#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# OCaml Logic Module

# Purpose: Installs OCaml and OPAM via unirtm.
# Delegate: Managed by unirtm (.unirtm.toml)
install_runtime_ocaml() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install OCaml via unirtm."
    return 0
  fi

  # shellcheck disable=SC2154
  "${_G_UNIRTM_BIN:-unirtm}" install "ocaml@$(get_unirtm_tool_version ocaml)"
}

# Purpose: Sets up OCaml environment for project.
setup_ocaml() {
  if ! has_lang_files "dune-project opam" "*.ml *.mli"; then
    return 0
  fi

  setup_registry_ocaml

  local _T0_OCM_RT
  _T0_OCM_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version ocaml)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "ocaml")

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "OCaml" "✅ Detected" "${_CUR_VER:-}" "0"
    return 0
  fi

  _log_setup "OCaml" "ocaml"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "Runtime" "OCaml" "⚖️ Previewed" "-" "0"
    return 0
  fi

  local _STAT_OCM_RT="✅ Installed"
  install_runtime_ocaml || _STAT_OCM_RT="❌ Failed"

  local _DUR_OCM_RT
  _DUR_OCM_RT=$(($(date +%s) - _T0_OCM_RT))
  log_summary "Runtime" "OCaml" "${_STAT_OCM_RT:-}" "$(get_version ocaml --version | head -n 1)" "${_DUR_OCM_RT:-}"
}

# Purpose: Checks if OCaml is available.
# Examples:
#   check_runtime_ocaml "Linter"
check_runtime_ocaml() {
  local _TOOL_DESC_OCM="${1:-OCaml}"
  if ! resolve_bin "ocaml" >/dev/null 2>&1; then
    log_warn "Required runtime 'ocaml' for ${_TOOL_DESC_OCM:-} is missing. Skipping."
    return 1
  fi
  return 0
}
