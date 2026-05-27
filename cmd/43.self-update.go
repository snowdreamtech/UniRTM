// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/spf13/cobra"
)

var (
	selfUpdateVersion string
	selfUpdateYes     bool
	execCommand       = exec.Command
)

func init() {
	selfUpdateCmd.Flags().StringVar(&selfUpdateVersion, "version", "", "target version to update to (default: latest)")
	selfUpdateCmd.Flags().BoolVarP(&selfUpdateYes, "yes", "y", false, "skip confirmation prompt")
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

Checks the latest release on GitHub, displays release notes, and
prompts before installing. On Linux/macOS uses the install script;
on Windows uses PowerShell.

Examples:
  # Update to latest
  unirtm self-update

  # Update without prompting
  unirtm self-update --yes

  # Update to a specific version
  unirtm self-update --version v1.2.3`,
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

	// Fetch release notes from GitHub API
	spinner, _ := pterm.DefaultSpinner.Start("Checking for updates...")
	releaseInfo, err := fetchGitHubRelease(target)
	if err != nil {
		spinner.Warning(fmt.Sprintf("Could not fetch release info: %v. Proceeding blindly...", err))
	} else {
		spinner.Success(fmt.Sprintf("Found %s release: %s", target, releaseInfo.TagName))

		if target == "latest" && current == releaseInfo.TagName {
			pterm.Info.Printfln("You are already using the latest version (%s).", current)
			if !selfUpdateYes {
				return nil
			}
		}

		fmt.Println()
		pterm.DefaultSection.Printfln("Release Notes for %s", releaseInfo.TagName)
		fmt.Println(pterm.FgGray.Sprint(strings.TrimSpace(releaseInfo.Body)))
		fmt.Println()
	}

	if !selfUpdateYes && !(yes) {
		confirm, err := pterm.DefaultInteractiveConfirm.WithDefaultText("Do you want to continue with the update?").Show()
		if err != nil || !confirm {
			pterm.Info.Println("Update cancelled.")
			return nil
		}
	}

	// Determine install method based on OS.
	switch runtime.GOOS {
	case "windows":
		return selfUpdateWindows(formatter, target)
	default:
		return selfUpdateUnix(formatter, target)
	}
}

type githubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	Body    string `json:"body"`
}

var fetchGitHubRelease = func(version string) (*githubRelease, error) {
	url := "https://api.github.com/repos/snowdreamtech/unirtm/releases/latest"
	if version != "latest" {
		url = fmt.Sprintf("https://api.github.com/repos/snowdreamtech/unirtm/releases/tags/%s", version)
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
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

	formatter.Info("Downloading and executing install script…", nil)

	// Build: curl -fsSL <url> | sh
	shellCmd := fmt.Sprintf("curl -fsSL %s | sh", scriptURL)
	if version != "latest" {
		shellCmd = fmt.Sprintf("curl -fsSL %s | sh -s -- --version %s", scriptURL, version)
	}

	_ = curlArgs
	c := execCommand("sh", "-c", shellCmd)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		formatter.Error("Self-update failed", map[string]interface{}{"error": err.Error()})
		return err
	}

	pterm.Success.Println("UniRTM updated successfully. Restart your shell to use the new version.")
	return nil
}

func selfUpdateWindows(formatter output.Formatter, version string) error {
	psScript := `irm https://github.com/snowdreamtech/unirtm/raw/main/install.ps1 | iex`
	if version != "latest" {
		psScript = fmt.Sprintf(`$env:UNIRTM_VERSION='%s'; irm https://github.com/snowdreamtech/unirtm/raw/main/install.ps1 | iex`,
			strings.ReplaceAll(version, "'", "''"))
	}

	formatter.Info("Downloading and executing install script…", nil)
	c := execCommand("powershell", "-ExecutionPolicy", "Bypass", "-Command", psScript)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		formatter.Error("Self-update failed", map[string]interface{}{"error": err.Error()})
		return err
	}
	pterm.Success.Println("UniRTM updated successfully. Restart your shell to use the new version.")
	return nil
}
