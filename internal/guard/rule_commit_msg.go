// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package guard

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// conventionalCommitHeaderRegex matches: <type>(<scope>)!: <subject> or <type>: <subject>
	conventionalCommitHeaderRegex = regexp.MustCompile(`^([a-zA-Z0-9_\-]+)(\(([a-zA-Z0-9_\-\.\/ ]+)\))?(!)?: (.+)$`)
	// allowedSpecialCommitRegex allows standard git merges, reverts, and interactive fixup commits
	allowedSpecialCommitRegex = regexp.MustCompile(`^(Merge |Revert |fixup! |squash! |amend! )`)
)

// AllowedCommitTypes lists all supported standard Conventional Commit types
var AllowedCommitTypes = []string{
	"feat", "fix", "docs", "style", "refactor",
	"perf", "test", "build", "ci", "chore", "revert",
}

const (
	// HeaderMaxLength specifies maximum allowed character count for commit header line (corresponding to commitlint header-max-length: 120)
	HeaderMaxLength = 120
	// SubjectMaxLength specifies maximum allowed character count for commit subject (corresponding to commitlint subject-max-length: 120)
	SubjectMaxLength = 120
)

// CommitMsgOptions contains configuration settings for Conventional Commits validation
type CommitMsgOptions struct {
	AllowedTypes     []string
	AllowedScopes    []string
	RequireScope     bool
	HeaderMaxLength  int
	SubjectMaxLength int
}

// DefaultCommitMsgOptions returns standard Conventional Commits default validation options
func DefaultCommitMsgOptions() CommitMsgOptions {
	return CommitMsgOptions{
		AllowedTypes:     AllowedCommitTypes,
		AllowedScopes:    nil,
		RequireScope:     false,
		HeaderMaxLength:  HeaderMaxLength,
		SubjectMaxLength: SubjectMaxLength,
	}
}

// ValidateCommitMessage validates raw commit message string against default Conventional Commits specification
func ValidateCommitMessage(msg string) (bool, string) {
	return ValidateCommitMessageWithOptions(msg, DefaultCommitMsgOptions())
}

// ValidateCommitMessageWithOptions validates raw commit message string against configured options
func ValidateCommitMessageWithOptions(msg string, opts CommitMsgOptions) (bool, string) {
	lines := strings.Split(strings.ReplaceAll(msg, "\r\n", "\n"), "\n")
	if len(lines) == 0 {
		return false, "Commit message cannot be empty"
	}

	// Filter out Git comment lines starting with #
	var contentLines []string
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if !strings.HasPrefix(trimmed, "#") {
			contentLines = append(contentLines, l)
		}
	}

	if len(contentLines) == 0 || strings.TrimSpace(contentLines[0]) == "" {
		return false, "Commit message cannot be empty"
	}

	header := strings.TrimSpace(contentLines[0])

	// Allow standard Git merge or fixup commits
	if allowedSpecialCommitRegex.MatchString(header) {
		return true, ""
	}

	maxHeader := opts.HeaderMaxLength
	if maxHeader <= 0 {
		maxHeader = HeaderMaxLength
	}
	if len([]rune(header)) > maxHeader {
		return false, fmt.Sprintf("Commit header line exceeds %d characters (current: %d)", maxHeader, len([]rune(header)))
	}

	matches := conventionalCommitHeaderRegex.FindStringSubmatch(header)
	allowedTypes := opts.AllowedTypes
	if len(allowedTypes) == 0 {
		allowedTypes = AllowedCommitTypes
	}

	if len(matches) == 0 {
		return false, fmt.Sprintf(
			"Invalid commit header format %q. Expected: <type>(<scope>): <subject> (valid types: %s)",
			header, strings.Join(allowedTypes, ", "),
		)
	}

	commitType := matches[1]
	hasScope := matches[2] != ""
	scope := matches[3]
	subject := strings.TrimSpace(matches[5])

	// Validate commit type
	typeValid := false
	for _, t := range allowedTypes {
		if commitType == t {
			typeValid = true
			break
		}
	}
	if !typeValid {
		return false, fmt.Sprintf(
			"Invalid commit header format %q. Expected: <type>(<scope>): <subject> (valid types: %s)",
			header, strings.Join(allowedTypes, ", "),
		)
	}

	// Validate scope requirement and allowed values
	if opts.RequireScope && !hasScope {
		return false, "Commit scope is required. Expected: <type>(<scope>): <subject>"
	}
	if len(opts.AllowedScopes) > 0 && hasScope {
		scopeValid := false
		for _, s := range opts.AllowedScopes {
			if scope == s {
				scopeValid = true
				break
			}
		}
		if !scopeValid {
			return false, fmt.Sprintf("Invalid commit scope %q. Expected one of: %s", scope, strings.Join(opts.AllowedScopes, ", "))
		}
	}

	// Validate subject
	if subject == "" {
		return false, "Commit message subject cannot be empty"
	}

	maxSubject := opts.SubjectMaxLength
	if maxSubject <= 0 {
		maxSubject = SubjectMaxLength
	}
	if len([]rune(subject)) > maxSubject {
		return false, fmt.Sprintf("Commit subject exceeds %d characters (current: %d)", maxSubject, len([]rune(subject)))
	}

	if strings.HasSuffix(subject, ".") {
		return false, "Commit subject must not end with a period '.'"
	}

	// Check if body separator is a blank line if body exists
	if len(contentLines) > 1 {
		secondLine := strings.TrimSpace(contentLines[1])
		if secondLine != "" {
			return false, "Second line of commit message must be empty (separating header from body)"
		}
	}

	return true, ""
}

// CommitMsgRule validates git commit messages against Conventional Commits specification
type CommitMsgRule struct {
	opts CommitMsgOptions
}

func NewCommitMsgRule() Rule {
	return &CommitMsgRule{
		opts: DefaultCommitMsgOptions(),
	}
}

func (r *CommitMsgRule) ID() string {
	return "builtin:commit-msg"
}

func (r *CommitMsgRule) Category() Category {
	return CategoryPolicy
}

func (r *CommitMsgRule) Phase() Phase {
	return PhaseLinter
}

func (r *CommitMsgRule) DefaultExtensions() []string {
	return []string{"COMMIT_EDITMSG"}
}

// Configure configures custom rules from [guard.rules."builtin:commit-msg"] in .unirtm.toml
func (r *CommitMsgRule) Configure(cfg RuleConfig) {
	if len(cfg.Types) > 0 {
		r.opts.AllowedTypes = cfg.Types
	}
	if len(cfg.Scopes) > 0 {
		r.opts.AllowedScopes = cfg.Scopes
	}
	r.opts.RequireScope = cfg.RequireScope
	if cfg.HeaderMaxLength > 0 {
		r.opts.HeaderMaxLength = cfg.HeaderMaxLength
	}
	if cfg.SubjectMaxLength > 0 {
		r.opts.SubjectMaxLength = cfg.SubjectMaxLength
	}
}

func (r *CommitMsgRule) Execute(ctx context.Context, files []string, opts ExecuteOptions) ([]Issue, error) {
	var issues []Issue

	for _, rel := range files {
		select {
		case <-ctx.Done():
			return issues, ctx.Err()
		default:
		}

		baseName := filepath.Base(rel)
		if baseName != "COMMIT_EDITMSG" && !strings.HasSuffix(rel, "COMMIT_EDITMSG") {
			continue
		}

		fullPath := rel
		if !filepath.IsAbs(fullPath) {
			fullPath = filepath.Join(opts.WorkspaceDir, rel)
		}

		content, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}

		ok, reason := ValidateCommitMessageWithOptions(string(content), r.opts)
		if !ok {
			issues = append(issues, Issue{
				FilePath: rel,
				Line:     1,
				RuleID:   r.ID(),
				Category: r.Category(),
				Phase:    r.Phase(),
				Severity: SeverityError,
				Message:  reason,
				Fixable:  false,
			})
		}
	}

	return issues, nil
}
