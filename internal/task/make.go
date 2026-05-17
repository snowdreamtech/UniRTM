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

// ListTasks returns all targets defined in the Makefile.
func (r *MakeRunner) ListTasks(dir string) ([]string, error) {
	if !r.CanExecute(dir, "") {
		return nil, nil
	}

	// Use make -qp to list all targets without executing anything
	cmd := exec.Command("make", "-qp")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		// some make versions return exit code 1 even with -qp if there are errors in Makefile
		// but they might still output the database
	}

	var targets []string
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		// Ignore comments, assignments, and non-target lines
		if len(line) == 0 || line[0] == '#' || line[0] == '\t' || line[0] == '.' || !strings.Contains(line, ":") {
			continue
		}

		// Targets start at the beginning of the line, followed by a colon
		parts := strings.SplitN(line, ":", 2)
		target := strings.TrimSpace(parts[0])

		// Ignore internal make targets, patterns, and uppercase variables
		if target != "" && !strings.Contains(target, "=") && !strings.Contains(target, "+") && !strings.Contains(target, "$") && !strings.Contains(target, "%") && !isAllUpper(target) {
			targets = append(targets, target)
		}
	}
	return targets, nil
}

func isAllUpper(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, r := range s {
		if (r < 'A' || r > 'Z') && r != '_' && (r < '0' || r > '9') {
			return false
		}
	}
	return true
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
