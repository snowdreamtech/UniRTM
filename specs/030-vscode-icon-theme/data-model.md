# Phase 1 Data Model: VS Code Icon Theme 100% Specification Alignment

**Branch**: `030-vscode-icon-theme` | **Date**: 2026-10-15

## 1. Entities & Type Definitions

```typescript
// frontend/src/types/icon-theme.ts

/**
 * Single icon presentation definition
 */
export interface IconDefinition {
  fontCharacter?: string; // Unicode escape sequence, e.g. "\\E024"
  fontColor?: string;     // Hex color, e.g. "#00ADD8"
  fontSize?: string;      // e.g. "120%"
  iconPath?: string;      // Relative path to SVG or image file if image-based
}

/**
 * Font declaration matching VS Code icon theme font format
 */
export interface IconFontDefinition {
  id: string;
  src: Array<{
    path: string;
    format: "woff" | "woff2" | "ttf" | "opentype";
  }>;
  weight?: string;
  style?: string;
  size?: string;
}

/**
 * 100% VS Code Compliant Icon Theme Specification Root
 */
export interface IconThemeManifest {
  fonts?: IconFontDefinition[];
  iconDefinitions: Record<string, IconDefinition>;
  file?: string;
  folder?: string;
  folderExpanded?: string;
  rootFolder?: string;
  rootFolderExpanded?: string;
  fileExtensions?: Record<string, string>;
  fileNames?: Record<string, string>;
  languageIds?: Record<string, string>;
  light?: Partial<IconThemeManifest>;
  highContrast?: Partial<IconThemeManifest>;
}

/**
 * Resolved icon outcome for component rendering
 */
export interface ResolvedIcon {
  type: "font" | "svg";
  fontFamily?: string;
  character?: string;
  color?: string;
  svgPath?: string;
  definitionKey: string;
}
```

## 2. Component Props Interface

```typescript
// frontend/src/components/common/TechIcon.vue

export interface TechIconProps {
  /**
   * Tool name, language ID, file name, or file extension (e.g., 'go', 'node', 'unirtm.lock', 'Cargo.toml')
   */
  id: string;

  /**
   * Icon size in pixels. Defaults to 16.
   */
  size?: number;

  /**
   * Whether to apply the theme's brand fontColor. Defaults to true.
   */
  colored?: boolean;

  /**
   * Optional custom CSS class
   */
  className?: string;

  /**
   * Fallback icon definition key if not found
   */
  fallbackKey?: string;
}
```

## 3. Resolution State Machine

```
Target ID (e.g. "go", "Cargo.toml", "unirtm.lock")
                      │
                      ▼
             Normalize: toLowerCase()
                      │
     ┌────────────────┼────────────────┐
     ▼                ▼                ▼
fileNames[id]   fileExtensions[ext]  languageIds[id]
     │                │                │
     └────────────────┼────────────────┘
                      │
               Match Found?
               ├── Yes ──► lookup iconDefinitions[matchedKey]
               └── No  ──► lookup iconDefinitions[manifest.file || "_file"]
                      │
                      ▼
               Return ResolvedIcon
```
