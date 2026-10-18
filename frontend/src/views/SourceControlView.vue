<template>
  <div class="scm-view-container" id="view-scm">
    <!-- Top Diff Toolbar (When viewing a specific file diff) -->
    <div v-if="activeDiffFile" class="diff-header-bar">
      <div class="diff-title-group">
        <span class="codicon codicon-diff"></span>
        <span class="diff-file-path font-mono">{{ activeDiffFile }}</span>
        <span
          v-if="isCommitDiff"
          class="diff-status-pill pill-commit"
          :title="`历史提交: ${commitHash}`"
        >
          提交 {{ commitHash.slice(0, 7) }}{{ parentHash ? ' vs ' + parentHash.slice(0, 7) : '' }}
        </span>
        <span
          v-else
          class="diff-status-pill"
          :class="isStaged ? 'pill-staged' : 'pill-unstaged'"
        >
          {{ isStaged ? "已暂存 (Index)" : "工作区 (Working Tree)" }}
        </span>
      </div>

      <div class="diff-actions">
        <!-- Toggle Inline vs Side-by-Side Diff -->
        <button
          class="diff-btn"
          :title="isInlineDiff ? '切换为并排比对 (Side by Side)' : '切换为内联比对 (Inline)'"
          @click="toggleInlineDiff"
        >
          <span class="codicon" :class="isInlineDiff ? 'codicon-split-horizontal' : 'codicon-list-flat'"></span>
          <span>{{ isInlineDiff ? "并排比对" : "内联比对" }}</span>
        </button>

        <!-- Working Tree Stage / Unstage / Discard Actions (only when not commit diff) -->
        <template v-if="!isCommitDiff">
          <!-- Stage / Unstage Action -->
          <button
            v-if="!isStaged"
            class="diff-btn primary"
            title="暂存此更改"
            @click="handleStageActiveFile"
          >
            <span class="codicon codicon-add"></span>
            <span>暂存更改</span>
          </button>
          <button
            v-else
            class="diff-btn"
            title="取消暂存此更改"
            @click="handleUnstageActiveFile"
          >
            <span class="codicon codicon-remove"></span>
            <span>取消暂存</span>
          </button>

          <!-- Discard Action (unstaged only) -->
          <button
            v-if="!isStaged"
            class="diff-btn danger"
            title="放弃此文件的所有更改"
            @click="handleDiscardActiveFile"
          >
            <span class="codicon codicon-discard"></span>
            <span>放弃更改</span>
          </button>
        </template>

        <!-- Open File in Editor -->
        <button
          class="diff-btn"
          title="在代码编辑器中打开"
          @click="handleOpenFileInEditor"
        >
          <span class="codicon codicon-go-to-file"></span>
          <span>打开文件</span>
        </button>

        <!-- Close Diff / Back to Overview or Graph -->
        <button
          class="diff-btn close-btn"
          title="关闭比对视图"
          @click="closeDiffView"
        >
          <span class="codicon codicon-close"></span>
        </button>
      </div>
    </div>

    <!-- 1. Active Monaco Diff Editor Container -->
    <div
      v-show="activeDiffFile"
      ref="diffEditorContainer"
      class="monaco-diff-container"
    ></div>

    <!-- 2. SCM Overview Hub (When no diff file is selected) -->
    <div v-if="!activeDiffFile" class="scm-overview-hub">
      <div class="overview-header">
        <div class="overview-title-row">
          <div class="title-with-icon">
            <span class="codicon codicon-source-control scm-icon-large"></span>
            <div>
              <h2 class="repo-title">Git 源代码管理</h2>
              <p class="repo-subtitle">统一版本控制中心 · 分支、变更与提交图谱融合管理</p>
            </div>
          </div>

          <div class="header-right-group">
            <!-- Segmented Mode Tabs Switcher -->
            <div class="view-mode-tabs">
              <button
                class="mode-tab-btn"
                :class="{ active: currentTab === 'changes' }"
                @click="switchTab('changes')"
              >
                <span class="codicon codicon-diff-modified"></span>
                <span>工作区变更 ({{ totalChanges }})</span>
              </button>
              <button
                class="mode-tab-btn"
                :class="{ active: currentTab === 'graph' }"
                @click="switchTab('graph')"
              >
                <span class="codicon codicon-git-commit"></span>
                <span>提交图谱 (Git Graph)</span>
              </button>
            </div>

            <div class="header-branch-pill">
              <span class="codicon codicon-git-branch"></span>
              <span class="branch-name">{{ currentBranch }}</span>
              <span v-if="status.ahead > 0 || status.behind > 0" class="branch-sync-badge">
                <span v-if="status.behind > 0">↓ {{ status.behind }}</span>
                <span v-if="status.ahead > 0">↑ {{ status.ahead }}</span>
              </span>
            </div>
          </div>
        </div>
      </div>

      <!-- Mode 1: Working Tree Changes & Metrics -->
      <template v-if="currentTab === 'changes'">
        <!-- Quick Metrics Cards -->
        <div class="scm-stats-grid">
          <div class="stat-card">
            <div class="stat-icon-wrapper blue">
              <span class="codicon codicon-git-branch"></span>
            </div>
            <div class="stat-details">
              <span class="stat-value">{{ currentBranch }}</span>
              <span class="stat-label">当前活跃分支</span>
            </div>
          </div>

          <div class="stat-card">
            <div class="stat-icon-wrapper green">
              <span class="codicon codicon-check-all"></span>
            </div>
            <div class="stat-details">
              <span class="stat-value">{{ status.staged.length }}</span>
              <span class="stat-label">暂存更改 (Staged)</span>
            </div>
          </div>

          <div class="stat-card">
            <div class="stat-icon-wrapper amber">
              <span class="codicon codicon-diff-modified"></span>
            </div>
            <div class="stat-details">
              <span class="stat-value">{{ status.unstaged.length }}</span>
              <span class="stat-label">未暂存更改 (Unstaged)</span>
            </div>
          </div>

          <div class="stat-card">
            <div class="stat-icon-wrapper purple">
              <span class="codicon codicon-history"></span>
            </div>
            <div class="stat-details">
              <span class="stat-value">{{ totalChanges }}</span>
              <span class="stat-label">待处理文件总数</span>
            </div>
          </div>
        </div>

        <!-- Quick Actions Bar -->
        <div class="overview-actions-card">
          <div class="actions-group">
            <button
              class="action-pill primary"
              :disabled="status.unstaged.length === 0"
              @click="stage([])"
            >
              <span class="codicon codicon-add"></span>
              <span>暂存所有更改 (Stage All)</span>
            </button>
            <button
              class="action-pill"
              :disabled="status.staged.length === 0"
              @click="unstage([])"
            >
              <span class="codicon codicon-remove"></span>
              <span>取消全部暂存 (Unstage All)</span>
            </button>
            <button
              class="action-pill"
              @click="handleManualRefresh"
            >
              <span class="codicon codicon-refresh"></span>
              <span>刷新状态 (Refresh)</span>
            </button>
          </div>
        </div>

        <!-- Split Columns: Staged and Unstaged Files -->
        <div class="changes-split-view">
          <!-- Staged Box -->
          <div class="changes-box">
            <div class="box-header">
              <span class="box-title">暂存更改 ({{ status.staged.length }})</span>
              <span v-if="status.staged.length > 0" class="box-badge green">已准备提交</span>
            </div>
            <div class="box-content">
              <div
                v-for="item in status.staged"
                :key="'staged:' + item.path"
                class="box-file-item"
                @click="openDiff(item.path, true)"
              >
                <TechIcon :id="getFileName(item.path)" :size="16" />
                <span class="file-name">{{ getFileName(item.path) }}</span>
                <span class="file-dir">{{ getDirName(item.path) }}</span>
                <span class="status-tag" :class="'tag-' + item.status">{{ item.status }}</span>
              </div>
              <div v-if="status.staged.length === 0" class="empty-state">
                <span class="codicon codicon-inbox"></span>
                <span>暂无已暂存的更改</span>
              </div>
            </div>
          </div>

          <!-- Unstaged Box -->
          <div class="changes-box">
            <div class="box-header">
              <span class="box-title">工作区更改 ({{ status.unstaged.length }})</span>
              <span v-if="status.unstaged.length > 0" class="box-badge amber">需处理</span>
            </div>
            <div class="box-content">
              <div
                v-for="item in status.unstaged"
                :key="'unstaged:' + item.path"
                class="box-file-item"
                @click="openDiff(item.path, false)"
              >
                <TechIcon :id="getFileName(item.path)" :size="16" />
                <span class="file-name">{{ getFileName(item.path) }}</span>
                <span class="file-dir">{{ getDirName(item.path) }}</span>
                <span class="status-tag" :class="'tag-' + item.status">{{ item.status }}</span>
              </div>
              <div v-if="status.unstaged.length === 0" class="empty-state">
                <span class="codicon codicon-pass"></span>
                <span>工作区树干净 (Working tree clean)</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Recent Commits Section -->
        <div class="commits-history-card">
          <div class="box-header">
            <div class="title-with-icon">
              <span class="codicon codicon-history"></span>
              <span class="box-title">最近提交历史 (Commit History)</span>
            </div>
            <button class="action-pill small" @click="fetchLog">
              <span class="codicon codicon-refresh"></span>
              <span>刷新</span>
            </button>
          <div class="commits-list">
            <div v-for="c in commits" :key="c.hash" class="commit-item">
              <span class="commit-hash font-mono">{{ c.shortHash }}</span>
              <span
                v-if="c.gpg && c.gpg.status !== 'none'"
                class="gpg-mini-badge"
                :class="'gpg-' + c.gpg.status"
                :title="c.gpg.status === 'verified' ? `GPG 签名已验证 • ${c.gpg.signer || c.gpg.keyId || ''}` : `GPG 状态: ${c.gpg.status}`"
              >
                <span class="codicon" :class="c.gpg.status === 'verified' ? 'codicon-verified' : 'codicon-shield'"></span>
                <span>{{ c.gpg.status === 'verified' ? 'Verified' : c.gpg.status }}</span>
              </span>
              <span class="commit-message">{{ c.message }}</span>
              <div class="commit-meta">
                <span class="commit-author">{{ c.author }}</span>
                <span class="commit-time">{{ formatTime(c.timestamp) }}</span>
              </div>
            </div>
            <div v-if="commits.length === 0" class="empty-state">
              <span class="codicon codicon-history"></span>
              <span>暂无提交历史记录</span>
            </div>
          </div>
        </div>
      </template>

      <!-- Mode 2: Interactive Git Graph View (Aligned with vscode-git-graph) -->
      <GitGraphView
        v-else-if="currentTab === 'graph'"
        class="git-graph-embedded"
        @open-commit-diff="handleOpenCommitDiff"
        @open-file="handleOpenFileFromGraph"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, watch, nextTick } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useGit } from "../composables/useGit";
import { useEditorTabs } from "../composables/useEditorTabs";
import { useTheme } from "../composables/useTheme";
import { loadMonaco, getLanguageForFile } from "../services/editor/monacoLoader";
import TechIcon from "../components/common/TechIcon.vue";
import GitGraphView from "../components/scm/GitGraphView.vue";

const route = useRoute();
const router = useRouter();
const { isDarkTheme } = useTheme();

const {
  status,
  currentBranch,
  totalChanges,
  commits,
  fetchStatus,
  fetchDiff,
  fetchCommitDiff,
  fetchLog,
  stage,
  unstage,
  discard,
} = useGit();

const currentTab = ref<"changes" | "graph">(
  route.query.tab === "graph" ? "graph" : "changes"
);

function switchTab(tab: "changes" | "graph") {
  currentTab.value = tab;
  router.replace({
    query: {
      ...route.query,
      tab,
    },
  });
}

function formatTime(ts: number): string {
  if (!ts) return "";
  const d = new Date(ts * 1000);
  return d.toLocaleDateString() + " " + d.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
}

const { openTab } = useEditorTabs();

const diffEditorContainer = ref<HTMLElement | null>(null);
let monacoDiffEditor: any = null;
let originalModel: any = null;
let modifiedModel: any = null;

const isInlineDiff = ref(false);

const activeDiffFile = computed(() => {
  return (route.query.path as string) || "";
});

const isStaged = computed(() => {
  return route.query.staged === "true";
});

const commitHash = computed(() => {
  return (route.query.commit as string) || "";
});

const parentHash = computed(() => {
  return (route.query.parent as string) || "";
});

const isCommitDiff = computed(() => {
  return !!commitHash.value;
});

function getFileName(filePath: string): string {
  const parts = filePath.split("/");
  return parts[parts.length - 1] || filePath;
}

function getDirName(filePath: string): string {
  const parts = filePath.split("/");
  if (parts.length <= 1) return "";
  return parts.slice(0, parts.length - 1).join("/");
}

async function renderDiff(filePath: string, staged: boolean) {
  if (!filePath) return;

  let diff: any = null;
  if (isCommitDiff.value) {
    diff = await fetchCommitDiff(commitHash.value, filePath, parentHash.value || undefined);
  } else {
    diff = await fetchDiff(filePath, staged);
  }
  if (!diff) return;

  await nextTick();
  if (!diffEditorContainer.value) return;

  const monaco = await loadMonaco();

  const lang = getLanguageForFile(getFileName(filePath));

  // Dispose previous models and editor if already present
  if (originalModel) {
    originalModel.dispose();
    originalModel = null;
  }
  if (modifiedModel) {
    modifiedModel.dispose();
    modifiedModel = null;
  }

  originalModel = monaco.editor.createModel(diff.original || "", lang);
  modifiedModel = monaco.editor.createModel(diff.modified || "", lang);

  if (!monacoDiffEditor) {
    monacoDiffEditor = monaco.editor.createDiffEditor(diffEditorContainer.value, {
      renderSideBySide: !isInlineDiff.value,
      readOnly: true,
      automaticLayout: true,
      theme: isDarkTheme.value ? "vs-dark" : "vs",
      minimap: { enabled: false },
      scrollBeyondLastLine: false,
      lineNumbers: "on",
      fontSize: 13,
      fontFamily: "'Fira Code', 'JetBrains Mono', Consolas, 'Courier New', monospace",
    });
  } else {
    monacoDiffEditor.updateOptions({
      renderSideBySide: !isInlineDiff.value,
    });
  }

  monacoDiffEditor.setModel({
    original: originalModel,
    modified: modifiedModel,
  });
}

function toggleInlineDiff() {
  isInlineDiff.value = !isInlineDiff.value;
  if (monacoDiffEditor) {
    monacoDiffEditor.updateOptions({
      renderSideBySide: !isInlineDiff.value,
    });
  }
}

watch(isDarkTheme, (dark) => {
  if (monacoDiffEditor) {
    loadMonaco().then((monaco) => {
      monaco.editor.setTheme(dark ? "vs-dark" : "vs");
    });
  }
});

async function handleStageActiveFile() {
  if (!activeDiffFile.value) return;
  await stage([activeDiffFile.value]);
  // update route to staged
  router.replace({
    name: "scm",
    query: {
      path: activeDiffFile.value,
      staged: "true",
    },
  });
}

async function handleUnstageActiveFile() {
  if (!activeDiffFile.value) return;
  await unstage([activeDiffFile.value]);
  // update route to unstaged
  router.replace({
    name: "scm",
    query: {
      path: activeDiffFile.value,
      staged: "false",
    },
  });
}

async function handleDiscardActiveFile() {
  if (!activeDiffFile.value) return;
  if (confirm(`确定放弃对 "${activeDiffFile.value}" 的所有更改吗？此操作无法撤销。`)) {
    await discard([activeDiffFile.value]);
    closeDiffView();
  }
}

function handleOpenFileInEditor() {
  if (!activeDiffFile.value) return;
  const fileName = getFileName(activeDiffFile.value);
  openTab({ path: activeDiffFile.value, name: fileName });
  router.push({ name: "explorer" }).catch(() => {});
}

function closeDiffView() {
  if (isCommitDiff.value) {
    router.replace({
      name: "scm",
      query: {
        scope: route.query.scope || "workspace",
        tab: "graph",
      },
    });
  } else {
    router.replace({
      name: "scm",
      query: {
        scope: route.query.scope || "workspace",
        tab: currentTab.value,
      },
    });
  }
}

function handleOpenCommitDiff(payload: { path: string; hash: string; parentHash?: string }) {
  router.push({
    name: "scm",
    query: {
      scope: route.query.scope || "workspace",
      tab: "graph",
      path: payload.path,
      commit: payload.hash,
      parent: payload.parentHash || "",
    },
  });
}

function handleOpenFileFromGraph(path: string) {
  const fileName = getFileName(path);
  openTab({ path, name: fileName });
  router.push({ name: "explorer" }).catch(() => {});
}

function openDiff(path: string, staged: boolean) {
  router.push({
    name: "scm",
    query: {
      path,
      staged: String(staged),
    },
  });
}

function handleManualRefresh() {
  fetchStatus().catch(() => {});
  fetchLog().catch(() => {});
}

watch(
  () => [route.query.path, route.query.staged, route.query.commit, route.query.parent],
  ([newPath, newStaged]) => {
    if (newPath && typeof newPath === "string") {
      renderDiff(newPath, newStaged === "true");
    } else {
      if (monacoDiffEditor) {
        monacoDiffEditor.dispose();
        monacoDiffEditor = null;
      }
    }
  },
  { immediate: true }
);

watch(
  () => route.query.tab,
  (tab) => {
    if (tab === "graph" || tab === "changes") {
      currentTab.value = tab;
    }
  }
);

onMounted(() => {
  fetchStatus().catch(() => {});
  fetchLog().catch(() => {});
});

onBeforeUnmount(() => {
  if (monacoDiffEditor) {
    monacoDiffEditor.dispose();
    monacoDiffEditor = null;
  }
  if (originalModel) {
    originalModel.dispose();
    originalModel = null;
  }
  if (modifiedModel) {
    modifiedModel.dispose();
    modifiedModel = null;
  }
});
</script>

<style scoped>
.scm-view-container {
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 100%;
  overflow: hidden;
  background-color: var(--bg-color);
  color: var(--text-main);
  transition: background-color 0.2s ease, color 0.2s ease;
}

/* Diff Header Bar */
.diff-header-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 16px;
  background-color: var(--card-bg);
  border-bottom: 1px solid var(--card-border);
  gap: 12px;
}

.diff-title-group {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.diff-file-path {
  font-weight: 500;
  font-size: 13px;
  color: var(--text-main);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.diff-status-pill {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 12px;
  font-weight: 500;
}

.pill-staged {
  background: rgba(16, 185, 129, 0.15);
  color: var(--success, #10b981);
  border: 1px solid rgba(16, 185, 129, 0.35);
}

.pill-unstaged {
  background: rgba(245, 158, 11, 0.15);
  color: var(--warning, #f59e0b);
  border: 1px solid rgba(245, 158, 11, 0.35);
}

:root[data-theme="light"] .pill-staged {
  background: rgba(5, 150, 105, 0.12);
  color: #059669;
  border-color: rgba(5, 150, 105, 0.3);
}

:root[data-theme="light"] .pill-unstaged {
  background: rgba(217, 119, 6, 0.12);
  color: #b45309;
  border-color: rgba(217, 119, 6, 0.3);
}

.diff-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.diff-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  font-size: 12px;
  border-radius: 4px;
  border: 1px solid var(--card-border);
  background: var(--card-bg);
  color: var(--text-main);
  cursor: pointer;
  transition: all 0.15s ease;
}

.diff-btn:hover {
  background: var(--section-bg);
}

.diff-btn.primary {
  background: var(--accent-cyan);
  color: #ffffff !important;
  border-color: var(--accent-cyan);
}

.diff-btn.primary:hover {
  filter: brightness(1.1);
}

.diff-btn.danger {
  background: rgba(239, 68, 68, 0.12);
  color: #ef4444;
  border-color: rgba(239, 68, 68, 0.3);
}

.diff-btn.danger:hover {
  background: rgba(239, 68, 68, 0.22);
}

:root[data-theme="light"] .diff-btn.danger {
  background: rgba(220, 38, 38, 0.08);
  color: #dc2626;
  border-color: rgba(220, 38, 38, 0.25);
}

.diff-btn.close-btn {
  padding: 4px 8px;
}

/* Monaco Container */
.monaco-diff-container {
  flex: 1;
  width: 100%;
  height: 100%;
  overflow: hidden;
}

/* SCM Overview Hub */
.scm-overview-hub {
  padding: 24px 32px;
  display: flex;
  flex-direction: column;
  gap: 24px;
  overflow-y: auto;
  height: 100%;
  box-sizing: border-box;
}

.overview-header {
  border-bottom: 1px solid var(--card-border);
  padding-bottom: 16px;
}

.overview-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.title-with-icon {
  display: flex;
  align-items: center;
  gap: 16px;
}

.scm-icon-large {
  font-size: 36px;
  color: var(--accent-cyan);
}

.repo-title {
  font-size: 20px;
  font-weight: 600;
  margin: 0;
  color: var(--text-main);
}

.repo-subtitle {
  font-size: 13px;
  color: var(--text-muted);
  margin: 4px 0 0 0;
}

.header-branch-pill {
  display: flex;
  align-items: center;
  gap: 8px;
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  color: var(--text-main);
  padding: 6px 14px;
  border-radius: 20px;
  font-weight: 500;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
}

.branch-sync-badge {
  font-size: 11px;
  background: var(--accent-cyan);
  color: #ffffff;
  padding: 1px 6px;
  border-radius: 10px;
}

/* Stats Grid */
.scm-stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
}

.stat-card {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 14px 18px;
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  border-radius: 6px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
  transition: transform 0.15s ease, box-shadow 0.15s ease;
}

.stat-card:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}

.stat-icon-wrapper {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 8px;
  font-size: 20px;
}

.stat-icon-wrapper.blue {
  background: rgba(2, 132, 199, 0.15);
  color: var(--accent-cyan);
}

.stat-icon-wrapper.green {
  background: rgba(16, 185, 129, 0.15);
  color: var(--success, #10b981);
}

.stat-icon-wrapper.amber {
  background: rgba(245, 158, 11, 0.15);
  color: var(--warning, #f59e0b);
}

.stat-icon-wrapper.purple {
  background: rgba(168, 85, 247, 0.15);
  color: var(--accent-purple, #a855f7);
}

.stat-details {
  display: flex;
  flex-direction: column;
}

.stat-value {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-main);
}

.stat-label {
  font-size: 12px;
  color: var(--text-muted);
}

/* Actions Card */
.overview-actions-card {
  display: flex;
  align-items: center;
}

.actions-group {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.action-pill {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  font-size: 13px;
  border-radius: 4px;
  border: 1px solid var(--card-border);
  background: var(--card-bg);
  color: var(--text-main);
  cursor: pointer;
  transition: all 0.15s ease;
}

.action-pill:hover:not(:disabled) {
  background: var(--section-bg);
}

.action-pill.primary {
  background: var(--accent-cyan);
  color: #ffffff !important;
  border-color: var(--accent-cyan);
}

.action-pill.primary:hover:not(:disabled) {
  filter: brightness(1.1);
}

.action-pill:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

/* Split Changes Boxes */
.changes-split-view {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
  flex: 1;
}

.changes-box {
  display: flex;
  flex-direction: column;
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  border-radius: 6px;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
}

.box-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px;
  background: var(--section-bg);
  border-bottom: 1px solid var(--card-border);
  color: var(--text-main);
}

.box-title {
  font-size: 12px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.box-badge {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 10px;
  font-weight: 500;
}

.box-badge.green {
  background: rgba(16, 185, 129, 0.15);
  color: var(--success, #10b981);
}

.box-badge.amber {
  background: rgba(245, 158, 11, 0.15);
  color: var(--warning, #f59e0b);
}

:root[data-theme="light"] .box-badge.green {
  background: rgba(5, 150, 105, 0.12);
  color: #059669;
}

:root[data-theme="light"] .box-badge.amber {
  background: rgba(217, 119, 6, 0.12);
  color: #b45309;
}

.box-content {
  display: flex;
  flex-direction: column;
  overflow-y: auto;
  max-height: 380px;
}

.box-file-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 14px;
  cursor: pointer;
  border-bottom: 1px solid var(--card-border);
  transition: background 0.15s ease;
}

.box-file-item:hover {
  background: var(--section-bg);
}

.file-name {
  font-weight: 500;
  color: var(--text-main);
}

.file-dir {
  flex: 1;
  font-size: 11px;
  color: var(--text-muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.status-tag {
  font-size: 11px;
  font-weight: 700;
  width: 16px;
  text-align: center;
}

.status-tag.tag-M {
  color: #f59e0b;
}

.status-tag.tag-A,
.status-tag.tag-\? {
  color: #10b981;
}

.status-tag.tag-D {
  color: #ef4444;
}

:root[data-theme="light"] .status-tag.tag-M {
  color: #b45309;
}

:root[data-theme="light"] .status-tag.tag-A,
:root[data-theme="light"] .status-tag.tag-\? {
  color: #059669;
}

:root[data-theme="light"] .status-tag.tag-D {
  color: #dc2626;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 36px 16px;
  gap: 8px;
  color: var(--text-muted);
  font-size: 12px;
}

.empty-state .codicon {
  font-size: 24px;
  opacity: 0.5;
}

/* Commits History Card */
.commits-history-card {
  display: flex;
  flex-direction: column;
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  border-radius: 6px;
  overflow: hidden;
  margin-bottom: 24px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
}

.action-pill.small {
  padding: 3px 8px;
  font-size: 11px;
}

.commits-list {
  display: flex;
  flex-direction: column;
  max-height: 280px;
  overflow-y: auto;
}

.commit-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 14px;
  border-bottom: 1px solid var(--card-border);
  font-size: 12px;
  transition: background 0.15s ease;
}

.commit-item:hover {
  background: var(--section-bg);
}

.commit-hash {
  color: var(--accent-cyan);
  font-weight: 600;
  min-width: 65px;
}

.gpg-mini-badge {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  font-size: 10px;
  font-weight: 600;
  padding: 1px 6px;
  border-radius: 10px;
  flex-shrink: 0;
}

.gpg-mini-badge .codicon {
  font-size: 11px;
}

.gpg-mini-badge.gpg-verified {
  background: rgba(46, 160, 67, 0.15);
  color: #3fb950;
  border: 1px solid rgba(46, 160, 67, 0.4);
}

:root[data-theme="light"] .gpg-mini-badge.gpg-verified {
  background: rgba(46, 160, 67, 0.12);
  color: #1a7f37;
  border-color: rgba(26, 127, 55, 0.35);
}

.gpg-mini-badge.gpg-unverified {
  background: rgba(210, 153, 34, 0.15);
  color: #e3b341;
  border: 1px solid rgba(210, 153, 34, 0.4);
}

:root[data-theme="light"] .gpg-mini-badge.gpg-unverified {
  background: rgba(154, 103, 0, 0.12);
  color: #9a6700;
  border-color: rgba(154, 103, 0, 0.35);
}

.gpg-mini-badge.gpg-bad,
.gpg-mini-badge.gpg-revoked {
  background: rgba(248, 81, 73, 0.15);
  color: #f85149;
  border: 1px solid rgba(248, 81, 73, 0.4);
}

.commit-message {
  flex: 1;
  color: var(--text-main);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.commit-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  color: var(--text-muted);
  font-size: 11px;
}

.commit-author {
  color: var(--text-muted);
}

.header-right-group {
  display: flex;
  align-items: center;
  gap: 12px;
}

.view-mode-tabs {
  display: flex;
  align-items: center;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  border-radius: 20px;
  padding: 2px;
  gap: 2px;
}

.mode-tab-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  background: transparent;
  border: none;
  color: var(--text-muted);
  font-size: 12px;
  font-weight: 500;
  padding: 4px 12px;
  border-radius: 16px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.mode-tab-btn:hover {
  color: var(--text-main);
}

.mode-tab-btn.active {
  background: var(--card-bg);
  color: var(--accent-cyan);
  font-weight: 600;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

:root[data-theme="light"] .mode-tab-btn.active {
  background: #ffffff;
  color: #0284c7;
}

.diff-status-pill.pill-commit {
  background: rgba(168, 85, 247, 0.15);
  color: #c084fc;
  border: 1px solid rgba(168, 85, 247, 0.3);
}

:root[data-theme="light"] .diff-status-pill.pill-commit {
  background: rgba(124, 58, 237, 0.1);
  color: #7c3aed;
  border-color: rgba(124, 58, 237, 0.25);
}

.git-graph-embedded {
  flex: 1;
  height: calc(100vh - 120px);
  border-radius: 6px;
  overflow: hidden;
  border: 1px solid var(--card-border);
}
</style>
