#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# ReScript Logic Module

# Purpose: Installs ReScript compiler (rescript) via unirtm (npm provider).
# Delegate: Managed by unirtm (.unirtm.toml)
install_runtime_rescript() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install ReScript compiler via mise npm provider."
    return 0
  fi

  # shellcheck disable=SC2154
  unirtm install "npm:rescript@$(get_unirtm_tool_version rescript)"
}

# Purpose: Sets up ReScript environment for project.
setup_rescript() {
  if ! has_lang_files "rescript.json" "*.res *.resi"; then
    return 0
  fi

  setup_registry_rescript

  local _T0_RES_RT
  _T0_RES_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version rescript)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "rescript")

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "ReScript" "✅ Detected" "${_CUR_VER:-}" "0"
    return 0
  fi

  _log_setup "ReScript" "rescript"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "Runtime" "ReScript" "⚖️ Previewed" "-" "0"
    return 0
  fi

  local _STAT_RES_RT="✅ Installed"
  install_runtime_rescript || _STAT_RES_RT="❌ Failed"

  local _DUR_RES_RT
  _DUR_RES_RT=$(($(date +%s) - _T0_RES_RT))
  log_summary "Runtime" "ReScript" "${_STAT_RES_RT:-}" "$(get_version rescript --version)" "${_DUR_RES_RT:-}"
}

# Purpose: Checks if ReScript is available.
# Examples:
#   check_runtime_rescript "Linter"
check_runtime_rescript() {
  local _TOOL_DESC_RES="${1:-ReScript}"
  if ! resolve_bin "rescript" >/dev/null 2>&1; then
    log_warn "Required runtime 'rescript' for ${_TOOL_DESC_RES:-} is missing. Skipping."
    return 1
  fi
  return 0
}
