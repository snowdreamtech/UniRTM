#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Dart Logic Module

# Purpose: Installs Dart runtime via unirtm.
# Delegate: Managed by unirtm (.unirtm.toml)
install_runtime_dart() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install Dart runtime."
    return 0
  fi
  unirtm install dart
}

# Purpose: Sets up Dart runtime.
setup_dart() {
  if ! has_lang_files "pubspec.yaml" "*.dart"; then
    return 0
  fi

  setup_registry_dart

  local _T0_DART_RT
  _T0_DART_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version dart)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "dart")

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "Dart" "✅ Detected" "${_CUR_VER:-}" "0"
    return 0
  fi

  _log_setup "Dart Runtime" "dart"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "Runtime" "Dart" "⚖️ Previewed" "-" "0"
    return 0
  fi

  local _STAT_DART_RT="✅ Installed"
  install_runtime_dart || _STAT_DART_RT="❌ Failed"

  local _DUR_DART_RT
  _DUR_DART_RT=$(($(date +%s) - _T0_DART_RT))
  log_summary "Runtime" "Dart" "${_STAT_DART_RT:-}" "$(get_version dart --version | head -n 1)" "${_DUR_DART_RT:-}"
}
# Purpose: Checks if Dart runtime is available.
# Examples:
#   check_runtime_dart "Linter"
check_runtime_dart() {
  local _TOOL_DESC_DART="${1:-Dart}"
  if ! resolve_bin "dart" >/dev/null 2>&1; then
    log_warn "Required runtime 'dart' for ${_TOOL_DESC_DART:-} is missing. Skipping."
    return 1
  fi
  return 0
}
