// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package output

import (
	"os"
	"sync"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
)

var (
	// globalFormatter is the global formatter instance
	globalFormatter Formatter
	// formatterMutex protects access to globalFormatter
	formatterMutex sync.RWMutex
)

// SetGlobalFormatter sets the global formatter instance
func SetGlobalFormatter(formatter Formatter) {
	formatterMutex.Lock()
	defer formatterMutex.Unlock()
	globalFormatter = formatter
}

// GetGlobalFormatter returns the global formatter instance
// If no formatter has been set, it returns a default human formatter
func GetGlobalFormatter() Formatter {
	formatterMutex.RLock()
	defer formatterMutex.RUnlock()

	if globalFormatter == nil {
		return DefaultFormatter()
	}

	return globalFormatter
}

// Info outputs an informational message using the global formatter
func Info(message string, fields ...map[string]interface{}) {
	GetGlobalFormatter().Info(message, fields...)
}

// Success outputs a success message using the global formatter
func Success(message string, fields ...map[string]interface{}) {
	GetGlobalFormatter().Success(message, fields...)
}

// Warning outputs a warning message using the global formatter
func Warning(message string, fields ...map[string]interface{}) {
	GetGlobalFormatter().Warning(message, fields...)
}

// Error outputs an error message using the global formatter
func Error(message string, fields ...map[string]interface{}) {
	GetGlobalFormatter().Error(message, fields...)
}

// Data outputs structured data using the global formatter
func Data(data interface{}) {
	GetGlobalFormatter().Data(data)
}

// Table outputs tabular data using the global formatter
func Table(headers []string, rows [][]string) {
	GetGlobalFormatter().Table(headers, rows)
}

// IsColorSupported checks if the terminal supports color output
func IsColorSupported() bool {
	// Check if NO_COLOR environment variable is set
	if env.Get("NO_COLOR") != "" {
		return false
	}

	// Check if TERM is set to a color-supporting terminal
	term := env.Get("TERM")
	if term == "" || term == "dumb" {
		return false
	}

	// Check if stdout is a terminal
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}

	// Check if it's a character device (terminal)
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}
