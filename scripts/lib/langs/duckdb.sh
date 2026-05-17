#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# DuckDB Logic Module

# Purpose: Installs DuckDB via unirtm.
# Delegate: Managed by unirtm (.unirtm.toml)
install_runtime_duckdb() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install DuckDB via mise."
    return 0
  fi

  # shellcheck disable=SC2154
  unirtm install "duckdb@$(get_unirtm_tool_version duckdb)"
}

# Purpose: Sets up DuckDB environment for project.
setup_duckdb() {
  if ! has_lang_files "" "*.sql *.duckdb"; then
    return 0
  fi

  local _T0_DUCK_RT
  _T0_DUCK_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version duckdb)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "duckdb")

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "DuckDB" "✅ Detected" "${_CUR_VER:-}" "0"
    return 0
  fi

  _log_setup "DuckDB" "duckdb"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "Runtime" "DuckDB" "⚖️ Previewed" "-" "0"
    return 0
  fi

  local _STAT_DUCK_RT="✅ Installed"
  install_runtime_duckdb || _STAT_DUCK_RT="❌ Failed"

  local _DUR_DUCK_RT
  _DUR_DUCK_RT=$(($(date +%s) - _T0_DUCK_RT))
  log_summary "Runtime" "DuckDB" "${_STAT_DUCK_RT:-}" "$(get_version duckdb --version | awk '{print $1}')" "${_DUR_DUCK_RT:-}"
}

# Purpose: Checks if DuckDB is available.
# Examples:
#   check_runtime_duckdb "Linter"
check_runtime_duckdb() {
  local _TOOL_DESC_DUCK="${1:-DuckDB}"
  if ! resolve_bin "duckdb" >/dev/null 2>&1; then
    log_warn "Required runtime 'duckdb' for ${_TOOL_DESC_DUCK:-} is missing. Skipping."
    return 1
  fi
  return 0
}
