<!-- Copyright (c) 2026 SnowdreamTech. All rights reserved. -->
<!-- Licensed under the MIT License. See LICENSE file in the project root for full license information. -->

<template>
  <div v-if="addFolderModal.visible" class="modal-backdrop" @click.self="closeAddFolderModal">
    <div class="modal-dialog glass-card add-folder-modal" role="dialog" aria-modal="true">
      <div class="modal-header">
        <div class="modal-title-row">
          <span class="modal-icon">
            <span class="codicon codicon-folder-library" style="font-size: 20px; color: var(--accent-cyan, #06b6d4);"></span>
          </span>
          <div>
            <h3 class="modal-title">{{ t('explorer.add_folder_title') || '将文件夹添加到工作区' }}</h3>
            <p class="modal-desc">{{ t('explorer.add_folder_desc') || '输入本地文件夹相对路径或绝对路径，将其作为顶级根挂载至当前工作区' }}</p>
          </div>
        </div>
        <button class="btn btn-icon modal-close-btn" @click="closeAddFolderModal" :title="t('common.close') || '关闭'">
          <span class="codicon codicon-close"></span>
        </button>
      </div>

      <div class="modal-body">
        <div class="form-group">
          <label class="form-label">{{ t('explorer.folder_path_label') || '文件夹路径' }}</label>
          <input
            ref="inputRef"
            v-model="inputPath"
            type="text"
            class="modal-text-input font-mono"
            :placeholder="t('explorer.folder_path_placeholder') || '例如: frontend 或 /path/to/folder'"
            @keydown.enter="handleSubmit"
            @keydown.esc="closeAddFolderModal"
          />
        </div>

        <div v-if="suggestedFolders.length > 0" class="suggestions-section">
          <span class="suggestions-label">{{ t('explorer.quick_suggestions') || '当前可用子文件夹快捷添加：' }}</span>
          <div class="suggestion-chips">
            <button
              v-for="folder in suggestedFolders"
              :key="folder.path"
              type="button"
              class="suggestion-chip font-mono"
              :class="{ selected: inputPath === folder.path }"
              @click="inputPath = folder.path"
            >
              <span class="codicon codicon-folder"></span>
              <span>{{ folder.name }}</span>
            </button>
          </div>
        </div>
      </div>

      <div class="modal-footer">
        <button type="button" class="btn btn-secondary" @click="closeAddFolderModal">
          {{ t('common.cancel') || '取消' }}
        </button>
        <button
          type="button"
          class="btn btn-primary"
          :disabled="!inputPath.trim() || submitting"
          @click="handleSubmit"
        >
          <span v-if="submitting">{{ t('common.saving') || '添加中...' }}</span>
          <span v-else>{{ t('explorer.confirm_add') || '添加到工作区' }}</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from "vue";
import { useFileTree } from "../../composables/useFileTree";
import { t } from "../../i18n";

const {
  addFolderModal,
  closeAddFolderModal,
  addFolderToWorkspace,
  tree,
  workspaceRoots,
} = useFileTree();

const inputRef = ref<HTMLInputElement | null>(null);
const inputPath = ref("");
const submitting = ref(false);

watch(
  () => addFolderModal.value.visible,
  (val) => {
    if (val) {
      inputPath.value = addFolderModal.value.defaultPath || "";
      nextTick(() => {
        inputRef.value?.focus();
        inputRef.value?.select();
      });
    }
  }
);

const suggestedFolders = computed(() => {
  if (!tree.value?.children) return [];
  const existingPaths = new Set(workspaceRoots.value.map((r) => r.path));
  return tree.value.children
    .filter((c) => c.is_dir && !existingPaths.has(c.path))
    .slice(0, 8);
});

async function handleSubmit() {
  const p = inputPath.value.trim();
  if (!p || submitting.value) return;
  submitting.value = true;
  try {
    const success = await addFolderToWorkspace(p);
    if (success) {
      closeAddFolderModal();
    }
  } finally {
    submitting.value = false;
  }
}
</script>

<style scoped>
.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.65);
  backdrop-filter: blur(6px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
}

.add-folder-modal {
  width: 90%;
  max-width: 520px;
  background: var(--card-bg, #18181b);
  border: 1px solid var(--card-border, rgba(255, 255, 255, 0.12));
  border-radius: 12px;
  box-shadow: 0 16px 40px rgba(0, 0, 0, 0.5);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.modal-header {
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid var(--border-color, rgba(255, 255, 255, 0.08));
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
}

.modal-title-row {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
}

.modal-title {
  margin: 0;
  font-size: 1.05rem;
  font-weight: 600;
  color: var(--text-main, #f1f5f9);
}

.modal-desc {
  margin: 0.25rem 0 0 0;
  font-size: 0.8rem;
  color: var(--text-muted, #94a3b8);
  line-height: 1.4;
}

.modal-body {
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.form-label {
  font-size: 0.82rem;
  font-weight: 600;
  color: var(--text-main, #e2e8f0);
}

.modal-text-input {
  width: 100%;
  padding: 0.6rem 0.8rem;
  font-size: 0.85rem;
  border-radius: 6px;
  background: var(--input-bg, rgba(255, 255, 255, 0.04));
  border: 1px solid var(--card-border, rgba(255, 255, 255, 0.15));
  color: var(--text-main, #f8fafc);
  outline: none;
  box-sizing: border-box;
  transition: all 0.2s ease;
}

.modal-text-input:focus {
  border-color: var(--accent-cyan, #06b6d4);
  box-shadow: 0 0 0 2px rgba(6, 182, 212, 0.2);
}

.suggestions-section {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.suggestions-label {
  font-size: 0.75rem;
  color: var(--text-muted, #94a3b8);
}

.suggestion-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.suggestion-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.25rem 0.6rem;
  font-size: 0.76rem;
  border-radius: 4px;
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: var(--text-main, #e2e8f0);
  cursor: pointer;
  transition: all 0.15s ease;
}

.suggestion-chip:hover,
.suggestion-chip.selected {
  background: rgba(6, 182, 212, 0.15);
  border-color: var(--accent-cyan, #06b6d4);
  color: var(--accent-cyan, #06b6d4);
}

.modal-footer {
  padding: 1rem 1.5rem;
  border-top: 1px solid var(--border-color, rgba(255, 255, 255, 0.08));
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
}

.btn {
  padding: 0.45rem 1rem;
  font-size: 0.82rem;
  font-weight: 500;
  border-radius: 6px;
  cursor: pointer;
  border: none;
  transition: all 0.15s ease;
}

.btn-secondary {
  background: transparent;
  color: var(--text-muted, #94a3b8);
  border: 1px solid var(--border-color, rgba(255, 255, 255, 0.15));
}

.btn-secondary:hover {
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-main, #f1f5f9);
}

.btn-primary {
  background: linear-gradient(135deg, var(--accent-cyan, #06b6d4), #0284c7);
  color: #fff;
}

.btn-primary:hover:not(:disabled) {
  opacity: 0.9;
  transform: translateY(-1px);
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
