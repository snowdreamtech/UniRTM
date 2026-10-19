// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { describe, it, expect } from "vitest";
import { parseToolSpec, composeToolSpec, COMMON_TOOL_PREFIXES } from "./toolSpecParser";

describe("toolSpecParser", () => {
  describe("COMMON_TOOL_PREFIXES", () => {
    it("includes standard prefixes in exact designated order", () => {
      expect(COMMON_TOOL_PREFIXES.map((p) => p.value)).toEqual([
        "",
        "github:",
        "npm:",
        "go:",
        "pipx:",
        "cargo:",
        "aqua:",
        "asdf:",
      ]);
    });
  });

  describe("parseToolSpec", () => {
    it("handles empty or whitespace string", () => {
      expect(parseToolSpec("")).toEqual({ prefix: "", name: "", version: "" });
      expect(parseToolSpec("   ")).toEqual({ prefix: "", name: "", version: "" });
    });

    it("parses tool with prefix and version (e.g. cargo:ripgrep@14.1.0)", () => {
      const res = parseToolSpec("cargo:ripgrep@14.1.0");
      expect(res.prefix).toBe("cargo:");
      expect(res.name).toBe("ripgrep");
      expect(res.version).toBe("14.1.0");
    });

    it("parses npm scoped package with version (e.g. npm:@biomejs/biome@1.9.4)", () => {
      const res = parseToolSpec("npm:@biomejs/biome@1.9.4");
      expect(res.prefix).toBe("npm:");
      expect(res.name).toBe("@biomejs/biome");
      expect(res.version).toBe("1.9.4");
    });

    it("parses npm scoped package without version", () => {
      const res = parseToolSpec("npm:@biomejs/biome");
      expect(res.prefix).toBe("npm:");
      expect(res.name).toBe("@biomejs/biome");
      expect(res.version).toBe("");
    });

    it("parses standard tool with version (e.g. zig@0.13.0)", () => {
      const res = parseToolSpec("zig@0.13.0");
      expect(res.prefix).toBe("");
      expect(res.name).toBe("zig");
      expect(res.version).toBe("0.13.0");
    });

    it("parses plain tool name without prefix or version", () => {
      const res = parseToolSpec("elixir");
      expect(res.prefix).toBe("");
      expect(res.name).toBe("elixir");
      expect(res.version).toBe("");
    });

    it("parses github release with repo and version", () => {
      const res = parseToolSpec("github:sharkdp/bat@v0.24.0");
      expect(res.prefix).toBe("github:");
      expect(res.name).toBe("sharkdp/bat");
      expect(res.version).toBe("v0.24.0");
    });
  });

  describe("composeToolSpec", () => {
    it("composes prefixed tool spec and preview toml", () => {
      const res = composeToolSpec("npm:", "@biomejs/biome", "1.9.4");
      expect(res.toolKey).toBe("npm:@biomejs/biome");
      expect(res.version).toBe("1.9.4");
      expect(res.previewToml).toBe('"npm:@biomejs/biome" = "1.9.4"');
    });

    it("composes core tool spec without prefix", () => {
      const res = composeToolSpec("", "zig", "0.13.0");
      expect(res.toolKey).toBe("zig");
      expect(res.version).toBe("0.13.0");
      expect(res.previewToml).toBe('"zig" = "0.13.0"');
    });

    it("handles placeholder when version is empty", () => {
      const res = composeToolSpec("cargo:", "ripgrep", "");
      expect(res.toolKey).toBe("cargo:ripgrep");
      expect(res.version).toBe("");
      expect(res.previewToml).toBe('"cargo:ripgrep" = "..."');
    });

    it("returns empty preview when name is empty", () => {
      const res = composeToolSpec("cargo:", "", "");
      expect(res.previewToml).toBe("");
    });
  });
});
