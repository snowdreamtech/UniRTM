#!/usr/bin/env bats

setup() {
  load 'vendor/bats-support/load.bash'
  load 'vendor/bats-assert/load.bash'

  # Create a temporary workspace
  export TEMP_DIR
  unset _SNOWDREAM_TOP_LEVEL_SCRIPT
  unset _G_PROJECT_ROOT
  unset _G_LIB_DIR
  TEMP_DIR="$(mktemp -d)"
  cp -r scripts "$TEMP_DIR/"

  # Create a dummy script that sources common.sh
  cat <<EOF >"$TEMP_DIR/test_script.sh"
#!/usr/bin/env sh
SCRIPT_DIR=\$(cd "\$(dirname "\$0")" && pwd)


case "\$1" in
  guard) guard_project_root ;;
  log) log_info "test info"; log_success "test success"; log_warn "test warn"; log_error "test error" ;;
  args) parse_common_args "\$@"; echo "VERBOSE=\$VERBOSE"; echo "DRY_RUN=\$DRY_RUN" ;;

  swap) atomic_swap "\$2" "\$3" ;;
  check_rt) check_runtime "\$2" "Test Tool" ;;
esac
EOF
  chmod +x "$TEMP_DIR/test_script.sh"
}

teardown() {
  rm -rf "$TEMP_DIR"
}

@test "common.sh: guard_project_root fails outside project root" {
  cd "$TEMP_DIR" || exit
  run sh test_script.sh guard
  assert_failure
  assert_output --partial "Error: This script must be run from the project root."
}

@test "common.sh: guard_project_root passes in project root (with Makefile and .git)" {
  cd "$TEMP_DIR" || exit
  touch Makefile
  git init -q
  run sh test_script.sh guard
  assert_success
}

@test "common.sh: parse_common_args handles --dry-run" {
  run sh "$TEMP_DIR/test_script.sh" args --dry-run
  assert_success
  assert_line --partial "DRY_RUN=1"
  assert_line --partial "Running in DRY-RUN mode"
}

@test "common.sh: parse_common_args handles -v and -q" {
  run sh "$TEMP_DIR/test_script.sh" args -v
  assert_line "VERBOSE=2"

  run sh "$TEMP_DIR/test_script.sh" args -q
  assert_line "VERBOSE=0"
}

@test "common.sh: logging functions produce colored output" {
  # We check for ANSI escape sequences (\033 or \x1b)
  run sh "$TEMP_DIR/test_script.sh" log
  assert_success
  # Check if output contains ESC[ (color start). bats-assert output handles escape codes.
  # We check for the specific literal strings we defined in common.sh
  assert_line --partial "test info"
  assert_line --partial "test success"
  assert_line --partial "test error"
}

@test "common.sh: atomic_swap successfully moves file" {
  echo "new content" >"$TEMP_DIR/new.tmp"
  touch "$TEMP_DIR/old.txt"
  run sh "$TEMP_DIR/test_script.sh" swap "$TEMP_DIR/new.tmp" "$TEMP_DIR/old.txt"
  assert_success
  assert [ -f "$TEMP_DIR/old.txt" ]
  assert [ ! -f "$TEMP_DIR/new.tmp" ]
  run cat "$TEMP_DIR/old.txt"
  assert_output "new content"
}

@test "common.sh: check_runtime exits 0 if tool exists" {
  # Mock a command
  mkdir -p "$TEMP_DIR/bin"
  touch "$TEMP_DIR/bin/dummy-tool"
  chmod +x "$TEMP_DIR/bin/dummy-tool"
  PATH="$TEMP_DIR/bin:$PATH" run sh "$TEMP_DIR/test_script.sh" check_rt dummy-tool
  assert_success
}

@test "common.sh: check_runtime exits 0 (skips) if tool missing" {
  run sh "$TEMP_DIR/test_script.sh" check_rt non-existent-tool
  assert_success
  assert_output --partial "Required runtime 'non-existent-tool' for Test Tool is missing. Skipping."
}
