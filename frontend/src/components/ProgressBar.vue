<template>
  <div class="progress-wrapper">
    <div class="progress-label">
      <span>{{ label }}</span>
      <span>{{ progress }}%</span>
    </div>
    <div class="progress-track">
      <div class="progress-fill" :style="{ width: progress + '%' }"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  label: string;
  progress: number;
}>();
</script>

<style scoped>
.progress-wrapper {
  margin: 1rem 0;
}

.progress-label {
  display: flex;
  justify-content: space-between;
  font-size: 0.85rem;
  color: var(--text-muted);
  margin-bottom: 0.4rem;
}

.progress-track {
  width: 100%;
  height: 8px;
  background: rgba(255, 255, 255, 0.08);
  border-radius: 999px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: linear-gradient(90deg, #00e5ff, #9d4edd);
  border-radius: 999px;
  transition: width 0.3s ease;
  box-shadow: 0 0 10px rgba(0, 229, 255, 0.5);
}
</style>
