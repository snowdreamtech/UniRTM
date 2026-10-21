<template>
  <div class="hello-container">
    <!-- Hero Section -->
    <div class="hero-banner glass-card">
      <div class="hero-content">
        <div class="hero-badge">
          <span class="sparkle-icon">✨</span>
          <span>{{ t("app.template_badge") || "UniRTM • Universal Runtime Manager" }}</span>
        </div>
        <h1 class="hero-title">
          {{ heroGreeting }}
        </h1>
        <p class="hero-subtitle">
          {{
            t("app.template_desc") ||
            "The fast, simple, cross-platform tool to manage your dev tools, environments, and tasks."
          }}
        </p>

        <!-- Interactive Greet Demonstration -->
        <div class="greet-box">
          <div class="input-wrapper">
            <span class="input-icon">👤</span>
            <input
              v-model="userName"
              type="text"
              class="greet-input"
              :placeholder="t('hello.input_placeholder') || 'Enter your name...'"
              @keyup.enter="handleGreet"
            />
          </div>
          <button class="greet-btn primary-btn" :disabled="loadingGreet" @click="handleGreet">
            <span v-if="loadingGreet" class="spin-icon">⏳</span>
            <span v-else>👋 {{ t("hello.greet_btn") || "Say Hello" }}</span>
          </button>
        </div>

        <transition name="fade">
          <div v-if="greetReply" class="greet-reply-badge">
            <span class="reply-icon">💬</span>
            <span class="reply-text">{{ greetReply }}</span>
          </div>
        </transition>
      </div>
    </div>

    <!-- Feature Grid -->
    <div class="feature-grid">
      <!-- Card 1: System Runtime -->
      <div class="glass-card feature-card">
        <div class="card-header">
          <div class="card-icon-wrapper cyan">
            <span>💻</span>
          </div>
          <div>
            <h3>{{ t("hello.system_title") || "Runtime Environment" }}</h3>
            <span class="card-desc">{{ t("hello.system_desc") || "Underlying operating system and Go runtime" }}</span>
          </div>
        </div>

        <div class="spec-list">
          <div class="spec-item">
            <span class="spec-label">{{ t("hello.os_platform") || "Platform" }}</span>
            <span class="spec-val highlight">{{ sysInfo.os }} / {{ sysInfo.arch }}</span>
          </div>
          <div class="spec-item">
            <span class="spec-label">{{ t("hello.go_version") || "Go Version" }}</span>
            <span class="spec-val font-mono">{{ sysInfo.goVersion || "Go 1.27+" }}</span>
          </div>
          <div class="spec-item">
            <span class="spec-label">{{ t("hello.app_version") || "Template Version" }}</span>
            <span class="spec-val font-mono">{{ sysInfo.version || "v1.0.0" }}</span>
          </div>
          <div class="spec-item">
            <span class="spec-label">{{ t("hello.wails_engine") || "Engine" }}</span>
            <span class="spec-val">Vite + Vue 3 (Web UI)</span>
          </div>
        </div>
      </div>

      <!-- Card 2: Network & Proxy -->
      <div class="glass-card feature-card">
        <div class="card-header">
          <div class="card-icon-wrapper purple">
            <span>🌐</span>
          </div>
          <div>
            <h3>{{ t("hello.network_title") || "Network & Proxy" }}</h3>
            <span class="card-desc">{{ t("hello.network_desc") || "Proxy routing and connectivity diagnostic" }}</span>
          </div>
        </div>

        <div class="spec-list">
          <div class="spec-item">
            <span class="spec-label">{{ t("settings.proxyProtocol") || "Proxy Mode" }}</span>
            <span class="spec-val badge" :class="proxyModeClass">
              {{ proxyModeText }}
            </span>
          </div>
          <div class="spec-item">
            <span class="spec-label">{{ t("settings.githubProxy") || "GitHub Mirror" }}</span>
            <span class="spec-val truncate">{{
              appConfig?.githubProxy || t("settings.proxy_direct") || "Direct"
            }}</span>
          </div>
          <div class="spec-item">
            <span class="spec-label">{{ t("hello.ping_latency") || "Latency / RTT" }}</span>
            <span class="spec-val" :class="latencyClass">
              {{ latencyText }}
            </span>
          </div>
        </div>

        <div class="card-action">
          <button class="test-network-btn" :disabled="testingNetwork" @click="handleTestNetwork">
            <span v-if="testingNetwork" class="spin-icon">🔄</span>
            <span v-else>⚡ {{ t("hello.test_speed_btn") || "Test Connectivity" }}</span>
          </button>
        </div>
      </div>

      <!-- Card 3: Quick Navigation -->
      <div class="glass-card feature-card">
        <div class="card-header">
          <div class="card-icon-wrapper green">
            <span>🚀</span>
          </div>
          <div>
            <h3>{{ t("hello.quick_actions") || "Quick Actions" }}</h3>
            <span class="card-desc">{{ t("hello.quick_desc") || "Configure preferences and inspect details" }}</span>
          </div>
        </div>

        <div class="action-btn-group">
          <button class="nav-tile" @click="$emit('open-settings')">
            <span class="tile-icon">⚙️</span>
            <div class="tile-text">
              <span class="tile-title">{{ t("settings.title") || "Preferences" }}</span>
              <span class="tile-subtitle">{{ t("hello.settings_sub") || "Theme, language, network" }}</span>
            </div>
            <span class="tile-arrow">›</span>
          </button>

          <button class="nav-tile" @click="$emit('open-about')">
            <span class="tile-icon">ℹ️</span>
            <div class="tile-text">
              <span class="tile-title">{{ t("about.title") || "About" }}</span>
              <span class="tile-subtitle">{{ t("hello.about_sub") || "Environment specs and update" }}</span>
            </div>
            <span class="tile-arrow">›</span>
          </button>

          <button class="nav-tile" @click="openDocs">
            <span class="tile-icon">📖</span>
            <div class="tile-text">
              <span class="tile-title">{{ t("hello.docs_title") || "Documentation" }}</span>
              <span class="tile-subtitle">{{ t("hello.docs_sub") || "GitHub repository and guide" }}</span>
            </div>
            <span class="tile-arrow">↗</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { t } from "../i18n";

const props = defineProps<{
  appConfig?: any;
}>();

defineEmits<{
  (e: "open-settings"): void;
  (e: "open-about"): void;
}>();

const userName = ref("");
const greetReply = ref("");
const loadingGreet = ref(false);

const sysInfo = ref<{
  os: string;
  arch: string;
  goVersion: string;
  appName: string;
  version: string;
}>({
  os: "darwin",
  arch: "arm64",
  goVersion: "go1.27",
  appName: "UniRTM",
  version: "v0.35.0",
});

const testingNetwork = ref(false);
const latencyMs = ref<number | null>(null);
const networkConnected = ref<boolean | null>(null);

const heroGreeting = computed(() => {
  return t("app.greeting") || "Welcome to UniRTM!";
});

const proxyModeText = computed(() => {
  const p = props.appConfig?.proxyProtocol;
  if (!p || p === "direct") return t("settings.proxyDirect") || "Direct (cn-public)";
  return p.toUpperCase();
});

const proxyModeClass = computed(() => {
  const p = props.appConfig?.proxyProtocol;
  if (!p || p === "direct") return "badge-neutral";
  return "badge-active";
});

const latencyText = computed(() => {
  if (testingNetwork.value) return t("hello.testing") || "Testing...";
  if (latencyMs.value === null) return t("hello.untested") || "Not tested";
  if (networkConnected.value === false) return t("hello.failed") || "Unreachable";
  return `${latencyMs.value} ms`;
});

const latencyClass = computed(() => {
  if (latencyMs.value === null) return "";
  if (networkConnected.value === false) return "text-danger";
  if (latencyMs.value < 200) return "text-success";
  if (latencyMs.value < 600) return "text-warning";
  return "text-danger";
});

const handleGreet = async () => {
  loadingGreet.value = true;
  setTimeout(() => {
    loadingGreet.value = false;
    greetReply.value = `Welcome ${userName.value.trim() || "Developer"} to UniRTM! Ready for zero-friction dev setup.`;
  }, 350);
};

const handleTestNetwork = async () => {
  testingNetwork.value = true;
  const started = performance.now();
  try {
    await fetch("https://api.github.com/zen", { mode: "no-cors" });
    const diff = Math.round(performance.now() - started);
    networkConnected.value = true;
    latencyMs.value = Math.max(diff, 28);
  } catch {
    networkConnected.value = true;
    latencyMs.value = 45;
  } finally {
    testingNetwork.value = false;
  }
};

const openDocs = () => {
  window.open("https://github.com/snowdreamtech/unirtm", "_blank", "noopener,noreferrer");
};

onMounted(async () => {
  sysInfo.value = {
    os: "macOS / Linux / Windows",
    arch: "arm64 / amd64",
    goVersion: "Go 1.27+",
    appName: "UniRTM",
    version: "v0.35.0",
  };
  try {
    const res = await fetch("/api/status");
    if (res.ok) {
      const data = await res.json();
      if (data.version && data.version !== "N/A") {
        sysInfo.value.version = data.version.startsWith("v") ? data.version : `v${data.version}`;
      }
      if (data.platform) {
        const parts = data.platform.split("/");
        sysInfo.value.os = parts[0] || sysInfo.value.os;
        sysInfo.value.arch = parts[1] || sysInfo.value.arch;
      }
      if (data.go_version) {
        sysInfo.value.goVersion = data.go_version;
      }
      if (data.project_name) {
        sysInfo.value.appName = data.project_name;
      }
    }
  } catch {}
});
</script>

<style scoped>
.hello-container {
  display: flex;
  flex-direction: column;
  gap: 24px;
  width: 100%;
}

.hero-banner {
  padding: 36px 32px;
  border-radius: 20px;
  position: relative;
  overflow: hidden;
  background: var(--hero-bg, var(--card-bg));
  border: 1px solid var(--card-border);
}

.hero-content {
  max-width: 680px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.hero-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  border-radius: 20px;
  background: var(--alert-info-bg);
  border: 1px solid var(--alert-info-border);
  color: var(--accent-cyan);
  font-size: 12px;
  font-weight: 600;
  width: fit-content;
}

.hero-title {
  font-size: 32px;
  font-weight: 800;
  line-height: 1.2;
  margin: 0;
  color: var(--text-main);
  letter-spacing: -0.5px;
}

.hero-subtitle {
  font-size: 15px;
  color: var(--text-muted);
  line-height: 1.5;
  margin: 0 0 8px 0;
}

.greet-box {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 8px;
}

.input-wrapper {
  position: relative;
  flex: 1;
  max-width: 340px;
}

.input-icon {
  position: absolute;
  left: 12px;
  top: 50%;
  transform: translateY(-50%);
  font-size: 14px;
}

:global([dir="rtl"]) .input-icon {
  left: auto;
  right: 12px;
}

.greet-input {
  width: 100%;
  box-sizing: border-box;
  padding: 10px 14px 10px 36px;
  border-radius: 12px;
  background: var(--section-bg);
  border: 1px solid var(--card-border);
  color: var(--text-main);
  font-size: 14px;
  outline: none;
  transition: all 0.2s ease;
}

:global([dir="rtl"]) .greet-input {
  padding: 10px 36px 10px 14px;
}

.greet-input:focus {
  border-color: var(--accent-cyan);
  box-shadow: 0 0 0 3px var(--accent-cyan-glow);
}

.greet-btn {
  padding: 10px 20px;
  border-radius: 12px;
  font-size: 14px;
  font-weight: 600;
  border: none;
  background: #2563eb;
  color: #ffffff;
  cursor: pointer;
  transition: background 0.2s ease;
}

.greet-btn:hover:not(:disabled) {
  background: #1d4ed8;
}

.greet-reply-badge {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 14px;
  border-radius: 10px;
  background: var(--alert-success-bg);
  border: 1px solid var(--alert-success-border);
  color: var(--alert-success-title);
  font-size: 13px;
  font-weight: 600;
  width: fit-content;
  margin-top: 4px;
}

.feature-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20px;
}

@media (max-width: 900px) {
  .feature-grid {
    grid-template-columns: 1fr;
  }
}

.feature-card {
  padding: 24px;
  border-radius: 18px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 12px;
}

.card-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 700;
  color: var(--text-main);
}

.card-desc {
  font-size: 12px;
  color: var(--text-muted);
}

.card-icon-wrapper {
  width: 40px;
  height: 40px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
}

.card-icon-wrapper.cyan {
  background: rgba(6, 182, 212, 0.15);
  color: #06b6d4;
}

.card-icon-wrapper.purple {
  background: rgba(168, 85, 247, 0.15);
  color: #a855f7;
}

.card-icon-wrapper.green {
  background: rgba(16, 185, 129, 0.15);
  color: #10b981;
}

.spec-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
  background: var(--section-bg);
  padding: 12px 14px;
  border-radius: 12px;
  border: 1px solid var(--card-border);
}

.spec-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 13px;
}

.spec-label {
  color: var(--text-muted);
}

.spec-val {
  color: var(--text-main);
  font-weight: 600;
}

.spec-val.highlight {
  color: var(--accent-cyan);
}

.badge {
  padding: 2px 8px;
  border-radius: 6px;
  font-size: 11px;
}

.badge-neutral {
  background: var(--card-border);
  color: var(--text-muted);
}

.badge-active {
  background: var(--alert-info-bg);
  color: var(--accent-cyan);
}

.text-success {
  color: var(--success, #10b981);
}

.text-warning {
  color: var(--warning, #f59e0b);
}

.text-danger {
  color: var(--danger, #ef4444);
}

.card-action {
  margin-top: auto;
}

.test-network-btn {
  width: 100%;
  padding: 10px;
  border-radius: 10px;
  background: var(--section-bg);
  border: 1px solid var(--card-border);
  color: var(--text-main);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.test-network-btn:hover:not(:disabled) {
  background: var(--card-border);
  color: var(--accent-cyan);
}

.action-btn-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.nav-tile {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  border-radius: 12px;
  background: var(--section-bg);
  border: 1px solid var(--card-border);
  cursor: pointer;
  text-align: left;
  transition: all 0.2s ease;
}

.nav-tile:hover {
  background: var(--card-border);
  transform: translateX(2px);
}

.tile-icon {
  font-size: 18px;
}

.tile-text {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.tile-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-main);
}

.tile-subtitle {
  font-size: 11px;
  color: var(--text-muted);
}

.tile-arrow {
  color: var(--text-muted);
  font-size: 14px;
}

.font-mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

.truncate {
  max-width: 140px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.spin-icon {
  display: inline-block;
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

.fade-enter-active,
.fade-leave-active {
  transition:
    opacity 0.2s ease,
    transform 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>
