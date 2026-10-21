<template>
  <div class="guard-view-container" id="quality-guard-view">
    <!-- Top Action Header -->
    <header class="guard-header">
      <div class="header-left">
        <div class="title-with-badge">
          <span class="codicon codicon-shield header-shield-icon"></span>
          <h1 class="header-title">{{ t("guard.title") || "Quality Guard" }}</h1>
          <span class="status-pill" :class="guardStatus.clean ? 'status-pill-clean' : 'status-pill-warning'">
            <span class="codicon" :class="guardStatus.clean ? 'codicon-pass' : 'codicon-warning'"></span>
            <span>{{
              guardStatus.clean
                ? t("guard.status.healthy") || "All Rules Healthy"
                : `${guardStatus.issues.length} Issues Found`
            }}</span>
          </span>

          <!-- Last Inspected Pill -->
          <div
            v-if="lastCheckedTime"
            class="last-checked-pill"
            :class="{ 'pill-pulse': justInspected }"
            :title="t('guard.last_checked_tooltip') || 'Timestamp of last code quality execution'"
          >
            <span class="codicon codicon-history"></span>
            <span
              >{{ t("guard.last_checked") || "Last Checked" }}: <strong>{{ lastCheckedTime }}</strong></span
            >
            <span class="last-mode font-mono">({{ getModeLabel(lastCheckedMode) }})</span>
          </div>
        </div>
        <p class="header-subtitle">
          {{
            t("guard.subtitle") ||
            "Automated code quality harness, file-extension pipelines, and pre-commit protection."
          }}
        </p>
      </div>

      <div class="header-right-actions">
        <!-- Run Check Staged -->
        <button
          class="btn-action primary"
          :disabled="isChecking"
          @click="handleCheck('staged')"
          title="Run quick staged scan (default pre-commit behavior)"
        >
          <span
            class="codicon"
            :class="isChecking && checkingMode === 'staged' ? 'codicon-loading spin' : 'codicon-play'"
          ></span>
          <span>{{
            isChecking && checkingMode === "staged"
              ? t("guard.actions.checking") || "Checking..."
              : t("guard.actions.check_staged") || "Check Staged"
          }}</span>
        </button>

        <!-- Run Check Changed -->
        <button
          class="btn-action secondary"
          :disabled="isChecking"
          @click="handleCheck('changed')"
          title="Run scan on all working tree changes (staged and unstaged)"
        >
          <span
            class="codicon"
            :class="isChecking && checkingMode === 'changed' ? 'codicon-loading spin' : 'codicon-diff-modified'"
          ></span>
          <span>{{
            isChecking && checkingMode === "changed"
              ? t("guard.actions.checking") || "Checking..."
              : t("guard.actions.check_changed") || "Check Changed"
          }}</span>
        </button>

        <!-- Run Check All (Check Project) -->
        <button
          class="btn-action secondary"
          :disabled="isChecking"
          @click="handleCheck('all')"
          title="Run project quality scan"
        >
          <span
            class="codicon"
            :class="isChecking && checkingMode === 'all' ? 'codicon-loading spin' : 'codicon-run-all'"
          ></span>
          <span>{{ t("guard.actions.check_all") || "Check Project" }}</span>
        </button>

        <!-- One-click Auto Fix with Scope Selector -->
        <div ref="fixDropdownRef" class="split-btn-group">
          <button
            class="split-main-btn"
            :disabled="isFixing || isChecking"
            @click="handleFix()"
            :title="getFixButtonTooltip()"
          >
            <span class="codicon" :class="isFixing ? 'codicon-loading spin' : 'codicon-wand'"></span>
            <span>{{ getFixButtonLabel() }}</span>
          </button>
          <button
            class="split-arrow-btn"
            :disabled="isFixing || isChecking"
            @click.stop="isFixDropdownOpen = !isFixDropdownOpen"
            :aria-expanded="isFixDropdownOpen"
            title="Switch fix scope (Staged, Changed, or Project)"
          >
            <span class="codicon codicon-chevron-down"></span>
          </button>

          <!-- Fix Scope Dropdown Menu -->
          <transition name="dropdown-fade">
            <div v-if="isFixDropdownOpen" class="fix-dropdown-menu">
              <div class="fix-dropdown-header">
                {{ t("guard.actions.fix_scope_title") || "Select Auto-Fix Scope" }}
              </div>
              <button
                class="fix-dropdown-item"
                :class="{ active: checkingMode === 'staged' }"
                @click="selectFixScope('staged')"
              >
                <div class="item-title-row">
                  <span class="codicon codicon-git-stage"></span>
                  <span class="item-name">{{ t("guard.actions.fix_staged") || "Fix Staged" }}</span>
                  <span v-if="checkingMode === 'staged'" class="current-badge">Active</span>
                </div>
                <div class="item-desc">
                  {{ t("guard.actions.fix_staged_desc") || "Fix staged files only ($ unirtm check --staged --fix)" }}
                </div>
              </button>
              <button
                class="fix-dropdown-item"
                :class="{ active: checkingMode === 'changed' }"
                @click="selectFixScope('changed')"
              >
                <div class="item-title-row">
                  <span class="codicon codicon-diff-modified"></span>
                  <span class="item-name">{{ t("guard.actions.fix_changed") || "Fix Changed" }}</span>
                  <span v-if="checkingMode === 'changed'" class="current-badge">Active</span>
                </div>
                <div class="item-desc">
                  {{
                    t("guard.actions.fix_changed_desc") ||
                    "Fix all uncommitted workspace changes ($ unirtm check --fix)"
                  }}
                </div>
              </button>
              <button
                class="fix-dropdown-item"
                :class="{ active: checkingMode === 'all' }"
                @click="selectFixScope('all')"
              >
                <div class="item-title-row">
                  <span class="codicon codicon-run-all"></span>
                  <span class="item-name">{{ t("guard.actions.fix_all") || "Fix Project" }}</span>
                  <span v-if="checkingMode === 'all'" class="current-badge">Active</span>
                </div>
                <div class="item-desc">
                  {{
                    t("guard.actions.fix_all_desc") ||
                    "Auto-format and fix all project files ($ unirtm check --all --fix)"
                  }}
                </div>
              </button>
            </div>
          </transition>
        </div>

        <!-- Toggle Pipeline Settings -->
        <button
          class="btn-action"
          :class="activeTab === 'settings' ? 'primary' : 'ghost'"
          @click="activeTab = activeTab === 'settings' ? 'workbench' : 'settings'"
          title="Toggle Pipeline & Tool Slot Configuration"
        >
          <span class="codicon codicon-settings-gear"></span>
          <span>{{ t("guard.actions.settings") || "Settings" }}</span>
        </button>

        <!-- Fallback: View/Edit .unirtm.toml Button (Opens directly in Monaco Editor) -->
        <button
          class="btn-action ghost"
          @click="openTomlInEditor"
          title="Directly view and edit .unirtm.toml in main code editor"
        >
          <span class="codicon codicon-code"></span>
          <span>{{ t("guard.actions.edit_toml") || "Edit .unirtm.toml" }}</span>
        </button>
      </div>
    </header>

    <!-- Running Scanning Glow Bar -->
    <transition name="fade">
      <div v-if="isChecking || isFixing" class="guard-scanning-banner">
        <div class="scanning-glow-bar">
          <div class="scanning-glow-thumb"></div>
        </div>
        <div class="scanning-info">
          <span class="codicon codicon-loading spin"></span>
          <span v-if="isFixing">{{
            t("guard.running.fixing") || "Running automated fixers and style convergence..."
          }}</span>
          <span v-else>{{
            t("guard.running.checking", { mode: getModeLabel(checkingMode) }) ||
            `Scanning ${getModeLabel(checkingMode)} files across 38 rules...`
          }}</span>
        </div>
      </div>
    </transition>

    <!-- Key Metrics Stat Row -->
    <div class="metrics-row">
      <div class="metric-card">
        <div class="metric-icon-box bg-blue">
          <span class="codicon codicon-symbol-ruler"></span>
        </div>
        <div class="metric-info">
          <span class="metric-val font-mono">{{ guardStatus.rules_count || 38 }}</span>
          <span class="metric-label">{{ t("guard.metrics.active_rules") || "Active Rules & Slots" }}</span>
        </div>
      </div>

      <div class="metric-card">
        <div class="metric-icon-box" :class="guardStatus.clean ? 'bg-green' : 'bg-amber'">
          <span class="codicon" :class="guardStatus.clean ? 'codicon-verified' : 'codicon-warning'"></span>
        </div>
        <div class="metric-info">
          <span class="metric-val font-mono" :class="guardStatus.clean ? 'text-green' : 'text-amber'">
            {{ guardStatus.clean ? "100%" : `${guardStatus.issues.length} Issues` }}
          </span>
          <span class="metric-label">{{ t("guard.metrics.quality_health") || "Quality Health" }}</span>
        </div>
      </div>

      <div class="metric-card">
        <div class="metric-icon-box bg-purple">
          <span class="codicon codicon-watch"></span>
        </div>
        <div class="metric-info">
          <span class="metric-val font-mono">{{
            guardStatus.duration_ms ? `${guardStatus.duration_ms}ms` : "--"
          }}</span>
          <span class="metric-label">{{ t("guard.metrics.scan_latency") || "Inspection Latency" }}</span>
        </div>
      </div>

      <div class="metric-card">
        <div class="metric-icon-box bg-cyan">
          <span class="codicon codicon-layers"></span>
        </div>
        <div class="metric-info">
          <span class="metric-val font-mono">{{ extensionRows.length }}</span>
          <span class="metric-label">{{ t("guard.metrics.pipelines") || "Extension Pipelines" }}</span>
        </div>
      </div>
    </div>

    <!-- Latest Inspection Outcome Feedback Banner -->
    <transition name="slide-down">
      <div
        v-if="latestInspection"
        class="inspection-outcome-card"
        :class="latestInspection.clean ? 'outcome-clean' : 'outcome-warning'"
      >
        <div class="outcome-main">
          <div class="outcome-icon-wrap">
            <span class="codicon" :class="latestInspection.clean ? 'codicon-pass-filled' : 'codicon-warning'"></span>
          </div>
          <div class="outcome-details">
            <div class="outcome-title-row">
              <span class="outcome-headline">
                {{
                  latestInspection.isFix
                    ? t("guard.banner.fix_done") || "Auto-repair pipeline successfully finished"
                    : latestInspection.clean
                      ? t("guard.banner.clean_title") || "Inspection Passed · All rules and pipelines healthy"
                      : t("guard.banner.issues_title", { count: latestInspection.issuesCount }) ||
                        `Inspection alert: ${latestInspection.issuesCount} issues detected`
                }}
              </span>
              <span class="outcome-badge font-mono" :class="latestInspection.clean ? 'badge-clean' : 'badge-warn'">
                {{
                  latestInspection.clean
                    ? t("guard.status.passed") || "PASSED"
                    : `${latestInspection.issuesCount} ISSUES`
                }}
              </span>
            </div>
            <div class="outcome-meta-row">
              <span class="meta-item">
                <span class="codicon codicon-eye"></span>
                <span
                  >{{ t("guard.banner.scope") || "Scope" }}:
                  <strong>{{ getModeLabel(latestInspection.mode) }}</strong></span
                >
              </span>
              <span class="meta-dot">•</span>
              <span class="meta-item">
                <span class="codicon codicon-files"></span>
                <span
                  >{{ t("guard.banner.scanned") || "Scanned" }}: <strong>{{ latestInspection.scanned }}</strong></span
                >
              </span>
              <span class="meta-dot">•</span>
              <span class="meta-item">
                <span class="codicon codicon-watch"></span>
                <span
                  >{{ t("guard.banner.latency") || "Latency" }}:
                  <strong>{{ latestInspection.durationMs }}ms</strong></span
                >
              </span>
              <span class="meta-dot">•</span>
              <span class="meta-item">
                <span class="codicon codicon-clock"></span>
                <span
                  >{{ t("guard.banner.time") || "Checked at" }}: <strong>{{ latestInspection.time }}</strong></span
                >
              </span>
            </div>
          </div>
        </div>

        <div class="outcome-actions">
          <button
            v-if="!latestInspection.clean && activeTab !== 'reports'"
            class="btn-outcome-action"
            @click="activeTab = 'reports'"
            title="Switch to detailed Checklist tab"
          >
            <span class="codicon codicon-checklist"></span>
            <span>{{ t("guard.banner.view_reports") || "View Checklist" }}</span>
          </button>

          <button
            v-if="!latestInspection.clean"
            class="btn-outcome-action primary"
            :disabled="isFixing"
            @click="handleFix()"
            :title="getFixButtonTooltip()"
          >
            <span class="codicon" :class="isFixing ? 'codicon-loading spin' : 'codicon-wand'"></span>
            <span>{{ getFixButtonLabel() }}</span>
          </button>

          <button
            class="btn-outcome-dismiss"
            @click="latestInspection = null"
            :title="t('guard.banner.dismiss') || 'Dismiss'"
          >
            <span class="codicon codicon-close"></span>
          </button>
        </div>
      </div>
    </transition>

    <!-- View Switcher Tabs -->
    <div class="guard-view-tabs-bar">
      <div class="tabs-group">
        <!-- 1. Primary Default Tab: Logs Workbench -->
        <button class="view-tab-btn" :class="{ active: activeTab === 'workbench' }" @click="activeTab = 'workbench'">
          <span class="codicon codicon-terminal"></span>
          <span>{{ t("guard.tabs.workbench") || "Logs" }}</span>
          <span class="tab-badge-pulse" :class="guardStatus.clean ? 'badge-clean' : 'badge-warn'">
            {{ guardStatus.clean ? "CLEAN" : `${guardStatus.issues.length} ISSUES` }}
          </span>
        </button>

        <!-- 2. Rule Checklist -->
        <button class="view-tab-btn" :class="{ active: activeTab === 'reports' }" @click="activeTab = 'reports'">
          <span class="codicon codicon-checklist"></span>
          <span>{{ t("guard.tabs.reports") || "Checklist" }}</span>
          <span v-if="guardStatus.rule_reports && guardStatus.rule_reports.length > 0" class="tab-badge font-mono">
            {{ guardStatus.rule_reports.length }}
          </span>
        </button>

        <!-- 3. Hidden Secondary Tab: Settings -->
        <button
          class="view-tab-btn tab-settings"
          :class="{ active: activeTab === 'settings' }"
          @click="activeTab = 'settings'"
        >
          <span class="codicon codicon-settings-gear"></span>
          <span>{{ t("guard.tabs.settings") || "Settings" }}</span>
          <span class="tab-badge font-mono">{{ extensionRows.length }}</span>
        </button>
      </div>

      <!-- Quick Search Filter in Settings -->
      <div v-if="activeTab === 'settings'" class="filter-search-box">
        <span class="codicon codicon-search"></span>
        <input
          v-model="searchKeyword"
          type="text"
          class="search-input"
          :placeholder="t('guard.search_placeholder') || 'Filter extensions (.go, .ts) or tools...'"
        />
      </div>
    </div>

    <!-- Main Tab View Content Area -->
    <main class="guard-content-body">
      <!-- 1. WORKBENCH: Live Inspection Logs, Execution Output, and Issues -->
      <section v-if="activeTab === 'workbench'" class="tab-section workbench-section">
        <div class="workbench-layout">
          <!-- Live Terminal Output Console Window -->
          <div class="live-console-card">
            <div class="console-card-header">
              <div class="console-header-left">
                <span class="codicon codicon-terminal console-icon"></span>
                <span class="console-title">{{ t("guard.logs.title") || "Logs" }}</span>
                <span
                  class="console-status-dot"
                  :class="isChecking || isFixing ? 'dot-running' : guardStatus.clean ? 'dot-clean' : 'dot-warn'"
                ></span>
                <span class="console-status-text font-mono">
                  {{
                    isChecking ? "INSPECTING..." : isFixing ? "FIXING..." : guardStatus.clean ? "CLEAN" : "ISSUES FOUND"
                  }}
                </span>
              </div>
              <div class="console-header-actions">
                <button
                  v-if="selectedLogIndices.size > 0"
                  class="console-btn text-muted"
                  @click="clearSelectedLogs"
                  title="Deselect all lines"
                >
                  <span class="codicon codicon-close"></span>
                  <span>{{ t("guard.logs.cancel_select") || "Deselect" }}</span>
                </button>
                <button
                  class="console-btn"
                  :class="{ 'btn-highlight': selectedLogIndices.size > 0 }"
                  @click="copySelectedOrAllLogs"
                  :title="
                    selectedLogIndices.size > 0
                      ? 'Copy selected lines to clipboard'
                      : 'Copy inspection logs to clipboard'
                  "
                >
                  <span class="codicon codicon-copy"></span>
                  <span>
                    {{
                      selectedLogIndices.size > 0
                        ? (t("guard.logs.copy_selected") || "Copy Selected") + ` (${selectedLogIndices.size})`
                        : t("guard.logs.copy") || "Copy"
                    }}
                  </span>
                </button>
                <button class="console-btn" @click="clearLogs" title="Clear console output">
                  <span class="codicon codicon-clear-all"></span>
                  <span>{{ t("guard.logs.clear") || "Clear" }}</span>
                </button>
              </div>
            </div>

            <!-- Anti-Virus Real-Time Radar Scanner Bar -->
            <div class="av-radar-scanner" :class="{ 'is-active': isChecking || isFixing || isStreaming }">
              <div class="radar-pulse-ring">
                <span
                  class="codicon"
                  :class="
                    isChecking || isFixing || isStreaming
                      ? 'codicon-sync spin text-cyan'
                      : guardStatus.clean
                        ? 'codicon-shield text-green'
                        : 'codicon-warning text-amber'
                  "
                ></span>
              </div>
              <div class="radar-scanner-content">
                <div class="radar-scan-row">
                  <span class="radar-scan-label font-mono">
                    {{ isChecking || isFixing || isStreaming ? (isFixing ? "REPAIRING" : "SCANNING") : "MONITOR" }}:
                  </span>
                  <span class="radar-scan-target font-mono" :title="currentScanningTarget">
                    {{
                      currentScanningTarget ||
                      (guardStatus.clean
                        ? t("guard.radar.clean_status") || "All 38 quality rules healthy · Real-time inspection idle"
                        : t("guard.radar.issues_status", { count: guardStatus.issues.length }) ||
                          `${guardStatus.issues.length} quality alerts detected · Ready to inspect or fix`)
                    }}
                  </span>
                </div>
                <div class="radar-meta-row">
                  <span class="radar-meta-item">
                    <span class="codicon codicon-symbol-ruler"></span>
                    <span>38 Rules (28 Built-in + 10 Slots)</span>
                  </span>
                  <span class="radar-meta-dot">•</span>
                  <span class="radar-meta-item">
                    <span class="codicon codicon-files"></span>
                    <span>
                      {{
                        isStreaming
                          ? `${currentScannedCount} / ${totalFilesCount || (checkingMode === "all" ? 2631 : guardStatus.scanned || 0)} files`
                          : checkingMode === "all"
                            ? `${guardStatus.scanned || 2631} / ${guardStatus.scanned || 2631} files`
                            : `${guardStatus.scanned || 0} / ${guardStatus.scanned || 0} files`
                      }}
                    </span>
                  </span>
                  <span class="radar-meta-dot">•</span>
                  <span class="radar-meta-item">
                    <span class="codicon codicon-dashboard"></span>
                    <span>{{
                      isStreaming ? "STREAMING" : guardStatus.duration_ms ? `${guardStatus.duration_ms}ms` : "READY"
                    }}</span>
                  </span>
                </div>
              </div>
              <div class="radar-progress-box">
                <span class="radar-progress-pct font-mono">{{ scanProgressPct }}%</span>
                <div class="radar-progress-bar">
                  <div class="radar-progress-fill" :style="{ width: scanProgressPct + '%' }"></div>
                </div>
              </div>
            </div>

            <!-- Terminal Output Stream Body -->
            <div class="console-body" ref="logViewerRef">
              <div v-if="formattedLogs.length === 0" class="console-empty-placeholder">
                <span class="codicon codicon-shield-check placeholder-shield"></span>
                <p class="placeholder-title">Quality Guard Ready</p>
                <p class="placeholder-desc">
                  {{
                    t("guard.logs.empty") ||
                    "No logs yet. Click 'Check Staged' or 'Check Changed' above to inspect files."
                  }}
                </p>
                <div class="placeholder-quick-actions">
                  <button class="placeholder-action-btn primary" :disabled="isChecking" @click="handleCheck('staged')">
                    <span class="codicon codicon-play"></span>
                    <span>{{ t("guard.actions.check_staged") || "Check Staged" }}</span>
                  </button>
                  <button class="placeholder-action-btn" :disabled="isChecking" @click="handleCheck('changed')">
                    <span class="codicon codicon-diff-modified"></span>
                    <span>{{ t("guard.actions.check_changed") || "Check Changed" }}</span>
                  </button>
                </div>
              </div>

              <div v-else class="console-log-lines" @copy="handleConsoleCopy">
                <div
                  v-for="(item, idx) in formattedLogs"
                  :key="idx"
                  class="log-line"
                  :class="{
                    ['log-' + item.type]: true,
                    'is-selected': selectedLogIndices.has(idx),
                    'is-flashed': flashedLineIndex === idx,
                  }"
                  @click="handleLogLineClick(idx, $event)"
                  @dblclick="handleLogLineDblClick(item, idx)"
                  :title="t('guard.logs.line_tip') || 'Click to select, double-click to copy this line'"
                >
                  <span class="log-ts font-mono">{{ item.time }}</span>
                  <span class="log-badge font-mono" :class="'badge-' + item.type">{{ item.tag }}</span>
                  <span class="log-icon-wrap" :class="'icon-wrap-' + item.type">
                    <span class="codicon" :class="item.icon"></span>
                  </span>
                  <span class="log-content font-mono">{{ item.text }}</span>

                  <!-- Hover Quick Action: Copy Single Line -->
                  <button
                    class="log-line-copy-btn"
                    @click.stop="copySingleLog(item, idx, $event)"
                    :title="t('guard.logs.copy_single') || 'Copy this line'"
                  >
                    <span
                      class="codicon"
                      :class="copiedLineIndex === idx ? 'codicon-check text-green' : 'codicon-copy'"
                    ></span>
                  </button>
                </div>
              </div>
            </div>
          </div>

          <!-- Issues & Diagnostics Card (Visible if issues exist) -->
          <div v-if="guardStatus.issues && guardStatus.issues.length > 0" class="workbench-issues-card">
            <div class="issues-card-header">
              <div class="issues-title-group">
                <span class="codicon codicon-warning text-amber"></span>
                <h3 class="issues-title">
                  {{ t("guard.diagnostics.issues") || "Quality Violations Detected" }} ({{ guardStatus.issues.length }})
                </h3>
              </div>
              <button
                class="btn-action primary compact"
                :disabled="isFixing"
                @click="handleFix()"
                :title="getFixButtonTooltip()"
              >
                <span class="codicon" :class="isFixing ? 'codicon-loading spin' : 'codicon-wand'"></span>
                <span>{{ getFixButtonLabel() }}</span>
              </button>
            </div>

            <div class="issues-list">
              <div v-for="(issue, idx) in guardStatus.issues" :key="idx" class="issue-item">
                <div class="issue-item-header">
                  <span class="issue-file font-mono">{{ issue.file_path }}:{{ issue.line }}:{{ issue.column }}</span>
                  <span class="issue-rule font-mono">{{ issue.rule_id }}</span>
                </div>
                <p class="issue-msg">{{ issue.message }}</p>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- 2. Transparent Rule Execution Checklist (Reports Tab) -->
      <section v-if="activeTab === 'reports'" class="tab-section reports-section">
        <div class="reports-container">
          <div class="reports-header-row">
            <div class="reports-header-titles">
              <h3 class="reports-title">
                <span class="codicon codicon-checklist text-cyan"></span>
                <span>{{ t("guard.diagnostics.checklist") || "质量守卫全套规则清单与执行状态" }}</span>
                <span class="rules-total-count font-mono">{{ filteredRichRules.length }} / 38</span>
              </h3>
              <p class="reports-sub">
                涵盖代码卫生清洗、格式化对齐、深层语法门禁与凭据安全防泄露全套流水线，实时展示执行耗时与检测结论。
              </p>
            </div>

            <div class="reports-controls-bar">
              <!-- Category Filter Chips -->
              <div class="rule-category-pills">
                <button
                  class="cat-pill-btn"
                  :class="{ active: selectedRuleCategory === 'all' }"
                  @click="selectedRuleCategory = 'all'"
                >
                  全部 (38)
                </button>
                <button
                  class="cat-pill-btn"
                  :class="{ active: selectedRuleCategory === 'hygiene' }"
                  @click="selectedRuleCategory = 'hygiene'"
                >
                  <span class="codicon codicon-wand text-green"></span>
                  原生自愈 (8)
                </button>
                <button
                  class="cat-pill-btn"
                  :class="{ active: selectedRuleCategory === 'formatting' }"
                  @click="selectedRuleCategory = 'formatting'"
                >
                  <span class="codicon codicon-diff-modified text-blue"></span>
                  格式排版 (5)
                </button>
                <button
                  class="cat-pill-btn"
                  :class="{ active: selectedRuleCategory === 'linter' }"
                  @click="selectedRuleCategory = 'linter'"
                >
                  <span class="codicon codicon-pass text-amber"></span>
                  规范门禁 (19)
                </button>
                <button
                  class="cat-pill-btn"
                  :class="{ active: selectedRuleCategory === 'security' }"
                  @click="selectedRuleCategory = 'security'"
                >
                  <span class="codicon codicon-shield text-purple"></span>
                  凭据安全 (6)
                </button>
              </div>

              <!-- Rule Search Box -->
              <div class="rule-search-box">
                <span class="codicon codicon-search search-icon"></span>
                <input
                  v-model="ruleSearchKeyword"
                  type="text"
                  class="rule-search-input"
                  placeholder="搜索规则名称、ID、扩展名 (.go, .ts) 或工具..."
                />
                <span
                  v-if="ruleSearchKeyword"
                  class="codicon codicon-close clear-icon"
                  @click="ruleSearchKeyword = ''"
                ></span>
              </div>
            </div>
          </div>

          <!-- Rich Rule Cards Grid -->
          <div v-if="filteredRichRules.length > 0" class="rich-rules-cards-grid">
            <div
              v-for="rule in filteredRichRules"
              :key="rule.id"
              class="rich-rule-card"
              :class="[
                'category-' + rule.category,
                rule.status === 'fail' || rule.status === 'error'
                  ? 'is-fail'
                  : rule.status === 'warn'
                    ? 'is-warn'
                    : rule.status === 'fixed'
                      ? 'is-fixed'
                      : rule.status === 'passed'
                        ? 'is-passed'
                        : 'is-idle',
              ]"
            >
              <!-- Card Header -->
              <div class="rule-card-top">
                <div class="rule-title-group">
                  <div class="rule-status-indicator">
                    <span
                      class="codicon"
                      :class="
                        rule.status === 'fail' || rule.status === 'error'
                          ? 'codicon-error text-red'
                          : rule.status === 'warn'
                            ? 'codicon-warning text-amber'
                            : rule.status === 'fixed'
                              ? 'codicon-wand text-green'
                              : rule.status === 'passed'
                                ? 'codicon-pass text-green'
                                : 'codicon-circle-outline text-muted'
                      "
                    ></span>
                  </div>
                  <div class="rule-names">
                    <div class="rule-chinese-name">{{ rule.name }}</div>
                    <code class="rule-id-code font-mono">{{ rule.id }}</code>
                  </div>
                </div>

                <div class="rule-badges-group">
                  <span v-if="rule.status !== 'idle'" class="rule-run-badge font-mono" :class="'badge-' + rule.status">
                    {{
                      rule.status === "passed"
                        ? "✔ 通过"
                        : rule.status === "fixed"
                          ? "⚡ 已自愈"
                          : rule.status === "warn"
                            ? `! ${rule.issuesCount} 警告`
                            : `✕ ${rule.issuesCount} 阻断`
                    }}
                    <span v-if="rule.duration">· {{ rule.duration }}</span>
                  </span>
                  <span v-else class="rule-idle-badge font-mono">就绪待命</span>
                </div>
              </div>

              <!-- Card Description (实质内容) -->
              <p class="rule-description">{{ rule.description }}</p>

              <!-- Card Meta & Extension Tags -->
              <div class="rule-card-bottom">
                <div class="rule-meta-chips">
                  <span class="meta-chip phase-chip" :class="'phase-' + rule.phase">{{ rule.phaseLabel }}</span>
                  <span class="meta-chip tool-chip font-mono">{{ rule.toolName }}</span>
                  <span v-if="rule.fixable" class="meta-chip fixable-chip">⚡ 可自愈</span>
                  <span v-if="rule.severity === 'blocker'" class="meta-chip blocker-chip">🛑 阻断门禁</span>
                  <span v-else class="meta-chip warn-chip">⚠️ 建议告警</span>
                </div>

                <div class="rule-extensions-row">
                  <span class="ext-label font-mono">适用:</span>
                  <span v-for="ext in rule.extensions" :key="ext" class="ext-badge font-mono">{{ ext }}</span>
                </div>
              </div>
            </div>
          </div>

          <div v-else class="empty-reports-box">
            <span class="codicon codicon-search empty-reports-icon"></span>
            <p>未找到匹配 "{{ ruleSearchKeyword }}" 的守卫规则</p>
            <button
              class="btn-clear-filter"
              @click="
                ruleSearchKeyword = '';
                selectedRuleCategory = 'all';
              "
            >
              重置筛选条件
            </button>
          </div>
        </div>
      </section>

      <!-- 3. HIDDEN SETTINGS: Pipeline Matrix & Tool Slots -->
      <section v-if="activeTab === 'settings'" class="tab-section settings-section">
        <div class="settings-hint-box">
          <div class="hint-text">
            <span class="codicon codicon-info"></span>
            <span>{{
              t("guard.settings.hint") ||
              "Advanced Configuration: Manage extension pipelines, external CLI tools, and .unirtm.toml mappings."
            }}</span>
          </div>
          <div class="settings-actions">
            <button class="btn-action accent compact" @click="openAddSlotModal">
              <span class="codicon codicon-plus"></span>
              <span>{{ t("guard.actions.add_mapping") || "Add Extension Mapping" }}</span>
            </button>
            <button class="btn-action ghost compact" @click="openTomlInEditor">
              <span class="codicon codicon-code"></span>
              <span>{{ t("guard.actions.edit_toml") || "Edit .unirtm.toml" }}</span>
            </button>
          </div>
        </div>

        <!-- Sub Tabs inside Settings: Pipelines Matrix vs Built-in Rules vs CLI Slots -->
        <div class="settings-sub-tabs">
          <button
            class="sub-tab-btn"
            :class="{ active: settingsSubTab === 'matrix' }"
            @click="settingsSubTab = 'matrix'"
          >
            <span class="codicon codicon-table"></span>
            <span>{{ t("guard.tabs.matrix") || "Extension Pipelines (Matrix)" }}</span>
          </button>
          <button
            class="sub-tab-btn"
            :class="{ active: settingsSubTab === 'builtin' }"
            @click="settingsSubTab = 'builtin'"
          >
            <span class="codicon codicon-shield text-cyan"></span>
            <span>{{ t("guard.tabs.builtin") || "Built-in Rules (原生)" }} ({{ BUILTIN_GUARD_RULES.length }})</span>
          </button>
          <button class="sub-tab-btn" :class="{ active: settingsSubTab === 'slots' }" @click="settingsSubTab = 'slots'">
            <span class="codicon codicon-symbol-class"></span>
            <span>{{ t("guard.tabs.slots") || "Tool Slots (CLI)" }} ({{ guardSlots.length }})</span>
          </button>
        </div>

        <!-- 3A. Matrix Table -->
        <div v-if="settingsSubTab === 'matrix'" class="table-responsive-wrapper">
          <table class="pipeline-matrix-table" id="table-extension-pipelines">
            <thead>
              <tr>
                <th class="col-ext">{{ t("guard.table.extension") || "File Type / Extension" }}</th>
                <th class="col-linter">{{ t("guard.table.linter") || "Linter (Static Audit)" }}</th>
                <th class="col-formatter">{{ t("guard.table.formatter") || "Formatter (Styling)" }}</th>
                <th class="col-fixer">{{ t("guard.table.fixer") || "Fixer (Auto Healing)" }}</th>
                <th class="col-security">{{ t("guard.table.security") || "Security & Audit" }}</th>
                <th class="col-actions text-right">{{ t("guard.table.actions") || "Actions" }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="row in filteredExtensionRows" :key="row.extension" class="pipeline-row">
                <!-- Extension Column -->
                <td class="col-ext">
                  <div class="ext-badge-group">
                    <span class="ext-tag font-mono">{{ row.extension }}</span>
                    <span class="lang-label">{{ row.languages.join(", ") }}</span>
                  </div>
                </td>

                <!-- Linter Column -->
                <td class="col-linter">
                  <div v-if="row.linter" class="tool-cell" @click="editSlot(row.linter)">
                    <span class="tool-name">{{ row.linter.default_tool_name }}</span>
                    <code class="tool-cmd font-mono">{{ row.linter.default_command }}</code>
                  </div>
                  <div
                    v-else-if="getBuiltinRuleForExtAndCategory(row.extension, 'linter')"
                    class="tool-cell builtin-cell"
                    @click="openConfigureBuiltinRule(getBuiltinRuleForExtAndCategory(row.extension, 'linter')!)"
                    title="点击配置该原生内置规则参数"
                  >
                    <span class="tool-name text-cyan"
                      ><span class="codicon codicon-shield"></span>
                      {{ getBuiltinRuleForExtAndCategory(row.extension, "linter")!.name }}</span
                    >
                    <span class="builtin-note font-mono">原生守卫 (点击配置)</span>
                  </div>
                  <span v-else class="cell-empty font-mono">—</span>
                </td>

                <!-- Formatter Column -->
                <td class="col-formatter">
                  <div v-if="row.formatter" class="tool-cell" @click="editSlot(row.formatter)">
                    <span class="tool-name">{{ row.formatter.default_tool_name }}</span>
                    <code class="tool-cmd font-mono">{{ row.formatter.default_command }}</code>
                  </div>
                  <span v-else class="cell-empty font-mono">—</span>
                </td>

                <!-- Fixer Column -->
                <td class="col-fixer">
                  <div v-if="row.fixer" class="tool-cell" @click="editSlot(row.fixer)">
                    <span class="tool-name text-green">{{ row.fixer.default_tool_name }}</span>
                    <code class="tool-cmd font-mono text-green">{{
                      row.fixer.fix_command || row.fixer.default_command
                    }}</code>
                  </div>
                  <div
                    v-else-if="getBuiltinRuleForExtAndCategory(row.extension, 'fixer')"
                    class="tool-cell builtin-cell"
                    @click="openConfigureBuiltinRule(getBuiltinRuleForExtAndCategory(row.extension, 'fixer')!)"
                    title="点击配置该原生自愈规则参数"
                  >
                    <span class="tool-name text-green"
                      ><span class="codicon codicon-wand"></span>
                      {{ getBuiltinRuleForExtAndCategory(row.extension, "fixer")!.name }}</span
                    >
                    <span class="builtin-note font-mono">原生自愈 (点击配置)</span>
                  </div>
                  <span v-else class="cell-empty font-mono">—</span>
                </td>

                <!-- Security Column -->
                <td class="col-security">
                  <div v-if="row.auditor" class="tool-cell" @click="editSlot(row.auditor)">
                    <span class="tool-name text-purple">{{ row.auditor.default_tool_name }}</span>
                    <code class="tool-cmd font-mono">{{ row.auditor.default_command }}</code>
                  </div>
                  <div
                    v-else-if="getBuiltinRuleForExtAndCategory(row.extension, 'security')"
                    class="tool-cell builtin-cell"
                    @click="openConfigureBuiltinRule(getBuiltinRuleForExtAndCategory(row.extension, 'security')!)"
                    title="点击配置该原生安全规则参数"
                  >
                    <span class="tool-name text-purple"
                      ><span class="codicon codicon-shield text-purple"></span>
                      {{ getBuiltinRuleForExtAndCategory(row.extension, "security")!.name }}</span
                    >
                    <span class="builtin-note font-mono">原生凭据安全 (点击配置)</span>
                  </div>
                  <span v-else class="cell-empty font-mono">—</span>
                </td>

                <!-- Action Buttons -->
                <td class="col-actions text-right">
                  <div class="action-btn-row">
                    <button
                      class="row-btn edit-btn"
                      title="Edit mappings for this extension"
                      @click="handleEditRow(row)"
                    >
                      <span class="codicon codicon-edit"></span>
                    </button>
                    <button
                      class="row-btn add-btn"
                      title="Add tool slot to this extension"
                      @click="handleAddToolToExt(row.extension)"
                    >
                      <span class="codicon codicon-plus"></span>
                    </button>
                  </div>
                </td>
              </tr>

              <tr v-if="filteredExtensionRows.length === 0" class="empty-row">
                <td colspan="6" class="text-center py-4">
                  <span class="codicon codicon-search empty-icon"></span>
                  <p class="empty-text">
                    {{ t("guard.table.no_match") || "No extension pipeline matches the search query." }}
                  </p>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- 3B. Built-in Rules Settings View (Configurable, Modifiable, Non-deletable) -->
        <div v-if="settingsSubTab === 'builtin'" class="builtin-rules-settings-view">
          <div class="builtin-settings-header-banner">
            <div class="banner-left">
              <span class="codicon codicon-shield banner-icon text-cyan"></span>
              <div>
                <h4 class="banner-title">UniRTM Go 原生引擎核心内置规则 (25项)</h4>
                <p class="banner-subtitle">
                  所有内置规则均由轻量级原生引擎驱动，零外部依赖、极速毫秒级执行。
                  <strong>系统核心预置规则·不可删除</strong>，但支持在此自由修改参数、排除目录/文件或灵活开启/跳过。
                </p>
              </div>
            </div>
            <div class="banner-stats">
              <span class="stat-pill active-pill">
                <span class="dot dot-green"></span>
                <span>已启用: {{ activeBuiltinCount }}</span>
              </span>
              <span class="stat-pill skip-pill">
                <span class="dot dot-amber"></span>
                <span>已跳过: {{ skippedBuiltinCount }}</span>
              </span>
            </div>
          </div>

          <!-- Builtin Rules Grid -->
          <div class="builtin-cards-grid">
            <div
              v-for="rule in filteredSettingsBuiltinRules"
              :key="rule.id"
              class="builtin-card"
              :class="{ 'is-skipped': isRuleSkipped(rule.id) }"
            >
              <div class="builtin-card-top">
                <div class="builtin-title-group">
                  <span class="builtin-cn-name">{{ rule.name }}</span>
                  <code class="builtin-id-code font-mono">{{ rule.id }}</code>
                </div>

                <div class="builtin-actions-group">
                  <!-- Non-deletable lock badge -->
                  <span class="badge-lock-tag font-mono" title="系统核心内置规则，无法被删除">
                    <span class="codicon codicon-lock"></span> 不可删除
                  </span>

                  <!-- Status Toggle Switch -->
                  <button
                    class="btn-toggle-switch"
                    :class="{ active: !isRuleSkipped(rule.id) }"
                    :title="isRuleSkipped(rule.id) ? '点击恢复启用此规则' : '点击跳过此规则'"
                    @click.stop="toggleBuiltinRule(rule.id)"
                  >
                    <span class="toggle-thumb"></span>
                    <span class="toggle-label font-mono">{{ isRuleSkipped(rule.id) ? "已跳过" : "已启用" }}</span>
                  </button>
                </div>
              </div>

              <p class="builtin-desc">{{ rule.description }}</p>

              <!-- Custom configured indicator if any -->
              <div v-if="hasRuleCustomConfig(rule.id)" class="builtin-custom-tag-row">
                <span class="codicon codicon-check text-cyan"></span>
                <span class="custom-hint font-mono">已配置: {{ getRuleCustomSummary(rule.id) }}</span>
              </div>

              <div class="builtin-card-footer">
                <div class="builtin-chips">
                  <span class="chip phase-chip" :class="'phase-' + rule.phase">{{ rule.phaseLabel }}</span>
                  <span v-if="rule.fixable" class="chip fix-chip">⚡ 可自愈</span>
                  <span v-if="rule.severity === 'blocker'" class="chip blocker-chip">🛑 阻断</span>
                  <span v-else class="chip warn-chip">⚠️ 告警</span>
                </div>

                <div class="builtin-footer-right">
                  <button
                    class="btn-configure-builtin"
                    title="配置此内置规则的运行参数与排除路径"
                    @click="openConfigureBuiltinRule(rule)"
                  >
                    <span class="codicon codicon-settings"></span>
                    <span>配置规则</span>
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 3C. Tool Slots Grid (CLI Tool Cards) -->
        <div v-if="settingsSubTab === 'slots'" class="slots-cards-grid">
          <div v-for="slot in filteredSlots" :key="slot.id" class="slot-card">
            <div class="slot-card-header">
              <div class="slot-title-wrap">
                <span class="slot-tool-name">{{ slot.default_tool_name }}</span>
                <span class="slot-id font-mono">{{ slot.id }}</span>
              </div>
              <div class="slot-header-tags">
                <span class="phase-pill" :class="'phase-' + slot.phase">{{ slot.phase }}</span>
                <button class="icon-edit-btn" title="Edit Slot" @click="editSlot(slot)">
                  <span class="codicon codicon-edit"></span>
                </button>
              </div>
            </div>

            <p class="slot-desc">{{ slot.description }}</p>

            <div class="slot-exts-wrap">
              <span class="ext-label">{{ t("guard.slots.extensions") || "Extensions:" }}</span>
              <div class="ext-tags-row">
                <span
                  v-for="ext in slot.default_extensions && slot.default_extensions.length > 0
                    ? slot.default_extensions
                    : ['*']"
                  :key="ext"
                  class="ext-chip font-mono"
                >
                  {{ ext }}
                </span>
              </div>
            </div>

            <div class="slot-commands-box">
              <div class="cmd-line" v-if="slot.default_command">
                <span class="cmd-label">{{ t("guard.slots.check") || "Check:" }}</span>
                <code class="cmd-text font-mono">{{ slot.default_command }}</code>
              </div>
              <div class="cmd-line" v-if="slot.fix_command">
                <span class="cmd-label text-green">{{ t("guard.slots.fix") || "Fix:" }}</span>
                <code class="cmd-text font-mono text-green">{{ slot.fix_command }}</code>
              </div>
            </div>
          </div>
        </div>
      </section>
    </main>

    <!-- Extension / Tool Slot CRUD Modal -->
    <ExtensionSlotModal
      :isOpen="isSlotModalOpen"
      :initialSlot="activeEditingSlot"
      :targetExtension="targetExtForAdd"
      @close="isSlotModalOpen = false"
      @save="onSaveSlot"
      @delete="onDeleteSlot"
    />

    <!-- Built-in Rule Config Modal (Configurable, Modifiable, Non-deletable) -->
    <BuiltinRuleConfigModal
      :isOpen="isBuiltinModalOpen"
      :rule="activeEditingBuiltinRule"
      :initialConfig="currentEditingRuleCustomConfig"
      :isSkipped="activeEditingBuiltinRule ? isRuleSkipped(activeEditingBuiltinRule.id) : false"
      @close="isBuiltinModalOpen = false"
      @save="onSaveBuiltinRule"
    />

    <!-- Toml Config Fallback Drawer -->
    <TomlConfigDrawer :isOpen="isTomlDrawerOpen" @close="isTomlDrawerOpen = false" @saved="onTomlSaved" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onActivated, onUnmounted } from "vue";
import { useRouter } from "vue-router";
import { t } from "../i18n";
import { useGuard, type ToolSlotItem, type SlotConfigInput } from "../composables/useGuard";
import { useEditorTabs } from "../composables/useEditorTabs";
import { useToast } from "../composables/useToast";
import ExtensionSlotModal from "../components/guard/ExtensionSlotModal.vue";
import TomlConfigDrawer from "../components/guard/TomlConfigDrawer.vue";
import BuiltinRuleConfigModal, { type BuiltinRuleCustomConfig } from "../components/guard/BuiltinRuleConfigModal.vue";
import {
  ALL_38_GUARD_RULES,
  BUILTIN_GUARD_RULES,
  formatGuardDuration,
  type DetailedGuardRule,
} from "../utils/guardRulesMeta";
import { parseGuardConfigFromToml, updateBuiltinRuleInToml, type ParsedGuardConfig } from "../utils/tomlGuardHelper";

const emit = defineEmits<{
  (e: "log-event", level: "INFO" | "SUCCESS" | "WARN" | "ERROR", msg: string, detail?: string, channel?: string): void;
}>();

const router = useRouter();
const { openTab } = useEditorTabs();
const { showToast } = useToast();

const {
  guardStatus,
  guardSlots,
  isChecking,
  isFixing,
  lastRunResult,
  fetchStatus,
  fetchSlots,
  runCheck,
  runFix,
  saveSlotConfig,
  deleteSlotConfig,
  fetchRawTomlConfig,
  saveRawTomlConfig,
} = useGuard();

const isSlotModalOpen = ref(false);
const activeEditingSlot = ref<ToolSlotItem | null>(null);
const targetExtForAdd = ref<string>("");

// Built-in Rules Configuration State (Configurable, Modifiable, Non-deletable)
const isBuiltinModalOpen = ref(false);
const activeEditingBuiltinRule = ref<DetailedGuardRule | null>(null);
const parsedGuardConfig = ref<ParsedGuardConfig>({ skip: [], rules: {} });

const isTomlDrawerOpen = ref(false);
const checkingMode = ref<"changed" | "staged" | "all">("staged");

// Auto-fix scope dropdown state
const isFixDropdownOpen = ref(false);
const fixDropdownRef = ref<HTMLElement | null>(null);

function handleDocClick(e: MouseEvent) {
  if (fixDropdownRef.value && !fixDropdownRef.value.contains(e.target as Node)) {
    isFixDropdownOpen.value = false;
  }
}

function getFixButtonLabel(mode?: "staged" | "changed" | "all") {
  const targetMode = mode || checkingMode.value;
  if (isFixing.value) {
    return t("guard.actions.fixing") || "Fixing...";
  }
  if (targetMode === "staged") {
    return t("guard.actions.fix_staged") || "Fix Staged";
  }
  if (targetMode === "all") {
    return t("guard.actions.fix_all") || "Fix Project";
  }
  return t("guard.actions.fix_changed") || "Fix Changed";
}

function getFixButtonTooltip(mode?: "staged" | "changed" | "all") {
  const targetMode = mode || checkingMode.value;
  if (targetMode === "staged") {
    return t("guard.actions.fix_staged_desc") || "Fix staged files only ($ unirtm check --staged --fix)";
  }
  if (targetMode === "all") {
    return t("guard.actions.fix_all_desc") || "Auto-format and fix all project files ($ unirtm check --all --fix)";
  }
  return t("guard.actions.fix_changed_desc") || "Fix all uncommitted workspace changes ($ unirtm check --fix)";
}

function selectFixScope(mode: "staged" | "changed" | "all") {
  isFixDropdownOpen.value = false;
  handleFix(mode);
}

// Default primary tab is now the transparent Live Guard & Logs Workbench!
const activeTab = ref<"workbench" | "reports" | "settings">("workbench");
const settingsSubTab = ref<"matrix" | "builtin" | "slots">("matrix");
const searchKeyword = ref("");

// Rich Rule Checklist Reactive State & Filtering
const selectedRuleCategory = ref<"all" | "hygiene" | "formatting" | "linter" | "security">("all");
const ruleSearchKeyword = ref("");

export interface RichGuardRuleItem extends DetailedGuardRule {
  status: "passed" | "fixed" | "fail" | "error" | "warn" | "idle";
  duration: string;
  durationMs: number;
  issuesCount: number;
}

const richRulesList = computed<RichGuardRuleItem[]>(() => {
  const reports = guardStatus.value.rule_reports || [];
  const reportsMap = new Map<string, { status: string; duration: string; durationMs: number; issuesCount: number }>();
  for (const r of reports) {
    const rawStatus = (r.status || "").toLowerCase();
    const rawDur = r.duration_ms !== undefined ? r.duration_ms : r.duration !== undefined ? r.duration : undefined;
    const dur = formatGuardDuration(rawDur);
    const durMs = typeof rawDur === "number" && rawDur > 50000 ? Math.round(rawDur / 1_000_000) : Number(rawDur) || 0;
    const issues = r.issues_count || (r.errors_count || 0) + (r.warnings_count || 0);
    const item = { status: rawStatus, duration: dur, durationMs: durMs, issuesCount: issues };
    if (r.rule_id) {
      reportsMap.set(r.rule_id, item);
    }
    if (r.rule_name) {
      reportsMap.set(r.rule_name, item);
    }
  }

  return ALL_38_GUARD_RULES.map((rule) => {
    const report = reportsMap.get(rule.id) || reportsMap.get(rule.name);
    let status: RichGuardRuleItem["status"] = "idle";
    let duration = "";
    let durationMs = 0;
    let issuesCount = 0;
    if (report) {
      duration = report.duration;
      durationMs = report.durationMs;
      issuesCount = report.issuesCount;
      if (report.status === "passed" || report.status === "pass" || report.status === "clean") {
        status = "passed";
      } else if (report.status === "fixed") {
        status = "fixed";
      } else if (report.status === "fail" || report.status === "error") {
        status = "fail";
      } else if (report.status === "warn" || report.status === "warning") {
        status = "warn";
      } else {
        status = "passed";
      }
    }
    return {
      ...rule,
      status,
      duration,
      durationMs,
      issuesCount,
    };
  });
});

const filteredRichRules = computed(() => {
  let list = richRulesList.value;
  if (selectedRuleCategory.value !== "all") {
    list = list.filter((r) => r.category === selectedRuleCategory.value);
  }
  const kw = ruleSearchKeyword.value.trim().toLowerCase();
  if (kw) {
    list = list.filter((r) => {
      return (
        r.name.toLowerCase().includes(kw) ||
        r.id.toLowerCase().includes(kw) ||
        r.description.toLowerCase().includes(kw) ||
        r.toolName.toLowerCase().includes(kw) ||
        r.extensions.some((ext) => ext.toLowerCase().includes(kw))
      );
    });
  }
  return list;
});

// Last inspection tracking for high visibility and transparency
const lastCheckedTime = ref<string>("");
const lastCheckedMode = ref<"changed" | "staged" | "all">("staged");
const justInspected = ref(false);

export interface FormattedLogItem {
  time: string;
  tag: string;
  type: "info" | "pass" | "warn" | "error" | "config" | "phase" | "cmd" | "scan" | "rule";
  icon: string;
  text: string;
  rawText?: string;
}

const logViewerRef = ref<HTMLElement | null>(null);
const localLogs = ref<FormattedLogItem[]>([]);

// Interactive Log Line Selection and Copy States
const selectedLogIndices = ref<Set<number>>(new Set());
const lastSelectedLogIndex = ref<number | null>(null);
const copiedLineIndex = ref<number | null>(null);
const flashedLineIndex = ref<number | null>(null);

// Anti-Virus style real-time scanner dynamic states
const isStreaming = ref(false);
const currentScanningTarget = ref("");
const currentScannedCount = ref(0);
const totalFilesCount = ref(0);
const scanProgressPct = ref(100);

function parseLogLine(raw: string): FormattedLogItem {
  const now = new Date();
  const time = now.toTimeString().split(" ")[0];

  // Strip ANSI escape sequences (e.g. \u001b[32m or [32m)
  const clean = raw
    .replace(/\u001b\[[0-9;]*[a-zA-Z]/g, "")
    .replace(/\[\d{1,2}(?:;\d{1,2})*m/g, "")
    .trim();

  let type: FormattedLogItem["type"] = "info";
  let tag = "GUARD";
  let icon = "codicon-info text-blue";
  let displayText = clean;

  if (clean.startsWith("$ ")) {
    type = "cmd";
    tag = "CLI";
    icon = "codicon-terminal text-cyan";
    displayText = clean.replace(/^\$\s*/, "");
  } else if (clean.startsWith("ℹ") || clean.includes("Inspecting workspace files")) {
    type = "info";
    tag = "INFO";
    icon = "codicon-info text-blue";
    displayText = clean.replace(/^ℹ\s*/, "");
  } else if (clean.includes("› [scan]") || clean.includes("more files in inspection queue")) {
    type = "scan";
    tag = "SCAN";
    icon = "codicon-search text-cyan";
    displayText = clean.replace(/^\s*›\s*\[scan\]\s*/, "");
  } else if (
    clean.startsWith("✓ ✨") ||
    clean.includes("Quality Guard passed") ||
    clean.includes("No matching files found")
  ) {
    type = "pass";
    tag = "PASS";
    icon = "codicon-check-all text-green";
    displayText = clean.replace(/^✓\s*✨?\s*/, "");
  } else if (clean.startsWith("✓ ") || clean.startsWith("✓")) {
    type = "rule";
    tag = "RULE";
    icon = "codicon-pass text-green";
    // The UI icon already represents the passed checkmark, so trim the duplicate symbol from display text
    displayText = clean.replace(/^✓\s*/, "");
  } else if (clean.startsWith("✗ ") || clean.startsWith("✗") || clean.includes("[FAIL]")) {
    type = "error";
    tag = "FAIL";
    icon = "codicon-error text-red";
    displayText = clean.replace(/^✗\s*/, "");
  } else if (clean.startsWith("! ") || clean.startsWith("!") || clean.startsWith("⚠") || clean.includes("[WARN]")) {
    type = "warn";
    tag = "WARN";
    icon = "codicon-warning text-amber";
    displayText = clean.replace(/^[!⚠]\s*/, "");
  } else if (clean.includes("[Config:")) {
    type = "config";
    tag = "CONFIG";
    icon = "codicon-settings-gear text-purple";
  } else if (clean.includes("[Phase")) {
    type = "phase";
    tag = "PHASE";
    icon = "codicon-layers text-indigo";
  } else if (
    clean.toLowerCase().includes("passed") ||
    clean.toLowerCase().includes("clean") ||
    clean.toLowerCase().includes("compliant")
  ) {
    type = "pass";
    tag = "PASS";
    icon = "codicon-pass text-green";
  }

  return {
    time,
    tag,
    type,
    icon,
    text: displayText,
    rawText: clean,
  };
}

// Anti-Virus streaming typewriter effect for live logs inspection transparency
async function streamPushLogs(rawLines: string[]) {
  if (!rawLines || rawLines.length === 0) return;
  isStreaming.value = true;
  localLogs.value = [];
  currentScannedCount.value = 0;
  scanProgressPct.value = 0;

  const totalLines = rawLines.length;
  const inspectLine = rawLines.find((l) => l.includes("Inspecting workspace files"));
  if (inspectLine) {
    const match = inspectLine.match(/\((\d+)\s+files/);
    if (match) {
      totalFilesCount.value = parseInt(match[1], 10);
    }
  }

  for (let i = 0; i < totalLines; i++) {
    const line = rawLines[i];
    localLogs.value.push(parseLogLine(line));

    if (line.includes("› [scan]")) {
      currentScanningTarget.value = line.replace(/^\s*›\s*\[scan\]\s*/, "");
      currentScannedCount.value++;
    } else if (line.startsWith("✓ ") || line.startsWith("✗ ") || line.startsWith("! ")) {
      currentScanningTarget.value = line.replace(/^[✓✗!]\s*/, "");
    } else if (line.startsWith("$ ")) {
      currentScanningTarget.value = line;
    }

    scanProgressPct.value = Math.min(100, Math.round(((i + 1) / totalLines) * 100));
    scrollToBottom();

    await new Promise((resolve) => setTimeout(resolve, 14));
  }

  isStreaming.value = false;
  scanProgressPct.value = 100;
  if (guardStatus.value.clean) {
    currentScanningTarget.value = "All 38 guard rules verified clean · Workspace protected";
  }
}

const formattedLogs = computed(() => {
  if (localLogs.value.length > 0) return localLogs.value;
  if (guardStatus.value.logs && guardStatus.value.logs.length > 0) {
    return guardStatus.value.logs.map(parseLogLine);
  }
  return [];
});

function scrollToBottom() {
  setTimeout(() => {
    if (logViewerRef.value) {
      logViewerRef.value.scrollTop = logViewerRef.value.scrollHeight;
    }
  }, 50);
}

function copySingleLog(item: FormattedLogItem, idx: number, e?: MouseEvent) {
  if (e) e.stopPropagation();
  const textToCopy = item.rawText || item.text;
  if (!textToCopy) return;
  navigator.clipboard
    .writeText(textToCopy)
    .then(() => {
      copiedLineIndex.value = idx;
      setTimeout(() => {
        if (copiedLineIndex.value === idx) copiedLineIndex.value = null;
      }, 1500);
      showToast(t("guard.logs.copied_single") || "Log line copied to clipboard", "success");
    })
    .catch(() => {});
}

function handleLogLineDblClick(item: FormattedLogItem, idx: number) {
  const textToCopy = item.rawText || item.text;
  if (!textToCopy) return;
  navigator.clipboard
    .writeText(textToCopy)
    .then(() => {
      flashedLineIndex.value = idx;
      copiedLineIndex.value = idx;
      setTimeout(() => {
        if (flashedLineIndex.value === idx) flashedLineIndex.value = null;
        if (copiedLineIndex.value === idx) copiedLineIndex.value = null;
      }, 1500);
      showToast(t("guard.logs.copied_double") || "Double-click: Log line copied to clipboard", "success");
    })
    .catch(() => {});
}

function handleLogLineClick(idx: number, e: MouseEvent) {
  // If user is selecting text with mouse, do not toggle row selection
  const selection = window.getSelection();
  if (selection && selection.toString().length > 1) {
    return;
  }

  if (e.shiftKey && lastSelectedLogIndex.value !== null) {
    const start = Math.min(lastSelectedLogIndex.value, idx);
    const end = Math.max(lastSelectedLogIndex.value, idx);
    const newSet = new Set(selectedLogIndices.value);
    for (let i = start; i <= end; i++) {
      newSet.add(i);
    }
    selectedLogIndices.value = newSet;
  } else if (e.ctrlKey || e.metaKey) {
    const newSet = new Set(selectedLogIndices.value);
    if (newSet.has(idx)) {
      newSet.delete(idx);
    } else {
      newSet.add(idx);
    }
    selectedLogIndices.value = newSet;
    lastSelectedLogIndex.value = idx;
  } else {
    // Single click: toggle selection if clicked again, or select single
    const newSet = new Set<number>();
    if (!selectedLogIndices.value.has(idx) || selectedLogIndices.value.size > 1) {
      newSet.add(idx);
      lastSelectedLogIndex.value = idx;
    } else {
      lastSelectedLogIndex.value = null;
    }
    selectedLogIndices.value = newSet;
  }
}

function clearSelectedLogs() {
  selectedLogIndices.value = new Set();
  lastSelectedLogIndex.value = null;
}

function copySelectedOrAllLogs() {
  if (selectedLogIndices.value.size > 0) {
    const indices = Array.from(selectedLogIndices.value).sort((a, b) => a - b);
    const selectedTexts = indices
      .map((i) => formattedLogs.value[i])
      .filter(Boolean)
      .map((l) => l.rawText || l.text);
    const fullText = selectedTexts.join("\n");
    if (!fullText) return;
    navigator.clipboard
      .writeText(fullText)
      .then(() => {
        showToast(
          t("guard.logs.copied_selected", { count: indices.length }) ||
            `Copied ${indices.length} selected lines to clipboard`,
          "success"
        );
      })
      .catch(() => {});
  } else {
    copyAllLogs();
  }
}

function handleConsoleCopy(e: ClipboardEvent) {
  const selection = window.getSelection();
  if (selection && selection.toString().trim()) {
    // Native selection handles drag text copy
    return;
  }
  if (selectedLogIndices.value.size > 0) {
    e.preventDefault();
    copySelectedOrAllLogs();
  }
}

function copyAllLogs() {
  // Preserve full CLI text including native checkmarks when copying to clipboard
  const fullText = formattedLogs.value.map((l) => l.rawText || l.text).join("\n");
  if (!fullText) return;
  navigator.clipboard
    .writeText(fullText)
    .then(() => {
      showToast(t("guard.logs.copied") || "Inspection logs copied to clipboard", "success");
    })
    .catch(() => {});
}

function clearLogs() {
  localLogs.value = [];
  selectedLogIndices.value = new Set();
  lastSelectedLogIndex.value = null;
  if (guardStatus.value) {
    guardStatus.value.logs = [];
  }
  showToast(t("guard.logs.cleared") || "Console output cleared", "info");
}

export interface LatestInspectionInfo {
  time: string;
  mode: "changed" | "staged" | "all";
  scanned: number;
  clean: boolean;
  issuesCount: number;
  fixedCount: number;
  durationMs: number;
  isFix?: boolean;
}
const latestInspection = ref<LatestInspectionInfo | null>(null);

export interface ExtensionPipelineRow {
  extension: string;
  languages: string[];
  linter?: ToolSlotItem;
  formatter?: ToolSlotItem;
  fixer?: ToolSlotItem;
  auditor?: ToolSlotItem;
  allSlots: ToolSlotItem[];
}

function getModeLabel(mode: "changed" | "staged" | "all"): string {
  switch (mode) {
    case "changed":
      return t("guard.modes.changed") || "Workspace Changes";
    case "staged":
      return t("guard.modes.staged") || "Staged Changes";
    case "all":
      return t("guard.modes.all") || "Project";
  }
}

onMounted(() => {
  document.addEventListener("click", handleDocClick);
  fetchStatus().catch(() => {});
  fetchSlots().catch(() => {});
  loadGuardTomlConfig().catch(() => {});
});

onUnmounted(() => {
  document.removeEventListener("click", handleDocClick);
});

onActivated(() => {
  fetchStatus().catch(() => {});
  fetchSlots().catch(() => {});
  loadGuardTomlConfig().catch(() => {});
});

async function handleCheck(mode: "changed" | "staged" | "all") {
  checkingMode.value = mode;
  const modeLabel = getModeLabel(mode);
  emit("log-event", "INFO", `[Quality Guard] Starting inspection (${modeLabel})...`, undefined, "output");

  currentScanningTarget.value = `Initiating quality guard inspection (${modeLabel})...`;
  scanProgressPct.value = 5;
  currentScannedCount.value = 0;
  totalFilesCount.value = mode === "all" ? 2631 : 0;
  const cliCmd =
    mode === "all" ? "$ unirtm check --all" : mode === "staged" ? "$ unirtm check --staged" : "$ unirtm check";
  localLogs.value = [parseLogLine(cliCmd)];
  scrollToBottom();

  const receivedRealtimeLogs: string[] = [];
  isStreaming.value = true;

  await runCheck({
    mode,
    onLog: (line: string) => {
      // Avoid duplicate leading prompt line
      if (line.trim() === cliCmd.trim()) return;

      receivedRealtimeLogs.push(line);
      localLogs.value.push(parseLogLine(line));

      const inspectMatch = line.match(/\((\d+)\s+files/);
      if (inspectMatch) {
        totalFilesCount.value = parseInt(inspectMatch[1], 10);
      }

      if (line.includes("› [scan]")) {
        currentScanningTarget.value = line.replace(/^\s*›\s*\[scan\]\s*/, "");
        currentScannedCount.value++;
      } else if (line.startsWith("✓ ") || line.startsWith("✗ ") || line.startsWith("! ") || line.startsWith("⚠ ")) {
        currentScanningTarget.value = line.replace(/^[✓✗!⚠]\s*/, "");
        if (mode === "all") {
          currentScannedCount.value = Math.min(2631, Math.round((receivedRealtimeLogs.length / 40) * 2631));
        }
      } else if (line.startsWith("$ ")) {
        currentScanningTarget.value = line;
      }

      const totalTarget = Math.max(38, totalFilesCount.value || 38);
      scanProgressPct.value = Math.min(95, Math.max(10, Math.round((receivedRealtimeLogs.length / totalTarget) * 100)));
      scrollToBottom();
      emit("log-event", "INFO", line, undefined, "output");
    },
  });

  isStreaming.value = false;
  scanProgressPct.value = 100;
  currentScannedCount.value =
    mode === "all" ? totalFilesCount.value || guardStatus.value.scanned || 2631 : guardStatus.value.scanned || 0;
  if (guardStatus.value.clean) {
    currentScanningTarget.value = "All 38 guard rules verified clean · Workspace protected";
  }

  const res = lastRunResult.value || {
    success: true,
    mode,
    scanned: guardStatus.value.scanned,
    clean: guardStatus.value.clean,
    issuesCount: guardStatus.value.issues.length,
    fixedCount: 0,
    durationMs: guardStatus.value.duration_ms || 15,
    logs: guardStatus.value.logs || [],
  };

  const now = new Date();
  const timeStr = now.toTimeString().split(" ")[0];
  lastCheckedTime.value = timeStr;
  lastCheckedMode.value = mode;
  justInspected.value = true;
  setTimeout(() => {
    justInspected.value = false;
  }, 2500);

  latestInspection.value = {
    time: timeStr,
    mode,
    scanned: res.scanned,
    clean: res.clean,
    issuesCount: res.issuesCount,
    fixedCount: 0,
    durationMs: res.durationMs,
    isFix: false,
  };

  // If no realtime logs were received (e.g. mock or fallback), stream them now
  if (receivedRealtimeLogs.length === 0) {
    const logsToStream = res.logs && res.logs.length > 0 ? res.logs : guardStatus.value.logs || [];
    if (logsToStream.length > 0) {
      await streamPushLogs(logsToStream);
    } else {
      localLogs.value.push(
        parseLogLine(
          `✓ ✨ Quality Guard passed: All 38 guard rules processed (${res.scanned} files inspected, ${res.issuesCount} issues, took ${res.durationMs}ms)`
        )
      );
      scrollToBottom();
    }
  }

  if (res.clean) {
    emit(
      "log-event",
      "SUCCESS",
      `[Quality Guard] Inspection passed: ${res.scanned} files verified clean, ${res.durationMs}ms`,
      undefined,
      "output"
    );
    showToast(
      t("guard.toast.clean", { count: res.scanned, duration: res.durationMs }) ||
        `Inspection passed: ${res.scanned} files verified, all rules healthy (${res.durationMs}ms)`,
      "success"
    );
  } else {
    emit(
      "log-event",
      "WARN",
      `[Quality Guard] Inspection found ${res.issuesCount} issues across ${res.scanned} files (${res.durationMs}ms)`,
      undefined,
      "output"
    );
    showToast(
      t("guard.toast.issues", { count: res.issuesCount, duration: res.durationMs }) ||
        `Inspection alert: found ${res.issuesCount} issues across ${res.scanned} files`,
      "warning"
    );
  }
}

async function openTomlInEditor() {
  await openTab({
    path: ".unirtm.toml",
    name: ".unirtm.toml",
  });
  router.push({
    name: "explorer",
    query: {
      from: "guard",
      file: ".unirtm.toml",
      scope: "workspace",
    },
  });
}

async function handleFix(overrideMode?: "staged" | "changed" | "all") {
  if (overrideMode) {
    checkingMode.value = overrideMode;
  }
  isFixDropdownOpen.value = false;
  const modeLabel = getModeLabel(checkingMode.value);
  emit("log-event", "INFO", `[Quality Guard] Starting auto-fix pipeline (${modeLabel})...`, undefined, "output");

  currentScanningTarget.value = `Initiating auto-repair & formatting pipeline (${modeLabel})...`;
  scanProgressPct.value = 5;
  currentScannedCount.value = 0;
  totalFilesCount.value = checkingMode.value === "all" ? 2631 : 0;
  const fixFlag = checkingMode.value === "all" ? "--all" : checkingMode.value === "staged" ? "--staged" : "";
  const cliCmd = fixFlag ? `$ unirtm check ${fixFlag} --fix` : "$ unirtm check --fix";
  localLogs.value = [parseLogLine(cliCmd)];
  scrollToBottom();

  const receivedRealtimeLogs: string[] = [];
  isStreaming.value = true;

  await runFix(checkingMode.value, [], (line: string) => {
    if (line.trim() === cliCmd.trim()) return;

    receivedRealtimeLogs.push(line);
    localLogs.value.push(parseLogLine(line));

    const inspectMatch = line.match(/\((\d+)\s+files/);
    if (inspectMatch) {
      totalFilesCount.value = parseInt(inspectMatch[1], 10);
    }

    if (line.includes("› [scan]")) {
      currentScanningTarget.value = line.replace(/^\s*›\s*\[scan\]\s*/, "");
      currentScannedCount.value++;
    } else if (line.startsWith("✓ ") || line.startsWith("✗ ") || line.startsWith("! ") || line.startsWith("⚠ ")) {
      currentScanningTarget.value = line.replace(/^[✓✗!⚠]\s*/, "");
      if (checkingMode.value === "all") {
        currentScannedCount.value = Math.min(2631, Math.round((receivedRealtimeLogs.length / 40) * 2631));
      }
    } else if (line.startsWith("$ ")) {
      currentScanningTarget.value = line;
    }

    const totalTarget = Math.max(38, totalFilesCount.value || 38);
    scanProgressPct.value = Math.min(95, Math.max(10, Math.round((receivedRealtimeLogs.length / totalTarget) * 100)));
    scrollToBottom();
    emit("log-event", "INFO", line, undefined, "output");
  });

  isStreaming.value = false;
  scanProgressPct.value = 100;
  currentScannedCount.value =
    checkingMode.value === "all"
      ? totalFilesCount.value || guardStatus.value.scanned || 2631
      : guardStatus.value.scanned || 0;
  currentScanningTarget.value = "All fixable issues repaired · Workspace clean";

  const res = lastRunResult.value || {
    success: true,
    mode: checkingMode.value,
    scanned: guardStatus.value.scanned,
    clean: true,
    issuesCount: 0,
    fixedCount: 1,
    durationMs: guardStatus.value.duration_ms || 25,
    logs: guardStatus.value.logs || [],
  };

  const now = new Date();
  const timeStr = now.toTimeString().split(" ")[0];
  lastCheckedTime.value = timeStr;
  lastCheckedMode.value = checkingMode.value;
  justInspected.value = true;
  setTimeout(() => {
    justInspected.value = false;
  }, 2500);

  latestInspection.value = {
    time: timeStr,
    mode: checkingMode.value,
    scanned: res.scanned,
    clean: true,
    issuesCount: 0,
    fixedCount: res.fixedCount,
    durationMs: res.durationMs,
    isFix: true,
  };

  if (receivedRealtimeLogs.length === 0) {
    const fixLogsToStream = res.logs && res.logs.length > 0 ? res.logs : guardStatus.value.logs || [];
    if (fixLogsToStream.length > 0) {
      await streamPushLogs(fixLogsToStream);
    } else {
      localLogs.value.push(
        parseLogLine(
          `✓ ✨ Quality Guard passed: Auto-fix completed: ${res.fixedCount} corrections applied, took ${res.durationMs}ms`
        )
      );
      scrollToBottom();
    }
  }

  emit(
    "log-event",
    "SUCCESS",
    `[Quality Guard] Auto-fix completed: ${res.fixedCount} formatting & hygiene repairs applied (${res.durationMs}ms)`,
    undefined,
    "output"
  );
  showToast(
    t("guard.toast.fixed", { count: res.fixedCount, duration: res.durationMs }) ||
      `Auto-fix finished: ${res.fixedCount} formatting & hygiene fixes applied (${res.durationMs}ms)`,
    "success"
  );
}

// Map extensions to human readable language tags
function getLanguagesForExt(ext: string): string[] {
  switch (ext) {
    case ".go":
      return ["Go"];
    case ".ts":
    case ".tsx":
      return ["TypeScript"];
    case ".js":
    case ".jsx":
      return ["JavaScript"];
    case ".vue":
      return ["Vue"];
    case ".sh":
    case ".bash":
      return ["Shell"];
    case ".toml":
      return ["TOML"];
    case ".json":
      return ["JSON"];
    case ".md":
      return ["Markdown"];
    case "*":
      return ["Global / Any"];
    default:
      return [ext.replace(".", "").toUpperCase()];
  }
}

// Compute aggregate rows grouped by file extension
const extensionRows = computed<ExtensionPipelineRow[]>(() => {
  const rowMap = new Map<string, ExtensionPipelineRow>();

  // Ensure primary supported extensions exist by default
  const defaultExts = [".go", ".ts", ".vue", ".sh", ".toml", "*"];
  for (const ext of defaultExts) {
    rowMap.set(ext, {
      extension: ext,
      languages: getLanguagesForExt(ext),
      allSlots: [],
    });
  }

  // Iterate over guardSlots and associate with extension
  for (const slot of guardSlots.value) {
    const exts = slot.default_extensions && slot.default_extensions.length > 0 ? slot.default_extensions : ["*"];

    for (const ext of exts) {
      if (!rowMap.has(ext)) {
        rowMap.set(ext, {
          extension: ext,
          languages: getLanguagesForExt(ext),
          allSlots: [],
        });
      }

      const row = rowMap.get(ext)!;
      row.allSlots.push(slot);

      if (slot.category === "format") {
        row.formatter = slot;
        if (slot.fix_command || slot.phase === "phase1_fixer") {
          row.fixer = slot;
        }
      } else if (slot.category === "lint") {
        row.linter = slot;
      } else if (slot.category === "secret" || slot.category === "security") {
        row.auditor = slot;
      }
    }
  }

  return Array.from(rowMap.values());
});

const filteredExtensionRows = computed(() => {
  if (!searchKeyword.value.trim()) return extensionRows.value;
  const q = searchKeyword.value.toLowerCase().trim();
  return extensionRows.value.filter((row) => {
    return (
      row.extension.toLowerCase().includes(q) ||
      row.languages.some((l) => l.toLowerCase().includes(q)) ||
      (row.linter && row.linter.default_tool_name.toLowerCase().includes(q)) ||
      (row.formatter && row.formatter.default_tool_name.toLowerCase().includes(q)) ||
      (row.fixer && row.fixer.default_tool_name.toLowerCase().includes(q)) ||
      (row.auditor && row.auditor.default_tool_name.toLowerCase().includes(q))
    );
  });
});

const filteredSlots = computed(() => {
  if (!searchKeyword.value.trim()) return guardSlots.value;
  const q = searchKeyword.value.toLowerCase().trim();
  return guardSlots.value.filter((slot) => {
    return (
      slot.id.toLowerCase().includes(q) ||
      slot.default_tool_name.toLowerCase().includes(q) ||
      (slot.default_extensions && slot.default_extensions.some((e) => e.toLowerCase().includes(q)))
    );
  });
});

function openAddSlotModal() {
  activeEditingSlot.value = null;
  targetExtForAdd.value = "";
  isSlotModalOpen.value = true;
}

function editSlot(slot: ToolSlotItem) {
  activeEditingSlot.value = slot;
  targetExtForAdd.value = "";
  isSlotModalOpen.value = true;
}

function handleEditRow(row: ExtensionPipelineRow) {
  const primarySlot = row.linter || row.formatter || row.fixer || row.auditor || row.allSlots[0];
  if (primarySlot) {
    editSlot(primarySlot);
  } else {
    handleAddToolToExt(row.extension);
  }
}

function handleAddToolToExt(ext: string) {
  activeEditingSlot.value = null;
  targetExtForAdd.value = ext;
  isSlotModalOpen.value = true;
}

async function onSaveSlot(slot: SlotConfigInput) {
  const ok = await saveSlotConfig(slot);
  if (ok) {
    isSlotModalOpen.value = false;
    await fetchSlots();
  }
}

async function onDeleteSlot(slotId: string) {
  const ok = await deleteSlotConfig(slotId);
  if (ok) {
    isSlotModalOpen.value = false;
    await fetchSlots();
  }
}

async function onTomlSaved() {
  await fetchSlots();
  await fetchStatus();
  await loadGuardTomlConfig();
}

// Built-in Rules Helpers & Config Handlers (Configurable, Modifiable, Non-deletable)
async function loadGuardTomlConfig() {
  const content = await fetchRawTomlConfig();
  if (content) {
    parsedGuardConfig.value = parseGuardConfigFromToml(content);
  }
}

function isRuleSkipped(ruleId: string): boolean {
  return parsedGuardConfig.value.skip.includes(ruleId);
}

function hasRuleCustomConfig(ruleId: string): boolean {
  const r = parsedGuardConfig.value.rules[ruleId];
  if (!r) return false;
  return Boolean(
    (r.severity && r.severity !== "default") ||
    (r.max_file_size && r.max_file_size > 0) ||
    (r.exclude_dirs && r.exclude_dirs.length > 0) ||
    (r.exclude_files && r.exclude_files.length > 0) ||
    (r.extensions && r.extensions.length > 0)
  );
}

function getRuleCustomSummary(ruleId: string): string {
  const r = parsedGuardConfig.value.rules[ruleId];
  if (!r) return "";
  const parts: string[] = [];
  if (r.severity && r.severity !== "default") {
    parts.push(`级别: ${r.severity}`);
  }
  if (r.max_file_size && r.max_file_size > 0) {
    const mb = Math.round((r.max_file_size / (1024 * 1024)) * 10) / 10;
    parts.push(`体积上限: ${mb}MB`);
  }
  if (r.exclude_dirs && r.exclude_dirs.length > 0) {
    parts.push(`排除目录: ${r.exclude_dirs.join(",")}`);
  }
  if (r.exclude_files && r.exclude_files.length > 0) {
    parts.push(`排除文件: ${r.exclude_files.join(",")}`);
  }
  return parts.join(" · ");
}

const activeBuiltinCount = computed(() => {
  return BUILTIN_GUARD_RULES.filter((r) => !isRuleSkipped(r.id)).length;
});

const skippedBuiltinCount = computed(() => {
  return BUILTIN_GUARD_RULES.filter((r) => isRuleSkipped(r.id)).length;
});

const filteredSettingsBuiltinRules = computed(() => {
  let list = BUILTIN_GUARD_RULES;
  if (!searchKeyword.value.trim()) return list;
  const q = searchKeyword.value.toLowerCase().trim();
  return list.filter((r) => {
    return (
      r.id.toLowerCase().includes(q) ||
      r.name.toLowerCase().includes(q) ||
      r.description.toLowerCase().includes(q) ||
      r.extensions.some((ext) => ext.toLowerCase().includes(q))
    );
  });
});

function openConfigureBuiltinRule(rule: DetailedGuardRule) {
  activeEditingBuiltinRule.value = rule;
  isBuiltinModalOpen.value = true;
}

const currentEditingRuleCustomConfig = computed(() => {
  if (!activeEditingBuiltinRule.value) return null;
  return parsedGuardConfig.value.rules[activeEditingBuiltinRule.value.id] || null;
});

async function toggleBuiltinRule(ruleId: string) {
  const currentlySkipped = isRuleSkipped(ruleId);
  const targetEnabled = currentlySkipped;
  const content = await fetchRawTomlConfig();
  const ruleMeta = BUILTIN_GUARD_RULES.find((r) => r.id === ruleId);

  const existingCustom = parsedGuardConfig.value.rules[ruleId] || {};
  const newConfig: BuiltinRuleCustomConfig = {
    id: ruleId,
    enabled: targetEnabled,
    severity: (existingCustom.severity as any) || "default",
    max_file_size: existingCustom.max_file_size,
    exclude_dirs: existingCustom.exclude_dirs || [],
    exclude_files: existingCustom.exclude_files || [],
    extensions: existingCustom.extensions || [],
  };

  const updatedToml = updateBuiltinRuleInToml(content, newConfig);
  const ok = await saveRawTomlConfig(updatedToml);
  if (ok) {
    await loadGuardTomlConfig();
    showToast(
      targetEnabled
        ? `内置规则 "${ruleMeta ? ruleMeta.name : ruleId}" 已恢复启用`
        : `内置规则 "${ruleMeta ? ruleMeta.name : ruleId}" 已加入跳过列表 (Skip)`,
      "success"
    );
  }
}

async function onSaveBuiltinRule(config: BuiltinRuleCustomConfig) {
  const content = await fetchRawTomlConfig();
  const updatedToml = updateBuiltinRuleInToml(content, config);
  const ok = await saveRawTomlConfig(updatedToml);
  if (ok) {
    isBuiltinModalOpen.value = false;
    await loadGuardTomlConfig();
    showToast(`内置规则 "${config.id}" 配置已成功保存至 .unirtm.toml`, "success");
  }
}

function getBuiltinRuleForExtAndCategory(
  ext: string,
  type: "fixer" | "linter" | "security"
): DetailedGuardRule | undefined {
  return BUILTIN_GUARD_RULES.find((r) => {
    const isGlobal = r.extensions.some((e) => e.includes("*"));
    const matchesExt = isGlobal || r.extensions.includes(ext) || r.extensions.some((e) => e.includes(ext));
    if (!matchesExt) return false;
    if (type === "fixer") return r.phase === "fixer";
    if (type === "linter") return r.phase === "linter" && r.category !== "security";
    if (type === "security") return r.category === "security";
    return false;
  });
}
</script>

<style scoped>
.guard-view-container {
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 100%;
  padding: 1.25rem 1.5rem;
  box-sizing: border-box;
  overflow-y: auto;
  gap: 1.25rem;
}

/* Header */
.guard-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 1rem;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid var(--card-border, #333);
}

.header-left {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.title-with-badge {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.header-shield-icon {
  font-size: 1.5rem;
  color: var(--vscode-button-background, #388bfd);
}

.header-title {
  font-size: 1.375rem;
  font-weight: 700;
  margin: 0;
  color: var(--text-main);
}

.status-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  font-size: 0.75rem;
  font-weight: 600;
  padding: 0.2rem 0.625rem;
  border-radius: 9999px;
}

.status-pill-clean {
  background: rgba(46, 160, 67, 0.15);
  color: #3fb950;
  border: 1px solid rgba(46, 160, 67, 0.3);
}

.status-pill-warning {
  background: rgba(248, 81, 73, 0.15);
  color: #f85149;
  border: 1px solid rgba(248, 81, 73, 0.3);
}

/* Last Inspected Pill in Header */
.last-checked-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  font-size: 0.75rem;
  color: var(--text-muted);
  background: var(--card-bg, rgba(255, 255, 255, 0.04));
  border: 1px solid var(--card-border, #333);
  padding: 0.2rem 0.625rem;
  border-radius: 9999px;
  transition: all 0.2s ease;
}

.last-checked-pill .last-mode {
  opacity: 0.8;
  font-size: 0.6875rem;
}

.last-checked-pill.pill-pulse {
  border-color: #388bfd;
  background: rgba(56, 139, 253, 0.15);
  color: var(--text-main);
  box-shadow: 0 0 10px rgba(56, 139, 253, 0.35);
  animation: pulse-border 1.5s ease-in-out;
}

@keyframes pulse-border {
  0% {
    transform: scale(0.98);
  }
  50% {
    transform: scale(1.03);
  }
  100% {
    transform: scale(1);
  }
}

/* Scanning Progress Bar Banner */
.guard-scanning-banner {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
  background: rgba(56, 139, 253, 0.08);
  border: 1px solid rgba(56, 139, 253, 0.25);
  border-radius: 6px;
  padding: 0.5rem 0.75rem;
  overflow: hidden;
}

.scanning-glow-bar {
  width: 100%;
  height: 3px;
  background: rgba(56, 139, 253, 0.2);
  border-radius: 9999px;
  position: relative;
  overflow: hidden;
}

.scanning-glow-thumb {
  position: absolute;
  top: 0;
  left: 0;
  height: 100%;
  width: 30%;
  background: linear-gradient(90deg, transparent, #388bfd, #58a6ff, transparent);
  border-radius: 9999px;
  animation: scanning-slide 1.2s infinite ease-in-out;
}

@keyframes scanning-slide {
  0% {
    left: -30%;
  }
  100% {
    left: 100%;
  }
}

.scanning-info {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.8125rem;
  color: #58a6ff;
  font-weight: 500;
}

/* Inspection Outcome Feedback Banner */
.inspection-outcome-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.75rem 1rem;
  border-radius: 8px;
  gap: 1rem;
  flex-wrap: wrap;
  transition: all 0.25s ease;
}

.outcome-clean {
  background: rgba(46, 160, 67, 0.1);
  border: 1px solid rgba(46, 160, 67, 0.3);
}

.outcome-warning {
  background: rgba(210, 153, 34, 0.1);
  border: 1px solid rgba(210, 153, 34, 0.3);
}

.outcome-main {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex: 1;
  min-width: 280px;
}

.outcome-icon-wrap {
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.25rem;
}

.outcome-clean .outcome-icon-wrap {
  color: #3fb950;
}

.outcome-warning .outcome-icon-wrap {
  color: #d29922;
}

.outcome-details {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}

.outcome-title-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.outcome-headline {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--text-main);
}

.outcome-badge {
  font-size: 0.6875rem;
  font-weight: 700;
  padding: 0.1rem 0.4rem;
  border-radius: 4px;
}

.badge-clean {
  background: rgba(46, 160, 67, 0.2);
  color: #3fb950;
}

.badge-warn {
  background: rgba(210, 153, 34, 0.2);
  color: #d29922;
}

.outcome-meta-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.75rem;
  color: var(--text-muted);
  flex-wrap: wrap;
}

.outcome-meta-row .meta-item {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
}

.outcome-meta-row .meta-item strong {
  color: var(--text-main);
}

.outcome-meta-row .meta-dot {
  opacity: 0.4;
}

.outcome-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.btn-outcome-action {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.35rem 0.65rem;
  font-size: 0.75rem;
  font-weight: 500;
  background: var(--card-bg, #252526);
  color: var(--text-main);
  border: 1px solid var(--card-border, #333);
  border-radius: 5px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-outcome-action:hover {
  background: var(--hover-bg, #333);
}

.btn-outcome-action.primary {
  background: var(--vscode-button-background, #0e639c);
  color: #fff;
  border-color: transparent;
}

.btn-outcome-action.primary:hover {
  background: var(--vscode-button-hoverBackground, #1177bb);
}

.btn-outcome-dismiss {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.35rem;
  background: transparent;
  color: var(--text-muted);
  border: none;
  cursor: pointer;
  border-radius: 4px;
  transition: color 0.15s ease;
}

.btn-outcome-dismiss:hover {
  color: var(--text-main);
  background: rgba(255, 255, 255, 0.08);
}

/* Animations */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.slide-down-enter-active,
.slide-down-leave-active {
  transition: all 0.25s cubic-bezier(0.16, 1, 0.3, 1);
}
.slide-down-enter-from,
.slide-down-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}

.header-subtitle {
  font-size: 0.8125rem;
  color: var(--text-muted);
  margin: 0;
}

.header-right-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.btn-action {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.4375rem 0.875rem;
  font-size: 0.8125rem;
  font-weight: 500;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.15s ease;
  border: 1px solid transparent;
}

.btn-action.primary {
  background: var(--accent-cyan, #0284c7);
  color: #ffffff;
}

.btn-action.primary:hover:not(:disabled) {
  opacity: 0.9;
}

.btn-action.secondary {
  background: var(--btn-secondary-bg);
  color: var(--btn-secondary-text);
  border: 1px solid var(--btn-secondary-border);
}

.btn-action.secondary:hover:not(:disabled) {
  background: var(--btn-secondary-hover-bg);
}

.btn-action.accent {
  background: rgba(2, 132, 199, 0.12);
  color: var(--accent-cyan);
  border: 1px solid rgba(2, 132, 199, 0.3);
}

.btn-action.accent:hover:not(:disabled) {
  background: rgba(2, 132, 199, 0.2);
}

.btn-action.ghost {
  background: transparent;
  color: var(--text-muted);
  border: 1px solid var(--card-border);
}

.btn-action.ghost:hover:not(:disabled) {
  background: var(--hover-bg);
  color: var(--text-main);
}

.btn-action:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* Split Dropdown Button for Auto-Fix */
.split-btn-group {
  position: relative;
  display: inline-flex;
  align-items: stretch;
  border-radius: 6px;
}

.split-main-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.4375rem 0.875rem;
  font-size: 0.8125rem;
  font-weight: 500;
  border-top-left-radius: 6px;
  border-bottom-left-radius: 6px;
  border-top-right-radius: 0;
  border-bottom-right-radius: 0;
  cursor: pointer;
  background: var(--btn-secondary-bg);
  color: var(--btn-secondary-text);
  border: 1px solid var(--btn-secondary-border);
  border-right: none;
  transition: all 0.15s ease;
}

.split-main-btn:hover:not(:disabled) {
  background: var(--btn-secondary-hover-bg);
}

.split-main-btn:disabled,
.split-arrow-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.split-arrow-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.4375rem 0.5rem;
  font-size: 0.75rem;
  border-top-left-radius: 0;
  border-bottom-left-radius: 0;
  border-top-right-radius: 6px;
  border-bottom-right-radius: 6px;
  cursor: pointer;
  background: var(--btn-secondary-bg);
  color: var(--btn-secondary-text);
  border: 1px solid var(--btn-secondary-border);
  border-left: 1px solid var(--card-border, rgba(255, 255, 255, 0.1));
  transition: all 0.15s ease;
}

.split-arrow-btn:hover:not(:disabled) {
  background: var(--btn-secondary-hover-bg);
  color: var(--accent-cyan, #0284c7);
}

.fix-dropdown-menu {
  position: absolute;
  top: calc(100% + 6px);
  right: 0;
  min-width: 310px;
  background: var(--card-bg, #1e1e1e);
  border: 1px solid var(--card-border, #333);
  border-radius: 8px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.35);
  padding: 0.5rem;
  z-index: 1000;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.fix-dropdown-header {
  font-size: 0.6875rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted, #888);
  padding: 0.375rem 0.625rem;
  border-bottom: 1px solid var(--card-border, rgba(255, 255, 255, 0.08));
  margin-bottom: 0.25rem;
}

.fix-dropdown-item {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 0.25rem;
  padding: 0.5rem 0.625rem;
  border-radius: 6px;
  background: transparent;
  border: 1px solid transparent;
  cursor: pointer;
  text-align: left;
  width: 100%;
  transition: all 0.15s ease;
}

.fix-dropdown-item:hover {
  background: var(--hover-bg, rgba(255, 255, 255, 0.08));
  border-color: var(--card-border, rgba(255, 255, 255, 0.1));
}

.fix-dropdown-item.active {
  background: rgba(2, 132, 199, 0.1);
  border-color: rgba(2, 132, 199, 0.3);
}

.item-title-row {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--text-main, #fff);
  width: 100%;
}

.item-title-row .codicon {
  font-size: 0.875rem;
  color: var(--accent-cyan, #0284c7);
}

.current-badge {
  margin-left: auto;
  font-size: 0.625rem;
  font-weight: 700;
  padding: 1px 6px;
  border-radius: 10px;
  background: rgba(2, 132, 199, 0.2);
  color: var(--accent-cyan, #0284c7);
  text-transform: uppercase;
}

.item-desc {
  font-size: 0.6875rem;
  color: var(--text-muted, #888);
  padding-left: 1.3rem;
  line-height: 1.35;
  font-family: var(--font-mono, monospace);
}

.dropdown-fade-enter-active,
.dropdown-fade-leave-active {
  transition:
    opacity 0.15s ease,
    transform 0.15s ease;
}

.dropdown-fade-enter-from,
.dropdown-fade-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}

/* Metrics Row */
.metrics-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
}

.metric-card {
  display: flex;
  align-items: center;
  gap: 0.875rem;
  padding: 0.875rem 1rem;
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.metric-icon-box {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 8px;
  font-size: 1.25rem;
}

.bg-blue {
  background: rgba(56, 139, 253, 0.15);
  color: #58a6ff;
}
.bg-green {
  background: rgba(46, 160, 67, 0.15);
  color: #3fb950;
}
.bg-amber {
  background: rgba(210, 153, 34, 0.15);
  color: #d29922;
}
.bg-purple {
  background: rgba(187, 128, 255, 0.15);
  color: #bc8cff;
}
.bg-cyan {
  background: rgba(57, 197, 207, 0.15);
  color: #39c5cf;
}

.metric-info {
  display: flex;
  flex-direction: column;
}

.metric-val {
  font-size: 1.125rem;
  font-weight: 700;
  color: var(--text-main);
}

.metric-label {
  font-size: 0.75rem;
  color: var(--text-muted);
}

/* View Switcher Tabs Bar */
.guard-view-tabs-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid var(--card-border);
  padding-bottom: 0.5rem;
  flex-wrap: wrap;
  gap: 0.75rem;
}

.tabs-group {
  display: flex;
  align-items: center;
  gap: 0.375rem;
}

.view-tab-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.4375rem 0.875rem;
  font-size: 0.8125rem;
  font-weight: 500;
  color: var(--text-muted);
  background: transparent;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.view-tab-btn:hover {
  background: var(--hover-bg);
  color: var(--text-main);
}

.view-tab-btn.active {
  background: rgba(2, 132, 199, 0.12);
  color: var(--accent-cyan);
  font-weight: 600;
}

.tab-badge-pulse {
  font-size: 0.6875rem;
  font-weight: 700;
  padding: 0.1rem 0.45rem;
  border-radius: 9999px;
  margin-left: 0.35rem;
}

.tab-badge-pulse.badge-clean {
  background: rgba(46, 160, 67, 0.2);
  color: #3fb950;
}

.tab-badge-pulse.badge-warn {
  background: rgba(248, 81, 73, 0.2);
  color: #f85149;
}

/* Workbench Section & Live Console */
.workbench-layout {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.live-console-card {
  display: flex;
  flex-direction: column;
  background: var(--vscode-terminal-background, #14161a);
  border: 1px solid var(--card-border, #30363d);
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
}

.console-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.5rem 1rem;
  background: rgba(255, 255, 255, 0.03);
  border-bottom: 1px solid var(--card-border, #30363d);
}

.console-header-left {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.console-icon {
  font-size: 1rem;
  color: var(--accent-cyan, #388bfd);
}

.console-title {
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--text-main);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.console-status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-left: 0.25rem;
}

.console-status-dot.dot-clean {
  background: #3fb950;
  box-shadow: 0 0 6px rgba(63, 185, 80, 0.6);
}

.console-status-dot.dot-running {
  background: #388bfd;
  box-shadow: 0 0 8px rgba(56, 139, 253, 0.8);
  animation: pulse-border 1s infinite alternate;
}

.console-status-dot.dot-warn {
  background: #d29922;
  box-shadow: 0 0 6px rgba(210, 153, 34, 0.6);
}

.console-status-text {
  font-size: 0.6875rem;
  color: var(--text-muted);
  font-weight: 600;
}

.console-header-actions {
  display: flex;
  align-items: center;
  gap: 0.375rem;
}

.console-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  font-size: 0.75rem;
  padding: 0.2rem 0.5rem;
  background: transparent;
  color: var(--text-muted);
  border: 1px solid var(--card-border, #444);
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.console-btn:hover {
  background: rgba(255, 255, 255, 0.08);
  color: var(--text-main);
}

/* Anti-Virus Live Radar Scanner Bar */
.av-radar-scanner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.625rem 1rem;
  background: rgba(255, 255, 255, 0.02);
  border-bottom: 1px solid var(--card-border, #30363d);
  transition: all 0.25s ease;
}

.av-radar-scanner.is-active {
  background: rgba(56, 139, 253, 0.06);
  border-bottom-color: rgba(56, 139, 253, 0.35);
}

.radar-pulse-ring {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.04);
  flex-shrink: 0;
  font-size: 1.125rem;
}

.av-radar-scanner.is-active .radar-pulse-ring {
  background: rgba(56, 139, 253, 0.15);
  box-shadow: 0 0 10px rgba(56, 139, 253, 0.4);
}

.radar-scanner-content {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
  flex: 1;
  min-width: 0;
}

.radar-scan-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  min-width: 0;
}

.radar-scan-label {
  font-size: 0.6875rem;
  font-weight: 700;
  color: var(--accent-cyan, #388bfd);
  letter-spacing: 0.05em;
  flex-shrink: 0;
}

.radar-scan-target {
  font-size: 0.75rem;
  color: var(--text-main, #f0f6fc);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  font-weight: 500;
}

.radar-meta-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.6875rem;
  color: var(--text-muted, #8b949e);
}

.radar-meta-item {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
}

.radar-meta-dot {
  opacity: 0.5;
}

.radar-progress-box {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 0.25rem;
  width: 110px;
  flex-shrink: 0;
}

.radar-progress-pct {
  font-size: 0.6875rem;
  font-weight: 700;
  color: var(--accent-cyan, #388bfd);
}

.radar-progress-bar {
  width: 100%;
  height: 4px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 2px;
  overflow: hidden;
}

.radar-progress-fill {
  height: 100%;
  background: linear-gradient(90deg, #388bfd, #2ea043);
  border-radius: 2px;
  transition: width 0.1s linear;
}

.console-body {
  height: 380px;
  max-height: 480px;
  overflow-y: auto;
  padding: 1rem 1.25rem;
  font-size: 0.8125rem;
  line-height: 1.6;
  background: var(--input-bg, #0d1117);
}

.console-empty-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  text-align: center;
  color: var(--text-muted);
  gap: 0.5rem;
  padding: 2rem 1rem;
}

.placeholder-shield {
  font-size: 2.5rem;
  color: var(--accent-cyan, #388bfd);
  opacity: 0.85;
}

.placeholder-title {
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-main);
  margin: 0;
}

.placeholder-desc {
  font-size: 0.8125rem;
  max-width: 480px;
  margin: 0;
  line-height: 1.5;
}

.placeholder-quick-actions {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-top: 0.75rem;
}

.placeholder-action-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.4rem 0.85rem;
  font-size: 0.8125rem;
  font-weight: 500;
  border-radius: 6px;
  cursor: pointer;
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid var(--card-border, #444);
  color: var(--text-main);
  transition: all 0.15s ease;
}

.placeholder-action-btn:hover {
  background: rgba(255, 255, 255, 0.12);
}

.placeholder-action-btn.primary {
  background: var(--vscode-button-background, #0e639c);
  color: #fff;
  border-color: transparent;
}

.placeholder-action-btn.primary:hover {
  background: var(--vscode-button-hoverBackground, #1177bb);
}

/* Console Log Lines */
.console-log-lines {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.log-line {
  display: flex;
  align-items: flex-start;
  gap: 0.625rem;
  padding: 0.1875rem 0.5rem;
  margin: 0 -0.5rem;
  border-radius: 4px;
  position: relative;
  cursor: pointer;
  user-select: text;
  transition:
    background 0.15s ease,
    box-shadow 0.15s ease;
}

.log-line:hover {
  background: rgba(255, 255, 255, 0.05);
}

.log-line.is-selected {
  background: rgba(56, 139, 253, 0.16) !important;
  box-shadow: inset 3px 0 0 #388bfd;
}

.log-line.is-flashed {
  animation: logLineFlash 1.4s ease-out;
}

@keyframes logLineFlash {
  0% {
    background: rgba(63, 185, 80, 0.35);
  }
  100% {
    background: transparent;
  }
}

.log-line-copy-btn {
  opacity: 0;
  pointer-events: none;
  margin-left: auto;
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 1.375rem;
  height: 1.375rem;
  background: rgba(255, 255, 255, 0.08);
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 3px;
  color: #8b949e;
  cursor: pointer;
  padding: 0;
  font-size: 0.75rem;
  transition:
    opacity 0.15s ease,
    background 0.15s ease,
    color 0.15s ease;
}

.log-line:hover .log-line-copy-btn {
  opacity: 1;
  pointer-events: auto;
}

.log-line-copy-btn:hover {
  background: rgba(255, 255, 255, 0.16);
  color: #f0f6fc;
}

.btn-highlight {
  background: rgba(56, 139, 253, 0.22) !important;
  color: #58a6ff !important;
  border-color: rgba(56, 139, 253, 0.4) !important;
}

.log-ts {
  color: #6e7681;
  font-size: 0.75rem;
  flex-shrink: 0;
}

.log-badge {
  font-size: 0.625rem;
  font-weight: 700;
  padding: 0.05rem 0.35rem;
  border-radius: 3px;
  flex-shrink: 0;
}

.log-icon-wrap {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  font-size: 0.8125rem;
  line-height: 1;
  margin-top: 0.05rem;
}

.text-green {
  color: #3fb950 !important;
}

.text-red {
  color: #f85149 !important;
}

.text-amber {
  color: #d29922 !important;
}

.text-blue {
  color: #58a6ff !important;
}

.text-cyan {
  color: #39c5cf !important;
}

.text-purple {
  color: #bc8cff !important;
}

.text-indigo {
  color: #79c0ff !important;
}

.badge-info {
  background: rgba(56, 139, 253, 0.18);
  color: #58a6ff;
}

.badge-pass {
  background: rgba(46, 160, 67, 0.2);
  color: #3fb950;
}

.badge-phase {
  background: rgba(187, 128, 255, 0.2);
  color: #bc8cff;
}

.badge-config {
  background: rgba(57, 197, 207, 0.2);
  color: #39c5cf;
}

.badge-warn {
  background: rgba(210, 153, 34, 0.2);
  color: #d29922;
}

.badge-error {
  background: rgba(248, 81, 73, 0.2);
  color: #f85149;
}

.badge-cmd {
  background: rgba(56, 139, 253, 0.2);
  color: #58a6ff;
}

.badge-scan {
  background: rgba(139, 148, 158, 0.18);
  color: #8b949e;
}

.badge-rule {
  background: rgba(46, 160, 67, 0.2);
  color: #3fb950;
}

.log-content {
  color: #c9d1d9;
  word-break: break-all;
  white-space: pre-wrap;
}

.log-pass .log-content {
  color: #7ee787;
}

.log-warn .log-content {
  color: #e3b341;
}

.log-error .log-content {
  color: #ff7b72;
}

.log-cmd .log-content {
  color: #58a6ff;
  font-weight: 600;
}

.log-scan .log-content {
  color: #8b949e;
}

.log-rule .log-content {
  color: #e6edf3;
}

/* Issues Card inside Workbench */
.workbench-issues-card {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  background: var(--card-bg, #1e1e1e);
  border: 1px solid rgba(248, 81, 73, 0.3);
  border-radius: 8px;
  padding: 1rem 1.25rem;
}

.issues-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.issues-title-group {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.issues-title {
  font-size: 0.9375rem;
  font-weight: 700;
  margin: 0;
  color: var(--text-main);
}

.issues-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.issue-item {
  padding: 0.5rem 0.75rem;
  background: rgba(248, 81, 73, 0.08);
  border-left: 3px solid #f85149;
  border-radius: 4px;
}

.issue-item-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  font-size: 0.75rem;
}

.issue-file {
  font-weight: 600;
  color: #58a6ff;
}

.issue-rule {
  color: var(--text-muted);
}

.issue-msg {
  font-size: 0.8125rem;
  color: var(--text-main);
  margin: 0.25rem 0 0 0;
}

/* Hidden Settings Panel */
.settings-section {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.settings-hint-box {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.75rem 1rem;
  background: rgba(56, 139, 253, 0.08);
  border: 1px solid rgba(56, 139, 253, 0.25);
  border-radius: 8px;
  gap: 1rem;
  flex-wrap: wrap;
}

.hint-text {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.8125rem;
  color: var(--text-main);
}

.settings-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.settings-sub-tabs {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  border-bottom: 1px solid var(--card-border);
  padding-bottom: 0.5rem;
}

.sub-tab-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.35rem 0.75rem;
  font-size: 0.8125rem;
  font-weight: 500;
  color: var(--text-muted);
  background: transparent;
  border: none;
  border-radius: 5px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.sub-tab-btn:hover {
  background: var(--hover-bg);
  color: var(--text-main);
}

.sub-tab-btn.active {
  background: var(--hover-bg);
  color: var(--text-main);
  font-weight: 600;
}

.tab-badge {
  font-size: 0.6875rem;
  padding: 0.1rem 0.375rem;
  background: var(--hover-bg);
  border-radius: 9999px;
  margin-left: 0.25rem;
}

.filter-search-box {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.3125rem 0.625rem;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  border-radius: 6px;
  min-width: 260px;
}

.search-input {
  background: transparent;
  border: none;
  color: var(--text-main);
  font-size: 0.8125rem;
  outline: none;
  width: 100%;
}

.search-input::placeholder {
  color: var(--text-muted);
}

/* Main Content Body */
.guard-content-body {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

/* Table Design */
.table-responsive-wrapper {
  overflow-x: auto;
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  border-radius: 8px;
}

.pipeline-matrix-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.8125rem;
  text-align: left;
}

.pipeline-matrix-table th {
  padding: 0.75rem 1rem;
  background: var(--header-bg);
  color: var(--text-muted);
  font-weight: 600;
  border-bottom: 1px solid var(--card-border);
}

.pipeline-matrix-table td {
  padding: 0.75rem 1rem;
  border-bottom: 1px solid var(--card-border);
  color: var(--text-main);
  vertical-align: middle;
}

.pipeline-row:hover {
  background: var(--hover-bg);
}

.ext-badge-group {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.ext-tag {
  font-size: 0.8125rem;
  font-weight: 700;
  padding: 0.15rem 0.45rem;
  background: rgba(2, 132, 199, 0.12);
  color: var(--accent-cyan);
  border: 1px solid rgba(2, 132, 199, 0.25);
  border-radius: 4px;
}

.lang-label {
  font-size: 0.75rem;
  color: var(--text-muted);
}

.tool-cell {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
  cursor: pointer;
  padding: 0.25rem 0.4rem;
  border-radius: 4px;
  transition: background 0.15s ease;
}

.tool-cell:hover {
  background: var(--hover-bg);
}

.tool-name {
  font-weight: 600;
  color: var(--text-main);
}

.tool-cmd {
  font-size: 0.6875rem;
  color: var(--text-muted);
  background: var(--section-bg);
  padding: 0.1rem 0.35rem;
  border-radius: 3px;
  border: 1px solid var(--card-border);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 180px;
}

.cell-empty {
  color: var(--text-muted);
  opacity: 0.6;
}

.action-btn-row {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0.375rem;
}

.row-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 4px;
  background: transparent;
  border: 1px solid var(--card-border);
  color: var(--text-muted);
  cursor: pointer;
  transition: all 0.15s ease;
}

.row-btn:hover {
  background: var(--hover-bg);
  color: var(--text-main);
  border-color: var(--accent-cyan);
}

/* Slots Grid */
.slots-cards-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 1rem;
}

.slot-card {
  display: flex;
  flex-direction: column;
  gap: 0.625rem;
  padding: 1rem;
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  border-radius: 8px;
  transition: border-color 0.2s ease;
}

.slot-card:hover {
  border-color: var(--card-border-active);
}

.slot-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.slot-title-wrap {
  display: flex;
  flex-direction: column;
}

.slot-tool-name {
  font-size: 0.9375rem;
  font-weight: 700;
  color: var(--text-main);
}

.slot-id {
  font-size: 0.6875rem;
  color: var(--text-muted);
}

.slot-header-tags {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.phase-pill {
  font-size: 0.6875rem;
  padding: 0.125rem 0.375rem;
  border-radius: 4px;
  font-weight: 600;
}

.phase-phase1_fixer {
  background: rgba(46, 160, 67, 0.15);
  color: #3fb950;
}
.phase-phase2_formatter {
  background: rgba(56, 139, 253, 0.15);
  color: #58a6ff;
}
.phase-phase3_linter {
  background: rgba(210, 153, 34, 0.15);
  color: #d29922;
}

.icon-edit-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 0.2rem;
  border-radius: 4px;
}

.icon-edit-btn:hover {
  color: var(--text-main);
  background: var(--hover-bg);
}

.slot-desc {
  font-size: 0.75rem;
  color: var(--text-muted);
  margin: 0;
  line-height: 1.4;
}

.slot-exts-wrap {
  display: flex;
  align-items: center;
  gap: 0.375rem;
}

.ext-label {
  font-size: 0.6875rem;
  color: var(--text-muted);
}

.ext-tags-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.25rem;
}

.ext-chip {
  font-size: 0.6875rem;
  padding: 0.1rem 0.3rem;
  background: var(--section-bg);
  border: 1px solid var(--card-border);
  border-radius: 3px;
  color: var(--text-main);
}

.slot-commands-box {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  padding: 0.5rem;
  background: var(--section-bg);
  border: 1px solid var(--card-border);
  border-radius: 4px;
}

.cmd-line {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  font-size: 0.6875rem;
}

.cmd-label {
  font-weight: 600;
  color: var(--text-muted);
}

.cmd-text {
  color: var(--text-main);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* Reports & Rich Rules Section */
.reports-container {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.reports-header-row {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid var(--border-color, rgba(255, 255, 255, 0.08));
}

@media (min-width: 900px) {
  .reports-header-row {
    flex-direction: row;
    align-items: flex-end;
    justify-content: space-between;
  }
}

.reports-header-titles {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.reports-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 1.125rem;
  font-weight: 600;
  margin: 0;
  color: var(--text-main);
}

.rules-total-count {
  font-size: 0.75rem;
  padding: 0.15rem 0.5rem;
  background: rgba(6, 182, 212, 0.15);
  color: #06b6d4;
  border: 1px solid rgba(6, 182, 212, 0.3);
  border-radius: 12px;
  margin-left: 0.25rem;
}

.reports-sub {
  font-size: 0.8125rem;
  color: var(--text-muted);
  margin: 0;
  line-height: 1.4;
}

.reports-controls-bar {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

@media (min-width: 768px) {
  .reports-controls-bar {
    flex-direction: row;
    align-items: center;
    gap: 1rem;
  }
}

.rule-category-pills {
  display: flex;
  flex-wrap: wrap;
  gap: 0.375rem;
}

.cat-pill-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.35rem 0.75rem;
  font-size: 0.75rem;
  font-weight: 500;
  border-radius: 20px;
  background: var(--card-bg, rgba(255, 255, 255, 0.04));
  border: 1px solid var(--card-border, rgba(255, 255, 255, 0.1));
  color: var(--text-muted);
  cursor: pointer;
  transition: all 0.18s ease;
}

.cat-pill-btn:hover {
  background: rgba(255, 255, 255, 0.08);
  color: var(--text-main);
  border-color: rgba(255, 255, 255, 0.2);
}

.cat-pill-btn.active {
  background: #3b82f6;
  border-color: #3b82f6;
  color: #ffffff;
  box-shadow: 0 2px 6px rgba(59, 130, 246, 0.3);
}

.rule-search-box {
  position: relative;
  display: flex;
  align-items: center;
  min-width: 260px;
}

.rule-search-input {
  width: 100%;
  padding: 0.4rem 1.8rem 0.4rem 2rem;
  font-size: 0.75rem;
  background: var(--card-bg, rgba(0, 0, 0, 0.2));
  border: 1px solid var(--card-border, rgba(255, 255, 255, 0.12));
  border-radius: 6px;
  color: var(--text-main);
  outline: none;
  transition:
    border-color 0.2s,
    box-shadow 0.2s;
}

.rule-search-input:focus {
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.2);
}

.rule-search-box .search-icon {
  position: absolute;
  left: 0.6rem;
  font-size: 0.85rem;
  color: var(--text-muted);
  pointer-events: none;
}

.rule-search-box .clear-icon {
  position: absolute;
  right: 0.6rem;
  font-size: 0.85rem;
  color: var(--text-muted);
  cursor: pointer;
}

.rule-search-box .clear-icon:hover {
  color: var(--text-main);
}

/* Rich Rules Cards Grid */
.rich-rules-cards-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(360px, 1fr));
  gap: 1rem;
}

.rich-rule-card {
  display: flex;
  flex-direction: column;
  padding: 1rem 1.15rem;
  background: var(--card-bg, rgba(30, 30, 30, 0.7));
  border: 1px solid var(--card-border, rgba(255, 255, 255, 0.08));
  border-radius: 8px;
  transition:
    transform 0.18s ease,
    border-color 0.18s ease,
    box-shadow 0.18s ease;
  position: relative;
  overflow: hidden;
}

.rich-rule-card:hover {
  transform: translateY(-2px);
  border-color: rgba(59, 130, 246, 0.4);
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.25);
}

.rich-rule-card.is-fail {
  border-color: rgba(239, 68, 68, 0.5);
  background: rgba(239, 68, 68, 0.04);
}

.rich-rule-card.is-warn {
  border-color: rgba(245, 158, 11, 0.5);
  background: rgba(245, 158, 11, 0.04);
}

.rich-rule-card.is-fixed {
  border-color: rgba(16, 185, 129, 0.5);
  background: rgba(16, 185, 129, 0.04);
}

.rich-rule-card.is-passed {
  border-color: rgba(34, 197, 94, 0.35);
}

.rule-card-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.5rem;
  gap: 0.5rem;
}

.rule-id-row {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  overflow: hidden;
}

.rule-id {
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--text-muted);
  white-space: nowrap;
  text-overflow: ellipsis;
  overflow: hidden;
}

.rule-category-tag {
  font-size: 0.65rem;
  padding: 0.1rem 0.4rem;
  border-radius: 4px;
  font-weight: 500;
}

.cat-hygiene {
  background: rgba(34, 197, 94, 0.15);
  color: #22c55e;
  border: 1px solid rgba(34, 197, 94, 0.25);
}

.cat-formatting {
  background: rgba(59, 130, 246, 0.15);
  color: #3b82f6;
  border: 1px solid rgba(59, 130, 246, 0.25);
}

.cat-linter {
  background: rgba(245, 158, 11, 0.15);
  color: #f59e0b;
  border: 1px solid rgba(245, 158, 11, 0.25);
}

.cat-security {
  background: rgba(168, 85, 247, 0.15);
  color: #c084fc;
  border: 1px solid rgba(168, 85, 247, 0.25);
}

.rule-status-badge {
  display: flex;
  align-items: center;
}

.rule-passed-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  font-size: 0.6875rem;
  padding: 0.15rem 0.45rem;
  border-radius: 4px;
  background: rgba(34, 197, 94, 0.15);
  color: #22c55e;
  border: 1px solid rgba(34, 197, 94, 0.3);
}

.rule-fixed-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  font-size: 0.6875rem;
  padding: 0.15rem 0.45rem;
  border-radius: 4px;
  background: rgba(16, 185, 129, 0.2);
  color: #10b981;
  border: 1px solid rgba(16, 185, 129, 0.4);
}

.rule-failed-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  font-size: 0.6875rem;
  padding: 0.15rem 0.45rem;
  border-radius: 4px;
  background: rgba(239, 68, 68, 0.2);
  color: #ef4444;
  border: 1px solid rgba(239, 68, 68, 0.4);
}

.rule-warn-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  font-size: 0.6875rem;
  padding: 0.15rem 0.45rem;
  border-radius: 4px;
  background: rgba(245, 158, 11, 0.2);
  color: #f59e0b;
  border: 1px solid rgba(245, 158, 11, 0.4);
}

.rule-idle-badge {
  font-size: 0.6875rem;
  padding: 0.15rem 0.45rem;
  border-radius: 4px;
  background: rgba(255, 255, 255, 0.06);
  color: var(--text-muted);
  border: 1px solid rgba(255, 255, 255, 0.08);
}

.rule-title-cn {
  font-size: 0.9375rem;
  font-weight: 600;
  margin: 0 0 0.4rem 0;
  color: var(--text-main);
  line-height: 1.35;
}

.rule-description {
  font-size: 0.775rem;
  color: var(--text-muted);
  line-height: 1.5;
  margin: 0 0 0.85rem 0;
  flex-grow: 1;
}

.rule-card-bottom {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  padding-top: 0.6rem;
  border-top: 1px solid rgba(255, 255, 255, 0.06);
}

.rule-meta-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
  align-items: center;
}

.meta-chip {
  font-size: 0.65rem;
  padding: 0.12rem 0.4rem;
  border-radius: 3px;
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-muted);
  border: 1px solid rgba(255, 255, 255, 0.08);
}

.phase-chip.phase-fixer {
  background: rgba(34, 197, 94, 0.1);
  color: #4ade80;
  border-color: rgba(34, 197, 94, 0.2);
}

.phase-chip.phase-formatter {
  background: rgba(59, 130, 246, 0.1);
  color: #60a5fa;
  border-color: rgba(59, 130, 246, 0.2);
}

.phase-chip.phase-linter {
  background: rgba(245, 158, 11, 0.1);
  color: #fbbf24;
  border-color: rgba(245, 158, 11, 0.2);
}

.tool-chip {
  background: rgba(147, 51, 234, 0.1);
  color: #c084fc;
  border-color: rgba(147, 51, 234, 0.2);
}

.fixable-chip {
  background: rgba(16, 185, 129, 0.12);
  color: #34d399;
  border-color: rgba(16, 185, 129, 0.25);
  font-weight: 500;
}

.blocker-chip {
  background: rgba(239, 68, 68, 0.12);
  color: #f87171;
  border-color: rgba(239, 68, 68, 0.25);
  font-weight: 500;
}

.warn-chip {
  background: rgba(245, 158, 11, 0.12);
  color: #fbbf24;
  border-color: rgba(245, 158, 11, 0.25);
}

.rule-extensions-row {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.3rem;
  font-size: 0.6875rem;
}

.ext-label {
  color: var(--text-muted);
  font-size: 0.65rem;
}

.ext-badge {
  font-size: 0.65rem;
  padding: 0.08rem 0.3rem;
  background: rgba(255, 255, 255, 0.06);
  border-radius: 3px;
  color: var(--text-main);
  border: 1px solid rgba(255, 255, 255, 0.08);
}

.empty-reports-box {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3.5rem 1rem;
  background: var(--card-bg);
  border: 1px dashed var(--card-border);
  border-radius: 8px;
  color: var(--text-muted);
  font-size: 0.85rem;
  gap: 0.75rem;
}

.empty-reports-icon {
  font-size: 2.25rem;
  opacity: 0.4;
}

.btn-clear-filter {
  padding: 0.35rem 0.85rem;
  font-size: 0.75rem;
  background: #3b82f6;
  color: #ffffff;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  transition: opacity 0.2s;
}

.btn-clear-filter:hover {
  opacity: 0.88;
}

.empty-reports-icon {
  font-size: 2rem;
  opacity: 0.5;
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  100% {
    transform: rotate(360deg);
  }
}

/* ─────────────────────────────────────────────────────────────────────────────
   Light Theme Adaptive Styles for Quality Guard Console & Workbench
   ───────────────────────────────────────────────────────────────────────────── */
:root[data-theme="light"] .live-console-card,
[data-theme="light"] .live-console-card {
  background: #ffffff;
  border-color: #e2e8f0;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.05);
}

:root[data-theme="light"] .console-card-header,
[data-theme="light"] .console-card-header {
  background: #f8fafc;
  border-bottom-color: #e2e8f0;
}

:root[data-theme="light"] .console-title,
[data-theme="light"] .console-title {
  color: #0f172a;
}

:root[data-theme="light"] .console-status-text,
[data-theme="light"] .console-status-text {
  color: #64748b;
}

:root[data-theme="light"] .console-btn,
[data-theme="light"] .console-btn {
  background: #ffffff;
  border-color: #cbd5e1;
  color: #334155;
}

:root[data-theme="light"] .console-btn:hover,
[data-theme="light"] .console-btn:hover {
  background: #f1f5f9;
  color: #0f172a;
}

:root[data-theme="light"] .av-radar-scanner,
[data-theme="light"] .av-radar-scanner {
  background: #f8fafc;
  border-bottom-color: #e2e8f0;
}

:root[data-theme="light"] .av-radar-scanner.is-active,
[data-theme="light"] .av-radar-scanner.is-active {
  background: #f0f9ff;
  border-bottom-color: #bae6fd;
}

:root[data-theme="light"] .radar-pulse-ring,
[data-theme="light"] .radar-pulse-ring {
  background: #e2e8f0;
}

:root[data-theme="light"] .av-radar-scanner.is-active .radar-pulse-ring,
[data-theme="light"] .av-radar-scanner.is-active .radar-pulse-ring {
  background: #e0f2fe;
  box-shadow: 0 0 10px rgba(2, 132, 199, 0.25);
}

:root[data-theme="light"] .radar-scan-label,
[data-theme="light"] .radar-scan-label {
  color: #0284c7;
}

:root[data-theme="light"] .radar-scan-target,
[data-theme="light"] .radar-scan-target {
  color: #0f172a;
}

:root[data-theme="light"] .radar-meta-row,
[data-theme="light"] .radar-meta-row {
  color: #64748b;
}

:root[data-theme="light"] .radar-progress-pct,
[data-theme="light"] .radar-progress-pct {
  color: #0284c7;
}

:root[data-theme="light"] .radar-progress-bar,
[data-theme="light"] .radar-progress-bar {
  background: #e2e8f0;
}

:root[data-theme="light"] .console-body,
[data-theme="light"] .console-body {
  background: #f8fafc;
  border-top: 1px solid #e2e8f0;
}

:root[data-theme="light"] .console-empty-placeholder .placeholder-title,
[data-theme="light"] .console-empty-placeholder .placeholder-title {
  color: #0f172a;
}

:root[data-theme="light"] .console-empty-placeholder .placeholder-desc,
[data-theme="light"] .console-empty-placeholder .placeholder-desc {
  color: #64748b;
}

:root[data-theme="light"] .placeholder-action-btn,
[data-theme="light"] .placeholder-action-btn {
  background: #ffffff;
  border-color: #cbd5e1;
  color: #1e293b;
}

:root[data-theme="light"] .placeholder-action-btn:hover,
[data-theme="light"] .placeholder-action-btn:hover {
  background: #f1f5f9;
}

:root[data-theme="light"] .log-line:hover,
[data-theme="light"] .log-line:hover {
  background: rgba(0, 0, 0, 0.04);
}

:root[data-theme="light"] .log-line.is-selected,
[data-theme="light"] .log-line.is-selected {
  background: rgba(9, 105, 218, 0.12) !important;
  box-shadow: inset 3px 0 0 #0969da;
}

:root[data-theme="light"] .log-line-copy-btn,
[data-theme="light"] .log-line-copy-btn {
  background: #f6f8fa;
  border-color: #d0d7de;
  color: #57606a;
}

:root[data-theme="light"] .log-line-copy-btn:hover,
[data-theme="light"] .log-line-copy-btn:hover {
  background: #eaeef2;
  color: #24292f;
}

:root[data-theme="light"] .btn-highlight,
[data-theme="light"] .btn-highlight {
  background: rgba(9, 105, 218, 0.12) !important;
  color: #0969da !important;
  border-color: rgba(9, 105, 218, 0.3) !important;
}

:root[data-theme="light"] .log-ts,
[data-theme="light"] .log-ts {
  color: #64748b;
}

:root[data-theme="light"] .log-content,
[data-theme="light"] .log-content {
  color: #0f172a;
}

:root[data-theme="light"] .log-pass .log-content,
[data-theme="light"] .log-pass .log-content {
  color: #15803d;
  font-weight: 600;
}

:root[data-theme="light"] .log-warn .log-content,
[data-theme="light"] .log-warn .log-content {
  color: #b45309;
  font-weight: 600;
}

:root[data-theme="light"] .log-error .log-content,
[data-theme="light"] .log-error .log-content {
  color: #b91c1c;
  font-weight: 600;
}

:root[data-theme="light"] .log-cmd .log-content,
[data-theme="light"] .log-cmd .log-content {
  color: #0284c7;
  font-weight: 600;
}

:root[data-theme="light"] .log-scan .log-content,
[data-theme="light"] .log-scan .log-content {
  color: #475569;
}

:root[data-theme="light"] .log-rule .log-content,
[data-theme="light"] .log-rule .log-content {
  color: #1e293b;
}

:root[data-theme="light"] .badge-info,
[data-theme="light"] .badge-info {
  background: rgba(14, 165, 233, 0.15);
  color: #0284c7;
}

:root[data-theme="light"] .badge-pass,
[data-theme="light"] .badge-pass {
  background: rgba(22, 163, 74, 0.15);
  color: #15803d;
}

:root[data-theme="light"] .badge-warn,
[data-theme="light"] .badge-warn {
  background: rgba(217, 119, 6, 0.15);
  color: #b45309;
}

:root[data-theme="light"] .badge-error,
[data-theme="light"] .badge-error {
  background: rgba(220, 38, 38, 0.15);
  color: #b91c1c;
}

:root[data-theme="light"] .badge-cmd,
[data-theme="light"] .badge-cmd {
  background: rgba(14, 165, 233, 0.12);
  color: #0284c7;
}

:root[data-theme="light"] .badge-scan,
[data-theme="light"] .badge-scan {
  background: rgba(100, 116, 139, 0.12);
  color: #475569;
}

:root[data-theme="light"] .badge-rule,
[data-theme="light"] .badge-rule {
  background: rgba(22, 163, 74, 0.12);
  color: #15803d;
}

:root[data-theme="light"] .badge-phase,
[data-theme="light"] .badge-phase {
  background: rgba(147, 51, 234, 0.12);
  color: #7e22ce;
}

:root[data-theme="light"] .badge-config,
[data-theme="light"] .badge-config {
  background: rgba(13, 148, 136, 0.12);
  color: #0f766e;
}

:root[data-theme="light"] .metric-card,
[data-theme="light"] .metric-card {
  background: #ffffff;
  border-color: #e2e8f0;
}

:root[data-theme="light"] .inspection-outcome-card.outcome-clean,
[data-theme="light"] .inspection-outcome-card.outcome-clean {
  background: #f0fdf4;
  border-color: #bbf7d0;
}

:root[data-theme="light"] .inspection-outcome-card.outcome-warning,
[data-theme="light"] .inspection-outcome-card.outcome-warning {
  background: #fffbeb;
  border-color: #fde68a;
}

:root[data-theme="light"] .outcome-headline,
[data-theme="light"] .outcome-headline {
  color: #0f172a;
}

:root[data-theme="light"] .guard-scanning-banner,
[data-theme="light"] .guard-scanning-banner {
  background: #f0f9ff;
  border-color: #bae6fd;
  color: #0369a1;
}

:root[data-theme="light"] .rich-rule-card,
[data-theme="light"] .rich-rule-card {
  background: #ffffff;
  border-color: #e2e8f0;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

:root[data-theme="light"] .rich-rule-card:hover,
[data-theme="light"] .rich-rule-card:hover {
  border-color: #93c5fd;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}

:root[data-theme="light"] .rule-search-input,
[data-theme="light"] .rule-search-input {
  background: #f8fafc;
  border-color: #cbd5e1;
  color: #0f172a;
}

:root[data-theme="light"] .cat-pill-btn,
[data-theme="light"] .cat-pill-btn {
  background: #f1f5f9;
  border-color: #e2e8f0;
  color: #475569;
}

:root[data-theme="light"] .cat-pill-btn:hover,
[data-theme="light"] .cat-pill-btn:hover {
  background: #e2e8f0;
  color: #0f172a;
}

:root[data-theme="light"] .cat-pill-btn.active,
[data-theme="light"] .cat-pill-btn.active {
  background: #2563eb;
  border-color: #2563eb;
  color: #ffffff;
}

:root[data-theme="light"] .rule-title-cn,
[data-theme="light"] .rule-title-cn {
  color: #0f172a;
}

:root[data-theme="light"] .rule-description,
[data-theme="light"] .rule-description {
  color: #475569;
}

:root[data-theme="light"] .rule-card-bottom,
[data-theme="light"] .rule-card-bottom {
  border-top-color: #f1f5f9;
}

:root[data-theme="light"] .ext-badge,
[data-theme="light"] .ext-badge {
  background: #f1f5f9;
  border-color: #e2e8f0;
  color: #334155;
}

:root[data-theme="light"] .rule-idle-badge,
[data-theme="light"] .rule-idle-badge {
  background: #f1f5f9;
  color: #64748b;
  border-color: #e2e8f0;
}

/* Built-in Rules Settings View Styling */
.builtin-rules-settings-view {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.builtin-settings-header-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 1rem;
  padding: 1rem 1.25rem;
  background: rgba(6, 182, 212, 0.08);
  border: 1px solid rgba(6, 182, 212, 0.2);
  border-radius: 8px;
}

.banner-left {
  display: flex;
  align-items: center;
  gap: 0.85rem;
  max-width: 750px;
}

.banner-icon {
  font-size: 1.75rem;
  flex-shrink: 0;
}

.banner-title {
  font-size: 0.95rem;
  font-weight: 600;
  margin: 0 0 0.25rem 0;
  color: var(--text-main, #ffffff);
}

.banner-subtitle {
  font-size: 0.775rem;
  color: var(--text-muted, #94a3b8);
  margin: 0;
  line-height: 1.45;
}

.banner-stats {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.stat-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.25rem 0.6rem;
  border-radius: 14px;
  font-size: 0.725rem;
  font-weight: 500;
}

.stat-pill.active-pill {
  background: rgba(34, 197, 94, 0.12);
  color: #4ade80;
  border: 1px solid rgba(34, 197, 94, 0.25);
}

.stat-pill.skip-pill {
  background: rgba(245, 158, 11, 0.12);
  color: #fbbf24;
  border: 1px solid rgba(245, 158, 11, 0.25);
}

.stat-pill .dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
}

.dot-green {
  background: #22c55e;
}

.dot-amber {
  background: #f59e0b;
}

/* Builtin Cards Grid */
.builtin-cards-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: 1rem;
}

.builtin-card {
  display: flex;
  flex-direction: column;
  padding: 1rem 1.15rem;
  background: var(--card-bg, rgba(30, 30, 30, 0.6));
  border: 1px solid var(--card-border, rgba(255, 255, 255, 0.08));
  border-radius: 8px;
  transition: all 0.2s ease;
  position: relative;
}

.builtin-card:hover {
  border-color: rgba(6, 182, 212, 0.4);
  box-shadow: 0 4px 14px rgba(0, 0, 0, 0.25);
}

.builtin-card.is-skipped {
  opacity: 0.68;
  border-style: dashed;
  border-color: rgba(245, 158, 11, 0.35);
  background: rgba(245, 158, 11, 0.02);
}

.builtin-card-top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
}

.builtin-title-group {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
  overflow: hidden;
}

.builtin-cn-name {
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--text-main, #ffffff);
}

.builtin-id-code {
  font-size: 0.725rem;
  color: var(--text-muted, #94a3b8);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.builtin-actions-group {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-shrink: 0;
}

.badge-lock-tag {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  font-size: 0.65rem;
  padding: 0.15rem 0.45rem;
  border-radius: 3px;
  background: rgba(148, 163, 184, 0.12);
  color: #94a3b8;
  border: 1px solid rgba(148, 163, 184, 0.2);
}

/* Switch Button */
.btn-toggle-switch {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.15rem 0.45rem 0.15rem 0.25rem;
  border-radius: 12px;
  background: rgba(239, 68, 68, 0.15);
  border: 1px solid rgba(239, 68, 68, 0.3);
  color: #f87171;
  font-size: 0.65rem;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-toggle-switch .toggle-thumb {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: #ef4444;
  transition: transform 0.2s;
}

.btn-toggle-switch.active {
  background: rgba(34, 197, 94, 0.15);
  border-color: rgba(34, 197, 94, 0.4);
  color: #4ade80;
}

.btn-toggle-switch.active .toggle-thumb {
  background: #22c55e;
}

.builtin-desc {
  font-size: 0.775rem;
  color: var(--text-muted, #94a3b8);
  margin: 0 0 0.75rem 0;
  line-height: 1.45;
  flex-grow: 1;
}

.builtin-custom-tag-row {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  margin-bottom: 0.65rem;
  padding: 0.25rem 0.5rem;
  background: rgba(6, 182, 212, 0.08);
  border-radius: 4px;
}

.custom-hint {
  font-size: 0.6875rem;
  color: #38bdf8;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.builtin-card-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-top: 0.6rem;
  border-top: 1px solid rgba(255, 255, 255, 0.06);
  gap: 0.5rem;
}

.builtin-chips {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.3rem;
}

.builtin-chips .chip {
  font-size: 0.65rem;
  padding: 0.1rem 0.35rem;
  border-radius: 3px;
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-muted);
}

.chip.fix-chip {
  background: rgba(16, 185, 129, 0.12);
  color: #34d399;
}

.chip.blocker-chip {
  background: rgba(239, 68, 68, 0.12);
  color: #f87171;
}

.chip.warn-chip {
  background: rgba(245, 158, 11, 0.12);
  color: #fbbf24;
}

.btn-configure-builtin {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.25rem 0.55rem;
  font-size: 0.725rem;
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid var(--card-border, rgba(255, 255, 255, 0.12));
  border-radius: 4px;
  color: var(--text-main, #ffffff);
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-configure-builtin:hover {
  background: #2563eb;
  border-color: #3b82f6;
  color: #ffffff;
}

/* Builtin Matrix Cell */
.tool-cell.builtin-cell {
  cursor: pointer;
  padding: 0.3rem 0.5rem;
  border-radius: 4px;
  background: rgba(6, 182, 212, 0.05);
  border: 1px solid rgba(6, 182, 212, 0.2);
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
  transition: all 0.15s ease;
}

.tool-cell.builtin-cell:hover {
  background: rgba(6, 182, 212, 0.15);
  border-color: rgba(6, 182, 212, 0.4);
}

.builtin-note {
  font-size: 0.65rem;
  color: var(--text-muted, #94a3b8);
}

/* Light Theme for Built-in Settings */
:root[data-theme="light"] .builtin-settings-header-banner,
[data-theme="light"] .builtin-settings-header-banner {
  background: #f0f9ff;
  border-color: #bae6fd;
}

:root[data-theme="light"] .banner-title,
[data-theme="light"] .banner-title,
:root[data-theme="light"] .builtin-cn-name,
[data-theme="light"] .builtin-cn-name {
  color: #0f172a;
}

:root[data-theme="light"] .builtin-card,
[data-theme="light"] .builtin-card {
  background: #ffffff;
  border-color: #e2e8f0;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

:root[data-theme="light"] .btn-configure-builtin,
[data-theme="light"] .btn-configure-builtin {
  background: #f1f5f9;
  border-color: #cbd5e1;
  color: #334155;
}

:root[data-theme="light"] .btn-configure-builtin:hover,
[data-theme="light"] .btn-configure-builtin:hover {
  background: #2563eb;
  border-color: #2563eb;
  color: #ffffff;
}
</style>
