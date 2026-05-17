// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package task_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/task"
)

// mockRunner is a fake runner for testing routing logic.
type mockRunner struct {
	name       string
	canExecute bool
	called     bool
}

func (m *mockRunner) Name() string                                { return m.name }
func (m *mockRunner) CanExecute(dir string, taskName string) bool { return m.canExecute }
func (m *mockRunner) Run(ctx context.Context, dir string, taskName string, args []string, env []string) error {
	m.called = true
	return nil
}
func (m *mockRunner) ListTasks(dir string) ([]string, error) { return nil, nil }

func TestEngineRouting(t *testing.T) {
	engine := task.NewEngine()

	mock1 := &mockRunner{name: "mock1", canExecute: false}
	mock2 := &mockRunner{name: "mock2", canExecute: true}
	mock3 := &mockRunner{name: "mock3", canExecute: true}

	engine.Register(mock1)
	engine.Register(mock2)
	engine.Register(mock3) // Should not be called because mock2 intercepts it

	err := engine.Execute(context.Background(), ".", "test_task", nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if mock1.called {
		t.Errorf("expected mock1 not to be called")
	}
	if !mock2.called {
		t.Errorf("expected mock2 to be called")
	}
	if mock3.called {
		t.Errorf("expected mock3 not to be called")
	}
}

func TestMakeRunnerCanExecute(t *testing.T) {
	tmpDir := t.TempDir()
	runner := task.NewMakeRunner()

	// Should be false initially
	if runner.CanExecute(tmpDir, "test") {
		t.Errorf("expected MakeRunner.CanExecute to be false when no Makefile exists")
	}

	// Create a Makefile
	if err := os.WriteFile(filepath.Join(tmpDir, "Makefile"), []byte("test:\n\techo test"), 0644); err != nil {
		t.Fatalf("failed to write Makefile: %v", err)
	}

	// Should be true now
	if !runner.CanExecute(tmpDir, "test") {
		t.Errorf("expected MakeRunner.CanExecute to be true when Makefile exists")
	}
}

func TestNativeRunner(t *testing.T) {
	tasks := map[string]config.Task{
		"build": {Run: "echo 'building'"},
	}
	runner := task.NewNativeRunner(tasks, config.Settings{})

	if !runner.CanExecute(".", "build") {
		t.Errorf("expected NativeRunner to return true for 'build' task")
	}

	if runner.CanExecute(".", "missing") {
		t.Errorf("expected NativeRunner to return false for 'missing' task")
	}

	err := runner.Run(context.Background(), ".", "missing", nil, nil)
	if err == nil {
		t.Errorf("expected error for missing task")
	}
}
