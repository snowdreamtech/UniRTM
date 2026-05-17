#!/usr/bin/env sh
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

set -eu

# Ada Logic Module

# Purpose: Installs Ada (GNAT) via unirtm.
# Delegate: Managed by unirtm (.unirtm.toml)
install_runtime_ada() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install Ada (GNAT) via unirtm."
    return 0
  fi

  # shellcheck disable=SC2154
  "${_G_UNIRTM_BIN:-unirtm}" install "gnat@$(get_unirtm_tool_version ada)"
}

# Purpose: Sets up Ada environment for project.
setup_ada() {
  if ! has_lang_files "default.gpr package.gpr" "*.ada *.adb *.ads"; then
    return 0
  fi

  setup_registry_ada

  local _T0_ADA_RT
  _T0_ADA_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version gnat)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "gnat")

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "Ada" "✅ Detected" "${_CUR_VER:-}" "0"
    return 0
  fi

  _log_setup "Ada" "gnat"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "Runtime" "Ada" "⚖️ Previewed" "-" "0"
    return 0
  fi

  local _STAT_ADA_RT="✅ Installed"
  install_runtime_ada || _STAT_ADA_RT="❌ Failed"

  local _DUR_ADA_RT
  _DUR_ADA_RT=$(($(date +%s) - _T0_ADA_RT))
  log_summary "Runtime" "Ada" "${_STAT_ADA_RT:-}" "$(get_version gnat --version | head -n 1 | awk '{print $NF}')" "${_DUR_ADA_RT:-}"
}

# Purpose: Checks if Ada (GNAT) is available.
# Examples:
#   check_runtime_ada "Linter"
check_runtime_ada() {
  local _TOOL_DESC_ADA="${1:-Ada}"
  if ! resolve_bin "gnat" >/dev/null 2>&1; then
    log_warn "Required runtime 'gnat' for ${_TOOL_DESC_ADA:-} is missing. Skipping."
    return 1
  fi
  return 0
}
