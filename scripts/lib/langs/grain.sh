#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Grain Logic Module

# Purpose: Installs Grain via unirtm.
install_runtime_grain() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install Grain via unirtm."
    return 0
  fi
  # Version pinned in scripts/lib/versions.sh (VER_GRAIN_PROVIDER, VER_GRAIN)
  "${_G_UNIRTM_BIN:-unirtm}" install "${VER_GRAIN_PROVIDER:-}@${VER_GRAIN:-}"
}

# Purpose: Sets up Grain environment for project.
setup_grain() {
  if ! has_lang_files "grain.toml" "*.gr"; then
    return 0
  fi

  setup_registry_grain

  local _T0_GRAIN_RT
  _T0_GRAIN_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version grain)
  local _REQ_VER="${VER_GRAIN:-}"

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "Grain" "✅ Detected" "${_CUR_VER:-}" "0"
    return 0
  fi

  _log_setup "Grain" "grain"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "Runtime" "Grain" "⚖️ Previewed" "-" "0"
    return 0
  fi

  local _STAT_GRAIN_RT="✅ Installed"
  install_runtime_grain || _STAT_GRAIN_RT="❌ Failed"

  local _DUR_GRAIN_RT
  _DUR_GRAIN_RT=$(($(date +%s) - _T0_GRAIN_RT))
  log_summary "Runtime" "Grain" "${_STAT_GRAIN_RT:-}" "$(get_version grain --version | head -n 1 | awk '{print $NF}')" "${_DUR_GRAIN_RT:-}"
}

# Purpose: Checks if Grain is available.
# Examples:
#   check_runtime_grain "Linter"
check_runtime_grain() {
  local _TOOL_DESC_GRAIN="${1:-Grain}"
  if ! resolve_bin "grain" >/dev/null 2>&1; then
    log_warn "Required runtime 'grain' for ${_TOOL_DESC_GRAIN:-} is missing. Skipping."
    return 1
  fi
  return 0
}
