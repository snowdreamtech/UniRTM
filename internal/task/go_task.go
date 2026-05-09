// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package task

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"

	gotask "github.com/go-task/task/v3"
)

// GoTaskRunner delegates task execution directly to the embedded `go-task` engine
// if a Taskfile.yml or Taskfile.yaml is detected in the working directory.
type GoTaskRunner struct{}

var envMutex sync.Mutex

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

// Run executes the task by delegating directly to the go-task library.
func (r *GoTaskRunner) Run(ctx context.Context, dir string, taskName string, args []string, env []string) error {
	// Temporarily inject environment variables since go-task inherently inherits from os.Environ
	envMutex.Lock()
	
	// Save existing environment to restore later
	originalEnv := os.Environ()
	
	// Apply the environment variables from UniRTM
	for _, e := range env {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			os.Setenv(parts[0], parts[1])
		}
	}
	
	defer func() {
		// Restore the entire environment
		os.Clearenv()
		for _, e := range originalEnv {
			parts := strings.SplitN(e, "=", 2)
			if len(parts) == 2 {
				os.Setenv(parts[0], parts[1])
			}
		}
		envMutex.Unlock()
	}()

	// Initialize the executor
	e := &gotask.Executor{
		Dir:    dir,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Stdin:  os.Stdin,
	}

	if err := e.Setup(); err != nil {
		return err
	}

	// Prepare task calls
	// If no task is specified, default is often "default" or empty.
	if taskName == "" {
		taskName = "default"
	}
	
	calls := []*gotask.Call{
		{Task: taskName},
	}

	return e.Run(ctx, calls...)
}
