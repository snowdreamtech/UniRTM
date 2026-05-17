#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Kotlin Logic Module

# Purpose: Installs Kotlin runtime via unirtm (version pinned in scripts/lib/versions.sh).
install_runtime_kotlin() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install Kotlin runtime."
    return 0
  fi
  "${_G_UNIRTM_BIN:-unirtm}" install "kotlin@${VER_KOTLIN:-}"
}

# Purpose: Installs ktlint (version pinned in scripts/lib/versions.sh).
install_ktlint() {
  if ! has_lang_files "build.gradle.kts build.gradle settings.gradle.kts" "*.kt *.kts"; then
    return 0
  fi

  setup_registry_ktlint
  install_tool_safe "ktlint" "${VER_KTLINT_PROVIDER:-}" "ktlint" "--version" 0 "*.kt *.kts" ""
}

# Purpose: Sets up Kotlin runtime and mandatory linting tools.
setup_kotlin() {
  if ! has_lang_files "build.gradle.kts build.gradle settings.gradle.kts" "*.kt *.kts"; then
    return 0
  fi

  setup_registry_kotlin

  local _T0_KOTLIN_RT
  _T0_KOTLIN_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version kotlin)
  local _REQ_VER="${VER_KOTLIN:-}"

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "Kotlin" "✅ Detected" "${_CUR_VER:-}" "0"
  else
    _log_setup "Kotlin Runtime" "kotlin"

    if [ "${DRY_RUN:-0}" -eq 1 ]; then
      log_summary "Runtime" "Kotlin" "⚖️ Previewed" "-" "0"
    else
      local _STAT_KOTLIN_RT="✅ Installed"
      install_runtime_kotlin || _STAT_KOTLIN_RT="❌ Failed"

      local _DUR_KOTLIN_RT
      _DUR_KOTLIN_RT=$(($(date +%s) - _T0_KOTLIN_RT))
      log_summary "Runtime" "Kotlin" "${_STAT_KOTLIN_RT:-}" "$(get_version kotlin -version | head -n 1)" "${_DUR_KOTLIN_RT:-}"
    fi
  fi

  # Also ensure linting tools are present
  install_ktlint
}
# Purpose: Checks if Kotlin runtime is available.
# Examples:
#   check_runtime_kotlin "Linter"
check_runtime_kotlin() {
  local _TOOL_DESC_KOTLIN="${1:-Kotlin}"
  if ! resolve_bin "kotlin" >/dev/null 2>&1; then
    log_warn "Required runtime 'kotlin' for ${_TOOL_DESC_KOTLIN:-} is missing. Skipping."
    return 1
  fi
  return 0
}
