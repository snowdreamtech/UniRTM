// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package guard

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type secretPattern struct {
	Name    string
	Regex   *regexp.Regexp
	MaskLen int
}

var cloudSecretPatterns = []secretPattern{
	{
		Name:    "AWS Access Key ID",
		Regex:   regexp.MustCompile(`\b(AKIA[0-9A-Z]{16})\b`),
		MaskLen: 4,
	},
	{
		Name:    "GitHub Personal Access Token",
		Regex:   regexp.MustCompile(`\b(gh[pousr]_[0-9a-zA-Z]{36})\b`),
		MaskLen: 4,
	},
	{
		Name:    "OpenAI API Key",
		Regex:   regexp.MustCompile(`\b(sk-[a-zA-Z0-9]{48}|sk-proj-[a-zA-Z0-9-_]{30,})\b`),
		MaskLen: 6,
	},
	{
		Name:    "Slack Token",
		Regex:   regexp.MustCompile(`\b(xox[baprs]-[0-9a-zA-Z]{10,48})\b`),
		MaskLen: 5,
	},
	{
		Name:    "Google API Key",
		Regex:   regexp.MustCompile(`\b(AIza[0-9A-Za-z\-_]{35})\b`),
		MaskLen: 5,
	},
}

// CloudSecretsRule detects common cloud credentials and API keys in committed files
type CloudSecretsRule struct{}

func NewCloudSecretsRule() Rule {
	return &CloudSecretsRule{}
}

func (r *CloudSecretsRule) ID() string {
	return "builtin:detect-cloud-secrets"
}

func (r *CloudSecretsRule) Category() Category {
	return CategorySecret
}

func (r *CloudSecretsRule) Phase() Phase {
	return PhaseLinter
}

func (r *CloudSecretsRule) DefaultExtensions() []string {
	return nil
}

func (r *CloudSecretsRule) Execute(ctx context.Context, files []string, opts ExecuteOptions) ([]Issue, error) {
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

		// Skip test files, mocks, lockfiles, and git configs
		slashRel := filepath.ToSlash(rel)
		if strings.Contains(slashRel, "testdata/") ||
			strings.Contains(slashRel, ".test.") ||
			strings.HasSuffix(slashRel, "_test.go") ||
			strings.HasSuffix(slashRel, ".lock") ||
			strings.HasSuffix(slashRel, ".sum") ||
			strings.HasSuffix(slashRel, ".gitleaks.toml") {
			continue
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

		// Detect binary file by looking for null byte
		checkLen := len(content)
		if checkLen > 8000 {
			checkLen = 8000
		}
		if bytes.IndexByte(content[:checkLen], 0) != -1 {
			continue
		}

		scanner := bufio.NewScanner(bytes.NewReader(content))
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			lineText := scanner.Text()

			for _, p := range cloudSecretPatterns {
				if match := p.Regex.FindString(lineText); match != "" {
					masked := match
					if len(match) > p.MaskLen {
						masked = match[:p.MaskLen] + strings.Repeat("*", len(match)-p.MaskLen)
					}
					issues = append(issues, Issue{
						FilePath: rel,
						Line:     lineNum,
						RuleID:   r.ID(),
						Category: r.Category(),
						Phase:    r.Phase(),
						Severity: SeverityError,
						Message:  fmt.Sprintf("High-risk secret detected: %s (%s)", p.Name, masked),
						Fixable:  false,
					})
				}
			}
		}
	}

	return issues, nil
}
