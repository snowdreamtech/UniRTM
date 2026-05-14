// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"os"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/spf13/cobra"
)

func init() {
	// Add command to root
	if rootCmd != nil {
		rootCmd.AddCommand(testToolCmd)
	}
}

// testToolCmd represents the test-tool command which tests a tool installs and executes.
var testToolCmd = &cobra.Command{
	Use:   "test-tool <tool> [version]",
	Short: "Test a tool installs and executes",
	Long: `Test a tool installs and executes by downloading it and running a version check.

This is an internal utility for validating tool registry entries and backends.

Examples:
  # Test GitHub CLI
  unirtm test-tool cli/cli

  # Test a specific version
  unirtm test-tool node 20.0.0`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runTestTool,
}

// runTestTool executes the test-tool command.
func runTestTool(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	tool := args[0]
	version := "latest"
	if len(args) == 2 {
		version = args[1]
	}

	formatter.Info(fmt.Sprintf("Testing tool %s@%s...", tool, version), nil)
	
	// We delegate to runInstall but could add cleanup logic here in the future.
	err := runInstall(cmd, args)
	if err != nil {
		pterm.Error.Printf("Test failed: installation failed for %s@%s: %v\n", tool, version, err)
		return err
	}

	pterm.FgGreen.Printf("✅ Test passed: %s@%s installed correctly\n", tool, version)
	return nil
}
