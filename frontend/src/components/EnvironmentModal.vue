<template>
  <div v-if="isOpen" class="modal-overlay" @click.self="$emit('close')" @keydown.esc="$emit('close')" tabindex="-1">
    <div class="modal-card glass-modal" id="env-editor-modal">
      <!-- Modal Header -->
      <div class="modal-header">
        <div class="header-title">
          <TechIcon id="workbench:environments" :size="20" />
          <div>
            <h3>{{ isEdit ? (t("environments.edit_btn") + " " + t("nav.environments")) : (t("environments.create_btn")) }}</h3>
            <span class="sub-title">
              {{ form.scope === 'user' ? (t("environments.scope_user") || "用户全局环境变量") : (t("environments.scope_workspace") || "工作区项目环境变量") }}
              <span class="scope-file-tag font-mono">{{ form.scope === 'user' ? '~/.config/unirtm/unirtm.toml' : '.unirtm.toml' }}</span>
            </span>
          </div>
        </div>
        <button class="close-btn" @click="$emit('close')" title="Close">✕</button>
      </div>

      <!-- Modal Body -->
      <div class="modal-body">
        <form @submit.prevent="handleSubmit">

          <!-- Variable Key -->
          <div class="form-group">
            <label class="form-label required">
              <span>{{ t("environments.key_label") || "变量名 (Key)" }}</span>
              <span class="label-hint">（大写字母、数字及下划线组合）</span>
            </label>
            <div class="input-with-action">
              <input
                v-model="form.key"
                type="text"
                class="form-input font-mono"
                :placeholder="t('environments.key_placeholder') || '如 GITHUB_TOKEN, PORT, NODE_ENV'"
                required
                :disabled="saving"
                @input="handleKeyInput"
              />
              <button
                type="button"
                class="btn-uppercase"
                @click="form.key = form.key.toUpperCase()"
                title="转换为全大写"
              >
                AA
              </button>
            </div>

            <!-- Preset Chips -->
            <div class="preset-chips-wrap">
              <span class="preset-hint">常用预设:</span>
              <button
                v-for="chip in PRESET_KEYS"
                :key="chip"
                type="button"
                class="preset-chip font-mono"
                @click="form.key = chip"
              >
                {{ chip }}
              </button>
            </div>

            <!-- Sensitive Detection Hint -->
            <div v-if="isSensitiveKey" class="sensitive-detected-banner">
              <span class="sec-icon">🔒</span>
              <span>系统已自动判定为<b>敏感密钥</b>，保存后默认脱密遮罩显示，需点击眼睛图标方可查看明文。</span>
            </div>
          </div>

          <!-- Variable Value -->
          <div class="form-group">
            <label class="form-label required">
              <span>{{ t("environments.value_label") || "变量值 (Value)" }}</span>
              <span class="label-hint">（点击小眼睛可自由切换密文/明文预览）</span>
            </label>
            <div class="value-input-wrapper">
              <input
                v-model="form.value"
                :type="isValueRevealed ? 'text' : 'password'"
                class="form-input font-mono value-input"
                :placeholder="t('environments.value_placeholder') || '请输入变量值'"
                required
                :disabled="saving"
              />
              <button
                type="button"
                class="btn-eye-toggle"
                @click="isValueRevealed = !isValueRevealed"
                :title="isValueRevealed ? '隐藏密码' : '显示明文'"
              >
                <span>{{ isValueRevealed ? '👁️' : '👁️‍🗨️' }}</span>
              </button>
            </div>
          </div>

          <!-- Modal Footer Actions -->
          <div class="modal-footer">
            <button
              type="button"
              class="btn-cancel"
              @click="$emit('close')"
              :disabled="saving"
            >
              {{ t("common.cancel") || "取消" }}
            </button>
            <button
              type="submit"
              class="btn-submit"
              :disabled="saving || !form.key.trim()"
            >
              <TechIcon v-if="saving" id="workbench:sync" :size="14" class="spin-icon" />
              <span>{{ saving ? "正在保存..." : (isEdit ? (t("common.save") || "保存更新") : (t("common.create") || "立即创建")) }}</span>
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from "vue";
import { t } from "../i18n";
import TechIcon from "./common/TechIcon.vue";

const PRESET_KEYS = ["NODE_ENV", "PORT", "GITHUB_TOKEN", "HTTP_PROXY", "HTTPS_PROXY", "DEBUG", "API_BASE_URL"];

const props = withDefaults(
  defineProps<{
    isOpen: boolean;
    mode?: "create" | "edit";
    initialData?: { name: string; value: string; scope?: "workspace" | "user" };
    activeScope?: "workspace" | "user";
  }>(),
  {
    mode: "create",
    activeScope: "workspace",
  }
);

const emit = defineEmits<{
  (e: "close"): void;
  (
    e: "save",
    payload: {
      key: string;
      original_key?: string;
      value: string;
      scope: "workspace" | "user";
    }
  ): void;
}>();

const isEdit = computed(() => props.mode === "edit");
const saving = ref(false);
const isValueRevealed = ref(false);

const form = ref({
  key: "",
  original_key: "",
  value: "",
  scope: "workspace" as "workspace" | "user",
});

const isSensitiveKey = computed(() => {
  const k = form.value.key || "";
  return /(TOKEN|KEY|SECRET|PASSWORD|PASS|AUTH|PRIVATE|CREDENTIAL|APIKEY|ACCESS|CERT|SALT|HASH|SIGNATURE)/i.test(k);
});

function handleKeyInput(e: Event) {
  const input = e.target as HTMLInputElement;
  form.value.key = input.value.trim();
}

watch(
  () => props.isOpen,
  (open) => {
    if (open) {
      isValueRevealed.value = false;
      if (props.mode === "edit" && props.initialData) {
        form.value = {
          key: props.initialData.name,
          original_key: props.initialData.name,
          value: props.initialData.value,
          scope: props.initialData.scope || props.activeScope,
        };
      } else {
        form.value = {
          key: "",
          original_key: "",
          value: "",
          scope: props.activeScope,
        };
      }
    }
  },
  { immediate: true }
);

function handleSubmit() {
  const cleanKey = form.value.key.trim().toUpperCase();
  if (!cleanKey) return;

  saving.value = true;
  emit("save", {
    key: cleanKey,
    original_key: isEdit.value ? form.value.original_key : undefined,
    value: form.value.value,
    scope: form.value.scope,
  });
  saving.value = false;
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 1000;
  background: rgba(0, 0, 0, 0.65);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
}

.modal-card {
  width: 100%;
  max-width: 540px;
  background: var(--card-bg, #0d1424);
  border: 1px solid var(--card-border, rgba(0, 229, 255, 0.25));
  border-radius: 14px;
  box-shadow: 0 16px 48px rgba(0, 0, 0, 0.45);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  animation: modal-enter 0.2s cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes modal-enter {
  from {
    opacity: 0;
    transform: scale(0.96) translateY(-8px);
  }
  to {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 1.25rem;
  border-bottom: 1px solid var(--card-border, rgba(255, 255, 255, 0.08));
}

.header-title {
  display: flex;
  align-items: center;
  gap: 10px;
}

.header-title h3 {
  margin: 0;
  font-size: 1.05rem;
  font-weight: 600;
  color: var(--text-main);
}

.sub-title {
  display: block;
  font-size: 0.76rem;
  color: var(--text-muted);
  margin-top: 2px;
}

.scope-file-tag {
  color: var(--accent-cyan);
  margin-left: 4px;
  background: rgba(0, 229, 255, 0.1);
  padding: 1px 6px;
  border-radius: 4px;
}

.close-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  font-size: 1.1rem;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 6px;
  transition: all 0.2s;
}

.close-btn:hover {
  color: var(--text-main);
  background: rgba(255, 255, 255, 0.08);
}

.modal-body {
  padding: 1.25rem;
}

.form-group {
  margin-bottom: 1.1rem;
}

.form-label {
  display: flex;
  align-items: baseline;
  gap: 6px;
  margin-bottom: 6px;
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text-main);
}

.form-label.required::after {
  content: "*";
  color: #ef4444;
}

.label-hint {
  font-size: 0.74rem;
  font-weight: normal;
  color: var(--text-muted);
}


.input-with-action {
  position: relative;
  display: flex;
  align-items: center;
}

.form-input {
  width: 100%;
  padding: 8px 12px;
  background: var(--input-bg, rgba(0, 0, 0, 0.25));
  border: 1px solid var(--card-border, rgba(255, 255, 255, 0.12));
  border-radius: 6px;
  color: var(--text-main);
  font-size: 0.88rem;
  box-sizing: border-box;
  transition: all 0.2s;
}

.form-input:focus {
  outline: none;
  border-color: var(--accent-cyan);
  box-shadow: 0 0 8px rgba(0, 229, 255, 0.2);
}

.btn-uppercase {
  position: absolute;
  right: 6px;
  padding: 3px 6px;
  font-size: 0.72rem;
  font-weight: 700;
  background: rgba(255, 255, 255, 0.08);
  border: 1px solid rgba(255, 255, 255, 0.15);
  border-radius: 4px;
  color: var(--text-muted);
  cursor: pointer;
}

.btn-uppercase:hover {
  color: var(--text-main);
  background: rgba(255, 255, 255, 0.16);
}

.preset-chips-wrap {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 6px;
}

.preset-hint {
  font-size: 0.72rem;
  color: var(--text-muted);
}

.preset-chip {
  padding: 2px 7px;
  font-size: 0.72rem;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid var(--card-border, rgba(255, 255, 255, 0.1));
  border-radius: 4px;
  color: var(--text-muted);
  cursor: pointer;
  transition: all 0.15s;
}

.preset-chip:hover {
  color: var(--accent-cyan);
  border-color: var(--accent-cyan);
  background: rgba(0, 229, 255, 0.08);
}

.sensitive-detected-banner {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 8px;
  padding: 6px 10px;
  background: rgba(245, 158, 11, 0.12);
  border: 1px solid rgba(245, 158, 11, 0.3);
  border-radius: 6px;
  color: #f59e0b;
  font-size: 0.76rem;
  line-height: 1.4;
}

.value-input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.value-input {
  padding-right: 36px;
}

.btn-eye-toggle {
  position: absolute;
  right: 8px;
  background: transparent;
  border: none;
  cursor: pointer;
  font-size: 0.95rem;
  padding: 2px 4px;
  border-radius: 4px;
  transition: opacity 0.2s;
  color: var(--text-muted);
}

.btn-eye-toggle:hover {
  opacity: 0.8;
}

.modal-footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 1.5rem;
  padding-top: 1rem;
  border-top: 1px solid var(--card-border, rgba(255, 255, 255, 0.08));
}

.btn-cancel {
  padding: 7px 16px;
  background: transparent;
  border: 1px solid var(--card-border, rgba(255, 255, 255, 0.15));
  border-radius: 6px;
  color: var(--text-muted);
  font-size: 0.84rem;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-cancel:hover {
  color: var(--text-main);
  background: rgba(255, 255, 255, 0.05);
}

.btn-submit {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 7px 18px;
  background: var(--accent-cyan);
  border: 1px solid var(--accent-cyan);
  border-radius: 6px;
  color: #050b14;
  font-size: 0.84rem;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-submit:hover:not(:disabled) {
  filter: brightness(1.1);
  transform: translateY(-1px);
}

.btn-submit:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Light Theme Adaptations */
:root[data-theme="light"] .modal-card,
[data-theme="light"] .modal-card {
  background: #ffffff;
  border-color: #cbd5e1;
  box-shadow: 0 16px 40px rgba(0, 0, 0, 0.15);
}

:root[data-theme="light"] .scope-radio-item,
[data-theme="light"] .scope-radio-item {
  background: #f8fafc;
  border-color: #e2e8f0;
}

:root[data-theme="light"] .scope-radio-item.selected,
[data-theme="light"] .scope-radio-item.selected {
  background: rgba(2, 132, 199, 0.08);
  border-color: #0284c7;
}

:root[data-theme="light"] .form-input,
[data-theme="light"] .form-input {
  background: #f8fafc;
  border-color: #cbd5e1;
  color: #0f172a;
}

:root[data-theme="light"] .btn-submit,
[data-theme="light"] .btn-submit {
  color: #ffffff !important;
}

:root[data-theme="light"] .preset-chip,
[data-theme="light"] .preset-chip {
  background: #f1f5f9;
  border-color: #cbd5e1;
  color: #475569;
}

:root[data-theme="light"] .preset-chip:hover,
[data-theme="light"] .preset-chip:hover {
  background: rgba(2, 132, 199, 0.1);
  color: #0284c7;
  border-color: #0284c7;
}
</style>
