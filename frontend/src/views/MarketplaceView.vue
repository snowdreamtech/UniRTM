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
        <span class="error-icon">⚠️</span>
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
                v-if="activeDetail.files && activeDetail.files.icon"
                :src="activeDetail.files.icon"
                class="hero-img-icon"
                alt="icon"
                @error="(e: any) => { if (e && e.target) e.target.style.display = 'none'; }"
              />
              <span class="hero-fallback-icon">🎨</span>
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
                <span class="empty-icon">📄</span>
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
                      v-if="activeDetail.is_installed"
                      class="btn-apply-sub"
                      @click="onApplySubColorTheme(sub.label)"
                    >
                      应用
                    </button>
                  </div>
                </div>
              </div>

              <!-- Icon Themes List -->
              <div
                v-if="activeDetail.contributes?.iconThemes && activeDetail.contributes.iconThemes.length > 0"
                class="contribute-group"
              >
                <h3 class="group-title">图标集 (Icon Themes)</h3>
                <div class="contribute-list">
                  <div
                    v-for="sub in activeDetail.contributes.iconThemes"
                    :key="sub.label"
                    class="contribute-item"
                  >
                    <div class="sub-info">
                      <span class="sub-label">{{ sub.label }}</span>
                      <span class="sub-mode-tag mode-icon">文件图标</span>
                    </div>
                    <button
                      v-if="activeDetail.is_installed"
                      class="btn-apply-sub"
                      @click="onApplySubIconTheme(sub.label)"
                    >
                      应用
                    </button>
                  </div>
                </div>
              </div>

              <div
                v-if="(!activeDetail.contributes?.themes || activeDetail.contributes.themes.length === 0) &&
                      (!activeDetail.contributes?.iconThemes || activeDetail.contributes.iconThemes.length === 0)"
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
                  <span class="link-icon">🐙</span>
                  <span>项目仓库 (Repository)</span>
                </a>
                <a
                  v-if="activeDetail.bugs"
                  :href="activeDetail.bugs"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="sidebar-link"
                >
                  <span class="link-icon">🐛</span>
                  <span>问题反馈 (Issues)</span>
                </a>
                <a
                  v-if="activeDetail.homepage"
                  :href="activeDetail.homepage"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="sidebar-link"
                >
                  <span class="link-icon">🏠</span>
                  <span>主页 (Homepage)</span>
                </a>
                <a
                  v-if="activeDetail.files && activeDetail.files.download"
                  :href="activeDetail.files.download"
                  download
                  class="sidebar-link"
                >
                  <span class="link-icon">📥</span>
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
            <span v-else>{{ t('marketplace.search') || '搜索' }}</span>
          </button>
        </div>

        <!-- Quick Chips & Results count -->
        <div class="market-chips-bar">
          <div class="chips-left">
            <span class="chips-label">{{ t('marketplace.recommended_categories') || '推荐分类' }}:</span>
            <button
              v-for="chip in themeChips"
              :key="chip.id"
              class="chip-btn"
              :class="{
                active: selectedFilter === chip.id,
                'chip-installed': chip.isInstalled
              }"
              @click="onSelectChip(chip)"
            >
              <span v-if="chip.isInstalled">
                📦 {{ (chip.labelKey ? t(chip.labelKey as any) : chip.name) || '已安装' }} ({{ installedExtensions.length }})
              </span>
              <span v-else>
                {{ (chip.labelKey ? t(chip.labelKey as any) : chip.name) || chip.id }}
              </span>
            </button>
          </div>
          <span v-if="selectedFilter === 'installed'" class="results-count-label">
            {{ t('marketplace.installed_count', { count: displayedExtensions.length }) || `已安装 ${displayedExtensions.length} 款主题与图标包` }}
          </span>
          <span v-else-if="mpTotalSize > 0" class="results-count-label">
            {{ t('marketplace.total_count', { count: mpTotalSize }) || `共收录 ${mpTotalSize} 款主题与图标包` }}
          </span>
        </div>

        <!-- Feedback Alert -->
        <div v-if="feedbackMsg" class="feedback-banner" :class="{ error: feedbackIsError }">
          <span>{{ feedbackMsg }}</span>
          <button class="btn-close-banner" @click="feedbackMsg = ''">✕</button>
        </div>

        <!-- Error Banner -->
        <div v-if="mpError && selectedFilter !== 'installed'" class="error-banner">
          <span>⚠️ {{ mpError }}</span>
          <button class="btn-retry" @click="onSearchThemes(mpSearchQuery)">{{ t('marketplace.retry') || '重试' }}</button>
        </div>

        <!-- Loading State -->
        <div v-if="mpIsSearching && selectedFilter !== 'installed'" class="market-loading-box glass-card">
          <div class="spinner-small"></div>
          <span>{{ t('marketplace.searching_openvsx') || '正在实时连接 Open VSX 检索主题与图标...' }}</span>
        </div>

        <!-- Empty State -->
        <div v-else-if="displayedExtensions.length === 0" class="market-empty-box glass-card">
          <span class="empty-icon">🎨</span>
          <p>{{ selectedFilter === 'installed' ? (t('marketplace.empty_installed') || '暂未安装任何第三方主题或图标包') : (t('marketplace.empty_results') || '未找到符合条件的主题或图标包') }}</p>
          <button class="btn-reset-chip" @click="onResetToPopular">{{ t('marketplace.view_popular') || '查看热门推荐' }}</button>
        </div>

        <!-- Themes Grid -->
        <div v-else class="extensions-grid">
          <div
            v-for="item in displayedExtensions"
            :key="(item.namespace || '') + '.' + item.name"
            class="glass-card theme-card clickable-card"
            :class="{ 'card-is-active': isExtensionActive(item) }"
            @click="navigateToDetail(item.namespace || '', item.name)"
          >
            <!-- Top Right Delete Button for Installed Themes -->
            <button
              v-if="isItemInstalled(item)"
              class="btn-card-delete"
              :title="t('common.delete') || '卸载此主题包'"
              @click.stop="onUninstallThemeCard(item)"
            >
              <svg class="trash-icon" viewBox="0 0 24 24" width="13" height="13" stroke="currentColor" stroke-width="2" fill="none">
                <polyline points="3 6 5 6 21 6"></polyline>
                <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
                <line x1="10" y1="11" x2="10" y2="17"></line>
                <line x1="14" y1="11" x2="14" y2="17"></line>
              </svg>
            </button>

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
                  <span v-if="isExtensionActive(item)" class="badge-active-pill">
                    <span class="active-dot"></span>
                    <span>{{ t('marketplace.active') || '生效中' }}</span>
                  </span>
                </div>
                <div class="card-subtitle-row">
                  <span class="card-namespace font-mono">{{ item.namespace }}</span>
                  <span
                    class="card-type-tag"
                    :class="getExtensionType(item) === 'icon' ? 'type-icon' : 'type-color'"
                  >
                    {{ getExtensionType(item) === 'icon' ? ('📁 ' + (t('marketplace.filter_icons') || '图标')) : ('🎨 ' + (t('marketplace.filter_themes') || '主题')) }}
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
                    {{ t('marketplace.currently_applied') || '当前生效' }}
                  </button>
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
                <span v-if="ext.installed" class="badge-installed">✓ {{ t("settings.installed") || "已启用" }}</span>
                <span v-else class="badge-soon">开发中</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { t } from "../i18n";
import { useThemeMarketplace } from "../composables/useThemeMarketplace";
import { useThemeManager } from "../composables/useThemeManager";
import { useToast } from "../composables/useToast";
import { renderMarkdown } from "../utils/markdown";
import type { MarketplaceExtension, MarketplaceExtensionDetail } from "../types/theme-manager";

const route = useRoute();
const router = useRouter();
const { showToast } = useToast();

// Channels: 'themes' (Open VSX) vs 'plugins' (UniRTM Native Plugins)
const currentChannel = ref<"themes" | "plugins">("themes");

// Detail view state
const activeDetailTab = ref<"readme" | "contributes" | "changelog">("readme");
const isUninstalling = ref<boolean>(false);
const changelogContent = ref<string>("");

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
  activeColorTheme,
  activeIconTheme,
  fetchThemes,
  setActiveTheme,
  deleteTheme,
} = useThemeManager();

const selectedFilter = ref<string>("all");

interface ThemeChipItem {
  id: string;
  labelKey?: string;
  name?: string;
  query: string;
  isInstalled?: boolean;
}

const themeChips = computed<ThemeChipItem[]>(() => [
  { id: "all", labelKey: "marketplace.filter_all", query: "" },
  { id: "installed", labelKey: "marketplace.filter_installed", query: "", isInstalled: true },
  { id: "material", name: "Material", query: "Material" },
  { id: "dracula", name: "Dracula", query: "Dracula" },
  { id: "one-dark", name: "One Dark", query: "One Dark" },
  { id: "catppuccin", name: "Catppuccin", query: "Catppuccin" },
  { id: "tokyo-night", name: "Tokyo Night", query: "Tokyo Night" },
  { id: "github", name: "GitHub", query: "GitHub" },
]);

const installedExtensions = computed<MarketplaceExtension[]>(() => {
  const list: MarketplaceExtension[] = [];
  const seen = new Set<string>();

  for (const c of colorThemes.value) {
    if (!c.is_builtin && c.package_id) {
      const key = `color:${c.package_id.toLowerCase()}`;
      if (seen.has(key)) continue;
      seen.add(key);

      const match = mpSearchResults.value.find(
        (s) => (s.package_id && s.package_id.toLowerCase() === c.package_id?.toLowerCase()) ||
               s.name.toLowerCase() === c.package_id?.toLowerCase()
      );
      if (match) {
        list.push({ ...match, is_installed: true });
      } else {
        list.push({
          namespace: c.package_id.includes(".") ? c.package_id.split(".")[0] : "",
          name: c.package_id.includes(".") ? c.package_id.split(".")[1] : c.package_id,
          displayName: c.name || c.package_id,
          version: c.version || "",
          description: "已安装的本地 VS Code 颜色主题扩展包",
          downloadCount: 0,
          averageRating: 5.0,
          is_installed: true,
          package_id: c.package_id,
          files: {
            download: "",
            icon: "",
          },
        });
      }
    }
  }

  for (const i of iconThemes.value) {
    if (!i.is_builtin && i.package_id) {
      const key = `icon:${i.package_id.toLowerCase()}`;
      if (seen.has(key)) continue;
      seen.add(key);

      const match = mpSearchResults.value.find(
        (s) => (s.package_id && s.package_id.toLowerCase() === i.package_id?.toLowerCase()) ||
               s.name.toLowerCase() === i.package_id?.toLowerCase()
      );
      if (match) {
        list.push({ ...match, is_installed: true });
      } else {
        list.push({
          namespace: i.package_id.includes(".") ? i.package_id.split(".")[0] : "",
          name: i.package_id.includes(".") ? i.package_id.split(".")[1] : i.package_id,
          displayName: i.name || i.package_id,
          version: i.version || "",
          description: "已安装的本地 VS Code 图标主题扩展包",
          downloadCount: 0,
          averageRating: 5.0,
          is_installed: true,
          package_id: i.package_id,
          files: {
            download: "",
            icon: "",
          },
        });
      }
    }
  }

  return list;
});

const displayedExtensions = computed<MarketplaceExtension[]>(() => {
  if (selectedFilter.value === "installed") {
    if (!mpSearchQuery.value.trim()) {
      return installedExtensions.value;
    }
    const q = mpSearchQuery.value.toLowerCase().trim();
    return installedExtensions.value.filter(
      (item) =>
        (item.displayName && item.displayName.toLowerCase().includes(q)) ||
        item.name.toLowerCase().includes(q) ||
        (item.description && item.description.toLowerCase().includes(q)) ||
        (item.package_id && item.package_id.toLowerCase().includes(q))
    );
  }
  return mpSearchResults.value;
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
  return tCount + iCount;
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

// Initialize channel and load installed themes on mount
onMounted(async () => {
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

function getExtensionType(item: MarketplaceExtension): "icon" | "theme" {
  // Check local installed themes first for 100% exact type match
  const pkgId = (item.package_id || item.name || "").toLowerCase();
  const name = (item.name || "").toLowerCase();
  if (iconThemes.value.some((i) => !i.is_builtin && ((i.package_id && i.package_id.toLowerCase() === pkgId) || i.id.toLowerCase() === name))) {
    return "icon";
  }
  if (colorThemes.value.some((c) => !c.is_builtin && ((c.package_id && c.package_id.toLowerCase() === pkgId) || c.id.toLowerCase() === name))) {
    return "theme";
  }

  // Name or displayName with 'icon' or 'icons' is an icon pack
  const nameText = `${item.name} ${item.displayName || ""}`.toLowerCase();
  if (nameText.includes("icon") || nameText.includes("icons")) {
    return "icon";
  }

  // Fallback to description
  const desc = (item.description || "").toLowerCase();
  if (desc.includes("icon theme") || desc.includes("icons theme") || desc.includes("file icons")) {
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
    return cPkg === pkgId || cPkg === name || cPkg === fullId || cId === pkgId || cId === name || (displayName && cName === displayName);
  });
  if (isColor) return true;

  const isIcon = iconThemes.value.some((i) => {
    if (i.is_builtin) return false;
    const iPkg = (i.package_id || "").toLowerCase();
    const iId = i.id.toLowerCase();
    const iName = i.name.toLowerCase();
    return iPkg === pkgId || iPkg === name || iPkg === fullId || iId === pkgId || iId === name || (displayName && iName === displayName);
  });
  return isIcon;
}

function isExtensionActive(item: { namespace?: string; name: string; displayName?: string; package_id?: string; is_installed?: boolean }): boolean {
  if (!isItemInstalled(item)) return false;

  const pkgId = (item.package_id || item.name || "").toLowerCase();
  const name = (item.name || "").toLowerCase();
  const ns = (item.namespace || "").toLowerCase();
  const fullId = `${ns}.${name}`;
  const displayName = (item.displayName || "").toLowerCase();

  // 1. Check against active color theme
  if (activeColorTheme.value) {
    const actColor = activeColorTheme.value.toLowerCase();
    if (actColor === pkgId || actColor === name || actColor === fullId || actColor === displayName) return true;
    const activeColorItem = colorThemes.value.find((c) => c.id.toLowerCase() === actColor);
    if (activeColorItem) {
      const cPkg = (activeColorItem.package_id || "").toLowerCase();
      const cId = activeColorItem.id.toLowerCase();
      const cName = activeColorItem.name.toLowerCase();
      if (cPkg && (cPkg === pkgId || cPkg === name || cPkg === fullId)) return true;
      if (cId === pkgId || cId === name || cId === fullId) return true;
      if (cName === displayName) return true;
    }
  }

  // 2. Check against active icon theme
  if (activeIconTheme.value) {
    const actIcon = activeIconTheme.value.toLowerCase();
    if (actIcon === pkgId || actIcon === name || actIcon === fullId || actIcon === displayName) return true;
    const activeIconItem = iconThemes.value.find((i) => i.id.toLowerCase() === actIcon);
    if (activeIconItem) {
      const iPkg = (activeIconItem.package_id || "").toLowerCase();
      const iId = activeIconItem.id.toLowerCase();
      const iName = activeIconItem.name.toLowerCase();
      if (iPkg && (iPkg === pkgId || iPkg === name || iPkg === fullId)) return true;
      if (iId === pkgId || iId === name || iId === fullId) return true;
      if (iName === displayName) return true;
    }
  }

  return false;
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
  if (selectedFilter.value === "installed") {
    return;
  }
  const query = q !== undefined ? q : mpSearchQuery.value;
  const match = themeChips.value.find((c) => !c.isInstalled && c.query.toLowerCase() === query.toLowerCase().trim());
  selectedFilter.value = match ? match.id : (query ? "" : "all");
  mpSearchThemes(query);
}

function onSelectChip(chip: ThemeChipItem) {
  selectedFilter.value = chip.id;
  if (chip.isInstalled) {
    mpSearchQuery.value = "";
    return;
  }
  mpSearchQuery.value = chip.query;
  mpSearchThemes(chip.query);
}

function onResetToPopular() {
  const allChip = themeChips.value[0];
  if (allChip) {
    onSelectChip(allChip);
  }
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
    feedbackMsg.value = `✓ 已应用颜色主题: ${match.name}`;
    feedbackIsError.value = false;
    showToast(`🎨 已应用颜色主题: ${match.name}`, "success");
  } else {
    feedbackMsg.value = `已激活主题包，可在设置中选择应用 ${label}`;
    feedbackIsError.value = false;
    showToast(`🎨 已激活主题包: ${label}`, "info");
  }
}

async function onApplySubIconTheme(label: string) {
  const match = iconThemes.value.find(
    (i) => i.name.toLowerCase() === label.toLowerCase() || i.id.toLowerCase() === label.toLowerCase()
  );
  if (match) {
    await setActiveTheme(undefined, match.id);
    feedbackMsg.value = `✓ 已应用图标主题: ${match.name}`;
    feedbackIsError.value = false;
    showToast(`📁 已应用图标主题: ${match.name}`, "success");
  } else {
    feedbackMsg.value = `已激活图标包，可在设置中选择应用 ${label}`;
    feedbackIsError.value = false;
    showToast(`📁 已激活图标包: ${label}`, "info");
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

    let applied = false;
    if (colorMatch) {
      await setActiveTheme(colorMatch.id, undefined);
      applied = true;
    }
    if (iconMatch) {
      await setActiveTheme(undefined, iconMatch.id);
      applied = true;
    }

    if (!applied) {
      // Fallback: apply directly by extension name according to type
      if (getExtensionType(item) === "icon") {
        await setActiveTheme(undefined, item.name);
      } else {
        await setActiveTheme(item.name, undefined);
      }
    }

    // Refresh themes so all UI states sync
    await fetchThemes();

    const targetName = item.displayName || item.name;
    feedbackMsg.value = `✓ 已启用主题: ${targetName}`;
    feedbackIsError.value = false;
    showToast(`✓ 已启用并应用主题: ${targetName}`, "success");
  } catch (err: any) {
    feedbackMsg.value = `✕ 激活主题失败: ${err.message || err}`;
    feedbackIsError.value = true;
    showToast(`✕ 激活主题失败: ${err.message || err}`, "error");
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
  padding-right: 28px;
}

.btn-card-delete {
  position: absolute;
  top: 10px;
  right: 10px;
  width: 26px;
  height: 26px;
  border-radius: 6px;
  background: rgba(239, 68, 68, 0.12);
  border: 1px solid rgba(239, 68, 68, 0.28);
  color: #ef4444;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s ease;
  z-index: 5;
}

.btn-card-delete:hover {
  background: #ef4444;
  border-color: #ef4444;
  color: #ffffff;
  transform: scale(1.1);
  box-shadow: 0 2px 10px rgba(239, 68, 68, 0.4);
}

.trash-icon {
  stroke: currentColor;
  stroke-width: 2;
  transition: transform 0.2s ease;
}

.btn-card-delete:hover .trash-icon {
  transform: scale(1.1);
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

/* Light Mode Overrides for Active Badges & Buttons */
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
</style>
