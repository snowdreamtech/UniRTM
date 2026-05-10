// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/spf13/cobra"
)

// init registers the shell command to the root command.
func init() {
	if rootCmd != nil {
		rootCmd.AddCommand(shellCmd)
	}
}

// shellCmd represents the shell command which outputs shell-specific export statements
// for a specific tool version, scoped only to the current shell session.
// Unlike `activate`, which modifies PATH permanently, `shell` emits environment
// variable exports that override the active version for the lifetime of the shell.
//
// Usage: eval "$(unirtm shell node@20.0.0)"
var shellCmd = &cobra.Command{
	Use:   "shell <tool>@<version> [tool@version ...]",
	Short: "Set tool version for the current shell session",
	Long: `Set tool version(s) for the current shell session only.

The shell command outputs shell-specific export statements that you can
eval to override the active tool version for the current shell session.
Unlike 'activate', these settings are not persisted and only affect the
current shell.

Examples:
  # Set node version for current shell session
  eval "$(unirtm shell node@20.0.0)"

  # Set multiple tools
  eval "$(unirtm shell node@20.0.0 python@3.11.0)"

  # Show what would be exported (without eval)
  unirtm shell node@20.0.0`,
	Aliases: []string{"sh"},
	Args:    cobra.RangeArgs(1, 2),
	RunE:    runShell,
}

// runShell executes the shell command.
// It outputs shell-specific export statements to stdout so they can be eval'd.
func runShell(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stderr,
		Quiet:   quiet,
		Verbose: verbose,
	})

	// Detect shell type
	shellType, err := resolveShellType("")
	if err != nil {
		shellType = "bash"
	}

	// Parse tool@version arguments
	type toolVersion struct {
		tool    string
		version string
	}
	pairs := make([]toolVersion, 0, len(args))
	for _, arg := range args {
		parts := strings.SplitN(arg, "@", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return fmt.Errorf("invalid format %q: expected <tool>@<version> (e.g. node@20.0.0)", arg)
		}
		pairs = append(pairs, toolVersion{tool: parts[0], version: parts[1]})
	}

	if dryRun {
		for _, p := range pairs {
			formatter.Info(fmt.Sprintf("[dry-run] Would export UNIRTM_%s_VERSION=%s for current shell", strings.ToUpper(p.tool), p.version), nil)
		}
		return nil
	}

	// Emit export statements based on shell type
	var sb strings.Builder
	for _, p := range pairs {
		envKey := fmt.Sprintf("UNIRTM_%s_VERSION", strings.ToUpper(strings.ReplaceAll(p.tool, "-", "_")))
		switch shellType {
		case "fish":
			sb.WriteString(fmt.Sprintf("set -gx %s %q;\n", envKey, p.version))
		case "powershell":
			sb.WriteString(fmt.Sprintf("$env:%s = %q\n", envKey, p.version))
		default: // bash, zsh
			sb.WriteString(fmt.Sprintf("export %s=%q\n", envKey, p.version))
		}
	}

	fmt.Print(sb.String())
	return nil
}
