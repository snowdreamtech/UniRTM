#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Tcl Logic Module

# Purpose: Installs Tcl via unirtm.
# Delegate: Managed by unirtm (.unirtm.toml)
install_runtime_tcl() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install Tcl via unirtm."
    return 0
  fi

  # shellcheck disable=SC2154
  unirtm install "tcl@$(get_unirtm_tool_version tcl)"
}

# Purpose: Sets up Tcl environment for project.
setup_tcl() {
  if ! has_lang_files "" "*.tcl *.tk *.exp"; then
    return 0
  fi

  setup_registry_tcl

  local _T0_TCL_RT
  _T0_TCL_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version tclsh)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "tclsh")

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "Tcl" "✅ Detected" "${_CUR_VER:-}" "0"
    return 0
  fi

  _log_setup "Tcl" "tclsh"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "Runtime" "Tcl" "⚖️ Previewed" "-" "0"
    return 0
  fi

  local _STAT_TCL_RT="✅ Installed"
  install_runtime_tcl || _STAT_TCL_RT="❌ Failed"

  local _DUR_TCL_RT
  _DUR_TCL_RT=$(($(date +%s) - _T0_TCL_RT))
  log_summary "Runtime" "Tcl" "${_STAT_TCL_RT:-}" "$(get_version tclsh "echo 'puts [info patchlevel]' | tclsh" | awk '{print $NF}')" "${_DUR_TCL_RT:-}"
}

# Purpose: Checks if Tcl is available.
# Examples:
#   check_runtime_tcl "Linter"
check_runtime_tcl() {
  local _TOOL_DESC_TCL="${1:-Tcl}"
  if ! resolve_bin "tclsh" >/dev/null 2>&1; then
    log_warn "Required runtime 'tclsh' for ${_TOOL_DESC_TCL:-} is missing. Skipping."
    return 1
  fi
  return 0
}
