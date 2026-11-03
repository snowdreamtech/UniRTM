import { describe, it, expect } from "vitest";
import { getXtermTheme } from "./terminalTheme";

describe("terminalTheme", () => {
  it("returns dark theme palette with high contrast foreground and 16 ANSI colors", () => {
    const theme = getXtermTheme(true);
    expect(theme.background).toBe("#18181b");
    expect(theme.foreground).toBe("#f1f5f9");
    expect(theme.cursor).toBe("#38bdf8");
    expect(theme.green).toBe("#4ade80");
    expect(theme.red).toBe("#f87171");
    expect(theme.cyan).toBe("#22d3ee");
  });

  it("returns light theme palette with pure white background and dark slate text", () => {
    const theme = getXtermTheme(false);
    expect(theme.background).toBe("#ffffff");
    expect(theme.foreground).toBe("#0f172a");
    expect(theme.cursor).toBe("#0284c7");
    expect(theme.green).toBe("#16a34a");
    expect(theme.red).toBe("#dc2626");
    expect(theme.cyan).toBe("#0891b2");
  });
});
