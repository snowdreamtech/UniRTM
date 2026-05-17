#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Java Logic Module

# Purpose: Installs Java runtime via unirtm (version pinned in scripts/lib/versions.sh).
install_runtime_java() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install Java runtime."
    return 0
  fi
  "${_G_UNIRTM_BIN:-unirtm}" install "java@${VER_JAVA:-}"
}

# Purpose: Installs google-java-format for Java project linting (version in versions.sh).
# WARNING: google-java-format has no prebuilt binary for linux/arm64.
#          On ARM64 Linux, this step may fail. Use: java -jar google-java-format.jar
install_java_lint() {
  if ! has_lang_files "pom.xml build.gradle" "*.java"; then
    return 0
  fi

  setup_registry_google_java_format
  install_tool_safe "google-java-format" "${VER_JAVA_FORMAT_PROVIDER:-}" "Java Lint" "--version" 0 "*.java" ""
}

# Purpose: Sets up Java runtime and mandatory linting tools.
# Delegate: Managed by unirtm (.unirtm.toml)
setup_java() {
  if ! has_lang_files "pom.xml build.gradle" "*.java"; then
    return 0
  fi

  # Dynamically register Java in .unirtm.toml if not already present.
  setup_registry_java

  local _T0_JAVA_RT
  _T0_JAVA_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version java)
  local _REQ_VER="${VER_JAVA:-}"

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "Java" "✅ Detected" "${_CUR_VER:-}" "0"
  else
    _log_setup "Java Runtime" "java"

    if [ "${DRY_RUN:-0}" -eq 1 ]; then
      log_summary "Runtime" "Java" "⚖️ Previewed" "-" "0"
    else
      local _STAT_JAVA_RT="✅ Installed"
      install_runtime_java || _STAT_JAVA_RT="❌ Failed"

      local _DUR_JAVA_RT
      _DUR_JAVA_RT=$(($(date +%s) - _T0_JAVA_RT))
      log_summary "Runtime" "Java" "${_STAT_JAVA_RT:-}" "$(get_version java)" "${_DUR_JAVA_RT:-}"
    fi
  fi

  # Also ensure linting tools are present
  install_java_lint
}
# Purpose: Checks if Java runtime is available.
# Examples:
#   check_runtime_java "Linter"
check_runtime_java() {
  local _TOOL_DESC_JAVA="${1:-Java}"
  if ! resolve_bin "java" >/dev/null 2>&1; then
    log_warn "Required runtime 'java' for ${_TOOL_DESC_JAVA:-} is missing. Skipping."
    return 1
  fi
  return 0
}
