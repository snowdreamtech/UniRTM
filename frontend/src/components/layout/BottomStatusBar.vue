<template>
  <footer class="status-bar" id="app-status-bar" role="contentinfo">
    <!-- Left Section: Overall Health & Diagnostics -->
    <div class="status-left">
      <div
        class="status-pill status-ready"
        id="status-health"
        :title="'System & Environment Status: Ready'"
        @click="$emit('select-tab', 'doctor')"
      >
        <span class="status-dot">●</span>
        <span class="status-text">{{ statusReadyText }}</span>
      </div>

      <div class="status-pill status-warnings" id="status-warnings" :title="'0 Security Warnings'">
        <span class="status-icon"><TechIcon id="workbench:warning" :size="14" /></span>
        <span class="status-text">0</span>
      </div>
    </div>

    <!-- Center Section: Active Background Task / Dynamic Notification -->
    <div class="status-center">
      <div v-if="activeTaskName" class="status-task-pill" id="status-running-task" @click="$emit('toggle-drawer')">
        <TechIcon id="workbench:sync" :size="13" class="spin-icon" />
        <span class="task-running-text">{{ runningTaskText }}: {{ activeTaskName }}</span>
      </div>
    </div>

    <!-- Right Section: Scope, Tool Count, Mirror Latency, Lockfile, Version -->
    <div class="status-right">
      <!-- Active Scope Switcher -->
      <button
        class="status-btn"
        id="btn-status-scope"
        :title="
          activeScope === 'user'
            ? '当前作用域: 用户全局 (~/.unirtm) • 点击切换为工作区 (Alt+S)'
            : '当前作用域: 工作区 (.unirtm.toml) • 点击切换为用户全局 (Alt+S)'
        "
        @click="toggleScope(router)"
      >
        <span class="btn-icon">
          <TechIcon
            :id="activeScope === 'user' ? 'workbench:settings' : 'workbench:folder'"
            :size="13"
            :colored="false"
          />
        </span>
        <span>{{ activeScope === "user" ? (isZh ? "用户全局" : "User") : isZh ? "工作区" : "Workspace" }}</span>
      </button>

      <!-- Active Toolchain Count -->
      <button
        class="status-btn"
        id="btn-status-tools"
        :title="
          activeScope === 'user'
            ? isZh
              ? '用户全局共 1 个工具'
              : '1 tool configured in user global'
            : isZh
              ? '工作区共 5 个工具'
              : '5 tools configured in workspace'
        "
        @click="$emit('select-tab', 'tools')"
      >
        <span class="btn-icon"><TechIcon id="workbench:tools" :size="14" :colored="false" /></span>
        <span>{{ activeScope === "user" ? 1 : 5 }} {{ toolsLabelText }}</span>
      </button>

      <!-- Network Mirror & Latency -->
      <button
        class="status-btn"
        id="btn-status-network"
        :title="'Mirror: npmmirror • Latency: ' + (latencyMs ? latencyMs + 'ms' : 'Ready')"
        @click="$emit('open-network-settings')"
      >
        <span class="btn-icon"><TechIcon id="workbench:runtimes" :size="14" :colored="false" /></span>
        <span>{{ mirrorLabel }}</span>
      </button>

      <!-- Lockfile status -->
      <div class="status-pill" id="status-lockfile" :title="'unirtm.lock is in sync with active dependencies'">
        <span class="btn-icon"><TechIcon id="unirtm.lock" :size="12" /></span>
        <span>{{ lockStatusText }}</span>
      </div>

      <!-- Core Version -->
      <div class="status-version font-mono" title="UniRTM Core Version">v0.35.0</div>
    </div>
  </footer>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useRouter } from "vue-router";
import TechIcon from "../common/TechIcon.vue";
import { useAppScope } from "../../composables/useAppScope";

const router = useRouter();
const { activeScope, toggleScope } = useAppScope();

const props = defineProps<{
  activeTaskName?: string;
  latencyMs?: number | null;
  currentLang?: string;
}>();

defineEmits<{
  (e: "select-tab", tabId: string): void;
  (e: "open-network-settings"): void;
  (e: "toggle-drawer"): void;
}>();

const isZh = computed(() => {
  return props.currentLang ? props.currentLang.startsWith("zh") : true;
});

const statusReadyText = computed(() => (isZh.value ? "环境就绪" : "Ready"));
const runningTaskText = computed(() => (isZh.value ? "正在执行" : "Running"));
const toolsLabelText = computed(() => (isZh.value ? "个工具" : "tools"));
const lockStatusText = computed(() => (isZh.value ? "锁定已同步" : "Lock Sync"));

const mirrorLabel = computed(() => {
  const lat = props.latencyMs ? ` (${props.latencyMs}ms)` : "";
  return `npmmirror${lat}`;
});
</script>

<style scoped>
.status-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 26px;
  min-height: 26px;
  background: var(--card-bg);
  border-top: 1px solid var(--card-border);
  padding: 0 0.65rem;
  font-size: 0.72rem;
  color: var(--text-muted);
  user-select: none;
  -webkit-user-select: none;
  box-sizing: border-box;
  z-index: 60;
  backdrop-filter: blur(8px);
}

.status-left,
.status-right {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.status-center {
  flex: 1;
  display: flex;
  justify-content: center;
}

.status-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.1rem 0.35rem;
  border-radius: 4px;
}

.status-ready {
  color: var(--success);
  cursor: pointer;
  font-weight: 500;
}

.status-ready:hover {
  background: rgba(16, 185, 129, 0.12);
}

.status-dot {
  font-size: 0.75rem;
  line-height: 1;
}

.status-warnings {
  color: var(--text-muted);
}

.status-task-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  background: rgba(0, 229, 255, 0.14);
  color: var(--accent-cyan);
  padding: 0.1rem 0.55rem;
  border-radius: 9999px;
  font-weight: 600;
  cursor: pointer;
  animation: pulse-task 2s infinite ease-in-out;
}

@keyframes pulse-task {
  0%,
  100% {
    opacity: 0.9;
  }
  50% {
    opacity: 0.6;
  }
}

.spin-icon {
  display: inline-block;
  animation: spin 1.2s infinite linear;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.status-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  background: transparent;
  border: none;
  color: var(--text-muted);
  font-size: 0.72rem;
  padding: 0.1rem 0.35rem;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.status-btn:hover {
  background: var(--input-bg);
  color: var(--text-main);
}

.btn-icon {
  font-size: 0.75rem;
}

.status-version {
  font-size: 0.68rem;
  opacity: 0.7;
  padding-left: 0.25rem;
}
</style>
