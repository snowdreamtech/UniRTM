<template>
  <Teleport to="body">
    <div
      v-if="isPromptOpen && pendingRequest"
      class="passphrase-modal-backdrop"
      @click.self="handleCancel"
      @keydown.esc="handleCancel"
    >
      <div class="passphrase-modal-dialog" role="dialog" aria-modal="true">
        <!-- Modal Header -->
        <div class="modal-header">
          <div class="modal-title-group">
            <span
              class="codicon"
              :class="pendingRequest.type === 'gpg' ? 'codicon-shield' : 'codicon-key'"
              :style="{
                color: pendingRequest.type === 'gpg' ? 'var(--accent-purple, #a855f7)' : 'var(--accent-cyan, #00e5ff)',
              }"
            ></span>
            <h3 class="modal-title">{{ modalTitle }}</h3>
          </div>
          <button class="modal-close-btn" title="取消 (Esc)" @click="handleCancel">
            <span class="codicon codicon-close"></span>
          </button>
        </div>

        <!-- Modal Body -->
        <div class="modal-body">
          <!-- Key / Prompt Context Pill -->
          <div class="prompt-context-pill">
            <div class="context-key-icon">
              <span class="codicon codicon-file-code"></span>
            </div>
            <div class="context-text-area">
              <div v-if="pendingRequest.targetKey" class="context-target-key font-mono">
                {{ pendingRequest.targetKey }}
              </div>
              <div class="context-prompt-message">
                {{ pendingRequest.prompt }}
              </div>
            </div>
          </div>

          <!-- Password Input -->
          <div class="form-group">
            <label class="form-label" for="passphrase-input">
              {{
                pendingRequest.type === "gpg"
                  ? t("credential.gpg_passphrase_label") || "GPG 签名私钥密码"
                  : t("credential.passphrase_label") || "私钥密码 (Passphrase)"
              }}
              <span class="required">*</span>
            </label>
            <div class="input-password-wrapper">
              <input
                id="passphrase-input"
                ref="inputRef"
                v-model="passphrase"
                :type="showPassword ? 'text' : 'password'"
                class="form-input font-mono"
                :placeholder="
                  pendingRequest.type === 'gpg'
                    ? t('credential.gpg_placeholder') || '输入 GPG 密码短语...'
                    : t('credential.ssh_placeholder') || '输入 SSH 私钥密码短语...'
                "
                autocomplete="off"
                spellcheck="false"
                @keydown.enter="handleConfirm"
                @keyup="handleKeyUp"
              />
              <button
                type="button"
                class="toggle-eye-btn"
                :title="showPassword ? t('common.hide') || '隐藏密码' : t('common.show') || '显示密码'"
                @click="showPassword = !showPassword"
              >
                <span class="codicon" :class="showPassword ? 'codicon-eye' : 'codicon-eye-closed'"></span>
              </button>
            </div>

            <!-- Caps Lock Indicator -->
            <div v-if="capsLockOn" class="caps-lock-warning">
              <span class="codicon codicon-warning"></span>
              <span>{{ t("credential.caps_lock") || "大写锁定已开启 (Caps Lock On)" }}</span>
            </div>
          </div>

          <!-- Session Remember Checkbox -->
          <div class="form-checkbox-row">
            <label class="checkbox-label">
              <input v-model="rememberSession" type="checkbox" class="form-checkbox" />
              <span>{{ t("credential.remember_session") || "在当前工作区会话期间记住此密码 (Session Cache)" }}</span>
            </label>
            <span class="checkbox-hint">{{
              t("credential.remember_hint") || "仅在当前运行的内存中临时安全保留，退出或重载后失效"
            }}</span>
          </div>

          <!-- Error Alert Banner -->
          <div v-if="lastError" class="error-banner">
            <span class="codicon codicon-error"></span>
            <span>{{ lastError }}</span>
          </div>
        </div>

        <!-- Modal Footer Actions -->
        <div class="modal-footer">
          <button type="button" class="btn btn-secondary" :disabled="isSubmitting" @click="handleCancel">
            {{ t("common.cancel") || "取消 (Esc)" }}
          </button>
          <button type="button" class="btn btn-primary" :disabled="!passphrase || isSubmitting" @click="handleConfirm">
            <span v-if="isSubmitting" class="codicon codicon-loading codicon-modifier-spin"></span>
            <span>{{
              isSubmitting
                ? t("credential.validating_btn") || "验证中..."
                : t("credential.unlock_btn") || "解锁并继续 (Enter)"
            }}</span>
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from "vue";
import { t } from "../../i18n";
import { useCredentialBroker } from "../../composables/useCredentialBroker";

const { pendingRequest, isPromptOpen, isSubmitting, lastError, respondCredential, cancelCredential } =
  useCredentialBroker();

const passphrase = ref("");
const rememberSession = ref(true);
const showPassword = ref(false);
const capsLockOn = ref(false);
const inputRef = ref<HTMLInputElement | null>(null);

const modalTitle = computed(() => {
  if (!pendingRequest.value) return t("credential.unlock_git_title") || "解锁凭据";
  switch (pendingRequest.value.type) {
    case "gpg":
      return t("credential.unlock_gpg_title") || "解锁 GPG 签名私钥";
    case "ssh":
      return t("credential.unlock_ssh_title") || "解锁 SSH 私钥凭据";
    default:
      return t("credential.unlock_git_title") || "输入 Git 认证凭据";
  }
});

watch(isPromptOpen, (isOpen) => {
  if (isOpen) {
    passphrase.value = "";
    showPassword.value = false;
    capsLockOn.value = false;
    nextTick(() => {
      inputRef.value?.focus();
    });
  }
});

function handleKeyUp(e: KeyboardEvent) {
  capsLockOn.value = e.getModifierState && e.getModifierState("CapsLock");
}

async function handleConfirm() {
  if (!pendingRequest.value || !passphrase.value || isSubmitting.value) return;
  await respondCredential(pendingRequest.value.id, passphrase.value, rememberSession.value);
}

function handleCancel() {
  if (isSubmitting.value) return;
  cancelCredential();
}
</script>

<style scoped>
.passphrase-modal-backdrop {
  position: fixed;
  inset: 0;
  z-index: 999999;
  background: rgba(0, 0, 0, 0.65);
  backdrop-filter: blur(6px);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
  animation: fadeIn 0.15s ease-out;
}

.passphrase-modal-dialog {
  width: 100%;
  max-width: 480px;
  background: var(--card-bg, #111827);
  border: 1px solid var(--card-border, #1f2937);
  border-radius: 8px;
  box-shadow:
    0 20px 25px -5px rgba(0, 0, 0, 0.5),
    0 10px 10px -5px rgba(0, 0, 0, 0.4);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  animation: scaleUp 0.15s cubic-bezier(0.16, 1, 0.3, 1);
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.75rem 1rem;
  border-bottom: 1px solid var(--card-border, #1f2937);
  background: rgba(255, 255, 255, 0.02);
}

.modal-title-group {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.modal-title {
  margin: 0;
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--text-main, #f3f4f6);
}

.modal-close-btn {
  background: transparent;
  border: none;
  color: var(--text-muted, #9ca3af);
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.modal-close-btn:hover {
  color: var(--text-main, #f3f4f6);
  background: rgba(255, 255, 255, 0.08);
}

.modal-body {
  padding: 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.9rem;
}

.prompt-context-pill {
  display: flex;
  align-items: flex-start;
  gap: 0.65rem;
  padding: 0.65rem 0.75rem;
  background: rgba(0, 229, 255, 0.05);
  border: 1px solid rgba(0, 229, 255, 0.15);
  border-radius: 6px;
}

.context-key-icon {
  color: var(--accent-cyan, #00e5ff);
  margin-top: 1px;
}

.context-text-area {
  display: flex;
  flex-direction: column;
  gap: 2px;
  overflow: hidden;
}

.context-target-key {
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--text-main, #f3f4f6);
  word-break: break-all;
}

.context-prompt-message {
  font-size: 0.7rem;
  color: var(--text-muted, #9ca3af);
  line-height: 1.35;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.form-label {
  font-size: 0.75rem;
  font-weight: 500;
  color: var(--text-muted, #9ca3af);
}

.required {
  color: #ef4444;
}

.input-password-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.form-input {
  width: 100%;
  padding: 0.45rem 2rem 0.45rem 0.65rem;
  background: var(--input-bg, #0b0f19);
  border: 1px solid var(--card-border, #1f2937);
  border-radius: 4px;
  color: var(--text-main, #f3f4f6);
  font-size: 0.8rem;
  transition:
    border-color 0.15s ease,
    box-shadow 0.15s ease;
}

.form-input:focus {
  outline: none;
  border-color: var(--accent-cyan, #00e5ff);
  box-shadow: 0 0 0 1px var(--accent-cyan, #00e5ff);
}

.toggle-eye-btn {
  position: absolute;
  right: 6px;
  background: transparent;
  border: none;
  color: var(--text-muted, #9ca3af);
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  display: inline-flex;
}

.toggle-eye-btn:hover {
  color: var(--text-main, #f3f4f6);
}

.caps-lock-warning {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  font-size: 0.7rem;
  color: #f59e0b;
  margin-top: 2px;
}

.form-checkbox-row {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  font-size: 0.75rem;
  color: var(--text-main, #f3f4f6);
  cursor: pointer;
}

.form-checkbox {
  cursor: pointer;
  accent-color: var(--accent-cyan, #00e5ff);
}

.checkbox-hint {
  font-size: 0.68rem;
  color: var(--text-muted, #9ca3af);
  margin-left: 1.45rem;
}

.error-banner {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  padding: 0.5rem 0.75rem;
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.3);
  border-radius: 4px;
  color: #ef4444;
  font-size: 0.75rem;
}

.modal-footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0.5rem;
  padding: 0.75rem 1rem;
  border-top: 1px solid var(--card-border, #1f2937);
  background: rgba(255, 255, 255, 0.02);
}

.btn {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.4rem 0.85rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
  border: none;
}

.btn-secondary {
  background: rgba(255, 255, 255, 0.08);
  color: var(--text-main, #f3f4f6);
}

.btn-secondary:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.12);
}

.btn-primary {
  background: var(--accent-cyan, #00e5ff);
  color: #0b0f19;
  font-weight: 600;
}

.btn-primary:hover:not(:disabled) {
  filter: brightness(1.1);
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

@keyframes scaleUp {
  from {
    opacity: 0;
    transform: scale(0.95);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}
</style>
