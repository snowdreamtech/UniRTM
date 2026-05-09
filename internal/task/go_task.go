// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package task

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
)

// GoTaskRunner delegates task execution to the system's `task` command
// if a Taskfile.yml or Taskfile.yaml is detected in the working directory.
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
func (r *GoTaskRunner) CanExecute(dir string) bool {
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

// Run executes the task by delegating to `task <taskName>`.
func (r *GoTaskRunner) Run(ctx context.Context, dir string, taskName string, args []string, env []string) error {
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
