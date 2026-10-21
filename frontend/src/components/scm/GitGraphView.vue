<template>
  <div class="git-graph-container" id="git-graph-view">
    <!-- Top Control & Filter Bar -->
    <div class="graph-toolbar">
      <!-- 0. Repo Selector (Multi-Repository) -->
      <div v-if="repositories.length > 0" class="toolbar-item repo-selector-item">
        <span class="toolbar-label">{{ t("scm.graph.repo_label") || "Repo:" }}</span>
        <div class="repo-select-wrapper">
          <select
            class="graph-repo-select"
            :value="activeRepo?.id || activeRepo?.name || ''"
            @change="handleRepoChange(($event.target as HTMLSelectElement).value)"
          >
            <option v-for="r in repositories" :key="r.id || r.rootPath" :value="r.id || r.name">
              {{ r.name }}
            </option>
          </select>
          <span class="codicon codicon-chevron-down repo-select-arrow"></span>
        </div>
      </div>

      <!-- 1. Branch Selector (Multi-Branch Filter Popover) -->
      <div class="toolbar-item">
        <GitBranchFilterPopover
          v-model="selectedBranches"
          :branches="branches"
          :current-branch="currentBranch"
          @change="handleBranchesChange"
        />
      </div>

      <!-- 2. File Filter Badge (When filtered by specific file) -->
      <div v-if="graphFilterPath" class="filter-file-badge">
        <span class="codicon codicon-file"></span>
        <span class="file-path-text font-mono" :title="graphFilterPath">{{ getFileName(graphFilterPath) }}</span>
        <button class="clear-filter-btn" :title="t('scm.graph.clear_filter')" @click="clearGraphFilterPath">
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

      <!-- 3.1 Conventional Commit Type Filter (Dropdown Select) -->
      <div class="toolbar-item type-selector-item">
        <div class="type-select-wrapper" :class="{ 'has-filter': selectedCommitType !== 'all' }">
          <span class="codicon codicon-tag type-select-icon"></span>
          <select v-model="selectedCommitType" class="graph-type-select" title="按提交类型过滤 (Conventional Commits)">
            <option v-for="tItem in commitTypeFilterOptions" :key="tItem.value" :value="tItem.value">
              {{ tItem.label }}
            </option>
          </select>
          <span class="codicon codicon-chevron-down type-select-arrow"></span>
        </div>
        <button
          v-if="selectedCommitType !== 'all'"
          class="clear-type-filter-btn"
          title="清除类型过滤"
          @click="selectedCommitType = 'all'"
        >
          <span class="codicon codicon-close"></span>
        </button>
      </div>

      <!-- 4. Refresh Button -->
      <button
        class="toolbar-btn"
        :class="{ loading: isLoadingGraph }"
        :title="t('scm.actions.refresh')"
        @click="handleManualRefresh"
      >
        <span class="codicon" :class="isLoadingGraph ? 'codicon-sync codicon-modifier-spin' : 'codicon-refresh'"></span>
        <span>{{ t("common.refresh") }}</span>
      </button>

      <!-- 5. Show Stashes Toggle Button -->
      <button
        class="toolbar-btn"
        :class="{ active: showStashes }"
        :title="t('scm.graph.toggle_stashes')"
        @click="toggleShowStashes"
      >
        <span class="codicon codicon-archive"></span>
        <span>{{ t("scm.graph.toggle_stashes") }}</span>
      </button>

      <!-- 5. Commit Count Badge -->
      <div
        class="commit-count-pill"
        :class="{ clickable: hasMoreCommits }"
        :title="hasMoreCommits ? t('scm.graph.click_to_load_more') : undefined"
        @click="hasMoreCommits ? handleLoadMore() : undefined"
      >
        <span v-if="isLoadingMore" class="codicon codicon-loading codicon-modifier-spin"></span>
        <span
          >{{ t("scm.graph.commits_count", { count: filteredCommits.length }) }}{{ hasMoreCommits ? "+" : "" }}</span
        >
      </div>
    </div>

    <!-- Main Content Area: Split Table & Commit Details Panel -->
    <div class="graph-main-split">
      <!-- Graph Table Area -->
      <div class="graph-table-wrapper" ref="tableWrapperRef" @scroll="handleScroll">
        <!-- Top Indeterminate Progress Bar for Graph Loading / Updating -->
        <div v-if="isGraphLoading" class="graph-loading-progressbar">
          <div class="loading-progress-indeterminate"></div>
        </div>

        <!-- Floating Refreshing Badge when commits are already loaded -->
        <div v-if="isGraphLoading && filteredCommits.length > 0" class="graph-refreshing-badge">
          <span class="codicon codicon-loading codicon-modifier-spin"></span>
          <span>{{ t("scm.graph.refreshing") || "正在更新 Git 图谱..." }}</span>
        </div>

        <div class="graph-table-header">
          <div class="col-header col-graph" :style="{ width: graphColWidth + 'px' }">
            {{ t("scm.graph.col_graph") }}
          </div>
          <div class="col-header col-desc">{{ t("scm.graph.col_desc") }}</div>
          <div class="col-header col-date">{{ t("scm.graph.col_date") }}</div>
          <div class="col-header col-author">{{ t("scm.graph.col_author") }}</div>
          <div class="col-header col-hash">{{ t("scm.graph.col_commit") }}</div>
          <div class="col-header col-actions">{{ t("scm.graph.col_actions") }}</div>
        </div>

        <div
          class="graph-table-body"
          :class="{ 'table-is-refreshing': isGraphLoading && filteredCommits.length > 0 }"
          :style="{ height: totalBodyHeight + 'px' }"
        >
          <!-- SVG DAG Overlay for Lanes, Nodes, Bezier Curves -->
          <svg class="graph-svg-layer" :style="{ width: graphColWidth + 'px', height: totalSvgHeight + 'px' }">
            <!-- 1. Connecting Paths (Vertical lines & Bezier curves) -->
            <path
              v-for="p in graphLayout.paths"
              :key="p.id"
              :d="p.d"
              :stroke="p.color"
              :stroke-dasharray="p.id.includes('*') ? '3 3' : undefined"
              stroke-width="2"
              fill="none"
              stroke-linecap="round"
              stroke-linejoin="round"
              class="graph-path-line"
              :class="{ 'uncommitted-path': p.id.includes('*') }"
            />

            <!-- 2. Commit Nodes (Dots) -->
            <g v-for="node in graphLayout.nodes" :key="'node:' + node.hash">
              <!-- Uncommitted Changes Hollow Ring -->
              <template v-if="node.hash === '*'">
                <circle
                  :cx="node.x"
                  :cy="node.y"
                  r="5.5"
                  fill="var(--card-bg, #081a24)"
                  :stroke="node.color"
                  stroke-width="2"
                  class="node-uncommitted-ring"
                />
              </template>
              <template v-else>
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
              </template>
            </g>
          </svg>

          <!-- Row Content Rows (HTML rows aligned with SVG rowHeight) -->
          <div
            v-for="(commit, idx) in filteredCommits"
            :key="commit.hash"
            class="graph-row"
            :class="{
              selected: isRowSelected(commit),
              'uncommitted-row': commit.hash === '*',
            }"
            :style="{ top: getRowTop(idx) + 'px', height: ROW_HEIGHT + 'px' }"
            @click="handleRowClick($event, commit)"
            @contextmenu.prevent="openContextMenu($event, commit)"
          >
            <!-- Col 1: Empty placeholder matching graph SVG width -->
            <div class="col-cell col-graph" :style="{ width: graphColWidth + 'px' }"></div>

            <!-- Col 2: Ref Badges & Commit Message -->
            <div class="col-cell col-desc">
              <!-- Uncommitted Changes virtual row -->
              <template v-if="commit.hash === '*'">
                <span class="uncommitted-circle-icon">○</span>
                <span class="commit-subject uncommitted-title" :title="commit.message">{{ commit.message }}</span>
              </template>
              <!-- Regular Commit row -->
              <template v-else>
                <!-- Ref Badges (Branches, Remotes, Tags) -->
                <div v-if="commit.refs && commit.refs.length > 0" class="ref-badges-group">
                  <span
                    v-for="(refItem, rIdx) in commit.refs"
                    :key="rIdx"
                    class="ref-badge"
                    :class="['badge-' + refItem.type, { current: refItem.current }]"
                    :title="refItem.name"
                    @contextmenu.stop.prevent="openRefContextMenu($event, refItem, commit)"
                  >
                    <span
                      class="codicon"
                      :class="
                        refItem.type === 'tag'
                          ? 'codicon-tag'
                          : refItem.type === 'remote'
                            ? 'codicon-cloud'
                            : refItem.type === 'stash'
                              ? 'codicon-archive'
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

                <!-- Message Text with Conventional Commits Pill Badges -->
                <span class="commit-subject" :title="commit.message">
                  <template v-if="parseCommitSubject(commit.message)">
                    <span class="cc-type-badge" :class="'cc-' + parseCommitSubject(commit.message)!.type">
                      {{ parseCommitSubject(commit.message)!.type }}
                    </span>
                    <span v-if="parseCommitSubject(commit.message)!.scope" class="cc-scope-badge">
                      {{ parseCommitSubject(commit.message)!.scope }}
                    </span>
                    <span v-if="parseCommitSubject(commit.message)!.isBreaking" class="cc-breaking-badge">!</span>
                    <span class="cc-desc-text">{{ parseCommitSubject(commit.message)!.description }}</span>
                  </template>
                  <template v-else>
                    {{ commit.message }}
                  </template>
                </span>
              </template>
            </div>

            <!-- Col 3: Date -->
            <div class="col-cell col-date" :title="commit.hash === '*' ? '*' : formatExactDate(commit.timestamp)">
              {{ commit.hash === "*" ? "*" : formatRelativeTime(commit.timestamp) }}
            </div>

            <!-- Col 4: Author -->
            <div class="col-cell col-author" :title="commit.hash === '*' ? '*' : `${commit.author} <${commit.email}>`">
              <template v-if="commit.hash === '*'">
                <span class="author-star font-mono">*</span>
              </template>
              <template v-else>
                <span class="author-avatar">{{ getInitials(commit.author) }}</span>
                <span class="author-name">{{ commit.author }}</span>
              </template>
            </div>

            <!-- Col 5: Short Hash -->
            <div class="col-cell col-hash">
              <span class="hash-text font-mono">{{ commit.shortHash }}</span>
              <button
                v-if="commit.hash !== '*'"
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
            v-if="selectedCommit && expandedCommitIndex !== -1"
            class="inline-commit-detail-panel"
            :style="{
              top: (expandedCommitIndex + 1) * ROW_HEIGHT + 'px',
              left: graphColWidth + 'px',
              height: EXPANDED_PANEL_HEIGHT + 'px',
              width: `calc(100% - ${graphColWidth}px)`,
            }"
            @click.stop
          >
            <!-- Inline Action Bar (Top) -->
            <div class="inline-detail-actions-bar">
              <div class="inline-detail-actions-left">
                <template v-if="isComparing && activeCompare">
                  <div class="inline-compare-title">
                    <span class="codicon codicon-diff"></span>
                    <span>{{
                      t("scm.compare.comparing", {
                        base: activeCompare.base.shortHash,
                        target: activeCompare.target.shortHash,
                      })
                    }}</span>
                  </div>
                  <button class="inline-tool-btn" :title="t('scm.compare.swap')" @click="handleSwapCompare">
                    <span class="codicon codicon-arrow-swap"></span>
                    <span>{{ t("scm.compare.swap") }}</span>
                  </button>
                </template>
                <template v-else-if="selectedCommit?.hash === '*'">
                  <button
                    class="inline-tool-btn"
                    :title="t('scm.uncommitted.stash')"
                    @click="openModal('stash-uncommitted', selectedCommit)"
                  >
                    <span class="codicon codicon-archive"></span>
                    <span>{{ t("scm.uncommitted.stash") }}</span>
                  </button>
                  <button
                    class="inline-tool-btn danger"
                    :title="t('scm.uncommitted.reset')"
                    @click="openModal('reset-uncommitted', selectedCommit)"
                  >
                    <span class="codicon codicon-discard"></span>
                    <span>{{ t("scm.uncommitted.reset") }}</span>
                  </button>
                  <button
                    class="inline-tool-btn danger"
                    :title="t('scm.uncommitted.clean')"
                    @click="openModal('clean-uncommitted', selectedCommit)"
                  >
                    <span class="codicon codicon-trash"></span>
                    <span>{{ t("scm.uncommitted.clean") }}</span>
                  </button>
                  <button
                    class="inline-tool-btn"
                    :title="t('scm.uncommitted.open_scm')"
                    @click="emit('switch-tab', 'changes')"
                  >
                    <span class="codicon codicon-source-control"></span>
                    <span>{{ t("scm.uncommitted.open_scm") }}</span>
                  </button>
                </template>
                <template v-else>
                  <button
                    class="inline-tool-btn"
                    :title="t('scm.actions.tag')"
                    @click="openModal('create-tag', selectedCommit)"
                  >
                    <span class="codicon codicon-tag"></span>
                    <span>{{ t("scm.actions.tag") }}</span>
                  </button>
                  <button
                    class="inline-tool-btn"
                    :title="t('scm.actions.branch')"
                    @click="openModal('create-branch', selectedCommit)"
                  >
                    <span class="codicon codicon-git-branch"></span>
                    <span>{{ t("scm.actions.branch") }}</span>
                  </button>
                </template>
              </div>

              <div class="inline-detail-actions-right">
                <!-- Collapse / Expand All (Tree mode only) -->
                <button
                  v-if="filesViewMode === 'tree'"
                  class="inline-icon-btn"
                  :title="t('scm.detail.collapse_all') || '折叠全部目录'"
                  @click="collapseAllFolders"
                >
                  <span class="codicon codicon-collapse-all"></span>
                </button>
                <button
                  v-if="filesViewMode === 'tree'"
                  class="inline-icon-btn"
                  :title="t('scm.detail.expand_all') || '展开全部目录'"
                  @click="expandAllFolders"
                >
                  <span class="codicon codicon-expand-all"></span>
                </button>
                <div v-if="filesViewMode === 'tree'" class="action-divider"></div>

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
                <button class="inline-close-btn" :title="t('common.close')" @click="closeDetailOrCompare">
                  <span class="codicon codicon-close"></span>
                </button>
              </div>
            </div>

            <!-- Inline Split Columns (Meta on Left, Files on Right) -->
            <div class="inline-detail-columns">
              <!-- Left Column: Commit Metadata & GPG & Message -->
              <div class="inline-meta-column">
                <template v-if="isComparing && activeCompare">
                  <!-- Compare Summary Stats Card -->
                  <div class="compare-stats-card">
                    <div class="compare-stats-header">
                      <span class="codicon codicon-diff-multiple"></span>
                      <span class="compare-stats-title">{{ t("scm.compare.title") }}</span>
                    </div>
                    <div class="compare-stats-pill font-mono">
                      <span class="stat-files">{{ activeCompare.files.length }} {{ t("scm.changes.total") }}</span>
                      <span class="stat-divider">•</span>
                      <span class="stat-add">+{{ activeCompare.totalAdditions }}</span>
                      <span class="stat-del">-{{ activeCompare.totalDeletions }}</span>
                    </div>
                  </div>

                  <!-- Base Commit Card -->
                  <div class="compare-commit-box">
                    <div class="compare-box-label font-mono">
                      <span class="box-badge base-badge">{{ t("scm.compare.base") }}</span>
                      <span class="hash-link" @click="handleCopyHash(activeCompare.base.hash)">
                        {{ activeCompare.base.shortHash }}
                        <span class="codicon codicon-copy"></span>
                      </span>
                    </div>
                    <div class="compare-box-subject" :title="activeCompare.base.message">
                      {{ activeCompare.base.message }}
                    </div>
                    <div class="compare-box-meta">
                      <span>{{ activeCompare.base.author }}</span>
                      <span class="meta-dot">•</span>
                      <span>{{ formatRelativeTime(activeCompare.base.timestamp) }}</span>
                    </div>
                  </div>

                  <div class="compare-arrow-divider">
                    <span class="codicon codicon-arrow-down"></span>
                  </div>

                  <!-- Target Commit Card -->
                  <div class="compare-commit-box">
                    <div class="compare-box-label font-mono">
                      <span class="box-badge target-badge">{{ t("scm.compare.target") }}</span>
                      <span class="hash-link" @click="handleCopyHash(activeCompare.target.hash)">
                        {{ activeCompare.target.shortHash }}
                        <span class="codicon codicon-copy"></span>
                      </span>
                    </div>
                    <div class="compare-box-subject" :title="activeCompare.target.message">
                      {{ activeCompare.target.message }}
                    </div>
                    <div class="compare-box-meta">
                      <span>{{ activeCompare.target.author }}</span>
                      <span class="meta-dot">•</span>
                      <span>{{ formatRelativeTime(activeCompare.target.timestamp) }}</span>
                    </div>
                  </div>
                </template>

                <template v-else-if="selectedCommit.hash === '*'">
                  <div class="uncommitted-meta-card">
                    <div class="uncommitted-card-header">
                      <span class="codicon codicon-diff-ignored"></span>
                      <span class="uncommitted-card-title">{{ t("scm.uncommitted.displaying") }}</span>
                    </div>
                    <div class="uncommitted-stats-list">
                      <div class="uncommitted-stat-row">
                        <span class="stat-label">{{ t("scm.changes.staged") }}:</span>
                        <span class="stat-num staged font-mono">{{ status.staged.length }}</span>
                      </div>
                      <div class="uncommitted-stat-row">
                        <span class="stat-label">{{ t("scm.changes.unstaged") }}:</span>
                        <span class="stat-num unstaged font-mono">{{ status.unstaged.length }}</span>
                      </div>
                      <div class="uncommitted-stat-row total-row">
                        <span class="stat-label">{{ t("scm.changes.total") }}:</span>
                        <span class="stat-num total font-mono">{{ totalChanges }}</span>
                      </div>
                    </div>
                    <div class="uncommitted-card-hint">
                      <span class="codicon codicon-info"></span>
                      <span>{{ t("scm.uncommitted.hint") }}</span>
                    </div>
                  </div>
                </template>

                <template v-else>
                  <div class="meta-field-group">
                    <div class="meta-field">
                      <span class="meta-field-label">{{ t("scm.detail.commit") }}:</span>
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
                      <span class="meta-field-label">{{ t("scm.detail.parents") }}:</span>
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
                      <span class="meta-field-label">{{ t("scm.detail.author") }}:</span>
                      <div class="meta-field-val">{{ selectedCommit.author }} &lt;{{ selectedCommit.email }}&gt;</div>
                    </div>

                    <div class="meta-field">
                      <span class="meta-field-label">{{ t("scm.detail.author_date") }}:</span>
                      <div class="meta-field-val">
                        {{ formatExactDate(selectedCommit.timestamp) }}
                      </div>
                    </div>

                    <div v-if="selectedCommitDetail?.committer" class="meta-field">
                      <span class="meta-field-label">{{ t("scm.detail.committer") }}:</span>
                      <div class="meta-field-val">
                        {{ selectedCommitDetail.committer }} &lt;{{ selectedCommitDetail.committerEmail }}&gt;
                      </div>
                    </div>

                    <div v-if="selectedCommitDetail?.committerTimestamp" class="meta-field">
                      <span class="meta-field-label">{{ t("scm.detail.committer_date") }}:</span>
                      <div class="meta-field-val">
                        {{ formatExactDate(selectedCommitDetail.committerTimestamp) }}
                      </div>
                    </div>
                  </div>

                  <!-- GPG Signature Card -->
                  <div v-if="currentGpgInfo" class="inline-gpg-card" :class="'gpg-card-' + currentGpgInfo.status">
                    <div class="inline-gpg-header">
                      <span class="codicon" :class="getGpgIcon(currentGpgInfo.status)"></span>
                      <span class="inline-gpg-title">{{ getGpgFullLabel(currentGpgInfo.status) }}</span>
                      <span class="inline-gpg-pill">{{ currentGpgInfo.status.toUpperCase() }}</span>
                    </div>
                    <div v-if="currentGpgInfo.status !== 'none'" class="inline-gpg-details font-mono">
                      <div v-if="currentGpgInfo.signer">
                        <span class="gpg-k">{{ t("scm.gpg.signer") }}:</span>
                        <span class="gpg-v">{{ currentGpgInfo.signer }}</span>
                      </div>
                      <div v-if="currentGpgInfo.keyId">
                        <span class="gpg-k">{{ t("scm.gpg.key_id") }}:</span>
                        <span class="gpg-v">{{ currentGpgInfo.keyId }}</span>
                      </div>
                      <div v-if="currentGpgInfo.fingerprint">
                        <span class="gpg-k">{{ t("scm.gpg.fingerprint") }}:</span>
                        <span class="gpg-v fingerprint">{{ currentGpgInfo.fingerprint }}</span>
                      </div>
                    </div>
                    <div v-else class="inline-gpg-unsigned">
                      <span>{{ t("scm.gpg.unsigned_hint") }}</span>
                    </div>
                  </div>

                  <!-- Full Commit Message with Structured Conventional Presentation -->
                  <div class="inline-message-box">
                    <template v-if="parseFullCommitMessage(selectedCommitDetail?.message || selectedCommit.message)">
                      <div class="inline-message-top">
                        <div class="inline-subject-row">
                          <template
                            v-if="
                              parseFullCommitMessage(selectedCommitDetail?.message || selectedCommit.message)
                                .parsedSubject
                            "
                          >
                            <span
                              class="cc-type-badge large"
                              :class="
                                'cc-' +
                                parseFullCommitMessage(selectedCommitDetail?.message || selectedCommit.message)
                                  .parsedSubject!.type
                              "
                            >
                              {{
                                parseFullCommitMessage(selectedCommitDetail?.message || selectedCommit.message)
                                  .parsedSubject!.type
                              }}
                            </span>
                            <span
                              v-if="
                                parseFullCommitMessage(selectedCommitDetail?.message || selectedCommit.message)
                                  .parsedSubject!.scope
                              "
                              class="cc-scope-badge large"
                            >
                              ({{
                                parseFullCommitMessage(selectedCommitDetail?.message || selectedCommit.message)
                                  .parsedSubject!.scope
                              }})
                            </span>
                            <span
                              v-if="
                                parseFullCommitMessage(selectedCommitDetail?.message || selectedCommit.message)
                                  .parsedSubject!.isBreaking
                              "
                              class="cc-breaking-badge large"
                              >BREAKING</span
                            >
                            <span class="inline-subject-title">{{
                              parseFullCommitMessage(selectedCommitDetail?.message || selectedCommit.message)
                                .parsedSubject!.description
                            }}</span>
                          </template>
                          <template v-else>
                            <span class="inline-subject-title">{{
                              parseFullCommitMessage(selectedCommitDetail?.message || selectedCommit.message).subject
                            }}</span>
                          </template>
                        </div>
                        <button
                          class="inline-copy-msg-btn"
                          :title="t('scm.actions.copy_message') || '复制提交信息'"
                          @click.stop="
                            handleCopyCommitDetailMessage(selectedCommitDetail?.message || selectedCommit.message)
                          "
                        >
                          <span class="codicon codicon-copy"></span>
                        </button>
                      </div>

                      <div
                        v-if="parseFullCommitMessage(selectedCommitDetail?.message || selectedCommit.message).body"
                        class="inline-body-content"
                      >
                        <pre class="inline-body-text">{{
                          parseFullCommitMessage(selectedCommitDetail?.message || selectedCommit.message).body
                        }}</pre>
                      </div>
                    </template>
                  </div>
                </template>
              </div>

              <!-- Vertical Divider -->
              <div class="inline-col-divider"></div>

              <!-- Right Column: Changed Files Tree / List -->
              <div class="inline-files-column">
                <div class="inline-files-header">
                  <div class="inline-files-count-group">
                    <span class="inline-files-count">
                      {{ t("scm.detail.changed_files") }} ({{ currentFiles.length }})
                    </span>
                    <span v-if="totalAdditions > 0 || totalDeletions > 0" class="inline-files-stats-pill">
                      <span v-if="totalAdditions > 0" class="stat-add">+{{ totalAdditions }}</span>
                      <span v-if="totalDeletions > 0" class="stat-del">-{{ totalDeletions }}</span>
                    </span>
                  </div>
                  <div class="inline-files-search">
                    <span class="codicon codicon-search"></span>
                    <input
                      v-model="filesSearchQuery"
                      type="text"
                      class="files-search-input"
                      :placeholder="t('scm.graph.search_placeholder')"
                    />
                    <button v-if="filesSearchQuery" class="files-clear-search" @click="filesSearchQuery = ''">
                      <span class="codicon codicon-close"></span>
                    </button>
                  </div>
                </div>

                <div v-if="isLoadingDetail || isLoadingCompare" class="inline-files-loading">
                  <span class="codicon codicon-loading codicon-modifier-spin"></span>
                  <span>{{ t("scm.graph.loading_files") }}</span>
                </div>

                <!-- Tree View -->
                <div v-else-if="filesViewMode === 'tree'" class="inline-files-tree">
                  <div
                    v-for="row in flatTreeRows"
                    :key="row.id"
                    class="tree-node-row"
                    :class="{ 'is-dir': row.isDir, 'is-file': !row.isDir }"
                    :style="{ paddingLeft: row.depth * 16 + 8 + 'px' }"
                    @click="row.isDir ? toggleFolder(row.path) : handleFileClick(row.path)"
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
                      <span class="file-status-pill" :class="'badge-' + (row.file?.status || 'M')">{{
                        row.file?.status
                      }}</span>
                      <span class="tree-node-name file-name font-mono" :title="row.path">{{ row.name }}</span>

                      <div v-if="row.file" class="file-diff-stats">
                        <span v-if="row.file.additions > 0" class="stat-add">+{{ row.file.additions }}</span>
                        <span v-if="row.file.deletions > 0" class="stat-del">-{{ row.file.deletions }}</span>
                      </div>

                      <div class="file-node-actions">
                        <button
                          class="node-action-btn"
                          :title="t('scm.actions.diff')"
                          @click.stop="handleFileClick(row.path)"
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
                    {{ t("scm.graph.no_commits") }}
                  </div>
                </div>

                <!-- List View -->
                <div v-else class="inline-files-list">
                  <div
                    v-for="file in flatFileList"
                    :key="file.path"
                    class="list-file-row"
                    @click="handleFileClick(file.path)"
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
                        @click.stop="handleFileClick(file.path)"
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
                    {{ t("scm.graph.no_commits") }}
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Bottom Pagination / Load More Row -->
          <div
            v-if="filteredCommits.length > 0"
            class="graph-load-more-row"
            :style="{
              top: loadMoreTop + 'px',
              height: LOAD_MORE_HEIGHT + 'px',
            }"
          >
            <!-- When there are more commits to load -->
            <template v-if="hasMoreCommits">
              <button class="load-more-btn" :disabled="isLoadingMore || isLoadingGraph" @click="handleLoadMore">
                <span
                  class="codicon"
                  :class="
                    isLoadingMore || isLoadingGraph ? 'codicon-loading codicon-modifier-spin' : 'codicon-arrow-down'
                  "
                ></span>
                <span>
                  {{
                    isLoadingMore || isLoadingGraph
                      ? t("scm.graph.loading_more")
                      : t("scm.graph.load_more_commits", { count: 200 })
                  }}
                </span>
              </button>
            </template>
            <!-- When all commits have been loaded -->
            <template v-else>
              <div class="all-commits-loaded-hint">
                <span class="codicon codicon-check"></span>
                <span>{{ t("scm.graph.all_commits_loaded") }}</span>
              </div>
            </template>
          </div>

          <!-- Loading State (Initial / Empty Graph) -->
          <div v-if="isGraphLoading && filteredCommits.length === 0" class="graph-loading-state">
            <div class="graph-loading-spinner-wrapper">
              <span class="codicon codicon-loading codicon-modifier-spin loading-spinner"></span>
            </div>
            <div class="loading-text-group">
              <span class="loading-primary-text">{{ t("scm.graph.loading") || "正在加载 Git 提交图谱..." }}</span>
              <span class="loading-secondary-text">{{
                t("scm.graph.loading_desc") || "正在解析提交拓扑与分支历史，请稍候..."
              }}</span>
            </div>
          </div>

          <!-- Empty State -->
          <div v-else-if="filteredCommits.length === 0 && !isGraphLoading" class="graph-empty-state">
            <span class="codicon codicon-inbox empty-icon"></span>
            <span>{{ t("scm.graph.no_commits") }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Right-click Context Menu (Teleported) -->
    <Teleport to="body">
      <!-- Commit Context Menu (Aligned with Git Graph) -->
      <div
        v-if="contextMenuVisible"
        class="graph-context-menu"
        :style="{ left: contextMenuPos.x + 'px', top: contextMenuPos.y + 'px' }"
        @click.stop
      >
        <!-- Uncommitted Changes Dedicated Menu (Aligned with Git Graph) -->
        <template v-if="contextCommit?.hash === '*'">
          <div class="menu-item" @click="handleContextAction('stash-uncommitted')">
            {{ t("scm.uncommitted.stash") }}
          </div>
          <div class="menu-separator"></div>
          <div class="menu-item danger" @click="handleContextAction('reset-uncommitted')">
            {{ t("scm.uncommitted.reset") }}
          </div>
          <div class="menu-item danger" @click="handleContextAction('clean-uncommitted')">
            {{ t("scm.uncommitted.clean") }}
          </div>
          <div class="menu-separator"></div>
          <div class="menu-item" @click="handleContextAction('open-scm-view')">
            {{ t("scm.uncommitted.open_scm") }}
          </div>
        </template>

        <!-- Regular Commit Context Menu -->
        <template v-else>
          <template v-if="comparePair">
            <div class="menu-item primary-action" @click="handleContextAction('compare')">
              {{
                t("scm.commit.compare_with", { base: comparePair.base.shortHash, target: comparePair.target.shortHash })
              }}
            </div>
            <div class="menu-separator"></div>
          </template>
          <div class="menu-item" @click="handleContextAction('add-tag')">
            {{ t("scm.commit.add_tag") }}
          </div>
          <div class="menu-item" @click="handleContextAction('create-branch')">
            {{ t("scm.commit.create_branch") }}
          </div>
          <div class="menu-separator"></div>
          <div class="menu-item" @click="handleContextAction('checkout')">
            {{ t("scm.commit.checkout") }}
          </div>
          <div class="menu-item" @click="handleContextAction('cherry-pick')">
            {{ t("scm.commit.cherry_pick") }}
          </div>
          <div class="menu-item" @click="handleContextAction('revert')">
            {{ t("scm.commit.revert") }}
          </div>
          <div class="menu-separator"></div>
          <div class="menu-item" @click="handleContextAction('merge')">
            {{ t("scm.commit.merge") }}
          </div>
          <div class="menu-item" @click="handleContextAction('rebase')">
            {{ t("scm.commit.rebase") }}
          </div>
          <div class="menu-item danger" @click="handleContextAction('reset')">
            {{ t("scm.commit.reset") }}
          </div>
          <div class="menu-separator"></div>
          <div class="menu-item" @click="handleContextAction('copy-hash')">
            {{ t("scm.commit.copy_hash") }}
          </div>
          <div class="menu-item" @click="handleContextAction('copy-subject')">
            {{ t("scm.commit.copy_subject") }}
          </div>
        </template>
      </div>

      <!-- Ref Badge Context Menu (Aligned with 图 1 & 图 2) -->
      <div
        v-if="refContextMenuVisible"
        class="graph-context-menu"
        :style="{ left: refContextMenuPos.x + 'px', top: refContextMenuPos.y + 'px' }"
        @click.stop
      >
        <!-- Branch Context Menu -->
        <template v-if="selectedRef?.type === 'branch' || selectedRef?.type === 'head'">
          <!-- 图 1: Current Branch Context Menu -->
          <template v-if="isRefCurrent">
            <div class="menu-item" @click="handleRefAction('rename-branch')">
              {{ t("scm.branch.rename") }}
            </div>
            <div class="menu-item" @click="handleRefAction('push-branch')">
              {{ t("scm.branch.push") }}
            </div>
            <div class="menu-separator"></div>
            <div class="menu-item" @click="handleRefAction('archive')">
              {{ t("scm.branch.archive") }}
            </div>
            <div class="menu-item" @click="handleRefAction('toggle-dropdown-filter')">
              {{ isBranchFiltered ? t("scm.branch.unselect_filter") : t("scm.branch.select_filter") }}
            </div>
            <div class="menu-separator"></div>
            <div class="menu-item" @click="handleRefAction('copy-name')">
              {{ t("scm.branch.copy_name") }}
            </div>
          </template>

          <!-- 图 2: Inactive Branch Context Menu -->
          <template v-else>
            <div class="menu-item" @click="handleRefAction('checkout-branch')">
              {{ t("scm.branch.checkout") }}
            </div>
            <div class="menu-item" @click="handleRefAction('rename-branch')">
              {{ t("scm.branch.rename") }}
            </div>
            <div class="menu-item danger" @click="handleRefAction('delete-branch')">
              {{ t("scm.branch.delete") }}
            </div>
            <div class="menu-item" @click="handleRefAction('merge-branch')">
              {{ t("scm.branch.merge") }}
            </div>
            <div class="menu-item" @click="handleRefAction('rebase-branch')">
              {{ t("scm.branch.rebase") }}
            </div>
            <div class="menu-item" @click="handleRefAction('push-branch')">
              {{ t("scm.branch.push") }}
            </div>
            <div class="menu-separator"></div>
            <div class="menu-item" @click="handleRefAction('archive')">
              {{ t("scm.branch.archive") }}
            </div>
            <div class="menu-item" @click="handleRefAction('toggle-dropdown-filter')">
              {{ isBranchFiltered ? t("scm.branch.unselect_filter") : t("scm.branch.select_filter") }}
            </div>
            <div class="menu-separator"></div>
            <div class="menu-item" @click="handleRefAction('copy-name')">
              {{ t("scm.branch.copy_name") }}
            </div>
          </template>
        </template>

        <!-- Tag Context Menu -->
        <template v-else-if="selectedRef?.type === 'tag'">
          <div class="menu-item" @click="handleRefAction('checkout-tag')">
            {{ t("scm.tag.checkout") }}
          </div>
          <div class="menu-item" @click="handleRefAction('create-branch-from-tag')">
            {{ t("scm.tag.create_branch") }}
          </div>
          <div class="menu-item danger" @click="handleRefAction('delete-tag')">
            {{ t("scm.tag.delete") }}
          </div>
          <div class="menu-item" @click="handleRefAction('push-tag')">
            {{ t("scm.tag.push") }}
          </div>
          <div class="menu-separator"></div>
          <div class="menu-item" @click="handleRefAction('archive')">
            {{ t("scm.branch.archive") }}
          </div>
          <div class="menu-separator"></div>
          <div class="menu-item" @click="handleRefAction('copy-name')">
            {{ t("scm.tag.copy_name") }}
          </div>
        </template>

        <!-- Stash Context Menu -->
        <template v-else-if="selectedRef?.type === 'stash'">
          <div class="menu-item" @click="handleRefAction('apply-stash')">
            {{ t("scm.stash.apply") }}
          </div>
          <div class="menu-item" @click="handleRefAction('pop-stash')">
            {{ t("scm.stash.pop") }}
          </div>
          <div class="menu-item danger" @click="handleRefAction('drop-stash')">
            {{ t("scm.stash.drop") }}
          </div>
          <div class="menu-separator"></div>
          <div class="menu-item" @click="handleRefAction('branch-from-stash')">
            {{ t("scm.stash.create_branch") }}
          </div>
          <div class="menu-separator"></div>
          <div class="menu-item" @click="handleRefAction('copy-name')">
            {{ t("scm.branch.copy_name") }}
          </div>
        </template>

        <!-- Remote Branch Context Menu -->
        <template v-else-if="selectedRef?.type === 'remote'">
          <div class="menu-item" @click="handleRefAction('checkout-remote')">
            {{ t("scm.remote.checkout") }}
          </div>
          <div class="menu-item" @click="handleRefAction('pull-remote')">
            {{ t("scm.remote.pull") }}
          </div>
          <div class="menu-item" @click="handleRefAction('merge-remote')">
            {{ t("scm.remote.merge") }}
          </div>
          <div class="menu-separator"></div>
          <div class="menu-item" @click="handleRefAction('archive')">
            {{ t("scm.branch.archive") }}
          </div>
          <div class="menu-separator"></div>
          <div class="menu-item" @click="handleRefAction('copy-name')">
            {{ t("scm.branch.copy_name") }}
          </div>
        </template>
      </div>
    </Teleport>

    <!-- High-order Git Action Modal Dialog -->
    <GitActionModal
      :visible="modalVisible"
      :action-type="activeModalType"
      :target-commit="modalTargetCommit"
      :target-ref-name="modalTargetRefName"
      :current-branch-name="currentBranchName"
      @cancel="modalVisible = false"
      @confirm="handleModalConfirm"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from "vue";
import TechIcon from "../common/TechIcon.vue";
import GitActionModal, { type GitModalActionType } from "./GitActionModal.vue";
import GitBranchFilterPopover from "./GitBranchFilterPopover.vue";
import {
  useGit,
  type GitGraphCommit,
  type GitGpgInfo,
  type GitCommitFileItem,
  type GitGraphRef,
} from "../../composables/useGit";
import { layoutGitGraph } from "../../utils/gitGraphLayout";
import { useToast } from "../../composables/useToast";
import { t } from "../../i18n";

const ROW_HEIGHT = 34;
const EXPANDED_PANEL_HEIGHT = 320;

const {
  repositories,
  activeRepo,
  fetchRepositories,
  setActiveRepo,
  branches,
  status,
  currentBranch,
  totalChanges,
  graphCommits,
  selectedCommit,
  selectedCommitDetail,
  graphFilterBranches,
  graphFilterPath,
  isLoadingGraph,
  fetchBranches,
  fetchGraph,
  fetchStatus,
  fetchCommitDetail,
  clearGraphFilterPath,
  checkout,
  createTag,
  createBranchFromCommit,
  renameBranch,
  deleteBranch,
  mergeBranch,
  rebaseBranch,
  pushBranch,
  deleteTag,
  pushTag,
  downloadArchive,
  cherryPick,
  revertCommit,
  resetToCommit,
  activeCompare,
  fetchCompare,
  stashUncommitted,
  stashApply,
  stashPop,
  stashDrop,
  stashBranch,
  resetUncommitted,
  cleanUntracked,
  hasMoreCommits,
  isLoadingMore,
  loadMoreCommits,
  resetGraphLimit,
} = useGit();

const { showToast } = useToast();

const selectedBranches = ref<string[]>(
  graphFilterBranches.value && graphFilterBranches.value.length > 0 ? [...graphFilterBranches.value] : ["all"]
);
const showStashes = ref(true);
const searchQuery = ref("");
const selectedCommitType = ref<string>("all");

const commitTypeFilterOptions = [
  { label: "全部类型", value: "all", title: "显示全部提交" },
  { label: "feat (新功能)", value: "feat", title: "仅看新功能 (Features)" },
  { label: "fix (缺陷修复)", value: "fix", title: "仅看缺陷修复 (Bug Fixes)" },
  { label: "docs (文档更新)", value: "docs", title: "仅看文档更新 (Docs)" },
  { label: "style (代码风格)", value: "style", title: "仅看代码风格 (Style)" },
  { label: "refactor (代码重构)", value: "refactor", title: "仅看代码重构 (Refactor)" },
  { label: "perf (性能优化)", value: "perf", title: "仅看性能优化 (Performance)" },
  { label: "test (测试用例)", value: "test", title: "仅看测试用例 (Test)" },
  { label: "build (构建依赖)", value: "build", title: "仅看构建依赖 (Build)" },
  { label: "ci (持续集成)", value: "ci", title: "仅看持续集成 (CI/CD)" },
  { label: "chore (日常维护)", value: "chore", title: "仅看日常维护 (Chore)" },
  { label: "revert (代码回滚)", value: "revert", title: "仅看代码回滚 (Revert)" },
];
const tableWrapperRef = ref<HTMLElement | null>(null);
const isInitialLoading = ref<boolean>(graphCommits.value.length === 0);
const isGraphLoading = computed(() => isInitialLoading.value || isLoadingGraph.value);
const isLoadingDetail = ref(false);
const isLoadingCompare = ref(false);
const compareCommit = ref<GitGraphCommit | null>(null);

const isComparing = computed(() => {
  return !!(selectedCommit.value && compareCommit.value && activeCompare.value);
});

function parseCommitSubject(msg: string) {
  if (!msg || msg.startsWith("*") || msg.startsWith("Uncommitted")) return null;
  const firstLine = msg.split(/\r?\n/)[0];
  const m = firstLine.match(
    /^(feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(\(([a-zA-Z0-9_.\-/]+)\))?(!)?: (.+)$/
  );
  if (!m) return null;
  return {
    type: m[1],
    scope: m[3] || null,
    isBreaking: !!m[4],
    description: m[5],
  };
}

function parseFullCommitMessage(rawMsg: string) {
  if (!rawMsg) return { subject: "", body: "", parsedSubject: null };
  const lines = rawMsg.split(/\r?\n/);
  const firstLine = lines[0] || "";
  const body = lines.slice(1).join("\n").trim();
  const parsedSubject = parseCommitSubject(firstLine);
  return {
    subject: firstLine,
    body,
    parsedSubject,
  };
}

async function handleCopyCommitDetailMessage(msg: string) {
  try {
    await navigator.clipboard.writeText(msg);
    showToast(t("scm.actions.copied") || "已复制提交说明", "info");
  } catch {
    showToast("Copy failed", "error");
  }
}

const comparePair = computed<{ base: GitGraphCommit; target: GitGraphCommit } | null>(() => {
  if (compareCommit.value && selectedCommit.value) {
    return { base: selectedCommit.value, target: compareCommit.value };
  }
  if (contextCommit.value && selectedCommit.value && contextCommit.value.hash !== selectedCommit.value.hash) {
    return { base: selectedCommit.value, target: contextCommit.value };
  }
  return null;
});

// Inline File Tree & Search State
const filesViewMode = ref<"tree" | "list">("tree");
const filesSearchQuery = ref("");
const collapsedFolders = ref<Set<string>>(new Set());

// Context Menu State
const contextMenuVisible = ref(false);
const contextMenuPos = ref({ x: 0, y: 0 });
const contextCommit = ref<GitGraphCommit | null>(null);

// Ref Context Menu State (Branch / Tag context menu)
const refContextMenuVisible = ref(false);
const refContextMenuPos = ref({ x: 0, y: 0 });
const selectedRef = ref<GitGraphRef | null>(null);
const selectedRefCommit = ref<GitGraphCommit | null>(null);

// Modal Dialog State
const modalVisible = ref(false);
const activeModalType = ref<GitModalActionType | null>(null);
const modalTargetCommit = ref<GitGraphCommit | null>(null);
const modalTargetRefName = ref<string | undefined>(undefined);

const currentBranchName = computed(() => {
  return branches.value.find((b) => b.current)?.name || "";
});

const isRefCurrent = computed(() => {
  if (!selectedRef.value) return false;
  if (selectedRef.value.current) return true;
  if (currentBranchName.value && selectedRef.value.name === currentBranchName.value) return true;
  return false;
});

const isBranchFiltered = computed(() => {
  return selectedBranches.value.includes(selectedRef.value?.name || "");
});

const emit = defineEmits<{
  (e: "open-commit-diff", payload: { path: string; hash: string; parentHash?: string }): void;
  (e: "open-diff", payload: { path: string; staged: boolean }): void;
  (e: "switch-tab", tab: "changes" | "graph"): void;
  (e: "open-file", path: string): void;
}>();

// Virtual Uncommitted Commit node injected at top when totalChanges > 0
const currentHeadCommit = computed(() => {
  return graphCommits.value.find((c) => c.refs?.some((r) => r.current || r.type === "head")) || graphCommits.value[0];
});

const uncommittedVirtualCommit = computed<GitGraphCommit | null>(() => {
  if (totalChanges.value <= 0) return null;
  const head = currentHeadCommit.value;
  return {
    hash: "*",
    shortHash: "*",
    parents: head ? [head.hash] : [],
    author: "*",
    email: "",
    timestamp: Math.floor(Date.now() / 1000),
    message: `${t("scm.uncommitted.title")} (${totalChanges.value})`,
    refs: [],
  };
});

const allGraphCommits = computed(() => {
  if (uncommittedVirtualCommit.value) {
    return [uncommittedVirtualCommit.value, ...graphCommits.value];
  }
  return graphCommits.value;
});

// Filtered Commits by live search and conventional type
const filteredCommits = computed(() => {
  let list = allGraphCommits.value;
  if (selectedCommitType.value !== "all") {
    list = list.filter((c) => {
      if (c.hash === "*") return true; // Keep uncommitted virtual row
      const parsed = parseCommitSubject(c.message);
      return parsed && parsed.type === selectedCommitType.value;
    });
  }
  const query = searchQuery.value.trim().toLowerCase();
  if (!query) return list;
  return list.filter((c) => {
    return (
      c.message.toLowerCase().includes(query) ||
      c.author.toLowerCase().includes(query) ||
      c.hash.toLowerCase().includes(query) ||
      c.shortHash.toLowerCase().includes(query) ||
      c.refs?.some((r) => r.name.toLowerCase().includes(query))
    );
  });
});

const expandedCommitIndex = computed(() => {
  if (!selectedCommit.value) return -1;
  const idxA = filteredCommits.value.findIndex((c) => c.hash === selectedCommit.value?.hash);
  if (!compareCommit.value) return idxA;
  const idxB = filteredCommits.value.findIndex((c) => c.hash === compareCommit.value?.hash);
  if (idxA === -1) return idxB;
  if (idxB === -1) return idxA;
  return Math.max(idxA, idxB);
});

function isRowSelected(commit: GitGraphCommit): boolean {
  return selectedCommit.value?.hash === commit.hash || compareCommit.value?.hash === commit.hash;
}

function getRowTop(idx: number): number {
  if (expandedCommitIndex.value === -1 || idx <= expandedCommitIndex.value) {
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
    expandedIndex: expandedCommitIndex.value,
    expandedHeight: EXPANDED_PANEL_HEIGHT,
  });
});

const graphColWidth = computed(() => {
  return Math.max(68, graphLayout.value.width);
});

const LOAD_MORE_HEIGHT = 46;

const loadMoreTop = computed(() => {
  return getRowTop(filteredCommits.value.length);
});

const totalBodyHeight = computed(() => {
  const base = filteredCommits.value.length * ROW_HEIGHT;
  const extra = expandedCommitIndex.value !== -1 ? EXPANDED_PANEL_HEIGHT : 0;
  const bottomBar = filteredCommits.value.length > 0 ? LOAD_MORE_HEIGHT : 0;
  return Math.max(ROW_HEIGHT, base + extra + bottomBar);
});

const totalSvgHeight = computed(() => {
  const base = filteredCommits.value.length * ROW_HEIGHT;
  const extra = expandedCommitIndex.value !== -1 ? EXPANDED_PANEL_HEIGHT : 0;
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

const currentFiles = computed<GitCommitFileItem[]>(() => {
  if (selectedCommit.value?.hash === "*") {
    const list: GitCommitFileItem[] = [];
    const seen = new Set<string>();
    for (const f of status.value.staged) {
      if (!seen.has(f.path)) {
        seen.add(f.path);
        list.push({
          path: f.path,
          status: f.status,
          additions: 0,
          deletions: 0,
        });
      }
    }
    for (const f of status.value.unstaged) {
      if (!seen.has(f.path)) {
        seen.add(f.path);
        list.push({
          path: f.path,
          status: f.status,
          additions: 0,
          deletions: 0,
        });
      }
    }
    return list;
  }
  if (isComparing.value && activeCompare.value) {
    return activeCompare.value.files;
  }
  return selectedCommitDetail.value?.files || [];
});

const fileTreeData = computed<FileTreeNode[]>(() => {
  const files = currentFiles.value;
  const query = filesSearchQuery.value.trim().toLowerCase();
  const filtered = query ? files.filter((f) => f.path.toLowerCase().includes(query)) : files;

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
  const files = currentFiles.value;
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

const totalAdditions = computed(() => {
  return currentFiles.value.reduce((acc, f) => acc + (f.additions || 0), 0);
});

const totalDeletions = computed(() => {
  return currentFiles.value.reduce((acc, f) => acc + (f.deletions || 0), 0);
});

function collapseAllFolders() {
  const allDirs = new Set<string>();
  function collect(nodes: FileTreeNode[]) {
    for (const n of nodes) {
      if (n.isDir) {
        allDirs.add(n.path);
        collect(n.children);
      }
    }
  }
  collect(fileTreeData.value);
  collapsedFolders.value = allDirs;
}

function expandAllFolders() {
  collapsedFolders.value = new Set();
}

function handleFileClick(path: string) {
  if (selectedCommit.value?.hash === "*") {
    const isStaged = status.value.staged.some((s) => s.path === path);
    emit("open-diff", { path, staged: isStaged });
    return;
  }
  if (isComparing.value && activeCompare.value) {
    handleOpenFileDiff(path, activeCompare.value.target.hash, activeCompare.value.base.hash);
  } else if (selectedCommit.value) {
    handleOpenFileDiff(path, selectedCommit.value.hash, selectedCommit.value.parents?.[0]);
  }
}

async function triggerCompare(baseHash: string, targetHash: string) {
  isLoadingCompare.value = true;
  try {
    await fetchCompare(baseHash, targetHash);
  } finally {
    isLoadingCompare.value = false;
  }
}

async function handleSwapCompare() {
  if (!selectedCommit.value || !compareCommit.value) return;
  const temp = selectedCommit.value;
  selectedCommit.value = compareCommit.value;
  compareCommit.value = temp;
  await triggerCompare(selectedCommit.value.hash, compareCommit.value.hash);
}

function closeDetailOrCompare() {
  if (compareCommit.value) {
    compareCommit.value = null;
    activeCompare.value = null;
  } else {
    selectedCommit.value = null;
    selectedCommitDetail.value = null;
  }
}

async function handleRowClick(event: MouseEvent, commit: GitGraphCommit) {
  if (event.ctrlKey || event.metaKey) {
    if (!selectedCommit.value) {
      await handleSelectCommit(commit);
    } else if (selectedCommit.value.hash === commit.hash) {
      if (compareCommit.value) {
        selectedCommit.value = compareCommit.value;
        compareCommit.value = null;
        activeCompare.value = null;
        await handleSelectCommit(selectedCommit.value);
      } else {
        selectedCommit.value = null;
        selectedCommitDetail.value = null;
      }
    } else if (compareCommit.value?.hash === commit.hash) {
      compareCommit.value = null;
      activeCompare.value = null;
    } else {
      compareCommit.value = commit;
      await triggerCompare(selectedCommit.value.hash, commit.hash);
    }
  } else {
    if (compareCommit.value) {
      compareCommit.value = null;
      activeCompare.value = null;
    }
    await handleSelectCommit(commit);
  }
}

async function handleRepoChange(repoId: string) {
  const target = repositories.value.find((r) => r.id === repoId || r.name === repoId || r.rootPath === repoId);
  if (target) {
    try {
      isInitialLoading.value = true;
      await setActiveRepo(target);
      selectedBranches.value = ["all"];
      await fetchGraph();
    } finally {
      isInitialLoading.value = false;
    }
  }
}

onMounted(async () => {
  try {
    isInitialLoading.value = true;
    if (repositories.value.length === 0) {
      await fetchRepositories();
    }
    await Promise.allSettled([fetchStatus(), fetchBranches()]);
    await refreshGraph();
  } finally {
    isInitialLoading.value = false;
  }
  window.addEventListener("click", closeAllContextMenus);
});

onBeforeUnmount(() => {
  window.removeEventListener("click", closeAllContextMenus);
});

async function handleManualRefresh() {
  await refreshGraph(true);
}

async function refreshGraph(resetPagination = true) {
  if (resetPagination) {
    resetGraphLimit();
  }
  await fetchStatus();
  await fetchGraph({
    branch: selectedBranches.value,
    path: graphFilterPath.value,
    showStashes: showStashes.value,
  });
}

function toggleShowStashes() {
  showStashes.value = !showStashes.value;
  refreshGraph(false);
}

function handleBranchesChange(newBranches: string[]) {
  selectedBranches.value = newBranches;
  graphFilterBranches.value = newBranches;
  refreshGraph(true);
}

async function handleSelectCommit(commit: GitGraphCommit) {
  if (selectedCommit.value?.hash === commit.hash) {
    selectedCommit.value = null;
    return;
  }
  selectedCommit.value = commit;
  if (commit.hash === "*") {
    await fetchStatus();
    return;
  }
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
  const target = graphCommits.value.find((c) => c.hash.startsWith(parentHash) || parentHash.startsWith(c.hash));
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
  closeAllContextMenus();

  if (!tableWrapperRef.value || isLoadingGraph.value || isLoadingMore.value || !hasMoreCommits.value) {
    return;
  }

  const el = tableWrapperRef.value;
  const scrollBottom = el.scrollHeight - el.scrollTop - el.clientHeight;
  if (scrollBottom < 300) {
    loadMoreCommits();
  }
}

async function handleLoadMore() {
  await loadMoreCommits();
}

// Context Menu Handling
function openContextMenu(event: MouseEvent, commit: GitGraphCommit) {
  closeAllContextMenus();
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

function closeRefContextMenu() {
  refContextMenuVisible.value = false;
}

function closeAllContextMenus() {
  closeContextMenu();
  closeRefContextMenu();
}

function openRefContextMenu(event: MouseEvent, refItem: GitGraphRef, commit: GitGraphCommit) {
  closeAllContextMenus();
  selectedRef.value = refItem;
  selectedRefCommit.value = commit;
  refContextMenuPos.value = {
    x: Math.min(event.clientX, window.innerWidth - 260),
    y: Math.min(event.clientY, window.innerHeight - 380),
  };
  refContextMenuVisible.value = true;
}

async function handleContextAction(action: string) {
  if (!contextCommit.value) return;
  const commit = contextCommit.value;
  closeContextMenu();

  if (action === "compare") {
    if (comparePair.value) {
      selectedCommit.value = comparePair.value.base;
      compareCommit.value = comparePair.value.target;
      await triggerCompare(comparePair.value.base.hash, comparePair.value.target.hash);
    }
    return;
  }

  if (action === "copy-hash") {
    handleCopyHash(commit.hash);
    return;
  }

  if (action === "copy-subject") {
    try {
      await navigator.clipboard.writeText(commit.message);
      showToast(t("scm.commit.copied_subject"), "info");
    } catch {
      showToast(commit.message, "info");
    }
    return;
  }

  if (action === "stash-uncommitted") {
    openModal("stash-uncommitted", commit);
    return;
  }

  if (action === "reset-uncommitted") {
    openModal("reset-uncommitted", commit);
    return;
  }

  if (action === "clean-uncommitted") {
    openModal("clean-uncommitted", commit);
    return;
  }

  if (action === "open-scm-view") {
    emit("switch-tab", "changes");
    return;
  }

  if (action === "add-tag") {
    openModal("create-tag", commit);
    return;
  }

  if (action === "merge") {
    openRefModal("merge-branch", commit.shortHash, commit);
    return;
  }

  if (action === "rebase") {
    openRefModal("rebase-branch", commit.shortHash, commit);
    return;
  }

  openModal(action as GitModalActionType, commit);
}

async function handleRefAction(action: string) {
  const refItem = selectedRef.value;
  const commit = selectedRefCommit.value;
  closeRefContextMenu();
  if (!refItem) return;

  if (action === "copy-name") {
    try {
      await navigator.clipboard.writeText(refItem.name);
      showToast(t("scm.branch.copied", { name: refItem.name }), "info");
    } catch {
      showToast(refItem.name, "info");
    }
    return;
  }

  if (action === "archive") {
    downloadArchive(refItem.name);
    return;
  }

  if (action === "toggle-dropdown-filter") {
    if (selectedBranches.value.includes(refItem.name)) {
      selectedBranches.value = selectedBranches.value.filter((b) => b !== refItem.name);
      if (selectedBranches.value.length === 0) {
        selectedBranches.value = ["all"];
      }
    } else {
      if (selectedBranches.value.includes("all")) {
        selectedBranches.value = [refItem.name];
      } else {
        selectedBranches.value.push(refItem.name);
      }
    }
    handleBranchesChange(selectedBranches.value);
    return;
  }

  if (action === "checkout-branch" || action === "checkout-tag") {
    try {
      await checkout(refItem.name);
      await fetchBranches();
      await refreshGraph();
    } catch (err: any) {
      showToast(err.message || t("scm.graph.action_failed"), "error");
    }
    return;
  }

  if (action === "create-branch-from-tag") {
    if (commit) {
      openModal("create-branch", commit);
    }
    return;
  }

  if (action === "apply-stash") {
    if (confirm(t("scm.stash.apply_confirm", { name: refItem.name }))) {
      await stashApply(refItem.name);
    }
    return;
  }

  if (action === "pop-stash") {
    if (confirm(t("scm.stash.pop_confirm", { name: refItem.name }))) {
      await stashPop(refItem.name);
    }
    return;
  }

  if (action === "drop-stash") {
    if (confirm(t("scm.stash.drop_confirm", { name: refItem.name }))) {
      await stashDrop(refItem.name);
    }
    return;
  }

  if (action === "branch-from-stash") {
    openRefModal("create-branch", refItem.name, commit);
    return;
  }

  if (action === "checkout-remote") {
    try {
      await checkout(refItem.name);
      await fetchBranches();
      await refreshGraph();
    } catch (err: any) {
      showToast(err.message || t("scm.graph.action_failed"), "error");
    }
    return;
  }

  openRefModal(action as GitModalActionType, refItem.name, commit);
}

function openRefModal(action: GitModalActionType, refName?: string, commit?: GitGraphCommit | null) {
  activeModalType.value = action;
  modalTargetRefName.value = refName;
  modalTargetCommit.value = commit || null;
  modalVisible.value = true;
}

function openModal(action: GitModalActionType, commit: GitGraphCommit) {
  activeModalType.value = action;
  modalTargetCommit.value = commit;
  modalTargetRefName.value = undefined;
  modalVisible.value = true;
}

async function handleModalConfirm(payload: {
  action: GitModalActionType;
  commit?: GitGraphCommit | null;
  targetRefName?: string;
  data?: any;
}) {
  modalVisible.value = false;
  const { action, commit, data } = payload;

  try {
    if (action === "stash-uncommitted") {
      await stashUncommitted(data?.message, data?.includeUntracked);
      await fetchStatus();
    } else if (action === "reset-uncommitted") {
      await resetUncommitted();
      await fetchStatus();
    } else if (action === "clean-uncommitted") {
      await cleanUntracked();
      await fetchStatus();
    } else if (action === "create-tag") {
      if (commit) await createTag(data.name, commit.hash, data.message);
    } else if (action === "create-branch") {
      if (payload.targetRefName?.startsWith("stash@{")) {
        await stashBranch(data.name, payload.targetRefName);
      } else if (commit) {
        await createBranchFromCommit(data.name, commit.hash, data.checkout);
      }
    } else if (action === "rename-branch") {
      await renameBranch(data.oldName, data.newName);
    } else if (action === "delete-branch") {
      await deleteBranch(data.name, data.force);
    } else if (action === "merge-branch") {
      await mergeBranch(data.branch, data.noFf, data.squash);
    } else if (action === "rebase-branch") {
      await rebaseBranch(data.branch);
    } else if (action === "push-branch") {
      await pushBranch(data.branch, data.setUpstream, data.force);
    } else if (action === "delete-tag") {
      await deleteTag(data.name);
    } else if (action === "push-tag") {
      await pushTag(data.name, data.force);
    } else if (action === "reset") {
      if (commit) await resetToCommit(commit.hash, data.mode);
    } else if (action === "cherry-pick") {
      if (commit) await cherryPick(commit.hash);
    } else if (action === "revert") {
      if (commit) await revertCommit(commit.hash);
    } else if (action === "checkout") {
      if (commit) await checkout(commit.hash);
    }
    await fetchBranches();
    await refreshGraph();
    if (totalChanges.value === 0 && selectedCommit.value?.hash === "*") {
      selectedCommit.value = null;
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
  min-width: 140px;
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

/* ── Conventional Commit Type Selector ── */
.type-selector-item {
  display: flex;
  align-items: center;
  gap: 4px;
}

.type-select-wrapper {
  position: relative;
  display: inline-flex;
  align-items: center;
  height: 28px;
  background: var(--input-bg, rgba(255, 255, 255, 0.06));
  border: 1px solid var(--card-border, #334155);
  border-radius: 4px;
  transition: all 0.15s ease;
}

.type-select-wrapper:hover {
  border-color: var(--accent-cyan, #0ea5e9);
}

.type-select-wrapper.has-filter {
  background: var(--accent-cyan-dim, rgba(14, 165, 233, 0.15));
  border-color: var(--accent-cyan, #0ea5e9);
}

:root[data-theme="light"] .type-select-wrapper {
  background: #ffffff;
  border-color: var(--card-border, #cbd5e1);
}

:root[data-theme="light"] .type-select-wrapper.has-filter {
  background: #e0f2fe;
  border-color: #38bdf8;
}

.type-select-icon {
  position: absolute;
  left: 8px;
  font-size: 13px;
  color: var(--text-muted, #94a3b8);
  pointer-events: none;
  transition: color 0.15s ease;
}

.type-select-wrapper.has-filter .type-select-icon {
  color: var(--accent-cyan, #0ea5e9);
}

:root[data-theme="light"] .type-select-wrapper.has-filter .type-select-icon {
  color: #0369a1;
}

.graph-type-select {
  appearance: none;
  background: transparent;
  color: var(--text-main);
  border: none;
  height: 100%;
  padding: 0 24px 0 26px;
  font-size: 12px;
  cursor: pointer;
  outline: none;
  font-family: inherit;
  font-weight: 500;
}

.type-select-wrapper.has-filter .graph-type-select {
  color: var(--accent-cyan, #0ea5e9);
  font-weight: 600;
}

:root[data-theme="light"] .type-select-wrapper.has-filter .graph-type-select {
  color: #0369a1;
}

.graph-type-select option {
  background: var(--card-bg, #1e293b);
  color: var(--text-main, #f8fafc);
}

:root[data-theme="light"] .graph-type-select option {
  background: #ffffff;
  color: #1e293b;
}

.type-select-arrow {
  position: absolute;
  right: 6px;
  pointer-events: none;
  font-size: 12px;
  color: var(--text-muted);
}

.type-select-wrapper.has-filter .type-select-arrow {
  color: var(--accent-cyan, #0ea5e9);
}

:root[data-theme="light"] .type-select-wrapper.has-filter .type-select-arrow {
  color: #0369a1;
}

.clear-type-filter-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  border-radius: 50%;
  padding: 0;
  transition: all 0.15s ease;
}

.clear-type-filter-btn:hover {
  background: var(--hover-bg, rgba(255, 255, 255, 0.1));
  color: var(--accent-cyan, #0ea5e9);
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

.commit-count-pill.clickable {
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 5px;
  transition: all 0.15s ease;
  user-select: none;
}

.commit-count-pill.clickable:hover {
  background: rgba(128, 128, 128, 0.2);
  color: var(--text-main);
  border-color: var(--primary, #3b82f6);
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
  0%,
  100% {
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
  cursor: pointer;
  user-select: none;
  transition: filter 0.15s ease;
}

.ref-badge:hover {
  filter: brightness(1.2);
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

.badge-stash {
  background: rgba(236, 72, 153, 0.15);
  color: #f472b6;
  border: 1px solid rgba(236, 72, 153, 0.35);
}

:root[data-theme="light"] .badge-stash {
  color: #db2777;
}

.commit-subject {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--text-main);
  display: inline-flex;
  align-items: center;
  gap: 5px;
}

.cc-type-badge {
  font-size: 10px;
  font-weight: 600;
  font-family: var(--font-family-mono, monospace);
  padding: 1px 5px;
  border-radius: 4px;
  line-height: 1.2;
  flex-shrink: 0;
  text-transform: lowercase;
}

.cc-type-badge.cc-feat {
  background: rgba(16, 185, 129, 0.18);
  color: #34d399;
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.cc-type-badge.cc-fix {
  background: rgba(239, 68, 68, 0.18);
  color: #f87171;
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.cc-type-badge.cc-docs {
  background: rgba(59, 130, 246, 0.18);
  color: #60a5fa;
  border: 1px solid rgba(59, 130, 246, 0.3);
}

.cc-type-badge.cc-refactor {
  background: rgba(168, 85, 247, 0.18);
  color: #c084fc;
  border: 1px solid rgba(168, 85, 247, 0.3);
}

.cc-type-badge.cc-test {
  background: rgba(245, 158, 11, 0.18);
  color: #fbbf24;
  border: 1px solid rgba(245, 158, 11, 0.3);
}

.cc-type-badge.cc-perf {
  background: rgba(6, 182, 212, 0.18);
  color: #22d3ee;
  border: 1px solid rgba(6, 182, 212, 0.3);
}

.cc-type-badge.cc-revert {
  background: rgba(244, 63, 94, 0.18);
  color: #fb7185;
  border: 1px solid rgba(244, 63, 94, 0.3);
}

.cc-type-badge.cc-chore,
.cc-type-badge.cc-ci,
.cc-type-badge.cc-build,
.cc-type-badge.cc-style {
  background: rgba(148, 163, 184, 0.15);
  color: #94a3b8;
  border: 1px solid rgba(148, 163, 184, 0.25);
}

:root[data-theme="light"] .cc-type-badge.cc-feat {
  background: #ecfdf5;
  color: #047857;
  border-color: #a7f3d0;
}

:root[data-theme="light"] .cc-type-badge.cc-fix {
  background: #fef2f2;
  color: #b91c1c;
  border-color: #fecaca;
}

:root[data-theme="light"] .cc-type-badge.cc-docs {
  background: #eff6ff;
  color: #1d4ed8;
  border-color: #bfdbfe;
}

:root[data-theme="light"] .cc-type-badge.cc-refactor {
  background: #faf5ff;
  color: #7e22ce;
  border-color: #e9d5ff;
}

:root[data-theme="light"] .cc-type-badge.cc-test {
  background: #fffbeb;
  color: #b45309;
  border-color: #fde68a;
}

.cc-scope-badge {
  font-size: 10px;
  font-family: var(--font-family-mono, monospace);
  color: var(--text-muted);
  background: rgba(255, 255, 255, 0.05);
  padding: 0 4px;
  border-radius: 3px;
  flex-shrink: 0;
}

.cc-breaking-badge {
  font-size: 10px;
  font-weight: 700;
  padding: 0 4px;
  border-radius: 3px;
  background: #ef4444;
  color: #ffffff;
  flex-shrink: 0;
}

.cc-desc-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
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

/* Loading State (Initial / Empty Graph) */
.graph-loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px 20px;
  gap: 16px;
  color: var(--text-main);
  animation: graphFadeIn 0.2s ease-out;
}

.graph-loading-spinner-wrapper {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background: var(--accent-cyan-glow, rgba(0, 229, 255, 0.12));
  border: 1px solid var(--card-border);
}

.graph-loading-spinner-wrapper .loading-spinner {
  font-size: 24px;
  color: var(--accent-cyan);
}

.loading-text-group {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
}

.loading-primary-text {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-main);
}

.loading-secondary-text {
  font-size: 12px;
  color: var(--text-muted);
}

/* Top Indeterminate Progress Line */
.graph-loading-progressbar {
  position: sticky;
  top: 32px;
  left: 0;
  right: 0;
  height: 2px;
  background: transparent;
  overflow: hidden;
  z-index: 20;
  pointer-events: none;
}

.loading-progress-indeterminate {
  height: 100%;
  width: 40%;
  background: linear-gradient(90deg, transparent, var(--accent-cyan), transparent);
  animation: progressIndeterminate 1.2s infinite cubic-bezier(0.4, 0, 0.2, 1);
}

@keyframes progressIndeterminate {
  0% {
    transform: translateX(-100%);
  }
  100% {
    transform: translateX(350%);
  }
}

/* Floating Refreshing Badge */
.graph-refreshing-badge {
  position: absolute;
  top: 40px;
  right: 16px;
  z-index: 25;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 6px 14px;
  font-size: 12px;
  color: var(--accent-cyan);
  background: var(--modal-bg, var(--card-bg));
  border: 1px solid var(--card-border);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.25);
  border-radius: 20px;
  backdrop-filter: blur(8px);
  animation: graphBadgeFadeIn 0.2s ease-out;
  pointer-events: none;
}

@keyframes graphFadeIn {
  from {
    opacity: 0;
    transform: translateY(6px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes graphBadgeFadeIn {
  from {
    opacity: 0;
    transform: translateY(-4px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* Table body when updating in background */
.graph-table-body.table-is-refreshing {
  opacity: 0.65;
  transition: opacity 0.25s ease;
  pointer-events: none;
}

/* Bottom Pagination / Load More Row */
.graph-load-more-row {
  position: absolute;
  left: 0;
  right: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 16px;
  box-sizing: border-box;
  border-top: 1px dashed var(--card-border);
  background: var(--card-bg, transparent);
  z-index: 5;
}

.load-more-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  border-radius: 4px;
  padding: 6px 18px;
  font-size: 12px;
  font-weight: 500;
  color: var(--text-main);
  cursor: pointer;
  transition: all 0.15s ease;
}

.load-more-btn:hover:not(:disabled) {
  background: var(--active-bg, rgba(255, 255, 255, 0.08));
  border-color: var(--primary, #3b82f6);
  color: var(--primary, #3b82f6);
}

.load-more-btn:disabled {
  opacity: 0.65;
  cursor: not-allowed;
}

.all-commits-loaded-hint {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--text-muted);
  padding: 6px 12px;
  user-select: none;
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
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.inline-message-top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 8px;
}

.inline-subject-row {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
  flex: 1;
}

.cc-type-badge.large {
  font-size: 11px;
  padding: 2px 7px;
  border-radius: 4px;
}

.cc-scope-badge.large {
  font-size: 11px;
  padding: 1px 6px;
}

.cc-breaking-badge.large {
  font-size: 11px;
  padding: 2px 6px;
}

.inline-subject-title {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-main);
  line-height: 1.4;
}

.inline-copy-msg-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 3px 5px;
  border-radius: 4px;
  display: inline-flex;
  transition: all 0.15s ease;
  flex-shrink: 0;
}

.inline-copy-msg-btn:hover {
  background: rgba(128, 128, 128, 0.15);
  color: var(--accent-cyan);
}

.inline-body-content {
  border-top: 1px dashed var(--card-border);
  padding-top: 6px;
}

.inline-body-text {
  margin: 0;
  font-family: inherit;
  font-size: 11px;
  line-height: 1.5;
  white-space: pre-wrap;
  color: var(--text-muted);
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

.inline-files-count-group {
  display: flex;
  align-items: center;
  gap: 8px;
}

.inline-files-count {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-main);
}

.inline-files-stats-pill {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 10px;
  font-family: var(--font-family-mono, monospace);
  padding: 1px 6px;
  border-radius: 10px;
  background: var(--bg-secondary);
  border: 1px solid var(--card-border);
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
  background: #081a24;
  border: 1px solid #143547;
  border-radius: 4px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.55);
  padding: 4px 0;
  min-width: 230px;
  z-index: 2000;
  display: flex;
  flex-direction: column;
}

:root[data-theme="light"] .graph-context-menu {
  background: #ffffff;
  border-color: #cbd5e1;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
}

.menu-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 5px 16px;
  font-size: 13px;
  color: #e2e8f0;
  cursor: pointer;
  transition:
    background-color 0.08s ease,
    color 0.08s ease;
  white-space: nowrap;
}

.menu-item:hover {
  background: #005f73;
  color: #ffffff;
}

:root[data-theme="light"] .menu-item {
  color: #334155;
}

:root[data-theme="light"] .menu-item:hover {
  background: #005f73;
  color: #ffffff;
}

.menu-item.danger:hover {
  background: #dc2626;
  color: #ffffff;
}

.menu-separator {
  height: 1px;
  background: #183745;
  margin: 4px 0;
}

:root[data-theme="light"] .menu-separator {
  background: #e2e8f0;
}

/* Compare Mode Styles */
.inline-compare-title {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  font-weight: 600;
  color: var(--accent-cyan, #00e5ff);
}

.compare-stats-card {
  padding: 8px 12px;
  margin-bottom: 10px;
  background: rgba(0, 229, 255, 0.05);
  border: 1px solid rgba(0, 229, 255, 0.2);
  border-radius: 4px;
}

.compare-stats-header {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  font-weight: 600;
  color: var(--accent-cyan, #00e5ff);
  margin-bottom: 4px;
}

.compare-stats-pill {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
}

.stat-divider {
  color: var(--text-muted);
}

.compare-commit-box {
  padding: 8px 12px;
  background: rgba(128, 128, 128, 0.06);
  border: 1px solid var(--card-border);
  border-radius: 4px;
  margin-bottom: 6px;
}

.compare-box-label {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 4px;
  font-size: 11px;
}

.box-badge {
  padding: 1px 5px;
  border-radius: 3px;
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
}

.box-badge.base-badge {
  background: rgba(255, 171, 0, 0.15);
  color: #ffab00;
  border: 1px solid rgba(255, 171, 0, 0.3);
}

.box-badge.target-badge {
  background: rgba(0, 230, 118, 0.15);
  color: #00e676;
  border: 1px solid rgba(0, 230, 118, 0.3);
}

.hash-link {
  cursor: pointer;
  color: var(--accent-cyan);
  display: flex;
  align-items: center;
  gap: 4px;
}

.hash-link:hover {
  text-decoration: underline;
}

.compare-box-subject {
  font-size: 12px;
  color: var(--text-main);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  margin-bottom: 4px;
  font-weight: 500;
}

.compare-box-meta {
  font-size: 11px;
  color: var(--text-muted);
  display: flex;
  align-items: center;
  gap: 4px;
}

.compare-arrow-divider {
  display: flex;
  justify-content: center;
  margin: 2px 0 6px;
  color: var(--text-muted);
  font-size: 12px;
}

.graph-context-menu .menu-item.primary-action {
  font-weight: 600;
  color: var(--accent-cyan, #00e5ff);
}

/* Uncommitted Changes Node and Row Styles */
.node-uncommitted-ring {
  stroke-dasharray: 2 1.5;
  filter: drop-shadow(0 0 3px rgba(255, 171, 0, 0.4));
  transition: all 0.2s ease;
}

.graph-path-line.uncommitted-path {
  stroke-dasharray: 3 3;
  opacity: 0.85;
}

.graph-row.uncommitted-row {
  font-style: normal;
}

.graph-row.uncommitted-row .commit-subject.uncommitted-title {
  color: #ffb74d;
  font-weight: 600;
}

.uncommitted-circle-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  line-height: 1;
  color: #ffb74d;
  margin-right: 6px;
  font-weight: 700;
}

.author-star {
  display: inline-block;
  color: var(--text-muted);
  font-size: 14px;
  line-height: 1;
}

/* Uncommitted Meta Card */
.uncommitted-meta-card {
  padding: 12px 14px;
  background: rgba(255, 171, 0, 0.05);
  border: 1px solid rgba(255, 171, 0, 0.25);
  border-radius: 6px;
  margin-bottom: 12px;
}

:root[data-theme="light"] .uncommitted-meta-card {
  background: rgba(245, 158, 11, 0.08);
  border-color: rgba(245, 158, 11, 0.3);
}

.uncommitted-card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  font-weight: 600;
  color: #ffb74d;
  margin-bottom: 10px;
}

:root[data-theme="light"] .uncommitted-card-header {
  color: #b45309;
}

.uncommitted-stats-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
  font-size: 11px;
}

.uncommitted-stat-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.uncommitted-stat-row .stat-label {
  color: var(--text-muted);
}

.uncommitted-stat-row .stat-num {
  font-weight: 600;
  padding: 1px 6px;
  border-radius: 3px;
}

.uncommitted-stat-row .stat-num.staged {
  background: rgba(34, 197, 94, 0.15);
  color: #4ade80;
}

.uncommitted-stat-row .stat-num.unstaged {
  background: rgba(234, 179, 8, 0.15);
  color: #facc15;
}

.uncommitted-stat-row.total-row {
  margin-top: 4px;
  padding-top: 6px;
  border-top: 1px dashed rgba(255, 255, 255, 0.1);
}

.uncommitted-stat-row .stat-num.total {
  background: rgba(56, 189, 248, 0.15);
  color: #38bdf8;
}

.uncommitted-card-hint {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 10px;
  font-size: 11px;
  color: var(--text-muted);
}

/* ── Repo Selector ── */
.repo-selector-item {
  display: flex;
  align-items: center;
  gap: 6px;
}

.toolbar-label {
  font-size: 12px;
  color: var(--text-muted);
  font-weight: 500;
}

.repo-select-wrapper {
  position: relative;
  display: inline-flex;
  align-items: center;
}

.graph-repo-select {
  appearance: none;
  background: var(--input-bg, rgba(255, 255, 255, 0.06));
  color: var(--text-main);
  border: 1px solid var(--border-color);
  border-radius: 4px;
  padding: 4px 24px 4px 8px;
  font-size: 12px;
  cursor: pointer;
  outline: none;
  font-family: inherit;
  font-weight: 500;
}

.graph-repo-select:focus {
  border-color: var(--accent-color, #3b82f6);
}

.repo-select-arrow {
  position: absolute;
  right: 6px;
  pointer-events: none;
  font-size: 12px;
  color: var(--text-muted);
}
</style>
