#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# scripts/lib/bootstrap.sh - Infrastructure bootstrapping logic.
#
# Purpose:
#   Provides specialized logic for installing and activating core tooling
#   (primarily mise) across different operating systems and shells.
#
# Standards:
#   - POSIX-compliant sh logic.
#   - Rule 01 (General, Network), Rule 08 (Dev Env).

# Purpose: Ensures mise is installed and available in the environment.
#          Downloads the standalone binary if missing (cross-platform).
# Examples:
#   bootstrap_mise
# Internal helper to detect the current user shell.
_mise_detect_shell() {
  local _M_SHELL="bash"
  local _PARENT_SHELL
  _PARENT_SHELL=$(ps -p "${PPID:-}" -o comm= 2>/dev/null | awk -F/ '{print $NF}' | tr -d '-')

  case "${_PARENT_SHELL:-}" in
  zsh | bash | fish | pwsh | powershell | nu | xonsh | elvish)
    _M_SHELL="${_PARENT_SHELL:-}"
    ;;
  *)
    case "${SHELL:-}" in
    *zsh*) _M_SHELL="zsh" ;;
    *bash*) _M_SHELL="bash" ;;
    *fish*) _M_SHELL="fish" ;;
    *pwsh*) _M_SHELL="pwsh" ;;
    *nu*) _M_SHELL="nu" ;;
    *xonsh*) _M_SHELL="xonsh" ;;
    *elvish*) _M_SHELL="elvish" ;;
    *) _M_SHELL="bash" ;;
    esac
    ;;
  esac
  echo "${_M_SHELL:-}"
}

# Internal helper to detect the CPU architecture.
_mise_detect_arch() {
  local _OS="${1:-}"
  local _ARCH
  _ARCH=$(uname -m)
  case "${_ARCH:-}" in
  x86_64 | amd64) _ARCH="x64" ;;
  arm64 | aarch64) _ARCH="arm64" ;;
  armv7*) _ARCH="armv7" ;;
  *) _ARCH="x64" ;;
  esac

  if [ "${_OS:-}" = "linux" ]; then
    # Detect musl libc (common in Alpine and minimal Docker images)
    if (ldd "$(command -v ls)" 2>&1 | grep -q "musl") || [ -f /lib/ld-musl-x86_64.so.1 ] || [ -f /lib/ld-musl-aarch64.so.1 ] || [ -f /lib/ld-musl-armhf.so.1 ]; then
      _ARCH="${_ARCH:-}-musl"
    fi
  fi
  echo "${_ARCH:-}"
}

# Internal helper to detect the OS type.
_mise_detect_os() {
  case "$(uname -s)" in
  Darwin) echo "macos" ;;
  Linux) echo "linux" ;;
  MINGW* | MSYS* | CYGWIN*) echo "windows" ;;
  *) echo "linux" ;;
  esac
}

# Tier 1: Official install script (mise.jdx.dev) - Supports version specification.
_mise_install_tier1() {
  log_info "Tier 1: Trying official install script..."
  if [ -n "${MISE_VERSION:-}" ]; then
    log_info "Installing mise version: ${MISE_VERSION:-}"
  else
    log_info "Installing latest mise version"
  fi

  _TMP_SH="${TMPDIR:-/tmp}/mise_install_$.sh"
  if curl --retry 5 --retry-delay 2 --retry-connrefused -sS -L "https://mise.jdx.dev/install.sh" -o"${_TMP_SH:-}"; then
    if sh "${_TMP_SH:-}"; then
      rm -f "${_TMP_SH:-}"
      return 0
    fi
  fi
  rm -f "${_TMP_SH:-}"
  return 1
}

# Tier 2: System Package Managers.
_mise_install_tier2() {
  log_info "Tier 2: Searching for system package managers..."
  if command -v brew >/dev/null 2>&1; then
    log_info "Detected Homebrew. Installing mise..."
    brew install mise && return 0
  elif command -v port >/dev/null 2>&1; then
    log_info "Detected MacPorts. Installing mise..."
    sudo port install mise && return 0
  elif command -v apk >/dev/null 2>&1; then
    log_info "Detected apk. Installing mise..."
    sudo apk add mise && return 0
  elif command -v apt-get >/dev/null 2>&1; then
    log_info "Detected apt. Installing mise..."
    sudo apt-get update && sudo apt-get install -y mise && return 0
  elif command -v dnf >/dev/null 2>&1; then
    log_info "Detected dnf. Installing mise..."
    sudo dnf install -y mise && return 0
  elif command -v pacman >/dev/null 2>&1; then
    log_info "Detected pacman. Installing mise..."
    sudo pacman -S --noconfirm mise && return 0
  elif command -v nix-env >/dev/null 2>&1; then
    log_info "Detected nix. Installing mise..."
    nix-env -iA mise && return 0
  elif command -v yum >/dev/null 2>&1; then
    log_info "Detected yum. Installing mise..."
    sudo yum install -y mise && return 0
  elif command -v zypper >/dev/null 2>&1; then
    log_info "Detected zypper. Installing mise..."
    sudo zypper install -y mise && return 0
  fi
  return 1
}

# Tier 3: Language-specific tools.
_mise_install_tier3() {
  log_info "Tier 3: Searching for language-specific tools..."
  if command -v cargo >/dev/null 2>&1; then
    log_info "Detected Cargo. Installing mise..."
    cargo install mise && return 0
  else
    local _m_bs
    for _m_bs in pnpm npm bun; do
      if command -v "${_m_bs:-}" >/dev/null 2>&1; then
        log_info "Detected $_m_bs. Installing mise..."
        "${_m_bs:-}" install -g @jdxcode/mise && return 0
      fi
    done
    if command -v yarn >/dev/null 2>&1; then
      log_info "Detected yarn. Installing mise..."
      yarn global add @jdxcode/mise && return 0
    fi
  fi
  return 1
}

# Tier 4: Manual Binary Download (GitHub Releases).
_mise_install_tier4() {
  local _OS="${1:-}"
  local _ARCH="${2:-}"
  local _VER="${3:-}"
  log_info "Tier 4: Performing manual binary download for ${_OS:-}-${_ARCH:-} (v${_VER:-})..."

  local _M_BIN_NAME="mise-v${_VER:-}-${_OS:-}-${_ARCH:-}"
  local _EXT=""
  if [ "${_OS:-}" = "windows" ]; then _EXT=".zip"; fi
  local _M_URL="https://github.com/jdx/mise/releases/download/v${_VER:-}/${_M_BIN_NAME:-}${_EXT:-}"
  if [ "${ENABLE_GITHUB_PROXY:-}" = "1" ] || [ "${ENABLE_GITHUB_PROXY:-}" = "true" ]; then
    _M_URL="${GITHUB_PROXY:-}${_M_URL:-}"
  fi
  local _DEST="${_G_MISE_BIN_BASE:-$HOME/.local/bin}/mise"
  if [ "${_OS:-}" = "windows" ]; then _DEST="${_DEST:-}.exe"; fi

  mkdir -p "$(dirname "${_DEST:-}")"

  _download() {
    if command -v curl >/dev/null 2>&1; then
      run_quiet curl --retry 5 --retry-delay 2 --retry-connrefused -fSL --connect-timeout 15 -o "${2:-}" "${1:-}"
    elif command -v wget >/dev/null 2>&1; then
      run_quiet wget --tries=5 --waitretry=2 -q --timeout=15 -O "${2:-}" "${1:-}"
    else
      log_error "Neither curl nor wget is available for manual download."
      return 1
    fi
  }

  if [ "${_OS:-}" = "windows" ]; then
    if ! command -v unzip >/dev/null 2>&1; then
      log_error "Error: 'unzip' is required for Windows manual bootstrap."
      return 1
    fi
    local _TMP_DIR
    _TMP_DIR=$(mktemp -d 2>/dev/null || echo "/tmp/mise_win_extract_$")
    local _TMP_ZIP="${_TMP_DIR:-}/mise.zip"

    if _download "${_M_URL:-}" "${_TMP_ZIP:-}"; then
      if unzip -q "${_TMP_ZIP:-}" -d "${_TMP_DIR:-}"; then
        # Robustly find mise.exe and mise-shim.exe in any extracted path
        local _FOUND_BIN
        _FOUND_BIN=$(find "${_TMP_DIR:-}" -maxdepth 3 -name "mise.exe" | head -n 1)
        if [ -n "${_FOUND_BIN:-}" ]; then
          mv "${_FOUND_BIN:-}" "${_DEST:-}"
          local _FOUND_SHIM
          _FOUND_SHIM=$(find "${_TMP_DIR:-}" -maxdepth 3 -name "mise-shim.exe" | head -n 1)
          if [ -n "${_FOUND_SHIM:-}" ]; then mv "${_FOUND_SHIM:-}" "$(dirname "${_DEST:-}")/mise-shim.exe"; fi
          rm -rf "${_TMP_DIR:-}"
          return 0
        fi
      fi
    fi
    rm -rf "${_TMP_DIR:-}"
  else
    if _download "${_M_URL:-}" "${_DEST:-}"; then
      chmod +x "${_DEST:-}"
      return 0
    fi
  fi
  return 1
}

# Setup shell completions.
_mise_setup_completions() {
  local _SHELL="${1:-}"
  log_info "Setting up mise completions for ${_SHELL:-}..."

  # mise completion performs better when 'usage' is installed.
  # However, it often hangs on Windows CI due to compilation or interactive prompts.
  # We skip 'usage' installation entirely in CI to guarantee fast bootstrap.
  if ! is_ci_env && [ "${USAGE_FORCE_INSTALL:-0}" -ne 1 ]; then
    run_quiet run_mise install usage || true
  fi

  case "${_SHELL:-}" in
  zsh)
    local _DIR="${ZDOTDIR:-${HOME:-}}/.zsh/completions"
    mkdir -p "${_DIR:-}"
    mise completion zsh >"${_DIR:-}/_mise" 2>/dev/null || true
    ;;
  bash)
    local _DIR="$HOME/.local/share/bash-completion/completions"
    mkdir -p "${_DIR:-}"
    mise completion bash >"${_DIR:-}/mise" 2>/dev/null || true
    ;;
  fish)
    local _DIR="$HOME/.config/fish/completions"
    mkdir -p "${_DIR:-}"
    mise completion fish >"${_DIR:-}/mise.fish" 2>/dev/null || true
    ;;
  pwsh | powershell)
    local _DIR="$HOME/Documents/PowerShell/Completions"
    mkdir -p "${_DIR:-}"
    # mise completion supports 'powershell' (or 'pwsh' as alias in newer versions)
    mise completion powershell >"${_DIR:-}/mise-completion.ps1" 2>/dev/null || true
    ;;
  esac
}

# Run mise doctor to verify health.
_mise_verify_health() {
  log_info "Verifying mise health..."
  if ! run_quiet mise doctor; then
    log_warn "mise doctor reported some issues. Please check 'mise doctor' manually."
  else
    log_success "mise health check passed."
  fi
}

# ── 🐚 Shell-Specific Activation Helpers ─────────────────────────────────────

_mise_activate_bash() {
  local _RC="$HOME/.bashrc"
  [ -f "${_RC:-}" ] || return 0
  local _MISE_BIN
  _MISE_BIN=$(command -v mise 2>/dev/null || echo "${_G_MISE_BIN_BASE:-$HOME/.local/bin}/mise")
  # shellcheck disable=SC2016
  # Check for any existing mise activate line (with or without full path)
  if ! grep -qE '(mise|\.local/bin/mise) activate bash' "${_RC:-}"; then
    {
      echo ''
      echo '# mise activation (added by snowdreamtech/ai-ide-template setup)'
      echo "eval \"\$(${_MISE_BIN} activate bash)\""
    } >>"${_RC:-}"
    log_debug "Added mise activation to ${_RC:-}"
  else
    log_debug "mise activation already present in ${_RC:-}"
  fi
}

_mise_activate_zsh() {
  local _RC="${ZDOTDIR-$HOME}/.zshrc"
  [ -f "${_RC:-}" ] || return 0
  local _MISE_BIN
  _MISE_BIN=$(command -v mise 2>/dev/null || echo "${_G_MISE_BIN_BASE:-$HOME/.local/bin}/mise")
  # shellcheck disable=SC2016
  # Check for any existing mise activate line (with or without full path)
  if ! grep -qE '(mise|\.local/bin/mise) activate zsh' "${_RC:-}"; then
    {
      echo ''
      echo '# mise activation (added by snowdreamtech/ai-ide-template setup)'
      echo "eval \"\$(${_MISE_BIN} activate zsh)\""
    } >>"${_RC:-}"
    log_debug "Added mise activation to ${_RC:-}"
  else
    log_debug "mise activation already present in ${_RC:-}"
  fi
}

_mise_activate_fish() {
  local _RC="$HOME/.config/fish/config.fish"
  mkdir -p "$(dirname "${_RC:-}")"
  local _MISE_BIN
  _MISE_BIN=$(command -v mise 2>/dev/null || echo "${_G_MISE_BIN_BASE:-$HOME/.local/bin}/mise")
  # Check for any existing mise activate line
  if ! grep -qE '(mise|\.local/bin/mise) activate fish' "${_RC:-}"; then
    {
      echo ''
      echo '# mise activation (added by snowdreamtech/ai-ide-template setup)'
      echo "${_MISE_BIN} activate fish | source"
    } >>"${_RC:-}"
    log_debug "Added mise activation to ${_RC:-}"
  else
    log_debug "mise activation already present in ${_RC:-}"
  fi
}

_mise_activate_pwsh() {
  # Powershell profile path varies, we use a common heuristic.
  local _RC="$HOME/Documents/PowerShell/Microsoft.PowerShell_profile.ps1"
  [ -d "$(dirname "${_RC:-}")" ] || mkdir -p "$(dirname "${_RC:-}")"
  # Check for any existing mise activate line
  if ! grep -qE '(mise|\.local/bin/mise) activate pwsh' "${_RC:-}" 2>/dev/null; then
    {
      echo ''
      echo '# mise activation (added by snowdreamtech/ai-ide-template setup)'
      echo '(&mise activate pwsh) | Out-String | Invoke-Expression'
    } >>"${_RC:-}"
    log_debug "Added mise activation to ${_RC:-}"
  else
    log_debug "mise activation already present in ${_RC:-}"
  fi
}

_mise_activate_nu() {
  # Nushell requires env.nu and config.nu updates.
  local _NU_DIR="$HOME/.config/nushell"
  [ -d "${_NU_DIR:-}" ] || return 0
  local _ENV="${_NU_DIR:-}/env.nu"
  local _CONF="${_NU_DIR:-}/config.nu"
  local _MISE_NU="${_NU_DIR:-}/mise.nu"

  if [ ! -f "${_MISE_NU:-}" ]; then
    mise activate nu >"${_MISE_NU:-}" 2>/dev/null || true
  fi

  # shellcheck disable=SC2016
  grep -q "mise.nu" "${_ENV:-}" 2>/dev/null || printf "let mise_path = \$nu.default-config-dir | path join mise.nu\n^mise activate nu | save \$mise_path --force\n" >>"${_ENV:-}"
  # shellcheck disable=SC2016
  grep -q "mise.nu" "${_CONF:-}" 2>/dev/null || printf "use (\$nu.default-config-dir | path join mise.nu)\n" >>"${_CONF:-}"
}

_mise_activate_xonsh() {
  local _RC="$HOME/.config/xonsh/rc.xsh"
  [ -d "$(dirname "${_RC:-}")" ] || mkdir -p "$(dirname "${_RC:-}")"
  # shellcheck disable=SC2016
  # Check for any existing mise activate line
  if ! grep -qE '(mise|\.local/bin/mise) activate xonsh' "${_RC:-}" 2>/dev/null; then
    {
      echo ''
      echo '# mise activation (added by snowdreamtech/ai-ide-template setup)'
      echo 'execx($(mise activate xonsh))'
    } >>"${_RC:-}"
    log_debug "Added mise activation to ${_RC:-}"
  else
    log_debug "mise activation already present in ${_RC:-}"
  fi
}

_mise_activate_elvish() {
  local _RC="$HOME/.config/elvish/rc.elv"
  [ -d "$(dirname "${_RC:-}")" ] || mkdir -p "$(dirname "${_RC:-}")"
  # shellcheck disable=SC2016
  # Check for any existing mise activate line
  if ! grep -qE '(mise|\.local/bin/mise) activate elvish' "${_RC:-}" 2>/dev/null; then
    {
      echo ''
      echo '# mise activation (added by snowdreamtech/ai-ide-template setup)'
      echo 'eval (mise activate elvish | slurp)'
    } >>"${_RC:-}"
    log_debug "Added mise activation to ${_RC:-}"
  else
    log_debug "mise activation already present in ${_RC:-}"
  fi
}

# Helper to ensure mise is activated in the current session and RC files.
_mise_apply_activation() {
  local _SHELL="${1:-}"
  log_info "Synchronizing mise activation for ${_SHELL:-}..."

  # 1. Permanent RC File Injection
  case "${_SHELL:-}" in
  zsh) _mise_activate_zsh ;;
  bash) _mise_activate_bash ;;
  fish) _mise_activate_fish ;;
  pwsh | powershell) _mise_activate_pwsh ;;
  nu | nushell) _mise_activate_nu ;;
  xonsh) _mise_activate_xonsh ;;
  elvish) _mise_activate_elvish ;;
  *) _mise_activate_bash ;;
  esac

  # 2. Ephemeral Session Activation (POSIX sh compatible)
  # CRITICAL: We are running in a POSIX sh script, so we CANNOT eval shell-specific
  # activation code (e.g., zsh's $+functions, bash's PROMPT_COMMAND, etc.).
  # Instead, we only update PATH to include mise bin and shims directories.
  # Full activation (hook-env, etc.) will happen when the user starts a new shell
  # session with the RC file changes we made above.

  # Ensure mise bin directory is in PATH
  if [ -d "${_G_MISE_BIN_BASE:-}" ]; then
    case ":${PATH:-}:" in
    *":${_G_MISE_BIN_BASE:-}:"*) ;;
    *) export PATH="${_G_MISE_BIN_BASE:-}:${PATH:-}" ;;
    esac
  fi

  # Ensure mise shims directory is in PATH
  if [ -d "${_G_MISE_SHIMS_BASE:-}" ]; then
    case ":${PATH:-}:" in
    *":${_G_MISE_SHIMS_BASE:-}:"*) ;;
    *) export PATH="${_G_MISE_SHIMS_BASE:-}:${PATH:-}" ;;
    esac
  fi

  log_debug "mise PATH synchronized for current session. Full activation will occur in new shell sessions."
}

bootstrap_mise() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_info "Dry-run: Skipping mise bootstrap."
    return 0
  fi

  if command -v mise >/dev/null 2>&1; then
    log_debug "mise is already installed."
    _mise_apply_activation "$(_mise_detect_shell)"
    return 0
  fi

  log_info "mise not found. Initiating multi-tier prioritized bootstrap..."
  optimize_network

  local _M_SHELL
  _M_SHELL=$(_mise_detect_shell)
  local _M_OS
  _M_OS=$(_mise_detect_os)
  local _M_ARCH
  _M_ARCH=$(_mise_detect_arch "${_M_OS:-}")

  # Priority 1: Official Install Script (Supports version specification)
  if _mise_install_tier1; then
    log_success "mise installed via Tier 1 (Official Install Script)."
  # Priority 2: System Package Managers
  elif _mise_install_tier2; then
    log_success "mise installed via Tier 2 (Package Manager)."
  # Priority 3: Manual Binary (Fast & cross-platform)
  elif _mise_install_tier4 "${_M_OS:-}" "${_M_ARCH:-}" "${MISE_VERSION#[vV]}"; then
    log_success "mise installed via Tier 3 (Manual Binary)."
  # Priority 4: Language Tools (Slowest fallback)
  elif _mise_install_tier3; then
    log_success "mise installed via Tier 4 (Language Tool)."
  else
    log_error "All mise installation tiers failed."
    return 1
  fi

  # Path Refresh: Ensure MISE is available for immediate setup
  if [ -d "$HOME/.local/bin" ]; then export PATH="$HOME/.local/bin:$PATH"; fi
  if [ -d "${_G_MISE_BIN_BASE:-}" ]; then export PATH="${_G_MISE_BIN_BASE:-}:$PATH"; fi
  if [ -d "${_G_MISE_SHIMS_BASE:-}" ]; then export PATH="${_G_MISE_SHIMS_BASE:-}:$PATH"; fi

  # ── 🏗️ Post-Install Configuration ──

  # Finalize Activation
  _mise_apply_activation "${_M_SHELL:-}"

  # Setup Completions
  _mise_setup_completions "${_M_SHELL:-}"

  # Security & Automation: Trust project config
  if [ -f ".mise.toml" ]; then
    log_info "Trusting local .mise.toml..."
    mise trust ".mise.toml" >/dev/null 2>&1 || true
  fi

  # Verify Health
  _mise_verify_health
}
