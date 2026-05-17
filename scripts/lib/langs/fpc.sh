#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Free Pascal (FPC) Logic Module

# Purpose: Installs Free Pascal Compiler (fpc) via unirtm.
# Delegate: Managed by unirtm (.unirtm.toml)
install_runtime_fpc() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install Free Pascal Compiler via unirtm."
    return 0
  fi

  # shellcheck disable=SC2154
  unirtm install "fpc@$(get_unirtm_tool_version fpc)"
}

# Purpose: Sets up FPC environment for project.
setup_fpc() {
  if ! has_lang_files "" "*.pas *.pp *.lpr"; then
    return 0
  fi

  setup_registry_fpc

  local _T0_FPC_RT
  _T0_FPC_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version fpc)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "fpc")

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "Free Pascal" "✅ Detected" "${_CUR_VER:-}" "0"
    return 0
  fi

  _log_setup "Free Pascal" "fpc"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "Runtime" "Free Pascal" "⚖️ Previewed" "-" "0"
    return 0
  fi

  local _STAT_FPC_RT="✅ Installed"
  install_runtime_fpc || _STAT_FPC_RT="❌ Failed"

  local _DUR_FPC_RT
  _DUR_FPC_RT=$(($(date +%s) - _T0_FPC_RT))
  log_summary "Runtime" "Free Pascal" "${_STAT_FPC_RT:-}" "$(get_version fpc -iV | head -n 1)" "${_DUR_FPC_RT:-}"
}

# Purpose: Checks if FPC is available.
# Examples:
#   check_runtime_fpc "Linter"
check_runtime_fpc() {
  local _TOOL_DESC_FPC="${1:-Free Pascal}"
  if ! resolve_bin "fpc" >/dev/null 2>&1; then
    log_warn "Required runtime 'fpc' for ${_TOOL_DESC_FPC:-} is missing. Skipping."
    return 1
  fi
  return 0
}
