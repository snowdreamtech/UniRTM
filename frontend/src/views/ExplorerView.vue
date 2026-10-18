<template>
  <div class="view-container explorer-view" id="view-explorer">
    <!-- 1. Tabs Bar (Displayed when one or more tabs are open) -->
    <div v-if="tabs.length > 0" class="editor-tabs-bar">
      <div class="tabs-scroll-container">
        <div
          v-for="tab in tabs"
          :key="tab.id"
          class="editor-tab"
          :class="{ active: tab.id === activeTabId }"
          @click="handleSelectTab(tab.id)"
          @auxclick.stop="closeTab(tab.id)"
          :title="tab.path"
        >
          <TechIcon :id="tab.name" :size="16" class="tab-icon" />
          <span class="tab-title">{{ tab.name }}</span>
          <span v-if="tab.isDirty" class="dirty-dot" title="未保存更改">●</span>
          <button
            class="tab-close-btn"
            title="关闭 (中键亦可关闭)"
            @click.stop="closeTab(tab.id)"
          >
            <span class="codicon codicon-close"></span>
          </button>
        </div>
      </div>

      <div class="tab-actions-right">
        <button
          v-if="activeTab?.isDirty"
          class="action-btn primary"
          title="保存当前文件 (Ctrl+S)"
          @click="handleSave"
        >
          <span class="codicon codicon-save"></span>
          <span>保存</span>
        </button>
        <button
          v-if="hasDirtyTabs"
          class="action-btn"
          title="全部保存"
          @click="saveAllTabs"
        >
          <span class="codicon codicon-save-all"></span>
        </button>
        <button
          v-if="activeTab"
          class="action-btn"
          title="复制相对路径"
          @click="handleCopyPath"
        >
          <span class="codicon codicon-copy"></span>
        </button>
        <button
          v-if="activeTab"
          class="action-btn"
          title="在集成终端中打开所在目录"
          @click="handleOpenParentTerminal"
        >
          <span class="codicon codicon-terminal"></span>
        </button>
        <button
          class="action-btn"
          title="关闭全部标签页"
          @click="closeAllTabs"
        >
          <span class="codicon codicon-close-all"></span>
        </button>
      </div>
    </div>

    <!-- 2. Active Tab Workspace (Code Editor or Image Preview) -->
    <div v-if="activeTab" class="editor-workspace">
      <!-- File Breadcrumb Bar -->
      <div class="editor-breadcrumb">
        <span class="codicon codicon-folder"></span>
        <span class="bc-path">{{ activeTab.path }}</span>
      </div>

      <!-- Editor Content Area -->
      <div class="editor-content-area">
        <!-- Image Preview -->
        <div v-if="isImageFile(activeTab.name)" class="image-preview-container">
          <div class="image-wrapper">
            <img :src="imagePreviewSrc" :alt="activeTab.name" class="preview-img" />
          </div>
          <div class="image-meta">
            <span>{{ activeTab.name }}</span>
          </div>
        </div>

        <!-- Monaco Code Editor -->
        <MonacoEditor
          v-else
          :tab="activeTab"
          @change="handleEditorChange"
          @cursor-change="handleCursorChange"
          @save="handleSave"
        />
      </div>

      <!-- Editor Footer Status Bar -->
      <div class="editor-footer">
        <div class="footer-left">
          <span class="footer-item">UTF-8</span>
          <span class="footer-item">LF</span>
          <span class="footer-item language-tag">{{ activeTabLanguage }}</span>
        </div>
        <div class="footer-right">
          <span class="footer-item">行 {{ cursorLine }}, 列 {{ cursorColumn }}</span>
        </div>
      </div>
    </div>

    <!-- 3. Active Directory Selected (No active tab) -->
    <div v-else-if="selectedNode && selectedNode.is_dir" class="folder-overview-workspace">
      <div class="folder-hero-card">
        <div class="folder-icon-large">
          <TechIcon id="codicon:folder-opened" :size="48" />
        </div>
        <div class="folder-info">
          <h2 class="folder-title">{{ selectedNode.name || "工作区根目录" }}</h2>
          <p class="folder-path">{{ selectedNode.path || "/" }}</p>
          <div class="folder-stats">
            <span class="badge">{{ selectedNode.children?.length || 0 }} 个项目</span>
            <span v-if="selectedNode.mod_time" class="mod-time">
              修改于: {{ new Date(selectedNode.mod_time).toLocaleString() }}
            </span>
          </div>
        </div>
      </div>

      <div class="folder-quick-actions">
        <h4 class="section-heading">快捷操作</h4>
        <div class="action-grid">
          <div class="action-card" @click="handleCreateFileInside">
            <span class="codicon codicon-new-file card-icon"></span>
            <div class="card-text">
              <span class="card-title">新建文件</span>
              <span class="card-desc">在此文件夹下创建新文件</span>
            </div>
          </div>
          <div class="action-card" @click="handleCreateFolderInside">
            <span class="codicon codicon-new-folder card-icon"></span>
            <div class="card-text">
              <span class="card-title">新建子文件夹</span>
              <span class="card-desc">创建嵌套子目录</span>
            </div>
          </div>
          <div class="action-card" @click="handleOpenFolderInTerminal">
            <span class="codicon codicon-terminal card-icon"></span>
            <div class="card-text">
              <span class="card-title">在终端中打开</span>
              <span class="card-desc">在下方终端中导航至此目录</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 4. No File/Folder Selected (Watermark Welcome Screen) -->
    <div v-else class="explorer-welcome-workspace">
      <div class="welcome-watermark">
        <TechIcon id="workbench:explorer" :size="56" class="watermark-icon" />
        <h2 class="welcome-title">{{ t("explorer.title") || "资源管理器" }}</h2>
        <p class="welcome-desc">
          在左侧选择文件以查看详情或编辑，支持文件浏览、增删改查与终端快捷联动。
        </p>

        <div class="welcome-shortcuts">
          <div class="shortcut-list">
            <div class="shortcut-row">
              <span class="key-badge">单击文件</span>
              <span class="shortcut-desc">打开文件代码编辑器标签页</span>
            </div>
            <div class="shortcut-row">
              <span class="key-badge">中键 / ✕</span>
              <span class="shortcut-desc">关闭当前文件标签页</span>
            </div>
            <div class="shortcut-row">
              <span class="key-badge">双击 / 箭头</span>
              <span class="shortcut-desc">展开或折叠子目录</span>
            </div>
            <div class="shortcut-row">
              <span class="key-badge">右键</span>
              <span class="shortcut-desc">弹出上下文菜单 (重命名、删除、在终端打开)</span>
            </div>
            <div class="shortcut-row">
              <span class="key-badge">F2</span>
              <span class="shortcut-desc">就地原地重命名文件或文件夹</span>
            </div>
            <div class="shortcut-row">
              <span class="key-badge">Ctrl + S</span>
              <span class="shortcut-desc">快速保存当前正在编辑的文件</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import { t } from "../i18n";
import TechIcon from "../components/common/TechIcon.vue";
import MonacoEditor from "../components/editor/MonacoEditor.vue";
import { useFileTree } from "../composables/useFileTree";
import { useEditorTabs } from "../composables/useEditorTabs";

const emit = defineEmits<{
  (e: "open-drawer", tab: string): void;
}>();

const {
  selectedNode,
  copyRelativePath,
  startInlineCreate,
} = useFileTree();

const {
  tabs,
  activeTabId,
  activeTab,
  hasDirtyTabs,
  switchTab,
  updateContent,
  saveTab,
  saveAllTabs,
  closeTab,
  closeAllTabs,
} = useEditorTabs();

const cursorLine = ref<number>(1);
const cursorColumn = ref<number>(1);

const activeTabLanguage = computed(() => {
  return (activeTab.value?.language || "plaintext").toUpperCase();
});

function isImageFile(filename: string): boolean {
  return /\.(png|jpe?g|gif|svg|webp|ico|bmp)$/i.test(filename);
}

const imagePreviewSrc = computed(() => {
  if (!activeTab.value) return "";
  return `/api/fs/read?path=${encodeURIComponent(activeTab.value.path)}`;
});

function handleSelectTab(tabId: string) {
  switchTab(tabId);
}

function handleEditorChange(newContent: string) {
  if (activeTab.value) {
    updateContent(activeTab.value.id, newContent);
  }
}

function handleCursorChange(line: number, col: number) {
  cursorLine.value = line;
  cursorColumn.value = col;
}

async function handleSave() {
  await saveTab();
}

function handleCopyPath() {
  if (activeTab.value) {
    copyRelativePath(activeTab.value.path);
  }
}

function handleOpenParentTerminal() {
  emit("open-drawer", "terminal");
}

function handleOpenFolderInTerminal() {
  emit("open-drawer", "terminal");
}

function handleCreateFileInside() {
  if (selectedNode.value?.is_dir) {
    startInlineCreate(selectedNode.value.path, "new-file");
  }
}

function handleCreateFolderInside() {
  if (selectedNode.value?.is_dir) {
    startInlineCreate(selectedNode.value.path, "new-folder");
  }
}
</script>

<style scoped>
.explorer-view {
  display: flex;
  flex-direction: column;
  height: 100%;
  width: 100%;
  padding: 0;
  margin: 0;
  box-sizing: border-box;
  background-color: var(--bg-color);
  color: var(--text-main);
  overflow: hidden;
}

/* 1. Editor Tabs Bar */
.editor-tabs-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background-color: var(--card-bg);
  height: 35px;
  border-bottom: 1px solid var(--card-border);
  flex-shrink: 0;
  overflow: hidden;
}

.tabs-scroll-container {
  display: flex;
  align-items: center;
  height: 100%;
  overflow-x: auto;
  overflow-y: hidden;
  scrollbar-width: none;
}

.tabs-scroll-container::-webkit-scrollbar {
  display: none;
}

.editor-tab {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 0 12px;
  height: 100%;
  background-color: var(--card-bg);
  color: var(--text-muted);
  border-right: 1px solid var(--card-border);
  font-size: 13px;
  cursor: pointer;
  user-select: none;
  white-space: nowrap;
  border-top: 2px solid transparent;
  transition: all 0.15s ease;
}

.editor-tab:hover {
  background-color: var(--bg-color);
  color: var(--text-main);
}

.editor-tab.active {
  background-color: var(--bg-color);
  color: var(--text-main);
  border-top: 2px solid var(--accent-cyan);
}

.tab-icon {
  margin-right: 2px;
}

.tab-title {
  max-width: 160px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.dirty-dot {
  color: var(--accent-cyan);
  font-size: 12px;
  line-height: 1;
}

.tab-close-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 2px;
  border-radius: 3px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
}

.tab-close-btn:hover {
  background-color: rgba(128, 128, 128, 0.2);
  color: var(--text-main);
}

.tab-actions-right {
  display: flex;
  align-items: center;
  gap: 4px;
  padding-right: 8px;
  flex-shrink: 0;
}

.action-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 4px 6px;
  border-radius: 3px;
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  transition: all 0.15s ease;
}

.action-btn:hover {
  background-color: rgba(128, 128, 128, 0.15);
  color: var(--text-main);
}

.action-btn.primary {
  background-color: var(--accent-cyan);
  color: #ffffff;
  padding: 3px 8px;
  font-weight: 500;
}

.action-btn.primary:hover {
  filter: brightness(1.1);
}

/* 2. Editor Workspace */
.editor-workspace {
  display: flex;
  flex-direction: column;
  flex: 1;
  width: 100%;
  height: 100%;
  min-height: 0;
  min-width: 0;
  overflow: hidden;
}

.editor-breadcrumb {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 16px;
  background-color: var(--section-bg, var(--bg-color));
  color: var(--text-muted);
  font-size: 11px;
  border-bottom: 1px solid var(--card-border);
  flex-shrink: 0;
}

.editor-content-area {
  flex: 1;
  position: relative;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 100%;
  min-height: 0;
  min-width: 0;
}

.image-preview-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  padding: 24px;
}

.image-wrapper {
  max-width: 80%;
  max-height: 70%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: repeating-conic-gradient(var(--card-bg) 0% 25%, var(--bg-color) 0% 50%) 50% / 16px 16px;
  border: 1px solid var(--card-border);
  border-radius: 4px;
  padding: 12px;
}

.preview-img {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}

.image-meta {
  margin-top: 12px;
  display: flex;
  gap: 16px;
  color: var(--text-muted);
  font-size: 12px;
}

.editor-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 24px;
  padding: 0 12px;
  background-color: var(--card-bg);
  border-top: 1px solid var(--card-border);
  color: var(--text-muted);
  font-size: 11px;
  flex-shrink: 0;
}

.footer-left,
.footer-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.footer-item {
  display: inline-flex;
  align-items: center;
}

.language-tag {
  color: var(--accent-cyan);
  font-weight: 500;
}

/* 3. Folder Overview */
.folder-overview-workspace {
  display: flex;
  flex-direction: column;
  padding: 40px;
  gap: 32px;
  max-width: 800px;
}

.folder-hero-card {
  display: flex;
  align-items: center;
  gap: 20px;
  padding: 24px;
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  border-radius: 8px;
}

.folder-icon-large {
  color: #dcb67a;
}

.folder-info {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.folder-title {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: var(--text-main);
}

.folder-path {
  margin: 0;
  font-size: 12px;
  color: var(--text-muted);
  font-family: monospace;
}

.folder-stats {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 4px;
}

.badge {
  background-color: var(--accent-cyan-glow, rgba(0, 229, 255, 0.15));
  color: var(--accent-cyan);
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;
}

.mod-time {
  font-size: 11px;
  color: var(--text-muted);
}

.folder-quick-actions {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.section-heading {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.action-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 12px;
}

.action-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 16px;
  background: var(--card-bg);
  border: 1px solid var(--card-border);
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.action-card:hover {
  background: var(--input-bg);
  border-color: var(--accent-cyan);
  transform: translateY(-1px);
}

.card-icon {
  font-size: 22px;
  color: var(--accent-cyan);
}

.card-text {
  display: flex;
  flex-direction: column;
}

.card-title {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-main);
}

.card-desc {
  font-size: 11px;
  color: var(--text-muted);
}

/* 4. Welcome Screen */
.explorer-welcome-workspace {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  width: 100%;
}

.welcome-watermark {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  max-width: 480px;
  padding: 24px;
}

.watermark-icon {
  color: var(--text-muted);
  margin-bottom: 12px;
  opacity: 0.35;
}

.welcome-title {
  margin: 0 0 6px;
  font-size: 18px;
  font-weight: 600;
  color: var(--text-main);
}

.welcome-desc {
  margin: 0 0 24px;
  font-size: 13px;
  line-height: 1.5;
  color: var(--text-muted);
}

.welcome-shortcuts {
  width: 100%;
  background: transparent;
  border: none;
  padding: 0;
}

.shortcut-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.shortcut-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 12px;
  padding: 4px 8px;
  border-radius: 4px;
  transition: background-color 0.15s ease;
}

.shortcut-row:hover {
  background-color: rgba(128, 128, 128, 0.08);
}

.key-badge {
  background: var(--input-bg);
  border: 1px solid var(--card-border);
  border-radius: 4px;
  padding: 2px 8px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 11px;
  color: var(--text-main);
}

.shortcut-desc {
  color: var(--text-muted);
}
</style>
