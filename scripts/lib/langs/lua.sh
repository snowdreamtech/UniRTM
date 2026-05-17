#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Lua Logic Module

# Purpose: Installs Lua runtime via unirtm.
# Delegate: Managed by unirtm (.unirtm.toml)
install_runtime_lua() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install Lua runtime."
    return 0
  fi
  # shellcheck disable=SC2154
  unirtm install "lua@$(get_unirtm_tool_version lua)"
}

# Purpose: Installs stylua.
# Delegate: Managed by unirtm (.unirtm.toml)
install_stylua() {
  setup_registry_stylua
  install_tool_safe "stylua" "${VER_STYLUA_PROVIDER:-}" "StyLua" "--version" 0 "*.lua" ""
}

# Purpose: Sets up Lua runtime and mandatory linting tools.
setup_lua() {
  if ! has_lang_files "" "*.lua"; then
    return 0
  fi

  setup_registry_lua

  local _T0_LUA_RT
  _T0_LUA_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version lua)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "lua")

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "Lua" "✅ Detected" "${_CUR_VER:-}" "0"
  else
    _log_setup "Lua Runtime" "lua"

    if [ "${DRY_RUN:-0}" -eq 1 ]; then
      log_summary "Runtime" "Lua" "⚖️ Previewed" "-" "0"
    else
      local _STAT_LUA_RT="✅ Installed"
      install_runtime_lua || _STAT_LUA_RT="❌ Failed"

      local _DUR_LUA_RT
      _DUR_LUA_RT=$(($(date +%s) - _T0_LUA_RT))
      log_summary "Runtime" "Lua" "${_STAT_LUA_RT:-}" "$(get_version lua -v | head -n 1)" "${_DUR_LUA_RT:-}"
    fi
  fi

  # Also ensure linting tools are present
  install_stylua
}
# Purpose: Checks if Lua runtime is available.
# Examples:
#   check_runtime_lua "Linter"
check_runtime_lua() {
  local _TOOL_DESC_LUA="${1:-Lua}"
  if ! resolve_bin "lua" >/dev/null 2>&1; then
    log_warn "Required runtime 'lua' for ${_TOOL_DESC_LUA:-} is missing. Skipping."
    return 1
  fi
  return 0
}
