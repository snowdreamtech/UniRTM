// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/database"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/provider"
	"github.com/snowdreamtech/unirtm/internal/repository/sqlite"
	"github.com/snowdreamtech/unirtm/internal/service"
	"github.com/spf13/cobra"
)

var (
	// activateShell specifies the shell type for activation script
	activateShell string
	// activateScope specifies the activation scope (global or project)
	activateScope string
	// activateProjectDir specifies the project directory for project-scoped activation
	activateProjectDir string
	// activateShims specifies whether to use shims instead of dynamic PATH mode
	activateShims bool
)

// init registers the activate command to the root command.
func init() {
	activateCmd.Flags().StringVarP(&activateShell, "shell", "s", "", "shell type (bash, zsh, fish, powershell) — auto-detected if not specified")
	activateCmd.Flags().StringVar(&activateScope, "scope", "global", "activation scope (global or project)")
	activateCmd.Flags().StringVar(&activateProjectDir, "project-dir", "", "project directory for project-scoped activation (default: current directory)")
	activateCmd.Flags().BoolVar(&activateShims, "shims", false, "use shims instead of dynamic PATH mode")

	if rootCmd != nil {
		rootCmd.AddCommand(activateCmd)
	}
}

// activateCmd represents the activate command which generates activation scripts.
var activateCmd = &cobra.Command{
	Use:   "activate [tool] [version]",
	Short: "Activate a tool version in the current shell",
	Long: `Activate a tool version in the current shell.

The activate command generates a shell activation script that sets up
the environment for using the specified tool version. You can source
this script directly or add it to your shell configuration.

Examples:
  # Activate all tools (generate activation script for current shell)
  eval "$(unirtm activate)"

  # Activate for a specific shell
  unirtm activate --shell bash

  # Activate with project scope
  unirtm activate --scope project --project-dir /path/to/project

  # Activate a specific tool version
  unirtm activate node 20.0.0`,
	Args: cobra.MaximumNArgs(2),
	RunE: runActivate,
}

// runActivate executes the activate command.
// It generates a shell activation script for the specified tool or all tools.
//
// Validates: Requirements 15.1, 15.2, 15.3, 23.2
func runActivate(cmd *cobra.Command, args []string) error {
	// Create output formatter
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stderr, // use stderr so that stdout can be eval'd
		Quiet:   quiet,
		Verbose: verbose,
	})

	// Initialize dependencies
	ctx := context.Background()

	// Detect shell if not specified in flags
	if activateShell == "" && len(args) > 0 {
		firstArg := strings.ToLower(args[0])
		// Check if first arg is a known shell name
		if firstArg == "bash" || firstArg == "zsh" || firstArg == "fish" || firstArg == "powershell" || firstArg == "pwsh" {
			activateShell = firstArg
			args = args[1:] // Shift args
		}
	}

	shellType, err := resolveShellType(activateShell)
	if err != nil {
		formatter.Error("Failed to detect shell", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("detect shell: %w", err)
	}

	// Resolve project directory
	projectDir := activateProjectDir
	if activateScope == "project" && projectDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			formatter.Error("Failed to get current directory", map[string]interface{}{
				"error": err.Error(),
			})
			return fmt.Errorf("get working directory: %w", err)
		}
		projectDir = wd
	}

	// Determine scope
	scope := service.ScopeGlobal
	if activateScope == "project" {
		scope = service.ScopeProject
	}

	// Initialize database to get active tool versions
	dbPath := env.GetDatabasePath()
	db, err := database.Open(ctx, database.Config{
		Path:    dbPath,
		WALMode: true,
	})
	if err != nil {
		formatter.Error("Failed to initialize database", map[string]interface{}{
			"error": err.Error(),
			"path":  dbPath,
		})
		return fmt.Errorf("initialize database: %w", err)
	}
	defer db.Close()

	installRepo, err := sqlite.NewInstallationRepository(db.Conn())
	if err != nil {
		formatter.Error("Failed to create installation repository", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("create installation repository: %w", err)
	}

	// Build tool versions map
	toolVersions := make(map[string]string)

	if len(args) == 2 {
		// Specific tool and version requested
		tool := args[0]
		version := args[1]

		// Verify installation exists
		installation, err := installRepo.FindByToolAndVersion(ctx, tool, version)
		if err != nil {
			formatter.Error(fmt.Sprintf("Tool %s@%s is not installed", tool, version), map[string]interface{}{
				"tool":    tool,
				"version": version,
				"error":   err.Error(),
			})
			return fmt.Errorf("tool %s@%s not found: %w", tool, version, err)
		}
		toolVersions[installation.Tool] = installation.Version
	} else if len(args) == 1 {
		// Only tool specified — get latest installed version
		tool := args[0]
		installations, err := installRepo.List(ctx)
		if err != nil {
			formatter.Error("Failed to list installations", map[string]interface{}{
				"error": err.Error(),
			})
			return fmt.Errorf("list installations: %w", err)
		}
		found := false
		for _, inst := range installations {
			if inst.Tool == tool {
				toolVersions[inst.Tool] = inst.Version
				found = true
				break
			}
		}
		if !found {
			formatter.Error(fmt.Sprintf("No installed version of %s found", tool), map[string]interface{}{
				"tool": tool,
			})
			return fmt.Errorf("no installed version of %s found", tool)
		}
	} else {
		// No args — activate all installed tools
		installations, err := installRepo.List(ctx)
		if err != nil {
			formatter.Error("Failed to list installations", map[string]interface{}{
				"error": err.Error(),
			})
			return fmt.Errorf("list installations: %w", err)
		}
		for _, inst := range installations {
			// Keep the latest version for each tool (last one in list wins)
			toolVersions[inst.Tool] = inst.Version
		}
	}

	// Get shims directory
	shimsDir := env.GetShimsDir()

	// Load configuration to get tools and [env] variables
	envVars := make(map[string]string)
	var sources []string
	if cfg, err := config.LoadFull(); err == nil {
		resolved, src, redacted, err := cfg.ResolveEnvironment()
		if err != nil {
			formatter.Error("Environment resolution error", map[string]interface{}{
				"error": err.Error(),
			})
		}

		// If we are activating "all" (no specific tool requested),
		// use the tools defined in the configuration.
		if len(args) == 0 {
			// Reset toolVersions to only include what's in the config
			toolVersions = make(map[string]string)
			for name, tc := range cfg.Tools {
				toolVersions[name] = tc.Version
			}
		}

		isRedacted := make(map[string]bool)
		for _, rk := range redacted {
			isRedacted[rk] = true
		}

		for k, v := range resolved {
			val := v
			if isRedacted[k] {
				val = "[REDACTED]"
			}
			envVars[k] = val
		}
		sources = src
	} else {
		// If no config found and no tools requested, we have nothing to activate
		if len(args) == 0 {
			toolVersions = make(map[string]string)
		}
	}

	// Create activation manager
	registry := provider.NewRegistry()
	activationManager := service.NewActivationManager(shimsDir, env.GetDataDir(), registry)

	// Get executable path for the hook
	exePath, err := os.Executable()

	// Populate InjectedPaths if not using shims (Env mode)
	var injectedPaths []string
	if !activateShims {
		installsDir := env.GetInstallsDir()
		for toolNameKey, version := range toolVersions {
			toolName := toolNameKey
			backendName := ""

			// Resolve backend and tool name from key if not explicit
			if idx := strings.Index(toolNameKey, ":"); idx != -1 {
				backendName = toolNameKey[:idx]
				toolName = toolNameKey[idx+1:]
			} else if strings.Contains(toolNameKey, "/") {
				backendName = "github"
			}

			// Intercept go: prefix (align with installation manager)
			if backendName == "go" || strings.HasPrefix(toolNameKey, "go:") {
				backendName = "go-pkg"
				if strings.HasPrefix(toolNameKey, "go:") {
					toolName = strings.TrimPrefix(toolNameKey, "go:")
				}
			}

			p := provider.DefaultRegistry.GetWithBackend(toolName, backendName)
			if p == nil {
				continue
			}
			fsToolName := env.GetFSToolName(toolName, backendName)
			installPath := filepath.Join(installsDir, fsToolName, version)

			binPaths, err := p.GetBinPaths(toolName, installPath, version)
			if err == nil {
				injectedPaths = append(injectedPaths, binPaths...)
			}

			// Add provider-specific environment variables (e.g., GOROOT)
			toolEnvVars, err := p.GetEnvVars(toolName, installPath, version)
			if err == nil {
				for k, v := range toolEnvVars {
					if existing, exists := envVars[k]; !exists {
						envVars[k] = v
					} else if k == "NODE_PATH" {
						// Concatenate NODE_PATH for multiple npm tools
						sep := string(os.PathListSeparator)
						if !strings.Contains(existing+sep, v+sep) {
							envVars[k] = existing + sep + v
						}
					}
				}
			}
		}
	}

	// Generate activation script
	activationConfig := service.ActivationConfig{
		Shell:         service.ShellType(shellType),
		Scope:         scope,
		ShimsDir:      shimsDir,
		ProjectDir:    projectDir,
		ToolVersions:  toolVersions,
		EnvVars:       envVars,
		Sources:       sources,
		ExePath:       exePath,
		UseShims:      activateShims,
		InjectedPaths: injectedPaths,
	}

	script, err := activationManager.GenerateActivationScript(ctx, activationConfig)
	if err != nil {
		formatter.Error("Failed to generate activation script", map[string]interface{}{
			"shell": shellType,
			"error": err.Error(),
		})
		return fmt.Errorf("generate activation script: %w", err)
	}

	// Print the activation script to stdout for eval
	fmt.Print(script.Content)

	// Print instructions to stderr
	if !quiet {
		formatter.Info(script.Instructions, nil)
	}

	return nil
}

// resolveShellType returns the shell type to use for activation.
// If shellType is empty, it auto-detects from the environment.
func resolveShellType(shellType string) (string, error) {
	if shellType != "" {
		// Normalize shell name
		switch strings.ToLower(shellType) {
		case "bash":
			return "bash", nil
		case "zsh":
			return "zsh", nil
		case "fish":
			return "fish", nil
		case "powershell", "pwsh":
			return "powershell", nil
		default:
			return "", fmt.Errorf("unsupported shell: %s (supported: bash, zsh, fish, powershell)", shellType)
		}
	}

	// Auto-detect from environment
	detected, err := service.DetectShell()
	if err != nil {
		return "bash", nil // fallback to bash
	}
	return string(detected), nil
}
