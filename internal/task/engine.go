// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package task

import (
	"context"
	"fmt"
)

// Engine is the central router for task execution.
// It iterates through registered runners and delegates execution to the first applicable one.
type Engine struct {
	runners []Runner
}

// NewEngine creates a new task routing engine.
func NewEngine() *Engine {
	return &Engine{
		runners: make([]Runner, 0),
	}
}

// Register adds a runner to the engine. Runners are evaluated in the order they are registered.
func (e *Engine) Register(r Runner) {
	e.runners = append(e.runners, r)
}

// Execute routes the task execution to the appropriate runner based on CanExecute.
func (e *Engine) Execute(ctx context.Context, dir string, taskName string, args []string, env []string) error {
	for _, r := range e.runners {
		if r.CanExecute(dir) {
			return r.Run(ctx, dir, taskName, args, env)
		}
	}
	return fmt.Errorf("no suitable task runner found for directory %s", dir)
}
