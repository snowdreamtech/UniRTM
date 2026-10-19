// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package builtin

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"regexp"

	"github.com/snowdreamtech/unirtm/internal/guard"
)

var trailingWhitespaceRegex = regexp.MustCompile(`[ \t]+(\r?\n|$)`)

type TrailingWhitespaceRule struct{}

func NewTrailingWhitespaceRule() guard.Rule {
	return &TrailingWhitespaceRule{}
}

func (r *TrailingWhitespaceRule) ID() string {
	return "builtin:trailing-whitespace"
}

func (r *TrailingWhitespaceRule) Category() guard.Category {
	return guard.CategoryHygiene
}

func (r *TrailingWhitespaceRule) Phase() guard.Phase {
	return guard.PhaseFixer
}

func (r *TrailingWhitespaceRule) DefaultExtensions() []string {
	return nil // Matches all text files unless excluded
}

func (r *TrailingWhitespaceRule) Execute(ctx context.Context, files []string, opts guard.ExecuteOptions) ([]guard.Issue, error) {
	var issues []guard.Issue
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

		fullPath := filepath.Join(opts.WorkspaceDir, rel)
		fi, err := os.Lstat(fullPath)
		if err != nil || fi.Mode()&os.ModeSymlink != 0 || fi.Size() > maxSize {
			continue
		}

		content, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}

		// Detect binary file by looking for null bytes in the first 8000 bytes
		checkLen := len(content)
		if checkLen > 8000 {
			checkLen = 8000
		}
		if bytes.IndexByte(content[:checkLen], 0) != -1 {
			continue
		}

		cleaned := trailingWhitespaceRegex.ReplaceAll(content, []byte("$1"))
		if !bytes.Equal(content, cleaned) {
			fixed := false
			if opts.Fix && !opts.DryRun {
				if err := os.WriteFile(fullPath, cleaned, fi.Mode().Perm()); err == nil {
					fixed = true
				}
			}

			issues = append(issues, guard.Issue{
				FilePath: rel,
				Line:     1,
				RuleID:   r.ID(),
				Category: r.Category(),
				Phase:    r.Phase(),
				Severity: guard.SeverityWarning,
				Message:  "Contains trailing whitespace on one or more lines",
				Fixable:  true,
				Fixed:    fixed,
			})
		}
	}

	return issues, nil
}
