#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Odin Logic Module

# Purpose: Installs Odin compiler via unirtm.
# Delegate: Managed by unirtm (.unirtm.toml)
install_runtime_odin() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install Odin compiler via mise."
    return 0
  fi

  # shellcheck disable=SC2154
  unirtm install "odin@$(get_unirtm_tool_version odin)"
}

# Purpose: Sets up Odin environment for project.
setup_odin() {
  if ! has_lang_files "" "*.odin"; then
    return 0
  fi

  setup_registry_odin

  local _T0_ODIN_RT
  _T0_ODIN_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version odin)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "odin")

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "Odin" "✅ Detected" "${_CUR_VER:-}" "0"
    return 0
  fi

  _log_setup "Odin" "odin"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "Runtime" "Odin" "⚖️ Previewed" "-" "0"
    return 0
  fi

  local _STAT_ODIN_RT="✅ Installed"
  install_runtime_odin || _STAT_ODIN_RT="❌ Failed"

  local _DUR_ODIN_RT
  _DUR_ODIN_RT=$(($(date +%s) - _T0_ODIN_RT))
  log_summary "Runtime" "Odin" "${_STAT_ODIN_RT:-}" "$(get_version odin version)" "${_DUR_ODIN_RT:-}"
}

# Purpose: Checks if Odin compiler is available.
# Examples:
#   check_runtime_odin "Linter"
check_runtime_odin() {
  local _TOOL_DESC_ODIN="${1:-Odin}"
  if ! resolve_bin "odin" >/dev/null 2>&1; then
    log_warn "Required runtime 'odin' for ${_TOOL_DESC_ODIN:-} is missing. Skipping."
    return 1
  fi
  return 0
}
