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
  background-color: var(--modal-bg, #252526);
  border: 1px solid var(--modal-border, #454545);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.5);
  border-radius: 6px;
  width: 90%;
  max-width: 420px;
  padding: 18px 20px;
  color: var(--text-main, #cccccc);
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
  color: var(--text-heading, #ffffff);
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
  color: var(--text-heading, #ffffff);
  word-break: break-all;
}

.modal-subtext {
  margin: 0;
  font-size: 12px;
  color: var(--text-muted, #858585);
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.btn {
  padding: 6px 14px;
  border-radius: 3px;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  border: 1px solid transparent;
  transition: all 0.15s ease;
}

.btn-secondary {
  background-color: transparent;
  border-color: var(--btn-secondary-border, #454545);
  color: var(--text-main, #cccccc);
}

.btn-secondary:hover {
  background-color: rgba(255, 255, 255, 0.08);
  color: #ffffff;
}

.btn-danger {
  background-color: var(--error, #e51400);
  color: #ffffff;
}

.btn-danger:hover {
  background-color: #d01200;
}
</style>
