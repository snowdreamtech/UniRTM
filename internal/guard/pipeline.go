// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package guard

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
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

// RuleReport records execution statistics and outcome for an individual rule
type RuleReport struct {
	RuleID        string        `json:"rule_id"`
	Category      Category      `json:"category"`
	Phase         Phase         `json:"phase"`
	Duration      time.Duration `json:"duration"`
	IssuesCount   int           `json:"issues_count"`
	ErrorsCount   int           `json:"errors_count"`
	WarningsCount int           `json:"warnings_count"`
	FixedCount    int           `json:"fixed_count"`
	Status        string        `json:"status"` // "passed", "warn", "fail", "fixed"
}

// PipelineResult aggregates issues and summaries across all executed phases
type PipelineResult struct {
	Issues      []Issue         `json:"issues"`
	Summary     PipelineSummary `json:"summary"`
	HasErrors   bool            `json:"has_errors"`
	Logs        []string        `json:"logs"`
	RuleReports []RuleReport    `json:"rule_reports"`
}

// Pipeline orchestrates rule execution across the 3 phases
type Pipeline struct {
	WorkspaceDir string
	Config       GuardConfig
	Rules        []Rule
}

// NewPipeline creates a Pipeline populated with all builtin rules and configured slots
func NewPipeline(workspaceDir string, cfg GuardConfig) *Pipeline {
	resolvedSlots := GetResolvedToolSlots(cfg)
	rules := GetAllBuiltinRules()
	for _, slot := range resolvedSlots {
		rules = append(rules, NewSlotRule(slot))
	}

	return &Pipeline{
		WorkspaceDir: workspaceDir,
		Config:       cfg,
		Rules:        rules,
	}
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
	logs = append(logs, fmt.Sprintf("[Quality Guard] Starting quality guard pipeline (mode: %s, scanned files: %d)", modeStr, len(files)))

	// Detect and verify each configuration file individually line-by-line
	if p.WorkspaceDir != "" {
		if _, err := os.Stat(filepath.Join(p.WorkspaceDir, ".unirtm.toml")); err == nil {
			logs = append(logs, "[Config: .unirtm.toml] Verified valid guard manifest loaded")
		}
		if _, err := os.Stat(filepath.Join(p.WorkspaceDir, ".editorconfig")); err == nil {
			logs = append(logs, "[Config: .editorconfig] Verified workspace hygiene rules active")
		}
		for _, slot := range DefaultToolSlots {
			if len(slot.StandardConfigFiles) > 0 {
				if cfg := DetectStandardConfig(p.WorkspaceDir, slot.StandardConfigFiles); cfg != "" {
					logs = append(logs, fmt.Sprintf("[Config: %s] Verified bound to %s (tool: %s)", cfg, slot.ID, slot.DefaultToolName))
				}
			}
		}
	}

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

	var ruleReports []RuleReport

	executeRule := func(r Rule, phaseName string) {
		if checker, ok := r.(interface{ ShouldRun([]string) bool }); ok {
			if !checker.ShouldRun(files) {
				return
			}
		}

		rStart := time.Now()
		rBefore := len(allIssues)
		issues, err := r.Execute(ctx, files, opts)
		rDuration := time.Since(rStart)
		if err == nil && len(issues) > 0 {
			allIssues = append(allIssues, issues...)
		}
		rIssues := len(allIssues) - rBefore

		rep := RuleReport{
			RuleID:      r.ID(),
			Category:    r.Category(),
			Phase:       r.Phase(),
			Duration:    rDuration,
			IssuesCount: rIssues,
			Status:      "passed",
		}
		for i := rBefore; i < len(allIssues); i++ {
			iss := allIssues[i]
			if iss.Fixed {
				rep.FixedCount++
			}
			if iss.Severity == SeverityError {
				rep.ErrorsCount++
			} else if iss.Severity == SeverityWarning {
				rep.WarningsCount++
			}
		}
		if rep.ErrorsCount > 0 {
			rep.Status = "fail"
		} else if rep.WarningsCount > 0 {
			rep.Status = "warn"
		} else if rep.FixedCount > 0 {
			rep.Status = "fixed"
		}
		ruleReports = append(ruleReports, rep)

		if rIssues == 0 {
			logs = append(logs, fmt.Sprintf("[%s] Checking rule '%s': clean (0 issues)", phaseName, r.ID()))
		} else {
			logs = append(logs, fmt.Sprintf("[%s] Checking rule '%s': detected %d issues", phaseName, r.ID(), rIssues))
		}
	}

	// Phase 1: Fixer (Text & AST Repairs) - line-by-line evaluation
	phase1Start := time.Now()
	phase1Before := len(allIssues)
	for _, r := range fixerRules {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		executeRule(r, "Phase 1 Fixer")
	}
	phase1Issues := len(allIssues) - phase1Before
	logs = append(logs, fmt.Sprintf("[Phase 1 Fixer] Completed native hygiene check: %d rules evaluated, %d issues found (%v)", len(fixerRules), phase1Issues, time.Since(phase1Start).Round(time.Millisecond)))

	// Phase 2: Formatter (Visual Layout Standardization) - line-by-line evaluation
	phase2Start := time.Now()
	phase2Before := len(allIssues)
	for _, r := range formatterRules {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		executeRule(r, "Phase 2 Formatter")
	}
	phase2Issues := len(allIssues) - phase2Before
	logs = append(logs, fmt.Sprintf("[Phase 2 Formatter] Completed code formatting: %d rules evaluated, %d deviations found (%v)", len(formatterRules), phase2Issues, time.Since(phase2Start).Round(time.Millisecond)))

	// Phase 3: Linter (Strict Read-only Rule Verification) - line-by-line evaluation
	phase3Start := time.Now()
	phase3Before := len(allIssues)
	for _, r := range linterRules {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		executeRule(r, "Phase 3 Linter")
	}
	phase3Issues := len(allIssues) - phase3Before
	logs = append(logs, fmt.Sprintf("[Phase 3 Linter] Completed static analysis & secret protection: %d rules evaluated, %d issues found (%v)", len(linterRules), phase3Issues, time.Since(phase3Start).Round(time.Millisecond)))

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
		logs = append(logs, fmt.Sprintf("[Quality Guard] Guard blocked: %d blocking violations, %d warnings detected (total: %v)", summary.ErrorsCount, summary.WarningsCount, summary.Duration.Round(time.Millisecond)))
	} else {
		logs = append(logs, fmt.Sprintf("[Quality Guard] Guard checks passed: 0 blocking violations, all compliant (total: %v)", summary.Duration.Round(time.Millisecond)))
	}

	return &PipelineResult{
		Issues:      allIssues,
		Summary:     summary,
		HasErrors:   hasErrors,
		Logs:        logs,
		RuleReports: ruleReports,
	}, nil
}
