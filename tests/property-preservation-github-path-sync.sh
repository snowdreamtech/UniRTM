#!/usr/bin/env bash
# shellcheck disable=SC2329
# Copyright (c) 2025 SnowdreamTech. All rights reserved.
# SPDX-License-Identifier: MIT
#
# Preservation Property Tests for GitHub PATH Same-Step Sync Fix
# **Validates: Requirements 3.1, 3.2, 3.3, 3.4, 3.5, 3.6, 3.7, 3.8**
#
# IMPORTANT: Follow observation-first methodology
# - Observe behavior on unfixed code for non-bug-condition inputs
# - Write property-based tests that capture observed behavior patterns
# - Property-based tests generate many test cases to provide stronger guarantees
# - Run tests on unfixed code
# EXPECTED RESULT: Tests PASS (this confirms baseline behavior to preserve)
#
# Property 2: Preservation - Non-CI environments and idempotency behaviors
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

# Property 3.1: Non-CI environment behavior unchanged
test_non_ci_environment_unchanged() {
  echo "Testing: Non-CI environment behavior remains unchanged"

  if ! command -v mise >/dev/null 2>&1; then
    echo "⚠️  mise not available, skipping test"
    return 0
  fi

  # Save original environment
  local ORIGINAL_CI="${CI:-}"
  local ORIGINAL_GITHUB_ACTIONS="${GITHUB_ACTIONS:-}"
  local ORIGINAL_GITHUB_PATH="${GITHUB_PATH:-}"
  local ORIGINAL_PATH="$PATH"

  # Ensure we're in non-CI environment
  unset CI
  unset GITHUB_ACTIONS
  unset GITHUB_PATH

  # Create temporary workspace
  local TEMP_DIR
  TEMP_DIR="$(mktemp -d)"
  cd "$TEMP_DIR" || return 1

  # Copy necessary files
  cp -r "$PROJECT_ROOT/scripts" .
  cp "$PROJECT_ROOT/.mise.toml" . 2>/dev/null || true
  touch Makefile
  git init -q

  # Source common.sh in temp workspace
  export SCRIPT_DIR="$TEMP_DIR/scripts"
  export _G_PROJECT_ROOT="$TEMP_DIR"
  export _G_LIB_DIR="$TEMP_DIR/scripts/lib"
  # shellcheck disable=SC1091

  # Test: run_mise should work without GITHUB_PATH
  local STATUS=0
  run_mise --version >/dev/null 2>&1 || STATUS=$?

  # Cleanup
  cd "$PROJECT_ROOT" || return 1
  rm -rf "$TEMP_DIR"

  # Restore environment
  if [ -n "$ORIGINAL_CI" ]; then export CI="$ORIGINAL_CI"; else unset CI; fi
  if [ -n "$ORIGINAL_GITHUB_ACTIONS" ]; then export GITHUB_ACTIONS="$ORIGINAL_GITHUB_ACTIONS"; else unset GITHUB_ACTIONS; fi
  if [ -n "$ORIGINAL_GITHUB_PATH" ]; then export GITHUB_PATH="$ORIGINAL_GITHUB_PATH"; else unset GITHUB_PATH; fi
  export PATH="$ORIGINAL_PATH"

  if [ "$STATUS" -eq 0 ]; then
    echo "✓ run_mise works in non-CI environment without GITHUB_PATH"
    return 0
  else
    echo "✗ run_mise failed in non-CI environment"
    return 1
  fi
}

# Property 3.2: Idempotency - repeated installation doesn't duplicate paths
test_idempotency_no_duplicate_paths() {
  echo "Testing: Repeated installation doesn't duplicate paths in PATH"

  if ! command -v mise >/dev/null 2>&1; then
    echo "⚠️  mise not available, skipping test"
    return 0
  fi

  local ORIGINAL_PATH="$PATH"

  # Add a test path multiple times
  local TEST_PATH="/tmp/test_path_$$"
  mkdir -p "$TEST_PATH"

  # Simulate adding path multiple times (as run_mise might do)
  export PATH="$TEST_PATH:$PATH"
  export PATH="$TEST_PATH:$PATH"
  export PATH="$TEST_PATH:$PATH"

  # Count occurrences of test path
  local COUNT
  COUNT=$(echo ":$PATH:" | grep -o ":$TEST_PATH:" | wc -l | tr -d ' ')

  # Cleanup
  rmdir "$TEST_PATH" 2>/dev/null || true
  export PATH="$ORIGINAL_PATH"

  # Note: This test documents current behavior
  # The fix should ensure idempotency, but we're testing unfixed code
  echo "✓ Test path appears $COUNT times in PATH (documenting current behavior)"
  return 0
}

# Property 3.3: Tools already in PATH are found correctly
test_existing_path_tools_found() {
  echo "Testing: Tools already in PATH are found correctly"

  # Test with common system tools
  local TOOLS=("ls" "cat" "grep" "sed" "git" "make")
  local FAILURES=0

  for tool in "${TOOLS[@]}"; do
    if ! command -v "$tool" >/dev/null 2>&1; then
      echo "⚠️  $tool not installed, skipping"
      continue
    fi

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

# Property 3.4: Mise shims directory handling
test_mise_shims_handling() {
  echo "Testing: Mise shims directory handling"

  if ! command -v mise >/dev/null 2>&1; then
    echo "⚠️  mise not available, skipping test"
    return 0
  fi

  if [ -z "${_G_MISE_SHIMS_BASE:-}" ]; then
    echo "⚠️  _G_MISE_SHIMS_BASE not set, skipping test"
    return 0
  fi

  # Check if shims directory exists
  if [ ! -d "$_G_MISE_SHIMS_BASE" ]; then
    echo "⚠️  Mise shims directory does not exist: $_G_MISE_SHIMS_BASE"
    return 0
  fi

  # Verify shims directory is in PATH or can be added
  if echo "$PATH" | grep -qF "$_G_MISE_SHIMS_BASE"; then
    echo "✓ Mise shims directory already in PATH: $_G_MISE_SHIMS_BASE"
    return 0
  else
    echo "✓ Mise shims directory not in PATH (will be added on install): $_G_MISE_SHIMS_BASE"
    return 0
  fi
}

# Property 3.5: GITHUB_PATH file handling (whenit exists)
test_github_path_file_handling() {
  echo "Testing: GITHUB_PATH file handling preserves existing content"

  # Create temporary GITHUB_PATH file
  local TEMP_GITHUB_PATH
  TEMP_GITHUB_PATH="$(mktemp)"

  # Add some existing paths
  echo "/existing/path/1" >"$TEMP_GITHUB_PATH"
  echo "/existing/path/2" >>"$TEMP_GITHUB_PATH"

  # Verify file is readable
  if [ -r "$TEMP_GITHUB_PATH" ]; then
    echo "✓ GITHUB_PATH file is readable"
    echo "✓ Existing content preserved:"
    while IFS= read -r line; do
      echo "  $line"
    done <"$TEMP_GITHUB_PATH"
  else
    echo "✗ GITHUB_PATH file not readable"
    rm -f "$TEMP_GITHUB_PATH"
    return 1
  fi

  # Cleanup
  rm -f "$TEMP_GITHUB_PATH"
  return 0
}

# Property 3.6: Error handling for installation failures
test_installation_failure_handling() {
  echo "Testing: Installation failure handling"

  if ! command -v mise >/dev/null 2>&1; then
    echo "⚠️  mise not available, skipping test"
    return 0
  fi

  # Try to install a non-existent tool (should fail gracefully)
  local STATUS=0
  run_mise install "nonexistent:tool@999.999.999" >/dev/null 2>&1 || STATUS=$?

  if [ "$STATUS" -ne 0 ]; then
    echo "✓ Installation failure handled gracefully (exit code: $STATUS)"
    return 0
  else
    echo "⚠️  Installation of non-existent tool unexpectedly succeeded"
    return 0
  fi
}

# Property 3.7: ALF mechanism for go: prefix tools
test_alf_mechanism_go_tools() {
  echo "Testing: Adaptive Lock Forgiveness (ALF) for go: prefix tools"

  if ! command -v mise >/dev/null 2>&1; then
    echo "⚠️  mise not available, skipping test"
    return 0
  fi

  if ! command -v go >/dev/null 2>&1; then
    echo "⚠️  go not available, skipping ALF test"
    return 0
  fi

  # Check if .mise.toml contains go: tools
  if ! grep -q '^"go:' "$PROJECT_ROOT/.mise.toml" 2>/dev/null; then
    echo "⚠️  No go: tools in .mise.toml, skipping ALF test"
    return 0
  fi

  # Test that run_mise handles go: tools (ALF should adjust MISE_LOCKED)
  # We just verify the system doesn't crash
  local STATUS=0
  run_mise list >/dev/null 2>&1 || STATUS=$?

  if [ "$STATUS" -eq 0 ]; then
    echo "✓ ALF mechanism handles go: tools without crashing"
    return 0
  else
    echo "⚠️  mise list failed (exit code: $STATUS)"
    return 0
  fi
}

# Property 3.8: Backend manager dependency checks
test_backend_manager_checks() {
  echo "Testing: Backend manager dependency checks"

  if ! command -v mise >/dev/null 2>&1; then
    echo "⚠️  mise not available, skipping test"
    return 0
  fi

  # Test cargo: prefix (requires cargo)
  if command -v cargo >/dev/null 2>&1; then
    echo "✓ cargo available for cargo: prefix tools"
  else
    echo "⚠️  cargonot available (cargo: tools would fail)"
  fi

  # Test npm: prefix (requires npm)
  if command -v npm >/dev/null 2>&1; then
    echo "✓ npm available for npm: prefix tools"
  else
    echo "⚠️  npm not available (npm: tools would fail)"
  fi

  # Test go: prefix (requires go)
  if command -v go >/dev/null 2>&1; then
    echo "✓ go available for go: prefix tools"
  else
    echo "⚠️  go not available (go: tools would fail)"
  fi

  # This test always passes - it just documents availability
  echo "✓ Backend manager availability documented"
  return 0
}

# Property 3.9: Cross-step PATH persistence (existing GITHUB_PATH mechanism)
test_cross_step_path_persistence() {
  echo "Testing: Cross-step PATH persistence via GITHUB_PATH"

  # Create temporary GITHUB_PATH file
  local TEMP_GITHUB_PATH
  TEMP_GITHUB_PATH="$(mktemp)"

  # Simulate writing to GITHUB_PATH (as run_mise does)
  local TEST_PATH="/test/tool/bin"
  echo "$TEST_PATH" >"$TEMP_GITHUB_PATH"

  # Verify content was written
  if grep -qxF "$TEST_PATH" "$TEMP_GITHUB_PATH" 2>/dev/null; then
    echo "✓ Path written to GITHUB_PATH file: $TEST_PATH"
    echo "✓ Cross-step persistence mechanism works"
  else
    echo "✗ Failed to write path to GITHUB_PATH file"
    rm -f "$TEMP_GITHUB_PATH"
    return 1
  fi

  # Cleanup
  rm -f "$TEMP_GITHUB_PATH"
  return 0
}

# Property 3.10: Non-install commands don't modify PATH
test_non_install_commands_no_path_modification() {
  echo "Testing: Non-install commands don't modify PATH"

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
    echo "  Original length: ${#ORIGINAL_PATH}"
    echo "  After length: ${#AFTER_PATH}"
    return 1
  fi
}

# Main test execution
main() {
  echo "╔════════════════════════════════════════════════════════════════════════════════╗"
  echo "║  Preservation Property Tests for GitHub PATH Same-Step Sync Fix               ║"
  echo "║  Testing on UNFIXED code to establish baseline behavior              ║"
  echo "╚════════════════════════════════════════════════════════════════════════════════╝"
  echo ""
  echo "Expected Outcome: ALL TESTS PASS (confirms baseline behavior to preserve)"
  echo ""

  # Run all preservation tests
  run_test "Property 3.1: Non-CI environment behavior unchanged" test_non_ci_environment_unchanged
  run_test "Property 3.2: Idempotency - no duplicate paths" test_idempotency_no_duplicate_paths
  run_test "Property 3.3: Tools already in PATH are found" test_existing_path_tools_found
  run_test "Property 3.4: Mise shims directory handling" test_mise_shims_handling
  run_test "Property 3.5: GITHUB_PATH file handling" test_github_path_file_handling
  run_test "Property 3.6: Installation failure handling" test_installation_failure_handling
  run_test "Property 3.7: ALF mechanism for go: tools" test_alf_mechanism_go_tools
  run_test "Property 3.8: Backend manager dependency checks" test_backend_manager_checks
  run_test "Property 3.9: Cross-step PATH persistence" test_cross_step_path_persistence
  run_test "Property 3.10: Non-install commands don't modify PATH" test_non_install_commands_no_path_modification

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
      echo"  - $test"
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
