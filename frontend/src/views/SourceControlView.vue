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
          :title="`${t('scm.diff.commit_badge')}: ${commitHash}`"
        >
          {{ t('scm.diff.commit_badge') }} {{ commitHash.slice(0, 7) }}{{ parentHash ? ' vs ' + parentHash.slice(0, 7) : '' }}
        </span>
        <span
          v-else
          class="diff-status-pill"
          :class="isStaged ? 'pill-staged' : 'pill-unstaged'"
        >
          {{ isStaged ? t('scm.diff.staged_badge') : t('scm.diff.unstaged_badge') }}
        </span>
      </div>

      <div class="diff-actions">
        <!-- Toggle Inline vs Side-by-Side Diff -->
        <button
          class="diff-btn"
          :title="isInlineDiff ? t('scm.diff.toggle_side_by_side') : t('scm.diff.toggle_inline')"
          @click="toggleInlineDiff"
        >
          <span class="codicon" :class="isInlineDiff ? 'codicon-split-horizontal' : 'codicon-list-flat'"></span>
          <span>{{ isInlineDiff ? t('scm.diff.side_by_side') : t('scm.diff.inline') }}</span>
        </button>

        <!-- Working Tree Stage / Unstage / Discard Actions (only when not commit diff) -->
        <template v-if="!isCommitDiff">
          <!-- Stage / Unstage Action -->
          <button
            v-if="!isStaged"
            class="diff-btn primary"
            :title="t('scm.diff.stage_file')"
            @click="handleStageActiveFile"
          >
            <span class="codicon codicon-add"></span>
            <span>{{ t('scm.diff.stage_file') }}</span>
          </button>
          <button
            v-else
            class="diff-btn"
            :title="t('scm.diff.unstage_file')"
            @click="handleUnstageActiveFile"
          >
            <span class="codicon codicon-remove"></span>
            <span>{{ t('scm.diff.unstage_file') }}</span>
          </button>

          <!-- Discard Action (unstaged only) -->
          <button
            v-if="!isStaged"
            class="diff-btn danger"
            :title="t('scm.diff.discard_file')"
            @click="handleDiscardActiveFile"
          >
            <span class="codicon codicon-discard"></span>
            <span>{{ t('scm.diff.discard_file') }}</span>
          </button>
        </template>

        <!-- Open File in Editor -->
        <button
          class="diff-btn"
          :title="t('scm.diff.open_file')"
          @click="handleOpenFileInEditor"
        >
          <span class="codicon codicon-go-to-file"></span>
          <span>{{ t('scm.diff.open_file') }}</span>
        </button>

        <!-- Reveal File in Explorer -->
        <button
          class="diff-btn"
          title="在资源管理器中揭示"
          @click="handleRevealFileInExplorer(activeDiffFile)"
        >
          <span class="codicon codicon-folder-opened"></span>
          <span>定位</span>
        </button>

        <!-- Close Diff / Back to Overview or Graph -->
        <button
          class="diff-btn close-btn"
          :title="t('scm.diff.close')"
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
              <h2 class="repo-title">{{ t('scm.title') }}</h2>
              <p class="repo-subtitle">{{ t('scm.subtitle') }}</p>
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
                <span>{{ t('scm.tabs.changes') }} ({{ totalChanges }})</span>
              </button>
              <button
                class="mode-tab-btn"
                :class="{ active: currentTab === 'graph' }"
                @click="switchTab('graph')"
              >
                <span class="codicon codicon-git-commit"></span>
                <span>{{ t('scm.tabs.graph') }}</span>
              </button>
              <button
                class="mode-tab-btn"
                :class="{ active: currentTab === 'guard' }"
                @click="switchTab('guard')"
              >
                <span class="codicon codicon-shield"></span>
                <span>{{ t('guard.title') || '质量守卫' }}</span>
                <span
                  v-if="guardIssuesCount > 0"
                  class="tab-count-pill"
                  :class="guardStatus.errors_count > 0 ? 'pill-error' : 'pill-warning'"
                >
                  {{ guardIssuesCount }}
                </span>
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
              <span class="stat-label">{{ t('scm.branch.active') }}</span>
            </div>
          </div>

          <div class="stat-card">
            <div class="stat-icon-wrapper green">
              <span class="codicon codicon-check-all"></span>
            </div>
            <div class="stat-details">
              <span class="stat-value">{{ status.staged.length }}</span>
              <span class="stat-label">{{ t('scm.changes.staged') }}</span>
            </div>
          </div>

          <div class="stat-card">
            <div class="stat-icon-wrapper amber">
              <span class="codicon codicon-diff-modified"></span>
            </div>
            <div class="stat-details">
              <span class="stat-value">{{ status.unstaged.length }}</span>
              <span class="stat-label">{{ t('scm.changes.unstaged') }}</span>
            </div>
          </div>

          <div class="stat-card">
            <div class="stat-icon-wrapper purple">
              <span class="codicon codicon-history"></span>
            </div>
            <div class="stat-details">
              <span class="stat-value">{{ totalChanges }}</span>
              <span class="stat-label">{{ t('scm.changes.total') }}</span>
            </div>
          </div>

          <div class="stat-card clickable-guard-card" :title="guardTooltip" @click="switchTab('guard')">
            <div class="stat-icon-wrapper" :class="guardCardIconColor">
              <span class="codicon" :class="guardCardIcon"></span>
            </div>
            <div class="stat-details">
              <span class="stat-value">{{ guardCardValue }}</span>
              <span class="stat-label">{{ t('guard.title') || '质量守卫' }}</span>
            </div>
          </div>
        </div>

        <!-- Quick Actions Bar -->
        <div class="overview-actions-card">
          <div class="actions-group">
            <button
              v-if="guardStatus.fixable_count > 0"
              class="action-pill warning"
              :disabled="isGuardFixing"
              @click="handleRunGuardFix"
            >
              <span class="codicon" :class="isGuardFixing ? 'codicon-loading codicon-modifier-spin' : 'codicon-wand'"></span>
              <span>{{ t('guard.actions.auto_fix') || '一键自动修复' }} ({{ guardStatus.fixable_count }})</span>
            </button>
            <button
              class="action-pill primary"
              :disabled="status.unstaged.length === 0"
              @click="stage([])"
            >
              <span class="codicon codicon-add"></span>
              <span>{{ t('scm.actions.stage_all') }}</span>
            </button>
            <button
              class="action-pill"
              :disabled="status.staged.length === 0"
              @click="unstage([])"
            >
              <span class="codicon codicon-remove"></span>
              <span>{{ t('scm.actions.unstage_all') }}</span>
            </button>
            <button
              class="action-pill"
              @click="handleManualRefresh"
            >
              <span class="codicon codicon-refresh"></span>
              <span>{{ t('scm.actions.refresh') }}</span>
            </button>
            <button
              class="action-pill"
              :title="t('scm.actions.more')"
              @click.stop="toggleMoreMenu"
            >
              <span class="codicon codicon-ellipsis"></span>
              <span>{{ t('scm.actions.more') }}</span>
            </button>
          </div>
        </div>

        <!-- Split Columns: Staged and Unstaged Files -->
        <div class="changes-split-view">
          <!-- Staged Box -->
          <div class="changes-box">
            <div class="box-header">
              <div class="box-header-left">
                <span class="box-title">{{ t('scm.staged.title') }} ({{ status.staged.length }})</span>
                <span v-if="status.staged.length > 0" class="box-badge green">{{ t('scm.diff.staged_badge') }}</span>
              </div>
              <div v-if="status.staged.length > 0" class="box-header-actions" @click.stop>
                <button
                  class="box-header-btn"
                  :title="t('scm.actions.unstage_all')"
                  @click="unstage([])"
                >
                  <span class="codicon codicon-remove"></span>
                  <span>{{ t('scm.actions.unstage_all') }}</span>
                </button>
              </div>
            </div>
            <div class="box-content">
              <div
                v-for="item in status.staged"
                :key="'staged:' + item.path"
                class="box-file-item"
                :title="item.path"
                @click="openDiff(item.path, true)"
                @contextmenu.prevent="openFileContextMenu($event, item, true)"
              >
                <div class="file-item-main">
                  <TechIcon :id="getFileName(item.path)" :size="16" />
                  <span class="file-name">{{ getFileName(item.path) }}</span>
                  <span class="file-dir">{{ getDirName(item.path) }}</span>
                </div>
                <div class="file-item-right">
                  <span class="status-tag" :class="'tag-' + item.status">{{ item.status }}</span>
                  <div class="item-hover-actions" @click.stop>
                    <button
                      class="file-action-icon-btn"
                      title="在资源管理器中揭示"
                      @click="handleRevealFileInExplorer(item.path)"
                    >
                      <span class="codicon codicon-folder-opened"></span>
                    </button>
                    <button
                      class="file-action-icon-btn"
                      :title="t('scm.diff.open_file')"
                      @click="handleOpenFile(item.path)"
                    >
                      <span class="codicon codicon-go-to-file"></span>
                    </button>
                    <button
                      class="file-action-icon-btn"
                      :title="t('scm.diff.unstage_file')"
                      @click="handleUnstageSingleFile(item.path)"
                    >
                      <span class="codicon codicon-remove"></span>
                    </button>
                  </div>

                </div>
              </div>
              <div v-if="status.staged.length === 0" class="empty-state">
                <span class="codicon codicon-inbox"></span>
                <span>{{ t('scm.staged.empty') }}</span>
              </div>
            </div>
          </div>

          <!-- Unstaged Box -->
          <div class="changes-box">
            <div class="box-header">
              <div class="box-header-left">
                <span class="box-title">{{ t('scm.unstaged.title') }} ({{ status.unstaged.length }})</span>
              </div>
              <div v-if="status.unstaged.length > 0" class="box-header-actions" @click.stop>
                <button
                  class="box-header-btn danger"
                  :title="t('scm.actions.discard_all')"
                  @click="handleDiscardAllUnstaged"
                >
                  <span class="codicon codicon-discard"></span>
                  <span>{{ t('scm.actions.discard_all') }}</span>
                </button>
                <button
                  class="box-header-btn primary"
                  :title="t('scm.actions.stage_all')"
                  @click="stage([])"
                >
                  <span class="codicon codicon-add"></span>
                  <span>{{ t('scm.actions.stage_all') }}</span>
                </button>
              </div>
            </div>
            <div class="box-content">
              <div
                v-for="item in status.unstaged"
                :key="'unstaged:' + item.path"
                class="box-file-item"
                :title="item.path"
                @click="openDiff(item.path, false)"
                @contextmenu.prevent="openFileContextMenu($event, item, false)"
              >
                <div class="file-item-main">
                  <TechIcon :id="getFileName(item.path)" :size="16" />
                  <span class="file-name">{{ getFileName(item.path) }}</span>
                  <span class="file-dir">{{ getDirName(item.path) }}</span>
                </div>
                <div class="file-item-right">
                  <span class="status-tag" :class="'tag-' + item.status">{{ item.status }}</span>
                  <div class="item-hover-actions" @click.stop>
                    <button
                      class="file-action-icon-btn"
                      title="在资源管理器中揭示"
                      @click="handleRevealFileInExplorer(item.path)"
                    >
                      <span class="codicon codicon-folder-opened"></span>
                    </button>
                    <button
                      class="file-action-icon-btn"
                      :title="t('scm.diff.open_file')"
                      @click="handleOpenFile(item.path)"
                    >
                      <span class="codicon codicon-go-to-file"></span>
                    </button>
                    <button
                      class="file-action-icon-btn danger"
                      :title="t('scm.diff.discard_file')"
                      @click="handleDiscardSingleFile(item.path)"
                    >
                      <span class="codicon codicon-discard"></span>
                    </button>
                    <button
                      class="file-action-icon-btn primary"
                      :title="t('scm.diff.stage_file')"
                      @click="handleStageSingleFile(item.path)"
                    >
                      <span class="codicon codicon-add"></span>
                    </button>
                  </div>

                </div>
              </div>
              <div v-if="status.unstaged.length === 0" class="empty-state">
                <span class="codicon codicon-pass"></span>
                <span>{{ t('scm.unstaged.empty') }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Recent Commits Section -->
        <div class="commits-history-card">
          <div class="box-header">
            <div class="title-with-icon">
              <span class="codicon codicon-history"></span>
              <span class="box-title">{{ t('scm.history.title') }}</span>
            </div>
            <button class="action-pill small" @click="fetchLog">
              <span class="codicon codicon-refresh"></span>
              <span>{{ t('common.refresh') }}</span>
            </button>
          </div>
          <div class="commits-list">
            <div v-for="c in commits" :key="c.hash" class="commit-item">
              <span class="commit-hash font-mono">{{ c.shortHash }}</span>
              <span
                v-if="c.gpg && c.gpg.status !== 'none'"
                class="gpg-mini-badge"
                :class="'gpg-' + c.gpg.status"
                :title="c.gpg.status === 'verified' ? (t('scm.gpg.verified') + ' • ' + (c.gpg.signer || c.gpg.keyId || '')) : (t('scm.gpg.unverified') + ': ' + c.gpg.status)"
              >
                <span class="codicon" :class="c.gpg.status === 'verified' ? 'codicon-verified' : 'codicon-shield'"></span>
                <span>{{ c.gpg.status === 'verified' ? t('scm.gpg.verified') : c.gpg.status }}</span>
              </span>
              <span class="commit-message">{{ c.message }}</span>
              <div class="commit-meta">
                <span class="commit-author">{{ c.author }}</span>
                <span class="commit-time">{{ formatTime(c.timestamp) }}</span>
              </div>
            </div>
            <div v-if="commits.length === 0" class="empty-state">
              <span class="codicon codicon-history"></span>
              <span>{{ t('scm.history.empty') }}</span>
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
        @open-diff="handleOpenDiffFromGraph"
        @switch-tab="switchTab"
      />

      <!-- Mode 3: VS Code Grade Quality Guard Hub -->
      <div v-else-if="currentTab === 'guard'" class="guard-hub-view">
        <div class="guard-hub-header">
          <div class="guard-hub-title-group">
            <div class="guard-badge-large" :class="guardCardIconColor">
              <span class="codicon" :class="guardCardIcon"></span>
            </div>
            <div class="guard-hub-title-text">
              <h2 class="guard-hub-title">{{ t('guard.title') || '质量守卫与合规中心' }}</h2>
              <p class="guard-hub-desc">
                Pre-commit & CI 双守卫流水线：Phase 1 Fixer 原生自愈 ➔ Phase 2 Formatter 格式化 ➔ Phase 3 Linter & Security 门禁阻断与密钥防泄漏
              </p>
            </div>
          </div>
          <div class="guard-hub-actions">
            <div class="guard-scope-toggle">
              <button
                class="scope-btn"
                :class="{ active: currentGuardScope === 'staged' }"
                @click="handleChangeGuardScope('staged')"
              >
                {{ t('scm.sidebar.staged') || '暂存区' }} ({{ status.staged.length }})
              </button>
              <button
                class="scope-btn"
                :class="{ active: currentGuardScope === 'all' }"
                @click="handleChangeGuardScope('all')"
              >
                {{ t('scm.changes.total') || '所有变更' }} ({{ totalChanges }})
              </button>
            </div>
            <button
              v-if="guardStatus.fixable_count > 0"
              class="action-pill warning"
              :disabled="isGuardFixing"
              @click="handleRunGuardFix"
            >
              <span class="codicon" :class="isGuardFixing ? 'codicon-loading codicon-modifier-spin' : 'codicon-wand'"></span>
              <span>{{ t('guard.actions.auto_fix') || '一键自动修复' }} ({{ guardStatus.fixable_count }})</span>
            </button>
            <button
              class="action-pill primary"
              :disabled="isGuardLoading"
              @click="handleRefreshGuard"
            >
              <span class="codicon codicon-refresh" :class="{ 'codicon-modifier-spin': isGuardLoading }"></span>
              <span>{{ t('guard.actions.rescan') || '重新扫描' }}</span>
            </button>
          </div>
        </div>

        <!-- 3-Phase Pipeline Architecture Cards -->
        <div class="guard-pipeline-grid">
          <div class="pipeline-step-card">
            <div class="step-header">
              <span class="step-badge">Phase 1</span>
              <span class="step-title">Fixer (自愈修复)</span>
            </div>
            <div class="step-desc">
              Go 原生极速自愈：末尾换行符补全、行尾多余空白清除、统一换行符、行尾空行收敛。
            </div>
            <div class="step-status">
              <span class="codicon codicon-check text-green"></span>
              <span>内置 32 项 Go 原生规则</span>
            </div>
          </div>

          <div class="pipeline-arrow">➔</div>

          <div class="pipeline-step-card">
            <div class="step-header">
              <span class="step-badge">Phase 2</span>
              <span class="step-title">Formatter (代码格式化)</span>
            </div>
            <div class="step-desc">
              标杆格式化工具插槽：gofumpt (Go)、prettier (Web/TS/CSS/JSON)、shfmt (Shell)。
            </div>
            <div class="step-status">
              <span class="codicon codicon-check text-green"></span>
              <span>自动识别语言格式化</span>
            </div>
          </div>

          <div class="pipeline-arrow">➔</div>

          <div class="pipeline-step-card">
            <div class="step-header">
              <span class="step-badge">Phase 3</span>
              <span class="step-title">Linter & Security (门禁阻断)</span>
            </div>
            <div class="step-desc">
              静态代码门禁与安全卡点：gitleaks 密钥敏感凭证防泄漏扫描 + golangci-lint。
            </div>
            <div class="step-status">
              <span class="codicon codicon-shield text-blue"></span>
              <span>Pre-commit & CI 双守卫</span>
            </div>
          </div>
        </div>

        <!-- Diagnostics Panel (VS Code Problems Explorer) -->
        <div class="guard-diagnostics-section">
          <div class="diagnostics-bar">
            <div class="diagnostics-summary">
              <span class="diag-title">{{ t('guard.diagnostics.title') || '质量守卫诊断结果' }}</span>
              <span class="diag-badge-clean" v-if="guardStatus.clean">🟢 {{ t('guard.status.passed') || '质量守卫通过' }}</span>
              <span class="diag-badge-errors" v-if="guardStatus.errors_count > 0">🔴 {{ guardStatus.errors_count }} 阻断违规</span>
              <span class="diag-badge-warnings" v-if="guardStatus.warnings_count > 0">🟡 {{ guardStatus.warnings_count }} 待优化项</span>
            </div>
            <span class="diag-files-count">已扫描 {{ guardStatus.scanned }} 个文件</span>
          </div>

          <!-- Issues Table / List -->
          <div v-if="guardStatus.issues.length > 0" class="guard-issues-table">
            <div
              v-for="(issue, idx) in guardStatus.issues"
              :key="idx"
              class="issue-row"
              :class="'severity-' + issue.severity"
              @click="handleOpenIssueFile(issue.file_path)"
            >
              <div class="issue-left">
                <span class="codicon" :class="issue.severity === 'error' ? 'codicon-error' : 'codicon-warning'"></span>
                <span class="issue-file font-mono">{{ issue.file_path }}<span v-if="issue.line" class="issue-line">:{{ issue.line }}</span></span>
                <span class="issue-rule font-mono">{{ issue.rule_id }}</span>
              </div>
              <div class="issue-msg">{{ issue.message }}</div>
              <div class="issue-right">
                <span v-if="issue.fixable" class="issue-fixable-pill">⚡ {{ t('guard.fixable') || '可自愈' }}</span>
                <span class="issue-phase-tag font-mono">{{ issue.phase }}</span>
              </div>
            </div>
          </div>

          <!-- Clean State Illustration -->
          <div v-else class="guard-clean-state">
            <span class="codicon codicon-shield-check clean-icon"></span>
            <h3 class="clean-title">{{ t('guard.status.passed') || '质量守卫已通过' }}</h3>
            <p class="clean-subtitle">
              当前扫描范围内未发现任何格式违规、代码坏味道或泄密风险，符合提交规范。
            </p>
          </div>
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
import { ref, computed, onMounted, onBeforeUnmount, watch, nextTick } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useGit } from "../composables/useGit";
import { useEditorTabs } from "../composables/useEditorTabs";
import { useWorkspaceBus } from "../composables/useWorkspaceBus";
import { useTheme } from "../composables/useTheme";

import { loadMonaco, getLanguageForFile } from "../services/editor/monacoLoader";
import TechIcon from "../components/common/TechIcon.vue";
import GitGraphView from "../components/scm/GitGraphView.vue";
import GitMoreActionsMenu from "../components/scm/GitMoreActionsMenu.vue";
import { useGuard } from "../composables/useGuard";
import GitFileContextMenu, { type ContextMenuTargetFile, type FileMenuActionType } from "../components/scm/GitFileContextMenu.vue";
import GitActionModal, { type GitModalActionType } from "../components/scm/GitActionModal.vue";
import { useToast } from "../composables/useToast";
import { t } from "../i18n";

const route = useRoute();
const router = useRouter();
const { isDarkTheme } = useTheme();
const { showToast } = useToast();

const {
  status,
  currentBranch,
  totalChanges,
  commits,
  fetchStatus,
  fetchBranches,
  fetchDiff,
  fetchCommitDiff,
  fetchLog,
  stage,
  unstage,
  discard,
  commit,
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

const currentTab = ref<"changes" | "graph" | "guard">(
  route.query.tab === "graph" ? "graph" : route.query.tab === "guard" ? "guard" : "changes"
);

const {
  status: guardStatus,
  isLoading: isGuardLoading,
  isFixing: isGuardFixing,
  fetchStatus: fetchGuardStatus,
  runFix: runGuardFix,
} = useGuard();

const currentGuardScope = ref<"staged" | "all">("staged");

const guardIssuesCount = computed(() => {
  return (guardStatus.value.errors_count || 0) + (guardStatus.value.warnings_count || 0);
});

async function handleChangeGuardScope(scope: "staged" | "all") {
  currentGuardScope.value = scope;
  await fetchGuardStatus(scope);
}

async function handleRefreshGuard() {
  await fetchGuardStatus(currentGuardScope.value);
}

const guardCardValue = computed(() => {
  if (isGuardLoading.value) return "...";
  if (guardStatus.value.errors_count > 0) return `${guardStatus.value.errors_count} 阻断`;
  if (guardStatus.value.fixable_count > 0) return `${guardStatus.value.fixable_count} 可修复`;
  return "已就绪";
});

const guardCardIcon = computed(() => {
  if (isGuardLoading.value || isGuardFixing.value) return "codicon-loading codicon-modifier-spin";
  if (guardStatus.value.errors_count > 0) return "codicon-shield-x";
  if (guardStatus.value.warnings_count > 0) return "codicon-shield-alert";
  return "codicon-shield-check";
});

const guardCardIconColor = computed(() => {
  if (guardStatus.value.errors_count > 0) return "red";
  if (guardStatus.value.warnings_count > 0) return "amber";
  return "green";
});

const guardTooltip = computed(() => {
  if (guardStatus.value.fixable_count > 0) {
    return `质量守卫发现 ${guardStatus.value.fixable_count} 处可修复问题，点击查看守卫中心`;
  }
  return "质量守卫通过，点击打开守卫中心";
});

async function handleRunGuardFix() {
  const ok = await runGuardFix(currentGuardScope.value);
  if (ok) {
    showToast(t("guard.toast.fixed") || "✨ 已自动完成代码格式与卫生修复", "success");
    await fetchStatus();
  }
}

function handleOpenIssueFile(path: string) {
  const fileName = getFileName(path);
  openTab({ path, name: fileName });
  router.push({ name: "explorer" }).catch(() => {});
}

const isMoreMenuOpen = ref(false);
const moreMenuPos = ref({ x: 0, y: 0 });

const isFileContextMenuOpen = ref(false);
const fileContextMenuPos = ref({ x: 0, y: 0 });
const fileContextMenuTarget = ref<ContextMenuTargetFile | null>(null);

const modalVisible = ref(false);
const activeModalType = ref<GitModalActionType | null>(null);
const modalTargetRefName = ref<string | undefined>(undefined);

function toggleMoreMenu(event: MouseEvent) {
  if (isMoreMenuOpen.value) {
    isMoreMenuOpen.value = false;
    return;
  }
  const btn = event.currentTarget as HTMLElement | null;
  if (btn) {
    const rect = btn.getBoundingClientRect();
    moreMenuPos.value = {
      x: rect.left,
      y: rect.bottom + 4,
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
    switchTab("graph");
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
    switchTab("changes");
    return;
  }

  if (action === "fetch") {
    await fetchRemote();
    return;
  }

  if (action === "commit") {
    await commit();
    return;
  }

  if (action === "commit-amend") {
    await commit(undefined, true);
    return;
  }

  if (action === "commit-staged") {
    await commit();
    return;
  }

  if (action === "commit-all") {
    await stage([]);
    await commit();
    return;
  }

  if (action === "stage-all") {
    await stage([]);
    return;
  }

  if (action === "unstage-all") {
    await unstage([]);
    return;
  }

  if (action === "discard-all") {
    const allPaths = status.value.unstaged.map((i) => i.path);
    if (confirm(t("scm.diff.discard_confirm", { path: "*" }))) {
      await discard(allPaths);
    }
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

function switchTab(tab: "changes" | "graph" | "guard") {
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
const { revealInExplorer } = useWorkspaceBus();


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
  if (confirm(t("scm.diff.discard_confirm", { path: activeDiffFile.value }))) {
    await discard([activeDiffFile.value]);
    closeDiffView();
  }
}

function handleOpenFile(targetPath?: string) {
  const p = targetPath || activeDiffFile.value;
  if (!p) return;
  const fileName = getFileName(p);
  openTab({ path: p, name: fileName });
  router.push({ name: "explorer" }).catch(() => {});
}

function handleRevealFileInExplorer(targetPath?: string) {
  const p = targetPath || activeDiffFile.value;
  if (!p) return;
  revealInExplorer(p);
  router.push({ name: "explorer" }).catch(() => {});
}

function handleOpenFileInEditor() {
  handleOpenFile();
}

async function handleStageSingleFile(path: string) {
  await stage([path]);
}

async function handleUnstageSingleFile(path: string) {
  await unstage([path]);
}

async function handleDiscardSingleFile(path: string) {
  if (confirm(t("scm.diff.discard_confirm", { path }))) {
    await discard([path]);
  }
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

async function handleFileMenuAction(action: FileMenuActionType, file: ContextMenuTargetFile) {
  isFileContextMenuOpen.value = false;
  switch (action) {
    case "open-changes":
      openDiff(file.path, file.isStaged);
      break;
    case "open-file":
      handleOpenFile(file.path);
      break;
    case "stage":
      await handleStageSingleFile(file.path);
      break;
    case "unstage":
      await handleUnstageSingleFile(file.path);
      break;
    case "discard":
      await handleDiscardSingleFile(file.path);
      break;
    case "reveal-explorer":
      handleRevealFileInExplorer(file.path);
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

async function handleDiscardAllUnstaged() {
  if (confirm(t("scm.diff.discard_confirm", { path: "*" }))) {
    const allPaths = status.value.unstaged.map((i) => i.path);
    await discard(allPaths);
  }
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

function handleOpenDiffFromGraph(payload: { path: string; staged: boolean }) {
  openDiff(payload.path, payload.staged);
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
  fetchGuardStatus("staged").catch(() => {});
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
    if (tab === "graph" || tab === "changes" || tab === "guard") {
      currentTab.value = tab;
    }
  }
);

onMounted(() => {
  fetchStatus().catch(() => {});
  fetchLog().catch(() => {});
  fetchGuardStatus("staged").catch(() => {});
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
  padding: 8px 14px;
  background: var(--section-bg);
  border-bottom: 1px solid var(--card-border);
  color: var(--text-main);
  min-height: 40px;
}

.box-header-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.box-header-actions {
  display: flex;
  align-items: center;
  gap: 6px;
}

.box-header-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 3px 8px;
  font-size: 11px;
  border-radius: 4px;
  border: 1px solid var(--card-border);
  background: var(--card-bg);
  color: var(--text-muted);
  cursor: pointer;
  transition: all 0.15s ease;
  line-height: 1;
}

.box-header-btn:hover {
  color: var(--text-main);
  background: var(--card-bg-hover, rgba(255, 255, 255, 0.08));
  border-color: var(--border-hover, rgba(255, 255, 255, 0.2));
}

.box-header-btn.primary:hover {
  color: #ffffff;
  background: var(--accent-cyan);
  border-color: var(--accent-cyan);
}

.box-header-btn.danger:hover {
  color: #ffffff;
  background: #ef4444;
  border-color: #ef4444;
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
  justify-content: space-between;
  gap: 12px;
  padding: 7px 14px;
  cursor: pointer;
  border-bottom: 1px solid var(--card-border);
  transition: background 0.15s ease;
}

.box-file-item:hover {
  background: var(--section-bg);
}

.file-item-main {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  min-width: 0;
}

.file-item-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.item-hover-actions {
  display: none;
  align-items: center;
  gap: 3px;
}

.box-file-item:hover .item-hover-actions {
  display: flex;
}

.file-action-icon-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  padding: 0;
  border: none;
  border-radius: 4px;
  background: transparent;
  color: var(--text-muted);
  cursor: pointer;
  transition: all 0.15s ease;
  font-size: 13px;
}

.file-action-icon-btn:hover {
  color: var(--text-main);
  background: rgba(255, 255, 255, 0.12);
}

.file-action-icon-btn.primary:hover {
  color: #ffffff;
  background: var(--accent-cyan);
}

.file-action-icon-btn.danger:hover {
  color: #ffffff;
  background: #ef4444;
}

:root[data-theme="light"] .file-action-icon-btn:hover {
  background: rgba(0, 0, 0, 0.08);
}

:root[data-theme="light"] .file-action-icon-btn.primary:hover {
  color: #ffffff;
  background: var(--accent-cyan);
}

:root[data-theme="light"] .file-action-icon-btn.danger:hover {
  color: #ffffff;
  background: #dc2626;
}

.file-name {
  font-weight: 500;
  color: var(--text-main);
  white-space: nowrap;
}

.file-dir {
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

.clickable-guard-card {
  cursor: pointer;
  transition: transform 0.15s ease, border-color 0.15s ease;
}

.clickable-guard-card:hover {
  transform: translateY(-1px);
  border-color: var(--accent-cyan);
}

.tab-count-pill {
  font-size: 10px;
  font-weight: 700;
  padding: 1px 6px;
  border-radius: 10px;
  line-height: 1.2;
}

.tab-count-pill.pill-error {
  background: rgba(248, 81, 73, 0.2);
  color: #f85149;
}

.tab-count-pill.pill-warning {
  background: rgba(210, 153, 34, 0.2);
  color: #d29922;
}

/* Quality Guard Hub Layout */
.guard-hub-view {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 20px 24px;
  background: var(--bg-color);
  color: var(--text-main);
  overflow-y: auto;
}

.guard-hub-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 16px 20px;
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.guard-hub-title-group {
  display: flex;
  align-items: center;
  gap: 14px;
}

.guard-badge-large {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  border-radius: 10px;
  font-size: 24px;
}

.guard-badge-large.green {
  background: rgba(16, 185, 129, 0.15);
  color: #34d399;
  border: 1px solid rgba(16, 185, 129, 0.3);
}

:root[data-theme="light"] .guard-badge-large.green {
  background: #ecfdf5;
  color: #047857;
  border: 1px solid #a7f3d0;
}

.guard-badge-large.amber {
  background: rgba(245, 158, 11, 0.15);
  color: #fbbf24;
  border: 1px solid rgba(245, 158, 11, 0.3);
}

:root[data-theme="light"] .guard-badge-large.amber {
  background: #fffbeb;
  color: #b45309;
  border: 1px solid #fde68a;
}

.guard-badge-large.red {
  background: rgba(239, 68, 68, 0.15);
  color: #f87171;
  border: 1px solid rgba(239, 68, 68, 0.3);
}

:root[data-theme="light"] .guard-badge-large.red {
  background: #fef2f2;
  color: #b91c1c;
  border: 1px solid #fecaca;
}

.guard-hub-title {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--text-main);
  line-height: 1.4;
}

.guard-hub-desc {
  margin: 4px 0 0 0;
  font-size: 12px;
  color: var(--text-muted);
  line-height: 1.5;
}

.guard-hub-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.guard-scope-toggle {
  display: inline-flex;
  background: var(--section-bg);
  padding: 2px;
  border-radius: 6px;
  border: 1px solid var(--card-border);
}

:root[data-theme="light"] .guard-scope-toggle {
  background: #f1f5f9;
  border-color: #cbd5e1;
}

.guard-scope-toggle .scope-btn {
  padding: 4px 10px;
  font-size: 12px;
  border: none;
  background: transparent;
  color: var(--text-muted);
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.15s ease;
  line-height: 1.4;
}

:root[data-theme="light"] .guard-scope-toggle .scope-btn {
  color: #64748b;
}

.guard-scope-toggle .scope-btn.active {
  background: var(--card-bg);
  color: var(--accent-cyan);
  font-weight: 600;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
}

:root[data-theme="light"] .guard-scope-toggle .scope-btn.active {
  background: #ffffff;
  color: #0284c7;
}

/* 3-Phase Pipeline Architecture Grid */
.guard-pipeline-grid {
  display: flex;
  align-items: center;
  gap: 12px;
}

.pipeline-step-card {
  flex: 1;
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  border-radius: 8px;
  padding: 14px 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

:root[data-theme="light"] .pipeline-step-card {
  background: #ffffff;
  border-color: #cbd5e1;
}

.pipeline-arrow {
  color: var(--text-muted);
  font-size: 18px;
  font-weight: bold;
}

.step-header {
  display: flex;
  align-items: center;
  gap: 8px;
}

.step-badge {
  font-size: 10px;
  font-weight: 700;
  padding: 1px 6px;
  border-radius: 4px;
  background: rgba(56, 189, 248, 0.15);
  color: #38bdf8;
  font-family: monospace;
}

.step-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-main);
  line-height: 1.4;
}

.step-desc {
  font-size: 12px;
  color: var(--text-muted);
  line-height: 1.5;
  min-height: 36px;
}

:root[data-theme="light"] .step-desc {
  color: #475569;
}

.step-status {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  color: var(--text-muted);
  border-top: 1px solid var(--card-border);
  padding-top: 8px;
  line-height: 1.4;
}

:root[data-theme="light"] .step-status {
  border-top-color: #f1f5f9;
  color: #64748b;
}

.text-green {
  color: #34d399;
}

:root[data-theme="light"] .text-green {
  color: #059669;
}

.text-blue {
  color: #38bdf8;
}

:root[data-theme="light"] .text-blue {
  color: #0284c7;
}

/* Diagnostics Panel */
.guard-diagnostics-section {
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  border-radius: 8px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

:root[data-theme="light"] .guard-diagnostics-section {
  background: #ffffff;
  border-color: #cbd5e1;
}

.diagnostics-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: var(--section-bg);
  border-bottom: 1px solid var(--card-border);
}

:root[data-theme="light"] .diagnostics-bar {
  background: #f8fafc;
  border-bottom-color: #e2e8f0;
}

.diagnostics-summary {
  display: flex;
  align-items: center;
  gap: 12px;
}

.diag-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-main);
  line-height: 1.4;
}

.diag-badge-clean {
  font-size: 12px;
  color: #34d399;
  font-weight: 500;
  line-height: 1.5;
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

:root[data-theme="light"] .diag-badge-clean {
  color: #047857;
}

.diag-badge-errors {
  font-size: 12px;
  color: #f87171;
  font-weight: 600;
  line-height: 1.5;
}

:root[data-theme="light"] .diag-badge-errors {
  color: #b91c1c;
}

.diag-badge-warnings {
  font-size: 12px;
  color: #fbbf24;
  font-weight: 600;
  line-height: 1.5;
}

:root[data-theme="light"] .diag-badge-warnings {
  color: #b45309;
}

.diag-files-count {
  font-size: 11px;
  color: var(--text-muted);
}

.guard-issues-table {
  display: flex;
  flex-direction: column;
}

.issue-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 10px 16px;
  border-bottom: 1px solid var(--card-border);
  font-size: 12px;
  line-height: 1.5;
  cursor: pointer;
  transition: background 0.1s ease;
}

:root[data-theme="light"] .issue-row {
  border-bottom-color: #f1f5f9;
}

.issue-row:hover {
  background: rgba(255, 255, 255, 0.04);
}

:root[data-theme="light"] .issue-row:hover {
  background: #f8fafc;
}

.issue-left {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 240px;
}

.issue-row.severity-error .codicon-error {
  color: #f87171;
}

:root[data-theme="light"] .issue-row.severity-error .codicon-error {
  color: #b91c1c;
}

.issue-row.severity-warning .codicon-warning {
  color: #fbbf24;
}

:root[data-theme="light"] .issue-row.severity-warning .codicon-warning {
  color: #b45309;
}

.issue-file {
  font-weight: 500;
  color: var(--text-main);
}

.issue-line {
  color: #38bdf8;
}

:root[data-theme="light"] .issue-line {
  color: #0284c7;
}

.issue-rule {
  font-size: 10px;
  padding: 1px 5px;
  border-radius: 3px;
  background: rgba(255, 255, 255, 0.06);
  color: var(--text-muted);
}

:root[data-theme="light"] .issue-rule {
  background: #f1f5f9;
  color: #475569;
}

.issue-msg {
  flex: 1;
  color: var(--text-muted);
}

:root[data-theme="light"] .issue-msg {
  color: #475569;
}

.issue-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.issue-fixable-pill {
  font-size: 10px;
  padding: 2px 6px;
  border-radius: 4px;
  background: rgba(16, 185, 129, 0.15);
  color: #34d399;
  font-weight: 600;
}

:root[data-theme="light"] .issue-fixable-pill {
  background: #ecfdf5;
  color: #047857;
  border: 1px solid #a7f3d0;
}

.issue-phase-tag {
  font-size: 10px;
  color: var(--text-muted);
  text-transform: uppercase;
}

.guard-clean-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px 20px;
  text-align: center;
}

.guard-clean-state .clean-icon {
  font-size: 48px;
  color: #34d399;
  margin-bottom: 12px;
}

:root[data-theme="light"] .guard-clean-state .clean-icon {
  color: #059669;
}

.guard-clean-state .clean-title {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--text-main);
  line-height: 1.5;
}

.guard-clean-state .clean-subtitle {
  margin: 8px 0 0 0;
  font-size: 13px;
  color: var(--text-muted);
  max-width: 480px;
  line-height: 1.6;
}

:root[data-theme="light"] .guard-clean-state .clean-subtitle {
  color: #64748b;
}
</style>
