<!-- Copyright (c) 2026 SnowdreamTech. All rights reserved. -->
<!-- Licensed under the MIT License. See LICENSE file in the project root for full license information. -->

<template>
  <div class="view-container doctor-view" id="view-doctor" @contextmenu.prevent>
    <!-- View Header -->
    <div class="view-header-row">
      <div>
        <h2 class="view-title" style="display: flex; align-items: center; gap: 8px">
          <TechIcon id="workbench:doctor" :size="20" />
          <span>{{ t("doctor.title") || "UniRTM Doctor Diagnostics" }}</span>
        </h2>
        <p class="view-desc">
          {{
            t("doctor.subtitle") ||
            "UniRTM Doctor diagnostic engine deeply inspects core runtimes, shims routing, disk footprints, cloud API quotas, and security posture."
          }}
        </p>
      </div>
      <div class="view-actions">
        <!-- Console output button -->
        <button
          class="doctor-action-btn secondary"
          @click="openConsoleLogs"
          :title="t('doctor.console_logs_tooltip') || 'View detailed diagnostic logs in bottom console drawer'"
        >
          <TechIcon id="workbench:terminal" :size="14" />
          <span>{{ t("doctor.console_logs") || "Console Logs" }}</span>
        </button>

        <!-- Export diagnostic report -->
        <button
          class="doctor-action-btn secondary"
          :disabled="isLoading || !report"
          @click="copyDiagnosticReport"
          :title="t('doctor.export_report_tooltip') || 'Copy full markdown diagnostic report to clipboard'"
        >
          <TechIcon id="workbench:copy" :size="14" />
          <span>{{ t("doctor.export_report") || "Export Report" }}</span>
        </button>

        <!-- Doctor run button -->
        <button
          class="doctor-action-btn primary"
          :disabled="isLoading"
          @click="fetchDoctorReport"
          :title="t('doctor.run_doctor_tooltip') || 'Run comprehensive system health check'"
        >
          <TechIcon id="workbench:sync" :size="14" :class="{ 'spin-icon': isLoading }" />
          <span>{{ isLoading ? t("doctor.running") || "Diagnosing..." : t("doctor.run_doctor") || "Run Doctor" }}</span>
        </button>
      </div>
    </div>

    <!-- Health Score Banner -->
    <div class="glass-card score-banner" :class="bannerStatusClass">
      <div class="score-left">
        <div class="score-circle" :class="circleScoreClass">
          <span v-if="isLoading" class="score-loading-icon spin-icon">
            <TechIcon id="workbench:sync" :size="28" />
          </span>
          <template v-else>
            <span class="score-number">{{ report ? report.score : "--" }}</span>
            <span class="score-unit">/ 100</span>
          </template>
        </div>
        <div class="score-text">
          <h3 class="score-title">
            {{
              isLoading
                ? t("doctor.running_desc") || "Running comprehensive UniRTM environment health diagnosis..."
                : report?.score === 100
                  ? t("doctor.healthy") || "Environment Fully Healthy"
                  : report && report.score >= 80
                    ? t("doctor.warning_title") || "Environment Ready with Optimization Suggestions"
                    : t("doctor.warning_desc") || "Configuration issues detected; review suggestions below to resolve."
            }}
          </h3>
          <p class="score-sub">
            <template v-if="report">
              {{ t("doctor.core_version") || "Core Version" }}: {{ report.core.version }} •
              {{ t("doctor.shims_status") || "Runtime Shims" }}:
              {{
                report.core.shims_on_path
                  ? t("doctor.shims_ready") || "Injected into PATH"
                  : t("doctor.shims_not_ready") || "Not in PATH"
              }}
              • {{ t("doctor.database") || "Database" }}:
              {{
                report.health.database_ok ? t("doctor.status_ok") || "Normal" : t("doctor.status_error") || "Abnormal"
              }}
              • {{ t("doctor.proxy_status") || "Network Proxy" }}:
              {{
                report.health.proxy_configured
                  ? t("doctor.proxy_configured") || "Proxy Configured"
                  : t("doctor.proxy_direct") || "Direct Connection"
              }}
              • {{ t("doctor.google_status") || "Global Uplink" }}:
              {{
                report.health.google_connected
                  ? t("doctor.google_connected") || "Google Reachable"
                  : t("doctor.direct_mode") || "Direct Mode"
              }}
              • {{ t("doctor.github_status") || "GitHub API" }}:
              {{
                report.health.github_connected
                  ? `${t("doctor.connected") || "Connected"} (HTTP ${report.health.github_status_code})`
                  : t("doctor.not_connected") || "Disconnected"
              }}
            </template>
            <template v-else>
              {{ t("doctor.parsing_env") || "Parsing local runtime and environment state..." }}
            </template>
          </p>
        </div>
      </div>
      <div v-if="report && lastCheckedTime" class="score-right-meta">
        <span class="meta-label">{{ t("doctor.last_check") || "Last Check Time" }}</span>
        <span class="meta-value font-mono">{{ lastCheckedTime }}</span>
      </div>
    </div>

    <!-- Fix Suggestions Banner (Actionable) -->
    <div v-if="report && report.suggestions && report.suggestions.length > 0" class="suggestions-container">
      <div v-for="sug in report.suggestions" :key="sug.id" class="glass-card suggestion-card" :class="sug.severity">
        <div class="sug-icon-box">
          <span v-if="sug.severity === 'error'" class="codicon codicon-error sug-icon error"></span>
          <span v-else class="codicon codicon-warning sug-icon warning"></span>
        </div>
        <div class="sug-body">
          <div class="sug-message font-mono">{{ sug.message }}</div>
          <div v-if="sug.command" class="sug-cmd-box font-mono">
            <span class="cmd-prompt">$</span>
            <span class="cmd-text">{{ sug.command }}</span>
          </div>
        </div>
        <div v-if="sug.command" class="sug-actions">
          <button
            class="sug-action-btn secondary"
            @click="copyCommand(sug.command)"
            :title="t('doctor.copy_cmd_tooltip') || 'Copy fix command'"
          >
            <TechIcon id="workbench:copy" :size="13" />
            <span>{{ t("doctor.copy_cmd") || "Copy Command" }}</span>
          </button>
          <button
            class="sug-action-btn primary"
            @click="runFixCommand(sug.command)"
            :title="t('doctor.run_in_terminal_tooltip') || 'Send command to terminal and run'"
          >
            <TechIcon id="workbench:terminal" :size="13" />
            <span>{{ t("doctor.run_in_terminal") || "Run in Terminal" }}</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Diagnostic Details Grid (5 Core Dimensions) -->
    <div v-if="report" class="doctor-grid">
      <!-- 1. Core & Shell Context -->
      <div id="doctor-section-core" class="glass-card doctor-card">
        <div class="card-header">
          <div class="card-title-group">
            <span class="card-icon"><TechIcon id="workbench:dashboard" :size="16" /></span>
            <span class="card-title">{{ t("doctor.dim_core") || "Core & Shell" }}</span>
          </div>
          <span class="badge-status" :class="report.core.shims_on_path ? 'pass' : 'warn'">
            {{
              report.core.shims_on_path
                ? t("doctor.shims_badge_ready") || "Shims Ready"
                : t("doctor.shims_badge_not_ready") || "Shims Need Config"
            }}
          </span>
        </div>
        <div class="card-kv-list">
          <div class="kv-row">
            <span class="kv-key">{{ t("doctor.unirtm_version") || "UniRTM Version" }}</span>
            <span class="kv-val font-mono">{{ report.core.version }} ({{ report.core.commit }})</span>
          </div>
          <div class="kv-row">
            <span class="kv-key">{{ t("doctor.target_arch") || "Target Arch" }}</span>
            <span class="kv-val font-mono">{{ report.build.target }} ({{ report.build.go_version }})</span>
          </div>
          <div class="kv-row">
            <span class="kv-key">{{ t("doctor.current_shell") || "Current Shell" }}</span>
            <span class="kv-val font-mono">{{ report.context.shell }} {{ report.context.shell_version }}</span>
          </div>
          <div class="kv-row">
            <span class="kv-key">{{ t("doctor.default_editor") || "Default Editor" }}</span>
            <span class="kv-val font-mono">{{ report.context.editor }} ({{ report.context.editor_source }})</span>
          </div>
          <div class="kv-row">
            <span class="kv-key">{{ t("doctor.active_env") || "Active Profile" }}</span>
            <span class="kv-val font-mono highlight">{{ report.context.active_env }}</span>
          </div>
        </div>
      </div>

      <!-- 2. Cloud Network & GitHub Quota -->
      <div id="doctor-section-network" class="glass-card doctor-card">
        <div class="card-header">
          <div class="card-title-group">
            <span class="card-icon"><TechIcon id="workbench:runtimes" :size="16" /></span>
            <span class="card-title">{{ t("doctor.dim_network") || "Network & Quotas" }}</span>
          </div>
          <span
            class="badge-status"
            :class="
              report.health.github_connected && report.health.google_connected
                ? 'pass'
                : report.health.github_connected
                  ? 'pass'
                  : 'fail'
            "
          >
            {{
              report.health.github_connected && report.health.google_connected
                ? t("doctor.network_global_ready") || "Global Connectivity Optimal"
                : report.health.github_connected
                  ? t("doctor.network_direct_ready") || "Direct Access Ready"
                  : t("doctor.network_error") || "Network Connection Issue"
            }}
          </span>
        </div>
        <div class="card-kv-list">
          <!-- 1. Network Proxy Environment -->
          <div class="kv-row">
            <span class="kv-key">{{ t("doctor.proxy_env") || "Network Proxy Environment" }}</span>
            <span class="kv-val font-mono" :class="{ highlight: !!report.health.proxy_configured }">
              {{
                report.health.proxy_configured ||
                t("doctor.proxy_not_detected") ||
                "Direct Connection (No proxy detected)"
              }}
            </span>
          </div>
          <!-- 2. Google Connectivity -->
          <div class="kv-row">
            <span class="kv-key">{{ t("doctor.google_connectivity") || "Google Connectivity" }}</span>
            <span
              class="kv-val font-mono"
              :style="{ color: report.health.google_connected ? '#10b981' : 'var(--text-muted)' }"
            >
              ●
              {{
                report.health.google_connected
                  ? `HTTP ${report.health.google_status_code || 204} (${t("doctor.connected") || "Connected"} · ${report.health.google_latency_ms || "<100"}ms)`
                  : t("doctor.google_not_connected") || "Unreachable (Proxy required)"
              }}
            </span>
          </div>
          <!-- 3. GitHub Connectivity -->
          <div class="kv-row">
            <span class="kv-key">{{ t("doctor.github_connectivity") || "GitHub Connectivity" }}</span>
            <span class="kv-val font-mono" :style="{ color: report.health.github_connected ? '#10b981' : '#ef4444' }">
              ●
              {{
                report.health.github_connected
                  ? `HTTP ${report.health.github_status_code} (${t("doctor.connected") || "Connected"})`
                  : t("doctor.not_connected") || "Disconnected"
              }}
            </span>
          </div>
          <!-- 4. GitHub API Quota -->
          <div v-if="report.health.github_rate_limit > 0" class="rate-limit-section">
            <div class="rate-limit-header font-mono">
              <span>{{ t("doctor.github_quota") || "GitHub API Remaining Quota" }}</span>
              <span
                >{{ report.health.github_rate_remaining.toLocaleString() }} /
                {{ report.health.github_rate_limit.toLocaleString() }}</span
              >
            </div>
            <div class="rate-bar-track">
              <div
                class="rate-bar-fill"
                :style="{
                  width: `${rateLimitPercent}%`,
                  backgroundColor: rateLimitColor,
                }"
              ></div>
            </div>
            <div v-if="report.health.github_rate_reset_text" class="rate-reset-hint">
              {{
                t("doctor.github_quota_reset", { time: report.health.github_rate_reset_text }) ||
                `Quota resets in approx ${report.health.github_rate_reset_text}`
              }}
            </div>
          </div>
          <div v-else class="kv-row">
            <span class="kv-key">{{ t("doctor.github_quota_config") || "GitHub API Quota Configuration" }}</span>
            <span class="kv-val font-mono">{{ t("doctor.anonymous_quota") || "Anonymous / Quota not returned" }}</span>
          </div>
        </div>
      </div>

      <!-- 3. Storage & Footprint -->
      <div id="doctor-section-directories" class="glass-card doctor-card">
        <div class="card-header">
          <div class="card-title-group">
            <span class="card-icon"><TechIcon id="workbench:folder" :size="16" /></span>
            <span class="card-title">{{ t("doctor.dim_directories") || "Storage & Directories" }}</span>
          </div>
        </div>
        <div class="card-dirs-list">
          <div v-for="dir in report.directories" :key="dir.type" class="dir-item font-mono">
            <div class="dir-type-side">
              <span class="dir-badge">{{ dir.type }}</span>
              <span class="dir-size" :title="t('doctor.dir_size_title', { size: dir.size }) || `Size: ${dir.size}`">{{
                dir.size
              }}</span>
            </div>
            <div class="dir-path" :title="dir.path">{{ dir.path }}</div>
          </div>
        </div>
      </div>

      <!-- 4. Database & Configuration -->
      <div id="doctor-section-database" class="glass-card doctor-card">
        <div class="card-header">
          <div class="card-title-group">
            <span class="card-icon"><TechIcon id="workbench:settings" :size="16" /></span>
            <span class="card-title">{{ t("doctor.dim_database") || "Database & Config" }}</span>
          </div>
          <span class="badge-status" :class="report.health.database_ok ? 'pass' : 'fail'">
            {{ report.health.database_ok ? "SQLite OK" : t("doctor.status_error") || "Error" }}
          </span>
        </div>
        <div class="card-kv-list">
          <div class="kv-row">
            <span class="kv-key">{{ t("doctor.sqlite_db") || "SQLite Database" }}</span>
            <span class="kv-val font-mono" :title="report.health.database_path">
              {{
                report.health.database_ok
                  ? t("doctor.connected_size", { size: report.health.database_size }) ||
                    `Connected (${report.health.database_size})`
                  : report.health.database_error
              }}
            </span>
          </div>
          <div class="kv-config-section">
            <span class="kv-key">{{ t("doctor.active_configs") || "Active Config Files" }}</span>
            <div class="config-files-box font-mono">
              <template v-if="report.configs && report.configs.length > 0">
                <div v-for="c in report.configs" :key="c.path" class="config-file-row">
                  <span class="cfg-scope" :class="c.scope">{{ c.scope }}</span>
                  <span class="cfg-path" :title="c.path">{{ c.path }}</span>
                </div>
              </template>
              <div v-else style="color: var(--text-muted); font-size: 0.78rem">
                {{ t("doctor.default_config_hint") || "Default config (No active TOML file detected)" }}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { t } from "../i18n";
import TechIcon from "../components/common/TechIcon.vue";
import { useToast } from "../composables/useToast";
import { useWorkspaceBus } from "../composables/useWorkspaceBus";

interface DoctorReport {
  timestamp: number;
  score: number;
  status: "healthy" | "warning" | "critical";
  core: {
    version: string;
    commit: string;
    activated: boolean;
    shims_on_path: boolean;
    shims_dir: string;
    trust_status: string;
  };
  build: {
    target: string;
    go_version: string;
    built: string;
    features: string[];
  };
  context: {
    shell: string;
    shell_version: string;
    editor: string;
    editor_source: string;
    work_dir: string;
    active_env: string;
  };
  directories: Array<{
    type: string;
    path: string;
    size: string;
    size_bytes?: number;
  }>;
  configs: Array<{
    path: string;
    loaded: boolean;
    scope: string;
  }>;
  settings: Array<{
    key: string;
    value: string;
  }>;
  backends: string[];
  tools: Array<{
    name: string;
    version: string;
    backend: string;
    installed: boolean;
  }>;
  tasks: Array<{
    name: string;
    command: string;
    description: string;
  }>;
  health: {
    database_path: string;
    database_size: string;
    database_ok: boolean;
    database_error?: string;
    github_connected: boolean;
    github_status_code: number;
    github_rate_limit: number;
    github_rate_remaining: number;
    github_rate_reset: number;
    github_rate_reset_text: string;
    google_connected?: boolean;
    google_status_code?: number;
    google_latency_ms?: number;
    proxy_configured?: string;
  };
  suggestions: Array<{
    id: string;
    severity: "warning" | "error" | "info";
    message: string;
    command?: string;
    actionable: boolean;
  }>;
}

const emit = defineEmits<{
  (
    e: "log-event",
    level: "INFO" | "SUCCESS" | "WARN" | "ERROR",
    msg: string,
    detail?: string,
    category?: "output" | "tasks" | "doctor"
  ): void;
  (e: "open-drawer", tab?: "output" | "tasks" | "doctor"): void;
}>();

const { showToast } = useToast();
const { sendToTerminal, notifyDoctorChanged } = useWorkspaceBus();

const report = ref<DoctorReport | null>(null);
const isLoading = ref(false);
const lastCheckedTime = ref<string>("");

// Compute visual score banner level
const bannerStatusClass = computed(() => {
  if (!report.value) return "";
  if (report.value.score === 100) return "status-perfect";
  if (report.value.score >= 80) return "status-good";
  if (report.value.score >= 50) return "status-warning";
  return "status-critical";
});

const circleScoreClass = computed(() => {
  if (!report.value) return "";
  if (report.value.score === 100) return "score-perfect";
  if (report.value.score >= 80) return "score-good";
  if (report.value.score >= 50) return "score-warning";
  return "score-critical";
});

// Compute GitHub API remaining quota percentage
const rateLimitPercent = computed(() => {
  if (!report.value || report.value.health.github_rate_limit <= 0) return 0;
  const pct = (report.value.health.github_rate_remaining / report.value.health.github_rate_limit) * 100;
  return Math.max(0, Math.min(100, pct));
});

const rateLimitColor = computed(() => {
  if (rateLimitPercent.value > 50) return "#10b981";
  if (rateLimitPercent.value > 20) return "#f59e0b";
  return "#ef4444";
});

async function fetchDoctorReport() {
  isLoading.value = true;
  emit("log-event", "INFO", "[Doctor] Triggering UniRTM Doctor diagnostics...", undefined, "doctor");
  try {
    const res = await fetch("/api/doctor");
    const contentType = res.headers.get("content-type") || "";
    if (res.ok && contentType.includes("json")) {
      const data: DoctorReport = await res.json();
      report.value = data;
      const d = new Date(data.timestamp ? data.timestamp * 1000 : Date.now());
      lastCheckedTime.value = d.toLocaleTimeString();
      notifyDoctorChanged();
      emit(
        "log-event",
        data.status === "healthy" ? "SUCCESS" : "WARN",
        `[Doctor] Diagnosis complete: Score ${data.score}/100, ${data.suggestions?.length || 0} suggestion(s)`,
        undefined,
        "doctor"
      );
    } else {
      showToast("Diagnostic API not ready or invalid JSON response", "error");
      emit("log-event", "ERROR", `[Doctor] Request failed (HTTP ${res.status})`, undefined, "doctor");
    }
  } catch (err: any) {
    showToast(err.message, "error");
    emit("log-event", "ERROR", `[Doctor] Error: ${err.message}`, undefined, "doctor");
  } finally {
    isLoading.value = false;
  }
}

function openConsoleLogs() {
  emit("open-drawer", "doctor");
}

function copyCommand(cmd: string) {
  if (navigator?.clipboard?.writeText) {
    navigator.clipboard
      .writeText(cmd)
      .then(() => {
        showToast(t("doctor.copied_command") || "Copied fix command to clipboard", "success");
      })
      .catch(() => {
        showToast(t("doctor.copy_failed") || "Failed to copy, please try again", "error");
      });
  }
}

function runFixCommand(cmd: string) {
  emit("open-drawer", "output");
  sendToTerminal({ text: cmd, addNewline: true });
  showToast(t("doctor.fix_sent") || "Command sent to terminal for execution", "success");
}

function copyDiagnosticReport() {
  if (!report.value) return;
  const r = report.value;
  const lines = [
    `# UniRTM Doctor Diagnostic Report`,
    `- Date: ${new Date().toISOString()}`,
    `- Health Score: ${r.score}/100 (${r.status})`,
    `- Core: UniRTM ${r.core.version} (${r.core.commit}), Shims On PATH: ${r.core.shims_on_path}`,
    `- Platform: ${r.build.target}, Go: ${r.build.go_version}`,
    `- Shell: ${r.context.shell} (${r.context.shell_version})`,
    `- Database: ${r.health.database_ok ? "OK" : "Error"} (${r.health.database_size})`,
    `- GitHub API: ${r.health.github_connected ? `Connected (${r.health.github_rate_remaining}/${r.health.github_rate_limit})` : "Disconnected"}`,
    `- Tasks Discovered: ${r.tasks.length}`,
    `- Suggestions: ${r.suggestions.length}`,
  ];
  if (r.suggestions.length > 0) {
    lines.push(`\n## Fix Suggestions:`);
    r.suggestions.forEach((s, idx) => {
      lines.push(`${idx + 1}. [${s.severity.toUpperCase()}] ${s.message} (${s.command || "N/A"})`);
    });
  }

  const markdown = lines.join("\n");
  if (navigator?.clipboard?.writeText) {
    navigator.clipboard.writeText(markdown).then(() => {
      showToast(t("doctor.copied_report") || "Copied full markdown diagnostic report to clipboard", "success");
    });
  }
}

onMounted(() => {
  fetchDoctorReport();
});
</script>

<style scoped>
.doctor-view {
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
  gap: 1rem;
  flex-wrap: wrap;
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

.view-actions {
  display: flex;
  align-items: center;
  gap: 0.6rem;
}

.doctor-action-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.45rem 0.95rem;
  border-radius: 6px;
  font-size: 0.82rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
  border: 1px solid transparent;
}

/* Dark mode primary button (Run Doctor) */
.doctor-action-btn.primary {
  background: linear-gradient(135deg, #00f0ff 0%, #38bdf8 100%);
  color: #050e1d;
  font-weight: 700;
  border: 1px solid rgba(0, 240, 255, 0.5);
  box-shadow: 0 2px 10px rgba(0, 229, 255, 0.28);
}

.doctor-action-btn.primary:hover:not(:disabled) {
  transform: translateY(-1px);
  background: linear-gradient(135deg, #38bdf8 0%, #7dd3fc 100%);
  color: #050e1d;
  border-color: #7dd3fc;
  box-shadow: 0 4px 16px rgba(0, 229, 255, 0.45);
}

/* Light mode primary button (Run Doctor) */
:root[data-theme="light"] .doctor-action-btn.primary,
[data-theme="light"] .doctor-action-btn.primary {
  background: linear-gradient(135deg, #0284c7 0%, #0369a1 100%);
  color: #ffffff !important;
  font-weight: 600;
  border: 1px solid #0284c7;
  box-shadow: 0 2px 6px rgba(2, 132, 199, 0.25);
}

:root[data-theme="light"] .doctor-action-btn.primary:hover:not(:disabled),
[data-theme="light"] .doctor-action-btn.primary:hover:not(:disabled) {
  transform: translateY(-1px);
  background: linear-gradient(135deg, #0369a1 0%, #075985 100%);
  border-color: #0369a1;
  color: #ffffff !important;
  box-shadow: 0 4px 12px rgba(2, 132, 199, 0.35);
}

/* Dark mode secondary button */
.doctor-action-btn.secondary {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid var(--card-border);
  color: var(--text-main);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
}

.doctor-action-btn.secondary:hover:not(:disabled) {
  transform: translateY(-1px);
  background: rgba(255, 255, 255, 0.1);
  border-color: var(--card-border-active);
  color: var(--text-main);
}

/* Light mode secondary button */
:root[data-theme="light"] .doctor-action-btn.secondary,
[data-theme="light"] .doctor-action-btn.secondary {
  background: #ffffff;
  border: 1px solid #cbd5e1;
  color: #334155;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
}

:root[data-theme="light"] .doctor-action-btn.secondary:hover:not(:disabled),
[data-theme="light"] .doctor-action-btn.secondary:hover:not(:disabled) {
  transform: translateY(-1px);
  background: #f8fafc;
  border-color: #0284c7;
  color: #0284c7;
}

.doctor-action-btn:disabled {
  opacity: 0.55;
  cursor: not-allowed;
  transform: none !important;
  box-shadow: none !important;
}

/* Score Banner */
.score-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1.75rem 2rem;
  min-height: 116px;
  border-radius: 14px;
  border: 1px solid var(--card-border);
  position: relative;
  overflow: hidden;
  transition: all 0.25s ease;
}

.score-banner.status-perfect {
  border-color: rgba(16, 185, 129, 0.4);
  background: linear-gradient(135deg, rgba(16, 185, 129, 0.08), transparent 70%);
}

.score-banner.status-good {
  border-color: rgba(0, 229, 255, 0.4);
  background: linear-gradient(135deg, rgba(0, 229, 255, 0.08), transparent 70%);
}

.score-banner.status-warning {
  border-color: rgba(245, 158, 11, 0.4);
  background: linear-gradient(135deg, rgba(245, 158, 11, 0.08), transparent 70%);
}

.score-banner.status-critical {
  border-color: rgba(239, 68, 68, 0.4);
  background: linear-gradient(135deg, rgba(239, 68, 68, 0.08), transparent 70%);
}

.score-left {
  display: flex;
  align-items: center;
  gap: 1.75rem;
}

.score-circle {
  width: 82px;
  height: 82px;
  border-radius: 50%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  position: relative;
  border: 3.5px solid;
  flex-shrink: 0;
  box-sizing: border-box;
}

.score-circle.score-perfect {
  border-color: #10b981;
  color: #10b981;
  box-shadow: 0 0 18px rgba(16, 185, 129, 0.28);
}

.score-circle.score-good {
  border-color: var(--accent-cyan);
  color: var(--accent-cyan);
  box-shadow: 0 0 18px rgba(0, 229, 255, 0.28);
}

.score-circle.score-warning {
  border-color: #f59e0b;
  color: #f59e0b;
  box-shadow: 0 0 18px rgba(245, 158, 11, 0.28);
}

.score-circle.score-critical {
  border-color: #ef4444;
  color: #ef4444;
  box-shadow: 0 0 18px rgba(239, 68, 68, 0.28);
}

.score-number {
  font-size: 1.55rem;
  font-weight: 800;
  font-family: monospace;
  line-height: 1;
  letter-spacing: -0.03em;
}

.score-unit {
  font-size: 0.65rem;
  font-weight: 600;
  color: var(--text-muted);
  line-height: 1;
  margin-top: 3px;
  letter-spacing: 0.02em;
  opacity: 0.85;
}

.score-text {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.score-title {
  margin: 0;
  font-size: 1.2rem;
  font-weight: 700;
  color: var(--text-main);
  letter-spacing: -0.01em;
}

.score-sub {
  margin: 0;
  font-size: 0.85rem;
  color: var(--text-muted);
  line-height: 1.6;
}

.score-right-meta {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 0.35rem;
}

.meta-label {
  font-size: 0.74rem;
  color: var(--text-muted);
}

.meta-value {
  font-size: 0.95rem;
  color: var(--text-main);
  font-weight: 600;
}

/* Actionable Suggestions */
.suggestions-container {
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
}

.suggestion-card {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 0.9rem 1.2rem;
  border-radius: 8px;
  border: 1px solid rgba(245, 158, 11, 0.35);
  background: rgba(245, 158, 11, 0.05);
}

.suggestion-card.error {
  border-color: rgba(239, 68, 68, 0.4);
  background: rgba(239, 68, 68, 0.06);
}

.sug-icon-box {
  font-size: 1.25rem;
  line-height: 1;
}

.sug-icon.warning {
  color: #f59e0b;
}

.sug-icon.error {
  color: #ef4444;
}

.sug-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  min-width: 0;
}

.sug-message {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text-main);
}

.sug-cmd-box {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 0.82rem;
  color: var(--accent-cyan);
}

.cmd-prompt {
  opacity: 0.6;
}

.sug-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.sug-action-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 0.35rem 0.75rem;
  border-radius: 6px;
  font-size: 0.78rem;
  font-weight: 600;
  cursor: pointer;
  border: 1px solid transparent;
  transition: all 0.15s ease;
}

.sug-action-btn.primary {
  background: rgba(0, 229, 255, 0.15);
  color: var(--accent-cyan);
  border-color: rgba(0, 229, 255, 0.35);
}

.sug-action-btn.primary:hover {
  background: var(--accent-cyan);
  color: #070a12;
}

.sug-action-btn.secondary {
  background: var(--card-bg);
  color: var(--text-muted);
  border-color: var(--card-border);
}

.sug-action-btn.secondary:hover {
  color: var(--text-main);
  background: var(--input-bg);
}

/* 5 Core Dimensions Grid */
.doctor-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(360px, 1fr));
  gap: 1rem;
}

.doctor-card {
  padding: 1.1rem 1.25rem;
  border-radius: 10px;
  display: flex;
  flex-direction: column;
  gap: 0.85rem;
  transition: all 0.25s ease;
}

.doctor-card.highlight-pulse {
  animation: pulse-border 1.5s ease-out;
}

@keyframes pulse-border {
  0% {
    box-shadow: 0 0 0 2px var(--accent-cyan);
    border-color: var(--accent-cyan);
  }
  50% {
    box-shadow: 0 0 16px rgba(0, 229, 255, 0.4);
    border-color: var(--accent-cyan);
  }
  100% {
    box-shadow: none;
    border-color: var(--card-border);
  }
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 0.6rem;
  border-bottom: 1px solid var(--card-border);
}

.card-title-group {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.card-icon {
  display: flex;
  align-items: center;
  color: var(--accent-cyan);
}

.card-title {
  font-size: 0.88rem;
  font-weight: 700;
  color: var(--text-main);
}

.badge-status {
  font-size: 0.7rem;
  padding: 0.15rem 0.5rem;
  border-radius: 10px;
  font-weight: 600;
}

.badge-status.pass {
  background: rgba(16, 185, 129, 0.15);
  color: #10b981;
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.badge-status.warn {
  background: rgba(245, 158, 11, 0.15);
  color: #f59e0b;
  border: 1px solid rgba(245, 158, 11, 0.3);
}

.badge-status.fail {
  background: rgba(239, 68, 68, 0.15);
  color: #ef4444;
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.card-kv-list {
  display: flex;
  flex-direction: column;
  gap: 0.55rem;
}

.kv-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.8rem;
  font-size: 0.82rem;
}

.kv-key {
  color: var(--text-muted);
}

.kv-val {
  color: var(--text-main);
  text-align: right;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.kv-val.highlight {
  color: var(--accent-cyan);
  font-weight: 600;
}

/* Rate Limit Bar */
.rate-limit-section {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  padding-top: 0.3rem;
}

.rate-limit-header {
  display: flex;
  justify-content: space-between;
  font-size: 0.78rem;
  color: var(--text-muted);
}

.rate-bar-track {
  width: 100%;
  height: 6px;
  background: var(--input-bg);
  border-radius: 3px;
  overflow: hidden;
}

.rate-bar-fill {
  height: 100%;
  border-radius: 3px;
  transition: width 0.3s ease;
}

.rate-reset-hint {
  font-size: 0.7rem;
  color: var(--text-muted);
  text-align: right;
}

/* Directories List */
.card-dirs-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.dir-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.6rem;
  font-size: 0.8rem;
  padding: 0.4rem 0.6rem;
  background: rgba(255, 255, 255, 0.02);
  border-radius: 6px;
  border: 1px solid var(--card-border);
}

.dir-type-side {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  flex-shrink: 0;
}

.dir-badge {
  font-size: 0.68rem;
  text-transform: uppercase;
  color: var(--accent-cyan);
  background: rgba(0, 229, 255, 0.1);
  padding: 0.1rem 0.35rem;
  border-radius: 3px;
}

.dir-size {
  color: var(--text-main);
  font-weight: 600;
}

.dir-path {
  color: var(--text-muted);
  font-size: 0.75rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* Config Section */
.kv-config-section {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  margin-top: 0.3rem;
}

.config-files-box {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.config-file-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.78rem;
}

.cfg-scope {
  font-size: 0.65rem;
  padding: 0.05rem 0.3rem;
  border-radius: 3px;
  text-transform: uppercase;
}

.cfg-scope.workspace {
  background: rgba(0, 229, 255, 0.15);
  color: var(--accent-cyan);
}

.cfg-scope.global {
  background: rgba(147, 51, 234, 0.15);
  color: #a855f7;
}

.cfg-path {
  color: var(--text-muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.spin-icon {
  animation: spin 1s linear infinite;
  display: inline-flex;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>
