<template>
  <div class="git-graph-container" id="git-graph-view">
    <!-- Top Control & Filter Bar -->
    <div class="graph-toolbar">
      <!-- 1. Branch Selector -->
      <div class="toolbar-item branch-select-group">
        <span class="codicon codicon-git-branch"></span>
        <select
          v-model="selectedBranch"
          class="graph-select"
          title="选择查看的分支"
          @change="handleBranchChange"
        >
          <option value="all">全部分支 (All Branches)</option>
          <option
            v-for="b in branches"
            :key="b.name"
            :value="b.name"
          >
            {{ b.name }}{{ b.current ? " (当前)" : "" }}
          </option>
        </select>
      </div>

      <!-- 2. File Filter Badge (When filtered by specific file) -->
      <div v-if="graphFilterPath" class="filter-file-badge">
        <span class="codicon codicon-file"></span>
        <span class="file-path-text font-mono" :title="graphFilterPath">{{ getFileName(graphFilterPath) }}</span>
        <button
          class="clear-filter-btn"
          title="清除文件过滤，查看全库提交图谱"
          @click="clearGraphFilterPath"
        >
          <span class="codicon codicon-close"></span>
        </button>
      </div>

      <!-- 3. Search Filter Input -->
      <div class="toolbar-item search-group">
        <span class="codicon codicon-search search-icon"></span>
        <input
          v-model="searchQuery"
          type="text"
          class="graph-search-input"
          placeholder="搜索提交信息 / 作者 / Hash..."
        />
        <button
          v-if="searchQuery"
          class="clear-search-btn"
          title="清空搜索"
          @click="searchQuery = ''"
        >
          <span class="codicon codicon-close"></span>
        </button>
      </div>

      <!-- 4. Refresh Button -->
      <button
        class="toolbar-btn"
        :class="{ loading: isLoadingGraph }"
        title="刷新 Git 提交图谱"
        @click="refreshGraph"
      >
        <span class="codicon" :class="isLoadingGraph ? 'codicon-sync codicon-modifier-spin' : 'codicon-refresh'"></span>
        <span>刷新</span>
      </button>

      <!-- 5. Commit Count Badge -->
      <div class="commit-count-pill">
        <span>{{ filteredCommits.length }} 次提交</span>
      </div>
    </div>

    <!-- Main Content Area: Split Table & Commit Details Panel -->
    <div class="graph-main-split">
      <!-- Graph Table Area -->
      <div class="graph-table-wrapper" ref="tableWrapperRef" @scroll="handleScroll">
        <div class="graph-table-header">
          <div class="col-header col-graph" :style="{ width: graphColWidth + 'px' }">图谱</div>
          <div class="col-header col-desc">描述 (Description)</div>
          <div class="col-header col-date">日期</div>
          <div class="col-header col-author">作者</div>
          <div class="col-header col-hash">Commit</div>
          <div class="col-header col-actions">操作</div>
        </div>

        <div class="graph-table-body" :style="{ height: totalBodyHeight + 'px' }">
          <!-- SVG DAG Overlay for Lanes, Nodes, Bezier Curves -->
          <svg
            class="graph-svg-layer"
            :style="{ width: graphColWidth + 'px', height: totalBodyHeight + 'px' }"
          >
            <!-- 1. Connecting Paths (Vertical lines & Bezier curves) -->
            <path
              v-for="p in graphLayout.paths"
              :key="p.id"
              :d="p.d"
              :stroke="p.color"
              stroke-width="2"
              fill="none"
              stroke-linecap="round"
              stroke-linejoin="round"
              class="graph-path-line"
            />

            <!-- 2. Commit Nodes (Dots) -->
            <g v-for="node in graphLayout.nodes" :key="'node:' + node.hash">
              <!-- Outer Ring for HEAD -->
              <circle
                v-if="node.isHead"
                :cx="node.x"
                :cy="node.y"
                r="7"
                fill="none"
                :stroke="node.color"
                stroke-width="1.8"
                class="node-head-ring"
              />
              <!-- Commit Dot -->
              <circle
                :cx="node.x"
                :cy="node.y"
                :r="node.isHead ? 4 : 4.5"
                :fill="node.color"
                stroke="var(--card-bg, #070a12)"
                stroke-width="1.5"
                class="node-circle"
                :class="{ active: selectedCommit?.hash === node.hash }"
              />
            </g>
          </svg>

          <!-- Row Content Rows (HTML rows aligned with SVG rowHeight) -->
          <div
            v-for="(commit, idx) in filteredCommits"
            :key="commit.hash"
            class="graph-row"
            :class="{ selected: selectedCommit?.hash === commit.hash }"
            :style="{ top: idx * ROW_HEIGHT + 'px', height: ROW_HEIGHT + 'px' }"
            @click="handleSelectCommit(commit)"
            @contextmenu.prevent="openContextMenu($event, commit)"
          >
            <!-- Col 1: Empty placeholder matching graph SVG width -->
            <div class="col-cell col-graph" :style="{ width: graphColWidth + 'px' }"></div>

            <!-- Col 2: Ref Badges & Commit Message -->
            <div class="col-cell col-desc">
              <!-- Ref Badges (Branches, Remotes, Tags) -->
              <div v-if="commit.refs && commit.refs.length > 0" class="ref-badges-group">
                <span
                  v-for="(refItem, rIdx) in commit.refs"
                  :key="rIdx"
                  class="ref-badge"
                  :class="'badge-' + refItem.type"
                  :title="refItem.name"
                >
                  <span
                    class="codicon"
                    :class="
                      refItem.type === 'tag'
                        ? 'codicon-tag'
                        : refItem.type === 'remote'
                        ? 'codicon-cloud'
                        : 'codicon-git-branch'
                    "
                  ></span>
                  <span class="ref-name">{{ refItem.name }}</span>
                </span>
              </div>

              <!-- Message Text -->
              <span class="commit-subject" :title="commit.message">{{ commit.message }}</span>
            </div>

            <!-- Col 3: Date -->
            <div class="col-cell col-date" :title="formatExactDate(commit.timestamp)">
              {{ formatRelativeTime(commit.timestamp) }}
            </div>

            <!-- Col 4: Author -->
            <div class="col-cell col-author" :title="`${commit.author} <${commit.email}>`">
              <span class="author-avatar">{{ getInitials(commit.author) }}</span>
              <span class="author-name">{{ commit.author }}</span>
            </div>

            <!-- Col 5: Short Hash -->
            <div class="col-cell col-hash">
              <span class="hash-text font-mono">{{ commit.shortHash }}</span>
              <button
                class="copy-hash-btn"
                title="复制 Commit SHA"
                @click.stop="handleCopyHash(commit.hash)"
              >
                <span class="codicon codicon-copy"></span>
              </button>
            </div>

            <!-- Col 6: Row Action Button -->
            <div class="col-cell col-actions">
              <button
                class="row-action-btn"
                title="高阶 Git 操作"
                @click.stop="openContextMenu($event, commit)"
              >
                <span class="codicon codicon-ellipsis"></span>
              </button>
            </div>
          </div>

          <!-- Empty State -->
          <div v-if="filteredCommits.length === 0 && !isLoadingGraph" class="graph-empty-state">
            <span class="codicon codicon-inbox empty-icon"></span>
            <span>没有匹配的提交记录</span>
          </div>
        </div>
      </div>

      <!-- Bottom / Split Commit Details Panel -->
      <div v-if="selectedCommit" class="commit-detail-panel">
        <div class="detail-header">
          <div class="detail-title-group">
            <span class="codicon codicon-git-commit text-cyan"></span>
            <span class="detail-hash font-mono">{{ selectedCommit.hash }}</span>
            <button
              class="copy-detail-hash-btn"
              title="复制完整 SHA"
              @click="handleCopyHash(selectedCommit.hash)"
            >
              <span class="codicon codicon-copy"></span>
            </button>
          </div>

          <div class="detail-actions">
            <button
              class="detail-action-btn"
              title="为此提交创建标签"
              @click="openModal('create-tag', selectedCommit)"
            >
              <span class="codicon codicon-tag"></span>
              <span>打标签</span>
            </button>
            <button
              class="detail-action-btn"
              title="从此提交检出并新建分支"
              @click="openModal('create-branch', selectedCommit)"
            >
              <span class="codicon codicon-git-branch"></span>
              <span>新建分支</span>
            </button>
            <button
              class="detail-close-btn"
              title="关闭详情"
              @click="selectedCommit = null"
            >
              <span class="codicon codicon-close"></span>
            </button>
          </div>
        </div>

        <div class="detail-content">
          <!-- Metadata Meta Row -->
          <div class="meta-row">
            <div class="meta-item">
              <span class="meta-label">作者:</span>
              <span class="meta-value">{{ selectedCommit.author }} &lt;{{ selectedCommit.email }}&gt;</span>
            </div>
            <div class="meta-item">
              <span class="meta-label">日期:</span>
              <span class="meta-value">{{ formatExactDate(selectedCommit.timestamp) }}</span>
            </div>
            <div v-if="selectedCommit.parents && selectedCommit.parents.length > 0" class="meta-item">
              <span class="meta-label">父提交:</span>
              <span class="meta-parents">
                <span
                  v-for="p in selectedCommit.parents"
                  :key="p"
                  class="parent-link font-mono"
                  title="跳转至父提交"
                  @click="jumpToCommit(p)"
                >
                  {{ p.slice(0, 7) }}
                </span>
              </span>
            </div>
          </div>

          <!-- Commit Message Body -->
          <div class="detail-message-box">
            <pre class="detail-message-text">{{ selectedCommitDetail?.message || selectedCommit.message }}</pre>
          </div>

          <!-- Changed Files Section -->
          <div class="detail-files-section">
            <div class="files-header">
              <span class="files-title">
                变更文件 ({{ selectedCommitDetail?.files?.length ?? 0 }})
              </span>
              <span class="files-hint text-muted">点击任意文件立即打开 Monaco 差异比对</span>
            </div>

            <div v-if="isLoadingDetail" class="files-loading">
              <span class="codicon codicon-loading codicon-modifier-spin"></span>
              <span>正在获取变更文件...</span>
            </div>

            <div v-else class="files-list">
              <div
                v-for="file in selectedCommitDetail?.files || []"
                :key="file.path"
                class="file-item-row"
                @click="handleOpenFileDiff(file.path, selectedCommit.hash, selectedCommit.parents?.[0])"
              >
                <TechIcon :id="getFileName(file.path)" :size="16" />
                <span class="file-status-badge" :class="'badge-' + file.status">{{ file.status }}</span>
                <span class="file-name font-mono">{{ file.path }}</span>

                <div class="file-stats">
                  <span v-if="file.additions > 0" class="stat-add">+{{ file.additions }}</span>
                  <span v-if="file.deletions > 0" class="stat-del">-{{ file.deletions }}</span>
                </div>

                <div class="file-item-actions">
                  <button
                    class="file-diff-btn"
                    title="在工作区 Monaco 编辑器中比对差异"
                    @click.stop="handleOpenFileDiff(file.path, selectedCommit.hash, selectedCommit.parents?.[0])"
                  >
                    <span class="codicon codicon-diff"></span>
                    <span>比对差异</span>
                  </button>
                  <button
                    class="file-open-btn"
                    title="在代码编辑器中打开此文件"
                    @click.stop="handleOpenFileInEditor(file.path)"
                  >
                    <span class="codicon codicon-go-to-file"></span>
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Right-click Context Menu (Teleported) -->
    <Teleport to="body">
      <div
        v-if="contextMenuVisible"
        class="graph-context-menu"
        :style="{ left: contextMenuPos.x + 'px', top: contextMenuPos.y + 'px' }"
        @click.stop
      >
        <div class="menu-item" @click="handleContextAction('checkout')">
          <span class="codicon codicon-check"></span>
          <span>检出此提交 (Checkout)</span>
        </div>
        <div class="menu-item" @click="handleContextAction('create-branch')">
          <span class="codicon codicon-git-branch"></span>
          <span>从此提交新建分支 (Create Branch)...</span>
        </div>
        <div class="menu-item" @click="handleContextAction('create-tag')">
          <span class="codicon codicon-tag"></span>
          <span>为此提交打标签 (Create Tag)...</span>
        </div>
        <div class="menu-separator"></div>
        <div class="menu-item" @click="handleContextAction('cherry-pick')">
          <span class="codicon codicon-git-pull-request"></span>
          <span>拣选此提交 (Cherry-Pick)</span>
        </div>
        <div class="menu-item" @click="handleContextAction('revert')">
          <span class="codicon codicon-history"></span>
          <span>撤销此提交 (Revert)</span>
        </div>
        <div class="menu-separator"></div>
        <div class="menu-item danger" @click="handleContextAction('reset')">
          <span class="codicon codicon-discard"></span>
          <span>重置分支到此提交 (Reset Branch)...</span>
        </div>
        <div class="menu-separator"></div>
        <div class="menu-item" @click="handleContextAction('copy-hash')">
          <span class="codicon codicon-copy"></span>
          <span>复制 Commit SHA</span>
        </div>
      </div>
    </Teleport>

    <!-- High-order Git Action Modal Dialog -->
    <GitActionModal
      :visible="modalVisible"
      :action-type="activeModalType"
      :target-commit="modalTargetCommit"
      @cancel="modalVisible = false"
      @confirm="handleModalConfirm"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from "vue";
import TechIcon from "../common/TechIcon.vue";
import GitActionModal, { type GitModalActionType } from "./GitActionModal.vue";
import {
  useGit,
  type GitGraphCommit,
} from "../../composables/useGit";
import { layoutGitGraph } from "../../utils/gitGraphLayout";
import { useToast } from "../../composables/useToast";

const ROW_HEIGHT = 34;

const {
  branches,
  graphCommits,
  selectedCommit,
  selectedCommitDetail,
  graphFilterBranch,
  graphFilterPath,
  isLoadingGraph,
  fetchBranches,
  fetchGraph,
  fetchCommitDetail,
  clearGraphFilterPath,
  checkout,
  createTag,
  createBranchFromCommit,
  cherryPick,
  revertCommit,
  resetToCommit,
} = useGit();

const { showToast } = useToast();

const selectedBranch = ref(graphFilterBranch.value || "all");
const searchQuery = ref("");
const tableWrapperRef = ref<HTMLElement | null>(null);
const isLoadingDetail = ref(false);

// Context Menu State
const contextMenuVisible = ref(false);
const contextMenuPos = ref({ x: 0, y: 0 });
const contextCommit = ref<GitGraphCommit | null>(null);

// Modal Dialog State
const modalVisible = ref(false);
const activeModalType = ref<GitModalActionType | null>(null);
const modalTargetCommit = ref<GitGraphCommit | null>(null);

const emit = defineEmits<{
  (
    e: "open-commit-diff",
    payload: { path: string; hash: string; parentHash?: string }
  ): void;
  (e: "open-file", path: string): void;
}>();

// Filtered Commits by live search
const filteredCommits = computed(() => {
  const query = searchQuery.value.trim().toLowerCase();
  if (!query) return graphCommits.value;
  return graphCommits.value.filter((c) => {
    return (
      c.message.toLowerCase().includes(query) ||
      c.author.toLowerCase().includes(query) ||
      c.hash.toLowerCase().includes(query) ||
      c.shortHash.toLowerCase().includes(query) ||
      c.refs?.some((r) => r.name.toLowerCase().includes(query))
    );
  });
});

// Computed SVG Graph Layout
const graphLayout = computed(() => {
  return layoutGitGraph(filteredCommits.value, {
    rowHeight: ROW_HEIGHT,
    laneWidth: 18,
    startX: 16,
  });
});

const graphColWidth = computed(() => {
  return Math.max(68, graphLayout.value.width);
});

const totalBodyHeight = computed(() => {
  return Math.max(ROW_HEIGHT, filteredCommits.value.length * ROW_HEIGHT);
});

onMounted(async () => {
  await fetchBranches();
  await refreshGraph();
  window.addEventListener("click", closeContextMenu);
});

onBeforeUnmount(() => {
  window.removeEventListener("click", closeContextMenu);
});

async function refreshGraph() {
  await fetchGraph({
    branch: selectedBranch.value,
    path: graphFilterPath.value,
  });
}

function handleBranchChange() {
  graphFilterBranch.value = selectedBranch.value;
  refreshGraph();
}

async function handleSelectCommit(commit: GitGraphCommit) {
  selectedCommit.value = commit;
  isLoadingDetail.value = true;
  try {
    await fetchCommitDetail(commit.hash);
  } finally {
    isLoadingDetail.value = false;
  }
}

function handleOpenFileDiff(path: string, hash: string, parentHash?: string) {
  emit("open-commit-diff", { path, hash, parentHash });
}

function handleOpenFileInEditor(path: string) {
  emit("open-file", path);
}

function handleCopyHash(hash: string) {
  navigator.clipboard.writeText(hash).then(() => {
    showToast(`已复制 Commit SHA: ${hash.slice(0, 8)}`, "info");
  });
}

function jumpToCommit(parentHash: string) {
  const target = graphCommits.value.find(
    (c) => c.hash.startsWith(parentHash) || parentHash.startsWith(c.hash)
  );
  if (target) {
    handleSelectCommit(target);
    const idx = filteredCommits.value.findIndex((c) => c.hash === target.hash);
    if (idx !== -1 && tableWrapperRef.value) {
      tableWrapperRef.value.scrollTop = Math.max(0, idx * ROW_HEIGHT - 100);
    }
  } else {
    showToast(`未在当前加载的提交列表中找到父提交: ${parentHash.slice(0, 7)}`, "warning");
  }
}

function handleScroll() {
  closeContextMenu();
}

// Context Menu Handling
function openContextMenu(event: MouseEvent, commit: GitGraphCommit) {
  contextCommit.value = commit;
  contextMenuPos.value = {
    x: Math.min(event.clientX, window.innerWidth - 240),
    y: Math.min(event.clientY, window.innerHeight - 300),
  };
  contextMenuVisible.value = true;
}

function closeContextMenu() {
  contextMenuVisible.value = false;
}

async function handleContextAction(
  action: GitModalActionType | "copy-hash"
) {
  if (!contextCommit.value) return;
  const commit = contextCommit.value;
  closeContextMenu();

  if (action === "copy-hash") {
    handleCopyHash(commit.hash);
    return;
  }

  openModal(action, commit);
}

function openModal(action: GitModalActionType, commit: GitGraphCommit) {
  activeModalType.value = action;
  modalTargetCommit.value = commit;
  modalVisible.value = true;
}

async function handleModalConfirm(payload: {
  action: GitModalActionType;
  commit: GitGraphCommit;
  data?: any;
}) {
  modalVisible.value = false;
  const { action, commit, data } = payload;

  try {
    if (action === "create-tag") {
      await createTag(data.name, commit.hash, data.message);
    } else if (action === "create-branch") {
      await createBranchFromCommit(data.name, commit.hash, data.checkout);
    } else if (action === "reset") {
      await resetToCommit(commit.hash, data.mode);
    } else if (action === "cherry-pick") {
      await cherryPick(commit.hash);
    } else if (action === "revert") {
      await revertCommit(commit.hash);
    } else if (action === "checkout") {
      await checkout(commit.hash);
    }
  } catch (err: any) {
    showToast(err.message || "Git 操作失败", "error");
  }
}

function getFileName(filePath: string): string {
  if (!filePath) return "";
  const parts = filePath.split("/");
  return parts[parts.length - 1] || filePath;
}

function getInitials(name: string): string {
  if (!name) return "U";
  const parts = name.trim().split(" ");
  if (parts.length > 1) {
    return (parts[0][0] + parts[1][0]).toUpperCase();
  }
  return name.slice(0, 2).toUpperCase();
}

function formatExactDate(timestamp: number): string {
  if (!timestamp) return "";
  return new Date(timestamp * 1000).toLocaleString();
}

function formatRelativeTime(timestamp: number): string {
  if (!timestamp) return "";
  const now = Math.floor(Date.now() / 1000);
  const diff = now - timestamp;
  if (diff < 60) return "刚刚";
  if (diff < 3600) return `${Math.floor(diff / 60)} 分钟前`;
  if (diff < 86400) return `${Math.floor(diff / 3600)} 小时前`;
  if (diff < 2592000) return `${Math.floor(diff / 86400)} 天前`;
  return new Date(timestamp * 1000).toLocaleDateString();
}
</script>

<style scoped>
.git-graph-container {
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 100%;
  background: var(--bg-color, #070a12);
  color: var(--text-main, #f8fafc);
  overflow: hidden;
  user-select: none;
}

/* Toolbar */
.graph-toolbar {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 14px;
  background: var(--card-bg, #121a2b);
  border-bottom: 1px solid var(--card-border);
  flex-shrink: 0;
}

.toolbar-item {
  display: flex;
  align-items: center;
  gap: 6px;
}

.branch-select-group {
  color: var(--accent-cyan);
}

.graph-select {
  background: var(--input-bg);
  color: var(--text-main);
  border: 1px solid var(--card-border);
  border-radius: 4px;
  padding: 4px 8px;
  font-size: 12px;
  outline: none;
  cursor: pointer;
}

.filter-file-badge {
  display: flex;
  align-items: center;
  gap: 6px;
  background: rgba(0, 229, 255, 0.12);
  border: 1px solid var(--accent-cyan);
  color: var(--accent-cyan);
  padding: 3px 8px;
  border-radius: 12px;
  font-size: 11px;
}

:root[data-theme="light"] .filter-file-badge {
  background: rgba(2, 132, 199, 0.1);
  color: #0284c7;
  border-color: #0284c7;
}

.file-path-text {
  max-width: 160px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.clear-filter-btn {
  background: transparent;
  border: none;
  color: inherit;
  cursor: pointer;
  padding: 0;
  display: flex;
  align-items: center;
}

.search-group {
  position: relative;
  flex: 1;
  max-width: 320px;
}

.search-icon {
  position: absolute;
  left: 8px;
  color: var(--text-muted);
  pointer-events: none;
}

.graph-search-input {
  width: 100%;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  color: var(--text-main);
  border-radius: 4px;
  padding: 4px 26px 4px 28px;
  font-size: 12px;
  outline: none;
}

.graph-search-input:focus {
  border-color: var(--accent-cyan);
}

.clear-search-btn {
  position: absolute;
  right: 6px;
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 2px;
  display: flex;
  align-items: center;
}

.toolbar-btn {
  display: flex;
  align-items: center;
  gap: 5px;
  background: var(--section-bg);
  border: 1px solid var(--card-border);
  color: var(--text-main);
  padding: 4px 10px;
  border-radius: 4px;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.toolbar-btn:hover {
  background: rgba(128, 128, 128, 0.15);
}

.commit-count-pill {
  margin-left: auto;
  font-size: 11px;
  color: var(--text-muted);
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  padding: 3px 8px;
  border-radius: 10px;
}

/* Split Main Container */
.graph-main-split {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  position: relative;
}

/* Table Area */
.graph-table-wrapper {
  flex: 1;
  overflow: auto;
  position: relative;
}

.graph-table-header {
  position: sticky;
  top: 0;
  z-index: 20;
  display: flex;
  align-items: center;
  background: var(--card-bg);
  border-bottom: 1px solid var(--card-border);
  font-size: 11px;
  font-weight: 600;
  color: var(--text-muted);
  height: 28px;
}

.col-header {
  padding: 0 10px;
  box-sizing: border-box;
}

.graph-table-body {
  position: relative;
  width: 100%;
}

.graph-svg-layer {
  position: absolute;
  top: 0;
  left: 0;
  z-index: 5;
  pointer-events: none;
}

.node-head-ring {
  animation: pulseHead 2s infinite ease-in-out;
}

@keyframes pulseHead {
  0%, 100% {
    stroke-opacity: 0.9;
    r: 7;
  }
  50% {
    stroke-opacity: 0.4;
    r: 8.5;
  }
}

.node-circle {
  transition: r 0.15s ease;
}

.node-circle.active {
  stroke: #ffffff;
  stroke-width: 2px;
}

/* Row Item */
.graph-row {
  position: absolute;
  left: 0;
  right: 0;
  display: flex;
  align-items: center;
  border-bottom: 1px solid rgba(128, 128, 128, 0.08);
  font-size: 12px;
  cursor: pointer;
  transition: background-color 0.1s ease;
  z-index: 10;
}

.graph-row:hover {
  background-color: var(--item-hover-bg, rgba(128, 128, 128, 0.08));
}

.graph-row.selected {
  background-color: rgba(0, 229, 255, 0.1);
  border-bottom-color: rgba(0, 229, 255, 0.3);
}

:root[data-theme="light"] .graph-row.selected {
  background-color: rgba(2, 132, 199, 0.1);
  border-bottom-color: rgba(2, 132, 199, 0.3);
}

.col-cell {
  padding: 0 10px;
  box-sizing: border-box;
  display: flex;
  align-items: center;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.col-graph {
  flex-shrink: 0;
}

.col-desc {
  flex: 1;
  min-width: 240px;
  gap: 8px;
}

.ref-badges-group {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-shrink: 0;
}

.ref-badge {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  font-size: 10px;
  font-weight: 600;
  padding: 1px 6px;
  border-radius: 10px;
}

.badge-branch {
  background: rgba(16, 185, 129, 0.15);
  color: var(--success, #10b981);
  border: 1px solid rgba(16, 185, 129, 0.3);
}

:root[data-theme="light"] .badge-branch {
  color: #059669;
}

.badge-remote {
  background: rgba(168, 85, 247, 0.15);
  color: #a855f7;
  border: 1px solid rgba(168, 85, 247, 0.3);
}

:root[data-theme="light"] .badge-remote {
  color: #7c3aed;
}

.badge-tag {
  background: rgba(245, 158, 11, 0.15);
  color: var(--warning, #f59e0b);
  border: 1px solid rgba(245, 158, 11, 0.3);
}

:root[data-theme="light"] .badge-tag {
  color: #b45309;
}

.commit-subject {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--text-main);
}

.col-date {
  width: 110px;
  flex-shrink: 0;
  font-size: 11px;
  color: var(--text-muted);
}

.col-author {
  width: 130px;
  flex-shrink: 0;
  gap: 6px;
  color: var(--text-muted);
}

.author-avatar {
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background: var(--section-bg);
  border: 1px solid var(--card-border);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 9px;
  font-weight: 700;
  color: var(--accent-cyan);
}

.col-hash {
  width: 85px;
  flex-shrink: 0;
  gap: 4px;
}

.hash-text {
  font-size: 11px;
  color: var(--accent-cyan);
}

.copy-hash-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 1px;
  display: none;
}

.graph-row:hover .copy-hash-btn {
  display: inline-flex;
}

.col-actions {
  width: 44px;
  flex-shrink: 0;
  justify-content: center;
}

.row-action-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 2px 4px;
  border-radius: 3px;
  display: flex;
  align-items: center;
}

.row-action-btn:hover {
  background: rgba(128, 128, 128, 0.15);
  color: var(--text-main);
}

/* Empty State */
.graph-empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  gap: 12px;
  color: var(--text-muted);
}

.empty-icon {
  font-size: 32px;
}

/* Commit Detail Panel (Bottom Drawer) */
.commit-detail-panel {
  max-height: 45%;
  min-height: 220px;
  background: var(--card-bg);
  border-top: 1px solid var(--card-border);
  display: flex;
  flex-direction: column;
  box-shadow: 0 -4px 16px rgba(0, 0, 0, 0.15);
  z-index: 30;
}

.detail-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 14px;
  background: var(--section-bg);
  border-bottom: 1px solid var(--card-border);
}

.detail-title-group {
  display: flex;
  align-items: center;
  gap: 8px;
}

.text-cyan {
  color: var(--accent-cyan);
}

.detail-hash {
  font-size: 12px;
  font-weight: 600;
  color: var(--accent-cyan);
}

.copy-detail-hash-btn,
.detail-close-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 2px 4px;
  border-radius: 3px;
  display: flex;
  align-items: center;
}

.copy-detail-hash-btn:hover,
.detail-close-btn:hover {
  background: rgba(128, 128, 128, 0.15);
  color: var(--text-main);
}

.detail-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.detail-action-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  color: var(--text-main);
  font-size: 11px;
  padding: 3px 8px;
  border-radius: 3px;
  cursor: pointer;
}

.detail-action-btn:hover {
  background: rgba(128, 128, 128, 0.15);
}

.detail-content {
  flex: 1;
  overflow-y: auto;
  padding: 12px 14px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.meta-row {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 16px;
  font-size: 11px;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 6px;
}

.meta-label {
  color: var(--text-muted);
}

.meta-value {
  color: var(--text-main);
}

.meta-parents {
  display: flex;
  align-items: center;
  gap: 6px;
}

.parent-link {
  color: var(--accent-cyan);
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  padding: 1px 5px;
  border-radius: 3px;
  cursor: pointer;
}

.parent-link:hover {
  text-decoration: underline;
}

.detail-message-box {
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  border-radius: 4px;
  padding: 8px 12px;
}

.detail-message-text {
  margin: 0;
  font-family: inherit;
  font-size: 12px;
  line-height: 1.5;
  white-space: pre-wrap;
  color: var(--text-main);
}

.detail-files-section {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.files-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.files-title {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-main);
}

.files-hint {
  font-size: 11px;
}

.files-loading {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: var(--text-muted);
  padding: 12px 0;
}

.files-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
  border: 1px solid var(--card-border);
  border-radius: 4px;
  background: var(--input-bg);
  overflow: hidden;
}

.file-item-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 10px;
  border-bottom: 1px solid rgba(128, 128, 128, 0.08);
  font-size: 12px;
  cursor: pointer;
}

.file-item-row:hover {
  background: var(--item-hover-bg, rgba(128, 128, 128, 0.08));
}

.file-status-badge {
  font-size: 10px;
  font-weight: 700;
  min-width: 14px;
  text-align: center;
}

.badge-M {
  color: var(--warning, #f59e0b);
}

.badge-A,
.badge-U {
  color: var(--success, #10b981);
}

.badge-D {
  color: var(--error, #ef4444);
}

:root[data-theme="light"] .badge-M {
  color: #b45309;
}

:root[data-theme="light"] .badge-A,
:root[data-theme="light"] .badge-U {
  color: #059669;
}

:root[data-theme="light"] .badge-D {
  color: #dc2626;
}

.file-name {
  flex: 1;
  color: var(--text-main);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-stats {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  font-family: monospace;
}

.stat-add {
  color: var(--success, #10b981);
}

.stat-del {
  color: var(--error, #ef4444);
}

:root[data-theme="light"] .stat-add {
  color: #059669;
}

:root[data-theme="light"] .stat-del {
  color: #dc2626;
}

.file-item-actions {
  display: none;
  align-items: center;
  gap: 4px;
}

.file-item-row:hover .file-item-actions {
  display: flex;
}

.file-diff-btn,
.file-open-btn {
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  color: var(--text-main);
  font-size: 11px;
  padding: 2px 6px;
  border-radius: 3px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 3px;
}

.file-diff-btn:hover,
.file-open-btn:hover {
  background: rgba(128, 128, 128, 0.15);
}

/* Context Menu */
.graph-context-menu {
  position: fixed;
  background: var(--card-bg, #182234);
  border: 1px solid var(--card-border, rgba(0, 229, 255, 0.2));
  border-radius: 6px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4);
  padding: 4px 0;
  min-width: 210px;
  z-index: 2000;
  display: flex;
  flex-direction: column;
}

.menu-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 14px;
  font-size: 12px;
  color: var(--text-main);
  cursor: pointer;
  transition: background-color 0.1s ease;
}

.menu-item:hover {
  background: rgba(0, 229, 255, 0.12);
  color: var(--accent-cyan);
}

:root[data-theme="light"] .menu-item:hover {
  background: rgba(2, 132, 199, 0.1);
  color: #0284c7;
}

.menu-item.danger:hover {
  background: rgba(239, 68, 68, 0.12);
  color: var(--error, #ef4444);
}

.menu-separator {
  height: 1px;
  background: var(--card-border);
  margin: 4px 0;
}
</style>
