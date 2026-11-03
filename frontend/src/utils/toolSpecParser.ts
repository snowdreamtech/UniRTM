// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

export interface ToolPrefixOption {
  value: string;
  label: string;
  placeholder: string;
  description?: string;
}

export const COMMON_TOOL_PREFIXES: ToolPrefixOption[] = [
  { value: "", label: "Default / Core", placeholder: "e.g.: zig, elixir, php, julia..." },
  { value: "github:", label: "github: (Releases)", placeholder: "e.g.: sharkdp/bat, BurntSushi/ripgrep..." },
  { value: "npm:", label: "npm: (Node.js)", placeholder: "e.g.: @biomejs/biome, prettier, pnpm..." },
  {
    value: "go:",
    label: "go: (Go Module)",
    placeholder: "e.g.: github.com/golangci/golangci-lint/cmd/golangci-lint...",
  },
  { value: "pipx:", label: "pipx: (Python)", placeholder: "e.g.: ruff, uv, black, poetry..." },
  { value: "cargo:", label: "cargo: (Rust)", placeholder: "e.g.: ripgrep, eza, bat, zellij..." },
  { value: "aqua:", label: "aqua: (Aqua)", placeholder: "e.g.: aquaproj/aqua..." },
  { value: "asdf:", label: "asdf: (Plugin)", placeholder: "e.g.: asdf-vm/asdf-nodejs..." },
];

export interface ParsedToolSpec {
  prefix: string;
  name: string;
  version: string;
}

const KNOWN_PREFIXES = [
  "npm:",
  "cargo:",
  "pipx:",
  "pypi:",
  "go:",
  "github:",
  "aqua:",
  "asdf:",
  "vfox:",
  "composer:",
  "gem:",
];

/**
 * Intelligently parse full tool spec string (e.g. "npm:@biomejs/biome@1.9.4", "cargo:ripgrep@14.1.0", or "zig@0.13.0")
 */
export function parseToolSpec(raw: string): ParsedToolSpec {
  const trimmed = (raw || "").trim();
  if (!trimmed) {
    return { prefix: "", name: "", version: "" };
  }

  let prefix = "";
  let rest = trimmed;

  // 1. Match known prefix
  for (const p of KNOWN_PREFIXES) {
    if (rest.toLowerCase().startsWith(p.toLowerCase())) {
      prefix = p;
      rest = rest.slice(p.length).trim();
      break;
    }
  }

  // 2. Match version (split by last @, excluding leading @ in npm scoped package)
  let name = rest;
  let version = "";

  const lastAtIdx = rest.lastIndexOf("@");
  // Must be index > 0 to ensure it is not leading @ in scoped package
  if (lastAtIdx > 0) {
    name = rest.slice(0, lastAtIdx).trim();
    version = rest.slice(lastAtIdx + 1).trim();
  }

  return { prefix, name, version };
}

/**
 * Combine prefix, tool name, and pinned version into finalized spec and preview string
 */
export function composeToolSpec(
  prefix: string,
  name: string,
  version: string
): { toolKey: string; version: string; previewToml: string } {
  const cleanName = (name || "").trim();
  const cleanVer = (version || "").trim();
  const cleanPrefix = (prefix || "").trim();

  const toolKey = cleanPrefix ? `${cleanPrefix}${cleanName}` : cleanName;
  const previewToml = cleanName ? `"${toolKey}" = "${cleanVer || "..."}"` : "";

  return {
    toolKey,
    version: cleanVer,
    previewToml,
  };
}
