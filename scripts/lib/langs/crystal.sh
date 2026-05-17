#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Crystal Logic Module

# Purpose: Installs Crystal and Shards via unirtm.
# Delegate: Managed by unirtm (.unirtm.toml)
install_runtime_crystal() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install Crystal via unirtm."
    return 0
  fi

  # shellcheck disable=SC2154
  unirtm install "crystal@$(get_unirtm_tool_version crystal)"
}

# Purpose: Sets up Crystal environment for project.
setup_crystal() {
  if ! has_lang_files "shard.yml" "*.cr"; then
    return 0
  fi

  setup_registry_crystal

  local _T0_CRY_RT
  _T0_CRY_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version crystal)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "crystal")

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "Crystal" "✅ Detected" "${_CUR_VER:-}" "0"
    return 0
  fi

  _log_setup "Crystal" "crystal"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "Runtime" "Crystal" "⚖️ Previewed" "-" "0"
    return 0
  fi

  local _STAT_CRY_RT="✅ Installed"
  install_runtime_crystal || _STAT_CRY_RT="❌ Failed"

  local _DUR_CRY_RT
  _DUR_CRY_RT=$(($(date +%s) - _T0_CRY_RT))
  log_summary "Runtime" "Crystal" "${_STAT_CRY_RT:-}" "$(get_version crystal --version | head -n 1 | awk '{print $2}')" "${_DUR_CRY_RT:-}"
}

# Purpose: Checks if Crystal is available.
# Examples:
#   check_runtime_crystal "Linter"
check_runtime_crystal() {
  local _TOOL_DESC_CRY="${1:-Crystal}"
  if ! resolve_bin "crystal" >/dev/null 2>&1; then
    log_warn "Required runtime 'crystal' for ${_TOOL_DESC_CRY:-} is missing. Skipping."
    return 1
  fi
  return 0
}
