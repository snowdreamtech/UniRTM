package cmd

import (
	"context"
	"fmt"
	"os"

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
	migrateCmd.Flags().StringVarP(&migrateOutput, "output", "o", "unirtm.toml", "output file path")

	if rootCmd != nil {
		rootCmd.AddCommand(migrateCmd)
	}
}

// migrateCmd converts mise/.tool-versions config to UniRTM format.
var migrateCmd = &cobra.Command{
	Use:   "migrate [source]",
	Short: "Migrate from mise or asdf configuration",
	Long: `Convert mise or asdf configuration to UniRTM format.

Supported source formats:
  • .mise.toml     — mise configuration file
  • mise.toml      — mise configuration file
  • .tool-versions — asdf/mise version pin file

If no source file is specified, the current directory is scanned
automatically.

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
				fmt.Println(mm.FormatReport(report))
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

	// ── Human-readable output ──────────────────────────────────────────────────
	for _, report := range reports {
		fmt.Println(mm.FormatReport(report))
	}

	for _, r := range reports {
		if len(r.Errors) > 0 {
			return fmt.Errorf("migration completed with errors — check the report above")
		}
	}

	if isDryRun {
		formatter.Info("Dry-run complete — no files were written", nil)
	}

	return nil
}
