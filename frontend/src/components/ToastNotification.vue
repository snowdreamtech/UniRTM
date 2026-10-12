<template>
  <transition name="toast-fade">
    <div v-if="message" class="global-toast" :class="type">
      <span class="toast-icon">
        <template v-if="type === 'warning'">⚠️</template>
        <template v-else-if="type === 'error'">❌</template>
        <template v-else-if="type === 'success'">🎉</template>
        <template v-else>ℹ️</template>
      </span>
      <span class="toast-text">{{ message }}</span>
      <button class="toast-close" title="Close" @click="$emit('close')">✕</button>
    </div>
  </transition>
</template>

<script setup lang="ts">
defineProps<{
  message: string;
  type?: "info" | "success" | "warning" | "error";
}>();

defineEmits<{
  (e: "close"): void;
}>();
</script>

<style scoped>
.global-toast {
  position: fixed;
  top: 1.5rem;
  right: 2rem;
  z-index: 2000;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 1.25rem;
  border-radius: 10px;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.4);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  font-size: 0.875rem;
  font-weight: 500;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.global-toast.info {
  background: rgba(15, 23, 42, 0.95);
  border-color: rgba(0, 229, 255, 0.4);
  color: #f8fafc;
}

.global-toast.success {
  background: rgba(6, 78, 59, 0.95);
  border-color: rgba(16, 185, 129, 0.4);
  color: #ecfdf5;
}

.global-toast.warning {
  background: rgba(120, 53, 15, 0.95);
  border-color: rgba(245, 158, 11, 0.4);
  color: #fffbeb;
}

.global-toast.error {
  background: rgba(127, 29, 29, 0.95);
  border-color: rgba(239, 68, 68, 0.4);
  color: #fef2f2;
}

.toast-close {
  background: none;
  border: none;
  color: inherit;
  opacity: 0.7;
  cursor: pointer;
  padding: 0.1rem 0.3rem;
  font-size: 0.9rem;
}

.toast-close:hover {
  opacity: 1;
}

.toast-fade-enter-active,
.toast-fade-leave-active {
  transition: all 0.2s ease;
}

.toast-fade-enter-from,
.toast-fade-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}
</style>
