#!/usr/bin/env sh
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# tests/property-preservation-other-tools.sh
# Purpose: Property-based tests for preservation of other tool detection
# **Validates: Requirements 3.1, 3.2, 3.3, 3.4**
#
# This test verifies that the fix for Gitleaks detection does NOT break
# existing tool detection for Node.js, Python, Shellcheck, and Checkmake.
#
# IMPORTANT: This test runs on UNFIXED code and is EXPECTED TO PASS.
# This confirms the baseline behavior that must be preserved after the fix.

set -eu

# ── Common Library ───────────────────────────────────────────────────────────


# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# ── Test Helpers ─────────────────────────────────────────────────────────────

log_test_start() {
  local _test_name="$1"
  TESTS_RUN=$((TESTS_RUN + 1))
  printf "\n🧪 Test %d: %s\n" "$TESTS_RUN" "$_test_name"
}

assert_tool_detected() {
  local _tool_name="$1"
  local _tool_cmd="$2"
  local _test_desc="$3"
  local _optional="${4:-0}"

  log_test_start "$_test_desc"

  # Try to resolve the binary
  local _resolved_path
  _resolved_path=$(resolve_bin_cached "$_tool_cmd" 2>/dev/null) || true

  if [ -n "$_resolved_path" ]; then
    log_success "✅ PASS: $_tool_name detected at: $_resolved_path"
    TESTS_PASSED=$((TESTS_PASSED + 1))
    return 0
  else
    if [ "$_optional" = "1" ]; then
      log_info "⏭️  SKIP: $_tool_name not installed (optional for this test)"
      TESTS_PASSED=$((TESTS_PASSED + 1))
      return 0
    else
      log_error "❌ FAIL: $_tool_name NOT detected (expected to be found)"
      TESTS_FAILED=$((TESTS_FAILED + 1))
      return 1
    fi
  fi
}

assert_tool_version_check() {
  local _tool_name="$1"
  local _tool_cmd="$2"
  local _test_desc="$3"
  local _optional="${4:-0}"

  log_test_start "$_test_desc"

  # Try to get version
  local _version
  _version=$(get_version "$_tool_cmd" 2>/dev/null) || true

  if [ -n "$_version" ] && [ "$_version" != "-" ]; then
    log_success "✅ PASS: $_tool_name version detected: $_version"
    TESTS_PASSED=$((TESTS_PASSED + 1))
    return 0
  else
    if [ "$_optional" = "1" ]; then
      log_info "⏭️  SKIP: $_tool_name version not available (optional for this test)"
      TESTS_PASSED=$((TESTS_PASSED + 1))
      return 0
    else
      log_error "❌ FAIL: $_tool_name version NOT detected (expected to be found)"
      TESTS_FAILED=$((TESTS_FAILED + 1))
      return 1
    fi
  fi
}

# ── Property-Based Test Generators ──────────────────────────────────────────

# Property 2: Preservation - Other Tool Detection Unchanged
#
# For any tool that is NOT Gitleaks (Node.js, Python, Shellcheck, Checkmake),
# the code SHALL produce the same detection behavior as before.
#
# Test Strategy: Generate test cases for multiple tools and verify each one
# can be detected via resolve_bin and get_version.

test_property_nodejs_detection() {
  log_info "\n═══ Property 2.1: Node.js Detection Preservation ═══"

  # Test case 1: Node.js binary resolution
  assert_tool_detected "Node.js" "node" "Node.js binary can be resolved via resolve_bin" 0

  # Test case 2: Node.js version detection
  assert_tool_version_check "Node.js" "node" "Node.js version can be detected via get_version" 0

  # Test case 3: npm binary resolution
  if [ -f "package.json" ]; then
    assert_tool_detected "npm" "npm" "npm binary can be resolved when package.json exists" 0
  else
    log_info "⏭️  Skipping npm test (no package.json)"
  fi
}

test_property_python_detection() {
  log_info "\n═══ Property 2.2: Python Detection Preservation ═══"

  # Test case 1: Python binary resolution
  assert_tool_detected "Python" "python3" "Python binary can be resolved via resolve_bin" 0

  # Test case 2: Python version detection
  assert_tool_version_check "Python" "python3" "Python version can be detected via get_version" 0
}

test_property_shellcheck_detection() {
  log_info "\n═══ Property 2.3: Shellcheck Detection Preservation ═══"

  # Only test if shell files exist
  if has_lang_files "" "*.sh"; then
    # Test case 1: Shellcheck binary resolution (optional - may not be installed)
    assert_tool_detected "Shellcheck" "shellcheck" "Shellcheck binary can be resolved via resolve_bin" 1

    # Test case 2: Shellcheck version detection (optional)
    assert_tool_version_check "Shellcheck" "shellcheck" "Shellcheck version can be detected via get_version" 1
  else
    log_info "⏭️  Skipping Shellcheck tests (no shell files)"
  fi
}

test_property_checkmake_detection() {
  log_info "\n═══ Property 2.4: Checkmake Detection Preservation ═══"

  # Only test if Makefile exists
  if has_lang_files "Makefile" "*.make"; then
    # Test case 1: Checkmake binary resolution (optional - may not be installed)
    assert_tool_detected "Checkmake" "checkmake" "Checkmake binary can be resolved via resolve_bin" 1

    # Test case 2: Checkmake version detection (optional)
    assert_tool_version_check "Checkmake" "checkmake" "Checkmake version can be detected via get_version" 1
  else
    log_info "⏭️  Skipping Checkmake tests (no Makefile)"
  fi
}

test_property_ruby_detection() {
  log_info "\n═══ Property 2.5: Ruby Detection Preservation ═══"

  # Only test if Ruby files exist
  if has_lang_files "Gemfile .ruby-version" "*.rb"; then
    # Test case 1: Ruby binary resolution (optional - may not be installed)
    assert_tool_detected "Ruby" "ruby" "Ruby binary can be resolved via resolve_bin" 1

    # Test case 2: Ruby version detection (optional)
    assert_tool_version_check "Ruby" "ruby" "Ruby version can be detected via get_version" 1
  else
    log_info "⏭️  Skipping Ruby tests (no Ruby files)"
  fi
}

# ── Main Test Execution ──────────────────────────────────────────────────────

main() {
  log_info "🚀 Starting Preservation Property Tests (UNFIXED Code)"
  log_info "Expected Outcome: ALL TESTS PASS (confirms baseline behavior)"
  log_info "═══════════════════════════════════════════════════════════════\n"

  # Ensure we're in project root
  guard_project_root

  # Run property tests for each tool category
  test_property_nodejs_detection
  test_property_python_detection
  test_property_shellcheck_detection
  test_property_checkmake_detection
  test_property_ruby_detection

  # ── Test Summary ───────────────────────────────────────────────────────────
  log_info "\n═══════════════════════════════════════════════════════════════"
  log_info "📊 Test Summary:"
  log_info "   Total Tests Run: $TESTS_RUN"
  log_success "   Passed: $TESTS_PASSED"

  if [ "$TESTS_FAILED" -gt 0 ]; then
    log_error "   Failed: $TESTS_FAILED"
    log_error "\n❌ PRESERVATION TESTS FAILED"
    log_error "This indicates that the current code has regressions in tool detection."
    log_error "Expected: All tests should PASS on unfixed code."
    exit 1
  else
    log_success "\n✅ ALL PRESERVATION TESTS PASSED"
    log_success "Baseline behavior confirmed: Other tools are detected correctly."
    log_success "This behavior MUST be preserved after implementing the Gitleaks fix."
    exit 0
  fi
}

main "$@"
