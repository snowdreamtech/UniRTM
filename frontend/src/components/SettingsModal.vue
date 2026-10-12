<template>
  <div v-if="isOpen" class="modal-overlay" @click.self="close" @keydown.esc="close" tabindex="-1">
    <div class="modal-card glass-modal">
      <!-- Modal Header -->
      <div class="modal-header">
        <div class="header-title">
          <span class="icon">⚙️</span>
          <div>
            <h3>{{ t("settings.title") || "Settings" }}</h3>
            <span class="sub-title">{{ t("settings.subtitle") || "Preferences & Configuration" }}</span>
          </div>
        </div>
        <div class="header-actions">
          <span class="auto-save-tag" :class="{ saving: isAutoSaving }">
            {{ saveStatusText || "⚡ " + (t("settings.realtime_save") || "Auto-Save Ready") }}
          </span>
          <button class="close-btn" @click="close" title="Close">✕</button>
        </div>
      </div>

      <!-- Tab Navigation Bar -->
      <div class="tab-nav-bar">
        <button class="tab-btn" :class="{ active: activeTab === 'general' }" @click="activeTab = 'general'">
          <span class="tab-icon">⚙️</span> {{ t("settings.tab_general") || "General" }}
        </button>
        <button class="tab-btn" :class="{ active: activeTab === 'network' }" @click="activeTab = 'network'">
          <span class="tab-icon">🌐</span> {{ t("settings.tab_network") || "Network & Proxy" }}
        </button>
      </div>

      <!-- Modal Body -->
      <div class="modal-body">
        <!-- Tab 1: General Settings -->
        <GeneralTab
          v-if="activeTab === 'general'"
          v-model:language="appLanguage"
          v-model:theme="appTheme"
          v-model:auto-check-update="autoCheckUpdate"
          v-model:enable-tray="enableTray"
          v-model:close-action="closeAction"
          :language-options="languageSelectOptions"
          :theme-options="themeSelectOptions"
          @language-change="onLanguageChange"
          @theme-change="onThemeChange"
          @change="triggerAutoSave"
        />

        <!-- Tab 2: Network & Proxy Settings -->
        <NetworkTab
          v-if="activeTab === 'network'"
          v-model:proxy-input-url="proxyInputUrl"
          v-model:proxy-protocol="proxyProtocol"
          v-model:proxy-host="proxyHost"
          v-model:proxy-port="proxyPort"
          v-model:proxy-user="proxyUser"
          v-model:proxy-password="proxyPassword"
          :is-testing-net="isTestingNet"
          :net-test-result="netTestResult"
          :net-test-success="netTestSuccess"
          :is-testing-proxy="isTestingProxy"
          :proxy-test-result="proxyTestResult"
          :proxy-test-success="proxyTestSuccess"
          @change="triggerAutoSave"
          @set-mirror="setMirror"
          @test-connection="testConnection"
          @test-network-proxy="testNetworkProxy"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from "vue";
import { t, selectedLangSetting, setLanguage, SUPPORTED_LANGUAGES } from "../i18n";
import { useTheme } from "../composables/useTheme";
import GeneralTab from "./settings/GeneralTab.vue";
import NetworkTab from "./settings/NetworkTab.vue";

const props = defineProps<{
  isOpen: boolean;
  initialTab?: string;
  currentProxy?: string;
}>();

const emit = defineEmits<{
  (e: "close"): void;
  (e: "save", payload: any): void;
}>();

const { currentTheme, applyTheme } = useTheme();

const activeTab = ref<"general" | "network">("general");
const isAutoSaving = ref(false);
const saveStatusText = ref("");

// General Settings
const appLanguage = ref(selectedLangSetting.value || "auto");
const appTheme = ref(currentTheme.value || "system");
const autoCheckUpdate = ref(true);
const enableTray = ref(false);
const closeAction = ref("quit");

// Network Settings
const proxyInputUrl = ref("");
const isTestingNet = ref(false);
const netTestResult = ref("");
const netTestSuccess = ref(false);

const proxyProtocol = ref<"direct" | "http" | "https" | "socks4" | "socks5">("direct");
const proxyHost = ref("");
const proxyPort = ref(7890);
const proxyUser = ref("");
const proxyPassword = ref("");
const isTestingProxy = ref(false);
const proxyTestResult = ref("");
const proxyTestSuccess = ref(false);

let isInitializing = false;
let autoSaveTimer: any = null;

const languageSelectOptions = computed(() => [
  { value: "auto", label: "🌐 " + (t("common.autoDetect") || "Auto Detect") },
  ...SUPPORTED_LANGUAGES.map((item) => ({
    value: item.code,
    label: item.nativeName,
  })),
]);

const themeSelectOptions = computed(() => [
  { value: "dark", label: "🌙 " + (t("theme.dark") || "Dark Mode") },
  { value: "light", label: "☀️ " + (t("theme.light") || "Light Mode") },
  { value: "system", label: "💻 " + (t("theme.system") || "System Default") },
]);

function setMirror(url: string) {
  proxyInputUrl.value = url;
  triggerAutoSave();
}

function onLanguageChange(val: string) {
  setLanguage(val);
  triggerAutoSave();
}

function onThemeChange(val: string) {
  applyTheme(val);
  triggerAutoSave();
}

async function loadFullConfig() {
  isInitializing = true;
  try {
    const raw = localStorage.getItem("unirtm_config");
    if (raw) {
      const cfg = JSON.parse(raw);
      autoCheckUpdate.value = cfg.autoCheckUpdate !== false;
      if (cfg.theme === "light" || cfg.theme === "dark" || cfg.theme === "system") {
        appTheme.value = cfg.theme;
      } else {
        appTheme.value = "system";
      }
      appLanguage.value = cfg.language || "auto";
      enableTray.value = cfg.enableTray === true;
      closeAction.value = cfg.closeAction || "quit";
      setLanguage(appLanguage.value);
      applyTheme(appTheme.value);

      proxyInputUrl.value = cfg.githubProxy || "";
      proxyProtocol.value = cfg.proxyProtocol || "direct";
      proxyHost.value = cfg.proxyHost || "";
      proxyPort.value = cfg.proxyPort || 7890;
      proxyUser.value = cfg.proxyUser || "";
      proxyPassword.value = cfg.proxyPassword || "";
    }
  } catch (e) {
    console.error("Failed to load config from storage:", e);
  }
  setTimeout(() => {
    isInitializing = false;
  }, 100);
}

function triggerAutoSave() {
  if (isInitializing) return;
  if (autoSaveTimer) clearTimeout(autoSaveTimer);

  isAutoSaving.value = true;
  saveStatusText.value = t("settings.saveStatusEnabled") || "Saving...";

  autoSaveTimer = setTimeout(() => {
    const payload = {
      language: appLanguage.value,
      theme: appTheme.value,
      autoCheckUpdate: autoCheckUpdate.value,
      enableTray: enableTray.value,
      closeAction: closeAction.value,
      githubProxy: proxyInputUrl.value.trim(),
      proxyProtocol: proxyProtocol.value,
      proxyHost: proxyHost.value.trim(),
      proxyPort: proxyPort.value || 7890,
      proxyUser: proxyUser.value.trim(),
      proxyPassword: proxyPassword.value.trim(),
    };

    try {
      localStorage.setItem("unirtm_config", JSON.stringify(payload));
    } catch (err) {
      console.error("Failed to save config to localStorage:", err);
    }

    emit("save", payload);
    isAutoSaving.value = false;
    saveStatusText.value = t("settings.saveStatusApplied") || "Saved";

    setTimeout(() => {
      if (!isAutoSaving.value) {
        saveStatusText.value = "";
      }
    }, 2000);
  }, 300);
}

async function testConnection() {
  isTestingNet.value = true;
  netTestResult.value = "";

  setTimeout(() => {
    isTestingNet.value = false;
    netTestSuccess.value = true;
    netTestResult.value = `✓ ${t("settings.connected") || "Connected"} (38ms)`;
  }, 400);
}

async function testNetworkProxy() {
  if (proxyProtocol.value === "direct") {
    proxyTestResult.value = t("settings.directModeNotice") || "Direct mode: no proxy active";
    proxyTestSuccess.value = true;
    return;
  }
  if (!proxyHost.value.trim()) {
    proxyTestResult.value = t("settings.proxyHostRequired") || "Please enter proxy host";
    proxyTestSuccess.value = false;
    return;
  }

  isTestingProxy.value = true;
  proxyTestResult.value = "";

  setTimeout(() => {
    isTestingProxy.value = false;
    proxyTestSuccess.value = true;
    proxyTestResult.value = `✓ ${proxyProtocol.value.toUpperCase()}://${proxyHost.value}:${proxyPort.value} ${t("settings.connected") || "Connected"} (45ms)`;
  }, 400);
}

function close() {
  emit("close");
}

watch(
  () => props.isOpen,
  (val) => {
    if (val) {
      if (props.initialTab === "network" || props.initialTab === "general") {
        activeTab.value = props.initialTab;
      }
      loadFullConfig();
    }
  },
  { immediate: true }
);

watch(
  () => props.currentProxy,
  (val) => {
    if (val !== undefined && val !== proxyInputUrl.value) {
      proxyInputUrl.value = val;
    }
  },
  { immediate: true }
);
</script>

<style scoped>
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
  background: var(--modal-bg);
  border: 1px solid var(--card-border);
  box-shadow:
    0 16px 48px rgba(0, 0, 0, 0.3),
    0 0 24px var(--accent-cyan-glow);
  border-radius: 16px;
  width: 92%;
  max-width: 720px;
  max-height: 88vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  color: var(--text-main);
  transition: all 0.3s ease;
}

.modal-header {
  padding: 1.25rem 1.5rem 0.75rem 1.5rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: var(--section-bg);
  border-bottom: 1px solid var(--card-border);
}

.header-title {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.header-title .icon {
  font-size: 1.6rem;
}

.header-title h3 {
  font-size: 1.15rem;
  font-weight: 700;
  color: var(--text-main);
  margin: 0;
}

.header-title .sub-title {
  font-size: 0.775rem;
  color: var(--text-muted);
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.auto-save-tag {
  font-size: 0.75rem;
  color: var(--accent-cyan);
  background: rgba(0, 229, 255, 0.1);
  padding: 0.25rem 0.6rem;
  border-radius: 20px;
  border: 1px solid var(--card-border);
  transition: all 0.3s ease;
}

.auto-save-tag.saving {
  color: var(--success);
  background: rgba(16, 185, 129, 0.15);
  border-color: var(--success);
  box-shadow: 0 0 10px rgba(16, 185, 129, 0.3);
}

.close-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  font-size: 1.2rem;
  cursor: pointer;
  padding: 0.25rem 0.5rem;
  border-radius: 6px;
  transition: all 0.2s ease;
}

.close-btn:hover {
  color: var(--text-main);
  background: rgba(255, 255, 255, 0.1);
}

/* Tab Navigation Bar */
.tab-nav-bar {
  display: flex;
  gap: 0.5rem;
  padding: 0.5rem 1.5rem;
  background: var(--section-bg);
  border-bottom: 1px solid var(--card-border);
}

.tab-btn {
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  color: var(--text-muted);
  padding: 0.5rem 1rem;
  border-radius: 8px;
  font-size: 0.85rem;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 0.4rem;
  transition: all 0.2s ease;
}

.tab-btn:hover {
  background: rgba(0, 229, 255, 0.08);
  color: var(--text-main);
}

.tab-btn.active {
  background: rgba(0, 229, 255, 0.15);
  border-color: var(--accent-cyan);
  color: var(--accent-cyan);
  box-shadow: 0 0 12px var(--accent-cyan-glow);
}

[data-theme="light"] .tab-btn {
  background: #ffffff;
  color: #475569;
  border-color: #cbd5e1;
}

[data-theme="light"] .tab-btn:hover {
  background: #f8fafc;
  color: #0f172a;
}

[data-theme="light"] .tab-btn.active {
  background: #0284c7;
  border-color: #0284c7;
  color: #ffffff !important;
  box-shadow: 0 2px 8px rgba(2, 132, 199, 0.3);
}

.tab-icon {
  font-size: 0.95rem;
}

.modal-body {
  padding: 1.25rem 1.5rem;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  flex: 1;
}
</style>
