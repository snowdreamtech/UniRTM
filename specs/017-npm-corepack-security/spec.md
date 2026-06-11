# Feature Specification: Npm Security & Corepack Alignment

**Feature Branch**: `[017-npm-corepack-security]`

**Created**: 2026-06-11

**Status**: Draft

**Input**: Align with the security mechanisms and frontend ecosystem standards of modern toolchains like `mise`.

## Overview

Implement atomic development mechanisms across platforms (POSIX/Windows) to achieve UniRTM's core concepts of "Security by Default" and "Ultimate Developer Experience":
1. **Block Zero-Day Supply Chain Attacks**: Intercept high-risk lifecycle scripts during npm package installations, balancing transparency and security.
2. **Zero-Configuration Node Ecosystem Support**: Provide native bridge support for Corepack to enable seamless invocation of `yarn` and `pnpm`.

## Detailed Plan

### 1. Interception and Warning for npm Lifecycle Scripts

#### Security Pain Points
The `preinstall` and `postinstall` lifecycle scripts in the global npm ecosystem are frequent vectors for supply chain poisoning. An atomic execution mechanism with "isolation by default, explicit authorization" must be enforced.

#### Technical Solution
- **Mandatory Isolation**: Inject the `--ignore-scripts=true` parameter during global installations in `internal/provider/npm.go` to enforce indiscriminate interception.
- **Escape Hatch**: Developers can explicitly bypass this security interception and allow lifecycle scripts to execute by setting the `UNIRTM_NPM_ALLOW_SCRIPTS=1` environment variable.
- **Intelligent Diagnostic Warnings**: After installation (atomic directory writing), read the `package.json` located in the respective component directory:
  - UNIX: `lib/node_modules/<tool>/package.json`
  - Windows: `node_modules/<tool>/package.json`
- **Dynamic Alerts**: If `postinstall` or `preinstall` attributes are found in `package.json`, emit a highly visible warning to the terminal via `logger.Warn`:
  `⚠️ WARNING: For security reasons, UniRTM has intercepted the installation scripts for this tool. If the tool fails to run, it may rely on lifecycle scripts. Please manually verify its safety.`

### 2. Native Support for Corepack Package Management

#### Cross-Platform Experience Consistency
Eliminate the need for developers to manually execute `corepack enable`. UniRTM's version switching will automatically support environment-aware execution for `yarn`, `yarnpkg`, `pnpm`, and `pnpx`.

#### Technical Solution
- **Shim Mounting**: Modify `internal/provider/node.go`.
- Add `corepack`, `yarn`, `yarnpkg`, `pnpm`, and `pnpx` to the `ListExecutables` whitelist and `GenerateShims` target generation list.
- **Delegate Execution**: For the generation of specialized shims like `yarn`/`pnpm`:
  - The internal `exePath` will not point to a non-existent `yarn.cmd`, but directly to the native `corepack` executable.
  - The generated shim script will prepend its own name as the first argument during execution: `exec "corepack" "yarn" "$@"` (POSIX) or `corepack.cmd "yarn" %*` (Windows).
