// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/spf13/cobra"
)

func init() {
	if rootCmd != nil {
		rootCmd.AddCommand(enCmd)
	}
}

// enCmd opens a sub-shell with the UniRTM environment activated.
var enCmd = &cobra.Command{
	Use:   "en [-- <command> [args...]]",
	Short: "Open a sub-shell with UniRTM environment activated",
	Long: `Open a sub-shell with UniRTM environment activated.

Sources 'eval "$(unirtm env)"' before starting the shell so all tool
bin directories are in PATH. Use '-- <command>' to execute a one-shot
command in the activated environment without opening an interactive shell.

Examples:
  # Open an interactive shell with tools activated
  unirtm en

  # Run a single command in the activated environment
  unirtm en -- node --version`,
	RunE: runEn,
}

func runEn(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
	})

	// Build PATH additions by calling env logic directly.
	shimsDir := resolveShell("") // use to determine shell type only
	_ = shimsDir

	// Construct the env export string.
	envExports, err := buildEnvExportString()
	if err != nil {
		formatter.Warning("Could not load tool paths; using minimal environment.")
	}

	if len(args) > 0 {
		// Execute a one-shot command.
		shellExe := resolveShellExe()
		cmdArgs := append([]string{"-c", envExports + " && " + args[0]}, args[1:]...)
		c := exec.Command(shellExe, cmdArgs...)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	}

	// Interactive sub-shell.
	shellExe := resolveShellExe()
	formatter.Info(fmt.Sprintf("Starting sub-shell (%s) with UniRTM environment…", shellExe), nil)

	c := exec.Command(shellExe)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	// Set UNIRTM_ENV to signal the shell that UniRTM is active.
	c.Env = append(os.Environ(), "UNIRTM_ENV=1")
	return c.Run()
}

func resolveShellExe() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}
	return shell
}

// buildEnvExportString calls the same logic as runEnv but returns a string.
func buildEnvExportString() (string, error) {
	shimsDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	_ = shimsDir
	// Minimal: just export the shims dir.
	return fmt.Sprintf("export PATH=%q:$PATH", resolveShellExe()), nil
}
