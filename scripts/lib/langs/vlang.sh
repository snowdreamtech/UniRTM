#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Vlang Logic Module

# Purpose: Installs V compiler via unirtm.
# Delegate: Managed by unirtm (.unirtm.toml)
install_runtime_vlang() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install Vlang via unirtm."
    return 0
  fi

  # shellcheck disable=SC2154
  unirtm install "vlang@$(get_unirtm_tool_version vlang)"
}

# Purpose: Sets up Vlang environment for project.
setup_vlang() {
  if ! has_lang_files "v.mod" "*.v"; then
    return 0
  fi

  setup_registry_vlang

  local _T0_VLG_RT
  _T0_VLG_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version v)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "v")

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "Vlang" "✅ Detected" "${_CUR_VER:-}" "0"
    return 0
  fi

  _log_setup "Vlang" "v"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "Runtime" "Vlang" "⚖️ Previewed" "-" "0"
    return 0
  fi

  local _STAT_VLG_RT="✅ Installed"
  install_runtime_vlang || _STAT_VLG_RT="❌ Failed"

  local _DUR_VLG_RT
  _DUR_VLG_RT=$(($(date +%s) - _T0_VLG_RT))
  log_summary "Runtime" "Vlang" "${_STAT_VLG_RT:-}" "$(get_version v version | head -n 1)" "${_DUR_VLG_RT:-}"
}

# Purpose: Checks if V compiler is available.
# Examples:
#   check_runtime_vlang "Linter"
check_runtime_vlang() {
  local _TOOL_DESC_VLG="${1:-Vlang}"
  if ! resolve_bin "v" >/dev/null 2>&1; then
    log_warn "Required runtime 'v' for ${_TOOL_DESC_VLG:-} is missing. Skipping."
    return 1
  fi
  return 0
}
