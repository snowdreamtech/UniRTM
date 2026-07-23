#!/usr/bin/env bash
set -eu

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
PYPI_DIR="${PROJECT_ROOT}/pypi"
DIST_DIR="${PROJECT_ROOT}/dist"

VERSION=""
while [ $# -gt 0 ]; do
  case "$1" in
  --version)
    VERSION="$2"
    shift 2
    ;;
  *)
    echo "Unknown argument: $1" >&2
    exit 1
    ;;
  esac
done

if [ -z "${VERSION}" ]; then
  echo "ERROR: --version not specified." >&2
  exit 1
fi

VERSION_PY="${VERSION#v}"
echo "Building PyPI packages for version: ${VERSION_PY}"

# Update version in pyproject.toml
sed -i.bak "s/version = \"0.0.0\"/version = \"${VERSION_PY}\"/g" "${PYPI_DIR}/pyproject.toml"
rm -f "${PYPI_DIR}/pyproject.toml.bak"

# Ensure bin directory exists
mkdir -p "${PYPI_DIR}/snowdreamtech_unirtm/bin"
# Copy README
cp "${PROJECT_ROOT}/README.md" "${PYPI_DIR}/"

cd "${PYPI_DIR}"

# map format: "dist_dir|pypi_plat|binary"
PLATFORMS="
unirtm_darwin_arm64_v8.0|macosx_11_0_arm64|unirtm
unirtm_darwin_amd64_v1|macosx_10_9_x86_64|unirtm
unirtm_linux_amd64_v1|manylinux2014_x86_64|unirtm
unirtm_linux_arm64_v8.0|manylinux2014_aarch64|unirtm
unirtm_linux_386_sse2|manylinux2014_i686|unirtm
unirtm_windows_amd64_v1|win_amd64|unirtm.exe
unirtm_windows_arm64_v8.0|win_arm64|unirtm.exe
unirtm_windows_386_sse2|win32|unirtm.exe
"

echo "${PLATFORMS}" | grep -v '^$' | while IFS='|' read -r _dist_subdir _pypi_plat _binary; do
  _src_binary="${DIST_DIR}/${_dist_subdir}/${_binary}"
  _dst_binary="snowdreamtech_unirtm/bin/${_binary}"

  if [ ! -f "${_src_binary}" ]; then
    echo "  ❌ Source binary not found: ${_src_binary}"
    continue
  fi

  echo "🔧 Building wheel for ${_pypi_plat}..."

  # Clean up previous binary
  rm -f snowdreamtech_unirtm/bin/*

  # Copy the binary
  cp "${_src_binary}" "${_dst_binary}"
  chmod +x "${_dst_binary}" 2>/dev/null || true

  # Build the wheel with the specific platform tag
  python3 setup.py bdist_wheel --plat-name="${_pypi_plat}"

  echo "  ✅ Generated wheel for ${_pypi_plat}"
done

echo ""
echo "✨ PyPI package build complete! Wheels are in pypi/dist/"
