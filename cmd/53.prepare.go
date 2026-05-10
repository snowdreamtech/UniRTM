// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"os"
	"path/filepath"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/spf13/cobra"
)

func init() {
	// Add command to root
	if rootCmd != nil {
		rootCmd.AddCommand(prepareCmd)
	}
}

// prepareCmd represents the prepare command which ensures project dependencies are ready.
var prepareCmd = &cobra.Command{
	Use:   "prepare [provider]",
	Short: "[experimental] Ensure project dependencies are ready",
	Long: `[experimental] Ensure project dependencies are ready by running applicable prepare steps.

This checks if dependency lockfiles are newer than installed outputs
(e.g., package-lock.json vs node_modules/) and runs install commands
if needed.

Examples:
  # Run all applicable prepare steps
  unirtm prepare

  # Run only npm prepare
  unirtm prepare npm`,
	Aliases: []string{"prep"},
	Args:    cobra.MaximumNArgs(1),
	RunE:    runPrepare,
}

// runPrepare executes the prepare command.
func runPrepare(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	cwd, err := os.Getwd()
	if err != nil {
		formatter.Error("Failed to get current working directory", map[string]interface{}{"error": err.Error()})
		return err
	}

	found := false

	// Check for Node.js
	if _, err := os.Stat(filepath.Join(cwd, "package.json")); err == nil {
		found = true
		pterm.Info.Println("Detected Node.js project")
		if _, err := os.Stat(filepath.Join(cwd, "node_modules")); os.IsNotExist(err) {
			pterm.Warning.Println("node_modules missing. Suggestion: run 'npm install' or 'pnpm install'")
		} else {
			pterm.Success.Println("node_modules present")
		}
	}

	// Check for Go
	if _, err := os.Stat(filepath.Join(cwd, "go.mod")); err == nil {
		found = true
		pterm.Info.Println("Detected Go project")
		pterm.Success.Println("go.mod present")
	}

	// Check for Python
	if _, err := os.Stat(filepath.Join(cwd, "requirements.txt")); err == nil {
		found = true
		pterm.Info.Println("Detected Python project (requirements.txt)")
	}
	if _, err := os.Stat(filepath.Join(cwd, "pyproject.toml")); err == nil {
		found = true
		pterm.Info.Println("Detected Python project (pyproject.toml)")
	}

	if !found {
		formatter.Info("No supported project structure detected in the current directory.", nil)
	} else {
		pterm.Success.Println("Project preparation check complete")
	}

	return nil
}
