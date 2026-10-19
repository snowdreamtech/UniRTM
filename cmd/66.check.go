// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/guard"
	"github.com/spf13/cobra"
)

var (
	checkAll     bool
	checkStaged  bool
	checkFix     bool
	checkProfile string
)

func init() {
	checkCmd.Flags().BoolVarP(&checkAll, "all", "a", false, "Scan all files across the entire workspace")
	checkCmd.Flags().BoolVarP(&checkStaged, "staged", "s", false, "Scan only Git staged files (pre-commit mode)")
	checkCmd.Flags().BoolVarP(&checkFix, "fix", "f", false, "Automatically fix fixable formatting and hygiene issues")
	checkCmd.Flags().StringVar(&checkProfile, "profile", "", "Select execution profile (e.g. pre-commit, ci)")

	fixCmd.Flags().BoolVarP(&checkAll, "all", "a", false, "Fix all files across the entire workspace")
	fixCmd.Flags().BoolVarP(&checkStaged, "staged", "s", false, "Fix only Git staged files")

	if rootCmd != nil {
		rootCmd.AddCommand(checkCmd)
		rootCmd.AddCommand(fixCmd)
	}
}

var checkCmd = &cobra.Command{
	Use:   "check [files...]",
	Short: "Inspect workspace files for quality, formatting, hygiene, and security issues",
	Long: `Inspect workspace files with UniRTM Quality Guard Harness.
By default, checks Git modified/staged files. Use --all for full-repository audit,
or --fix to auto-resolve fixable issues in-place.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGuard(cmd.Context(), args, checkFix)
	},
}

var fixCmd = &cobra.Command{
	Use:   "fix [files...]",
	Short: "Automatically repair and format workspace files in-place",
	Long:  `Run Phase 1 (Fixer) and Phase 2 (Formatter) to auto-resolve formatting, text hygiene, and style issues in-place.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGuard(cmd.Context(), args, true)
	},
}

func runGuard(ctx context.Context, explicitFiles []string, autoFix bool) error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	mode := "changed"
	if checkProfile == "pre-commit" {
		mode = "staged"
	} else if checkProfile == "ci" {
		mode = "all"
		autoFix = false // CI must be strict read-only
	}

	if checkAll {
		mode = "all"
	} else if checkStaged {
		mode = "staged"
	}

	walkOpts := guard.WalkOptions{
		WorkspaceDir: pwd,
		Mode:         mode,
	}

	var files []string
	if len(explicitFiles) > 0 {
		files = explicitFiles
	} else {
		files, err = guard.CollectFiles(ctx, walkOpts)
		if err != nil {
			return fmt.Errorf("failed to collect files: %w", err)
		}
	}

	if len(files) == 0 {
		output.Success("✨ No matching files found to inspect.")
		return nil
	}

	cfg := guard.GuardConfig{
		Enabled: true,
		Mode:    mode,
	}

	pipeline := guard.NewPipeline(pwd, cfg)
	res, err := pipeline.Run(ctx, files, guard.ExecuteOptions{
		Fix: autoFix,
	})
	if err != nil {
		return fmt.Errorf("guard pipeline failed: %w", err)
	}

	// Output structured reports
	if len(res.Issues) == 0 {
		output.Success(fmt.Sprintf("✨ Quality Guard passed! (%d files inspected in %v, 0 issues)",
			res.Summary.TotalFilesScanned, res.Summary.Duration.Round(100*1000))) // Round to 0.1ms
		return nil
	}

	pterm.Println()
	for _, iss := range res.Issues {
		badge := pterm.Yellow("WARN")
		if iss.Severity == guard.SeverityError {
			badge = pterm.Red("FAIL")
		} else if iss.Fixed {
			badge = pterm.Green("FIXED")
		}

		location := iss.FilePath
		if iss.Line > 0 {
			location = fmt.Sprintf("%s:%d", iss.FilePath, iss.Line)
		}

		pterm.Printf("  [%s] %s  %s: %s\n",
			badge,
			pterm.Cyan(location),
			pterm.Gray(iss.RuleID),
			iss.Message,
		)
	}
	pterm.Println()

	summaryStr := fmt.Sprintf("Guard Summary: %d files scanned, %d errors, %d warnings, %d fixed (%v)",
		res.Summary.TotalFilesScanned,
		res.Summary.ErrorsCount,
		res.Summary.WarningsCount,
		res.Summary.FixedCount,
		res.Summary.Duration.Round(100*1000),
	)

	if res.HasErrors {
		output.Error(summaryStr)
		return fmt.Errorf("quality guard check failed with %d error(s)", res.Summary.ErrorsCount)
	}

	if res.Summary.FixedCount > 0 {
		output.Success("✨ " + summaryStr)
	} else {
		output.Warning("⚠️  " + summaryStr)
	}

	return nil
}
