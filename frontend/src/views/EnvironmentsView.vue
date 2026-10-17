<template>
  <div class="view-container environments-view" id="view-environments">
    <!-- View Header -->
    <div class="view-header-row">
      <div>
        <h2 class="view-title" style="display: flex; align-items: center; gap: 8px;">
          <TechIcon id="workbench:environments" :size="20" />
          <span>{{ t("environments.title") || "环境变量管理" }}</span>
        </h2>
        <p class="view-desc">
          {{ t("environments.desc") || "统一管理工作区 (.unirtm.toml) 与用户全局 (~/.config/unirtm/unirtm.toml) 配置文件中的 [env] 环境变量。敏感密钥默认脱密遮蔽。" }}
        </p>
      </div>
      <div class="view-actions" style="display: flex; gap: 8px; align-items: center;">
        <button class="tool-action-btn secondary" @click="fetchEnvVars" :disabled="loading" title="刷新环境变量">
          <TechIcon id="workbench:sync" :size="14" :class="{ 'spin-icon': loading }" />
          <span>{{ loading ? (t("common.refreshing") || "刷新中...") : (t("common.refresh") || "刷新") }}</span>
        </button>
      </div>
    </div>

    <!-- Current Scope Context Banner -->
    <div class="scope-selection-bar glass-card" id="environments-scope-bar">
      <div class="scope-context-pill">
        <span class="scope-badge" :class="activeScope">
          <TechIcon :id="activeScope === 'user' ? 'workbench:settings' : 'workbench:folder'" :size="13" :colored="false" />
          {{ activeScope === 'user' ? (t("environments.scope_user") || "用户") : (t("environments.scope_workspace") || "工作区") }}
        </span>
        <span class="scope-hint-text">
          {{ activeScope === 'user'
              ? (t("environments.scope_user_desc") || "定义在用户全局 ~/.config/unirtm/unirtm.toml 的 [env] 配置块中")
              : (t("environments.scope_workspace_desc") || "定义在当前项目 .unirtm.toml 的 [env] 配置块中") }}
        </span>
      </div>
      <div class="scope-context-config font-mono" :title="activeScope === 'user' ? '~/.config/unirtm/unirtm.toml' : '.unirtm.toml'">
        <span class="file-icon">📄</span>
        <span>{{ activeScope === 'user' ? '~/.config/unirtm/unirtm.toml' : '.unirtm.toml' }}</span>
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
          placeholder="搜索环境变量名称、关键字或值..."
        />
      </div>


      <div class="env-stats-badges">
        <span class="stat-badge">
          <b>{{ filteredEnvList.length }}</b> 个变量
        </span>
        <span v-if="sensitiveCount > 0" class="stat-badge sensitive-stat" title="已自动启用默认脱密保护">
          <TechIcon id="workbench:lock" :size="12" />
          <b>{{ sensitiveCount }}</b> 个敏感
        </span>
        <button
          class="tool-action-btn primary small"
          id="btn-add-env"
          @click="openCreateModal"
          title="新建环境变量"
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
        <span style="font-size: 0.85rem; color: var(--text-muted);">正在读取 [env] 配置...</span>
      </div>

      <!-- Empty State -->
      <div v-else-if="filteredEnvList.length === 0" class="env-empty-box">
        <TechIcon id="workbench:environments" :size="32" :colored="false" />
        <div class="empty-title">
          {{ searchKey ? `未找到匹配 "${searchKey}" 的环境变量` : (t("environments.empty_title") || "暂未配置任何环境变量") }}
        </div>
        <p class="empty-desc">
          {{ activeScope === 'user'
              ? '可在用户全局配置 (~/.config/unirtm/unirtm.toml) 中添加 [env] 区块，为全局所有开发环境提供兜底环境变量。'
              : '可在当前项目根目录的 .unirtm.toml 中添加 [env] 区块，锁定当前项目的专属环境变量与服务端口。' }}
        </p>
        <button v-if="!searchKey" class="tool-action-btn primary" style="margin-top: 0.75rem;" @click="openCreateModal">
          <TechIcon id="workbench:edit" :size="14" />
          <span>{{ t("environments.create_first") || "创建第一个环境变量" }}</span>
        </button>
      </div>

      <!-- Environments Grid / Cards -->
      <div v-else class="env-grid">
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
                :title="item.is_sensitive ? '敏感字段：默认进行脱密遮掩' : '普通公开环境变量'"
              >
                <TechIcon :id="item.is_sensitive ? 'workbench:lock' : 'workbench:globe'" :size="11" />
                <span>{{ item.is_sensitive ? (t('environments.sensitive_badge') || '敏感密钥') : (t('environments.public_badge') || '公开变量') }}</span>
              </span>
              <span class="scope-tag" :class="item.scope">
                {{ item.scope === 'user' ? (t("environments.scope_user") || "用户") : (t("environments.scope_workspace") || "工作区") }}
              </span>
            </div>

            <!-- Card Actions: Edit & Delete -->
            <div class="card-btn-group">
              <button
                class="action-btn edit-btn"
                @click="openEditModal(item)"
                :title="t('environments.edit_btn') || '编辑'"
              >
                <TechIcon id="workbench:edit" :size="14" />
              </button>
              <button
                class="action-btn delete-btn"
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
              <span v-if="item.is_sensitive && !item.isRevealed" class="masked-dots">
                ••••••••••••••••
              </span>
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
                :title="item.isRevealed ? (t('environments.conceal') || '脱密隐藏') : (t('environments.reveal') || '显示明文')"
              >
                <TechIcon :id="item.isRevealed ? 'workbench:eye-closed' : 'workbench:eye'" :size="14" />
                <span class="btn-text-sm">{{ item.isRevealed ? '密文' : '明文' }}</span>
              </button>

              <!-- One-click Copy Button -->
              <button
                class="value-btn copy-btn"
                @click="copyValue(item)"
                :title="t('environments.copy_val') || '复制真实值'"
              >
                <TechIcon v-if="copiedKey === item.name" id="workbench:success" :size="13" class="copy-check" />
                <TechIcon v-else id="workbench:copy" :size="13" />
                <span class="btn-text-sm">{{ copiedKey === item.name ? '已复制' : '复制' }}</span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Create / Edit Environment Modal -->
    <EnvironmentModal
      :is-open="isModalOpen"
      :mode="modalMode"
      :initial-data="selectedEnv"
      :active-scope="activeScope"
      @close="isModalOpen = false"
      @save="handleSaveEnv"
    />

    <!-- Delete Confirmation Modal -->
    <div v-if="deleteConfirmKey" class="confirm-overlay" @click.self="deleteConfirmKey = ''">
      <div class="confirm-modal glass-card">
        <div class="confirm-header">
          <TechIcon id="workbench:warning" :size="22" />
          <h4>确认删除环境变量</h4>
        </div>
        <p class="confirm-desc">
          确定要从 <b>{{ activeScope === 'user' ? '用户全局配置' : '当前工作区配置' }}</b> 中彻底删除环境变量
          <code class="font-mono">{{ deleteConfirmKey }}</code> 吗？
          删除后将不可恢复。
        </p>
        <div class="confirm-actions">
          <button class="tool-action-btn secondary" @click="deleteConfirmKey = ''">取消</button>
          <button class="tool-action-btn danger" @click="executeDelete(deleteConfirmKey)">确认删除</button>
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
  is_sensitive: boolean;
  isRevealed?: boolean;
}

const route = useRoute();
const { activeScope, syncFromRoute } = useAppScope();

const envList = ref<EnvItem[]>([]);
const loading = ref(false);
const searchKey = ref("");
const isModalOpen = ref(false);
const modalMode = ref<"create" | "edit">("create");
const selectedEnv = ref<{ name: string; value: string; scope: "workspace" | "user" } | undefined>(undefined);
const deleteConfirmKey = ref("");
const copiedKey = ref("");
const toastMessage = ref("");

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
    fetchEnvVars();
  }
);

async function fetchEnvVars() {
  loading.value = true;
  try {
    const res = await fetch(`/api/env?scope=${activeScope.value}`);
    if (res.ok) {
      const data = await res.json();
      if (Array.isArray(data.env)) {
        envList.value = data.env.map((e: any) => ({
          name: e.name,
          value: e.value,
          scope: e.scope || activeScope.value,
          is_sensitive: Boolean(e.is_sensitive),
          isRevealed: false,
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
  if (!searchKey.value) return envList.value;
  const q = searchKey.value.toLowerCase().trim();
  return envList.value.filter(
    (e) => e.name.toLowerCase().includes(q) || e.value.toLowerCase().includes(q)
  );
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
    showToast(`✓ 已复制 ${item.name} 真实值到剪贴板`);
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
}) {
  try {
    const res = await fetch("/api/env", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    if (res.ok) {
      isModalOpen.value = false;
      showToast(modalMode.value === "create" ? `✓ 成功创建环境变量 ${payload.key}` : `✓ 成功更新环境变量 ${payload.key}`);
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
  try {
    const res = await fetch(`/api/env?key=${encodeURIComponent(name)}&scope=${activeScope.value}`, {
      method: "DELETE",
    });
    if (res.ok) {
      deleteConfirmKey.value = "";
      showToast(`✓ 已成功删除环境变量 ${name}`);
      await fetchEnvVars();
    } else {
      const err = await res.json();
      showToast("删除失败: " + (err.error || "未知错误"));
    }
  } catch (err: any) {
    showToast("网络删除错误: " + err.message);
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
  background: rgba(0, 0, 0, 0.2);
  padding: 6px 10px;
  border-radius: 6px;
  border: 1px solid rgba(255, 255, 255, 0.05);
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
[data-theme="light"] .env-card-value-row {
  background: #f8fafc;
  border-color: #e2e8f0;
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
