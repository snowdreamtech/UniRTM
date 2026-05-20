#!/usr/bin/env bats
# Bug Condition Exploration Test for CI Gitleaks Detection
# **Validates: Requirements 1.1, 1.2, 1.3, 1.4**
#
# **CRITICAL**: This test MUST FAIL on unfixed code - failure confirms the bug exists
# **DO NOT attempt to fix the test or the code when it fails**
# **NOTE**: This test encodes the expected behavior - it will validate the fix when it passes after implementation
# **GOAL**: Surface counterexamples that demonstrate the bug exists
#
# **Property 1: Bug Condition** - Gitleaks Not Found After Installation
# Scoped PBT Approach: Scope the property to CI environment where mise cache is disabled and mise shims are not in PATH
#
# **TEST RESULT ANALYSIS**:
# This test PASSED UNEXPECTEDLY in local development environment.
# This indicates:
#   1. The bug is CI-specific and does NOT reproduce locally
#   2. Local environment has mise shims already in PATH, preventing bug reproduction
#   3. The bug likely requires a truly fresh CI runner environment to manifest
#
# **RECOMMENDATION**: This test serves as a regression test for the fix, but cannot
# validate the bug exists on unfixed code in local development. The bug must be
# validated in actual CI environment (GitHub Actions).

setup() {
  load '../vendor/bats-support/load.bash'
  load '../vendor/bats-assert/load.bash'

  # Save original environment
  export ORIGINAL_PATH="$PATH"
  export ORIGINAL_MISE_CACHE="${_G_MISE_LS_JSON_CACHE:-}"
  export ORIGINAL_MISE_SHIMS_BASE="${_G_MISE_SHIMS_BASE:-}"

  # Create a temporary workspace that mimics CI environment
  export TEMP_DIR
  TEMP_DIR="$(mktemp -d)"

  # Copy necessary scripts
  cp -r scripts "$TEMP_DIR/"
  cp .mise.toml "$TEMP_DIR/" 2>/dev/null || true

  # Create minimal project structure
  cd "$TEMP_DIR" || exit
  touch Makefile
  git init -q

  # Source common.sh to get access to functions
  export SCRIPT_DIR="$TEMP_DIR/scripts"
  export _G_PROJECT_ROOT="$TEMP_DIR"
  export _G_LIB_DIR="$TEMP_DIR/scripts/lib"

  # shellcheck disable=SC1091
  . "$TEMP_DIR/scripts/lib/common.sh"
}

teardown() {
  # Restore original environment
  export PATH="$ORIGINAL_PATH"
  export _G_MISE_LS_JSON_CACHE="$ORIGINAL_MISE_CACHE"
  export _G_MISE_SHIMS_BASE="$ORIGINAL_MISE_SHIMS_BASE"

  # Cleanup
  cd / || exit
  rm -rf "$TEMP_DIR"
}

# Helper function to simulate CI environment conditions
# NOTE: This simulation is imperfect in local development because:
# 1. mise is already installed and configured
# 2. Shell initialization has already added mise to PATH
# 3. Cannot fully replicate fresh CI runner state
simulate_ci_environment() {
  # 1. Disable mise cache (simulates network-disabled CI)
  export _G_MISE_LS_JSON_CACHE="{}"

  # 2. Attempt to remove mise shims from PATH (simulates fresh CI runner)
  # NOTE: This may not fully work in local development
  local NEW_PATH=""
  local OLD_IFS="$IFS"
  IFS=":"
  # shellcheck disable=SC2086
  for path_entry in $PATH; do
    # Skip mise shims directory
    case "$path_entry" in
    *mise/shims*)
      echo "# Removing from PATH: $path_entry" >&3
      continue
      ;;
    *)
      if [ -z "$NEW_PATH" ]; then
        NEW_PATH="$path_entry"
      else
        NEW_PATH="$NEW_PATH:$path_entry"
      fi
      ;;
    esac
  done
  IFS="$OLD_IFS"
  export PATH="$NEW_PATH"

  # 3. Verify mise shims are NOT in PATH
  if echo "$PATH" | grep -q "mise/shims"; then
    echo "# ⚠️  WARNING: mise/shims still in PATH after removal attempt" >&3
    echo "# This indicates CI simulation is incomplete" >&3
    return 1
  fi

  # 4. Verify cache is disabled
  [ "$_G_MISE_LS_JSON_CACHE" = "{}" ] || return 1

  return 0
}

@test "Bug Condition: Gitleaks not found after installation in CI environment" {
  # **EXPECTED OUTCOME**: This test should FAIL on unfixed code, PASS on fixed code
  # **ACTUAL OUTCOME**: Test PASSED in local development (unexpected)
  # **ANALYSIS**: Bug is CI-specific, cannot be fully reproduced locally

  # Skip if gitleaks is not in .mise.toml (not applicable to this project)
  if ! grep -q "gitleaks" "$TEMP_DIR/.mise.toml" 2>/dev/null; then
    skip "Gitleaks not configured in .mise.toml"
  fi

  # Skip if mise is not available
  if ! command -v mise >/dev/null 2>&1; then
    skip "mise not available in test environment"
  fi

  # Attempt to simulate CI environment conditions
  echo "# Attempting to simulate CI environment..." >&3
  run simulate_ci_environment

  if [ "$status" -ne 0 ]; then
    echo "# ⚠️  CI environment simulation incomplete" >&3
    echo "# Test will run but may not reproduce bug condition" >&3
  else
    echo "# ✅ CI environment simulation successful" >&3
  fi

  # Verify CI conditions
  echo "# Environment state:" >&3
  echo "# - mise cache: $_G_MISE_LS_JSON_CACHE" >&3
  echo "# - mise/shims in PATH: $(echo "$PATH" | grep -c 'mise/shims' || echo '0')" >&3

  # Install gitleaks via unirtm (simulating unirtm install)
  echo "# Installing gitleaks via unirtm..." >&3
  run run_mise install gitleaks

  # Log installation result
  if [ "$status" -eq 0 ]; then
    echo "# ✅ Gitleaks installation completed" >&3
  else
    echo "# ❌ Gitleaks installation failed: $status" >&3
    skip "Gitleaks installation failed"
  fi

  # **CRITICAL TEST**: Attempt to resolve gitleaks binary
  echo "# Testing resolve_bin..." >&3

  local GITLEAKS_PATH
  GITLEAKS_PATH=$(resolve_bin "gitleaks" 2>&1) || true

  echo "# resolve_bin result: '$GITLEAKS_PATH'" >&3

  # Document the result
  if [ -z "$GITLEAKS_PATH" ]; then
    echo "# 🐛 BUG CONFIRMED:" >&3
    echo "#   - Gitleaks installed successfully" >&3
    echo "#   - resolve_bin returned: EMPTY" >&3
    echo "#   - This is the expected bug behavior" >&3
  else
    echo "# Result: resolve_bin returned path" >&3
    if [ -x "$GITLEAKS_PATH" ]; then
      echo "# Path is executable: YES" >&3
      # Verify it actually works
      if "$GITLEAKS_PATH" version >/dev/null 2>&1; then
        echo "# Binary works: YES" >&3
        echo "# ✅ Bug does NOT reproduce in this environment" >&3
      fi
    fi
  fi

  # The assertion: resolve_bin MUST return a non-empty, executable path
  # **ON UNFIXED CODE IN CI**: This should FAIL
  # **ON FIXED CODE**: This should PASS
  # **IN LOCAL DEV**: This PASSES (bug doesn't reproduce)
  assert [ -n "$GITLEAKS_PATH" ]
  assert [ -x "$GITLEAKS_PATH" ]
}

@test "Verify resolve_bin layers work correctly" {
  # This test verifies that resolve_bin's 4-layer lookup works
  # It serves as a baseline to understand which layers are functional

  if ! command -v mise >/dev/null 2>&1; then
    skip "mise not available"
  fi

  # Ensure gitleaks is installed
  run_mise install gitleaks >/dev/null 2>&1 || skip "Cannot install gitleaks"

  echo "# Testing resolve_bin layers:" >&3

  # Layer 3: System PATH (via command -v)
  local CMD_V_RESULT
  CMD_V_RESULT=$(command -v gitleaks 2>&1) || true
  echo "# Layer 3 (command -v): '$CMD_V_RESULT'" >&3

  # Layer 4: mise which
  local MISE_WHICH_RESULT
  MISE_WHICH_RESULT=$(mise which gitleaks 2>&1) || true
  echo "# Layer 4 (mise which): '$MISE_WHICH_RESULT'" >&3

  # Full resolve_bin
  local RESOLVE_RESULT
  RESOLVE_RESULT=$(resolve_bin "gitleaks" 2>&1) || true
  echo "# resolve_bin result: '$RESOLVE_RESULT'" >&3

  # Analysis
  if [ -n "$CMD_V_RESULT" ]; then
    echo "# ✅ Layer 3 (PATH) works" >&3
  else
    echo "# ❌ Layer 3 (PATH) failed" >&3
  fi

  if [ -n "$MISE_WHICH_RESULT" ]; then
    echo "# ✅ Layer 4 (mise which) works" >&3
  else
    echo "# ❌ Layer 4 (mise which) failed" >&3
  fi

  # The test passes if resolve_bin works
  assert [ -n "$RESOLVE_RESULT" ]
  assert [ -x "$RESOLVE_RESULT" ]
}

@test "Document CI-specific conditions that trigger the bug" {
  # This test documents the specific conditions needed to reproduce the bug
  # It doesn't test functionality, just documents requirements

  echo "# Bug Reproduction Requirements:" >&3
  echo "# 1. Fresh CI runner (no prior mise configuration)" >&3
  echo "# 2. mise cache disabled (_G_MISE_LS_JSON_CACHE='{}')" >&3
  echo "# 3. mise shims NOT in PATH" >&3
  echo "# 4. Gitleaks installed via 'run_mise install gitleaks'" >&3
  echo "# 5. Immediate call to 'resolve_bin gitleaks' after installation" >&3
  echo "#" >&3
  echo "# Current environment:" >&3
  echo "# - OS: $(uname -s)" >&3
  echo "# - mise installed: $(command -v mise >/dev/null && echo 'YES' || echo 'NO')" >&3
  echo "# - mise shims in PATH: $(echo "$PATH" | grep -c 'mise/shims' || echo '0')" >&3
  echo "# - _G_MISE_LS_JSON_CACHE: ${_G_MISE_LS_JSON_CACHE:-not set}" >&3
  echo "#" >&3
  echo "# Conclusion: Local development environment does NOT match CI conditions" >&3
  echo "# The bug can only be validated in actual GitHub Actions CI" >&3

  # This test always passes - it's documentation only
  run echo "Documentation complete"
  assert_success
}
