// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import type { SemanticIconMapping, WorkbenchIconKey } from "../../types/icon-theme";

/**
 * Workbench Semantic Icon Dictionary.
 * Maps high-level UI intents to candidate file/language/tool identifiers across standard VS Code icon themes,
 * with precise Seti UI fallback keys when missing.
 */
export const workbenchSemanticIcons: Record<WorkbenchIconKey, SemanticIconMapping> = {
  "workbench:dashboard": {
    candidates: ["dashboard", "chart", "analytics", "status", "overview", "json"],
    fallbackSetiKey: "_json",
  },
  "workbench:explorer": {
    candidates: ["explorer", "files", "folder", "default_folder", "directory"],
    fallbackSetiKey: "_default",
  },
  "workbench:runtimes": {
    candidates: ["runtimes", "globe", "environment", "html", "javascript"],
    fallbackSetiKey: "_html",
  },
  "workbench:tools": {
    candidates: ["tools", "toolchain", "build", "makefile", "c"],
    fallbackSetiKey: "_c",
  },
  "workbench:tasks": {
    candidates: ["tasks", "task", "sh", "bash", "shell", "terminal"],
    fallbackSetiKey: "_shell",
  },
  "workbench:marketplace": {
    candidates: ["marketplace", "extensions", "plugin", "package.json", "zip"],
    fallbackSetiKey: "_zip",
  },
  "workbench:doctor": {
    candidates: ["doctor", "check", "test", "spec", "search"],
    fallbackSetiKey: "_search",
  },
  "workbench:settings": {
    candidates: ["settings", "config", "settings.json", "preferences"],
    fallbackSetiKey: "_settings",
  },
  "workbench:about": {
    candidates: ["about", "info", "readme.md", "markdown"],
    fallbackSetiKey: "_info",
  },
  "workbench:terminal": {
    candidates: ["terminal", "console", "sh", "bash", "shell"],
    fallbackSetiKey: "_shell",
  },
  "workbench:logs": {
    candidates: ["logs", "log", "output", "text", "markdown"],
    fallbackSetiKey: "_markdown",
  },
  "workbench:lock": {
    candidates: ["lock", "unirtm.lock", "package-lock.json", "yarn.lock"],
    fallbackSetiKey: "_lock",
  },
  "workbench:environments": {
    candidates: ["variable-group", "environment", "env", "dotenv"],
    fallbackSetiKey: "_config",
  },
  "workbench:variable": {
    candidates: ["symbol-variable", "variable", "var"],
    fallbackSetiKey: "_config",
  },
  "workbench:key": {
    candidates: ["key", "symbol-key", "lock", "secret"],
    fallbackSetiKey: "_lock",
  },
  "workbench:eye": {
    candidates: ["eye", "preview", "view"],
    fallbackSetiKey: "_info",
  },
  "workbench:eye-closed": {
    candidates: ["eye-closed", "hide", "conceal"],
    fallbackSetiKey: "_info",
  },
  "workbench:copy": {
    candidates: ["copy", "clone", "clipboard"],
    fallbackSetiKey: "_default",
  },
  "workbench:config": {
    candidates: ["config", "unirtm.toml", "configuration", "settings.json"],
    fallbackSetiKey: "_config",
  },
  "workbench:success": {
    candidates: ["success", "pass", "valid", "check"],
    fallbackSetiKey: "_search",
  },
  "workbench:warning": {
    candidates: ["warning", "alert", "warn"],
    fallbackSetiKey: "_settings",
  },
  "workbench:error": {
    candidates: ["error", "fail", "danger"],
    fallbackSetiKey: "_default",
  },
  "workbench:info": {
    candidates: ["info", "information", "hint"],
    fallbackSetiKey: "_info",
  },
  "workbench:sync": {
    candidates: ["sync", "refresh", "reload", "git", "git_commit"],
    fallbackSetiKey: "_git",
  },
  "workbench:search": {
    candidates: ["search", "find", "doctor", "zoom"],
    fallbackSetiKey: "_search",
  },
  "workbench:delete": {
    candidates: ["delete", "trash", "remove"],
    fallbackSetiKey: "_default",
  },
  "workbench:folder": {
    candidates: ["folder", "directory", "dir", "files"],
    fallbackSetiKey: "_default",
  },
  "workbench:package": {
    candidates: ["package", "zip", "bundle", "tar", "box"],
    fallbackSetiKey: "_zip",
  },
  "workbench:globe": {
    candidates: ["globe", "network", "internet", "http", "html"],
    fallbackSetiKey: "_html",
  },
  "workbench:speed": {
    candidates: ["speed", "lightning", "zap", "performance", "flash"],
    fallbackSetiKey: "_shell",
  },
  "workbench:star": {
    candidates: ["star", "favorite", "json"],
    fallbackSetiKey: "_json",
  },
  "workbench:user": {
    candidates: ["user", "account", "author", "person"],
    fallbackSetiKey: "_info",
  },
  "workbench:edit": {
    candidates: ["edit", "pencil", "write", "config", "settings"],
    fallbackSetiKey: "_settings",
  },
  "workbench:github": {
    candidates: ["github", "git", "repo", "repository"],
    fallbackSetiKey: "_git",
  },
  "workbench:sparkle": {
    candidates: ["sparkle", "ai", "magic", "smart", "json"],
    fallbackSetiKey: "_json",
  },
  "workbench:cards": {
    candidates: ["editor-layout", "layout", "grid", "json"],
    fallbackSetiKey: "_default",
  },
  "workbench:list": {
    candidates: ["table", "list-flat", "list", "text"],
    fallbackSetiKey: "_default",
  },
};

/**
 * Helper to check if a string is a workbench semantic icon key
 */
export function isWorkbenchSemanticIcon(id: string): id is WorkbenchIconKey {
  return typeof id === "string" && id.startsWith("workbench:") && id in workbenchSemanticIcons;
}

/**
 * Built-in high fidelity SVG fallbacks for system workbench actions
 * when the active file icon theme does not provide a dedicated action glyph.
 */
export const WORKBENCH_BUILTIN_SVGS: Partial<Record<WorkbenchIconKey, string>> = {
  "workbench:delete":
    "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 16 16' width='16' height='16' fill='%23ef4444'%3E%3Cpath fill-rule='evenodd' clip-rule='evenodd' d='M10 3h3v1h-1v9a1 1 0 0 1-1 1H5a1 1 0 0 1-1-1V4H3V3h3V2a1 1 0 0 1 1-1h2a1 1 0 0 1 1 1v1zM7 2v1h2V2H7zm2 3H8v7h1V5zm-2 0H6v7h1V5z'/%3E%3C/svg%3E",
};
