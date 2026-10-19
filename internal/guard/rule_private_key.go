// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package guard

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"

)

var privateKeyRegex = regexp.MustCompile(`-----BEGIN (?:[A-Z0-9 ]+ )?PRIVATE KEY-----`)

type PrivateKeyRule struct{}

func NewPrivateKeyRule() Rule {
	return &PrivateKeyRule{}
}

func (r *PrivateKeyRule) ID() string {
	return "builtin:detect-private-key"
}

func (r *PrivateKeyRule) Category() Category {
	return CategorySecret
}

func (r *PrivateKeyRule) Phase() Phase {
	return PhaseLinter
}

func (r *PrivateKeyRule) DefaultExtensions() []string {
	return nil
}

func (r *PrivateKeyRule) Execute(ctx context.Context, files []string, opts ExecuteOptions) ([]Issue, error) {
	var issues []Issue

	for _, rel := range files {
		select {
		case <-ctx.Done():
			return issues, ctx.Err()
		default:
		}

		base := strings.ToLower(filepath.Base(rel))
		if base == "id_rsa" || base == "id_dsa" || base == "id_ecdsa" || base == "id_ed25519" {
			issues = append(issues, Issue{
				FilePath: rel,
				Line:     1,
				RuleID:   r.ID(),
				Category: r.Category(),
				Phase:    r.Phase(),
				Severity: SeverityError,
				Message:  "Private key file detected by filename: " + base,
				Fixable:  false,
			})
			continue
		}

		fullPath := rel; if !filepath.IsAbs(fullPath) { fullPath = filepath.Join(opts.WorkspaceDir, rel) }
		fi, err := os.Lstat(fullPath)
		if err != nil || fi.Mode()&os.ModeSymlink != 0 || fi.Size() > 2*1024*1024 {
			continue
		}

		content, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}

		// Skip binaries
		checkLen := len(content)
		if checkLen > 4000 {
			checkLen = 4000
		}
		if bytes.IndexByte(content[:checkLen], 0) != -1 {
			continue
		}

		if loc := privateKeyRegex.FindIndex(content); loc != nil {
			// Find line number
			line := bytes.Count(content[:loc[0]], []byte("\n")) + 1
			issues = append(issues, Issue{
				FilePath: rel,
				Line:     line,
				RuleID:   r.ID(),
				Category: r.Category(),
				Phase:    r.Phase(),
				Severity: SeverityError,
				Message:  "Sensitive private key header pattern detected (leak hazard)",
				Fixable:  false,
			})
		}
	}

	return issues, nil
}
