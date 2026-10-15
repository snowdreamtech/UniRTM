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
};

/**
 * Helper to check if a string is a workbench semantic icon key
 */
export function isWorkbenchSemanticIcon(id: string): id is WorkbenchIconKey {
  return typeof id === "string" && id.startsWith("workbench:") && id in workbenchSemanticIcons;
}
