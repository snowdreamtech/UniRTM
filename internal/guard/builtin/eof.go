// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package builtin

import (
	"bytes"
	"context"
	"os"
	"path/filepath"

	"github.com/snowdreamtech/unirtm/internal/guard"
)

type EndOfFileRule struct{}

func NewEndOfFileRule() guard.Rule {
	return &EndOfFileRule{}
}

func (r *EndOfFileRule) ID() string {
	return "builtin:end-of-file"
}

func (r *EndOfFileRule) Category() guard.Category {
	return guard.CategoryHygiene
}

func (r *EndOfFileRule) Phase() guard.Phase {
	return guard.PhaseFixer
}

func (r *EndOfFileRule) DefaultExtensions() []string {
	return nil
}

func (r *EndOfFileRule) Execute(ctx context.Context, files []string, opts guard.ExecuteOptions) ([]guard.Issue, error) {
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
		if err != nil || fi.Mode()&os.ModeSymlink != 0 || fi.Size() == 0 || fi.Size() > maxSize {
			continue
		}

		content, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}

		// Skip binary files
		checkLen := len(content)
		if checkLen > 8000 {
			checkLen = 8000
		}
		if bytes.IndexByte(content[:checkLen], 0) != -1 {
			continue
		}

		// Normalize trailing newlines: strip trailing \r and \n, then append single \n (or \r\n if file used CRLF)
		isCRLF := bytes.Contains(content, []byte("\r\n"))
		trimmed := bytes.TrimRight(content, "\r\n")

		var cleaned []byte
		if isCRLF {
			cleaned = append(trimmed, []byte("\r\n")...)
		} else {
			cleaned = append(trimmed, '\n')
		}

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
				Message:  "File does not end with a single newline",
				Fixable:  true,
				Fixed:    fixed,
			})
		}
	}

	return issues, nil
}
