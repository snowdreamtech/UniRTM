// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDoctorStructure(t *testing.T) {
	assert.Contains(t, doctorCmd.Use, "doctor", "doctorCmd command use should contain 'doctor'")
	assert.NotEmpty(t, doctorCmd.Short, "doctorCmd command short description should not be empty")
	assert.True(t, doctorCmd.Run != nil || doctorCmd.RunE != nil, "Run or RunE function should be set for doctorCmd")
}

func TestRunDoctor(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("UNIRTM_DATA_DIR", tmpDir)

	cmd := doctorCmd
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := runDoctor(cmd, []string{})
	assert.NoError(t, err)
}
