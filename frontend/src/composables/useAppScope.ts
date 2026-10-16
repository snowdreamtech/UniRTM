// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { ref, computed } from "vue";
import type { Router, RouteLocationNormalizedLoaded } from "vue-router";

export type AppScope = "workspace" | "user";

export const SCOPE_STORAGE_KEY = "unirtm:active_scope";

function getSafeStorage(): Storage | null {
  if (typeof window !== "undefined" && window.localStorage) {
    return window.localStorage;
  }
  if (typeof globalThis !== "undefined" && (globalThis as any).localStorage) {
    try {
      const s = (globalThis as any).localStorage;
      s.getItem("__probe__");
      return s;
    } catch {
      return null;
    }
  }
  return null;
}

function getInitialScope(): AppScope {
  if (typeof window !== "undefined") {
    // 1. Check URL search param first
    try {
      const urlParams = new URLSearchParams(window.location.search);
      const urlScope = urlParams.get("scope");
      if (urlScope === "user" || urlScope === "workspace") {
        return urlScope;
      }
    } catch {}
  }

  // 2. Check localStorage
  const storage = getSafeStorage();
  if (storage) {
    try {
      const saved = storage.getItem(SCOPE_STORAGE_KEY);
      if (saved === "user" || saved === "workspace") {
        return saved;
      }
    } catch {}
  }

  return "workspace";
}

const activeScope = ref<AppScope>(getInitialScope());

export function useAppScope() {
  function setScope(scope: AppScope, router?: Router) {
    if (activeScope.value === scope) return;
    activeScope.value = scope;

    const storage = getSafeStorage();
    if (storage) {
      try {
        storage.setItem(SCOPE_STORAGE_KEY, scope);
      } catch {}
    }

    if (router && router.currentRoute?.value) {
      const cur = router.currentRoute.value;
      if (cur.query.scope !== scope) {
        router.replace({
          name: cur.name || undefined,
          params: cur.params,
          query: {
            ...cur.query,
            scope,
          },
        }).catch(() => {});
      }
    }
  }

  function toggleScope(router?: Router) {
    setScope(activeScope.value === "workspace" ? "user" : "workspace", router);
  }

  function syncFromRoute(route: RouteLocationNormalizedLoaded) {
    const scope = route.query.scope;
    if (scope === "user" || scope === "workspace") {
      if (activeScope.value !== scope) {
        activeScope.value = scope;
        const storage = getSafeStorage();
        if (storage) {
          try {
            storage.setItem(SCOPE_STORAGE_KEY, scope);
          } catch {}
        }
      }
    }
  }

  const isWorkspace = computed(() => activeScope.value === "workspace");
  const isUser = computed(() => activeScope.value === "user");

  return {
    activeScope,
    isWorkspace,
    isUser,
    setScope,
    toggleScope,
    syncFromRoute,
  };
}
