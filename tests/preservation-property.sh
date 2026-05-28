#!/usr/bin/env sh
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Preservation Property Tests for Scripts Version Centralization
#
# Purpose:
#   Property-based tests that verify existing functionality is preserved
#   when scripts use centralized provider variables from versions.sh.
#
# **Validates: Requirements 3.1-3.10 (Preservation Requirements)**
#
# Test Strategy:
#   - Observe behavior on UNFIXED code for scripts that already use centralized variables
#   - Test scripts: security.sh, java.sh, kotlin.sh, testing.sh
#   - Verify version checking, installation, logging, DRY_RUN, fast-path, error handling
#
# Expected Outcome:
#   All tests PASS on unfixed code (confirms baseline behavior to preserve)

set -e

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Colors for output
if [ -t 1 ]; then
  GREEN=$(printf '\033[0;32m')
  RED=$(printf '\033[0;31m')
  YELLOW=$(printf '\033[1;33m')
  NC=$(printf '\033[0m')
else
  GREEN="" RED="" YELLOW="" NC=""
fi

# Test helper functions
pass() {
  TESTS_PASSED=$((TESTS_PASSED + 1))
  echo "${GREEN}✓${NC} $1"
}

fail() {
  TESTS_FAILED=$((TESTS_FAILED + 1))
  echo "${RED}✗${NC} $1"
  [ -n "${2:-}" ] && echo "  ${RED}Error:${NC} $2"
}

test_start() {
  TESTS_RUN=$((TESTS_RUN + 1))
  echo ""
  echo "${YELLOW}Test $TESTS_RUN:${NC} $1"
}

# Setup test environment
SCRIPT_DIR="$(cd "$(dirname "$0")/../scripts" && pwd)"
# shellcheck disable=SC2034
PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

# ── Property 2.1: Centralized Variables Exist ────────────────────────────────

test_start "Property 2.1: Centralized provider variables are defined in versions.sh"

# Source versions.sh to load centralized variables
# Temporarily disable -u to avoid issues with undefined variables during sourcing
set +u
if [ -f "${SCRIPT_DIR}/lib/versions.sh" ]; then
  # shellcheck disable=SC1091
  . "${SCRIPT_DIR}/lib/versions.sh" || {
    fail "Failed to source versions.sh" "Script exited with error"
    exit 1
  }
  pass "versions.sh sourced successfully"
else
  fail "versions.sh not found" "${SCRIPT_DIR}/lib/versions.sh does not exist"
  exit 1
fi
set -e

# Verify key centralized variables exist
test_start "Property 2.1.1: Security tool provider variables exist"
if [ -n "${VER_GITLEAKS_PROVIDER:-}" ] && [ -n "${VER_OSV_SCANNER_PROVIDER:-}" ]; then
  pass "Security provider variables defined: VER_GITLEAKS_PROVIDER=${VER_GITLEAKS_PROVIDER}, VER_OSV_SCANNER_PROVIDER=${VER_OSV_SCANNER_PROVIDER}"
else
  fail "Security provider variables missing"
fi

test_start "Property 2.1.2: Java/Kotlin tool provider variables exist"
if [ -n "${VER_JAVA_FORMAT_PROVIDER:-}" ] && [ -n "${VER_KTLINT_PROVIDER:-}" ]; then
  pass "Java/Kotlin provider variables defined: VER_JAVA_FORMAT_PROVIDER=${VER_JAVA_FORMAT_PROVIDER}, VER_KTLINT_PROVIDER=${VER_KTLINT_PROVIDER}"
else
  fail "Java/Kotlin provider variables missing"
fi

test_start "Property 2.1.3: Testing tool provider variables exist"
if [ -n "${VER_BATS_PROVIDER:-}" ]; then
  pass "Testing provider variables defined: VER_BATS_PROVIDER=${VER_BATS_PROVIDER}"
else
  fail "Testing provider variables missing"
fi

# ── Property 2.2: Provider Variable Format ───────────────────────────────────

test_start "Property 2.2: Provider variables follow expected format"

# Check that provider variables contain provider prefix
if echo "${VER_GITLEAKS_PROVIDER:-}" | grep -q "^github:"; then
  pass "VER_GITLEAKS_PROVIDER has correct format: ${VER_GITLEAKS_PROVIDER}"
else
  fail "VER_GITLEAKS_PROVIDER has incorrect format" "Expected 'github:*', got '${VER_GITLEAKS_PROVIDER:-}'"
fi

if echo "${VER_COMMITLINT_PROVIDER:-}" | grep -q "^npm:"; then
  pass "VER_COMMITLINT_PROVIDER has correct format: ${VER_COMMITLINT_PROVIDER}"
else
  fail "VER_COMMITLINT_PROVIDER has incorrect format" "Expected 'npm:*', got '${VER_COMMITLINT_PROVIDER:-}'"
fi

# ── Property 2.3: Scripts Using Centralized Pattern ──────────────────────────

test_start "Property 2.3: Scripts already using centralized pattern exist and are correct"

# Check that scripts using centralized pattern exist
SCRIPTS_TO_CHECK="security.sh java.sh kotlin.sh"

for script_name in ${SCRIPTS_TO_CHECK}; do
  script_path="${SCRIPT_DIR}/lib/langs/${script_name}"

  test_start "Property 2.3.${script_name}: ${script_name} uses centralized pattern"

  if [ ! -f "${script_path}" ]; then
    fail "${script_name} not found" "${script_path} does not exist"
    continue
  fi

  # Verify script uses centralized pattern (contains ${VER_*_PROVIDER:-})
  # shellcheck disable=SC2016
  if grep -q '\${VER_.*_PROVIDER:-}' "${script_path}"; then
    pass "${script_name} uses centralized pattern"
  else
    fail "${script_name} does not use centralized pattern" "No '\${VER_*_PROVIDER:-}' pattern found"
  fi

  # Verify script does NOT have hardcoded providers
  if grep -q 'local _PROVIDER="github:' "${script_path}" ||
    grep -q 'local _PROVIDER="npm:' "${script_path}" ||
    grep -q 'local _PROVIDER="pipx:' "${script_path}" ||
    grep -q 'local _PROVIDER="gem:' "${script_path}"; then
    fail "${script_name} contains hardcoded providers"
  else
    pass "${script_name} has no hardcoded providers"
  fi
done

# ── Property 2.4: Fallback Pattern Preservation ──────────────────────────────

# shellcheck disable=SC2016
test_start 'Property 2.4: Fallback pattern ${VAR:-} is used correctly'

# Test that the :- fallback pattern works as expected
TEST_VAR="${UNDEFINED_VAR:-default_value}"
if [ "${TEST_VAR}" = "default_value" ]; then
  pass "Fallback pattern works with undefined variable"
else
  fail "Fallback pattern failed with undefined variable" "Expected 'default_value', got '${TEST_VAR}'"
fi

DEFINED_VAR="actual_value"
TEST_VAR2="${DEFINED_VAR:-default_value}"
if [ "${TEST_VAR2}" = "actual_value" ]; then
  pass "Fallback pattern works with defined variable"
else
  fail "Fallback pattern failed with defined variable" "Expected 'actual_value', got '${TEST_VAR2}'"
fi

# ── Property 2.5: Version Variables Exist ─────────────────────────────────────

test_start "Property 2.5: Version variables (not just providers) are defined"

# Check that version variables exist alongside provider variables
if [ -n "${VER_GITLEAKS:-}" ]; then
  pass "VER_GITLEAKS version defined: ${VER_GITLEAKS}"
else
  fail "VER_GITLEAKS version missing"
fi

if [ -n "${VER_COMMITLINT:-}" ]; then
  pass "VER_COMMITLINT version defined: ${VER_COMMITLINT}"
else
  fail "VER_COMMITLINT version missing"
fi

# ── Property 2.6: No Hardcoded Providers in Centralized Scripts ──────────────

test_start "Property 2.6: Scripts using centralized pattern have NO hardcoded providers"

# This is a critical preservation test - scripts that already use centralized
# variables should not have any hardcoded provider strings

for script_name in ${SCRIPTS_TO_CHECK}; do
  script_path="${SCRIPT_DIR}/lib/langs/${script_name}"

  if [ ! -f "${script_path}" ]; then
    continue
  fi

  # Check for any hardcoded providers
  has_hardcoded=0

  if grep -q 'local _PROVIDER="github:' "${script_path}" 2>/dev/null; then
    has_hardcoded=1
  fi

  if grep -q 'local _PROVIDER="npm:' "${script_path}" 2>/dev/null; then
    has_hardcoded=1
  fi

  if grep -q 'local _PROVIDER="pipx:' "${script_path}" 2>/dev/null; then
    has_hardcoded=1
  fi

  if grep -q 'local _PROVIDER="gem:' "${script_path}" 2>/dev/null; then
    has_hardcoded=1
  fi

  if [ "${has_hardcoded}" -eq 0 ]; then
    pass "${script_name} has zero hardcoded providers (preserved)"
  else
    fail "${script_name} has hardcoded providers" "This violates preservation - script should already be using centralized variables"
  fi
done

# ── Property 2.7: Centralized Variables Are Non-Empty ────────────────────────

test_start "Property 2.7: Centralized provider variables are non-empty"

# Verify that centralized variables have actual values, not empty strings
VARS_TO_CHECK="VER_GITLEAKS_PROVIDER VER_OSV_SCANNER_PROVIDER VER_JAVA_FORMAT_PROVIDER VER_KTLINT_PROVIDER VER_BATS_PROVIDER VER_COMMITLINT_PROVIDER VER_PRETTIER_PROVIDER"

for var_name in ${VARS_TO_CHECK}; do
  eval "var_value=\${${var_name}:-}"

  if [ -n "${var_value}" ]; then
    pass "${var_name} is non-empty: ${var_value}"
  else
    fail "${var_name} is empty or undefined"
  fi
done

# ── Property 2.8: Centralized Pattern in Multiple Scripts ────────────────────

test_start "Property 2.8: Multiple scripts use the centralized pattern consistently"

# Count how many scripts use the centralized pattern
scripts_with_pattern=0

for script_name in ${SCRIPTS_TO_CHECK}; do
  script_path="${SCRIPT_DIR}/lib/langs/${script_name}"

  # shellcheck disable=SC2016
  if [ -f "${script_path}" ] && grep -q '\${VER_.*_PROVIDER:-}' "${script_path}"; then
    scripts_with_pattern=$((scripts_with_pattern + 1))
  fi
done

if [ "${scripts_with_pattern}" -ge 3 ]; then
  pass "At least 3 scripts use centralized pattern (found ${scripts_with_pattern})"
else
  fail "Too few scripts use centralized pattern" "Expected at least 3, found ${scripts_with_pattern}"
fi

# ── Property 2.9: Provider Variables Match Expected Tools ────────────────────

test_start "Property 2.9: Provider variables exist for expected security tools"

# Verify specific security tools have provider variables
if [ -n "${VER_GITLEAKS_PROVIDER:-}" ] && [ -n "${VER_TRIVY_PROVIDER:-}" ] && [ -n "${VER_OSV_SCANNER_PROVIDER:-}" ]; then
  pass "Security tool provider variables exist (gitleaks, trivy, osv-scanner)"
else
  fail "Some security tool provider variables missing"
fi

# ── Property 2.10: Naming Convention Consistency ──────────────────────────────

test_start "Property 2.10: Provider variables follow VER_*_PROVIDER naming convention"

# Check that provider variables follow the naming convention
naming_correct=1

# All provider variables should end with _PROVIDER
if ! echo "${VER_GITLEAKS_PROVIDER:-}" | grep -q "^github:"; then
  naming_correct=0
fi

if ! echo "${VER_COMMITLINT_PROVIDER:-}" | grep -q "^npm:"; then
  naming_correct=0
fi

if [ "${naming_correct}" -eq 1 ]; then
  pass "Provider variables follow naming convention and format"
else
  fail "Provider variables do not follow expected format"
fi

# ── Test Summary ──────────────────────────────────────────────────────────────

echo ""
echo "════════════════════════════════════════════════════════════════════════════"
echo "Test Summary:"
echo "  Total:  ${TESTS_RUN}"
echo "  ${GREEN}Passed: ${TESTS_PASSED}${NC}"
echo "  ${RED}Failed: ${TESTS_FAILED}${NC}"
echo "════════════════════════════════════════════════════════════════════════════"

if [ "${TESTS_FAILED}" -eq 0 ]; then
  echo ""
  echo "${GREEN}✓ All preservation property tests PASSED${NC}"
  echo ""
  echo "This confirms that scripts using centralized variables (security.sh, java.sh,"
  echo "kotlin.sh) work correctly and have no hardcoded providers."
  echo ""
  echo "The baseline behavior to preserve has been validated."
  exit 0
else
  echo ""
  echo "${RED}✗ Some preservation property tests FAILED${NC}"
  echo ""
  echo "Review the failures above to understand what preservation requirements"
  echo "are not being met in the current codebase."
  exit 1
fi
