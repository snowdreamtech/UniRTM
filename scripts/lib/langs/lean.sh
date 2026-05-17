#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Lean 4 Logic Module

# Purpose: Installs Lean 4 via unirtm.
# Delegate: Managed by unirtm (.unirtm.toml)
install_runtime_lean() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install Lean 4 via unirtm."
    return 0
  fi

  # shellcheck disable=SC2154
  unirtm install "${VER_LEAN_PROVIDER:-}@${VER_LEAN:-}"
}

# Purpose: Sets up Lean 4 environment for project.
setup_lean() {
  if ! has_lang_files "lean-toolchain lakefile.lean" "*.lean"; then
    return 0
  fi

  setup_registry_lean

  local _T0_LEAN_RT
  _T0_LEAN_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version lean)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "lean")

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "Lean 4" "✅ Detected" "${_CUR_VER:-}" "0"
    return 0
  fi

  _log_setup "Lean 4" "lean"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "Runtime" "Lean 4" "⚖️ Previewed" "-" "0"
    return 0
  fi

  local _STAT_LEAN_RT="✅ Installed"
  install_runtime_lean || _STAT_LEAN_RT="❌ Failed"

  local _DUR_LEAN_RT
  _DUR_LEAN_RT=$(($(date +%s) - _T0_LEAN_RT))
  log_summary "Runtime" "Lean 4" "${_STAT_LEAN_RT:-}" "$(get_version lean --version | awk '{print $NF}')" "${_DUR_LEAN_RT:-}"
}

# Purpose: Checks if Lean 4 is available.
# Delegate: N/A
# Examples:
#   check_runtime_lean "Linter"
check_runtime_lean() {
  local _TOOL_DESC_LEAN="${1:-Lean 4}"
  if ! resolve_bin "lean" >/dev/null 2>&1; then
    log_warn "Required runtime 'lean' for ${_TOOL_DESC_LEAN:-} is missing. Skipping."
    return 1
  fi
  return 0
}
