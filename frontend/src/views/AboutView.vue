<template>
  <div class="view-container about-view" id="view-about">
    <!-- Top Header Bar -->
    <div class="view-header-row">
      <div>
        <h2 class="view-title">ℹ️ {{ t("about.title") || "关于" }} (About UniRTM)</h2>
        <p class="view-desc">
          {{ t("about.subtitle") || "通用运行时与开发环境多版本统一管理器" }}
        </p>
      </div>
      <div class="view-actions">
        <button class="about-action-btn secondary" @click="copySystemInfo" :disabled="copied">
          <svg viewBox="0 0 24 24" width="15" height="15" stroke="currentColor" stroke-width="2" fill="none">
            <path d="M16 4h2a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h2"></path>
            <rect x="8" y="2" width="8" height="4" rx="1" ry="1"></rect>
          </svg>
          <span>{{ copied ? t("about.copied") || "已复制到剪贴板" : t("about.copyInfo") || "复制系统信息" }}</span>
        </button>

        <button
          v-if="!hasUpdateAvailable && !isRestartReady"
          class="about-action-btn primary"
          @click="handleCheckUpdate"
          :disabled="checking || updating"
        >
          <span v-if="checking" class="spin-icon">🔄</span>
          <span v-else>🔍 {{ t("about.checkUpdate") || "检查更新" }}</span>
        </button>

        <button
          v-else-if="isRestartReady"
          class="about-action-btn restart"
          @click="handleRestartApp"
          :disabled="restarting"
        >
          <span>🚀 {{ restarting ? t("about.restarting") || "正在重启..." : t("about.restartToApply") || "立即重启生效" }}</span>
        </button>

        <button v-else class="about-action-btn update" @click="handlePerformUpdate" :disabled="updating">
          <span>⬇️ {{ updating ? t("about.updating", { progress: updateProgress }) : t("about.updateTo", { tag: latestTag }) }}</span>
        </button>
      </div>
    </div>

    <!-- Update Feedback Banner -->
    <div v-if="updateMessage" class="update-status-banner glass-card" :class="updateStatusClass">
      <span class="banner-icon">{{ isRestartReady ? '🎉' : 'ℹ️' }}</span>
      <span class="banner-text">{{ updateMessage }}</span>
    </div>

    <!-- Progress Bar during download -->
    <div v-if="updating" class="update-progress-card glass-card">
      <div class="progress-bar-track">
        <div class="progress-bar-fill" :style="{ width: updateProgress + '%' }"></div>
      </div>
      <span class="progress-text">{{ updateStatusText || t("about.downloading", { progress: updateProgress }) }}</span>
    </div>

    <!-- Hero Brand Banner -->
    <div class="glass-card about-hero-banner">
      <div class="hero-left">
        <div class="logo-wrapper">
          <div class="logo-glow"></div>
          <img src="/logo.png" alt="UniRTM" class="app-logo-image" />
        </div>
        <div class="hero-details">
          <div class="title-with-badge">
            <h1 class="app-brand-name">{{ t("app.title") || "UniRTM" }}</h1>
            <span class="version-pill font-mono">
              <span class="dot"></span>
              {{ displayVersion }}
            </span>
          </div>
          <p class="hero-tagline">Universal Runtime Manager — Next-Gen Modern Dev Environment Workspace</p>
          <div class="hero-tags">
            <span class="hero-tag">⚡ Fast AST Engine</span>
            <span class="hero-tag">🔒 Multi-Scope Isolation</span>
            <span class="hero-tag">🌐 Embedded Workbench</span>
            <span class="hero-tag">📦 Zero-Dependency</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Core Specs Grid -->
    <div class="specs-grid">
      <div class="glass-card spec-card">
        <span class="spec-label">{{ t("about.gitTag") || "发行版本 (Git Tag)" }}</span>
        <span class="spec-value font-mono">{{ displayGitTag }}</span>
      </div>

      <div class="glass-card spec-card">
        <span class="spec-label">{{ t("about.commitHash") || "提交哈希 (Commit)" }}</span>
        <span class="spec-value font-mono">{{ displayCommitHash }}</span>
      </div>

      <div class="glass-card spec-card">
        <span class="spec-label">{{ t("about.buildTime") || "编译时间 (Build Time)" }}</span>
        <span class="spec-value font-mono">{{ displayBuildTime }}</span>
      </div>

      <div class="glass-card spec-card">
        <span class="spec-label">{{ t("about.environment") || "运行环境 (Platform)" }}</span>
        <span class="spec-value font-mono">{{ appInfo.osArch || "Linux / Darwin / Windows" }} ({{ appInfo.goVersion || "Go 1.27+" }})</span>
      </div>
    </div>

    <!-- Architecture Highlights Grid -->
    <div class="highlights-grid">
      <div class="glass-card highlight-card">
        <div class="highlight-icon-wrap">🛠️</div>
        <div class="highlight-body">
          <h3>多运行时声明式管控</h3>
          <p>无缝管理 Node.js、Go、Python、Rust、Java 等多语言版本，支持工具链锁定与跨平台环境对齐。</p>
        </div>
      </div>

      <div class="glass-card highlight-card">
        <div class="highlight-icon-wrap">🔒</div>
        <div class="highlight-body">
          <h3>三层命名空间与排他锁</h3>
          <p>严格隔离核心系统设置与第三方插件参数，写入自动持有文件锁并完整保留原有 TOML 注释排版。</p>
        </div>
      </div>

      <div class="glass-card highlight-card">
        <div class="highlight-icon-wrap">⚡</div>
        <div class="highlight-body">
          <h3>极速无感即时自动保存</h3>
          <p>遵循现代专业 IDE 哲学，界面选项修改即时生效落盘，杜绝手动冗余保存按钮与配置丢失风险。</p>
        </div>
      </div>
    </div>

    <!-- Footer Links & Copyright -->
    <div class="glass-card about-footer-card">
      <div class="footer-links">
        <button @click="openUrl('https://github.com/snowdreamtech/unirtm')" class="footer-link-btn">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
            <path
              d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0024 12c0-6.63-5.37-12-12-12z"
            />
          </svg>
          <span>GitHub Repository</span>
        </button>
        <span class="link-separator">•</span>
        <button @click="openUrl('https://github.com/snowdreamtech/unirtm/blob/main/LICENSE')" class="footer-link-btn">
          <span>MIT License</span>
        </button>
      </div>
      <p class="copyright-text">
        {{ appInfo.copyright || "Copyright © 2026-present SnowdreamTech Inc. All rights reserved." }}
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { t } from "../i18n";
import { formatUpdateProgress, type UpdateProgressPayload } from "../utils/updater";

interface AppInfo {
  projectName?: string;
  version?: string;
  gitTag?: string;
  commitHash?: string;
  commitHashFull?: string;
  buildTime?: string;
  author?: string;
  copyright?: string;
  license?: string;
  goVersion?: string;
  osArch?: string;
}

const appInfo = ref<AppInfo>({
  projectName: "UniRTM",
  version: "v0.1.0",
  gitTag: "v0.1.0",
  commitHash: "HEAD",
  buildTime: new Date().toISOString().slice(0, 10),
  copyright: "Copyright © 2026 SnowdreamTech. All rights reserved.",
  goVersion: "Go 1.27+",
  osArch: "Web / Universal",
});

const displayVersion = computed(() => appInfo.value.version || appInfo.value.gitTag || "v0.1.0");
const displayGitTag = computed(() => appInfo.value.gitTag || "v0.1.0");
const displayCommitHash = computed(() => appInfo.value.commitHash || "dev");
const displayBuildTime = computed(() => appInfo.value.buildTime || "N/A");

const copied = ref(false);
const checking = ref(false);
const updating = ref(false);
const isRestartReady = ref(false);
const restarting = ref(false);
const updateProgress = ref(0);
const currentProgress = ref<UpdateProgressPayload | null>(null);

const updateStatusText = computed(() => {
  return formatUpdateProgress(currentProgress.value, updateProgress.value);
});

const hasUpdateAvailable = ref(false);
const latestTag = ref("");
const updateMessage = ref("");
const updateStatusClass = ref("");

const openUrl = (url: string) => {
  window.open(url, "_blank", "noopener,noreferrer");
};

const copySystemInfo = async () => {
  const diagnosticText = `--- UniRTM Diagnostic Info ---
Version: ${displayVersion.value} (${displayGitTag.value})
Commit: ${displayCommitHash.value}
Build Time: ${displayBuildTime.value}
OS/Arch: ${appInfo.value.osArch || "Web / Universal"}
Go Runtime: ${appInfo.value.goVersion || "Go 1.27+"}
License: MIT License
------------------------------`;

  try {
    if (navigator.clipboard && navigator.clipboard.writeText) {
      await navigator.clipboard.writeText(diagnosticText);
    } else {
      const textArea = document.createElement("textarea");
      textArea.value = diagnosticText;
      document.body.appendChild(textArea);
      textArea.select();
      document.execCommand("copy");
      document.body.removeChild(textArea);
    }
    copied.value = true;
    setTimeout(() => {
      copied.value = false;
    }, 2500);
  } catch (err) {
    console.error("Failed to copy system info:", err);
  }
};

const handleCheckUpdate = async () => {
  checking.value = true;
  updateMessage.value = "";
  setTimeout(() => {
    checking.value = false;
    hasUpdateAvailable.value = false;
    updateMessage.value = t("about.isLatest") || "当前已是最新版本 (v0.1.0)";
    updateStatusClass.value = "is-latest";
  }, 600);
};

const handlePerformUpdate = async () => {
  updating.value = true;
  updateProgress.value = 0;
  currentProgress.value = { stage: "checking", percentage: 0 };
  updateMessage.value = "";

  let p = 0;
  const interval = setInterval(() => {
    p += 25;
    updateProgress.value = Math.min(p, 100);
    currentProgress.value = {
      percentage: updateProgress.value,
      stage: "downloading",
      loadedBytes: Math.round(10240000 * (updateProgress.value / 100)),
      totalBytes: 10240000,
    };
    if (p >= 100) {
      clearInterval(interval);
      updating.value = false;
      isRestartReady.value = true;
      updateMessage.value = t("about.updateCompleteRestart") || "升级包已就绪，点击重启即可应用。";
      updateStatusClass.value = "is-latest";
    }
  }, 250);
};

const handleRestartApp = async () => {
  restarting.value = true;
  updateMessage.value = t("about.restarting") || "正在刷新重载...";
  setTimeout(() => {
    window.location.reload();
  }, 600);
};

onMounted(() => {
  appInfo.value = {
    projectName: "UniRTM",
    version: "v0.1.0",
    gitTag: "v0.1.0",
    commitHash: "HEAD",
    buildTime: new Date().toISOString().slice(0, 10),
    osArch: "Web Client",
    goVersion: "Go 1.27+",
    copyright: "Copyright © 2026 SnowdreamTech. All rights reserved.",
  };
});
</script>

<style scoped>
.about-view {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  padding: 1.5rem;
  max-width: 1200px;
  margin: 0 auto;
  width: 100%;
}

.view-header-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.view-title {
  font-size: 1.4rem;
  font-weight: 700;
  color: var(--text-main);
  margin: 0 0 0.25rem;
}

.view-desc {
  font-size: 0.85rem;
  color: var(--text-muted);
  margin: 0;
}

.view-actions {
  display: flex;
  align-items: center;
  gap: 0.65rem;
}

.about-action-btn {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  padding: 7px 14px;
  border-radius: 6px;
  font-size: 0.82rem;
  font-weight: 600;
  cursor: pointer;
  border: 1px solid transparent;
  transition: all 0.2s ease;
}

.about-action-btn.primary {
  background: var(--accent-cyan);
  color: #070a12;
}

.about-action-btn.secondary {
  background: var(--input-bg);
  color: var(--text-main);
  border-color: var(--card-border);
}

.about-action-btn.restart {
  background: var(--success);
  color: #070a12;
}

.about-action-btn.update {
  background: var(--accent-purple);
  color: #ffffff;
}

.about-action-btn:hover:not(:disabled) {
  opacity: 0.9;
  transform: translateY(-1px);
}

.about-action-btn:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

/* Update Banner */
.update-status-banner {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  padding: 10px 14px;
  border-radius: 8px;
  background: rgba(16, 185, 129, 0.1);
  border: 1px solid rgba(16, 185, 129, 0.25);
  color: var(--success);
  font-size: 0.85rem;
}

.update-progress-card {
  padding: 12px 16px;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.progress-bar-track {
  width: 100%;
  height: 6px;
  background: var(--section-bg);
  border: 1px solid var(--card-border);
  border-radius: 3px;
  overflow: hidden;
}

.progress-bar-fill {
  height: 100%;
  background: linear-gradient(90deg, var(--accent-cyan), var(--accent-purple));
  transition: width 0.2s ease;
}

.progress-text {
  font-size: 0.75rem;
  color: var(--text-muted);
}

/* Hero Banner */
.about-hero-banner {
  padding: 1.5rem 1.75rem;
  border-radius: 14px;
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.04);
}

:root[data-theme="dark"] .about-hero-banner,
[data-theme="dark"] .about-hero-banner {
  background: linear-gradient(135deg, rgba(18, 26, 43, 0.95) 0%, rgba(13, 20, 36, 0.98) 100%);
  border-color: var(--card-border);
}

:root[data-theme="light"] .about-hero-banner,
[data-theme="light"] .about-hero-banner {
  background: linear-gradient(135deg, #ffffff 0%, #f8fafc 100%);
  border-color: var(--card-border);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.04);
}

.hero-left {
  display: flex;
  align-items: center;
  gap: 1.5rem;
}

.logo-wrapper {
  position: relative;
  width: 76px;
  height: 76px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.logo-glow {
  position: absolute;
  width: 90px;
  height: 90px;
  border-radius: 50%;
  background: radial-gradient(circle, var(--accent-cyan-glow) 0%, transparent 70%);
}

.app-logo-image {
  width: 64px;
  height: 64px;
  object-fit: contain;
  position: relative;
  z-index: 1;
}

.hero-details {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.title-with-badge {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.app-brand-name {
  font-size: 1.6rem;
  font-weight: 800;
  color: var(--text-main);
  margin: 0;
  letter-spacing: -0.5px;
}

.version-pill {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 2px 10px;
  border-radius: 12px;
  background: var(--accent-cyan-glow);
  border: 1px solid var(--card-border);
  color: var(--accent-cyan);
  font-size: 0.78rem;
  font-weight: 600;
}

.version-pill .dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--accent-cyan);
}

.hero-tagline {
  font-size: 0.85rem;
  color: var(--text-muted);
  margin: 0;
}

.hero-tags {
  display: flex;
  gap: 0.5rem;
  margin-top: 0.35rem;
  flex-wrap: wrap;
}

.hero-tag {
  font-size: 0.72rem;
  padding: 3px 9px;
  border-radius: 4px;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  color: var(--text-muted);
}

/* Specs Grid */
.specs-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 1rem;
}

.spec-card {
  padding: 1rem 1.15rem;
  border-radius: 10px;
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.spec-label {
  font-size: 0.75rem;
  color: var(--text-muted);
}

.spec-value {
  font-size: 0.88rem;
  font-weight: 600;
  color: var(--text-main);
}

/* Highlights Grid */
.highlights-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 1rem;
}

.highlight-card {
  padding: 1.25rem;
  border-radius: 10px;
  display: flex;
  gap: 1rem;
  align-items: flex-start;
}

.highlight-icon-wrap {
  font-size: 1.6rem;
  line-height: 1;
}

.highlight-body h3 {
  font-size: 0.95rem;
  font-weight: 700;
  color: var(--text-main);
  margin: 0 0 0.35rem;
}

.highlight-body p {
  font-size: 0.8rem;
  color: var(--text-muted);
  line-height: 1.45;
  margin: 0;
}

/* Footer Card */
.about-footer-card {
  padding: 1.25rem;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 1rem;
}

.footer-links {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.footer-link-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  background: transparent;
  border: none;
  color: var(--accent-cyan);
  font-size: 0.82rem;
  cursor: pointer;
  padding: 0;
  transition: opacity 0.2s ease;
}

.footer-link-btn:hover {
  text-decoration: underline;
  opacity: 0.85;
}

.link-separator {
  color: var(--text-muted);
}

.copyright-text {
  font-size: 0.78rem;
  color: var(--text-muted);
  margin: 0;
}

.spin-icon {
  display: inline-block;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
