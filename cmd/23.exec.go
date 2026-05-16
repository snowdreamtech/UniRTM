// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/service"
	"github.com/spf13/cobra"
	"runtime"
	"syscall"
)

func init() {
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
	Aliases:            []string{"x"},
	Args:               cobra.MinimumNArgs(1),
	DisableFlagParsing: true,
	RunE:               runExec,
}

// runExec executes the exec command.
// It injects tool version environment variables and then execs the command.
func runExec(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	cfg, _ := config.LoadFull()

	if len(args) == 0 {
		return fmt.Errorf("no command specified: usage: unirtm exec [tool@version] -- <command> [args...]")
	}

	// 1. Separate context tools (node@20) and command args
	contextTools := []string{}
	commandArgs := []string{}
	foundSeparator := false

	for i, arg := range args {
		if arg == "--" {
			foundSeparator = true
			commandArgs = args[i+1:]
			break
		}
		if !foundSeparator {
			contextTools = append(contextTools, arg)
		}
	}

	// Fallback: if no --, the first arg is the command (mise style is stricter but we can be flexible)
	if !foundSeparator && len(args) > 0 {
		commandArgs = args
		contextTools = []string{}
	}

	if len(commandArgs) == 0 {
		return fmt.Errorf("no command specified after '--'")
	}

	// 2. Initialize Installation Manager
	installManager, err := getInstallationManager(ctx, cfg)
	if err != nil {
		if verbose {
			pterm.Warning.Printf("Failed to initialize installation manager: %v\n", err)
		}
	}

	// 3. Auto-install missing tools if enabled
	if installManager != nil && cfg != nil && (cfg.Settings.AutoInstall == nil || *cfg.Settings.AutoInstall) {
		// Ensure tools defined in config are installed
		if err := installManager.EnsureInstalled(ctx, cfg.Tools); err != nil {
			if verbose {
				pterm.Warning.Printf("Auto-install failed: %v\n", err)
			}
		}
	}

	// 4. Resolve Environment
	// Inject specified tools into the context
	toolsToEnsure := make(map[string]service.ToolSpec)
	for _, arg := range contextTools {
		// Use centralized ToolSpec parsing
		backendName, toolName, version, _ := installManager.ParseToolSpec(arg)

		// Record the tool to ensure it's available before execution
		toolsToEnsure[toolName] = service.ToolSpec{
			Name:        toolName,
			Version:     version,
			BackendName: backendName,
		}
		
		// Map to environment variable
		envKey := fmt.Sprintf("UNIRTM_%s_VERSION", strings.ToUpper(strings.ReplaceAll(toolName, "-", "_")))
		os.Setenv(envKey, version)
	}

	// Ensure context tools are installed
	if installManager != nil && len(toolsToEnsure) > 0 && (cfg == nil || cfg.Settings.AutoInstall == nil || *cfg.Settings.AutoInstall) {
		if err := installManager.EnsureInstalledFromSpecs(ctx, toolsToEnsure); err != nil {
			if verbose {
				pterm.Warning.Printf("Failed to install context tools: %v\n", err)
			}
		}
	}

	// Ensure shims is in PATH
	shimsDir := env.GetShimsDir()
	os.Setenv("PATH", fmt.Sprintf("%s%c%s", shimsDir, os.PathListSeparator, os.Getenv("PATH")))

	// 5. Visual Feedback (Silent for execution transparency)
	if verbose {
		pterm.Info.Printf("Executing: %s\n", strings.Join(commandArgs, " "))
		if len(contextTools) > 0 {
			pterm.Info.Printf("Context: %s\n", pterm.LightCyan(strings.Join(contextTools, ", ")))
		}
	}

	// 6. Execution
	name := commandArgs[0]
	// Look up the full path of the binary
	binary, err := exec.LookPath(name)
	if err != nil {
		return fmt.Errorf("command not found: %s", name)
	}

	if dryRun {
		pterm.Info.Printf("[dry-run] Executing: %s\n", strings.Join(commandArgs, " "))
		return nil
	}

	// [Platform Specific] Unix syscall.Exec for zero-overhead
	if runtime.GOOS != "windows" {
		// Prepare env for syscall
		env := os.Environ()
		// syscall.Exec(path, argv, envv)
		err := execUnix(binary, commandArgs, env)
		if err != nil {
			return fmt.Errorf("syscall exec failed: %w", err)
		}
		return nil
	}

	// Windows fallback using os/exec
	c := exec.Command(commandArgs[0], commandArgs[1:]...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Env = os.Environ()

	if err := c.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return err
	}
	return nil
}

func execUnix(binary string, args []string, env []string) error {
	return syscall.Exec(binary, args, env)
}
