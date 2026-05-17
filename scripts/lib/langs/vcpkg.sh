#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# VCPKG Logic Module

# Purpose: Installs VCPKG (often used alongside C++).
# Delegate: Managed by unirtm (.unirtm.toml)
install_runtime_vcpkg() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install VCPKG."
    return 0
  fi

  # Note: vcpkg is often installed via git or system package manager.
  # Here we look for asdf/unirtm plugin or system command.
  if ! resolve_bin "vcpkg" >/dev/null 2>&1; then
    log_info "VCPKG not found. Attempting to install via unirtm (if plugin available)..."
    # shellcheck disable=SC2154
    unirtm install vcpkg || log_warn "Could not install vcpkg via unirtm. Please install it manually."
  fi
}

# Purpose: Sets up VCPKG environment for project.
setup_vcpkg() {
  if ! has_lang_files "vcpkg.json vcpkg-configuration.json" ""; then
    return 0
  fi

  local _T0_VCPKG_RT
  _T0_VCPKG_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version vcpkg)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "vcpkg")

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "VCPKG" "✅ Detected" "${_CUR_VER:-}" "0"
    return 0
  fi

  _log_setup "VCPKG" "vcpkg"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "Runtime" "VCPKG" "⚖️ Previewed" "-" "0"
    return 0
  fi

  local _STAT_VCPKG_RT="✅ Detected"
  install_runtime_vcpkg || _STAT_VCPKG_RT="❌ Missing"

  local _DUR_VCPKG_RT
  _DUR_VCPKG_RT=$(($(date +%s) - _T0_VCPKG_RT))
  log_summary "Runtime" "VCPKG" "${_STAT_VCPKG_RT:-}" "$(get_version vcpkg version | awk '{print $NF}')" "${_DUR_VCPKG_RT:-}"
}

# Purpose: Checks if VCPKG is available.
# Examples:
#   check_runtime_vcpkg "Linter"
check_runtime_vcpkg() {
  local _TOOL_DESC_VCPKG="${1:-VCPKG}"
  if ! resolve_bin "vcpkg" >/dev/null 2>&1; then
    log_warn "Required tool 'vcpkg' for ${_TOOL_DESC_VCPKG:-} is missing. Skipping."
    return 1
  fi
  return 0
}
