#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# scripts/check-env.sh - Environment Health Auditor
#
# Purpose:
#   Validates the developer workstation against project-required runtimes and tools.
#   Identifies missing dependencies or version mismatches before development starts.
#
# Usage:
#   sh scripts/check-env.sh [OPTIONS]
#
# Standards:
#   - POSIX-compliant sh logic.
#   - "World Class" AI Documentation (English-only).
#   - Rule 01 (General), Rule 08 (Dev Env).
#
# Features:
#   - POSIX compliant, encapsulated main() pattern.
#   - Language-aware runtime detection (Node, Go, Python, etc.).
#   - High-performance, non-destructive validation scans.

# set -e removed to allow full diagnostic reporting

# ── Common Library ───────────────────────────────────────────────────────────
SCRIPT_DIR=$(cd "$(dirname "${0:-}")" && pwd)
. "${SCRIPT_DIR:-}/lib/common.sh"

# ── Global State (Scoped to script) ──────────────────────────────────────────
HEALTHY_ST=0
CORE_HEALTHY_ST=0

# ── Functions ────────────────────────────────────────────────────────────────

# Purpose: Displays usage information for the environment health auditor.
# Examples:
#   show_help
# shellcheck disable=SC2317,SC2329
show_help() {
  cat <<EOF
Usage: $0 [OPTIONS]

Checks the health of the development environment.

Options:
  -q, --quiet      Suppress informational output.
  -v, --verbose    Enable verbose/debug output.
  -h, --help       Show this help message.

EOF
}

# Purpose: Internal helper for version checking of a specific tool.
# Params:
#   $1 - Human-readable name
#   $2 - Command to verify
#   $3 - Minimum required version
#   $4 - Version check command
#   $5 - Critical flag (1 for core, 0 for optional)
#   $6 - CI-only flag (1 to skip locally with info, 0 to always check)
#
# Behavior Matrix:
#   (1, 1) Critical CI-only:  Local=skip, CI=must exist (fail if missing)
#          Use for: Universal security scanners (OSV-scanner, Zizmor)
#   (0, 1) Optional CI-only:  Local=skip, CI=optional (warn if missing)
#          Use for: Language-specific security tools (govulncheck, cargo-audit, etc.)
#   (1, 0) Critical Always:   Local=must exist, CI=must exist
#          Use for: Core infrastructure (Git, Make)
#   (0, 0) Optional Always:   Local=optional, CI=optional
#          Use for: Language runtimes, linters (when source files present)
#
# Examples:
#   check_tool_version "Git" "git" "2.30.0" "git --version" 1 0
#   check_tool_version "OSV-scanner" "osv-scanner" "1.0.0" "osv-scanner --version" 1 1 "osv-scanner" "OSV_FORCE_INSTALL"
#   check_tool_version "Govulncheck" "govulncheck" "latest" "govulncheck ./..." 0 1 "govulncheck" "GOVULN_FORCE_INSTALL"
check_tool_version() {
  local _LV_NAME="${1:-}"
  local _LV_CMD="${2:-}"
  local _LV_MIN_VER="${3:-}"
  local _LV_VER_CMD="${4:-}"
  local _LV_CRITICAL="${5:-0}"
  local _LV_CI_ONLY="${6:-0}"

  log_debug "Checking ${_LV_NAME:-} (min: ${_LV_MIN_VER:-})..."

  # 1. Availability-first detection:
  # Check if the tool is ALREADY resolved or available in the environment.
  local _LV_RESOLVED
  _LV_RESOLVED=$(resolve_bin "${_LV_CMD:-}") || true

  # 2. Path A: Tool is FOUND (Active)
  if [ -n "${_LV_RESOLVED:-}" ]; then
    local _LV_MISE_KEY="${7:-${_LV_CMD:-}}"
    local _LV_CURRENT_VER
    _LV_CURRENT_VER=$(get_version "${_LV_CMD:-}" "" "${_LV_MISE_KEY:-}" | tr -d '\r')
    [ "${_LV_CURRENT_VER:-}" = "-" ] && _LV_CURRENT_VER="0.0"

    # Strip leading 'v' or 'V' from version for consistent comparison
    _LV_CURRENT_VER=$(echo "${_LV_CURRENT_VER:-}" | sed 's/^[vV]//')

    # If requirement is empty or -, allow anything
    if [ -z "${_LV_MIN_VER:-}" ] || [ "${_LV_MIN_VER:-}" = "-" ]; then
      log_success "✅ ${_LV_NAME:-}: v${_LV_CURRENT_VER:-} (Active)"
      return 0
    fi

    # Strip leading 'v' from requirement as well
    local _LV_MIN_VER_CLEAN
    _LV_MIN_VER_CLEAN=$(echo "${_LV_MIN_VER:-}" | sed 's/^[vV]//')

    # Canonicalize versions to 3 components to avoid revision suffix mismatches (e.g. 1.7.11.24)
    local _LV_MIN_CANON _LV_CUR_CANON
    _LV_MIN_CANON=$(echo "${_LV_MIN_VER_CLEAN:-}" | cut -d. -f1-3)
    _LV_CUR_CANON=$(echo "${_LV_CURRENT_VER:-}" | cut -d. -f1-3)

    local _LV_LOWER_VER
    _LV_LOWER_VER=$(printf "%s\n%s" "${_LV_MIN_CANON:-}" "${_LV_CUR_CANON:-}" | sort -V | head -n1)

    if [ "${_LV_LOWER_VER:-}" = "${_LV_MIN_CANON:-}" ] || [ "${_LV_CUR_CANON:-}" = "${_LV_MIN_CANON:-}" ]; then
      log_success "✅ ${_LV_NAME:-}: v${_LV_CURRENT_VER:-} (Active, matches v${_LV_MIN_VER_CLEAN:-})"
    else
      # Version mismatch - warn but don't fail (tool exists, just old version)
      log_warn "⚠️  ${_LV_NAME:-}: v${_LV_CURRENT_VER:-} (below recommended v${_LV_MIN_VER_CLEAN:-})"
      HEALTHY_ST=1
      # Don't set CORE_HEALTHY_ST for version mismatches - tool is present and functional
    fi
    return 0
  fi

  # 3. Path B: Tool is MISSING
  local _LV_FORCE_VAR="${8:-}"
  # If tool is marked as CI-only and we are in local dev, it's optional.
  if [ "${_LV_CI_ONLY:-0}" -eq 1 ] && ! is_ci_env; then
    # Only fail if FORCE_INSTALL was explicitly requested but tool is missing.
    if [ -n "${_LV_FORCE_VAR:-}" ] && [ "$(eval echo "\${$_LV_FORCE_VAR:-0}")" -eq 1 ]; then
      log_warn "❌ ${_LV_NAME:-}: Not found (Forced check failed)."
      HEALTHY_ST=1
      return 1
    fi
    log_info "⏭️  ${_LV_NAME:-}: Optional (CI-only by default)"
    return 0
  fi

  # Tool is missing - determine severity based on criticality
  if [ "${_LV_CRITICAL:-0}" -eq 1 ]; then
    # Critical tool missing - this is a failure
    log_warn "❌ ${_LV_NAME:-}: Not found."
    HEALTHY_ST=1
    CORE_HEALTHY_ST=1
    return 1
  else
    # Non-critical tool missing - just a warning, don't fail
    log_warn "⚠️  ${_LV_NAME:-}: Not found (optional)."
    return 0
  fi
}

# Purpose: Main entry point for the environment health auditing engine.
# Params:
#   $@ - Command line arguments
# Examples:
#   main --verbose
main() {
  export _G_AUDIT_MODE=1

  # 1. Execution Context Guard
  guard_project_root

  # 2. Argument Parsing
  parse_common_args "$@"

  log_info "🔍 Checking Development Environment Health...\n"

  # 3. Group: Core Infrastructure
  log_info "── Core Infrastructure ──"
  check_tool_version "Git" "git" "2.30.0" "git --version" 1

  if resolve_bin "make" >/dev/null 2>&1; then
    log_success "✅ Make: Installed"
  else
    log_error "❌ Make: Not found."
    HEALTHY_ST=1
    CORE_HEALTHY_ST=1
  fi

  if resolve_bin "docker" >/dev/null 2>&1; then
    log_success "✅ Docker: Installed"
  else
    log_warn "⚠️  Docker: Not found (optional for some tasks)"
  fi
  printf "\n"

  # 4. Group: Language Runtimes (Dynamic Modular Verification)
  log_info "── Language Runtimes ──"

  # Node.js
  if [ -f "${PACKAGE_JSON:-}" ]; then
    check_tool_version "Node.js" "node" "$(get_unirtm_tool_version node)" "node -v" 0
    check_tool_version "${NPM:-}" "${NPM:-}" "$(get_unirtm_tool_version "${NPM:-}")" "$NPM -v"
    check_runtime "node" "Node.js (Modular)"
  else
    log_info "⏭️  Node.js/$NPM: Skipped (no package.json)"
  fi

  # Python
  if has_lang_files "requirements.txt requirements-dev.txt pyproject.toml" "*.py"; then
    check_tool_version "Python" "${PYTHON:-}" "$(get_unirtm_tool_version python)" "$PYTHON --version" 0
    check_runtime "python" "Python (Modular)"
  else
    log_info "⏭️  Python: Skipped (no python files)"
  fi

  # Go
  if has_lang_files "go.mod" "*.go"; then
    check_tool_version "Go" "go" "1.21.0" "go version" 0
    check_runtime "go" "Go (Modular)"
  else
    log_info "⏭️  Go: Skipped (no go files)"
  fi

  # Deno
  if has_lang_files "deno.json deno.jsonc" "*.ts *.tsx *.js *.jsx"; then
    check_runtime "deno" "Deno (Modular)"
  else
    log_info "⏭️  Deno: Skipped (no deno files)"
  fi

  # Bun
  if has_lang_files "bun.lockb package.json" "*.ts *.tsx *.js *.jsx"; then
    check_runtime "bun" "Bun (Modular)"
  else
    log_info "⏭️  Bun: Skipped (no bun files)"
  fi

  # Ruby
  if has_lang_files "Gemfile .ruby-version package.json" "*.rb"; then
    check_tool_version "Ruby" "ruby" "3.0.0" "ruby -v" 0
    check_runtime "ruby" "Ruby (Modular)"
  else
    log_info "⏭️  Ruby: Skipped (no ruby files)"
  fi

  # Java
  if has_lang_files "pom.xml build.gradle" "*.java"; then
    check_tool_version "Java" "java" "17" "java -version" 0
    check_runtime "java" "Java (Modular)"
  else
    log_info "⏭️  Java: Skipped (no java files)"
  fi

  # PHP
  if has_lang_files "composer.json" "*.php"; then
    check_tool_version "PHP" "php" "8.0.0" "php -v" 0
    check_runtime "php" "PHP (Modular)"
  else
    log_info "⏭️  PHP: Skipped (no php files)"
  fi

  # .NET
  if has_lang_files "global.json" "*.csproj *.sln *.cs"; then
    check_tool_version ".NET" "dotnet" "6.0.0" "dotnet --version" 0
    check_runtime "dotnet" "Dotnet (Modular)"
  else
    log_info "⏭️  .NET: Skipped (no dotnet files)"
  fi

  # Rust
  if has_lang_files "Cargo.toml" "*.rs"; then
    check_tool_version "Rust" "cargo" "1.70.0" "cargo --version" 0
    check_runtime "rust" "Rust (Modular)"
  else
    log_info "⏭️  Rust: Skipped (no rust files)"
  fi

  # C/C++
  if has_lang_files "*.c *.cpp *.h *.hpp"; then
    check_runtime "cpp" "C/C++"
  else
    log_info "⏭️  C/C++: Skipped (no C/C++ files)"
  fi

  # Terraform
  if has_lang_files "" "*.tf *.tfvars *.hcl"; then
    check_runtime "terraform" "Terraform (Modular)"
  else
    log_info "⏭️  Terraform: Skipped (no Terraform files)"
  fi

  # Solidity
  if has_lang_files "" "*.sol"; then
    check_runtime "solidity" "Solidity (Modular)"
  else
    log_info "⏭️  Solidity: Skipped (no Solidity files)"
  fi

  # Odin
  if has_lang_files "" "*.odin"; then
    check_runtime "odin" "Odin (Modular)"
  else
    log_info "⏭️  Odin: Skipped (no Odin files)"
  fi

  # Nim
  if has_lang_files "nim.cfg nimble.ini" "*.nim *.nims *.nimble"; then
    check_runtime "nim" "Nim (Modular)"
  else
    log_info "⏭️  Nim: Skipped (no Nim files)"
  fi

  # Clojure
  if has_lang_files "project.clj deps.edn bb.edn" "*.clj *.cljs *.cljc *.edn"; then
    check_runtime "clojure" "Clojure (Modular)"
  else
    log_info "⏭️  Clojure: Skipped (no Clojure files)"
  fi

  # Gleam
  if has_lang_files "gleam.toml" "*.gleam"; then
    check_runtime "gleam" "Gleam (Modular)"
  else
    log_info "⏭️  Gleam: Skipped (no Gleam files)"
  fi

  # Mojo
  if has_lang_files "" "*.mojo *.fire"; then
    check_runtime "mojo" "Mojo (Modular)"
  else
    log_info "⏭️  Mojo: Skipped (no Mojo files)"
  fi

  # OCaml
  if has_lang_files "dune-project dune opam" "*.ml *.mli *.mll *.mly"; then
    check_runtime "ocaml" "OCaml (Modular)"
  else
    log_info "⏭️  OCaml: Skipped (no OCaml files)"
  fi

  # Erlang
  if has_lang_files "rebar.config erlang.mk" "*.erl *.hrl"; then
    check_runtime "erlang" "Erlang (Modular)"
  else
    log_info "⏭️  Erlang: Skipped (no Erlang files)"
  fi

  # Vlang
  if has_lang_files "v.mod" "*.v *.vsh"; then
    check_runtime "vlang" "Vlang (Modular)"
  else
    log_info "⏭️  Vlang: Skipped (no Vlang files)"
  fi

  # Crystal
  if has_lang_files "shard.yml" "*.cr"; then
    check_runtime "crystal" "Crystal (Modular)"
  else
    log_info "⏭️  Crystal: Skipped (no Crystal files)"
  fi

  # Dlang
  if has_lang_files "dub.json dub.sdl" "*.d"; then
    check_runtime "dlang" "Dlang (Modular)"
  else
    log_info "⏭️  Dlang: Skipped (no Dlang files)"
  fi

  # Haxe
  if has_lang_files "project.xml build.hxml" "*.hx"; then
    check_runtime "haxe" "Haxe (Modular)"
  else
    log_info "⏭️  Haxe: Skipped (no Haxe files)"
  fi

  # AssemblyScript
  if has_lang_files "asconfig.json" "*.as"; then
    check_runtime "assemblyscript" "AssemblyScript (Modular)"
  else
    log_info "⏭️  AssemblyScript: Skipped (no AssemblyScript files)"
  fi

  # Ballerina
  if has_lang_files "Ballerina.toml" "*.bal"; then
    check_runtime "ballerina" "Ballerina (Modular)"
  else
    log_info "⏭️  Ballerina: Skipped (no Ballerina files)"
  fi

  # KCL
  if has_lang_files "kcl.mod" "*.k"; then
    check_runtime "kcl" "KCL (Modular)"
  else
    log_info "⏭️  KCL: Skipped (no KCL files)"
  fi

  # Pkl
  if has_lang_files "PklProject" "*.pkl"; then
    check_runtime "pkl" "Pkl (Modular)"
  else
    log_info "⏭️  Pkl: Skipped (no Pkl files)"
  fi

  # Move
  if has_lang_files "Move.toml" "*.move"; then
    check_runtime "move" "Move (Modular)"
  else
    log_info "⏭️  Move: Skipped (no Move files)"
  fi

  # Elm
  if has_lang_files "elm.json" "*.elm"; then
    check_runtime "elm" "Elm (Modular)"
  else
    log_info "⏭️  Elm: Skipped (no Elm files)"
  fi

  # ReScript
  if has_lang_files "rescript.json bsconfig.json" "*.res *.resi"; then
    check_runtime "rescript" "ReScript (Modular)"
  else
    log_info "⏭️  ReScript: Skipped (no ReScript files)"
  fi

  # Ada
  if has_lang_files "*.adb *.ads *.gpr"; then
    check_runtime "ada" "Ada (Modular)"
  else
    log_info "⏭️  Ada: Skipped (no Ada files)"
  fi

  # Luau
  if has_lang_files "*.luau"; then
    check_runtime "luau" "Luau (Modular)"
  else
    log_info "⏭️  Luau: Skipped (no Luau files)"
  fi

  # Raku
  if has_lang_files "META6.json" "*.raku *.rakumod *.p6 *.pm6"; then
    check_runtime "raku" "Raku (Modular)"
  else
    log_info "⏭️  Raku: Skipped (no Raku files)"
  fi

  # Vala
  if has_lang_files "*.vala *.vapi"; then
    check_runtime "vala" "Vala (Modular)"
  else
    log_info "⏭️  Vala: Skipped (no Vala files)"
  fi

  # Free Pascal
  if has_lang_files "*.pas *.pp *.inc *.lpr"; then
    check_runtime "fpc" "Free Pascal (Modular)"
  else
    log_info "⏭️  Free Pascal: Skipped (no Pascal files)"
  fi

  # Lean 4
  if has_lang_files "lean-toolchain lakefile.lean" "*.lean"; then
    check_runtime "lean" "Lean 4 (Modular)"
  else
    log_info "⏭️  Lean 4: Skipped (no Lean files)"
  fi

  # Common Lisp
  if has_lang_files "*.lisp *.cl *.asd"; then
    check_runtime "lisp" "Common Lisp (Modular)"
  else
    log_info "⏭️  Common Lisp: Skipped (no Lisp files)"
  fi

  # Racket
  if has_lang_files "*.rkt *.rktl"; then
    check_runtime "racket" "Racket (Modular)"
  else
    log_info "⏭️  Racket: Skipped (no Racket files)"
  fi

  # Prolog
  if has_lang_files "*.pl *.pro *.prolog"; then
    check_runtime "prolog" "Prolog (Modular)"
  else
    log_info "⏭️  Prolog: Skipped (no Prolog files)"
  fi

  # Fortran
  if has_lang_files "*.f *.for *.f90 *.f95"; then
    check_runtime "fortran" "Fortran (Modular)"
  else
    log_info "⏭️  Fortran: Skipped (no Fortran files)"
  fi

  # WebAssembly Text
  if has_lang_files "*.wat *.wasm"; then
    check_runtime "wat" "WebAssembly (Modular)"
  else
    log_info "⏭️  WebAssembly: Skipped (no Wasm files)"
  fi

  # MoonBit
  if has_lang_files "moon.pkg.json" "*.mbt"; then
    check_runtime "moonbit" "MoonBit (Modular)"
  else
    log_info "⏭️  MoonBit: Skipped (no MoonBit files)"
  fi

  # Grain
  if has_lang_files "*.gr"; then
    check_runtime "grain" "Grain (Modular)"
  else
    log_info "⏭️  Grain: Skipped (no Grain files)"
  fi

  # Jsonnet
  if has_lang_files "*.jsonnet *.libsonnet"; then
    check_runtime "jsonnet" "Jsonnet (Modular)"
  else
    log_info "⏭️  Jsonnet: Skipped (no Jsonnet files)"
  fi

  # Starlark
  if has_lang_files "*.star *.bzl BUILD WORKSPACE MODULE.bazel"; then
    check_runtime "starlark" "Starlark (Modular)"
  else
    log_info "⏭️  Starlark: Skipped (no Starlark files)"
  fi

  # Tcl
  if has_lang_files "*.tcl *.tk"; then
    check_runtime "tcl" "Tcl (Modular)"
  else
    log_info "⏭️  Tcl: Skipped (no Tcl files)"
  fi

  # DuckDB
  if has_lang_files "*.sql *.duckdb"; then
    check_runtime "duckdb" "DuckDB (Modular)"
  else
    log_info "⏭️  DuckDB: Skipped (no SQL/DuckDB files)"
  fi

  # VCPKG
  if has_lang_files "vcpkg.json vcpkg-configuration.json"; then
    check_runtime "vcpkg" "VCPKG (Modular)"
  else
    log_info "⏭️  VCPKG: Skipped (no VCPKG files)"
  fi

  # Terragrunt
  if has_lang_files "terragrunt.hcl"; then
    check_runtime "terragrunt" "Terragrunt (Modular)"
  else
    log_info "⏭️  Terragrunt: Skipped (no Terragrunt files)"
  fi

  # Apache Spark
  if has_lang_files "spark-defaults.conf *.pyspark"; then
    check_runtime "spark" "Apache Spark (Modular)"
  else
    log_info "⏭️  Apache Spark: Skipped (no Spark files)"
  fi

  # Helm
  if has_lang_files "Chart.yaml"; then
    check_runtime "helm" "Helm (Modular)"
  else
    log_info "⏭️  Helm: Skipped (no Chart.yaml)"
  fi

  # Typst
  if has_lang_files "*.typ"; then
    check_runtime "typst" "Typst (Modular)"
  else
    log_info "⏭️  Typst: Skipped (no Typst files)"
  fi

  # Perl
  if has_lang_files "cpanfile Makefile.PL" "*.pl *.pm"; then
    check_runtime "perl" "Perl (Modular)"
  else
    log_info "⏭️  Perl: Skipped (no perl files)"
  fi

  # Julia
  if has_lang_files "Project.toml Manifest.toml" "*.jl"; then
    check_runtime "julia" "Julia (Modular)"
  else
    log_info "⏭️  Julia: Skipped (no julia files)"
  fi

  # Groovy
  if has_lang_files "build.gradle build.gradle.kts" "*.groovy *.gvy"; then
    check_runtime "groovy" "Groovy (Modular)"
  else
    log_info "⏭️  Groovy: Skipped (no groovy files)"
  fi

  # Zig
  if has_lang_files "build.zig" "*.zig"; then
    check_runtime "zig" "Zig (Modular)"
  else
    log_info "⏭️  Zig: Skipped (no zig files)"
  fi

  # Tofu
  if has_lang_files "tofu.lock" "*.tf"; then
    check_runtime "tofu" "OpenTofu (Modular)"
  else
    log_info "⏭️  Tofu: Skipped (no tofu files)"
  fi

  # Pulumi
  if has_lang_files "Pulumi.yaml Pulumi.stack.yaml"; then
    check_runtime "pulumi" "Pulumi (Modular)"
  else
    log_info "⏭️  Pulumi: Skipped (no pulumi files)"
  fi

  # Elixir
  if has_lang_files "mix.exs" "*.ex *.exs"; then
    check_runtime "elixir" "Elixir (Modular)"
  else
    log_info "⏭️  Elixir: Skipped (no elixir files)"
  fi

  # Haskell
  if has_lang_files "stack.yaml cabal.project package.yaml" "*.hs *.lhs"; then
    check_runtime "haskell" "Haskell (Modular)"
  else
    log_info "⏭️  Haskell: Skipped (no haskell files)"
  fi

  # Scala
  if has_lang_files "build.sbt build.gradle.kts" "*.scala *.sc"; then
    check_runtime "scala" "Scala (Modular)"
  else
    log_info "⏭️  Scala: Skipped (no scala files)"
  fi

  # Lua
  if has_lang_files "*.lua"; then
    check_runtime "lua" "Lua (Modular)"
  else
    log_info "⏭️  Lua: Skipped (no lua files)"
  fi

  # R
  if has_lang_files "DESCRIPTION" "*.R *.r"; then
    check_runtime "r" "R (Modular)"
  else
    log_info "⏭️  R: Skipped (no R files)"
  fi

  # Mobile Support
  if has_lang_files "Package.swift pubspec.yaml build.gradle.kts" "*.swift *.kt *.dart"; then
    log_info "── Mobile Support ──"
    if has_lang_files "Package.swift" "*.swift"; then check_tool_version "Swift" "swift" "5.0" "swift --version" 0; fi
    if has_lang_files "build.gradle.kts" "*.kt *.kts"; then
      check_tool_version "Kotlin" "kotlin" "1.9.0" "kotlin -version" 0
      check_runtime "kotlin" "Kotlin (Modular)"
    fi
    if [ -f "pubspec.yaml" ] || has_lang_files "" "*.dart"; then
      if resolve_bin "flutter" >/dev/null 2>&1; then
        check_tool_version "Flutter" "flutter" "3.0.0" "flutter --version" 0
      else check_tool_version "Dart" "dart" "3.0.0" "dart --version" 0; fi
    fi
    printf "\n"
  fi

  log_info "── Toolchain Manager ──"
  if resolve_bin "unirtm" >/dev/null 2>&1; then
    log_success "✅ unirtm: Active ($(get_version unirtm))"
  elif resolve_bin "mise" >/dev/null 2>&1; then
    log_success "✅ unirtm/unirtm (legacy fallback): Active ($(get_version unirtm))"
  else
    log_warn "❌ unirtm/unirtm: Not found. (Mandatory for toolchain management)"
    HEALTHY_ST=1
  fi
  printf "\n"

  # 7. Group: Security & Quality Tools
  log_info "── Security & Quality Tools ──"
  check_tool_version "Gitleaks" "gitleaks" "$(get_unirtm_tool_version "github:gitleaks/gitleaks")" "gitleaks version" 0 0 "github:gitleaks/gitleaks"
  # Universal security scanners - Critical in CI, optional locally
  # (1, 1) = Critical in CI (fail if missing), skip locally with info message
  check_tool_version "OSV-scanner" "osv-scanner" "$(get_unirtm_tool_version "github:google/osv-scanner")" "osv-scanner --version" 1 1 "github:google/osv-scanner" "OSV_FORCE_INSTALL"
  # NOTE: Trivy version check removed — scanning delegated to aquasecurity/trivy-action.
  check_tool_version "Zizmor" "zizmor" "$(get_unirtm_tool_version "github:zizmorcore/zizmor")" "zizmor --version" 1 1 "github:zizmorcore/zizmor" "ZIZMOR_FORCE_INSTALL"

  log_info "── Lint & Quality Tools ──"
  check_tool_version "Shfmt" "shfmt" "$(get_unirtm_tool_version "github:mvdan/sh")" "shfmt --version" 0 0 "github:mvdan/sh"
  check_tool_version "Shellcheck" "shellcheck" "$(get_unirtm_tool_version "github:koalaman/shellcheck")" "shellcheck --version" 0 0 "github:koalaman/shellcheck"
  check_tool_version "Actionlint" "actionlint" "$(get_unirtm_tool_version "github:rhysd/actionlint")" "actionlint --version" 0 0 "github:rhysd/actionlint"

  if [ -f "Dockerfile" ] || [ -f "docker-compose.yml" ]; then
    check_tool_version "Hadolint" "hadolint" "$(get_unirtm_tool_version "github:hadolint/hadolint")" "hadolint --version" 0 0 "github:hadolint/hadolint"
  fi
  if has_lang_files "go.mod" "*.go"; then
    check_tool_version "golangci-lint" "golangci-lint" "$(get_unirtm_tool_version golangci-lint)" "golangci-lint --version" 0 0 "golangci-lint"
    check_tool_version "Govulncheck" "govulncheck" "latest" "govulncheck ./..." 0 1 "govulncheck" "GOVULN_FORCE_INSTALL"
  fi
  if has_lang_files "Makefile" "*.make"; then
    check_tool_version "Checkmake" "checkmake" "$(get_unirtm_tool_version "github:checkmake/checkmake")" "checkmake --version" 0 0 "github:checkmake/checkmake"
  fi
  if has_lang_files "Cargo.toml" "*.rs"; then
    check_tool_version "Cargo-audit" "cargo-audit" "latest" "cargo-audit --version" 0 1 "cargo-audit" "CA_FORCE_INSTALL"
  fi
  if [ -f "package.json" ]; then
    check_tool_version "npm-audit" "npm" "$(get_version npm)" "npm audit --version" 0 1 "npm" "NPM_AUDIT_FORCE_INSTALL"
  fi
  if has_lang_files "requirements.txt pyproject.toml" "*.py"; then
    check_tool_version "Pip-audit" "pip-audit" "$(get_unirtm_tool_version pip-audit)" "pip-audit --version" 0 1 "pip-audit" "PA_FORCE_INSTALL"
  fi
  printf "\n"

  # 7. Project File Integrity
  log_info "── Project Integrity ──"
  local _f_chk _has_config=0
  for _f_chk in ".unirtm.toml" ".mise.toml"; do
    if [ -f "${_f_chk:-}" ]; then
      _has_config=1
      break
    fi
  done
  if [ "${_has_config:-0}" -eq 1 ]; then
    log_debug "Found configuration file"
  else
    log_error "❌ Missing critical config file: .unirtm.toml or .mise.toml"
    HEALTHY_ST=1
    CORE_HEALTHY_ST=1
  fi

  for _f_chk in "README.md" ".agent/rules/01-general.md"; do
    if [ -f "${_f_chk:-}" ]; then
      log_debug "Found $_f_chk"
    else
      log_error "❌ Missing critical file: $_f_chk"
      HEALTHY_ST=1
      CORE_HEALTHY_ST=1
    fi
  done
  [ "${CORE_HEALTHY_ST:-0}" -eq 0 ] && log_success "✅ Basic project structure is intact."

  # 8. Script Permissions Audit
  log_info "── Script Permissions ──"
  local _s_chk _p_err=0
  # Use a temporary file to avoid subshell variable loss with find | while read
  local _TMP_SCRIPTS
  _TMP_SCRIPTS=$(mktemp)
  trap 'rm -f "${_TMP_SCRIPTS:-}"' EXIT INT TERM
  find scripts -name "*.sh" -type f >"${_TMP_SCRIPTS:-}"
  while read -r _s_chk; do
    if [ -f "${_s_chk:-}" ]; then
      # Check if it has a shebang
      if head -n 1 "${_s_chk:-}" | grep -q "^#!"; then
        if [ ! -x "${_s_chk:-}" ]; then
          log_error "❌ Script lacks executable bit: $_s_chk"
          _p_err=1
          HEALTHY_ST=1
        fi
      fi
    fi
  done <"${_TMP_SCRIPTS:-}"
  rm -f "${_TMP_SCRIPTS:-}"

  if [ "${_p_err:-}" -eq 0 ]; then
    log_success "✅ All automation scripts have correct executable permissions."
  fi

  finalize_summary_table

  # Final combined health check
  if [ "${HEALTHY_ST:-0}" -eq 0 ]; then
    log_success "\n✨ Environment is HEALTHY! Ready for development."
    exit 0
  elif [ "${CORE_HEALTHY_ST:-0}" -eq 0 ]; then
    log_warn "\n🛠️  Environment is FUNCTIONAL but has warnings (missing recommended/optional tools)."
    log_warn "💡 Run 'unirtm run setup' to address the warnings above."
    exit 0
  else
    log_error "\n❌ Environment is BROKEN. Critical tools or files are missing."
    log_error "Please fix the issues above to proceed."
    exit 1
  fi
}

main "$@"
