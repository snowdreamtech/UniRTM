<template>
  <transition name="modal-fade">
    <div v-if="isOpen" class="shortcuts-modal-overlay" @click="handleOverlayClick" @keydown.esc="close" tabindex="-1">
      <div class="shortcuts-modal-container" role="dialog" aria-modal="true" @click.stop>
        <!-- Modal Header -->
        <div class="modal-header">
          <div class="header-title">
            <span class="codicon codicon-keyboard header-icon"></span>
            <h2>{{ t("shortcuts.modal_title") || "Keyboard Shortcuts (VS Code Aligned)" }}</h2>
          </div>
          <button class="close-btn" :title="t('common.close') || 'Close (Esc)'" @click="close">
            <span class="codicon codicon-close"></span>
          </button>
        </div>

        <!-- Search & Filter Bar -->
        <div class="filter-bar">
          <div class="search-input-wrapper">
            <span class="codicon codicon-search search-icon"></span>
            <input
              ref="searchInputRef"
              v-model="searchQuery"
              type="text"
              class="search-input"
              :placeholder="t('shortcuts.search_placeholder') || 'Search shortcuts or command names...'"
              @keydown.esc.stop="close"
            />
            <button
              v-if="searchQuery"
              class="clear-search-btn"
              :title="t('shortcuts.clear_search') || 'Clear search'"
              @click="searchQuery = ''"
            >
              <span class="codicon codicon-close"></span>
            </button>
          </div>

          <div class="category-tabs">
            <button
              v-for="cat in categories"
              :key="cat.key"
              class="cat-tab"
              :class="{ active: activeCategory === cat.key }"
              @click="activeCategory = cat.key"
            >
              {{ cat.label }}
            </button>
          </div>
        </div>

        <!-- Shortcuts Table -->
        <div class="table-container">
          <table class="shortcuts-table">
            <thead>
              <tr>
                <th class="col-command">{{ t("shortcuts.col_command") || "Command" }}</th>
                <th class="col-category">{{ t("shortcuts.col_category") || "Category" }}</th>
                <th class="col-keys">{{ t("shortcuts.col_keybinding") || "Keybinding" }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in filteredShortcuts" :key="item.id" class="shortcut-row">
                <td class="col-command">
                  <div class="cmd-info">
                    <div class="cmd-title">{{ getShortcutTitle(item) }}</div>
                    <div class="cmd-desc">{{ getShortcutDesc(item) }}</div>
                  </div>
                </td>
                <td class="col-category">
                  <span class="cat-badge">{{ getCategoryLabel(item.category) }}</span>
                </td>
                <td class="col-keys">
                  <div class="keys-wrapper">
                    <template v-for="(chord, cIdx) in splitKeys(formatShortcut(item.effectiveKey))" :key="cIdx">
                      <span v-if="cIdx > 0" class="chord-separator">+</span>
                      <kbd class="key-pill" :class="{ 'custom-pill': item.isCustom }">{{ chord }}</kbd>
                    </template>
                  </div>
                </td>
              </tr>
              <tr v-if="filteredShortcuts.length === 0" class="empty-row">
                <td colspan="3">
                  <div class="empty-state">
                    <span class="codicon codicon-search empty-icon"></span>
                    <p>{{ t("shortcuts.no_matches") || "No matching shortcuts found" }}</p>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Footer -->
        <div class="modal-footer">
          <div class="footer-hint">
            <span class="hint-pill"
              ><kbd>{{ isMac ? "⌘K ⌘S" : "Ctrl+K Ctrl+S" }}</kbd>
              {{ t("shortcuts.hint_shortcuts") || "Shortcuts" }}</span
            >
            <span class="hint-pill"
              ><kbd>{{ isMac ? "⌘⇧P" : "Ctrl+Shift+P" }}</kbd>
              {{ t("shortcuts.hint_command_palette") || "Command Palette" }}</span
            >
            <span class="hint-pill"><kbd>ESC</kbd> {{ t("common.close") || "Close" }}</span>
          </div>
          <div class="footer-right">
            <button class="footer-settings-btn" @click="$emit('open-settings')">
              <span class="codicon codicon-settings-gear"></span>
              <span>{{ t("shortcuts.customize_in_settings") || "Customize in Settings" }}</span>
            </button>
            <div class="footer-count">
              {{
                t("shortcuts.count_summary", { count: filteredShortcuts.length }) ||
                filteredShortcuts.length + " shortcuts"
              }}
            </div>
          </div>
        </div>
      </div>
    </div>
  </transition>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from "vue";
import { useKeyboardShortcuts, getShortcutTitle, getShortcutDesc } from "../../composables/useKeyboardShortcuts";
import { t } from "../../i18n";
import { isMac, formatShortcut } from "../../utils/platform";

const props = defineProps<{
  isOpen: boolean;
}>();

const emit = defineEmits<{
  (e: "close"): void;
  (e: "open-settings"): void;
}>();

const { activeShortcuts } = useKeyboardShortcuts();

const searchQuery = ref("");
const activeCategory = ref("all");
const searchInputRef = ref<HTMLInputElement | null>(null);

const categories = computed(() => [
  { key: "all", label: t("shortcuts.cat_all") || "All" },
  { key: "general", label: t("shortcuts.cat_general") || "General" },
  { key: "panel", label: t("shortcuts.cat_panel") || "Panel & Terminal" },
  { key: "view", label: t("shortcuts.cat_view") || "Views" },
  { key: "editor", label: t("shortcuts.cat_editor") || "Editor" },
]);

function getCategoryLabel(catKey: string): string {
  const found = categories.value.find((c) => c.key === catKey);
  return found ? found.label : catKey;
}

function splitKeys(keys: string): string[] {
  return keys.split(" ");
}

const filteredShortcuts = computed(() => {
  const q = searchQuery.value.trim().toLowerCase();
  const cat = activeCategory.value;

  return activeShortcuts.value.filter((item) => {
    if (cat !== "all" && item.category !== cat) {
      return false;
    }
    if (!q) return true;

    const title = getShortcutTitle(item).toLowerCase();
    const desc = getShortcutDesc(item).toLowerCase();
    const key = item.effectiveKey.toLowerCase();
    const id = item.id.toLowerCase();
    return key.includes(q) || title.includes(q) || desc.includes(q) || id.includes(q);
  });
});

function close() {
  emit("close");
}

function handleOverlayClick(event: MouseEvent) {
  if (event.target === event.currentTarget) {
    close();
  }
}

watch(
  () => props.isOpen,
  (val) => {
    if (val) {
      searchQuery.value = "";
      activeCategory.value = "all";
      nextTick(() => {
        searchInputRef.value?.focus();
      });
    }
  }
);
</script>

<style scoped>
.shortcuts-modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 9999;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.65);
  backdrop-filter: blur(8px);
}

.shortcuts-modal-container {
  width: 90vw;
  max-width: 820px;
  max-height: 85vh;
  display: flex;
  flex-direction: column;
  background: #0d1424;
  border: 1px solid rgba(0, 229, 255, 0.25);
  border-radius: 12px;
  box-shadow:
    0 20px 50px rgba(0, 0, 0, 0.7),
    0 0 24px rgba(0, 229, 255, 0.15);
  overflow: hidden;
  color: #e2e8f0;
  font-family: inherit;
}

/* Light Mode Overrides */
:root[data-theme="light"] .shortcuts-modal-container {
  background: #ffffff;
  border: 1px solid rgba(15, 23, 42, 0.12);
  box-shadow:
    0 20px 50px rgba(15, 23, 42, 0.15),
    0 4px 16px rgba(15, 23, 42, 0.08);
  color: #0f172a;
}

/* Header */
.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

:root[data-theme="light"] .modal-header {
  border-bottom: 1px solid rgba(15, 23, 42, 0.08);
}

.header-title {
  display: flex;
  align-items: center;
  gap: 10px;
}

.header-icon {
  font-size: 20px;
  color: #00e5ff;
}

:root[data-theme="light"] .header-icon {
  color: #0284c7;
}

.header-title h2 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #ffffff;
}

:root[data-theme="light"] .header-title h2 {
  color: #0f172a;
}

.close-btn {
  background: transparent;
  border: none;
  color: #94a3b8;
  padding: 6px;
  border-radius: 6px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s ease;
}

.close-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  color: #ffffff;
}

:root[data-theme="light"] .close-btn:hover {
  background: rgba(15, 23, 42, 0.06);
  color: #0f172a;
}

/* Filter Bar */
.filter-bar {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 14px 20px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  background: rgba(0, 0, 0, 0.15);
}

:root[data-theme="light"] .filter-bar {
  border-bottom: 1px solid rgba(15, 23, 42, 0.08);
  background: rgba(248, 250, 252, 0.8);
}

.search-input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
  width: 100%;
}

.search-icon {
  position: absolute;
  left: 12px;
  font-size: 16px;
  color: #94a3b8;
  pointer-events: none;
}

.search-input {
  width: 100%;
  padding: 10px 36px 10px 38px;
  background: #151f33;
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 8px;
  color: #ffffff;
  font-size: 14px;
  outline: none;
  transition:
    border-color 0.2s,
    box-shadow 0.2s;
}

:root[data-theme="light"] .search-input {
  background: #ffffff;
  border: 1px solid #cbd5e1;
  color: #0f172a;
}

.search-input:focus {
  border-color: #00e5ff;
  box-shadow: 0 0 0 2px rgba(0, 229, 255, 0.2);
}

:root[data-theme="light"] .search-input:focus {
  border-color: #0284c7;
  box-shadow: 0 0 0 2px rgba(2, 132, 199, 0.2);
}

.clear-search-btn {
  position: absolute;
  right: 10px;
  background: transparent;
  border: none;
  color: #94a3b8;
  cursor: pointer;
  display: flex;
  align-items: center;
}

.category-tabs {
  display: flex;
  gap: 8px;
  overflow-x: auto;
}

.cat-tab {
  padding: 6px 12px;
  font-size: 13px;
  border-radius: 6px;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.08);
  color: #94a3b8;
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.15s ease;
}

:root[data-theme="light"] .cat-tab {
  background: #f1f5f9;
  border: 1px solid #e2e8f0;
  color: #475569;
}

.cat-tab:hover {
  color: #ffffff;
  background: rgba(255, 255, 255, 0.1);
}

:root[data-theme="light"] .cat-tab:hover {
  color: #0f172a;
  background: #e2e8f0;
}

.cat-tab.active {
  background: rgba(0, 229, 255, 0.15);
  border-color: #00e5ff;
  color: #00e5ff;
  font-weight: 500;
}

:root[data-theme="light"] .cat-tab.active {
  background: #e0f2fe;
  border-color: #0284c7;
  color: #0284c7;
}

/* Table */
.table-container {
  flex: 1;
  overflow-y: auto;
  padding: 0;
}

.shortcuts-table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
}

.shortcuts-table th {
  padding: 10px 20px;
  font-size: 12px;
  font-weight: 600;
  color: #94a3b8;
  text-transform: uppercase;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  position: sticky;
  top: 0;
  background: #0d1424;
  z-index: 1;
}

:root[data-theme="light"] .shortcuts-table th {
  color: #64748b;
  border-bottom: 1px solid #e2e8f0;
  background: #ffffff;
}

.col-command {
  width: 55%;
}

.col-category {
  width: 20%;
}

.col-keys {
  width: 25%;
  text-align: right;
  padding-right: 20px;
}

.shortcut-row {
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  content-visibility: auto;
  contain-intrinsic-size: 0 48px;
  transition: background-color 0.15s ease;
}

:root[data-theme="light"] .shortcut-row {
  border-bottom: 1px solid rgba(15, 23, 42, 0.05);
}

.shortcut-row:hover {
  background: rgba(0, 229, 255, 0.06);
}

:root[data-theme="light"] .shortcut-row:hover {
  background: #f8fafc;
}

.shortcut-row td {
  padding: 12px 20px;
  vertical-align: middle;
}

.cmd-title {
  font-size: 14px;
  font-weight: 500;
  color: #ffffff;
}

:root[data-theme="light"] .cmd-title {
  color: #0f172a;
}

.cmd-desc {
  font-size: 12px;
  color: #94a3b8;
  margin-top: 2px;
}

:root[data-theme="light"] .cmd-desc {
  color: #475569;
}

.cat-badge {
  font-size: 11px;
  padding: 3px 8px;
  border-radius: 4px;
  background: rgba(255, 255, 255, 0.06);
  color: #94a3b8;
}

:root[data-theme="light"] .cat-badge {
  background: #f1f5f9;
  color: #475569;
}

.keys-wrapper {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 6px;
  flex-wrap: wrap;
}

.chord-separator {
  color: #64748b;
  font-size: 12px;
}

.key-pill {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 12px;
  font-weight: 600;
  padding: 3px 8px;
  border-radius: 4px;
  background: #1e293b;
  border: 1px solid rgba(255, 255, 255, 0.18);
  box-shadow: 0 2px 0 rgba(0, 0, 0, 0.3);
  color: #00e5ff;
}

:root[data-theme="light"] .key-pill {
  background: #f8fafc;
  border: 1px solid #cbd5e1;
  box-shadow: 0 2px 0 rgba(15, 23, 42, 0.1);
  color: #0284c7;
}

.empty-row td {
  padding: 40px 20px;
  text-align: center;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  color: #94a3b8;
}

.empty-icon {
  font-size: 32px;
  margin-bottom: 8px;
  opacity: 0.5;
}

/* Footer */
.modal-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 20px;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
  background: rgba(0, 0, 0, 0.2);
  font-size: 12px;
  color: #94a3b8;
}

:root[data-theme="light"] .modal-footer {
  border-top: 1px solid rgba(15, 23, 42, 0.08);
  background: #f8fafc;
  color: #64748b;
}

.key-pill.custom-pill {
  color: #38bdf8;
  border-color: rgba(56, 189, 248, 0.4);
}

:root[data-theme="light"] .key-pill.custom-pill {
  color: #0284c7;
  border-color: #38bdf8;
}

.footer-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.footer-settings-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: transparent;
  border: none;
  color: #00e5ff;
  font-size: 12px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 4px;
  transition: all 0.15s ease;
}

:root[data-theme="light"] .footer-settings-btn {
  color: #0284c7;
}

.footer-settings-btn:hover {
  background: rgba(0, 229, 255, 0.1);
}

:root[data-theme="light"] .footer-settings-btn:hover {
  background: #e0f2fe;
}

.footer-hint {
  display: flex;
  align-items: center;
  gap: 12px;
}

.hint-pill kbd {
  font-family: inherit;
  padding: 2px 6px;
  border-radius: 3px;
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.15);
  color: #e2e8f0;
}

:root[data-theme="light"] .hint-pill kbd {
  background: #e2e8f0;
  border: 1px solid #cbd5e1;
  color: #0f172a;
}

/* Transition animations */
.modal-fade-enter-active,
.modal-fade-leave-active {
  transition: opacity 0.2s ease;
}

.modal-fade-enter-from,
.modal-fade-leave-to {
  opacity: 0;
}

.modal-fade-enter-active .shortcuts-modal-container {
  transition: transform 0.2s cubic-bezier(0.16, 1, 0.3, 1);
}

.modal-fade-enter-from .shortcuts-modal-container {
  transform: translateY(-16px) scale(0.98);
}
</style>
