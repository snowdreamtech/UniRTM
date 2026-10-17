<template>
  <div v-if="isOpen" class="bottom-drawer" id="app-bottom-drawer">
    <!-- Drawer Header Bar -->
    <div class="drawer-header">
      <!-- Tabs on the left -->
      <div class="drawer-tabs">
        <button
          class="drawer-tab"
          :class="{ active: currentTab === 'terminal' }"
          @click="currentTab = 'terminal'"
        >
          <span style="display: inline-flex; align-items: center; gap: 6px;">
            <TechIcon id="workbench:terminal" :size="13" />
            {{ t("drawer.tab_terminal") || "终端" }}
          </span>
        </button>
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

      <!-- Center: Terminal Shell Selector (Only when in terminal tab) -->
      <div v-if="currentTab === 'terminal'" class="terminal-shell-selector-wrap" @click.stop>
        <button
          class="shell-select-pill"
          @click="isShellMenuOpen = !isShellMenuOpen"
          :title="t('terminal.switch_shell') || '切换终端 Shell'"
        >
          <TechIcon id="workbench:terminal" :size="12" />
          <span class="shell-name font-mono">{{ currentTerminal?.name || 'Shell' }}</span>
          <span v-if="currentTerminal?.is_default" class="shell-default-tag">{{ t('terminal.default_badge') || '系统默认' }}</span>
          <svg class="chevron-icon" :class="{ open: isShellMenuOpen }" viewBox="0 0 16 16" width="10" height="10" fill="currentColor">
            <path d="M4.5 6l3.5 3.5 3.5-3.5" />
          </svg>
        </button>

        <!-- Dropdown Shell List -->
        <div v-if="isShellMenuOpen" class="shell-dropdown-menu">
          <div class="shell-dropdown-header">
            <span>{{ t("terminal.switch_shell") || "切换终端 Shell" }}</span>
            <button class="rescan-btn" @click.stop="handleRescan" :title="t('terminal.rescan') || '重新检测'">
              <svg viewBox="0 0 16 16" width="11" height="11" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M14 8A6 6 0 1 1 9.5 2.5L14 3" />
                <path d="M14 1v2.5H11.5" />
              </svg>
            </button>
          </div>
          <div class="shell-dropdown-list">
            <div
              v-for="term in terminals"
              :key="term.path"
              class="shell-option"
              :class="{ active: term.path === currentTerminal?.path }"
              @click="handleSelectTerminal(term.path)"
            >
              <div class="shell-option-main">
                <span class="shell-option-name font-mono">{{ term.name }}</span>
                <span v-if="term.is_default" class="shell-default-pill">{{ t('terminal.default_badge') || '系统默认' }}</span>
              </div>
              <span class="shell-option-path font-mono">{{ term.path }}</span>
            </div>
            <div v-if="terminals.length === 0" class="shell-dropdown-empty">
              <span>{{ t('terminal.rescan') || '正在扫描可用 Shell...' }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Action Tools on the right -->
      <div class="drawer-actions">
        <!-- Terminal Special Action Buttons (VSCode Style) -->
        <template v-if="currentTab === 'terminal'">
          <!-- 1. New Terminal Button Group (+) with Shell Dropdown -->
          <div class="terminal-action-group">
            <button
              class="drawer-action-btn term-action-btn"
              :title="t('terminal.new_terminal') || '新建终端'"
              :data-tooltip="t('terminal.new_terminal') || '新建终端'"
              @click="handleCreateDefaultSession"
            >
              <svg viewBox="0 0 16 16" width="13" height="13" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M8 3v10M3 8h10" />
              </svg>
            </button>
            <button
              class="drawer-action-btn term-action-arrow"
              :class="{ active: isNewMenuOpen }"
              title="选择 Shell 新建终端"
              @click.stop="isNewMenuOpen = !isNewMenuOpen"
            >
              <svg viewBox="0 0 16 16" width="9" height="9" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M4 6l4 4 4-4" />
              </svg>
            </button>

            <!-- Dropdown to select Shell for new session -->
            <div v-if="isNewMenuOpen" class="term-new-dropdown font-mono" @click.stop>
              <div class="dropdown-header">{{ t('terminal.new_terminal') || '新建终端' }}</div>
              <div
                v-for="term in terminals"
                :key="term.path"
                class="shell-option"
                @click="handleCreateSessionWithShell(term)"
              >
                <div class="shell-option-main">
                  <span class="shell-option-name font-mono">{{ term.name }}</span>
                  <span v-if="term.is_default" class="shell-default-pill">{{ t('terminal.default_badge') || '系统默认' }}</span>
                </div>
                <span class="shell-option-path font-mono">{{ term.path }}</span>
              </div>
            </div>
          </div>

          <!-- 2. Vertical Split Button [ | ] (一分为二) -->
          <button
            class="drawer-action-btn"
            :title="t('terminal.split_vertical') || '竖直拆分终端 (一分为二)'"
            :data-tooltip="t('terminal.split_vertical') || '竖直拆分终端 (一分为二)'"
            @click="handleSplitVertical"
          >
            <svg viewBox="0 0 16 16" width="13" height="13" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
              <rect x="2" y="2.5" width="12" height="11" rx="1.5" />
              <path d="M8 2.5v11" />
            </svg>
          </button>

          <!-- 3. Horizontal Split Button [ — ] (一分为二) -->
          <button
            class="drawer-action-btn"
            :title="t('terminal.split_horizontal') || '横向拆分终端 (一分为二)'"
            :data-tooltip="t('terminal.split_horizontal') || '横向拆分终端 (一分为二)'"
            @click="handleSplitHorizontal"
          >
            <svg viewBox="0 0 16 16" width="13" height="13" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
              <rect x="2.5" y="2" width="11" height="12" rx="1.5" />
              <path d="M2.5 8h11" />
            </svg>
          </button>

          <!-- 4. Kill Current Terminal (Trash) -->
          <button
            class="drawer-action-btn term-kill-btn"
            :title="t('terminal.kill_terminal') || '关闭当前终端'"
            :data-tooltip="t('terminal.kill_terminal') || '关闭当前终端'"
            @click="handleKillCurrentTerminal"
          >
            <svg viewBox="0 0 16 16" width="13" height="13" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
              <path d="M3 4.5h10M6 4.5V3a1 1 0 0 1 1-1h2a1 1 0 0 1 1 1v1.5M12 4.5v8.5a1.5 1.5 0 0 1-1.5 1.5h-5A1.5 1.5 0 0 1 4 13V4.5" />
              <path d="M6.5 7.5v4M9.5 7.5v4" />
            </svg>
          </button>
        </template>

        <!-- General Action Buttons -->
        <button
          class="drawer-action-btn"
          :title="clearLogsTitle"
          :data-tooltip="clearLogsTitle"
          @click="handleClearCurrent"
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
          v-if="currentTab !== 'terminal'"
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

    <!-- 1. Terminal View Body with VSCode Split Panes & Tabs Sidebar -->
    <div
      v-if="currentTab === 'terminal'"
      class="terminal-wrapper font-mono"
      :style="{ height: isMaximized ? '420px' : '200px' }"
    >
      <!-- Main: Terminal Panes Work Area -->
      <div
        class="terminal-panes-area"
        :class="['split-' + (activeSession?.splitDirection || 'none')]"
      >
        <div
          v-for="pane in activeSession?.panes"
          :key="pane.id"
          class="terminal-pane"
          :class="{ active: pane.id === activeSession?.activePaneId }"
          @click="setActivePane(pane.id)"
        >
          <!-- Pane Top Header / Status Banner -->
          <div class="terminal-pane-banner">
            <div class="pane-banner-left">
              <span class="term-status-dot"></span>
              <span class="term-shell-info font-mono">{{ pane.shellName }}</span>
              <span class="term-shell-path font-mono">({{ pane.shellPath }})</span>
              <span v-if="activeSession?.splitDirection !== 'none'" class="term-split-pill">
                {{ activeSession?.splitDirection === 'vertical' ? '竖直拆分' : '横向拆分' }}
              </span>
            </div>
            <div class="pane-banner-right">
              <button
                v-if="(activeSession?.panes.length || 0) > 1"
                class="pane-close-btn"
                :title="t('terminal.kill_terminal') || '关闭此窗格'"
                @click.stop="closePane(pane.id)"
              >
                <svg viewBox="0 0 16 16" width="11" height="11" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M3.5 3.5l9 9M12.5 3.5l-9 9" />
                </svg>
              </button>
            </div>
          </div>

          <!-- Pane Screen Output and Interactive Input -->
          <div class="terminal-screen" :ref="el => setPaneRef(pane.id, el)">
            <div class="terminal-welcome">
              <div class="term-text-cyan">UniRTM Workbench Terminal ({{ pane.shellName }})</div>
              <div class="term-text-muted">{{ t('drawer.empty_terminal') || '集成终端准备就绪。在上方选择 Shell 或通过快捷指令运行任务。' }}</div>
            </div>

            <div v-for="(line, idx) in pane.lines" :key="idx" class="terminal-line" :class="'line-' + line.type">
              <span v-if="line.type === 'input'" class="term-prompt">❯</span>
              <span class="term-content">{{ line.text }}</span>
            </div>

            <!-- Terminal Interactive Input Row -->
            <div class="terminal-input-row">
              <span class="term-prompt">❯</span>
              <input
                v-model="pane.input"
                @keydown.enter="handlePaneRunCommand(pane)"
                @keydown.up.prevent="handlePaneHistoryUp(pane)"
                @keydown.down.prevent="handlePaneHistoryDown(pane)"
                class="terminal-cmd-input font-mono"
                :placeholder="t('drawer.empty_terminal') || '输入命令并回车运行 (如 unirtm --help)...'"
                spellcheck="false"
                @focus="setActivePane(pane.id)"
              />
            </div>
          </div>
        </div>
      </div>

      <!-- Right: VSCode-style Terminal Tabs Sidebar -->
      <div class="terminal-sidebar">
        <div class="terminal-sidebar-header">
          <span class="sidebar-title">{{ t('terminal.tabs_title') || '终端' }}</span>
          <button
            class="sidebar-add-btn"
            :title="t('terminal.new_terminal') || '新建终端'"
            @click="handleCreateDefaultSession"
          >
            <svg viewBox="0 0 16 16" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M8 3v10M3 8h10" />
            </svg>
          </button>
        </div>
        <div class="terminal-tabs-list">
          <div
            v-for="s in sessions"
            :key="s.id"
            class="terminal-tab-item"
            :class="{ active: s.id === activeSessionId }"
            @click="setActiveSession(s.id)"
          >
            <div class="tab-item-icon">
              <svg viewBox="0 0 16 16" width="13" height="13" fill="none" stroke="currentColor" stroke-width="1.8">
                <path d="M3.5 5.5l3.5 2.5-3.5 2.5M8.5 10.5h4" />
              </svg>
            </div>
            <div class="tab-item-info">
              <span class="tab-item-name font-mono">{{ s.title }}</span>
              <span v-if="s.splitDirection !== 'none'" class="tab-item-split-badge">
                {{ s.splitDirection === 'vertical' ? '[|]' : '[—]' }}
              </span>
            </div>
            <button
              class="tab-item-close-btn"
              :title="t('terminal.kill_terminal') || '关闭终端'"
              @click.stop="closeSession(s.id)"
            >
              <svg viewBox="0 0 16 16" width="10" height="10" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M3.5 3.5l9 9M12.5 3.5l-9 9" />
              </svg>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 2. Logs Drawer Body (Output, Tasks, Doctor) -->
    <div
      v-else
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
import { ref, computed, watch, nextTick, onMounted, onUnmounted } from "vue";
import { t } from "../../i18n";
import TechIcon from "../common/TechIcon.vue";
import { useToast } from "../../composables/useToast";
import { useTerminal, TerminalItem } from "../../composables/useTerminal";
import { useTerminalSessions, TerminalPane } from "../../composables/useTerminalSessions";

export interface LogItem {
  timestamp: string;
  level: "INFO" | "SUCCESS" | "WARN" | "ERROR";
  message: string;
  details?: string;
  category?: "output" | "tasks" | "doctor";
}

export type DrawerTab = "terminal" | "output" | "tasks" | "doctor";

const props = withDefaults(
  defineProps<{
    isOpen: boolean;
    logs: LogItem[];
    activeTab?: DrawerTab;
  }>(),
  {
    activeTab: undefined,
  }
);

const emit = defineEmits<{
  (e: "close"): void;
  (e: "clear-logs", category?: DrawerTab): void;
  (e: "update:activeTab", tab: DrawerTab): void;
}>();

const { showToast } = useToast();
const {
  terminals,
  currentTerminal,
  fetchTerminals,
  setPreferredTerminal,
} = useTerminal();

const {
  sessions,
  activeSessionId,
  activeSession,
  activePane,
  createSession,
  splitActiveSession,
  setActiveSession,
  setActivePane,
  closePane,
  closeSession,
  closeActiveTerminal,
  clearActivePane,
  ensureInitialSession,
} = useTerminalSessions();

const isShellMenuOpen = ref(false);
const isNewMenuOpen = ref(false);
const paneRefs = new Map<string, HTMLElement>();

const DRAWER_TAB_KEY = "unirtm_drawer_tab";
const DRAWER_MAX_KEY = "unirtm_drawer_maximized";
const DRAWER_AUTOSCROLL_KEY = "unirtm_drawer_autoscroll";

function getInitialTab(): DrawerTab {
  if (typeof localStorage === "undefined") return "output";
  const saved = localStorage.getItem(DRAWER_TAB_KEY);
  if (saved === "terminal" || saved === "output" || saved === "tasks" || saved === "doctor") {
    return saved as DrawerTab;
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

const internalActiveTab = ref<DrawerTab>(getInitialTab());

const currentTab = computed<DrawerTab>({
  get: () => props.activeTab ?? internalActiveTab.value,
  set: (val: DrawerTab) => {
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
  isShellMenuOpen.value = false;
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
  return (
    t("drawer.empty_logs") ||
    "暂无系统输出日志。系统运行与扩展事件将在此显示。"
  );
});

const clearLogsTitle = computed(() => t("drawer.clear_logs") || "清空当前面板");
const copyAllTitle = computed(() => t("drawer.copy_all") || "复制全部输出");
const autoScrollTitle = computed(() =>
  autoScroll.value ? "自动滚动 (开启)" : "自动滚动 (关闭)"
);
const maximizeTitle = computed(() =>
  isMaximized.value ? "还原面板高度" : "最大化面板高度"
);
const closeTitle = computed(() => "关闭面板 (Ctrl+` 或 Esc)");

const isCopiedAll = ref(false);
const copiedRowIdx = ref<number | null>(null);

function formatLogLine(log: LogItem): string {
  const parts = [`[${log.timestamp}]`, `[${log.level}]`, log.message];
  if (log.details) {
    parts.push(`\n${log.details}`);
  }
  return parts.join(" ");
}

async function copyText(text: string): Promise<boolean> {
  try {
    if (typeof navigator !== "undefined" && navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(text);
      return true;
    }
  } catch (e) {
    // fallback
  }
  try {
    if (typeof document !== "undefined") {
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
    }
    return false;
  } catch (e) {
    return false;
  }
}

async function handleCopyAll() {
  if (currentTab.value === "terminal") {
    if (!activePane.value || activePane.value.lines.length === 0) {
      showToast(t("drawer.empty_copy") || "当前暂无日志可复制", "warning");
      return;
    }
    const text = activePane.value.lines.map((l) => l.text).join("\n");
    const success = await copyText(text);
    if (success) {
      isCopiedAll.value = true;
      setTimeout(() => {
        isCopiedAll.value = false;
      }, 1500);
      showToast(t("drawer.copy_all_success") || "已复制当前面板全部日志到剪贴板", "success");
    }
    return;
  }

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

function handleClearCurrent() {
  if (currentTab.value === "terminal") {
    clearActivePane();
  }
  emit("clear-logs", currentTab.value);
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

function setPaneRef(id: string, el: any) {
  if (el) {
    paneRefs.set(id, el as HTMLElement);
  } else {
    paneRefs.delete(id);
  }
}

function scrollToPaneBottom(id: string) {
  const el = paneRefs.get(id);
  if (el) {
    el.scrollTop = el.scrollHeight;
  }
}

async function handlePaneRunCommand(pane: TerminalPane) {
  const cmd = pane.input.trim();
  if (!cmd) return;

  pane.lines.push({ type: "input", text: cmd });
  pane.history.push(cmd);
  pane.historyIndex = -1;
  pane.input = "";

  if (cmd === "clear" || cmd === "cls") {
    pane.lines = [];
    return;
  }

  if (cmd === "exit") {
    closePane(pane.id);
    return;
  }

  try {
    const res = await fetch("/api/tasks/run", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name: cmd, scope: "workspace" }),
    });
    if (res.ok) {
      const data = await res.json();
      if (data.output) {
        pane.lines.push({ type: "output", text: data.output });
      } else {
        pane.lines.push({
          type: "info",
          text: `[${pane.shellName}] Command executed (exit code: 0).`,
        });
      }
    } else {
      pane.lines.push({ type: "error", text: "Command returned non-zero exit code." });
    }
  } catch (err: any) {
    pane.lines.push({ type: "error", text: err.message || "Command execution error" });
  }

  nextTick(() => {
    scrollToPaneBottom(pane.id);
  });
}

function handlePaneHistoryUp(pane: TerminalPane) {
  if (pane.history.length === 0) return;
  if (pane.historyIndex === -1) {
    pane.historyIndex = pane.history.length - 1;
  } else if (pane.historyIndex > 0) {
    pane.historyIndex--;
  }
  pane.input = pane.history[pane.historyIndex] || "";
}

function handlePaneHistoryDown(pane: TerminalPane) {
  if (pane.historyIndex === -1) return;
  if (pane.historyIndex < pane.history.length - 1) {
    pane.historyIndex++;
    pane.input = pane.history[pane.historyIndex] || "";
  } else {
    pane.historyIndex = -1;
    pane.input = "";
  }
}

function handleCreateDefaultSession() {
  const s = createSession(
    currentTerminal.value
      ? { name: currentTerminal.value.name, path: currentTerminal.value.path }
      : undefined
  );
  showToast(`${t("terminal.new_terminal") || "新建终端"}: ${s.title}`, "info");
}

function handleCreateSessionWithShell(term: TerminalItem) {
  isNewMenuOpen.value = false;
  const s = createSession({ name: term.name, path: term.path });
  showToast(`${t("terminal.new_terminal") || "新建终端"}: ${s.title}`, "info");
}

function handleSplitVertical() {
  splitActiveSession("vertical");
  showToast(t("terminal.split_vertical") || "竖直拆分终端 (一分为二)", "info");
}

function handleSplitHorizontal() {
  splitActiveSession("horizontal");
  showToast(t("terminal.split_horizontal") || "横向拆分终端 (一分为二)", "info");
}

function handleKillCurrentTerminal() {
  closeActiveTerminal();
}

async function handleSelectTerminal(path: string) {
  isShellMenuOpen.value = false;
  await setPreferredTerminal(path);
  if (activePane.value) {
    const term = terminals.value.find((t) => t.path === path);
    activePane.value.shellPath = path;
    activePane.value.shellName = term?.name || path;
    activePane.value.lines.push({
      type: "info",
      text: `[UniRTM] Active shell switched to: ${activePane.value.shellName} (${path})`,
    });
  }
}

async function handleRescan() {
  await fetchTerminals();
  showToast(t("terminal.rescan") || "已重新检测终端环境", "info");
}

function handleGlobalClick() {
  isShellMenuOpen.value = false;
  isNewMenuOpen.value = false;
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

onMounted(() => {
  fetchTerminals();
  ensureInitialSession(currentTerminal.value);
  window.addEventListener("click", handleGlobalClick);
});

onUnmounted(() => {
  window.removeEventListener("click", handleGlobalClick);
});
</script>

<style scoped>
.bottom-drawer {
  width: 100%;
  flex-shrink: 0;
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
  position: relative;
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

/* Shell Selector Pill & Dropdown */
.terminal-shell-selector-wrap {
  position: relative;
  display: flex;
  align-items: center;
  margin-left: 0.5rem;
}

.shell-select-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 2px 8px;
  border-radius: 4px;
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  color: var(--text-main);
  font-size: 0.72rem;
  cursor: pointer;
  transition: all 0.15s ease;
}

.shell-select-pill:hover {
  border-color: var(--accent-cyan);
  background: var(--btn-hover, rgba(0, 229, 255, 0.05));
}

.shell-name {
  font-weight: 600;
  color: var(--accent-cyan);
}

.shell-default-tag {
  font-size: 0.65rem;
  padding: 0 4px;
  border-radius: 3px;
  background: rgba(16, 185, 129, 0.15);
  color: #10b981;
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.chevron-icon {
  transition: transform 0.2s ease;
  color: var(--text-muted);
}

.chevron-icon.open {
  transform: rotate(180deg);
}

.shell-dropdown-menu {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  min-width: 260px;
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  border-radius: 6px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.35);
  z-index: 120;
  overflow: hidden;
}

.shell-dropdown-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 10px;
  font-size: 0.7rem;
  font-weight: 600;
  color: var(--text-muted);
  border-bottom: 1px solid var(--card-border);
  background: var(--input-bg);
}

.rescan-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 2px 4px;
  border-radius: 3px;
  display: inline-flex;
  align-items: center;
}

.rescan-btn:hover {
  color: var(--accent-cyan);
}

.shell-dropdown-list {
  max-height: 220px;
  overflow-y: auto;
  padding: 4px;
}

.shell-option {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 5px 8px;
  border-radius: 4px;
  cursor: pointer;
  transition: background 0.15s ease;
}

.shell-option:hover {
  background: rgba(0, 229, 255, 0.08);
}

.shell-option.active {
  background: rgba(0, 229, 255, 0.14);
}

.shell-option-main {
  display: flex;
  align-items: center;
  gap: 6px;
}

.shell-option-name {
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--text-main);
}

.shell-option.active .shell-option-name {
  color: var(--accent-cyan);
}

.shell-default-pill {
  font-size: 0.62rem;
  padding: 0 4px;
  border-radius: 3px;
  background: rgba(16, 185, 129, 0.15);
  color: #10b981;
}

.shell-option-path {
  font-size: 0.65rem;
  color: var(--text-muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.shell-dropdown-empty {
  padding: 10px;
  font-size: 0.72rem;
  color: var(--text-muted);
  text-align: center;
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

/* Terminal View Styles (VSCode Multi-Terminal & Split Panes) */
.terminal-action-group {
  position: relative;
  display: inline-flex;
  align-items: center;
}

.term-action-btn {
  border-top-right-radius: 0 !important;
  border-bottom-right-radius: 0 !important;
}

.term-action-arrow {
  width: 14px !important;
  border-top-left-radius: 0 !important;
  border-bottom-left-radius: 0 !important;
  margin-left: -2px;
  opacity: 0.8;
}

.term-action-arrow:hover,
.term-action-arrow.active {
  opacity: 1;
  background: var(--card-bg);
  color: var(--accent-cyan);
}

.term-new-dropdown {
  position: absolute;
  top: calc(100% + 4px);
  right: 0;
  width: 220px;
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.45);
  border-radius: 6px;
  padding: 4px;
  z-index: 120;
}

.dropdown-header {
  padding: 4px 8px;
  font-size: 0.68rem;
  font-weight: 700;
  color: var(--text-muted);
  border-bottom: 1px solid var(--card-border);
  margin-bottom: 4px;
}

.term-kill-btn:hover {
  color: #ef4444 !important;
  background: rgba(239, 68, 68, 0.12) !important;
}

.terminal-wrapper {
  display: flex;
  flex-direction: row;
  overflow: hidden;
  background: var(--section-bg, #0b0f19);
  transition: height 0.2s ease;
}

.terminal-panes-area {
  flex: 1;
  min-width: 0;
  display: flex;
  height: 100%;
  overflow: hidden;
}

.terminal-panes-area.split-vertical {
  flex-direction: row;
}

.terminal-panes-area.split-horizontal {
  flex-direction: column;
}

.terminal-panes-area.split-none {
  flex-direction: row;
}

.terminal-pane {
  flex: 1 1 50%;
  min-width: 0;
  min-height: 0;
  display: flex;
  flex-direction: column;
  position: relative;
  background: var(--section-bg, #0b0f19);
  box-sizing: border-box;
}

.terminal-panes-area.split-vertical .terminal-pane:not(:first-child) {
  border-left: 1px solid var(--card-border);
}

.terminal-panes-area.split-horizontal .terminal-pane:not(:first-child) {
  border-top: 1px solid var(--card-border);
}

.terminal-pane.active {
  outline: 1.5px solid rgba(0, 229, 255, 0.6);
  outline-offset: -1px;
  z-index: 2;
}

.terminal-pane-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.45rem;
  padding: 3px 0.65rem;
  background: rgba(0, 0, 0, 0.2);
  border-bottom: 1px solid var(--card-border);
  font-size: 0.68rem;
  color: var(--text-muted);
  flex-shrink: 0;
}

.pane-banner-left {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.term-shell-path {
  font-size: 0.62rem;
  opacity: 0.7;
}

.term-split-pill {
  font-size: 0.6rem;
  padding: 0 4px;
  border-radius: 3px;
  background: rgba(0, 229, 255, 0.15);
  color: var(--accent-cyan);
}

.pane-close-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 2px;
  border-radius: 3px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s ease;
}

.pane-close-btn:hover {
  color: #ef4444;
  background: rgba(239, 68, 68, 0.15);
}

/* VSCode-style Terminal Tabs Sidebar */
.terminal-sidebar {
  width: 165px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  background: rgba(0, 0, 0, 0.15);
  border-left: 1px solid var(--card-border);
  overflow: hidden;
}

.terminal-sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 4px 8px;
  font-size: 0.68rem;
  font-weight: 700;
  color: var(--text-muted);
  border-bottom: 1px solid var(--card-border);
  background: rgba(0, 0, 0, 0.1);
}

.sidebar-add-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 2px 4px;
  border-radius: 3px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.sidebar-add-btn:hover {
  color: var(--accent-cyan);
  background: var(--card-bg);
}

.terminal-tabs-list {
  flex: 1;
  overflow-y: auto;
  padding: 4px;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.terminal-tab-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 6px;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.15s ease;
  position: relative;
}

.terminal-tab-item:hover {
  background: rgba(255, 255, 255, 0.05);
}

.terminal-tab-item.active {
  background: rgba(0, 229, 255, 0.12);
  color: var(--accent-cyan);
}

.tab-item-icon {
  display: inline-flex;
  align-items: center;
  color: var(--text-muted);
  flex-shrink: 0;
}

.terminal-tab-item.active .tab-item-icon {
  color: var(--accent-cyan);
}

.tab-item-info {
  flex: 1;
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 4px;
}

.tab-item-name {
  font-size: 0.72rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tab-item-split-badge {
  font-size: 0.62rem;
  opacity: 0.75;
  color: var(--accent-cyan);
}

.tab-item-close-btn {
  opacity: 0;
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 2px;
  border-radius: 3px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s ease;
  flex-shrink: 0;
}

.terminal-tab-item:hover .tab-item-close-btn {
  opacity: 0.8;
}

.tab-item-close-btn:hover {
  opacity: 1 !important;
  color: #ef4444;
  background: rgba(239, 68, 68, 0.15);
}

.term-status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #10b981;
}

.term-shell-info {
  font-weight: 500;
}

.terminal-screen {
  flex: 1;
  overflow-y: auto;
  padding: 0.5rem 0.65rem;
  font-size: 0.75rem;
  line-height: 1.5;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.terminal-welcome {
  margin-bottom: 0.35rem;
}

.term-text-cyan {
  color: var(--accent-cyan);
  font-weight: 600;
}

.term-text-muted {
  color: var(--text-muted);
  font-size: 0.7rem;
}

.terminal-line {
  display: flex;
  align-items: flex-start;
  gap: 6px;
  word-break: break-all;
}

.term-prompt {
  color: var(--accent-cyan);
  font-weight: 700;
  user-select: none;
  flex-shrink: 0;
}

.line-input {
  color: var(--text-main);
  font-weight: 600;
}

.line-output {
  color: var(--text-main);
  white-space: pre-wrap;
}

.line-info {
  color: var(--accent-cyan);
  font-style: italic;
}

.line-error {
  color: #ef4444;
}

.terminal-input-row {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 2px;
}

.terminal-cmd-input {
  flex: 1;
  background: transparent;
  border: none;
  outline: none;
  color: var(--text-main);
  font-size: 0.75rem;
  caret-color: var(--accent-cyan);
  padding: 0;
}

.terminal-cmd-input::placeholder {
  color: var(--text-muted);
  opacity: 0.5;
}

/* Classic Log View Styles */
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
