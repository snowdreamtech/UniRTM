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

let observer: MutationObserver | null = null;
let isEventsRegistered = false;

function translateSingleLabelElement(el: Element): void {
  const raw = el.textContent?.trim();
  if (raw) {
    const translated = translateMonacoLabel(raw);
    if (translated !== raw) {
      el.textContent = translated;
    }
  }
}

function translateTree(node: Node): void {
  if (node.nodeType === Node.TEXT_NODE) {
    const parent = (node as Text).parentElement;
    if (parent?.classList.contains("action-label")) {
      translateSingleLabelElement(parent);
    }
    return;
  }

  if (node.nodeType !== Node.ELEMENT_NODE) return;
  const el = node as Element;

  if (el.classList.contains("action-label")) {
    translateSingleLabelElement(el);
    return;
  }

  const actionLabels = el.querySelectorAll?.(".action-label");
  if (actionLabels && actionLabels.length > 0) {
    actionLabels.forEach((item) => translateSingleLabelElement(item));
  }
}

/**
 * 手动对当前已经展开或 DOM 中存在的菜单进行全量翻译检查
 */
export function translateActiveMonacoMenus(): void {
  if (typeof document === "undefined") return;
  const labels = document.querySelectorAll(".action-label");
  labels.forEach((item) => translateSingleLabelElement(item));
}

/**
 * 本地化指定 Monaco Standalone Editor 实例中注册的 Action 的 label 属性
 */
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

/**
 * 监听 Monaco 上下文菜单弹出并执行本地化（支持首次及后续每次弹出）
 */
export function setupMonacoLocalization(): void {
  if (typeof document === "undefined") return;

  if (!observer) {
    observer = new MutationObserver((mutations) => {
      for (const mutation of mutations) {
        if (mutation.type === "characterData") {
          const parent = mutation.target.parentElement;
          if (parent?.classList.contains("action-label")) {
            translateSingleLabelElement(parent);
          }
          continue;
        }

        for (const node of mutation.addedNodes) {
          translateTree(node);
        }
      }
    });

    observer.observe(document.body, {
      childList: true,
      subtree: true,
      characterData: true,
    });
  }

  if (!isEventsRegistered && typeof window !== "undefined") {
    isEventsRegistered = true;

    // 当用户触发右键上下文菜单时，多阶段检查确保后续第 2、3...次弹出均被精确汉化
    window.addEventListener(
      "contextmenu",
      () => {
        translateActiveMonacoMenus();
        requestAnimationFrame(() => {
          translateActiveMonacoMenus();
        });
        setTimeout(translateActiveMonacoMenus, 16);
        setTimeout(translateActiveMonacoMenus, 60);
        setTimeout(translateActiveMonacoMenus, 150);
      },
      { capture: true }
    );

    // 鼠标移入菜单时保持本地化
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

