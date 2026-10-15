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
