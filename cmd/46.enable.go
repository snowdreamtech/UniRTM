// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/service"
	"github.com/spf13/cobra"
)

func init() {
	if rootCmd != nil {
		rootCmd.AddCommand(enableCmd)
	}
}

// enableCmd intelligently enables UniRTM shell activation.
var enableCmd = &cobra.Command{
	Use:     "enable [unirtm|mise]",
	Aliases: []string{"en"},
	Short:   "Intelligently enable UniRTM or mise in your shell configuration",
	Long: `Intelligently enable UniRTM or mise in your shell configuration.

This command auto-detects your current shell, identifies the corresponding
configuration file (e.g., ~/.zshrc, ~/.bashrc), and appends the necessary
activation command if it's not already present.

By default, it enables UniRTM, but you can specify 'mise' as an argument
to enable mise instead. This is useful for switching between the two tools.`,
	Args:      cobra.MaximumNArgs(1),
	ValidArgs: []string{"unirtm", "mise"},
	RunE:      runEnable,
}

func runEnable(cmd *cobra.Command, args []string) error {
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

	// 1. Detect shell
	shell, err := service.DetectShell()
	if err != nil {
		return fmt.Errorf("failed to detect shell: %w", err)
	}

	// 2. Resolve config file and activation command
	var configFile string
	var activationCmd string
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	// Get the absolute path of the target executable
	var exePath string
	if targetTool == "unirtm" {
		exePath, err = os.Executable()
		if err != nil {
			exePath = "unirtm"
		}
	} else {
		// Try to find mise in PATH
		exePath = "mise"
		if p, err := exec.LookPath("mise"); err == nil {
			if abs, err := filepath.Abs(p); err == nil {
				exePath = abs
			}
		}
	}

	switch shell {
	case service.ShellZsh:
		configFile = filepath.Join(home, ".zshrc")
		activationCmd = fmt.Sprintf(`eval "$(%s activate zsh)"`, exePath)
	case service.ShellBash:
		configFile = filepath.Join(home, ".bashrc")
		activationCmd = fmt.Sprintf(`eval "$(%s activate bash)"`, exePath)
	case service.ShellFish:
		configFile = filepath.Join(home, ".config/fish/config.fish")
		activationCmd = fmt.Sprintf(`%s activate fish | source`, exePath)
	case service.ShellPowerShell:
		configFile = os.Getenv("PROFILE")
		if configFile == "" {
			configFile = filepath.Join(home, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1")
		}
		activationCmd = fmt.Sprintf(`%s activate powershell | Out-String | Invoke-Expression`, exePath)
	default:
		return fmt.Errorf("unsupported shell for auto-enable: %s", shell)
	}

	formatter.Info(fmt.Sprintf("Detected shell: %s", shell), nil)
	formatter.Info(fmt.Sprintf("Enabling %s via: %s", targetTool, exePath), nil)

	// 3. Inject configuration
	scm := service.NewShellConfigManager(formatter, dryRun)
	if err := scm.Inject(shell, targetTool, activationCmd); err != nil {
		return err
	}

	configFilePath, _ := scm.GetConfigPath(shell)
	fmt.Printf("\nPlease restart your shell or run: source %s\n", configFilePath)

	return nil
}
