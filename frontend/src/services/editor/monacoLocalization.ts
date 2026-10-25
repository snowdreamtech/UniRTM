// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { watch } from "vue";
import { t, currentLocale, loadedDictionaries, type TranslationDict } from "../../i18n";

/**
 * Maps standard Monaco Editor context menu labels to i18n translation keys.
 */
export const MONACO_LABEL_TO_KEY: Record<string, keyof TranslationDict> = {
  "Run in Active Terminal": "editor.run_in_terminal",
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
 * Global bidirectional registry mapping any known text (English, Simplified Chinese, Traditional Chinese, Japanese, etc.)
 * back to its canonical TranslationDict key.
 */
const LABEL_TO_KEY_CACHE = new Map<string, keyof TranslationDict>();

function registerLabelKey(rawLabel: string, key: keyof TranslationDict) {
  if (!rawLabel) return;
  const trimmed = rawLabel.trim();
  LABEL_TO_KEY_CACHE.set(trimmed, key);
  LABEL_TO_KEY_CACHE.set(trimmed.toLowerCase(), key);

  if (trimmed.endsWith("...")) {
    const withoutDots = trimmed.slice(0, -3).trim();
    LABEL_TO_KEY_CACHE.set(withoutDots, key);
    LABEL_TO_KEY_CACHE.set(withoutDots.toLowerCase(), key);
  } else {
    LABEL_TO_KEY_CACHE.set(trimmed + "...", key);
    LABEL_TO_KEY_CACHE.set((trimmed + "...").toLowerCase(), key);
  }
}

const MONACO_ACTION_KEYS = Object.values(MONACO_LABEL_TO_KEY);

function indexDictionary(dict: TranslationDict | undefined): void {
  if (!dict) return;
  for (let i = 0; i < MONACO_ACTION_KEYS.length; i++) {
    const key = MONACO_ACTION_KEYS[i];
    const val = dict[key];
    if (val) {
      registerLabelKey(val, key);
    }
  }
}

// 1. Register canonical English labels
for (const [englishLabel, key] of Object.entries(MONACO_LABEL_TO_KEY)) {
  registerLabelKey(englishLabel, key);
}

// 2. Reactively index any loaded dictionaries in O(1)

// 3. Register common Monaco NLS aliases using unicode escapes
registerLabelKey("\u67e5\u770b\u5b9a\u4e49", "editor.action.peek_definition");
registerLabelKey("\u67e5\u770b\u5f15\u7528", "editor.action.peek_references");
registerLabelKey("\u67e5\u770b\u5b9e\u73b0", "editor.action.peek_implementation");
registerLabelKey("\u67e5\u770b\u7c7b\u578b\u5b9a\u4e49", "editor.action.peek_type_definition");

// 3. Reactively index any loaded dictionaries in O(1)
watch(
  () => Object.values(loadedDictionaries.value),
  (dicts) => {
    for (let i = 0; i < dicts.length; i++) {
      indexDictionary(dicts[i]);
    }
  },
  { immediate: true, deep: false }
);

/**
 * Resolves an arbitrary context menu label (in ANY language) back to its TranslationDict key in O(1).
 */
export function getActionKeyForLabel(label: string): keyof TranslationDict | undefined {
  if (!label) return undefined;
  const trimmed = label.trim();
  const direct = LABEL_TO_KEY_CACHE.get(trimmed) || LABEL_TO_KEY_CACHE.get(trimmed.toLowerCase());
  if (direct) return direct;

  if (trimmed.endsWith("...")) {
    const withoutDots = trimmed.slice(0, -3).trim();
    return LABEL_TO_KEY_CACHE.get(withoutDots) || LABEL_TO_KEY_CACHE.get(withoutDots.toLowerCase());
  }

  return undefined;
}

/**
 * Translates an action label in ANY language into the currently selected locale.
 */
export function translateMonacoLabel(label: string): string {
  const trimmed = label.trim();
  const key = getActionKeyForLabel(trimmed);
  if (key) {
    return t(key) || trimmed;
  }
  return trimmed;
}

/**
 * Action icon mapping to Codicon symbols (100% aligned with FileContextMenu system).
 */
export const ACTION_KEY_ICON_MAP: Record<string, string> = {
  "editor.run_in_terminal": "codicon-terminal",
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
  const key = getActionKeyForLabel(trimmed);
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
 * Localizes action label properties registered on the Monaco Standalone Editor instance.
 */
export function localizeEditorActions(editor: any): void {
  if (!editor) return;
  try {
    const termAct = typeof editor.getAction === "function" ? editor.getAction("unirtm.sendSelectionToTerminal") : null;
    if (termAct) {
      const termLabel = t("editor.run_in_terminal") || "Run in Active Terminal";
      termAct.label = termLabel;
      if (termAct._descriptor) {
        termAct._descriptor.label = termLabel;
      }
    }

    if (typeof editor.getSupportedActions === "function") {
      const actions = editor.getSupportedActions();
      for (const act of actions) {
        if (act && act.label) {
          if (!act._actionKey) {
            act._actionKey = getActionKeyForLabel(act.label);
          }
          if (act._actionKey) {
            const translated = t(act._actionKey);
            if (translated && act.label !== translated) {
              act.label = translated;
              if (act._descriptor) {
                act._descriptor.label = translated;
              }
            }
          }
        }
      }
    }
  } catch {
    // ignore
  }
}

let isTranslating = false;

/**
 * Translates all active Monaco context menu and submenu items in the DOM,
 * injecting unified Codicon menu icons aligned with the workspace design system.
 */
export function translateActiveMonacoMenus(root: ParentNode = document): void {
  if (isTranslating) return;
  isTranslating = true;
  try {
    const labels = root.querySelectorAll<HTMLElement>(
      ".context-view .action-label:not(.separator), .monaco-menu .action-label:not(.separator)"
    );
    if (!labels.length) return;

    for (let i = 0; i < labels.length; i++) {
      const el = labels[i];
      let actionKey = el.dataset.actionKey as keyof TranslationDict | undefined;
      const currentText = (el.textContent || "").trim();

      if (!actionKey && currentText) {
        actionKey = getActionKeyForLabel(currentText);
        if (actionKey) {
          el.dataset.actionKey = actionKey;
        }
      }

      if (actionKey) {
        const translated = t(actionKey);
        if (translated && el.textContent !== translated) {
          el.textContent = translated;
        }
      }

      const parent = el.parentElement;
      if (parent && (parent.classList.contains("action-menu-item") || parent.matches?.("a[role='menuitem']"))) {
        const iconClass =
          actionKey && ACTION_KEY_ICON_MAP[actionKey] ? ACTION_KEY_ICON_MAP[actionKey] : getIconForLabel(currentText);
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
  } finally {
    isTranslating = false;
  }
}

let globalMenuObserver: MutationObserver | null = null;

function ensureGlobalMenuObserver(): void {
  if (globalMenuObserver || typeof document === "undefined") return;
  globalMenuObserver = new MutationObserver((mutations) => {
    if (isTranslating) return;
    let shouldTranslate = false;
    for (let i = 0; i < mutations.length; i++) {
      const mut = mutations[i];
      if (mut.addedNodes.length > 0) {
        for (let j = 0; j < mut.addedNodes.length; j++) {
          const node = mut.addedNodes[j];
          if (node.nodeType === Node.ELEMENT_NODE) {
            const el = node as HTMLElement;
            if (
              el.classList?.contains("context-view") ||
              el.classList?.contains("monaco-menu-container") ||
              el.classList?.contains("monaco-menu") ||
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
  });

  // Watch locale changes to automatically re-localize any open context menus
  watch(currentLocale, () => {
    translateActiveMonacoMenus();
  });
}

/**
 * Initializes Monaco NLS module if available, and starts global context menu observer.
 */
export async function setupMonacoLocalization(): Promise<void> {
  ensureGlobalMenuObserver();
}

/**
 * Binds context menu localization capture hook to an editor container.
 */
export function setupEditorContextMenuLocalization(editor: any, container: HTMLElement): void {
  if (!container || !editor) return;
  ensureGlobalMenuObserver();

  const handleContextMenu = () => {
    requestAnimationFrame(() => {
      translateActiveMonacoMenus();
    });
  };

  container.addEventListener("contextmenu", handleContextMenu, { capture: true, passive: true });
}
