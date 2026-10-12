import { ref } from "vue";
import { logUserAction } from "../utils/logger";

const toastMessage = ref("");
const toastType = ref<"info" | "warning" | "error" | "success">("info");
let toastTimer: number | undefined;

export function useToast() {
  function showToast(msg: string, type: "info" | "warning" | "error" | "success" = "info") {
    const cleanMsg = msg ? msg.replace(/^[\s\uFE0F]*[⚠️❌🎉ℹ️✅🚨⚡️❗][\s\uFE0F]*/, "").trim() : "";
    toastMessage.value = cleanMsg || msg;
    toastType.value = type;
    if (toastTimer) clearTimeout(toastTimer);
    toastTimer = window.setTimeout(() => {
      toastMessage.value = "";
    }, 4000);
  }

  function dismissToast() {
    toastMessage.value = "";
    logUserAction("DEBUG", "User dismissed toast notification");
  }

  return {
    toastMessage,
    toastType,
    showToast,
    dismissToast,
  };
}
