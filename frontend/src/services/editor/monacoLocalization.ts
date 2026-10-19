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
    "Share": "共享",
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
    "Share": "共用",
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

function translateMenuLabels(container: Element): void {
  const labels = container.querySelectorAll(".action-label");
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

/**
 * 监听 Monaco 上下文菜单弹出并执行本地化
 */
export function setupMonacoLocalization(): void {
  if (typeof document === "undefined" || observer) return;

  observer = new MutationObserver((mutations) => {
    for (const mutation of mutations) {
      for (const node of mutation.addedNodes) {
        if (node.nodeType !== Node.ELEMENT_NODE) continue;
        const el = node as Element;
        if (el.classList.contains("context-view") || el.classList.contains("monaco-menu") || el.querySelector?.(".monaco-menu")) {
          translateMenuLabels(el);
        }
      }
    }
  });

  observer.observe(document.body, {
    childList: true,
    subtree: true,
  });
}

/**
 * 手动对当前已经展开的菜单进行一次翻译检查
 */
export function translateActiveMonacoMenus(): void {
  if (typeof document === "undefined") return;
  const menus = document.querySelectorAll(".context-view, .monaco-menu");
  menus.forEach((menu) => translateMenuLabels(menu));
}
