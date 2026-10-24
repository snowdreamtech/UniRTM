<template>
  <div class="view-container dashboard-view" id="view-dashboard" @contextmenu.prevent>
    <!-- Hero Cockpit Banner -->
    <div class="hero-cockpit glass-card" id="dashboard-section-hero">
      <div class="hero-badge">
        <span class="sparkle-icon">
          <TechIcon id="workbench:sparkle" :size="14" />
        </span>
        <span>{{ t("app.template_badge") || "UniRTM • Universal Runtime Manager" }}</span>
      </div>
      <h1 class="hero-title">{{ t("app.greeting") || "Welcome to UniRTM!" }}</h1>
      <p class="hero-desc">
        {{
          t("app.template_desc") ||
          "The fast, simple, cross-platform tool to manage your dev tools, environments, and tasks."
        }}
      </p>

      <!-- 4 KPI Metrics Grid -->
      <div class="kpi-grid" id="dashboard-section-kpi">
        <div class="kpi-card" @click="$emit('select-tab', 'tools')">
          <span class="kpi-icon"><TechIcon id="workbench:tools" :size="30" /></span>
          <div class="kpi-info">
            <span class="kpi-val"
              >{{ installedToolsCount !== undefined ? installedToolsCount : 5 }}
              {{ t("explorer.items_suffix") || "items" }}</span
            >
            <span class="kpi-lbl">{{ t("dashboard.kpi_tools") || "Installed Tools" }}</span>
          </div>
        </div>

        <div class="kpi-card" @click="$emit('select-tab', 'tasks')">
          <span class="kpi-icon"><TechIcon id="workbench:tasks" :size="30" /></span>
          <div class="kpi-info">
            <span class="kpi-val"
              >{{ totalTasksCount !== undefined ? totalTasksCount : 4 }} {{ t("tasks.count_suffix") || "Tasks" }}</span
            >
            <span class="kpi-lbl">{{ t("dashboard.kpi_tasks") || "Tasks" }}</span>
          </div>
        </div>

        <div class="kpi-card" @click="handleTestNetwork">
          <span class="kpi-icon"><TechIcon id="workbench:runtimes" :size="30" /></span>
          <div class="kpi-info">
            <span class="kpi-val">{{
              latencyMs !== null ? latencyMs + "ms" : t("dashboard.network_optimal") || "Optimal"
            }}</span>
            <span class="kpi-lbl">{{ t("dashboard.kpi_mirror") || "Mirror Network" }}</span>
          </div>
        </div>

        <div class="kpi-card" @click="$emit('select-tab', 'doctor')">
          <span class="kpi-icon"><TechIcon id="workbench:doctor" :size="30" /></span>
          <div class="kpi-info">
            <span class="kpi-val text-success">100%</span>
            <span class="kpi-lbl">{{ t("dashboard.kpi_security") }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 2 Column Layout: Runtime Details & Quick Actions -->
    <div class="dash-columns">
      <!-- Left Card: Runtime Environment Specs -->
      <div class="glass-card dash-card" id="dashboard-section-specs">
        <div class="card-title-row">
          <div class="card-icon cyan"><TechIcon id="workbench:dashboard" :size="24" /></div>
          <div>
            <h3>{{ t("hello.system_title") || "Runtime Environment" }}</h3>
            <span class="sub-muted">{{
              t("hello.system_desc") || "Operating system, architecture, and Go runtime specs"
            }}</span>
          </div>
        </div>

        <div class="specs-table">
          <div class="spec-row">
            <span class="spec-name">{{ t("hello.os_platform") || "Platform & Arch" }}</span>
            <span class="spec-val font-mono">Linux x86_64</span>
          </div>
          <div class="spec-row">
            <span class="spec-name">{{ t("hello.go_version") || "Go Runtime Version" }}</span>
            <span class="spec-val font-mono"
              ><TechIcon id="go" :size="18" style="margin-right: 6px; vertical-align: middle" />go1.24.0
              linux/amd64</span
            >
          </div>
          <div class="spec-row">
            <span class="spec-name">{{ t("hello.app_version") || "Template Version" }}</span>
            <span class="spec-val font-mono">UniRTM v0.35.0 (dev)</span>
          </div>
          <div class="spec-row">
            <span class="spec-name">{{ t("hello.wails_engine") || "GUI Engine" }}</span>
            <span class="spec-val font-mono">Vite + Vue 3 Standalone</span>
          </div>
        </div>
      </div>

      <!-- Right Card: Network & Connectivity Diagnosis -->
      <div class="glass-card dash-card" id="dashboard-section-network">
        <div class="card-title-row">
          <div class="card-icon purple"><TechIcon id="workbench:runtimes" :size="24" /></div>
          <div>
            <h3>{{ t("hello.network_title") || "Network & Proxy" }}</h3>
            <span class="sub-muted">{{
              t("hello.network_desc") || "Proxy routing and GitHub mirror connectivity diagnostic"
            }}</span>
          </div>
        </div>

        <div class="specs-table">
          <div class="spec-row">
            <span class="spec-name">{{ t("network.col_profile") || "Active Profile" }}</span>
            <span class="spec-badge badge-neutral font-mono">
              {{ networkStatus?.active_profile || "official" }}
            </span>
          </div>
          <div class="spec-row">
            <span class="spec-name">{{ t("network.tab_proxy") || "Network Proxy" }}</span>
            <span class="spec-badge font-mono" :class="networkStatus?.proxy?.enabled ? 'badge-on' : 'badge-neutral'">
              {{
                networkStatus?.proxy?.enabled
                  ? networkStatus.proxy.url || t("network.proxy_enabled") || "Enable Network Proxy"
                  : t("network.direct_mode") || "Direct Mode"
              }}
            </span>
          </div>
          <div class="spec-row">
            <span class="spec-name">{{ t("hello.ping_latency") || "Round-Trip Latency (RTT)" }}</span>
            <span class="spec-val font-mono" :class="{ 'text-success': networkConnected }">
              {{ latencyDisplay }}
            </span>
          </div>
        </div>

        <div class="card-action-bar">
          <button class="dash-btn primary" :disabled="testingNetwork" @click="handleTestNetwork">
            <span v-if="testingNetwork" class="btn-content-loading">
              <span class="spin-icon">
                <TechIcon id="workbench:sync" :size="16" />
              </span>
              <span>{{ t("hello.testing") || "Testing..." }}</span>
            </span>
            <span v-else class="btn-content">
              <TechIcon id="workbench:tasks" :size="16" :colored="false" />
              <span>{{ t("hello.test_speed_btn") || "Test Connectivity" }}</span>
            </span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { t } from "../i18n";
import TechIcon from "../components/common/TechIcon.vue";

defineProps<{
  appConfig?: any;
}>();

const emit = defineEmits<{
  (e: "select-tab", tabId: string): void;
  (e: "log-event", level: "INFO" | "SUCCESS" | "WARN" | "ERROR", msg: string, detail?: string): void;
}>();

const testingNetwork = ref(false);
const latencyMs = ref<number | null>(null);
const networkConnected = ref<boolean | null>(null);
const networkStatus = ref<any>(null);

const latencyDisplay = computed(() => {
  if (testingNetwork.value) return t("network.testing") || "Testing...";
  if (latencyMs.value !== null) return `${latencyMs.value} ms`;
  return t("network.untested") || "Untested";
});

async function loadNetworkStatus() {
  try {
    const res = await fetch("/api/network/status");
    if (res.ok) {
      const data = await res.json();
      networkStatus.value = data.status;
    }
  } catch (err) {
    // ignore
  }
}

onMounted(() => {
  loadNetworkStatus();
});

async function handleTestNetwork() {
  testingNetwork.value = true;
  emit("log-event", "INFO", "Initiating network diagnostic probe...");

  try {
    const res = await fetch("/api/network/ping");
    if (res.ok) {
      const data = await res.json();
      const results = data.results || [];
      const okItem = results.find((r: any) => r.ok && r.latency_ms > 0);
      if (okItem) {
        latencyMs.value = okItem.latency_ms;
        networkConnected.value = true;
        emit(
          "log-event",
          "SUCCESS",
          `Network latency probe completed: ${okItem.latency_ms}ms`,
          `Target: ${okItem.endpoint}`
        );
      } else {
        networkConnected.value = false;
        latencyMs.value = null;
        emit("log-event", "WARN", "Network latency probe completed with errors");
      }
    }
  } catch (err: any) {
    networkConnected.value = false;
    emit("log-event", "ERROR", `Network diagnostic probe failed: ${err.message}`);
  } finally {
    testingNetwork.value = false;
  }
}
</script>

<style scoped>
.view-container {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  padding: 1.25rem;
  height: 100%;
  overflow-y: auto;
  box-sizing: border-box;
}

.hero-cockpit {
  padding: 1.5rem 1.75rem;
  border-radius: 14px;
}

.hero-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  background: rgba(0, 229, 255, 0.12);
  border: 1px solid rgba(0, 229, 255, 0.3);
  color: var(--accent-cyan);
  padding: 0.2rem 0.65rem;
  border-radius: 9999px;
  font-size: 0.78rem;
  font-weight: 600;
  margin-bottom: 0.75rem;
}

.hero-title {
  font-size: 1.45rem;
  font-weight: 700;
  margin: 0 0 0.45rem;
  color: var(--text-main);
  background: linear-gradient(135deg, var(--text-main) 60%, var(--accent-cyan));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.hero-desc {
  font-size: 0.88rem;
  color: var(--text-muted);
  margin: 0 0 1.25rem;
  line-height: 1.5;
  max-width: 780px;
}

.kpi-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 0.85rem;
}

.kpi-card {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  border-radius: 10px;
  padding: 0.85rem 1rem;
  cursor: pointer;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

.kpi-card:hover {
  border-color: var(--card-border-active);
  transform: translateY(-2px);
  background: rgba(0, 229, 255, 0.05);
}

.kpi-icon {
  font-size: 1.6rem;
}

.kpi-info {
  display: flex;
  flex-direction: column;
}

.kpi-val {
  font-size: 1.15rem;
  font-weight: 700;
  color: var(--text-main);
}

.kpi-lbl {
  font-size: 0.72rem;
  color: var(--text-muted);
}

.dash-columns {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(360px, 1fr));
  gap: 1.25rem;
}

.dash-card {
  display: flex;
  flex-direction: column;
  padding: 1.25rem;
  border-radius: 12px;
}

.card-title-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 1.1rem;
}

.card-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 10px;
  font-size: 1.3rem;
  flex-shrink: 0;
}

.card-icon.cyan {
  background: rgba(0, 229, 255, 0.15);
}

.card-icon.purple {
  background: rgba(157, 78, 221, 0.15);
}

.card-title-row h3 {
  font-size: 0.95rem;
  font-weight: 600;
  margin: 0;
  color: var(--text-main);
}

.sub-muted {
  font-size: 0.75rem;
  color: var(--text-muted);
}

.specs-table {
  display: flex;
  flex-direction: column;
  gap: 0.65rem;
  flex: 1;
}

.spec-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.35rem 0;
  border-bottom: 1px dashed var(--card-border);
  font-size: 0.82rem;
}

.spec-name {
  color: var(--text-muted);
}

.spec-val {
  color: var(--text-main);
  font-weight: 500;
}

.spec-badge {
  font-size: 0.72rem;
  padding: 0.15rem 0.45rem;
  border-radius: 4px;
}

.badge-neutral {
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  color: var(--accent-cyan);
}

.badge-on {
  background: rgba(16, 185, 129, 0.15);
  border: 1px solid rgba(16, 185, 129, 0.4);
  color: var(--success);
}

.card-action-bar {
  margin-top: 1.1rem;
}

.dash-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.5rem 1.15rem;
  min-width: 148px;
  border-radius: 8px;
  font-size: 0.82rem;
  font-weight: 600;
  cursor: pointer;
  border: 1px solid transparent;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  user-select: none;
  white-space: nowrap;
  box-sizing: border-box;
  line-height: 1.4;
}

.btn-content,
.btn-content-loading {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

/* Dark Theme Primary Button (Default) */
.dash-btn.primary {
  background: linear-gradient(135deg, #0ea5e9 0%, #0284c7 100%);
  color: #ffffff !important;
  border: 1px solid rgba(255, 255, 255, 0.18);
  box-shadow: 0 2px 8px rgba(2, 132, 199, 0.28);
}

.dash-btn.primary:hover:not(:disabled) {
  background: linear-gradient(135deg, #38bdf8 0%, #0ea5e9 100%);
  border-color: #38bdf8;
  color: #ffffff !important;
  box-shadow: 0 4px 14px rgba(2, 132, 199, 0.42);
  transform: translateY(-1px);
}

.dash-btn.primary:active:not(:disabled) {
  transform: translateY(0);
}

.dash-btn:disabled {
  opacity: 0.55;
  cursor: not-allowed;
  transform: none !important;
  box-shadow: none !important;
}

/* Light Theme Adaptations */
:root[data-theme="light"] .dash-btn.primary,
[data-theme="light"] .dash-btn.primary {
  background: linear-gradient(135deg, #0284c7 0%, #0369a1 100%);
  color: #ffffff !important;
  border: 1px solid rgba(2, 132, 199, 0.4);
  box-shadow: 0 2px 8px rgba(2, 132, 199, 0.2);
}

:root[data-theme="light"] .dash-btn.primary:hover:not(:disabled),
[data-theme="light"] .dash-btn.primary:hover:not(:disabled) {
  background: linear-gradient(135deg, #0369a1 0%, #075985 100%);
  border-color: #0284c7;
  color: #ffffff !important;
  box-shadow: 0 4px 14px rgba(2, 132, 199, 0.32);
  transform: translateY(-1px);
}

.text-success {
  color: var(--success);
}

.highlight-pulse {
  outline: 2px solid var(--color-primary, #00f0ff) !important;
  box-shadow: 0 0 24px rgba(0, 240, 255, 0.45) !important;
  transition: all 0.3s ease !important;
}
</style>
