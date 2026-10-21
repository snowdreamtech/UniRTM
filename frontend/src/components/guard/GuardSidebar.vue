<template>
  <div class="guard-sidebar" id="guard-sidebar-panel">
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
          <span class="stat-value font-mono" :class="guardStatus.issues.length > 0 ? 'text-amber' : 'text-green'">
            {{ guardStatus.issues.length }}
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
          <span class="codicon" :class="isChecking ? 'codicon-loading spin' : 'codicon-play'"></span>
          <span>{{
            isChecking
              ? t("guard.actions.checking") || "Checking..."
              : t("guard.actions.check_staged") || "Check Staged"
          }}</span>
        </button>
        <button class="sidebar-action-btn secondary" :disabled="isChecking" @click="handleRunCheck('all')">
          <span class="codicon codicon-run-all"></span>
          <span>{{ t("guard.actions.check_all") || "Check Entire Repo" }}</span>
        </button>
        <button class="sidebar-action-btn secondary" :disabled="isFixing" @click="handleRunFix">
          <span class="codicon" :class="isFixing ? 'codicon-loading spin' : 'codicon-wand'"></span>
          <span>{{
            isFixing ? t("guard.actions.fixing") || "Fixing..." : t("guard.actions.auto_fix") || "Auto-Fix All"
          }}</span>
        </button>
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
          @click="selectedCategory = cat.id"
        >
          <span class="codicon" :class="cat.icon"></span>
          <span class="cat-label">{{ cat.label }}</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import { t } from "../../i18n";
import { useGuard } from "../../composables/useGuard";

const { guardStatus, isChecking, isFixing, runCheck, runFix } = useGuard();

const selectedCategory = ref("all");

const categories = computed(() => [
  { id: "all", label: t("guard.cat.all") || "All Domains", icon: "codicon-layers" },
  { id: "hygiene", label: t("guard.cat.hygiene") || "Hygiene & Safety", icon: "codicon-shield" },
  { id: "format", label: t("guard.cat.format") || "Code Formatting", icon: "codicon-symbol-ruler" },
  { id: "lint", label: t("guard.cat.lint") || "Static Analysis", icon: "codicon-search" },
  { id: "secret", label: t("guard.cat.secret") || "Secret Leak Prevention", icon: "codicon-lock" },
]);

function handleRunCheck(mode: "staged" | "all") {
  runCheck({ mode }).catch(() => {});
}

function handleRunFix() {
  runFix().catch(() => {});
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
  background: var(--hover-bg);
  color: var(--text-main);
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
</style>
