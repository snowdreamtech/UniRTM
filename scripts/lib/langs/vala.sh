#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Vala Logic Module

# Purpose: Installs Vala via unirtm.
# Delegate: Managed by unirtm (.unirtm.toml)
install_runtime_vala() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install Vala via unirtm."
    return 0
  fi

  # shellcheck disable=SC2154
  unirtm install "${VER_VALA_PROVIDER:-}@${VER_VALA:-}"
}

# Purpose: Sets up Vala environment for project.
setup_vala() {
  if ! has_lang_files "" "*.vala *.vapi"; then
    return 0
  fi

  local _T0_VALA_RT
  _T0_VALA_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version valac)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "valac")

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "Vala" "✅ Detected" "${_CUR_VER:-}" "0"
    return 0
  fi

  _log_setup "Vala" "valac"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "Runtime" "Vala" "⚖️ Previewed" "-" "0"
    return 0
  fi

  local _STAT_VALA_RT="✅ Installed"
  install_runtime_vala || _STAT_VALA_RT="❌ Failed"

  local _DUR_VALA_RT
  _DUR_VALA_RT=$(($(date +%s) - _T0_VALA_RT))
  log_summary "Runtime" "Vala" "${_STAT_VALA_RT:-}" "$(get_version valac --version | head -n 1 | awk '{print $NF}')" "${_DUR_VALA_RT:-}"
}

# Purpose: Checks if Vala (valac) is available.
# Examples:
#   check_runtime_vala "Linter"
check_runtime_vala() {
  local _TOOL_DESC_VALA="${1:-Vala}"
  if ! resolve_bin "valac" >/dev/null 2>&1; then
    log_warn "Required runtime 'valac' for ${_TOOL_DESC_VALA:-} is missing. Skipping."
    return 1
  fi
  return 0
}
