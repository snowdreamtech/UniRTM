// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/task"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// init registers the run command to the root command.
func init() {
	if rootCmd != nil {
		rootCmd.AddCommand(runCmd)
	}

	// Scheme 3: Add --output flag for task output mode selection
	runCmd.Flags().String("output", "", "Task output mode: spinner, prefix, or interleaved (default: auto-detect)")
	runCmd.Flags().Bool("fix", false, "Apply automatic fixes if supported by the task")
}

// runCmd represents the run command which executes a task via the routing engine.
var runCmd = &cobra.Command{
	Use:   "run [task] [args...]",
	Short: "Run a task using the multi-modal routing engine",
	Long: `Run a task using the multi-modal routing engine.

UniRTM delegates tasks to professional tools (go-task, make, just)
if their configuration files are detected, or falls back to executing
tasks defined in unirtm.toml.

If no task is provided, it lists all available tasks.

Examples:
  # List all available tasks
  unirtm run

  # Run a build task
  unirtm run build

  # Run a task with arguments
  unirtm run test -- -v

  # Run with custom output mode
  unirtm run test --output interleaved`,
	Aliases:            []string{"r"},
	Args:               cobra.MinimumNArgs(0),
	DisableFlagParsing: false,
	RunE:               runTaskCommand,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		ctx := context.Background()
		configManager := config.NewConfigManager()
		cfg, err := configManager.LoadHierarchy(ctx)
		if err != nil {
			fmt.Println("Config error:", err)
			cfg = &config.Config{}
		}

		engine := task.NewEngine()
		engine.Register(task.NewNativeRunner(cfg.Tasks, cfg.Settings))
		engine.Register(task.NewGoTaskRunner())
		engine.Register(task.NewMakeRunner())
		engine.Register(task.NewJustRunner())

		cwd, _ := os.Getwd()
		tasks := engine.ListTasks(cwd)
		return tasks, cobra.ShellCompDirectiveNoFileComp
	},
}

// runTaskCommand executes the task routing.
func runTaskCommand(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Scheme 3: Get output mode from flag and set environment variable
	outputMode, _ := cmd.Flags().GetString("output")
	if outputMode != "" {
		os.Setenv("UNIRTM_TASK_OUTPUT", outputMode)
	}

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
		fmt.Println("Config error:", err)
		cfg = &config.Config{
			Tasks: make(map[string]config.Task),
		}
	} else if cfg != nil {
		// Apply [env] variables from config to current process
		cfg.ApplyEnvironment()
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

	if len(args) == 0 {
		tasks := engine.ListTasks(cwd)
		if len(tasks) == 0 {
			output.Info("No tasks available in current context.")
			return nil
		}

		isTerminal := term.IsTerminal(int(os.Stdout.Fd()))
		if isTerminal {
			pterm.DefaultSection.Println("Available Tasks")
		} else {
			fmt.Println("Available Tasks:")
		}

		var taskItems []pterm.BulletListItem
		for _, t := range tasks {
			taskItems = append(taskItems, pterm.BulletListItem{Level: 0, Text: pterm.FgCyan.Sprint(t)})
		}
		pterm.DefaultBulletList.WithItems(taskItems).Render()
		return nil
	}

	taskName := args[0]
	taskArgs := args[1:]

	// Prepare environment injects
	shimsDir := env.GetShimsDir()
	envInjects := []string{
		fmt.Sprintf("PATH=%s:%s", shimsDir, env.Get("PATH")),
	}

	isFix, _ := cmd.Flags().GetBool("fix")
	if isFix {
		envInjects = append(envInjects, "UNIRTM_FIX=1")
	}

	// Execute task
	if err := engine.Execute(ctx, cwd, taskName, taskArgs, envInjects); err != nil {
		if strings.Contains(err.Error(), "no suitable task runner found") {
			// Suggest similar tasks or commands if not found
			var candidates []string
			for name := range cfg.Tasks {
				candidates = append(candidates, name)
			}
			if rootCmd != nil {
				for _, cmd := range rootCmd.Commands() {
					candidates = append(candidates, cmd.Name())
					candidates = append(candidates, cmd.Aliases...)
				}
			}
			for name := range cfg.Tools {
				candidates = append(candidates, name)
			}

			output.Suggest(os.Stderr, taskName, candidates)
		}

		return fmt.Errorf("task execution failed: %w", err)
	}

	return nil
}
