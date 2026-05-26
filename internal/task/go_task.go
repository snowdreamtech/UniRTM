// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package task

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GoTaskRunner delegates task execution to the system's `task` command
// if a Taskfile.yml, Taskfile.yaml or Taskfile.dist.yml is detected in the working directory.
type GoTaskRunner struct{}

// NewGoTaskRunner creates a new GoTaskRunner instance.
func NewGoTaskRunner() *GoTaskRunner {
	return &GoTaskRunner{}
}

// Name returns the name of this runner.
func (r *GoTaskRunner) Name() string {
	return "go-task"
}

// CanExecute returns true if a Taskfile.yml, Taskfile.yaml or Taskfile.dist.yml exists.
func (r *GoTaskRunner) CanExecute(dir string, taskName string) bool {
	files := []string{
		"Taskfile.yml",
		"Taskfile.yaml",
		"Taskfile.dist.yml",
		"Taskfile.dist.yaml",
	}
	for _, file := range files {
		if _, err := os.Stat(filepath.Join(dir, file)); err == nil {
			return true
		}
	}
	return false
}

// ListTasks returns all tasks defined in the Taskfile.
func (r *GoTaskRunner) ListTasks(dir string) ([]string, error) {
	if !r.CanExecute(dir, "") {
		return nil, nil
	}

	// Use task --list-all to list all tasks
	cmd := exec.Command("task", "--list-all")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return nil, nil // Silently fail if task is not installed or error occurs
	}

	var tasks []string
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// go-task --list-all usually outputs: "* task-name:   Description"
		if strings.HasPrefix(line, "* ") {
			parts := strings.SplitN(line[2:], ":", 2)
			taskName := strings.TrimSpace(parts[0])
			if taskName != "" {
				tasks = append(tasks, taskName)
			}
		}
	}
	return tasks, nil
}

// Run executes the task by delegating to `task <taskName>`.
func (r *GoTaskRunner) Run(ctx context.Context, dir string, taskName string, args []string, env []string) error {
	if taskName == "" {
		taskName = "default"
	}
	
	cmdArgs := []string{taskName}
	cmdArgs = append(cmdArgs, args...)

	cmd := exec.CommandContext(ctx, "task", cmdArgs...)
	cmd.Dir = dir

	// Pass through the environment variables injected by UniRTM
	cmd.Env = append(os.Environ(), env...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
