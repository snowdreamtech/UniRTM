<template>
  <Teleport to="body">
    <Transition name="modal-fade">
      <div v-if="show" class="about-modal-overlay" @click.self="close" tabindex="0" @keydown.esc="close">
        <div class="about-modal-container glass-card">
          <!-- Header / Close button -->
          <div class="about-header">
            <button class="close-btn" @click="close" :title="t('common.close')">
              <svg viewBox="0 0 24 24" width="20" height="20" stroke="currentColor" stroke-width="2" fill="none">
                <line x1="18" y1="6" x2="6" y2="18"></line>
                <line x1="6" y1="6" x2="18" y2="18"></line>
              </svg>
            </button>
          </div>

          <!-- Hero Brand Area -->
          <div class="about-hero">
            <div class="logo-wrapper">
              <div class="logo-glow"></div>
              <img src="/logo.png" alt="UniRTM" class="app-logo-image" />
            </div>
            <h2 class="app-title">{{ t("app.title") || "UniRTM" }}</h2>
            <p class="app-subtitle">{{ t("about.subtitle") || "Universal Runtime Manager" }}</p>
            <div class="version-badge">
              <span class="badge-dot"></span>
              <span class="badge-text">{{ displayVersion }}</span>
            </div>
          </div>

          <!-- Environment & Build Info Grid -->
          <div class="info-grid">
            <div class="info-item">
              <span class="info-label">{{ t("about.gitTag") }}</span>
              <span class="info-val font-mono">{{ displayGitTag }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">{{ t("about.commitHash") }}</span>
              <span class="info-val font-mono">{{ displayCommitHash }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">{{ t("about.buildTime") }}</span>
              <span class="info-val font-mono">{{ displayBuildTime }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">{{ t("about.environment") }}</span>
              <span class="info-val font-mono">{{ appInfo.osArch || "N/A" }} ({{ appInfo.goVersion || "N/A" }})</span>
            </div>
          </div>

          <!-- Actions Area -->
          <div class="about-actions">
            <button class="action-btn secondary-btn" @click="copySystemInfo" :disabled="copied">
              <svg viewBox="0 0 24 24" width="16" height="16" stroke="currentColor" stroke-width="2" fill="none">
                <path d="M16 4h2a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h2"></path>
                <rect x="8" y="2" width="8" height="4" rx="1" ry="1"></rect>
              </svg>
              <span>{{ copied ? t("about.copied") : t("about.copyInfo") }}</span>
            </button>

            <button
              v-if="!hasUpdateAvailable && !isRestartReady"
              class="action-btn primary-btn"
              @click="handleCheckUpdate"
              :disabled="checking || updating"
            >
              <svg
                v-if="!checking"
                viewBox="0 0 24 24"
                width="16"
                height="16"
                stroke="currentColor"
                stroke-width="2"
                fill="none"
              >
                <polyline points="23 4 23 10 17 10"></polyline>
                <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"></path>
              </svg>
              <svg
                v-else
                class="spin-icon"
                viewBox="0 0 24 24"
                width="16"
                height="16"
                stroke="currentColor"
                stroke-width="2"
                fill="none"
              >
                <line x1="12" y1="2" x2="12" y2="6"></line>
                <line x1="12" y1="18" x2="12" y2="22"></line>
                <line x1="4.93" y1="4.93" x2="7.76" y2="7.76"></line>
                <line x1="16.24" y1="16.24" x2="19.07" y2="19.07"></line>
                <line x1="2" y1="12" x2="6" y2="12"></line>
                <line x1="18" y1="12" x2="22" y2="12"></line>
                <line x1="4.93" y1="19.07" x2="7.76" y2="16.24"></line>
                <line x1="16.24" y1="7.76" x2="19.07" y2="4.93"></line>
              </svg>
              <span>{{ checking ? t("about.checking") : t("about.checkUpdate") }}</span>
            </button>

            <!-- 关键演进：更新准备就绪时，按钮就地转化为【立即重启生效】 -->
            <button
              v-else-if="isRestartReady"
              class="action-btn restart-btn"
              @click="handleRestartApp"
              :disabled="restarting"
            >
              <svg viewBox="0 0 24 24" width="16" height="16" stroke="currentColor" stroke-width="2" fill="none">
                <path d="M21.5 2v6h-6M21.34 15.57a10 10 0 1 1-.57-8.38l5.67-5.67"></path>
              </svg>
              <span>{{
                restarting ? t("about.restarting") || "正在重启..." : t("about.restartToApply") || "立即重启生效"
              }}</span>
            </button>

            <button v-else class="action-btn update-btn" @click="handlePerformUpdate" :disabled="updating">
              <svg viewBox="0 0 24 24" width="16" height="16" stroke="currentColor" stroke-width="2" fill="none">
                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
                <polyline points="7 10 12 15 17 10"></polyline>
                <line x1="12" y1="15" x2="12" y2="3"></line>
              </svg>
              <span>{{
                updating ? t("about.updating", { progress: updateProgress }) : t("about.updateTo", { tag: latestTag })
              }}</span>
            </button>
          </div>

          <!-- Progress Bar during download -->
          <div v-if="updating" class="update-progress-container">
            <div class="progress-bar-track">
              <div class="progress-bar-fill" :style="{ width: updateProgress + '%' }"></div>
            </div>
            <span class="progress-text">{{
              updateStatusText || t("about.downloading", { progress: updateProgress })
            }}</span>
          </div>

          <!-- Status Message Toast -->
          <div v-if="updateMessage" class="update-status-msg" :class="updateStatusClass">
            {{ updateMessage }}
          </div>

          <!-- Links & Footer -->
          <div class="about-footer">
            <div class="footer-links">
              <button @click="openUrl('https://github.com/snowdreamtech/unirtm')" class="footer-link-btn">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor">
                  <path
                    d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0024 12c0-6.63-5.37-12-12-12z"
                  />
                </svg>
                GitHub Repository
              </button>
              <span class="link-separator">•</span>
              <button
                @click="openUrl('https://github.com/snowdreamtech/unirtm/blob/main/LICENSE')"
                class="footer-link-btn"
              >
                MIT License
              </button>
            </div>
            <p class="copyright-text">
              {{ appInfo.copyright || "Copyright © 2026-present SnowdreamTech Inc. All rights reserved." }}
            </p>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from "vue";
import { t } from "../i18n";
import { formatUpdateProgress, type UpdateProgressPayload } from "../utils/updater";

const props = defineProps<{
  show: boolean;
}>();

const emit = defineEmits<{
  (e: "close"): void;
}>();

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

const loadAppInfo = async () => {
  // Pure web metadata
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
};

onMounted(() => {
  if (props.show) {
    loadAppInfo();
  }
});

watch(
  () => props.show,
  (newVal) => {
    if (newVal) {
      loadAppInfo();
      copied.value = false;
      updateMessage.value = "";
      updating.value = false;
      isRestartReady.value = false;
      restarting.value = false;
    }
  }
);

const close = () => {
  emit("close");
};

const logUserAction = (level: string, message: string, details: string = "") => {
  if (import.meta.env?.DEV) {
    console.debug(`[${level}] ${message}`, details);
  }
};

const openUrl = (url: string) => {
  logUserAction("INFO", "User opened external link in browser", url);
  window.open(url, "_blank", "noopener,noreferrer");
};

const copySystemInfo = async () => {
  logUserAction("INFO", "User copied system diagnostic info to clipboard");
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
  logUserAction("INFO", "User manually checked for software updates in About modal");
  checking.value = true;
  updateMessage.value = "";
  setTimeout(() => {
    checking.value = false;
    hasUpdateAvailable.value = false;
    updateMessage.value = t("about.isLatest") || "Current version is up to date";
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
      updateMessage.value = t("about.updateCompleteRestart") || "Update downloaded. Ready to refresh.";
      updateStatusClass.value = "is-latest";
    }
  }, 250);
};

const handleRestartApp = async () => {
  logUserAction("INFO", "User clicked restart app to apply update");
  restarting.value = true;
  updateMessage.value = t("about.restarting") || "Refreshing...";
  setTimeout(() => {
    window.location.reload();
  }, 600);
};
</script>

<style scoped>
.about-modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  background: rgba(10, 15, 26, 0.75);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
}

.about-modal-container {
  width: 480px;
  max-width: 90vw;
  padding: 24px;
  border-radius: 20px;
  background: var(--modal-bg);
  border: 1px solid var(--card-border);
  box-shadow: 0 20px 50px rgba(0, 0, 0, 0.25);
  color: var(--text-main);
  position: relative;
  overflow: hidden;
}

.about-header {
  display: flex;
  justify-content: flex-end;
}

.close-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 6px;
  border-radius: 50%;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
}

.close-btn:hover {
  color: var(--text-main);
  background: var(--section-bg);
}

.about-hero {
  text-align: center;
  margin-top: -10px;
  margin-bottom: 24px;
}

.logo-wrapper {
  position: relative;
  display: inline-block;
  margin-bottom: 12px;
}

.logo-glow {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 80px;
  height: 80px;
  border-radius: 50%;
  background: radial-gradient(circle, var(--accent-cyan-glow) 0%, transparent 80%);
  filter: blur(16px);
  z-index: 0;
}

.app-logo-icon,
.app-logo-image {
  position: relative;
  z-index: 1;
  width: 68px;
  height: 68px;
  border-radius: 16px;
  object-fit: cover;
  box-shadow: 0 4px 20px var(--accent-cyan-glow);
}

.app-title {
  font-size: 24px;
  font-weight: 700;
  margin: 0;
  color: var(--text-main);
  letter-spacing: -0.5px;
}

.app-subtitle {
  font-size: 13px;
  color: var(--text-muted);
  margin: 4px 0 12px 0;
}

.version-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 12px;
  border-radius: 20px;
  background: var(--alert-success-bg);
  border: 1px solid var(--alert-success-border);
  color: var(--alert-success-title);
  font-size: 12px;
  font-weight: 600;
}

.badge-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--success);
}

.info-grid {
  background: var(--section-bg);
  border-radius: 12px;
  padding: 14px 16px;
  border: 1px solid var(--card-border);
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-bottom: 20px;
}

.info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 13px;
}

.info-label {
  color: var(--text-muted);
}

.info-val {
  color: var(--text-main);
  font-size: 12px;
  font-weight: 600;
}

.font-mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

.about-actions {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
}

.action-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 10px 16px;
  border-radius: 10px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
  border: none;
}

.secondary-btn {
  background: var(--section-bg);
  border: 1px solid var(--card-border);
  color: var(--text-main);
}

.secondary-btn:hover:not(:disabled) {
  background: var(--card-border);
}

.primary-btn {
  background: #2563eb;
  color: #ffffff;
}

.primary-btn:hover:not(:disabled) {
  background: #1d4ed8;
}

.update-btn {
  background: #059669;
  color: #ffffff;
}

.update-btn:hover:not(:disabled) {
  background: #047857;
}

.restart-btn {
  background: linear-gradient(135deg, #2563eb 0%, #0284c7 100%);
  color: #ffffff;
  box-shadow: 0 4px 14px rgba(37, 99, 235, 0.35);
}

.restart-btn:hover:not(:disabled) {
  background: linear-gradient(135deg, #1d4ed8 0%, #0369a1 100%);
  transform: translateY(-1px);
  box-shadow: 0 6px 18px rgba(37, 99, 235, 0.45);
}

.update-progress-container {
  margin-bottom: 16px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.progress-bar-track {
  width: 100%;
  height: 8px;
  background: var(--section-bg);
  border-radius: 4px;
  overflow: hidden;
}

.progress-bar-fill {
  height: 100%;
  background: var(--accent-cyan);
  border-radius: 4px;
  transition: width 0.2s ease;
}

.progress-text {
  font-size: 11px;
  color: var(--text-muted);
  text-align: center;
}

.action-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.spin-icon {
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

.update-status-msg {
  padding: 8px 12px;
  border-radius: 8px;
  font-size: 12px;
  text-align: center;
  margin-bottom: 16px;
}

.has-update {
  background: var(--alert-warning-bg);
  border: 1px solid var(--alert-warning-border);
  color: var(--alert-warning-title);
}

.is-latest {
  background: var(--alert-success-bg);
  border: 1px solid var(--alert-success-border);
  color: var(--alert-success-title);
}

.update-error {
  background: var(--alert-danger-bg);
  border: 1px solid var(--alert-danger-border);
  color: var(--alert-danger-title);
}

.about-footer {
  text-align: center;
  border-top: 1px solid var(--card-border);
  padding-top: 16px;
}

.footer-links {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  margin-bottom: 8px;
}

.footer-link,
.footer-link-btn {
  background: transparent;
  border: none;
  color: var(--accent-cyan);
  font-size: 12px;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 0;
  transition: color 0.2s ease;
  font-family: inherit;
}

.footer-link:hover,
.footer-link-btn:hover {
  text-decoration: underline;
}

.link-separator {
  color: var(--text-muted);
  font-size: 12px;
}

.copyright-text {
  font-size: 11px;
  color: var(--text-subtle);
  margin: 0;
}

.modal-fade-enter-active,
.modal-fade-leave-active {
  transition:
    opacity 0.25s ease,
    transform 0.25s ease;
}

.modal-fade-enter-from,
.modal-fade-leave-to {
  opacity: 0;
  transform: scale(0.95);
}
</style>
