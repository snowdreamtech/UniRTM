#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# KCL Logic Module

# Purpose: Installs KCL via unirtm (version pinned in scripts/lib/versions.sh).
install_runtime_kcl() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install KCL via unirtm."
    return 0
  fi
  unirtm install "${VER_KCL_PROVIDER:-}@${VER_KCL:-}"
}

# Purpose: Sets up KCL environment for project.
setup_kcl() {
  if ! has_lang_files "kcl.mod" "*.k"; then
    return 0
  fi

  setup_registry_kcl

  local _T0_KCL_RT
  _T0_KCL_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version kcl)
  local _REQ_VER="${VER_KCL:-}"

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "KCL" "✅ Detected" "${_CUR_VER:-}" "0"
    return 0
  fi

  _log_setup "KCL" "kcl"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "Runtime" "KCL" "⚖️ Previewed" "-" "0"
    return 0
  fi

  local _STAT_KCL_RT="✅ Installed"
  install_runtime_kcl || _STAT_KCL_RT="❌ Failed"

  local _DUR_KCL_RT
  _DUR_KCL_RT=$(($(date +%s) - _T0_KCL_RT))
  log_summary "Runtime" "KCL" "${_STAT_KCL_RT:-}" "$(get_version kcl --version | head -n 1 | awk '{print $3}')" "${_DUR_KCL_RT:-}"
}

# Purpose: Checks if KCL is available.
check_runtime_kcl() {
  local _TOOL_DESC_KCL="${1:-KCL}"
  if ! resolve_bin "kcl" >/dev/null 2>&1; then
    log_warn "Required runtime 'kcl' for ${_TOOL_DESC_KCL:-} is missing. Skipping."
    return 1
  fi
  return 0
}
