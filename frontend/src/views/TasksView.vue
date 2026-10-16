<template>
  <div class="view-container tasks-view" id="view-tasks">
    <!-- View Header -->
    <div class="view-header-row">
      <div>
        <h2 class="view-title" style="display: flex; align-items: center; gap: 8px;">
          <TechIcon id="workbench:tasks" :size="20" />
          <span>{{ t("tasks.title") || "任务流与工程自动化" }}</span>
        </h2>
        <p class="view-desc">
          {{ t("tasks.desc") || "由 UniRTM 统一编排运行工程构建、依赖安全审计、代码格式化与自定义脚本，保障跨平台执行一致性。" }}
        </p>
      </div>
      <div class="view-actions" style="display: flex; gap: 8px; align-items: center;">
        <button class="tool-action-btn secondary" @click="fetchTasks" :disabled="loading" title="刷新任务列表">
          <TechIcon id="workbench:sync" :size="14" :class="{ 'spin-icon': loading }" />
          <span>{{ loading ? t("common.refreshing") || "刷新中..." : t("common.refresh") || "刷新" }}</span>
        </button>
        <button class="tool-action-btn primary" @click="openCreateModal" title="新建工程自动化任务">
          <TechIcon id="workbench:edit" :size="14" />
          <span>{{ t("tasks.create_btn") || "新建任务" }}</span>
        </button>
      </div>
    </div>

    <!-- Current Scope Context Banner -->
    <div class="scope-selection-bar glass-card" id="tasks-scope-bar">
      <div class="scope-context-pill">
        <span class="scope-badge" :class="activeScope">
          <TechIcon :id="activeScope === 'user' ? 'workbench:settings' : 'workbench:folder'" :size="13" :colored="false" />
          {{ activeScope === 'user' ? (t("runtimes.scope_user") || "用户全局任务") : (t("runtimes.scope_workspace") || "工作区项目任务") }}
        </span>
        <span class="scope-hint-text">
          {{ activeScope === 'user'
              ? (t("runtimes.scope_user_desc") || "展示当前用户全局配置文件 (~/.config/unirtm/unirtm.toml) 定义的任务")
              : (t("runtimes.scope_workspace_desc") || "解析当前工作区项目配置文件 (.unirtm.toml) 定义的自动化任务") }}
        </span>
      </div>
      <div class="scope-context-config font-mono" :title="activeScope === 'user' ? '~/.config/unirtm/unirtm.toml' : '.unirtm.toml'">
        <span class="file-icon">📄</span>
        <span>{{ activeScope === 'user' ? '~/.config/unirtm/unirtm.toml' : '.unirtm.toml' }}</span>
      </div>
    </div>

    <!-- Filter & Search Toolbar -->
    <div class="tasks-filter-bar glass-card">
      <div class="search-input-wrapper">
        <span class="search-icon">
          <TechIcon id="workbench:doctor" :size="14" :colored="false" />
        </span>
        <input
          v-model="searchKey"
          type="text"
          class="task-search-input"
          placeholder="搜索任务名称、依赖、命令或描述..."
        />
      </div>
      <div class="task-stats-badges">
        <span class="stat-badge">
          <b>{{ filteredTasks.length }}</b> 个任务
        </span>
      </div>
    </div>

    <!-- Tasks List Container -->
    <div class="tasks-table-container glass-card">
      <!-- Loading State -->
      <div v-if="loading && tasks.length === 0" class="tasks-empty-box">
        <TechIcon id="workbench:sync" :size="24" class="spin-icon" />
        <span style="font-size: 0.85rem; color: var(--text-muted);">正在解析配置文件中的任务...</span>
      </div>

      <!-- Empty State -->
      <div v-else-if="filteredTasks.length === 0" class="tasks-empty-box">
        <TechIcon id="workbench:tasks" :size="32" :colored="false" />
        <div class="empty-title">
          {{ searchKey ? `未找到匹配 "${searchKey}" 的任务` : (activeScope === 'user' ? '用户全局环境暂未定义任务' : '当前工作区配置文件暂未定义 [tasks]') }}
        </div>
        <p class="empty-desc">
          {{ activeScope === 'user'
              ? '可在用户配置文件 (~/.config/unirtm/unirtm.toml) 中配置 [tasks.<name>] 定义通用任务。'
              : '可在项目根目录的 .unirtm.toml 中添加 [tasks.<name>] 定义构建、测试与安全扫描等自动化脚本。' }}
        </p>
        <button v-if="!searchKey" class="tool-action-btn primary" style="margin-top: 0.75rem;" @click="openCreateModal">
          <TechIcon id="workbench:edit" :size="14" />
          <span>{{ t("tasks.create_first") || "创建第一个任务" }}</span>
        </button>
      </div>

      <!-- Tasks List -->
      <div v-else class="tasks-list">
        <div
          v-for="task in filteredTasks"
          :key="task.name"
          class="task-row-card"
          :class="{ running: runningTask === task.name }"
        >
          <div class="task-info-side">
            <div class="task-name-line">
              <span class="task-title font-mono" style="display: inline-flex; align-items: center; gap: 6px;">
                <TechIcon id="workbench:tasks" :size="14" />
                {{ task.name }}
              </span>
              <span v-if="task.env" class="task-env-pill font-mono" :title="`环境变量: ${task.env}`">
                env: {{ task.env }}
              </span>
              <div v-if="task.depends && task.depends.length > 0" class="task-depends-wrap">
                <span class="depends-label">依赖:</span>
                <span
                  v-for="dep in task.depends"
                  :key="dep"
                  class="depends-pill font-mono"
                >
                  {{ dep }}
                </span>
              </div>
            </div>

            <div class="task-cmd-line font-mono" :title="task.command">
              <span class="cmd-prompt">$</span> {{ task.command }}
            </div>

            <div v-if="task.description" class="task-desc-line">
              {{ task.description }}
            </div>
          </div>

          <div class="task-actions-side">
            <button
              class="run-task-btn"
              :disabled="runningTask !== null"
              @click="executeTask(task)"
              :title="`运行任务: unirtm run ${task.name}`"
            >
              <span v-if="runningTask === task.name" style="display: inline-flex; align-items: center; gap: 6px;">
                <TechIcon id="workbench:sync" :size="12" class="spin-icon" />
                <span>{{ t("tasks.running") || "运行中..." }}</span>
              </span>
              <span v-else style="display: inline-flex; align-items: center; gap: 4px;">
                <TechIcon id="workbench:tasks" :size="12" :colored="false" />
                <span>{{ t("tasks.run") || "运行" }}</span>
              </span>
            </button>
            <button
              class="task-icon-btn edit"
              @click="openEditModal(task)"
              :title="`编辑任务: ${task.name}`"
            >
              <TechIcon id="workbench:edit" :size="13" />
            </button>
            <button
              class="task-icon-btn delete"
              @click="openDeleteConfirm(task)"
              :title="`删除任务: ${task.name}`"
            >
              <TechIcon id="workbench:delete" :size="13" />
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Task Editor Modal -->
    <TaskModal
      :is-open="isTaskModalOpen"
      :is-edit="isEditMode"
      :initial-task="editingTask"
      :scope="activeScope"
      :all-task-names="allTaskNames"
      :saving="savingTask"
      @close="isTaskModalOpen = false"
      @save="handleSaveTask"
    />

    <!-- Delete Confirmation Modal -->
    <div v-if="taskToDelete" class="modal-overlay" @click.self="taskToDelete = null">
      <div class="modal-card glass-modal delete-confirm-modal">
        <div class="modal-header">
          <div class="header-title" style="color: #ef4444;">
            <TechIcon id="workbench:warning" :size="20" />
            <h3 style="margin: 0; font-size: 1.05rem;">{{ t("tasks.delete_title") || "确认删除任务" }}</h3>
          </div>
          <button class="close-btn" @click="taskToDelete = null">✕</button>
        </div>
        <div class="modal-body" style="padding: 1.25rem 1.4rem;">
          <p style="margin: 0 0 0.75rem 0; font-size: 0.85rem; color: var(--text-main);">
            确定要删除任务 <b class="font-mono" style="color: var(--accent-cyan);">{{ taskToDelete.name }}</b> 吗？
          </p>
          <p style="margin: 0; font-size: 0.75rem; color: var(--text-muted); line-height: 1.4;">
            此操作将从 {{ activeScope === 'user' ? '~/.config/unirtm/unirtm.toml' : '.unirtm.toml' }} 中移除该任务配置，删除后不可撤销。
          </p>
          <div class="modal-footer" style="display: flex; justify-content: flex-end; gap: 0.75rem; margin-top: 1.25rem; padding-top: 1rem; border-top: 1px solid rgba(255, 255, 255, 0.08);">
            <button class="tool-action-btn secondary" @click="taskToDelete = null" :disabled="deletingTask">
              {{ t("common.cancel") || "取消" }}
            </button>
            <button class="tool-action-btn danger" @click="handleConfirmDelete" :disabled="deletingTask">
              <TechIcon v-if="deletingTask" id="workbench:sync" :size="14" class="spin-icon" />
              <span>{{ deletingTask ? (t("common.deleting") || "删除中...") : (t("common.delete") || "确认删除") }}</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from "vue";
import { t } from "../i18n";
import TechIcon from "../components/common/TechIcon.vue";
import TaskModal, { type TaskFormData } from "../components/TaskModal.vue";
import { useAppScope } from "../composables/useAppScope";
import { useToast } from "../composables/useToast";

const props = defineProps<{
  activeSubItem?: string;
  initialScope?: string;
}>();

const emit = defineEmits<{
  (e: "log-event", level: "INFO" | "SUCCESS" | "WARN" | "ERROR", msg: string, detail?: string, category?: "output" | "tasks" | "doctor"): void;
  (e: "task-started", taskName: string): void;
  (e: "task-finished", taskName: string): void;
}>();

const { activeScope } = useAppScope();
const { showToast } = useToast();

export interface ProjectTask {
  name: string;
  command: string;
  description: string;
  depends?: string[];
  env?: string;
  output?: string;
  timeout?: number;
}

const tasks = ref<ProjectTask[]>([]);
const loading = ref(false);
const runningTask = ref<string | null>(null);
const searchKey = ref("");

// Task modal state
const isTaskModalOpen = ref(false);
const isEditMode = ref(false);
const editingTask = ref<ProjectTask | null>(null);
const savingTask = ref(false);

// Delete confirm modal state
const taskToDelete = ref<ProjectTask | null>(null);
const deletingTask = ref(false);

const allTaskNames = computed(() => tasks.value.map((t) => t.name));

function openCreateModal() {
  editingTask.value = null;
  isEditMode.value = false;
  isTaskModalOpen.value = true;
}

function openEditModal(task: ProjectTask) {
  editingTask.value = task;
  isEditMode.value = true;
  isTaskModalOpen.value = true;
}

function openDeleteConfirm(task: ProjectTask) {
  taskToDelete.value = task;
}

async function handleSaveTask(formData: TaskFormData) {
  savingTask.value = true;
  try {
    const res = await fetch("/api/tasks", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        ...formData,
        scope: activeScope.value,
      }),
    });
    const data = await res.json();
    if (res.ok && data.success) {
      showToast(isEditMode.value ? `任务 ${formData.name} 已更新` : `任务 ${formData.name} 创建成功`, "success");
      emit("log-event", "SUCCESS", `[任务管理] 任务 ${formData.name} 保存成功`, undefined, "tasks");
      isTaskModalOpen.value = false;
      await fetchTasks();
    } else {
      showToast(data.error || "保存任务失败", "error");
      emit("log-event", "ERROR", `[任务管理] 任务 ${formData.name} 保存失败: ${data.error}`, undefined, "tasks");
    }
  } catch (err: any) {
    showToast(`保存异常: ${err.message}`, "error");
  } finally {
    savingTask.value = false;
  }
}

async function handleConfirmDelete() {
  if (!taskToDelete.value) return;
  deletingTask.value = true;
  const name = taskToDelete.value.name;
  try {
    const res = await fetch(`/api/tasks?name=${encodeURIComponent(name)}&scope=${activeScope.value}`, {
      method: "DELETE",
    });
    const data = await res.json();
    if (res.ok && data.success) {
      showToast(`任务 ${name} 已从配置中删除`, "success");
      emit("log-event", "SUCCESS", `[任务管理] 任务 ${name} 已删除`, undefined, "tasks");
      taskToDelete.value = null;
      await fetchTasks();
    } else {
      showToast(data.error || "删除任务失败", "error");
    }
  } catch (err: any) {
    showToast(`删除异常: ${err.message}`, "error");
  } finally {
    deletingTask.value = false;
  }
}

const filteredTasks = computed(() => {
  const q = searchKey.value.trim().toLowerCase();
  if (!q) return tasks.value;
  return tasks.value.filter((task) => {
    return (
      task.name.toLowerCase().includes(q) ||
      task.description.toLowerCase().includes(q) ||
      task.command.toLowerCase().includes(q) ||
      (task.depends && task.depends.some((d) => d.toLowerCase().includes(q))) ||
      (task.env && task.env.toLowerCase().includes(q))
    );
  });
});

async function fetchTasks() {
  loading.value = true;
  try {
    const res = await fetch(`/api/tasks?scope=${activeScope.value}`);
    if (res.ok) {
      const data = await res.json();
      if (Array.isArray(data.tasks)) {
        tasks.value = data.tasks;
      }
    }
  } catch (err: any) {
    emit("log-event", "ERROR", `获取任务列表失败: ${err.message}`, undefined, "tasks");
  } finally {
    loading.value = false;
  }
}

watch(
  () => activeScope.value,
  () => {
    fetchTasks();
  }
);

watch(
  () => props.activeSubItem,
  (newSubItem) => {
    if (newSubItem) {
      searchKey.value = newSubItem;
    }
  },
  { immediate: true }
);

onMounted(() => {
  fetchTasks();
});

async function executeTask(task: ProjectTask) {
  if (runningTask.value !== null) return;
  runningTask.value = task.name;
  emit("task-started", task.name);
  emit("log-event", "INFO", `[unirtm run ${task.name}] 正在启动任务...`, `执行命令: ${task.command}`, "tasks");

  try {
    const res = await fetch("/api/tasks/run", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        name: task.name,
        scope: activeScope.value,
      }),
    });

    const data = await res.json();
    if (data.success) {
      emit(
        "log-event",
        "SUCCESS",
        `[${task.name}] 任务执行成功 (退出码: 0)`,
        data.output ? data.output : undefined,
        "tasks"
      );
    } else {
      emit(
        "log-event",
        "ERROR",
        `[${task.name}] 任务执行失败 (退出码: ${data.exit_code ?? 1})`,
        data.output || data.error || undefined,
        "tasks"
      );
    }
  } catch (err: any) {
    emit("log-event", "ERROR", `[${task.name}] 执行网络异常: ${err.message}`, undefined, "tasks");
  } finally {
    emit("task-finished", task.name);
    runningTask.value = null;
  }
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

/* Scope Banner */
.scope-selection-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.65rem 1rem;
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.03);
  gap: 1rem;
}

.scope-context-pill {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.scope-badge {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-size: 0.72rem;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 9999px;
  background: rgba(2, 132, 199, 0.12);
  color: var(--accent-cyan);
  border: 1px solid rgba(2, 132, 199, 0.25);
}

.scope-badge.user {
  background: rgba(168, 85, 247, 0.12);
  color: #c084fc;
  border-color: rgba(168, 85, 247, 0.25);
}

.scope-context-config {
  font-size: 0.72rem;
  color: var(--text-muted);
  opacity: 0.8;
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.scope-hint-text {
  font-size: 0.78rem;
  color: var(--text-muted);
}

/* Filter Bar */
.tasks-filter-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.65rem 0.85rem;
  border-radius: 10px;
  gap: 1rem;
}

.search-input-wrapper {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex: 1;
  max-width: 450px;
}

.search-icon {
  color: var(--text-muted);
  display: inline-flex;
  align-items: center;
}

.task-search-input {
  width: 100%;
  padding: 0.4rem 0.65rem;
  font-size: 0.82rem;
  border-radius: 6px;
  border: 1px solid var(--card-border);
  background: var(--input-bg);
  color: var(--text-main);
  outline: none;
  transition: border-color 0.15s ease;
}

.task-search-input:focus {
  border-color: var(--accent-cyan);
}

.task-stats-badges {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.stat-badge {
  font-size: 0.78rem;
  color: var(--text-muted);
  padding: 0.25rem 0.6rem;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  border-radius: 6px;
}

.stat-badge b {
  color: var(--text-main);
}

/* Table Container & List */
.tasks-table-container {
  padding: 1rem;
  border-radius: 12px;
}

.tasks-empty-box {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3rem 1.5rem;
  gap: 0.75rem;
  text-align: center;
  color: var(--text-muted);
}

.empty-title {
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--text-main);
}

.empty-desc {
  font-size: 0.8rem;
  margin: 0;
  max-width: 480px;
  line-height: 1.45;
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
  gap: 1rem;
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
  gap: 0.4rem;
  flex: 1;
  min-width: 0;
}

.task-name-line {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  flex-wrap: wrap;
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
  padding: 0.08rem 0.45rem;
  border-radius: 4px;
}

.task-depends-wrap {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  flex-wrap: wrap;
}

.depends-label {
  font-size: 0.72rem;
  color: var(--text-muted);
}

.depends-pill {
  font-size: 0.68rem;
  background: rgba(148, 163, 184, 0.12);
  border: 1px solid rgba(148, 163, 184, 0.25);
  color: var(--text-muted);
  padding: 0.05rem 0.35rem;
  border-radius: 4px;
}

.task-cmd-line {
  font-size: 0.78rem;
  color: var(--accent-cyan);
  background: rgba(0, 0, 0, 0.2);
  padding: 0.25rem 0.6rem;
  border-radius: 6px;
  width: fit-content;
  max-width: 100%;
  overflow-x: auto;
  white-space: pre-wrap;
  word-break: break-all;
}

.cmd-prompt {
  color: var(--text-muted);
  margin-right: 0.25rem;
}

.task-desc-line {
  font-size: 0.76rem;
  color: var(--text-muted);
  line-height: 1.4;
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
  white-space: nowrap;
}

.run-task-btn:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 14px rgba(0, 229, 255, 0.35);
}

.run-task-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.tool-action-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.4rem 0.75rem;
  font-size: 0.78rem;
  font-weight: 600;
  border-radius: 6px;
  border: 1px solid var(--border-color);
  background: var(--button-secondary-bg, rgba(255, 255, 255, 0.05));
  color: var(--text-main);
  cursor: pointer;
  transition: all 0.2s ease;
}

.tool-action-btn:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.1);
  border-color: var(--card-border-active);
}

.spin-icon {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.task-icon-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 6px;
  border: 1px solid var(--border-color, rgba(255, 255, 255, 0.1));
  background: var(--button-secondary-bg, rgba(255, 255, 255, 0.05));
  color: var(--text-muted, #94a3b8);
  cursor: pointer;
  transition: all 0.2s ease;
}

.task-icon-btn:hover {
  background: rgba(255, 255, 255, 0.12);
  color: var(--text-main, #fff);
  border-color: var(--accent-cyan, #00e5ff);
  transform: translateY(-1px);
}

.task-icon-btn.delete:hover {
  background: rgba(239, 68, 68, 0.15);
  border-color: rgba(239, 68, 68, 0.4);
  color: #ef4444;
}

.tool-action-btn.primary {
  background: linear-gradient(135deg, var(--accent-cyan, #00e5ff), #0284c7);
  color: #070a12;
  border: none;
}

.tool-action-btn.primary:hover:not(:disabled) {
  box-shadow: 0 4px 14px rgba(0, 229, 255, 0.35);
  transform: translateY(-1px);
}

.tool-action-btn.danger {
  background: #dc2626;
  color: #fff;
  border: none;
}

.tool-action-btn.danger:hover:not(:disabled) {
  background: #ef4444;
  box-shadow: 0 4px 14px rgba(239, 68, 68, 0.35);
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(4, 8, 16, 0.75);
  backdrop-filter: blur(8px);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.glass-modal {
  background: var(--card-bg, #111520);
  border: 1px solid var(--card-border, rgba(255, 255, 255, 0.1));
  box-shadow: 0 16px 48px rgba(0, 0, 0, 0.4), 0 0 24px rgba(0, 229, 255, 0.15);
  border-radius: 14px;
  width: 92%;
  max-width: 440px;
  overflow: hidden;
  color: var(--text-main, #e2e8f0);
}

.modal-header {
  padding: 1rem 1.3rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: rgba(255, 255, 255, 0.02);
  border-bottom: 1px solid var(--card-border, rgba(255, 255, 255, 0.08));
}

.header-title {
  display: flex;
  align-items: center;
  gap: 0.6rem;
}

.close-btn {
  background: transparent;
  border: none;
  color: var(--text-muted, #94a3b8);
  font-size: 1.1rem;
  cursor: pointer;
  padding: 0.2rem 0.4rem;
  border-radius: 4px;
  transition: all 0.2s;
}

.close-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  color: var(--text-main, #fff);
}
</style>
