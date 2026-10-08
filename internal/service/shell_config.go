// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
)

// ShellConfigManager handles persistent configuration changes in shell RC files.
type ShellConfigManager struct {
	formatter output.Formatter
	dryRun    bool
}

// NewShellConfigManager creates a new ShellConfigManager.
func NewShellConfigManager(formatter output.Formatter, dryRun bool) *ShellConfigManager {
	return &ShellConfigManager{
		formatter: formatter,
		dryRun:    dryRun,
	}
}

// GetConfigPath returns the standard configuration file path for the given shell.
func (m *ShellConfigManager) GetConfigPath(shell ShellType) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	switch shell {
	case ShellZsh:
		return filepath.Join(home, ".zshrc"), nil
	case ShellBash:
		return filepath.Join(home, ".bashrc"), nil
	case ShellFish:
		return filepath.Join(home, ".config/fish/config.fish"), nil
	case ShellPowerShell:
		configFile := env.Get("PROFILE")
		if configFile != "" {
			return configFile, nil
		}
		pwshCore := filepath.Join(home, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1")
		if env.RuntimeGOOS == "windows" {
			winPwsh := filepath.Join(home, "Documents", "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1")
			if _, err := os.Stat(winPwsh); err == nil {
				if _, err := os.Stat(pwshCore); os.IsNotExist(err) {
					return winPwsh, nil
				}
			}
		}
		return pwshCore, nil
	default:
		return "", fmt.Errorf("unsupported shell: %s", shell)
	}
}

// Inject appends or updates a configuration block in the shell RC file.
func (m *ShellConfigManager) Inject(shell ShellType, marker string, content string) error {
	configFile, err := m.GetConfigPath(shell)
	if err != nil {
		return err
	}

	// 1. Read existing content
	rawContent, err := os.ReadFile(configFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	searchPattern := fmt.Sprintf("unirtm %s activation", marker)
	fullBlock := fmt.Sprintf("\n# %s\n%s\n", searchPattern, content)

	rawContentStr := string(rawContent)
	if strings.Contains(rawContentStr, searchPattern) {
		// Already present, check if we need to update
		if strings.Contains(rawContentStr, content) {
			m.formatter.Info(fmt.Sprintf("UniRTM %s logic already up to date in %s", marker, configFile), nil)
			return nil
		}

		if m.dryRun {
			m.formatter.Info(fmt.Sprintf("[dry-run] Would update %s activation logic in %s", marker, configFile), nil)
			return nil
		}

		// Update by replacing the old block
		lines := strings.Split(rawContentStr, "\n")
		var newLines []string
		inBlock := false
		replaced := false
		seenActivation := false
		inIfWrapper := false
		markerLower := strings.ToLower(marker)

		for i := 0; i < len(lines); i++ {
			line := lines[i]
			trimmed := strings.TrimSpace(line)
			trimmedLower := strings.ToLower(trimmed)

			if strings.Contains(line, searchPattern) {
				inBlock = true
				seenActivation = false
				inIfWrapper = false
				if !replaced {
					// Add the new block here
					newLines = append(newLines, "# "+searchPattern)
					newLines = append(newLines, content)
					replaced = true
				}
				continue
			}

			if inBlock {
				if strings.Contains(trimmedLower, markerLower) &&
					(strings.Contains(trimmedLower, "activate") || strings.Contains(trimmedLower, "hook-env") || strings.Contains(trimmedLower, "eval") || strings.Contains(trimmedLower, "source")) {
					seenActivation = true
				}

				if strings.HasPrefix(trimmedLower, "if ") || strings.Contains(trimmedLower, "get-command") {
					inIfWrapper = true
				}

				if seenActivation {
					if inIfWrapper {
						if trimmed == "fi" || trimmed == "end" || trimmed == "}" {
							inBlock = false
						}
					} else {
						inBlock = false
					}
				} else if !isBlockComponent(trimmedLower) {
					inBlock = false
				}
				continue
			}

			newLines = append(newLines, lines[i])
		}

		newContent := strings.Join(newLines, "\n")
		if err := os.WriteFile(configFile, []byte(newContent), 0600); err != nil {
			return err
		}
		m.formatter.Success(fmt.Sprintf("Updated %s activation logic in %s", marker, configFile))
		return nil
	}

	if m.dryRun {
		m.formatter.Info(fmt.Sprintf("[dry-run] Would add %s activation logic to %s", marker, configFile), nil)
		return nil
	}

	// 2. Ensure file and directory exist
	if os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(configFile), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(configFile, []byte(""), 0600); err != nil {
			return err
		}
	}

	// 3. Prepare content with consistent spacing
	cleanContent := strings.TrimRight(rawContentStr, " \t\r\n")

	// 4. Append block
	f, err := os.OpenFile(configFile, os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString(cleanContent + "\n" + fullBlock); err != nil {
		return err
	}

	m.formatter.Success(fmt.Sprintf("Added %s activation logic to %s", marker, configFile))
	return nil
}

// Remove removes configuration lines related to the given marker from the shell RC file.
func (m *ShellConfigManager) Remove(shell ShellType, marker string) error {
	configFile, err := m.GetConfigPath(shell)
	if err != nil {
		return err
	}

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil
	}

	if m.dryRun {
		m.formatter.Info(fmt.Sprintf("[dry-run] Would remove %s logic from %s", marker, configFile), nil)
		return nil
	}

	rawBytes, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	contentStr := string(rawBytes)
	isCRLF := strings.Contains(contentStr, "\r\n")
	normalized := strings.ReplaceAll(contentStr, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")

	markerLower := strings.ToLower(marker)
	possibleComments := []string{
		fmt.Sprintf("unirtm %s activation", markerLower),
		fmt.Sprintf("unirtm %s", markerLower),
		fmt.Sprintf("%s activation", markerLower),
	}
	if markerLower == "unirtm" {
		possibleComments = append(possibleComments, "unirtm activation")
	}

	isCommentMarker := func(line string) bool {
		trimmed := strings.ToLower(strings.TrimSpace(line))
		if !strings.HasPrefix(trimmed, "#") {
			return false
		}
		for _, pat := range possibleComments {
			if strings.Contains(trimmed, pat) {
				return true
			}
		}
		return false
	}

	isActivationCmd := func(line string) bool {
		trimmed := strings.ToLower(strings.TrimSpace(line))
		if strings.HasPrefix(trimmed, "#") {
			return false
		}
		// Must reference the target tool
		if !strings.Contains(trimmed, markerLower) {
			return false
		}
		// And must contain activation keywords
		return strings.Contains(trimmed, "activate") ||
			strings.Contains(trimmed, "hook-env") ||
			strings.Contains(trimmed, "source") ||
			strings.Contains(trimmed, "eval") ||
			strings.Contains(trimmed, "invoke-expression") ||
			strings.Contains(trimmed, "out-string") ||
			strings.Contains(trimmed, "invoke-restmethod")
	}

	var newLines []string
	removedCount := 0
	inBlock := false
	seenActivation := false
	inIfWrapper := false

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		trimmedLower := strings.ToLower(trimmed)

		// 1. Check if this line starts a managed block
		if isCommentMarker(line) {
			inBlock = true
			seenActivation = false
			inIfWrapper = false
			removedCount++
			continue
		}

		// 2. If inside a managed block, remove all lines of the block
		if inBlock {
			removedCount++

			if strings.Contains(trimmedLower, markerLower) &&
				(strings.Contains(trimmedLower, "activate") || strings.Contains(trimmedLower, "hook-env") || strings.Contains(trimmedLower, "eval") || strings.Contains(trimmedLower, "source")) {
				seenActivation = true
			}

			if strings.HasPrefix(trimmedLower, "if ") || strings.Contains(trimmedLower, "get-command") {
				inIfWrapper = true
			}

			if seenActivation {
				if inIfWrapper {
					if trimmed == "fi" || trimmed == "end" || trimmed == "}" {
						inBlock = false
					}
				} else {
					inBlock = false
				}
			} else if !isBlockComponent(trimmedLower) {
				inBlock = false
			}
			continue
		}

		// 3. Fallback: Standalone activation command without comment marker
		if isActivationCmd(line) {
			removedCount++
			continue
		}

		newLines = append(newLines, line)
	}

	if removedCount == 0 {
		return nil
	}

	// Clean up trailing empty lines
	for len(newLines) > 0 && strings.TrimSpace(newLines[len(newLines)-1]) == "" {
		newLines = newLines[:len(newLines)-1]
	}

	lineSep := "\n"
	if isCRLF {
		lineSep = "\r\n"
	}
	output := strings.Join(newLines, lineSep)
	if len(newLines) > 0 {
		output += lineSep
	}

	if err := os.WriteFile(configFile, []byte(output), 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	m.formatter.Success(fmt.Sprintf("Removed %s logic from %s (%d lines removed)", marker, configFile, removedCount))
	return nil
}

func isBlockComponent(lineLower string) bool {
	if lineLower == "" || strings.HasPrefix(lineLower, "#") {
		return true
	}
	return strings.Contains(lineLower, "path") ||
		strings.Contains(lineLower, "userbin") ||
		strings.Contains(lineLower, "if ") ||
		strings.Contains(lineLower, "then") ||
		strings.Contains(lineLower, "foreach") ||
		strings.Contains(lineLower, "test ") ||
		lineLower == "fi" ||
		lineLower == "end" ||
		lineLower == "}" ||
		strings.Contains(lineLower, "command -v") ||
		strings.Contains(lineLower, "type -q") ||
		strings.Contains(lineLower, "get-command")
}

