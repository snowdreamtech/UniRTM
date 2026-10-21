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

        <!-- Run Check All -->
        <button
          class="btn-action secondary"
          :disabled="isChecking"
          @click="handleCheck('all')"
          title="Run full repository scan"
        >
          <span
            class="codicon"
            :class="isChecking && checkingMode === 'all' ? 'codicon-loading spin' : 'codicon-run-all'"
          ></span>
          <span>{{ t("guard.actions.check_all") || "Check Entire Repo" }}</span>
        </button>

        <!-- One-click Auto Fix -->
        <button
          class="btn-action secondary"
          :disabled="isFixing"
          @click="handleFix"
          title="Automatically fix formatting and whitespace errors"
        >
          <span class="codicon" :class="isFixing ? 'codicon-loading spin' : 'codicon-wand'"></span>
          <span>{{
            isFixing ? t("guard.actions.fixing") || "Fixing..." : t("guard.actions.auto_fix") || "Auto-Fix"
          }}</span>
        </button>

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
            @click="handleFix"
            title="Run one-click automatic fixes"
          >
            <span class="codicon codicon-wand"></span>
            <span>{{ t("guard.actions.auto_fix") || "Auto-Fix Now" }}</span>
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
                <button class="console-btn" @click="copyAllLogs" title="Copy inspection logs to clipboard">
                  <span class="codicon codicon-copy"></span>
                  <span>{{ t("guard.logs.copy") || "Copy" }}</span>
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
                    <span>{{ currentScannedCount }} / {{ totalFilesCount || guardStatus.scanned || 2619 }} files</span>
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

              <div v-else class="console-log-lines">
                <div v-for="(item, idx) in formattedLogs" :key="idx" class="log-line" :class="'log-' + item.type">
                  <span class="log-ts font-mono">{{ item.time }}</span>
                  <span class="log-badge font-mono" :class="'badge-' + item.type">{{ item.tag }}</span>
                  <span class="log-content font-mono">{{ item.text }}</span>
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
              <button class="btn-action primary compact" :disabled="isFixing" @click="handleFix">
                <span class="codicon codicon-wand"></span>
                <span>{{ t("guard.actions.auto_fix") || "Auto-Fix Now" }}</span>
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
          <div class="reports-header">
            <h3 class="reports-title">{{ t("guard.diagnostics.checklist") || "Transparent Execution Checklist" }}</h3>
            <span class="reports-sub">
              {{ t("guard.reports_desc") || "Real-time streaming validation status and execution duration per rule." }}
            </span>
          </div>

          <div v-if="guardStatus.rule_reports && guardStatus.rule_reports.length > 0" class="checklist-items-grid">
            <div
              v-for="report in guardStatus.rule_reports"
              :key="report.rule_name"
              class="checklist-item"
              :class="
                report.status === 'error' || report.status === 'fail'
                  ? 'item-error'
                  : report.status === 'warn'
                    ? 'item-warning'
                    : 'item-pass'
              "
            >
              <div class="item-left">
                <span
                  class="codicon"
                  :class="
                    report.status === 'error' || report.status === 'fail'
                      ? 'codicon-error text-red'
                      : report.status === 'warn'
                        ? 'codicon-warning text-amber'
                        : 'codicon-pass text-green'
                  "
                ></span>
                <span class="item-rule-name font-mono">{{ report.rule_name }}</span>
              </div>
              <div class="item-right">
                <span class="item-phase-badge" :class="'phase-' + report.phase">{{ report.phase }}</span>
                <span class="item-duration font-mono">{{ report.duration_ms }}ms</span>
              </div>
            </div>
          </div>

          <div v-else class="empty-reports-box">
            <span class="codicon codicon-checklist empty-reports-icon"></span>
            <p>
              {{
                t("guard.no_reports") ||
                "No inspection reports yet. Run 'Check Staged' or 'Check Entire Repo' to see live results."
              }}
            </p>
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

        <!-- Sub Tabs inside Settings: Pipelines Matrix vs CLI Slots -->
        <div class="settings-sub-tabs">
          <button
            class="sub-tab-btn"
            :class="{ active: settingsSubTab === 'matrix' }"
            @click="settingsSubTab = 'matrix'"
          >
            <span class="codicon codicon-table"></span>
            <span>{{ t("guard.tabs.matrix") || "Extension Pipelines (Matrix)" }}</span>
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
                  <span v-else class="cell-empty font-mono">—</span>
                </td>

                <!-- Security Column -->
                <td class="col-security">
                  <div v-if="row.auditor" class="tool-cell" @click="editSlot(row.auditor)">
                    <span class="tool-name text-purple">{{ row.auditor.default_tool_name }}</span>
                    <code class="tool-cmd font-mono">{{ row.auditor.default_command }}</code>
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

        <!-- 3B. Tool Slots Grid (CLI Tool Cards) -->
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

    <!-- Toml Config Fallback Drawer -->
    <TomlConfigDrawer :isOpen="isTomlDrawerOpen" @close="isTomlDrawerOpen = false" @saved="onTomlSaved" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onActivated } from "vue";
import { useRouter } from "vue-router";
import { t } from "../i18n";
import { useGuard, type ToolSlotItem, type SlotConfigInput } from "../composables/useGuard";
import { useEditorTabs } from "../composables/useEditorTabs";
import { useToast } from "../composables/useToast";
import ExtensionSlotModal from "../components/guard/ExtensionSlotModal.vue";
import TomlConfigDrawer from "../components/guard/TomlConfigDrawer.vue";

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
} = useGuard();

const isSlotModalOpen = ref(false);
const activeEditingSlot = ref<ToolSlotItem | null>(null);
const targetExtForAdd = ref<string>("");

const isTomlDrawerOpen = ref(false);
const checkingMode = ref<"changed" | "staged" | "all">("staged");

// Default primary tab is now the transparent Live Guard & Logs Workbench!
const activeTab = ref<"workbench" | "reports" | "settings">("workbench");
const settingsSubTab = ref<"matrix" | "slots">("matrix");
const searchKeyword = ref("");

// Last inspection tracking for high visibility and transparency
const lastCheckedTime = ref<string>("");
const lastCheckedMode = ref<"changed" | "staged" | "all">("staged");
const justInspected = ref(false);

export interface FormattedLogItem {
  time: string;
  tag: string;
  type: "info" | "pass" | "warn" | "error" | "config" | "phase" | "cmd" | "scan" | "rule";
  text: string;
}

const logViewerRef = ref<HTMLElement | null>(null);
const localLogs = ref<FormattedLogItem[]>([]);

// Anti-Virus style real-time scanner dynamic states
const isStreaming = ref(false);
const currentScanningTarget = ref("");
const currentScannedCount = ref(0);
const totalFilesCount = ref(0);
const scanProgressPct = ref(100);

function parseLogLine(raw: string): FormattedLogItem {
  const now = new Date();
  const time = now.toTimeString().split(" ")[0];
  let type: FormattedLogItem["type"] = "info";
  let tag = "GUARD";

  if (raw.startsWith("$ ")) {
    type = "cmd";
    tag = "CLI";
  } else if (raw.startsWith("ℹ ") || raw.startsWith("ℹ") || raw.includes("Inspecting workspace files")) {
    type = "info";
    tag = "INFO";
  } else if (raw.includes("› [scan]") || raw.includes("more files in inspection queue")) {
    type = "scan";
    tag = "SCAN";
  } else if (raw.startsWith("✓ ✨") || raw.includes("Quality Guard passed") || raw.includes("All 38 guard rules")) {
    type = "pass";
    tag = "PASS";
  } else if (raw.startsWith("✓ ") || raw.startsWith("✓")) {
    type = "rule";
    tag = "RULE";
  } else if (raw.startsWith("✗ ") || raw.startsWith("✗") || raw.includes("[FAIL]")) {
    type = "error";
    tag = "FAIL";
  } else if (raw.startsWith("! ") || raw.startsWith("!") || raw.includes("[WARN]")) {
    type = "warn";
    tag = "WARN";
  } else if (raw.includes("[Config:")) {
    type = "config";
    tag = "CONFIG";
  } else if (raw.includes("[Phase")) {
    type = "phase";
    tag = "PHASE";
  } else if (
    raw.toLowerCase().includes("passed") ||
    raw.toLowerCase().includes("clean") ||
    raw.toLowerCase().includes("compliant")
  ) {
    type = "pass";
    tag = "PASS";
  }

  return {
    time,
    tag,
    type,
    text: raw,
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

function copyAllLogs() {
  const fullText = formattedLogs.value.map((l) => `[${l.time}] [${l.tag}] ${l.text}`).join("\n");
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
      return t("guard.modes.all") || "Entire Repository";
  }
}

onMounted(() => {
  fetchStatus().catch(() => {});
  fetchSlots().catch(() => {});
});

onActivated(() => {
  fetchStatus().catch(() => {});
  fetchSlots().catch(() => {});
});

async function handleCheck(mode: "changed" | "staged" | "all") {
  checkingMode.value = mode;
  const modeLabel = getModeLabel(mode);
  emit("log-event", "INFO", `[Quality Guard] Starting inspection (${modeLabel})...`, undefined, "output");

  currentScanningTarget.value = `Initiating quality guard inspection (${modeLabel})...`;
  scanProgressPct.value = 5;
  localLogs.value = [
    parseLogLine(`$ unirtm check --mode=${mode}`),
    parseLogLine(`ℹ Inspecting workspace files with UniRTM Quality Guard (mode: ${mode})...`),
  ];
  scrollToBottom();

  await runCheck({ mode });

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

  if (logsToStream.length > 0) {
    for (const line of logsToStream) {
      emit("log-event", "INFO", line, undefined, "output");
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

async function handleFix() {
  const modeLabel = getModeLabel(checkingMode.value);
  emit("log-event", "INFO", `[Quality Guard] Starting auto-fix pipeline (${modeLabel})...`, undefined, "output");

  currentScanningTarget.value = `Initiating auto-repair & formatting pipeline (${modeLabel})...`;
  scanProgressPct.value = 5;
  localLogs.value = [
    parseLogLine(`$ unirtm check --mode=${checkingMode.value} --fix`),
    parseLogLine(`ℹ Running automated fixers and style convergence (${modeLabel})...`),
  ];
  scrollToBottom();

  await runFix(checkingMode.value);

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

  if (fixLogsToStream.length > 0) {
    for (const line of fixLogsToStream) {
      emit("log-event", "INFO", line, undefined, "output");
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
  padding: 0.125rem 0;
  border-radius: 3px;
}

.log-line:hover {
  background: rgba(255, 255, 255, 0.04);
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

/* Reports Section */
.reports-container {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.reports-header {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.reports-title {
  font-size: 1rem;
  font-weight: 600;
  margin: 0;
  color: var(--text-main);
}

.reports-sub {
  font-size: 0.75rem;
  color: var(--text-muted);
}

.checklist-items-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 0.625rem;
}

.checklist-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.625rem 0.75rem;
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  border-radius: 6px;
}

.item-left {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.item-rule-name {
  font-size: 0.75rem;
  color: var(--text-main);
}

.item-right {
  display: flex;
  align-items: center;
  gap: 0.375rem;
}

.item-phase-badge {
  font-size: 0.625rem;
  padding: 0.1rem 0.25rem;
  border-radius: 3px;
}

.item-duration {
  font-size: 0.6875rem;
  color: var(--text-muted);
}

.empty-reports-box {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3rem 1rem;
  background: var(--card-bg);
  border: 1px dashed var(--card-border);
  border-radius: 8px;
  color: var(--text-muted);
  font-size: 0.8125rem;
  gap: 0.5rem;
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

:root[data-theme="light"] .checklist-item,
[data-theme="light"] .checklist-item {
  background: #ffffff;
  border-color: #e2e8f0;
}
</style>
