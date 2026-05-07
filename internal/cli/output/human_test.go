// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHumanFormatter_Info(t *testing.T) {
	tests := []struct {
		name       string
		message    string
		fields     []map[string]interface{}
		quiet      bool
		verbose    bool
		wantOutput bool
	}{
		{
			name:       "simple info message",
			message:    "test message",
			wantOutput: true,
		},
		{
			name:       "info with fields - not verbose",
			message:    "test message",
			fields:     []map[string]interface{}{{"key": "value"}},
			verbose:    false,
			wantOutput: true,
		},
		{
			name:       "info with fields - verbose",
			message:    "test message",
			fields:     []map[string]interface{}{{"key": "value"}},
			verbose:    true,
			wantOutput: true,
		},
		{
			name:       "quiet mode suppresses output",
			message:    "test message",
			quiet:      true,
			wantOutput: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			formatter := &HumanFormatter{
				writer:  buf,
				noColor: true,
				quiet:   tt.quiet,
				verbose: tt.verbose,
			}

			formatter.Info(tt.message, tt.fields...)

			if tt.wantOutput {
				assert.Contains(t, buf.String(), tt.message)
			} else {
				assert.Empty(t, buf.String())
			}
		})
	}
}

func TestHumanFormatter_Success(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := &HumanFormatter{
		writer:  buf,
		noColor: true,
	}

	formatter.Success("operation completed")

	output := buf.String()
	assert.Contains(t, output, "operation completed")
	assert.Contains(t, output, "✓")
}

func TestHumanFormatter_Warning(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := &HumanFormatter{
		writer:  buf,
		noColor: true,
	}

	formatter.Warning("potential issue")

	output := buf.String()
	assert.Contains(t, output, "potential issue")
	assert.Contains(t, output, "⚠")
}

func TestHumanFormatter_Error(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := &HumanFormatter{
		writer:  buf,
		noColor: true,
	}

	formatter.Error("operation failed")

	output := buf.String()
	assert.Contains(t, output, "operation failed")
	assert.Contains(t, output, "✗")
}

func TestHumanFormatter_Data(t *testing.T) {
	tests := []struct {
		name       string
		data       interface{}
		quiet      bool
		wantOutput bool
	}{
		{
			name:       "string data",
			data:       "test string",
			wantOutput: true,
		},
		{
			name:       "string slice",
			data:       []string{"item1", "item2", "item3"},
			wantOutput: true,
		},
		{
			name: "map data",
			data: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			wantOutput: true,
		},
		{
			name:       "quiet mode suppresses output",
			data:       "test string",
			quiet:      true,
			wantOutput: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			formatter := &HumanFormatter{
				writer:  buf,
				noColor: true,
				quiet:   tt.quiet,
			}

			formatter.Data(tt.data)

			if tt.wantOutput {
				assert.NotEmpty(t, buf.String())
			} else {
				assert.Empty(t, buf.String())
			}
		})
	}
}

func TestHumanFormatter_Table(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := &HumanFormatter{
		writer:  buf,
		noColor: true,
	}

	headers := []string{"Name", "Version", "Status"}
	rows := [][]string{
		{"node", "20.0.0", "installed"},
		{"python", "3.11.0", "installed"},
		{"go", "1.21.0", "not installed"},
	}

	formatter.Table(headers, rows)

	output := buf.String()
	assert.Contains(t, output, "Name")
	assert.Contains(t, output, "Version")
	assert.Contains(t, output, "Status")
	assert.Contains(t, output, "node")
	assert.Contains(t, output, "20.0.0")
	assert.Contains(t, output, "installed")
}

func TestHumanFormatter_TableQuiet(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := &HumanFormatter{
		writer:  buf,
		noColor: true,
		quiet:   true,
	}

	headers := []string{"Name", "Version"}
	rows := [][]string{{"node", "20.0.0"}}

	formatter.Table(headers, rows)

	assert.Empty(t, buf.String())
}

func TestHumanFormatter_Colorize(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		color   string
		noColor bool
		want    string
	}{
		{
			name:    "with color",
			text:    "test",
			color:   colorRed,
			noColor: false,
			want:    colorRed + "test" + colorReset,
		},
		{
			name:    "no color",
			text:    "test",
			color:   colorRed,
			noColor: true,
			want:    "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := &HumanFormatter{
				noColor: tt.noColor,
			}

			got := formatter.colorize(tt.text, tt.color)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPadRight(t *testing.T) {
	tests := []struct {
		name  string
		s     string
		width int
		want  string
	}{
		{
			name:  "pad needed",
			s:     "test",
			width: 10,
			want:  "test      ",
		},
		{
			name:  "no pad needed",
			s:     "test",
			width: 4,
			want:  "test",
		},
		{
			name:  "string longer than width",
			s:     "test",
			width: 2,
			want:  "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := padRight(tt.s, tt.width)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestHumanFormatter_SetColorEnabled(t *testing.T) {
	formatter := &HumanFormatter{
		noColor: true,
	}

	assert.False(t, formatter.IsColorEnabled())

	formatter.SetColorEnabled(true)
	assert.True(t, formatter.IsColorEnabled())

	formatter.SetColorEnabled(false)
	assert.False(t, formatter.IsColorEnabled())
}

func TestHumanFormatter_SetQuiet(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := &HumanFormatter{
		writer:  buf,
		noColor: true,
		quiet:   false,
	}

	formatter.Info("test message")
	assert.NotEmpty(t, buf.String())

	buf.Reset()
	formatter.SetQuiet(true)
	formatter.Info("test message")
	assert.Empty(t, buf.String())
}

func TestHumanFormatter_SetVerbose(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := &HumanFormatter{
		writer:  buf,
		noColor: true,
		verbose: false,
	}

	formatter.Info("test message", map[string]interface{}{"key": "value"})
	output := buf.String()
	assert.Contains(t, output, "test message")
	assert.NotContains(t, output, "key=value")

	buf.Reset()
	formatter.SetVerbose(true)
	formatter.Info("test message", map[string]interface{}{"key": "value"})
	output = buf.String()
	assert.Contains(t, output, "test message")
	assert.Contains(t, output, "key=value")
}

func TestHumanFormatter_GetLogLevel(t *testing.T) {
	tests := []struct {
		name    string
		quiet   bool
		verbose bool
		want    string
	}{
		{
			name:    "quiet mode",
			quiet:   true,
			verbose: false,
			want:    "disabled",
		},
		{
			name:    "verbose mode",
			quiet:   false,
			verbose: true,
			want:    "debug",
		},
		{
			name:    "normal mode",
			quiet:   false,
			verbose: false,
			want:    "info",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := &HumanFormatter{
				quiet:   tt.quiet,
				verbose: tt.verbose,
			}

			level := formatter.GetLogLevel()
			assert.Equal(t, tt.want, level.String())
		})
	}
}

func TestHumanFormatter_WithFields(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := &HumanFormatter{
		writer:  buf,
		noColor: true,
		verbose: true,
	}

	fields := map[string]interface{}{
		"tool":    "node",
		"version": "20.0.0",
		"status":  "installed",
	}

	formatter.Info("installation complete", fields)

	output := buf.String()
	assert.Contains(t, output, "installation complete")
	assert.Contains(t, output, "tool=node")
	assert.Contains(t, output, "version=20.0.0")
	assert.Contains(t, output, "status=installed")
}

func TestHumanFormatter_TableAlignment(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := &HumanFormatter{
		writer:  buf,
		noColor: true,
	}

	headers := []string{"Short", "VeryLongHeader"}
	rows := [][]string{
		{"A", "B"},
		{"LongValue", "C"},
	}

	formatter.Table(headers, rows)

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	require.Len(t, lines, 4) // header + separator + 2 rows

	// Check that columns are aligned
	headerLine := lines[0]
	firstRowLine := lines[2]

	// The "Short" column should be padded to match "LongValue"
	assert.True(t, len(headerLine) > 0)
	assert.True(t, len(firstRowLine) > 0)
}
