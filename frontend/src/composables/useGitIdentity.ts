import { ref, computed } from "vue";
import type { GitProfile, SystemKeysSummary, NativeAgentStatus, CredentialAuditRecord } from "../types/credential";

const profiles = ref<GitProfile[]>([]);
const activeProfileId = ref<string>("");
const systemKeys = ref<SystemKeysSummary>({ sshKeys: [], gpgKeys: [] });
const agentStatus = ref<NativeAgentStatus>({
  sshAgentActive: false,
  sshAgentKeys: 0,
  gpgAgentActive: false,
});
const auditRecords = ref<CredentialAuditRecord[]>([]);
const isLoading = ref<boolean>(false);
const errorMsg = ref<string | null>(null);

let inFlightProfilesPromise: Promise<void> | null = null;

export function useGitIdentity() {
  const activeProfile = computed<GitProfile | undefined>(() =>
    profiles.value.find((p: GitProfile) => p.id === activeProfileId.value)
  );

  const loadProfiles = async (): Promise<void> => {
    if (inFlightProfilesPromise) {
      return inFlightProfilesPromise;
    }
    inFlightProfilesPromise = (async () => {
      try {
        isLoading.value = true;
        errorMsg.value = null;
        const res = await fetch("/api/git/profiles", { cache: "no-store" });
        if (res.ok) {
          const data = await res.json();
          profiles.value = data.profiles || [];
          activeProfileId.value = data.activeProfileId || "";
        }
      } catch (err: unknown) {
        errorMsg.value = err instanceof Error ? err.message : String(err);
      } finally {
        isLoading.value = false;
        inFlightProfilesPromise = null;
      }
    })();
    return inFlightProfilesPromise;
  };

  const saveProfile = async (profile: Partial<GitProfile>): Promise<boolean> => {
    try {
      isLoading.value = true;
      errorMsg.value = null;
      const res = await fetch("/api/git/profiles", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(profile),
      });
      if (res.ok) {
        await loadProfiles();
        return true;
      }
      return false;
    } catch (err: unknown) {
      errorMsg.value = err instanceof Error ? err.message : String(err);
      return false;
    } finally {
      isLoading.value = false;
    }
  };

  const deleteProfile = async (id: string): Promise<boolean> => {
    try {
      isLoading.value = true;
      errorMsg.value = null;
      const res = await fetch(`/api/git/profiles?id=${encodeURIComponent(id)}`, {
        method: "DELETE",
      });
      if (res.ok) {
        await loadProfiles();
        return true;
      }
      return false;
    } catch (err: unknown) {
      errorMsg.value = err instanceof Error ? err.message : String(err);
      return false;
    } finally {
      isLoading.value = false;
    }
  };

  const applyProfile = async (profileId: string, workspaceDir = ""): Promise<boolean> => {
    try {
      isLoading.value = true;
      errorMsg.value = null;
      const res = await fetch("/api/git/profile/apply", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ profileId, workspaceDir }),
      });
      if (res.ok) {
        activeProfileId.value = profileId;
        return true;
      }
      const data = await res.json().catch(() => ({}));
      if (data.error) {
        errorMsg.value = data.error;
      }
      return false;
    } catch (err: unknown) {
      errorMsg.value = err instanceof Error ? err.message : String(err);
      return false;
    } finally {
      isLoading.value = false;
    }
  };

  const loadSystemKeys = async (): Promise<void> => {
    try {
      const res = await fetch("/api/git/system-keys", { cache: "no-store" });
      if (res.ok) {
        const data = await res.json();
        systemKeys.value = data;
      }
    } catch (err: unknown) {
      // Graceful fallback for non-fatal key detection
    }
  };

  const loadAgentStatus = async (): Promise<void> => {
    try {
      const res = await fetch("/api/git/agent-status", { cache: "no-store" });
      if (res.ok) {
        const data = await res.json();
        agentStatus.value = data;
      }
    } catch (err: unknown) {
      // Graceful fallback
    }
  };

  const loadAuditRecords = async (): Promise<void> => {
    try {
      const res = await fetch("/api/git/credential-audit", {
        cache: "no-store",
      });
      if (res.ok) {
        const data = await res.json();
        auditRecords.value = data.records || [];
      }
    } catch (err: unknown) {
      // Graceful fallback
    }
  };

  const clearKeyringCache = async (): Promise<boolean> => {
    try {
      const res = await fetch("/api/git/credential-clear-cache", {
        method: "POST",
      });
      return res.ok;
    } catch {
      return false;
    }
  };

  return {
    profiles,
    activeProfileId,
    activeProfile,
    systemKeys,
    agentStatus,
    auditRecords,
    isLoading,
    errorMsg,
    loadProfiles,
    saveProfile,
    deleteProfile,
    applyProfile,
    loadSystemKeys,
    loadAgentStatus,
    loadAuditRecords,
    clearKeyringCache,
  };
}
