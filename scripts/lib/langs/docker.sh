#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Docker Logic Module

# Purpose: Installs hadolint for Dockerfile linting.
# Delegate: Managed by unirtm (.unirtm.toml)
install_hadolint() {
  install_tool_safe "hadolint" "${VER_HADOLINT_PROVIDER:-}" "Hadolint" "--version" 0 "Dockerfile docker-compose.yaml docker-compose.yml compose.yaml compose.yml" ""
}

# Purpose: Installs dockerfile-utils for Dockerfile management.
# Delegate: Managed by unirtm (.unirtm.toml)
install_dockerfile_utils() {
  install_tool_safe "dockerfile-utils" "${VER_DOCKERFILE_UTILS_PROVIDER:-}" "Dockerfile Utils" "--version" 0 "Dockerfile docker-compose.yaml docker-compose.yml compose.yaml compose.yml" ""
}

# Purpose: Sets up Docker environment.
setup_docker() {
  install_hadolint
  install_dockerfile_utils
}
