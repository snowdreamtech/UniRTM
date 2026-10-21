import { ref } from "vue";
import { logUserAction } from "../utils/logger";

const toastMessage = ref("");
const toastType = ref<"info" | "warning" | "error" | "success">("info");
let toastTimer: number | undefined;

export function useToast() {
  function showToast(msg: string, type: "info" | "warning" | "error" | "success" = "info") {
    // Strip any leading emoji prefixes from the message text to avoid duplicate icons in Toast
    const cleanMsg = msg
      ? msg.replace(/^[\s\uFE0F]*[\u{1F300}-\u{1F9FF}\u{2600}-\u{26FF}\u{2700}-\u{27BF}][\s\uFE0F]*/u, "").trim()
      : "";
    toastMessage.value = cleanMsg || msg;
    toastType.value = type;
    if (toastTimer) clearTimeout(toastTimer);
    toastTimer = setTimeout(() => {
      toastMessage.value = "";
    }, 4000) as unknown as number;
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
