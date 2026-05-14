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

var (
	enableAll   bool
	enableShims bool
)

// enableCmd intelligently enables UniRTM shell activation.
var enableCmd = &cobra.Command{
	Use:     "enable [unirtm|mise]",
	Aliases: []string{"en"},
	Short: "Setup UniRTM to start automatically in your shell",
	Long: `Setup UniRTM to start automatically by adding it to your shell's configuration file (like .zshrc or .bashrc).

This command auto-detects your shell and injects the activation command so you don't have to do it manually.
Once enabled, UniRTM will automatically manage your project environments whenever you open a new terminal window.

By default, it enables UniRTM, but you can specify 'mise' as an argument to enable mise instead.`,
	Args:      cobra.MaximumNArgs(1),
	ValidArgs: []string{"unirtm", "mise"},
	RunE:      runEnable,
}

func init() {
	enableCmd.Flags().BoolVarP(&enableAll, "all", "a", false, "Enable for all supported shells (zsh, bash, fish, powershell)")
	enableCmd.Flags().BoolVar(&enableShims, "shims", false, "Use shims mode for activation (default is path mode)")
	if rootCmd != nil {
		rootCmd.AddCommand(enableCmd)
	}
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

	scm := service.NewShellConfigManager(formatter, dryRun)
	shells := []service.ShellType{service.ShellZsh, service.ShellBash, service.ShellFish, service.ShellPowerShell}

	// 1. Handle --all mode
	if enableAll {
		for _, st := range shells {
			configPath, _ := scm.GetConfigPath(st)
			if _, err := os.Stat(configPath); err == nil {
				formatter.Info(fmt.Sprintf("Enabling %s for %s...", targetTool, st), nil)
				activationCmd, err := getActivationCmd(targetTool, st, enableShims)
				if err != nil {
					formatter.Warning(fmt.Sprintf("Failed to get activation command for %s: %v", st, err), nil)
					continue
				}
				if err := scm.Inject(st, targetTool, activationCmd); err != nil {
					formatter.Warning(fmt.Sprintf("Failed to enable for %s: %v", st, err), nil)
				}
			}
		}
		return nil
	}

	// 2. Detect shell
	shell, err := service.DetectShell()
	if err != nil {
		return fmt.Errorf("failed to detect shell: %w", err)
	}

	// 3. Resolve activation command
	activationCmd, err := getActivationCmd(targetTool, shell, enableShims)
	if err != nil {
		return err
	}

	formatter.Info(fmt.Sprintf("Detected shell: %s", shell), nil)

	// 4. Inject configuration
	if err := scm.Inject(shell, targetTool, activationCmd); err != nil {
		return err
	}

	configFilePath, _ := scm.GetConfigPath(shell)
	fmt.Printf("\nPlease restart your shell or run: source %s\n", configFilePath)

	return nil
}

func getActivationCmd(targetTool string, shell service.ShellType, useShims bool) (string, error) {
	// Get the absolute path of the target executable
	var exePath string
	var err error
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
	
	// Prepare flags
	flags := ""
	if useShims {
		flags = " --shims"
	}

	switch shell {
	case service.ShellZsh:
		return fmt.Sprintf(`eval "$(%s activate%s zsh)"`, exePath, flags), nil
	case service.ShellBash:
		return fmt.Sprintf(`eval "$(%s activate%s bash)"`, exePath, flags), nil
	case service.ShellFish:
		return fmt.Sprintf(`%s activate%s fish | source`, exePath, flags), nil
	case service.ShellPowerShell:
		return fmt.Sprintf(`%s activate%s powershell | Out-String | Invoke-Expression`, exePath, flags), nil
	default:
		return "", fmt.Errorf("unsupported shell: %s", shell)
	}
}
