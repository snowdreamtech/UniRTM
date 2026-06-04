// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package task

import (
	"context"
	"os"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/stretchr/testify/require"
)

func TestNativeRunner_runTaskWithGraph_Cycle(t *testing.T) {
	tasks := map[string]config.Task{
		"A": {Run: config.StringArray{"echo A"}, Depends: []string{"B"}},
		"B": {Run: config.StringArray{"echo B"}, Depends: []string{"A"}},
	}
	runner := NewNativeRunner(tasks, config.Settings{})

	err := runner.Run(context.Background(), "/tmp", "A", nil, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "circular dependency")
}

func TestNativeRunner_runTaskWithGraph_MissingDep(t *testing.T) {
	tasks := map[string]config.Task{
		"A": {Run: config.StringArray{"echo A"}, Depends: []string{"B"}},
	}
	runner := NewNativeRunner(tasks, config.Settings{})

	err := runner.Run(context.Background(), "/tmp", "A", nil, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "dependency")
}

func TestPrefixWriter_Write(t *testing.T) {
	pw := &prefixWriter{
		prefix: "PREFIX",
		w:      os.Stdout,
	}

	n, err := pw.Write([]byte("hello\nworld\n"))
	require.NoError(t, err)
	require.Equal(t, 12, n)
}

func TestNativeRunner_runTaskWithGraph_Normal(t *testing.T) {
	tasks := map[string]config.Task{
		"A": {Run: config.StringArray{"echo A"}, Env: map[string]interface{}{"FOO": "BAR"}, Timeout: 5, Output: "interleaved"},
	}
	runner := NewNativeRunner(tasks, config.Settings{})

	err := runner.Run(context.Background(), "/tmp", "A", []string{"arg1"}, []string{"ENV1=1"})
	require.NoError(t, err)
}

func TestNativeRunner_runTaskWithGraph_InvalidScript(t *testing.T) {
	tasks := map[string]config.Task{
		"A": {Run: config.StringArray{"echo \"unclosed quote"}, Output: "prefix"},
	}
	runner := NewNativeRunner(tasks, config.Settings{})

	err := runner.Run(context.Background(), "/tmp", "A", nil, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse task script")
}
