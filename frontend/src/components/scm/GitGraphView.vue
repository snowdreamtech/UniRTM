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
          :title="t('scm.branch.switch')"
          @change="handleBranchChange"
        >
          <option value="all">{{ t('scm.graph.all_branches') }}</option>
          <option
            v-for="b in branches"
            :key="b.name"
            :value="b.name"
          >
            {{ b.name }}{{ b.current ? ` (${t('scm.graph.current_badge')})` : '' }}
          </option>
        </select>
      </div>

      <!-- 2. File Filter Badge (When filtered by specific file) -->
      <div v-if="graphFilterPath" class="filter-file-badge">
        <span class="codicon codicon-file"></span>
        <span class="file-path-text font-mono" :title="graphFilterPath">{{ getFileName(graphFilterPath) }}</span>
        <button
          class="clear-filter-btn"
          :title="t('scm.graph.clear_filter')"
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
          :placeholder="t('scm.graph.search_placeholder')"
        />
        <button
          v-if="searchQuery"
          class="clear-search-btn"
          :title="t('scm.graph.clear_search')"
          @click="searchQuery = ''"
        >
          <span class="codicon codicon-close"></span>
        </button>
      </div>

      <!-- 4. Refresh Button -->
      <button
        class="toolbar-btn"
        :class="{ loading: isLoadingGraph }"
        :title="t('scm.actions.refresh')"
        @click="refreshGraph"
      >
        <span class="codicon" :class="isLoadingGraph ? 'codicon-sync codicon-modifier-spin' : 'codicon-refresh'"></span>
        <span>{{ t('scm.graph.refresh') }}</span>
      </button>

      <!-- 5. Commit Count Badge -->
      <div class="commit-count-pill">
        <span>{{ t('scm.graph.commits_count', { count: filteredCommits.length }) }}</span>
      </div>
    </div>

    <!-- Main Content Area: Split Table & Commit Details Panel -->
    <div class="graph-main-split">
      <!-- Graph Table Area -->
      <div class="graph-table-wrapper" ref="tableWrapperRef" @scroll="handleScroll">
        <div class="graph-table-header">
          <div class="col-header col-graph" :style="{ width: graphColWidth + 'px' }">{{ t('scm.graph.col_graph') }}</div>
          <div class="col-header col-desc">{{ t('scm.graph.col_desc') }}</div>
          <div class="col-header col-date">{{ t('scm.graph.col_date') }}</div>
          <div class="col-header col-author">{{ t('scm.graph.col_author') }}</div>
          <div class="col-header col-hash">{{ t('scm.graph.col_commit') }}</div>
          <div class="col-header col-actions">{{ t('scm.graph.col_actions') }}</div>
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

              <!-- GPG Signature Badge -->
              <span
                v-if="commit.gpg && commit.gpg.status !== 'none'"
                class="gpg-badge"
                :class="'gpg-' + commit.gpg.status"
                :title="getGpgTooltip(commit.gpg)"
              >
                <span class="codicon" :class="getGpgIcon(commit.gpg.status)"></span>
                <span class="gpg-badge-label">{{ getGpgBadgeLabel(commit.gpg.status) }}</span>
              </span>

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
                :title="t('scm.actions.copy_hash')"
                @click.stop="handleCopyHash(commit.hash)"
              >
                <span class="codicon codicon-copy"></span>
              </button>
            </div>

            <!-- Col 6: Row Action Button -->
            <div class="col-cell col-actions">
              <button
                class="row-action-btn"
                :title="t('scm.modal.title_action')"
                @click.stop="openContextMenu($event, commit)"
              >
                <span class="codicon codicon-ellipsis"></span>
              </button>
            </div>
          </div>

          <!-- Empty State -->
          <div v-if="filteredCommits.length === 0 && !isLoadingGraph" class="graph-empty-state">
            <span class="codicon codicon-inbox empty-icon"></span>
            <span>{{ t('scm.graph.no_commits') }}</span>
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
              :title="t('scm.actions.copy_hash')"
              @click="handleCopyHash(selectedCommit.hash)"
            >
              <span class="codicon codicon-copy"></span>
            </button>
          </div>

          <div class="detail-actions">
            <button
              class="detail-action-btn"
              :title="t('scm.actions.tag')"
              @click="openModal('create-tag', selectedCommit)"
            >
              <span class="codicon codicon-tag"></span>
              <span>{{ t('scm.actions.tag') }}</span>
            </button>
            <button
              class="detail-action-btn"
              :title="t('scm.actions.branch')"
              @click="openModal('create-branch', selectedCommit)"
            >
              <span class="codicon codicon-git-branch"></span>
              <span>{{ t('scm.actions.branch') }}</span>
            </button>
            <button
              class="detail-close-btn"
              :title="t('common.close')"
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
              <span class="meta-label">{{ t('scm.detail.author') }}:</span>
              <span class="meta-value">{{ selectedCommit.author }} &lt;{{ selectedCommit.email }}&gt;</span>
            </div>
            <div class="meta-item">
              <span class="meta-label">{{ t('scm.detail.date') }}:</span>
              <span class="meta-value">{{ formatExactDate(selectedCommit.timestamp) }}</span>
            </div>
            <div v-if="selectedCommit.parents && selectedCommit.parents.length > 0" class="meta-item">
              <span class="meta-label">{{ t('scm.detail.parents') }}:</span>
              <span class="meta-parents">
                <span
                  v-for="p in selectedCommit.parents"
                  :key="p"
                  class="parent-link font-mono"
                  :title="t('scm.detail.parents')"
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

          <!-- GPG Signature Security Card -->
          <div
            v-if="currentGpgInfo"
            class="detail-gpg-card"
            :class="'gpg-card-' + currentGpgInfo.status"
          >
            <div class="gpg-card-header">
              <div class="gpg-card-status">
                <span class="codicon" :class="getGpgIcon(currentGpgInfo.status)"></span>
                <span class="gpg-status-text">{{ getGpgFullLabel(currentGpgInfo.status) }}</span>
                <span class="gpg-status-pill">{{ currentGpgInfo.status.toUpperCase() }}</span>
              </div>
              <span class="gpg-algorithm-badge">GnuPG / OpenPGP</span>
            </div>

            <div v-if="currentGpgInfo.status !== 'none'" class="gpg-card-grid">
              <div v-if="currentGpgInfo.signer" class="gpg-grid-item col-span-2">
                <span class="gpg-grid-label">{{ t('scm.gpg.signer') }}:</span>
                <span class="gpg-grid-value font-mono">{{ currentGpgInfo.signer }}</span>
              </div>
              <div v-if="currentGpgInfo.keyId" class="gpg-grid-item">
                <span class="gpg-grid-label">{{ t('scm.gpg.key_id') }}:</span>
                <div class="gpg-grid-copyable">
                  <span class="gpg-grid-value font-mono">{{ currentGpgInfo.keyId }}</span>
                  <button
                    class="gpg-copy-btn"
                    :title="t('scm.gpg.key_id')"
                    @click="handleCopyHash(currentGpgInfo.keyId)"
                  >
                    <span class="codicon codicon-copy"></span>
                  </button>
                </div>
              </div>
              <div v-if="currentGpgInfo.fingerprint" class="gpg-grid-item col-span-2">
                <span class="gpg-grid-label">{{ t('scm.gpg.fingerprint') }}:</span>
                <div class="gpg-grid-copyable">
                  <span class="gpg-grid-value font-mono fingerprint-text">{{ currentGpgInfo.fingerprint }}</span>
                  <button
                    class="gpg-copy-btn"
                    :title="t('scm.gpg.fingerprint')"
                    @click="handleCopyHash(currentGpgInfo.fingerprint)"
                  >
                    <span class="codicon codicon-copy"></span>
                  </button>
                </div>
              </div>
            </div>
            <div v-else class="gpg-unsigned-hint">
              <span class="codicon codicon-info"></span>
              <span>{{ t('scm.gpg.unsigned_hint') }}</span>
            </div>
          </div>

          <!-- Changed Files Section -->
          <div class="detail-files-section">
            <div class="files-header">
              <span class="files-title">
                {{ t('scm.detail.changed_files') }} ({{ selectedCommitDetail?.files?.length ?? 0 }})
              </span>
              <span class="files-hint text-muted">{{ t('scm.detail.diff_hint') }}</span>
            </div>

            <div v-if="isLoadingDetail" class="files-loading">
              <span class="codicon codicon-loading codicon-modifier-spin"></span>
              <span>{{ t('scm.graph.loading_files') }}</span>
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
                    :title="t('scm.actions.diff')"
                    @click.stop="handleOpenFileDiff(file.path, selectedCommit.hash, selectedCommit.parents?.[0])"
                  >
                    <span class="codicon codicon-diff"></span>
                    <span>{{ t('scm.actions.diff') }}</span>
                  </button>
                  <button
                    class="file-open-btn"
                    :title="t('scm.diff.open_file')"
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
          <span>{{ t('scm.context.checkout') }}</span>
        </div>
        <div class="menu-item" @click="handleContextAction('create-branch')">
          <span class="codicon codicon-git-branch"></span>
          <span>{{ t('scm.context.create_branch') }}</span>
        </div>
        <div class="menu-item" @click="handleContextAction('create-tag')">
          <span class="codicon codicon-tag"></span>
          <span>{{ t('scm.context.create_tag') }}</span>
        </div>
        <div class="menu-separator"></div>
        <div class="menu-item" @click="handleContextAction('cherry-pick')">
          <span class="codicon codicon-git-pull-request"></span>
          <span>{{ t('scm.context.cherry_pick') }}</span>
        </div>
        <div class="menu-item" @click="handleContextAction('revert')">
          <span class="codicon codicon-history"></span>
          <span>{{ t('scm.context.revert') }}</span>
        </div>
        <div class="menu-separator"></div>
        <div class="menu-item danger" @click="handleContextAction('reset')">
          <span class="codicon codicon-discard"></span>
          <span>{{ t('scm.context.reset') }}</span>
        </div>
        <div class="menu-separator"></div>
        <div class="menu-item" @click="handleContextAction('copy-hash')">
          <span class="codicon codicon-copy"></span>
          <span>{{ t('scm.context.copy_hash') }}</span>
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
  type GitGpgInfo,
} from "../../composables/useGit";
import { layoutGitGraph } from "../../utils/gitGraphLayout";
import { useToast } from "../../composables/useToast";
import { t } from "../../i18n";

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
    showToast(t("scm.graph.copied_hash", { hash: hash.slice(0, 8) }), "info");
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
    showToast(t("scm.graph.parent_not_found", { hash: parentHash.slice(0, 7) }), "warning");
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
    showToast(err.message || t("scm.graph.action_failed"), "error");
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
  if (diff < 60) return t("scm.graph.now");
  if (diff < 3600) return t("scm.graph.minutes_ago", { count: Math.floor(diff / 60) });
  if (diff < 86400) return t("scm.graph.hours_ago", { count: Math.floor(diff / 3600) });
  if (diff < 2592000) return t("scm.graph.days_ago", { count: Math.floor(diff / 86400) });
  return new Date(timestamp * 1000).toLocaleDateString();
}

const currentGpgInfo = computed<GitGpgInfo | undefined>(() => {
  return selectedCommitDetail.value?.gpg || selectedCommit.value?.gpg;
});

function getGpgIcon(status?: string): string {
  switch (status) {
    case "verified":
      return "codicon-verified";
    case "unverified":
      return "codicon-shield";
    case "bad":
      return "codicon-warning";
    case "expired":
      return "codicon-history";
    case "revoked":
      return "codicon-error";
    default:
      return "codicon-shield";
  }
}

function getGpgBadgeLabel(status?: string): string {
  switch (status) {
    case "verified":
      return t("scm.gpg.verified");
    case "unverified":
      return t("scm.gpg.unverified");
    case "bad":
      return t("scm.gpg.bad");
    case "expired":
      return t("scm.gpg.expired");
    case "revoked":
      return t("scm.gpg.revoked");
    default:
      return "";
  }
}

function getGpgFullLabel(status?: string): string {
  switch (status) {
    case "verified":
      return t("scm.gpg.verified");
    case "unverified":
      return t("scm.gpg.unverified");
    case "bad":
      return t("scm.gpg.bad");
    case "expired":
      return t("scm.gpg.expired");
    case "revoked":
      return t("scm.gpg.revoked");
    case "none":
    default:
      return t("scm.gpg.none");
  }
}

function getGpgTooltip(gpg?: GitGpgInfo): string {
  if (!gpg || gpg.status === "none") return t("scm.gpg.none");
  const signer = gpg.signer ? ` • ${t("scm.gpg.signer")}: ${gpg.signer}` : "";
  const key = gpg.keyId ? ` • ${t("scm.gpg.key_id")}: ${gpg.keyId}` : "";
  return `${getGpgFullLabel(gpg.status)}${signer}${key}`;
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

/* GPG Badges in Git Graph Rows */
.gpg-badge {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  font-size: 10px;
  font-weight: 600;
  padding: 1px 6px;
  border-radius: 10px;
  flex-shrink: 0;
  letter-spacing: 0.2px;
}

.gpg-badge .codicon {
  font-size: 11px;
}

.gpg-badge.gpg-verified {
  background: rgba(46, 160, 67, 0.15);
  color: #3fb950;
  border: 1px solid rgba(46, 160, 67, 0.4);
}

:root[data-theme="light"] .gpg-badge.gpg-verified {
  background: rgba(46, 160, 67, 0.12);
  color: #1a7f37;
  border-color: rgba(26, 127, 55, 0.35);
}

.gpg-badge.gpg-unverified {
  background: rgba(210, 153, 34, 0.15);
  color: #e3b341;
  border: 1px solid rgba(210, 153, 34, 0.4);
}

:root[data-theme="light"] .gpg-badge.gpg-unverified {
  background: rgba(154, 103, 0, 0.12);
  color: #9a6700;
  border-color: rgba(154, 103, 0, 0.35);
}

.gpg-badge.gpg-bad,
.gpg-badge.gpg-revoked {
  background: rgba(248, 81, 73, 0.15);
  color: #f85149;
  border: 1px solid rgba(248, 81, 73, 0.4);
}

:root[data-theme="light"] .gpg-badge.gpg-bad,
:root[data-theme="light"] .gpg-badge.gpg-revoked {
  background: rgba(207, 34, 46, 0.12);
  color: #cf222e;
  border-color: rgba(207, 34, 46, 0.35);
}

.gpg-badge.gpg-expired {
  background: rgba(219, 109, 40, 0.15);
  color: #db6d28;
  border: 1px solid rgba(219, 109, 40, 0.4);
}

:root[data-theme="light"] .gpg-badge.gpg-expired {
  background: rgba(188, 76, 0, 0.12);
  color: #bc4c00;
  border-color: rgba(188, 76, 0, 0.35);
}

/* GPG Signature Detail Card in Commit Drawer */
.detail-gpg-card {
  border-radius: 6px;
  padding: 10px 14px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  transition: all 0.2s ease;
}

.detail-gpg-card.gpg-card-verified {
  background: linear-gradient(135deg, rgba(46, 160, 67, 0.08) 0%, rgba(13, 25, 18, 0.4) 100%);
  border: 1px solid rgba(46, 160, 67, 0.35);
}

:root[data-theme="light"] .detail-gpg-card.gpg-card-verified {
  background: linear-gradient(135deg, rgba(46, 160, 67, 0.08) 0%, rgba(240, 253, 244, 0.95) 100%);
  border: 1px solid rgba(26, 127, 55, 0.3);
}

.detail-gpg-card.gpg-card-unverified {
  background: rgba(210, 153, 34, 0.08);
  border: 1px solid rgba(210, 153, 34, 0.35);
}

.detail-gpg-card.gpg-card-bad,
.detail-gpg-card.gpg-card-revoked {
  background: rgba(248, 81, 73, 0.08);
  border: 1px solid rgba(248, 81, 73, 0.35);
}

.detail-gpg-card.gpg-card-none {
  background: var(--input-bg);
  border: 1px solid var(--card-border);
}

.gpg-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.gpg-card-status {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  font-weight: 600;
}

.gpg-card-verified .gpg-card-status {
  color: #3fb950;
}

:root[data-theme="light"] .gpg-card-verified .gpg-card-status {
  color: #1a7f37;
}

.gpg-card-unverified .gpg-card-status {
  color: #e3b341;
}

.gpg-card-bad .gpg-card-status,
.gpg-card-revoked .gpg-card-status {
  color: #f85149;
}

.gpg-card-none .gpg-card-status {
  color: var(--text-muted);
}

.gpg-status-pill {
  font-size: 9px;
  font-weight: 700;
  padding: 1px 5px;
  border-radius: 4px;
  background: rgba(128, 128, 128, 0.15);
  letter-spacing: 0.5px;
}

.gpg-algorithm-badge {
  font-size: 10px;
  color: var(--text-muted);
  font-family: monospace;
}

.gpg-card-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 6px 16px;
  font-size: 11px;
}

.gpg-grid-item {
  display: flex;
  align-items: center;
  gap: 6px;
}

.gpg-grid-item.col-span-2 {
  grid-column: span 2;
}

.gpg-grid-label {
  color: var(--text-muted);
  flex-shrink: 0;
}

.gpg-grid-value {
  color: var(--text-main);
  word-break: break-all;
}

.gpg-grid-copyable {
  display: flex;
  align-items: center;
  gap: 6px;
}

.fingerprint-text {
  font-size: 11px;
  letter-spacing: 0.5px;
}

.gpg-copy-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 1px 4px;
  border-radius: 3px;
  display: flex;
  align-items: center;
}

.gpg-copy-btn:hover {
  background: rgba(128, 128, 128, 0.15);
  color: var(--text-main);
}

.gpg-unsigned-hint {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  color: var(--text-muted);
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
