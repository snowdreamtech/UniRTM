#!/usr/bin/env sh
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Purpose: Simulate CI environment for local testing
# Usage:
#   . scripts/simulate-ci.sh          # Enable CI mode (POSIX)
#   source scripts/simulate-ci.sh     # Enable CI mode (bash/zsh)
#   . scripts/simulate-ci.sh reset    # Disable CI mode
#
# This script sets environment variables to make the local environment
# behave like a CI environment, useful for testing CI-specific behaviors.

set -eu

# Color codes for output (POSIX compatible)
# shellcheck disable=SC2034
if [ -t 1 ]; then
  BLUE=$(printf '\033[0;34m')
  GREEN=$(printf '\033[0;32m')
  YELLOW=$(printf '\033[1;33m')
  RED=$(printf '\033[0;31m')
  NC=$(printf '\033[0m')
else
  BLUE="" GREEN="" YELLOW="" RED="" NC=""
fi

# Function to enable CI mode
enable_ci_mode() {
  printf '%s🔧 Enabling CI simulation mode...%s\n' "${BLUE}" "${NC}"

  # Core CI detection variables
  CI=true
  GITHUB_ACTIONS=true

  # GitHub Actions specific variables
  GITHUB_WORKFLOW="Local CI Simulation"
  GITHUB_RUN_ID="local-$(date +%s)"
  GITHUB_RUN_NUMBER="1"
  GITHUB_ACTOR="${USER:-local-user}"
  GITHUB_REPOSITORY="local/simulation"
  GITHUB_REF="refs/heads/main"
  GITHUB_SHA="$(git rev-parse HEAD 2>/dev/null || echo 'local-sha')"

  # CI behavior flags
  UNIRTM_YES=true
  UNIRTM_NON_INTERACTIVE=true
  UNIRTM_QUIET=true

  # Force cache refresh in CI
  _G_IS_CI=1

  # Create CI step summary file
  GITHUB_STEP_SUMMARY="${PWD}/.ci_summary.log"
  : >"${GITHUB_STEP_SUMMARY}"

  # Export all variables (POSIX compatible)
  export CI GITHUB_ACTIONS GITHUB_WORKFLOW GITHUB_RUN_ID GITHUB_RUN_NUMBER
  export GITHUB_ACTOR GITHUB_REPOSITORY GITHUB_REF GITHUB_SHA
  export UNIRTM_YES UNIRTM_NON_INTERACTIVE UNIRTM_QUIET
  export _G_IS_CI GITHUB_STEP_SUMMARY

  printf '%s✅ CI mode enabled%s\n' "${GREEN}" "${NC}"
  printf '\n'
  printf '%sActive CI variables:%s\n' "${YELLOW}" "${NC}"
  printf '  CI=%s\n' "${CI}"
  printf '  GITHUB_ACTIONS=%s\n' "${GITHUB_ACTIONS}"
  printf '  GITHUB_WORKFLOW=%s\n' "${GITHUB_WORKFLOW}"
  printf '  GITHUB_STEP_SUMMARY=%s\n' "${GITHUB_STEP_SUMMARY}"
  printf '  _G_IS_CI=%s\n' "${_G_IS_CI}"
  printf '\n'
  printf '%s💡 Tip: Run '\''make verify'\'' or '\''make audit'\'' to test CI behavior%s\n' "${BLUE}" "${NC}"
  printf '%s💡 Tip: Run '\''. scripts/simulate-ci.sh reset'\'' to disable%s\n' "${BLUE}" "${NC}"
}

# Function to disable CI mode
disable_ci_mode() {
  printf '%s🔧 Disabling CI simulation mode...%s\n' "${BLUE}" "${NC}"

  # Unset CI variables
  unset CI || true
  unset GITHUB_ACTIONS || true
  unset GITHUB_WORKFLOW || true
  unset GITHUB_RUN_ID || true
  unset GITHUB_RUN_NUMBER || true
  unset GITHUB_ACTOR || true
  unset GITHUB_REPOSITORY || true
  unset GITHUB_REF || true
  unset GITHUB_SHA || true
  unset _G_IS_CI || true
  unset GITHUB_STEP_SUMMARY || true

  printf '%s✅ CI mode disabled (back to local mode)%s\n' "${GREEN}" "${NC}"
}

# Main logic
case "${1:-enable}" in
reset | disable | off)
  disable_ci_mode
  ;;
enable | on | *)
  enable_ci_mode
  ;;
esac
