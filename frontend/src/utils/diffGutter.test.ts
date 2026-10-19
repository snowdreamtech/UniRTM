// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { describe, it, expect } from "vitest";
import { computeLineDiff } from "./diffGutter";

describe("diffGutter utility", () => {
  it("returns empty array when contents are identical", () => {
    const text = "line1\nline2\nline3";
    expect(computeLineDiff(text, text)).toEqual([]);
  });

  it("detects added lines accurately", () => {
    const orig = "line1\nline3";
    const mod = "line1\nline2\nline3";
    const diffs = computeLineDiff(orig, mod);

    expect(diffs.length).toBe(1);
    expect(diffs[0]).toEqual({
      type: "added",
      startLineNumber: 2,
      endLineNumber: 2,
    });
  });

  it("detects modified lines accurately", () => {
    const orig = "alpha\nbeta\ngamma";
    const mod = "alpha\nBETA\ngamma";
    const diffs = computeLineDiff(orig, mod);

    expect(diffs.length).toBe(1);
    expect(diffs[0]).toEqual({
      type: "modified",
      startLineNumber: 2,
      endLineNumber: 2,
    });
  });

  it("detects deleted line position accurately", () => {
    const orig = "line1\nline2\nline3";
    const mod = "line1\nline3";
    const diffs = computeLineDiff(orig, mod);

    expect(diffs.length).toBe(1);
    expect(diffs[0].type).toBe("deleted");
  });

  it("handles empty initial text turning into new lines", () => {
    const diffs = computeLineDiff("", "hello\nworld");
    expect(diffs.length).toBe(1);
    expect(diffs[0]).toEqual({
      type: "added",
      startLineNumber: 1,
      endLineNumber: 2,
    });
  });
});
