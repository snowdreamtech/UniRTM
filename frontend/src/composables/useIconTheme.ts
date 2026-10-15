// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { ref, onMounted, onUnmounted } from "vue";
import { iconThemeService } from "../services/icon-theme/iconThemeService";
import type { ResolvedIcon } from "../types/icon-theme";

const themeRevision = ref(0);

export function useIconTheme() {
  iconThemeService.ensureStylesInjected();

  onMounted(() => {
    iconThemeService.ensureStylesInjected();
  });

  const unsubscribe = iconThemeService.subscribe(() => {
    themeRevision.value++;
  });

  onUnmounted(() => {
    unsubscribe();
  });

  function resolveIcon(id: string): ResolvedIcon {
    // Access themeRevision so that Vue computed properties track this reactive dependency
    void themeRevision.value;
    return iconThemeService.resolveIcon(id);
  }

  return {
    resolveIcon,
    iconThemeService,
    themeRevision,
  };
}
