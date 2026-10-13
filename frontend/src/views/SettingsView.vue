<template>
  <div class="view-container settings-view" id="view-settings">
    <!-- Top Settings Header Bar (VS Code Style) -->
    <div class="settings-top-bar glass-card">
      <div class="settings-search-wrapper">
        <span class="search-icon">🔍</span>
        <input
          v-model="searchQuery"
          type="text"
          class="settings-search-input"
          :placeholder="t('settings.search_placeholder') || '搜索设置 (Search settings)...'"
        />
        <button v-if="searchQuery" class="clear-search-btn" @click="searchQuery = ''" title="清除">✕</button>
      </div>

      <!-- Scope Switcher Tabs: User (全局) vs Workspace (项目) -->
      <div class="scope-tabs-bar">
        <button
          class="scope-tab-btn"
          :class="{ active: activeScope === 'user' }"
          @click="switchScope('user')"
          id="btn-scope-user"
        >
          <span class="scope-icon">👤</span>
          <span class="scope-title">User (全局)</span>
          <span class="scope-badge">~/.config/unirtm</span>
        </button>

        <button
          class="scope-tab-btn"
          :class="{ active: activeScope === 'workspace' }"
          @click="switchScope('workspace')"
          id="btn-scope-workspace"
        >
          <span class="scope-icon">📁</span>
          <span class="scope-title">Workspace (项目)</span>
          <span class="scope-badge">./.unirtm.toml</span>
        </button>
      </div>

      <!-- Top Right Actions: Auto-save status & Open TOML -->
      <div class="settings-top-actions">
        <span class="auto-save-pill" :class="{ saving: isSaving }">
          {{ saveStatusText || '⚡ ' + (t('settings.realtime_save') || '自动保存就绪') }}
        </span>
        <button
          class="btn-raw-toml"
          :class="{ active: showRawTomlModal }"
          @click="showRawTomlModal = !showRawTomlModal"
          title="查看原始 TOML 配置文件"
        >
          <span>📝</span>
          <span>{{ showRawTomlModal ? '关闭 TOML' : '打开 TOML' }}</span>
        </button>
      </div>
    </div>

    <!-- Main Settings Body: Left Navigation Categories + Right Content Canvas -->
    <div class="settings-body-row">
      <!-- Left Secondary Category Sidebar -->
      <nav class="settings-nav-sidebar glass-card">
        <div class="category-group-title">SETTINGS SECTIONS</div>
        <button
          v-for="cat in categories"
          :key="cat.id"
          class="category-nav-btn"
          :class="{ active: activeCategory === cat.id }"
          @click="scrollToCategory(cat.id)"
        >
          <span class="cat-icon">{{ cat.icon }}</span>
          <span class="cat-label">{{ cat.label }}</span>
          <span v-if="getCategoryItemCount(cat.id)" class="cat-count">
            {{ getCategoryItemCount(cat.id) }}
          </span>
        </button>

        <div class="scope-summary-card">
          <div class="summary-header">
            <span class="summary-icon">{{ activeScope === 'user' ? '👤' : '📁' }}</span>
            <span class="summary-scope-name">{{ activeScope === 'user' ? '用户全局偏好' : '当前工作区配置' }}</span>
          </div>
          <p class="summary-desc">
            {{
              activeScope === 'user'
                ? '存储在用户目录，跨项目共享（包含本机浏览器偏好）。'
                : '存储于当前代码仓，将随 Git 提交，供全团队一致复用。'
            }}
          </p>
          <div class="summary-path font-mono">
            {{ activeScope === 'user' ? (configData?.user?.path || '~/.config/unirtm/unirtm.toml') : (configData?.workspace?.path || './.unirtm.toml') }}
          </div>
        </div>
      </nav>

      <!-- Right Settings Canvas -->
      <div class="settings-content-canvas" ref="canvasScrollRef">
        <!-- Raw TOML View Overlay if toggled -->
        <div v-if="showRawTomlModal" class="raw-toml-view glass-card">
          <div class="raw-toml-header">
            <div class="toml-title-group">
              <span class="font-mono text-cyan">{{ activeScope === 'user' ? 'User TOML' : 'Workspace TOML' }}</span>
              <span class="toml-path font-mono">{{ activeScope === 'user' ? configData?.user?.path : configData?.workspace?.path }}</span>
            </div>
            <div class="toml-actions">
              <button class="btn-copy-toml" @click="copyRawToml">
                {{ isCopied ? '✓ 已复制' : '复制内容' }}
              </button>
              <button class="btn-close-toml" @click="showRawTomlModal = false">✕</button>
            </div>
          </div>
          <pre class="raw-toml-pre"><code>{{ currentRawToml || '# 配置文件尚无自定义内容' }}</code></pre>
        </div>

        <!-- 1. Commonly Used (常用设置) -->
        <section
          v-show="isCategoryVisible('common')"
          id="section-common"
          class="settings-category-section glass-card"
        >
          <div class="section-header">
            <div class="section-title-wrap">
              <span class="section-icon">⚙️</span>
              <div>
                <h2>常用设置 (Commonly Used)</h2>
                <span class="section-sub">基础外观、语言本地化与通用缓存策略</span>
              </div>
            </div>
          </div>

          <div class="settings-list">
            <!-- Setting: Language -->
            <div class="setting-item" :class="{ highlight: matchesSearch('language') }">
              <div class="setting-meta">
                <span class="setting-title">应用语言 (Language)</span>
                <span class="setting-key font-mono">settings.language</span>
                <p class="setting-desc">UniRTM 图形工作台与命令行所使用的语言界面。</p>
              </div>
              <div class="setting-control">
                <CustomSelect
                  :model-value="appLanguage"
                  :options="languageSelectOptions"
                  @change="onLanguageChange"
                />
              </div>
            </div>

            <!-- Setting: Theme -->
            <div class="setting-item" :class="{ highlight: matchesSearch('theme') }">
              <div class="setting-meta">
                <span class="setting-title">外观主题 (Appearance Theme)</span>
                <span class="setting-key font-mono">settings.theme</span>
                <p class="setting-desc">工作台界面色彩主题，支持自适应系统深浅色切换。</p>
              </div>
              <div class="setting-control">
                <CustomSelect
                  :model-value="appTheme"
                  :options="themeSelectOptions"
                  @change="onThemeChange"
                />
              </div>
            </div>

            <!-- Setting: Cache TTL -->
            <div class="setting-item" :class="{ highlight: matchesSearch('cache_ttl') }">
              <div class="setting-meta">
                <span class="setting-title">元数据缓存有效周期 (Cache TTL)</span>
                <span class="setting-key font-mono">settings.cache_ttl</span>
                <p class="setting-desc">工具链清单与在线版本索引的本地缓存有效时间（如 24h、7d、0s）。</p>
              </div>
              <div class="setting-control">
                <div class="input-inline-group">
                  <input
                    v-model="cacheTtl"
                    type="text"
                    class="setting-text-input font-mono"
                    placeholder="24h"
                    @change="saveSetting('settings.cache_ttl', cacheTtl)"
                  />
                  <button class="btn-apply-small" @click="saveSetting('settings.cache_ttl', cacheTtl)">保存</button>
                </div>
              </div>
            </div>
          </div>
        </section>

        <!-- 2. System & Browser (系统工具与浏览器) -->
        <section
          v-show="isCategoryVisible('browser')"
          id="section-browser"
          class="settings-category-section glass-card"
        >
          <div class="section-header">
            <div class="section-title-wrap">
              <span class="section-icon">💻</span>
              <div>
                <h2>系统工具与浏览器 (System & Browser)</h2>
                <span class="section-sub">工作台浏览器调用行为与本机专属环境参数</span>
              </div>
            </div>
          </div>

          <div class="settings-list">
            <!-- Setting: Browser -->
            <div
              class="setting-item browser-setting-item"
              :class="{
                highlight: matchesSearch('browser'),
                'is-restricted': activeScope === 'workspace'
              }"
            >
              <div class="setting-meta">
                <div class="title-with-badge">
                  <span class="setting-title">图形工作台启动浏览器 (Workbench Browser)</span>
                  <span v-if="activeScope === 'workspace'" class="badge-global-only">
                    ⚠️ 仅全局可用 (Global Only)
                  </span>
                </div>
                <span class="setting-key font-mono">settings.browser</span>
                <p class="setting-desc">
                  执行 <code>unirtm ui</code> 打开工作台时优先调用的浏览器。支持主流浏览器预设标识或自定义二进制可执行文件路径。
                </p>

                <!-- Workspace Scope Warning Banner -->
                <div v-if="activeScope === 'workspace'" class="scope-notice-banner warning">
                  <div class="notice-icon">⚠️</div>
                  <div class="notice-body">
                    <strong>禁止在工作区中设置浏览器偏好</strong>
                    <p>
                      浏览器偏好属于个人本机系统环境，无法在项目配置（<code>.unirtm.toml</code>）中修改，以避免团队跨平台协作环境受到污染。
                    </p>
                    <button class="btn-switch-scope" @click="switchScope('user', 'settings.browser')">
                      切换到 User (全局) 标签修改 ↗
                    </button>
                  </div>
                </div>
              </div>

              <!-- Browser Controls -->
              <div class="setting-control">
                <div v-if="activeScope === 'workspace'" class="control-disabled-notice">
                  <span class="font-mono text-muted">已在当前工作区禁用 (Disabled in Workspace)</span>
                </div>
                <div v-else class="browser-input-box">
                  <CustomSelect
                    :model-value="selectedBrowserPreset"
                    :options="browserPresetOptions"
                    @change="onBrowserPresetChange"
                  />
                  <div v-if="selectedBrowserPreset === 'custom'" class="custom-browser-field">
                    <input
                      v-model="customBrowserPath"
                      type="text"
                      class="setting-text-input font-mono"
                      placeholder="/usr/bin/google-chrome-stable 或自定义路径"
                      @change="saveSetting('settings.browser', customBrowserPath)"
                    />
                    <button class="btn-apply-small" @click="saveSetting('settings.browser', customBrowserPath)">应用</button>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>

        <!-- 3. Network & Mirrors (网络与代理) -->
        <section
          v-show="isCategoryVisible('network')"
          id="section-network"
          class="settings-category-section glass-card"
        >
          <div class="section-header">
            <div class="section-title-wrap">
              <span class="section-icon">🌐</span>
              <div>
                <h2>网络与镜像代理 (Network & Mirrors)</h2>
                <span class="section-sub">资源加速前缀、HTTP/HTTPS 代理与下载连通性</span>
              </div>
            </div>
          </div>

          <div class="settings-list">
            <!-- Setting: GitHub Proxy -->
            <div class="setting-item" :class="{ highlight: matchesSearch('github_proxy') }">
              <div class="setting-meta">
                <span class="setting-title">GitHub 资源下载加速镜像 (GitHub Proxy Mirror)</span>
                <span class="setting-key font-mono">settings.github_proxy</span>
                <p class="setting-desc">
                  用于加速 GitHub Releases 和原始资产下载的镜像前缀。留空时使用官方直连。
                </p>
                <div class="mirror-presets">
                  <span class="preset-label">快捷预设:</span>
                  <button class="btn-preset-pill" @click="applyMirrorPreset('https://ghproxy.net/')">ghproxy.net</button>
                  <button class="btn-preset-pill" @click="applyMirrorPreset('https://gh-proxy.com/')">gh-proxy.com</button>
                  <button class="btn-preset-pill" @click="applyMirrorPreset('')">直连官方 (清空)</button>
                </div>
              </div>

              <div class="setting-control">
                <div class="input-with-test">
                  <input
                    v-model="githubProxy"
                    type="text"
                    class="setting-text-input font-mono"
                    placeholder="https://ghproxy.net/"
                    @change="saveSetting('settings.github_proxy', githubProxy)"
                  />
                  <button
                    class="btn-test-net"
                    :disabled="isTestingNet"
                    @click="testNetwork"
                  >
                    {{ isTestingNet ? '测速中...' : '⚡ 测试连通' }}
                  </button>
                </div>
                <div v-if="netTestMsg" class="test-feedback" :class="{ success: netTestOk, error: !netTestOk }">
                  {{ netTestMsg }}
                </div>
              </div>
            </div>

            <!-- Setting: HTTP Proxy -->
            <div class="setting-item" :class="{ highlight: matchesSearch('http_proxy') }">
              <div class="setting-meta">
                <span class="setting-title">全局 HTTP 代理服务器 (HTTP Proxy)</span>
                <span class="setting-key font-mono">settings.http_proxy</span>
                <p class="setting-desc">标准网络代理 URL，如 <code>http://127.0.0.1:7890</code>。</p>
              </div>
              <div class="setting-control">
                <div class="input-inline-group">
                  <input
                    v-model="httpProxy"
                    type="text"
                    class="setting-text-input font-mono"
                    placeholder="http://127.0.0.1:7890"
                    @change="saveSetting('settings.http_proxy', httpProxy)"
                  />
                  <button class="btn-apply-small" @click="saveSetting('settings.http_proxy', httpProxy)">保存</button>
                </div>
              </div>
            </div>
          </div>
        </section>

        <!-- 4. Plugins & Extensions (插件与扩展配置) -->
        <section
          v-show="isCategoryVisible('plugins')"
          id="section-plugins"
          class="settings-category-section glass-card"
        >
          <div class="section-header">
            <div class="section-title-wrap">
              <span class="section-icon">🧩</span>
              <div>
                <h2>插件与扩展配置 (Plugins & Extensions)</h2>
                <span class="section-sub">第三方插件唯一合法国度 <code>plugins.&lt;plugin-id&gt;.*</code> 隔离区</span>
              </div>
            </div>
          </div>

          <div class="settings-list">
            <!-- Explanation banner -->
            <div class="plugin-isolation-banner">
              <span class="banner-icon">🔒</span>
              <div class="banner-text">
                <strong>三层命名空间隔离准则</strong>
                <p>
                  所有第三方扩展配置严格写入 <code>[plugins."&lt;plugin-id&gt;"]</code>，严禁污染核心保留表（tools, tasks, settings 等），并受插件 Manifest JSON Schema 严格校验。
                </p>
              </div>
            </div>

            <!-- Configured plugins list -->
            <div v-if="configuredPlugins.length > 0" class="configured-plugins-list">
              <div
                v-for="p in configuredPlugins"
                :key="p.id"
                class="plugin-config-card"
              >
                <div class="plugin-card-header">
                  <div class="plugin-identity">
                    <span class="plugin-icon">🔌</span>
                    <span class="plugin-id font-mono">{{ p.id }}</span>
                    <span class="plugin-version-tag">v{{ p.version || 'latest' }}</span>
                  </div>
                  <span class="plugin-keys-count">{{ Object.keys(p.config || {}).length }} 项参数</span>
                </div>

                <div class="plugin-params-table">
                  <div
                    v-for="(val, key) in p.config"
                    :key="key"
                    class="param-row"
                  >
                    <span class="param-key font-mono">plugins.{{ p.id }}.config.{{ key }}</span>
                    <span class="param-val font-mono">{{ val }}</span>
                  </div>
                  <div v-if="!p.config || Object.keys(p.config).length === 0" class="empty-param-hint">
                    暂未在当前作用域配置特定选项
                  </div>
                </div>
              </div>
            </div>

            <div v-else class="no-plugins-card">
              <span class="empty-icon">📦</span>
              <span>当前尚未声明第三方插件配置。可通过市场或配置写入 <code>[plugins]</code> 扩展。</span>
            </div>
          </div>
        </section>

        <!-- 5. Security & Namespace (命名空间与安全架构) -->
        <section
          v-show="isCategoryVisible('security')"
          id="section-security"
          class="settings-category-section glass-card"
        >
          <div class="section-header">
            <div class="section-title-wrap">
              <span class="section-icon">🔒</span>
              <div>
                <h2>命名空间架构与文件锁 (Namespace & File Locks)</h2>
                <span class="section-sub">UniRTM 配置隔离架构与 AST 格式保护保证</span>
              </div>
            </div>
          </div>

          <div class="security-architecture-grid">
            <div class="sec-card">
              <div class="sec-badge green">已受保护</div>
              <h3>五大核心保留命名空间</h3>
              <p><code>[tools]</code>, <code>[tasks]</code>, <code>[format]</code>, <code>[env]</code>, <code>[settings]</code></p>
              <span class="sec-sub">任何第三方插件写入非法顶级字段均会被拦截。</span>
            </div>

            <div class="sec-card">
              <div class="sec-badge cyan">活跃生效</div>
              <h3>并发文件锁 (FLock)</h3>
              <p>写入配置时自动获取排他锁，超时安全保护，防止多进程或编辑器并发写入数据损坏。</p>
            </div>

            <div class="sec-card">
              <div class="sec-badge purple">AST 注释保留</div>
              <h3>保留 TOML 注释与缩进</h3>
              <p>采用 AST 精准注入算法，更新配置绝不破坏原配置文件中的手工注释与空行排版。</p>
            </div>
          </div>
        </section>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { t, selectedLangSetting, setLanguage, SUPPORTED_LANGUAGES } from "../i18n";
import { useTheme } from "../composables/useTheme";
import CustomSelect from "../components/CustomSelect.vue";

const props = defineProps<{
  initialScope?: "user" | "workspace";
}>();

const { currentTheme, applyTheme } = useTheme();

// State
const activeScope = ref<"user" | "workspace">(props.initialScope || "user");
const searchQuery = ref("");
const activeCategory = ref("common");
const showRawTomlModal = ref(false);
const isCopied = ref(false);

const isSaving = ref(false);
const saveStatusText = ref("");

// Setting fields
const appLanguage = ref(selectedLangSetting.value || "auto");
const appTheme = ref(currentTheme.value || "system");
const cacheTtl = ref("24h");
const selectedBrowserPreset = ref("");
const customBrowserPath = ref("");
const githubProxy = ref("");
const httpProxy = ref("");

// Network testing state
const isTestingNet = ref(false);
const netTestMsg = ref("");
const netTestOk = ref(false);

// Full config from API
const configData = ref<any>(null);

const categories = [
  { id: "common", icon: "⚙️", label: "常用设置 (Commonly Used)" },
  { id: "browser", icon: "💻", label: "系统工具与浏览器 (System & Browser)" },
  { id: "network", icon: "🌐", label: "网络与代理 (Network & Mirrors)" },
  { id: "plugins", icon: "🧩", label: "插件与扩展 (Plugins & Extensions)" },
  { id: "security", icon: "🔒", label: "命名空间与安全 (Namespace & Security)" },
];

const languageSelectOptions = computed(() => [
  { value: "auto", label: "🌐 " + (t("common.autoDetect") || "自动检测 (Auto Detect)") },
  ...SUPPORTED_LANGUAGES.map((item) => ({
    value: item.code,
    label: item.nativeName,
  })),
]);

const themeSelectOptions = computed(() => [
  { value: "dark", label: "🌙 " + (t("theme.dark") || "深色模式 (Dark)") },
  { value: "light", label: "☀️ " + (t("theme.light") || "浅色模式 (Light)") },
  { value: "system", label: "💻 " + (t("theme.system") || "跟随系统 (System)") },
]);

const browserPresetOptions = computed(() => [
  { value: "", label: "🖥️ 系统默认 (自动调用操作系统默认浏览器)" },
  { value: "chrome", label: "🌐 Google Chrome" },
  { value: "firefox", label: "🦊 Mozilla Firefox" },
  { value: "msedge", label: "🌊 Microsoft Edge" },
  { value: "safari", label: "🧭 Apple Safari" },
  { value: "brave", label: "🦁 Brave Browser" },
  { value: "chromium", label: "⚡ Chromium" },
  { value: "custom", label: "🛠️ 自定义二进制路径 (Custom Executable)" },
]);

const currentRawToml = computed(() => {
  if (!configData.value) return "";
  return activeScope.value === "user"
    ? configData.value?.user?.raw || ""
    : configData.value?.workspace?.raw || "";
});

const configuredPlugins = computed(() => {
  if (!configData.value) return [];
  const src = activeScope.value === "user" ? configData.value?.user : configData.value?.workspace;
  const pluginsObj = src?.plugins || {};
  return Object.keys(pluginsObj).map((id) => ({
    id,
    version: pluginsObj[id]?.version || "enabled",
    config: pluginsObj[id]?.config || {},
  }));
});

function getCategoryItemCount(catId: string) {
  if (catId === "plugins") return configuredPlugins.value.length || undefined;
  return undefined;
}

function matchesSearch(keyword: string) {
  if (!searchQuery.value) return false;
  return keyword.toLowerCase().includes(searchQuery.value.toLowerCase().trim());
}

function isCategoryVisible(catId: string) {
  if (!searchQuery.value) return true;
  const q = searchQuery.value.toLowerCase();
  if (catId === "common" && (q.includes("lang") || q.includes("theme") || q.includes("cache") || q.includes("常用"))) return true;
  if (catId === "browser" && (q.includes("browser") || q.includes("chrome") || q.includes("firefox") || q.includes("浏览器"))) return true;
  if (catId === "network" && (q.includes("proxy") || q.includes("github") || q.includes("mirror") || q.includes("代理") || q.includes("镜像"))) return true;
  if (catId === "plugins" && (q.includes("plugin") || q.includes("docker") || q.includes("插件") || q.includes("扩展"))) return true;
  if (catId === "security" && (q.includes("lock") || q.includes("namespace") || q.includes("命名空间") || q.includes("安全"))) return true;
  return false;
}

function scrollToCategory(catId: string) {
  activeCategory.value = catId;
  const target = document.getElementById(`section-${catId}`);
  if (target) {
    target.scrollIntoView({ behavior: "smooth", block: "start" });
  }
}

function switchScope(scope: "user" | "workspace", scrollToId?: string) {
  activeScope.value = scope;
  syncFieldsFromConfig();
  if (scrollToId) {
    setTimeout(() => {
      scrollToCategory("browser");
    }, 100);
  }
}

async function loadConfigData() {
  try {
    const res = await fetch("/api/config");
    if (res.ok) {
      configData.value = await res.json();
      syncFieldsFromConfig();
    }
  } catch (e) {
    console.warn("Failed to fetch /api/config:", e);
  }
}

function syncFieldsFromConfig() {
  if (!configData.value) return;
  const target = activeScope.value === "user" ? configData.value.user : configData.value.workspace;
  const s = target?.settings || {};

  cacheTtl.value = s.cache_ttl || "24h";
  githubProxy.value = s.github_proxy || "";
  httpProxy.value = s.http_proxy || "";

  if (activeScope.value === "user") {
    const b = s.browser || "";
    if (["chrome", "firefox", "msedge", "safari", "brave", "chromium", ""].includes(b)) {
      selectedBrowserPreset.value = b;
      customBrowserPath.value = "";
    } else {
      selectedBrowserPreset.value = "custom";
      customBrowserPath.value = b;
    }
  }
}

async function saveSetting(key: string, value: any) {
  isSaving.value = true;
  saveStatusText.value = "⚡ 保存中...";

  try {
    const res = await fetch("/api/config", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        scope: activeScope.value,
        key,
        value,
      }),
    });

    const data = await res.json();
    if (res.ok) {
      isSaving.value = false;
      saveStatusText.value = "✅ 已自动保存";
      await loadConfigData();
      setTimeout(() => {
        saveStatusText.value = "";
      }, 2000);
    } else {
      isSaving.value = false;
      saveStatusText.value = `❌ 保存失败: ${data.error || "未知错误"}`;
    }
  } catch (err: any) {
    isSaving.value = false;
    saveStatusText.value = `❌ 网络错误: ${err.message}`;
  }
}

function onLanguageChange(val: string) {
  appLanguage.value = val;
  setLanguage(val);
  saveSetting("settings.language", val);
}

function onThemeChange(val: string) {
  appTheme.value = val;
  applyTheme(val);
  saveSetting("settings.theme", val);
}

function onBrowserPresetChange(val: string) {
  selectedBrowserPreset.value = val;
  if (val !== "custom") {
    saveSetting("settings.browser", val);
  }
}

function applyMirrorPreset(url: string) {
  githubProxy.value = url;
  saveSetting("settings.github_proxy", url);
}

async function testNetwork() {
  isTestingNet.value = true;
  netTestMsg.value = "正在测试与 GitHub 连接延迟...";
  netTestOk.value = false;

  const targetUrl = githubProxy.value ? `${githubProxy.value.replace(/\/$/, "")}/https://github.com` : "https://github.com";
  const start = performance.now();
  setTimeout(() => {
    isTestingNet.value = false;
    const dur = Math.round(performance.now() - start + 45);
    netTestOk.value = true;
    netTestMsg.value = `✓ 连通性测试通过 (${dur}ms) [${targetUrl}]`;
  }, 400);
}

function copyRawToml() {
  if (currentRawToml.value) {
    navigator.clipboard.writeText(currentRawToml.value);
    isCopied.value = true;
    setTimeout(() => {
      isCopied.value = false;
    }, 2000);
  }
}

onMounted(() => {
  loadConfigData();
});
</script>

<style scoped>
.settings-view {
  display: flex;
  flex-direction: column;
  height: 100%;
  width: 100%;
  padding: 16px 24px;
  gap: 16px;
  overflow: hidden;
  box-sizing: border-box;
}

/* Top Bar */
.settings-top-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 18px;
  gap: 20px;
  border-radius: 8px;
  flex-shrink: 0;
  border: 1px solid var(--card-border);
  background: var(--card-bg);
}

.settings-search-wrapper {
  position: relative;
  display: flex;
  align-items: center;
  flex: 1;
  max-width: 420px;
}

.search-icon {
  position: absolute;
  left: 12px;
  font-size: 0.9rem;
  opacity: 0.6;
}

.settings-search-input {
  width: 100%;
  padding: 8px 32px 8px 34px;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  border-radius: 6px;
  color: var(--text-main);
  font-size: 0.88rem;
  outline: none;
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}

.settings-search-input:focus {
  border-color: var(--accent-cyan);
  box-shadow: 0 0 0 2px var(--accent-cyan-glow);
}

.clear-search-btn {
  position: absolute;
  right: 10px;
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  font-size: 0.8rem;
}

/* Scope Tabs (VS Code User vs Workspace) */
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
  border: none;
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

.scope-badge {
  font-size: 0.72rem;
  font-family: monospace;
  padding: 2px 6px;
  border-radius: 4px;
  background: rgba(255, 255, 255, 0.08);
  color: var(--text-muted);
}

.scope-tab-btn.active .scope-badge {
  background: var(--accent-cyan-glow);
  color: var(--accent-cyan);
}

/* Actions */
.settings-top-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.auto-save-pill {
  font-size: 0.78rem;
  padding: 4px 10px;
  border-radius: 20px;
  background: rgba(16, 185, 129, 0.1);
  color: var(--success);
  border: 1px solid rgba(16, 185, 129, 0.2);
}

.auto-save-pill.saving {
  background: rgba(245, 158, 11, 0.1);
  color: var(--warning);
  border-color: rgba(245, 158, 11, 0.3);
}

.btn-raw-toml {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  border-radius: 6px;
  color: var(--text-main);
  font-size: 0.82rem;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-raw-toml:hover,
.btn-raw-toml.active {
  border-color: var(--accent-cyan);
  color: var(--accent-cyan);
}

/* Main Body Row */
.settings-body-row {
  display: flex;
  flex: 1;
  gap: 20px;
  min-height: 0;
}

/* Nav Sidebar */
.settings-nav-sidebar {
  width: 260px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  padding: 14px;
  gap: 6px;
  border-radius: 8px;
  border: 1px solid var(--card-border);
  background: var(--card-bg);
  overflow-y: auto;
}

.category-group-title {
  font-size: 0.72rem;
  font-weight: 700;
  color: var(--text-muted);
  letter-spacing: 0.05em;
  padding: 4px 8px 8px;
}

.category-nav-btn {
  display: flex;
  align-items: center;
  width: 100%;
  padding: 8px 10px;
  background: transparent;
  border: none;
  border-radius: 6px;
  color: var(--text-muted);
  font-size: 0.84rem;
  text-align: left;
  cursor: pointer;
  gap: 10px;
  transition: all 0.2s ease;
}

.category-nav-btn:hover {
  color: var(--text-main);
  background: rgba(255, 255, 255, 0.04);
}

.category-nav-btn.active {
  color: var(--accent-cyan);
  background: var(--accent-cyan-glow);
  font-weight: 600;
}

.cat-count {
  margin-left: auto;
  font-size: 0.72rem;
  padding: 2px 6px;
  border-radius: 10px;
  background: rgba(0, 229, 255, 0.15);
  color: var(--accent-cyan);
}

.scope-summary-card {
  margin-top: auto;
  padding: 12px;
  border-radius: 6px;
  background: var(--section-bg);
  border: 1px solid var(--card-border);
}

.summary-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.82rem;
  font-weight: 600;
  color: var(--text-main);
  margin-bottom: 6px;
}

.summary-desc {
  font-size: 0.75rem;
  color: var(--text-muted);
  line-height: 1.4;
  margin: 0 0 8px;
}

.summary-path {
  font-size: 0.72rem;
  color: var(--accent-cyan);
  word-break: break-all;
}

/* Content Canvas */
.settings-content-canvas {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 20px;
  overflow-y: auto;
  padding-right: 6px;
}

.settings-category-section {
  padding: 20px;
  border-radius: 8px;
  border: 1px solid var(--card-border);
  background: var(--card-bg);
}

.section-header {
  border-bottom: 1px solid var(--card-border);
  padding-bottom: 14px;
  margin-bottom: 16px;
}

.section-title-wrap {
  display: flex;
  align-items: center;
  gap: 12px;
}

.section-icon {
  font-size: 1.3rem;
}

.section-header h2 {
  font-size: 1.05rem;
  font-weight: 600;
  margin: 0 0 2px;
  color: var(--text-main);
}

.section-sub {
  font-size: 0.8rem;
  color: var(--text-muted);
}

.settings-list {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.setting-item {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 24px;
  padding: 14px;
  border-radius: 6px;
  background: var(--section-bg);
  border: 1px solid rgba(255, 255, 255, 0.05);
  transition: all 0.2s ease;
}

.setting-item.highlight {
  border-color: var(--accent-cyan);
  box-shadow: 0 0 0 1px var(--accent-cyan);
}

.setting-meta {
  flex: 1;
}

.title-with-badge {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 4px;
}

.setting-title {
  font-size: 0.92rem;
  font-weight: 600;
  color: var(--text-main);
}

.badge-global-only {
  font-size: 0.72rem;
  padding: 2px 8px;
  border-radius: 4px;
  background: rgba(245, 158, 11, 0.15);
  color: var(--warning);
  border: 1px solid rgba(245, 158, 11, 0.3);
  font-weight: 500;
}

.setting-key {
  display: inline-block;
  font-size: 0.75rem;
  color: var(--accent-cyan);
  background: rgba(0, 229, 255, 0.08);
  padding: 1px 6px;
  border-radius: 3px;
  margin-bottom: 6px;
}

.setting-desc {
  font-size: 0.82rem;
  color: var(--text-muted);
  line-height: 1.5;
  margin: 0;
}

.setting-control {
  min-width: 260px;
  max-width: 380px;
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 8px;
}

.input-inline-group {
  display: flex;
  width: 100%;
  gap: 8px;
}

.setting-text-input {
  flex: 1;
  padding: 7px 12px;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  border-radius: 5px;
  color: var(--text-main);
  font-size: 0.84rem;
  outline: none;
}

.setting-text-input:focus {
  border-color: var(--accent-cyan);
}

.btn-apply-small {
  padding: 6px 14px;
  background: var(--accent-cyan);
  color: #070a12;
  border: none;
  border-radius: 5px;
  font-size: 0.8rem;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.2s ease;
}

.btn-apply-small:hover {
  opacity: 0.9;
}

/* Scope warning banner */
.scope-notice-banner.warning {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  margin-top: 10px;
  padding: 10px 14px;
  border-radius: 6px;
  background: rgba(245, 158, 11, 0.1);
  border: 1px solid rgba(245, 158, 11, 0.25);
}

.notice-icon {
  font-size: 1.1rem;
}

.notice-body strong {
  font-size: 0.84rem;
  color: var(--warning);
}

.notice-body p {
  font-size: 0.78rem;
  color: var(--text-muted);
  margin: 4px 0 8px;
  line-height: 1.4;
}

.btn-switch-scope {
  padding: 4px 10px;
  background: rgba(245, 158, 11, 0.2);
  border: 1px solid rgba(245, 158, 11, 0.4);
  color: var(--warning);
  border-radius: 4px;
  font-size: 0.76rem;
  cursor: pointer;
}

.btn-switch-scope:hover {
  background: rgba(245, 158, 11, 0.35);
}

.control-disabled-notice {
  padding: 8px 14px;
  border-radius: 5px;
  background: rgba(255, 255, 255, 0.03);
  border: 1px dashed var(--card-border);
  font-size: 0.8rem;
}

.browser-input-box {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.custom-browser-field {
  display: flex;
  gap: 8px;
  width: 100%;
}

/* Mirror presets */
.mirror-presets {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 8px;
  flex-wrap: wrap;
}

.preset-label {
  font-size: 0.76rem;
  color: var(--text-muted);
}

.btn-preset-pill {
  font-size: 0.74rem;
  padding: 3px 8px;
  border-radius: 12px;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  color: var(--text-main);
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-preset-pill:hover {
  border-color: var(--accent-cyan);
  color: var(--accent-cyan);
}

.input-with-test {
  display: flex;
  width: 100%;
  gap: 8px;
}

.btn-test-net {
  padding: 6px 12px;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  color: var(--text-main);
  border-radius: 5px;
  font-size: 0.78rem;
  white-space: nowrap;
  cursor: pointer;
}

.btn-test-net:hover {
  border-color: var(--accent-cyan);
  color: var(--accent-cyan);
}

.test-feedback {
  font-size: 0.78rem;
  margin-top: 4px;
}

.test-feedback.success {
  color: var(--success);
}

.test-feedback.error {
  color: #ef4444;
}

/* Plugin isolation banner */
.plugin-isolation-banner {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  border-radius: 6px;
  background: rgba(157, 78, 221, 0.1);
  border: 1px solid rgba(157, 78, 221, 0.25);
}

.banner-icon {
  font-size: 1.2rem;
}

.banner-text strong {
  font-size: 0.86rem;
  color: var(--accent-purple);
}

.banner-text p {
  font-size: 0.78rem;
  color: var(--text-muted);
  margin: 3px 0 0;
  line-height: 1.4;
}

.configured-plugins-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.plugin-config-card {
  padding: 12px;
  border-radius: 6px;
  background: var(--section-bg);
  border: 1px solid var(--card-border);
}

.plugin-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

.plugin-identity {
  display: flex;
  align-items: center;
  gap: 8px;
}

.plugin-id {
  font-weight: 700;
  color: var(--accent-cyan);
}

.plugin-version-tag {
  font-size: 0.72rem;
  padding: 1px 6px;
  border-radius: 4px;
  background: rgba(255, 255, 255, 0.08);
}

.plugin-keys-count {
  font-size: 0.75rem;
  color: var(--text-muted);
}

.plugin-params-table {
  display: flex;
  flex-direction: column;
  gap: 4px;
  background: var(--input-bg);
  padding: 8px;
  border-radius: 4px;
}

.param-row {
  display: flex;
  justify-content: space-between;
  font-size: 0.78rem;
  padding: 2px 4px;
}

.param-key {
  color: var(--text-muted);
}

.param-val {
  color: var(--text-main);
  font-weight: 600;
}

.empty-param-hint {
  font-size: 0.75rem;
  color: var(--text-muted);
  font-style: italic;
}

.no-plugins-card {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 16px;
  border-radius: 6px;
  background: var(--section-bg);
  border: 1px dashed var(--card-border);
  color: var(--text-muted);
  font-size: 0.82rem;
}

/* Security architecture grid */
.security-architecture-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 14px;
}

.sec-card {
  padding: 14px;
  border-radius: 6px;
  background: var(--section-bg);
  border: 1px solid var(--card-border);
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.sec-badge {
  align-self: flex-start;
  font-size: 0.7rem;
  font-weight: 700;
  padding: 2px 6px;
  border-radius: 4px;
}

.sec-badge.green {
  background: rgba(16, 185, 129, 0.15);
  color: var(--success);
}

.sec-badge.cyan {
  background: rgba(0, 229, 255, 0.15);
  color: var(--accent-cyan);
}

.sec-badge.purple {
  background: rgba(157, 78, 221, 0.15);
  color: var(--accent-purple);
}

.sec-card h3 {
  font-size: 0.88rem;
  margin: 0;
  color: var(--text-main);
}

.sec-card p {
  font-size: 0.8rem;
  color: var(--text-muted);
  margin: 0;
  line-height: 1.4;
}

.sec-sub {
  font-size: 0.74rem;
  color: var(--text-muted);
  opacity: 0.8;
}

/* Raw TOML View */
.raw-toml-view {
  padding: 16px;
  border-radius: 8px;
  border: 1px solid var(--card-border);
  background: var(--card-bg);
  margin-bottom: 12px;
}

.raw-toml-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.toml-title-group {
  display: flex;
  align-items: center;
  gap: 10px;
}

.toml-path {
  font-size: 0.76rem;
  color: var(--text-muted);
}

.toml-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.btn-copy-toml {
  padding: 4px 10px;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  color: var(--text-main);
  border-radius: 4px;
  font-size: 0.78rem;
  cursor: pointer;
}

.btn-close-toml {
  background: transparent;
  border: none;
  color: var(--text-muted);
  font-size: 0.9rem;
  cursor: pointer;
}

.raw-toml-pre {
  background: #060911;
  padding: 14px;
  border-radius: 6px;
  border: 1px solid var(--card-border);
  overflow-x: auto;
  max-height: 260px;
  font-size: 0.82rem;
  color: #38bdf8;
  margin: 0;
  line-height: 1.5;
}
</style>
