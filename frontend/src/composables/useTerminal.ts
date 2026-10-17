import { ref, computed } from "vue";
import { useToast } from "./useToast";
import { t } from "../i18n";

export interface TerminalItem {
  id: string;
  name: string;
  path: string;
  is_default: boolean;
}

const terminals = ref<TerminalItem[]>([]);
const defaultTerminalPath = ref<string>("");
const currentTerminalPath = ref<string>("");
const isLoadingTerminals = ref<boolean>(false);
const terminalError = ref<string | null>(null);

export function useTerminal() {
  const { showToast } = useToast();

  const currentTerminal = computed(() => {
    if (!currentTerminalPath.value && terminals.value.length > 0) {
      return terminals.value.find((t) => t.is_default) || terminals.value[0];
    }
    return (
      terminals.value.find((t) => t.path === currentTerminalPath.value || t.id === currentTerminalPath.value) ||
      terminals.value.find((t) => t.is_default) ||
      terminals.value[0] ||
      null
    );
  });

  const defaultTerminal = computed(() => {
    return terminals.value.find((t) => t.is_default) || terminals.value[0] || null;
  });

  async function fetchTerminals() {
    isLoadingTerminals.value = true;
    terminalError.value = null;
    try {
      const res = await fetch("/api/terminals");
      if (res.ok) {
        const data = await res.json();
        terminals.value = data.terminals || [];
        defaultTerminalPath.value = data.default || "";
        currentTerminalPath.value = data.current || data.default || "";
      } else {
        terminalError.value = "Failed to load terminals";
      }
    } catch (err: any) {
      terminalError.value = err.message || "Network error loading terminals";
    } finally {
      isLoadingTerminals.value = false;
    }
  }

  async function setPreferredTerminal(terminalPath: string) {
    currentTerminalPath.value = terminalPath;
    try {
      const res = await fetch("/api/config", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          scope: "user",
          key: "settings.terminal",
          value: terminalPath,
        }),
      });
      if (res.ok) {
        showToast(t("terminal.save_success") || "默认终端已更新", "success");
        return true;
      } else {
        const errData = await res.json().catch(() => ({}));
        showToast(errData.error || t("terminal.save_failed") || "保存终端设置失败", "error");
        return false;
      }
    } catch (err: any) {
      showToast(err.message || t("terminal.save_failed") || "保存终端设置失败", "error");
      return false;
    }
  }

  return {
    terminals,
    defaultTerminalPath,
    currentTerminalPath,
    currentTerminal,
    defaultTerminal,
    isLoadingTerminals,
    terminalError,
    fetchTerminals,
    setPreferredTerminal,
  };
}
