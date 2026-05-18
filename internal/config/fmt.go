// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
)

// FormatFile formats the configuration file at the given path with canonical ordering.
// It returns true if the file content was modified, and an error if one occurred.
// If fmtCheck is true, it does not write changes back to the disk.
func FormatFile(path string, fmtCheck bool) (bool, error) {
	original, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}

	var formatted []byte
	if strings.HasSuffix(path, ".toml") {
		result, err := FormatTOML(string(original))
		if err != nil {
			return false, fmt.Errorf("format TOML: %w", err)
		}
		formatted = []byte(result)

		// Optionally run taplo if available, just like mise does
		if _, err := exec.LookPath("taplo"); err == nil {
			cmd := exec.Command("taplo", "format", "-")
			cmd.Stdin = bytes.NewReader(formatted)
			var out bytes.Buffer
			cmd.Stdout = &out
			if err := cmd.Run(); err == nil {
				formatted = out.Bytes()
			}
		}
	} else {
		formatted = []byte(strings.TrimSpace(string(original)) + "\n")
	}

	if bytes.Equal(original, formatted) {
		return false, nil
	}

	if !fmtCheck {
		if err := os.WriteFile(path, formatted, 0o644); err != nil {
			return false, err
		}
	}

	return true, nil
}

type tomlBlock struct {
	TopKey string
	Lines  []string
}

func getSectionOrder(key string) int {
	switch key {
	case "min_version":
		return 0
	case "env_file":
		return 1
	case "env_path":
		return 2
	case "":
		return 3
	case "env":
		return 4
	case "vars":
		return 5
	case "hooks":
		return 6
	case "watch_files":
		return 7
	case "tools":
		return 8
	case "tasks":
		return 10
	case "task_config":
		return 11
	case "redactions":
		return 12
	case "alias":
		return 13
	case "plugins":
		return 14
	case "settings":
		return 15
	default:
		return 9
	}
}

// FormatTOML formats TOML content strings with canonical key ordering.
func FormatTOML(content string) (string, error) {
	lines := strings.Split(content, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	blocksMap := make(map[string]*tomlBlock)
	var blocksList []*tomlBlock

	getBlock := func(key string) *tomlBlock {
		if b, exists := blocksMap[key]; exists {
			return b
		}
		b := &tomlBlock{TopKey: key}
		blocksMap[key] = b
		blocksList = append(blocksList, b)
		return b
	}

	currentTopKey := ""
	var pendingComments []string
	inMultiLineSingle := false
	inMultiLineDouble := false

	headerRegex := regexp.MustCompile(`^\[\[?\s*"?([a-zA-Z0-9_-]+)"?`)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		isCommentOrBlank := trimmed == "" || strings.HasPrefix(trimmed, "#")
		inMultiLine := inMultiLineSingle || inMultiLineDouble

		if !inMultiLine && !isCommentOrBlank && strings.HasPrefix(trimmed, "[") {
			matches := headerRegex.FindStringSubmatch(trimmed)
			if len(matches) > 1 {
				// If this is the very first section and we have pending comments,
				// they are likely file-level comments. Flush them to the root block instead.
				if currentTopKey == "" && len(pendingComments) > 0 {
					rootBlock := getBlock("")
					rootBlock.Lines = append(rootBlock.Lines, pendingComments...)
					pendingComments = nil
				}

				currentTopKey = matches[1]
				b := getBlock(currentTopKey)
				b.Lines = append(b.Lines, pendingComments...)
				pendingComments = nil
				b.Lines = append(b.Lines, line)
				continue
			}
		}

		if !inMultiLine && isCommentOrBlank {
			pendingComments = append(pendingComments, line)
			continue
		}

		b := getBlock(currentTopKey)
		b.Lines = append(b.Lines, pendingComments...)
		pendingComments = nil
		b.Lines = append(b.Lines, line)

		if !inMultiLineSingle {
			doubleCount := strings.Count(line, `"""`) - strings.Count(line, `\"""`)
			if doubleCount%2 != 0 {
				inMultiLineDouble = !inMultiLineDouble
			}
		}
		if !inMultiLineDouble {
			singleCount := strings.Count(line, `'''`)
			if singleCount%2 != 0 {
				inMultiLineSingle = !inMultiLineSingle
			}
		}
	}

	b := getBlock(currentTopKey)
	b.Lines = append(b.Lines, pendingComments...)

	sort.SliceStable(blocksList, func(i, j int) bool {
		o1 := getSectionOrder(blocksList[i].TopKey)
		o2 := getSectionOrder(blocksList[j].TopKey)
		if o1 == o2 {
			return blocksList[i].TopKey < blocksList[j].TopKey
		}
		return o1 < o2
	})

	var out []string
	for _, b := range blocksList {
		out = append(out, b.Lines...)
	}

	return strings.Join(out, "\n") + "\n", nil
}
