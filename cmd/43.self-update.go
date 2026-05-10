// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/spf13/cobra"
)

var (
	selfUpdateVersion string
)

func init() {
	selfUpdateCmd.Flags().StringVar(&selfUpdateVersion, "version", "", "target version to update to (default: latest)")
	if rootCmd != nil {
		rootCmd.AddCommand(selfUpdateCmd)
	}
}

// selfUpdateCmd upgrades the unirtm binary itself.
var selfUpdateCmd = &cobra.Command{
	Use:     "self-update",
	Short:   "Update UniRTM to the latest (or specified) version",
	Aliases: []string{"upgrade"},
	Long: `Update UniRTM to the latest (or specified) version.

Downloads and replaces the current binary. On Linux/macOS uses the
install script; on Windows uses PowerShell. The current binary is
backed up as unirtm.bak in the same directory.

Examples:
  # Update to latest
  unirtm self-update

  # Update to a specific version
  unirtm self-update --version 1.2.3`,
	Args: cobra.NoArgs,
	RunE: runSelfUpdate,
}

func runSelfUpdate(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	current := env.GitTag
	if current == "" {
		current = "dev"
	}

	target := selfUpdateVersion
	if target == "" {
		target = "latest"
	}

	formatter.Info(fmt.Sprintf("Current version: %s", current), nil)
	formatter.Info(fmt.Sprintf("Target version:  %s", target), nil)

	// Determine install method based on OS.
	switch runtime.GOOS {
	case "windows":
		return selfUpdateWindows(formatter, target)
	default:
		return selfUpdateUnix(formatter, target)
	}
}

func selfUpdateUnix(formatter output.Formatter, version string) error {
	// Use curl to download and run the install script.
	scriptURL := "https://github.com/snowdreamtech/unirtm/raw/main/install.sh"

	var curlArgs []string
	if version != "latest" {
		curlArgs = []string{"-fsSL", scriptURL, "|", "sh", "-s", "--", "--version", version}
	} else {
		curlArgs = []string{"-fsSL", scriptURL}
	}

	// Check if curl is available.
	if _, err := exec.LookPath("curl"); err != nil {
		formatter.Error("curl is required for self-update but was not found in PATH")
		return fmt.Errorf("curl not found")
	}

	formatter.Info("Downloading install script…", nil)

	// Build: curl -fsSL <url> | sh
	shellCmd := fmt.Sprintf("curl -fsSL %s | sh", scriptURL)
	if version != "latest" {
		shellCmd = fmt.Sprintf("curl -fsSL %s | sh -s -- --version %s", scriptURL, version)
	}

	_ = curlArgs
	c := exec.Command("sh", "-c", shellCmd)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		formatter.Error(fmt.Sprintf("Self-update failed: %v", err))
		return err
	}

	formatter.Success("UniRTM updated successfully. Restart your shell to use the new version.", nil)
	return nil
}

func selfUpdateWindows(formatter output.Formatter, version string) error {
	psScript := `irm https://github.com/snowdreamtech/unirtm/raw/main/install.ps1 | iex`
	if version != "latest" {
		psScript = fmt.Sprintf(`$env:UNIRTM_VERSION='%s'; irm https://github.com/snowdreamtech/unirtm/raw/main/install.ps1 | iex`,
			strings.ReplaceAll(version, "'", "''"))
	}

	formatter.Info("Downloading install script…", nil)
	c := exec.Command("powershell", "-ExecutionPolicy", "Bypass", "-Command", psScript)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		formatter.Error(fmt.Sprintf("Self-update failed: %v", err))
		return err
	}
	formatter.Success("UniRTM updated successfully.", nil)
	return nil
}
