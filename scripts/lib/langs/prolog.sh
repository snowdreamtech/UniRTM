#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Prolog Logic Module

# Purpose: Installs SWI-Prolog via unirtm.
# Delegate: Managed by unirtm (.unirtm.toml)
install_runtime_prolog() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install SWI-Prolog via unirtm."
    return 0
  fi

  # shellcheck disable=SC2154
  "${_G_UNIRTM_BIN:-unirtm}" install "swiprolog@$(get_unirtm_tool_version swiprolog)"
}

# Purpose: Sets up Prolog environment for project.
setup_prolog() {
  if ! has_lang_files "" "*.pl *.pro *.P"; then
    return 0
  fi

  setup_registry_prolog

  local _T0_PROLOG_RT
  _T0_PROLOG_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version swipl)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "swipl")

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "Prolog" "✅ Detected" "${_CUR_VER:-}" "0"
    return 0
  fi

  _log_setup "Prolog" "swipl"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "Runtime" "Prolog" "⚖️ Previewed" "-" "0"
    return 0
  fi

  local _STAT_PROLOG_RT="✅ Installed"
  install_runtime_prolog || _STAT_PROLOG_RT="❌ Failed"

  local _DUR_PROLOG_RT
  _DUR_PROLOG_RT=$(($(date +%s) - _T0_PROLOG_RT))
  log_summary "Runtime" "Prolog" "${_STAT_PROLOG_RT:-}" "$(get_version swipl --version | awk '{print $NF}')" "${_DUR_PROLOG_RT:-}"
}

# Purpose: Checks if SWI-Prolog is available.
# Examples:
#   check_runtime_prolog "Linter"
check_runtime_prolog() {
  local _TOOL_DESC_PROLOG="${1:-Prolog}"
  if ! resolve_bin "swipl" >/dev/null 2>&1; then
    log_warn "Required runtime 'swipl' for ${_TOOL_DESC_PROLOG:-} is missing. Skipping."
    return 1
  fi
  return 0
}
