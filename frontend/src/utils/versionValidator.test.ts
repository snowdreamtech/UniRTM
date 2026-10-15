// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { describe, it, expect } from "vitest";
import { validateRuntimeVersion } from "./versionValidator";

describe("validateRuntimeVersion", () => {
  it("rejects empty or whitespace inputs", () => {
    expect(validateRuntimeVersion("go", "").valid).toBe(false);
    expect(validateRuntimeVersion("go", "   ").valid).toBe(false);
  });

  it("rejects illegal characters", () => {
    const res = validateRuntimeVersion("go", "1.24@beta!");
    expect(res.valid).toBe(false);
    expect(res.error).toBeDefined();
  });

  it("rejects leading zero non-semver numbers like 01.2.3", () => {
    const res = validateRuntimeVersion("go", "01.2.3");
    expect(res.valid).toBe(false);
  });

  it("rejects 0.x versions for Go (e.g. 0.1111)", () => {
    const res = validateRuntimeVersion("go", "0.1111");
    expect(res.valid).toBe(false);
    expect(res.error).toContain("Go 官方版本均在 1.0 及以上");
  });

  it("accepts valid Go semantic versions", () => {
    expect(validateRuntimeVersion("go", "1.24.0").valid).toBe(true);
    expect(validateRuntimeVersion("go", "1.24").valid).toBe(true);
    expect(validateRuntimeVersion("go", "v1.23.6").valid).toBe(true);
    expect(validateRuntimeVersion("go", "1.24-rc1").valid).toBe(true);
  });

  it("rejects 0.x for Rust SemVer", () => {
    const res = validateRuntimeVersion("rust", "0.99.1");
    expect(res.valid).toBe(false);
    expect(res.error).toContain("Rust 官方稳定版本均在 1.0 及以上");
  });

  it("accepts valid Rust versions", () => {
    expect(validateRuntimeVersion("rust", "1.85.0").valid).toBe(true);
  });

  it("accepts CalVer (Calendar Versioning) with 4-digit and 2-digit years", () => {
    // 4-digit year with 2-digit month and day (e.g. 2024.03.15)
    const cal1 = validateRuntimeVersion("python", "2024.03.15");
    expect(cal1.valid).toBe(true);
    expect(cal1.kind).toBe("calver");

    // 4-digit year with single-digit month (e.g. 2024.1.2)
    const cal2 = validateRuntimeVersion("python", "2024.1.2");
    expect(cal2.valid).toBe(true);
    expect(cal2.kind).toBe("calver");

    // 2-digit year (e.g. 24.04, 24.04.1)
    const cal3 = validateRuntimeVersion("general", "24.04");
    expect(cal3.valid).toBe(true);
    expect(cal3.kind).toBe("calver");

    const cal4 = validateRuntimeVersion("general", "24.04.1");
    expect(cal4.valid).toBe(true);
    expect(cal4.kind).toBe("calver");
  });

  it("accepts ISO date-based versions (YYYY-MM-DD)", () => {
    const res = validateRuntimeVersion("general", "2024-03-15");
    expect(res.valid).toBe(true);
    expect(res.kind).toBe("calver");
  });

  it("accepts toolchain date channels (e.g. Rust nightly-YYYY-MM-DD)", () => {
    const res = validateRuntimeVersion("rust", "nightly-2024-03-01");
    expect(res.valid).toBe(true);
    expect(res.kind).toBe("channel_date");

    const resBeta = validateRuntimeVersion("rust", "beta-2024-01-15");
    expect(resBeta.valid).toBe(true);
    expect(resBeta.kind).toBe("channel_date");
  });

  it("accepts channel aliases (nightly, beta, stable, lts, latest)", () => {
    expect(validateRuntimeVersion("rust", "nightly").valid).toBe(true);
    expect(validateRuntimeVersion("node", "lts").valid).toBe(true);
    expect(validateRuntimeVersion("go", "latest").valid).toBe(true);
  });

  it("rejects Python versions outside 2.x and 3.x for standard semver", () => {
    expect(validateRuntimeVersion("python", "0.5.0").valid).toBe(false);
    expect(validateRuntimeVersion("python", "5.1.0").valid).toBe(false);
    expect(validateRuntimeVersion("python", "3.12.2").valid).toBe(true);
  });

  it("validates Node.js legacy 0.x vs modern", () => {
    expect(validateRuntimeVersion("node", "0.10.48").valid).toBe(true);
    expect(validateRuntimeVersion("node", "0.12.18").valid).toBe(true);
    expect(validateRuntimeVersion("node", "0.1111").valid).toBe(false);
    expect(validateRuntimeVersion("node", "20.18.3").valid).toBe(true);
  });
});
