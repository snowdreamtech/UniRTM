// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package guard

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type LargeFilesRule struct{}

func NewLargeFilesRule() Rule {
	return &LargeFilesRule{}
}

func (r *LargeFilesRule) ID() string {
	return "builtin:check-added-large-files"
}

func (r *LargeFilesRule) Category() Category {
	return CategoryHygiene
}

func (r *LargeFilesRule) Phase() Phase {
	return PhaseLinter
}

func (r *LargeFilesRule) DefaultExtensions() []string {
	return nil
}

func (r *LargeFilesRule) Execute(ctx context.Context, files []string, opts ExecuteOptions) ([]Issue, error) {
	var issues []Issue
	limit := opts.MaxFileSize
	if limit <= 0 {
		limit = 10 * 1024 * 1024 // 10MB default
	}

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
		fi, err := os.Lstat(fullPath)
		if err != nil || fi.Mode()&os.ModeSymlink != 0 {
			continue
		}

		lowerRel := strings.ToLower(filepath.ToSlash(rel))
		if strings.Contains(lowerRel, "examples/plugin-go/unirtm-plugin-example") ||
			strings.Contains(lowerRel, "docs/tapes/demo.gif") {
			continue
		}

		if fi.Size() > limit {
			issues = append(issues, Issue{
				FilePath: rel,
				Line:     1,
				RuleID:   r.ID(),
				Category: r.Category(),
				Phase:    r.Phase(),
				Severity: SeverityError,
				Message: fmt.Sprintf("File size (%.2f MB) exceeds threshold (%.2f MB). Use Git LFS or .gitignore.",
					float64(fi.Size())/(1024*1024), float64(limit)/(1024*1024)),
				Fixable: false,
			})
		}
	}

	return issues, nil
}
