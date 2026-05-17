#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Luau Logic Module

# Purpose: Installs Luau via unirtm.
# Delegate: Managed by unirtm (.unirtm.toml)
install_runtime_luau() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install Luau via unirtm."
    return 0
  fi

  # shellcheck disable=SC2154
  "${_G_UNIRTM_BIN:-unirtm}" install "luau@$(get_unirtm_tool_version luau)"
}

# Purpose: Sets up Luau environment for project.
setup_luau() {
  if ! has_lang_files "selene.toml" "*.luau"; then
    return 0
  fi

  setup_registry_luau

  local _T0_LUAU_RT
  _T0_LUAU_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version luau)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "luau")

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "Luau" "✅ Detected" "${_CUR_VER:-}" "0"
    return 0
  fi

  _log_setup "Luau" "luau"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "Runtime" "Luau" "⚖️ Previewed" "-" "0"
    return 0
  fi

  local _STAT_LUAU_RT="✅ Installed"
  install_runtime_luau || _STAT_LUAU_RT="❌ Failed"

  local _DUR_LUAU_RT
  _DUR_LUAU_RT=$(($(date +%s) - _T0_LUAU_RT))
  log_summary "Runtime" "Luau" "${_STAT_LUAU_RT:-}" "$(get_version luau --version)" "${_DUR_LUAU_RT:-}"
}

# Purpose: Checks if Luau is available.
# Examples:
#   check_runtime_luau "Linter"
check_runtime_luau() {
  local _TOOL_DESC_LUAU="${1:-Luau}"
  if ! resolve_bin "luau" >/dev/null 2>&1; then
    log_warn "Required runtime 'luau' for ${_TOOL_DESC_LUAU:-} is missing. Skipping."
    return 1
  fi
  return 0
}
