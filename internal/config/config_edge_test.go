// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDurationToSeconds_Coverage(t *testing.T) {
	tests := []struct {
		input   string
		want    int
		wantErr bool
	}{
		{"", 0, false},
		{"60", 60, false},
		{"60s", 60, false},
		{"30m", 30 * 60, false},
		{"2h", 2 * 3600, false},
		{"7d", 7 * 86400, false},
		{"2w", 2 * 86400 * 7, false},
		{"1sec", 1, false},
		{"5min", 300, false},
		{"1hour", 3600, false},
		{"3day", 3 * 86400, false},
		{"1week", 7 * 86400, false},
		{"invalid", 0, true},
		{"10xyz", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseDurationToSeconds(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestConfig_Merge_EdgeCases(t *testing.T) {
	t.Run("nil receiver is no-op", func(t *testing.T) {
		var c *Config
		c.Merge(&Config{Tools: map[string]ToolConfig{"go": {Version: "1.22"}}})
		// Should not panic
	})

	t.Run("nil other is no-op", func(t *testing.T) {
		c := &Config{Tools: map[string]ToolConfig{"go": {Version: "1.22"}}}
		c.Merge(nil)
		assert.Equal(t, "1.22", c.Tools["go"].Version)
	})

	t.Run("merge tools with existing key not overwritten", func(t *testing.T) {
		c1 := &Config{Tools: map[string]ToolConfig{"go": {Version: "1.22"}}}
		c2 := &Config{Tools: map[string]ToolConfig{"go": {Version: "1.21"}, "node": {Version: "20.0"}}}
		c1.Merge(c2)
		assert.Equal(t, "1.22", c1.Tools["go"].Version) // c1 wins
		assert.Equal(t, "20.0", c1.Tools["node"].Version)
	})

	t.Run("merge env with existing key not overwritten", func(t *testing.T) {
		c1 := &Config{Env: map[string]interface{}{"A": "1"}}
		c2 := &Config{Env: map[string]interface{}{"A": "2", "B": "3"}}
		c1.Merge(c2)
		assert.Equal(t, "1", c1.Env["A"]) // c1 wins
		assert.Equal(t, "3", c1.Env["B"])
	})

	t.Run("merge tasks with existing key not overwritten", func(t *testing.T) {
		c1 := &Config{Tasks: map[string]Task{"build": {Run: StringArray{"make all"}}}}
		c2 := &Config{Tasks: map[string]Task{"build": {Run: StringArray{"go build ./..."}}, "test": {Run: StringArray{"go test ./..."}}}}
		c1.Merge(c2)
		assert.Equal(t, StringArray{"make all"}, c1.Tasks["build"].Run) // c1 wins
		assert.Equal(t, StringArray{"go test ./..."}, c1.Tasks["test"].Run)
	})

	t.Run("merge environments", func(t *testing.T) {
		c1 := &Config{Environments: map[string]EnvironmentConfig{"prod": {Tools: map[string]ToolConfig{"go": {Version: "1.22"}}}}}
		c2 := &Config{Environments: map[string]EnvironmentConfig{"dev": {Tools: map[string]ToolConfig{"node": {Version: "20.0"}}}}}
		c1.Merge(c2)
		assert.Contains(t, c1.Environments, "prod")
		assert.Contains(t, c1.Environments, "dev")
	})

	t.Run("merge aliases deep merge", func(t *testing.T) {
		c1 := &Config{
			Aliases: map[string]map[string]string{
				"node": {"lts": "18.0"},
			},
		}
		c2 := &Config{
			Aliases: map[string]map[string]string{
				"node": {"lts": "20.0", "stable": "20.0"},
				"go":   {"latest": "1.22"},
			},
		}
		c1.Merge(c2)
		// node lts: c1 wins
		assert.Equal(t, "18.0", c1.Aliases["node"]["lts"])
		// node stable: from c2 (not in c1)
		assert.Equal(t, "20.0", c1.Aliases["node"]["stable"])
		// go: from c2
		assert.Equal(t, "1.22", c1.Aliases["go"]["latest"])
	})

	t.Run("merge with nil tools in c1 gets allocated", func(t *testing.T) {
		c1 := &Config{}
		c2 := &Config{Tools: map[string]ToolConfig{"go": {Version: "1.22"}}}
		c1.Merge(c2)
		assert.Equal(t, "1.22", c1.Tools["go"].Version)
	})
}

func TestStringArray_UnmarshalYAML(t *testing.T) {
	// Test via JSON (fallback path for non-string non-array)
	var sa StringArray
	err := sa.UnmarshalJSON([]byte(`"single"`))
	require.NoError(t, err)
	assert.Equal(t, StringArray{"single"}, sa)

	err = sa.UnmarshalJSON([]byte(`["a","b"]`))
	require.NoError(t, err)
	assert.Equal(t, StringArray{"a", "b"}, sa)

	// Invalid JSON (error path)
	err = sa.UnmarshalJSON([]byte(`123`))
	assert.Error(t, err)
}
