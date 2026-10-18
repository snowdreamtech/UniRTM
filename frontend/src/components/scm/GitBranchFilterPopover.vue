<!-- Copyright (c) 2026 SnowdreamTech. All rights reserved. -->
<!-- Licensed under the MIT License. See LICENSE file in the project root for full license information. -->

<template>
  <div class="branch-filter-popover-container" ref="containerRef">
    <!-- Trigger Pill -->
    <button
      class="branch-filter-trigger"
      :class="{ active: isOpen }"
      type="button"
      :title="t('scm.branch.switch')"
      @click.stop="togglePopover"
    >
      <span class="codicon codicon-git-branch branch-icon"></span>
      <span class="branch-display-text">{{ displayLabel }}</span>
      <span v-if="selectedCountBadge" class="branch-count-badge">{{ selectedCountBadge }}</span>
      <span class="codicon codicon-chevron-down arrow-icon" :class="{ open: isOpen }"></span>
    </button>

    <!-- Teleported Popover -->
    <Teleport to="body">
      <div
        v-if="isOpen"
        class="branch-popover-overlay"
        @click="closePopover"
        @contextmenu.prevent="closePopover"
      >
        <div
          class="branch-popover-dropdown"
          :style="popoverStyle"
          @click.stop
        >
          <!-- Header & Search Box -->
          <div class="popover-header">
            <div class="popover-search-wrap">
              <span class="codicon codicon-search search-icon"></span>
              <input
                ref="searchInputRef"
                v-model="searchQuery"
                type="text"
                class="popover-search-input"
                :placeholder="t('scm.graph.branch_filter.filter_placeholder')"
                @keydown.esc="closePopover"
              />
              <button
                v-if="searchQuery"
                class="search-clear-btn"
                @click="searchQuery = ''"
              >
                <span class="codicon codicon-close"></span>
              </button>
            </div>

            <!-- Quick Action Toolbar -->
            <div class="popover-quick-actions">
              <button
                type="button"
                class="quick-btn"
                :class="{ active: isAllSelected }"
                @click="selectAllBranches"
              >
                {{ t('scm.graph.branch_filter.all') }}
              </button>
              <button
                type="button"
                class="quick-btn"
                :class="{ active: isCurrentSelectedOnly }"
                @click="selectCurrentBranch"
              >
                {{ t('scm.graph.branch_filter.current') }}
              </button>
              <button
                type="button"
                class="quick-btn"
                @click="clearSelection"
              >
                {{ t('scm.graph.branch_filter.clear') }}
              </button>
            </div>
          </div>

          <!-- Branch List Items -->
          <div class="popover-branch-list">
            <!-- 1. All Branches Item -->
            <div
              v-if="!searchQuery || 'all branches'.includes(searchQuery.toLowerCase()) || t('scm.graph.all_branches').toLowerCase().includes(searchQuery.toLowerCase())"
              class="branch-list-item"
              :class="{ selected: isAllSelected }"
              @click="toggleAll"
            >
              <span class="custom-checkbox" :class="{ checked: isAllSelected }">
                <span v-if="isAllSelected" class="codicon codicon-check check-icon"></span>
              </span>
              <span class="codicon codicon-git-branch-create item-type-icon"></span>
              <span class="item-name font-semibold">{{ t('scm.graph.all_branches') }}</span>
            </div>

            <div v-if="!searchQuery" class="list-divider"></div>

            <!-- 2. Local Branches Section -->
            <div v-if="filteredLocalBranches.length > 0" class="branch-section">
              <div class="section-title">{{ t('scm.graph.branch_filter.local') }} ({{ filteredLocalBranches.length }})</div>
              <div
                v-for="b in filteredLocalBranches"
                :key="b.name"
                class="branch-list-item"
                :class="{ selected: isBranchSelected(b.name) }"
                @click="toggleBranch(b.name)"
              >
                <span class="custom-checkbox" :class="{ checked: isBranchSelected(b.name) }">
                  <span v-if="isBranchSelected(b.name)" class="codicon codicon-check check-icon"></span>
                </span>
                <span class="codicon codicon-git-branch item-type-icon"></span>
                <span class="item-name" :class="{ 'font-bold': b.current }">{{ b.name }}</span>
                <span v-if="b.current" class="current-badge">{{ t('scm.graph.current_badge') }}</span>
              </div>
            </div>

            <!-- 3. Remote Branches Section -->
            <div v-if="filteredRemoteBranches.length > 0" class="branch-section">
              <div class="section-title">{{ t('scm.graph.branch_filter.remote') }} ({{ filteredRemoteBranches.length }})</div>
              <div
                v-for="b in filteredRemoteBranches"
                :key="b.name"
                class="branch-list-item"
                :class="{ selected: isBranchSelected(b.name) }"
                @click="toggleBranch(b.name)"
              >
                <span class="custom-checkbox" :class="{ checked: isBranchSelected(b.name) }">
                  <span v-if="isBranchSelected(b.name)" class="codicon codicon-check check-icon"></span>
                </span>
                <span class="codicon codicon-cloud item-type-icon remote"></span>
                <span class="item-name">{{ b.name }}</span>
              </div>
            </div>

            <!-- Empty State -->
            <div
              v-if="filteredLocalBranches.length === 0 && filteredRemoteBranches.length === 0 && searchQuery"
              class="empty-state"
            >
              <span>{{ t('scm.graph.branch_filter.no_branches_found') }}</span>
            </div>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, onMounted, onBeforeUnmount } from "vue";
import type { GitBranchItem } from "../../composables/useGit";
import { t } from "../../i18n";

const props = withDefaults(
  defineProps<{
    modelValue: string[];
    branches: GitBranchItem[];
    currentBranch?: string;
  }>(),
  {
    modelValue: () => ["all"],
    branches: () => [],
    currentBranch: "HEAD",
  }
);

const emit = defineEmits<{
  (e: "update:modelValue", value: string[]): void;
  (e: "change", value: string[]): void;
}>();

const containerRef = ref<HTMLElement | null>(null);
const searchInputRef = ref<HTMLInputElement | null>(null);
const isOpen = ref(false);
const searchQuery = ref("");
const triggerRect = ref<DOMRect | null>(null);

const isAllSelected = computed(() => {
  return props.modelValue.length === 0 || props.modelValue.includes("all");
});

const isCurrentSelectedOnly = computed(() => {
  if (isAllSelected.value) return false;
  const current = props.currentBranch;
  return (
    props.modelValue.length === 1 &&
    current &&
    (props.modelValue[0] === current || props.modelValue[0] === `refs/heads/${current}`)
  );
});

const displayLabel = computed(() => {
  if (isAllSelected.value) {
    return t("scm.graph.all_branches");
  }
  if (props.modelValue.length === 1) {
    const single = props.modelValue[0];
    const isCurrent =
      single === props.currentBranch ||
      single === `refs/heads/${props.currentBranch}`;
    return isCurrent ? `${single} (${t("scm.graph.current_badge")})` : single;
  }
  return t("scm.graph.branch_filter.multiple_selected", {
    count: props.modelValue.length,
  });
});

const selectedCountBadge = computed(() => {
  if (isAllSelected.value || props.modelValue.length <= 1) return null;
  return String(props.modelValue.length);
});

const filteredLocalBranches = computed(() => {
  const q = searchQuery.value.trim().toLowerCase();
  const locals = props.branches.filter((b) => !b.remote);
  if (!q) return locals;
  return locals.filter((b) => b.name.toLowerCase().includes(q));
});

const filteredRemoteBranches = computed(() => {
  const q = searchQuery.value.trim().toLowerCase();
  const remotes = props.branches.filter((b) => b.remote);
  if (!q) return remotes;
  return remotes.filter((b) => b.name.toLowerCase().includes(q));
});

function isBranchSelected(branchName: string): boolean {
  if (isAllSelected.value) return false;
  return props.modelValue.includes(branchName);
}

function updateSelection(newSelection: string[]) {
  emit("update:modelValue", newSelection);
  emit("change", newSelection);
}

function selectAllBranches() {
  updateSelection(["all"]);
}

function selectCurrentBranch() {
  const target = props.currentBranch || "HEAD";
  updateSelection([target]);
}

function clearSelection() {
  // If cleared, default back to showing all
  updateSelection(["all"]);
}

function toggleAll() {
  if (isAllSelected.value) {
    // If was all, select current branch
    selectCurrentBranch();
  } else {
    selectAllBranches();
  }
}

function toggleBranch(branchName: string) {
  if (isAllSelected.value) {
    // Switching from "all" to this single branch
    updateSelection([branchName]);
    return;
  }

  const current = [...props.modelValue];
  const idx = current.indexOf(branchName);
  if (idx !== -1) {
    current.splice(idx, 1);
  } else {
    current.push(branchName);
  }

  if (current.length === 0) {
    // Fallback to all if none selected
    updateSelection(["all"]);
  } else {
    updateSelection(current);
  }
}

function togglePopover() {
  if (isOpen.value) {
    closePopover();
  } else {
    openPopover();
  }
}

function openPopover() {
  if (containerRef.value) {
    triggerRect.value = containerRef.value.getBoundingClientRect();
  }
  isOpen.value = true;
  nextTick(() => {
    searchInputRef.value?.focus();
  });
}

function closePopover() {
  isOpen.value = false;
  searchQuery.value = "";
}

const POPOVER_WIDTH = 280;
const POPOVER_MAX_HEIGHT = 380;

const popoverStyle = computed(() => {
  if (!triggerRect.value) return {};

  let left = triggerRect.value.left;
  let top = triggerRect.value.bottom + 4;

  if (left + POPOVER_WIDTH > window.innerWidth - 10) {
    left = window.innerWidth - POPOVER_WIDTH - 10;
  }
  if (top + POPOVER_MAX_HEIGHT > window.innerHeight - 10) {
    top = Math.max(10, triggerRect.value.top - POPOVER_MAX_HEIGHT - 4);
  }

  return {
    left: `${Math.max(10, left)}px`,
    top: `${Math.max(10, top)}px`,
    width: `${POPOVER_WIDTH}px`,
  };
});

function handleWindowResize() {
  if (isOpen.value && containerRef.value) {
    triggerRect.value = containerRef.value.getBoundingClientRect();
  }
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === "Escape" && isOpen.value) {
    closePopover();
  }
}

onMounted(() => {
  window.addEventListener("resize", handleWindowResize);
  window.addEventListener("keydown", handleKeydown);
});

onBeforeUnmount(() => {
  window.removeEventListener("resize", handleWindowResize);
  window.removeEventListener("keydown", handleKeydown);
});
</script>

<style scoped>
.branch-filter-popover-container {
  position: relative;
  display: inline-block;
}

.branch-filter-trigger {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 28px;
  padding: 0 10px;
  background: var(--input-bg, #1e293b);
  border: 1px solid var(--input-border, #334155);
  border-radius: 4px;
  color: var(--text-main, #e2e8f0);
  font-size: 12px;
  font-family: inherit;
  cursor: pointer;
  transition: all 0.15s ease;
  user-select: none;
  max-width: 220px;
}

.branch-filter-trigger:hover,
.branch-filter-trigger.active {
  background: var(--hover-bg, rgba(56, 189, 248, 0.12));
  border-color: var(--accent-cyan, #00e5ff);
  color: var(--text-main, #ffffff);
}

:root[data-theme="light"] .branch-filter-trigger {
  background: #ffffff;
  border-color: #cbd5e1;
  color: #1e293b;
}

:root[data-theme="light"] .branch-filter-trigger:hover,
:root[data-theme="light"] .branch-filter-trigger.active {
  background: #f1f5f9;
  border-color: #0284c7;
  color: #0284c7;
}

.branch-icon {
  font-size: 13px;
  color: var(--accent-cyan, #00e5ff);
  flex-shrink: 0;
}

.branch-display-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-weight: 500;
}

.branch-count-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 16px;
  height: 16px;
  padding: 0 4px;
  font-size: 10px;
  font-weight: 600;
  background: var(--accent-cyan, #00e5ff);
  color: #0b1320;
  border-radius: 8px;
  line-height: 1;
}

.arrow-icon {
  font-size: 11px;
  color: var(--text-muted, #94a3b8);
  transition: transform 0.15s ease;
  flex-shrink: 0;
}

.arrow-icon.open {
  transform: rotate(180deg);
}

/* Popover Overlay */
.branch-popover-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 9999;
  background: transparent;
}

/* Popover Dropdown Card */
.branch-popover-dropdown {
  position: fixed;
  background: var(--card-bg, #0d1626);
  border: 1px solid var(--card-border, #1f2e48);
  border-radius: 6px;
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.45);
  display: flex;
  flex-direction: column;
  z-index: 10000;
  overflow: hidden;
  font-size: 12px;
  color: var(--text-main, #e2e8f0);
}

:root[data-theme="light"] .branch-popover-dropdown {
  background: #ffffff;
  border-color: #e2e8f0;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.12);
  color: #1e293b;
}

/* Header & Search */
.popover-header {
  padding: 8px;
  border-bottom: 1px solid var(--card-border, #1f2e48);
  background: var(--surface-bg, rgba(255, 255, 255, 0.02));
}

:root[data-theme="light"] .popover-header {
  border-bottom-color: #f1f5f9;
  background: #f8fafc;
}

.popover-search-wrap {
  position: relative;
  display: flex;
  align-items: center;
}

.popover-search-wrap .search-icon {
  position: absolute;
  left: 8px;
  font-size: 12px;
  color: var(--text-muted, #94a3b8);
  pointer-events: none;
}

.popover-search-input {
  width: 100%;
  height: 28px;
  padding: 0 24px 0 26px;
  background: var(--input-bg, #111c30);
  border: 1px solid var(--input-border, #253552);
  border-radius: 4px;
  color: var(--text-main, #f8fafc);
  font-size: 11px;
  font-family: inherit;
  outline: none;
}

.popover-search-input:focus {
  border-color: var(--accent-cyan, #00e5ff);
}

:root[data-theme="light"] .popover-search-input {
  background: #ffffff;
  border-color: #cbd5e1;
  color: #1e293b;
}

.search-clear-btn {
  position: absolute;
  right: 6px;
  background: transparent;
  border: none;
  color: var(--text-muted, #94a3b8);
  cursor: pointer;
  padding: 2px;
}

/* Quick Actions */
.popover-quick-actions {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-top: 6px;
}

.quick-btn {
  flex: 1;
  height: 22px;
  background: var(--btn-secondary-bg, rgba(255, 255, 255, 0.05));
  border: 1px solid var(--card-border, #23344f);
  border-radius: 3px;
  color: var(--text-muted, #94a3b8);
  font-size: 11px;
  cursor: pointer;
  transition: all 0.12s ease;
}

.quick-btn:hover,
.quick-btn.active {
  background: var(--hover-bg, rgba(56, 189, 248, 0.15));
  border-color: var(--accent-cyan, #00e5ff);
  color: var(--accent-cyan, #00e5ff);
}

:root[data-theme="light"] .quick-btn {
  background: #f1f5f9;
  border-color: #e2e8f0;
  color: #64748b;
}

:root[data-theme="light"] .quick-btn:hover,
:root[data-theme="light"] .quick-btn.active {
  background: #e0f2fe;
  border-color: #0284c7;
  color: #0284c7;
}

/* Branch List */
.popover-branch-list {
  max-height: 280px;
  overflow-y: auto;
  padding: 4px 0;
}

.list-divider {
  height: 1px;
  background: var(--card-border, #1f2e48);
  margin: 4px 0;
}

:root[data-theme="light"] .list-divider {
  background: #f1f5f9;
}

.branch-section {
  padding: 2px 0;
}

.section-title {
  padding: 4px 12px;
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
  color: var(--text-muted, #64748b);
  letter-spacing: 0.5px;
}

.branch-list-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  cursor: pointer;
  transition: background 0.12s ease;
}

.branch-list-item:hover {
  background: var(--hover-bg, rgba(56, 189, 248, 0.12));
}

:root[data-theme="light"] .branch-list-item:hover {
  background: #f1f5f9;
}

.branch-list-item.selected {
  background: rgba(56, 189, 248, 0.08);
}

/* Custom Checkbox */
.custom-checkbox {
  width: 14px;
  height: 14px;
  border: 1px solid var(--input-border, #475569);
  border-radius: 3px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--input-bg, #0f172a);
  flex-shrink: 0;
  transition: all 0.12s ease;
}

.custom-checkbox.checked {
  background: var(--accent-cyan, #00e5ff);
  border-color: var(--accent-cyan, #00e5ff);
}

:root[data-theme="light"] .custom-checkbox {
  background: #ffffff;
  border-color: #cbd5e1;
}

:root[data-theme="light"] .custom-checkbox.checked {
  background: #0284c7;
  border-color: #0284c7;
}

.check-icon {
  font-size: 11px;
  color: #0b1320;
  font-weight: 900;
}

:root[data-theme="light"] .check-icon {
  color: #ffffff;
}

.item-type-icon {
  font-size: 13px;
  color: var(--text-muted, #94a3b8);
  flex-shrink: 0;
}

.item-type-icon.remote {
  color: #38bdf8;
}

.item-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.current-badge {
  font-size: 9px;
  padding: 1px 5px;
  border-radius: 3px;
  background: rgba(34, 197, 94, 0.15);
  color: #4ade80;
  border: 1px solid rgba(34, 197, 94, 0.3);
  font-weight: 600;
}

.empty-state {
  padding: 16px 12px;
  text-align: center;
  color: var(--text-muted, #64748b);
  font-size: 11px;
}
</style>
