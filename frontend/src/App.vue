<template>
  <div class="vscode-app-layout" id="app-workbench">
    <!-- Global Toast Notification -->
    <ToastNotification :message="toastMessage" :type="toastType" @close="dismissToast" />

    <!-- 1. Top TitleBar (VS Code Titlebar / Command Bar) -->
    <TitleBar
      :currentLang="selectedLangSetting"
      :currentTheme="currentTheme"
      @toggle-theme="toggleTheme"
      @open-palette="openCommandPalette"
      @select-lang="selectLanguage"
    />

    <!-- 2. Main Workbench Row (ActivityBar + SidebarPanel + Canvas Workspace) -->
    <div class="workbench-main-row">
      <!-- Activity Bar (Far Left Icon Navigation) -->
      <ActivityBar
        :currentTab="currentTab"
        :isDrawerOpen="isDrawerOpen"
        @select-tab="handleSelectActivityTab"
        @open-settings="openSettings"
        @open-about="openAbout"
        @toggle-drawer="toggleDrawer"
      />

      <!-- Primary Collapsible Sidebar Panel -->
      <SidebarPanel
        :isOpen="isSidebarOpen"
        :currentTab="currentTab"
        :activeSubItem="activeSubItem"
        @toggle-collapse="isSidebarOpen = false"
        @select-sub-item="activeSubItem = $event"
        @run-task="handleTaskStarted"
      />

      <!-- Main Canvas / Editor Space -->
      <main class="workbench-canvas" id="workbench-canvas">
        <!-- Canvas Breadcrumbs Navigation Bar -->
        <div class="canvas-breadcrumb-bar">
          <span class="bc-item">UniRTM</span>
          <span class="bc-separator">/</span>
          <span class="bc-item active">{{ currentTabLabel }}</span>
          <span v-if="activeSubItem" class="bc-separator">/</span>
          <span v-if="activeSubItem" class="bc-item sub font-mono">{{ activeSubItem }}</span>
        </div>

        <!-- Canvas Views -->
        <div class="canvas-content-wrapper">
          <DashboardView
            v-if="currentTab === 'dashboard'"
            :appConfig="appConfig"
            @select-tab="handleDirectSelectTab"
            @log-event="addLog"
          />

          <RuntimesView
            v-else-if="currentTab === 'runtimes'"
            @log-event="addLog"
          />

          <ToolsView
            v-else-if="currentTab === 'tools'"
            @log-event="addLog"
          />

          <TasksView
            v-else-if="currentTab === 'tasks'"
            @task-started="handleTaskStarted"
            @task-finished="handleTaskFinished"
            @log-event="addLog"
          />

          <MarketplaceView
            v-else-if="currentTab === 'marketplace'"
            @log-event="addLog"
          />

          <DoctorView
            v-else-if="currentTab === 'doctor'"
            @log-event="addLog"
            @open-drawer="handleOpenDrawerTab"
          />

          <SettingsView
            v-else-if="currentTab === 'settings'"
            :initialScope="settingsScope"
            :activeCategory="activeSubItem"
            @select-category="activeSubItem = $event"
          />

          <AboutView
            v-else-if="currentTab === 'about'"
          />
        </div>
      </main>
    </div>

    <!-- 3. Bottom Drawer (Console / Task Terminal / Logs) -->
    <BottomDrawer
      :isOpen="isDrawerOpen"
      :logs="runtimeLogs"
      :activeTab="drawerActiveTab"
      @update:activeTab="drawerActiveTab = $event"
      @close="isDrawerOpen = false"
      @clear-logs="handleClearLogs"
    />

    <!-- 4. Bottom StatusBar (VS Code Clean Status Bar) -->
    <BottomStatusBar
      :activeTaskName="activeRunningTask"
      :latencyMs="currentLatencyMs"
      :currentLang="selectedLangSetting"
      @select-tab="handleDirectSelectTab"
      @open-network-settings="openSettings"
      @toggle-drawer="toggleDrawer"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from "vue";
import TitleBar from "./components/layout/TitleBar.vue";
import ActivityBar from "./components/layout/ActivityBar.vue";
import SidebarPanel from "./components/layout/SidebarPanel.vue";
import BottomStatusBar from "./components/layout/BottomStatusBar.vue";
import BottomDrawer, { type LogItem } from "./components/layout/BottomDrawer.vue";

import DashboardView from "./views/DashboardView.vue";
import RuntimesView from "./views/RuntimesView.vue";
import ToolsView from "./views/ToolsView.vue";
import TasksView from "./views/TasksView.vue";
import MarketplaceView from "./views/MarketplaceView.vue";
import DoctorView from "./views/DoctorView.vue";
import SettingsView from "./views/SettingsView.vue";
import AboutView from "./views/AboutView.vue";

import ToastNotification from "./components/ToastNotification.vue";

import { t, selectedLangSetting, setLanguage } from "./i18n";
import { useTheme } from "./composables/useTheme";
import { useToast } from "./composables/useToast";

export interface AppConfig {
  language?: string;
  theme?: string;
  autoCheckUpdate?: boolean;
  enableTray?: boolean;
  closeAction?: string;
  githubProxy?: string;
  proxyProtocol?: string;
  proxyHost?: string;
  proxyPort?: number;
  proxyUser?: string;
  proxyPassword?: string;
}

const { currentTheme, applyTheme, toggleTheme: baseToggleTheme } = useTheme();
const { toastMessage, toastType, showToast, dismissToast } = useToast();

const DRAWER_STORAGE_KEY = "unirtm_drawer_open";

function getInitialDrawerOpen(): boolean {
  if (typeof localStorage === "undefined") return false;
  return localStorage.getItem(DRAWER_STORAGE_KEY) === "true";
}

const currentTab = ref<string>("dashboard");
const activeSubItem = ref<string>("overview");
const isSidebarOpen = ref<boolean>(true);
const isDrawerOpen = ref<boolean>(getInitialDrawerOpen());
const settingsScope = ref<"user" | "workspace">("user");

watch(isDrawerOpen, (newVal) => {
  if (typeof localStorage !== "undefined") {
    localStorage.setItem(DRAWER_STORAGE_KEY, String(newVal));
  }
});

const activeRunningTask = ref<string | undefined>(undefined);
const currentLatencyMs = ref<number | null>(18);
const drawerActiveTab = ref<"output" | "tasks" | "doctor">("output");

const runtimeLogs = ref<LogItem[]>([
  {
    timestamp: new Date().toLocaleTimeString(),
    level: "INFO",
    message: "UniRTM Workbench initialised successfully",
    details: "Pure Go shims router + VS Code style workbench layout",
    category: "output",
  },
  {
    timestamp: new Date().toLocaleTimeString(),
    level: "SUCCESS",
    message: "Environment verified: Node.js 22.14, Go 1.24, Python 3.12 ready",
    category: "output",
  },
]);

const appConfig = ref<AppConfig | null>(null);

const currentTabLabel = computed(() => {
  switch (currentTab.value) {
    case "dashboard":
      return t("nav.dashboard") || "Dashboard";
    case "runtimes":
      return t("nav.runtimes") || "Runtimes";
    case "tools":
      return t("nav.tools") || "Toolchains";
    case "tasks":
      return t("nav.tasks") || "Task Runner";
    case "marketplace":
      return t("nav.marketplace") || "Extensions Marketplace";
    case "doctor":
      return t("nav.doctor") || "Doctor Diagnostics";
    case "settings":
      return t("settings.title") || "设置";
    case "about":
      return t("about.title") || "关于";
    default:
      return currentTab.value.toUpperCase();
  }
});

function handleSelectActivityTab(tabId: string) {
  if (currentTab.value === tabId) {
    // If clicking same active tab in activity bar, toggle sidebar visibility (classic VS Code behavior)
    isSidebarOpen.value = !isSidebarOpen.value;
  } else {
    currentTab.value = tabId;
    isSidebarOpen.value = true;
    activeSubItem.value = "";
  }
}

function handleDirectSelectTab(tabId: string) {
  currentTab.value = tabId;
  isSidebarOpen.value = true;
}

function toggleDrawer() {
  isDrawerOpen.value = !isDrawerOpen.value;
}

function openSettings(scope?: "user" | "workspace") {
  if (scope) {
    settingsScope.value = scope;
  }
  currentTab.value = "settings";
  isSidebarOpen.value = true;
}

function openAbout() {
  currentTab.value = "about";
  activeSubItem.value = "";
  isSidebarOpen.value = true;
}

function openCommandPalette() {
  showToast("Command Palette: Press Ctrl+K (All quick commands available)", "info");
}

function handleOpenDrawerTab(tab: "output" | "tasks" | "doctor" = "output") {
  isDrawerOpen.value = true;
  drawerActiveTab.value = tab;
}

function handleClearLogs(category?: "output" | "tasks" | "doctor") {
  if (category) {
    runtimeLogs.value = runtimeLogs.value.filter((log) => {
      const logCat = log.category || "output";
      return logCat !== category;
    });
  } else {
    runtimeLogs.value = [];
  }
}

function addLog(
  level: "INFO" | "SUCCESS" | "WARN" | "ERROR",
  msg: string,
  detail?: string,
  category: "output" | "tasks" | "doctor" = "output"
) {
  runtimeLogs.value.push({
    timestamp: new Date().toLocaleTimeString(),
    level,
    message: msg,
    details: detail,
    category,
  });
}

function handleTaskStarted(taskName: string) {
  activeRunningTask.value = taskName;
  isDrawerOpen.value = true;
  drawerActiveTab.value = "tasks";
  addLog("INFO", `Running project task: unirtm run ${taskName}...`, undefined, "tasks");
  setTimeout(() => {
    activeRunningTask.value = undefined;
    addLog("SUCCESS", `Task ${taskName} executed successfully.`, undefined, "tasks");
  }, 1200);
}

function handleTaskFinished(taskName: string) {
  activeRunningTask.value = undefined;
  addLog("SUCCESS", `Task ${taskName} finished.`, undefined, "tasks");
}

async function saveGlobalSetting(key: string, value: any) {
  try {
    await fetch("/api/config", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        scope: "user",
        key,
        value,
      }),
    });
  } catch (err) {
    console.warn("Failed to persist setting to config:", err);
  }
}

function selectLanguage(langVal: string) {
  setLanguage(langVal);
  if (appConfig.value) {
    appConfig.value.language = langVal;
  }
  saveGlobalSetting("settings.language", langVal);
  showToast(`Language switched to ${langVal}`, "info");
}

function toggleTheme() {
  const nextTheme = baseToggleTheme();
  if (appConfig.value) {
    appConfig.value.theme = nextTheme;
  }
  saveGlobalSetting("settings.theme", nextTheme);
}

async function loadConfig() {
  try {
    const raw = localStorage.getItem("unirtm_config");
    if (raw) {
      const cfg = JSON.parse(raw);
      appConfig.value = cfg;
      if (cfg.theme) {
        applyTheme(cfg.theme);
      }
      if (cfg.language) {
        setLanguage(cfg.language);
      }
    }
  } catch (err) {
    console.warn("Failed to load initial configuration:", err);
  }

  try {
    const res = await fetch("/api/config");
    if (res.ok) {
      const data = await res.json();
      const userSettings = data?.user?.settings || {};
      if (userSettings.theme) {
        applyTheme(userSettings.theme);
      }
      if (userSettings.language) {
        setLanguage(userSettings.language);
      }
    }
  } catch (err) {
    console.warn("Failed to sync /api/config on startup:", err);
  }
}

function handleGlobalKeydown(e: KeyboardEvent) {
  // Ctrl+B / Cmd+B: Toggle Sidebar (VS Code shortcut)
  if ((e.ctrlKey || e.metaKey) && (e.key === "b" || e.key === "B")) {
    e.preventDefault();
    isSidebarOpen.value = !isSidebarOpen.value;
  }
  // Ctrl+` / Cmd+`: Toggle Console Drawer (VS Code shortcut)
  else if ((e.ctrlKey || e.metaKey) && e.key === "`") {
    e.preventDefault();
    isDrawerOpen.value = !isDrawerOpen.value;
  }
  // Ctrl+K: Command search
  else if ((e.ctrlKey || e.metaKey) && (e.key === "k" || e.key === "K")) {
    e.preventDefault();
    openCommandPalette();
  }
  // Ctrl+, / Cmd+,: Open Settings (VS Code shortcut)
  else if ((e.ctrlKey || e.metaKey) && e.key === ",") {
    e.preventDefault();
    openSettings();
  }
  // Escape
  else if (e.key === "Escape") {
    if (isDrawerOpen.value) {
      isDrawerOpen.value = false;
    }
  }
}

onMounted(() => {
  loadConfig();
  window.addEventListener("keydown", handleGlobalKeydown);
});

onUnmounted(() => {
  window.removeEventListener("keydown", handleGlobalKeydown);
});
</script>

<style scoped>
.vscode-app-layout {
  display: flex;
  flex-direction: column;
  width: 100vw;
  height: 100vh;
  overflow: hidden;
  background-color: var(--bg-color);
  color: var(--text-main);
  box-sizing: border-box;
}

.workbench-main-row {
  flex: 1;
  display: flex;
  overflow: hidden;
  position: relative;
}

.workbench-canvas {
  flex: 1;
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
  background: var(--bg-color);
  position: relative;
}

.canvas-breadcrumb-bar {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  height: 28px;
  min-height: 28px;
  background: var(--card-bg);
  border-bottom: 1px solid var(--card-border);
  padding: 0 1rem;
  font-size: 0.72rem;
  color: var(--text-muted);
  user-select: none;
  -webkit-user-select: none;
}

.bc-item {
  color: var(--text-muted);
}

.bc-item.active {
  color: var(--text-main);
  font-weight: 600;
}

.bc-item.sub {
  color: var(--accent-cyan);
}

.bc-separator {
  opacity: 0.4;
}

.canvas-content-wrapper {
  flex: 1;
  height: calc(100% - 28px);
  overflow-y: auto;
  position: relative;
}
</style>
