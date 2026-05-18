// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// ReadFileOrEmpty reads a file content or returns an empty string if the file doesn't exist.
func ReadFileOrEmpty(path string) (string, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// UpsertEnvVar adds or updates an environment variable entry in the TOML [env] section.
func UpsertEnvVar(content, key, value string) string {
	lines := strings.Split(content, "\n")
	newEntry := fmt.Sprintf("%s = %q", key, value)

	inEnv := false
	envStart := -1
	envEnd := -1
	envLineIdx := -1

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "[env]" {
			inEnv = true
			envStart = i
			continue
		}
		if inEnv {
			if strings.HasPrefix(trimmed, "[") && trimmed != "[env]" {
				envEnd = i
				inEnv = false
				break
			}
			if strings.HasPrefix(trimmed, key+"=") || strings.HasPrefix(trimmed, key+" =") {
				envLineIdx = i
			}
		}
	}
	if inEnv {
		envEnd = len(lines)
	}

	if envStart == -1 {
		if content != "" && !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		content += "\n[env]\n" + newEntry + "\n"
		return content
	}

	if envLineIdx != -1 {
		lines[envLineIdx] = newEntry
		return strings.Join(lines, "\n")
	}

	insertAt := envEnd
	newLines := make([]string, 0, len(lines)+1)
	newLines = append(newLines, lines[:insertAt]...)
	newLines = append(newLines, newEntry)
	newLines = append(newLines, lines[insertAt:]...)
	return strings.Join(newLines, "\n")
}

// UnsetEnvVar removes an environment variable entry from the TOML [env] section.
func UnsetEnvVar(content, key string) (string, bool) {
	lines := strings.Split(content, "\n")

	inEnv := false
	envLineIdx := -1

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "[env]" {
			inEnv = true
			continue
		}
		if inEnv {
			if strings.HasPrefix(trimmed, "[") && trimmed != "[env]" {
				break
			}
			if strings.HasPrefix(trimmed, key+"=") || strings.HasPrefix(trimmed, key+" =") {
				envLineIdx = i
				break
			}
		}
	}

	if envLineIdx != -1 {
		// Remove the line
		lines = append(lines[:envLineIdx], lines[envLineIdx+1:]...)
		return strings.Join(lines, "\n"), true
	}

	return content, false
}

// UpsertToolVersion adds or updates a tool version entry in the TOML [tools] section.
func UpsertToolVersion(content, tool, version string) string {
	lines := strings.Split(content, "\n")
	newEntry := fmt.Sprintf("%s = %q", tool, version)

	// Look for [tools] section
	inTools := false
	toolsStart := -1
	toolsEnd := -1
	toolLineIdx := -1

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "[tools]" {
			inTools = true
			toolsStart = i
			continue
		}
		if inTools {
			if strings.HasPrefix(trimmed, "[") && trimmed != "[tools]" {
				// New section started
				toolsEnd = i
				inTools = false
				break
			}
			if strings.HasPrefix(trimmed, tool+"=") || strings.HasPrefix(trimmed, tool+" =") {
				toolLineIdx = i
			}
		}
	}
	if inTools {
		toolsEnd = len(lines)
	}

	if toolsStart == -1 {
		// No [tools] section — append it
		if content != "" && !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		content += "\n[tools]\n" + newEntry + "\n"
		return content
	}

	if toolLineIdx != -1 {
		// Update existing line
		lines[toolLineIdx] = newEntry
		return strings.Join(lines, "\n")
	}

	// Insert before toolsEnd
	insertAt := toolsEnd
	newLines := make([]string, 0, len(lines)+1)
	newLines = append(newLines, lines[:insertAt]...)
	newLines = append(newLines, newEntry)
	newLines = append(newLines, lines[insertAt:]...)
	return strings.Join(newLines, "\n")
}

// LoadRawTOML reads a TOML file into a generic map.
// Returns an empty map when the file does not yet exist.
func LoadRawTOML(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return make(map[string]interface{}), nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var m map[string]interface{}
	if err := toml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if m == nil {
		m = make(map[string]interface{})
	}
	return m, nil
}

// SaveRawTOML writes a generic map to a TOML file and formats it.
func SaveRawTOML(path string, m map[string]interface{}) error {
	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	if err := enc.Encode(m); err != nil {
		return fmt.Errorf("encode TOML: %w", err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		return err
	}
	// Always format the file to enforce standard section ordering
	_, _ = FormatFile(path, false)
	return nil
}
