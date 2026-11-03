// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package guard

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
)

var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

type Utf8BOMRule struct{}

func NewUtf8BOMRule() Rule {
	return &Utf8BOMRule{}
}

func (r *Utf8BOMRule) ID() string {
	return "builtin:check-utf8-bom"
}

func (r *Utf8BOMRule) Category() Category {
	return CategoryHygiene
}

func (r *Utf8BOMRule) Phase() Phase {
	return PhaseFixer
}

func (r *Utf8BOMRule) DefaultExtensions() []string {
	return nil
}

func (r *Utf8BOMRule) Execute(ctx context.Context, files []string, opts ExecuteOptions) ([]Issue, error) {
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
		if err != nil || fi.Mode()&os.ModeSymlink != 0 || fi.Size() < 3 {
			continue
		}

		content, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}

		if bytes.HasPrefix(content, utf8BOM) {
			fixed := false
			if opts.Fix && !opts.DryRun {
				cleaned := content[3:]
				if err := os.WriteFile(fullPath, cleaned, fi.Mode().Perm()); err == nil {
					fixed = true
				}
			}

			issues = append(issues, Issue{
				FilePath: rel,
				Line:     1,
				RuleID:   r.ID(),
				Category: r.Category(),
				Phase:    r.Phase(),
				Severity: SeverityWarning,
				Message:  "Contains prohibited UTF-8 Byte Order Mark (BOM)",
				Fixable:  true,
				Fixed:    fixed,
			})
		}
	}

	return issues, nil
}
