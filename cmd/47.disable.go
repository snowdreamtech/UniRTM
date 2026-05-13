// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/service"
	"github.com/spf13/cobra"
)

var (
	disableAll bool
)

// disableCmd intelligently disables UniRTM shell activation.
var disableCmd = &cobra.Command{
	Use:     "disable [unirtm|mise]",
	Aliases: []string{"dis"},
	Short:   "Intelligently disable UniRTM or mise in your shell configuration",
	Long: `Intelligently disable UniRTM or mise in your shell configuration.

This command auto-detects your current shell, identifies the corresponding
configuration file (e.g., ~/.zshrc, ~/.bashrc), and removes any activation
commands for the specified tool.

By default, it disables UniRTM, but you can specify 'mise' as an argument
to disable mise instead.`,
	Args:      cobra.MaximumNArgs(1),
	ValidArgs: []string{"unirtm", "mise"},
	RunE:      runDisable,
}

func init() {
	disableCmd.Flags().BoolVarP(&disableAll, "all", "a", false, "Disable for all supported shells (zsh, bash, fish, powershell)")
	if rootCmd != nil {
		rootCmd.AddCommand(disableCmd)
	}
}

func runDisable(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
	})

	targetTool := "unirtm"
	if len(args) > 0 {
		targetTool = strings.ToLower(args[0])
	}

	if targetTool != "unirtm" && targetTool != "mise" {
		return fmt.Errorf("unsupported tool: %s. Supported tools: unirtm, mise", targetTool)
	}

	scm := service.NewShellConfigManager(formatter, dryRun)
	shells := []service.ShellType{service.ShellZsh, service.ShellBash, service.ShellFish, service.ShellPowerShell}

	// 1. Handle --all mode
	if disableAll {
		for _, st := range shells {
			formatter.Info(fmt.Sprintf("Disabling %s for %s...", targetTool, st), nil)
			if err := scm.Remove(st, targetTool); err != nil {
				formatter.Warning(fmt.Sprintf("Failed to disable for %s: %v", st, err), nil)
			}
		}
		return nil
	}

	// 2. Detect shell
	shell, err := service.DetectShell()
	if err != nil {
		return fmt.Errorf("failed to detect shell: %w", err)
	}

	formatter.Info(fmt.Sprintf("Detected shell: %s", shell), nil)
	formatter.Info(fmt.Sprintf("Disabling %s...", targetTool), nil)

	// 3. Remove configuration
	if err := scm.Remove(shell, targetTool); err != nil {
		return err
	}

	fmt.Printf("\nPlease restart your shell to apply changes.\n")

	return nil
}
