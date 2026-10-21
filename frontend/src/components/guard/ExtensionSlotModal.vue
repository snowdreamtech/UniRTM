<template>
  <div v-if="isOpen" class="modal-backdrop" @click.self="$emit('close')">
    <div class="modal-dialog" role="dialog" aria-modal="true">
      <!-- Modal Header -->
      <div class="modal-header">
        <div class="header-title-wrap">
          <span class="codicon codicon-symbol-class header-icon"></span>
          <h2 class="modal-title">
            {{
              isEditing
                ? t("guard.modal.edit_title") || "Edit Tool Slot & Extension Mapping"
                : t("guard.modal.create_title") || "Add Tool Slot & Extension Mapping"
            }}
          </h2>
        </div>
        <button class="btn-close" @click="$emit('close')">
          <span class="codicon codicon-close"></span>
        </button>
      </div>

      <!-- Modal Body -->
      <div class="modal-body">
        <!-- Quick Preset Pick from Installed Tools -->
        <div class="form-group preset-box">
          <label class="form-label">
            <span class="codicon codicon-tools"></span>
            <span>{{ t("guard.modal.pick_tool") || "Pick from UniRTM Managed Tools (Auto-fill)" }}</span>
          </label>
          <div class="select-wrapper">
            <select v-model="selectedPresetTool" class="form-select" @change="applyToolPreset">
              <option value="">{{ t("guard.modal.choose_preset") || "-- Select a recommended tool --" }}</option>
              <option v-for="preset in toolPresets" :key="preset.name" :value="preset.name">
                {{ preset.label }} ({{ preset.category }})
              </option>
            </select>
          </div>
          <span class="form-hint">
            {{
              t("guard.modal.pick_tool_hint") ||
              "Selecting a preset tool automatically populates the category, phase, extensions, and CLI command."
            }}
          </span>
        </div>

        <!-- Slot ID -->
        <div class="form-row">
          <div class="form-group flex-1">
            <label class="form-label required">{{ t("guard.modal.slot_id") || "Slot ID" }}</label>
            <input
              v-model="formData.id"
              type="text"
              class="form-input font-mono"
              placeholder="e.g. web.formatter, go.linter, python.suite"
              :disabled="isEditing"
            />
          </div>

          <div class="form-group flex-1">
            <label class="form-label required">{{ t("guard.modal.tool_name") || "Tool Name" }}</label>
            <input
              v-model="formData.tool"
              type="text"
              class="form-input"
              placeholder="e.g. prettier, eslint, gofmt, ruff"
            />
          </div>
        </div>

        <!-- Category & Phase -->
        <div class="form-row">
          <div class="form-group flex-1">
            <label class="form-label required">{{ t("guard.modal.category") || "Category (Pillar)" }}</label>
            <select v-model="formData.category" class="form-select">
              <option value="format">format (Code Formatting)</option>
              <option value="lint">lint (Static Analysis)</option>
              <option value="secret">secret (Credential & Leak Prevention)</option>
              <option value="security">security (SCA / Dependency Audit)</option>
              <option value="hygiene">hygiene (Workspace Hygiene)</option>
            </select>
          </div>

          <div class="form-group flex-1">
            <label class="form-label required">{{ t("guard.modal.phase") || "Execution Phase" }}</label>
            <select v-model="formData.phase" class="form-select">
              <option value="phase1_fixer">Phase 1: Fixer (In-place healing)</option>
              <option value="phase2_formatter">Phase 2: Formatter (Visual alignment)</option>
              <option value="phase3_linter">Phase 3: Linter (Gate decision)</option>
            </select>
          </div>
        </div>

        <!-- File Extensions Tag Input -->
        <div class="form-group">
          <label class="form-label">{{ t("guard.modal.extensions") || "Mapped File Extensions" }}</label>
          <div class="tags-container">
            <span v-for="ext in formData.extensions" :key="ext" class="ext-chip font-mono">
              {{ ext }}
              <span class="remove-tag" @click="removeExtension(ext)">×</span>
            </span>
            <input
              v-model="newExtInput"
              type="text"
              class="tag-inline-input font-mono"
              placeholder="+ add extension (.ts, .go)"
              @keydown.enter.prevent="addExtension"
            />
          </div>

          <!-- Quick Extension Presets -->
          <div class="quick-tags-row">
            <span class="quick-label">{{ t("guard.modal.quick_add") || "Quick Add:" }}</span>
            <button
              v-for="presetExt in commonExtensions"
              :key="presetExt"
              type="button"
              class="quick-tag-btn font-mono"
              @click="addPresetExtension(presetExt)"
            >
              {{ presetExt }}
            </button>
          </div>
        </div>

        <!-- Inspection Command -->
        <div class="form-group">
          <label class="form-label required">{{ t("guard.modal.check_command") || "Inspection Command" }}</label>
          <div class="input-with-hint">
            <input
              v-model="formData.command"
              type="text"
              class="form-input font-mono"
              placeholder="e.g. prettier --check {{files}} or gofmt -l {{files}}"
            />
            <span class="form-hint">
              {{ commandHintText }}
            </span>
          </div>
        </div>

        <!-- Fix / Self-healing Command -->
        <div class="form-group">
          <label class="form-label">{{
            t("guard.modal.fix_command") || "Auto-Healing / Fix Command (Optional)"
          }}</label>
          <div class="input-with-hint">
            <input
              v-model="formData.fix_command"
              type="text"
              class="form-input font-mono"
              placeholder="e.g. prettier --write {{files}} or gofmt -w {{files}}"
            />
            <span class="form-hint text-green">
              {{
                t("guard.modal.fix_command_hint") ||
                "Executed during 'unirtm fix' or Phase 1 to automatically heal styling issues."
              }}
            </span>
          </div>
        </div>

        <!-- Enabled Toggle Switch -->
        <div class="form-group toggle-row">
          <div class="toggle-info">
            <span class="toggle-title">{{ t("guard.modal.enable_slot") || "Enable this Slot" }}</span>
            <span class="toggle-sub">{{
              t("guard.modal.enable_desc") || "When disabled, Quality Guard bypasses this slot during checks."
            }}</span>
          </div>
          <label class="switch">
            <input type="checkbox" v-model="formData.enabled" />
            <span class="slider round"></span>
          </label>
        </div>
      </div>

      <!-- Modal Footer -->
      <div class="modal-footer">
        <div class="footer-left">
          <button v-if="isEditing" type="button" class="btn-modal-danger" @click="handleDelete">
            <span class="codicon codicon-trash"></span>
            <span>{{ t("guard.actions.delete") || "Delete / Reset" }}</span>
          </button>
        </div>

        <div class="footer-right">
          <button type="button" class="btn-modal-cancel" @click="$emit('close')">
            {{ t("common.cancel") || "Cancel" }}
          </button>
          <button type="button" class="btn-modal-primary" :disabled="!isValid || isSaving" @click="handleSave">
            <span class="codicon" :class="isSaving ? 'codicon-loading spin' : 'codicon-save'"></span>
            <span>{{
              isSaving ? t("guard.actions.saving") || "Saving..." : t("guard.actions.save") || "Save to .unirtm.toml"
            }}</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, watch, computed } from "vue";
import { t } from "../../i18n";
import type { ToolSlotItem, SlotConfigInput } from "../../composables/useGuard";

const props = defineProps<{
  isOpen: boolean;
  initialSlot?: ToolSlotItem | null;
  targetExtension?: string;
}>();

const emit = defineEmits<{
  (e: "close"): void;
  (e: "save", slot: SlotConfigInput): void;
  (e: "delete", slotId: string): void;
}>();

const isSaving = ref(false);
const selectedPresetTool = ref("");
const newExtInput = ref("");

const isEditing = computed(() => !!props.initialSlot?.id);

const commandHintText = computed(() => {
  return t("guard.modal.command_hint") || "Use {{files}} placeholder for the list of staged/inspected files.";
});

const formData = reactive<SlotConfigInput>({
  id: "",
  tool: "",
  command: "",
  fix_command: "",
  category: "format",
  phase: "phase2_formatter",
  extensions: [],
  enabled: true,
});

const commonExtensions = [".go", ".ts", ".tsx", ".js", ".vue", ".sh", ".toml", ".py", ".rs", "*"];

const toolPresets = [
  {
    name: "gofmt",
    label: "gofmt (Go Formatter)",
    id: "go.formatter",
    tool: "gofmt",
    category: "format",
    phase: "phase2_formatter",
    command: "gofmt -l {{files}}",
    fix_command: "gofmt -w {{files}}",
    extensions: [".go"],
  },
  {
    name: "goimports",
    label: "goimports (Go Imports & Fixer)",
    id: "go.imports",
    tool: "goimports",
    category: "hygiene",
    phase: "phase1_fixer",
    command: "goimports -l {{files}}",
    fix_command: "goimports -w {{files}}",
    extensions: [".go"],
  },
  {
    name: "golangci-lint",
    label: "golangci-lint (Go Static Linter)",
    id: "go.linter",
    tool: "golangci-lint",
    category: "lint",
    phase: "phase3_linter",
    command: "golangci-lint run",
    extensions: [".go"],
  },
  {
    name: "prettier",
    label: "Prettier (Web Formatter)",
    id: "web.formatter",
    tool: "prettier",
    category: "format",
    phase: "phase2_formatter",
    command: "prettier --check {{files}}",
    fix_command: "prettier --write {{files}}",
    extensions: [".ts", ".tsx", ".vue", ".js", ".json", ".md"],
  },
  {
    name: "eslint",
    label: "ESLint (Web Code Quality)",
    id: "web.linter",
    tool: "eslint",
    category: "lint",
    phase: "phase3_linter",
    command: "eslint {{files}}",
    fix_command: "eslint --fix {{files}}",
    extensions: [".ts", ".tsx", ".vue", ".js"],
  },
  {
    name: "biome",
    label: "Biome (Next-Gen Fast Web Formatter)",
    id: "web.formatter",
    tool: "biome",
    category: "format",
    phase: "phase2_formatter",
    command: "biome format {{files}}",
    fix_command: "biome format --write {{files}}",
    extensions: [".ts", ".tsx", ".js", ".json"],
  },
  {
    name: "ruff",
    label: "Ruff (Python Formatter & Linter)",
    id: "python.suite",
    tool: "ruff",
    category: "lint",
    phase: "phase3_linter",
    command: "ruff check {{files}}",
    fix_command: "ruff check --fix {{files}}",
    extensions: [".py"],
  },
  {
    name: "shfmt",
    label: "shfmt (Shell Script Formatter)",
    id: "shell.formatter",
    tool: "shfmt",
    category: "format",
    phase: "phase2_formatter",
    command: "shfmt -d {{files}}",
    fix_command: "shfmt -w {{files}}",
    extensions: [".sh", ".bash"],
  },
  {
    name: "shellcheck",
    label: "ShellCheck (Shell Quality Checker)",
    id: "shell.linter",
    tool: "shellcheck",
    category: "lint",
    phase: "phase3_linter",
    command: "shellcheck {{files}}",
    extensions: [".sh", ".bash"],
  },
  {
    name: "gitleaks",
    label: "Gitleaks (Secret & Token Protector)",
    id: "secret.gitleaks",
    tool: "gitleaks",
    category: "secret",
    phase: "phase3_linter",
    command: "gitleaks protect --staged --redact",
    extensions: ["*"],
  },
];

watch(
  () => props.isOpen,
  (open) => {
    if (open) {
      if (props.initialSlot) {
        formData.id = props.initialSlot.id;
        formData.tool = props.initialSlot.default_tool_name;
        formData.command = props.initialSlot.default_command;
        formData.fix_command = props.initialSlot.fix_command || "";
        formData.category = props.initialSlot.category || "format";
        formData.phase = props.initialSlot.phase || "phase2_formatter";
        formData.extensions = props.initialSlot.default_extensions ? [...props.initialSlot.default_extensions] : [];
        formData.enabled = props.initialSlot.enabled !== false;
      } else {
        formData.id = "";
        formData.tool = "";
        formData.command = "";
        formData.fix_command = "";
        formData.category = "format";
        formData.phase = "phase2_formatter";
        formData.extensions = props.targetExtension ? [props.targetExtension] : [".ts", ".vue"];
        formData.enabled = true;
      }
      selectedPresetTool.value = "";
      newExtInput.value = "";
    }
  },
  { immediate: true }
);

function applyToolPreset() {
  const p = toolPresets.find((item) => item.name === selectedPresetTool.value);
  if (!p) return;
  if (!formData.id || !isEditing.value) {
    formData.id = p.id;
  }
  formData.tool = p.tool;
  formData.category = p.category;
  formData.phase = p.phase;
  formData.command = p.command;
  formData.fix_command = p.fix_command || "";
  formData.extensions = [...p.extensions];
}

function addExtension() {
  const val = newExtInput.value.trim();
  if (val && !formData.extensions?.includes(val)) {
    const formatted = val.startsWith(".") || val === "*" ? val : `.${val}`;
    formData.extensions = [...(formData.extensions || []), formatted];
  }
  newExtInput.value = "";
}

function addPresetExtension(ext: string) {
  if (!formData.extensions?.includes(ext)) {
    formData.extensions = [...(formData.extensions || []), ext];
  }
}

function removeExtension(ext: string) {
  formData.extensions = formData.extensions?.filter((e) => e !== ext) || [];
}

const isValid = computed(() => {
  return formData.id.trim() !== "" && formData.command.trim() !== "";
});

function handleSave() {
  if (!isValid.value) return;
  emit("save", { ...formData });
}

function handleDelete() {
  if (formData.id) {
    emit("delete", formData.id);
  }
}
</script>

<style scoped>
.modal-backdrop {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
}

.modal-dialog {
  width: 580px;
  max-width: 90vw;
  max-height: 85vh;
  background: var(--vscode-sideBar-background, #1e1e1e);
  border: 1px solid var(--card-border, #333);
  border-radius: 8px;
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.4);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.875rem 1.25rem;
  border-bottom: 1px solid var(--card-border, #333);
  background: var(--card-bg, #252526);
}

.header-title-wrap {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.header-icon {
  font-size: 1.125rem;
  color: #58a6ff;
}

.modal-title {
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--text-primary, #fff);
  margin: 0;
}

.btn-close {
  background: transparent;
  border: none;
  color: var(--text-secondary, #8b949e);
  cursor: pointer;
  padding: 0.25rem;
  border-radius: 4px;
}

.btn-close:hover {
  color: var(--text-primary, #fff);
  background: rgba(255, 255, 255, 0.08);
}

.modal-body {
  padding: 1.25rem;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.preset-box {
  background: rgba(56, 139, 253, 0.06);
  border: 1px dashed rgba(56, 139, 253, 0.3);
  border-radius: 6px;
  padding: 0.75rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

.form-row {
  display: flex;
  gap: 1rem;
}

.flex-1 {
  flex: 1;
}

.form-label {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--text-secondary, #8b949e);
}

.form-label.required::after {
  content: " *";
  color: #f85149;
}

.form-input,
.form-select {
  padding: 0.4375rem 0.625rem;
  font-size: 0.8125rem;
  background: var(--input-bg, #181818);
  border: 1px solid var(--input-border, #333);
  border-radius: 4px;
  color: var(--text-primary, #fff);
  outline: none;
}

.form-input:focus,
.form-select:focus {
  border-color: var(--vscode-focusBorder, #0078d4);
}

.form-hint {
  font-size: 0.6875rem;
  color: var(--text-secondary, #888);
}

.tags-container {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.375rem;
  padding: 0.375rem 0.5rem;
  background: var(--input-bg, #181818);
  border: 1px solid var(--input-border, #333);
  border-radius: 4px;
  min-height: 36px;
}

.ext-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.125rem 0.375rem;
  background: rgba(56, 139, 253, 0.15);
  color: #58a6ff;
  border-radius: 3px;
  font-size: 0.75rem;
}

.remove-tag {
  cursor: pointer;
  font-size: 0.875rem;
  color: #8b949e;
}

.remove-tag:hover {
  color: #f85149;
}

.tag-inline-input {
  background: transparent;
  border: none;
  outline: none;
  color: var(--text-primary, #fff);
  font-size: 0.75rem;
  flex: 1;
  min-width: 140px;
}

.quick-tags-row {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.25rem;
  margin-top: 0.25rem;
}

.quick-label {
  font-size: 0.6875rem;
  color: var(--text-secondary, #888);
}

.quick-tag-btn {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: var(--text-secondary, #ccc);
  font-size: 0.6875rem;
  padding: 0.1rem 0.3rem;
  border-radius: 3px;
  cursor: pointer;
}

.quick-tag-btn:hover {
  background: rgba(56, 139, 253, 0.2);
  color: #58a6ff;
}

.toggle-row {
  flex-direction: row;
  align-items: center;
  justify-content: space-between;
  padding: 0.5rem 0;
}

.toggle-info {
  display: flex;
  flex-direction: column;
}

.toggle-title {
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--text-primary, #fff);
}

.toggle-sub {
  font-size: 0.6875rem;
  color: var(--text-secondary, #8b949e);
}

/* Switch */
.switch {
  position: relative;
  display: inline-block;
  width: 36px;
  height: 20px;
}

.switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: #444;
  transition: 0.2s;
  border-radius: 20px;
}

.slider:before {
  position: absolute;
  content: "";
  height: 14px;
  width: 14px;
  left: 3px;
  bottom: 3px;
  background-color: white;
  transition: 0.2s;
  border-radius: 50%;
}

input:checked + .slider {
  background-color: var(--vscode-button-background, #0078d4);
}

input:checked + .slider:before {
  transform: translateX(16px);
}

/* Footer */
.modal-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.875rem 1.25rem;
  background: var(--card-bg, #252526);
  border-top: 1px solid var(--card-border, #333);
}

.footer-right {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.btn-modal-primary {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.4375rem 1rem;
  font-size: 0.8125rem;
  font-weight: 600;
  background: var(--vscode-button-background, #0078d4);
  color: var(--vscode-button-foreground, #fff);
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.btn-modal-primary:hover:not(:disabled) {
  background: var(--vscode-button-hoverBackground, #026ec1);
}

.btn-modal-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-modal-cancel {
  padding: 0.4375rem 0.875rem;
  font-size: 0.8125rem;
  background: transparent;
  color: var(--text-secondary, #8b949e);
  border: 1px solid var(--card-border, #444);
  border-radius: 4px;
  cursor: pointer;
}

.btn-modal-cancel:hover {
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-primary, #fff);
}

.btn-modal-danger {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.4375rem 0.75rem;
  font-size: 0.8125rem;
  color: #f85149;
  background: rgba(248, 81, 73, 0.1);
  border: 1px solid rgba(248, 81, 73, 0.3);
  border-radius: 4px;
  cursor: pointer;
}

.btn-modal-danger:hover {
  background: rgba(248, 81, 73, 0.2);
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
