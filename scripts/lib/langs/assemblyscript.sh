#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# AssemblyScript Logic Module

# Purpose: Installs AssemblyScript (asc) via unirtm (npm provider).
# Delegate: Managed by unirtm (.unirtm.toml)
install_runtime_assemblyscript() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install AssemblyScript (asc) via unirtm npm provider."
    return 0
  fi

  # shellcheck disable=SC2154
  unirtm install "npm:assemblyscript@$(get_unirtm_tool_version assemblyscript)"
}

# Purpose: Sets up AssemblyScript environment for project.
setup_assemblyscript() {
  if ! has_lang_files "asconfig.json" "*.as"; then
    return 0
  fi

  local _T0_AS_RT
  _T0_AS_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version asc)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "asc")

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "AssemblyScript" "✅ Detected" "${_CUR_VER:-}" "0"
    return 0
  fi

  _log_setup "AssemblyScript" "asc"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "Runtime" "AssemblyScript" "⚖️ Previewed" "-" "0"
    return 0
  fi

  local _STAT_AS_RT="✅ Installed"
  install_runtime_assemblyscript || _STAT_AS_RT="❌ Failed"

  local _DUR_AS_RT
  _DUR_AS_RT=$(($(date +%s) - _T0_AS_RT))
  log_summary "Runtime" "AssemblyScript" "${_STAT_AS_RT:-}" "$(get_version asc --version)" "${_DUR_AS_RT:-}"
}

# Purpose: Checks if AssemblyScript is available.
# Examples:
#   check_runtime_assemblyscript "Linter"
check_runtime_assemblyscript() {
  local _TOOL_DESC_AS="${1:-AssemblyScript}"
  if ! resolve_bin "asc" >/dev/null 2>&1; then
    log_warn "Required runtime 'asc' for ${_TOOL_DESC_AS:-} is missing. Skipping."
    return 1
  fi
  return 0
}
