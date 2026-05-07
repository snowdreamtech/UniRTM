// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONFormatter_Info(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := &JSONFormatter{
		writer: buf,
	}

	formatter.Info("test message")

	var msg OutputMessage
	err := json.Unmarshal(buf.Bytes(), &msg)
	require.NoError(t, err)

	assert.Equal(t, "info", msg.Level)
	assert.Equal(t, "test message", msg.Message)
	assert.NotEmpty(t, msg.Timestamp)
}

func TestJSONFormatter_InfoWithFields(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := &JSONFormatter{
		writer: buf,
	}

	fields := map[string]interface{}{
		"tool":    "node",
		"version": "20.0.0",
	}

	formatter.Info("installation started", fields)

	var msg OutputMessage
	err := json.Unmarshal(buf.Bytes(), &msg)
	require.NoError(t, err)

	assert.Equal(t, "info", msg.Level)
	assert.Equal(t, "installation started", msg.Message)
	assert.Equal(t, "node", msg.Fields["tool"])
	assert.Equal(t, "20.0.0", msg.Fields["version"])
}

func TestJSONFormatter_Success(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := &JSONFormatter{
		writer: buf,
	}

	formatter.Success("operation completed")

	var msg OutputMessage
	err := json.Unmarshal(buf.Bytes(), &msg)
	require.NoError(t, err)

	assert.Equal(t, "success", msg.Level)
	assert.Equal(t, "operation completed", msg.Message)
}

func TestJSONFormatter_Warning(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := &JSONFormatter{
		writer: buf,
	}

	formatter.Warning("potential issue")

	var msg OutputMessage
	err := json.Unmarshal(buf.Bytes(), &msg)
	require.NoError(t, err)

	assert.Equal(t, "warning", msg.Level)
	assert.Equal(t, "potential issue", msg.Message)
}

func TestJSONFormatter_Error(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := &JSONFormatter{
		writer: buf,
	}

	formatter.Error("operation failed")

	var msg OutputMessage
	err := json.Unmarshal(buf.Bytes(), &msg)
	require.NoError(t, err)

	assert.Equal(t, "error", msg.Level)
	assert.Equal(t, "operation failed", msg.Message)
}

func TestJSONFormatter_Data(t *testing.T) {
	tests := []struct {
		name string
		data interface{}
	}{
		{
			name: "string data",
			data: "test string",
		},
		{
			name: "map data",
			data: map[string]interface{}{
				"key1": "value1",
				"key2": 42,
			},
		},
		{
			name: "slice data",
			data: []string{"item1", "item2", "item3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			formatter := &JSONFormatter{
				writer: buf,
			}

			formatter.Data(tt.data)

			var output DataOutput
			err := json.Unmarshal(buf.Bytes(), &output)
			require.NoError(t, err)

			assert.NotEmpty(t, output.Timestamp)
			assert.NotNil(t, output.Data)
		})
	}
}

func TestJSONFormatter_Table(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := &JSONFormatter{
		writer: buf,
	}

	headers := []string{"Name", "Version", "Status"}
	rows := [][]string{
		{"node", "20.0.0", "installed"},
		{"python", "3.11.0", "installed"},
	}

	formatter.Table(headers, rows)

	var output TableOutput
	err := json.Unmarshal(buf.Bytes(), &output)
	require.NoError(t, err)

	assert.NotEmpty(t, output.Timestamp)
	assert.Equal(t, headers, output.Headers)
	assert.Equal(t, rows, output.Rows)
}

func TestJSONFormatter_Quiet(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := &JSONFormatter{
		writer: buf,
		quiet:  true,
	}

	formatter.Info("test message")
	assert.Empty(t, buf.String())

	formatter.Data("test data")
	assert.Empty(t, buf.String())

	formatter.Table([]string{"Header"}, [][]string{{"Value"}})
	assert.Empty(t, buf.String())

	// Success, Warning, and Error should still output in quiet mode
	formatter.Success("success")
	assert.NotEmpty(t, buf.String())
}

func TestJSONFormatter_SetWriter(t *testing.T) {
	buf1 := &bytes.Buffer{}
	buf2 := &bytes.Buffer{}

	formatter := &JSONFormatter{
		writer: buf1,
	}

	formatter.Info("message 1")
	assert.NotEmpty(t, buf1.String())
	assert.Empty(t, buf2.String())

	formatter.SetWriter(buf2)
	formatter.Info("message 2")
	assert.Contains(t, buf1.String(), "message 1")
	assert.NotContains(t, buf1.String(), "message 2")
	assert.Contains(t, buf2.String(), "message 2")
}

func TestJSONFormatter_SetQuiet(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := &JSONFormatter{
		writer: buf,
		quiet:  false,
	}

	formatter.Info("test message")
	assert.NotEmpty(t, buf.String())

	buf.Reset()
	formatter.SetQuiet(true)
	formatter.Info("test message")
	assert.Empty(t, buf.String())
}

func TestJSONFormatter_MultipleFields(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := &JSONFormatter{
		writer: buf,
	}

	fields1 := map[string]interface{}{"key1": "value1"}
	fields2 := map[string]interface{}{"key2": "value2"}

	formatter.Info("test message", fields1, fields2)

	var msg OutputMessage
	err := json.Unmarshal(buf.Bytes(), &msg)
	require.NoError(t, err)

	assert.Equal(t, "value1", msg.Fields["key1"])
	assert.Equal(t, "value2", msg.Fields["key2"])
}

func TestJSONFormatter_ValidJSON(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := &JSONFormatter{
		writer: buf,
	}

	// Test all output methods produce valid JSON
	formatter.Info("info message")
	formatter.Success("success message")
	formatter.Warning("warning message")
	formatter.Error("error message")
	formatter.Data(map[string]interface{}{"key": "value"})
	formatter.Table([]string{"Header"}, [][]string{{"Value"}})

	// Split by newlines and validate each JSON object
	output := strings.TrimSpace(buf.String())
	if output == "" {
		t.Fatal("expected output, got empty string")
	}

	// Use a JSON decoder to parse each complete JSON object
	decoder := json.NewDecoder(strings.NewReader(output))
	count := 0
	for decoder.More() {
		var obj interface{}
		err := decoder.Decode(&obj)
		assert.NoError(t, err, "JSON object %d should be valid", count)
		count++
	}

	// We should have 6 JSON objects (one for each formatter call)
	assert.Equal(t, 6, count, "expected 6 JSON objects")
}

func TestJSONFormatter_TimestampFormat(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := &JSONFormatter{
		writer: buf,
	}

	formatter.Info("test message")

	var msg OutputMessage
	err := json.Unmarshal(buf.Bytes(), &msg)
	require.NoError(t, err)

	// Verify timestamp is in RFC3339 format
	assert.Regexp(t, `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`, msg.Timestamp)
}
