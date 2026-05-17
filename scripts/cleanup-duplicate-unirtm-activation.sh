#!/usr/bin/env sh
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License.

# scripts/cleanup-duplicate-unirtm-activation.sh
#
# Purpose: Remove duplicate unirtm activation lines from shell RC files
#
# Usage:
#   sh scripts/cleanup-duplicate-unirtm-activation.sh [--dry-run]
#
# This script fixes the issue where multiple 'make setup' runs created
# duplicate unirtm activation lines in shell configuration files.

set -eu

DRY_RUN=0
if [ "${1:-}" = "--dry-run" ]; then
  DRY_RUN=1
  echo "🔍 DRY RUN MODE - No files will be modified"
  echo ""
fi

cleanup_rc_file() {
  local _file="${1:-}"
  local _shell="${2:-}"

  [ ! -f "${_file:-}" ] && return 0

  echo "Checking ${_file:-}..."

  # Count unirtm activation lines
  local _count
  _count=$(grep -cE '(unirtm|\.local/bin/unirtm) activate' "${_file:-}" 2>/dev/null || true)
  _count=${_count:-0}

  if [ "${_count:-0}" -eq 0 ]; then
    echo "  ✅ No unirtm activation lines found"
    return 0
  elif [ "${_count:-0}" -eq 1 ]; then
    echo "  ✅ Single unirtm activation line (OK)"
    return 0
  fi

  echo "  ⚠️  Found ${_count:-} duplicate unirtm activation lines"

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    echo "  📋 Would remove duplicates and keep one line"
    grep -nE '(unirtm|\.local/bin/unirtm) activate' "${_file:-}" | head -5
    return 0
  fi

  # Create backup
  cp "${_file:-}" "${_file:-}.backup-$(date +%Y%m%d-%H%M%S)"
  echo "  💾 Created backup: ${_file:-}.backup-$(date +%Y%m%d-%H%M%S)"

  # Remove all unirtm activation lines and related comments
  grep -vE '(unirtm|\.local/bin/unirtm) activate|# unirtm activation' "${_file:-}" >"${_file:-}.tmp" || true

  # Add back a single unirtm activation line at the end
  # Try to detect unirtm binary location, fallback to standard location
  local _unirtm_bin
  if command -v unirtm >/dev/null 2>&1; then
    _unirtm_bin=$(command -v unirtm)
  else
    # Fallback: try common locations
    if [ -f "$HOME/.local/bin/unirtm" ]; then
      _unirtm_bin="$HOME/.local/bin/unirtm"
    elif [ -f "$HOME/Library/Application Support/unirtm/bin/unirtm" ]; then
      _unirtm_bin="$HOME/Library/Application Support/unirtm/bin/unirtm"
    else
      _unirtm_bin="unirtm" # Hope it's in PATH
    fi
  fi

  {
    cat "${_file:-}.tmp"
    echo ""
    echo "# unirtm activation (added by snowdreamtech/ai-ide-template setup)"
    case "${_shell:-}" in
    bash)
      echo "eval \"\$(${_unirtm_bin} activate bash)\""
      ;;
    zsh)
      echo "eval \"\$(${_unirtm_bin} activate zsh)\""
      ;;
    fish)
      echo "${_unirtm_bin} activate fish | source"
      ;;
    *)
      echo "eval \"\$(${_unirtm_bin} activate ${_shell:-})\""
      ;;
    esac
  } >"${_file:-}"

  rm "${_file:-}.tmp"
  echo "  ✅ Cleaned up duplicates, kept one activation line"
}

echo "🧹 Cleaning up duplicate unirtm activation lines..."
echo ""

# Bash
cleanup_rc_file "$HOME/.bashrc" "bash"

# Zsh
cleanup_rc_file "${ZDOTDIR-$HOME}/.zshrc" "zsh"

# Fish
cleanup_rc_file "$HOME/.config/fish/config.fish" "fish"

# PowerShell
cleanup_rc_file "$HOME/Documents/PowerShell/Microsoft.PowerShell_profile.ps1" "pwsh"

# Xonsh
cleanup_rc_file "$HOME/.config/xonsh/rc.xsh" "xonsh"

# Elvish
cleanup_rc_file "$HOME/.config/elvish/rc.elv" "elvish"

echo ""
if [ "${DRY_RUN:-0}" -eq 1 ]; then
  echo "✅ Dry run complete. Run without --dry-run to apply changes."
else
  echo "✅ Cleanup complete!"
  echo ""
  echo "📝 Backups were created with timestamp suffix (.backup-YYYYMMDD-HHMMSS)"
  echo "🔄 Please restart your shell or run: source ~/.zshrc (or ~/.bashrc)"
fi
