<template>
  <div class="view-container explorer-view" id="view-explorer">
    <!-- 1. Tabs Bar (Displayed when one or more tabs are open) -->
    <div v-if="tabs.length > 0" class="editor-tabs-bar">
      <div class="tabs-scroll-container">
        <div
          v-for="tab in tabs"
          :key="tab.id"
          class="editor-tab"
          :class="{
            active: tab.id === activeTabId,
            'is-pinned': tab.isPinned,
            'is-diff': tab.isDiff,
            'has-git-status': !!tab.gitStatus
          }"
          @click="handleSelectTab(tab.id)"
          @auxclick.middle.stop="closeTab(tab.id)"
          @contextmenu.prevent.stop="openTabContextMenu($event, tab)"
          :title="tab.path + (tab.isPinned ? ' (已固定)' : '') + (tab.gitStatus ? ` [${tab.gitStatus}]` : '')"
        >
          <span v-if="tab.isDiff" class="codicon codicon-diff tab-icon diff-icon"></span>
          <TechIcon v-else :id="tab.name" :size="16" class="tab-icon" />
          <span
            class="tab-title"
            :class="[
              tab.gitStatus ? 'git-color-' + tab.gitStatus : '',
              tab.isDiff ? 'tab-title-diff' : ''
            ]"
          >
            {{ tab.name }}
          </span>
          <span v-if="tab.gitStatus" class="tab-git-badge" :class="'badge-' + tab.gitStatus">{{ tab.gitStatus }}</span>
          <span v-if="tab.isDirty" class="dirty-dot" title="未保存更改">●</span>
          <button
            v-if="tab.isPinned"
            class="tab-pin-btn"
            title="取消固定"
            @click.stop="togglePinTab(tab.id)"
          >
            <span class="codicon codicon-pinned"></span>
          </button>
          <button
            v-else
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
          v-if="activeTab?.isDirty && !activeTab?.isBinary"
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
          @click="() => closeAllTabs()"
        >
          <span class="codicon codicon-close-all"></span>
        </button>
      </div>
    </div>

    <!-- 2. Active Tab Workspace (Code Editor or Image Preview) -->
    <div v-if="activeTab" class="editor-workspace">
      <!-- File Breadcrumbs Navigation Bar (VS Code aligned) -->
      <nav class="editor-breadcrumb" aria-label="文件导航路径">
        <div class="bc-items-scroll">
          <template v-for="(item, index) in breadcrumbs" :key="item.id">
            <span
              v-if="index > 0"
              class="codicon codicon-chevron-right bc-separator"
              aria-hidden="true"
            ></span>
            <button
              type="button"
              class="bc-item"
              :class="{
                'is-file': item.isFile,
                'is-dir': item.isDir,
                'is-root': item.isRoot,
                'is-active': index === breadcrumbs.length - 1,
              }"
              :title="item.path ? `${item.path} (单击在资源管理器中定位)` : item.name"
              @click="handleBreadcrumbClick(item)"
            >
              <span class="bc-icon-wrapper">
                <template v-if="item.isRoot">
                  <span class="codicon codicon-root-folder bc-icon"></span>
                </template>
                <template v-else-if="item.isDir">
                  <span class="codicon codicon-folder bc-icon"></span>
                </template>
                <template v-else>
                  <TechIcon :id="item.name" :size="14" class="bc-icon" />
                </template>
              </span>
              <span class="bc-label">{{ item.name }}</span>
            </button>
          </template>
        </div>
      </nav>

      <!-- Editor Content Area -->
      <div class="editor-content-area">
        <!-- Dual Pane Split View Layout -->
        <div
          v-if="splitState.active && splitState.tab"
          class="editor-split-container"
          :class="[
            'split-' + splitState.direction,
            'split-pos-' + splitState.position,
          ]"
        >
          <!-- Primary Pane (Active Tab) -->
          <div class="split-pane split-primary-pane">
            <div class="split-pane-header">
              <div class="split-pane-info">
                <TechIcon :id="activeTab.name" :size="14" class="split-pane-icon" />
                <span class="split-pane-name">{{ activeTab.name }}</span>
                <span v-if="activeTab.isDirty" class="dirty-dot">●</span>
              </div>
            </div>
            <div class="split-pane-body">
              <div v-if="isImageFile(activeTab.name)" class="image-preview-container">
                <div class="image-wrapper">
                  <img :src="imagePreviewSrc" :alt="activeTab.name" class="preview-img" />
                </div>
              </div>
              <div v-else-if="activeTab.isBinary" class="binary-file-container compact">
                <div class="binary-card">
                  <span class="codicon codicon-file-binary binary-hero-icon"></span>
                  <h4 class="binary-title">{{ activeTab.name }}</h4>
                  <p class="binary-message">无法在拆分视图中编辑二进制文件</p>
                </div>
              </div>
              <MonacoEditor
                v-else
                :key="'primary-' + activeTab.id"
                :tab="activeTab"
                @change="handleEditorChange"
                @cursor-change="handleCursorChange"
                @save="handleSave"
              />
            </div>
          </div>

          <!-- Split Sash / Divider -->
          <div class="split-divider" :class="'divider-' + splitState.direction"></div>

          <!-- Secondary Pane (Split Tab) -->
          <div class="split-pane split-secondary-pane">
            <div class="split-pane-header">
              <div class="split-pane-info">
                <TechIcon :id="splitState.tab.name" :size="14" class="split-pane-icon" />
                <span class="split-pane-name">{{ splitState.tab.name }}{{ splitState.inGroup ? ' (组内拆分)' : '' }}</span>
                <span v-if="splitState.tab.isDirty" class="dirty-dot">●</span>
              </div>
              <div class="split-pane-tools">
                <button
                  type="button"
                  class="split-tool-btn"
                  title="关闭拆分视图"
                  @click="closeSplitEditor"
                >
                  <span class="codicon codicon-close"></span>
                </button>
              </div>
            </div>
            <div class="split-pane-body">
              <div v-if="isImageFile(splitState.tab.name)" class="image-preview-container">
                <div class="image-wrapper">
                  <img
                    :src="'/api/fs/read?path=' + encodeURIComponent(splitState.tab.path)"
                    :alt="splitState.tab.name"
                    class="preview-img"
                  />
                </div>
              </div>
              <div v-else-if="splitState.tab.isBinary" class="binary-file-container compact">
                <div class="binary-card">
                  <span class="codicon codicon-file-binary binary-hero-icon"></span>
                  <h4 class="binary-title">{{ splitState.tab.name }}</h4>
                  <p class="binary-message">无法在拆分视图中编辑二进制文件</p>
                </div>
              </div>
              <MonacoEditor
                v-else
                :key="'split-' + splitState.tab.id"
                :tab="splitState.tab"
                @change="handleSplitEditorChange"
                @save="handleSplitEditorSave"
              />
            </div>
          </div>
        </div>

        <!-- Normal Single Editor Workspace -->
        <template v-else>
          <!-- 1. Image Preview -->
          <div v-if="isImageFile(activeTab.name)" class="image-preview-container">
            <div class="image-wrapper">
              <img :src="imagePreviewSrc" :alt="activeTab.name" class="preview-img" />
            </div>
            <div class="image-meta">
              <span>{{ activeTab.name }}</span>
            </div>
          </div>

          <!-- 2. Binary or Large File Notice (VS Code Aligned) -->
          <div v-else-if="activeTab.isBinary" class="binary-file-container">
            <div class="binary-card">
              <div class="binary-icon-box">
                <span class="codicon codicon-file-binary binary-hero-icon"></span>
              </div>
              <h3 class="binary-title">{{ activeTab.name }}</h3>
              <div class="binary-meta">
                <span v-if="activeTab.size" class="binary-badge">{{ formatFileSize(activeTab.size) }}</span>
                <span class="binary-badge tag">{{ activeTab.binaryReason === 'too_large' ? '大文件' : '二进制文件' }}</span>
              </div>
              <p class="binary-message">
                {{ activeTab.binaryReason === 'too_large' 
                   ? '该文件体积过大 (超过 10 MB)，无法在文本编辑器中显示。' 
                   : '该文件未在文本编辑器中显示，因为它是二进制文件或使用了不受支持的文本编码。' }}
              </p>
              <div class="binary-actions">
                <button type="button" class="action-btn-primary" @click="handleOpenParentTerminal">
                  <span class="codicon codicon-terminal"></span>
                  <span>在终端中查看 / 运行</span>
                </button>
                <button type="button" class="action-btn-secondary" @click="handleCopyPath">
                  <span class="codicon codicon-copy"></span>
                  <span>复制相对路径</span>
                </button>
              </div>
            </div>
          </div>

          <!-- 3. Monaco Diff Editor (Diff Tab) -->
          <MonacoDiffEditor
            v-else-if="activeTab.isDiff"
            :key="'diff-' + activeTab.id"
            :tab="activeTab"
          />

          <!-- 4. Monaco Code Editor (Normal Tab) -->
          <MonacoEditor
            v-else
            :key="activeTab.id"
            :tab="activeTab"
            @change="handleEditorChange"
            @cursor-change="handleCursorChange"
            @save="handleSave"
          />

        </template>
      </div>

      <!-- Editor Footer Status Bar -->
      <div class="editor-footer">
        <div class="footer-left">
          <template v-if="activeTab.isBinary">
            <span class="footer-item">BINARY</span>
            <span v-if="activeTab.size" class="footer-item">{{ formatFileSize(activeTab.size) }}</span>
          </template>
          <template v-else>
            <span class="footer-item">UTF-8</span>
            <span class="footer-item">LF</span>
            <span class="footer-item language-tag">{{ activeTabLanguage }}</span>
          </template>
        </div>
        <div class="footer-right">
          <span v-if="!activeTab.isBinary" class="footer-item">行 {{ cursorLine }}, 列 {{ cursorColumn }}</span>
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

    <!-- 5. Tab Context Menu (VS Code Aligned 5 Groups) -->
    <Teleport to="body">
      <div
        v-if="isContextMenuVisible && targetTab"
        ref="contextMenuRef"
        class="context-view monaco-menu-container tab-context-menu"
        :class="{ 'submenu-left': isSubmenuNearRightEdge }"
        :style="{ left: contextMenuPos.x + 'px', top: contextMenuPos.y + 'px' }"
        @contextmenu.prevent
      >
        <div class="monaco-menu">
          <div class="monaco-action-bar vertical">
            <ul class="actions-container" role="menu">
              <!-- Group 1: Close Actions -->
              <li class="action-item" role="none" @click="handleCloseCurrent">
                <a class="action-menu-item" role="menuitem">
                  <span class="action-label">关闭</span>
                  <span class="keybinding">{{ isMac ? '⌘W' : 'Ctrl+W' }}</span>
                </a>
              </li>
              <li
                class="action-item"
                :class="{ disabled: !canCloseOthers }"
                role="none"
                @click="handleCloseOthers"
              >
                <a class="action-menu-item" role="menuitem">
                  <span class="action-label">关闭其他</span>
                  <span class="keybinding">{{ isMac ? '⌥⌘T' : 'Alt+Ctrl+T' }}</span>
                </a>
              </li>
              <li
                class="action-item"
                :class="{ disabled: !canCloseToRight }"
                role="none"
                @click="handleCloseToRight"
              >
                <a class="action-menu-item" role="menuitem">
                  <span class="action-label">关闭右侧标签页</span>
                </a>
              </li>
              <li
                class="action-item"
                :class="{ disabled: !canCloseSaved }"
                role="none"
                @click="handleCloseSaved"
              >
                <a class="action-menu-item" role="menuitem">
                  <span class="action-label">关闭已保存</span>
                  <span class="keybinding">{{ isMac ? '[⌘K U]' : '[Ctrl+K U]' }}</span>
                </a>
              </li>
              <li class="action-item" role="none" @click="handleCloseAll">
                <a class="action-menu-item" role="menuitem">
                  <span class="action-label">全部关闭</span>
                  <span class="keybinding">{{ isMac ? '[⌘K W]' : '[Ctrl+K W]' }}</span>
                </a>
              </li>

              <!-- Separator -->
              <li class="action-separator" role="separator"></li>

              <!-- Group 2: Copy Paths -->
              <li class="action-item" role="none" @click="handleCopyAbsolutePath">
                <a class="action-menu-item" role="menuitem">
                  <span class="action-label">复制路径</span>
                  <span class="keybinding">{{ isMac ? '⌥⌘C' : 'Shift+Alt+C' }}</span>
                </a>
              </li>
              <li class="action-item" role="none" @click="handleCopyRelativePath">
                <a class="action-menu-item" role="menuitem">
                  <span class="action-label">复制相对路径</span>
                  <span class="keybinding">{{ isMac ? '⌥⇧⌘C' : 'Ctrl+K Ctrl+C' }}</span>
                </a>
              </li>

              <!-- Separator -->
              <li class="action-separator" role="separator"></li>

              <!-- Group 3: Reveal in Explorer View -->
              <li class="action-item" role="none" @click="handleRevealInExplorer">
                <a class="action-menu-item" role="menuitem">
                  <span class="action-label">在资源管理器视图中显示</span>
                </a>
              </li>

              <!-- Separator -->
              <li class="action-separator" role="separator"></li>

              <!-- Git Graph & Diff Actions -->
              <li class="action-item" role="none" @click="handleViewFileGitGraph">
                <a class="action-menu-item" role="menuitem">
                  <span class="action-label">在 Git Graph 中查看该文件提交历史</span>
                </a>
              </li>
              <li v-if="targetTabHasGitChanges" class="action-item" role="none" @click="handleOpenTabDiff">
                <a class="action-menu-item" role="menuitem">
                  <span class="action-label">打开差异比对 (Diff)</span>
                </a>
              </li>

              <!-- Separator -->
              <li class="action-separator" role="separator"></li>

              <!-- Group 4: Pin / Unpin Tab -->
              <li class="action-item" role="none" @click="handleTogglePin">
                <a class="action-menu-item" role="menuitem">
                  <span class="action-label">{{ targetTab?.isPinned ? '取消固定' : '固定' }}</span>
                  <span class="keybinding">{{
                    targetTab?.isPinned
                      ? (isMac ? '[⌘K ⇧Enter]' : '[Ctrl+K Shift+Enter]')
                      : (isMac ? '[⌘K Enter]' : '[Ctrl+K Enter]')
                  }}</span>
                </a>
              </li>

              <!-- Separator -->
              <li class="action-separator" role="separator"></li>

              <!-- Group 5: Split Editor -->
              <li class="action-item" role="none" @click="handleSplitRight">
                <a class="action-menu-item" role="menuitem">
                  <span class="action-label">向右拆分</span>
                </a>
              </li>
              <li
                ref="submenuTriggerRef"
                class="action-item has-submenu"
                :class="{ focused: isSubmenuVisible }"
                role="none"
                @mouseenter="handleSubmenuTriggerEnter"
                @mouseleave="handleSubmenuTriggerLeave"
              >
                <a class="action-menu-item" role="menuitem">
                  <span class="action-label">拆分和移动</span>
                  <span class="codicon codicon-chevron-right submenu-arrow"></span>
                </a>
              </li>
            </ul>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- 6. Independent Secondary Submenu (Completely separate from parent container) -->
    <Teleport to="body">
      <div
        v-if="isContextMenuVisible && isSubmenuVisible"
        ref="submenuContentRef"
        class="context-view monaco-menu-container tab-submenu-independent"
        :style="{ left: submenuPos.x + 'px', top: submenuPos.y + 'px' }"
        @mouseenter="handleSubmenuContentEnter"
        @mouseleave="handleSubmenuContentLeave"
        @contextmenu.prevent
      >
        <div class="monaco-menu">
          <div class="monaco-action-bar vertical">
            <ul class="actions-container" role="menu">
              <li class="action-item" role="none" @click.stop="handleSplitEditor('up')">
                <a class="action-menu-item" role="menuitem">
                  <span class="action-label">向上拆分</span>
                </a>
              </li>
              <li class="action-item" role="none" @click.stop="handleSplitEditor('down')">
                <a class="action-menu-item" role="menuitem">
                  <span class="action-label">向下拆分</span>
                </a>
              </li>
              <li class="action-item" role="none" @click.stop="handleSplitEditor('left')">
                <a class="action-menu-item" role="menuitem">
                  <span class="action-label">向左拆分</span>
                </a>
              </li>
              <li class="action-item" role="none" @click.stop="handleSplitEditor('right')">
                <a class="action-menu-item" role="menuitem">
                  <span class="action-label">向右拆分</span>
                </a>
              </li>

              <li class="action-separator" role="separator"></li>

              <li class="action-item" role="none" @click.stop="handleSplitEditor('up')">
                <a class="action-menu-item" role="menuitem">
                  <span class="action-label">上移</span>
                </a>
              </li>
              <li class="action-item" role="none" @click.stop="handleSplitEditor('down')">
                <a class="action-menu-item" role="menuitem">
                  <span class="action-label">下移</span>
                </a>
              </li>
              <li class="action-item" role="none" @click.stop="handleSplitEditor('left')">
                <a class="action-menu-item" role="menuitem">
                  <span class="action-label">左移</span>
                </a>
              </li>
              <li class="action-item" role="none" @click.stop="handleSplitEditor('right')">
                <a class="action-menu-item" role="menuitem">
                  <span class="action-label">右移</span>
                </a>
              </li>

              <li class="action-separator" role="separator"></li>

              <li class="action-item" role="none" @click.stop="handleSplitInGroup">
                <a class="action-menu-item" role="menuitem">
                  <span class="action-label">{{ splitState.active && splitState.inGroup ? '加入编辑器组' : '在组中拆分' }}</span>
                </a>
              </li>
            </ul>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, watch } from "vue";
import { useRouter, useRoute } from "vue-router";
import { t } from "../i18n";
import TechIcon from "../components/common/TechIcon.vue";
import MonacoEditor from "../components/editor/MonacoEditor.vue";
import MonacoDiffEditor from "../components/editor/MonacoDiffEditor.vue";
import { useFileTree } from "../composables/useFileTree";
import { useEditorTabs, type EditorTab } from "../composables/useEditorTabs";
import { useWorkspaceBus } from "../composables/useWorkspaceBus";
import { useTerminalSessions } from "../composables/useTerminalSessions";
import { useToast } from "../composables/useToast";
import { useGit } from "../composables/useGit";

const router = useRouter();
const route = useRoute();
const { getGitStatus, setGraphFilterPath, fetchStatus } = useGit();
const { onOpenDiff, setActiveFilePath, onGitStatusChanged } = useWorkspaceBus();

const emit = defineEmits<{
  (e: "open-drawer", tab: string): void;
}>();

const { showToast } = useToast();

const {
  tree,
  selectedNode,
  copyRelativePath,
  startInlineCreate,
  expandPath,
  selectNode,
  fetchTree,
} = useFileTree();

const {
  tabs,
  activeTabId,
  activeTab,
  hasDirtyTabs,
  openDiffTab,
  switchTab,
  updateContent,

  saveTab,
  saveAllTabs,
  closeTab,
  closeOtherTabs,
  closeTabsToRight,
  closeSavedTabs,
  closeAllTabs,
  togglePinTab,
} = useEditorTabs();


// Tab Context Menu State
const contextMenuRef = ref<HTMLDivElement | null>(null);
const isContextMenuVisible = ref(false);
const contextMenuPos = ref({ x: 0, y: 0 });
const targetTab = ref<EditorTab | null>(null);

// Independent Submenu State
const submenuTriggerRef = ref<HTMLElement | null>(null);
const submenuContentRef = ref<HTMLDivElement | null>(null);
const isSubmenuVisible = ref(false);
const submenuPos = ref({ x: 0, y: 0 });
let submenuCloseTimer: ReturnType<typeof setTimeout> | null = null;

function handleSubmenuTriggerEnter(e: MouseEvent) {
  if (submenuCloseTimer) {
    clearTimeout(submenuCloseTimer);
    submenuCloseTimer = null;
  }
  const target = (e.currentTarget as HTMLElement) || submenuTriggerRef.value;
  if (!target) return;
  const rect = target.getBoundingClientRect();
  const submenuWidth = 175;
  const submenuHeight = 260;

  let x = rect.right + 2;
  let y = rect.top - 4;

  if (x + submenuWidth > window.innerWidth - 8) {
    x = rect.left - submenuWidth - 2;
  }

  if (y + submenuHeight > window.innerHeight - 8) {
    y = Math.max(8, window.innerHeight - submenuHeight - 8);
  }

  submenuPos.value = { x, y };
  isSubmenuVisible.value = true;
}

function handleSubmenuTriggerLeave() {
  if (submenuCloseTimer) clearTimeout(submenuCloseTimer);
  submenuCloseTimer = setTimeout(() => {
    isSubmenuVisible.value = false;
  }, 160);
}

function handleSubmenuContentEnter() {
  if (submenuCloseTimer) {
    clearTimeout(submenuCloseTimer);
    submenuCloseTimer = null;
  }
  isSubmenuVisible.value = true;
}

function handleSubmenuContentLeave() {
  if (submenuCloseTimer) clearTimeout(submenuCloseTimer);
  submenuCloseTimer = setTimeout(() => {
    isSubmenuVisible.value = false;
  }, 160);
}

const isMac = computed(() => {
  if (typeof navigator === "undefined") return false;
  return /Mac|iPhone|iPod|iPad/i.test(navigator.platform || navigator.userAgent);
});

const isSubmenuNearRightEdge = computed(() => {
  if (typeof window === "undefined") return false;
  return contextMenuPos.value.x + 250 + 160 > window.innerWidth;
});

const targetTabIndex = computed(() => {
  if (!targetTab.value) return -1;
  return tabs.value.findIndex((t) => t.id === targetTab.value?.id);
});

const canCloseOthers = computed(() => {
  if (!targetTab.value) return false;
  return tabs.value.filter((t) => t.id !== targetTab.value?.id).length > 0;
});

const canCloseToRight = computed(() => {
  if (targetTabIndex.value === -1) return false;
  return targetTabIndex.value < tabs.value.length - 1;
});

const canCloseSaved = computed(() => {
  return tabs.value.some((t) => !t.isDirty && !t.isPinned);
});

function openTabContextMenu(e: MouseEvent, tab: EditorTab) {
  e.preventDefault();
  e.stopPropagation();
  targetTab.value = tab;
  switchTab(tab.id);
  const menuWidth = 260;
  const menuHeight = 360;
  const x = Math.min(e.clientX, window.innerWidth - menuWidth - 10);
  const y = Math.min(e.clientY, window.innerHeight - menuHeight - 10);
  contextMenuPos.value = { x: Math.max(10, x), y: Math.max(10, y) };
  isContextMenuVisible.value = true;
  isSubmenuVisible.value = false;
  if (submenuCloseTimer) {
    clearTimeout(submenuCloseTimer);
    submenuCloseTimer = null;
  }
}

function closeContextMenu() {
  isContextMenuVisible.value = false;
  isSubmenuVisible.value = false;
  if (submenuCloseTimer) {
    clearTimeout(submenuCloseTimer);
    submenuCloseTimer = null;
  }
  targetTab.value = null;
}

function handleCloseCurrent() {
  if (targetTab.value) {
    closeTab(targetTab.value.id);
  }
  closeContextMenu();
}

function handleCloseOthers() {
  if (targetTab.value && canCloseOthers.value) {
    closeOtherTabs(targetTab.value.id);
  }
  closeContextMenu();
}

function handleCloseToRight() {
  if (targetTab.value && canCloseToRight.value) {
    closeTabsToRight(targetTab.value.id);
  }
  closeContextMenu();
}

function handleCloseSaved() {
  if (canCloseSaved.value) {
    closeSavedTabs();
  }
  closeContextMenu();
}

function handleCloseAll() {
  closeAllTabs();
  closeContextMenu();
}

function handleCopyAbsolutePath() {
  if (!targetTab.value) return;
  const rel = targetTab.value.path;
  const root = tree.value?.path || "";
  let fullPath = rel;
  if (!rel.startsWith("/")) {
    fullPath = root ? (root.endsWith("/") ? `${root}${rel}` : `${root}/${rel}`) : rel;
  }
  navigator.clipboard?.writeText(fullPath).then(() => {
    showToast(t("explorer.copied_full_path") || `已复制完整路径: ${fullPath}`, "success");
  }).catch(() => {
    showToast("复制路径失败", "error");
  });
  closeContextMenu();
}

function handleCopyRelativePath() {
  if (!targetTab.value) return;
  const rel = targetTab.value.path;
  navigator.clipboard?.writeText(rel).then(() => {
    showToast(t("explorer.copied_path") || `已复制相对路径: ${rel}`, "success");
  }).catch(() => {
    showToast("复制相对路径失败", "error");
  });
  closeContextMenu();
}

function handleRevealInExplorer() {
  if (!targetTab.value) return;
  const relPath = targetTab.value.path;
  const lastSlash = relPath.lastIndexOf("/");
  if (lastSlash > 0) {
    const parentDir = relPath.slice(0, lastSlash);
    expandPath(parentDir);
  }
  const findNode = (node: any): any => {
    if (!node) return null;
    if (node.path === relPath) return node;
    if (node.children) {
      for (const child of node.children) {
        const found = findNode(child);
        if (found) return found;
      }
    }
    return null;
  };
  if (tree.value) {
    const node = findNode(tree.value);
    if (node) {
      selectNode(node);
      setTimeout(() => {
        const el =
          document.querySelector(`[data-path="${CSS.escape(relPath)}"]`) ||
          document.querySelector(`[title*="${CSS.escape(targetTab.value?.name || "")}"]`);
        if (el) {
          el.scrollIntoView({ behavior: "smooth", block: "nearest" });
        }
      }, 100);
    }
  }
  closeContextMenu();
}

function handleTogglePin() {
  if (targetTab.value) {
    togglePinTab(targetTab.value.id);
  }
  closeContextMenu();
}

const targetTabHasGitChanges = computed(() => {
  if (!targetTab.value) return false;
  return !!getGitStatus(targetTab.value.path);
});

function handleViewFileGitGraph() {
  if (!targetTab.value) return;
  const filePath = targetTab.value.path;
  closeContextMenu();
  setGraphFilterPath(filePath);
  router.push({
    name: "scm",
    query: {
      scope: route.query.scope || "workspace",
      tab: "graph",
      path: filePath,
    },
  });
}

function handleOpenTabDiff() {
  if (!targetTab.value) return;
  const filePath = targetTab.value.path;
  closeContextMenu();
  router.push({
    name: "scm",
    query: {
      scope: route.query.scope || "workspace",
      path: filePath,
      staged: "false",
    },
  });
}

// Split Editor State
const splitState = ref<{
  active: boolean;
  direction: "horizontal" | "vertical";
  position: "up" | "down" | "left" | "right";
  tab: EditorTab | null;
  inGroup?: boolean;
}>({
  active: false,
  direction: "horizontal",
  position: "right",
  tab: null,
  inGroup: false,
});

function handleSplitRight() {
  handleSplitEditor("right");
}

function handleSplitEditor(direction: "up" | "down" | "left" | "right") {
  if (!targetTab.value) return;
  splitState.value = {
    active: true,
    direction: direction === "up" || direction === "down" ? "vertical" : "horizontal",
    position: direction,
    tab: targetTab.value,
    inGroup: false,
  };
  closeContextMenu();
}

function handleSplitInGroup() {
  const currentTab = targetTab.value || activeTab.value;
  if (!currentTab) return;
  if (splitState.value.active && splitState.value.inGroup) {
    closeSplitEditor();
  } else {
    splitState.value = {
      active: true,
      direction: "horizontal",
      position: "right",
      tab: currentTab,
      inGroup: true,
    };
  }
  closeContextMenu();
}

function closeSplitEditor() {
  splitState.value = {
    active: false,
    direction: "horizontal",
    position: "right",
    tab: null,
    inGroup: false,
  };
}

function handleSplitEditorChange(newContent: string) {
  if (splitState.value.tab) {
    updateContent(splitState.value.tab.id, newContent);
  }
}

async function handleSplitEditorSave() {
  if (splitState.value.tab) {
    await saveTab(splitState.value.tab.id);
  }
}

watch(tabs, (currTabs) => {
  if (splitState.value.active && splitState.value.tab) {
    if (!currTabs.some((t) => t.id === splitState.value.tab?.id)) {
      closeSplitEditor();
    }
  }
});

function handleGlobalPointerDown(e: MouseEvent) {
  if (isContextMenuVisible.value) {
    const el = contextMenuRef.value;
    const subEl = submenuContentRef.value;
    const target = e.target as Node;
    if (
      (!el || !el.contains(target)) &&
      (!subEl || !subEl.contains(target))
    ) {
      closeContextMenu();
    }
  }
}

function handleGlobalKeyDown(e: KeyboardEvent) {
  if (e.key === "Escape" && isContextMenuVisible.value) {
    closeContextMenu();
  }
}

const busCleanups: (() => void)[] = [];

onMounted(() => {
  window.addEventListener("pointerdown", handleGlobalPointerDown);
  window.addEventListener("keydown", handleGlobalKeyDown);

  busCleanups.push(
    onOpenDiff((req) => {
      openDiffTab({
        path: req.path,
        name: req.name,
        originalContent: req.originalContent,
        modifiedContent: req.modifiedContent,
        isStaged: req.isStaged,
        commitHash: req.commitHash,
        parentHash: req.parentHash,
      });
    }),
    onGitStatusChanged(() => {
      fetchStatus();
      fetchTree();
    })
  );
});


onBeforeUnmount(() => {
  window.removeEventListener("pointerdown", handleGlobalPointerDown);
  window.removeEventListener("keydown", handleGlobalKeyDown);
  busCleanups.forEach((fn) => fn());
});

watch(
  () => activeTab.value?.path,
  (newPath) => {
    setActiveFilePath(newPath || "");
  },
  { immediate: true }
);

watch(
  () => tabs.value,
  () => {
    tabs.value.forEach((tab) => {
      if (!tab.isDiff && tab.path) {
        const status = getGitStatus(tab.path);
        if (status) {
          tab.gitStatus = status;
        }
      }
    });
  },
  { immediate: true, deep: true }
);


interface BreadcrumbItem {
  id: string;
  name: string;
  path: string;
  isRoot: boolean;
  isDir: boolean;
  isFile: boolean;
}

const breadcrumbs = computed<BreadcrumbItem[]>(() => {
  if (!activeTab.value) return [];
  const rawPath = activeTab.value.path || activeTab.value.name;
  const parts = rawPath.split("/").filter(Boolean);

  const rootName = tree.value?.name || "UniRTM";
  const items: BreadcrumbItem[] = [
    {
      id: "bc-root",
      name: rootName,
      path: "",
      isRoot: true,
      isDir: true,
      isFile: false,
    },
  ];

  let accumulated = "";
  for (let i = 0; i < parts.length; i++) {
    const part = parts[i];
    accumulated = accumulated ? `${accumulated}/${part}` : part;
    const isLast = i === parts.length - 1;
    items.push({
      id: `bc-${accumulated}`,
      name: part,
      path: accumulated,
      isRoot: false,
      isDir: !isLast,
      isFile: isLast,
    });
  }

  return items;
});

function handleBreadcrumbClick(item: BreadcrumbItem) {
  if (item.isDir) {
    if (item.path) {
      expandPath(item.path);
    }
    const findNodeByPath = (node: any, targetPath: string): any => {
      if (node.path === targetPath) return node;
      if (node.children) {
        for (const child of node.children) {
          const found = findNodeByPath(child, targetPath);
          if (found) return found;
        }
      }
      return null;
    };
    if (tree.value) {
      const target = item.path === "" ? tree.value : findNodeByPath(tree.value, item.path);
      if (target) {
        selectNode(target);
      }
    }
  }
}

const cursorLine = ref<number>(1);
const cursorColumn = ref<number>(1);

const activeTabLanguage = computed(() => {
  return (activeTab.value?.language || "plaintext").toUpperCase();
});

function isImageFile(filename: string): boolean {
  return /\.(png|jpe?g|gif|svg|webp|ico|bmp)$/i.test(filename);
}

function formatFileSize(bytes?: number): string {
  if (!bytes || bytes <= 0) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(bytes) / Math.log(1024));
  return `${(bytes / Math.pow(1024, i)).toFixed(i === 0 ? 0 : 1)} ${units[i]}`;
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
  const { ensureInitialSession, activePane, runCommand } = useTerminalSessions();
  ensureInitialSession();
  const pane = activePane.value;
  if (activeTab.value) {
    const fileName = activeTab.value.name;
    const filePath = activeTab.value.path;
    const lastSlash = filePath.lastIndexOf("/");
    const parentDir = lastSlash >= 0 ? filePath.slice(0, lastSlash) : "";

    setTimeout(() => {
      const paneId = pane?.id || "";
      if (parentDir) {
        runCommand(paneId, `cd "${parentDir}" && ./${fileName}`);
      } else {
        runCommand(paneId, `if [ -f "./${fileName}" ]; then ./${fileName}; else cd .. && ./${fileName}; fi`);
      }
    }, 150);
  }
}

function handleOpenFolderInTerminal() {
  emit("open-drawer", "terminal");
  if (selectedNode.value) {
    const { ensureInitialSession, activePane, runCommand } = useTerminalSessions();
    ensureInitialSession();
    const pane = activePane.value;
    const targetPath = selectedNode.value.path || ".";
    setTimeout(() => {
      runCommand(pane?.id || "", `cd "${targetPath}"`);
    }, 150);
  }
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

.diff-icon {
  color: #38bdf8;
}

.tab-title-diff {
  font-style: italic;
}

.tab-git-badge {
  font-size: 10px;
  font-weight: 700;
  padding: 1px 4px;
  border-radius: 3px;
  line-height: 1;
}

.git-color-M {
  color: #eab308 !important;
}

.badge-M {
  background-color: rgba(234, 179, 8, 0.2);
  color: #eab308;
}

.git-color-U,
.git-color-A {
  color: #22c55e !important;
}

.badge-U,
.badge-A {
  background-color: rgba(34, 197, 94, 0.2);
  color: #22c55e;
}

.git-color-D {
  color: #ef4444 !important;
}

.badge-D {
  background-color: rgba(239, 68, 68, 0.2);
  color: #ef4444;
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
  height: 26px;
  min-height: 26px;
  padding: 0 12px;
  background-color: var(--section-bg, var(--bg-color));
  color: var(--text-muted);
  font-size: 11.5px;
  border-bottom: 1px solid var(--card-border);
  flex-shrink: 0;
  overflow: hidden;
  user-select: none;
}

.bc-items-scroll {
  display: flex;
  align-items: center;
  gap: 2px;
  overflow-x: auto;
  scrollbar-width: none;
  white-space: nowrap;
  max-width: 100%;
}

.bc-items-scroll::-webkit-scrollbar {
  display: none;
}

.bc-item {
  display: inline-flex;
  align-items: center;
  gap: 4.5px;
  padding: 2px 5px;
  border-radius: 4px;
  color: var(--text-muted);
  font-size: 11.5px;
  font-family: inherit;
  background: transparent;
  border: none;
  cursor: pointer;
  transition: background-color 0.15s ease, color 0.15s ease;
  line-height: 1;
}

.bc-item:hover {
  background-color: rgba(125, 125, 125, 0.12);
  color: var(--text-color, #0f172a);
}

.bc-item.is-active {
  color: var(--text-color, #0f172a);
  font-weight: 500;
}

.bc-icon-wrapper {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  line-height: 1;
}

.bc-icon {
  font-size: 13px;
  opacity: 0.9;
}

.bc-label {
  line-height: 1.2;
  letter-spacing: -0.1px;
}

.bc-separator {
  font-size: 10px;
  color: var(--text-muted);
  opacity: 0.55;
  margin: 0 1px;
  display: inline-flex;
  align-items: center;
  user-select: none;
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

/* Binary / Large File Placeholder (VS Code Aligned) */
.binary-file-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  padding: 32px 24px;
  background-color: var(--bg-color);
  user-select: none;
}

.binary-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  max-width: 460px;
  padding: 36px 32px;
  background-color: var(--card-bg, rgba(255, 255, 255, 0.03));
  border: 1px solid var(--card-border, rgba(255, 255, 255, 0.08));
  border-radius: 12px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.08);
}

.binary-icon-box {
  width: 64px;
  height: 64px;
  border-radius: 16px;
  background: rgba(148, 163, 184, 0.1);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 16px;
}

.binary-hero-icon {
  font-size: 32px;
  color: var(--text-muted);
}

.binary-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-color, #0f172a);
  margin: 0 0 10px 0;
  word-break: break-all;
}

.binary-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
}

.binary-badge {
  font-size: 11px;
  font-weight: 500;
  padding: 2px 8px;
  border-radius: 4px;
  background: rgba(148, 163, 184, 0.15);
  color: var(--text-muted);
}

.binary-badge.tag {
  background: rgba(2, 132, 199, 0.12);
  color: #0284c7;
}

[data-theme="dark"] .binary-badge.tag {
  background: rgba(0, 229, 255, 0.15);
  color: #00e5ff;
}

.binary-message {
  font-size: 13px;
  line-height: 1.6;
  color: var(--text-muted);
  margin: 0 0 24px 0;
}

.binary-actions {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
  justify-content: center;
}

.action-btn-primary {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 7px 14px;
  border-radius: 6px;
  background-color: var(--accent-cyan, #0284c7);
  color: #ffffff;
  border: none;
  font-size: 12.5px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
}

.action-btn-primary:hover {
  filter: brightness(1.1);
  transform: translateY(-1px);
}

.action-btn-secondary {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 7px 14px;
  border-radius: 6px;
  background: transparent;
  color: var(--text-color, #0f172a);
  border: 1px solid var(--card-border, rgba(0, 0, 0, 0.15));
  font-size: 12.5px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.action-btn-secondary:hover {
  background-color: rgba(125, 125, 125, 0.1);
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

/* Pinned Tabs */
.editor-tab.is-pinned {
  border-top: 2px solid var(--accent-cyan);
  font-weight: 500;
}

.editor-tab .tab-pin-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 2px 4px;
  border-radius: 3px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  transition: all 0.15s ease;
}

.editor-tab .tab-pin-btn:hover {
  background-color: rgba(128, 128, 128, 0.2);
  color: var(--text-main);
}

/* Dual Pane Split Editor Container */
.editor-split-container {
  display: flex;
  width: 100%;
  height: 100%;
  overflow: hidden;
}

.editor-split-container.split-horizontal {
  flex-direction: row;
}

.editor-split-container.split-horizontal.split-pos-left {
  flex-direction: row-reverse;
}

.editor-split-container.split-vertical {
  flex-direction: column;
}

.editor-split-container.split-vertical.split-pos-up {
  flex-direction: column-reverse;
}

.split-pane {
  flex: 1 1 50%;
  min-width: 0;
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background-color: var(--bg-color);
}

.split-pane-header {
  height: 28px;
  min-height: 28px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 10px;
  background-color: var(--card-bg);
  border-bottom: 1px solid var(--card-border);
  user-select: none;
}

.split-pane-info {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--text-main);
}

.split-pane-name {
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.split-pane-tools {
  display: flex;
  align-items: center;
}

.split-tool-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 2px 4px;
  border-radius: 3px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  transition: all 0.15s ease;
}

.split-tool-btn:hover {
  background-color: rgba(128, 128, 128, 0.2);
  color: var(--text-main);
}

.split-pane-body {
  flex: 1;
  min-width: 0;
  min-height: 0;
  position: relative;
  overflow: hidden;
}

.split-divider {
  background-color: var(--card-border);
  flex-shrink: 0;
  z-index: 5;
}

.split-divider.divider-horizontal {
  width: 2px;
  cursor: col-resize;
}

.split-divider.divider-vertical {
  height: 2px;
  cursor: row-resize;
}
</style>
