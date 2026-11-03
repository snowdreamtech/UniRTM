# Feature Spec: Auto Update Notifier

## Overview

## Requirements

1. **Background Check**:
   - Periodically check for new releases from GitHub API (`https://api.github.com/repos/snowdreamtech/UniRTM/releases/latest`).
   - The check must be fully asynchronous and must not block command execution.
   - Cache results locally (e.g., in `~/.unirtm/update-cache.json`).

2. **Cool-down / Debounce**:
   - Check frequency: Check GitHub API at most once every 24 hours.
   - Prompt frequency: Display the update prompt to the user at most once every 24 hours.

3. **Atomic & Safe Execution**:
   - The prompt should only be displayed to an interactive terminal (`stderr`).
   - It should NOT be displayed when running non-interactive scripts or when piping output.
   - It should NOT be displayed for blacklisted commands (`env`, `completion`, `version`, `self-update`) to prevent breaking tool integrations or user experiences.

4. **User Experience**:
   - The warning should look like:

     ```
     unirtm WARN  unirtm version X.X.X available
     unirtm WARN  To update, run unirtm self-update
     ```
