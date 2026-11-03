// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

/**
 * Detects whether the current runtime environment is macOS/iOS.
 */
export const isMac: boolean =
  typeof navigator !== "undefined" &&
  /Mac|Macintosh|Mac OS X|iPhone|iPad|iPod/i.test(navigator.platform || navigator.userAgent);

/**
 * Formats a shortcut string according to the operating system.
 * On macOS, converts "Ctrl+K" / "Cmd+K" to standard Apple symbol "⌘K",
 * "Shift" to "⇧", and "Alt" to "⌥".
 * On Windows/Linux, keeps "Ctrl+K" style.
 *
 * @param keys The shortcut string, e.g. "Ctrl+K", "Ctrl+Shift+P", "Ctrl+K Ctrl+S"
 * @param mac Optional override for platform detection (defaults to runtime isMac)
 * @returns Localized/platform-friendly shortcut label
 */
export function formatShortcut(keys: string, mac: boolean = isMac): string {
  if (!keys) return "";

  if (!mac) {
    return keys.replace(/\bCmd\b/gi, "Ctrl");
  }

  return keys
    .split(" ")
    .map((part) => {
      return part
        .replace(/Ctrl\+/gi, "⌘")
        .replace(/Cmd\+/gi, "⌘")
        .replace(/Shift\+/gi, "⇧")
        .replace(/Alt\+/gi, "⌥")
        .replace(/Option\+/gi, "⌥");
    })
    .join(" ");
}
