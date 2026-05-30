// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
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
	Aliases: []string{"deactive"},
	Args:    cobra.NoArgs,
	RunE:    runDeactivate,
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
	// Detect shell if not specified in flags
	shellToUse := deactivateShell
	if shellToUse == "" && len(args) > 0 {
		firstArg := strings.ToLower(args[0])
		// Check if first arg is a known shell name
		if firstArg == "bash" || firstArg == "zsh" || firstArg == "fish" || firstArg == "powershell" || firstArg == "pwsh" {
			shellToUse = firstArg
			args = args[1:] // Shift args
		}
	}

	shellType, err := resolveShellType(shellToUse)
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
	shimsDir := env.GetShimsDir()

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
# 1. Clean up shims and injected paths from PATH
_unirtm_clean_path() {
  local result=""
  local IFS=:
  for _p in $PATH; do
    case ":$UNIRTM_PATH:%s:" in
      *":$_p:"*) ;;
      *) result="${result:+$result:}$_p" ;;
    esac
  done
  echo "$result"
}
export PATH="$(_unirtm_clean_path)"
unset -f _unirtm_clean_path

# 2. Unset UniRTM environment variables
unset UNIRTM_PATH
unset UNIRTM_ACTIVATION_SCOPE
unset UNIRTM_PROJECT_DIR

# 3. Unset tool version variables (pattern UNIRTM_*_VERSION)
for var in $(env | grep '^UNIRTM_.*_VERSION=' | cut -d= -f1); do
    unset "$var"
done

# 4. Remove hook if present
unset -f unirtm
unset -f _unirtm_hook
`, shimsDir)
}

// generateFishDeactivationScript generates a fish shell deactivation script.
func generateFishDeactivationScript(shimsDir string) string {
	return fmt.Sprintf(`# UniRTM deactivation script (fish)
# 1. Clean up PATH
set -l new_path
for p in $PATH
    if not contains $p $UNIRTM_PATH; and [ "$p" != "%s" ]
        set -a new_path $p
    end
end
set -gx PATH $new_path

# 2. Unset environment variables
set -e UNIRTM_PATH
set -e UNIRTM_ACTIVATION_SCOPE
set -e UNIRTM_PROJECT_DIR

# 3. Unset tool version variables
for var in (env | grep '^UNIRTM_.*_VERSION=' | string replace -r '=.*' '')
    set -e $var
end

# 4. Remove hook
functions -e unirtm
functions -e _unirtm_hook
`, shimsDir)
}

// generatePowerShellDeactivationScript generates a PowerShell deactivation script.
func generatePowerShellDeactivationScript(shimsDir string) string {
	// Convert path separators for Windows
	if runtime.GOOS == "windows" {
		shimsDir = filepath.FromSlash(shimsDir)
	}

	return fmt.Sprintf(`# UniRTM deactivation script (PowerShell)
# 1. Clean up PATH
$unirtmPaths = @()
if ($env:UNIRTM_PATH) {
    $unirtmPaths = $env:UNIRTM_PATH -split '%c'
}
$shimsDir = "%s"
$env:PATH = ($env:PATH -split '%c' | Where-Object { $unirtmPaths -notcontains $_ -and $_ -ne $shimsDir }) -join '%c'

# 2. Unset environment variables
Remove-Item Env:\UNIRTM_PATH -ErrorAction SilentlyContinue
Remove-Item Env:\UNIRTM_ACTIVATION_SCOPE -ErrorAction SilentlyContinue
Remove-Item Env:\UNIRTM_PROJECT_DIR -ErrorAction SilentlyContinue

# 3. Unset tool version variables
Get-ChildItem Env: | Where-Object { $_.Name -match '^UNIRTM_.*_VERSION$' } | ForEach-Object {
    Remove-Item "Env:\$($_.Name)" -ErrorAction SilentlyContinue
}

# 4. Remove hook
if (Test-Path Function:\unirtm) { Remove-Item Function:\unirtm }
`, os.PathListSeparator, shimsDir, os.PathListSeparator, os.PathListSeparator)
}
