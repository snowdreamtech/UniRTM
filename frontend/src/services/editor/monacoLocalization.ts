import { currentLocale } from "../../i18n";

/**
 * Monaco Editor 上下文菜单与浮动组件的标准多语言映射表（对齐 VS Code 官方译名）
 */
const MONACO_LABEL_TRANSLATIONS: Record<string, Record<string, string>> = {
  "zh-CN": {
    "Change All Occurrences": "更改所有匹配项",
    "Cut": "剪切",
    "Copy": "复制",
    "Paste": "粘贴",
    "Command Palette": "命令面板...",
    "Command Palette...": "命令面板...",
    "Format Document": "格式化文档",
    "Format Selection": "格式化选定内容",
    "Refactor...": "重构...",
    "Source Action...": "源代码操作...",
    "Rename Symbol": "重命名符号",
    "Go to Definition": "转到定义",
    "Peek Definition": "查看定义",
    "Go to References": "转到引用",
    "Peek References": "查看引用",
    "Go to Implementation": "转到实现",
    "Peek Implementation": "查看实现",
    "Go to Type Definition": "转到类型定义",
    "Peek Type Definition": "查看类型定义",
    "Find All References": "查找所有引用",
    "Share": "共享",
    "Select All": "全选",
    "Undo": "撤销",
    "Redo": "恢复",
    "Cut Line": "剪切行",
    "Copy Line Up": "向上复制行",
    "Copy Line Down": "向下复制行",
    "Move Line Up": "向上移动行",
    "Move Line Down": "向下移动行",
    "Delete Line": "删除行",
    "Toggle Line Comment": "切换行注释",
    "Toggle Block Comment": "切换块注释",
    "Fold": "折叠",
    "Unfold": "展开",
    "Fold All": "全部折叠",
    "Unfold All": "全部展开",
    "Quick Fix...": "快速修复...",
    "Show Call Hierarchy": "查看调用层次结构",
    "Peek Call Hierarchy": "速览调用层次结构",
  },
  "zh-TW": {
    "Change All Occurrences": "變更所有相符項",
    "Cut": "剪下",
    "Copy": "複製",
    "Paste": "貼上",
    "Command Palette": "命令選擇區...",
    "Command Palette...": "命令選擇區...",
    "Format Document": "格式化文件",
    "Format Selection": "格式化選取項目",
    "Refactor...": "重構...",
    "Source Action...": "原始程式碼動作...",
    "Rename Symbol": "重新命名符號",
    "Go to Definition": "前往定義",
    "Peek Definition": "預覽定義",
    "Go to References": "前往參考",
    "Peek References": "預覽參考",
    "Go to Implementation": "前往實作",
    "Peek Implementation": "預覽實作",
    "Go to Type Definition": "前往類型定義",
    "Peek Type Definition": "預覽類型定義",
    "Find All References": "尋找所有參考",
    "Share": "共用",
    "Select All": "全選",
    "Undo": "復原",
    "Redo": "取消復原",
    "Cut Line": "剪下行",
    "Copy Line Up": "向上複製行",
    "Copy Line Down": "向下複製行",
    "Move Line Up": "向上移動行",
    "Move Line Down": "向下移動行",
    "Delete Line": "刪除行",
    "Toggle Line Comment": "切換行註解",
    "Toggle Block Comment": "切換區塊註解",
    "Fold": "摺疊",
    "Unfold": "展開",
    "Fold All": "全部摺疊",
    "Unfold All": "全部展開",
    "Quick Fix...": "快速修正...",
    "Show Call Hierarchy": "檢視呼叫階層",
    "Peek Call Hierarchy": "速覽呼叫階層",
  },
};

export function translateMonacoLabel(label: string, lang?: string): string {
  const currentLang = lang || currentLocale.value || "en-US";
  const dict = MONACO_LABEL_TRANSLATIONS[currentLang] || (currentLang.startsWith("zh") ? MONACO_LABEL_TRANSLATIONS["zh-CN"] : null);
  if (!dict) return label;
  const trimmed = label.trim();
  return dict[trimmed] || label;
}

// ---------------------------------------------------------------------------
// 获取高反差配色（深色模式 → 浅色菜单，浅色模式 → 深色菜单）
// ---------------------------------------------------------------------------
function getContrastMenuColors() {
  const isLight = typeof document !== "undefined"
    ? document.documentElement.getAttribute("data-theme") === "light"
    : false;
  return isLight
    ? { bg: "#1e1e1e", fg: "#cccccc", border: "#3c3c3c", shadow: "0 12px 32px rgba(0,0,0,0.55)", hoverBg: "#04395e", hoverFg: "#ffffff", sep: "#333333", keybinding: "#858585" }
    : { bg: "#ffffff", fg: "#1e293b", border: "#cbd5e1", shadow: "0 10px 30px -4px rgba(0,0,0,0.4)", hoverBg: "#0060c0", hoverFg: "#ffffff", sep: "#e2e8f0", keybinding: "#64748b" };
}

// ---------------------------------------------------------------------------
// 对单个 .context-view 节点注入实心反差样式 + 翻译标签
// ---------------------------------------------------------------------------
function patchContextView(contextView: HTMLElement): void {
  const c = getContrastMenuColors();

  // 1) 外层容器 (context-view 本身，或 monaco-menu-container)
  contextView.style.setProperty("background-color", c.bg, "important");
  contextView.style.setProperty("color", c.fg, "important");
  contextView.style.setProperty("border", `1px solid ${c.border}`, "important");
  contextView.style.setProperty("box-shadow", c.shadow, "important");
  contextView.style.setProperty("border-radius", "6px", "important");
  contextView.style.setProperty("overflow", "hidden", "important");

  // 2) 所有子层：scrollable-element, monaco-menu, action-bar
  const innerContainers = contextView.querySelectorAll<HTMLElement>(
    ".monaco-scrollable-element, .monaco-menu, .monaco-action-bar.vertical"
  );
  innerContainers.forEach((el) => {
    el.style.setProperty("background-color", c.bg, "important");
    el.style.setProperty("color", c.fg, "important");
  });

  // 3) 菜单项：覆盖 Monaco 的 inline style（applyStyle() 在 render/focus/blur 时设置）
  const menuItems = contextView.querySelectorAll<HTMLElement>(".action-menu-item");
  menuItems.forEach((item) => {
    item.style.setProperty("color", c.fg, "important");
    // 不设置 backgroundColor（让 hover 态由下面处理）
    if (item.style.backgroundColor) {
      item.style.removeProperty("background-color");
    }
  });

  // 4) keybinding
  const keybindings = contextView.querySelectorAll<HTMLElement>(".keybinding");
  keybindings.forEach((kb) => {
    kb.style.setProperty("color", c.keybinding, "important");
  });

  // 5) 分割线
  const separators = contextView.querySelectorAll<HTMLElement>(".action-label.separator");
  separators.forEach((sep) => {
    sep.style.setProperty("border-bottom-color", c.sep, "important");
  });

  // 6) 翻译所有 action-label
  const labels = contextView.querySelectorAll<HTMLElement>(".action-label:not(.separator)");
  labels.forEach((el) => {
    const raw = el.textContent?.trim();
    if (raw) {
      const translated = translateMonacoLabel(raw);
      if (translated !== raw) {
        el.textContent = translated;
      }
    }
  });
}

// ---------------------------------------------------------------------------
// 全局翻译（兼容自定义菜单和 Monaco 菜单）
// ---------------------------------------------------------------------------
export function translateActiveMonacoMenus(): void {
  if (typeof document === "undefined") return;
  document.querySelectorAll<HTMLElement>(".context-view").forEach(patchContextView);
}

// ---------------------------------------------------------------------------
// 本地化 Monaco Editor Action 实例的 label（首次打开有效）
// ---------------------------------------------------------------------------
export function localizeEditorActions(editor: any): void {
  if (!editor || typeof editor.getSupportedActions !== "function") return;
  try {
    const actions = editor.getSupportedActions();
    for (const act of actions) {
      if (act && act.label) {
        const translated = translateMonacoLabel(act.label);
        if (translated !== act.label) {
          act.label = translated;
        }
      }
    }
  } catch {
    // ignore
  }
}

// ---------------------------------------------------------------------------
// 核心：监听 context-view 插入并持续修补
// ---------------------------------------------------------------------------
let observer: MutationObserver | null = null;
let isEventsRegistered = false;

/**
 * 当检测到 .context-view 出现在 DOM 中时，启动一个短时高频轮询来持续修补。
 * 这是因为 Monaco 的 Menu 构造过程是：
 *   1. 创建 context-view 容器并 append 到 body
 *   2. 异步/帧内逐步 push action items → 触发 render → updateLabel → applyStyle
 * 我们需要在所有 action items 完成 render 之后再做修补。
 * 高频轮询 (每 20ms) 持续 800ms 确保覆盖 Monaco 的所有异步阶段。
 */
function startPatchCycle(): void {
  if (typeof document === "undefined") return;
  // 立即执行一次
  translateActiveMonacoMenus();
  let ticks = 0;
  const timer = setInterval(() => {
    translateActiveMonacoMenus();
    ticks++;
    if (ticks > 40) { // 40 × 20ms = 800ms
      clearInterval(timer);
    }
  }, 20);
}

export function setupMonacoLocalization(): void {
  if (typeof document === "undefined") return;

  if (!observer) {
    observer = new MutationObserver((mutations) => {
      let hasNewMenu = false;
      for (const mutation of mutations) {
        for (const node of mutation.addedNodes) {
          if (node.nodeType !== Node.ELEMENT_NODE) continue;
          const el = node as HTMLElement;
          // 检测 context-view 节点被插入 DOM
          if (el.classList?.contains("context-view") || el.querySelector?.(".context-view")) {
            hasNewMenu = true;
          }
        }
      }
      if (hasNewMenu) {
        startPatchCycle();
      }
    });

    observer.observe(document.documentElement, {
      childList: true,
      subtree: true,
    });
  }

  if (!isEventsRegistered && typeof window !== "undefined") {
    isEventsRegistered = true;

    // 右键、pointerdown(button=2)、Shift+F10、F1 触发时启动修补
    const triggerPatch = () => startPatchCycle();

    window.addEventListener("contextmenu", triggerPatch, { capture: true });
    window.addEventListener(
      "pointerdown",
      (e) => { if (e.button === 2) triggerPatch(); },
      { capture: true }
    );
    window.addEventListener(
      "keydown",
      (e) => { if ((e.shiftKey && e.key === "F10") || e.key === "F1") triggerPatch(); },
      { capture: true }
    );

    // 鼠标移入菜单区域时也修补（防止 focus/blur 时 Monaco 的 applyStyle 覆盖颜色）
    document.addEventListener(
      "pointerover",
      (e) => {
        const target = e.target as Element | null;
        if (target?.closest?.(".context-view, .monaco-menu")) {
          translateActiveMonacoMenus();
        }
      },
      { capture: true, passive: true }
    );
  }
}
