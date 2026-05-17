#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Swift Logic Module

# Purpose: Installs Swift runtime via unirtm.
# Delegate: Managed by unirtm (.unirtm.toml)
install_runtime_swift() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install Swift runtime."
    return 0
  fi
  # shellcheck disable=SC2154
  unirtm install "swift@$(get_unirtm_tool_version swift)"
}

# Purpose: Sets up Swift runtime and mandatory linting tools.
# shellcheck disable=SC2329
setup_swift() {
  if ! has_lang_files "Package.swift" "*.swift"; then
    return 0
  fi

  setup_registry_swift

  local _T0_SWIFT_RT
  _T0_SWIFT_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version swift)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "swift")

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "Swift" "✅ Detected" "${_CUR_VER:-}" "0"
  else
    _log_setup "Swift Runtime" "swift"

    if [ "${DRY_RUN:-0}" -eq 0 ]; then
      log_warn "Swift installation is large (~1GB+) and may take several minutes. Please wait..."
      install_runtime_swift || return 1
    fi
    log_summary "Runtime" "Swift" "✅ Installed" "$(get_version swift --version | head -n 1)" "$(($(date +%s) - _T0_SWIFT_RT))"
  fi

  # Also ensure linting tools are present
  install_swiftformat
  install_swiftlint
}

# Purpose: Installs swiftformat for Swift linting.
# Delegate: Managed by unirtm (.unirtm.toml)
install_swiftformat() {
  if ! has_lang_files "Package.swift" "*.swift"; then
    return 0
  fi

  setup_registry_swiftformat
  install_tool_safe "swiftformat" "${VER_SWIFTFORMAT_PROVIDER:-}" "SwiftFormat" "--version" 0 "*.swift" ""
}

# Purpose: Installs swiftlint for Swift linting.
# Delegate: Managed by unirtm (.unirtm.toml)
install_swiftlint() {
  if ! has_lang_files "Package.swift" "*.swift"; then
    return 0
  fi

  setup_registry_swiftlint
  install_tool_safe "swiftlint" "${VER_SWIFTLINT_PROVIDER:-}" "SwiftLint" "version" 0 "*.swift" ""
}

# Purpose: Checks if Swift runtime is available.
# Examples:
#   check_runtime_swift "Linter"
check_runtime_swift() {
  local _TOOL_DESC_SWIFT="${1:-Swift}"
  if ! resolve_bin "swift" >/dev/null 2>&1; then
    log_warn "Required runtime 'swift' for ${_TOOL_DESC_SWIFT:-} is missing. Skipping."
    return 1
  fi
  return 0
}
