#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Rust Logic Module

# Purpose: Installs Rust runtime via unirtm.
# Delegate: Managed by unirtm (.unirtm.toml)
install_runtime_rust() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install Rust runtime."
    return 0
  fi

  # Runtime initialization
  "${_G_UNIRTM_BIN:-unirtm}" install rust
}

# Purpose: Sets up Rust runtime for project.
# Delegate: Managed by unirtm (.unirtm.toml)
setup_rust() {
  if ! has_lang_files "Cargo.toml Cargo.lock" "*.rs"; then
    return 0
  fi

  # Dynamically register Rust in .unirtm.toml if not already present.
  setup_registry_rust

  local _T0_RUST_RT
  _T0_RUST_RT=$(date +%s)
  local _TITLE="Rust Runtime"
  local _PROVIDER="rust"

  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version rust)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "${_PROVIDER:-}")

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "Rust" "✅ Detected" "${_CUR_VER:-}" "0"
    return 0
  fi

  _log_setup "${_TITLE:-}" "${_PROVIDER:-}"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "Runtime" "Rust" "⚖️ Previewed" "-" "0"
    return 0
  fi

  local _STAT_RUST_RT="✅ Installed"
  install_runtime_rust || _STAT_RUST_RT="❌ Failed"

  local _DUR_RUST_RT
  _DUR_RUST_RT=$(($(date +%s) - _T0_RUST_RT))
  log_summary "Runtime" "Rust" "${_STAT_RUST_RT:-}" "$(get_version rustc)" "${_DUR_RUST_RT:-}"
}
# Purpose: Checks if Rust runtime is available.
# Examples:
#   check_runtime_rust "Linter"
check_runtime_rust() {
  local _TOOL_DESC_RUST="${1:-Rust}"
  if ! resolve_bin "cargo" >/dev/null 2>&1; then
    log_warn "Required runtime 'rust' for ${_TOOL_DESC_RUST:-} is missing. Skipping."
    return 1
  fi
  return 0
}
