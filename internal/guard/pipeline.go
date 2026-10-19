// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package guard

import (
	"context"
	"fmt"
	"sort"
	"time"
)

// PipelineSummary provides statistical metrics for a guard execution run
type PipelineSummary struct {
	TotalFilesScanned int           `json:"total_files_scanned"`
	TotalIssues       int           `json:"total_issues"`
	ErrorsCount       int           `json:"errors_count"`
	WarningsCount     int           `json:"warnings_count"`
	FixedCount        int           `json:"fixed_count"`
	Duration          time.Duration `json:"duration"`
}

// PipelineResult aggregates issues and summaries across all executed phases
type PipelineResult struct {
	Issues    []Issue         `json:"issues"`
	Summary   PipelineSummary `json:"summary"`
	HasErrors bool            `json:"has_errors"`
	Logs      []string        `json:"logs"`
}

// Pipeline orchestrates rule execution across the 3 phases
type Pipeline struct {
	WorkspaceDir string
	Config       GuardConfig
	Rules        []Rule
}

// NewPipeline creates a Pipeline populated with all builtin rules and configured slots
func NewPipeline(workspaceDir string, cfg GuardConfig) *Pipeline {
	p := &Pipeline{
		WorkspaceDir: workspaceDir,
		Config:       cfg,
		Rules:        GetAllBuiltinRules(),
	}
	return p
}

// Run executes the full or filtered pipeline against candidate files
func (p *Pipeline) Run(ctx context.Context, files []string, opts ExecuteOptions) (*PipelineResult, error) {
	start := time.Now()
	opts.WorkspaceDir = p.WorkspaceDir

	var allIssues []Issue
	var logs []string
	modeStr := opts.Mode
	if modeStr == "" {
		modeStr = "staged"
	}
	logs = append(logs, fmt.Sprintf("[%s] [Quality Guard] Starting quality guard pipeline (mode: %s, scanned files: %d)", time.Now().Format("15:04:05"), modeStr, len(files)))

	// Group rules by phase
	var fixerRules, formatterRules, linterRules []Rule
	for _, r := range p.Rules {
		// Check if rule is disabled in config
		if rCfg, ok := p.Config.Rules[r.ID()]; ok && rCfg.Enabled != nil && !*rCfg.Enabled {
			continue
		}

		switch r.Phase() {
		case PhaseFixer:
			fixerRules = append(fixerRules, r)
		case PhaseFormatter:
			formatterRules = append(formatterRules, r)
		case PhaseLinter:
			linterRules = append(linterRules, r)
		}
	}

	// Phase 1: Fixer (Text & AST Repairs)
	phase1Start := time.Now()
	phase1Before := len(allIssues)
	for _, r := range fixerRules {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		issues, err := r.Execute(ctx, files, opts)
		if err == nil && len(issues) > 0 {
			allIssues = append(allIssues, issues...)
		}
	}
	phase1Issues := len(allIssues) - phase1Before
	logs = append(logs, fmt.Sprintf("[%s] [Phase 1 Fixer] Completed native hygiene check: %d rules evaluated, %d issues found (%v)", time.Now().Format("15:04:05"), len(fixerRules), phase1Issues, time.Since(phase1Start).Round(time.Millisecond)))

	// Phase 2: Formatter (Visual Layout Standardization)
	phase2Start := time.Now()
	phase2Before := len(allIssues)
	for _, r := range formatterRules {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		issues, err := r.Execute(ctx, files, opts)
		if err == nil && len(issues) > 0 {
			allIssues = append(allIssues, issues...)
		}
	}
	phase2Issues := len(allIssues) - phase2Before
	logs = append(logs, fmt.Sprintf("[%s] [Phase 2 Formatter] Completed code formatting: %d rules evaluated, %d deviations found (%v)", time.Now().Format("15:04:05"), len(formatterRules), phase2Issues, time.Since(phase2Start).Round(time.Millisecond)))

	// Phase 3: Linter (Strict Read-only Rule Verification)
	phase3Start := time.Now()
	phase3Before := len(allIssues)
	for _, r := range linterRules {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		issues, err := r.Execute(ctx, files, opts)
		if err == nil && len(issues) > 0 {
			allIssues = append(allIssues, issues...)
		}
	}
	phase3Issues := len(allIssues) - phase3Before
	logs = append(logs, fmt.Sprintf("[%s] [Phase 3 Linter] Completed static analysis & secret protection: %d rules evaluated, %d issues found (%v)", time.Now().Format("15:04:05"), len(linterRules), phase3Issues, time.Since(phase3Start).Round(time.Millisecond)))

	// Calculate summary metrics
	summary := PipelineSummary{
		TotalFilesScanned: len(files),
		TotalIssues:       len(allIssues),
		Duration:          time.Since(start),
	}

	hasErrors := false
	for _, iss := range allIssues {
		if iss.Fixed {
			summary.FixedCount++
		}
		if iss.Severity == SeverityError {
			summary.ErrorsCount++
			hasErrors = true
		} else if iss.Severity == SeverityWarning {
			summary.WarningsCount++
		}
	}

	// Sort issues deterministically by FilePath and Line
	sort.Slice(allIssues, func(i, j int) bool {
		if allIssues[i].FilePath != allIssues[j].FilePath {
			return allIssues[i].FilePath < allIssues[j].FilePath
		}
		return allIssues[i].Line < allIssues[j].Line
	})

	if hasErrors {
		logs = append(logs, fmt.Sprintf("[%s] [Quality Guard] Guard blocked: %d blocking violations, %d warnings detected (total: %v)", time.Now().Format("15:04:05"), summary.ErrorsCount, summary.WarningsCount, summary.Duration.Round(time.Millisecond)))
	} else {
		logs = append(logs, fmt.Sprintf("[%s] [Quality Guard] Guard checks passed: 0 blocking violations, all compliant (total: %v)", time.Now().Format("15:04:05"), summary.Duration.Round(time.Millisecond)))
	}

	return &PipelineResult{
		Issues:    allIssues,
		Summary:   summary,
		HasErrors: hasErrors,
		Logs:      logs,
	}, nil
}
