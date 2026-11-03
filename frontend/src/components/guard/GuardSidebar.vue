<template>
  <div class="guard-sidebar" id="guard-sidebar-panel" @contextmenu.prevent>
    <!-- Quick Status Banner -->
    <div class="sidebar-section">
      <div class="section-header">
        <span class="section-title">{{ t("guard.sidebar.status_title") || "GUARD STATUS" }}</span>
        <span
          class="status-indicator"
          :class="guardStatus.clean ? 'status-pass' : 'status-fail'"
          :title="guardStatus.clean ? 'All guard rules passed' : 'Issues detected'"
        >
          <span class="codicon" :class="guardStatus.clean ? 'codicon-pass' : 'codicon-warning'"></span>
          <span>{{
            guardStatus.clean ? t("guard.status.passed") || "PASSED" : t("guard.status.failed") || "ISSUES"
          }}</span>
        </span>
      </div>

      <div class="status-meta-card">
        <div class="meta-stat">
          <span class="stat-value font-mono">{{ guardStatus.rules_count || 38 }}</span>
          <span class="stat-label">{{ t("guard.rules.total") || "Rules" }}</span>
        </div>
        <div class="meta-stat">
          <span
            class="stat-value font-mono"
            :class="(guardStatus.issues?.length || 0) > 0 ? 'text-amber' : 'text-green'"
          >
            {{ guardStatus.issues?.length || 0 }}
          </span>
          <span class="stat-label">{{ t("guard.diagnostics.issues") || "Issues" }}</span>
        </div>
        <div class="meta-stat">
          <span class="stat-value font-mono">{{
            guardStatus.duration_ms ? `${guardStatus.duration_ms}ms` : "--"
          }}</span>
          <span class="stat-label">{{ t("guard.status.duration") || "Duration" }}</span>
        </div>
      </div>
    </div>

    <!-- Quick Action Triggers -->
    <div class="sidebar-section">
      <div class="section-header">
        <span class="section-title">{{ t("guard.sidebar.actions_title") || "ACTIONS" }}</span>
      </div>
      <div class="action-buttons-group">
        <button class="sidebar-action-btn primary" :disabled="isChecking" @click="handleRunCheck('staged')">
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
        <button class="sidebar-action-btn secondary" :disabled="isChecking" @click="handleRunCheck('changed')">
          <span
            class="codicon"
            :class="isChecking && checkingMode === 'changed' ? 'codicon-loading spin' : 'codicon-diff-modified'"
          ></span>
          <span>{{
            isChecking && checkingMode === "changed"
              ? t("guard.actions.checking") || "Checking..."
              : t("guard.actions.check_changed") || "Check Workspace"
          }}</span>
        </button>
        <button class="sidebar-action-btn secondary" :disabled="isChecking" @click="handleRunCheck('all')">
          <span
            class="codicon"
            :class="isChecking && checkingMode === 'all' ? 'codicon-loading spin' : 'codicon-run-all'"
          ></span>
          <span>{{ t("guard.actions.check_all") || "Check Project" }}</span>
        </button>

        <!-- Split Dropdown Auto-Fix Button -->
        <div ref="fixDropdownRef" class="sidebar-split-btn-group">
          <button
            class="sidebar-split-main-btn"
            :disabled="isFixing || isChecking"
            @click="handleRunFix()"
            :title="getFixButtonTooltip()"
          >
            <span class="codicon" :class="isFixing ? 'codicon-loading spin' : 'codicon-wand'"></span>
            <span>{{ getFixButtonLabel() }}</span>
          </button>
          <button
            class="sidebar-split-arrow-btn"
            :disabled="isFixing || isChecking"
            @click.stop="isFixDropdownOpen = !isFixDropdownOpen"
            :aria-expanded="isFixDropdownOpen"
            :title="t('guard.actions.fix_scope_title') || 'Select Auto-Fix Scope'"
          >
            <span class="codicon codicon-chevron-down"></span>
          </button>

          <!-- Dropdown Menu -->
          <transition name="dropdown-fade">
            <div v-if="isFixDropdownOpen" class="sidebar-fix-dropdown-menu">
              <div class="sidebar-fix-dropdown-header">
                {{ t("guard.actions.fix_scope_title") || "Select Auto-Fix Scope" }}
              </div>
              <button
                class="sidebar-fix-dropdown-item"
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
                class="sidebar-fix-dropdown-item"
                :class="{ active: checkingMode === 'changed' }"
                @click="selectFixScope('changed')"
              >
                <div class="item-title-row">
                  <span class="codicon codicon-diff-modified"></span>
                  <span class="item-name">{{ t("guard.actions.fix_changed") || "Fix Workspace" }}</span>
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
                class="sidebar-fix-dropdown-item"
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
      </div>
    </div>

    <!-- Navigation / Categories Filter -->
    <div class="sidebar-section">
      <div class="section-header">
        <span class="section-title">{{ t("guard.sidebar.categories_title") || "CATEGORIES" }}</span>
      </div>
      <div class="category-nav-list">
        <button
          v-for="cat in categories"
          :key="cat.id"
          class="cat-nav-item"
          :class="{ active: selectedCategory === cat.id }"
          @click="selectCategory(cat.id)"
        >
          <span class="codicon" :class="cat.icon"></span>
          <span class="cat-label">{{ cat.label }}</span>
        </button>
      </div>
    </div>

    <!-- Rules List in Sidebar -->
    <div class="sidebar-section">
      <div class="section-header">
        <span class="section-title">{{ t("guard.tabs.reports") || "RULES LIST" }}</span>
        <span class="section-badge font-mono">{{ filteredRules.length }}</span>
      </div>
      <div class="sidebar-rules-list">
        <button
          v-for="rule in filteredRules"
          :key="rule.id"
          class="sidebar-rule-item"
          :class="{ active: selectedRuleId === rule.id }"
          @click="handleSelectRule(rule)"
          :title="rule.description || rule.displayName"
        >
          <span class="rule-status-dot" :class="rule.statusClass"></span>
          <div class="rule-item-content">
            <span class="rule-item-name">{{ rule.displayName }}</span>
            <span class="rule-item-meta">{{ rule.categoryLabel }}</span>
          </div>
          <span v-if="rule.issuesCount > 0" class="rule-item-issues-badge">{{ rule.issuesCount }}</span>
        </button>
        <div v-if="filteredRules.length === 0" class="sidebar-empty-rules">
          <span>{{ t("common.no_results") || "No matching rules found" }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from "vue";
import { t } from "../../i18n";
import { useGuard } from "../../composables/useGuard";
import { useToast } from "../../composables/useToast";
import { ALL_38_GUARD_RULES, getGuardRuleName } from "../../utils/guardRulesMeta";

const props = defineProps<{ searchQuery?: string }>();

const {
  guardStatus,
  isChecking,
  isFixing,
  lastRunResult,
  runCheck,
  runFix,
  setActiveGuardTab,
  selectedGuardCategory,
  guardFilterKeyword,
} = useGuard();
const { showToast } = useToast();

const selectedCategory = ref(selectedGuardCategory.value || "all");
const selectedRuleId = ref<string>("");
const checkingMode = ref<"changed" | "staged" | "all">("staged");

// Fix dropdown state
const isFixDropdownOpen = ref(false);
const fixDropdownRef = ref<HTMLElement | null>(null);

function handleDocClick(e: MouseEvent) {
  if (!isFixDropdownOpen.value) return;
  if (fixDropdownRef.value && !fixDropdownRef.value.contains(e.target as Node)) {
    isFixDropdownOpen.value = false;
  }
}

onMounted(() => {
  document.addEventListener("click", handleDocClick);
});

onUnmounted(() => {
  document.removeEventListener("click", handleDocClick);
});

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
  return t("guard.actions.fix_changed") || "Fix Workspace";
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
  handleRunFix(mode);
}

const categories = computed(() => [
  { id: "all", label: t("guard.cat.all") || "All Domains", icon: "codicon-layers" },
  { id: "hygiene", label: t("guard.cat.hygiene") || "Hygiene & Safety", icon: "codicon-shield" },
  { id: "format", label: t("guard.cat.format") || "Code Formatting", icon: "codicon-symbol-ruler" },
  { id: "lint", label: t("guard.cat.lint") || "Static Analysis", icon: "codicon-search" },
  { id: "secret", label: t("guard.cat.secret") || "Secret Leak Prevention", icon: "codicon-lock" },
]);

function selectCategory(catId: string) {
  selectedCategory.value = catId;
  selectedGuardCategory.value = catId;
  setActiveGuardTab("reports");
}

function handleSelectRule(rule: any) {
  selectedRuleId.value = rule.id;
  selectedGuardCategory.value = rule.category;
  guardFilterKeyword.value = rule.displayName || rule.name;
  setActiveGuardTab("reports");
}

watch(
  () => props.searchQuery,
  (val) => {
    guardFilterKeyword.value = val || "";
  }
);

watch(selectedGuardCategory, (val) => {
  if (val && val !== selectedCategory.value) {
    selectedCategory.value = val;
  }
});

const reportsMap = computed(() => {
  const map = new Map<string, { status: string; issuesCount: number }>();
  const reports = guardStatus.value.rule_reports || [];
  for (const r of reports) {
    const rawStatus = (r.status || "").toLowerCase();
    const issues = r.issues_count || (r.errors_count || 0) + (r.warnings_count || 0);
    const item = { status: rawStatus, issuesCount: issues };
    if (r.rule_id) map.set(r.rule_id, item);
    if (r.rule_name) map.set(r.rule_name, item);
  }
  return map;
});

const filteredRules = computed(() => {
  const query = (props.searchQuery || "").trim().toLowerCase();
  const cat = selectedCategory.value;

  return ALL_38_GUARD_RULES.map((rule) => {
    const rep = reportsMap.value.get(rule.id) || reportsMap.value.get(rule.name);
    let statusClass = "status-idle";
    let issuesCount = 0;
    if (rep) {
      issuesCount = rep.issuesCount;
      if (rep.status === "passed" || rep.status === "pass" || rep.status === "clean") {
        statusClass = "status-passed";
      } else if (rep.status === "fixed") {
        statusClass = "status-fixed";
      } else if (rep.status === "fail" || rep.status === "error") {
        statusClass = "status-fail";
      } else if (rep.status === "warn") {
        statusClass = "status-warn";
      }
    }
    return {
      ...rule,
      displayName: getGuardRuleName(rule),
      statusClass,
      issuesCount,
    };
  }).filter((rule) => {
    if (cat !== "all") {
      const ruleCat = rule.category;
      if (cat === "hygiene" && ruleCat !== "hygiene") return false;
      if (cat === "format" && ruleCat !== "formatting") return false;
      if (cat === "lint" && ruleCat !== "linter") return false;
      if (cat === "secret" && ruleCat !== "security") return false;
    }
    if (query) {
      const matchName = rule.displayName.toLowerCase().includes(query) || rule.name.toLowerCase().includes(query);
      const matchId = rule.id.toLowerCase().includes(query);
      const matchDesc = (rule.description || "").toLowerCase().includes(query);
      if (!matchName && !matchId && !matchDesc) return false;
    }
    return true;
  });
});

async function handleRunCheck(mode: "changed" | "staged" | "all") {
  checkingMode.value = mode;
  await runCheck({ mode });
  const res = lastRunResult.value;
  if (!res || res.clean) {
    showToast(
      t("guard.toast.clean", {
        count: res?.scanned ?? guardStatus.value.scanned,
        duration: res?.durationMs ?? guardStatus.value.duration_ms ?? 15,
      }) || `Inspection passed: ${res?.scanned ?? guardStatus.value.scanned} files clean`,
      "success"
    );
  } else {
    showToast(
      t("guard.toast.issues", {
        count: res.issuesCount,
        duration: res.durationMs,
      }) || `Inspection alert: found ${res.issuesCount} issues across ${res.scanned} files`,
      "warning"
    );
  }
}

async function handleRunFix(overrideMode?: "staged" | "changed" | "all") {
  if (overrideMode) {
    checkingMode.value = overrideMode;
  }
  isFixDropdownOpen.value = false;
  await runFix(checkingMode.value);
  const res = lastRunResult.value;
  showToast(
    t("guard.toast.fixed", {
      count: res?.fixedCount ?? 1,
      duration: res?.durationMs ?? guardStatus.value.duration_ms ?? 25,
    }) || `Auto-fix finished: formatting & hygiene fixes applied`,
    "success"
  );
}
</script>

<style scoped>
.guard-sidebar {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding: 0.75rem 0.5rem;
}

.sidebar-section {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 0.25rem;
}

.section-title {
  font-size: 0.6875rem;
  font-weight: 700;
  letter-spacing: 0.05em;
  color: var(--text-muted);
  text-transform: uppercase;
}

.status-indicator {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  font-size: 0.6875rem;
  font-weight: 600;
  padding: 0.125rem 0.375rem;
  border-radius: 4px;
}

.status-indicator.status-pass {
  background: rgba(46, 160, 67, 0.15);
  color: #3fb950;
  border: 1px solid rgba(46, 160, 67, 0.3);
}

.status-indicator.status-fail {
  background: rgba(248, 81, 73, 0.15);
  color: #f85149;
  border: 1px solid rgba(248, 81, 73, 0.3);
}

.status-meta-card {
  display: flex;
  align-items: center;
  justify-content: space-around;
  padding: 0.625rem;
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  border-radius: 6px;
}

.meta-stat {
  display: flex;
  flex-direction: column;
  align-items: center;
  flex: 1;
  min-width: 0;
  text-align: center;
}

.stat-value {
  font-size: 1rem;
  font-weight: 700;
  color: var(--text-main);
}

.stat-label {
  font-size: 0.6875rem;
  color: var(--text-muted);
  margin-top: 0.125rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  width: 100%;
}

.action-buttons-group {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

.sidebar-action-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.4375rem 0.75rem;
  font-size: 0.75rem;
  font-weight: 500;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.15s ease;
  border: 1px solid transparent;
}

.sidebar-action-btn.primary {
  background: var(--accent-cyan, #0284c7);
  color: #ffffff;
}

.sidebar-action-btn.primary:hover:not(:disabled) {
  opacity: 0.9;
}

.sidebar-action-btn.secondary {
  background: var(--btn-secondary-bg);
  color: var(--btn-secondary-text);
  border: 1px solid var(--btn-secondary-border);
}

.sidebar-action-btn.secondary:hover:not(:disabled) {
  background: var(--btn-secondary-hover-bg);
}

.sidebar-action-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* Sidebar Split Dropdown Button for Auto-Fix */
.sidebar-split-btn-group {
  position: relative;
  display: flex;
  align-items: stretch;
  width: 100%;
}

.sidebar-split-main-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
  padding: 0.4375rem 0.625rem;
  font-size: 0.75rem;
  font-weight: 500;
  border-top-left-radius: 4px;
  border-bottom-left-radius: 4px;
  border-top-right-radius: 0;
  border-bottom-right-radius: 0;
  cursor: pointer;
  background: var(--btn-secondary-bg);
  color: var(--btn-secondary-text);
  border: 1px solid var(--btn-secondary-border);
  border-right: none;
  transition: all 0.15s ease;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.sidebar-split-main-btn:hover:not(:disabled) {
  background: var(--btn-secondary-hover-bg);
}

.sidebar-split-main-btn:disabled,
.sidebar-split-arrow-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.sidebar-split-arrow-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.4375rem 0.5rem;
  font-size: 0.75rem;
  border-top-left-radius: 0;
  border-bottom-left-radius: 0;
  border-top-right-radius: 4px;
  border-bottom-right-radius: 4px;
  cursor: pointer;
  background: var(--btn-secondary-bg);
  color: var(--btn-secondary-text);
  border: 1px solid var(--btn-secondary-border);
  border-left: 1px solid var(--card-border, rgba(255, 255, 255, 0.1));
  transition: all 0.15s ease;
}

.sidebar-split-arrow-btn:hover:not(:disabled) {
  background: var(--btn-secondary-hover-bg);
  color: var(--accent-cyan, #0284c7);
}

.sidebar-fix-dropdown-menu {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  min-width: 270px;
  background: var(--card-bg, #1e1e1e);
  border: 1px solid var(--card-border, #333);
  border-radius: 6px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.35);
  padding: 0.375rem;
  z-index: 1000;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.sidebar-fix-dropdown-header {
  font-size: 0.625rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted, #888);
  padding: 0.25rem 0.5rem;
  border-bottom: 1px solid var(--card-border, rgba(255, 255, 255, 0.08));
  margin-bottom: 0.125rem;
}

.sidebar-fix-dropdown-item {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 0.15rem;
  padding: 0.375rem 0.5rem;
  border-radius: 4px;
  background: transparent;
  border: 1px solid transparent;
  cursor: pointer;
  text-align: left;
  width: 100%;
  transition: all 0.15s ease;
}

.sidebar-fix-dropdown-item:hover {
  background: var(--hover-bg, rgba(255, 255, 255, 0.08));
  border-color: var(--card-border, rgba(255, 255, 255, 0.1));
}

.sidebar-fix-dropdown-item.active {
  background: rgba(2, 132, 199, 0.1);
  border-color: rgba(2, 132, 199, 0.3);
}

.item-title-row {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--text-main, #fff);
  width: 100%;
}

.item-title-row .codicon {
  font-size: 0.8125rem;
  color: var(--accent-cyan, #0284c7);
}

.current-badge {
  margin-left: auto;
  font-size: 0.625rem;
  font-weight: 700;
  padding: 1px 5px;
  border-radius: 8px;
  background: rgba(2, 132, 199, 0.2);
  color: var(--accent-cyan, #0284c7);
  text-transform: uppercase;
}

.item-desc {
  font-size: 0.625rem;
  color: var(--text-muted, #888);
  padding-left: 1.2rem;
  line-height: 1.3;
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

.category-nav-list {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
}

.cat-nav-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  width: 100%;
  padding: 0.375rem 0.5rem;
  font-size: 0.75rem;
  color: var(--text-muted);
  background: transparent;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  text-align: left;
  transition: all 0.15s ease;
}

.cat-nav-item:hover {
  background-color: rgba(128, 128, 128, 0.12);
  color: var(--text-main);
}

:root[data-theme="dark"] .cat-nav-item:hover,
[data-theme="dark"] .cat-nav-item:hover {
  background-color: rgba(255, 255, 255, 0.08);
}

:root[data-theme="light"] .cat-nav-item:hover,
[data-theme="light"] .cat-nav-item:hover {
  background-color: rgba(0, 0, 0, 0.06);
}

.cat-nav-item.active {
  background: rgba(2, 132, 199, 0.12);
  color: var(--accent-cyan);
  font-weight: 600;
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  100% {
    transform: rotate(360deg);
  }
}

/* Sidebar Rules List */
.sidebar-rules-list {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
  max-height: 380px;
  overflow-y: auto;
  overflow-x: hidden;
  padding-right: 2px;
}

.sidebar-rule-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  width: 100%;
  padding: 0.35rem 0.5rem;
  font-size: 0.75rem;
  color: var(--text-muted);
  background: transparent;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  text-align: left;
  transition: all 0.15s ease;
}

.sidebar-rule-item:hover {
  background-color: rgba(128, 128, 128, 0.12);
  color: var(--text-main);
}

.sidebar-rule-item.active {
  background: rgba(2, 132, 199, 0.15);
  color: var(--accent-cyan, #388bfd);
  font-weight: 600;
}

.rule-status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  flex-shrink: 0;
  background-color: #8b949e;
}

.rule-status-dot.status-passed {
  background-color: #3fb950;
  box-shadow: 0 0 4px rgba(63, 185, 80, 0.5);
}

.rule-status-dot.status-fixed {
  background-color: #58a6ff;
}

.rule-status-dot.status-warn {
  background-color: #d29922;
}

.rule-status-dot.status-fail {
  background-color: #f85149;
  box-shadow: 0 0 4px rgba(248, 81, 73, 0.5);
}

.rule-item-content {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-width: 0;
}

.rule-item-name {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  font-size: 0.75rem;
}

.rule-item-meta {
  font-size: 0.625rem;
  color: var(--text-muted);
  opacity: 0.8;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.rule-item-issues-badge {
  font-size: 0.625rem;
  font-weight: 700;
  padding: 1px 5px;
  border-radius: 9999px;
  background: rgba(210, 153, 34, 0.2);
  color: #d29922;
  flex-shrink: 0;
}

.section-badge {
  font-size: 0.6875rem;
  color: var(--text-muted);
  background: var(--card-bg, rgba(255, 255, 255, 0.05));
  padding: 0.1rem 0.4rem;
  border-radius: 9999px;
  border: 1px solid var(--card-border, #333);
}

.sidebar-empty-rules {
  padding: 0.5rem;
  font-size: 0.75rem;
  color: var(--text-muted);
  text-align: center;
  font-style: italic;
}
</style>
