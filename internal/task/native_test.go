package task

import (
	"context"
	"os"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNativeRunner_runTaskWithGraph_Timeout(t *testing.T) {
	tasks := map[string]config.Task{
		"slow": {
			Run:     "sleep 5",
			Timeout: 1,
		},
	}
	runner := NewNativeRunner(tasks, config.Settings{})

	err := runner.Run(context.Background(), ".", "slow", nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "killed") // Or context deadline exceeded
}

func TestNativeRunner_runTaskWithGraph_OutputStyles(t *testing.T) {
	tasks := map[string]config.Task{
		"hello_spinner": {
			Run:    "echo hello",
			Output: "spinner",
		},
		"hello_prefix": {
			Run:    "echo hello",
			Output: "prefix",
		},
		"hello_interleaved": {
			Run:    "echo hello",
			Output: "interleaved",
		},
		"hello_env": {
			Run: "echo hello",
		},
		"fail_spinner": {
			Run:    "exit 1",
			Output: "spinner",
		},
	}
	runner := NewNativeRunner(tasks, config.Settings{})

	// Test spinner
	err := runner.Run(context.Background(), ".", "hello_spinner", nil, nil)
	assert.NoError(t, err)

	// Test prefix
	err = runner.Run(context.Background(), ".", "hello_prefix", nil, nil)
	assert.NoError(t, err)

	// Test interleaved
	err = runner.Run(context.Background(), ".", "hello_interleaved", nil, nil)
	assert.NoError(t, err)

	// Test fail spinner
	err = runner.Run(context.Background(), ".", "fail_spinner", nil, nil)
	assert.Error(t, err)

	// Test env override
	os.Setenv("UNIRTM_TASK_OUTPUT", "interleaved")
	defer os.Unsetenv("UNIRTM_TASK_OUTPUT")
	err = runner.Run(context.Background(), ".", "hello_env", nil, nil)
	assert.NoError(t, err)
	
	// Test CI override
	os.Setenv("CI", "true")
	os.Unsetenv("UNIRTM_TASK_OUTPUT")
	defer os.Unsetenv("CI")
	err = runner.Run(context.Background(), ".", "hello_env", nil, nil)
	assert.NoError(t, err)
}

func TestNativeRunner_runTaskWithGraph_Dependencies(t *testing.T) {
	tasks := map[string]config.Task{
		"build": {
			Run:     "echo build",
			Depends: []string{"test"},
		},
		"test": {
			Run: "echo test",
		},
		"cycle1": {
			Run:     "echo c1",
			Depends: []string{"cycle2"},
		},
		"cycle2": {
			Run:     "echo c2",
			Depends: []string{"cycle1"},
		},
	}
	runner := NewNativeRunner(tasks, config.Settings{})

	// Success dependency
	err := runner.Run(context.Background(), ".", "build", nil, nil)
	assert.NoError(t, err)

	// Cycle error
	err = runner.Run(context.Background(), ".", "cycle1", nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circular dependency detected")
}
