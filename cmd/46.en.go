// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/service"
	"github.com/spf13/cobra"
)

func init() {
	if rootCmd != nil {
		rootCmd.AddCommand(enCmd)
	}
}

// enCmd intelligently enables UniRTM shell activation.
var enCmd = &cobra.Command{
	Use:   "en",
	Short: "Intelligently enable UniRTM in your shell configuration",
	Long: `Intelligently enable UniRTM in your shell configuration.

This command auto-detects your current shell, identifies the corresponding
configuration file (e.g., ~/.zshrc, ~/.bashrc), and appends the necessary
activation command if it's not already present. It is idempotent and safe
 to run multiple times.`,
	RunE: runEn,
}

func runEn(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
	})

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

	switch shell {
	case service.ShellZsh:
		configFile = filepath.Join(home, ".zshrc")
		activationCmd = `eval "$(unirtm activate zsh)"`
	case service.ShellBash:
		configFile = filepath.Join(home, ".bashrc")
		activationCmd = `eval "$(unirtm activate bash)"`
	case service.ShellFish:
		configFile = filepath.Join(home, ".config/fish/config.fish")
		activationCmd = `unirtm activate fish | source`
	case service.ShellPowerShell:
		// PowerShell profile is complex, but we can try to find it
		configFile = os.Getenv("PROFILE")
		if configFile == "" {
			configFile = filepath.Join(home, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1")
		}
		activationCmd = `unirtm activate powershell | Out-String | Invoke-Expression`
	default:
		return fmt.Errorf("unsupported shell for auto-enable: %s", shell)
	}

	formatter.Info(fmt.Sprintf("Detected shell: %s", shell), nil)
	formatter.Info(fmt.Sprintf("Target configuration file: %s", configFile), nil)

	// 3. Check if already enabled
	if exists, _ := fileExists(configFile); !exists {
		// Create file if it doesn't exist (e.g. fish config might not exist yet)
		if err := os.MkdirAll(filepath.Dir(configFile), 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
		f, err := os.Create(configFile)
		if err != nil {
			return fmt.Errorf("failed to create config file: %w", err)
		}
		f.Close()
	}

	file, err := os.Open(configFile)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	alreadyEnabled := false
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "unirtm activate") {
			alreadyEnabled = true
			break
		}
	}

	if alreadyEnabled {
		formatter.Success("UniRTM is already enabled in your shell configuration.")
		return nil
	}

	// 4. Append to config file
	f, err := os.OpenFile(configFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open config file for writing: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(fmt.Sprintf("\n# UniRTM shell activation\n%s\n", activationCmd)); err != nil {
		return fmt.Errorf("failed to write to config file: %w", err)
	}

	formatter.Success(fmt.Sprintf("Successfully enabled UniRTM in %s", configFile))
	fmt.Printf("\nPlease restart your shell or run: source %s\n", configFile)

	return nil
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
