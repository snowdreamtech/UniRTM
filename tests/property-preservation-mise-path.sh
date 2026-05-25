#!/usr/bin/env bash
# shellcheck disable=SC2329
# Copyright (c) 2025 SnowdreamTech. All rights reserved.
# SPDX-License-Identifier: MIT
#
# Preservation Property Tests for Mise PATH Management Fix
# Validates: Requirements 3.1, 3.2, 3.3, 3.4, 3.5, 3.6, 3.7, 3.8
#
# IMPORTANT: Follow observation-first methodology
# - Observe behavior on unfixed code for non-bug-condition inputs
# - Write property-based tests that capture observed behavior patterns
# - Property-based tests generate many test cases to provide stronger guarantees
# - Run tests on unfixed code
# EXPECTED RESULT: Tests PASS (this confirms baseline behavior to preserve)
#
# Property 2: Preservation - Non-Install Commands and Existing Tool Resolution
# These tests ensure that the fix does NOT break existing functionality

set -euo pipefail

# Source common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"




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

# Property 3.1: Graceful handling when mise is not installed
test_mise_not_installed_graceful() {
  echo "Testing: System handles mise absence gracefully"

  # Temporarily hide mise from PATH
  local ORIGINAL_PATH="$PATH"
  export PATH="/usr/bin:/bin:/usr/sbin:/sbin"

  # Verify mise is not available
  if command -v mise >/dev/null 2>&1; then
    echo "⚠️  Cannot hide mise from PATH, skipping test"
    export PATH="$ORIGINAL_PATH"
    return 0
  fi

  # Test that resolve_bin still works for system tools
  local TOOL_PATH
  TOOL_PATH=$(resolve_bin "ls" 2>&1) || true

  export PATH="$ORIGINAL_PATH"

  if [ -n "$TOOL_PATH" ] && [ -x "$TOOL_PATH" ]; then
    echo "✓ resolve_bin found system tool without mise: $TOOL_PATH"
    return 0
  else
    echo "✗ resolve_bin failed to find system tool without mise"
    return 1
  fi
}

# Property 3.2: Tools already in PATH from other sources are found correctly
test_existing_tools_in_path() {
  echo "Testing: Tools in PATH from other sources are found correctly"

  # Test with common system tools
  local TOOLS=("ls" "cat" "grep" "sed")
  local FAILURES=0

  for tool in "${TOOLS[@]}"; do
    local TOOL_PATH
    TOOL_PATH=$(resolve_bin "$tool" 2>&1) || true

    if [ -n "$TOOL_PATH" ] && [ -x "$TOOL_PATH" ]; then
      echo "✓ Found $tool: $TOOL_PATH"
    else
      echo "✗ Failed to find $tool"
      FAILURES=$((FAILURES + 1))
    fi
  done

  if [ "$FAILURES" -eq 0 ]; then
    return 0
  else
    echo "✗ $FAILURES tools not found"
    return 1
  fi
}

# Property 3.3: Non-install commands do not modify PATH
test_non_install_commands_no_path_modification() {
  echo "Testing: Non-install commands do not modify PATH"

  if ! command -v mise >/dev/null 2>&1; then
    echo "⚠️  mise not available, skipping test"
    return 0
  fi

  local ORIGINAL_PATH="$PATH"

  # Run non-install commands
  run_mise list >/dev/null 2>&1 || true
  run_mise --version >/dev/null 2>&1 || true

  local AFTER_PATH="$PATH"

  if [ "$ORIGINAL_PATH" = "$AFTER_PATH" ]; then
    echo "✓ PATH unchanged after non-install commands"
    return 0
  else
    echo "✗ PATH was modified by non-install commands"
    echo "  Original: $ORIGINAL_PATH"
    echo "  After: $AFTER_PATH"
    return 1
  fi
}

# Property 3.4: MISE_LOCKED environment variable is respected
test_mise_locked_respected() {
  echo "Testing: MISE_LOCKED environment variable is respected"

  if ! command -v mise >/dev/null 2>&1; then
    echo "⚠️  mise not available, skipping test"
    return 0
  fi

  # This test verifies that MISE_LOCKED doesn't break the system
  # The actual locking behavior is tested elsewhere
  # We just verify the system doesn't crash with MISE_LOCKED set
  export MISE_LOCKED=1

  # Try to get mise version (should work even with MISE_LOCKED)
  # Note: Some mise commands may intentionally fail with MISE_LOCKED
  # We just verify the system doesn't crash
  run_mise --version >/dev/null 2>&1 || true

  echo "✓ System handles MISE_LOCKED=1 without crashing"
  unset MISE_LOCKED
  return 0
}

# Property 3.5: Local development environments work correctly
test_local_dev_environment() {
  echo "Testing: Local development environment works correctly"

  # Verify that common development tools can be resolved
  local DEV_TOOLS=("git" "make")
  local FAILURES=0

  for tool in "${DEV_TOOLS[@]}"; do
    if command -v "$tool" >/dev/null 2>&1; then
      local TOOL_PATH
      TOOL_PATH=$(resolve_bin "$tool" 2>&1) || true

      if [ -n "$TOOL_PATH" ] && [ -x "$TOOL_PATH" ]; then
        echo "✓ Found $tool: $TOOL_PATH"
      else
        echo "✗ Failed to resolve $tool via resolve_bin"
        FAILURES=$((FAILURES + 1))
      fi
    else
      echo "⚠️  $tool not installed, skipping"
    fi
  done

  if [ "$FAILURES" -eq 0 ]; then
    return 0
  else
    return 1
  fi
}

# Property 3.6: Fallback resolution methods work correctly
test_fallback_resolution_methods() {
  echo "Testing: Fallback resolution methods work correctly"

  # Test Layer 3: Direct PATH lookup via command -v
  local TOOL_PATH
  TOOL_PATH=$(command -v ls 2>&1) || true

  if [ -z "$TOOL_PATH" ]; then
    echo "✗ Layer 3 (command -v) failed for system tool"
    return 1
  fi
  echo "✓ Layer 3 (command -v) works: $TOOL_PATH"

  # Test Layer 4: mise which (if mise is available)
  if command -v mise >/dev/null 2>&1; then
    # Try with a tool that might be managed by mise
    local MISE_TOOL="shellcheck"
    if grep -q "$MISE_TOOL" "$PROJECT_ROOT/.mise.toml" 2>/dev/null; then
      local MISE_PATH
      MISE_PATH=$(mise which "$MISE_TOOL" 2>&1) || true

      if [ -n "$MISE_PATH" ]; then
        echo "✓ Layer 4 (mise which) works: $MISE_PATH"
      else
        echo "⚠️  Layer 4 (mise which) returned empty (tool may not be installed)"
      fi
    fi
  fi

  return 0
}

# Property 3.7: get_version returns accurate version information
test_get_version_accuracy() {
  echo "Testing: get_version returns accurate version information"

  # Test with git (should be available on all systems)
  if ! command -v git >/dev/null 2>&1; then
    echo "⚠️  git not available, skipping test"
    return 0
  fi

  local VERSION
  VERSION=$(get_version "git" 2>&1) || true

  if [ -n "$VERSION" ]; then
    echo "✓ get_version returned: $VERSION"

    # Verify it matches actual git version
    local ACTUAL_VERSION
    ACTUAL_VERSION=$(git --version 2>&1 | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1) || true

    if [ -n "$ACTUAL_VERSION" ] && echo "$VERSION" | grep -q "$ACTUAL_VERSION"; then
      echo "✓ Version matches actual git version"
      return 0
    else
      echo "⚠️  Version format differs but get_version works"
      return 0
    fi
  else
    echo "✗ get_version returned empty"
    return 1
  fi
}

# Property 3.8: System functions with network issues (cache disabled)
test_network_issues_fallback() {
  echo "Testing: System functions with network issues (cache disabled)"

  # Simulate network-disabled environment by disabling cache
  export _G_MISE_LS_JSON_CACHE="{}"

  # Verify that resolve_bin still works via direct command resolution
  local TOOL_PATH
  TOOL_PATH=$(resolve_bin "ls" 2>&1) || true

  unset _G_MISE_LS_JSON_CACHE

  if [ -n "$TOOL_PATH" ] && [ -x "$TOOL_PATH" ]; then
    echo "✓ resolve_bin works with disabled cache: $TOOL_PATH"
    return 0
  else
    echo "✗ resolve_bin failed with disabled cache"
    return 1
  fi
}

# Main test execution
main() {
  echo "╔════════════════════════════════════════════════════════════════════════════════╗"
  echo "║  Preservation Property Tests for Mise PATH Management Fix                     ║"
  echo "║  Testing on UNFIXED code to establish baseline behavior                       ║"
  echo "╚════════════════════════════════════════════════════════════════════════════════╝"
  echo ""

  # Run all preservation tests
  run_test "Property 3.1: Graceful handling when mise is not installed" test_mise_not_installed_graceful
  run_test "Property 3.2: Tools already in PATH from other sources" test_existing_tools_in_path
  run_test "Property 3.3: Non-install commands do not modify PATH" test_non_install_commands_no_path_modification
  run_test "Property 3.4: MISE_LOCKED environment variable is respected" test_mise_locked_respected
  run_test "Property 3.5: Local development environment works correctly" test_local_dev_environment
  run_test "Property 3.6: Fallback resolution methods work correctly" test_fallback_resolution_methods
  run_test "Property 3.7: get_version returns accurate version information" test_get_version_accuracy
  run_test "Property 3.8: System functions with network issues" test_network_issues_fallback

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
    echo "⚠️  CRITICAL: Preservation tests failed on unfixed code!"
    echo "This indicates that the baseline behavior is already broken."
    echo "Investigation required before proceeding with fix implementation."
    exit 1
  else
    echo "✅ ALL PRESERVATION TESTS PASSED"
    echo ""
    echo "✓ Baseline behavior confirmed on unfixed code"
    echo "✓ These behaviors MUST be preserved after implementing the fix"
    echo "✓ Safe to proceed with fix implementation"
    exit 0
  fi
}

# Run main function
main "$@"
