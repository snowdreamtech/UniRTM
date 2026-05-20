// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unirtm/internal/cli/output"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/spf13/cobra"
)

func init() {
	tasksCmd.AddCommand(tasksListCmd)
	tasksCmd.AddCommand(tasksInfoCmd)
	tasksCmd.AddCommand(tasksDepsCmd)
	tasksCmd.AddCommand(tasksEditCmd)
	if rootCmd != nil {
		rootCmd.AddCommand(tasksCmd)
	}
}

// tasksCmd is the root of the tasks sub-command group.
var tasksCmd = &cobra.Command{
	Use:     "tasks",
	Short:   "Manage and inspect tasks defined in the config file",
	Aliases: []string{"t"},
	Long: `Manage and inspect tasks defined in the config file.

Tasks are defined in the [tasks] section of unirtm.toml.
Use 'unirtm run <task>' to execute them.

Sub-commands:
  list   List all defined tasks
  info   Show details about a specific task
  deps   Show task dependency graph
  edit   Open a task in $EDITOR

Examples:
  unirtm tasks
  unirtm tasks list
  unirtm tasks info build
  unirtm tasks deps
  unirtm tasks edit test`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTasksList(cmd, args)
	},
}

// tasksListCmd lists all defined tasks.
var tasksListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all tasks defined in the config file",
	Aliases: []string{"ls"},
	Args:    cobra.NoArgs,
	RunE:    runTasksList,
}

// tasksInfoCmd shows details about a specific task.
var tasksInfoCmd = &cobra.Command{
	Use:   "info <task>",
	Short: "Show details about a specific task",
	Args:  cobra.ExactArgs(1),
	RunE:  runTasksInfo,
}

// tasksDepsCmd shows the task dependency graph.
var tasksDepsCmd = &cobra.Command{
	Use:   "deps",
	Short: "Show task dependency graph",
	Args:  cobra.NoArgs,
	RunE:  runTasksDeps,
}

// tasksEditCmd opens a task definition in $EDITOR.
var tasksEditCmd = &cobra.Command{
	Use:   "edit <task>",
	Short: "Open task definition in $EDITOR",
	Args:  cobra.ExactArgs(1),
	RunE:  runTasksEdit,
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func loadTasksConfig() (*config.Config, error) {
	cm := config.NewConfigManager()
	cfg, err := cm.LoadHierarchy(context.Background())
	if err != nil || cfg == nil {
		return &config.Config{Tasks: make(map[string]config.Task)}, nil
	}
	if cfg.Tasks == nil {
		cfg.Tasks = make(map[string]config.Task)
	}
	return cfg, nil
}

// ─── list ─────────────────────────────────────────────────────────────────────

func runTasksList(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	cfg, _ := loadTasksConfig()

	if len(cfg.Tasks) == 0 {
		cfgPath := filepath.Base(resolveConfigFilePath(false))
		formatter.Info(fmt.Sprintf("No tasks defined. Add [tasks] to %s.", cfgPath), nil)
		return nil
	}

	names := make([]string, 0, len(cfg.Tasks))
	for name := range cfg.Tasks {
		names = append(names, name)
	}
	sort.Strings(names)

	if jsonOutput {
		type jsonTask struct {
			Name        string   `json:"name"`
			Description string   `json:"description"`
			Depends     []string `json:"depends,omitempty"`
		}
		tasks := make([]jsonTask, 0, len(names))
		for _, n := range names {
			t := cfg.Tasks[n]
			tasks = append(tasks, jsonTask{
				Name:        n,
				Description: t.Description,
				Depends:     t.Depends,
			})
		}
		formatter.Success("Tasks", map[string]interface{}{"count": len(tasks), "tasks": tasks})
		return nil
	}

	tableData := pterm.TableData{{"TASK", "DESCRIPTION", "DEPENDS"}}
	for _, n := range names {
		t := cfg.Tasks[n]
		desc := t.Description
		if desc == "" {
			desc = pterm.FgDefault.Sprint("─")
		}
		deps := "─"
		if len(t.Depends) > 0 {
			deps = strings.Join(t.Depends, ", ")
		}
		tableData = append(tableData, []string{
			pterm.FgCyan.Sprint(n),
			desc,
			deps,
		})
	}

	fmt.Println()
	pterm.DefaultTable.
		WithHasHeader(true).
		WithSeparator("   ").
		WithHeaderStyle(pterm.NewStyle(pterm.FgCyan, pterm.Bold)).
		WithData(tableData).
		Render()
	return nil
}

// ─── info ─────────────────────────────────────────────────────────────────────

func runTasksInfo(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	taskName := args[0]
	cfg, _ := loadTasksConfig()

	t, ok := cfg.Tasks[taskName]
	if !ok {
		formatter.Error(fmt.Sprintf("Task %q not found. Run 'unirtm tasks' to list all tasks.", taskName))
		return fmt.Errorf("task not found: %s", taskName)
	}

	if jsonOutput {
		formatter.Success(fmt.Sprintf("Task: %s", taskName), map[string]interface{}{
			"name":        taskName,
			"description": t.Description,
			"run":         t.Run,
			"depends":     t.Depends,
			"env":         t.Env,
		})
		return nil
	}

	fmt.Println()
	pterm.DefaultSection.Printf("Task: %s", pterm.FgCyan.Sprint(taskName))

	rows := pterm.TableData{}
	if t.Description != "" {
		rows = append(rows, []string{"Description", t.Description})
	}
	if len(t.Depends) > 0 {
		rows = append(rows, []string{"Depends", strings.Join(t.Depends, " → ")})
	}
	if len(t.Env) > 0 {
		envPairs := make([]string, 0, len(t.Env))
		for k, v := range t.Env {
			envPairs = append(envPairs, fmt.Sprintf("%s=%v", k, v))
		}
		rows = append(rows, []string{"Env", strings.Join(envPairs, "  ")})
	}

	if len(rows) > 0 {
		pterm.DefaultTable.WithSeparator("   ").WithData(rows).Render()
	}

	if t.Run != "" {
		fmt.Printf("\n%s\n%s\n", pterm.FgDefault.Sprint("Run:"), pterm.FgYellow.Sprint(t.Run))
	}
	return nil
}

// ─── deps ─────────────────────────────────────────────────────────────────────

func runTasksDeps(cmd *cobra.Command, args []string) error {
	cfg, _ := loadTasksConfig()

	if len(cfg.Tasks) == 0 {
		fmt.Println("No tasks defined.")
		return nil
	}

	names := make([]string, 0, len(cfg.Tasks))
	for n := range cfg.Tasks {
		names = append(names, n)
	}
	sort.Strings(names)

	fmt.Println("\nTask dependency graph:")
	for _, n := range names {
		t := cfg.Tasks[n]
		if len(t.Depends) == 0 {
			fmt.Printf("  %s\n", pterm.FgCyan.Sprint(n))
		} else {
			fmt.Printf("  %s ← %s\n",
				pterm.FgCyan.Sprint(n),
				pterm.FgYellow.Sprint(strings.Join(t.Depends, ", ")),
			)
		}
	}
	return nil
}

// ─── edit ─────────────────────────────────────────────────────────────────────

func runTasksEdit(cmd *cobra.Command, args []string) error {
	formatter := output.NewFormatter(output.FormatterOptions{
		Format:  getOutputFormat(),
		NoColor: false,
		Writer:  os.Stdout,
		Quiet:   quiet,
		Verbose: verbose,
	})

	taskName := args[0]
	cfg, _ := loadTasksConfig()
	if _, ok := cfg.Tasks[taskName]; !ok {
		formatter.Warning(fmt.Sprintf("Task %q not found, but will open config for editing.", taskName))
	}

	cfgPath := resolveConfigFilePath(false)
	editor := env.Get("VISUAL")
	if editor == "" {
		editor = env.Get("EDITOR")
	}
	if editor == "" && cfg != nil && cfg.Settings.Editor != "" {
		editor = cfg.Settings.Editor
	}
	if editor == "" {
		editor = "vi"
	}

	editorCmd := exec.Command(editor, cfgPath)
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr
	return editorCmd.Run()
}
