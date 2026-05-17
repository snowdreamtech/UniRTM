#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Zig Logic Module

# Purpose: Installs Zig runtime via unirtm.
# Delegate: Managed by unirtm (.unirtm.toml)
install_runtime_zig() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install Zig runtime."
    return 0
  fi
  unirtm install zig
}

# Purpose: Sets up Zig runtime.
setup_zig() {
  if ! has_lang_files "build.zig" "*.zig"; then
    return 0
  fi

  # Dynamically register Zig in .unirtm.toml if not already present.
  setup_registry_zig

  local _T0_ZIG_RT
  _T0_ZIG_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version zig)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "zig")

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "Zig" "✅ Detected" "${_CUR_VER:-}" "0"
    return 0
  fi

  _log_setup "Zig Runtime" "zig"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "Runtime" "Zig" "⚖️ Previewed" "-" "0"
    return 0
  fi

  local _STAT_ZIG_RT="✅ Installed"
  install_runtime_zig || _STAT_ZIG_RT="❌ Failed"

  local _DUR_ZIG_RT
  _DUR_ZIG_RT=$(($(date +%s) - _T0_ZIG_RT))
  log_summary "Runtime" "Zig" "${_STAT_ZIG_RT:-}" "$(get_version zig version)" "${_DUR_ZIG_RT:-}"
}
# Purpose: Checks if Zig runtime is available.
# Examples:
#   check_runtime_zig "Linter"
check_runtime_zig() {
  local _TOOL_DESC_ZIG="${1:-Zig}"
  if ! resolve_bin "zig" >/dev/null 2>&1; then
    log_warn "Required runtime 'zig' for ${_TOOL_DESC_ZIG:-} is missing. Skipping."
    return 1
  fi
  return 0
}
