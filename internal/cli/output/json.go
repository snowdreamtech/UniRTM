// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package output

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// JSONFormatter implements JSON output for scripting and automation
type JSONFormatter struct {
	writer io.Writer
	quiet  bool
}

// OutputMessage represents a JSON output message
type OutputMessage struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// DataOutput represents structured data output
type DataOutput struct {
	Timestamp string      `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// TableOutput represents tabular data output
type TableOutput struct {
	Timestamp string     `json:"timestamp"`
	Headers   []string   `json:"headers"`
	Rows      [][]string `json:"rows"`
}

// Info outputs an informational message in JSON format
func (j *JSONFormatter) Info(message string, fields ...map[string]interface{}) {
	if j.quiet {
		return
	}

	msg := OutputMessage{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     "info",
		Message:   message,
	}

	if len(fields) > 0 {
		msg.Fields = mergeFields(fields...)
	}

	j.writeJSON(msg)
}

// Infof outputs an informational message in JSON format with string formatting
func (j *JSONFormatter) Infof(format string, a ...interface{}) {
	j.Info(fmt.Sprintf(format, a...))
}

// Success outputs a success message in JSON format
func (j *JSONFormatter) Success(message string, fields ...map[string]interface{}) {
	msg := OutputMessage{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     "success",
		Message:   message,
	}

	if len(fields) > 0 {
		msg.Fields = mergeFields(fields...)
	}

	j.writeJSON(msg)
}

// Successf outputs a success message in JSON format with string formatting
func (j *JSONFormatter) Successf(format string, a ...interface{}) {
	j.Success(fmt.Sprintf(format, a...))
}

// Warning outputs a warning message in JSON format
func (j *JSONFormatter) Warning(message string, fields ...map[string]interface{}) {
	msg := OutputMessage{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     "warning",
		Message:   message,
	}

	if len(fields) > 0 {
		msg.Fields = mergeFields(fields...)
	}

	j.writeJSON(msg)
}

// Warningf outputs a warning message in JSON format with string formatting
func (j *JSONFormatter) Warningf(format string, a ...interface{}) {
	j.Warning(fmt.Sprintf(format, a...))
}

// Error outputs an error message in JSON format
func (j *JSONFormatter) Error(message string, fields ...map[string]interface{}) {
	msg := OutputMessage{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     "error",
		Message:   message,
	}

	if len(fields) > 0 {
		msg.Fields = mergeFields(fields...)
	}

	j.writeJSON(msg)
}

// Errorf outputs an error message in JSON format with string formatting
func (j *JSONFormatter) Errorf(format string, a ...interface{}) {
	j.Error(fmt.Sprintf(format, a...))
}

// Data outputs structured data in JSON format
func (j *JSONFormatter) Data(data interface{}) {
	if j.quiet {
		return
	}

	output := DataOutput{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Data:      data,
	}

	j.writeJSON(output)
}

// Table outputs tabular data in JSON format
func (j *JSONFormatter) Table(headers []string, rows [][]string) {
	if j.quiet {
		return
	}

	output := TableOutput{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Headers:   headers,
		Rows:      rows,
	}

	j.writeJSON(output)
}

// SetWriter sets the output writer
func (j *JSONFormatter) SetWriter(w io.Writer) {
	j.writer = w
}

// SetQuiet enables or disables quiet mode
func (j *JSONFormatter) SetQuiet(quiet bool) {
	j.quiet = quiet
}

// writeJSON writes a JSON object to the output writer
func (j *JSONFormatter) writeJSON(v interface{}) {
	encoder := json.NewEncoder(j.writer)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(v); err != nil {
		// Fallback to error message if encoding fails
		fmt.Fprintf(j.writer, `{"error": "failed to encode JSON: %s"}`+"\n", err)
	}
}
