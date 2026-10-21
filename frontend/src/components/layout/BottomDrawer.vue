<template>
  <div v-show="isOpen" class="bottom-drawer" :class="{ 'is-resizing': isDrawerHeightDragging }" id="app-bottom-drawer">
    <!-- Top Horizontal Sash for vertical drag-resizing (VS Code style) -->
    <div
      class="drawer-top-resizer-sash"
      :class="{ active: isDrawerHeightDragging }"
      @pointerdown="handleDrawerHeightPointerDown"
      @dblclick="handleToggleMaximizeOrReset"
      title="双击切换最大化，上下拖拽调节高度"
    >
      <div class="sash-indicator-horizontal"></div>
    </div>

    <!-- Drawer Header Bar -->
    <div class="drawer-header">
      <!-- Tabs on the left (draggable with VS Code drop indicator) -->
      <div class="drawer-tabs" @contextmenu.prevent="handleTabContextMenu($event)">
        <button
          v-for="tabId in visibleTabs"
          :key="tabId"
          class="drawer-tab"
          :class="{
            active: currentTab === tabId,
            'is-dragging': draggedTabId === tabId,
            'drop-indicator-left': dropTargetTabId === tabId && dropTargetPosition === 'before',
            'drop-indicator-right': dropTargetTabId === tabId && dropTargetPosition === 'after',
          }"
          draggable="true"
          :title="t('drawer.drag_reorder_tip') || '拖拽调整标签页顺序，右键重置'"
          @dragstart="handleTabDragStart($event, tabId)"
          @dragover="handleTabDragOver($event, tabId)"
          @dragleave="handleTabDragLeave($event, tabId)"
          @drop="handleTabDrop($event, tabId)"
          @dragend="handleTabDragEnd"
          @click="handleTabClick(tabId)"
        >
          <span class="tab-label-inner">
            <TechIcon :id="tabDefinitions[tabId].icon" :size="13" />
            {{ tabDefinitions[tabId].label }}
          </span>
          <span v-if="tabDefinitions[tabId].badge && tabDefinitions[tabId].badge > 0" class="tab-badge">
            {{ tabDefinitions[tabId].badge }}
          </span>
        </button>
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
              <svg
                viewBox="0 0 16 16"
                width="13"
                height="13"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
              >
                <path d="M8 3v10M3 8h10" />
              </svg>
            </button>
            <button
              class="drawer-action-btn term-action-arrow"
              :class="{ active: isNewMenuOpen }"
              title="选择 Shell 新建终端"
              @click.stop="isNewMenuOpen = !isNewMenuOpen"
            >
              <svg
                viewBox="0 0 16 16"
                width="9"
                height="9"
                fill="none"
                stroke="currentColor"
                stroke-width="2.2"
                stroke-linecap="round"
                stroke-linejoin="round"
              >
                <path d="M4 6l4 4 4-4" />
              </svg>
            </button>

            <!-- Dropdown to select Shell for new session -->
            <div v-if="isNewMenuOpen" class="term-new-dropdown font-mono" @click.stop>
              <div class="dropdown-header">{{ t("terminal.new_terminal") || "新建终端" }}</div>
              <div
                v-for="term in terminals"
                :key="term.path"
                class="shell-option"
                @click="handleCreateSessionWithShell(term)"
              >
                <div class="shell-option-main">
                  <span class="shell-option-name font-mono">{{ term.name }}</span>
                  <span v-if="term.is_default" class="shell-default-pill">{{
                    t("terminal.default_badge") || "系统默认"
                  }}</span>
                </div>
                <span class="shell-option-path font-mono">{{ term.path }}</span>
              </div>
            </div>
          </div>
        </template>

        <!-- "更多" (More Actions) Dropdown Menu (复制日志, 清空日志, 关闭终端) -->
        <div class="drawer-more-wrapper">
          <button
            class="drawer-action-btn"
            :class="{ active: isMoreMenuOpen }"
            title="更多操作"
            data-tooltip="更多操作"
            @click.stop="isMoreMenuOpen = !isMoreMenuOpen"
          >
            <svg viewBox="0 0 16 16" width="14" height="14" fill="currentColor">
              <circle cx="3" cy="8" r="1.5" />
              <circle cx="8" cy="8" r="1.5" />
              <circle cx="13" cy="8" r="1.5" />
            </svg>
          </button>

          <!-- Dropdown menu with icons + text in exact required order -->
          <div v-if="isMoreMenuOpen" class="drawer-more-dropdown font-mono" @click.stop>
            <!-- 1. 复制日志 -->
            <button class="more-menu-item" @click="handleMenuCopyLogs">
              <svg
                viewBox="0 0 16 16"
                width="13"
                height="13"
                fill="none"
                stroke="currentColor"
                stroke-width="1.8"
                stroke-linecap="round"
                stroke-linejoin="round"
              >
                <rect x="5.5" y="5.5" width="8" height="8" rx="1.5" />
                <path d="M10.5 5.5V3.5a1.5 1.5 0 0 0-1.5-1.5h-5A1.5 1.5 0 0 0 2.5 3.5v5a1.5 1.5 0 0 0 1.5 1.5h1.5" />
              </svg>
              <span>{{ t("drawer.copy_all") || "复制日志" }}</span>
            </button>

            <!-- 2. 清空日志 -->
            <button class="more-menu-item" @click="handleMenuClearLogs">
              <svg
                viewBox="0 0 16 16"
                width="13"
                height="13"
                fill="none"
                stroke="currentColor"
                stroke-width="1.8"
                stroke-linecap="round"
                stroke-linejoin="round"
              >
                <circle cx="8" cy="8" r="6" />
                <path d="M3.8 3.8l8.4 8.4" />
              </svg>
              <span>{{ t("drawer.clear_logs") || "清空日志" }}</span>
            </button>

            <!-- 3. 关闭终端 (在终端 Tab 下提供) -->
            <button v-if="currentTab === 'terminal'" class="more-menu-item danger" @click="handleMenuKillTerminal">
              <svg
                viewBox="0 0 16 16"
                width="13"
                height="13"
                fill="none"
                stroke="currentColor"
                stroke-width="1.8"
                stroke-linecap="round"
                stroke-linejoin="round"
              >
                <path
                  d="M3 4.5h10M6 4.5V3a1 1 0 0 1 1-1h2a1 1 0 0 1 1 1v1.5M12 4.5v8.5a1.5 1.5 0 0 1-1.5 1.5h-5A1.5 1.5 0 0 1 4 13V4.5"
                />
                <path d="M6.5 7.5v4M9.5 7.5v4" />
              </svg>
              <span>{{ t("terminal.kill_terminal") || "关闭终端" }}</span>
            </button>
          </div>
        </div>
        <button
          v-if="currentTab === 'output' || currentTab === 'tasks' || currentTab === 'doctor'"
          class="drawer-action-btn"
          :class="{ active: autoScroll }"
          :title="autoScrollTitle"
          :data-tooltip="autoScrollTitle"
          @click="autoScroll = !autoScroll"
        >
          <svg
            viewBox="0 0 16 16"
            width="13"
            height="13"
            fill="none"
            stroke="currentColor"
            stroke-width="1.8"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
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
          <svg
            v-if="!isMaximized"
            viewBox="0 0 16 16"
            width="13"
            height="13"
            fill="none"
            stroke="currentColor"
            stroke-width="2.2"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <path d="M2.5 10.5L8 5l5.5 5.5" />
          </svg>
          <svg
            v-else
            viewBox="0 0 16 16"
            width="13"
            height="13"
            fill="none"
            stroke="currentColor"
            stroke-width="2.2"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <path d="M2.5 5.5L8 11l5.5-5.5" />
          </svg>
        </button>
        <button class="drawer-action-btn" :title="closeTitle" :data-tooltip="closeTitle" @click="$emit('close')">
          <svg
            viewBox="0 0 16 16"
            width="13"
            height="13"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <path d="M3.5 3.5l9 9M12.5 3.5l-9 9" />
          </svg>
        </button>
      </div>
    </div>

    <!-- 1. Terminal View Body with VSCode Split Panes & Tabs Sidebar -->
    <div
      v-show="currentTab === 'terminal'"
      class="terminal-wrapper font-mono"
      :class="{ 'is-resizing': isTermSidebarDragging || isDrawerHeightDragging }"
      :style="{ height: `${effectiveBodyHeight}px` }"
    >
      <!-- Main: Terminal Panes Work Area -->
      <div class="terminal-panes-area" :class="['split-' + (activeSession?.splitDirection || 'none')]">
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
              <span
                v-if="getActiveProcess(pane.id)"
                class="term-process-badge"
                :style="{
                  backgroundColor: `${getActiveProcess(pane.id)?.color}25`,
                  color: getActiveProcess(pane.id)?.color,
                  borderColor: `${getActiveProcess(pane.id)?.color}60`,
                }"
                :title="`Running: ${getActiveProcess(pane.id)?.name}`"
              >
                {{ getActiveProcess(pane.id)?.name }}
              </span>
              <span class="term-shell-info font-mono">{{ pane.shellName }}</span>
              <span class="term-shell-path font-mono">({{ pane.shellPath }})</span>
              <span v-if="activeSession?.splitDirection !== 'none'" class="term-split-pill">
                {{ activeSession?.splitDirection === "vertical" ? "竖直拆分" : "横向拆分" }}
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

          <!-- Pane Screen Output and Interactive xterm.js PTY Canvas -->
          <div class="terminal-screen">
            <XTerminalPane
              :pane-id="pane.id"
              :shell-name="pane.shellName"
              :shell-path="pane.shellPath"
              :is-active="pane.id === activeSession?.activePaneId"
              @focus="setActivePane(pane.id)"
            />
          </div>
        </div>
      </div>

      <!-- Resizer Sash between terminal panes and terminal tabs sidebar -->
      <div
        class="terminal-sidebar-sash"
        :class="{ active: isTermSidebarDragging }"
        @pointerdown="handleTermSidebarPointerDown"
        @dblclick="resetTermSidebarWidth"
        title="双击恢复默认宽度，拖拽调节大小"
      >
        <div class="sash-indicator"></div>
      </div>

      <!-- Right: VSCode-style Terminal Tabs Sidebar -->
      <div class="terminal-sidebar" :style="{ width: `${termSidebarWidth}px` }">
        <div class="terminal-sidebar-header">
          <span class="sidebar-title">{{ t("terminal.tabs_title") || "终端" }}</span>
          <button
            class="sidebar-add-btn"
            :title="t('terminal.new_terminal') || '新建终端'"
            @click="handleCreateDefaultSession"
          >
            <svg
              viewBox="0 0 16 16"
              width="14"
              height="14"
              fill="none"
              stroke="currentColor"
              stroke-width="2.2"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
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
            @dblclick.stop="startRenameSession(s.id, getEffectiveTitle(s.id, s.title))"
          >
            <div class="tab-item-icon">
              <span
                v-if="getActiveProcess(s.activePaneId)"
                class="tab-process-indicator font-mono"
                :style="{ color: getActiveProcess(s.activePaneId)?.color }"
                :title="getActiveProcess(s.activePaneId)?.name"
              >
                ●
              </span>
              <svg
                v-else
                viewBox="0 0 16 16"
                width="13"
                height="13"
                fill="none"
                stroke="currentColor"
                stroke-width="1.8"
              >
                <path d="M3.5 5.5l3.5 2.5-3.5 2.5M8.5 10.5h4" />
              </svg>
            </div>
            <div class="tab-item-info">
              <input
                v-if="editingSessionId === s.id"
                :class="`tab-rename-input tab-rename-input-${s.id}`"
                v-model="editingTitleInput"
                @blur="finishRenameSession(s.id)"
                @keydown.enter.prevent="finishRenameSession(s.id)"
                @keydown.esc.prevent="cancelRenameSession"
                @click.stop
              />
              <span v-else class="tab-item-name font-mono" :title="getEffectiveTitle(s.id, s.title)">
                {{ getEffectiveTitle(s.id, s.title) }}
              </span>
              <span v-if="s.splitDirection !== 'none'" class="tab-item-split-badge">
                {{ s.splitDirection === "vertical" ? "[|]" : "[—]" }}
              </span>
            </div>
            <!-- Split actions: on terminal tab before trash icon, active only -->
            <div v-if="s.id === activeSessionId" class="tab-item-split-actions">
              <button
                class="tab-action-btn"
                :title="t('terminal.split_vertical') || '竖直拆分终端 (一分为二)'"
                @click.stop="handleSplitVertical"
              >
                <svg
                  viewBox="0 0 16 16"
                  width="12"
                  height="12"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="1.8"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                >
                  <rect x="2" y="2.5" width="12" height="11" rx="1.5" />
                  <path d="M8 2.5v11" />
                </svg>
              </button>
              <button
                class="tab-action-btn"
                :title="t('terminal.split_horizontal') || '水平拆分终端 (一分为二)'"
                @click.stop="handleSplitHorizontal"
              >
                <svg
                  viewBox="0 0 16 16"
                  width="12"
                  height="12"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="1.8"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                >
                  <rect x="2.5" y="2" width="11" height="12" rx="1.5" />
                  <path d="M2.5 8h11" />
                </svg>
              </button>
            </div>

            <button
              class="tab-item-close-btn"
              :title="t('terminal.kill_terminal') || '关闭终端'"
              @click.stop="closeSession(s.id)"
            >
              <svg
                viewBox="0 0 16 16"
                width="13"
                height="13"
                fill="none"
                stroke="currentColor"
                stroke-width="1.8"
                stroke-linecap="round"
                stroke-linejoin="round"
              >
                <path
                  d="M3 4.5h10M6 4.5V3a1 1 0 0 1 1-1h2a1 1 0 0 1 1 1v1.5M12 4.5v8.5a1.5 1.5 0 0 1-1.5 1.5h-5A1.5 1.5 0 0 1 4 13V4.5"
                />
                <path d="M6.5 7.5v4M9.5 7.5v4" />
              </svg>
            </button>
          </div>
        </div>
      </div>

      <!-- Dragging overlay for terminal sidebar -->
      <div v-if="isTermSidebarDragging" class="term-sash-drag-overlay"></div>
    </div>

    <!-- 2. Logs Drawer Body (Output, Tasks, Doctor) -->
    <div
      v-if="currentTab === 'output' || currentTab === 'tasks' || currentTab === 'doctor'"
      class="drawer-body font-mono"
      ref="logContainerRef"
      :style="{ height: `${effectiveBodyHeight}px` }"
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
          <svg
            v-if="copiedRowIdx !== idx"
            viewBox="0 0 16 16"
            width="11"
            height="11"
            fill="none"
            stroke="currentColor"
            stroke-width="1.8"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <rect x="5.5" y="5.5" width="8" height="8" rx="1.5" />
            <path d="M10.5 5.5V3.5a1.5 1.5 0 0 0-1.5-1.5h-5A1.5 1.5 0 0 0 2.5 3.5v5a1.5 1.5 0 0 0 1.5 1.5h1.5" />
          </svg>
          <svg
            v-else
            viewBox="0 0 16 16"
            width="11"
            height="11"
            fill="none"
            stroke="currentColor"
            stroke-width="2.2"
            stroke-linecap="round"
            stroke-linejoin="round"
            style="color: var(--success, #10b981)"
          >
            <path d="M3.5 8.5l3 3 6-6" />
          </svg>
        </button>
      </div>
    </div>

    <!-- 3. Ports View Body -->
    <div
      v-else-if="currentTab === 'ports'"
      class="ports-drawer-wrapper"
      :style="{ height: `${effectiveBodyHeight}px` }"
    >
      <PortsPanelView />
    </div>

    <!-- Dragging overlay for drawer height -->
    <div v-if="isDrawerHeightDragging" class="drawer-height-drag-overlay"></div>

    <!-- Tab Context Menu (VS Code style right-click menu with visibility checks) -->
    <Teleport to="body">
      <div
        v-if="isTabContextMenuOpen"
        class="tab-context-menu"
        :style="{ top: `${tabContextMenuY}px`, left: `${tabContextMenuX}px` }"
        @click.stop
      >
        <div class="context-menu-header">{{ t("drawer.panel_views") || "面板视图" }}</div>
        <div
          v-for="tabId in DEFAULT_TAB_ORDER"
          :key="tabId"
          class="tab-context-menu-item"
          :class="{ disabled: !hiddenTabs.has(tabId) && visibleTabs.length <= 1 }"
          @click="toggleTabVisibility(tabId)"
        >
          <span class="context-check-icon">
            <svg
              v-if="!hiddenTabs.has(tabId)"
              viewBox="0 0 16 16"
              width="13"
              height="13"
              fill="none"
              stroke="currentColor"
              stroke-width="2.2"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <polyline points="3 8.5 6.5 12 13 4" />
            </svg>
          </span>
          <TechIcon :id="tabDefinitions[tabId].icon" :size="13" />
          <span class="context-item-label">{{ tabDefinitions[tabId].label }}</span>
        </div>

        <div class="context-menu-divider"></div>

        <div class="tab-context-menu-item" @click="handleResetTabOrder">
          <span class="context-check-icon"></span>
          <svg
            viewBox="0 0 16 16"
            width="13"
            height="13"
            fill="none"
            stroke="currentColor"
            stroke-width="1.8"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <path d="M2.5 8a5.5 5.5 0 1 0 1.5-3.8L2 6" />
            <path d="M2 2.5v3.5h3.5" />
          </svg>
          <span>{{ t("drawer.reset_tab_order") || "重置标签页顺序与可见性" }}</span>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick, onMounted, onUnmounted } from "vue";
import { t } from "../../i18n";
import TechIcon from "../common/TechIcon.vue";
import { useToast } from "../../composables/useToast";
import { useTerminal, TerminalItem } from "../../composables/useTerminal";
import { useTerminalSessions } from "../../composables/useTerminalSessions";
import XTerminalPane from "../terminal/XTerminalPane.vue";
import PortsPanelView from "../ports/PortsPanelView.vue";
import { usePorts } from "../../composables/usePorts";
import { useTerminalProcessWatcher } from "../../composables/useTerminalProcessWatcher";

export interface LogItem {
  timestamp: string;
  level: "INFO" | "SUCCESS" | "WARN" | "ERROR";
  message: string;
  details?: string;
  category?: "output" | "tasks" | "doctor";
}

export type DrawerTab = "terminal" | "output" | "tasks" | "doctor" | "ports";

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
const { terminals, currentTerminal, fetchTerminals } = useTerminal();

const { activePortsCount } = usePorts();

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

const { getActiveProcess, getEffectiveTitle, setCustomTitle } = useTerminalProcessWatcher();

// 双击就地重命名终端 Tab
const editingSessionId = ref<string | null>(null);
const editingTitleInput = ref("");

function startRenameSession(sessionId: string, currentTitle: string) {
  editingSessionId.value = sessionId;
  editingTitleInput.value = currentTitle;
  nextTick(() => {
    const inputEl = document.querySelector(`.tab-rename-input-${sessionId}`) as HTMLInputElement;
    if (inputEl) {
      inputEl.focus();
      inputEl.select();
    }
  });
}

function finishRenameSession(sessionId: string) {
  if (editingSessionId.value === sessionId) {
    if (editingTitleInput.value.trim()) {
      setCustomTitle(sessionId, editingTitleInput.value.trim());
    }
    editingSessionId.value = null;
  }
}

function cancelRenameSession() {
  editingSessionId.value = null;
}

const isNewMenuOpen = ref(false);
const isMoreMenuOpen = ref(false);

const DRAWER_TAB_KEY = "unirtm_drawer_tab";
const DRAWER_MAX_KEY = "unirtm_drawer_maximized";
const DRAWER_AUTOSCROLL_KEY = "unirtm_drawer_autoscroll";

const TERM_SIDEBAR_WIDTH_KEY = "unirtm_term_sidebar_width";
const DEFAULT_TERM_SIDEBAR_WIDTH = 215;
const MIN_TERM_SIDEBAR_WIDTH = 130;
const MAX_TERM_SIDEBAR_WIDTH = 500;

function getInitialTermSidebarWidth(): number {
  if (typeof localStorage === "undefined") return DEFAULT_TERM_SIDEBAR_WIDTH;
  try {
    const saved = localStorage.getItem(TERM_SIDEBAR_WIDTH_KEY);
    if (saved) {
      const parsed = Number(saved);
      if (!isNaN(parsed) && parsed >= MIN_TERM_SIDEBAR_WIDTH && parsed <= MAX_TERM_SIDEBAR_WIDTH) {
        return parsed;
      }
    }
  } catch {}
  return DEFAULT_TERM_SIDEBAR_WIDTH;
}

const termSidebarWidth = ref(getInitialTermSidebarWidth());
const isTermSidebarDragging = ref(false);

function resetTermSidebarWidth() {
  termSidebarWidth.value = DEFAULT_TERM_SIDEBAR_WIDTH;
  try {
    localStorage.setItem(TERM_SIDEBAR_WIDTH_KEY, String(DEFAULT_TERM_SIDEBAR_WIDTH));
  } catch {}
}

let termDragStartX = 0;
let termDragStartWidth = 0;

function handleTermSidebarPointerDown(e: PointerEvent) {
  e.preventDefault();
  isTermSidebarDragging.value = true;
  termDragStartX = e.clientX;
  termDragStartWidth = termSidebarWidth.value;

  document.body.style.cursor = "col-resize";
  document.body.style.userSelect = "none";

  const onPointerMove = (ev: PointerEvent) => {
    if (!isTermSidebarDragging.value) return;
    const delta = termDragStartX - ev.clientX; // dragging left expands right sidebar
    const maxAllowed = Math.min(MAX_TERM_SIDEBAR_WIDTH, Math.floor(window.innerWidth * 0.5));
    const newWidth = Math.max(MIN_TERM_SIDEBAR_WIDTH, Math.min(maxAllowed, termDragStartWidth + delta));
    termSidebarWidth.value = Math.round(newWidth);
  };

  const onPointerUp = () => {
    if (isTermSidebarDragging.value) {
      isTermSidebarDragging.value = false;
      document.body.style.cursor = "";
      document.body.style.userSelect = "";
      try {
        localStorage.setItem(TERM_SIDEBAR_WIDTH_KEY, String(termSidebarWidth.value));
      } catch {}
    }
    window.removeEventListener("pointermove", onPointerMove);
    window.removeEventListener("pointerup", onPointerUp);
    window.removeEventListener("pointercancel", onPointerUp);
  };

  window.addEventListener("pointermove", onPointerMove);
  window.addEventListener("pointerup", onPointerUp);
  window.addEventListener("pointercancel", onPointerUp);
}

const DRAWER_HEIGHT_KEY = "unirtm_drawer_height";
const DEFAULT_DRAWER_BODY_HEIGHT = 200;
const MIN_DRAWER_BODY_HEIGHT = 100;
const MAX_DRAWER_BODY_HEIGHT = 800;

function getInitialDrawerBodyHeight(): number {
  if (typeof localStorage === "undefined") return DEFAULT_DRAWER_BODY_HEIGHT;
  try {
    const saved = localStorage.getItem(DRAWER_HEIGHT_KEY);
    if (saved) {
      const parsed = Number(saved);
      if (!isNaN(parsed) && parsed >= MIN_DRAWER_BODY_HEIGHT && parsed <= MAX_DRAWER_BODY_HEIGHT) {
        return parsed;
      }
    }
  } catch {}
  return DEFAULT_DRAWER_BODY_HEIGHT;
}

const drawerBodyHeight = ref(getInitialDrawerBodyHeight());
const isDrawerHeightDragging = ref(false);

const effectiveBodyHeight = computed(() => {
  if (isMaximized.value) {
    if (typeof window !== "undefined") {
      return Math.max(380, Math.floor(window.innerHeight * 0.65));
    }
    return 420;
  }
  return drawerBodyHeight.value;
});

function handleToggleMaximizeOrReset() {
  if (isMaximized.value) {
    isMaximized.value = false;
  } else {
    if (drawerBodyHeight.value !== DEFAULT_DRAWER_BODY_HEIGHT) {
      drawerBodyHeight.value = DEFAULT_DRAWER_BODY_HEIGHT;
      try {
        localStorage.setItem(DRAWER_HEIGHT_KEY, String(DEFAULT_DRAWER_BODY_HEIGHT));
      } catch {}
    } else {
      isMaximized.value = true;
    }
  }
}

let heightDragStartY = 0;
let heightDragStartHeight = 0;

function handleDrawerHeightPointerDown(e: PointerEvent) {
  e.preventDefault();
  isDrawerHeightDragging.value = true;
  heightDragStartY = e.clientY;

  if (isMaximized.value) {
    heightDragStartHeight = effectiveBodyHeight.value;
    isMaximized.value = false;
  } else {
    heightDragStartHeight = drawerBodyHeight.value;
  }

  document.body.style.cursor = "row-resize";
  document.body.style.userSelect = "none";

  const onPointerMove = (ev: PointerEvent) => {
    if (!isDrawerHeightDragging.value) return;
    const deltaY = heightDragStartY - ev.clientY; // dragging up increases height
    const maxAllowed = Math.min(MAX_DRAWER_BODY_HEIGHT, Math.floor(window.innerHeight * 0.8));
    const newHeight = Math.max(MIN_DRAWER_BODY_HEIGHT, Math.min(maxAllowed, heightDragStartHeight + deltaY));
    drawerBodyHeight.value = Math.round(newHeight);
  };

  const onPointerUp = () => {
    if (isDrawerHeightDragging.value) {
      isDrawerHeightDragging.value = false;
      document.body.style.cursor = "";
      document.body.style.userSelect = "";
      try {
        localStorage.setItem(DRAWER_HEIGHT_KEY, String(drawerBodyHeight.value));
      } catch {}
    }
    window.removeEventListener("pointermove", onPointerMove);
    window.removeEventListener("pointerup", onPointerUp);
    window.removeEventListener("pointercancel", onPointerUp);
  };

  window.addEventListener("pointermove", onPointerMove);
  window.addEventListener("pointerup", onPointerUp);
  window.addEventListener("pointercancel", onPointerUp);
}

function getInitialTab(): DrawerTab {
  if (typeof localStorage === "undefined") return "terminal";
  const saved = localStorage.getItem(DRAWER_TAB_KEY);
  if (saved === "terminal" || saved === "output" || saved === "tasks" || saved === "doctor" || saved === "ports") {
    return saved as DrawerTab;
  }
  return "terminal";
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
  if (hiddenTabs.value.has(val)) {
    const newHidden = new Set(hiddenTabs.value);
    newHidden.delete(val);
    hiddenTabs.value = newHidden;
    try {
      localStorage.setItem(DRAWER_TAB_HIDDEN_KEY, JSON.stringify([...newHidden]));
    } catch {}
  }
  if (typeof localStorage !== "undefined") {
    localStorage.setItem(DRAWER_TAB_KEY, val);
  }
  isNewMenuOpen.value = false;
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

const outputLogs = computed(() => props.logs.filter((l) => !l.category || l.category === "output"));
const tasksLogs = computed(() => props.logs.filter((l) => l.category === "tasks"));
const doctorLogs = computed(() => props.logs.filter((l) => l.category === "doctor"));

const DRAWER_TAB_ORDER_KEY = "unirtm_drawer_tab_order";
const DRAWER_TAB_HIDDEN_KEY = "unirtm_drawer_tab_hidden";
const DEFAULT_TAB_ORDER: DrawerTab[] = ["terminal", "output", "tasks", "doctor", "ports"];

function getInitialTabOrder(): DrawerTab[] {
  if (typeof localStorage === "undefined") return [...DEFAULT_TAB_ORDER];
  try {
    const saved = localStorage.getItem(DRAWER_TAB_ORDER_KEY);
    if (saved) {
      const parsed = JSON.parse(saved);
      if (Array.isArray(parsed)) {
        const valid = parsed.filter((id): id is DrawerTab => DEFAULT_TAB_ORDER.includes(id as DrawerTab));
        const missing = DEFAULT_TAB_ORDER.filter((id) => !valid.includes(id));
        return [...valid, ...missing];
      }
    }
  } catch {}
  return [...DEFAULT_TAB_ORDER];
}

function getInitialHiddenTabs(): Set<DrawerTab> {
  if (typeof localStorage === "undefined") return new Set();
  try {
    const saved = localStorage.getItem(DRAWER_TAB_HIDDEN_KEY);
    if (saved) {
      const parsed = JSON.parse(saved);
      if (Array.isArray(parsed)) {
        return new Set(parsed.filter((id): id is DrawerTab => DEFAULT_TAB_ORDER.includes(id as DrawerTab)));
      }
    }
  } catch {}
  return new Set();
}

const tabOrder = ref<DrawerTab[]>(getInitialTabOrder());
const hiddenTabs = ref<Set<DrawerTab>>(getInitialHiddenTabs());

const visibleTabs = computed<DrawerTab[]>(() => {
  const visible = tabOrder.value.filter((id) => !hiddenTabs.value.has(id));
  return visible.length > 0 ? visible : ["terminal"];
});

interface TabDefinition {
  id: DrawerTab;
  label: string;
  icon: string;
  badge?: number;
}

const tabDefinitions = computed<Record<DrawerTab, TabDefinition>>(() => ({
  terminal: {
    id: "terminal",
    label: t("drawer.tab_terminal") || "终端",
    icon: "workbench:terminal",
  },
  output: {
    id: "output",
    label: t("drawer.tab_output") || "输出",
    icon: "workbench:logs",
    badge: outputLogs.value.length,
  },
  tasks: {
    id: "tasks",
    label: t("drawer.tab_tasks") || "任务",
    icon: "workbench:tasks",
    badge: tasksLogs.value.length,
  },
  doctor: {
    id: "doctor",
    label: t("drawer.tab_doctor") || "诊断",
    icon: "workbench:doctor",
    badge: doctorLogs.value.length,
  },
  ports: {
    id: "ports",
    label: t("drawer.tab_ports") || "端口",
    icon: "workbench:ports",
    badge: activePortsCount.value,
  },
}));

// Tab 拖拽排序逻辑 (VS Code style Drag and Drop)
const draggedTabId = ref<DrawerTab | null>(null);
const dropTargetTabId = ref<DrawerTab | null>(null);
const dropTargetPosition = ref<"before" | "after" | null>(null);
let justDragged = false;

function handleTabDragStart(e: DragEvent, tabId: DrawerTab) {
  draggedTabId.value = tabId;
  if (e.dataTransfer) {
    e.dataTransfer.effectAllowed = "move";
    e.dataTransfer.setData("text/plain", tabId);
  }
}

function handleTabDragOver(e: DragEvent, targetTabId: DrawerTab) {
  if (!draggedTabId.value || draggedTabId.value === targetTabId) {
    dropTargetTabId.value = null;
    dropTargetPosition.value = null;
    return;
  }
  e.preventDefault();
  if (e.dataTransfer) {
    e.dataTransfer.dropEffect = "move";
  }

  const el = e.currentTarget as HTMLElement;
  const rect = el.getBoundingClientRect();
  const isLeft = e.clientX - rect.left < rect.width / 2;

  dropTargetTabId.value = targetTabId;
  dropTargetPosition.value = isLeft ? "before" : "after";
}

function handleTabDragLeave(e: DragEvent, tabId: DrawerTab) {
  const el = e.currentTarget as HTMLElement;
  if (e.relatedTarget && el.contains(e.relatedTarget as Node)) {
    return;
  }
  if (dropTargetTabId.value === tabId) {
    dropTargetTabId.value = null;
    dropTargetPosition.value = null;
  }
}

function handleTabDrop(e: DragEvent, targetTabId: DrawerTab) {
  e.preventDefault();
  const sourceId = draggedTabId.value;
  if (!sourceId || sourceId === targetTabId) {
    clearDragState();
    return;
  }

  const currentOrder = [...tabOrder.value];
  const oldIndex = currentOrder.indexOf(sourceId);
  if (oldIndex === -1) {
    clearDragState();
    return;
  }
  currentOrder.splice(oldIndex, 1);

  const targetIndex = currentOrder.indexOf(targetTabId);
  if (targetIndex !== -1) {
    const insertIndex = dropTargetPosition.value === "after" ? targetIndex + 1 : targetIndex;
    currentOrder.splice(insertIndex, 0, sourceId);
    tabOrder.value = currentOrder;
    try {
      localStorage.setItem(DRAWER_TAB_ORDER_KEY, JSON.stringify(currentOrder));
    } catch {}
  }

  justDragged = true;
  setTimeout(() => {
    justDragged = false;
  }, 100);

  clearDragState();
}

function handleTabDragEnd() {
  justDragged = true;
  setTimeout(() => {
    justDragged = false;
  }, 100);
  clearDragState();
}

function clearDragState() {
  draggedTabId.value = null;
  dropTargetTabId.value = null;
  dropTargetPosition.value = null;
}

function handleTabClick(tabId: DrawerTab) {
  if (justDragged) return;
  currentTab.value = tabId;
}

// Tab 右键菜单 (VS Code style right-click to reset)
const isTabContextMenuOpen = ref(false);
const tabContextMenuX = ref(0);
const tabContextMenuY = ref(0);

function handleTabContextMenu(e: MouseEvent) {
  e.preventDefault();
  tabContextMenuX.value = Math.min(e.clientX, window.innerWidth - 180);
  tabContextMenuY.value = Math.min(e.clientY, window.innerHeight - 100);
  isTabContextMenuOpen.value = true;
}

function toggleTabVisibility(tabId: DrawerTab) {
  const newHidden = new Set(hiddenTabs.value);
  if (newHidden.has(tabId)) {
    newHidden.delete(tabId);
  } else {
    // 保护：至少保留 1 个可见 Tab
    if (visibleTabs.value.length <= 1) {
      return;
    }
    newHidden.add(tabId);

    // 如果被隐藏的 Tab 正好是当前激活的 Tab，则自动切换到剩余可见的第一个 Tab
    if (currentTab.value === tabId) {
      const remaining = tabOrder.value.filter((id) => !newHidden.has(id));
      if (remaining.length > 0) {
        currentTab.value = remaining[0];
      }
    }
  }
  hiddenTabs.value = newHidden;
  try {
    localStorage.setItem(DRAWER_TAB_HIDDEN_KEY, JSON.stringify([...newHidden]));
  } catch {}
}

function handleResetTabOrder() {
  tabOrder.value = [...DEFAULT_TAB_ORDER];
  hiddenTabs.value = new Set();
  try {
    localStorage.removeItem(DRAWER_TAB_ORDER_KEY);
    localStorage.removeItem(DRAWER_TAB_HIDDEN_KEY);
  } catch {}
  isTabContextMenuOpen.value = false;
}

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
    return t("drawer.empty_tasks") || "暂无任务执行日志。在任务视图中点击“运行”以在此查看执行终端输出。";
  }
  if (currentTab.value === "doctor") {
    return t("drawer.empty_doctor") || "暂无环境体检报告。在环境诊断视图中执行“重新全面体检”以查看详情。";
  }
  return t("drawer.empty_logs") || "暂无系统输出日志。系统运行与扩展事件将在此显示。";
});

const autoScrollTitle = computed(() => (autoScroll.value ? "自动滚动 (开启)" : "自动滚动 (关闭)"));
const maximizeTitle = computed(() => (isMaximized.value ? "还原面板高度" : "最大化面板高度"));
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

function handleCreateDefaultSession() {
  const s = createSession(
    currentTerminal.value ? { name: currentTerminal.value.name, path: currentTerminal.value.path } : undefined
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

function handleMenuCopyLogs() {
  isMoreMenuOpen.value = false;
  handleCopyAll();
}

function handleMenuClearLogs() {
  isMoreMenuOpen.value = false;
  handleClearCurrent();
}

function handleMenuKillTerminal() {
  isMoreMenuOpen.value = false;
  handleKillCurrentTerminal();
}

function handleGlobalClick() {
  isNewMenuOpen.value = false;
  isMoreMenuOpen.value = false;
  isTabContextMenuOpen.value = false;
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
  position: relative;
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

.bottom-drawer.is-resizing {
  user-select: none;
}

.bottom-drawer.is-resizing .terminal-panes-area,
.bottom-drawer.is-resizing .terminal-screen {
  pointer-events: none;
}

.drawer-top-resizer-sash {
  position: absolute;
  top: -3px;
  left: 0;
  right: 0;
  height: 6px;
  cursor: row-resize;
  z-index: 70;
  display: flex;
  align-items: center;
  justify-content: center;
  touch-action: none;
}

.drawer-top-resizer-sash .sash-indicator-horizontal {
  width: 100%;
  height: 2px;
  background: transparent;
  transition: background 0.15s ease;
}

.drawer-top-resizer-sash:hover .sash-indicator-horizontal,
.drawer-top-resizer-sash.active .sash-indicator-horizontal {
  background: var(--accent-cyan, #00e5ff);
}

.drawer-height-drag-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 99999;
  cursor: row-resize;
  background: transparent;
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
  position: relative;
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
  cursor: grab;
  user-select: none;
  -webkit-user-select: none;
  transition:
    color 0.15s ease,
    background 0.15s ease,
    opacity 0.15s ease;
}

.drawer-tab:active {
  cursor: grabbing;
}

.drawer-tab:hover {
  color: var(--text-main);
}

.drawer-tab.active {
  color: var(--accent-cyan);
  border-bottom-color: var(--accent-cyan);
  font-weight: 600;
}

.drawer-tab.is-dragging {
  opacity: 0.35;
  background: rgba(255, 255, 255, 0.04);
}

/* VS Code 风格的插入垂直指示线 */
.drawer-tab.drop-indicator-left::before,
.drawer-tab.drop-indicator-right::after {
  content: "";
  position: absolute;
  top: 5px;
  bottom: 5px;
  width: 2px;
  background: var(--accent-cyan, #00e5ff);
  box-shadow: 0 0 8px var(--accent-cyan, #00e5ff);
  border-radius: 1px;
  z-index: 20;
  pointer-events: none;
}

.drawer-tab.drop-indicator-left::before {
  left: -2px;
}

.drawer-tab.drop-indicator-right::after {
  right: -2px;
}

.tab-label-inner {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  pointer-events: none;
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
  pointer-events: none;
}

.drawer-tab.active .tab-badge {
  background: rgba(34, 211, 238, 0.15);
  color: var(--accent-cyan);
}

/* Tab 右键上下文菜单 (VS Code style) */
.tab-context-menu {
  position: fixed;
  z-index: 100000;
  background: var(--card-bg, #111827);
  border: 1px solid var(--card-border, #1f2937);
  box-shadow:
    0 10px 25px -5px rgba(0, 0, 0, 0.6),
    0 8px 10px -6px rgba(0, 0, 0, 0.4);
  border-radius: 6px;
  padding: 4px;
  min-width: 155px;
  font-size: 0.75rem;
  color: var(--text-main, #f3f4f6);
  user-select: none;
  animation: tab-context-menu-fade 0.12s ease-out;
}

@keyframes tab-context-menu-fade {
  from {
    opacity: 0;
    transform: scale(0.96);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}

.context-menu-header {
  padding: 4px 10px 6px 10px;
  font-size: 0.68rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: var(--text-muted, #9ca3af);
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
  margin-bottom: 3px;
}

.context-menu-divider {
  height: 1px;
  background: var(--card-border, #1f2937);
  margin: 4px 0;
}

.context-check-icon {
  width: 14px;
  height: 14px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: var(--accent-cyan, #00e5ff);
  flex-shrink: 0;
}

.context-item-label {
  flex: 1;
}

.tab-context-menu-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 10px;
  border-radius: 4px;
  cursor: pointer;
  transition:
    background 0.12s ease,
    color 0.12s ease;
}

.tab-context-menu-item:hover:not(.disabled) {
  background: rgba(0, 229, 255, 0.12);
  color: var(--accent-cyan, #00e5ff);
}

.tab-context-menu-item.disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

/* Shell options for New Terminal dropdown */

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
  transition:
    opacity 0.15s ease,
    transform 0.15s ease;
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
  bottom: calc(100% + 6px);
  top: auto;
  right: 0;
  width: 240px;
  max-height: 260px;
  overflow-y: auto;
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  box-shadow: 0 -8px 24px rgba(0, 0, 0, 0.45);
  border-radius: 6px;
  padding: 4px;
  z-index: 120;
}

.dropdown-header {
  padding: 5px 8px;
  font-size: 0.68rem;
  font-weight: 700;
  color: var(--text-muted);
  border-bottom: 1px solid var(--card-border);
  margin-bottom: 4px;
}

.drawer-more-wrapper {
  position: relative;
  display: inline-flex;
  align-items: center;
}

.drawer-more-dropdown {
  position: absolute;
  bottom: calc(100% + 6px);
  top: auto;
  right: 0;
  min-width: 130px;
  background: var(--card-bg, #1e293b);
  border: 1px solid var(--card-border, #334155);
  box-shadow: 0 -8px 24px rgba(0, 0, 0, 0.45);
  border-radius: 6px;
  padding: 4px;
  z-index: 120;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.more-menu-item {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  padding: 6px 10px;
  font-size: 0.74rem;
  background: transparent;
  border: none;
  border-radius: 4px;
  color: var(--text-main, #f1f5f9);
  cursor: pointer;
  text-align: left;
  transition: all 0.15s ease;
  white-space: nowrap;
}

.more-menu-item:hover {
  background: rgba(255, 255, 255, 0.08);
  color: var(--accent-cyan, #38bdf8);
}

.more-menu-item.danger:hover {
  background: rgba(239, 68, 68, 0.15);
  color: #ef4444;
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
  position: relative;
}

.terminal-wrapper.is-resizing {
  user-select: none;
}

.terminal-wrapper.is-resizing .terminal-panes-area {
  pointer-events: none;
}

/* Draggable resizer sash between terminal panes and terminal sidebar */
.terminal-sidebar-sash {
  position: relative;
  width: 0;
  height: 100%;
  z-index: 15;
  flex-shrink: 0;
  user-select: none;
  cursor: col-resize;
  touch-action: none;
}

.terminal-sidebar-sash .sash-indicator {
  position: absolute;
  top: 0;
  bottom: 0;
  left: -2px;
  width: 5px;
  background: transparent;
  transition: background 0.15s ease;
}

.terminal-sidebar-sash:hover .sash-indicator,
.terminal-sidebar-sash.active .sash-indicator {
  background: var(--accent-cyan, #00e5ff);
}

.term-sash-drag-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 99999;
  cursor: col-resize;
  background: transparent;
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
  outline: none;
}

.terminal-panes-area:not(.split-none) .terminal-pane.active {
  outline: 1.5px solid rgba(0, 229, 255, 0.6);
  outline-offset: -1px;
  z-index: 2;
}

.terminal-pane-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.45rem;
  padding: 4px 0.75rem;
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
  padding: 3px;
  border-radius: 4px;
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
  min-width: 0;
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
  padding: 5px 8px 5px 10px;
  font-size: 0.68rem;
  font-weight: 700;
  color: var(--text-muted);
  border-bottom: 1px solid var(--card-border);
  background: rgba(0, 0, 0, 0.08);
}

.sidebar-add-btn {
  background: transparent;
  border: 1px solid transparent;
  color: var(--text-muted);
  cursor: pointer;
  width: 22px;
  height: 22px;
  margin-right: 6px;
  border-radius: 4px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s ease;
}

.sidebar-add-btn:hover {
  color: var(--accent-cyan);
  background: var(--card-bg);
  border-color: var(--card-border);
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
  gap: 8px;
  padding: 5px 8px 5px 10px;
  border-radius: 5px;
  cursor: pointer;
  transition: all 0.15s ease;
  position: relative;
  margin: 1px 2px;
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

.tab-rename-input {
  width: 100%;
  font-size: 0.72rem;
  font-family: inherit;
  background: var(--bg-surface-elevated, #1e1e1e);
  color: var(--text-main, #ffffff);
  border: 1px solid var(--accent-cyan, #00e5ff);
  border-radius: 3px;
  padding: 1px 4px;
  outline: none;
}

.tab-process-indicator {
  font-size: 0.65rem;
  line-height: 1;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.term-process-badge {
  font-size: 0.62rem;
  padding: 1px 5px;
  border-radius: 3px;
  border: 1px solid;
  font-weight: 600;
  letter-spacing: 0.02em;
  text-transform: uppercase;
}

.tab-item-split-badge {
  font-size: 0.62rem;
  opacity: 0.75;
  color: var(--accent-cyan);
}

.tab-item-close-btn {
  opacity: 0;
  pointer-events: none;
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  width: 22px;
  height: 22px;
  margin-right: 4px;
  border-radius: 4px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition:
    opacity 0.2s cubic-bezier(0.4, 0, 0.2, 1),
    background 0.2s cubic-bezier(0.4, 0, 0.2, 1),
    color 0.2s cubic-bezier(0.4, 0, 0.2, 1),
    transform 0.15s ease;
  flex-shrink: 0;
}

.terminal-tab-item:hover .tab-item-close-btn {
  opacity: 0.85;
  pointer-events: auto;
}

.tab-item-close-btn:hover {
  opacity: 1 !important;
  color: #ef4444 !important;
  background: rgba(239, 68, 68, 0.18) !important;
  transform: scale(1.08);
}

.tab-item-split-actions {
  display: flex;
  align-items: center;
  gap: 2px;
  margin-right: 2px;
  flex-shrink: 0;
}

.tab-action-btn {
  opacity: 0;
  pointer-events: none;
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  width: 22px;
  height: 22px;
  border-radius: 4px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition:
    opacity 0.2s cubic-bezier(0.4, 0, 0.2, 1),
    background 0.2s cubic-bezier(0.4, 0, 0.2, 1),
    color 0.2s cubic-bezier(0.4, 0, 0.2, 1),
    transform 0.15s ease;
  flex-shrink: 0;
}

.terminal-tab-item:hover .tab-action-btn {
  opacity: 0.85;
  pointer-events: auto;
}

.tab-action-btn:hover {
  opacity: 1 !important;
  color: var(--accent-cyan) !important;
  background: rgba(0, 229, 255, 0.15) !important;
  transform: scale(1.08);
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
  position: relative;
  overflow: hidden;
  padding: 0;
  display: flex;
  flex-direction: column;
  background: transparent;
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

/* Light Theme Polish for Integrated Terminal (VSCode Light Standards) */
[data-theme="light"] .terminal-wrapper {
  background: #ffffff;
  color: #1e293b;
}

[data-theme="light"] .terminal-pane {
  background: #ffffff;
}

[data-theme="light"] .terminal-pane.active {
  outline: none;
}

[data-theme="light"] .terminal-panes-area:not(.split-none) .terminal-pane.active {
  outline: 1.5px solid #0284c7;
}

[data-theme="light"] .terminal-pane-banner {
  background: #f1f5f9;
  border-color: #e2e8f0;
  color: #475569;
}

[data-theme="light"] .term-shell-path {
  color: #64748b;
}

[data-theme="light"] .term-split-pill {
  background: rgba(2, 132, 199, 0.1);
  color: #0284c7;
}

[data-theme="light"] .terminal-screen {
  background: #ffffff;
}

[data-theme="light"] .term-text-cyan {
  color: #0284c7;
}

[data-theme="light"] .term-text-muted {
  color: #64748b;
}

[data-theme="light"] .term-prompt {
  color: #0284c7;
}

[data-theme="light"] .line-input {
  color: #0f172a;
}

[data-theme="light"] .line-output {
  color: #334155;
}

[data-theme="light"] .line-info {
  color: #0284c7;
}

[data-theme="light"] .terminal-cmd-input {
  color: #0f172a;
  caret-color: #0284c7;
}

[data-theme="light"] .terminal-cmd-input::placeholder {
  color: #94a3b8;
}

[data-theme="light"] .terminal-sidebar {
  background: #f8fafc;
  border-left-color: #e2e8f0;
}

[data-theme="light"] .terminal-sidebar-header {
  background: #f1f5f9;
  border-bottom-color: #e2e8f0;
  color: #475569;
}

[data-theme="light"] .sidebar-add-btn:hover {
  background: #ffffff;
  border-color: #cbd5e1;
  color: #0284c7;
}

[data-theme="light"] .terminal-tab-item {
  color: #334155;
}

[data-theme="light"] .terminal-tab-item:hover {
  background: rgba(0, 0, 0, 0.04);
}

[data-theme="light"] .terminal-tab-item.active {
  background: rgba(2, 132, 199, 0.12);
  color: #0284c7;
}

[data-theme="light"] .terminal-tab-item.active .tab-item-icon {
  color: #0284c7;
}

[data-theme="light"] .term-new-dropdown {
  background: #ffffff;
  border-color: #e2e8f0;
  box-shadow: 0 -8px 24px rgba(0, 0, 0, 0.15);
}

[data-theme="light"] .dropdown-header {
  background: #f8fafc;
  border-bottom-color: #e2e8f0;
  color: #475569;
}

[data-theme="light"] .shell-option:hover {
  background: #f1f5f9;
}

[data-theme="light"] .shell-option.active {
  background: rgba(2, 132, 199, 0.1);
}

[data-theme="light"] .shell-option-name {
  color: #1e293b;
}

[data-theme="light"] .shell-option-path {
  color: #64748b;
}

[data-theme="light"] .shell-option.active .shell-option-name {
  color: #0284c7;
}

[data-theme="light"] .drawer-more-dropdown {
  background: #ffffff;
  border-color: #cbd5e1;
  box-shadow: 0 -4px 16px rgba(0, 0, 0, 0.1);
}

[data-theme="light"] .more-menu-item {
  color: #334155;
}

[data-theme="light"] .more-menu-item:hover {
  background: #f1f5f9;
  color: #0284c7;
}

[data-theme="light"] .more-menu-item.danger:hover {
  background: #fee2e2;
  color: #dc2626;
}

[data-theme="light"] .tab-action-btn:hover {
  color: #0284c7 !important;
  background: rgba(2, 132, 199, 0.12) !important;
}

/* Ports View Wrapper */
.ports-drawer-wrapper {
  display: flex;
  flex-direction: column;
  height: 100%;
  width: 100%;
  overflow: hidden;
  background: var(--card-bg, #070a12);
}

[data-theme="light"] .ports-drawer-wrapper {
  background: #ffffff;
}
</style>
