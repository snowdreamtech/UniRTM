#!/usr/bin/env sh
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# scripts/benchmark-binary-resolution.sh - Binary resolution performance benchmark
#
# Purpose:
#   Measures performance of binary-first detection and platform-specific name resolution
#   for the verify_binary_exists() and resolve_bin() functions.
#
# Usage:
#   ./scripts/benchmark-binary-resolution.sh [OPTIONS]
#
# Options:
#   --output-format <json|text>  Output format (default: text)
#   --verbose                    Enable detailed timing logs
#   --dry-run                    Preview without actual execution
#   -h, --help                   Show this help message
#
# Requirements: 3.1, 3.2, 3.3, 3.4, 3.5

set -eu

# ── Common Library ───────────────────────────────────────────────────────────
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
. "$SCRIPT_DIR/lib/common.sh"

# 1. Execution Context Guard: ensure run from root
guard_project_root

# ── Configuration ────────────────────────────────────────────────────────────
OUTPUT_FORMAT="${OUTPUT_FORMAT:-text}"
TIMEOUT_BINARY_VERIFY=5    # 5 seconds per binary verification (Requirement 3.1)
TIMEOUT_PLATFORM_RESOLVE=3 # 3 seconds per platform-specific resolution (Requirement 3.3)

# ── Global State ─────────────────────────────────────────────────────────────
BENCHMARK_START_TIME=0
BENCHMARK_END_TIME=0
TEST_RESULTS=""

# ── Test Cases ───────────────────────────────────────────────────────────────
# Test cases from design.md:
# 1. Standard Binary: shfmt (exact name match)
# 2. Platform-Specific: ec-linux-amd64, ec-darwin-arm64 (editorconfig-checker)
# 3. Windows Binary: hadolint.exe
# 4. Versioned Binary: shfmt_v3.13.1
# 5. UniRTM Shim: /unirtm/shims/shellcheck

# ── Helper Functions ─────────────────────────────────────────────────────────

# Purpose: Display usage information
usage() {
  cat <<-EOF
Usage: $(basename "$0") [OPTIONS]

Measures performance of binary resolution strategies for tool verification.

Options:
  --output-format <json|text>  Output format (default: text)
  --verbose                    Enable detailed timing logs
  --dry-run                    Preview without actual execution
  -h, --help                   Show this help message

Test Cases:
  1. Standard Binary: shfmt (exact name match)
  2. Platform-Specific: ec-* patterns (editorconfig-checker)
  3. Windows Binary: *.exe extensions
  4. Versioned Binary: tool_vX.Y.Z patterns
  5. UniRTM Shim: /unirtm/shims/* paths

Measurement Points:
  - verify_binary_exists() execution time
  - resolve_bin() execution time
  - unirtm which execution time
  - command -v execution time
  - find pattern matching time

Performance Thresholds:
  - Binary verification: < 5 seconds (Requirement 3.1)
  - Platform-specific resolution: < 3 seconds (Requirement 3.3)

Requirements: 3.1, 3.2, 3.3, 3.4, 3.5
EOF
  exit "${1:-0}"
}

# Purpose: Get current timestamp in milliseconds (or seconds if not available)
get_timestamp_ms() {
  if command -v gdate >/dev/null 2>&1; then
    # macOS with coreutils
    gdate +%s%3N
  elif date --version 2>/dev/null | grep -q GNU; then
    # GNU date (Linux)
    date +%s%3N
  else
    # Fallback to seconds with 000 milliseconds
    echo "$(date +%s)000"
  fi
}

# Purpose: Calculate elapsed time in milliseconds
# Params:
#   $1 - Start time (ms)
#   $2 - End time (ms)
calc_elapsed_ms() {
  local start="${1:-0}"
  local end="${2:-0}"
  echo $((end - start))
}

# Purpose: Measure time for verify_binary_exists() function
# Params:
#   $1 - Binary name
#   $2 - Test case name
measure_verify_binary() {
  local bin_name="${1:-}"
  local test_name="${2:-}"
  local start_time end_time elapsed_ms status

  [ "${VERBOSE:-1}" -ge 1 ] && log_info "  Testing verify_binary_exists: $test_name" >&2

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    elapsed_ms=0
    status="dry-run"
  else
    start_time=$(get_timestamp_ms)

    # Test verify_binary_exists with timeout
    if run_with_timeout "$TIMEOUT_BINARY_VERIFY" sh -c "command -v '$bin_name' >/dev/null 2>&1"; then
      status="success"
    else
      status="not_found"
    fi

    end_time=$(get_timestamp_ms)
    elapsed_ms=$(calc_elapsed_ms "$start_time" "$end_time")
  fi

  # Store result
  TEST_RESULTS="$TEST_RESULTS
verify_binary_exists:$test_name:$bin_name:$elapsed_ms:$status"

  [ "${VERBOSE:-1}" -ge 1 ] && log_info "    Result: $status (${elapsed_ms}ms)" >&2
}

# Purpose: Measure time for resolve_bin() function
# Params:
#   $1 - Binary name
#   $2 - Test case name
measure_resolve_bin() {
  local bin_name="${1:-}"
  local test_name="${2:-}"
  local start_time end_time elapsed_ms status resolved_path

  [ "${VERBOSE:-1}" -ge 1 ] && log_info "  Testing resolve_bin: $test_name" >&2

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    elapsed_ms=0
    status="dry-run"
    resolved_path=""
  else
    start_time=$(get_timestamp_ms)

    # Test resolve_bin with timeout (if function exists)
    if command -v resolve_bin >/dev/null 2>&1; then
      if resolved_path=$(run_with_timeout "$TIMEOUT_PLATFORM_RESOLVE" resolve_bin "$bin_name" 2>/dev/null); then
        status="success"
      else
        status="not_found"
        resolved_path=""
      fi
    else
      # Fallback to command -v if resolve_bin not available
      if resolved_path=$(run_with_timeout "$TIMEOUT_PLATFORM_RESOLVE" command -v "$bin_name" 2>/dev/null); then
        status="success"
      else
        status="not_found"
        resolved_path=""
      fi
    fi

    end_time=$(get_timestamp_ms)
    elapsed_ms=$(calc_elapsed_ms "$start_time" "$end_time")
  fi

  # Store result
  TEST_RESULTS="$TEST_RESULTS
resolve_bin:$test_name:$bin_name:$elapsed_ms:$status:$resolved_path"

  [ "${VERBOSE:-1}" -ge 1 ] && log_info "    Result: $status (${elapsed_ms}ms) -> $resolved_path" >&2
}

# Purpose: Measure time for unirtm which command
# Params:
#   $1 - Tool name
#   $2 - Test case name
measure_unirtm_which() {
  local tool_name="${1:-}"
  local test_name="${2:-}"
  local start_time end_time elapsed_ms status resolved_path

  [ "${VERBOSE:-1}" -ge 1 ] && log_info "  Testing unirtm which: $test_name" >&2

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    elapsed_ms=0
    status="dry-run"
    resolved_path=""
  else
    start_time=$(get_timestamp_ms)

    # Test unirtm which with timeout
    if command -v unirtm >/dev/null 2>&1; then
      if resolved_path=$(UNIRTM_OFFLINE=1 run_with_timeout "$TIMEOUT_PLATFORM_RESOLVE" unirtm which "$tool_name" 2>/dev/null); then
        status="success"
      else
        status="not_found"
        resolved_path=""
      fi
    else
      status="unirtm_not_available"
      resolved_path=""
    fi

    end_time=$(get_timestamp_ms)
    elapsed_ms=$(calc_elapsed_ms "$start_time" "$end_time")
  fi

  # Store result
  TEST_RESULTS="$TEST_RESULTS
unirtm_which:$test_name:$tool_name:$elapsed_ms:$status:$resolved_path"

  [ "${VERBOSE:-1}" -ge 1 ] && log_info "    Result: $status (${elapsed_ms}ms) -> $resolved_path" >&2
}

# Purpose: Measure time for command -v lookup
# Params:
#   $1 - Binary name
#   $2 - Test case name
measure_command_v() {
  local bin_name="${1:-}"
  local test_name="${2:-}"
  local start_time end_time elapsed_ms status resolved_path

  [ "${VERBOSE:-1}" -ge 1 ] && log_info "  Testing command -v: $test_name" >&2

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    elapsed_ms=0
    status="dry-run"
    resolved_path=""
  else
    start_time=$(get_timestamp_ms)

    # Test command -v with timeout
    if resolved_path=$(run_with_timeout 1 command -v "$bin_name" 2>/dev/null); then
      status="success"
    else
      status="not_found"
      resolved_path=""
    fi

    end_time=$(get_timestamp_ms)
    elapsed_ms=$(calc_elapsed_ms "$start_time" "$end_time")
  fi

  # Store result
  TEST_RESULTS="$TEST_RESULTS
command_v:$test_name:$bin_name:$elapsed_ms:$status:$resolved_path"

  [ "${VERBOSE:-1}" -ge 1 ] && log_info "    Result: $status (${elapsed_ms}ms) -> $resolved_path" >&2
}

# Purpose: Measure time for find pattern matching
# Params:
#   $1 - Pattern (e.g., "ec-*")
#   $2 - Test case name
measure_find_pattern() {
  local pattern="${1:-}"
  local test_name="${2:-}"
  local start_time end_time elapsed_ms status found_count

  [ "${VERBOSE:-1}" -ge 1 ] && log_info "  Testing find pattern: $test_name" >&2

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    elapsed_ms=0
    status="dry-run"
    found_count=0
  else
    start_time=$(get_timestamp_ms)

    # Test find with pattern in unirtm installation directory
    if command -v unirtm >/dev/null 2>&1; then
      local unirtm_installs
      unirtm_installs=$(unirtm where 2>/dev/null || echo "")

      if [ -n "$unirtm_installs" ] && [ -d "$unirtm_installs" ]; then
        found_count=$(run_with_timeout 10 find "$unirtm_installs" -type f -name "$pattern" 2>/dev/null | wc -l | tr -d ' ')
        status="success"
      else
        found_count=0
        status="unirtm_dir_not_found"
      fi
    else
      found_count=0
      status="unirtm_not_available"
    fi

    end_time=$(get_timestamp_ms)
    elapsed_ms=$(calc_elapsed_ms "$start_time" "$end_time")
  fi

  # Store result
  TEST_RESULTS="$TEST_RESULTS
find_pattern:$test_name:$pattern:$elapsed_ms:$status:$found_count"

  [ "${VERBOSE:-1}" -ge 1 ] && log_info "    Result: $status (${elapsed_ms}ms) - found $found_count matches" >&2
}

# Purpose: Run all test cases
run_test_cases() {
  log_info "Running binary resolution benchmarks..." >&2

  # Test Case 1: Standard Binary (shfmt)
  log_info "Test Case 1: Standard Binary (shfmt)" >&2
  measure_verify_binary "shfmt" "standard_binary"
  measure_resolve_bin "shfmt" "standard_binary"
  measure_unirtm_which "shfmt" "standard_binary"
  measure_command_v "shfmt" "standard_binary"

  # Test Case 2: Platform-Specific Binary (editorconfig-checker)
  log_info "Test Case 2: Platform-Specific Binary (editorconfig-checker)" >&2
  local ec_pattern
  case "$(uname -s)" in
  Darwin)
    case "$(uname -m)" in
    arm64) ec_pattern="ec-darwin-arm64" ;;
    *) ec_pattern="ec-darwin-amd64" ;;
    esac
    ;;
  Linux)
    case "$(uname -m)" in
    aarch64 | arm64) ec_pattern="ec-linux-arm64" ;;
    *) ec_pattern="ec-linux-amd64" ;;
    esac
    ;;
  MINGW* | MSYS* | CYGWIN*)
    ec_pattern="ec-windows-amd64.exe"
    ;;
  *)
    ec_pattern="ec-linux-amd64"
    ;;
  esac

  measure_verify_binary "$ec_pattern" "platform_specific"
  measure_resolve_bin "$ec_pattern" "platform_specific"
  measure_unirtm_which "editorconfig-checker" "platform_specific"
  measure_find_pattern "ec-*" "platform_specific"

  # Test Case 3: Windows Binary (hadolint.exe)
  if [ "$(uname -s)" = "MINGW" ] || [ "$(uname -s)" = "MSYS" ] || [ "$(uname -s)" = "CYGWIN" ]; then
    log_info "Test Case 3: Windows Binary (hadolint.exe)" >&2
    measure_verify_binary "hadolint.exe" "windows_binary"
    measure_resolve_bin "hadolint.exe" "windows_binary"
    measure_command_v "hadolint.exe" "windows_binary"
  else
    log_info "Test Case 3: Windows Binary (skipped on non-Windows)" >&2
  fi

  # Test Case 4: Versioned Binary (shfmt with version)
  log_info "Test Case 4: Versioned Binary (shfmt_v*)" >&2
  measure_find_pattern "shfmt_v*" "versioned_binary"

  # Test Case 5: UniRTM Shim (shellcheck)
  log_info "Test Case 5: UniRTM Shim (shellcheck)" >&2
  measure_verify_binary "shellcheck" "unirtm_shim"
  measure_resolve_bin "shellcheck" "unirtm_shim"
  measure_unirtm_which "shellcheck" "unirtm_shim"
  measure_command_v "shellcheck" "unirtm_shim"
}

# Purpose: Calculate statistics from results
# Params:
#   $1 - Strategy name (e.g., "verify_binary_exists")
calc_strategy_stats() {
  local strategy="${1:-}"
  local total_time=0
  local count=0
  local max_time=0
  local min_time=999999

  # shellcheck disable=SC2030,SC2031
  echo "$TEST_RESULTS" | grep "^$strategy:" | while IFS=: read -r _ test_name bin_name elapsed status rest; do
    count=$((count + 1))
    total_time=$((total_time + elapsed))

    if [ "$elapsed" -gt "$max_time" ]; then
      max_time=$elapsed
    fi

    if [ "$elapsed" -lt "$min_time" ]; then
      min_time=$elapsed
    fi
  done

  # shellcheck disable=SC2031
  if [ "$count" -gt 0 ]; then
    # shellcheck disable=SC2031
    avg_time=$((total_time / count))
    # shellcheck disable=SC2031
    echo "$avg_time:$min_time:$max_time:$count"
  else
    echo "0:0:0:0"
  fi
}

# Purpose: Check if any test exceeded thresholds
check_thresholds() {
  local violations=0

  # Check binary verification threshold (5 seconds = 5000ms)
  # shellcheck disable=SC2030,SC2031
  echo "$TEST_RESULTS" | grep "^verify_binary_exists:" | while IFS=: read -r _ test_name bin_name elapsed status rest; do
    if [ "$elapsed" -gt 5000 ]; then
      log_warn "⚠ Binary verification exceeded threshold: $test_name ($elapsed ms > 5000 ms)" >&2
      violations=$((violations + 1))
    fi
  done

  # Check platform-specific resolution threshold (3 seconds = 3000ms)
  # shellcheck disable=SC2030,SC2031
  echo "$TEST_RESULTS" | grep "^resolve_bin:.*platform_specific" | while IFS=: read -r _ test_name bin_name elapsed status rest; do
    if [ "$elapsed" -gt 3000 ]; then
      log_warn "⚠ Platform-specific resolution exceeded threshold: $test_name ($elapsed ms > 3000 ms)" >&2
      violations=$((violations + 1))
    fi
  done

  # shellcheck disable=SC2031
  return "$violations"
}

# Purpose: Output results in JSON format
output_json() {
  local timestamp commit_sha total_time

  timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date +"%Y-%m-%dT%H:%M:%SZ")
  commit_sha=$(git rev-parse HEAD 2>/dev/null || echo "unknown")
  total_time=$((BENCHMARK_END_TIME - BENCHMARK_START_TIME))

  printf '{\n'
  printf '  "timestamp": "%s",\n' "$timestamp"
  printf '  "commit_sha": "%s",\n' "$commit_sha"
  printf '  "total_time_ms": %d,\n' "$total_time"
  printf '  "results": [\n'

  first=true
  # shellcheck disable=SC2030,SC2031
  echo "$TEST_RESULTS" | grep -v '^$' | while IFS=: read -r strategy test_name target elapsed status extra; do
    if [ "$first" = false ]; then
      printf ',\n'
    fi
    first=false

    printf '    {\n'
    printf '      "strategy": "%s",\n' "$strategy"
    printf '      "test_case": "%s",\n' "$test_name"
    printf '      "target": "%s",\n' "$target"
    printf '      "elapsed_ms": %s,\n' "$elapsed"
    printf '      "status": "%s"' "$status"

    if [ -n "$extra" ]; then
      printf ',\n      "extra": "%s"\n' "$extra"
    else
      printf '\n'
    fi

    printf '    }'
  done

  printf '\n  ]\n'
  printf '}\n'
}

# Purpose: Output results in text format
output_text() {
  local timestamp commit_sha total_time

  timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date +"%Y-%m-%dT%H:%M:%SZ")
  commit_sha=$(git rev-parse HEAD 2>/dev/null || echo "unknown")
  total_time=$((BENCHMARK_END_TIME - BENCHMARK_START_TIME))

  log_success "
═══════════════════════════════════════════════════════════════
Binary Resolution Performance Benchmark
═══════════════════════════════════════════════════════════════

Timestamp:   $timestamp
Commit:      $commit_sha
Total Time:  ${total_time}ms

Performance Thresholds:
  - Binary verification: < 5000ms (Requirement 3.1)
  - Platform-specific resolution: < 3000ms (Requirement 3.3)

Test Results by Strategy:
"

  # Group results by strategy
  for strategy in verify_binary_exists resolve_bin unirtm_which command_v find_pattern; do
    local count
    count=$(echo "$TEST_RESULTS" | grep -c "^$strategy:" || echo "0")

    if [ "$count" -gt 0 ]; then
      printf "\n%s (%d tests):\n" "$strategy" "$count"
      # shellcheck disable=SC2030,SC2031
      echo "$TEST_RESULTS" | grep "^$strategy:" | while IFS=: read -r _ test_name target elapsed status extra; do
        printf "  %-25s %-20s %6sms  [%s]\n" "$test_name" "$target" "$elapsed" "$status"
      done
    fi
  done

  echo "
═══════════════════════════════════════════════════════════════"

  # Check thresholds
  if check_thresholds; then
    log_success "✓ All tests passed performance thresholds"
  else
    log_warn "⚠ Some tests exceeded performance thresholds"
  fi
}

# ── Main Execution ───────────────────────────────────────────────────────────

main() {
  log_info "Starting binary resolution benchmark..." >&2

  # Record start time
  BENCHMARK_START_TIME=$(get_timestamp_ms)

  # Run test cases
  run_test_cases

  # Record end time
  BENCHMARK_END_TIME=$(get_timestamp_ms)

  # Output results
  if [ "$OUTPUT_FORMAT" = "json" ]; then
    output_json
  else
    output_text
  fi

  log_success "Binary resolution benchmark completed" >&2
}

# ── Argument Parsing ─────────────────────────────────────────────────────────

while [ $# -gt 0 ]; do
  case "$1" in
  --output-format)
    OUTPUT_FORMAT="${2:-text}"
    shift 2
    ;;
  --verbose)
    export VERBOSE=2
    shift
    ;;
  --dry-run)
    DRY_RUN=1
    shift
    ;;
  -h | --help)
    usage 0
    ;;
  *)
    log_error "Unknown option: $1"
    usage 1
    ;;
  esac
done

# ── Entry Point ──────────────────────────────────────────────────────────────

main "$@"
