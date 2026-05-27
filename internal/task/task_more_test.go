package task

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/config"
)

func TestEngine_ListTasksAndExecute(t *testing.T) {
	engine := NewEngine()
	tasks := map[string]config.Task{
		"build": {Run: "echo build"},
	}
	settings := config.Settings{}
	
	engine.Register(NewNativeRunner(tasks, settings))
	engine.Register(NewGoTaskRunner())
	engine.Register(NewMakeRunner())
	engine.Register(NewJustRunner())

	dir := t.TempDir()

	// Write Taskfile.yml for GoTaskRunner
	taskYaml := `
version: '3'
tasks:
  hello:
    cmds:
      - echo "hello"
`
	os.WriteFile(filepath.Join(dir, "Taskfile.yml"), []byte(taskYaml), 0644)

	// Test Engine ListTasks
	allTasks := engine.ListTasks(dir)
	if len(allTasks) == 0 {
		t.Fatalf("expected tasks, got 0")
	}

	// Test GoTaskRunner ListTasks parsing
	goTask := NewGoTaskRunner()
	if !goTask.CanExecute(dir, "") {
		t.Fatal("GoTaskRunner should be able to execute")
	}
	_, _ = goTask.ListTasks(dir)

	// Test NativeRunner executing
	native := NewNativeRunner(tasks, settings)
	_ = native.Run(context.Background(), dir, "build", nil, nil)
}

func TestNativeRunner_CycleDetection(t *testing.T) {
	tasks := map[string]config.Task{
		"a": {Run: "echo a", Depends: []string{"b"}},
		"b": {Run: "echo b", Depends: []string{"c"}},
		"c": {Run: "echo c", Depends: []string{"a"}},
	}
	settings := config.Settings{}
	native := NewNativeRunner(tasks, settings)
	
	err := native.Run(context.Background(), ".", "a", nil, nil)
	if err == nil {
		t.Fatal("expected cycle error, got nil")
	}
}

func TestPrefixWriter_Write(t *testing.T) {
	var buf bytes.Buffer
	pw := &prefixWriter{
		w:       &buf,
		prefix:  "[test] ",
		atStart: true,
	}

	_, err := pw.Write([]byte("line 1\nline 2\n"))
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}

	expected := "[test] line 1\n[test] line 2\n"
	if buf.String() != expected {
		t.Errorf("expected %q, got %q", expected, buf.String())
	}
}
