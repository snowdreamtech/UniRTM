#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Haskell Logic Module

# Purpose: Installs Haskell (GHC) runtime via unirtm.
# Delegate: Managed by unirtm (.unirtm.toml)
install_runtime_haskell() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install Haskell runtime."
    return 0
  fi
  # shellcheck disable=SC2154
  "${_G_UNIRTM_BIN:-unirtm}" install "ghc@$(get_unirtm_tool_version ghc)"
}

# Purpose: Installs Haskell linter (ormolu).
install_haskell_lint() {
  if ! has_lang_files "package.yaml stack.yaml *.cabal" "*.hs"; then
    return 0
  fi

  setup_registry_ormolu
  install_tool_safe "ormolu" "ormolu" "Haskell Lint" "--version" 0 "*.hs" ""
}

# Purpose: Sets up Haskell runtime.
setup_haskell() {
  if ! has_lang_files "package.yaml stack.yaml *.cabal" "*.hs"; then
    return 0
  fi

  setup_registry_ghc

  local _T0_HASKELL_RT
  _T0_HASKELL_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version haskell)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "haskell")

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "Haskell" "✅ Detected" "${_CUR_VER:-}" "0"
    return 0
  fi

  _log_setup "Haskell Runtime" "haskell"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "Runtime" "Haskell" "⚖️ Previewed" "-" "0"
    return 0
  fi

  local _STAT_HASKELL_RT="✅ Installed"
  install_runtime_haskell || _STAT_HASKELL_RT="❌ Failed"

  local _DUR_HASKELL_RT
  _DUR_HASKELL_RT=$(($(date +%s) - _T0_HASKELL_RT))
  log_summary "Runtime" "Haskell" "${_STAT_HASKELL_RT:-}" "$(get_version ghc --version)" "${_DUR_HASKELL_RT:-}"

  install_haskell_lint
}
# Purpose: Checks if Haskell runtime is available.
# Examples:
#   check_runtime_haskell "Linter"
check_runtime_haskell() {
  local _TOOL_DESC_GHC="${1:-Haskell}"
  if ! resolve_bin "ghc" >/dev/null 2>&1; then
    log_warn "Required runtime 'ghc' for ${_TOOL_DESC_GHC:-} is missing. Skipping."
    return 1
  fi
  return 0
}
