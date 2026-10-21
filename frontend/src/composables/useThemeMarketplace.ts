// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { ref, reactive, computed } from "vue";
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
const currentPage = ref<number>(1);
const pageSize = ref<number>(24);
const isSearching = ref<boolean>(false);
const installingMap = reactive<Record<string, boolean>>({});
const marketplaceError = ref<string | null>(null);
const activeDetail = ref<MarketplaceExtensionDetail | null>(null);
const isLoadingDetail = ref<boolean>(false);

const totalPages = computed(() => {
  if (totalSize.value <= 0 || pageSize.value <= 0) return 1;
  return Math.ceil(totalSize.value / pageSize.value);
});

export function useThemeMarketplace() {
  const { fetchThemes } = useThemeManager();

  const searchThemes = async (q?: string, page: number = 1, size?: number) => {
    isSearching.value = true;
    marketplaceError.value = null;
    if (q !== undefined) {
      searchQuery.value = q;
    }
    if (size !== undefined && size > 0) {
      pageSize.value = size;
    }
    currentPage.value = Math.max(1, page);

    const offset = (currentPage.value - 1) * pageSize.value;

    try {
      const url = `/api/themes/marketplace/search?q=${encodeURIComponent(searchQuery.value)}&size=${pageSize.value}&offset=${offset}`;
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

  const goToPage = async (page: number) => {
    const target = Math.max(1, Math.min(page, totalPages.value));
    if (target === currentPage.value && searchResults.value.length > 0) return;
    await searchThemes(undefined, target);
  };

  const setPageSize = async (size: number) => {
    if (size === pageSize.value) return;
    await searchThemes(undefined, 1, size);
  };

  const isExtensionInstalling = (
    item?: { namespace?: string; name?: string; displayName?: string } | null
  ): boolean => {
    if (!item) return false;
    if (item.namespace && item.name && installingMap[`${item.namespace}.${item.name}`]) return true;
    if (item.name && installingMap[item.name]) return true;
    if (item.displayName && installingMap[item.displayName]) return true;
    return false;
  };

  const installThemeOnline = async (item: MarketplaceExtension): Promise<ThemeItem[]> => {
    if (!item.files?.download) {
      throw new Error("No download URL available for this package");
    }

    const keys: string[] = [];
    if (item.name) keys.push(item.name);
    if (item.displayName) keys.push(item.displayName);
    if (item.namespace && item.name) keys.push(`${item.namespace}.${item.name}`);

    for (const k of keys) {
      installingMap[k] = true;
    }
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
      for (const k of keys) {
        installingMap[k] = false;
      }
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
    currentPage,
    pageSize,
    totalPages,
    isSearching,
    installingMap,
    marketplaceError,
    activeDetail,
    isLoadingDetail,
    searchThemes,
    goToPage,
    setPageSize,
    isExtensionInstalling,
    installThemeOnline,
    fetchExtensionDetail,
    clearActiveDetail,
  };
}
