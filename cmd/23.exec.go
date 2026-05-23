// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/backend"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/service"
	"github.com/spf13/cobra"
)

var (
	// execCommandStr holds the value of -x/--exec-command (a shell command string).
	execCommandStr string

	// execRaw pipes stdin/stdout/stderr directly without buffering (Windows only;
	// on Unix syscall.Exec already achieves this).
	execRaw bool

	// execFreshEnv forces environment recomputation, bypassing any cache.
	execFreshEnv bool

	// execNoDeps skips automatic dependency installation.
	execNoDeps bool
)

func init() {
	if rootCmd != nil {
		rootCmd.AddCommand(execCmd)
	}

	// Per-command flags — they do NOT conflict with the global -c/--config flag.
	execCmd.Flags().StringVarP(&execCommandStr, "exec-command", "x", "",
		"execute this shell command string (analogous to mise exec --command)")
	execCmd.Flags().BoolVar(&execRaw, "raw", false,
		"pipe stdin/stdout/stderr directly (sets --jobs=1)")
	execCmd.Flags().BoolVar(&execFreshEnv, "fresh-env", false,
		"bypass environment cache and recompute the environment")
	execCmd.Flags().BoolVar(&execNoDeps, "no-deps", false,
		"skip automatic dependency preparation")
}

// execCmd represents the exec command which runs a command within the tool environment.
// It is functionally equivalent to `mise exec` / `mise x`.
//
//   - Positional args before "--" are treated as tool@version specifiers.
//   - Args after "--" are the command to execute.
//   - Alternatively, use -x/--exec-command to pass a shell string.
//
// Key behaviours that match or exceed mise:
//   - Resolves tool@version overrides on top of unirtm.toml context tools.
//   - Injects bin-dirs into PATH and provider env vars (GOROOT, JAVA_HOME …).
//   - Replaces the process image via syscall.Exec on Unix (zero overhead).
//   - Falls back to os/exec on Windows with Ctrl-C forwarded to the child.
//   - Supports --dry-run, --verbose, --no-deps, --fresh-env, --raw, -x.
var execCmd = &cobra.Command{
	Use:   "exec [tool@version...] [-- <command> [args...]]",
	Short: "Execute a command with tool environment variables set",
	Long: `Execute a command with one or more tools set.

Use this to run ad-hoc commands with specific tool versions without modifying
the current shell session. Tools are loaded from the nearest unirtm.toml and
can be overridden by passing tool@version specifiers before "--".

The "--" separator distinguishes tool specifiers from the command to execute.
If "--" is omitted, all positional arguments are treated as the command itself
(loose mode, no tool overrides — compatible with mise).

Examples:
  # Run node v20 with an explicit version override
  unirtm exec node@20 -- node --version

  # Combine multiple tool overrides
  unirtm exec node@20 python@3.11 -- bash -c "node -v && python -V"

  # Shell-string mode (analogous to mise exec --command "...")
  unirtm exec node@20 -x "node --version && npm -v"

  # Use tools already declared in unirtm.toml
  unirtm exec -- make build

  # Alias: 'x'
  unirtm x node@20 -- node server.js

  # Dry-run: see what would be executed
  unirtm exec --dry-run node@20 -- node app.js`,
	Aliases: []string{"x"},
	// DisableFlagParsing is intentionally NOT set so Cobra parses -x/--raw/etc.
	// The "--" flag terminator is still honoured by Cobra natively.
	RunE: runExec,
}

// mergeEnvMaps merges src into dst.  PATH values are combined additively so
// that each tool's bin directory is prepended in order.
func mergeEnvMaps(dst, src map[string]string) {
	for k, v := range src {
		if k == "PATH" {
			if dst["PATH"] != "" {
				dst["PATH"] = v + string(os.PathListSeparator) + dst["PATH"]
			} else {
				dst["PATH"] = v
			}
		} else if k == "NODE_PATH" {
			if dst["NODE_PATH"] != "" {
				// Avoid duplicates
				sep := string(os.PathListSeparator)
				if !strings.Contains(dst["NODE_PATH"]+sep, v+sep) {
					dst["NODE_PATH"] = dst["NODE_PATH"] + sep + v
				}
			} else {
				dst["NODE_PATH"] = v
			}
		} else {
			dst[k] = v
		}
	}
}

// applyEnvMap writes all entries of envMap to the current process environment.
// PATH is prepended to the existing system value rather than overwriting it,
// ensuring host tools remain accessible.
func applyEnvMap(envMap map[string]string) {
	for k, v := range envMap {
		if k == "PATH" && v != "" {
			existing := os.Getenv("PATH")
			if existing != "" {
				os.Setenv(k, v+string(os.PathListSeparator)+existing)
			} else {
				os.Setenv(k, v)
			}
		} else if k == "NODE_PATH" && v != "" {
			existing := os.Getenv("NODE_PATH")
			if existing != "" {
				os.Setenv(k, existing+string(os.PathListSeparator)+v)
			} else {
				os.Setenv(k, v)
			}
		} else if v != "" {
			os.Setenv(k, v)
		}
	}
}

// runExec is the entry point for the exec sub-command.
func runExec(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	cfg, _ := config.LoadFull()

	// ── 1. Parse tool@version specifiers and the command to run ───────────────

	var contextTools []string
	var commandArgs []string

	if execCommandStr != "" {
		// Shell-string mode (-x): every positional arg is a tool specifier.
		contextTools = args
		commandArgs = nil // built from execCommandStr below
	} else {
		// Standard positional mode: [tool@version...] [-- command [args...]]
		// Cobra natively consumes and strips the "--" separator from the args slice.
		// We reconstruct the partition by tracing args against raw os.Args.
		separatorIdx := -1
		argsPtr := 0
		seenSubcommand := false
		for _, token := range os.Args {
			if !seenSubcommand {
				if token == "exec" || token == "x" {
					seenSubcommand = true
				}
				continue
			}
			if token == "--" {
				separatorIdx = argsPtr
				break
			}
			if argsPtr < len(args) && token == args[argsPtr] {
				argsPtr++
			}
		}

		if separatorIdx >= 0 {
			contextTools = args[:separatorIdx]
			commandArgs = args[separatorIdx:]
		} else {
			// Loose compatibility mode: no "--" → everything is the command.
			commandArgs = args
			contextTools = []string{}
		}
	}

	// ── 2. Derive the effective program and its arguments ─────────────────────

	var program string
	var programArgs []string

	if execCommandStr != "" {
		// Delegate to the user's shell.
		shell := os.Getenv("SHELL")
		if shell == "" {
			shell = "sh"
			if runtime.GOOS == "windows" {
				shell = "cmd"
			}
		}
		program = shell
		if runtime.GOOS == "windows" {
			programArgs = []string{"/c", execCommandStr}
		} else {
			programArgs = []string{"-c", execCommandStr}
		}
	} else {
		if len(commandArgs) == 0 || (len(commandArgs) == 1 && commandArgs[0] == "--") {
			return fmt.Errorf(
				"no command specified; usage: unirtm exec [tool@version...] -- <command> [args...]")
		}
		program = commandArgs[0]
		programArgs = commandArgs[1:]
	}

	// ── 3. Initialise installation manager ───────────────────────────────────

	installManager, imErr := getInstallationManager(ctx, cfg)
	if imErr != nil && verbose {
		pterm.Warning.Printf("Failed to initialise installation manager: %v\n", imErr)
	}

	// ── 4. Auto-install config tools if enabled ───────────────────────────────

	autoInstall := cfg != nil && (cfg.Settings.AutoInstall == nil || *cfg.Settings.AutoInstall)

	if installManager != nil && autoInstall && !execNoDeps && cfg != nil && len(cfg.Tools) > 0 {
		if err := installManager.EnsureInstalled(ctx, cfg.Tools); err != nil && verbose {
			pterm.Warning.Printf("Auto-install failed: %v\n", err)
		}
	}

	// ── 5. Resolve environment variables for all tools ──────────────────────

	// additionalEnv accumulates PATH extensions and tool-specific env vars.
	additionalEnv := make(map[string]string)

	if installManager != nil {
		// 5.1 First, inject tools from configuration (lower priority)
		if cfg != nil {
			for rawToolName, toolSpec := range cfg.Tools {
				// Parse the tool name using InstallationManager to correctly extract backend and tool
				parsedBackend, parsedTool, _, _ := installManager.ParseToolSpec(rawToolName)

				// Use backend defined in config, or fallback to the parsed backend
				backendName := toolSpec.Backend
				if backendName == "" {
					backendName = parsedBackend
				}

				// Resolve the version (might be a ref or alias)
				version := toolSpec.Version

				// Gather env vars (PATH, GOROOT, etc.) using the parsed tool name
				toolEnv := installManager.ResolveToolEnvBySpec(parsedTool, version, backendName)
				if len(toolEnv) > 0 {
					mergeEnvMaps(additionalEnv, toolEnv)
				}
			}
		}

		// 5.2 Second, inject context tools from command line (higher priority)
		for _, arg := range contextTools {
			backendName, toolName, version, _ := installManager.ParseToolSpec(arg)

			// Auto-install context tools when requested.
			if autoInstall && !execNoDeps {
				spec := map[string]service.ToolSpec{
					toolName: {
						Name:        toolName,
						Version:     version,
						BackendName: backendName,
					},
				}
				if err := installManager.EnsureInstalledFromSpecs(ctx, spec); err != nil && verbose {
					pterm.Warning.Printf("Failed to install context tool %s: %v\n", arg, err)
				}
			}

			// Gather env vars (overriding config tools if same variables)
			toolEnv := installManager.ResolveToolEnvBySpec(toolName, version, backendName)
			if len(toolEnv) > 0 {
				mergeEnvMaps(additionalEnv, toolEnv)
			}
		}
	}

	// ── 6. Ensure tool binaries take precedence over shims ────────────────────

	shimsDir := env.GetShimsDir()
	// Tool bins are already in additionalEnv["PATH"]. We want them to be searched FIRST.
	// Then shims, then the original system PATH.
	if existing := additionalEnv["PATH"]; existing != "" {
		additionalEnv["PATH"] = existing + string(os.PathListSeparator) + shimsDir
	} else {
		additionalEnv["PATH"] = shimsDir
	}

	// Apply all collected env overrides onto the current process.
	applyEnvMap(additionalEnv)

	// --fresh-env: remove any cached environment key so child processes
	// recompute their environments from scratch.
	if execFreshEnv {
		os.Unsetenv("UNIRTM_ENV_CACHE_KEY")
		os.Unsetenv("__MISE_ENV_CACHE_KEY")
	}

	// ── 7. Verbose diagnostic output ─────────────────────────────────────────

	if verbose {
		displayCmd := program
		if len(programArgs) > 0 {
			displayCmd += " " + strings.Join(programArgs, " ")
		}
		pterm.Info.Printf("Executing: %s\n", displayCmd)
		if len(contextTools) > 0 {
			pterm.Info.Printf("Tool context: %s\n",
				pterm.LightCyan(strings.Join(contextTools, ", ")))
		}
		if execFreshEnv {
			pterm.Info.Println("fresh-env: environment cache cleared")
		}
		if execNoDeps {
			pterm.Info.Println("no-deps: dependency installation skipped")
		}
		if execRaw {
			pterm.Info.Println("raw: direct I/O pipe active")
		}
	}

	// ── 8. Dry-run short-circuit ──────────────────────────────────────────────

	if dryRun {
		displayCmd := program
		if len(programArgs) > 0 {
			displayCmd += " " + strings.Join(programArgs, " ")
		}
		pterm.Info.Printf("[dry-run] Would execute: %s\n", displayCmd)
		if verbose && len(additionalEnv) > 0 {
			pterm.Info.Println("[dry-run] Injected environment:")
			for k, v := range additionalEnv {
				pterm.Info.Printf("  %s=%s\n", k, v)
			}
		}
		return nil
	}

	// ── 9. Resolve absolute binary path ───────────────────────────────────────

	// First, try to resolve via InstallationManager (more accurate for UniRTM tools)
	var binary string
	if installManager != nil {
		if resolved, _, err := installManager.ResolveExecutable(ctx, program, backend.CurrentPlatform()); err == nil {
			binary = resolved
		}
	}

	// If not found in UniRTM tools, fallback to system PATH (using the modified PATH)
	if binary == "" {
		var err error
		binary, err = exec.LookPath(program)
		if err != nil {
			return fmt.Errorf("command not found: %s (checked UniRTM tools and PATH)", program)
		}
	}

	// ── 10. Execute ──────────────────────────────────────────────────────────

	if runtime.GOOS != "windows" {
		// Unix: replace the current process image — zero overhead, no wrapper.
		allArgs := append([]string{binary}, programArgs...)
		return execUnix(binary, allArgs, os.Environ())
	}

	// Windows: spawn child and forward exit code.
	return execWindows(binary, programArgs)
}

// execUnix replaces the current process image via syscall.Exec (Unix only).
// On success this function never returns; the kernel replaces the process.
func execUnix(binary string, args, environ []string) error {
	if err := syscall.Exec(binary, args, environ); err != nil {
		return fmt.Errorf("syscall exec failed: %w", err)
	}
	// Unreachable on success.
	return nil
}

// execWindows runs the command as a child process on Windows and propagates
// its exit code.  Ctrl-C is forwarded to the child naturally because we do
// not install a signal handler in the parent process.
func execWindows(binary string, args []string) error {
	c := exec.Command(binary, args...)
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
