<template>
  <div v-if="deleteConfirmModal.visible" class="modal-backdrop" @click="handleCancel">
    <div class="modal-dialog" @click.stop>
      <div class="modal-header">
        <span class="modal-icon codicon codicon-warning"></span>
        <h3 class="modal-title">确认删除</h3>
      </div>
      <div class="modal-body">
        <p class="modal-message">
          确定要彻底删除
          <strong class="target-name">{{ deleteConfirmModal.node?.name }}</strong>
          吗？
        </p>
        <p v-if="deleteConfirmModal.node?.is_dir" class="modal-subtext">
          该文件夹内的所有子文件和目录也将被一并永久删除，此操作不可逆。
        </p>
        <p v-else class="modal-subtext">
          此文件将被永久删除，此操作不可逆。
        </p>
      </div>
      <div class="modal-footer">
        <button class="btn btn-secondary" @click="handleCancel">取消</button>
        <button class="btn btn-danger" @click="handleConfirm">删除</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted } from "vue";
import { useFileTree } from "../../composables/useFileTree";

const { deleteConfirmModal, deleteItem } = useFileTree();

function handleCancel() {
  deleteConfirmModal.value = { visible: false, node: null };
}

async function handleConfirm() {
  const node = deleteConfirmModal.value.node;
  if (node) {
    await deleteItem(node.path);
  }
}

function handleKeydown(e: KeyboardEvent) {
  if (!deleteConfirmModal.value.visible) return;
  if (e.key === "Escape") {
    handleCancel();
  } else if (e.key === "Enter") {
    handleConfirm();
  }
}

onMounted(() => {
  window.addEventListener("keydown", handleKeydown);
});

onUnmounted(() => {
  window.removeEventListener("keydown", handleKeydown);
});
</script>

<style scoped>
.modal-backdrop {
  position: fixed;
  inset: 0;
  background-color: rgba(0, 0, 0, 0.65);
  backdrop-filter: blur(2px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10000;
}

.modal-dialog {
  background-color: var(--modal-bg, var(--card-bg));
  border: 1px solid var(--card-border);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);
  border-radius: 8px;
  width: 90%;
  max-width: 420px;
  padding: 20px 24px;
  color: var(--text-main);
  animation: modal-pop 0.15s cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes modal-pop {
  from {
    opacity: 0;
    transform: scale(0.96);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}

.modal-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
}

.modal-icon {
  font-size: 20px;
  color: var(--warning, #f59e0b);
}

.modal-title {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  color: var(--text-main);
}

.modal-body {
  font-size: 13px;
  line-height: 1.5;
  margin-bottom: 20px;
}

.modal-message {
  margin: 0 0 6px;
}

.target-name {
  color: var(--accent-cyan);
  word-break: break-all;
}

.modal-subtext {
  margin: 0;
  font-size: 12px;
  color: var(--text-muted);
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.btn {
  padding: 6px 14px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  border: 1px solid transparent;
  transition: all 0.15s ease;
}

.btn-secondary {
  background-color: transparent;
  border-color: var(--card-border);
  color: var(--text-main);
}

.btn-secondary:hover {
  background-color: rgba(128, 128, 128, 0.15);
}

.btn-danger {
  background-color: var(--error, #e51400);
  color: #ffffff;
}

.btn-danger:hover {
  filter: brightness(1.1);
}
</style>
