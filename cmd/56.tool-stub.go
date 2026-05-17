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
		rootCmd.AddCommand(toolStubCmd)
	}
}

// toolStubCmd represents the tool-stub command which executes a tool stub.
var toolStubCmd = &cobra.Command{
	Use:   "tool-stub <file> [args...]",
	Short: "Execute a tool stub",
	Long: `Execute a tool stub.

Stubs are placeholders for tools that are not yet installed but are available
in the registry. When a stub is executed, UniRTM prompts to install the tool.`,
	Args: cobra.MinimumNArgs(1),
	RunE: runToolStub,
}

// runToolStub executes the tool-stub command.
func runToolStub(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	file := args[0]
	formatter.Info(fmt.Sprintf("Executing tool stub for %s...", file), nil)

	// Placeholder for actual stub logic.
	pterm.Warning.Println("Tool stub logic is currently a placeholder.")
	pterm.Info.Printf("In a full implementation, this would prompt to install the tool providing %s.\n", file)

	return nil
}
