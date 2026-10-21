import { currentLocale } from "../../i18n";

/**
 * Monaco Editor 上下文菜单与浮动组件的标准多语言映射表（对齐 VS Code 官方译名）
 */
const MONACO_LABEL_TRANSLATIONS: Record<string, Record<string, string>> = {
  "zh-CN": {
    "Change All Occurrences": "更改所有匹配项",
    Cut: "剪切",
    Copy: "复制",
    Paste: "粘贴",
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
    Share: "共享",
    "Select All": "全选",
    Undo: "撤销",
    Redo: "恢复",
    "Cut Line": "剪切行",
    "Copy Line Up": "向上复制行",
    "Copy Line Down": "向下复制行",
    "Move Line Up": "向上移动行",
    "Move Line Down": "向下移动行",
    "Delete Line": "删除行",
    "Toggle Line Comment": "切换行注释",
    "Toggle Block Comment": "切换块注释",
    Fold: "折叠",
    Unfold: "展开",
    "Fold All": "全部折叠",
    "Unfold All": "全部展开",
    "Quick Fix...": "快速修复...",
    "Show Call Hierarchy": "查看调用层次结构",
    "Peek Call Hierarchy": "速览调用层次结构",
  },
  "zh-TW": {
    "Change All Occurrences": "變更所有相符項",
    Cut: "剪下",
    Copy: "複製",
    Paste: "貼上",
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
    Share: "共用",
    "Select All": "全選",
    Undo: "復原",
    Redo: "取消復原",
    "Cut Line": "剪下行",
    "Copy Line Up": "向上複製行",
    "Copy Line Down": "向下複製行",
    "Move Line Up": "向上移動行",
    "Move Line Down": "向下移動行",
    "Delete Line": "刪除行",
    "Toggle Line Comment": "切換行註解",
    "Toggle Block Comment": "切換區塊註解",
    Fold: "摺疊",
    Unfold: "展開",
    "Fold All": "全部摺疊",
    "Unfold All": "全部展開",
    "Quick Fix...": "快速修正...",
    "Show Call Hierarchy": "檢視呼叫階層",
    "Peek Call Hierarchy": "速覽呼叫階層",
  },
};

export function translateMonacoLabel(label: string, lang?: string): string {
  const currentLang = lang || currentLocale.value || "en-US";
  const dict =
    MONACO_LABEL_TRANSLATIONS[currentLang] ||
    (currentLang.startsWith("zh") ? MONACO_LABEL_TRANSLATIONS["zh-CN"] : null);
  if (!dict) return label;
  const trimmed = label.trim();
  return dict[trimmed] || label;
}

/**
 * 本地化指定 Monaco Standalone Editor 实例中注册的 Action 的 label 属性。
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
 * Monaco Editor 上下文菜单与常用操作的精致 Codicon 图标映射表（100% 对齐 FileContextMenu 体系）
 */
const ACTION_ICON_MAP: Record<string, string> = {
  // Terminal / Run
  "在活动终端中运行 (Run in Active Terminal)": "codicon-terminal",
  "Run in Active Terminal": "codicon-terminal",
  在活动终端中运行: "codicon-terminal",
  在集成终端中打开: "codicon-terminal",

  // Matching & Multi-cursor
  更改所有匹配项: "codicon-edit",
  "Change All Occurrences": "codicon-edit",
  變更所有相符項: "codicon-edit",

  // Clipboard
  剪切: "codicon-scissors",
  Cut: "codicon-scissors",
  剪下: "codicon-scissors",
  复制: "codicon-copy",
  Copy: "codicon-copy",
  複製: "codicon-copy",
  粘贴: "codicon-clippy",
  Paste: "codicon-clippy",
  貼上: "codicon-clippy",

  // Command Palette
  命令面板: "codicon-terminal",
  "命令面板...": "codicon-terminal",
  "Command Palette": "codicon-terminal",
  "Command Palette...": "codicon-terminal",
  "命令選擇區...": "codicon-terminal",

  // Format & Refactor
  格式化文档: "codicon-code",
  "Format Document": "codicon-code",
  格式化文件: "codicon-code",
  格式化选定内容: "codicon-code",
  "Format Selection": "codicon-code",
  格式化選取項目: "codicon-code",
  "重构...": "codicon-tools",
  "Refactor...": "codicon-tools",
  "重構...": "codicon-tools",
  "源代码操作...": "codicon-source-control",
  "Source Action...": "codicon-source-control",
  "原始程式碼動作...": "codicon-source-control",
  重命名符号: "codicon-edit",
  "Rename Symbol": "codicon-edit",
  重新命名符號: "codicon-edit",

  // Navigation & Definition
  转到定义: "codicon-go-to-file",
  "Go to Definition": "codicon-go-to-file",
  前往定義: "codicon-go-to-file",
  查看定义: "codicon-preview",
  "Peek Definition": "codicon-preview",
  預覽定義: "codicon-preview",
  转到引用: "codicon-references",
  "Go to References": "codicon-references",
  前往參考: "codicon-references",
  查看引用: "codicon-preview",
  "Peek References": "codicon-preview",
  預覽參考: "codicon-preview",
  转到实现: "codicon-symbol-event",
  "Go to Implementation": "codicon-symbol-event",
  前往實作: "codicon-symbol-event",
  查看实现: "codicon-preview",
  "Peek Implementation": "codicon-preview",
  預覽實作: "codicon-preview",
  转到类型定义: "codicon-symbol-class",
  "Go to Type Definition": "codicon-symbol-class",
  前往類型定義: "codicon-symbol-class",
  查看类型定义: "codicon-preview",
  "Peek Type Definition": "codicon-preview",
  預覽類型定義: "codicon-preview",
  查找所有引用: "codicon-search",
  "Find All References": "codicon-search",
  尋找所有參考: "codicon-search",

  // Quick Fix & Actions
  "快速修复...": "codicon-lightbulb",
  "Quick Fix...": "codicon-lightbulb",
  "快速修正...": "codicon-lightbulb",
  全选: "codicon-selection",
  "Select All": "codicon-selection",
  撤销: "codicon-discard",
  Undo: "codicon-discard",
  恢复: "codicon-redo",
  Redo: "codicon-redo",
  剪切行: "codicon-scissors",
  "Cut Line": "codicon-scissors",
  向上复制行: "codicon-copy",
  "Copy Line Up": "codicon-copy",
  向下复制行: "codicon-copy",
  "Copy Line Down": "codicon-copy",
  删除行: "codicon-trash",
  "Delete Line": "codicon-trash",
  切换行注释: "codicon-comment",
  "Toggle Line Comment": "codicon-comment",
  切换块注释: "codicon-comment-discussion",
  "Toggle Block Comment": "codicon-comment-discussion",
  全部折叠: "codicon-collapse-all",
  "Fold All": "codicon-collapse-all",
  全部展开: "codicon-expand-all",
  "Unfold All": "codicon-expand-all",
  查看调用层次结构: "codicon-type-hierarchy",
  "Show Call Hierarchy": "codicon-type-hierarchy",
  速览调用层次结构: "codicon-type-hierarchy",
  "Peek Call Hierarchy": "codicon-type-hierarchy",
};

export function getIconForLabel(label: string): string {
  const trimmed = label.trim();
  if (ACTION_ICON_MAP[trimmed]) return ACTION_ICON_MAP[trimmed];
  const lower = trimmed.toLowerCase();
  if (lower.includes("terminal") || lower.includes("终端")) return "codicon-terminal";
  if (lower.includes("cut") || lower.includes("剪切") || lower.includes("剪下")) return "codicon-scissors";
  if (lower.includes("copy") || lower.includes("复制") || lower.includes("複製")) return "codicon-copy";
  if (lower.includes("paste") || lower.includes("粘贴") || lower.includes("貼上")) return "codicon-clippy";
  if (lower.includes("command") || lower.includes("命令")) return "codicon-terminal";
  if (lower.includes("definition") || lower.includes("定义") || lower.includes("定義")) return "codicon-go-to-file";
  if (lower.includes("reference") || lower.includes("引用") || lower.includes("參考")) return "codicon-references";
  if (lower.includes("format") || lower.includes("格式化")) return "codicon-code";
  if (lower.includes("fix") || lower.includes("修复") || lower.includes("修正")) return "codicon-lightbulb";
  if (lower.includes("fold") || lower.includes("折叠") || lower.includes("摺疊")) return "codicon-collapse-all";
  if (lower.includes("unfold") || lower.includes("展开")) return "codicon-expand-all";
  if (lower.includes("comment") || lower.includes("注释") || lower.includes("註解")) return "codicon-comment";
  if (lower.includes("delete") || lower.includes("删除") || lower.includes("刪除")) return "codicon-trash";
  if (lower.includes("find") || lower.includes("search") || lower.includes("查找") || lower.includes("寻找"))
    return "codicon-search";
  if (lower.includes("rename") || lower.includes("重命名")) return "codicon-edit";
  return "codicon-symbol-misc";
}

/**
 * 翻译当前 DOM 中所有活动的 Monaco 菜单项（包括主右键菜单与子菜单），并注入 Codicon 菜单图标。
 * 存储原始文本到 dataset.origLabel，以便多次右键或切换语言时准确翻译。
 */
export function translateActiveMonacoMenus(root: ParentNode = document): void {
  try {
    const labels = root.querySelectorAll<HTMLElement>(
      ".context-view .action-label:not(.separator), .monaco-menu .action-label:not(.separator)"
    );
    for (let i = 0; i < labels.length; i++) {
      const el = labels[i];
      const orig = el.dataset.origLabel || el.textContent || "";
      if (!el.dataset.origLabel && orig) {
        el.dataset.origLabel = orig;
      }
      const translated = translateMonacoLabel(orig);
      if (translated && el.textContent !== translated) {
        el.textContent = translated;
      }

      // 注入与更新统一的精致 Codicon 菜单图标（对齐 FileContextMenu）
      const parent = el.parentElement;
      if (parent && (parent.classList.contains("action-menu-item") || parent.matches?.("a[role='menuitem']"))) {
        const iconClass = getIconForLabel(translated || orig);
        const existingIcon = parent.querySelector(":scope > .menu-icon");
        if (!existingIcon) {
          const iconSpan = document.createElement("span");
          iconSpan.className = `menu-icon codicon ${iconClass}`;
          parent.insertBefore(iconSpan, el);
        } else if (!existingIcon.classList.contains(iconClass)) {
          existingIcon.className = `menu-icon codicon ${iconClass}`;
        }
      }
    }
  } catch {
    // ignore
  }
}

let globalMenuObserver: MutationObserver | null = null;

function ensureGlobalMenuObserver(): void {
  if (globalMenuObserver || typeof document === "undefined") return;
  globalMenuObserver = new MutationObserver((mutations) => {
    let shouldTranslate = false;
    for (let i = 0; i < mutations.length; i++) {
      const mut = mutations[i];
      const target = mut.target as HTMLElement;
      if (
        (target && target.nodeType === Node.ELEMENT_NODE && target.closest?.(".context-view, .monaco-menu")) ||
        (target && target.parentElement?.closest?.(".context-view, .monaco-menu"))
      ) {
        shouldTranslate = true;
        break;
      }
      if (mut.addedNodes.length > 0) {
        for (let j = 0; j < mut.addedNodes.length; j++) {
          const node = mut.addedNodes[j];
          if (node.nodeType === Node.ELEMENT_NODE) {
            const el = node as HTMLElement;
            if (
              el.matches?.(".action-label, .action-menu-item, .monaco-menu, .context-view") ||
              el.classList?.contains("context-view") ||
              el.classList?.contains("monaco-menu-container") ||
              el.querySelector?.(".action-label")
            ) {
              shouldTranslate = true;
              break;
            }
          }
        }
      }
      if (shouldTranslate) break;
    }
    if (shouldTranslate) {
      translateActiveMonacoMenus();
    }
  });

  globalMenuObserver.observe(document.body, {
    childList: true,
    subtree: true,
    characterData: true,
  });
}

/**
 * 初始化 Monaco 官方 NLS 本地化模块并启动全局右键菜单翻译监听器。
 */
export async function setupMonacoLocalization(): Promise<void> {
  try {
    const lang = currentLocale.value || "en-US";
    if (lang === "zh-TW" || lang === "zh-HK" || lang === "zh-Hant") {
      // @ts-ignore
      await import("monaco-editor/min/vs/nls/lang/zh-tw.js");
    } else if (lang.startsWith("zh")) {
      // @ts-ignore
      await import("monaco-editor/min/vs/nls/lang/zh-cn.js");
    } else if (lang.startsWith("ja")) {
      // @ts-ignore
      await import("monaco-editor/min/vs/nls/lang/ja.js");
    } else if (lang.startsWith("ko")) {
      // @ts-ignore
      await import("monaco-editor/min/vs/nls/lang/ko.js");
    } else if (lang.startsWith("de")) {
      // @ts-ignore
      await import("monaco-editor/min/vs/nls/lang/de.js");
    } else if (lang.startsWith("fr")) {
      // @ts-ignore
      await import("monaco-editor/min/vs/nls/lang/fr.js");
    } else if (lang.startsWith("es")) {
      // @ts-ignore
      await import("monaco-editor/min/vs/nls/lang/es.js");
    } else if (lang.startsWith("ru")) {
      // @ts-ignore
      await import("monaco-editor/min/vs/nls/lang/ru.js");
    } else if (lang.startsWith("it")) {
      // @ts-ignore
      await import("monaco-editor/min/vs/nls/lang/it.js");
    } else if (lang === "pt-BR") {
      // @ts-ignore
      await import("monaco-editor/min/vs/nls/lang/pt-br.js");
    } else if (lang.startsWith("tr")) {
      // @ts-ignore
      await import("monaco-editor/min/vs/nls/lang/tr.js");
    } else if (lang.startsWith("pl")) {
      // @ts-ignore
      await import("monaco-editor/min/vs/nls/lang/pl.js");
    } else if (lang.startsWith("cs")) {
      // @ts-ignore
      await import("monaco-editor/min/vs/nls/lang/cs.js");
    }
  } catch (err) {
    console.warn("[monacoLocalization] Failed to load Monaco NLS messages:", err);
  }

  ensureGlobalMenuObserver();
}

/**
 * 为指定编辑器绑定右键菜单本地化捕获钩子。
 */
export function setupEditorContextMenuLocalization(editor: any, container: HTMLElement): void {
  if (!container || !editor) return;
  ensureGlobalMenuObserver();

  const handleContextMenu = () => {
    localizeEditorActions(editor);
    // 菜单由 Monaco 异步挂载到 DOM，多次微任务与动画帧轮询确保即时汉化，杜绝二次右键时变回英文
    translateActiveMonacoMenus();
    requestAnimationFrame(() => {
      translateActiveMonacoMenus();
      setTimeout(translateActiveMonacoMenus, 10);
      setTimeout(translateActiveMonacoMenus, 30);
      setTimeout(translateActiveMonacoMenus, 80);
      setTimeout(translateActiveMonacoMenus, 150);
    });
  };

  container.addEventListener("contextmenu", handleContextMenu, { capture: true });
}
