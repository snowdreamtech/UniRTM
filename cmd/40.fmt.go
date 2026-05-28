// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/spf13/cobra"
)

var (
	fmtCheck      bool
	fmtRecursive  bool
	fmtAllConfigs bool
)

func init() {
	fmtCmd.Flags().BoolVar(&fmtCheck, "check", false, "check if config is formatted without modifying it (exit 1 if not)")
	fmtCmd.Flags().BoolVarP(&fmtRecursive, "recursive", "r", false, "format all unirtm.toml files in subdirectories")
	fmtCmd.Flags().BoolVarP(&fmtAllConfigs, "all", "a", false, "format all supported config files (unirtm.toml, .tool-versions)")

	if rootCmd != nil {
		rootCmd.AddCommand(fmtCmd)
	}
}

// fmtCmd formats unirtm.toml with canonical key ordering and indentation.
var fmtCmd = &cobra.Command{
	Use:   "fmt",
	Short: "Format configuration files with canonical key ordering",
	Long: `Format configuration files with canonical key ordering and indentation.

Reads UniRTM configuration files, normalizes section ordering, and writes them
back in-place. Use --check in CI to verify formatting without modifying files.

Canonical Section Order:
  [env] -> [tools] -> [tasks] -> [settings] -> [plugins] -> [alias]

Examples:
  # Format the project config file
  unirtm fmt

  # Format all configs in current and subdirectories
  unirtm fmt -r

  # CI mode: exit 1 if files are not formatted
  unirtm fmt --check`,
	Args: cobra.NoArgs,
	RunE: runFmt,
}

func runFmt(cmd *cobra.Command, args []string) error {
	// Transactional command: Keep it clean and quiet without the verbose header

	// pterm.SpinnerPrinter.Start() in v0.12.83 unconditionally spawns a goroutine
	// that reads IsActive in a tight loop, while Stop() writes IsActive without
	// synchronisation — a data race detected by 'go test -race'. Skip the spinner
	// entirely when running under 'go test'; use plain pterm printers instead.
	var spinner *pterm.SpinnerPrinter
	if !testing.Testing() {
		spinner, _ = pterm.DefaultSpinner.Start("Scanning configuration files...")
	}

	// spinnerWarn / spinnerSuccess are helpers that route to the spinner when
	// available, or fall back to plain pterm printers in test mode.
	spinnerWarn := func(msg string) {
		if spinner != nil {
			spinner.Warning(msg)
		} else {
			pterm.Warning.Println(msg)
		}
	}
	spinnerSuccess := func(msg string) {
		if spinner != nil {
			spinner.Success(msg)
		} else {
			pterm.FgGreen.Println(msg)
		}
	}

	var files []string
	if fmtRecursive {
		filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() {
				// Skip hidden directories and common heavy folders
				name := info.Name()
				if (strings.HasPrefix(name, ".") && name != ".") || name == "node_modules" || name == "vendor" || name == "dist" {
					return filepath.SkipDir
				}
				return nil
			}

			name := info.Name()
			if name == "unirtm.toml" || name == ".unirtm.toml" || (fmtAllConfigs && name == ".tool-versions") {
				files = append(files, path)
			}
			return nil
		})
	} else {
		cfgPath := resolveConfigFilePath(false)
		if _, err := os.Stat(cfgPath); err == nil {
			files = append(files, cfgPath)
		}
		if fmtAllConfigs {
			if _, err := os.Stat(".tool-versions"); err == nil {
				files = append(files, ".tool-versions")
			}
		}
	}

	if len(files) == 0 {
		spinnerWarn("No configuration files found to format.")
		return nil
	}

	spinnerSuccess(fmt.Sprintf("Found %d file(s)", len(files)))
	fmt.Println()

	var modifiedCount, errorCount int
	for _, path := range files {
		isModified, err := config.FormatFile(path, fmtCheck)
		if err != nil {
			pterm.Error.Prefix = pterm.Prefix{Text: "FAILED", Style: pterm.NewStyle(pterm.BgRed, pterm.FgWhite)}
			pterm.Error.Printf("%s: %v\n", path, err)
			errorCount++
			continue
		}

		if isModified {
			if fmtCheck {
				pterm.Warning.Prefix = pterm.Prefix{Text: "CHECK", Style: pterm.NewStyle(pterm.BgYellow, pterm.FgBlack)}
				pterm.Warning.Printf("%s: Needs formatting\n", path)
				modifiedCount++
			} else {
				pterm.FgGreen.Printf("%s: Formatted ✓\n", path)
				modifiedCount++
			}
		} else {
			pterm.Info.Printf("%s: Already formatted\n", path)
		}
	}

	fmt.Println()
	summary := pterm.DefaultTable.WithData(pterm.TableData{
		{"Metric", "Value"},
		{"Total Processed", fmt.Sprintf("%d", len(files))},
		{"Modified/Dirty", fmt.Sprintf("%d", modifiedCount)},
		{"Errors", fmt.Sprintf("%d", errorCount)},
	})
	summary.Render()

	if fmtCheck && modifiedCount > 0 {
		return fmt.Errorf("%d file(s) are not formatted", modifiedCount)
	}

	if errorCount > 0 {
		return fmt.Errorf("formatting completed with %d error(s)", errorCount)
	}

	return nil
}
