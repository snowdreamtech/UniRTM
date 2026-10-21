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
  color: var(--text-secondary);
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
  background: var(--card-bg, #1e1e1e);
  border: 1px solid var(--card-border, #2d2d2d);
  border-radius: 6px;
}

.meta-stat {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.stat-value {
  font-size: 1rem;
  font-weight: 700;
  color: var(--text-primary);
}

.stat-label {
  font-size: 0.6875rem;
  color: var(--text-secondary);
  margin-top: 0.125rem;
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
  background: var(--vscode-button-background, #0078d4);
  color: var(--vscode-button-foreground, #ffffff);
}

.sidebar-action-btn.primary:hover:not(:disabled) {
  background: var(--vscode-button-hoverBackground, #026ec1);
}

.sidebar-action-btn.secondary {
  background: var(--vscode-button-secondaryBackground, #3a3d41);
  color: var(--vscode-button-secondaryForeground, #ffffff);
  border: 1px solid var(--card-border, #45494e);
}

.sidebar-action-btn.secondary:hover:not(:disabled) {
  background: var(--vscode-button-secondaryHoverBackground, #45494e);
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
  color: var(--text-secondary);
  background: transparent;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  text-align: left;
  transition: all 0.15s ease;
}

.cat-nav-item:hover {
  background: var(--hover-bg, rgba(255, 255, 255, 0.05));
  color: var(--text-primary);
}

.cat-nav-item.active {
  background: var(--selected-bg, rgba(0, 120, 212, 0.15));
  color: var(--vscode-button-background, #388bfd);
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
