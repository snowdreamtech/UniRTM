// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { describe, it, expect } from "vitest";
import { formatShortcut } from "./platform";

describe("platform utility formatShortcut", () => {
  it("formats Ctrl+K for macOS as ⌘K", () => {
    expect(formatShortcut("Ctrl+K", true)).toBe("⌘K");
  });

  it("formats Ctrl+K for Windows/Linux as Ctrl+K", () => {
    expect(formatShortcut("Ctrl+K", false)).toBe("Ctrl+K");
  });

  it("formats Cmd+K for Windows/Linux as Ctrl+K", () => {
    expect(formatShortcut("Cmd+K", false)).toBe("Ctrl+K");
  });

  it("formats chords like Ctrl+K Ctrl+S for macOS as ⌘K ⌘S", () => {
    expect(formatShortcut("Ctrl+K Ctrl+S", true)).toBe("⌘K ⌘S");
  });

  it("formats complex combo like Ctrl+Shift+P for macOS as ⌘⇧P", () => {
    expect(formatShortcut("Ctrl+Shift+P", true)).toBe("⌘⇧P");
  });

  it("handles empty or falsy string gracefully", () => {
    expect(formatShortcut("", true)).toBe("");
  });
});
