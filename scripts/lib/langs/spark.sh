#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Apache Spark Logic Module

# Purpose: Installs Apache Spark via unirtm.
# Delegate: Managed by unirtm (.unirtm.toml)
install_runtime_spark() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install Apache Spark via unirtm."
    return 0
  fi

  # shellcheck disable=SC2154
  unirtm install "spark@$(get_unirtm_tool_version spark)"
}

# Purpose: Sets up Apache Spark environment for project.
setup_spark() {
  if ! has_lang_files "spark-submit" "*.scala *.ipynb"; then
    return 0
  fi

  setup_registry_spark

  local _T0_SPARK_RT
  _T0_SPARK_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version spark-shell)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "spark-shell")

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "Apache Spark" "✅ Detected" "${_CUR_VER:-}" "0"
    return 0
  fi

  _log_setup "Apache Spark" "spark-shell"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "Runtime" "Apache Spark" "⚖️ Previewed" "-" "0"
    return 0
  fi

  local _STAT_SPARK_RT="✅ Installed"
  install_runtime_spark || _STAT_SPARK_RT="❌ Failed"

  local _DUR_SPARK_RT
  _DUR_SPARK_RT=$(($(date +%s) - _T0_SPARK_RT))
  log_summary "Runtime" "Apache Spark" "${_STAT_SPARK_RT:-}" "$(get_version spark-shell --version 2>&1 | grep "version" | head -1 | awk '{print $NF}')" "${_DUR_SPARK_RT:-}"
}

# Purpose: Checks if Apache Spark is available.
# Examples:
#   check_runtime_spark "Linter"
check_runtime_spark() {
  local _TOOL_DESC_SPARK="${1:-Apache Spark}"
  if ! resolve_bin "spark-shell" >/dev/null 2>&1; then
    log_warn "Required tool 'spark-shell' for ${_TOOL_DESC_SPARK:-} is missing. Skipping."
    return 1
  fi
  return 0
}
