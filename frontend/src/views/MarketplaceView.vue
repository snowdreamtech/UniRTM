<template>
  <div class="view-container marketplace-view" id="view-marketplace">
    <!-- VIEW MODE 1: EXTENSION DETAIL PAGE -->
    <div v-if="route.params.namespace && route.params.name" class="market-detail-container">
      <!-- Breadcrumb Bar -->
      <div class="detail-nav-bar glass-card">
        <button class="btn-back-market" @click="goBackToList">
          <span class="back-arrow">←</span>
          <span>返回市场列表</span>
        </button>
        <div class="breadcrumb-trail">
          <span class="crumb-link" @click="goBackToList">扩展与主题市场</span>
          <span class="crumb-sep">/</span>
          <span class="crumb-ns font-mono">{{ route.params.namespace }}</span>
          <span class="crumb-sep">/</span>
          <span class="crumb-curr">{{ activeDetail?.displayName || activeDetail?.name || route.params.name }}</span>
        </div>
      </div>

      <!-- Feedback Alert -->
      <div v-if="feedbackMsg" class="feedback-banner" :class="{ error: feedbackIsError }">
        <span>{{ feedbackMsg }}</span>
        <button class="btn-close-banner" @click="feedbackMsg = ''">✕</button>
      </div>

      <!-- Loading State -->
      <div v-if="isLoadingDetail" class="detail-loading-box glass-card">
        <div class="spinner-small"></div>
        <span>正在实时获取扩展详细信息、README 与主题清单...</span>
      </div>

      <!-- Error State -->
      <div v-else-if="!activeDetail" class="detail-error-box glass-card">
        <span class="error-icon">
          <TechIcon id="workbench:warning" :size="32" />
        </span>
        <h3>未找到该扩展主题信息</h3>
        <p>{{ mpError || "网络请求失败或该扩展不存在于 Open VSX 仓库。" }}</p>
        <div class="detail-error-actions">
          <button class="btn-action-primary" @click="loadCurrentDetail">重新尝试</button>
          <button class="btn-action-secondary" @click="goBackToList">返回市场</button>
        </div>
      </div>

      <!-- Detail Main Content -->
      <div v-else class="detail-main-layout">
        <!-- Hero Header -->
        <div class="detail-hero glass-card">
          <div class="hero-left">
            <div class="hero-icon-box">
              <img
                v-if="hasValidIcon(activeDetail)"
                :src="activeDetail.files!.icon"
                class="hero-img-icon"
                alt="icon"
                @error="handleIconError(`${activeDetail.namespace || ''}.${activeDetail.name}`)"
              />
              <span v-else class="hero-fallback-icon">
                <TechIcon id="workbench:marketplace" :size="32" />
              </span>
            </div>
            <div class="hero-info">
              <div class="hero-title-row">
                <h1 class="hero-title">{{ activeDetail.displayName || activeDetail.name }}</h1>
                <span class="hero-version font-mono">v{{ activeDetail.version }}</span>
                <span v-if="isDetailActive" class="badge-active-pill">● 当前生效使用中</span>
                <span v-else-if="activeDetail.is_installed" class="badge-installed">✓ {{ t("settings.installed") || "已安装" }}</span>
              </div>
              <div class="hero-meta-row">
                <span class="hero-author font-mono">
                  发布者: <strong>{{ activeDetail.namespace }}</strong>
                </span>
                <span class="meta-dot">•</span>
                <span class="hero-stat">⬇️ {{ formatDownloads(activeDetail.downloadCount) }} 次下载</span>
                <span v-if="activeDetail.averageRating" class="meta-dot">•</span>
                <span v-if="activeDetail.averageRating" class="hero-stat">
                  ⭐ {{ activeDetail.averageRating.toFixed(1) }} ({{ activeDetail.reviewCount || 1 }} 评价)
                </span>
                <span v-if="activeDetail.license" class="meta-dot">•</span>
                <span v-if="activeDetail.license" class="hero-license font-mono">{{ activeDetail.license }}</span>
              </div>
              <p class="hero-desc">{{ activeDetail.description || "无详细描述" }}</p>
            </div>
          </div>

          <div class="hero-actions">
            <button
              v-if="!activeDetail.is_installed"
              class="btn-install-primary"
              :disabled="mpIsInstalling(activeDetail)"
              @click="onInstallDetail(activeDetail)"
            >
              <span v-if="mpIsInstalling(activeDetail)" class="spinner-small"></span>
              <span>{{ mpIsInstalling(activeDetail) ? (t("settings.installing") || "安装中...") : (t("settings.install") || "安装") }}</span>
            </button>
            <div v-else class="installed-action-buttons">
              <span v-if="isDetailActive" class="badge-hero-applied">
                <span class="active-dot"></span>
                <span>当前主题已生效</span>
              </span>
              <button
                v-else
                class="btn-apply-detail-primary"
                @click="onApplyDetailTheme(activeDetail)"
              >
                应用此主题
              </button>
              <button
                class="btn-uninstall-danger"
                :disabled="isUninstalling"
                @click="onUninstallDetail(activeDetail)"
              >
                {{ isUninstalling ? "卸载中..." : "卸载" }}
              </button>
            </div>
            <a
              :href="'https://open-vsx.org/extension/' + activeDetail.namespace + '/' + activeDetail.name"
              target="_blank"
              rel="noopener noreferrer"
              class="btn-external-link"
              title="在 Open VSX 官方页面打开"
            >
              <span>在 Open VSX 查看</span>
              <span>↗</span>
            </a>
          </div>
        </div>

        <!-- Detail Body Grid: Main Content + Sidebar -->
        <div class="detail-body-grid">
          <!-- Main Tabs Area -->
          <div class="detail-body-main glass-card">
            <div class="detail-subtabs">
              <button
                class="subtab-btn"
                :class="{ active: activeDetailTab === 'readme' }"
                @click="activeDetailTab = 'readme'"
              >
                📖 详情说明 (README)
              </button>
              <button
                class="subtab-btn"
                :class="{ active: activeDetailTab === 'contributes' }"
                @click="activeDetailTab = 'contributes'"
              >
                🎨 包含主题与图标
                <span v-if="contributedThemesCount > 0" class="tab-badge">{{ contributedThemesCount }}</span>
              </button>
              <button
                v-if="activeDetail.files && activeDetail.files.changelog"
                class="subtab-btn"
                :class="{ active: activeDetailTab === 'changelog' }"
                @click="activeDetailTab = 'changelog'"
              >
                📝 更新日志
              </button>
            </div>

            <!-- Tab 1: README (Markdown with Screenshots) -->
            <div v-show="activeDetailTab === 'readme'" class="tab-content tab-readme">
              <div
                v-if="activeDetail.readme"
                class="markdown-body"
                v-html="renderedReadme"
              ></div>
              <div v-else class="readme-empty-hint">
                <span class="empty-icon">
                  <TechIcon id="workbench:about" :size="32" />
                </span>
                <p>该扩展未提供 README 文档说明</p>
              </div>
            </div>

            <!-- Tab 2: Contributed Sub-themes -->
            <div v-show="activeDetailTab === 'contributes'" class="tab-content tab-contributes">
              <!-- Color Themes List -->
              <div
                v-if="activeDetail.contributes?.themes && activeDetail.contributes.themes.length > 0"
                class="contribute-group"
              >
                <h3 class="group-title">颜色主题 (Color Themes)</h3>
                <div class="contribute-list">
                  <div
                    v-for="sub in activeDetail.contributes.themes"
                    :key="sub.label"
                    class="contribute-item"
                  >
                    <div class="sub-info">
                      <span class="sub-label">{{ sub.label }}</span>
                      <span
                        class="sub-mode-tag"
                        :class="sub.uiTheme === 'vs' ? 'mode-light' : 'mode-dark'"
                      >
                        {{ formatUITheme(sub.uiTheme) }}
                      </span>
                    </div>
                    <button
                      v-if="activeDetail.is_installed && isSubColorActive(sub.label, sub.id)"
                      class="btn-applied-current"
                      disabled
                      :title="t('marketplace.current_active') || '当前正在生效的主题'"
                    >
                      {{ t('marketplace.currently_applied') || '已应用' }}
                    </button>
                    <button
                      v-else-if="activeDetail.is_installed"
                      class="btn-apply-sub"
                      @click="onApplySubColorTheme(sub.label)"
                    >
                      {{ t('settings.apply') || '应用' }}
                    </button>
                  </div>
                </div>
              </div>

              <!-- Icon Themes List -->
              <div
                v-if="activeDetail.contributes?.iconThemes && activeDetail.contributes.iconThemes.length > 0"
                class="contribute-group"
              >
                <h3 class="group-title">文件图标 (Icon Themes)</h3>
                <div class="contribute-list">
                  <div
                    v-for="sub in activeDetail.contributes.iconThemes"
                    :key="sub.label"
                    class="contribute-item"
                  >
                    <div class="sub-info">
                      <span class="sub-label">{{ sub.label }}</span>
                      <span class="sub-mode-tag mode-icon">{{ t('marketplace.filter_icons') || '文件图标' }}</span>
                    </div>
                    <button
                      v-if="activeDetail.is_installed && isSubIconActive(sub.label, sub.id)"
                      class="btn-applied-current"
                      disabled
                      :title="t('marketplace.current_active') || '当前正在生效的主题'"
                    >
                      {{ t('marketplace.currently_applied') || '已应用' }}
                    </button>
                    <button
                      v-else-if="activeDetail.is_installed"
                      class="btn-apply-sub"
                      @click="onApplySubIconTheme(sub.label)"
                    >
                      {{ t('settings.apply') || '应用' }}
                    </button>
                  </div>
                </div>
              </div>

              <!-- Product Icon Themes List -->
              <div
                v-if="activeDetail.contributes?.productIconThemes && activeDetail.contributes.productIconThemes.length > 0"
                class="contribute-group"
              >
                <h3 class="group-title">界面图标 (Product Icon Themes)</h3>
                <div class="contribute-list">
                  <div
                    v-for="sub in activeDetail.contributes.productIconThemes"
                    :key="sub.label"
                    class="contribute-item"
                  >
                    <div class="sub-info">
                      <span class="sub-label">{{ sub.label }}</span>
                      <span class="sub-mode-tag mode-product">{{ t('marketplace.filter_product_icons') || '界面图标' }}</span>
                    </div>
                    <button
                      v-if="activeDetail.is_installed && isSubProductIconActive(sub.label, sub.id)"
                      class="btn-applied-current"
                      disabled
                      :title="t('marketplace.current_active') || '当前正在生效的主题'"
                    >
                      {{ t('marketplace.currently_applied') || '已应用' }}
                    </button>
                    <button
                      v-else-if="activeDetail.is_installed"
                      class="btn-apply-sub"
                      @click="onApplySubProductIconTheme(sub.id || sub.label)"
                    >
                      {{ t('settings.apply') || '应用' }}
                    </button>
                  </div>
                </div>
              </div>

              <div
                v-if="(!activeDetail.contributes?.themes || activeDetail.contributes.themes.length === 0) &&
                      (!activeDetail.contributes?.iconThemes || activeDetail.contributes.iconThemes.length === 0) &&
                      (!activeDetail.contributes?.productIconThemes || activeDetail.contributes.productIconThemes.length === 0)"
                class="readme-empty-hint"
              >
                <p>该扩展清单中未发现显式声明的主题子集</p>
              </div>
            </div>

            <!-- Tab 3: Changelog -->
            <div v-show="activeDetailTab === 'changelog'" class="tab-content tab-changelog">
              <div
                v-if="changelogContent"
                class="markdown-body"
                v-html="renderedChangelog"
              ></div>
              <div v-else class="readme-empty-hint">
                <p>正在读取或该扩展未提供独立更新日志</p>
              </div>
            </div>
          </div>

          <!-- Right Sidebar Metadata -->
          <div class="detail-sidebar glass-card">
            <!-- Resources -->
            <div class="sidebar-section">
              <h4 class="sidebar-heading">资源与链接</h4>
              <div class="sidebar-links">
                <a
                  v-if="activeDetail.repository"
                  :href="cleanGitUrl(activeDetail.repository)"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="sidebar-link"
                >
                  <span class="link-icon">
                    <TechIcon id="workbench:github" :size="14" />
                  </span>
                  <span>项目仓库 (Repository)</span>
                </a>
                <a
                  v-if="activeDetail.bugs"
                  :href="activeDetail.bugs"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="sidebar-link"
                >
                  <span class="link-icon">
                    <TechIcon id="workbench:doctor" :size="14" />
                  </span>
                  <span>问题反馈 (Issues)</span>
                </a>
                <a
                  v-if="activeDetail.homepage"
                  :href="activeDetail.homepage"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="sidebar-link"
                >
                  <span class="link-icon">
                    <TechIcon id="workbench:dashboard" :size="14" />
                  </span>
                  <span>主页 (Homepage)</span>
                </a>
                <a
                  v-if="activeDetail.files && activeDetail.files.download"
                  :href="activeDetail.files.download"
                  download
                  class="sidebar-link"
                >
                  <span class="link-icon">
                    <TechIcon id="workbench:package" :size="14" />
                  </span>
                  <span>下载 .vsix 离线安装包</span>
                </a>
              </div>
            </div>

            <!-- Properties -->
            <div class="sidebar-section">
              <h4 class="sidebar-heading">元数据属性</h4>
              <div class="prop-list">
                <div class="prop-item">
                  <span class="prop-key">命名空间</span>
                  <span class="prop-val font-mono">{{ activeDetail.namespace }}</span>
                </div>
                <div class="prop-item">
                  <span class="prop-key">扩展标识</span>
                  <span class="prop-val font-mono">{{ activeDetail.namespace }}.{{ activeDetail.name }}</span>
                </div>
                <div class="prop-item">
                  <span class="prop-key">最新版本</span>
                  <span class="prop-val font-mono">v{{ activeDetail.version }}</span>
                </div>
                <div class="prop-item">
                  <span class="prop-key">发布日期</span>
                  <span class="prop-val">{{ formatDate(activeDetail.timestamp) }}</span>
                </div>
                <div class="prop-item">
                  <span class="prop-key">许可证</span>
                  <span class="prop-val">{{ activeDetail.license || "开源协议" }}</span>
                </div>
              </div>
            </div>

            <!-- Tags -->
            <div v-if="activeDetail.tags && activeDetail.tags.length > 0" class="sidebar-section">
              <h4 class="sidebar-heading">标签分类</h4>
              <div class="sidebar-tags">
                <span
                  v-for="tag in activeDetail.tags"
                  :key="tag"
                  class="tag-chip"
                >
                  {{ tag }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- VIEW MODE 2: MARKETPLACE DISCOVERY LIST (DUAL CHANNELS) -->
    <template v-else>
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
            <TechIcon id="workbench:marketplace" :size="16" />
            <span>{{ t("settings.theme_tab_marketplace") || "主题与图标 (Open VSX)" }}</span>
          </button>
          <button
            class="channel-btn"
            :class="{ active: currentChannel === 'plugins' }"
            @click="switchChannel('plugins')"
          >
            <TechIcon id="workbench:tools" :size="16" />
            <span>{{ t("settings.cat_plugins") || "功能插件 (UniRTM 生态)" }}</span>
          </button>
        </div>
      </div>

      <!-- CHANNEL 1: Themes & Icons from Open VSX -->
      <div v-show="currentChannel === 'themes'" class="market-channel-content">
        <!-- Structural Type Dimension Switcher Tabs -->
        <div class="market-type-tabs glass-card">
          <button
            class="type-tab-btn"
            :class="{ active: activeThemeTypeTab === 'all' }"
            @click="switchThemeTypeTab('all')"
          >
            <TechIcon id="workbench:marketplace" :size="15" />
            <span>{{ t('marketplace.filter_all') || '全部' }}</span>
          </button>
          <button
            class="type-tab-btn"
            :class="{ active: activeThemeTypeTab === 'color' }"
            @click="switchThemeTypeTab('color')"
          >
            <span>🎨</span>
            <span>{{ t('marketplace.filter_themes') || '颜色主题' }}</span>
          </button>
          <button
            class="type-tab-btn"
            :class="{ active: activeThemeTypeTab === 'icon' }"
            @click="switchThemeTypeTab('icon')"
          >
            <span>📁</span>
            <span>{{ t('marketplace.filter_icons') || '文件图标' }}</span>
          </button>
          <button
            class="type-tab-btn"
            :class="{ active: activeThemeTypeTab === 'product_icon' }"
            @click="switchThemeTypeTab('product_icon')"
          >
            <span>⚡</span>
            <span>{{ t('marketplace.filter_product_icons') || '界面图标' }}</span>
          </button>
          <button
            class="type-tab-btn type-tab-installed"
            :class="{ active: activeThemeTypeTab === 'installed' }"
            @click="switchThemeTypeTab('installed')"
          >
            <span>📦</span>
            <span>{{ t('marketplace.filter_installed') || '已安装' }}</span>
            <span v-if="installedExtensions.length > 0" class="tab-badge font-mono">{{ installedExtensions.length }}</span>
          </button>
        </div>

        <!-- Search Filter Bar -->
        <div class="market-filter-bar glass-card">
          <div class="search-input-wrapper">
            <span class="search-icon">
              <TechIcon id="workbench:search" :size="14" />
            </span>
            <input
              v-model="mpSearchQuery"
              type="text"
              class="market-search-input"
              :placeholder="searchPlaceholder"
              @keydown.enter="onSearchThemes(mpSearchQuery)"
            />
            <button
              v-if="mpSearchQuery"
              class="btn-clear-search"
              title="清空"
              @click="onClearSearch"
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
            <span v-else>{{ t('marketplace.search') || '搜索' }}</span>
          </button>

          <!-- View Mode Switcher -->
          <div class="view-mode-toggle" id="marketplace-view-mode-toggle">
            <button
              type="button"
              class="view-mode-btn"
              :class="{ active: viewMode === 'cards' }"
              @click="setViewMode('cards')"
              :title="t('environments.view_cards') || '卡片视图'"
            >
              <TechIcon id="workbench:cards" :size="14" />
            </button>
            <button
              type="button"
              class="view-mode-btn"
              :class="{ active: viewMode === 'list' }"
              @click="setViewMode('list')"
              :title="t('environments.view_list') || '列表视图'"
            >
              <TechIcon id="workbench:list" :size="14" />
            </button>
          </div>
        </div>

        <!-- Quick Chips & Results count -->
        <div class="market-chips-bar">
          <div class="chips-left">
            <span class="chips-label">{{ activeThemeTypeTab === 'installed' ? '类型过滤:' : (t('marketplace.recommended_categories') || '推荐分类') + ':' }}</span>
            <button
              v-for="chip in contextualChips"
              :key="chip.id"
              class="chip-btn"
              :class="{
                active: selectedFilter === chip.id,
                'chip-installed': chip.isInstalled
              }"
              @click="onSelectContextualChip(chip)"
            >
              <span>{{ (chip.labelKey ? t(chip.labelKey as any) : chip.name) || chip.id }}</span>
            </button>
          </div>
          <span v-if="activeThemeTypeTab === 'installed'" class="results-count-label">
            {{ t('marketplace.installed_count', { count: displayedExtensions.length }) || `已安装 ${displayedExtensions.length} 款主题与图标包` }}
          </span>
          <span v-else class="results-count-label">
            {{ activeThemeTypeTab !== 'all' ? `筛选后呈现 ${displayedExtensions.length} 款` : (mpTotalSize > 0 ? (t('marketplace.total_count', { count: mpTotalSize }) || `共收录 ${mpTotalSize} 款主题与图标包`) : '') }}
          </span>
        </div>

        <!-- Feedback Alert -->
        <div v-if="feedbackMsg" class="feedback-banner" :class="{ error: feedbackIsError }">
          <span>{{ feedbackMsg }}</span>
          <button class="btn-close-banner" @click="feedbackMsg = ''">✕</button>
        </div>

        <!-- Error Banner -->
        <div v-if="mpError && activeThemeTypeTab !== 'installed'" class="error-banner">
          <span style="display: inline-flex; align-items: center; gap: 6px;">
            <TechIcon id="workbench:warning" :size="14" />
            <span>{{ mpError }}</span>
          </span>
          <button class="btn-retry" @click="onSearchThemes(mpSearchQuery)">{{ t('marketplace.retry') || '重试' }}</button>
        </div>

        <!-- Loading State -->
        <div v-if="mpIsSearching && activeThemeTypeTab !== 'installed'" class="market-loading-box glass-card">
          <div class="spinner-small"></div>
          <span>{{ t('marketplace.searching_openvsx') || '正在实时连接 Open VSX 检索主题与图标...' }}</span>
        </div>

        <!-- Empty State -->
        <div v-else-if="displayedExtensions.length === 0" class="market-empty-box glass-card">
          <span class="empty-icon">
            <TechIcon id="workbench:marketplace" :size="36" />
          </span>
          <p>{{ emptyStateMessage }}</p>
          <button class="btn-reset-chip" @click="onResetToPopular">{{ t('marketplace.view_popular') || '查看热门推荐' }}</button>
        </div>

        <!-- Themes Cards View -->
        <div v-else-if="viewMode === 'cards'" class="extensions-grid">
          <div
            v-for="item in displayedExtensions"
            :key="(item.namespace || '') + '.' + item.name"
            class="glass-card theme-card clickable-card"
            :class="{ 'card-is-active': isExtensionActive(item) }"
            @click="navigateToDetail(item.namespace || '', item.name)"
          >

            <div class="card-header-row">
              <div class="card-icon-wrap">
                <img
                  v-if="hasValidIcon(item)"
                  :src="item.files!.icon"
                  class="card-img-icon"
                  loading="lazy"
                  alt="icon"
                  @error="handleIconError(`${item.namespace || ''}.${item.name}`)"
                />
                <span v-else class="card-fallback-icon">
                  <TechIcon id="workbench:marketplace" :size="24" />
                </span>
              </div>
              <div class="card-title-col">
                <div class="card-name-row">
                  <span class="card-name" :title="item.displayName || item.name">
                    {{ item.displayName || item.name }}
                  </span>
                  <span class="card-version font-mono">v{{ item.version }}</span>
                  <span v-if="isExtensionActive(item)" class="badge-active-pill">
                    <span class="active-dot"></span>
                    <span>{{ t('marketplace.active') || '生效中' }}</span>
                  </span>
                </div>
                <div class="card-subtitle-row">
                  <span class="card-namespace font-mono">{{ item.namespace }}</span>
                  <span
                    class="card-type-tag"
                    :class="{
                      'type-product-icon': getCardDisplayType(item) === 'product_icon',
                      'type-icon': getCardDisplayType(item) === 'icon',
                      'type-color': getCardDisplayType(item) === 'theme'
                    }"
                    style="display: inline-flex; align-items: center; gap: 4px;"
                  >
                    <TechIcon :id="getCardDisplayType(item) === 'theme' ? 'workbench:marketplace' : 'workbench:tools'" :size="12" :colored="false" />
                    {{
                      getCardDisplayType(item) === 'product_icon'
                        ? (t('marketplace.filter_product_icons') || '界面图标')
                        : (getCardDisplayType(item) === 'icon'
                            ? (t('marketplace.filter_icons') || '文件图标')
                            : (t('marketplace.filter_themes') || '颜色主题'))
                    }}
                  </span>
                </div>
              </div>
            </div>

            <p class="card-desc" :title="item.description">
              {{ item.description || "No description provided." }}
            </p>

            <div class="card-footer-row">
              <div class="card-stats">
                <span class="stat-item" :title="'Downloads: ' + item.downloadCount">
                  ⬇️ {{ formatDownloads(item.downloadCount) }}
                </span>
                <span v-if="item.averageRating" class="stat-item">
                  ⭐ {{ item.averageRating.toFixed(1) }}
                </span>
              </div>

              <div class="card-actions" @click.stop>
                <div v-if="isItemInstalled(item)" class="installed-group">
                  <span :class="isExtensionActive(item) ? 'badge-active-applied' : 'badge-installed'">
                    ✓ {{ isExtensionActive(item) ? (t('marketplace.in_use') || "使用中") : (t("settings.installed") || "已安装") }}
                  </span>
                  <button
                    v-if="!isExtensionActive(item)"
                    class="btn-apply-theme"
                    :title="t('settings.apply_theme') || '应用此主题'"
                    @click="onApplyTheme(item)"
                  >
                    {{ t("settings.apply") || "应用" }}
                  </button>
                  <button
                    v-else
                    class="btn-applied-current"
                    disabled
                    :title="t('marketplace.current_active') || '当前正在生效的主题'"
                  >
                    {{ t('marketplace.currently_applied') || '已应用' }}
                  </button>
                  <!-- More Menu for Installed Themes (Uninstall) -->
                  <div
                    v-if="activeThemeTypeTab === 'installed'"
                    class="theme-more-menu-wrap"
                    @click.stop
                  >
                    <button
                      class="btn-theme-more"
                      :class="{ 'is-active': activeMenuThemePkgId === (item.package_id || item.name) }"
                      @click.stop="toggleThemeMenu(item.package_id || item.name)"
                      :title="t('common.more_actions') || '更多操作'"
                    >
                      <TechIcon id="workbench:more" :size="14" />
                    </button>
                    <transition name="dropdown-fade">
                      <div
                        v-if="activeMenuThemePkgId === (item.package_id || item.name)"
                        class="theme-dropdown-menu"
                        @click.stop
                      >
                        <button
                          class="dropdown-item danger"
                          @click.stop="handleUninstallFromMenu(item)"
                          :title="t('common.delete') || '卸载此主题包'"
                        >
                          <TechIcon id="workbench:delete" :size="13" />
                          <span>{{ t('common.delete') || '卸载此主题包' }}</span>
                        </button>
                      </div>
                    </transition>
                  </div>
                </div>
                <button
                  v-else
                  class="btn-install-theme"
                  :disabled="mpIsInstalling(item)"
                  @click="onInstallTheme(item)"
                >
                  <span v-if="mpIsInstalling(item)" class="spinner-small"></span>
                  <span>{{ mpIsInstalling(item) ? (t("settings.installing") || "安装中...") : (t("settings.install") || "安装") }}</span>
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- Themes High-Density List / Table View -->
        <div v-else class="market-list-table-wrapper glass-card">
          <table class="market-table">
            <thead>
              <tr>
                <th class="th-extension">{{ t("marketplace.col_extension") || "扩展与主题" }}</th>
                <th class="th-publisher">{{ t("marketplace.col_publisher") || "发布者与分类" }}</th>
                <th class="th-desc">{{ t("marketplace.col_desc") || "描述" }}</th>
                <th class="th-stats">{{ t("marketplace.col_stats") || "数据" }}</th>
                <th class="th-actions">{{ t("marketplace.col_actions") || "操作" }}</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="item in displayedExtensions"
                :key="(item.namespace || '') + '.' + item.name"
                class="market-table-row clickable-row"
                :class="{ 'row-is-active': isExtensionActive(item) }"
                @click="navigateToDetail(item.namespace || '', item.name)"
              >
                <!-- Col: Extension Info -->
                <td class="td-extension">
                  <div class="table-ext-cell">
                    <div class="table-icon-wrap">
                      <img
                        v-if="hasValidIcon(item)"
                        :src="item.files!.icon"
                        class="table-img-icon"
                        loading="lazy"
                        alt="icon"
                        @error="handleIconError(`${item.namespace || ''}.${item.name}`)"
                      />
                      <span v-else class="table-fallback-icon">
                        <TechIcon id="workbench:marketplace" :size="16" />
                      </span>
                    </div>
                    <div class="table-ext-info">
                      <div class="table-ext-title-row">
                        <span class="ext-name">{{ item.displayName || item.name }}</span>
                        <span class="ext-ver font-mono">v{{ item.version }}</span>
                        <span v-if="isExtensionActive(item)" class="badge-active-pill small">
                          <span class="active-dot"></span>
                          <span>{{ t('marketplace.active') || '生效中' }}</span>
                        </span>
                      </div>
                    </div>
                  </div>
                </td>

                <!-- Col: Publisher & Category -->
                <td class="td-publisher">
                  <div class="table-publisher-cell">
                    <span class="publisher-ns font-mono">{{ item.namespace }}</span>
                    <span
                      class="card-type-tag"
                      :class="{
                        'type-product-icon': getCardDisplayType(item) === 'product_icon',
                        'type-icon': getCardDisplayType(item) === 'icon',
                        'type-color': getCardDisplayType(item) === 'theme'
                      }"
                      style="display: inline-flex; align-items: center; gap: 4px;"
                    >
                      <TechIcon :id="getCardDisplayType(item) === 'theme' ? 'workbench:marketplace' : 'workbench:tools'" :size="11" :colored="false" />
                      {{
                        getCardDisplayType(item) === 'product_icon'
                          ? (t('marketplace.filter_product_icons') || '界面图标')
                          : (getCardDisplayType(item) === 'icon'
                              ? (t('marketplace.filter_icons') || '文件图标')
                              : (t('marketplace.filter_themes') || '颜色主题'))
                      }}
                    </span>
                  </div>
                </td>

                <!-- Col: Description -->
                <td class="td-desc">
                  <span class="table-desc-text" :title="item.description">
                    {{ item.description || "No description provided." }}
                  </span>
                </td>

                <!-- Col: Stats -->
                <td class="td-stats">
                  <div class="table-stats-cell">
                    <span class="stat-pill" :title="'Downloads: ' + item.downloadCount">
                      ⬇️ {{ formatDownloads(item.downloadCount) }}
                    </span>
                    <span v-if="item.averageRating" class="stat-pill">
                      ⭐ {{ item.averageRating.toFixed(1) }}
                    </span>
                  </div>
                </td>

                <!-- Col: Actions -->
                <td class="td-actions" @click.stop>
                  <div class="table-actions-cell">
                    <div v-if="isItemInstalled(item)" class="installed-group">
                      <button
                        v-if="!isExtensionActive(item)"
                        class="btn btn-capsule btn-apply-theme"
                        :title="t('settings.apply_theme') || '应用此主题'"
                        @click="onApplyTheme(item)"
                      >
                        {{ t("settings.apply") || "应用" }}
                      </button>
                      <button
                        v-else
                        class="btn btn-capsule applied btn-applied-current"
                        disabled
                        :title="t('marketplace.current_active') || '当前正在生效的主题'"
                      >
                        {{ t('marketplace.currently_applied') || '已应用' }}
                      </button>
                      <!-- More Menu for Installed Themes (Uninstall) -->
                      <div
                        v-if="activeThemeTypeTab === 'installed'"
                        class="theme-more-menu-wrap"
                        @click.stop
                      >
                        <button
                          class="btn-theme-more table-more-btn"
                          :class="{ 'is-active': activeMenuThemePkgId === (item.package_id || item.name) }"
                          @click.stop="toggleThemeMenu(item.package_id || item.name)"
                          :title="t('common.more_actions') || '更多操作'"
                        >
                          <TechIcon id="workbench:more" :size="13" />
                        </button>
                        <transition name="dropdown-fade">
                          <div
                            v-if="activeMenuThemePkgId === (item.package_id || item.name)"
                            class="theme-dropdown-menu table-menu"
                            @click.stop
                          >
                            <button
                              class="dropdown-item danger"
                              @click.stop="handleUninstallFromMenu(item)"
                              :title="t('common.delete') || '卸载此主题包'"
                            >
                              <TechIcon id="workbench:delete" :size="13" />
                              <span>{{ t('common.delete') || '卸载此主题包' }}</span>
                            </button>
                          </div>
                        </transition>
                      </div>
                    </div>
                    <button
                      v-else
                      class="btn btn-capsule btn-install-theme"
                      :disabled="mpIsInstalling(item)"
                      @click="onInstallTheme(item)"
                    >
                      <span v-if="mpIsInstalling(item)" class="spinner-small"></span>
                      <span>{{ mpIsInstalling(item) ? (t("settings.installing") || "安装中...") : (t("settings.install") || "安装") }}</span>
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Pagination Bar -->
        <div v-if="selectedFilter !== 'installed' && mpTotalPages > 1" class="market-pagination-bar glass-card">
          <div class="pagination-summary">
            <span>{{ t('marketplace.total_count', { count: mpTotalSize }) || `共收录 ${mpTotalSize} 款主题与图标包` }}</span>
            <span class="pagination-divider">•</span>
            <span>{{ t('marketplace.page_info', { current: mpCurrentPage, total: mpTotalPages }) || `第 ${mpCurrentPage} / ${mpTotalPages} 页` }}</span>
          </div>

          <div class="pagination-buttons">
            <button
              class="btn-page-nav"
              :disabled="mpCurrentPage <= 1 || mpIsSearching"
              :title="t('marketplace.first_page') || '第一页'"
              @click="onPageChange(1)"
            >
              «
            </button>
            <button
              class="btn-page-nav"
              :disabled="mpCurrentPage <= 1 || mpIsSearching"
              :title="t('marketplace.prev_page') || '上一页'"
              @click="onPageChange(mpCurrentPage - 1)"
            >
              ‹
            </button>

            <template v-for="(p, idx) in visiblePages" :key="idx">
              <span v-if="p === -1" class="page-ellipsis">…</span>
              <button
                v-else
                class="btn-page-num"
                :class="{ active: p === mpCurrentPage }"
                :disabled="mpIsSearching"
                @click="onPageChange(p)"
              >
                {{ p }}
              </button>
            </template>

            <button
              class="btn-page-nav"
              :disabled="mpCurrentPage >= mpTotalPages || mpIsSearching"
              :title="t('marketplace.next_page') || '下一页'"
              @click="onPageChange(mpCurrentPage + 1)"
            >
              ›
            </button>
            <button
              class="btn-page-nav"
              :disabled="mpCurrentPage >= mpTotalPages || mpIsSearching"
              :title="t('marketplace.last_page') || '最后一页'"
              @click="onPageChange(mpTotalPages)"
            >
              »
            </button>
          </div>

          <div class="pagination-tools">
            <div class="page-jump-box">
              <span>{{ t('marketplace.jump_to') || '跳至' }}</span>
              <input
                v-model.number="jumpPageInput"
                type="number"
                min="1"
                :max="mpTotalPages"
                class="jump-input"
                :placeholder="t('marketplace.page_info', { current: '', total: '' }).replace(/[^0-9a-zA-Z]/g, '') || '页码'"
                @keydown.enter="onJumpPage"
              />
              <button class="btn-jump-submit" :disabled="mpIsSearching" @click="onJumpPage">
                {{ t('marketplace.jump_btn') || '跳转' }}
              </button>
            </div>

            <div class="page-size-box">
              <select
                :value="mpPageSize"
                class="page-size-select"
                :disabled="mpIsSearching"
                @change="onChangePageSize(Number(($event.target as HTMLSelectElement).value))"
              >
                <option :value="12">{{ t('marketplace.page_size_item', { size: 12 }) || '12 条 / 页' }}</option>
                <option :value="24">{{ t('marketplace.page_size_item', { size: 24 }) || '24 条 / 页' }}</option>
                <option :value="48">{{ t('marketplace.page_size_item', { size: 48 }) || '48 条 / 页' }}</option>
              </select>
            </div>
          </div>
        </div>
      </div>

      <!-- CHANNEL 2: Native Plugins & Ecosystem (UniRTM Extensions) -->
      <div v-show="currentChannel === 'plugins'" class="market-channel-content">
        <!-- Ecosystem Banner -->
        <div class="plugin-isolation-banner glass-card">
          <span class="banner-icon">
            <TechIcon id="workbench:lock" :size="20" />
          </span>
          <div class="banner-text">
            <strong>三层隔离架构与工具提供者准则</strong>
            <p>
              UniRTM 原生功能插件运行于独立外部进程中，绝不臃肿核心运行时。所有插件配置与存储严格限定于 plugins.&lt;plugin-id&gt;.* 命名空间内。
            </p>
          </div>
        </div>

        <!-- Native Plugins Cards View -->
        <div v-if="viewMode === 'cards'" class="extensions-grid">
          <div
            v-for="ext in nativePlugins"
            :key="ext.id"
            class="glass-card plugin-card"
          >
            <div class="card-header-row">
              <div class="card-icon-box">
                <TechIcon :id="ext.id" :size="28" :colored="true" />
              </div>
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
                <span class="stat-item" style="display: inline-flex; align-items: center; gap: 4px;">
                  <TechIcon id="workbench:star" :size="12" />
                  <span>{{ ext.stars }}</span>
                </span>
                <span class="stat-item" style="display: inline-flex; align-items: center; gap: 4px;">
                  <TechIcon id="workbench:package" :size="12" />
                  <span>{{ ext.downloads }}</span>
                </span>
              </div>

              <div class="card-actions">
                <span v-if="ext.installed" class="badge-installed">✓ {{ t("settings.installed") || "已启用" }}</span>
                <span v-else class="badge-soon">开发中</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Native Plugins High-Density List / Table View -->
        <div v-else class="market-list-table-wrapper glass-card">
          <table class="market-table">
            <thead>
              <tr>
                <th class="th-extension">{{ t("marketplace.col_extension") || "插件" }}</th>
                <th class="th-publisher">{{ t("marketplace.col_publisher") || "作者与标签" }}</th>
                <th class="th-desc">{{ t("marketplace.col_desc") || "描述" }}</th>
                <th class="th-stats">{{ t("marketplace.col_stats") || "数据" }}</th>
                <th class="th-actions">{{ t("marketplace.col_actions") || "状态" }}</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="ext in nativePlugins"
                :key="ext.id"
                class="market-table-row"
              >
                <!-- Col: Extension Info -->
                <td class="td-extension">
                  <div class="table-ext-cell">
                    <TechIcon :id="ext.id" :size="20" :colored="true" />
                    <div class="table-ext-info">
                      <div class="table-ext-title-row">
                        <span class="ext-name">{{ ext.name }}</span>
                        <span class="ext-ver font-mono">v{{ ext.version }}</span>
                      </div>
                    </div>
                  </div>
                </td>

                <!-- Col: Author & Tags -->
                <td class="td-publisher">
                  <div class="table-publisher-cell">
                    <span class="publisher-ns font-mono">{{ ext.author }}</span>
                    <div class="plugin-tags-row small" style="margin-top: 2px;">
                      <span v-for="tag in ext.tags" :key="tag" class="plugin-tag">{{ tag }}</span>
                    </div>
                  </div>
                </td>

                <!-- Col: Description -->
                <td class="td-desc">
                  <span class="table-desc-text" :title="ext.description">
                    {{ ext.description }}
                  </span>
                </td>

                <!-- Col: Stats -->
                <td class="td-stats">
                  <div class="table-stats-cell">
                    <span class="stat-pill" style="display: inline-flex; align-items: center; gap: 4px;">
                      <TechIcon id="workbench:star" :size="11" />
                      <span>{{ ext.stars }}</span>
                    </span>
                    <span class="stat-pill" style="display: inline-flex; align-items: center; gap: 4px;">
                      <TechIcon id="workbench:package" :size="11" />
                      <span>{{ ext.downloads }}</span>
                    </span>
                  </div>
                </td>

                <!-- Col: Status -->
                <td class="td-actions">
                  <span v-if="ext.installed" class="badge-installed">✓ {{ t("settings.installed") || "已启用" }}</span>
                  <span v-else class="badge-soon">开发中</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { t } from "../i18n";
import { useThemeMarketplace } from "../composables/useThemeMarketplace";
import { useThemeManager } from "../composables/useThemeManager";
import { useToast } from "../composables/useToast";
import { renderMarkdown } from "../utils/markdown";
import type { MarketplaceExtension, MarketplaceExtensionDetail } from "../types/theme-manager";
import TechIcon from "../components/common/TechIcon.vue";

const route = useRoute();
const router = useRouter();
const { showToast } = useToast();

// Channels: 'themes' (Open VSX) vs 'plugins' (UniRTM Native Plugins)
const currentChannel = ref<"themes" | "plugins">("themes");

// View mode: 'cards' or 'list'
const viewMode = ref<"cards" | "list">(
  (localStorage.getItem("unirtm_view_mode_marketplace") as "cards" | "list") || "cards"
);
function setViewMode(mode: "cards" | "list") {
  viewMode.value = mode;
  localStorage.setItem("unirtm_view_mode_marketplace", mode);
}

// Detail view state
const activeDetailTab = ref<"readme" | "contributes" | "changelog">("readme");
const isUninstalling = ref<boolean>(false);
const changelogContent = ref<string>("");

// Failed icons tracking to ensure fallback and img are strictly mutually exclusive
const failedIconKeys = ref<Set<string>>(new Set());

function handleIconError(key: string) {
  if (key) {
    failedIconKeys.value.add(key);
  }
}

function hasValidIcon(item: { namespace?: string; name?: string; files?: { icon?: string } } | null | undefined): boolean {
  if (!item || !item.files?.icon) return false;
  const key = `${item.namespace || ""}.${item.name || ""}`;
  return !failedIconKeys.value.has(key);
}

// --- Theme Marketplace Composable Integration ---
const {
  searchQuery: mpSearchQuery,
  searchResults: mpSearchResults,
  totalSize: mpTotalSize,
  currentPage: mpCurrentPage,
  pageSize: mpPageSize,
  totalPages: mpTotalPages,
  isSearching: mpIsSearching,
  marketplaceError: mpError,
  activeDetail,
  isLoadingDetail,
  searchThemes: mpSearchThemes,
  goToPage: mpGoToPage,
  setPageSize: mpSetPageSize,
  isExtensionInstalling: mpIsInstalling,
  installThemeOnline: mpInstallThemeOnline,
  fetchExtensionDetail,
  clearActiveDetail,
} = useThemeMarketplace();

const {
  colorThemes,
  iconThemes,
  productIconThemes,
  activeColorTheme,
  activeIconTheme,
  activeProductIconTheme,
  fetchThemes,
  setActiveTheme,
  deleteTheme,
} = useThemeManager();

type ThemeTypeTab = "all" | "color" | "icon" | "product_icon" | "installed";
const activeThemeTypeTab = ref<ThemeTypeTab>("all");
const selectedInstalledSubFilter = ref<"all" | "color" | "icon" | "product_icon">("all");
const selectedFilter = ref<string>("all");

interface ThemeChipItem {
  id: string;
  labelKey?: string;
  name?: string;
  query: string;
  isInstalled?: boolean;
}

const contextualChips = computed<ThemeChipItem[]>(() => {
  if (activeThemeTypeTab.value === "installed") {
    return [
      { id: "installed-all", name: t("marketplace.filter_all") || "全部已安装", query: "", isInstalled: true },
      { id: "installed-color", name: "🎨 " + (t("marketplace.filter_themes") || "颜色主题"), query: "", isInstalled: true },
      { id: "installed-icon", name: "📁 " + (t("marketplace.filter_icons") || "文件图标"), query: "", isInstalled: true },
      { id: "installed-product", name: "⚡ " + (t("marketplace.filter_product_icons") || "界面图标"), query: "", isInstalled: true },
    ];
  }

  if (activeThemeTypeTab.value === "color") {
    return [
      { id: "color-all", name: t("marketplace.filter_all") || "全部颜色主题", query: "" },
      { id: "dracula", name: "Dracula", query: "Dracula" },
      { id: "one-dark", name: "One Dark", query: "One Dark" },
      { id: "catppuccin", name: "Catppuccin", query: "Catppuccin" },
      { id: "tokyo-night", name: "Tokyo Night", query: "Tokyo Night" },
      { id: "github", name: "GitHub", query: "GitHub" },
      { id: "nord", name: "Nord", query: "Nord" },
      { id: "monokai", name: "Monokai", query: "Monokai" },
    ];
  }

  if (activeThemeTypeTab.value === "icon") {
    return [
      { id: "icon-all", name: t("marketplace.filter_all") || "全部文件图标", query: "icon" },
      { id: "material-icons", name: "Material Icon", query: "Material Icon Theme" },
      { id: "vscode-icons", name: "VSCode Icons", query: "vscode-icons" },
      { id: "seti", name: "Seti Icons", query: "seti-icons" },
      { id: "catppuccin-icons", name: "Catppuccin Icons", query: "catppuccin icons" },
      { id: "symbols", name: "Symbols", query: "symbols" },
      { id: "ayu-icons", name: "Ayu Icons", query: "ayu icon" },
    ];
  }

  if (activeThemeTypeTab.value === "product_icon") {
    return [
      { id: "product-all", name: t("marketplace.filter_all") || "全部界面图标", query: "product icon" },
      { id: "fluent-icons", name: "Fluent Icons", query: "fluent-icons" },
      { id: "carbon-icons", name: "Carbon Product Icons", query: "carbon product icons" },
      { id: "codicons", name: "Codicons", query: "codicons" },
    ];
  }

  // "all"
  return [
    { id: "all", labelKey: "marketplace.filter_all", query: "" },
    { id: "fluent", name: "Fluent Icons (界面)", query: "fluent-icons" },
    { id: "material", name: "Material (文件)", query: "Material Icon" },
    { id: "dracula", name: "Dracula (颜色)", query: "Dracula" },
    { id: "one-dark", name: "One Dark", query: "One Dark" },
    { id: "catppuccin", name: "Catppuccin", query: "Catppuccin" },
    { id: "tokyo-night", name: "Tokyo Night", query: "Tokyo Night" },
    { id: "github", name: "GitHub", query: "GitHub" },
  ];
});

const searchPlaceholder = computed(() => {
  if (activeThemeTypeTab.value === "installed") {
    return "搜索已安装的颜色主题、文件图标或界面图标...";
  }
  if (activeThemeTypeTab.value === "color") {
    return "在 Open VSX 搜索颜色主题 (如 Dracula, One Dark, Catppuccin)...";
  }
  if (activeThemeTypeTab.value === "icon") {
    return "在 Open VSX 搜索文件图标主题 (如 Material Icon, VSCode Icons)...";
  }
  if (activeThemeTypeTab.value === "product_icon") {
    return "在 Open VSX 搜索界面图标主题 (如 Fluent Icons, Codicons)...";
  }
  return t("settings.marketplace_search_placeholder") || "在 Open VSX 搜索主题或图标集 (如 Dracula, Material, Catppuccin)...";
});

const emptyStateMessage = computed(() => {
  if (activeThemeTypeTab.value === "installed") {
    return t("marketplace.empty_installed") || "暂未安装任何第三方主题或图标包";
  }
  if (activeThemeTypeTab.value === "color") {
    return "未找到符合条件的颜色主题";
  }
  if (activeThemeTypeTab.value === "icon") {
    return "未找到符合条件的文件图标主题";
  }
  if (activeThemeTypeTab.value === "product_icon") {
    return "未找到符合条件的界面图标主题";
  }
  return t("marketplace.empty_results") || "未找到符合条件的主题或图标包";
});

const installedExtensions = computed<MarketplaceExtension[]>(() => {
  const list: MarketplaceExtension[] = [];
  const seen = new Set<string>();

  const processThemeItem = (themeItem: any, fallbackDesc: string) => {
    if (themeItem.is_builtin || !themeItem.package_id) return;
    const pkg = themeItem.package_id.toLowerCase();
    if (seen.has(pkg)) return;
    seen.add(pkg);

    const match = mpSearchResults.value.find(
      (s) => (s.package_id && s.package_id.toLowerCase() === pkg) ||
             s.name.toLowerCase() === pkg
    );
    if (match) {
      list.push({ ...match, is_installed: true });
    } else {
      list.push({
        namespace: themeItem.package_id.includes(".") ? themeItem.package_id.split(".")[0] : "",
        name: themeItem.package_id.includes(".") ? themeItem.package_id.split(".")[1] : themeItem.package_id,
        displayName: themeItem.name || themeItem.package_id,
        version: themeItem.version || "",
        description: fallbackDesc,
        downloadCount: 0,
        averageRating: 5.0,
        is_installed: true,
        package_id: themeItem.package_id,
        files: {
          download: "",
          icon: "",
        },
      });
    }
  };

  for (const c of colorThemes.value) {
    processThemeItem(c, "已安装的本地 VS Code 颜色主题扩展包");
  }
  for (const i of iconThemes.value) {
    processThemeItem(i, "已安装的本地 VS Code 文件图标主题扩展包");
  }
  for (const p of productIconThemes.value) {
    processThemeItem(p, "已安装的本地 VS Code 界面图标主题扩展包");
  }

  return list;
});

const displayedExtensions = computed<MarketplaceExtension[]>(() => {
  let list: MarketplaceExtension[] = [];

  if (activeThemeTypeTab.value === "installed") {
    list = [...installedExtensions.value];
    if (selectedInstalledSubFilter.value === "color") {
      list = list.filter((item) => hasInstalledColorTheme(item) || getExtensionType(item) === "theme");
    } else if (selectedInstalledSubFilter.value === "icon") {
      list = list.filter((item) => hasInstalledIconTheme(item) || getExtensionType(item) === "icon");
    } else if (selectedInstalledSubFilter.value === "product_icon") {
      list = list.filter((item) => hasInstalledProductIconTheme(item) || getExtensionType(item) === "product_icon");
    }
    if (mpSearchQuery.value.trim()) {
      const q = mpSearchQuery.value.toLowerCase().trim();
      list = list.filter(
        (item) =>
          (item.displayName && item.displayName.toLowerCase().includes(q)) ||
          item.name.toLowerCase().includes(q) ||
          (item.description && item.description.toLowerCase().includes(q)) ||
          (item.package_id && item.package_id.toLowerCase().includes(q))
      );
    }
    return list;
  }

  list = [...mpSearchResults.value];

  // Strict Category Type filtering across upstream results
  if (activeThemeTypeTab.value === "color") {
    list = list.filter((item) => hasInstalledColorTheme(item) || getExtensionType(item) === "theme");
  } else if (activeThemeTypeTab.value === "icon") {
    list = list.filter((item) => hasInstalledIconTheme(item) || getExtensionType(item) === "icon");
  } else if (activeThemeTypeTab.value === "product_icon") {
    list = list.filter((item) => hasInstalledProductIconTheme(item) || getExtensionType(item) === "product_icon");
  }

  return list;
});

const feedbackMsg = ref("");
const feedbackIsError = ref(false);

// Markdown computed properties
const renderedReadme = computed(() => {
  if (!activeDetail.value?.readme) return "";
  return renderMarkdown(activeDetail.value.readme);
});

const renderedChangelog = computed(() => {
  if (!changelogContent.value) return "";
  return renderMarkdown(changelogContent.value);
});

const contributedThemesCount = computed(() => {
  const tCount = activeDetail.value?.contributes?.themes?.length || 0;
  const iCount = activeDetail.value?.contributes?.iconThemes?.length || 0;
  const pCount = activeDetail.value?.contributes?.productIconThemes?.length || 0;
  return tCount + iCount + pCount;
});

// Watch route params for detail page view
const loadCurrentDetail = async () => {
  const ns = route.params.namespace as string;
  const name = route.params.name as string;
  if (ns && name) {
    changelogContent.value = "";
    activeDetailTab.value = "readme";
    const detail = await fetchExtensionDetail(ns, name);
    if (detail?.files?.changelog) {
      try {
        const clRes = await fetch(detail.files.changelog);
        if (clRes.ok) {
          changelogContent.value = await clRes.text();
        }
      } catch {}
    }
  } else {
    clearActiveDetail();
  }
};

watch(
  () => [route.params.namespace, route.params.name],
  () => {
    loadCurrentDetail();
  },
  { immediate: true }
);

// Overflow action menu for installed themes
const activeMenuThemePkgId = ref<string | null>(null);

function toggleThemeMenu(pkgId: string) {
  activeMenuThemePkgId.value = activeMenuThemePkgId.value === pkgId ? null : pkgId;
}

function closeThemeMenu() {
  activeMenuThemePkgId.value = null;
}

function handleUninstallFromMenu(item: MarketplaceExtension) {
  closeThemeMenu();
  onUninstallThemeCard(item);
}

function handleGlobalClick() {
  if (activeMenuThemePkgId.value) {
    closeThemeMenu();
  }
}

function handleGlobalKeyDown(e: KeyboardEvent) {
  if (e.key === "Escape" && activeMenuThemePkgId.value) {
    closeThemeMenu();
  }
}

// Initialize channel and load installed themes on mount
onMounted(async () => {
  window.addEventListener("click", handleGlobalClick);
  window.addEventListener("keydown", handleGlobalKeyDown);
  if (route.query.tab === "plugins") {
    currentChannel.value = "plugins";
  } else {
    currentChannel.value = "themes";
  }
  // Load installed themes and active theme state
  await fetchThemes();
  // Trigger initial theme marketplace search if on list view
  if (!route.params.namespace || !route.params.name) {
    mpSearchThemes("");
  }
});

onUnmounted(() => {
  window.removeEventListener("click", handleGlobalClick);
  window.removeEventListener("keydown", handleGlobalKeyDown);
});

function switchChannel(channel: "themes" | "plugins") {
  currentChannel.value = channel;
}

// --- Pagination State & Handlers ---
const jumpPageInput = ref<number | string>("");

const visiblePages = computed(() => {
  const total = mpTotalPages.value;
  const current = mpCurrentPage.value;
  if (total <= 7) {
    return Array.from({ length: total }, (_, i) => i + 1);
  }
  if (current <= 4) {
    return [1, 2, 3, 4, 5, -1, total];
  }
  if (current >= total - 3) {
    return [1, -1, total - 4, total - 3, total - 2, total - 1, total];
  }
  return [1, -1, current - 1, current, current + 1, -1, total];
});

async function onPageChange(page: number) {
  if (page < 1 || page > mpTotalPages.value || page === mpCurrentPage.value) return;
  await mpGoToPage(page);
  window.scrollTo({ top: 0, behavior: "smooth" });
}

async function onJumpPage() {
  const target = Number(jumpPageInput.value);
  if (target >= 1 && target <= mpTotalPages.value) {
    jumpPageInput.value = "";
    await onPageChange(target);
  }
}

async function onChangePageSize(size: number) {
  await mpSetPageSize(size);
  window.scrollTo({ top: 0, behavior: "smooth" });
}

function hasInstalledColorTheme(item: { package_id?: string; name: string }): boolean {
  const pkgId = (item.package_id || item.name || "").toLowerCase();
  const name = (item.name || "").toLowerCase();
  return colorThemes.value.some((c) => !c.is_builtin && ((c.package_id && c.package_id.toLowerCase() === pkgId) || c.id.toLowerCase() === name || (c.package_id && c.package_id.toLowerCase() === name)));
}

function hasInstalledIconTheme(item: { package_id?: string; name: string }): boolean {
  const pkgId = (item.package_id || item.name || "").toLowerCase();
  const name = (item.name || "").toLowerCase();
  return iconThemes.value.some((i) => !i.is_builtin && ((i.package_id && i.package_id.toLowerCase() === pkgId) || i.id.toLowerCase() === name || (i.package_id && i.package_id.toLowerCase() === name)));
}

function hasInstalledProductIconTheme(item: { package_id?: string; name: string }): boolean {
  const pkgId = (item.package_id || item.name || "").toLowerCase();
  const name = (item.name || "").toLowerCase();
  return productIconThemes.value.some((p) => !p.is_builtin && ((p.package_id && p.package_id.toLowerCase() === pkgId) || p.id.toLowerCase() === name || (p.package_id && p.package_id.toLowerCase() === name)));
}

function getCardDisplayType(item: MarketplaceExtension): "product_icon" | "icon" | "theme" {
  if (activeThemeTypeTab.value === "product_icon") return "product_icon";
  if (activeThemeTypeTab.value === "icon") return "icon";
  if (activeThemeTypeTab.value === "color") return "theme";
  if (activeThemeTypeTab.value === "installed") {
    if (selectedInstalledSubFilter.value === "product_icon") return "product_icon";
    if (selectedInstalledSubFilter.value === "icon") return "icon";
    if (selectedInstalledSubFilter.value === "color") return "theme";
  }
  return getExtensionType(item);
}

function getExtensionType(item: MarketplaceExtension): "product_icon" | "icon" | "theme" {
  const pkgId = (item.package_id || item.name || "").toLowerCase();
  const name = (item.name || "").toLowerCase();
  const ns = (item.namespace || "").toLowerCase();
  const fullId = `${ns}.${name}`.toLowerCase();

  // 1. Calculate count of each contributed type from locally installed themes
  const colorCount = colorThemes.value.filter((c) => !c.is_builtin && ((c.package_id && c.package_id.toLowerCase() === pkgId) || c.id.toLowerCase() === name || (c.package_id && c.package_id.toLowerCase() === fullId))).length;
  const iconCount = iconThemes.value.filter((i) => !i.is_builtin && ((i.package_id && i.package_id.toLowerCase() === pkgId) || i.id.toLowerCase() === name || (i.package_id && i.package_id.toLowerCase() === fullId))).length;
  const productCount = productIconThemes.value.filter((p) => !p.is_builtin && ((p.package_id && p.package_id.toLowerCase() === pkgId) || p.id.toLowerCase() === name || (p.package_id && p.package_id.toLowerCase() === fullId))).length;

  if (colorCount > 0 || iconCount > 0 || productCount > 0) {
    if (productCount > colorCount && productCount > iconCount) return "product_icon";
    if (iconCount > colorCount && iconCount >= productCount) return "icon";
    if (colorCount >= iconCount && colorCount >= productCount) return "theme";
  }

  // 2. Curated known fingerprint sets to avoid any misclassification
  const KNOWN_PRODUCT_ICONS = [
    "fluent-icons",
    "miguelsolorio.fluent-icons",
    "carbon-product-icons",
    "codicons",
    "vscode-codicons",
    "jetbrains-product-icons",
    "workbench-icons",
  ];
  if (KNOWN_PRODUCT_ICONS.some((k) => name === k || fullId === k || pkgId === k || name.includes("product-icon") || name.includes("producticon"))) {
    return "product_icon";
  }

  const KNOWN_FILE_ICONS = [
    "material-icon-theme",
    "pkief.material-icon-theme",
    "vscode-icons",
    "vscode-icons-team.vscode-icons",
    "catppuccin-vsc-icons",
    "catppuccin.catppuccin-vsc-icons",
    "symbols",
    "miguelsolorio.symbols",
    "chalice-icon-theme",
    "ayu-icon-theme",
    "seti-icons",
  ];
  if (KNOWN_FILE_ICONS.some((k) => name === k || fullId === k || pkgId === k)) {
    return "icon";
  }

  const nameText = `${item.name} ${item.displayName || ""}`.toLowerCase();
  const desc = (item.description || "").toLowerCase();
  const fullText = `${nameText} ${desc}`;

  // 3. Check for product icon keywords
  if (
    fullText.includes("product icon") ||
    fullText.includes("product icons") ||
    fullText.includes("product-icon") ||
    fullText.includes("producticontheme") ||
    fullText.includes("workbench icon") ||
    fullText.includes("fluent-icons") ||
    fullText.includes("fluent icons") ||
    fullText.includes("codicons") ||
    fullText.includes("carbon product icons")
  ) {
    return "product_icon";
  }

  // 4. Check for file icon keywords
  if (
    fullText.includes("file icon") ||
    fullText.includes("file-icon") ||
    fullText.includes("folder icon") ||
    fullText.includes("folder-icon") ||
    fullText.includes("icon theme") ||
    nameText.includes("icon-theme") ||
    nameText.includes("icons-theme") ||
    nameText.includes("file icons") ||
    nameText.endsWith("-icons") ||
    nameText.endsWith("-icon") ||
    nameText.endsWith(" icons") ||
    nameText.endsWith(" icon")
  ) {
    return "icon";
  }

  return "theme";
}

function isItemInstalled(item: { namespace?: string; name: string; displayName?: string; package_id?: string; is_installed?: boolean }): boolean {
  if (item.is_installed) return true;
  const pkgId = (item.package_id || item.name || "").toLowerCase();
  const name = (item.name || "").toLowerCase();
  const ns = (item.namespace || "").toLowerCase();
  const fullId = `${ns}.${name}`.toLowerCase();
  const displayName = (item.displayName || "").toLowerCase();

  const isColor = colorThemes.value.some((c) => {
    if (c.is_builtin) return false;
    const cPkg = (c.package_id || "").toLowerCase();
    const cId = c.id.toLowerCase();
    const cName = c.name.toLowerCase();
    return cPkg === pkgId || cPkg === name || cPkg === fullId || cId === pkgId || cId === name || (Boolean(displayName) && cName === displayName);
  });
  if (isColor) return true;

  const isIcon = iconThemes.value.some((i) => {
    if (i.is_builtin) return false;
    const iPkg = (i.package_id || "").toLowerCase();
    const iId = i.id.toLowerCase();
    const iName = i.name.toLowerCase();
    return iPkg === pkgId || iPkg === name || iPkg === fullId || iId === pkgId || iId === name || (Boolean(displayName) && iName === displayName);
  });
  if (isIcon) return true;

  const isProduct = productIconThemes.value.some((p) => {
    if (p.is_builtin) return false;
    const pPkg = (p.package_id || "").toLowerCase();
    const pId = p.id.toLowerCase();
    const pName = p.name.toLowerCase();
    return pPkg === pkgId || pPkg === name || pPkg === fullId || pId === pkgId || pId === name || (Boolean(displayName) && pName === displayName);
  });
  return isProduct;
}

function isColorThemeActive(item: { namespace?: string; name: string; displayName?: string; package_id?: string; is_installed?: boolean }): boolean {
  if (!isItemInstalled(item) || !activeColorTheme.value) return false;
  const actColor = activeColorTheme.value.toLowerCase();
  const pkgId = (item.package_id || item.name || "").toLowerCase();
  const name = (item.name || "").toLowerCase();
  const ns = (item.namespace || "").toLowerCase();
  const fullId = `${ns}.${name}`.toLowerCase();
  const displayName = (item.displayName || "").toLowerCase();

  if (actColor === pkgId || actColor === name || actColor === fullId || actColor === displayName) return true;

  return colorThemes.value.some((c) => {
    if (c.is_builtin) return false;
    const cPkg = (c.package_id || "").toLowerCase();
    const cId = c.id.toLowerCase();
    const cName = c.name.toLowerCase();
    const belongs = cPkg === pkgId || cPkg === name || cPkg === fullId || cId === pkgId || cId === name || (Boolean(displayName) && cName === displayName);
    if (!belongs) return false;
    return cId === actColor || cName === actColor;
  });
}

function isFileIconActive(item: { namespace?: string; name: string; displayName?: string; package_id?: string; is_installed?: boolean }): boolean {
  if (!isItemInstalled(item) || !activeIconTheme.value) return false;
  const actIcon = activeIconTheme.value.toLowerCase();
  const pkgId = (item.package_id || item.name || "").toLowerCase();
  const name = (item.name || "").toLowerCase();
  const ns = (item.namespace || "").toLowerCase();
  const fullId = `${ns}.${name}`.toLowerCase();
  const displayName = (item.displayName || "").toLowerCase();

  if (actIcon === pkgId || actIcon === name || actIcon === fullId || actIcon === displayName) return true;

  return iconThemes.value.some((i) => {
    if (i.is_builtin) return false;
    const iPkg = (i.package_id || "").toLowerCase();
    const iId = i.id.toLowerCase();
    const iName = i.name.toLowerCase();
    const belongs = iPkg === pkgId || iPkg === name || iPkg === fullId || iId === pkgId || iId === name || (Boolean(displayName) && iName === displayName);
    if (!belongs) return false;
    return iId === actIcon || iName === actIcon;
  });
}

function isProductIconActive(item: { namespace?: string; name: string; displayName?: string; package_id?: string; is_installed?: boolean }): boolean {
  if (!isItemInstalled(item) || !activeProductIconTheme.value) return false;
  const actProd = activeProductIconTheme.value.toLowerCase();
  const pkgId = (item.package_id || item.name || "").toLowerCase();
  const name = (item.name || "").toLowerCase();
  const ns = (item.namespace || "").toLowerCase();
  const fullId = `${ns}.${name}`.toLowerCase();
  const displayName = (item.displayName || "").toLowerCase();

  if (actProd === pkgId || actProd === name || actProd === fullId || actProd === displayName) return true;

  return productIconThemes.value.some((p) => {
    if (p.is_builtin) return false;
    const pPkg = (p.package_id || "").toLowerCase();
    const pId = p.id.toLowerCase();
    const pName = p.name.toLowerCase();
    const belongs = pPkg === pkgId || pPkg === name || pPkg === fullId || pId === pkgId || pId === name || (Boolean(displayName) && pName === displayName);
    if (!belongs) return false;
    return pId === actProd || pName === actProd;
  });
}

function isExtensionActive(
  item: { namespace?: string; name: string; displayName?: string; package_id?: string; is_installed?: boolean },
  explicitType?: "product_icon" | "icon" | "theme"
): boolean {
  if (!isItemInstalled(item)) return false;

  const targetType = explicitType || getCardDisplayType(item as MarketplaceExtension);
  if (targetType === "product_icon") {
    return isProductIconActive(item);
  }
  if (targetType === "icon") {
    return isFileIconActive(item);
  }
  return isColorThemeActive(item);
}

function isSubColorActive(label?: string, id?: string): boolean {
  if (!activeColorTheme.value) return false;
  const act = activeColorTheme.value.toLowerCase();
  return (Boolean(label) && label!.toLowerCase() === act) || (Boolean(id) && id!.toLowerCase() === act);
}

function isSubIconActive(label?: string, id?: string): boolean {
  if (!activeIconTheme.value) return false;
  const act = activeIconTheme.value.toLowerCase();
  return (Boolean(label) && label!.toLowerCase() === act) || (Boolean(id) && id!.toLowerCase() === act);
}

function isSubProductIconActive(label?: string, id?: string): boolean {
  if (!activeProductIconTheme.value) return false;
  const act = activeProductIconTheme.value.toLowerCase();
  return (Boolean(label) && label!.toLowerCase() === act) || (Boolean(id) && id!.toLowerCase() === act);
}

const isDetailActive = computed(() => {
  if (!activeDetail.value || !isItemInstalled(activeDetail.value)) return false;
  return isExtensionActive(activeDetail.value);
});

async function onApplyDetailTheme(detail: MarketplaceExtensionDetail) {
  const item: MarketplaceExtension = {
    namespace: detail.namespace,
    name: detail.name,
    displayName: detail.displayName || detail.name,
    version: detail.version,
    description: detail.description,
    downloadCount: detail.downloadCount,
    averageRating: detail.averageRating,
    files: detail.files,
    is_installed: detail.is_installed,
    package_id: detail.package_id || detail.name,
  };
  await onApplyTheme(item);
}

function onSearchThemes(q?: string) {
  if (activeThemeTypeTab.value === "installed") {
    return;
  }
  const query = q !== undefined ? q : mpSearchQuery.value;
  const match = contextualChips.value.find((c) => !c.isInstalled && c.query.toLowerCase() === query.toLowerCase().trim());
  selectedFilter.value = match ? match.id : (query ? "" : (contextualChips.value[0]?.id || "all"));
  mpSearchThemes(query);
}

function onSelectContextualChip(chip: ThemeChipItem) {
  selectedFilter.value = chip.id;
  if (activeThemeTypeTab.value === "installed") {
    if (chip.id === "installed-color") selectedInstalledSubFilter.value = "color";
    else if (chip.id === "installed-icon") selectedInstalledSubFilter.value = "icon";
    else if (chip.id === "installed-product") selectedInstalledSubFilter.value = "product_icon";
    else selectedInstalledSubFilter.value = "all";
    return;
  }
  mpSearchQuery.value = chip.query;
  mpSearchThemes(chip.query);
}

function switchThemeTypeTab(tab: ThemeTypeTab) {
  activeThemeTypeTab.value = tab;
  selectedInstalledSubFilter.value = "all";

  if (tab === "installed") {
    selectedFilter.value = "installed-all";
    mpSearchQuery.value = "";
    return;
  }

  if (tab === "color") {
    selectedFilter.value = "color-all";
    if (mpSearchQuery.value === "icon" || mpSearchQuery.value === "product icon" || mpSearchQuery.value === "fluent-icons") {
      mpSearchQuery.value = "";
    }
    mpSearchThemes(mpSearchQuery.value);
  } else if (tab === "icon") {
    selectedFilter.value = "icon-all";
    if (!mpSearchQuery.value || mpSearchQuery.value === "product icon") {
      mpSearchQuery.value = "icon";
    }
    mpSearchThemes(mpSearchQuery.value);
  } else if (tab === "product_icon") {
    selectedFilter.value = "product-all";
    if (!mpSearchQuery.value || mpSearchQuery.value === "icon") {
      mpSearchQuery.value = "product icon";
    }
    mpSearchThemes(mpSearchQuery.value);
  } else {
    selectedFilter.value = "all";
    mpSearchThemes(mpSearchQuery.value);
  }
}

function onClearSearch() {
  mpSearchQuery.value = "";
  if (activeThemeTypeTab.value === "installed") return;
  if (activeThemeTypeTab.value === "icon") {
    mpSearchThemes("icon");
  } else if (activeThemeTypeTab.value === "product_icon") {
    mpSearchThemes("product icon");
  } else {
    mpSearchThemes("");
  }
}

function onResetToPopular() {
  switchThemeTypeTab("all");
}

function navigateToDetail(namespace: string, name: string) {
  if (!namespace) {
    return;
  }
  router.push(`/marketplace/${namespace}/${name}`);
}

function goBackToList() {
  router.push({ path: "/marketplace", query: { tab: currentChannel.value } });
}

async function onUninstallThemeCard(item: MarketplaceExtension) {
  const name = item.displayName || item.name;
  const confirmMsg = t("marketplace.uninstall_confirm", { name }) || `确定要卸载主题包 "${name}" 吗？`;
  if (!confirm(confirmMsg)) {
    return;
  }
  const pkgId = item.package_id || item.name;
  feedbackMsg.value = "";
  feedbackIsError.value = false;
  try {
    await deleteTheme("color", pkgId);
    await deleteTheme("icon", pkgId);
    await deleteTheme("product", pkgId);
    item.is_installed = false;
    const found = mpSearchResults.value.find(
      (s) => s.name === item.name || (s.package_id && s.package_id === pkgId)
    );
    if (found) found.is_installed = false;
    await fetchThemes();
    feedbackMsg.value = t("marketplace.uninstalled_success", { name }) || `✓ 已卸载 ${name}`;
    feedbackIsError.value = false;
  } catch (err: any) {
    feedbackMsg.value = `✕ 卸载失败: ${err.message || err}`;
    feedbackIsError.value = true;
  }
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
      } else if (first.type === "product") {
        await setActiveTheme(undefined, undefined, undefined, first.id);
      }
    }
  } catch (err: any) {
    feedbackIsError.value = true;
    feedbackMsg.value = `✕ 安装失败: ${err.message || err}`;
  }
}

async function onInstallDetail(detail: MarketplaceExtensionDetail) {
  const item: MarketplaceExtension = {
    namespace: detail.namespace,
    name: detail.name,
    displayName: detail.displayName || detail.name,
    version: detail.version,
    description: detail.description,
    downloadCount: detail.downloadCount,
    averageRating: detail.averageRating,
    files: {
      download: detail.files.download,
      icon: detail.files.icon,
    },
    package_id: detail.package_id || detail.name,
  };
  await onInstallTheme(item);
  detail.is_installed = true;
}

async function onUninstallDetail(detail: MarketplaceExtensionDetail) {
  const pkgId = detail.package_id || detail.name;
  isUninstalling.value = true;
  feedbackMsg.value = "";
  feedbackIsError.value = false;
  try {
    await deleteTheme("color", pkgId);
    await deleteTheme("icon", pkgId);
    await deleteTheme("product", pkgId);
    detail.is_installed = false;
    const found = mpSearchResults.value.find((s) => s.namespace === detail.namespace && s.name === detail.name);
    if (found) found.is_installed = false;
    feedbackMsg.value = `✓ 已卸载 ${detail.displayName || detail.name}`;
  } catch (err: any) {
    feedbackMsg.value = `✕ 卸载失败: ${err.message || err}`;
    feedbackIsError.value = true;
  } finally {
    isUninstalling.value = false;
  }
}

async function onApplySubColorTheme(label: string) {
  const match = colorThemes.value.find(
    (c) => c.name.toLowerCase() === label.toLowerCase() || c.id.toLowerCase() === label.toLowerCase()
  );
  if (match) {
    await setActiveTheme(match.id, undefined);
    await fetchThemes();
    feedbackMsg.value = `✓ 已应用颜色主题: ${match.name}`;
    feedbackIsError.value = false;
    showToast(`已应用颜色主题: ${match.name}`, "success");
  } else {
    feedbackMsg.value = `已激活主题包，可在设置中选择应用 ${label}`;
    feedbackIsError.value = false;
    showToast(`已激活主题包: ${label}`, "info");
  }
}

async function onApplySubIconTheme(label: string) {
  const match = iconThemes.value.find(
    (i) => i.name.toLowerCase() === label.toLowerCase() || i.id.toLowerCase() === label.toLowerCase()
  );
  if (match) {
    await setActiveTheme(undefined, match.id);
    await fetchThemes();
    feedbackMsg.value = `✓ 已应用文件图标主题: ${match.name}`;
    feedbackIsError.value = false;
    showToast(`已应用文件图标主题: ${match.name}`, "success");
  } else {
    feedbackMsg.value = `已激活文件图标包，可在设置中选择应用 ${label}`;
    feedbackIsError.value = false;
    showToast(`已激活文件图标包: ${label}`, "info");
  }
}

async function onApplySubProductIconTheme(labelOrId: string) {
  const match = productIconThemes.value.find(
    (p) => p.name.toLowerCase() === labelOrId.toLowerCase() || p.id.toLowerCase() === labelOrId.toLowerCase()
  );
  if (match) {
    await setActiveTheme(undefined, undefined, undefined, match.id);
    await fetchThemes();
    feedbackMsg.value = `✓ 已应用界面图标主题: ${match.name}`;
    feedbackIsError.value = false;
    showToast(`已应用界面图标主题: ${match.name}`, "success");
  } else {
    feedbackMsg.value = `已激活界面图标包，可在设置中选择应用 ${labelOrId}`;
    feedbackIsError.value = false;
    showToast(`已激活界面图标包: ${labelOrId}`, "info");
  }
}

async function onApplyTheme(item: MarketplaceExtension) {
  try {
    // If local themes list isn't populated yet, fetch first
    if (colorThemes.value.length <= 4 && iconThemes.value.length <= 2) {
      await fetchThemes();
    }

    const pkgID = (item.package_id || item.name || "").toLowerCase();
    const name = item.name.toLowerCase();
    const fullId = `${item.namespace || ""}.${item.name}`.toLowerCase();
    const displayName = (item.displayName || "").toLowerCase();

    // Find matching color theme
    const colorMatch = colorThemes.value.find((c) => {
      const cPkg = (c.package_id || "").toLowerCase();
      const cId = c.id.toLowerCase();
      const cName = c.name.toLowerCase();
      return cPkg === pkgID || cPkg === name || cPkg === fullId || cId === pkgID || cId === name || cName === displayName;
    });

    // Find matching icon theme
    const iconMatch = iconThemes.value.find((i) => {
      const iPkg = (i.package_id || "").toLowerCase();
      const iId = i.id.toLowerCase();
      const iName = i.name.toLowerCase();
      return iPkg === pkgID || iPkg === name || iPkg === fullId || iId === pkgID || iId === name || iName === displayName;
    });

    // Find matching product icon theme
    const productMatch = productIconThemes.value.find((p) => {
      const pPkg = (p.package_id || "").toLowerCase();
      const pId = p.id.toLowerCase();
      const pName = p.name.toLowerCase();
      return pPkg === pkgID || pPkg === name || pPkg === fullId || pId === pkgID || pId === name || pName === displayName;
    });

    const targetType = getCardDisplayType(item);

    if (targetType === "product_icon") {
      if (productMatch) {
        await setActiveTheme(undefined, undefined, undefined, productMatch.id);
      } else {
        await setActiveTheme(undefined, undefined, undefined, item.name);
      }
    } else if (targetType === "icon") {
      if (iconMatch) {
        await setActiveTheme(undefined, iconMatch.id);
      } else {
        await setActiveTheme(undefined, item.name);
      }
    } else {
      if (colorMatch) {
        await setActiveTheme(colorMatch.id, undefined);
      } else {
        await setActiveTheme(item.name, undefined);
      }
    }

    // Refresh themes so all UI states sync
    await fetchThemes();

    const targetName = item.displayName || item.name;
    feedbackMsg.value = `已启用主题: ${targetName}`;
    feedbackIsError.value = false;
    showToast(`已启用并应用主题: ${targetName}`, "success");
  } catch (err: any) {
    feedbackMsg.value = `激活主题失败: ${err.message || err}`;
    feedbackIsError.value = true;
    showToast(`激活主题失败: ${err.message || err}`, "error");
  }
}

function formatDownloads(count: number): string {
  if (!count) return "0";
  if (count >= 1000000) return (count / 1000000).toFixed(1) + "M";
  if (count >= 1000) return (count / 1000).toFixed(1) + "k";
  return String(count);
}

function formatUITheme(uiTheme?: string): string {
  if (uiTheme === "vs") return "浅色 (Light)";
  if (uiTheme === "hc-black" || uiTheme === "hc-light") return "高对比度";
  return "深色 (Dark)";
}

function formatDate(dateStr?: string): string {
  if (!dateStr) return "近期更新";
  try {
    const d = new Date(dateStr);
    return d.toLocaleDateString();
  } catch {
    return dateStr;
  }
}

function cleanGitUrl(url: string): string {
  return url.replace(/^git\+/, "").replace(/\.git$/, "");
}

// --- Channel 2: UniRTM Ecosystem Native Plugins ---
interface PluginItem {
  id: string;
  name: string;
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

/* Header & Tabs */
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
  gap: 6px;
}

.channel-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  border: 1px solid transparent;
  border-radius: 6px;
  background: transparent;
  color: var(--text-muted);
  font-size: 0.84rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.channel-btn:hover {
  color: var(--text-main);
  background: rgba(255, 255, 255, 0.05);
}

.channel-btn.active {
  color: var(--accent-cyan);
  background: rgba(0, 229, 255, 0.12);
  border-color: rgba(0, 229, 255, 0.3);
}

/* Channel Content */
.market-channel-content {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

/* Structural Type Dimension Tabs */
.market-type-tabs {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 10px;
  border-radius: 8px;
  overflow-x: auto;
}

.type-tab-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 7px 16px;
  border: 1px solid transparent;
  border-radius: 6px;
  background: transparent;
  color: var(--text-muted);
  font-size: 0.86rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  white-space: nowrap;
}

.type-tab-btn:hover {
  color: var(--text-main);
  background: rgba(255, 255, 255, 0.05);
}

.type-tab-btn.active {
  color: var(--accent-cyan);
  background: var(--accent-cyan-glow, rgba(0, 229, 255, 0.14));
  border-color: var(--card-border-active, rgba(0, 229, 255, 0.35));
  box-shadow: 0 0 12px var(--accent-cyan-glow, rgba(0, 229, 255, 0.15));
  font-weight: 600;
}

.type-tab-btn.type-tab-installed.active {
  color: var(--success, #10b981);
  background: rgba(16, 185, 129, 0.16);
  border-color: var(--success, #10b981);
  box-shadow: 0 0 12px rgba(16, 185, 129, 0.15);
  font-weight: 600;
}

.type-tab-btn .tab-badge {
  background: rgba(16, 185, 129, 0.2);
  color: #10b981;
  border: 1px solid rgba(16, 185, 129, 0.4);
  border-radius: 10px;
  padding: 1px 7px;
  font-size: 0.72rem;
  font-weight: 600;
}

/* Filter Bar */
.market-filter-bar {
  display: flex;
  gap: 10px;
  padding: 10px 14px;
  align-items: center;
}

.search-input-wrapper {
  position: relative;
  flex: 1;
  display: flex;
  align-items: center;
}

.search-icon {
  position: absolute;
  left: 12px;
  font-size: 0.9rem;
  opacity: 0.6;
  pointer-events: none;
}

.market-search-input {
  width: 100%;
  padding: 8px 36px 8px 36px;
  background: var(--input-bg, rgba(0, 0, 0, 0.2));
  border: 1px solid var(--card-border, rgba(255, 255, 255, 0.1));
  border-radius: 6px;
  color: var(--text-main);
  font-size: 0.88rem;
  outline: none;
  transition: border-color 0.2s;
}

.market-search-input:focus {
  border-color: var(--accent-cyan);
}

.btn-clear-search {
  position: absolute;
  right: 12px;
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
}

.btn-search-action {
  padding: 8px 20px;
  background: rgba(0, 229, 255, 0.14);
  border: 1px solid rgba(0, 229, 255, 0.35);
  color: var(--accent-cyan);
  border-radius: 6px;
  font-size: 0.86rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-search-action:hover:not(:disabled) {
  background: rgba(0, 229, 255, 0.25);
  border-color: var(--accent-cyan);
}

/* Chips Bar */
.market-chips-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
}

.chips-left {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.chips-label {
  font-size: 0.8rem;
  color: var(--text-muted);
}

.chip-btn {
  background: var(--card-bg, rgba(255, 255, 255, 0.03));
  border: 1px solid var(--card-border, rgba(255, 255, 255, 0.1));
  color: var(--text-muted);
  padding: 4px 12px;
  border-radius: 14px;
  font-size: 0.78rem;
  cursor: pointer;
  transition: all 0.2s;
}

.chip-btn:hover {
  color: var(--text-main);
  border-color: rgba(255, 255, 255, 0.25);
}

.chip-btn.active {
  background: rgba(0, 229, 255, 0.12);
  border-color: var(--accent-cyan);
  color: var(--accent-cyan);
}

.chip-btn.chip-installed {
  font-weight: 600;
}

.chip-btn.chip-installed.active {
  background: rgba(16, 185, 129, 0.15);
  border-color: #10b981;
  color: #10b981;
}

.results-count-label {
  font-size: 0.78rem;
  color: var(--text-muted);
}

/* Banners */
.feedback-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 14px;
  background: rgba(16, 185, 129, 0.12);
  border: 1px solid rgba(16, 185, 129, 0.35);
  color: var(--success, #10b981);
  border-radius: 6px;
  font-size: 0.84rem;
}

.feedback-banner.error {
  background: rgba(239, 68, 68, 0.12);
  border-color: rgba(239, 68, 68, 0.35);
  color: #ef4444;
}

.btn-close-banner {
  background: transparent;
  border: none;
  color: inherit;
  cursor: pointer;
}

.error-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px;
  background: rgba(239, 68, 68, 0.12);
  border: 1px solid rgba(239, 68, 68, 0.35);
  color: #ef4444;
  border-radius: 6px;
  font-size: 0.84rem;
}

.btn-retry {
  background: transparent;
  border: 1px solid rgba(239, 68, 68, 0.4);
  color: #ef4444;
  padding: 2px 10px;
  border-radius: 4px;
  cursor: pointer;
}

/* Loading & Empty */
.market-loading-box,
.market-empty-box,
.detail-loading-box,
.detail-error-box {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 40px 20px;
  color: var(--text-muted);
  font-size: 0.88rem;
}

.spinner-small {
  display: inline-block;
  width: 12px;
  height: 12px;
  border: 2px solid rgba(255, 255, 255, 0.25);
  border-top-color: currentColor;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  vertical-align: middle;
  flex-shrink: 0;
  box-sizing: border-box;
}

.market-loading-box .spinner-small,
.detail-loading-box .spinner-small {
  width: 26px;
  height: 26px;
  border-width: 3px;
  border-color: rgba(0, 229, 255, 0.2);
  border-top-color: var(--accent-cyan);
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.empty-icon,
.error-icon {
  font-size: 2.2rem;
}

.btn-reset-chip,
.btn-action-primary {
  padding: 6px 16px;
  background: rgba(0, 229, 255, 0.14);
  border: 1px solid rgba(0, 229, 255, 0.35);
  color: var(--accent-cyan);
  border-radius: 6px;
  font-size: 0.84rem;
  cursor: pointer;
}

.btn-action-secondary {
  padding: 6px 16px;
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  color: var(--text-main);
  border-radius: 6px;
  font-size: 0.84rem;
  cursor: pointer;
}

.detail-error-actions {
  display: flex;
  gap: 10px;
  margin-top: 6px;
}

/* Extensions Grid */
.extensions-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(290px, 1fr));
  gap: 14px;
}

.theme-card,
.plugin-card {
  position: relative;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  padding: 14px;
  border-radius: 10px;
  transition: transform 0.2s ease, border-color 0.2s ease, box-shadow 0.2s ease, background 0.2s ease;
}

.clickable-card {
  cursor: pointer;
}

.clickable-card:hover {
  transform: translateY(-2px);
  border-color: rgba(0, 229, 255, 0.4);
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.2);
}

.card-is-active {
  position: relative;
  border: 2px solid #00e5ff !important;
  background: linear-gradient(145deg, rgba(0, 229, 255, 0.22) 0%, rgba(14, 165, 233, 0.14) 40%, rgba(15, 23, 42, 0.96) 100%) !important;
  box-shadow: 0 0 26px rgba(0, 229, 255, 0.42), 0 8px 24px rgba(0, 0, 0, 0.5), inset 0 0 18px rgba(0, 229, 255, 0.15) !important;
}

.card-is-active::before {
  content: "";
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: linear-gradient(90deg, #00e5ff, #38bdf8, #818cf8);
  box-shadow: 0 0 14px rgba(0, 229, 255, 0.9);
  pointer-events: none;
  z-index: 2;
}

.card-is-active:hover {
  transform: translateY(-2px);
  box-shadow: 0 0 34px rgba(0, 229, 255, 0.55), 0 10px 28px rgba(0, 0, 0, 0.6), inset 0 0 22px rgba(0, 229, 255, 0.2) !important;
  border-color: #00e5ff !important;
}

.card-is-active .card-name {
  color: #00e5ff !important;
  font-weight: 700 !important;
  text-shadow: 0 0 12px rgba(0, 229, 255, 0.45);
}

.card-is-active .card-icon-wrap,
.card-is-active .card-icon-box {
  border-color: #00e5ff !important;
  background: rgba(0, 229, 255, 0.18) !important;
  box-shadow: 0 0 14px rgba(0, 229, 255, 0.4) !important;
}

/* Light Mode Overrides for Active Theme Card */
:root[data-theme="light"] .card-is-active,
[data-theme="light"] .card-is-active {
  border: 2px solid #0284c7 !important;
  background: linear-gradient(145deg, rgba(2, 132, 199, 0.14) 0%, rgba(240, 249, 255, 0.97) 100%) !important;
  box-shadow: 0 6px 20px rgba(2, 132, 199, 0.25), inset 0 0 14px rgba(2, 132, 199, 0.08) !important;
}

:root[data-theme="light"] .card-is-active::before,
[data-theme="light"] .card-is-active::before {
  background: linear-gradient(90deg, #0284c7, #38bdf8);
  box-shadow: 0 0 10px rgba(2, 132, 199, 0.5);
}

:root[data-theme="light"] .card-is-active .card-name,
[data-theme="light"] .card-is-active .card-name {
  color: #0284c7 !important;
  text-shadow: none;
}

:root[data-theme="light"] .card-is-active .card-icon-wrap,
[data-theme="light"] .card-is-active .card-icon-box {
  border-color: #0284c7 !important;
  background: rgba(2, 132, 199, 0.1) !important;
  box-shadow: 0 0 10px rgba(2, 132, 199, 0.25) !important;
}

.card-header-row {
  display: flex;
  gap: 12px;
  align-items: flex-start;
}

/* Theme More Menu and Dropdown Styles (Installed Tab) */
.theme-more-menu-wrap {
  position: relative;
  display: inline-flex;
}

.btn-theme-more {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  padding: 0;
  border-radius: 6px;
  border: 1px solid rgba(148, 163, 184, 0.25);
  background: rgba(148, 163, 184, 0.08);
  color: var(--text-muted);
  cursor: pointer;
  transition: all 0.18s cubic-bezier(0.4, 0, 0.2, 1);
  flex-shrink: 0;
}

.btn-theme-more:hover,
.btn-theme-more.is-active {
  background: rgba(2, 132, 199, 0.15);
  border-color: var(--accent-cyan);
  color: var(--accent-cyan);
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(2, 132, 199, 0.2);
}

[data-theme="light"] .btn-theme-more {
  background: #f8fafc;
  border-color: #cbd5e1;
  color: #64748b;
}

[data-theme="light"] .btn-theme-more:hover,
[data-theme="light"] .btn-theme-more.is-active {
  background: rgba(2, 132, 199, 0.08);
  border-color: var(--accent-cyan);
  color: var(--accent-cyan);
}

.btn-theme-more.table-more-btn {
  width: 26px;
  height: 26px;
}

.theme-dropdown-menu {
  position: absolute;
  bottom: calc(100% + 6px);
  right: 0;
  background: var(--card-bg, #1e293b);
  border: 1px solid var(--card-border, rgba(255, 255, 255, 0.1));
  border-radius: 8px;
  box-shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.35), 0 0 12px var(--accent-cyan-glow, rgba(2, 132, 199, 0.15));
  padding: 0.35rem;
  z-index: 50;
  min-width: 140px;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.theme-dropdown-menu.table-menu {
  top: calc(100% + 4px);
  bottom: auto;
  right: 0;
  text-align: left;
}

[data-theme="light"] .theme-dropdown-menu {
  background: #ffffff;
  border-color: #e2e8f0;
  box-shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.1), 0 0 12px rgba(2, 132, 199, 0.1);
}

.theme-dropdown-menu .dropdown-item {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  padding: 0.5rem 0.75rem;
  font-size: 0.82rem;
  font-weight: 500;
  color: var(--text-main);
  background: transparent;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  width: 100%;
  text-align: left;
  transition: all 0.15s ease;
  white-space: nowrap;
}

.theme-dropdown-menu .dropdown-item:hover {
  background: var(--input-bg, rgba(255, 255, 255, 0.06));
  color: var(--text-main);
}

[data-theme="light"] .theme-dropdown-menu .dropdown-item:hover {
  background: #f1f5f9;
}

.theme-dropdown-menu .dropdown-item.danger {
  color: #ef4444;
}

.theme-dropdown-menu .dropdown-item.danger:hover {
  background: rgba(239, 68, 68, 0.12);
  color: #ef4444;
}

[data-theme="dark"] .theme-dropdown-menu .dropdown-item.danger {
  color: #f87171;
}

.card-icon-wrap,
.card-icon-box {
  width: 44px;
  height: 44px;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid var(--card-border, rgba(255, 255, 255, 0.1));
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  position: relative;
  flex-shrink: 0;
  font-size: 1.4rem;
  padding: 4px;
  box-sizing: border-box;
}

.card-img-icon {
  width: 100%;
  height: 100%;
  object-fit: contain;
  border-radius: 6px;
  display: block;
}

.card-fallback-icon {
  font-size: 1.3rem;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  color: var(--text-muted);
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
  gap: 8px;
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
  font-size: 0.72rem;
  color: var(--text-muted);
  background: rgba(255, 255, 255, 0.06);
  padding: 1px 6px;
  border-radius: 4px;
  flex-shrink: 0;
}

.card-subtitle-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 6px;
  margin-top: 2px;
}

.card-namespace {
  font-size: 0.76rem;
  color: var(--text-muted);
}

.card-type-tag {
  font-size: 0.7rem;
  padding: 1px 6px;
  border-radius: 4px;
  font-weight: 500;
  white-space: nowrap;
}

.card-type-tag.type-color {
  background: rgba(168, 85, 247, 0.12);
  color: #c084fc;
  border: 1px solid rgba(168, 85, 247, 0.25);
}

.card-type-tag.type-icon {
  background: rgba(6, 182, 212, 0.12);
  color: #22d3ee;
  border: 1px solid rgba(6, 182, 212, 0.25);
}

.card-type-tag.type-product-icon {
  background: rgba(245, 158, 11, 0.12);
  color: #fbbf24;
  border: 1px solid rgba(245, 158, 11, 0.25);
}

.card-desc {
  font-size: 0.8rem;
  color: var(--text-muted);
  margin: 10px 0;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  min-height: 2.25em;
}

.card-footer-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 6px;
  padding-top: 10px;
  border-top: 1px solid rgba(255, 255, 255, 0.05);
}

.card-stats {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 0.76rem;
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
  font-size: 0.74rem;
  padding: 2px 7px;
  background: rgba(16, 185, 129, 0.15);
  border: 1px solid rgba(16, 185, 129, 0.35);
  color: var(--success, #10b981);
  border-radius: 4px;
}

.badge-active-pill {
  font-size: 0.72rem;
  padding: 2px 8px;
  border-radius: 12px;
  background: rgba(0, 229, 255, 0.22);
  border: 1px solid #00e5ff;
  color: #00e5ff;
  font-weight: 700;
  display: inline-flex;
  align-items: center;
  gap: 5px;
  flex-shrink: 0;
  box-shadow: 0 0 10px rgba(0, 229, 255, 0.45);
  letter-spacing: 0.2px;
}

.active-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #00e5ff;
  box-shadow: 0 0 8px #00e5ff;
  animation: activeDotPulse 1.8s infinite ease-in-out;
}

@keyframes activeDotPulse {
  0%, 100% {
    opacity: 1;
    transform: scale(1);
    box-shadow: 0 0 8px #00e5ff;
  }
  50% {
    opacity: 0.35;
    transform: scale(0.8);
    box-shadow: 0 0 2px #00e5ff;
  }
}

.badge-active-applied {
  font-size: 0.74rem;
  padding: 3px 8px;
  background: rgba(0, 229, 255, 0.2);
  border: 1px solid #00e5ff;
  color: #00e5ff;
  border-radius: 4px;
  font-weight: 700;
  box-shadow: 0 0 10px rgba(0, 229, 255, 0.35);
}

.badge-hero-applied {
  font-size: 0.88rem;
  padding: 8px 18px;
  background: rgba(0, 229, 255, 0.22);
  border: 1.5px solid #00e5ff;
  color: #00e5ff;
  border-radius: 8px;
  font-weight: 700;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  box-shadow: 0 0 16px rgba(0, 229, 255, 0.45);
}

.btn-applied-current {
  padding: 4px 12px;
  background: #00e5ff !important;
  border: 1px solid #00e5ff !important;
  color: #070a12 !important;
  border-radius: 4px;
  font-size: 0.76rem;
  font-weight: 700;
  cursor: default;
  box-shadow: 0 0 14px rgba(0, 229, 255, 0.5) !important;
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

/* Light Mode Overrides for Active Badges, Tabs & Buttons */
:root[data-theme="light"] .type-tab-btn:hover,
[data-theme="light"] .type-tab-btn:hover {
  color: var(--text-main);
  background: rgba(0, 0, 0, 0.05);
}

:root[data-theme="light"] .type-tab-btn.active,
[data-theme="light"] .type-tab-btn.active {
  color: #0284c7 !important;
  background: rgba(2, 132, 199, 0.12) !important;
  border-color: #0284c7 !important;
  box-shadow: 0 1px 6px rgba(2, 132, 199, 0.2) !important;
  font-weight: 600 !important;
}

:root[data-theme="light"] .type-tab-btn.active span,
[data-theme="light"] .type-tab-btn.active span {
  color: inherit !important;
}

:root[data-theme="light"] .type-tab-btn.type-tab-installed.active,
[data-theme="light"] .type-tab-btn.type-tab-installed.active {
  color: #059669 !important;
  background: rgba(5, 150, 105, 0.12) !important;
  border-color: #059669 !important;
  box-shadow: 0 1px 6px rgba(5, 150, 105, 0.2) !important;
  font-weight: 600 !important;
}

:root[data-theme="light"] .type-tab-btn.type-tab-installed.active span,
[data-theme="light"] .type-tab-btn.type-tab-installed.active span {
  color: inherit !important;
}

:root[data-theme="light"] .type-tab-btn .tab-badge,
[data-theme="light"] .type-tab-btn .tab-badge {
  background: rgba(5, 150, 105, 0.12);
  color: #059669;
  border-color: rgba(5, 150, 105, 0.35);
}

:root[data-theme="light"] .channel-btn:hover,
[data-theme="light"] .channel-btn:hover {
  color: var(--text-main);
  background: rgba(0, 0, 0, 0.05);
}

:root[data-theme="light"] .channel-btn.active,
[data-theme="light"] .channel-btn.active {
  color: #0284c7;
  background: rgba(2, 132, 199, 0.12);
  border-color: rgba(2, 132, 199, 0.4);
  font-weight: 600;
}

:root[data-theme="light"] .btn-search-action,
[data-theme="light"] .btn-search-action {
  background: rgba(2, 132, 199, 0.1);
  border-color: rgba(2, 132, 199, 0.35);
  color: #0284c7;
}

:root[data-theme="light"] .btn-search-action:hover:not(:disabled),
[data-theme="light"] .btn-search-action:hover:not(:disabled) {
  background: rgba(2, 132, 199, 0.18);
  border-color: #0284c7;
}

:root[data-theme="light"] .chip-btn.active,
[data-theme="light"] .chip-btn.active {
  background: rgba(2, 132, 199, 0.1);
  border-color: rgba(2, 132, 199, 0.35);
  color: #0284c7;
}

:root[data-theme="light"] .chip-btn.chip-installed.active,
[data-theme="light"] .chip-btn.chip-installed.active {
  background: rgba(5, 150, 105, 0.1);
  border-color: rgba(5, 150, 105, 0.35);
  color: #047857;
}

:root[data-theme="light"] .btn-page-num.active,
[data-theme="light"] .btn-page-num.active {
  background: #0284c7;
  border-color: #0284c7;
  color: #ffffff;
  box-shadow: 0 2px 8px rgba(2, 132, 199, 0.25);
}

:root[data-theme="light"] .btn-install-primary,
[data-theme="light"] .btn-install-primary,
:root[data-theme="light"] .btn-apply-detail-primary,
[data-theme="light"] .btn-apply-detail-primary {
  background: #0284c7;
  border-color: #0284c7;
  color: #ffffff !important;
  box-shadow: 0 2px 10px rgba(2, 132, 199, 0.25);
}

:root[data-theme="light"] .subtab-btn:hover,
[data-theme="light"] .subtab-btn:hover {
  color: var(--text-main);
  background: rgba(0, 0, 0, 0.05);
}

:root[data-theme="light"] .subtab-btn.active,
[data-theme="light"] .subtab-btn.active {
  color: #0284c7;
  background: rgba(2, 132, 199, 0.12);
  border-color: rgba(2, 132, 199, 0.35);
  font-weight: 600;
}

:root[data-theme="light"] .badge-active-pill,
[data-theme="light"] .badge-active-pill {
  background: rgba(2, 132, 199, 0.12);
  border-color: #0284c7;
  color: #0284c7;
  box-shadow: 0 0 6px rgba(2, 132, 199, 0.2);
}

:root[data-theme="light"] .badge-active-pill .active-dot,
[data-theme="light"] .badge-active-pill .active-dot {
  background: #0284c7;
  box-shadow: 0 0 6px #0284c7;
}

:root[data-theme="light"] .badge-active-applied,
[data-theme="light"] .badge-active-applied {
  background: rgba(2, 132, 199, 0.15);
  border-color: #0284c7;
  color: #0284c7;
  box-shadow: none;
}

:root[data-theme="light"] .btn-applied-current,
[data-theme="light"] .btn-applied-current {
  background: #0284c7 !important;
  border-color: #0284c7 !important;
  color: #ffffff !important;
  box-shadow: 0 2px 8px rgba(2, 132, 199, 0.3) !important;
}

:root[data-theme="light"] .badge-hero-applied,
[data-theme="light"] .badge-hero-applied {
  background: rgba(2, 132, 199, 0.12);
  border-color: #0284c7;
  color: #0284c7;
  box-shadow: 0 2px 10px rgba(2, 132, 199, 0.2);
}

.btn-apply-detail-primary {
  padding: 8px 18px;
  background: var(--accent-cyan);
  border: 1px solid var(--accent-cyan);
  color: #0b1120;
  border-radius: 6px;
  font-size: 0.88rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-apply-detail-primary:hover {
  opacity: 0.92;
  transform: translateY(-1px);
}

.badge-soon {
  font-size: 0.74rem;
  padding: 2px 7px;
  background: rgba(255, 255, 255, 0.06);
  color: var(--text-muted);
  border-radius: 4px;
}

.btn-apply-theme {
  padding: 4px 10px;
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
  background: rgba(0, 229, 255, 0.25);
  border-color: var(--accent-cyan);
}

.btn-install-theme:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* Plugin Isolation Banner */
.plugin-isolation-banner {
  display: flex;
  gap: 14px;
  align-items: center;
  padding: 12px 18px;
  background: rgba(0, 229, 255, 0.06);
  border-color: rgba(0, 229, 255, 0.2);
}

.banner-icon {
  font-size: 1.6rem;
}

.banner-text strong {
  display: block;
  font-size: 0.9rem;
  color: var(--accent-cyan);
  margin-bottom: 2px;
}

.banner-text p {
  margin: 0;
  font-size: 0.8rem;
  color: var(--text-muted);
  line-height: 1.4;
}

.plugin-tags-row {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
  margin: 4px 0 10px;
}

.plugin-tag {
  font-size: 0.72rem;
  padding: 2px 6px;
  border-radius: 4px;
  background: rgba(255, 255, 255, 0.06);
  color: var(--text-muted);
}

/* ========================================================================= */
/* MARKETPLACE PAGINATION STYLES */
/* ========================================================================= */
.market-pagination-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 14px 20px;
  margin-top: 1.5rem;
  flex-wrap: wrap;
  font-size: 0.88rem;
  color: var(--text-muted);
  border-radius: 12px;
}

.pagination-summary {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.86rem;
}

.pagination-summary strong {
  color: var(--text-main);
  font-weight: 600;
}

.pagination-divider {
  opacity: 0.3;
}

.pagination-buttons {
  display: flex;
  align-items: center;
  gap: 6px;
}

.btn-page-nav,
.btn-page-num {
  min-width: 34px;
  height: 34px;
  padding: 0 8px;
  border-radius: 6px;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid var(--card-border);
  color: var(--text-muted);
  font-size: 0.88rem;
  font-weight: 500;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s;
  user-select: none;
}

.btn-page-nav:hover:not(:disabled),
.btn-page-num:hover:not(:disabled) {
  background: rgba(0, 229, 255, 0.12);
  border-color: rgba(0, 229, 255, 0.4);
  color: var(--accent-cyan);
}

.btn-page-num.active {
  background: var(--accent-cyan);
  border-color: var(--accent-cyan);
  color: #0b1120;
  font-weight: 600;
  box-shadow: 0 2px 8px rgba(0, 229, 255, 0.3);
}

.btn-page-nav:disabled,
.btn-page-num:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.page-ellipsis {
  padding: 0 4px;
  color: var(--text-muted);
  opacity: 0.5;
  user-select: none;
}

.pagination-tools {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
}

.page-jump-box {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 0.84rem;
}

.jump-input {
  width: 54px;
  height: 32px;
  padding: 0 6px;
  text-align: center;
  border-radius: 6px;
  background: rgba(0, 0, 0, 0.2);
  border: 1px solid var(--card-border);
  color: var(--text-main);
  font-size: 0.84rem;
  outline: none;
  transition: border-color 0.2s;
}

.jump-input:focus {
  border-color: var(--accent-cyan);
}

.jump-input::-webkit-inner-spin-button,
.jump-input::-webkit-outer-spin-button {
  -webkit-appearance: none;
  margin: 0;
}

.btn-jump-submit {
  height: 32px;
  padding: 0 10px;
  background: rgba(255, 255, 255, 0.08);
  border: 1px solid var(--card-border);
  color: var(--text-main);
  border-radius: 6px;
  font-size: 0.82rem;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-jump-submit:hover:not(:disabled) {
  background: rgba(0, 229, 255, 0.15);
  border-color: var(--accent-cyan);
  color: var(--accent-cyan);
}

.btn-jump-submit:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.page-size-select {
  height: 32px;
  padding: 0 8px;
  border-radius: 6px;
  background: rgba(0, 0, 0, 0.2);
  border: 1px solid var(--card-border);
  color: var(--text-main);
  font-size: 0.82rem;
  outline: none;
  cursor: pointer;
  transition: border-color 0.2s;
}

.page-size-select:focus {
  border-color: var(--accent-cyan);
}

/* ========================================================================= */
/* DETAIL VIEW STYLING */
/* ========================================================================= */
.market-detail-container {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  max-width: 1440px;
  width: 100%;
  margin: 0 auto;
  box-sizing: border-box;
}

.detail-main-layout {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  width: 100%;
}

.detail-nav-bar {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 10px 18px;
  font-size: 0.88rem;
}

.btn-back-market {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 5px 12px;
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid var(--card-border);
  color: var(--text-main);
  border-radius: 6px;
  font-size: 0.84rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-back-market:hover {
  background: rgba(0, 229, 255, 0.14);
  border-color: var(--accent-cyan);
  color: var(--accent-cyan);
  transform: translateX(-2px);
}

.breadcrumb-trail {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--text-muted);
}

.crumb-link {
  cursor: pointer;
}

.crumb-link:hover {
  color: var(--accent-cyan);
  text-decoration: underline;
}

.crumb-sep {
  opacity: 0.4;
}

.crumb-curr {
  color: var(--text-main);
  font-weight: 500;
}

/* Detail Hero Header */
.detail-hero {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  padding: 24px 28px;
  flex-wrap: wrap;
}

.hero-left {
  display: flex;
  gap: 20px;
  align-items: flex-start;
  flex: 1;
  min-width: 340px;
}

.hero-icon-box {
  width: 80px;
  height: 80px;
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid var(--card-border);
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  overflow: hidden;
  flex-shrink: 0;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.2);
}

.hero-img-icon {
  width: 100%;
  height: 100%;
  object-fit: cover;
  position: absolute;
  top: 0;
  left: 0;
}

.hero-fallback-icon {
  font-size: 2.6rem;
}

.hero-info {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.hero-title-row {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.hero-title {
  font-size: 1.55rem;
  font-weight: 700;
  margin: 0;
  color: var(--text-main);
  letter-spacing: -0.2px;
}

.hero-version {
  font-size: 0.8rem;
  background: rgba(255, 255, 255, 0.08);
  padding: 2px 8px;
  border-radius: 4px;
  color: var(--text-muted);
}

.hero-meta-row {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 0.84rem;
  color: var(--text-muted);
  flex-wrap: wrap;
}

.meta-dot {
  opacity: 0.35;
}

.hero-desc {
  font-size: 0.88rem;
  color: var(--text-muted);
  margin: 4px 0 0;
  line-height: 1.5;
  max-width: 800px;
}

.hero-actions {
  display: flex;
  align-items: center;
  gap: 14px;
  flex-wrap: wrap;
}

.btn-install-primary {
  padding: 9px 26px;
  background: var(--accent-cyan);
  border: 1px solid var(--accent-cyan);
  color: #0b1120;
  border-radius: 6px;
  font-size: 0.92rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  box-shadow: 0 2px 10px rgba(0, 229, 255, 0.25);
}

.btn-install-primary:hover:not(:disabled) {
  opacity: 0.92;
  transform: translateY(-1px);
  box-shadow: 0 4px 16px rgba(0, 229, 255, 0.35);
}

.btn-install-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-uninstall-danger {
  padding: 8px 18px;
  background: rgba(239, 68, 68, 0.12);
  border: 1px solid rgba(239, 68, 68, 0.35);
  color: #ef4444;
  border-radius: 6px;
  font-size: 0.86rem;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-uninstall-danger:hover:not(:disabled) {
  background: rgba(239, 68, 68, 0.22);
}

.btn-external-link {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid var(--card-border);
  color: var(--text-muted);
  border-radius: 6px;
  font-size: 0.86rem;
  text-decoration: none;
  transition: all 0.2s;
}

.btn-external-link:hover {
  color: var(--text-main);
  border-color: rgba(255, 255, 255, 0.3);
  background: rgba(255, 255, 255, 0.08);
}

/* Detail Body Grid Layout - Spacious 2 Columns with Sticky Sidebar */
.detail-body-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 300px;
  gap: 1.5rem;
  align-items: flex-start;
  width: 100%;
}

@media (max-width: 960px) {
  .detail-body-grid {
    grid-template-columns: 1fr;
  }
}

.detail-body-main {
  padding: 24px 28px;
  min-height: 480px;
  min-width: 0;
  max-width: 100%;
  overflow: hidden;
  box-sizing: border-box;
}

/* Subtabs */
.detail-subtabs {
  display: flex;
  gap: 8px;
  border-bottom: 1px solid var(--card-border);
  padding-bottom: 12px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.subtab-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 7px 16px;
  background: transparent;
  border: 1px solid transparent;
  border-radius: 6px;
  color: var(--text-muted);
  font-size: 0.88rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.subtab-btn:hover {
  color: var(--text-main);
  background: rgba(255, 255, 255, 0.05);
}

.subtab-btn.active {
  color: var(--accent-cyan);
  background: rgba(0, 229, 255, 0.1);
  border-color: rgba(0, 229, 255, 0.25);
  font-weight: 600;
}

.tab-badge {
  font-size: 0.72rem;
  padding: 1px 7px;
  background: rgba(0, 229, 255, 0.2);
  border-radius: 10px;
  color: var(--accent-cyan);
}

/* Tab Content */
.tab-content {
  line-height: 1.7;
  min-width: 0;
  max-width: 100%;
  overflow-x: hidden;
}

.readme-empty-hint {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 48px;
  color: var(--text-muted);
  font-size: 0.9rem;
}

/* Contributed Themes List */
.contribute-group {
  margin-bottom: 24px;
}

.group-title {
  font-size: 1rem;
  color: var(--text-main);
  margin: 0 0 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  font-weight: 600;
}

.contribute-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.contribute-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 18px;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid var(--card-border);
  border-radius: 8px;
  transition: border-color 0.2s;
}

.contribute-item:hover {
  border-color: rgba(0, 229, 255, 0.3);
}

.sub-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.sub-label {
  font-size: 0.92rem;
  font-weight: 500;
  color: var(--text-main);
}

.sub-mode-tag {
  font-size: 0.74rem;
  padding: 3px 8px;
  border-radius: 4px;
}

.mode-dark {
  background: rgba(99, 102, 241, 0.15);
  color: #a5b4fc;
  border: 1px solid rgba(99, 102, 241, 0.3);
}

.mode-light {
  background: rgba(245, 158, 11, 0.15);
  color: #fcd34d;
  border: 1px solid rgba(245, 158, 11, 0.3);
}

.mode-icon {
  background: rgba(16, 185, 129, 0.15);
  color: #6ee7b7;
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.mode-product {
  background: rgba(245, 158, 11, 0.15);
  color: #fbbf24;
  border: 1px solid rgba(245, 158, 11, 0.3);
}

.btn-apply-sub {
  padding: 5px 14px;
  background: rgba(0, 229, 255, 0.12);
  border: 1px solid rgba(0, 229, 255, 0.3);
  color: var(--accent-cyan);
  border-radius: 5px;
  font-size: 0.8rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-apply-sub:hover {
  background: rgba(0, 229, 255, 0.25);
  border-color: var(--accent-cyan);
}

/* Sticky Sidebar */
.detail-sidebar {
  display: flex;
  flex-direction: column;
  gap: 20px;
  padding: 20px;
  position: sticky;
  top: 1rem;
}

.sidebar-heading {
  font-size: 0.82rem;
  text-transform: uppercase;
  letter-spacing: 0.6px;
  color: var(--text-muted);
  margin: 0 0 12px;
  font-weight: 600;
}

.sidebar-links {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.sidebar-link {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 0.84rem;
  color: var(--text-main);
  text-decoration: none;
  padding: 8px 10px;
  border-radius: 6px;
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid transparent;
  transition: all 0.2s;
}

.sidebar-link:hover {
  background: rgba(0, 229, 255, 0.1);
  border-color: rgba(0, 229, 255, 0.25);
  color: var(--accent-cyan);
  transform: translateX(2px);
}

.link-icon {
  font-size: 1rem;
}

.prop-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.prop-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.82rem;
  padding-bottom: 6px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.04);
}

.prop-key {
  color: var(--text-muted);
}

.prop-val {
  color: var(--text-main);
}

.sidebar-tags {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.tag-chip {
  font-size: 0.74rem;
  padding: 3px 9px;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid var(--card-border);
  border-radius: 12px;
  color: var(--text-muted);
}

/* ========================================================================= */
/* RENDERED MARKDOWN STYLES (Strict Container Fitting, Auto-Expanded Details) */
/* ========================================================================= */
:deep(.markdown-body) {
  color: var(--text-main);
  font-size: 0.92rem;
  line-height: 1.75;
  min-width: 0;
  max-width: 100%;
  overflow-x: hidden;
  word-break: break-word;
  overflow-wrap: break-word;
}

:deep(.markdown-body h1),
:deep(.markdown-body h2),
:deep(.markdown-body h3),
:deep(.markdown-body h4) {
  color: var(--text-main);
  margin-top: 1.6em;
  margin-bottom: 0.6em;
  font-weight: 600;
  letter-spacing: -0.2px;
}

:deep(.markdown-body h1) {
  font-size: 1.45rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  padding-bottom: 8px;
}

:deep(.markdown-body h2) {
  font-size: 1.25rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  padding-bottom: 6px;
}

:deep(.markdown-body h3) {
  font-size: 1.08rem;
}

:deep(.markdown-body p) {
  margin: 0.8em 0;
}

:deep(.markdown-body a.md-link),
:deep(.markdown-body a) {
  color: var(--accent-cyan);
  text-decoration: underline;
  text-underline-offset: 3px;
}

/* Robust Image Containment - NEVER blow up or break the container */
:deep(.markdown-body img),
:deep(.markdown-body img.md-img) {
  max-width: 100% !important;
  height: auto !important;
  display: block;
  margin: 16px auto;
  border-radius: 8px;
  box-sizing: border-box;
  object-fit: contain;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.25);
  border: 1px solid var(--card-border, rgba(255, 255, 255, 0.1));
}

/* Inline Badges & Shields (shields.io) should stay inline cleanly */
:deep(.markdown-body a img),
:deep(.markdown-body p > a > img),
:deep(.markdown-body p > img[src*="shields.io"]),
:deep(.markdown-body p > img[src*="badge"]) {
  display: inline-block !important;
  margin: 4px 6px 4px 0 !important;
  box-shadow: none !important;
  border: none !important;
  vertical-align: middle;
}

/* Centered paragraphs with images */
:deep(.markdown-body p[align="center"]) {
  text-align: center;
}

:deep(.markdown-body p[align="center"] img) {
  margin-left: auto;
  margin-right: auto;
}

/* Collapsible Details & Summaries (Expanded by default, styled beautifully) */
:deep(.markdown-body details.md-details),
:deep(.markdown-body details) {
  margin: 20px 0;
  border: 1px solid var(--card-border, rgba(255, 255, 255, 0.12));
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.02);
  overflow: hidden;
  transition: border-color 0.2s ease;
}

:deep(.markdown-body details[open]) {
  border-color: rgba(0, 229, 255, 0.35);
  background: rgba(0, 229, 255, 0.02);
}

:deep(.markdown-body summary.md-summary),
:deep(.markdown-body summary) {
  padding: 12px 18px;
  cursor: pointer;
  font-weight: 600;
  font-size: 0.94rem;
  color: var(--accent-cyan, #00e5ff);
  background: rgba(0, 229, 255, 0.06);
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
  user-select: none;
  transition: background 0.2s ease;
}

:deep(.markdown-body summary:hover) {
  background: rgba(0, 229, 255, 0.12);
}

:deep(.markdown-body details > *:not(summary)) {
  padding-left: 18px;
  padding-right: 18px;
}

/* Code Blocks & Inline Code */
:deep(.markdown-body code.md-inline-code),
:deep(.markdown-body code) {
  background: rgba(255, 255, 255, 0.08);
  padding: 2px 7px;
  border-radius: 4px;
  font-family: monospace;
  font-size: 0.84rem;
  color: var(--accent-cyan);
}

:deep(.markdown-body pre.md-code-block),
:deep(.markdown-body pre) {
  background: rgba(0, 0, 0, 0.4);
  border: 1px solid var(--card-border);
  border-radius: 8px;
  padding: 14px 16px;
  overflow-x: auto;
  font-family: monospace;
  font-size: 0.84rem;
  margin: 14px 0;
  max-width: 100%;
}

:deep(.markdown-body pre code) {
  background: transparent;
  padding: 0;
  color: inherit;
}

/* Blockquotes */
:deep(.markdown-body blockquote.md-blockquote),
:deep(.markdown-body blockquote) {
  margin: 14px 0;
  padding: 10px 16px;
  border-left: 3px solid var(--accent-cyan);
  background: rgba(0, 229, 255, 0.04);
  color: var(--text-muted);
  border-radius: 0 6px 6px 0;
}

/* Lists */
:deep(.markdown-body ul.md-ul),
:deep(.markdown-body ol.md-ol),
:deep(.markdown-body ul),
:deep(.markdown-body ol) {
  padding-left: 24px;
  margin: 8px 0;
}

:deep(.markdown-body li.md-li),
:deep(.markdown-body li) {
  margin: 4px 0;
}

:deep(.markdown-body hr.md-hr),
:deep(.markdown-body hr) {
  border: none;
  border-top: 1px solid var(--card-border);
  margin: 20px 0;
}

/* Tables */
:deep(.markdown-body .md-table-wrap) {
  overflow-x: auto;
  margin: 16px 0;
  border: 1px solid var(--card-border);
  border-radius: 8px;
}

:deep(.markdown-body table.md-table),
:deep(.markdown-body table) {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.86rem;
}

:deep(.markdown-body table th),
:deep(.markdown-body table td) {
  border: 1px solid var(--card-border);
  padding: 8px 14px;
  text-align: left;
}

:deep(.markdown-body table th) {
  background: rgba(255, 255, 255, 0.05);
  font-weight: 600;
  color: var(--text-main);
}

:deep(.markdown-body table tr:nth-child(even)) {
  background: rgba(255, 255, 255, 0.015);
}

/* Marketplace View Mode Switcher */
.view-mode-toggle {
  display: inline-flex;
  align-items: center;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  border-radius: 6px;
  padding: 2px;
  gap: 2px;
  margin-left: auto;
}

.view-mode-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 26px;
  height: 26px;
  padding: 0;
  border: none;
  background: transparent;
  color: var(--text-muted);
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.view-mode-btn:hover {
  color: var(--text-main);
  background: rgba(255, 255, 255, 0.05);
}

.view-mode-btn.active {
  color: var(--accent-cyan);
  background: rgba(0, 229, 255, 0.12);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
}

/* Marketplace High-Density Table Styles */
.market-list-table-wrapper {
  width: 100%;
  overflow-x: auto;
  border: 1px solid var(--card-border);
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.015);
  margin-bottom: 1.5rem;
}

.market-table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
  font-size: 0.84rem;
}

.market-table thead tr {
  background: rgba(255, 255, 255, 0.03);
  border-bottom: 1px solid var(--card-border);
}

.market-table th {
  padding: 0.75rem 1rem;
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: var(--text-muted);
  white-space: nowrap;
}

.market-table th.th-extension {
  width: 260px;
}
.market-table th.th-publisher {
  width: 170px;
}
.market-table th.th-desc {
  min-width: 220px;
}
.market-table th.th-stats {
  width: 130px;
}
.market-table th.th-actions {
  text-align: right;
  width: 180px;
}

.market-table-row {
  border-bottom: 1px solid var(--card-border, rgba(255, 255, 255, 0.05));
  transition: background 0.15s ease;
}

.market-table-row:last-child {
  border-bottom: none;
}

.market-table-row:hover {
  background: rgba(255, 255, 255, 0.035);
}

.market-table-row.clickable-row {
  cursor: pointer;
}

.market-table-row.row-is-active {
  border-left: 3px solid var(--accent-cyan);
  background: rgba(0, 229, 255, 0.05);
}

.market-table td {
  padding: 0.65rem 1rem;
  vertical-align: middle;
}

.market-table td.td-actions {
  text-align: right;
}

.table-ext-cell {
  display: inline-flex;
  align-items: center;
  gap: 10px;
}

.table-icon-wrap {
  width: 30px;
  height: 30px;
  border-radius: 6px;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid var(--card-border, rgba(255, 255, 255, 0.08));
  flex-shrink: 0;
  box-sizing: border-box;
  padding: 2px;
}

.table-img-icon {
  width: 100%;
  height: 100%;
  object-fit: contain;
  border-radius: 4px;
  display: block;
}

.table-fallback-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  color: var(--text-muted);
  opacity: 0.7;
}

.table-ext-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.table-ext-title-row {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.table-ext-title-row .ext-name {
  font-weight: 600;
  color: var(--text-main);
  letter-spacing: 0.2px;
}

.table-ext-title-row .ext-ver {
  font-size: 0.74rem;
  color: var(--text-muted);
}

.badge-active-pill.small {
  padding: 1px 6px;
  font-size: 0.68rem;
}

.table-publisher-cell {
  display: flex;
  flex-direction: column;
  gap: 3px;
  align-items: flex-start;
}

.publisher-ns {
  font-size: 0.78rem;
  color: var(--text-muted);
}

.table-desc-text {
  font-size: 0.78rem;
  color: var(--text-muted);
  max-width: 380px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  line-height: 1.35;
}

.table-stats-cell {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 0.75rem;
  color: var(--text-muted);
}

.stat-pill {
  display: inline-flex;
  align-items: center;
  padding: 2px 6px;
  border-radius: 4px;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  font-size: 0.74rem;
}

.table-actions-cell {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  justify-content: flex-end;
}

.btn-apply-theme.small,
.btn-install-theme.small,
.btn-applied-current.small {
  padding: 4px 10px;
  font-size: 0.75rem;
}


.plugin-tags-row.small {
  gap: 3px;
}

.plugin-tags-row.small .plugin-tag {
  font-size: 0.68rem;
  padding: 1px 5px;
}

/* Light Theme Overrides for Marketplace Table */
:root[data-theme="light"] .market-list-table-wrapper,
[data-theme="light"] .market-list-table-wrapper {
  background: #ffffff;
  border-color: #e2e8f0;
}

:root[data-theme="light"] .market-table thead tr,
[data-theme="light"] .market-table thead tr {
  background: #f8fafc;
  border-bottom-color: #e2e8f0;
}

:root[data-theme="light"] .market-table th,
[data-theme="light"] .market-table th {
  color: #64748b;
}

:root[data-theme="light"] .market-table-row,
[data-theme="light"] .market-table-row {
  border-bottom-color: #f1f5f9;
}

:root[data-theme="light"] .market-table-row:hover,
[data-theme="light"] .market-table-row:hover {
  background: rgba(2, 132, 199, 0.03);
}

:root[data-theme="light"] .stat-pill,
[data-theme="light"] .stat-pill {
  background: #f8fafc;
  border-color: #e2e8f0;
  color: #0f172a;
}

:root[data-theme="light"] .table-icon-wrap,
[data-theme="light"] .table-icon-wrap {
  background: #f1f5f9;
}

:root[data-theme="light"] .view-mode-toggle,
[data-theme="light"] .view-mode-toggle {
  background: #f8fafc;
  border-color: #cbd5e1;
}

:root[data-theme="light"] .view-mode-btn,
[data-theme="light"] .view-mode-btn {
  color: #64748b;
}

:root[data-theme="light"] .view-mode-btn:hover,
[data-theme="light"] .view-mode-btn:hover {
  background: rgba(0, 0, 0, 0.05);
  color: #0f172a;
}

:root[data-theme="light"] .view-mode-btn.active,
[data-theme="light"] .view-mode-btn.active {
  background: #ffffff;
  color: #0284c7;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}
</style>
