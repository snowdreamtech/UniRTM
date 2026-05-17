#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Dotnet Logic Module

# Purpose: Installs .NET runtime via unirtm.
# Delegate: Managed by unirtm (.unirtm.toml)
install_runtime_dotnet() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install .NET runtime."
    return 0
  fi

  # Runtime initialization
  unirtm install dotnet
}

# Purpose: Sets up Dotnet runtime.
setup_dotnet() {
  if ! has_lang_files "global.json *.csproj *.fsproj *.sln" ""; then
    return 0
  fi

  # Dynamically register .NET in .unirtm.toml if not already present.
  setup_registry_dotnet

  local _T0_DOTNET_RT
  _T0_DOTNET_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version dotnet)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "dotnet")

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" ".NET" "✅ Detected" "${_CUR_VER:-}" "0"
  else

    _log_setup ".NET Runtime" "dotnet"

    if [ "${DRY_RUN:-0}" -eq 0 ]; then
      install_runtime_dotnet || return 1
    fi
    log_summary "Runtime" ".NET" "✅ Installed" "$(get_version dotnet --version)" "$(($(date +%s) - _T0_DOTNET_RT))"
  fi

  # Install .NET specific tools
  install_dotnet_format
}

# Purpose: Installs dotnet-format for .NET linting.
# Delegate: Managed by unirtm (.unirtm.toml)
install_dotnet_format() {
  local _T0_DNF
  _T0_DNF=$(date +%s)
  _log_setup "Dotnet Format" "dotnet"
  local _STAT_DNF="✅ Available"
  # dotnet-format is now built-in as 'dotnet format'
  if ! dotnet format --version >/dev/null 2>&1; then
    _STAT_DNF="❌ Missing"
  fi
  log_summary "Dotnet" "Dotnet Format" "${_STAT_DNF:-}" "$(dotnet format --version 2>/dev/null || echo "-")" "$(($(date +%s) - _T0_DNF))"
}
# Purpose: Checks if .NET runtime is available.
# Examples:
#   check_runtime_dotnet "Linter"
check_runtime_dotnet() {
  local _TOOL_DESC_DOTNET="${1:-.NET}"
  if ! resolve_bin "dotnet" >/dev/null 2>&1; then
    log_warn "Required runtime 'dotnet' for ${_TOOL_DESC_DOTNET:-} is missing. Skipping."
    return 1
  fi
  return 0
}
