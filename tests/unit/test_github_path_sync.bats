#!/usr/bin/env bats
# Bug Condition Exploration Test for GitHub PATH Same-Step Sync
# Validates: Requirements 1.1, 1.2, 1.3, 1.4
#
# **Validates: Requirements 1.1, 1.2, 1.3, 1.4**
#
# CRITICAL: This test MUST FAIL on unfixed code - failure confirms the bug exists
# DO NOT attempt to fix the test or the code when it fails
# NOTE: This test encodes the expected behavior - it will validate the fix when it passes after implementation
# GOAL: Surface counterexamples that demonstrate the bug exists
#
# Property 1: Bug Condition - CI environment same-step PATH synchronization
# Scoped PBT Approach: Scope the property to specific failing scenario where:
# - CI environment (GITHUB_PATH is set)
# - run_mise install succeeds and writes to GITHUB_PATH file
# - Current shell PATH does NOT contain the tool path (BUG)
# - resolve_bin cannot find the tool immediately (BUG)
#
# Expected Behavior (after fix): After successful run_mise install in CI,
# the tool path should be:
# 1. Written to GITHUB_PATH file (for cross-step persistence)
# 2. Added to current shell PATH (for same-step availability)
# 3. Immediately resolvable via resolve_bin

setup() {
  load '../vendor/bats-support/load.bash'
  load '../vendor/bats-assert/load.bash'

  # Save original environment
  export ORIGINAL_PATH="$PATH"
  export ORIGINAL_GITHUB_PATH="${GITHUB_PATH:-}"
  export ORIGINAL_CI="${CI:-}"
  export ORIGINAL_GITHUB_ACTIONS="${GITHUB_ACTIONS:-}"

  # Create a temporary workspace
  export TEMP_DIR
  TEMP_DIR="$(mktemp -d)"

  # Create temporary GITHUB_PATH file
  export GITHUB_PATH="$TEMP_DIR/github_path"
  touch "$GITHUB_PATH"

  # Copy necessary scripts
  cp -r scripts "$TEMP_DIR/"
  cp .mise.toml "$TEMP_DIR/" 2>/dev/null || true

  # Create minimal project structure
  cd "$TEMP_DIR" || exit
  touch Makefile
  git init -q

  # Set up CI environment
  export CI="true"
  export GITHUB_ACTIONS="true"

  # Source common.sh to get access to functions
  export SCRIPT_DIR="$TEMP_DIR/scripts"
  export _G_PROJECT_ROOT="$TEMP_DIR"
  export _G_LIB_DIR="$TEMP_DIR/scripts/lib"

  # shellcheck disable=SC1091

}

teardown() {
  # Restore original environment
  export PATH="$ORIGINAL_PATH"
  if [ -n "$ORIGINAL_GITHUB_PATH" ]; then
    export GITHUB_PATH="$ORIGINAL_GITHUB_PATH"
  else
    unset GITHUB_PATH
  fi
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

  # Cleanup
  cd / || exit
  rm -rf "$TEMP_DIR"
}

# Helper function to remove tool paths from current PATH
remove_tool_from_path() {
  local tool_pattern="${1:-mise}"
  local NEW_PATH=""
  local OLD_IFS="$IFS"
  IFS=":"
  for path_entry in $PATH; do
    case "$path_entry" in
    *"$tool_pattern"*)
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
}

@test "Bug Condition: Tool path written to GITHUB_PATH but not synced to current shell PATH" {
  # EXPECTED OUTCOME: This test should FAIL on unfixed code, PASS on fixed code
  # Bug Condition: In CI environment, after run_mise install:
  # - Tool path IS written to GITHUB_PATH file
  # - Tool path IS NOT in current shell PATH (BUG)
  # - resolve_bin CANNOT find the tool (BUG)

  if ! command -v mise >/dev/null 2>&1; then
    skip "mise not available in test environment"
  fi

  # Use pipx:zizmor as test tool (as specified in task details)
  local TEST_TOOL="pipx:zizmor"
  local TOOL_NAME="zizmor"

  echo "# Test Setup:" >&3
  echo "# - CI Environment: CI=$CI, GITHUB_ACTIONS=$GITHUB_ACTIONS" >&3
  echo "# - GITHUB_PATH file: $GITHUB_PATH" >&3
  echo "# - Test Tool: $TEST_TOOL" >&3

  # Remove any existing tool paths from PATH to simulate fresh CI step
  echo "# Removing existing tool paths from PATH..." >&3
  remove_tool_from_path "mise"
  remove_tool_from_path "zizmor"

  echo "# - Current PATH (cleaned): $PATH" >&3

  # Clear GITHUB_PATH file
  true >"$GITHUB_PATH"

  # Install tool via run_mise
  # CRITICAL: Do NOT use 'run' command here because it creates a subshell
  # and PATH modifications won't propagate back to the test shell.
  # We need to call run_mise directly to test same-shell PATH synchronization.
  echo "# Installing $TEST_TOOL via run_mise..." >&3
  if ! run_mise install "$TEST_TOOL" >&3 2>&3; then
    echo "# Installation failed" >&3
    skip "$TEST_TOOL installation failed"
  fi

  echo "# Installation completed successfully" >&3

  # VERIFICATION 1: Check if tool path was written to GITHUB_PATH file
  echo "# Checking GITHUB_PATH file contents..." >&3
  local GITHUB_PATH_CONTENTS
  GITHUB_PATH_CONTENTS=$(cat "$GITHUB_PATH" 2>/dev/null || echo "")
  echo "# GITHUB_PATH contents:" >&3
  printf '%s\n' "$GITHUB_PATH_CONTENTS" | sed 's/^/#   /' >&3

  local TOOL_PATH_IN_FILE=0
  if [ -n "$GITHUB_PATH_CONTENTS" ] && echo "$GITHUB_PATH_CONTENTS" | grep -q "$TOOL_NAME"; then
    TOOL_PATH_IN_FILE=1
    echo "# ✓ Tool path found in GITHUB_PATH file" >&3
  else
    echo "# ✗ Tool path NOT found in GITHUB_PATH file" >&3
    skip "Tool path not written to GITHUB_PATH - prerequisite not met"
  fi

  # VERIFICATION 2: Check if tool path is in current shell PATH
  echo "# Checking current shell PATH..." >&3
  local TOOL_PATH_IN_CURRENT_PATH=0
  if echo "$PATH" | grep -q "$TOOL_NAME"; then
    TOOL_PATH_IN_CURRENT_PATH=1
    echo "# ✓ Tool path found in current PATH" >&3
  else
    echo "# ✗ BUG CONFIRMED: Tool path NOT in current PATH" >&3
    echo "#   - GITHUB_PATH file contains tool path" >&3
    echo "#   - Current shell PATH does NOT contain tool path" >&3
    echo "#   - This proves the bug: GITHUB_PATH not synced to current shell" >&3
  fi

  # VERIFICATION 3: Check if resolve_bin can find the tool
  echo "# Testing resolve_bin for $TOOL_NAME..." >&3
  local RESOLVED_PATH
  RESOLVED_PATH=$(resolve_bin "$TOOL_NAME" 2>&1) || true

  echo "# - resolve_bin result: '${RESOLVED_PATH:-EMPTY}'" >&3

  local TOOL_RESOLVABLE=0
  if [ -n "$RESOLVED_PATH" ] && [ -x "$RESOLVED_PATH" ]; then
    TOOL_RESOLVABLE=1
    echo "# ✓ resolve_bin found executable: $RESOLVED_PATH" >&3
  else
    echo "# ✗ BUG CONFIRMED: resolve_bin CANNOT find tool" >&3
    echo "#   - Expected: resolve_bin should find tool immediately after install" >&3
    echo "#   - Actual: resolve_bin returned empty or non-executable path" >&3
  fi

  # VERIFICATION 4: Check if command -v can find the tool
  echo "# Testing command -v for $TOOL_NAME..." >&3
  local COMMAND_V_PATH
  COMMAND_V_PATH=$(command -v "$TOOL_NAME" 2>&1) || true

  echo "# - command -v result: '${COMMAND_V_PATH:-EMPTY}'" >&3

  local TOOL_IN_COMMAND_V=0
  if [ -n "$COMMAND_V_PATH" ]; then
    TOOL_IN_COMMAND_V=1
    echo "# ✓ command -v found tool: $COMMAND_V_PATH" >&3
  else
    echo "# ✗ BUG CONFIRMED: command -v CANNOT find tool" >&3
  fi

  # SUMMARY
  echo "# Bug Condition Summary:" >&3
  echo "#   1. Tool path in GITHUB_PATH file: $TOOL_PATH_IN_FILE" >&3
  echo "#   2. Tool path in current PATH: $TOOL_PATH_IN_CURRENT_PATH" >&3
  echo "#   3. Tool resolvable viaresolve_bin: $TOOL_RESOLVABLE" >&3
  echo "#   4. Tool findable via command -v: $TOOL_IN_COMMAND_V" >&3

  # ASSERTIONS: These encode the expected behavior
  # ON UNFIXED CODE: These should FAIL
  # ON FIXED CODE: These should PASS

  # After successful run_mise install in CI, tool path MUST be in current PATH
  assert [ "$TOOL_PATH_IN_CURRENT_PATH" -eq 1 ]

  # After successful run_mise install in CI, resolve_bin MUST find the tool
  assert [ "$TOOL_RESOLVABLE" -eq 1 ]
  assert [ -n "$RESOLVED_PATH" ]
  assert [ -x"$RESOLVED_PATH" ]
}

@test "Bug Condition: Mise shims directory written to GITHUB_PATH but not synced to current shell PATH" {
  # EXPECTED OUTCOME: This test should FAIL on unfixed code, PASS on fixed code
  # Bug Condition: In CI environment, after run_mise install:
  # - Mise shims directory IS written to GITHUB_PATH file
  # - Mise shims directory IS NOT in current shell PATH (BUG)

  if ! command -v mise >/dev/null 2>&1; then
    skip "mise not available in test environment"
  fi

  if [ -z "${_G_MISE_SHIMS_BASE:-}" ]; then
    skip "_G_MISE_SHIMS_BASE not set - cannot test"
  fi

  local TEST_TOOL="shellcheck"

  echo "# Test Setup:" >&3
  echo "# - CI Environment: CI=$CI, GITHUB_ACTIONS=$GITHUB_ACTIONS" >&3
  echo "# - GITHUB_PATH file: $GITHUB_PATH" >&3
  echo "# - _G_MISE_SHIMS_BASE: $_G_MISE_SHIMS_BASE" >&3
  echo "# - Test Tool: $TEST_TOOL" >&3

  # Remove mise shims from PATH
  echo "# Removing mise shims from PATH..." >&3
  remove_tool_from_path "mise"

  # Clear GITHUB_PATH file
  true >"$GITHUB_PATH"

  # Install tool via run_mise
  # CRITICAL: Do NOT use 'run' command here because it creates a subshell
  # and PATH modifications won't propagate back to the test shell.
  # We need to call run_mise directly to test same-shell PATH synchronization.
  echo "# Installing $TEST_TOOL via run_mise..." >&3
  if ! run_mise install "$TEST_TOOL" >&3 2>&3; then
    echo "# Installation failed" >&3
    skip "$TEST_TOOL installation failed"
  fi

  echo "# Installation completed successfully" >&3

  # VERIFICATION 1: Check if mise shims path was written to GITHUB_PATH file
  echo "# Checking GITHUB_PATH file for mise shims..." >&3
  local SHIMS_IN_FILE=0
  if grep -qxF "$_G_MISE_SHIMS_BASE" "$GITHUB_PATH" 2>/dev/null; then
    SHIMS_IN_FILE=1
    echo "# ✓ Mise shims path found in GITHUB_PATH file" >&3
  else
    echo "# ✗ Mise shims path NOT found in GITHUB_PATH file" >&3
  fi

  # VERIFICATION 2: Check if mise shims path is in current shell PATH
  echo "# Checking current shell PATH for mise shims..." >&3
  local SHIMS_IN_CURRENT_PATH=0
  if echo "$PATH" | grep -qF "$_G_MISE_SHIMS_BASE"; then
    SHIMS_IN_CURRENT_PATH=1
    echo "# ✓ Mise shims path found in current PATH" >&3
  else
    echo "# ✗ BUG CONFIRMED: Mise shims path NOT in current PATH" >&3
    echo "#   - GITHUB_PATH file contains mise shims path" >&3
    echo "#   - Current shell PATH does NOT contain mise shims path" >&3
    echo "#   - This proves the bug: GITHUB_PATH not synced to current shell" >&3
  fi

  # SUMMARY
  echo "# Bug Condition Summary:" >&3
  echo "#   1. Mise shims in GITHUB_PATH file: $SHIMS_IN_FILE" >&3
  echo "#   2. Mise shims in current PATH: $SHIMS_IN_CURRENT_PATH" >&3

  # ASSERTIONS: These encode the expected behavior
  # ON UNFIXED CODE: These should FAIL
  # ON FIXED CODE: These should PASS

  # If mise shims were written to GITHUB_PATH, they MUST also be in current PATH
  if [ "$SHIMS_IN_FILE" -eq 1 ]; then
    assert [ "$SHIMS_IN_CURRENT_PATH" -eq 1 ]
  fi
}

@test "Document bug reproduction conditions and root cause" {
  # Documentation test - always passes

  echo "# Bug Reproduction Conditions:" >&3
  echo "# 1. CI environment (GITHUB_PATH environment variable is set)" >&3
  echo "# 2. Fresh shell session where tool paths are NOT in PATH" >&3
  echo "# 3. Call run_mise install <tool> (succeeds with exit code 0)" >&3
  echo "# 4. Tool path is written to GITHUB_PATH file" >&3
  echo "# 5. Immediately call resolve_bin <tool> or command -v <tool>" >&3
  echo "# 6. Expected: Tool should be found immediately" >&3
  echo "# 7. Actual (unfixed): Tool NOT found - resolve_bin returns empty" >&3
  echo "#" >&3
  echo "# Root Cause:" >&3
  echo "# - run_mise() writes tool paths to GITHUB_PATH file for cross-step persistence" >&3
  echo "# - GitHub Actions GITHUB_PATH mechanism only applies BETWEEN steps" >&3
  echo "# - GITHUB_PATH does NOT automatically update current shell's PATH" >&3
  echo "# - run_mise() does NOT manually sync GITHUB_PATH to current shell PATH" >&3
  echo "# - Result: Tools installed in same step are NOT immediately available" >&3
  echo "#" >&3
  echo "# Impact:" >&3
  echo "# - Affects ALL 20+ tools installed via mise in CI" >&3
  echo "# - Breaks same-step workflows: unirtm install && make check-env" >&3
  echo "# - Security scanners (Gitleaks, Zizmor, OSV-Scanner) not found" >&3
  echo "# - Code quality tools (Shellcheck, Shfmt, Actionlint) not found" >&3
  echo "#" >&3
  echo "# Expected Fix:" >&3
  echo "# - After writing to GITHUB_PATH file, immediately read it" >&3
  echo "# - Add all paths from GITHUB_PATH to current shell's export PATH" >&3
  echo "# - Maintain idempotency (don't duplicate paths already in PATH)" >&3
  echo "# - Ensure tools are available in same step AND subsequent steps" >&3

  run echo "Documentation complete"
  assert_success
}

@test "Document affected scenarios" {
  # Documentation test - always passes

  echo "# Affected Scenarios:" >&3
  echo "#" >&3
  echo "# Scenario 1: Same-step tool verification" >&3
  echo "#   Command: unirtm install && make check-env" >&3
  echo "#   Expected: check-env finds all installed tools" >&3
  echo "#   Actual (unfixed): check-env reports tools not found" >&3
  echo "#" >&3
  echo "# Scenario 2: Dynamic tool installation and immediate use" >&3
  echo "#   Command: run_mise install pipx:zizmor && zizmor --version" >&3
  echo "#   Expected: zizmor command executes successfully" >&3
  echo "#   Actual (unfixed): zizmor: command not found" >&3
  echo "#" >&3
  echo "# Scenario 3: CI workflow single-step execution" >&3
  echo "#   Workflow: Install tools -> Run security scans in same step" >&3
  echo "#   Expected: Security scans execute with installed tools" >&3
  echo "#   Actual (unfixed): Security scans fail - tools not found" >&3
  echo "#" >&3
  echo "# Scenario 4: resolve_bin immediate lookup" >&3
  echo "#   Command: run_mise install go:github.com/.../osv-scanner && resolve_bin osv-scanner" >&3
  echo "#   Expected: resolve_bin returns valid executable path" >&3
  echo "#   Actual (unfixed): resolve_bin returns empty string" >&3

  run echo "Documentation complete"
  assert_success
}
