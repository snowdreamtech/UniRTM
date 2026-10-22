<!-- Copyright (c) 2026 SnowdreamTech. All rights reserved. -->
<!-- Licensed under the MIT License. See LICENSE file in the project root for full license information. -->

<template>
  <div class="view-container doctor-view" id="view-doctor">
    <!-- View Header -->
    <div class="view-header-row">
      <div>
        <h2 class="view-title" style="display: flex; align-items: center; gap: 8px">
          <TechIcon id="workbench:doctor" :size="20" />
          <span>{{ t("doctor.title") || "项目与环境全面体检" }}</span>
        </h2>
        <p class="view-desc">
          基于 UniRTM Doctor 系统诊断引擎，全景探测核心运行时、Shims 路由、磁盘足迹、云端 API 配额及安全配置。
        </p>
      </div>
      <div class="view-actions">
        <!-- 查看原始控制台输出 -->
        <button class="doctor-action-btn secondary" @click="openConsoleLogs" title="在下方控制台抽屉中查看详细体检日志">
          <TechIcon id="workbench:terminal" :size="14" />
          <span>控制台输出</span>
        </button>

        <!-- 导出诊断报告 -->
        <button
          class="doctor-action-btn secondary"
          :disabled="isLoading || !report"
          @click="copyDiagnosticReport"
          title="复制完整的环境诊断报告 Markdown 到剪贴板"
        >
          <TechIcon id="workbench:copy" :size="14" />
          <span>导出报告</span>
        </button>

        <!-- 重新全面体检主按钮 -->
        <button
          class="doctor-action-btn primary"
          :disabled="isLoading"
          @click="fetchDoctorReport"
          title="重新执行系统健康全景诊断"
        >
          <TechIcon id="workbench:sync" :size="14" :class="{ 'spin-icon': isLoading }" />
          <span>{{ isLoading ? "正在体检..." : "重新全面体检" }}</span>
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
                ? "正在执行 UniRTM 全景环境健康检查..."
                : report?.score === 100
                  ? t("doctor.healthy") || "项目环境完全健康"
                  : report && report.score >= 80
                    ? "环境基本就绪，存在优化建议"
                    : "检测到环境配置隐患，建议核查修复"
            }}
          </h3>
          <p class="score-sub">
            <template v-if="report">
              核心版本: {{ report.core.version }} • 运行时代理:
              {{ report.core.shims_on_path ? "已注入 PATH" : "未注入 PATH" }} • 数据库:
              {{ report.health.database_ok ? "正常" : "异常" }} • GitHub API:
              {{ report.health.github_connected ? `连接正常 (HTTP ${report.health.github_status_code})` : "未连接" }}
            </template>
            <template v-else> 正在解析本地运行时与环境状态... </template>
          </p>
        </div>
      </div>
      <div v-if="report && lastCheckedTime" class="score-right-meta">
        <span class="meta-label">最后体检时间</span>
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
          <button class="sug-action-btn secondary" @click="copyCommand(sug.command)" title="复制修复命令">
            <TechIcon id="workbench:copy" :size="13" />
            <span>复制命令</span>
          </button>
          <button class="sug-action-btn primary" @click="runFixCommand(sug.command)" title="发送命令到终端直接运行">
            <TechIcon id="workbench:terminal" :size="13" />
            <span>在终端运行</span>
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
            <span class="card-title">核心与执行环境 (Core & Shell)</span>
          </div>
          <span class="badge-status" :class="report.core.shims_on_path ? 'pass' : 'warn'">
            {{ report.core.shims_on_path ? "Shims 已就绪" : "Shims 需配置" }}
          </span>
        </div>
        <div class="card-kv-list">
          <div class="kv-row">
            <span class="kv-key">UniRTM 版本</span>
            <span class="kv-val font-mono">{{ report.core.version }} ({{ report.core.commit }})</span>
          </div>
          <div class="kv-row">
            <span class="kv-key">目标架构</span>
            <span class="kv-val font-mono">{{ report.build.target }} ({{ report.build.go_version }})</span>
          </div>
          <div class="kv-row">
            <span class="kv-key">当前 Shell</span>
            <span class="kv-val font-mono">{{ report.context.shell }} {{ report.context.shell_version }}</span>
          </div>
          <div class="kv-row">
            <span class="kv-key">默认编辑器</span>
            <span class="kv-val font-mono">{{ report.context.editor }} ({{ report.context.editor_source }})</span>
          </div>
          <div class="kv-row">
            <span class="kv-key">生效环境</span>
            <span class="kv-val font-mono highlight">{{ report.context.active_env }}</span>
          </div>
        </div>
      </div>

      <!-- 2. Cloud Network & GitHub Quota -->
      <div id="doctor-section-network" class="glass-card doctor-card">
        <div class="card-header">
          <div class="card-title-group">
            <span class="card-icon"><TechIcon id="workbench:runtimes" :size="16" /></span>
            <span class="card-title">网络联通与 API 配额 (Network)</span>
          </div>
          <span class="badge-status" :class="report.health.github_connected ? 'pass' : 'fail'">
            {{ report.health.github_connected ? "GitHub API 正常" : "网络断开" }}
          </span>
        </div>
        <div class="card-kv-list">
          <div class="kv-row">
            <span class="kv-key">GitHub 联通状态</span>
            <span class="kv-val font-mono" :style="{ color: report.health.github_connected ? '#10b981' : '#ef4444' }">
              ● {{ report.health.github_connected ? `HTTP ${report.health.github_status_code} (已连通)` : "未连接" }}
            </span>
          </div>
          <!-- Rate Limit Progress Bar -->
          <div v-if="report.health.github_rate_limit > 0" class="rate-limit-section">
            <div class="rate-limit-header font-mono">
              <span>GitHub API 剩余配额</span>
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
              配额将在约 {{ report.health.github_rate_reset_text }} 后重置
            </div>
          </div>
          <div v-else class="kv-row">
            <span class="kv-key">API 速率限制</span>
            <span class="kv-val font-mono">匿名/未返回配额</span>
          </div>
        </div>
      </div>

      <!-- 3. Storage & Footprint -->
      <div id="doctor-section-directories" class="glass-card doctor-card">
        <div class="card-header">
          <div class="card-title-group">
            <span class="card-icon"><TechIcon id="workbench:folder" :size="16" /></span>
            <span class="card-title">磁盘占用与存储路径 (Directories)</span>
          </div>
        </div>
        <div class="card-dirs-list">
          <div v-for="dir in report.directories" :key="dir.type" class="dir-item font-mono">
            <div class="dir-type-side">
              <span class="dir-badge">{{ dir.type }}</span>
              <span class="dir-size" :title="`占用大小: ${dir.size}`">{{ dir.size }}</span>
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
            <span class="card-title">状态数据库与配置层叠</span>
          </div>
          <span class="badge-status" :class="report.health.database_ok ? 'pass' : 'fail'">
            {{ report.health.database_ok ? "SQLite OK" : "DB 异常" }}
          </span>
        </div>
        <div class="card-kv-list">
          <div class="kv-row">
            <span class="kv-key">SQLite 数据库</span>
            <span class="kv-val font-mono" :title="report.health.database_path">
              {{ report.health.database_ok ? `已连接 (${report.health.database_size})` : report.health.database_error }}
            </span>
          </div>
          <div class="kv-config-section">
            <span class="kv-key">已生效配置文件</span>
            <div class="config-files-box font-mono">
              <template v-if="report.configs && report.configs.length > 0">
                <div v-for="c in report.configs" :key="c.path" class="config-file-row">
                  <span class="cfg-scope" :class="c.scope">{{ c.scope }}</span>
                  <span class="cfg-path" :title="c.path">{{ c.path }}</span>
                </div>
              </template>
              <div v-else style="color: var(--text-muted); font-size: 0.78rem">
                默认配置 (未检测到生效的 toml 配置文件)
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 5. Task & Workflow Discovery -->
      <div id="doctor-section-tasks" class="glass-card doctor-card doctor-card-full">
        <div class="card-header">
          <div class="card-title-group">
            <span class="card-icon"><TechIcon id="workbench:tasks" :size="16" /></span>
            <span class="card-title">自动化工程任务与就绪度 (Tasks & Tools)</span>
          </div>
          <span class="badge-status pass"> {{ report.tasks.length }} 个任务已发现 </span>
        </div>
        <div class="tasks-pill-container font-mono">
          <div v-for="task in report.tasks" :key="task.name" class="task-chip" :title="task.command">
            <span class="task-chip-icon">⚡</span>
            <span class="task-chip-name">{{ task.name }}</span>
            <span v-if="task.description" class="task-chip-desc">{{ task.description }}</span>
          </div>
          <div v-if="report.tasks.length === 0" style="color: var(--text-muted); font-size: 0.8rem; padding: 0.5rem">
            当前工程尚未定义自动化 [tasks] 任务流。
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

// 计算健康分横幅的视觉色阶
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

// 计算 GitHub API 剩余配额百分比
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
  emit("log-event", "INFO", "正在触发 UniRTM Doctor 系统诊断...", undefined, "doctor");
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
        `[Doctor] 体检完成: 评分 ${data.score}/100, 发现 ${data.suggestions?.length || 0} 项修复建议`,
        undefined,
        "doctor"
      );
    } else {
      showToast("诊断接口未就绪或未返回 JSON", "error");
      emit("log-event", "ERROR", `[Doctor] 诊断接口请求失败 (HTTP ${res.status})`, undefined, "doctor");
    }
  } catch (err: any) {
    showToast(`诊断异常: ${err.message}`, "error");
    emit("log-event", "ERROR", `[Doctor] 异常: ${err.message}`, undefined, "doctor");
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
        showToast(`已复制修复命令: ${cmd}`, "success");
      })
      .catch(() => {
        showToast("复制失败", "error");
      });
  }
}

function runFixCommand(cmd: string) {
  emit("open-drawer", "output");
  sendToTerminal({ text: cmd, addNewline: true });
  showToast(`已向集成终端派发命令: ${cmd}`, "success");
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
      showToast("已复制诊断报告 Markdown 到剪贴板", "success");
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
  padding: 0.45rem 0.9rem;
  border-radius: 6px;
  font-size: 0.82rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
  border: 1px solid transparent;
}

.doctor-action-btn.primary {
  background: linear-gradient(135deg, var(--accent-cyan), #0284c7);
  color: #070a12;
  box-shadow: 0 2px 8px var(--accent-cyan-glow);
}

.doctor-action-btn.primary:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px var(--accent-cyan-glow);
}

.doctor-action-btn.secondary {
  background: var(--card-bg);
  border-color: var(--card-border);
  color: var(--text-main);
}

.doctor-action-btn.secondary:hover:not(:disabled) {
  background: var(--input-bg);
  border-color: var(--card-border-active);
}

.doctor-action-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
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

.doctor-card-full {
  grid-column: 1 / -1;
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

/* Task Pills */
.tasks-pill-container {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.task-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.35rem 0.65rem;
  border-radius: 6px;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  font-size: 0.78rem;
  transition: all 0.15s ease;
}

.task-chip:hover {
  border-color: var(--accent-cyan);
  background: rgba(0, 229, 255, 0.05);
}

.task-chip-icon {
  color: #f59e0b;
}

.task-chip-name {
  color: var(--text-main);
  font-weight: 600;
}

.task-chip-desc {
  color: var(--text-muted);
  font-size: 0.72rem;
  max-width: 220px;
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
