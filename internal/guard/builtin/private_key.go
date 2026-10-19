// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package builtin

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/snowdreamtech/unirtm/internal/guard"
)

var privateKeyRegex = regexp.MustCompile(`-----BEGIN (?:[A-Z0-9 ]+ )?PRIVATE KEY-----`)

type PrivateKeyRule struct{}

func NewPrivateKeyRule() guard.Rule {
	return &PrivateKeyRule{}
}

func (r *PrivateKeyRule) ID() string {
	return "builtin:detect-private-key"
}

func (r *PrivateKeyRule) Category() guard.Category {
	return guard.CategorySecret
}

func (r *PrivateKeyRule) Phase() guard.Phase {
	return guard.PhaseLinter
}

func (r *PrivateKeyRule) DefaultExtensions() []string {
	return nil
}

func (r *PrivateKeyRule) Execute(ctx context.Context, files []string, opts guard.ExecuteOptions) ([]guard.Issue, error) {
	var issues []guard.Issue

	for _, rel := range files {
		select {
		case <-ctx.Done():
			return issues, ctx.Err()
		default:
		}

		base := strings.ToLower(filepath.Base(rel))
		if base == "id_rsa" || base == "id_dsa" || base == "id_ecdsa" || base == "id_ed25519" {
			issues = append(issues, guard.Issue{
				FilePath: rel,
				Line:     1,
				RuleID:   r.ID(),
				Category: r.Category(),
				Phase:    r.Phase(),
				Severity: guard.SeverityError,
				Message:  "Private key file detected by filename: " + base,
				Fixable:  false,
			})
			continue
		}

		fullPath := filepath.Join(opts.WorkspaceDir, rel)
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
			issues = append(issues, guard.Issue{
				FilePath: rel,
				Line:     line,
				RuleID:   r.ID(),
				Category: r.Category(),
				Phase:    r.Phase(),
				Severity: guard.SeverityError,
				Message:  "Sensitive private key header pattern detected (leak hazard)",
				Fixable:  false,
			})
		}
	}

	return issues, nil
}
