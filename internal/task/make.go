// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package task

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
)

// MakeRunner delegates task execution to the system's `make` command
// if a Makefile is detected in the working directory.
type MakeRunner struct{}

// NewMakeRunner creates a new MakeRunner instance.
func NewMakeRunner() *MakeRunner {
	return &MakeRunner{}
}

// Name returns the name of this runner.
func (r *MakeRunner) Name() string {
	return "make"
}

// CanExecute returns true if a Makefile or makefile exists in the target directory.
func (r *MakeRunner) CanExecute(dir string, taskName string) bool {
	if _, err := os.Stat(filepath.Join(dir, "Makefile")); err == nil {
		return true
	}
	if _, err := os.Stat(filepath.Join(dir, "makefile")); err == nil {
		return true
	}
	return false
}

// Run executes the task by delegating to `make <taskName>`.
func (r *MakeRunner) Run(ctx context.Context, dir string, taskName string, args []string, env []string) error {
	cmdArgs := []string{taskName}
	cmdArgs = append(cmdArgs, args...)
	
	cmd := exec.CommandContext(ctx, "make", cmdArgs...)
	cmd.Dir = dir
	
	// Pass through the environment variables injected by UniRTM
	cmd.Env = append(os.Environ(), env...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
