#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# R Logic Module

# Purpose: Installs R runtime via unirtm.
# Delegate: Managed by unirtm (.unirtm.toml)
install_runtime_r() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install R runtime."
    return 0
  fi
  # shellcheck disable=SC2154
  unirtm install "R@$(get_unirtm_tool_version r)"
}

# Purpose: Sets up R runtime.
setup_r() {
  if ! has_lang_files "DESCRIPTION" "*.R *.Rmd"; then
    return 0
  fi

  setup_registry_r

  local _T0_R_RT
  _T0_R_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version R)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "R")

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "R" "✅ Detected" "${_CUR_VER:-}" "0"
    return 0
  fi

  _log_setup "R Runtime" "R"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "Runtime" "R" "⚖️ Previewed" "-" "0"
    return 0
  fi

  local _STAT_R_RT="✅ Installed"
  install_runtime_r || _STAT_R_RT="❌ Failed"

  local _DUR_R_RT
  _DUR_R_RT=$(($(date +%s) - _T0_R_RT))
  log_summary "Runtime" "R" "${_STAT_R_RT:-}" "$(get_version R --version | head -n 1)" "${_DUR_R_RT:-}"
}
# Purpose: Checks if R runtime is available.
# Examples:
#   check_runtime_r "Linter"
check_runtime_r() {
  local _TOOL_DESC_R="${1:-R}"
  if ! command -v R >/dev/null 2>&1; then
    log_warn "Required runtime 'R' for ${_TOOL_DESC_R:-} is missing. Skipping."
    return 1
  fi
  return 0
}
