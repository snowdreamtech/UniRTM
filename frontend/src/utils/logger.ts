// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

export function logUserAction(level: string, message: string, details: string = ""): void {
  if (import.meta.env?.DEV) {
    console.debug(`[${level}] ${message}`, details);
  }
}
