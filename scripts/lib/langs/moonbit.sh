#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# MoonBit Logic Module

# Purpose: Installs MoonBit via unirtm (version pinned in scripts/lib/versions.sh).
install_runtime_moonbit() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install MoonBit via unirtm."
    return 0
  fi
  unirtm install "${VER_MOONBIT_PROVIDER:-}@${VER_MOONBIT:-}"
}

# Purpose: Sets up MoonBit environment for project.
setup_moonbit() {
  if ! has_lang_files "moon.mod.json" "*.mbt"; then
    return 0
  fi

  setup_registry_moonbit

  local _T0_MOON_RT
  _T0_MOON_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version moon)
  local _REQ_VER="${VER_MOONBIT:-}"

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "MoonBit" "✅ Detected" "${_CUR_VER:-}" "0"
    return 0
  fi

  _log_setup "MoonBit" "moon"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "Runtime" "MoonBit" "⚖️ Previewed" "-" "0"
    return 0
  fi

  local _STAT_MOON_RT="✅ Installed"
  install_runtime_moonbit || _STAT_MOON_RT="❌ Failed"

  local _DUR_MOON_RT
  _DUR_MOON_RT=$(($(date +%s) - _T0_MOON_RT))
  log_summary "Runtime" "MoonBit" "${_STAT_MOON_RT:-}" "$(get_version moon version | head -n 1 | awk '{print $NF}')" "${_DUR_MOON_RT:-}"
}

# Purpose: Checks if MoonBit is available.
# Examples:
#   check_runtime_moonbit "Linter"
check_runtime_moonbit() {
  local _TOOL_DESC_MOON="${1:-MoonBit}"
  if ! resolve_bin "moon" >/dev/null 2>&1; then
    log_warn "Required runtime 'moon' for ${_TOOL_DESC_MOON:-} is missing. Skipping."
    return 1
  fi
  return 0
}
