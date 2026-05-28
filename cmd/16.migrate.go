// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/service"
	"github.com/spf13/cobra"
)

var (
	migrateDryRun bool
	migrateOutput string
)

// init registers the migrate command to the root command.
func init() {
	migrateCmd.Flags().BoolVar(&migrateDryRun, "dry-run", false, "preview migration without writing files")
	migrateCmd.Flags().StringVarP(&migrateOutput, "output", "o", ".unirtm.toml", "output file path")

	if rootCmd != nil {
		rootCmd.AddCommand(migrateCmd)
	}
}

// migrateCmd converts mise/.tool-versions config to UniRTM format.
var migrateCmd = &cobra.Command{
	Use:   "migrate [source]",
	Short: "Migrate from mise or asdf configuration with visual side-by-side diff",
	Long: `Convert mise or asdf configuration to UniRTM format.

Supported source formats:
  • .mise.toml     — mise configuration file
  • mise.toml      — mise configuration file
  • .tool-versions — asdf/mise version pin file

If no source file is specified, the current directory is scanned
automatically. The migration report shows a visual side-by-side
comparison of what was found in the source and what will be written.

Examples:
  unirtm migrate
  unirtm migrate .mise.toml
  unirtm migrate .tool-versions --output .unirtm.toml
  unirtm migrate --dry-run`,
	Args: cobra.MaximumNArgs(1),
	RunE: runMigrate,
}

// runMigrate executes the migrate command.
//
// Validates: Requirements 21.1, 21.2, 21.3, 21.4, 21.5, 21.6, 21.7
func runMigrate(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	// Honour both the global --dry-run and the local --dry-run flag
	isDryRun := dryRun || migrateDryRun

	ctx := context.Background()
	mm := service.NewMigrationManager()

	var reports []*service.MigrationReport

	if len(args) == 1 {
		// Explicit source file provided
		report, migrateErr := mm.MigrateFile(ctx, args[0], migrateOutput, isDryRun)
		if migrateErr != nil {
			formatter.Error("Migration failed", map[string]interface{}{
				"source": args[0],
				"error":  migrateErr.Error(),
			})
			if report != nil {
				printMigrateReport(report, isDryRun)
			}
			return migrateErr
		}
		reports = []*service.MigrationReport{report}
	} else {
		// Auto-detect source files in the current directory
		var dirErr error
		reports, dirErr = mm.MigrateDirectory(ctx, ".", isDryRun)
		if dirErr != nil {
			formatter.Error("No source files found", map[string]interface{}{
				"error": dirErr.Error(),
			})
			fmt.Fprintln(os.Stderr, "\nTip: Specify a source file explicitly:")
			fmt.Fprintln(os.Stderr, "  unirtm migrate .mise.toml")
			fmt.Fprintln(os.Stderr, "  unirtm migrate .tool-versions")
			return dirErr
		}
	}

	// ── JSON output ────────────────────────────────────────────────────────────
	if jsonOutput {
		type jsonTool struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		}
		type jsonReport struct {
			Source    string     `json:"source"`
			Output    string     `json:"output"`
			ToolCount int        `json:"tool_count"`
			Tools     []jsonTool `json:"tools"`
			DryRun    bool       `json:"dry_run"`
			Warnings  []string   `json:"warnings,omitempty"`
			Errors    []string   `json:"errors,omitempty"`
		}
		var jsonReports []jsonReport
		for _, r := range reports {
			jt := make([]jsonTool, 0, len(r.Tools))
			for _, t := range r.Tools {
				jt = append(jt, jsonTool{Name: t.Name, Version: t.Version})
			}
			jsonReports = append(jsonReports, jsonReport{
				Source:    r.Source,
				Output:    r.OutputFile,
				ToolCount: len(r.Tools),
				Tools:     jt,
				DryRun:    r.DryRun,
				Warnings:  r.UnsupportedFields,
				Errors:    r.Errors,
			})
		}
		formatter.Success("Migration complete", map[string]interface{}{
			"reports": jsonReports,
		})
		return nil
	}

	// ── Visual pterm output ────────────────────────────────────────────────────
	for _, report := range reports {
		printMigrateReport(report, isDryRun)
	}

	for _, r := range reports {
		if len(r.Errors) > 0 {
			return fmt.Errorf("migration completed with errors — check the report above")
		}
	}

	if isDryRun {
		pterm.Info.Println("Dry-run complete — no files were written")
	}

	return nil
}

// printMigrateReport renders a rich pterm migration report with a side-by-side diff table.
func printMigrateReport(report *service.MigrationReport, isDryRun bool) {
	mode := "live"
	if isDryRun {
		mode = pterm.FgYellow.Sprint("dry-run")
	}

	pterm.DefaultSection.Printfln("Migration Report  [mode: %s]", mode)

	// ── Meta info panel ──────────────────────────────────────────────────────
	_ = pterm.DefaultTable.
		WithData(pterm.TableData{
			{"Source", pterm.FgCyan.Sprint(report.Source)},
			{"Output", pterm.FgGreen.Sprint(report.OutputFile)},
			{"Tools found", fmt.Sprintf("%d", len(report.Tools))},
		}).
		WithSeparator("  →  ").
		Render()
	fmt.Println()

	if len(report.Tools) == 0 {
		pterm.Warning.Println("No tools were found to migrate.")
		return
	}

	// ── Side-by-side diff table ───────────────────────────────────────────────
	pterm.DefaultSection.WithLevel(2).Println("Tool Mapping  (Source → UniRTM)")

	tableData := pterm.TableData{
		{"#", "TOOL", "SOURCE VERSION", "→", "UNIRTM FORMAT", "BACKEND"},
	}
	for i, t := range report.Tools {
		sourceEntry := fmt.Sprintf("%s = \"%s\"", t.Name, t.Version) // .tool-versions style
		if report.Source == "mise.toml" {
			sourceEntry = fmt.Sprintf("[tools.%s]\nversion = \"%s\"", t.Name, t.Version) // mise style
		}

		unirtmEntry := fmt.Sprintf("%s = \"%s\"", t.Name, t.Version)
		if t.Backend != "" || t.Provider != "" {
			unirtmEntry = fmt.Sprintf("%s.version = \"%s\"", t.Name, t.Version)
		}

		backend := t.Backend
		if backend == "" {
			backend = pterm.FgGray.Sprint("auto")
		} else {
			backend = pterm.FgMagenta.Sprint(backend)
		}

		tableData = append(tableData, []string{
			fmt.Sprintf("%d", i+1),
			pterm.FgCyan.Sprint(t.Name),
			pterm.FgYellow.Sprint(sourceEntry),
			pterm.FgGreen.Sprint("→"),
			pterm.FgGreen.Sprint(unirtmEntry),
			backend,
		})
	}

	_ = pterm.DefaultTable.
		WithHasHeader(true).
		WithSeparator("  ").
		WithHeaderStyle(pterm.NewStyle(pterm.FgLightCyan, pterm.Bold)).
		WithData(tableData).
		Render()

	// ── Warnings ──────────────────────────────────────────────────────────────
	if len(report.UnsupportedFields) > 0 {
		fmt.Println()
		pterm.DefaultSection.WithLevel(2).Println("Warnings (manual review needed)")
		for _, w := range report.UnsupportedFields {
			pterm.Warning.Println(w)
		}
	}

	// ── Errors ────────────────────────────────────────────────────────────────
	if len(report.Errors) > 0 {
		fmt.Println()
		pterm.DefaultSection.WithLevel(2).Println("Errors")
		for _, e := range report.Errors {
			pterm.Error.Println(e)
		}
		return
	}

	// ── Success summary ───────────────────────────────────────────────────────
	fmt.Println()
	if isDryRun {
		pterm.Info.Printfln("Would write %d tool(s) to %s", len(report.Tools), pterm.FgGreen.Sprint(report.OutputFile))
		fmt.Println()
		pterm.DefaultBox.
			WithTitle("Preview Output").
			WithTitleTopLeft().
			Println(buildTomlPreview(report))
	} else {
		pterm.FgGreen.Printfln(
			"Migrated %d tool(s)  →  %s",
			len(report.Tools),
			pterm.FgGreen.Sprint(report.OutputFile),
		)
		pterm.Info.Println("Review the generated file and remove any unwanted entries.")
	}
}

// buildTomlPreview builds a short TOML preview string for dry-run display.
func buildTomlPreview(report *service.MigrationReport) string {
	var sb strings.Builder
	sb.WriteString("# UniRTM configuration (preview)\n\n[tools]\n")
	for _, t := range report.Tools {
		sb.WriteString(fmt.Sprintf("  %s = %q\n", t.Name, t.Version))
	}
	return sb.String()
}
