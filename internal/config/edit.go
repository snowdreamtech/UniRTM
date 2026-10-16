// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// ReadFileOrEmpty reads a file content or returns an empty string if the file doesn't exist.
func ReadFileOrEmpty(path string) (string, error) {
	data, err := OsReadFile(path)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// quoteTOMLValue formats a string value with double quotes following standard TOML convention.
func quoteTOMLValue(val string) string {
	return fmt.Sprintf("%q", val)
}

// extractTOMLKey extracts the key from a "key = value" TOML line.
func extractTOMLKey(line string) string {
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, "#") || trimmed == "" || strings.HasPrefix(trimmed, "[") {
		return ""
	}
	parts := strings.SplitN(trimmed, "=", 2)
	if len(parts) != 2 {
		return ""
	}
	k := strings.TrimSpace(parts[0])
	k = strings.Trim(k, `"'`)
	return k
}

// UpsertEnvVar adds or updates an environment variable entry in the TOML [env] section.
// Following Postel's Law and POSIX environment standards, environment variable keys are
// strictly normalized to UPPERCASE when written to TOML configuration.
// Existing entries with matching key (case-insensitively) are replaced and deduplicated.
func UpsertEnvVar(content, key, value string) string {
	upperKey := strings.ToUpper(strings.TrimSpace(key))
	lines := strings.Split(content, "\n")
	newEntry := fmt.Sprintf("%s = %s", quoteTOMLKey(upperKey), quoteTOMLValue(value))

	inEnv := false
	envStart := -1
	envEnd := -1
	var matchingIndices []int

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
			existingKey := extractTOMLKey(line)
			if strings.EqualFold(existingKey, upperKey) {
				matchingIndices = append(matchingIndices, i)
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

	if len(matchingIndices) > 0 {
		// Update first matching occurrence with uppercase key
		lines[matchingIndices[0]] = newEntry
		// Remove any duplicate occurrences (e.g. if both lowercase and uppercase existed previously)
		if len(matchingIndices) > 1 {
			toRemove := make(map[int]bool)
			for _, idx := range matchingIndices[1:] {
				toRemove[idx] = true
			}
			filtered := make([]string, 0, len(lines)-len(matchingIndices)+1)
			for i, l := range lines {
				if !toRemove[i] {
					filtered = append(filtered, l)
				}
			}
			lines = filtered
		}
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
// Key matching is case-insensitive, ensuring any existing casing variants are removed.
func UnsetEnvVar(content, key string) (string, bool) {
	matchKey := strings.TrimSpace(key)
	lines := strings.Split(content, "\n")

	inEnv := false
	var matchingIndices []int

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
			existingKey := extractTOMLKey(line)
			if strings.EqualFold(existingKey, matchKey) {
				matchingIndices = append(matchingIndices, i)
			}
		}
	}

	if len(matchingIndices) > 0 {
		toRemove := make(map[int]bool)
		for _, idx := range matchingIndices {
			toRemove[idx] = true
		}
		newLines := make([]string, 0, len(lines)-len(matchingIndices))
		for i, l := range lines {
			if !toRemove[i] {
				newLines = append(newLines, l)
			}
		}
		return strings.Join(newLines, "\n"), true
	}

	return content, false
}

func quoteTOMLKey(key string) string {
	for _, c := range key {
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			return fmt.Sprintf("%q", key) // Use double quotes for keys with special chars
		}
	}
	return key
}

// UpsertToolVersion updates or inserts a tool version in the [tools] section of a TOML file.
func UpsertToolVersion(content string, tool string, version string) string {
	lines := strings.Split(content, "\n")
	newEntry := fmt.Sprintf("%s = %s", quoteTOMLKey(tool), quoteTOMLValue(version))

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
			} else {
				parts := strings.SplitN(trimmed, "=", 2)
				if len(parts) == 2 {
					keyPart := strings.TrimSpace(parts[0])
					keyPart = strings.Trim(keyPart, `"'`)
					if keyPart == tool {
						toolLineIdx = i
					}
				}
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

// RemoveToolVersion removes a tool entry from the [tools] section of a TOML file,
// preserving all comments, indentation, multi-line strings, and formatting across the rest of the file.
func RemoveToolVersion(content string, tool string) (string, bool) {
	lines := strings.Split(content, "\n")
	inTools := false
	toolLineIdx := -1

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "[tools]" {
			inTools = true
			continue
		}
		if inTools {
			if strings.HasPrefix(trimmed, "[") && trimmed != "[tools]" {
				break
			}
			parts := strings.SplitN(trimmed, "=", 2)
			if len(parts) == 2 {
				keyPart := strings.TrimSpace(parts[0])
				keyPart = strings.Trim(keyPart, `"'`)
				if keyPart == tool {
					toolLineIdx = i
					break
				}
			}
		}
	}

	if toolLineIdx != -1 {
		lines = append(lines[:toolLineIdx], lines[toolLineIdx+1:]...)
		return strings.Join(lines, "\n"), true
	}

	return content, false
}

// LoadRawTOML reads a TOML file into a generic map.
// Returns an empty map when the file does not yet exist.
func LoadRawTOML(path string) (map[string]interface{}, error) {
	data, err := OsReadFile(path)
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

// ValidateTOML verifies that the content parses as valid TOML syntax.
func ValidateTOML(content string) error {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return nil
	}
	var dummy map[string]interface{}
	if err := toml.Unmarshal([]byte(content), &dummy); err != nil {
		return fmt.Errorf("invalid TOML syntax: %w", err)
	}
	return nil
}

// WriteTOMLFile validates that content is syntactically valid TOML and writes it to path.
// If the content fails ValidateTOML, the write is aborted to protect config file integrity.
func WriteTOMLFile(path string, content []byte) error {
	if err := ValidateTOML(string(content)); err != nil {
		return fmt.Errorf("cannot write invalid TOML to %s: %w", path, err)
	}
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		_ = os.MkdirAll(dir, 0755)
	}
	return OsWriteFile(path, content, 0o644)
}

// SaveRawTOML writes a generic map to a TOML file safely without global reformatting.
func SaveRawTOML(path string, m map[string]interface{}) error {
	if envMap, ok := m["env"].(map[string]interface{}); ok {
		normalizedEnv := make(map[string]interface{}, len(envMap))
		for k, v := range envMap {
			normalizedEnv[strings.ToUpper(k)] = v
		}
		m["env"] = normalizedEnv
	}
	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	if err := enc.Encode(m); err != nil {
		return fmt.Errorf("encode TOML: %w", err)
	}
	return WriteTOMLFile(path, buf.Bytes())
}

// formatTOMLValue formats a value as a valid TOML literal.
func formatTOMLValue(val interface{}) string {
	switch v := val.(type) {
	case string:
		return quoteTOMLValue(v)
	case bool:
		if v {
			return "true"
		}
		return "false"
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return fmt.Sprintf("%v", v)
	default:
		data, err := toml.Marshal(v)
		if err == nil {
			return strings.TrimSpace(string(data))
		}
		return fmt.Sprintf("%v", v)
	}
}

// UpsertTOMLSetting updates or inserts a setting at keySegments in TOML content while
// strictly preserving all comments, formatting, and unknown keys.
func UpsertTOMLSetting(content string, keySegments []string, value interface{}) (string, error) {
	if len(keySegments) == 0 {
		return content, fmt.Errorf("key segments cannot be empty")
	}

	var sectionSegments []string
	var targetKey string
	if len(keySegments) == 1 {
		targetKey = keySegments[0]
	} else {
		sectionSegments = keySegments[:len(keySegments)-1]
		targetKey = keySegments[len(keySegments)-1]
	}

	isEnv := len(sectionSegments) == 1 && strings.EqualFold(sectionSegments[0], "env")
	if isEnv {
		targetKey = strings.ToUpper(targetKey)
	}

	// Format section header
	var sectionHeader string
	if len(sectionSegments) > 0 {
		var quotedSecs []string
		for _, s := range sectionSegments {
			quotedSecs = append(quotedSecs, quoteTOMLKey(s))
		}
		sectionHeader = "[" + strings.Join(quotedSecs, ".") + "]"
	}

	newEntry := fmt.Sprintf("%s = %s", quoteTOMLKey(targetKey), formatTOMLValue(value))

	lines := strings.Split(content, "\n")

	// If root section (no header)
	if sectionHeader == "" {
		keyLineIdx := -1
		firstSectionIdx := -1
		for i, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "[") {
				firstSectionIdx = i
				break
			}
			parts := strings.SplitN(trimmed, "=", 2)
			if len(parts) == 2 {
				k := strings.Trim(strings.TrimSpace(parts[0]), `"'`)
				if k == targetKey {
					keyLineIdx = i
					break
				}
			}
		}
		if keyLineIdx != -1 {
			lines[keyLineIdx] = newEntry
			return strings.Join(lines, "\n"), nil
		}
		if firstSectionIdx != -1 {
			newLines := make([]string, 0, len(lines)+1)
			newLines = append(newLines, lines[:firstSectionIdx]...)
			newLines = append(newLines, newEntry)
			newLines = append(newLines, lines[firstSectionIdx:]...)
			return strings.Join(newLines, "\n"), nil
		}
		lines = append(lines, newEntry)
		return strings.Join(lines, "\n"), nil
	}

	// Non-root section
	inSection := false
	secStart := -1
	secEnd := -1
	keyLineIdx := -1

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") && !strings.HasPrefix(trimmed, "[[") {
			if normalizeHeader(trimmed) == normalizeHeader(sectionHeader) {
				inSection = true
				secStart = i
				continue
			} else if inSection {
				secEnd = i
				inSection = false
				break
			}
		}
		if inSection {
			parts := strings.SplitN(trimmed, "=", 2)
			if len(parts) == 2 {
				k := strings.Trim(strings.TrimSpace(parts[0]), `"'`)
				if (isEnv && strings.EqualFold(k, targetKey)) || k == targetKey {
					keyLineIdx = i
				}
			}
		}
	}

	if inSection {
		secEnd = len(lines)
	}

	if secStart == -1 {
		// Section not found; append to end
		if content != "" && !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		content += "\n" + sectionHeader + "\n" + newEntry + "\n"
		return content, nil
	}

	if keyLineIdx != -1 {
		// Update existing line
		lines[keyLineIdx] = newEntry
		return strings.Join(lines, "\n"), nil
	}

	// Insert before secEnd
	insertAt := secEnd
	newLines := make([]string, 0, len(lines)+1)
	newLines = append(newLines, lines[:insertAt]...)
	newLines = append(newLines, newEntry)
	newLines = append(newLines, lines[insertAt:]...)
	return strings.Join(newLines, "\n"), nil
}

func normalizeHeader(h string) string {
	h = strings.TrimSpace(h)
	h = strings.TrimPrefix(h, "[")
	h = strings.TrimSuffix(h, "]")
	h = strings.TrimSpace(h)
	segments := SplitConfigKey(h)
	var norm []string
	for _, s := range segments {
		norm = append(norm, quoteTOMLKey(s))
	}
	return "[" + strings.Join(norm, ".") + "]"
}

// UpsertPin adds or updates a tool pin entry in the TOML [pins] section.
func UpsertPin(content, tool, reason string) string {
	lines := strings.Split(content, "\n")
	valToSet := quoteTOMLValue(reason)
	if reason == "" {
		valToSet = "true"
	}
	newEntry := fmt.Sprintf("%s = %s", quoteTOMLKey(tool), valToSet)

	inPins := false
	pinsStart := -1
	pinsEnd := -1
	matchingIdx := -1

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "[pins]" {
			inPins = true
			pinsStart = i
			continue
		}
		if inPins {
			if strings.HasPrefix(trimmed, "[") && trimmed != "[pins]" {
				pinsEnd = i
				inPins = false
				break
			}
			existingKey := extractTOMLKey(line)
			if existingKey == tool {
				matchingIdx = i
				break
			}
		}
	}
	if inPins {
		pinsEnd = len(lines)
	}

	if pinsStart == -1 {
		if content != "" && !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		content += "\n[pins]\n" + newEntry + "\n"
		return content
	}

	if matchingIdx != -1 {
		lines[matchingIdx] = newEntry
		return strings.Join(lines, "\n")
	}

	insertAt := pinsEnd
	newLines := make([]string, 0, len(lines)+1)
	newLines = append(newLines, lines[:insertAt]...)
	newLines = append(newLines, newEntry)
	newLines = append(newLines, lines[insertAt:]...)
	return strings.Join(newLines, "\n")
}

// RemovePin removes a tool pin entry from the [pins] section of a TOML file.
func RemovePin(content string, tool string) (string, bool) {
	lines := strings.Split(content, "\n")
	inPins := false
	pinLineIdx := -1

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "[pins]" {
			inPins = true
			continue
		}
		if inPins {
			if strings.HasPrefix(trimmed, "[") && trimmed != "[pins]" {
				break
			}
			existingKey := extractTOMLKey(line)
			if existingKey == tool {
				pinLineIdx = i
				break
			}
		}
	}

	if pinLineIdx != -1 {
		lines = append(lines[:pinLineIdx], lines[pinLineIdx+1:]...)
		return strings.Join(lines, "\n"), true
	}

	return content, false
}

