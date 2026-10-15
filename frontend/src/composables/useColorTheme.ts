// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { ref } from "vue";
import type { ThemeItem } from "../types/theme-manager";

const STYLE_TAG_ID = "vscode-color-theme-override";

/**
 * Strip comments and trailing commas from VS Code JSONC theme files
 */
function parseJSONC(text: string): Record<string, any> {
  const cleaned = text
    .replace(/\/\*[\s\S]*?\*\//g, "")
    .replace(/\/\/[^\n\r]*/g, "")
    .replace(/,\s*([\]}])/g, "$1");
  return JSON.parse(cleaned);
}

export function useColorTheme() {
  const currentThemeId = ref<string>("dark");

  const removeCustomThemeStyle = () => {
    const existing = document.getElementById(STYLE_TAG_ID);
    if (existing) {
      existing.remove();
    }
  };

  const applyBuiltinTheme = (themeId: string) => {
    removeCustomThemeStyle();
    currentThemeId.value = themeId;

    if (themeId === "system") {
      const prefersDark = window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)").matches;
      document.documentElement.setAttribute("data-theme", prefersDark ? "dark" : "light");
    } else {
      document.documentElement.setAttribute("data-theme", themeId === "light" ? "light" : "dark");
    }
  };

  const applyCustomVSCodeTheme = async (themeItem: ThemeItem) => {
    if (!themeItem.package_id || !themeItem.relative_path) {
      applyBuiltinTheme("dark");
      return;
    }

    try {
      const cleanRel = themeItem.relative_path.startsWith("./")
        ? themeItem.relative_path.slice(2)
        : themeItem.relative_path.startsWith("/")
        ? themeItem.relative_path.slice(1)
        : themeItem.relative_path;

      const url = `/api/themes/assets/color/${encodeURIComponent(themeItem.package_id)}/${cleanRel}`;
      const res = await fetch(url);
      if (!res.ok) {
        throw new Error(`Failed to fetch theme asset: ${res.statusText}`);
      }

      const text = await res.text();
      const themeData = parseJSONC(text);
      const colors = themeData.colors || {};

      // Determine UI theme base (light or dark)
      const isLight = themeItem.ui_theme === "vs" || themeData.type === "light";
      document.documentElement.setAttribute("data-theme", isLight ? "light" : "dark");

      // Map VS Code workbench tokens to CSS variables
      const cssRules: string[] = [];

      if (colors["editor.background"]) {
        cssRules.push(`--bg-color: ${colors["editor.background"]} !important;`);
      }
      if (colors["sideBar.background"] || colors["editorGroupHeader.tabsBackground"]) {
        const sideBg = colors["sideBar.background"] || colors["editorGroupHeader.tabsBackground"];
        cssRules.push(`--card-bg: ${sideBg} !important;`);
        cssRules.push(`--modal-bg: ${sideBg} !important;`);
      }
      if (colors["input.background"]) {
        cssRules.push(`--input-bg: ${colors["input.background"]} !important;`);
      }
      if (colors["activityBar.background"]) {
        cssRules.push(`--section-bg: ${colors["activityBar.background"]} !important;`);
      }
      if (colors["editor.foreground"] || colors["foreground"]) {
        const fg = colors["editor.foreground"] || colors["foreground"];
        cssRules.push(`--text-main: ${fg} !important;`);
      }
      if (colors["descriptionForeground"]) {
        cssRules.push(`--text-muted: ${colors["descriptionForeground"]} !important;`);
      }
      if (colors["button.background"] || colors["activityBarBadge.background"]) {
        const accent = colors["button.background"] || colors["activityBarBadge.background"];
        cssRules.push(`--accent-cyan: ${accent} !important;`);
        cssRules.push(`--accent-cyan-glow: ${accent}55 !important;`);
      }
      if (colors["focusBorder"] || colors["activityBar.activeBorder"]) {
        const borderActive = colors["focusBorder"] || colors["activityBar.activeBorder"];
        cssRules.push(`--card-border-active: ${borderActive} !important;`);
      }
      if (colors["sideBar.border"] || colors["panel.border"]) {
        const border = colors["sideBar.border"] || colors["panel.border"];
        cssRules.push(`--card-border: ${border} !important;`);
      }

      // Inject or update override style tag
      removeCustomThemeStyle();
      const styleEl = document.createElement("style");
      styleEl.id = STYLE_TAG_ID;
      styleEl.textContent = `:root, [data-theme] { ${cssRules.join(" ")} }`;
      document.head.appendChild(styleEl);

      currentThemeId.value = themeItem.id;
    } catch (err) {
      console.error("[useColorTheme] Failed to apply custom VS Code color theme:", err);
      applyBuiltinTheme(themeItem.ui_theme === "vs" ? "light" : "dark");
    }
  };

  const applyTheme = async (theme: string | ThemeItem, allThemes?: ThemeItem[]) => {
    if (typeof theme === "string") {
      if (theme === "dark" || theme === "light" || theme === "system") {
        applyBuiltinTheme(theme);
        return;
      }
      // If an ID was passed, find corresponding item in allThemes
      const found = allThemes?.find((t) => t.id === theme);
      if (found) {
        await applyCustomVSCodeTheme(found);
      } else {
        applyBuiltinTheme("dark");
      }
    } else {
      if (theme.is_builtin) {
        applyBuiltinTheme(theme.id);
      } else {
        await applyCustomVSCodeTheme(theme);
      }
    }
  };

  return {
    currentThemeId,
    applyTheme,
    applyBuiltinTheme,
    applyCustomVSCodeTheme,
  };
}
