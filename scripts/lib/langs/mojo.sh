#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Mojo Logic Module

# Purpose: Installs Mojo runtime via unirtm.
# Delegate: Managed by unirtm (.unirtm.toml)
install_runtime_mojo() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install Mojo runtime via unirtm."
    return 0
  fi

  # shellcheck disable=SC2154
  unirtm install "mojo@$(get_unirtm_tool_version mojo)"
}

# Purpose: Sets up Mojo environment for project.
setup_mojo() {
  if ! has_lang_files "mojoproject.toml" "*.mojo *.ipynb"; then
    return 0
  fi

  setup_registry_mojo

  local _T0_MOJ_RT
  _T0_MOJ_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version mojo)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "mojo")

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "Mojo" "✅ Detected" "${_CUR_VER:-}" "0"
    return 0
  fi

  _log_setup "Mojo" "mojo"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "Runtime" "Mojo" "⚖️ Previewed" "-" "0"
    return 0
  fi

  local _STAT_MOJ_RT="✅ Installed"
  install_runtime_mojo || _STAT_MOJ_RT="❌ Failed"

  local _DUR_MOJ_RT
  _DUR_MOJ_RT=$(($(date +%s) - _T0_MOJ_RT))
  log_summary "Runtime" "Mojo" "${_STAT_MOJ_RT:-}" "$(get_version mojo --version | head -n 1)" "${_DUR_MOJ_RT:-}"
}

# Purpose: Checks if Mojo is available.
# Examples:
#   check_runtime_mojo "Linter"
check_runtime_mojo() {
  local _TOOL_DESC_MOJ="${1:-Mojo}"
  if ! resolve_bin "mojo" >/dev/null 2>&1; then
    log_warn "Required runtime 'mojo' for ${_TOOL_DESC_MOJ:-} is missing. Skipping."
    return 1
  fi
  return 0
}
