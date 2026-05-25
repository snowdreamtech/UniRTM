#!/usr/bin/env bash
# Integration Test: Dynamic Tool Installation Without .mise.toml Pollution
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# SPDX-License-Identifier: MIT
# shellcheck disable=SC2329
# shellcheck shell=bash

set -euo pipefail

# Source required modules
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"



# shellcheck source=../../scripts/lib/timeout.sh
. "$PROJECT_ROOT/scripts/lib/timeout.sh"
# shellcheck source=../../scripts/lib/registry.sh
. "$PROJECT_ROOT/scripts/lib/registry.sh"



# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Test result tracking
declare -a FAILED_TESTS=()

# Helper function to run a test
run_test() {
  local test_name="$1"
  local test_func="$2"

  TESTS_RUN=$((TESTS_RUN + 1))
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "Test $TESTS_RUN: $test_name"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

  if $test_func; then
    TESTS_PASSED=$((TESTS_PASSED + 1))
    echo "✅ PASSED: $test_name"
  else
    TESTS_FAILED=$((TESTS_FAILED + 1))
    FAILED_TESTS+=("$test_name")
    echo "❌ FAILED: $test_name"
  fi
  echo ""
}

# Test 1: .mise.toml remains clean after dynamic install
test_mise_toml_clean() {
  echo "Testing: .mise.toml remains clean after dynamic install"

  # Backup .mise.toml
  cp "$PROJECT_ROOT/.mise.toml" "$PROJECT_ROOT/.mise.toml.backup"

  # Simulate CI environment (must be exported for subshells)
  export CI=true
  export GITHUB_ACTIONS=true

  # Install Zizmor dynamically
  setup_registry_zizmor >/dev/null 2>&1

  # Check .mise.toml
  local RESULT=0
  if grep -q "pipx:zizmor" "$PROJECT_ROOT/.mise.toml"; then
    echo "✗ zizmor was added to .mise.toml"
    RESULT=1
  else
    echo "✓ .mise.toml remains clean"
    RESULT=0
  fi

  # Restore .mise.toml
  mv "$PROJECT_ROOT/.mise.toml.backup" "$PROJECT_ROOT/.mise.toml"

  # Cleanup CI vars
  unset CI
  unset GITHUB_ACTIONS

  return "$RESULT"
}

# Test 2: get_mise_tool_version works for dynamic tools
test_get_mise_tool_version() {
  echo "Testing: get_mise_tool_version for dynamic tools"

  local VERSION
  VERSION=$(get_mise_tool_version "pipx:zizmor")

  if [ "$VERSION" = "${VER_ZIZMOR:-}" ]; then
    echo "✓ get_mise_tool_version returned: $VERSION"
    return 0
  else
    echo "✗ Expected ${VER_ZIZMOR:-}, got $VERSION"
    return 1
  fi
}

# Test 3: get_version works for dynamic tools
test_get_version() {
  echo "Testing: get_version for dynamic tools"

  local VERSION
  VERSION=$(get_version zizmor)

  if [ "$VERSION" != "-" ] && [ -n "$VERSION" ]; then
    echo "✓ get_version returned: $VERSION"
    return 0
  else
    echo "✗ get_version failed to detect version"
    return 1
  fi
}

# Test 4: resolve_bin works for dynamic tools
test_resolve_bin() {
  echo "Testing: resolve_bin for dynamic tools"

  local TOOL_PATH
  TOOL_PATH=$(resolve_bin zizmor 2>&1) || true

  if [ -n "$TOOL_PATH" ] && [ -x "$TOOL_PATH" ]; then
    echo "✓ resolve_bin found: $TOOL_PATH"
    return 0
  else
    echo "✗ resolve_bin failed to find executable"
    return 1
  fi
}

# Test 5: Tool is actually executable
test_tool_executable() {
  echo "Testing: Tool is actually executable"

  if command -v zizmor >/dev/null 2>&1; then
    local VERSION_OUTPUT
    VERSION_OUTPUT=$(zizmor --version 2>&1 | head -1)
    echo "✓ zizmor is executable: $VERSION_OUTPUT"
    return 0
  else
    echo "✗ zizmor not found in PATH"
    return 1
  fi
}

# Test 6: PATH management works
test_path_management() {
  echo "Testing: PATH contains tool directories"

  local FOUND_SHIMS=0
  local FOUND_BIN=0

  if echo "$PATH" | grep -q "mise/shims"; then
    FOUND_SHIMS=1
    echo "✓ mise shims in PATH"
  fi

  if echo "$PATH" | grep -q "mise/installs.*zizmor.*bin"; then
    FOUND_BIN=1
    echo "✓ zizmor bin directory in PATH"
  fi

  if [ "$FOUND_SHIMS" -eq 1 ] || [ "$FOUND_BIN" -eq 1 ]; then
    return 0
  else
    echo "✗ Neither shims nor bin directory found in PATH"
    return 1
  fi
}

# Test 7: CI PATH persistence
test_ci_path_persistence() {
  echo "Testing: CI PATH persistence"

  # Create temporary GITHUB_PATH
  local TEMP_GITHUB_PATH
  TEMP_GITHUB_PATH=$(mktemp)
  export GITHUB_PATH="$TEMP_GITHUB_PATH"

  # Simulate CI environment
  export CI=true
  export GITHUB_ACTIONS=true

  # Force a fresh install by temporarily clearing the cache
  # This ensures run_mise actually performs installation and triggers PATH persistence
  local OLD_CACHE="${_G_MISE_LS_JSON_CACHE:-}"
  export _G_MISE_LS_JSON_CACHE=""

  # Install a test tool to trigger PATH persistence
  run_mise install pipx:zizmor@"${VER_ZIZMOR:-latest}" >/dev/null 2>&1

  # Restore cache
  export _G_MISE_LS_JSON_CACHE="$OLD_CACHE"

  # Check GITHUB_PATH
  local RESULT=0
  if [ -f "$GITHUB_PATH" ] && [ -s "$GITHUB_PATH" ]; then
    echo "✓ GITHUB_PATH was updated:"
    cat "$GITHUB_PATH" | sed 's/^/  /'
    RESULT=0
  else
    echo "✗ GITHUB_PATH not updated"
    RESULT=1
  fi

  # Cleanup
  rm -f "$TEMP_GITHUB_PATH"
  unset GITHUB_PATH
  unset CI
  unset GITHUB_ACTIONS

  return "$RESULT"
}

# Main test execution
main() {
  echo "╔════════════════════════════════════════════════════════════════════════════════╗"
  echo "║  Integration Test: Dynamic Tool Installation                                  ║"
  echo "║  Testing dynamic install without .mise.toml pollution                          ║"
  echo "╚════════════════════════════════════════════════════════════════════════════════╝"
  echo ""

  # Run all tests
  run_test "Test 1: .mise.toml remains clean" test_mise_toml_clean
  run_test "Test 2: get_mise_tool_version works" test_get_mise_tool_version
  run_test "Test 3: get_version works" test_get_version
  run_test "Test 4: resolve_bin works" test_resolve_bin
  run_test "Test 5: Tool is executable" test_tool_executable
  run_test "Test 6: PATH management works" test_path_management
  run_test "Test 7: CI PATH persistence" test_ci_path_persistence

  # Print summary
  echo "╔════════════════════════════════════════════════════════════════════════════════╗"
  echo "║  Test Summary                                                                  ║"
  echo "╚════════════════════════════════════════════════════════════════════════════════╝"
  echo "Total tests run: $TESTS_RUN"
  echo "Tests passed: $TESTS_PASSED"
  echo "Tests failed: $TESTS_FAILED"
  echo ""

  if [ "$TESTS_FAILED" -gt 0 ]; then
    echo "❌ FAILED TESTS:"
    for test in "${FAILED_TESTS[@]}"; do
      echo "  - $test"
    done
    echo ""
    exit 1
  else
    echo "✅ ALL TESTS PASSED"
    echo ""
    echo "Dynamic tool installation works correctly:"
    echo "  ✓ Tools install without polluting .mise.toml"
    echo "  ✓ All version/resolution functions work"
    echo "  ✓ PATH management is automatic"
    echo "  ✓ CI persistence works"
    exit 0
  fi
}

# Run main function
main "$@"
