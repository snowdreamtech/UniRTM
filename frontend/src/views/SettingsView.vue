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

      <!-- Scope Switcher Tabs: User vs Workspace -->
      <div class="scope-switcher-group">
        <div class="scope-tabs-bar">
          <button
            class="scope-tab-btn"
            :class="{ active: activeScope === 'user' }"
            @click="switchScope('user')"
            id="btn-scope-user"
            :title="configData?.user?.path || '~/.config/unirtm/unirtm.toml'"
          >
            <span class="scope-icon">👤</span>
            <span class="scope-title">{{ t("settings.scope_user") || "用户" }}</span>
          </button>

          <button
            class="scope-tab-btn"
            :class="{ active: activeScope === 'workspace' }"
            @click="switchScope('workspace')"
            id="btn-scope-workspace"
            :title="configData?.workspace?.path || './.unirtm.toml'"
          >
            <span class="scope-icon">📁</span>
            <span class="scope-title">{{ t("settings.scope_workspace") || "工作区" }}</span>
          </button>
        </div>
        <span
          class="active-scope-badge font-mono"
          :title="activeScope === 'user' ? (configData?.user?.path || '~/.config/unirtm/unirtm.toml') : (configData?.workspace?.path || './.unirtm.toml')"
        >
          {{ activeScope === 'user' ? (configData?.user?.path || '~/.config/unirtm/unirtm.toml') : (configData?.workspace?.path || './.unirtm.toml') }}
        </span>
      </div>

      <!-- Top Right Actions: Auto-save status & Open TOML -->
      <div class="settings-top-actions">
        <span class="auto-save-pill" :class="{ saving: isSaving }">
          {{ saveStatusText || '⚡ ' + (t('settings.realtime_save') || '修改已实时生效') }}
        </span>
        <button
          class="btn-raw-toml"
          :class="{ active: showRawTomlModal }"
          @click="showRawTomlModal = !showRawTomlModal"
          :title="t('settings.open_toml') || '打开 TOML'"
        >
          <span>📝</span>
          <span>{{ showRawTomlModal ? (t('settings.close_toml') || '关闭 TOML') : (t('settings.open_toml') || '打开 TOML') }}</span>
        </button>
      </div>
    </div>

    <!-- Main Settings Body: Full-width Content Canvas -->
    <div class="settings-content-canvas" ref="canvasScrollRef">
        <!-- Raw TOML View Overlay if toggled -->
        <div v-if="showRawTomlModal" class="raw-toml-view glass-card">
          <div class="raw-toml-header">
            <div class="toml-title-group">
              <span class="font-mono text-cyan">{{ activeScope === 'user' ? (t('settings.user_toml') || '用户配置') : (t('settings.workspace_toml') || '工作区配置') }}</span>
              <span class="toml-path font-mono">{{ activeScope === 'user' ? configData?.user?.path : configData?.workspace?.path }}</span>
            </div>
            <div class="toml-actions">
              <button class="btn-copy-toml" @click="copyRawToml">
                {{ isCopied ? ('✓ ' + (t('settings.copied') || '已复制')) : (t('settings.copy_content') || '复制内容') }}
              </button>
              <button class="btn-close-toml" @click="showRawTomlModal = false">✕</button>
            </div>
          </div>
          <pre class="raw-toml-pre"><code>{{ currentRawToml || (t('settings.empty_toml') || '# 配置文件尚无自定义内容') }}</code></pre>
        </div>

        <!-- 1. General Settings -->
        <section
          v-show="isCategoryVisible('general')"
          id="section-general"
          class="settings-category-section glass-card"
        >
          <div class="section-header">
            <div class="section-title-wrap">
              <span class="section-icon">⚙️</span>
              <div>
                <h2>{{ t('settings.cat_general') || '常规' }}</h2>
                <span class="section-sub">{{ t('settings.cat_general_sub') || '语言、主题、浏览器、缓存' }}</span>
              </div>
            </div>
          </div>

          <div class="settings-list">
            <!-- Setting: Language -->
            <div class="setting-item" :class="{ highlight: matchesSearch('language') }">
              <div class="setting-meta">
                <span class="setting-title">{{ t('settings.language') || '语言' }}</span>
                <span class="setting-key font-mono">settings.language</span>
                <p class="setting-desc">{{ t('settings.language_desc') || '工作台界面的显示语言。' }}</p>
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
                <span class="setting-title">{{ t('settings.theme') || '主题' }}</span>
                <span class="setting-key font-mono">settings.theme</span>
                <p class="setting-desc">{{ t('settings.theme_desc') || '工作台外观色彩主题。' }}</p>
              </div>
              <div class="setting-control">
                <CustomSelect
                  :model-value="appTheme"
                  :options="themeSelectOptions"
                  @change="onThemeChange"
                />
              </div>
            </div>

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
                  <span class="setting-title">{{ t('settings.browser_title') || '浏览器' }}</span>
                  <span v-if="activeScope === 'workspace'" class="badge-global-only">
                    {{ t('settings.scope_global_only') || '仅限用户' }}
                  </span>
                </div>
                <span class="setting-key font-mono">settings.browser</span>
                <p class="setting-desc">
                  {{ t('settings.browser_desc') || '启动工作台时调用的浏览器。' }}
                </p>

                <!-- Workspace Scope Warning Banner -->
                <div v-if="activeScope === 'workspace'" class="scope-notice-banner warning">
                  <div class="notice-icon">⚠️</div>
                  <div class="notice-body">
                    <strong>{{ t('settings.scope_global_only_banner') || '无法在工作区中配置浏览器' }}</strong>
                    <p>
                      {{ t('settings.scope_global_only_desc') || '浏览器偏好属于个人系统环境，无法在工作区配置，以避免污染团队共享环境。' }}
                    </p>
                    <button class="btn-switch-scope" @click="switchScope('user', 'settings.browser')">
                      {{ t('settings.switch_to_user') || '在用户中编辑' }} ↗
                    </button>
                  </div>
                </div>
              </div>

              <!-- Browser Controls -->
              <div class="setting-control">
                <div v-if="activeScope === 'workspace'" class="control-disabled-notice">
                  <span class="font-mono text-muted">{{ t('settings.disabled_in_workspace') || '在工作区中禁用' }}</span>
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
                      :placeholder="t('settings.browser_custom_placeholder') || '/usr/bin/google-chrome-stable 或自定义路径'"
                      @change="saveSetting('settings.browser', customBrowserPath)"
                      @blur="saveSetting('settings.browser', customBrowserPath)"
                      @keydown.enter="saveSetting('settings.browser', customBrowserPath)"
                    />
                  </div>
                </div>
              </div>
            </div>

            <!-- Setting: Cache TTL -->
            <div class="setting-item" :class="{ highlight: matchesSearch('cache_ttl') }">
              <div class="setting-meta">
                <span class="setting-title">{{ t('settings.cache_ttl') || '缓存周期' }}</span>
                <span class="setting-key font-mono">settings.cache_ttl</span>
                <p class="setting-desc">{{ t('settings.cache_ttl_desc') || '工具清单与索引的本地缓存有效时间。' }}</p>
              </div>
              <div class="setting-control">
                <input
                  v-model="cacheTtl"
                  type="text"
                  class="setting-text-input font-mono"
                  placeholder="24h"
                  @change="saveSetting('settings.cache_ttl', cacheTtl)"
                  @blur="saveSetting('settings.cache_ttl', cacheTtl)"
                  @keydown.enter="saveSetting('settings.cache_ttl', cacheTtl)"
                />
              </div>
            </div>
          </div>
        </section>

        <!-- 3. Network & Mirrors -->
        <section
          v-show="isCategoryVisible('network')"
          id="section-network"
          class="settings-category-section glass-card"
        >
          <div class="section-header">
            <div class="section-title-wrap">
              <span class="section-icon">🌐</span>
              <div>
                <h2>{{ t('settings.cat_network') || '网络' }}</h2>
                <span class="section-sub">{{ t('settings.cat_network_sub') || '镜像加速与网络代理' }}</span>
              </div>
            </div>
            <div class="section-header-actions">
              <button class="btn-batch-test" :disabled="isProbing" @click="handleRunProbeAll">
                <span>{{ isProbing ? '🔄 ' + (t('network.testing') || '测速中...') : '⚡ ' + (t('network.batch_test') || '全生态并发测速') }}</span>
              </button>
            </div>
          </div>

          <!-- Section A: Network Proxy (Default strictly disabled, empty, no 3rd-party github proxies) -->
          <div class="network-subsection">
            <div class="subsection-title-bar">
              <span class="sub-icon">🔒</span>
              <h3>{{ t('network.tab_proxy') || '网络代理' }}</h3>
              <span class="proxy-badge" :class="networkStatus?.proxy?.enabled ? 'badge-on' : 'badge-off'">
                {{ networkStatus?.proxy?.enabled ? (t('network.proxy_enabled') || '已启用') : (t('network.proxy_disabled') || '直连模式 (已关闭)') }}
              </span>
            </div>

            <!-- Proxy Security Callout -->
            <div class="security-callout">
              <span class="callout-icon">🛡️</span>
              <p>{{ t('network.proxy_security_notice') || '系统默认官方直连，严禁且不内置任何第三方公网代理或 GitHub 代理，保障源码与密钥绝对安全。' }}</p>
            </div>

            <div class="proxy-form-grid" v-if="networkStatus">
              <div class="proxy-toggle-row">
                <label class="toggle-switch-label">
                  <input
                    type="checkbox"
                    v-model="networkStatus.proxy.enabled"
                    @change="handleSaveProxy"
                  />
                  <span class="slider round"></span>
                </label>
                <span class="toggle-text">{{ networkStatus.proxy.enabled ? (t('network.proxy_enabled') || '全局代理已开启') : (t('network.proxy_disabled') || '全局代理已关闭 (直连)') }}</span>
              </div>

              <div class="proxy-fields" :class="{ 'fields-disabled': !networkStatus.proxy.enabled }">
                <div class="proxy-field-item">
                  <label>{{ t('network.proxy_http') || 'HTTP / HTTPS 代理地址' }}</label>
                  <input
                    v-model="networkStatus.proxy.http"
                    type="text"
                    class="setting-text-input font-mono"
                    placeholder="http://127.0.0.1:7890 (留空为不使用)"
                    :disabled="!networkStatus.proxy.enabled"
                    @blur="handleSaveProxy"
                    @keydown.enter="handleSaveProxy"
                  />
                </div>

                <div class="proxy-field-item">
                  <label>{{ t('network.proxy_socks5') || 'SOCKS5 代理地址' }}</label>
                  <input
                    v-model="networkStatus.proxy.socks5"
                    type="text"
                    class="setting-text-input font-mono"
                    placeholder="socks5://127.0.0.1:10808 (可选)"
                    :disabled="!networkStatus.proxy.enabled"
                    @blur="handleSaveProxy"
                    @keydown.enter="handleSaveProxy"
                  />
                </div>

                <div class="proxy-field-item full-width">
                  <label>{{ t('network.proxy_noproxy') || '直连白名单 (NO_PROXY)' }}</label>
                  <input
                    v-model="networkStatus.proxy.no_proxy"
                    type="text"
                    class="setting-text-input font-mono"
                    placeholder="localhost,127.0.0.1,::1,*.local,*.internal"
                    :disabled="!networkStatus.proxy.enabled"
                    @blur="handleSaveProxy"
                    @keydown.enter="handleSaveProxy"
                  />
                </div>
              </div>
            </div>
          </div>

          <!-- Section B: Profiles and Mirrors Matrix -->
          <div class="network-subsection">
            <div class="subsection-title-bar">
              <span class="sub-icon">📦</span>
              <h3>{{ t('network.tab_mirrors') || '镜像源矩阵' }}</h3>
              <div class="profile-actions-bar">
                <button
                  class="btn-profile-action"
                  @click="showCreateProfileModal = true"
                  :title="t('network.create_profile') || '新建档案'"
                >
                  ＋ {{ t('network.create_profile') || '新建档案' }}
                </button>
                <button
                  v-if="networkStatus?.active_profile?.builtin"
                  class="btn-profile-action secondary"
                  @click="handleResetBuiltinProfile(networkStatus.active_profile_id)"
                  :title="t('network.reset_factory') || '恢复出厂默认值'"
                >
                  ↺ {{ t('network.reset_factory') || '恢复出厂默认值' }}
                </button>
              </div>
            </div>

            <!-- Profile Cards -->
            <div class="profile-cards-grid" v-if="networkStatus">
              <div
                v-for="(p, pId) in networkStatus.profiles"
                :key="pId"
                class="profile-card"
                :class="{ active: pId === networkStatus.active_profile_id, builtin: p.builtin }"
                @click="handleUseProfile(String(pId))"
              >
                <div class="profile-card-header">
                  <span class="profile-name">{{ p.name }}</span>
                  <span v-if="pId === networkStatus.active_profile_id" class="active-tag">✔ {{ t('network.active_tag') || '生效中' }}</span>
                </div>
                <p class="profile-desc">{{ p.description }}</p>
                <div class="profile-footer">
                  <span class="profile-id font-mono">id: {{ pId }}</span>
                  <span v-if="p.builtin" class="pill-builtin">{{ t('network.tag_builtin') || '官方内置' }}</span>
                  <span v-else class="pill-custom">{{ t('network.tag_custom') || '用户自定义' }}</span>
                </div>
              </div>
            </div>

            <!-- Save as Modal -->
            <div v-if="showCreateProfileModal" class="create-profile-modal glass-card">
              <h4>{{ t('network.create_profile') || '创建新网络预设档案' }}</h4>
              <div class="modal-form">
                <input
                  v-model="newProfileId"
                  type="text"
                  class="setting-text-input font-mono"
                  placeholder="档案标识 ID (例如 my-company)"
                />
                <input
                  v-model="newProfileName"
                  type="text"
                  class="setting-text-input"
                  placeholder="展示名称 (例如 公司内网高速)"
                />
                <div class="modal-actions">
                  <button class="btn-modal-cancel" @click="showCreateProfileModal = false">取消</button>
                  <button class="btn-modal-confirm" :disabled="!newProfileId" @click="handleSaveAsProfile">创建并激活</button>
                </div>
              </div>
            </div>

            <!-- Ecosystem Registries Matrix Table -->
            <div class="registries-matrix" v-if="networkStatus">
              <div
                v-for="eco in ['npm', 'go', 'python', 'rust', 'ruby']"
                :key="eco"
                class="registry-matrix-row"
              >
                <div class="row-eco-info">
                  <span class="eco-badge font-mono">{{ eco.toUpperCase() }}</span>
                  <span
                    class="status-indicator-badge"
                    :class="networkStatus.effective[eco]?.is_overridden ? 'badge-override' : 'badge-inherited'"
                  >
                    {{ networkStatus.effective[eco]?.is_overridden ? (t('network.badge_overridden') || '单项自定义覆盖') : (t('network.badge_inherited') || '跟随档案预设') }}
                  </span>
                </div>

                <div class="row-controls">
                  <!-- Quick Preset Dropdown -->
                  <div class="preset-dropdown-wrap" v-if="recommendedMirrors[eco]">
                    <select
                      class="preset-select"
                      @change="handleSelectMirrorPreset(eco, ($event.target as HTMLSelectElement).value)"
                    >
                      <option value="">快速推荐镜像源...</option>
                      <option
                        v-for="m in recommendedMirrors[eco]"
                        :key="m.url"
                        :value="m.url"
                      >
                        {{ m.name }}
                      </option>
                    </select>
                  </div>

                  <!-- Text input for registry URL -->
                  <input
                    type="text"
                    class="setting-text-input font-mono registry-url-input"
                    :value="networkStatus.effective[eco]?.url"
                    @change="handleOverrideRegistry(eco, ($event.target as HTMLInputElement).value)"
                    @blur="handleOverrideRegistry(eco, ($event.target as HTMLInputElement).value)"
                  />

                  <!-- Reset Override Button if Overridden -->
                  <button
                    v-if="networkStatus.effective[eco]?.is_overridden"
                    class="btn-reset-override"
                    :title="t('network.restore_inherited') || '恢复跟随预设'"
                    @click="handleResetRegistry(eco)"
                  >
                    ↺ {{ t('network.restore_inherited') || '恢复跟随' }}
                  </button>
                </div>
              </div>
            </div>

            <!-- Latency Probe Results Table if run -->
            <div v-if="probeResults.length > 0" class="probe-results-card glass-card">
              <h4>{{ t('network.speed_test') || '并发网络延迟探测结果' }}</h4>
              <div class="probe-grid">
                <div
                  v-for="res in probeResults"
                  :key="res.endpoint"
                  class="probe-node-item"
                  :class="'node-' + res.status"
                >
                  <div class="node-header">
                    <span class="node-eco font-mono">{{ res.ecosystem.toUpperCase() }}</span>
                    <span class="node-name">{{ res.name }}</span>
                  </div>
                  <div class="node-rtt font-mono">
                    <span v-if="res.status !== 'timeout'">{{ res.rtt_ms }} ms</span>
                    <span v-else class="text-error">TIMEOUT</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>

        <!-- 4. Plugins & Extensions -->
        <section
          v-show="isCategoryVisible('plugins')"
          id="section-plugins"
          class="settings-category-section glass-card"
        >
          <div class="section-header">
            <div class="section-title-wrap">
              <span class="section-icon">🧩</span>
              <div>
                <h2>{{ t('settings.cat_plugins') || '扩展' }}</h2>
                <span class="section-sub">{{ t('settings.cat_plugins_sub') || '插件配置与隔离' }}</span>
              </div>
            </div>
          </div>

          <div class="settings-list">
            <!-- Explanation banner -->
            <div class="plugin-isolation-banner">
              <span class="banner-icon">🔒</span>
              <div class="banner-text">
                <strong>{{ t('settings.plugin_isolation_title') || '三层命名空间隔离准则' }}</strong>
                <p>
                  {{ t('settings.plugin_isolation_desc') || '所有第三方扩展配置严格隔离于 plugins.<plugin-id>.* 区域，严禁侵入核心保留表。' }}
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
                  <span class="plugin-keys-count">{{ Object.keys(p.config || {}).length }}</span>
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
                    {{ t('settings.empty_toml') || '暂无配置项' }}
                  </div>
                </div>
              </div>
            </div>

            <div v-else class="no-plugins-card">
              <span class="empty-icon">📦</span>
              <span>{{ t('settings.no_plugins') || '当前尚未声明第三方插件配置。可通过扩展市场或配置文件添加扩展。' }}</span>
            </div>
          </div>
        </section>

        <!-- 5. Security & Namespace -->
        <section
          v-show="isCategoryVisible('security')"
          id="section-security"
          class="settings-category-section glass-card"
        >
          <div class="section-header">
            <div class="section-title-wrap">
              <span class="section-icon">🔒</span>
              <div>
                <h2>{{ t('settings.cat_security') || '安全' }}</h2>
                <span class="section-sub">{{ t('settings.cat_security_sub') || '保留表与文件排他锁' }}</span>
              </div>
            </div>
          </div>

          <div class="security-architecture-grid">
            <div class="sec-card">
              <div class="sec-badge green">{{ t('settings.status_protected') || '已受保护' }}</div>
              <h3>{{ t('settings.sec_core_title') || '核心保留命名空间' }}</h3>
              <p>{{ t('settings.sec_core_desc') || '严格保留 [tools], [tasks], [format], [env], [settings]，任何第三方插件写入非法顶级字段均会被拦截。' }}</p>
            </div>

            <div class="sec-card">
              <div class="sec-badge cyan">{{ t('settings.status_active') || '活跃生效' }}</div>
              <h3>{{ t('settings.sec_lock_title') || '并发文件排他锁' }}</h3>
              <p>{{ t('settings.sec_lock_desc') || '写入配置时自动获取排他锁，防止多进程或编辑器并发保存损坏数据。' }}</p>
            </div>

            <div class="sec-card">
              <div class="sec-badge purple">{{ t('settings.status_ast_safe') || '格式安全' }}</div>
              <h3>{{ t('settings.sec_ast_title') || 'AST 注释与格式保留' }}</h3>
              <p>{{ t('settings.sec_ast_desc') || '采用语法树精准注入算法，更新配置绝不破坏原配置文件中的手工注释与空行排版。' }}</p>
            </div>
          </div>
        </section>
      </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from "vue";
import { t, selectedLangSetting, setLanguage, SUPPORTED_LANGUAGES } from "../i18n";
import { useTheme } from "../composables/useTheme";
import CustomSelect from "../components/CustomSelect.vue";

const props = defineProps<{
  initialScope?: "user" | "workspace";
  activeCategory?: string;
}>();

const emit = defineEmits<{
  (e: "select-category", catId: string): void;
}>();

const { currentTheme, applyTheme } = useTheme();

// State
const activeScope = ref<"user" | "workspace">(props.initialScope || "user");
const searchQuery = ref("");
const activeCategory = ref(props.activeCategory || "general");
const showRawTomlModal = ref(false);
const isCopied = ref(false);

watch(
  () => props.activeCategory,
  (newCat) => {
    if (newCat) {
      activeCategory.value = newCat;
      scrollToCategory(newCat);
    }
  }
);

const isSaving = ref(false);
const saveStatusText = ref("");

// Setting fields
const appLanguage = ref(selectedLangSetting.value || "auto");
const appTheme = ref(currentTheme.value || "system");

watch(currentTheme, (newTheme) => {
  if (newTheme && appTheme.value !== newTheme) {
    appTheme.value = newTheme;
  }
});

watch(selectedLangSetting, (newLang) => {
  if (newLang && appLanguage.value !== newLang) {
    appLanguage.value = newLang;
  }
});

const cacheTtl = ref("24h");
const selectedBrowserPreset = ref("");
const customBrowserPath = ref("");
const githubProxy = ref("");
const httpProxy = ref("");


// Network matrix and profile state
const networkStatus = ref<any>(null);
const recommendedMirrors = ref<any>({});
const isProbing = ref(false);
const probeResults = ref<any[]>([]);
const showCreateProfileModal = ref(false);
const newProfileId = ref("");
const newProfileName = ref("");

// Full config from API
const configData = ref<any>(null);


const languageSelectOptions = computed(() => [
  { value: "auto", label: "🌐 " + (t("common.autoDetect") || "自动检测") },
  ...SUPPORTED_LANGUAGES.map((item) => ({
    value: item.code,
    label: item.nativeName,
  })),
]);

const themeSelectOptions = computed(() => [
  { value: "dark", label: "🌙 " + (t("theme.dark") || "深色模式") },
  { value: "light", label: "☀️ " + (t("theme.light") || "浅色模式") },
  { value: "system", label: "💻 " + (t("theme.system") || "跟随系统") },
]);

interface DetectedBrowser {
  id: string;
  name: string;
  path: string;
  is_default?: boolean;
}

function detectClientBrowsers(): DetectedBrowser[] {
  if (typeof navigator === "undefined") return [];

  const ua = navigator.userAgent || "";
  const isMac = /Macintosh|Mac OS X/i.test(ua);
  const isWindows = /Windows NT/i.test(ua);
  const detected: DetectedBrowser[] = [];

  // Detect currently active browser from User Agent
  if (/Edg\//i.test(ua)) {
    detected.push({ id: "msedge", name: "Microsoft Edge", path: "msedge", is_default: false });
  } else if (/Chrome\//i.test(ua) && !/Chromium|Edg|OPR|Brave/i.test(ua)) {
    detected.push({ id: "chrome", name: "Google Chrome", path: "chrome", is_default: false });
  } else if (/Firefox\//i.test(ua)) {
    detected.push({ id: "firefox", name: "Mozilla Firefox", path: "firefox", is_default: false });
  } else if (/Safari\//i.test(ua) && !/Chrome|Chromium|Edg/i.test(ua)) {
    detected.push({ id: "safari", name: "Apple Safari", path: "safari", is_default: false });
  }

  // Detect OS platform native built-ins
  if (isMac && !detected.some((b) => b.id === "safari")) {
    detected.push({ id: "safari", name: "Apple Safari", path: "safari", is_default: false });
  } else if (isWindows && !detected.some((b) => b.id === "msedge")) {
    detected.push({ id: "msedge", name: "Microsoft Edge", path: "msedge", is_default: false });
  }

  return detected;
}

function mergeDetectedBrowsers(serverList?: DetectedBrowser[]): DetectedBrowser[] {
  const map = new Map<string, DetectedBrowser>();

  // If server returned browsers (e.g. local desktop unirtm execution), add them
  if (serverList && serverList.length > 0) {
    for (const b of serverList) {
      map.set(b.id.toLowerCase(), b);
    }
  }

  // Merge client-side detected browsers (vital for web UI / remote SSH sessions)
  const clientList = detectClientBrowsers();
  for (const b of clientList) {
    if (!map.has(b.id.toLowerCase())) {
      map.set(b.id.toLowerCase(), b);
    }
  }

  return Array.from(map.values());
}

const detectedBrowsers = ref<DetectedBrowser[]>(mergeDetectedBrowsers());

function getBrowserIcon(id: string): string {
  switch (id.toLowerCase()) {
    case "chrome": return "🌐";
    case "firefox": return "🦊";
    case "msedge": case "edge": return "🌊";
    case "safari": return "🧭";
    case "brave": return "🦁";
    case "chromium": return "⚡";
    case "arc": return "🎨";
    case "opera": return "⭕";
    case "vivaldi": return "🔺";
    default: return "💻";
  }
}

const browserPresetOptions = computed(() => {
  const options: Array<{ value: string; label: string }> = [
    { value: "", label: "🖥️ " + (t("settings.browser_auto") || "系统默认") },
  ];

  for (const b of detectedBrowsers.value) {
    const icon = getBrowserIcon(b.id);
    options.push({
      value: b.id,
      label: `${icon} ${b.name}`,
    });
  }

  options.push({
    value: "custom",
    label: "🛠️ " + (t("settings.browser_custom") || "自定义路径"),
  });

  return options;
});

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


function matchesSearch(keyword: string) {
  if (!searchQuery.value) return false;
  return keyword.toLowerCase().includes(searchQuery.value.toLowerCase().trim());
}

function isCategoryVisible(catId: string) {
  if (!searchQuery.value) return true;
  const q = searchQuery.value.toLowerCase().trim();
  if (
    catId === "general" &&
    (q.includes("lang") ||
      q.includes("theme") ||
      q.includes("cache") ||
      q.includes("browser") ||
      q.includes("常规") ||
      q.includes("常用") ||
      q.includes("语言") ||
      q.includes("主题") ||
      q.includes("浏览器") ||
      q.includes("缓存"))
  )
    return true;
  if (
    catId === "network" &&
    (q.includes("proxy") ||
      q.includes("github") ||
      q.includes("mirror") ||
      q.includes("net") ||
      q.includes("网络") ||
      q.includes("代理") ||
      q.includes("镜像"))
  )
    return true;
  if (
    catId === "plugins" &&
    (q.includes("plugin") ||
      q.includes("ext") ||
      q.includes("docker") ||
      q.includes("插件") ||
      q.includes("扩展"))
  )
    return true;
  if (
    catId === "security" &&
    (q.includes("sec") ||
      q.includes("lock") ||
      q.includes("ast") ||
      q.includes("namespace") ||
      q.includes("safe") ||
      q.includes("安全") ||
      q.includes("锁") ||
      q.includes("命名空间") ||
      q.includes("保留"))
  )
    return true;
  return false;
}

function scrollToCategory(catId: string) {
  activeCategory.value = catId;
  emit("select-category", catId);
  const target = document.getElementById(`section-${catId}`);
  if (target) {
    target.scrollIntoView({ behavior: "smooth", block: "start" });
  }
}

function switchScope(scope: "user" | "workspace", scrollToId?: string) {
  activeScope.value = scope;
  syncFieldsFromConfig();
  loadNetworkStatus();
  if (scrollToId) {
    setTimeout(() => {
      scrollToCategory("general");
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

  if (configData.value?.detected_browsers) {
    detectedBrowsers.value = mergeDetectedBrowsers(configData.value.detected_browsers);
  }

  cacheTtl.value = s.cache_ttl || "24h";
  githubProxy.value = s.github_proxy || "";
  httpProxy.value = s.http_proxy || "";
  if (s.theme) {
    appTheme.value = s.theme;
  }
  if (s.language) {
    appLanguage.value = s.language;
  }

  if (activeScope.value === "user") {
    const b = s.browser || "";
    const knownIds = ["chrome", "firefox", "msedge", "safari", "brave", "chromium", "arc", "opera", "vivaldi", "epiphany", ""];
    const isDetectedOrKnown = knownIds.includes(b.toLowerCase()) || detectedBrowsers.value.some(d => d.id === b.toLowerCase());
    if (isDetectedOrKnown) {
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


function copyRawToml() {
  if (currentRawToml.value) {
    navigator.clipboard.writeText(currentRawToml.value);
    isCopied.value = true;
    setTimeout(() => {
      isCopied.value = false;
    }, 2000);
  }
}

async function loadNetworkStatus() {
  try {
    const res = await fetch(`/api/network/status?scope=${activeScope.value}`);
    if (res.ok) {
      const data = await res.json();
      networkStatus.value = data.status;
      recommendedMirrors.value = data.recommended || {};
    }
  } catch (err) {
    console.warn("Failed to load network status:", err);
  }
}

async function handleUseProfile(id: string) {
  try {
    const res = await fetch("/api/network/profile/use", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ profile_id: id, scope: activeScope.value }),
    });
    if (res.ok) {
      await loadNetworkStatus();
      loadConfigData();
    }
  } catch (err) {
    console.error(err);
  }
}

async function handleSaveAsProfile() {
  if (!newProfileId.value) return;
  try {
    const res = await fetch("/api/network/profile/save", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        profile_id: newProfileId.value,
        name: newProfileName.value || newProfileId.value,
        scope: activeScope.value,
      }),
    });
    if (res.ok) {
      showCreateProfileModal.value = false;
      newProfileId.value = "";
      newProfileName.value = "";
      await loadNetworkStatus();
      loadConfigData();
    }
  } catch (err) {
    console.error(err);
  }
}

async function handleResetBuiltinProfile(id: string) {
  try {
    const res = await fetch("/api/network/profile/reset", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ profile_id: id, scope: activeScope.value }),
    });
    if (res.ok) {
      await loadNetworkStatus();
      loadConfigData();
    }
  } catch (err) {
    console.error(err);
  }
}

async function handleOverrideRegistry(ecosystem: string, url: string) {
  try {
    const res = await fetch("/api/network/registry/override", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ ecosystem, url, scope: activeScope.value }),
    });
    if (res.ok) {
      await loadNetworkStatus();
      loadConfigData();
    }
  } catch (err) {
    console.error(err);
  }
}

function handleSelectMirrorPreset(ecosystem: string, url: string) {
  if (!url) return;
  handleOverrideRegistry(ecosystem, url);
}

async function handleResetRegistry(ecosystem: string) {
  try {
    const res = await fetch("/api/network/registry/reset", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ ecosystem, scope: activeScope.value }),
    });
    if (res.ok) {
      await loadNetworkStatus();
      loadConfigData();
    }
  } catch (err) {
    console.error(err);
  }
}

async function handleSaveProxy() {
  if (!networkStatus.value) return;
  try {
    const res = await fetch("/api/network/proxy", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        ...networkStatus.value.proxy,
        scope: activeScope.value,
      }),
    });
    if (res.ok) {
      await loadNetworkStatus();
      loadConfigData();
    }
  } catch (err) {
    console.error(err);
  }
}

async function handleRunProbeAll() {
  isProbing.value = true;
  try {
    const res = await fetch("/api/network/ping");
    if (res.ok) {
      const data = await res.json();
      probeResults.value = data.results || [];
    }
  } catch (err) {
    console.error(err);
  } finally {
    isProbing.value = false;
  }
}

onMounted(() => {
  loadConfigData();
  loadNetworkStatus();
  if (props.activeCategory) {
    setTimeout(() => {
      scrollToCategory(props.activeCategory!);
    }, 150);
  }
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

/* Scope Switcher & Path Badge */
.scope-switcher-group {
  display: flex;
  align-items: center;
  gap: 12px;
}

.active-scope-badge {
  font-size: 0.72rem;
  color: var(--text-muted);
  background: var(--input-bg);
  padding: 4px 10px;
  border-radius: 4px;
  border: 1px solid var(--card-border);
  max-width: 320px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
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

.badge-detected {
  font-size: 0.72rem;
  padding: 2px 8px;
  border-radius: 4px;
  background: rgba(16, 185, 129, 0.15);
  color: var(--success);
  border: 1px solid rgba(16, 185, 129, 0.3);
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

/* ─── Network & Mirror Matrix Styles ─── */
.section-header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.btn-batch-test {
  padding: 6px 14px;
  background: rgba(34, 211, 238, 0.12);
  border: 1px solid rgba(34, 211, 238, 0.35);
  color: var(--accent-cyan);
  border-radius: 6px;
  font-size: 0.8rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-batch-test:hover:not(:disabled) {
  background: rgba(34, 211, 238, 0.25);
  transform: translateY(-1px);
}

.btn-batch-test:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.network-subsection {
  display: flex;
  flex-direction: column;
  gap: 14px;
  padding: 18px 20px;
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid var(--card-border);
  border-radius: 10px;
  margin-top: 14px;
}

.subsection-title-bar {
  display: flex;
  align-items: center;
  gap: 10px;
}

.subsection-title-bar h3 {
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-main);
  margin: 0;
}

.proxy-badge {
  font-size: 0.72rem;
  padding: 2px 8px;
  border-radius: 9999px;
  font-weight: 600;
}

.badge-on {
  background: rgba(16, 185, 129, 0.18);
  color: #10b981;
  border: 1px solid rgba(16, 185, 129, 0.35);
}

.badge-off {
  background: rgba(148, 163, 184, 0.12);
  color: var(--text-muted);
  border: 1px solid var(--card-border);
}

.security-callout {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 10px 14px;
  background: rgba(56, 189, 248, 0.06);
  border-left: 3px solid var(--accent-cyan);
  border-radius: 4px;
}

.security-callout p {
  margin: 0;
  font-size: 0.78rem;
  color: var(--text-muted);
  line-height: 1.5;
}

.proxy-toggle-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.toggle-switch-label {
  position: relative;
  display: inline-block;
  width: 42px;
  height: 22px;
}

.toggle-switch-label input {
  opacity: 0;
  width: 0;
  height: 0;
}

.slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: var(--card-border);
  transition: 0.25s;
}

.slider.round {
  border-radius: 22px;
}

.slider.round:before {
  position: absolute;
  content: "";
  height: 16px;
  width: 16px;
  left: 3px;
  bottom: 3px;
  background-color: white;
  transition: 0.25s;
  border-radius: 50%;
}

input:checked + .slider {
  background-color: var(--accent-cyan);
}

input:checked + .slider:before {
  transform: translateX(20px);
}

.toggle-text {
  font-size: 0.84rem;
  font-weight: 500;
  color: var(--text-main);
}

.proxy-fields {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px;
  margin-top: 10px;
}

.proxy-fields.fields-disabled {
  opacity: 0.45;
  pointer-events: none;
}

.proxy-field-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.proxy-field-item.full-width {
  grid-column: 1 / -1;
}

.proxy-field-item label {
  font-size: 0.78rem;
  color: var(--text-muted);
  font-weight: 500;
}

.profile-actions-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-left: auto;
}

.btn-profile-action {
  padding: 4px 10px;
  background: rgba(34, 211, 238, 0.1);
  border: 1px solid rgba(34, 211, 238, 0.3);
  color: var(--accent-cyan);
  border-radius: 5px;
  font-size: 0.76rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-profile-action:hover {
  background: rgba(34, 211, 238, 0.2);
}

.btn-profile-action.secondary {
  background: var(--input-bg);
  border-color: var(--card-border);
  color: var(--text-muted);
}

.profile-cards-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 12px;
}

.profile-card {
  padding: 12px 14px;
  border-radius: 8px;
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  cursor: pointer;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.profile-card:hover {
  border-color: rgba(34, 211, 238, 0.4);
  transform: translateY(-1px);
}

.profile-card.active {
  border-color: var(--accent-cyan);
  background: rgba(34, 211, 238, 0.06);
  box-shadow: 0 0 12px rgba(34, 211, 238, 0.15);
}

.profile-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.profile-name {
  font-size: 0.88rem;
  font-weight: 600;
  color: var(--text-main);
}

.active-tag {
  font-size: 0.68rem;
  padding: 1px 6px;
  border-radius: 9999px;
  background: rgba(34, 211, 238, 0.2);
  color: var(--accent-cyan);
  font-weight: 600;
}

.profile-desc {
  font-size: 0.74rem;
  color: var(--text-muted);
  margin: 0;
  line-height: 1.4;
  flex: 1;
}

.profile-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 4px;
}

.profile-id {
  font-size: 0.68rem;
  color: var(--text-muted);
}

.pill-builtin, .pill-custom {
  font-size: 0.65rem;
  padding: 1px 5px;
  border-radius: 3px;
  font-weight: 500;
}

.pill-builtin {
  background: rgba(59, 130, 246, 0.15);
  color: #60a5fa;
}

.pill-custom {
  background: rgba(168, 85, 247, 0.15);
  color: #c084fc;
}

.create-profile-modal {
  padding: 14px 18px;
  border-radius: 8px;
  background: #090e1a;
  border: 1px solid var(--accent-cyan);
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.create-profile-modal h4 {
  margin: 0;
  font-size: 0.9rem;
  color: var(--accent-cyan);
}

.modal-form {
  display: flex;
  gap: 10px;
  align-items: center;
}

.modal-actions {
  display: flex;
  gap: 6px;
}

.btn-modal-cancel {
  padding: 6px 12px;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  color: var(--text-muted);
  border-radius: 5px;
  cursor: pointer;
  font-size: 0.78rem;
}

.btn-modal-confirm {
  padding: 6px 14px;
  background: var(--accent-cyan);
  color: #0b1120;
  font-weight: 600;
  border: none;
  border-radius: 5px;
  cursor: pointer;
  font-size: 0.78rem;
}

.btn-modal-confirm:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.registries-matrix {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-top: 6px;
}

.registry-matrix-row {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 10px 14px;
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  border-radius: 7px;
}

.row-eco-info {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 180px;
}

.eco-badge {
  font-size: 0.74rem;
  font-weight: 700;
  padding: 3px 8px;
  background: rgba(255, 255, 255, 0.06);
  border-radius: 4px;
  color: var(--text-main);
}

.status-indicator-badge {
  font-size: 0.68rem;
  padding: 2px 6px;
  border-radius: 4px;
  font-weight: 600;
}

.badge-override {
  background: rgba(245, 158, 11, 0.15);
  color: #f59e0b;
  border: 1px solid rgba(245, 158, 11, 0.35);
}

.badge-inherited {
  background: rgba(16, 185, 129, 0.12);
  color: #10b981;
}

.row-controls {
  display: flex;
  align-items: center;
  gap: 10px;
  flex: 1;
}

.preset-dropdown-wrap {
  width: 170px;
}

.preset-select {
  width: 100%;
  padding: 6px 8px;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  border-radius: 5px;
  color: var(--text-main);
  font-size: 0.76rem;
  outline: none;
}

.registry-url-input {
  flex: 1;
  font-size: 0.8rem;
  padding: 6px 10px;
}

.btn-reset-override {
  padding: 4px 10px;
  background: rgba(245, 158, 11, 0.12);
  border: 1px solid rgba(245, 158, 11, 0.3);
  color: #f59e0b;
  border-radius: 4px;
  font-size: 0.74rem;
  font-weight: 600;
  cursor: pointer;
  white-space: nowrap;
}

.btn-reset-override:hover {
  background: rgba(245, 158, 11, 0.25);
}

.probe-results-card {
  margin-top: 14px;
  padding: 14px 18px;
}

.probe-results-card h4 {
  margin: 0 0 10px;
  font-size: 0.88rem;
  color: var(--text-main);
}

.probe-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(210px, 1fr));
  gap: 8px;
}

.probe-node-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  border-radius: 6px;
  background: var(--card-bg);
  border: 1px solid var(--card-border);
}

.node-eco {
  font-size: 0.68rem;
  color: var(--text-muted);
  margin-right: 6px;
}

.node-name {
  font-size: 0.78rem;
  font-weight: 500;
  color: var(--text-main);
}

.node-rtt {
  font-size: 0.78rem;
  font-weight: 700;
}

.node-fast .node-rtt {
  color: #10b981;
}

.node-normal .node-rtt {
  color: #38bdf8;
}

.node-slow .node-rtt {
  color: #f59e0b;
}

.node-timeout .node-rtt {
  color: #ef4444;
}
</style>
