#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Elixir Logic Module

# Purpose: Installs Elixir/Erlang runtime via unirtm.
# Delegate: Managed by unirtm (.unirtm.toml)
install_runtime_elixir() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install Elixir/Erlang runtime."
    return 0
  fi
  # shellcheck disable=SC2154
  "${_G_UNIRTM_BIN:-unirtm}" install "erlang@$(get_unirtm_tool_version erlang)"
  # shellcheck disable=SC2154
  "${_G_UNIRTM_BIN:-unirtm}" install "elixir@$(get_unirtm_tool_version elixir)"
}

# Purpose: Sets up Elixir runtime and mandatory linting tools.
setup_elixir() {
  if ! has_lang_files "mix.exs" "*.ex *.exs"; then
    return 0
  fi

  setup_registry_elixir

  local _T0_ELIXIR_RT
  _T0_ELIXIR_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version elixir)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "elixir")

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "Elixir" "✅ Detected" "${_CUR_VER:-}" "0"
    return 0
  fi

  _log_setup "Elixir Runtime" "elixir"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "Runtime" "Elixir" "⚖️ Previewed" "-" "0"
    return 0
  fi

  local _STAT_ELIXIR_RT="✅ Installed"
  install_runtime_elixir || _STAT_ELIXIR_RT="❌ Failed"

  local _DUR_ELIXIR_RT
  _DUR_ELIXIR_RT=$(($(date +%s) - _T0_ELIXIR_RT))
  log_summary "Runtime" "Elixir" "${_STAT_ELIXIR_RT:-}" "$(get_version elixir --version | grep 'Elixir' | head -n 1)" "${_DUR_ELIXIR_RT:-}"
}
# Purpose: Checks if Elixir runtime is available.
# Examples:
#   check_runtime_elixir "Linter"
check_runtime_elixir() {
  local _TOOL_DESC_ELIXIR="${1:-Elixir}"
  if ! resolve_bin "elixir" >/dev/null 2>&1; then
    log_warn "Required runtime 'elixir' for ${_TOOL_DESC_ELIXIR:-} is missing. Skipping."
    return 1
  fi
  return 0
}
