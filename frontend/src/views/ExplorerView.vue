<template>
  <div class="view-container explorer-view" id="view-explorer">
    <!-- 1. Active File Selected -->
    <div v-if="selectedNode && !selectedNode.is_dir" class="editor-workspace">
      <!-- VS Code Style Tab Bar -->
      <div class="editor-tabs-bar">
        <div class="editor-tab active">
          <TechIcon :id="selectedNode.name" :size="16" class="tab-icon" />
          <span class="tab-title">{{ selectedNode.name }}</span>
          <span v-if="isModified" class="dirty-dot" title="未保存更改">●</span>
          <button class="tab-close-btn" title="关闭" @click="closeActiveFile">
            <span class="codicon codicon-close"></span>
          </button>
        </div>

        <div class="tab-actions-right">
          <button
            v-if="isModified"
            class="action-btn primary"
            title="保存更改 (Ctrl+S)"
            @click="handleSave"
          >
            <span class="codicon codicon-save"></span>
            <span>保存</span>
          </button>
          <button
            class="action-btn"
            title="复制相对路径"
            @click="copyRelativePath(selectedNode.path)"
          >
            <span class="codicon codicon-copy"></span>
          </button>
          <button
            class="action-btn"
            title="在终端中打开所在目录"
            @click="handleOpenParentTerminal"
          >
            <span class="codicon codicon-terminal"></span>
          </button>
          <button
            class="action-btn danger"
            title="删除文件"
            @click="promptDelete(selectedNode)"
          >
            <span class="codicon codicon-trash"></span>
          </button>
        </div>
      </div>

      <!-- File Breadcrumb Bar -->
      <div class="editor-breadcrumb">
        <span class="codicon codicon-folder"></span>
        <span class="bc-path">{{ selectedNode.path }}</span>
      </div>

      <!-- Editor Content Area -->
      <div class="editor-content-area">
        <!-- Loading State -->
        <div v-if="isFileLoading" class="editor-loading">
          <span class="codicon codicon-loading codicon-modifier-spin"></span>
          <span>加载中...</span>
        </div>

        <!-- Image Preview -->
        <div v-else-if="isImageFile(selectedNode.name)" class="image-preview-container">
          <div class="image-wrapper">
            <img :src="imagePreviewSrc" :alt="selectedNode.name" class="preview-img" />
          </div>
          <div class="image-meta">
            <span>{{ selectedNode.name }}</span>
            <span>{{ formatFileSize(selectedNode.size) }}</span>
          </div>
        </div>

        <!-- Text / Code Editor Area -->
        <div v-else class="text-editor-container">
          <div class="line-numbers">
            <span v-for="n in lineCount" :key="n" class="line-num">{{ n }}</span>
          </div>
          <textarea
            v-model="editableContent"
            class="code-textarea"
            spellcheck="false"
            autocomplete="off"
            autocapitalize="off"
            @keydown="handleKeydown"
          ></textarea>
        </div>
      </div>

      <!-- Editor Footer Status Bar -->
      <div class="editor-footer">
        <div class="footer-left">
          <span class="footer-item">UTF-8</span>
          <span class="footer-item">LF</span>
          <span class="footer-item">{{ formatFileSize(selectedNode.size) }}</span>
        </div>
        <div class="footer-right">
          <span class="footer-item">行 {{ lineCount }}, 列 {{ cursorColumn }}</span>
        </div>
      </div>
    </div>

    <!-- 2. Active Directory Selected -->
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

    <!-- 3. No File/Folder Selected (VS Code Watermark Welcome Screen) -->
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
              <span class="key-badge">单击</span>
              <span class="shortcut-desc">选择文件并在主视图中查看详情</span>
            </div>
            <div class="shortcut-row">
              <span class="key-badge">双击 / 箭头</span>
              <span class="shortcut-desc">展开或折叠子目录</span>
            </div>
            <div class="shortcut-row">
              <span class="key-badge">右键</span>
              <span class="shortcut-desc">弹出上下文菜单 (重命名、删除、在终端中打开)</span>
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
import { ref, computed, watch } from "vue";
import { t } from "../i18n";
import TechIcon from "../components/common/TechIcon.vue";
import { useFileTree } from "../composables/useFileTree";

const emit = defineEmits<{
  (e: "open-drawer", tab: string): void;
}>();

const {
  selectedNode,
  activeFileContent,
  isFileLoading,
  saveFile,
  copyRelativePath,
  promptDelete,
  startInlineCreate,
} = useFileTree();

const editableContent = ref<string>("");
const cursorColumn = ref<number>(1);

watch(
  activeFileContent,
  (val) => {
    editableContent.value = val;
  },
  { immediate: true }
);

const isModified = computed(() => {
  return editableContent.value !== activeFileContent.value;
});

const lineCount = computed(() => {
  if (!editableContent.value) return 1;
  return editableContent.value.split("\n").length;
});

function isImageFile(filename: string): boolean {
  return /\.(png|jpe?g|gif|svg|webp|ico|bmp)$/i.test(filename);
}

const imagePreviewSrc = computed(() => {
  if (!selectedNode.value) return "";
  return `/api/fs/read?path=${encodeURIComponent(selectedNode.value.path)}`;
});

function formatFileSize(bytes: number): string {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
}

function closeActiveFile() {
  selectedNode.value = null;
  activeFileContent.value = "";
}

async function handleSave() {
  if (!selectedNode.value) return;
  await saveFile(selectedNode.value.path, editableContent.value);
}

function handleKeydown(e: KeyboardEvent) {
  if ((e.ctrlKey || e.metaKey) && e.key === "s") {
    e.preventDefault();
    handleSave();
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

/* 1. Editor Workspace */
.editor-workspace {
  display: flex;
  flex-direction: column;
  height: 100%;
  width: 100%;
}

.editor-tabs-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background-color: var(--card-bg);
  height: 35px;
  border-bottom: 1px solid var(--card-border);
}

.editor-tab {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 0 12px;
  height: 100%;
  background-color: var(--bg-color);
  color: var(--text-main);
  border-top: 2px solid var(--accent-cyan);
  border-right: 1px solid var(--card-border);
  font-size: 13px;
  cursor: pointer;
  user-select: none;
}

.tab-icon {
  margin-right: 2px;
}

.tab-title {
  max-width: 180px;
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
  background-color: rgba(128, 128, 128, 0.15);
  color: var(--text-main);
}

.tab-actions-right {
  display: flex;
  align-items: center;
  gap: 4px;
  padding-right: 8px;
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
}

.action-btn:hover {
  background-color: rgba(128, 128, 128, 0.12);
  color: var(--text-main);
}

.action-btn.primary {
  background-color: var(--accent-cyan);
  color: #ffffff;
  padding: 3px 8px;
}

.action-btn.primary:hover {
  filter: brightness(1.1);
}

.action-btn.danger:hover {
  color: var(--error, #ef4444);
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
}

.editor-content-area {
  flex: 1;
  position: relative;
  overflow: hidden;
  display: flex;
}

.editor-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  gap: 8px;
  color: var(--text-muted);
  font-size: 13px;
}

.text-editor-container {
  display: flex;
  width: 100%;
  height: 100%;
}

.line-numbers {
  display: flex;
  flex-direction: column;
  padding: 8px 12px;
  background-color: var(--bg-color);
  border-right: 1px solid var(--card-border);
  user-select: none;
  color: var(--text-muted);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
  font-size: 13px;
  line-height: 20px;
  text-align: right;
  min-width: 32px;
  opacity: 0.65;
}

.code-textarea {
  flex: 1;
  background-color: var(--bg-color);
  color: var(--text-main);
  border: none;
  outline: none;
  resize: none;
  padding: 8px 12px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
  font-size: 13px;
  line-height: 20px;
  white-space: pre;
  overflow: auto;
  tab-size: 2;
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

/* 2. Folder Overview */
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

/* 3. Welcome Screen (VS Code Style Watermark) */
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
