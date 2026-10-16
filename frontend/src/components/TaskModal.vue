<template>
  <div v-if="isOpen" class="modal-overlay" @click.self="$emit('close')" @keydown.esc="$emit('close')" tabindex="-1">
    <div class="modal-card glass-modal" id="task-editor-modal">
      <!-- Modal Header -->
      <div class="modal-header">
        <div class="header-title">
          <TechIcon id="workbench:tasks" :size="20" />
          <div>
            <h3>{{ isEdit ? (t("tasks.edit_title") || "编辑任务") : (t("tasks.create_title") || "新建工程任务") }}</h3>
            <span class="sub-title">
              {{ scope === 'user' ? (t("runtimes.scope_user") || "用户全局配置") : (t("runtimes.scope_workspace") || "工作区项目配置") }}
              <span class="scope-file-tag font-mono">{{ scope === 'user' ? '~/.config/unirtm/unirtm.toml' : '.unirtm.toml' }}</span>
            </span>
          </div>
        </div>
        <button class="close-btn" @click="$emit('close')" title="Close">✕</button>
      </div>

      <!-- Modal Body -->
      <div class="modal-body">
        <form @submit.prevent="handleSubmit">
          <!-- Task Name -->
          <div class="form-group">
            <label class="form-label required">
              <span>{{ t("tasks.field_name") || "任务名称" }}</span>
              <span class="label-hint">{{ t("tasks.name_hint") || "（唯一标识，如 build, test, lint:fix）" }}</span>
            </label>
            <input
              v-model="form.name"
              type="text"
              class="form-input font-mono"
              placeholder="例如: build:prod"
              required
              :disabled="saving"
            />
          </div>

          <!-- Command / Run Script -->
          <div class="form-group">
            <label class="form-label required">
              <span>{{ t("tasks.field_command") || "执行命令 (Command / Script)" }}</span>
              <span class="label-hint">{{ t("tasks.command_hint") || "（支持单行命令或多行脚本）" }}</span>
            </label>
            <textarea
              v-model="form.command"
              class="form-textarea font-mono"
              placeholder="例如: npm run build&#10;或: go test -v ./..."
              rows="4"
              required
              :disabled="saving"
            ></textarea>
          </div>

          <!-- Description -->
          <div class="form-group">
            <label class="form-label">
              <span>{{ t("tasks.field_desc") || "任务描述" }}</span>
              <span class="label-hint">{{ t("tasks.desc_hint") || "（选填，说明该任务的用途与触发时机）" }}</span>
            </label>
            <input
              v-model="form.description"
              type="text"
              class="form-input"
              placeholder="例如: 编译前端代码并输出到 dist 目录"
              :disabled="saving"
            />
          </div>

          <!-- Depends (Dependencies) -->
          <div class="form-group">
            <label class="form-label">
              <span>{{ t("tasks.field_depends") || "依赖前置任务 (Depends)" }}</span>
              <span class="label-hint">{{ t("tasks.depends_hint") || "（选填，执行当前任务前必须先完成的任务）" }}</span>
            </label>
            <div class="depends-selector-wrap">
              <div v-if="filteredAvailableTasks.length > 0" class="dep-pills-list">
                <button
                  v-for="taskName in filteredAvailableTasks"
                  :key="taskName"
                  type="button"
                  class="dep-pill-btn font-mono"
                  :class="{ selected: form.depends.includes(taskName) }"
                  @click="toggleDependency(taskName)"
                >
                  <span class="check-mark">{{ form.depends.includes(taskName) ? '✓' : '+' }}</span>
                  <span>{{ taskName }}</span>
                </button>
              </div>
              <div class="dep-custom-input-row">
                <input
                  v-model="customDepInput"
                  type="text"
                  class="form-input font-mono small-input"
                  placeholder="手动输入依赖任务名并按回车..."
                  @keydown.enter.prevent="addCustomDependency"
                  :disabled="saving"
                />
                <button
                  type="button"
                  class="tool-action-btn secondary small-btn"
                  @click="addCustomDependency"
                  :disabled="!customDepInput.trim() || saving"
                >
                  {{ t("tasks.add_dep") || "添加依赖" }}
                </button>
              </div>
              <div v-if="form.depends.length > 0" class="current-selected-deps">
                <span class="selected-label">{{ t("tasks.dep_order") || "当前选定顺序:" }}</span>
                <span
                  v-for="dep in form.depends"
                  :key="dep"
                  class="selected-dep-badge font-mono"
                >
                  {{ dep }}
                  <span class="remove-dep" @click="removeDependency(dep)">✕</span>
                </span>
              </div>
            </div>
          </div>

          <!-- Modal Footer Actions -->
          <div class="modal-footer">
            <button
              type="button"
              class="tool-action-btn secondary"
              @click="$emit('close')"
              :disabled="saving"
            >
              {{ t("common.cancel") || "取消" }}
            </button>
            <button
              type="submit"
              class="tool-action-btn primary"
              :disabled="saving || !form.name.trim() || !form.command.trim()"
            >
              <TechIcon v-if="saving" id="workbench:sync" :size="14" class="spin-icon" />
              <span>{{ saving ? (t("common.saving") || "保存中...") : (isEdit ? (t("common.save_changes") || "保存修改") : (t("tasks.create_btn") || "创建任务")) }}</span>
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, computed } from "vue";
import { t } from "../i18n";
import TechIcon from "./common/TechIcon.vue";

export interface TaskFormData {
  name: string;
  originalName?: string;
  command: string;
  description: string;
  depends: string[];
}

const props = defineProps<{
  isOpen: boolean;
  isEdit: boolean;
  initialTask?: {
    name: string;
    command: string;
    description: string;
    depends?: string[];
  } | null;
  scope: string;
  allTaskNames: string[];
  saving?: boolean;
}>();

const emit = defineEmits<{
  (e: "close"): void;
  (e: "save", data: TaskFormData): void;
}>();

const form = ref<TaskFormData>({
  name: "",
  command: "",
  description: "",
  depends: [],
});

const customDepInput = ref("");

watch(
  () => props.isOpen,
  (open) => {
    if (open) {
      if (props.isEdit && props.initialTask) {
        form.value = {
          name: props.initialTask.name,
          originalName: props.initialTask.name,
          command: props.initialTask.command || "",
          description: props.initialTask.description || "",
          depends: props.initialTask.depends ? [...props.initialTask.depends] : [],
        };
      } else {
        form.value = {
          name: "",
          command: "",
          description: "",
          depends: [],
        };
      }
      customDepInput.value = "";
    }
  },
  { immediate: true }
);

const filteredAvailableTasks = computed(() => {
  return props.allTaskNames.filter((n) => n !== form.value.name);
});

function toggleDependency(depName: string) {
  const idx = form.value.depends.indexOf(depName);
  if (idx > -1) {
    form.value.depends.splice(idx, 1);
  } else {
    form.value.depends.push(depName);
  }
}

function addCustomDependency() {
  const val = customDepInput.value.trim();
  if (val && !form.value.depends.includes(val) && val !== form.value.name) {
    form.value.depends.push(val);
  }
  customDepInput.value = "";
}

function removeDependency(depName: string) {
  const idx = form.value.depends.indexOf(depName);
  if (idx > -1) {
    form.value.depends.splice(idx, 1);
  }
}

function handleSubmit() {
  if (!form.value.name.trim() || !form.value.command.trim()) return;
  emit("save", {
    name: form.value.name.trim(),
    originalName: form.value.originalName,
    command: form.value.command.trim(),
    description: form.value.description.trim(),
    depends: [...form.value.depends],
  });
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.glass-modal {
  background: var(--modal-bg);
  border: 1px solid var(--card-border);
  box-shadow:
    0 20px 48px -10px rgba(0, 0, 0, 0.35),
    0 0 24px var(--accent-cyan-glow);
  border-radius: 16px;
  width: 92%;
  max-width: 620px;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  color: var(--text-main);
  transition: all 0.25s ease;
}

.modal-header {
  padding: 1.15rem 1.4rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: var(--section-bg);
  border-bottom: 1px solid var(--card-border);
}

.header-title {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.header-title h3 {
  margin: 0;
  font-size: 1.05rem;
  font-weight: 700;
  color: var(--text-main);
}

.sub-title {
  display: flex;
  align-items: center;
  font-size: 0.75rem;
  color: var(--text-muted);
  margin-top: 3px;
}

.scope-file-tag {
  background: var(--accent-cyan-glow);
  color: var(--accent-cyan);
  border: 1px solid var(--card-border);
  padding: 1px 7px;
  border-radius: 4px;
  margin-left: 6px;
  font-size: 0.72rem;
  font-weight: 600;
}

.close-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  font-size: 1.1rem;
  cursor: pointer;
  padding: 0.25rem 0.5rem;
  border-radius: 6px;
  transition: all 0.2s;
}

.close-btn:hover {
  background: var(--card-bg);
  color: var(--text-main);
}

.modal-body {
  padding: 1.25rem 1.4rem;
  overflow-y: auto;
  box-sizing: border-box;
  background: var(--modal-bg);
}

.form-group {
  margin-bottom: 1.15rem;
}

.form-label {
  display: flex;
  align-items: center;
  font-size: 0.82rem;
  font-weight: 600;
  margin-bottom: 0.45rem;
  color: var(--text-main);
}

.form-label.required::after {
  content: " *";
  color: #ef4444;
}

.label-hint {
  font-size: 0.72rem;
  font-weight: normal;
  color: var(--text-muted);
  margin-left: 4px;
}

.form-input,
.form-textarea {
  width: 100%;
  box-sizing: border-box;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  border-radius: 8px;
  padding: 0.55rem 0.8rem;
  font-size: 0.82rem;
  color: var(--text-main);
  outline: none;
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}

.form-input:focus,
.form-textarea:focus {
  border-color: var(--card-border-active);
  box-shadow: 0 0 0 3px var(--accent-cyan-glow);
}

.form-textarea {
  resize: vertical;
  min-height: 85px;
  line-height: 1.45;
}

.depends-selector-wrap {
  display: flex;
  flex-direction: column;
  gap: 0.65rem;
  background: var(--section-bg);
  border: 1px solid var(--card-border);
  border-radius: 10px;
  padding: 0.85rem;
}

.dep-pills-list {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
}

.dep-pill-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  color: var(--text-muted);
  font-size: 0.75rem;
  padding: 0.25rem 0.55rem;
  border-radius: 5px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.dep-pill-btn:hover {
  background: var(--input-bg);
  color: var(--text-main);
  border-color: var(--card-border-active);
  transform: translateY(-1px);
}

.dep-pill-btn.selected {
  background: var(--accent-cyan-glow);
  border-color: var(--accent-cyan);
  color: var(--accent-cyan);
  font-weight: 600;
}

.dep-custom-input-row {
  display: flex;
  gap: 0.5rem;
}

.small-input {
  flex: 1;
  padding: 0.4rem 0.65rem;
  font-size: 0.78rem;
}

.small-btn {
  padding: 0.4rem 0.8rem;
  font-size: 0.75rem;
  white-space: nowrap;
}

.current-selected-deps {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  flex-wrap: wrap;
  padding-top: 0.4rem;
  border-top: 1px dashed var(--card-border);
}

.selected-label {
  font-size: 0.72rem;
  color: var(--text-muted);
  font-weight: 500;
}

.selected-dep-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  background: var(--accent-cyan-glow);
  border: 1px solid var(--accent-cyan);
  color: var(--accent-cyan);
  font-size: 0.72rem;
  font-weight: 600;
  padding: 0.15rem 0.45rem;
  border-radius: 4px;
}

.remove-dep {
  cursor: pointer;
  opacity: 0.7;
  font-size: 0.75rem;
}

.remove-dep:hover {
  opacity: 1;
  color: #ef4444;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  margin-top: 1.5rem;
  padding-top: 1rem;
  border-top: 1px solid var(--card-border);
}

.tool-action-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.45rem 1.15rem;
  font-size: 0.82rem;
  font-weight: 600;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
  border: none;
}

.tool-action-btn.primary {
  background: linear-gradient(135deg, var(--accent-cyan), #0284c7);
  color: #ffffff;
  border: none;
  font-weight: 600;
  box-shadow: 0 2px 8px var(--accent-cyan-glow);
}

.tool-action-btn.primary:hover:not(:disabled) {
  box-shadow: 0 4px 14px var(--accent-cyan-glow);
  transform: translateY(-1px);
}

.tool-action-btn.secondary {
  border: 1px solid var(--card-border);
  background: var(--card-bg);
  color: var(--text-main);
}

.tool-action-btn.secondary:hover:not(:disabled) {
  background: var(--input-bg);
  border-color: var(--card-border-active);
  color: var(--text-main);
}

.tool-action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.spin-icon {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
