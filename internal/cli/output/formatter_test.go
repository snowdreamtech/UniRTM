// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package output

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFormatter(t *testing.T) {
	tests := []struct {
		name     string
		opts     FormatterOptions
		wantType interface{}
	}{
		{
			name: "human formatter",
			opts: FormatterOptions{
				Format: FormatHuman,
			},
			wantType: &HumanFormatter{},
		},
		{
			name: "json formatter",
			opts: FormatterOptions{
				Format: FormatJSON,
			},
			wantType: &JSONFormatter{},
		},
		{
			name: "default formatter",
			opts: FormatterOptions{
				Format: "",
			},
			wantType: &HumanFormatter{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewFormatter(tt.opts)
			require.NotNil(t, formatter)
			assert.IsType(t, tt.wantType, formatter)
		})
	}
}

func TestDefaultFormatter(t *testing.T) {
	formatter := DefaultFormatter()
	require.NotNil(t, formatter)
	assert.IsType(t, &HumanFormatter{}, formatter)
}

func TestMergeFields(t *testing.T) {
	tests := []struct {
		name   string
		fields []map[string]interface{}
		want   map[string]interface{}
	}{
		{
			name:   "empty fields",
			fields: []map[string]interface{}{},
			want:   map[string]interface{}{},
		},
		{
			name: "single field map",
			fields: []map[string]interface{}{
				{"key1": "value1", "key2": "value2"},
			},
			want: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name: "multiple field maps",
			fields: []map[string]interface{}{
				{"key1": "value1"},
				{"key2": "value2"},
			},
			want: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name: "overlapping keys - last wins",
			fields: []map[string]interface{}{
				{"key1": "value1"},
				{"key1": "value2"},
			},
			want: map[string]interface{}{
				"key1": "value2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeFields(tt.fields...)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFormatJSON(t *testing.T) {
	tests := []struct {
		name    string
		data    interface{}
		wantErr bool
	}{
		{
			name: "simple map",
			data: map[string]interface{}{
				"key": "value",
			},
			wantErr: false,
		},
		{
			name: "nested structure",
			data: map[string]interface{}{
				"key1": "value1",
				"key2": map[string]interface{}{
					"nested": "value",
				},
			},
			wantErr: false,
		},
		{
			name:    "string",
			data:    "test string",
			wantErr: false,
		},
		{
			name:    "number",
			data:    42,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := formatJSON(tt.data)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, got)
			}
		})
	}
}

func TestFormatterSetWriter(t *testing.T) {
	buf := &bytes.Buffer{}

	// Test HumanFormatter
	humanFormatter := NewFormatter(FormatterOptions{
		Format: FormatHuman,
	})
	humanFormatter.SetWriter(buf)
	humanFormatter.Info("test message")
	assert.NotEmpty(t, buf.String())

	// Test JSONFormatter
	buf.Reset()
	jsonFormatter := NewFormatter(FormatterOptions{
		Format: FormatJSON,
	})
	jsonFormatter.SetWriter(buf)
	jsonFormatter.Info("test message")
	assert.NotEmpty(t, buf.String())
}
