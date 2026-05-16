// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/snowdreamtech/unirtm/internal/database"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/repository/sqlite"
	"github.com/spf13/cobra"
	"context"
)

var (
	// envShell selects output format for shell integration.
	// Supported: bash, zsh, fish, nu. Default: auto-detect from $SHELL.
	envShell string

	// envLegacy prints the legacy version/build-info output instead of tool PATH.
	envLegacy bool
)

func init() {
	envCmd.Flags().StringVar(&envShell, "shell", "", "shell format (bash, zsh, fish, nu). Default: auto-detect")
	envCmd.Flags().BoolVar(&envLegacy, "info", false, "print build/version info instead of tool environment")
	rootCmd.AddCommand(envCmd)
}

// envCmd outputs the shell environment for activating UniRTM tools.
//
// Primary use (mirrors mise env):
//
//	eval "$(unirtm env)"           # bash / zsh
//	unirtm env | source            # fish
//
// Legacy use:
//
//	unirtm env --info              # print version/build info
var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Export shell environment variables for activated tools",
	Long: `Export shell environment variables for activated tools.

The primary use of 'env' is shell integration — it outputs export statements
that add tool bin directories to PATH and set version-specific variables.

  # bash / zsh
  eval "$(unirtm env)"

  # fish
  unirtm env --shell fish | source

  # JSON output (for scripting)
  unirtm env --json

Use --info to print UniRTM build/version information instead.`,
	Aliases: []string{"e"},
	Args:    cobra.NoArgs,
	RunE:    runEnv,
}

// ─── env info structs ─────────────────────────────────────────────────────────

type envVarEntry struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type envOutput struct {
	Vars    []envVarEntry `json:"vars"`
	PathAdd []string      `json:"path_add"`
}

// ─── runEnv ───────────────────────────────────────────────────────────────────

func runEnv(cmd *cobra.Command, args []string) error {
	// Legacy --info mode: print build/version info (backwards compatible).
	if envLegacy {
		return runEnvInfo()
	}

	// Detect shell format.
	shell := resolveShell(envShell)

	// Load activated tools from DB.
	ctx := context.Background()
	dbPath := env.GetDatabasePath()
	db, err := database.Open(ctx, database.Config{Path: dbPath, WALMode: true})
	if err != nil {
		// Non-fatal: output shims dir at minimum so basic shell integration works.
		return outputMinimalEnv(shell)
	}
	defer db.Close()

	installRepo, err := sqlite.NewInstallationRepository(db.Conn())
	if err != nil {
		return outputMinimalEnv(shell)
	}

	installations, err := installRepo.List(ctx)
	if err != nil {
		return outputMinimalEnv(shell)
	}

	// Collect PATH additions: shims dir first, then each install's bin dir.
	shimsDir := env.GetShimsDir()
	installsDir := env.GetInstallsDir()

	pathDirs := []string{shimsDir}
	vars := []envVarEntry{}

	// Load configuration to get [env] variables
	var sources []string
	if cfg, err := config.Load(); err == nil {
		resolved, src, redacted, err := cfg.ResolveEnvironment()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		}
		sources = src

		// Create a map for quick redacted key lookup
		isRedacted := make(map[string]bool)
		for _, rk := range redacted {
			isRedacted[rk] = true
		}

		for k, v := range resolved {
			if k == "PATH" {
				// Special handling for PATH - prepend to pathDirs
				// Split the rendered PATH and add parts that are not already in pathDirs
				parts := strings.Split(v, string(os.PathListSeparator))
				for i := len(parts) - 1; i >= 0; i-- {
					p := parts[i]
					if p != "" && p != "$PATH" {
						// Prepend to pathDirs
						pathDirs = append([]string{p}, pathDirs...)
					}
				}
				continue
			}

			val := v
			if isRedacted[k] {
				val = "[REDACTED]"
			}
			vars = append(vars, envVarEntry{Name: k, Value: val})
		}
	}

	seen := make(map[string]bool)
	for _, inst := range installations {
		binDir := filepath.Join(installsDir, inst.Tool, inst.Version, "bin")
		if _, statErr := os.Stat(binDir); statErr == nil && !seen[binDir] {
			pathDirs = append(pathDirs, binDir)
			seen[binDir] = true
		}
		// Add a canonical version variable, e.g. UNIRTM_NODE_VERSION=22.14.0
		varName := "UNIRTM_" + strings.ToUpper(strings.ReplaceAll(inst.Tool, "-", "_")) + "_VERSION"
		vars = append(vars, envVarEntry{Name: varName, Value: inst.Version})
	}

	// JSON output.
	if jsonOutput {
		out := struct {
			Vars    []envVarEntry `json:"vars"`
			PathAdd []string      `json:"path_add"`
			Sources []string      `json:"sources"`
		}{Vars: vars, PathAdd: pathDirs, Sources: sources}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(out)
	}

	// Shell-specific export statements.
	return emitShellEnv(shell, pathDirs, vars, sources)
}

// outputMinimalEnv emits the shims directory into PATH when the DB is unavailable.
func outputMinimalEnv(shell string) error {
	return emitShellEnv(shell,
		[]string{env.GetShimsDir()},
		[]envVarEntry{},
		[]string{},
	)
}

// emitShellEnv writes the appropriate export statements for the detected shell.
func emitShellEnv(shell string, pathDirs []string, vars []envVarEntry, sources []string) error {
	switch shell {
	case "fish":
		// fish uses 'set -gx'
		if len(pathDirs) > 0 {
			fmt.Printf("set -gx PATH %s $PATH;\n", strings.Join(quoteFish(pathDirs), " "))
		}
		for _, v := range vars {
			fmt.Printf("set -gx %s %q;\n", v.Name, v.Value)
		}
		for _, s := range sources {
			fmt.Printf("source %q;\n", s)
		}
	case "nu":
		// nushell uses $env.PATH
		if len(pathDirs) > 0 {
			fmt.Printf("$env.PATH = (%s ++ $env.PATH)\n",
				"["+strings.Join(quoteNu(pathDirs), " ")+"]")
		}
		for _, v := range vars {
			fmt.Printf("$env.%s = %q\n", v.Name, v.Value)
		}
		for _, s := range sources {
			fmt.Printf("source %q\n", s)
		}
	case "powershell", "pwsh":
		// powershell uses $env:VAR
		if len(pathDirs) > 0 {
			separator := ";"
			if runtime.GOOS != "windows" {
				separator = ":"
			}
			fmt.Printf("$env:PATH = %q\n",
				strings.Join(pathDirs, separator)+separator+"$env:PATH")
		}
		for _, v := range vars {
			fmt.Printf("$env:%s = %q\n", v.Name, v.Value)
		}
		for _, s := range sources {
			fmt.Printf(". %q\n", s)
		}
	default:
		// bash / zsh / posix sh
		if len(pathDirs) > 0 {
			fmt.Printf("export PATH=%q\n",
				strings.Join(pathDirs, string(os.PathListSeparator))+string(os.PathListSeparator)+"$PATH")
		}
		for _, v := range vars {
			fmt.Printf("export %s=%q\n", v.Name, v.Value)
		}
		for _, s := range sources {
			fmt.Printf("source %q\n", s)
		}
	}
	return nil
}

// resolveShell returns the canonical shell name from the flag or $SHELL.
func resolveShell(flag string) string {
	if flag != "" {
		return strings.ToLower(flag)
	}
	shellEnv := filepath.Base(env.Get("SHELL"))
	switch shellEnv {
	case "fish":
		return "fish"
	case "nu", "nushell":
		return "nu"
	case "powershell", "pwsh", "pwsh.exe", "powershell.exe":
		return "powershell"
	default:
		return "bash" // covers bash, zsh, sh
	}
}

func quoteFish(dirs []string) []string {
	out := make([]string, len(dirs))
	for i, d := range dirs {
		out[i] = fmt.Sprintf("%q", d)
	}
	return out
}

func quoteNu(dirs []string) []string {
	out := make([]string, len(dirs))
	for i, d := range dirs {
		out[i] = fmt.Sprintf("%q", d)
	}
	return out
}

// runEnvInfo prints legacy build/version information.
func runEnvInfo() error {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("ProjectName=%s\n", env.ProjectName))
	builder.WriteString(fmt.Sprintf("Author=%s\n", env.Author))
	builder.WriteString(fmt.Sprintf("BuildTime=%s\n", env.BuildTime))
	builder.WriteString(fmt.Sprintf("GitTag=%s\n", env.GitTag))
	builder.WriteString(fmt.Sprintf("CommitHash=%s\n", env.CommitHash))
	builder.WriteString(fmt.Sprintf("CommitHashFull=%s\n", env.CommitHashFull))
	builder.WriteString(fmt.Sprintf("GOOS=%s\n", runtime.GOOS))
	builder.WriteString(fmt.Sprintf("GOARCH=%s\n", runtime.GOARCH))
	builder.WriteString(fmt.Sprintf("GOVERSION=%s\n", runtime.Version()))
	builder.WriteString(fmt.Sprintf("Copyright=%s\n", env.COPYRIGHT))
	builder.WriteString(fmt.Sprintf("LICENSE=%s\n", env.LICENSE))
	fmt.Print(builder.String())
	return nil
}
