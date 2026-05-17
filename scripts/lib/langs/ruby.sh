#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Ruby Logic Module

# Purpose: Installs Ruby runtime via unirtm.
# Delegate: Managed by unirtm (.unirtm.toml)
install_runtime_ruby() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install Ruby runtime."
    return 0
  fi

  # Runtime initialization
  unirtm install ruby

  # Project dependencies
  if [ -f "Gemfile" ]; then
    run_quiet bundle install
  fi
}

# Purpose: Sets up Rubocop for Ruby project linting.
# Params:
#   None
# Examples:
#   install_ruby_lint
install_ruby_lint() {
  if ! has_lang_files "Gemfile" "*.rb"; then
    return 0
  fi

  setup_registry_rubocop
  install_tool_safe "rubocop" "${VER_RUBOCOP_PROVIDER:-}" "Rubocop" "--version" 0 "*.rb" ""
}

# Purpose: Sets up Ruby runtime and mandatory linting tools.
# Delegate: Managed by unirtm (.unirtm.toml)
setup_ruby() {
  if ! has_lang_files "Gemfile Gemfile.lock Rakefile" "*.rb *.rake"; then
    return 0
  fi

  setup_registry_ruby

  local _T0_RUBY_RT
  _T0_RUBY_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version ruby)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "ruby")

  if is_version_match "${_CUR_VER:-}" "${_REQ_VER:-}"; then
    log_summary "Runtime" "Ruby" "✅ Detected" "${_CUR_VER:-}" "0"
  else
    _log_setup "Ruby Runtime" "ruby"

    if [ "${DRY_RUN:-0}" -eq 1 ]; then
      log_summary "Runtime" "Ruby" "⚖️ Previewed" "-" "0"
    else
      local _STAT_RUBY_RT="✅ Installed"
      install_runtime_ruby || _STAT_RUBY_RT="❌ Failed"

      local _DUR_RUBY_RT
      _DUR_RUBY_RT=$(($(date +%s) - _T0_RUBY_RT))
      log_summary "Runtime" "Ruby" "${_STAT_RUBY_RT:-}" "$(get_version ruby)" "${_DUR_RUBY_RT:-}"
    fi
  fi

  # Also ensure linting tools are present
  install_ruby_lint
}
# Purpose: Checks if Ruby runtime is available.
# Examples:
#   check_runtime_ruby "Linter"
check_runtime_ruby() {
  local _TOOL_DESC_RUBY="${1:-Ruby}"
  if ! resolve_bin "ruby" >/dev/null 2>&1; then
    log_warn "Required runtime 'ruby' for ${_TOOL_DESC_RUBY:-} is missing. Skipping."
    return 1
  fi
  return 0
}
