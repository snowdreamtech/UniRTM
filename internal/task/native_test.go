// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package task

import (
	"context"
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
	runner := NewNativeRunner(tasks, config.Settings{TaskOutput: "interleaved"})

	err := runner.Run(context.Background(), ".", "slow", nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

func TestNativeRunner_runTaskWithGraph_OutputStyles(t *testing.T) {
	tasks := map[string]config.Task{
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
	}
	runner := NewNativeRunner(tasks, config.Settings{TaskOutput: "interleaved"})

	// Test prefix
	err := runner.Run(context.Background(), ".", "hello_prefix", nil, nil)
	assert.NoError(t, err)

	// Test interleaved
	err = runner.Run(context.Background(), ".", "hello_interleaved", nil, nil)
	assert.NoError(t, err)

	// Test env override
	t.Setenv("UNIRTM_TASK_OUTPUT", "interleaved")
	err = runner.Run(context.Background(), ".", "hello_env", nil, nil)
	assert.NoError(t, err)

	// Test CI override
	t.Setenv("CI", "true")
	t.Setenv("UNIRTM_TASK_OUTPUT", "")
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
		"verify": {
			Depends: []string{"lint", "test"},
		},
		"lint": {
			Run: "echo lint",
		},
	}
	runner := NewNativeRunner(tasks, config.Settings{TaskOutput: "interleaved"})

	// Success dependency
	err := runner.Run(context.Background(), ".", "build", nil, nil)
	assert.NoError(t, err)

	// Cycle error
	err = runner.Run(context.Background(), ".", "cycle1", nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circular dependency detected")

	// Task with only depends and no run
	err = runner.Run(context.Background(), ".", "verify", nil, nil)
	assert.NoError(t, err)
}
