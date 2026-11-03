// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"bufio"
	"bytes"
	"path/filepath"
	"strings"
	"unicode"
)

// LegacyFileInfo captures metadata about an individual legacy version file.
type LegacyFileInfo struct {
	FilePath string `json:"file_path"`
	FileName string `json:"file_name"`
	Tool     string `json:"tool"`
	Version  string `json:"version"`
}

// cleanVersionString normalizes version strings extracted from legacy files:
// - Removes comments (# ...)
// - Trims whitespace
// - Removes leading 'v' or 'V' for Node.js versions (e.g. v20.18.0 -> 20.18.0)
// - Removes leading 'go' for Go versions (e.g. go1.23.0 -> 1.23.0)
func cleanVersionString(val, tool string) string {
	val = strings.TrimSpace(val)
	if idx := strings.Index(val, "#"); idx != -1 {
		val = strings.TrimSpace(val[:idx])
	}
	if val == "" {
		return ""
	}

	// Normalise Node.js versions
	if tool == "node" {
		if (strings.HasPrefix(val, "v") || strings.HasPrefix(val, "V")) && len(val) > 1 && unicode.IsDigit(rune(val[1])) {
			val = val[1:]
		}
	}

	// Normalise Go versions
	if tool == "go" {
		if strings.HasPrefix(val, "go") && len(val) > 2 && unicode.IsDigit(rune(val[2])) {
			val = val[2:]
		}
	}

	return val
}

// readFirstLine reads the first non-empty, non-comment line from a file.
func readFirstLine(path, tool string) string {
	data, err := OsReadFile(path)
	if err != nil {
		return ""
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if cleaned := cleanVersionString(line, tool); cleaned != "" {
			return cleaned
		}
	}
	return ""
}

// parseToolVersions parses an asdf / mise style .tool-versions file.
// Format:
//
//	nodejs 20.18.0
//	golang 1.23.0
//	python 3.12.0
func parseToolVersions(path string) (ToolMap, map[string]string) {
	tools := make(ToolMap)
	sources := make(map[string]string)

	data, err := OsReadFile(path)
	if err != nil {
		return tools, sources
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if idx := strings.Index(line, "#"); idx != -1 {
			line = strings.TrimSpace(line[:idx])
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		rawTool := parts[0]
		rawVer := parts[1]

		// Map common aliases to canonical UniRTM tool names
		toolName := rawTool
		switch strings.ToLower(rawTool) {
		case "nodejs":
			toolName = "node"
		case "golang":
			toolName = "go"
		}

		cleanVer := cleanVersionString(rawVer, toolName)
		if cleanVer != "" {
			tools[toolName] = ToolConfig{Version: cleanVer}
			sources[toolName] = filepath.Base(path)
		}
	}

	return tools, sources
}

// ReadLegacyVersionFiles discovers and parses legacy version files in the specified directory:
//   - .nvmrc          -> node
//   - .node-version   -> node
//   - .go-version     -> go
//   - .python-version -> python
//   - .tool-versions  -> multiple tools (asdf/mise format)
//
// Returns a ToolMap of discovered tools and a mapping of toolName -> sourceFileName.
func ReadLegacyVersionFiles(dir string) (ToolMap, map[string]string) {
	tools := make(ToolMap)
	sources := make(map[string]string)

	// 1. Check Node.js legacy files: .nvmrc takes precedence over .node-version
	nvmrcPath := filepath.Join(dir, ".nvmrc")
	if ver := readFirstLine(nvmrcPath, "node"); ver != "" {
		tools["node"] = ToolConfig{Version: ver}
		sources["node"] = ".nvmrc"
	} else {
		nodeVerPath := filepath.Join(dir, ".node-version")
		if ver := readFirstLine(nodeVerPath, "node"); ver != "" {
			tools["node"] = ToolConfig{Version: ver}
			sources["node"] = ".node-version"
		}
	}

	// 2. Check Go legacy files: .go-version
	goVerPath := filepath.Join(dir, ".go-version")
	if ver := readFirstLine(goVerPath, "go"); ver != "" {
		tools["go"] = ToolConfig{Version: ver}
		sources["go"] = ".go-version"
	}

	// 3. Check Python legacy files: .python-version
	pyVerPath := filepath.Join(dir, ".python-version")
	if ver := readFirstLine(pyVerPath, "python"); ver != "" {
		tools["python"] = ToolConfig{Version: ver}
		sources["python"] = ".python-version"
	}

	// 4. Check .tool-versions (asdf format) - fills in any tools not already set by specific files
	toolVersionsPath := filepath.Join(dir, ".tool-versions")
	tvTools, tvSources := parseToolVersions(toolVersionsPath)
	for t, cfg := range tvTools {
		if _, exists := tools[t]; !exists {
			tools[t] = cfg
			sources[t] = tvSources[t]
		}
	}

	return tools, sources
}
