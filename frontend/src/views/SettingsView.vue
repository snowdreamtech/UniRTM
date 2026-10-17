<template>
  <div class="view-container settings-view" id="view-settings">
    <!-- Top Settings Header Bar (VS Code Style) -->
    <div class="settings-top-bar glass-card">
      <div class="settings-search-wrapper">
        <span class="search-icon">
          <TechIcon id="workbench:doctor" :size="14" :colored="false" />
        </span>
        <input
          v-model="searchQuery"
          type="text"
          class="settings-search-input"
          :placeholder="t('settings.search_placeholder') || '搜索设置 (Search settings)...'"
        />
        <button v-if="searchQuery" class="clear-search-btn" @click="searchQuery = ''" title="清除">✕</button>
      </div>

      <!-- Scope Context Info: User vs Workspace -->
      <div class="scope-switcher-group">
        <span
          class="active-scope-badge font-mono"
          :title="activeScope === 'user' ? (configData?.user?.path || '~/.config/unirtm/unirtm.toml') : (configData?.workspace?.path || './.unirtm.toml')"
        >
          <span style="font-weight: 600; color: var(--accent-cyan); margin-right: 6px;">
            [{{ activeScope === 'user' ? (t('settings.scope_user') || '用户全局配置') : (t('settings.scope_workspace') || '工作区配置') }}]
          </span>
          {{ activeScope === 'user' ? (configData?.user?.path || '~/.config/unirtm/unirtm.toml') : (configData?.workspace?.path || './.unirtm.toml') }}
        </span>
      </div>

      <!-- Top Right Actions: Auto-save status & Open TOML -->
      <div class="settings-top-actions">
        <span class="auto-save-pill" :class="[saveStatusType, { saving: isSaving }]" style="display: inline-flex; align-items: center; gap: 6px;">
          <TechIcon
            :id="saveStatusType === 'saving' ? 'workbench:sync' : saveStatusType === 'success' ? 'workbench:success' : saveStatusType === 'error' ? 'workbench:error' : 'workbench:speed'"
            :size="12"
            :class="{ spinning: saveStatusType === 'saving' }"
          />
          <span>{{ saveStatusText || t('settings.realtime_save') || '修改已实时生效' }}</span>
        </span>
        <button
          class="btn-raw-toml"
          :class="{ active: showRawTomlModal }"
          @click="showRawTomlModal = !showRawTomlModal"
          :title="t('settings.open_toml') || '打开 TOML'"
        >
          <TechIcon id="workbench:edit" :size="14" />
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
              <span class="section-icon">
                <TechIcon id="workbench:settings" :size="20" />
              </span>
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

            <!-- Setting: Appearance Mode -->
            <div class="setting-item" :class="{ highlight: matchesSearch('appearance') }">
              <div class="setting-meta">
                <span class="setting-title">{{ t('settings.appearance_mode') || '外观模式' }}</span>
                <p class="setting-desc">{{ t('settings.appearance_mode_desc') || '工作台界面的明暗模式，可跟随操作系统或手动指定。' }}</p>
              </div>
              <div class="setting-control">
                <div class="segmented-control" id="appearance-segmented-control">
                  <button
                    type="button"
                    class="segment-btn"
                    :class="{ active: activeAppearance === 'system' }"
                    @click="onAppearanceChange('system')"
                    id="btn-appearance-system"
                  >
                    <TechIcon id="workbench:dashboard" :size="14" :colored="false" />
                    <span>{{ t('theme.system') || '跟随系统' }}</span>
                  </button>
                  <button
                    type="button"
                    class="segment-btn"
                    :class="{ active: activeAppearance === 'light' }"
                    @click="onAppearanceChange('light')"
                    id="btn-appearance-light"
                  >
                    <TechIcon id="workbench:about" :size="14" :colored="false" />
                    <span>{{ t('theme.light') || '浅色模式' }}</span>
                  </button>
                  <button
                    type="button"
                    class="segment-btn"
                    :class="{ active: activeAppearance === 'dark' }"
                    @click="onAppearanceChange('dark')"
                    id="btn-appearance-dark"
                  >
                    <TechIcon id="workbench:terminal" :size="14" :colored="false" />
                    <span>{{ t('theme.dark') || '深色模式' }}</span>
                  </button>
                </div>
              </div>
            </div>

            <!-- Setting: Color Theme -->
            <div class="setting-item" :class="{ highlight: matchesSearch('theme') }">
              <div class="setting-meta">
                <span class="setting-title">{{ t('settings.color_theme') || '配色主题' }}</span>
                <p class="setting-desc">{{ t('settings.color_theme_desc') || '工作台界面配色主题，完全兼容 VS Code 主题标准。' }}</p>
              </div>
              <div class="setting-control">
                <CustomSelect
                  :model-value="activeColorTheme"
                  :options="colorThemeOptions"
                  @change="onColorThemeChange"
                />
              </div>
            </div>

            <!-- Setting: Icon Theme -->
            <div class="setting-item" :class="{ highlight: matchesSearch('icon') }">
              <div class="setting-meta">
                <span class="setting-title">{{ t('settings.icon_theme') || '文件图标主题' }}</span>
                <p class="setting-desc">{{ t('settings.icon_theme_desc') || '工作台侧边栏与文件/工具列表图标集，100% 对齐 VS Code 标准结构。' }}</p>
              </div>
              <div class="setting-control">
                <CustomSelect
                  :model-value="activeIconTheme"
                  :options="iconThemeOptions"
                  @change="onIconThemeChange"
                />
              </div>
            </div>

            <!-- Setting: Product Icon Theme -->
            <div class="setting-item" :class="{ highlight: matchesSearch('product_icon') || matchesSearch('icon') }">
              <div class="setting-meta">
                <span class="setting-title">{{ t('settings.product_icon_theme') || '界面图标主题' }}</span>
                <p class="setting-desc">{{ t('settings.product_icon_theme_desc') || '工作台按钮、操作与状态等界面产品图标主题，支持切换 Codicons 与第三方界面图标扩展。' }}</p>
              </div>
              <div class="setting-control">
                <CustomSelect
                  :model-value="activeProductIconTheme"
                  :options="productIconThemeOptions"
                  @change="onProductIconThemeChange"
                />
              </div>
            </div>

            <!-- Setting: Theme Package Manager -->
            <div class="setting-item theme-manager-item" :class="{ highlight: matchesSearch('vsix') }">
              <div class="setting-meta full-width">
                <div class="title-with-badge">
                  <span class="setting-title">{{ t('settings.theme_pkg_manager') || 'VS Code 主题与图标扩展管理' }}</span>
                  <span class="badge-custom-pkg">VSIX / ZIP</span>
                </div>
                <p class="setting-desc">{{ t('settings.theme_pkg_desc') || '导入本地官方 VS Code 扩展包（.vsix），纯静态解析，自动保存在用户本地目录。' }}</p>

                <!-- Actions Toolbar -->
                <div class="theme-actions-toolbar">
                  <button class="btn-jump-marketplace" @click="navigateToMarketplace('themes')">
                    <span class="jump-icon">
                      <TechIcon id="workbench:marketplace" :size="14" />
                    </span>
                    <span>前往扩展与主题市场挑选... ↗</span>
                  </button>

                  <div class="local-vsix-upload-wrap">
                    <input
                      ref="vsixInputRef"
                      type="file"
                      accept=".vsix,.zip"
                      class="hidden-file-input"
                      @change="onVsixFileSelected"
                    />
                    <button
                      class="btn-install-vsix"
                      :disabled="isInstallingTheme"
                      @trigger="triggerVsixUpload"
                      @click="triggerVsixUpload"
                    >
                      <span v-if="isInstallingTheme" class="spinner-small"></span>
                      <span v-else class="btn-icon">
                        <TechIcon id="workbench:folder" :size="14" />
                      </span>
                      <span>{{ isInstallingTheme ? (t('settings.installing') || '正在解析安装...') : (t('settings.install_vsix') || '手动导入本地 .vsix 包') }}</span>
                    </button>
                  </div>
                </div>

                <!-- Feedback Alert in Installed Tab -->
                <div v-if="themeInstallMsg" class="install-feedback-banner" :class="{ error: themeInstallError }">
                  <span>{{ themeInstallMsg }}</span>
                  <button class="btn-close-msg" @click="themeInstallMsg = ''">✕</button>
                </div>
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
                <p class="setting-desc">
                  {{ t('settings.browser_desc') || '启动工作台时调用的浏览器。' }}
                </p>

                <!-- Workspace Scope Warning Banner -->
                <div v-if="activeScope === 'workspace'" class="scope-notice-banner warning">
                  <div class="notice-icon">
                    <TechIcon id="workbench:warning" :size="16" />
                  </div>
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

            <!-- Setting: Concurrency (Jobs) -->
            <div class="setting-item" :class="{ highlight: matchesSearch('jobs') || matchesSearch('concurrency') }">
              <div class="setting-meta">
                <span class="setting-title">{{ t('settings.jobs') || '并发线程数 (Jobs)' }}</span>
                <p class="setting-desc">{{ t('settings.jobs_desc') || '工具并发下载及任务并行执行的线程数量上限。设为 0 表示自适应系统 CPU 核心数。' }}</p>
              </div>
              <div class="setting-control">
                <input
                  v-model.number="jobs"
                  type="number"
                  min="0"
                  class="setting-text-input font-mono"
                  :placeholder="t('settings.jobs_auto') || '0 (自动调度)'"
                  @change="onJobsChange"
                  @blur="onJobsChange"
                  @keydown.enter="onJobsChange"
                />
              </div>
            </div>

            <!-- Setting: Auto Install Missing Tools -->
            <div class="setting-item" :class="{ highlight: matchesSearch('auto_install') || matchesSearch('install') }">
              <div class="setting-meta">
                <span class="setting-title">{{ t('settings.auto_install') || '自动补齐工具 (Auto Install)' }}</span>
                <p class="setting-desc">{{ t('settings.auto_install_desc') || '执行任务遇到未安装的环境依赖工具时，是否自动在后台静默拉取并安装。' }}</p>
              </div>
              <div class="setting-control">
                <label class="toggle-switch-label">
                  <input
                    type="checkbox"
                    :checked="autoInstall"
                    @change="onAutoInstallToggle"
                  />
                  <span class="slider round"></span>
                </label>
              </div>
            </div>

            <!-- Setting: Task Timeout -->
            <div class="setting-item" :class="{ highlight: matchesSearch('timeout') || matchesSearch('task_timeout') }">
              <div class="setting-meta">
                <span class="setting-title">{{ t('settings.task_timeout') || '任务超时时长 (Task Timeout)' }}</span>
                <p class="setting-desc">{{ t('settings.task_timeout_desc') || '任务运行的全局超时时间上限（例如 10m、30m 或 600s）。未设置或留空表示遵循任务自身定义。' }}</p>
              </div>
              <div class="setting-control">
                <input
                  v-model="taskTimeout"
                  type="text"
                  class="setting-text-input font-mono"
                  placeholder="例如: 10m, 600"
                  @change="onTaskTimeoutChange"
                  @blur="onTaskTimeoutChange"
                  @keydown.enter="onTaskTimeoutChange"
                />
              </div>
            </div>

            <!-- Setting: Task Output Mode -->
            <div class="setting-item" :class="{ highlight: matchesSearch('output') || matchesSearch('task_output') }">
              <div class="setting-meta">
                <span class="setting-title">{{ t('settings.task_output') || '任务日志输出模式 (Task Output)' }}</span>
                <p class="setting-desc">{{ t('settings.task_output_desc') || '多任务并发执行时的控制台日志汇总与渲染风格。' }}</p>
              </div>
              <div class="setting-control">
                <CustomSelect
                  :model-value="taskOutput"
                  :options="taskOutputOptions"
                  @change="onTaskOutputChange"
                />
              </div>
            </div>

            <!-- Setting: Minimum Release Age -->
            <div class="setting-item" :class="{ highlight: matchesSearch('age') || matchesSearch('minimum_release_age') || matchesSearch('security') }">
              <div class="setting-meta">
                <div class="title-with-badge">
                  <span class="setting-title">{{ t('settings.minimum_release_age') || '发布安全冷却时间 (Release Age)' }}</span>
                  <span class="badge-custom-pkg" style="background: rgba(16, 185, 129, 0.12); color: #10b981; border-color: rgba(16, 185, 129, 0.25);">
                    🛡️ 安全防护
                  </span>
                </div>
                <p class="setting-desc">{{ t('settings.minimum_release_age_desc') || '防范供应链投毒与零日漏洞。系统默认遵循 7 天安全冷却期（7d），延迟接纳刚发布的新版本。输入 0 可关闭冷却以立即获取绝对最新版本。' }}</p>
                <div class="release-age-presets-row" style="margin-top: 8px; display: inline-flex; gap: 6px; flex-wrap: wrap;">
                  <button
                    type="button"
                    class="preset-chip-btn"
                    :class="{ active: (minimumReleaseAge || '7d') === '7d' }"
                    @click="applyReleaseAgePreset('7d')"
                  >
                    <span>🛡️ {{ t('settings.release_age_preset_7d') || '7d (推荐默认)' }}</span>
                  </button>
                  <button
                    type="button"
                    class="preset-chip-btn"
                    :class="{ active: minimumReleaseAge === '14d' }"
                    @click="applyReleaseAgePreset('14d')"
                  >
                    <span>🔒 {{ t('settings.release_age_preset_14d') || '14d (严格保守)' }}</span>
                  </button>
                  <button
                    type="button"
                    class="preset-chip-btn"
                    :class="{ active: minimumReleaseAge === '0' }"
                    @click="applyReleaseAgePreset('0')"
                  >
                    <span>⚡ {{ t('settings.release_age_preset_0') || '0 (即刻获取)' }}</span>
                  </button>
                </div>
              </div>
              <div class="setting-control">
                <input
                  v-model="minimumReleaseAge"
                  type="text"
                  class="setting-text-input font-mono"
                  :placeholder="t('settings.minimum_release_age_placeholder') || '7d (系统默认推荐)'"
                  @change="onMinimumReleaseAgeChange"
                  @blur="onMinimumReleaseAgeChange"
                  @keydown.enter="onMinimumReleaseAgeChange"
                />
              </div>
            </div>

            <!-- Setting: Experimental Features -->
            <div class="setting-item" :class="{ highlight: matchesSearch('experimental') }">
              <div class="setting-meta">
                <div class="title-with-badge">
                  <span class="setting-title">{{ t('settings.experimental') || '实验性预览特性 (Experimental)' }}</span>
                  <span class="badge-custom-pkg" style="background: rgba(245, 158, 11, 0.12); color: #d97706; border-color: rgba(245, 158, 11, 0.25);">
                    🧪 预览阶段
                  </span>
                </div>
                <p class="setting-desc">{{ t('settings.experimental_desc') || '启用 UniRTM 处于早期开发预览阶段的实验性功能与架构能力。' }}</p>
              </div>
              <div class="setting-control">
                <label class="toggle-switch-label">
                  <input
                    type="checkbox"
                    :checked="experimental"
                    @change="onExperimentalToggle"
                  />
                  <span class="slider round"></span>
                </label>
              </div>
            </div>
          </div>
        </section>

        <!-- 2. Git & Lifecycle Hooks -->
        <section
          v-show="isCategoryVisible('hooks')"
          id="section-hooks"
          class="settings-category-section glass-card"
        >
          <div class="section-header">
            <div class="section-title-wrap">
              <span class="section-icon">
                <TechIcon id="workbench:git" :size="20" />
              </span>
              <div>
                <div class="title-with-badge" style="display: flex; align-items: center; gap: 8px;">
                  <h2>{{ t('settings.cat_hooks') || '生命周期与 Git 钩子 (Hooks)' }}</h2>
                  <span v-if="activeScope === 'user'" class="badge-global-only" style="background: rgba(6, 182, 212, 0.12); color: var(--accent-cyan); border-color: rgba(6, 182, 212, 0.25);">
                    {{ t('settings.scope_workspace_only') || '仅限工作区' }}
                  </span>
                </div>
                <span class="section-sub">{{ t('settings.cat_hooks_sub') || 'Git 提交与操作生命周期自动化事件脚本' }}</span>
              </div>
            </div>
            <div v-if="activeScope !== 'user'" class="hooks-header-actions" style="display: flex; gap: 8px; align-items: center;">
              <button
                type="button"
                class="btn-bridge-install"
                :disabled="isInstallingBridges"
                @click="installAllGitBridges"
                :title="t('settings.hooks_install_bridges_tip') || '自动将 UniRTM 路由脚本注入到工作区 .git/hooks 目录'"
              >
                <TechIcon id="workbench:sync" :size="13" :class="{ spinning: isInstallingBridges }" />
                <span>{{ isInstallingBridges ? (t('settings.hooks_installing_bridges') || '正在注入 Git 桥接...') : (t('settings.hooks_install_bridges') || '注入 Git 桥接') }}</span>
              </button>
              <button
                type="button"
                class="btn-add-hook"
                @click="openAddHookModal()"
              >
                <TechIcon id="workbench:add" :size="13" />
                <span>{{ t('settings.hooks_add_btn') || '添加钩子' }}</span>
              </button>
            </div>
          </div>

          <!-- Hooks Content List or Empty State -->
          <div class="hooks-container">
            <!-- User Scope Warning Banner -->
            <div v-if="activeScope === 'user'" class="scope-notice-banner warning" style="margin-top: 4px;">
              <div class="notice-icon">
                <TechIcon id="workbench:warning" :size="16" />
              </div>
              <div class="notice-body">
                <strong>{{ t('settings.hooks_workspace_only_banner') || 'Git 生命周期钩子仅限工作区生效' }}</strong>
                <p>
                  {{ t('settings.hooks_workspace_only_desc') || 'Git 钩子（如 pre-commit、commit-msg）直接与当前代码工程的 .git 目录及项目规范强绑定，无法在用户全局层级配置。' }}
                </p>
                <button class="btn-switch-scope" @click="switchScope('workspace', 'section-hooks')">
                  {{ t('settings.switch_to_workspace') || '切换到工作区配置' }} ↗
                </button>
              </div>
            </div>

            <div v-else-if="isLoadingHooks" class="hooks-loading">
              <span class="spinner-small"></span>
              <span>{{ t('settings.hooks_loading') || '正在加载生命周期钩子...' }}</span>
            </div>

            <div v-else-if="hooksList.length === 0" class="hooks-empty-state">
              <div class="empty-icon-wrap">
                <TechIcon id="workbench:git" :size="32" />
              </div>
              <h4>{{ t('settings.hooks_empty_title') || '尚未配置任何生命周期钩子' }}</h4>
              <p>{{ t('settings.hooks_empty_desc') || '配置 Git 钩子可在提交、推送或分支检出前后自动执行代码格式化、质量检查与测试。' }}</p>
              <div class="hooks-preset-recommendations">
                <span class="preset-label">{{ t('settings.hooks_recommended_label') || '常用推荐:' }}</span>
                <button
                  type="button"
                  class="preset-chip-btn"
                  @click="openAddHookModal('pre-commit', 'npm test')"
                >
                  <span>⚡ pre-commit (测试与质检)</span>
                </button>
                <button
                  type="button"
                  class="preset-chip-btn"
                  @click="openAddHookModal('commit-msg', 'npx commitlint --edit $1')"
                >
                  <span>📝 commit-msg (提交信息规范)</span>
                </button>
                <button
                  type="button"
                  class="preset-chip-btn"
                  @click="openAddHookModal('pre-push', 'npm run lint && npm test')"
                >
                  <span>🛡️ pre-push (推送前全量校验)</span>
                </button>
              </div>
            </div>

            <div v-else class="hooks-grid">
              <div
                v-for="hook in hooksList"
                :key="hook.name"
                class="hook-card"
                :class="{ 'highlight': matchesSearch(hook.name) || matchesSearch(hook.command) }"
              >
                <div class="hook-card-header">
                  <div class="hook-name-group">
                    <span class="hook-badge-icon">
                      <TechIcon id="workbench:git" :size="14" />
                    </span>
                    <span class="hook-name font-mono">{{ hook.name }}</span>
                    <span class="hook-bridge-badge" :class="hook.installed ? 'installed' : 'not-installed'">
                      {{ hook.installed ? ('✓ ' + (t('settings.hooks_badge_bridged') || 'Git 已桥接')) : ('⚠ ' + (t('settings.hooks_badge_unbridged') || '未注入桥接')) }}
                    </span>
                  </div>
                  <div class="hook-card-actions">
                    <button
                      type="button"
                      class="btn-icon-action"
                      :title="t('settings.hooks_run_tip') || '测试执行该钩子'"
                      @click="runHookTest(hook.name)"
                    >
                      <TechIcon id="workbench:tasks" :size="14" />
                      <span>{{ t('settings.hooks_run_btn') || '测试' }}</span>
                    </button>
                    <button
                      type="button"
                      class="btn-icon-action"
                      :title="t('settings.hooks_edit_tip') || '编辑脚本命令'"
                      @click="openEditHookModal(hook)"
                    >
                      <TechIcon id="workbench:edit" :size="14" />
                    </button>
                    <button
                      type="button"
                      class="btn-icon-action danger"
                      :title="t('settings.hooks_delete_tip') || '删除此钩子'"
                      @click="deleteHook(hook.name)"
                    >
                      <TechIcon id="workbench:error" :size="14" />
                    </button>
                  </div>
                </div>

                <div class="hook-command-wrap">
                  <div class="hook-command-code font-mono">
                    <code>{{ hook.command }}</code>
                  </div>
                  <button
                    type="button"
                    class="btn-copy-code"
                    :title="t('settings.copy_command') || '复制命令'"
                    @click="copyCommand(hook.command)"
                  >
                    <TechIcon id="workbench:edit" :size="12" />
                  </button>
                </div>
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
              <span class="section-icon">
                <TechIcon id="workbench:runtimes" :size="20" />
              </span>
              <div>
                <h2>{{ t('settings.cat_network') || '网络' }}</h2>
                <span class="section-sub">{{ t('settings.cat_network_sub') || '镜像加速与网络代理' }}</span>
              </div>
            </div>
          </div>

          <!-- Section A: Network Proxy (Default strictly disabled, empty, no 3rd-party github proxies) -->
          <div class="network-subsection">
            <div class="subsection-title-bar">
              <span class="sub-icon">
                <TechIcon id="workbench:lock" :size="16" />
              </span>
              <h3>{{ t('network.tab_proxy') || '网络代理' }}</h3>
              <span class="proxy-badge" :class="networkStatus?.proxy?.enabled ? 'badge-on' : 'badge-off'">
                {{ networkStatus?.proxy?.enabled ? (t('network.proxy_enabled') || '已启用') : (t('network.proxy_disabled') || '直连模式 (已关闭)') }}
              </span>
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

                <button
                  type="button"
                  class="btn-test-proxy"
                  :disabled="isTestingProxy || !networkStatus.proxy.enabled || (!networkStatus.proxy.http && !networkStatus.proxy.socks5)"
                  @click="handleTestProxy"
                  title="发起真实网络连接检测该代理是否可用"
                >
                  <span v-if="isTestingProxy" style="display: inline-flex; align-items: center; gap: 4px;">
                    <TechIcon id="workbench:sync" :size="13" class="spinning" />
                    <span>检测中...</span>
                  </span>
                  <span v-else style="display: inline-flex; align-items: center; gap: 4px;">
                    <TechIcon id="workbench:speed" :size="13" />
                    <span>连通性检测</span>
                  </span>
                </button>

                <button
                  type="button"
                  class="btn-detect-proxy"
                  :disabled="isDetectingProxy"
                  @click="handleDetectProxy(true)"
                  title="自动探测宿主机运行中的代理客户端 (sing-box, Clash 等)"
                >
                  <span v-if="isDetectingProxy" style="display: inline-flex; align-items: center; gap: 4px;">
                    <TechIcon id="workbench:sync" :size="13" class="spinning" />
                    <span>探测中...</span>
                  </span>
                  <span v-else style="display: inline-flex; align-items: center; gap: 4px;">
                    <TechIcon id="workbench:search" :size="13" />
                    <span>自动探测宿主代理</span>
                  </span>
                </button>
              </div>

              <!-- Detected Proxy Banner -->
              <div v-if="detectedProxy?.detected" class="detected-proxy-banner">
                <div class="detected-banner-left">
                  <span class="detected-pill">{{ detectedProxy.type.toUpperCase() }}</span>
                  <span class="detected-desc">{{ detectedProxy.description }}</span>
                </div>
                <button
                  type="button"
                  class="btn-apply-detected"
                  @click="applyDetectedProxy"
                >
                  <span style="display: inline-flex; align-items: center; gap: 4px;">
                    <TechIcon id="workbench:speed" :size="13" />
                    <span>一键导入</span>
                  </span>
                </button>
              </div>

              <!-- Proxy Test Result Status Banner -->
              <div
                v-if="proxyTestResult"
                class="proxy-test-status-banner"
                :class="{ 'status-ok': proxyTestResult.success, 'status-fail': !proxyTestResult.success }"
              >
                <div class="test-status-left">
                  <span class="status-icon">
                    <TechIcon :id="proxyTestResult.success ? 'workbench:success' : 'workbench:error'" :size="14" />
                  </span>
                  <span class="status-message">{{ proxyTestResult.message || proxyTestResult.error }}</span>
                </div>
                <div v-if="proxyTestResult.success && proxyTestResult.rtt_ms" class="test-rtt-tag" style="display: inline-flex; align-items: center; gap: 4px;">
                  <TechIcon id="workbench:speed" :size="12" />
                  <span>延迟 {{ proxyTestResult.rtt_ms }}ms</span>
                </div>
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
                    @change="handleSaveProxy"
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
                    @change="handleSaveProxy"
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
                    @change="handleSaveProxy"
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
              <span class="sub-icon">
                <TechIcon id="workbench:package" :size="16" />
              </span>
              <h3>{{ t('network.tab_mirrors') || '镜像源矩阵' }}</h3>
              <button
                class="btn-batch-test"
                :disabled="isProbing"
                @click="handleRunProbeAll"
                title="并发测试当前镜像源矩阵中所有生态的延迟"
              >
                <span v-if="isProbing" style="display: inline-flex; align-items: center; gap: 4px;">
                  <TechIcon id="workbench:sync" :size="13" class="spinning" />
                  <span>测速中...</span>
                </span>
                <span v-else style="display: inline-flex; align-items: center; gap: 4px;">
                  <TechIcon id="workbench:speed" :size="13" />
                  <span>一键测速</span>
                </span>
              </button>

              <div class="profile-actions-bar">
                <div class="profile-scroll-nav" v-if="orderedProfiles.length > 3">
                  <button class="btn-scroll-arrow" @click="scrollProfiles('left')" title="向左滚动">‹</button>
                  <button class="btn-scroll-arrow" @click="scrollProfiles('right')" title="向右滚动">›</button>
                </div>
                <button
                  class="btn-profile-action"
                  @click="showCreateProfileModal = true"
                  :title="t('network.create_profile') || '新建档案'"
                >
                  <TechIcon id="workbench:tools" :size="12" />
                  <span>{{ t('network.create_profile') || '新建档案' }}</span>
                </button>
                <button
                  class="btn-profile-action"
                  @click="openAddEcoModal()"
                  :title="t('network.add_ecosystem') || '添加生态'"
                >
                  <TechIcon id="workbench:runtimes" :size="12" />
                  <span>{{ t('network.add_ecosystem') || '添加生态' }}</span>
                </button>
                <button
                  class="btn-profile-action secondary"
                  @click="handleResetAllEcosystems()"
                  :title="t('network.reset_all_ecosystems') || '恢复默认生态'"
                >
                  <TechIcon id="workbench:sync" :size="12" />
                  <span>{{ t('network.reset_all_ecosystems') || '恢复默认生态' }}</span>
                </button>
                <button
                  v-if="networkStatus?.active_profile?.builtin"
                  class="btn-profile-action secondary"
                  @click="handleResetBuiltinProfile(networkStatus.active_profile_id)"
                  :title="t('network.reset_factory') || '恢复出厂默认值'"
                >
                  <TechIcon id="workbench:sync" :size="12" />
                  <span>{{ t('network.reset_factory') || '恢复出厂默认值' }}</span>
                </button>
              </div>
            </div>

            <!-- Profile Cards Track -->
            <div class="profile-cards-wrapper" v-if="networkStatus">
              <div
                ref="profileTrackRef"
                class="profile-cards-track"
                @wheel="onProfileTrackWheel"
              >
                <div
                  v-for="{ id: pId, profile: p } in orderedProfiles"
                  :key="pId"
                  class="profile-card"
                  :class="{ active: pId === networkStatus.active_profile_id, builtin: p.builtin }"
                  @click="handleUseProfile(String(pId))"
                >
                  <div class="profile-card-header">
                    <span class="profile-name">{{ getProfileDisplayName(String(pId), p.name) }}</span>
                    <span v-if="pId === networkStatus.active_profile_id" class="active-tag">✔ {{ t('network.active_tag') }}</span>
                  </div>
                  <p class="profile-desc">{{ getProfileDisplayDesc(String(pId), p.description) }}</p>
                  <div class="profile-footer">
                    <span class="profile-id font-mono">id: {{ pId }}</span>
                    <span v-if="p.builtin" class="pill-builtin">{{ t('network.tag_builtin') }}</span>
                    <span v-else class="pill-custom">{{ t('network.tag_custom') }}</span>
                  </div>
                </div>
              </div>
            </div>

            <!-- Create/Save Profile Modal Dialog -->
            <div v-if="showCreateProfileModal" class="modal-backdrop" @click.self="showCreateProfileModal = false">
              <div class="modal-dialog glass-card">
                <div class="modal-header">
                  <div class="modal-title-group">
                    <span class="modal-icon">
                      <TechIcon id="workbench:config" :size="20" />
                    </span>
                    <div>
                      <h4>{{ t('network.create_profile') || '新建网络预设档案' }}</h4>
                      <span class="modal-subtitle">将当前网络与各生态镜像配置保存为独立的 Profile 方案</span>
                    </div>
                  </div>
                  <button class="btn-modal-close" @click="showCreateProfileModal = false">✕</button>
                </div>

                <div class="modal-body">
                  <div class="modal-field">
                    <label>档案标识 (ID) <span class="text-danger">*</span></label>
                    <input
                      v-model="newProfileId"
                      type="text"
                      class="setting-text-input font-mono"
                      placeholder="例如: company-internal 或 fast-mirror"
                      autofocus
                    />
                    <span class="field-hint">用于 CLI 命令行切换，仅允许英文字母、数字、短横线与下划线</span>
                  </div>

                  <div class="modal-field">
                    <label>显示名称</label>
                    <input
                      v-model="newProfileName"
                      type="text"
                      class="setting-text-input"
                      placeholder="例如: 公司专线加速"
                    />
                  </div>
                </div>

                <div class="modal-footer">
                  <button class="btn-modal-cancel" @click="showCreateProfileModal = false">取消</button>
                  <button
                    class="btn-modal-confirm"
                    :disabled="!newProfileId.trim() || isSavingProfile"
                    @click="handleSaveAsProfile"
                  >
                    <span v-if="isSavingProfile">正在创建...</span>
                    <span v-else>创建并激活</span>
                  </button>
                </div>
              </div>
            </div>

            <!-- Add Ecosystem Modal Dialog -->
            <div v-if="showAddEcoModal" class="modal-backdrop" @click.self="showAddEcoModal = false">
              <div class="modal-dialog glass-card">
                <div class="modal-header">
                  <div class="modal-title-group">
                    <span class="modal-icon">
                      <TechIcon id="workbench:runtimes" :size="20" />
                    </span>
                    <div>
                      <h4>{{ t('network.add_ecosystem_title') || '添加语言生态镜像源' }}</h4>
                      <span class="modal-subtitle">{{ t('network.add_ecosystem_desc') || '选择常用语言生态预设或输入自定义生态标识与镜像地址' }}</span>
                    </div>
                  </div>
                  <button class="btn-modal-close" @click="showAddEcoModal = false">✕</button>
                </div>

                <div class="modal-body">
                  <div class="modal-field">
                    <label>{{ t('network.quick_templates') || '快速选择预设生态' }}</label>
                    <div class="quick-templates-grid">
                      <button
                        v-for="tpl in commonEcoTemplates"
                        :key="tpl.id"
                        type="button"
                        class="btn-template-chip"
                        :class="{ active: newEcoId === tpl.id }"
                        @click="applyEcoTemplate(tpl)"
                      >
                        {{ tpl.label }}
                      </button>
                    </div>
                  </div>

                  <div class="modal-field">
                    <label>{{ t('network.eco_identifier') || '生态标识 (ID)' }} <span class="text-danger">*</span></label>
                    <input
                      v-model="newEcoId"
                      type="text"
                      class="setting-text-input font-mono"
                      placeholder="例如: composer, maven, docker, nuget 等"
                      autofocus
                    />
                    <span class="field-hint">小写字母、数字或短横线，如 composer、maven、docker 等</span>
                  </div>

                  <div class="modal-field">
                    <label>{{ t('network.eco_registry_url') || '镜像源地址 (URL)' }} <span class="text-danger">*</span></label>
                    <input
                      v-model="newEcoUrl"
                      type="text"
                      class="setting-text-input font-mono"
                      placeholder="例如: https://repo.packagist.org"
                    />
                  </div>
                </div>

                <div class="modal-footer">
                  <button class="btn-modal-cancel" @click="showAddEcoModal = false">取消</button>
                  <button
                    class="btn-modal-confirm"
                    :disabled="!newEcoId.trim() || !newEcoUrl.trim() || isAddingEco"
                    @click="handleAddEcosystem"
                  >
                    <span v-if="isAddingEco">正在添加...</span>
                    <span v-else>添加并保存</span>
                  </button>
                </div>
              </div>
            </div>

            <!-- Ecosystem Registries Matrix Table -->
            <div class="registries-matrix" v-if="networkStatus">
              <div
                v-for="eco in orderedEffectiveEcosystems"
                :key="eco"
                class="registry-matrix-row"
              >
                <div class="row-eco-info">
                  <span class="eco-badge font-mono">{{ eco.toUpperCase() }}</span>
                  <span
                    class="status-indicator-badge"
                    :class="getEcoBadgeClass(eco)"
                    :title="getEcoBadgeTitle(eco)"
                  >
                    {{ getEcoBadgeLabel(eco) }}
                  </span>
                </div>

                <div class="row-controls">
                  <div class="preset-dropdown-wrap" v-if="recommendedMirrors[eco] && recommendedMirrors[eco].length">
                    <select
                      class="preset-select"
                      :value="getPresetSelectValue(eco)"
                      @change="handleSelectMirrorPreset(eco, ($event.target as HTMLSelectElement).value)"
                    >
                      <option
                        v-for="m in recommendedMirrors[eco]"
                        :key="m.url"
                        :value="m.url"
                      >
                        {{ m.name }}
                      </option>
                      <option
                        v-if="!isUrlInPresets(eco, networkStatus?.effective?.[eco]?.url)"
                        :value="networkStatus?.effective?.[eco]?.url"
                      >
                        {{ t('network.custom_source') || '自定义镜像源' }}
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
                    @keydown.enter="handleOverrideRegistry(eco, ($event.target as HTMLInputElement).value)"
                  />

                  <!-- Dynamic Ping / Latency Badge for this ecosystem -->
                  <button
                    class="btn-eco-ping"
                    :class="getEcoPingClass(eco)"
                    :disabled="ecoPingStatus[eco]?.loading"
                    :title="getEcoPingTitle(eco)"
                    @click="handlePingSingle(eco)"
                  >
                    <span v-if="ecoPingStatus[eco]?.loading" class="ping-icon spinning">
                      <TechIcon id="workbench:sync" :size="12" />
                    </span>
                    <span v-else-if="ecoPingStatus[eco]?.status === 'fast'" class="ping-text ping-fast" style="display: inline-flex; align-items: center; gap: 4px;">
                      <TechIcon id="workbench:speed" :size="11" />
                      <span>{{ (ecoPingStatus[eco]?.rtt_ms ?? 0) }}ms</span>
                    </span>
                    <span v-else-if="ecoPingStatus[eco]?.status === 'normal'" class="ping-text ping-normal" style="display: inline-flex; align-items: center; gap: 4px;">
                      <TechIcon id="workbench:speed" :size="11" />
                      <span>{{ (ecoPingStatus[eco]?.rtt_ms ?? 0) }}ms</span>
                    </span>
                    <span v-else-if="ecoPingStatus[eco]?.status === 'slow'" class="ping-text ping-slow" style="display: inline-flex; align-items: center; gap: 4px;">
                      <TechIcon id="workbench:speed" :size="11" />
                      <span>{{ (ecoPingStatus[eco]?.rtt_ms ?? 0) }}ms</span>
                    </span>
                    <span v-else-if="ecoPingStatus[eco]?.status === 'timeout'" class="ping-text ping-timeout" style="display: inline-flex; align-items: center; gap: 4px;">
                      <TechIcon id="workbench:warning" :size="11" />
                      <span>超时</span>
                    </span>
                    <span v-else-if="ecoPingStatus[eco]?.status === 'error'" class="ping-text ping-error" style="display: inline-flex; align-items: center; gap: 4px;">
                      <TechIcon id="workbench:warning" :size="11" />
                      <span>失败</span>
                    </span>
                    <span v-else class="ping-text ping-idle" style="display: inline-flex; align-items: center; gap: 4px;">
                      <TechIcon id="workbench:speed" :size="11" />
                      <span>测速</span>
                    </span>
                  </button>

                  <!-- Reset Override Button if Overridden -->
                  <button
                    v-if="networkStatus.effective[eco]?.is_overridden"
                    class="btn-reset-override"
                    :title="t('network.restore_inherited') || '恢复预设'"
                    @click="handleResetRegistry(eco)"
                  >
                    <TechIcon id="workbench:sync" :size="12" />
                  </button>

                  <!-- Delete Ecosystem Button -->
                  <button
                    class="btn-delete-eco"
                    :title="t('network.delete_ecosystem') || '从矩阵中移除'"
                    @click="handleDeleteEcosystem(eco)"
                  >
                    <TechIcon id="workbench:delete" :size="13" />
                  </button>
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
              <span class="section-icon">
                <TechIcon id="workbench:marketplace" :size="20" />
              </span>
              <div>
                <h2>{{ t('settings.cat_plugins') || '扩展' }}</h2>
                <span class="section-sub">{{ t('settings.cat_plugins_sub') || '插件配置与隔离' }}</span>
              </div>
            </div>
          </div>

          <div class="settings-list">
            <!-- Explanation banner -->
            <div class="plugin-isolation-banner">
              <span class="banner-icon">
                <TechIcon id="workbench:lock" :size="16" />
              </span>
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
                    <span class="plugin-icon">
                      <TechIcon id="workbench:marketplace" :size="16" />
                    </span>
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
              <span class="empty-icon">
                <TechIcon id="workbench:marketplace" :size="32" />
              </span>
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
              <span class="section-icon">
                <TechIcon id="workbench:lock" :size="20" />
              </span>
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

        <!-- Add/Edit Hook Modal Dialog -->
        <div v-if="showHookModal" class="modal-backdrop" @click.self="showHookModal = false">
          <div class="modal-dialog glass-card">
            <div class="modal-header">
              <div class="modal-title-group">
                <span class="modal-icon">
                  <TechIcon id="workbench:git" :size="20" />
                </span>
                <div>
                  <h4>{{ isEditingHook ? (t('settings.hooks_modal_edit_title') || '编辑生命周期钩子') : (t('settings.hooks_modal_add_title') || '添加生命周期钩子') }}</h4>
                  <span class="modal-subtitle">{{ t('settings.hooks_modal_sub') || '指定 Git 触发阶段与待执行的 Shell 脚本或任务命令' }}</span>
                </div>
              </div>
              <button class="btn-modal-close" @click="showHookModal = false">✕</button>
            </div>
            <div class="modal-body">
              <div class="modal-field">
                <label class="modal-label">{{ t('settings.hooks_form_stage') || 'Git 触发阶段 (Hook Stage)' }}</label>
                <CustomSelect
                  v-if="!isEditingHook"
                  :model-value="hookFormName"
                  :options="standardHookPresets"
                  @change="(val) => hookFormName = val"
                />
                <input
                  v-if="hookFormName === 'custom' || isEditingHook"
                  v-model="hookFormCustomName"
                  type="text"
                  class="modal-input font-mono"
                  style="margin-top: 6px;"
                  :placeholder="t('settings.hooks_form_name_placeholder') || '输入钩子名称 (例如 pre-commit, commit-msg)'"
                  :disabled="isEditingHook"
                />
              </div>

              <div class="modal-field">
                <label class="modal-label">{{ t('settings.hooks_form_cmd') || '执行命令 / 脚本 (Command)' }}</label>
                <textarea
                  v-model="hookFormCommand"
                  class="modal-textarea font-mono"
                  rows="3"
                  :placeholder="t('settings.hooks_form_cmd_placeholder') || '例如: npm test 或 unirtm run check'"
                ></textarea>
              </div>

              <div v-if="activeScope !== 'user'" class="modal-field-checkbox">
                <label class="modal-checkbox-row" for="hook-auto-bridge">
                  <input
                    type="checkbox"
                    id="hook-auto-bridge"
                    v-model="hookFormAutoBridge"
                  />
                  <span>{{ t('settings.hooks_form_auto_bridge') || '同时自动将 UniRTM 路由脚本注入到 .git/hooks/' }}</span>
                </label>
              </div>
            </div>
            <div class="modal-footer">
              <button class="btn-modal-cancel" @click="showHookModal = false">{{ t('common.cancel') || '取消' }}</button>
              <button
                class="btn-modal-confirm"
                :disabled="isSavingHook"
                @click="saveHook"
              >
                <span v-if="isSavingHook" class="spinner-small"></span>
                <span>{{ isSavingHook ? (t('common.saving') || '保存中...') : (t('common.save') || '保存钩子') }}</span>
              </button>
            </div>
          </div>
        </div>

        <!-- Hook Test Run Modal Dialog -->
        <div v-if="showHookRunModal" class="modal-backdrop" @click.self="showHookRunModal = false">
          <div class="modal-dialog glass-card" style="max-width: 600px;">
            <div class="modal-header">
              <div class="modal-title-group">
                <span class="modal-icon">
                  <TechIcon id="workbench:tasks" :size="20" />
                </span>
                <div>
                  <h4>{{ t('settings.hooks_run_modal_title') || '测试执行生命周期钩子' }}: <span class="font-mono text-cyan">{{ runningHookName }}</span></h4>
                  <span class="modal-subtitle">{{ t('settings.hooks_run_modal_sub') || '在当前工作区环境下通过 UniRTM Multi-Engine 调度运行' }}</span>
                </div>
              </div>
              <button class="btn-modal-close" @click="showHookRunModal = false">✕</button>
            </div>
            <div class="modal-body">
              <div v-if="isRunningHook" class="hook-run-loading" style="display: flex; align-items: center; gap: 10px; padding: 24px; justify-content: center; color: var(--text-muted);">
                <span class="spinner-small"></span>
                <span>{{ t('settings.hooks_running') || '正在执行钩子脚本，请稍候...' }}</span>
              </div>
              <div v-else-if="hookRunResult" class="hook-run-result">
                <div class="run-status-bar" style="margin-bottom: 12px; display: flex; align-items: center; justify-content: space-between;">
                  <span class="run-status-badge" :style="{ color: hookRunResult.success ? '#10b981' : '#ef4444', fontWeight: '600' }">
                    {{ hookRunResult.success ? '✓ 执行成功 (Exit Code 0)' : `✕ 执行失败 (Exit Code ${hookRunResult.exit_code})` }}
                  </span>
                </div>
                <div class="terminal-output-box" style="background: #090d16; border: 1px solid var(--card-border); border-radius: 6px; padding: 12px; font-family: monospace; font-size: 0.85rem; max-height: 300px; overflow-y: auto; color: #e2e8f0; white-space: pre-wrap;">{{ hookRunResult.output || '(无标准输出)' }}</div>
              </div>
            </div>
            <div class="modal-footer">
              <button class="btn-modal-confirm" @click="showHookRunModal = false">{{ t('common.close') || '关闭' }}</button>
            </div>
          </div>
        </div>
      </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from "vue";
import { useRouter } from "vue-router";
import { t, selectedLangSetting, setLanguage, SUPPORTED_LANGUAGES } from "../i18n";
import { useTheme } from "../composables/useTheme";
import { useThemeManager } from "../composables/useThemeManager";
import { useToast } from "../composables/useToast";
import { useAppScope } from "../composables/useAppScope";
import CustomSelect from "../components/CustomSelect.vue";
import TechIcon from "../components/common/TechIcon.vue";

const router = useRouter();
const { showToast } = useToast();
const { activeScope, setScope } = useAppScope();

function navigateToMarketplace(tab = "themes") {
  router.push({ path: "/marketplace", query: { tab } });
}

const props = defineProps<{
  initialScope?: "user" | "workspace";
  activeCategory?: string;
}>();

const emit = defineEmits<{
  (e: "select-category", catId: string): void;
}>();

const { currentTheme } = useTheme();
const {
  colorThemes,
  iconThemes,
  productIconThemes,
  activeColorTheme,
  activeIconTheme,
  activeProductIconTheme,
  activeAppearance,
  fetchThemes,
  installTheme,
  setActiveTheme,
} = useThemeManager();

// State
const searchQuery = ref("");
const activeCategory = ref(props.activeCategory || "general");
const showRawTomlModal = ref(false);
const isCopied = ref(false);

// Hooks Lifecycle Management State
interface HookItem {
  name: string;
  command: string;
  installed?: boolean;
}

const hooksList = ref<HookItem[]>([]);
const isLoadingHooks = ref(false);
const isInstallingBridges = ref(false);

const showHookModal = ref(false);
const isEditingHook = ref(false);
const hookFormName = ref("pre-commit");
const hookFormCustomName = ref("");
const hookFormCommand = ref("");
const hookFormAutoBridge = ref(true);
const isSavingHook = ref(false);

const showHookRunModal = ref(false);
const runningHookName = ref("");
const isRunningHook = ref(false);
const hookRunResult = ref<{ success: boolean; exit_code: number; output: string } | null>(null);

const standardHookPresets = [
  { value: "pre-commit", label: "pre-commit (提交前校验 / Lint & Test)" },
  { value: "prepare-commit-msg", label: "prepare-commit-msg (准备提交信息模板)" },
  { value: "commit-msg", label: "commit-msg (提交信息格式规范检查)" },
  { value: "post-commit", label: "post-commit (提交完成后续动作)" },
  { value: "pre-push", label: "pre-push (推送远程前全量校验)" },
  { value: "post-checkout", label: "post-checkout (检出切换分支依赖维护)" },
  { value: "post-merge", label: "post-merge (合并后自动拉取/安装)" },
  { value: "pre-rebase", label: "pre-rebase (变基前安全防御)" },
  { value: "custom", label: "自定义名称 (Custom Hook)..." },
];

async function fetchHooksList() {
  isLoadingHooks.value = true;
  try {
    const res = await fetch(`/api/hooks?scope=${activeScope.value}`);
    if (res.ok) {
      const data = await res.json();
      hooksList.value = data.hooks || [];
    }
  } catch (err) {
    console.warn("Failed to fetch /api/hooks:", err);
  } finally {
    isLoadingHooks.value = false;
  }
}

function openAddHookModal(presetName?: string, presetCmd?: string) {
  isEditingHook.value = false;
  hookFormName.value = presetName || "pre-commit";
  hookFormCustomName.value = presetName || "";
  hookFormCommand.value = presetCmd || "";
  hookFormAutoBridge.value = activeScope.value !== "user";
  showHookModal.value = true;
}

function openEditHookModal(hook: HookItem) {
  isEditingHook.value = true;
  const isStandard = standardHookPresets.some(p => p.value === hook.name && p.value !== "custom");
  if (isStandard) {
    hookFormName.value = hook.name;
    hookFormCustomName.value = "";
  } else {
    hookFormName.value = "custom";
    hookFormCustomName.value = hook.name;
  }
  hookFormCommand.value = hook.command;
  hookFormAutoBridge.value = hook.installed || false;
  showHookModal.value = true;
}

async function saveHook() {
  const finalName = hookFormName.value === "custom" ? hookFormCustomName.value.trim() : hookFormName.value.trim();
  if (!finalName) {
    showToast("请输入 Hook 名称", "warning");
    return;
  }
  if (!hookFormCommand.value.trim()) {
    showToast("请输入执行命令", "warning");
    return;
  }

  isSavingHook.value = true;
  try {
    const res = await fetch("/api/hooks", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        scope: activeScope.value,
        name: finalName,
        command: hookFormCommand.value.trim(),
        install_bridge: hookFormAutoBridge.value,
      }),
    });
    if (res.ok) {
      showToast(isEditingHook.value ? "Hook 更新成功" : "Hook 创建成功", "success");
      showHookModal.value = false;
      await fetchHooksList();
      await loadConfigData();
    } else {
      const data = await res.json();
      showToast(`保存失败: ${data.error || "未知错误"}`, "error");
    }
  } catch (err: any) {
    showToast(`保存失败: ${err.message}`, "error");
  } finally {
    isSavingHook.value = false;
  }
}

async function deleteHook(name: string) {
  if (!confirm(`确定要移除钩子 ${name} 吗？`)) return;
  try {
    const res = await fetch(`/api/hooks?name=${encodeURIComponent(name)}&scope=${activeScope.value}`, {
      method: "DELETE",
    });
    if (res.ok) {
      showToast(`已移除 ${name}`, "success");
      await fetchHooksList();
      await loadConfigData();
    } else {
      const data = await res.json();
      showToast(`删除失败: ${data.error}`, "error");
    }
  } catch (err: any) {
    showToast(`删除失败: ${err.message}`, "error");
  }
}

async function installAllGitBridges() {
  isInstallingBridges.value = true;
  try {
    const res = await fetch("/api/hooks/install", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ all: true }),
    });
    const data = await res.json();
    if (res.ok) {
      showToast(`✨ Git 桥接注入成功: ${data.installed?.join(", ") || "已生效"}`, "success");
      await fetchHooksList();
    } else {
      showToast(`注入失败: ${data.error}`, "error");
    }
  } catch (err: any) {
    showToast(`注入失败: ${err.message}`, "error");
  } finally {
    isInstallingBridges.value = false;
  }
}

async function runHookTest(name: string) {
  runningHookName.value = name;
  isRunningHook.value = true;
  hookRunResult.value = null;
  showHookRunModal.value = true;
  try {
    const res = await fetch("/api/hooks/run", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name, scope: activeScope.value }),
    });
    const data = await res.json();
    hookRunResult.value = data;
  } catch (err: any) {
    hookRunResult.value = {
      success: false,
      exit_code: 1,
      output: `执行失败: ${err.message}`,
    };
  } finally {
    isRunningHook.value = false;
  }
}

function copyCommand(cmd: string) {
  if (navigator?.clipboard?.writeText) {
    navigator.clipboard.writeText(cmd);
    showToast("已复制命令到剪贴板", "success");
  }
}

watch(activeScope, () => {
  loadConfigData();
  loadNetworkStatus();
  fetchHooksList();
});

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
const saveStatusType = ref<"idle" | "saving" | "success" | "error">("idle");

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

// Advanced Settings (Concurrency, Automation, Safety, Experimental)
const jobs = ref<number | string>(0);
const autoInstall = ref<boolean>(true);
const taskTimeout = ref<string>("");
const taskOutput = ref<"interleaved" | "prefix" | "quiet">("interleaved");
const minimumReleaseAge = ref<string>("");
const experimental = ref<boolean>(false);

const taskOutputOptions = computed(() => [
  { value: "interleaved", label: t("settings.task_output_interleaved") || "交叉流式 (Interleaved)" },
  { value: "prefix", label: t("settings.task_output_prefix") || "任务前缀 (Prefix)" },
  { value: "quiet", label: t("settings.task_output_quiet") || "静默模式 (Quiet)" },
]);


// Network matrix and profile state
const isSavingProfile = ref(false);
const showCreateProfileModal = ref(false);
const newProfileId = ref("");
const newProfileName = ref("");
const isProbing = ref(false);
const detectedProxy = ref<any>(null);
const isDetectingProxy = ref(false);
const isTestingProxy = ref(false);
const proxyTestResult = ref<{
  success: boolean;
  rtt_ms?: number;
  message?: string;
  error?: string;
} | null>(null);
interface EcoPingResult {
  rtt_ms?: number;
  status: 'idle' | 'loading' | 'fast' | 'normal' | 'slow' | 'timeout' | 'error';
  loading?: boolean;
  error?: string;
}
const ecoPingStatus = ref<Record<string, EcoPingResult>>({});
const showAddEcoModal = ref(false);
const isAddingEco = ref(false);
const newEcoId = ref("");
const newEcoUrl = ref("");

const commonEcoTemplates = [
  { id: "cran", label: "R (CRAN)", url: "https://cran.r-project.org" },
  { id: "julia", label: "Julia (Pkg)", url: "https://pkg.julialang.org" },
  { id: "cpan", label: "Perl (CPAN)", url: "https://www.cpan.org" },
  { id: "hackage", label: "Haskell (Hackage)", url: "https://hackage.haskell.org" },
  { id: "stackage", label: "Haskell (Stackage)", url: "https://www.stackage.org" },
  { id: "ctan", label: "TeX (CTAN)", url: "https://mirror.ctan.org/" },
  { id: "elpa", label: "Emacs (ELPA)", url: "https://elpa.gnu.org/packages/" },
  { id: "clojars", label: "Clojure (Clojars)", url: "https://repo.clojars.org" },
  { id: "deno", label: "Deno", url: "https://deno.land/x" },
];

const recommendedMirrors = ref<Record<string, Array<{ name: string; url: string }>>>({
  npm: [
    { name: "官方直连 (npmjs.org)", url: "https://registry.npmjs.org" },
    { name: "npmmirror (淘宝/阿里云)", url: "https://registry.npmmirror.com" },
    { name: "腾讯云镜像", url: "https://mirrors.cloud.tencent.com/npm/" },
    { name: "华为云镜像", url: "https://repo.huaweicloud.com/repository/npm/" },
  ],
  go: [
    { name: "官方直连 (Google)", url: "https://proxy.golang.org,direct" },
    { name: "goproxy.cn (七牛云/社区)", url: "https://goproxy.cn,direct" },
    { name: "阿里云 GoProxy", url: "https://mirrors.aliyun.com/goproxy/,direct" },
    { name: "goproxy.io", url: "https://goproxy.io,direct" },
  ],
  python: [
    { name: "官方直连 (PyPI)", url: "https://pypi.org/simple" },
    { name: "清华大学 TUNA", url: "https://pypi.tuna.tsinghua.edu.cn/simple" },
    { name: "中国科学技术大学 USTC", url: "https://mirrors.ustc.edu.cn/pypi/simple" },
    { name: "阿里云 PyPI", url: "https://mirrors.aliyun.com/pypi/simple/" },
    { name: "腾讯云 PyPI", url: "https://mirrors.cloud.tencent.com/pypi/simple/" },
    { name: "华为云 PyPI", url: "https://repo.huaweicloud.com/repository/pypi/simple/" },
  ],
  rust: [
    { name: "官方直连 (crates.io)", url: "https://crates.io" },
    { name: "rsproxy.cn (字节跳动/社区)", url: "https://rsproxy.cn/crates.io-index" },
    { name: "中国科学技术大学 USTC", url: "https://mirrors.ustc.edu.cn/crates.io-index" },
    { name: "清华大学 TUNA", url: "https://mirrors.tuna.tsinghua.edu.cn/git/crates.io-index.git" },
  ],
  ruby: [
    { name: "官方直连 (RubyGems)", url: "https://rubygems.org" },
    { name: "清华大学 TUNA", url: "https://mirrors.tuna.tsinghua.edu.cn/rubygems/" },
    { name: "中国科学技术大学 USTC", url: "https://mirrors.ustc.edu.cn/rubygems/" },
    { name: "腾讯云 RubyGems", url: "https://mirrors.cloud.tencent.com/rubygems/" },
  ],
  composer: [
    { name: "官方直连 (Packagist)", url: "https://repo.packagist.org" },
    { name: "阿里云 Composer", url: "https://mirrors.aliyun.com/composer/" },
    { name: "腾讯云 Composer", url: "https://mirrors.cloud.tencent.com/composer/" },
    { name: "华为云 Composer", url: "https://repo.huaweicloud.com/repository/php/" },
  ],
  maven: [
    { name: "官方直连 (Maven Central)", url: "https://repo1.maven.org/maven2" },
    { name: "阿里云 Maven", url: "https://maven.aliyun.com/repository/public" },
    { name: "华为云 Maven", url: "https://repo.huaweicloud.com/repository/maven/" },
    { name: "清华大学 TUNA", url: "https://mirrors.tuna.tsinghua.edu.cn/maven/" },
  ],
  nuget: [
    { name: "官方直连 (NuGet.org)", url: "https://api.nuget.org/v3/index.json" },
    { name: "华为云 NuGet", url: "https://repo.huaweicloud.com/repository/nuget/v3/index.json" },
    { name: "腾讯云 NuGet", url: "https://mirrors.cloud.tencent.com/nuget/" },
  ],
  docker: [
    { name: "官方直连 (Docker Hub)", url: "https://registry-1.docker.io" },
    { name: "阿里云 Docker 镜像", url: "https://registry.cn-hangzhou.aliyuncs.com" },
    { name: "网易 Docker 镜像", url: "https://hub-mirror.c.163.com" },
    { name: "百度 Docker 镜像", url: "https://mirror.baidubce.com" },
  ],
  dart: [
    { name: "官方直连 (pub.dev)", url: "https://pub.dev" },
    { name: "清华大学 TUNA", url: "https://mirrors.tuna.tsinghua.edu.cn/dart-pub" },
    { name: "上海交通大学 SJTUG", url: "https://mirrors.sjtug.sjtu.edu.cn/dart-pub" },
  ],
  cocoapods: [
    { name: "官方直连 (CocoaPods CDN)", url: "https://cdn.cocoapods.org" },
    { name: "清华大学 TUNA", url: "https://mirrors.tuna.tsinghua.edu.cn/git/CocoaPods/Specs.git" },
    { name: "中国科学技术大学 USTC", url: "https://mirrors.ustc.edu.cn/CocoaPods-Specs.git" },
  ],
  conda: [
    { name: "官方直连 (Anaconda)", url: "https://repo.anaconda.com/pkgs/main" },
    { name: "清华大学 TUNA", url: "https://mirrors.tuna.tsinghua.edu.cn/anaconda/pkgs/main" },
    { name: "北京外国语大学 BFSU", url: "https://mirrors.bfsu.edu.cn/anaconda/pkgs/main" },
    { name: "中国科学技术大学 USTC", url: "https://mirrors.ustc.edu.cn/anaconda/pkgs/main" },
  ],
  homebrew: [
    { name: "官方直连 (Homebrew Bottles)", url: "https://ghcr.io/v2/homebrew/core" },
    { name: "清华大学 TUNA", url: "https://mirrors.tuna.tsinghua.edu.cn/homebrew-bottles" },
    { name: "中国科学技术大学 USTC", url: "https://mirrors.ustc.edu.cn/homebrew-bottles" },
    { name: "阿里云 Homebrew", url: "https://mirrors.aliyun.com/homebrew/homebrew-bottles" },
  ],
  cran: [
    { name: "官方直连 (CRAN)", url: "https://cran.r-project.org" },
    { name: "清华大学 TUNA", url: "https://mirrors.tuna.tsinghua.edu.cn/CRAN/" },
    { name: "中国科学技术大学 USTC", url: "https://mirrors.ustc.edu.cn/CRAN/" },
    { name: "南方科技大学 SUSTech", url: "https://mirrors.sustech.edu.cn/CRAN/" },
    { name: "阿里云 CRAN", url: "https://mirrors.aliyun.com/CRAN/" },
  ],
  julia: [
    { name: "官方直连 (Julia Pkg)", url: "https://pkg.julialang.org" },
    { name: "清华大学 TUNA", url: "https://mirrors.tuna.tsinghua.edu.cn/julia" },
    { name: "中国科学技术大学 USTC", url: "https://mirrors.ustc.edu.cn/julia" },
    { name: "北京外国语大学 BFSU", url: "https://mirrors.bfsu.edu.cn/julia" },
  ],
  cpan: [
    { name: "官方直连 (CPAN)", url: "https://www.cpan.org" },
    { name: "清华大学 TUNA", url: "https://mirrors.tuna.tsinghua.edu.cn/CPAN/" },
    { name: "中国科学技术大学 USTC", url: "https://mirrors.ustc.edu.cn/CPAN/" },
    { name: "阿里云 CPAN", url: "https://mirrors.aliyun.com/CPAN/" },
  ],
  hackage: [
    { name: "官方直连 (Hackage)", url: "https://hackage.haskell.org" },
    { name: "清华大学 TUNA", url: "https://mirrors.tuna.tsinghua.edu.cn/hackage" },
    { name: "中国科学技术大学 USTC", url: "https://mirrors.ustc.edu.cn/hackage" },
  ],
  clojars: [
    { name: "官方直连 (Clojars)", url: "https://repo.clojars.org" },
    { name: "清华大学 TUNA", url: "https://mirrors.tuna.tsinghua.edu.cn/clojars" },
  ],
  ctan: [
    { name: "官方直连 (CTAN)", url: "https://mirror.ctan.org/" },
    { name: "中国科学技术大学 USTC", url: "https://mirrors.ustc.edu.cn/CTAN/" },
    { name: "清华大学 TUNA", url: "https://mirrors.tuna.tsinghua.edu.cn/CTAN/" },
  ],
  elpa: [
    { name: "官方直连 (GNU ELPA)", url: "https://elpa.gnu.org/packages/" },
    { name: "中国科学技术大学 USTC", url: "https://mirrors.ustc.edu.cn/elpa/" },
    { name: "清华大学 TUNA", url: "https://mirrors.tuna.tsinghua.edu.cn/elpa/" },
  ],
  stackage: [
    { name: "官方直连 (Stackage)", url: "https://www.stackage.org" },
    { name: "中国科学技术大学 USTC", url: "https://mirrors.ustc.edu.cn/stackage" },
  ],
});

const networkStatus = ref<any>({
  active_profile_id: "official",
  active_profile: {
    id: "official",
    name: "官方直连",
    description: "全量使用官方源与官方二进制 CDN，不经任何代理或第三方镜像",
    builtin: true,
  },
  proxy: {
    enabled: false,
    http: "",
    https: "",
    socks5: "",
    no_proxy: "localhost,127.0.0.1,::1,127.0.0.0/8,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16,169.254.0.0/16,*.local,*.internal,*.lan,*.cluster.local",
  },
  effective: {
    npm: { ecosystem: "npm", url: "https://registry.npmjs.org", source: "profile", is_overridden: false },
    go: { ecosystem: "go", url: "https://proxy.golang.org,direct", source: "profile", is_overridden: false },
    python: { ecosystem: "python", url: "https://pypi.org/simple", source: "profile", is_overridden: false },
    rust: { ecosystem: "rust", url: "https://crates.io", source: "profile", is_overridden: false },
    ruby: { ecosystem: "ruby", url: "https://rubygems.org", source: "profile", is_overridden: false },
    composer: { ecosystem: "composer", url: "https://repo.packagist.org", source: "profile", is_overridden: false },
    maven: { ecosystem: "maven", url: "https://repo1.maven.org/maven2", source: "profile", is_overridden: false },
    nuget: { ecosystem: "nuget", url: "https://api.nuget.org/v3/index.json", source: "profile", is_overridden: false },
    docker: { ecosystem: "docker", url: "https://registry-1.docker.io", source: "profile", is_overridden: false },
    dart: { ecosystem: "dart", url: "https://pub.dev", source: "profile", is_overridden: false },
    cocoapods: { ecosystem: "cocoapods", url: "https://cdn.cocoapods.org", source: "profile", is_overridden: false },
    conda: { ecosystem: "conda", url: "https://repo.anaconda.com/pkgs/main", source: "profile", is_overridden: false },
    homebrew: { ecosystem: "homebrew", url: "https://ghcr.io/v2/homebrew/core", source: "profile", is_overridden: false },
  },
  profiles: {
    official: {
      id: "official",
      name: "官方直连",
      description: "全量使用官方源与官方二进制 CDN，不经任何代理或第三方镜像",
      builtin: true,
    },
    cn: {
      id: "cn",
      name: "中国镜像加速",
      description: "中国顶级高校与大厂公共镜像加速矩阵，保障编译前依赖极速拉取",
      builtin: true,
    },
  },
});

const orderedProfiles = computed(() => {
  if (!networkStatus.value?.profiles) return [];
  const list = Object.entries(networkStatus.value.profiles).map(([id, profile]) => ({
    id,
    profile: profile as any,
  }));
  const rank: Record<string, number> = {
    official: 1,
    cn: 2,
    tuna: 3,
    aliyun: 4,
  };
  return list.sort((a, b) => {
    const rankA = rank[a.id] || 99;
    const rankB = rank[b.id] || 99;
    if (rankA !== rankB) return rankA - rankB;
    return a.id.localeCompare(b.id);
  });
});

function getProfileDisplayName(id: string, fallbackName?: string): string {
  if (id === "cn") {
    return t("network.profile_cn") || "中国镜像加速";
  }
  if (id === "official") {
    return t("network.profile_official") || "官方直连";
  }
  if (id === "tuna") {
    return t("network.profile_tuna") || "清华大学 TUNA 专线";
  }
  if (id === "aliyun") {
    return t("network.profile_aliyun") || "阿里云镜像专线";
  }
  return fallbackName || id;
}

function getProfileDisplayDesc(id: string, fallbackDesc?: string): string {
  if (id === "cn") {
    return t("network.profile_cn_desc") || "中国顶级高校与大厂公共镜像加速矩阵，保障编译前依赖极速拉取";
  }
  if (id === "official") {
    return t("network.profile_official_desc") || "全量使用官方源与官方二进制 CDN，不经任何代理或第三方镜像";
  }
  if (id === "tuna") {
    return t("network.profile_tuna_desc") || "优先全量采用清华大学开源软件镜像站 (TUNA) 教育网与公网优质骨干源";
  }
  if (id === "aliyun") {
    return t("network.profile_aliyun_desc") || "优先全量采用阿里巴巴开源镜像站生态高速 CDN 加速矩阵";
  }
  return fallbackDesc || "";
}

const profileTrackRef = ref<HTMLElement | null>(null);

function scrollProfiles(direction: 'left' | 'right') {
  if (!profileTrackRef.value) return;
  const scrollAmount = 280;
  profileTrackRef.value.scrollBy({
    left: direction === 'left' ? -scrollAmount : scrollAmount,
    behavior: 'smooth',
  });
}

function onProfileTrackWheel(e: WheelEvent) {
  if (!profileTrackRef.value) return;
  if (Math.abs(e.deltaY) > Math.abs(e.deltaX) && profileTrackRef.value.scrollWidth > profileTrackRef.value.clientWidth) {
    e.preventDefault();
    profileTrackRef.value.scrollLeft += e.deltaY;
  }
}

const orderedEffectiveEcosystems = computed(() => {
  if (!networkStatus.value?.effective) return [];
  const baseOrder = [
    "npm", "go", "python", "rust", "ruby", "composer", "maven", "nuget", "docker",
    "dart", "cocoapods", "conda", "homebrew"
  ];
  const effKeys = Object.keys(networkStatus.value.effective);
  const ordered: string[] = [];
  for (const eco of baseOrder) {
    if (effKeys.includes(eco)) {
      ordered.push(eco);
    }
  }
  const extras = effKeys.filter((eco) => !baseOrder.includes(eco)).sort();
  return [...ordered, ...extras];
});

function getEcoBadgeClass(eco: string): string {
  const eff = networkStatus.value?.effective?.[eco];
  if (eff?.is_overridden) return 'badge-override';
  return 'badge-inherited';
}

function getEcoBadgeLabel(eco: string): string {
  const eff = networkStatus.value?.effective?.[eco];
  if (eff?.is_overridden) return t('network.badge_overridden') || '自定义';
  return t('network.badge_inherited') || '预设';
}

function getEcoBadgeTitle(eco: string): string {
  const eff = networkStatus.value?.effective?.[eco];
  if (eff?.source?.startsWith('fallback_')) {
    return t('network.badge_fallback_cn_tip') || '当前专线未收录该生态，已自动回退到中国镜像加速矩阵补齐';
  }
  return '';
}

function openAddEcoModal() {
  newEcoId.value = "";
  newEcoUrl.value = "";
  showAddEcoModal.value = true;
}

function applyEcoTemplate(tpl: { id: string; url: string }) {
  newEcoId.value = tpl.id;
  newEcoUrl.value = tpl.url;
}

async function handleAddEcosystem() {
  const eco = newEcoId.value.trim().toLowerCase();
  const url = newEcoUrl.value.trim();
  if (!eco || !url) return;
  isAddingEco.value = true;
  isSaving.value = true;
  saveStatusType.value = "saving";
  saveStatusText.value = "正在添加生态...";
  try {
    const res = await fetch("/api/network/registry/add", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ ecosystem: eco, url, scope: activeScope.value }),
    });
    if (res.ok) {
      showAddEcoModal.value = false;
      newEcoId.value = "";
      newEcoUrl.value = "";
      saveStatusType.value = "success";
      saveStatusText.value = "已自动保存";
      await loadNetworkStatus();
      await loadConfigData();
      showToast(`已添加生态镜像源: ${eco}`, "success");
      setTimeout(() => {
        saveStatusType.value = "idle";
        saveStatusText.value = "";
      }, 2000);
    } else {
      const err = await res.json().catch(() => ({}));
      saveStatusType.value = "error";
      saveStatusText.value = `添加失败: ${err.error || "未知错误"}`;
    }
  } catch (err: any) {
    console.error(err);
    saveStatusType.value = "error";
    saveStatusText.value = `网络错误: ${err.message}`;
  } finally {
    isAddingEco.value = false;
    isSaving.value = false;
  }
}

async function handleDeleteEcosystem(eco: string) {
  isSaving.value = true;
  saveStatusType.value = "saving";
  saveStatusText.value = "正在移除生态...";
  try {
    const res = await fetch("/api/network/registry/delete", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ ecosystem: eco, scope: activeScope.value }),
    });
    if (res.ok) {
      saveStatusType.value = "success";
      saveStatusText.value = "已移除生态";
      await loadNetworkStatus();
      await loadConfigData();
      showToast(`已从矩阵中移除生态: ${eco}`, "success");
      setTimeout(() => {
        saveStatusType.value = "idle";
        saveStatusText.value = "";
      }, 2000);
    } else {
      const err = await res.json().catch(() => ({}));
      saveStatusType.value = "error";
      saveStatusText.value = `移除失败: ${err.error || "未知错误"}`;
    }
  } catch (err: any) {
    console.error(err);
    saveStatusType.value = "error";
    saveStatusText.value = `网络错误: ${err.message}`;
  } finally {
    isSaving.value = false;
  }
}

// Full config from API
const configData = ref<any>(null);


const languageSelectOptions = computed(() => [
  { value: "auto", label: t("common.autoDetect") || "自动检测" },
  ...SUPPORTED_LANGUAGES.map((item) => ({
    value: item.code,
    label: item.nativeName,
  })),
]);

function cleanThemeLabel(rawName: string): string {
  if (!rawName) return "";
  let name = rawName;
  if (/^seti/i.test(name)) {
    return "Seti UI";
  }
  // 移除任何括号及其内容，如 (v13.0.0), (VS Code Default), (默认), [v1.0] 等
  name = name.replace(/\s*[\(\[\（\【][^\)\]\）\】]*[\)\]\）\】]\s*/g, " ");
  // 移除裸版本号标识，如 v13.0.0, 1.2.3 等
  name = name.replace(/\b[vV]?\d+(\.\d+)+(-[a-zA-Z0-9.]+)?\b/g, "");
  // 规范化空格
  name = name.trim().replace(/\s+/g, " ");
  return name || rawName;
}

const colorThemeOptions = computed(() => {
  const options: Array<{ value: string; label: string }> = [
    { value: "default", label: "UniRTM 默认自适应" },
    { value: "light", label: "UniRTM 官方浅色" },
    { value: "dark", label: "UniRTM 官方深色" },
  ];

  for (const item of colorThemes.value) {
    if (item.id === "default" || item.id === "dark" || item.id === "light" || item.id === "system") {
      continue;
    }
    options.push({
      value: item.id,
      label: cleanThemeLabel(item.name),
    });
  }

  return options;
});

const iconThemeOptions = computed(() => {
  return iconThemes.value.map((item) => {
    return {
      value: item.id,
      label: cleanThemeLabel(item.name),
    };
  });
});

const productIconThemeOptions = computed(() => {
  return productIconThemes.value.map((item) => {
    return {
      value: item.id,
      label: cleanThemeLabel(item.name),
    };
  });
});


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

const browserPresetOptions = computed(() => {
  const options: Array<{ value: string; label: string }> = [
    { value: "", label: t("settings.browser_auto") || "系统默认 (自动探测)" },
  ];

  for (const b of detectedBrowsers.value) {
    options.push({
      value: b.id,
      label: b.name,
    });
  }

  options.push({
    value: "custom",
    label: t("settings.browser_custom") || "自定义路径",
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
      q.includes("缓存") ||
      q.includes("job") ||
      q.includes("concur") ||
      q.includes("task") ||
      q.includes("timeout") ||
      q.includes("auto") ||
      q.includes("exper") ||
      q.includes("age") ||
      q.includes("并发") ||
      q.includes("任务") ||
      q.includes("超时") ||
      q.includes("实验"))
  )
    return true;
  if (
    catId === "hooks" &&
    (q.includes("hook") ||
      q.includes("git") ||
      q.includes("commit") ||
      q.includes("push") ||
      q.includes("钩子") ||
      q.includes("生命周期") ||
      q.includes("自动化") ||
      q.includes("脚本") ||
      q.includes("提交") ||
      q.includes("推送"))
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
  setScope(scope, router);
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

  jobs.value = s.jobs ?? 0;
  autoInstall.value = s.auto_install !== false;
  taskTimeout.value = s.task_timeout !== undefined ? String(s.task_timeout) : "";
  taskOutput.value = s.task_output || "interleaved";
  minimumReleaseAge.value = s.minimum_release_age !== undefined && s.minimum_release_age !== "" ? s.minimum_release_age : "7d";
  experimental.value = Boolean(s.experimental);

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

function onJobsChange() {
  const num = parseInt(String(jobs.value), 10);
  const normalized = isNaN(num) || num < 0 ? 0 : num;
  jobs.value = normalized;
  saveSetting("settings.jobs", normalized);
}

function onAutoInstallToggle() {
  autoInstall.value = !autoInstall.value;
  saveSetting("settings.auto_install", autoInstall.value);
}

function onTaskTimeoutChange() {
  const val = String(taskTimeout.value).trim();
  taskTimeout.value = val;
  saveSetting("settings.task_timeout", val);
}

function onTaskOutputChange(val: string) {
  taskOutput.value = val as any;
  saveSetting("settings.task_output", val);
}

function onMinimumReleaseAgeChange() {
  const val = String(minimumReleaseAge.value).trim();
  minimumReleaseAge.value = val;
  saveSetting("settings.minimum_release_age", val);
}

function applyReleaseAgePreset(val: string) {
  minimumReleaseAge.value = val;
  saveSetting("settings.minimum_release_age", val);
}

function onExperimentalToggle() {
  experimental.value = !experimental.value;
  saveSetting("settings.experimental", experimental.value);
}

async function saveSetting(key: string, value: any) {
  isSaving.value = true;
  saveStatusType.value = "saving";
  saveStatusText.value = "保存中...";

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
      saveStatusType.value = "success";
      saveStatusText.value = "已自动保存";
      await loadConfigData();
      setTimeout(() => {
        saveStatusType.value = "idle";
        saveStatusText.value = "";
      }, 2000);
    } else {
      isSaving.value = false;
      saveStatusType.value = "error";
      saveStatusText.value = `保存失败: ${data.error || "未知错误"}`;
    }
  } catch (err: any) {
    isSaving.value = false;
    saveStatusType.value = "error";
    saveStatusText.value = `网络错误: ${err.message}`;
  }
}

function onLanguageChange(val: string) {
  appLanguage.value = val;
  setLanguage(val);
  saveSetting("settings.language", val);
}

const vsixInputRef = ref<HTMLInputElement | null>(null);
const isInstallingTheme = ref(false);
const themeInstallMsg = ref("");
const themeInstallError = ref(false);

function triggerVsixUpload() {
  vsixInputRef.value?.click();
}

async function onVsixFileSelected(e: Event) {
  const target = e.target as HTMLInputElement;
  const files = target.files;
  if (!files || files.length === 0) return;
  const file = files[0];
  isInstallingTheme.value = true;
  themeInstallMsg.value = "";
  themeInstallError.value = false;

  try {
    const installed = await installTheme(file);
    const colorItems = installed.filter((i) => i.type === "color");
    const iconItems = installed.filter((i) => i.type === "icon");
    const productItems = installed.filter((i) => i.type === "product");

    const parts: string[] = [];
    if (colorItems.length > 0) parts.push(`${colorItems.length} 个颜色主题`);
    if (iconItems.length > 0) parts.push(`${iconItems.length} 个文件图标`);
    if (productItems.length > 0) parts.push(`${productItems.length} 个界面图标`);

    const typeSummary = parts.length > 0 ? parts.join("，") : `${installed.length} 项`;
    themeInstallMsg.value = `✓ 成功导入 ${typeSummary}！`;

    if (installed.length > 0) {
      const first = installed[0];
      if (first.type === "color") {
        await setActiveTheme(first.id, undefined);
      } else if (first.type === "icon") {
        await setActiveTheme(undefined, first.id);
      } else if (first.type === "product") {
        await setActiveTheme(undefined, undefined, undefined, first.id);
      }
    }
    await fetchThemes();
    await loadConfigData();
    showToast(`成功导入 VS Code 扩展包，识别到: ${typeSummary}`, "success");
  } catch (err: any) {
    themeInstallError.value = true;
    themeInstallMsg.value = `✕ 安装失败: ${err.message || err}`;
    showToast(`导入失败: ${err.message || err}`, "error");
  } finally {
    isInstallingTheme.value = false;
    target.value = "";
  }
}

async function onAppearanceChange(mode: "system" | "light" | "dark") {
  isSaving.value = true;
  saveStatusType.value = "saving";
  saveStatusText.value = "正在切换外观...";
  try {
    await setActiveTheme(undefined, undefined, mode);
    saveStatusType.value = "success";
    saveStatusText.value = "已自动保存";
    await loadConfigData();
    showToast(
      mode === "system"
        ? "已设置为跟随系统外观"
        : mode === "light"
        ? "已切换为浅色模式"
        : "已切换为深色模式",
      "info"
    );
    setTimeout(() => {
      saveStatusType.value = "idle";
      saveStatusText.value = "";
    }, 2000);
  } catch (err: any) {
    saveStatusType.value = "error";
    saveStatusText.value = `外观设置失败: ${err.message || err}`;
  } finally {
    isSaving.value = false;
  }
}

async function onColorThemeChange(val: string) {
  isSaving.value = true;
  saveStatusType.value = "saving";
  saveStatusText.value = "正在应用主题...";
  try {
    await setActiveTheme(val, undefined, undefined);
    appTheme.value = val;
    const themeObj = colorThemes.value.find((c) => c.id === val);
    const themeName = cleanThemeLabel(themeObj ? themeObj.name : val);
    saveStatusType.value = "success";
    saveStatusText.value = "已自动保存";
    await loadConfigData();
    showToast(`已应用配色主题: ${themeName}`, "success");
    setTimeout(() => {
      saveStatusType.value = "idle";
      saveStatusText.value = "";
    }, 2000);
  } catch (err: any) {
    saveStatusType.value = "error";
    saveStatusText.value = `应用主题失败: ${err.message || err}`;
  } finally {
    isSaving.value = false;
  }
}

async function onIconThemeChange(val: string) {
  isSaving.value = true;
  saveStatusType.value = "saving";
  saveStatusText.value = "正在应用文件图标主题...";
  try {
    await setActiveTheme(undefined, val, undefined);
    const iconObj = iconThemes.value.find((i) => i.id === val);
    const iconName = cleanThemeLabel(iconObj ? iconObj.name : val);
    saveStatusType.value = "success";
    saveStatusText.value = "已自动保存";
    await loadConfigData();
    showToast(`已应用文件图标主题: ${iconName}`, "success");
    setTimeout(() => {
      saveStatusType.value = "idle";
      saveStatusText.value = "";
    }, 2000);
  } catch (err: any) {
    saveStatusType.value = "error";
    saveStatusText.value = `应用文件图标失败: ${err.message || err}`;
  } finally {
    isSaving.value = false;
  }
}

async function onProductIconThemeChange(val: string) {
  isSaving.value = true;
  saveStatusType.value = "saving";
  saveStatusText.value = "正在应用界面图标主题...";
  try {
    await setActiveTheme(undefined, undefined, undefined, val);
    const prodObj = productIconThemes.value.find((p) => p.id === val);
    const prodName = cleanThemeLabel(prodObj ? prodObj.name : val);
    saveStatusType.value = "success";
    saveStatusText.value = "已自动保存";
    await loadConfigData();
    showToast(`已应用界面图标主题: ${prodName}`, "success");
    setTimeout(() => {
      saveStatusType.value = "idle";
      saveStatusText.value = "";
    }, 2000);
  } catch (err: any) {
    saveStatusType.value = "error";
    saveStatusText.value = `应用界面图标失败: ${err.message || err}`;
  } finally {
    isSaving.value = false;
  }
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
  isSaving.value = true;
  saveStatusType.value = "saving";
  saveStatusText.value = "正在切换档案...";
  try {
    const res = await fetch("/api/network/profile/use", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ profile_id: id, scope: activeScope.value }),
    });
    const data = await res.json().catch(() => ({}));
    if (res.ok) {
      saveStatusType.value = "success";
      saveStatusText.value = "已应用档案";
      await loadNetworkStatus();
      await loadConfigData();
      showToast(`已切换网络档案为: ${id}`, "success");
      setTimeout(() => {
        saveStatusType.value = "idle";
        saveStatusText.value = "";
      }, 2000);
    } else {
      saveStatusType.value = "error";
      saveStatusText.value = `切换档案失败: ${data.error || "未知错误"}`;
      showToast(`切换档案失败: ${data.error || "未知错误"}`, "error");
    }
  } catch (err: any) {
    saveStatusType.value = "error";
    saveStatusText.value = `网络错误: ${err.message}`;
  } finally {
    isSaving.value = false;
  }
}

async function handleSaveAsProfile() {
  const pid = newProfileId.value.trim();
  if (!pid) return;
  isSavingProfile.value = true;
  isSaving.value = true;
  saveStatusType.value = "saving";
  saveStatusText.value = "正在创建档案...";
  try {
    const res = await fetch("/api/network/profile/save", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        profile_id: pid,
        name: newProfileName.value.trim() || pid,
        scope: activeScope.value,
      }),
    });
    if (res.ok) {
      showCreateProfileModal.value = false;
      newProfileId.value = "";
      newProfileName.value = "";
      saveStatusType.value = "success";
      saveStatusText.value = "已创建并激活档案";
      await loadNetworkStatus();
      await loadConfigData();
      showToast(`已创建并激活档案: ${pid}`, "success");
      setTimeout(() => {
        saveStatusType.value = "idle";
        saveStatusText.value = "";
      }, 2000);
    } else {
      const err = await res.json().catch(() => ({}));
      saveStatusType.value = "error";
      saveStatusText.value = `创建档案失败: ${err.error || "未知错误"}`;
      alert(err.error || "创建档案失败");
    }
  } catch (err: any) {
    console.error(err);
    saveStatusType.value = "error";
    saveStatusText.value = `请求失败: ${err.message}`;
    alert("请求失败: " + err.message);
  } finally {
    isSavingProfile.value = false;
    isSaving.value = false;
  }
}

async function handleResetBuiltinProfile(id: string) {
  isSaving.value = true;
  saveStatusType.value = "saving";
  saveStatusText.value = "正在重置档案...";
  try {
    const res = await fetch("/api/network/profile/reset", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ profile_id: id, scope: activeScope.value }),
    });
    if (res.ok) {
      saveStatusType.value = "success";
      saveStatusText.value = "已恢复出厂默认值";
      await loadNetworkStatus();
      await loadConfigData();
      showToast("已恢复出厂默认值", "success");
      setTimeout(() => {
        saveStatusType.value = "idle";
        saveStatusText.value = "";
      }, 2000);
    } else {
      const err = await res.json().catch(() => ({}));
      saveStatusType.value = "error";
      saveStatusText.value = `重置失败: ${err.error || "未知错误"}`;
    }
  } catch (err: any) {
    saveStatusType.value = "error";
    saveStatusText.value = `网络错误: ${err.message}`;
  } finally {
    isSaving.value = false;
  }
}

async function handleResetAllEcosystems() {
  isSaving.value = true;
  saveStatusType.value = "saving";
  saveStatusText.value = "正在恢复默认生态...";
  try {
    const res = await fetch("/api/network/registry/reset-all", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ scope: activeScope.value }),
    });
    if (res.ok) {
      saveStatusType.value = "success";
      saveStatusText.value = "已恢复默认生态列表";
      await loadNetworkStatus();
      await loadConfigData();
      showToast("已恢复默认生态列表", "success");
      setTimeout(() => {
        saveStatusType.value = "idle";
        saveStatusText.value = "";
      }, 2000);
    } else {
      const err = await res.json().catch(() => ({}));
      saveStatusType.value = "error";
      saveStatusText.value = `恢复失败: ${err.error || "未知错误"}`;
    }
  } catch (err: any) {
    saveStatusType.value = "error";
    saveStatusText.value = `网络错误: ${err.message}`;
  } finally {
    isSaving.value = false;
  }
}

async function handleOverrideRegistry(ecosystem: string, url: string) {
  if (!ecosystem || !url) return;
  isSaving.value = true;
  saveStatusType.value = "saving";
  saveStatusText.value = "正在更新镜像源...";
  try {
    const res = await fetch("/api/network/registry/override", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ ecosystem, url, scope: activeScope.value }),
    });
    if (res.ok) {
      saveStatusType.value = "success";
      saveStatusText.value = "已自动保存";
      await loadNetworkStatus();
      await loadConfigData();
      setTimeout(() => {
        saveStatusType.value = "idle";
        saveStatusText.value = "";
      }, 2000);
    } else {
      const err = await res.json().catch(() => ({}));
      saveStatusType.value = "error";
      saveStatusText.value = `保存失败: ${err.error || "未知错误"}`;
    }
  } catch (err: any) {
    saveStatusType.value = "error";
    saveStatusText.value = `网络错误: ${err.message}`;
  } finally {
    isSaving.value = false;
  }
}

function isUrlInPresets(eco: string, url?: string): boolean {
  if (!url) return false;
  const list = recommendedMirrors.value[eco] || [];
  return list.some((m: any) => m.url === url);
}

function getPresetSelectValue(eco: string): string {
  return networkStatus.value?.effective?.[eco]?.url || "";
}

function handleSelectMirrorPreset(ecosystem: string, url: string) {
  if (!url) return;
  const profileUrl = networkStatus.value?.active_profile?.registries?.[ecosystem];
  if (profileUrl && url === profileUrl) {
    handleResetRegistry(ecosystem);
  } else {
    handleOverrideRegistry(ecosystem, url);
  }
}

async function handleResetRegistry(ecosystem: string) {
  isSaving.value = true;
  saveStatusType.value = "saving";
  saveStatusText.value = "正在恢复预设...";
  try {
    const res = await fetch("/api/network/registry/reset", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ ecosystem, scope: activeScope.value }),
    });
    if (res.ok) {
      saveStatusType.value = "success";
      saveStatusText.value = "已恢复预设";
      await loadNetworkStatus();
      await loadConfigData();
      setTimeout(() => {
        saveStatusType.value = "idle";
        saveStatusText.value = "";
      }, 2000);
    } else {
      const err = await res.json().catch(() => ({}));
      saveStatusType.value = "error";
      saveStatusText.value = `恢复失败: ${err.error || "未知错误"}`;
    }
  } catch (err) {
    saveStatusType.value = "error";
    saveStatusText.value = "网络请求失败";
  } finally {
    isSaving.value = false;
  }
}

async function handleSaveProxy() {
  if (!networkStatus.value) return;
  isSaving.value = true;
  saveStatusType.value = "saving";
  saveStatusText.value = "保存中...";
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
      saveStatusType.value = "success";
      saveStatusText.value = "已自动保存";
      await loadNetworkStatus();
      await loadConfigData();
      setTimeout(() => {
        saveStatusType.value = "idle";
        saveStatusText.value = "";
      }, 2000);
    } else {
      const err = await res.json().catch(() => ({}));
      saveStatusType.value = "error";
      saveStatusText.value = `保存失败: ${err.error || "未知错误"}`;
    }
  } catch (err: any) {
    saveStatusType.value = "error";
    saveStatusText.value = `网络错误: ${err.message}`;
  } finally {
    isSaving.value = false;
  }
}

async function handleDetectProxy(applyImmediately: boolean = false) {
  isDetectingProxy.value = true;
  try {
    const res = await fetch("/api/network/proxy/detect");
    if (res.ok) {
      const data = await res.json();
      if (data && data.detected) {
        detectedProxy.value = data;
        if (applyImmediately) {
          applyDetectedProxy();
        }
      }
    }
  } catch (err) {
    console.error("Failed to detect system proxy:", err);
  } finally {
    isDetectingProxy.value = false;
  }
}

async function applyDetectedProxy() {
  if (!detectedProxy.value || !networkStatus.value) return;
  networkStatus.value.proxy.http = detectedProxy.value.http || "";
  networkStatus.value.proxy.https = detectedProxy.value.https || detectedProxy.value.http || "";
  networkStatus.value.proxy.socks5 = detectedProxy.value.socks5 || "";
  networkStatus.value.proxy.enabled = true;
  await handleSaveProxy();
  await handleTestProxy();
}

async function handleTestProxy() {
  if (!networkStatus.value?.proxy) return;
  const http = networkStatus.value.proxy.http;
  const https = networkStatus.value.proxy.https;
  const socks5 = networkStatus.value.proxy.socks5;
  if (!http && !https && !socks5) {
    proxyTestResult.value = {
      success: false,
      error: "未填写任何代理服务器地址 (HTTP 或 SOCKS5)",
    };
    return;
  }

  isTestingProxy.value = true;
  proxyTestResult.value = null;

  try {
    const res = await fetch("/api/network/proxy/test", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ http, https, socks5 }),
    });
    if (res.ok) {
      const data = await res.json();
      proxyTestResult.value = data;
    } else {
      proxyTestResult.value = {
        success: false,
        error: `代理测试接口请求失败 (HTTP ${res.status})`,
      };
    }
  } catch (err: any) {
    proxyTestResult.value = {
      success: false,
      error: `测试请求发生异常: ${err.message}`,
    };
  } finally {
    isTestingProxy.value = false;
  }
}

function getEcoPingClass(eco: string) {
  const p = ecoPingStatus.value[eco];
  if (!p || p.status === 'idle') return 'ping-idle';
  if (p.loading) return 'ping-loading';
  return `ping-${p.status}`;
}

function getEcoPingTitle(eco: string) {
  const p = ecoPingStatus.value[eco];
  if (!p) return '点击测试此镜像延迟';
  if (p.loading) return '正在测速...';
  if (p.status === 'timeout') return '连接超时';
  if (p.status === 'error') return p.error || '连接失败';
  return `延迟: ${p.rtt_ms}ms (${p.status})`;
}

async function handlePingSingle(eco: string) {
  const url = networkStatus.value?.effective?.[eco]?.url;
  if (!url) return;

  ecoPingStatus.value[eco] = {
    ...ecoPingStatus.value[eco],
    loading: true,
    status: 'loading'
  };

  try {
    const res = await fetch("/api/network/ping", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ ecosystem: eco, url })
    });
    if (res.ok) {
      const data = await res.json();
      const item = data.results?.[0];
      if (item) {
        ecoPingStatus.value[eco] = {
          rtt_ms: item.rtt_ms ?? item.latency_ms ?? 0,
          status: item.status,
          loading: false,
          error: item.error
        };
      }
    } else {
      ecoPingStatus.value[eco] = { status: 'error', loading: false, error: '请求失败' };
    }
  } catch (err: any) {
    ecoPingStatus.value[eco] = { status: 'error', loading: false, error: err?.message || '网络异常' };
  }
}

async function handleRunProbeAll() {
  if (!networkStatus.value?.effective) return;
  isProbing.value = true;

  const targets = orderedEffectiveEcosystems.value
    .map(eco => ({
      ecosystem: eco,
      name: eco,
      url: networkStatus.value?.effective?.[eco]?.url || ''
    }))
    .filter(t => t.url);

  targets.forEach(t => {
    ecoPingStatus.value[t.ecosystem] = {
      ...ecoPingStatus.value[t.ecosystem],
      loading: true,
      status: 'loading'
    };
  });

  try {
    const res = await fetch("/api/network/ping", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ targets })
    });
    if (res.ok) {
      const data = await res.json();
      const results: any[] = data.results || [];
      results.forEach(r => {
        if (r.ecosystem) {
          ecoPingStatus.value[r.ecosystem] = {
            rtt_ms: r.rtt_ms ?? r.latency_ms ?? 0,
            status: r.status,
            loading: false,
            error: r.error
          };
        }
      });
    }
  } catch (err) {
    console.error(err);
  } finally {
    isProbing.value = false;
    Object.keys(ecoPingStatus.value).forEach(k => {
      if (ecoPingStatus.value[k]?.loading) {
        ecoPingStatus.value[k].loading = false;
      }
    });
  }
}

onMounted(() => {
  loadConfigData();
  loadNetworkStatus();
  fetchHooksList();
  handleDetectProxy(false);
  fetchThemes();
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
  flex-shrink: 0;
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
  flex-shrink: 0;
}

.active-scope-badge {
  font-size: 0.72rem;
  color: var(--text-muted);
  background: var(--input-bg);
  padding: 4px 10px;
  border-radius: 4px;
  border: 1px solid var(--card-border);
  width: 250px;
  flex-shrink: 0;
  box-sizing: border-box;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  user-select: all;
  transition: border-color 0.15s ease, color 0.15s ease;
}

.active-scope-badge:hover {
  border-color: var(--card-border-active);
  color: var(--text-main);
}

@media (max-width: 960px) {
  .active-scope-badge {
    display: none;
  }
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
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 3px 10px;
  background: rgba(34, 211, 238, 0.12);
  border: 1px solid rgba(34, 211, 238, 0.35);
  color: var(--accent-cyan);
  border-radius: 5px;
  font-size: 0.74rem;
  font-weight: 600;
  cursor: pointer;
  margin-left: 2px;
  transition: all 0.2s ease;
}

:root[data-theme="light"] .btn-batch-test,
[data-theme="light"] .btn-batch-test {
  background: rgba(2, 132, 199, 0.1);
  border-color: rgba(2, 132, 199, 0.3);
  color: #0284c7;
}

.btn-batch-test:hover:not(:disabled) {
  background: rgba(34, 211, 238, 0.25);
  transform: translateY(-1px);
}

:root[data-theme="light"] .btn-batch-test:hover:not(:disabled),
[data-theme="light"] .btn-batch-test:hover:not(:disabled) {
  background: rgba(2, 132, 199, 0.18);
}

.btn-batch-test:disabled {
  opacity: 0.55;
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


.proxy-toggle-row {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
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

.btn-detect-proxy {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  color: var(--accent-cyan);
  border-radius: 5px;
  font-size: 0.76rem;
  font-weight: 600;
  cursor: pointer;
  margin-left: auto;
  transition: all 0.2s;
}

.btn-detect-proxy:hover:not(:disabled) {
  border-color: var(--accent-cyan);
  background: rgba(34, 211, 238, 0.1);
}

.btn-detect-proxy:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-test-proxy {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 11px;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  color: var(--accent-cyan);
  border-radius: 5px;
  font-size: 0.76rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-test-proxy:hover:not(:disabled) {
  border-color: var(--accent-cyan);
  background: rgba(34, 211, 238, 0.1);
}

.btn-test-proxy:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.proxy-test-status-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 8px 14px;
  border-radius: 6px;
  font-size: 0.8rem;
  margin-top: 10px;
  transition: all 0.2s;
}

.proxy-test-status-banner.status-ok {
  background: rgba(16, 185, 129, 0.1);
  border: 1px solid rgba(16, 185, 129, 0.35);
  color: #10b981;
}

:root[data-theme="light"] .proxy-test-status-banner.status-ok,
[data-theme="light"] .proxy-test-status-banner.status-ok {
  background: #ecfdf5;
  border: 1px solid #6ee7b7;
  color: #047857;
}

.proxy-test-status-banner.status-fail {
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.35);
  color: #ef4444;
}

:root[data-theme="light"] .proxy-test-status-banner.status-fail,
[data-theme="light"] .proxy-test-status-banner.status-fail {
  background: #fef2f2;
  border: 1px solid #fca5a5;
  color: #b91c1c;
}

.test-status-left {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
}

.test-rtt-tag {
  font-size: 0.72rem;
  font-weight: 700;
  padding: 2px 7px;
  border-radius: 4px;
  background: rgba(16, 185, 129, 0.2);
}

:root[data-theme="light"] .test-rtt-tag,
[data-theme="light"] .test-rtt-tag {
  background: #d1fae5;
  color: #047857;
}

.detected-proxy-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 10px 14px;
  background: rgba(34, 211, 238, 0.08);
  border: 1px solid rgba(34, 211, 238, 0.25);
  border-radius: 8px;
  margin-top: 10px;
}

.detected-banner-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.detected-pill {
  font-size: 0.68rem;
  font-weight: 700;
  padding: 2px 7px;
  border-radius: 4px;
  background: var(--accent-cyan);
  color: #050b14;
}

:root[data-theme="light"] .detected-pill,
[data-theme="light"] .detected-pill {
  background: #0284c7;
  color: #ffffff;
}

.detected-desc {
  font-size: 0.8rem;
  color: var(--text-main);
  font-weight: 500;
}

.btn-apply-detected {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 5px 14px;
  background: var(--accent-cyan);
  color: #050b14;
  border: 1px solid transparent;
  border-radius: 6px;
  font-size: 0.78rem;
  font-weight: 700;
  cursor: pointer;
  white-space: nowrap;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.15);
  transition: all 0.2s ease;
}

:root[data-theme="light"] .btn-apply-detected,
[data-theme="light"] .btn-apply-detected {
  background: #0284c7;
  color: #ffffff !important;
  box-shadow: 0 1px 4px rgba(2, 132, 199, 0.3);
}

.btn-apply-detected:hover {
  filter: brightness(1.1);
  transform: translateY(-1px);
}

.btn-apply-detected:active {
  transform: translateY(0);
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

.profile-scroll-nav {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-right: 2px;
}

.btn-scroll-arrow {
  width: 28px;
  height: 28px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  color: var(--text-main);
  cursor: pointer;
  font-size: 1.15rem;
  font-weight: bold;
  line-height: 1;
  transition: all 0.15s ease;
}

.btn-scroll-arrow:hover {
  background: rgba(34, 211, 238, 0.15);
  border-color: var(--accent-cyan);
  color: var(--accent-cyan);
}

.profile-cards-wrapper {
  position: relative;
  width: 100%;
}

.profile-cards-track {
  display: flex;
  gap: 12px;
  overflow-x: auto;
  scroll-behavior: smooth;
  padding: 4px 2px 10px 2px;
  scrollbar-width: thin;
  scrollbar-color: rgba(34, 211, 238, 0.25) transparent;
}

.profile-cards-track::-webkit-scrollbar {
  height: 6px;
}

.profile-cards-track::-webkit-scrollbar-track {
  background: transparent;
}

.profile-cards-track::-webkit-scrollbar-thumb {
  background: rgba(34, 211, 238, 0.2);
  border-radius: 9999px;
}

.profile-cards-track::-webkit-scrollbar-thumb:hover {
  background: rgba(34, 211, 238, 0.4);
}

.profile-card {
  flex: 1 0 240px;
  min-width: 240px;
  max-width: 320px;
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

.modal-backdrop {
  position: fixed;
  inset: 0;
  z-index: 9999;
  background: rgba(0, 0, 0, 0.45);
  backdrop-filter: blur(6px);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}

.modal-dialog {
  width: 100%;
  max-width: 460px;
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  border-radius: 12px;
  box-shadow: 0 16px 40px rgba(0, 0, 0, 0.25);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid var(--card-border);
}

.modal-title-group {
  display: flex;
  align-items: center;
  gap: 10px;
}

.modal-icon {
  font-size: 1.35rem;
}

.modal-header h4 {
  margin: 0;
  font-size: 0.98rem;
  font-weight: 600;
  color: var(--text-main);
}

.modal-subtitle {
  font-size: 0.74rem;
  color: var(--text-muted);
  display: block;
  margin-top: 2px;
}

.btn-modal-close {
  background: transparent;
  border: none;
  font-size: 1.1rem;
  color: var(--text-muted);
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 4px;
}

.btn-modal-close:hover {
  background: rgba(255, 255, 255, 0.08);
  color: var(--text-main);
}

.modal-body {
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.modal-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.modal-field label {
  font-size: 0.8rem;
  font-weight: 500;
  color: var(--text-main);
}

.modal-field .text-danger {
  color: #ef4444;
}

.modal-field .field-hint {
  font-size: 0.72rem;
  color: var(--text-muted);
  line-height: 1.4;
}

.modal-field-checkbox {
  display: flex;
  align-items: center;
  justify-content: flex-start;
  margin-top: 4px;
}

.modal-checkbox-row {
  display: inline-flex !important;
  flex-direction: row !important;
  align-items: center !important;
  justify-content: flex-start !important;
  gap: 8px !important;
  cursor: pointer;
  user-select: none;
  font-size: 0.85rem !important;
  color: var(--text-muted) !important;
  text-align: left !important;
  margin: 0 !important;
  padding: 0 !important;
}

.modal-checkbox-row:hover {
  color: var(--text-main) !important;
}

.modal-checkbox-row input[type="checkbox"] {
  display: inline-block !important;
  margin: 0 !important;
  width: 15px !important;
  height: 15px !important;
  min-width: 15px !important;
  cursor: pointer;
  accent-color: var(--accent-cyan);
  flex-shrink: 0 !important;
}

.modal-checkbox-row span {
  display: inline !important;
  line-height: 1.4 !important;
  text-align: left !important;
}

.modal-footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 10px;
  padding: 12px 20px;
  border-top: 1px solid var(--card-border);
  background: rgba(0, 0, 0, 0.03);
}

.btn-modal-cancel {
  padding: 7px 16px;
  border-radius: 6px;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  color: var(--text-main);
  font-size: 0.8rem;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-modal-cancel:hover {
  background: rgba(255, 255, 255, 0.08);
}

.btn-modal-confirm {
  padding: 7px 20px;
  border-radius: 6px;
  background: var(--accent-cyan);
  border: 1px solid var(--accent-cyan);
  color: #050b14;
  font-size: 0.8rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-modal-confirm:hover:not(:disabled) {
  opacity: 0.9;
  box-shadow: 0 0 14px rgba(0, 229, 255, 0.35);
}

.btn-modal-confirm:disabled {
  opacity: 0.45;
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

.badge-fallback {
  background: rgba(14, 165, 233, 0.15);
  color: #0ea5e9;
  border: 1px dashed rgba(14, 165, 233, 0.45);
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

.btn-delete-eco {
  padding: 4px 8px;
  background: transparent;
  border: 1px solid transparent;
  color: var(--text-muted);
  border-radius: 4px;
  font-size: 0.82rem;
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.2s ease;
}

.btn-delete-eco:hover {
  background: rgba(239, 68, 68, 0.15);
  border-color: rgba(239, 68, 68, 0.35);
  color: #ef4444;
}

.quick-templates-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 6px;
}

.btn-template-chip {
  padding: 4px 10px;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  border-radius: 6px;
  font-size: 0.76rem;
  color: var(--text-main);
  cursor: pointer;
  transition: all 0.2s;
}

.btn-template-chip:hover, .btn-template-chip.active {
  background: rgba(59, 130, 246, 0.15);
  border-color: #3b82f6;
  color: #60a5fa;
}

.btn-eco-ping {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  min-width: 82px;
  height: 30px;
  padding: 0 10px;
  border-radius: 6px;
  font-size: 0.76rem;
  font-weight: 600;
  cursor: pointer;
  border: 1px solid var(--card-border);
  background: var(--input-bg);
  color: var(--text-muted);
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  white-space: nowrap;
  font-family: inherit;
}

.btn-eco-ping:hover:not(:disabled) {
  border-color: rgba(34, 211, 238, 0.4);
  color: var(--accent-cyan);
  background: rgba(34, 211, 238, 0.08);
}

.btn-eco-ping:disabled {
  cursor: wait;
  opacity: 0.8;
}

.btn-eco-ping.ping-fast {
  background: rgba(16, 185, 129, 0.12);
  border-color: rgba(16, 185, 129, 0.35);
  color: #10b981;
}
.btn-eco-ping.ping-fast:hover:not(:disabled) {
  background: rgba(16, 185, 129, 0.22);
}

.btn-eco-ping.ping-normal {
  background: rgba(56, 189, 248, 0.12);
  border-color: rgba(56, 189, 248, 0.35);
  color: #38bdf8;
}
.btn-eco-ping.ping-normal:hover:not(:disabled) {
  background: rgba(56, 189, 248, 0.22);
}

.btn-eco-ping.ping-slow {
  background: rgba(245, 158, 11, 0.12);
  border-color: rgba(245, 158, 11, 0.35);
  color: #f59e0b;
}
.btn-eco-ping.ping-slow:hover:not(:disabled) {
  background: rgba(245, 158, 11, 0.22);
}

.btn-eco-ping.ping-timeout,
.btn-eco-ping.ping-error {
  background: rgba(239, 68, 68, 0.12);
  border-color: rgba(239, 68, 68, 0.35);
  color: #ef4444;
}
.btn-eco-ping.ping-timeout:hover:not(:disabled),
.btn-eco-ping.ping-error:hover:not(:disabled) {
  background: rgba(239, 68, 68, 0.22);
}

.btn-eco-ping .ping-icon.spinning {
  display: inline-block;
  animation: eco-spin 1s linear infinite;
  font-size: 0.8rem;
}

@keyframes eco-spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

/* Theme Management Styles */
.theme-manager-item {
  flex-direction: column;
  align-items: stretch !important;
}

.theme-manager-item .setting-meta.full-width {
  width: 100%;
}

.badge-custom-pkg {
  font-size: 0.72rem;
  padding: 2px 7px;
  border-radius: 4px;
  background: rgba(0, 229, 255, 0.12);
  color: var(--accent-cyan);
  border: 1px solid var(--accent-cyan-glow);
  font-weight: 500;
}

.theme-install-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 10px;
}

.hidden-file-input {
  display: none;
}

.btn-install-vsix {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 6px 14px;
  border-radius: 6px;
  background: var(--input-bg);
  color: var(--text-main);
  border: 1px solid var(--card-border);
  font-size: 0.84rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
}

.btn-install-vsix .btn-icon {
  font-size: 0.95rem;
  line-height: 1;
}

.btn-install-vsix:hover:not(:disabled) {
  border-color: var(--accent-cyan);
  color: var(--accent-cyan);
  background: rgba(0, 229, 255, 0.08);
  box-shadow: 0 0 0 2px var(--accent-cyan-glow);
  transform: translateY(-1px);
}

.btn-install-vsix:active:not(:disabled) {
  transform: translateY(0);
}

.btn-install-vsix:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

:root[data-theme="light"] .btn-install-vsix,
[data-theme="light"] .btn-install-vsix {
  background: #ffffff;
  color: #1e293b;
  border-color: #cbd5e1;
}

:root[data-theme="light"] .btn-install-vsix:hover:not(:disabled),
[data-theme="light"] .btn-install-vsix:hover:not(:disabled) {
  border-color: #0284c7;
  color: #0284c7;
  background: rgba(2, 132, 199, 0.06);
  box-shadow: 0 0 0 2px rgba(2, 132, 199, 0.15);
}

.spinner-small {
  display: inline-block;
  width: 13px;
  height: 13px;
  border: 2px solid var(--card-border);
  border-top-color: var(--accent-cyan);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.install-feedback-msg {
  font-size: 0.82rem;
  color: var(--success);
}

.install-feedback-msg.error {
  color: #ef4444;
}

.installed-themes-container {
  margin-top: 14px;
  background: var(--section-bg);
  border: 1px solid var(--card-border);
  border-radius: 6px;
  padding: 12px;
}

.installed-themes-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.82rem;
  font-weight: 600;
  color: var(--text-muted);
  margin-bottom: 10px;
}

.count-badge {
  background: rgba(255, 255, 255, 0.1);
  padding: 1px 6px;
  border-radius: 10px;
  font-size: 0.75rem;
}

.installed-themes-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.installed-theme-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  border-radius: 6px;
  transition: border-color 0.2s;
}

.installed-theme-card:hover {
  border-color: var(--accent-cyan);
}

.card-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.theme-type-tag {
  font-size: 0.72rem;
  padding: 2px 6px;
  border-radius: 4px;
}

.theme-type-tag.color {
  background: rgba(157, 78, 221, 0.15);
  color: var(--accent-purple);
  border: 1px solid rgba(157, 78, 221, 0.3);
}

.theme-type-tag.icon {
  background: rgba(0, 229, 255, 0.15);
  color: var(--accent-cyan);
  border: 1px solid rgba(0, 229, 255, 0.3);
}

.theme-info-texts {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.theme-card-title {
  font-size: 0.86rem;
  font-weight: 500;
  color: var(--text-main);
}

.theme-card-sub {
  font-size: 0.75rem;
  color: var(--text-muted);
}

.btn-delete-theme {
  background: transparent;
  border: 1px solid rgba(239, 68, 68, 0.4);
  color: #ef4444;
  padding: 4px 10px;
  border-radius: 4px;
  font-size: 0.78rem;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-delete-theme:hover {
  background: rgba(239, 68, 68, 0.15);
  border-color: #ef4444;
}

/* Toolbar for Theme Manager Actions */
.theme-actions-toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin: 14px 0 16px 0;
  flex-wrap: wrap;
}

.btn-jump-marketplace {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background: rgba(0, 229, 255, 0.1);
  border: 1px solid rgba(0, 229, 255, 0.3);
  color: var(--accent-cyan);
  border-radius: 6px;
  font-size: 0.84rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-jump-marketplace:hover {
  background: rgba(0, 229, 255, 0.22);
  border-color: var(--accent-cyan);
  transform: translateY(-1px);
}

.jump-icon {
  font-size: 0.95rem;
}

.local-vsix-upload-wrap {
  display: inline-flex;
}

/* Appearance Segmented Control */
.segmented-control {
  display: inline-flex;
  align-items: center;
  background: var(--input-bg);
  padding: 3px;
  border-radius: 8px;
  border: 1px solid var(--card-border);
  gap: 4px;
}

.segment-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  border: none;
  border-radius: 6px;
  background: transparent;
  color: var(--text-muted);
  font-size: 0.85rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  user-select: none;
}

.segment-btn:hover {
  color: var(--text-main);
  background: rgba(255, 255, 255, 0.05);
}

.segment-btn.active {
  color: var(--text-main);
  background: var(--card-bg);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.18);
  border: 1px solid var(--card-border-active);
  font-weight: 600;
}

/* Release Age Preset Chips */
.release-age-presets-row {
  display: inline-flex;
  gap: 6px;
  flex-wrap: wrap;
}

.preset-chip-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 3px 8px;
  font-size: 0.74rem;
  border-radius: 4px;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  color: var(--text-muted);
  cursor: pointer;
  transition: all 0.15s ease;
}

.preset-chip-btn:hover {
  background: rgba(255, 255, 255, 0.08);
  color: var(--text-main);
  border-color: var(--border-color);
}

.preset-chip-btn.active {
  background: rgba(16, 185, 129, 0.12);
  border-color: rgba(16, 185, 129, 0.35);
  color: #10b981;
  font-weight: 600;
}

:root[data-theme="light"] .preset-chip-btn,
[data-theme="light"] .preset-chip-btn {
  background: #f8fafc;
  border-color: #cbd5e1;
  color: #64748b;
}

:root[data-theme="light"] .preset-chip-btn:hover,
[data-theme="light"] .preset-chip-btn:hover {
  background: #f1f5f9;
  color: #0f172a;
}

:root[data-theme="light"] .preset-chip-btn.active,
[data-theme="light"] .preset-chip-btn.active {
  background: #ecfdf5;
  border-color: #a7f3d0;
  color: #059669;
}

/* Hooks Section Styling */
.btn-bridge-install,
.btn-add-hook {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  font-size: 0.8rem;
  font-weight: 500;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-bridge-install {
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  color: var(--text-muted);
}

.btn-bridge-install:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.08);
  color: var(--text-main);
  border-color: var(--border-color);
}

.btn-bridge-install:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-add-hook {
  background: var(--accent-cyan);
  border: 1px solid transparent;
  color: #0b0f19;
  font-weight: 600;
}

.btn-add-hook:hover {
  filter: brightness(1.1);
  box-shadow: 0 0 12px rgba(6, 182, 212, 0.35);
}

.hooks-container {
  margin-top: 12px;
}

.hooks-loading {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 24px;
  justify-content: center;
  color: var(--text-muted);
}

.hooks-empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: 32px 16px;
  background: rgba(255, 255, 255, 0.02);
  border: 1px dashed var(--card-border);
  border-radius: 8px;
}

.empty-icon-wrap {
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: rgba(6, 182, 212, 0.1);
  color: var(--accent-cyan);
  margin-bottom: 12px;
}

.hooks-empty-state h4 {
  margin: 0 0 6px 0;
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--text-main);
}

.hooks-empty-state p {
  margin: 0 0 16px 0;
  font-size: 0.82rem;
  color: var(--text-muted);
  max-width: 480px;
  line-height: 1.5;
}

.hooks-preset-recommendations {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  justify-content: center;
}

.preset-label {
  font-size: 0.78rem;
  color: var(--text-muted);
  font-weight: 500;
}

.hooks-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(360px, 1fr));
  gap: 12px;
}

.hook-card {
  display: flex;
  flex-direction: column;
  gap: 8px;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  border-radius: 8px;
  padding: 12px 14px;
  transition: all 0.2s ease;
}

.hook-card:hover {
  border-color: var(--border-color);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.hook-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.hook-name-group {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.hook-badge-icon {
  color: var(--accent-cyan);
  display: flex;
  align-items: center;
}

.hook-name {
  font-size: 0.88rem;
  font-weight: 600;
  color: var(--text-main);
}

.hook-bridge-badge {
  font-size: 0.7rem;
  padding: 2px 6px;
  border-radius: 4px;
  font-weight: 500;
}

.hook-bridge-badge.installed {
  background: rgba(16, 185, 129, 0.12);
  color: #10b981;
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.hook-bridge-badge.not-installed {
  background: rgba(245, 158, 11, 0.12);
  color: #f59e0b;
  border: 1px solid rgba(245, 158, 11, 0.3);
}

.hook-card-actions {
  display: flex;
  align-items: center;
  gap: 4px;
}

.btn-icon-action {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  font-size: 0.74rem;
  border-radius: 4px;
  background: transparent;
  border: 1px solid var(--card-border);
  color: var(--text-muted);
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-icon-action:hover {
  background: rgba(255, 255, 255, 0.08);
  color: var(--text-main);
}

.btn-icon-action.danger:hover {
  background: rgba(239, 68, 68, 0.15);
  border-color: rgba(239, 68, 68, 0.4);
  color: #ef4444;
}

.hook-command-wrap {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: rgba(0, 0, 0, 0.25);
  border: 1px solid var(--card-border);
  border-radius: 6px;
  padding: 6px 10px;
  gap: 8px;
}

.hook-command-code {
  font-size: 0.8rem;
  color: #38bdf8;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.hook-command-code code {
  font-family: var(--font-mono, monospace);
}

.btn-copy-code {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 2px 4px;
  border-radius: 3px;
  transition: color 0.15s ease;
}

.btn-copy-code:hover {
  color: var(--text-main);
}

.modal-textarea {
  width: 100%;
  box-sizing: border-box;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  border-radius: 6px;
  color: var(--text-main);
  padding: 8px 10px;
  font-size: 0.85rem;
  resize: vertical;
  line-height: 1.4;
  outline: none;
}

.modal-textarea:focus {
  border-color: var(--accent-cyan);
}

/* Light Theme Overrides */
:root[data-theme="light"] .btn-bridge-install,
[data-theme="light"] .btn-bridge-install {
  background: #f8fafc;
  border-color: #cbd5e1;
  color: #475569;
}

:root[data-theme="light"] .btn-bridge-install:hover:not(:disabled),
[data-theme="light"] .btn-bridge-install:hover:not(:disabled) {
  background: #f1f5f9;
  color: #0f172a;
}

:root[data-theme="light"] .btn-add-hook,
[data-theme="light"] .btn-add-hook {
  background: #0284c7;
  color: #ffffff;
}

:root[data-theme="light"] .hooks-empty-state,
[data-theme="light"] .hooks-empty-state {
  background: #f8fafc;
  border-color: #e2e8f0;
}

:root[data-theme="light"] .empty-icon-wrap,
[data-theme="light"] .empty-icon-wrap {
  background: #e0f2fe;
  color: #0284c7;
}

:root[data-theme="light"] .hook-card,
[data-theme="light"] .hook-card {
  background: #ffffff;
  border-color: #e2e8f0;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

:root[data-theme="light"] .hook-card:hover,
[data-theme="light"] .hook-card:hover {
  border-color: #cbd5e1;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}

:root[data-theme="light"] .hook-command-wrap,
[data-theme="light"] .hook-command-wrap {
  background: #f8fafc;
  border-color: #e2e8f0;
}

:root[data-theme="light"] .hook-command-code,
[data-theme="light"] .hook-command-code {
  color: #0284c7;
}

:root[data-theme="light"] .btn-icon-action,
[data-theme="light"] .btn-icon-action {
  border-color: #cbd5e1;
  color: #64748b;
}

:root[data-theme="light"] .btn-icon-action:hover,
[data-theme="light"] .btn-icon-action:hover {
  background: #f1f5f9;
  color: #0f172a;
}

:root[data-theme="light"] .modal-textarea,
[data-theme="light"] .modal-textarea {
  background: #ffffff;
  border-color: #cbd5e1;
  color: #0f172a;
}
</style>

