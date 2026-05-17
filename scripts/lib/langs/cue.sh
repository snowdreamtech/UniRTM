#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# CUE Logic Module

# Purpose: Installs CUE via unirtm.
# Delegate: Managed by unirtm (.unirtm.toml)
install_runtime_cue() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install CUE via unirtm."
    return 0
  fi

  # shellcheck disable=SC2154
  "${_G_UNIRTM_BIN:-unirtm}" install "cue@$(get_unirtm_tool_version cue)"
}

# Purpose: Sets up CUE environment for project.
# Delegate: Managed by unirtm (.unirtm.toml)
# Examples:
#   setup_cue
setup_cue() {
  if ! has_lang_files "" "*.cue"; then
    return 0
  fi

  local _T0_CUE_RT
  _T0_CUE_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version cue)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "cue")

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "CUE" "✅ Detected" "${_CUR_VER:-}" "0"
    return 0
  fi

  _log_setup "CUE" "cue"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "Runtime" "CUE" "⚖️ Previewed" "-" "0"
    return 0
  fi

  local _STAT_CUE_RT="✅ Installed"
  install_runtime_cue || _STAT_CUE_RT="❌ Failed"

  local _DUR_CUE_RT
  _DUR_CUE_RT=$(($(date +%s) - _T0_CUE_RT))
  log_summary "Runtime" "CUE" "${_STAT_CUE_RT:-}" "$(get_version cue version | head -n 1)" "${_DUR_CUE_RT:-}"
}

# Purpose: Checks if CUE is available.
# Examples:
#   check_runtime_cue "Linter"
check_runtime_cue() {
  local _TOOL_DESC_CUE="${1:-CUE}"
  if ! resolve_bin "cue" >/dev/null 2>&1; then
    log_warn "Required runtime 'cue' for ${_TOOL_DESC_CUE:-} is missing. Skipping."
    return 1
  fi
  return 0
}
