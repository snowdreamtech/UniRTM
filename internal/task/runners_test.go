// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package task

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/config"
)

func TestGoTaskRunner_CanExecute(t *testing.T) {
	r := NewGoTaskRunner()
	if r.Name() != "go-task" {
		t.Errorf("expected go-task, got %s", r.Name())
	}

	tmpDir := t.TempDir()

	// Should be false initially
	if r.CanExecute(tmpDir, "build") {
		t.Error("expected CanExecute to be false")
	}

	// Create Taskfile.yml
	taskfilePath := filepath.Join(tmpDir, "Taskfile.yml")
	if err := os.WriteFile(taskfilePath, []byte("version: '3'\ntasks:\n  build:\n    cmds:\n      - echo build\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Should be true now
	if !r.CanExecute(tmpDir, "build") {
		t.Error("expected CanExecute to be true after creating Taskfile.yml")
	}

	// ListTasks
	tasks, err := r.ListTasks(tmpDir)
	if err != nil {
		t.Logf("unexpected error listing tasks: %v (maybe task is not installed)", err)
	} else if len(tasks) > 0 {
		// If task is installed, it might parse something, though "build" would fail without a valid Taskfile format
		// Our mock Taskfile is valid YAML so if `task` is installed, it would find it.
		found := false
		for _, task := range tasks {
			if task == "build" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected to find 'build' in tasks, got %v", tasks)
		}
	}

	// Run
	ctx := context.Background()
	if err := r.Run(ctx, tmpDir, "build", []string{}, []string{"FOO=bar"}); err != nil {
		t.Logf("expected run to succeed, got %v (maybe task is not installed)", err)
	}
}

func TestMakeRunner_CanExecute(t *testing.T) {
	r := NewMakeRunner()
	if r.Name() != "make" {
		t.Errorf("expected make, got %s", r.Name())
	}

	tmpDir := t.TempDir()

	if r.CanExecute(tmpDir, "build") {
		t.Error("expected CanExecute to be false")
	}

	makefilePath := filepath.Join(tmpDir, "Makefile")
	if err := os.WriteFile(makefilePath, []byte("build:\n\t@echo build\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if !r.CanExecute(tmpDir, "build") {
		t.Error("expected CanExecute to be true after creating Makefile")
	}

	// We can't really test MakeRunner.Run easily unless `make` is installed,
	// but CanExecute should suffice for structural testing
	tasks, err := r.ListTasks(tmpDir)
	if err != nil {
		// Just log, depends on whether `make` is installed on the testing machine
		t.Logf("ListTasks err: %v", err)
	} else {
		found := false
		for _, task := range tasks {
			if task == "build" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected to find 'build' in tasks, got %v", tasks)
		}
	}

	// Run
	ctx := context.Background()
	err = r.Run(ctx, tmpDir, "build", []string{}, []string{"FOO=bar"})
	if err != nil {
		t.Logf("expected run to fail if make is missing, got: %v", err)
	}
}

func TestJustRunner_CanExecute(t *testing.T) {
	r := NewJustRunner()
	if r.Name() != "just" {
		t.Errorf("expected just, got %s", r.Name())
	}

	tmpDir := t.TempDir()
	if r.CanExecute(tmpDir, "build") {
		t.Error("expected CanExecute to be false")
	}

	justfilePath := filepath.Join(tmpDir, "justfile")
	if err := os.WriteFile(justfilePath, []byte("build:\n    echo build\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if !r.CanExecute(tmpDir, "build") {
		t.Error("expected CanExecute to be true after creating justfile")
	}

	// ListTasks
	tasks, err := r.ListTasks(tmpDir)
	if err != nil {
		t.Logf("ListTasks err: %v", err)
	} else {
		found := false
		for _, task := range tasks {
			if task == "build" {
				found = true
				break
			}
		}
		if !found {
			t.Logf("expected to find 'build' in tasks, got %v", tasks)
		}
	}

	// Run
	ctx := context.Background()
	err = r.Run(ctx, tmpDir, "build", []string{}, []string{"FOO=bar"})
	if err != nil {
		t.Logf("expected run to fail if just is missing, got: %v", err)
	}
}

func TestNativeRunner(t *testing.T) {
	tasks := map[string]config.Task{
		"build": {Run: "echo build"},
		"test":  {Run: "echo test", Depends: []string{"build"}},
	}
	r := NewNativeRunner(tasks, config.Settings{TaskOutput: "interleaved"})
	if r.Name() != "native" {
		t.Errorf("expected native, got %s", r.Name())
	}

	tmpDir := t.TempDir()

	// CanExecute
	if !r.CanExecute(tmpDir, "build") {
		t.Error("expected native runner to always report true for known tasks")
	}
	if r.CanExecute(tmpDir, "unknown") {
		t.Error("expected native runner to report false for unknown tasks")
	}

	// ListTasks
	listedTasks, err := r.ListTasks(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error listing tasks: %v", err)
	}
	if len(listedTasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(listedTasks))
	}

	// Run
	ctx := context.Background()
	if err := r.Run(ctx, tmpDir, "test", []string{}, []string{"TEST_ENV=1"}); err != nil {
		t.Errorf("expected run to succeed, got %v", err)
	}

	// Error on unknown task
	if err := r.Run(ctx, tmpDir, "nonexistent", []string{}, []string{}); err == nil {
		t.Error("expected error when running non-existent task")
	}
}
