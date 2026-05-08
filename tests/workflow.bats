#!/usr/bin/env bats

setup() {
  load 'vendor/bats-support/load.bash'
  load 'vendor/bats-assert/load.bash'

  # Create a temporary workspace
  export TEMP_DIR
  unset _SNOWDREAM_TOP_LEVEL_SCRIPT
  unset _G_PROJECT_ROOT
  unset _G_LIB_DIR
  PROJECT_ROOT="$(pwd)"
  TEMP_DIR="$(mktemp -d)"
  # Copy functional scripts
  cp -r scripts "$TEMP_DIR/"

  # Create a dummy project structure
  cd "$TEMP_DIR" || exit
  touch Makefile LICENSE README.md package.json
  mkdir -p docs tests .agent/rules && touch .agent/rules/01-general.md
  echo "repos: []" >.pre-commit-config.yaml
  # Copy ruff.toml so format.sh uses the same line-length config (120) as the project
  cp "$PROJECT_ROOT/ruff.toml" "$TEMP_DIR/ruff.toml" 2>/dev/null || true
  # Copy .prettierrc so format.sh uses the same printWidth config (120) as the project
  cp "$PROJECT_ROOT/.prettierrc" "$TEMP_DIR/.prettierrc" 2>/dev/null || true
  git init -q

  # Create a dummy .mise.toml for get_mise_tool_version
  cat <<EOF >.mise.toml
[tools]
node = "20.18.3"
pnpm = "9.0.0"
python = "3.12.9"
gitleaks = "8.30.0"
osv-scanner = "2.3.3"
trivy = "0.69.3"
zizmor = "1.23.1"
shfmt-py = "3.12.0.2"
shellcheck-py = "0.11.0.1"
actionlint-py = "1.7.11.24"
editorconfig-checker = "3.6.1"
checkmake = "0.3.2"
EOF

  # Mock tools for check-env.sh (called by orchestration scripts)
  mkdir -p "$TEMP_DIR/bin"
  export PATH="$TEMP_DIR/bin:$PATH"

  printf '#!/usr/bin/env sh\necho "v20.18.3"\n' >"$TEMP_DIR/bin/node"
  printf '#!/usr/bin/env sh\necho "9.0.0"\n' >"$TEMP_DIR/bin/pnpm"
  printf '#!/usr/bin/env sh\necho "git version 2.30.0"\n' >"$TEMP_DIR/bin/git"
  printf '#!/usr/bin/env sh\necho "GNU Make 3.81"\n' >"$TEMP_DIR/bin/make"
  printf '#!/usr/bin/env sh\necho "3.6.1"\n' >"$TEMP_DIR/bin/editorconfig-checker"
  printf '#!/usr/bin/env sh\necho "v2026.3.8"\n' >"$TEMP_DIR/bin/mise"
  printf '#!/usr/bin/env sh\necho "v8.30.0"\n' >"$TEMP_DIR/bin/gitleaks"
  printf '#!/usr/bin/env sh\necho "v2.3.3"\n' >"$TEMP_DIR/bin/osv-scanner"
  printf '#!/usr/bin/env sh\necho "v0.69.3"\n' >"$TEMP_DIR/bin/trivy"
  printf '#!/usr/bin/env sh\necho "0.3.2"\n' >"$TEMP_DIR/bin/checkmake"
  printf '#!/usr/bin/env sh\necho "v1.23.1"\n' >"$TEMP_DIR/bin/zizmor"
  printf '#!/usr/bin/env sh\necho "3.12.0.2"\n' >"$TEMP_DIR/bin/shfmt"
  printf '#!/usr/bin/env sh\necho "0.11.0.1"\n' >"$TEMP_DIR/bin/shellcheck"
  printf '#!/usr/bin/env sh\necho "1.7.11.24"\n' >"$TEMP_DIR/bin/actionlint"
  chmod +x "$TEMP_DIR/bin/"*
}

teardown() {
  rm -rf "$TEMP_DIR"
}

@test "workflow: build.sh dry-run reports build activities" {
  run sh scripts/build.sh --dry-run
  assert_success
  assert_output --partial "Starting Project Build"
}

@test "workflow: format.sh dry-run reports formatting tasks" {
  run sh scripts/format.sh --dry-run
  assert_success
  assert_output --partial "Starting Unified Project Formatter"
}

@test "workflow: lint.sh dry-run reports linting scans" {
  run sh scripts/lint.sh --dry-run
  assert_success
  assert_output --partial "Starting Unified Project Linter"
}

@test "workflow: test.sh dry-run reports test execution" {
  run sh scripts/test.sh --dry-run
  assert_success
  assert_output --partial "Starting Unified Test Runner"
}

@test "workflow: docs.sh dry-run reports documentation generation" {
  run sh scripts/docs.sh --dry-run
  assert_success
  assert_output --partial "Documentation Manager"
}

@test "workflow: bench.sh dry-run reports benchmarking" {
  run sh scripts/bench.sh --dry-run
  assert_success
  assert_output --partial "Starting Performance Benchmarker"
}
