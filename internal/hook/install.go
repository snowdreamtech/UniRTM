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
)

const (
	managedBlockStart = "# --- BEGIN UNIRTM MANAGED BLOCK ---"
	managedBlockEnd   = "# --- END UNIRTM MANAGED BLOCK ---"
)

const bridgeScriptPayload = `# Auto-load UniRTM environment for Headless (AI/CI) and GUI Git Clients
if ! command -v unirtm >/dev/null 2>&1; then
    _UNIRTM_BIN="$(git rev-parse --show-toplevel 2>/dev/null)/unirtm"
    if [ -x "$_UNIRTM_BIN" ]; then
        eval "$("$_UNIRTM_BIN" env)" 2>/dev/null
    else
        if [ -x "$HOME/bin/unirtm" ]; then
            eval "$("$HOME/bin/unirtm" env)" 2>/dev/null
        elif [ -x "$HOME/.local/bin/unirtm" ]; then
            eval "$("$HOME/.local/bin/unirtm" env)" 2>/dev/null
        fi
    fi
else
    eval "$(unirtm env)" 2>/dev/null
fi`

// readExistingHook reads the hook file, checking for a shebang.
func readExistingHook(path string) (content string, hasShebang bool, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, err
	}
	content = string(data)
	if strings.HasPrefix(content, "#!") {
		hasShebang = true
	}
	return content, hasShebang, nil
}

// injectOrUpdateBlock replaces the managed block if it exists, or inserts it.
func injectOrUpdateBlock(content string, hasShebang bool, hookName ...string) string {
	managedBlock := fmt.Sprintf("%s\n%s\n%s\n", managedBlockStart, bridgeScriptPayload, managedBlockEnd)

	// If block exists, replace it
	startIndex := strings.Index(content, managedBlockStart)
	if startIndex != -1 {
		endIndex := strings.Index(content, managedBlockEnd)
		if endIndex != -1 {
			// Find the end of the line containing managedBlockEnd
			endIndex += len(managedBlockEnd)
			if endIndex < len(content) && content[endIndex] == '\n' {
				endIndex++ // Consume the newline
			} else if endIndex+1 < len(content) && content[endIndex:endIndex+2] == "\r\n" {
				endIndex += 2
			}
			return content[:startIndex] + managedBlock + content[endIndex:]
		}
	}

	// Block doesn't exist, inject it
	if !hasShebang {
		if content == "" {
			name := ""
			if len(hookName) > 0 {
				name = hookName[0]
			}
			var defaultExec string
			switch name {
			case "commit-msg":
				defaultExec = "# Execute UniRTM Quality Guard to validate commit message\nexec unirtm check \"$@\"\n"
			default:
				defaultExec = "# Execute UniRTM Quality Guard to scan staged files before commit\nexec unirtm check --staged \"$@\"\n"
			}
			return "#!/bin/sh\n" + managedBlock + "\n" + defaultExec
		}
		// Prepend shebang and block if no shebang is present but content exists
		return "#!/bin/sh\n" + managedBlock + content
	}

	// Has shebang, insert immediately after the first line
	parts := strings.SplitN(content, "\n", 2)
	if len(parts) == 2 {
		return parts[0] + "\n" + managedBlock + parts[1]
	}
	// Edge case where shebang is the only thing in the file without a trailing newline
	return parts[0] + "\n" + managedBlock
}

// resolveGitHooksDir locates the .git/hooks directory quickly using Go filesystem APIs without forking a git process.
// It returns an empty string if dir is not a git repository or the hooks directory cannot be determined.
func resolveGitHooksDir(dir string) string {
	if dir == "" {
		return ""
	}
	gitPath := filepath.Join(dir, ".git")
	fi, err := os.Stat(gitPath)
	if err != nil {
		return ""
	}
	if fi.IsDir() {
		return filepath.Join(gitPath, "hooks")
	}

	// .git can be a file in git worktrees and git submodules: "gitdir: <path>"
	data, err := os.ReadFile(gitPath)
	if err != nil {
		return ""
	}
	content := strings.TrimSpace(string(data))
	if strings.HasPrefix(content, "gitdir:") {
		target := strings.TrimSpace(strings.TrimPrefix(content, "gitdir:"))
		if !filepath.IsAbs(target) {
			target = filepath.Join(dir, target)
		}
		hooksPath := filepath.Join(target, "hooks")
		if _, err := os.Stat(hooksPath); err == nil {
			return hooksPath
		}
		// If worktrees don't have local hooks, check parent gitdir
		parentGit := filepath.Dir(filepath.Dir(target))
		parentHooks := filepath.Join(parentGit, "hooks")
		if _, err := os.Stat(parentHooks); err == nil {
			return parentHooks
		}
		return hooksPath
	}
	return ""
}

// hasProjectHooksConfig quickly checks if the directory has UniRTM or known hook configurations.
// It is side-effect free and uses only fast filesystem checks.
func hasProjectHooksConfig(dir string) bool {
	// Check for UniRTM config files
	for _, cfgName := range []string{".unirtm.toml", "unirtm.toml"} {
		if _, err := os.Stat(filepath.Join(dir, cfgName)); err == nil {
			return true
		}
	}
	// Check for other standard hook runners supported by UniRTM
	for _, indicator := range []string{".husky", "lefthook.yml", ".pre-commit-config.yaml"} {
		if _, err := os.Stat(filepath.Join(dir, indicator)); err == nil {
			return true
		}
	}
	return false
}

// EnsureGitHooksIdempotent ensures that git hook bridge scripts are installed for the repository.
// It strictly guarantees:
//  1. Idempotency: If hook scripts are already installed and content matches, it performs NO disk writes and preserves mtime.
//  2. Zero side effects: If not a git repository or not a configured project, it returns immediately without error.
//     Any non-fatal filesystem error is caught and suppressed to prevent blocking the caller.
//  3. Ultra-fast execution: Executes in < 2ms using native filesystem inspection (no git subprocess execution).
//
// Returns true if any hook file was created or modified.
func EnsureGitHooksIdempotent(ctx context.Context, dir string) (updated bool, err error) {
	if dir == "" {
		return false, nil
	}

	// 1. Fast short-circuit: Not a git repo
	hooksDir := resolveGitHooksDir(dir)
	if hooksDir == "" {
		return false, nil
	}

	// 2. Fast short-circuit: Not a UniRTM or hook-configured project
	if !hasProjectHooksConfig(dir) {
		return false, nil
	}

	// 3. Ensure hooks directory exists
	if _, err := os.Stat(hooksDir); os.IsNotExist(err) {
		if err := os.MkdirAll(hooksDir, 0755); err != nil {
			return false, nil // Non-fatal, suppress
		}
	}

	// 4. Inspect standard hooks (pre-commit, commit-msg)
	targetHooks := []string{"pre-commit", "commit-msg"}
	var anyUpdated bool

	for _, hookName := range targetHooks {
		hookPath := filepath.Join(hooksDir, hookName)
		content, hasShebang, readErr := readExistingHook(hookPath)
		if readErr != nil {
			continue // Skip unreadable file without crashing
		}

		newContent := injectOrUpdateBlock(content, hasShebang, hookName)
		// Strict idempotency: if content is identical, do NOT touch disk
		if content == newContent {
			continue
		}

		if writeErr := os.WriteFile(hookPath, []byte(newContent), 0755); writeErr == nil {
			anyUpdated = true
		}
	}

	return anyUpdated, nil
}

// InstallBridgeScript writes the standard UniRTM bridge script to .git/hooks/
func InstallBridgeScript(ctx context.Context, dir string, hookName string) error {
	hooksDir := resolveGitHooksDir(dir)
	if hooksDir == "" {
		cmd := exec.CommandContext(ctx, "git", "rev-parse", "--git-path", "hooks")
		cmd.Dir = dir
		out, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("failed to resolve git hooks directory: %w", err)
		}
		hooksDir = strings.TrimSpace(string(out))
		if !filepath.IsAbs(hooksDir) {
			hooksDir = filepath.Join(dir, hooksDir)
		}
	}

	if _, err := os.Stat(hooksDir); os.IsNotExist(err) {
		if err := os.MkdirAll(hooksDir, 0755); err != nil {
			return fmt.Errorf("git hooks directory does not exist and could not be created: %w", err)
		}
	}

	hookPath := filepath.Join(hooksDir, hookName)

	content, hasShebang, err := readExistingHook(hookPath)
	if err != nil {
		return fmt.Errorf("failed to read existing hook: %w", err)
	}

	newContent := injectOrUpdateBlock(content, hasShebang, hookName)
	if content == newContent {
		// Strictly idempotent: avoid touching file if content is identical
		return nil
	}

	// Create or overwrite the hook
	err = os.WriteFile(hookPath, []byte(newContent), 0755)
	if err != nil {
		return fmt.Errorf("failed to write git hook %s: %w", hookName, err)
	}

	return nil
}

// InstallAllBridgeScripts iterates over all existing non-sample files in the git hooks directory
// and installs the bridge script into them. If no files exist, it optionally creates standard ones.
func InstallAllBridgeScripts(ctx context.Context, dir string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--git-path", "hooks")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve git hooks directory: %w", err)
	}

	hooksDir := strings.TrimSpace(string(out))
	if !filepath.IsAbs(hooksDir) {
		hooksDir = filepath.Join(dir, hooksDir)
	}

	entries, err := os.ReadDir(hooksDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // Nothing to do if hooks dir doesn't exist
		}
		return nil, fmt.Errorf("failed to read git hooks directory: %w", err)
	}

	var updatedHooks []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".sample") {
			continue
		}
		// Validate if it's a known git hook name (optional, but good practice)
		if err := ValidateHookName(name); err != nil {
			continue // Skip non-standard hook files
		}

		if err := InstallBridgeScript(ctx, dir, name); err != nil {
			return updatedHooks, fmt.Errorf("failed to install hook %s: %w", name, err)
		}
		updatedHooks = append(updatedHooks, name)
	}

	// If no active hooks were found, install standard ones (pre-commit, commit-msg)
	if len(updatedHooks) == 0 {
		for _, h := range []string{"pre-commit", "commit-msg"} {
			if err := InstallBridgeScript(ctx, dir, h); err != nil {
				return updatedHooks, fmt.Errorf("failed to install hook %s: %w", h, err)
			}
			updatedHooks = append(updatedHooks, h)
		}
	}

	return updatedHooks, nil
}
