#!/usr/bin/env sh
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# scripts/env.sh - Environment Configuration Orchestrator
#
# Purpose:
#   Standardizes management of .env files, validation, and template synchronization.
#   Ensures mandatory environment variables are populated before runtime.
#
# Usage:
#   sh scripts/env.sh [OPTIONS] [COMMAND]
#
# Standards:
#   - POSIX-compliant sh logic.
#   - "World Class" AI Documentation (English-only).
#   - Rule 01 (General), Rule 04 (Security).
#
# Features:
#   - POSIX compliant, encapsulated main() pattern.
#   - Safe template extraction and missing variable detection.
#   - Conditional overwrite protection for local secrets.

set -eu

# ── Common Library ───────────────────────────────────────────────────────────
SCRIPT_DIR=$(cd "$(dirname "${0:-}")" && pwd)
. "${SCRIPT_DIR:-}/lib/common.sh"

# ── Functions ────────────────────────────────────────────────────────────────

# Purpose: Displays usage information for the environment configuration manager.
# Examples:
#   show_help
show_help() {
  cat <<EOF
Usage: $0 [OPTIONS] [COMMAND]

Manages environment configuration files (.env).

Commands:
  setup            Create .env from .env.example if it doesn't exist.
  check            Validate .env against .env.example (check for missing keys).
  sync             Interactively sync missing keys from .env.example to .env.

Options:
  --dry-run        Preview changes without applying them.
  -q, --quiet      Suppress informational output.
  -v, --verbose    Enable verbose/debug output.
  -h, --help       Show this help message.

EOF
}

# Purpose: Copies .env.example to .env if it doesn't already exist.
# Examples:
#   run_env_setup
run_env_setup() {
  if [ ! -f ".env.example" ]; then
    log_warn "Warning: .env.example not found. Skipping setup."
    return
  fi

  if [ -f ".env" ]; then
    log_info ".env already exists. No action needed."
  else
    if [ "${DRY_RUN:-0}" -eq 1 ]; then
      log_info "DRY-RUN: Would copy .env.example to .env"
    else
      log_info "Creating .env from .env.example..."
      cp ".env.example" ".env"
      log_success "Successfully created .env"
    fi
  fi
}

# Purpose: Internal helper to parse .env.example and execute an action (check or sync) for missing keys.
_process_env_template() {
  local _MODE="${1:-}" # "check" or "sync"
  local _MISSING_COUNT=0
  local _ADDED_COUNT=0

  while IFS= read -r line_parsed || [ -n "${line_parsed:-}" ]; do
    case "${line_parsed:-}" in
    \#* | '') continue ;;
    esac

    local _KEY_PARSED
    _KEY_PARSED=$(echo "${line_parsed:-}" | cut -d'=' -f1 | sed 's/[[:space:]]*$//')

    if [ "${DRY_RUN:-0}" -eq 1 ]; then
      if [ "${_MODE:-}" = "check" ]; then
        log_debug "DRY-RUN: Would check if ${_KEY_PARSED:-} exists in .env"
      else
        log_info "DRY-RUN: Would add ${_KEY_PARSED:-} to .env"
      fi
      continue
    fi

    if ! grep -q "^${_KEY_PARSED:-}=" ".env" && ! grep -q "^${_KEY_PARSED:-} =" ".env"; then
      if [ "${_MODE:-}" = "check" ]; then
        log_warn "Missing key: ${_KEY_PARSED:-}"
        _MISSING_COUNT=$((_MISSING_COUNT + 1))
      elif [ "${_MODE:-}" = "sync" ]; then
        printf "Missing key found: %b%s%b. Add to .env? (y/N): " "${YELLOW:-}" "${_KEY_PARSED:-}" "${NC:-}"
        local _CONFIRM_SYNC
        read -r _CONFIRM_SYNC
        case "${_CONFIRM_SYNC:-}" in
        [yY]*)
          echo "${line_parsed:-}" >>.env
          log_success "Added ${_KEY_PARSED:-}"
          _ADDED_COUNT=$((_ADDED_COUNT + 1))
          ;;
        *) log_info "Skipped ${_KEY_PARSED:-}" ;;
        esac
      fi
    fi
  done <".env.example"

  if [ "${_MODE:-}" = "check" ]; then
    if [ "${_MISSING_COUNT:-}" -eq 0 ]; then
      log_success "All environment keys are present."
    else
      log_error "Validation failed: ${_MISSING_COUNT:-} keys missing from .env"
      log_info "💡 Run 'scripts/env.sh sync' to add missing keys (interactive)."
      exit 1
    fi
  elif [ "${_MODE:-}" = "sync" ]; then
    log_success "Sync complete. ${_ADDED_COUNT:-} keys added."
  fi
}

# Purpose: Validates that all keys in .env.example are present in .env.
# Examples:
#   run_env_check
run_env_check() {
  if [ ! -f ".env.example" ]; then
    log_error "Error: .env.example not found in the project root."
    log_info "💡 Ensure you are running this script from the workspace root."
    exit 1
  fi
  if [ ! -f ".env" ] && [ "${DRY_RUN:-0}" -ne 1 ]; then
    log_warn ".env not found. Please run 'scripts/env.sh setup' first."
    return
  fi

  log_info "Checking .env for missing keys from .env.example..."
  _process_env_template "check"
}

# Purpose: Interactively synchronizes missing keys from .env.example to .env.
# Examples:
#   run_env_sync
run_env_sync() {
  if [ ! -f ".env.example" ]; then
    log_error "Error: .env.example not found."
    exit 1
  fi
  if [ ! -f ".env" ]; then
    log_info "Initializing .env from .env.example..."
    cp ".env.example" ".env"
  fi

  log_info "Syncing missing keys from .env.example to .env..."
  _process_env_template "sync"
}

# Purpose: Main entry point for the environment configuration orchestrator.
# Params:
#   $@ - Command line arguments and optional command
# Examples:
#   main sync
main() {
  # 1. Execution Context Guard
  guard_project_root

  # 2. Argument Parsing
  local _COMMAND_ENV="setup"
  local _arg_env
  for _arg_env in "$@"; do
    case "${_arg_env:-}" in
    setup | check | sync) _COMMAND_ENV="${_arg_env:-}" ;;
    esac
  done
  parse_common_args "$@"

  log_info "🔧 Environment Configuration Manager (${_COMMAND_ENV:-})...\n"

  case "${_COMMAND_ENV:-}" in
  setup) run_env_setup ;;
  check) run_env_check ;;
  sync) run_env_sync ;;
  esac

  log_success "\n✨ Environment management task complete."

  # 4. Standardized Next Actions
  if [ "${DRY_RUN:-0}" -eq 0 ] && [ "${_IS_TOP_LEVEL:-}" = "true" ]; then
    printf "\n%bNext Actions:%b\n" "${YELLOW:-}" "${NC:-}"
    printf "  - Run %bunirtm install%b to prepare your development environment and synchronize project dependencies.\n" "${GREEN:-}" "${NC:-}"
  fi
}

main "$@"
