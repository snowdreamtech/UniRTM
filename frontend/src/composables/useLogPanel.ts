import { ref, computed, watch } from "vue";
import type { LogItem } from "../components/LogViewerModal.vue";
import { formatLogsToText } from "../utils/logFormatter";
import { logUserAction } from "../utils/logger";

export interface UseLogPanelOptions {
  t: (key: any, named?: Record<string, any>) => string;
  showToast: (msg: string, type?: "info" | "success" | "warning" | "error") => void;
}

export function useLogPanel(options: UseLogPanelOptions) {
  const { t, showToast } = options;

  const runtimeLogs = ref<LogItem[]>([]);

  const savedLogCardVisible = localStorage.getItem("unirtm_log_card_visible");
  const isLogCardVisible = ref(savedLogCardVisible !== null ? savedLogCardVisible === "true" : true);

  function toggleLogCard() {
    isLogCardVisible.value = !isLogCardVisible.value;
    localStorage.setItem("unirtm_log_card_visible", String(isLogCardVisible.value));
  }

  const savedAutoScroll = localStorage.getItem("unirtm_embedded_log_autoscroll");
  const embeddedAutoScroll = ref(savedAutoScroll !== null ? savedAutoScroll === "true" : true);
  const currentEmbeddedLogFilter = ref<string>("ALL");

  watch(embeddedAutoScroll, (val) => {
    localStorage.setItem("unirtm_embedded_log_autoscroll", String(val));
  });

  watch(currentEmbeddedLogFilter, (val) => {
    logUserAction("DEBUG", "User switched log filter tab in embedded log viewer", val);
  });

  watch(isLogCardVisible, (val) => {
    logUserAction("INFO", "User toggled log card visibility", val ? "expanded" : "collapsed");
  });

  const logLevels = computed(() => [
    { key: "ALL", label: t("log.level_all") },
    { key: "INFO", label: t("log.level_info") },
    { key: "WARN", label: t("log.level_warn") },
    { key: "ERROR", label: t("log.level_error") },
    { key: "DEBUG", label: t("log.level_debug") },
  ]);

  const filteredEmbeddedLogs = computed(() => {
    return runtimeLogs.value.filter((log) => {
      if (currentEmbeddedLogFilter.value === "ALL") return true;
      return (log.level || "").toUpperCase() === currentEmbeddedLogFilter.value;
    });
  });

  function handleCopyEmbeddedLogs() {
    logUserAction("INFO", "User copied embedded logs to clipboard");
    if (filteredEmbeddedLogs.value.length === 0) {
      showToast(t("log.empty"), "info");
      return;
    }
    const text = formatLogsToText(filteredEmbeddedLogs.value);
    navigator.clipboard.writeText(text);
    showToast(t("log.copied_toast"), "success");
  }

  async function handleExportEmbeddedLogs() {
    logUserAction("INFO", "User exported embedded logs");
    if (filteredEmbeddedLogs.value.length === 0) {
      showToast(t("log.empty"), "info");
      return;
    }
    const text = formatLogsToText(filteredEmbeddedLogs.value);
    const blob = new Blob([text], { type: "text/plain;charset=utf-8" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `unirtm-log-${new Date().toISOString().slice(0, 10)}.log`;
    a.click();
    URL.revokeObjectURL(url);
    showToast(t("log.exported_toast") || "Logs exported", "success");
  }

  function handleClearEmbeddedLogs() {
    runtimeLogs.value = [];
  }

  function appendLogEntry(entry: any) {
    if (entry) {
      runtimeLogs.value.push({
        id: entry.id,
        timestamp: entry.timestamp,
        level: entry.level || "INFO",
        message: entry.message || "",
        details: entry.details || "",
      });
      if (runtimeLogs.value.length > 500) {
        runtimeLogs.value.shift();
      }
    }
  }

  return {
    runtimeLogs,
    isLogCardVisible,
    toggleLogCard,
    embeddedAutoScroll,
    currentEmbeddedLogFilter,
    logLevels,
    filteredEmbeddedLogs,
    handleCopyEmbeddedLogs,
    handleExportEmbeddedLogs,
    handleClearEmbeddedLogs,
    appendLogEntry,
  };
}
