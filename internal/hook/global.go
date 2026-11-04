// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package hook

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
)

// GetGlobalHooksDir returns the path to the global git hooks directory managed by UniRTM.
func GetGlobalHooksDir() string {
	return filepath.Join(env.GetConfigDir(), "hooks")
}

// GenerateGlobalHookScript generates a robust, zero-side-effect global hook script.
// When git invokes this hook via core.hooksPath across ANY repository on the machine:
//  1. It auto-loads the UniRTM environment for GUI clients and headless AI Agents.
//  2. If the current repository is NOT a UniRTM project, it transparently delegates
//     to the repository's local .git/hooks/<hookName> (if executable) or exits 0.
//  3. If the repository IS a UniRTM project, it invokes UniRTM Quality Guard or hook engine.
func GenerateGlobalHookScript(hookName string) string {
	managedBlock := fmt.Sprintf("%s\n%s\n%s\n", managedBlockStart, bridgeScriptPayload, managedBlockEnd)

	script := `#!/bin/sh
` + managedBlock + `
# Locate repository root
REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null)"
if [ -z "$REPO_ROOT" ]; then
    exit 0
fi

HOOK_NAME="$(basename "$0")"

# Transparent pass-through: if not a UniRTM project, delegate to local repo hook or exit 0
is_unirtm_project() {
    [ -f "$REPO_ROOT/.unirtm.toml" ] || [ -f "$REPO_ROOT/unirtm.toml" ] || \
    [ -d "$REPO_ROOT/.husky" ] || [ -f "$REPO_ROOT/lefthook.yml" ] || [ -f "$REPO_ROOT/.pre-commit-config.yaml" ]
}

if ! is_unirtm_project; then
    GIT_DIR="$(git rev-parse --git-dir 2>/dev/null)"
    if [ -n "$GIT_DIR" ] && [ -x "$GIT_DIR/hooks/$HOOK_NAME" ]; then
        exec "$GIT_DIR/hooks/$HOOK_NAME" "$@"
    fi
    exit 0
fi

# Execute Quality Guard or hook engine for UniRTM projects
case "$HOOK_NAME" in
    commit-msg)
        exec unirtm check "$@"
        ;;
    pre-commit)
        exec unirtm check --staged "$@"
        ;;
    *)
        exec unirtm hook run "$HOOK_NAME" "$@"
        ;;
esac
`
	return script
}

// InstallGlobalBridgeScript writes a global hook script into the UniRTM global hooks directory.
func InstallGlobalBridgeScript(ctx context.Context, hookName string) error {
	if err := ValidateHookName(hookName); err != nil {
		return err
	}

	globalDir := GetGlobalHooksDir()
	if err := os.MkdirAll(globalDir, 0755); err != nil {
		return fmt.Errorf("failed to create global hooks directory %s: %w", globalDir, err)
	}

	hookPath := filepath.Join(globalDir, hookName)
	content := GenerateGlobalHookScript(hookName)

	if existing, err := os.ReadFile(hookPath); err == nil && string(existing) == content {
		return nil // Strictly idempotent
	}

	if err := os.WriteFile(hookPath, []byte(content), 0755); err != nil {
		return fmt.Errorf("failed to write global hook %s: %w", hookName, err)
	}

	return nil
}

// ConfigureGitCoreHooksPath sets git config --global core.hooksPath to the UniRTM global hooks directory.
func ConfigureGitCoreHooksPath(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	globalDir := GetGlobalHooksDir()
	cmd := exec.CommandContext(ctx, "git", "config", "--global", "core.hooksPath", globalDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to set git core.hooksPath: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

// UnsetGitCoreHooksPath unsets git config --global core.hooksPath if it points to UniRTM's directory.
func UnsetGitCoreHooksPath(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	cmdGet := exec.CommandContext(ctx, "git", "config", "--global", "--get", "core.hooksPath")
	out, err := cmdGet.Output()
	if err != nil {
		return nil // Not set or error, nothing to unset
	}

	currentPath := strings.TrimSpace(string(out))
	globalDir := GetGlobalHooksDir()
	if currentPath == globalDir {
		cmdUnset := exec.CommandContext(ctx, "git", "config", "--global", "--unset", "core.hooksPath")
		if unsetOut, unsetErr := cmdUnset.CombinedOutput(); unsetErr != nil {
			return fmt.Errorf("failed to unset git core.hooksPath: %s: %w", strings.TrimSpace(string(unsetOut)), unsetErr)
		}
	}
	return nil
}

// InstallAllGlobalBridgeScripts installs standard global hooks and configures git core.hooksPath.
func InstallAllGlobalBridgeScripts(ctx context.Context) ([]string, error) {
	hooks := []string{"pre-commit", "commit-msg"}
	for _, h := range hooks {
		if err := InstallGlobalBridgeScript(ctx, h); err != nil {
			return nil, err
		}
	}

	if err := ConfigureGitCoreHooksPath(ctx); err != nil {
		return nil, err
	}

	return hooks, nil
}
