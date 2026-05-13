// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bufio"
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
	formatter.Info(fmt.Sprintf("Target configuration file: %s", configFile), nil)
	formatter.Info(fmt.Sprintf("Enabling %s via: %s", targetTool, exePath), nil)

	// 3. Check if already enabled
	if exists, _ := fileExists(configFile); !exists {
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
	searchPattern := fmt.Sprintf("%s activate", targetTool)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), searchPattern) {
			alreadyEnabled = true
			break
		}
	}

	if alreadyEnabled {
		formatter.Success(fmt.Sprintf("%s is already enabled in your shell configuration.", targetTool))
		return nil
	}

	// 4. Read entire content and trim trailing whitespace to prevent accumulating newlines
	content, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	cleanContent := strings.TrimRight(string(content), " \t\r\n")

	// 5. Write back with consistent spacing
	f, err := os.OpenFile(configFile, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open config file for writing: %w", err)
	}
	defer f.Close()

	activationBlock := fmt.Sprintf("\n\n# %s shell activation\n%s\n", targetTool, activationCmd)
	if _, err := f.WriteString(cleanContent + activationBlock); err != nil {
		return fmt.Errorf("failed to write to config file: %w", err)
	}

	formatter.Success(fmt.Sprintf("Successfully enabled %s in %s", targetTool, configFile))
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
