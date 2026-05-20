#!/usr/bin/env sh
set -eu
# Copyright (c) 2026 SnowdreamTech. All rights reserved.
# Licensed under the MIT License. See LICENSE file in the project root for full license information.

# Runner Logic Module

# Purpose: Installs Just (modern runner).
# Delegate: Managed by unirtm (.unirtm.toml)
install_just() {
  setup_registry_just
  install_tool_safe "just" "${VER_JUST_PROVIDER:-}" "Just" "--version" 0 "JUST" ""
}

# Purpose: Installs Task (modern runner).
# Delegate: Managed by unirtm (.unirtm.toml)
install_task() {
  setup_registry_task
  install_tool_safe "task" "${VER_TASK_PROVIDER:-}" "Task" "--version" 0 "TASK" ""
}

# Purpose: Sets up Runners environment.
setup_runners() {
  # Skip if no runner config files exist (unless forced)
  if [ "${FORCE_SETUP:-0}" -eq 0 ]; then
    if ! has_lang_files "" "justfile Justfile Taskfile.yml Taskfile.yaml"; then
      log_info "⏭️  Skipping runners: No justfile or Taskfile detected"
      log_summary "Runners" "just/task" "⏭️ Skipped" "-" "0"
      return 0
    fi
  fi

  install_just
  install_task
}
