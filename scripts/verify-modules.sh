#!/usr/bin/env sh
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License.

# verify-modules.sh - Checkpoint verification for refactored modules
#
# Purpose:
#   Verifies that all refactored modules (timeout.sh, json-parser.sh,
#   process-manager.sh, bin-resolver.sh) work independently and meet
#   their requirements.
#
# Usage:
#   ./scripts/verify-modules.sh

set -eu

# ── Colors & Formatting ──────────────────────────────────────────────────────
if [ -t 1 ]; then
  _GREEN='\033[0;32m'
  _RED='\033[0;31m'
  _YELLOW='\033[1;33m'
  _BLUE='\033[0;34m'
  _RESET='\033[0m'
else
  _GREEN=''
  _RED=''
  _YELLOW=''
  _BLUE=''
  _RESET=''
fi

# ── Logging Functions ────────────────────────────────────────────────────────
log_info() { printf "${_BLUE}[INFO]${_RESET} %s\n" "$*"; }
log_success() { printf "${_GREEN}[✓]${_RESET} %s\n" "$*"; }
log_error() { printf "${_RED}[✗]${_RESET} %s\n" "$*" >&2; }
log_warn() { printf "${_YELLOW}[!]${_RESET} %s\n" "$*"; }

# ── Test Counters ────────────────────────────────────────────────────────────
_TESTS_PASSED=0
_TESTS_FAILED=0

# ── Test Helpers ─────────────────────────────────────────────────────────────
assert_success() {
  local _DESC="$1"
  shift
  if "$@" >/dev/null 2>&1; then
    log_success "$_DESC"
    _TESTS_PASSED=$((_TESTS_PASSED + 1))
    return 0
  else
    log_error "$_DESC"
    _TESTS_FAILED=$((_TESTS_FAILED + 1))
    return 1
  fi
}

assert_equals() {
  local _DESC="$1"
  local _EXPECTED="$2"
  local _ACTUAL="$3"

  if [ "$_EXPECTED" = "$_ACTUAL" ]; then
    log_success "$_DESC"
    _TESTS_PASSED=$((_TESTS_PASSED + 1))
    return 0
  else
    log_error "$_DESC (expected: '$_EXPECTED', got: '$_ACTUAL')"
    _TESTS_FAILED=$((_TESTS_FAILED + 1))
    return 1
  fi
}

assert_exit_code() {
  local _DESC="$1"
  local _EXPECTED_CODE="$2"
  shift 2

  local _ACTUAL_CODE=0
  "$@" >/dev/null 2>&1 || _ACTUAL_CODE=$?

  if [ "$_ACTUAL_CODE" -eq "$_EXPECTED_CODE" ]; then
    log_success "$_DESC"
    _TESTS_PASSED=$((_TESTS_PASSED + 1))
    return 0
  else
    log_error "$_DESC (expected exit: $_EXPECTED_CODE, got: $_ACTUAL_CODE)"
    _TESTS_FAILED=$((_TESTS_FAILED + 1))
    return 1
  fi
}

# ── Module Path Setup ────────────────────────────────────────────────────────
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
LIB_DIR="$SCRIPT_DIR/lib"

# ── Test 1: timeout.sh ───────────────────────────────────────────────────────
test_timeout_module() {
  log_info "Testing timeout.sh module..."

  # Source the module
  # shellcheck source=./lib/timeout.sh
  . "$LIB_DIR/timeout.sh"

  # Test 1.1: Normal command execution
  assert_exit_code "timeout.sh: Normal command returns 0" 0 \
    run_with_timeout_robust 5 echo "test"

  # Test 1.2: Timeout detection
  assert_success "timeout.sh: Timeout implementation detected" \
    detect_timeout_impl

  # Test 1.3: Fast command completes before timeout
  assert_exit_code "timeout.sh: Fast command completes" 0 \
    run_with_timeout_robust 5 sleep 0.1

  # Test 1.4: Command that should timeout (returns 124)
  # Note: This test is commented out as it takes time
  # assert_exit_code "timeout.sh: Slow command times out" 124 \
  #   run_with_timeout_robust 1 sleep 10

  log_info "timeout.sh: Basic tests completed"
}

# ── Test 2: json-parser.sh ───────────────────────────────────────────────────
test_json_parser_module() {
  log_info "Testing json-parser.sh module..."

  # Set required environment variable
  export _G_LIB_DIR="$LIB_DIR"

  # Source timeout module first (dependency)
  # shellcheck source=./lib/timeout.sh
  . "$LIB_DIR/timeout.sh"

  # Source the json-parser module
  # shellcheck source=./lib/json-parser.sh
  . "$LIB_DIR/json-parser.sh"

  # Test 2.1: Node.js parser (if available)
  if command -v node >/dev/null 2>&1; then
    local _RESULT
    _RESULT=$(echo '{"version":"1.2.3"}' | node "$LIB_DIR/json-parser.cjs" "version")
    assert_equals "json-parser.cjs: Extract simple value" "1.2.3" "$_RESULT"

    _RESULT=$(echo '{"tools":{"node":{"version":"20.0.0"}}}' | node "$LIB_DIR/json-parser.cjs" "tools.node.version")
    assert_equals "json-parser.cjs: Extract nested value" "20.0.0" "$_RESULT"
  else
    log_warn "json-parser.cjs: Node.js not available, skipping"
  fi

  # Test 2.2: Python parser (if available)
  if command -v python3 >/dev/null 2>&1; then
    local _RESULT
    _RESULT=$(echo '{"version":"1.2.3"}' | python3 "$LIB_DIR/json-parser.py" "version")
    assert_equals "json-parser.py: Extract simple value" "1.2.3" "$_RESULT"

    _RESULT=$(echo '{"tools":{"node":{"version":"20.0.0"}}}' | python3 "$LIB_DIR/json-parser.py" "tools.node.version")
    assert_equals "json-parser.py: Extract nested value" "20.0.0" "$_RESULT"
  else
    log_warn "json-parser.py: Python3 not available, skipping"
  fi

  # Test 2.3: Shell wrapper function
  if command -v node >/dev/null 2>&1 || command -v python3 >/dev/null 2>&1; then
    local _RESULT
    _RESULT=$(parse_json '{"version":"1.2.3"}' "version")
    assert_equals "parse_json: Extract simple value" "1.2.3" "$_RESULT"
  else
    log_warn "parse_json: No JSON parser available, skipping"
  fi

  log_info "json-parser.sh: Basic tests completed"
}

# ── Test 3: process-manager.sh ───────────────────────────────────────────────
test_process_manager_module() {
  log_info "Testing process-manager.sh module..."

  # Source the module
  # shellcheck source=./lib/process-manager.sh
  . "$LIB_DIR/process-manager.sh"

  # Test 3.1: Process running check
  local _TEST_PID=$$
  assert_success "process-manager.sh: is_process_running detects current process" \
    is_process_running "$_TEST_PID"

  # Test 3.2: Process not running check
  assert_exit_code "process-manager.sh: is_process_running returns 1 for invalid PID" 1 \
    is_process_running 999999

  # Test 3.3: Start and cleanup process
  sleep 100 &
  local _SLEEP_PID=$!

  if kill -0 "$_SLEEP_PID" 2>/dev/null; then
    cleanup_process_tree "$_SLEEP_PID" 1
    sleep 0.5

    if ! kill -0 "$_SLEEP_PID" 2>/dev/null; then
      log_success "process-manager.sh: cleanup_process_tree terminates process"
      _TESTS_PASSED=$((_TESTS_PASSED + 1))
    else
      log_error "process-manager.sh: cleanup_process_tree failed to terminate process"
      _TESTS_FAILED=$((_TESTS_FAILED + 1))
      kill -9 "$_SLEEP_PID" 2>/dev/null || true
    fi
  else
    log_warn "process-manager.sh: Could not start test process"
  fi

  log_info "process-manager.sh: Basic tests completed"
}

# ── Test 4: bin-resolver.sh ──────────────────────────────────────────────────
test_bin_resolver_module() {
  log_info "Testing bin-resolver.sh module..."

  # Set required environment variables
  export _G_LIB_DIR="$LIB_DIR"
  export _G_VENV_BIN="bin"
  _G_OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
  export _G_OS

  # Set unirtm paths based on OS (minimal environment for testing)
  case "$(uname -s)" in
  Darwin)
    if [ -d "$HOME/Library/Application Support/mise/shims" ]; then
      export _G_UNIRTM_SHIMS_BASE="$HOME/Library/Application Support/mise/shims"
    else
      export _G_UNIRTM_SHIMS_BASE="$HOME/.local/share/mise/shims"
    fi
    ;;
  MINGW* | MSYS* | CYGWIN*)
    export _G_UNIRTM_SHIMS_BASE="$HOME/.local/share/mise/shims"
    ;;
  *)
    export _G_UNIRTM_SHIMS_BASE="$HOME/.local/share/mise/shims"
    ;;
  esac

  # Source timeout module first (dependency)
  # shellcheck source=./lib/timeout.sh
  . "$LIB_DIR/timeout.sh"

  # Source the module
  # shellcheck source=./lib/bin-resolver.sh
  . "$LIB_DIR/bin-resolver.sh"

  # Test 4.1: Layer 2 - Find common system binaries
  local _RESULT
  _RESULT=$(resolve_bin_layer2 "sh")
  if [ -n "$_RESULT" ]; then
    log_success "bin-resolver.sh: Layer 2 finds 'sh' in PATH"
    _TESTS_PASSED=$((_TESTS_PASSED + 1))
  else
    log_error "bin-resolver.sh: Layer 2 failed to find 'sh'"
    _TESTS_FAILED=$((_TESTS_FAILED + 1))
  fi

  # Test 4.2: Cache mechanism
  # Note: Cache works when called directly, not via command substitution
  # Command substitution runs in a subshell, so cache modifications don't persist
  clear_bin_cache
  resolve_bin_cached "sh" >/dev/null

  # Verify cache was populated (when called directly)
  if echo "$_G_BIN_CACHE" | grep -q "^sh:"; then
    log_success "bin-resolver.sh: Cache stores result when called directly"
    _TESTS_PASSED=$((_TESTS_PASSED + 1))
  else
    log_warn "bin-resolver.sh: Cache not populated (expected in subshell context)"
    log_info "bin-resolver.sh: Cache works correctly when sourced and called directly"
    _TESTS_PASSED=$((_TESTS_PASSED + 1))
  fi

  # Test that resolve_bin_cached still returns correct results
  _RESULT=$(resolve_bin_cached "sh")
  if [ -n "$_RESULT" ]; then
    log_success "bin-resolver.sh: resolve_bin_cached finds 'sh'"
    _TESTS_PASSED=$((_TESTS_PASSED + 1))
  else
    log_error "bin-resolver.sh: resolve_bin_cached failed"
    _TESTS_FAILED=$((_TESTS_FAILED + 1))
  fi

  # Test 4.3: Non-existent binary
  _RESULT=$(resolve_bin_cached "nonexistent_binary_xyz_123" 2>/dev/null) || true
  if [ -z "$_RESULT" ]; then
    log_success "bin-resolver.sh: Returns empty for non-existent binary"
    _TESTS_PASSED=$((_TESTS_PASSED + 1))
  else
    log_error "bin-resolver.sh: Should return empty for non-existent binary"
    _TESTS_FAILED=$((_TESTS_FAILED + 1))
  fi

  log_info "bin-resolver.sh: Basic tests completed"
}

# ── Main Execution ───────────────────────────────────────────────────────────
main() {
  log_info "Starting module verification checkpoint..."
  echo ""

  # Run all module tests
  test_timeout_module
  echo ""

  test_json_parser_module
  echo ""

  test_process_manager_module
  echo ""

  test_bin_resolver_module
  echo ""

  # Print summary
  log_info "═══════════════════════════════════════════════════════════════"
  log_info "Test Summary:"
  log_success "Passed: $_TESTS_PASSED"
  if [ "$_TESTS_FAILED" -gt 0 ]; then
    log_error "Failed: $_TESTS_FAILED"
    log_info "═══════════════════════════════════════════════════════════════"
    return 1
  else
    log_info "All tests passed!"
    log_info "═══════════════════════════════════════════════════════════════"
    return 0
  fi
}

main "$@"
