// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package guard

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// DebugStatementsRule detects leftover debugging statements such as debugger, console.debug, pdb.set_trace()
type DebugStatementsRule struct {
	patterns []*regexp.Regexp
}

func NewDebugStatementsRule() Rule {
	return &DebugStatementsRule{
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`\bdebugger\s*;?`),
			regexp.MustCompile(`\bconsole\.(debug|trace)\s*\(`),
			regexp.MustCompile(`\bbreakpoint\s*\(`),
			regexp.MustCompile(`\bpdb\.set_trace\s*\(`),
			regexp.MustCompile(`\bbinding\.pry\b`),
			regexp.MustCompile(`\bvar_dump\s*\(`),
		},
	}
}

func (r *DebugStatementsRule) ID() string {
	return "builtin:forbid-debug-statements"
}

func (r *DebugStatementsRule) Category() Category {
	return CategoryHygiene
}

func (r *DebugStatementsRule) Phase() Phase {
	return PhaseLinter
}

func (r *DebugStatementsRule) DefaultExtensions() []string {
	return []string{".js", ".ts", ".jsx", ".tsx", ".vue", ".py", ".rb", ".php"}
}

func (r *DebugStatementsRule) Execute(ctx context.Context, files []string, opts ExecuteOptions) ([]Issue, error) {
	var issues []Issue

	extMap := make(map[string]bool)
	for _, ext := range r.DefaultExtensions() {
		extMap[strings.ToLower(ext)] = true
	}

	for _, rel := range files {
		select {
		case <-ctx.Done():
			return issues, ctx.Err()
		default:
		}

		ext := strings.ToLower(filepath.Ext(rel))
		if !extMap[ext] {
			continue
		}

		// Skip test files, mocks, vendor
		norm := filepath.ToSlash(rel)
		if strings.Contains(norm, "/tests/") || strings.Contains(norm, "/test/") ||
			strings.Contains(norm, ".test.") || strings.Contains(norm, ".spec.") ||
			strings.Contains(norm, "/mock/") || strings.Contains(norm, "/mocks/") ||
			strings.HasSuffix(norm, "/logger.ts") || strings.HasSuffix(norm, "/logger.js") {
			continue
		}

		fullPath := rel
		if !filepath.IsAbs(fullPath) {
			fullPath = filepath.Join(opts.WorkspaceDir, rel)
		}

		fi, err := os.Lstat(fullPath)
		if err != nil || fi.IsDir() || fi.Mode()&os.ModeSymlink != 0 {
			continue
		}

		if fi.Size() > 2*1024*1024 {
			// Skip files > 2MB
			continue
		}

		content, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(bytes.NewReader(content))
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/*") || strings.HasPrefix(line, "*") || strings.HasPrefix(line, "#") {
				continue
			}

			for _, pattern := range r.patterns {
				if match := pattern.FindString(line); match != "" {
					issues = append(issues, Issue{
						FilePath: rel,
						Line:     lineNum,
						RuleID:   r.ID(),
						Category: r.Category(),
						Phase:    r.Phase(),
						Severity: SeverityWarning,
						Message:  fmt.Sprintf("Found leftover debug statement '%s' on line %d", match, lineNum),
						Fixable:  false,
					})
					break
				}
			}
		}
	}

	return issues, nil
}
