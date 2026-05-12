// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/task"
	"github.com/spf13/cobra"
)

// init registers the run command to the root command.
func init() {
	if rootCmd != nil {
		rootCmd.AddCommand(runCmd)
	}
}

// runCmd represents the run command which executes a task via the routing engine.
var runCmd = &cobra.Command{
	Use:   "run <task> [args...]",
	Short: "Run a task using the multi-modal routing engine",
	Long: `Run a task using the multi-modal routing engine.

UniRTM delegates tasks to professional tools (go-task, make, just) 
if their configuration files are detected, or falls back to executing 
tasks defined in unirtm.toml.

Examples:
  # Run a build task
  unirtm run build

  # Run a task with arguments
  unirtm run test -- -v`,
	Aliases:            []string{"r"},
	Args:               cobra.MinimumNArgs(1),
	DisableFlagParsing: false,
	RunE:               runTaskCommand,
}

// runTaskCommand executes the task routing.
func runTaskCommand(cmd *cobra.Command, args []string) error {
	taskName := args[0]
	taskArgs := args[1:]

	// Separate args if there's a "--" separator, but cobra handles arguments
	// after -- properly into args depending on config, but if they put it we might need to parse.
	// Actually args[1:] is fine.

	ctx := context.Background()

	// Load configuration
	configManager := config.NewConfigManager()
	var cfg *config.Config
	var err error
	
	if configPath != "" {
		cfg, err = configManager.Load(ctx, configPath)
	} else {
		cfg, err = configManager.LoadHierarchy(ctx)
	}

	if err != nil {
		// Log error but continue with an empty config, since external runners
		// (go-task, make, just) might not need unirtm.toml to work.
		cfg = &config.Config{
			Tasks: make(map[string]config.Task),
		}
	}

	// Auto-install missing tools if enabled
	if cfg.Settings.AutoInstall == nil || *cfg.Settings.AutoInstall {
		installManager, err := getInstallationManager(ctx, cfg)
		if err == nil {
			if err := installManager.EnsureInstalled(ctx, cfg.Tools); err != nil {
				// Log warning but continue
				fmt.Fprintf(os.Stderr, "⚠ auto-install warning: %v\n", err)
			}
		}
	}

	// Setup Engine
	engine := task.NewEngine()

	// Register runners in priority order
	// NativeRunner (unirtm.toml) has highest priority if task exists there.
	engine.Register(task.NewNativeRunner(cfg.Tasks, cfg.Settings))
	engine.Register(task.NewGoTaskRunner())
	engine.Register(task.NewMakeRunner())
	engine.Register(task.NewJustRunner())

	// Get current directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Prepare environment injects
	shimsDir := env.GetShimsDir()
	envInjects := []string{
		fmt.Sprintf("PATH=%s:%s", shimsDir, os.Getenv("PATH")),
	}

	// Execute task
	if err := engine.Execute(ctx, cwd, taskName, taskArgs, envInjects); err != nil {
		return fmt.Errorf("task execution failed: %w", err)
	}

	return nil
}
