#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Jsonnet Logic Module

# Purpose: Installs go-jsonnet via unirtm.
# Delegate: Managed by unirtm (.unirtm.toml)
install_runtime_jsonnet() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install go-jsonnet via unirtm."
    return 0
  fi

  # shellcheck disable=SC2154
  "${_G_UNIRTM_BIN:-unirtm}" install "go-jsonnet@$(get_unirtm_tool_version jsonnet)"
}

# Purpose: Sets up Jsonnet environment for project.
# Delegate: Managed by unirtm (.unirtm.toml)
# Examples:
#   setup_jsonnet
setup_jsonnet() {
  if ! has_lang_files "" "*.jsonnet *.libsonnet"; then
    return 0
  fi

  local _T0_JSONNET_RT
  _T0_JSONNET_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version jsonnet)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "jsonnet")

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "Jsonnet" "✅ Detected" "${_CUR_VER:-}" "0"
    return 0
  fi

  _log_setup "Jsonnet" "jsonnet"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "Runtime" "Jsonnet" "⚖️ Previewed" "-" "0"
    return 0
  fi

  local _STAT_JSONNET_RT="✅ Installed"
  install_runtime_jsonnet || _STAT_JSONNET_RT="❌ Failed"

  local _DUR_JSONNET_RT
  _DUR_JSONNET_RT=$(($(date +%s) - _T0_JSONNET_RT))
  log_summary "Runtime" "Jsonnet" "${_STAT_JSONNET_RT:-}" "$(get_version jsonnet --version | head -n 1 | awk '{print $NF}')" "${_DUR_JSONNET_RT:-}"
}

# Purpose: Checks if Jsonnet is available.
# Examples:
#   check_runtime_jsonnet "Linter"
check_runtime_jsonnet() {
  local _TOOL_DESC_JSONNET="${1:-Jsonnet}"
  if ! resolve_bin "jsonnet" >/dev/null 2>&1; then
    log_warn "Required runtime 'jsonnet' for ${_TOOL_DESC_JSONNET:-} is missing. Skipping."
    return 1
  fi
  return 0
}
