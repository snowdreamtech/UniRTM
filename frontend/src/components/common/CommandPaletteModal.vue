<!-- Copyright (c) 2026 SnowdreamTech. All rights reserved. -->
<!-- Licensed under the MIT License. See LICENSE file in the project root for full license information. -->

<template>
  <transition name="palette-fade">
    <div
      v-if="isOpen"
      class="palette-overlay"
      id="modal-command-palette"
      @click="handleOverlayClick"
      @keydown.esc="close"
    >
      <div class="palette-container" @click.stop @contextmenu.prevent>
        <!-- Search Input Bar -->
        <div class="palette-search-row">
          <span class="palette-search-icon">
            <TechIcon id="workbench:search" :size="20" />
          </span>
          <input
            ref="inputRef"
            v-model="query"
            type="text"
            class="palette-input"
            :placeholder="placeholderText"
            @keydown.down.prevent="moveSelection(1)"
            @keydown.up.prevent="moveSelection(-1)"
            @keydown.enter.prevent="executeSelected"
          />
          <button v-if="query" class="palette-clear-btn" title="Clear" @click="query = ''">×</button>
          <kbd class="palette-esc-hint">ESC</kbd>
        </div>

        <!-- Command List -->
        <div class="palette-list-wrapper">
          <div v-if="filteredCommands.length === 0" class="palette-empty-state">
            <span class="empty-icon"><TechIcon id="workbench:search" :size="36" /></span>
            <span class="empty-text">{{ emptyText }}</span>
          </div>

          <div v-else class="palette-items-list">
            <div
              v-for="(item, index) in filteredCommands"
              :key="item.id"
              class="palette-item"
              :class="{ selected: selectedIndex === index }"
              @mouseenter="selectedIndex = index"
              @click="executeItem(item)"
            >
              <div class="item-left">
                <span class="item-icon">
                  <TechIcon v-if="item.techIcon" :id="item.techIcon" :size="20" />
                  <span v-else-if="item.codicon" class="codicon" :class="item.codicon"></span>
                  <span v-else class="codicon codicon-terminal"></span>
                </span>
                <div class="item-details">
                  <span class="item-title">{{ item.title }}</span>
                  <span v-if="item.desc" class="item-desc">{{ item.desc }}</span>
                </div>
              </div>

              <div class="item-right">
                <span v-if="item.category" class="category-pill">{{ item.category }}</span>
                <kbd v-if="item.shortcut" class="shortcut-pill">{{ formatShortcut(item.shortcut) }}</kbd>
                <span class="enter-arrow">↵</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Footer Bar -->
        <div class="palette-footer">
          <div class="footer-tips">
            <span class="tip-item"><kbd>↑</kbd> <kbd>↓</kbd> {{ navTipText }}</span>
            <span class="tip-item"><kbd>↵</kbd> {{ execTipText }}</span>
            <span class="tip-item"><kbd>ESC</kbd> {{ closeTipText }}</span>
          </div>
          <div class="footer-count">{{ countText }}</div>
        </div>
      </div>
    </div>
  </transition>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from "vue";
import { useRouter } from "vue-router";
import { t } from "../../i18n";
import TechIcon from "./TechIcon.vue";
import type { DrawerTab } from "../layout/BottomDrawer.vue";
import { isMac, formatShortcut } from "../../utils/platform";

export interface CommandItem {
  id: string;
  title: string;
  desc?: string;
  category: string;
  shortcut?: string;
  techIcon?: string;
  codicon?: string;
  action: () => void;
}

const props = defineProps<{
  isOpen: boolean;
}>();

const emit = defineEmits<{
  (e: "close"): void;
  (e: "select-tab", tabId: string): void;
  (e: "toggle-theme"): void;
  (e: "open-drawer", tab?: DrawerTab): void;
  (e: "toggle-scope"): void;
  (e: "open-shortcuts"): void;
  (e: "new-terminal"): void;
}>();

const router = useRouter();
const query = ref("");
const selectedIndex = ref(0);
const inputRef = ref<HTMLInputElement | null>(null);

const placeholderText = computed(() => {
  const shortcutHint = isMac ? "⌘K" : "Ctrl+K";
  return (
    t("cmdpalette.placeholder", { shortcut: shortcutHint }) ||
    `Search features, navigate views, or execute quick commands (${shortcutHint})...`
  );
});

const emptyText = computed(() => {
  return t("cmdpalette.empty") || "No matching features or commands found";
});

const navTipText = computed(() => t("cmdpalette.nav_tip") || "Navigate");
const execTipText = computed(() => t("cmdpalette.exec_tip") || "Execute");
const closeTipText = computed(() => t("cmdpalette.close_tip") || "Close");

const countText = computed(() => {
  const n = filteredCommands.value.length;
  return t("cmdpalette.available_commands", { n }) || `${n} commands available`;
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
      query.value = "";
      selectedIndex.value = 0;
      nextTick(() => {
        inputRef.value?.focus();
      });
    }
  }
);

const allCommands = computed<CommandItem[]>(() => {
  const catView = t("cmdpalette.cat_view") || "View";
  const catGit = "Git";
  const catDiag = t("cmdpalette.cat_diagnosis") || "Diagnosis";
  const catTool = t("cmdpalette.cat_tools") || "Tools";
  const catSys = t("cmdpalette.cat_system") || "System";

  return [
    // 1. Navigation views (VS Code aligned)
    {
      id: "nav-explorer",
      title: t("cmdpalette.cmd_explorer_title") || "File Explorer (Explorer)",
      desc: t("cmdpalette.cmd_explorer_desc") || "Browse workspace directories and view editor files",
      category: catView,
      shortcut: "Ctrl+Shift+E",
      techIcon: "workbench:folder",
      action: () => {
        emit("select-tab", "explorer");
        router.push("/explorer");
      },
    },
    {
      id: "nav-scm",
      title: t("cmdpalette.cmd_scm_title") || "Source Control (SCM / Git)",
      desc: t("cmdpalette.cmd_scm_desc") || "View changes, stage files, switch branches, and view git commit graphs",
      category: catGit,
      shortcut: "Ctrl+Shift+G",
      techIcon: "workbench:git",
      action: () => {
        emit("select-tab", "scm");
        router.push("/scm");
      },
    },
    {
      id: "nav-doctor",
      title: t("cmdpalette.cmd_doctor_title") || "System Health Diagnosis (Doctor)",
      desc: t("cmdpalette.cmd_doctor_desc") || "Probe core runtimes, shims routing, disk footprints, and API quotas",
      category: catDiag,
      shortcut: "Ctrl+Shift+D",
      techIcon: "workbench:doctor",
      action: () => {
        emit("select-tab", "doctor");
        router.push("/doctor");
      },
    },
    {
      id: "nav-marketplace",
      title: t("cmdpalette.cmd_marketplace_title") || "Extension Marketplace (Marketplace)",
      desc: t("cmdpalette.cmd_marketplace_desc") || "Discover and install themes, plugins, and ecosystem extensions",
      category: catView,
      shortcut: "Ctrl+Shift+X",
      techIcon: "workbench:marketplace",
      action: () => {
        emit("select-tab", "marketplace");
        router.push("/marketplace");
      },
    },
    {
      id: "nav-dashboard",
      title: t("cmdpalette.cmd_dashboard_title") || "Project Overview Dashboard (Dashboard)",
      desc: t("cmdpalette.cmd_dashboard_desc") || "UniRTM project status overview, runtime metrics, and quick actions",
      category: catView,
      shortcut: "Alt+1",
      techIcon: "workbench:dashboard",
      action: () => {
        emit("select-tab", "dashboard");
        router.push("/dashboard");
      },
    },
    {
      id: "nav-guard",
      title: t("cmdpalette.cmd_guard_title") || "Quality Guard Harness (Guard)",
      desc: t("cmdpalette.cmd_guard_desc") || "Pre-commit security scanning, leak interception, and format compliance",
      category: catDiag,
      shortcut: "Ctrl+Shift+Q",
      techIcon: "workbench:guard",
      action: () => {
        emit("select-tab", "guard");
        router.push("/guard");
      },
    },
    {
      id: "nav-tools",
      title: t("cmdpalette.cmd_tools_title") || "Toolchains & Runtimes (Tools)",
      desc: t("cmdpalette.cmd_tools_desc") || "Quickly manage Node.js, Python, Go, Rust and runtime versions",
      category: catTool,
      shortcut: "Alt+2",
      techIcon: "workbench:tools",
      action: () => {
        emit("select-tab", "tools");
        router.push("/tools");
      },
    },
    {
      id: "nav-tasks",
      title: t("cmdpalette.cmd_tasks_title") || "Project Automation Tasks (Tasks)",
      desc: t("cmdpalette.cmd_tasks_desc") || "One-click build, test, lint, and workflow pipeline tasks",
      category: catTool,
      shortcut: "Alt+3",
      techIcon: "workbench:tasks",
      action: () => {
        emit("select-tab", "tasks");
        router.push("/tasks");
      },
    },
    {
      id: "nav-environments",
      title: t("cmdpalette.cmd_environments_title") || "Environment Variables (Environments)",
      desc:
        t("cmdpalette.cmd_environments_desc") || "Configure and manage workspace and user global environment variables",
      category: catTool,
      shortcut: "Alt+4",
      techIcon: "workbench:environments",
      action: () => {
        emit("select-tab", "environments");
        router.push("/environments");
      },
    },
    {
      id: "nav-settings",
      title: t("cmdpalette.cmd_settings_title") || "Preferences & Settings (Settings)",
      desc: t("cmdpalette.cmd_settings_desc") || "Configure themes, network proxies, auto-update, and security keys",
      category: catSys,
      shortcut: "Ctrl+,",
      techIcon: "workbench:settings",
      action: () => {
        emit("select-tab", "settings");
        router.push("/settings");
      },
    },
    {
      id: "nav-about",
      title: t("cmdpalette.cmd_about_title") || "About UniRTM (About)",
      desc: t("cmdpalette.cmd_about_desc") || "View version build info, license information, and changelog",
      category: catSys,
      techIcon: "workbench:about",
      action: () => {
        emit("select-tab", "about");
        router.push("/about");
      },
    },

    // 2. Global Actions & Helpers
    {
      id: "act-shortcuts",
      title: t("cmdpalette.cmd_shortcuts_title") || "Keyboard Shortcuts Reference (VS Code Aligned)",
      desc: t("cmdpalette.cmd_shortcuts_desc") || "View full list of keyboard shortcuts aligned with VS Code",
      category: catSys,
      shortcut: "Ctrl+K Ctrl+S",
      codicon: "codicon-keyboard",
      action: () => {
        emit("open-shortcuts");
      },
    },
    {
      id: "act-theme",
      title: t("cmdpalette.cmd_theme_title") || "Toggle Light / Dark Appearance Mode",
      desc: t("cmdpalette.cmd_theme_desc") || "Toggle between high-contrast dark mode and clean light theme",
      category: catSys,
      codicon: "codicon-color-mode",
      action: () => {
        emit("toggle-theme");
      },
    },
    {
      id: "act-scope",
      title: t("cmdpalette.cmd_scope_title") || "Toggle Scope (Workspace / User Global)",
      desc: t("cmdpalette.cmd_scope_desc") || "Toggle active configuration between workspace and global ~/.unirtm",
      category: catSys,
      shortcut: "Alt+S",
      codicon: "codicon-layers",
      action: () => {
        emit("toggle-scope");
      },
    },
    {
      id: "act-terminal",
      title: t("cmdpalette.cmd_terminal_title") || "Toggle Integrated Terminal (Terminal)",
      desc: t("cmdpalette.cmd_terminal_desc") || "Reveal or toggle bottom drawer integrated shell session",
      category: catTool,
      shortcut: "Ctrl+`",
      codicon: "codicon-terminal",
      action: () => {
        emit("open-drawer", "terminal");
      },
    },
    {
      id: "act-new-terminal",
      title: t("cmdpalette.cmd_new_terminal_title") || "Create New Terminal Session",
      desc: t("cmdpalette.cmd_new_terminal_desc") || "Create a new terminal shell session in bottom drawer",
      category: catTool,
      shortcut: "Ctrl+Shift+`",
      codicon: "codicon-terminal",
      action: () => {
        emit("new-terminal");
      },
    },
    {
      id: "act-output",
      title: t("cmdpalette.cmd_output_title") || "Show Runtime Output & Logs",
      desc: t("cmdpalette.cmd_output_desc") || "Inspect background events, system notifications and build outputs",
      category: catTool,
      shortcut: "Ctrl+Shift+U",
      codicon: "codicon-output",
      action: () => {
        emit("open-drawer", "output");
      },
    },
  ];
});

const filteredCommands = computed(() => {
  const q = query.value.trim().toLowerCase();
  if (!q) return allCommands.value;
  return allCommands.value.filter((cmd) => {
    return (
      cmd.title.toLowerCase().includes(q) ||
      (cmd.desc && cmd.desc.toLowerCase().includes(q)) ||
      cmd.category.toLowerCase().includes(q) ||
      cmd.id.toLowerCase().includes(q)
    );
  });
});

watch(filteredCommands, (newFiltered) => {
  if (selectedIndex.value >= newFiltered.length) {
    selectedIndex.value = Math.max(0, newFiltered.length - 1);
  }
});

function moveSelection(step: number) {
  const total = filteredCommands.value.length;
  if (total === 0) return;
  selectedIndex.value = (selectedIndex.value + step + total) % total;
}

function executeSelected() {
  const item = filteredCommands.value[selectedIndex.value];
  if (item) {
    executeItem(item);
  }
}

function executeItem(item: CommandItem) {
  close();
  item.action();
}
</script>

<style scoped>
/* Backdrop Overlay */
.palette-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
  backdrop-filter: blur(5px);
  z-index: 99999;
  display: flex;
  justify-content: center;
  align-items: flex-start;
  padding-top: 6vh;
}

:root[data-theme="light"] .palette-overlay,
[data-theme="light"] .palette-overlay {
  background: rgba(15, 23, 42, 0.38);
}

/* Modal Container */
.palette-container {
  width: 660px;
  max-width: 92vw;
  background: var(--modal-bg, #0d1424);
  border: 1px solid var(--card-border, rgba(0, 229, 255, 0.22));
  border-radius: 12px;
  box-shadow:
    0 24px 64px rgba(0, 0, 0, 0.65),
    0 0 0 1px rgba(255, 255, 255, 0.08);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  color: var(--text-main, #f8fafc);
  transition:
    background 0.2s ease,
    border-color 0.2s ease,
    box-shadow 0.2s ease;
}

/* Light Mode Container */
:root[data-theme="light"] .palette-container,
[data-theme="light"] .palette-container {
  background: #ffffff;
  border: 1px solid #cbd5e1;
  box-shadow:
    0 20px 48px rgba(15, 23, 42, 0.18),
    0 4px 12px rgba(15, 23, 42, 0.08);
  color: #0f172a;
}

/* Top Search Input Row */
.palette-search-row {
  display: flex;
  align-items: center;
  padding: 0.85rem 1.15rem;
  border-bottom: 1px solid var(--card-border, rgba(255, 255, 255, 0.1));
  gap: 0.85rem;
  background: var(--input-bg, rgba(8, 14, 26, 0.75));
}

:root[data-theme="light"] .palette-search-row,
[data-theme="light"] .palette-search-row {
  background: #f8fafc;
  border-bottom: 1px solid #e2e8f0;
}

.palette-search-icon {
  display: flex;
  align-items: center;
  color: var(--accent-cyan, #00e5ff);
}

:root[data-theme="light"] .palette-search-icon,
[data-theme="light"] .palette-search-icon {
  color: #0284c7;
}

.palette-input {
  flex: 1;
  background: transparent;
  border: none;
  outline: none;
  color: var(--text-main, #f8fafc);
  font-size: 0.98rem;
  font-weight: 500;
  font-family: inherit;
}

:root[data-theme="light"] .palette-input,
[data-theme="light"] .palette-input {
  color: #0f172a;
}

.palette-input::placeholder {
  color: var(--text-muted, #94a3b8);
}

:root[data-theme="light"] .palette-input::placeholder,
[data-theme="light"] .palette-input::placeholder {
  color: #94a3b8;
}

.palette-clear-btn {
  background: transparent;
  border: none;
  font-size: 1.25rem;
  line-height: 1;
  color: var(--text-muted, #94a3b8);
  cursor: pointer;
  padding: 0 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
}

.palette-clear-btn:hover {
  color: var(--text-main, #ffffff);
}

:root[data-theme="light"] .palette-clear-btn:hover,
[data-theme="light"] .palette-clear-btn:hover {
  color: #0f172a;
  background: #e2e8f0;
}

.palette-esc-hint {
  font-size: 0.7rem;
  font-weight: 600;
  padding: 3px 6px;
  border-radius: 4px;
  background: rgba(255, 255, 255, 0.08);
  border: 1px solid rgba(255, 255, 255, 0.16);
  color: var(--text-muted, #94a3b8);
  font-family: var(--font-mono, monospace);
}

:root[data-theme="light"] .palette-esc-hint,
[data-theme="light"] .palette-esc-hint {
  background: #f1f5f9;
  border: 1px solid #cbd5e1;
  color: #475569;
}

/* List Wrapper */
.palette-list-wrapper {
  max-height: 400px;
  overflow-y: auto;
  padding: 0.5rem;
}

.palette-empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3rem 1.5rem;
  gap: 0.8rem;
  color: var(--text-muted, #94a3b8);
}

:root[data-theme="light"] .palette-empty-state,
[data-theme="light"] .palette-empty-state {
  color: #64748b;
}

.empty-text {
  font-size: 0.9rem;
  font-weight: 500;
}

.palette-items-list {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

/* Item Entry */
.palette-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.65rem 0.85rem;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.15s ease;
  user-select: none;
  border-left: 3px solid transparent;
}

/* Dark Mode Selected State */
.palette-item:hover,
.palette-item.selected {
  background: rgba(0, 229, 255, 0.12);
  border-left-color: #00e5ff;
}

/* Light Mode Selected State */
:root[data-theme="light"] .palette-item:hover,
[data-theme="light"] .palette-item.selected {
  background: #f0f9ff;
  border-left-color: #0284c7;
}

.item-left {
  display: flex;
  align-items: center;
  gap: 0.85rem;
  overflow: hidden;
}

.item-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  color: var(--accent-cyan, #00e5ff);
  flex-shrink: 0;
}

:root[data-theme="light"] .item-icon,
[data-theme="light"] .item-icon {
  color: #0284c7;
}

.item-details {
  display: flex;
  flex-direction: column;
  gap: 3px;
  overflow: hidden;
}

.item-title {
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--text-main, #f8fafc);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

:root[data-theme="light"] .item-title,
[data-theme="light"] .item-title {
  color: #0f172a;
}

.palette-item.selected .item-title {
  color: var(--accent-cyan, #00e5ff);
}

:root[data-theme="light"] .palette-item.selected .item-title,
[data-theme="light"] .palette-item.selected .item-title {
  color: #0284c7;
}

.item-desc {
  font-size: 0.76rem;
  color: var(--text-muted, #94a3b8);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

:root[data-theme="light"] .item-desc,
[data-theme="light"] .item-desc {
  color: #475569;
}

.palette-item.selected .item-desc {
  color: #cbd5e1;
}

:root[data-theme="light"] .palette-item.selected .item-desc,
[data-theme="light"] .palette-item.selected .item-desc {
  color: #0369a1;
}

.item-right {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  flex-shrink: 0;
}

.category-pill {
  font-size: 0.7rem;
  padding: 2px 7px;
  border-radius: 4px;
  background: rgba(255, 255, 255, 0.07);
  color: var(--text-muted, #94a3b8);
  border: 1px solid rgba(255, 255, 255, 0.08);
}

:root[data-theme="light"] .category-pill,
[data-theme="light"] .category-pill {
  background: #f1f5f9;
  color: #475569;
  border: 1px solid #e2e8f0;
}

.shortcut-pill {
  font-size: 0.7rem;
  padding: 2px 6px;
  border-radius: 4px;
  background: rgba(0, 229, 255, 0.1);
  border: 1px solid rgba(0, 229, 255, 0.25);
  color: #38bdf8;
  font-family: var(--font-mono, monospace);
  font-weight: 500;
}

:root[data-theme="light"] .shortcut-pill,
[data-theme="light"] .shortcut-pill {
  background: #e0f2fe;
  border: 1px solid #bae6fd;
  color: #0284c7;
}

.enter-arrow {
  opacity: 0;
  font-size: 0.92rem;
  color: var(--accent-cyan, #00e5ff);
  transition: opacity 0.15s ease;
}

:root[data-theme="light"] .enter-arrow,
[data-theme="light"] .enter-arrow {
  color: #0284c7;
}

.palette-item.selected .enter-arrow {
  opacity: 1;
}

/* Footer Status Tips */
.palette-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.55rem 1.15rem;
  background: rgba(0, 0, 0, 0.28);
  border-top: 1px solid var(--card-border, rgba(255, 255, 255, 0.08));
  font-size: 0.74rem;
  color: var(--text-muted, #94a3b8);
}

:root[data-theme="light"] .palette-footer,
[data-theme="light"] .palette-footer {
  background: #f8fafc;
  border-top: 1px solid #e2e8f0;
  color: #475569;
}

.footer-tips {
  display: flex;
  gap: 0.85rem;
}

.tip-item kbd {
  font-family: var(--font-mono, monospace);
  background: rgba(255, 255, 255, 0.09);
  padding: 2px 5px;
  border-radius: 3px;
  color: #f1f5f9;
  border: 1px solid rgba(255, 255, 255, 0.15);
}

:root[data-theme="light"] .tip-item kbd,
[data-theme="light"] .tip-item kbd {
  background: #ffffff;
  color: #0f172a;
  border: 1px solid #cbd5e1;
}

.footer-count {
  font-weight: 500;
}

/* Animations */
.palette-fade-enter-active,
.palette-fade-leave-active {
  transition:
    opacity 0.18s cubic-bezier(0.16, 1, 0.3, 1),
    transform 0.18s cubic-bezier(0.16, 1, 0.3, 1);
}

.palette-fade-enter-from,
.palette-fade-leave-to {
  opacity: 0;
  transform: translateY(-10px) scale(0.98);
}
</style>
