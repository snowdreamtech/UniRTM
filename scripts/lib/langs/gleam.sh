#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Gleam Logic Module

# Purpose: Installs Gleam runtime via unirtm.
# Delegate: Managed by unirtm (.unirtm.toml)
install_runtime_gleam() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install Gleam runtime via unirtm."
    return 0
  fi

  # shellcheck disable=SC2154
  unirtm install "gleam@$(get_unirtm_tool_version gleam)"
}

# Purpose: Sets up Gleam environment for project.
setup_gleam() {
  if ! has_lang_files "gleam.toml" "*.gleam"; then
    return 0
  fi

  setup_registry_gleam

  local _T0_GLM_RT
  _T0_GLM_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version gleam)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "gleam")

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "Gleam" "✅ Detected" "${_CUR_VER:-}" "0"
    return 0
  fi

  _log_setup "Gleam" "gleam"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "Runtime" "Gleam" "⚖️ Previewed" "-" "0"
    return 0
  fi

  local _STAT_GLM_RT="✅ Installed"
  install_runtime_gleam || _STAT_GLM_RT="❌ Failed"

  local _DUR_GLM_RT
  _DUR_GLM_RT=$(($(date +%s) - _T0_GLM_RT))
  log_summary "Runtime" "Gleam" "${_STAT_GLM_RT:-}" "$(get_version gleam --version | head -n 1)" "${_DUR_GLM_RT:-}"
}

# Purpose: Checks if Gleam is available.
# Examples:
#   check_runtime_gleam "Linter"
check_runtime_gleam() {
  local _TOOL_DESC_GLM="${1:-Gleam}"
  if ! resolve_bin "gleam" >/dev/null 2>&1; then
    log_warn "Required runtime 'gleam' for ${_TOOL_DESC_GLM:-} is missing. Skipping."
    return 1
  fi
  return 0
}
