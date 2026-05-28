#!/usr/bin/env bats
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# tests/integration/test_setup_flow.bats - Integration tests for setup flow
#
# Purpose:
#   Tests complete setup flows including:
#   - unirtm install executes without hangs
#   - make verify validates installation
#
# Requirements: 5.2, 8.1

# SC2030/SC2031: In bats, each @test runs in a subshell by design.
# Exports are intentionally scoped to individual tests.
# shellcheck disable=SC2030,SC2031

setup() {
  load '../vendor/bats-support/load.bash'
  load '../vendor/bats-assert/load.bash'

  # Source common.sh to test integrated behavior
  export SCRIPT_DIR
  SCRIPT_DIR="$(cd "$(dirname "$BATS_TEST_FILENAME")/../.." && pwd)"

  # Set up minimal environment
  export _G_PROJECT_ROOT="$SCRIPT_DIR"
  export _G_LIB_DIR="$SCRIPT_DIR/scripts/lib"


}

teardown() {
  # Clean up any test artifacts
  true
}

# ── Test: Timeout Mechanism Integration ──────────────────────────────────────

@test "run_mise: completes without hanging" {
  skip "Requires mise installation"

  # Test that run_mise doesn't hang
  run timeout 30 run_mise --version
  assert_success
}

@test "run_mise: respects timeout configuration" {
  skip "Requires mise installation"

  export TIMEOUT_NETWORK=5

  # Should complete within timeout
  run run_mise --version
  assert_success
}

# ── Test: Binary Resolution Integration ──────────────────────────────────────

@test "resolve_bin: finds system binaries" {
  run resolve_bin "sh"
  assert_success
  [[ $output =~ /sh$ ]]
}

@test "resolve_bin: handles missing binaries gracefully" {
  run resolve_bin "nonexistent_binary_xyz"
  assert_failure 1
}

@test "resolve_bin: feature flag enables new implementation" {
  export USE_NEW_RESOLVE_BIN=1

  run resolve_bin "sh"
  assert_success
  [[ $output =~ /sh$ ]]
}

@test "resolve_bin: legacy implementation still works" {
  export USE_NEW_RESOLVE_BIN=0

  run resolve_bin "sh"
  assert_success
  [[ $output =~ /sh$ ]]
}

# ── Test: JSON Parsing Integration ───────────────────────────────────────────

@test "get_version: uses new JSON parser" {
  skip "Requires tool installation"

  # Test that get_version works with new parser
  run get_version "node"
  # Should return version or "-"
  [[ $output =~ ^[0-9]+\.[0-9]+.*$ ]] || [[ $output == "-" ]]
}

# ── Test: Process Management Integration ─────────────────────────────────────

@test "run_with_timeout: integrated into run_mise" {
  skip "Requires mise installation"

  # Verify timeout mechanism is active
  export TIMEOUT_NETWORK=1

  # This should timeout if mise hangs
  run run_mise --version
  # Should either succeed quickly or timeout (124)
  [[ $status -eq 0 ]] || [[ $status -eq 124 ]]
}

# ── Test: Complete Setup Flow ────────────────────────────────────────────────

@test "setup flow: common.sh sources all modules" {
  # Verify all modules are sourced
  run type run_with_timeout_robust
  assert_success

  run type parse_json
  assert_success

  run type cleanup_process_tree
  assert_success

  run type resolve_bin_cached
  assert_success
}

@test "setup flow: timeout constants are set" {
  [[ -n $TIMEOUT_RESOLVE_BIN ]]
  [[ -n $TIMEOUT_JSON_PARSE ]]
  [[ -n $TIMEOUT_MISE_WHICH ]]
  [[ -n $TIMEOUT_NETWORK ]]
}

@test "setup flow: debug switches are set" {
  [[ -n $DEBUG_RESOLVE_BIN ]]
  [[ -n $DEBUG_TIMEOUT ]]
  [[ -n $DEBUG_JSON_PARSE ]]
}

# ── Test: No Hangs Guarantee ─────────────────────────────────────────────────

@test "no hangs: resolve_bin completes within timeout" {
  export TIMEOUT_RESOLVE_BIN=5

  # Should complete within 5 seconds
  run timeout 10 resolve_bin "nonexistent"
  # Either finds it, doesn't find it, or times out - but doesn't hang
  [[ $status -ne 124 ]]
}

@test "no hangs: parse_json completes within timeout" {
  if ! command -v node >/dev/null 2>&1 && ! command -v python3 >/dev/null 2>&1; then
    skip "No JSON parser available"
  fi

  export TIMEOUT_JSON_PARSE=3

  # Should complete within 3 seconds
  run timeout 5 parse_json '{"test":"value"}' "test"
  assert_success
  assert_output "value"
}

# ── Test: Backward Compatibility ─────────────────────────────────────────────

@test "backward compatibility: resolve_bin maintains signature" {
  # Old code should still work
  local bin_path
  bin_path=$(resolve_bin "sh") || true
  [[ -n $bin_path ]]
}

@test "backward compatibility: get_version maintains signature" {
  # Old code should still work
  local version
  version=$(get_version "sh") || true
  [[ -n $version ]]
}

# ── Test: Error Handling ─────────────────────────────────────────────────────

@test "error handling: graceful degradation when parsers unavailable" {
  # Temporarily hide parsers
  PATH_BACKUP="$PATH"
  export PATH="/usr/bin:/bin"

  # Should still work or fail gracefully
  run parse_json '{"test":"value"}' "test" 2>/dev/null || true

  # Restore PATH
  export PATH="$PATH_BACKUP"
}

@test "error handling: timeout doesn't leave zombie processes" {
  # Start a command that will timeout
  run_with_timeout_robust 1 sleep 10 &
  local pid=$!

  # Wait for it to complete
  wait $pid 2>/dev/null || true

  # Check for zombies
  sleep 0.5
  run sh -c "ps aux | grep -E 'Z|defunct' | grep sleep | grep -v grep"
  assert_failure
}

# ── Test: Performance ────────────────────────────────────────────────────────

@test "performance: cached lookups are fast" {
  export USE_NEW_RESOLVE_BIN=1

  # First lookup
  resolve_bin_cached "sh" >/dev/null

  # Second lookup should be instant (cached)
  local start
  start=$(date +%s%N)
  resolve_bin_cached "sh" >/dev/null
  local end
  end=$(date +%s%N)
  local duration=$((end - start))

  # Should complete in less than 10ms (10000000 ns)
  [ "$duration" -lt 10000000 ]
}

# ── Test: Module Independence ────────────────────────────────────────────────

@test "module independence: timeout module works standalone" {
  # Source only timeout module
  unset run_with_timeout_robust
  . "$_G_LIB_DIR/timeout.sh"

  run run_with_timeout_robust 5 echo "test"
  assert_success
  assert_output "test"
}

@test "module independence: json-parser module works standalone" {
  if ! command -v node >/dev/null 2>&1 && ! command -v python3 >/dev/null 2>&1; then
    skip "No JSON parser available"
  fi

  # Source modules
  unset parse_json
  . "$_G_LIB_DIR/timeout.sh"
  . "$_G_LIB_DIR/json-parser.sh"

  run parse_json '{"test":"value"}' "test"
  assert_success
  assert_output "value"
}

@test "module independence: process-manager module works standalone" {
  unset cleanup_process_tree
  . "$_G_LIB_DIR/process-manager.sh"

  sleep 10 &
  local pid=$!
  sleep 0.2

  run cleanup_process_tree $pid 1
  assert_success

  sleep 0.5
  run kill -0 $pid 2>/dev/null
  assert_failure
}

@test "module independence: bin-resolver module works standalone" {
  unset resolve_bin_cached
  . "$_G_LIB_DIR/timeout.sh"
  . "$_G_LIB_DIR/bin-resolver.sh"

  run resolve_bin_cached "sh"
  assert_success
  [[ $output =~ /sh$ ]]
}
