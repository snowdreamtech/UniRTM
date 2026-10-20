// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package guard

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var nonScriptExtensions = map[string]bool{
	".bat": true, ".cmd": true, ".ps1": true, ".exe": true,
	".txt": true, ".pdf": true, ".doc": true, ".docx": true,
	".ini": true, ".editorconfig": true, ".properties": true,
	".lock": true, ".lockb": true, ".sum": true, ".resolved": true,
	".json": true, ".yaml": true, ".yml": true, ".toml": true,
	".md": true, ".html": true, ".css": true, ".png": true, ".jpg": true, ".svg": true,
	".go": true, ".ts": true, ".tsx": true, ".js": true, ".jsx": true, ".vue": true,
}

// ExecutablesHaveShebangsRule ensures that executable text scripts start with a shebang (#!)
type ExecutablesHaveShebangsRule struct{}

func NewExecutablesHaveShebangsRule() Rule {
	return &ExecutablesHaveShebangsRule{}
}

func (r *ExecutablesHaveShebangsRule) ID() string {
	return "builtin:check-executables-have-shebangs"
}

func (r *ExecutablesHaveShebangsRule) Category() Category {
	return CategoryHygiene
}

func (r *ExecutablesHaveShebangsRule) Phase() Phase {
	return PhaseLinter
}

func (r *ExecutablesHaveShebangsRule) DefaultExtensions() []string {
	return nil
}

func (r *ExecutablesHaveShebangsRule) Execute(ctx context.Context, files []string, opts ExecuteOptions) ([]Issue, error) {
	if runtime.GOOS == "windows" {
		return nil, nil
	}

	var issues []Issue

	for _, rel := range files {
		select {
		case <-ctx.Done():
			return issues, ctx.Err()
		default:
		}

		ext := strings.ToLower(filepath.Ext(rel))
		if nonScriptExtensions[ext] {
			continue
		}

		baseName := filepath.Base(rel)
		if strings.EqualFold(baseName, "Makefile") || strings.EqualFold(baseName, "LICENSE") {
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

		mode := fi.Mode()
		// Check if any executable bit is set (Owner, Group, Others)
		if mode&0111 != 0 {
			fh, err := os.Open(fullPath)
			if err != nil {
				continue
			}

			head := make([]byte, 4)
			n, _ := io.ReadFull(fh, head)
			_ = fh.Close()

			if n >= 4 {
				// Skip binary executable headers (ELF, Mach-O, Windows PE)
				if bytes.Equal(head[:4], []byte{0x7f, 'E', 'L', 'F'}) || // ELF
					bytes.Equal(head[:4], []byte{0xfe, 0xed, 0xfa, 0xce}) || // Mach-O 32
					bytes.Equal(head[:4], []byte{0xfe, 0xed, 0xfa, 0xcf}) || // Mach-O 64
					bytes.Equal(head[:4], []byte{0xcf, 0xfa, 0xed, 0xfe}) ||
					bytes.Equal(head[:2], []byte{'M', 'Z'}) { // PE
					continue
				}
			}

			if n < 2 || head[0] != '#' || head[1] != '!' {
				issues = append(issues, Issue{
					FilePath: rel,
					Line:     1,
					RuleID:   r.ID(),
					Category: r.Category(),
					Phase:    r.Phase(),
					Severity: SeverityError,
					Message:  "Executable script lacks shebang (#!) header",
					Fixable:  false,
				})
			}
		}
	}

	return issues, nil
}

// ShebangScriptsAreExecutableRule ensures scripts with shebang headers have executable bit (+x)
type ShebangScriptsAreExecutableRule struct{}

func NewShebangScriptsAreExecutableRule() Rule {
	return &ShebangScriptsAreExecutableRule{}
}

func (r *ShebangScriptsAreExecutableRule) ID() string {
	return "builtin:check-shebang-scripts-are-executable"
}

func (r *ShebangScriptsAreExecutableRule) Category() Category {
	return CategoryHygiene
}

func (r *ShebangScriptsAreExecutableRule) Phase() Phase {
	return PhaseFixer
}

func (r *ShebangScriptsAreExecutableRule) DefaultExtensions() []string {
	return nil
}

func (r *ShebangScriptsAreExecutableRule) Execute(ctx context.Context, files []string, opts ExecuteOptions) ([]Issue, error) {
	if runtime.GOOS == "windows" {
		return nil, nil
	}

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

		// Read first two bytes
		fh, err := os.Open(fullPath)
		if err != nil {
			continue
		}
		head := make([]byte, 2)
		n, _ := io.ReadFull(fh, head)
		_ = fh.Close()

		if n == 2 && head[0] == '#' && head[1] == '!' {
			mode := fi.Mode()
			// If missing any executable bit
			if mode&0111 == 0 {
				fixed := false
				if opts.Fix && !opts.DryRun {
					// Set +x (e.g. 0755 or preserve existing permissions + 0111)
					newMode := mode.Perm() | 0111
					if err := os.Chmod(fullPath, newMode); err == nil {
						fixed = true
					}
				}

				issues = append(issues, Issue{
					FilePath: rel,
					Line:     1,
					RuleID:   r.ID(),
					Category: r.Category(),
					Phase:    r.Phase(),
					Severity: SeverityError,
					Message:  "Script starts with shebang (#!) but lacks executable permission (+x)",
					Fixable:  true,
					Fixed:    fixed,
				})
			}
		}
	}

	return issues, nil
}
