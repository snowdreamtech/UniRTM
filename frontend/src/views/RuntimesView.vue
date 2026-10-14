<template>
  <div class="view-container runtimes-view" id="view-runtimes">
    <!-- View Header Row -->
    <div class="view-header-row">
      <div>
        <h2 class="view-title">🌐 {{ t("runtimes.title") || "Multi-Language Runtimes" }}</h2>
        <p class="view-desc">
          {{ t("runtimes.desc") || "Govern Node.js, Go, Python, and more — replacing NVM, GVM, and Pyenv with unified cross-platform versioning." }}
        </p>
      </div>
      <div class="view-actions">
        <button class="tool-action-btn secondary" @click="fetchStatus" :disabled="loading">
          <span :class="{ 'spin-icon': loading }">🔄</span>
          <span>{{ loading ? "Refreshing..." : "Refresh" }}</span>
        </button>
      </div>
    </div>

    <!-- Legacy Version Files Banner -->
    <div v-if="detectedLegacies.length > 0" class="legacy-banner glass-card" id="legacy-detected-banner">
      <div class="legacy-banner-left">
        <span class="legacy-banner-icon">💡</span>
        <div class="legacy-banner-content">
          <span class="legacy-banner-title">
            {{ t("runtimes.legacy_banner", { files: detectedLegacies.join(", ") }) }}
          </span>
          <span class="legacy-banner-sub">
            Seamlessly active via zero-configuration compatibility. You can keep them or import to .unirtm.toml for single-file governance.
          </span>
        </div>
      </div>
      <div class="legacy-banner-actions">
        <button
          class="tool-action-btn primary small"
          id="btn-migrate-legacy"
          @click="migrateLegacy"
          :disabled="migrating"
        >
          <span>✨ {{ migrating ? "Importing..." : (t("runtimes.import_legacy") || "Import to .unirtm.toml") }}</span>
        </button>
      </div>
    </div>

    <!-- Scope Filter & Stats Toolbar -->
    <div class="runtimes-filter-bar glass-card">
      <div class="search-input-wrapper">
        <span class="search-icon">🔍</span>
        <input
          v-model="searchKey"
          type="text"
          class="tool-search-input"
          :placeholder="t('runtimes.search_placeholder') || 'Search or filter runtimes...'"
        />
      </div>
      <div class="runtime-stats-badges">
        <span class="stat-badge">
          <b>{{ runtimes.length }}</b> Runtimes
        </span>
        <span class="stat-badge highlight">
          <b>{{ totalInstalled }}</b> Installed Releases
        </span>
      </div>
    </div>

    <!-- Runtimes Grid Cards -->
    <div class="runtimes-grid">
      <div
        v-for="rt in filteredRuntimes"
        :key="rt.id"
        class="glass-card runtime-card"
        :class="{ active: rt.active_version }"
        :id="'runtime-card-' + rt.id"
      >
        <!-- Card Header -->
        <div class="runtime-card-header">
          <div class="runtime-icon-badge">{{ getRuntimeIcon(rt.id) }}</div>
          <div class="runtime-title-block">
            <div class="runtime-name-row">
              <span class="runtime-name">{{ rt.name }}</span>
              <a
                v-if="rt.homepage"
                :href="rt.homepage"
                target="_blank"
                rel="noreferrer"
                class="runtime-homepage-link"
                title="Official Homepage"
              >
                ↗
              </a>
            </div>
            <span class="runtime-desc">{{ rt.description }}</span>
          </div>
        </div>

        <!-- Active Version & Origin Source -->
        <div class="runtime-status-box">
          <div class="status-row">
            <span class="status-label">{{ t("runtimes.active_version") || "Active Version" }}:</span>
            <div v-if="rt.active_version" class="status-val-wrap">
              <span class="active-pill font-mono">v{{ rt.active_version }}</span>
              <span class="source-tag" :class="getSourceTagClass(rt.source)">
                {{ rt.source || "default" }}
              </span>
            </div>
            <span v-else class="text-muted font-mono italic">None</span>
          </div>

          <div v-if="rt.global_version && rt.global_version !== rt.active_version" class="status-row sub-row">
            <span class="status-label">{{ t("runtimes.global_default") || "Global Default" }}:</span>
            <span class="font-mono global-val">v{{ rt.global_version }}</span>
          </div>
        </div>

        <!-- Installed Versions List -->
        <div class="installed-versions-section">
          <div class="section-title-row">
            <span class="section-title">{{ t("runtimes.installed_versions") || "Installed Versions" }}</span>
            <span class="count-pill">{{ rt.installed_versions ? rt.installed_versions.length : 0 }}</span>
          </div>

          <div v-if="rt.installed_versions && rt.installed_versions.length > 0" class="versions-pills-row">
            <button
              v-for="ver in rt.installed_versions"
              :key="ver"
              class="ver-pill font-mono"
              :class="{ selected: ver === rt.active_version }"
              :title="'Click to switch to ' + ver"
              @click="quickSwitch(rt.id, ver)"
            >
              <span v-if="ver === rt.active_version" class="active-dot">●</span>
              <span>v{{ ver }}</span>
            </button>
          </div>
          <div v-else class="empty-installed-text">
            {{ t("runtimes.no_installed") || "No versions installed locally" }}
          </div>
        </div>

        <!-- Card Actions & Switcher -->
        <div class="runtime-card-footer">
          <div class="switch-control-group">
            <select
              v-if="rt.installed_versions && rt.installed_versions.length > 0"
              class="runtime-select"
              :value="rt.active_version"
              @change="handleSelectChange(rt.id, $event)"
            >
              <option value="" disabled>Switch version...</option>
              <option
                v-for="ver in rt.installed_versions"
                :key="ver"
                :value="ver"
              >
                v{{ ver }}{{ ver === rt.active_version ? ' (current)' : '' }}
              </option>
            </select>
          </div>

          <button
            class="install-remote-btn"
            @click="openRemoteModal(rt)"
          >
            <span>+ {{ t("runtimes.install_new") || "Install Version" }}</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Remote Version Browser Modal -->
    <div v-if="isModalOpen" class="modal-backdrop" @click.self="isModalOpen = false">
      <div class="modal-dialog glass-card">
        <div class="modal-header">
          <div class="modal-title-row">
            <span class="modal-icon">{{ getRuntimeIcon(selectedRuntime?.id || '') }}</span>
            <div>
              <h3 class="modal-title">
                {{ selectedRuntime?.name }} — {{ t("runtimes.modal_title") || "Install Remote Version" }}
              </h3>
              <p class="modal-desc">{{ t("runtimes.modal_desc") }}</p>
            </div>
          </div>
          <button class="modal-close-btn" @click="isModalOpen = false">✕</button>
        </div>

        <div class="modal-search-row">
          <input
            v-model="remoteSearchKey"
            type="text"
            class="modal-search-input"
            placeholder="Filter remote versions (e.g. 20, lts, 1.23)..."
          />
        </div>

        <div class="modal-body-versions">
          <div v-if="loadingRemote" class="remote-loading-box">
            <span class="spin-icon">🔄</span>
            <span>Fetching remote release metadata...</span>
          </div>
          <div v-else-if="filteredRemoteVersions.length === 0" class="remote-empty-box">
            No matching versions found.
          </div>
          <div v-else class="remote-versions-list">
            <div
              v-for="v in filteredRemoteVersions.slice(0, 50)"
              :key="v.version"
              class="remote-ver-row"
            >
              <div class="remote-ver-info">
                <span class="font-mono ver-name">v{{ v.version }}</span>
                <span v-if="isAlreadyInstalled(selectedRuntime?.id || '', v.version)" class="installed-tag">
                  ✓ Installed
                </span>
              </div>
              <button
                class="install-action-btn"
                :disabled="isAlreadyInstalled(selectedRuntime?.id || '', v.version) || installingVer === v.version"
                @click="installRemoteVersion(selectedRuntime?.id || '', v.version)"
              >
                <span>{{ installingVer === v.version ? "Installing..." : "Install" }}</span>
              </button>
            </div>
          </div>
        </div>

        <div class="modal-footer">
          <button class="tool-action-btn secondary small" @click="isModalOpen = false">Close</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { t } from "../i18n";

export interface RuntimeItem {
  id: string;
  name: string;
  description: string;
  homepage: string;
  installed_versions: string[];
  active_version: string;
  global_version: string;
  source: string;
  is_legacy: boolean;
  legacy_file: string;
}

const runtimes = ref<RuntimeItem[]>([]);
const totalInstalled = ref(0);
const detectedLegacies = ref<string[]>([]);
const loading = ref(false);
const migrating = ref(false);
const searchKey = ref("");

// Remote modal state
const isModalOpen = ref(false);
const selectedRuntime = ref<RuntimeItem | null>(null);
const remoteVersions = ref<{ version: string }[]>([]);
const remoteSearchKey = ref("");
const loadingRemote = ref(false);
const installingVer = ref<string | null>(null);

const emit = defineEmits<{
  (e: "log-event", level: "INFO" | "SUCCESS" | "WARN" | "ERROR", msg: string, detail?: string): void;
}>();

const filteredRuntimes = computed(() => {
  if (!searchKey.value.trim()) return runtimes.value;
  const q = searchKey.value.toLowerCase();
  return runtimes.value.filter(
    (rt) =>
      rt.name.toLowerCase().includes(q) ||
      rt.id.toLowerCase().includes(q) ||
      rt.description.toLowerCase().includes(q) ||
      rt.active_version.includes(q)
  );
});

const filteredRemoteVersions = computed(() => {
  if (!remoteSearchKey.value.trim()) return remoteVersions.value;
  const q = remoteSearchKey.value.toLowerCase();
  return remoteVersions.value.filter((v) => v.version.toLowerCase().includes(q));
});

function getRuntimeIcon(id: string): string {
  switch (id) {
    case "node":
    case "nodejs":
      return "🟩";
    case "go":
    case "golang":
      return "🔷";
    case "python":
    case "python3":
      return "🐍";
    case "rust":
      return "🦀";
    case "java":
      return "☕";
    case "ruby":
      return "💎";
    default:
      return "⚙️";
  }
}

function getSourceTagClass(source: string): string {
  if (!source) return "source-default";
  if (source.includes(".nvmrc") || source.includes(".node-version") || source.includes(".go-version")) {
    return "source-legacy";
  }
  if (source.includes("unirtm.toml")) return "source-project";
  if (source === "global") return "source-global";
  return "source-default";
}

function isAlreadyInstalled(toolId: string, ver: string): boolean {
  const rt = runtimes.value.find((r) => r.id === toolId);
  if (!rt || !rt.installed_versions) return false;
  return rt.installed_versions.includes(ver) || rt.installed_versions.includes(ver.replace(/^v/, ""));
}

async function fetchStatus() {
  loading.value = true;
  try {
    const res = await fetch("/api/runtimes/status");
    if (res.ok) {
      const data = await res.json();
      runtimes.value = data.runtimes || [];
      totalInstalled.value = data.total_installed || 0;
      detectedLegacies.value = data.detected_legacies || [];
    }
  } catch (err: any) {
    emit("log-event", "ERROR", "Failed to fetch runtimes status: " + err.message);
  } finally {
    loading.value = false;
  }
}

async function quickSwitch(toolId: string, ver: string) {
  try {
    const res = await fetch("/api/runtimes/use", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ tool: toolId, version: ver, global: false }),
    });
    if (res.ok) {
      emit("log-event", "SUCCESS", `Switched ${toolId} to ${ver}`);
      await fetchStatus();
    }
  } catch (err: any) {
    emit("log-event", "ERROR", `Failed to switch version: ${err.message}`);
  }
}

function handleSelectChange(toolId: string, evt: Event) {
  const select = evt.target as HTMLSelectElement;
  if (select && select.value) {
    quickSwitch(toolId, select.value);
  }
}

async function migrateLegacy() {
  migrating.value = true;
  try {
    const res = await fetch("/api/runtimes/migrate-legacy", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({}),
    });
    if (res.ok) {
      emit("log-event", "SUCCESS", "Successfully migrated legacy version files into .unirtm.toml");
      await fetchStatus();
    }
  } catch (err: any) {
    emit("log-event", "ERROR", `Migration failed: ${err.message}`);
  } finally {
    migrating.value = false;
  }
}

async function openRemoteModal(rt: RuntimeItem) {
  selectedRuntime.value = rt;
  isModalOpen.value = true;
  remoteVersions.value = [];
  remoteSearchKey.value = "";
  loadingRemote.value = true;

  try {
    const res = await fetch(`/api/runtimes/remote?tool=${rt.id}`);
    if (res.ok) {
      const data = await res.json();
      remoteVersions.value = data.versions || [];
    }
  } catch (err: any) {
    emit("log-event", "ERROR", `Failed to fetch remote versions: ${err.message}`);
  } finally {
    loadingRemote.value = false;
  }
}

async function installRemoteVersion(toolId: string, ver: string) {
  installingVer.value = ver;
  try {
    // Write and activate version
    await fetch("/api/runtimes/use", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ tool: toolId, version: ver, global: false }),
    });
    emit("log-event", "INFO", `Initiated install for ${toolId}@${ver}`);
    isModalOpen.value = false;
    await fetchStatus();
  } catch (err: any) {
    emit("log-event", "ERROR", `Failed to install version: ${err.message}`);
  } finally {
    installingVer.value = null;
  }
}

onMounted(() => {
  fetchStatus();
});
</script>

<style scoped>
.runtimes-view {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  padding: 1.5rem;
}

.view-header-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.view-title {
  font-size: 1.5rem;
  font-weight: 700;
  margin: 0 0 0.35rem 0;
  color: var(--color-foreground, #fff);
}

.view-desc {
  font-size: 0.875rem;
  color: var(--color-foreground-muted, #94a3b8);
  margin: 0;
}

.tool-action-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.45rem 0.85rem;
  border-radius: 6px;
  border: 1px solid rgba(255, 255, 255, 0.15);
  background: rgba(255, 255, 255, 0.08);
  color: #fff;
  cursor: pointer;
  font-size: 0.85rem;
  transition: all 0.2s ease;
}

.tool-action-btn.primary {
  background: #3b82f6;
  border-color: #2563eb;
}

.tool-action-btn:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.15);
}

.tool-action-btn.primary:hover:not(:disabled) {
  background: #2563eb;
}

.spin-icon {
  display: inline-block;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  100% {
    transform: rotate(360deg);
  }
}

/* Legacy Banner */
.legacy-banner {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.85rem 1.25rem;
  border-left: 4px solid #f59e0b;
  background: rgba(245, 158, 11, 0.08);
}

.legacy-banner-left {
  display: flex;
  align-items: center;
  gap: 0.85rem;
}

.legacy-banner-icon {
  font-size: 1.4rem;
}

.legacy-banner-content {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}

.legacy-banner-title {
  font-size: 0.875rem;
  font-weight: 600;
  color: #fbbf24;
}

.legacy-banner-sub {
  font-size: 0.775rem;
  color: var(--color-foreground-muted, #94a3b8);
}

/* Filter bar */
.runtimes-filter-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.65rem 1rem;
}

.search-input-wrapper {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex: 1;
  max-width: 400px;
}

.tool-search-input {
  width: 100%;
  background: transparent;
  border: none;
  outline: none;
  color: inherit;
  font-size: 0.875rem;
}

.runtime-stats-badges {
  display: flex;
  gap: 0.5rem;
}

.stat-badge {
  font-size: 0.775rem;
  padding: 0.25rem 0.65rem;
  border-radius: 9999px;
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.stat-badge.highlight {
  background: rgba(59, 130, 246, 0.15);
  border-color: rgba(59, 130, 246, 0.3);
  color: #60a5fa;
}

/* Grid */
.runtimes-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 1.25rem;
}

.runtime-card {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding: 1.25rem;
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  background: rgba(30, 41, 59, 0.4);
  transition: transform 0.2s ease, border-color 0.2s ease;
}

.runtime-card:hover {
  border-color: rgba(59, 130, 246, 0.4);
  transform: translateY(-2px);
}

.runtime-card-header {
  display: flex;
  align-items: flex-start;
  gap: 0.85rem;
}

.runtime-icon-badge {
  font-size: 1.75rem;
  line-height: 1;
}

.runtime-title-block {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
  flex: 1;
}

.runtime-name-row {
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

.runtime-name {
  font-size: 1.1rem;
  font-weight: 600;
  color: #fff;
}

.runtime-homepage-link {
  color: var(--color-foreground-muted, #94a3b8);
  text-decoration: none;
  font-size: 0.8rem;
}

.runtime-desc {
  font-size: 0.775rem;
  color: var(--color-foreground-muted, #94a3b8);
  line-height: 1.35;
}

/* Status box */
.runtime-status-box {
  background: rgba(15, 23, 42, 0.5);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 6px;
  padding: 0.65rem 0.85rem;
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.status-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 0.8rem;
}

.status-label {
  color: var(--color-foreground-muted, #94a3b8);
}

.status-val-wrap {
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

.active-pill {
  font-size: 0.775rem;
  font-weight: 600;
  color: #34d399;
}

.source-tag {
  font-size: 0.65rem;
  padding: 0.1rem 0.4rem;
  border-radius: 4px;
}

.source-legacy {
  background: rgba(245, 158, 11, 0.2);
  color: #f59e0b;
}

.source-project {
  background: rgba(59, 130, 246, 0.2);
  color: #60a5fa;
}

.source-global {
  background: rgba(168, 85, 247, 0.2);
  color: #c084fc;
}

.source-default {
  background: rgba(255, 255, 255, 0.08);
  color: #cbd5e1;
}

.global-val {
  font-size: 0.775rem;
  color: var(--color-foreground-muted, #94a3b8);
}

/* Installed versions */
.installed-versions-section {
  display: flex;
  flex-direction: column;
  gap: 0.45rem;
}

.section-title-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.75rem;
  color: var(--color-foreground-muted, #94a3b8);
}

.count-pill {
  font-size: 0.7rem;
  padding: 0.05rem 0.4rem;
  background: rgba(255, 255, 255, 0.06);
  border-radius: 9999px;
}

.versions-pills-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
}

.ver-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  font-size: 0.75rem;
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: inherit;
  cursor: pointer;
  transition: all 0.15s ease;
}

.ver-pill:hover {
  background: rgba(59, 130, 246, 0.2);
  border-color: #3b82f6;
}

.ver-pill.selected {
  background: rgba(16, 185, 129, 0.2);
  border-color: #10b981;
  color: #34d399;
}

.active-dot {
  font-size: 0.6rem;
}

.empty-installed-text {
  font-size: 0.75rem;
  color: var(--color-foreground-muted, #64748b);
  font-style: italic;
}

/* Card footer */
.runtime-card-footer {
  margin-top: auto;
  display: flex;
  gap: 0.5rem;
  align-items: center;
}

.switch-control-group {
  flex: 1;
}

.runtime-select {
  width: 100%;
  padding: 0.35rem 0.5rem;
  border-radius: 4px;
  background: rgba(15, 23, 42, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.12);
  color: inherit;
  font-size: 0.8rem;
  outline: none;
}

.install-remote-btn {
  padding: 0.35rem 0.65rem;
  border-radius: 4px;
  background: rgba(59, 130, 246, 0.15);
  border: 1px solid rgba(59, 130, 246, 0.3);
  color: #60a5fa;
  font-size: 0.8rem;
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.15s ease;
}

.install-remote-btn:hover {
  background: rgba(59, 130, 246, 0.3);
}

/* Modal */
.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.65);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-dialog {
  width: 90%;
  max-width: 520px;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
  padding: 1.25rem;
  background: #1e293b;
  border: 1px solid rgba(255, 255, 255, 0.15);
  border-radius: 8px;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.5);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 0.85rem;
}

.modal-title-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.modal-icon {
  font-size: 1.75rem;
}

.modal-title {
  margin: 0;
  font-size: 1.1rem;
  font-weight: 600;
  color: #fff;
}

.modal-desc {
  margin: 0.2rem 0 0 0;
  font-size: 0.775rem;
  color: var(--color-foreground-muted, #94a3b8);
}

.modal-close-btn {
  background: transparent;
  border: none;
  color: var(--color-foreground-muted, #94a3b8);
  font-size: 1.2rem;
  cursor: pointer;
}

.modal-search-row {
  margin-bottom: 0.75rem;
}

.modal-search-input {
  width: 100%;
  padding: 0.5rem 0.75rem;
  border-radius: 6px;
  background: rgba(15, 23, 42, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.15);
  color: inherit;
  font-size: 0.85rem;
  outline: none;
  box-sizing: border-box;
}

.modal-body-versions {
  flex: 1;
  overflow-y: auto;
  max-height: 380px;
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.remote-loading-box,
.remote-empty-box {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 2rem 0;
  color: var(--color-foreground-muted, #94a3b8);
  font-size: 0.875rem;
}

.remote-ver-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.5rem 0.75rem;
  border-radius: 4px;
  background: rgba(255, 255, 255, 0.04);
}

.remote-ver-row:hover {
  background: rgba(255, 255, 255, 0.08);
}

.remote-ver-info {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.ver-name {
  font-size: 0.85rem;
  font-weight: 500;
}

.installed-tag {
  font-size: 0.7rem;
  color: #34d399;
}

.install-action-btn {
  padding: 0.25rem 0.65rem;
  font-size: 0.75rem;
  border-radius: 4px;
  background: #3b82f6;
  border: 1px solid #2563eb;
  color: #fff;
  cursor: pointer;
  transition: background 0.15s ease;
}

.install-action-btn:hover:not(:disabled) {
  background: #2563eb;
}

.install-action-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.modal-footer {
  margin-top: 0.85rem;
  display: flex;
  justify-content: flex-end;
}
</style>
