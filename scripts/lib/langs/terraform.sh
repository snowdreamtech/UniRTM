#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Terraform Logic Module

# Purpose: Installs Terraform via unirtm.
# Delegate: Managed by unirtm (.unirtm.toml)
install_runtime_terraform() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install Terraform via unirtm."
    return 0
  fi

  # shellcheck disable=SC2154
  unirtm install "terraform@$(get_unirtm_tool_version terraform)"
}

# Purpose: Installs TFLint.
# Delegate: Managed by unirtm (.unirtm.toml)
install_tflint() {
  setup_registry_tflint
  install_tool_safe "tflint" "${VER_TFLINT_PROVIDER:-}" "TFLint" "--version" 0 "*.tf" ""
}

# Purpose: Sets up Terraform environment for project.
setup_terraform() {
  if ! has_lang_files "" "*.tf *.tfvars *.hcl"; then
    return 0
  fi

  local _T0_TF_RT
  _T0_TF_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version terraform)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "terraform")

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "Terraform" "✅ Detected" "${_CUR_VER:-}" "0"
  else
    _log_setup "Terraform" "terraform"

    if [ "${DRY_RUN:-0}" -eq 1 ]; then
      log_summary "Runtime" "Terraform" "⚖️ Previewed" "-" "0"
    else
      local _STAT_TF_RT="✅ Installed"
      install_runtime_terraform || _STAT_TF_RT="❌ Failed"

      local _DUR_TF_RT
      _DUR_TF_RT=$(($(date +%s) - _T0_TF_RT))
      log_summary "Runtime" "Terraform" "${_STAT_TF_RT:-}" "$(get_version terraform)" "${_DUR_TF_RT:-}"
    fi
  fi

  # Setup related tools
  install_tflint
}

# Purpose: Checks if Terraform is available.
# Examples:
#   check_runtime_terraform "Linter"
check_runtime_terraform() {
  local _TOOL_DESC_TF="${1:-Terraform}"
  if ! resolve_bin "terraform" >/dev/null 2>&1; then
    log_warn "Required runtime 'terraform' for ${_TOOL_DESC_TF:-} is missing. Skipping."
    return 1
  fi
  return 0
}
