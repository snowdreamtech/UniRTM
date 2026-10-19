<!-- Copyright (c) 2026 SnowdreamTech. All rights reserved. -->
<!-- Licensed under the MIT License. See LICENSE file in the project root for full license information. -->
<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from "vue";
import { t } from "../../i18n";

const props = defineProps<{
  visible: boolean;
  x: number;
  y: number;
  hasSelection: boolean;
}>();

const emit = defineEmits<{
  (e: "close"): void;
  (e: "copy"): void;
  (e: "paste"): void;
  (e: "select-all"): void;
  (e: "clear"): void;
  (e: "clear-history"): void;
  (e: "export-log"): void;
  (e: "find"): void;
  (e: "restart"): void;
}>();

const menuRef = ref<HTMLElement | null>(null);

const adjustedStyle = computed(() => {
  const menuWidth = 190;
  const menuHeight = 220;
  let posX = props.x;
  let posY = props.y;

  if (typeof window !== "undefined") {
    if (posX + menuWidth > window.innerWidth) {
      posX = Math.max(10, window.innerWidth - menuWidth - 10);
    }
    if (posY + menuHeight > window.innerHeight) {
      posY = Math.max(10, window.innerHeight - menuHeight - 10);
    }
  }

  return {
    left: `${posX}px`,
    top: `${posY}px`,
  };
});

function handleItemClick(action: () => void) {
  emit("close");
  action();
}

function handleOutsideClick(e: MouseEvent) {
  if (props.visible && menuRef.value && !menuRef.value.contains(e.target as Node)) {
    emit("close");
  }
}

onMounted(() => {
  window.addEventListener("click", handleOutsideClick);
  window.addEventListener("contextmenu", handleOutsideClick);
});

onBeforeUnmount(() => {
  window.removeEventListener("click", handleOutsideClick);
  window.removeEventListener("contextmenu", handleOutsideClick);
});
</script>

<template>
  <div
    v-if="visible"
    ref="menuRef"
    class="terminal-context-menu"
    :style="adjustedStyle"
    role="menu"
    aria-label="Terminal Context Menu"
    @click.stop
  >
    <!-- Copy -->
    <div
      class="menu-item"
      :class="{ disabled: !hasSelection }"
      role="menuitem"
      tabindex="0"
      @click="hasSelection && handleItemClick(() => emit('copy'))"
    >
      <span class="menu-icon codicon codicon-copy"></span>
      <span class="menu-label">{{ t('terminal.menu.copy') }}</span>
      <span class="menu-shortcut">Ctrl+Shift+C</span>
    </div>

    <!-- Paste -->
    <div
      class="menu-item"
      role="menuitem"
      tabindex="0"
      @click="handleItemClick(() => emit('paste'))"
    >
      <span class="menu-icon codicon codicon-clippy"></span>
      <span class="menu-label">{{ t('terminal.menu.paste') }}</span>
      <span class="menu-shortcut">Ctrl+Shift+V</span>
    </div>

    <div class="menu-separator"></div>

    <!-- Find -->
    <div
      class="menu-item"
      role="menuitem"
      tabindex="0"
      @click="handleItemClick(() => emit('find'))"
    >
      <span class="menu-icon codicon codicon-search"></span>
      <span class="menu-label">{{ t('terminal.menu.find') }}</span>
      <span class="menu-shortcut">Ctrl+F</span>
    </div>

    <!-- Select All -->
    <div
      class="menu-item"
      role="menuitem"
      tabindex="0"
      @click="handleItemClick(() => emit('select-all'))"
    >
      <span class="menu-icon codicon codicon-check-all"></span>
      <span class="menu-label">{{ t('terminal.menu.select_all') }}</span>
      <span class="menu-shortcut">Ctrl+A</span>
    </div>

    <!-- Export Log -->
    <div
      class="menu-item"
      role="menuitem"
      tabindex="0"
      @click="handleItemClick(() => emit('export-log'))"
    >
      <span class="menu-icon codicon codicon-cloud-download"></span>
      <span class="menu-label">{{ t('terminal.menu.export_log') }}</span>
    </div>

    <!-- Clear Buffer -->
    <div
      class="menu-item"
      role="menuitem"
      tabindex="0"
      @click="handleItemClick(() => emit('clear'))"
    >
      <span class="menu-icon codicon codicon-clear-all"></span>
      <span class="menu-label">{{ t('terminal.menu.clear') }}</span>
      <span class="menu-shortcut">Ctrl+K</span>
    </div>

    <!-- Clear Persistent Buffer -->
    <div
      class="menu-item"
      role="menuitem"
      tabindex="0"
      @click="handleItemClick(() => emit('clear-history'))"
    >
      <span class="menu-icon codicon codicon-trash"></span>
      <span class="menu-label">{{ t('terminal.menu.clear_history') }}</span>
    </div>

    <div class="menu-separator"></div>

    <!-- Restart Session -->
    <div
      class="menu-item danger"
      role="menuitem"
      tabindex="0"
      @click="handleItemClick(() => emit('restart'))"
    >
      <span class="menu-icon codicon codicon-refresh"></span>
      <span class="menu-label">{{ t('terminal.menu.restart') }}</span>
    </div>
  </div>
</template>

<style scoped>
.terminal-context-menu {
  position: fixed;
  z-index: 1000;
  min-width: 180px;
  background-color: var(--vscode-menu-background, #1e1e1e);
  color: var(--vscode-menu-foreground, #cccccc);
  border: 1px solid var(--vscode-menu-border, #454545);
  border-radius: 5px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.45);
  padding: 4px 0;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  font-size: 12px;
  user-select: none;
}

.menu-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 5px 12px;
  cursor: pointer;
  transition: background-color 0.1s ease;
}

.menu-item:hover:not(.disabled) {
  background-color: var(--vscode-menu-selectionBackground, #04395e);
  color: var(--vscode-menu-selectionForeground, #ffffff);
}

.menu-item.disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.menu-item.danger:hover {
  background-color: #7f1d1d;
  color: #fca5a5;
}

.menu-icon {
  font-size: 14px;
  width: 14px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.menu-label {
  flex: 1;
}

.menu-shortcut {
  font-size: 11px;
  color: var(--vscode-menu-shortcutForeground, #888888);
  margin-left: 12px;
  font-family: monospace;
}

.menu-separator {
  height: 1px;
  background-color: var(--vscode-menu-separatorBackground, #333333);
  margin: 4px 0;
}
</style>
