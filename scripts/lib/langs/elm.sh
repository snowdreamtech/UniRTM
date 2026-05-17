#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Elm Logic Module

# Purpose: Installs Elm via unirtm.
# Delegate: Managed by unirtm (.unirtm.toml)
install_runtime_elm() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install Elm via unirtm."
    return 0
  fi

  # shellcheck disable=SC2154
  "${_G_UNIRTM_BIN:-unirtm}" install "elm@$(get_unirtm_tool_version elm)"
}

# Purpose: Sets up Elm environment for project.
setup_elm() {
  if ! has_lang_files "elm.json" "*.elm"; then
    return 0
  fi

  setup_registry_elm

  local _T0_ELM_RT
  _T0_ELM_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version elm)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "elm")

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "Elm" "✅ Detected" "${_CUR_VER:-}" "0"
    return 0
  fi

  _log_setup "Elm" "elm"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "Runtime" "Elm" "⚖️ Previewed" "-" "0"
    return 0
  fi

  local _STAT_ELM_RT="✅ Installed"
  install_runtime_elm || _STAT_ELM_RT="❌ Failed"

  local _DUR_ELM_RT
  _DUR_ELM_RT=$(($(date +%s) - _T0_ELM_RT))
  log_summary "Runtime" "Elm" "${_STAT_ELM_RT:-}" "$(get_version elm --version)" "${_DUR_ELM_RT:-}"
}

# Purpose: Checks if Elm is available.
# Examples:
#   check_runtime_elm "Linter"
check_runtime_elm() {
  local _TOOL_DESC_ELM="${1:-Elm}"
  if ! resolve_bin "elm" >/dev/null 2>&1; then
    log_warn "Required runtime 'elm' for ${_TOOL_DESC_ELM:-} is missing. Skipping."
    return 1
  fi
  return 0
}
