// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { describe, it, expect } from "vitest";
import { formatGuardDuration } from "./guardRulesMeta";

describe("formatGuardDuration", () => {
  it("formats undefined or empty values safely", () => {
    expect(formatGuardDuration(undefined)).toBe("");
    expect(formatGuardDuration(null as any)).toBe("");
    expect(formatGuardDuration("")).toBe("");
  });

  it("formats sub-millisecond durations as '< 1ms' instead of deceptive 0ms", () => {
    expect(formatGuardDuration(0)).toBe("< 1ms");
    expect(formatGuardDuration(0.2)).toBe("< 1ms");
    expect(formatGuardDuration("0ms")).toBe("< 1ms");
    expect(formatGuardDuration("0.3ms")).toBe("< 1ms");
  });

  it("converts raw nanoseconds to milliseconds", () => {
    // 250,000 ns = 0.25ms -> "< 1ms"
    expect(formatGuardDuration(250000)).toBe("< 1ms");
    // 15,000,000 ns = 15ms -> "15ms"
    expect(formatGuardDuration(15000000)).toBe("15ms");
  });

  it("formats standard millisecond values", () => {
    expect(formatGuardDuration(18)).toBe("18ms");
    expect(formatGuardDuration("120ms")).toBe("120ms");
    expect(formatGuardDuration(450)).toBe("450ms");
  });

  it("formats seconds for large durations >= 1000ms", () => {
    expect(formatGuardDuration(1200)).toBe("1.2s");
    expect(formatGuardDuration("2500ms")).toBe("2.5s");
  });
});
