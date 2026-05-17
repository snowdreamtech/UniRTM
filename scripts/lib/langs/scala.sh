#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Scala Logic Module

# Purpose: Installs Scala runtime via unirtm.
# Delegate: Managed by unirtm (.unirtm.toml)
install_runtime_scala() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install Scala runtime."
    return 0
  fi
  # shellcheck disable=SC2154
  unirtm install "scala@$(get_unirtm_tool_version scala)"
}

# Purpose: Installs Scala linter.
install_scala_lint() {
  if ! has_lang_files "build.sbt" "*.scala *.sc"; then
    return 0
  fi

  setup_registry_scalafmt
  install_tool_safe "scalafmt" "scalafmt" "Scala Lint" "--version" 0 "*.scala *.sc" ""
}

# Purpose: Sets up Scala runtime.
setup_scala() {
  if ! has_lang_files "build.sbt" "*.scala *.sc"; then
    return 0
  fi

  setup_registry_scala

  local _T0_SCALA_RT
  _T0_SCALA_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version scala)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "scala")

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "Scala" "✅ Detected" "${_CUR_VER:-}" "0"
    return 0
  fi

  _log_setup "Scala Runtime" "scala"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "Runtime" "Scala" "⚖️ Previewed" "-" "0"
    return 0
  fi

  local _STAT_SCALA_RT="✅ Installed"
  install_runtime_scala || _STAT_SCALA_RT="❌ Failed"

  local _DUR_SCALA_RT
  _DUR_SCALA_RT=$(($(date +%s) - _T0_SCALA_RT))
  log_summary "Runtime" "Scala" "${_STAT_SCALA_RT:-}" "$(get_version scala -version | head -n 1)" "${_DUR_SCALA_RT:-}"

  install_scala_lint
}
# Purpose: Checks if Scala runtime is available.
# Examples:
#   check_runtime_scala "Linter"
check_runtime_scala() {
  local _TOOL_DESC_SCALA="${1:-Scala}"
  if ! resolve_bin "scala" >/dev/null 2>&1; then
    log_warn "Required runtime 'scala' for ${_TOOL_DESC_SCALA:-} is missing. Skipping."
    return 1
  fi
  return 0
}
