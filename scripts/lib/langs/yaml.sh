#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# YAML Logic Module

# Purpose: Installs yamllint for YAML linting.
# Delegate: Managed by unirtm (.unirtm.toml)
install_yamllint() {
  install_tool_safe "yamllint" "${VER_YAMLLINT_PROVIDER:-}" "Yamllint" "--version" 0 "*.yaml *.yml" ""
}

# Purpose: Installs dotenv-linter for .env file linting.
# Delegate: Managed by unirtm (.unirtm.toml)
install_dotenv_linter() {
  install_tool_safe "dotenv-linter" "${VER_DOTENV_LINTER_PROVIDER:-}" "Dotenv-Linter" "--version" 0 "" ".env .env.example .env.template"
}

# Purpose: Sets up YAML and Env environment.
setup_yaml() {
  install_yamllint
  install_dotenv_linter
}
