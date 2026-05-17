#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Solidity Logic Module

# Purpose: Installs Solidity compiler (solc) via unirtm.
# Delegate: Managed by unirtm (.unirtm.toml)
install_runtime_solidity() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install Solidity compiler (solc) via mise."
    return 0
  fi

  # shellcheck disable=SC2154
  unirtm install "solidity@$(get_unirtm_tool_version solidity)"
}

# Purpose: Sets up Solidity environment for project.
setup_solidity() {
  if ! has_lang_files "foundry.toml hardhat.config.js" "*.sol"; then
    return 0
  fi

  setup_registry_solc

  local _T0_SOL_RT
  _T0_SOL_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version solc)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "solc")

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "Solidity" "✅ Detected" "${_CUR_VER:-}" "0"
    return 0
  fi

  _log_setup "Solidity" "solc"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "Runtime" "Solidity" "⚖️ Previewed" "-" "0"
    return 0
  fi

  local _STAT_SOL_RT="✅ Installed"
  install_runtime_solidity || _STAT_SOL_RT="❌ Failed"

  local _DUR_SOL_RT
  _DUR_SOL_RT=$(($(date +%s) - _T0_SOL_RT))
  log_summary "Runtime" "Solidity" "${_STAT_SOL_RT:-}" "$(get_version solc)" "${_DUR_SOL_RT:-}"
}

# Purpose: Checks if Solidity compiler is available.
# Examples:
#   check_runtime_solidity "Linter"
check_runtime_solidity() {
  local _TOOL_DESC_SOL="${1:-Solidity}"
  if ! resolve_bin "solc" >/dev/null 2>&1; then
    log_warn "Required runtime 'solc' for ${_TOOL_DESC_SOL:-} is missing. Skipping."
    return 1
  fi
  return 0
}
