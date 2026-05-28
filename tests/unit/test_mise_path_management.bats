#!/usr/bin/env bats
# Bug Condition Exploration Test for Mise Tools PATH Management
# Validates: Requirements 1.1, 1.2, 1.3, 1.4, 1.5, 1.6
#
# CRITICAL: This test MUST FAIL on unfixed code - failure confirms the bug exists
# DO NOT attempt to fix the test or the code when it fails
# NOTE: This test encodes the expected behavior - it will validate the fix when it passes after implementation
# GOAL: Surface counterexamples that demonstrate the bug exists
#
# Property 1: Bug Condition - Automatic PATH Management After Mise Install
# Scoped PBT Approach: Scope the property to specific scenario where:
# - mise shims are NOT in current session PATH
# - run_mise install succeeds (exit code 0)
# - resolve_bin is called immediately after
#
# Expected Behavior: After successful run_mise install, the tool should be
# immediately resolvable via resolve_bin without manual PATH manipulation

setup() {
  load '../vendor/bats-support/load.bash'
  load '../vendor/bats-assert/load.bash'

  # Save original environment
  export ORIGINAL_PATH="$PATH"
  export ORIGINAL_MISE_CACHE="${_G_MISE_LS_JSON_CACHE:-}"
  export ORIGINAL_MISE_SHIMS_BASE="${_G_MISE_SHIMS_BASE:-}"

  # Create a temporary workspace
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

# Helper function to remove mise shims from PATH
remove_mise_from_path() {
  local NEW_PATH=""
  local OLD_IFS="$IFS"
  IFS=":"
  for path_entry in $PATH; do
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

  # Verify removal
  if echo "$PATH" | grep -q "mise/shims"; then
    echo "# WARNING: mise/shims still in PATH after removal" >&3
    return 1
  fi
  return 0
}

@test "Bug Condition: PATH not automatically managed after run_mise install" {
  # EXPECTED OUTCOME: This test should FAIL on unfixed code, PASS on fixed code
  # Bug Condition: After run_mise install succeeds, _G_MISE_SHIMS_BASE
  # is NOT automatically added to PATH, causing resolve_bin to fail

  if ! command -v mise >/dev/null 2>&1; then
    skip "mise not available in test environment"
  fi

  local TEST_TOOL="shellcheck"

  if ! grep -q "$TEST_TOOL" "$TEMP_DIR/.mise.toml" 2>/dev/null; then
    skip "$TEST_TOOL not configured in .mise.toml"
  fi

  echo "# Test Setup:" >&3
  echo "# - Tool: $TEST_TOOL" >&3

  # Remove mise shims from PATH to simulate fresh session
  echo "# Removing mise shims from PATH..." >&3
  run remove_mise_from_path
  if [ "$status" -ne 0 ]; then
    echo "# WARNING: Could not fully remove mise from PATH" >&3
  fi

  echo "# - mise/shims in PATH: $(echo "$PATH" | grep -c 'mise/shims' || echo '0')" >&3
  echo "# - _G_MISE_SHIMS_BASE: ${_G_MISE_SHIMS_BASE:-NOT SET}" >&3

  if [ -z "$_G_MISE_SHIMS_BASE" ]; then
    skip "_G_MISE_SHIMS_BASE not set - cannot test"
  fi

  # Install tool via run_mise
  echo "# Installing $TEST_TOOL via run_mise..." >&3
  run run_mise install "$TEST_TOOL"

  if [ "$status" -ne 0 ]; then
    echo "# Installation failed with exit code: $status" >&3
    skip "$TEST_TOOL installation failed"
  fi

  echo "# Installation completed successfully" >&3

  # CRITICAL TEST: Check if _G_MISE_SHIMS_BASE was added to PATH
  echo "# Checking PATH after installation..." >&3

  local SHIMS_IN_PATH=0
  if echo "$PATH" | grep -q "$_G_MISE_SHIMS_BASE"; then
    SHIMS_IN_PATH=1
    echo "# _G_MISE_SHIMS_BASE found in PATH" >&3
  else
    echo "# BUG CONFIRMED: _G_MISE_SHIMS_BASE NOT in PATH after installation" >&3
    echo "#   - Expected: $_G_MISE_SHIMS_BASE should be in PATH" >&3
    echo "#   - Actual: PATH does not contain mise shims" >&3
  fi

  # Test resolve_bin
  echo "# Testing resolve_bin for $TEST_TOOL..." >&3
  local TOOL_PATH
  TOOL_PATH=$(resolve_bin "$TEST_TOOL" 2>&1) || true

  echo "# - resolve_bin result: '${TOOL_PATH:-EMPTY}'" >&3

  if [ -z "$TOOL_PATH" ]; then
    echo "# BUG CONFIRMED: resolve_bin returned EMPTY" >&3
  elif [ -x "$TOOL_PATH" ]; then
    echo "# resolve_bin found executable: $TOOL_PATH" >&3
  else
    echo "# resolve_bin returned non-executable path" >&3
  fi

  # ASSERTION: After successful run_mise install, PATH must contain _G_MISE_SHIMS_BASE
  # ON UNFIXED CODE: This should FAIL
  # ON FIXED CODE: This should PASS
  assert [ "$SHIMS_IN_PATH" -eq 1 ]

  # ASSERTION: resolve_bin must return a valid executable path
  assert [ -n "$TOOL_PATH" ]
  assert [ -x "$TOOL_PATH" ]
}

@test "Document bug reproduction conditions" {
  # Documentation test - always passes

  echo "# Bug Reproduction Conditions:" >&3
  echo "# 1. Session where _G_MISE_SHIMS_BASE is NOT in PATH" >&3
  echo "# 2. Call run_mise install <tool> (succeeds with exit code 0)" >&3
  echo "# 3. Immediately call resolve_bin <tool>" >&3
  echo "# 4. Expected: resolve_bin should find the tool" >&3
  echo "# 5. Actual (unfixed): resolve_bin returns empty or fails" >&3
  echo "#" >&3
  echo "# Root Cause:" >&3
  echo "# - run_mise does NOT automatically add _G_MISE_SHIMS_BASE to PATH" >&3
  echo "# - resolve_bin relies on PATH (Layer 3) or mise which (Layer 4)" >&3
  echo "# - If PATH doesn't contain shims, Layer 3 fails" >&3
  echo "# - Layer 4 (mise which) may also fail if cache is disabled" >&3
  echo "#" >&3
  echo "# Affected Tools:" >&3
  echo "# - ALL tools installed via mise (20+ tools)" >&3
  echo "# - Particularly impacts CI environments with fresh runners" >&3

  run echo "Documentation complete"
  assert_success
}
