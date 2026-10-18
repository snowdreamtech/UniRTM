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
        <span>{{ t('common.refresh') }}</span>
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
            :style="{ top: getRowTop(idx) + 'px', height: ROW_HEIGHT + 'px' }"
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

          <!-- Inline Expanded Commit Detail Panel -->
          <div
            v-if="selectedCommit && selectedCommitIndex !== -1"
            class="inline-commit-detail-panel"
            :style="{
              top: ((selectedCommitIndex + 1) * ROW_HEIGHT) + 'px',
              left: graphColWidth + 'px',
              height: EXPANDED_PANEL_HEIGHT + 'px',
              width: `calc(100% - ${graphColWidth}px)`
            }"
            @click.stop
          >
            <!-- Inline Action Bar (Top) -->
            <div class="inline-detail-actions-bar">
              <div class="inline-detail-actions-left">
                <button
                  class="inline-tool-btn"
                  :title="t('scm.actions.tag')"
                  @click="openModal('create-tag', selectedCommit)"
                >
                  <span class="codicon codicon-tag"></span>
                  <span>{{ t('scm.actions.tag') }}</span>
                </button>
                <button
                  class="inline-tool-btn"
                  :title="t('scm.actions.branch')"
                  @click="openModal('create-branch', selectedCommit)"
                >
                  <span class="codicon codicon-git-branch"></span>
                  <span>{{ t('scm.actions.branch') }}</span>
                </button>
              </div>

              <div class="inline-detail-actions-right">
                <!-- Tree vs List View Toggle -->
                <button
                  class="inline-icon-btn"
                  :class="{ active: filesViewMode === 'tree' }"
                  :title="t('scm.detail.view_tree')"
                  @click="filesViewMode = 'tree'"
                >
                  <span class="codicon codicon-list-tree"></span>
                </button>
                <button
                  class="inline-icon-btn"
                  :class="{ active: filesViewMode === 'list' }"
                  :title="t('scm.detail.view_list')"
                  @click="filesViewMode = 'list'"
                >
                  <span class="codicon codicon-list-flat"></span>
                </button>
                <div class="action-divider"></div>
                <!-- Close Button -->
                <button
                  class="inline-close-btn"
                  :title="t('common.close')"
                  @click="selectedCommit = null"
                >
                  <span class="codicon codicon-close"></span>
                </button>
              </div>
            </div>

            <!-- Inline Split Columns (Meta on Left, Files on Right) -->
            <div class="inline-detail-columns">
              <!-- Left Column: Commit Metadata & GPG & Message -->
              <div class="inline-meta-column">
                <div class="meta-field-group">
                  <div class="meta-field">
                    <span class="meta-field-label">{{ t('scm.detail.commit') }}:</span>
                    <div class="meta-field-val hash-val font-mono">
                      <span>{{ selectedCommit.hash }}</span>
                      <button
                        class="inline-copy-btn"
                        :title="t('scm.actions.copy_hash')"
                        @click="handleCopyHash(selectedCommit.hash)"
                      >
                        <span class="codicon codicon-copy"></span>
                      </button>
                    </div>
                  </div>

                  <div v-if="selectedCommit.parents && selectedCommit.parents.length > 0" class="meta-field">
                    <span class="meta-field-label">{{ t('scm.detail.parents') }}:</span>
                    <div class="meta-field-val parents-val">
                      <span
                        v-for="p in selectedCommit.parents"
                        :key="p"
                        class="inline-parent-link font-mono"
                        :title="p"
                        @click="jumpToCommit(p)"
                      >
                        {{ p.slice(0, 7) }}
                      </span>
                    </div>
                  </div>

                  <div class="meta-field">
                    <span class="meta-field-label">{{ t('scm.detail.author') }}:</span>
                    <div class="meta-field-val">
                      {{ selectedCommit.author }} &lt;{{ selectedCommit.email }}&gt;
                    </div>
                  </div>

                  <div class="meta-field">
                    <span class="meta-field-label">{{ t('scm.detail.author_date') }}:</span>
                    <div class="meta-field-val">
                      {{ formatExactDate(selectedCommit.timestamp) }}
                    </div>
                  </div>

                  <div v-if="selectedCommitDetail?.committer" class="meta-field">
                    <span class="meta-field-label">{{ t('scm.detail.committer') }}:</span>
                    <div class="meta-field-val">
                      {{ selectedCommitDetail.committer }} &lt;{{ selectedCommitDetail.committerEmail }}&gt;
                    </div>
                  </div>

                  <div v-if="selectedCommitDetail?.committerTimestamp" class="meta-field">
                    <span class="meta-field-label">{{ t('scm.detail.committer_date') }}:</span>
                    <div class="meta-field-val">
                      {{ formatExactDate(selectedCommitDetail.committerTimestamp) }}
                    </div>
                  </div>
                </div>

                <!-- GPG Signature Card -->
                <div
                  v-if="currentGpgInfo"
                  class="inline-gpg-card"
                  :class="'gpg-card-' + currentGpgInfo.status"
                >
                  <div class="inline-gpg-header">
                    <span class="codicon" :class="getGpgIcon(currentGpgInfo.status)"></span>
                    <span class="inline-gpg-title">{{ getGpgFullLabel(currentGpgInfo.status) }}</span>
                    <span class="inline-gpg-pill">{{ currentGpgInfo.status.toUpperCase() }}</span>
                  </div>
                  <div v-if="currentGpgInfo.status !== 'none'" class="inline-gpg-details font-mono">
                    <div v-if="currentGpgInfo.signer">
                      <span class="gpg-k">{{ t('scm.gpg.signer') }}:</span>
                      <span class="gpg-v">{{ currentGpgInfo.signer }}</span>
                    </div>
                    <div v-if="currentGpgInfo.keyId">
                      <span class="gpg-k">{{ t('scm.gpg.key_id') }}:</span>
                      <span class="gpg-v">{{ currentGpgInfo.keyId }}</span>
                    </div>
                    <div v-if="currentGpgInfo.fingerprint">
                      <span class="gpg-k">{{ t('scm.gpg.fingerprint') }}:</span>
                      <span class="gpg-v fingerprint">{{ currentGpgInfo.fingerprint }}</span>
                    </div>
                  </div>
                  <div v-else class="inline-gpg-unsigned">
                    <span>{{ t('scm.gpg.unsigned_hint') }}</span>
                  </div>
                </div>

                <!-- Full Commit Message -->
                <div class="inline-message-box">
                  <pre class="inline-message-text">{{ selectedCommitDetail?.message || selectedCommit.message }}</pre>
                </div>
              </div>

              <!-- Vertical Divider -->
              <div class="inline-col-divider"></div>

              <!-- Right Column: Changed Files Tree / List -->
              <div class="inline-files-column">
                <div class="inline-files-header">
                  <span class="inline-files-count">
                    {{ t('scm.detail.changed_files') }} ({{ selectedCommitDetail?.files?.length ?? 0 }})
                  </span>
                  <div class="inline-files-search">
                    <span class="codicon codicon-search"></span>
                    <input
                      v-model="filesSearchQuery"
                      type="text"
                      class="files-search-input"
                      :placeholder="t('scm.graph.search_placeholder')"
                    />
                    <button
                      v-if="filesSearchQuery"
                      class="files-clear-search"
                      @click="filesSearchQuery = ''"
                    >
                      <span class="codicon codicon-close"></span>
                    </button>
                  </div>
                </div>

                <div v-if="isLoadingDetail" class="inline-files-loading">
                  <span class="codicon codicon-loading codicon-modifier-spin"></span>
                  <span>{{ t('scm.graph.loading_files') }}</span>
                </div>

                <!-- Tree View -->
                <div
                  v-else-if="filesViewMode === 'tree'"
                  class="inline-files-tree"
                >
                  <div
                    v-for="row in flatTreeRows"
                    :key="row.id"
                    class="tree-node-row"
                    :class="{ 'is-dir': row.isDir, 'is-file': !row.isDir }"
                    :style="{ paddingLeft: (row.depth * 16 + 8) + 'px' }"
                    @click="row.isDir ? toggleFolder(row.path) : handleOpenFileDiff(row.path, selectedCommit.hash, selectedCommit.parents?.[0])"
                  >
                    <!-- Folder row -->
                    <template v-if="row.isDir">
                      <span
                        class="codicon tree-arrow"
                        :class="row.collapsed ? 'codicon-chevron-right' : 'codicon-chevron-down'"
                      ></span>
                      <span
                        class="codicon tree-folder-icon"
                        :class="row.collapsed ? 'codicon-folder' : 'codicon-folder-opened'"
                      ></span>
                      <span class="tree-node-name folder-name">{{ row.name }}</span>
                    </template>

                    <!-- File row -->
                    <template v-else>
                      <span class="tree-spacer"></span>
                      <TechIcon :id="getFileName(row.path)" :size="15" />
                      <span class="file-status-pill" :class="'badge-' + (row.file?.status || 'M')">{{ row.file?.status }}</span>
                      <span class="tree-node-name file-name font-mono" :title="row.path">{{ row.name }}</span>

                      <div v-if="row.file" class="file-diff-stats">
                        <span v-if="row.file.additions > 0" class="stat-add">+{{ row.file.additions }}</span>
                        <span v-if="row.file.deletions > 0" class="stat-del">-{{ row.file.deletions }}</span>
                      </div>

                      <div class="file-node-actions">
                        <button
                          class="node-action-btn"
                          :title="t('scm.actions.diff')"
                          @click.stop="handleOpenFileDiff(row.path, selectedCommit.hash, selectedCommit.parents?.[0])"
                        >
                          <span class="codicon codicon-diff"></span>
                        </button>
                        <button
                          class="node-action-btn"
                          :title="t('scm.diff.open_file')"
                          @click.stop="handleOpenFileInEditor(row.path)"
                        >
                          <span class="codicon codicon-go-to-file"></span>
                        </button>
                      </div>
                    </template>
                  </div>

                  <div v-if="flatTreeRows.length === 0" class="no-files-hint">
                    {{ t('scm.graph.no_commits') }}
                  </div>
                </div>

                <!-- List View -->
                <div
                  v-else
                  class="inline-files-list"
                >
                  <div
                    v-for="file in flatFileList"
                    :key="file.path"
                    class="list-file-row"
                    @click="handleOpenFileDiff(file.path, selectedCommit.hash, selectedCommit.parents?.[0])"
                  >
                    <TechIcon :id="getFileName(file.path)" :size="15" />
                    <span class="file-status-pill" :class="'badge-' + file.status">{{ file.status }}</span>
                    <span class="list-file-path font-mono" :title="file.path">{{ file.path }}</span>

                    <div class="file-diff-stats">
                      <span v-if="file.additions > 0" class="stat-add">+{{ file.additions }}</span>
                      <span v-if="file.deletions > 0" class="stat-del">-{{ file.deletions }}</span>
                    </div>

                    <div class="file-node-actions">
                      <button
                        class="node-action-btn"
                        :title="t('scm.actions.diff')"
                        @click.stop="handleOpenFileDiff(file.path, selectedCommit.hash, selectedCommit.parents?.[0])"
                      >
                        <span class="codicon codicon-diff"></span>
                      </button>
                      <button
                        class="node-action-btn"
                        :title="t('scm.diff.open_file')"
                        @click.stop="handleOpenFileInEditor(file.path)"
                      >
                        <span class="codicon codicon-go-to-file"></span>
                      </button>
                    </div>
                  </div>

                  <div v-if="flatFileList.length === 0" class="no-files-hint">
                    {{ t('scm.graph.no_commits') }}
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Empty State -->
          <div v-if="filteredCommits.length === 0 && !isLoadingGraph" class="graph-empty-state">
            <span class="codicon codicon-inbox empty-icon"></span>
            <span>{{ t('scm.graph.no_commits') }}</span>
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
  type GitCommitFileItem,
} from "../../composables/useGit";
import { layoutGitGraph } from "../../utils/gitGraphLayout";
import { useToast } from "../../composables/useToast";
import { t } from "../../i18n";

const ROW_HEIGHT = 34;
const EXPANDED_PANEL_HEIGHT = 320;

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

// Inline File Tree & Search State
const filesViewMode = ref<"tree" | "list">("tree");
const filesSearchQuery = ref("");
const collapsedFolders = ref<Set<string>>(new Set());

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

const selectedCommitIndex = computed(() => {
  if (!selectedCommit.value) return -1;
  return filteredCommits.value.findIndex(
    (c) => c.hash === selectedCommit.value?.hash
  );
});

function getRowTop(idx: number): number {
  if (selectedCommitIndex.value === -1 || idx <= selectedCommitIndex.value) {
    return idx * ROW_HEIGHT;
  }
  return idx * ROW_HEIGHT + EXPANDED_PANEL_HEIGHT;
}

// Computed SVG Graph Layout
const graphLayout = computed(() => {
  return layoutGitGraph(filteredCommits.value, {
    rowHeight: ROW_HEIGHT,
    laneWidth: 18,
    startX: 16,
    expandedIndex: selectedCommitIndex.value,
    expandedHeight: EXPANDED_PANEL_HEIGHT,
  });
});

const graphColWidth = computed(() => {
  return Math.max(68, graphLayout.value.width);
});

const totalBodyHeight = computed(() => {
  const base = filteredCommits.value.length * ROW_HEIGHT;
  const extra = selectedCommitIndex.value !== -1 ? EXPANDED_PANEL_HEIGHT : 0;
  return Math.max(ROW_HEIGHT, base + extra);
});

interface FileTreeNode {
  name: string;
  path: string;
  isDir: boolean;
  children: FileTreeNode[];
  file?: GitCommitFileItem;
}

interface FlatTreeRow {
  id: string;
  name: string;
  path: string;
  isDir: boolean;
  depth: number;
  file?: GitCommitFileItem;
  collapsed?: boolean;
}

const fileTreeData = computed<FileTreeNode[]>(() => {
  const files = selectedCommitDetail.value?.files || [];
  const query = filesSearchQuery.value.trim().toLowerCase();
  const filtered = query
    ? files.filter((f) => f.path.toLowerCase().includes(query))
    : files;

  const root: FileTreeNode = { name: "", path: "", isDir: true, children: [] };
  for (const f of filtered) {
    const parts = f.path.split("/");
    let curr = root;
    for (let i = 0; i < parts.length; i++) {
      const part = parts[i];
      const isLast = i === parts.length - 1;
      let child = curr.children.find((c) => c.name === part);
      if (!child) {
        child = {
          name: part,
          path: parts.slice(0, i + 1).join("/"),
          isDir: !isLast,
          children: [],
          file: isLast ? f : undefined,
        };
        curr.children.push(child);
      }
      curr = child;
    }
  }

  function compactTree(nodes: FileTreeNode[]): FileTreeNode[] {
    return nodes.map((node) => {
      if (node.isDir) {
        let current = node;
        let combinedName = current.name;
        while (current.children.length === 1 && current.children[0].isDir) {
          current = current.children[0];
          combinedName += ` / ${current.name}`;
        }
        return {
          name: combinedName,
          path: current.path,
          isDir: true,
          children: compactTree(current.children),
        };
      }
      return node;
    });
  }

  return compactTree(root.children);
});

const flatTreeRows = computed<FlatTreeRow[]>(() => {
  const rows: FlatTreeRow[] = [];
  function traverse(nodes: FileTreeNode[], depth: number) {
    for (const node of nodes) {
      const isCollapsed = collapsedFolders.value.has(node.path);
      rows.push({
        id: node.path,
        name: node.name,
        path: node.path,
        isDir: node.isDir,
        depth,
        file: node.file,
        collapsed: isCollapsed,
      });
      if (node.isDir && !isCollapsed) {
        traverse(node.children, depth + 1);
      }
    }
  }
  traverse(fileTreeData.value, 0);
  return rows;
});

const flatFileList = computed(() => {
  const files = selectedCommitDetail.value?.files || [];
  const query = filesSearchQuery.value.trim().toLowerCase();
  if (!query) return files;
  return files.filter((f) => f.path.toLowerCase().includes(query));
});

function toggleFolder(folderPath: string) {
  if (collapsedFolders.value.has(folderPath)) {
    collapsedFolders.value.delete(folderPath);
  } else {
    collapsedFolders.value.add(folderPath);
  }
}

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
  if (selectedCommit.value?.hash === commit.hash) {
    selectedCommit.value = null;
    return;
  }
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
      tableWrapperRef.value.scrollTop = Math.max(0, getRowTop(idx) - 100);
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

/* Inline Expanded Commit Detail Panel */
.inline-commit-detail-panel {
  position: absolute;
  box-sizing: border-box;
  background: var(--card-bg, #0f172a);
  border-left: 2px solid var(--accent-cyan, #00e5ff);
  border-top: 1px solid var(--card-border);
  border-bottom: 1px solid var(--card-border);
  border-right: 1px solid var(--card-border);
  z-index: 15;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  box-shadow: 0 6px 18px rgba(0, 0, 0, 0.25);
}

.inline-detail-actions-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 4px 10px;
  background: var(--section-bg, #1e293b);
  border-bottom: 1px solid var(--card-border);
  flex-shrink: 0;
  height: 28px;
}

.inline-detail-actions-left,
.inline-detail-actions-right {
  display: flex;
  align-items: center;
  gap: 6px;
}

.inline-tool-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  background: transparent;
  border: 1px solid var(--card-border);
  color: var(--text-main);
  font-size: 11px;
  padding: 2px 7px;
  border-radius: 3px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.inline-tool-btn:hover {
  background: rgba(128, 128, 128, 0.15);
  border-color: var(--accent-cyan);
}

.inline-icon-btn,
.inline-close-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 3px 5px;
  border-radius: 3px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  transition: all 0.15s ease;
}

.inline-icon-btn:hover,
.inline-close-btn:hover {
  background: rgba(128, 128, 128, 0.15);
  color: var(--text-main);
}

.inline-icon-btn.active {
  background: rgba(0, 229, 255, 0.15);
  color: var(--accent-cyan);
}

.action-divider {
  width: 1px;
  height: 14px;
  background: var(--card-border);
  margin: 0 2px;
}

/* Two-column layout */
.inline-detail-columns {
  flex: 1;
  display: flex;
  min-height: 0;
  overflow: hidden;
}

.inline-meta-column {
  flex: 0 0 45%;
  min-width: 280px;
  max-width: 50%;
  display: flex;
  flex-direction: column;
  padding: 10px 12px;
  gap: 8px;
  overflow-y: auto;
  border-right: 1px solid var(--card-border);
  box-sizing: border-box;
}

.meta-field-group {
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: 11px;
}

.meta-field {
  display: flex;
  align-items: flex-start;
  gap: 6px;
  line-height: 1.4;
}

.meta-field-label {
  color: var(--text-muted);
  font-weight: 600;
  min-width: 82px;
  flex-shrink: 0;
}

.meta-field-val {
  color: var(--text-main);
  word-break: break-all;
}

.hash-val {
  display: flex;
  align-items: center;
  gap: 4px;
  color: var(--accent-cyan);
}

.inline-copy-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 1px 3px;
  border-radius: 2px;
  display: inline-flex;
}

.inline-copy-btn:hover {
  background: rgba(128, 128, 128, 0.15);
  color: var(--text-main);
}

.parents-val {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.inline-parent-link {
  color: var(--accent-cyan);
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  padding: 1px 5px;
  border-radius: 3px;
  cursor: pointer;
}

.inline-parent-link:hover {
  text-decoration: underline;
}

.inline-gpg-card {
  padding: 6px 8px;
  border-radius: 4px;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  font-size: 11px;
}

.inline-gpg-card.gpg-card-verified {
  border-color: rgba(46, 160, 67, 0.4);
  background: rgba(46, 160, 67, 0.08);
}

.inline-gpg-card.gpg-card-unverified {
  border-color: rgba(245, 158, 11, 0.4);
  background: rgba(245, 158, 11, 0.08);
}

.inline-gpg-header {
  display: flex;
  align-items: center;
  gap: 6px;
  font-weight: 600;
}

.inline-gpg-pill {
  font-size: 9px;
  padding: 1px 4px;
  border-radius: 3px;
  background: rgba(128, 128, 128, 0.2);
  margin-left: auto;
}

.inline-gpg-details {
  margin-top: 4px;
  display: flex;
  flex-direction: column;
  gap: 2px;
  font-size: 10px;
  color: var(--text-muted);
}

.inline-gpg-unsigned {
  font-size: 10px;
  color: var(--text-muted);
}

.inline-message-box {
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  border-radius: 4px;
  padding: 8px 10px;
  flex: 1;
  min-height: 60px;
  overflow-y: auto;
}

.inline-message-text {
  margin: 0;
  font-family: inherit;
  font-size: 11px;
  line-height: 1.5;
  white-space: pre-wrap;
  color: var(--text-main);
}

/* Right Column: Files */
.inline-files-column {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  overflow: hidden;
  background: var(--bg-color, #070a12);
}

.inline-files-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 12px;
  background: var(--card-bg);
  border-bottom: 1px solid var(--card-border);
  flex-shrink: 0;
}

.inline-files-count {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-main);
}

.inline-files-search {
  display: flex;
  align-items: center;
  gap: 4px;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  border-radius: 4px;
  padding: 2px 6px;
  width: 140px;
}

.inline-files-search .codicon {
  font-size: 11px;
  color: var(--text-muted);
}

.files-search-input {
  background: transparent;
  border: none;
  outline: none;
  color: var(--text-main);
  font-size: 11px;
  width: 100%;
}

.files-clear-search {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 0;
  display: flex;
}

.inline-files-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 30px;
  color: var(--text-muted);
  font-size: 12px;
}

.inline-files-tree,
.inline-files-list {
  flex: 1;
  overflow-y: auto;
  padding: 4px 0;
}

.tree-node-row,
.list-file-row {
  display: flex;
  align-items: center;
  gap: 6px;
  height: 24px;
  padding: 0 10px;
  cursor: pointer;
  font-size: 11px;
  transition: background-color 0.1s ease;
  user-select: none;
}

.tree-node-row:hover,
.list-file-row:hover {
  background-color: var(--item-hover-bg, rgba(128, 128, 128, 0.08));
}

.tree-arrow {
  font-size: 11px;
  color: var(--text-muted);
  width: 12px;
  display: inline-flex;
  justify-content: center;
}

.tree-folder-icon {
  font-size: 13px;
  color: #38bdf8;
}

.tree-spacer {
  width: 12px;
}

.tree-node-name {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.folder-name {
  font-weight: 500;
  color: var(--text-main);
}

.file-name,
.list-file-path {
  flex: 1;
  color: var(--text-main);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.file-status-pill {
  font-size: 9px;
  font-weight: 700;
  padding: 0 4px;
  border-radius: 2px;
  line-height: 14px;
}

.file-status-pill.badge-M {
  background: rgba(245, 158, 11, 0.2);
  color: #fbbf24;
}

.file-status-pill.badge-A {
  background: rgba(46, 160, 67, 0.2);
  color: #4ade80;
}

.file-status-pill.badge-D {
  background: rgba(239, 68, 68, 0.2);
  color: #f87171;
}

.file-diff-stats {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 10px;
  font-family: monospace;
  margin-left: auto;
}

.stat-add {
  color: #4ade80;
}

.stat-del {
  color: #f87171;
}

.file-node-actions {
  display: none;
  align-items: center;
  gap: 2px;
  margin-left: 6px;
}

.tree-node-row:hover .file-node-actions,
.list-file-row:hover .file-node-actions {
  display: flex;
}

.node-action-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 2px;
  border-radius: 2px;
  display: flex;
  align-items: center;
}

.node-action-btn:hover {
  background: rgba(128, 128, 128, 0.15);
  color: var(--accent-cyan);
}

.no-files-hint {
  padding: 20px;
  text-align: center;
  color: var(--text-muted);
  font-size: 11px;
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
