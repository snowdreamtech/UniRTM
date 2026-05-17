#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# WebAssembly Text (WAT) Logic Module

# Purpose: Installs Wasmtime via unirtm.
# Delegate: Managed by unirtm (.unirtm.toml)
install_runtime_wat() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install Wasmtime via unirtm."
    return 0
  fi

  # shellcheck disable=SC2154
  "${_G_UNIRTM_BIN:-unirtm}" install "wasmtime@$(get_unirtm_tool_version wasmtime)"
}

# Purpose: Sets up WAT environment for project.
setup_wat() {
  if ! has_lang_files "" "*.wat *.wasm"; then
    return 0
  fi

  setup_registry_wasmtime

  local _T0_WAT_RT
  _T0_WAT_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version wasmtime)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "wasmtime")

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "WebAssembly" "✅ Detected" "${_CUR_VER:-}" "0"
    return 0
  fi

  _log_setup "WebAssembly" "wasmtime"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "Runtime" "WebAssembly" "⚖️ Previewed" "-" "0"
    return 0
  fi

  local _STAT_WAT_RT="✅ Installed"
  install_runtime_wat || _STAT_WAT_RT="❌ Failed"

  local _DUR_WAT_RT
  _DUR_WAT_RT=$(($(date +%s) - _T0_WAT_RT))
  log_summary "Runtime" "WebAssembly" "${_STAT_WAT_RT:-}" "$(get_version wasmtime --version | head -n 1 | awk '{print $NF}')" "${_DUR_WAT_RT:-}"
}

# Purpose: Checks if Wasmtime is available.
# Examples:
#   check_runtime_wat "Linter"
check_runtime_wat() {
  local _TOOL_DESC_WAT="${1:-WebAssembly}"
  if ! resolve_bin "wasmtime" >/dev/null 2>&1; then
    log_warn "Required runtime 'wasmtime' for ${_TOOL_DESC_WAT:-} is missing. Skipping."
    return 1
  fi
  return 0
}
