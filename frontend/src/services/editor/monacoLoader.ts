// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import editorWorker from "monaco-editor/esm/vs/editor/editor.worker?worker";
import jsonWorker from "monaco-editor/esm/vs/language/json/json.worker?worker";
import cssWorker from "monaco-editor/esm/vs/language/css/css.worker?worker";
import htmlWorker from "monaco-editor/esm/vs/language/html/html.worker?worker";
import tsWorker from "monaco-editor/esm/vs/language/typescript/ts.worker?worker";

let isEnvConfigured = false;

export function setupMonacoEnvironment(): void {
  if (isEnvConfigured || typeof window === "undefined") return;

  (window as any).MonacoEnvironment = {
    getWorker(_: any, label: string) {
      if (label === "json") {
        return new jsonWorker();
      }
      if (label === "css" || label === "scss" || label === "less") {
        return new cssWorker();
      }
      if (label === "html" || label === "handlebars" || label === "razor") {
        return new htmlWorker();
      }
      if (label === "typescript" || label === "javascript") {
        return new tsWorker();
      }
      return new editorWorker();
    },
  };

  isEnvConfigured = true;
}

let monacoPromise: Promise<typeof import("monaco-editor")> | null = null;

/**
 * Asynchronously load the full Monaco Editor on-demand.
 * 100% offline via local Vite worker bundles.
 */
export function loadMonaco(): Promise<typeof import("monaco-editor")> {
  if (!monacoPromise) {
    monacoPromise = (async () => {
      setupMonacoEnvironment();
      const monaco = await import("monaco-editor");
      return monaco;
    })();
  }
  return monacoPromise;
}

const EXTENSION_LANGUAGE_MAP: Record<string, string> = {
  go: "go",
  ts: "typescript",
  tsx: "typescript",
  js: "javascript",
  jsx: "javascript",
  mjs: "javascript",
  cjs: "javascript",
  json: "json",
  jsonc: "json",
  vue: "html",
  html: "html",
  htm: "html",
  css: "css",
  scss: "scss",
  less: "less",
  py: "python",
  pyw: "python",
  rs: "rust",
  md: "markdown",
  markdown: "markdown",
  yml: "yaml",
  yaml: "yaml",
  toml: "ini",
  ini: "ini",
  cfg: "ini",
  sh: "shell",
  bash: "shell",
  zsh: "shell",
  bat: "bat",
  cmd: "bat",
  ps1: "powershell",
  psm1: "powershell",
  sql: "sql",
  xml: "xml",
  svg: "xml",
  c: "c",
  cpp: "cpp",
  h: "c",
  hpp: "cpp",
  java: "java",
  lua: "lua",
  php: "php",
  rb: "ruby",
  swift: "swift",
  kt: "kotlin",
};

const FILENAME_LANGUAGE_MAP: Record<string, string> = {
  dockerfile: "dockerfile",
  makefile: "makefile",
  ".gitignore": "ignore",
  ".npmrc": "ini",
  ".editorconfig": "ini",
  ".env": "ini",
};

export function getLanguageForFile(filename: string): string {
  if (!filename) return "plaintext";

  const lower = filename.toLowerCase();
  if (FILENAME_LANGUAGE_MAP[lower]) {
    return FILENAME_LANGUAGE_MAP[lower];
  }

  const lastDot = lower.lastIndexOf(".");
  if (lastDot >= 0) {
    const ext = lower.slice(lastDot + 1);
    if (EXTENSION_LANGUAGE_MAP[ext]) {
      return EXTENSION_LANGUAGE_MAP[ext];
    }
  }

  return "plaintext";
}
