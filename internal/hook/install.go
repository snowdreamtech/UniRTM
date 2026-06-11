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
        if [ -x "$HOME/.local/bin/unirtm" ]; then
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
func injectOrUpdateBlock(content string, hasShebang bool) string {
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
			return "#!/bin/sh\n" + managedBlock
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

// InstallBridgeScript writes the standard UniRTM bridge script to .git/hooks/
func InstallBridgeScript(ctx context.Context, dir string, hookName string) error {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--git-path", "hooks")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to resolve git hooks directory: %w", err)
	}

	hooksDir := strings.TrimSpace(string(out))
	if !filepath.IsAbs(hooksDir) {
		hooksDir = filepath.Join(dir, hooksDir)
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

	newContent := injectOrUpdateBlock(content, hasShebang)

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

	return updatedHooks, nil
}
