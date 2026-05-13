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

// JustRunner delegates task execution to the system's `just` command
// if a Justfile is detected in the working directory.
type JustRunner struct{}

// NewJustRunner creates a new JustRunner instance.
func NewJustRunner() *JustRunner {
	return &JustRunner{}
}

// Name returns the name of this runner.
func (r *JustRunner) Name() string {
	return "just"
}

// CanExecute returns true if a Justfile or justfile exists in the target directory.
func (r *JustRunner) CanExecute(dir string, taskName string) bool {
	if _, err := os.Stat(filepath.Join(dir, "Justfile")); err == nil {
		return true
	}
	if _, err := os.Stat(filepath.Join(dir, "justfile")); err == nil {
		return true
	}
	return false
}

// ListTasks returns all targets defined in the Justfile.
func (r *JustRunner) ListTasks(dir string) ([]string, error) {
	if !r.CanExecute(dir, "") {
		return nil, nil
	}

	// Use just --summary to list all targets
	cmd := exec.Command("just", "--summary")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return nil, nil
	}

	return strings.Fields(string(out)), nil
}

// Run executes the task by delegating to `just <taskName>`.
func (r *JustRunner) Run(ctx context.Context, dir string, taskName string, args []string, env []string) error {
	cmdArgs := []string{taskName}
	cmdArgs = append(cmdArgs, args...)
	
	cmd := exec.CommandContext(ctx, "just", cmdArgs...)
	cmd.Dir = dir
	
	// Pass through the environment variables injected by UniRTM
	cmd.Env = append(os.Environ(), env...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
