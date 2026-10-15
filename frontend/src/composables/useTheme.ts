// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { ref, computed } from "vue";

export type AppTheme = "dark" | "light" | "system";

const STORAGE_KEY = "unirtm_theme_cache";

function getSystemPreferredTheme(): "dark" | "light" {
  if (typeof window !== "undefined" && window.matchMedia) {
    return window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
  }
  return "dark";
}

function resolveEffectiveTheme(themeName: string): "dark" | "light" {
  if (themeName === "system") {
    return getSystemPreferredTheme();
  }
  return themeName === "light" ? "light" : "dark";
}

function getInitialTheme(): AppTheme {
  try {
    const cached = localStorage.getItem(STORAGE_KEY) || localStorage.getItem("unigo_theme_cache");
    if (cached === "light" || cached === "dark" || cached === "system") {
      return cached as AppTheme;
    }
  } catch (e) {
    // localStorage may be unavailable
  }

  return "system";
}

const currentTheme = ref<string>(getInitialTheme());
const effectiveTheme = computed<"dark" | "light">(() => resolveEffectiveTheme(currentTheme.value));
const isDarkTheme = computed(() => effectiveTheme.value === "dark");

function updateDomTheme(applied: "dark" | "light") {
  if (typeof document !== "undefined") {
    document.documentElement.setAttribute("data-theme", applied);
  }
}

// Apply initial theme immediately upon module evaluation
updateDomTheme(resolveEffectiveTheme(currentTheme.value));

// Global listener for OS system theme changes
if (typeof window !== "undefined" && window.matchMedia) {
  const mql = window.matchMedia("(prefers-color-scheme: dark)");
  const handleSystemThemeChange = (e: MediaQueryListEvent | MediaQueryList) => {
    if (currentTheme.value === "system") {
      const nextApplied = e.matches ? "dark" : "light";
      updateDomTheme(nextApplied);
    }
  };

  if (typeof mql.addEventListener === "function") {
    mql.addEventListener("change", handleSystemThemeChange);
  } else if (typeof (mql as any).addListener === "function") {
    (mql as any).addListener(handleSystemThemeChange);
  }
}

export function useTheme() {
  function applyTheme(themeName: string) {
    let theme = "system";
    if (themeName === "light" || themeName === "dark" || themeName === "system") {
      theme = themeName;
    }

    currentTheme.value = theme;
    const applied = resolveEffectiveTheme(theme);

    try {
      localStorage.setItem(STORAGE_KEY, theme);
    } catch (e) {
      // localStorage may be unavailable
    }

    updateDomTheme(applied);
  }

  function getActiveTheme(): AppTheme {
    return currentTheme.value as AppTheme;
  }

  function toggleTheme(): string {
    const active = resolveEffectiveTheme(currentTheme.value);
    const nextTheme = active === "light" ? "dark" : "light";
    applyTheme(nextTheme);
    return nextTheme;
  }

  return {
    appearance: currentTheme,
    currentTheme,
    effectiveTheme,
    isDarkTheme,
    applyTheme,
    setAppearance: applyTheme,
    getActiveTheme,
    toggleTheme,
  };
}

export function syncWindowTheme() {
  updateDomTheme(resolveEffectiveTheme(currentTheme.value));
}
