<template>
  <div class="app-container">
    <!-- Global Toast Notification -->
    <ToastNotification :message="toastMessage" :type="toastType" @close="dismissToast" />

    <!-- Header Component -->
    <AppHeader
      :currentLang="selectedLangSetting"
      :currentTheme="currentTheme"
      @toggle-theme="toggleTheme"
      @open-settings="openSettings"
      @open-about="openAbout"
      @open-logs="isLogViewerOpen = true"
      @select-lang="selectLanguage"
    />

    <!-- Main Content Area -->
    <main class="main-content">
      <HelloPanel :appConfig="appConfig" @open-settings="openSettings" @open-about="openAbout" />
    </main>

    <!-- Settings Modal -->
    <SettingsModal
      :isOpen="isSettingsOpen"
      :currentProxy="appConfig?.githubProxy"
      @close="isSettingsOpen = false"
      @save="onSaveSettings"
    />

    <!-- About Modal -->
    <AboutModal :show="isAboutOpen" @close="isAboutOpen = false" />

    <!-- Log Viewer Modal -->
    <LogViewerModal
      :isOpen="isLogViewerOpen"
      :logs="runtimeLogs"
      @close="isLogViewerOpen = false"
      @clear="handleClearLogs"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from "vue";
import AppHeader from "./components/AppHeader.vue";
import HelloPanel from "./components/HelloPanel.vue";
import SettingsModal from "./components/SettingsModal.vue";
import AboutModal from "./components/AboutModal.vue";
import LogViewerModal, { type LogItem } from "./components/LogViewerModal.vue";
import ToastNotification from "./components/ToastNotification.vue";
import { selectedLangSetting, setLanguage } from "./i18n";
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

const isSettingsOpen = ref(false);
const isAboutOpen = ref(false);
const isLogViewerOpen = ref(false);
const runtimeLogs = ref<LogItem[]>([
  {
    timestamp: new Date().toLocaleTimeString(),
    level: "INFO",
    message: "UniRTM Web dashboard initialized",
    details: "Runtime: Web / Localhost UI",
  },
  {
    timestamp: new Date().toLocaleTimeString(),
    level: "SUCCESS",
    message: "Design system & i18n modules loaded successfully",
  },
]);
const appConfig = ref<AppConfig | null>(null);

function handleClearLogs() {
  runtimeLogs.value = [];
}

function selectLanguage(langVal: string) {
  setLanguage(langVal);
  if (appConfig.value) {
    appConfig.value.language = langVal;
  }
  showToast(`Language switched to ${langVal}`, "info");
}

function toggleTheme() {
  const nextTheme = baseToggleTheme();
  if (appConfig.value) {
    appConfig.value.theme = nextTheme;
  }
}

function openSettings() {
  isSettingsOpen.value = true;
}

function openAbout() {
  isAboutOpen.value = true;
}

function onSaveSettings(savedCfg: any) {
  appConfig.value = { ...(appConfig.value || {}), ...savedCfg };
  if (savedCfg.theme) {
    applyTheme(savedCfg.theme);
  }
  if (savedCfg.language) {
    setLanguage(savedCfg.language);
  }
  try {
    localStorage.setItem("unirtm_config", JSON.stringify(appConfig.value));
  } catch (e) {
    console.error("Failed to persist config:", e);
  }
  showToast("Settings saved successfully", "success");
}

function loadConfig() {
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
}

function handleGlobalKeydown(e: KeyboardEvent) {
  if (e.key === "Escape") {
    if (isAboutOpen.value) {
      isAboutOpen.value = false;
    } else if (isSettingsOpen.value) {
      isSettingsOpen.value = false;
    } else if (isLogViewerOpen.value) {
      isLogViewerOpen.value = false;
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
.app-container {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  width: 100%;
  max-width: 1200px;
  margin: 0 auto;
  background-color: var(--bg-color);
  color: var(--text-main);
  padding: 2.25rem 2rem 1.5rem;
  box-sizing: border-box;
}

.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
}
</style>
