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

// BinaryArtifactsRule prevents accidental commit of compiled binaries, archives, and build artifacts
type BinaryArtifactsRule struct {
	forbiddenExtensions map[string]string
}

func NewBinaryArtifactsRule() Rule {
	return &BinaryArtifactsRule{
		forbiddenExtensions: map[string]string{
			".exe":    "Windows executable",
			".dll":    "Windows dynamic library",
			".so":     "Linux shared library",
			".dylib":  "macOS dynamic library",
			".class":  "Java compiled bytecode",
			".pyc":    "Python compiled bytecode",
			".pyo":    "Python optimized bytecode",
			".o":      "C/C++ object file",
			".a":      "Static library archive",
			".jar":    "Java archive",
			".war":    "Java web application archive",
			".tar.gz": "Compressed tarball",
			".tgz":    "Compressed tarball",
			".zip":    "Zip archive",
			".7z":     "7-Zip archive",
			".rar":    "RAR archive",
		},
	}
}

func (r *BinaryArtifactsRule) ID() string {
	return "builtin:forbid-binary-artifacts"
}

func (r *BinaryArtifactsRule) Category() Category {
	return CategoryHygiene
}

func (r *BinaryArtifactsRule) Phase() Phase {
	return PhaseLinter
}

func (r *BinaryArtifactsRule) DefaultExtensions() []string {
	return []string{
		".exe", ".dll", ".so", ".dylib", ".class", ".pyc", ".pyo",
		".o", ".a", ".jar", ".war", ".tar.gz", ".tgz", ".zip", ".7z", ".rar",
	}
}

func (r *BinaryArtifactsRule) Execute(ctx context.Context, files []string, opts ExecuteOptions) ([]Issue, error) {
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
		if err != nil || fi.IsDir() || fi.Mode()&os.ModeSymlink != 0 {
			continue
		}

		lowerRel := strings.ToLower(rel)
		for ext, desc := range r.forbiddenExtensions {
			if strings.HasSuffix(lowerRel, ext) {
				issues = append(issues, Issue{
					FilePath: rel,
					Line:     1,
					RuleID:   r.ID(),
					Category: r.Category(),
					Phase:    r.Phase(),
					Severity: SeverityError,
					Message:  fmt.Sprintf("Found forbidden binary artifact '%s' (%s). Binary files should not be committed to Git. Please remove it or add it to .gitignore", rel, desc),
					Fixable:  false,
				})
				break
			}
		}
	}

	return issues, nil
}
