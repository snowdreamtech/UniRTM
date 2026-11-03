<template>
  <div v-if="isOpen" class="modal-backdrop" @click.self="$emit('close')">
    <div class="modal-dialog" role="dialog" aria-modal="true">
      <!-- Modal Header -->
      <div class="modal-header">
        <div class="header-title-wrap">
          <span class="codicon codicon-shield header-icon text-cyan"></span>
          <div>
            <h2 class="modal-title">
              {{ t("guard.builtin_modal.title") || "Configure Built-in Guard Rule" }}
            </h2>
            <div class="rule-id-sub font-mono">{{ rule?.id }}</div>
          </div>
        </div>
        <button class="btn-close" @click="$emit('close')">
          <span class="codicon codicon-close"></span>
        </button>
      </div>

      <!-- Built-in Unremovable Notice Banner -->
      <div class="builtin-notice-box">
        <span class="codicon codicon-info notice-icon"></span>
        <div class="notice-text">
          <strong>{{ t("guard.builtin_modal.notice_bold") || "Core Built-in Rule (Non-deletable)" }}</strong>
          <span>
            {{
              t("guard.builtin_modal.notice_desc") ||
              "Built-in rules run at native speed via the UniRTM binary. You can adjust active status, thresholds, or excluded paths here; all settings auto-sync to .unirtm.toml."
            }}
          </span>
        </div>
      </div>

      <!-- Modal Body -->
      <div v-if="rule" class="modal-body">
        <!-- Rule Basic Overview -->
        <div class="rule-summary-card">
          <div class="summary-top">
            <span class="summary-name">{{ getGuardRuleName(rule) }}</span>
            <span class="summary-phase-badge font-mono">{{ rule.phaseLabel }}</span>
          </div>
          <p class="summary-desc">{{ getGuardRuleDesc(rule) }}</p>
          <div class="summary-meta-row">
            <span class="summary-meta-item">
              <strong>{{ t("guard.modal.category_label") || "Category:" }}</strong> {{ rule.categoryLabel }}
            </span>
            <span class="summary-meta-item">
              <strong>{{ t("guard.modal.healing_label") || "Auto-healing:" }}</strong>
              {{
                rule.fixable
                  ? t("guard.modal.supports_fix") || "⚡ Auto-fix Supported"
                  : t("guard.modal.read_only") || "Read-only Inspection"
              }}
            </span>
            <span class="summary-meta-item">
              <strong>{{ t("guard.modal.default_applies") || "Default Scope:" }}</strong>
              {{ rule.extensions.join(", ") }}
            </span>
          </div>
        </div>

        <!-- Form: Enable / Skip Toggle -->
        <div class="form-group toggle-group-box">
          <label class="form-label font-bold">
            {{ t("guard.builtin_modal.status_label") || "Rule Active Status" }}
          </label>
          <div class="status-toggle-wrapper">
            <button type="button" class="toggle-choice-btn" :class="{ active: isEnabled }" @click="isEnabled = true">
              <span class="codicon codicon-check"></span>
              <span>{{ t("guard.modal.active_opt") || "Active" }}</span>
            </button>
            <button
              type="button"
              class="toggle-choice-btn btn-skip"
              :class="{ active: !isEnabled }"
              @click="isEnabled = false"
            >
              <span class="codicon codicon-debug-step-over"></span>
              <span>{{ t("guard.modal.skip_opt") || "Skip" }}</span>
            </button>
          </div>
          <span class="form-hint">
            {{
              isEnabled
                ? t("guard.modal.active_desc") || "Rule runs normally during pre-commit and check inspections."
                : t("guard.modal.skip_desc") ||
                  "Rule will be added to [guard] skip list in .unirtm.toml and skipped during checks."
            }}
          </span>
        </div>

        <!-- Form: Severity Level -->
        <div class="form-group">
          <label class="form-label font-bold">
            {{ t("guard.builtin_modal.severity_label") || "Severity Level" }}
          </label>
          <div class="select-wrapper">
            <select v-model="formData.severity" class="form-select">
              <option value="default">
                {{
                  t("guard.modal.follow_default", {
                    sev:
                      rule.severity === "blocker"
                        ? t("guard.chip_blocker") || "Blocker"
                        : t("guard.chip_warning") || "Advisory",
                  }) || "Follow Default (" + rule.severity + ")"
                }}
              </option>
              <option value="blocker">
                {{ t("guard.modal.severity_blocker_opt") || "🛑 Blocker (Prevent commit and build on violation)" }}
              </option>
              <option value="warning">
                {{ t("guard.modal.severity_warn_opt") || "⚠️ Warning (Highlight violation without breaking workflow)" }}
              </option>
            </select>
          </div>
        </div>

        <!-- Form: Max File Size (Specifically useful for large-files rule) -->
        <div v-if="rule.id === 'builtin:check-added-large-files'" class="form-group">
          <label class="form-label font-bold">
            {{ t("guard.builtin_modal.max_file_size") || "Max File Size Threshold (MB)" }}
          </label>
          <div class="input-with-unit">
            <input
              v-model.number="maxFileSizeMB"
              type="number"
              step="0.5"
              min="0.1"
              class="form-input"
              placeholder="e.g. 5"
            />
            <span class="unit-badge">MB</span>
          </div>
          <span class="form-hint">
            {{
              t("guard.modal.max_size_hint") ||
              "Intercept large files exceeding threshold from being committed into Git repository."
            }}
          </span>
        </div>

        <!-- Form: Exclude Dirs -->
        <div class="form-group">
          <label class="form-label font-bold">
            {{ t("guard.builtin_modal.exclude_dirs") || "Exclude Directories" }}
          </label>
          <input
            v-model="excludeDirsStr"
            type="text"
            class="form-input font-mono"
            placeholder="e.g. vendor, testdata, dist, tmp"
          />
          <span class="form-hint">
            {{ t("guard.modal.exclude_dirs_hint") || "Comma-separated directories to exclude from this rule." }}
          </span>
        </div>

        <!-- Form: Exclude Files -->
        <div class="form-group">
          <label class="form-label font-bold">
            {{ t("guard.builtin_modal.exclude_files") || "Exclude Specific Files" }}
          </label>
          <input
            v-model="excludeFilesStr"
            type="text"
            class="form-input font-mono"
            placeholder="e.g. *.min.js, config/local.json, test.env"
          />
          <span class="form-hint">
            {{ t("guard.modal.exclude_files_hint") || "Comma-separated filenames or glob patterns to exclude." }}
          </span>
        </div>

        <!-- Form: Extensions Override -->
        <div class="form-group">
          <label class="form-label font-bold">
            {{ t("guard.builtin_modal.extensions_override") || "Custom File Extensions Override" }}
          </label>
          <input
            v-model="extensionsStr"
            type="text"
            class="form-input font-mono"
            :placeholder="rule.extensions.join(', ')"
          />
          <span class="form-hint">{{
            t("guard.modal.ext_hint") ||
            "Leave empty to use defaults. Comma-separated extensions (e.g. .go, .ts, .vue)."
          }}</span>
        </div>
      </div>

      <!-- Modal Footer -->
      <div class="modal-footer">
        <div class="footer-left">
          <span class="badge-unremovable font-mono">
            <span class="codicon codicon-lock"></span> {{ t("guard.cannot_delete") || "Locked" }}
          </span>
          <button type="button" class="btn-reset-default" @click="resetToDefault">
            <span class="codicon codicon-discard"></span>
            <span>{{ t("guard.modal.restore_defaults") || "Restore Defaults" }}</span>
          </button>
        </div>

        <div class="footer-right">
          <button type="button" class="btn-cancel" @click="$emit('close')">
            {{ t("guard.actions.cancel") || "Cancel" }}
          </button>
          <button type="button" class="btn-save" @click="handleSave">
            <span class="codicon codicon-save"></span>
            <span>{{ t("guard.actions.save_config") || "Save to .unirtm.toml" }}</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from "vue";
import { t } from "../../i18n";
import type { DetailedGuardRule } from "../../utils/guardRulesMeta";
import { getGuardRuleName, getGuardRuleDesc } from "../../utils/guardRulesMeta";
import type { BuiltinRuleCustomConfig } from "../../utils/tomlGuardHelper";

export type { BuiltinRuleCustomConfig };

const props = defineProps<{
  isOpen: boolean;
  rule: DetailedGuardRule | null;
  initialConfig?: Partial<BuiltinRuleCustomConfig> | null;
  isSkipped?: boolean;
}>();

const emit = defineEmits<{
  (e: "close"): void;
  (e: "save", config: BuiltinRuleCustomConfig): void;
}>();

const isEnabled = ref(true);
const maxFileSizeMB = ref<number | "">("");
const excludeDirsStr = ref("");
const excludeFilesStr = ref("");
const extensionsStr = ref("");

const formData = ref<{
  severity: "default" | "blocker" | "warning";
}>({
  severity: "default",
});

watch(
  () => [props.isOpen, props.rule, props.initialConfig, props.isSkipped],
  () => {
    if (!props.isOpen || !props.rule) return;

    isEnabled.value = !props.isSkipped;
    const cfg = props.initialConfig;

    if (cfg) {
      if (cfg.enabled !== undefined) {
        isEnabled.value = cfg.enabled;
      }
      formData.value.severity = cfg.severity || "default";
      if (cfg.max_file_size && cfg.max_file_size > 0) {
        maxFileSizeMB.value = Math.round((cfg.max_file_size / (1024 * 1024)) * 10) / 10;
      } else {
        maxFileSizeMB.value = "";
      }
      excludeDirsStr.value = (cfg.exclude_dirs || []).join(", ");
      excludeFilesStr.value = (cfg.exclude_files || []).join(", ");
      extensionsStr.value = (cfg.extensions || []).join(", ");
    } else {
      resetToDefault();
    }
  },
  { immediate: true }
);

function resetToDefault() {
  isEnabled.value = true;
  formData.value.severity = "default";
  maxFileSizeMB.value = "";
  excludeDirsStr.value = "";
  excludeFilesStr.value = "";
  extensionsStr.value = "";
}

function handleSave() {
  if (!props.rule) return;

  const parseList = (str: string) =>
    str
      .split(",")
      .map((s) => s.trim())
      .filter(Boolean);

  let maxSizeBytes: number | undefined = undefined;
  if (typeof maxFileSizeMB.value === "number" && maxFileSizeMB.value > 0) {
    maxSizeBytes = Math.round(maxFileSizeMB.value * 1024 * 1024);
  }

  const result: BuiltinRuleCustomConfig = {
    id: props.rule.id,
    enabled: isEnabled.value,
    severity: formData.value.severity,
    max_file_size: maxSizeBytes,
    exclude_dirs: parseList(excludeDirsStr.value),
    exclude_files: parseList(excludeFilesStr.value),
    extensions: parseList(extensionsStr.value),
  };

  emit("save", result);
}
</script>

<style scoped>
.modal-backdrop {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.65);
  backdrop-filter: blur(4px);
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
}

.modal-dialog {
  background: var(--bg-primary, #1e1e1e);
  border: 1px solid var(--border-color, rgba(255, 255, 255, 0.12));
  border-radius: 10px;
  width: 100%;
  max-width: 620px;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 16px 36px rgba(0, 0, 0, 0.5);
  overflow: hidden;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 1.25rem;
  border-bottom: 1px solid var(--border-color, rgba(255, 255, 255, 0.08));
}

.header-title-wrap {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.header-icon {
  font-size: 1.5rem;
}

.modal-title {
  font-size: 1.1rem;
  font-weight: 600;
  margin: 0;
  color: var(--text-main, #ffffff);
}

.rule-id-sub {
  font-size: 0.75rem;
  color: var(--text-muted, #94a3b8);
  margin-top: 0.1rem;
}

.btn-close {
  background: transparent;
  border: none;
  color: var(--text-muted, #94a3b8);
  cursor: pointer;
  padding: 0.25rem;
  border-radius: 4px;
  transition: all 0.15s;
}

.btn-close:hover {
  color: var(--text-main, #ffffff);
  background: rgba(255, 255, 255, 0.08);
}

/* Built-in Unremovable Notice */
.builtin-notice-box {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  padding: 0.75rem 1.25rem;
  background: rgba(6, 182, 212, 0.08);
  border-bottom: 1px solid rgba(6, 182, 212, 0.2);
  color: #38bdf8;
  font-size: 0.775rem;
  line-height: 1.45;
}

.builtin-notice-box .notice-icon {
  font-size: 1.1rem;
  margin-top: 0.1rem;
  flex-shrink: 0;
}

.modal-body {
  padding: 1.25rem;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 1.15rem;
}

.rule-summary-card {
  padding: 0.85rem 1rem;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 6px;
}

.summary-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.35rem;
}

.summary-name {
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--text-main, #ffffff);
}

.summary-phase-badge {
  font-size: 0.6875rem;
  padding: 0.1rem 0.4rem;
  border-radius: 3px;
  background: rgba(255, 255, 255, 0.08);
  color: var(--text-muted, #94a3b8);
}

.summary-desc {
  font-size: 0.775rem;
  color: var(--text-muted, #94a3b8);
  margin: 0 0 0.5rem 0;
  line-height: 1.45;
}

.summary-meta-row {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  font-size: 0.725rem;
  color: var(--text-muted, #94a3b8);
}

.summary-meta-item strong {
  color: var(--text-main, #ffffff);
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.form-label {
  font-size: 0.8rem;
  color: var(--text-main, #ffffff);
}

.form-hint {
  font-size: 0.725rem;
  color: var(--text-muted, #94a3b8);
  line-height: 1.35;
}

.form-input,
.form-select {
  width: 100%;
  padding: 0.45rem 0.65rem;
  font-size: 0.8rem;
  background: var(--input-bg, rgba(0, 0, 0, 0.3));
  border: 1px solid var(--border-color, rgba(255, 255, 255, 0.15));
  border-radius: 6px;
  color: var(--text-main, #ffffff);
  outline: none;
  transition: border-color 0.2s;
}

.form-input:focus,
.form-select:focus {
  border-color: #3b82f6;
}

.status-toggle-wrapper {
  display: flex;
  gap: 0.65rem;
}

.toggle-choice-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
  padding: 0.5rem 0.85rem;
  border-radius: 6px;
  font-size: 0.8rem;
  font-weight: 500;
  cursor: pointer;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: var(--text-muted, #94a3b8);
  transition: all 0.18s ease;
}

.toggle-choice-btn.active {
  background: rgba(34, 197, 94, 0.15);
  border-color: rgba(34, 197, 94, 0.5);
  color: #4ade80;
}

.toggle-choice-btn.btn-skip.active {
  background: rgba(239, 68, 68, 0.15);
  border-color: rgba(239, 68, 68, 0.5);
  color: #f87171;
}

.input-with-unit {
  position: relative;
  display: flex;
  align-items: center;
}

.input-with-unit .unit-badge {
  position: absolute;
  right: 0.75rem;
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--text-muted, #94a3b8);
}

.modal-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.85rem 1.25rem;
  border-top: 1px solid var(--border-color, rgba(255, 255, 255, 0.08));
  background: rgba(0, 0, 0, 0.15);
}

.footer-left {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.badge-unremovable {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  font-size: 0.7rem;
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
  background: rgba(148, 163, 184, 0.12);
  color: #94a3b8;
  border: 1px solid rgba(148, 163, 184, 0.25);
}

.btn-reset-default {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  font-size: 0.75rem;
  background: transparent;
  border: none;
  color: var(--text-muted, #94a3b8);
  cursor: pointer;
  padding: 0.2rem 0.4rem;
  border-radius: 4px;
  transition: all 0.15s;
}

.btn-reset-default:hover {
  color: #38bdf8;
  background: rgba(56, 189, 248, 0.08);
}

.footer-right {
  display: flex;
  align-items: center;
  gap: 0.65rem;
}

.btn-cancel {
  padding: 0.45rem 0.95rem;
  font-size: 0.8rem;
  background: transparent;
  border: 1px solid var(--border-color, rgba(255, 255, 255, 0.15));
  border-radius: 6px;
  color: var(--text-main, #ffffff);
  cursor: pointer;
  transition: background 0.15s;
}

.btn-cancel:hover {
  background: rgba(255, 255, 255, 0.08);
}

.btn-save {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.45rem 1.05rem;
  font-size: 0.8rem;
  font-weight: 500;
  background: #2563eb;
  border: 1px solid #3b82f6;
  border-radius: 6px;
  color: #ffffff;
  cursor: pointer;
  transition: background 0.15s;
}

.btn-save:hover {
  background: #1d4ed8;
}

/* Light Theme */
:root[data-theme="light"] .modal-dialog,
[data-theme="light"] .modal-dialog {
  background: #ffffff;
  border-color: #e2e8f0;
  box-shadow: 0 16px 36px rgba(0, 0, 0, 0.15);
}

:root[data-theme="light"] .modal-title,
[data-theme="light"] .modal-title,
:root[data-theme="light"] .summary-name,
[data-theme="light"] .summary-name,
:root[data-theme="light"] .form-label,
[data-theme="light"] .form-label {
  color: #0f172a;
}

:root[data-theme="light"] .builtin-notice-box,
[data-theme="light"] .builtin-notice-box {
  background: #f0f9ff;
  border-bottom-color: #bae6fd;
  color: #0369a1;
}

:root[data-theme="light"] .rule-summary-card,
[data-theme="light"] .rule-summary-card {
  background: #f8fafc;
  border-color: #e2e8f0;
}

:root[data-theme="light"] .form-input,
[data-theme="light"] .form-input,
:root[data-theme="light"] .form-select,
[data-theme="light"] .form-select {
  background: #ffffff;
  border-color: #cbd5e1;
  color: #0f172a;
}

:root[data-theme="light"] .toggle-choice-btn,
[data-theme="light"] .toggle-choice-btn {
  background: #f1f5f9;
  border-color: #cbd5e1;
  color: #475569;
}

:root[data-theme="light"] .modal-footer,
[data-theme="light"] .modal-footer {
  background: #f8fafc;
  border-top-color: #e2e8f0;
}
</style>
