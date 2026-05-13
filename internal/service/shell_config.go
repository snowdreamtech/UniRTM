// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/snowdreamtech/unirtm/internal/cli/output"
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
		configFile := os.Getenv("PROFILE")
		if configFile == "" {
			configFile = filepath.Join(home, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1")
		}
		return configFile, nil
	default:
		return "", fmt.Errorf("unsupported shell: %s", shell)
	}
}

// Inject appends a configuration block to the shell RC file if it doesn't already exist.
func (m *ShellConfigManager) Inject(shell ShellType, marker string, content string) error {
	configFile, err := m.GetConfigPath(shell)
	if err != nil {
		return err
	}

	// 1. Check if already present
	rawContent, err := os.ReadFile(configFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	searchPattern := fmt.Sprintf("# unirtm %s", marker)
	if strings.Contains(string(rawContent), searchPattern) {
		m.formatter.Info(fmt.Sprintf("UniRTM %s logic already present in %s", marker, configFile), nil)
		return nil
	}

	if m.dryRun {
		m.formatter.Info(fmt.Sprintf("[dry-run] Would update %s with %s activation logic", configFile, marker), nil)
		return nil
	}

	// 2. Ensure file and directory exist
	if os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(configFile), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(configFile, []byte(""), 0644); err != nil {
			return err
		}
	}

	// 3. Prepare content with consistent spacing
	cleanContent := strings.TrimRight(string(rawContent), " \t\r\n")
	
	// 4. Append block
	f, err := os.OpenFile(configFile, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	block := fmt.Sprintf("\n\n# unirtm %s activation\n%s\n", marker, content)
	if _, err := f.WriteString(cleanContent + block); err != nil {
		return err
	}

	m.formatter.Success(fmt.Sprintf("Updated %s with %s logic", configFile, marker))
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

	content, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	searchPattern := fmt.Sprintf("unirtm %s", marker)
	if !strings.Contains(string(content), searchPattern) {
		return nil
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string
	removedCount := 0
	
	for _, line := range lines {
		// Remove the comment marker line
		if strings.Contains(line, searchPattern) {
			removedCount++
			continue
		}
		// Also remove the specific source/activation line if we find it
		if strings.Contains(line, "unirtm") && strings.Contains(line, marker) && 
			(strings.Contains(line, "source") || strings.Contains(line, "activate") || strings.Contains(line, "eval")) {
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

	output := strings.Join(newLines, "\n") + "\n"
	if err := os.WriteFile(configFile, []byte(output), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	m.formatter.Success(fmt.Sprintf("Removed %s logic from %s (%d lines removed)", marker, configFile, removedCount))
	return nil
}
