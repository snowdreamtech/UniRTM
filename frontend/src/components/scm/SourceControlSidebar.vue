<template>
  <div class="scm-sidebar-container" id="scm-sidebar">
    <!-- 1. SCM Header: Branch info & Sync -->
    <div class="scm-header-bar">
      <div class="branch-selector" :title="`当前分支: ${currentBranch}${status.upstream ? ' -> ' + status.upstream : ''}`" @click="isBranchMenuOpen = !isBranchMenuOpen">
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
          title="查看 Git Graph 提交图谱 (对齐 vscode-git-graph)"
          @click.stop="openGitGraph"
        >
          <span class="codicon codicon-git-commit"></span>
        </button>
        <button
          class="scm-action-btn"
          title="刷新状态"
          :class="{ 'is-loading': isLoading }"
          @click.stop="handleRefresh"
        >
          <span class="codicon codicon-refresh"></span>
        </button>
      </div>
    </div>

    <!-- Branch Dropdown Popover -->
    <div v-if="isBranchMenuOpen" class="branch-dropdown-menu">
      <div class="branch-dropdown-header">
        <span>切换分支</span>
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
          <span v-if="b.remote" class="branch-remote-tag">remote</span>
        </div>
      </div>
    </div>

    <!-- 2. Commit Message Box -->
    <div class="scm-commit-section">
      <div class="commit-input-wrapper">
        <textarea
          v-model="commitMessage"
          class="commit-textarea"
          placeholder="提交信息 (Ctrl+Enter 提交)"
          rows="3"
          @keydown.ctrl.enter.prevent="handleCommit"
          @keydown.meta.enter.prevent="handleCommit"
        ></textarea>
      </div>

      <div class="commit-options-bar">
        <label class="gpg-sign-toggle" title="使用 GPG 签名签署提交 (-S)">
          <input type="checkbox" v-model="signCommit" class="gpg-sign-checkbox" />
          <span class="codicon codicon-verified gpg-sign-icon"></span>
          <span class="gpg-sign-text">GPG 签名提交</span>
        </label>
      </div>

      <button
        class="commit-btn"
        :disabled="isCommitting || !commitMessage.trim()"
        @click="handleCommit"
      >
        <span v-if="isCommitting" class="codicon codicon-loading codicon-modifier-spin"></span>
        <span v-else class="codicon codicon-check"></span>
        <span>提交</span>
      </button>
    </div>

    <!-- 3. Staged Changes Section -->
    <div class="scm-changes-group">
      <div class="group-header" @click="isStagedExpanded = !isStagedExpanded">
        <div class="group-left">
          <span class="codicon twistie" :class="isStagedExpanded ? 'codicon-chevron-down' : 'codicon-chevron-right'"></span>
          <span class="group-title">暂存的更改</span>
          <span class="group-count">{{ status.staged.length }}</span>
        </div>
        <div class="group-actions" @click.stop>
          <button
            v-if="status.staged.length > 0"
            class="group-btn"
            title="全部取消暂存"
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
              title="在编辑器中打开"
              @click="handleOpenFileInEditor(item.path)"
            >
              <span class="codicon codicon-go-to-file"></span>
            </button>
            <button
              class="item-action-btn"
              title="取消暂存更改"
              @click="handleUnstageFile(item.path)"
            >
              <span class="codicon codicon-remove"></span>
            </button>
          </div>
        </div>

        <div v-if="status.staged.length === 0" class="empty-hint">
          <span>无暂存更改</span>
        </div>
      </div>
    </div>

    <!-- 4. Changes (Unstaged) Section -->
    <div class="scm-changes-group">
      <div class="group-header" @click="isUnstagedExpanded = !isUnstagedExpanded">
        <div class="group-left">
          <span class="codicon twistie" :class="isUnstagedExpanded ? 'codicon-chevron-down' : 'codicon-chevron-right'"></span>
          <span class="group-title">更改</span>
          <span class="group-count">{{ status.unstaged.length }}</span>
        </div>
        <div class="group-actions" @click.stop>
          <button
            v-if="status.unstaged.length > 0"
            class="group-btn"
            title="全部放弃更改"
            @click="handleDiscardAll"
          >
            <span class="codicon codicon-discard"></span>
          </button>
          <button
            v-if="status.unstaged.length > 0"
            class="group-btn"
            title="全部暂存更改"
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
              title="在编辑器中打开"
              @click="handleOpenFileInEditor(item.path)"
            >
              <span class="codicon codicon-go-to-file"></span>
            </button>
            <button
              class="item-action-btn"
              title="放弃更改"
              @click="handleDiscardFile(item.path)"
            >
              <span class="codicon codicon-discard"></span>
            </button>
            <button
              class="item-action-btn"
              title="暂存更改"
              @click="handleStageFile(item.path)"
            >
              <span class="codicon codicon-add"></span>
            </button>
          </div>
        </div>

        <div v-if="status.unstaged.length === 0" class="empty-hint">
          <span>工作区干净，无未暂存更改</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useRouter, useRoute } from "vue-router";
import { useGit } from "../../composables/useGit";
import { useEditorTabs } from "../../composables/useEditorTabs";
import TechIcon from "../common/TechIcon.vue";

const router = useRouter();
const route = useRoute();
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
} = useGit();

const { openTab } = useEditorTabs();

const isStagedExpanded = ref(true);
const isUnstagedExpanded = ref(true);
const isBranchMenuOpen = ref(false);
const signCommit = ref(true);

const selectedDiffPath = ref<string>("");
const selectedIsStaged = ref<boolean>(false);

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
  if (confirm(`确定放弃对 "${path}" 的所有未暂存更改吗？此操作无法撤销。`)) {
    await discard([path]);
  }
}

async function handleDiscardAll() {
  if (confirm("确定放弃所有未暂存的更改吗？此操作无法撤销。")) {
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
