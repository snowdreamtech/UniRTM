// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/backend"
	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/database"
	"github.com/snowdreamtech/unirtm/internal/pkg/download"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/provider"
	"github.com/snowdreamtech/unirtm/internal/repository/sqlite"
	"github.com/snowdreamtech/unirtm/internal/service"
	"github.com/snowdreamtech/unirtm/internal/transaction"
	"github.com/spf13/cobra"
	"go.uber.org/automaxprocs/maxprocs"
)

var (
	rootCmd *cobra.Command
)

func init() {
	// Disable automaxprocs log
	// https://github.com/uber-go/automaxprocs/issues/19#issuecomment-557382150
	nopLog := func(string, ...interface{}) {}
	maxprocs.Set(maxprocs.Logger(nopLog))
}

// Execute executes the root command.
func Execute() {
	exeName := filepath.Base(os.Args[0])
	// Handle asdf alias/symlink
	if exeName == "asdf" {
		handleAsdfAlias()
		return
	}

	// Handle shim mode: if invoked as a tool (e.g. "go") instead of "unirtm"
	if !isUniRTMBinary(exeName) {
		invokeShimMode(exeName)
		return
	}

	err := rootCmd.Execute()
	if err != nil {
		if strings.Contains(err.Error(), "unknown command") {
			target := ""
			if len(os.Args) > 1 {
				target = os.Args[len(os.Args)-1]
			}

			// Suggest from ALL commands and subcommands
			var candidates []string
			var collectCmds func(*cobra.Command)
			collectCmds = func(c *cobra.Command) {
				for _, sub := range c.Commands() {
					if sub.Hidden {
						continue
					}
					candidates = append(candidates, sub.Name())
					candidates = append(candidates, sub.Aliases...)
					collectCmds(sub)
				}
			}
			collectCmds(rootCmd)

			// De-duplicate candidates
			unique := make(map[string]struct{})
			var finalCandidates []string
			for _, c := range candidates {
				if _, ok := unique[c]; !ok {
					unique[c] = struct{}{}
					finalCandidates = append(finalCandidates, c)
				}
			}

			output.Suggest(os.Stderr, target, finalCandidates)
		}

		formatter := output.DefaultFormatter()
		formatter.Error(err.Error(), nil)
		os.Exit(1)
	}
}

// isUniRTMBinary checks if the given name is one of UniRTM's own binary names.
func isUniRTMBinary(name string) bool {
	// Remove path and extension
	name = filepath.Base(name)
	name = strings.TrimSuffix(name, ".exe")
	return name == "unirtm" || name == "unirtm-test" || name == "main" || name == "unirtm-debug"
}

// handleAsdfAlias handles legacy asdf commands for compatibility.
func handleAsdfAlias() {
	if len(os.Args) > 1 {
		command := os.Args[1]
		switch command {
		case "reshim", "update-nodebuild", "update-ruby-build":
			// Silently succeed for these commands common in plugins
			os.Exit(0)
		default:
			// For other asdf commands, we could eventually map them to unirtm commands
			fmt.Printf("unirtm (as asdf) - intercepted %s command\n", command)
			os.Exit(0)
		}
	}
	os.Exit(0)
}

// invokeShimMode resolves and executes a tool when UniRTM is invoked via a symlink.
func invokeShimMode(exeName string) {
	ctx := context.Background()

	// 1. Load configuration to find which tool provides this executable
	cfg, err := loadConfig(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: failed to load config for shim %s: %v\n", exeName, err)
		os.Exit(1)
	}

	// 2. Get installation manager
	im, err := getInstallationManager(ctx, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: failed to initialize installation manager for shim %s: %v\n", exeName, err)
		os.Exit(1)
	}

	// 3. Resolve tool path for the executable name
	// This will look through active tools in the current directory context
	platform := backend.CurrentPlatform()
	binPath, envVars, err := im.ResolveExecutable(ctx, exeName, platform)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unirtm: '%s' is not installed or binary not found.\n", exeName)
		fmt.Fprintf(os.Stderr, "  Run: unirtm install [tool]\n")
		os.Exit(1)
	}

	// 4. Execute the target binary with all arguments
	// Merge with provided env vars
	for k, v := range envVars {
		os.Setenv(k, v)
	}

	// Filter out host environment variables that might pollute the unirtm managed tools.
	// E.g., if the user has GOROOT set for a system Go installation, it will break unirtm's Go.
	if exeName == "go" || exeName == "gofmt" {
		os.Unsetenv("GOROOT")
	}

	// Prepare for execution
	args := os.Args
	if len(args) > 0 {
		args[0] = binPath
	}

	// Use syscall.Exec for a clean handoff on Unix, or fallback on Windows
	if err := service.ExecuteBinary(binPath, args); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: failed to execute %s: %v\n", binPath, err)
		os.Exit(1)
	}
}

// loadConfig loads the project configuration hierarchy.
func loadConfig(ctx context.Context) (*config.Config, error) {
	configManager := config.NewConfigManager()
	if configPath != "" {
		return configManager.Load(ctx, configPath)
	}
	return configManager.LoadHierarchy(ctx)
}

// getFormatter returns a configured formatter based on global flags and settings.
func getFormatter(cfg *config.Config) output.Formatter {
	colorMode := "auto"
	if cfg != nil && cfg.Settings.Color != "" {
		colorMode = cfg.Settings.Color
	}

	if colorMode == "always" {
		pterm.EnableColor()
	} else if colorMode == "never" {
		pterm.DisableColor()
	}

	return output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		Color:   colorMode,
		Writer:  os.Stderr,
		Quiet:   quiet,
		Verbose: verbose,
	})
}

// getInstallationManager returns a configured installation manager.
func getInstallationManager(ctx context.Context, cfg *config.Config) (*service.InstallationManager, error) {
	// Initialize registries
	backendRegistry := backend.NewRegistry()
	providerRegistry := provider.NewRegistry()
	downloadManager := download.NewManager()

	// Setup database
	db, err := database.Open(ctx, database.Config{
		Path: env.GetDatabasePath(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create repository
	installRepo, err := sqlite.NewInstallationRepository(db.Conn())
	if err != nil {
		return nil, fmt.Errorf("failed to create installation repository: %w", err)
	}

	// Create transaction manager
	txManager := transaction.NewSQLiteTransactionManager(db.Conn())

	// Create lock service if lockfile exists
	var lockSvc *service.LockService
	lockPath := env.GetLockFilePath()
	if _, err := os.Stat(lockPath); err == nil {
		lockSvc, _ = service.NewLockService(service.LockServiceOptions{
			LockfilePath: lockPath,
		})
		if lockSvc != nil {
			lockSvc.SetBackendRegistry(backendRegistry)
		}
	}

	// Create installation manager
	var settings *config.Settings
	if cfg != nil {
		settings = &cfg.Settings
	}

	im := service.NewInstallationManagerWithLock(
		backendRegistry,
		providerRegistry,
		downloadManager,
		installRepo,
		txManager,
		lockSvc,
		settings,
	)
	// Transfer ownership of the DB to the manager so callers can close it
	// via im.Close() when the command finishes.
	im.SetDB(db)

	if cfg != nil {
		im.SetAliases(cfg.Aliases)
		im.SetToolConfigs(cfg.Tools)
	}

	return im, nil
}

func getBestEditorWithSource(cfg *config.Config) (string, string) {
	// 1. Env vars
	for _, e := range []string{"UNIRTM_EDITOR", "VISUAL", "EDITOR"} {
		if v := env.Get(e); v != "" {
			return v, "$" + e
		}
	}

	// 2. Config settings
	if cfg != nil && cfg.Settings.Editor != "" {
		return cfg.Settings.Editor, "unirtm settings"
	}

	// 3. System defaults
	var defaults []string
	if runtime.GOOS == "windows" {
		defaults = []string{"notepad", "code", "vim"}
	} else {
		defaults = []string{"vim", "vi", "nano", "code", "emacs"}
	}

	for _, d := range defaults {
		if _, err := exec.LookPath(d); err == nil {
			return d, "system default"
		}
	}

	return "vi", "fallback"
}
