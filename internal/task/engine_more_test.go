// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package task

import (
	"context"
	"errors"
	"testing"
)

type errorRunner struct{}

func (e *errorRunner) CanExecute(dir, task string) bool { return false }
func (e *errorRunner) Run(ctx context.Context, dir, task string, args, env []string) error {
	return nil
}
func (e *errorRunner) ListTasks(dir string) ([]string, error) { return nil, errors.New("err") }
func (e *errorRunner) Name() string                           { return "errorRunner" }

func TestEngine_Execute_NotFound(t *testing.T) {
	eng := NewEngine()
	err := eng.Execute(context.Background(), ".", "nonexistent_task_12345", nil, nil)
	if err == nil {
		t.Errorf("expected error for nonexistent task")
	}
}

func TestEngine_ListTasks_Error(t *testing.T) {
	eng := NewEngine()
	eng.runners = append(eng.runners, &errorRunner{})

	// Error from runner should be ignored
	eng.ListTasks(".")
}
