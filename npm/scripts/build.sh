#!/usr/bin/env sh
# build.sh - Build npm platform packages from GoReleaser dist/ output
#
# Usage:
#   sh npm/scripts/build.sh [--version <version>] [--dist-dir <path>] [--npm-dir <path>]
#
# Reads version from VERSION file if not specified.
# Must be run from the project root.

set -eu

# ---------------------------------------------------------------------------
# Script location & project root detection
# ---------------------------------------------------------------------------
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

# ---------------------------------------------------------------------------
# Default paths
# ---------------------------------------------------------------------------
DIST_DIR="${PROJECT_ROOT}/dist"
NPM_DIR="${PROJECT_ROOT}/npm"
VERSION_FILE="${PROJECT_ROOT}/VERSION"

# ---------------------------------------------------------------------------
# Parse arguments
# ---------------------------------------------------------------------------
VERSION=""
while [ $# -gt 0 ]; do
	case "$1" in
	--version)
		VERSION="$2"
		shift 2
		;;
	--dist-dir)
		DIST_DIR="$2"
		shift 2
		;;
	--npm-dir)
		NPM_DIR="$2"
		shift 2
		;;
	*)
		printf 'Unknown argument: %s\n' "$1" >&2
		exit 1
		;;
	esac
done

# ---------------------------------------------------------------------------
# Resolve version
# ---------------------------------------------------------------------------
if [ -z "${VERSION}" ]; then
	if [ -f "${VERSION_FILE}" ]; then
		VERSION="$(cat "${VERSION_FILE}" | tr -d '[:space:]')"
	else
		printf 'ERROR: VERSION file not found and --version not specified.\n' >&2
		exit 1
	fi
fi

# Strip leading 'v' for npm semver compatibility
VERSION_NPM="${VERSION#v}"

printf '✅ Building npm packages for version: %s\n' "${VERSION_NPM}"
printf '   dist: %s\n' "${DIST_DIR}"
printf '   npm:  %s\n' "${NPM_DIR}"

# ---------------------------------------------------------------------------
# Platform mapping: <npm-package-dir>:<dist-subdir>:<binary-name>
# Format: "npm_dir|dist_dir|binary"
# ---------------------------------------------------------------------------
PLATFORMS="
unirtm-darwin-arm64|unirtm_darwin_arm64_v8.0|unirtm
unirtm-darwin-x64|unirtm_darwin_amd64_v1|unirtm
unirtm-linux-x64|unirtm_linux_amd64_v1|unirtm
unirtm-linux-arm64|unirtm_linux_arm64_v8.0|unirtm
unirtm-linux-ia32|unirtm_linux_386_sse2|unirtm
unirtm-linux-arm|unirtm_linux_arm_7|unirtm
unirtm-linux-arm-5|unirtm_linux_arm_5|unirtm
unirtm-linux-arm-6|unirtm_linux_arm_6|unirtm
unirtm-linux-loong64|unirtm_linux_loong64|unirtm
unirtm-linux-ppc64le|unirtm_linux_ppc64le_power8|unirtm
unirtm-linux-riscv64|unirtm_linux_riscv64_rva20u64|unirtm
unirtm-linux-s390x|unirtm_linux_s390x|unirtm
unirtm-windows-x64|unirtm_windows_amd64_v1|unirtm.exe
unirtm-windows-arm64|unirtm_windows_arm64_v8.0|unirtm.exe
unirtm-windows-ia32|unirtm_windows_386_sse2|unirtm.exe
"

# ---------------------------------------------------------------------------
# Helper: generate package.json from .tpl file
# ---------------------------------------------------------------------------
generate_package_json() {
	local _pkg_dir="$1"
	local _version="$2"
	local _tpl="${_pkg_dir}/package.json.tpl"
	local _out="${_pkg_dir}/package.json"

	if [ ! -f "${_tpl}" ]; then
		printf '  ⚠️  Template not found: %s (skipping package.json generation)\n' "${_tpl}"
		return 0
	fi

	# Replace {{VERSION}} placeholder
	sed "s/{{VERSION}}/${_version}/g" "${_tpl}" >"${_out}"
	printf '  ✅ Generated: %s\n' "${_out#"${PROJECT_ROOT}"/}"
}

# ---------------------------------------------------------------------------
# Helper: copy documentation files
# ---------------------------------------------------------------------------
copy_docs() {
	local _pkg_dir="$1"

	for _doc in LICENSE README.md README_zh-CN.md; do
		if [ -f "${PROJECT_ROOT}/${_doc}" ]; then
			cp "${PROJECT_ROOT}/${_doc}" "${_pkg_dir}/${_doc}"
		fi
	done
}

# ---------------------------------------------------------------------------
# Process each platform
# ---------------------------------------------------------------------------
printf '\n📦 Processing platform packages...\n'

printf '%s' "${PLATFORMS}" | grep -v '^$' | while IFS='|' read -r _npm_pkg _dist_subdir _binary; do
	_pkg_dir="${NPM_DIR}/${_npm_pkg}"
	_src_binary="${DIST_DIR}/${_dist_subdir}/${_binary}"
	_bin_dir="${_pkg_dir}/bin"
	_dst_binary="${_bin_dir}/${_binary}"

	printf '\n🔧 %s\n' "${_npm_pkg}"

	# Verify source binary exists
	if [ ! -f "${_src_binary}" ]; then
		printf '  ❌ Source binary not found: %s\n' "${_src_binary}"
		printf '     Skipping (run GoReleaser build first)\n'
		continue
	fi

	# Create bin directory
	mkdir -p "${_bin_dir}"

	# Copy binary
	cp "${_src_binary}" "${_dst_binary}"

	# Set executable permission (no-op on Windows binaries, harmless)
	chmod +x "${_dst_binary}"

	printf '  ✅ Binary: %s -> %s\n' \
		"${_src_binary#"${PROJECT_ROOT}"/}" \
		"${_dst_binary#"${PROJECT_ROOT}"/}"

	# Generate package.json from template
	generate_package_json "${_pkg_dir}" "${VERSION_NPM}"

	# Copy documentation
	copy_docs "${_pkg_dir}"
done

# ---------------------------------------------------------------------------
# Process root package
# ---------------------------------------------------------------------------
printf '\n🔧 unirtm (root package)\n'
_root_pkg_dir="${NPM_DIR}/unirtm"
generate_package_json "${_root_pkg_dir}" "${VERSION_NPM}"
copy_docs "${_root_pkg_dir}"

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
printf '\n✨ npm package build complete!\n'
printf '   Version: %s\n' "${VERSION_NPM}"
printf '\nNext steps:\n'
printf '  1. Publish platform packages first:\n'
# shellcheck disable=SC2016
printf '%s\n' '       for pkg in npm/unirtm-*/; do npm publish "$pkg" --access public; done'
printf '  2. Then publish root package:\n'
printf '       npm publish npm/unirtm/ --access public\n'
