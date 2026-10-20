// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package guard

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// MixedLineEndingsRule detects and repairs mixed line endings (CRLF vs LF) within the same file
type MixedLineEndingsRule struct{}

func NewMixedLineEndingsRule() Rule {
	return &MixedLineEndingsRule{}
}

func (r *MixedLineEndingsRule) ID() string {
	return "builtin:mixed-line-endings"
}

func (r *MixedLineEndingsRule) Category() Category {
	return CategoryHygiene
}

func (r *MixedLineEndingsRule) Phase() Phase {
	return PhaseFixer
}

func (r *MixedLineEndingsRule) DefaultExtensions() []string {
	return nil
}

func (r *MixedLineEndingsRule) Execute(ctx context.Context, files []string, opts ExecuteOptions) ([]Issue, error) {
	var issues []Issue
	maxSize := opts.MaxFileSize
	if maxSize <= 0 {
		maxSize = 2 * 1024 * 1024 // 2MB
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
		if err != nil || fi.IsDir() || fi.Mode()&os.ModeSymlink != 0 || fi.Size() > maxSize {
			continue
		}

		content, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}

		// Skip binary files (null byte check in first 8000 bytes)
		checkLen := len(content)
		if checkLen > 8000 {
			checkLen = 8000
		}
		if bytes.IndexByte(content[:checkLen], 0) != -1 {
			continue
		}

		crlfCount := bytes.Count(content, []byte("\r\n"))
		totalLfCount := bytes.Count(content, []byte("\n"))
		standaloneLfCount := totalLfCount - crlfCount

		// Check for mixed line endings
		if crlfCount > 0 && standaloneLfCount > 0 {
			fixed := false
			if opts.Fix && !opts.DryRun {
				// Normalize to uniform LF
				normalized := bytes.ReplaceAll(content, []byte("\r\n"), []byte("\n"))
				if err := os.WriteFile(fullPath, normalized, fi.Mode().Perm()); err == nil {
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
				Message:  fmt.Sprintf("Mixed line endings detected (%d CRLF and %d LF lines)", crlfCount, standaloneLfCount),
				Fixable:  true,
				Fixed:    fixed,
			})
		}
	}

	return issues, nil
}
