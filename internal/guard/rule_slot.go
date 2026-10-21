// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package guard

import (
	"bytes"
	"context"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// SlotRule adapts an external ToolSlot into the standard Guard Rule interface
type SlotRule struct {
	slot ToolSlot
}

// NewSlotRule constructs a Rule backed by a ToolSlot definition
func NewSlotRule(slot ToolSlot) *SlotRule {
	return &SlotRule{slot: slot}
}

// ID returns the unique identifier for the slot rule (prefixed with slot:)
func (s *SlotRule) ID() string {
	return "slot:" + s.slot.ID
}

// Category returns the functional pillar of this slot
func (s *SlotRule) Category() Category {
	return s.slot.Category
}

// Phase returns the pipeline phase (fixer, formatter, or linter)
func (s *SlotRule) Phase() Phase {
	return s.slot.Phase
}

// DefaultExtensions returns target file extensions
func (s *SlotRule) DefaultExtensions() []string {
	return s.slot.DefaultExtensions
}

// FilterMatchingFiles filters candidate files by matching slot extensions
func (s *SlotRule) FilterMatchingFiles(files []string) []string {
	if len(s.slot.DefaultExtensions) == 0 {
		return files
	}
	var matched []string
	for _, f := range files {
		ext := strings.ToLower(filepath.Ext(f))
		for _, defExt := range s.slot.DefaultExtensions {
			if ext == defExt {
				matched = append(matched, f)
				break
			}
		}
	}
	return matched
}

// ShouldRun determines whether this slot rule applies to candidate files and current environment
func (s *SlotRule) ShouldRun(files []string) bool {
	if len(s.slot.DefaultExtensions) > 0 {
		if len(s.FilterMatchingFiles(files)) == 0 {
			return false
		}
	}
	if s.slot.DefaultToolName != "" {
		if _, err := exec.LookPath(s.slot.DefaultToolName); err != nil {
			return false
		}
	}
	return true
}

// Execute triggers the external tool associated with the slot
func (s *SlotRule) Execute(ctx context.Context, files []string, opts ExecuteOptions) ([]Issue, error) {
	matchingFiles := s.FilterMatchingFiles(files)
	if len(s.slot.DefaultExtensions) > 0 && len(matchingFiles) == 0 {
		return nil, nil
	}
	if len(s.slot.DefaultExtensions) == 0 {
		matchingFiles = files
	}

	toolName := s.slot.DefaultToolName
	if toolName == "" {
		return nil, nil
	}

	toolBin, err := exec.LookPath(toolName)
	if err != nil {
		return nil, nil
	}

	cmdStr := s.slot.DefaultCommand
	if opts.Fix && s.slot.FixCommand != "" {
		cmdStr = s.slot.FixCommand
	}
	if cmdStr == "" {
		cmdStr = toolName
	}

	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		parts = []string{toolName}
	}

	var cmdArgs []string
	hasPlaceholder := false

	for _, p := range parts[1:] {
		if p == "{{files}}" {
			cmdArgs = append(cmdArgs, matchingFiles...)
			hasPlaceholder = true
		} else if strings.Contains(p, "{{files}}") {
			replaced := strings.ReplaceAll(p, "{{files}}", strings.Join(matchingFiles, " "))
			cmdArgs = append(cmdArgs, replaced)
			hasPlaceholder = true
		} else {
			cmdArgs = append(cmdArgs, p)
		}
	}

	if !hasPlaceholder && len(matchingFiles) > 0 && !strings.Contains(s.slot.ID, "linter") && !strings.Contains(s.slot.ID, "secret") {
		cmdArgs = append(cmdArgs, matchingFiles...)
	}

	cmd := exec.CommandContext(ctx, toolBin, cmdArgs...)
	cmd.Stdin = bytes.NewReader(nil)
	if opts.WorkspaceDir != "" {
		cmd.Dir = opts.WorkspaceDir
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()

	// Special check for gofmt/goimports in listing mode without fix flag
	if runErr == nil && !opts.Fix && (toolName == "gofmt" || toolName == "goimports") {
		outLines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
		var issues []Issue
		for _, l := range outLines {
			fileLine := strings.TrimSpace(l)
			if fileLine == "" {
				continue
			}
			issues = append(issues, Issue{
				FilePath: fileLine,
				Line:     1,
				RuleID:   s.ID(),
				Category: s.Category(),
				Phase:    s.Phase(),
				Severity: SeverityWarning,
				Message:  "File formatting differs from canonical " + toolName + " style",
				Fixable:  s.slot.FixCommand != "",
			})
		}
		return issues, nil
	}

	if runErr == nil {
		return nil, nil
	}

	outputCombined := strings.TrimSpace(stderr.String())
	if outputCombined == "" {
		outputCombined = strings.TrimSpace(stdout.String())
	}
	if outputCombined == "" {
		outputCombined = runErr.Error()
	}

	locRegex := regexp.MustCompile(`(?m)^([a-zA-Z0-9_\-\./\\]+):(\d+)(?::(\d+))?:\s*(.+)$`)
	matches := locRegex.FindAllStringSubmatch(outputCombined, 50)

	var issues []Issue
	if len(matches) > 0 {
		for _, m := range matches {
			line, _ := strconv.Atoi(m[2])
			col := 0
			if len(m) > 3 && m[3] != "" {
				col, _ = strconv.Atoi(m[3])
			}
			msg := m[len(m)-1]
			issues = append(issues, Issue{
				FilePath: m[1],
				Line:     line,
				Column:   col,
				RuleID:   s.ID(),
				Category: s.Category(),
				Phase:    s.Phase(),
				Severity: SeverityError,
				Message:  msg,
				Fixable:  s.slot.FixCommand != "",
			})
		}
	} else {
		firstFile := "workspace"
		if len(matchingFiles) > 0 {
			firstFile = matchingFiles[0]
		}
		firstLine := strings.Split(outputCombined, "\n")[0]
		issues = append(issues, Issue{
			FilePath: firstFile,
			Line:     1,
			RuleID:   s.ID(),
			Category: s.Category(),
			Phase:    s.Phase(),
			Severity: SeverityError,
			Message:  firstLine,
			Fixable:  s.slot.FixCommand != "",
		})
	}

	return issues, nil
}
