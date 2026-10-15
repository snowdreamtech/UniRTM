// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { t } from "../i18n";

export interface VersionValidationResult {
  valid: boolean;
  error?: string;
  isCompleteSemver?: boolean;
}

/**
 * Validates whether an input string is a legitimate semantic version
 * and satisfies tool-specific constraints (e.g. Go starting at 1.x).
 */
export function validateRuntimeVersion(toolId: string, input: string): VersionValidationResult {
  const val = input.trim().replace(/^v/i, "");
  if (!val) {
    return { valid: false };
  }

  // 1. Basic allowed characters: alphanumeric, dots, hyphens, plus, underscores
  if (!/^[0-9a-zA-Z._+-]+$/.test(val)) {
    return {
      valid: false,
      error: t("runtimes.version_invalid_format") || "版本号格式不合规（包含非法字符）",
    };
  }

  // 2. Strict SemVer check: (Major).(Minor)(.(Patch))?(-prerelease)?(+build)?
  // SemVer specification requires numeric identifiers not to have leading zeroes (e.g. '01' is invalid)
  const semverRegex = /^(0|[1-9]\d*)\.(0|[1-9]\d*)(?:\.(0|[1-9]\d*))?(?:-([0-9A-Za-z.-]+))?(?:\+([0-9A-Za-z.-]+))?$/;
  const match = val.match(semverRegex);
  if (!match) {
    return {
      valid: false,
      error: t("runtimes.version_invalid_format") || "版本号格式不合规（须符合语义化版本规范，例如 1.23.0）",
    };
  }

  const major = parseInt(match[1], 10);
  const minor = parseInt(match[2], 10);
  const patch = match[3] !== undefined ? parseInt(match[3], 10) : undefined;

  // 3. Tool-specific runtime version validations
  const tid = toolId.toLowerCase();
  if (tid === "go" || tid === "golang") {
    // Go public releases are >= 1.0 (e.g. 1.24.0, 1.23.6)
    if (major < 1) {
      return {
        valid: false,
        error: `“${val}” 不是有效的 Go 版本号（Go 官方版本均在 1.0 及以上，如 1.24.0，不存在 0.x 版本）`,
      };
    }
    if (major > 30) {
      return { valid: false, error: `Go 主版本号超出合理范围（v${major}）` };
    }
  } else if (tid === "rust") {
    // Rust stable toolchains are >= 1.0
    if (major < 1) {
      return {
        valid: false,
        error: `“${val}” 不是有效的 Rust 版本号（Rust 官方稳定版本均在 1.0 及以上，如 1.85.0）`,
      };
    }
  } else if (tid === "python") {
    // Python modern versions are 3.x (or 2.7)
    if (major < 2 || major > 4) {
      return {
        valid: false,
        error: `“${val}” 不是有效的 Python 版本号（Python 版本通常为 3.x，如 3.12.0）`,
      };
    }
  } else if (tid === "node") {
    // Node.js releases: 0.10.x, 0.12.x or modern >= 4.0.0
    if (major === 0) {
      if (minor !== 10 && minor !== 12) {
        return {
          valid: false,
          error: `“${val}” 不是有效的 Node.js 版本（Node 0.x 仅支持 0.10 或 0.12，现代版本为 18.x、20.x 等）`,
        };
      }
    } else if (major < 4 || major > 50) {
      return { valid: false, error: `Node.js 版本超出合理范围` };
    }
  }

  return { valid: true, isCompleteSemver: patch !== undefined };
}
