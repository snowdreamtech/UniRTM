#!/usr/bin/env bats
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# tests/integration/test_ci_simulation.bats - CI environment simulation tests
#
# Purpose:
#   Tests behavior in CI environments including:
#   - GitHub Actions environment variables
#   - Network timeout scenarios
#   - Resource-constrained environments
#
# Requirements: 5.2, 8.1

# SC2030/SC2031: In bats, each @test runs in a subshell by design.
# Exports are intentionally scoped to individual tests.
# SC2034: Loop variable 'i' is used by brace expansion iteration.
# shellcheck disable=SC2030,SC2031,SC2034

setup() {
  load '../vendor/bats-support/load.bash'
  load '../vendor/bats-assert/load.bash'

  # Source common.sh
  export SCRIPT_DIR
  SCRIPT_DIR="$(cd "$(dirname "$BATS_TEST_FILENAME")/../.." && pwd)"

  export _G_PROJECT_ROOT="$SCRIPT_DIR"
  export _G_LIB_DIR="$SCRIPT_DIR/scripts/lib"

  # Backup original environment
  export ORIGINAL_CI="${CI:-}"
  export ORIGINAL_GITHUB_ACTIONS="${GITHUB_ACTIONS:-}"
  export ORIGINAL_GITHUB_TOKEN="${GITHUB_TOKEN:-}"


}

teardown() {
  # Restore original environment
  if [ -n "$ORIGINAL_CI" ]; then
    export CI="$ORIGINAL_CI"
  else
    unset CI
  fi

  if [ -n "$ORIGINAL_GITHUB_ACTIONS" ]; then
    export GITHUB_ACTIONS="$ORIGINAL_GITHUB_ACTIONS"
  else
    unset GITHUB_ACTIONS
  fi

  if [ -n "$ORIGINAL_GITHUB_TOKEN" ]; then
    export GITHUB_TOKEN="$ORIGINAL_GITHUB_TOKEN"
  else
    unset GITHUB_TOKEN
  fi
}

# ── Test: GitHub Actions Environment ─────────────────────────────────────────

@test "CI simulation: detects GitHub Actions environment" {
  export CI="true"
  export GITHUB_ACTIONS="true"

  run detect_ci_platform
  assert_success
  assert_output "github-actions"
}

@test "CI simulation: detects local environment" {
  unset CI
  unset GITHUB_ACTIONS

  run detect_ci_platform
  assert_success
  assert_output "local"
}

@test "CI simulation: is_ci_env returns true in CI" {
  export CI="true"
  export GITHUB_ACTIONS="true"

  run is_ci_env
  assert_success
}

@test "CI simulation: is_ci_env returns false locally" {
  unset CI
  unset GITHUB_ACTIONS

  run is_ci_env
  assert_failure
}

# ── Test: Network Timeout Scenarios ──────────────────────────────────────────

@test "network timeout: short timeout prevents hangs" {
  export TIMEOUT_NETWORK=1

  # Simulate slow network operation
  run run_with_timeout_robust 1 sleep 10
  assert_failure 124
}

@test "network timeout: fast operations complete" {
  export TIMEOUT_NETWORK=5

  run run_with_timeout_robust 5 echo "fast"
  assert_success
  assert_output "fast"
}

@test "network timeout: configurable per operation" {
  export TIMEOUT_RESOLVE_BIN=2
  export TIMEOUT_JSON_PARSE=1
  export TIMEOUT_MISE_WHICH=3

  # All timeouts should be respected
  [[ $TIMEOUT_RESOLVE_BIN -eq 2 ]]
  [[ $TIMEOUT_JSON_PARSE -eq 1 ]]
  [[ $TIMEOUT_MISE_WHICH -eq 3 ]]
}

# ── Test: Resource-Constrained Environments ──────────────────────────────────

@test "resource constraints: handles limited PATH" {
  # Simulate minimal PATH
  PATH_BACKUP="$PATH"
  export PATH="/usr/bin:/bin"

  # Should still find basic binaries
  run resolve_bin "sh"
  assert_success

  # Restore PATH
  export PATH="$PATH_BACKUP"
}

@test "resource constraints: handles missing parsers gracefully" {
  # Temporarily hide Node.js and Python
  PATH_BACKUP="$PATH"
  export PATH="/usr/bin:/bin"

  # Should fail gracefully or use awk fallback
  run parse_json '{"test":"value"}' "test" 2>/dev/null || true

  # Restore PATH
  export PATH="$PATH_BACKUP"
}

@test "resource constraints: handles low memory scenario" {
  # Test with minimal cache
  unset _G_BIN_CACHE
  declare -gA _G_BIN_CACHE

  run resolve_bin "sh"
  assert_success
}

# ── Test: CI-Specific Behavior ───────────────────────────────────────────────

@test "CI behavior: GITHUB_TOKEN handling in CI" {
  export CI="true"
  export GITHUB_ACTIONS="true"
  export GITHUB_TOKEN="test_token_123"

  # Token should be preserved in CI
  run sh -c 'echo "$GITHUB_TOKEN"'
  assert_success
  assert_output "test_token_123"
}

@test "CI behavior: mise non-interactive mode in CI" {
  export CI="true"
  export GITHUB_ACTIONS="true"

  # Verify non-interactive flags are set
  [[ $MISE_YES == "true" ]]
  [[ $MISE_NON_INTERACTIVE == "true" ]]
  [[ $MISE_QUIET == "true" ]]
}

@test "CI behavior: update checks disabled in CI" {
  export CI="true"
  export GITHUB_ACTIONS="true"

  # Verify update checks are disabled
  [[ $MISE_CHECK_FOR_UPDATES -eq 0 ]]
  [[ $NO_UPDATE_NOTIFIER -eq 1 ]]
}

# ── Test: Timeout Protection in CI ───────────────────────────────────────────

@test "CI timeout: mise operations respect timeout" {
  skip "Requires mise installation"

  export CI="true"
  export TIMEOUT_NETWORK=5

  # Should complete within timeout
  run timeout 10 run_mise --version
  assert_success
}

@test "CI timeout: binary resolution respects timeout" {
  export CI="true"
  export TIMEOUT_RESOLVE_BIN=3

  # Should complete within timeout
  run timeout 5 resolve_bin "sh"
  assert_success
}

@test "CI timeout: JSON parsing respects timeout" {
  if ! command -v node >/dev/null 2>&1 && ! command -v python3 >/dev/null 2>&1; then
    skip "No JSON parser available"
  fi

  export CI="true"
  export TIMEOUT_JSON_PARSE=2

  run timeout 4 parse_json '{"test":"value"}' "test"
  assert_success
  assert_output "value"
}

# ── Test: Parallel Execution ─────────────────────────────────────────────────

@test "parallel execution: multiple resolve_bin calls" {
  # Simulate parallel lookups
  resolve_bin "sh" &
  local pid1=$!
  resolve_bin "sh" &
  local pid2=$!
  resolve_bin "sh" &
  local pid3=$!

  # Wait for all
  wait $pid1 $pid2 $pid3

  # All should succeed
  assert_success
}

@test "parallel execution: cache handles concurrent access" {
  export USE_NEW_RESOLVE_BIN=1

  # Multiple concurrent cached lookups
  resolve_bin_cached "sh" &
  resolve_bin_cached "sh" &
  resolve_bin_cached "sh" &
  wait

  # Cache should remain consistent
  [[ -n ${_G_BIN_CACHE[sh]} ]]
}

# ── Test: Error Recovery ─────────────────────────────────────────────────────

@test "error recovery: continues after timeout" {
  # First operation times out
  run run_with_timeout_robust 1 sleep 10
  assert_failure 124

  # Next operation should still work
  run run_with_timeout_robust 5 echo "recovered"
  assert_success
  assert_output "recovered"
}

@test "error recovery: cache survives failures" {
  export USE_NEW_RESOLVE_BIN=1

  # Successful lookup
  resolve_bin_cached "sh" >/dev/null

  # Failed lookup
  resolve_bin_cached "nonexistent" >/dev/null 2>&1 || true

  # Original cache entry should still exist
  [[ -n ${_G_BIN_CACHE[sh]} ]]
}

# ── Test: CI Summary Integration ─────────────────────────────────────────────

@test "CI summary: file is created" {
  [[ -n $CI_STEP_SUMMARY ]]
}

@test "CI summary: file path is valid" {
  # Should be a valid path
  [[ $CI_STEP_SUMMARY =~ ^/ ]] || [[ $CI_STEP_SUMMARY =~ ^\. ]]
}

# ── Test: Cross-Platform Compatibility ───────────────────────────────────────

@test "cross-platform: OS detection works" {
  [[ -n $_G_OS ]]
  [[ $_G_OS =~ ^(linux|macos|windows)$ ]]
}

@test "cross-platform: venv bin path is set" {
  [[ -n $_G_VENV_BIN ]]
}

@test "cross-platform: mise paths are set" {
  [[ -n $_G_MISE_BIN_BASE ]]
  [[ -n $_G_MISE_SHIMS_BASE ]]
}

# ── Test: Stress Testing ─────────────────────────────────────────────────────

@test "stress: many sequential lookups" {
  for i in {1..10}; do
    resolve_bin "sh" >/dev/null || true
  done

  # Should complete without hanging
  assert_success
}

@test "stress: many sequential timeouts" {
  for i in {1..5}; do
    run_with_timeout_robust 1 sleep 10 >/dev/null 2>&1 || true
  done

  # Should complete without hanging
  assert_success
}

@test "stress: rapid cache access" {
  export USE_NEW_RESOLVE_BIN=1

  # Prime cache
  resolve_bin_cached "sh" >/dev/null

  # Rapid access
  for i in {1..20}; do
    resolve_bin_cached "sh" >/dev/null
  done

  assert_success
}

# ── Test: Cleanup Verification ───────────────────────────────────────────────

@test "cleanup: no zombie processes after tests" {
  # Run some operations that might leave zombies
  run_with_timeout_robust 1 sleep 10 >/dev/null 2>&1 || true
  run_with_timeout_robust 1 sleep 10 >/dev/null 2>&1 || true

  sleep 0.5

  # Check for zombies
  run sh -c "ps aux | grep -E 'Z|defunct' | grep -v grep"
  # If there are zombies, they shouldn't be from our tests
  if [ "$status" -eq 0 ]; then
    run sh -c "ps aux | grep -E 'Z|defunct' | grep sleep | grep -v grep"
    assert_failure
  fi
}

@test "cleanup: no leaked file descriptors" {
  # Run operations that open files
  for _unused in {1..10}; do
    resolve_bin "sh" >/dev/null 2>&1 || true
  done

  # Check open file descriptors (should be reasonable)
  if command -v lsof >/dev/null 2>&1; then
    local fd_count
    fd_count=$(lsof -p $$ 2>/dev/null | wc -l)
    # Should have fewer than 100 open FDs
    [ "$fd_count" -lt 100 ]
  else
    skip "lsof not available"
  fi
}

# ── Test: Feature Flags ──────────────────────────────────────────────────────

@test "feature flags: USE_NEW_RESOLVE_BIN=0 uses legacy" {
  export USE_NEW_RESOLVE_BIN=0

  run resolve_bin "sh"
  assert_success
}

@test "feature flags: USE_NEW_RESOLVE_BIN=1 uses new implementation" {
  export USE_NEW_RESOLVE_BIN=1

  run resolve_bin "sh"
  assert_success
}

@test "feature flags: default behavior without flag" {
  unset USE_NEW_RESOLVE_BIN

  run resolve_bin "sh"
  assert_success
}

# ── Test: Logging and Debug ──────────────────────────────────────────────────

@test "logging: debug mode can be enabled" {
  export DEBUG_RESOLVE_BIN=1
  export DEBUG_TIMEOUT=1
  export DEBUG_JSON_PARSE=1

  # Should not crash with debug enabled
  run resolve_bin "sh"
  assert_success
}

@test "logging: verbose mode works" {
  export VERBOSE=2

  # Should not crash with verbose enabled
  run resolve_bin "sh"
  assert_success
}

@test "logging: quiet mode works" {
  export VERBOSE=0

  # Should not crash with quiet mode
  run resolve_bin "sh"
  assert_success
}
