// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { ref, computed } from "vue";
import { useEditorTabs } from "./useEditorTabs";

export type DrawerTab = "terminal" | "output" | "tasks" | "doctor" | "ports";

export interface ShortcutItem {
  id: string;
  category: "general" | "view" | "panel" | "editor" | "navigation";
  keys: string;
  macKeys?: string;
  titleZh: string;
  titleEn: string;
  descZh: string;
  descEn: string;
  when?: string;
}

export const SHORTCUT_DEFINITIONS: ShortcutItem[] = [
  // 1. General & Palette
  {
    id: "workbench.action.showCommands",
    category: "general",
    keys: "Ctrl+Shift+P / F1",
    macKeys: "Cmd+Shift+P / F1",
    titleZh: "显示所有命令 (命令面板)",
    titleEn: "Show All Commands (Command Palette)",
    descZh: "打开全局命令面板搜索任意功能、页面或执行快捷指令",
    descEn: "Open Command Palette to search features, pages, or execute commands",
  },
  {
    id: "workbench.action.quickOpen",
    category: "general",
    keys: "Ctrl+P",
    macKeys: "Cmd+P",
    titleZh: "快速打开 / 页面跳转",
    titleEn: "Quick Open / Navigate",
    descZh: "快速搜索并跳转至页面、功能或执行任务",
    descEn: "Quickly navigate to files, views, or execute actions",
  },
  {
    id: "workbench.action.openSettings",
    category: "general",
    keys: "Ctrl+,",
    macKeys: "Cmd+,",
    titleZh: "打开系统设置",
    titleEn: "Open Settings",
    descZh: "打开系统配置偏好面板 (外观、网络、代理与环境)",
    descEn: "Open preferences and system settings modal",
  },
  {
    id: "workbench.action.openGlobalKeybindings",
    category: "general",
    keys: "Ctrl+K Ctrl+S",
    macKeys: "Cmd+K Cmd+S",
    titleZh: "打开键盘快捷键参考",
    titleEn: "Open Keyboard Shortcuts",
    descZh: "查看 UniRTM 与 VS Code 深度对齐的所有键盘快捷键",
    descEn: "View keyboard shortcuts aligned with VS Code",
  },
  {
    id: "workbench.action.closeModalOrDrawer",
    category: "general",
    keys: "Escape",
    macKeys: "Escape",
    titleZh: "关闭当前弹窗 / 浮层 / 底部控制台",
    titleEn: "Close Current Modal / Flyout / Bottom Drawer",
    descZh: "退出搜索栏、设置窗口、快捷键面板或折叠底部控制台",
    descEn: "Dismiss search palette, active modals, or bottom drawer",
  },

  // 2. Sidebar & Bottom Panel
  {
    id: "workbench.action.toggleSidebarVisibility",
    category: "panel",
    keys: "Ctrl+B",
    macKeys: "Cmd+B",
    titleZh: "显示 / 隐藏主侧边栏",
    titleEn: "Toggle Primary Side Bar",
    descZh: "折叠或展开左侧主导航与子模块侧边栏",
    descEn: "Toggle primary collapsible side bar navigation",
  },
  {
    id: "workbench.action.togglePanel",
    category: "panel",
    keys: "Ctrl+J",
    macKeys: "Cmd+J",
    titleZh: "显示 / 隐藏底部控制台面板",
    titleEn: "Toggle Bottom Panel",
    descZh: "展开或收起底部抽屉 (集成终端 / 任务输出 / 运行日志)",
    descEn: "Toggle bottom drawer for terminal, tasks, and output",
  },
  {
    id: "workbench.action.terminal.toggleTerminal",
    category: "panel",
    keys: "Ctrl+`",
    macKeys: "Cmd+`",
    titleZh: "切换集成终端",
    titleEn: "Toggle Integrated Terminal",
    descZh: "快速呼出或隐藏底部集成终端会话",
    descEn: "Focus or toggle active integrated terminal",
  },
  {
    id: "workbench.action.terminal.new",
    category: "panel",
    keys: "Ctrl+Shift+`",
    macKeys: "Cmd+Shift+`",
    titleZh: "新建终端会话",
    titleEn: "Create New Terminal Session",
    descZh: "在底部终端抽屉中创建并激活一个新的 Shell 会话",
    descEn: "Spawn a new shell tab in integrated terminal",
  },
  {
    id: "workbench.action.output.show",
    category: "panel",
    keys: "Ctrl+Shift+U",
    macKeys: "Cmd+Shift+U",
    titleZh: "显示输出通道与系统日志",
    titleEn: "Show Output Channel & Logs",
    descZh: "切换至底部抽屉的日志输出控制台",
    descEn: "Switch bottom drawer to runtime output and system logs",
  },
  {
    id: "workbench.action.problems.show",
    category: "panel",
    keys: "Ctrl+Shift+M",
    macKeys: "Cmd+Shift+M",
    titleZh: "打开系统诊断与环境体检 (Doctor)",
    titleEn: "Show System Health & Doctor",
    descZh: "直接跳转至系统全面体检与依赖诊断视图",
    descEn: "Jump to Doctor health diagnostics view",
  },

  // 3. Activity Bar Views Navigation
  {
    id: "workbench.view.explorer",
    category: "view",
    keys: "Ctrl+Shift+E",
    macKeys: "Cmd+Shift+E",
    titleZh: "显示资源管理器 (Explorer)",
    titleEn: "Show Explorer",
    descZh: "切换至工作区文件目录树与代码编辑区",
    descEn: "Switch to workspace file explorer and editor",
  },
  {
    id: "workbench.view.scm",
    category: "view",
    keys: "Ctrl+Shift+G",
    macKeys: "Cmd+Shift+G",
    titleZh: "显示源代码管理 (SCM / Git)",
    titleEn: "Show Source Control (SCM)",
    descZh: "切换至 Git 变更、暂存、提交历史与图谱视图",
    descEn: "Switch to Git status, staging and commit graph view",
  },
  {
    id: "workbench.view.doctor",
    category: "view",
    keys: "Ctrl+Shift+D",
    macKeys: "Cmd+Shift+D",
    titleZh: "显示环境诊断与运行状态 (Doctor)",
    titleEn: "Show Diagnosis (Doctor)",
    descZh: "探测 Shims 路由、磁盘足迹、云服务联通性",
    descEn: "Inspect shims, runtime health, and connectivity",
  },
  {
    id: "workbench.view.extensions",
    category: "view",
    keys: "Ctrl+Shift+X",
    macKeys: "Cmd+Shift+X",
    titleZh: "显示扩展与主题市场 (Marketplace)",
    titleEn: "Show Extensions & Marketplace",
    descZh: "浏览并安装外观主题、IDE 扩展及后端插件",
    descEn: "Explore themes, extensions, and plugins",
  },
  {
    id: "workbench.action.toggleScope",
    category: "view",
    keys: "Alt+S",
    macKeys: "Option+S",
    titleZh: "切换配置作用域 (工作区 / 用户全局)",
    titleEn: "Toggle Scope (Workspace / User Global)",
    descZh: "在本地 .unirtm.toml 与全局 ~/.unirtm 之间切换",
    descEn: "Toggle active configuration scope",
  },
  {
    id: "workbench.view.jumpByNumber",
    category: "navigation",
    keys: "Alt+1 ~ Alt+9",
    macKeys: "Option+1 ~ Option+9",
    titleZh: "直达活动栏面板 (1~9)",
    titleEn: "Direct Jump to Activity Bar (1-9)",
    descZh: "1: 仪表盘, 2: 资源管理器, 3: Git, 4: 质量守护, 5: 工具链, 6: 任务, 7: 环境变量, 8: 扩展市场, 9: 诊断",
    descEn: "1: Dashboard, 2: Explorer, 3: SCM, 4: Guard, 5: Tools, 6: Tasks, 7: Env, 8: Marketplace, 9: Doctor",
  },

  // 4. Editor & Tabs Management
  {
    id: "workbench.action.files.save",
    category: "editor",
    keys: "Ctrl+S",
    macKeys: "Cmd+S",
    titleZh: "保存当前文件",
    titleEn: "Save Active File",
    descZh: "将编辑器中当前文件的修改写入磁盘",
    descEn: "Save active modified file to disk",
  },
  {
    id: "workbench.action.files.saveAll",
    category: "editor",
    keys: "Ctrl+Shift+S",
    macKeys: "Cmd+Shift+S",
    titleZh: "保存所有已修改文件",
    titleEn: "Save All Files",
    descZh: "一键保存所有标签页中处于脏状态的代码文件",
    descEn: "Save all dirty tabs across the workbench",
  },
  {
    id: "workbench.action.closeActiveEditor",
    category: "editor",
    keys: "Ctrl+W / Ctrl+F4",
    macKeys: "Cmd+W",
    titleZh: "关闭当前编辑器标签页",
    titleEn: "Close Active Editor Tab",
    descZh: "关闭当前聚焦的文件编辑或差异对比标签页",
    descEn: "Close active editor tab in workspace",
  },
  {
    id: "workbench.action.closeAllEditors",
    category: "editor",
    keys: "Ctrl+K Ctrl+W",
    macKeys: "Cmd+K Cmd+W",
    titleZh: "关闭所有编辑器标签页",
    titleEn: "Close All Editor Tabs",
    descZh: "一键关闭工作区中打开的所有文件和 Diff 标签页",
    descEn: "Close all opened editor tabs",
  },
  {
    id: "workbench.action.nextEditor",
    category: "editor",
    keys: "Ctrl+PageDown",
    macKeys: "Cmd+Alt+Right",
    titleZh: "切换到下一个标签页",
    titleEn: "Next Editor Tab",
    descZh: "在已打开的编辑器标签页之间向前切换",
    descEn: "Cycle forward through open editor tabs",
  },
  {
    id: "workbench.action.previousEditor",
    category: "editor",
    keys: "Ctrl+PageUp",
    macKeys: "Cmd+Alt+Left",
    titleZh: "切换到上一个标签页",
    titleEn: "Previous Editor Tab",
    descZh: "在已打开的编辑器标签页之间向后切换",
    descEn: "Cycle backward through open editor tabs",
  },
  {
    id: "workbench.action.selectTheme",
    category: "general",
    keys: "Ctrl+K Ctrl+T",
    macKeys: "Cmd+K Cmd+T",
    titleZh: "偏好设置: 外观颜色主题",
    titleEn: "Preferences: Color Theme",
    descZh: "打开扩展市场的颜色主题分类进行浏览与配置",
    descEn: "Browse and select color themes in marketplace",
  },
];

export interface ShortcutHandlers {
  openCommandPalette: () => void;
  closeCommandPalette: () => void;
  isCommandPaletteOpen: () => boolean;
  openSettings: () => void;
  toggleSidebar: () => void;
  toggleDrawer: () => void;
  openDrawerTab: (tab: DrawerTab) => void;
  newTerminalSession: () => void;
  navigateToTab: (tabId: string) => void;
  toggleTheme: () => void;
  toggleScope: () => void;
  isDrawerOpen: () => boolean;
  closeDrawer: () => void;
}

const isShortcutsModalOpen = ref(false);

export function useKeyboardShortcuts() {
  function openShortcutsModal() {
    isShortcutsModalOpen.value = true;
  }

  function closeShortcutsModal() {
    isShortcutsModalOpen.value = false;
  }

  const { tabs, activeTabId, switchTab, saveTab, saveAllTabs, closeTab, closeAllTabs } = useEditorTabs();

  function nextEditorTab() {
    if (tabs.value.length <= 1) return;
    const idx = tabs.value.findIndex((t) => t.id === activeTabId.value);
    const nextIdx = (idx + 1) % tabs.value.length;
    switchTab(tabs.value[nextIdx].id);
  }

  function prevEditorTab() {
    if (tabs.value.length <= 1) return;
    const idx = tabs.value.findIndex((t) => t.id === activeTabId.value);
    const prevIdx = (idx - 1 + tabs.value.length) % tabs.value.length;
    switchTab(tabs.value[prevIdx].id);
  }

  function closeActiveEditorTab() {
    if (activeTabId.value) {
      closeTab(activeTabId.value);
    }
  }

  function saveActiveEditorTab() {
    if (activeTabId.value) {
      saveTab(activeTabId.value);
    }
  }

  // Chord state tracker (e.g. Ctrl+K chord)
  let pendingChord: "Ctrl+K" | null = null;
  let chordTimer: any = null;

  function clearChord() {
    pendingChord = null;
    if (chordTimer) {
      clearTimeout(chordTimer);
      chordTimer = null;
    }
  }

  const tabOrder = ["dashboard", "explorer", "scm", "guard", "tools", "tasks", "environments", "marketplace", "doctor"];

  function createKeydownHandler(handlers: ShortcutHandlers) {
    return function handleKeydown(e: KeyboardEvent) {
      const isMac =
        typeof navigator !== "undefined" && navigator.platform
          ? navigator.platform.toUpperCase().indexOf("MAC") >= 0
          : false;
      const mod = isMac ? e.metaKey : e.ctrlKey;
      const key = e.key;

      // Handle active chord state
      if (pendingChord === "Ctrl+K") {
        clearChord();

        // Ctrl+K Ctrl+S or 's'/'S': Open Keyboard Shortcuts
        if ((mod && (key === "s" || key === "S")) || (!mod && (key === "s" || key === "S"))) {
          e.preventDefault();
          openShortcutsModal();
          return;
        }

        // Ctrl+K Ctrl+W or 'w'/'W': Close All Editor Tabs
        if ((mod && (key === "w" || key === "W")) || (!mod && (key === "w" || key === "W"))) {
          e.preventDefault();
          closeAllTabs();
          return;
        }

        // Ctrl+K Ctrl+T or 't'/'T': Select Color Theme
        if ((mod && (key === "t" || key === "T")) || (!mod && (key === "t" || key === "T"))) {
          e.preventDefault();
          handlers.navigateToTab("marketplace");
          return;
        }

        // If another chord wasn't matched, fall through
      }

      // Check if user is typing in standard inputs
      const target = e.target as HTMLElement | null;
      const isInput =
        target &&
        (target.tagName === "INPUT" ||
          target.tagName === "TEXTAREA" ||
          target.tagName === "SELECT" ||
          target.isContentEditable);

      // Escape key: dismiss modals/palette/drawers
      if (key === "Escape") {
        if (isShortcutsModalOpen.value) {
          e.preventDefault();
          closeShortcutsModal();
          return;
        }
        if (handlers.isCommandPaletteOpen()) {
          e.preventDefault();
          handlers.closeCommandPalette();
          return;
        }
        if (handlers.isDrawerOpen()) {
          e.preventDefault();
          handlers.closeDrawer();
          return;
        }
        return;
      }

      // If shortcuts modal is open, don't execute background hotkeys
      if (isShortcutsModalOpen.value) {
        return;
      }

      // --- VS Code Command Palette & Quick Open ---
      // Ctrl+Shift+P / Cmd+Shift+P / F1: Command Palette
      if ((mod && e.shiftKey && (key === "p" || key === "P")) || key === "F1") {
        e.preventDefault();
        handlers.openCommandPalette();
        return;
      }

      // Ctrl+P / Cmd+P: Quick Open / Command Search
      if (mod && !e.shiftKey && !e.altKey && (key === "p" || key === "P")) {
        e.preventDefault();
        handlers.openCommandPalette();
        return;
      }

      // Ctrl+K / Cmd+K: Chord initiator or Command Palette
      if (mod && !e.shiftKey && !e.altKey && (key === "k" || key === "K")) {
        e.preventDefault();
        if (handlers.isCommandPaletteOpen()) {
          handlers.closeCommandPalette();
          return;
        }
        // Start chord detection with a 380ms timeout
        pendingChord = "Ctrl+K";
        chordTimer = setTimeout(() => {
          if (pendingChord === "Ctrl+K") {
            clearChord();
            handlers.openCommandPalette();
          }
        }, 380);
        return;
      }

      // Ctrl+, / Cmd+,: Open Settings
      if (mod && !e.shiftKey && !e.altKey && key === ",") {
        e.preventDefault();
        handlers.openSettings();
        return;
      }

      // --- Side Bar & Panel Layout ---
      // Ctrl+B / Cmd+B: Toggle Primary Side Bar
      if (mod && !e.shiftKey && !e.altKey && (key === "b" || key === "B")) {
        e.preventDefault();
        handlers.toggleSidebar();
        return;
      }

      // Ctrl+J / Cmd+J: Toggle Bottom Panel (Drawer)
      if (mod && !e.shiftKey && !e.altKey && (key === "j" || key === "J")) {
        e.preventDefault();
        handlers.toggleDrawer();
        return;
      }

      // Ctrl+` / Cmd+`: Toggle Integrated Terminal
      if (mod && !e.shiftKey && !e.altKey && (key === "`" || key === "~")) {
        e.preventDefault();
        handlers.openDrawerTab("terminal");
        return;
      }

      // Ctrl+Shift+` / Cmd+Shift+`: Create New Terminal Session
      if (mod && e.shiftKey && (key === "`" || key === "~")) {
        e.preventDefault();
        handlers.newTerminalSession();
        return;
      }

      // Ctrl+Shift+U / Cmd+Shift+U: Show Output Channel
      if (mod && e.shiftKey && (key === "u" || key === "U")) {
        e.preventDefault();
        handlers.openDrawerTab("output");
        return;
      }

      // Ctrl+Shift+M / Cmd+Shift+M: Show Problems & Doctor
      if (mod && e.shiftKey && (key === "m" || key === "M")) {
        e.preventDefault();
        handlers.navigateToTab("doctor");
        return;
      }

      // --- Activity Bar Views Navigation ---
      // Ctrl+Shift+E: Show Explorer
      if (mod && e.shiftKey && (key === "e" || key === "E")) {
        e.preventDefault();
        handlers.navigateToTab("explorer");
        return;
      }

      // Ctrl+Shift+G: Show SCM / Git
      if (mod && e.shiftKey && (key === "g" || key === "G")) {
        e.preventDefault();
        handlers.navigateToTab("scm");
        return;
      }

      // Ctrl+Shift+D: Show Doctor (Run & Debug position)
      if (mod && e.shiftKey && (key === "d" || key === "D")) {
        e.preventDefault();
        handlers.navigateToTab("doctor");
        return;
      }

      // Ctrl+Shift+X: Show Extensions / Marketplace
      if (mod && e.shiftKey && (key === "x" || key === "X")) {
        e.preventDefault();
        handlers.navigateToTab("marketplace");
        return;
      }

      // Alt+S: Toggle Scope (Workspace / User)
      if (e.altKey && !mod && !e.shiftKey && (key === "s" || key === "S")) {
        e.preventDefault();
        handlers.toggleScope();
        return;
      }

      // Alt+1 ~ Alt+9: Direct View Switch
      if (e.altKey && !mod && !e.shiftKey && /^[1-9]$/.test(key)) {
        e.preventDefault();
        const tabIndex = parseInt(key, 10) - 1;
        if (tabOrder[tabIndex]) {
          handlers.navigateToTab(tabOrder[tabIndex]);
        }
        return;
      }

      // --- Editor & Tabs Management ---
      // Ctrl+Shift+S / Cmd+Shift+S: Save All Files
      if (mod && e.shiftKey && (key === "s" || key === "S")) {
        e.preventDefault();
        saveAllTabs();
        return;
      }

      // Ctrl+S / Cmd+S: Save Active File (unless focused on input where user needs browser default)
      if (mod && !e.shiftKey && !e.altKey && (key === "s" || key === "S")) {
        e.preventDefault();
        saveActiveEditorTab();
        return;
      }

      // Ctrl+W / Cmd+W or Ctrl+F4: Close Active Editor Tab
      if ((mod && !e.shiftKey && !e.altKey && (key === "w" || key === "W") && !isInput) || (mod && key === "F4")) {
        if (activeTabId.value) {
          e.preventDefault();
          closeActiveEditorTab();
          return;
        }
      }

      // Ctrl+PageDown / Ctrl+PageUp: Switch Next / Previous Editor Tab
      if (mod && !e.shiftKey && !e.altKey && key === "PageDown") {
        e.preventDefault();
        nextEditorTab();
        return;
      }
      if (mod && !e.shiftKey && !e.altKey && key === "PageUp") {
        e.preventDefault();
        prevEditorTab();
        return;
      }
    };
  }

  return {
    isShortcutsModalOpen,
    openShortcutsModal,
    closeShortcutsModal,
    allShortcuts: computed(() => SHORTCUT_DEFINITIONS),
    createKeydownHandler,
    nextEditorTab,
    prevEditorTab,
    closeActiveEditorTab,
    saveActiveEditorTab,
  };
}
