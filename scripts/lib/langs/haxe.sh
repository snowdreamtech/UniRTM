#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Haxe Logic Module

# Purpose: Installs Haxe via unirtm.
# Delegate: Managed by unirtm (.unirtm.toml)
install_runtime_haxe() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install Haxe via mise."
    return 0
  fi

  # shellcheck disable=SC2154
  unirtm install "haxe@$(get_unirtm_tool_version haxe)"
}

# Purpose: Sets up Haxe environment for project.
setup_haxe() {
  if ! has_lang_files "*.hxml" "*.hx"; then
    return 0
  fi

  setup_registry_haxe

  local _T0_HX_RT
  _T0_HX_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version haxe)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "haxe")

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "Haxe" "✅ Detected" "${_CUR_VER:-}" "0"
    return 0
  fi

  _log_setup "Haxe" "haxe"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "Runtime" "Haxe" "⚖️ Previewed" "-" "0"
    return 0
  fi

  local _STAT_HX_RT="✅ Installed"
  install_runtime_haxe || _STAT_HX_RT="❌ Failed"

  local _DUR_HX_RT
  _DUR_HX_RT=$(($(date +%s) - _T0_HX_RT))
  log_summary "Runtime" "Haxe" "${_STAT_HX_RT:-}" "$(get_version haxe -version | head -n 1)" "${_DUR_HX_RT:-}"
}

# Purpose: Checks if Haxe is available.
# Examples:
#   check_runtime_haxe "Linter"
check_runtime_haxe() {
  local _TOOL_DESC_HX="${1:-Haxe}"
  if ! resolve_bin "haxe" >/dev/null 2>&1; then
    log_warn "Required runtime 'haxe' for ${_TOOL_DESC_HX:-} is missing. Skipping."
    return 1
  fi
  return 0
}
