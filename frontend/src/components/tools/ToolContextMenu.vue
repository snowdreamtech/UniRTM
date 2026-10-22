<!-- Copyright (c) 2026 SnowdreamTech. All rights reserved. -->
<!-- Licensed under the MIT License. See LICENSE file in the project root for full license information. -->

<template>
  <Teleport to="body">
    <div v-if="visible && tool" class="context-menu-backdrop" @click="handleClose" @contextmenu.prevent="handleClose">
      <div ref="menuRef" class="tool-context-menu" :style="adjustedStyle" @click.stop>
        <!-- 1. 切换 (Switch) -->
        <div class="menu-item" @click="handleAction('switch')">
          <span class="menu-icon codicon codicon-arrow-swap"></span>
          <span class="menu-label">{{ t("runtimes.action_switch") || "切换" }}</span>
        </div>

        <!-- 2. 安装 (Install) -->
        <div class="menu-item" @click="handleAction('install')">
          <span class="menu-icon codicon codicon-cloud-download"></span>
          <span class="menu-label">{{ t("runtimes.action_install") || "安装" }}</span>
        </div>

        <!-- 3. 锁定 / 解锁 (Lock / Unlock) -->
        <div class="menu-item" @click="handleAction('toggle-pin')">
          <span class="menu-icon codicon" :class="isPinned ? 'codicon-unlock' : 'codicon-lock'"></span>
          <span class="menu-label">
            {{ isPinned ? t("runtimes.action_unlock") || "解锁" : t("runtimes.action_lock") || "锁定" }}
          </span>
        </div>

        <div class="menu-separator"></div>

        <!-- 4. 删除 (Delete) -->
        <div class="menu-item danger" @click="handleAction('delete')">
          <span class="menu-icon codicon codicon-trash"></span>
          <span class="menu-label">{{ t("runtimes.action_delete") || "删除" }}</span>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, ref, onMounted, onBeforeUnmount } from "vue";
import { t } from "../../i18n";

export type ToolMenuActionType = "switch" | "install" | "toggle-pin" | "delete";

const props = defineProps<{
  visible: boolean;
  position: { x: number; y: number };
  tool: any | null;
}>();

const emit = defineEmits<{
  (e: "close"): void;
  (e: "action", action: ToolMenuActionType, tool: any): void;
}>();

const menuRef = ref<HTMLElement | null>(null);

const isPinned = computed(() => {
  if (!props.tool) return false;
  return Boolean(props.tool.pinned || props.tool.raw?.pinned);
});

const MENU_WIDTH = 140;
const MENU_HEIGHT = 145;

const adjustedStyle = computed(() => {
  let { x, y } = props.position;
  const padding = 8;
  const viewportWidth = window.innerWidth;
  const viewportHeight = window.innerHeight;

  if (x + MENU_WIDTH > viewportWidth - padding) {
    x = Math.max(padding, viewportWidth - MENU_WIDTH - padding);
  }

  if (y + MENU_HEIGHT > viewportHeight - padding) {
    y = Math.max(padding, viewportHeight - MENU_HEIGHT - padding);
  }

  return {
    left: `${x}px`,
    top: `${y}px`,
  };
});

function handleClose() {
  emit("close");
}

function handleAction(action: ToolMenuActionType) {
  if (!props.tool) return;
  emit("action", action, props.tool);
  emit("close");
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === "Escape" && props.visible) {
    handleClose();
  }
}

onMounted(() => {
  window.addEventListener("keydown", handleKeydown);
  window.addEventListener("resize", handleClose);
});

onBeforeUnmount(() => {
  window.removeEventListener("keydown", handleKeydown);
  window.removeEventListener("resize", handleClose);
});
</script>

<style scoped>
.context-menu-backdrop {
  position: fixed;
  inset: 0;
  z-index: var(--z-context-menu-backdrop, 1000);
  background: transparent;
}

.tool-context-menu {
  position: fixed;
  z-index: var(--z-context-menu-popover, 1001);
  background-color: var(--modal-bg, var(--card-bg));
  border: 1px solid var(--card-border);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.35);
  border-radius: 6px;
  padding: 4px 0;
  min-width: 130px;
  font-size: 13px;
  color: var(--text-main);
  user-select: none;
}

.menu-item {
  display: flex;
  align-items: center;
  padding: 6px 14px;
  cursor: pointer;
  gap: 8px;
  transition: all 0.1s ease;
}

.menu-item:hover {
  background-color: var(--accent-cyan-glow, rgba(0, 229, 255, 0.15));
  color: var(--accent-cyan);
}

.menu-item.danger,
.menu-item.danger .menu-label,
.menu-item.danger .menu-icon {
  color: var(--error, #f87171);
}

:root[data-theme="light"] .menu-item.danger,
:root[data-theme="light"] .menu-item.danger .menu-label,
:root[data-theme="light"] .menu-item.danger .menu-icon {
  color: #ef4444;
}

.menu-item.danger:hover {
  background-color: rgba(239, 68, 68, 0.15);
  color: #ef4444;
}

.menu-icon {
  font-size: 14px;
  width: 16px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.menu-label {
  flex: 1;
  font-weight: 500;
}

.menu-separator {
  height: 1px;
  background-color: var(--card-border);
  margin: 4px 0;
}
</style>
