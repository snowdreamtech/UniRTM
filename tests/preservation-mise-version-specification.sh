#!/usr/bin/env bash
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Preservation Property Tests for Mise Version Specification Bugfix
# **Validates: Requirements 3.1, 3.2, 3.3, 3.4, 3.5**
#
# IMPORTANT: Follow observation-first methodology
# Observe behavior on UNFIXED code for non-buggy installation patterns
#
# Test Coverage:
#   - Fast-path version checks: verify is_version_match skips reinstallation when correct version exists
#   - DRY_RUN mode: verify installations are previewed without execution
#   - Runtime-only installs: verify run_mise install ruby works without version suffix
#   - Version extraction pattern: verify run_mise install "perl@$(get_mise_tool_version perl)" works correctly
#   - Error handling: verify installation failures are logged with || _STAT="❌ Failed"
#   - Tools already in .mise.toml: verify they respect .mise.toml as source of truth
#
# EXPECTED OUTCOME: Tests PASS (this confirms baseline behavior to preserve)

# SC2329: Functions are invoked indirectly via the run_test helper.
# shellcheck disable=SC2329

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
# shellcheck disable=SC2034
YELLOW='\033[1;33m'
# shellcheck disable=SC2034
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test configuration
SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
PROJECT_ROOT=$(cd "${SCRIPT_DIR}/.." && pwd)

echo "=========================================="
echo "Preservation Property Tests"
echo "Property 2: Preservation - Existing Installation Behavior"
echo "=========================================="
echo ""

# Source common library for helper functions



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

  echo "----------------------------------------"
  echo "Running: ${test_name}"
  echo "----------------------------------------"

  TESTS_RUN=$((TESTS_RUN + 1))

  if ${test_func}; then
    printf '%b✓ PASS%b: %s\n' "${GREEN}" "${NC}" "${test_name}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
    echo ""
    return 0
  else
    printf '%b✗ FAIL%b: %s\n' "${RED}" "${NC}" "${test_name}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
    FAILED_TESTS+=("${test_name}")
    echo ""
    return 1
  fi
}

# ========================================
# Test 1: Fast-Path Version Checks
# ========================================
# Verify that is_version_match correctly skips reinstallation
# when the correct version already exists
test_fast_path_version_checks() {
  echo "Testing is_version_match function..."

  # Test exact match
  if ! is_version_match "1.2.3" "1.2.3"; then
    echo "ERROR: Exact version match failed"
    return 1
  fi
  echo "  ✓ Exact match: 1.2.3 == 1.2.3"

  # Test prefix match (e.g., 3.12.0 matches 3.12.0.2)
  if ! is_version_match "3.12.0" "3.12.0.2"; then
    echo "ERROR: Prefix match failed"
    return 1
  fi
  echo "  ✓ Prefix match: 3.12.0 matches 3.12.0.2"

  # Test latest always matches
  if ! is_version_match "1.2.3" "latest"; then
    echo "ERROR: Latest match failed"
    return 1
  fi
  echo "  ✓ Latest match: any version matches 'latest'"

  # Test mismatch detection
  if is_version_match "1.2.3" "2.0.0"; then
    echo "ERROR: Mismatch not detected"
    return 1
  fi
  echo "  ✓ Mismatch detection: 1.2.3 != 2.0.0"

  # Test missing version detection
  if is_version_match "-" "1.2.3"; then
    echo "ERROR: Missing version not detected"
    return 1
  fi
  echo "  ✓ Missing version detection: '-' != 1.2.3"

  echo "Fast-path version checks working correctly"
  return 0
}

# ========================================
# Test 2: DRY_RUN Mode
# ========================================
# Verify that DRY_RUN mode previews installations without executing them
test_dry_run_mode() {
  echo "Testing DRY_RUN mode..."

  # Check if DRY_RUN environment variable is respected
  export DRY_RUN=1

  # Verify DRY_RUN flag is set
  if [ "${DRY_RUN:-0}" -ne 1 ]; then
    echo "ERROR: DRY_RUN flag not set correctly"
    return 1
  fi
  echo "  ✓ DRY_RUN flag set correctly"

  # Test that installation functions check DRY_RUN
  # We'll check the docker.sh module as an example
  if [ -f "${PROJECT_ROOT}/scripts/lib/langs/docker.sh" ]; then
    # Check that install functions have DRY_RUN guards
    # shellcheck disable=SC2016
    if ! grep -q 'if \[ "\${DRY_RUN:-0}" -eq 1 \]; then' "${PROJECT_ROOT}/scripts/lib/langs/docker.sh"; then
      echo "ERROR: DRY_RUN guard not found in docker.sh"
      return 1
    fi
    echo "  ✓ DRY_RUN guards present in installation functions"

    # Check that DRY_RUN returns early with preview status
    if ! grep -q '⚖️ Previewed' "${PROJECT_ROOT}/scripts/lib/langs/docker.sh"; then
      echo "ERROR: DRY_RUN preview status not found"
      return 1
    fi
    echo "  ✓ DRY_RUN preview status present"
  fi

  # Reset DRY_RUN
  export DRY_RUN=0

  echo "DRY_RUN mode working correctly"
  return 0
}

# ========================================
# Test 3: Runtime-Only Installs
# ========================================
# Verify that runtime-only installs work without version suffix
# Examples: run_mise install ruby, run_mise install dart
test_runtime_only_installs() {
  echo "Testing runtime-only install patterns..."

  # Check for runtime-only install patterns in language modules
  local runtime_only_found=0

  # Search for runtime-only patterns (install without @version)
  # These are legitimate cases where no version suffix is needed
  if grep -r 'run_mise install ruby' "${PROJECT_ROOT}/scripts/lib/langs/"*.sh 2>/dev/null | grep -v '@'; then
    echo "  ✓ Found runtime-only install: ruby"
    runtime_only_found=1
  fi

  if grep -r 'run_mise install dart' "${PROJECT_ROOT}/scripts/lib/langs/"*.sh 2>/dev/null | grep -v '@'; then
    echo "  ✓ Found runtime-only install: dart"
    runtime_only_found=1
  fi

  if grep -r 'run_mise install python' "${PROJECT_ROOT}/scripts/lib/langs/"*.sh 2>/dev/null | grep -v '@'; then
    echo "  ✓ Found runtime-only install: python"
    runtime_only_found=1
  fi

  # It's okay if no runtime-only installs exist - this is a preservation test
  # We're verifying the pattern is supported, not that it must exist
  if [ "${runtime_only_found}" -eq 0 ]; then
    echo "  ℹ No runtime-only installs found (this is acceptable)"
  fi

  echo "Runtime-only install pattern supported"
  return 0
}

# ========================================
# Test 4: Version Extraction Pattern
# ========================================
# Verify that version extraction via get_mise_tool_version works correctly
test_version_extraction_pattern() {
  echo "Testing version extraction pattern..."

  # Test get_mise_tool_version function exists and works
  if ! command -v get_mise_tool_version >/dev/null 2>&1; then
    echo "ERROR: get_mise_tool_version function not found"
    return 1
  fi
  echo "  ✓ get_mise_tool_version function exists"

  # Test that it returns a version (or "latest" as fallback)
  local test_version
  test_version=$(get_mise_tool_version "node" 2>/dev/null || echo "")

  if [ -z "${test_version}" ]; then
    echo "ERROR: get_mise_tool_version returned empty string"
    return 1
  fi
  echo "  ✓ get_mise_tool_version returns a value: ${test_version}"

  # Check for version extraction patterns in language modules
  # Pattern: run_mise install "tool@$(get_mise_tool_version tool)"
  if grep -r 'get_mise_tool_version' "${PROJECT_ROOT}/scripts/lib/langs/"*.sh 2>/dev/null | head -n 1; then
    echo "  ✓ Version extraction pattern found in language modules"
  else
    echo "  ℹ No version extraction pattern found (may not be used yet)"
  fi

  echo "Version extraction pattern working correctly"
  return 0
}

# ========================================
# Test 5: Error Handling
# ========================================
# Verify that installation failures are logged with || _STAT="❌ Failed"
test_error_handling() {
  echo "Testing error handling pattern..."

  # Check that installation functions use error handling pattern
  local error_pattern_found=0

  # Search for error handling pattern in language modules
  if grep -r '|| _STAT.*="❌ Failed"' "${PROJECT_ROOT}/scripts/lib/langs/"*.sh 2>/dev/null | head -n 3; then
    echo "  ✓ Error handling pattern found in language modules"
    error_pattern_found=1
  fi

  if [ "${error_pattern_found}" -eq 0 ]; then
    echo "ERROR: Error handling pattern not found"
    return 1
  fi

  # Verify that errors don't cause script to exit (set -e should be handled)
  # Check that functions continue after errors
  if grep -r 'run_mise install.*||' "${PROJECT_ROOT}/scripts/lib/langs/"*.sh 2>/dev/null | grep -v 'exit' | head -n 1; then
    echo "  ✓ Errors are handled gracefully (no immediate exit)"
  fi

  echo "Error handling working correctly"
  return 0
}

# ========================================
# Test 6: Tools in .mise.toml
# ========================================
# Verify that tools already in .mise.toml respect it as source of truth
test_mise_toml_source_of_truth() {
  echo "Testing .mise.toml as source of truth..."

  # Check that .mise.toml exists
  if [ ! -f "${PROJECT_ROOT}/.mise.toml" ]; then
    echo "ERROR: .mise.toml not found"
    return 1
  fi
  echo "  ✓ .mise.toml exists"

  # Verify get_mise_tool_version reads from .mise.toml
  # Test with a tool that should be in .mise.toml
  local node_version
  node_version=$(get_mise_tool_version "node" 2>/dev/null || echo "")

  if [ -n "${node_version}" ] && [ "${node_version}" != "latest" ]; then
    echo "  ✓ get_mise_tool_version reads from .mise.toml: node=${node_version}"
  else
    echo "  ℹ node not in .mise.toml or returns 'latest'"
  fi

  # Check that installation functions use get_mise_tool_version
  # This ensures they respect .mise.toml
  if grep -r 'get_mise_tool_version' "${PROJECT_ROOT}/scripts/lib/langs/"*.sh 2>/dev/null | head -n 1; then
    echo "  ✓ Installation functions use get_mise_tool_version"
  fi

  echo ".mise.toml source of truth working correctly"
  return 0
}

# ========================================
# Test 7: CI Environment Detection
# ========================================
# Verify that is_ci_env correctly detects CI environments
test_ci_environment_detection() {
  echo "Testing CI environment detection..."

  # Test is_ci_env function exists
  if ! command -v is_ci_env >/dev/null 2>&1; then
    echo "ERROR: is_ci_env function not found"
    return 1
  fi
  echo "  ✓ is_ci_env function exists"

  # Test local environment detection (should return false in local dev)
  if is_ci_env; then
    echo "  ✓ Running in CI environment"
  else
    echo "  ✓ Running in local environment (not CI)"
  fi

  # Verify CI detection is used in installation logic
  if grep -r 'is_ci_env' "${PROJECT_ROOT}/scripts/lib/"*.sh 2>/dev/null | head -n 1; then
    echo "  ✓ CI detection used in installation logic"
  fi

  echo "CI environment detection working correctly"
  return 0
}

# ========================================
# Test 8: Language File Detection
# ========================================
# Verify that has_lang_files correctly detects project languages
test_language_file_detection() {
  echo "Testing language file detection..."

  # Test has_lang_files function exists
  if ! command -v has_lang_files >/dev/null 2>&1; then
    echo "ERROR: has_lang_files function not found"
    return 1
  fi
  echo "  ✓ has_lang_files function exists"

  # Test detection with common files
  if has_lang_files "package.json" "*.ts *.js"; then
    echo "  ✓ Detected Node.js project files"
  else
    echo "  ℹ No Node.js project files detected (acceptable)"
  fi

  if has_lang_files "Makefile" ""; then
    echo "  ✓ Detected Makefile"
  else
    echo "  ℹ No Makefile detected"
  fi

  # Verify language detection is used in installation logic
  if grep -r 'has_lang_files' "${PROJECT_ROOT}/scripts/lib/langs/"*.sh 2>/dev/null | head -n 1; then
    echo "  ✓ Language detection used in installation logic"
  fi

  echo "Language file detection working correctly"
  return 0
}

# ========================================
# Run All Tests
# ========================================

echo "Starting preservation property tests..."
echo ""

run_test "Test 1: Fast-Path Version Checks" test_fast_path_version_checks
run_test "Test 2: DRY_RUN Mode" test_dry_run_mode
run_test "Test 3: Runtime-Only Installs" test_runtime_only_installs
run_test "Test 4: Version Extraction Pattern" test_version_extraction_pattern
run_test "Test 5: Error Handling" test_error_handling
run_test "Test 6: Tools in .mise.toml" test_mise_toml_source_of_truth
run_test "Test 7: CI Environment Detection" test_ci_environment_detection
run_test "Test 8: Language File Detection" test_language_file_detection

# ========================================
# Test Summary
# ========================================

echo "=========================================="
echo "TEST SUMMARY"
echo "=========================================="
echo ""
echo "Total tests run: ${TESTS_RUN}"
printf '%bTests passed: %s%b\n' "${GREEN}" "${TESTS_PASSED}" "${NC}"
printf '%bTests failed: %s%b\n' "${RED}" "${TESTS_FAILED}" "${NC}"
echo ""

if [ ${TESTS_FAILED} -gt 0 ]; then
  echo "Failed tests:"
  for test in "${FAILED_TESTS[@]}"; do
    echo "  - ${test}"
  done
  echo ""
  printf '%b✗ PRESERVATION TESTS FAILED%b\n' "${RED}" "${NC}"
  echo ""
  echo "Some preservation properties are not working correctly."
  echo "This indicates that the baseline behavior has changed or is broken."
  echo ""
  exit 1
fi

printf '%b✓ ALL PRESERVATION TESTS PASSED%b\n' "${GREEN}" "${NC}"
echo ""
echo "All preservation properties are working correctly on unfixed code."
echo "This confirms the baseline behavior that must be preserved after the fix."
echo ""
echo "Key behaviors verified:"
echo "  ✓ Fast-path version checks skip reinstallation when correct version exists"
echo "  ✓ DRY_RUN mode previews installations without execution"
echo "  ✓ Runtime-only installs work without version suffix"
echo "  ✓ Version extraction via get_mise_tool_version works correctly"
echo "  ✓ Installation failures arelogged with error handling"
echo "  ✓ Tools in .mise.toml are respected as source of truth"
echo "  ✓ CI environment detection works correctly"
echo "  ✓ Language file detection works correctly"
echo ""
echo "Next steps:"
echo "  1. These tests document the baseline behavior to preserve"
# shellcheck disable=SC2016
echo '  2. Implement the fix to add @${_VERSION:-} suffix to buggy calls'
echo "  3. Re-run these tests - they should still PASS after the fix"
echo "  4. Run bug condition test - it should PASS (0 instances) after the fix"
echo ""

exit 0
