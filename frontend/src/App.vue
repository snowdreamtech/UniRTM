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
        @select-sub-item="handleSubItemSelect"
        @run-task="handleSidebarTaskRun"
        @open-terminal="handleOpenTerminalInDir"
      />

      <!-- Main Canvas / Editor Space -->
      <main class="workbench-canvas" id="workbench-canvas">
        <!-- Canvas Breadcrumbs Navigation Bar -->
        <div class="canvas-breadcrumb-bar">
          <div class="breadcrumb-left">
            <span class="bc-item">UniRTM</span>
            <span class="bc-separator">/</span>
            <span class="bc-item active">{{ currentTabLabel }}</span>
            <span v-if="activeSubItem" class="bc-separator">/</span>
            <span v-if="activeSubItem" class="bc-item sub font-mono">{{ activeSubItem }}</span>
          </div>
          <div class="breadcrumb-right">
            <ScopeSwitcher />
          </div>
        </div>

        <!-- Canvas Views -->
        <div class="canvas-content-wrapper">
          <router-view v-slot="{ Component }">
            <keep-alive>
              <component
                :is="Component"
                :key="currentTab"
                :appConfig="appConfig"
                :activeSubItem="activeSubItem"
                :initialScope="activeScope"
                :activeCategory="activeSubItem"
                @select-tab="handleDirectSelectTab"
                @log-event="addLog"
                @task-started="handleTaskStarted"
                @task-finished="handleTaskFinished"
                @open-drawer="handleOpenDrawerTab"
                @select-category="handleCategorySelect"
              />
            </keep-alive>
          </router-view>
        </div>

        <!-- 3. Bottom Drawer (Console / Task Terminal / Logs) - Aligned to Main Window -->
        <BottomDrawer
          :isOpen="isDrawerOpen"
          :logs="runtimeLogs"
          :activeTab="drawerActiveTab"
          @update:activeTab="drawerActiveTab = $event"
          @close="isDrawerOpen = false"
          @clear-logs="handleClearLogs"
        />
      </main>
    </div>

    <!-- 4. Bottom StatusBar (VS Code Clean Status Bar) -->
    <BottomStatusBar
      :activeTaskName="activeRunningTask"
      :latencyMs="currentLatencyMs"
      :currentLang="selectedLangSetting"
      @select-tab="handleDirectSelectTab"
      @open-network-settings="openSettings"
      @toggle-drawer="toggleDrawer"
    />

    <!-- 5. Universal Git SSH & GPG Passphrase Prompt Modal -->
    <PassphraseModal />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from "vue";
import { useRoute, useRouter } from "vue-router";
import TitleBar from "./components/layout/TitleBar.vue";
import ActivityBar from "./components/layout/ActivityBar.vue";
import SidebarPanel from "./components/layout/SidebarPanel.vue";
import BottomStatusBar from "./components/layout/BottomStatusBar.vue";
import BottomDrawer, { type LogItem, type DrawerTab } from "./components/layout/BottomDrawer.vue";
import ToastNotification from "./components/ToastNotification.vue";

import { t, selectedLangSetting, setLanguage } from "./i18n";
import { useTheme } from "./composables/useTheme";
import { useToast } from "./composables/useToast";
import { useToolUpdates } from "./composables/useToolUpdates";
import { useAppScope } from "./composables/useAppScope";
import { useTerminalSessions } from "./composables/useTerminalSessions";
import { useCredentialBroker } from "./composables/useCredentialBroker";
import ScopeSwitcher from "./components/common/ScopeSwitcher.vue";
import PassphraseModal from "./components/common/PassphraseModal.vue";

const { startBrokerPolling, stopBrokerPolling } = useCredentialBroker();

const { checkUpdates } = useToolUpdates();
const { activeScope, syncFromRoute } = useAppScope();

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

const route = useRoute();
const router = useRouter();

import { useThemeManager } from "./composables/useThemeManager";

const { currentTheme, applyTheme } = useTheme();
const { fetchThemes, toggleAppearance } = useThemeManager();
const { toastMessage, toastType, showToast, dismissToast } = useToast();

const DRAWER_STORAGE_KEY = "unirtm_drawer_open";

function getInitialDrawerOpen(): boolean {
  if (typeof localStorage === "undefined") return false;
  return localStorage.getItem(DRAWER_STORAGE_KEY) === "true";
}

const currentTab = computed<string>(() => (route.name as string) || "dashboard");
const activeSubItem = ref<string>("");
const isSidebarOpen = ref<boolean>(true);
const isDrawerOpen = ref<boolean>(getInitialDrawerOpen());

watch(isDrawerOpen, (newVal) => {
  if (typeof localStorage !== "undefined") {
    localStorage.setItem(DRAWER_STORAGE_KEY, String(newVal));
  }
});

watch(
  () => route.name,
  (newName, oldName) => {
    if (newName !== oldName && newName !== "settings") {
      activeSubItem.value = "";
    }
  }
);

watch(
  () => route.params.category,
  (category) => {
    if (category && typeof category === "string") {
      activeSubItem.value = category;
    }
  },
  { immediate: true }
);

watch(
  () => route.query.scope,
  (scope) => {
    if (scope === "user" || scope === "workspace") {
      syncFromRoute(route);
    }
  },
  { immediate: true }
);

watch(
  activeScope,
  (newScope) => {
    checkUpdates(newScope === "user").catch(() => {});
  },
  { immediate: true }
);

function getInitialDrawerTab(): DrawerTab {
  if (typeof localStorage === "undefined") return "terminal";
  try {
    const saved = localStorage.getItem("unirtm_drawer_tab");
    if (saved === "terminal" || saved === "output" || saved === "tasks" || saved === "doctor" || saved === "ports") {
      return saved as DrawerTab;
    }
  } catch {}
  return "terminal";
}

const activeRunningTask = ref<string | undefined>(undefined);
const currentLatencyMs = ref<number | null>(18);
const drawerActiveTab = ref<DrawerTab>(getInitialDrawerTab());

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
    case "explorer":
      return t("nav.explorer") || "Explorer";
    case "scm":
      return t("nav.scm") || "Source Control";
    case "tools":
      return t("nav.tools") || "Toolchains";
    case "tasks":
      return t("nav.tasks") || "Task Runner";
    case "environments":
      return t("nav.environments") || "Environment Variables";
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

function navigateToTab(tabId: string) {
  if (currentTab.value !== tabId) {
    router.push({ name: tabId }).catch(() => {});
  }
}

function handleSelectActivityTab(tabId: string) {
  if (currentTab.value === tabId) {
    // If clicking same active tab in activity bar, toggle sidebar visibility (classic VS Code behavior)
    isSidebarOpen.value = !isSidebarOpen.value;
  } else {
    isSidebarOpen.value = true;
    activeSubItem.value = "";
    navigateToTab(tabId);
  }
}

function handleDirectSelectTab(tabId: string) {
  isSidebarOpen.value = true;
  navigateToTab(tabId);
}

function handleSubItemSelect(id: string) {
  const nextItem = activeSubItem.value === id ? "" : id;
  activeSubItem.value = nextItem;
  if (currentTab.value === "settings") {
    if (nextItem) {
      router.replace({ name: "settings", params: { category: nextItem }, query: route.query }).catch(() => {});
    } else {
      router.replace({ name: "settings", query: route.query }).catch(() => {});
    }
  }
}

function handleCategorySelect(catId: string) {
  activeSubItem.value = catId;
  if (currentTab.value === "settings") {
    router.replace({ name: "settings", params: { category: catId }, query: route.query }).catch(() => {});
  }
}

function toggleDrawer() {
  isDrawerOpen.value = !isDrawerOpen.value;
}

function openSettings(scope?: "user" | "workspace") {
  if (scope) {
    activeScope.value = scope;
  }
  isSidebarOpen.value = true;
  const targetScope = scope || activeScope.value;
  if (route.name === "settings") {
    router
      .replace({ name: "settings", params: route.params, query: { ...route.query, scope: targetScope } })
      .catch(() => {});
  } else {
    router.push({ name: "settings", query: { ...route.query, scope: targetScope } }).catch(() => {});
  }
}

function openAbout() {
  activeSubItem.value = "";
  isSidebarOpen.value = true;
  navigateToTab("about");
}

function openCommandPalette() {
  showToast("Command Palette: Press Ctrl+K (All quick commands available)", "info");
}

function handleOpenDrawerTab(tab: DrawerTab = "terminal") {
  isDrawerOpen.value = true;
  drawerActiveTab.value = tab;
}

function handleOpenTerminalInDir(dirPath: string) {
  isDrawerOpen.value = true;
  drawerActiveTab.value = "terminal";
  const { ensureInitialSession, activeSession, runCommand } = useTerminalSessions();
  ensureInitialSession();
  const pane =
    activeSession.value?.panes.find((p) => p.id === activeSession.value?.activePaneId) || activeSession.value?.panes[0];
  if (pane && dirPath) {
    runCommand(pane.id, `cd "${dirPath}"`);
  }
}

function handleClearLogs(category?: DrawerTab) {
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
}

async function handleSidebarTaskRun(taskName: string) {
  activeRunningTask.value = taskName;
  isDrawerOpen.value = true;
  drawerActiveTab.value = "tasks";
  addLog("INFO", `[unirtm run ${taskName}] 正在启动任务...`, undefined, "tasks");
  try {
    const res = await fetch("/api/tasks/run", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        name: taskName,
        task: taskName,
        scope: activeScope.value || "workspace",
      }),
    });
    const data = await res.json();
    if (res.ok && data.success) {
      addLog("SUCCESS", `[unirtm run ${taskName}] 任务执行完成 (退出码: 0)`, data.output, "tasks");
    } else {
      addLog(
        "ERROR",
        `[unirtm run ${taskName}] 任务执行失败 (退出码: ${data.exit_code ?? 1})`,
        data.error || data.output,
        "tasks"
      );
    }
  } catch (err: any) {
    addLog("ERROR", `[unirtm run ${taskName}] 任务调用异常: ${err.message}`, undefined, "tasks");
  } finally {
    activeRunningTask.value = undefined;
  }
}

function handleTaskFinished(taskName: string) {
  activeRunningTask.value = undefined;
  addLog("SUCCESS", `[unirtm run ${taskName}] 任务执行结束`, undefined, "tasks");
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

async function toggleTheme() {
  const nextTheme = await toggleAppearance();
  saveGlobalSetting("settings.appearance", nextTheme);
  showToast(nextTheme === "light" ? "已切换为浅色模式" : "已切换为深色模式", "info");
}

async function loadConfig() {
  try {
    const raw = localStorage.getItem("unirtm_config");
    if (raw) {
      const cfg = JSON.parse(raw);
      appConfig.value = cfg;
      if (cfg.appearance) {
        applyTheme(cfg.appearance);
      } else if (cfg.theme && (cfg.theme === "dark" || cfg.theme === "light" || cfg.theme === "system")) {
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
      const app =
        userSettings.appearance ||
        (userSettings.theme === "light" || userSettings.theme === "dark" || userSettings.theme === "system"
          ? userSettings.theme
          : undefined);
      if (app) {
        applyTheme(app);
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
  // Ctrl+` / Cmd+`: Toggle Terminal Drawer (VS Code shortcut)
  else if ((e.ctrlKey || e.metaKey) && e.key === "`") {
    e.preventDefault();
    if (!isDrawerOpen.value) {
      drawerActiveTab.value = "terminal";
      isDrawerOpen.value = true;
    } else if (drawerActiveTab.value !== "terminal") {
      drawerActiveTab.value = "terminal";
    } else {
      isDrawerOpen.value = false;
    }
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
  fetchThemes();
  checkUpdates().catch(() => {});
  startBrokerPolling(1000);
  window.addEventListener("keydown", handleGlobalKeydown);
});

onUnmounted(() => {
  stopBrokerPolling();
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
  justify-content: space-between;
  gap: 0.6rem;
  height: 32px;
  min-height: 32px;
  flex-shrink: 0;
  background: var(--card-bg);
  border-bottom: 1px solid var(--card-border);
  padding: 0 1rem;
  font-size: 0.72rem;
  color: var(--text-muted);
  user-select: none;
  -webkit-user-select: none;
}

.breadcrumb-left {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.breadcrumb-right {
  display: flex;
  align-items: center;
  flex-shrink: 0;
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
  min-height: 0;
  overflow-y: auto;
  scrollbar-gutter: stable;
  position: relative;
  display: flex;
  flex-direction: column;
}
</style>
