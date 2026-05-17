#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Starlark Logic Module

# Purpose: Installs Bazel (for Starlark) via unirtm.
# Delegate: Managed by unirtm (.unirtm.toml)
install_runtime_starlark() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install Bazel via mise."
    return 0
  fi

  # shellcheck disable=SC2154
  unirtm install "bazel@$(get_unirtm_tool_version bazel)"
}

# Purpose: Sets up Starlark environment for project.
setup_starlark() {
  if ! has_lang_files "BUILD WORKSPACE MODULE.bazel" "*.star *.bzl"; then
    return 0
  fi

  setup_registry_bazel

  local _T0_STAR_RT
  _T0_STAR_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version bazel)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "bazel")

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "Starlark" "✅ Detected" "${_CUR_VER:-}" "0"
    return 0
  fi

  _log_setup "Starlark" "bazel"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "Runtime" "Starlark" "⚖️ Previewed" "-" "0"
    return 0
  fi

  local _STAT_STAR_RT="✅ Installed"
  install_runtime_starlark || _STAT_STAR_RT="❌ Failed"

  local _DUR_STAR_RT
  _DUR_STAR_RT=$(($(date +%s) - _T0_STAR_RT))
  log_summary "Runtime" "Starlark" "${_STAT_STAR_RT:-}" "$(get_version bazel --version | awk '{print $NF}')" "${_DUR_STAR_RT:-}"
}

# Purpose: Checks if Bazel is available.
# Examples:
#   check_runtime_starlark "Linter"
check_runtime_starlark() {
  local _TOOL_DESC_STAR="${1:-Starlark}"
  if ! resolve_bin "bazel" >/dev/null 2>&1; then
    log_warn "Required runtime 'bazel' for ${_TOOL_DESC_STAR:-} is missing. Skipping."
    return 1
  fi
  return 0
}
