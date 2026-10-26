<template>
  <div class="view-container tasks-view" id="view-tasks" @contextmenu.prevent>
    <!-- View Header -->
    <div class="view-header-row">
      <div>
        <h2 class="view-title" style="display: flex; align-items: center; gap: 8px">
          <TechIcon id="workbench:tasks" :size="20" />
          <span>{{ t("tasks.title") || "Unified Task Runner" }}</span>
        </h2>
        <p class="view-desc">
          {{ t("tasks.desc") || "Cross-language tasks defined in .unirtm.toml, executed cross-platform in one click." }}
        </p>
      </div>
      <div class="view-actions" style="display: flex; gap: 8px; align-items: center">
        <button
          class="tool-action-btn secondary"
          @click="fetchTasks"
          :disabled="loading"
          :title="t('tasks.refresh_title') || 'Refresh tasks list'"
        >
          <TechIcon id="workbench:sync" :size="14" :class="{ 'spin-icon': loading }" />
          <span>{{ loading ? t("common.refreshing") || "Refreshing..." : t("common.refresh") || "Refresh" }}</span>
        </button>
      </div>
    </div>

    <!-- Current Scope Context Banner -->
    <div class="scope-selection-bar glass-card" id="tasks-scope-bar">
      <div class="scope-context-pill">
        <span class="scope-badge" :class="activeScope">
          <TechIcon
            :id="activeScope === 'user' ? 'workbench:settings' : 'workbench:folder'"
            :size="13"
            :colored="false"
          />
          {{
            activeScope === "user" ? t("runtimes.scope_user") || "User" : t("runtimes.scope_workspace") || "Workspace"
          }}
        </span>
        <span class="scope-hint-text">
          {{
            activeScope === "user"
              ? t("runtimes.scope_user_desc") || "Default fallback environment for user profile"
              : t("runtimes.scope_workspace_desc") || "Lock language versions for current workspace"
          }}
        </span>
      </div>
      <div
        class="scope-context-config font-mono"
        :title="activeScope === 'user' ? '~/.config/unirtm/unirtm.toml' : 'unirtm.toml'"
      >
        <span class="file-icon">📄</span>
        <span>{{ activeScope === "user" ? "~/.config/unirtm/unirtm.toml" : "unirtm.toml" }}</span>
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
          :placeholder="t('tasks.search_placeholder') || 'Search task name, dependency, command, or description...'"
        />
        <button
          v-if="searchKey"
          type="button"
          class="search-clear-btn"
          :title="t('tasks.clear_search') || 'Clear search'"
          @click="searchKey = ''"
        >
          <span class="codicon codicon-close"></span>
        </button>
      </div>
      <div class="task-stats-badges">
        <span class="stat-badge">
          <b>{{ filteredTasks.length }}</b> {{ t("tasks.count_suffix") || "Tasks" }}
        </span>

        <!-- View Mode Switcher -->
        <div class="view-mode-toggle" id="tasks-view-mode-toggle">
          <button
            type="button"
            class="view-mode-btn"
            :class="{ active: viewMode === 'cards' }"
            @click="setViewMode('cards')"
            :title="t('environments.view_cards') || 'Cards View'"
          >
            <TechIcon id="workbench:cards" :size="14" />
          </button>
          <button
            type="button"
            class="view-mode-btn"
            :class="{ active: viewMode === 'list' }"
            @click="setViewMode('list')"
            :title="t('environments.view_list') || 'List View'"
          >
            <TechIcon id="workbench:list" :size="14" />
          </button>
        </div>

        <button
          class="tool-action-btn primary small"
          id="btn-add-task"
          @click="openCreateModal"
          :title="t('tasks.create_title') || 'Create new automated project task'"
        >
          <span>+ {{ t("tasks.create_btn") || "New Task" }}</span>
        </button>
      </div>
    </div>

    <!-- Tasks List Container -->
    <div class="tasks-table-container glass-card">
      <!-- Loading State -->
      <div v-if="loading && tasks.length === 0" class="tasks-empty-box">
        <TechIcon id="workbench:sync" :size="24" class="spin-icon" />
        <span style="font-size: 0.85rem; color: var(--text-muted)">{{
          t("tasks.loading_tasks") || "Parsing tasks from configuration..."
        }}</span>
      </div>

      <!-- Empty State -->
      <div v-else-if="filteredTasks.length === 0" class="tasks-empty-box">
        <TechIcon id="workbench:tasks" :size="32" :colored="false" />
        <div class="empty-title">
          {{
            searchKey
              ? t("tasks.empty_search", { query: searchKey }) || `No tasks found matching "${searchKey}"`
              : activeScope === "user"
                ? t("tasks.empty_user") || "No tasks defined in user global profile"
                : t("tasks.empty_workspace") || "No [tasks] defined in current workspace config"
          }}
        </div>
        <p class="empty-desc">
          {{
            searchKey
              ? t("tasks.empty_search_desc") ||
                "Try shortening keywords or searching for other commands or dependencies."
              : activeScope === "user"
                ? t("tasks.empty_user_desc") ||
                  "Configure [tasks.task-name] in user profile (~/.config/unirtm/unirtm.toml) to define common tasks."
                : t("tasks.empty_workspace_desc") ||
                  "Add [tasks.task-name] in workspace unirtm.toml to define builds, tests, and security scans."
          }}
        </p>
        <button v-if="searchKey" class="tool-action-btn secondary" style="margin-top: 0.75rem" @click="searchKey = ''">
          <span class="codicon codicon-clear-all"></span>
          <span>{{ t("tasks.clear_search_btn") || "Clear Search Filter" }}</span>
        </button>
        <button v-else class="tool-action-btn primary" style="margin-top: 0.75rem" @click="openCreateModal">
          <TechIcon id="workbench:edit" :size="14" />
          <span>{{ t("tasks.create_first") || "Create First Task" }}</span>
        </button>
      </div>

      <!-- Tasks Cards View -->
      <div v-else-if="viewMode === 'cards'" class="tasks-list">
        <div
          v-for="task in filteredTasks"
          :key="task.name"
          class="task-row-card"
          :class="{ running: runningTask === task.name }"
        >
          <div class="task-info-side">
            <div class="task-name-line">
              <span class="task-title font-mono" style="display: inline-flex; align-items: center; gap: 6px">
                <TechIcon id="workbench:tasks" :size="14" />
                {{ task.name }}
              </span>
              <span
                v-if="task.env"
                class="task-env-pill font-mono"
                :title="`${t('tasks.col_env') || 'Environment'}: ${task.env}`"
              >
                env: {{ task.env }}
              </span>
              <div v-if="task.depends && task.depends.length > 0" class="task-depends-wrap">
                <span class="depends-label">{{ t("tasks.depends_label") || "Depends:" }}</span>
                <span v-for="dep in task.depends" :key="dep" class="depends-pill font-mono">
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
            <div class="task-more-menu-wrap" @click.stop>
              <button
                class="task-more-btn"
                :class="{ active: activeMenuTask === task.name }"
                @click.stop="toggleTaskMenu(task.name)"
                :title="`${t('common.more_actions') || 'More Actions'}: ${task.name}`"
              >
                <TechIcon id="workbench:more" :size="15" />
              </button>
              <transition name="dropdown-fade">
                <div v-if="activeMenuTask === task.name" class="task-dropdown-menu">
                  <button class="dropdown-item" @click.stop="handleEditFromMenu(task)">
                    <TechIcon id="workbench:edit" :size="13" />
                    <span>{{ t("common.edit") || "Edit" }}</span>
                  </button>
                  <button class="dropdown-item danger" @click.stop="handleDeleteFromMenu(task)">
                    <TechIcon id="workbench:delete" :size="13" />
                    <span>{{ t("common.delete") || "Delete" }}</span>
                  </button>
                </div>
              </transition>
            </div>
            <button
              class="run-task-btn"
              :disabled="runningTask !== null"
              @click="executeTask(task)"
              :title="`${t('tasks.run_task') || 'Run Task'}: unirtm run ${task.name}`"
            >
              <span v-if="runningTask === task.name" style="display: inline-flex; align-items: center; gap: 6px">
                <TechIcon id="workbench:sync" :size="12" class="spin-icon" />
                <span>{{ t("tasks.running") || "Running" }}</span>
              </span>
              <span v-else style="display: inline-flex; align-items: center; gap: 4px">
                <TechIcon id="workbench:tasks" :size="12" :colored="false" />
                <span>{{ t("tasks.run") || "Run" }}</span>
              </span>
            </button>
          </div>
        </div>
      </div>

      <!-- Tasks High-Density Table / List View -->
      <div v-else class="tasks-list-table-wrapper">
        <table class="tasks-table">
          <thead>
            <tr>
              <th class="th-name">{{ t("tasks.col_name") || "Task Name" }}</th>
              <th class="th-command">{{ t("tasks.col_command") || "Command" }}</th>
              <th class="th-depends">{{ t("tasks.col_depends") || "Dependencies & Info" }}</th>
              <th class="th-actions">{{ t("tasks.col_actions") || "Actions" }}</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="task in filteredTasks"
              :key="task.name"
              class="task-table-row"
              :class="{ running: runningTask === task.name }"
            >
              <!-- Name & Env -->
              <td class="td-name">
                <div class="table-name-cell">
                  <TechIcon id="workbench:tasks" :size="15" class="task-icon" />
                  <span class="task-name font-mono">{{ task.name }}</span>
                  <span
                    v-if="task.env"
                    class="task-env-pill font-mono"
                    :title="`${t('tasks.col_env') || 'Environment'}: ${task.env}`"
                  >
                    env: {{ task.env }}
                  </span>
                </div>
              </td>

              <!-- Command -->
              <td class="td-command">
                <div class="table-cmd-box font-mono" :title="task.command">
                  <span class="cmd-prompt">$</span>
                  <span class="cmd-text">{{ task.command }}</span>
                </div>
              </td>

              <!-- Dependencies & Description -->
              <td class="td-depends">
                <div class="table-meta-cell">
                  <div v-if="task.depends && task.depends.length > 0" class="task-depends-wrap">
                    <span v-for="dep in task.depends" :key="dep" class="depends-pill font-mono">
                      {{ dep }}
                    </span>
                  </div>
                  <span v-if="task.description" class="table-task-desc" :title="task.description">
                    {{ task.description }}
                  </span>
                  <span
                    v-else-if="!task.depends || task.depends.length === 0"
                    style="color: var(--text-muted); opacity: 0.5"
                    >-</span
                  >
                </div>
              </td>

              <!-- Actions -->
              <td class="td-actions">
                <div class="table-row-actions">
                  <!-- Row 1: More Actions Menu -->
                  <div class="task-more-menu-wrap" @click.stop>
                    <button
                      class="task-more-btn table-more-btn"
                      :class="{ active: activeMenuTask === task.name }"
                      @click.stop="toggleTaskMenu(task.name)"
                      :title="`${t('common.more_actions') || 'More Actions'}: ${task.name}`"
                    >
                      <TechIcon id="workbench:more" :size="14" />
                    </button>
                    <transition name="dropdown-fade">
                      <div v-if="activeMenuTask === task.name" class="task-dropdown-menu table-menu" @click.stop>
                        <button class="dropdown-item" @click.stop="handleEditFromMenu(task)">
                          <TechIcon id="workbench:edit" :size="13" />
                          <span>{{ t("common.edit") || "Edit" }}</span>
                        </button>
                        <button class="dropdown-item danger" @click.stop="handleDeleteFromMenu(task)">
                          <TechIcon id="workbench:delete" :size="13" />
                          <span>{{ t("common.delete") || "Delete" }}</span>
                        </button>
                      </div>
                    </transition>
                  </div>

                  <!-- Row 2: Run Task Button -->
                  <button
                    class="run-task-btn small"
                    :disabled="runningTask !== null"
                    @click="executeTask(task)"
                    :title="`${t('tasks.run_task') || 'Run Task'}: unirtm run ${task.name}`"
                  >
                    <span v-if="runningTask === task.name" style="display: inline-flex; align-items: center; gap: 4px">
                      <TechIcon id="workbench:sync" :size="12" class="spin-icon" />
                      <span>{{ t("tasks.running") || "Running" }}</span>
                    </span>
                    <span v-else style="display: inline-flex; align-items: center; gap: 4px">
                      <TechIcon id="workbench:tasks" :size="12" :colored="false" />
                      <span>{{ t("tasks.run") || "Run" }}</span>
                    </span>
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
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
          <div class="header-title" style="color: #ef4444">
            <TechIcon id="workbench:warning" :size="20" />
            <h3 style="margin: 0; font-size: 1.05rem">{{ t("tasks.delete_title") || "Delete Task" }}</h3>
          </div>
          <button class="btn btn-icon close-btn" @click="taskToDelete = null" :title="t('common.close') || 'Close'">
            <TechIcon id="workbench:close" :size="14" />
          </button>
        </div>
        <div class="modal-body" style="padding: 1.25rem 1.4rem">
          <p style="margin: 0 0 0.75rem 0; font-size: 0.85rem; color: var(--text-main)">
            {{ t("tasks.delete_confirm", { name: taskToDelete.name }) }}
          </p>
          <p style="margin: 0; font-size: 0.75rem; color: var(--text-muted); line-height: 1.4">
            {{
              t("tasks.delete_warning", {
                file: activeScope === "user" ? "~/.config/unirtm/unirtm.toml" : "unirtm.toml",
              })
            }}
          </p>
          <div
            class="modal-footer"
            style="
              display: flex;
              justify-content: flex-end;
              gap: 0.75rem;
              margin-top: 1.25rem;
              padding-top: 1rem;
              border-top: 1px solid var(--card-border);
            "
          >
            <button class="tool-action-btn secondary" @click="taskToDelete = null" :disabled="deletingTask">
              {{ t("common.cancel") || "Cancel" }}
            </button>
            <button class="tool-action-btn danger" @click="handleConfirmDelete" :disabled="deletingTask">
              <TechIcon v-if="deletingTask" id="workbench:sync" :size="14" class="spin-icon" />
              <span>{{ deletingTask ? t("common.deleting") || "Deleting..." : t("common.delete") || "Delete" }}</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from "vue";
import { t } from "../i18n";
import TechIcon from "../components/common/TechIcon.vue";
import TaskModal, { type TaskFormData } from "../components/TaskModal.vue";
import { useAppScope } from "../composables/useAppScope";
import { useToast } from "../composables/useToast";
import { useWorkspaceBus, type TaskActionRequest } from "../composables/useWorkspaceBus";

const props = defineProps<{
  activeSubItem?: string;
  initialScope?: string;
}>();

const emit = defineEmits<{
  (
    e: "log-event",
    level: "INFO" | "SUCCESS" | "WARN" | "ERROR",
    msg: string,
    detail?: string,
    category?: "output" | "tasks" | "doctor"
  ): void;
  (e: "task-started", taskName: string): void;
  (e: "task-finished", taskName: string): void;
}>();

const { activeScope } = useAppScope();
const { showToast } = useToast();
const { onTaskAction, notifyTasksChanged } = useWorkspaceBus();
let unsubTaskAction: (() => void) | null = null;

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

const viewMode = ref<"cards" | "list">((localStorage.getItem("unirtm_view_mode_tasks") as "cards" | "list") || "cards");

function setViewMode(mode: "cards" | "list") {
  viewMode.value = mode;
  localStorage.setItem("unirtm_view_mode_tasks", mode);
}

// Task modal state
const isTaskModalOpen = ref(false);
const isEditMode = ref(false);
const editingTask = ref<ProjectTask | null>(null);
const savingTask = ref(false);

// Delete confirm modal state
const taskToDelete = ref<ProjectTask | null>(null);
const deletingTask = ref(false);

// Three-dots more menu state
const activeMenuTask = ref<string | null>(null);

function toggleTaskMenu(name: string) {
  activeMenuTask.value = activeMenuTask.value === name ? null : name;
}

function handleEditFromMenu(task: ProjectTask) {
  activeMenuTask.value = null;
  openEditModal(task);
}

function handleDeleteFromMenu(task: ProjectTask) {
  activeMenuTask.value = null;
  openDeleteConfirm(task);
}

function closeTaskMenu() {
  if (activeMenuTask.value === null) return;
  activeMenuTask.value = null;
}

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
      showToast(
        isEditMode.value
          ? t("tasks.updated_msg", { name: formData.name }) || `Task ${formData.name} updated successfully`
          : t("tasks.created_msg", { name: formData.name }) || `Task ${formData.name} created successfully`,
        "success"
      );
      emit("log-event", "SUCCESS", `[Tasks] ${formData.name} saved successfully`, undefined, "tasks");
      isTaskModalOpen.value = false;
      await fetchTasks();
      notifyTasksChanged();
    } else {
      showToast(data.error || t("tasks.save_failed") || "Failed to save task", "error");
      emit("log-event", "ERROR", `[Tasks] ${formData.name} save failed: ${data.error}`, undefined, "tasks");
    }
  } catch (err: any) {
    showToast(err.message, "error");
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
      showToast(t("tasks.deleted_msg", { name }) || `Task ${name} deleted successfully`, "success");
      emit("log-event", "SUCCESS", `[Tasks] ${name} deleted`, undefined, "tasks");
      taskToDelete.value = null;
      await fetchTasks();
      notifyTasksChanged();
    } else {
      showToast(data.error || t("tasks.delete_failed") || "Failed to delete task", "error");
    }
  } catch (err: any) {
    showToast(err.message, "error");
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
    emit(
      "log-event",
      "ERROR",
      `${t("tasks.fetch_failed") || "Failed to fetch task list"}: ${err.message}`,
      undefined,
      "tasks"
    );
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

function handleGlobalKeyDown(e: KeyboardEvent) {
  if (e.key === "Escape" && activeMenuTask.value) {
    closeTaskMenu();
  }
}

function handleBusTaskAction(req: TaskActionRequest) {
  let target = tasks.value.find((t) => t.name === req.taskName);
  if (!target) {
    target = {
      name: req.taskName,
      command: req.command || req.taskCmd || (req.task as any)?.cmd || "",
      description: (req.task as any)?.desc || "",
    };
  }

  switch (req.action) {
    case "run":
      executeTask(target);
      break;
    case "edit":
      openEditModal(target);
      break;
    case "delete":
      openDeleteConfirm(target);
      break;
  }
}

onMounted(() => {
  fetchTasks();
  window.addEventListener("click", closeTaskMenu);
  window.addEventListener("keydown", handleGlobalKeyDown);
  unsubTaskAction = onTaskAction(handleBusTaskAction);
});

onUnmounted(() => {
  window.removeEventListener("click", closeTaskMenu);
  window.removeEventListener("keydown", handleGlobalKeyDown);
  if (unsubTaskAction) {
    unsubTaskAction();
    unsubTaskAction = null;
  }
});

async function executeTask(task: ProjectTask) {
  if (runningTask.value !== null) return;
  const taskName = task?.name || (task as any)?.id;
  if (!taskName) return;
  runningTask.value = taskName;
  emit("task-started", taskName);
  emit(
    "log-event",
    "INFO",
    t("tasks.running_task", { taskName }) || `[unirtm run ${taskName}] Starting task...`,
    t("tasks.exec_command", { command: task.command }) || `Executing command: ${task.command}`,
    "tasks"
  );

  try {
    const res = await fetch("/api/tasks/run", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        name: taskName,
        task: taskName,
        scope: activeScope.value,
      }),
    });

    const data = await res.json();
    if (data.success) {
      emit(
        "log-event",
        "SUCCESS",
        t("tasks.task_success", { taskName }) || `[${taskName}] Task executed successfully (exit code: 0)`,
        data.output ? data.output : undefined,
        "tasks"
      );
    } else {
      emit(
        "log-event",
        "ERROR",
        t("tasks.task_failed", { taskName, code: data.exit_code ?? 1 }) ||
          `[${taskName}] Task execution failed (exit code: ${data.exit_code ?? 1})`,
        data.output || data.error || undefined,
        "tasks"
      );
    }
  } catch (err: any) {
    emit(
      "log-event",
      "ERROR",
      t("tasks.network_error", { taskName, error: err.message }) ||
        `[${taskName}] Network execution error: ${err.message}`,
      undefined,
      "tasks"
    );
  } finally {
    emit("task-finished", taskName);
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
  position: relative;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex: 1;
  max-width: 450px;
}

.search-clear-btn {
  position: absolute;
  right: 8px;
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2px;
  border-radius: 4px;
  transition: all 0.15s ease;
}

.search-clear-btn:hover {
  color: var(--text-main);
  background: rgba(128, 128, 128, 0.15);
}

.search-icon {
  color: var(--text-muted);
  display: inline-flex;
  align-items: center;
}

.task-search-input {
  width: 100%;
  padding: 0.4rem 2rem 0.4rem 0.65rem;
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
  align-items: stretch;
  justify-content: space-between;
  padding: 0.95rem 1.15rem;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  border-radius: 10px;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  gap: 1.25rem;
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
  color: #fff;
  box-shadow: 0 2px 8px var(--accent-cyan-glow);
  transition: all 0.2s ease;
  white-space: nowrap;
}

.run-task-btn:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 14px var(--accent-cyan-glow);
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
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.task-actions-side {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 0.15rem 0;
  flex-shrink: 0;
}

.task-more-menu-wrap {
  position: relative;
  display: inline-flex;
}

.task-more-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  padding: 0;
  border-radius: 7px;
  border: 1px solid var(--card-border, rgba(255, 255, 255, 0.12));
  background: var(--section-bg);
  color: var(--text-muted);
  cursor: pointer;
  opacity: 0.75;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  flex-shrink: 0;
}

.task-more-btn:hover,
.task-more-btn.active {
  background: var(--input-bg);
  color: var(--text-main);
  border-color: var(--card-border-active);
  opacity: 1;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.12);
  transform: translateY(-1px);
}

[data-theme="light"] .task-more-btn {
  background: #f8fafc;
  border-color: #cbd5e1;
  color: #64748b;
}

[data-theme="light"] .task-more-btn:hover,
[data-theme="light"] .task-more-btn.active {
  background: rgba(2, 132, 199, 0.08);
  border-color: var(--accent-cyan);
  color: var(--accent-cyan);
}

.task-row-card:hover .task-more-btn {
  opacity: 0.85;
}

.task-row-card:hover .task-more-btn:hover,
.task-row-card:hover .task-more-btn.active {
  opacity: 1;
}

.task-dropdown-menu {
  position: absolute;
  top: calc(100% + 6px);
  right: 0;
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  border-radius: 8px;
  box-shadow:
    0 10px 25px -5px rgba(0, 0, 0, 0.25),
    0 0 12px var(--accent-cyan-glow);
  padding: 0.35rem;
  z-index: 50;
  min-width: 130px;
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.task-dropdown-menu.table-menu {
  top: calc(100% + 4px);
  right: 0;
  text-align: left;
}

[data-theme="light"] .task-dropdown-menu {
  background: #ffffff;
  border-color: #e2e8f0;
  box-shadow:
    0 10px 25px -5px rgba(0, 0, 0, 0.1),
    0 0 12px rgba(2, 132, 199, 0.1);
}

.task-dropdown-menu .dropdown-item {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  padding: 0.55rem 0.8rem;
  font-size: 0.82rem;
  font-weight: 500;
  color: var(--text-main);
  background: transparent;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  width: 100%;
  text-align: left;
  transition: all 0.15s ease;
  white-space: nowrap;
}

.task-dropdown-menu .dropdown-item:hover:not(.danger) {
  background-color: var(--accent-cyan-glow, rgba(0, 229, 255, 0.15));
  color: var(--accent-cyan);
}
[data-theme="light"] .task-dropdown-menu .dropdown-item:hover:not(.danger) {
  background-color: rgba(2, 132, 199, 0.12);
  color: #0284c7;
}

.task-dropdown-menu .dropdown-item.danger,
.task-dropdown-menu .dropdown-item.danger span {
  color: #ef4444;
}

[data-theme="dark"] .task-dropdown-menu .dropdown-item.danger,
[data-theme="dark"] .task-dropdown-menu .dropdown-item.danger span {
  color: #f87171;
}

.task-dropdown-menu .dropdown-item.danger:hover {
  background: rgba(239, 68, 68, 0.12);
  color: #ef4444;
}
[data-theme="dark"] .task-dropdown-menu .dropdown-item.danger:hover {
  background: rgba(239, 68, 68, 0.15);
  color: #f87171;
}

/* Dropdown menu animation */
.dropdown-fade-enter-active,
.dropdown-fade-leave-active {
  transition: all 0.15s cubic-bezier(0.16, 1, 0.3, 1);
}

.dropdown-fade-enter-from,
.dropdown-fade-leave-to {
  opacity: 0;
  transform: translateY(-4px) scale(0.98);
}

.tool-action-btn.primary {
  background: linear-gradient(135deg, var(--accent-cyan), #0284c7);
  color: #fff;
  font-weight: 600;
  border: none;
  box-shadow: 0 2px 8px var(--accent-cyan-glow);
}

.tool-action-btn.primary:hover:not(:disabled) {
  box-shadow: 0 4px 14px var(--accent-cyan-glow);
  transform: translateY(-1px);
}

.tool-action-btn.small {
  padding: 0.35rem 0.75rem;
  font-size: 0.78rem;
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
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.delete-confirm-modal {
  background: var(--modal-bg);
  border: 1px solid var(--card-border);
  box-shadow: 0 20px 48px -10px rgba(0, 0, 0, 0.35);
  border-radius: 16px;
  width: 92%;
  max-width: 440px;
  overflow: hidden;
  color: var(--text-main);
}

.modal-header {
  padding: 1.15rem 1.4rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: var(--section-bg);
  border-bottom: 1px solid var(--card-border);
}

.header-title {
  display: flex;
  align-items: center;
  gap: 0.6rem;
}

.close-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  font-size: 1.1rem;
  cursor: pointer;
  padding: 0.25rem 0.5rem;
  border-radius: 6px;
  transition: all 0.2s;
}

.close-btn:hover {
  background: var(--card-bg);
  color: var(--text-main);
}

/* View Mode Switcher Toggle */
.view-mode-toggle {
  display: inline-flex;
  align-items: center;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  border-radius: 6px;
  padding: 2px;
  gap: 2px;
}

.view-mode-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 26px;
  height: 26px;
  padding: 0;
  border: none;
  background: transparent;
  color: var(--text-muted);
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.view-mode-btn:hover {
  color: var(--text-main);
  background: rgba(255, 255, 255, 0.05);
}

.view-mode-btn.active {
  color: var(--accent-cyan);
  background: rgba(0, 229, 255, 0.12);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
}

/* High-Density Tasks Table */
.tasks-list-table-wrapper {
  width: 100%;
  overflow-x: auto;
  border: 1px solid var(--card-border);
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.015);
}

.tasks-table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
  font-size: 0.84rem;
}

.tasks-table thead tr {
  background: rgba(255, 255, 255, 0.03);
  border-bottom: 1px solid var(--card-border);
}

.tasks-table th {
  padding: 0.75rem 1rem;
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: var(--text-muted);
  white-space: nowrap;
}

.tasks-table th.th-actions {
  text-align: right;
  width: 100px;
}

.task-table-row {
  border-bottom: 1px solid var(--card-border, rgba(255, 255, 255, 0.05));
  transition: background 0.15s ease;
}

.task-table-row:last-child {
  border-bottom: none;
}

.task-table-row:hover {
  background: rgba(255, 255, 255, 0.035);
}

.task-table-row.running {
  border-left: 3px solid var(--accent-cyan);
  background: rgba(0, 229, 255, 0.06);
}

.tasks-table td {
  padding: 0.65rem 1rem;
  vertical-align: middle;
}

.tasks-table td.td-actions {
  text-align: right;
}

.table-name-cell {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.table-name-cell .task-name {
  font-weight: 600;
  color: var(--text-main);
  letter-spacing: 0.3px;
}

.table-cmd-box {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  border-radius: 6px;
  padding: 3px 8px;
  max-width: 360px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 0.78rem;
  color: var(--accent-cyan);
}

.table-meta-cell {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.table-task-desc {
  font-size: 0.76rem;
  color: var(--text-muted);
  max-width: 300px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.table-row-actions {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  justify-content: center;
  gap: 5px;
}

.run-task-btn.small {
  padding: 3px 9px;
  font-size: 0.75rem;
}

.task-more-btn.table-more-btn {
  width: 26px;
  height: 26px;
  padding: 0;
  border-radius: 6px;
  opacity: 0.75;
}

.task-more-btn.table-more-btn:hover,
.task-more-btn.table-more-btn.active {
  opacity: 1;
}

/* Light Theme Overrides for Tasks Table */
:root[data-theme="light"] .tasks-list-table-wrapper,
[data-theme="light"] .tasks-list-table-wrapper {
  background: #ffffff;
  border-color: #e2e8f0;
}

:root[data-theme="light"] .tasks-table thead tr,
[data-theme="light"] .tasks-table thead tr {
  background: #f8fafc;
  border-bottom-color: #e2e8f0;
}

:root[data-theme="light"] .tasks-table th,
[data-theme="light"] .tasks-table th {
  color: #64748b;
}

:root[data-theme="light"] .task-table-row,
[data-theme="light"] .task-table-row {
  border-bottom-color: #f1f5f9;
}

:root[data-theme="light"] .task-table-row:hover,
[data-theme="light"] .task-table-row:hover {
  background: rgba(2, 132, 199, 0.03);
}

:root[data-theme="light"] .table-cmd-box,
[data-theme="light"] .table-cmd-box {
  background: #f8fafc;
  border-color: #e2e8f0;
  color: #0284c7;
}

:root[data-theme="light"] .view-mode-toggle,
[data-theme="light"] .view-mode-toggle {
  background: #f8fafc;
  border-color: #cbd5e1;
}

:root[data-theme="light"] .view-mode-btn,
[data-theme="light"] .view-mode-btn {
  color: #64748b;
}

:root[data-theme="light"] .view-mode-btn:hover,
[data-theme="light"] .view-mode-btn:hover {
  background: rgba(0, 0, 0, 0.05);
  color: #0f172a;
}

:root[data-theme="light"] .view-mode-btn.active,
[data-theme="light"] .view-mode-btn.active {
  background: #ffffff;
  color: #0284c7;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

:root[data-theme="light"] .action-btn,
[data-theme="light"] .action-btn {
  color: #64748b;
}

:root[data-theme="light"] .action-btn:hover,
[data-theme="light"] .action-btn:hover {
  background: rgba(0, 0, 0, 0.05);
  color: #0f172a;
}
</style>
