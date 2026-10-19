// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package guard

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

)

type CaseConflictRule struct{}

func NewCaseConflictRule() Rule {
	return &CaseConflictRule{}
}

func (r *CaseConflictRule) ID() string {
	return "builtin:check-case-conflict"
}

func (r *CaseConflictRule) Category() Category {
	return CategoryHygiene
}

func (r *CaseConflictRule) Phase() Phase {
	return PhaseLinter
}

func (r *CaseConflictRule) DefaultExtensions() []string {
	return nil
}

func (r *CaseConflictRule) Execute(ctx context.Context, files []string, opts ExecuteOptions) ([]Issue, error) {
	var issues []Issue
	// Group files by lowercase path
	seen := make(map[string]string) // lowerPath -> originalPath

	for _, rel := range files {
		select {
		case <-ctx.Done():
			return issues, ctx.Err()
		default:
		}

		cleanPath := filepath.Clean(rel)
		lower := strings.ToLower(cleanPath)

		if existing, exists := seen[lower]; exists && existing != cleanPath {
			issues = append(issues, Issue{
				FilePath: cleanPath,
				Line:     1,
				RuleID:   r.ID(),
				Category: r.Category(),
				Phase:    r.Phase(),
				Severity: SeverityError,
				Message: fmt.Sprintf("Case-insensitive filename conflict detected between %q and %q (breaks macOS/Windows checkout)",
					cleanPath, existing),
				Fixable: false,
			})
		} else {
			seen[lower] = cleanPath
		}
	}

	return issues, nil
}
