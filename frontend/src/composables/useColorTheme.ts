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

const currentThemeId = ref<string>("default");

export function useColorTheme() {
  const removeCustomThemeStyle = () => {
    const existing = document.getElementById(STYLE_TAG_ID);
    if (existing) {
      existing.remove();
    }
  };

  const applyBuiltinTheme = (themeId: string) => {
    removeCustomThemeStyle();
    currentThemeId.value = themeId;
  };

  const applyCustomVSCodeTheme = async (themeItem: ThemeItem) => {
    if (!themeItem.package_id || !themeItem.relative_path) {
      applyBuiltinTheme("default");
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
      const targetScope = isLight ? '[data-theme="light"]' : '[data-theme="dark"]';

      // Comprehensive mapping of VS Code workbench tokens to UniRTM layout and components
      const rootVars: string[] = [];
      const componentRules: string[] = [];

      const editorBg = colors["editor.background"];
      const sideBg = colors["sideBar.background"] || colors["editorGroupHeader.tabsBackground"] || colors["panel.background"];
      const titleBg = colors["titleBar.activeBackground"] || sideBg;
      const activityBg = colors["activityBar.background"] || sideBg;
      const statusBg = colors["statusBar.background"] || activityBg;
      const panelBg = colors["panel.background"] || sideBg;
      const breadcrumbBg = colors["editorGroupHeader.tabsBackground"] || sideBg;
      const inputBg = colors["input.background"] || colors["dropdown.background"];
      const textFg = colors["editor.foreground"] || colors["foreground"] || (isLight ? "#0f172a" : "#abb2bf");
      const textMuted = colors["descriptionForeground"] || colors["sideBarTitle.foreground"] || colors["tab.inactiveForeground"] || (isLight ? "#64748b" : "#5c6370");
      const accent = colors["button.background"] || colors["activityBarBadge.background"] || colors["focusBorder"] || "#00e5ff";
      const border = colors["sideBar.border"] || colors["panel.border"] || colors["editorGroup.border"] || (isLight ? "rgba(0,0,0,0.12)" : "rgba(255,255,255,0.08)");
      const borderActive = colors["focusBorder"] || colors["activityBar.activeBorder"] || accent;

      if (editorBg) rootVars.push(`--bg-color: ${editorBg} !important;`);
      if (sideBg) {
        rootVars.push(`--card-bg: ${sideBg} !important;`);
        rootVars.push(`--modal-bg: ${sideBg} !important;`);
      }
      if (inputBg) rootVars.push(`--input-bg: ${inputBg} !important;`);
      if (activityBg) rootVars.push(`--section-bg: ${activityBg} !important;`);
      if (textFg) rootVars.push(`--text-main: ${textFg} !important;`);
      if (textMuted) rootVars.push(`--text-muted: ${textMuted} !important;`);
      if (accent) {
        rootVars.push(`--accent-cyan: ${accent} !important;`);
        rootVars.push(`--accent-cyan-glow: ${accent}40 !important;`);
      }
      if (border) rootVars.push(`--card-border: ${border} !important;`);
      if (borderActive) rootVars.push(`--card-border-active: ${borderActive} !important;`);

      // Targeted layout overrides strictly scoped to this theme mode
      if (titleBg) {
        componentRules.push(`
          ${targetScope} .titlebar {
            background: ${titleBg} !important;
            border-bottom: 1px solid ${border} !important;
            color: ${colors["titleBar.activeForeground"] || textFg} !important;
          }
        `);
      }
      if (activityBg) {
        componentRules.push(`
          ${targetScope} .activitybar {
            background: ${activityBg} !important;
            border-right: 1px solid ${border} !important;
          }
        `);
      }
      if (sideBg) {
        componentRules.push(`
          ${targetScope} .sidebar-panel {
            background: ${sideBg} !important;
            border-right: 1px solid ${border} !important;
            color: ${colors["sideBar.foreground"] || textFg} !important;
          }
        `);
      }
      if (statusBg) {
        componentRules.push(`
          ${targetScope} .bottom-statusbar {
            background: ${statusBg} !important;
            border-top: 1px solid ${border} !important;
            color: ${colors["statusBar.foreground"] || textFg} !important;
          }
        `);
      }
      if (panelBg) {
        componentRules.push(`
          ${targetScope} .bottom-drawer {
            background: ${panelBg} !important;
            border-top: 1px solid ${border} !important;
          }
        `);
      }
      if (breadcrumbBg) {
        componentRules.push(`
          ${targetScope} .canvas-breadcrumb-bar {
            background: ${breadcrumbBg} !important;
            border-bottom: 1px solid ${border} !important;
          }
        `);
      }
      if (editorBg) {
        componentRules.push(`
          ${targetScope} .workbench-canvas {
            background: ${editorBg} !important;
          }
        `);
      }

      // Smooth visual cross-fade transition
      componentRules.push(`
        body, .vscode-app-layout, .workbench-canvas, .titlebar, .activitybar, .sidebar-panel, .bottom-statusbar, .glass-card, .bottom-drawer {
          transition: background-color 0.28s cubic-bezier(0.4, 0, 0.2, 1), border-color 0.28s ease, color 0.28s ease !important;
        }
      `);

      // Inject or update override style tag
      removeCustomThemeStyle();
      const styleEl = document.createElement("style");
      styleEl.id = STYLE_TAG_ID;
      styleEl.textContent = `${targetScope} { ${rootVars.join(" ")} }\n${componentRules.join("\n")}`;
      document.head.appendChild(styleEl);

      currentThemeId.value = themeItem.id;
    } catch (err) {
      console.error("[useColorTheme] Failed to apply custom VS Code color theme:", err);
      applyBuiltinTheme("default");
    }
  };

  const applyTheme = async (theme: string | ThemeItem, allThemes?: ThemeItem[]) => {
    if (typeof theme === "string") {
      if (theme === "default" || theme === "dark" || theme === "light" || theme === "system") {
        applyBuiltinTheme(theme);
        return;
      }
      const found = allThemes?.find((t) => t.id === theme);
      if (found) {
        await applyCustomVSCodeTheme(found);
      } else {
        applyBuiltinTheme("default");
      }
    } else {
      if (theme.is_builtin || theme.id === "default") {
        applyBuiltinTheme("default");
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
