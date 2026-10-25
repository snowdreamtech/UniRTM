<template>
  <div
    class="explorer-sidebar-container"
    @click="handleBackgroundClick"
    @contextmenu.prevent="handleBackgroundContextMenu"
  >
    <!-- VS Code Section Header -->
    <div class="explorer-section-header">
      <div class="workspace-title-group" @click.stop="toggleRootExpand">
        <span class="codicon" :class="isRootExpanded ? 'codicon-chevron-down' : 'codicon-chevron-right'"></span>
        <span class="workspace-title">{{ workspaceName }}</span>
      </div>

      <!-- Action Toolbar (New File, New Folder, Refresh, Collapse All) -->
      <div class="explorer-toolbar">
        <button
          class="toolbar-btn"
          :title="t('explorer.toolbar_new_file') || 'New File'"
          @click.stop="handleToolbarNewFile"
        >
          <span class="codicon codicon-new-file"></span>
        </button>
        <button
          class="toolbar-btn"
          :title="t('explorer.toolbar_new_folder') || 'New Folder'"
          @click.stop="handleToolbarNewFolder"
        >
          <span class="codicon codicon-new-folder"></span>
        </button>
        <button
          class="toolbar-btn"
          :class="{ 'is-active': isGitFilterActive }"
          :title="
            isGitFilterActive
              ? t('explorer.git_filter_exit') || 'Show All Files (Exit Git Focus)'
              : t('explorer.git_filter_enter') || 'Show Only Git Changes (Git Focus Mode)'
          "
          @click.stop="toggleGitFilter"
        >
          <span class="codicon codicon-source-control"></span>
        </button>
        <button
          class="toolbar-btn"
          :title="t('explorer.refresh_explorer') || 'Refresh Explorer'"
          :class="{ 'is-loading': isLoading }"
          @click.stop="handleRefresh"
        >
          <span class="codicon codicon-refresh"></span>
        </button>
        <button
          class="toolbar-btn"
          :title="t('explorer.add_folder_to_workspace') || 'Add Folder to Workspace...'"
          @click.stop="openAddFolderModal('', '')"
        >
          <span class="codicon codicon-folder-library"></span>
        </button>
        <button class="toolbar-btn" :title="t('explorer.collapse_all') || 'Collapse All'" @click.stop="collapseAll">
          <span class="codicon codicon-collapse-all"></span>
        </button>
      </div>
    </div>

    <!-- Tree Content Area -->
    <div v-show="isRootExpanded" class="explorer-tree-body" ref="treeBodyRef" @scroll="handleTreeScroll">
      <!-- Loading State -->
      <div v-if="isLoading && !filteredTree" class="tree-status-state">
        <span class="codicon codicon-loading codicon-modifier-spin"></span>
        <span class="status-text">{{ t("explorer.loading_tree") || "Loading directory tree..." }}</span>
      </div>

      <!-- Error State -->
      <div v-else-if="error" class="tree-status-state is-error">
        <span class="codicon codicon-error"></span>
        <span class="status-text">{{ error }}</span>
        <button class="retry-btn" @click.stop="fetchTree">{{ t("explorer.retry") || "Retry" }}</button>
      </div>

      <!-- Tree Nodes -->
      <div v-else-if="filteredTree" class="tree-nodes-list">
        <!-- Root Inline Creation Input -->
        <div v-if="isInlineCreatingRoot" class="root-inline-row" :style="{ paddingLeft: '20px' }">
          <span class="root-twistie">
            <span v-if="inlineAction?.type === 'new-folder'" class="codicon codicon-chevron-right"></span>
          </span>
          <span class="root-icon">
            <TechIcon
              :id="inlineAction?.type === 'new-folder' ? 'codicon:folder' : createRootInputValue || 'file.txt'"
              :size="16"
            />
          </span>
          <div class="root-input-box">
            <input
              ref="rootCreateInputRef"
              v-model="createRootInputValue"
              type="text"
              class="tree-inline-input"
              :placeholder="
                inlineAction?.type === 'new-folder'
                  ? t('tree.folder_name_placeholder') || 'Folder name'
                  : t('tree.file_name_placeholder') || 'File name'
              "
              @keydown.enter.stop="commitRootCreate"
              @keydown.esc.stop="cancelInlineAction"
              @blur="commitRootCreate"
              @click.stop
            />
          </div>
        </div>

        <!-- Render flat tree nodes (Virtual Scrolling for large trees) -->
        <template v-if="flatVisibleNodes.length > 0">
          <div
            v-if="flatVisibleNodes.length > 50"
            class="tree-virtual-container"
            :style="{ height: virtualTreeTotalHeight + 'px', position: 'relative' }"
          >
            <div
              class="tree-virtual-slider"
              :style="{
                transform: `translateY(${virtualTreeRows[0]?.top || 0}px)`,
                position: 'absolute',
                top: 0,
                left: 0,
                right: 0,
              }"
            >
              <FileTreeNode
                v-for="vItem in virtualTreeRows"
                :key="vItem.item.key"
                :node="vItem.item.node"
                :depth="vItem.item.depth"
                :flat="true"
              />
            </div>
          </div>
          <template v-else>
            <FileTreeNode
              v-for="item in flatVisibleNodes"
              :key="item.key"
              :node="item.node"
              :depth="item.depth"
              :flat="true"
            />
          </template>
        </template>

        <div v-else-if="!isInlineCreatingRoot" class="tree-empty-state">
          <span v-if="isGitFilterActive">{{ t("explorer.no_git_changes") || "No Git changes (Workspace clean)" }}</span>
          <span v-else>{{ t("explorer.no_files") || "No files" }}</span>
        </div>
      </div>
    </div>

    <!-- Global Context Menu, Delete Modal & Add Folder Modal -->
    <FileContextMenu @open-terminal="handleOpenTerminal" />
    <DeleteConfirmModal />
    <AddFolderModal />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, nextTick, watch } from "vue";
import { useFileTree, flattenVisibleTree, type FlatTreeNode } from "../../composables/useFileTree";
import { useVirtualList } from "../../composables/useVirtualList";
import FileTreeNode from "./FileTreeNode.vue";
import FileContextMenu from "./FileContextMenu.vue";
import DeleteConfirmModal from "./DeleteConfirmModal.vue";
import AddFolderModal from "./AddFolderModal.vue";
import TechIcon from "../common/TechIcon.vue";
import { t } from "../../i18n";

import { useWorkspaceBus } from "../../composables/useWorkspaceBus";

const emit = defineEmits<{
  (e: "open-terminal", dirPath: string): void;
}>();

const {
  tree,
  filteredTree,
  isGitFilterActive,
  toggleGitFilter,
  isLoading,
  error,
  selectedNode,
  expandedPaths,
  inlineAction,
  fetchTree,
  toggleExpand,
  revealPath,
  collapseAll,

  isMultiRoot,
  workspaceRoots,
  openAddFolderModal,
  startInlineCreate,
  cancelInlineAction,
  createFile,
  createFolder,
  openContextMenu,
} = useFileTree();

const { onRevealFile } = useWorkspaceBus();

const rootCreateInputRef = ref<HTMLInputElement | null>(null);
const createRootInputValue = ref("");

const workspaceName = computed(() => {
  if (isMultiRoot.value) {
    return (
      t("explorer.workspace_multi_title", { count: workspaceRoots.value.length }) ||
      t("explorer.workspace_folders_count", { count: workspaceRoots.value.length }) ||
      `Workspace (${workspaceRoots.value.length} folders)`
    );
  }
  return (tree.value?.name || "UNIRTM").toUpperCase();
});

const isRootExpanded = computed(() => {
  return expandedPaths.value.has("");
});

function toggleRootExpand() {
  toggleExpand("");
}

const isInlineCreatingRoot = computed(() => {
  return (
    (inlineAction.value?.type === "new-file" || inlineAction.value?.type === "new-folder") &&
    inlineAction.value.parentPath === ""
  );
});

watch(
  isInlineCreatingRoot,
  (val) => {
    if (val) {
      createRootInputValue.value = "";
      nextTick(() => {
        rootCreateInputRef.value?.focus();
      });
    }
  },
  { immediate: true }
);

const flatVisibleNodes = computed(() => {
  const rootChildren = filteredTree.value?.children || [];
  return flattenVisibleTree(rootChildren, expandedPaths.value, 0);
});

const treeBodyRef = ref<HTMLElement | null>(null);

const {
  visibleItems: virtualTreeRows,
  totalHeight: virtualTreeTotalHeight,
  handleScroll: handleTreeScroll,
  scrollToIndex: scrollToTreeIndex,
} = useVirtualList<FlatTreeNode>({
  items: flatVisibleNodes,
  itemHeight: 26,
  containerRef: treeBodyRef,
  overscan: 10,
});

function handleToolbarNewFile() {
  const parent = selectedNode.value?.is_dir ? selectedNode.value.path : "";
  startInlineCreate(parent, "new-file");
}

function handleToolbarNewFolder() {
  const parent = selectedNode.value?.is_dir ? selectedNode.value.path : "";
  startInlineCreate(parent, "new-folder");
}

function handleRefresh() {
  fetchTree();
}

async function commitRootCreate() {
  if (!isInlineCreatingRoot.value) return;
  const val = createRootInputValue.value.trim();
  if (!val) {
    cancelInlineAction();
    return;
  }
  const isDir = inlineAction.value?.type === "new-folder";
  if (isDir) {
    await createFolder("", val);
  } else {
    await createFile("", val);
  }
}

function handleBackgroundClick() {
  selectedNode.value = null;
}

function handleBackgroundContextMenu(e: MouseEvent) {
  if (tree.value) {
    openContextMenu(e, {
      name: tree.value.name,
      path: "",
      is_dir: true,
      size: 0,
      mod_time: new Date().toISOString(),
    });
  }
}

function handleOpenTerminal(dirPath: string) {
  emit("open-terminal", dirPath);
}

let revealCleanup: (() => void) | null = null;

onMounted(() => {
  if (!tree.value) {
    fetchTree();
  }

  revealCleanup = onRevealFile((req) => {
    revealPath(req.path);
    nextTick(() => {
      const targetIdx = flatVisibleNodes.value.findIndex(
        (item) => item.node.path === req.path || item.node.path.endsWith("/" + req.path)
      );
      if (targetIdx !== -1) {
        scrollToTreeIndex(targetIdx, "center");
      }
      nextTick(() => {
        const fileName = req.path.split("/").pop() || req.path;
        const el =
          document.querySelector(`.tree-node-row[title*="${CSS.escape(req.path)}"]`) ||
          document.querySelector(`.tree-node-row[title*="${CSS.escape(fileName)}"]`);
        if (el) {
          el.scrollIntoView({ behavior: "smooth", block: "nearest" });
          el.classList.add("reveal-flash");
          setTimeout(() => {
            el.classList.remove("reveal-flash");
          }, 1800);
        }
      });
    });
  });
});

onBeforeUnmount(() => {
  if (revealCleanup) {
    revealCleanup();
  }
});
</script>

<style scoped>
.explorer-sidebar-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  width: 100%;
  overflow: hidden;
  user-select: none;
}

.explorer-section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 28px;
  padding: 0 8px 0 6px;
  background-color: var(--section-bg, transparent);
  color: var(--text-muted);
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.5px;
  border-top: 1px solid var(--card-border);
  border-bottom: 1px solid var(--card-border);
}

.workspace-title-group {
  display: flex;
  align-items: center;
  gap: 4px;
  cursor: pointer;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--text-main);
}

.workspace-title {
  text-transform: uppercase;
  font-weight: 700;
}

.explorer-toolbar {
  display: flex;
  align-items: center;
  gap: 2px;
}

.toolbar-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 2px 4px;
  border-radius: 3px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  line-height: 1;
  transition: all 0.15s ease;
}

.toolbar-btn:hover {
  background-color: rgba(128, 128, 128, 0.15);
  color: var(--text-main);
}

.toolbar-btn.is-active {
  background-color: var(--accent-cyan-glow, rgba(0, 229, 255, 0.15)) !important;
  color: var(--accent-cyan, #00e5ff) !important;
}

:root[data-theme="light"] .toolbar-btn.is-active {
  background-color: rgba(2, 132, 199, 0.15) !important;
  color: #0284c7 !important;
}

.toolbar-btn.is-loading .codicon-refresh {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.explorer-tree-body {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  padding-top: 2px;
}

.tree-status-state {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 24px 12px;
  color: var(--text-muted);
  font-size: 12px;
}

.tree-status-state.is-error {
  flex-direction: column;
  color: var(--error, #ef4444);
}

.retry-btn {
  margin-top: 6px;
  background-color: transparent;
  border: 1px solid var(--card-border);
  color: var(--text-main);
  padding: 3px 10px;
  border-radius: 3px;
  cursor: pointer;
  font-size: 11px;
}

.retry-btn:hover {
  background-color: rgba(128, 128, 128, 0.12);
  color: var(--text-main);
}

.tree-empty-state {
  padding: 20px 16px;
  color: var(--text-muted);
  font-size: 12px;
  text-align: center;
}

.root-inline-row {
  display: flex;
  align-items: center;
  height: 24px;
  margin: 1px 4px;
}

.root-twistie {
  width: 18px;
  height: 18px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.root-icon {
  width: 18px;
  height: 18px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  margin-right: 6px;
}

.root-input-box {
  flex: 1;
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

:deep(.reveal-flash) {
  animation: revealPulse 1.8s ease-out;
}

@keyframes revealPulse {
  0% {
    background-color: rgba(56, 189, 248, 0.45);
    outline: 2px solid #38bdf8;
  }
  50% {
    background-color: rgba(56, 189, 248, 0.25);
    outline: 1px solid #38bdf8;
  }
  100% {
    background-color: transparent;
    outline: none;
  }
}
</style>
