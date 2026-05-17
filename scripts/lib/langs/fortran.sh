#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Fortran Logic Module

# Purpose: Installs GFortran (via GCC) via unirtm.
# Delegate: Managed by unirtm (.unirtm.toml)
install_runtime_fortran() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install GFortran via mise."
    return 0
  fi

  # shellcheck disable=SC2154
  unirtm install "gcc@$(get_unirtm_tool_version gcc)"
}

# Purpose: Sets up Fortran environment for project.
setup_fortran() {
  if ! has_lang_files "" "*.f *.for *.f90 *.f95"; then
    return 0
  fi

  local _T0_FORT_RT
  _T0_FORT_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version gfortran)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "gfortran")

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "Fortran" "✅ Detected" "${_CUR_VER:-}" "0"
    return 0
  fi

  _log_setup "Fortran" "gfortran"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "Runtime" "Fortran" "⚖️ Previewed" "-" "0"
    return 0
  fi

  local _STAT_FORT_RT="✅ Installed"
  install_runtime_fortran || _STAT_FORT_RT="❌ Failed"

  local _DUR_FORT_RT
  _DUR_FORT_RT=$(($(date +%s) - _T0_FORT_RT))
  log_summary "Runtime" "Fortran" "${_STAT_FORT_RT:-}" "$(get_version gfortran --version | head -n 1 | awk '{print $NF}')" "${_DUR_FORT_RT:-}"
}

# Purpose: Checks if GFortran is available.
# Examples:
#   check_runtime_fortran "Linter"
check_runtime_fortran() {
  local _TOOL_DESC_FORT="${1:-Fortran}"
  if ! resolve_bin "gfortran" >/dev/null 2>&1; then
    log_warn "Required runtime 'gfortran' for ${_TOOL_DESC_FORT:-} is missing. Skipping."
    return 1
  fi
  return 0
}
