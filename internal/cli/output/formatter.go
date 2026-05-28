// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package output provides CLI output formatting capabilities including
// progress indicators, JSON output, and color-coded output.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/rs/zerolog"
)

// OutputFormat represents the output format type
type OutputFormat string

const (
	// FormatHuman is the human-readable output format with colors
	FormatHuman OutputFormat = "human"
	// FormatJSON is the JSON output format for scripting
	FormatJSON OutputFormat = "json"
)

// Formatter defines the interface for output formatting
type Formatter interface {
	// Info outputs an informational message
	Info(message string, fields ...map[string]interface{})

	// Infof outputs an informational message with string formatting
	Infof(format string, a ...interface{})

	// Success outputs a success message
	Success(message string, fields ...map[string]interface{})

	// Successf outputs a success message with string formatting
	Successf(format string, a ...interface{})

	// Warning outputs a warning message
	Warning(message string, fields ...map[string]interface{})

	// Warningf outputs a warning message with string formatting
	Warningf(format string, a ...interface{})

	// Error outputs an error message
	Error(message string, fields ...map[string]interface{})

	// Errorf outputs an error message with string formatting
	Errorf(format string, a ...interface{})

	// Data outputs structured data
	Data(data interface{})

	// Table outputs tabular data
	Table(headers []string, rows [][]string)

	// SetWriter sets the output writer
	SetWriter(w io.Writer)
}

// FormatterOptions contains options for creating a formatter
type FormatterOptions struct {
	// Format specifies the output format (human or json)
	Format OutputFormat

	// NoColor disables color output for human format
	NoColor bool

	// Color specifies the color mode (auto, always, never)
	Color string

	// Writer is the output writer (defaults to os.Stdout)
	Writer io.Writer

	// Quiet suppresses non-essential output
	Quiet bool

	// Verbose enables verbose output
	Verbose bool
}

// NewFormatter creates a new formatter based on the specified options
func NewFormatter(opts FormatterOptions) Formatter {
	if opts.Writer == nil {
		opts.Writer = os.Stdout
	}

	switch opts.Format {
	case FormatJSON:
		return &JSONFormatter{
			writer: opts.Writer,
			quiet:  opts.Quiet,
		}
	case FormatHuman:
		fallthrough
	default:
		noColor := opts.NoColor || zerolog.GlobalLevel() == zerolog.Disabled
		if opts.Color == "always" {
			noColor = false
		} else if opts.Color == "never" {
			noColor = true
		}
		return &HumanFormatter{
			writer:  opts.Writer,
			noColor: noColor,
			quiet:   opts.Quiet,
			verbose: opts.Verbose,
		}
	}
}

// DefaultFormatter returns a formatter with default settings
func DefaultFormatter() Formatter {
	return NewFormatter(FormatterOptions{
		Format: FormatHuman,
		Writer: os.Stdout,
	})
}

// mergeFields merges multiple field maps into a single map
func mergeFields(fieldMaps ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, fields := range fieldMaps {
		for k, v := range fields {
			result[k] = v
		}
	}
	return result
}

// formatJSON formats data as JSON
func formatJSON(data interface{}) (string, error) {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal JSON: %w", err)
	}
	return string(bytes), nil
}
