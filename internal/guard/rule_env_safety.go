// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package guard

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var sensitiveKeyRegex = regexp.MustCompile(`(?i)(?:AWS_SECRET|GITHUB_TOKEN|PRIVATE_KEY|DATABASE_URL|SECRET_KEY|AUTH_TOKEN|PASSWORD|API_KEY)\s*=\s*['"]?([^\s'"]+)`)

type EnvSafetyRule struct{}

func NewEnvSafetyRule() Rule {
	return &EnvSafetyRule{}
}

func (r *EnvSafetyRule) ID() string {
	return "builtin:check-env-safety"
}

func (r *EnvSafetyRule) Category() Category {
	return CategorySecret
}

func (r *EnvSafetyRule) Phase() Phase {
	return PhaseLinter
}

func (r *EnvSafetyRule) DefaultExtensions() []string {
	return nil
}

func (r *EnvSafetyRule) Execute(ctx context.Context, files []string, opts ExecuteOptions) ([]Issue, error) {
	var issues []Issue

	for _, rel := range files {
		select {
		case <-ctx.Done():
			return issues, ctx.Err()
		default:
		}

		base := strings.ToLower(filepath.Base(rel))
		// Only check .env files
		if !strings.HasPrefix(base, ".env") && !strings.HasSuffix(base, ".env") {
			continue
		}

		// Allow safe template files: .env.example, .env.sample, .env.template
		if strings.HasSuffix(base, ".example") || strings.HasSuffix(base, ".sample") || strings.HasSuffix(base, ".template") {
			continue
		}

		fullPath := rel
		if !filepath.IsAbs(fullPath) {
			fullPath = filepath.Join(opts.WorkspaceDir, rel)
		}
		f, err := os.Open(fullPath)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(f)
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			matches := sensitiveKeyRegex.FindStringSubmatch(line)
			if len(matches) > 1 {
				val := matches[1]
				// Ignore placeholder values like xxx, your-key-here, test, localhost, 127.0.0.1
				lowerVal := strings.ToLower(val)
				if lowerVal == "xxx" || strings.Contains(lowerVal, "your_") || strings.Contains(lowerVal, "placeholder") || strings.Contains(lowerVal, "example") || lowerVal == "password" {
					continue
				}

				issues = append(issues, Issue{
					FilePath: rel,
					Line:     lineNum,
					RuleID:   r.ID(),
					Category: r.Category(),
					Phase:    r.Phase(),
					Severity: SeverityError,
					Message:  "Potential raw production secret detected in .env file (leak risk)",
					Fixable:  false,
				})
				break
			}
		}
		f.Close()
	}

	return issues, nil
}
