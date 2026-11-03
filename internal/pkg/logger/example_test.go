// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package logger_test

import (
	"errors"
	"fmt"

	"github.com/snowdreamtech/unirtm/internal/pkg/logger"
)

// ExampleInitUniRTMLogger demonstrates how to initialize the UniRTM logger
// with rotating file writers for operation and error logs.
func ExampleInitUniRTMLogger() {
	// Initialize logger with custom paths
	opWriter, errWriter := logger.InitUniRTMLogger("unirtm.log", "error.log")

	fmt.Printf("Operation writer: %v\n", opWriter != nil)
	fmt.Printf("Error writer: %v\n", errWriter != nil)

	// Output:
	// Operation writer: true
	// Error writer: true
}

// ExampleInfo demonstrates basic info-level logging.
func ExampleInfo() {
	// Initialize logger
	logger.InitUniRTMLogger("unirtm.log", "error.log")

	// Log a simple info message
	logger.Info("Application started successfully")

	// Log with context fields
	logger.Info("Tool installed", map[string]interface{}{
		"tool":    "node",
		"version": "20.0.0",
		"path":    "/usr/local/bin/node",
	})
}

// ExampleError demonstrates error-level logging with automatic stack trace capture.
func ExampleError() {
	// Initialize logger
	logger.InitUniRTMLogger("unirtm.log", "error.log")

	// Log a simple error message (stack trace is automatically captured)
	logger.Error("Failed to connect to database")

	// Log with context fields
	logger.Error("Installation failed", map[string]interface{}{
		"tool":    "python",
		"version": "3.11.0",
		"reason":  "checksum mismatch",
	})
}

// ExampleErrorWithErr demonstrates logging errors with error objects.
func ExampleErrorWithErr() {
	// Initialize logger
	logger.InitUniRTMLogger("unirtm.log", "error.log")

	// Create an error
	err := errors.New("network timeout")

	// Log error with message
	logger.ErrorWithErr(err, "Failed to download artifact")

	// Log error with context fields
	logger.ErrorWithErr(err, "Download failed", map[string]interface{}{
		"url":     "https://example.com/artifact.tar.gz",
		"retries": 5,
		"timeout": "30s",
	})
}

// ExampleWithContext demonstrates structured logging with context fields.
func ExampleWithContext() {
	// Initialize logger
	logger.InitUniRTMLogger("unirtm.log", "error.log")

	// Create a logger with context
	contextLogger := logger.WithContext(map[string]interface{}{
		"request_id": "req-123456",
		"user_id":    "user-789",
		"operation":  "install",
	})

	// Use the context logger
	contextLogger.Info().Msg("Starting installation")
	contextLogger.Debug().Msg("Downloading artifact")
	contextLogger.Info().Msg("Installation completed")
}

// ExampleWithFields demonstrates structured logging with variadic key-value pairs.
func ExampleWithFields() {
	// Initialize logger
	logger.InitUniRTMLogger("unirtm.log", "error.log")

	// Create a logger with fields
	fieldsLogger := logger.WithFields(
		"tool", "go",
		"version", "1.21.0",
		"backend", "github",
	)

	// Use the fields logger
	fieldsLogger.Info().Msg("Tool resolved")
	fieldsLogger.Debug().Msg("Downloading from GitHub")
	fieldsLogger.Info().Msg("Installation successful")
}
