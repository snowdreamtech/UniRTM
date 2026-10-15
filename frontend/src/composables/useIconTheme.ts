// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { onMounted } from "vue";
import { iconThemeService } from "../services/icon-theme/iconThemeService";
import type { ResolvedIcon } from "../types/icon-theme";

export function useIconTheme() {
  iconThemeService.ensureStylesInjected();

  onMounted(() => {
    iconThemeService.ensureStylesInjected();
  });

  function resolveIcon(id: string): ResolvedIcon {
    return iconThemeService.resolveIcon(id);
  }

  return {
    resolveIcon,
    iconThemeService,
  };
}
