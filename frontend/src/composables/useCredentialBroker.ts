import { ref } from "vue";
import type { CredentialRequest, CredentialResponse } from "../types/credential";

const pendingRequest = ref<CredentialRequest | null>(null);
const isPromptOpen = ref(false);
const isSubmitting = ref(false);
const lastError = ref<string | null>(null);

// In-memory session keyring cache
const sessionKeyring = ref<Map<string, string>>(new Map());

const DEFAULT_ACTIVE_INTERVAL = 800;
const DEFAULT_IDLE_INTERVAL = 5000;
const BOOST_DURATION_MS = 15000;

let pollTimer: ReturnType<typeof setTimeout> | null = null;
let isPollingActive = false;
let boostUntil = 0;
let configuredActiveInterval = DEFAULT_ACTIVE_INTERVAL;

export function useCredentialBroker() {
  /**
   * Trigger prompt modal for credential request
   */
  function promptCredential(req: CredentialRequest) {
    // Check if session keyring has cached passphrase
    const cacheKey = req.targetKey || req.prompt;
    if (sessionKeyring.value.has(cacheKey)) {
      const cachedPassphrase = sessionKeyring.value.get(cacheKey)!;
      // Cache hit: automatically respond without prompt
      respondCredential(req.id, cachedPassphrase, false);
      return;
    }

    pendingRequest.value = req;
    lastError.value = null;
    isPromptOpen.value = true;
  }

  /**
   * Respond to credential request
   */
  async function respondCredential(id: string, passphrase: string, rememberSession: boolean = false): Promise<boolean> {
    isSubmitting.value = true;
    lastError.value = null;
    try {
      if (rememberSession && pendingRequest.value) {
        const cacheKey = pendingRequest.value.targetKey || pendingRequest.value.prompt;
        sessionKeyring.value.set(cacheKey, passphrase);
      }

      const payload: CredentialResponse = {
        id,
        passphrase,
        rememberSession,
        canceled: false,
      };

      const res = await fetch("/api/git/credential-respond", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });

      if (!res.ok) {
        throw new Error(`Credential respond failed: HTTP ${res.status}`);
      }

      const data = await res.json();
      if (!data.success) {
        throw new Error(data.error || "Credential validation failed");
      }

      // Close prompt modal upon successful response
      isPromptOpen.value = false;
      pendingRequest.value = null;
      return true;
    } catch (err: any) {
      lastError.value = err.message || String(err);
      return false;
    } finally {
      isSubmitting.value = false;
    }
  }

  /**
   * Cancel current credential request
   */
  async function cancelCredential(): Promise<void> {
    if (!pendingRequest.value) {
      isPromptOpen.value = false;
      return;
    }

    const id = pendingRequest.value.id;
    try {
      await fetch("/api/git/credential-respond", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          id,
          passphrase: "",
          rememberSession: false,
          canceled: true,
        }),
      });
    } catch (e) {
      console.error("Cancel credential request failed:", e);
    } finally {
      isPromptOpen.value = false;
      pendingRequest.value = null;
      lastError.value = null;
    }
  }

  /**
   * Poll backend for pending credential requests
   */
  async function checkPendingRequest(): Promise<CredentialRequest | null> {
    try {
      const res = await fetch("/api/git/credential-pending");
      if (!res.ok) return null;
      const data = await res.json();
      if (data && data.request && (!pendingRequest.value || pendingRequest.value.id !== data.request.id)) {
        promptCredential(data.request);
        return data.request;
      }
      return null;
    } catch {
      return null;
    }
  }

  let isChecking = false;

  async function runPollLoop() {
    if (!isPollingActive) return;

    if (typeof document === "undefined" || !document.hidden) {
      if (!isPromptOpen.value && !isChecking) {
        isChecking = true;
        try {
          await checkPendingRequest();
        } finally {
          isChecking = false;
        }
      }
    }

    if (!isPollingActive) return;

    const isBoosted = Date.now() < boostUntil;
    const nextDelay = isBoosted ? configuredActiveInterval : DEFAULT_IDLE_INTERVAL;
    pollTimer = setTimeout(runPollLoop, nextDelay);
  }

  /**
   * Boost polling to high-frequency active rate for durationMs
   */
  function boostPolling(durationMs: number = BOOST_DURATION_MS) {
    boostUntil = Math.max(boostUntil, Date.now() + durationMs);
    if (!isPollingActive) return;
    if (pollTimer) {
      clearTimeout(pollTimer);
      pollTimer = null;
    }
    if (!isChecking) {
      runPollLoop();
    }
  }

  /**
   * Start background adaptive polling
   */
  function startBrokerPolling(intervalMs: number = DEFAULT_ACTIVE_INTERVAL) {
    if (isPollingActive) return;
    isPollingActive = true;
    configuredActiveInterval = intervalMs;
    runPollLoop();
  }

  /**
   * Stop polling
   */
  function stopBrokerPolling() {
    isPollingActive = false;
    if (pollTimer) {
      clearTimeout(pollTimer);
      pollTimer = null;
    }
  }

  /**
   * Clear all cached session passphrases
   */
  function clearSessionKeyring() {
    sessionKeyring.value.clear();
  }

  return {
    pendingRequest,
    isPromptOpen,
    isSubmitting,
    lastError,
    sessionKeyring,
    promptCredential,
    respondCredential,
    cancelCredential,
    checkPendingRequest,
    startBrokerPolling,
    stopBrokerPolling,
    boostPolling,
    clearSessionKeyring,
  };
}
