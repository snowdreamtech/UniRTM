#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Markdown Logic Module

# Purpose: Installs markdownlint for Markdown linting.
# Delegate: Managed by unirtm (.unirtm.toml)
install_markdownlint() {
  install_tool_safe "markdownlint-cli2" "${VER_MARKDOWNLINT_PROVIDER:-}" "Markdownlint" "--version" 0 "*.md" ""
}

# Purpose: Sets up Markdown environment.
setup_markdown() {
  install_markdownlint
}
