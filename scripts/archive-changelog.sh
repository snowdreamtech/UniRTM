#!/usr/bin/env sh
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# scripts/archive-changelog.sh - Automate major-version changelog archiving
#
# Purpose:
#   Moves entries of previous major versions from CHANGELOG.md to archival files.
#   Ensures the primary changelog remains concise while preserving historical data.
#
# Usage:
#   sh scripts/archive-changelog.sh [OPTIONS]
#
# Standards:
#   - POSIX-compliant sh logic.
#   - "World Class" AI Documentation (English-only).
#   - Rule 01 (General), Rule 03 (Architecture), Rule 07 (Git).
#
# Features:
#   - POSIX compliant, encapsulated main() pattern.
#   - Atomic Operations & Safety Traps (atomic_swap).
#   - Deduplication and History Sorting logic.

set -eu

# ── Common Library ───────────────────────────────────────────────────────────
SCRIPT_DIR=$(cd "$(dirname "${0:-}")" && pwd)
. "${SCRIPT_DIR:-}/lib/common.sh"

# ── Functions ────────────────────────────────────────────────────────────────

# Purpose: Displays usage information for the changelog archival manager.
# Examples:
#   show_help
show_help() {
  cat <<EOF
Usage: $0 [OPTIONS]

Automates the archival of old major versions from CHANGELOG.md.

Options:
  --dry-run        Preview archival actions without modifying files.
  -q, --quiet      Suppress informational output.
  -v, --verbose    Enable verbose/debug output.
  -h, --help       Show this help message.

EOF
}

# Purpose: Safely removes temporary files and releases concurrency locks.
# Params:
#   $1 - Temporary header file
#   $2 - New changelog file
#   $3 - Temporary summary file
#   $4 - Temporary archive directory
#   $5 - Dry run flag
# Examples:
#   run_archival_cleanup "${_TMP_HEADER:-}" "${_NEW_CHANGELOG:-}" "${_TMP_SUMMARY:-}" "${_TMP_ARCHIVE_PREFIX:-}" "${DRY_RUN:-}"
run_archival_cleanup() {
  local _TMP_HEADER_CLN="${1:-}"
  local _NEW_CHANGELOG_CLN="${2:-}"
  local _TMP_SUMMARY_CLN="${3:-}"
  local _TMP_ARCH_DIR_CLN="${4:-}"
  local _TMP_FILT_CLN="${5:-}"
  local _TMP_SWAP_CLN="${6:-}"
  local _TMP_HIST_CLN="${7:-}"
  local _DRY_RUN_CLN="${8:-}"

  log_debug "Cleaning up temporary files..."
  rm -f "${_TMP_HEADER_CLN:-}" "${_NEW_CHANGELOG_CLN:-}" "${_TMP_SUMMARY_CLN:-}" "${_TMP_FILT_CLN:-}" "${_TMP_SWAP_CLN:-}" "${_TMP_HIST_CLN:-}"
  rm -rf "${_TMP_ARCH_DIR_CLN:-}"
  if [ "${_DRY_RUN_CLN:-0}" -eq 0 ]; then
    rmdir "${LOCK_DIR:-}" 2>/dev/null || true
  fi
}

# Purpose: Main entry point for the changelog archival engine.
# Params:
#   $@ - Command line arguments
# Examples:
#   main --dry-run
main() {
  # 1. Execution Context Guard: Ensure we are in the project root
  guard_project_root
  if [ ! -f "${CHANGELOG:-}" ]; then
    log_error "Error: $CHANGELOG not found in project root."
    exit 1
  fi

  # 2. Argument Parsing
  parse_common_args "$@"

  log_info "💧 Starting Changelog Archival Process...\n"

  # Smart Directory Discovery
  local _REL_ARCHIVE_DIR="${ARCHIVE_DIR:-}"
  if [ -z "${_REL_ARCHIVE_DIR:-}" ]; then
    if [ -d "docs/changelogs" ]; then
      _REL_ARCHIVE_DIR="docs/changelogs"
    elif [ -d "changelogs" ]; then
      _REL_ARCHIVE_DIR="changelogs"
    else
      _REL_ARCHIVE_DIR="."
    fi
  fi

  # Concurrency Lock (Atomic mkdir is standard POSIX lock)
  if [ "${DRY_RUN:-0}" -eq 0 ]; then
    if ! mkdir "${LOCK_DIR:-}" 2>/dev/null; then
      log_error "Error: Another archival process seems to be running (lock exists: $LOCK_DIR)."
      exit 1
    fi
    log_debug "Lock acquired: $LOCK_DIR"
  fi

  # Preparation of local scratch variables
  local _TMP_HDR
  local _TMP_ARCH_PFX
  local _NEW_CHLOG
  local _TMP_SUM
  local _FILTERED_CONTENT_ARCH
  local _TMP_ARCH_SWAP
  local _HISTORY_LINKS_SCRATCH

  _TMP_HDR=$(mktemp)
  _TMP_ARCH_PFX=$(mktemp -d)
  _NEW_CHLOG=$(mktemp)
  _TMP_SUM=$(mktemp)
  _FILTERED_CONTENT_ARCH=$(mktemp)
  _TMP_ARCH_SWAP=$(mktemp)
  _HISTORY_LINKS_SCRATCH=$(mktemp)

  # Setup Trap for cleanup
  trap 'run_archival_cleanup "${_TMP_HDR:-}" "${_NEW_CHLOG:-}" "${_TMP_SUM:-}" "${_TMP_ARCH_PFX:-}" "${_FILTERED_CONTENT_ARCH:-}" "${_TMP_ARCH_SWAP:-}" "${_HISTORY_LINKS_SCRATCH:-}" "${DRY_RUN:-0}"' EXIT INT TERM

  # Get current major version (Universal Source)
  local _RAW_VER_ARCH
  _RAW_VER_ARCH=$(get_project_version)
  case "${_RAW_VER_ARCH:-}" in
  '' | '-' | 'null')
    log_warn "Could not determine version from project files. Falling back to Git tags..."
    local _GIT_TAG_ARCH
    _GIT_TAG_ARCH=$(git describe --tags --abbrev=0 2>/dev/null || true)
    _RAW_VER_ARCH="${_GIT_TAG_ARCH:-}"
    ;;
  esac

  # Extract major version
  local _CURRENT_MAJOR_ARCH
  _CURRENT_MAJOR_ARCH=$(echo "${_RAW_VER_ARCH:-}" | sed 's/^v//' | cut -d. -f1)

  # Input Validation
  case "${_CURRENT_MAJOR_ARCH:-}" in
  '' | *[!0-9]*)
    log_error "Error: Could not determine a valid numeric major version (Found: '${_CURRENT_MAJOR_ARCH:-}' from '${_RAW_VER_ARCH:-}')"
    exit 1
    ;;
  esac

  log_info "Targeting major version: ${_CURRENT_MAJOR_ARCH:-}"

  # 1. Extract the header (everything up to the first version header)
  local _FIRST_H2_LINE_ARCH
  _FIRST_H2_LINE_ARCH=$(grep -n "^## \[" "${CHANGELOG:-}" | head -n1 | cut -d: -f1)
  if [ -z "${_FIRST_H2_LINE_ARCH:-}" ]; then
    log_warn "No version headers found in $CHANGELOG. Nothing to archive."
    return 0
  fi

  log_debug "Header ends at line ${_FIRST_H2_LINE_ARCH:-}"
  head -n "$((_FIRST_H2_LINE_ARCH - 1))" "${CHANGELOG:-}" >"${_TMP_HDR:-}"

  # 2. Process all versions using a portable awk approach
  log_info "Scanning $CHANGELOG for older major versions..."
  tail -n +"${_FIRST_H2_LINE_ARCH:-}" "${CHANGELOG:-}" | awk -v major="${_CURRENT_MAJOR_ARCH:-}" -v prefix="${_TMP_ARCH_PFX:-}" '
    /^## History/ { skip_history = 1; next; }
    /^## \[/ {
      skip_history = 0;
      line = $0;
      sub(/^## \[/, "", line);
      sub(/^v/, "", line);
      split(line, parts, ".");
      v_major = parts[1];

      if ($0 ~ /Unreleased/) {
        output = "current";
      } else if (v_major == major) {
        output = "current";
      } else if (v_major != "") {
        output = "v" v_major;
      } else {
        output = "current";
      }
    }
    {
      if (skip_history) next;
      if (output == "") output = "current";
      target = prefix "/" output ".md";
      print >> target;
    }
  '

  # 3. Preparation for rebuild (Atomic)
  cat "${_TMP_HDR:-}" >"${_NEW_CHLOG:-}"
  if [ -f "${_TMP_ARCH_PFX:-}/current.md" ]; then
    log_debug "Preserving current major version content."
    cat "${_TMP_ARCH_PFX:-}/current.md" >>"${_NEW_CHLOG:-}"
  fi

  # 4. Handle archives with deduplication
  if [ "${_ARCH_SUMMARY_INITIALIZED:-false}" != "true" ] && ! check_ci_summary "### Archival Execution Summary"; then
    {
      printf "### Archival Execution Summary\n\n"
    } >>"${CI_STEP_SUMMARY:-}"
    [ -n "${GITHUB_ENV:-}" ] && echo "_ARCH_SUMMARY_INITIALIZED=true" >>"${GITHUB_ENV:-}"
    export _ARCH_SUMMARY_INITIALIZED=true
  fi

  # Provide table header if not already present
  if [ "${_SUMMARY_TABLE_HEADER_SENTINEL:-false}" != "true" ] && ! check_ci_summary "| Major Version | Action | Destination |"; then
    {
      printf "| Major Version | Action | Destination |\n"
      printf "| :--- | :--- | :--- |\n"
    } >>"${CI_STEP_SUMMARY:-}"
    [ -n "${GITHUB_ENV:-}" ] && echo "_SUMMARY_TABLE_HEADER_SENTINEL=true" >>"${GITHUB_ENV:-}"
    export _SUMMARY_TABLE_HEADER_SENTINEL=true
  fi
  local _ARCHIVE_COUNT_ARCH=0

  local _arch_file_iter
  for _arch_file_iter in "${_TMP_ARCH_PFX:-}"/v*.md; do
    [ -e "${_arch_file_iter:-}" ] || continue

    local _V_TAG_PROC
    _V_TAG_PROC=$(basename "${_arch_file_iter:-}" .md)
    local _V_NUM_PROC
    _V_NUM_PROC=$(echo "${_V_TAG_PROC:-}" | sed 's/^v//')
    local _FINAL_ARCH_FILE="${_REL_ARCHIVE_DIR:-}/CHANGELOG-v${_V_NUM_PROC:-}.md"
    log_debug "Processing archival block for major version ${_V_NUM_PROC:-}..."

    if [ -f "${_FINAL_ARCH_FILE:-}" ]; then
      # _FILTERED_CONTENT_ARCH is reused safely because we empty it
      : >"${_FILTERED_CONTENT_ARCH:-}"
      awk -v arch="${_FINAL_ARCH_FILE:-}" '
        BEGIN {
          while ((getline line < arch) > 0) {
            if (line ~ /^## \[/) { headers[line] = 1; }
          }
          close(arch);
        }
        /^## \[/ {
          if (headers[$0]) { skip = 1; }
          else { skip = 0; }
        }
        { if (!skip) print; }
      ' "${_arch_file_iter:-}" >"${_FILTERED_CONTENT_ARCH:-}"

      if [ -s "${_FILTERED_CONTENT_ARCH:-}" ]; then
        if [ "${DRY_RUN:-0}" -eq 1 ]; then
          log_warn "DRY-RUN: Would prepend new entries to ${_FINAL_ARCH_FILE:-}"
          printf "| v%s | Prepend (Dry Run) | %s |\n" "${_V_NUM_PROC:-}" "${_FINAL_ARCH_FILE:-}" >>"${CI_STEP_SUMMARY:-}"
        else
          log_info "Updating archive: ${_FINAL_ARCH_FILE:-}"
          # _TMP_ARCH_SWAP is reused safely because we empty it
          : >"${_TMP_ARCH_SWAP:-}"
          cat "${_FILTERED_CONTENT_ARCH:-}" >"${_TMP_ARCH_SWAP:-}"
          printf "\n" >>"${_TMP_ARCH_SWAP:-}"
          sed '1,2d' "${_FINAL_ARCH_FILE:-}" >>"${_TMP_ARCH_SWAP:-}"
          atomic_swap "${_TMP_ARCH_SWAP:-}" "${_FINAL_ARCH_FILE:-}"
          printf "| v%s | Updated | %s |\n" "${_V_NUM_PROC:-}" "${_FINAL_ARCH_FILE:-}" >>"${CI_STEP_SUMMARY:-}"
        fi
        _ARCHIVE_COUNT_ARCH=$((_ARCHIVE_COUNT_ARCH + 1))
      else
        log_debug "No new entries for v${_V_NUM_PROC:-}, skipping."
      fi
      rm -f "${_FILTERED_CONTENT_ARCH:-}"
    else
      if [ "${DRY_RUN:-0}" -eq 1 ]; then
        log_warn "DRY-RUN: Would create ${_FINAL_ARCH_FILE:-}"
        printf "| v%s | Create (Dry Run) | %s |\n" "${_V_NUM_PROC:-}" "${_FINAL_ARCH_FILE:-}" >>"${CI_STEP_SUMMARY:-}"
      else
        log_success "Creating new archive: ${_FINAL_ARCH_FILE:-}"
        mkdir -p "${_REL_ARCHIVE_DIR:-}"
        printf "# Changelog Archive v%s\n\n" "${_V_NUM_PROC:-}" >"${_FINAL_ARCH_FILE:-}"
        cat "${_arch_file_iter:-}" >>"${_FINAL_ARCH_FILE:-}"
        printf "| v%s | Created | %s |\n" "${_V_NUM_PROC:-}" "${_FINAL_ARCH_FILE:-}" >>"${CI_STEP_SUMMARY:-}"
      fi
      _ARCHIVE_COUNT_ARCH=$((_ARCHIVE_COUNT_ARCH + 1))
    fi
  done

  if [ "${_ARCHIVE_COUNT_ARCH:-}" -eq 0 ]; then
    printf "| N/A | No changes required | %s is up to date |\n" "${CHANGELOG:-}" >>"${_TMP_SUM:-}"
  fi

  # 5. Rebuild History section in NEW_CHLOG (Sorted)
  log_info "Rebuilding History section (Archive Directory: ${_REL_ARCHIVE_DIR:-})..."
  # _HISTORY_LINKS_SCRATCH is emptied for reuse
  : >"${_HISTORY_LINKS_SCRATCH:-}"
  # Check filesystem for existing archives in _REL_ARCHIVE_DIR
  local _f_hist
  for _f_hist in "${_REL_ARCHIVE_DIR:-}"/CHANGELOG-v*.md; do
    [ -e "${_f_hist:-}" ] || continue
    local _F_NAME_HIST
    _F_NAME_HIST=$(basename "${_f_hist:-}")
    local _V_NUM_HIST
    _V_NUM_HIST=$(echo "${_F_NAME_HIST:-}" | sed 's/^CHANGELOG-v//;s/\.md$//')
    printf "%s|%s\n" "${_V_NUM_HIST:-}" "- [v${_V_NUM_HIST:-}.x.x Archive](./$_f_hist)" >>"${_HISTORY_LINKS_SCRATCH:-}"
  done

  # Merge with newly planned archives from _TMP_ARCH_PFX
  for _f_hist in "${_TMP_ARCH_PFX:-}"/v*.md; do
    [ -e "${_f_hist:-}" ] || continue
    local _V_NUM_HIST
    _V_NUM_HIST=$(echo "${_f_hist:-}" | sed 's|^.*/v||;s/.md$//')
    local _LINK_HIST
    _LINK_HIST="- [v${_V_NUM_HIST:-}.x.x Archive](./${_REL_ARCHIVE_DIR:-}/CHANGELOG-v${_V_NUM_HIST:-}.md)"
    # Normalize link paths (remove repeated ./)
    _LINK_HIST=$(echo "${_LINK_HIST:-}" | sed 's|\./\./|\./|g')
    if ! grep -qF -- "${_LINK_HIST:-}" "${_HISTORY_LINKS_SCRATCH:-}" 2>/dev/null; then
      printf "%s|%s\n" "${_V_NUM_HIST:-}" "${_LINK_HIST:-}" >>"${_HISTORY_LINKS_SCRATCH:-}"
    fi
  done

  if [ -s "${_HISTORY_LINKS_SCRATCH:-}" ]; then
    printf "\n## History\n" >>"${_NEW_CHLOG:-}"
    sort -t'|' -k1,1nr "${_HISTORY_LINKS_SCRATCH:-}" | cut -d'|' -f2 >>"${_NEW_CHLOG:-}"
  fi
  rm -f "${_HISTORY_LINKS_SCRATCH:-}"

  # 6. Final Swap (Atomic)
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_warn "DRY-RUN: History section preview:"
    grep -A 100 "^## History" "${_NEW_CHLOG:-}" || true
  else
    atomic_swap "${_NEW_CHLOG:-}" "${CHANGELOG:-}"
    log_success "Successfully updated $CHANGELOG"
  fi

  # 7. Write Summary Log
  if [ -n "${CI_STEP_SUMMARY:-}" ]; then
    log_debug "Writing archival logs to $CI_STEP_SUMMARY..."
    cat "${_TMP_SUM:-}" >>"${CI_STEP_SUMMARY:-}" 2>/dev/null || true
  fi

  # 8. Standardized Next Actions
  if [ "${DRY_RUN:-0}" -eq 0 ] && [ "${_IS_TOP_LEVEL:-}" = "true" ]; then
    printf "\n%bNext Actions:%b\n" "${YELLOW:-}" "${NC:-}"
  fi
}

main "$@"
