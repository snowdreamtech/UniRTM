<template>
  <div class="view-container explorer-view" id="view-explorer">
    <div class="explorer-placeholder">
      <TechIcon id="workbench:explorer" :size="48" :colored="false" />
      <h3 class="placeholder-title">{{ t("explorer.title") || "资源管理器" }}</h3>
      <p class="placeholder-desc">{{ t("explorer.empty_folder") || "请在左侧资源管理器中选择文件或新建文件。" }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { t } from "../i18n";
import TechIcon from "../components/common/TechIcon.vue";
</script>

<style scoped>
.explorer-view {
  display: flex;
  flex-direction: column;
  height: 100%;
  width: 100%;
  padding: 1.5rem;
  box-sizing: border-box;
}

.explorer-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: var(--text-muted);
  gap: 0.75rem;
}

.placeholder-title {
  margin: 0;
  font-size: 1.2rem;
  font-weight: 600;
  color: var(--text-main);
}

.placeholder-desc {
  margin: 0;
  font-size: 0.85rem;
}
</style>
