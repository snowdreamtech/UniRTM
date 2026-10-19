// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { describe, it, expect, vi, beforeEach } from "vitest";
import { useGuard } from "./useGuard";

describe("useGuard", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it("fetches guard status successfully", async () => {
    const mockStatus = {
      mode: "staged",
      scanned: 3,
      clean: false,
      errors_count: 0,
      warnings_count: 2,
      fixable_count: 2,
      issues: [
        {
          file_path: "src/dirty.ts",
          line: 1,
          column: 0,
          rule_id: "builtin:trailing-whitespace",
          category: "hygiene",
          phase: "fixer",
          severity: "warning",
          message: "Contains trailing whitespace",
          fixable: true,
        },
      ],
    };

    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: vi.fn().mockResolvedValue(mockStatus),
    });

    const { status, fetchStatus } = useGuard();
    const res = await fetchStatus("staged");

    expect(res).toEqual(mockStatus);
    expect(status.value.scanned).toBe(3);
    expect(status.value.fixable_count).toBe(2);
    expect(status.value.issues.length).toBe(1);
  });

  it("executes runFix and re-fetches status", async () => {
    const mockPostRes = { success: true };
    const mockCleanStatus = {
      mode: "staged",
      scanned: 3,
      clean: true,
      errors_count: 0,
      warnings_count: 0,
      fixable_count: 0,
      issues: [],
    };

    global.fetch = vi
      .fn()
      .mockResolvedValueOnce({
        ok: true,
        json: vi.fn().mockResolvedValue(mockPostRes),
      })
      .mockResolvedValueOnce({
        ok: true,
        json: vi.fn().mockResolvedValue(mockCleanStatus),
      });

    const { status, runFix } = useGuard();
    const ok = await runFix("staged");

    expect(ok).toBe(true);
    expect(status.value.clean).toBe(true);
    expect(status.value.fixable_count).toBe(0);
  });
});
