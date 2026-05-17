#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Move Logic Module

# Purpose: Installs Move toolchain (via aptos CLI) via unirtm.
# Delegate: Managed by unirtm (.unirtm.toml)
install_runtime_move() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install Aptos CLI (Move toolchain) via mise."
    return 0
  fi

  # shellcheck disable=SC2154
  unirtm install "${VER_APTOS_PROVIDER:-}@${VER_APTOS:-}"
}

# Purpose: Sets up Move environment for project.
setup_move() {
  if ! has_lang_files "Move.toml" "*.move"; then
    return 0
  fi

  setup_registry_move

  local _T0_MOVE_RT
  _T0_MOVE_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version aptos)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "aptos")

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "Move" "✅ Detected" "${_CUR_VER:-}" "0"
    return 0
  fi

  _log_setup "Move" "aptos"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "Runtime" "Move" "⚖️ Previewed" "-" "0"
    return 0
  fi

  local _STAT_MOVE_RT="✅ Installed"
  install_runtime_move || _STAT_MOVE_RT="❌ Failed"

  local _DUR_MOVE_RT
  _DUR_MOVE_RT=$(($(date +%s) - _T0_MOVE_RT))
  log_summary "Runtime" "Move" "${_STAT_MOVE_RT:-}" "$(get_version aptos --version | awk '{print $NF}')" "${_DUR_MOVE_RT:-}"
}

# Purpose: Checks if Move (aptos) is available.
# Examples:
#   check_runtime_move "Linter"
check_runtime_move() {
  local _TOOL_DESC_MOVE="${1:-Move}"
  if ! resolve_bin "aptos" >/dev/null 2>&1; then
    log_warn "Required runtime 'aptos' for ${_TOOL_DESC_MOVE:-} is missing. Skipping."
    return 1
  fi
  return 0
}
