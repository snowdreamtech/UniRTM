<template>
  <div v-if="isOpen" class="drawer-backdrop" @click.self="$emit('close')">
    <div class="drawer-panel" role="dialog" aria-modal="true">
      <!-- Drawer Header -->
      <div class="drawer-header">
        <div class="header-title-box">
          <span class="codicon codicon-code text-blue"></span>
          <div class="header-titles">
            <h2 class="drawer-title">{{ t("guard.toml.title") || ".unirtm.toml Fallback Editor" }}</h2>
            <span class="drawer-sub">{{
              t("guard.toml.subtitle") || "Directly inspect or edit the project's source-of-truth configuration."
            }}</span>
          </div>
        </div>
        <div class="header-actions">
          <button class="btn-refresh" :title="t('common.refresh') || 'Reload'" @click="loadToml">
            <span class="codicon codicon-refresh" :class="{ spin: isLoading }"></span>
          </button>
          <button class="btn-close" @click="$emit('close')">
            <span class="codicon codicon-close"></span>
          </button>
        </div>
      </div>

      <!-- Drawer Content -->
      <div class="drawer-body">
        <div class="editor-toolbar">
          <span class="file-path-badge font-mono">.unirtm.toml</span>
          <span v-if="hasChanges" class="dirty-badge">● {{ t("guard.toml.unsaved") || "Unsaved Changes" }}</span>
        </div>

        <textarea
          v-model="tomlContent"
          class="toml-textarea font-mono"
          spellcheck="false"
          placeholder="# .unirtm.toml configuration content..."
        ></textarea>

        <div v-if="errorMessage" class="error-banner">
          <span class="codicon codicon-error text-red"></span>
          <span>{{ errorMessage }}</span>
        </div>
      </div>

      <!-- Drawer Footer -->
      <div class="drawer-footer">
        <div class="footer-left">
          <span class="footer-hint">
            {{ t("guard.toml.hint") || "Edits are verified for TOML syntax validity before saving." }}
          </span>
        </div>
        <div class="footer-right">
          <button type="button" class="btn-cancel" @click="$emit('close')">
            {{ t("common.cancel") || "Cancel" }}
          </button>
          <button type="button" class="btn-save" :disabled="isSaving || !hasChanges" @click="handleSave">
            <span class="codicon" :class="isSaving ? 'codicon-loading spin' : 'codicon-save'"></span>
            <span>{{
              isSaving ? t("guard.actions.saving") || "Saving..." : t("guard.actions.save") || "Save .unirtm.toml"
            }}</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, computed } from "vue";
import { t } from "../../i18n";
import { useGuard } from "../../composables/useGuard";

const props = defineProps<{
  isOpen: boolean;
}>();

const emit = defineEmits<{
  (e: "close"): void;
  (e: "saved"): void;
}>();

const { fetchRawTomlConfig, saveRawTomlConfig } = useGuard();

const tomlContent = ref("");
const originalContent = ref("");
const isLoading = ref(false);
const isSaving = ref(false);
const errorMessage = ref("");

const hasChanges = computed(() => tomlContent.value !== originalContent.value);

watch(
  () => props.isOpen,
  (open) => {
    if (open) {
      loadToml();
    }
  },
  { immediate: true }
);

async function loadToml() {
  isLoading.value = true;
  errorMessage.value = "";
  try {
    const c = await fetchRawTomlConfig();
    tomlContent.value = c;
    originalContent.value = c;
  } catch (err: any) {
    errorMessage.value = err.message || "Failed to load TOML configuration";
  } finally {
    isLoading.value = false;
  }
}

async function handleSave() {
  isSaving.value = true;
  errorMessage.value = "";
  try {
    const success = await saveRawTomlConfig(tomlContent.value);
    if (success) {
      originalContent.value = tomlContent.value;
      emit("saved");
      emit("close");
    } else {
      errorMessage.value = "Failed to save .unirtm.toml. Please ensure TOML syntax is valid.";
    }
  } catch (err: any) {
    errorMessage.value = err.message || "Failed to save TOML configuration";
  } finally {
    isSaving.value = false;
  }
}
</script>

<style scoped>
.drawer-backdrop {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(2px);
  display: flex;
  justify-content: flex-end;
  z-index: 9999;
}

.drawer-panel {
  width: 620px;
  max-width: 90vw;
  height: 100vh;
  background: var(--modal-bg);
  border-left: 1px solid var(--card-border);
  box-shadow: -10px 0 30px rgba(0, 0, 0, 0.3);
  display: flex;
  flex-direction: column;
  animation: slideIn 0.2s cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes slideIn {
  from {
    transform: translateX(100%);
  }
  to {
    transform: translateX(0);
  }
}

.drawer-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 1.25rem;
  background: var(--card-bg);
  border-bottom: 1px solid var(--card-border);
}

.header-title-box {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.header-titles {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
}

.drawer-title {
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--text-main);
  margin: 0;
}

.drawer-sub {
  font-size: 0.6875rem;
  color: var(--text-muted);
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 0.375rem;
}

.btn-refresh,
.btn-close {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 0.375rem;
  border-radius: 4px;
}

.btn-refresh:hover,
.btn-close:hover {
  background: var(--hover-bg);
  color: var(--text-main);
}

.drawer-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 1rem;
  gap: 0.5rem;
  overflow: hidden;
}

.editor-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.25rem 0.5rem;
  background: var(--section-bg);
  border: 1px solid var(--card-border);
  border-radius: 4px;
}

.file-path-badge {
  font-size: 0.75rem;
  color: var(--accent-cyan);
  font-weight: 600;
}

.dirty-badge {
  font-size: 0.6875rem;
  color: #d29922;
}

.toml-textarea {
  flex: 1;
  width: 100%;
  box-sizing: border-box;
  background: var(--input-bg);
  color: var(--text-main);
  border: 1px solid var(--card-border);
  border-radius: 6px;
  padding: 0.75rem;
  font-size: 0.8125rem;
  line-height: 1.5;
  resize: none;
  outline: none;
  white-space: pre;
}

.toml-textarea:focus {
  border-color: var(--accent-cyan);
}

.error-banner {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.75rem;
  background: rgba(248, 81, 73, 0.1);
  border: 1px solid rgba(248, 81, 73, 0.3);
  border-radius: 4px;
  font-size: 0.75rem;
  color: #f85149;
}

.drawer-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.875rem 1.25rem;
  background: var(--card-bg);
  border-top: 1px solid var(--card-border);
}

.footer-hint {
  font-size: 0.6875rem;
  color: var(--text-muted);
}

.footer-right {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.btn-save {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.4375rem 1rem;
  font-size: 0.8125rem;
  font-weight: 600;
  background: var(--accent-cyan, #0284c7);
  color: #ffffff;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.btn-save:hover:not(:disabled) {
  opacity: 0.9;
}

.btn-save:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-cancel {
  padding: 0.4375rem 0.875rem;
  font-size: 0.8125rem;
  background: transparent;
  color: var(--text-muted);
  border: 1px solid var(--card-border);
  border-radius: 4px;
  cursor: pointer;
}

.btn-cancel:hover {
  background: var(--hover-bg);
  color: var(--text-main);
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
