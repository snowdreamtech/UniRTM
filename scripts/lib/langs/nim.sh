#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Nim Logic Module

# Purpose: Installs Nim compiler via unirtm.
# Delegate: Managed by unirtm (.unirtm.toml)
install_runtime_nim() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install Nim compiler via unirtm."
    return 0
  fi

  # shellcheck disable=SC2154
  unirtm install "${VER_NIM_PROVIDER:-}@${VER_NIM:-}"
}

# Purpose: Sets up Nim environment for project.
setup_nim() {
  if ! has_lang_files "nim.cfg *.nimble" "*.nim *.nims"; then
    return 0
  fi

  setup_registry_nim

  local _T0_NIM_RT
  _T0_NIM_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version nim)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "nim")

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "Nim" "✅ Detected" "${_CUR_VER:-}" "0"
    return 0
  fi

  _log_setup "Nim" "nim"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "Runtime" "Nim" "⚖️ Previewed" "-" "0"
    return 0
  fi

  local _STAT_NIM_RT="✅ Installed"
  install_runtime_nim || _STAT_NIM_RT="❌ Failed"

  local _DUR_NIM_RT
  _DUR_NIM_RT=$(($(date +%s) - _T0_NIM_RT))
  log_summary "Runtime" "Nim" "${_STAT_NIM_RT:-}" "$(get_version nim --version | head -n 1)" "${_DUR_NIM_RT:-}"
}

# Purpose: Checks if Nim compiler is available.
# Examples:
#   check_runtime_nim "Linter"
check_runtime_nim() {
  local _TOOL_DESC_NIM="${1:-Nim}"
  if ! resolve_bin "nim" >/dev/null 2>&1; then
    log_warn "Required runtime 'nim' for ${_TOOL_DESC_NIM:-} is missing. Skipping."
    return 1
  fi
  return 0
}
