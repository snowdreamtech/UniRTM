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

// EmptyCodeFilesRule flags empty source code files (0 bytes) to prevent accidental commit of unfinished stubs
type EmptyCodeFilesRule struct {
	whitelistNames map[string]bool
	supportedExts  map[string]bool
}

func NewEmptyCodeFilesRule() Rule {
	exts := []string{
		".go", ".ts", ".js", ".vue", ".jsx", ".tsx",
		".py", ".rs", ".java", ".c", ".cpp", ".h", ".cs", ".rb", ".php",
	}
	extMap := make(map[string]bool, len(exts))
	for _, ext := range exts {
		extMap[ext] = true
	}

	return &EmptyCodeFilesRule{
		whitelistNames: map[string]bool{
			"__init__.py":  true,
			"__init__.pyi": true,
			".gitkeep":     true,
			".keep":        true,
			".empty":       true,
		},
		supportedExts: extMap,
	}
}

func (r *EmptyCodeFilesRule) ID() string {
	return "builtin:forbid-empty-code-files"
}

func (r *EmptyCodeFilesRule) Category() Category {
	return CategoryHygiene
}

func (r *EmptyCodeFilesRule) Phase() Phase {
	return PhaseLinter
}

func (r *EmptyCodeFilesRule) DefaultExtensions() []string {
	return []string{
		".go", ".ts", ".js", ".vue", ".jsx", ".tsx",
		".py", ".rs", ".java", ".c", ".cpp", ".h", ".cs", ".rb", ".php",
	}
}

func (r *EmptyCodeFilesRule) Execute(ctx context.Context, files []string, opts ExecuteOptions) ([]Issue, error) {
	var issues []Issue

	for _, rel := range files {
		select {
		case <-ctx.Done():
			return issues, ctx.Err()
		default:
		}

		baseName := strings.ToLower(filepath.Base(rel))
		if r.whitelistNames[baseName] {
			continue
		}

		ext := strings.ToLower(filepath.Ext(rel))
		if !r.supportedExts[ext] {
			continue
		}

		fullPath := rel
		if !filepath.IsAbs(fullPath) {
			fullPath = filepath.Join(opts.WorkspaceDir, rel)
		}

		fi, err := os.Lstat(fullPath)
		if err != nil || fi.IsDir() || fi.Mode()&os.ModeSymlink != 0 {
			continue
		}

		if fi.Size() == 0 {
			issues = append(issues, Issue{
				FilePath: rel,
				Line:     1,
				RuleID:   r.ID(),
				Category: r.Category(),
				Phase:    r.Phase(),
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("Empty code file '%s' (0 bytes). Empty source files are typically unfinished stubs or accidental commits.", rel),
				Fixable:  false,
			})
		}
	}

	return issues, nil
}
