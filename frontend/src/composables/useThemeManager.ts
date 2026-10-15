// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { ref } from "vue";
import type { ThemeItem } from "../types/theme-manager";
import { useColorTheme } from "./useColorTheme";
import { iconThemeService } from "../services/icon-theme/iconThemeService";

const defaultColors: ThemeItem[] = [
  { id: "dark", name: "Dark (UniRTM Default)", type: "color", is_builtin: true, ui_theme: "vs-dark" },
  { id: "light", name: "Light", type: "color", is_builtin: true, ui_theme: "vs" },
  { id: "system", name: "Follow System", type: "color", is_builtin: true, ui_theme: "vs-dark" },
];
const defaultIcons: ThemeItem[] = [
  { id: "seti", name: "Seti UI (VS Code Default)", type: "icon", is_builtin: true, description: "Official Visual Studio Code default file icon theme" },
];

const colorThemes = ref<ThemeItem[]>(defaultColors);
const iconThemes = ref<ThemeItem[]>(defaultIcons);
const activeColorTheme = ref<string>("dark");
const activeIconTheme = ref<string>("seti");
const loading = ref<boolean>(false);
const error = ref<string | null>(null);

const { applyTheme: applyColorTheme } = useColorTheme();

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
      activeColorTheme.value = data.active?.color || data.active_color_theme || "dark";
      activeIconTheme.value = data.active?.icon || data.active_icon_theme || "seti";

      // Apply active themes hot
      await applyColorTheme(activeColorTheme.value, colorThemes.value);
      const activeIconItem = iconThemes.value.find((i) => i.id === activeIconTheme.value);
      if (activeIconItem) {
        await iconThemeService.loadAndApplyTheme(activeIconItem);
      } else {
        await iconThemeService.switchTheme("seti");
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

  const deleteTheme = async (type: "color" | "icon", packageId: string) => {
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
          await setActiveTheme("dark", undefined);
        }
      } else if (type === "icon") {
        const deletedIcon = iconThemes.value.find((i) => i.package_id === packageId && i.id === activeIconTheme.value);
        if (deletedIcon) {
          await setActiveTheme(undefined, "seti");
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

  const setActiveTheme = async (colorId?: string, iconId?: string) => {
    loading.value = true;
    error.value = null;
    try {
      const payload: Record<string, string> = {};
      if (colorId) payload.color = colorId;
      if (iconId) payload.icon = iconId;

      const res = await fetch("/api/themes/apply", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });

      const data = await res.json();
      if (!res.ok || data.error) {
        throw new Error(data.error || "Failed to apply theme");
      }

      if (colorId) {
        activeColorTheme.value = colorId;
        await applyColorTheme(colorId, colorThemes.value);
      }

      if (iconId) {
        activeIconTheme.value = iconId;
        const iconItem = iconThemes.value.find((i) => i.id === iconId);
        if (iconItem) {
          await iconThemeService.loadAndApplyTheme(iconItem);
        } else {
          await iconThemeService.switchTheme("seti");
        }
      }
    } catch (err: any) {
      error.value = err.message || "Failed to apply theme";
      throw err;
    } finally {
      loading.value = false;
    }
  };

  return {
    colorThemes,
    iconThemes,
    activeColorTheme,
    activeIconTheme,
    loading,
    error,
    fetchThemes,
    installTheme,
    deleteTheme,
    setActiveTheme,
  };
}
