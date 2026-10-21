// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package guard

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// LoadGuardConfig discovers and unmarshals the [guard] section from workspace TOML configurations
func LoadGuardConfig(workspaceDir string) (GuardConfig, error) {
	defaultCfg := GuardConfig{
		Enabled: true,
		Mode:    "changed",
	}

	if workspaceDir == "" {
		return defaultCfg, nil
	}

	candidates := []string{
		filepath.Join(workspaceDir, ".unirtm.toml"),
		filepath.Join(workspaceDir, "unirtm.toml"),
		filepath.Join(workspaceDir, ".unirtm.local.toml"),
	}

	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var raw struct {
			Guard GuardConfig `toml:"guard"`
		}

		if err := toml.Unmarshal(data, &raw); err != nil {
			return defaultCfg, err
		}

		// Ensure defaults if omitted in config
		cfg := raw.Guard
		if cfg.Mode == "" {
			cfg.Mode = "changed"
		}
		return cfg, nil
	}

	return defaultCfg, nil
}

// MatchSlotHeader checks if a line is a [guard.slots."<id>"] or [guard.slots.<id>] table header
func MatchSlotHeader(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "[") || !strings.HasSuffix(trimmed, "]") {
		return "", false
	}
	content := strings.TrimSpace(trimmed[1 : len(trimmed)-1])
	prefix := "guard.slots."
	if !strings.HasPrefix(content, prefix) {
		return "", false
	}
	rawID := strings.TrimPrefix(content, prefix)
	rawID = strings.Trim(rawID, `"'`)
	return rawID, true
}

// formatSlotBlock formats a SlotConfig into canonical TOML table lines
func formatSlotBlock(slot SlotConfig) []string {
	var lines []string
	header := fmt.Sprintf("[guard.slots.%q]", slot.ID)
	lines = append(lines, header)

	if slot.Tool != "" {
		lines = append(lines, fmt.Sprintf("tool = %q", slot.Tool))
	}
	if slot.Command != "" {
		lines = append(lines, fmt.Sprintf("command = %q", slot.Command))
	}
	if slot.FixCommand != "" {
		lines = append(lines, fmt.Sprintf("fix_command = %q", slot.FixCommand))
	}
	if slot.Category != "" {
		lines = append(lines, fmt.Sprintf("category = %q", string(slot.Category)))
	}
	if slot.Phase != "" {
		lines = append(lines, fmt.Sprintf("phase = %q", string(slot.Phase)))
	}
	if len(slot.Extensions) > 0 {
		var quotedExts []string
		for _, ext := range slot.Extensions {
			quotedExts = append(quotedExts, fmt.Sprintf("%q", ext))
		}
		lines = append(lines, fmt.Sprintf("extensions = [%s]", strings.Join(quotedExts, ", ")))
	}
	if slot.Enabled != nil {
		lines = append(lines, fmt.Sprintf("enabled = %t", *slot.Enabled))
	}
	return lines
}

// UpsertGuardSlot adds or updates a [guard.slots."<id>"] table in TOML content preserving file structure
func UpsertGuardSlot(content string, slot SlotConfig) string {
	lines := strings.Split(content, "\n")
	slotStart := -1
	slotEnd := -1

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		id, ok := MatchSlotHeader(line)
		if ok {
			if slotStart != -1 {
				slotEnd = i
				break
			}
			if id == slot.ID {
				slotStart = i
				continue
			}
		} else if slotStart != -1 && strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			slotEnd = i
			break
		}
	}

	newBlock := formatSlotBlock(slot)

	if slotStart != -1 {
		if slotEnd == -1 {
			slotEnd = len(lines)
		}
		newLines := make([]string, 0, len(lines)-(slotEnd-slotStart)+len(newBlock))
		newLines = append(newLines, lines[:slotStart]...)
		newLines = append(newLines, newBlock...)
		newLines = append(newLines, lines[slotEnd:]...)
		return strings.Join(newLines, "\n")
	}

	// Not found, append at the end of the file
	if len(content) > 0 && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	return content + "\n" + strings.Join(newBlock, "\n") + "\n"
}

// RemoveGuardSlot removes a [guard.slots."<id>"] table from TOML content
func RemoveGuardSlot(content string, slotID string) (string, bool) {
	lines := strings.Split(content, "\n")
	slotStart := -1
	slotEnd := -1

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		id, ok := MatchSlotHeader(line)
		if ok {
			if slotStart != -1 {
				slotEnd = i
				break
			}
			if id == slotID {
				slotStart = i
				continue
			}
		} else if slotStart != -1 && strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			slotEnd = i
			break
		}
	}

	if slotStart != -1 {
		if slotEnd == -1 {
			slotEnd = len(lines)
		}
		newLines := make([]string, 0, len(lines)-(slotEnd-slotStart))
		newLines = append(newLines, lines[:slotStart]...)
		newLines = append(newLines, lines[slotEnd:]...)
		return strings.Join(newLines, "\n"), true
	}

	return content, false
}

// resolveWorkspaceConfigPath locates or creates the standard .unirtm.toml path
func resolveWorkspaceConfigPath(workspaceDir string) string {
	candidates := []string{
		filepath.Join(workspaceDir, ".unirtm.toml"),
		filepath.Join(workspaceDir, "unirtm.toml"),
	}
	for _, p := range candidates {
		if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
			return p
		}
	}
	return filepath.Join(workspaceDir, ".unirtm.toml")
}

// SaveSlotConfig writes or updates a slot definition in the workspace .unirtm.toml
func SaveSlotConfig(workspaceDir string, slot SlotConfig) error {
	path := resolveWorkspaceConfigPath(workspaceDir)
	var content string
	if data, err := os.ReadFile(path); err == nil {
		content = string(data)
	}

	updated := UpsertGuardSlot(content, slot)

	// Validate resulting TOML syntax
	var check map[string]interface{}
	if err := toml.Unmarshal([]byte(updated), &check); err != nil {
		return fmt.Errorf("generated invalid TOML syntax: %w", err)
	}

	return os.WriteFile(path, []byte(updated), 0644)
}

// DeleteSlotConfig deletes a custom slot override from workspace .unirtm.toml
func DeleteSlotConfig(workspaceDir string, slotID string) error {
	path := resolveWorkspaceConfigPath(workspaceDir)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	updated, changed := RemoveGuardSlot(string(data), slotID)
	if !changed {
		return nil
	}

	// Validate resulting TOML syntax
	var check map[string]interface{}
	if err := toml.Unmarshal([]byte(updated), &check); err != nil {
		return fmt.Errorf("resulting TOML has syntax error: %w", err)
	}

	return os.WriteFile(path, []byte(updated), 0644)
}

// GetGuardToml reads the workspace .unirtm.toml content
func GetGuardToml(workspaceDir string) (string, error) {
	path := resolveWorkspaceConfigPath(workspaceDir)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

// SaveGuardToml validates and writes raw TOML content to the workspace .unirtm.toml
func SaveGuardToml(workspaceDir string, content string) error {
	var check map[string]interface{}
	if err := toml.Unmarshal([]byte(content), &check); err != nil {
		return fmt.Errorf("invalid TOML syntax: %w", err)
	}
	path := resolveWorkspaceConfigPath(workspaceDir)
	return os.WriteFile(path, []byte(content), 0644)
}
