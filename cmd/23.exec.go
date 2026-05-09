// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/spf13/cobra"
)

var (
	// execTool specifies the tool to set up before running the command
	execTool string
	// execVersion specifies the tool version to use
	execVersion string
)

// init registers the exec command to the root command.
func init() {
	execCmd.Flags().StringVarP(&execTool, "tool", "t", "", "tool to activate (e.g. node)")
	execCmd.Flags().StringVarP(&execVersion, "version", "V", "", "tool version to use (default: latest installed)")

	if rootCmd != nil {
		rootCmd.AddCommand(execCmd)
	}
}

// execCmd represents the exec command which runs a command within a tool's environment.
// Equivalent to `mise exec` / `mise run`.
var execCmd = &cobra.Command{
	Use:   "exec -- <command> [args...]",
	Short: "Execute a command with a tool's environment",
	Long: `Execute a command with a specific tool's environment variables set.

The exec command sets the UNIRTM_<TOOL>_VERSION environment variable
and then executes the given command, making the tool available via shims.

Examples:
  # Run node with a specific version
  unirtm exec --tool node --version 20.0.0 -- node --version

  # Run npm install using the active node version
  unirtm exec --tool node -- npm install

  # Run a command without specifying a tool (uses activated tools)
  unirtm exec -- make build`,
	Args:               cobra.MinimumNArgs(1),
	DisableFlagParsing: false,
	RunE:               runExec,
}

// runExec executes the exec command.
// It injects tool version environment variables and then execs the command.
func runExec(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stderr,
		Quiet:   quiet,
		Verbose: verbose,
	})

	if len(args) == 0 {
		return fmt.Errorf("no command specified: usage: unirtm exec -- <command> [args...]")
	}

	// Find the separator "--" position and split args accordingly
	commandArgs := args
	for i, a := range args {
		if a == "--" {
			commandArgs = args[i+1:]
			break
		}
	}
	if len(commandArgs) == 0 {
		return fmt.Errorf("no command specified after '--'")
	}

	if dryRun {
		parts := []string{}
		if execTool != "" && execVersion != "" {
			parts = append(parts, fmt.Sprintf("UNIRTM_%s_VERSION=%s", strings.ToUpper(execTool), execVersion))
		}
		parts = append(parts, commandArgs...)
		formatter.Info(fmt.Sprintf("[dry-run] Would run: %s", strings.Join(parts, " ")), nil)
		return nil
	}

	// Build the child command
	name := commandArgs[0]
	childArgs := commandArgs[1:]

	c := exec.Command(name, childArgs...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	// Inherit current environment
	c.Env = os.Environ()

	// Inject tool version env var if specified
	if execTool != "" && execVersion != "" {
		envKey := fmt.Sprintf("UNIRTM_%s_VERSION", strings.ToUpper(strings.ReplaceAll(execTool, "-", "_")))
		c.Env = append(c.Env, fmt.Sprintf("%s=%s", envKey, execVersion))
		if verbose {
			formatter.Info(fmt.Sprintf("Setting %s=%s", envKey, execVersion), nil)
		}
	}

	// Prepend shims dir to PATH so shims take precedence
	shimsDir := env.GetShimsDir()
	for i, e := range c.Env {
		if strings.HasPrefix(e, "PATH=") {
			c.Env[i] = fmt.Sprintf("PATH=%s:%s", shimsDir, strings.TrimPrefix(e, "PATH="))
			break
		}
	}

	if err := c.Run(); err != nil {
		// Return exit code transparently
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return fmt.Errorf("exec %s: %w", name, err)
	}
	return nil
}
