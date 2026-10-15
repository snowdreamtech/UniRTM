// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import type { IconThemeManifest, ResolvedIcon } from "../../types/icon-theme";
import type { ThemeItem } from "../../types/theme-manager";
import defaultSetiManifest from "../../assets/themes/seti/seti.json";
import setiWoff2Url from "../../assets/themes/seti/seti.woff2";
import setiWoffUrl from "../../assets/themes/seti/seti.woff";

function parseJSONC(text: string): Record<string, any> {
  const cleaned = text
    .replace(/\/\*[\s\S]*?\*\//g, "")
    .replace(/\/\/[^\n\r]*/g, "")
    .replace(/,\s*([\]}])/g, "$1");
  return JSON.parse(cleaned);
}

class IconThemeService {
  private activeManifest: IconThemeManifest;
  private isStyleInjected = false;
  private currentBaseUrl = "";
  private activeThemeId = "seti";
  private listeners: Array<() => void> = [];

  constructor() {
    this.activeManifest = defaultSetiManifest as unknown as IconThemeManifest;
  }

  public subscribe(cb: () => void): () => void {
    this.listeners.push(cb);
    return () => {
      this.listeners = this.listeners.filter((l) => l !== cb);
    };
  }

  private notify(): void {
    for (const listener of this.listeners) {
      try {
        listener();
      } catch {
        // ignore
      }
    }
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
    this.notify();
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
    if (themeId === "seti") {
      this.activeManifest = defaultSetiManifest as unknown as IconThemeManifest;
      this.currentBaseUrl = "";
      this.activeThemeId = "seti";
      this.notify();
      return true;
    }
    const target = this.customThemes.get(themeId);
    if (target) {
      this.activeManifest = target;
      this.activeThemeId = themeId;
      this.notify();
      return true;
    }
    return false;
  }

  /**
   * Get current active theme ID.
   */
  public getActiveThemeId(): string {
    return this.activeThemeId;
  }

  /**
   * Load and apply icon theme from server API or builtin
   */
  public async loadAndApplyTheme(themeItem: ThemeItem): Promise<boolean> {
    if (themeItem.id === "seti" || themeItem.is_builtin) {
      this.activeManifest = defaultSetiManifest as unknown as IconThemeManifest;
      this.currentBaseUrl = "";
      this.activeThemeId = "seti";
      this.notify();
      return true;
    }

    if (!themeItem.package_id || !themeItem.relative_path) {
      return false;
    }

    const cleanRel = themeItem.relative_path.startsWith("./")
      ? themeItem.relative_path.slice(2)
      : themeItem.relative_path.startsWith("/")
      ? themeItem.relative_path.slice(1)
      : themeItem.relative_path;

    const manifestUrl = `/api/themes/assets/icon/${encodeURIComponent(themeItem.package_id)}/${cleanRel}`;
    const lastSlash = cleanRel.lastIndexOf("/");
    const manifestDir = lastSlash !== -1 ? cleanRel.substring(0, lastSlash) : "";
    this.currentBaseUrl = `/api/themes/assets/icon/${encodeURIComponent(themeItem.package_id)}/${manifestDir ? manifestDir + "/" : ""}`;

    try {
      const res = await fetch(manifestUrl);
      if (!res.ok) {
        throw new Error(`Failed to load icon theme: ${res.statusText}`);
      }
      const text = await res.text();
      const manifest = parseJSONC(text) as IconThemeManifest;

      if (manifest.fonts && manifest.fonts.length > 0) {
        let fontFaceCSS = "";
        for (const font of manifest.fonts) {
          const srcRules = font.src
            .map((s) => {
              const fontRel = s.path.startsWith("./") ? s.path.slice(2) : s.path;
              const fullFontUrl = this.currentBaseUrl + fontRel;
              return `url('${fullFontUrl}') format('${s.format}')`;
            })
            .join(",\n       ");

          fontFaceCSS += `
@font-face {
  font-family: '${font.id}';
  src: ${srcRules};
  font-weight: ${font.weight || "normal"};
  font-style: ${font.style || "normal"};
  font-display: block;
}
`;
        }

        const customStyleId = "unirtm-custom-icon-fonts";
        let customStyleEl = document.getElementById(customStyleId) as HTMLStyleElement | null;
        if (!customStyleEl) {
          customStyleEl = document.createElement("style");
          customStyleEl.id = customStyleId;
          document.head.appendChild(customStyleEl);
        }
        customStyleEl.textContent = fontFaceCSS;
      }

      this.activeManifest = manifest;
      this.activeThemeId = themeItem.id;
      this.notify();
      return true;
    } catch (err) {
      console.error("[IconThemeService] Failed to load custom icon theme:", err);
      return false;
    }
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
      const cleanHex = rawChar.replace(/^\\+u?/i, "").replace(/^0x/i, "");
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

    let toolName = lower;
    if (toolName.includes(":")) {
      toolName = toolName.split(":").pop() || toolName;
    }
    if (toolName.includes("/")) {
      toolName = toolName.split("/").pop() || toolName;
    }

    const manifest = this.activeManifest;
    let defKey: string | undefined;

    // 1. Check exact fileNames
    if (manifest.fileNames) {
      defKey = manifest.fileNames[trimmed] || manifest.fileNames[lower] || manifest.fileNames[toolName];
    }

    // 2. Check fileExtensions
    if (!defKey && manifest.fileExtensions) {
      if (trimmed.includes(".")) {
        const parts = trimmed.split(".");
        const ext = parts[parts.length - 1].toLowerCase();
        defKey = manifest.fileExtensions[ext];
      } else {
        defKey = manifest.fileExtensions[toolName];
      }
    }

    // 3. Check languageIds
    if (!defKey && manifest.languageIds) {
      defKey = manifest.languageIds[toolName] || manifest.languageIds[lower];
    }

    if (defKey && manifest.iconDefinitions && manifest.iconDefinitions[defKey]) {
      const def = manifest.iconDefinitions[defKey];
      if (def.iconPath) {
        const cleanPath = def.iconPath.startsWith("./") ? def.iconPath.slice(2) : def.iconPath;
        const fullSvgPath = this.currentBaseUrl ? this.currentBaseUrl + cleanPath : def.iconPath;
        return {
          type: "svg",
          svgPath: fullSvgPath,
          definitionKey: defKey,
        };
      }
      return {
        type: "font",
        fontFamily: (manifest.fonts && manifest.fonts[0]?.id) || "seti",
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
    if (def?.iconPath) {
      const cleanPath = def.iconPath.startsWith("./") ? def.iconPath.slice(2) : def.iconPath;
      const fullSvgPath = this.currentBaseUrl ? this.currentBaseUrl + cleanPath : def.iconPath;
      return {
        type: "svg",
        svgPath: fullSvgPath,
        definitionKey: defaultKey,
      };
    }
    return {
      type: "font",
      fontFamily: (this.activeManifest.fonts && this.activeManifest.fonts[0]?.id) || "seti",
      character: this.parseCharacter(def?.fontCharacter || "\\E023"),
      color: def?.fontColor || "#d4d7d6",
      definitionKey: defaultKey,
    };
  }
}

export const iconThemeService = new IconThemeService();
