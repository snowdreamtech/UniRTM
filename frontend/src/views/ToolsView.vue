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
      <div class="view-actions" style="display: flex; gap: 8px; align-items: center;">
        <button
          class="tool-action-btn secondary"
          id="btn-check-updates"
          @click="checkRemoteUpdates"
          :disabled="checkingUpdates || loading"
          title="检测所有工具最新远程版本与升级差异"
        >
          <TechIcon id="workbench:sync" :size="14" :class="{ 'spin-icon': checkingUpdates }" />
          <span>{{ checkingUpdates ? "正在检测更新..." : "检测更新" }}</span>
        </button>
        <button class="tool-action-btn secondary" @click="fetchStatus" :disabled="loading">
          <TechIcon id="workbench:sync" :size="14" :class="{ 'spin-icon': loading }" />
          <span>{{ loading ? t("common.refreshing") : t("common.refresh") }}</span>
        </button>
      </div>
    </div>

    <!-- Current Scope Context Banner -->
    <div class="scope-selection-bar glass-card" id="runtimes-scope-bar">
      <div class="scope-context-pill">
        <span class="scope-badge" :class="currentScope">
          <TechIcon :id="currentScope === 'user' ? 'workbench:settings' : 'workbench:folder'" :size="13" :colored="false" />
          {{ currentScope === 'user' ? (t("runtimes.scope_user") || "用户全局环境") : (t("runtimes.scope_workspace") || "工作区项目环境") }}
        </span>
        <span class="scope-hint-text">
          {{ currentScope === 'user' ? (t("runtimes.scope_user_desc") || "配置当前用户所有未指定工作区时的默认兜底环境 (~/.unirtm)") : (t("runtimes.scope_workspace_desc") || "锁定当前工作区工具版本，团队协作检出即生效 (.unirtm.toml)") }}
        </span>
      </div>
      <div class="scope-context-config font-mono" :title="currentScope === 'user' ? '~/.unirtm' : '.unirtm.toml'">
        <span class="file-icon">📄</span>
        <span>{{ currentScope === 'user' ? '~/.unirtm' : '.unirtm.toml' }}</span>
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

        <!-- 待安装统计徽标：遵循低噪音原则，仅在 > 0 时醒目呈现，支持点击一键切换筛选 -->
        <span
          v-if="scopeMissingCount > 0"
          class="stat-badge stat-badge-missing clickable"
          :class="{ active: filterOnlyMissing }"
          :title="t('runtimes.missing_tooltip', { count: scopeMissingCount }) || `当前作用域有 ${scopeMissingCount} 个工具配置了但本地尚未安装，点击快速筛选`"
          @click="filterOnlyMissing = !filterOnlyMissing"
        >
          <span class="badge-dot warning-dot">●</span>
          <b>{{ scopeMissingCount }}</b> {{ t("runtimes.count_missing_suffix") || "个待安装" }}
        </span>

        <!-- 可升级统计徽标：绿色专属升级状态 -->
        <span
          class="stat-badge"
          :class="{ highlight: scopeUpgradableCount > 0, clickable: scopeUpgradableCount > 0 }"
          :title="scopeUpgradableCount > 0 ? (t('runtimes.upgradable_tooltip', { count: scopeUpgradableCount }) || `当前作用域有 ${scopeUpgradableCount} 个工具可升级，点击查看详情`) : ''"
          @click="scopeUpgradableCount > 0 ? (isDiffDrawerOpen = true) : undefined"
        >
          <b>{{ scopeUpgradableCount }}</b> {{ t("runtimes.count_upgradable_suffix") || "个可升级" }}
        </span>

        <!-- View Mode Switcher -->
        <div class="view-mode-toggle" id="tools-view-mode-toggle">
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

        <button
          class="tool-action-btn primary small"
          id="btn-add-runtime"
          @click="openAddModal"
        >
          <span>+ {{ t("runtimes.add_runtime") || "添加运行环境" }}</span>
        </button>
      </div>
    </div>

    <!-- Empty State When Scope Has No Configured Tools -->
    <div v-if="filteredRuntimes.length === 0" class="glass-card runtimes-empty-box" id="runtimes-empty-state">
      <div class="empty-icon-wrap">
        <TechIcon id="workbench:tools" :size="32" :colored="false" />
      </div>
      <div class="empty-title">
        {{ currentScope === 'user' ? '用户默认环境暂未配置工具' : '当前工作区暂未配置工具' }}
      </div>
      <p class="empty-desc">
        {{ currentScope === 'user'
            ? '用户区配置文件 (~/.config/unirtm/unirtm.toml) 提供未绑定工作区时的全局兜底环境，可通过上方「+ 添加运行环境」进行配置。'
            : '工作区配置文件 (.unirtm.toml) 暂无匹配工具，可通过上方「+ 添加运行环境」添加。' }}
      </p>
      <button class="tool-action-btn primary small" @click="openAddModal">
        <span>+ {{ t("runtimes.add_runtime") || "添加运行环境" }}</span>
      </button>
    </div>

    <!-- Runtimes Grid Cards View -->
    <div v-else-if="viewMode === 'cards'" class="runtimes-grid">
      <div
        v-for="rt in filteredRuntimes"
        :key="rt.id"
        class="glass-card runtime-card"
        :class="{ active: currentScope === 'workspace' ? rt.active_version : rt.global_version }"
        :id="'runtime-card-' + rt.id"
      >
        <!-- Card Header -->
        <div class="runtime-card-header">
          <div class="runtime-icon-badge"><TechIcon :id="rt.id" :size="36" :colored="true" /></div>
          <div class="runtime-title-block">
            <div class="runtime-name-row">
              <a
                :href="getToolHomepage(rt)"
                target="_blank"
                rel="noopener noreferrer"
                class="runtime-name"
                :title="`${formatToolTitle(rt.name)} (访问官方主页/代码仓库)`"
              >
                {{ formatToolTitle(rt.name) }}
              </a>
              <div class="runtime-header-badges">
                <!-- 1. 未配置态 -->
                <span
                  v-if="getToolStatus(rt).state === 'unconfigured'"
                  class="runtime-version-badge is-muted font-mono"
                  :title="t('runtimes.unconfigured')"
                >
                  {{ t('runtimes.unconfigured') }}
                </span>

                <!-- 2. 已配置但未安装 (Missing) -> 琥珀警示色 -->
                <span
                  v-else-if="getToolStatus(rt).state === 'missing'"
                  class="runtime-version-badge is-missing font-mono"
                  :title="`工作区配置了 ${formatVersion(getToolStatus(rt).configuredVersion)}，但本地磁盘尚未安装`"
                >
                  <span class="missing-dot">●</span>
                  <span>{{ formatVersion(getToolStatus(rt).configuredVersion) }} ({{ t('runtimes.status_missing') || '待安装' }})</span>
                </span>

                <!-- 3. 已安装且已就绪 (Ready) -> 沉稳中性低噪音，不展示任何图标 -->
                <span
                  v-else
                  class="runtime-version-badge is-ready font-mono"
                  :title="formatVersion(getToolStatus(rt).configuredVersion)"
                >
                  {{ formatVersion(getToolStatus(rt).configuredVersion) }}
                </span>

                <!-- 4. 专属绿色升级徽标 (Upgradable) -> 绿色留给可升级状态 -->
                <button
                  v-if="getToolStatus(rt).state === 'upgradable' && getToolStatus(rt).upgradableVersion"
                  type="button"
                  class="runtime-upgrade-pill font-mono"
                  :title="`可升级至最新版本 ${formatVersion(getToolStatus(rt).upgradableVersion!)}`"
                  @click.stop="openRemoteModal(rt)"
                >
                  <span class="upgrade-arrow">↑</span>
                  <span>{{ formatVersion(getToolStatus(rt).upgradableVersion!) }}</span>
                </button>
              </div>
            </div>
            <span class="runtime-desc">{{ rt.description }}</span>
            <div v-if="rt.pinned && rt.pin_reason" class="pin-reason-tag" :title="rt.pin_reason">
              <span class="pin-reason-icon">📌</span>
              <span class="pin-reason-text">{{ rt.pin_reason }}</span>
            </div>
          </div>
        </div>

        <!-- Card Actions: Primary [切换] [安装] + More Menu [···] (锁定, 删除) -->
        <div class="runtime-card-footer">
          <button
            class="card-action-btn action-switch"
            :id="'btn-switch-' + rt.id"
            @click.stop="openSwitchVersionModal(rt)"
            :title="t('runtimes.action_switch')"
          >
            {{ t("runtimes.action_switch") }}
          </button>
          <button
            class="card-action-btn action-install"
            :id="'btn-install-' + rt.id"
            @click.stop="openRemoteModal(rt)"
            :title="t('runtimes.action_install')"
          >
            {{ t("runtimes.action_install") }}
          </button>
          <div class="card-more-action-wrap">
            <button
              class="card-action-btn action-more"
              :class="{ 'is-active': activeMenuRuntimeId === rt.id }"
              :id="'btn-more-' + rt.id"
              @click.stop="toggleRuntimeMenu(rt.id)"
              :title="t('common.more_actions') || '更多操作'"
            >
              <TechIcon id="workbench:more" :size="15" />
            </button>
            <transition name="dropdown-fade">
              <div v-if="activeMenuRuntimeId === rt.id" class="runtime-dropdown-menu" @click.stop>
                <button
                  class="dropdown-item"
                  :class="{ 'is-pinned': rt.pinned }"
                  :id="'btn-pin-' + rt.id"
                  @click.stop="handleTogglePinFromMenu(rt)"
                  :title="rt.pinned ? (rt.pin_reason ? `${t('runtimes.action_unlock')}：${rt.pin_reason}` : t('runtimes.action_unlock')) : t('runtimes.action_lock')"
                >
                  <TechIcon :id="rt.pinned ? 'workbench:unlock' : 'workbench:lock'" :size="13" />
                  <span>{{ rt.pinned ? (t("runtimes.action_unlock") || "解除锁定") : (t("runtimes.action_lock") || "锁定版本") }}</span>
                </button>
                <div class="dropdown-divider"></div>
                <button
                  class="dropdown-item danger"
                  :id="'btn-delete-' + rt.id"
                  @click.stop="handleRemoveFromMenu(rt)"
                  :title="t('runtimes.action_delete')"
                >
                  <TechIcon id="workbench:delete" :size="13" />
                  <span>{{ t("runtimes.action_delete") || "移除工具" }}</span>
                </button>
              </div>
            </transition>
          </div>
        </div>
      </div>
    </div>

    <!-- Runtimes High-Density List / Table View -->
    <div v-else class="tools-list-table-wrapper glass-card">
      <table class="tools-table">
        <thead>
          <tr>
            <th class="th-tool">{{ t("runtimes.col_tool") || "工具" }}</th>
            <th class="th-version">{{ t("runtimes.col_version") || "版本" }}</th>
            <th class="th-desc">{{ t("runtimes.col_desc") || "描述" }}</th>
            <th class="th-actions">{{ t("runtimes.col_actions") || "操作" }}</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="rt in filteredRuntimes"
            :key="rt.id"
            class="tool-table-row"
            :class="{ active: currentScope === 'workspace' ? rt.active_version : rt.global_version }"
            :id="'tool-row-' + rt.id"
          >
            <!-- Col: Tool -->
            <td class="td-tool">
              <div class="table-tool-cell">
                <TechIcon :id="rt.id" :size="24" :colored="true" class="tool-icon" />
                <a
                  :href="getToolHomepage(rt)"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="tool-title font-mono"
                  :title="`${formatToolTitle(rt.name)} (访问官方主页/代码仓库)`"
                >
                  {{ formatToolTitle(rt.name) }}
                </a>
              </div>
            </td>

            <!-- Col: Version -->
            <td class="td-version">
              <div class="table-version-cell">
                <span
                  v-if="getToolStatus(rt).state === 'unconfigured'"
                  class="runtime-version-badge is-muted font-mono"
                  :title="t('runtimes.unconfigured')"
                >
                  {{ t('runtimes.unconfigured') }}
                </span>
                <span
                  v-else-if="getToolStatus(rt).state === 'missing'"
                  class="runtime-version-badge is-missing font-mono"
                  :title="`配置了 ${formatVersion(getToolStatus(rt).configuredVersion)}，但本地尚未下载安装`"
                >
                  <span class="missing-dot">●</span>
                  <span>{{ formatVersion(getToolStatus(rt).configuredVersion) }} ({{ t('runtimes.status_missing') || '待安装' }})</span>
                </span>
                <span
                  v-else
                  class="runtime-version-badge is-ready font-mono"
                  :title="formatVersion(getToolStatus(rt).configuredVersion)"
                >
                  {{ formatVersion(getToolStatus(rt).configuredVersion) }}
                </span>
                <button
                  v-if="getToolStatus(rt).state === 'upgradable' && getToolStatus(rt).upgradableVersion"
                  type="button"
                  class="runtime-upgrade-pill font-mono"
                  :title="`可升级至 ${formatVersion(getToolStatus(rt).upgradableVersion!)}`"
                  @click.stop="openRemoteModal(rt)"
                >
                  <span class="upgrade-arrow">↑</span>
                  <span>{{ formatVersion(getToolStatus(rt).upgradableVersion!) }}</span>
                </button>
                <span v-if="rt.pinned" class="pin-indicator" :title="rt.pin_reason || t('runtimes.action_lock')">📌</span>
              </div>
            </td>

            <!-- Col: Description -->
            <td class="td-desc">
              <div class="table-desc-cell">
                <span class="tool-desc-text" :title="rt.description">{{ rt.description }}</span>
                <span v-if="rt.pinned && rt.pin_reason" class="pin-reason-text-sm" :title="rt.pin_reason">
                  📌 {{ rt.pin_reason }}
                </span>
              </div>
            </td>

            <!-- Col: Actions (Switch, Install + More Menu) -->
            <td class="td-actions">
              <div class="table-actions-group">
                <button
                  class="row-action-btn action-switch"
                  :id="'btn-switch-' + rt.id"
                  @click.stop="openSwitchVersionModal(rt)"
                  :title="t('runtimes.action_switch')"
                >
                  {{ t("runtimes.action_switch") }}
                </button>
                <button
                  class="row-action-btn action-install"
                  :id="'btn-install-' + rt.id"
                  @click.stop="openRemoteModal(rt)"
                  :title="t('runtimes.action_install')"
                >
                  {{ t("runtimes.action_install") }}
                </button>
                <div class="table-more-action-wrap">
                  <button
                    class="row-action-btn action-more"
                    :class="{ 'is-active': activeMenuRuntimeId === rt.id }"
                    :id="'btn-more-row-' + rt.id"
                    @click.stop="toggleRuntimeMenu(rt.id)"
                    :title="t('common.more_actions') || '更多操作'"
                  >
                    <TechIcon id="workbench:more" :size="14" />
                  </button>
                  <transition name="dropdown-fade">
                    <div v-if="activeMenuRuntimeId === rt.id" class="runtime-dropdown-menu table-menu" @click.stop>
                      <button
                        class="dropdown-item"
                        :class="{ 'is-pinned': rt.pinned }"
                        :id="'btn-pin-row-' + rt.id"
                        @click.stop="handleTogglePinFromMenu(rt)"
                        :title="rt.pinned ? (rt.pin_reason ? `${t('runtimes.action_unlock')}：${rt.pin_reason}` : t('runtimes.action_unlock')) : t('runtimes.action_lock')"
                      >
                        <TechIcon :id="rt.pinned ? 'workbench:unlock' : 'workbench:lock'" :size="13" />
                        <span>{{ rt.pinned ? (t("runtimes.action_unlock") || "解除锁定") : (t("runtimes.action_lock") || "锁定版本") }}</span>
                      </button>
                      <div class="dropdown-divider"></div>
                      <button
                        class="dropdown-item danger"
                        :id="'btn-delete-row-' + rt.id"
                        @click.stop="handleRemoveFromMenu(rt)"
                        :title="t('runtimes.action_delete')"
                      >
                        <TechIcon id="workbench:delete" :size="13" />
                        <span>{{ t("runtimes.action_delete") || "移除工具" }}</span>
                      </button>
                    </div>
                  </transition>
                </div>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>


    <!-- Switch Version Modal -->
    <div
      v-if="isSwitchModalOpen && switchModalRuntime"
      class="modal-backdrop switch-modal-backdrop"
      @click.self="closeSwitchVersionModal"
      @keydown.esc="closeSwitchVersionModal"
    >
      <div class="modal-dialog glass-card switch-modal-dialog" role="dialog" aria-modal="true">
        <div class="modal-header">
          <div class="modal-title-row">
            <span class="modal-icon">
              <TechIcon :id="switchModalRuntime.id" :size="24" :colored="true" />
            </span>
            <div>
              <h3 class="modal-title">
                {{ formatToolTitle(switchModalRuntime.name) }} · {{ t("runtimes.switch_version") }}
              </h3>
              <p class="modal-desc">
                {{ currentScope === 'user' ? t('runtimes.switch_scope_user_tip') : t('runtimes.switch_scope_workspace_tip') }}
              </p>
            </div>
          </div>
          <button class="btn btn-icon modal-close-btn" @click="closeSwitchVersionModal" :title="t('runtimes.close')">
            <TechIcon id="workbench:close" :size="14" />
          </button>
        </div>

        <div class="switch-modal-body">
          <div
            v-if="switchModalRuntime.installed_versions && switchModalRuntime.installed_versions.length > 0"
            class="switch-pills-container"
          >
            <button
              v-for="ver in switchModalRuntime.installed_versions"
              :key="ver"
              class="switch-ver-pill font-mono"
              :class="{
                'is-active': ver === getCurrentVersion(switchModalRuntime),
                'is-inactive': ver !== getCurrentVersion(switchModalRuntime),
              }"
              :title="ver === getCurrentVersion(switchModalRuntime) ? t('runtimes.active_badge') : `${t('runtimes.action_switch')}: ${formatVersion(ver)}`"
              @click="handleSelectVersionAndClose(switchModalRuntime, ver)"
            >
              <span v-if="ver === getCurrentVersion(switchModalRuntime)" class="pill-active-dot">●</span>
              <span class="pill-ver-text">{{ formatVersion(ver) }}</span>
              <span v-if="ver === getCurrentVersion(switchModalRuntime)" class="pill-active-tag">{{ t("runtimes.active_badge") }}</span>
            </button>
          </div>

          <div v-else class="switch-empty-box">
            <p class="switch-empty-tip">{{ t("runtimes.no_installed_versions") }}</p>
            <button class="switch-install-now-btn" @click="openInstallFromSwitch(switchModalRuntime)">
              + {{ t("runtimes.action_install") }}
            </button>
          </div>
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
              <div class="modal-title-text-wrap">
                <h3 class="modal-title">
                  {{ selectedRuntime?.name }} {{ t("runtimes.install_version") || "Install Version" }}
                </h3>
                <span
                  v-if="selectedRuntime?.pinned"
                  class="source-tag source-pinned"
                  :title="selectedRuntime.pin_reason ? `版本锁定: ${selectedRuntime.pin_reason}` : '版本已锁定'"
                >
                  🔒 已锁定{{ selectedRuntime.pin_reason ? `: ${selectedRuntime.pin_reason}` : "" }}
                </span>
              </div>
              <p class="modal-desc">
                {{ t("runtimes.remote_list_tip") || "Available official upstream releases" }}
              </p>
            </div>
          </div>
          <button class="btn btn-icon modal-close-btn" @click="closeRemoteModal" :title="t('runtimes.close') || '关闭'">
            <TechIcon id="workbench:close" :size="14" />
          </button>
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
        <div v-if="quickFilterTags.length > 0 || (selectedRuntime?.installed_versions && selectedRuntime.installed_versions.length > 0)" class="quick-filter-chips">
          <button
            class="filter-chip"
            :class="{ active: selectedMajorFilter === '' }"
            @click="selectedMajorFilter = ''"
          >
            {{ t('runtimes.all_versions') || '全部' }}
          </button>
          <button
            v-if="selectedRuntime?.installed_versions && selectedRuntime.installed_versions.length > 0"
            class="filter-chip installed-chip"
            :class="{ active: selectedMajorFilter === 'installed' }"
            @click="selectedMajorFilter = selectedMajorFilter === 'installed' ? '' : 'installed'"
            title="只查看本地共享安装池中已有的版本"
          >
            <span class="chip-installed-dot">●</span>
            <span>已安装 ({{ selectedRuntime.installed_versions.length }})</span>
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
                <span class="font-mono ver-name">{{ formatVersion(cleanCustomVersion) }}</span>
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
              <span class="font-mono ver-name">{{ formatVersion(rv.version) }}</span>
              <span
                v-if="isAlreadyInstalled(selectedRuntime?.id || '', rv.version)"
                class="installed-tag"
                title="已保存在本地共享安装池中，切换无需下载"
              >
                已安装
              </span>
            </div>

            <button
              class="install-action-btn"
              :class="{
                'is-active-ver': isActiveInModalScope(selectedRuntime?.id || '', rv.version),
                'is-installed': isAlreadyInstalled(selectedRuntime?.id || '', rv.version)
              }"
              :disabled="installingVer === rv.version || isActiveInModalScope(selectedRuntime?.id || '', rv.version)"
              @click="installRemoteVersion(selectedRuntime?.id || '', rv.version)"
            >
              <span v-if="installingVer === rv.version">{{ t("runtimes.installing") || "Installing..." }}</span>
              <span v-else-if="isActiveInModalScope(selectedRuntime?.id || '', rv.version)">
                当前使用
              </span>
              <span v-else-if="isAlreadyInstalled(selectedRuntime?.id || '', rv.version)">
                {{ modalInstallScope === 'user' ? (t("runtimes.set_user") || '设为用户') : (t("runtimes.set_workspace") || '设为工作区') }}
              </span>
              <span v-else>{{ t("runtimes.install_apply") || "安装并应用" }}</span>
            </button>
          </div>
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
          <button class="btn btn-icon modal-close-btn" @click="isAddModalOpen = false" :title="t('runtimes.close') || '关闭'">
            <TechIcon id="workbench:close" :size="14" />
          </button>
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
                    <span class="preset-tag font-mono">{{ formatVersion(preset.defaultVersion) }}</span>
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

    <!-- Version Diff & Batch Update Drawer/Modal -->
    <div
      v-if="isDiffDrawerOpen"
      class="modal-backdrop diff-drawer-backdrop"
      @click.self="isDiffDrawerOpen = false"
    >
      <div class="diff-drawer glass-card" role="dialog" aria-modal="true" id="diff-drawer-container">
        <!-- Header -->
        <div class="diff-drawer-header">
          <div>
            <h3 class="diff-drawer-title">可升级工具</h3>
            <p class="diff-drawer-desc">
              {{ filteredUpdateResults.length > 0
                ? `发现 ${filteredUpdateResults.length} 个工具可升级至最新版本，请勾选后确认应用。`
                : "当前所有工具均已是最新版本。" }}
            </p>
          </div>
          <button class="btn btn-icon modal-close-btn" @click="isDiffDrawerOpen = false" :title="t('common.close') || '关闭'">
            <TechIcon id="workbench:close" :size="14" />
          </button>
        </div>

        <!-- Select All Bar -->
        <div v-if="filteredUpdateResults.length > 0" class="diff-select-all-bar">
          <label class="diff-select-all-label">
            <input
              type="checkbox"
              :checked="isAllUpdatableSelected"
              @change="toggleSelectAllUpdatable"
            />
            <span>全选所有可升级工具 ({{ filteredUpdateResults.length }})</span>
          </label>
        </div>

        <!-- Body / Items List -->
        <div class="diff-drawer-body">
          <div v-if="filteredUpdateResults.length === 0" class="empty-diff-state">
            <span class="empty-diff-icon">✓</span>
            <p class="empty-diff-title">所有工具均已是最新版本</p>
            <p class="empty-diff-sub">没有需要升级的工具</p>
          </div>

          <div v-else class="diff-items-list">
            <label
              v-for="item in filteredUpdateResults"
              :key="item.id"
              class="diff-item-row"
              :class="{ 'is-selected': selectedUpdateTools.includes(item.id) }"
            >
              <div class="diff-item-left">
                <input
                  type="checkbox"
                  :value="item.id"
                  v-model="selectedUpdateTools"
                />
                <TechIcon :id="item.id" :size="24" :colored="true" />
                <div class="diff-tool-name-col">
                  <span class="diff-tool-title font-mono">{{ formatToolTitle(item.name || item.id) }}</span>
                  <span
                    v-if="item.id !== item.name && !item.name.toLowerCase().includes(item.id.toLowerCase())"
                    class="diff-tool-sub font-mono"
                  >
                    {{ item.id }}
                  </span>
                </div>
              </div>
              <div class="diff-version-transition">
                <span class="diff-ver-old font-mono" :title="`当前版本: ${formatVersion(item.current_version)}`">
                  {{ formatVersion(item.current_version) || '未指定' }}
                </span>
                <span class="diff-ver-arrow">→</span>
                <span class="diff-ver-new font-mono" :title="`可升级至最新: ${formatVersion(item.latest_version)}`">
                  {{ formatVersion(item.latest_version) }}
                </span>
              </div>
            </label>
          </div>
        </div>

        <!-- Footer -->
        <div class="diff-drawer-footer">
          <div class="diff-footer-info">
            <span class="diff-selection-summary">
              已选中 <b>{{ selectedUpdateTools.length }}</b> 个待升级工具
            </span>
          </div>
          <div class="diff-footer-actions">
            <button
              class="tool-action-btn secondary"
              @click="isDiffDrawerOpen = false"
              :disabled="batchUpdating"
            >
              取消
            </button>
            <button
              class="tool-action-btn primary"
              id="btn-apply-batch-update"
              @click="runBatchUpdate"
              :disabled="selectedUpdateTools.length === 0 || batchUpdating"
            >
              <TechIcon v-if="batchUpdating" id="workbench:sync" :size="14" class="spin-icon" />
              <span>{{ batchUpdating ? "正在批量升级..." : `一键升级选中项 (${selectedUpdateTools.length})` }}</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>

</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount } from "vue";
import { t } from "../i18n";
import { useToast } from "../composables/useToast";
import TechIcon from "../components/common/TechIcon.vue";
import { validateRuntimeVersion } from "../utils/versionValidator";
import { useToolUpdates, type ToolUpdateItem } from "../composables/useToolUpdates";
import { getToolHomepage } from "../utils/toolHomepage";
import { computeToolStatus, type ToolStatusResult } from "../utils/toolStatus";

import { useToolScope } from "../composables/useToolScope";

const { showToast } = useToast();
const { activeScope: currentScope } = useToolScope();
const {
  updateResults,
  isChecking: checkingUpdates,
  checkUpdates,
  markToolUpdated,
  syncFromToolsList,
} = useToolUpdates();

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
    source: ".unirtm.toml",
    is_legacy: false,
    legacy_file: "",
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
];

const runtimes = ref<RuntimeItem[]>([...defaultRuntimes]);
const totalInstalled = ref(0);
const detectedLegacies = ref<string[]>([]);
const loading = ref(false);
const migrating = ref(false);
const searchKey = ref("");

// View mode: 'cards' or 'list'
const viewMode = ref<"cards" | "list">(
  (localStorage.getItem("unirtm_view_mode_tools") as "cards" | "list") || "cards"
);
function setViewMode(mode: "cards" | "list") {
  viewMode.value = mode;
  localStorage.setItem("unirtm_view_mode_tools", mode);
}

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

// Remote update detection & Diff Drawer state
export type ToolUpdateResult = ToolUpdateItem;

const isDiffDrawerOpen = ref(false);
const batchUpdating = ref(false);
const selectedUpdateTools = ref<string[]>([]);

const isAllUpdatableSelected = computed(() => {
  const updatableIds = filteredUpdateResults.value.map((r) => r.id);
  if (updatableIds.length === 0) return false;
  return updatableIds.every((id) => selectedUpdateTools.value.includes(id));
});

function toggleSelectAllUpdatable() {
  const updatableIds = filteredUpdateResults.value.map((r) => r.id);
  if (isAllUpdatableSelected.value) {
    selectedUpdateTools.value = selectedUpdateTools.value.filter(
      (id) => !updatableIds.includes(id)
    );
  } else {
    const newSet = new Set([...selectedUpdateTools.value, ...updatableIds]);
    selectedUpdateTools.value = Array.from(newSet);
  }
}

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
  if (rt.pinned) {
    const reasonText = rt.pin_reason ? `（锁定原因：${rt.pin_reason}）` : "";
    if (!window.confirm(`警告：工具 ${rt.name} 当前处于版本锁定保护状态${reasonText}。\n确定要强制将其从${scopeName}中移除吗？`)) {
      return;
    }
  }
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

const filterOnlyMissing = ref(false);

const scopeRuntimes = computed(() => {
  let list = runtimes.value;
  if (currentScope.value === "user") {
    return list.filter((rt) => Boolean(rt.global_version && rt.global_version.trim()));
  } else {
    return list.filter((rt) => Boolean((rt.active_version && rt.active_version.trim()) || rt.is_legacy));
  }
});

const scopeMissingCount = computed(() => {
  return scopeRuntimes.value.filter((rt) => getToolStatus(rt).state === "missing").length;
});

const filteredRuntimes = computed(() => {
  let list = scopeRuntimes.value;

  if (filterOnlyMissing.value) {
    list = list.filter((rt) => getToolStatus(rt).state === "missing");
  }

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
      (currentScope.value === "user" ? rt.global_version.includes(q) : rt.active_version.includes(q))
  );
});
function isToolUpgradable(rt: RuntimeItem): boolean {
  if (rt.pinned) return false;
  return Boolean(
    getLocalUpgradableVersion(rt, currentScope.value) ||
    getRemoteUpgradableVersion(rt.id)
  );
}

const scopeUpgradableCount = computed(() => {
  return scopeRuntimes.value.filter(isToolUpgradable).length;
});

const filteredUpdateResults = computed(() => {
  return updateResults.value.filter((r) => {
    if (!r.has_update || r.pinned) return false;
    if (currentScope.value === "user") {
      if (r.in_user !== undefined) return Boolean(r.in_user);
      if (r.scope !== undefined) return r.scope === "user" || r.scope === "both";
      return false;
    } else {
      if (r.in_workspace !== undefined) return Boolean(r.in_workspace);
      if (r.scope !== undefined) return r.scope === "workspace" || r.scope === "both";
      return true;
    }
  });
});

watch(
  () => props.activeSubItem,
  (newVal: string | undefined) => {
    if (newVal) {
      const toolId = newVal.trim().toLowerCase();
      const target = runtimes.value.find((rt) => rt.id.toLowerCase() === toolId);
      if (target) {
        if (currentScope.value === "user" && !target.global_version && target.active_version) {
          currentScope.value = "workspace";
        } else if (currentScope.value === "workspace" && !target.active_version && target.global_version) {
          currentScope.value = "user";
        }
      }
    }
  }
);

// Format tool title with 2-line middle-truncation
function formatToolTitle(title: string, maxLen = 44): string {
  if (!title || title.length <= maxLen) return title;
  const keepStart = 24;
  const keepEnd = maxLen - keepStart - 3;
  return `${title.slice(0, keepStart)}...${title.slice(title.length - keepEnd)}`;
}

// Clean and format version strings strictly with a single 'v', avoiding 'vv'
function formatVersion(ver: string | undefined | null): string {
  if (!ver) return "";
  const clean = ver.trim();
  const stripped = clean.replace(/^[vV]+/, "");
  return `v${stripped}`;
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

// Detect the highest locally installed version that is strictly greater than the current version
function getLocalUpgradableVersion(rt: RuntimeItem, scope: "workspace" | "user" = currentScope.value): string | null {
  if (rt.pinned) return null;

  const currentVer = scope === "workspace" ? (rt.active_version || rt.global_version) : rt.global_version;
  if (!currentVer || !rt.installed_versions || rt.installed_versions.length === 0) {
    return null;
  }

  const cleanCurrent = currentVer.replace(/^v/i, "").trim();
  if (!/[0-9]/.test(cleanCurrent)) return null;

  let highest: string | null = null;

  for (const installed of rt.installed_versions) {
    const cleanInstalled = installed.replace(/^v/i, "").trim();
    if (!cleanInstalled || cleanInstalled === cleanCurrent) continue;

    // compareSemverDesc(cleanInstalled, cleanCurrent) < 0 means cleanInstalled > cleanCurrent
    if (compareSemverDesc(cleanInstalled, cleanCurrent) < 0) {
      if (!highest || compareSemverDesc(cleanInstalled, highest.replace(/^v/i, "").trim()) < 0) {
        highest = installed;
      }
    }
  }

  return highest;
}

function getRemoteUpgradableVersion(toolId: string): string | null {
  const normId = (id: string) => (id || "").replace(/^github:/, "").toLowerCase();
  const shortName = (id: string) => normId(id).split("/").pop() || "";
  const targetNorm = normId(toolId);
  const targetShort = shortName(toolId);

  const match = updateResults.value.find((r) => {
    if (r.id === toolId || r.name === toolId) return true;
    const rNorm = normId(r.id || "");
    const rShort = shortName(r.id || "");
    const rName = (r.name || "").toLowerCase();
    return (
      (targetNorm && rNorm === targetNorm) ||
      (targetShort && (rShort === targetShort || rName === targetShort))
    );
  });
  if (!match || !match.has_update || match.pinned || !match.latest_version) {
    return null;
  }
  const rt = runtimes.value.find((r) => {
    if (r.id === toolId || r.name === toolId) return true;
    const rNorm = normId(r.id || "");
    const rShort = shortName(r.id || "");
    return (
      (targetNorm && rNorm === targetNorm) ||
      (targetShort && rShort === targetShort)
    );
  });
  if (rt) {
    const cur = currentScope.value === "workspace" ? (rt.active_version || rt.global_version) : rt.global_version;
    if (cur) {
      const cleanCur = cur.replace(/^[vV]+/, "").trim();
      const cleanLatest = match.latest_version.replace(/^[vV]+/, "").trim();
      if (cleanLatest === cleanCur || compareSemverDesc(cleanLatest, cleanCur) >= 0) {
        return null;
      }
    }
  }
  return match.latest_version;
}

function getCurrentVersion(rt: RuntimeItem): string {
  if (currentScope.value === "user") {
    return rt.global_version || "";
  }
  return rt.active_version || rt.global_version || "";
}

function getToolStatus(rt: RuntimeItem): ToolStatusResult {
  return computeToolStatus(rt, currentScope.value, updateResults.value);
}

const isSwitchModalOpen = ref(false);
const switchModalRuntime = ref<RuntimeItem | null>(null);

function openSwitchVersionModal(rt: RuntimeItem) {
  switchModalRuntime.value = rt;
  isSwitchModalOpen.value = true;
}

function closeSwitchVersionModal() {
  isSwitchModalOpen.value = false;
  switchModalRuntime.value = null;
}

function handleSelectVersionAndClose(rt: RuntimeItem, ver: string) {
  quickSwitch(rt.id, ver, currentScope.value === "user");
  closeSwitchVersionModal();
}

function openInstallFromSwitch(rt: RuntimeItem) {
  closeSwitchVersionModal();
  openRemoteModal(rt);
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

  // Filter by major tag or installed status if selected
  if (selectedMajorFilter.value === "installed") {
    list = list.filter((v) => isAlreadyInstalled(selectedRuntime.value?.id || "", v.version));
  } else if (selectedMajorFilter.value) {
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

function isActiveInModalScope(toolId: string, ver: string): boolean {
  const rt = runtimes.value.find((r) => r.id === toolId);
  if (!rt) return false;
  const targetVer = modalInstallScope.value === "user" ? rt.global_version : rt.active_version;
  if (!targetVer) return false;
  const cleanVer = ver.replace(/^[vV]+/, "").trim();
  const cleanTarget = targetVer.replace(/^[vV]+/, "").trim();
  return cleanVer === cleanTarget;
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
        syncFromToolsList(toolsList);
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

  // Check if tool is currently pinned
  const targetRt = runtimes.value.find((r) => r.id === toolId);
  if (targetRt && targetRt.pinned) {
    const reasonText = targetRt.pin_reason ? `（锁定原因：${targetRt.pin_reason}）` : "";
    const confirmSwitch = window.confirm(
      `工具 ${targetRt.name} 当前处于版本锁定状态${reasonText}。\n应用此版本将自动解除此锁定，确定要继续切换至 v${ver} 吗？`
    );
    if (!confirmSwitch) {
      return;
    }
    // Automatically unpin
    targetRt.pinned = false;
    targetRt.pin_reason = "";
    if (selectedRuntime.value && selectedRuntime.value.id === toolId) {
      selectedRuntime.value.pinned = false;
      selectedRuntime.value.pin_reason = "";
    }
    fetch("/api/tools/pin", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ tool: toolId, pinned: false, global: isGlobal }),
    }).catch(() => {});
  }

  const prevRuntimes = JSON.parse(JSON.stringify(runtimes.value));
  const prevSelectedRuntime = selectedRuntime.value ? JSON.parse(JSON.stringify(selectedRuntime.value)) : null;

  const updateRuntimeState = (newVer: string, addInstalled: boolean) => {
    runtimes.value = runtimes.value.map((rt) => {
      if (rt.id === toolId) {
        const set = new Set(rt.installed_versions || []);
        if (addInstalled) set.add(newVer);
        return {
          ...rt,
          installed_versions: Array.from(set),
          ...(isGlobal
            ? { global_version: newVer }
            : { active_version: newVer, is_legacy: false, source: ".unirtm.toml" }),
        };
      }
      return rt;
    });

    if (selectedRuntime.value && selectedRuntime.value.id === toolId) {
      const set = new Set(selectedRuntime.value.installed_versions || []);
      if (addInstalled) set.add(newVer);
      selectedRuntime.value = {
        ...selectedRuntime.value,
        installed_versions: Array.from(set),
        ...(isGlobal
          ? { global_version: newVer }
          : { active_version: newVer, is_legacy: false, source: ".unirtm.toml" }),
      };
    }
  };

  // If already installed locally, perform optimistic in-place update without auto-closing modal
  if (alreadyInstalled) {
    updateRuntimeState(ver, false);
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
      updateRuntimeState(ver, true);
    }

    showToast(`成功将 ${toolId}@${ver} 设置为${scopeName}！`, "success");
    emit("log-event", "INFO", `Set ${toolId}@${ver} as ${scopeName}`);
  } catch (err: any) {
    runtimes.value = prevRuntimes;
    if (prevSelectedRuntime) {
      selectedRuntime.value = prevSelectedRuntime;
    }
    showToast(`设置失败: ${err.message}`, "error");
    emit("log-event", "ERROR", `Failed to set version: ${err.message}`);
  } finally {
    installingVer.value = null;
  }
}

async function checkRemoteUpdates() {
  try {
    const isGlobal = currentScope.value === "user";
    const list = await checkUpdates(isGlobal);

    const autoSelected = list
      .filter((item: ToolUpdateItem) => item.has_update && !item.pinned)
      .map((item: ToolUpdateItem) => item.id);
    selectedUpdateTools.value = autoSelected;

    isDiffDrawerOpen.value = true;

    const pinnedCount = list.filter((r) => r.pinned).length;
    if (autoSelected.length > 0) {
      showToast(
        `检测完成：发现 ${autoSelected.length} 个工具可升级${pinnedCount > 0 ? ` (已自动保护 ${pinnedCount} 个锁定工具)` : ""}`,
        "success"
      );
    } else {
      showToast("检测完成：所有工具均已是最新版本或受保护锁定", "info");
    }
  } catch (err: any) {
    showToast(`检测更新失败: ${err.message}`, "error");
  }
}

async function runBatchUpdate() {
  if (selectedUpdateTools.value.length === 0) return;
  batchUpdating.value = true;

  const isGlobal = currentScope.value === "user";
  const toolsPayload = selectedUpdateTools.value
    .map((toolId) => {
      const match = updateResults.value.find((r) => r.id === toolId);
      // Strictly respect lock status: never upgrade pinned tools
      if (match && match.pinned) return null;
      return match ? { id: toolId, version: match.latest_version } : null;
    })
    .filter(Boolean);

  try {
    const res = await fetch("/api/tools/batch-update", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        tools: toolsPayload,
        global: isGlobal,
      }),
    });
    if (!res.ok) {
      const errData = await res.json().catch(() => ({}));
      throw new Error(errData.error || `HTTP ${res.status}`);
    }
    const data = await res.json();
    const updatedCount = data.updated ? data.updated.length : 0;
    const skippedCount = data.skipped ? data.skipped.length : 0;

    let msg = `批量升级完成：已成功升级 ${updatedCount} 个工具`;
    if (skippedCount > 0) {
      msg += `，跳过 ${skippedCount} 个保护项`;
    }
    for (const toolId of selectedUpdateTools.value) {
      markToolUpdated(toolId);
    }
    showToast(msg, "success");
    isDiffDrawerOpen.value = false;
    emit(
      "log-event",
      "INFO",
      `Batch updated ${updatedCount} tools (${isGlobal ? "user" : "workspace"})`
    );
    await fetchStatus();
  } catch (err: any) {
    showToast(`批量升级失败: ${err.message}`, "error");
  } finally {
    batchUpdating.value = false;
  }
}

// Overflow action menu state
const activeMenuRuntimeId = ref<string | null>(null);

function toggleRuntimeMenu(id: string) {
  if (activeMenuRuntimeId.value === id) {
    activeMenuRuntimeId.value = null;
  } else {
    activeMenuRuntimeId.value = id;
  }
}

function closeRuntimeMenu() {
  activeMenuRuntimeId.value = null;
}

function handleTogglePinFromMenu(rt: RuntimeItem) {
  closeRuntimeMenu();
  togglePin(rt);
}

function handleRemoveFromMenu(rt: RuntimeItem) {
  closeRuntimeMenu();
  confirmRemoveRuntime(rt);
}

function handleDocumentClick() {
  if (activeMenuRuntimeId.value) {
    closeRuntimeMenu();
  }
}

function handleGlobalKeyDown(e: KeyboardEvent) {
  if (e.key === "Escape") {
    if (activeMenuRuntimeId.value) {
      closeRuntimeMenu();
    }
    if (isSwitchModalOpen.value) {
      closeSwitchVersionModal();
    }
  }
}

onMounted(() => {
  fetchStatus();
  window.addEventListener("keydown", handleGlobalKeyDown);
  document.addEventListener("click", handleDocumentClick);
});

onBeforeUnmount(() => {
  window.removeEventListener("keydown", handleGlobalKeyDown);
  document.removeEventListener("click", handleDocumentClick);
  if (remoteModalAbortController) {
    remoteModalAbortController.abort();
    remoteModalAbortController = null;
  }
});

</script>

<style scoped>
.view-container,
.tools-view,
.runtimes-view {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  padding: 1.25rem 1.5rem 2.5rem 1.5rem;
  min-height: 100%;
  overflow-y: auto;
  scrollbar-gutter: stable;
  box-sizing: border-box;
}

.view-header-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-shrink: 0;
}

.view-title {
  font-size: 1.25rem;
  font-weight: 700;
  margin: 0 0 0.2rem 0;
  color: var(--text-main);
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.view-icon {
  font-size: 1.3rem;
}

.view-desc {
  font-size: 0.8rem;
  color: var(--text-muted);
  margin: 0;
  line-height: 1.4;
}

/* Scope Selection Bar */
.scope-selection-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.5rem 1rem;
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.03);
  gap: 1rem;
  flex-shrink: 0;
}

.scope-context-pill {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.scope-badge {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-size: 0.72rem;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 9999px;
  background: rgba(2, 132, 199, 0.12);
  color: var(--accent-cyan);
  border: 1px solid rgba(2, 132, 199, 0.25);
}

.scope-badge.user {
  background: rgba(168, 85, 247, 0.12);
  color: #c084fc;
  border-color: rgba(168, 85, 247, 0.25);
}

.scope-context-config {
  font-size: 0.72rem;
  color: var(--text-muted);
  opacity: 0.8;
  display: inline-flex;
  align-items: center;
  gap: 4px;
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
  padding: 0.65rem 1.15rem;
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.03);
  gap: 1rem;
  flex-shrink: 0;
}

.search-input-wrapper {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  flex: 1;
  max-width: 520px;
}

.search-icon {
  font-size: 0.9rem;
  color: var(--text-muted);
}

.tool-search-input {
  width: 100%;
  background: transparent;
  border: none;
  outline: none;
  color: var(--text-main);
  font-size: 0.86rem;
}

.runtime-stats-badges {
  display: flex;
  align-items: center;
  gap: 0.65rem;
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
  background: rgba(16, 185, 129, 0.12);
  border-color: rgba(16, 185, 129, 0.35);
  color: #10b981;
  font-weight: 600;
}

[data-theme="light"] .stat-badge.highlight,
:root[data-theme="light"] .stat-badge.highlight {
  background: #ecfdf5;
  border-color: #a7f3d0;
  color: #059669;
}

.stat-badge-missing {
  background: rgba(245, 158, 11, 0.12);
  border-color: rgba(245, 158, 11, 0.35);
  color: #f59e0b;
  font-weight: 600;
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.stat-badge-missing.active {
  background: rgba(245, 158, 11, 0.25);
  border-color: #f59e0b;
  box-shadow: 0 0 10px rgba(245, 158, 11, 0.35);
}

.stat-badge-missing:hover {
  filter: brightness(1.15);
  box-shadow: 0 0 10px rgba(245, 158, 11, 0.3);
  transform: translateY(-1px);
}

[data-theme="light"] .stat-badge-missing,
:root[data-theme="light"] .stat-badge-missing {
  background: #fef3c7;
  border-color: #fcd34d;
  color: #b45309;
}

[data-theme="light"] .stat-badge-missing.active,
:root[data-theme="light"] .stat-badge-missing.active {
  background: #fde68a;
  border-color: #d97706;
}

.stat-badge.clickable {
  cursor: pointer;
  transition: all 0.2s ease;
}

.stat-badge.clickable:hover {
  filter: brightness(1.15);
  box-shadow: 0 0 8px rgba(16, 185, 129, 0.25);
  transform: translateY(-1px);
}

/* Grid */
.runtimes-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(380px, 1fr));
  gap: 1.25rem;
  min-height: 500px;
  align-content: start;
}

.runtime-card {
  display: flex;
  flex-direction: column;
  gap: 1.15rem;
  padding: 1.4rem 1.5rem;
  border-radius: 16px;
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  box-shadow: 0 4px 18px rgba(0, 0, 0, 0.04);
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  min-height: 200px;
}

.runtime-card:hover {
  border-color: var(--card-border-active);
  transform: translateY(-2px);
  box-shadow: 0 8px 24px var(--accent-cyan-glow);
}

.runtime-card-header {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
}

.runtime-icon-badge {
  font-size: 2.2rem;
  line-height: 1;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.runtime-title-block {
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
  flex: 1;
  min-width: 0;
}

.runtime-name-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.6rem;
  width: 100%;
}

.runtime-name {
  font-size: 1.1rem;
  font-weight: 700;
  color: var(--text-main);
  line-height: 1.35;
  min-height: 1.35em;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
  word-break: break-all;
  overflow-wrap: anywhere;
  flex: 1;
  min-width: 0;
  text-decoration: none;
  cursor: pointer;
  transition: color 0.15s ease, text-decoration 0.15s ease;
}

.runtime-name:hover {
  color: var(--accent-cyan);
  text-decoration: underline;
  text-underline-offset: 3px;
}

[data-theme="light"] .runtime-name:hover,
:root[data-theme="light"] .runtime-name:hover {
  color: #0284c7;
}

.runtime-header-badges {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  flex-shrink: 0;
}

.runtime-version-badge {
  font-size: 0.82rem;
  font-weight: 700;
  padding: 0.22rem 0.65rem;
  border-radius: 6px;
  line-height: 1.3;
  display: inline-flex;
  align-items: center;
}

/* 1. 已安装且就绪（正常态）：沉稳低噪音中性色，不展示任何小图标 */
.runtime-version-badge.is-ready,
.runtime-version-badge.is-active {
  background: rgba(255, 255, 255, 0.08);
  color: #f1f5f9;
  border: 1px solid rgba(255, 255, 255, 0.16);
}

[data-theme="light"] .runtime-version-badge.is-ready,
[data-theme="light"] .runtime-version-badge.is-active,
:root[data-theme="light"] .runtime-version-badge.is-ready,
:root[data-theme="light"] .runtime-version-badge.is-active {
  background: #f1f5f9;
  color: #0f172a;
  border: 1px solid #cbd5e1;
}

/* 2. 已配置但未安装（缺失态）：醒目琥珀警示色，提醒需要下载安装 */
.runtime-version-badge.is-missing {
  background: rgba(245, 158, 11, 0.15);
  color: #f59e0b;
  border: 1px solid rgba(245, 158, 11, 0.38);
  gap: 5px;
}

[data-theme="light"] .runtime-version-badge.is-missing,
:root[data-theme="light"] .runtime-version-badge.is-missing {
  background: #fef3c7;
  color: #b45309;
  border: 1px solid #fcd34d;
}

.missing-dot {
  font-size: 0.6rem;
  line-height: 1;
}

/* 3. 可升级状态（专属绿色）：绿色独享给升级状态 */
.runtime-upgrade-pill {
  font-size: 0.72rem;
  font-weight: 700;
  padding: 0.18rem 0.5rem;
  border-radius: 5px;
  background: rgba(16, 185, 129, 0.16);
  color: #10b981;
  border: 1px solid rgba(16, 185, 129, 0.4);
  display: inline-flex;
  align-items: center;
  gap: 3px;
  cursor: pointer;
  transition: all 0.18s ease;
  line-height: 1.2;
}

.runtime-upgrade-pill:hover {
  background: rgba(16, 185, 129, 0.28);
  border-color: #10b981;
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(16, 185, 129, 0.3);
}

[data-theme="light"] .runtime-upgrade-pill,
:root[data-theme="light"] .runtime-upgrade-pill {
  background: #ecfdf5;
  color: #059669;
  border: 1px solid #a7f3d0;
}

[data-theme="light"] .runtime-upgrade-pill:hover,
:root[data-theme="light"] .runtime-upgrade-pill:hover {
  background: #d1fae5;
  border-color: #34d399;
}

.upgrade-arrow {
  font-weight: 900;
  font-size: 0.76rem;
}


.runtime-version-badge.is-muted {
  background: rgba(148, 163, 184, 0.12);
  color: #94a3b8;
  border: 1px solid rgba(148, 163, 184, 0.25);
  font-weight: 500;
}

.runtime-desc {
  font-size: 0.82rem;
  color: var(--text-muted);
  line-height: 1.5;
  min-height: 2.6em;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

/* Card Meta Info Bar */
.runtime-card-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 0.55rem 0.85rem;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  border-radius: 8px;
  font-size: 0.76rem;
}

.runtime-card-meta .meta-item {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  overflow: hidden;
}

.runtime-card-meta .meta-label {
  color: var(--text-muted);
}

.runtime-card-meta .meta-value {
  color: var(--text-main);
  font-weight: 600;
  text-overflow: ellipsis;
  overflow: hidden;
  white-space: nowrap;
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

.modal-title-text-wrap {
  display: flex;
  align-items: center;
  gap: 0.5rem;
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

/* Card footer: Primary [切换] [安装] + More Menu [···] */
.runtime-card-footer {
  margin-top: auto;
  display: flex;
  align-items: center;
  gap: 0.6rem;
  padding-top: 0.85rem;
  border-top: 1px solid var(--card-border-subtle, rgba(255, 255, 255, 0.06));
}

.runtime-card-footer .card-action-btn.action-switch,
.runtime-card-footer .card-action-btn.action-install {
  flex: 1;
  min-width: 0;
}

.card-action-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.48rem 0.65rem;
  font-size: 0.82rem;
  font-weight: 600;
  border-radius: 7px;
  border: 1px solid var(--card-border, rgba(255, 255, 255, 0.12));
  background: var(--button-secondary-bg, rgba(255, 255, 255, 0.04));
  color: var(--text-main);
  cursor: pointer;
  transition: all 0.18s cubic-bezier(0.4, 0, 0.2, 1);
  white-space: nowrap;
}

/* 主操作：切换 */
.card-action-btn.action-switch {
  background: rgba(2, 132, 199, 0.08);
  border-color: rgba(2, 132, 199, 0.28);
  color: var(--accent-cyan);
}
.card-action-btn.action-switch:hover {
  background: rgba(2, 132, 199, 0.18);
  border-color: var(--accent-cyan);
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(2, 132, 199, 0.2);
}

/* 状态操作：安装 */
.card-action-btn.action-install {
  background: rgba(16, 185, 129, 0.08);
  border-color: rgba(16, 185, 129, 0.28);
  color: #10b981;
}
.card-action-btn.action-install:hover {
  background: rgba(16, 185, 129, 0.18);
  border-color: #10b981;
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(16, 185, 129, 0.2);
}
[data-theme="dark"] .card-action-btn.action-install {
  color: #34d399;
}

/* 更多操作容器与按钮 */
.card-more-action-wrap {
  position: relative;
  display: inline-flex;
}

.card-action-btn.action-more {
  width: 34px;
  height: 34px;
  padding: 0;
  flex-shrink: 0;
  background: rgba(148, 163, 184, 0.08);
  border: 1px solid rgba(148, 163, 184, 0.25);
  color: var(--text-muted);
}
.card-action-btn.action-more:hover,
.card-action-btn.action-more.is-active {
  background: rgba(2, 132, 199, 0.15);
  border-color: var(--accent-cyan);
  color: var(--accent-cyan);
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(2, 132, 199, 0.2);
}
[data-theme="light"] .card-action-btn.action-more {
  background: #f8fafc;
  border-color: #cbd5e1;
  color: #64748b;
}
[data-theme="light"] .card-action-btn.action-more:hover,
[data-theme="light"] .card-action-btn.action-more.is-active {
  background: rgba(2, 132, 199, 0.08);
  border-color: var(--accent-cyan);
  color: var(--accent-cyan);
}

/* 下拉菜单通用样式 */
.runtime-dropdown-menu {
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

[data-theme="light"] .runtime-dropdown-menu {
  background: #ffffff;
  border-color: #e2e8f0;
  box-shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.1), 0 0 12px rgba(2, 132, 199, 0.1);
}

.runtime-dropdown-menu .dropdown-item {
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

.runtime-dropdown-menu .dropdown-item:hover {
  background: var(--input-bg, rgba(255, 255, 255, 0.06));
  color: var(--text-main);
}
[data-theme="light"] .runtime-dropdown-menu .dropdown-item:hover {
  background: #f1f5f9;
}

.runtime-dropdown-menu .dropdown-item.is-pinned {
  color: #d97706;
}
[data-theme="dark"] .runtime-dropdown-menu .dropdown-item.is-pinned {
  color: #fbbf24;
}

.runtime-dropdown-menu .dropdown-item.danger {
  color: #ef4444;
}
.runtime-dropdown-menu .dropdown-item.danger:hover {
  background: rgba(239, 68, 68, 0.12);
  color: #ef4444;
}
[data-theme="dark"] .runtime-dropdown-menu .dropdown-item.danger {
  color: #f87171;
}

.runtime-dropdown-menu .dropdown-divider {
  height: 1px;
  margin: 3px 0;
  background: var(--card-border-subtle, rgba(255, 255, 255, 0.08));
}
[data-theme="light"] .runtime-dropdown-menu .dropdown-divider {
  background: #e2e8f0;
}

/* 下拉菜单动画 */
.dropdown-fade-enter-active,
.dropdown-fade-leave-active {
  transition: all 0.15s cubic-bezier(0.16, 1, 0.3, 1);
}

.dropdown-fade-enter-from,
.dropdown-fade-leave-to {
  opacity: 0;
  transform: translateY(-4px) scale(0.98);
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

/* Switch Version Modal */
.switch-modal-dialog {
  max-width: 440px;
  width: 90%;
  border-radius: 14px;
  padding: 1.25rem 1.4rem;
  box-shadow: 0 16px 40px rgba(0, 0, 0, 0.4);
}

.switch-modal-body {
  padding-top: 0.85rem;
}

.switch-pills-container {
  display: flex;
  flex-wrap: wrap;
  gap: 0.6rem;
  max-height: 280px;
  overflow-y: auto;
  padding: 0.25rem 0.15rem;
}

.switch-ver-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.45rem 0.95rem;
  border-radius: 8px;
  font-size: 0.88rem;
  font-weight: 600;
  cursor: pointer;
  border: 1px solid transparent;
  transition: all 0.18s cubic-bezier(0.4, 0, 0.2, 1);
  outline: none;
}

/* 当前生效版本：绿色高亮与描边 */
.switch-ver-pill.is-active {
  background: rgba(16, 185, 129, 0.15);
  border-color: #10b981;
  color: #10b981;
  box-shadow: 0 0 12px rgba(16, 185, 129, 0.25);
  font-weight: 700;
}

[data-theme="dark"] .switch-ver-pill.is-active {
  background: rgba(16, 185, 129, 0.22);
  border-color: #34d399;
  color: #34d399;
  box-shadow: 0 0 14px rgba(52, 211, 153, 0.3);
}

.pill-active-dot {
  font-size: 0.65rem;
  line-height: 1;
}

.pill-active-tag {
  font-size: 0.7rem;
  padding: 0.05rem 0.35rem;
  border-radius: 4px;
  background: rgba(16, 185, 129, 0.2);
  margin-left: 0.2rem;
}

/* 未生效版本：灰色，悬停变色 */
.switch-ver-pill.is-inactive {
  background: var(--input-bg);
  border-color: var(--card-border);
  color: var(--text-muted);
}

.switch-ver-pill.is-inactive:hover {
  background: rgba(2, 132, 199, 0.1);
  border-color: var(--accent-cyan);
  color: var(--accent-cyan);
  transform: translateY(-1px);
}

.switch-empty-box {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.8rem;
  padding: 1.5rem 0.5rem;
  text-align: center;
}

.switch-empty-tip {
  color: var(--text-muted);
  font-size: 0.88rem;
}

.switch-install-now-btn {
  padding: 0.45rem 1rem;
  border-radius: 6px;
  background: rgba(16, 185, 129, 0.12);
  border: 1px solid rgba(16, 185, 129, 0.35);
  color: #10b981;
  font-size: 0.82rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.switch-install-now-btn:hover {
  background: #10b981;
  color: #fff;
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

.filter-chip.installed-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  border-color: rgba(16, 185, 129, 0.35);
  color: #10b981;
}

.filter-chip.installed-chip:hover {
  background: rgba(16, 185, 129, 0.12);
  border-color: #10b981;
  color: #10b981;
}

.filter-chip.installed-chip.active {
  background: rgba(16, 185, 129, 0.2);
  border-color: #10b981;
  color: #10b981;
  font-weight: 700;
}

.chip-installed-dot {
  font-size: 0.65rem;
  line-height: 1;
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

.runtimes-empty-box {
  grid-column: 1 / -1;
  display: flex !important;
  flex-direction: column !important;
  align-items: center !important;
  justify-content: center !important;
  gap: 0.75rem !important;
  padding: 3.5rem 2rem !important;
  text-align: center !important;
  width: 100% !important;
  box-sizing: border-box !important;
}

.runtimes-empty-box .empty-desc {
  font-size: 0.85rem;
  color: var(--text-muted);
  max-width: 500px;
  line-height: 1.5;
  margin: 0;
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
  color: #10b981;
  background: rgba(16, 185, 129, 0.12);
  border: 1px solid rgba(16, 185, 129, 0.3);
  padding: 0.12rem 0.5rem;
  border-radius: 4px;
  font-weight: 600;
}

.installed-tag.active-scope-tag {
  color: var(--accent-cyan);
  background: rgba(6, 182, 212, 0.15);
  border: 1px solid rgba(6, 182, 212, 0.4);
  box-shadow: 0 0 8px rgba(6, 182, 212, 0.18);
  font-weight: 700;
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

.install-action-btn.is-active-ver {
  background: rgba(6, 182, 212, 0.1);
  border: 1px solid rgba(6, 182, 212, 0.3);
  color: var(--accent-cyan);
  opacity: 0.85;
  cursor: default;
}

.install-action-btn:disabled:not(.is-active-ver) {
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

/* Diff Drawer & Batch Upgrade Styles */
.diff-drawer {
  width: 90%;
  max-width: 540px;
  max-height: 85vh;
  display: flex;
  flex-direction: column;
  padding: 1.25rem 1.4rem;
  background: var(--modal-bg);
  border: 1px solid var(--card-border);
  border-radius: 14px;
  box-shadow: 0 20px 45px rgba(0, 0, 0, 0.35);
  color: var(--text-main);
  animation: diffDrawerIn 0.22s cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes diffDrawerIn {
  from {
    opacity: 0;
    transform: scale(0.96) translateY(12px);
  }
  to {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}

.diff-drawer-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 0.75rem;
}

.diff-drawer-title {
  margin: 0;
  font-size: 1.15rem;
  font-weight: 700;
  color: var(--text-main);
}

.diff-drawer-desc {
  margin: 0.25rem 0 0 0;
  font-size: 0.8rem;
  color: var(--text-muted);
  line-height: 1.4;
}

.diff-select-all-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.45rem 0.65rem;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  border-radius: 8px;
  margin-bottom: 0.5rem;
}

.diff-select-all-label {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--text-main);
  cursor: pointer;
  user-select: none;
}

.diff-drawer-body {
  flex: 1;
  overflow-y: auto;
  margin: 0.25rem 0 0.5rem 0;
  max-height: 50vh;
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  padding: 0.1rem 0;
}

.empty-diff-state {
  padding: 2.5rem 1rem;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  gap: 0.4rem;
}

.empty-diff-icon {
  font-size: 2rem;
  color: #10b981;
  margin-bottom: 0.25rem;
}

.empty-diff-title {
  margin: 0;
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--text-main);
}

.empty-diff-sub {
  margin: 0;
  font-size: 0.78rem;
  color: var(--text-muted);
}

.diff-items-list {
  display: flex;
  flex-direction: column;
  gap: 0.45rem;
}

.diff-item-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.65rem 0.85rem;
  background: var(--card-bg, rgba(255, 255, 255, 0.04));
  border: 1px solid var(--card-border);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.15s ease;
  user-select: none;
}

.diff-item-row:hover {
  background: rgba(2, 132, 199, 0.05);
  border-color: rgba(2, 132, 199, 0.3);
}

.diff-item-row.is-selected {
  background: rgba(2, 132, 199, 0.08);
  border-color: rgba(2, 132, 199, 0.35);
}

.diff-item-left {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  min-width: 0;
  flex: 1;
}

.diff-tool-name-col {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.diff-tool-title {
  font-size: 0.86rem;
  font-weight: 600;
  color: var(--text-main);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.diff-tool-sub {
  font-size: 0.72rem;
  color: var(--text-muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.diff-version-transition {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  flex-shrink: 0;
  margin-left: 0.75rem;
}

.diff-ver-old {
  font-size: 0.78rem;
  color: var(--text-muted);
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  padding: 0.15rem 0.45rem;
  border-radius: 5px;
  text-decoration: line-through;
  opacity: 0.85;
}

.diff-ver-arrow {
  font-size: 0.85rem;
  font-weight: 700;
  color: #10b981;
}

.diff-ver-new {
  font-size: 0.82rem;
  font-weight: 700;
  color: #10b981;
  background: rgba(16, 185, 129, 0.12);
  border: 1px solid rgba(16, 185, 129, 0.3);
  padding: 0.15rem 0.5rem;
  border-radius: 5px;
}

[data-theme="dark"] .diff-ver-new {
  color: #34d399;
  background: rgba(16, 185, 129, 0.18);
  border-color: rgba(52, 211, 153, 0.35);
}

.diff-drawer-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 0.85rem;
  border-top: 1px solid var(--card-border);
  margin-top: 0.25rem;
}

.diff-footer-info {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}

.diff-selection-summary {
  font-size: 0.82rem;
  color: var(--text-muted);
}

.diff-selection-summary b {
  color: var(--accent-cyan);
}

.diff-footer-actions {
  display: flex;
  align-items: center;
  gap: 0.65rem;
}

/* Tools View Mode Switcher */
.view-mode-toggle {
  display: inline-flex;
  align-items: center;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  border-radius: 6px;
  padding: 2px;
  gap: 2px;
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

/* Tools High-Density Table Styles */
.tools-list-table-wrapper {
  width: 100%;
  min-height: 700px;
  overflow-x: auto;
  border: 1px solid var(--card-border);
  border-radius: 14px;
  background: var(--card-bg);
  box-shadow: 0 4px 18px rgba(0, 0, 0, 0.03);
  margin-bottom: 2rem;
}

.tools-table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
  font-size: 0.86rem;
}

.tools-table thead tr {
  background: rgba(255, 255, 255, 0.03);
  border-bottom: 1px solid var(--card-border);
}

.tools-table th {
  padding: 0.85rem 1.25rem;
  font-size: 0.78rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: var(--text-muted);
  white-space: nowrap;
}

.tools-table th.th-tool {
  width: 220px;
}
.tools-table th.th-version {
  width: 160px;
}
.tools-table th.th-actions {
  text-align: right;
  width: 180px;
}

.tool-table-row {
  border-bottom: 1px solid var(--card-border, rgba(255, 255, 255, 0.05));
  transition: background 0.15s ease;
}

.tool-table-row:last-child {
  border-bottom: none;
}

.tool-table-row:hover {
  background: rgba(255, 255, 255, 0.035);
}

.tools-table td {
  padding: 0.8rem 1.25rem;
  vertical-align: middle;
}

.tools-table td.td-actions {
  text-align: right;
}

.table-tool-cell {
  display: inline-flex;
  align-items: center;
  gap: 10px;
}

.table-tool-cell .tool-icon {
  flex-shrink: 0;
}

.table-tool-cell .tool-title {
  font-size: 0.92rem;
  font-weight: 600;
  color: var(--text-main);
  text-decoration: none;
  transition: color 0.15s;
}

.table-tool-cell .tool-title:hover {
  color: var(--accent-cyan);
}

.table-version-cell {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.table-desc-cell {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.tool-desc-text {
  font-size: 0.82rem;
  color: var(--text-muted);
  max-width: 640px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.pin-reason-text-sm {
  font-size: 0.74rem;
  color: #d97706;
}

/* Table Row Actions */
.table-actions-group {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  justify-content: flex-end;
}

.row-action-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 4px 10px;
  font-size: 0.78rem;
  font-weight: 500;
  border-radius: 6px;
  border: 1px solid var(--card-border);
  background: var(--button-secondary-bg, rgba(255, 255, 255, 0.04));
  color: var(--text-main);
  cursor: pointer;
  transition: all 0.15s ease;
  white-space: nowrap;
}

.row-action-btn.action-switch {
  background: rgba(2, 132, 199, 0.08);
  border-color: rgba(2, 132, 199, 0.28);
  color: var(--accent-cyan);
}
.row-action-btn.action-switch:hover {
  background: rgba(2, 132, 199, 0.18);
  border-color: var(--accent-cyan);
}

.row-action-btn.action-install {
  background: rgba(16, 185, 129, 0.08);
  border-color: rgba(16, 185, 129, 0.28);
  color: #10b981;
}
.row-action-btn.action-install:hover {
  background: rgba(16, 185, 129, 0.18);
  border-color: #10b981;
}
[data-theme="dark"] .row-action-btn.action-install {
  color: #34d399;
}

.table-more-action-wrap {
  position: relative;
  display: inline-flex;
}

.row-action-btn.action-more {
  width: 28px;
  height: 28px;
  padding: 0;
  background: rgba(148, 163, 184, 0.08);
  border: 1px solid rgba(148, 163, 184, 0.25);
  color: var(--text-muted);
}
.row-action-btn.action-more:hover,
.row-action-btn.action-more.is-active {
  background: rgba(2, 132, 199, 0.15);
  border-color: var(--accent-cyan);
  color: var(--accent-cyan);
}
[data-theme="light"] .row-action-btn.action-more {
  background: #f8fafc;
  border-color: #cbd5e1;
  color: #64748b;
}
[data-theme="light"] .row-action-btn.action-more:hover,
[data-theme="light"] .row-action-btn.action-more.is-active {
  background: rgba(2, 132, 199, 0.08);
  border-color: var(--accent-cyan);
  color: var(--accent-cyan);
}

.table-more-action-wrap .runtime-dropdown-menu.table-menu {
  top: calc(100% + 4px);
  bottom: auto;
  right: 0;
  text-align: left;
}

/* Light Theme Overrides for Tools Table */
:root[data-theme="light"] .tools-list-table-wrapper,
[data-theme="light"] .tools-list-table-wrapper {
  background: #ffffff;
  border-color: #e2e8f0;
}

:root[data-theme="light"] .tools-table thead tr,
[data-theme="light"] .tools-table thead tr {
  background: #f8fafc;
  border-bottom-color: #e2e8f0;
}

:root[data-theme="light"] .tools-table th,
[data-theme="light"] .tools-table th {
  color: #64748b;
}

:root[data-theme="light"] .tool-table-row,
[data-theme="light"] .tool-table-row {
  border-bottom-color: #f1f5f9;
}

:root[data-theme="light"] .tool-table-row:hover,
[data-theme="light"] .tool-table-row:hover {
  background: rgba(2, 132, 199, 0.03);
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
