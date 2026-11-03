// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package guard

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// forbiddenPattern matches common temp, editor backup, OS cache, and debug remnant files
var forbiddenPattern = regexp.MustCompile(`(?i)(^|/)(\.DS_Store|Thumbs\.db|npm-debug\.log|yarn-error\.log|pnpm-debug\.log|debug\.log|tmp\.json|test\.txt)$|\.(swp|swo|bak|orig|rej|coverage)$|~$`)

type ForbiddenFilesRule struct{}

func NewForbiddenFilesRule() Rule {
	return &ForbiddenFilesRule{}
}

func (r *ForbiddenFilesRule) ID() string {
	return "builtin:forbidden-files"
}

func (r *ForbiddenFilesRule) Category() Category {
	return CategoryHygiene
}

func (r *ForbiddenFilesRule) Phase() Phase {
	return PhaseLinter
}

func (r *ForbiddenFilesRule) DefaultExtensions() []string {
	return nil
}

func (r *ForbiddenFilesRule) Execute(ctx context.Context, files []string, opts ExecuteOptions) ([]Issue, error) {
	var issues []Issue

	for _, rel := range files {
		select {
		case <-ctx.Done():
			return issues, ctx.Err()
		default:
		}

		fullPath := rel
		if !filepath.IsAbs(fullPath) {
			fullPath = filepath.Join(opts.WorkspaceDir, rel)
		}

		// Check if file exists
		if _, err := os.Lstat(fullPath); err != nil {
			continue
		}

		cleanRel := filepath.ToSlash(filepath.Clean(rel))
		baseName := filepath.Base(cleanRel)

		if forbiddenPattern.MatchString(cleanRel) || forbiddenPattern.MatchString(baseName) {
			issues = append(issues, Issue{
				FilePath: rel,
				Line:     1,
				RuleID:   r.ID(),
				Category: r.Category(),
				Phase:    r.Phase(),
				Severity: SeverityError,
				Message:  fmt.Sprintf("Forbidden temporary or remnant file detected: %s", baseName),
				Fixable:  false,
			})
		} else if strings.HasSuffix(baseName, ".map") && (strings.HasPrefix(cleanRel, "src/") || strings.HasPrefix(cleanRel, "cmd/")) {
			// Source maps in source directories
			issues = append(issues, Issue{
				FilePath: rel,
				Line:     1,
				RuleID:   r.ID(),
				Category: r.Category(),
				Phase:    r.Phase(),
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("Source map file (%s) should not be committed into source directories", baseName),
				Fixable:  false,
			})
		}
	}

	return issues, nil
}
