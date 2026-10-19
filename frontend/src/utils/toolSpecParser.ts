// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

export interface ToolPrefixOption {
  value: string;
  label: string;
  placeholder: string;
  description?: string;
}

export const COMMON_TOOL_PREFIXES: ToolPrefixOption[] = [
  { value: "", label: "默认 / 原生 (Core)", placeholder: "例如: zig, elixir, php, julia..." },
  { value: "npm:", label: "npm: (Node.js)", placeholder: "例如: @biomejs/biome, prettier, pnpm..." },
  { value: "cargo:", label: "cargo: (Rust)", placeholder: "例如: ripgrep, eza, bat, zellij..." },
  { value: "pipx:", label: "pipx: (Python)", placeholder: "例如: ruff, uv, black, poetry..." },
  { value: "go:", label: "go: (Go Module)", placeholder: "例如: github.com/golangci/golangci-lint/cmd/golangci-lint..." },
  { value: "github:", label: "github: (Releases)", placeholder: "例如: sharkdp/bat, BurntSushi/ripgrep..." },
  { value: "aqua:", label: "aqua: (Aqua)", placeholder: "例如: aquaproj/aqua..." },
  { value: "asdf:", label: "asdf: (Plugin)", placeholder: "例如: asdf-vm/asdf-nodejs..." },
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
 * 智能解析全格式工具规范字符串（如 "npm:@biomejs/biome@1.9.4" 或 "cargo:ripgrep@14.1.0" 或 "zig@0.13.0"）
 */
export function parseToolSpec(raw: string): ParsedToolSpec {
  const trimmed = (raw || "").trim();
  if (!trimmed) {
    return { prefix: "", name: "", version: "" };
  }

  let prefix = "";
  let rest = trimmed;

  // 1. 匹配已知前缀
  for (const p of KNOWN_PREFIXES) {
    if (rest.toLowerCase().startsWith(p.toLowerCase())) {
      prefix = p;
      rest = rest.slice(p.length).trim();
      break;
    }
  }

  // 2. 匹配版本号 (按最后一个 @ 拆分，但要排除 npm 作用域包首字符 @，例如 @biomejs/biome)
  let name = rest;
  let version = "";

  const lastAtIdx = rest.lastIndexOf("@");
  // 必须在 index > 0 处，确保不是 npm 作用域包开头的 @，例如 "@scope/pkg"
  if (lastAtIdx > 0) {
    name = rest.slice(0, lastAtIdx).trim();
    version = rest.slice(lastAtIdx + 1).trim();
  }

  return { prefix, name, version };
}

/**
 * 将前缀、工具名与固定版本组合为最终持久化规范与配置预览字符串
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
  const previewToml = cleanName
    ? `"${toolKey}" = "${cleanVer || "..."}"`
    : "";

  return {
    toolKey,
    version: cleanVer,
    previewToml,
  };
}
