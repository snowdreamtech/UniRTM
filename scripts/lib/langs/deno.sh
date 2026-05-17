#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Deno Logic Module

# Purpose: Installs Deno runtime via unirtm.
# Delegate: Managed by unirtm (.unirtm.toml)
install_runtime_deno() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install Deno runtime."
    return 0
  fi
  unirtm install deno
}

# Purpose: Sets up Deno runtime.
setup_deno() {
  if ! has_lang_files "deno.json deno.jsonc" ""; then
    return 0
  fi

  # Dynamically register Deno in .unirtm.toml if not already present.
  setup_registry_deno

  local _T0_DENO_RT
  _T0_DENO_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version deno)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "deno")

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "Deno" "✅ Detected" "${_CUR_VER:-}" "0"
    return 0
  fi

  _log_setup "Deno Runtime" "deno"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "Runtime" "Deno" "⚖️ Previewed" "-" "0"
    return 0
  fi

  local _STAT_DENO_RT="✅ Installed"
  install_runtime_deno || _STAT_DENO_RT="❌ Failed"

  local _DUR_DENO_RT
  _DUR_DENO_RT=$(($(date +%s) - _T0_DENO_RT))
  log_summary "Runtime" "Deno" "${_STAT_DENO_RT:-}" "$(get_version deno --version | head -n 1 | awk '{print $2}')" "${_DUR_DENO_RT:-}"
}
# Purpose: Checks if Deno runtime is available.
# Examples:
#   check_runtime_deno "Linter"
check_runtime_deno() {
  local _TOOL_DESC_DENO="${1:-Deno}"
  if ! resolve_bin "deno" >/dev/null 2>&1; then
    log_warn "Required runtime 'deno' for ${_TOOL_DESC_DENO:-} is missing. Skipping."
    return 1
  fi
  return 0
}
