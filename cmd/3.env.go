// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/provider"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	// envShell selects output format for shell integration.
	envShell string

	// envLegacy prints the legacy version/build-info output.
	envLegacy bool
)

func init() {
	envCmd.Flags().StringVar(&envShell, "shell", "", "shell format (bash, zsh, fish, nu, powershell). Default: auto-detect")
	envCmd.Flags().BoolVar(&envLegacy, "info", false, "print build/version info instead of tool environment")
	if rootCmd != nil {
		rootCmd.AddCommand(envCmd)
	}
}

// envCmd outputs the shell environment for activating UniRTM tools.
var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Export shell environment variables for activated tools",
	Long: `Display or export the environment variables for the current UniRTM context.

When run in an interactive terminal, it provides a beautiful, data-rich dashboard
of your current environment, including PATH additions and variable sources.

When redirected or used with 'eval', it outputs shell-specific export statements
suitable for shell integration.

Examples:
  # Display interactive environment dashboard
  unirtm env

  # Export variables to current shell
  eval "$(unirtm env)"

  # Get environment in JSON format
  unirtm env --json`,
	Aliases: []string{"e"},
	Args:    cobra.NoArgs,
	RunE:    runEnv,
}

func runEnv(cmd *cobra.Command, args []string) error {
	// 1. Handle legacy --info mode
	if envLegacy {
		return runEnvInfoWithStyle()
	}

	cfg, _ := config.LoadFull()

	// Collect environment data
	shell := resolveShell(envShell)
	shimsDir := env.GetShimsDir()
	installsDir := env.GetInstallsDir()

	pathDirs := []string{shimsDir}
	vars := []envVarEntry{}
	var sources []string
	isRedacted := make(map[string]bool)

	// Get active tools from configuration
	toolVersions := make(map[string]string)
	if cfg != nil {
		for name, tc := range cfg.Tools {
			toolVersions[name] = tc.Version
		}
	}

	registry := provider.DefaultRegistry
	seen := make(map[string]bool)
	providerEnvVars := make(map[string]string)

	for toolNameKey, version := range toolVersions {
		toolName := toolNameKey
		backendName := ""

		if idx := strings.Index(toolNameKey, ":"); idx != -1 {
			backendName = toolNameKey[:idx]
			toolName = toolNameKey[idx+1:]
		} else if strings.Contains(toolNameKey, "/") {
			backendName = "github"
		}

		if backendName == "go" || strings.HasPrefix(toolNameKey, "go:") {
			backendName = "go-pkg"
			if strings.HasPrefix(toolNameKey, "go:") {
				toolName = strings.TrimPrefix(toolNameKey, "go:")
			}
		}

		p := registry.GetWithBackend(toolName, backendName)
		if p == nil {
			continue
		}
		fsToolName := env.GetFSToolName(toolName, backendName)
		installPath := filepath.Join(installsDir, fsToolName, version)

		// Add bin paths
		binPaths, err := p.GetBinPaths(toolName, installPath, version)
		if err == nil {
			for _, binDir := range binPaths {
				if _, statErr := os.Stat(binDir); statErr == nil && !seen[binDir] {
					pathDirs = append(pathDirs, binDir)
					seen[binDir] = true
				}
			}
		}

		// Add version variables
		varName := "UNIRTM_" + strings.ToUpper(strings.ReplaceAll(fsToolName, "-", "_")) + "_VERSION"
		vars = append(vars, envVarEntry{Name: varName, Value: version, Source: "tool:" + toolNameKey})

		// Add provider env vars
		toolEnvVars, err := p.GetEnvVars(toolName, installPath, version)
		if err == nil {
			for k, v := range toolEnvVars {
				if existing, exists := providerEnvVars[k]; !exists {
					providerEnvVars[k] = v
				} else if k == "NODE_PATH" {
					sep := string(os.PathListSeparator)
					if !strings.Contains(existing+sep, v+sep) {
						providerEnvVars[k] = existing + sep + v
					}
				}
			}
		}
	}

	for k, v := range providerEnvVars {
		vars = append(vars, envVarEntry{Name: k, Value: v, Source: "provider"})
	}

	// Load config [env] variables
	if cfg != nil {
		resolved, src, redacted, err := cfg.ResolveEnvironment()
		if err == nil {
			sources = src
			for _, rk := range redacted {
				isRedacted[rk] = true
			}

			for k, v := range resolved {
				if k == "PATH" {
					parts := strings.Split(v, string(os.PathListSeparator))
					for i := len(parts) - 1; i >= 0; i-- {
						p := parts[i]
						if p != "" && p != "$PATH" {
							pathDirs = append([]string{p}, pathDirs...)
						}
					}
					continue
				}

				val := v
				if isRedacted[k] {
					val = "[REDACTED]"
				}
				vars = append(vars, envVarEntry{Name: k, Value: val, Source: "config"})
			}
		}
	}

	// Determine if we should show the interactive TUI
	// We show the TUI ONLY if:
	// 1. Output is a terminal
	// 2. No --shell flag is provided
	// 3. No --json flag is provided
	isTerminal := term.IsTerminal(int(os.Stdout.Fd())) && !jsonOutput && envShell == ""

	if isTerminal {
		return renderInteractiveEnv(cfg, pathDirs, vars, sources)
	}

	// Scripting/Shell mode
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

	return emitShellEnv(shell, pathDirs, vars, sources)
}

func renderInteractiveEnv(cfg *config.Config, pathDirs []string, vars []envVarEntry, sources []string) error {
	// 1. Active Environment Info
	activeEnv := "base"
	if e := env.Get("ENV"); e != "" {
		activeEnv = e
	}
	pterm.DefaultSection.Printf("Context: %s\n", pterm.LightCyan(activeEnv))

	// 2. PATH Hierarchy
	pterm.DefaultSection.Println("🛤️  PATH Additions (Priority Order)")
	for i, p := range pathDirs {
		label := pterm.FgGray.Sprint("(managed)")
		if strings.Contains(p, "shims") {
			label = pterm.LightMagenta("(shims)")
		} else if strings.Contains(p, "bin") {
			label = pterm.LightBlue("(tool bin)")
		}
		prefix := "  "
		if i == 0 {
			prefix = "-> "
		}
		fmt.Printf("%s%s %s\n", prefix, p, label)
	}

	// 3. Variables Table
	pterm.DefaultSection.Println("🔑 Exported Variables")
	if len(vars) > 0 {
		var data [][]string
		data = append(data, []string{"Variable", "Value", "Source"})
		sort.Slice(vars, func(i, j int) bool { return vars[i].Name < vars[j].Name })

		for _, v := range vars {
			displayVal := v.Value
			if len(displayVal) > 50 {
				displayVal = displayVal[:47] + "..."
			}
			data = append(data, []string{
				pterm.Bold.Sprint(v.Name),
				pterm.LightCyan(displayVal),
				pterm.FgGray.Sprint(v.Source),
			})
		}
		pterm.DefaultTable.WithHasHeader().WithData(data).Render()
	} else {
		pterm.Info.Println("No additional variables exported.")
	}

	// 4. Configuration Sources
	if len(sources) > 0 {
		pterm.DefaultSection.Println("📝 Loaded Config Sources")
		for _, s := range sources {
			pterm.Success.Println(pterm.FgGray.Sprint(s))
		}
	}

	fmt.Println()
	pterm.Info.Println("To apply this environment, run: " + pterm.LightMagenta("eval \"$(unirtm env)\""))

	return nil
}

type envVarEntry struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Source string `json:"source,omitempty"`
}

func runEnvInfoWithStyle() error {
	data := [][]string{
		{"Project", env.ProjectName},
		{"Version", env.GitTag},
		{"Commit", env.CommitHash},
		{"Built", env.BuildTime},
		{"Platform", fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)},
		{"Go", runtime.Version()},
		{"License", env.LICENSE},
	}
	pterm.DefaultTable.WithData(data).Render()
	return nil
}

// emitShellEnv writes the appropriate export statements for the detected shell.
func emitShellEnv(shell string, pathDirs []string, vars []envVarEntry, sources []string) error {
	switch shell {
	case "fish":
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
		return "bash"
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
