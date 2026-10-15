// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

export interface ThemeItem {
  id: string;
  name: string;
  type: "color" | "icon";
  is_builtin: boolean;
  ui_theme?: string;
  package_id?: string;
  relative_path?: string;
  description?: string;
  version?: string;
}

export interface ThemesListResponse {
  colors: ThemeItem[];
  icons: ThemeItem[];
  active: {
    color: string;
    icon: string;
  };
}

export interface MarketplaceExtension {
  namespace: string;
  name: string;
  displayName: string;
  version: string;
  description: string;
  downloadCount: number;
  averageRating: number;
  files: {
    download: string;
    icon?: string;
  };
  url?: string;
  is_installed?: boolean;
  package_id?: string;
}

export interface MarketplaceSearchResponse {
  offset: number;
  totalSize: number;
  extensions: MarketplaceExtension[];
}

export interface ContributedSubTheme {
  id?: string;
  label: string;
  uiTheme?: string;
  path?: string;
}

export interface ContributedSubIconTheme {
  id?: string;
  label: string;
  path?: string;
}

export interface ExtensionContributions {
  themes?: ContributedSubTheme[];
  iconThemes?: ContributedSubIconTheme[];
}

export interface MarketplaceExtensionDetail {
  namespace: string;
  name: string;
  displayName: string;
  version: string;
  description: string;
  downloadCount: number;
  averageRating: number;
  reviewCount: number;
  repository?: string;
  homepage?: string;
  bugs?: string;
  license?: string;
  timestamp?: string;
  tags?: string[];
  categories?: string[];
  files: {
    download: string;
    icon?: string;
    readme?: string;
    changelog?: string;
    manifest?: string;
  };
  url?: string;
  is_installed?: boolean;
  package_id?: string;
  readme?: string;
  contributes?: ExtensionContributions;
}

