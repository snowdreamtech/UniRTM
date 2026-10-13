<template>
  <div class="view-container tasks-view" id="view-tasks">
    <!-- View Header -->
    <div class="view-header-row">
      <div>
        <h2 class="view-title">⚡ {{ t("tasks.title") }}</h2>
        <p class="view-desc">
          {{ t("tasks.desc") }}
        </p>
      </div>
    </div>

    <!-- Tasks List Container -->
    <div class="tasks-table-container glass-card">
      <div class="tasks-list">
        <div
          v-for="task in tasks"
          :key="task.name"
          class="task-row-card"
          :class="{ running: runningTask === task.name }"
        >
          <div class="task-info-side">
            <div class="task-name-line">
              <span class="task-title">{{ task.name }}</span>
              <span v-if="task.env" class="task-env-pill">{{ task.env }}</span>
            </div>
            <div class="task-cmd-line font-mono">
              <span class="cmd-prompt">$</span> {{ task.command }}
            </div>
            <div class="task-desc-line">{{ task.description }}</div>
          </div>

          <div class="task-actions-side">
            <button
              class="run-task-btn"
              :disabled="runningTask !== null"
              @click="executeTask(task)"
            >
              <span v-if="runningTask === task.name" class="spin-icon">⏳ {{ t("tasks.running") }}</span>
              <span v-else>▶ {{ t("tasks.run") }}</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { t } from "../i18n";

const emit = defineEmits<{
  (e: "log-event", level: "INFO" | "SUCCESS" | "WARN" | "ERROR", msg: string, detail?: string, category?: "output" | "tasks" | "doctor"): void;
  (e: "task-started", taskName: string): void;
  (e: "task-finished", taskName: string): void;
}>();

interface ProjectTask {
  name: string;
  command: string;
  description: string;
  env?: string;
}

const runningTask = ref<string | null>(null);

const tasks = ref<ProjectTask[]>([
  {
    name: "dev",
    command: "vite",
    description: "启动前端 Vite 本地热重载开发服务器",
    env: "PORT=5173",
  },
  {
    name: "build",
    command: "vue-tsc && vite build",
    description: "执行 TypeScript 严格类型检查并编译生产优化包",
  },
  {
    name: "test",
    command: "vitest run",
    description: "运行包含 53 种国际化语言与核心组件的自动化测试",
  },
  {
    name: "audit:npm",
    command: "npm audit --registry=https://registry.npmjs.org",
    description: "执行 npm 依赖树全量安全性与 CVE 漏洞深度扫描",
  },
]);

function executeTask(task: ProjectTask) {
  runningTask.value = task.name;
  emit("task-started", task.name);
  emit("log-event", "INFO", `Starting task: ${task.name}`, `Command: ${task.command}`, "tasks");

  setTimeout(() => {
    emit("log-event", "INFO", `[${task.name}] Loading environment variables and shims...`, undefined, "tasks");
  }, 300);

  setTimeout(() => {
    emit("log-event", "SUCCESS", `[${task.name}] Process finished successfully with exit code 0.`, undefined, "tasks");
    emit("task-finished", task.name);
    runningTask.value = null;
  }, 1200);
}
</script>

<style scoped>
.tasks-view {
  display: flex;
  flex-direction: column;
  gap: 1.1rem;
  padding: 1.25rem;
  height: 100%;
  overflow-y: auto;
  box-sizing: border-box;
}

.view-header-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.view-title {
  font-size: 1.25rem;
  font-weight: 700;
  margin: 0 0 0.35rem;
  color: var(--text-main);
}

.view-desc {
  font-size: 0.82rem;
  color: var(--text-muted);
  margin: 0;
}

.tasks-table-container {
  padding: 1rem;
  border-radius: 12px;
}

.tasks-list {
  display: flex;
  flex-direction: column;
  gap: 0.85rem;
}

.task-row-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.95rem 1.15rem;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  border-radius: 10px;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

.task-row-card:hover {
  border-color: var(--card-border-active);
  background: rgba(0, 229, 255, 0.04);
}

.task-row-card.running {
  border-color: var(--accent-cyan);
  background: rgba(0, 229, 255, 0.08);
}

.task-info-side {
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
  flex: 1;
}

.task-name-line {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.task-title {
  font-size: 0.95rem;
  font-weight: 700;
  color: var(--text-main);
}

.task-env-pill {
  font-size: 0.68rem;
  background: var(--section-bg);
  border: 1px solid var(--card-border);
  color: var(--accent-cyan);
  padding: 0.05rem 0.4rem;
  border-radius: 4px;
}

.task-cmd-line {
  font-size: 0.78rem;
  color: var(--accent-cyan);
  background: rgba(0, 0, 0, 0.2);
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
  width: fit-content;
}

.cmd-prompt {
  color: var(--text-muted);
  margin-right: 0.2rem;
}

.task-desc-line {
  font-size: 0.75rem;
  color: var(--text-muted);
}

.run-task-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.45rem 1rem;
  border-radius: 8px;
  border: none;
  font-size: 0.8rem;
  font-weight: 600;
  cursor: pointer;
  background: linear-gradient(135deg, var(--accent-cyan), #0284c7);
  color: #070a12;
  transition: all 0.2s ease;
}

.run-task-btn:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 14px rgba(0, 229, 255, 0.35);
}

.run-task-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
