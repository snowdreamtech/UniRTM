// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { ref } from "vue";
import { iconThemeService } from "../services/icon-theme/iconThemeService";
import type { ResolvedIcon } from "../types/icon-theme";

const themeRevision = ref(0);
let isSubscribed = false;

function ensureSubscription(): void {
  if (isSubscribed) return;
  iconThemeService.ensureStylesInjected();
  iconThemeService.subscribe(() => {
    themeRevision.value++;
  });
  isSubscribed = true;
}

export function useIconTheme() {
  ensureSubscription();

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
