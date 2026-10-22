<template>
  <aside
    v-if="isOpen"
    class="sidebar-panel"
    :class="{ 'is-resizing': isDragging }"
    id="app-sidebar-panel"
    :style="{ width: `${sidebarWidth}px`, minWidth: `${sidebarWidth}px` }"
  >
    <!-- Panel Header -->
    <div class="sidebar-header">
      <span class="panel-title">{{ headerTitle }}</span>
      <div class="panel-tools">
        <button class="panel-tool-btn" title="Collapse Sidebar" @click="$emit('toggle-collapse')">
          <span>◀</span>
        </button>
      </div>
    </div>

    <!-- Filter input -->
    <div class="sidebar-search">
      <input v-model="searchQuery" type="text" class="sidebar-search-input" :placeholder="searchPlaceholder" />
    </div>

    <!-- Panel Content based on currentTab -->
    <div class="sidebar-content-list">
      <!-- DASHBOARD Navigation Items -->
      <template v-if="currentTab === 'dashboard'">
        <div class="nav-group-title">{{ t("sidebar.group_overview") || "PROJECT OVERVIEW" }}</div>
        <button
          v-for="item in filteredDashboardItems"
          :key="item.id"
          class="sidebar-item"
          :class="{ active: activeSubItem === item.id }"
          @click="handleSubItemClick(item.id)"
        >
          <span class="item-icon"><TechIcon :id="item.icon" :size="20" /></span>
          <span class="item-label">{{ item.label }}</span>
          <span v-if="item.meta" class="item-meta">{{ item.meta }}</span>
        </button>
      </template>

      <!-- EXPLORER Navigation Items -->
      <template v-if="currentTab === 'explorer'">
        <ExplorerSidebar @open-terminal="$emit('open-terminal', $event)" />
      </template>

      <!-- SCM Navigation Items -->
      <template v-if="currentTab === 'scm'">
        <SourceControlSidebar />
      </template>

      <!-- GUARD Navigation Items -->
      <template v-if="currentTab === 'guard'">
        <GuardSidebar />
      </template>

      <!-- TOOLS Navigation Items -->
      <template v-if="currentTab === 'tools'">
        <div class="nav-group-title">
          {{ activeScope === "user" ? "用户工具" : t("sidebar.group_tools") || "项目工具" }}
        </div>

        <!-- 1. ⚠️ 待安装 (N)：仅当 N > 0 时显示在最顶部，默认展开，琥珀色警示 -->
        <div v-if="missingTools.length > 0" class="tools-accordion-group group-missing">
          <div
            class="accordion-header missing-header"
            :class="{ collapsed: !isMissingExpanded }"
            @click="isMissingExpanded = !isMissingExpanded"
            :title="t('sidebar.tools_missing_hint') || '工作区已配置但尚未在本地安装的工具'"
          >
            <div class="accordion-header-left">
              <span
                class="codicon twistie"
                :class="isMissingExpanded ? 'codicon-chevron-down' : 'codicon-chevron-right'"
              ></span>
              <span class="accordion-status-icon warning">⚠️</span>
              <span class="accordion-title">{{ t("sidebar.tools_missing") || "待安装" }}</span>
            </div>
            <span class="accordion-count badge-warning">{{ missingTools.length }}</span>
          </div>

          <div v-show="isMissingExpanded" class="accordion-body">
            <button
              v-for="tool in missingTools"
              :key="tool.name"
              class="sidebar-item item-missing"
              :class="{ active: activeSubItem === tool.name || activeSubItem === tool.id }"
              @click="handleSubItemClick(tool.id || tool.name)"
              @contextmenu.prevent="handleToolContextMenu($event, tool)"
            >
              <span class="item-icon"><TechIcon :id="tool.id || tool.name" :size="20" /></span>
              <div class="item-dual">
                <span class="item-label">{{ tool.name }}</span>
                <span class="item-sub text-warning">{{ tool.version }}</span>
              </div>
              <span class="status-indicator-badge warning" :title="t('tools.status.missing') || '待安装'">
                <span class="codicon codicon-warning"></span>
              </span>
            </button>
          </div>
        </div>

        <!-- 2. 🟢 可升级 (N)：仅当 N > 0 时展示，默认展开，绿色升级小圆点 -->
        <div v-if="upgradableTools.length > 0" class="tools-accordion-group group-upgradable">
          <div
            class="accordion-header upgradable-header"
            :class="{ collapsed: !isUpgradableExpanded }"
            @click="isUpgradableExpanded = !isUpgradableExpanded"
            :title="t('sidebar.tools_upgradable_hint') || '发现更高可用版本，支持一键升级'"
          >
            <div class="accordion-header-left">
              <span
                class="codicon twistie"
                :class="isUpgradableExpanded ? 'codicon-chevron-down' : 'codicon-chevron-right'"
              ></span>
              <span class="accordion-status-dot update">●</span>
              <span class="accordion-title">{{ t("sidebar.tools_upgradable") || "可升级" }}</span>
            </div>
            <span class="accordion-count badge-success">{{ upgradableTools.length }}</span>
          </div>

          <div v-show="isUpgradableExpanded" class="accordion-body">
            <button
              v-for="tool in upgradableTools"
              :key="tool.name"
              class="sidebar-item item-upgradable"
              :class="{ active: activeSubItem === tool.name || activeSubItem === tool.id }"
              @click="handleSubItemClick(tool.id || tool.name)"
              @contextmenu.prevent="handleToolContextMenu($event, tool)"
            >
              <span class="item-icon"><TechIcon :id="tool.id || tool.name" :size="20" /></span>
              <div class="item-dual">
                <span class="item-label">{{ tool.name }}</span>
                <span class="item-sub">{{ tool.version }}</span>
              </div>
              <span
                class="status-indicator-dot update"
                :title="
                  tool.upgradableVersion
                    ? `可升级至 v${tool.upgradableVersion}`
                    : t('tools.actions.upgrade') || '可升级'
                "
                >●</span
              >
            </button>
          </div>
        </div>

        <!-- 3. 已就绪 / 已安装 (N)：所有正常就绪的工具，支持折叠/展开 -->
        <div class="tools-accordion-group group-ready">
          <div
            class="accordion-header ready-header"
            :class="{ collapsed: !isReadyExpanded }"
            @click="isReadyExpanded = !isReadyExpanded"
          >
            <div class="accordion-header-left">
              <span
                class="codicon twistie"
                :class="isReadyExpanded ? 'codicon-chevron-down' : 'codicon-chevron-right'"
              ></span>
              <span class="accordion-title">{{ t("sidebar.tools_ready") || "已就绪" }}</span>
            </div>
            <span class="accordion-count badge-neutral">{{ readyTools.length }}</span>
          </div>

          <div v-show="isReadyExpanded" class="accordion-body">
            <button
              v-for="tool in readyTools"
              :key="tool.name"
              class="sidebar-item"
              :class="{ active: activeSubItem === tool.name || activeSubItem === tool.id }"
              @click="handleSubItemClick(tool.id || tool.name)"
              @contextmenu.prevent="handleToolContextMenu($event, tool)"
            >
              <span class="item-icon"><TechIcon :id="tool.id || tool.name" :size="20" /></span>
              <div class="item-dual">
                <span class="item-label">{{ tool.name }}</span>
                <span class="item-sub">{{ tool.version }}</span>
              </div>
            </button>
          </div>
        </div>

        <!-- Empty State when searching -->
        <div v-if="allFilteredTools.length === 0" class="sidebar-empty-state">
          <span class="empty-text">{{ t("common.no_results") || "未找到匹配工具" }}</span>
        </div>
      </template>

      <!-- TASKS Navigation Items -->
      <template v-if="currentTab === 'tasks'">
        <div class="nav-group-title">
          {{
            activeScope === "user" ? "用户任务 (~/.unirtm.toml)" : t("sidebar.group_tasks") || "项目任务 (.unirtm.toml)"
          }}
        </div>
        <button
          v-for="task in filteredTaskItems"
          :key="task.name"
          class="sidebar-item"
          :class="{ active: activeSubItem === task.name }"
          @click="handleSubItemClick(task.name)"
          @contextmenu.prevent="handleTaskContextMenu($event, task)"
        >
          <span class="item-icon"><TechIcon id="workbench:tasks" :size="20" /></span>
          <div class="item-dual">
            <span class="item-label">{{ task.name }}</span>
            <span class="item-sub">{{ task.cmd }}</span>
          </div>
          <span class="task-play-hint" @click.stop="$emit('run-task', task.name)">▶</span>
        </button>
      </template>

      <!-- ENVIRONMENTS Navigation Items -->
      <template v-if="currentTab === 'environments'">
        <div class="nav-group-title">
          {{
            activeScope === "user"
              ? t("environments.scope_user") || "用户"
              : t("environments.scope_workspace") || "工作区"
          }}
        </div>
        <button
          v-for="env in filteredEnvItems"
          :key="env.name"
          class="sidebar-item"
          :class="{ active: activeSubItem === env.name }"
          @click="handleSubItemClick(env.name)"
          @contextmenu.prevent="handleEnvContextMenu($event, env)"
        >
          <span class="item-icon">
            <TechIcon
              :id="env.isSensitive ? 'workbench:key' : 'workbench:variable'"
              :size="18"
              :style="{ color: env.isSensitive ? '#f59e0b' : 'inherit' }"
            />
          </span>
          <div class="item-dual">
            <span class="item-label font-mono">{{ env.name }}</span>
            <span class="item-sub font-mono">{{ env.isSensitive ? "••••••••" : env.value }}</span>
          </div>
          <span
            v-if="env.isSensitive"
            class="badge-installed"
            style="
              background: rgba(245, 158, 11, 0.15);
              color: #f59e0b;
              border-color: rgba(245, 158, 11, 0.3);
              display: inline-flex;
              align-items: center;
              padding: 2px 4px;
            "
            title="敏感字段"
          >
            <TechIcon id="workbench:lock" :size="11" />
          </span>
        </button>
      </template>

      <!-- MARKETPLACE Navigation Items -->
      <template v-if="currentTab === 'marketplace'">
        <div class="nav-group-title">{{ t("sidebar.group_marketplace") || "MARKETPLACE" }}</div>
        <button
          class="sidebar-item"
          :class="{ active: activeSubItem === 'themes' || (!activeSubItem && currentTab === 'marketplace') }"
          @click="handleSubItemClick('themes')"
        >
          <span class="item-icon"><TechIcon id="workbench:marketplace" :size="20" /></span>
          <div class="item-dual">
            <span class="item-label">{{ t("settings.theme_tab_marketplace") || "主题与图标" }}</span>
            <span class="item-sub">Open VSX 扩展源</span>
          </div>
        </button>

        <button
          class="sidebar-item"
          :class="{ active: activeSubItem === 'plugins' }"
          @click="handleSubItemClick('plugins')"
        >
          <span class="item-icon"><TechIcon id="workbench:tools" :size="20" /></span>
          <div class="item-dual">
            <span class="item-label">{{ t("settings.cat_plugins") || "功能插件" }}</span>
            <span class="item-sub">敬请期待</span>
          </div>
          <span class="badge-soon" style="font-size: 0.6rem; padding: 1px 4px">开发中</span>
        </button>
      </template>

      <!-- DOCTOR Navigation Items -->
      <template v-if="currentTab === 'doctor'">
        <div class="nav-group-title">{{ t("sidebar.group_doctor") || "系统诊断" }}</div>
        <button
          v-for="check in filteredDoctorChecks"
          :key="check.id"
          class="sidebar-item"
          :class="{ active: activeSubItem === check.id }"
          @click="handleDoctorItemClick(check.id)"
        >
          <span class="item-icon"><TechIcon :id="check.icon" :size="20" /></span>
          <span class="item-label">{{ check.label }}</span>
          <span class="status-indicator-dot" :class="check.status">●</span>
        </button>
      </template>

      <!-- SETTINGS Navigation Items -->
      <template v-if="currentTab === 'settings'">
        <button
          v-for="item in settingsNavItems"
          :key="item.id"
          class="sidebar-item"
          :class="{ active: activeSubItem === item.id }"
          @click="handleSubItemClick(item.id)"
        >
          <span class="item-icon"><TechIcon :id="item.icon" :size="20" /></span>
          <span class="item-label">{{ item.label }}</span>
        </button>
      </template>

      <!-- ABOUT Navigation Items -->
      <template v-if="currentTab === 'about'">
        <div class="nav-group-title">ABOUT UNIRTM</div>
        <div class="sidebar-item active">
          <span class="item-icon"><TechIcon id="workbench:about" :size="20" /></span>
          <span class="item-label">{{ t("about.title") || "关于" }}</span>
        </div>
      </template>
    </div>

    <!-- Draggable Resizer Sash on the right border -->
    <div
      class="sidebar-resizer-sash"
      :class="{ active: isDragging }"
      @pointerdown="handlePointerDown"
      @dblclick="resetWidth"
      title="双击恢复默认宽度，拖拽调节大小"
    >
      <div class="sash-indicator"></div>
    </div>

    <!-- Dragging overlay to prevent cursor/pointer capture interference -->
    <div v-if="isDragging" class="sidebar-drag-overlay"></div>

    <!-- Tool Item Context Menu -->
    <ToolContextMenu
      :visible="isToolContextMenuVisible"
      :position="toolContextMenuPosition"
      :tool="toolContextMenuTarget"
      @close="isToolContextMenuVisible = false"
      @action="handleToolContextMenuAction"
    />

    <!-- Task Item Context Menu -->
    <TaskContextMenu
      :visible="isTaskContextMenuVisible"
      :position="taskContextMenuPosition"
      :task="taskContextMenuTarget"
      @close="isTaskContextMenuVisible = false"
      @action="handleTaskContextMenuAction"
    />

    <!-- Environment Item Context Menu -->
    <EnvironmentContextMenu
      :visible="isEnvContextMenuVisible"
      :position="envContextMenuPosition"
      :env="envContextMenuTarget"
      @close="isEnvContextMenuVisible = false"
      @action="handleEnvContextMenuAction"
    />
  </aside>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount } from "vue";
import { t } from "../../i18n";
import TechIcon from "../common/TechIcon.vue";
import ExplorerSidebar from "../explorer/ExplorerSidebar.vue";
import SourceControlSidebar from "../scm/SourceControlSidebar.vue";
import GuardSidebar from "../guard/GuardSidebar.vue";
import ToolContextMenu, { type ToolMenuActionType } from "../tools/ToolContextMenu.vue";
import TaskContextMenu, { type TaskMenuActionType } from "../tasks/TaskContextMenu.vue";
import EnvironmentContextMenu, { type EnvMenuActionType } from "../environments/EnvironmentContextMenu.vue";
import { useFileTree } from "../../composables/useFileTree";
import { useToolUpdates } from "../../composables/useToolUpdates";
import { useToolScope } from "../../composables/useToolScope";
import { useWorkspaceBus } from "../../composables/useWorkspaceBus";
import { categorizeTools } from "../../utils/toolStatus";

const SIDEBAR_WIDTH_KEY = "unirtm_sidebar_width";
const DEFAULT_SIDEBAR_WIDTH = 230;
const MIN_SIDEBAR_WIDTH = 160;
const MAX_SIDEBAR_WIDTH = 550;

function getInitialSidebarWidth(): number {
  if (typeof localStorage === "undefined") return DEFAULT_SIDEBAR_WIDTH;
  try {
    const saved = localStorage.getItem(SIDEBAR_WIDTH_KEY);
    if (saved) {
      const parsed = Number(saved);
      if (!isNaN(parsed) && parsed >= MIN_SIDEBAR_WIDTH && parsed <= MAX_SIDEBAR_WIDTH) {
        return parsed;
      }
    }
  } catch {}
  return DEFAULT_SIDEBAR_WIDTH;
}

const sidebarWidth = ref(getInitialSidebarWidth());
const isDragging = ref(false);

function resetWidth() {
  sidebarWidth.value = DEFAULT_SIDEBAR_WIDTH;
  try {
    localStorage.setItem(SIDEBAR_WIDTH_KEY, String(DEFAULT_SIDEBAR_WIDTH));
  } catch {}
}

let startX = 0;
let startWidth = 0;

function handlePointerDown(e: PointerEvent) {
  e.preventDefault();
  isDragging.value = true;
  startX = e.clientX;
  startWidth = sidebarWidth.value;

  document.body.style.cursor = "col-resize";
  document.body.style.userSelect = "none";

  const onPointerMove = (ev: PointerEvent) => {
    if (!isDragging.value) return;
    const delta = ev.clientX - startX;
    const maxAllowed = Math.min(MAX_SIDEBAR_WIDTH, Math.floor(window.innerWidth * 0.6));
    const newWidth = Math.max(MIN_SIDEBAR_WIDTH, Math.min(maxAllowed, startWidth + delta));
    sidebarWidth.value = Math.round(newWidth);
  };

  const onPointerUp = () => {
    if (isDragging.value) {
      isDragging.value = false;
      document.body.style.cursor = "";
      document.body.style.userSelect = "";
      try {
        localStorage.setItem(SIDEBAR_WIDTH_KEY, String(sidebarWidth.value));
      } catch {}
    }
    window.removeEventListener("pointermove", onPointerMove);
    window.removeEventListener("pointerup", onPointerUp);
    window.removeEventListener("pointercancel", onPointerUp);
  };

  window.addEventListener("pointermove", onPointerMove);
  window.addEventListener("pointerup", onPointerUp);
  window.addEventListener("pointercancel", onPointerUp);
}

const { syncFromToolsList, updateResults } = useToolUpdates();
const { activeScope } = useToolScope();
const {
  triggerToolAction,
  onToolsChanged,
  triggerTaskAction,
  onTasksChanged,
  onDoctorChanged,
  triggerEnvAction,
  onEnvsChanged,
} = useWorkspaceBus();

const isToolContextMenuVisible = ref(false);
const toolContextMenuPosition = ref({ x: 0, y: 0 });
const toolContextMenuTarget = ref<any | null>(null);

function handleToolContextMenu(event: MouseEvent, tool: any) {
  toolContextMenuTarget.value = tool;
  toolContextMenuPosition.value = { x: event.clientX, y: event.clientY };
  isToolContextMenuVisible.value = true;
}

function handleToolContextMenuAction(action: ToolMenuActionType, tool: any) {
  isToolContextMenuVisible.value = false;
  const toolId = tool.id || tool.name;
  triggerToolAction({
    action,
    toolId,
    toolName: tool.name,
    tool: tool.raw || tool,
  });
}

const isTaskContextMenuVisible = ref(false);
const taskContextMenuPosition = ref({ x: 0, y: 0 });
const taskContextMenuTarget = ref<SidebarTaskItem | null>(null);

function handleTaskContextMenu(event: MouseEvent, task: SidebarTaskItem) {
  taskContextMenuTarget.value = task;
  taskContextMenuPosition.value = { x: event.clientX, y: event.clientY };
  isTaskContextMenuVisible.value = true;
}

function handleTaskContextMenuAction(action: TaskMenuActionType, task: SidebarTaskItem) {
  isTaskContextMenuVisible.value = false;
  triggerTaskAction({
    action,
    taskName: task.name,
    command: task.cmd,
    taskCmd: task.cmd,
    task: task,
  });
}

const isEnvContextMenuVisible = ref(false);
const envContextMenuPosition = ref({ x: 0, y: 0 });
const envContextMenuTarget = ref<SidebarEnvItem | null>(null);

function handleEnvContextMenu(event: MouseEvent, env: SidebarEnvItem) {
  envContextMenuTarget.value = env;
  envContextMenuPosition.value = { x: event.clientX, y: event.clientY };
  isEnvContextMenuVisible.value = true;
}

function handleEnvContextMenuAction(action: EnvMenuActionType, env: SidebarEnvItem) {
  isEnvContextMenuVisible.value = false;
  if (action === "copy-value") {
    try {
      if (navigator.clipboard) {
        navigator.clipboard.writeText(env.value);
      }
    } catch {}
  } else if (action === "copy-name") {
    try {
      if (navigator.clipboard) {
        navigator.clipboard.writeText(env.name);
      }
    } catch {}
  }
  triggerEnvAction({
    action,
    envName: env.name,
    env,
  });
}

const props = defineProps<{
  isOpen: boolean;
  currentTab: string;
  activeSubItem: string;
}>();

const emit = defineEmits<{
  (e: "toggle-collapse"): void;
  (e: "select-sub-item", id: string): void;
  (e: "run-task", taskName: string): void;
  (e: "open-terminal", dirPath: string): void;
}>();

function handleSubItemClick(id: string) {
  if (props.activeSubItem === id) {
    emit("select-sub-item", "");
  } else {
    emit("select-sub-item", id);
  }
}

const searchQuery = ref("");

const { searchQuery: fileTreeSearchQuery } = useFileTree();

watch(searchQuery, (newVal) => {
  if (props.currentTab === "explorer") {
    fileTreeSearchQuery.value = newVal;
  }
});

const headerTitle = computed(() => {
  switch (props.currentTab) {
    case "dashboard":
      return t("nav.dashboard") || "DASHBOARD";
    case "explorer":
      return t("nav.explorer") || "EXPLORER";
    case "scm":
      return t("nav.scm") || "SOURCE CONTROL";
    case "guard":
      return t("nav.guard") || "QUALITY GUARD";
    case "tools":
      return t("nav.tools") || "TOOLCHAINS";
    case "tasks":
      return t("nav.tasks") || "TASK RUNNER";
    case "environments":
      return t("nav.environments") || "ENVIRONMENTS";
    case "marketplace":
      return t("nav.marketplace") || "MARKETPLACE";
    case "doctor":
      return t("nav.doctor") || "DOCTOR";
    case "settings":
      return t("settings.title") || "SETTINGS";
    case "about":
      return t("about.title") || "ABOUT";
    default:
      return "EXPLORER";
  }
});

const settingsNavItems = computed(() => [
  { id: "general", icon: "workbench:settings", label: t("settings.cat_general") || "常规" },
  { id: "network", icon: "workbench:runtimes", label: t("settings.cat_network") || "网络" },
  { id: "plugins", icon: "workbench:marketplace", label: t("settings.cat_plugins") || "扩展" },
  { id: "security", icon: "workbench:lock", label: t("settings.cat_security") || "安全" },
]);

const searchPlaceholder = computed(() => {
  return t("common.filter") || "Filter...";
});

// Dashboard mock items
const dashboardItems = [
  { id: "overview", icon: "workbench:dashboard", label: "Project Status", meta: "100%" },
  { id: "specs", icon: "go", label: "Runtime Specs", meta: "Go 1.24" },
  { id: "connectivity", icon: "workbench:runtimes", label: "Network & Mirrors", meta: "OK" },
  { id: "actions", icon: "workbench:tasks", label: "Quick Actions", meta: "3" },
];

const filteredDashboardItems = computed(() => {
  if (!searchQuery.value) return dashboardItems;
  return dashboardItems.filter((i) => i.label.toLowerCase().includes(searchQuery.value.toLowerCase()));
});

// Tools items (Synchronized with active toolchains and runtimes by Scope)
const isMissingExpanded = ref(true);
const isUpgradableExpanded = ref(true);
const isReadyExpanded = ref(true);

const rawTools = ref<any[]>([
  { id: "node", name: "Node.js", active_version: "26.8.1", global_version: "26.8.1", installed_versions: ["26.8.1"] },
  { id: "go", name: "Go", active_version: "1.27.2", global_version: "", installed_versions: ["1.27.2"] },
  { id: "python", name: "Python", active_version: "3.14.7", global_version: "", installed_versions: ["3.14.7"] },
]);

const categorizedTools = computed(() => {
  return categorizeTools(
    rawTools.value,
    activeScope.value as "user" | "workspace",
    updateResults.value,
    searchQuery.value
  );
});

const missingTools = computed(() => categorizedTools.value.missing);
const upgradableTools = computed(() => categorizedTools.value.upgradable);
const readyTools = computed(() => categorizedTools.value.ready);
const allFilteredTools = computed(() => categorizedTools.value.allFiltered);

// Tasks items dynamically synchronized with .unirtm.toml
interface SidebarTaskItem {
  name: string;
  cmd: string;
  desc?: string;
}

const rawTasks = ref<SidebarTaskItem[]>([]);

async function fetchSidebarTasks() {
  try {
    const res = await fetch(`/api/tasks?scope=${activeScope.value}`);
    if (res.ok) {
      const data = await res.json();
      if (Array.isArray(data.tasks)) {
        rawTasks.value = data.tasks.map((t: any) => ({
          name: t.name,
          cmd: t.command || t.description || "",
          desc: t.description || "",
        }));
      }
    }
  } catch {
    // dev mode fallback
  }
}

const filteredTaskItems = computed(() => {
  if (!searchQuery.value) return rawTasks.value;
  const q = searchQuery.value.toLowerCase().trim();
  return rawTasks.value.filter(
    (t) =>
      t.name.toLowerCase().includes(q) ||
      t.cmd.toLowerCase().includes(q) ||
      (t.desc && t.desc.toLowerCase().includes(q))
  );
});

watch(
  () => activeScope.value,
  () => {
    fetchSidebarTasks();
    fetchSidebarEnvs();
  }
);

watch(
  () => props.currentTab,
  (newTab) => {
    if (newTab === "tasks") {
      fetchSidebarTasks();
    } else if (newTab === "environments") {
      fetchSidebarEnvs();
    } else if (newTab === "doctor") {
      fetchSidebarDoctorStatus();
    }
  }
);

// Environment items dynamically synchronized with [env]
interface SidebarEnvItem {
  name: string;
  value: string;
  isSensitive: boolean;
  scope?: "workspace" | "user";
  isOverride?: boolean;
}

const rawEnvs = ref<SidebarEnvItem[]>([]);

async function fetchSidebarEnvs() {
  try {
    const res = await fetch(`/api/env?scope=${activeScope.value}`);
    if (res.ok) {
      const data = await res.json();
      if (Array.isArray(data.env)) {
        rawEnvs.value = data.env.map((e: any) => ({
          name: e.name,
          value: e.value,
          isSensitive: Boolean(e.is_sensitive || e.isSensitive),
          isOverride: Boolean(e.is_override || e.isOverride),
          scope: e.scope,
        }));
      }
    }
  } catch {}
}

const filteredEnvItems = computed(() => {
  if (!searchQuery.value) return rawEnvs.value;
  const q = searchQuery.value.toLowerCase().trim();
  return rawEnvs.value.filter((e) => e.name.toLowerCase().includes(q) || e.value.toLowerCase().includes(q));
});

// Doctor checks - 真实对应 UniRTM 全景诊断 5 大核心板块
interface DoctorCheckItem {
  id: string;
  icon: string;
  label: string;
  status: "success" | "warning" | "error";
}

const doctorReport = ref<any>(null);

async function fetchSidebarDoctorStatus() {
  try {
    const res = await fetch("/api/doctor");
    const contentType = res.headers.get("content-type") || "";
    if (res.ok && contentType.includes("json")) {
      doctorReport.value = await res.json();
    }
  } catch {}
}

const doctorChecks = computed<DoctorCheckItem[]>(() => {
  const r = doctorReport.value;
  return [
    {
      id: "core",
      icon: "workbench:dashboard",
      label: t("doctor.dim_core") || "核心与执行环境",
      status: r ? (r.core?.shims_on_path ? "success" : "warning") : "success",
    },
    {
      id: "network",
      icon: "workbench:runtimes",
      label: t("doctor.dim_network") || "网络联通与 API 配额",
      status: r ? (r.health?.github_connected ? "success" : "error") : "success",
    },
    {
      id: "directories",
      icon: "workbench:folder",
      label: t("doctor.dim_directories") || "磁盘占用与存储路径",
      status: "success",
    },
    {
      id: "database",
      icon: "workbench:settings",
      label: t("doctor.dim_database") || "状态数据库与配置",
      status: r ? (r.health?.database_ok ? "success" : "error") : "success",
    },
  ];
});

const filteredDoctorChecks = computed(() => {
  if (!searchQuery.value) return doctorChecks.value;
  const q = searchQuery.value.toLowerCase().trim();
  return doctorChecks.value.filter((c) => c.label.toLowerCase().includes(q) || c.id.toLowerCase().includes(q));
});

function handleDoctorItemClick(id: string) {
  handleSubItemClick(id);
  const target = document.getElementById("doctor-section-" + id);
  if (target) {
    target.scrollIntoView({ behavior: "smooth", block: "center" });
    target.classList.add("highlight-pulse");
    setTimeout(() => {
      target.classList.remove("highlight-pulse");
    }, 1500);
  }
}

async function fetchSidebarToolsStatus() {
  try {
    const res = await fetch("/api/tools/status");
    if (res.ok) {
      const data = await res.json();
      const list = data.tools || data.runtimes;
      if (list && Array.isArray(list)) {
        syncFromToolsList(list);
        rawTools.value = list;
      }
    }
  } catch {
    // dev mode fallback
  }
}

let unsubToolsChanged: (() => void) | null = null;
let unsubTasksChanged: (() => void) | null = null;
let unsubDoctorChanged: (() => void) | null = null;
let unsubEnvsChanged: (() => void) | null = null;

onMounted(async () => {
  fetchSidebarTasks();
  fetchSidebarEnvs();
  fetchSidebarToolsStatus();
  fetchSidebarDoctorStatus();
  unsubToolsChanged = onToolsChanged(fetchSidebarToolsStatus);
  unsubTasksChanged = onTasksChanged(fetchSidebarTasks);
  unsubDoctorChanged = onDoctorChanged(fetchSidebarDoctorStatus);
  unsubEnvsChanged = onEnvsChanged(fetchSidebarEnvs);
});

onBeforeUnmount(() => {
  if (unsubToolsChanged) {
    unsubToolsChanged();
    unsubToolsChanged = null;
  }
  if (unsubTasksChanged) {
    unsubTasksChanged();
    unsubTasksChanged = null;
  }
  if (unsubDoctorChanged) {
    unsubDoctorChanged();
    unsubDoctorChanged = null;
  }
  if (unsubEnvsChanged) {
    unsubEnvsChanged();
    unsubEnvsChanged = null;
  }
});
</script>

<style scoped>
.sidebar-panel {
  display: flex;
  flex-direction: column;
  position: relative;
  height: 100%;
  background: var(--card-bg);
  border-right: 1px solid var(--card-border);
  box-sizing: border-box;
  user-select: none;
  -webkit-user-select: none;
  overflow: visible;
  backdrop-filter: blur(10px);
  flex-shrink: 0;
}

.sidebar-panel.is-resizing {
  user-select: none;
}

.sidebar-resizer-sash {
  position: absolute;
  top: 0;
  right: -3px;
  width: 6px;
  height: 100%;
  cursor: col-resize;
  z-index: 25;
  display: flex;
  align-items: center;
  justify-content: center;
  touch-action: none;
}

.sidebar-resizer-sash .sash-indicator {
  width: 2px;
  height: 100%;
  background: transparent;
  transition: background 0.15s ease;
}

.sidebar-resizer-sash:hover .sash-indicator,
.sidebar-resizer-sash.active .sash-indicator {
  background: var(--accent-cyan, #00e5ff);
}

.sidebar-drag-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 99999;
  cursor: col-resize;
  background: transparent;
}

.sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.65rem 0.85rem;
  border-bottom: 1px solid var(--card-border);
}

.panel-title {
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  color: var(--text-muted);
}

.panel-tool-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 0.15rem 0.35rem;
  border-radius: 4px;
  font-size: 0.75rem;
  transition: all 0.2s ease;
}

.panel-tool-btn:hover {
  background: var(--input-bg);
  color: var(--text-main);
}

.sidebar-search {
  padding: 0.5rem 0.65rem;
}

.sidebar-search-input {
  width: 100%;
  padding: 0.32rem 0.6rem;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  border-radius: 6px;
  color: var(--text-main);
  font-size: 0.78rem;
  box-sizing: border-box;
  outline: none;
  transition: border-color 0.2s ease;
}

.sidebar-search-input:focus {
  border-color: var(--card-border-active);
}

.sidebar-content-list {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  scrollbar-gutter: stable;
  padding: 0.2rem 0.5rem;
}

.nav-group-title {
  font-size: 0.68rem;
  font-weight: 700;
  color: var(--text-muted);
  padding: 0.45rem 0.45rem 0.25rem;
  letter-spacing: 0.05em;
  opacity: 0.8;
}

.sidebar-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  width: 100%;
  padding: 0.38rem 0.5rem;
  background: transparent;
  border: none;
  border-radius: 6px;
  color: var(--text-main);
  cursor: pointer;
  font-size: 0.8rem;
  text-align: left;
  transition: all 0.15s ease;
  box-sizing: border-box;
}

.sidebar-item:hover {
  background-color: rgba(128, 128, 128, 0.12);
}

:root[data-theme="dark"] .sidebar-item:hover,
[data-theme="dark"] .sidebar-item:hover {
  background-color: rgba(255, 255, 255, 0.08);
}

:root[data-theme="light"] .sidebar-item:hover,
[data-theme="light"] .sidebar-item:hover {
  background-color: rgba(0, 0, 0, 0.06);
}

.sidebar-item.active {
  background: rgba(0, 229, 255, 0.12);
  color: var(--accent-cyan);
  font-weight: 600;
}

[data-theme="light"] .sidebar-item.active,
:root[data-theme="light"] .sidebar-item.active {
  background: rgba(2, 132, 199, 0.12);
  color: #0369a1;
}

[data-theme="light"] .sidebar-item.active .item-sub,
:root[data-theme="light"] .sidebar-item.active .item-sub {
  color: #0284c7;
}

.item-icon {
  font-size: 0.95rem;
  line-height: 1;
}

.item-label {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.item-dual {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.item-sub {
  font-size: 0.7rem;
  color: var(--text-muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.item-meta {
  font-size: 0.7rem;
  color: var(--text-muted);
  background: var(--input-bg);
  padding: 0.1rem 0.35rem;
  border-radius: 4px;
}

.status-indicator-dot {
  font-size: 0.65rem;
  line-height: 1;
}

.status-indicator-dot.success,
.status-indicator-dot.update {
  color: var(--success, #10b981);
}

.status-indicator-dot.warning {
  color: #f59e0b;
}

.status-indicator-dot.error {
  color: #ef4444;
}

[data-theme="light"] .status-indicator-dot.success,
[data-theme="light"] .status-indicator-dot.update,
:root[data-theme="light"] .status-indicator-dot.update {
  color: var(--success, #059669);
}

.task-play-hint {
  font-size: 0.7rem;
  padding: 0.15rem 0.35rem;
  border-radius: 4px;
  color: var(--text-muted);
  transition: all 0.15s ease;
}

.task-play-hint:hover {
  background: rgba(0, 229, 255, 0.2);
  color: var(--accent-cyan);
}

.badge-installed {
  font-size: 0.75rem;
  font-weight: 700;
  color: var(--success);
}

/* Tools Accordion Groups */
.tools-accordion-group {
  margin-bottom: 6px;
}

.accordion-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 4px 8px;
  cursor: pointer;
  user-select: none;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: var(--text-muted);
  transition:
    background 0.15s ease,
    color 0.15s ease;
}

.accordion-header:hover {
  background-color: rgba(128, 128, 128, 0.12);
  color: var(--text-main);
}

:root[data-theme="dark"] .accordion-header:hover,
[data-theme="dark"] .accordion-header:hover {
  background-color: rgba(255, 255, 255, 0.08);
}

:root[data-theme="light"] .accordion-header:hover,
[data-theme="light"] .accordion-header:hover {
  background-color: rgba(0, 0, 0, 0.06);
}

.accordion-header-left {
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
}

.accordion-header .twistie {
  font-size: 10px;
  color: var(--text-muted);
  transition: transform 0.15s ease;
  width: 14px;
  text-align: center;
}

.accordion-status-icon.warning {
  font-size: 11px;
  line-height: 1;
}

.accordion-status-dot.update {
  font-size: 10px;
  color: #22c55e;
  line-height: 1;
}

.accordion-title {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.accordion-count {
  font-size: 10px;
  font-weight: 700;
  padding: 1px 6px;
  border-radius: 10px;
  line-height: 1.2;
}

.badge-warning {
  background: rgba(245, 158, 11, 0.18);
  color: #f59e0b;
  border: 1px solid rgba(245, 158, 11, 0.35);
}

.badge-success {
  background: rgba(34, 197, 94, 0.18);
  color: #22c55e;
  border: 1px solid rgba(34, 197, 94, 0.35);
}

.badge-neutral {
  background: var(--input-bg);
  color: var(--text-muted);
  border: 1px solid var(--card-border);
}

.accordion-body {
  display: flex;
  flex-direction: column;
  gap: 2px;
  margin-top: 2px;
}

.status-indicator-badge.warning {
  display: inline-flex;
  align-items: center;
  color: #f59e0b;
  font-size: 13px;
}

.text-warning {
  color: #f59e0b !important;
}

.sidebar-empty-state {
  padding: 16px 12px;
  text-align: center;
  font-size: 12px;
  color: var(--text-muted);
}
</style>
