#!/usr/bin/env bats

setup() {
  # Provide a mock unirtm binary and network tools in PATH
  MOCK_BIN_DIR="$(mktemp -d)"
  export PATH="${MOCK_BIN_DIR}:${PATH}"

  # Mock curl
  cat <<'EOF' >"${MOCK_BIN_DIR}/curl"
#!/usr/bin/env bash
ORIG_ARGS="$*"
if [[ "$ORIG_ARGS" == *"api.github.com"* ]]; then
    echo '{"tag_name": "v99.9.9"}'
    exit 0
fi

# Find the -o argument value
OUTFILE=""
while [ $# -gt 0 ]; do
    if [ "$1" = "-o" ]; then
        OUTFILE="$2"
        shift 2
    else
        shift
    fi
done

if [[ "$ORIG_ARGS" == *"/releases/download/"* ]]; then
    if [[ "$ORIG_ARGS" == *"checksums.txt"* ]]; then
        echo "fakechecksum unirtm_Darwin_arm64.tar.gz" > "$OUTFILE"
    else
        echo "dummy archive content" > "$OUTFILE"
    fi
    exit 0
fi
echo "Mock curl called: $ORIG_ARGS" >&2
exit 0
EOF
  chmod +x "${MOCK_BIN_DIR}/curl"

  # Mock tar
  cat <<'EOF' >"${MOCK_BIN_DIR}/tar"
#!/usr/bin/env bash
# Mock tar just creates a fake unirtm binary in the target directory
TARGET_DIR="${@: -1}"
mkdir -p "$TARGET_DIR"
touch "$TARGET_DIR/unirtm"
chmod +x "$TARGET_DIR/unirtm"
EOF
  chmod +x "${MOCK_BIN_DIR}/tar"

  # Mock unirtm itself for post-install verification
  cat <<'EOF' >"${MOCK_BIN_DIR}/unirtm"
#!/bin/sh
echo "unirtm version v99.9.9"
EOF
  chmod +x "${MOCK_BIN_DIR}/unirtm"
}

teardown() {
  rm -rf "${MOCK_BIN_DIR}"
}

@test "install.sh runs successfully with --help" {
  run ./install.sh --help
  [ "$status" -eq 0 ]
  [[ $output == *"Usage:"* ]]
}

@test "install.sh fails gracefully when --version is missing its argument" {
  # Temporarily disable set -u for the test environment caller if needed,
  # but the script itself uses set -u. It should fail with unbound variable.
  run ./install.sh --version
  [ "$status" -ne 0 ]
  [[ $output == *"unbound variable"* ]] || [[ $output == *"parameter not set"* ]]
}

@test "install.sh executes full mocked installation path" {
  TEST_INSTALL_DIR="$(mktemp -d)"
  run ./install.sh --install-dir "$TEST_INSTALL_DIR"

  if [ "$status" -ne 0 ]; then
    echo "Failed Output: $output" >&3
  fi

  # Check if the mock binary was "installed"
  [ "$status" -eq 0 ]
  [ -f "${TEST_INSTALL_DIR}/unirtm" ]

  rm -rf "$TEST_INSTALL_DIR"
}
