// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package cmd contains all the command-line interface definitions and implementations
// for the unirtm application. This file implements shell completion generation commands.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/service"
	"github.com/spf13/cobra"
)

var (
	completionInstall   bool
	completionUninstall bool
)

// init registers the completion command and its subcommands to the root command.
func init() {
	completionCmd.Flags().BoolVarP(&completionInstall, "install", "i", false, "Intelligently install completion script to your shell configuration")
	completionCmd.Flags().BoolVarP(&completionUninstall, "uninstall", "u", false, "Intelligently uninstall completion script from your shell configuration")
	rootCmd.AddCommand(completionCmd)
}

// completionCmd represents the completion command which generates shell completion scripts.
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate or install shell completion script",
	Long: `Generate or install shell completion script for UniRTM.

By default, it auto-detects your current shell and prints the completion script.
Use the --install (-i) flag to automatically save the script and enable it in your shell configuration.

Examples:
  # Auto-detect and print to stdout
  unirtm completion

  # Auto-detect and install persistently
  unirtm completion -i

  # Generate for a specific shell and print
  unirtm completion zsh`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MaximumNArgs(1),
	RunE:                  runCompletion,
}

// runCompletion generates or installs the shell completion script.
func runCompletion(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
	})

	// 1. Detect/Select shell
	var shellType service.ShellType
	if len(args) > 0 {
		shellType = service.ShellType(args[0])
	} else {
		var err error
		shellType, err = service.DetectShell()
		if err != nil {
			return fmt.Errorf("failed to detect shell: %w. Please specify shell as argument", err)
		}
	}

	// 2. If uninstalling
	if completionUninstall {
		return uninstallCompletion(formatter, shellType)
	}

	// 3. If not installing, just print to stdout
	if !completionInstall {
		return generateCompletion(cmd, shellType, os.Stdout)
	}

	// 4. Install persistently (Plan B style)
	return installCompletion(formatter, cmd, shellType)
}

func generateCompletion(cmd *cobra.Command, shellType service.ShellType, out *os.File) error {
	switch shellType {
	case service.ShellBash:
		return cmd.Root().GenBashCompletion(out)
	case service.ShellZsh:
		return cmd.Root().GenZshCompletion(out)
	case service.ShellFish:
		return cmd.Root().GenFishCompletion(out, true)
	case service.ShellPowerShell:
		return cmd.Root().GenPowerShellCompletionWithDesc(out)
	default:
		return fmt.Errorf("unsupported shell: %s", shellType)
	}
}

func installCompletion(formatter output.Formatter, cmd *cobra.Command, shellType service.ShellType) error {
	home, _ := os.UserHomeDir()
	dataDir := env.GetDataDir()
	compDir := filepath.Join(dataDir, "completions")
	
	if err := os.MkdirAll(compDir, 0755); err != nil {
		return fmt.Errorf("failed to create completions directory: %w", err)
	}

	var compFile string
	var configFile string
	var activationCmd string

	switch shellType {
	case service.ShellZsh:
		compFile = filepath.Join(compDir, "unirtm.zsh")
		configFile = filepath.Join(home, ".zshrc")
		activationCmd = fmt.Sprintf(`[[ -f %s ]] && source %s`, compFile, compFile)
	case service.ShellBash:
		compFile = filepath.Join(compDir, "unirtm.bash")
		configFile = filepath.Join(home, ".bashrc")
		activationCmd = fmt.Sprintf(`[[ -f %s ]] && source %s`, compFile, compFile)
	case service.ShellFish:
		// Fish has a standard completion path
		compFile = filepath.Join(home, ".config/fish/completions/unirtm.fish")
		// No need for activationCmd in fish if placed in standard path
	case service.ShellPowerShell:
		compFile = filepath.Join(compDir, "unirtm.ps1")
		configFile = os.Getenv("PROFILE")
		if configFile == "" {
			configFile = filepath.Join(home, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1")
		}
		activationCmd = fmt.Sprintf(`. %s`, compFile)
	default:
		return fmt.Errorf("auto-install not supported for shell: %s", shellType)
	}

	// Write completion file
	if dryRun {
		formatter.Info(fmt.Sprintf("[dry-run] Would save completion script to %s", compFile), nil)
	} else {
		f, err := os.Create(compFile)
		if err != nil {
			return fmt.Errorf("failed to create completion file: %w", err)
		}
		if err := generateCompletion(cmd, shellType, f); err != nil {
			f.Close()
			return err
		}
		f.Close()
		formatter.Success(fmt.Sprintf("Completion script saved to %s", compFile))
	}

	// Update RC file if needed
	if configFile != "" {
		scm := service.NewShellConfigManager(formatter, dryRun)
		if err := scm.Inject(shellType, "completion", activationCmd); err != nil {
			return err
		}
	}

	if dryRun {
		formatter.Success(fmt.Sprintf("[dry-run] UniRTM completion for %s is ready to be enabled.", shellType))
	} else {
		formatter.Success(fmt.Sprintf("UniRTM completion for %s is now enabled.", shellType))
		fmt.Printf("\nPlease restart your shell or run: source %s\n", configFile)
	}
	
	return nil
}

func uninstallCompletion(formatter output.Formatter, shellType service.ShellType) error {
	home, _ := os.UserHomeDir()
	dataDir := env.GetDataDir()
	compDir := filepath.Join(dataDir, "completions")
	
	var compFile string
	var configFile string

	switch shellType {
	case service.ShellZsh:
		compFile = filepath.Join(compDir, "unirtm.zsh")
		configFile = filepath.Join(home, ".zshrc")
	case service.ShellBash:
		compFile = filepath.Join(compDir, "unirtm.bash")
		configFile = filepath.Join(home, ".bashrc")
	case service.ShellFish:
		compFile = filepath.Join(home, ".config/fish/completions/unirtm.fish")
	case service.ShellPowerShell:
		compFile = filepath.Join(compDir, "unirtm.ps1")
		configFile = os.Getenv("PROFILE")
		if configFile == "" {
			configFile = filepath.Join(home, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1")
		}
	default:
		return fmt.Errorf("auto-uninstall not supported for shell: %s", shellType)
	}

	// Remove completion file
	if dryRun {
		formatter.Info(fmt.Sprintf("[dry-run] Would remove completion file %s", compFile), nil)
	} else {
		if err := os.Remove(compFile); err == nil {
			formatter.Success(fmt.Sprintf("Removed completion file: %s", compFile))
		} else if !os.IsNotExist(err) {
			formatter.Warning(fmt.Sprintf("Failed to remove completion file: %v", err), nil)
		}
	}

	// Update RC file if needed
	if configFile != "" {
		scm := service.NewShellConfigManager(formatter, dryRun)
		if err := scm.Remove(shellType, "completion"); err != nil {
			return err
		}
	}

	formatter.Success(fmt.Sprintf("UniRTM completion for %s has been disabled.", shellType))
	return nil
}
