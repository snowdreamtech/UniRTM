// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTasksCommandStructure(t *testing.T) {
	assert.Contains(t, tasksCmd.Use, "tasks")
	assert.Contains(t, tasksCmd.Aliases, "t")
}

func TestTasksListSubcommand(t *testing.T) {
	assert.NotNil(t, tasksListCmd.RunE)
}

func TestTasksInfoSubcommand(t *testing.T) {
	assert.NotNil(t, tasksInfoCmd.RunE)
	err := tasksInfoCmd.Args(tasksInfoCmd, []string{"build"})
	assert.NoError(t, err)
	err = tasksInfoCmd.Args(tasksInfoCmd, []string{})
	assert.Error(t, err)
}

func TestTasksDepsSubcommand(t *testing.T) {
	assert.NotNil(t, tasksDepsCmd.RunE)
}

func TestTasksEditSubcommand(t *testing.T) {
	assert.NotNil(t, tasksEditCmd.RunE)
	err := tasksEditCmd.Args(tasksEditCmd, []string{"test"})
	assert.NoError(t, err)
}

func TestTasksListRun(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)
	t.Setenv("UNIRTM_CONFIG_DIR", tmpDir)

	err := tasksListCmd.RunE(tasksListCmd, []string{})
	assert.NoError(t, err)
}

func TestTasksInfoRun(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)
	t.Setenv("UNIRTM_CONFIG_DIR", tmpDir)

	err := tasksInfoCmd.RunE(tasksInfoCmd, []string{"dummy-task"})
	assert.Error(t, err)
}

func TestTasksDepsRun(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)
	t.Setenv("UNIRTM_CONFIG_DIR", tmpDir)

	err := tasksDepsCmd.RunE(tasksDepsCmd, []string{"dummy-task"})
	assert.NoError(t, err)
}

