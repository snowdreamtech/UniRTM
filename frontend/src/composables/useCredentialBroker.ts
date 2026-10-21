import { ref } from "vue";
import type { CredentialRequest, CredentialResponse } from "../types/credential";

const pendingRequest = ref<CredentialRequest | null>(null);
const isPromptOpen = ref(false);
const isSubmitting = ref(false);
const lastError = ref<string | null>(null);

// 前端会话安全缓存 (纯内存)
const sessionKeyring = ref<Map<string, string>>(new Map());

let pollTimer: ReturnType<typeof setInterval> | null = null;

export function useCredentialBroker() {
  /**
   * 触发显示凭证输入弹窗
   */
  function promptCredential(req: CredentialRequest) {
    // 检查是否已有会话缓存
    const cacheKey = req.targetKey || req.prompt;
    if (sessionKeyring.value.has(cacheKey)) {
      const cachedPassphrase = sessionKeyring.value.get(cacheKey)!;
      // 命中会话缓存，自动响应免密通过
      respondCredential(req.id, cachedPassphrase, false);
      return;
    }

    pendingRequest.value = req;
    lastError.value = null;
    isPromptOpen.value = true;
  }

  /**
   * 响应凭证请求
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
        throw new Error(data.error || "密码验证失败");
      }

      // 成功响应后关闭弹窗
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
   * 取消当前凭证请求
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
   * 轮询后端是否有正在等待的凭证请求
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

  /**
   * 启动后台轮询监听
   */
  function startBrokerPolling(intervalMs: number = 800) {
    if (pollTimer) return;
    pollTimer = setInterval(checkPendingRequest, intervalMs);
  }

  /**
   * 停止轮询监听
   */
  function stopBrokerPolling() {
    if (pollTimer) {
      clearInterval(pollTimer);
      pollTimer = null;
    }
  }

  /**
   * 清理全部会话密码缓存
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
    clearSessionKeyring,
  };
}
