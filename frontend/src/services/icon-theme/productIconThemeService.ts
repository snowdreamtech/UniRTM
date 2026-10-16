// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import type { ThemeItem } from "../../types/theme-manager";

function parseJSONC(text: string): Record<string, any> {
  const cleaned = text
    .replace(/\/\*[\s\S]*?\*\//g, "")
    .replace(/\/\/[^\n\r]*/g, "")
    .replace(/,\s*([\]}])/g, "$1");
  return JSON.parse(cleaned);
}

function resolveRelativePath(baseFile: string, relPath: string): string {
  const cleanBase = baseFile.replace(/^\.\//, "");
  const dirParts = cleanBase.split("/").slice(0, -1);
  const segments = relPath.replace(/^\.\//, "").split("/");
  for (const seg of segments) {
    if (seg === "..") {
      dirParts.pop();
    } else if (seg && seg !== ".") {
      dirParts.push(seg);
    }
  }
  return dirParts.join("/");
}

export class ProductIconThemeService {
  private activeThemeId = "codicons";
  private styleElementId = "unirtm-product-icon-theme-styles";

  public getActiveThemeId(): string {
    return this.activeThemeId;
  }

  public resetToDefault(): void {
    this.activeThemeId = "codicons";
    if (typeof document === "undefined") return;
    const styleEl = document.getElementById(this.styleElementId);
    if (styleEl) {
      styleEl.textContent = "";
    }
  }

  public async loadAndApplyTheme(theme: ThemeItem): Promise<void> {
    if (!theme || theme.id === "codicons" || theme.is_builtin) {
      this.resetToDefault();
      return;
    }

    if (!theme.package_id || !theme.relative_path) {
      console.warn("[ProductIconThemeService] Missing package_id or relative_path", theme);
      return;
    }

    try {
      const cleanPath = theme.relative_path.replace(/^\.\//, "");
      const manifestUrl = `/api/themes/assets/product/${encodeURIComponent(theme.package_id)}/${cleanPath}`;
      const res = await fetch(manifestUrl);
      if (!res.ok) {
        throw new Error(`Failed to load product icon manifest: ${res.statusText}`);
      }

      const text = await res.text();
      const manifest = parseJSONC(text);

      let css = "";
      const fontFamilies: string[] = [];

      if (Array.isArray(manifest.fonts)) {
        for (const font of manifest.fonts) {
          const fontId = font.id || theme.id;
          fontFamilies.push(`'${fontId}'`);

          const srcParts: string[] = [];
          if (Array.isArray(font.src)) {
            for (const s of font.src) {
              const resolvedFontPath = resolveRelativePath(theme.relative_path, s.path);
              const fontUrl = `/api/themes/assets/product/${theme.package_id}/${resolvedFontPath}`;
              const format = s.format ? ` format('${s.format}')` : "";
              srcParts.push(`url('${fontUrl}')${format}`);
            }
          }

          if (srcParts.length > 0) {
            css += `
@font-face {
  font-family: '${fontId}';
  src: ${srcParts.join(", ")};
  font-weight: ${font.weight || "normal"};
  font-style: ${font.style || "normal"};
  font-display: block;
}
`;
          }
        }
      }

      const primaryFont = fontFamilies.length > 0 ? fontFamilies.join(", ") + ", " : "";
      css += `
.codicon {
  font-family: ${primaryFont}'codicon' !important;
}
`;

      if (manifest.iconDefinitions && typeof manifest.iconDefinitions === "object") {
        for (const [iconName, def] of Object.entries<any>(manifest.iconDefinitions)) {
          if (def && def.fontCharacter) {
            let charCode = def.fontCharacter;
            if (charCode.startsWith("\\")) {
              charCode = charCode.slice(1);
            }
            css += `
.codicon-${iconName}:before {
  content: "\\${charCode}" !important;
}
`;
          }
        }
      }

      if (typeof document !== "undefined") {
        let styleEl = document.getElementById(this.styleElementId) as HTMLStyleElement | null;
        if (!styleEl) {
          styleEl = document.createElement("style");
          styleEl.id = this.styleElementId;
          document.head.appendChild(styleEl);
        }
        styleEl.textContent = css;
      }

      this.activeThemeId = theme.id;
    } catch (err) {
      console.error("[ProductIconThemeService] Failed to apply product icon theme:", err);
      this.resetToDefault();
      throw err;
    }
  }
}

export const productIconThemeService = new ProductIconThemeService();
