// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { ref, reactive } from "vue";
import type {
  MarketplaceExtension,
  MarketplaceSearchResponse,
  MarketplaceExtensionDetail,
  ThemeItem,
} from "../types/theme-manager";
import { useThemeManager } from "./useThemeManager";

const searchQuery = ref<string>("");
const searchResults = ref<MarketplaceExtension[]>([]);
const totalSize = ref<number>(0);
const isSearching = ref<boolean>(false);
const installingMap = reactive<Record<string, boolean>>({});
const marketplaceError = ref<string | null>(null);
const activeDetail = ref<MarketplaceExtensionDetail | null>(null);
const isLoadingDetail = ref<boolean>(false);

export function useThemeMarketplace() {
  const { fetchThemes } = useThemeManager();

  const searchThemes = async (q?: string) => {
    isSearching.value = true;
    marketplaceError.value = null;
    if (q !== undefined) {
      searchQuery.value = q;
    }

    try {
      const url = `/api/themes/marketplace/search?q=${encodeURIComponent(searchQuery.value)}&size=30`;
      const res = await fetch(url);
      if (!res.ok) {
        throw new Error(`Failed to search Open VSX: ${res.statusText}`);
      }
      const data: MarketplaceSearchResponse = await res.json();
      searchResults.value = data.extensions || [];
      totalSize.value = data.totalSize || 0;
    } catch (err: any) {
      marketplaceError.value = err.message || "Failed to query themes";
      console.error("[useThemeMarketplace] search error:", err);
    } finally {
      isSearching.value = false;
    }
  };

  const installThemeOnline = async (item: MarketplaceExtension): Promise<ThemeItem[]> => {
    if (!item.files?.download) {
      throw new Error("No download URL available for this package");
    }

    const key = item.name || item.displayName;
    installingMap[key] = true;
    marketplaceError.value = null;

    try {
      const res = await fetch("/api/themes/marketplace/install", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          download_url: item.files.download,
        }),
      });

      const data = await res.json();
      if (!res.ok || data.error) {
        throw new Error(data.error || "Online installation failed");
      }

      item.is_installed = true;
      // Refresh local theme manager state so new themes appear in dropdown
      await fetchThemes();
      return (data.installed as ThemeItem[]) || [];
    } catch (err: any) {
      marketplaceError.value = err.message || "Failed to install theme";
      throw err;
    } finally {
      installingMap[key] = false;
    }
  };

  const fetchExtensionDetail = async (namespace: string, name: string): Promise<MarketplaceExtensionDetail | null> => {
    isLoadingDetail.value = true;
    marketplaceError.value = null;
    try {
      const url = `/api/themes/marketplace/detail?namespace=${encodeURIComponent(namespace)}&name=${encodeURIComponent(name)}`;
      const res = await fetch(url);
      if (!res.ok) {
        throw new Error(`Failed to fetch extension details: ${res.statusText}`);
      }
      const data: MarketplaceExtensionDetail = await res.json();
      activeDetail.value = data;
      return data;
    } catch (err: any) {
      marketplaceError.value = err.message || "Failed to fetch extension detail";
      console.error("[useThemeMarketplace] detail error:", err);
      return null;
    } finally {
      isLoadingDetail.value = false;
    }
  };

  const clearActiveDetail = () => {
    activeDetail.value = null;
  };

  return {
    searchQuery,
    searchResults,
    totalSize,
    isSearching,
    installingMap,
    marketplaceError,
    activeDetail,
    isLoadingDetail,
    searchThemes,
    installThemeOnline,
    fetchExtensionDetail,
    clearActiveDetail,
  };
}

