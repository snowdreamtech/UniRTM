#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Testing Logic Module

# Purpose: Installs bats (Bash Automated Testing System).
# Delegate: Managed by unirtm (.unirtm.toml)
install_bats() {
  install_tool_safe "bats" "${VER_BATS_PROVIDER:-npm:bats}" "Bats" "--version" 0 "*.bats" ""
}

# Purpose: Installs bats-libs (helper libraries for bats).
install_bats_libs() {
  if ! has_lang_files "" "*.bats"; then
    return 0
  fi
  log_summary "Testing" "Bats-Libs" "✅ Active" "-" "0"
}

# Purpose: Checks for Playwright deployment configurations.
install_playwright() {
  if ! has_lang_files "" "PLAYWRIGHT"; then
    return 0
  fi
  log_summary "Testing" "Playwright" "✅ Detected" "-" "0"
}

# Purpose: Checks for Cypress deployment configurations.
install_cypress() {
  if ! has_lang_files "" "CYPRESS"; then
    return 0
  fi
  log_summary "Testing" "Cypress" "✅ Detected" "-" "0"
}

# Purpose: Checks for Vitest deployment configurations.
install_vitest() {
  if ! has_lang_files "" "VITEST"; then
    return 0
  fi
  log_summary "Testing" "Vitest" "✅ Detected" "-" "0"
}

# Purpose: Sets up Testing environment.
setup_testing() {
  install_bats
  install_bats_libs
  install_playwright
  install_cypress
  install_vitest
}
