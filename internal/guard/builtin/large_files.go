// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package builtin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/snowdreamtech/unirtm/internal/guard"
)

type LargeFilesRule struct{}

func NewLargeFilesRule() guard.Rule {
	return &LargeFilesRule{}
}

func (r *LargeFilesRule) ID() string {
	return "builtin:check-added-large-files"
}

func (r *LargeFilesRule) Category() guard.Category {
	return guard.CategoryHygiene
}

func (r *LargeFilesRule) Phase() guard.Phase {
	return guard.PhaseLinter
}

func (r *LargeFilesRule) DefaultExtensions() []string {
	return nil
}

func (r *LargeFilesRule) Execute(ctx context.Context, files []string, opts guard.ExecuteOptions) ([]guard.Issue, error) {
	var issues []guard.Issue
	limit := opts.MaxFileSize
	if limit <= 0 {
		limit = 1024 * 1024 // 1MB default
	}

	for _, rel := range files {
		select {
		case <-ctx.Done():
			return issues, ctx.Err()
		default:
		}

		fullPath := filepath.Join(opts.WorkspaceDir, rel)
		fi, err := os.Lstat(fullPath)
		if err != nil || fi.Mode()&os.ModeSymlink != 0 {
			continue
		}

		if fi.Size() > limit {
			issues = append(issues, guard.Issue{
				FilePath: rel,
				Line:     1,
				RuleID:   r.ID(),
				Category: r.Category(),
				Phase:    r.Phase(),
				Severity: guard.SeverityError,
				Message: fmt.Sprintf("File size (%.2f MB) exceeds threshold (%.2f MB). Use Git LFS or .gitignore.",
					float64(fi.Size())/(1024*1024), float64(limit)/(1024*1024)),
				Fixable: false,
			})
		}
	}

	return issues, nil
}
