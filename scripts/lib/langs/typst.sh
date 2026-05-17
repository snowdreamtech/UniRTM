#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Typst Logic Module

# Purpose: Installs Typst via unirtm.
# Delegate: Managed by unirtm (.unirtm.toml)
install_runtime_typst() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install Typst via unirtm."
    return 0
  fi

  # shellcheck disable=SC2154
  "${_G_UNIRTM_BIN:-unirtm}" install "typst@$(get_unirtm_tool_version typst)"
}

# Purpose: Sets up Typst environment for project.
setup_typst() {
  if ! has_lang_files "" "*.typ"; then
    return 0
  fi

  local _T0_TYPST
  _T0_TYPST=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version typst)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "typst")

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Docs" "Typst" "✅ Detected" "${_CUR_VER:-}" "0"
    return 0
  fi

  _log_setup "Typst" "typst"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "Docs" "Typst" "⚖️ Previewed" "-" "0"
    return 0
  fi

  local _STAT_TYPST="✅ Installed"
  install_runtime_typst || _STAT_TYPST="❌ Failed"

  local _DUR_TYPST
  _DUR_TYPST=$(($(date +%s) - _T0_TYPST))
  log_summary "Docs" "Typst" "${_STAT_TYPST:-}" "$(get_version typst --version | awk '{print $2}')" "${_DUR_TYPST:-}"
}

# Purpose: Checks if Typst is available.
# Examples:
#   check_runtime_typst "Linter"
check_runtime_typst() {
  local _TOOL_DESC_TYPST="${1:-Typst}"
  if ! resolve_bin "typst" >/dev/null 2>&1; then
    log_warn "Required tool 'typst' for ${_TOOL_DESC_TYPST:-} is missing. Skipping."
    return 1
  fi
  return 0
}
