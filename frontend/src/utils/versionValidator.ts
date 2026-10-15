// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

import { t } from "../i18n";

export type VersionFormatKind = "semver" | "calver" | "channel_date" | "channel_alias";

export interface VersionValidationResult {
  valid: boolean;
  kind?: VersionFormatKind;
  error?: string;
  isCompleteSemver?: boolean;
}

/**
 * Validates whether an input string is a legitimate semantic version (SemVer),
 * calendar version (CalVer, e.g. 2024.03.15, 2024.1, 24.04), date-based toolchain
 * (e.g. nightly-2024-03-01), or release channel alias (e.g. nightly, lts).
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

  const tid = toolId.toLowerCase();

  // 2. Check for toolchain channel aliases (e.g. "nightly", "beta", "stable", "latest", "lts")
  if (/^(nightly|beta|stable|latest|lts)$/i.test(val)) {
    return { valid: true, kind: "channel_alias" };
  }

  // 3. Check for toolchain date channels (e.g. Rust "nightly-2024-03-01", "beta-2024-03-01")
  const channelDateRegex = /^(nightly|beta|stable)-\d{4}-(0?[1-9]|1[0-2])-(0?[1-9]|[12]\d|3[01])$/i;
  if (channelDateRegex.test(val)) {
    return { valid: true, kind: "channel_date" };
  }

  // 4. Check for ISO Date format (e.g. "2024-03-15", "2024-01-01")
  const isoDateRegex = /^\d{4}-(0?[1-9]|1[0-2])-(0?[1-9]|[12]\d|3[01])$/;
  if (isoDateRegex.test(val)) {
    return { valid: true, kind: "calver" };
  }

  // 5. Check for CalVer (Calendar Versioning):
  // 4-digit Year (1990-2099): 2024.03.15, 2024.1, 2024.05.1
  // 2-digit Year (10-99): 24.04, 24.04.1
  const calverRegex = /^(19\d\d|20\d\d|[1-9]\d)\.(0?[1-9]|1[0-2])(?:\.(0?[1-9]|[12]\d|3[01]|\d+))?(?:[._-][0-9A-Za-z.-]+)?$/;
  const calverMatch = val.match(calverRegex);
  if (calverMatch) {
    const year = parseInt(calverMatch[1], 10);
    if (year >= 1990 || (year >= 10 && year <= 99)) {
      return { valid: true, kind: "calver" };
    }
  }

  // 6. Strict SemVer check: (Major).(Minor)(.(Patch))?(-prerelease)?(+build)?
  // SemVer specification requires numeric identifiers not to have leading zeroes (e.g. '01' is invalid)
  const semverRegex = /^(0|[1-9]\d*)\.(0|[1-9]\d*)(?:\.(0|[1-9]\d*))?(?:-([0-9A-Za-z.-]+))?(?:\+([0-9A-Za-z.-]+))?$/;
  const match = val.match(semverRegex);
  if (!match) {
    return {
      valid: false,
      error: t("runtimes.version_invalid_format") || "版本号格式不合规（须符合语义化版本或年月日 CalVer 规范）",
    };
  }

  const major = parseInt(match[1], 10);
  const minor = parseInt(match[2], 10);
  const patch = match[3] !== undefined ? parseInt(match[3], 10) : undefined;

  // 7. Tool-specific runtime version validations
  if (tid === "go" || tid === "golang") {
    // Go public releases are >= 1.0 (e.g. 1.24.0, 1.23.6)
    if (major < 1) {
      return {
        valid: false,
        error: `版本号无效：Go 官方版本均为 1.x（如 1.24.0）`,
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
        error: `版本号无效：Rust 稳定版本均为 1.x（如 1.85.0）`,
      };
    }
  } else if (tid === "python") {
    // Python modern versions are 3.x (or 2.7)
    if (major < 2 || major > 4) {
      return {
        valid: false,
        error: `版本号无效：Python 版本通常为 3.x（如 3.12.0）`,
      };
    }
  } else if (tid === "node") {
    // Node.js releases: 0.10.x, 0.12.x or modern >= 4.0.0
    if (major === 0) {
      if (minor !== 10 && minor !== 12) {
        return {
          valid: false,
          error: `版本号无效：Node.js 现代版本为 18.x、20.x 等`,
        };
      }
    } else if (major < 4 || major > 50) {
      return { valid: false, error: `Node.js 版本超出合理范围` };
    }
  }

  return { valid: true, kind: "semver", isCompleteSemver: patch !== undefined };
}
