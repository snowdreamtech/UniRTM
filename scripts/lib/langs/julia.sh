#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Julia Logic Module

# Purpose: Installs Julia runtime via unirtm.
install_runtime_julia() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install Julia runtime."
    return 0
  fi
  # shellcheck disable=SC2154
  "${_G_UNIRTM_BIN:-unirtm}" install "julia@$(get_unirtm_tool_version julia)"
}

# Purpose: Sets up Julia runtime.
setup_julia() {
  if ! has_lang_files "Project.toml" "*.jl"; then
    return 0
  fi

  setup_registry_julia

  local _T0_JULIA_RT
  _T0_JULIA_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version julia)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "julia")

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "Julia" "✅ Detected" "${_CUR_VER:-}" "0"
    return 0
  fi

  _log_setup "Julia Runtime" "julia"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "Runtime" "Julia" "⚖️ Previewed" "-" "0"
    return 0
  fi

  local _STAT_JULIA_RT="✅ Installed"
  install_runtime_julia || _STAT_JULIA_RT="❌ Failed"

  local _DUR_JULIA_RT
  _DUR_JULIA_RT=$(($(date +%s) - _T0_JULIA_RT))
  log_summary "Runtime" "Julia" "${_STAT_JULIA_RT:-}" "$(get_version julia -v)" "${_DUR_JULIA_RT:-}"
}
# Purpose: Checks if Julia runtime is available.
# Examples:
#   check_runtime_julia "Linter"
check_runtime_julia() {
  local _TOOL_DESC_JULIA="${1:-Julia}"
  if ! resolve_bin "julia" >/dev/null 2>&1; then
    log_warn "Required runtime 'julia' for ${_TOOL_DESC_JULIA:-} is missing. Skipping."
    return 1
  fi
  return 0
}
