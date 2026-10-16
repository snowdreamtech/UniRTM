<template>
  <div class="view-container tools-view" id="view-tools">
    <!-- View Header Row -->
    <div class="view-header-row">
      <div>
        <h2 class="view-title" style="display: flex; align-items: center; gap: 8px;">
          <span class="view-icon">
            <TechIcon id="workbench:tools" :size="20" />
          </span>
          <span>{{ t("nav.tools") || "开发环境与工具链" }}</span>
        </h2>
        <p class="view-desc">
          由 UniRTM 统一版本锁定与高性能垫片（Go Shims）路由，替换 NVM、GVM、Pyenv 等异构管理器，提供跨平台一致的工具链环境。
        </p>
      </div>
      <div class="view-actions">
        <button class="tool-action-btn secondary" @click="fetchStatus" :disabled="loading">
          <TechIcon id="workbench:sync" :size="14" :class="{ 'spin-icon': loading }" />
          <span>{{ loading ? t("common.refreshing") : t("common.refresh") }}</span>
        </button>
      </div>
    </div>

    <!-- Scope Selection Bar (User Scope vs Workspace Scope) -->
    <div class="scope-selection-bar glass-card" id="runtimes-scope-bar">
      <div class="scope-tabs-bar">
        <button
          class="scope-tab-btn"
          :class="{ active: currentScope === 'workspace' }"
          @click="currentScope = 'workspace'"
          id="btn-scope-workspace"
        >
          <span class="scope-icon">
            <TechIcon id="workbench:tools" :size="14" :colored="false" />
          </span>
          <span class="scope-title">{{ t("runtimes.scope_workspace") || "工作区" }}</span>
        </button>

        <button
          class="scope-tab-btn"
          :class="{ active: currentScope === 'user' }"
          @click="currentScope = 'user'"
          id="btn-scope-user"
        >
          <span class="scope-icon">
            <TechIcon id="workbench:settings" :size="14" :colored="false" />
          </span>
          <span class="scope-title">{{ t("runtimes.scope_user") || "用户" }}</span>
        </button>
      </div>

      <div class="scope-hint-text">
        <span v-if="currentScope === 'user'">
          {{ t("runtimes.scope_user_desc") || "配置当前用户所有未指定工作区时的默认兜底环境" }}
        </span>
        <span v-else>
          {{ t("runtimes.scope_workspace_desc") || "锁定当前工作区语言版本，团队协作检出即生效" }}
        </span>
      </div>
    </div>

    <!-- Legacy Version Files Banner (Workspace Scope Only) -->
    <div v-if="currentScope === 'workspace' && detectedLegacies.length > 0" class="legacy-banner glass-card" id="legacy-detected-banner">
      <div class="legacy-banner-left">
        <span class="legacy-banner-icon">
          <TechIcon id="workbench:about" :size="16" />
        </span>
        <div class="legacy-banner-content">
          <span class="legacy-banner-title">
            {{ t("runtimes.legacy_banner", { files: detectedLegacies.join(", ") }) }}
          </span>
          <span class="legacy-banner-sub">
            Seamlessly active via zero-configuration compatibility. You can keep them or import to workspace configuration for unified governance.
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
          <span style="display: inline-flex; align-items: center; gap: 4px;">
            <TechIcon id="workbench:tasks" :size="12" :colored="false" />
            {{ migrating ? "Importing..." : (t("runtimes.import_legacy") || "一键导入至工作区") }}
          </span>
        </button>
      </div>
    </div>

    <!-- Scope Filter & Stats Toolbar -->
    <div class="runtimes-filter-bar glass-card">
      <div class="search-input-wrapper">
        <span class="search-icon">
          <TechIcon id="workbench:doctor" :size="14" :colored="false" />
        </span>
        <input
          v-model="searchKey"
          type="text"
          class="tool-search-input"
          :placeholder="t('runtimes.search_placeholder') || 'Search or filter runtimes...'"
        />
      </div>
      <div class="runtime-stats-badges">
        <span class="stat-badge">
          <b>{{ filteredRuntimes.length }}</b> {{ t("runtimes.count_runtimes_suffix") || "个工具" }}
        </span>
        <span class="stat-badge highlight">
          <b>{{ totalInstalled }}</b> {{ t("runtimes.count_installed_suffix") || "个已装版本" }}
        </span>
        <button
          class="tool-action-btn primary small"
          id="btn-add-runtime"
          @click="openAddModal"
        >
          <span>+ {{ t("runtimes.add_runtime") || "添加运行环境" }}</span>
        </button>
      </div>
    </div>

    <!-- Runtimes Grid Cards -->
    <div class="runtimes-grid">
      <div
        v-for="rt in filteredRuntimes"
        :key="rt.id"
        class="glass-card runtime-card"
        :class="{ active: currentScope === 'workspace' ? rt.active_version : rt.global_version }"
        :id="'runtime-card-' + rt.id"
      >
        <!-- Card Header -->
        <div class="runtime-card-header">
          <div class="runtime-icon-badge"><TechIcon :id="rt.id" :size="32" :colored="true" /></div>
          <div class="runtime-title-block">
            <div class="runtime-name-row">
              <span class="runtime-name" :title="rt.name">{{ formatToolTitle(rt.name) }}</span>
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
              <button
                class="runtime-pin-btn"
                :class="{ 'is-pinned': rt.pinned }"
                :title="rt.pinned ? (rt.pin_reason ? `版本已锁定：${rt.pin_reason} (点击解锁)` : '版本已锁定 (点击解锁)') : '锁定当前版本，防止意外升级'"
                @click.stop="togglePin(rt)"
              >
                <span class="pin-icon">{{ rt.pinned ? '🔒' : '🔓' }}</span>
                <span v-if="rt.pinned" class="pin-label">已锁定</span>
              </button>
              <button
                class="runtime-delete-btn"
                :title="(t('runtimes.remove_runtime') || '移除环境') + ' ' + rt.name"
                @click.stop="confirmRemoveRuntime(rt)"
              >
                <TechIcon id="workbench:delete" :size="17" />
              </button>
            </div>
            <span class="runtime-desc">{{ rt.description }}</span>
            <div v-if="rt.pinned && rt.pin_reason" class="pin-reason-tag" :title="rt.pin_reason">
              <span class="pin-reason-icon">📌</span>
              <span class="pin-reason-text">{{ rt.pin_reason }}</span>
            </div>
          </div>
        </div>


        <!-- Dual Scope Version Display -->
        <div class="runtime-status-box">
          <!-- Workspace Scope Version (First) -->
          <div class="status-row" :class="{ 'highlight-scope': currentScope === 'workspace' }">
            <span class="status-label">
              <span class="scope-mini-icon">
                <TechIcon id="workbench:tools" :size="15" :colored="false" />
              </span>
              {{ t("runtimes.scope_workspace") || "工作区" }}:
            </span>
            <div v-if="rt.active_version" class="status-val-wrap">
              <span class="active-pill font-mono">{{ formatVersion(rt.active_version) }}</span>
              <span v-if="rt.pinned" class="source-tag source-pinned" :title="rt.pin_reason ? `版本锁定: ${rt.pin_reason}` : '版本已锁定'">
                🔒 锁定
              </span>
              <span v-else-if="rt.is_legacy" class="source-tag source-legacy">
                {{ rt.source }}
              </span>
            </div>
            <div v-else-if="rt.global_version" class="status-val-wrap">
              <span class="active-pill font-mono muted">{{ formatVersion(rt.global_version) }}</span>
              <span class="source-tag source-default">{{ t("runtimes.follow_user") || "跟随用户" }}</span>
            </div>
            <span v-else class="text-muted font-mono italic">未配置</span>
          </div>

          <!-- User Scope Version (Second) -->
          <div class="status-row" :class="{ 'highlight-scope': currentScope === 'user' }">
            <span class="status-label">
              <span class="scope-mini-icon">
                <TechIcon id="workbench:settings" :size="15" :colored="false" />
              </span>
              {{ t("runtimes.scope_user") || "用户" }}:
            </span>
            <div v-if="rt.global_version" class="status-val-wrap">
              <span class="global-pill font-mono">{{ formatVersion(rt.global_version) }}</span>
            </div>
            <span v-else class="text-muted font-mono italic">未配置</span>
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
              :class="{
                selected: currentScope === 'workspace' ? ver === rt.active_version : ver === rt.global_version
              }"
              :title="'点击设为' + (currentScope === 'user' ? '用户' : '工作区') + '版本 (v' + ver + ')'"
              @click="quickSwitch(rt.id, ver)"
            >
              <span
                v-if="(currentScope === 'workspace' && ver === rt.active_version) || (currentScope === 'user' && ver === rt.global_version)"
                class="active-dot"
              >●</span>
              <span>{{ formatVersion(ver) }}</span>
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
              :value="currentScope === 'workspace' ? rt.active_version : rt.global_version"
              @change="handleSelectChange(rt.id, $event)"
            >
              <option value="" disabled>
                {{ currentScope === 'user' ? '切换用户默认版本...' : '切换工作区版本...' }}
              </option>
              <option
                v-for="ver in rt.installed_versions"
                :key="ver"
                :value="ver"
              >
                {{ formatVersion(ver) }}{{ (currentScope === 'workspace' ? ver === rt.active_version : ver === rt.global_version) ? ' (当前)' : '' }}
              </option>
            </select>
          </div>
          <button
            class="install-remote-btn"
            :id="'btn-install-remote-' + rt.id"
            @click="openRemoteModal(rt)"
          >
            <span>+ {{ t("runtimes.install_version") || "Install Version" }}</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Remote Versions Modal -->
    <div v-if="isModalOpen" class="modal-backdrop" @click.self="closeRemoteModal">
      <div class="modal-dialog glass-card" role="dialog" aria-modal="true">
        <div class="modal-header">
          <div class="modal-title-row">
            <span class="modal-icon"><TechIcon :id="selectedRuntime?.id || ''" :size="20" :colored="true" /></span>
            <div>
              <h3 class="modal-title">
                {{ selectedRuntime?.name }} {{ t("runtimes.install_version") || "Install Version" }}
              </h3>
              <p class="modal-desc">
                {{ t("runtimes.remote_list_tip") || "Available official upstream releases" }}
              </p>
            </div>
          </div>
          <button class="modal-close-btn" @click="closeRemoteModal">✕</button>
        </div>

        <!-- Scope Option Inside Modal -->
        <div class="modal-scope-picker">
          <label class="modal-scope-item" :class="{ active: modalInstallScope === 'workspace' }">
            <input type="radio" value="workspace" v-model="modalInstallScope" name="install-scope" />
            <span class="scope-icon">
              <TechIcon id="workbench:workspace" :size="16" :colored="false" />
            </span>
            <div class="scope-opt-text">
              <span class="scope-opt-name">{{ t("runtimes.scope_workspace") || "工作区" }}</span>
              <span class="scope-opt-desc">{{ t("runtimes.scope_workspace_tip") || "应用于当前工作区，随团队项目协作生效" }}</span>
            </div>
          </label>
          <label class="modal-scope-item" :class="{ active: modalInstallScope === 'user' }">
            <input type="radio" value="user" v-model="modalInstallScope" name="install-scope" />
            <span class="scope-icon">
              <TechIcon id="workbench:user" :size="16" :colored="false" />
            </span>
            <div class="scope-opt-text">
              <span class="scope-opt-name">{{ t("runtimes.scope_user") || "用户" }}</span>
              <span class="scope-opt-desc">{{ t("runtimes.scope_user_tip") || "应用于本机所有工作区的默认兜底版本" }}</span>
            </div>
          </label>
        </div>

        <div class="modal-search-row">
          <input
            v-model="remoteSearchKey"
            type="text"
            class="modal-search-input"
            :placeholder="t('runtimes.search_remote_placeholder') || 'Filter remote versions (e.g. 20, lts, 1.23)...'"
          />
          <span class="version-count-pill font-mono">
            {{ filteredRemoteVersions.length }} {{ t("runtimes.version_count") || "个可用版本" }}
          </span>
        </div>

        <!-- Quick Major Filter Tags -->
        <div v-if="quickFilterTags.length > 0" class="quick-filter-chips">
          <button
            class="filter-chip"
            :class="{ active: selectedMajorFilter === '' }"
            @click="selectedMajorFilter = ''"
          >
            {{ t('runtimes.all_versions') || '全部' }}
          </button>
          <button
            v-for="tag in quickFilterTags"
            :key="tag"
            class="filter-chip"
            :class="{ active: selectedMajorFilter === tag }"
            @click="selectedMajorFilter = selectedMajorFilter === tag ? '' : tag"
          >
            {{ tag }}
          </button>
        </div>

        <div class="modal-body-versions">
          <div v-if="loadingRemote" class="remote-loading-box">
            <span class="spin-icon">
              <TechIcon id="workbench:sync" :size="16" />
            </span>
            <span>{{ t("runtimes.fetching_releases") || "Fetching available releases..." }}</span>
          </div>

          <div v-else-if="filteredRemoteVersions.length === 0" class="remote-empty-box">
            <div class="empty-icon-wrap">
              <TechIcon id="workbench:search" :size="24" />
            </div>
            <div class="empty-title">{{ t("runtimes.no_matching_releases") || "未找到匹配的官方发布版本" }}</div>

            <!-- Version format / rule validation error -->
            <div v-if="versionValidation.error" class="version-validation-error">
              <span class="warn-icon">⚠️</span>
              <span class="warn-text">{{ versionValidation.error }}</span>
            </div>

            <!-- Strictly validated custom version card when unlisted -->
            <div v-else-if="canInstallCustomVersion" class="custom-ver-install-box">
              <div class="custom-ver-header">
                <span class="font-mono ver-name">v{{ cleanCustomVersion }}</span>
                <span class="custom-badge font-mono">
                  {{ t('runtimes.custom_version_unlisted') || '官方未收录' }}
                </span>
              </div>
              <p class="custom-ver-desc">
                {{ t('runtimes.custom_version_desc') || '该版本未包含在当前官方版本清单中。若您确认存在此构建，可尝试自定义安装。' }}
              </p>
              <button
                class="install-action-btn"
                :disabled="installingVer === cleanCustomVersion"
                @click="installRemoteVersion(selectedRuntime?.id || '', cleanCustomVersion)"
              >
                <span v-if="installingVer === cleanCustomVersion">{{ t("runtimes.installing") || "Installing..." }}</span>
                <span v-else>{{ t("runtimes.install_custom_btn") || "尝试安装自定义版本" }}</span>
              </button>
            </div>
          </div>

          <div
            v-else
            v-for="rv in filteredRemoteVersions"
            :key="rv.version"
            class="remote-ver-row"
          >
            <div class="remote-ver-info">
              <span class="font-mono ver-name">v{{ rv.version }}</span>
              <span
                v-if="isAlreadyInstalled(selectedRuntime?.id || '', rv.version)"
                class="installed-tag"
              >
                {{ t("runtimes.installed") || "已安装" }}
              </span>
            </div>

            <button
              class="install-action-btn"
              :class="{ 'is-installed': isAlreadyInstalled(selectedRuntime?.id || '', rv.version) }"
              :disabled="installingVer === rv.version"
              @click="installRemoteVersion(selectedRuntime?.id || '', rv.version)"
            >
              <span v-if="installingVer === rv.version">{{ t("runtimes.installing") || "Installing..." }}</span>
              <span v-else-if="isAlreadyInstalled(selectedRuntime?.id || '', rv.version)">
                {{ modalInstallScope === 'user' ? (t("runtimes.set_user") || '设为用户') : (t("runtimes.set_workspace") || '设为工作区') }}
              </span>
              <span v-else>{{ t("runtimes.install_apply") || "安装并应用" }}</span>
            </button>
          </div>
        </div>

        <div class="modal-footer">
          <button class="tool-action-btn secondary" @click="closeRemoteModal">
            {{ t("common.close") || "Close" }}
          </button>
        </div>
      </div>
    </div>

    <!-- Add New Runtime Modal -->
    <div v-if="isAddModalOpen" class="modal-backdrop" @click.self="isAddModalOpen = false">
      <div class="modal-dialog glass-card add-runtime-dialog" role="dialog" aria-modal="true">
        <div class="modal-header">
          <div class="modal-title-row">
            <span class="modal-icon">
              <TechIcon id="workbench:tools" :size="20" />
            </span>
            <div>
              <h3 class="modal-title">{{ t("runtimes.add_runtime") || "添加工具" }}</h3>
              <p class="modal-desc">为当前{{ currentScope === 'user' ? '用户默认' : '工作区' }}添加工具与版本</p>
            </div>
          </div>
          <button class="modal-close-btn" @click="isAddModalOpen = false">✕</button>
        </div>

        <div class="add-preset-section">
          <div class="preset-section-title">官方支持推荐语言</div>
          <div class="add-preset-grid">
            <div
              v-for="preset in availablePresetRuntimes"
              :key="preset.id"
              class="add-preset-card"
            >
              <div class="preset-card-left">
                <span class="preset-icon"><TechIcon :id="preset.id" :size="20" :colored="true" /></span>
                <div class="preset-info">
                  <div class="preset-name-row">
                    <span class="preset-name">{{ preset.name }}</span>
                    <span class="preset-tag font-mono">v{{ preset.defaultVersion }}</span>
                  </div>
                  <span class="preset-desc">{{ preset.description }}</span>
                </div>
              </div>
              <button
                class="tool-action-btn primary small"
                @click="addRuntimeFromPreset(preset)"
              >
                + 添加
              </button>
            </div>
          </div>
        </div>

        <div class="custom-add-section">
          <div class="preset-section-title">或输入自定义工具名</div>
          <div class="custom-add-form">
            <input
              v-model="customToolInput"
              type="text"
              class="custom-tool-input font-mono"
              placeholder="例如: elixir, zig, php..."
              @keyup.enter="addCustomRuntime"
            />
            <button
              class="tool-action-btn primary small"
              :disabled="!customToolInput.trim()"
              @click="addCustomRuntime"
            >
              添加
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>

</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from "vue";
import { t } from "../i18n";
import { useToast } from "../composables/useToast";
import TechIcon from "../components/common/TechIcon.vue";
import { validateRuntimeVersion } from "../utils/versionValidator";

const { showToast } = useToast();

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
  pinned?: boolean;
  pin_reason?: string;
}

const props = defineProps<{
  activeSubItem?: string;
}>();

const currentScope = ref<"user" | "workspace">("workspace");
const modalInstallScope = ref<"user" | "workspace">("workspace");

const defaultRuntimes: RuntimeItem[] = [
  {
    id: "node",
    name: "Node.js",
    description: "JavaScript / TypeScript runtime (NVM superset)",
    homepage: "https://nodejs.org",
    installed_versions: ["20.18.3", "26.8.1"],
    active_version: "20.18.3",
    global_version: "26.8.1",
    source: ".nvmrc",
    is_legacy: true,
    legacy_file: ".nvmrc",
  },
  {
    id: "go",
    name: "Go",
    description: "Go programming language (GVM superset)",
    homepage: "https://go.dev",
    installed_versions: ["1.27.1", "1.27.2"],
    active_version: "1.27.2",
    global_version: "1.27.2",
    source: ".unirtm.toml",
    is_legacy: false,
    legacy_file: "",
  },
  {
    id: "python",
    name: "Python",
    description: "Python language interpreter (Pyenv superset)",
    homepage: "https://python.org",
    installed_versions: ["3.12.2", "3.14.7"],
    active_version: "3.14.7",
    global_version: "3.14.7",
    source: ".unirtm.toml",
    is_legacy: false,
    legacy_file: "",
  },
  {
    id: "rust",
    name: "Rust",
    description: "Rust toolchain & cargo build system",
    homepage: "https://rust-lang.org",
    installed_versions: ["1.85.0"],
    active_version: "1.85.0",
    global_version: "1.85.0",
    source: ".unirtm.toml",
    is_legacy: false,
    legacy_file: "",
  },
];

const runtimes = ref<RuntimeItem[]>([...defaultRuntimes]);
const totalInstalled = ref(7);
const detectedLegacies = ref<string[]>([".nvmrc", ".go-version"]);
const loading = ref(false);
const migrating = ref(false);
const searchKey = ref("");

// Remote modal state
const isModalOpen = ref(false);
const selectedRuntime = ref<RuntimeItem | null>(null);
const remoteVersions = ref<{ version: string }[]>([]);
const remoteSearchKey = ref("");
const selectedMajorFilter = ref("");
const loadingRemote = ref(false);
const installingVer = ref<string | null>(null);
let remoteModalAbortController: AbortController | null = null;

function closeRemoteModal() {
  if (remoteModalAbortController) {
    remoteModalAbortController.abort();
    remoteModalAbortController = null;
  }
  isModalOpen.value = false;
  selectedRuntime.value = null;
  remoteVersions.value = [];
  remoteSearchKey.value = "";
  selectedMajorFilter.value = "";
  loadingRemote.value = false;
}

// Add Runtime modal state
const isAddModalOpen = ref(false);
const customToolInput = ref("");

interface PresetRuntime {
  id: string;
  name: string;
  description: string;
  homepage: string;
  defaultVersion: string;
}

const presetRuntimes: PresetRuntime[] = [
  {
    id: "bun",
    name: "Bun",
    description: "Incredibly fast all-in-one JavaScript & TypeScript toolkit",
    homepage: "https://bun.sh",
    defaultVersion: "1.2.4",
  },
  {
    id: "deno",
    name: "Deno",
    description: "Modern, secure runtime for JavaScript and TypeScript",
    homepage: "https://deno.com",
    defaultVersion: "2.2.2",
  },
  {
    id: "node",
    name: "Node.js",
    description: "JavaScript / TypeScript runtime (NVM superset)",
    homepage: "https://nodejs.org",
    defaultVersion: "22.14.0",
  },
  {
    id: "go",
    name: "Go",
    description: "Go programming language (GVM superset)",
    homepage: "https://go.dev",
    defaultVersion: "1.24.0",
  },
  {
    id: "python",
    name: "Python",
    description: "Python language interpreter (Pyenv superset)",
    homepage: "https://python.org",
    defaultVersion: "3.13.2",
  },
  {
    id: "rust",
    name: "Rust",
    description: "Rust toolchain & cargo build system",
    homepage: "https://rust-lang.org",
    defaultVersion: "1.85.0",
  },
  {
    id: "java",
    name: "Java (OpenJDK)",
    description: "Enterprise multi-platform Java Development Kit & Virtual Machine",
    homepage: "https://openjdk.org",
    defaultVersion: "21.0.2",
  },
  {
    id: "zig",
    name: "Zig",
    description: "General-purpose programming language and toolchain for robust software",
    homepage: "https://ziglang.org",
    defaultVersion: "0.13.0",
  },
  {
    id: "ruby",
    name: "Ruby",
    description: "Dynamic, open source programming language with a focus on simplicity",
    homepage: "https://ruby-lang.org",
    defaultVersion: "3.4.1",
  },
  {
    id: "php",
    name: "PHP",
    description: "Popular general-purpose scripting language suited for web development",
    homepage: "https://php.net",
    defaultVersion: "8.4.4",
  },
  {
    id: "dart",
    name: "Dart",
    description: "Client-optimized language for fast apps on any platform",
    homepage: "https://dart.dev",
    defaultVersion: "3.6.1",
  },
  {
    id: "flutter",
    name: "Flutter",
    description: "Google's UI toolkit for multi-platform applications",
    homepage: "https://flutter.dev",
    defaultVersion: "3.29.0",
  },
  {
    id: "julia",
    name: "Julia",
    description: "High-performance dynamic language for technical & scientific computing",
    homepage: "https://julialang.org",
    defaultVersion: "1.11.3",
  },
  {
    id: "elixir",
    name: "Elixir",
    description: "Dynamic, functional language for building scalable & maintainable software",
    homepage: "https://elixir-lang.org",
    defaultVersion: "1.18.2",
  },
  {
    id: "erlang",
    name: "Erlang / OTP",
    description: "Soft real-time distributed system language & OTP runtime",
    homepage: "https://erlang.org",
    defaultVersion: "27.2.1",
  },
  {
    id: "haskell",
    name: "Haskell (GHC)",
    description: "Purely functional programming language with Glasgow Haskell Compiler",
    homepage: "https://haskell.org",
    defaultVersion: "9.12.1",
  },
];

const availablePresetRuntimes = computed(() => {
  const currentIds = new Set(runtimes.value.map((r) => r.id.toLowerCase()));
  return presetRuntimes.filter((p) => !currentIds.has(p.id.toLowerCase()));
});

function openAddModal() {
  customToolInput.value = "";
  isAddModalOpen.value = true;
}

async function addRuntimeFromPreset(preset: PresetRuntime) {
  const isGlobal = currentScope.value === "user";
  const newItem: RuntimeItem = {
    id: preset.id,
    name: preset.name,
    description: preset.description,
    homepage: preset.homepage,
    installed_versions: [preset.defaultVersion],
    active_version: isGlobal ? "" : preset.defaultVersion,
    global_version: isGlobal ? preset.defaultVersion : "",
    source: isGlobal ? "~/.config/unirtm/unirtm.toml" : ".unirtm.toml",
    is_legacy: false,
    legacy_file: "",
  };

  runtimes.value.push(newItem);
  isAddModalOpen.value = false;
  showToast(`已成功添加 ${preset.name} 环境 (v${preset.defaultVersion})`, "success");

  // Persist to .unirtm.toml & unirtm.lock
  await quickSwitch(preset.id, preset.defaultVersion, isGlobal);
}

async function addCustomRuntime() {
  const toolName = customToolInput.value.trim().toLowerCase();
  if (!toolName) return;

  const isGlobal = currentScope.value === "user";
  const newItem: RuntimeItem = {
    id: toolName,
    name: toolName.charAt(0).toUpperCase() + toolName.slice(1),
    description: `${toolName} runtime environment`,
    homepage: "",
    installed_versions: ["latest"],
    active_version: isGlobal ? "" : "latest",
    global_version: isGlobal ? "latest" : "",
    source: isGlobal ? "~/.config/unirtm/unirtm.toml" : ".unirtm.toml",
    is_legacy: false,
    legacy_file: "",
  };

  runtimes.value.push(newItem);
  customToolInput.value = "";
  isAddModalOpen.value = false;
  showToast(`已成功添加 ${newItem.name} 环境`, "success");

  await quickSwitch(toolName, "latest", isGlobal);
}

async function confirmRemoveRuntime(rt: RuntimeItem) {
  const isGlobal = currentScope.value === "user";
  const scopeName = isGlobal ? "用户配置" : "工作区";
  const msg = (t("runtimes.remove_confirm") || "确定从当前工作区中移除 {name} 吗？").replace("{name}", rt.name);
  if (!window.confirm(msg)) {
    return;
  }

  // 1. Optimistic removal
  const prevRuntimes = JSON.parse(JSON.stringify(runtimes.value));
  runtimes.value = runtimes.value.filter((r) => r.id !== rt.id);
  showToast(`正在从${scopeName}移除 ${rt.name}...`, "info");

  try {
    const res = await fetch("/api/tools/remove", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ tool: rt.id, global: isGlobal }),
    });
    if (!res.ok) {
      const errData = await res.json().catch(() => ({}));
      throw new Error(errData.error || `HTTP ${res.status}`);
    }
    showToast(`已成功移除 ${rt.name}`, "success");
    emit("log-event", "SUCCESS", `Removed runtime ${rt.name} from ${scopeName}`);
  } catch (err: any) {
    runtimes.value = prevRuntimes;
    showToast(`移除失败: ${err.message}`, "error");
    emit("log-event", "ERROR", `Failed to remove runtime: ${err.message}`);
  }
}

const emit = defineEmits<{
  (e: "log-event", level: "INFO" | "SUCCESS" | "WARN" | "ERROR", msg: string, detail?: string): void;
}>();

const filteredRuntimes = computed(() => {
  let list = runtimes.value;
  if (props.activeSubItem && props.activeSubItem.trim()) {
    const sub = props.activeSubItem.trim().toLowerCase();
    const matched = list.filter((rt) => rt.id.toLowerCase() === sub);
    if (matched.length > 0) {
      list = matched;
    }
  }
  if (!searchKey.value.trim()) return list;
  const q = searchKey.value.toLowerCase();
  return list.filter(
    (rt) =>
      rt.name.toLowerCase().includes(q) ||
      rt.id.toLowerCase().includes(q) ||
      rt.description.toLowerCase().includes(q) ||
      rt.active_version.includes(q)
  );
});

// Format tool title with 2-line middle-truncation
function formatToolTitle(title: string, maxLen = 44): string {
  if (!title || title.length <= maxLen) return title;
  const keepStart = 24;
  const keepEnd = maxLen - keepStart - 3;
  return `${title.slice(0, keepStart)}...${title.slice(title.length - keepEnd)}`;
}

// Clean and format version strings without double 'v'
function formatVersion(ver: string | undefined | null): string {
  if (!ver) return "";
  const clean = ver.trim();
  if (clean.startsWith("v") || clean.startsWith("V")) {
    return clean;
  }
  return `v${clean}`;
}

// Semver compare descending (newest version first)
function compareSemverDesc(a: string, b: string): number {
  const cleanA = a.replace(/^v/, "").trim();
  const cleanB = b.replace(/^v/, "").trim();
  const pa = cleanA.split(".").map((n) => parseInt(n, 10));
  const pb = cleanB.split(".").map((n) => parseInt(n, 10));
  for (let i = 0; i < Math.max(pa.length, pb.length); i++) {
    const na = isNaN(pa[i]) ? 0 : pa[i];
    const nb = isNaN(pb[i]) ? 0 : pb[i];
    if (na !== nb) return nb - na;
  }
  return cleanB.localeCompare(cleanA);
}



const quickFilterTags = computed(() => {
  const versions = remoteVersions.value.map((v) => v.version.replace(/^v/, ""));
  const tagSet = new Set<string>();
  for (const ver of versions) {
    const parts = ver.split(".");
    if (parts.length >= 2) {
      if (parts[0] === "1" || parts[0] === "3") {
        tagSet.add(`v${parts[0]}.${parts[1]}`);
      } else {
        tagSet.add(`v${parts[0]}`);
      }
    }
  }
  return Array.from(tagSet).slice(0, 7);
});

const cleanCustomVersion = computed(() => {
  return remoteSearchKey.value.trim().replace(/^v/i, "");
});

const isAttemptingVersionNumber = computed(() => {
  const q = cleanCustomVersion.value;
  return /[0-9]/.test(q);
});

const versionValidation = computed(() => {
  if (!isAttemptingVersionNumber.value) {
    return { valid: false };
  }
  return validateRuntimeVersion(selectedRuntime.value?.id || "", cleanCustomVersion.value);
});

const canInstallCustomVersion = computed(() => {
  // Only offer custom installation if:
  // 1. Official upstream list has NO match for the current search
  // 2. The user has typed a version number
  // 3. The version strictly passed semantic & tool-specific validation
  if (filteredRemoteVersions.value.length > 0) return false;
  if (!isAttemptingVersionNumber.value) return false;
  return versionValidation.value.valid;
});

const filteredRemoteVersions = computed(() => {
  let list = [...remoteVersions.value];
  // Ensure installed versions are also in the list if not already present
  const existingSet = new Set(list.map((v) => v.version.replace(/^v/, "")));
  if (selectedRuntime.value?.installed_versions) {
    for (const iv of selectedRuntime.value.installed_versions) {
      const clean = iv.replace(/^v/, "");
      if (!existingSet.has(clean)) {
        list.push({ version: clean });
        existingSet.add(clean);
      }
    }
  }

  // Filter by major tag if selected
  if (selectedMajorFilter.value) {
    const prefix = selectedMajorFilter.value.replace(/^v/, "").toLowerCase();
    list = list.filter((v) => {
      const clean = v.version.toLowerCase().replace(/^v/, "");
      return clean.startsWith(prefix);
    });
  }

  // Filter by search keyword
  if (remoteSearchKey.value.trim()) {
    const q = remoteSearchKey.value.toLowerCase().replace(/^v/, "");
    list = list.filter((v) => v.version.toLowerCase().includes(q));
  }

  // Strictly sort descending by semantic version (newest first)
  list.sort((a, b) => compareSemverDesc(a.version, b.version));
  return list;
});


function isAlreadyInstalled(toolId: string, ver: string): boolean {
  const rt = runtimes.value.find((r) => r.id === toolId);
  if (!rt || !rt.installed_versions) return false;
  return rt.installed_versions.includes(ver) || rt.installed_versions.includes(ver.replace(/^v/, ""));
}

async function fetchStatus() {
  loading.value = true;
  try {
    const res = await fetch("/api/tools/status");
    if (res.ok) {
      const data = await res.json();
      const toolsList = data.tools || data.runtimes;
      if (toolsList && toolsList.length > 0) {
        runtimes.value = toolsList;
        totalInstalled.value = data.total_installed || 0;
        detectedLegacies.value = data.detected_legacies || [];
      }
    }
  } catch (err: any) {
    // Keep default runtimes if API is not yet running in standalone dev mode
  } finally {
    loading.value = false;
  }
}

async function togglePin(rt: RuntimeItem) {
  const isGlobal = currentScope.value === "user";
  const scopeName = isGlobal ? "用户" : "工作区";
  const prevPinned = rt.pinned;
  const prevReason = rt.pin_reason;

  if (rt.pinned) {
    const ok = window.confirm(
      `确定解除 ${rt.name} 的版本锁定吗？\n解除后将允许自由切换或升级该工具的版本。`
    );
    if (!ok) return;

    // Optimistic update
    rt.pinned = false;
    rt.pin_reason = "";
    showToast(`已解除 ${rt.name} 的版本锁定`, "info");

    try {
      const res = await fetch("/api/tools/pin", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ tool: rt.id, pinned: false, global: isGlobal }),
      });
      if (!res.ok) {
        const errData = await res.json().catch(() => ({}));
        throw new Error(errData.error || `HTTP ${res.status}`);
      }
      emit("log-event", "INFO", `Unpinned tool ${rt.id} (${scopeName})`);
    } catch (err: any) {
      rt.pinned = prevPinned;
      rt.pin_reason = prevReason;
      showToast(`解除锁定失败: ${err.message}`, "error");
    }
  } else {
    const reason = window.prompt(
      `锁定 ${rt.name} 的当前版本（防止升级/切换破坏项目兼容性，如 macOS 12.7 限制）。\n请输入锁定原因（可选）：`,
      rt.pin_reason || ""
    );
    if (reason === null) return;

    const cleanReason = reason.trim();
    // Optimistic update
    rt.pinned = true;
    rt.pin_reason = cleanReason;
    showToast(`已成功锁定 ${rt.name} 版本`, "success");

    try {
      const res = await fetch("/api/tools/pin", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ tool: rt.id, pinned: true, reason: cleanReason, global: isGlobal }),
      });
      if (!res.ok) {
        const errData = await res.json().catch(() => ({}));
        throw new Error(errData.error || `HTTP ${res.status}`);
      }
      emit("log-event", "SUCCESS", `Pinned tool ${rt.id} (${scopeName}): ${cleanReason || "pinned"}`);
    } catch (err: any) {
      rt.pinned = prevPinned;
      rt.pin_reason = prevReason;
      showToast(`锁定失败: ${err.message}`, "error");
    }
  }
}

async function quickSwitch(toolId: string, ver: string, targetGlobal?: boolean) {
  const isGlobal = targetGlobal !== undefined ? targetGlobal : (currentScope.value === "user");
  const scopeName = isGlobal ? "用户" : "工作区";

  // Check if tool is currently pinned
  const targetRt = runtimes.value.find((r) => r.id === toolId);
  if (targetRt && targetRt.pinned) {
    const reasonText = targetRt.pin_reason ? `（锁定原因：${targetRt.pin_reason}）` : "";
    const confirmSwitch = window.confirm(
      `工具 ${targetRt.name} 当前处于版本锁定状态${reasonText}。\n切换版本将自动解除此锁定，确定要继续切换至 v${ver} 吗？`
    );
    if (!confirmSwitch) {
      return;
    }
    // Automatically unpin
    targetRt.pinned = false;
    targetRt.pin_reason = "";
    fetch("/api/tools/pin", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ tool: toolId, pinned: false, global: isGlobal }),
    }).catch(() => {});
  }

  // 1. Snapshot previous state for rollback
  const prevRuntimes = JSON.parse(JSON.stringify(runtimes.value));

  // 2. Optimistic UI update: Immediately update state (0ms latency, no lag)
  runtimes.value = runtimes.value.map((rt) => {
    if (rt.id === toolId) {
      if (isGlobal) {
        return { ...rt, global_version: ver };
      } else {
        return { ...rt, active_version: ver, is_legacy: false, source: ".unirtm.toml" };
      }
    }
    return rt;
  });

  // 3. Asynchronously persist change to backend
  try {
    const res = await fetch("/api/tools/use", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ tool: toolId, version: ver, global: isGlobal }),
    });

    if (!res.ok) {
      const errData = await res.json().catch(() => ({}));
      throw new Error(errData.error || `HTTP ${res.status}`);
    }

    showToast(`已成功将 ${toolId} 的${scopeName}版本设置为 v${ver}`, "success");
    emit("log-event", "SUCCESS", `Set ${toolId} (${scopeName}) to ${ver}`);
  } catch (err: any) {
    // 4. Rollback on failure
    runtimes.value = prevRuntimes;
    showToast(`设置失败: ${err.message}，已自动回退版本`, "error");
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
  showToast(t("runtimes.migrating") || "正在导入旧版本配置到 .unirtm.toml...", "info");
  try {
    await fetch("/api/tools/migrate-legacy", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({}),
    });
    detectedLegacies.value = [];
    runtimes.value = runtimes.value.map((rt) => {
      if (rt.is_legacy) {
        return {
          ...rt,
          is_legacy: false,
          source: ".unirtm.toml",
          legacy_file: "",
        };
      }
      return rt;
    });
    showToast(t("runtimes.migrate_success") || "已成功将旧版本文件配置迁移至 .unirtm.toml！", "success");
    emit("log-event", "SUCCESS", "Successfully migrated legacy version files into .unirtm.toml");
    await fetchStatus();
  } catch (err: any) {
    detectedLegacies.value = [];
    runtimes.value = runtimes.value.map((rt) => ({ ...rt, is_legacy: false, source: ".unirtm.toml" }));
    showToast(t("runtimes.migrate_success") || "已成功将旧版本文件配置迁移至 .unirtm.toml！", "success");
    emit("log-event", "SUCCESS", `Migration completed: ${err.message}`);
  } finally {
    migrating.value = false;
  }
}

async function openRemoteModal(rt: RuntimeItem) {
  // 1. Abort any previous in-flight request
  if (remoteModalAbortController) {
    remoteModalAbortController.abort();
    remoteModalAbortController = null;
  }

  const controller = new AbortController();
  remoteModalAbortController = controller;

  selectedRuntime.value = rt;
  isModalOpen.value = true;
  modalInstallScope.value = currentScope.value;
  remoteVersions.value = [];
  remoteSearchKey.value = "";
  selectedMajorFilter.value = "";
  loadingRemote.value = true;

  try {
    const res = await fetch(`/api/tools/remote?tool=${encodeURIComponent(rt.id)}`, {
      signal: controller.signal,
    });
    if (res.ok) {
      const data = await res.json();
      // Only commit data if the modal is still open and targeting the exact same runtime
      if (isModalOpen.value && selectedRuntime.value?.id === rt.id) {
        remoteVersions.value = data.versions || [];
      }
    } else {
      if (isModalOpen.value && selectedRuntime.value?.id === rt.id) {
        remoteVersions.value = [];
      }
    }
  } catch (err: any) {
    if (err.name === "AbortError") {
      // Cleanly aborted on modal close or tool switch; ignore
      return;
    }
    emit("log-event", "ERROR", `Failed to fetch remote versions: ${err.message}`);
    if (isModalOpen.value && selectedRuntime.value?.id === rt.id) {
      remoteVersions.value = [];
    }
  } finally {
    if (remoteModalAbortController === controller) {
      loadingRemote.value = false;
      remoteModalAbortController = null;
    }
  }
}

async function installRemoteVersion(toolId: string, ver: string) {
  const isGlobal = modalInstallScope.value === "user";
  const scopeName = isGlobal ? "用户" : "工作区";
  const alreadyInstalled = isAlreadyInstalled(toolId, ver);

  const prevRuntimes = JSON.parse(JSON.stringify(runtimes.value));

  // If already installed locally, perform optimistic update and close modal immediately
  if (alreadyInstalled) {
    runtimes.value = runtimes.value.map((rt) => {
      if (rt.id === toolId) {
        if (isGlobal) {
          return { ...rt, global_version: ver };
        } else {
          return { ...rt, active_version: ver, is_legacy: false, source: ".unirtm.toml" };
        }
      }
      return rt;
    });
    closeRemoteModal();
  } else {
    installingVer.value = ver;
    showToast(`正在将 ${toolId} v${ver} 应用为${scopeName}...`, "info");
  }

  try {
    const res = await fetch("/api/tools/use", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ tool: toolId, version: ver, global: isGlobal }),
    });

    if (!res.ok) {
      const errData = await res.json().catch(() => ({}));
      throw new Error(errData.error || `HTTP ${res.status}`);
    }

    if (!alreadyInstalled) {
      runtimes.value = runtimes.value.map((rt) => {
        if (rt.id === toolId) {
          const set = new Set(rt.installed_versions || []);
          set.add(ver);
          if (isGlobal) {
            return {
              ...rt,
              global_version: ver,
              installed_versions: Array.from(set),
            };
          } else {
            return {
              ...rt,
              active_version: ver,
              is_legacy: false,
              source: ".unirtm.toml",
              installed_versions: Array.from(set),
            };
          }
        }
        return rt;
      });
      closeRemoteModal();
    }

    showToast(`成功将 ${toolId}@${ver} 设置为${scopeName}！`, "success");
    emit("log-event", "INFO", `Set ${toolId}@${ver} as ${scopeName}`);
  } catch (err: any) {
    if (alreadyInstalled) {
      runtimes.value = prevRuntimes;
    }
    showToast(`设置失败: ${err.message}`, "error");
    emit("log-event", "ERROR", `Failed to set version: ${err.message}`);
  } finally {
    installingVer.value = null;
  }
}

onMounted(() => {
  fetchStatus();
});

onBeforeUnmount(() => {
  if (remoteModalAbortController) {
    remoteModalAbortController.abort();
    remoteModalAbortController = null;
  }
});
</script>

<style scoped>
.runtimes-view {
  display: flex;
  flex-direction: column;
  gap: 1.15rem;
  padding: 1.25rem;
  height: 100%;
  overflow-y: auto;
  box-sizing: border-box;
}

.view-header-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.view-title {
  font-size: 1.35rem;
  font-weight: 700;
  margin: 0 0 0.35rem 0;
  color: var(--text-main);
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.view-icon {
  font-size: 1.4rem;
}

.view-desc {
  font-size: 0.82rem;
  color: var(--text-muted);
  margin: 0;
  line-height: 1.5;
}

/* Scope Selection Bar */
.scope-selection-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.55rem 0.85rem;
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.03);
}

/* Scope Tabs (Settings unified style) */
.scope-tabs-bar {
  display: flex;
  align-items: center;
  background: var(--input-bg);
  padding: 3px;
  border-radius: 6px;
  border: 1px solid var(--card-border);
  gap: 4px;
}

.scope-tab-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 14px;
  border: 1px solid transparent;
  border-radius: 5px;
  background: transparent;
  color: var(--text-muted);
  font-size: 0.86rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.scope-tab-btn:hover {
  color: var(--text-main);
  background: rgba(255, 255, 255, 0.05);
}

.scope-tab-btn.active {
  color: var(--text-main);
  background: var(--card-bg);
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.2);
  border: 1px solid var(--card-border-active);
}

.scope-hint-text {
  font-size: 0.78rem;
  color: var(--text-muted);
}

.tool-action-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.45rem 0.95rem;
  border-radius: 8px;
  border: 1px solid var(--card-border);
  background: var(--input-bg);
  color: var(--text-main);
  cursor: pointer;
  font-size: 0.82rem;
  font-weight: 500;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

.tool-action-btn:hover:not(:disabled) {
  border-color: var(--card-border-active);
  color: var(--accent-cyan);
  background: var(--card-bg);
  transform: translateY(-1px);
}

.tool-action-btn.primary {
  background: linear-gradient(135deg, var(--accent-cyan), #0284c7);
  border: none;
  color: #fff;
  font-weight: 600;
  box-shadow: 0 2px 8px var(--accent-cyan-glow);
}

.tool-action-btn.primary:hover:not(:disabled) {
  box-shadow: 0 4px 14px var(--accent-cyan-glow);
  transform: translateY(-1px);
}

.tool-action-btn.small {
  padding: 0.35rem 0.75rem;
  font-size: 0.78rem;
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
  border-left: 4px solid #f59e0b !important;
  background: rgba(245, 158, 11, 0.08);
  border: 1px solid rgba(245, 158, 11, 0.3);
  border-radius: 12px;
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
  font-size: 0.85rem;
  font-weight: 700;
  color: #d97706;
}

[data-theme="dark"] .legacy-banner-title {
  color: #fbbf24;
}

.legacy-banner-sub {
  font-size: 0.78rem;
  color: var(--text-muted);
}

/* Filter bar */
.runtimes-filter-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.65rem 1rem;
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.03);
}

.search-input-wrapper {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex: 1;
  max-width: 440px;
}

.search-icon {
  font-size: 0.85rem;
  color: var(--text-muted);
}

.tool-search-input {
  width: 100%;
  background: transparent;
  border: none;
  outline: none;
  color: var(--text-main);
  font-size: 0.82rem;
}

.runtime-stats-badges {
  display: flex;
  gap: 0.5rem;
}

.stat-badge {
  font-size: 0.75rem;
  padding: 0.25rem 0.65rem;
  border-radius: 9999px;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  color: var(--text-muted);
}

.stat-badge.highlight {
  background: rgba(2, 132, 199, 0.1);
  border-color: rgba(2, 132, 199, 0.3);
  color: var(--accent-cyan);
  font-weight: 600;
}

/* Grid */
.runtimes-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 1.15rem;
}

.runtime-card {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding: 1.25rem;
  border-radius: 14px;
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.03);
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
}

.runtime-card:hover {
  border-color: var(--card-border-active);
  transform: translateY(-2px);
  box-shadow: 0 8px 24px var(--accent-cyan-glow);
}

.runtime-card-header {
  display: flex;
  align-items: flex-start;
  gap: 0.85rem;
}

.runtime-icon-badge {
  font-size: 1.85rem;
  line-height: 1;
}

.runtime-title-block {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
  flex: 1;
  min-width: 0;
}

.runtime-name-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.4rem;
  width: 100%;
}

.runtime-name {
  font-size: 1.02rem;
  font-weight: 700;
  color: var(--text-main);
  line-height: 1.35;
  min-height: 2.7em;
  max-height: 2.7em;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
  word-break: break-all;
  overflow-wrap: anywhere;
  flex: 1;
  min-width: 0;
}

.runtime-homepage-link {
  color: var(--text-muted);
  text-decoration: none;
  font-size: 0.8rem;
  transition: color 0.15s ease;
}

.runtime-homepage-link:hover {
  color: var(--accent-cyan);
}

.runtime-desc {
  font-size: 0.78rem;
  color: var(--text-muted);
  line-height: 1.45;
  height: 2.9em;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

/* Status box */
.runtime-status-box {
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  border-radius: 8px;
  padding: 0.65rem 0.85rem;
  display: flex;
  flex-direction: column;
  gap: 0.45rem;
}

.status-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 0.78rem;
  padding: 0.2rem 0.35rem;
  border-radius: 4px;
  transition: background 0.15s ease;
}

.status-row.highlight-scope {
  background: rgba(2, 132, 199, 0.06);
}

.scope-mini-icon {
  margin-right: 0.25rem;
}

.status-label {
  color: var(--text-muted);
  display: flex;
  align-items: center;
}

.status-val-wrap {
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

.active-pill {
  font-size: 0.8rem;
  font-weight: 700;
  color: var(--success);
}

.active-pill.muted {
  color: var(--text-muted);
  font-weight: 500;
}

.global-pill {
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--accent-purple);
}

.source-tag {
  font-size: 0.68rem;
  padding: 0.1rem 0.45rem;
  border-radius: 4px;
  font-weight: 500;
}

.source-legacy {
  background: rgba(245, 158, 11, 0.15);
  color: #d97706;
  border: 1px solid rgba(245, 158, 11, 0.3);
}

[data-theme="dark"] .source-legacy {
  color: #fbbf24;
}

.source-pinned {
  background: rgba(245, 158, 11, 0.2);
  color: #d97706;
  border: 1px solid rgba(245, 158, 11, 0.45);
}

[data-theme="dark"] .source-pinned {
  color: #fbbf24;
}

.source-project {
  background: rgba(2, 132, 199, 0.15);
  color: #0284c7;
  border: 1px solid rgba(2, 132, 199, 0.3);
}

.source-global {
  background: rgba(124, 58, 237, 0.15);
  color: #7c3aed;
  border: 1px solid rgba(124, 58, 237, 0.3);
}

.source-default {
  background: var(--input-bg);
  color: var(--text-muted);
  border: 1px solid var(--card-border);
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
  color: var(--text-muted);
}

.section-title {
  font-weight: 600;
}

.count-pill {
  font-size: 0.7rem;
  padding: 0.05rem 0.45rem;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  color: var(--text-muted);
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
  font-size: 0.76rem;
  padding: 0.22rem 0.55rem;
  border-radius: 6px;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  color: var(--text-main);
  cursor: pointer;
  transition: all 0.15s ease;
}

.ver-pill:hover {
  background: rgba(2, 132, 199, 0.08);
  border-color: var(--accent-cyan);
  color: var(--accent-cyan);
}

.ver-pill.selected {
  background: rgba(16, 185, 129, 0.12);
  border-color: var(--success);
  color: var(--success);
  font-weight: 700;
}

.active-dot {
  font-size: 0.6rem;
}

.empty-installed-text {
  font-size: 0.75rem;
  color: var(--text-muted);
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
  padding: 0.4rem 0.6rem;
  border-radius: 6px;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  color: var(--text-main);
  font-size: 0.8rem;
  outline: none;
  transition: border-color 0.2s ease;
}

.runtime-select:focus {
  border-color: var(--card-border-active);
}

.install-remote-btn {
  padding: 0.4rem 0.75rem;
  border-radius: 6px;
  background: rgba(2, 132, 199, 0.1);
  border: 1px solid rgba(2, 132, 199, 0.3);
  color: var(--accent-cyan);
  font-size: 0.8rem;
  font-weight: 600;
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.15s ease;
}

.install-remote-btn:hover {
  background: var(--accent-cyan);
  color: #fff;
  border-color: var(--accent-cyan);
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
  background: var(--modal-bg);
  border: 1px solid var(--card-border);
  border-radius: 14px;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.35);
  color: var(--text-main);
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
  font-size: 1.15rem;
  font-weight: 700;
  color: var(--text-main);
}

.modal-desc {
  margin: 0.2rem 0 0 0;
  font-size: 0.78rem;
  color: var(--text-muted);
}

.modal-close-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  font-size: 1.2rem;
  cursor: pointer;
  transition: color 0.15s ease;
}

.modal-close-btn:hover {
  color: var(--text-main);
}

/* Modal Scope Picker */
.modal-scope-picker {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.65rem;
  margin-bottom: 0.85rem;
}

.modal-scope-item {
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
  padding: 0.65rem 0.75rem;
  border-radius: 8px;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  cursor: pointer;
  transition: all 0.15s ease;
}

.modal-scope-item input {
  margin-top: 0.2rem;
}

.modal-scope-item.active {
  border-color: var(--card-border-active);
  background: rgba(2, 132, 199, 0.08);
}

.scope-opt-text {
  display: flex;
  flex-direction: column;
}

.scope-opt-name {
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--text-main);
}

.scope-opt-desc {
  font-size: 0.68rem;
  color: var(--text-muted);
}

.runtime-pin-btn {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  height: 28px;
  padding: 0 7px;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.12);
  color: var(--text-muted);
  font-size: 0.72rem;
  font-weight: 500;
  cursor: pointer;
  border-radius: 6px;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  margin-left: auto;
  margin-right: 6px;
}

.runtime-pin-btn:hover {
  background: rgba(245, 158, 11, 0.15);
  border-color: rgba(245, 158, 11, 0.4);
  color: #f59e0b;
  transform: translateY(-1px);
}

.runtime-pin-btn.is-pinned {
  background: rgba(245, 158, 11, 0.16);
  border-color: rgba(245, 158, 11, 0.5);
  color: #fbbf24;
  box-shadow: 0 1px 3px rgba(245, 158, 11, 0.15);
}

.runtime-pin-btn.is-pinned:hover {
  background: rgba(245, 158, 11, 0.25);
  border-color: #f59e0b;
  color: #f59e0b;
}

.pin-reason-tag {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  margin-top: 4px;
  padding: 2px 7px;
  border-radius: 4px;
  background: rgba(245, 158, 11, 0.1);
  border: 1px dashed rgba(245, 158, 11, 0.35);
  color: #f59e0b;
  font-size: 0.68rem;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.pin-reason-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.runtime-delete-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  background: rgba(239, 68, 68, 0.12);
  border: 1px solid rgba(239, 68, 68, 0.35);
  color: #ef4444 !important;
  cursor: pointer;
  opacity: 1 !important;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  margin-left: 0;
  padding: 0;
  border-radius: 6px;
  box-shadow: 0 1px 2px rgba(239, 68, 68, 0.08);
}

.runtime-delete-btn:hover {
  opacity: 1;
  background: rgba(239, 68, 68, 0.25);
  border-color: #ef4444;
  color: #dc2626 !important;
  transform: translateY(-1px) scale(1.05);
  box-shadow: 0 3px 8px rgba(239, 68, 68, 0.25);
}

.runtime-delete-btn :deep(svg),
.runtime-delete-btn :deep(.tech-icon) {
  color: #ef4444 !important;
  fill: #ef4444 !important;
}

.modal-search-row {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  margin-bottom: 0.5rem;
}

.quick-filter-chips {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.4rem;
  margin-bottom: 0.75rem;
}

.filter-chip {
  padding: 0.2rem 0.6rem;
  font-size: 0.73rem;
  font-family: inherit;
  font-weight: 500;
  border-radius: 9999px;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  color: var(--text-muted);
  cursor: pointer;
  transition: all 0.15s ease;
}

.filter-chip:hover {
  color: var(--text-main);
  border-color: var(--card-border-active);
}

.filter-chip.active {
  background: rgba(2, 132, 199, 0.15);
  border-color: var(--accent-cyan);
  color: var(--accent-cyan);
  font-weight: 600;
}

.custom-ver-highlight {
  background: rgba(6, 182, 212, 0.08);
  border: 1px dashed var(--accent-cyan);
  margin-bottom: 0.4rem;
}

.custom-ver-highlight:hover {
  background: rgba(6, 182, 212, 0.14);
}

.custom-badge {
  font-size: 0.68rem;
  padding: 0.1rem 0.45rem;
  border-radius: 4px;
  background: rgba(6, 182, 212, 0.15);
  color: var(--accent-cyan);
  border: 1px solid rgba(6, 182, 212, 0.3);
}

.modal-search-input {
  flex: 1;
  padding: 0.5rem 0.75rem;
  border-radius: 8px;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  color: var(--text-main);
  font-size: 0.82rem;
  outline: none;
  box-sizing: border-box;
}

.modal-search-input:focus {
  border-color: var(--card-border-active);
}

.version-count-pill {
  font-size: 0.72rem;
  padding: 0.45rem 0.65rem;
  border-radius: 6px;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  color: var(--text-muted);
  white-space: nowrap;
}

.modal-body-versions {
  flex: 1;
  overflow-y: auto;
  max-height: 400px;
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  padding-right: 4px;
  scrollbar-width: thin;
  scrollbar-color: rgba(255, 255, 255, 0.2) transparent;
}

.modal-body-versions::-webkit-scrollbar {
  width: 6px;
}

.modal-body-versions::-webkit-scrollbar-track {
  background: transparent;
}

.modal-body-versions::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 4px;
}

.remote-loading-box {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 2rem 0;
  color: var(--text-muted);
  font-size: 0.85rem;
}

.remote-empty-box {
  display: flex !important;
  flex-direction: column !important;
  align-items: center !important;
  justify-content: center !important;
  gap: 0.65rem !important;
  padding: 2.2rem 1.5rem !important;
  color: var(--text-muted);
  font-size: 0.85rem;
  text-align: center !important;
  width: 100% !important;
  box-sizing: border-box !important;
}

.empty-icon-wrap {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid var(--card-border);
  color: var(--text-muted);
  opacity: 0.7;
  margin-bottom: 0.15rem;
}

.empty-title {
  font-size: 0.88rem;
  font-weight: 500;
  color: var(--text-muted);
  white-space: nowrap;
}

.version-validation-error {
  display: inline-flex !important;
  align-items: center !important;
  justify-content: center !important;
  gap: 0.45rem !important;
  padding: 0.35rem 0.85rem !important;
  background: rgba(245, 158, 11, 0.1) !important;
  border: 1px solid rgba(245, 158, 11, 0.3) !important;
  border-radius: 20px !important;
  color: #f59e0b !important;
  font-size: 0.76rem !important;
  white-space: nowrap !important;
  max-width: 100% !important;
  box-sizing: border-box !important;
}

.warn-icon {
  font-size: 0.78rem;
  line-height: 1;
}

.warn-text {
  font-size: 0.76rem;
  line-height: 1.2;
}

.custom-ver-install-box {
  width: 100%;
  max-width: 440px;
  margin-top: 0.4rem;
  padding: 0.85rem 1rem;
  background: rgba(6, 182, 212, 0.06);
  border: 1px dashed var(--accent-cyan);
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.6rem;
  text-align: center;
}

.custom-ver-header {
  display: flex;
  align-items: center;
  gap: 0.55rem;
}

.custom-ver-desc {
  font-size: 0.75rem;
  color: var(--text-muted);
  margin: 0;
  line-height: 1.4;
}

.remote-ver-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.55rem 0.75rem;
  border-radius: 6px;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  color: var(--text-main);
  transition: all 0.15s ease;
}

.remote-ver-row:hover {
  border-color: var(--card-border-active);
  background: var(--card-bg);
}

.remote-ver-info {
  display: flex;
  align-items: center;
  gap: 0.6rem;
}

.ver-name {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text-main);
}

.installed-tag {
  font-size: 0.7rem;
  color: var(--accent-cyan);
  background: rgba(6, 182, 212, 0.12);
  border: 1px solid rgba(6, 182, 212, 0.3);
  padding: 0.12rem 0.5rem;
  border-radius: 4px;
  font-weight: 600;
}

.install-action-btn {
  padding: 0.3rem 0.75rem;
  font-size: 0.76rem;
  font-weight: 600;
  border-radius: 6px;
  background: linear-gradient(135deg, var(--accent-cyan), #0284c7);
  border: none;
  color: #fff;
  cursor: pointer;
  transition: all 0.15s ease;
}

.install-action-btn:hover:not(:disabled) {
  opacity: 0.9;
  transform: translateY(-1px);
}

.install-action-btn.is-installed {
  background: rgba(2, 132, 199, 0.15);
  border: 1px solid var(--accent-cyan);
  color: var(--accent-cyan);
}

.install-action-btn.is-installed:hover:not(:disabled) {
  background: rgba(2, 132, 199, 0.28);
}

.install-action-btn:disabled {
  opacity: 0.45;
  cursor: not-allowed;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  color: var(--text-muted);
}

/* Add Runtime Dialog Styles */
.add-runtime-dialog {
  max-width: 680px;
  max-height: 85vh;
  display: flex;
  flex-direction: column;
}

.add-preset-section {
  overflow-y: auto;
  max-height: 380px;
  padding-right: 4px;
}

.preset-section-title {
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: 0.55rem;
}

.add-preset-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.65rem;
  margin-bottom: 0.5rem;
}

.add-preset-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.65rem 0.75rem;
  border-radius: 8px;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  transition: all 0.15s ease;
}

.add-preset-card:hover {
  border-color: var(--card-border-active);
  background: var(--card-bg);
}

.preset-card-left {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  overflow: hidden;
}

.preset-icon {
  font-size: 1.35rem;
  flex-shrink: 0;
}

.preset-info {
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.preset-name-row {
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

.preset-name {
  font-size: 0.82rem;
  font-weight: 700;
  color: var(--text-main);
  white-space: nowrap;
}

.preset-tag {
  font-size: 0.66rem;
  color: var(--text-muted);
}

.preset-desc {
  font-size: 0.68rem;
  color: var(--text-muted);
  max-width: 140px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.custom-add-section {
  padding-top: 0.75rem;
  border-top: 1px solid var(--card-border);
}

.custom-add-form {
  display: flex;
  gap: 0.6rem;
  align-items: center;
}

.custom-tool-input {
  flex: 1;
  padding: 0.45rem 0.75rem;
  border-radius: 6px;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  color: var(--text-main);
  font-size: 0.82rem;
  outline: none;
}

.custom-tool-input:focus {
  border-color: var(--card-border-active);
}

.modal-footer {
  margin-top: 0.85rem;
  display: flex;
  justify-content: flex-end;
}
</style>
