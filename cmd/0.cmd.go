// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
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
	// Handle asdf alias/symlink
	if filepath.Base(os.Args[0]) == "asdf" {
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

	if cfg != nil {
		im.SetAliases(cfg.Aliases)
		im.SetToolConfigs(cfg.Tools)
	}

	return im, nil
}
