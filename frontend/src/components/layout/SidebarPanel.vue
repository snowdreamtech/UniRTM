<template>
  <aside v-if="isOpen" class="sidebar-panel" id="app-sidebar-panel">
    <!-- Panel Header -->
    <div class="sidebar-header">
      <span class="panel-title">{{ headerTitle }}</span>
      <div class="panel-tools">
        <button
          class="panel-tool-btn"
          title="Collapse Sidebar"
          @click="$emit('toggle-collapse')"
        >
          <span>◀</span>
        </button>
      </div>
    </div>

    <!-- Filter input -->
    <div class="sidebar-search">
      <input
        v-model="searchQuery"
        type="text"
        class="sidebar-search-input"
        :placeholder="searchPlaceholder"
      />
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



      <!-- TOOLS Navigation Items -->
      <template v-if="currentTab === 'tools'">
        <div class="nav-group-title">{{ t("sidebar.group_tools") || "PROJECT TOOLS" }}</div>
        <button
          v-for="tool in filteredToolItems"
          :key="tool.name"
          class="sidebar-item"
          :class="{ active: activeSubItem === tool.name || activeSubItem === tool.id }"
          @click="handleSubItemClick(tool.id || tool.name)"
        >
          <span class="item-icon"><TechIcon :id="tool.id || tool.name" :size="20" /></span>
          <div class="item-dual">
            <span class="item-label">{{ tool.name }}</span>
            <span class="item-sub">{{ tool.version }}</span>
          </div>
          <span class="status-indicator-dot" :class="tool.status">●</span>
        </button>
      </template>

      <!-- TASKS Navigation Items -->
      <template v-if="currentTab === 'tasks'">
        <div class="nav-group-title">{{ t("sidebar.group_tasks") || "UNIRTM TASKS (.unirtm.toml)" }}</div>
        <button
          v-for="task in filteredTaskItems"
          :key="task.name"
          class="sidebar-item"
          :class="{ active: activeSubItem === task.name }"
          @click="handleSubItemClick(task.name)"
        >
          <span class="item-icon"><TechIcon id="workbench:tasks" :size="20" /></span>
          <div class="item-dual">
            <span class="item-label">{{ task.name }}</span>
            <span class="item-sub">{{ task.cmd }}</span>
          </div>
          <span class="task-play-hint" @click.stop="$emit('run-task', task.name)">▶</span>
        </button>
      </template>

      <!-- MARKETPLACE Navigation Items -->
      <template v-if="currentTab === 'marketplace'">
        <div class="nav-group-title">{{ t("sidebar.group_marketplace") || "FEATURED EXTENSIONS" }}</div>
        <button
          v-for="ext in filteredExtensions"
          :key="ext.id"
          class="sidebar-item"
          :class="{ active: activeSubItem === ext.id }"
          @click="handleSubItemClick(ext.id)"
        >
          <span class="item-icon"><TechIcon :id="ext.icon" :size="20" /></span>
          <div class="item-dual">
            <span class="item-label">{{ ext.name }}</span>
            <span class="item-sub">{{ ext.desc }}</span>
          </div>
          <span v-if="ext.installed" class="badge-installed">✓</span>
        </button>
      </template>

      <!-- DOCTOR Navigation Items -->
      <template v-if="currentTab === 'doctor'">
        <div class="nav-group-title">{{ t("sidebar.group_doctor") || "DIAGNOSTIC CHECKS" }}</div>
        <button
          v-for="check in doctorChecks"
          :key="check.id"
          class="sidebar-item"
          :class="{ active: activeSubItem === check.id }"
          @click="handleSubItemClick(check.id)"
        >
          <span class="item-icon"><TechIcon :id="check.icon" :size="20" /></span>
          <span class="item-label">{{ check.label }}</span>
          <span class="status-indicator-dot success">●</span>
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
  </aside>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { t } from "../../i18n";
import TechIcon from "../common/TechIcon.vue";

const props = defineProps<{
  isOpen: boolean;
  currentTab: string;
  activeSubItem: string;
}>();

const emit = defineEmits<{
  (e: "toggle-collapse"): void;
  (e: "select-sub-item", id: string): void;
  (e: "run-task", taskName: string): void;
}>();

function handleSubItemClick(id: string) {
  if (props.activeSubItem === id) {
    emit("select-sub-item", "");
  } else {
    emit("select-sub-item", id);
  }
}

const searchQuery = ref("");

const headerTitle = computed(() => {
  switch (props.currentTab) {
    case "dashboard":
      return t("nav.dashboard") || "DASHBOARD";
    case "tools":
      return t("nav.tools") || "TOOLCHAINS";
    case "tasks":
      return t("nav.tasks") || "TASK RUNNER";
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



// Tools items (Synchronized with active toolchains and runtimes)
const toolItems = ref([
  { id: "node", name: "Node.js", icon: "node", version: "20.18.3", status: "success" },
  { id: "go", name: "Go", icon: "go", version: "1.27.2", status: "success" },
  { id: "python", name: "Python", icon: "python", version: "3.14.7", status: "success" },
  { id: "rust", name: "Rust", icon: "rust", version: "1.85.0", status: "success" },
]);

const filteredToolItems = computed(() => {
  if (!searchQuery.value) return toolItems.value;
  const q = searchQuery.value.toLowerCase().trim();
  return toolItems.value.filter((t) => t.name.toLowerCase().includes(q) || t.id.toLowerCase().includes(q));
});

// Tasks items
const taskItems = [
  { name: "dev", cmd: "vite" },
  { name: "build", cmd: "vue-tsc && vite build" },
  { name: "test", cmd: "vitest run" },
  { name: "audit:npm", cmd: "npm audit" },
];

const filteredTaskItems = computed(() => {
  if (!searchQuery.value) return taskItems;
  return taskItems.filter((t) => t.name.toLowerCase().includes(searchQuery.value.toLowerCase()));
});

// Marketplace extensions
const extensions = [
  { id: "uniagent", icon: "workbench:doctor", name: "UniAgent", desc: "AI Troubleshooter", installed: true },
  { id: "security-suite", icon: "workbench:lock", name: "Security Suite", desc: "Trivy & Syft Scans", installed: true },
  { id: "rust-provider", icon: "rust", name: "Rust Provider", desc: "Cargo & Rust Toolchain", installed: false },
  { id: "datascience", icon: "python", name: "DataScience Pack", desc: "Python uv & Jupyter", installed: false },
];

const filteredExtensions = computed(() => {
  if (!searchQuery.value) return extensions;
  return extensions.filter((e) => e.name.toLowerCase().includes(searchQuery.value.toLowerCase()));
});

// Doctor checks
const doctorChecks = [
  { id: "os-arch", icon: "workbench:dashboard", label: "OS & Arch Support" },
  { id: "lockfile", icon: "workbench:lock", label: "Lockfile Hygiene" },
  { id: "env-sandbox", icon: "workbench:config", label: "Environment Isolation" },
  { id: "mirror-speed", icon: "workbench:tasks", label: "Mirror Latency" },
];

onMounted(async () => {
  try {
    const res = await fetch("/api/tools/status");
    if (res.ok) {
      const data = await res.json();
      const list = data.tools || data.runtimes;
      if (list && Array.isArray(list)) {
        for (const rt of list) {
          const matched = toolItems.value.find((t) => t.id === rt.id || t.name === rt.id);
          if (matched) {
            if (rt.active_version) matched.version = rt.active_version;
            else if (rt.global_version) matched.version = rt.global_version;
          } else {
            toolItems.value.push({
              id: rt.id,
              name: rt.name,
              icon: rt.id,
              version: rt.active_version || rt.global_version || "installed",
              status: "success",
            });
          }
        }
      }
    }
  } catch {
    // dev mode fallback
  }
});
</script>

<style scoped>
.sidebar-panel {
  display: flex;
  flex-direction: column;
  width: 230px;
  min-width: 230px;
  height: 100%;
  background: var(--card-bg);
  border-right: 1px solid var(--card-border);
  box-sizing: border-box;
  user-select: none;
  -webkit-user-select: none;
  overflow: hidden;
  backdrop-filter: blur(10px);
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
  background: var(--input-bg);
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
}

.status-indicator-dot.success {
  color: var(--success);
}

.status-indicator-dot.warning {
  color: var(--warning);
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
</style>
