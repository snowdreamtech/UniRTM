// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package task

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEngine_ListTasks_Empty(t *testing.T) {
	engine := NewEngine()
	// No runners registered
	tasks := engine.ListTasks("/tmp")
	require.Empty(t, tasks)
}

func TestEngine_Execute_Error(t *testing.T) {
	engine := NewEngine()
	// No runners registered
	err := engine.Execute(context.Background(), "/tmp", "test", nil, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no suitable task runner found")
}
