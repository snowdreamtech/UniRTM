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

  it("rejects 0.x for Rust", () => {
    const res = validateRuntimeVersion("rust", "0.99.1");
    expect(res.valid).toBe(false);
    expect(res.error).toContain("Rust 官方稳定版本均在 1.0 及以上");
  });

  it("accepts valid Rust versions", () => {
    expect(validateRuntimeVersion("rust", "1.85.0").valid).toBe(true);
  });

  it("rejects Python versions outside 2.x and 3.x", () => {
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
