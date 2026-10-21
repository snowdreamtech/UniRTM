<template>
  <div class="view-container environments-view" id="view-environments">
    <!-- View Header -->
    <div class="view-header-row">
      <div>
        <h2 class="view-title" style="display: flex; align-items: center; gap: 8px">
          <TechIcon id="workbench:environments" :size="20" />
          <span>{{ t("environments.title") || "环境变量管理" }}</span>
        </h2>
        <p class="view-desc">
          {{
            t("environments.desc") ||
            "统一管理工作区 (.unirtm.toml) 与用户全局 (~/.config/unirtm/unirtm.toml) 配置文件中的 [env] 环境变量。敏感密钥默认脱密遮蔽。"
          }}
        </p>
      </div>
      <div class="view-actions" style="display: flex; gap: 8px; align-items: center">
        <button
          class="tool-action-btn secondary"
          @click="fetchEnvVars"
          :disabled="loading"
          :title="t('common.refresh') || '刷新'"
        >
          <TechIcon id="workbench:sync" :size="14" :class="{ 'spin-icon': loading }" />
          <span>{{ loading ? t("common.refreshing") || "刷新中..." : t("common.refresh") || "刷新" }}</span>
        </button>
      </div>
    </div>

    <!-- Current Scope Context Banner -->
    <div class="scope-selection-bar glass-card" id="environments-scope-bar">
      <div class="scope-context-pill">
        <span class="scope-badge" :class="activeScope">
          <TechIcon
            :id="activeScope === 'user' ? 'workbench:settings' : 'workbench:folder'"
            :size="13"
            :colored="false"
          />
          {{
            activeScope === "user"
              ? t("environments.scope_user") || "用户"
              : t("environments.scope_workspace") || "工作区"
          }}
        </span>
        <span class="scope-hint-text">
          {{
            activeScope === "user"
              ? t("environments.scope_user_desc") || "定义在用户全局 ~/.config/unirtm/unirtm.toml 的 [env] 配置块中"
              : t("environments.scope_workspace_desc") || "定义在当前项目 .unirtm.toml 的 [env] 配置块中"
          }}
        </span>
      </div>
      <div
        class="scope-context-config font-mono"
        :title="activeScope === 'user' ? '~/.config/unirtm/unirtm.toml' : '.unirtm.toml'"
      >
        <span class="file-icon">📄</span>
        <span>{{ activeScope === "user" ? "~/.config/unirtm/unirtm.toml" : ".unirtm.toml" }}</span>
      </div>
    </div>

    <!-- Environment Profiles Capsule Bar (Workspace Only) -->
    <div v-if="activeScope === 'workspace'" class="profiles-bar glass-card" id="env-profiles-bar">
      <div class="profiles-list">
        <!-- Default Baseline Tab -->
        <button
          type="button"
          class="profile-tab-btn"
          :class="{ active: activeProfile === '' }"
          @click="selectProfile('')"
        >
          <span class="tab-icon">✨</span>
          <span class="tab-label">{{ t("environments.profile_default") || "默认基准 (Default)" }}</span>
        </button>

        <!-- User Defined Profiles -->
        <div v-for="p in profilesList" :key="p" class="profile-tab-item" :class="{ active: activeProfile === p }">
          <button
            type="button"
            class="profile-tab-btn profile-custom"
            :class="{ active: activeProfile === p }"
            @click="selectProfile(p)"
          >
            <span class="tab-icon">{{ getProfileIcon(p) }}</span>
            <span class="tab-label font-mono">{{ p }}</span>
            <span v-if="getProfileOverrideCount(p) > 0" class="profile-override-chip">
              {{ getProfileOverrideCount(p) }} {{ t("environments.profile_overrides_chip") || "覆盖" }}
            </span>
          </button>
          <button
            type="button"
            class="btn btn-icon danger btn-delete-profile"
            :title="t('environments.delete_profile') || '删除此环境配置'"
            @click.stop="confirmDeleteProfile(p)"
          >
            <TechIcon id="workbench:close" :size="11" />
          </button>
        </div>

        <!-- Add Profile Trigger -->
        <button
          type="button"
          class="profile-tab-btn btn-add-profile"
          @click="openNewProfileModal"
          :title="t('environments.create_profile') || '新建环境 Profile'"
        >
          <span>+ {{ t("environments.create_profile") || "新建环境 Profile" }}</span>
        </button>
      </div>

      <!-- Active Profile Context Indicator -->
      <div v-if="activeProfile" class="active-profile-hint">
        <span class="profile-current-tag"> {{ getProfileIcon(activeProfile) }} {{ activeProfile }} </span>
        <button
          type="button"
          class="filter-toggle-btn"
          :class="{ active: showOnlyOverrides }"
          @click="showOnlyOverrides = !showOnlyOverrides"
        >
          <span>{{ showOnlyOverrides ? "● 仅看特有与覆盖项" : "全部生效变量" }}</span>
        </button>
      </div>
    </div>

    <!-- Filter & Search Toolbar -->
    <div class="env-filter-bar glass-card">
      <div class="search-input-wrapper">
        <span class="search-icon">
          <TechIcon id="workbench:search" :size="14" :colored="false" />
        </span>
        <input
          v-model="searchKey"
          type="text"
          class="env-search-input"
          :placeholder="t('environments.search_placeholder') || '搜索环境变量名称、关键字或值...'"
        />
      </div>

      <div class="env-stats-badges">
        <span class="stat-badge">
          <b>{{ filteredEnvList.length }}</b> {{ t("environments.count_suffix") || "个变量" }}
        </span>
        <span
          v-if="sensitiveCount > 0"
          class="stat-badge sensitive-stat"
          :title="t('environments.sensitive_hint') || '已自动启用默认脱密保护'"
        >
          <TechIcon id="workbench:lock" :size="12" />
          <b>{{ sensitiveCount }}</b> {{ t("environments.sensitive_count_suffix") || "个敏感" }}
        </span>

        <!-- View Mode Switcher -->
        <div class="view-mode-toggle" id="env-view-mode-toggle">
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
          id="btn-add-env"
          @click="openCreateModal"
          :title="t('environments.create_btn') || '新建变量'"
        >
          <span>+ {{ t("environments.create_btn") || "新建变量" }}</span>
        </button>
      </div>
    </div>

    <!-- Environments List Container -->
    <div class="env-table-container glass-card">
      <!-- Loading State -->
      <div v-if="loading && envList.length === 0" class="env-empty-box">
        <TechIcon id="workbench:sync" :size="24" class="spin-icon" />
        <span style="font-size: 0.85rem; color: var(--text-muted)">{{
          t("environments.loading") || "正在读取 [env] 配置..."
        }}</span>
      </div>

      <!-- Empty State -->
      <div v-else-if="filteredEnvList.length === 0" class="env-empty-box">
        <TechIcon id="workbench:environments" :size="32" :colored="false" />
        <div class="empty-title">
          {{
            searchKey
              ? t("environments.empty_search", { query: searchKey }) || `未找到匹配 "${searchKey}" 的环境变量`
              : t("environments.empty_title") || "暂未配置任何环境变量"
          }}
        </div>
        <p class="empty-desc">
          {{
            activeScope === "user"
              ? t("environments.empty_user_desc") ||
                "可在用户全局配置 (~/.config/unirtm/unirtm.toml) 中添加 [env] 区块，为全局所有开发环境提供兜底环境变量。"
              : t("environments.empty_workspace_desc") ||
                "可在当前项目根目录的 .unirtm.toml 中添加 [env] 区块，锁定当前项目的专属环境变量与服务端口。"
          }}
        </p>
        <button v-if="!searchKey" class="tool-action-btn primary" style="margin-top: 0.75rem" @click="openCreateModal">
          <TechIcon id="workbench:edit" :size="14" />
          <span>{{ t("environments.create_first") || "创建第一个环境变量" }}</span>
        </button>
      </div>

      <!-- Environments Grid / Cards View -->
      <div v-else-if="viewMode === 'cards'" class="env-grid">
        <div
          v-for="item in filteredEnvList"
          :key="item.name"
          class="env-card"
          :class="{ sensitive: item.is_sensitive }"
        >
          <!-- Card Header Line -->
          <div class="env-card-header">
            <div class="env-key-wrap">
              <TechIcon
                :id="item.is_sensitive ? 'workbench:key' : 'workbench:variable'"
                :size="15"
                :style="{ color: item.is_sensitive ? '#f59e0b' : 'var(--accent-cyan)' }"
                class="key-type-icon"
              />
              <span class="env-key font-mono">{{ item.name }}</span>
              <span
                class="sec-tag"
                :class="item.is_sensitive ? 'sensitive' : 'public'"
                :title="
                  item.is_sensitive
                    ? t('environments.sensitive_badge') || '敏感密钥'
                    : t('environments.public_badge') || '公开变量'
                "
              >
                <TechIcon :id="item.is_sensitive ? 'workbench:lock' : 'workbench:globe'" :size="11" />
                <span>{{
                  item.is_sensitive
                    ? t("environments.sensitive_badge") || "敏感密钥"
                    : t("environments.public_badge") || "公开变量"
                }}</span>
              </span>
              <!-- Profile Override Badges -->
              <span v-if="item.is_override" class="sec-tag override" :title="'基准默认值: ' + item.base_value">
                <span>🔄 {{ t("environments.badge_override") || "已覆盖" }}</span>
              </span>
              <span v-else-if="item.is_profile_only" class="sec-tag profile-only">
                <span>🌱 {{ t("environments.badge_profile_only") || "环境专属" }}</span>
              </span>
              <span v-else-if="item.is_inherited && activeProfile" class="sec-tag inherited">
                <span>⚡ {{ t("environments.badge_inherited") || "继承基准" }}</span>
              </span>

              <span class="scope-tag" :class="item.scope">
                {{
                  item.scope === "user"
                    ? t("environments.scope_user") || "用户"
                    : t("environments.scope_workspace") || "工作区"
                }}
              </span>
            </div>

            <!-- Card Actions: Edit & Delete -->
            <div class="card-btn-group">
              <button
                v-if="item.is_override"
                class="btn btn-icon success restore-btn"
                @click="restoreBaseValue(item.name)"
                :title="t('environments.restore_base') || '恢复基准值'"
              >
                <TechIcon id="workbench:sync" :size="13" />
              </button>
              <button
                class="btn btn-icon accent edit-btn"
                @click="openEditModal(item)"
                :title="t('environments.edit_btn') || '编辑'"
              >
                <TechIcon id="workbench:edit" :size="14" />
              </button>
              <button
                class="btn btn-icon danger delete-btn"
                @click="confirmDelete(item.name)"
                :title="t('environments.delete_btn') || '删除'"
              >
                <TechIcon id="workbench:delete" :size="14" />
              </button>
            </div>
          </div>

          <!-- Value Display Row with Mask & Eye Toggle -->
          <div class="env-card-value-row">
            <div class="value-display-box font-mono" :class="{ masked: item.is_sensitive && !item.isRevealed }">
              <span v-if="item.is_sensitive && !item.isRevealed" class="masked-dots"> •••••••••••••••• </span>
              <span v-else class="revealed-text">
                {{ item.value }}
              </span>
            </div>

            <div class="value-actions">
              <!-- Eye Reveal / Mask Toggle Button -->
              <button
                v-if="item.is_sensitive"
                class="value-btn eye-btn"
                :class="{ active: item.isRevealed }"
                @click="toggleReveal(item)"
                :title="
                  item.isRevealed ? t('environments.conceal') || '脱密隐藏' : t('environments.reveal') || '显示明文'
                "
              >
                <TechIcon :id="item.isRevealed ? 'workbench:eye-closed' : 'workbench:eye'" :size="14" />
                <span class="btn-text-sm">{{
                  item.isRevealed ? t("environments.conceal_short") || "密文" : t("environments.reveal_short") || "明文"
                }}</span>
              </button>

              <!-- One-click Copy Button -->
              <button
                class="value-btn copy-btn"
                @click="copyValue(item)"
                :title="t('environments.copy_val') || '复制真实值'"
              >
                <TechIcon v-if="copiedKey === item.name" id="workbench:success" :size="13" class="copy-check" />
                <TechIcon v-else id="workbench:copy" :size="13" />
                <span class="btn-text-sm">{{
                  copiedKey === item.name ? t("common.copied") || "已复制" : t("common.copy") || "复制"
                }}</span>
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Environments High-Density List / Table View -->
      <div v-else class="env-list-table-wrapper">
        <table class="env-table">
          <thead>
            <tr>
              <th class="th-key">{{ t("environments.col_key") || "变量名" }}</th>
              <th class="th-scope">{{ t("environments.col_scope") || "作用域" }}</th>
              <th class="th-value">{{ t("environments.col_value") || "变量值" }}</th>
              <th class="th-actions">{{ t("environments.col_actions") || "操作" }}</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="item in filteredEnvList"
              :key="item.name"
              class="env-table-row"
              :class="{ sensitive: item.is_sensitive }"
            >
              <!-- Col: Key -->
              <td class="td-key">
                <div class="table-key-cell">
                  <TechIcon
                    :id="item.is_sensitive ? 'workbench:key' : 'workbench:variable'"
                    :size="15"
                    :style="{ color: item.is_sensitive ? '#f59e0b' : 'var(--accent-cyan)' }"
                    class="key-type-icon"
                  />
                  <span class="key-name font-mono">{{ item.name }}</span>
                  <span
                    v-if="item.is_sensitive"
                    class="sec-tag sensitive mini"
                    :title="t('environments.sensitive_badge') || '敏感密钥'"
                  >
                    <TechIcon id="workbench:lock" :size="10" />
                    <span>{{ t("environments.sensitive_badge") || "敏感" }}</span>
                  </span>
                  <!-- Profile Override Badges in Table View -->
                  <span v-if="item.is_override" class="sec-tag override mini" :title="'基准默认值: ' + item.base_value">
                    <span>🔄 {{ t("environments.badge_override") || "已覆盖" }}</span>
                  </span>
                  <span v-else-if="item.is_profile_only" class="sec-tag profile-only mini">
                    <span>🌱 {{ t("environments.badge_profile_only") || "环境专属" }}</span>
                  </span>
                  <span v-else-if="item.is_inherited && activeProfile" class="sec-tag inherited mini">
                    <span>⚡ {{ t("environments.badge_inherited") || "继承基准" }}</span>
                  </span>
                </div>
              </td>

              <!-- Col: Scope -->
              <td class="td-scope">
                <span class="scope-tag" :class="item.scope">
                  {{
                    item.scope === "user"
                      ? t("environments.scope_user") || "用户"
                      : t("environments.scope_workspace") || "工作区"
                  }}
                </span>
              </td>

              <!-- Col: Value -->
              <td class="td-value font-mono">
                <div class="table-value-cell">
                  <div class="table-value-text" :class="{ masked: item.is_sensitive && !item.isRevealed }">
                    <span v-if="item.is_sensitive && !item.isRevealed" class="masked-dots"> •••••••••••••••• </span>
                    <span v-else class="revealed-text">
                      {{ item.value }}
                    </span>
                  </div>
                  <div class="table-value-actions">
                    <button
                      v-if="item.is_sensitive"
                      class="value-btn eye-btn mini"
                      :class="{ active: item.isRevealed }"
                      @click="toggleReveal(item)"
                      :title="
                        item.isRevealed
                          ? t('environments.conceal') || '脱密隐藏'
                          : t('environments.reveal') || '显示明文'
                      "
                    >
                      <TechIcon :id="item.isRevealed ? 'workbench:eye-closed' : 'workbench:eye'" :size="13" />
                    </button>
                    <button
                      class="value-btn copy-btn mini"
                      @click="copyValue(item)"
                      :title="t('environments.copy_val') || '复制真实值'"
                    >
                      <TechIcon v-if="copiedKey === item.name" id="workbench:success" :size="13" class="copy-check" />
                      <TechIcon v-else id="workbench:copy" :size="13" />
                    </button>
                  </div>
                </div>
              </td>

              <!-- Col: Actions -->
              <td class="td-actions">
                <div class="table-row-actions">
                  <button
                    v-if="item.is_override"
                    class="btn btn-icon success restore-btn"
                    @click="restoreBaseValue(item.name)"
                    :title="t('environments.restore_base') || '恢复基准值'"
                  >
                    <TechIcon id="workbench:sync" :size="13" />
                  </button>
                  <button
                    class="btn btn-icon accent edit-btn"
                    @click="openEditModal(item)"
                    :title="t('environments.edit_btn') || '编辑'"
                  >
                    <TechIcon id="workbench:edit" :size="14" />
                  </button>
                  <button
                    class="btn btn-icon danger delete-btn"
                    @click="confirmDelete(item.name)"
                    :title="t('environments.delete_btn') || '删除'"
                  >
                    <TechIcon id="workbench:delete" :size="14" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Create / Edit Environment Modal -->
    <EnvironmentModal
      :is-open="isModalOpen"
      :mode="modalMode"
      :initial-data="selectedEnv"
      :active-scope="activeScope"
      :active-profile="activeScope === 'workspace' ? activeProfile : ''"
      @close="isModalOpen = false"
      @save="handleSaveEnv"
    />

    <!-- Create Profile Modal -->
    <div v-if="isNewProfileModalOpen" class="confirm-overlay" @click.self="isNewProfileModalOpen = false">
      <div class="confirm-modal glass-card" style="max-width: 440px">
        <div class="confirm-header">
          <TechIcon id="workbench:environments" :size="20" />
          <h4>{{ t("environments.create_profile_title") || "新建环境 Profile" }}</h4>
        </div>
        <p class="confirm-desc">
          {{
            t("environments.create_profile_desc") ||
            "为特定场景（如本地开发、CI 检查、测试环境）定义差异化环境变量覆盖。"
          }}
        </p>
        <div class="profile-modal-body" style="display: flex; flex-direction: column; gap: 10px; margin: 12px 0">
          <input
            v-model="newProfileName"
            type="text"
            class="env-search-input font-mono"
            style="width: 100%; border: 1px solid var(--card-border); padding: 8px 12px; border-radius: 6px"
            :placeholder="t('environments.profile_name_placeholder') || '例如 development, ci, staging'"
            @keydown.enter="executeCreateProfile"
          />
          <div style="display: flex; gap: 6px; flex-wrap: wrap">
            <button
              v-for="p in ['development', 'ci', 'staging', 'production', 'testing']"
              :key="p"
              type="button"
              class="preset-chip font-mono"
              @click="newProfileName = p"
            >
              {{ p }}
            </button>
          </div>
        </div>
        <div class="confirm-actions">
          <button class="tool-action-btn secondary" @click="isNewProfileModalOpen = false">
            {{ t("common.cancel") || "取消" }}
          </button>
          <button class="tool-action-btn primary" :disabled="!newProfileName.trim()" @click="executeCreateProfile">
            <span>{{ t("common.create") || "创建" }}</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Delete Profile Confirmation Modal -->
    <div v-if="deleteProfileConfirm" class="confirm-overlay" @click.self="deleteProfileConfirm = ''">
      <div class="confirm-modal glass-card">
        <div class="confirm-header">
          <TechIcon id="workbench:warning" :size="22" />
          <h4>{{ t("environments.delete_profile_title") || "确认删除环境 Profile" }}</h4>
        </div>
        <p class="confirm-desc">
          {{ t("environments.delete_profile_warning", { profile: deleteProfileConfirm }) }}
        </p>
        <div class="confirm-actions">
          <button class="tool-action-btn secondary" @click="deleteProfileConfirm = ''">
            {{ t("common.cancel") || "取消" }}
          </button>
          <button class="tool-action-btn danger" @click="executeDeleteProfile(deleteProfileConfirm)">
            <span>{{ t("common.delete") || "删除" }}</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Delete Confirmation Modal -->
    <div v-if="deleteConfirmKey" class="confirm-overlay" @click.self="deleteConfirmKey = ''">
      <div class="confirm-modal glass-card">
        <div class="confirm-header">
          <TechIcon id="workbench:warning" :size="22" />
          <h4>{{ t("environments.delete_title") || "确认删除环境变量" }}</h4>
        </div>
        <p class="confirm-desc">
          {{
            t("environments.delete_warning", {
              key: deleteConfirmKey,
              scope:
                activeScope === "user"
                  ? t("environments.scope_user") || "用户"
                  : t("environments.scope_workspace") || "工作区",
            }) ||
            `确定要从 ${activeScope === "user" ? "用户全局配置" : "当前工作区配置"} 中彻底删除环境变量 ${deleteConfirmKey} 吗？删除后将不可恢复。`
          }}
        </p>
        <div class="confirm-actions">
          <button class="tool-action-btn secondary" @click="deleteConfirmKey = ''" :disabled="isDeleting">
            {{ t("common.cancel") || "取消" }}
          </button>
          <button class="tool-action-btn danger" @click="executeDelete(deleteConfirmKey)" :disabled="isDeleting">
            <TechIcon v-if="isDeleting" id="workbench:sync" :size="14" class="spin-icon" />
            <span>{{ isDeleting ? t("common.deleting") || "删除中..." : t("common.delete") || "删除" }}</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Floating Toast Notification -->
    <transition name="toast-fade">
      <div v-if="toastMessage" class="toast-banner">
        <span>{{ toastMessage }}</span>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from "vue";
import { useRoute } from "vue-router";
import { t } from "../i18n";
import TechIcon from "../components/common/TechIcon.vue";
import EnvironmentModal from "../components/EnvironmentModal.vue";
import { useAppScope } from "../composables/useAppScope";

interface EnvItem {
  name: string;
  value: string;
  scope: "workspace" | "user";
  profile?: string;
  is_sensitive: boolean;
  isRevealed?: boolean;
  is_override?: boolean;
  base_value?: string;
  is_profile_only?: boolean;
  is_inherited?: boolean;
}

const route = useRoute();
const { activeScope, syncFromRoute } = useAppScope();

const envList = ref<EnvItem[]>([]);
const profilesList = ref<string[]>([]);
const activeProfile = ref<string>("");
const showOnlyOverrides = ref(false);
const isNewProfileModalOpen = ref(false);
const newProfileName = ref("");
const deleteProfileConfirm = ref("");

const loading = ref(false);
const searchKey = ref("");
const isModalOpen = ref(false);
const modalMode = ref<"create" | "edit">("create");
const selectedEnv = ref<{ name: string; value: string; scope: "workspace" | "user" } | undefined>(undefined);
const deleteConfirmKey = ref("");
const isDeleting = ref(false);
const copiedKey = ref("");
const toastMessage = ref("");

function getProfileIcon(p: string) {
  const low = p.toLowerCase();
  if (low.includes("dev")) return "🌱";
  if (low.includes("ci")) return "⚡";
  if (low.includes("prod")) return "🚀";
  if (low.includes("stag")) return "📦";
  if (low.includes("test")) return "🧪";
  return "🏷️";
}

function getProfileOverrideCount(p: string) {
  if (p === activeProfile.value) {
    return envList.value.filter((e) => e.is_override || e.is_profile_only).length;
  }
  return 0;
}

function selectProfile(p: string) {
  activeProfile.value = p;
  fetchEnvVars();
}

function openNewProfileModal() {
  newProfileName.value = "";
  isNewProfileModalOpen.value = true;
}

async function executeCreateProfile() {
  const p = newProfileName.value.trim().toLowerCase();
  if (!p) return;
  try {
    const res = await fetch("/api/env", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        scope: "workspace",
        profile: p,
        action: "create_profile",
      }),
    });
    if (res.ok) {
      isNewProfileModalOpen.value = false;
      showToast(`✓ ${t("environments.create_profile") || "创建环境 Profile"}: ${p}`);
      activeProfile.value = p;
      await fetchEnvVars();
    } else {
      const err = await res.json();
      showToast("创建环境失败: " + (err.error || "未知错误"));
    }
  } catch (err: any) {
    showToast("创建错误: " + err.message);
  }
}

function confirmDeleteProfile(p: string) {
  deleteProfileConfirm.value = p;
}

async function executeDeleteProfile(p: string) {
  try {
    const res = await fetch(`/api/env?scope=workspace&profile=${encodeURIComponent(p)}&action=delete_profile`, {
      method: "DELETE",
    });
    if (res.ok) {
      deleteProfileConfirm.value = "";
      showToast(`✓ 已成功删除环境 Profile: ${p}`);
      if (activeProfile.value === p) {
        activeProfile.value = "";
      }
      await fetchEnvVars();
    } else {
      const err = await res.json();
      showToast("删除环境失败: " + (err.error || "未知错误"));
    }
  } catch (err: any) {
    showToast("删除错误: " + err.message);
  }
}

async function restoreBaseValue(key: string) {
  if (!activeProfile.value) return;
  try {
    const res = await fetch(
      `/api/env?key=${encodeURIComponent(key)}&scope=workspace&profile=${encodeURIComponent(activeProfile.value)}`,
      {
        method: "DELETE",
      }
    );
    if (res.ok) {
      showToast(`✓ 已恢复 ${key} 为默认基准值`);
      await fetchEnvVars();
    } else {
      const err = await res.json();
      showToast("恢复基准失败: " + (err.error || "未知错误"));
    }
  } catch (err: any) {
    showToast("操作失败: " + err.message);
  }
}

const viewMode = ref<"cards" | "list">(
  (localStorage.getItem("unirtm_view_mode_environments") as "cards" | "list") || "cards"
);

function setViewMode(mode: "cards" | "list") {
  viewMode.value = mode;
  localStorage.setItem("unirtm_view_mode_environments", mode);
}

function showToast(msg: string) {
  toastMessage.value = msg;
  setTimeout(() => {
    if (toastMessage.value === msg) {
      toastMessage.value = "";
    }
  }, 2500);
}

watch(
  () => route.query.scope,
  () => {
    syncFromRoute(route);
  },
  { immediate: true }
);

watch(
  () => activeScope.value,
  () => {
    activeProfile.value = "";
    fetchEnvVars();
  }
);

async function fetchEnvVars() {
  loading.value = true;
  try {
    const pParam =
      activeScope.value === "workspace" && activeProfile.value
        ? `&profile=${encodeURIComponent(activeProfile.value)}`
        : "";
    const res = await fetch(`/api/env?scope=${activeScope.value}${pParam}`);
    if (res.ok) {
      const data = await res.json();
      if (Array.isArray(data.profiles)) {
        profilesList.value = data.profiles;
      }
      if (Array.isArray(data.env)) {
        envList.value = data.env.map((e: any) => ({
          name: e.name,
          value: e.value,
          scope: e.scope || activeScope.value,
          profile: e.profile || "",
          is_sensitive: Boolean(e.is_sensitive),
          isRevealed: false,
          is_override: Boolean(e.is_override),
          base_value: e.base_value || "",
          is_profile_only: Boolean(e.is_profile_only),
          is_inherited: Boolean(e.is_inherited),
        }));
      }
    }
  } catch (err: any) {
    showToast("读取环境变量失败: " + err.message);
  } finally {
    loading.value = false;
  }
}

const filteredEnvList = computed(() => {
  let list = envList.value;
  if (activeScope.value === "workspace" && activeProfile.value && showOnlyOverrides.value) {
    list = list.filter((e) => e.is_override || e.is_profile_only);
  }
  if (!searchKey.value) return list;
  const q = searchKey.value.toLowerCase().trim();
  return list.filter((e) => e.name.toLowerCase().includes(q) || e.value.toLowerCase().includes(q));
});

const sensitiveCount = computed(() => {
  return filteredEnvList.value.filter((e) => e.is_sensitive).length;
});

function toggleReveal(item: EnvItem) {
  item.isRevealed = !item.isRevealed;
}

async function copyValue(item: EnvItem) {
  try {
    if (navigator?.clipboard?.writeText) {
      await navigator.clipboard.writeText(item.value);
    } else {
      const input = document.createElement("input");
      input.value = item.value;
      document.body.appendChild(input);
      input.select();
      document.execCommand("copy");
      document.body.removeChild(input);
    }
    copiedKey.value = item.name;
    showToast(
      `✓ ${t("environments.copied_to_clipboard", { key: item.name }) || "已复制 " + item.name + " 真实值到剪贴板"}`
    );
    setTimeout(() => {
      if (copiedKey.value === item.name) {
        copiedKey.value = "";
      }
    }, 2000);
  } catch {
    showToast("复制失败，请手动选择复制");
  }
}

function openCreateModal() {
  modalMode.value = "create";
  selectedEnv.value = undefined;
  isModalOpen.value = true;
}

function openEditModal(item: EnvItem) {
  modalMode.value = "edit";
  selectedEnv.value = {
    name: item.name,
    value: item.value,
    scope: item.scope,
  };
  isModalOpen.value = true;
}

async function handleSaveEnv(payload: {
  key: string;
  original_key?: string;
  value: string;
  scope: "workspace" | "user";
  profile?: string;
}) {
  try {
    const res = await fetch("/api/env", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    if (res.ok) {
      isModalOpen.value = false;
      showToast(
        modalMode.value === "create"
          ? `✓ ${t("environments.create_success", { key: payload.key }) || "成功创建环境变量 " + payload.key}`
          : `✓ ${t("environments.update_success", { key: payload.key }) || "成功更新环境变量 " + payload.key}`
      );
      await fetchEnvVars();
    } else {
      const err = await res.json();
      showToast("保存失败: " + (err.error || "未知错误"));
    }
  } catch (err: any) {
    showToast("网络保存错误: " + err.message);
  }
}

function confirmDelete(name: string) {
  deleteConfirmKey.value = name;
}

async function executeDelete(name: string) {
  isDeleting.value = true;
  try {
    const pParam =
      activeScope.value === "workspace" && activeProfile.value
        ? `&profile=${encodeURIComponent(activeProfile.value)}`
        : "";
    const res = await fetch(`/api/env?key=${encodeURIComponent(name)}&scope=${activeScope.value}${pParam}`, {
      method: "DELETE",
    });
    if (res.ok) {
      deleteConfirmKey.value = "";
      showToast(`✓ ${t("environments.delete_success", { key: name }) || "已成功删除环境变量 " + name}`);
      await fetchEnvVars();
    } else {
      const err = await res.json();
      showToast("删除失败: " + (err.error || "未知错误"));
    }
  } catch (err: any) {
    showToast("网络删除错误: " + err.message);
  } finally {
    isDeleting.value = false;
  }
}

onMounted(() => {
  fetchEnvVars();
});
</script>

<style scoped>
.environments-view {
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
  margin: 0;
  font-size: 1.35rem;
  font-weight: 700;
  color: var(--text-main);
}

.view-desc {
  margin: 4px 0 0;
  font-size: 0.85rem;
  color: var(--text-muted);
}

/* Scope Banner */
.scope-selection-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.65rem 1rem;
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.03);
  gap: 1rem;
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

/* Environment Profiles Capsule Bar */
.profiles-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.55rem 0.85rem;
  border-radius: 10px;
  gap: 1rem;
  flex-wrap: wrap;
}

.profiles-list {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.profile-tab-item {
  display: inline-flex;
  align-items: center;
  border-radius: 8px;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  transition: all 0.2s;
}

.profile-tab-item.active {
  border-color: var(--accent-cyan);
  background: rgba(0, 229, 255, 0.08);
}

.profile-tab-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 5px 12px;
  font-size: 0.8rem;
  font-weight: 500;
  color: var(--text-muted);
  background: transparent;
  border: none;
  cursor: pointer;
  border-radius: 8px;
  transition: all 0.2s;
}

.profile-tab-btn:hover {
  color: var(--text-main);
}

.profile-tab-btn.active {
  color: var(--accent-cyan);
  font-weight: 600;
}

.profiles-list > .profile-tab-btn {
  background: var(--input-bg);
  border: 1px solid var(--card-border);
}

.profiles-list > .profile-tab-btn.active {
  background: rgba(0, 229, 255, 0.08);
  border-color: var(--accent-cyan);
}

.btn-delete-profile {
  background: transparent;
  border: none;
  color: var(--text-muted);
  font-size: 0.75rem;
  padding: 4px 6px;
  cursor: pointer;
  border-radius: 4px;
  line-height: 1;
  transition: all 0.15s;
}

.btn-delete-profile:hover {
  color: #ef4444;
  background: rgba(239, 68, 68, 0.1);
}

.profile-override-chip {
  font-size: 0.68rem;
  padding: 1px 6px;
  border-radius: 9999px;
  background: rgba(16, 185, 129, 0.15);
  color: #10b981;
  font-weight: 600;
}

.btn-add-profile {
  font-size: 0.78rem;
  color: var(--text-muted);
  border: 1px dashed var(--card-border) !important;
}

.btn-add-profile:hover {
  border-color: var(--accent-cyan) !important;
  color: var(--accent-cyan);
}

.active-profile-hint {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.78rem;
}

.profile-current-tag {
  color: var(--accent-cyan);
  font-weight: 600;
}

.filter-toggle-btn {
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  color: var(--text-muted);
  padding: 3px 10px;
  border-radius: 6px;
  font-size: 0.75rem;
  cursor: pointer;
  transition: all 0.2s;
}

.filter-toggle-btn.active {
  color: #10b981;
  border-color: #10b981;
  background: rgba(16, 185, 129, 0.1);
}

.preset-chip {
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  color: var(--text-muted);
  padding: 3px 8px;
  border-radius: 6px;
  font-size: 0.75rem;
  cursor: pointer;
  transition: all 0.15s;
}

.preset-chip:hover {
  color: var(--accent-cyan);
  border-color: var(--accent-cyan);
  background: rgba(0, 229, 255, 0.05);
}

/* Filter Bar */
.env-filter-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.65rem 0.85rem;
  border-radius: 10px;
  gap: 1rem;
}

.search-input-wrapper {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex: 1;
  max-width: 450px;
}

.search-icon {
  color: var(--text-muted);
  display: flex;
  align-items: center;
}

.env-search-input {
  width: 100%;
  padding: 0.45rem 0.75rem;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  border-radius: 8px;
  color: var(--text-main);
  font-size: 0.82rem;
  outline: none;
  transition: all 0.2s;
}

.env-search-input:focus {
  border-color: var(--accent-cyan);
  box-shadow: 0 0 0 2px var(--accent-cyan-glow);
}

.env-stats-badges {
  display: flex;
  align-items: center;
  gap: 0.65rem;
}

.stat-badge {
  font-size: 0.78rem;
  color: var(--text-muted);
  background: var(--input-bg);
  padding: 0.3rem 0.65rem;
  border-radius: 6px;
  border: 1px solid var(--card-border);
}

.stat-badge.sensitive-stat {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  color: var(--warning, #f59e0b);
  background: rgba(245, 158, 11, 0.1);
  border-color: rgba(245, 158, 11, 0.25);
}

.tool-action-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  border-radius: 6px;
  font-size: 0.82rem;
  font-weight: 600;
  cursor: pointer;
  border: 1px solid transparent;
  transition: all 0.2s;
}

.tool-action-btn.primary {
  background: var(--accent-cyan);
  color: #050b14;
}

.tool-action-btn.secondary {
  background: rgba(255, 255, 255, 0.08);
  border-color: var(--card-border, rgba(255, 255, 255, 0.15));
  color: var(--text-main);
}

.tool-action-btn.danger {
  background: #ef4444;
  color: #ffffff;
}

.tool-action-btn.small {
  padding: 5px 12px;
  font-size: 0.78rem;
}

.tool-action-btn:hover:not(:disabled) {
  filter: brightness(1.1);
  transform: translateY(-1px);
}

/* Env Grid & Cards */
.env-table-container {
  padding: 1.25rem;
  border-radius: 12px;
  min-height: 240px;
}

.env-empty-box {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3rem 1rem;
  text-align: center;
  gap: 8px;
}

.empty-title {
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-main);
}

.empty-desc {
  font-size: 0.82rem;
  color: var(--text-muted);
  max-width: 480px;
  line-height: 1.5;
  margin: 0;
}

.env-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(360px, 1fr));
  gap: 14px;
}

.env-card {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  padding: 12px 14px;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid var(--card-border, rgba(255, 255, 255, 0.1));
  border-radius: 10px;
  gap: 10px;
  transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
}

.env-card:hover {
  border-color: var(--accent-cyan);
  box-shadow: 0 4px 16px rgba(0, 229, 255, 0.1);
  transform: translateY(-1px);
}

.env-card.sensitive {
  border-left: 3px solid #f59e0b;
}

.env-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.env-key-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.env-key {
  font-size: 0.95rem;
  font-weight: 700;
  color: var(--text-main);
  letter-spacing: 0.5px;
}

.key-type-icon {
  flex-shrink: 0;
}

.sec-tag {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 0.68rem;
  padding: 2px 8px;
  border-radius: 9999px;
  font-weight: 600;
}

.sec-tag.sensitive {
  background: rgba(245, 158, 11, 0.12);
  color: var(--warning, #f59e0b);
  border: 1px solid rgba(245, 158, 11, 0.25);
}

.sec-tag.public {
  background: rgba(2, 132, 199, 0.12);
  color: var(--accent-cyan);
  border: 1px solid rgba(2, 132, 199, 0.25);
}

.sec-tag.override {
  background: rgba(16, 185, 129, 0.15);
  color: #10b981;
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.sec-tag.profile-only {
  background: rgba(59, 130, 246, 0.15);
  color: #60a5fa;
  border: 1px solid rgba(59, 130, 246, 0.3);
}

.sec-tag.inherited {
  background: rgba(148, 163, 184, 0.12);
  color: var(--text-muted);
  border: 1px solid rgba(148, 163, 184, 0.2);
}

.scope-tag {
  font-size: 0.68rem;
  padding: 2px 8px;
  border-radius: 9999px;
  font-weight: 600;
  background: rgba(2, 132, 199, 0.12);
  color: var(--accent-cyan);
  border: 1px solid rgba(2, 132, 199, 0.25);
}

.scope-tag.user {
  background: rgba(168, 85, 247, 0.12);
  color: #c084fc;
  border-color: rgba(168, 85, 247, 0.25);
}

.card-btn-group {
  display: flex;
  align-items: center;
  gap: 4px;
}

.action-btn {
  background: transparent;
  border: 1px solid transparent;
  color: var(--text-muted);
  padding: 4px 6px;
  border-radius: 5px;
  cursor: pointer;
  transition: all 0.15s;
}

.action-btn.restore-btn {
  color: #10b981;
  font-weight: bold;
}

.action-btn.restore-btn:hover {
  background: rgba(16, 185, 129, 0.15);
  color: #34d399;
}

.action-btn:hover {
  background: rgba(255, 255, 255, 0.08);
  color: var(--text-main);
}

.action-btn.delete-btn:hover {
  color: #ef4444;
  background: rgba(239, 68, 68, 0.12);
}

/* Value row & masking */
.env-card-value-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  background: var(--input-bg);
  padding: 6px 10px;
  border-radius: 6px;
  border: 1px solid var(--card-border);
}

.value-display-box {
  flex: 1;
  font-size: 0.84rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--text-main);
}

.value-display-box.masked .masked-dots {
  letter-spacing: 2px;
  color: var(--text-muted);
  opacity: 0.8;
  font-size: 0.95rem;
}

.value-actions {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
}

.value-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 3px 7px;
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid var(--card-border, rgba(255, 255, 255, 0.1));
  border-radius: 5px;
  color: var(--text-muted);
  font-size: 0.72rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s;
}

.value-btn:hover {
  color: var(--accent-cyan);
  border-color: var(--accent-cyan);
  background: rgba(0, 229, 255, 0.08);
}

.value-btn.eye-btn.active {
  color: #f59e0b;
  border-color: rgba(245, 158, 11, 0.4);
  background: rgba(245, 158, 11, 0.12);
}

.copy-check {
  color: #10b981;
  font-weight: 700;
}

/* View Mode Switcher Toggle */
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

/* High-Density List / Table View */
.env-list-table-wrapper {
  width: 100%;
  overflow-x: auto;
  border: 1px solid var(--card-border);
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.015);
}

.env-table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
  font-size: 0.84rem;
}

.env-table thead tr {
  background: rgba(255, 255, 255, 0.03);
  border-bottom: 1px solid var(--card-border);
}

.env-table th {
  padding: 0.75rem 1rem;
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: var(--text-muted);
  white-space: nowrap;
}

.env-table th.th-actions {
  text-align: right;
  width: 90px;
}

.env-table-row {
  border-bottom: 1px solid var(--card-border, rgba(255, 255, 255, 0.05));
  transition: background 0.15s ease;
}

.env-table-row:last-child {
  border-bottom: none;
}

.env-table-row:hover {
  background: rgba(255, 255, 255, 0.035);
}

.env-table-row.sensitive {
  border-left: 3px solid #f59e0b;
}

.env-table td {
  padding: 0.65rem 1rem;
  vertical-align: middle;
}

.env-table td.td-actions {
  text-align: right;
}

.table-key-cell {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.table-key-cell .key-name {
  font-weight: 600;
  color: var(--text-main);
  letter-spacing: 0.4px;
}

.sec-tag.mini {
  padding: 1px 5px;
  font-size: 0.68rem;
  gap: 3px;
}

.table-value-cell {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  border-radius: 6px;
  padding: 4px 10px;
  max-width: 460px;
  min-height: 28px;
  box-sizing: border-box;
}

.table-value-text {
  flex: 1;
  font-size: 0.82rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--text-main);
}

.table-value-text.masked .masked-dots {
  letter-spacing: 2px;
  color: var(--text-muted);
  opacity: 0.8;
  font-size: 0.92rem;
}

.table-value-actions {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  flex-shrink: 0;
}

.value-btn.mini {
  padding: 2px 6px;
  font-size: 0.72rem;
  gap: 2px;
}

.table-row-actions {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  justify-content: flex-end;
}

/* Delete confirmation modal */
.confirm-overlay {
  position: fixed;
  inset: 0;
  z-index: 1000;
  background: rgba(0, 0, 0, 0.65);
  backdrop-filter: blur(8px);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
}

.confirm-modal {
  width: 100%;
  max-width: 440px;
  padding: 1.5rem;
  border-radius: 12px;
}

.confirm-header {
  display: flex;
  align-items: center;
  gap: 10px;
  color: #ef4444;
}

.confirm-header h4 {
  margin: 0;
  font-size: 1.05rem;
  font-weight: 600;
  color: var(--text-main);
}

.confirm-desc {
  font-size: 0.85rem;
  color: var(--text-muted);
  line-height: 1.5;
  margin: 12px 0 16px;
}

.confirm-desc code {
  color: var(--accent-cyan);
  background: rgba(0, 229, 255, 0.1);
  padding: 2px 6px;
  border-radius: 4px;
}

.confirm-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 10px;
}

/* Floating Toast */
.toast-banner {
  position: fixed;
  bottom: 2rem;
  right: 2rem;
  z-index: 2000;
  background: #0f172a;
  color: #f8fafc;
  border: 1px solid var(--accent-cyan);
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.4);
  border-radius: 8px;
  padding: 10px 18px;
  font-size: 0.86rem;
  font-weight: 600;
}

.toast-fade-enter-active,
.toast-fade-leave-active {
  transition: all 0.25s ease;
}

.toast-fade-enter-from,
.toast-fade-leave-to {
  opacity: 0;
  transform: translateY(10px);
}

.spin-icon {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

/* Light Theme Overrides */
:root[data-theme="light"] .scope-badge.user,
[data-theme="light"] .scope-badge.user,
:root[data-theme="light"] .scope-tag.user,
[data-theme="light"] .scope-tag.user {
  background: rgba(124, 58, 237, 0.1);
  color: #7c3aed;
  border-color: rgba(124, 58, 237, 0.25);
}

:root[data-theme="light"] .scope-badge,
[data-theme="light"] .scope-badge,
:root[data-theme="light"] .scope-tag,
[data-theme="light"] .scope-tag,
:root[data-theme="light"] .sec-tag.public,
[data-theme="light"] .sec-tag.public {
  background: rgba(2, 132, 199, 0.12);
  color: #0284c7;
  border-color: rgba(2, 132, 199, 0.25);
}

:root[data-theme="light"] .tool-action-btn.primary,
[data-theme="light"] .tool-action-btn.primary {
  background: linear-gradient(135deg, #0284c7, #0369a1);
  color: #ffffff;
  border-color: #0284c7;
}

:root[data-theme="light"] .env-card,
[data-theme="light"] .env-card {
  background: #ffffff;
  border-color: #e2e8f0;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.04);
}

:root[data-theme="light"] .env-card:hover,
[data-theme="light"] .env-card:hover {
  border-color: #0284c7;
  box-shadow: 0 4px 14px rgba(2, 132, 199, 0.15);
}

:root[data-theme="light"] .env-card-value-row,
[data-theme="light"] .env-card-value-row,
:root[data-theme="light"] .table-value-cell,
[data-theme="light"] .table-value-cell {
  background: #f8fafc;
  border-color: #e2e8f0;
}

:root[data-theme="light"] .value-display-box,
[data-theme="light"] .value-display-box,
:root[data-theme="light"] .table-value-text,
[data-theme="light"] .table-value-text {
  color: #0f172a;
}

:root[data-theme="light"] .value-btn,
[data-theme="light"] .value-btn {
  background: #ffffff;
  border-color: #e2e8f0;
  color: #475569;
}

:root[data-theme="light"] .value-btn:hover,
[data-theme="light"] .value-btn:hover {
  background: rgba(2, 132, 199, 0.08);
  border-color: #0284c7;
  color: #0284c7;
}

:root[data-theme="light"] .env-list-table-wrapper,
[data-theme="light"] .env-list-table-wrapper {
  background: #ffffff;
  border-color: #e2e8f0;
}

:root[data-theme="light"] .env-table thead tr,
[data-theme="light"] .env-table thead tr {
  background: #f8fafc;
  border-bottom-color: #e2e8f0;
}

:root[data-theme="light"] .env-table th,
[data-theme="light"] .env-table th {
  color: #64748b;
}

:root[data-theme="light"] .env-table-row,
[data-theme="light"] .env-table-row {
  border-bottom-color: #f1f5f9;
}

:root[data-theme="light"] .env-table-row:hover,
[data-theme="light"] .env-table-row:hover {
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

:root[data-theme="light"] .env-search-input,
[data-theme="light"] .env-search-input {
  background: #f8fafc;
  border-color: #cbd5e1;
  color: #0f172a;
}

:root[data-theme="light"] .toast-banner,
[data-theme="light"] .toast-banner {
  background: #ffffff;
  color: #0f172a;
  border-color: #0284c7;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
}
</style>
