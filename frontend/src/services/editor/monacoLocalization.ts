// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { t, currentLocale, type TranslationDict } from "../../i18n";

/**
 * Maps standard Monaco Editor context menu labels to i18n translation keys.
 */
const MONACO_LABEL_TO_KEY: Record<string, keyof TranslationDict> = {
  "Change All Occurrences": "editor.action.change_all",
  Cut: "editor.action.cut",
  Copy: "editor.action.copy",
  Paste: "editor.action.paste",
  "Command Palette": "editor.action.command_palette",
  "Command Palette...": "editor.action.command_palette",
  "Format Document": "editor.action.format_document",
  "Format Selection": "editor.action.format_selection",
  "Refactor...": "editor.action.refactor",
  "Source Action...": "editor.action.source_action",
  "Rename Symbol": "editor.action.rename_symbol",
  "Go to Definition": "editor.action.goto_definition",
  "Peek Definition": "editor.action.peek_definition",
  "Go to References": "editor.action.goto_references",
  "Peek References": "editor.action.peek_references",
  "Go to Implementation": "editor.action.goto_implementation",
  "Peek Implementation": "editor.action.peek_implementation",
  "Go to Type Definition": "editor.action.goto_type_definition",
  "Peek Type Definition": "editor.action.peek_type_definition",
  "Find All References": "editor.action.find_all_references",
  Share: "editor.action.share",
  "Select All": "editor.action.select_all",
  Undo: "editor.action.undo",
  Redo: "editor.action.redo",
  "Cut Line": "editor.action.cut_line",
  "Copy Line Up": "editor.action.copy_line_up",
  "Copy Line Down": "editor.action.copy_line_down",
  "Move Line Up": "editor.action.move_line_up",
  "Move Line Down": "editor.action.move_line_down",
  "Delete Line": "editor.action.delete_line",
  "Toggle Line Comment": "editor.action.toggle_line_comment",
  "Toggle Block Comment": "editor.action.toggle_block_comment",
  Fold: "editor.action.fold",
  Unfold: "editor.action.unfold",
  "Fold All": "editor.action.fold_all",
  "Unfold All": "editor.action.unfold_all",
  "Quick Fix...": "editor.action.quick_fix",
  "Show Call Hierarchy": "editor.action.show_call_hierarchy",
  "Peek Call Hierarchy": "editor.action.peek_call_hierarchy",
};

/**
 * Reverse mapping from known action translations back to English canonical label,
 * ensuring repeated context menu calls or DOM queries don't lose the original key.
 */
const REVERSE_LABEL_MAP: Record<string, string> = {};
for (const canonicalLabel of Object.keys(MONACO_LABEL_TO_KEY)) {
  REVERSE_LABEL_MAP[canonicalLabel] = canonicalLabel;
  REVERSE_LABEL_MAP[canonicalLabel.toLowerCase()] = canonicalLabel;
}

export function translateMonacoLabel(label: string): string {
  const trimmed = label.trim();
  const canonical = REVERSE_LABEL_MAP[trimmed] || REVERSE_LABEL_MAP[trimmed.toLowerCase()] || trimmed;
  const key = MONACO_LABEL_TO_KEY[canonical];
  if (key) {
    return t(key) || canonical;
  }
  return trimmed;
}

/**
 * Localizes action label properties registered on the Monaco Standalone Editor instance.
 */
export function localizeEditorActions(editor: any): void {
  if (!editor || typeof editor.getSupportedActions !== "function") return;
  try {
    const actions = editor.getSupportedActions();
    for (const act of actions) {
      if (act && act.label) {
        if (!act._origLabel) {
          act._origLabel = act.label;
        }
        const translated = translateMonacoLabel(act._origLabel);
        if (translated) {
          act.label = translated;
        }
      }
    }
  } catch {
    // ignore
  }
}

/**
 * Action icon mapping to Codicon symbols (100% aligned with FileContextMenu system).
 */
const ACTION_KEY_ICON_MAP: Record<string, string> = {
  "editor.action.change_all": "codicon-edit",
  "editor.action.cut": "codicon-scissors",
  "editor.action.copy": "codicon-copy",
  "editor.action.paste": "codicon-clippy",
  "editor.action.command_palette": "codicon-terminal",
  "editor.action.format_document": "codicon-code",
  "editor.action.format_selection": "codicon-code",
  "editor.action.refactor": "codicon-tools",
  "editor.action.source_action": "codicon-source-control",
  "editor.action.rename_symbol": "codicon-edit",
  "editor.action.goto_definition": "codicon-go-to-file",
  "editor.action.peek_definition": "codicon-preview",
  "editor.action.goto_references": "codicon-references",
  "editor.action.peek_references": "codicon-preview",
  "editor.action.goto_implementation": "codicon-symbol-event",
  "editor.action.peek_implementation": "codicon-preview",
  "editor.action.goto_type_definition": "codicon-symbol-class",
  "editor.action.peek_type_definition": "codicon-preview",
  "editor.action.find_all_references": "codicon-search",
  "editor.action.share": "codicon-symbol-misc",
  "editor.action.select_all": "codicon-selection",
  "editor.action.undo": "codicon-discard",
  "editor.action.redo": "codicon-redo",
  "editor.action.cut_line": "codicon-scissors",
  "editor.action.copy_line_up": "codicon-copy",
  "editor.action.copy_line_down": "codicon-copy",
  "editor.action.move_line_up": "codicon-arrow-up",
  "editor.action.move_line_down": "codicon-arrow-down",
  "editor.action.delete_line": "codicon-trash",
  "editor.action.toggle_line_comment": "codicon-comment",
  "editor.action.toggle_block_comment": "codicon-comment-discussion",
  "editor.action.fold": "codicon-chevron-right",
  "editor.action.unfold": "codicon-chevron-down",
  "editor.action.fold_all": "codicon-collapse-all",
  "editor.action.unfold_all": "codicon-expand-all",
  "editor.action.quick_fix": "codicon-lightbulb",
  "editor.action.show_call_hierarchy": "codicon-type-hierarchy",
  "editor.action.peek_call_hierarchy": "codicon-type-hierarchy",
};

export function getIconForLabel(label: string): string {
  const trimmed = label.trim();
  const canonical = REVERSE_LABEL_MAP[trimmed] || REVERSE_LABEL_MAP[trimmed.toLowerCase()] || trimmed;
  const key = MONACO_LABEL_TO_KEY[canonical];
  if (key && ACTION_KEY_ICON_MAP[key]) {
    return ACTION_KEY_ICON_MAP[key];
  }

  const lower = trimmed.toLowerCase();
  if (lower.includes("terminal") || lower.includes("run")) return "codicon-terminal";
  if (lower.includes("cut")) return "codicon-scissors";
  if (lower.includes("copy")) return "codicon-copy";
  if (lower.includes("paste")) return "codicon-clippy";
  if (lower.includes("command")) return "codicon-terminal";
  if (lower.includes("definition")) return "codicon-go-to-file";
  if (lower.includes("reference")) return "codicon-references";
  if (lower.includes("format")) return "codicon-code";
  if (lower.includes("fix")) return "codicon-lightbulb";
  if (lower.includes("fold")) return "codicon-collapse-all";
  if (lower.includes("unfold")) return "codicon-expand-all";
  if (lower.includes("comment")) return "codicon-comment";
  if (lower.includes("delete")) return "codicon-trash";
  if (lower.includes("find") || lower.includes("search")) return "codicon-search";
  if (lower.includes("rename")) return "codicon-edit";
  return "codicon-symbol-misc";
}

/**
 * Translates all active Monaco context menu and submenu items in the DOM,
 * injecting unified Codicon menu icons aligned with the workspace design system.
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

      const parent = el.parentElement;
      if (parent && (parent.classList.contains("action-menu-item") || parent.matches?.("a[role='menuitem']"))) {
        const iconClass = getIconForLabel(orig);
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
 * Initializes Monaco NLS module if available, and starts global context menu observer.
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
 * Binds context menu localization capture hook to an editor container.
 */
export function setupEditorContextMenuLocalization(editor: any, container: HTMLElement): void {
  if (!container || !editor) return;
  ensureGlobalMenuObserver();

  const handleContextMenu = () => {
    localizeEditorActions(editor);
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
