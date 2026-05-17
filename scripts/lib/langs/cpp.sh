#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# C/C++ Logic Module

# Purpose: Installs C/C++ toolchain via unirtm or system package manager.
# Delegate: Managed by unirtm (.unirtm.toml)
install_runtime_cpp() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log_debug "DRY_RUN: Would install C/C++ toolchain (gcc, clang, cmake, ninja)."
    return 0
  fi

  # Prefer native tools (brew/port/apt) for speed, fallback to unirtm for consistency
  ensure_tool cmake || return 1
  ensure_tool ninja || return 1

  # System level check for compiler
  if ! resolve_bin "gcc" >/dev/null 2>&1 && ! resolve_bin "clang" >/dev/null 2>&1; then
    log_warn "No C/C++ compiler (gcc/clang) detected. Please install build-essential or xcode-select."
    return 1
  fi
}

# Purpose: Installs clang-format.
# Delegate: Managed by unirtm (.unirtm.toml)
install_clang_format() {
  install_tool_safe "clang-format" "${VER_CLANG_FORMAT_PROVIDER:-}" "clang-format" "--version" 0 "*.c *.cpp *.h *.hpp *.cc *.m *.mm" ""
}

# Purpose: Sets up C/C++ environment for project.
setup_cpp() {
  if ! has_lang_files "CMakeLists.txt WORKSPACE .clang-format .clang-tidy" "*.cpp *.cc *.cxx *.h *.hpp *.hh"; then
    return 0
  fi

  local _T0_CPP_RT
  _T0_CPP_RT=$(date +%s)
  # Fast-path: Check version-aware existence
  local _CUR_VER
  _CUR_VER=$(get_version cpp)
  local _REQ_VER
  _REQ_VER=$(get_unirtm_tool_version "cpp")

  if [ "${_CUR_VER:-}" != "-" ] && { [ "${_REQ_VER:-}" = "latest" ] || [ "${_CUR_VER:-}" = "${_REQ_VER:-}" ]; }; then
    log_summary "Runtime" "C/C++" "✅ Detected" "${_CUR_VER:-}" "0"
  else
    _log_setup "C/C++ Toolchain" "cpp"

    if [ "${DRY_RUN:-0}" -eq 1 ]; then
      log_summary "Runtime" "C/C++" "⚖️ Previewed" "-" "0"
    else
      local _STAT_CPP_RT="✅ Installed"
      install_runtime_cpp || _STAT_CPP_RT="❌ Failed"

      local _DUR_CPP_RT
      _DUR_CPP_RT=$(($(date +%s) - _T0_CPP_RT))

      local _CPP_VER
      if resolve_bin "clang" >/dev/null 2>&1; then
        _CPP_VER=$(get_version clang | head -n 1)
      else
        _CPP_VER=$(get_version gcc | head -n 1)
      fi

      log_summary "Runtime" "C/C++" "${_STAT_CPP_RT:-}" "${_CPP_VER:-}" "${_DUR_CPP_RT:-}"
    fi
  fi

  # Setup related tools
  install_clang_format
}

# Purpose: Checks if C/C++ toolchain is available.
# Examples:
#   check_runtime_cpp "Linter"
check_runtime_cpp() {
  local _TOOL_DESC_CPP="${1:-C/C++}"
  if ! resolve_bin "gcc" >/dev/null 2>&1 && ! resolve_bin "clang" >/dev/null 2>&1; then
    log_warn "Required runtime 'gcc' or 'clang' for ${_TOOL_DESC_CPP:-} is missing. Skipping."
    return 1
  fi
  return 0
}
