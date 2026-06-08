// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package cmd contains all the command-line interface definitions and implementations
// for the unirtm application. This file implements shell completion generation commands.
package cmd

import (
	"fmt"
	"io"
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
	completionAll       bool
	// completionDir is the target directory for exporting completion scripts via -d flag.
	completionDir string
)

// allShells is the canonical list of all supported shells.
var allShells = []service.ShellType{
	service.ShellZsh,
	service.ShellBash,
	service.ShellFish,
	service.ShellPowerShell,
}

// completionFileNames maps each ShellType to its output filename.
var completionFileNames = map[service.ShellType]string{
	service.ShellZsh:        "unirtm.zsh",
	service.ShellBash:       "unirtm.bash",
	service.ShellFish:       "unirtm.fish",
	service.ShellPowerShell: "unirtm.ps1",
}

// init registers the completion command and its subcommands to the root command.
func init() {
	completionCmd.Flags().BoolVarP(&completionInstall, "install", "i", false, "Intelligently install completion script to your shell configuration")
	completionCmd.Flags().BoolVarP(&completionUninstall, "uninstall", "u", false, "Intelligently uninstall completion script from your shell configuration")
	completionCmd.Flags().BoolVarP(&completionAll, "all", "a", false, "Generate/install/uninstall for all supported shells (zsh, bash, fish, powershell)")
	completionCmd.Flags().StringVarP(&completionDir, "dir", "d", "", "Export all completion scripts to the specified directory (implies --all)")
	rootCmd.AddCommand(completionCmd)
}

// completionCmd represents the completion command which generates shell completion scripts.
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate or install shell completion script",
	Long: `Generate or install shell completion script for UniRTM.

By default, it auto-detects your current shell and prints the completion script.
Use the --install (-i) flag to automatically save the script and enable it in your shell configuration.
Use the --dir (-d) flag to export all four completion scripts to a specified directory.
  NOTE: --dir only writes files; it never modifies any shell configuration.
        --dir and --install/--uninstall are mutually exclusive.

Examples:
  # Auto-detect and print to stdout
  unirtm completion

  # Auto-detect and install persistently
  unirtm completion -i

  # Generate for a specific shell and print
  unirtm completion zsh

  # Generate all scripts, install only for shells present on the system
  unirtm completion -i --all

  # Export all four completion scripts to a directory (no shell config changes)
  unirtm completion -d ./completions`,

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
		Writer:  cmd.OutOrStdout(),
		Quiet:   quiet,
	})

	// 1. Handle -d / --dir mode: export all four scripts to the specified directory.
	// -d is mutually exclusive with -i / -u: it only writes files, never modifies shell configs.
	if completionDir != "" {
		if completionInstall {
			return fmt.Errorf("--dir (-d) and --install (-i) are mutually exclusive: -d only exports scripts without modifying shell configuration")
		}
		if completionUninstall {
			return fmt.Errorf("--dir (-d) and --uninstall (-u) are mutually exclusive")
		}
		return exportCompletionsToDir(formatter, cmd, completionDir)
	}

	// 2. Handle --all mode (without -d).
	if completionAll {
		if !completionInstall && !completionUninstall {
			return fmt.Errorf("--all flag must be used with --install (-i), --uninstall (-u), or --dir (-d)")
		}

		if completionInstall {
			return installAllCompletions(formatter, cmd)
		}
		if completionUninstall {
			return uninstallAllCompletions(formatter)
		}
	}

	// 3. Single-shell mode: detect or use the provided shell argument.
	var shellType service.ShellType
	if len(args) > 0 {
		shellType = service.ShellType(args[0])
	} else {
		var err error
		shellType, err = service.DetectShell()
		if err != nil || shellType == "" {
			output.Error("Failed to detect shell. Please specify shell as argument (bash|zsh|fish|powershell)")
			return fmt.Errorf("shell detection failed")
		}
	}

	// 4. If uninstalling (single shell)
	if completionUninstall {
		spinner, _ := output.StartSpinner(fmt.Sprintf("Uninstalling completion for %s...", shellType))
		err := uninstallCompletion(formatter, shellType)
		if err != nil {
			spinner.Fail(err.Error())
		} else {
			spinner.Success(fmt.Sprintf("Completion for %s has been disabled", shellType))
		}
		return err
	}

	// 5. If not installing, just print to stdout.
	if !completionInstall {
		return generateCompletion(cmd, shellType, cmd.OutOrStdout())
	}

	// 6. Install persistently for a single shell.
	spinner, _ := output.StartSpinner(fmt.Sprintf("Installing completion for %s...", shellType))
	err := installCompletion(formatter, cmd, shellType)
	if err != nil {
		spinner.Fail(err.Error())
	} else {
		spinner.Success(fmt.Sprintf("Completion for %s is now enabled", shellType))
	}
	return err
}

// exportCompletionsToDir generates all four completion scripts and writes them to destDir.
// The directory is created if it does not exist. This is idempotent.
func exportCompletionsToDir(formatter output.Formatter, cmd *cobra.Command, destDir string) error {
	if dryRun {
		formatter.Info(fmt.Sprintf("[dry-run] Would export all completion scripts to %s", destDir), nil)
		for _, st := range allShells {
			formatter.Info(fmt.Sprintf("[dry-run] Would write %s", filepath.Join(destDir, completionFileNames[st])), nil)
		}
		return nil
	}

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", destDir, err)
	}

	var errs []error
	for _, st := range allShells {
		filename := completionFileNames[st]
		destFile := filepath.Join(destDir, filename)

		spinner, _ := output.StartSpinner(fmt.Sprintf("Generating completion for %s...", st))
		if err := writeCompletionFile(cmd, st, destFile); err != nil {
			spinner.Warning(fmt.Sprintf("Failed to generate %s: %v", filename, err))
			errs = append(errs, err)
		} else {
			spinner.Success(fmt.Sprintf("Written %s", destFile))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("some completion scripts failed to generate; see above for details")
	}
	formatter.Success(fmt.Sprintf("All completion scripts exported to %s", destDir))
	return nil
}

// installAllCompletions generates all four completion scripts to the data directory and
// injects the source/activation line only into shell config files that already exist on
// the system. Missing shell configs are silently skipped (never created automatically).
func installAllCompletions(formatter output.Formatter, cmd *cobra.Command) error {
	dataDir := env.GetDataDir()
	compDir := filepath.Join(dataDir, "completions")

	if dryRun {
		formatter.Info(fmt.Sprintf("[dry-run] Would ensure completion directory: %s", compDir), nil)
	} else {
		if err := os.MkdirAll(compDir, 0755); err != nil {
			return fmt.Errorf("failed to create completions directory: %w", err)
		}
	}

	scm := service.NewShellConfigManager(formatter, dryRun)

	for _, st := range allShells {
		filename := completionFileNames[st]
		compFile := filepath.Join(compDir, filename)

		// Always generate the completion script file (all 4 shells, unconditionally).
		spinner, _ := output.StartSpinner(fmt.Sprintf("Generating completion script for %s...", st))
		if dryRun {
			formatter.Info(fmt.Sprintf("[dry-run] Would write completion script to %s", compFile), nil)
			spinner.Success(fmt.Sprintf("[dry-run] %s", compFile))
		} else {
			if err := writeCompletionFile(cmd, st, compFile); err != nil {
				spinner.Warning(fmt.Sprintf("Failed to generate completion script for %s: %v", st, err))
				continue
			}
			spinner.Success(fmt.Sprintf("Completion script written to %s", compFile))
		}

		// Only inject source line if the shell config file already exists on this system.
		configPath, err := scm.GetConfigPath(st)
		if err != nil {
			// Unsupported shell; skip quietly.
			continue
		}

		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			// Shell config not found on this system; skip source injection.
			formatter.Info(fmt.Sprintf("Skipping %s source injection (config not found: %s)", st, configPath), nil)
			continue
		}

		// Config file exists: inject the activation line.
		activationCmd := buildActivationCmd(st, compFile)
		if activationCmd == "" {
			// Fish uses a standard completion path; no source injection needed.
			continue
		}

		installSpinner, _ := output.StartSpinner(fmt.Sprintf("Injecting completion activation for %s...", st))
		if err := scm.Inject(st, "completion", activationCmd); err != nil {
			installSpinner.Warning(fmt.Sprintf("Failed to inject activation for %s: %v", st, err))
		} else {
			installSpinner.Success(fmt.Sprintf("Activated completion for %s in %s", st, configPath))
		}
	}

	return nil
}

// uninstallAllCompletions removes the completion scripts and source lines for all shells.
func uninstallAllCompletions(formatter output.Formatter) error {
	for _, st := range allShells {
		spinner, _ := output.StartSpinner(fmt.Sprintf("Uninstalling completion for %s...", st))
		if err := uninstallCompletion(formatter, st); err != nil {
			spinner.Warning(fmt.Sprintf("Failed to uninstall completion for %s: %v", st, err))
		} else {
			spinner.Success(fmt.Sprintf("Uninstalled completion for %s", st))
		}
	}
	return nil
}

// generateCompletion writes the shell completion script for the given shell to out.
func generateCompletion(cmd *cobra.Command, shellType service.ShellType, out io.Writer) error {
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

// writeCompletionFile generates a completion script for shellType and writes it to destFile.
// The file is created (or overwritten) atomically. This operation is idempotent.
func writeCompletionFile(cmd *cobra.Command, shellType service.ShellType, destFile string) error {
	if err := os.MkdirAll(filepath.Dir(destFile), 0755); err != nil {
		return fmt.Errorf("failed to create parent directory for %s: %w", destFile, err)
	}
	f, err := os.Create(destFile)
	if err != nil {
		return fmt.Errorf("failed to create completion file %s: %w", destFile, err)
	}
	defer f.Close()
	return generateCompletion(cmd, shellType, f)
}

// buildActivationCmd returns the shell-specific source/activation line for the completion file.
// Returns an empty string for shells that use a standard completion path (e.g., Fish).
func buildActivationCmd(shellType service.ShellType, compFile string) string {
	switch shellType {
	case service.ShellZsh, service.ShellBash:
		return fmt.Sprintf(`[[ -f %s ]] && source %s`, compFile, compFile)
	case service.ShellPowerShell:
		return fmt.Sprintf(`. %s`, compFile)
	case service.ShellFish:
		// Fish picks up completions from ~/.config/fish/completions/ automatically.
		return ""
	default:
		return ""
	}
}

// installCompletion installs the completion script for a single shell and injects the
// activation line into its RC file. The RC file is created if it does not exist.
func installCompletion(formatter output.Formatter, cmd *cobra.Command, shellType service.ShellType) error {
	home, _ := os.UserHomeDir()
	dataDir := env.GetDataDir()
	compDir := filepath.Join(dataDir, "completions")

	if err := os.MkdirAll(compDir, 0755); err != nil {
		return fmt.Errorf("failed to create completions directory: %w", err)
	}

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
		// Fish uses a standard completion path; place the file there directly.
		compFile = filepath.Join(home, ".config/fish/completions/unirtm.fish")
	case service.ShellPowerShell:
		compFile = filepath.Join(compDir, "unirtm.ps1")
		configFile = env.Get("PROFILE")
		if configFile == "" {
			configFile = filepath.Join(home, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1")
		}
	default:
		return fmt.Errorf("auto-install not supported for shell: %s", shellType)
	}

	// Write completion file.
	if dryRun {
		formatter.Info(fmt.Sprintf("[dry-run] Would save completion script to %s", compFile), nil)
	} else {
		if err := writeCompletionFile(cmd, shellType, compFile); err != nil {
			return err
		}
		formatter.Success(fmt.Sprintf("Completion script saved to %s", compFile))
	}

	// Update RC file if needed.
	if configFile != "" {
		activationCmd := buildActivationCmd(shellType, compFile)
		if activationCmd != "" {
			scm := service.NewShellConfigManager(formatter, dryRun)
			if err := scm.Inject(shellType, "completion", activationCmd); err != nil {
				return err
			}
		}
	}

	if dryRun {
		formatter.Success(fmt.Sprintf("[dry-run] UniRTM completion for %s is ready to be enabled.", shellType))
	} else {
		formatter.Success(fmt.Sprintf("UniRTM completion for %s is now enabled.", shellType))
		if configFile != "" {
			fmt.Printf("\nPlease restart your shell or run: source %s\n", configFile)
		}
	}

	return nil
}

// uninstallCompletion removes the completion script and its activation line for a single shell.
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
		configFile = env.Get("PROFILE")
		if configFile == "" {
			configFile = filepath.Join(home, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1")
		}
	default:
		return fmt.Errorf("auto-uninstall not supported for shell: %s", shellType)
	}

	// Remove completion file.
	if dryRun {
		formatter.Info(fmt.Sprintf("[dry-run] Would remove completion file %s", compFile), nil)
	} else {
		if err := os.Remove(compFile); err == nil {
			formatter.Success(fmt.Sprintf("Removed completion file: %s", compFile))
		} else if !os.IsNotExist(err) {
			formatter.Warning(fmt.Sprintf("Failed to remove completion file: %v", err), nil)
		}
	}

	// Update RC file if needed.
	if configFile != "" {
		scm := service.NewShellConfigManager(formatter, dryRun)
		if err := scm.Remove(shellType, "completion"); err != nil {
			return err
		}
	}

	formatter.Success(fmt.Sprintf("UniRTM completion for %s has been disabled.", shellType))
	return nil
}
