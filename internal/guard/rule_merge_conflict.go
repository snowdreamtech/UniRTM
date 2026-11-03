// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package guard

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
)

type MergeConflictRule struct{}

func NewMergeConflictRule() Rule {
	return &MergeConflictRule{}
}

func (r *MergeConflictRule) ID() string {
	return "builtin:check-merge-conflict"
}

func (r *MergeConflictRule) Category() Category {
	return CategoryHygiene
}

func (r *MergeConflictRule) Phase() Phase {
	return PhaseLinter
}

func (r *MergeConflictRule) DefaultExtensions() []string {
	return nil
}

func (r *MergeConflictRule) Execute(ctx context.Context, files []string, opts ExecuteOptions) ([]Issue, error) {
	var issues []Issue
	maxSize := opts.MaxFileSize
	if maxSize <= 0 {
		maxSize = 5 * 1024 * 1024 // 5MB
	}

	for _, rel := range files {
		select {
		case <-ctx.Done():
			return issues, ctx.Err()
		default:
		}

		lowerRel := strings.ToLower(filepath.ToSlash(rel))
		if strings.HasSuffix(lowerRel, "_test.go") || strings.Contains(lowerRel, "testdata/") {
			continue
		}

		fullPath := rel
		if !filepath.IsAbs(fullPath) {
			fullPath = filepath.Join(opts.WorkspaceDir, rel)
		}
		fi, err := os.Lstat(fullPath)
		if err != nil || fi.Mode()&os.ModeSymlink != 0 || fi.Size() > maxSize {
			continue
		}

		f, err := os.Open(fullPath)
		if err != nil {
			continue
		}

		// Read first chunk to check if binary
		header := make([]byte, 1024)
		n, _ := f.Read(header)
		if bytes.IndexByte(header[:n], 0) != -1 {
			f.Close()
			continue
		}

		f.Seek(0, 0)
		scanner := bufio.NewScanner(f)
		lineNum := 0
		hasConflict := false

		for scanner.Scan() {
			lineNum++
			line := scanner.Text()
			if strings.HasPrefix(line, "<<<<<<< ") ||
				strings.HasPrefix(line, "=======") && len(strings.TrimSpace(line)) == 7 ||
				strings.HasPrefix(line, ">>>>>>> ") {
				issues = append(issues, Issue{
					FilePath: rel,
					Line:     lineNum,
					RuleID:   r.ID(),
					Category: r.Category(),
					Phase:    r.Phase(),
					Severity: SeverityError,
					Message:  "Unresolved Git merge conflict marker detected: " + strings.TrimSpace(line),
					Fixable:  false,
				})
				hasConflict = true
				break // One report per file is sufficient
			}
		}
		f.Close()

		if hasConflict {
			continue
		}
	}

	return issues, nil
}
