<template>
  <div class="scm-view-container" id="view-scm">
    <!-- Top Diff Toolbar (When viewing a specific file diff) -->
    <div v-if="activeDiffFile" class="diff-header-bar">
      <div class="diff-title-group">
        <span class="codicon codicon-diff"></span>
        <span class="diff-file-path font-mono">{{ activeDiffFile }}</span>
        <span class="diff-status-pill" :class="isStaged ? 'pill-staged' : 'pill-unstaged'">
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

        <!-- Open File in Editor -->
        <button
          class="diff-btn"
          title="在代码编辑器中打开"
          @click="handleOpenFileInEditor"
        >
          <span class="codicon codicon-go-to-file"></span>
          <span>打开文件</span>
        </button>

        <!-- Close Diff / Back to Overview -->
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
              <p class="repo-subtitle">统一版本控制中心 · 分支与变更比对管理</p>
            </div>
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
        </div>
        <div class="commits-list">
          <div v-for="c in commits" :key="c.hash" class="commit-item">
            <span class="commit-hash font-mono">{{ c.shortHash }}</span>
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
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, watch, nextTick } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useGit } from "../composables/useGit";
import { useEditorTabs } from "../composables/useEditorTabs";
import { loadMonaco, getLanguageForFile } from "../services/editor/monacoLoader";
import TechIcon from "../components/common/TechIcon.vue";

const route = useRoute();
const router = useRouter();

const {
  status,
  currentBranch,
  totalChanges,
  commits,
  fetchStatus,
  fetchDiff,
  fetchLog,
  stage,
  unstage,
  discard,
} = useGit();

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
  const diff = await fetchDiff(filePath, staged);
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
      theme: "vs-dark",
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
  router.replace({ name: "scm" });
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
}

watch(
  () => [route.query.path, route.query.staged],
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
  background: var(--bg-canvas, #1e1e1e);
  color: var(--text-main, #cccccc);
}

/* Diff Header Bar */
.diff-header-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 16px;
  background: var(--bg-header, #252526);
  border-bottom: 1px solid var(--border-color, rgba(255, 255, 255, 0.08));
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
  background: rgba(115, 201, 145, 0.2);
  color: #73c991;
  border: 1px solid rgba(115, 201, 145, 0.4);
}

.pill-unstaged {
  background: rgba(226, 192, 141, 0.2);
  color: #e2c08d;
  border: 1px solid rgba(226, 192, 141, 0.4);
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
  border: 1px solid var(--border-color, rgba(255, 255, 255, 0.15));
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-main, #cccccc);
  cursor: pointer;
  transition: all 0.15s ease;
}

.diff-btn:hover {
  background: rgba(255, 255, 255, 0.12);
  color: #ffffff;
}

.diff-btn.primary {
  background: var(--accent-color, #007acc);
  color: #ffffff;
  border-color: var(--accent-color, #007acc);
}

.diff-btn.primary:hover {
  background: #0062a3;
}

.diff-btn.danger {
  background: rgba(235, 87, 87, 0.15);
  color: #eb5757;
  border-color: rgba(235, 87, 87, 0.4);
}

.diff-btn.danger:hover {
  background: rgba(235, 87, 87, 0.25);
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
  border-bottom: 1px solid var(--border-color, rgba(255, 255, 255, 0.08));
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
  color: var(--accent-color, #007acc);
}

.repo-title {
  font-size: 20px;
  font-weight: 600;
  margin: 0;
  color: #ffffff;
}

.repo-subtitle {
  font-size: 13px;
  color: var(--text-muted, #858585);
  margin: 4px 0 0 0;
}

.header-branch-pill {
  display: flex;
  align-items: center;
  gap: 8px;
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(255, 255, 255, 0.12);
  padding: 6px 14px;
  border-radius: 20px;
  font-weight: 500;
}

.branch-sync-badge {
  font-size: 11px;
  background: var(--accent-color, #007acc);
  color: #fff;
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
  background: var(--bg-card, rgba(255, 255, 255, 0.03));
  border: 1px solid var(--border-color, rgba(255, 255, 255, 0.08));
  border-radius: 6px;
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
  background: rgba(0, 122, 204, 0.15);
  color: #007acc;
}

.stat-icon-wrapper.green {
  background: rgba(115, 201, 145, 0.15);
  color: #73c991;
}

.stat-icon-wrapper.amber {
  background: rgba(226, 192, 141, 0.15);
  color: #e2c08d;
}

.stat-icon-wrapper.purple {
  background: rgba(179, 136, 255, 0.15);
  color: #b388ff;
}

.stat-details {
  display: flex;
  flex-direction: column;
}

.stat-value {
  font-size: 16px;
  font-weight: 600;
  color: #ffffff;
}

.stat-label {
  font-size: 12px;
  color: var(--text-muted, #858585);
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
  border: 1px solid var(--border-color, rgba(255, 255, 255, 0.15));
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-main, #cccccc);
  cursor: pointer;
  transition: all 0.15s ease;
}

.action-pill:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.12);
  color: #ffffff;
}

.action-pill.primary {
  background: var(--accent-color, #007acc);
  color: #ffffff;
  border-color: var(--accent-color, #007acc);
}

.action-pill.primary:hover:not(:disabled) {
  background: #0062a3;
}

.action-pill:disabled {
  opacity: 0.4;
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
  background: var(--bg-card, rgba(255, 255, 255, 0.02));
  border: 1px solid var(--border-color, rgba(255, 255, 255, 0.08));
  border-radius: 6px;
  overflow: hidden;
}

.box-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px;
  background: rgba(255, 255, 255, 0.04);
  border-bottom: 1px solid var(--border-color, rgba(255, 255, 255, 0.06));
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
}

.box-badge.green {
  background: rgba(115, 201, 145, 0.2);
  color: #73c991;
}

.box-badge.amber {
  background: rgba(226, 192, 141, 0.2);
  color: #e2c08d;
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
  border-bottom: 1px solid rgba(255, 255, 255, 0.03);
  transition: background 0.15s ease;
}

.box-file-item:hover {
  background: rgba(255, 255, 255, 0.06);
}

.file-name {
  font-weight: 500;
  color: #ffffff;
}

.file-dir {
  flex: 1;
  font-size: 11px;
  color: var(--text-muted, #777777);
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
  color: #e2c08d;
}

.status-tag.tag-A,
.status-tag.tag-\? {
  color: #73c991;
}

.status-tag.tag-D {
  color: #c74e39;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 36px 16px;
  gap: 8px;
  color: var(--text-muted, #666666);
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
  background: var(--bg-card, rgba(255, 255, 255, 0.02));
  border: 1px solid var(--border-color, rgba(255, 255, 255, 0.08));
  border-radius: 6px;
  overflow: hidden;
  margin-bottom: 24px;
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
  border-bottom: 1px solid rgba(255, 255, 255, 0.03);
  font-size: 12px;
}

.commit-item:hover {
  background: rgba(255, 255, 255, 0.04);
}

.commit-hash {
  color: var(--accent-color, #007acc);
  font-weight: 600;
  min-width: 65px;
}

.commit-message {
  flex: 1;
  color: #ffffff;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.commit-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  color: var(--text-muted, #777777);
  font-size: 11px;
}

.commit-author {
  color: var(--text-muted, #999999);
}
</style>
