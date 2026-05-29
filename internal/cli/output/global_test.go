// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package output

import (
	"bytes"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetGetGlobalFormatter(t *testing.T) {
	// Save original formatter
	originalFormatter := globalFormatter
	defer func() {
		globalFormatter = originalFormatter
	}()

	buf := &bytes.Buffer{}
	formatter := NewFormatter(FormatterOptions{
		Format: FormatHuman,
		Writer: buf,
	})

	SetGlobalFormatter(formatter)
	got := GetGlobalFormatter()

	assert.Equal(t, formatter, got)
}

func TestGetGlobalFormatter_Default(t *testing.T) {
	// Save original formatter
	originalFormatter := globalFormatter
	defer func() {
		globalFormatter = originalFormatter
	}()

	// Reset global formatter
	globalFormatter = nil

	formatter := GetGlobalFormatter()
	require.NotNil(t, formatter)
	assert.IsType(t, &HumanFormatter{}, formatter)
}

func TestGlobalInfo(t *testing.T) {
	// Save original formatter
	originalFormatter := globalFormatter
	defer func() {
		globalFormatter = originalFormatter
	}()

	buf := &bytes.Buffer{}
	formatter := NewFormatter(FormatterOptions{
		Format:  FormatHuman,
		Writer:  buf,
		NoColor: true,
	})
	SetGlobalFormatter(formatter)

	Info("test message")

	output := buf.String()
	assert.Contains(t, output, "test message")
}

func TestGlobalSuccess(t *testing.T) {
	// Save original formatter
	originalFormatter := globalFormatter
	defer func() {
		globalFormatter = originalFormatter
	}()

	buf := &bytes.Buffer{}
	formatter := NewFormatter(FormatterOptions{
		Format:  FormatHuman,
		Writer:  buf,
		NoColor: true,
	})
	SetGlobalFormatter(formatter)

	Success("operation completed")

	output := buf.String()
	assert.Contains(t, output, "operation completed")
	assert.Contains(t, output, "✓")
}

func TestGlobalWarning(t *testing.T) {
	// Save original formatter
	originalFormatter := globalFormatter
	defer func() {
		globalFormatter = originalFormatter
	}()

	buf := &bytes.Buffer{}
	formatter := NewFormatter(FormatterOptions{
		Format:  FormatHuman,
		Writer:  buf,
		NoColor: true,
	})
	SetGlobalFormatter(formatter)

	Warning("potential issue")

	output := buf.String()
	assert.Contains(t, output, "potential issue")
	assert.Contains(t, output, "⚠")
}

func TestGlobalError(t *testing.T) {
	// Save original formatter
	originalFormatter := globalFormatter
	defer func() {
		globalFormatter = originalFormatter
	}()

	buf := &bytes.Buffer{}
	formatter := NewFormatter(FormatterOptions{
		Format:  FormatHuman,
		Writer:  buf,
		NoColor: true,
	})
	SetGlobalFormatter(formatter)

	Error("operation failed")

	output := buf.String()
	assert.Contains(t, output, "operation failed")
	assert.Contains(t, output, "✗")
}

func TestGlobalData(t *testing.T) {
	// Save original formatter
	originalFormatter := globalFormatter
	defer func() {
		globalFormatter = originalFormatter
	}()

	buf := &bytes.Buffer{}
	formatter := NewFormatter(FormatterOptions{
		Format:  FormatHuman,
		Writer:  buf,
		NoColor: true,
	})
	SetGlobalFormatter(formatter)

	Data("test data")

	output := buf.String()
	assert.Contains(t, output, "test data")
}

func TestGlobalTable(t *testing.T) {
	// Save original formatter
	originalFormatter := globalFormatter
	defer func() {
		globalFormatter = originalFormatter
	}()

	buf := &bytes.Buffer{}
	formatter := NewFormatter(FormatterOptions{
		Format:  FormatHuman,
		Writer:  buf,
		NoColor: true,
	})
	SetGlobalFormatter(formatter)

	headers := []string{"Name", "Version"}
	rows := [][]string{
		{"node", "20.0.0"},
		{"python", "3.11.0"},
	}

	Table(headers, rows)

	output := buf.String()
	assert.Contains(t, output, "Name")
	assert.Contains(t, output, "Version")
	assert.Contains(t, output, "node")
	assert.Contains(t, output, "20.0.0")
}

func TestGlobalWithFields(t *testing.T) {
	// Save original formatter
	originalFormatter := globalFormatter
	defer func() {
		globalFormatter = originalFormatter
	}()

	buf := &bytes.Buffer{}
	formatter := NewFormatter(FormatterOptions{
		Format:  FormatHuman,
		Writer:  buf,
		NoColor: true,
		Verbose: true,
	})
	SetGlobalFormatter(formatter)

	fields := map[string]interface{}{
		"tool":    "node",
		"version": "20.0.0",
	}

	Info("installation complete", fields)

	output := buf.String()
	assert.Contains(t, output, "installation complete")
	assert.Contains(t, output, "tool=node")
	assert.Contains(t, output, "version=20.0.0")
}

func TestIsColorSupported(t *testing.T) {
	// Save original environment
	originalNoColor := env.Get("NO_COLOR")
	originalTerm := env.Get("TERM")
	defer func() {
		t.Setenv("NO_COLOR", originalNoColor)
		t.Setenv("TERM", originalTerm)
	}()

	tests := []struct {
		name    string
		noColor string
		term    string
		want    bool
	}{
		{
			name:    "NO_COLOR set",
			noColor: "1",
			term:    "xterm-256color",
			want:    false,
		},
		{
			name:    "TERM empty",
			noColor: "",
			term:    "",
			want:    false,
		},
		{
			name:    "TERM dumb",
			noColor: "",
			term:    "dumb",
			want:    false,
		},
		{
			name:    "TERM xterm",
			noColor: "",
			term:    "xterm-256color",
			want:    false, // Will be false because stdout is not a terminal in tests
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("NO_COLOR", tt.noColor)
			t.Setenv("TERM", tt.term)

			got := IsColorSupported()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGlobalFormatter_ThreadSafety(t *testing.T) {
	// Save original formatter
	originalFormatter := globalFormatter
	defer func() {
		globalFormatter = originalFormatter
	}()

	// Test concurrent access
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func() {
			buf := &bytes.Buffer{}
			formatter := NewFormatter(FormatterOptions{
				Format: FormatHuman,
				Writer: buf,
			})
			SetGlobalFormatter(formatter)
			_ = GetGlobalFormatter()
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Should not panic
	formatter := GetGlobalFormatter()
	require.NotNil(t, formatter)
}

func TestGlobalFormatter_JSONOutput(t *testing.T) {
	// Save original formatter
	originalFormatter := globalFormatter
	defer func() {
		globalFormatter = originalFormatter
	}()

	buf := &bytes.Buffer{}
	formatter := NewFormatter(FormatterOptions{
		Format: FormatJSON,
		Writer: buf,
	})
	SetGlobalFormatter(formatter)

	Info("test message")

	output := buf.String()
	assert.Contains(t, output, `"level"`)
	assert.Contains(t, output, `"message"`)
	assert.Contains(t, output, `"test message"`)
}
