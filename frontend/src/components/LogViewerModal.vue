<template>
  <div v-if="isOpen" class="modal-overlay" @click.self="$emit('close')">
    <div class="modal-content glass-card log-modal">
      <div class="modal-header">
        <div class="title-with-badge">
          <h3>📋 {{ t("log.title") }}</h3>
          <span class="badge live-badge">● {{ t("log.live") }}</span>
        </div>
        <button class="btn-close" @click="$emit('close')">✕</button>
      </div>

      <div class="modal-body">
        <!-- Log Filter & Search Bar -->
        <div class="log-controls">
          <div class="filter-tabs">
            <button
              v-for="level in logLevels"
              :key="level.key"
              class="btn-tab"
              :class="{ active: currentFilter === level.key, [level.key.toLowerCase()]: true }"
              @click="currentFilter = level.key"
            >
              {{ level.label }} ({{ getLevelCount(level.key) }})
            </button>
          </div>

          <div class="search-box">
            <input type="text" v-model="searchQuery" :placeholder="t('log.search_placeholder')" class="search-input" />
          </div>
        </div>

        <!-- Terminal Log Window -->
        <div class="terminal-window" ref="terminalRef">
          <div v-if="filteredLogs.length === 0" class="empty-logs">
            {{ t("log.empty") }}
          </div>
          <div
            v-for="log in filteredLogs"
            :key="log.id || String(log.timestamp)"
            class="log-row"
            :class="log.level.toLowerCase()"
          >
            <span class="log-time"
              ><bdi>{{ formatLogTime(log.timestamp) }}</bdi></span
            >
            <span class="log-level-badge" :class="log.level.toLowerCase()"
              ><bdi>[{{ log.level }}]</bdi></span
            >
            <div class="log-content">
              <span class="log-msg"
                ><bdi>{{ log.message }}</bdi></span
              >
              <span v-if="log.details" class="log-details"
                ><bdi>{{ log.details }}</bdi></span
              >
            </div>
          </div>
        </div>
      </div>

      <div class="modal-footer">
        <div class="footer-left">
          <label class="auto-scroll-label">
            <input type="checkbox" v-model="autoScroll" />
            {{ t("log.auto_scroll") }}
          </label>
        </div>
        <div class="footer-actions">
          <button class="btn btn-secondary" @click="copyAllLogs">📋 {{ t("log.copy") }}</button>
          <button class="btn btn-secondary" @click="exportLogFile">📥 {{ t("log.export") }}</button>
          <button class="btn btn-danger" @click="$emit('clear')">🗑️ {{ t("log.clear") }}</button>
          <button class="btn btn-primary" @click="$emit('close')">
            {{ t("common.close") }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from "vue";
import { t } from "../i18n";
import { formatLogTime, formatLogsToText } from "../utils/logFormatter";

export interface LogItem {
  id?: number;
  timestamp: string | Date;
  level: string;
  message: string;
  details?: string;
}

const props = defineProps<{
  isOpen: boolean;
  logs: LogItem[];
}>();

defineEmits(["close", "clear"]);

const currentFilter = ref("ALL");
const searchQuery = ref("");
const savedAutoScroll = localStorage.getItem("unigodesktop_log_autoscroll");
const autoScroll = ref(savedAutoScroll !== null ? savedAutoScroll === "true" : true);
const terminalRef = ref<HTMLDivElement | null>(null);

watch(autoScroll, (val) => {
  localStorage.setItem("unigodesktop_log_autoscroll", String(val));
});

const logLevels = computed(() => [
  { key: "ALL", label: t("log.level_all") },
  { key: "INFO", label: t("log.level_info") },
  { key: "WARN", label: t("log.level_warn") },
  { key: "ERROR", label: t("log.level_error") },
  { key: "DEBUG", label: t("log.level_debug") },
]);

function getLevelCount(level: string): number {
  if (level === "ALL") return props.logs.length;
  return props.logs.filter((l) => (l.level || "").toUpperCase() === level).length;
}

const filteredLogs = computed(() => {
  return props.logs.filter((log) => {
    const matchesLevel = currentFilter.value === "ALL" || (log.level || "").toUpperCase() === currentFilter.value;
    const query = searchQuery.value.trim().toLowerCase();
    const matchesQuery =
      !query || log.message.toLowerCase().includes(query) || (log.details && log.details.toLowerCase().includes(query));
    return matchesLevel && matchesQuery;
  });
});

function scrollToBottom() {
  if (autoScroll.value && terminalRef.value) {
    nextTick(() => {
      if (terminalRef.value) {
        terminalRef.value.scrollTop = terminalRef.value.scrollHeight;
      }
    });
  }
}

const logUserAction = (level: string, message: string, details: string = "") => {
  const app = (window as any)?.go?.main?.App;
  if (app && typeof app.LogAction === "function") {
    app.LogAction(level, message, details);
  }
};

watch(currentFilter, (val) => {
  logUserAction("DEBUG", "User switched log level filter tab in full Log Viewer modal", val);
});

watch(autoScroll, (val) => {
  logUserAction("DEBUG", "User toggled auto-scroll in full Log Viewer modal", val ? "enabled" : "disabled");
});

watch(
  () => props.logs.length,
  () => {
    scrollToBottom();
  }
);

watch(
  () => props.isOpen,
  (newVal) => {
    if (newVal) {
      logUserAction("INFO", "User opened full Log Viewer modal");
      scrollToBottom();
    }
  }
);

function copyAllLogs() {
  logUserAction("INFO", "User copied logs from full Log Viewer modal");
  const text = formatLogsToText(filteredLogs.value);
  navigator.clipboard.writeText(text);
  alert(t("log.copied_toast"));
}

function exportLogFile() {
  logUserAction("INFO", "User exported logs from full Log Viewer modal");
  const text = formatLogsToText(filteredLogs.value);
  const blob = new Blob([text], { type: "text/plain;charset=utf-8" });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = `unigo_log_${new Date().toISOString().slice(0, 10)}.log`;
  a.click();
  URL.revokeObjectURL(url);
}
</script>

<style scoped>
.log-modal {
  max-width: 900px;
  width: 90vw;
  max-height: 85vh;
  display: flex;
  flex-direction: column;
}

.title-with-badge {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.live-badge {
  background: rgba(16, 185, 129, 0.15);
  color: #10b981;
  font-size: 0.75rem;
  padding: 0.2rem 0.6rem;
  border-radius: 12px;
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0% {
    opacity: 1;
  }
  50% {
    opacity: 0.4;
  }
  100% {
    opacity: 1;
  }
}

.log-controls {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.85rem;
  gap: 1rem;
}

.filter-tabs {
  display: flex;
  gap: 0.4rem;
}

.btn-tab {
  background: var(--subtab-container-bg);
  border: 1px solid var(--card-border);
  color: var(--subtab-btn-text);
  padding: 0.3rem 0.7rem;
  border-radius: 6px;
  font-size: 0.8rem;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-tab.active {
  background: var(--tab-btn-active-bg);
  color: var(--tab-btn-active-text);
  border-color: var(--subtab-btn-active-border);
  font-weight: 600;
}

.search-input {
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  color: var(--text-main);
  padding: 0.35rem 0.75rem;
  border-radius: 6px;
  font-size: 0.85rem;
  width: 220px;
}

.terminal-window {
  background: var(--terminal-bg);
  border: 1px solid var(--terminal-border);
  color: var(--terminal-text);
  border-radius: 10px;
  padding: 1rem 1.15rem;
  font-family: "JetBrains Mono", "Fira Code", "Courier New", monospace;
  font-size: 0.82rem;
  line-height: 1.6;
  height: 480px;
  overflow-y: auto;
  overflow-x: hidden;
  border: 1px solid rgba(255, 255, 255, 0.08);
}

.terminal-window .empty-logs {
  color: #64748b;
  text-align: center;
  padding: 4rem 0;
}

.terminal-window .log-row {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
  padding: 0.35rem 0;
  border-bottom: 1px dashed rgba(255, 255, 255, 0.05);
  box-sizing: border-box;
  width: 100%;
}

.terminal-window .log-time {
  color: #64748b;
  font-size: 0.78rem;
  font-family: "JetBrains Mono", monospace;
  white-space: nowrap;
  flex-shrink: 0;
  width: 100px;
  min-width: 100px;
  height: 22px;
  line-height: 22px;
  display: inline-flex;
  align-items: center;
}

.terminal-window .log-level-badge {
  font-size: 0.7rem;
  font-weight: 700;
  height: 20px;
  line-height: 18px;
  padding: 0 0.5rem;
  border-radius: 4px;
  white-space: nowrap;
  flex-shrink: 0;
  width: 64px;
  min-width: 64px;
  box-sizing: border-box;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  letter-spacing: 0.5px;
  margin-top: 1px;
}

.terminal-window .log-level-badge.info {
  background: rgba(16, 185, 129, 0.18);
  color: #34d399;
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.terminal-window .log-level-badge.warn {
  background: rgba(245, 158, 11, 0.18);
  color: #fbbf24;
  border: 1px solid rgba(245, 158, 11, 0.3);
}

.terminal-window .log-level-badge.error {
  background: rgba(239, 68, 68, 0.2);
  color: #f87171;
  border: 1px solid rgba(239, 68, 68, 0.35);
}

.terminal-window .log-level-badge.debug {
  background: rgba(168, 85, 247, 0.18);
  color: #c084fc;
  border: 1px solid rgba(168, 85, 247, 0.3);
}

.terminal-window .log-content {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: 0.5rem;
}

.terminal-window .log-msg {
  color: #e2e8f0;
  font-size: 0.82rem;
  line-height: 22px;
  white-space: pre-wrap;
  word-break: break-word;
  overflow-wrap: anywhere;
}

.terminal-window .log-row.info .log-msg {
  color: #f1f5f9;
}
.terminal-window .log-row.warn .log-msg {
  color: #fde047;
}
.terminal-window .log-row.error .log-msg {
  color: #fca5a5;
}
.terminal-window .log-row.debug .log-msg {
  color: #c084fc;
}

.terminal-window .log-details {
  color: #94a3b8;
  font-size: 0.78rem;
  line-height: 22px;
  white-space: pre-wrap;
  word-break: break-word;
  overflow-wrap: anywhere;
  direction: ltr !important;
  text-align: left !important;
  unicode-bidi: embed;
}

.footer-left {
  display: flex;
  align-items: center;
}

.auto-scroll-label {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  font-size: 0.85rem;
  cursor: pointer;
}

.footer-actions {
  display: flex;
  gap: 0.5rem;
}

/* Light Mode Overrides for LogViewerModal */
[data-theme="light"] .log-modal {
  background: #ffffff;
  border-color: #cbd5e1;
  box-shadow: 0 16px 48px rgba(15, 23, 42, 0.2);
}

[data-theme="light"] .btn-tab {
  background: #f1f5f9;
  border-color: #cbd5e1;
  color: #475569;
  font-weight: 600;
}

[data-theme="light"] .btn-tab:hover {
  background: #e2e8f0;
  color: #0f172a;
}

[data-theme="light"] .btn-tab.active {
  background: #0284c7;
  color: #ffffff;
  border-color: #0284c7;
}

[data-theme="light"] .search-input {
  background: #ffffff;
  border-color: #cbd5e1;
  color: #0f172a;
}

[data-theme="light"] .auto-scroll-label {
  color: #334155;
  font-weight: 500;
}

/* Light Mode Terminal Window & Log Row Colors for Modal */
[data-theme="light"] .terminal-window {
  background: #f8fafc;
  border-color: #cbd5e1;
  box-shadow: inset 0 2px 4px rgba(15, 23, 42, 0.04);
}

[data-theme="light"] .terminal-window .empty-logs {
  color: #94a3b8;
}

[data-theme="light"] .terminal-window .log-row {
  border-bottom-color: #e2e8f0;
}

[data-theme="light"] .terminal-window .log-time {
  color: #64748b;
}

[data-theme="light"] .terminal-window .log-msg {
  color: #0f172a;
}

[data-theme="light"] .terminal-window .log-details {
  color: #64748b;
}

[data-theme="light"] .terminal-window .log-row.info .log-msg {
  color: #0f172a;
}

[data-theme="light"] .terminal-window .log-row.warn .log-msg {
  color: #b45309;
}

[data-theme="light"] .terminal-window .log-row.error .log-msg {
  color: #dc2626;
}

[data-theme="light"] .terminal-window .log-row.debug .log-msg {
  color: #6b21a8;
}

[data-theme="light"] .terminal-window .log-level-badge.info {
  background: #ecfdf5;
  color: #047857;
  border: 1px solid #a7f3d0;
}

[data-theme="light"] .terminal-window .log-level-badge.warn {
  background: #fffbeb;
  color: #b45309;
  border: 1px solid #fde68a;
}

[data-theme="light"] .terminal-window .log-level-badge.error {
  background: #fef2f2;
  color: #b91c1c;
  border: 1px solid #fca5a5;
}

[data-theme="light"] .terminal-window .log-level-badge.debug {
  background: #f3e8ff;
  color: #6b21a8;
  border: 1px solid #e9d5ff;
}
</style>
