#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# OpenTofu Logic Module

# Purpose: Installs OpenTofu via unirtm.
# Delegate: Managed by unirtm (.unirtm.toml)
install_runtime_tofu() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install OpenTofu."
    return 0
  fi
  # shellcheck disable=SC2154
  setup_registry_tofu
  unirtm install "opentofu@$(get_unirtm_tool_version opentofu)"
}

# Purpose: Sets up OpenTofu IaC.
setup_tofu() {
  if ! has_lang_files "" "HCL *.tf *.tfvars"; then
    return 0
  fi

  local _T0_TOFU
  _T0_TOFU=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version tofu)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "tofu")

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "IaC" "OpenTofu" "✅ Detected" "${_CUR_VER:-}" "0"
    return 0
  fi

  _log_setup "OpenTofu" "tofu"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "IaC" "OpenTofu" "⚖️ Previewed" "-" "0"
    return 0
  fi

  local _STAT_TO_RT="✅ Installed"
  install_runtime_tofu || _STAT_TO_RT="❌ Failed"

  local _DUR_TOFU
  _DUR_TOFU=$(($(date +%s) - _T0_TOFU))
  log_summary "IaC" "OpenTofu" "${_STAT_TO_RT:-}" "$(get_version tofu version)" "${_DUR_TOFU:-}"
}
# Purpose: Checks if OpenTofu is available.
# Examples:
#   check_runtime_tofu "Linter"
check_runtime_tofu() {
  local _TOOL_DESC_TOFU="${1:-OpenTofu}"
  if ! resolve_bin "tofu" >/dev/null 2>&1; then
    log_warn "Required runtime 'tofu' for ${_TOOL_DESC_TOFU:-} is missing. Skipping."
    return 1
  fi
  return 0
}
