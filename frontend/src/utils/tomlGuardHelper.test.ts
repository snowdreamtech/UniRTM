// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { describe, it, expect } from "vitest";
import { parseGuardConfigFromToml, updateBuiltinRuleInToml } from "./tomlGuardHelper";

describe("tomlGuardHelper", () => {
  const sampleToml = `
[guard]
enabled = true
skip = [
  "builtin:forbid-debug-statements",
]

[guard.rules."builtin:check-added-large-files"]
max_file_size = 5242880
exclude_dirs = ["vendor", "testdata"]

[guard.slots."go.linter"]
tool = "go"
command = "go vet ./..."
`;

  it("parses skip list and custom rules from TOML", () => {
    const parsed = parseGuardConfigFromToml(sampleToml);
    expect(parsed.skip).toContain("builtin:forbid-debug-statements");
    expect(parsed.rules["builtin:check-added-large-files"]).toBeDefined();
    expect(parsed.rules["builtin:check-added-large-files"].max_file_size).toBe(5242880);
    expect(parsed.rules["builtin:check-added-large-files"].exclude_dirs).toEqual(["vendor", "testdata"]);
  });

  it("adds a rule to skip list when disabled", () => {
    const updated = updateBuiltinRuleInToml(sampleToml, {
      id: "builtin:check-symlinks",
      enabled: false,
      severity: "default",
      exclude_dirs: [],
      exclude_files: [],
      extensions: [],
    });

    const parsed = parseGuardConfigFromToml(updated);
    expect(parsed.skip).toContain("builtin:forbid-debug-statements");
    expect(parsed.skip).toContain("builtin:check-symlinks");
  });

  it("removes a rule from skip list when re-enabled", () => {
    const updated = updateBuiltinRuleInToml(sampleToml, {
      id: "builtin:forbid-debug-statements",
      enabled: true,
      severity: "default",
      exclude_dirs: [],
      exclude_files: [],
      extensions: [],
    });

    const parsed = parseGuardConfigFromToml(updated);
    expect(parsed.skip).not.toContain("builtin:forbid-debug-statements");
  });

  it("updates custom rule properties", () => {
    const updated = updateBuiltinRuleInToml(sampleToml, {
      id: "builtin:check-added-large-files",
      enabled: true,
      severity: "blocker",
      max_file_size: 10485760,
      exclude_dirs: ["vendor", "dist"],
      exclude_files: ["*.bundle.js"],
      extensions: [],
    });

    const parsed = parseGuardConfigFromToml(updated);
    expect(parsed.rules["builtin:check-added-large-files"].severity).toBe("blocker");
    expect(parsed.rules["builtin:check-added-large-files"].max_file_size).toBe(10485760);
    expect(parsed.rules["builtin:check-added-large-files"].exclude_dirs).toEqual(["vendor", "dist"]);
    expect(parsed.rules["builtin:check-added-large-files"].exclude_files).toEqual(["*.bundle.js"]);
  });
});
