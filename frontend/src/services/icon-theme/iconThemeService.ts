// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import type { IconThemeManifest, ResolvedIcon } from "../../types/icon-theme";
import defaultSetiManifest from "../../assets/themes/seti/seti.json";
import setiWoff2Url from "../../assets/themes/seti/seti.woff2";
import setiWoffUrl from "../../assets/themes/seti/seti.woff";

class IconThemeService {
  private activeManifest: IconThemeManifest;
  private isStyleInjected = false;

  constructor() {
    this.activeManifest = defaultSetiManifest as unknown as IconThemeManifest;
  }

  /**
   * Inject the @font-face and base icon styles dynamically into the DOM head.
   */
  public ensureStylesInjected(): void {
    if (this.isStyleInjected || typeof document === "undefined") {
      return;
    }

    const styleId = "unirtm-icon-theme-styles";
    let styleEl = document.getElementById(styleId) as HTMLStyleElement | null;
    if (!styleEl) {
      styleEl = document.createElement("style");
      styleEl.id = styleId;
      document.head.appendChild(styleEl);
    }

    styleEl.textContent = `
@font-face {
  font-family: 'seti';
  src: url('${setiWoff2Url}') format('woff2'),
       url('${setiWoffUrl}') format('woff');
  font-weight: normal;
  font-style: normal;
  font-display: block;
}

.tech-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-family: 'seti', -apple-system, BlinkMacSystemFont, sans-serif;
  font-style: normal;
  font-variant: normal;
  text-rendering: auto;
  line-height: 1;
  vertical-align: middle;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}
`;
    this.isStyleInjected = true;
  }

  private customThemes: Map<string, IconThemeManifest> = new Map();

  /**
   * Set or update active theme manifest with validation and safe fallback.
   */
  public setManifest(manifest: IconThemeManifest): boolean {
    if (!manifest || typeof manifest !== "object" || !manifest.iconDefinitions) {
      console.warn("[IconThemeService] Rejected invalid manifest. Keeping fallback.");
      return false;
    }
    this.activeManifest = manifest;
    return true;
  }

  /**
   * Register a custom theme manifest.
   */
  public registerCustomTheme(themeId: string, manifest: IconThemeManifest): boolean {
    if (!manifest || typeof manifest !== "object" || !manifest.iconDefinitions) {
      console.warn(`[IconThemeService] Invalid theme manifest rejected for "${themeId}".`);
      return false;
    }
    this.customThemes.set(themeId, manifest);
    return true;
  }

  /**
   * Switch to a registered custom theme.
   */
  public switchTheme(themeId: string): boolean {
    const target = this.customThemes.get(themeId);
    if (target) {
      this.activeManifest = target;
      return true;
    }
    return false;
  }

  /**
   * Get current active manifest.
   */
  public getManifest(): IconThemeManifest {
    return this.activeManifest;
  }

  /**
   * Parse hex escape code (e.g. "\\E024", "\uE024", "E024") into actual character.
   */
  public parseCharacter(rawChar?: string): string {
    if (!rawChar) return "";
    try {
      const cleanHex = rawChar.replace(/^\\+[EeUu]?/, "").replace(/^0x/i, "");
      const codePoint = parseInt(cleanHex, 16);
      if (!isNaN(codePoint)) {
        return String.fromCodePoint(codePoint);
      }
    } catch {
      // ignore
    }
    return rawChar;
  }

  /**
   * Resolve icon definition key using VS Code standard priority:
   * 1. fileNames
   * 2. fileExtensions
   * 3. languageIds
   * 4. file (default)
   */
  public resolveIcon(rawIdentifier: string): ResolvedIcon {
    if (!rawIdentifier) {
      return this.getDefaultIcon();
    }

    const trimmed = rawIdentifier.trim();
    const lower = trimmed.toLowerCase();

    // Clean compound prefixes/paths like "go:mvdan.cc/sh/v3/cmd/shfmt" -> "shfmt"
    let toolName = lower;
    if (toolName.includes(":")) {
      toolName = toolName.split(":").pop() || toolName;
    }
    if (toolName.includes("/")) {
      toolName = toolName.split("/").pop() || toolName;
    }

    const manifest = this.activeManifest;
    let defKey: string | undefined;

    // 1. Check exact fileNames (e.g. 'unirtm.lock', 'Cargo.toml')
    if (manifest.fileNames) {
      defKey = manifest.fileNames[trimmed] || manifest.fileNames[lower] || manifest.fileNames[toolName];
    }

    // 2. Check fileExtensions if identifier has an extension (e.g. 'app.go' -> 'go')
    if (!defKey && manifest.fileExtensions) {
      if (trimmed.includes(".")) {
        const parts = trimmed.split(".");
        const ext = parts[parts.length - 1].toLowerCase();
        defKey = manifest.fileExtensions[ext];
      } else {
        // Also check if tool itself matches an extension (e.g. 'py', 'rs', 'go')
        defKey = manifest.fileExtensions[toolName];
      }
    }

    // 3. Check languageIds (e.g. 'go', 'golang', 'rust', 'python', 'shfmt')
    if (!defKey && manifest.languageIds) {
      defKey = manifest.languageIds[toolName] || manifest.languageIds[lower];
    }

    if (defKey && manifest.iconDefinitions && manifest.iconDefinitions[defKey]) {
      const def = manifest.iconDefinitions[defKey];
      return {
        type: "font",
        fontFamily: "seti",
        character: this.parseCharacter(def.fontCharacter),
        color: def.fontColor,
        fontSize: def.fontSize,
        definitionKey: defKey,
      };
    }

    return this.getDefaultIcon();
  }

  private getDefaultIcon(): ResolvedIcon {
    const defaultKey = this.activeManifest.file || "_default";
    const def = this.activeManifest.iconDefinitions?.[defaultKey];
    return {
      type: "font",
      fontFamily: "seti",
      character: this.parseCharacter(def?.fontCharacter || "\\E023"),
      color: def?.fontColor || "#d4d7d6",
      definitionKey: defaultKey,
    };
  }
}

export const iconThemeService = new IconThemeService();
