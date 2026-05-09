// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package task provides multi-modal task routing and execution.
package task

import "context"

// Runner defines the interface for task execution backends.
// This allows UniRTM to seamlessly delegate tasks to tools like make, just, or go-task.
type Runner interface {
	// Name returns the name of the runner (e.g., "native", "make").
	Name() string

	// CanExecute checks if this runner is applicable for the given directory
	// (e.g., checks for Taskfile.yml, Makefile).
	CanExecute(dir string) bool

	// Run executes the specified task within the given directory and environment.
	// taskName is the name of the task to run.
	// args are additional arguments passed to the task.
	// env is a slice of "KEY=VALUE" strings to inject into the execution environment.
	Run(ctx context.Context, dir string, taskName string, args []string, env []string) error
}
