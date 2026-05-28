// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
	"github.com/spf13/cobra"
)

var (
	selfUpdateVersion string
	selfUpdateYes     bool
	execCommand       = exec.Command
)

// installMethod represents the detected installation source.
type installMethod int

const (
	installMethodUnknown  installMethod = iota
	installMethodScript                 // installed via install.sh / install.ps1
	installMethodNpm                    // npm install -g unirtm
	installMethodPip                    // pip install unirtm
	installMethodBrew                   // brew install unirtm
	installMethodScoop                  // scoop install unirtm
	installMethodChoco                  // choco install unirtm
	installMethodCargo                  // cargo install unirtm
	installMethodGo                     // go install unirtm
	installMethodNix                    // nix-env -iA nixpkgs.unirtm
	installMethodSnap                   // snap install unirtm
	installMethodAsdf                   // asdf install unirtm latest
	installMethodMacPorts               // port install unirtm
	installMethodPkgx                   // pkgx unirtm
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
	Aliases: []string{"self-upgrade", "self-up"},
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

// detectInstallMethod inspects the executable's path to infer how it was installed.
// After filepath.ToSlash, all path separators are '/', so only forward-slash patterns
// are needed — backslash patterns are dead code and must not be used.
func detectInstallMethod(execPath string) installMethod {
	// Normalize: lowercase + convert OS separators to '/' for uniform matching.
	normalized := filepath.ToSlash(strings.ToLower(execPath))

	switch {
	// npm / pnpm / yarn
	case strings.Contains(normalized, "node_modules") ||
		strings.Contains(normalized, "/npm/") ||
		strings.Contains(normalized, "/pnpm/") ||
		strings.Contains(normalized, "/yarn/"):
		return installMethodNpm

	// pip / virtualenv / conda
	case strings.Contains(normalized, "site-packages") ||
		strings.Contains(normalized, "dist-packages") ||
		strings.Contains(normalized, "/pip/") ||
		strings.Contains(normalized, "/venv/") ||
		strings.Contains(normalized, "/.venv/"):
		return installMethodPip

	// Homebrew (macOS and Linux)
	case strings.Contains(normalized, "homebrew") ||
		strings.Contains(normalized, "linuxbrew") ||
		strings.Contains(normalized, "/cellar/"):
		return installMethodBrew

	// Scoop (Windows)
	case strings.Contains(normalized, "/scoop/"):
		return installMethodScoop

	// Chocolatey (Windows)
	case strings.Contains(normalized, "/chocolatey/") ||
		strings.Contains(normalized, "/choco/"):
		return installMethodChoco

	// Nix / nixpkgs
	case strings.Contains(normalized, "/nix/store/") ||
		strings.Contains(normalized, "/.nix-profile/"):
		return installMethodNix

	// Snap (Linux)
	case strings.Contains(normalized, "/snap/") &&
		strings.Contains(normalized, "/bin/"):
		return installMethodSnap

	// asdf (version manager, cross-platform)
	case strings.Contains(normalized, "/.asdf/"):
		return installMethodAsdf

	// MacPorts (/opt/local/bin)
	case strings.Contains(normalized, "/opt/local/"):
		return installMethodMacPorts

	// pkgx (~/.pkgx/)
	case strings.Contains(normalized, "/.pkgx/"):
		return installMethodPkgx

	// Cargo (Rust toolchain)
	case strings.Contains(normalized, "/.cargo/"):
		return installMethodCargo

	// go install: verify against GOPATH/bin or default ~/go/bin to avoid false positives.
	// Do NOT rely on "/go/bin/" alone — it can match project subdirectories.
	case isGoInstall(normalized):
		return installMethodGo
	}

	return installMethodScript
}

// isGoInstall returns true when the executable path matches a known `go install` bin directory.
// It checks against $GOPATH/bin and the default ~/go/bin to avoid false positives from
// project directories that happen to contain "/go/bin/" in their path.
func isGoInstall(normalizedPath string) bool {
	// Check $GOPATH/bin
	if gopath := os.Getenv("GOPATH"); gopath != "" {
		gopathBin := filepath.ToSlash(strings.ToLower(filepath.Join(gopath, "bin")))
		if strings.HasPrefix(normalizedPath, gopathBin) {
			return true
		}
	}

	// Check default ~/go/bin (Go's default when GOPATH is not set)
	if home, err := os.UserHomeDir(); err == nil {
		defaultBin := filepath.ToSlash(strings.ToLower(filepath.Join(home, "go", "bin")))
		if strings.HasPrefix(normalizedPath, defaultBin) {
			return true
		}
	}

	return false
}

// installMethodHint returns a human-readable upgrade hint for a given install method.
func installMethodHint(method installMethod) string {
	switch method {
	case installMethodNpm:
		return "npm update -g unirtm"
	case installMethodPip:
		return "pip install --upgrade unirtm"
	case installMethodBrew:
		return "brew upgrade unirtm"
	case installMethodScoop:
		return "scoop update unirtm"
	case installMethodChoco:
		return "choco upgrade unirtm"
	case installMethodCargo:
		return "cargo install unirtm --force"
	case installMethodGo:
		return "go install github.com/snowdreamtech/unirtm@latest"
	case installMethodNix:
		return "nix-env -uA nixpkgs.unirtm"
	case installMethodSnap:
		return "snap refresh unirtm"
	case installMethodAsdf:
		return "asdf install unirtm latest && asdf global unirtm latest"
	case installMethodMacPorts:
		return "sudo port upgrade unirtm"
	case installMethodPkgx:
		return "pkgx install unirtm"
	default:
		return ""
	}
}

// normalizeVersion strips a leading 'v' or 'V' prefix for comparison.
func normalizeVersion(v string) string {
	return strings.TrimPrefix(strings.TrimPrefix(v, "v"), "V")
}

func runSelfUpdate(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	// --- Detect installation source ---
	execPath, err := os.Executable()
	if err == nil {
		// Resolve symlinks so we get the real path
		if resolved, rerr := filepath.EvalSymlinks(execPath); rerr == nil {
			execPath = resolved
		}
		method := detectInstallMethod(execPath)
		if hint := installMethodHint(method); hint != "" {
			pterm.Warning.Printfln(
				"UniRTM appears to have been installed via a package manager.\n"+
					"Running self-update may conflict with your package manager.\n\n"+
					"  👉  To upgrade safely, please run:\n\n"+
					"      %s\n\n"+
					"  Use --yes to force self-update anyway.",
				pterm.LightCyan(hint),
			)
			if !selfUpdateYes {
				return nil
			}
		}
	}

	// --- Resolve current and target versions ---
	current := env.GitTag
	if current == "" {
		current = "dev"
	}

	target := selfUpdateVersion
	if target == "" {
		target = "latest"
	}

	// --- Fetch release info ---
	spinner, _ := pterm.DefaultSpinner.Start("Checking for updates...")
	releaseInfo, fetchErr := fetchGitHubRelease(target)
	if fetchErr != nil {
		spinner.Warning(fmt.Sprintf("Could not fetch release info: %v", fetchErr))
		if !selfUpdateYes {
			pterm.Warning.Println("Use --yes to force update without version information.")
			return fmt.Errorf("fetch release info: %w", fetchErr)
		}
	} else {
		spinner.Success(fmt.Sprintf("Found release: %s", releaseInfo.TagName))

		// Version comparison with normalized tags (strip leading 'v')
		if target == "latest" && normalizeVersion(current) == normalizeVersion(releaseInfo.TagName) {
			pterm.Info.Printfln("You are already using the latest version (%s).", current)
			if !selfUpdateYes {
				return nil
			}
		}

		// Show release notes
		fmt.Println()
		pterm.DefaultSection.Printfln("Release Notes for %s", releaseInfo.TagName)
		fmt.Println(pterm.FgGray.Sprint(strings.TrimSpace(releaseInfo.Body)))
		fmt.Println()
	}

	// --- User confirmation ---
	if !selfUpdateYes && !yes {
		confirm, promptErr := pterm.DefaultInteractiveConfirm.
			WithDefaultText("Do you want to continue with the update?").
			Show()
		if promptErr != nil || !confirm {
			pterm.Info.Println("Update cancelled.")
			return nil
		}
	}

	// Resolve the resolved tag for anchoring the script URL
	resolvedTag := target
	if releaseInfo != nil {
		resolvedTag = releaseInfo.TagName
	}

	// --- Execute platform-specific update ---
	switch runtime.GOOS {
	case "windows":
		return selfUpdateWindows(formatter, resolvedTag)
	default:
		return selfUpdateUnix(formatter, resolvedTag)
	}
}

// githubRelease holds the subset of GitHub Release API fields we need.
type githubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	Body    string `json:"body"`
}

// fetchGitHubRelease retrieves release metadata from the GitHub API.
// Uses pkgHttp.NewClientWithTimeout for timeout + proxy support.
var fetchGitHubRelease = func(version string) (*githubRelease, error) {
	url := "https://api.github.com/repos/snowdreamtech/unirtm/releases/latest"
	if version != "latest" {
		// Normalize version tag for URL
		tag := version
		if !strings.HasPrefix(tag, "v") {
			tag = "v" + tag
		}
		url = fmt.Sprintf("https://api.github.com/repos/snowdreamtech/unirtm/releases/tags/%s", tag)
	}

	client := pkgHttp.NewClientWithTimeout(30 * time.Second)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "unirtm/"+env.GitTag)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned HTTP %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &release, nil
}

// downloadToTempFn is the function used to download a URL to a temp file.
// It is a variable so that tests can replace it with a mock.
var downloadToTempFn = downloadToTempImpl

// downloadToTempImpl downloads a URL into a temporary file and returns its path.
// The caller is responsible for removing the file.
func downloadToTempImpl(url, suffix string) (string, error) {
	client := pkgHttp.NewClientWithTimeout(120 * time.Second)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "unirtm/"+env.GitTag)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download returned HTTP %d", resp.StatusCode)
	}

	tmpFile, err := os.CreateTemp("", "unirtm-install-*"+suffix)
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	defer tmpFile.Close()

	n, err := io.Copy(tmpFile, resp.Body)
	if err != nil {
		_ = os.Remove(tmpFile.Name())
		return "", fmt.Errorf("write temp file: %w", err)
	}
	if n == 0 {
		_ = os.Remove(tmpFile.Name())
		return "", fmt.Errorf("downloaded script is empty (0 bytes)")
	}

	return tmpFile.Name(), nil
}

// selfUpdateUnix downloads and executes the install.sh anchored to the given tag.
func selfUpdateUnix(formatter output.Formatter, tag string) error {
	// Check if curl is available
	if _, err := exec.LookPath("curl"); err != nil {
		formatter.Error("curl is required for self-update but was not found in PATH", nil)
		return fmt.Errorf("curl not found")
	}

	// Anchor the script URL to the specific release tag (not main branch)
	scriptURL := fmt.Sprintf("https://raw.githubusercontent.com/snowdreamtech/UniRTM/%s/install.sh", tag)

	formatter.Info(fmt.Sprintf("Downloading install script for %s...", tag), nil)

	// Download to temp file first (safe: avoids curl | sh pipe risk)
	tmpScript, err := downloadToTempFn(scriptURL, ".sh")
	if err != nil {
		formatter.Error("Failed to download install script", map[string]interface{}{"error": err.Error(), "url": scriptURL})
		return fmt.Errorf("download install script: %w", err)
	}
	defer os.Remove(tmpScript)

	// Make executable
	if err := os.Chmod(tmpScript, 0o755); err != nil {
		return fmt.Errorf("chmod install script: %w", err)
	}

	// Build the command: sh <tmpscript> [--version <tag>]
	scriptArgs := []string{tmpScript}
	if tag != "latest" {
		scriptArgs = append(scriptArgs, "--version", tag)
	}

	formatter.Info("Executing install script...", nil)

	c := execCommand("sh", scriptArgs...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin

	if err := c.Run(); err != nil {
		formatter.Error("Self-update failed", map[string]interface{}{"error": err.Error()})
		return fmt.Errorf("execute install script: %w", err)
	}

	pterm.Success.Println("UniRTM updated successfully.")
	return verifySelfUpdate()
}

// selfUpdateWindows downloads and executes the install.ps1 anchored to the given tag.
func selfUpdateWindows(formatter output.Formatter, tag string) error {
	scriptURL := fmt.Sprintf("https://raw.githubusercontent.com/snowdreamtech/UniRTM/%s/install.ps1", tag)

	formatter.Info(fmt.Sprintf("Downloading install script for %s...", tag), nil)

	// Download to temp file first
	tmpScript, err := downloadToTempFn(scriptURL, ".ps1")
	if err != nil {
		formatter.Error("Failed to download install script", map[string]interface{}{"error": err.Error(), "url": scriptURL})
		return fmt.Errorf("download install script: %w", err)
	}
	defer os.Remove(tmpScript)

	// Build PowerShell args
	psArgs := []string{"-NoProfile", "-ExecutionPolicy", "Bypass", "-File", tmpScript}
	if tag != "latest" {
		psArgs = append(psArgs, "-Version", tag)
	}

	formatter.Info("Executing install script...", nil)

	c := execCommand("powershell", psArgs...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin

	if err := c.Run(); err != nil {
		formatter.Error("Self-update failed", map[string]interface{}{"error": err.Error()})
		return fmt.Errorf("execute install script: %w", err)
	}

	pterm.Success.Println("UniRTM updated successfully.")
	return verifySelfUpdate()
}

// verifySelfUpdate runs `unirtm version` to confirm the new binary is functional.
func verifySelfUpdate() error {
	pterm.Info.Println("Verifying installation...")

	// Prefer the binary at the standard install location
	candidates := []string{"unirtm"}
	if home, err := os.UserHomeDir(); err == nil {
		candidates = append([]string{
			filepath.Join(home, ".unirtm", "bin", "unirtm"),
			filepath.Join(home, ".local", "bin", "unirtm"),
			"/usr/local/bin/unirtm",
		}, candidates...)
	}

	for _, candidate := range candidates {
		if _, err := exec.LookPath(candidate); err == nil || filepath.IsAbs(candidate) {
			c := execCommand(candidate, "version")
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			if err := c.Run(); err == nil {
				return nil
			}
		}
	}

	pterm.Warning.Println("Could not verify installed version. Restart your terminal and run 'unirtm version' manually.")
	return nil
}
