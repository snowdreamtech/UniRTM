// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

/**
 * 100% VS Code Compatible File Icon Theme Types
 * Aligned with https://code.visualstudio.com/api/extension-guides/file-icon-theme
 */

export interface IconDefinition {
  fontCharacter?: string;
  fontColor?: string;
  fontSize?: string;
  iconPath?: string;
}

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

export interface ResolvedIcon {
  type: "font" | "svg";
  fontFamily?: string;
  character?: string;
  color?: string;
  fontSize?: string;
  svgPath?: string;
  definitionKey: string;
  isFallback?: boolean;
}

export type WorkbenchIconKey =
  | "workbench:dashboard"
  | "workbench:runtimes"
  | "workbench:tools"
  | "workbench:tasks"
  | "workbench:marketplace"
  | "workbench:doctor"
  | "workbench:settings"
  | "workbench:about"
  | "workbench:terminal"
  | "workbench:logs"
  | "workbench:lock"
  | "workbench:config"
  | "workbench:success"
  | "workbench:warning"
  | "workbench:error"
  | "workbench:info"
  | "workbench:sync"
  | "workbench:search"
  | "workbench:delete"
  | "workbench:folder"
  | "workbench:package"
  | "workbench:globe"
  | "workbench:speed"
  | "workbench:star"
  | "workbench:user"
  | "workbench:edit"
  | "workbench:github"
  | "workbench:sparkle";

export interface SemanticIconMapping {
  candidates: string[];
  fallbackSetiKey: string;
}
