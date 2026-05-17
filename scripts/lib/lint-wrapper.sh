#!/usr/bin/env sh
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# scripts/lib/lint-wrapper.sh - Robust wrapper for pre-commit hooks.
#
# Purpose:
#   Ensures that optional linters skip gracefully if specialized tools
#   or runtimes are missing, maintaining cross-platform integrity.
#
# Usage:
#   sh scripts/lib/lint-wrapper.sh LINTER_NAME [ARGS...]
#
# Standards:
#   - POSIX-compliant sh logic.
#   - "World Class" AI Documentation (English-only).
#   - Rule 01 (General), Rule 02 (Coding Style).
#
# Features:
#   - Dynamic binary resolution (.venv, node_modules, PATH).
#   - Native runtime detection (Java, Ruby, Node, Dart, DOTNET).
#   - OS-specific guards (Apple Swift).
set -eu

# ── Common Library ───────────────────────────────────────────────────────────
SCRIPT_DIR=$(cd "$(dirname "${0:-}")" && pwd)
. "${SCRIPT_DIR:-}/common.sh"

# ── Functions ────────────────────────────────────────────────────────────────

# Purpose: Main entry point for the linter delegation engine.
#          Resolves the linter binary path and performs language-specific checks.
# Params:
#   $1 - Linter binary/hook name (e.g., "eslint", "gofmt")
#   $@ - Arguments passed to the linter
# Examples:
#   main eslint --fix path/to/file.js
main() {
  local _LINTER_WRAP="${1:-}"
  [ -z "${_LINTER_WRAP:-}" ] && return 0
  shift

  # Early Return: Optional Security Tools
  # These heavy scanners are NOT registered in .unirtm.toml and have dedicated
  # coverage in CI audit stages (trivy-action, CodeQL, SARIF uploads).
  # They should skip gracefully in ALL environments (CI and local) when not
  # installed, instead of hard-failing the lint pipeline.
  case "${_LINTER_WRAP:-}" in
  trivy | checkov | semgrep | bandit | zizmor)
    log_warn "⏭️  ${_LINTER_WRAP:-} not available. Skipping (covered by CI audit stage)."
    exit 0
    ;;
  esac

  # 1. Resolve Binary Path
  local _LINTER_BIN="${_LINTER_WRAP:-}"
  local _MISE_TOOL_SPEC=""
  case "${_LINTER_WRAP:-}" in
  psscriptanalyzer) _LINTER_BIN="pwsh" ;;
  osv_scanner) _LINTER_BIN="osv-scanner" ;;
  node-audit)
    # node-audit is a logical tool, we resolve the package manager instead
    _LINTER_BIN="${NPM:-pnpm}"
    ;;
  # Map tool names to unirtm tool specs for tools with different binary names
  # SECURITY: Explicitly specify full tool specs to avoid unirtm's default registry
  # which may redirect to different backends (e.g., aqua:mrtazz/checkmake instead
  # of github:checkmake/checkmake). This prevents supply chain attacks via
  # implicit registry redirections.
  checkmake)
    _MISE_TOOL_SPEC="github:checkmake/checkmake"
    _LINTER_BIN="checkmake"
    ;;
  shfmt)
    _MISE_TOOL_SPEC="github:mvdan/sh"
    _LINTER_BIN="shfmt"
    ;;
  taplo)
    _MISE_TOOL_SPEC="npm:@taplo/cli"
    _LINTER_BIN="taplo"
    ;;
  editorconfig-checker)
    _MISE_TOOL_SPEC="github:editorconfig-checker/editorconfig-checker"
    # Binary name is 'ec' with platform-specific suffixes
    # The unirtm tool spec uses bin = "ec-*" to match all variants
    _LINTER_BIN="ec"
    ;;
  esac

  local _RESOLVED_BIN_WRAP
  _RESOLVED_BIN_WRAP=$(resolve_bin "${_LINTER_BIN:-}") || true

  # 2. Check Existence
  if [ -z "${_RESOLVED_BIN_WRAP:-}" ]; then
    # Dynamic Handler: On-demand Tier 2 tools (not registered in .unirtm.toml)
    if [ "${_LINTER_WRAP:-}" = "zizmor" ]; then
      local _ZM_SPEC="${VER_ZIZMOR_PROVIDER:-zizmor}@${VER_ZIZMOR:-latest}"
      log_info "── Executing ${_LINTER_WRAP:-} (dynamic) ──"
      # Execute directly with unirtm exec
      # shellcheck disable=SC2093
      exec mise exec "${_ZM_SPEC:-}" -- zizmor "$@"
    fi

    # CI Fallback: Try unirtm exec directly if tool not resolved
    # This handles cases where unirtm shims exist but resolve_bin fails
    if is_ci_env; then
      # Special handling for node-audit: it's a logical tool, not a real binary
      if [ "${_LINTER_WRAP:-}" = "node-audit" ]; then
        log_info "=== CI Fallback for node-audit ==="
        local _PKG_MGR="${NPM:-npm}"
        # local _PKG_MGR="${NPM:-pnpm}"
        log_info "Using package manager: ${_PKG_MGR:-}"

        # Try to execute audit directly
        if command -v "${_PKG_MGR:-}" >/dev/null 2>&1; then
          log_info "✓ Package manager found, executing audit..."
          "${_PKG_MGR:-}" audit --registry=https://registry.npmjs.org "$@"
          exit $?
        else
          log_error "Package manager '${_PKG_MGR:-}' not found"
          exit 1
        fi
      fi

      log_info "=== CI Tool Resolution Fallback for ${_LINTER_WRAP:-} ==="
      log_info "Tool spec: ${_MISE_TOOL_SPEC:-${_LINTER_BIN:-}}"
      log_info "Binary name: ${_LINTER_BIN:-}"

      # Use tool spec if available, otherwise use binary name
      local _EXEC_TARGET="${_MISE_TOOL_SPEC:-${_LINTER_BIN:-}}"

      # Step 1: Try to execute the tool first
      log_debug "Step 1: Attempting mise exec..."
      if mise exec "${_EXEC_TARGET:-}" -- "${_LINTER_BIN:-}" --version >/dev/null 2>&1; then
        log_info "✓ Tool found via unirtm exec, executing..."
        # shellcheck disable=SC2093
        exec mise exec "${_EXEC_TARGET:-}" -- "${_LINTER_BIN:-}" "$@"
      fi
      log_warn "✗ unirtm exec failed"

      # Step 2: Check if tool is installed in unirtm
      log_debug "Step 2: Checking mise installation status..."
      if mise list 2>/dev/null | grep -q "${_EXEC_TARGET:-}"; then
        log_info "Tool is registered in unirtm, attempting uninstall..."
        mise uninstall "${_EXEC_TARGET:-}" 2>/dev/null || true
      else
        log_info "Tool not found in unirtm registry"
      fi

      # Step 3: Install the tool
      log_info "Step 3: Installing tool..."
      if mise install "${_EXEC_TARGET:-}"; then
        log_info "✓ Installation successful"

        # Step 4: Refresh unirtm state
        log_debug "Step 4: Refreshing mise state..."
        unirtm reshim 2>/dev/null || log_warn "reshim failed"
        sleep 1

        # Step 5: Try unirtm exec again
        log_debug "Step 5: Retrying mise exec..."
        if mise exec "${_EXEC_TARGET:-}" -- "${_LINTER_BIN:-}" --version >/dev/null 2>&1; then
          log_info "✓ Tool now executable via unirtm exec"
          # shellcheck disable=SC2093
          exec mise exec "${_EXEC_TARGET:-}" -- "${_LINTER_BIN:-}" "$@"
        fi
        log_warn "✗ unirtm exec still failed after installation"

        # Step 6: Try direct execution from install path
        log_debug "Step 6: Attempting direct execution..."
        local _INSTALL_PATH
        _INSTALL_PATH=$(mise where "${_EXEC_TARGET:-}" 2>/dev/null || true)
        if [ -n "${_INSTALL_PATH:-}" ]; then
          log_info "Install path: ${_INSTALL_PATH:-}"

          # Try multiple possible binary locations
          for _BIN_DIR in "bin" "."; do
            # First try exact match
            local _FULL_PATH="${_INSTALL_PATH:-}/${_BIN_DIR}/${_LINTER_BIN:-}"
            if [ -f "${_FULL_PATH:-}" ]; then
              # On Windows, skip executable check (files are executable by extension)
              if [ "${_G_OS:-}" = "windows" ] || [ -x "${_FULL_PATH:-}" ]; then
                log_info "✓ Found executable at ${_FULL_PATH:-}"
                exec "${_FULL_PATH:-}" "$@"
              fi
            fi

            # Then try pattern match (for binaries with version/platform suffixes)
            local _PATTERN_PATH
            # Find all matching files first (without permission check)
            local _CANDIDATES
            _CANDIDATES=$(find "${_INSTALL_PATH:-}/${_BIN_DIR}" -maxdepth 1 -name "${_LINTER_BIN:-}*" -type f 2>/dev/null || true)

            # Filter for executable files
            if [ -n "${_CANDIDATES:-}" ]; then
              for _CANDIDATE in ${_CANDIDATES}; do
                # On Windows, skip executable check (files are executable by extension)
                if [ "${_G_OS:-}" = "windows" ] || [ -x "${_CANDIDATE:-}" ]; then
                  _PATTERN_PATH="${_CANDIDATE}"
                  break
                fi
              done
            fi

            if [ -n "${_PATTERN_PATH:-}" ] && [ -f "${_PATTERN_PATH:-}" ]; then
              # On Windows, skip executable check
              if [ "${_G_OS:-}" = "windows" ] || [ -x "${_PATTERN_PATH:-}" ]; then
                log_info "✓ Found executable via pattern match at ${_PATTERN_PATH:-}"
                exec "${_PATTERN_PATH:-}" "$@"
              fi
            fi
          done
          log_warn "✗ Binary not found in install path"
        else
          log_warn "✗ Could not determine install path"
        fi
      else
        log_error "✗ Installation failed"
      fi

      # All attempts failed
      log_error "❌ ${_LINTER_WRAP:-} not found in CI after all attempts"
      log_info "💡 Debugging information:"
      log_info "   - Tool spec: ${_EXEC_TARGET:-}"
      log_info "   - Binary: ${_LINTER_BIN:-}"
      log_info "   - unirtm list output:"
      mise list 2>&1 | grep -E "(${_EXEC_TARGET:-}|${_LINTER_BIN:-})" || echo "     (no matches)"
      exit 1
    fi
    log_warn "⚠️  ${_LINTER_WRAP:-} not found locally. Skipping linting for this module."
    log_info "💡 Run 'make setup' to install required tools."
    exit 0
  fi

  # 3. Special Runtime Checks (Fail-Fast for missing language foundations)
  case "${_LINTER_WRAP:-}" in
  rubocop) check_runtime ruby "${_LINTER_WRAP:-}" ;;
  dart) check_runtime dart "${_LINTER_WRAP:-}" ;;
  mix) check_runtime elixir "${_LINTER_WRAP:-}" ;;
  scalafmt | google-java-format | ktlint) check_runtime java "${_LINTER_WRAP:-}" ;;
  ormolu)
    if [ "${_G_OS:-}" = "windows" ]; then
      log_info "⏭️  ormolu (Haskell) is skipped on Windows CI due to GHC candidate availability. Skipping."
      exit 0
    fi
    check_runtime haskell "${_LINTER_WRAP:-}"
    ;;
  eslint | prettier | stylelint | spectral | sort-package-json | markdownlint-cli2 | taplo | dockerfile-utils | commitlint | node-audit)
    check_runtime node "${_LINTER_WRAP:-}"
    ;;
  psscriptanalyzer)
    # Binary is pwsh, but check_runtime can use pwsh direct or a module if we had one
    check_runtime pwsh "${_LINTER_WRAP:-}"
    ;;
  swiftformat | swiftlint)
    if [ "$(uname -s)" != "Darwin" ]; then
      log_info "⏭️  ${_LINTER_WRAP:-} is only supported on macOS. Skipping."
      exit 0
    fi
    check_runtime swift "${_LINTER_WRAP:-}"
    ;;
  dotnet) check_runtime dotnet "${_LINTER_WRAP:-}" ;;
  gofmt) check_runtime go "${_LINTER_WRAP:-}" ;;
  cargo) check_runtime rust "${_LINTER_WRAP:-}" ;;
  osv_scanner | trivy | checkov | semgrep)
    # Generic binaries or cross-language tools
    # check_runtime will use resolve_bin to find them
    # For osv_scanner, we mapped it to osv-scanner binary in step 1
    check_runtime "${_LINTER_BIN:-}" "${_LINTER_WRAP:-}"
    ;;
  bandit) check_runtime python "${_LINTER_WRAP:-}" ;;
  gosec | govulncheck) check_runtime go "${_LINTER_WRAP:-}" ;;
  cargo-audit) check_runtime rust "${_LINTER_WRAP:-}" ;;
  *)
    # Generic fallback for other binary tools (yamllint, etc.)
    # If a module exists for the tool name, it will be used.
    check_runtime "${_LINTER_WRAP:-}" "${_LINTER_WRAP:-}"
    ;;
  esac

  # 4. Execute Linter
  # shellcheck disable=SC2086
  if [ "${_LINTER_WRAP:-}" = "node-audit" ]; then
    # logical case: call package manager audit command
    # NOTE: npm has retired the old audit endpoints (/-/npm/v1/security/audits/*).
    # Modern package managers (pnpm 9.0+, npm 7+, yarn 2+) now use the bulk advisory endpoint.
    # However, some mirror registries (like npmmirror.com) don't support audit endpoints at all.
    # We must use the official npm registry for audit operations.

    # On Windows, we need to handle .cmd/.exe wrappers carefully
    # Use the package manager command directly instead of relying on resolved path
    local _PKG_MGR="${NPM:-npm}"
    # local _PKG_MGR="${NPM:-pnpm}"

    # Verify the package manager is available
    if ! command -v "${_PKG_MGR:-}" >/dev/null 2>&1; then
      log_error "Package manager '${_PKG_MGR:-}' not found in PATH"
      exit 1
    fi

    # Execute audit with explicit command invocation (not exec)
    # This ensures proper argument passing on Windows
    # Use official npm registry for audit (required for mirrors that don't support audit)
    "${_PKG_MGR:-}" audit --registry=https://registry.npmjs.org "$@"
    exit $?
  fi

  exec "${_RESOLVED_BIN_WRAP:-}" "$@"
}

main "$@"
