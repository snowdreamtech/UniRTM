#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Erlang Logic Module

# Purpose: Installs Erlang runtime via unirtm.
install_runtime_erlang() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install Erlang via unirtm."
    return 0
  fi

  # shellcheck disable=SC2154
  "${_G_UNIRTM_BIN:-unirtm}" install "erlang@$(get_unirtm_tool_version erlang)"
}

# Purpose: Sets up Erlang environment for project.
setup_erlang() {
  if ! has_lang_files "rebar.config erlang.mk" "*.erl *.hrl"; then
    return 0
  fi

  setup_registry_erlang

  local _T0_ERL_RT
  _T0_ERL_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version erlang)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "erlang")

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "Erlang" "✅ Detected" "${_CUR_VER:-}" "0"
    return 0
  fi

  _log_setup "Erlang" "erlang"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "Runtime" "Erlang" "⚖️ Previewed" "-" "0"
    return 0
  fi

  local _STAT_ERL_RT="✅ Installed"
  install_runtime_erlang || _STAT_ERL_RT="❌ Failed"

  local _DUR_ERL_RT
  _DUR_ERL_RT=$(($(date +%s) - _T0_ERL_RT))
  log_summary "Runtime" "Erlang" "${_STAT_ERL_RT:-}" "$(get_version erl -version | head -n 1 | awk '{print $NF}')" "${_DUR_ERL_RT:-}"
}

# Purpose: Checks if Erlang is available.
check_runtime_erlang() {
  local _TOOL_DESC_ERL="${1:-Erlang}"
  if ! resolve_bin "erl" >/dev/null 2>&1; then
    log_warn "Required runtime 'erl' for ${_TOOL_DESC_ERL:-} is missing. Skipping."
    return 1
  fi
  return 0
}
