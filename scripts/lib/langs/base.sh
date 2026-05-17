#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Base Logic Module

# Purpose: Installs pipx.
# Delegate: Managed by unirtm (.unirtm.toml)
install_pipx() {
  local _T0_PIPX
  _T0_PIPX=$(date +%s)
  local _TITLE="Pipx"
  local _PROVIDER="${VER_PIPX_PROVIDER:-pip:pipx}"
  if resolve_bin "pipx" >/dev/null 2>&1; then
    log_summary "Base" "Pipx" "✅ Exists" "$(get_version pipx)" "$(($(date +%s) - _T0_PIPX))"
    return 0
  fi

  _log_setup "${_TITLE:-}" "${_PROVIDER:-}"

  # Proactive pipx installation via pip (Universal fallback)
  # This ensures pipx is available regardless of unirtm's provider compatibility.
  if ! resolve_bin "pipx" >/dev/null 2>&1; then
    log_info "Ensuring pipx is available via pip..."
    # Use python -m pip to ensure we use the correct python instance
    # Note: --user may fail inside a virtualenv, so we try with it first and fallback if needed.
    local _V_PIPX="${VER_PIPX:-latest}"
    if ! "${PYTHON:-}" -m pip install --user "pipx==${_V_PIPX:-}" --quiet 2>/dev/null; then
      log_debug "Pipx: --user install failed or not available, trying standard install..."
      "${PYTHON:-}" -m pip install "pipx==${_V_PIPX:-}" --quiet || true
    fi

    # Add user scripts to PATH immediately if not present
    local _USER_BIN
    local _PY_PREFIX
    _PY_PREFIX=$("${PYTHON:-}" -c "import sys; print(sys.prefix)" | tr -d '\r')
    if [ "${_G_OS:-}" = "windows" ]; then
      [ -d "${_PY_PREFIX:-}/Scripts" ] && _USER_BIN="${_PY_PREFIX:-}/Scripts"
    else
      [ -d "${_PY_PREFIX:-}/bin" ] && _USER_BIN="${_PY_PREFIX:-}/bin"
    fi

    # Fallback to User Base (for --user installs) if not found in prefix
    if [ -z "${_USER_BIN:-}" ]; then
      if [ "${_G_OS:-}" = "windows" ]; then
        local _USER_BASE
        _USER_BASE=$("${PYTHON:-}" -m site --user-base 2>/dev/null | tr -d '\r')
        [ -n "${_USER_BASE:-}" ] && _USER_BIN="${_USER_BASE:-}/Scripts"
      else
        # On macOS/Linux, pipx usually installs to ~/.local/bin or similar
        # shellcheck disable=SC2155
        _USER_BIN=$(python3 -m site --user-base 2>/dev/null)/bin
      fi
    fi

    if [ -n "${_USER_BIN:-}" ]; then
      case ":$PATH:" in
      *":${_USER_BIN:-}:"*) ;;
      *) export PATH="${_USER_BIN:-}:$PATH" ;;
      esac
    fi

  fi

  # Note: unirtm managed installation removed to avoid aqua backend non-Windows compatibility issues.
  # pipx is correctly available via the pip-based bypass above.
  local _STAT_PIPX="✅ pip"
  log_summary "Base" "Pipx" "${_STAT_PIPX:-}" "$(get_version pipx)" "$(($(date +%s) - _T0_PIPX))"
}

# Purpose: Installs Gitleaks for secrets scanning.
# Delegate: Managed by unirtm (.unirtm.toml)
install_gitleaks() {
  # In CI, always install. Locally, skip if not a git repository
  local _SKIP_CHECK=1
  if [ ! -d ".git" ] && ! is_ci_env; then
    log_info "⏭️  Gitleaks: Skipped (not a git repository)"
    log_summary "Base" "Gitleaks" "⏭️ Skipped" "-" "0"
    return 0
  fi
  install_tool_safe "gitleaks" "${VER_GITLEAKS_PROVIDER:-}" "Gitleaks" "version" 1
}

# Purpose: Installs checkmake for Makefile linting.
# Delegate: Managed by unirtm (.unirtm.toml)
install_checkmake() {
  install_tool_safe "checkmake" "${VER_CHECKMAKE_PROVIDER:-}" "Checkmake" "--version" 0 "Makefile *.make" ""
}

# Purpose: Installs pre-commit runtime via pipx.
install_runtime_hooks() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install pre-commit via pipx."
    return 0
  fi
  unirtm install pipx:pre-commit
}

# Purpose: Activates git pre-commit hooks.
setup_hooks() {
  local _T0_HOOK
  _T0_HOOK=$(date +%s)
  # 2. Fast-path: Check if hooks already exist
  if [ -f ".git/hooks/pre-commit" ]; then
    log_summary "Base" "Hooks" "✅ Activated" "4.5.1" "0"
    return 0
  fi

  # 3. Action Required (Real or Preview)
  _log_setup "Pre-commit Hooks" "pipx:pre-commit"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "Base" "Hooks" '⚖️ Previewed' "-" '0'
    return 0
  fi
  local _STAT_HOOK="✅ Activated"
  install_runtime_hooks || _STAT_HOOK="❌ Failed"

  local _DUR_HOOK
  _DUR_HOOK=$(($(date +%s) - _T0_HOOK))
  log_summary "Base" "Hooks" "${_STAT_HOOK:-}" "$(get_version pre-commit --version)" "${_DUR_HOOK:-}"
}

# Purpose: Installs editorconfig-checker.
# Delegate: Managed by unirtm (.unirtm.toml)
install_editorconfig_checker() {
  # Skip if no .editorconfig file
  if [ ! -f ".editorconfig" ]; then
    return 0
  fi

  # Note: editorconfig-checker has platform-specific binary names (ec-linux-amd64, ec-darwin-amd64, etc.)
  # The actual binary name is resolved automatically by install_tool_safe()
  install_tool_safe "ec" "${VER_EDITORCONFIG_CHECKER_PROVIDER:-}" "Editorconfig-Checker" "--version" 1
}

# Purpose: Installs GoReleaser as a universal release automation tool.
# Note: goreleaser supports multi-language projects (Go, Rust, Python, Node, etc.)
#       It is installed globally regardless of project language.
install_goreleaser() {
  # CI-only: Optional release tool, local dev skips to avoid slow GitHub download.
  if ! is_ci_env && [ "${GORELEASER_FORCE_INSTALL:-0}" -ne 1 ]; then
    return 0
  fi
  install_tool_safe "goreleaser" "${VER_GORELEASER_PROVIDER:-}" "GoReleaser" "--version" 1
}

# Purpose: Sets up Base environment.
setup_base() {
  install_pipx
  install_gitleaks
  install_checkmake
  setup_hooks
  install_editorconfig_checker
  install_goreleaser
}
