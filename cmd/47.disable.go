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
		rootCmd.AddCommand(disableCmd)
	}
}

// disableCmd intelligently disables UniRTM shell activation.
var disableCmd = &cobra.Command{
	Use:     "disable",
	Aliases: []string{"dis"},
	Short:   "Intelligently disable UniRTM in your shell configuration",
	Long: `Intelligently disable UniRTM in your shell configuration.

This command auto-detects your current shell, identifies the corresponding
configuration file (e.g., ~/.zshrc, ~/.bashrc), and removes any UniRTM
activation commands. It is safe to run multiple times.`,
	RunE: runDisable,
}

func runDisable(cmd *cobra.Command, args []string) error {
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

	// 2. Resolve config file
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	var configFile string
	switch shell {
	case service.ShellZsh:
		configFile = filepath.Join(home, ".zshrc")
	case service.ShellBash:
		configFile = filepath.Join(home, ".bashrc")
	case service.ShellFish:
		configFile = filepath.Join(home, ".config/fish/config.fish")
	case service.ShellPowerShell:
		configFile = os.Getenv("PROFILE")
		if configFile == "" {
			configFile = filepath.Join(home, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1")
		}
	default:
		return fmt.Errorf("unsupported shell for auto-disable: %s", shell)
	}

	if exists, _ := fileExists(configFile); !exists {
		formatter.Success("UniRTM is not enabled (config file does not exist).")
		return nil
	}

	formatter.Info(fmt.Sprintf("Detected shell: %s", shell), nil)
	formatter.Info(fmt.Sprintf("Target configuration file: %s", configFile), nil)

	// 3. Read and filter config file
	file, err := os.Open(configFile)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	
	var lines []string
	removedCount := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "unirtm activate") || strings.Contains(line, "# UniRTM shell activation") {
			removedCount++
			continue
		}
		lines = append(lines, line)
	}
	file.Close()

	if removedCount == 0 {
		formatter.Success("UniRTM activation not found in your shell configuration.")
		return nil
	}

	// 4. Write back filtered content
	f, err := os.OpenFile(configFile, os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open config file for writing: %w", err)
	}
	defer f.Close()

	writer := bufio.NewWriter(f)
	for i, line := range lines {
		// Avoid leading empty lines if we removed something at the top
		if i == 0 && line == "" {
			continue
		}
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("failed to write to config file: %w", err)
		}
	}
	writer.Flush()

	formatter.Success(fmt.Sprintf("Successfully disabled UniRTM in %s (%d lines removed)", configFile, removedCount))
	fmt.Printf("\nPlease restart your shell to apply changes.\n")

	return nil
}
