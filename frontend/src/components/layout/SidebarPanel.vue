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
        <div class="nav-group-title">PROJECT OVERVIEW</div>
        <button
          v-for="item in filteredDashboardItems"
          :key="item.id"
          class="sidebar-item"
          :class="{ active: activeSubItem === item.id }"
          @click="$emit('select-sub-item', item.id)"
        >
          <span class="item-icon">{{ item.icon }}</span>
          <span class="item-label">{{ item.label }}</span>
          <span v-if="item.meta" class="item-meta">{{ item.meta }}</span>
        </button>
      </template>

      <!-- TOOLS Navigation Items -->
      <template v-if="currentTab === 'tools'">
        <div class="nav-group-title">PROJECT RUNTIMES & TOOLS</div>
        <button
          v-for="tool in filteredToolItems"
          :key="tool.name"
          class="sidebar-item"
          :class="{ active: activeSubItem === tool.name }"
          @click="$emit('select-sub-item', tool.name)"
        >
          <span class="item-icon">{{ tool.icon }}</span>
          <div class="item-dual">
            <span class="item-label">{{ tool.name }}</span>
            <span class="item-sub">{{ tool.version }}</span>
          </div>
          <span class="status-indicator-dot" :class="tool.status">●</span>
        </button>
      </template>

      <!-- TASKS Navigation Items -->
      <template v-if="currentTab === 'tasks'">
        <div class="nav-group-title">UNIRTM TASKS (.unirtm.toml)</div>
        <button
          v-for="task in filteredTaskItems"
          :key="task.name"
          class="sidebar-item"
          :class="{ active: activeSubItem === task.name }"
          @click="$emit('select-sub-item', task.name)"
        >
          <span class="item-icon">⚡</span>
          <div class="item-dual">
            <span class="item-label">{{ task.name }}</span>
            <span class="item-sub">{{ task.cmd }}</span>
          </div>
          <span class="task-play-hint" @click.stop="$emit('run-task', task.name)">▶</span>
        </button>
      </template>

      <!-- MARKETPLACE Navigation Items -->
      <template v-if="currentTab === 'marketplace'">
        <div class="nav-group-title">FEATURED EXTENSIONS</div>
        <button
          v-for="ext in filteredExtensions"
          :key="ext.id"
          class="sidebar-item"
          :class="{ active: activeSubItem === ext.id }"
          @click="$emit('select-sub-item', ext.id)"
        >
          <span class="item-icon">{{ ext.icon }}</span>
          <div class="item-dual">
            <span class="item-label">{{ ext.name }}</span>
            <span class="item-sub">{{ ext.desc }}</span>
          </div>
          <span v-if="ext.installed" class="badge-installed">✓</span>
        </button>
      </template>

      <!-- DOCTOR Navigation Items -->
      <template v-if="currentTab === 'doctor'">
        <div class="nav-group-title">DIAGNOSTIC CHECKS</div>
        <button
          v-for="check in doctorChecks"
          :key="check.id"
          class="sidebar-item"
          :class="{ active: activeSubItem === check.id }"
          @click="$emit('select-sub-item', check.id)"
        >
          <span class="item-icon">{{ check.icon }}</span>
          <span class="item-label">{{ check.label }}</span>
          <span class="status-indicator-dot success">●</span>
        </button>
      </template>

      <!-- SETTINGS Navigation Items -->
      <template v-if="currentTab === 'settings'">
        <div class="nav-group-title">{{ t('settings.sections') || '设置分组' }}</div>
        <button
          v-for="item in settingsNavItems"
          :key="item.id"
          class="sidebar-item"
          :class="{ active: activeSubItem === item.id }"
          @click="$emit('select-sub-item', item.id)"
        >
          <span class="item-icon">{{ item.icon }}</span>
          <div class="item-dual">
            <span class="item-label">{{ item.label }}</span>
            <span class="item-sub">{{ item.sub }}</span>
          </div>
        </button>
      </template>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import { t } from "../../i18n";

const props = defineProps<{
  isOpen: boolean;
  currentTab: string;
  activeSubItem: string;
}>();

defineEmits<{
  (e: "toggle-collapse"): void;
  (e: "select-sub-item", id: string): void;
  (e: "run-task", taskName: string): void;
}>();

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
    default:
      return "EXPLORER";
  }
});

const settingsNavItems = computed(() => [
  { id: "common", icon: "⚙️", label: t("settings.cat_common") || "常用设置", sub: t("settings.section_appearance") || "外观与语言" },
  { id: "browser", icon: "💻", label: t("settings.cat_browser") || "系统与浏览器", sub: t("settings.scope_global_only") || "仅限用户设置" },
  { id: "network", icon: "🌐", label: t("settings.cat_network") || "网络与代理", sub: t("settings.githubProxy") || "镜像与代理" },
  { id: "plugins", icon: "🧩", label: t("settings.cat_plugins") || "插件与扩展", sub: t("settings.plugin_isolation_title") || "命名空间隔离" },
  { id: "security", icon: "🔒", label: t("settings.cat_security") || "安全与命名空间", sub: t("settings.sec_lock_title") || "并发文件锁" },
]);

const searchPlaceholder = computed(() => {
  return "Filter...";
});

// Dashboard mock items
const dashboardItems = [
  { id: "overview", icon: "📊", label: "Project Status", meta: "100%" },
  { id: "specs", icon: "💻", label: "Runtime Specs", meta: "Go 1.24" },
  { id: "connectivity", icon: "🌐", label: "Network & Mirrors", meta: "OK" },
  { id: "actions", icon: "🚀", label: "Quick Actions", meta: "3" },
];

const filteredDashboardItems = computed(() => {
  if (!searchQuery.value) return dashboardItems;
  return dashboardItems.filter((i) => i.label.toLowerCase().includes(searchQuery.value.toLowerCase()));
});

// Tools items
const toolItems = [
  { name: "node", icon: "🟢", version: "22.14.0", status: "success" },
  { name: "go", icon: "🔷", version: "1.24.0", status: "success" },
  { name: "python", icon: "🐍", version: "3.12.2", status: "success" },
  { name: "pre-commit", icon: "🪝", version: "4.7.0", status: "success" },
  { name: "trivy", icon: "🛡️", version: "0.50.1", status: "success" },
  { name: "eslint", icon: "⚡", version: "10.12.0", status: "success" },
];

const filteredToolItems = computed(() => {
  if (!searchQuery.value) return toolItems;
  return toolItems.filter((t) => t.name.toLowerCase().includes(searchQuery.value.toLowerCase()));
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
  { id: "uniagent", icon: "🤖", name: "UniAgent", desc: "AI Troubleshooter", installed: true },
  { id: "security-suite", icon: "🛡️", name: "Security Suite", desc: "Trivy & Syft Scans", installed: true },
  { id: "rust-provider", icon: "🦀", name: "Rust Provider", desc: "Cargo & Rust Toolchain", installed: false },
  { id: "datascience", icon: "🔬", name: "DataScience Pack", desc: "Python uv & Jupyter", installed: false },
];

const filteredExtensions = computed(() => {
  if (!searchQuery.value) return extensions;
  return extensions.filter((e) => e.name.toLowerCase().includes(searchQuery.value.toLowerCase()));
});

// Doctor checks
const doctorChecks = [
  { id: "os-arch", icon: "💻", label: "OS & Arch Support" },
  { id: "lockfile", icon: "🔒", label: "Lockfile Hygiene" },
  { id: "env-sandbox", icon: "📦", label: "Environment Isolation" },
  { id: "mirror-speed", icon: "⚡", label: "Mirror Latency" },
];
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
