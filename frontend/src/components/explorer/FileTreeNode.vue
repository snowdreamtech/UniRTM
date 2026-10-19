<template>
  <div class="tree-node-wrapper" :class="{ 'is-dir': node.is_dir }">
    <!-- Row Item -->
    <div
      class="tree-node-row"
      :class="{
        'is-selected': isSelected,
        'is-ignored': node.is_ignored,
        'is-renaming': isRenaming,
        'is-root-node': node.isRoot,
      }"
      :style="{ paddingLeft: `${depth * 14 + 6}px` }"
      @click="handleClick"
      @dblclick="handleDblClick"
      @contextmenu="handleContextMenu"
      :title="(node.path || node.name) + (node.is_ignored ? ' (已由 .gitignore 排除)' : '')"
    >

      <!-- Indent guide lines for deep nesting -->
      <div
        v-for="i in depth"
        :key="i"
        class="tree-indent-guide"
        :style="{ left: `${(i - 1) * 14 + 12}px` }"
      ></div>

      <!-- Twistie / Chevron Indicator -->
      <span class="tree-twistie" @click.stop="handleTwistieClick">
        <span
          v-if="node.is_dir"
          class="codicon"
          :class="isExpanded ? 'codicon-chevron-down' : 'codicon-chevron-right'"
        ></span>
        <span v-else class="twistie-spacer"></span>
      </span>

      <!-- Folder or File Icon -->
      <span class="tree-icon">
        <template v-if="node.is_dir">
          <TechIcon
            :id="isExpanded ? 'codicon:folder-opened' : 'codicon:folder'"
            :size="16"
          />
        </template>
        <template v-else>
          <TechIcon :id="node.name" :size="16" />
        </template>
      </span>

      <!-- Label or Inline Rename Input -->
      <div class="tree-label-container">
        <input
          v-if="isRenaming"
          ref="renameInputRef"
          v-model="renameValue"
          type="text"
          class="tree-inline-input"
          @keydown.enter.stop="commitRename"
          @keydown.esc.stop="cancelInlineAction"
          @blur="commitRename"
          @click.stop
        />
        <span
          v-else
          class="tree-label"
          :class="[
            gitStatus ? 'git-color-' + gitStatus : '',
            node.is_ignored ? 'is-ignored-text' : '',
            dirGitCount > 0 ? 'git-has-changes' : ''
          ]"
        >
          {{ node.name }}
        </span>
        <span v-if="gitStatus" class="git-tree-badge" :class="'badge-' + gitStatus">{{ gitStatus }}</span>
        <span v-else-if="dirGitCount > 0 && !isExpanded" class="git-dir-badge" :title="`包含 ${dirGitCount} 个待提交的变动文件`">+{{ dirGitCount }}</span>
      </div>


      <!-- Actions on Hover (Rename, Delete) -->
      <div class="tree-row-actions">
        <button
          v-if="!node.is_dir"
          class="tree-action-btn"
          title="重命名 (F2)"
          @click.stop="startInlineRename(node)"
        >
          <span class="codicon codicon-edit"></span>
        </button>
        <button
          class="tree-action-btn danger"
          title="删除"
          @click.stop="promptDelete(node)"
        >
          <span class="codicon codicon-trash"></span>
        </button>
      </div>
    </div>

    <!-- Inline Child Creation Input (if creating inside this folder) -->
    <div
      v-if="isInlineCreatingHere"
      class="tree-node-row is-creating"
      :style="{ paddingLeft: `${(depth + 1) * 14 + 6}px` }"
    >
      <div
        v-for="i in (depth + 1)"
        :key="i"
        class="tree-indent-guide"
        :style="{ left: `${(i - 1) * 14 + 12}px` }"
      ></div>
      <span class="tree-twistie">
        <span v-if="inlineAction?.type === 'new-folder'" class="codicon codicon-chevron-right"></span>
        <span v-else class="twistie-spacer"></span>
      </span>
      <span class="tree-icon">
        <TechIcon
          :id="inlineAction?.type === 'new-folder' ? 'codicon:folder' : (createInputValue || 'newfile.txt')"
          :size="16"
        />
      </span>
      <div class="tree-label-container">
        <input
          ref="createInputRef"
          v-model="createInputValue"
          type="text"
          class="tree-inline-input"
          :placeholder="inlineAction?.type === 'new-folder' ? '文件夹名称' : '文件名'"
          @keydown.enter.stop="commitCreate"
          @keydown.esc.stop="cancelInlineAction"
          @blur="commitCreate"
          @click.stop
        />
      </div>
    </div>

    <!-- Recursive Children -->
    <div v-if="node.is_dir && isExpanded && node.children && node.children.length > 0" class="tree-children">
      <FileTreeNode
        v-for="child in sortedChildren"
        :key="child.path || child.name"
        :node="child"
        :depth="depth + 1"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, watch } from "vue";
import type { FSNode } from "../../composables/useFileTree";
import { useFileTree } from "../../composables/useFileTree";
import { useGit } from "../../composables/useGit";
import TechIcon from "../common/TechIcon.vue";

const props = defineProps<{
  node: FSNode;
  depth: number;
}>();

const {
  selectedNode,
  expandedPaths,
  inlineAction,
  selectNode,
  toggleExpand,
  renameItem,
  createFile,
  createFolder,
  startInlineRename,
  cancelInlineAction,
  openContextMenu,
  promptDelete,
} = useFileTree();

const renameInputRef = ref<HTMLInputElement | null>(null);
const createInputRef = ref<HTMLInputElement | null>(null);
const renameValue = ref("");
const createInputValue = ref("");

const isExpanded = computed(() => {
  return expandedPaths.value.has(props.node.path);
});

const { getGitStatus, getDirGitCount } = useGit();

const gitStatus = computed(() => {
  if (props.node.is_dir) return undefined;
  return getGitStatus(props.node.path);
});

const dirGitCount = computed(() => {
  if (!props.node.is_dir) return 0;
  return getDirGitCount(props.node.path);
});

const isSelected = computed(() => {
  return selectedNode.value?.path === props.node.path;
});

const isRenaming = computed(() => {
  return (
    inlineAction.value?.type === "rename" &&
    inlineAction.value.targetPath === props.node.path
  );
});

const isInlineCreatingHere = computed(() => {
  return (
    (inlineAction.value?.type === "new-file" ||
      inlineAction.value?.type === "new-folder") &&
    inlineAction.value.parentPath === props.node.path &&
    isExpanded.value
  );
});

// Sort children: directories first, then alphabetical
const sortedChildren = computed(() => {
  if (!props.node.children) return [];
  return [...props.node.children].sort((a, b) => {
    if (a.is_dir && !b.is_dir) return -1;
    if (!a.is_dir && b.is_dir) return 1;
    return a.name.localeCompare(b.name, undefined, { numeric: true, sensitivity: "base" });
  });
});

watch(
  isRenaming,
  (val) => {
    if (val) {
      renameValue.value = props.node.name;
      nextTick(() => {
        if (renameInputRef.value) {
          renameInputRef.value.focus();
          // Select file name without extension
          const lastDot = props.node.name.lastIndexOf(".");
          if (!props.node.is_dir && lastDot > 0) {
            renameInputRef.value.setSelectionRange(0, lastDot);
          } else {
            renameInputRef.value.select();
          }
        }
      });
    }
  },
  { immediate: true }
);

watch(
  isInlineCreatingHere,
  (val) => {
    if (val) {
      createInputValue.value = "";
      nextTick(() => {
        createInputRef.value?.focus();
      });
    }
  },
  { immediate: true }
);

function handleClick() {
  selectNode(props.node);
}

function handleDblClick() {
  if (props.node.is_dir) {
    toggleExpand(props.node.path);
  }
}

function handleTwistieClick() {
  if (props.node.is_dir) {
    toggleExpand(props.node.path);
  }
}

function handleContextMenu(e: MouseEvent) {
  openContextMenu(e, props.node);
}

async function commitRename() {
  if (!isRenaming.value) return;
  const val = renameValue.value.trim();
  if (!val || val === props.node.name) {
    cancelInlineAction();
    return;
  }
  await renameItem(props.node.path, val);
}

async function commitCreate() {
  if (!isInlineCreatingHere.value) return;
  const val = createInputValue.value.trim();
  if (!val) {
    cancelInlineAction();
    return;
  }
  const isDir = inlineAction.value?.type === "new-folder";
  if (isDir) {
    await createFolder(props.node.path, val);
  } else {
    await createFile(props.node.path, val);
  }
}
</script>

<style scoped>
.tree-node-wrapper {
  position: relative;
  user-select: none;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
  font-size: 13px;
  line-height: 22px;
}

.tree-node-row {
  position: relative;
  display: flex;
  align-items: center;
  height: 24px;
  cursor: pointer;
  border-radius: 4px;
  margin: 1px 4px;
  padding-right: 6px;
  color: var(--text-main);
  transition: background-color 0.1s ease;
}

.tree-node-row:hover {
  background-color: rgba(128, 128, 128, 0.08);
}

.tree-node-row.is-selected {
  background-color: var(--accent-cyan-glow, rgba(0, 229, 255, 0.15));
  color: var(--accent-cyan);
  font-weight: 500;
}

.tree-node-row.is-root-node .tree-label {
  font-weight: 700;
  color: var(--text-main, #f1f5f9);
  letter-spacing: 0.02em;
}

.tree-node-row.is-ignored {
  opacity: 0.6;
}

.tree-node-row.is-ignored .tree-label {
  color: #8c8c8c !important;
  font-style: italic;
}

.tree-node-row.is-ignored .tree-icon {
  filter: grayscale(100%) opacity(0.4) !important;
}

.tree-node-row.is-ignored:hover {
  opacity: 0.9;
}



.tree-indent-guide {
  position: absolute;
  top: 0;
  bottom: 0;
  width: 1px;
  background-color: var(--card-border, rgba(128, 128, 128, 0.15));
  pointer-events: none;
}

.tree-twistie {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  flex-shrink: 0;
  color: var(--text-muted);
  font-size: 12px;
}

.twistie-spacer {
  width: 18px;
  height: 18px;
}

.tree-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  margin-right: 6px;
  flex-shrink: 0;
}

.tree-label-container {
  flex: 1;
  min-width: 0;
  display: flex;
  align-items: center;
}

.tree-label {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tree-label.git-color-M {
  color: #e2c08d !important;
}

.tree-label.git-color-A,
.tree-label.git-color-U,
.tree-label.git-color-\? {
  color: #73c991 !important;
}

.tree-label.git-color-D {
  color: #c74e39 !important;
}

.tree-label.git-color-R {
  color: #4ec9b0 !important;
}

.git-tree-badge {
  font-size: 11px;
  font-weight: 700;
  margin-left: auto;
  padding: 0 4px;
}

.git-tree-badge.badge-M {
  color: #e2c08d;
}

.git-tree-badge.badge-A,
.git-tree-badge.badge-U,
.git-tree-badge.badge-\? {
  color: #73c991;
}

.git-tree-badge.badge-D {
  color: #c74e39;
}

.git-tree-badge.badge-R {
  color: #4ec9b0;
}

.git-dir-badge {
  font-size: 10px;
  font-weight: 600;
  margin-left: auto;
  padding: 1px 6px;
  border-radius: 10px;
  background: rgba(226, 192, 141, 0.2);
  color: #e2c08d;
  line-height: 14px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  user-select: none;
}

:root[data-theme="light"] .tree-label.git-color-M {
  color: #b45309 !important;
}

:root[data-theme="light"] .tree-label.git-color-A,
:root[data-theme="light"] .tree-label.git-color-U,
:root[data-theme="light"] .tree-label.git-color-\? {
  color: #059669 !important;
}

:root[data-theme="light"] .tree-label.git-color-D {
  color: #dc2626 !important;
}

:root[data-theme="light"] .tree-label.git-color-R {
  color: #0284c7 !important;
}

:root[data-theme="light"] .git-tree-badge.badge-M {
  color: #b45309;
}

:root[data-theme="light"] .git-tree-badge.badge-A,
:root[data-theme="light"] .git-tree-badge.badge-U,
:root[data-theme="light"] .git-tree-badge.badge-\? {
  color: #059669;
}

:root[data-theme="light"] .git-tree-badge.badge-D {
  color: #dc2626;
}

:root[data-theme="light"] .git-tree-badge.badge-R {
  color: #0284c7;
}

:root[data-theme="light"] .git-dir-badge {
  background: rgba(180, 83, 9, 0.12);
  color: #b45309;
}

:root[data-theme="light"] .tree-node-row.is-ignored .tree-label {
  color: #94a3b8 !important;
}

:root[data-theme="light"] .tree-node-row.is-ignored .tree-icon {
  filter: grayscale(100%) opacity(0.35) !important;
}

.tree-inline-input {
  width: 100%;
  background-color: var(--input-bg);
  color: var(--text-main);
  border: 1px solid var(--accent-cyan, #00e5ff);
  border-radius: 3px;
  padding: 1px 6px;
  font-size: 12px;
  line-height: 18px;
  outline: none;
  box-shadow: 0 0 0 1px var(--accent-cyan, #00e5ff);
}

.tree-row-actions {
  display: none;
  align-items: center;
  gap: 2px;
  margin-left: auto;
}

.tree-node-row:hover .tree-row-actions {
  display: flex;
}

.tree-action-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 2px 4px;
  border-radius: 3px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
}

.tree-action-btn:hover {
  background-color: rgba(128, 128, 128, 0.15);
  color: var(--text-main);
}

.tree-action-btn.danger:hover {
  color: var(--error, #ef4444);
}
</style>
