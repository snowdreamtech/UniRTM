#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Bun Logic Module

# Purpose: Installs Bun runtime via unirtm.
# Delegate: Managed by unirtm (.unirtm.toml)
install_runtime_bun() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install Bun runtime."
    return 0
  fi
  "${_G_UNIRTM_BIN:-unirtm}" install bun
}

# Purpose: Sets up Bun runtime.
setup_bun() {
  if ! has_lang_files "bun.lockb" ""; then
    return 0
  fi

  # Dynamically register Bun in .unirtm.toml if not already present.
  setup_registry_bun

  local _T0_BUN_RT
  _T0_BUN_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version bun)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "bun")

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "Bun" "✅ Detected" "${_CUR_VER:-}" "0"
    return 0
  fi

  _log_setup "Bun Runtime" "bun"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "Runtime" "Bun" "⚖️ Previewed" "-" "0"
    return 0
  fi

  local _STAT_BUN_RT="✅ Installed"
  install_runtime_bun || _STAT_BUN_RT="❌ Failed"

  local _DUR_BUN_RT
  _DUR_BUN_RT=$(($(date +%s) - _T0_BUN_RT))
  log_summary "Runtime" "Bun" "${_STAT_BUN_RT:-}" "$(get_version bun --version)" "${_DUR_BUN_RT:-}"
}
# Purpose: Checks if Bun runtime is available.
# Examples:
#   check_runtime_bun "Linter"
check_runtime_bun() {
  local _TOOL_DESC_BUN="${1:-Bun}"
  if ! resolve_bin "bun" >/dev/null 2>&1; then
    log_warn "Required runtime 'bun' for ${_TOOL_DESC_BUN:-} is missing. Skipping."
    return 1
  fi
  return 0
}
