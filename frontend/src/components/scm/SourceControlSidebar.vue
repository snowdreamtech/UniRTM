<template>
  <div class="scm-sidebar-container" id="scm-sidebar">
    <!-- 1. SCM Header: Branch info & Sync -->
    <div class="scm-header-bar">
      <div class="branch-selector" :title="`${t('scm.branch.current')}: ${currentBranch}${status.upstream ? ' -> ' + status.upstream : ''}`" @click="isBranchMenuOpen = !isBranchMenuOpen">
        <span class="codicon codicon-git-branch"></span>
        <span class="branch-name">{{ currentBranch }}</span>
        <span v-if="status.ahead > 0 || status.behind > 0" class="branch-sync-pill">
          <span v-if="status.behind > 0">↓{{ status.behind }}</span>
          <span v-if="status.ahead > 0">↑{{ status.ahead }}</span>
        </span>
        <span class="codicon codicon-chevron-down twistie"></span>
      </div>

      <div class="scm-header-actions">
        <button
          class="scm-action-btn"
          :title="t('scm.tabs.graph')"
          @click.stop="openGitGraph"
        >
          <span class="codicon codicon-git-commit"></span>
        </button>
        <button
          class="scm-action-btn"
          :title="t('scm.actions.refresh')"
          :class="{ 'is-loading': isLoading }"
          @click.stop="handleRefresh"
        >
          <span class="codicon codicon-refresh"></span>
        </button>
        <button
          class="scm-action-btn"
          :title="t('scm.actions.more')"
          @click.stop="toggleMoreMenu"
        >
          <span class="codicon codicon-ellipsis"></span>
        </button>
      </div>
    </div>

    <!-- Branch Dropdown Popover -->
    <div v-if="isBranchMenuOpen" class="branch-dropdown-menu">
      <div class="branch-dropdown-header">
        <span>{{ t('scm.branch.switch') }}</span>
        <button class="branch-close-btn" @click="isBranchMenuOpen = false">×</button>
      </div>
      <div class="branch-list">
        <div
          v-for="b in branches"
          :key="b.name"
          class="branch-item"
          :class="{ 'is-current': b.current }"
          @click="handleSelectBranch(b.name)"
        >
          <span class="codicon" :class="b.current ? 'codicon-check' : 'codicon-git-branch'"></span>
          <span class="branch-item-name">{{ b.name }}</span>
          <span v-if="b.remote" class="branch-remote-tag">{{ t('scm.branch.remote') }}</span>
        </div>
      </div>
    </div>

    <!-- 2. Commit Message Box -->
    <div class="scm-commit-section">
      <div class="commit-input-wrapper">
        <textarea
          v-model="commitMessage"
          class="commit-textarea"
          :placeholder="t('scm.sidebar.commit_placeholder')"
          rows="3"
          @keydown.ctrl.enter.prevent="handleCommit"
          @keydown.meta.enter.prevent="handleCommit"
        ></textarea>
      </div>

      <div class="commit-options-bar">
        <label class="gpg-sign-toggle" :title="t('scm.sidebar.gpg_sign')">
          <input type="checkbox" v-model="signCommit" class="gpg-sign-checkbox" />
          <span class="codicon codicon-verified gpg-sign-icon"></span>
          <span class="gpg-sign-text">{{ t('scm.sidebar.gpg_sign') }}</span>
        </label>
      </div>

      <button
        class="commit-btn"
        :disabled="isCommitting || !commitMessage.trim()"
        @click="handleCommit"
      >
        <span v-if="isCommitting" class="codicon codicon-loading codicon-modifier-spin"></span>
        <span v-else class="codicon codicon-check"></span>
        <span>{{ t('scm.sidebar.commit') }}</span>
      </button>
    </div>

    <!-- 3. Staged Changes Section -->
    <div class="scm-changes-group">
      <div class="group-header" @click="isStagedExpanded = !isStagedExpanded">
        <div class="group-left">
          <span class="codicon twistie" :class="isStagedExpanded ? 'codicon-chevron-down' : 'codicon-chevron-right'"></span>
          <span class="group-title">{{ t('scm.sidebar.staged') }}</span>
          <span class="group-count">{{ status.staged.length }}</span>
        </div>
        <div class="group-actions" @click.stop>
          <button
            v-if="status.staged.length > 0"
            class="group-btn"
            :title="t('scm.actions.unstage_all')"
            @click="handleUnstageAll"
          >
            <span class="codicon codicon-remove"></span>
          </button>
        </div>
      </div>

      <div v-show="isStagedExpanded" class="group-items-list">
        <div
          v-for="item in status.staged"
          :key="'staged:' + item.path"
          class="change-item"
          :class="{ active: selectedDiffPath === item.path && selectedIsStaged === true }"
          @click="handleSelectFileDiff(item.path, true)"
          @contextmenu.prevent="openFileContextMenu($event, item, true)"
          :title="item.path"
        >
          <span class="change-icon">
            <TechIcon :id="getFileName(item.path)" :size="16" />
          </span>
          <div class="change-info">
            <span class="change-filename">{{ getFileName(item.path) }}</span>
            <span class="change-dirname">{{ getDirName(item.path) }}</span>
          </div>
          <span class="change-status-badge" :class="'badge-' + item.status">{{ item.status }}</span>
          <div class="change-item-actions" @click.stop>
            <button
              class="item-action-btn"
              :title="t('scm.diff.open_file')"
              @click="handleOpenFileInEditor(item.path)"
            >
              <span class="codicon codicon-go-to-file"></span>
            </button>
            <button
              class="item-action-btn"
              :title="t('scm.diff.unstage_file')"
              @click="handleUnstageFile(item.path)"
            >
              <span class="codicon codicon-remove"></span>
            </button>
          </div>
        </div>

        <div v-if="status.staged.length === 0" class="empty-hint">
          <span>{{ t('scm.sidebar.no_staged') }}</span>
        </div>
      </div>
    </div>

    <!-- 4. Changes (Unstaged) Section -->
    <div class="scm-changes-group">
      <div class="group-header" @click="isUnstagedExpanded = !isUnstagedExpanded">
        <div class="group-left">
          <span class="codicon twistie" :class="isUnstagedExpanded ? 'codicon-chevron-down' : 'codicon-chevron-right'"></span>
          <span class="group-title">{{ t('scm.sidebar.unstaged') }}</span>
          <span class="group-count">{{ status.unstaged.length }}</span>
        </div>
        <div class="group-actions" @click.stop>
          <button
            v-if="status.unstaged.length > 0"
            class="group-btn"
            :title="t('scm.actions.discard_all')"
            @click="handleDiscardAll"
          >
            <span class="codicon codicon-discard"></span>
          </button>
          <button
            v-if="status.unstaged.length > 0"
            class="group-btn"
            :title="t('scm.actions.stage_all')"
            @click="handleStageAll"
          >
            <span class="codicon codicon-add"></span>
          </button>
        </div>
      </div>

      <div v-show="isUnstagedExpanded" class="group-items-list">
        <div
          v-for="item in status.unstaged"
          :key="'unstaged:' + item.path"
          class="change-item"
          :class="{ active: selectedDiffPath === item.path && selectedIsStaged === false }"
          @click="handleSelectFileDiff(item.path, false)"
          @contextmenu.prevent="openFileContextMenu($event, item, false)"
          :title="item.path"
        >
          <span class="change-icon">
            <TechIcon :id="getFileName(item.path)" :size="16" />
          </span>
          <div class="change-info">
            <span class="change-filename">{{ getFileName(item.path) }}</span>
            <span class="change-dirname">{{ getDirName(item.path) }}</span>
          </div>
          <span class="change-status-badge" :class="'badge-' + item.status">{{ item.status }}</span>
          <div class="change-item-actions" @click.stop>
            <button
              class="item-action-btn"
              :title="t('scm.diff.open_file')"
              @click="handleOpenFileInEditor(item.path)"
            >
              <span class="codicon codicon-go-to-file"></span>
            </button>
            <button
              class="item-action-btn"
              :title="t('scm.diff.discard_file')"
              @click="handleDiscardFile(item.path)"
            >
              <span class="codicon codicon-discard"></span>
            </button>
            <button
              class="item-action-btn"
              :title="t('scm.diff.stage_file')"
              @click="handleStageFile(item.path)"
            >
              <span class="codicon codicon-add"></span>
            </button>
          </div>
        </div>

        <div v-if="status.unstaged.length === 0" class="empty-hint">
          <span>{{ t('scm.sidebar.clean_working_tree') }}</span>
        </div>
      </div>
    </div>

    <!-- Repository More Actions Menu (Teleported) -->
    <GitMoreActionsMenu
      :visible="isMoreMenuOpen"
      :position="moreMenuPos"
      @close="isMoreMenuOpen = false"
      @action="handleMoreAction"
    />

    <!-- File Context Menu (Teleported) -->
    <GitFileContextMenu
      :visible="isFileContextMenuOpen"
      :position="fileContextMenuPos"
      :file="fileContextMenuTarget"
      @close="isFileContextMenuOpen = false"
      @action="handleFileMenuAction"
    />

    <!-- Git Action Modal Dialog -->
    <GitActionModal
      :visible="modalVisible"
      :action-type="activeModalType"
      :target-commit="null"
      :target-ref-name="modalTargetRefName"
      :current-branch-name="currentBranch"
      @cancel="modalVisible = false"
      @confirm="handleModalConfirm"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useRouter, useRoute } from "vue-router";
import { useGit } from "../../composables/useGit";
import { useEditorTabs } from "../../composables/useEditorTabs";
import { useWorkspaceBus } from "../../composables/useWorkspaceBus";
import { useToast } from "../../composables/useToast";

import { t } from "../../i18n";
import TechIcon from "../common/TechIcon.vue";
import GitMoreActionsMenu from "./GitMoreActionsMenu.vue";
import GitFileContextMenu, { type ContextMenuTargetFile, type FileMenuActionType } from "./GitFileContextMenu.vue";
import GitActionModal, { type GitModalActionType } from "./GitActionModal.vue";

const router = useRouter();
const route = useRoute();
const { showToast } = useToast();
const {
  status,
  branches,
  currentBranch,
  commitMessage,
  isLoading,
  isCommitting,
  fetchStatus,
  fetchBranches,
  stage,
  unstage,
  discard,
  commit,
  checkout,
  pull,
  fetchRemote,
  sync,
  createBranchFromCommit,
  renameBranch,
  deleteBranch,
  mergeBranch,
  rebaseBranch,
  pushBranch,
  createTag,
  deleteTag,
  pushTag,
  stashUncommitted,
  resetUncommitted,
  cleanUntracked,
} = useGit();

const { openTab } = useEditorTabs();

const isStagedExpanded = ref(true);
const isUnstagedExpanded = ref(true);
const isBranchMenuOpen = ref(false);
const signCommit = ref(true);

const isMoreMenuOpen = ref(false);
const moreMenuPos = ref({ x: 0, y: 0 });

const modalVisible = ref(false);
const activeModalType = ref<GitModalActionType | null>(null);
const modalTargetRefName = ref<string | undefined>(undefined);

const selectedDiffPath = ref<string>("");
const selectedIsStaged = ref<boolean>(false);

const isFileContextMenuOpen = ref(false);
const fileContextMenuPos = ref({ x: 0, y: 0 });
const fileContextMenuTarget = ref<ContextMenuTargetFile | null>(null);

onMounted(async () => {
  await Promise.all([fetchStatus(), fetchBranches()]);
});

function handleRefresh() {
  fetchStatus().catch(() => {});
  fetchBranches().catch(() => {});
}

function openGitGraph() {
  router.push({
    name: "scm",
    query: {
      scope: route.query.scope || "workspace",
      tab: "graph",
    },
  });
}

function getFileName(filePath: string): string {
  const parts = filePath.split("/");
  return parts[parts.length - 1] || filePath;
}

function getDirName(filePath: string): string {
  const parts = filePath.split("/");
  if (parts.length <= 1) return "";
  return parts.slice(0, parts.length - 1).join("/");
}

async function handleSelectBranch(branchName: string) {
  isBranchMenuOpen.value = false;
  await checkout(branchName);
}

async function handleCommit() {
  await commit(undefined, false, signCommit.value);
}

async function handleStageFile(path: string) {
  await stage([path]);
}

async function handleStageAll() {
  await stage([]);
}

async function handleUnstageFile(path: string) {
  await unstage([path]);
}

async function handleUnstageAll() {
  await unstage([]);
}

async function handleDiscardFile(path: string) {
  if (confirm(t("scm.diff.discard_confirm", { path }))) {
    await discard([path]);
  }
}

async function handleDiscardAll() {
  if (confirm(t("scm.diff.discard_confirm", { path: "*" }))) {
    const allPaths = status.value.unstaged.map((i) => i.path);
    await discard(allPaths);
  }
}

function handleOpenFileInEditor(path: string) {
  const fileName = getFileName(path);
  openTab({ path, name: fileName });
  router.push({ name: "explorer" }).catch(() => {});
}

function handleSelectFileDiff(path: string, staged: boolean) {
  selectedDiffPath.value = path;
  selectedIsStaged.value = staged;
  // Navigate to SCM view with query parameters
  router.push({
    name: "scm",
    query: {
      path,
      staged: String(staged),
    },
  }).catch(() => {});
}

function openFileContextMenu(event: MouseEvent, item: { path: string; status?: string }, isStaged: boolean) {
  fileContextMenuTarget.value = {
    path: item.path,
    status: item.status,
    isStaged,
  };
  fileContextMenuPos.value = { x: event.clientX, y: event.clientY };
  isFileContextMenuOpen.value = true;
}

const { revealInExplorer } = useWorkspaceBus();

async function handleFileMenuAction(action: FileMenuActionType, file: ContextMenuTargetFile) {
  isFileContextMenuOpen.value = false;
  switch (action) {
    case "open-changes":
      handleSelectFileDiff(file.path, file.isStaged);
      break;
    case "open-file":
      handleOpenFileInEditor(file.path);
      break;
    case "stage":
      await handleStageFile(file.path);
      break;
    case "unstage":
      await handleUnstageFile(file.path);
      break;
    case "discard":
      await handleDiscardFile(file.path);
      break;
    case "reveal-explorer":
      revealInExplorer(file.path);
      router.push({ name: "explorer" }).catch(() => {});
      break;
    case "copy-path":

      try {
        const fullPath = file.path.startsWith("/") ? file.path : `/${file.path}`;
        await navigator.clipboard.writeText(fullPath);
        showToast(t("scm.file_menu.copied_path"), "info");
      } catch {
        showToast("Copy failed", "error");
      }
      break;
    case "copy-relative-path":
      try {
        await navigator.clipboard.writeText(file.path);
        showToast(t("scm.file_menu.copied_relative_path"), "info");
      } catch {
        showToast("Copy failed", "error");
      }
      break;
  }
}

function toggleMoreMenu(event: MouseEvent) {
  if (isMoreMenuOpen.value) {
    isMoreMenuOpen.value = false;
    return;
  }
  const btn = event.currentTarget as HTMLElement | null;
  if (btn) {
    const rect = btn.getBoundingClientRect();
    moreMenuPos.value = {
      x: rect.right + 4,
      y: rect.top,
    };
  } else {
    moreMenuPos.value = { x: event.clientX, y: event.clientY };
  }
  isMoreMenuOpen.value = true;
}

function openModal(action: GitModalActionType, targetRefName?: string) {
  activeModalType.value = action;
  modalTargetRefName.value = targetRefName || currentBranch.value;
  modalVisible.value = true;
}

async function handleMoreAction(action: string) {
  isMoreMenuOpen.value = false;

  if (action === "view-graph") {
    openGitGraph();
    return;
  }

  if (action === "pull") {
    await pull(false);
    return;
  }

  if (action === "pull-rebase") {
    await pull(true);
    return;
  }

  if (action === "push") {
    await pushBranch(currentBranch.value, false, false);
    return;
  }

  if (action === "push-force") {
    await pushBranch(currentBranch.value, false, true);
    return;
  }

  if (action === "clone") {
    showToast(t("scm.actions.more") + ": Clone", "info");
    return;
  }

  if (action === "checkout") {
    isBranchMenuOpen.value = true;
    return;
  }

  if (action === "fetch") {
    await fetchRemote();
    return;
  }

  if (action === "commit") {
    await commit(undefined, false, signCommit.value);
    return;
  }

  if (action === "commit-amend") {
    await commit(undefined, true, signCommit.value);
    return;
  }

  if (action === "commit-staged") {
    await commit(undefined, false, signCommit.value);
    return;
  }

  if (action === "commit-all") {
    await stage([]);
    await commit(undefined, false, signCommit.value);
    return;
  }

  if (action === "stage-all") {
    await handleStageAll();
    return;
  }

  if (action === "unstage-all") {
    await handleUnstageAll();
    return;
  }

  if (action === "discard-all") {
    await handleDiscardAll();
    return;
  }

  if (action === "sync") {
    await sync();
    return;
  }

  if (action === "create-branch") {
    openModal("create-branch");
    return;
  }

  if (action === "rename-branch") {
    openModal("rename-branch");
    return;
  }

  if (action === "delete-branch") {
    openModal("delete-branch");
    return;
  }

  if (action === "merge-branch") {
    openModal("merge-branch");
    return;
  }

  if (action === "rebase-branch") {
    openModal("rebase-branch");
    return;
  }

  if (action === "stash") {
    openModal("stash-uncommitted");
    return;
  }

  if (action === "stash-untracked") {
    openModal("stash-uncommitted");
    return;
  }

  if (action === "reset-uncommitted") {
    openModal("reset-uncommitted");
    return;
  }

  if (action === "clean-untracked") {
    openModal("clean-uncommitted");
    return;
  }

  if (action === "create-tag") {
    openModal("create-tag");
    return;
  }

  if (action === "push-tags") {
    await pushTag("", false);
    return;
  }

  if (action === "push-tags-force") {
    await pushTag("", true);
    return;
  }

  if (action === "delete-tag") {
    openModal("delete-tag");
    return;
  }

  if (action === "show-output") {
    router.push({ name: "terminal" }).catch(() => {});
    return;
  }
}

async function handleModalConfirm(payload: {
  action: GitModalActionType;
  commit?: any;
  targetRefName?: string;
  data?: any;
}) {
  modalVisible.value = false;
  const { action, data } = payload;
  try {
    if (action === "create-tag") {
      await createTag(data.name, "", data.message);
    } else if (action === "create-branch") {
      await createBranchFromCommit(data.name, "", data.checkout);
    } else if (action === "rename-branch") {
      await renameBranch(data.oldName, data.newName);
    } else if (action === "delete-branch") {
      await deleteBranch(data.name, data.force);
    } else if (action === "merge-branch") {
      await mergeBranch(data.branch, data.noFf, data.squash);
    } else if (action === "rebase-branch") {
      await rebaseBranch(data.branch);
    } else if (action === "delete-tag") {
      await deleteTag(data.name);
    } else if (action === "stash-uncommitted") {
      await stashUncommitted(data?.message, data?.includeUntracked);
    } else if (action === "reset-uncommitted") {
      await resetUncommitted();
    } else if (action === "clean-uncommitted") {
      await cleanUntracked();
    }
    await fetchBranches();
    await fetchStatus();
  } catch (err: any) {
    showToast(err.message || t("scm.graph.action_failed"), "error");
  }
}
</script>

<style scoped>
.scm-sidebar-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  color: var(--text-main);
  font-size: 13px;
  user-select: none;
  background-color: var(--card-bg);
  position: relative;
  transition: background-color 0.2s ease, color 0.2s ease;
}

/* Branch bar */
.scm-header-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 12px;
  border-bottom: 1px solid var(--card-border);
}

.branch-selector {
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
  padding: 3px 6px;
  border-radius: 4px;
  transition: background 0.15s ease;
  font-weight: 500;
  max-width: 80%;
}

.branch-selector:hover {
  background: var(--section-bg);
}

.branch-name {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  color: var(--text-main);
}

.branch-sync-pill {
  font-size: 11px;
  background: var(--section-bg);
  color: var(--text-muted);
  padding: 1px 6px;
  border-radius: 10px;
  font-family: monospace;
}

.twistie {
  font-size: 12px;
  opacity: 0.8;
  color: var(--text-muted);
}

.scm-header-actions {
  display: flex;
  align-items: center;
  gap: 4px;
}

.scm-action-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 4px;
  border-radius: 3px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.scm-action-btn:hover {
  color: var(--text-main);
  background: var(--section-bg);
}

.scm-action-btn.is-loading .codicon {
  animation: codicon-spin 1s linear infinite;
}

/* Branch Dropdown */
.branch-dropdown-menu {
  position: absolute;
  top: 36px;
  left: 10px;
  right: 10px;
  background: var(--modal-bg, var(--card-bg));
  border: 1px solid var(--card-border);
  border-radius: 6px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.2);
  z-index: 100;
  max-height: 240px;
  display: flex;
  flex-direction: column;
}

.branch-dropdown-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 6px 10px;
  font-size: 11px;
  text-transform: uppercase;
  color: var(--text-muted);
  border-bottom: 1px solid var(--card-border);
}

.branch-close-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  font-size: 14px;
  cursor: pointer;
}

.branch-close-btn:hover {
  color: var(--text-main);
}

.branch-list {
  overflow-y: auto;
  padding: 4px 0;
}

.branch-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  cursor: pointer;
  color: var(--text-main);
  font-size: 12px;
}

.branch-item:hover {
  background: var(--section-bg);
}

.branch-item.is-current {
  color: var(--accent-cyan);
  font-weight: 600;
}

.branch-item-name {
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.branch-remote-tag {
  font-size: 10px;
  color: var(--text-muted);
  border: 1px solid var(--card-border);
  border-radius: 3px;
  padding: 0 4px;
}

/* Commit Section */
.scm-commit-section {
  padding: 10px 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  border-bottom: 1px solid var(--card-border);
}

.commit-input-wrapper {
  position: relative;
}

.commit-textarea {
  width: 100%;
  box-sizing: border-box;
  background: var(--input-bg);
  color: var(--text-main);
  border: 1px solid var(--card-border);
  border-radius: 4px;
  padding: 6px 8px;
  font-family: inherit;
  font-size: 12px;
  resize: vertical;
  min-height: 54px;
}

.commit-textarea:focus {
  outline: 1px solid var(--accent-cyan);
  border-color: var(--accent-cyan);
}

.commit-options-bar {
  display: flex;
  align-items: center;
  justify-content: flex-start;
  padding: 2px 2px;
}

.gpg-sign-toggle {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  cursor: pointer;
  font-size: 11px;
  color: var(--text-muted);
  user-select: none;
}

.gpg-sign-toggle:hover {
  color: var(--text-main);
}

.gpg-sign-checkbox {
  cursor: pointer;
  accent-color: var(--accent-cyan);
}

.gpg-sign-icon {
  font-size: 13px;
  color: #3fb950;
}

:root[data-theme="light"] .gpg-sign-icon {
  color: #1a7f37;
}

.gpg-sign-text {
  font-size: 11px;
  font-weight: 500;
}

.commit-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  background: var(--accent-cyan);
  color: #ffffff !important;
  border: none;
  padding: 6px 12px;
  border-radius: 4px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
}

.commit-btn:hover:not(:disabled) {
  filter: brightness(1.1);
}

.commit-btn:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

/* Changes Groups */
.scm-changes-group {
  display: flex;
  flex-direction: column;
}

.group-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 10px;
  cursor: pointer;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.5px;
  color: var(--text-muted);
  text-transform: uppercase;
}

.group-header:hover {
  background: var(--section-bg);
}

.group-left {
  display: flex;
  align-items: center;
  gap: 6px;
}

.group-count {
  background: var(--section-bg);
  color: var(--text-muted);
  padding: 1px 6px;
  border-radius: 10px;
  font-size: 10px;
}

.group-actions {
  display: flex;
  align-items: center;
  gap: 2px;
}

.group-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 2px 4px;
  border-radius: 3px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.group-btn:hover {
  color: var(--text-main);
  background: var(--section-bg);
}

.group-items-list {
  display: flex;
  flex-direction: column;
}

.change-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 12px;
  cursor: pointer;
  font-size: 13px;
  position: relative;
  transition: background 0.1s ease;
}

.change-item:hover {
  background: var(--section-bg);
}

.change-item.active {
  background: var(--accent-cyan-glow, rgba(0, 229, 255, 0.15));
  color: var(--accent-cyan);
}

.change-icon {
  display: flex;
  align-items: center;
  min-width: 16px;
}

.change-info {
  display: flex;
  align-items: baseline;
  gap: 6px;
  flex: 1;
  overflow: hidden;
  white-space: nowrap;
}

.change-filename {
  font-weight: 400;
  color: var(--text-main);
}

.change-dirname {
  font-size: 11px;
  color: var(--text-muted);
  overflow: hidden;
  text-overflow: ellipsis;
}

.change-status-badge {
  font-size: 11px;
  font-weight: 700;
  width: 16px;
  text-align: center;
}

.change-status-badge.badge-M {
  color: var(--warning, #f59e0b);
}

.change-status-badge.badge-A,
.change-status-badge.badge-U,
.change-status-badge.badge-\? {
  color: var(--success, #10b981);
}

.change-status-badge.badge-D {
  color: #ef4444;
}

.change-status-badge.badge-R {
  color: #06b6d4;
}

:root[data-theme="light"] .change-status-badge.badge-M {
  color: #b45309;
}

:root[data-theme="light"] .change-status-badge.badge-A,
:root[data-theme="light"] .change-status-badge.badge-U,
:root[data-theme="light"] .change-status-badge.badge-\? {
  color: #059669;
}

:root[data-theme="light"] .change-status-badge.badge-D {
  color: #dc2626;
}

:root[data-theme="light"] .change-status-badge.badge-R {
  color: #0284c7;
}

.change-item-actions {
  display: none;
  align-items: center;
  gap: 2px;
  position: absolute;
  right: 8px;
  background: var(--card-bg);
  padding-left: 4px;
}

.change-item:hover .change-item-actions {
  display: flex;
}

.item-action-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 2px 4px;
  border-radius: 3px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.item-action-btn:hover {
  color: var(--text-main);
  background: var(--section-bg);
}

.empty-hint {
  padding: 8px 16px;
  font-size: 12px;
  color: var(--text-muted);
  font-style: italic;
}
</style>
