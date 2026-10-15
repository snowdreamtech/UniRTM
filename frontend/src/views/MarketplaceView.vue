<template>
  <div class="view-container marketplace-view" id="view-marketplace">
    <!-- View Header -->
    <div class="view-header-row">
      <div>
        <h2 class="view-title">{{ t("nav.marketplace") || "扩展与主题市场" }}</h2>
        <p class="view-desc">
          探索与安装 VS Code 兼容的主题与图标集，以及扩展 UniRTM 原生能力的开发者功能套件。
        </p>
      </div>

      <!-- Channel Switcher Tabs -->
      <div class="channel-tabs glass-card">
        <button
          class="channel-btn"
          :class="{ active: currentChannel === 'themes' }"
          @click="switchChannel('themes')"
        >
          <span>🎨</span>
          <span>{{ t("settings.theme_tab_marketplace") || "主题与图标 (Open VSX)" }}</span>
        </button>
        <button
          class="channel-btn"
          :class="{ active: currentChannel === 'plugins' }"
          @click="switchChannel('plugins')"
        >
          <span>🧩</span>
          <span>{{ t("settings.cat_plugins") || "功能插件 (UniRTM 生态)" }}</span>
        </button>
      </div>
    </div>

    <!-- CHANNEL 1: Themes & Icons from Open VSX -->
    <div v-show="currentChannel === 'themes'" class="market-channel-content">
      <!-- Search Filter Bar -->
      <div class="market-filter-bar glass-card">
        <div class="search-input-wrapper">
          <span class="search-icon">🔍</span>
          <input
            v-model="mpSearchQuery"
            type="text"
            class="market-search-input"
            :placeholder="t('settings.marketplace_search_placeholder') || '在 Open VSX 搜索主题或图标集 (如 Dracula, Material, Catppuccin)...'"
            @keydown.enter="onSearchThemes(mpSearchQuery)"
          />
          <button
            v-if="mpSearchQuery"
            class="btn-clear-search"
            title="清空"
            @click="mpSearchQuery = ''; onSearchThemes('')"
          >
            ✕
          </button>
        </div>
        <button
          class="btn-search-action"
          :disabled="mpIsSearching"
          @click="onSearchThemes(mpSearchQuery)"
        >
          <span v-if="mpIsSearching" class="spinner-small"></span>
          <span v-else>搜索</span>
        </button>
      </div>

      <!-- Quick Chips & Results count -->
      <div class="market-chips-bar">
        <div class="chips-left">
          <span class="chips-label">推荐分类:</span>
          <button
            v-for="chip in themeChips"
            :key="chip"
            class="chip-btn"
            :class="{ active: (chip === '全部' && !mpSearchQuery) || mpSearchQuery.toLowerCase() === chip.toLowerCase() }"
            @click="onSelectChip(chip)"
          >
            {{ chip }}
          </button>
        </div>
        <span v-if="mpTotalSize > 0" class="results-count-label">
          共收录 {{ mpTotalSize }} 款外观扩展
        </span>
      </div>

      <!-- Feedback Alert -->
      <div v-if="feedbackMsg" class="feedback-banner" :class="{ error: feedbackIsError }">
        <span>{{ feedbackMsg }}</span>
        <button class="btn-close-banner" @click="feedbackMsg = ''">✕</button>
      </div>

      <!-- Error Banner -->
      <div v-if="mpError" class="error-banner">
        <span>⚠️ {{ mpError }}</span>
        <button class="btn-retry" @click="onSearchThemes(mpSearchQuery)">重试</button>
      </div>

      <!-- Loading State -->
      <div v-if="mpIsSearching" class="market-loading-box glass-card">
        <div class="spinner-small"></div>
        <span>正在实时连接 Open VSX 检索主题与图标...</span>
      </div>

      <!-- Empty State -->
      <div v-else-if="mpSearchResults.length === 0" class="market-empty-box glass-card">
        <span class="empty-icon">🎨</span>
        <p>未找到符合条件的主题或图标包</p>
        <button class="btn-reset-chip" @click="onSelectChip('全部')">查看热门推荐</button>
      </div>

      <!-- Themes Grid -->
      <div v-else class="extensions-grid">
        <div
          v-for="item in mpSearchResults"
          :key="item.namespace + '.' + item.name"
          class="glass-card theme-card"
        >
          <div class="card-header-row">
            <div class="card-icon-wrap">
              <img
                v-if="item.files && item.files.icon"
                :src="item.files.icon"
                class="card-img-icon"
                loading="lazy"
                alt="icon"
                @error="(e: any) => { if (e && e.target) e.target.style.display = 'none'; }"
              />
              <span class="card-fallback-icon">🎨</span>
            </div>
            <div class="card-title-col">
              <div class="card-name-row">
                <span class="card-name" :title="item.displayName || item.name">
                  {{ item.displayName || item.name }}
                </span>
                <span class="card-version font-mono">v{{ item.version }}</span>
              </div>
              <span class="card-namespace font-mono">{{ item.namespace }}</span>
            </div>
          </div>

          <p class="card-desc" :title="item.description">
            {{ item.description || "无详细描述" }}
          </p>

          <div class="card-footer-row">
            <div class="card-stats">
              <span class="stat-item" :title="'下载量: ' + item.downloadCount">
                ⬇️ {{ formatDownloads(item.downloadCount) }}
              </span>
              <span v-if="item.averageRating" class="stat-item">
                ⭐ {{ item.averageRating.toFixed(1) }}
              </span>
            </div>

            <div class="card-actions">
              <div v-if="item.is_installed" class="installed-group">
                <span class="badge-installed">✓ {{ t("settings.installed") || "已安装" }}</span>
                <button
                  class="btn-apply-theme"
                  :title="t('settings.apply_theme') || '应用此主题'"
                  @click="onApplyTheme(item)"
                >
                  {{ t("settings.apply") || "应用" }}
                </button>
              </div>
              <button
                v-else
                class="btn-install-theme"
                :disabled="mpInstallingMap[item.namespace + '.' + item.name]"
                @click="onInstallTheme(item)"
              >
                <span v-if="mpInstallingMap[item.namespace + '.' + item.name]" class="spinner-small"></span>
                <span>{{ mpInstallingMap[item.namespace + '.' + item.name] ? (t("settings.installing") || "安装中...") : (t("settings.install") || "安装") }}</span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- CHANNEL 2: Native Plugins & Ecosystem (UniRTM Extensions) -->
    <div v-show="currentChannel === 'plugins'" class="market-channel-content">
      <!-- Ecosystem Banner -->
      <div class="plugin-isolation-banner glass-card">
        <span class="banner-icon">🔒</span>
        <div class="banner-text">
          <strong>三层隔离架构与工具提供者准则</strong>
          <p>
            UniRTM 原生功能插件运行于独立外部进程中，绝不臃肿核心运行时。所有插件配置与存储严格限定于 plugins.&lt;plugin-id&gt;.* 命名空间内。
          </p>
        </div>
      </div>

      <!-- Native Plugins Grid -->
      <div class="extensions-grid">
        <div
          v-for="ext in nativePlugins"
          :key="ext.id"
          class="glass-card plugin-card"
        >
          <div class="card-header-row">
            <div class="card-icon-box">{{ ext.icon }}</div>
            <div class="card-title-col">
              <div class="card-name-row">
                <span class="card-name">{{ ext.name }}</span>
                <span class="card-version font-mono">v{{ ext.version }}</span>
              </div>
              <span class="card-namespace font-mono">{{ ext.author }}</span>
            </div>
          </div>

          <p class="card-desc">{{ ext.description }}</p>

          <div class="plugin-tags-row">
            <span v-for="tag in ext.tags" :key="tag" class="plugin-tag">{{ tag }}</span>
          </div>

          <div class="card-footer-row">
            <div class="card-stats">
              <span class="stat-item">⭐ {{ ext.stars }}</span>
              <span class="stat-item">📥 {{ ext.downloads }}</span>
            </div>

            <div class="card-actions">
              <span v-if="ext.installed" class="badge-installed">✓ 已启用</span>
              <button
                v-else
                class="btn-install-theme"
                @click="ext.installed = true; feedbackMsg = `✓ 已启用插件 ${ext.name}`"
              >
                {{ t("settings.install") || "安装" }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useRoute } from "vue-router";
import { t } from "../i18n";
import { useThemeMarketplace } from "../composables/useThemeMarketplace";
import { useThemeManager } from "../composables/useThemeManager";
import type { MarketplaceExtension } from "../types/theme-manager";

const route = useRoute();

// Channels: 'themes' (Open VSX) vs 'plugins' (UniRTM Native Plugins)
const currentChannel = ref<"themes" | "plugins">("themes");

// Initialize channel from query parameter if present
onMounted(() => {
  if (route.query.tab === "plugins") {
    currentChannel.value = "plugins";
  } else {
    currentChannel.value = "themes";
  }
  // Trigger initial theme marketplace search
  mpSearchThemes("");
});

function switchChannel(channel: "themes" | "plugins") {
  currentChannel.value = channel;
}

// --- Theme Marketplace Composable Integration ---
const {
  searchQuery: mpSearchQuery,
  searchResults: mpSearchResults,
  totalSize: mpTotalSize,
  isSearching: mpIsSearching,
  installingMap: mpInstallingMap,
  marketplaceError: mpError,
  searchThemes: mpSearchThemes,
  installThemeOnline: mpInstallThemeOnline,
} = useThemeMarketplace();

const {
  colorThemes,
  iconThemes,
  setActiveTheme,
} = useThemeManager();

const themeChips = [
  "全部",
  "Material",
  "Dracula",
  "One Dark",
  "Catppuccin",
  "Tokyo Night",
  "GitHub",
  "Icons",
];

const feedbackMsg = ref("");
const feedbackIsError = ref(false);

function onSearchThemes(q?: string) {
  mpSearchThemes(q);
}

function onSelectChip(chip: string) {
  const query = chip === "全部" ? "" : chip;
  mpSearchQuery.value = query;
  mpSearchThemes(query);
}

async function onInstallTheme(item: MarketplaceExtension) {
  try {
    feedbackMsg.value = "";
    feedbackIsError.value = false;
    const installed = await mpInstallThemeOnline(item);
    feedbackMsg.value = `✓ 成功安装 ${item.displayName || item.name}！`;
    if (installed && installed.length > 0) {
      const first = installed[0];
      if (first.type === "color") {
        await setActiveTheme(first.id, undefined);
      } else if (first.type === "icon") {
        await setActiveTheme(undefined, first.id);
      }
    }
  } catch (err: any) {
    feedbackIsError.value = true;
    feedbackMsg.value = `✕ 安装失败: ${err.message || err}`;
  }
}

async function onApplyTheme(item: MarketplaceExtension) {
  const pkgID = item.package_id || item.name;
  const colorMatch = colorThemes.value.find((c) => c.package_id === pkgID);
  if (colorMatch) {
    await setActiveTheme(colorMatch.id, undefined);
  }
  const iconMatch = iconThemes.value.find((i) => i.package_id === pkgID);
  if (iconMatch) {
    await setActiveTheme(undefined, iconMatch.id);
  }
  feedbackMsg.value = `✓ 已激活主题 ${item.displayName || item.name}`;
}

function formatDownloads(count: number): string {
  if (!count) return "0";
  if (count >= 1000000) return (count / 1000000).toFixed(1) + "M";
  if (count >= 1000) return (count / 1000).toFixed(1) + "k";
  return String(count);
}

// --- Channel 2: UniRTM Ecosystem Native Plugins ---
interface PluginItem {
  id: string;
  name: string;
  icon: string;
  version: string;
  author: string;
  description: string;
  tags: string[];
  stars: string;
  downloads: string;
  installed: boolean;
}

const nativePlugins = ref<PluginItem[]>([
  {
    id: "uniagent",
    name: "UniAgent 智能排错助手",
    icon: "🤖",
    version: "0.1.0",
    author: "SnowdreamTech 官方认证",
    description: "专注于开发编译与环境配置诊断的 AI 智能助手，通过独立 MCP 协议与工作台通讯，零内核侵入。",
    tags: ["AI Agent", "MCP Server", "Troubleshooter"],
    stars: "1.2k",
    downloads: "28k",
    installed: true,
  },
  {
    id: "security-suite",
    name: "Security Suite Pack",
    icon: "🛡️",
    version: "1.2.0",
    author: "SnowdreamTech 官方认证",
    description: "集成 Trivy、Syft、Gitleaks 深度源码与依赖漏洞扫描，自动生成 SBOM 软件物料清单。",
    tags: ["Trivy", "CVE Scan", "SBOM"],
    stars: "890",
    downloads: "15k",
    installed: true,
  },
  {
    id: "rust-provider",
    name: "Rust & Cargo Provider",
    icon: "🦀",
    version: "1.85.0",
    author: "社区维护",
    description: "为 UniRTM 提供原生 Rust/Cargo 极速版本下载与跨平台构建环境隔离支持。",
    tags: ["Rust", "Cargo", "Tool Provider"],
    stars: "640",
    downloads: "9.2k",
    installed: false,
  },
  {
    id: "datascience-pack",
    name: "Python DataScience Pack",
    icon: "🔬",
    version: "2.0.0",
    author: "社区维护",
    description: "预置 uv、ruff、ipython、jupyter 数据科学轻量级任务流与环境配置模板。",
    tags: ["Python", "uv", "Jupyter"],
    stars: "520",
    downloads: "7.1k",
    installed: false,
  },
]);
</script>

<style scoped>
.marketplace-view {
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
  gap: 16px;
  flex-wrap: wrap;
}

.view-title {
  font-size: 1.35rem;
  font-weight: 700;
  margin: 0 0 0.35rem;
  color: var(--text-main);
}

.view-desc {
  font-size: 0.84rem;
  color: var(--text-muted);
  margin: 0;
}

.channel-tabs {
  display: flex;
  padding: 4px;
  border-radius: 8px;
  gap: 4px;
}

.channel-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  border-radius: 6px;
  font-size: 0.84rem;
  font-weight: 500;
  color: var(--text-muted);
  background: transparent;
  border: 1px solid transparent;
  cursor: pointer;
  transition: all 0.2s ease;
}

.channel-btn:hover {
  color: var(--text-main);
  background: rgba(255, 255, 255, 0.05);
}

.channel-btn.active {
  color: var(--accent-cyan);
  background: rgba(0, 229, 255, 0.1);
  border-color: rgba(0, 229, 255, 0.3);
  font-weight: 600;
}

.market-channel-content {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.market-filter-bar {
  display: flex;
  gap: 8px;
  align-items: center;
  padding: 8px 12px;
  border-radius: 10px;
}

.search-input-wrapper {
  position: relative;
  flex: 1;
  display: flex;
  align-items: center;
}

.search-icon {
  position: absolute;
  left: 10px;
  font-size: 0.85rem;
  opacity: 0.6;
  pointer-events: none;
}

.market-search-input {
  width: 100%;
  padding: 7px 32px;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  border-radius: 6px;
  color: var(--text-main);
  font-size: 0.85rem;
  outline: none;
  transition: border-color 0.2s;
}

.market-search-input:focus {
  border-color: var(--accent-cyan);
}

.btn-clear-search {
  position: absolute;
  right: 10px;
  background: transparent;
  border: none;
  color: var(--text-muted);
  font-size: 0.75rem;
  cursor: pointer;
}

.btn-search-action {
  padding: 7px 20px;
  background: rgba(0, 229, 255, 0.12);
  border: 1px solid rgba(0, 229, 255, 0.3);
  color: var(--accent-cyan);
  border-radius: 6px;
  font-size: 0.84rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-search-action:hover:not(:disabled) {
  background: rgba(0, 229, 255, 0.22);
  border-color: var(--accent-cyan);
}

.market-chips-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  flex-wrap: wrap;
}

.chips-left {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.chips-label {
  font-size: 0.78rem;
  color: var(--text-muted);
  margin-right: 2px;
}

.chip-btn {
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  color: var(--text-muted);
  padding: 3px 10px;
  border-radius: 12px;
  font-size: 0.75rem;
  cursor: pointer;
  transition: all 0.2s;
}

.chip-btn:hover {
  color: var(--text-main);
  border-color: rgba(255, 255, 255, 0.2);
}

.chip-btn.active {
  background: rgba(0, 229, 255, 0.1);
  border-color: var(--accent-cyan);
  color: var(--accent-cyan);
  font-weight: 600;
}

.results-count-label {
  font-size: 0.76rem;
  color: var(--text-muted);
}

.feedback-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  background: rgba(16, 185, 129, 0.1);
  border: 1px solid rgba(16, 185, 129, 0.3);
  color: var(--success);
  border-radius: 6px;
  font-size: 0.82rem;
}

.feedback-banner.error {
  background: rgba(239, 68, 68, 0.1);
  border-color: rgba(239, 68, 68, 0.3);
  color: #ef4444;
}

.btn-close-banner {
  background: transparent;
  border: none;
  color: inherit;
  cursor: pointer;
  font-size: 0.8rem;
}

.error-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.3);
  color: #ef4444;
  border-radius: 6px;
  font-size: 0.82rem;
}

.btn-retry {
  background: transparent;
  border: 1px solid rgba(239, 68, 68, 0.4);
  color: #ef4444;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 0.76rem;
  cursor: pointer;
}

.market-loading-box,
.market-empty-box {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 40px 16px;
  color: var(--text-muted);
  font-size: 0.86rem;
  border-radius: 12px;
}

.empty-icon {
  font-size: 2rem;
}

.btn-reset-chip {
  margin-top: 6px;
  padding: 5px 14px;
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  color: var(--accent-cyan);
  border-radius: 6px;
  font-size: 0.8rem;
  cursor: pointer;
}

/* Extensions Grid */
.extensions-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 12px;
}

.theme-card,
.plugin-card {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  padding: 14px;
  border-radius: 12px;
  transition: transform 0.2s, border-color 0.2s, box-shadow 0.2s;
}

.theme-card:hover,
.plugin-card:hover {
  transform: translateY(-2px);
  border-color: rgba(0, 229, 255, 0.4);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.2);
}

.card-header-row {
  display: flex;
  gap: 10px;
  align-items: flex-start;
}

.card-icon-wrap {
  width: 44px;
  height: 44px;
  border-radius: 8px;
  background: var(--section-bg);
  border: 1px solid var(--card-border);
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  position: relative;
  flex-shrink: 0;
}

.card-img-icon {
  width: 100%;
  height: 100%;
  object-fit: cover;
  position: absolute;
  top: 0;
  left: 0;
  z-index: 1;
}

.card-fallback-icon {
  font-size: 1.3rem;
  z-index: 0;
}

.card-icon-box {
  font-size: 2rem;
  line-height: 1;
}

.card-title-col {
  display: flex;
  flex-direction: column;
  min-width: 0;
  flex: 1;
}

.card-name-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 6px;
}

.card-name {
  font-size: 0.92rem;
  font-weight: 600;
  color: var(--text-main);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.card-version {
  font-size: 0.7rem;
  color: var(--text-muted);
  background: rgba(255, 255, 255, 0.06);
  padding: 1px 4px;
  border-radius: 4px;
  flex-shrink: 0;
}

.card-namespace {
  font-size: 0.74rem;
  color: var(--text-muted);
  margin-top: 2px;
}

.card-desc {
  font-size: 0.78rem;
  color: var(--text-muted);
  margin: 10px 0;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  min-height: 2.2em;
}

.plugin-tags-row {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
  margin-bottom: 8px;
}

.plugin-tag {
  font-size: 0.7rem;
  background: rgba(0, 229, 255, 0.08);
  border: 1px solid rgba(0, 229, 255, 0.2);
  color: var(--accent-cyan);
  padding: 1px 6px;
  border-radius: 4px;
}

.card-footer-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 6px;
  padding-top: 10px;
  border-top: 1px solid rgba(255, 255, 255, 0.06);
}

.card-stats {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.74rem;
  color: var(--text-muted);
}

.card-actions {
  display: flex;
  align-items: center;
  gap: 6px;
}

.installed-group {
  display: flex;
  align-items: center;
  gap: 6px;
}

.badge-installed {
  font-size: 0.72rem;
  padding: 3px 8px;
  background: rgba(16, 185, 129, 0.15);
  border: 1px solid rgba(16, 185, 129, 0.3);
  color: var(--success);
  border-radius: 4px;
}

.btn-apply-theme {
  padding: 4px 12px;
  background: rgba(0, 229, 255, 0.12);
  border: 1px solid rgba(0, 229, 255, 0.3);
  color: var(--accent-cyan);
  border-radius: 4px;
  font-size: 0.76rem;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-apply-theme:hover {
  background: rgba(0, 229, 255, 0.25);
  border-color: var(--accent-cyan);
}

.btn-install-theme {
  padding: 4px 14px;
  background: rgba(0, 229, 255, 0.12);
  border: 1px solid rgba(0, 229, 255, 0.3);
  color: var(--accent-cyan);
  border-radius: 4px;
  font-size: 0.78rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.btn-install-theme:hover:not(:disabled) {
  background: rgba(0, 229, 255, 0.22);
  border-color: var(--accent-cyan);
}

.btn-install-theme:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.plugin-isolation-banner {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 12px 16px;
  border-radius: 10px;
  background: rgba(0, 229, 255, 0.05);
  border: 1px solid rgba(0, 229, 255, 0.18);
}

.banner-icon {
  font-size: 1.4rem;
}

.banner-text strong {
  display: block;
  font-size: 0.88rem;
  color: var(--text-main);
  margin-bottom: 2px;
}

.banner-text p {
  font-size: 0.8rem;
  color: var(--text-muted);
  margin: 0;
  line-height: 1.4;
}
</style>
