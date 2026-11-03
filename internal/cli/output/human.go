// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/rs/zerolog"
)

// HumanFormatter implements human-readable output with color support
type HumanFormatter struct {
	writer  io.Writer
	noColor bool
	quiet   bool
	verbose bool
}

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
	colorBold   = "\033[1m"
)

// colorize wraps text in ANSI color codes if colors are enabled
func (h *HumanFormatter) colorize(text string, color string) string {
	if h.noColor {
		return text
	}
	return color + text + colorReset
}

// Info outputs an informational message
func (h *HumanFormatter) Info(message string, fields ...map[string]interface{}) {
	if h.quiet {
		return
	}

	fmt.Fprintf(h.writer, "ℹ %s", message)

	if len(fields) > 0 && h.verbose {
		h.printFields(mergeFields(fields...))
	}

	fmt.Fprintln(h.writer)
}

// Infof outputs an informational message with string formatting
func (h *HumanFormatter) Infof(format string, a ...interface{}) {
	h.Info(fmt.Sprintf(format, a...))
}

// Success outputs a success message
func (h *HumanFormatter) Success(message string, fields ...map[string]interface{}) {
	prefix := h.colorize("✓", colorGreen)
	fmt.Fprintf(h.writer, "%s %s", prefix, h.colorize(message, colorGreen))

	if len(fields) > 0 && h.verbose {
		h.printFields(mergeFields(fields...))
	}

	fmt.Fprintln(h.writer)
}

// Successf outputs a success message with string formatting
func (h *HumanFormatter) Successf(format string, a ...interface{}) {
	h.Success(fmt.Sprintf(format, a...))
}

// Warning outputs a warning message
func (h *HumanFormatter) Warning(message string, fields ...map[string]interface{}) {
	prefix := h.colorize("⚠", colorYellow)
	fmt.Fprintf(h.writer, "%s %s", prefix, h.colorize(message, colorYellow))

	if len(fields) > 0 && h.verbose {
		h.printFields(mergeFields(fields...))
	}

	fmt.Fprintln(h.writer)
}

// Warningf outputs a warning message with string formatting
func (h *HumanFormatter) Warningf(format string, a ...interface{}) {
	h.Warning(fmt.Sprintf(format, a...))
}

// Error outputs an error message
func (h *HumanFormatter) Error(message string, fields ...map[string]interface{}) {
	prefix := h.colorize("✗", colorRed)
	fmt.Fprintf(h.writer, "%s %s", prefix, h.colorize(message, colorRed))

	if len(fields) > 0 {
		h.printFields(mergeFields(fields...))
	}

	fmt.Fprintln(h.writer)
}

// Errorf outputs an error message with string formatting
func (h *HumanFormatter) Errorf(format string, a ...interface{}) {
	h.Error(fmt.Sprintf(format, a...))
}

// Data outputs structured data in a human-readable format
func (h *HumanFormatter) Data(data interface{}) {
	if h.quiet {
		return
	}

	// For simple types, just print them
	switch v := data.(type) {
	case string:
		fmt.Fprintln(h.writer, v)
	case []string:
		for _, item := range v {
			fmt.Fprintln(h.writer, item)
		}
	case map[string]interface{}:
		h.printFields(v)
	default:
		// For complex types, fall back to JSON formatting
		jsonStr, err := formatJSON(data)
		if err != nil {
			fmt.Fprintf(h.writer, "%v\n", data)
		} else {
			fmt.Fprintln(h.writer, jsonStr)
		}
	}
}

// Table outputs tabular data
func (h *HumanFormatter) Table(headers []string, rows [][]string) {
	if h.quiet {
		return
	}

	// Calculate column widths
	colWidths := make([]int, len(headers))
	for i, header := range headers {
		colWidths[i] = len(header)
	}

	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	// Print header
	headerParts := make([]string, len(headers))
	for i, header := range headers {
		headerParts[i] = h.colorize(padRight(header, colWidths[i]), colorBold)
	}
	fmt.Fprintln(h.writer, strings.Join(headerParts, "  "))

	// Print separator
	separatorParts := make([]string, len(headers))
	for i := range headers {
		separatorParts[i] = strings.Repeat("-", colWidths[i])
	}
	fmt.Fprintln(h.writer, h.colorize(strings.Join(separatorParts, "  "), colorGray))

	// Print rows
	for _, row := range rows {
		rowParts := make([]string, len(headers))
		for i := range headers {
			if i < len(row) {
				rowParts[i] = padRight(row[i], colWidths[i])
			} else {
				rowParts[i] = padRight("", colWidths[i])
			}
		}
		fmt.Fprintln(h.writer, strings.Join(rowParts, "  "))
	}
}

// SetWriter sets the output writer
func (h *HumanFormatter) SetWriter(w io.Writer) {
	h.writer = w
}

// printFields prints key-value fields
func (h *HumanFormatter) printFields(fields map[string]interface{}) {
	if len(fields) == 0 {
		return
	}

	fmt.Fprint(h.writer, " ")
	first := true
	for key, value := range fields {
		if !first {
			fmt.Fprint(h.writer, ", ")
		}
		fmt.Fprintf(h.writer, "%s=%v", h.colorize(key, colorCyan), value)
		first = false
	}
}

// padRight pads a string to the right with spaces
func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

// IsColorEnabled returns true if color output is enabled
func (h *HumanFormatter) IsColorEnabled() bool {
	return !h.noColor
}

// SetColorEnabled enables or disables color output
func (h *HumanFormatter) SetColorEnabled(enabled bool) {
	h.noColor = !enabled
}

// SetQuiet enables or disables quiet mode
func (h *HumanFormatter) SetQuiet(quiet bool) {
	h.quiet = quiet
}

// SetVerbose enables or disables verbose mode
func (h *HumanFormatter) SetVerbose(verbose bool) {
	h.verbose = verbose
}

// GetLogLevel returns the appropriate zerolog level based on formatter settings
func (h *HumanFormatter) GetLogLevel() zerolog.Level {
	if h.quiet {
		return zerolog.Disabled
	}
	if h.verbose {
		return zerolog.DebugLevel
	}
	return zerolog.InfoLevel
}
