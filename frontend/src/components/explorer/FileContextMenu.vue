<template>
  <div
    v-if="contextMenu.visible"
    ref="menuRef"
    class="file-context-menu"
    :style="{ top: `${contextMenu.y}px`, left: `${contextMenu.x}px` }"
    @click.stop
  >
    <!-- New File / New Folder (if directory, create inside; if file, create in parent directory) -->
    <div class="menu-item" @click="handleNewFile">
      <span class="menu-icon codicon codicon-new-file"></span>
      <span class="menu-label">新建文件</span>
    </div>
    <div class="menu-item" @click="handleNewFolder">
      <span class="menu-icon codicon codicon-new-folder"></span>
      <span class="menu-label">新建文件夹</span>
    </div>

    <div class="menu-separator"></div>

    <!-- Copy Relative Path -->
    <div class="menu-item" @click="handleCopyPath">
      <span class="menu-icon codicon codicon-copy"></span>
      <span class="menu-label">复制相对路径</span>
    </div>

    <!-- Copy Name -->
    <div class="menu-item" @click="handleCopyName">
      <span class="menu-icon codicon codicon-files"></span>
      <span class="menu-label">复制名称</span>
    </div>

    <div class="menu-separator"></div>

    <!-- Rename -->
    <div class="menu-item" @click="handleRename">
      <span class="menu-icon codicon codicon-edit"></span>
      <span class="menu-label">重命名</span>
      <span class="menu-shortcut">F2</span>
    </div>

    <!-- Delete -->
    <div class="menu-item danger" @click="handleDelete">
      <span class="menu-icon codicon codicon-trash"></span>
      <span class="menu-label">删除</span>
      <span class="menu-shortcut">Del</span>
    </div>

    <div class="menu-separator"></div>

    <!-- Open in Integrated Terminal -->
    <div class="menu-item" @click="handleOpenTerminal">
      <span class="menu-icon codicon codicon-terminal"></span>
      <span class="menu-label">在集成终端中打开</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from "vue";
import { useFileTree } from "../../composables/useFileTree";

const emit = defineEmits<{
  (e: "open-terminal", dirPath: string): void;
}>();

const {
  contextMenu,
  closeContextMenu,
  startInlineCreate,
  startInlineRename,
  promptDelete,
  copyRelativePath,
  copyFileName,
} = useFileTree();

const menuRef = ref<HTMLElement | null>(null);

function getTargetParent(): string {
  const node = contextMenu.value.node;
  if (!node) return "";
  if (node.is_dir) return node.path;
  const parts = node.path.split("/").filter(Boolean);
  parts.pop();
  return parts.join("/");
}

function handleNewFile() {
  const parent = getTargetParent();
  closeContextMenu();
  startInlineCreate(parent, "new-file");
}

function handleNewFolder() {
  const parent = getTargetParent();
  closeContextMenu();
  startInlineCreate(parent, "new-folder");
}

function handleCopyPath() {
  const node = contextMenu.value.node;
  if (node) {
    copyRelativePath(node.path);
  }
  closeContextMenu();
}

function handleCopyName() {
  const node = contextMenu.value.node;
  if (node) {
    copyFileName(node.name);
  }
  closeContextMenu();
}

function handleRename() {
  const node = contextMenu.value.node;
  closeContextMenu();
  if (node) {
    startInlineRename(node);
  }
}

function handleDelete() {
  const node = contextMenu.value.node;
  closeContextMenu();
  if (node) {
    promptDelete(node);
  }
}

function handleOpenTerminal() {
  const node = contextMenu.value.node;
  closeContextMenu();
  if (node) {
    const dir = node.is_dir ? node.path : getTargetParent();
    emit("open-terminal", dir);
  }
}

function handleClickOutside(e: MouseEvent) {
  if (contextMenu.value.visible && menuRef.value && !menuRef.value.contains(e.target as Node)) {
    closeContextMenu();
  }
}

onMounted(() => {
  window.addEventListener("click", handleClickOutside);
  window.addEventListener("contextmenu", handleClickOutside);
});

onUnmounted(() => {
  window.removeEventListener("click", handleClickOutside);
  window.removeEventListener("contextmenu", handleClickOutside);
});
</script>

<style scoped>
.file-context-menu {
  position: fixed;
  z-index: 9999;
  background-color: var(--dropdown-background, #252526);
  border: 1px solid var(--dropdown-border, #454545);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.4);
  border-radius: 5px;
  padding: 4px 0;
  min-width: 190px;
  font-size: 13px;
  color: var(--dropdown-foreground, #cccccc);
  user-select: none;
}

.menu-item {
  display: flex;
  align-items: center;
  padding: 5px 12px;
  cursor: pointer;
  gap: 8px;
  transition: background-color 0.1s;
}

.menu-item:hover {
  background-color: var(--list-active-selection-bg, #094771);
  color: #ffffff;
}

.menu-item.danger:hover {
  background-color: rgba(239, 68, 68, 0.2);
  color: #f87171;
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
}

.menu-shortcut {
  font-size: 11px;
  color: var(--text-muted, #858585);
  margin-left: 12px;
}

.menu-separator {
  height: 1px;
  background-color: var(--dropdown-border, #3a3d41);
  margin: 4px 0;
}
</style>
