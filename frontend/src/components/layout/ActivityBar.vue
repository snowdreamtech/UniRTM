<template>
  <aside class="activity-bar" id="app-activity-bar" role="navigation" aria-label="Activity Bar">
    <!-- Top Action Icons -->
    <div class="activity-top">
      <button
        v-for="item in primaryTabs"
        :key="item.id"
        class="activity-btn"
        :id="'btn-activity-' + item.id"
        :class="{ active: currentTab === item.id }"
        :title="item.label"
        @click="$emit('select-tab', item.id)"
      >
        <span class="activity-indicator" v-if="currentTab === item.id"></span>
        <span class="activity-icon">{{ item.icon }}</span>
        <span v-if="item.badge" class="activity-badge">{{ item.badge }}</span>
      </button>
    </div>

    <!-- Bottom Utility Icons -->
    <div class="activity-bottom">
      <button
        class="activity-btn"
        id="btn-activity-settings"
        :title="t('settings.title') || 'Settings'"
        @click="$emit('open-settings')"
      >
        <span class="activity-icon">⚙️</span>
      </button>

      <button
        class="activity-btn"
        id="btn-activity-about"
        :title="t('about.title') || 'About'"
        @click="$emit('open-about')"
      >
        <span class="activity-icon">ℹ️</span>
      </button>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { t } from "../../i18n";

export interface ActivityTab {
  id: string;
  icon: string;
  label: string;
  badge?: string | number;
}

const props = defineProps<{
  currentTab: string;
  toolCount?: number;
  taskCount?: number;
}>();

defineEmits<{
  (e: "select-tab", tabId: string): void;
  (e: "open-settings"): void;
  (e: "open-about"): void;
}>();

const primaryTabs = computed<ActivityTab[]>(() => [
  {
    id: "dashboard",
    icon: "📊",
    label: t("nav.dashboard") || "Dashboard (项目总览)",
  },
  {
    id: "tools",
    icon: "🛠️",
    label: t("nav.tools") || "Toolchains (工具链)",
    badge: props.toolCount || 5,
  },
  {
    id: "tasks",
    icon: "⚡",
    label: t("nav.tasks") || "Task Runner (任务系统)",
    badge: props.taskCount || 4,
  },
  {
    id: "marketplace",
    icon: "🧩",
    label: t("nav.marketplace") || "Extensions Marketplace (扩展市场)",
  },
  {
    id: "doctor",
    icon: "🩺",
    label: t("nav.doctor") || "Doctor Diagnostics (健康诊断)",
  },
]);
</script>

<style scoped>
.activity-bar {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  width: 50px;
  min-width: 50px;
  height: 100%;
  background: var(--card-bg);
  border-right: 1px solid var(--card-border);
  box-sizing: border-box;
  padding: 0.5rem 0;
  user-select: none;
  -webkit-user-select: none;
  z-index: 50;
  backdrop-filter: blur(8px);
}

.activity-top,
.activity-bottom {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.35rem;
}

.activity-btn {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  background: transparent;
  border: none;
  border-radius: 8px;
  color: var(--text-muted);
  cursor: pointer;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  padding: 0;
}

.activity-btn:hover {
  background: var(--input-bg);
  color: var(--text-main);
  transform: translateY(-1px);
}

.activity-btn.active {
  color: var(--accent-cyan);
  background: rgba(0, 229, 255, 0.12);
}

.activity-indicator {
  position: absolute;
  left: -5px;
  top: 8px;
  bottom: 8px;
  width: 3px;
  background: var(--accent-cyan);
  border-radius: 0 4px 4px 0;
  box-shadow: 0 0 8px var(--accent-cyan);
}

.activity-icon {
  font-size: 1.25rem;
  line-height: 1;
}

.activity-badge {
  position: absolute;
  top: 3px;
  right: 2px;
  background: var(--accent-purple);
  color: #ffffff;
  font-size: 0.62rem;
  font-weight: 700;
  padding: 0.05rem 0.3rem;
  border-radius: 9999px;
  line-height: 1.1;
  box-shadow: 0 2px 5px rgba(0, 0, 0, 0.3);
}
</style>
