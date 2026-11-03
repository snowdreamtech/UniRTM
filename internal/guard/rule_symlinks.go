// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package guard

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

type SymlinksRule struct{}

func NewSymlinksRule() Rule {
	return &SymlinksRule{}
}

func (r *SymlinksRule) ID() string {
	return "builtin:check-symlinks"
}

func (r *SymlinksRule) Category() Category {
	return CategoryHygiene
}

func (r *SymlinksRule) Phase() Phase {
	return PhaseLinter
}

func (r *SymlinksRule) DefaultExtensions() []string {
	return nil
}

func (r *SymlinksRule) Execute(ctx context.Context, files []string, opts ExecuteOptions) ([]Issue, error) {
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

		fi, err := os.Lstat(fullPath)
		if err != nil {
			continue
		}

		if fi.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(fullPath)
			if err != nil {
				issues = append(issues, Issue{
					FilePath: rel,
					Line:     1,
					RuleID:   r.ID(),
					Category: r.Category(),
					Phase:    r.Phase(),
					Severity: SeverityError,
					Message:  fmt.Sprintf("Failed to read symlink target: %v", err),
					Fixable:  false,
				})
				continue
			}

			if filepath.IsAbs(target) {
				issues = append(issues, Issue{
					FilePath: rel,
					Line:     1,
					RuleID:   r.ID(),
					Category: r.Category(),
					Phase:    r.Phase(),
					Severity: SeverityError,
					Message:  fmt.Sprintf("Absolute symlink forbidden: points to absolute path %s", target),
					Fixable:  false,
				})
			} else {
				// Check broken symlink (os.Stat follows symlinks)
				if _, err := os.Stat(fullPath); os.IsNotExist(err) {
					issues = append(issues, Issue{
						FilePath: rel,
						Line:     1,
						RuleID:   r.ID(),
						Category: r.Category(),
						Phase:    r.Phase(),
						Severity: SeverityError,
						Message:  fmt.Sprintf("Broken symlink: target does not exist (%s -> %s)", rel, target),
						Fixable:  false,
					})
				}
			}
		}
	}

	return issues, nil
}
