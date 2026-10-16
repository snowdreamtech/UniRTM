<template>
  <div v-if="isOpen" class="bottom-drawer" id="app-bottom-drawer">
    <!-- Drawer Header Bar -->
    <div class="drawer-header">
      <!-- Tabs on the left -->
      <div class="drawer-tabs">
        <button
          class="drawer-tab"
          :class="{ active: currentTab === 'output' }"
          @click="currentTab = 'output'"
        >
          <span style="display: inline-flex; align-items: center; gap: 6px;">
            <TechIcon id="workbench:logs" :size="13" />
            {{ t("drawer.tab_output") }}
          </span>
          <span v-if="outputLogs.length > 0" class="tab-badge">{{ outputLogs.length }}</span>
        </button>
        <button
          class="drawer-tab"
          :class="{ active: currentTab === 'tasks' }"
          @click="currentTab = 'tasks'"
        >
          <span style="display: inline-flex; align-items: center; gap: 6px;">
            <TechIcon id="workbench:tasks" :size="13" />
            {{ t("drawer.tab_tasks") }}
          </span>
          <span v-if="tasksLogs.length > 0" class="tab-badge">{{ tasksLogs.length }}</span>
        </button>
        <button
          class="drawer-tab"
          :class="{ active: currentTab === 'doctor' }"
          @click="currentTab = 'doctor'"
        >
          <span style="display: inline-flex; align-items: center; gap: 6px;">
            <TechIcon id="workbench:doctor" :size="13" />
            {{ t("drawer.tab_doctor") }}
          </span>
          <span v-if="doctorLogs.length > 0" class="tab-badge">{{ doctorLogs.length }}</span>
        </button>
      </div>

      <!-- Action Tools on the right -->
      <div class="drawer-actions">
        <button
          class="drawer-action-btn"
          :title="clearLogsTitle"
          :data-tooltip="clearLogsTitle"
          @click="$emit('clear-logs', currentTab)"
        >
          <svg viewBox="0 0 16 16" width="13" height="13" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="8" cy="8" r="6" />
            <path d="M3.8 3.8l8.4 8.4" />
          </svg>
        </button>
        <button
          class="drawer-action-btn"
          :class="{ active: isCopiedAll }"
          :title="copyAllTitle"
          :data-tooltip="copyAllTitle"
          @click="handleCopyAll"
        >
          <svg v-if="!isCopiedAll" viewBox="0 0 16 16" width="13" height="13" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
            <rect x="5.5" y="5.5" width="8" height="8" rx="1.5" />
            <path d="M10.5 5.5V3.5a1.5 1.5 0 0 0-1.5-1.5h-5A1.5 1.5 0 0 0 2.5 3.5v5a1.5 1.5 0 0 0 1.5 1.5h1.5" />
          </svg>
          <svg v-else viewBox="0 0 16 16" width="13" height="13" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" style="color: var(--success, #10b981);">
            <path d="M3.5 8.5l3 3 6-6" />
          </svg>
        </button>
        <button
          class="drawer-action-btn"
          :class="{ active: autoScroll }"
          :title="autoScrollTitle"
          :data-tooltip="autoScrollTitle"
          @click="autoScroll = !autoScroll"
        >
          <svg viewBox="0 0 16 16" width="13" height="13" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
            <path d="M8 2.5v8.5" />
            <path d="M4.5 7.5L8 11l3.5-3.5" />
            <path d="M2.5 13.5h11" />
          </svg>
        </button>
        <button
          class="drawer-action-btn"
          :class="{ active: isMaximized }"
          :title="maximizeTitle"
          :data-tooltip="maximizeTitle"
          @click="isMaximized = !isMaximized"
        >
          <svg v-if="!isMaximized" viewBox="0 0 16 16" width="13" height="13" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M2.5 10.5L8 5l5.5 5.5" />
          </svg>
          <svg v-else viewBox="0 0 16 16" width="13" height="13" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M2.5 5.5L8 11l5.5-5.5" />
          </svg>
        </button>
        <button
          class="drawer-action-btn"
          :title="closeTitle"
          :data-tooltip="closeTitle"
          @click="$emit('close')"
        >
          <svg viewBox="0 0 16 16" width="13" height="13" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M3.5 3.5l9 9M12.5 3.5l-9 9" />
          </svg>
        </button>
      </div>
    </div>

    <!-- Drawer Body -->
    <div
      class="drawer-body font-mono"
      ref="logContainerRef"
      :style="{ height: isMaximized ? '380px' : '170px' }"
    >
      <div v-if="displayLogs.length === 0" class="empty-logs">
        <span>{{ emptyStateText }}</span>
      </div>

      <div
        v-for="(log, idx) in displayLogs"
        :key="idx"
        class="log-row"
        :class="['level-' + log.level.toLowerCase(), { copied: copiedRowIdx === idx }]"
        @dblclick="handleCopyRow(log, idx)"
      >
        <span class="log-time">{{ log.timestamp }}</span>
        <span class="log-badge" :class="'badge-' + log.level.toLowerCase()">
          {{ log.level }}
        </span>
        <span class="log-msg">{{ log.message }}</span>
        <span v-if="log.details" class="log-details">{{ log.details }}</span>

        <!-- Single Line Copy Button -->
        <button
          class="log-row-copy-btn"
          :class="{ copied: copiedRowIdx === idx }"
          @click.stop="handleCopyRow(log, idx)"
          :title="t('drawer.copy_row') || '复制单行日志'"
        >
          <svg v-if="copiedRowIdx !== idx" viewBox="0 0 16 16" width="11" height="11" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
            <rect x="5.5" y="5.5" width="8" height="8" rx="1.5" />
            <path d="M10.5 5.5V3.5a1.5 1.5 0 0 0-1.5-1.5h-5A1.5 1.5 0 0 0 2.5 3.5v5a1.5 1.5 0 0 0 1.5 1.5h1.5" />
          </svg>
          <svg v-else viewBox="0 0 16 16" width="11" height="11" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" style="color: var(--success, #10b981);">
            <path d="M3.5 8.5l3 3 6-6" />
          </svg>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from "vue";
import { t } from "../../i18n";
import TechIcon from "../common/TechIcon.vue";
import { useToast } from "../../composables/useToast";

export interface LogItem {
  timestamp: string;
  level: "INFO" | "SUCCESS" | "WARN" | "ERROR";
  message: string;
  details?: string;
  category?: "output" | "tasks" | "doctor";
}

const props = withDefaults(
  defineProps<{
    isOpen: boolean;
    logs: LogItem[];
    activeTab?: "output" | "tasks" | "doctor";
  }>(),
  {
    activeTab: undefined,
  }
);

const emit = defineEmits<{
  (e: "close"): void;
  (e: "clear-logs", category?: "output" | "tasks" | "doctor"): void;
  (e: "update:activeTab", tab: "output" | "tasks" | "doctor"): void;
}>();

const DRAWER_TAB_KEY = "unirtm_drawer_tab";
const DRAWER_MAX_KEY = "unirtm_drawer_maximized";
const DRAWER_AUTOSCROLL_KEY = "unirtm_drawer_autoscroll";

function getInitialTab(): "output" | "tasks" | "doctor" {
  if (typeof localStorage === "undefined") return "output";
  const saved = localStorage.getItem(DRAWER_TAB_KEY);
  if (saved === "output" || saved === "tasks" || saved === "doctor") {
    return saved;
  }
  return "output";
}

function getInitialMaximized(): boolean {
  if (typeof localStorage === "undefined") return false;
  return localStorage.getItem(DRAWER_MAX_KEY) === "true";
}

function getInitialAutoScroll(): boolean {
  if (typeof localStorage === "undefined") return true;
  const saved = localStorage.getItem(DRAWER_AUTOSCROLL_KEY);
  if (saved === null) return true;
  return saved === "true";
}

const internalActiveTab = ref<"output" | "tasks" | "doctor">(getInitialTab());

const currentTab = computed<"output" | "tasks" | "doctor">({
  get: () => props.activeTab ?? internalActiveTab.value,
  set: (val: "output" | "tasks" | "doctor") => {
    internalActiveTab.value = val;
    emit("update:activeTab", val);
  },
});

const autoScroll = ref(getInitialAutoScroll());
const isMaximized = ref(getInitialMaximized());
const logContainerRef = ref<HTMLElement | null>(null);

watch(currentTab, (val) => {
  if (typeof localStorage !== "undefined") {
    localStorage.setItem(DRAWER_TAB_KEY, val);
  }
});

watch(isMaximized, (val) => {
  if (typeof localStorage !== "undefined") {
    localStorage.setItem(DRAWER_MAX_KEY, String(val));
  }
});

watch(autoScroll, (val) => {
  if (typeof localStorage !== "undefined") {
    localStorage.setItem(DRAWER_AUTOSCROLL_KEY, String(val));
  }
});

const outputLogs = computed(() =>
  props.logs.filter((l) => !l.category || l.category === "output")
);
const tasksLogs = computed(() =>
  props.logs.filter((l) => l.category === "tasks")
);
const doctorLogs = computed(() =>
  props.logs.filter((l) => l.category === "doctor")
);

const displayLogs = computed(() => {
  if (currentTab.value === "tasks") {
    return tasksLogs.value;
  }
  if (currentTab.value === "doctor") {
    return doctorLogs.value;
  }
  return outputLogs.value;
});

const emptyStateText = computed(() => {
  if (currentTab.value === "tasks") {
    return (
      t("drawer.empty_tasks") ||
      "暂无任务执行日志。在任务视图中点击“运行”以在此查看执行终端输出。"
    );
  }
  if (currentTab.value === "doctor") {
    return (
      t("drawer.empty_doctor") ||
      "暂无环境体检报告。在环境诊断视图中执行“重新全面体检”以查看详情。"
    );
  }
  return t("drawer.empty_logs") || "暂无系统输出日志。系统运行与扩展事件将在此显示。";
});

const clearLogsTitle = computed(() => t("drawer.clear_logs") || "清空控制台日志");
const copyAllTitle = computed(() => t("drawer.copy_all") || "复制全部日志");
const autoScrollTitle = computed(() =>
  autoScroll.value
    ? t("drawer.scroll_enabled") || "自动滚动（已开启）"
    : t("drawer.scroll_disabled") || "自动滚动（已暂停）"
);
const maximizeTitle = computed(() =>
  isMaximized.value
    ? t("drawer.restore") || "还原视窗默认高度"
    : t("drawer.maximize") || "最大化视窗高度"
);
const closeTitle = computed(() => t("drawer.close_hint") || "关闭控制台面板");

const { showToast } = useToast();
const isCopiedAll = ref(false);
const copiedRowIdx = ref<number | null>(null);

function formatLogLine(log: LogItem): string {
  const parts = [`[${log.timestamp}]`, `[${log.level}]`, log.message];
  if (log.details) {
    parts.push(log.details);
  }
  return parts.join(" ");
}

async function copyText(text: string): Promise<boolean> {
  try {
    if (navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(text);
      return true;
    }
  } catch (e) {
    // fallback
  }
  try {
    const textArea = document.createElement("textarea");
    textArea.value = text;
    textArea.style.position = "fixed";
    textArea.style.left = "-9999px";
    textArea.style.top = "-9999px";
    document.body.appendChild(textArea);
    textArea.focus();
    textArea.select();
    const successful = document.execCommand("copy");
    document.body.removeChild(textArea);
    return successful;
  } catch (e) {
    return false;
  }
}

async function handleCopyAll() {
  if (displayLogs.value.length === 0) {
    showToast(t("drawer.empty_copy") || "当前暂无日志可复制", "warning");
    return;
  }
  const text = displayLogs.value.map(formatLogLine).join("\n");
  const success = await copyText(text);
  if (success) {
    isCopiedAll.value = true;
    setTimeout(() => {
      isCopiedAll.value = false;
    }, 1500);
    showToast(t("drawer.copy_all_success") || "已复制当前面板全部日志到剪贴板", "success");
  }
}

async function handleCopyRow(log: LogItem, idx: number) {
  const lineText = formatLogLine(log);
  const success = await copyText(lineText);
  if (success) {
    copiedRowIdx.value = idx;
    setTimeout(() => {
      if (copiedRowIdx.value === idx) {
        copiedRowIdx.value = null;
      }
    }, 1500);
    showToast(t("drawer.copy_row_success") || "已复制单行日志", "success");
  }
}

watch(
  () => props.logs.length,
  () => {
    if (autoScroll.value && logContainerRef.value) {
      nextTick(() => {
        if (logContainerRef.value) {
          logContainerRef.value.scrollTop = logContainerRef.value.scrollHeight;
        }
      });
    }
  }
);
</script>

<style scoped>
.bottom-drawer {
  width: 100%;
  background: var(--card-bg);
  border-top: 1px solid var(--card-border);
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  z-index: 55;
  backdrop-filter: blur(12px);
  animation: drawer-slide-up 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

@keyframes drawer-slide-up {
  from {
    transform: translateY(10px);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}

.drawer-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 32px;
  min-height: 32px;
  background: var(--input-bg);
  border-bottom: 1px solid var(--card-border);
  padding: 0 0.5rem;
  user-select: none;
  -webkit-user-select: none;
}

.drawer-tabs {
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.drawer-tab {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  background: transparent;
  border: none;
  border-bottom: 2px solid transparent;
  color: var(--text-muted);
  font-size: 0.75rem;
  font-weight: 500;
  padding: 0.35rem 0.65rem;
  cursor: pointer;
  transition: all 0.15s ease;
}

.drawer-tab:hover {
  color: var(--text-main);
}

.drawer-tab.active {
  color: var(--accent-cyan);
  border-bottom-color: var(--accent-cyan);
  font-weight: 600;
}

.tab-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 0.65rem;
  font-weight: 600;
  padding: 0.05rem 0.35rem;
  min-width: 1rem;
  height: 1rem;
  border-radius: 9999px;
  background: var(--card-border);
  color: var(--text-muted);
  line-height: 1;
}

.drawer-tab.active .tab-badge {
  background: rgba(34, 211, 238, 0.15);
  color: var(--accent-cyan);
}

.drawer-actions {
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.drawer-action-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  width: 24px;
  height: 24px;
  padding: 0;
  border-radius: 4px;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s ease;
}

.drawer-action-btn:hover {
  background: var(--card-bg);
  color: var(--text-main);
}

.drawer-action-btn.active {
  color: var(--accent-cyan);
  background: rgba(0, 229, 255, 0.14);
}

.drawer-action-btn[data-tooltip] {
  position: relative;
}

.drawer-action-btn[data-tooltip]::after {
  content: attr(data-tooltip);
  position: absolute;
  bottom: calc(100% + 6px);
  right: 0;
  padding: 4px 8px;
  background: var(--card-bg);
  color: var(--text-main);
  border: 1px solid var(--card-border);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.25);
  font-size: 0.7rem;
  font-weight: 500;
  white-space: nowrap;
  border-radius: 4px;
  pointer-events: none;
  opacity: 0;
  transform: translateY(4px);
  transition: opacity 0.15s ease, transform 0.15s ease;
  z-index: 100;
}

.drawer-action-btn[data-tooltip]:hover::after {
  opacity: 1;
  transform: translateY(0);
}

.drawer-body {
  overflow-y: auto;
  padding: 0.45rem 0.75rem;
  font-size: 0.75rem;
  line-height: 1.5;
  background: var(--section-bg);
  transition: height 0.2s ease;
}

.empty-logs {
  color: var(--text-muted);
  font-style: italic;
  padding: 0.75rem 0;
  text-align: center;
}

.log-row {
  position: relative;
  display: flex;
  align-items: center;
  gap: 0.6rem;
  padding: 0.18rem 0.45rem;
  border-radius: 4px;
  word-break: break-all;
  transition: background 0.15s ease;
}

.log-row:hover {
  background: rgba(255, 255, 255, 0.05);
}

[data-theme="light"] .log-row:hover {
  background: rgba(0, 0, 0, 0.04);
}

.log-row.copied {
  background: rgba(16, 185, 129, 0.15) !important;
}

.log-row-copy-btn {
  margin-left: auto;
  opacity: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  border-radius: 4px;
  border: 1px solid var(--card-border);
  background: var(--card-bg);
  color: var(--text-muted);
  cursor: pointer;
  flex-shrink: 0;
  transition: all 0.15s ease;
  padding: 0;
}

.log-row:hover .log-row-copy-btn,
.log-row-copy-btn.copied {
  opacity: 0.75;
}

.log-row-copy-btn:hover,
.log-row-copy-btn.copied {
  opacity: 1;
  color: var(--accent-cyan);
  border-color: var(--card-border-active);
  background: var(--input-bg);
}

.log-time {
  color: var(--text-muted);
  font-size: 0.7rem;
  flex-shrink: 0;
}

.log-badge {
  font-size: 0.65rem;
  font-weight: 700;
  padding: 0.05rem 0.3rem;
  border-radius: 3px;
  flex-shrink: 0;
}

.badge-info {
  background: rgba(0, 229, 255, 0.15);
  color: var(--accent-cyan);
}

.badge-success {
  background: rgba(16, 185, 129, 0.15);
  color: var(--success);
}

.badge-warn {
  background: rgba(245, 158, 11, 0.15);
  color: var(--warning);
}

.badge-error {
  background: rgba(239, 68, 68, 0.15);
  color: #ef4444;
}

.log-msg {
  color: var(--text-main);
}

.log-details {
  color: var(--text-muted);
  font-size: 0.7rem;
}
</style>
