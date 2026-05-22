#!/usr/bin/env node
// install.js - Root package entry point for @snowdreamtech/unirtm
//
// This script is both the `bin` entry and the `postinstall` hook.
// It locates the correct platform-specific binary and either:
//   - Runs it (when invoked as CLI)
//   - Validates the installation (when invoked via postinstall)

"use strict";

const { execFileSync, spawnSync } = require("child_process");
const path = require("path");
const fs = require("fs");
const os = require("os");

// ---------------------------------------------------------------------------
// Platform detection
// ---------------------------------------------------------------------------

/**
 * Map Node.js process.platform + process.arch to npm package suffix.
 * @returns {string} package suffix e.g. "linux-x64"
 */
function getPlatformPackageSuffix() {
  const platform = process.platform; // 'darwin' | 'linux' | 'win32'
  const arch = process.arch; // 'x64' | 'arm64' | 'arm' | 'ia32'

  // ARM sub-arch detection (armv5/armv6/armv7)
  // process.config.variables.arm_version is available in some Node builds
  const armVersion = (process.config && process.config.variables && process.config.variables.arm_version) || "";

  if (platform === "darwin") {
    if (arch === "arm64") return "darwin-arm64";
    if (arch === "x64") return "darwin-x64";
  }

  if (platform === "linux") {
    if (arch === "x64") return "linux-x64";
    if (arch === "arm64") return "linux-arm64";
    if (arch === "ia32") return "linux-ia32";
    if (arch === "arm") {
      if (armVersion === "5") return "linux-arm-5";
      if (armVersion === "6") return "linux-arm-6";
      return "linux-arm"; // armv7 default
    }
    if (arch === "loong64") return "linux-loong64";
    if (arch === "ppc64") return "linux-ppc64le";
    if (arch === "riscv64") return "linux-riscv64";
    if (arch === "s390x") return "linux-s390x";
  }

  if (platform === "win32") {
    if (arch === "x64") return "windows-x64";
    if (arch === "arm64") return "windows-arm64";
    if (arch === "ia32") return "windows-ia32";
  }

  throw new Error(
    `Unsupported platform: ${platform}-${arch}\n` +
      "Please open an issue at https://github.com/snowdreamtech/unirtm/issues"
  );
}

// ---------------------------------------------------------------------------
// Binary resolution
// ---------------------------------------------------------------------------

/**
 * Locate the platform binary from the optional dependency package.
 * @param {string} suffix  e.g. "linux-x64"
 * @returns {string} absolute path to the binary
 */
function resolveBinary(suffix) {
  const pkgName = `@snowdreamtech/unirtm-${suffix}`;
  const binaryName = process.platform === "win32" ? "unirtm.exe" : "unirtm";

  // Strategy 1: resolve via require.resolve (works when package is installed)
  try {
    const pkgJsonPath = require.resolve(`${pkgName}/package.json`);
    const pkgDir = path.dirname(pkgJsonPath);
    const binPath = path.join(pkgDir, "bin", binaryName);
    if (fs.existsSync(binPath)) {
      return binPath;
    }
  } catch (_) {
    // package not found via require.resolve, try manual search
  }

  // Strategy 2: walk up node_modules hierarchy
  let dir = __dirname;
  for (let i = 0; i < 10; i++) {
    const candidate = path.join(dir, "node_modules", pkgName, "bin", binaryName);
    if (fs.existsSync(candidate)) {
      return candidate;
    }
    const parent = path.dirname(dir);
    if (parent === dir) break;
    dir = parent;
  }

  throw new Error(
    `Could not find binary for ${pkgName}.\n` +
      "Try reinstalling @snowdreamtech/unirtm:\n" +
      "  npm install @snowdreamtech/unirtm\n\n" +
      `Platform: ${process.platform}-${process.arch}`
  );
}

// ---------------------------------------------------------------------------
// Entry point
// ---------------------------------------------------------------------------

const isPostInstall = process.env.npm_lifecycle_event === "postinstall";

if (isPostInstall) {
  // Postinstall: just validate that the binary can be found
  try {
    const suffix = getPlatformPackageSuffix();
    const binPath = resolveBinary(suffix);
    const result = spawnSync(binPath, ["--version"], { encoding: "utf8" });
    if (result.status === 0) {
      process.stdout.write(`✅ unirtm installed successfully (${binPath})\n`);
    }
  } catch (err) {
    // Non-fatal: postinstall failure should not break npm install
    process.stderr.write(`⚠️  unirtm: ${err.message}\n`);
  }
} else {
  // CLI invocation: forward all arguments to the native binary
  try {
    const suffix = getPlatformPackageSuffix();
    const binPath = resolveBinary(suffix);
    const args = process.argv.slice(2);

    const result = spawnSync(binPath, args, {
      stdio: "inherit",
      windowsHide: false,
    });

    process.exit(result.status !== null ? result.status : 1);
  } catch (err) {
    process.stderr.write(`Error: ${err.message}\n`);
    process.exit(1);
  }
}
