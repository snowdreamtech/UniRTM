<template>
  <aside
    class="activity-bar"
    id="app-activity-bar"
    role="navigation"
    aria-label="Activity Bar"
    @contextmenu.prevent="handleContextMenu($event)"
  >
    <!-- Top Action Icons -->
    <div class="activity-top">
      <button
        v-for="item in visibleTabs"
        :key="item.id"
        class="activity-btn"
        :id="'btn-activity-' + item.id"
        :class="{
          active: currentTab === item.id,
          'is-dragging': draggedTabId === item.id,
          'drop-indicator-top': dropTargetTabId === item.id && dropTargetPosition === 'before',
          'drop-indicator-bottom': dropTargetTabId === item.id && dropTargetPosition === 'after',
        }"
        draggable="true"
        :title="item.label + ' (' + (t('activity_bar.drag_reorder_tip') || '拖拽调整顺序，右键重置') + ')'"
        @dragstart="handleDragStart($event, item.id as ActivityTabId)"
        @dragover="handleDragOver($event, item.id as ActivityTabId)"
        @dragleave="handleDragLeave($event, item.id as ActivityTabId)"
        @drop="handleDrop($event, item.id as ActivityTabId)"
        @dragend="handleDragEnd"
        @click="handleTabClick(item.id)"
      >
        <span class="activity-indicator" v-if="currentTab === item.id"></span>
        <span class="activity-icon">
          <TechIcon :id="item.icon" :size="24" />
        </span>
        <span v-if="item.badge" class="activity-badge" :class="'badge-' + item.id">{{ item.badge }}</span>
      </button>
    </div>

    <!-- Bottom Utility Icons -->
    <div class="activity-bottom">
      <button
        class="activity-btn"
        id="btn-activity-settings"
        :class="{ active: currentTab === 'settings' }"
        :title="t('settings.title') || 'Settings'"
        @click="$emit('select-tab', 'settings')"
      >
        <span class="activity-indicator" v-if="currentTab === 'settings'"></span>
        <span class="activity-icon">
          <TechIcon id="workbench:settings" :size="22" />
        </span>
      </button>

      <button
        class="activity-btn"
        id="btn-activity-about"
        :class="{ active: currentTab === 'about' }"
        :title="t('about.title') || 'About'"
        @click="$emit('select-tab', 'about')"
      >
        <span class="activity-indicator" v-if="currentTab === 'about'"></span>
        <span class="activity-icon">
          <TechIcon id="workbench:about" :size="22" />
        </span>
      </button>

      <!-- Logs / Terminal Drawer Toggle Button -->
      <button
        class="activity-btn"
        id="btn-activity-logs"
        :class="{ active: isDrawerOpen }"
        :title="t('console.title') || 'Terminal / Logs'"
        @click="$emit('toggle-drawer')"
      >
        <span class="activity-indicator" v-if="isDrawerOpen"></span>
        <span class="activity-icon">
          <TechIcon id="workbench:terminal" :size="22" />
        </span>
      </button>
    </div>

    <!-- Activity Bar Context Menu (VS Code style right-click to reorder / toggle visibility) -->
    <Teleport to="body">
      <div
        v-if="isContextMenuOpen"
        class="activity-context-menu"
        :style="{ top: `${contextMenuY}px`, left: `${contextMenuX}px` }"
        @click.stop
      >
        <div class="context-menu-header">{{ t("activity_bar.views") || "活动栏视图" }}</div>
        <div
          v-for="tab in rawTabsList"
          :key="tab.id"
          class="activity-context-menu-item"
          :class="{ disabled: !hiddenTabs.has(tab.id as ActivityTabId) && visibleTabs.length <= 1 }"
          @click="toggleVisibility(tab.id as ActivityTabId)"
        >
          <span class="context-check-icon">
            <svg
              v-if="!hiddenTabs.has(tab.id as ActivityTabId)"
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
          <TechIcon :id="tab.icon" :size="14" />
          <span class="context-item-label">{{ tab.label }}</span>
        </div>

        <div class="context-menu-divider"></div>

        <div class="activity-context-menu-item" @click="handleResetOrder">
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
          <span>{{ t("activity_bar.reset_order") || "重置图标顺序与可见性" }}</span>
        </div>
      </div>
    </Teleport>
  </aside>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from "vue";
import { t } from "../../i18n";
import TechIcon from "../common/TechIcon.vue";
import { useToolUpdates } from "../../composables/useToolUpdates";
import { useGit } from "../../composables/useGit";
import { useGuard } from "../../composables/useGuard";
import {
  ActivityTabId,
  ACTIVITY_BAR_ORDER_KEY,
  ACTIVITY_BAR_HIDDEN_KEY,
  DEFAULT_ACTIVITY_TAB_ORDER,
  getInitialActivityTabOrder,
  getInitialHiddenActivityTabs,
  reorderActivityTabs,
  toggleActivityTabVisibility,
} from "./activityBarTabs";

const { updatableCount } = useToolUpdates();
const { totalChanges, fetchStatus } = useGit();
const { guardStatus } = useGuard();

export interface ActivityTab {
  id: string;
  icon: string;
  label: string;
  badge?: string | number;
}

const props = defineProps<{
  currentTab: string;
  isDrawerOpen?: boolean;
  toolCount?: number;
}>();

const emit = defineEmits<{
  (e: "select-tab", tabId: string): void;
  (e: "open-settings"): void;
  (e: "open-about"): void;
  (e: "toggle-drawer"): void;
}>();

const tabOrder = ref<ActivityTabId[]>(getInitialActivityTabOrder());
const hiddenTabs = ref<Set<ActivityTabId>>(getInitialHiddenActivityTabs());

const rawTabsList = computed<ActivityTab[]>(() => [
  {
    id: "dashboard",
    icon: "workbench:dashboard",
    label: t("nav.dashboard") || "Dashboard",
  },
  {
    id: "explorer",
    icon: "workbench:explorer",
    label: t("nav.explorer") || "Explorer",
  },
  {
    id: "scm",
    icon: "workbench:scm",
    label: t("nav.scm") || "源代码管理",
    badge: totalChanges.value > 0 ? totalChanges.value : undefined,
  },
  {
    id: "guard",
    icon: "workbench:shield",
    label: t("nav.guard") || "Quality Guard",
    badge:
      !guardStatus.value.clean && guardStatus.value.issues.length > 0 ? guardStatus.value.issues.length : undefined,
  },
  {
    id: "tools",
    icon: "workbench:tools",
    label: t("nav.tools") || "Toolchains",
    badge: (updatableCount.value > 0 ? updatableCount.value : undefined) ?? props.toolCount,
  },
  {
    id: "tasks",
    icon: "workbench:tasks",
    label: t("nav.tasks") || "Tasks",
  },
  {
    id: "environments",
    icon: "workbench:environments",
    label: t("nav.environments") || "Environments",
  },
  {
    id: "marketplace",
    icon: "workbench:marketplace",
    label: t("nav.marketplace") || "Marketplace",
  },
  {
    id: "doctor",
    icon: "workbench:doctor",
    label: t("nav.doctor") || "Doctor",
  },
]);

const tabMap = computed(() => {
  const map: Record<string, ActivityTab> = {};
  for (const item of rawTabsList.value) {
    map[item.id] = item;
  }
  return map;
});

const visibleTabs = computed<ActivityTab[]>(() => {
  const list: ActivityTab[] = [];
  for (const id of tabOrder.value) {
    if (!hiddenTabs.value.has(id) && tabMap.value[id]) {
      list.push(tabMap.value[id]);
    }
  }
  return list;
});

// Drag and drop state
const draggedTabId = ref<ActivityTabId | null>(null);
const dropTargetTabId = ref<ActivityTabId | null>(null);
const dropTargetPosition = ref<"before" | "after" | null>(null);
let justDragged = false;

function handleDragStart(e: DragEvent, tabId: ActivityTabId) {
  draggedTabId.value = tabId;
  if (e.dataTransfer) {
    e.dataTransfer.effectAllowed = "move";
    e.dataTransfer.setData("text/plain", tabId);
  }
}

function handleDragOver(e: DragEvent, targetId: ActivityTabId) {
  if (!draggedTabId.value || draggedTabId.value === targetId) {
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
  const isTop = e.clientY - rect.top < rect.height / 2;

  dropTargetTabId.value = targetId;
  dropTargetPosition.value = isTop ? "before" : "after";
}

function handleDragLeave(e: DragEvent, tabId: ActivityTabId) {
  const el = e.currentTarget as HTMLElement;
  if (e.relatedTarget && el.contains(e.relatedTarget as Node)) {
    return;
  }
  if (dropTargetTabId.value === tabId) {
    dropTargetTabId.value = null;
    dropTargetPosition.value = null;
  }
}

function handleDrop(e: DragEvent, targetId: ActivityTabId) {
  e.preventDefault();
  const sourceId = draggedTabId.value;
  if (!sourceId || sourceId === targetId || !dropTargetPosition.value) {
    clearDragState();
    return;
  }

  const newOrder = reorderActivityTabs(tabOrder.value, sourceId, targetId, dropTargetPosition.value);
  tabOrder.value = newOrder;

  try {
    localStorage.setItem(ACTIVITY_BAR_ORDER_KEY, JSON.stringify(newOrder));
  } catch {}

  justDragged = true;
  setTimeout(() => {
    justDragged = false;
  }, 100);

  clearDragState();
}

function handleDragEnd() {
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

function handleTabClick(id: string) {
  if (justDragged) return;
  emit("select-tab", id);
}

// Right-click context menu
const isContextMenuOpen = ref(false);
const contextMenuX = ref(0);
const contextMenuY = ref(0);

function handleContextMenu(e: MouseEvent) {
  e.preventDefault();
  const menuEstimatedHeight = 390;
  const menuEstimatedWidth = 185;
  contextMenuX.value = Math.max(10, Math.min(e.clientX + 5, window.innerWidth - menuEstimatedWidth));
  contextMenuY.value = Math.max(10, Math.min(e.clientY, window.innerHeight - menuEstimatedHeight));
  isContextMenuOpen.value = true;
}

function toggleVisibility(id: ActivityTabId) {
  const { newHidden, newCurrentTab } = toggleActivityTabVisibility(
    hiddenTabs.value,
    tabOrder.value,
    id,
    props.currentTab
  );
  hiddenTabs.value = newHidden;
  try {
    localStorage.setItem(ACTIVITY_BAR_HIDDEN_KEY, JSON.stringify([...newHidden]));
  } catch {}

  if (newCurrentTab !== props.currentTab) {
    emit("select-tab", newCurrentTab);
  }
}

function handleResetOrder() {
  tabOrder.value = [...DEFAULT_ACTIVITY_TAB_ORDER];
  hiddenTabs.value = new Set();
  try {
    localStorage.removeItem(ACTIVITY_BAR_ORDER_KEY);
    localStorage.removeItem(ACTIVITY_BAR_HIDDEN_KEY);
  } catch {}
  isContextMenuOpen.value = false;
}

function handleGlobalClick() {
  isContextMenuOpen.value = false;
}

onMounted(() => {
  fetchStatus().catch(() => {});
  window.addEventListener("click", handleGlobalClick);
});

onUnmounted(() => {
  window.removeEventListener("click", handleGlobalClick);
});
</script>

<style scoped>
.activity-bar {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  width: 50px;
  min-width: 50px;
  height: 100%;
  background: var(--card-bg);
  border-right: 1px solid var(--card-border);
  box-sizing: border-box;
  padding: 0.5rem 0;
  user-select: none;
  -webkit-user-select: none;
  z-index: 50;
  backdrop-filter: blur(8px);
}

.activity-top,
.activity-bottom {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.35rem;
}

.activity-btn {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  background: transparent;
  border: none;
  border-radius: 8px;
  color: var(--text-muted);
  cursor: pointer;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  padding: 0;
}

.activity-btn:hover {
  background: var(--input-bg);
  color: var(--text-main);
  transform: translateY(-1px);
}

.activity-btn.active {
  color: var(--accent-cyan);
  background: rgba(0, 229, 255, 0.12);
}

.activity-btn.is-dragging {
  opacity: 0.35;
  background: rgba(255, 255, 255, 0.05);
}

/* VS Code 风格的横向插入指示线 */
.activity-btn.drop-indicator-top::before,
.activity-btn.drop-indicator-bottom::after {
  content: "";
  position: absolute;
  left: 6px;
  right: 6px;
  height: 2px;
  background: var(--accent-cyan, #00e5ff);
  box-shadow: 0 0 8px var(--accent-cyan, #00e5ff);
  border-radius: 1px;
  z-index: 30;
  pointer-events: none;
}

.activity-btn.drop-indicator-top::before {
  top: -2px;
}

.activity-btn.drop-indicator-bottom::after {
  bottom: -2px;
}

.activity-indicator {
  position: absolute;
  left: -5px;
  top: 8px;
  bottom: 8px;
  width: 3px;
  background: var(--accent-cyan);
  border-radius: 0 4px 4px 0;
  box-shadow: 0 0 8px var(--accent-cyan);
}

.activity-icon {
  font-size: 1.25rem;
  line-height: 1;
}

.activity-badge {
  position: absolute;
  top: 3px;
  right: 2px;
  background: var(--success, #10b981);
  color: #ffffff;
  font-size: 0.62rem;
  font-weight: 800;
  padding: 0.05rem 0.3rem;
  border-radius: 9999px;
  line-height: 1.1;
  box-shadow: 0 2px 5px rgba(0, 0, 0, 0.3);
}

.activity-badge.badge-tools,
.activity-badge.badge-scm {
  background: var(--success, #10b981);
  color: #ffffff;
  font-weight: 800;
}

/* Activity Bar Context Menu */
.activity-context-menu {
  position: fixed;
  z-index: var(--z-context-menu-popover, 1001);
  background: var(--card-bg, #111827);
  border: 1px solid var(--card-border, #1f2937);
  box-shadow:
    0 10px 25px -5px rgba(0, 0, 0, 0.45),
    0 8px 10px -6px rgba(0, 0, 0, 0.3);
  border-radius: 6px;
  padding: 4px;
  min-width: 175px;
  font-size: 0.75rem;
  color: var(--text-main, #f3f4f6);
  user-select: none;
  animation: activity-context-menu-fade 0.12s ease-out;
}

:root[data-theme="light"] .activity-context-menu {
  box-shadow:
    0 10px 25px -5px rgba(0, 0, 0, 0.12),
    0 8px 10px -6px rgba(0, 0, 0, 0.08);
}

@keyframes activity-context-menu-fade {
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

:root[data-theme="light"] .context-menu-header {
  border-bottom-color: #e2e8f0;
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

.activity-context-menu-item {
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

.activity-context-menu-item:hover:not(.disabled) {
  background: rgba(0, 229, 255, 0.12);
  color: var(--accent-cyan, #00e5ff);
}

:root[data-theme="light"] .activity-context-menu-item:hover:not(.disabled) {
  background: #e0f2fe;
  color: #0284c7;
}

.activity-context-menu-item.disabled {
  opacity: 0.45;
  cursor: not-allowed;
}
</style>
