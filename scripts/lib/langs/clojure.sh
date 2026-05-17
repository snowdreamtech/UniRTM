#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Clojure Logic Module

# Purpose: Installs Clojure and Leiningen via unirtm.
# Delegate: Managed by unirtm (.unirtm.toml)
install_runtime_clojure() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install Clojure and Leiningen via unirtm."
    return 0
  fi

  unirtm install clojure
  # shellcheck disable=SC2154
  unirtm install "clojure@$(get_unirtm_tool_version clojure)"
}

# Purpose: Sets up Clojure environment for project.
setup_clojure() {
  if ! has_lang_files "deps.edn project.clj build.boot" "*.clj *.cljs *.cljc *.edn"; then
    return 0
  fi

  setup_registry_clojure

  local _T0_CLJ_RT
  _T0_CLJ_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version clojure)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "clojure")

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "Clojure" "✅ Detected" "${_CUR_VER:-}" "0"
    return 0
  fi

  _log_setup "Clojure" "clojure"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_summary "Runtime" "Clojure" "⚖️ Previewed" "-" "0"
    return 0
  fi

  local _STAT_CLJ_RT="✅ Installed"
  install_runtime_clojure || _STAT_CLJ_RT="❌ Failed"

  local _DUR_CLJ_RT
  _DUR_CLJ_RT=$(($(date +%s) - _T0_CLJ_RT))
  log_summary "Runtime" "Clojure" "${_STAT_CLJ_RT:-}" "$(get_version clojure --version | head -n 1)" "${_DUR_CLJ_RT:-}"
}

# Purpose: Checks if Clojure is available.
# Examples:
#   check_runtime_clojure "Linter"
check_runtime_clojure() {
  local _TOOL_DESC_CLJ="${1:-Clojure}"
  if ! resolve_bin "clojure" >/dev/null 2>&1; then
    log_warn "Required runtime 'clojure' for ${_TOOL_DESC_CLJ:-} is missing. Skipping."
    return 1
  fi
  return 0
}
