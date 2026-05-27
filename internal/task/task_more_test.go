package task

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGoTaskRunner_More(t *testing.T) {
	// create fake task binary
	binDir := filepath.Join(t.TempDir(), "bin")
	os.MkdirAll(binDir, 0755)
	
	// Create fake task script
	fakeTask := filepath.Join(binDir, "task")
	script := `#!/bin/sh
if [ "$1" = "--list-all" ]; then
  echo "* build:   Build the project"
  echo "* test:    Run tests"
  echo "  some other line"
  exit 0
fi
exit 0
`
	os.WriteFile(fakeTask, []byte(script), 0755)

	// Update PATH
	oldPath := os.Getenv("PATH")
	defer os.Setenv("PATH", oldPath)
	os.Setenv("PATH", binDir+string(os.PathListSeparator)+oldPath)

	r := NewGoTaskRunner()
	dir := t.TempDir()
	
	// CanExecute -> false if no Taskfile
	assert.False(t, r.CanExecute(dir, ""))

	// create Taskfile.yaml
	os.WriteFile(filepath.Join(dir, "Taskfile.yaml"), []byte("version: '3'"), 0644)
	assert.True(t, r.CanExecute(dir, ""))

	// ListTasks
	tasks, err := r.ListTasks(dir)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"build", "test"}, tasks)

	// Run
	err = r.Run(context.Background(), dir, "build", []string{"--verbose"}, nil)
	assert.NoError(t, err)
}

func TestJustRunner_More(t *testing.T) {
	// create fake just binary
	binDir := filepath.Join(t.TempDir(), "bin")
	os.MkdirAll(binDir, 0755)
	
	// Create fake just script
	fakeJust := filepath.Join(binDir, "just")
	script := `#!/bin/sh
if [ "$1" = "--summary" ]; then
  echo "build test lint"
  exit 0
fi
exit 0
`
	os.WriteFile(fakeJust, []byte(script), 0755)

	// Update PATH
	oldPath := os.Getenv("PATH")
	defer os.Setenv("PATH", oldPath)
	os.Setenv("PATH", binDir+string(os.PathListSeparator)+oldPath)

	r := NewJustRunner()
	dir := t.TempDir()
	
	// CanExecute -> false if no justfile
	assert.False(t, r.CanExecute(dir, ""))

	// create justfile
	os.WriteFile(filepath.Join(dir, "justfile"), []byte("build:"), 0644)
	assert.True(t, r.CanExecute(dir, ""))

	// ListTasks
	tasks, err := r.ListTasks(dir)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"build", "test", "lint"}, tasks)
}

func TestPrefixWriter(t *testing.T) {
	var buf bytes.Buffer
	pw := &prefixWriter{w: &buf, prefix: "[prefix] ", atStart: true}
	pw.Write([]byte("hello\nworld\n"))
	assert.Equal(t, "[prefix] hello\n[prefix] world\n", buf.String())
}
