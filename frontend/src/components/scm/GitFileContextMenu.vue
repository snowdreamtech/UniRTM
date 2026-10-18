<!-- Copyright (c) 2026 SnowdreamTech. All rights reserved. -->
<!-- Licensed under the MIT License. See LICENSE file in the project root for full license information. -->

<template>
  <Teleport to="body">
    <div
      v-if="visible && file"
      class="git-file-menu-overlay"
      @click="handleOverlayClick"
      @contextmenu.prevent="handleOverlayClick"
    >
      <div
        class="git-file-menu"
        :style="menuStyle"
        @click.stop
      >
        <!-- 1. Open Changes -->
        <div class="menu-item" @click="triggerAction('open-changes')">
          <span class="codicon codicon-diff-single item-icon"></span>
          <span class="item-label">{{ t('scm.file_menu.open_changes') }}</span>
        </div>

        <!-- 2. Open File -->
        <div class="menu-item" @click="triggerAction('open-file')">
          <span class="codicon codicon-go-to-file item-icon"></span>
          <span class="item-label">{{ t('scm.file_menu.open_file') }}</span>
        </div>

        <div class="menu-separator"></div>

        <!-- 3. Stage / Unstage & Discard -->
        <template v-if="file.isStaged">
          <div class="menu-item" @click="triggerAction('unstage')">
            <span class="codicon codicon-remove item-icon"></span>
            <span class="item-label">{{ t('scm.file_menu.unstage') }}</span>
          </div>
        </template>
        <template v-else>
          <div class="menu-item" @click="triggerAction('stage')">
            <span class="codicon codicon-add item-icon"></span>
            <span class="item-label">{{ t('scm.file_menu.stage') }}</span>
          </div>
          <div class="menu-item danger" @click="triggerAction('discard')">
            <span class="codicon codicon-discard item-icon"></span>
            <span class="item-label">{{ t('scm.file_menu.discard') }}</span>
          </div>
        </template>

        <div class="menu-separator"></div>

        <!-- 4. Reveal in Explorer View -->
        <div class="menu-item" @click="triggerAction('reveal-explorer')">
          <span class="codicon codicon-folder-opened item-icon"></span>
          <span class="item-label">{{ t('scm.file_menu.reveal_explorer') }}</span>
        </div>

        <!-- 5. Copy Path & Relative Path -->
        <div class="menu-item" @click="triggerAction('copy-path')">
          <span class="codicon codicon-copy item-icon"></span>
          <span class="item-label">{{ t('scm.file_menu.copy_path') }}</span>
        </div>
        <div class="menu-item" @click="triggerAction('copy-relative-path')">
          <span class="codicon codicon-files item-icon"></span>
          <span class="item-label">{{ t('scm.file_menu.copy_relative_path') }}</span>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, onMounted, onBeforeUnmount } from "vue";
import { t } from "../../i18n";

export interface ContextMenuTargetFile {
  path: string;
  status?: string;
  isStaged: boolean;
}

export type FileMenuActionType =
  | "open-changes"
  | "open-file"
  | "stage"
  | "unstage"
  | "discard"
  | "reveal-explorer"
  | "copy-path"
  | "copy-relative-path";

const props = defineProps<{
  visible: boolean;
  position: { x: number; y: number };
  file: ContextMenuTargetFile | null;
}>();

const emit = defineEmits<{
  (e: "close"): void;
  (e: "action", action: FileMenuActionType, file: ContextMenuTargetFile): void;
}>();

const MENU_WIDTH = 230;
const ESTIMATED_HEIGHT = 260;

const menuStyle = computed(() => {
  let left = props.position.x;
  let top = props.position.y;

  // Prevent right edge overflow
  if (left + MENU_WIDTH > window.innerWidth - 10) {
    left = Math.max(10, window.innerWidth - MENU_WIDTH - 10);
  }
  // Prevent bottom edge overflow
  if (top + ESTIMATED_HEIGHT > window.innerHeight - 10) {
    top = Math.max(10, window.innerHeight - ESTIMATED_HEIGHT - 10);
  }

  return {
    left: `${Math.max(10, left)}px`,
    top: `${Math.max(10, top)}px`,
  };
});

function triggerAction(action: FileMenuActionType) {
  if (props.file) {
    emit("action", action, props.file);
  }
  emit("close");
}

function handleOverlayClick() {
  emit("close");
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === "Escape" && props.visible) {
    emit("close");
  }
}

onMounted(() => {
  window.addEventListener("keydown", handleKeydown);
  window.addEventListener("resize", handleOverlayClick);
});

onBeforeUnmount(() => {
  window.removeEventListener("keydown", handleKeydown);
  window.removeEventListener("resize", handleOverlayClick);
});
</script>

<style scoped>
.git-file-menu-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 9999;
  background: transparent;
}

.git-file-menu {
  position: fixed;
  background: var(--card-bg, #0d1626);
  border: 1px solid var(--card-border, #1f2e48);
  border-radius: 6px;
  box-shadow: 0 10px 28px rgba(0, 0, 0, 0.45);
  padding: 5px 0;
  font-size: 12px;
  color: var(--text-main, #e2e8f0);
  user-select: none;
  z-index: 10000;
  min-width: 210px;
}

:root[data-theme="light"] .git-file-menu {
  background: #ffffff;
  border-color: #e2e8f0;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.12);
  color: #1e293b;
}

.menu-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 14px;
  cursor: pointer;
  position: relative;
  transition: background 0.12s ease;
}

.menu-item:hover {
  background: var(--hover-bg, rgba(56, 189, 248, 0.12));
  color: var(--accent-cyan, #00e5ff);
}

:root[data-theme="light"] .menu-item:hover {
  background: #f1f5f9;
  color: #0284c7;
}

.item-icon {
  font-size: 14px;
  width: 16px;
  text-align: center;
  color: var(--text-muted, #94a3b8);
}

.menu-item:hover .item-icon {
  color: inherit;
}

.item-label {
  flex: 1;
  white-space: nowrap;
}

.menu-separator {
  height: 1px;
  background: var(--card-border, #1e293b);
  margin: 4px 0;
}

:root[data-theme="light"] .menu-separator {
  background: #e2e8f0;
}

.menu-item.danger {
  color: #f87171;
}

.menu-item.danger:hover {
  background: rgba(239, 68, 68, 0.15);
  color: #ef4444;
}
</style>
