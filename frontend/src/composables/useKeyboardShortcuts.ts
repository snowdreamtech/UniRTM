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

export interface CustomKeybinding {
  command: string;
  key: string;
  when?: string;
}

export interface ActiveShortcutItem extends ShortcutItem {
  effectiveKey: string;
  isCustom: boolean;
  source: "user" | "default";
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
    descZh: "查看与自定义 UniRTM 键盘快捷键 (对齐 VS Code keybindings.json)",
    descEn: "View and customize keyboard shortcuts (aligned with VS Code keybindings.json)",
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

// Global state for shortcuts and keybindings
const isShortcutsModalOpen = ref(false);
const customKeybindings = ref<CustomKeybinding[]>([]);
const keybindingsPath = ref<string>("~/.config/unirtm/keybindings.json");
const isKeybindingsLoading = ref(false);
const rawKeybindingsJson = ref<string>("[]");

export function parseKeyCombo(combo: string): {
  ctrl: boolean;
  shift: boolean;
  alt: boolean;
  meta: boolean;
  key: string;
} {
  const tokens = combo
    .split("+")
    .map((t) => t.trim().toLowerCase())
    .filter(Boolean);
  let ctrl = false;
  let shift = false;
  let alt = false;
  let meta = false;
  let mainKey = "";

  for (const token of tokens) {
    if (token === "ctrl" || token === "control") {
      ctrl = true;
    } else if (token === "shift") {
      shift = true;
    } else if (token === "alt" || token === "option") {
      alt = true;
    } else if (token === "cmd" || token === "meta" || token === "command") {
      meta = true;
    } else {
      mainKey = token;
    }
  }

  return { ctrl, shift, alt, meta, key: mainKey };
}

export function matchesSingleCombo(e: KeyboardEvent, comboStr: string): boolean {
  const parsed = parseKeyCombo(comboStr);
  const isMac =
    typeof navigator !== "undefined" && navigator.platform
      ? navigator.platform.toUpperCase().indexOf("MAC") >= 0
      : false;

  // Modifiers check
  const requiresMod = parsed.ctrl || parsed.meta;
  const hasMod = isMac ? e.metaKey || e.ctrlKey : e.ctrlKey;
  if (requiresMod && !hasMod) return false;
  if (!requiresMod && hasMod) return false;

  if (parsed.shift !== e.shiftKey) return false;
  if (parsed.alt !== e.altKey) return false;

  // Main key check
  const eventKey = e.key.toLowerCase();
  const targetKey = parsed.key.toLowerCase();

  if (targetKey === "pagedown") return eventKey === "pagedown";
  if (targetKey === "pageup") return eventKey === "pageup";
  if (targetKey === "`" || targetKey === "~") return eventKey === "`" || eventKey === "~";
  if (targetKey === ",") return eventKey === ",";

  return eventKey === targetKey;
}

export function matchesKeyCombo(e: KeyboardEvent, keysString: string): boolean {
  // If string contains multiple options split by " / "
  const options = keysString.split("/").map((s) => s.trim());
  return options.some((opt) => matchesSingleCombo(e, opt));
}

export function normalizeKeyComboFromEvent(e: KeyboardEvent): string {
  const parts: string[] = [];
  if (e.ctrlKey) parts.push("Ctrl");
  if (e.altKey) parts.push("Alt");
  if (e.shiftKey) parts.push("Shift");
  if (e.metaKey) parts.push("Cmd");

  let keyName = e.key;
  if (["Control", "Shift", "Alt", "Meta"].includes(keyName)) {
    return parts.join("+");
  }

  if (keyName === " ") keyName = "Space";
  if (keyName.length === 1) keyName = keyName.toUpperCase();
  parts.push(keyName);

  return parts.join("+");
}

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

  const activeShortcuts = computed<ActiveShortcutItem[]>(() => {
    return SHORTCUT_DEFINITIONS.map((def) => {
      const custom = customKeybindings.value.find((c) => c.command === def.id);
      if (custom && custom.key) {
        return {
          ...def,
          effectiveKey: custom.key,
          isCustom: true,
          source: "user",
          when: custom.when || def.when,
        };
      }
      return {
        ...def,
        effectiveKey: def.keys,
        isCustom: false,
        source: "default",
      };
    });
  });

  async function fetchKeybindings(): Promise<void> {
    isKeybindingsLoading.value = true;
    try {
      const res = await fetch("/api/keybindings");
      if (res.ok) {
        const data = await res.json();
        if (data.path) keybindingsPath.value = data.path;
        if (data.raw) rawKeybindingsJson.value = data.raw;
        if (Array.isArray(data.keybindings)) {
          customKeybindings.value = data.keybindings;
        }
      }
    } catch (err) {
      console.warn("Failed to fetch keybindings from /api/keybindings:", err);
    } finally {
      isKeybindingsLoading.value = false;
    }
  }

  async function saveKeybinding(commandId: string, newKey: string, when?: string): Promise<boolean> {
    const list = [...customKeybindings.value];
    const existingIdx = list.findIndex((c) => c.command === commandId);
    if (existingIdx !== -1) {
      list[existingIdx] = { command: commandId, key: newKey, when };
    } else {
      list.push({ command: commandId, key: newKey, when });
    }
    customKeybindings.value = list;

    try {
      const res = await fetch("/api/keybindings", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ keybindings: list }),
      });
      if (res.ok) {
        const data = await res.json();
        if (data.raw) rawKeybindingsJson.value = data.raw;
        return true;
      }
    } catch (err) {
      console.error("Failed to persist keybinding to /api/keybindings:", err);
    }
    return false;
  }

  async function resetKeybinding(commandId: string): Promise<boolean> {
    const list = customKeybindings.value.filter((c) => c.command !== commandId);
    customKeybindings.value = list;

    try {
      const res = await fetch("/api/keybindings", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ keybindings: list }),
      });
      if (res.ok) {
        const data = await res.json();
        if (data.raw) rawKeybindingsJson.value = data.raw;
        return true;
      }
    } catch (err) {
      console.error("Failed to reset keybinding:", err);
    }
    return false;
  }

  async function resetAllKeybindings(): Promise<boolean> {
    customKeybindings.value = [];
    try {
      const res = await fetch("/api/keybindings", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ keybindings: [] }),
      });
      if (res.ok) {
        rawKeybindingsJson.value = "[]\n";
        return true;
      }
    } catch (err) {
      console.error("Failed to reset all keybindings:", err);
    }
    return false;
  }

  async function saveRawKeybindings(raw: string): Promise<boolean> {
    try {
      const parsed = JSON.parse(raw);
      if (!Array.isArray(parsed)) {
        throw new Error("Keybindings JSON must be an array of objects");
      }
      const res = await fetch("/api/keybindings", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ raw }),
      });
      if (res.ok) {
        const data = await res.json();
        rawKeybindingsJson.value = data.raw || raw;
        if (Array.isArray(data.keybindings)) {
          customKeybindings.value = data.keybindings;
        }
        return true;
      }
    } catch (err) {
      console.error("Failed to save raw keybindings:", err);
    }
    return false;
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

  function getEffectiveKey(commandId: string, defaultKeys: string): string {
    const found = activeShortcuts.value.find((s) => s.id === commandId);
    return found ? found.effectiveKey : defaultKeys;
  }

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

        // Ctrl+K Ctrl+S: Open Keyboard Shortcuts
        if ((mod && (key === "s" || key === "S")) || (!mod && (key === "s" || key === "S"))) {
          e.preventDefault();
          openShortcutsModal();
          return;
        }

        // Ctrl+K Ctrl+W: Close All Editor Tabs
        if ((mod && (key === "w" || key === "W")) || (!mod && (key === "w" || key === "W"))) {
          e.preventDefault();
          closeAllTabs();
          return;
        }

        // Ctrl+K Ctrl+T: Select Color Theme
        if ((mod && (key === "t" || key === "T")) || (!mod && (key === "t" || key === "T"))) {
          e.preventDefault();
          handlers.navigateToTab("marketplace");
          return;
        }
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

      // --- 1. Dynamic Matching of Custom / Effective Shortcuts ---
      // Command Palette
      const showCmdsKey = getEffectiveKey("workbench.action.showCommands", "Ctrl+Shift+P / F1");
      if (matchesKeyCombo(e, showCmdsKey)) {
        e.preventDefault();
        handlers.openCommandPalette();
        return;
      }

      // Quick Open
      const quickOpenKey = getEffectiveKey("workbench.action.quickOpen", "Ctrl+P");
      if (matchesKeyCombo(e, quickOpenKey)) {
        e.preventDefault();
        handlers.openCommandPalette();
        return;
      }

      // Open Settings
      const settingsKey = getEffectiveKey("workbench.action.openSettings", "Ctrl+,");
      if (matchesKeyCombo(e, settingsKey)) {
        e.preventDefault();
        handlers.openSettings();
        return;
      }

      // Open Keyboard Shortcuts
      const kbShortcutsKey = getEffectiveKey("workbench.action.openGlobalKeybindings", "Ctrl+K Ctrl+S");
      if (kbShortcutsKey !== "Ctrl+K Ctrl+S" && matchesKeyCombo(e, kbShortcutsKey)) {
        e.preventDefault();
        openShortcutsModal();
        return;
      }

      // Toggle Sidebar
      const sidebarKey = getEffectiveKey("workbench.action.toggleSidebarVisibility", "Ctrl+B");
      if (matchesKeyCombo(e, sidebarKey)) {
        e.preventDefault();
        handlers.toggleSidebar();
        return;
      }

      // Toggle Panel
      const panelKey = getEffectiveKey("workbench.action.togglePanel", "Ctrl+J");
      if (matchesKeyCombo(e, panelKey)) {
        e.preventDefault();
        handlers.toggleDrawer();
        return;
      }

      // Toggle Integrated Terminal
      const termKey = getEffectiveKey("workbench.action.terminal.toggleTerminal", "Ctrl+`");
      if (matchesKeyCombo(e, termKey)) {
        e.preventDefault();
        handlers.openDrawerTab("terminal");
        return;
      }

      // New Terminal Session
      const newTermKey = getEffectiveKey("workbench.action.terminal.new", "Ctrl+Shift+`");
      if (matchesKeyCombo(e, newTermKey)) {
        e.preventDefault();
        handlers.newTerminalSession();
        return;
      }

      // Output Channel
      const outputKey = getEffectiveKey("workbench.action.output.show", "Ctrl+Shift+U");
      if (matchesKeyCombo(e, outputKey)) {
        e.preventDefault();
        handlers.openDrawerTab("output");
        return;
      }

      // Problems / Doctor
      const problemsKey = getEffectiveKey("workbench.action.problems.show", "Ctrl+Shift+M");
      if (matchesKeyCombo(e, problemsKey)) {
        e.preventDefault();
        handlers.navigateToTab("doctor");
        return;
      }

      // Explorer
      const explorerKey = getEffectiveKey("workbench.view.explorer", "Ctrl+Shift+E");
      if (matchesKeyCombo(e, explorerKey)) {
        e.preventDefault();
        handlers.navigateToTab("explorer");
        return;
      }

      // SCM / Git
      const scmKey = getEffectiveKey("workbench.view.scm", "Ctrl+Shift+G");
      if (matchesKeyCombo(e, scmKey)) {
        e.preventDefault();
        handlers.navigateToTab("scm");
        return;
      }

      // Doctor
      const doctorKey = getEffectiveKey("workbench.view.doctor", "Ctrl+Shift+D");
      if (matchesKeyCombo(e, doctorKey)) {
        e.preventDefault();
        handlers.navigateToTab("doctor");
        return;
      }

      // Extensions / Marketplace
      const extKey = getEffectiveKey("workbench.view.extensions", "Ctrl+Shift+X");
      if (matchesKeyCombo(e, extKey)) {
        e.preventDefault();
        handlers.navigateToTab("marketplace");
        return;
      }

      // Scope Switcher
      const scopeKey = getEffectiveKey("workbench.action.toggleScope", "Alt+S");
      if (matchesKeyCombo(e, scopeKey)) {
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

      // Editor: Save All Files
      const saveAllKey = getEffectiveKey("workbench.action.files.saveAll", "Ctrl+Shift+S");
      if (matchesKeyCombo(e, saveAllKey)) {
        e.preventDefault();
        saveAllTabs();
        return;
      }

      // Editor: Save Active File
      const saveKey = getEffectiveKey("workbench.action.files.save", "Ctrl+S");
      if (matchesKeyCombo(e, saveKey)) {
        e.preventDefault();
        saveActiveEditorTab();
        return;
      }

      // Editor: Close Active Tab
      const closeTabKey = getEffectiveKey("workbench.action.closeActiveEditor", "Ctrl+W / Ctrl+F4");
      if (matchesKeyCombo(e, closeTabKey) && !isInput) {
        if (activeTabId.value) {
          e.preventDefault();
          closeActiveEditorTab();
          return;
        }
      }

      // Editor: Next / Previous Tab
      const nextTabKey = getEffectiveKey("workbench.action.nextEditor", "Ctrl+PageDown");
      if (matchesKeyCombo(e, nextTabKey)) {
        e.preventDefault();
        nextEditorTab();
        return;
      }

      const prevTabKey = getEffectiveKey("workbench.action.previousEditor", "Ctrl+PageUp");
      if (matchesKeyCombo(e, prevTabKey)) {
        e.preventDefault();
        prevEditorTab();
        return;
      }

      // Ctrl+K Chord Initiator (Default VS Code Chord)
      if (mod && !e.shiftKey && !e.altKey && (key === "k" || key === "K")) {
        e.preventDefault();
        if (handlers.isCommandPaletteOpen()) {
          handlers.closeCommandPalette();
          return;
        }
        pendingChord = "Ctrl+K";
        chordTimer = setTimeout(() => {
          if (pendingChord === "Ctrl+K") {
            clearChord();
            handlers.openCommandPalette();
          }
        }, 380);
        return;
      }
    };
  }

  return {
    isShortcutsModalOpen,
    openShortcutsModal,
    closeShortcutsModal,
    allShortcuts: computed(() => SHORTCUT_DEFINITIONS),
    activeShortcuts,
    customKeybindings,
    keybindingsPath,
    isKeybindingsLoading,
    rawKeybindingsJson,
    fetchKeybindings,
    saveKeybinding,
    resetKeybinding,
    resetAllKeybindings,
    saveRawKeybindings,
    createKeydownHandler,
    nextEditorTab,
    prevEditorTab,
    closeActiveEditorTab,
    saveActiveEditorTab,
  };
}
