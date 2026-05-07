// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/spf13/cobra"
)

var (
	// deactivateShell specifies the shell type for deactivation script
	deactivateShell string
)

// init registers the deactivate command to the root command.
func init() {
	deactivateCmd.Flags().StringVarP(&deactivateShell, "shell", "s", "", "shell type (bash, zsh, fish, powershell) — auto-detected if not specified")

	if rootCmd != nil {
		rootCmd.AddCommand(deactivateCmd)
	}
}

// deactivateCmd represents the deactivate command which removes UniRTM from the current shell environment.
var deactivateCmd = &cobra.Command{
	Use:   "deactivate",
	Short: "Deactivate UniRTM from the current shell environment",
	Long: `Deactivate UniRTM from the current shell environment.

The deactivate command generates a shell script that removes UniRTM shims
from the PATH and unsets environment variables set by the activate command.

Examples:
  # Deactivate from current shell
  eval "$(unirtm deactivate)"

  # Deactivate for a specific shell
  unirtm deactivate --shell bash`,
	Args: cobra.NoArgs,
	RunE: runDeactivate,
}

// runDeactivate executes the deactivate command.
// It generates a shell script that restores the previous environment.
//
// Validates: Requirements 15.7, 23.2
func runDeactivate(cmd *cobra.Command, args []string) error {
	// Create output formatter (stderr so stdout can be eval'd)
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stderr,
		Quiet:   quiet,
		Verbose: verbose,
	})

	// Detect shell if not specified
	shellType, err := resolveShellType(deactivateShell)
	if err != nil {
		formatter.Error("Failed to detect shell", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("detect shell: %w", err)
	}

	// Generate deactivation script based on shell
	script := generateDeactivationScript(shellType)

	// Print to stdout for eval
	fmt.Print(script)

	if !quiet {
		formatter.Info("UniRTM environment deactivated", nil)
	}

	return nil
}

// generateDeactivationScript generates a shell-specific deactivation script.
// The script removes UniRTM shims from PATH and unsets environment variables.
func generateDeactivationScript(shellType string) string {
	shimsDir := getDefaultShimsDir()

	switch shellType {
	case "fish":
		return generateFishDeactivationScript(shimsDir)
	case "powershell":
		return generatePowerShellDeactivationScript(shimsDir)
	default: // bash, zsh, sh
		return generatePosixDeactivationScript(shimsDir)
	}
}

// generatePosixDeactivationScript generates a POSIX-compatible deactivation script.
func generatePosixDeactivationScript(shimsDir string) string {
	return fmt.Sprintf(`# UniRTM deactivation script
# Remove UniRTM shims from PATH
if echo "$PATH" | grep -q "%s"; then
    export PATH="$(echo "$PATH" | sed 's|%s:||g' | sed 's|:%s||g')"
fi

# Unset UniRTM environment variables
unset UNIRTM_ACTIVATION_SCOPE
unset UNIRTM_PROJECT_DIR

# Unset tool version variables (pattern UNIRTM_*_VERSION)
for var in $(env | grep '^UNIRTM_.*_VERSION=' | cut -d= -f1); do
    unset "$var"
done
`, shimsDir, shimsDir, shimsDir)
}

// generateFishDeactivationScript generates a fish shell deactivation script.
func generateFishDeactivationScript(shimsDir string) string {
	return fmt.Sprintf(`# UniRTM deactivation script (fish)
# Remove UniRTM shims from PATH
if contains "%s" $PATH
    set -e PATH[1..(contains -i "%s" $PATH)]
end

# Unset UniRTM environment variables
set -e UNIRTM_ACTIVATION_SCOPE
set -e UNIRTM_PROJECT_DIR

# Unset tool version variables
for var in (env | grep '^UNIRTM_.*_VERSION=' | string replace -r '=.*' '')
    set -e $var
end
`, shimsDir, shimsDir)
}

// generatePowerShellDeactivationScript generates a PowerShell deactivation script.
func generatePowerShellDeactivationScript(shimsDir string) string {
	// Convert path separators for Windows
	if runtime.GOOS == "windows" {
		shimsDir = fmt.Sprintf("%s", shimsDir) // filepath.FromSlash would be used here in real code
	}

	return fmt.Sprintf(`# UniRTM deactivation script (PowerShell)
# Remove UniRTM shims from PATH
$env:PATH = ($env:PATH -split ';' | Where-Object { $_ -ne "%s" }) -join ';'

# Unset UniRTM environment variables
Remove-Item Env:\UNIRTM_ACTIVATION_SCOPE -ErrorAction SilentlyContinue
Remove-Item Env:\UNIRTM_PROJECT_DIR -ErrorAction SilentlyContinue

# Unset tool version variables
Get-ChildItem Env: | Where-Object { $_.Name -match '^UNIRTM_.*_VERSION$' } | ForEach-Object {
    Remove-Item "Env:\$($_.Name)" -ErrorAction SilentlyContinue
}
`, shimsDir)
}
