#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# PHP Logic Module

# Purpose: Installs PHP runtime via unirtm.
# Delegate: Managed via unirtm (.unirtm.toml) and composer.
install_runtime_php() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install PHP runtime."
    return 0
  fi

  # 1. Runtime initialization
  unirtm install php

  # 2. Dependency resolution
  if [ -f "composer.json" ]; then
    run_quiet composer install --no-interaction --prefer-dist
  fi
}

# Purpose: Sets up PHP runtime.
# Delegate: Managed by unirtm (.unirtm.toml)
setup_php() {
  if ! has_lang_files "composer.json composer.lock" "*.php"; then
    return 0
  fi

  setup_registry_php

  local _T0_PHP_RT
  _T0_PHP_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version php)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "php")

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "PHP" "✅ Detected" "${_CUR_VER:-}" "0"
    return 0
  fi

  _log_setup "PHP Runtime" "php"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "Runtime" "PHP" "⚖️ Previewed" "-" "0"
    return 0
  fi

  local _STAT_PHP_RT="✅ Installed"
  install_runtime_php || _STAT_PHP_RT="❌ Failed"

  local _DUR_PHP_RT
  _DUR_PHP_RT=$(($(date +%s) - _T0_PHP_RT))
  log_summary "Runtime" "PHP" "${_STAT_PHP_RT:-}" "$(get_version php)" "${_DUR_PHP_RT:-}"
}
# Purpose: Checks if PHP runtime is available.
# Examples:
#   check_runtime_php "Linter"
check_runtime_php() {
  local _TOOL_DESC_PHP="${1:-PHP}"
  if ! resolve_bin "php" >/dev/null 2>&1; then
    log_warn "Required runtime 'php' for ${_TOOL_DESC_PHP:-} is missing. Skipping."
    return 1
  fi
  return 0
}
