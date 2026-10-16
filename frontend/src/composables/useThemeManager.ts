// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { ref, watch } from "vue";
import type { ThemeItem } from "../types/theme-manager";
import { useColorTheme } from "./useColorTheme";
import { useTheme } from "./useTheme";
import { iconThemeService } from "../services/icon-theme/iconThemeService";
import { productIconThemeService } from "../services/icon-theme/productIconThemeService";

const PREF_DARK_KEY = "unirtm_pref_dark_theme";
const PREF_LIGHT_KEY = "unirtm_pref_light_theme";

const defaultColors: ThemeItem[] = [
  { id: "default", name: "UniRTM (默认)", type: "color", is_builtin: true, ui_theme: "vs-dark" },
  { id: "dark", name: "UniRTM Dark", type: "color", is_builtin: true, ui_theme: "vs-dark" },
  { id: "light", name: "UniRTM Light", type: "color", is_builtin: true, ui_theme: "vs" },
];
const defaultIcons: ThemeItem[] = [
  { id: "seti", name: "Seti UI", type: "icon", is_builtin: true, description: "Official Visual Studio Code default file icon theme" },
];
const defaultProducts: ThemeItem[] = [
  { id: "codicons", name: "Codicons (默认)", type: "product", is_builtin: true, description: "Official Visual Studio Code product icon theme" },
];

const colorThemes = ref<ThemeItem[]>(defaultColors);
const iconThemes = ref<ThemeItem[]>(defaultIcons);
const productIconThemes = ref<ThemeItem[]>(defaultProducts);

const activeColorTheme = ref<string>("default");
const activeIconTheme = ref<string>("seti");
const activeProductIconTheme = ref<string>("codicons");
const activeAppearance = ref<string>("system");

const preferredDarkTheme = ref<string>(
  (typeof localStorage !== "undefined" && localStorage.getItem(PREF_DARK_KEY)) || "default"
);
const preferredLightTheme = ref<string>(
  (typeof localStorage !== "undefined" && localStorage.getItem(PREF_LIGHT_KEY)) || "default"
);

const loading = ref<boolean>(false);
const error = ref<string | null>(null);

const { applyTheme: applyColorTheme } = useColorTheme();
const { applyTheme: applyAppearanceTheme, effectiveTheme } = useTheme();

// Automatically update theme when effective appearance changes (e.g. OS scheme changes in system mode)
watch(effectiveTheme, async (newEffective) => {
  const targetTheme = newEffective === "light" ? preferredLightTheme.value : preferredDarkTheme.value;
  activeColorTheme.value = targetTheme;
  await applyColorTheme(targetTheme, colorThemes.value);
});

export function useThemeManager() {
  const fetchThemes = async () => {
    loading.value = true;
    error.value = null;
    try {
      const res = await fetch("/api/themes");
      if (!res.ok) {
        throw new Error(`Failed to fetch themes: ${res.statusText}`);
      }
      const data: any = await res.json();
      colorThemes.value = data.colors || data.color_themes || [];
      iconThemes.value = data.icons || data.icon_themes || [];
      productIconThemes.value = data.products || data.product_icon_themes || defaultProducts;

      const serverColor = data.active?.color || data.active_color_theme || "default";
      activeIconTheme.value = data.active?.icon || data.active_icon_theme || "seti";
      activeProductIconTheme.value = data.active?.product || data.active_product_icon_theme || "codicons";
      activeAppearance.value = data.active?.appearance || "system";

      // Apply appearance mode first
      applyAppearanceTheme(activeAppearance.value);

      // Initialize preferred themes from server or localStorage
      const serverThemeItem = colorThemes.value.find((c) => c.id === serverColor);
      const isServerLight = serverColor === "light" || (serverThemeItem && (serverThemeItem.ui_theme === "vs" || serverThemeItem.ui_theme === "light"));

      if (!localStorage.getItem(PREF_DARK_KEY)) {
        if (!isServerLight && serverColor !== "default") {
          preferredDarkTheme.value = serverColor;
          localStorage.setItem(PREF_DARK_KEY, serverColor);
        } else {
          preferredDarkTheme.value = "default";
        }
      } else {
        preferredDarkTheme.value = localStorage.getItem(PREF_DARK_KEY)!;
      }

      if (!localStorage.getItem(PREF_LIGHT_KEY)) {
        if (isServerLight) {
          preferredLightTheme.value = serverColor;
          localStorage.setItem(PREF_LIGHT_KEY, serverColor);
        } else {
          preferredLightTheme.value = "default";
        }
      } else {
        preferredLightTheme.value = localStorage.getItem(PREF_LIGHT_KEY)!;
      }

      // Determine active theme based on effective appearance mode
      const targetTheme = effectiveTheme.value === "light" ? preferredLightTheme.value : preferredDarkTheme.value;
      activeColorTheme.value = targetTheme;

      // Apply active themes hot
      await applyColorTheme(targetTheme, colorThemes.value);

      const activeIconItem = iconThemes.value.find((i) => i.id === activeIconTheme.value);
      if (activeIconItem) {
        await iconThemeService.loadAndApplyTheme(activeIconItem);
      } else {
        await iconThemeService.switchTheme("seti");
      }

      const activeProductItem = productIconThemes.value.find((p) => p.id === activeProductIconTheme.value);
      if (activeProductItem) {
        await productIconThemeService.loadAndApplyTheme(activeProductItem);
      } else {
        productIconThemeService.resetToDefault();
      }
    } catch (err: any) {
      error.value = err.message || "Failed to load themes";
      console.error("[useThemeManager] fetchThemes error:", err);
    } finally {
      loading.value = false;
    }
  };

  const installTheme = async (file: File) => {
    loading.value = true;
    error.value = null;
    try {
      const formData = new FormData();
      formData.append("file", file);

      const res = await fetch("/api/themes/install", {
        method: "POST",
        body: formData,
      });

      const data = await res.json();
      if (!res.ok || data.error) {
        throw new Error(data.error || "Installation failed");
      }

      await fetchThemes();
      return data.installed as ThemeItem[];
    } catch (err: any) {
      error.value = err.message || "Failed to install theme";
      throw err;
    } finally {
      loading.value = false;
    }
  };

  const deleteTheme = async (type: "color" | "icon" | "product", packageId: string) => {
    loading.value = true;
    error.value = null;
    try {
      const res = await fetch(`/api/themes?type=${encodeURIComponent(type)}&id=${encodeURIComponent(packageId)}`, {
        method: "DELETE",
      });

      const data = await res.json();
      if (!res.ok || data.error) {
        throw new Error(data.error || "Delete failed");
      }

      // If active theme was deleted, reset to default
      if (type === "color") {
        const deletedColor = colorThemes.value.find((c) => c.package_id === packageId && c.id === activeColorTheme.value);
        if (deletedColor) {
          if (preferredDarkTheme.value === deletedColor.id) {
            preferredDarkTheme.value = "default";
            localStorage.setItem(PREF_DARK_KEY, "default");
          }
          if (preferredLightTheme.value === deletedColor.id) {
            preferredLightTheme.value = "default";
            localStorage.setItem(PREF_LIGHT_KEY, "default");
          }
          await setActiveTheme("default", undefined);
        }
      } else if (type === "icon") {
        const deletedIcon = iconThemes.value.find((i) => i.package_id === packageId && i.id === activeIconTheme.value);
        if (deletedIcon) {
          await setActiveTheme(undefined, "seti");
        }
      } else if (type === "product") {
        const deletedProduct = productIconThemes.value.find((p) => p.package_id === packageId && p.id === activeProductIconTheme.value);
        if (deletedProduct) {
          await setActiveTheme(undefined, undefined, undefined, "codicons");
        }
      }

      await fetchThemes();
    } catch (err: any) {
      error.value = err.message || "Failed to delete theme";
      throw err;
    } finally {
      loading.value = false;
    }
  };

  const setActiveTheme = async (colorId?: string, iconId?: string, appearance?: string, productId?: string) => {
    loading.value = true;
    error.value = null;
    try {
      const payload: Record<string, string> = {};

      if (appearance) {
        activeAppearance.value = appearance;
        applyAppearanceTheme(appearance);
        payload.appearance = appearance;

        // When appearance changes and no explicit colorId is passed, apply preferred theme for this mode
        if (!colorId) {
          const modeTheme = effectiveTheme.value === "light" ? preferredLightTheme.value : preferredDarkTheme.value;
          activeColorTheme.value = modeTheme;
          await applyColorTheme(modeTheme, colorThemes.value);
          payload.color = modeTheme;
        }
      }

      if (colorId) {
        payload.color = colorId;
        const themeItem = colorThemes.value.find((c) => c.id === colorId);
        const isLight = colorId === "light" || (themeItem && (themeItem.ui_theme === "vs" || themeItem.ui_theme === "light"));
        const isDark = colorId === "dark" || (themeItem && (themeItem.ui_theme === "vs-dark" || themeItem.ui_theme === "dark"));

        if (colorId === "default") {
          if (effectiveTheme.value === "light") {
            preferredLightTheme.value = "default";
            localStorage.setItem(PREF_LIGHT_KEY, "default");
          } else {
            preferredDarkTheme.value = "default";
            localStorage.setItem(PREF_DARK_KEY, "default");
          }
        } else if (isLight) {
          preferredLightTheme.value = colorId;
          localStorage.setItem(PREF_LIGHT_KEY, colorId);
          // If user picked a light theme while currently in dark mode, auto-switch to light appearance
          if (activeAppearance.value !== "light") {
            activeAppearance.value = "light";
            applyAppearanceTheme("light");
            payload.appearance = "light";
          }
        } else if (isDark) {
          preferredDarkTheme.value = colorId;
          localStorage.setItem(PREF_DARK_KEY, colorId);
          // If user picked a dark theme while currently in light mode, auto-switch to dark appearance
          if (activeAppearance.value !== "dark") {
            activeAppearance.value = "dark";
            applyAppearanceTheme("dark");
            payload.appearance = "dark";
          }
        }

        activeColorTheme.value = colorId;
        await applyColorTheme(colorId, colorThemes.value);
      }

      if (iconId) {
        payload.icon = iconId;
        activeIconTheme.value = iconId;
        const iconItem = iconThemes.value.find((i) => i.id === iconId);
        if (iconItem) {
          await iconThemeService.loadAndApplyTheme(iconItem);
        } else {
          await iconThemeService.switchTheme("seti");
        }
      }

      if (productId) {
        payload.product = productId;
        activeProductIconTheme.value = productId;
        const prodItem = productIconThemes.value.find((p) => p.id === productId);
        if (prodItem) {
          await productIconThemeService.loadAndApplyTheme(prodItem);
        } else {
          productIconThemeService.resetToDefault();
        }
      }

      const res = await fetch("/api/themes/apply", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });

      const data = await res.json();
      if (!res.ok || data.error) {
        throw new Error(data.error || "Failed to apply theme");
      }
    } catch (err: any) {
      error.value = err.message || "Failed to apply theme";
      throw err;
    } finally {
      loading.value = false;
    }
  };

  const toggleAppearance = async (): Promise<string> => {
    const nextMode = activeAppearance.value === "light" ? "dark" : "light";
    await setActiveTheme(undefined, undefined, nextMode);
    return nextMode;
  };

  return {
    colorThemes,
    iconThemes,
    productIconThemes,
    activeColorTheme,
    activeIconTheme,
    activeProductIconTheme,
    activeAppearance,
    preferredDarkTheme,
    preferredLightTheme,
    loading,
    error,
    fetchThemes,
    installTheme,
    deleteTheme,
    setActiveTheme,
    toggleAppearance,
  };
}
