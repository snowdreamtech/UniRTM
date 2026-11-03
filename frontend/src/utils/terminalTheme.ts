// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import type { ITheme } from "@xterm/xterm";

/**
 * Returns a high-contrast, beautiful ANSI palette for xterm.js matching UniRTM's theme.
 */
export function getXtermTheme(isDark: boolean): ITheme {
  if (isDark) {
    return {
      background: "#18181b",
      foreground: "#f1f5f9",
      cursor: "#38bdf8",
      cursorAccent: "#18181b",
      selectionBackground: "rgba(56, 189, 248, 0.35)",
      selectionForeground: "#ffffff",
      black: "#1e293b",
      red: "#f87171",
      green: "#4ade80",
      yellow: "#facc15",
      blue: "#60a5fa",
      magenta: "#c084fc",
      cyan: "#22d3ee",
      white: "#f8fafc",
      brightBlack: "#475569",
      brightRed: "#ef4444",
      brightGreen: "#22c55e",
      brightYellow: "#eab308",
      brightBlue: "#3b82f6",
      brightMagenta: "#a855f7",
      brightCyan: "#06b6d4",
      brightWhite: "#ffffff",
    };
  }

  // Light theme: High contrast, dark slate text, bright colorful ANSI
  return {
    background: "#ffffff",
    foreground: "#0f172a",
    cursor: "#0284c7",
    cursorAccent: "#ffffff",
    selectionBackground: "rgba(2, 132, 199, 0.25)",
    selectionForeground: "#0f172a",
    black: "#0f172a",
    red: "#dc2626",
    green: "#16a34a",
    yellow: "#ca8a04",
    blue: "#2563eb",
    magenta: "#9333ea",
    cyan: "#0891b2",
    white: "#e2e8f0",
    brightBlack: "#475569",
    brightRed: "#b91c1c",
    brightGreen: "#15803d",
    brightYellow: "#a16207",
    brightBlue: "#1d4ed8",
    brightMagenta: "#7e22ce",
    brightCyan: "#0e7490",
    brightWhite: "#0f172a",
  };
}
