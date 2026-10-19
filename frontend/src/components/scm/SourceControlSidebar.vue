<template>
  <div class="scm-sidebar-container" id="scm-sidebar">
    <!-- 0. Repositories Section (Multi-repository) -->
    <div v-if="repositories.length > 0" class="scm-changes-group scm-repositories-group">
      <div class="group-header" @click="isReposExpanded = !isReposExpanded">
        <div class="group-left">
          <span class="codicon twistie" :class="isReposExpanded ? 'codicon-chevron-down' : 'codicon-chevron-right'"></span>
          <span class="group-title">{{ t('scm.repositories') || '存储库' }}</span>
          <span class="group-count">{{ repositories.length }}</span>
        </div>
        <div class="group-actions" @click.stop>
          <button class="group-btn" :title="t('scm.actions.refresh')" @click="handleRefreshRepos">
            <span class="codicon codicon-refresh" :class="{ 'codicon-modifier-spin': isLoadingRepositories }"></span>
          </button>
        </div>
      </div>

      <div v-show="isReposExpanded" class="group-items-list repo-items-list">
        <div
          v-for="repo in repositories"
          :key="repo.id || repo.rootPath"
          class="repo-item"
          :class="{ active: activeRepo?.id === repo.id || activeRepo?.rootPath === repo.rootPath }"
          @click="handleSelectRepo(repo)"
          :title="`${repo.name} (${repo.branch})${repo.rootPath ? ' - ' + repo.rootPath : ''}`"
        >
          <span class="codicon codicon-repo repo-icon"></span>
          <span class="repo-name">{{ repo.name }}</span>
          <span class="repo-branch-pill">
            <span class="repo-branch-name">{{ repo.branch }}{{ repo.isDirty ? '*' : '' }}</span>
            <span v-if="repo.behind > 0 || repo.ahead > 0" class="repo-sync-pill">
              <span v-if="repo.ahead > 0">↑{{ repo.ahead }}</span>
              <span v-if="repo.behind > 0">↓{{ repo.behind }}</span>
            </span>
          </span>
          <div class="repo-item-actions" @click.stop>
            <button
              class="item-action-btn"
              :title="t('scm.actions.more')"
              @click="openRepoActions($event, repo)"
            >
              <span class="codicon codicon-ellipsis"></span>
            </button>
          </div>
        </div>
      </div>
    </div>

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
          :placeholder="commitPlaceholder"
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

      <!-- Quality Guard Status Capsule & Actions -->
      <div class="guard-capsule-bar">
        <div class="guard-status-capsule" :class="guardCapsuleClass" :title="guardTooltip" @click="handleCapsuleClick">
          <span class="codicon" :class="guardIconClass"></span>
          <span class="guard-capsule-text">{{ guardLabelText }}</span>
          <span v-if="guardIssuesCount > 0" class="guard-count-pill">{{ guardIssuesCount }}</span>
        </div>
        <div class="guard-actions-group">
          <button
            v-if="guardStatus.fixable_count > 0"
            class="guard-btn-fix"
            :disabled="isGuardFixing"
            :title="t('guard.actions.auto_fix') || '一键自动修复'"
            @click.stop="handleRunGuardFix"
          >
            <span class="codicon" :class="isGuardFixing ? 'codicon-loading codicon-modifier-spin' : 'codicon-wand'"></span>
            <span>{{ t('guard.actions.fix') || '自动修复' }}</span>
          </button>
          <button
            class="guard-btn-refresh"
            :disabled="isGuardLoading"
            :title="t('guard.actions.rescan') || '重新质检'"
            @click.stop="handleRefreshGuard"
          >
            <span class="codicon codicon-refresh" :class="{ 'codicon-modifier-spin': isGuardLoading }"></span>
          </button>
        </div>
      </div>

      <!-- Expandable Guard Diagnostics Dropdown (Only when there are issues) -->
      <div v-if="isGuardDetailsOpen && guardStatus.issues.length > 0" class="guard-diagnostics-drawer">
        <div class="guard-drawer-header">
          <span class="guard-drawer-title">{{ t('guard.diagnostics.title') || '质量守卫诊断结果' }}</span>
          <button class="guard-close-btn" @click="isGuardDetailsOpen = false">×</button>
        </div>
        <div class="guard-issues-list">
          <div
            v-for="(iss, idx) in guardStatus.issues"
            :key="idx"
            class="guard-issue-item"
            :class="'severity-' + iss.severity"
            @click="handleOpenFileInEditor(iss.file_path)"
          >
            <span class="codicon" :class="iss.severity === 'error' ? 'codicon-error' : 'codicon-warning'"></span>
            <div class="guard-issue-info">
              <span class="guard-issue-file font-mono">{{ iss.file_path }}{{ iss.line ? ':' + iss.line : '' }}</span>
              <span class="guard-issue-msg">{{ iss.message }}</span>
            </div>
            <span v-if="iss.fixable" class="guard-fixable-tag">{{ t('guard.fixable') || '可修复' }}</span>
          </div>
        </div>
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
import { ref, computed, onMounted } from "vue";
import { useRouter, useRoute } from "vue-router";
import { useGit, type GitRepositoryItem } from "../../composables/useGit";
import { useEditorTabs } from "../../composables/useEditorTabs";
import { useWorkspaceBus } from "../../composables/useWorkspaceBus";
import { useToast } from "../../composables/useToast";

import { t } from "../../i18n";
import TechIcon from "../common/TechIcon.vue";
import GitMoreActionsMenu from "./GitMoreActionsMenu.vue";
import GitFileContextMenu, { type ContextMenuTargetFile, type FileMenuActionType } from "./GitFileContextMenu.vue";
import GitActionModal, { type GitModalActionType } from "./GitActionModal.vue";
import { useGuard } from "../../composables/useGuard";

const router = useRouter();
const route = useRoute();
const { showToast } = useToast();
const {
  repositories,
  activeRepo,
  isLoadingRepositories,
  fetchRepositories,
  setActiveRepo,
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

const isReposExpanded = ref(true);
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

const commitPlaceholder = computed(() => {
  return t('scm.sidebar.commit_placeholder_branch', { branch: currentBranch.value }) ||
    `消息(按Enter在“${currentBranch.value}”提交)`;
});

onMounted(async () => {
  await fetchRepositories();
  await Promise.all([fetchStatus(), fetchBranches()]);
});

async function handleSelectRepo(repo: GitRepositoryItem) {
  if (activeRepo.value?.id === repo.id || activeRepo.value?.rootPath === repo.rootPath) {
    return;
  }
  await setActiveRepo(repo);
}

async function handleRefreshRepos() {
  await fetchRepositories();
  await Promise.all([fetchStatus(), fetchBranches()]);
}

async function openRepoActions(event: MouseEvent, repo: GitRepositoryItem) {
  if (activeRepo.value?.id !== repo.id && activeRepo.value?.rootPath !== repo.rootPath) {
    await setActiveRepo(repo);
  }
  moreMenuPos.value = { x: event.clientX, y: event.clientY };
  isMoreMenuOpen.value = true;
}

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

const {
  status: guardStatus,
  isLoading: isGuardLoading,
  isFixing: isGuardFixing,
  fetchStatus: fetchGuardStatus,
  runFix: runGuardFix,
} = useGuard();

const isGuardDetailsOpen = ref(false);

const guardIssuesCount = computed(() => {
  return (guardStatus.value.errors_count || 0) + (guardStatus.value.warnings_count || 0);
});

const guardCapsuleClass = computed(() => {
  if (isGuardLoading.value || isGuardFixing.value) return "status-loading";
  if (guardStatus.value.errors_count > 0) return "status-error";
  if (guardStatus.value.warnings_count > 0) return "status-warning";
  return "status-pass";
});

const guardIconClass = computed(() => {
  if (isGuardLoading.value || isGuardFixing.value) return "codicon-loading codicon-modifier-spin";
  if (guardStatus.value.errors_count > 0) return "codicon-error";
  if (guardStatus.value.warnings_count > 0) return "codicon-warning";
  return "codicon-check";
});

const guardLabelText = computed(() => {
  if (isGuardLoading.value) return t("guard.status.checking") || "守卫质检中...";
  if (isGuardFixing.value) return t("guard.status.fixing") || "正在自动修复...";
  if (guardStatus.value.errors_count > 0) return t("guard.status.blocked") || "拦截违规";
  if (guardStatus.value.fixable_count > 0) return `${guardStatus.value.fixable_count} ` + (t("guard.status.fixable") || "处可自愈");
  return t("guard.status.passed") || "质量守卫通过";
});

const guardTooltip = computed(() => {
  return `${guardLabelText.value} (${guardStatus.value.scanned} files inspected)`;
});

function handleCapsuleClick() {
  if (guardIssuesCount.value === 0) {
    showToast(t("guard.status.passed") || "质量守卫通过，未检测到违规项", "success");
    isGuardDetailsOpen.value = false;
  } else {
    isGuardDetailsOpen.value = !isGuardDetailsOpen.value;
  }
}

async function handleRunGuardFix() {
  const ok = await runGuardFix("staged");
  if (ok) {
    showToast(t("guard.toast.fixed") || "✨ 已自动完成代码格式与卫生修复", "success");
    await fetchStatus();
  }
}

async function handleRefreshGuard() {
  await fetchGuardStatus("staged");
}

async function handleCommit() {
  if (guardStatus.value.errors_count > 0) {
    isGuardDetailsOpen.value = true;
    showToast(t("guard.toast.blocked") || "质量守卫检测到阻断违规，请先解决或修复后提交", "error");
    return;
  }
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

/* ── Repositories Group ── */
.scm-repositories-group {
  border-bottom: 1px solid var(--border-color);
  padding-bottom: 4px;
}

.repo-items-list {
  display: flex;
  flex-direction: column;
  gap: 1px;
  padding: 2px 4px;
}

.repo-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 5px 8px;
  border-radius: 4px;
  cursor: pointer;
  position: relative;
  font-size: 13px;
  color: var(--text-main);
  transition: background 0.15s ease;
}

.repo-item:hover {
  background: var(--theme-list-hover-bg, rgba(255, 255, 255, 0.05));
}

.repo-item.active {
  background: var(--theme-list-active-selection-bg, rgba(55, 148, 255, 0.15));
  color: var(--theme-list-active-selection-fg, var(--text-main));
  font-weight: 500;
}

.repo-item.active::before {
  content: "";
  position: absolute;
  left: 0;
  top: 4px;
  bottom: 4px;
  width: 3px;
  background: var(--accent-color, #3b82f6);
  border-radius: 2px;
}

.repo-icon {
  font-size: 14px;
  color: var(--text-muted);
}

.repo-item.active .repo-icon {
  color: var(--accent-color, #3b82f6);
}

.repo-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.repo-branch-pill {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 11px;
  color: var(--text-muted);
  font-family: var(--font-mono, monospace);
}

.repo-branch-name {
  max-width: 120px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.repo-item.active .repo-branch-name {
  color: var(--text-main);
}

.repo-sync-pill {
  display: inline-flex;
  gap: 3px;
  font-size: 10px;
  background: rgba(255, 255, 255, 0.08);
  padding: 1px 4px;
  border-radius: 3px;
}

.repo-item-actions {
  display: none;
  align-items: center;
  gap: 2px;
  margin-left: 4px;
}

.repo-item:hover .repo-item-actions {
  display: flex;
}

/* Quality Guard Capsule Bar */
.guard-capsule-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
  gap: 8px;
}

.guard-status-capsule {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  min-height: 24px;
  border-radius: 12px;
  font-size: 11px;
  line-height: 1.5;
  cursor: pointer;
  user-select: none;
  transition: all 0.15s ease;
  border: 1px solid transparent;
  box-sizing: border-box;
}

.guard-capsule-text {
  display: inline-block;
  line-height: 1.5;
  vertical-align: middle;
}

.guard-status-capsule.status-pass {
  background: rgba(16, 185, 129, 0.12);
  color: #34d399;
  border-color: rgba(16, 185, 129, 0.3);
}

:root[data-theme="light"] .guard-status-capsule.status-pass {
  background: #ecfdf5;
  color: #047857;
  border-color: #a7f3d0;
}

.guard-status-capsule.status-warning {
  background: rgba(245, 158, 11, 0.15);
  color: #fbbf24;
  border-color: rgba(245, 158, 11, 0.3);
}

:root[data-theme="light"] .guard-status-capsule.status-warning {
  background: #fffbeb;
  color: #b45309;
  border-color: #fde68a;
}

.guard-status-capsule.status-error {
  background: rgba(239, 68, 68, 0.15);
  color: #f87171;
  border-color: rgba(239, 68, 68, 0.3);
}

:root[data-theme="light"] .guard-status-capsule.status-error {
  background: #fef2f2;
  color: #b91c1c;
  border-color: #fecaca;
}

.guard-status-capsule.status-loading {
  background: rgba(255, 255, 255, 0.08);
  color: var(--text-muted, #94a3b8);
}

:root[data-theme="light"] .guard-status-capsule.status-loading {
  background: #f1f5f9;
  color: #64748b;
}

.guard-count-pill {
  padding: 1px 5px;
  border-radius: 8px;
  font-size: 10px;
  font-weight: 600;
  background: currentColor;
  color: var(--card-bg, #1e1e1e);
}

:root[data-theme="light"] .guard-count-pill {
  color: #ffffff;
}

.guard-actions-group {
  display: flex;
  align-items: center;
  gap: 4px;
}

.guard-btn-fix {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 3px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;
  background: #0078d4;
  color: #fff;
  border: none;
  cursor: pointer;
  transition: background 0.15s ease;
}

.guard-btn-fix:hover:not(:disabled) {
  background: #026ec1;
}

.guard-btn-refresh {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  background: transparent;
  color: var(--text-muted, #94a3b8);
  border: none;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.guard-btn-refresh:hover:not(:disabled) {
  background: var(--section-bg, rgba(255, 255, 255, 0.1));
  color: var(--text-main, #fff);
}

:root[data-theme="light"] .guard-btn-refresh {
  color: #64748b;
}

:root[data-theme="light"] .guard-btn-refresh:hover:not(:disabled) {
  background: #e2e8f0;
  color: #0f172a;
}

/* Guard Diagnostics Drawer */
.guard-diagnostics-drawer {
  background: var(--card-bg, #121a2b);
  border: 1px solid var(--card-border, rgba(0, 229, 255, 0.15));
  border-radius: 6px;
  padding: 10px;
  margin-bottom: 8px;
  max-height: 220px;
  overflow-y: auto;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.18);
}

:root[data-theme="light"] .guard-diagnostics-drawer {
  background: #ffffff;
  border-color: #cbd5e1;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.06);
}

.guard-drawer-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
  font-size: 11px;
  font-weight: 600;
  color: var(--text-main, #f8fafc);
}

:root[data-theme="light"] .guard-drawer-header {
  color: #0f172a;
}

.guard-close-btn {
  background: transparent;
  border: none;
  color: var(--text-muted, #94a3b8);
  cursor: pointer;
  font-size: 14px;
  line-height: 1;
}

:root[data-theme="light"] .guard-close-btn {
  color: #64748b;
}

.guard-issues-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.guard-issue-item {
  display: flex;
  align-items: flex-start;
  gap: 6px;
  padding: 5px 8px;
  border-radius: 4px;
  font-size: 11px;
  line-height: 1.4;
  background: rgba(255, 255, 255, 0.04);
  cursor: pointer;
  transition: background 0.1s ease;
}

:root[data-theme="light"] .guard-issue-item {
  background: #f8fafc;
  border: 1px solid #f1f5f9;
}

.guard-issue-item:hover {
  background: rgba(255, 255, 255, 0.08);
}

:root[data-theme="light"] .guard-issue-item:hover {
  background: #f1f5f9;
}

.guard-issue-item.severity-error {
  color: #f87171;
}

:root[data-theme="light"] .guard-issue-item.severity-error {
  color: #b91c1c;
}

.guard-issue-item.severity-warning {
  color: #fbbf24;
}

:root[data-theme="light"] .guard-issue-item.severity-warning {
  color: #b45309;
}

.guard-issue-info {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-width: 0;
}

.guard-issue-file {
  font-size: 10px;
  color: var(--text-main, #f8fafc);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

:root[data-theme="light"] .guard-issue-file {
  color: #0f172a;
}

.guard-issue-msg {
  font-size: 11px;
  color: var(--text-muted, #94a3b8);
}

:root[data-theme="light"] .guard-issue-msg {
  color: #475569;
}

.guard-fixable-tag {
  font-size: 9px;
  padding: 1px 4px;
  border-radius: 3px;
  background: rgba(16, 185, 129, 0.2);
  color: #34d399;
}

:root[data-theme="light"] .guard-fixable-tag {
  background: #ecfdf5;
  color: #047857;
  border: 1px solid #a7f3d0;
}

.guard-issues-empty {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 10px;
  font-size: 11px;
  line-height: 1.5;
  color: #34d399;
  background: rgba(16, 185, 129, 0.12);
  border: 1px solid rgba(16, 185, 129, 0.25);
  border-radius: 4px;
  box-sizing: border-box;
}

:root[data-theme="light"] .guard-issues-empty {
  color: #047857;
  background: #ecfdf5;
  border-color: #a7f3d0;
}
</style>
