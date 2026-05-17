#!/usr/bin/env sh
# shellcheck disable=SC2034
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# scripts/lib/timeout.sh - Timeout mechanism module
#
# Purpose:
#   Provides unified timeout execution functions with zero-hang guarantees.
#   Supports multiple timeout implementations and ensures proper process cleanup.
#
# Standards:
#   - POSIX-compliant sh logic.
#   - "World Class" AI Documentation (English-only).
#   - Requirements: 2.3.1, 2.3.2, 2.4.1, 2.4.2

# ── 🕐 Timeout Configuration ─────────────────────────────────────────────────

# Default timeout values (seconds)
TIMEOUT_RESOLVE_BIN="${TIMEOUT_RESOLVE_BIN:-5}"
TIMEOUT_JSON_PARSE="${TIMEOUT_JSON_PARSE:-3}"
TIMEOUT_MISE_WHICH="${TIMEOUT_MISE_WHICH:-5}"
TIMEOUT_FIND_BINARY="${TIMEOUT_FIND_BINARY:-10}"
TIMEOUT_NETWORK="${TIMEOUT_NETWORK:-30}"

# ── 🔧 Timeout Implementation Detection ──────────────────────────────────────

# Purpose: Detects available timeout implementation
# Returns:
#   "timeout" - GNU timeout available
#   "gtimeout" - macOS gtimeout available
#   "bash" - Bash native fallback
# Examples:
#   IMPL=$(detect_timeout_impl)
detect_timeout_impl() {
  if command -v timeout >/dev/null 2>&1; then
    echo "timeout"
  elif command -v gtimeout >/dev/null 2>&1; then
    echo "gtimeout"
  else
    echo "bash"
  fi
}

# ── 🛡️ Process Group Management ──────────────────────────────────────────────

# Purpose: Starts a command in a new process group
# Params:
#   $@ - Command and arguments to execute
# Returns:
#   PID of the process group leader
# Examples:
#   start_process_group sleep 100
start_process_group() {
  if command -v setsid >/dev/null 2>&1; then
    # Use setsid to create new session and process group
    setsid "$@" &
  else
    # Fallback: use subshell
    ("$@") &
  fi
  echo $!
}

# Purpose: Cleans up a process and all its children
# Params:
#   $1 - Process PID
#   $2 - Grace period in seconds (default: 2)
# Examples:
#   cleanup_process_tree 12345 2
cleanup_process_tree() {
  local _PID="${1:-}"
  local _GRACE="${2:-2}"

  [ -z "${_PID:-}" ] && return 0

  # Check if process exists
  if ! kill -0 "${_PID:-}" 2>/dev/null; then
    return 0
  fi

  # 1. Send SIGTERM to the process itself
  # Note: Avoid using group kill (-PID) unless we are sure it's in a new PGID (e.g. via setsid)
  # as it can kill the caller on some systems if not careful.
  kill -TERM "${_PID:-}" 2>/dev/null || true

  # 2. Wait for graceful shutdown
  local _WAITED=0
  while [ "${_WAITED:-0}" -lt "${_GRACE:-2}" ]; do
    if ! kill -0 "${_PID:-}" 2>/dev/null; then
      return 0
    fi
    sleep 1
    _WAITED=$((_WAITED + 1))
  done

  # 3. Force kill if still running
  kill -KILL "${_PID:-}" 2>/dev/null || true

  # 4. Clean up any remaining child processes
  if command -v pkill >/dev/null 2>&1; then
    pkill -KILL -P "${_PID:-}" 2>/dev/null || true
  fi

  return 0
}

# ── ⏱️ Robust Timeout Execution ──────────────────────────────────────────────

# Purpose: Executes a command with timeout and proper process cleanup
# Params:
#   $1 - Timeout in seconds
#   $2+ - Command and arguments to execute
# Returns:
#   Command exit code, or 124 if timeout occurred
# Examples:
#   run_with_timeout_robust 5 curl https://example.com
#   run_with_timeout_robust 10 unirtm which node
run_with_timeout_robust() {
  local _TIMEOUT="${1:-}"
  shift

  [ -z "${_TIMEOUT:-}" ] && return 1
  [ $# -eq 0 ] && return 1

  local _IMPL
  _IMPL=$(detect_timeout_impl)

  case "${_IMPL:-}" in
  timeout)
    # GNU timeout with SIGTERM -> SIGKILL escalation
    timeout --kill-after=2 "${_TIMEOUT:-}" "$@"
    return $?
    ;;
  gtimeout)
    # macOS gtimeout with SIGTERM -> SIGKILL escalation
    gtimeout --kill-after=2 "${_TIMEOUT:-}" "$@"
    return $?
    ;;
  bash)
    # Bash native implementation with process group cleanup
    _run_with_timeout_native "${_TIMEOUT:-}" "$@"
    return $?
    ;;
  esac
}

# Purpose: Native Bash/POSIX timeout implementation (internal)
# Params:
#   $1 - Timeout in seconds
#   $2+ - Command and arguments
# Returns:
#  Command exit code, or 124 if timeout
_run_with_timeout_native() {
  local _TIMEOUT="${1:-}"
  shift

  # Start command in process group
  "$@" &
  local _CMD_PID=$!

  # Start timeout watcher in background
  (
    echo "[$(date)] Watcher started for PID ${_CMD_PID:-} with timeout ${_TIMEOUT:-}s" >>/tmp/timeout_debug.log
    sleep "${_TIMEOUT:-}" 2>/dev/null || sleep "${_TIMEOUT:-}"
    echo "[$(date)] Watcher woke up for PID ${_CMD_PID:-}" >>/tmp/timeout_debug.log
    if kill -0 "${_CMD_PID:-}" 2>/dev/null; then
      echo "[$(date)] Calling cleanup for PID ${_CMD_PID:-}" >>/tmp/timeout_debug.log
      # Timeout occurred - clean up process tree
      cleanup_process_tree "${_CMD_PID:-}" 2
      echo "[$(date)] Cleanup finished for PID ${_CMD_PID:-}" >>/tmp/timeout_debug.log
    fi
  ) >/dev/null 2>&1 </dev/null &
  local _WATCH_PID=$!

  # Wait for command to complete
  local _RET=0
  wait "${_CMD_PID:-}" 2>/dev/null || _RET=$?

  # Kill the watcher if command finished before timeout
  kill "${_WATCH_PID:-}" 2>/dev/null || true
  wait "${_WATCH_PID:-}" 2>/dev/null || true

  # Check if timeout occurred
  if [ "${_RET:-}" -ne 0 ] && ! kill -0 "${_CMD_PID:-}" 2>/dev/null; then
    # Process is gone - check if it was killed by timeout
    if [ "${_RET:-}" -eq 143 ] || [ "${_RET:-}" -eq 137 ]; then
      return 124
    fi
  fi

  return "${_RET:-}"
}

# ── 📊 Timeout Statistics (Debug Mode) ───────────────────────────────────────

# Purpose: Wraps run_with_timeout_robust with timing statistics
# Params:
#   $1 - Timeout in seconds
#   $2+ - Command and arguments
# Returns:
#   Command exit code
# Examples:
#   run_with_timeout_stats 5 curl https://example.com
run_with_timeout_stats() {
  local _TIMEOUT="${1:-}"
  shift

  if [ "${DEBUG_TIMEOUT:-0}" != "1" ]; then
    run_with_timeout_robust "${_TIMEOUT:-}" "$@"
    return $?
  fi

  local _START
  _START=$(date +%s 2>/dev/null || echo "0")

  run_with_timeout_robust "${_TIMEOUT:-}" "$@"
  local _RET=$?

  local _END
  _END=$(date +%s 2>/dev/null || echo "0")
  local _DURATION=$((_END - _START))

  if [ "${_RET:-}" -eq 124 ]; then
    printf "[TIMEOUT] Command timed out after %ds (limit: %ds)\n" "${_DURATION:-}" "${_TIMEOUT:-}" >&2
  else
    printf "[TIMEOUT] Command completed in %ds (limit: %ds, exit: %d)\n" "${_DURATION:-}" "${_TIMEOUT:-}" "${_RET:-}" >&2
  fi

  return "${_RET:-}"
}

# Export functions for use in other scripts
# Note: export -f is Bash-specific, but we provide the functions for sourcing
