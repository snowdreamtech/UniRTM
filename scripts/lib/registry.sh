#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Tool Registry - Centralized version management for dynamic registration

# Purpose: Registers a tool in .unirtm.toml if it's not already present.
#          Uses direct TOML injection (awk) instead of `unirtm use` to avoid
#          hitting the GitHub API during registration. This prevents 403
#          rate-limit errors when registering multiple GitHub-hosted tools.
# Params:
#   $1 - Tool name (internal)
#   $2 - UniRTM provider/name (e.g. asdf:ghc)
#   $3 - Version string
# Internal helper: checks for CI environment robustly even if common.sh is not pre-sourced.
_is_ci() {
  # Priority 1: Check dynamically set CI environment variables first
  # This allows tests and scripts to override CI detection
  if [ "${CI:-}" = "true" ] || [ "${GITHUB_ACTIONS:-}" = "true" ]; then
    return 0
  fi

  # Priority 2: Use is_ci_env if available (from common.sh)
  if command -v is_ci_env >/dev/null 2>&1; then
    is_ci_env
    return $?
  fi

  # Priority 3: Fallback to cached global flag
  [ "${_G_IS_CI:-0}" -eq 1 ]
}

register_unirtm_tool() {
  local _NAME="${1:-}"
  local _PROVIDER="${2:-}"
  local _VERSION="${3:-}"
  local _MISE_TOML
  _MISE_TOML=$(get_project_root)/.unirtm.toml
  if [ ! -f "${_MISE_TOML:-}" ]; then
    _MISE_TOML="$(get_project_root)/.mise.toml"
  fi

  # Check if already in config file
  if grep -qE "^\"?${_PROVIDER:-}\"?[[:space:]]*=" "${_MISE_TOML:-}" 2>/dev/null; then
    return 0
  fi

  log_info "Dynamically registering ${_NAME:-} SDK (${_PROVIDER:-}@${_VERSION:-})...."

  # In CI environment, we only install it but skip registry in config to keep it clean.
  # This prevents CI from dirtying the workspace and leaking CI-only tools back into the repo.
  if _is_ci; then
    # We must explicitly install here because since it is not in config,
    # the general 'unirtm install' at the end of setup won't capture it.
    if ! unirtm install "${_PROVIDER:-}@${_VERSION:-}"; then
      log_error "Failed to install ${_NAME:-} (${_PROVIDER:-}@${_VERSION:-}) in CI."
      return 1
    fi
    return 0
  fi

  # Inject directly into [tools] section via awk to avoid API calls.
  awk -v inject="\"${_PROVIDER:-}\" = \"${_VERSION:-}\"" '
    /^\[tools\]/ { print; print inject; next }
    { print }
  ' "${_MISE_TOML:-}" >"${_MISE_TOML:-}.tmp" && mv "${_MISE_TOML:-}.tmp" "${_MISE_TOML:-}"
}

# Backward compatibility wrapper
register_unirtm_tool() {
  register_unirtm_tool "$@"
}

# Purpose: Registers a tool in config using a complex TOML value (e.g., dictionary with asset matches).
# Params:
#   $1 - Tool name (internal)
#   $2 - Provider/name
#   $3 - TOML representation (dictionary map)
register_unirtm_tool_complex() {
  local _NAME="${1:-}"
  local _TOOL="${2:-}"
  local _TOML_VALUE="${3:-}"
  local _MISE_TOML
  _MISE_TOML=$(get_project_root)/.unirtm.toml
  if [ ! -f "${_MISE_TOML:-}" ]; then
    _MISE_TOML="$(get_project_root)/.mise.toml"
  fi

  # Check if already in config file
  if grep -qE "^\"?${_TOOL:-}\"?[[:space:]]*=" "${_MISE_TOML:-}" 2>/dev/null; then
    return 0
  fi

  log_info "Dynamically registering ${_NAME:-} SDK (${_TOOL:-}) with complex assets..."

  # In CI environment, we only install it but skip registry in config to keep it clean.
  # This prevents CI from dirtying the workspace and leaking CI-only tools back into the repo.
  if _is_ci; then
    unirtm install "${_TOOL:-}"
    return 0
  fi

  awk -v inject="\"${_TOOL:-}\" = ${_TOML_VALUE:-}" '
    /^\[tools\]/ { print; print inject; next }
    { print }
  ' "${_MISE_TOML:-}" >"${_MISE_TOML:-}.tmp" && mv "${_MISE_TOML:-}.tmp" "${_MISE_TOML:-}"

  unirtm install "${_TOOL:-}"
}

# Backward compatibility wrapper
register_unirtm_tool_complex() {
  register_unirtm_tool_complex "$@"
}

# --- Registry Data ---
# Note: Core runtimes (Node, Python) are always in .unirtm.toml.
# Secondary runtimes (Go, Rust, Java, etc.) are dynamically registered
# only when their source files are detected or explicitly requested.

setup_registry_go() { register_unirtm_tool "Go" "go" "${VER_GO:-}"; }
setup_registry_rust() { register_unirtm_tool "Rust" "rust" "${VER_RUST:-}"; }
setup_registry_java() { register_unirtm_tool "Java" "java" "${VER_JAVA:-}"; }
setup_registry_dotnet() { register_unirtm_tool ".NET" "dotnet" "${VER_DOTNET:-}"; }
setup_registry_zig() { register_unirtm_tool "Zig" "zig" "${VER_ZIG:-}"; }
setup_registry_bun() { register_unirtm_tool "Bun" "bun" "${VER_BUN:-}"; }
setup_registry_deno() { register_unirtm_tool "Deno" "deno" "${VER_DENO:-}"; }

setup_registry_ada() { register_unirtm_tool "Ada" "asdf:ada" "${VER_ADA:-14.2.0}"; }
setup_registry_clojure() { register_unirtm_tool "Clojure" "asdf:clojure" "${VER_CLOJURE:-1.12.0.1479}"; }
setup_registry_crystal() { register_unirtm_tool "Crystal" "asdf:crystal" "${VER_CRYSTAL:-1.15.1}"; }
setup_registry_dart() { register_unirtm_tool "Dart" "asdf:dart" "${VER_DART:-3.7.0}"; }
setup_registry_dlang() { register_unirtm_tool "Dlang" "asdf:dlang" "${VER_DLANG:-2.109.1}"; }
setup_registry_elixir() { register_unirtm_tool "Elixir" "elixir" "${VER_ELIXIR:-1.18.2-otp-27}"; }
setup_registry_elm() { register_unirtm_tool "Elm" "elm" "${VER_ELM:-0.19.1}"; }
setup_registry_erlang() { register_unirtm_tool "Erlang" "erlang" "${VER_ERLANG:-27.2.2}"; }
setup_registry_fpc() { register_unirtm_tool "FPC" "asdf:fpc" "${VER_FPC:-3.2.2}"; }
setup_registry_gcc() { register_unirtm_tool "GCC" "asdf:gcc" "${VER_GCC:-15.2}"; }
setup_registry_ghc() { register_unirtm_tool "GHC" "asdf:ghc" "${VER_GHC:-9.10.1}"; }
setup_registry_ormolu() { register_unirtm_tool "Ormolu" "ormolu" "${VER_ORMOLU:-0.7.7.0}"; }
setup_registry_gleam() { register_unirtm_tool "Gleam" "asdf:gleam" "${VER_GLEAM:-1.8.1}"; }
setup_registry_groovy() { register_unirtm_tool "Groovy" "asdf:groovy" "${VER_GROOVY:-4.0.25}"; }
setup_registry_haxe() { register_unirtm_tool "Haxe" "asdf:haxe" "${VER_HAXE:-4.3.6}"; }
setup_registry_julia() { register_unirtm_tool "Julia" "asdf:julia" "${VER_JULIA:-1.11.3}"; }
setup_registry_kotlin() { register_unirtm_tool "Kotlin" "asdf:kotlin" "${VER_KOTLIN:-2.1.10}"; }
setup_registry_lean() { register_unirtm_tool "Lean" "asdf:lean" "${VER_LEAN:-4.26.0}"; }
setup_registry_llvm() { register_unirtm_tool "LLVM" "asdf:llvm" "${VER_LLVM:-19.1.7}"; }
setup_registry_lua() { register_unirtm_tool "Lua" "asdf:lua" "${VER_LUA:-5.4.7}"; }
setup_registry_luau() { register_unirtm_tool "Luau" "asdf:luau" "${VER_LUAU:-0.712}"; }
setup_registry_mojo() { register_unirtm_tool "Mojo" "asdf:mojo" "${VER_MOJO:-0.26.1}"; }
setup_registry_nim() { register_unirtm_tool "Nim" "asdf:nim" "${VER_NIM:-2.2.0}"; }
setup_registry_ocaml() { register_unirtm_tool "OCaml" "asdf:ocaml" "${VER_OCAML:-5.3.0}"; }
setup_registry_odin() { register_unirtm_tool "Odin" "asdf:odin" "${VER_ODIN:-dev-2026-03}"; }
setup_registry_perl() { register_unirtm_tool "Perl" "asdf:perl" "${VER_PERL:-5.40.0}"; }
setup_registry_php() { register_unirtm_tool "PHP" "asdf:php" "${VER_PHP:-8.3.16}"; }
setup_registry_r() { register_unirtm_tool "R" "asdf:R" "${VER_R:-4.4.2}"; }
setup_registry_racket() { register_unirtm_tool "Racket" "asdf:racket" "${VER_RACKET:-9.1}"; }
setup_registry_raku() { register_unirtm_tool "Raku" "asdf:raku" "${VER_RAKU:-2026.02}"; }
setup_registry_rescript() { register_unirtm_tool "ReScript" "asdf:rescript" "${VER_RESCRIPT:-12.0.0}"; }
setup_registry_ruby() { register_unirtm_tool "Ruby" "ruby" "${VER_RUBY:-3.4.2}"; }
setup_registry_sbcl() { register_unirtm_tool "SBCL" "asdf:sbcl" "${VER_SBCL:-2.6.2}"; }
setup_registry_scala() { register_unirtm_tool "Scala" "asdf:scala" "${VER_SCALA:-3.6.3}"; }
setup_registry_scalafmt() { register_unirtm_tool "Scalafmt" "scalafmt" "${VER_SCALAFMT:-3.8.3}"; }
setup_registry_solc() { register_unirtm_tool "Solc" "asdf:solc" "${VER_SOLC:-0.8.28}"; }
setup_registry_swift() { register_unirtm_tool "Swift" "asdf:swift" "${VER_SWIFT:-6.0.3}"; }
setup_registry_prolog() { register_unirtm_tool "Prolog" "asdf:swi-prolog" "${VER_PROLOG:-10.1.5}"; }
setup_registry_tcl() { register_unirtm_tool "Tcl" "asdf:tcl" "${VER_TCL:-9.0.3}"; }
setup_registry_vlang() { register_unirtm_tool "Vlang" "asdf:vlang" "${VER_VLANG:-0.5.1}"; }
setup_registry_wasmtime() { register_unirtm_tool "Wasmtime" "asdf:wasmtime" "${VER_WASMTIME:-42.0.1}"; }
setup_registry_grain() { register_unirtm_tool "Grain" "${VER_GRAIN_PROVIDER:-github:grain-lang/grain}" "${VER_GRAIN:-0.7.2}"; }
setup_registry_moonbit() { register_unirtm_tool "Moonbit" "${VER_MOONBIT_PROVIDER:-github:moonbitlang/moonbit-compiler}" "${VER_MOONBIT:-0.8.0}"; }
setup_registry_kcl() { register_unirtm_tool "KCL" "${VER_KCL_PROVIDER:-asdf:kcl}" "${VER_KCL:-0.11.1}"; }
setup_registry_move() { register_unirtm_tool "Move" "asdf:move" "${VER_MOVE:-1.2.0}"; }
setup_registry_pkl() { register_unirtm_tool "Pkl" "${VER_PKL_PROVIDER:-asdf:pkl}" "${VER_PKL:-0.31.0}"; }
setup_registry_bazel() { register_unirtm_tool "Bazel" "${VER_BAZEL_PROVIDER:-asdf:bazel}" "${VER_BAZEL:-9.0.1}"; }
setup_registry_spark() { register_unirtm_tool "Spark" "asdf:spark" "${VER_SPARK:-4.1.1}"; }
setup_registry_ballerina() { register_unirtm_tool "Ballerina" "${VER_BALLERINA_PROVIDER:-github:ballerina-platform/ballerina-distribution}" "${VER_BALLERINA:-2201.11.0}"; }

# -- Secondary Tooling / Infrastructure / Linting --
setup_registry_kube_linter() { register_unirtm_tool "Kube-Linter" "${VER_KUBE_LINTER_PROVIDER:-github:stackrox/kube-linter}" "${VER_KUBE_LINTER:-latest}"; }
setup_registry_spectral() { register_unirtm_tool "Spectral" "${VER_SPECTRAL_PROVIDER:-npm:@stoplight/spectral-cli}" "${VER_SPECTRAL:-latest}"; }
setup_registry_buf() { register_unirtm_tool "Buf" "${VER_BUF_PROVIDER:-github:bufbuild/buf}" "${VER_BUF:-latest}"; }
# NOTE: setup_registry_trivy() removed — scanning delegated to aquasecurity/trivy-action.
setup_registry_osv_scanner() { register_unirtm_tool "OSV-Scanner" "${VER_OSV_SCANNER_PROVIDER:-github:google/osv-scanner}" "${VER_OSV_SCANNER:-latest}"; }
setup_registry_cargo_audit() { register_unirtm_tool "Cargo-Audit" "${VER_CARGO_AUDIT_PROVIDER:-cargo:cargo-audit}" "${VER_CARGO_AUDIT:-latest}"; }
setup_registry_zizmor() { register_unirtm_tool "Zizmor" "${VER_ZIZMOR_PROVIDER:-github:zizmorcore/zizmor}" "${VER_ZIZMOR:-latest}"; }
setup_registry_tflint() { register_unirtm_tool "TFLint" "${VER_TFLINT_PROVIDER:-github:terraform-linters/tflint}" "${VER_TFLINT:-latest}"; }
setup_registry_tofu() { register_unirtm_tool "OpenTofu" "${VER_TOFU_PROVIDER:-github:opentofu/opentofu}" "${VER_TOFU:-latest}"; }
setup_registry_just() { register_unirtm_tool "Just" "${VER_JUST_PROVIDER:-github:casey/just}" "${VER_JUST:-latest}"; }
setup_registry_task() { register_unirtm_tool "Task" "${VER_TASK_PROVIDER:-github:go-task/task}" "${VER_TASK:-latest}"; }
setup_registry_ktlint() { register_unirtm_tool "ktlint" "${VER_KTLINT_PROVIDER:-npm:@naturalcycles/ktlint}" "${VER_KTLINT:-latest}"; }
setup_registry_swiftformat() { register_unirtm_tool "SwiftFormat" "${VER_SWIFTFORMAT_PROVIDER:-github:nicklockwood/SwiftFormat}" "${VER_SWIFTFORMAT:-latest}"; }
setup_registry_swiftlint() { register_unirtm_tool "SwiftLint" "${VER_SWIFTLINT_PROVIDER:-github:realm/SwiftLint}" "${VER_SWIFTLINT:-latest}"; }
setup_registry_rubocop() { register_unirtm_tool "Rubocop" "${VER_RUBOCOP_PROVIDER:-gem:rubocop}" "${VER_RUBOCOP:-latest}"; }
setup_registry_google_java_format() { register_unirtm_tool_complex "Google Java Format" "${VER_JAVA_FORMAT_PROVIDER:-github:google/google-java-format}" "{ version = \"${VER_JAVA_FORMAT:-latest}\", asset = [ { match = \"darwin-arm64\" }, { match = \"linux-x86-64\" }, { match = \"linux-arm64\" }, { match = \"windows-x86-64\" }, { match = \"all-deps.jar\" } ] }"; }
setup_registry_stylua() { register_unirtm_tool "StyLua" "${VER_STYLUA_PROVIDER:-github:JohnnyMorganz/StyLua}" "${VER_STYLUA:-latest}"; }
