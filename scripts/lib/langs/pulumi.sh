#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Pulumi Logic Module

# Purpose: Installs Pulumi CLI via unirtm.
# Delegate: Managed by unirtm (.unirtm.toml)
install_runtime_pulumi() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install Pulumi CLI."
    return 0
  fi
  # shellcheck disable=SC2154
  unirtm install "pulumi@$(get_unirtm_tool_version pulumi)"
}

# Purpose: Sets up Pulumi IaC.
setup_pulumi() {
  if ! has_lang_files "" "PULUMI"; then
    return 0
  fi

  local _T0_PULUMI
  _T0_PULUMI=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version pulumi)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "pulumi")

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "IaC" "Pulumi" "✅ Detected" "${_CUR_VER:-}" "0"
    return 0
  fi

  _log_setup "Pulumi" "pulumi"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "IaC" "Pulumi" "⚖️ Previewed" "-" "0"
    return 0
  fi

  local _STAT_PULUMI_RT="✅ Installed"
  install_runtime_pulumi || _STAT_PULUMI_RT="❌ Failed"

  local _DUR_PULUMI_RT
  _DUR_PULUMI_RT=$(($(date +%s) - _T0_PULUMI))
  log_summary "IaC" "Pulumi" "${_STAT_PULUMI_RT:-}" "$(get_version pulumi version)" "${_DUR_PULUMI_RT:-}"
}
# Purpose: Checks if Pulumi CLI is available.
# Examples:
#   check_runtime_pulumi "Linter"
check_runtime_pulumi() {
  local _TOOL_DESC_PULUMI="${1:-Pulumi}"
  if ! resolve_bin "pulumi" >/dev/null 2>&1; then
    log_warn "Required runtime 'pulumi' for ${_TOOL_DESC_PULUMI:-} is missing. Skipping."
    return 1
  fi
  return 0
}
