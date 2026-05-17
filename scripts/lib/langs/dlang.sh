#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Dlang Logic Module

# Purpose: Installs Dlang (dmd) via unirtm.
# Delegate: Managed by unirtm (.unirtm.toml)
install_runtime_dlang() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install Dlang (dmd) via unirtm."
    return 0
  fi

  # shellcheck disable=SC2154
  unirtm install "dmd@$(get_unirtm_tool_version dlang)"
}

# Purpose: Sets up Dlang environment for project.
setup_dlang() {
  if ! has_lang_files "dub.json dub.sdl" "*.d *.di"; then
    return 0
  fi

  setup_registry_dlang

  local _T0_D_RT
  _T0_D_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version dmd)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "dmd")

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "Dlang" "✅ Detected" "${_CUR_VER:-}" "0"
    return 0
  fi

  _log_setup "Dlang" "dmd"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "Runtime" "Dlang" "⚖️ Previewed" "-" "0"
    return 0
  fi

  local _STAT_D_RT="✅ Installed"
  install_runtime_dlang || _STAT_D_RT="❌ Failed"

  local _DUR_D_RT
  _DUR_D_RT=$(($(date +%s) - _T0_D_RT))
  log_summary "Runtime" "Dlang" "${_STAT_D_RT:-}" "$(get_version dmd --version | head -n 1 | awk '{print $4}')" "${_DUR_D_RT:-}"
}

# Purpose: Checks if Dlang is available.
# Examples:
#   check_runtime_dlang "Linter"
check_runtime_dlang() {
  local _TOOL_DESC_D="${1:-Dlang}"
  if ! resolve_bin "dmd" >/dev/null 2>&1; then
    log_warn "Required runtime 'dmd' for ${_TOOL_DESC_D:-} is missing. Skipping."
    return 1
  fi
  return 0
}
