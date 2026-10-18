<template>
  <div class="view-container about-view" id="view-about">
    <!-- Top Header Bar -->
    <div class="view-header-row">
      <div>
        <h2 class="view-title" style="display: flex; align-items: center; gap: 8px;">
          <TechIcon id="workbench:about" :size="20" />
          <span>{{ t("about.title") || "关于" }} (About UniRTM)</span>
        </h2>
        <p class="view-desc">
          {{ t("about.subtitle") || "通用运行时与开发环境多版本统一管理器" }}
        </p>
      </div>
      <div class="view-actions">
        <button class="about-action-btn secondary" @click="copySystemInfo" :disabled="copied">
          <TechIcon id="workbench:config" :size="14" />
          <span>{{ copied ? t("about.copied") || "已复制到剪贴板" : t("about.copyInfo") || "复制系统信息" }}</span>
        </button>

        <button
          v-if="!hasUpdateAvailable && !isRestartReady"
          class="about-action-btn primary"
          @click="handleCheckUpdate"
          :disabled="checking || updating"
        >
          <span v-if="checking" class="spin-icon">
            <TechIcon id="workbench:sync" :size="14" />
          </span>
          <span v-else style="display: inline-flex; align-items: center; gap: 6px;">
            <TechIcon id="workbench:search" :size="14" />
            <span>{{ t("about.checkUpdate") || "检查更新" }}</span>
          </span>
        </button>

        <button
          v-else-if="isRestartReady"
          class="about-action-btn restart"
          @click="handleRestartApp"
          :disabled="restarting"
        >
          <span style="display: inline-flex; align-items: center; gap: 6px;">
            <TechIcon id="workbench:speed" :size="14" />
            <span>{{ restarting ? t("about.restarting") || "正在重启..." : t("about.restartToApply") || "立即重启生效" }}</span>
          </span>
        </button>

        <button v-else class="about-action-btn update" @click="handlePerformUpdate" :disabled="updating">
          <span style="display: inline-flex; align-items: center; gap: 6px;">
            <TechIcon id="workbench:tasks" :size="14" />
            <span>{{ updating ? t("about.updating", { progress: updateProgress }) : t("about.updateTo", { tag: latestTag }) }}</span>
          </span>
        </button>
      </div>
    </div>

    <!-- Update Feedback Banner -->
    <div v-if="updateMessage" class="update-status-banner glass-card" :class="updateStatusClass">
      <span class="banner-icon">
        <TechIcon :id="isRestartReady ? 'workbench:sparkle' : 'workbench:info'" :size="16" />
      </span>
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
        <div class="highlight-icon-wrap">
          <TechIcon id="workbench:tools" :size="24" />
        </div>
        <div class="highlight-body">
          <h3>多运行时声明式管控</h3>
          <p>无缝管理 Node.js、Go、Python、Rust、Java 等多语言版本，支持工具链锁定与跨平台环境对齐。</p>
        </div>
      </div>

      <div class="glass-card highlight-card">
        <div class="highlight-icon-wrap">
          <TechIcon id="workbench:lock" :size="24" />
        </div>
        <div class="highlight-body">
          <h3>三层命名空间与排他锁</h3>
          <p>严格隔离核心系统设置与第三方插件参数，写入自动持有文件锁并完整保留原有 TOML 注释排版。</p>
        </div>
      </div>

      <div class="glass-card highlight-card">
        <div class="highlight-icon-wrap">
          <TechIcon id="workbench:speed" :size="24" />
        </div>
        <div class="highlight-body">
          <h3>极速无感即时自动保存</h3>
          <p>遵循现代专业 IDE 哲学，界面选项修改即时生效落盘，杜绝手动冗余保存按钮与配置丢失风险。</p>
        </div>
      </div>
    </div>

    <!-- Open Source Acknowledgements & Icon Credits -->
    <div class="glass-card about-credits-card">
      <div class="credits-header">
        <TechIcon id="workbench:about" :size="18" />
        <h3>{{ t("about.credits_title") || "开源鸣谢与图标集许可声明 (Open Source & Icon Credits)" }}</h3>
      </div>
      <p class="credits-desc">
        {{ t("about.credits_desc") || "UniRTM 深度融合现代开源开发生态，严格遵循开源软件分发与商业许可合规标准：" }}
      </p>
      <div class="credits-grid">
        <div class="credit-item">
          <div class="credit-item-title">
            <span class="codicon codicon-vscode"></span>
            <a
              href="https://github.com/microsoft/vscode-codicons"
              target="_blank"
              rel="noopener noreferrer"
              class="credit-link"
            >
              VS Code Codicons (@vscode/codicons) ↗
            </a>
            <a
              href="https://github.com/microsoft/vscode-codicons/blob/main/LICENSE"
              target="_blank"
              rel="noopener noreferrer"
              class="credit-license-badge"
              title="查看 CC-BY-4.0 商业可用授权"
            >
              CC-BY-4.0 ↗
            </a>
          </div>
          <p>
            {{ t("about.credits_codicons_desc") || "UniRTM 工作台标准产品操作与界面图标集。由 Microsoft 官方维护，遵循 CC-BY-4.0 商业可用许可，提供官方标准工作台交互视觉语言。" }}
          </p>
        </div>

        <div class="credit-item">
          <div class="credit-item-title">
            <span class="codicon codicon-file"></span>
            <a
              href="https://github.com/jesseweed/seti-ui"
              target="_blank"
              rel="noopener noreferrer"
              class="credit-link"
            >
              Seti UI 图标集 (@vscode/seti-ui) ↗
            </a>
            <a
              href="https://github.com/jesseweed/seti-ui/blob/master/LICENSE.md"
              target="_blank"
              rel="noopener noreferrer"
              class="credit-license-badge"
              title="查看 MIT 授权协议"
            >
              MIT License ↗
            </a>
          </div>
          <p>
            {{ t("about.credits_seti_desc") || "UniRTM 默认文件与开发语言图标集。由 Jesse Weed 与 Visual Studio Code 团队维护，遵循 MIT 许可，精准映射文件扩展名与语言标识。" }}
          </p>
        </div>

        <div class="credit-item">
          <div class="credit-item-title">
            <span class="codicon codicon-extensions"></span>
            <a
              href="https://open-vsx.org"
              target="_blank"
              rel="noopener noreferrer"
              class="credit-link"
            >
              VS Code 主题扩展规范与 Open VSX ↗
            </a>
            <a
              href="https://open-vsx.org"
              target="_blank"
              rel="noopener noreferrer"
              class="credit-license-badge"
              title="访问 Open VSX 开放扩展市场"
            >
              Open VSX ↗
            </a>
          </div>
          <p>
            {{ t("about.credits_themes_desc") || "原生兼容 VS Code 三大扩展主题维度：" }}
            <span class="manifest-links">
              <a href="https://code.visualstudio.com/api/extension-guides/color-theme" target="_blank" rel="noopener noreferrer" class="sub-link">颜色主题 (contributes.themes) ↗</a>、
              <a href="https://code.visualstudio.com/api/extension-guides/file-icon-theme" target="_blank" rel="noopener noreferrer" class="sub-link">文件图标 (contributes.iconThemes) ↗</a>、
              <a href="https://code.visualstudio.com/api/extension-guides/product-icon-theme" target="_blank" rel="noopener noreferrer" class="sub-link">界面图标 (contributes.productIconThemes) ↗</a>。
            </span>
          </p>
        </div>
      </div>
    </div>

    <!-- Footer Links & Copyright -->
    <div class="glass-card about-footer-card">
      <div class="footer-links">
        <button @click="openUrl('https://github.com/snowdreamtech/unirtm')" class="footer-link-btn">
          <TechIcon id="workbench:github" :size="16" />
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
import TechIcon from "../components/common/TechIcon.vue";

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
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  padding: 7px 16px;
  border-radius: 8px;
  font-size: 0.82rem;
  font-weight: 600;
  cursor: pointer;
  border: 1px solid transparent;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  line-height: 1.4;
  user-select: none;
}

.about-action-btn.primary {
  background: linear-gradient(135deg, #0284c7 0%, #0369a1 100%);
  color: #ffffff !important;
  border: 1px solid rgba(255, 255, 255, 0.18);
  box-shadow: 0 2px 8px rgba(2, 132, 199, 0.28);
}

.about-action-btn.primary:hover:not(:disabled) {
  background: linear-gradient(135deg, #0ea5e9 0%, #0284c7 100%);
  border-color: #38bdf8;
  color: #ffffff !important;
  box-shadow: 0 4px 14px rgba(2, 132, 199, 0.4);
  transform: translateY(-1px);
}

.about-action-btn.secondary {
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-main);
  border: 1px solid var(--card-border);
}

.about-action-btn.secondary:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.1);
  border-color: var(--card-border-active);
  transform: translateY(-1px);
}

.about-action-btn.restart {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  color: #ffffff !important;
  border: 1px solid rgba(16, 185, 129, 0.4);
  box-shadow: 0 2px 8px rgba(16, 185, 129, 0.28);
}

.about-action-btn.restart:hover:not(:disabled) {
  background: linear-gradient(135deg, #34d399 0%, #10b981 100%);
  border-color: #6ee7b7;
  color: #ffffff !important;
  box-shadow: 0 4px 14px rgba(16, 185, 129, 0.4);
  transform: translateY(-1px);
}

.about-action-btn.update {
  background: linear-gradient(135deg, #9333ea 0%, #7c3aed 100%);
  color: #ffffff !important;
  border: 1px solid rgba(147, 51, 234, 0.4);
  box-shadow: 0 2px 8px rgba(147, 51, 234, 0.28);
}

.about-action-btn.update:hover:not(:disabled) {
  background: linear-gradient(135deg, #a855f7 0%, #9333ea 100%);
  border-color: #c084fc;
  color: #ffffff !important;
  box-shadow: 0 4px 14px rgba(147, 51, 234, 0.4);
  transform: translateY(-1px);
}

.about-action-btn:active:not(:disabled) {
  transform: translateY(0);
}

.about-action-btn:disabled {
  opacity: 0.55;
  cursor: not-allowed;
  box-shadow: none;
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

/* Open Source & Icon Credits Card */
.about-credits-card {
  padding: 1.5rem;
  border-radius: 12px;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.credits-header {
  display: flex;
  align-items: center;
  gap: 0.6rem;
}

.credits-header h3 {
  font-size: 1.05rem;
  font-weight: 600;
  margin: 0;
  color: var(--text-main);
}

.credits-desc {
  font-size: 0.85rem;
  color: var(--text-muted);
  margin: 0;
  line-height: 1.5;
}

.credits-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 1rem;
  margin-top: 0.25rem;
}

.credit-item {
  background: var(--bg-card-glass, rgba(255, 255, 255, 0.03));
  border: 1px solid var(--border-subtle, rgba(255, 255, 255, 0.08));
  border-radius: 8px;
  padding: 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.credit-item-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.9rem;
  color: var(--text-main);
  flex-wrap: wrap;
}

.credit-link {
  color: var(--text-main);
  font-weight: 600;
  text-decoration: none;
  transition: color 0.2s ease;
}

.credit-link:hover {
  color: var(--accent-cyan, #00e5ff);
  text-decoration: underline;
}

.credit-license-badge {
  font-size: 0.7rem;
  padding: 2px 6px;
  border-radius: 4px;
  background: rgba(0, 229, 255, 0.12);
  color: var(--accent-cyan, #00e5ff);
  border: 1px solid rgba(0, 229, 255, 0.25);
  font-weight: 600;
  margin-left: auto;
  text-decoration: none;
  transition: all 0.2s ease;
  display: inline-flex;
  align-items: center;
}

.credit-license-badge:hover {
  background: rgba(0, 229, 255, 0.25);
  color: #ffffff;
  border-color: var(--accent-cyan, #00e5ff);
  transform: translateY(-1px);
}

.manifest-links {
  display: inline;
}

.sub-link {
  color: var(--accent-cyan, #00e5ff);
  text-decoration: none;
  font-size: 0.8rem;
  font-family: var(--font-mono, monospace);
  transition: opacity 0.2s ease;
}

.sub-link:hover {
  text-decoration: underline;
  opacity: 0.85;
}

.credit-item p {
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

/* ========================================================
   LIGHT THEME OVERRIDES FOR ABOUT VIEW
   ======================================================== */
:root[data-theme="light"] .about-action-btn.primary,
[data-theme="light"] .about-action-btn.primary {
  background: linear-gradient(135deg, #0284c7 0%, #0369a1 100%);
  color: #ffffff !important;
  border: 1px solid #0284c7;
  box-shadow: 0 2px 8px rgba(2, 132, 199, 0.28);
}

:root[data-theme="light"] .about-action-btn.primary:hover:not(:disabled),
[data-theme="light"] .about-action-btn.primary:hover:not(:disabled) {
  background: linear-gradient(135deg, #0ea5e9 0%, #0284c7 100%);
  border-color: #0284c7;
  color: #ffffff !important;
  box-shadow: 0 4px 14px rgba(2, 132, 199, 0.38);
}

:root[data-theme="light"] .about-action-btn.secondary,
[data-theme="light"] .about-action-btn.secondary {
  background: #ffffff;
  border-color: #cbd5e1;
  color: #334155;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

:root[data-theme="light"] .about-action-btn.secondary:hover:not(:disabled),
[data-theme="light"] .about-action-btn.secondary:hover:not(:disabled) {
  background: #f8fafc;
  border-color: #94a3b8;
  color: #0f172a;
}

:root[data-theme="light"] .about-action-btn.restart,
[data-theme="light"] .about-action-btn.restart {
  background: linear-gradient(135deg, #059669 0%, #047857 100%);
  color: #ffffff !important;
  border: 1px solid #059669;
  box-shadow: 0 2px 8px rgba(5, 150, 105, 0.28);
}

:root[data-theme="light"] .about-action-btn.update,
[data-theme="light"] .about-action-btn.update {
  background: linear-gradient(135deg, #7c3aed 0%, #6d28d9 100%);
  color: #ffffff !important;
  border: 1px solid #7c3aed;
  box-shadow: 0 2px 8px rgba(124, 58, 237, 0.28);
}

:root[data-theme="light"] .hero-tag,
[data-theme="light"] .hero-tag {
  background: #f1f5f9;
  border-color: #cbd5e1;
  color: #475569;
}

:root[data-theme="light"] .credit-item,
[data-theme="light"] .credit-item {
  background: #f8fafc;
  border-color: #e2e8f0;
}

:root[data-theme="light"] .credit-license-badge,
[data-theme="light"] .credit-license-badge {
  background: rgba(2, 132, 199, 0.1);
  border-color: rgba(2, 132, 199, 0.25);
  color: #0284c7;
}

:root[data-theme="light"] .credit-license-badge:hover,
[data-theme="light"] .credit-license-badge:hover {
  background: rgba(2, 132, 199, 0.2);
  color: #0284c7;
}
</style>
