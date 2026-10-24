<template>
  <div class="view-container settings-view" id="view-settings" @contextmenu.prevent>
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
          :placeholder="t('settings.search_placeholder') || 'Search settings'"
        />
        <button
          v-if="searchQuery"
          class="btn btn-icon clear-search-btn"
          @click="searchQuery = ''"
          :title="t('common.clear') || 'Clear'"
        >
          <TechIcon id="workbench:close" :size="12" />
        </button>
      </div>

      <!-- Scope Context Info: User vs Workspace -->
      <div class="scope-switcher-group">
        <span
          class="active-scope-badge font-mono"
          :title="
            activeScope === 'user'
              ? configData?.user?.path || '~/.config/unirtm/unirtm.toml'
              : configData?.workspace?.path || './.unirtm.toml'
          "
        >
          <span style="font-weight: 600; color: var(--accent-cyan); margin-right: 6px">
            [{{
              activeScope === "user"
                ? t("settings.scope_user") || "User"
                : t("settings.scope_workspace") || "Workspace"
            }}]
          </span>
          {{
            activeScope === "user"
              ? configData?.user?.path || "~/.config/unirtm/unirtm.toml"
              : configData?.workspace?.path || "./.unirtm.toml"
          }}
        </span>
      </div>

      <!-- Top Right Actions: Auto-save status & Open TOML -->
      <div class="settings-top-actions">
        <span
          class="auto-save-pill"
          :class="[saveStatusType, { saving: isSaving }]"
          style="display: inline-flex; align-items: center; gap: 6px"
        >
          <TechIcon
            :id="
              saveStatusType === 'saving'
                ? 'workbench:sync'
                : saveStatusType === 'success'
                  ? 'workbench:success'
                  : saveStatusType === 'error'
                    ? 'workbench:error'
                    : 'workbench:speed'
            "
            :size="12"
            :class="{ spinning: saveStatusType === 'saving' }"
          />
          <span>{{ saveStatusText || t("settings.realtime_save") || "Changes saved in real-time" }}</span>
        </span>
        <button
          class="btn-raw-toml"
          :class="{ active: showRawTomlModal }"
          @click="showRawTomlModal ? handleCloseTomlEditor() : handleOpenTomlEditor()"
          :title="showRawTomlModal ? t('settings.close_toml') || 'Close TOML' : t('settings.open_toml') || 'Open TOML'"
        >
          <TechIcon id="workbench:edit" :size="14" />
          <span>{{
            showRawTomlModal ? t("settings.close_toml") || "Close TOML" : t("settings.open_toml") || "Open TOML"
          }}</span>
        </button>
      </div>
    </div>

    <!-- In-Place Settings TOML Editor Modal (Does NOT leave settings view) -->
    <Teleport to="body">
      <div v-if="showRawTomlModal" class="config-editor-modal-overlay" @click.self="handleCloseTomlEditor">
        <div class="config-editor-modal" role="dialog" aria-modal="true">
          <div class="config-editor-header">
            <div class="header-left">
              <span class="codicon codicon-file-code file-icon"></span>
              <div class="header-titles">
                <div class="file-name">{{ activeTomlFileName }}</div>
                <div class="file-desc">{{ activeTomlFilePath }}</div>
              </div>
              <span class="config-badge" :class="isTomlExists ? 'badge-active' : 'badge-draft'">
                {{
                  isTomlExists
                    ? t("guard.config.active") || "Active"
                    : t("guard.config.unconfigured") || "Not configured (Using default)"
                }}
              </span>
              <span class="scope-tag font-mono">
                [{{
                  activeScope === "user"
                    ? t("settings.scope_user") || "User"
                    : t("settings.scope_workspace") || "Workspace"
                }}]
              </span>
            </div>
            <div class="header-actions">
              <button
                class="modal-action-btn secondary"
                @click="handleOpenInExplorer(activeTomlFileName)"
                :title="t('guard.config.open_in_explorer_title') || 'Open in Explorer with full editor tabs'"
              >
                <span class="codicon codicon-screen-full"></span>
                <span>{{ t("guard.config.open_in_explorer") || "Fullscreen" }}</span>
              </button>
              <button class="modal-action-btn copy" @click="copyRawToml" :title="t('settings.copy_content') || 'Copy'">
                <span class="codicon" :class="isCopied ? 'codicon-check' : 'codicon-copy'"></span>
                <span>{{ isCopied ? t("settings.copied") || "Copied" : t("common.copy") || "Copy" }}</span>
              </button>
              <button
                class="modal-action-btn close"
                @click="handleCloseTomlEditor"
                :title="t('settings.close_toml') || 'Close (Esc)'"
              >
                <span class="codicon codicon-close"></span>
              </button>
            </div>
          </div>

          <div class="config-editor-body">
            <div v-if="isTomlLoading" class="editor-loading">
              <span class="codicon codicon-loading codicon-modifier-spin"></span>
              <span>{{ t("guard.config.loading") || "Loading configuration file..." }}</span>
            </div>
            <div ref="tomlMonacoContainer" class="config-monaco-host" v-show="!isTomlLoading"></div>
          </div>

          <div class="config-editor-footer">
            <div class="footer-status">
              <span v-if="isTomlDirty" class="status-dirty"
                >● {{ t("guard.config.unsaved_changes") || "Unsaved changes" }}</span
              >
              <span v-else class="status-clean">✔ {{ t("guard.config.all_saved") || "All changes saved" }}</span>
            </div>
            <div class="footer-actions">
              <button class="modal-btn secondary" @click="handleCloseTomlEditor">
                {{ t("common.close") || "Close" }}
              </button>
              <button class="modal-btn primary" :disabled="isTomlSaving" @click="handleSaveToml">
                <span v-if="isTomlSaving" class="codicon codicon-loading codicon-modifier-spin"></span>
                <span v-else class="codicon codicon-save"></span>
                <span>{{ t("common.save") || "Save" }}</span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Main Settings Body: Full-width Content Canvas -->
    <div class="settings-content-canvas" ref="canvasScrollRef">
      <!-- 1. General Settings -->
      <section v-show="isCategoryVisible('general')" id="section-general" class="settings-category-section glass-card">
        <div class="section-header">
          <div class="section-title-wrap">
            <span class="section-icon">
              <TechIcon id="workbench:settings" :size="20" />
            </span>
            <div>
              <h2>{{ t("settings.cat_general") || "General" }}</h2>
              <span class="section-sub">{{
                t("settings.cat_general_sub") || "Language, theme, browser, and cache"
              }}</span>
            </div>
          </div>
        </div>

        <div class="settings-list">
          <!-- Setting: Language -->
          <div class="setting-item" :class="{ highlight: matchesSearch('language') }">
            <div class="setting-meta">
              <span class="setting-title">{{ t("settings.language") || "Language" }}</span>
              <p class="setting-desc">
                {{ t("settings.language_desc") || "Display language for the workbench interface." }}
              </p>
            </div>
            <div class="setting-control">
              <CustomSelect :model-value="appLanguage" :options="languageSelectOptions" @change="onLanguageChange" />
            </div>
          </div>

          <!-- Setting: Appearance Mode -->
          <div class="setting-item" :class="{ highlight: matchesSearch('appearance') }">
            <div class="setting-meta">
              <span class="setting-title">{{ t("settings.appearance_mode") || "Appearance Theme" }}</span>
              <p class="setting-desc">
                {{ t("settings.appearance_mode_desc") || "Select your preferred color theme for UniRTM." }}
              </p>
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
                  <span>{{ t("theme.system") || "System Default" }}</span>
                </button>
                <button
                  type="button"
                  class="segment-btn"
                  :class="{ active: activeAppearance === 'light' }"
                  @click="onAppearanceChange('light')"
                  id="btn-appearance-light"
                >
                  <TechIcon id="workbench:about" :size="14" :colored="false" />
                  <span>{{ t("theme.light") || "Light Mode" }}</span>
                </button>
                <button
                  type="button"
                  class="segment-btn"
                  :class="{ active: activeAppearance === 'dark' }"
                  @click="onAppearanceChange('dark')"
                  id="btn-appearance-dark"
                >
                  <TechIcon id="workbench:terminal" :size="14" :colored="false" />
                  <span>{{ t("theme.dark") || "Dark Mode" }}</span>
                </button>
              </div>
            </div>
          </div>

          <!-- Setting: Color Theme -->
          <div class="setting-item" :class="{ highlight: matchesSearch('theme') }">
            <div class="setting-meta">
              <span class="setting-title">{{ t("settings.color_theme") || "Color Theme" }}</span>
              <p class="setting-desc">
                {{
                  t("settings.color_theme_desc") ||
                  "Workbench color theme, 100% compatible with VS Code theme standard."
                }}
              </p>
            </div>
            <div class="setting-control">
              <CustomSelect :model-value="activeColorTheme" :options="colorThemeOptions" @change="onColorThemeChange" />
            </div>
          </div>

          <!-- Setting: Icon Theme -->
          <div class="setting-item" :class="{ highlight: matchesSearch('icon') }">
            <div class="setting-meta">
              <span class="setting-title">{{ t("settings.icon_theme") || "File Icon Theme" }}</span>
              <p class="setting-desc">
                {{
                  t("settings.icon_theme_desc") ||
                  "File and tool icon theme for workbench, 100% aligned with VS Code icon standard."
                }}
              </p>
            </div>
            <div class="setting-control">
              <CustomSelect :model-value="activeIconTheme" :options="iconThemeOptions" @change="onIconThemeChange" />
            </div>
          </div>

          <!-- Setting: Product Icon Theme -->
          <div class="setting-item" :class="{ highlight: matchesSearch('product_icon') || matchesSearch('icon') }">
            <div class="setting-meta">
              <span class="setting-title">{{ t("settings.product_icon_theme") || "Product Icon Theme" }}</span>
              <p class="setting-desc">
                {{
                  t("settings.product_icon_theme_desc") ||
                  "Product and workbench UI icon theme for buttons, actions, and status indicators. Supports Codicons and third-party product icon extensions."
                }}
              </p>
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
                <span class="setting-title">{{
                  t("settings.theme_pkg_manager") || "VS Code Theme Package Manager"
                }}</span>
                <span class="badge-custom-pkg">VSIX / ZIP</span>
              </div>
              <p class="setting-desc">
                {{
                  t("settings.theme_pkg_desc") ||
                  "Import official VS Code extension packages (.vsix) with pure static extraction."
                }}
              </p>

              <!-- Actions Toolbar -->
              <div class="theme-actions-toolbar">
                <button class="btn-jump-marketplace" @click="navigateToMarketplace('themes')">
                  <span class="jump-icon">
                    <TechIcon id="workbench:marketplace" :size="14" />
                  </span>
                  <span>{{
                    t("settings.marketplace_browse_link") || "Browse extension & theme marketplace... ↗"
                  }}</span>
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
                    <span>{{
                      isInstallingTheme
                        ? t("settings.installing") || "Installing package..."
                        : t("settings.install_vsix") || "Install Theme Package (.vsix)"
                    }}</span>
                  </button>
                </div>
              </div>

              <!-- Feedback Alert in Installed Tab -->
              <div v-if="themeInstallMsg" class="install-feedback-banner" :class="{ error: themeInstallError }">
                <span>{{ themeInstallMsg }}</span>
                <button
                  class="btn btn-icon btn-close-msg"
                  @click="themeInstallMsg = ''"
                  :title="t('common.close') || 'Close'"
                >
                  <TechIcon id="workbench:close" :size="12" />
                </button>
              </div>
            </div>
          </div>

          <!-- Setting: Browser (User Scope Only) -->
          <div
            v-if="activeScope === 'user'"
            class="setting-item browser-setting-item"
            :class="{
              highlight: matchesSearch('browser'),
            }"
          >
            <div class="setting-meta">
              <span class="setting-title">{{ t("settings.browser_title") || "Browser" }}</span>
              <p class="setting-desc">
                {{ t("settings.browser_desc") || "Browser used when launching unirtm ui." }}
              </p>
            </div>

            <!-- Browser Controls -->
            <div class="setting-control">
              <div class="browser-input-box">
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
                    :placeholder="
                      t('settings.browser_custom_placeholder') || '/usr/bin/google-chrome-stable or custom path'
                    "
                    @change="saveSetting('settings.browser', customBrowserPath)"
                    @blur="saveSetting('settings.browser', customBrowserPath)"
                    @keydown.enter="saveSetting('settings.browser', customBrowserPath)"
                  />
                </div>
              </div>
            </div>
          </div>

          <!-- Setting: Integrated Terminal (User Scope Only) -->
          <div
            v-if="activeScope === 'user'"
            class="setting-item terminal-setting-item"
            :class="{
              highlight: matchesSearch('terminal') || matchesSearch('shell') || matchesSearch(t('settings.shell')),
            }"
          >
            <div class="setting-meta">
              <span class="setting-title">{{ t("terminal.title") || "Integrated Terminal" }}</span>
              <p class="setting-desc">
                {{
                  t("terminal.desc") ||
                  "Configure the default shell environment for UniRTM integrated terminal and task execution"
                }}
              </p>
              <div
                v-if="terminals.length > 0"
                class="terminal-detected-chips"
                style="margin-top: 8px; display: inline-flex; gap: 6px; flex-wrap: wrap"
              >
                <button
                  v-for="tItem in terminals"
                  :key="tItem.path"
                  type="button"
                  class="preset-chip-btn font-mono"
                  :class="{ active: currentTerminal?.path === tItem.path }"
                  @click="onSelectTerminalSetting(tItem.path)"
                  :title="tItem.path"
                >
                  {{ tItem.name }}
                  <span v-if="tItem.is_default" class="chip-badge">{{
                    t("terminal.default_badge") || "System Default"
                  }}</span>
                </button>
              </div>
            </div>

            <!-- Terminal Controls -->
            <div class="setting-control">
              <div class="terminal-input-box">
                <CustomSelect
                  :model-value="selectedTerminalPath"
                  :options="terminalPresetOptions"
                  @change="onTerminalSelectChange"
                />
                <div v-if="selectedTerminalPath === 'custom'" class="custom-terminal-field" style="margin-top: 6px">
                  <input
                    v-model="customTerminalPath"
                    type="text"
                    class="setting-text-input font-mono"
                    :placeholder="t('settings.shell_placeholder') || '/bin/bash or custom shell path'"
                    @change="saveSetting('settings.terminal', customTerminalPath)"
                    @blur="saveSetting('settings.terminal', customTerminalPath)"
                    @keydown.enter="saveSetting('settings.terminal', customTerminalPath)"
                  />
                </div>
              </div>
            </div>
          </div>

          <!-- Setting: Preferred Editor (User Scope Only) -->
          <div v-if="activeScope === 'user'" class="setting-item" :class="{ highlight: matchesSearch('editor') }">
            <div class="setting-meta">
              <span class="setting-title">{{ t("settings.editor") || "Default Text Editor" }}</span>
              <p class="setting-desc">
                {{
                  t("settings.editor_desc") ||
                  "Preferred program used when launching an external editor to inspect or edit files. System-dependent, user scope only."
                }}
              </p>
              <div
                class="release-age-presets-row"
                style="margin-top: 8px; display: inline-flex; gap: 6px; flex-wrap: wrap"
              >
                <button
                  type="button"
                  class="preset-chip-btn"
                  :class="{ active: editor === 'code' }"
                  @click="applyEditorPreset('code')"
                >
                  <span>VS Code (code)</span>
                </button>
                <button
                  type="button"
                  class="preset-chip-btn"
                  :class="{ active: editor === 'cursor' }"
                  @click="applyEditorPreset('cursor')"
                >
                  <span>Cursor</span>
                </button>
                <button
                  type="button"
                  class="preset-chip-btn"
                  :class="{ active: editor === 'nvim' }"
                  @click="applyEditorPreset('nvim')"
                >
                  <span>Neovim</span>
                </button>
                <button
                  type="button"
                  class="preset-chip-btn"
                  :class="{ active: editor === 'vim' }"
                  @click="applyEditorPreset('vim')"
                >
                  <span>Vim</span>
                </button>
                <button
                  type="button"
                  class="preset-chip-btn"
                  :class="{ active: editor === 'nano' }"
                  @click="applyEditorPreset('nano')"
                >
                  <span>Nano</span>
                </button>
              </div>
            </div>
            <div class="setting-control">
              <input
                v-model="editor"
                type="text"
                class="setting-text-input font-mono"
                :placeholder="t('settings.editor_placeholder') || 'code or /usr/bin/nvim'"
                @change="onEditorChange"
                @blur="onEditorChange"
                @keydown.enter="onEditorChange"
              />
            </div>
          </div>

          <!-- Setting: Minimum Required UniRTM Version (Workspace Scope Only) -->
          <div
            v-if="activeScope === 'workspace'"
            class="setting-item"
            :class="{ highlight: matchesSearch('min_version') || matchesSearch('version') }"
          >
            <div class="setting-meta">
              <span class="setting-title">{{ t("settings.min_version") || "Minimum Version (min_version)" }}</span>
              <p class="setting-desc">
                {{
                  t("settings.min_version_desc") ||
                  "Minimum required UniRTM CLI version for project team collaboration, ensuring development environment consistency. Workspace only."
                }}
              </p>
              <div
                v-if="currentCliVersion"
                class="release-age-presets-row"
                style="margin-top: 8px; display: inline-flex; gap: 6px; flex-wrap: wrap"
              >
                <button
                  type="button"
                  class="preset-chip-btn"
                  :class="{ active: minVersion === currentCliVersion }"
                  @click="applyMinVersionCurrent"
                >
                  <span
                    >{{ t("settings.set_current_version") || "Set to Current Version" }} (v{{
                      currentCliVersion
                    }})</span
                  >
                </button>
                <button v-if="minVersion" type="button" class="preset-chip-btn" @click="clearMinVersion">
                  <span>{{ t("settings.clear_min_version") || "Remove Requirement" }}</span>
                </button>
              </div>
            </div>
            <div class="setting-control">
              <input
                v-model="minVersion"
                type="text"
                class="setting-text-input font-mono"
                :placeholder="
                  t('settings.min_version_placeholder') || 'e.g. 0.35.0 (leave empty to remove requirement)'
                "
                @change="onMinVersionChange"
                @blur="onMinVersionChange"
                @keydown.enter="onMinVersionChange"
              />
            </div>
          </div>

          <!-- Setting: Cache TTL -->
          <div class="setting-item" :class="{ highlight: matchesSearch('cache_ttl') }">
            <div class="setting-meta">
              <span class="setting-title">{{ t("settings.cache_ttl") || "Cache Duration" }}</span>
              <p class="setting-desc">
                {{ t("settings.cache_ttl_desc") || "Local cache duration for tool manifests and version indexes." }}
              </p>
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

          <!-- Setting: Always Keep Download Archives -->
          <div
            class="setting-item"
            :class="{
              highlight: matchesSearch('keep') || matchesSearch('download') || matchesSearch('always_keep_download'),
            }"
          >
            <div class="setting-meta">
              <span class="setting-title">{{ t("settings.always_keep_download") || "Keep Download Archives" }}</span>
              <p class="setting-desc">
                {{
                  t("settings.always_keep_download_desc") ||
                  "Whether to retain downloaded archives (.tar.gz / .zip) after installing tools. Enabling speeds up reinstalls and saves bandwidth; disabling saves local disk space."
                }}
              </p>
            </div>
            <div class="setting-control">
              <label class="toggle-switch-label">
                <input type="checkbox" :checked="alwaysKeepDownload" @change="onAlwaysKeepDownloadToggle" />
                <span class="slider round"></span>
              </label>
            </div>
          </div>

          <!-- Setting: Concurrency (Jobs) -->
          <div class="setting-item" :class="{ highlight: matchesSearch('jobs') || matchesSearch('concurrency') }">
            <div class="setting-meta">
              <span class="setting-title">{{ t("settings.jobs") || "Concurrent Jobs Limit" }}</span>
              <p class="setting-desc">
                {{
                  t("settings.jobs_desc") ||
                  "Maximum concurrent downloads and task executions. 0 defaults to system CPU core count."
                }}
              </p>
            </div>
            <div class="setting-control">
              <input
                v-model.number="jobs"
                type="number"
                min="0"
                class="setting-text-input font-mono"
                :placeholder="t('settings.jobs_auto') || '0 (Auto)'"
                @change="onJobsChange"
                @blur="onJobsChange"
                @keydown.enter="onJobsChange"
              />
            </div>
          </div>

          <!-- Setting: Auto Install Missing Tools -->
          <div class="setting-item" :class="{ highlight: matchesSearch('auto_install') || matchesSearch('install') }">
            <div class="setting-meta">
              <span class="setting-title">{{ t("settings.auto_install") || "Auto Install Tools" }}</span>
              <p class="setting-desc">
                {{
                  t("settings.auto_install_desc") ||
                  "Automatically download and install missing tools in background when executing tasks."
                }}
              </p>
            </div>
            <div class="setting-control">
              <label class="toggle-switch-label">
                <input type="checkbox" :checked="autoInstall" @change="onAutoInstallToggle" />
                <span class="slider round"></span>
              </label>
            </div>
          </div>

          <!-- Setting: Task Timeout -->
          <div class="setting-item" :class="{ highlight: matchesSearch('timeout') || matchesSearch('task_timeout') }">
            <div class="setting-meta">
              <span class="setting-title">{{ t("settings.task_timeout") || "Task Timeout" }}</span>
              <p class="setting-desc">
                {{
                  t("settings.task_timeout_desc") ||
                  "Global execution timeout for tasks (e.g. 10m, 30m, 600s). Leave blank to follow task-level settings."
                }}
              </p>
            </div>
            <div class="setting-control">
              <input
                v-model="taskTimeout"
                type="text"
                class="setting-text-input font-mono"
                :placeholder="t('settings.timeout_placeholder_example') || 'e.g. 10m, 600'"
                @change="onTaskTimeoutChange"
                @blur="onTaskTimeoutChange"
                @keydown.enter="onTaskTimeoutChange"
              />
            </div>
          </div>

          <!-- Setting: HTTP Request Timeout -->
          <div class="setting-item" :class="{ highlight: matchesSearch('http_timeout') || matchesSearch('http') }">
            <div class="setting-meta">
              <span class="setting-title">{{ t("settings.http_timeout") || "HTTP Timeout" }}</span>
              <p class="setting-desc">
                {{
                  t("settings.http_timeout_desc") ||
                  "Timeout limit for downloading tool packages and fetching remote manifests (e.g. 30s, 60s, or 2m)."
                }}
              </p>
              <div
                class="release-age-presets-row"
                style="margin-top: 8px; display: inline-flex; gap: 6px; flex-wrap: wrap"
              >
                <button
                  type="button"
                  class="preset-chip-btn"
                  :class="{ active: (httpTimeout || '30s') === '15s' }"
                  @click="applyHttpTimeoutPreset('15s')"
                >
                  <span>⚡ 15s</span>
                </button>
                <button
                  type="button"
                  class="preset-chip-btn"
                  :class="{ active: (httpTimeout || '30s') === '30s' }"
                  @click="applyHttpTimeoutPreset('30s')"
                >
                  <span>{{ t("settings.timeout_preset_30s") || "🌐 30s (Default)" }}</span>
                </button>
                <button
                  type="button"
                  class="preset-chip-btn"
                  :class="{ active: (httpTimeout || '30s') === '60s' }"
                  @click="applyHttpTimeoutPreset('60s')"
                >
                  <span>⏱️ 60s</span>
                </button>
                <button
                  type="button"
                  class="preset-chip-btn"
                  :class="{ active: (httpTimeout || '30s') === '120s' }"
                  @click="applyHttpTimeoutPreset('120s')"
                >
                  <span>{{ t("settings.timeout_preset_120s") || "📦 120s (Large Files)" }}</span>
                </button>
              </div>
            </div>
            <div class="setting-control">
              <input
                v-model="httpTimeout"
                type="text"
                class="setting-text-input font-mono"
                :placeholder="t('settings.http_timeout_placeholder') || '30s (Default)'"
                @change="onHttpTimeoutChange"
                @blur="onHttpTimeoutChange"
                @keydown.enter="onHttpTimeoutChange"
              />
            </div>
          </div>

          <!-- Setting: Task Output Mode -->
          <div class="setting-item" :class="{ highlight: matchesSearch('output') || matchesSearch('task_output') }">
            <div class="setting-meta">
              <span class="setting-title">{{ t("settings.task_output") || "Task Output Mode" }}</span>
              <p class="setting-desc">
                {{
                  t("settings.task_output_desc") ||
                  "Console output formatting and aggregation mode for concurrent task runs."
                }}
              </p>
            </div>
            <div class="setting-control">
              <CustomSelect :model-value="taskOutput" :options="taskOutputOptions" @change="onTaskOutputChange" />
            </div>
          </div>

          <!-- Setting: Shell (Task Shell) -->
          <div class="setting-item" :class="{ highlight: matchesSearch('shell') }">
            <div class="setting-meta">
              <span class="setting-title">{{ t("settings.shell") || "Default Task Shell" }}</span>
              <p class="setting-desc">
                {{
                  t("settings.shell_desc") ||
                  "Default interpreter used when executing project tasks and lifecycle hook scripts. Leave empty for system default."
                }}
              </p>
            </div>
            <div class="setting-control">
              <div style="display: flex; flex-direction: column; gap: 6px; width: 100%; max-width: 320px">
                <CustomSelect
                  :model-value="selectedShellPreset"
                  :options="shellPresetOptions"
                  @change="onShellPresetChange"
                />
                <input
                  v-if="selectedShellPreset === 'custom'"
                  v-model="customShellPath"
                  type="text"
                  class="setting-text-input font-mono"
                  :placeholder="t('settings.shell_custom_placeholder') || '/bin/bash or custom executable path'"
                  @change="saveSetting('settings.shell', customShellPath)"
                  @blur="saveSetting('settings.shell', customShellPath)"
                  @keydown.enter="saveSetting('settings.shell', customShellPath)"
                />
              </div>
            </div>
          </div>

          <!-- Setting: Minimum Release Age -->
          <div
            class="setting-item"
            :class="{
              highlight: matchesSearch('age') || matchesSearch('minimum_release_age') || matchesSearch('security'),
            }"
          >
            <div class="setting-meta">
              <div class="title-with-badge">
                <span class="setting-title">{{ t("settings.minimum_release_age") || "Release Age Requirement" }}</span>
                <span
                  class="badge-custom-pkg"
                  style="background: rgba(16, 185, 129, 0.12); color: #10b981; border-color: rgba(16, 185, 129, 0.25)"
                >
                  {{ t("settings.sec_guard_badge") || "🛡️ Security Guard" }}
                </span>
              </div>
              <p class="setting-desc">
                {{
                  t("settings.minimum_release_age_desc") ||
                  "Safeguard against supply-chain attacks. Defaults to a 7-day safety quarantine (7d) before adopting newly published releases. Set to 0 to bypass and get absolute latest immediately."
                }}
              </p>
              <div
                class="release-age-presets-row"
                style="margin-top: 8px; display: inline-flex; gap: 6px; flex-wrap: wrap"
              >
                <button
                  type="button"
                  class="preset-chip-btn"
                  :class="{ active: (minimumReleaseAge || '7d') === '7d' }"
                  @click="applyReleaseAgePreset('7d')"
                >
                  <span>🛡️ {{ t("settings.release_age_preset_7d") || "7d (Default)" }}</span>
                </button>
                <button
                  type="button"
                  class="preset-chip-btn"
                  :class="{ active: minimumReleaseAge === '14d' }"
                  @click="applyReleaseAgePreset('14d')"
                >
                  <span>🔒 {{ t("settings.release_age_preset_14d") || "14d (Strict)" }}</span>
                </button>
                <button
                  type="button"
                  class="preset-chip-btn"
                  :class="{ active: minimumReleaseAge === '0' }"
                  @click="applyReleaseAgePreset('0')"
                >
                  <span>⚡ {{ t("settings.release_age_preset_0") || "0 (Immediate)" }}</span>
                </button>
              </div>
            </div>
            <div class="setting-control">
              <input
                v-model="minimumReleaseAge"
                type="text"
                class="setting-text-input font-mono"
                :placeholder="t('settings.minimum_release_age_placeholder') || '7d (Recommended default)'"
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
                <span class="setting-title">{{ t("settings.experimental") || "Experimental Features" }}</span>
                <span
                  class="badge-custom-pkg"
                  style="background: rgba(245, 158, 11, 0.12); color: #d97706; border-color: rgba(245, 158, 11, 0.25)"
                >
                  {{ t("settings.preview_badge") || "🧪 Preview Stage" }}
                </span>
              </div>
              <p class="setting-desc">
                {{
                  t("settings.experimental_desc") || "Enable preview capabilities and experimental runtime features."
                }}
              </p>
            </div>
            <div class="setting-control">
              <label class="toggle-switch-label">
                <input type="checkbox" :checked="experimental" @change="onExperimentalToggle" />
                <span class="slider round"></span>
              </label>
            </div>
          </div>
        </div>
      </section>

      <!-- 1.5 Keyboard Shortcuts Section (VS Code Aligned) -->
      <section
        v-show="isCategoryVisible('shortcuts')"
        id="section-shortcuts"
        class="settings-category-section glass-card"
      >
        <KeyboardShortcutsSettings :active-scope="activeScope" />
      </section>

      <!-- 2. Git & Lifecycle Hooks (Workspace Only) -->
      <section
        v-if="activeScope !== 'user'"
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
              <h2>{{ t("settings.cat_hooks") || "Lifecycle & Git Hooks" }}</h2>
              <span class="section-sub">{{
                t("settings.cat_hooks_sub") || "Git lifecycle event hooks and automation scripts"
              }}</span>
            </div>
          </div>
          <div class="hooks-header-actions" style="display: flex; gap: 8px; align-items: center">
            <button
              type="button"
              class="btn-bridge-install"
              :disabled="isInstallingBridges"
              @click="installAllGitBridges"
              :title="
                t('settings.hooks_install_bridges_tip') ||
                'Automatically inject UniRTM route bridges into .git/hooks directory'
              "
            >
              <TechIcon id="workbench:sync" :size="13" :class="{ spinning: isInstallingBridges }" />
              <span>{{
                isInstallingBridges
                  ? t("settings.hooks_installing_bridges") || "Injecting Git bridges..."
                  : t("settings.hooks_install_bridges") || "Inject Git Bridges"
              }}</span>
            </button>
            <button type="button" class="btn-add-hook" @click="openAddHookModal()">
              <TechIcon id="workbench:add" :size="13" />
              <span>{{ t("settings.hooks_add_btn") || "Add Hook" }}</span>
            </button>
          </div>
        </div>

        <!-- Hooks Content List or Empty State -->
        <div class="hooks-container">
          <div v-if="isLoadingHooks" class="hooks-loading">
            <span class="spinner-small"></span>
            <span>{{ t("settings.hooks_loading") || "Loading lifecycle hooks..." }}</span>
          </div>

          <div v-else-if="hooksList.length === 0" class="hooks-empty-state">
            <div class="empty-icon-wrap">
              <TechIcon id="workbench:git" :size="32" />
            </div>
            <h4>{{ t("settings.hooks_empty_title") || "No lifecycle hooks configured yet" }}</h4>
            <p>
              {{
                t("settings.hooks_empty_desc") ||
                "Configure Git hooks to automatically run formatters, linters, and tests before commits or pushes."
              }}
            </p>
            <div class="hooks-preset-recommendations">
              <span class="preset-label">{{ t("settings.hooks_recommended_label") || "Recommended:" }}</span>
              <button type="button" class="preset-chip-btn" @click="openAddHookModal('pre-commit', 'npm test')">
                <span>{{ t("settings.hook_preset_precommit") || "⚡ pre-commit (Lint & Test)" }}</span>
              </button>
              <button
                type="button"
                class="preset-chip-btn"
                @click="openAddHookModal('commit-msg', 'npx commitlint --edit $1')"
              >
                <span>{{ t("settings.hook_preset_commitmsg") || "📝 commit-msg (Commit Message Format)" }}</span>
              </button>
              <button
                type="button"
                class="preset-chip-btn"
                @click="openAddHookModal('pre-push', 'npm run lint && npm test')"
              >
                <span>{{ t("settings.hook_preset_prepush") || "🛡️ pre-push (Pre-push Full Verification)" }}</span>
              </button>
            </div>
          </div>

          <div v-else class="hooks-grid">
            <div
              v-for="hook in hooksList"
              :key="hook.name"
              class="hook-card"
              :class="{ highlight: matchesSearch(hook.name) || matchesSearch(hook.command) }"
            >
              <div class="hook-card-header">
                <div class="hook-name-group">
                  <span class="hook-badge-icon">
                    <TechIcon id="workbench:git" :size="14" />
                  </span>
                  <span class="hook-name font-mono">{{ hook.name }}</span>
                  <span class="hook-bridge-badge" :class="hook.installed ? 'installed' : 'not-installed'">
                    {{
                      hook.installed
                        ? "✓ " + (t("settings.hooks_badge_bridged") || "Git Bridged")
                        : "⚠ " + (t("settings.hooks_badge_unbridged") || "Not Bridged")
                    }}
                  </span>
                </div>
                <div class="hook-card-actions">
                  <button
                    type="button"
                    class="btn-icon-action"
                    :title="t('settings.hooks_run_tip') || 'Run and test this hook'"
                    @click="runHookTest(hook.name)"
                  >
                    <TechIcon id="workbench:tasks" :size="14" />
                    <span>{{ t("settings.hooks_run_btn") || "Test" }}</span>
                  </button>
                  <button
                    type="button"
                    class="btn-icon-action"
                    :title="t('settings.hooks_edit_tip') || 'Edit hook command'"
                    @click="openEditHookModal(hook)"
                  >
                    <TechIcon id="workbench:edit" :size="14" />
                  </button>
                  <button
                    type="button"
                    class="btn-icon-action danger"
                    :title="t('settings.hooks_delete_tip') || 'Delete this hook'"
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
                  :title="t('settings.copy_command') || 'Copy Command'"
                  @click="copyCommand(hook.command)"
                >
                  <TechIcon id="workbench:edit" :size="12" />
                </button>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- 3. Tool Version Aliases -->
      <section v-show="isCategoryVisible('aliases')" id="section-aliases" class="settings-category-section glass-card">
        <div class="section-header">
          <div class="section-title-wrap">
            <span class="section-icon">
              <TechIcon id="workbench:tag" :size="20" />
            </span>
            <div>
              <h2>{{ t("settings.cat_aliases") || "Tool Version Aliases" }}</h2>
              <span class="section-sub">{{
                t("settings.cat_aliases_sub") ||
                "Semantic version alias mappings for tools (e.g. node.lts -> 20, python.stable -> 3.12)"
              }}</span>
            </div>
          </div>
          <div class="aliases-header-actions" style="display: flex; gap: 8px; align-items: center">
            <button type="button" class="btn-add-alias" @click="openAddAliasModal()">
              <TechIcon id="workbench:add" :size="13" />
              <span>{{ t("settings.aliases_add_btn") || "Add Alias" }}</span>
            </button>
          </div>
        </div>

        <!-- Aliases Container -->
        <div class="aliases-container">
          <div v-if="isLoadingAliases" class="aliases-loading">
            <span class="spinner-small"></span>
            <span>{{ t("settings.aliases_loading") || "Loading tool version aliases..." }}</span>
          </div>

          <!-- Aliases List by Tool Group -->
          <div v-else-if="toolNamesWithAliases.length > 0" class="aliases-groups-list">
            <div
              v-for="tool in toolNamesWithAliases"
              :key="tool"
              class="alias-tool-card"
              :class="{ highlight: matchesSearch(tool) }"
            >
              <div class="alias-tool-header">
                <div class="tool-info-left">
                  <span class="tool-icon-wrapper">
                    <TechIcon :id="tool" :size="18" />
                  </span>
                  <span class="tool-name font-mono">{{ tool }}</span>
                  <span class="alias-count-badge">{{ Object.keys(aliasesData[tool] || {}).length }}</span>
                </div>
                <button
                  type="button"
                  class="btn-tool-add-alias"
                  :title="t('settings.aliases_add_for_tool') || 'Add alias for this tool'"
                  @click="openAddAliasModal(tool)"
                >
                  <TechIcon id="workbench:add" :size="12" />
                  <span>{{ t("settings.aliases_add_short") || "Add" }}</span>
                </button>
              </div>

              <div class="alias-chips-grid">
                <div
                  v-for="(targetVersion, aliasName) in aliasesData[tool]"
                  :key="aliasName"
                  class="alias-chip-item"
                  :class="{ highlight: matchesSearch(String(aliasName)) || matchesSearch(String(targetVersion)) }"
                >
                  <div class="alias-chip-main">
                    <span class="alias-tag-name font-mono">{{ aliasName }}</span>
                    <span class="alias-arrow">→</span>
                    <span class="alias-target-version font-mono" :title="String(targetVersion)">{{
                      targetVersion
                    }}</span>
                  </div>
                  <div class="alias-chip-actions">
                    <button
                      type="button"
                      class="btn-chip-action"
                      :title="t('common.copy') || 'Copy'"
                      @click="copyCommand(String(targetVersion))"
                    >
                      <TechIcon id="workbench:edit" :size="11" />
                    </button>
                    <button
                      type="button"
                      class="btn-chip-action"
                      :title="t('common.edit') || 'Edit'"
                      @click="openEditAliasModal(tool, String(aliasName), String(targetVersion))"
                    >
                      <TechIcon id="workbench:edit" :size="11" />
                    </button>
                    <button
                      type="button"
                      class="btn-chip-action action-delete"
                      :title="t('common.delete') || 'Delete'"
                      @click="deleteAlias(tool, String(aliasName))"
                    >
                      <TechIcon id="workbench:close" :size="11" />
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Empty State -->
          <div v-else class="aliases-empty-state">
            <div class="empty-icon-wrap">
              <TechIcon id="workbench:tag" :size="24" />
            </div>
            <h4>{{ t("settings.aliases_empty_title") || "No Tool Version Aliases" }}</h4>
            <p style="margin: 0">
              {{
                t("settings.aliases_empty_desc") ||
                "Define semantic version mappings for tools, such as mapping lts to 20 or stable to 3.12."
              }}
            </p>
          </div>
        </div>
      </section>

      <!-- 4. Network & Mirrors -->
      <section v-show="isCategoryVisible('network')" id="section-network" class="settings-category-section glass-card">
        <div class="section-header">
          <div class="section-title-wrap">
            <span class="section-icon">
              <TechIcon id="workbench:runtimes" :size="20" />
            </span>
            <div>
              <h2>{{ t("settings.cat_network") || "Network" }}</h2>
              <span class="section-sub">{{ t("settings.cat_network_sub") || "Mirrors and proxies" }}</span>
            </div>
          </div>
        </div>

        <!-- Section A: Network Proxy (Default strictly disabled, empty, no 3rd-party github proxies) -->
        <div class="network-subsection">
          <div class="subsection-title-bar">
            <span class="sub-icon">
              <TechIcon id="workbench:lock" :size="16" />
            </span>
            <h3>{{ t("network.tab_proxy") || "Network Proxy" }}</h3>
            <span class="proxy-badge" :class="networkStatus?.proxy?.enabled ? 'badge-on' : 'badge-off'">
              {{
                networkStatus?.proxy?.enabled
                  ? t("network.proxy_enabled") || "Enable Network Proxy"
                  : t("network.proxy_disabled") || "Direct Mode (Proxy Disabled)"
              }}
            </span>
          </div>

          <div class="proxy-form-grid" v-if="networkStatus">
            <div class="proxy-toggle-row">
              <label class="toggle-switch-label">
                <input type="checkbox" v-model="networkStatus.proxy.enabled" @change="handleSaveProxy" />
                <span class="slider round"></span>
              </label>
              <span class="toggle-text">{{
                networkStatus.proxy.enabled
                  ? t("network.proxy_enabled") || "Enable Network Proxy"
                  : t("network.proxy_disabled") || "Direct Mode (Proxy Disabled)"
              }}</span>

              <button
                type="button"
                class="btn-test-proxy"
                :disabled="
                  isTestingProxy ||
                  !networkStatus.proxy.enabled ||
                  (!networkStatus.proxy.http && !networkStatus.proxy.socks5)
                "
                @click="handleTestProxy"
                :title="t('settings.proxy_test_tip') || 'Send real network request to check proxy connectivity'"
              >
                <span v-if="isTestingProxy" style="display: inline-flex; align-items: center; gap: 4px">
                  <TechIcon id="workbench:sync" :size="13" class="spinning" />
                  <span>{{ t("settings.proxy_testing") || "Testing..." }}</span>
                </span>
                <span v-else style="display: inline-flex; align-items: center; gap: 4px">
                  <TechIcon id="workbench:speed" :size="13" />
                  <span>{{ t("settings.proxy_test_btn") || "Test Connectivity" }}</span>
                </span>
              </button>

              <button
                type="button"
                class="btn-detect-proxy"
                :disabled="isDetectingProxy"
                @click="handleDetectProxy(true)"
                :title="
                  t('settings.proxy_detect_tip') || 'Auto detect running proxy clients on host (sing-box, Clash, etc.)'
                "
              >
                <span v-if="isDetectingProxy" style="display: inline-flex; align-items: center; gap: 4px">
                  <TechIcon id="workbench:sync" :size="13" class="spinning" />
                  <span>{{ t("settings.proxy_detecting") || "Detecting..." }}</span>
                </span>
                <span v-else style="display: inline-flex; align-items: center; gap: 4px">
                  <TechIcon id="workbench:search" :size="13" />
                  <span>{{ t("settings.proxy_detect_btn") || "Auto-detect Host Proxy" }}</span>
                </span>
              </button>
            </div>

            <!-- Detected Proxy Banner -->
            <div v-if="detectedProxy?.detected" class="detected-proxy-banner">
              <div class="detected-banner-left">
                <span class="detected-pill">{{ detectedProxy.type.toUpperCase() }}</span>
                <span class="detected-desc">{{ detectedProxy.description }}</span>
              </div>
              <button type="button" class="btn-apply-detected" @click="applyDetectedProxy">
                <span style="display: inline-flex; align-items: center; gap: 4px">
                  <TechIcon id="workbench:speed" :size="13" />
                  <span>{{ t("settings.proxy_import_btn") || "Import" }}</span>
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
              <div
                v-if="proxyTestResult.success && proxyTestResult.rtt_ms"
                class="test-rtt-tag"
                style="display: inline-flex; align-items: center; gap: 4px"
              >
                <TechIcon id="workbench:speed" :size="12" />
                <span>{{
                  t("settings.proxy_latency", { rtt: proxyTestResult.rtt_ms }) ||
                  "Latency " + proxyTestResult.rtt_ms + "ms"
                }}</span>
              </div>
            </div>

            <div class="proxy-fields" :class="{ 'fields-disabled': !networkStatus.proxy.enabled }">
              <div class="proxy-field-item">
                <label>{{ t("network.proxy_http") || "HTTP / HTTPS Proxy URL" }}</label>
                <input
                  v-model="networkStatus.proxy.http"
                  type="text"
                  class="setting-text-input font-mono"
                  :placeholder="
                    t('settings.proxy_http_placeholder') || 'http://127.0.0.1:7890 (leave blank to disable)'
                  "
                  :disabled="!networkStatus.proxy.enabled"
                  @change="handleSaveProxy"
                  @blur="handleSaveProxy"
                  @keydown.enter="handleSaveProxy"
                />
              </div>

              <div class="proxy-field-item">
                <label>{{ t("network.proxy_socks5") || "SOCKS5 Proxy URL" }}</label>
                <input
                  v-model="networkStatus.proxy.socks5"
                  type="text"
                  class="setting-text-input font-mono"
                  :placeholder="t('settings.proxy_socks5_placeholder') || 'socks5://127.0.0.1:10808 (optional)'"
                  :disabled="!networkStatus.proxy.enabled"
                  @change="handleSaveProxy"
                  @blur="handleSaveProxy"
                  @keydown.enter="handleSaveProxy"
                />
              </div>

              <div class="proxy-field-item full-width">
                <label>{{ t("network.proxy_noproxy") || "Bypass Proxy List (NO_PROXY)" }}</label>
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
            <h3>{{ t("network.tab_mirrors") || "Mirror Matrix" }}</h3>
            <button
              class="btn-batch-test"
              :disabled="isProbing"
              @click="handleRunProbeAll"
              :title="
                t('settings.mirrors_ping_all_tip') || 'Concurrently test latency for all ecosystem mirrors in matrix'
              "
            >
              <span v-if="isProbing" style="display: inline-flex; align-items: center; gap: 4px">
                <TechIcon id="workbench:sync" :size="13" class="spinning" />
                <span>{{ t("settings.mirrors_pinging") || "Testing speed..." }}</span>
              </span>
              <span v-else style="display: inline-flex; align-items: center; gap: 4px">
                <TechIcon id="workbench:speed" :size="13" />
                <span>{{ t("settings.mirrors_ping_all_btn") || "Test All Speeds" }}</span>
              </span>
            </button>

            <div class="profile-actions-bar">
              <div class="profile-scroll-nav" v-if="orderedProfiles.length > 3">
                <button
                  class="btn-scroll-arrow"
                  @click="scrollProfiles('left')"
                  :title="t('settings.scroll_left') || 'Scroll left'"
                >
                  ‹
                </button>
                <button
                  class="btn-scroll-arrow"
                  @click="scrollProfiles('right')"
                  :title="t('settings.scroll_right') || 'Scroll right'"
                >
                  ›
                </button>
              </div>
              <button
                class="btn-profile-action"
                @click="showCreateProfileModal = true"
                :title="t('network.create_profile') || 'New Profile'"
              >
                <TechIcon id="workbench:tools" :size="12" />
                <span>{{ t("network.create_profile") || "New Profile" }}</span>
              </button>
              <button
                class="btn-profile-action"
                @click="openAddEcoModal()"
                :title="t('network.add_ecosystem') || 'Add Ecosystem'"
              >
                <TechIcon id="workbench:runtimes" :size="12" />
                <span>{{ t("network.add_ecosystem") || "Add Ecosystem" }}</span>
              </button>
              <button
                class="btn-profile-action secondary"
                @click="handleResetAllEcosystems()"
                :title="t('network.reset_all_ecosystems') || 'Reset Ecosystems'"
              >
                <TechIcon id="workbench:sync" :size="12" />
                <span>{{ t("network.reset_all_ecosystems") || "Reset Ecosystems" }}</span>
              </button>
              <button
                v-if="networkStatus?.active_profile?.builtin"
                class="btn-profile-action secondary"
                @click="handleResetBuiltinProfile(networkStatus.active_profile_id)"
                :title="t('network.reset_factory') || 'Reset to Factory Defaults'"
              >
                <TechIcon id="workbench:sync" :size="12" />
                <span>{{ t("network.reset_factory") || "Reset to Factory Defaults" }}</span>
              </button>
            </div>
          </div>

          <!-- Profile Cards Track -->
          <div class="profile-cards-wrapper" v-if="networkStatus">
            <div ref="profileTrackRef" class="profile-cards-track" @wheel="onProfileTrackWheel">
              <div
                v-for="{ id: pId, profile: p } in orderedProfiles"
                :key="pId"
                class="profile-card"
                :class="{ active: pId === networkStatus.active_profile_id, builtin: p.builtin }"
                @click="handleUseProfile(String(pId))"
              >
                <div class="profile-card-header">
                  <span class="profile-name">{{ getProfileDisplayName(String(pId), p.name) }}</span>
                  <span v-if="pId === networkStatus.active_profile_id" class="active-tag"
                    >✔ {{ t("network.active_tag") }}</span
                  >
                </div>
                <p class="profile-desc">{{ getProfileDisplayDesc(String(pId), p.description) }}</p>
                <div class="profile-footer">
                  <span class="profile-id font-mono">id: {{ pId }}</span>
                  <span v-if="p.builtin" class="pill-builtin">{{ t("network.tag_builtin") }}</span>
                  <span v-else class="pill-custom">{{ t("network.tag_custom") }}</span>
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
                    <h4>{{ t("network.create_profile") || "New Profile" }}</h4>
                    <span class="modal-subtitle">{{
                      t("settings.profile_modal_sub") ||
                      "Save current network and ecosystem mirror settings as an independent Profile"
                    }}</span>
                  </div>
                </div>
                <button
                  class="btn btn-icon btn-modal-close"
                  @click="showCreateProfileModal = false"
                  :title="t('common.close') || 'Close'"
                >
                  <TechIcon id="workbench:close" :size="14" />
                </button>
              </div>

              <div class="modal-body">
                <div class="modal-field">
                  <label>{{ t("settings.profile_id_label") || "Profile ID" }} <span class="text-danger">*</span></label>
                  <input
                    v-model="newProfileId"
                    type="text"
                    class="setting-text-input font-mono"
                    :placeholder="t('settings.profile_id_placeholder') || 'e.g. company-internal or fast-mirror'"
                    autofocus
                  />
                  <span class="field-hint">{{
                    t("settings.profile_id_hint") ||
                    "Used for CLI switching; alphanumeric characters, hyphens, and underscores only"
                  }}</span>
                </div>

                <div class="modal-field">
                  <label>{{ t("settings.profile_name_label") || "Display Name" }}</label>
                  <input
                    v-model="newProfileName"
                    type="text"
                    class="setting-text-input"
                    :placeholder="t('settings.profile_name_placeholder') || 'e.g. Corporate Direct Line'"
                  />
                </div>
              </div>

              <div class="modal-footer">
                <button class="btn-modal-cancel" @click="showCreateProfileModal = false">
                  {{ t("common.cancel") || "Cancel" }}
                </button>
                <button
                  class="btn-modal-confirm"
                  :disabled="!newProfileId.trim() || isSavingProfile"
                  @click="handleSaveAsProfile"
                >
                  <span v-if="isSavingProfile">{{ t("settings.profile_creating") || "Creating..." }}</span>
                  <span v-else>{{ t("settings.profile_create_and_activate") || "Create & Activate" }}</span>
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
                    <h4>{{ t("network.add_ecosystem_title") || "Add Language Ecosystem Mirror" }}</h4>
                    <span class="modal-subtitle">{{
                      t("network.add_ecosystem_desc") ||
                      "Configure any package manager ecosystem (Composer, Maven, NuGet, Docker, etc.) or custom internal registry"
                    }}</span>
                  </div>
                </div>
                <button
                  class="btn btn-icon btn-modal-close"
                  @click="showAddEcoModal = false"
                  :title="t('common.close') || 'Close'"
                >
                  <TechIcon id="workbench:close" :size="14" />
                </button>
              </div>

              <div class="modal-body">
                <div class="modal-field">
                  <label>{{ t("network.quick_templates") || "Quick Presets" }}</label>
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
                  <label
                    >{{ t("network.eco_identifier") || "Ecosystem Identifier (ID)" }}
                    <span class="text-danger">*</span></label
                  >
                  <input
                    v-model="newEcoId"
                    type="text"
                    class="setting-text-input font-mono"
                    :placeholder="t('settings.eco_name_placeholder') || 'e.g. composer, maven, docker, nuget, etc.'"
                    autofocus
                  />
                  <span class="field-hint">{{
                    t("settings.eco_name_hint") ||
                    "Lowercase letters, numbers, or hyphens (e.g. composer, maven, docker)"
                  }}</span>
                </div>

                <div class="modal-field">
                  <label
                    >{{ t("network.eco_registry_url") || "Registry URL" }} <span class="text-danger">*</span></label
                  >
                  <input
                    v-model="newEcoUrl"
                    type="text"
                    class="setting-text-input font-mono"
                    :placeholder="t('settings.eco_url_placeholder') || 'e.g. https://repo.packagist.org'"
                  />
                </div>
              </div>

              <div class="modal-footer">
                <button class="btn-modal-cancel" @click="showAddEcoModal = false">
                  {{ t("common.cancel") || "Cancel" }}
                </button>
                <button
                  class="btn-modal-confirm"
                  :disabled="!newEcoId.trim() || !newEcoUrl.trim() || isAddingEco"
                  @click="handleAddEcosystem"
                >
                  <span v-if="isAddingEco">{{ t("settings.eco_adding") || "Adding..." }}</span>
                  <span v-else>{{ t("settings.eco_add_and_save") || "Add & Save" }}</span>
                </button>
              </div>
            </div>
          </div>

          <!-- Ecosystem Registries Matrix Table -->
          <div class="registries-matrix" v-if="networkStatus">
            <div v-for="eco in orderedEffectiveEcosystems" :key="eco" class="registry-matrix-row">
              <div class="row-eco-info">
                <span class="eco-badge font-mono">{{ eco.toUpperCase() }}</span>
                <span class="status-indicator-badge" :class="getEcoBadgeClass(eco)" :title="getEcoBadgeTitle(eco)">
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
                    <option v-for="m in recommendedMirrors[eco]" :key="m.url" :value="m.url">
                      {{ m.name }}
                    </option>
                    <option
                      v-if="!isUrlInPresets(eco, networkStatus?.effective?.[eco]?.url)"
                      :value="networkStatus?.effective?.[eco]?.url"
                    >
                      {{ t("network.custom_source") || "Custom Mirror" }}
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
                  <span
                    v-else-if="ecoPingStatus[eco]?.status === 'fast'"
                    class="ping-text ping-fast"
                    style="display: inline-flex; align-items: center; gap: 4px"
                  >
                    <TechIcon id="workbench:speed" :size="11" />
                    <span>{{ ecoPingStatus[eco]?.rtt_ms ?? 0 }}ms</span>
                  </span>
                  <span
                    v-else-if="ecoPingStatus[eco]?.status === 'normal'"
                    class="ping-text ping-normal"
                    style="display: inline-flex; align-items: center; gap: 4px"
                  >
                    <TechIcon id="workbench:speed" :size="11" />
                    <span>{{ ecoPingStatus[eco]?.rtt_ms ?? 0 }}ms</span>
                  </span>
                  <span
                    v-else-if="ecoPingStatus[eco]?.status === 'slow'"
                    class="ping-text ping-slow"
                    style="display: inline-flex; align-items: center; gap: 4px"
                  >
                    <TechIcon id="workbench:speed" :size="11" />
                    <span>{{ ecoPingStatus[eco]?.rtt_ms ?? 0 }}ms</span>
                  </span>
                  <span
                    v-else-if="ecoPingStatus[eco]?.status === 'timeout'"
                    class="ping-text ping-timeout"
                    style="display: inline-flex; align-items: center; gap: 4px"
                  >
                    <TechIcon id="workbench:warning" :size="11" />
                    <span>{{ t("settings.ping_timeout") || "Timeout" }}</span>
                  </span>
                  <span
                    v-else-if="ecoPingStatus[eco]?.status === 'error'"
                    class="ping-text ping-error"
                    style="display: inline-flex; align-items: center; gap: 4px"
                  >
                    <TechIcon id="workbench:warning" :size="11" />
                    <span>{{ t("settings.ping_failed") || "Failed" }}</span>
                  </span>
                  <span v-else class="ping-text ping-idle" style="display: inline-flex; align-items: center; gap: 4px">
                    <TechIcon id="workbench:speed" :size="11" />
                    <span>{{ t("settings.ping_action") || "Ping" }}</span>
                  </span>
                </button>

                <!-- Reset Override Button if Overridden -->
                <button
                  v-if="networkStatus.effective[eco]?.is_overridden"
                  class="btn-reset-override"
                  :title="t('network.restore_inherited') || 'Reset to Preset'"
                  @click="handleResetRegistry(eco)"
                >
                  <TechIcon id="workbench:sync" :size="12" />
                </button>

                <!-- Delete Ecosystem Button -->
                <button
                  class="btn-delete-eco"
                  :title="t('network.delete_ecosystem') || 'Remove from Matrix'"
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
      <section v-show="isCategoryVisible('plugins')" id="section-plugins" class="settings-category-section glass-card">
        <div class="section-header">
          <div class="section-title-wrap">
            <span class="section-icon">
              <TechIcon id="workbench:marketplace" :size="20" />
            </span>
            <div>
              <h2>{{ t("settings.cat_plugins") || "Extensions" }}</h2>
              <span class="section-sub">{{
                t("settings.cat_plugins_sub") || "Plugin configuration and isolation"
              }}</span>
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
              <strong>{{ t("settings.plugin_isolation_title") || "Namespace Isolation" }}</strong>
              <p>
                {{
                  t("settings.plugin_isolation_desc") ||
                  "All extension configurations are strictly isolated under plugins.[id]."
                }}
              </p>
            </div>
          </div>

          <!-- Configured plugins list -->
          <div v-if="configuredPlugins.length > 0" class="configured-plugins-list">
            <div v-for="p in configuredPlugins" :key="p.id" class="plugin-config-card">
              <div class="plugin-card-header">
                <div class="plugin-identity">
                  <span class="plugin-icon">
                    <TechIcon id="workbench:marketplace" :size="16" />
                  </span>
                  <span class="plugin-id font-mono">{{ p.id }}</span>
                  <span class="plugin-version-tag">v{{ p.version || "latest" }}</span>
                </div>
                <span class="plugin-keys-count">{{ Object.keys(p.config || {}).length }}</span>
              </div>

              <div class="plugin-params-table">
                <div v-for="(val, key) in p.config" :key="key" class="param-row">
                  <span class="param-key font-mono">plugins.{{ p.id }}.config.{{ key }}</span>
                  <span class="param-val font-mono">{{ val }}</span>
                </div>
                <div v-if="!p.config || Object.keys(p.config).length === 0" class="empty-param-hint">
                  {{ t("settings.empty_toml") || "# No custom settings yet" }}
                </div>
              </div>
            </div>
          </div>

          <div v-else class="no-plugins-card">
            <span class="empty-icon">
              <TechIcon id="workbench:marketplace" :size="32" />
            </span>
            <span>{{ t("settings.no_plugins") || "No plugin configurations yet." }}</span>
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
              <h2>{{ t("settings.cat_security") || "Security" }}</h2>
              <span class="section-sub">{{ t("settings.cat_security_sub") || "Reserved tables and file locks" }}</span>
            </div>
          </div>
        </div>

        <div class="security-architecture-grid">
          <div class="sec-card">
            <div class="sec-badge green">{{ t("settings.status_protected") || "Protected" }}</div>
            <h3>{{ t("settings.sec_core_title") || "Reserved Tables" }}</h3>
            <p>
              {{
                t("settings.sec_core_desc") ||
                "Strictly protects tools, tasks, format, env, settings from third-party overwrites."
              }}
            </p>
          </div>

          <div class="sec-card">
            <div class="sec-badge cyan">{{ t("settings.status_active") || "Active" }}</div>
            <h3>{{ t("settings.sec_lock_title") || "File Lock" }}</h3>
            <p>
              {{ t("settings.sec_lock_desc") || "Acquires an exclusive lock to prevent concurrent write corruption." }}
            </p>
          </div>

          <div class="sec-card">
            <div class="sec-badge purple">{{ t("settings.status_ast_safe") || "AST Safe" }}</div>
            <h3>{{ t("settings.sec_ast_title") || "Comment Preservation" }}</h3>
            <p>
              {{ t("settings.sec_ast_desc") || "Updates preserve existing comments and formatting." }}
            </p>
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
                <h4>
                  {{
                    isEditingHook
                      ? t("settings.hooks_modal_edit_title") || "Edit Lifecycle Hook"
                      : t("settings.hooks_modal_add_title") || "Add Lifecycle Hook"
                  }}
                </h4>
                <span class="modal-subtitle">{{
                  t("settings.hooks_modal_sub") || "Specify Git lifecycle trigger stage and shell command to execute"
                }}</span>
              </div>
            </div>
            <button
              class="btn btn-icon btn-modal-close"
              @click="showHookModal = false"
              :title="t('common.close') || 'Close'"
            >
              <TechIcon id="workbench:close" :size="14" />
            </button>
          </div>
          <div class="modal-body">
            <div class="modal-field">
              <label class="modal-label">{{ t("settings.hooks_form_stage") || "Git Hook Stage" }}</label>
              <CustomSelect
                v-if="!isEditingHook"
                :model-value="hookFormName"
                :options="standardHookPresets"
                @change="(val) => (hookFormName = val)"
              />
              <input
                v-if="hookFormName === 'custom' || isEditingHook"
                v-model="hookFormCustomName"
                type="text"
                class="modal-input font-mono"
                style="margin-top: 6px"
                :placeholder="
                  t('settings.hooks_form_name_placeholder') || 'Enter hook name (e.g. pre-commit, commit-msg)'
                "
                :disabled="isEditingHook"
              />
            </div>

            <div class="modal-field">
              <label class="modal-label">{{ t("settings.hooks_form_cmd") || "Command / Script" }}</label>
              <textarea
                v-model="hookFormCommand"
                class="modal-textarea font-mono"
                rows="3"
                :placeholder="t('settings.hooks_form_cmd_placeholder') || 'e.g. npm test or unirtm run check'"
              ></textarea>
            </div>

            <div v-if="activeScope !== 'user'" class="modal-field-checkbox">
              <label class="modal-checkbox-row" for="hook-auto-bridge">
                <input type="checkbox" id="hook-auto-bridge" v-model="hookFormAutoBridge" />
                <span>{{
                  t("settings.hooks_form_auto_bridge") || "Automatically inject UniRTM bridge into .git/hooks/"
                }}</span>
              </label>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn-modal-cancel" @click="showHookModal = false">
              {{ t("common.cancel") || "Cancel" }}
            </button>
            <button class="btn-modal-confirm" :disabled="isSavingHook" @click="saveHook">
              <span v-if="isSavingHook" class="spinner-small"></span>
              <span>{{ isSavingHook ? t("common.saving") || "Saving..." : t("common.save") || "Save" }}</span>
            </button>
          </div>
        </div>
      </div>

      <!-- Hook Test Run Modal Dialog -->
      <div v-if="showHookRunModal" class="modal-backdrop" @click.self="showHookRunModal = false">
        <div class="modal-dialog glass-card" style="max-width: 600px">
          <div class="modal-header">
            <div class="modal-title-group">
              <span class="modal-icon">
                <TechIcon id="workbench:tasks" :size="20" />
              </span>
              <div>
                <h4>
                  {{ t("settings.hooks_run_modal_title") || "Test Execute Lifecycle Hook" }}:
                  <span class="font-mono text-cyan">{{ runningHookName }}</span>
                </h4>
                <span class="modal-subtitle">{{
                  t("settings.hooks_run_modal_sub") ||
                  "Dispatch and execute via UniRTM multi-engine in current workspace"
                }}</span>
              </div>
            </div>
            <button
              class="btn btn-icon btn-modal-close"
              @click="showHookRunModal = false"
              :title="t('common.close') || 'Close'"
            >
              <TechIcon id="workbench:close" :size="14" />
            </button>
          </div>
          <div class="modal-body">
            <div
              v-if="isRunningHook"
              class="hook-run-loading"
              style="
                display: flex;
                align-items: center;
                gap: 10px;
                padding: 24px;
                justify-content: center;
                color: var(--text-muted);
              "
            >
              <span class="spinner-small"></span>
              <span>{{ t("settings.hooks_running") || "Executing hook script, please wait..." }}</span>
            </div>
            <div v-else-if="hookRunResult" class="hook-run-result">
              <div
                class="run-status-bar"
                style="margin-bottom: 12px; display: flex; align-items: center; justify-content: space-between"
              >
                <span
                  class="run-status-badge"
                  :style="{ color: hookRunResult.success ? '#10b981' : '#ef4444', fontWeight: '600' }"
                >
                  {{
                    hookRunResult.success
                      ? t("settings.hook_run_success") || "✓ Execution Succeeded (Exit Code 0)"
                      : t("settings.hook_run_failed", { code: hookRunResult.exit_code }) ||
                        `✕ Execution Failed (Exit Code ${hookRunResult.exit_code})`
                  }}
                </span>
              </div>
              <div
                class="terminal-output-box"
                style="
                  background: #090d16;
                  border: 1px solid var(--card-border);
                  border-radius: 6px;
                  padding: 12px;
                  font-family: monospace;
                  font-size: 0.85rem;
                  max-height: 300px;
                  overflow-y: auto;
                  color: #e2e8f0;
                  white-space: pre-wrap;
                "
              >
                {{ hookRunResult.output || t("settings.hook_no_stdout") || "(No standard output)" }}
              </div>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn-modal-confirm" @click="showHookRunModal = false">
              {{ t("common.close") || "Close" }}
            </button>
          </div>
        </div>
      </div>

      <!-- Alias Add/Edit Modal Dialog -->
      <div v-if="showAliasModal" class="modal-backdrop" @click.self="showAliasModal = false">
        <div class="modal-dialog glass-card">
          <div class="modal-header">
            <div class="modal-title-group">
              <span class="modal-icon">
                <TechIcon id="workbench:tag" :size="20" />
              </span>
              <div>
                <h4>
                  {{
                    isEditingAlias
                      ? t("settings.aliases_modal_edit_title") || "Edit Version Alias"
                      : t("settings.aliases_modal_add_title") || "Add Version Alias"
                  }}
                </h4>
                <span class="modal-subtitle">{{
                  t("settings.aliases_modal_sub") || "Specify tool, alias identifier and target version"
                }}</span>
              </div>
            </div>
            <button
              class="btn btn-icon btn-modal-close"
              @click="showAliasModal = false"
              :title="t('common.close') || 'Close'"
            >
              <TechIcon id="workbench:close" :size="14" />
            </button>
          </div>
          <div class="modal-body">
            <div class="modal-field">
              <label class="modal-label">{{ t("settings.aliases_form_tool") || "Tool Name" }}</label>
              <CustomSelect
                v-if="!isEditingAlias"
                :model-value="aliasFormTool"
                :options="standardToolPresets"
                @change="(val) => (aliasFormTool = val)"
              />
              <input
                v-if="aliasFormTool === 'custom' || isEditingAlias"
                v-model="aliasFormCustomTool"
                type="text"
                class="modal-input font-mono"
                style="margin-top: 6px"
                :placeholder="t('settings.aliases_form_tool_placeholder') || 'Enter tool name (e.g. node, python, go)'"
                :disabled="isEditingAlias"
              />
            </div>

            <div class="modal-field">
              <label class="modal-label">{{ t("settings.aliases_form_name") || "Alias Name" }}</label>
              <input
                v-model="aliasFormName"
                type="text"
                class="modal-input font-mono"
                :placeholder="t('settings.aliases_form_name_placeholder') || 'e.g. lts, stable, latest, nightly'"
                :disabled="isEditingAlias"
              />
            </div>

            <div class="modal-field">
              <label class="modal-label">{{ t("settings.aliases_form_version") || "Target Version" }}</label>
              <input
                v-model="aliasFormVersion"
                type="text"
                class="modal-input font-mono"
                :placeholder="t('settings.aliases_form_version_placeholder') || 'e.g. 20, 3.12, 1.22.0'"
                @keydown.enter="saveAlias"
              />
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn-modal-cancel" @click="showAliasModal = false">
              {{ t("common.cancel") || "Cancel" }}
            </button>
            <button class="btn-modal-confirm" :disabled="isSavingAlias" @click="saveAlias">
              <span v-if="isSavingAlias" class="spinner-small"></span>
              <span>{{ isSavingAlias ? t("common.saving") || "Saving..." : t("common.save") || "Save" }}</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onActivated, onDeactivated, onBeforeUnmount, watch, nextTick } from "vue";
import { useRoute, useRouter } from "vue-router";
import { loadMonaco } from "../services/editor/monacoLoader";
import { t, selectedLangSetting, setLanguage, SUPPORTED_LANGUAGES } from "../i18n";
import { useTheme } from "../composables/useTheme";
import { useThemeManager } from "../composables/useThemeManager";
import { useToast } from "../composables/useToast";
import { useAppScope } from "../composables/useAppScope";
import { useTerminal } from "../composables/useTerminal";
import CustomSelect from "../components/CustomSelect.vue";
import TechIcon from "../components/common/TechIcon.vue";
import KeyboardShortcutsSettings from "../components/settings/KeyboardShortcutsSettings.vue";

const router = useRouter();
const route = useRoute();
const { showToast } = useToast();
const { activeScope } = useAppScope();

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
  { value: "pre-commit", label: t("settings.hook_opt_precommit") || "pre-commit (Lint & Test)" },
  {
    value: "prepare-commit-msg",
    label: t("settings.hook_opt_preparecommitmsg") || "prepare-commit-msg (Prepare commit message template)",
  },
  { value: "commit-msg", label: t("settings.hook_opt_commitmsg") || "commit-msg (Commit message format check)" },
  { value: "post-commit", label: t("settings.hook_opt_postcommit") || "post-commit (Post-commit action)" },
  { value: "pre-push", label: t("settings.hook_opt_prepush") || "pre-push (Full verification before push)" },
  {
    value: "post-checkout",
    label: t("settings.hook_opt_postcheckout") || "post-checkout (Post-checkout branch maintenance)",
  },
  { value: "post-merge", label: t("settings.hook_opt_postmerge") || "post-merge (Auto pull/install after merge)" },
  { value: "pre-rebase", label: t("settings.hook_opt_prerebase") || "pre-rebase (Safety defense before rebase)" },
  { value: "custom", label: t("settings.hook_opt_custom") || "Custom Name (Custom Hook)..." },
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
  const isStandard = standardHookPresets.some((p) => p.value === hook.name && p.value !== "custom");
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
    showToast(t("settings.toast_hook_name_required") || "Hook name is required", "warning");
    return;
  }
  if (!hookFormCommand.value.trim()) {
    showToast(t("settings.toast_hook_cmd_required") || "Execution command is required", "warning");
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
      showToast(
        isEditingHook.value
          ? t("settings.toast_hook_updated") || "Hook updated successfully"
          : t("settings.toast_hook_created") || "Hook created successfully",
        "success"
      );
      showHookModal.value = false;
      await fetchHooksList();
      await loadConfigData();
    } else {
      const data = await res.json();
      showToast(
        t("settings.status_save_failed", { error: data.error || "Unknown error" }) ||
          `Save failed: ${data.error || "Unknown error"}`,
        "error"
      );
    }
  } catch (err: any) {
    showToast(t("settings.status_save_failed", { error: err.message }) || `Save failed: ${err.message}`, "error");
  } finally {
    isSavingHook.value = false;
  }
}

async function deleteHook(name: string) {
  if (!confirm(t("settings.confirm_remove_hook", { name }) || `Are you sure you want to remove hook ${name}?`)) return;
  try {
    const res = await fetch(`/api/hooks?name=${encodeURIComponent(name)}&scope=${activeScope.value}`, {
      method: "DELETE",
    });
    if (res.ok) {
      showToast(t("settings.toast_removed_name", { name }) || `Removed ${name}`, "success");
      await fetchHooksList();
      await loadConfigData();
    } else {
      const data = await res.json();
      showToast(t("settings.toast_delete_failed", { error: data.error }) || `Failed to delete: ${data.error}`, "error");
    }
  } catch (err: any) {
    showToast(t("settings.toast_delete_failed", { error: err.message }) || `Failed to delete: ${err.message}`, "error");
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
      showToast(
        t("settings.toast_git_bridge_success", { targets: data.installed?.join(", ") || "Active" }) ||
          `✨ Git bridge injected successfully: ${data.installed?.join(", ") || "Active"}`,
        "success"
      );
      await fetchHooksList();
    } else {
      showToast(
        t("settings.toast_bridge_inject_fail", { error: data.error }) || `Failed to inject bridge: ${data.error}`,
        "error"
      );
    }
  } catch (err: any) {
    showToast(
      t("settings.toast_bridge_inject_fail", { error: err.message }) || `Failed to inject bridge: ${err.message}`,
      "error"
    );
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
      output: t("settings.hook_run_failed", { code: err.message }) || `Execution failed: ${err.message}`,
    };
  } finally {
    isRunningHook.value = false;
  }
}

function copyCommand(cmd: string) {
  if (navigator?.clipboard?.writeText) {
    navigator.clipboard.writeText(cmd);
    showToast(t("settings.toast_cmd_copied") || "Command copied to clipboard", "success");
  }
}

// --- Aliases Management ---
const aliasesData = ref<Record<string, Record<string, string>>>({});
const isLoadingAliases = ref(false);
const showAliasModal = ref(false);
const isEditingAlias = ref(false);
const isSavingAlias = ref(false);
const aliasFormTool = ref("node");
const aliasFormCustomTool = ref("");
const aliasFormName = ref("");
const aliasFormVersion = ref("");

const standardToolPresets = [
  { value: "node", label: "Node.js (node)" },
  { value: "python", label: "Python (python)" },
  { value: "go", label: "Go (go)" },
  { value: "rust", label: "Rust (rust)" },
  { value: "bun", label: "Bun (bun)" },
  { value: "deno", label: "Deno (deno)" },
  { value: "java", label: "Java (java)" },
  { value: "pnpm", label: "pnpm (pnpm)" },
  { value: "custom", label: t("settings.aliases_opt_custom") || "Custom Tool..." },
];

const toolNamesWithAliases = computed(() => {
  return Object.keys(aliasesData.value || {}).filter((k) => {
    const group = aliasesData.value[k];
    return group && Object.keys(group).length > 0;
  });
});

async function fetchAliasesList() {
  isLoadingAliases.value = true;
  try {
    const res = await fetch(`/api/aliases?scope=${activeScope.value}`);
    if (res.ok) {
      const data = await res.json();
      aliasesData.value = data.aliases || {};
    }
  } catch (err) {
    console.warn("Failed to fetch /api/aliases:", err);
  } finally {
    isLoadingAliases.value = false;
  }
}

function openAddAliasModal(toolName?: string) {
  isEditingAlias.value = false;
  if (toolName) {
    const isStandard = standardToolPresets.some((p) => p.value === toolName && p.value !== "custom");
    if (isStandard) {
      aliasFormTool.value = toolName;
      aliasFormCustomTool.value = "";
    } else {
      aliasFormTool.value = "custom";
      aliasFormCustomTool.value = toolName;
    }
  } else {
    aliasFormTool.value = "node";
    aliasFormCustomTool.value = "";
  }
  aliasFormName.value = "";
  aliasFormVersion.value = "";
  showAliasModal.value = true;
}

function openEditAliasModal(tool: string, alias: string, version: string) {
  isEditingAlias.value = true;
  const isStandard = standardToolPresets.some((p) => p.value === tool && p.value !== "custom");
  if (isStandard) {
    aliasFormTool.value = tool;
    aliasFormCustomTool.value = tool;
  } else {
    aliasFormTool.value = "custom";
    aliasFormCustomTool.value = tool;
  }
  aliasFormName.value = alias;
  aliasFormVersion.value = version;
  showAliasModal.value = true;
}

async function saveAlias() {
  const finalTool = (aliasFormTool.value === "custom" ? aliasFormCustomTool.value : aliasFormTool.value)
    .trim()
    .toLowerCase();
  const finalName = aliasFormName.value.trim().toLowerCase();
  const finalVer = aliasFormVersion.value.trim();

  if (!finalTool) {
    showToast(t("settings.aliases_err_tool_empty") || "Tool name cannot be empty", "error");
    return;
  }
  if (!finalName) {
    showToast(t("settings.aliases_err_name_empty") || "Alias name cannot be empty", "error");
    return;
  }
  if (!finalVer) {
    showToast(t("settings.aliases_err_ver_empty") || "Target version cannot be empty", "error");
    return;
  }

  isSavingAlias.value = true;
  try {
    const res = await fetch("/api/aliases", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        tool: finalTool,
        alias: finalName,
        version: finalVer,
        scope: activeScope.value,
      }),
    });
    if (!res.ok) {
      const errData = await res.json().catch(() => ({}));
      showToast(errData.error || t("settings.aliases_save_fail") || "Failed to save alias", "error");
      return;
    }
    showToast(t("settings.aliases_save_success") || "Tool version alias saved successfully", "success");
    showAliasModal.value = false;
    await fetchAliasesList();
  } catch (err: any) {
    showToast(err.message || t("settings.aliases_save_fail") || "Failed to save alias", "error");
  } finally {
    isSavingAlias.value = false;
  }
}

async function deleteAlias(tool: string, alias: string) {
  const confirmed = confirm(
    t("settings.aliases_delete_confirm", { tool, alias }) ||
      `Are you sure you want to delete alias ${alias} for tool ${tool}?`
        .replace("{tool}", tool)
        .replace("{alias}", alias)
  );
  if (!confirmed) return;

  try {
    const res = await fetch(
      `/api/aliases?tool=${encodeURIComponent(tool)}&alias=${encodeURIComponent(alias)}&scope=${activeScope.value}`,
      {
        method: "DELETE",
      }
    );
    if (res.ok) {
      showToast(t("settings.aliases_delete_success") || "Alias deleted successfully", "success");
      await fetchAliasesList();
    } else {
      showToast(t("settings.aliases_delete_fail") || "Failed to delete alias", "error");
    }
  } catch (err: any) {
    showToast(err.message || t("settings.aliases_delete_fail") || "Failed to delete alias", "error");
  }
}

watch(activeScope, () => {
  loadConfigData();
  loadNetworkStatus();
  fetchHooksList();
  fetchAliasesList();
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

const defaultShell = ref("");
const selectedShellPreset = ref("");
const customShellPath = ref("");

const { terminals, currentTerminal, fetchTerminals, setPreferredTerminal } = useTerminal();
const customTerminalPath = ref("");

const selectedTerminalPath = computed(() => {
  if (!currentTerminal.value) return "";
  const found = terminals.value.some((t) => t.path === currentTerminal.value?.path);
  if (found) return currentTerminal.value.path;
  return "custom";
});

const terminalPresetOptions = computed(() => {
  const opts: Array<{ value: string; label: string }> = [];
  for (const tItem of terminals.value) {
    const isDef = tItem.is_default ? ` (${t("terminal.default_badge") || "System Default"})` : "";
    opts.push({
      value: tItem.path,
      label: `${tItem.name}${isDef} · ${tItem.path}`,
    });
  }
  opts.push({
    value: "custom",
    label: t("settings.browser_custom") || "Custom Executable",
  });
  return opts;
});

async function onTerminalSelectChange(val: string) {
  if (val === "custom") {
    customTerminalPath.value = currentTerminal.value?.path || "";
  } else {
    customTerminalPath.value = "";
    await setPreferredTerminal(val);
  }
}

async function onSelectTerminalSetting(path: string) {
  customTerminalPath.value = "";
  await setPreferredTerminal(path);
}

const alwaysKeepDownload = ref(false);
const httpTimeout = ref("");
const editor = ref("");
const minVersion = ref("");
const currentCliVersion = ref("");

const shellPresetOptions = computed(() => [
  { value: "", label: t("settings.shell_auto") || "System Default" },
  { value: "bash", label: "Bash (/bin/bash)" },
  { value: "zsh", label: "Zsh (/bin/zsh)" },
  { value: "sh", label: "POSIX Shell (/bin/sh)" },
  { value: "pwsh", label: "PowerShell (pwsh)" },
  { value: "fish", label: "Fish (fish)" },
  { value: "custom", label: t("settings.shell_custom") || "Custom Executable..." },
]);

const taskOutputOptions = computed(() => [
  { value: "interleaved", label: t("settings.task_output_interleaved") || "Interleaved (Real-time live streaming)" },
  { value: "prefix", label: t("settings.task_output_prefix") || "Prefix (Prepend task name to each line)" },
  { value: "quiet", label: t("settings.task_output_quiet") || "Quiet (Output only on failure)" },
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
  status: "idle" | "loading" | "fast" | "normal" | "slow" | "timeout" | "error";
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
    { name: "Official Direct (npmjs.org)", url: "https://registry.npmjs.org" },
    { name: "npmmirror (Taobao/Aliyun)", url: "https://registry.npmmirror.com" },
    { name: "Tencent Cloud", url: "https://mirrors.cloud.tencent.com/npm/" },
    { name: "Huawei Cloud", url: "https://repo.huaweicloud.com/repository/npm/" },
  ],
  go: [
    { name: "Official Direct (Google)", url: "https://proxy.golang.org,direct" },
    { name: "goproxy.cn (Qiniu/Community)", url: "https://goproxy.cn,direct" },
    { name: "Alibaba Cloud GoProxy", url: "https://mirrors.aliyun.com/goproxy/,direct" },
    { name: "goproxy.io", url: "https://goproxy.io,direct" },
  ],
  python: [
    { name: "Official Direct (PyPI)", url: "https://pypi.org/simple" },
    { name: "Tsinghua University (TUNA)", url: "https://pypi.tuna.tsinghua.edu.cn/simple" },
    { name: "USTC", url: "https://mirrors.ustc.edu.cn/pypi/simple" },
    { name: "Alibaba Cloud PyPI", url: "https://mirrors.aliyun.com/pypi/simple/" },
    { name: "Tencent Cloud PyPI", url: "https://mirrors.cloud.tencent.com/pypi/simple/" },
    { name: "Huawei Cloud PyPI", url: "https://repo.huaweicloud.com/repository/pypi/simple/" },
  ],
  rust: [
    { name: "Official Direct (crates.io)", url: "https://crates.io" },
    { name: "rsproxy.cn (ByteDance/Community)", url: "https://rsproxy.cn/crates.io-index" },
    { name: "USTC", url: "https://mirrors.ustc.edu.cn/crates.io-index" },
    { name: "Tsinghua University (TUNA)", url: "https://mirrors.tuna.tsinghua.edu.cn/git/crates.io-index.git" },
  ],
  ruby: [
    { name: "Official Direct (RubyGems)", url: "https://rubygems.org" },
    { name: "Tsinghua University (TUNA)", url: "https://mirrors.tuna.tsinghua.edu.cn/rubygems/" },
    { name: "USTC", url: "https://mirrors.ustc.edu.cn/rubygems/" },
    { name: "Tencent Cloud RubyGems", url: "https://mirrors.cloud.tencent.com/rubygems/" },
  ],
  composer: [
    { name: "Official Direct (Packagist)", url: "https://repo.packagist.org" },
    { name: "Alibaba Cloud Composer", url: "https://mirrors.aliyun.com/composer/" },
    { name: "Tencent Cloud Composer", url: "https://mirrors.cloud.tencent.com/composer/" },
    { name: "Huawei Cloud Composer", url: "https://repo.huaweicloud.com/repository/php/" },
  ],
  maven: [
    { name: "Official Direct (Maven Central)", url: "https://repo1.maven.org/maven2" },
    { name: "Alibaba Cloud Maven", url: "https://maven.aliyun.com/repository/public" },
    { name: "Huawei Cloud Maven", url: "https://repo.huaweicloud.com/repository/maven/" },
    { name: "Tsinghua University (TUNA)", url: "https://mirrors.tuna.tsinghua.edu.cn/maven/" },
  ],
  nuget: [
    { name: "Official Direct (NuGet.org)", url: "https://api.nuget.org/v3/index.json" },
    { name: "Huawei Cloud NuGet", url: "https://repo.huaweicloud.com/repository/nuget/v3/index.json" },
    { name: "Tencent Cloud NuGet", url: "https://mirrors.cloud.tencent.com/nuget/" },
  ],
  docker: [
    { name: "Official Direct (Docker Hub)", url: "https://registry-1.docker.io" },
    { name: "Alibaba Cloud Docker Mirror", url: "https://registry.cn-hangzhou.aliyuncs.com" },
    { name: "NetEase Docker Mirror", url: "https://hub-mirror.c.163.com" },
    { name: "Baidu Docker Mirror", url: "https://mirror.baidubce.com" },
  ],
  dart: [
    { name: "Official Direct (pub.dev)", url: "https://pub.dev" },
    { name: "Tsinghua University (TUNA)", url: "https://mirrors.tuna.tsinghua.edu.cn/dart-pub" },
    { name: "SJTUG", url: "https://mirrors.sjtug.sjtu.edu.cn/dart-pub" },
  ],
  cocoapods: [
    { name: "Official Direct (CocoaPods CDN)", url: "https://cdn.cocoapods.org" },
    { name: "Tsinghua University (TUNA)", url: "https://mirrors.tuna.tsinghua.edu.cn/git/CocoaPods/Specs.git" },
    { name: "USTC", url: "https://mirrors.ustc.edu.cn/CocoaPods-Specs.git" },
  ],
  conda: [
    { name: "Official Direct (Anaconda)", url: "https://repo.anaconda.com/pkgs/main" },
    { name: "Tsinghua University (TUNA)", url: "https://mirrors.tuna.tsinghua.edu.cn/anaconda/pkgs/main" },
    { name: "BFSU", url: "https://mirrors.bfsu.edu.cn/anaconda/pkgs/main" },
    { name: "USTC", url: "https://mirrors.ustc.edu.cn/anaconda/pkgs/main" },
  ],
  homebrew: [
    { name: "Official Direct (Homebrew Bottles)", url: "https://ghcr.io/v2/homebrew/core" },
    { name: "Tsinghua University (TUNA)", url: "https://mirrors.tuna.tsinghua.edu.cn/homebrew-bottles" },
    { name: "USTC", url: "https://mirrors.ustc.edu.cn/homebrew-bottles" },
    { name: "Alibaba Cloud Homebrew", url: "https://mirrors.aliyun.com/homebrew/homebrew-bottles" },
  ],
  cran: [
    { name: "Official Direct (CRAN)", url: "https://cran.r-project.org" },
    { name: "Tsinghua University (TUNA)", url: "https://mirrors.tuna.tsinghua.edu.cn/CRAN/" },
    { name: "USTC", url: "https://mirrors.ustc.edu.cn/CRAN/" },
    { name: "SUSTech", url: "https://mirrors.sustech.edu.cn/CRAN/" },
    { name: "Alibaba Cloud CRAN", url: "https://mirrors.aliyun.com/CRAN/" },
  ],
  julia: [
    { name: "Official Direct (Julia Pkg)", url: "https://pkg.julialang.org" },
    { name: "Tsinghua University (TUNA)", url: "https://mirrors.tuna.tsinghua.edu.cn/julia" },
    { name: "USTC", url: "https://mirrors.ustc.edu.cn/julia" },
    { name: "BFSU", url: "https://mirrors.bfsu.edu.cn/julia" },
  ],
  cpan: [
    { name: "Official Direct (CPAN)", url: "https://www.cpan.org" },
    { name: "Tsinghua University (TUNA)", url: "https://mirrors.tuna.tsinghua.edu.cn/CPAN/" },
    { name: "USTC", url: "https://mirrors.ustc.edu.cn/CPAN/" },
    { name: "Alibaba Cloud CPAN", url: "https://mirrors.aliyun.com/CPAN/" },
  ],
  hackage: [
    { name: "Official Direct (Hackage)", url: "https://hackage.haskell.org" },
    { name: "Tsinghua University (TUNA)", url: "https://mirrors.tuna.tsinghua.edu.cn/hackage" },
    { name: "USTC", url: "https://mirrors.ustc.edu.cn/hackage" },
  ],
  clojars: [
    { name: "Official Direct (Clojars)", url: "https://repo.clojars.org" },
    { name: "Tsinghua University (TUNA)", url: "https://mirrors.tuna.tsinghua.edu.cn/clojars" },
  ],
  ctan: [
    { name: "Official Direct (CTAN)", url: "https://mirror.ctan.org/" },
    { name: "USTC", url: "https://mirrors.ustc.edu.cn/CTAN/" },
    { name: "Tsinghua University (TUNA)", url: "https://mirrors.tuna.tsinghua.edu.cn/CTAN/" },
  ],
  elpa: [
    { name: "Official Direct (GNU ELPA)", url: "https://elpa.gnu.org/packages/" },
    { name: "USTC", url: "https://mirrors.ustc.edu.cn/elpa/" },
    { name: "Tsinghua University (TUNA)", url: "https://mirrors.tuna.tsinghua.edu.cn/elpa/" },
  ],
  stackage: [
    { name: "Official Direct (Stackage)", url: "https://www.stackage.org" },
    { name: "USTC", url: "https://mirrors.ustc.edu.cn/stackage" },
  ],
});

const networkStatus = ref<any>({
  active_profile_id: "official",
  active_profile: {
    id: "official",
    name: t("settings.profile_official_name") || "Official Direct",
    description:
      t("settings.profile_official_desc") ||
      "Full official sources and official binary CDNs without proxies or third-party mirrors",
    builtin: true,
  },
  proxy: {
    enabled: false,
    http: "",
    https: "",
    socks5: "",
    no_proxy:
      "localhost,127.0.0.1,::1,127.0.0.0/8,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16,169.254.0.0/16,*.local,*.internal,*.lan,*.cluster.local",
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
    homebrew: {
      ecosystem: "homebrew",
      url: "https://ghcr.io/v2/homebrew/core",
      source: "profile",
      is_overridden: false,
    },
  },
  profiles: {
    official: {
      id: "official",
      name: t("settings.profile_official_name") || "Official Direct",
      description:
        t("settings.profile_official_desc") ||
        "Full official sources and official binary CDNs without proxies or third-party mirrors",
      builtin: true,
    },
    cn: {
      id: "cn",
      name: t("settings.profile_china_name") || "China Mirror Acceleration",
      description:
        t("settings.profile_china_desc") ||
        "Public mirror matrix from top Chinese universities and cloud providers for rapid pre-build dependency downloads",
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
    return t("network.profile_cn") || "China Fast Mirrors";
  }
  if (id === "official") {
    return t("network.profile_official") || "Official Direct";
  }
  if (id === "tuna") {
    return t("network.profile_tuna") || "Tsinghua TUNA Dedicated";
  }
  if (id === "aliyun") {
    return t("network.profile_aliyun") || "Alibaba Cloud Dedicated";
  }
  return fallbackName || id;
}

function getProfileDisplayDesc(id: string, fallbackDesc?: string): string {
  if (id === "cn") {
    return (
      t("network.profile_cn_desc") ||
      "Curated public mirror matrix from top Chinese universities and cloud providers for rapid dependency retrieval"
    );
  }
  if (id === "official") {
    return (
      t("network.profile_official_desc") ||
      "Full official upstream registries and binary CDNs without proxies or third-party mirrors"
    );
  }
  if (id === "tuna") {
    return (
      t("network.profile_tuna_desc") ||
      "High-speed educational backbone mirrors prioritized via Tsinghua University Open Source Mirror (TUNA)"
    );
  }
  if (id === "aliyun") {
    return (
      t("network.profile_aliyun_desc") ||
      "High-availability package registry mirrors prioritized via Alibaba Cloud CDN infrastructure"
    );
  }
  return fallbackDesc || "";
}

const profileTrackRef = ref<HTMLElement | null>(null);

function scrollProfiles(direction: "left" | "right") {
  if (!profileTrackRef.value) return;
  const scrollAmount = 280;
  profileTrackRef.value.scrollBy({
    left: direction === "left" ? -scrollAmount : scrollAmount,
    behavior: "smooth",
  });
}

function onProfileTrackWheel(e: WheelEvent) {
  if (!profileTrackRef.value) return;
  if (
    Math.abs(e.deltaY) > Math.abs(e.deltaX) &&
    profileTrackRef.value.scrollWidth > profileTrackRef.value.clientWidth
  ) {
    e.preventDefault();
    profileTrackRef.value.scrollLeft += e.deltaY;
  }
}

const orderedEffectiveEcosystems = computed(() => {
  if (!networkStatus.value?.effective) return [];
  const baseOrder = [
    "npm",
    "go",
    "python",
    "rust",
    "ruby",
    "composer",
    "maven",
    "nuget",
    "docker",
    "dart",
    "cocoapods",
    "conda",
    "homebrew",
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
  if (eff?.is_overridden) return "badge-override";
  return "badge-inherited";
}

function getEcoBadgeLabel(eco: string): string {
  const eff = networkStatus.value?.effective?.[eco];
  if (eff?.is_overridden) return t("network.badge_overridden") || "Custom";
  return t("network.badge_inherited") || "Preset";
}

function getEcoBadgeTitle(eco: string): string {
  const eff = networkStatus.value?.effective?.[eco];
  if (eff?.source?.startsWith("fallback_")) {
    return (
      t("network.badge_fallback_cn_tip") ||
      "Not hosted on this dedicated mirror; automatically fallen back to China Mirror Matrix"
    );
  }
  return "";
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
  saveStatusText.value = t("settings.eco_adding") || "Adding...";
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
      saveStatusText.value = t("settings.status_auto_saved") || "Auto saved";
      await loadNetworkStatus();
      await loadConfigData();
      showToast(t("settings.toast_added_eco", { eco }) || `Added ecosystem mirror: ${eco}`, "success");
      setTimeout(() => {
        saveStatusType.value = "idle";
        saveStatusText.value = "";
      }, 2000);
    } else {
      const err = await res.json().catch(() => ({}));
      saveStatusType.value = "error";
      saveStatusText.value =
        t("settings.status_save_failed", { error: err.error || "Unknown error" }) ||
        `Failed to add: ${err.error || "Unknown error"}`;
    }
  } catch (err: any) {
    console.error(err);
    saveStatusType.value = "error";
    saveStatusText.value =
      t("settings.status_network_error", { error: err.message }) || `Network error: ${err.message}`;
  } finally {
    isAddingEco.value = false;
    isSaving.value = false;
  }
}

async function handleDeleteEcosystem(eco: string) {
  isSaving.value = true;
  saveStatusType.value = "saving";
  saveStatusText.value = t("settings.status_removing_eco") || "Removing ecosystem...";
  try {
    const res = await fetch("/api/network/registry/delete", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ ecosystem: eco, scope: activeScope.value }),
    });
    if (res.ok) {
      saveStatusType.value = "success";
      saveStatusText.value = t("settings.status_removed_eco") || "Ecosystem removed";
      await loadNetworkStatus();
      await loadConfigData();
      showToast(t("settings.toast_removed_eco", { eco }) || `Removed ecosystem from matrix: ${eco}`, "success");
      setTimeout(() => {
        saveStatusType.value = "idle";
        saveStatusText.value = "";
      }, 2000);
    } else {
      const err = await res.json().catch(() => ({}));
      saveStatusType.value = "error";
      saveStatusText.value =
        t("settings.status_save_failed", { error: err.error || "Unknown error" }) ||
        `Save failed: ${err.error || "Unknown error"}`;
    }
  } catch (err: any) {
    console.error(err);
    saveStatusType.value = "error";
    saveStatusText.value =
      t("settings.status_network_error", { error: err.message }) || `Network error: ${err.message}`;
  } finally {
    isSaving.value = false;
  }
}

// Full config from API
const configData = ref<any>(null);

const languageSelectOptions = computed(() => [
  { value: "auto", label: t("common.autoDetect") || "Auto Detect" },
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
  // Remove any parentheses and their contents, e.g. (v13.0.0), (VS Code Default), (default), [v1.0]
  name = name.replace(/\s*[\(\[\（\【][^\)\]\）\】]*[\)\]\）\】]\s*/g, " ");
  // Remove bare version strings, e.g. v13.0.0, 1.2.3
  name = name.replace(/\b[vV]?\d+(\.\d+)+(-[a-zA-Z0-9.]+)?\b/g, "");
  // Normalize whitespace
  name = name.trim().replace(/\s+/g, " ");
  return name || rawName;
}

const colorThemeOptions = computed(() => {
  const options: Array<{ value: string; label: string }> = [
    { value: "default", label: t("settings.appearance_unirtm_default") || "UniRTM Default Adaptive" },
    { value: "light", label: t("settings.appearance_unirtm_light") || "UniRTM Official Light" },
    { value: "dark", label: t("settings.appearance_unirtm_dark") || "UniRTM Official Dark" },
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
    { value: "", label: t("settings.browser_auto") || "System Default" },
  ];

  for (const b of detectedBrowsers.value) {
    options.push({
      value: b.id,
      label: b.name,
    });
  }

  options.push({
    value: "custom",
    label: t("settings.browser_custom") || "Custom Executable",
  });

  return options;
});

const currentRawToml = computed(() => {
  if (!configData.value) return "";
  return activeScope.value === "user" ? configData.value?.user?.raw || "" : configData.value?.workspace?.raw || "";
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
      q.includes("general") ||
      q.includes("common") ||
      q.includes("job") ||
      q.includes("concur") ||
      q.includes("task") ||
      q.includes("timeout") ||
      q.includes("auto") ||
      q.includes("exper") ||
      q.includes("age") ||
      q.includes(t("settings.tab_general").toLowerCase()) ||
      q.includes(t("settings.language").toLowerCase()) ||
      q.includes(t("settings.theme").toLowerCase()))
  )
    return true;
  if (
    catId === "shortcuts" &&
    (q.includes("short") ||
      q.includes("key") ||
      q.includes("bind") ||
      q.includes("hotkey") ||
      q.includes(t("settings.cat_shortcuts").toLowerCase()) ||
      q.includes("ctrl") ||
      q.includes("cmd") ||
      q.includes("alt") ||
      q.includes("shift"))
  )
    return true;
  if (
    catId === "hooks" &&
    (q.includes("hook") ||
      q.includes("git") ||
      q.includes("commit") ||
      q.includes("push") ||
      q.includes(t("settings.cat_hooks").toLowerCase()))
  )
    return true;
  if (
    catId === "network" &&
    (q.includes("proxy") ||
      q.includes("github") ||
      q.includes("mirror") ||
      q.includes("net") ||
      q.includes(t("settings.cat_network").toLowerCase()))
  )
    return true;
  if (
    catId === "plugins" &&
    (q.includes("plugin") ||
      q.includes("ext") ||
      q.includes("docker") ||
      q.includes(t("settings.cat_plugins").toLowerCase()))
  )
    return true;
  if (
    catId === "aliases" &&
    (q.includes("alias") || q.includes("version") || q.includes(t("settings.cat_aliases").toLowerCase()))
  )
    return true;
  if (
    catId === "security" &&
    (q.includes("sec") ||
      q.includes("lock") ||
      q.includes("ast") ||
      q.includes("namespace") ||
      q.includes("safe") ||
      q.includes(t("settings.cat_security").toLowerCase()))
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
  minimumReleaseAge.value =
    s.minimum_release_age !== undefined && s.minimum_release_age !== "" ? s.minimum_release_age : "7d";
  experimental.value = Boolean(s.experimental);

  alwaysKeepDownload.value = Boolean(s.always_keep_download);
  httpTimeout.value = s.http_timeout !== undefined ? String(s.http_timeout) : "";
  const sh = s.shell || "";
  const knownShells = ["", "bash", "zsh", "sh", "pwsh", "fish"];
  if (knownShells.includes(sh)) {
    selectedShellPreset.value = sh;
    customShellPath.value = "";
  } else {
    selectedShellPreset.value = "custom";
    customShellPath.value = sh;
  }
  defaultShell.value = sh;

  if (activeScope.value === "user") {
    editor.value = s.editor || "";
    const b = s.browser || "";
    const knownIds = [
      "chrome",
      "firefox",
      "msedge",
      "safari",
      "brave",
      "chromium",
      "arc",
      "opera",
      "vivaldi",
      "epiphany",
      "",
    ];
    const isDetectedOrKnown =
      knownIds.includes(b.toLowerCase()) || detectedBrowsers.value.some((d) => d.id === b.toLowerCase());
    if (isDetectedOrKnown) {
      selectedBrowserPreset.value = b;
      customBrowserPath.value = "";
    } else {
      selectedBrowserPreset.value = "custom";
      customBrowserPath.value = b;
    }
  } else {
    minVersion.value = target?.min_version || "";
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

function onShellPresetChange(val: string) {
  selectedShellPreset.value = val;
  if (val !== "custom") {
    defaultShell.value = val;
    customShellPath.value = "";
    saveSetting("settings.shell", val);
  }
}

function onAlwaysKeepDownloadToggle() {
  alwaysKeepDownload.value = !alwaysKeepDownload.value;
  saveSetting("settings.always_keep_download", alwaysKeepDownload.value);
}

function onHttpTimeoutChange() {
  const val = String(httpTimeout.value).trim();
  httpTimeout.value = val;
  saveSetting("settings.http_timeout", val);
}

function applyHttpTimeoutPreset(preset: string) {
  httpTimeout.value = preset;
  saveSetting("settings.http_timeout", preset);
}

function onEditorChange() {
  const val = String(editor.value).trim();
  editor.value = val;
  saveSetting("settings.editor", val);
}

function applyEditorPreset(preset: string) {
  editor.value = preset;
  saveSetting("settings.editor", preset);
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

function onMinVersionChange() {
  const clean = minVersion.value.trim().replace(/^v/, "");
  minVersion.value = clean;
  saveSetting("min_version", clean);
}

function applyMinVersionCurrent() {
  if (!currentCliVersion.value) return;
  minVersion.value = currentCliVersion.value;
  saveSetting("min_version", currentCliVersion.value);
}

function clearMinVersion() {
  minVersion.value = "";
  saveSetting("min_version", "");
}

async function loadSystemStatus() {
  try {
    const res = await fetch("/api/status");
    if (res.ok) {
      const data = await res.json();
      if (data.version && data.version !== "N/A") {
        currentCliVersion.value = data.version.replace(/^v/, "");
      }
    }
  } catch (e) {
    console.warn("Failed to fetch /api/status:", e);
  }
}

async function saveSetting(key: string, value: any) {
  isSaving.value = true;
  saveStatusType.value = "saving";
  saveStatusText.value = t("settings.status_saving") || "Saving...";

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
      saveStatusText.value = t("settings.status_auto_saved") || "Auto saved";
      await loadConfigData();
      setTimeout(() => {
        saveStatusType.value = "idle";
        saveStatusText.value = "";
      }, 2000);
    } else {
      isSaving.value = false;
      saveStatusType.value = "error";
      saveStatusText.value =
        t("settings.status_save_failed", { error: data.error || "Unknown error" }) ||
        `Save failed: ${data.error || "Unknown error"}`;
    }
  } catch (err: any) {
    isSaving.value = false;
    saveStatusType.value = "error";
    saveStatusText.value =
      t("settings.status_network_error", { error: err.message }) || `Network error: ${err.message}`;
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
    if (colorItems.length > 0)
      parts.push(
        t("settings.theme_import_colors", { count: colorItems.length }) || `${colorItems.length} color theme(s)`
      );
    if (iconItems.length > 0)
      parts.push(t("settings.theme_import_icons", { count: iconItems.length }) || `${iconItems.length} file icon(s)`);
    if (productItems.length > 0)
      parts.push(
        t("settings.theme_import_products", { count: productItems.length }) || `${productItems.length} product icon(s)`
      );

    const typeSummary =
      parts.length > 0
        ? parts.join(", ")
        : t("settings.theme_import_items", { count: installed.length }) || `${installed.length} items`;
    themeInstallMsg.value =
      t("settings.theme_import_success", { summary: typeSummary }) || `✓ Successfully imported ${typeSummary}!`;

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
    showToast(
      t("settings.toast_theme_imported", { summary: typeSummary }) ||
        `Successfully imported VS Code extension package, identified: ${typeSummary}`,
      "success"
    );
  } catch (err: any) {
    themeInstallError.value = true;
    themeInstallMsg.value =
      t("settings.theme_install_failed", { error: err.message || err }) || `✕ Failed to install: ${err.message || err}`;
    showToast(
      t("settings.toast_import_failed", { error: err.message || err }) || `Failed to import: ${err.message || err}`,
      "error"
    );
  } finally {
    isInstallingTheme.value = false;
    target.value = "";
  }
}

async function onAppearanceChange(mode: "system" | "light" | "dark") {
  isSaving.value = true;
  saveStatusType.value = "saving";
  saveStatusText.value = t("settings.status_changing_appearance") || "Changing appearance...";
  try {
    await setActiveTheme(undefined, undefined, mode);
    saveStatusType.value = "success";
    saveStatusText.value = t("settings.status_auto_saved") || "Auto saved";
    await loadConfigData();
    showToast(
      mode === "system"
        ? t("settings.appearance_follow_system") || "Set to follow system appearance"
        : mode === "light"
          ? t("settings.appearance_switched_light") || "Switched to light mode"
          : t("settings.appearance_switched_dark") || "Switched to dark mode",
      "info"
    );
    setTimeout(() => {
      saveStatusType.value = "idle";
      saveStatusText.value = "";
    }, 2000);
  } catch (err: any) {
    saveStatusType.value = "error";
    saveStatusText.value =
      t("settings.status_save_failed", { error: err.message || err }) || `Save failed: ${err.message || err}`;
  } finally {
    isSaving.value = false;
  }
}

async function onColorThemeChange(val: string) {
  isSaving.value = true;
  saveStatusType.value = "saving";
  saveStatusText.value = t("settings.status_applying_theme") || "Applying theme...";
  try {
    await setActiveTheme(val, undefined, undefined);
    appTheme.value = val;
    const themeObj = colorThemes.value.find((c) => c.id === val);
    const themeName = cleanThemeLabel(themeObj ? themeObj.name : val);
    saveStatusType.value = "success";
    saveStatusText.value = t("settings.status_auto_saved") || "Auto saved";
    await loadConfigData();
    showToast(
      t("settings.toast_applied_color_theme", { name: themeName }) || `Applied color theme: ${themeName}`,
      "success"
    );
    setTimeout(() => {
      saveStatusType.value = "idle";
      saveStatusText.value = "";
    }, 2000);
  } catch (err: any) {
    saveStatusType.value = "error";
    saveStatusText.value =
      t("settings.status_save_failed", { error: err.message || err }) || `Save failed: ${err.message || err}`;
  } finally {
    isSaving.value = false;
  }
}

async function onIconThemeChange(val: string) {
  isSaving.value = true;
  saveStatusType.value = "saving";
  saveStatusText.value = t("settings.status_applying_file_icons") || "Applying file icons...";
  try {
    await setActiveTheme(undefined, val, undefined);
    const iconObj = iconThemes.value.find((i) => i.id === val);
    const iconName = cleanThemeLabel(iconObj ? iconObj.name : val);
    saveStatusType.value = "success";
    saveStatusText.value = t("settings.status_auto_saved") || "Auto saved";
    await loadConfigData();
    showToast(
      t("settings.toast_applied_icon_theme", { name: iconName }) || `Applied file icon theme: ${iconName}`,
      "success"
    );
    setTimeout(() => {
      saveStatusType.value = "idle";
      saveStatusText.value = "";
    }, 2000);
  } catch (err: any) {
    saveStatusType.value = "error";
    saveStatusText.value =
      t("settings.status_save_failed", { error: err.message || err }) || `Save failed: ${err.message || err}`;
  } finally {
    isSaving.value = false;
  }
}

async function onProductIconThemeChange(val: string) {
  isSaving.value = true;
  saveStatusType.value = "saving";
  saveStatusText.value = t("settings.status_applying_product_icons") || "Applying product icons...";
  try {
    await setActiveTheme(undefined, undefined, undefined, val);
    const prodObj = productIconThemes.value.find((p) => p.id === val);
    const prodName = cleanThemeLabel(prodObj ? prodObj.name : val);
    saveStatusType.value = "success";
    saveStatusText.value = t("settings.status_auto_saved") || "Auto saved";
    await loadConfigData();
    showToast(
      t("settings.toast_applied_product_theme", { name: prodName }) || `Applied product icon theme: ${prodName}`,
      "success"
    );
    setTimeout(() => {
      saveStatusType.value = "idle";
      saveStatusText.value = "";
    }, 2000);
  } catch (err: any) {
    saveStatusType.value = "error";
    saveStatusText.value =
      t("settings.status_save_failed", { error: err.message || err }) || `Save failed: ${err.message || err}`;
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

const tomlMonacoContainer = ref<HTMLDivElement | null>(null);
let tomlMonacoEditor: any = null;
let tomlMonacoModel: any = null;
let originalTomlContent = "";
const isTomlLoading = ref(false);
const isTomlSaving = ref(false);
const isTomlDirty = ref(false);

const activeTomlFileName = computed(() => {
  return activeScope.value === "user" ? "unirtm.toml" : ".unirtm.toml";
});

const activeTomlFilePath = computed(() => {
  return activeScope.value === "user"
    ? configData.value?.user?.path || "~/.config/unirtm/unirtm.toml"
    : configData.value?.workspace?.path || "./.unirtm.toml";
});

const isTomlExists = computed(() => {
  return activeScope.value === "user" ? !!configData.value?.user?.exists : !!configData.value?.workspace?.exists;
});

async function handleOpenTomlEditor() {
  showRawTomlModal.value = true;
  isTomlLoading.value = true;
  isTomlDirty.value = false;

  let content = "";
  try {
    const res = await fetch(`/api/config/raw?scope=${activeScope.value}`);
    if (res.ok) {
      const data = await res.json();
      content = data.content ?? "";
    } else {
      content = currentRawToml.value || "# UniRTM Configuration\n";
    }
  } catch {
    content = currentRawToml.value || "# UniRTM Configuration\n";
  } finally {
    isTomlLoading.value = false;
  }

  originalTomlContent = content;

  await nextTick();
  if (!tomlMonacoContainer.value) return;

  const monaco = await loadMonaco();
  const isDark = document.documentElement.getAttribute("data-theme") !== "light";

  if (tomlMonacoModel) {
    tomlMonacoModel.dispose();
  }
  tomlMonacoModel = monaco.editor.createModel(content, "toml");

  if (!tomlMonacoEditor) {
    tomlMonacoEditor = monaco.editor.create(tomlMonacoContainer.value, {
      model: tomlMonacoModel,
      theme: isDark ? "vs-dark" : "vs",
      automaticLayout: true,
      fontSize: 13,
      lineNumbers: "on",
      minimap: { enabled: false },
      scrollBeyondLastLine: false,
      wordWrap: "on",
      tabSize: 2,
    });

    tomlMonacoEditor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyS, () => {
      handleSaveToml();
    });
  } else {
    tomlMonacoEditor.setModel(tomlMonacoModel);
    monaco.editor.setTheme(isDark ? "vs-dark" : "vs");
  }

  tomlMonacoModel.onDidChangeContent(() => {
    if (!tomlMonacoModel) return;
    isTomlDirty.value = tomlMonacoModel.getValue() !== originalTomlContent;
  });
}

async function handleSaveToml() {
  if (!tomlMonacoModel) return;
  isTomlSaving.value = true;
  const content = tomlMonacoModel.getValue();
  try {
    const res = await fetch("/api/config/raw", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        scope: activeScope.value,
        content,
      }),
    });
    const data = await res.json();
    if (res.ok) {
      originalTomlContent = content;
      isTomlDirty.value = false;
      showToast(t("settings.save_toml_success") || "Configuration file saved successfully", "success");
      await loadConfigData();
    } else {
      showToast(data.error || t("settings.save_toml_failed") || "Failed to save configuration file", "error");
    }
  } catch (err: any) {
    showToast(t("settings.status_save_failed", { error: err.message }) || `Save failed: ${err.message}`, "error");
  } finally {
    isTomlSaving.value = false;
  }
}

function handleCloseTomlEditor() {
  if (tomlMonacoModel) {
    tomlMonacoModel.dispose();
    tomlMonacoModel = null;
  }
  if (tomlMonacoEditor) {
    tomlMonacoEditor.dispose();
    tomlMonacoEditor = null;
  }
  showRawTomlModal.value = false;
  isTomlDirty.value = false;
  if (route.name === "settings" && (route.query.open_toml || route.query.file)) {
    const q = { ...route.query };
    delete q.open_toml;
    delete q.file;
    router.replace({ query: q });
  }
}

function handleOpenInExplorer(fileName: string) {
  if (tomlMonacoModel) {
    tomlMonacoModel.dispose();
    tomlMonacoModel = null;
  }
  if (tomlMonacoEditor) {
    tomlMonacoEditor.dispose();
    tomlMonacoEditor = null;
  }
  showRawTomlModal.value = false;
  isTomlDirty.value = false;

  router
    .push({
      name: "explorer",
      query: {
        from: "settings",
        file: fileName,
        scope: activeScope.value,
      },
    })
    .catch(() => {});
  showToast(t("settings.toast_opened_in_editor", { fileName }) || `Opened ${fileName} in code editor`, "info");
}

function copyRawToml() {
  const content = tomlMonacoModel ? tomlMonacoModel.getValue() : currentRawToml.value;
  if (content) {
    navigator.clipboard.writeText(content);
    isCopied.value = true;
    showToast(t("settings.copied") || "Copied", "success");
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
  saveStatusText.value = t("settings.status_switching_profile") || "Switching profile...";
  try {
    const res = await fetch("/api/network/profile/use", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ profile_id: id, scope: activeScope.value }),
    });
    const data = await res.json().catch(() => ({}));
    if (res.ok) {
      saveStatusType.value = "success";
      saveStatusText.value = t("settings.status_applied_profile") || "Profile applied";
      await loadNetworkStatus();
      await loadConfigData();
      showToast(t("settings.toast_switched_profile", { id }) || `Switched network profile to: ${id}`, "success");
      setTimeout(() => {
        saveStatusType.value = "idle";
        saveStatusText.value = "";
      }, 2000);
    } else {
      saveStatusType.value = "error";
      saveStatusText.value =
        t("settings.status_save_failed", { error: data.error || "Unknown error" }) ||
        `Save failed: ${data.error || "Unknown error"}`;
      showToast(
        t("settings.status_save_failed", { error: data.error || "Unknown error" }) ||
          `Save failed: ${data.error || "Unknown error"}`,
        "error"
      );
    }
  } catch (err: any) {
    saveStatusType.value = "error";
    saveStatusText.value =
      t("settings.status_network_error", { error: err.message }) || `Network error: ${err.message}`;
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
  saveStatusText.value = t("settings.status_creating_profile") || "Creating profile...";
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
      saveStatusText.value = t("settings.status_created_profile") || "Profile created and activated";
      await loadNetworkStatus();
      await loadConfigData();
      showToast(t("settings.toast_created_profile", { id: pid }) || `Created and activated profile: ${pid}`, "success");
      setTimeout(() => {
        saveStatusType.value = "idle";
        saveStatusText.value = "";
      }, 2000);
    } else {
      const err = await res.json().catch(() => ({}));
      saveStatusType.value = "error";
      saveStatusText.value =
        t("settings.status_save_failed", { error: err.error || "Unknown error" }) ||
        `Save failed: ${err.error || "Unknown error"}`;
      alert(err.error || t("settings.status_save_failed", { error: "Unknown error" }) || "Failed to create profile");
    }
  } catch (err: any) {
    console.error(err);
    saveStatusType.value = "error";
    saveStatusText.value =
      t("settings.status_network_error", { error: err.message }) || `Request failed: ${err.message}`;
    alert(t("settings.status_network_error", { error: err.message }) || "Request failed: " + err.message);
  } finally {
    isSavingProfile.value = false;
    isSaving.value = false;
  }
}

async function handleResetBuiltinProfile(id: string) {
  isSaving.value = true;
  saveStatusType.value = "saving";
  saveStatusText.value = t("settings.status_resetting_profile") || "Resetting profile...";
  try {
    const res = await fetch("/api/network/profile/reset", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ profile_id: id, scope: activeScope.value }),
    });
    if (res.ok) {
      saveStatusType.value = "success";
      saveStatusText.value = t("settings.status_restored_defaults") || "Restored to factory defaults";
      await loadNetworkStatus();
      await loadConfigData();
      showToast(t("settings.status_restored_defaults") || "Restored to factory defaults", "success");
      setTimeout(() => {
        saveStatusType.value = "idle";
        saveStatusText.value = "";
      }, 2000);
    } else {
      const err = await res.json().catch(() => ({}));
      saveStatusType.value = "error";
      saveStatusText.value =
        t("settings.status_save_failed", { error: err.error || "Unknown error" }) ||
        `Save failed: ${err.error || "Unknown error"}`;
    }
  } catch (err: any) {
    saveStatusType.value = "error";
    saveStatusText.value =
      t("settings.status_network_error", { error: err.message }) || `Network error: ${err.message}`;
  } finally {
    isSaving.value = false;
  }
}

async function handleResetAllEcosystems() {
  isSaving.value = true;
  saveStatusType.value = "saving";
  saveStatusText.value = t("settings.status_restoring_ecos") || "Restoring default ecosystems...";
  try {
    const res = await fetch("/api/network/registry/reset-all", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ scope: activeScope.value }),
    });
    if (res.ok) {
      saveStatusType.value = "success";
      saveStatusText.value = t("settings.status_restored_ecos") || "Default ecosystem list restored";
      await loadNetworkStatus();
      await loadConfigData();
      showToast(t("settings.status_restored_ecos") || "Default ecosystem list restored", "success");
      setTimeout(() => {
        saveStatusType.value = "idle";
        saveStatusText.value = "";
      }, 2000);
    } else {
      const err = await res.json().catch(() => ({}));
      saveStatusType.value = "error";
      saveStatusText.value =
        t("settings.status_save_failed", { error: err.error || "Unknown error" }) ||
        `Save failed: ${err.error || "Unknown error"}`;
    }
  } catch (err: any) {
    saveStatusType.value = "error";
    saveStatusText.value =
      t("settings.status_network_error", { error: err.message }) || `Network error: ${err.message}`;
  } finally {
    isSaving.value = false;
  }
}

async function handleOverrideRegistry(ecosystem: string, url: string) {
  if (!ecosystem || !url) return;
  isSaving.value = true;
  saveStatusType.value = "saving";
  saveStatusText.value = t("settings.status_updating_mirrors") || "Updating mirror sources...";
  try {
    const res = await fetch("/api/network/registry/override", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ ecosystem, url, scope: activeScope.value }),
    });
    if (res.ok) {
      saveStatusType.value = "success";
      saveStatusText.value = t("settings.status_auto_saved") || "Auto saved";
      await loadNetworkStatus();
      await loadConfigData();
      setTimeout(() => {
        saveStatusType.value = "idle";
        saveStatusText.value = "";
      }, 2000);
    } else {
      const err = await res.json().catch(() => ({}));
      saveStatusType.value = "error";
      saveStatusText.value =
        t("settings.status_save_failed", { error: err.error || "Unknown error" }) ||
        `Save failed: ${err.error || "Unknown error"}`;
    }
  } catch (err: any) {
    saveStatusType.value = "error";
    saveStatusText.value =
      t("settings.status_network_error", { error: err.message }) || `Network error: ${err.message}`;
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
  saveStatusText.value = t("settings.status_restoring_presets") || "Restoring presets...";
  try {
    const res = await fetch("/api/network/registry/reset", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ ecosystem, scope: activeScope.value }),
    });
    if (res.ok) {
      saveStatusType.value = "success";
      saveStatusText.value = t("settings.status_restored_presets") || "Presets restored";
      await loadNetworkStatus();
      await loadConfigData();
      setTimeout(() => {
        saveStatusType.value = "idle";
        saveStatusText.value = "";
      }, 2000);
    } else {
      const err = await res.json().catch(() => ({}));
      saveStatusType.value = "error";
      saveStatusText.value =
        t("settings.status_save_failed", { error: err.error || "Unknown error" }) ||
        `Save failed: ${err.error || "Unknown error"}`;
    }
  } catch (err) {
    saveStatusType.value = "error";
    saveStatusText.value = t("settings.status_network_error", { error: "" }) || "Network request failed";
  } finally {
    isSaving.value = false;
  }
}

async function handleSaveProxy() {
  if (!networkStatus.value) return;
  isSaving.value = true;
  saveStatusType.value = "saving";
  saveStatusText.value = t("settings.status_saving") || "Saving...";
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
      saveStatusText.value = t("settings.status_auto_saved") || "Auto saved";
      await loadNetworkStatus();
      await loadConfigData();
      setTimeout(() => {
        saveStatusType.value = "idle";
        saveStatusText.value = "";
      }, 2000);
    } else {
      const err = await res.json().catch(() => ({}));
      saveStatusType.value = "error";
      saveStatusText.value =
        t("settings.status_save_failed", { error: err.error || "Unknown error" }) ||
        `Save failed: ${err.error || "Unknown error"}`;
    }
  } catch (err: any) {
    saveStatusType.value = "error";
    saveStatusText.value =
      t("settings.status_network_error", { error: err.message }) || `Network error: ${err.message}`;
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
      error: t("settings.proxy_err_no_address") || "No proxy server address provided (HTTP or SOCKS5)",
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
        error:
          t("settings.proxy_err_status", { status: res.status }) || `Proxy test request failed (HTTP ${res.status})`,
      };
    }
  } catch (err: any) {
    proxyTestResult.value = {
      success: false,
      error:
        t("settings.proxy_err_exception", { error: err.message }) ||
        `Proxy test request encountered an exception: ${err.message}`,
    };
  } finally {
    isTestingProxy.value = false;
  }
}

function getEcoPingClass(eco: string) {
  const p = ecoPingStatus.value[eco];
  if (!p || p.status === "idle") return "ping-idle";
  if (p.loading) return "ping-loading";
  return `ping-${p.status}`;
}

function getEcoPingTitle(eco: string) {
  const p = ecoPingStatus.value[eco];
  if (!p) return t("settings.ping_click_hint") || "Click to test mirror latency";
  if (p.loading) return t("settings.ping_in_progress") || "Testing speed...";
  if (p.status === "timeout") return t("settings.ping_connection_timeout") || "Connection timed out";
  if (p.status === "error") return p.error || t("settings.ping_connection_failed") || "Connection failed";
  return (
    t("settings.ping_result_format", { ms: p.rtt_ms ?? 0, status: p.status }) || `Latency: ${p.rtt_ms}ms (${p.status})`
  );
}

async function handlePingSingle(eco: string) {
  const url = networkStatus.value?.effective?.[eco]?.url;
  if (!url) return;

  ecoPingStatus.value[eco] = {
    ...ecoPingStatus.value[eco],
    loading: true,
    status: "loading",
  };

  try {
    const res = await fetch("/api/network/ping", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ ecosystem: eco, url }),
    });
    if (res.ok) {
      const data = await res.json();
      const item = data.results?.[0];
      if (item) {
        ecoPingStatus.value[eco] = {
          rtt_ms: item.rtt_ms ?? item.latency_ms ?? 0,
          status: item.status,
          loading: false,
          error: item.error,
        };
      }
    } else {
      ecoPingStatus.value[eco] = {
        status: "error",
        loading: false,
        error: t("settings.ping_failed") || "Request failed",
      };
    }
  } catch (err: any) {
    ecoPingStatus.value[eco] = {
      status: "error",
      loading: false,
      error: err?.message || t("settings.network_anomaly") || "Network exception",
    };
  }
}

async function handleRunProbeAll() {
  if (!networkStatus.value?.effective) return;
  isProbing.value = true;

  const targets = orderedEffectiveEcosystems.value
    .map((eco) => ({
      ecosystem: eco,
      name: eco,
      url: networkStatus.value?.effective?.[eco]?.url || "",
    }))
    .filter((t) => t.url);

  targets.forEach((t) => {
    ecoPingStatus.value[t.ecosystem] = {
      ...ecoPingStatus.value[t.ecosystem],
      loading: true,
      status: "loading",
    };
  });

  try {
    const res = await fetch("/api/network/ping", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ targets }),
    });
    if (res.ok) {
      const data = await res.json();
      const results: any[] = data.results || [];
      results.forEach((r) => {
        if (r.ecosystem) {
          ecoPingStatus.value[r.ecosystem] = {
            rtt_ms: r.rtt_ms ?? r.latency_ms ?? 0,
            status: r.status,
            loading: false,
            error: r.error,
          };
        }
      });
    }
  } catch (err) {
    console.error(err);
  } finally {
    isProbing.value = false;
    Object.keys(ecoPingStatus.value).forEach((k) => {
      if (ecoPingStatus.value[k]?.loading) {
        ecoPingStatus.value[k].loading = false;
      }
    });
  }
}

onMounted(() => {
  loadConfigData();
  loadSystemStatus();
  loadNetworkStatus();
  fetchHooksList();
  fetchAliasesList();
  handleDetectProxy(false);
  fetchThemes();
  fetchTerminals();
  if (props.activeCategory) {
    setTimeout(() => {
      scrollToCategory(props.activeCategory!);
    }, 150);
  }
  if (
    route.name === "settings" &&
    (route.query.open_toml === "true" || route.query.file === ".unirtm.toml" || route.query.file === "unirtm.toml")
  ) {
    handleOpenTomlEditor();
  }
});

onActivated(() => {
  if (
    route.name === "settings" &&
    (route.query.open_toml === "true" || route.query.file === ".unirtm.toml" || route.query.file === "unirtm.toml")
  ) {
    if (!showRawTomlModal.value) {
      handleOpenTomlEditor();
    }
  }
});

onDeactivated(() => {
  if (showRawTomlModal.value) {
    handleCloseTomlEditor();
  }
});

onBeforeUnmount(() => {
  if (tomlMonacoModel) {
    tomlMonacoModel.dispose();
    tomlMonacoModel = null;
  }
  if (tomlMonacoEditor) {
    tomlMonacoEditor.dispose();
    tomlMonacoEditor = null;
  }
});

watch(
  () => [route.name, route.query.open_toml, route.query.file],
  ([name, openToml, file]) => {
    if (name !== "settings") {
      return;
    }
    if (openToml === "true" || file === ".unirtm.toml" || file === "unirtm.toml") {
      if (!showRawTomlModal.value) {
        handleOpenTomlEditor();
      }
    }
  }
);
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
  transition:
    border-color 0.2s ease,
    box-shadow 0.2s ease;
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
  transition:
    border-color 0.15s ease,
    color 0.15s ease;
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

/* Settings TOML Monaco Editor Modal */
.config-editor-modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.65);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  padding: 24px;
}

.config-editor-modal {
  width: 920px;
  max-width: 94vw;
  height: 660px;
  max-height: 90vh;
  background: var(--vscode-editor-background, #1e1e1e);
  border: 1px solid var(--vscode-panel-border, rgba(255, 255, 255, 0.12));
  border-radius: 8px;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.5);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

:root[data-theme="light"] .config-editor-modal {
  background: #ffffff;
  border-color: #e2e8f0;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.15);
}

.config-editor-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 18px;
  background: var(--vscode-editorGroupHeader-tabsBackground, #252526);
  border-bottom: 1px solid var(--vscode-panel-border, rgba(255, 255, 255, 0.08));
}

:root[data-theme="light"] .config-editor-header {
  background: #f8fafc;
  border-bottom-color: #e2e8f0;
}

.config-editor-header .header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.config-editor-header .file-icon {
  font-size: 20px;
  color: #38bdf8;
}

.config-editor-header .file-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-main, #fff);
  font-family: monospace;
}

.config-editor-header .file-desc {
  font-size: 11px;
  color: var(--text-muted, #94a3b8);
}

.scope-tag {
  font-size: 11px;
  color: var(--accent-cyan, #06b6d4);
  background: rgba(6, 182, 212, 0.1);
  padding: 2px 6px;
  border-radius: 4px;
}

.config-badge {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 10px;
  font-weight: 500;
}

.config-badge.badge-active {
  background: rgba(16, 185, 129, 0.15);
  color: #10b981;
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.config-badge.badge-draft {
  background: rgba(245, 158, 11, 0.15);
  color: #f59e0b;
  border: 1px solid rgba(245, 158, 11, 0.3);
}

.config-editor-header .header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.modal-action-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 5px 12px;
  font-size: 12px;
  font-weight: 500;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.modal-action-btn.secondary {
  background: #2563eb;
  color: #ffffff;
  border: 1px solid #1d4ed8;
}

.modal-action-btn.secondary:hover {
  background: #1d4ed8;
  border-color: #1e40af;
  color: #ffffff;
}

:root[data-theme="light"] .modal-action-btn.secondary {
  background: #0284c7;
  color: #ffffff;
  border: 1px solid #0369a1;
}

:root[data-theme="light"] .modal-action-btn.secondary:hover {
  background: #0369a1;
  color: #ffffff;
}

.modal-action-btn.copy {
  background: rgba(255, 255, 255, 0.1);
  color: #e2e8f0;
  border: 1px solid rgba(255, 255, 255, 0.2);
}

.modal-action-btn.copy:hover {
  background: rgba(255, 255, 255, 0.18);
  color: #ffffff;
}

:root[data-theme="light"] .modal-action-btn.copy {
  background: #f1f5f9;
  color: #334155;
  border: 1px solid #cbd5e1;
}

:root[data-theme="light"] .modal-action-btn.copy:hover {
  background: #e2e8f0;
  color: #0f172a;
}

.modal-action-btn.close {
  background: rgba(255, 255, 255, 0.1);
  color: #e2e8f0;
  border: 1px solid rgba(255, 255, 255, 0.2);
  font-size: 14px;
  padding: 5px 8px;
  border-radius: 4px;
}

.modal-action-btn.close:hover {
  color: #ffffff;
  background: #ef4444;
  border-color: #dc2626;
}

:root[data-theme="light"] .modal-action-btn.close {
  background: #f1f5f9;
  color: #334155;
  border: 1px solid #cbd5e1;
}

:root[data-theme="light"] .modal-action-btn.close:hover {
  color: #ffffff;
  background: #ef4444;
  border-color: #dc2626;
}

.config-editor-body {
  flex: 1;
  min-height: 0;
  position: relative;
  background: var(--vscode-editor-background, #1e1e1e);
}

:root[data-theme="light"] .config-editor-body {
  background: #ffffff;
}

.config-monaco-host {
  width: 100%;
  height: 100%;
}

.editor-loading {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  color: var(--text-muted, #94a3b8);
  font-size: 13px;
}

.config-editor-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 18px;
  background: var(--vscode-editorGroupHeader-tabsBackground, #252526);
  border-top: 1px solid var(--vscode-panel-border, rgba(255, 255, 255, 0.08));
}

:root[data-theme="light"] .config-editor-footer {
  background: #f8fafc;
  border-top-color: #e2e8f0;
}

.footer-status .status-dirty {
  color: #f59e0b;
  font-size: 12px;
}

.footer-status .status-clean {
  color: #10b981;
  font-size: 12px;
}

.footer-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.modal-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  font-size: 12px;
  font-weight: 500;
  border-radius: 4px;
  cursor: pointer;
  border: 1px solid transparent;
  transition: all 0.15s ease;
}

.modal-btn.secondary {
  background: transparent;
  border-color: var(--vscode-panel-border, rgba(255, 255, 255, 0.15));
  color: var(--text-main, #e2e8f0);
}

.modal-btn.secondary:hover {
  background: rgba(255, 255, 255, 0.08);
}

:root[data-theme="light"] .modal-btn.secondary {
  border-color: #cbd5e1;
  color: #475569;
}

:root[data-theme="light"] .modal-btn.secondary:hover {
  background: #f1f5f9;
}

.modal-btn.primary {
  background: #0284c7;
  color: #ffffff;
}

.modal-btn.primary:hover:not(:disabled) {
  background: #0369a1;
}

.modal-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
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

.pill-builtin,
.pill-custom {
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

.btn-template-chip:hover,
.btn-template-chip.active {
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
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
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
  to {
    transform: rotate(360deg);
  }
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

/* --- Tool Version Aliases Styles --- */
.aliases-container {
  margin-top: 12px;
}

.aliases-loading {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 24px;
  justify-content: center;
  color: var(--text-muted);
}

.btn-add-alias {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  border-radius: 6px;
  font-size: 0.8rem;
  cursor: pointer;
  background: var(--accent-cyan);
  border: 1px solid transparent;
  color: #0b0f19;
  font-weight: 600;
  transition: all 0.2s ease;
}

.btn-add-alias:hover {
  filter: brightness(1.1);
  box-shadow: 0 0 12px rgba(6, 182, 212, 0.35);
}

.aliases-groups-list {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.alias-tool-card {
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  border-radius: 8px;
  padding: 14px 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  transition: all 0.2s ease;
}

.alias-tool-card:hover {
  border-color: var(--border-color);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.alias-tool-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  padding-bottom: 8px;
}

.tool-info-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.tool-icon-wrapper {
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.tool-name {
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--text-main);
}

.alias-count-badge {
  font-size: 0.7rem;
  padding: 1px 6px;
  border-radius: 10px;
  background: rgba(6, 182, 212, 0.12);
  color: var(--accent-cyan);
  border: 1px solid rgba(6, 182, 212, 0.25);
}

.btn-tool-add-alias {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 3px 8px;
  font-size: 0.74rem;
  border-radius: 4px;
  background: transparent;
  border: 1px solid var(--card-border);
  color: var(--text-muted);
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-tool-add-alias:hover {
  background: rgba(255, 255, 255, 0.08);
  color: var(--text-main);
  border-color: var(--border-color);
}

.alias-chips-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 10px;
}

.alias-chip-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  border-radius: 6px;
  background: rgba(0, 0, 0, 0.2);
  border: 1px solid var(--card-border);
  transition: all 0.15s ease;
}

.alias-chip-item:hover {
  background: rgba(0, 0, 0, 0.28);
  border-color: rgba(6, 182, 212, 0.3);
}

.alias-chip-main {
  display: flex;
  align-items: center;
  gap: 6px;
  overflow: hidden;
}

.alias-tag-name {
  font-size: 0.82rem;
  font-weight: 600;
  color: #38bdf8;
  background: rgba(56, 189, 248, 0.1);
  padding: 2px 6px;
  border-radius: 4px;
}

.alias-arrow {
  color: var(--text-muted);
  font-size: 0.8rem;
}

.alias-target-version {
  font-size: 0.82rem;
  color: var(--text-main);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.alias-chip-actions {
  display: flex;
  align-items: center;
  gap: 4px;
}

.btn-chip-action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 3px 5px;
  border-radius: 4px;
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-chip-action:hover {
  color: var(--text-main);
  background: rgba(255, 255, 255, 0.08);
}

.btn-chip-action.action-delete:hover {
  color: #ef4444;
  background: rgba(239, 68, 68, 0.15);
}

.aliases-empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: 32px 16px;
  background: rgba(255, 255, 255, 0.02);
  border: 1px dashed var(--card-border);
  border-radius: 8px;
}

.aliases-empty-state h4 {
  margin: 0 0 6px 0;
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--text-main);
}

.aliases-empty-state p {
  margin: 0 0 16px 0;
  font-size: 0.82rem;
  color: var(--text-muted);
  max-width: 480px;
  line-height: 1.5;
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
[data-theme="light"] .btn-add-hook,
:root[data-theme="light"] .btn-add-alias,
[data-theme="light"] .btn-add-alias {
  background: #0284c7 !important;
  color: #ffffff !important;
  border-color: #0284c7 !important;
}

:root[data-theme="light"] .btn-add-hook:hover,
[data-theme="light"] .btn-add-hook:hover,
:root[data-theme="light"] .btn-add-alias:hover,
[data-theme="light"] .btn-add-alias:hover {
  background: #0369a1 !important;
  color: #ffffff !important;
  box-shadow: 0 0 10px rgba(2, 132, 199, 0.35);
}

:root[data-theme="light"] .hooks-empty-state,
[data-theme="light"] .hooks-empty-state,
:root[data-theme="light"] .aliases-empty-state,
[data-theme="light"] .aliases-empty-state {
  background: #f8fafc;
  border-color: #e2e8f0;
}

:root[data-theme="light"] .alias-tool-card,
[data-theme="light"] .alias-tool-card {
  background: #ffffff;
  border-color: #e2e8f0;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

:root[data-theme="light"] .alias-tool-card:hover,
[data-theme="light"] .alias-tool-card:hover {
  border-color: #cbd5e1;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}

:root[data-theme="light"] .alias-tool-header,
[data-theme="light"] .alias-tool-header {
  border-bottom-color: #f1f5f9;
}

:root[data-theme="light"] .alias-chip-item,
[data-theme="light"] .alias-chip-item {
  background: #f8fafc;
  border-color: #e2e8f0;
}

:root[data-theme="light"] .alias-chip-item:hover,
[data-theme="light"] .alias-chip-item:hover {
  background: #f1f5f9;
  border-color: #0284c7;
}

:root[data-theme="light"] .alias-tag-name,
[data-theme="light"] .alias-tag-name {
  color: #0284c7;
  background: #e0f2fe;
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
