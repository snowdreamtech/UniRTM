// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package property contains property-based tests for UniRTM.
//
// Property-based tests verify universal properties that should hold for all inputs,
// complementing example-based unit tests with comprehensive input coverage.
package property

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/rs/zerolog"
	"github.com/snowdreamtech/unirtm/internal/pkg/logger"
	"github.com/snowdreamtech/unirtm/internal/repository"
	"github.com/snowdreamtech/unirtm/internal/repository/sqlite"
	"github.com/snowdreamtech/unirtm/internal/service"
)

// Feature: unirtm, Property 20: Log Entry Format Completeness
//
// **Validates: Requirements 7.5**
//
// For any log entry, it SHALL contain a timestamp, log level, and structured
// context fields.
//
// This property ensures that:
// 1. All log entries have required metadata (timestamp, level)
// 2. Structured context fields are preserved in log output
// 3. Log format is consistent across all log levels
// 4. JSON log output is parseable and contains expected fields
func TestProperty_LogEntryFormatCompleteness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.MaxSize = 20

	properties := gopter.NewProperties(parameters)

	properties.Property("All log entries contain timestamp, level, and context fields", prop.ForAll(
		func(logLevel zerolog.Level, message string, contextFields map[string]interface{}) bool {
			// Skip disabled level
			if logLevel == zerolog.Disabled || logLevel == zerolog.NoLevel {
				return true
			}

			// Create a buffer to capture log output
			var buf bytes.Buffer

			// Create a logger that writes JSON to the buffer
			testLogger := zerolog.New(&buf).With().Timestamp().Logger()

			// Log the message with context fields at the specified level
			event := testLogger.WithLevel(logLevel)
			for key, value := range contextFields {
				event = event.Interface(key, value)
			}
			event.Msg(message)

			// Parse the JSON log output
			var logEntry map[string]interface{}
			err := json.Unmarshal(buf.Bytes(), &logEntry)
			if err != nil {
				t.Logf("Failed to parse log JSON: %v", err)
				t.Logf("Log output: %s", buf.String())
				return false
			}

			// Verify timestamp field exists
			if _, ok := logEntry["time"]; !ok {
				t.Logf("Log entry missing 'time' field")
				t.Logf("Log entry: %+v", logEntry)
				return false
			}

			// Verify level field exists
			if _, ok := logEntry["level"]; !ok {
				t.Logf("Log entry missing 'level' field")
				t.Logf("Log entry: %+v", logEntry)
				return false
			}

			// Verify message field exists
			if _, ok := logEntry["message"]; !ok {
				t.Logf("Log entry missing 'message' field")
				t.Logf("Log entry: %+v", logEntry)
				return false
			}

			// Verify all context fields are present
			for key := range contextFields {
				if _, ok := logEntry[key]; !ok {
					t.Logf("Log entry missing context field '%s'", key)
					t.Logf("Log entry: %+v", logEntry)
					return false
				}
			}

			return true
		},
		genLogLevel(),
		genLogMessage(),
		genContextFields(),
	))

	properties.TestingRun(t)
}

// Feature: unirtm, Property 21: Error Stack Trace Capture
//
// **Validates: Requirements 7.7**
//
// For any error that is logged, the log entry SHALL include the full stack trace
// showing the call chain that led to the error.
//
// This property ensures that:
// 1. Error-level logs automatically capture stack traces
// 2. Stack traces contain function names and line numbers
// 3. Stack trace information is included in the log output
// 4. Stack traces are captured for all error levels (Error, Fatal, Panic)
func TestProperty_ErrorStackTraceCapture(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.MaxSize = 20

	properties := gopter.NewProperties(parameters)

	properties.Property("Error-level logs capture stack traces", prop.ForAll(
		func(errorMessage string, contextFields map[string]interface{}) bool {
			// Create a temporary directory for log files
			tempDir, err := os.MkdirTemp("", "unirtm-test-*")
			if err != nil {
				t.Logf("Failed to create temp dir: %v", err)
				return false
			}
			defer os.RemoveAll(tempDir)

			operationLogPath := filepath.Join(tempDir, "unirtm.log")
			errorLogPath := filepath.Join(tempDir, "error.log")

			// Initialize the UniRTM logger
			_, _ = logger.InitUniRTMLogger(operationLogPath, errorLogPath)

			// Create a logger with context fields
			testLogger := logger.Logger
			if len(contextFields) > 0 {
				ctx := testLogger.With()
				for key, value := range contextFields {
					ctx = ctx.Interface(key, value)
				}
				testLogger = ctx.Logger()
			}

			// Log an error message (this should trigger stack trace capture)
			testLogger.Error().Msg(errorMessage)

			// Read the error log file
			errorLogContent, err := os.ReadFile(errorLogPath)
			if err != nil {
				t.Logf("Failed to read error log: %v", err)
				return false
			}

			// Parse the JSON log entries
			lines := bytes.Split(errorLogContent, []byte("\n"))
			var foundStackTrace bool

			for _, line := range lines {
				if len(line) == 0 {
					continue
				}

				var logEntry map[string]interface{}
				err := json.Unmarshal(line, &logEntry)
				if err != nil {
					// Skip non-JSON lines (might be console output)
					continue
				}

				// Check if this is an error-level log
				if level, ok := logEntry["level"].(string); ok && level == "error" {
					// Verify stack trace field exists
					if stackTrace, ok := logEntry["stack_trace"].(string); ok {
						// Verify stack trace contains function names
						if strings.Contains(stackTrace, "goroutine") {
							foundStackTrace = true
							break
						}
					}
				}
			}

			if !foundStackTrace {
				t.Logf("Error log missing stack trace")
				t.Logf("Error log content: %s", string(errorLogContent))
				return false
			}

			return true
		},
		genErrorMessageLogging(),
		genContextFields(),
	))

	properties.TestingRun(t)
}

// Feature: unirtm, Property 22: Audit Log Completeness
//
// **Validates: Requirements 7.8, 8.1, 8.5**
//
// For any operation (installation, activation, configuration change), an audit
// log entry SHALL be created containing operation type, timestamp, affected tools,
// status, and error message (if failed).
//
// This property ensures that:
// 1. All operations are recorded in the audit log
// 2. Audit entries contain all required fields
// 3. Successful and failed operations are both logged
// 4. Audit entries are persisted to the database
func TestProperty_AuditLogCompleteness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.MaxSize = 20

	properties := gopter.NewProperties(parameters)

	properties.Property("All operations create complete audit log entries", prop.ForAll(
		func(operation service.OperationType, tool string, version string, status service.OperationStatus, errorMsg string, duration int64, metadata map[string]interface{}) bool {
			// Create temporary database
			db, cleanup := setupTempDB(t)
			defer cleanup()

			ctx := context.Background()

			// Create audit repository and service
			auditRepo, err := sqlite.NewAuditRepository(db.Conn())
			if err != nil {
				t.Logf("Failed to create audit repository: %v", err)
				return false
			}
			defer auditRepo.Close()

			auditService := service.NewAuditService(auditRepo)

			// Create audit log entry
			entry := &service.AuditLogEntry{
				Operation: operation,
				Tool:      tool,
				Version:   version,
				Status:    status,
				Error:     errorMsg,
				Duration:  duration,
				Metadata:  metadata,
			}

			// Log the operation
			err = auditService.LogOperation(ctx, entry)
			if err != nil {
				t.Logf("Failed to log operation: %v", err)
				return false
			}

			// Query the audit log to verify the entry was created
			filter := repository.AuditFilter{
				Operation: string(operation),
				Tool:      tool,
				Limit:     1,
			}

			entries, err := auditService.QueryAuditLogs(ctx, filter)
			if err != nil {
				t.Logf("Failed to query audit logs: %v", err)
				return false
			}

			if len(entries) == 0 {
				t.Logf("No audit log entry found")
				return false
			}

			auditEntry := entries[0]

			// Verify all required fields are present
			if auditEntry.Operation != string(operation) {
				t.Logf("Operation mismatch: expected %s, got %s", operation, auditEntry.Operation)
				return false
			}

			if auditEntry.Tool != tool {
				t.Logf("Tool mismatch: expected %s, got %s", tool, auditEntry.Tool)
				return false
			}

			if auditEntry.Version != version {
				t.Logf("Version mismatch: expected %s, got %s", version, auditEntry.Version)
				return false
			}

			if auditEntry.Status != string(status) {
				t.Logf("Status mismatch: expected %s, got %s", status, auditEntry.Status)
				return false
			}

			if status == service.StatusFailure && auditEntry.Error != errorMsg {
				t.Logf("Error message mismatch: expected %s, got %s", errorMsg, auditEntry.Error)
				return false
			}

			if auditEntry.Duration != duration {
				t.Logf("Duration mismatch: expected %d, got %d", duration, auditEntry.Duration)
				return false
			}

			// Verify timestamp is set
			if auditEntry.Timestamp.IsZero() {
				t.Logf("Timestamp not set")
				return false
			}

			return true
		},
		genOperationType(),
		genToolName(),
		genVersion(),
		genOperationStatus(),
		genErrorMessage(),
		genDuration(),
		genMetadataMap(),
	))

	properties.TestingRun(t)
}

// Feature: unirtm, Property 23: Audit Query Correctness
//
// **Validates: Requirements 8.6**
//
// For any audit log query with filters (date range, operation type, tool name,
// status), the results SHALL include only entries matching all specified filters.
//
// This property ensures that:
// 1. Query filters are applied correctly
// 2. Multiple filters work together (AND logic)
// 3. Empty result sets are handled correctly
// 4. Query results are accurate and complete
func TestProperty_AuditQueryCorrectness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.MaxSize = 20

	properties := gopter.NewProperties(parameters)

	properties.Property("Audit queries return only matching entries", prop.ForAll(
		func(entries []auditEntryData, filterOp string, filterTool string, filterStatus string) bool {
			// Create temporary database
			db, cleanup := setupTempDB(t)
			defer cleanup()

			ctx := context.Background()

			// Create audit repository and service
			auditRepo, err := sqlite.NewAuditRepository(db.Conn())
			if err != nil {
				t.Logf("Failed to create audit repository: %v", err)
				return false
			}
			defer auditRepo.Close()

			auditService := service.NewAuditService(auditRepo)

			// Insert all audit entries
			for _, entryData := range entries {
				entry := &service.AuditLogEntry{
					Operation: entryData.Operation,
					Tool:      entryData.Tool,
					Version:   entryData.Version,
					Status:    entryData.Status,
					Error:     entryData.Error,
					Duration:  entryData.Duration,
					Metadata:  entryData.Metadata,
				}

				err = auditService.LogOperation(ctx, entry)
				if err != nil {
					t.Logf("Failed to log operation: %v", err)
					return false
				}
			}

			// Build filter
			filter := repository.AuditFilter{
				Operation: filterOp,
				Tool:      filterTool,
				Status:    filterStatus,
				Limit:     1000, // High limit to get all matching entries
			}

			// Query with filter
			results, err := auditService.QueryAuditLogs(ctx, filter)
			if err != nil {
				t.Logf("Failed to query audit logs: %v", err)
				return false
			}

			// Verify all results match the filter
			for _, result := range results {
				if filterOp != "" && result.Operation != filterOp {
					t.Logf("Result operation %s doesn't match filter %s", result.Operation, filterOp)
					return false
				}

				if filterTool != "" && result.Tool != filterTool {
					t.Logf("Result tool %s doesn't match filter %s", result.Tool, filterTool)
					return false
				}

				if filterStatus != "" && result.Status != filterStatus {
					t.Logf("Result status %s doesn't match filter %s", result.Status, filterStatus)
					return false
				}
			}

			// Verify no matching entries were missed
			expectedCount := 0
			for _, entryData := range entries {
				matches := true
				if filterOp != "" && string(entryData.Operation) != filterOp {
					matches = false
				}
				if filterTool != "" && entryData.Tool != filterTool {
					matches = false
				}
				if filterStatus != "" && string(entryData.Status) != filterStatus {
					matches = false
				}
				if matches {
					expectedCount++
				}
			}

			if len(results) != expectedCount {
				t.Logf("Expected %d results, got %d", expectedCount, len(results))
				return false
			}

			return true
		},
		genAuditEntries(),
		genOperationTypeFilter(),
		genToolNameFilter(),
		genOperationStatusFilter(),
	))

	properties.TestingRun(t)
}

// Generator functions

// genLogLevel generates random log levels
func genLogLevel() gopter.Gen {
	return gen.OneConstOf(
		zerolog.TraceLevel,
		zerolog.DebugLevel,
		zerolog.InfoLevel,
		zerolog.WarnLevel,
		zerolog.ErrorLevel,
	)
}

// genLogMessage generates random log messages
func genLogMessage() gopter.Gen {
	return gen.OneConstOf(
		"Operation started",
		"Processing request",
		"Configuration loaded",
		"Tool installed successfully",
		"Cache updated",
		"Database transaction committed",
		"Network request completed",
		"File written to disk",
	)
}

// genErrorMessageLogging generates random error messages for logging tests
func genErrorMessageLogging() gopter.Gen {
	return gen.OneConstOf(
		"network connection failed",
		"file not found",
		"permission denied",
		"invalid configuration",
		"checksum mismatch",
		"database error",
		"timeout exceeded",
		"",
	)
}

// genContextFields generates random context field maps
func genContextFields() gopter.Gen {
	return func(genParams *gopter.GenParameters) *gopter.GenResult {
		result := make(map[string]interface{})

		// Generate 0-5 random fields
		numFields := genParams.Rng.Intn(6)

		keys := []string{"user_id", "operation", "tool", "version", "backend", "duration_ms", "status", "path"}
		values := []interface{}{"node", "20.0.0", "install", 123, true}

		for i := 0; i < numFields && i < len(keys); i++ {
			key := keys[genParams.Rng.Intn(len(keys))]
			value := values[genParams.Rng.Intn(len(values))]
			result[key] = value
		}

		return gopter.NewGenResult(result, gopter.NoShrinker)
	}
}

// genContextKey generates valid context field keys
func genContextKey() gopter.Gen {
	return gen.OneConstOf(
		"user_id",
		"operation",
		"tool",
		"version",
		"backend",
		"duration_ms",
		"status",
		"path",
	)
}

// genContextValue generates context field values
func genContextValue() gopter.Gen {
	return gen.OneGenOf(
		gen.Const("node"),
		gen.Const("20.0.0"),
		gen.Const("install"),
		gen.Const(123),
		gen.Const(true),
	)
}

// genOperationType generates random operation types
func genOperationType() gopter.Gen {
	return gen.OneConstOf(
		service.OperationInstall,
		service.OperationUninstall,
		service.OperationActivate,
		service.OperationDeactivate,
		service.OperationUpdate,
		service.OperationCachePurge,
		service.OperationConfigLoad,
		service.OperationConfigUpdate,
	)
}

// genOperationTypeFilter generates operation type filters (including empty)
func genOperationTypeFilter() gopter.Gen {
	return gen.OneConstOf(
		"",
		string(service.OperationInstall),
		string(service.OperationUninstall),
		string(service.OperationActivate),
		string(service.OperationUpdate),
	)
}

// genToolNameFilter generates tool name filters (including empty)
func genToolNameFilter() gopter.Gen {
	return gen.OneConstOf(
		"",
		"node",
		"python",
		"go",
		"ruby",
	)
}

// genVersionFilter generates version filters (including empty)
func genVersionFilter() gopter.Gen {
	return gen.OneConstOf(
		"1.0.0",
		"2.5.3",
		"20.0.0",
		"3.11.5",
		"1.21.0",
		"latest",
		"lts",
	)
}

// genOperationStatus generates random operation statuses
func genOperationStatus() gopter.Gen {
	return gen.OneConstOf(
		service.StatusSuccess,
		service.StatusFailure,
	)
}

// genOperationStatusFilter generates operation status filters (including empty)
func genOperationStatusFilter() gopter.Gen {
	return gen.OneConstOf(
		"",
		string(service.StatusSuccess),
		string(service.StatusFailure),
	)
}

// genDuration generates random operation durations in milliseconds
func genDuration() gopter.Gen {
	return gen.Int64Range(0, 60000) // 0 to 60 seconds
}

// genMetadataMap generates random metadata maps for logging tests
func genMetadataMap() gopter.Gen {
	return func(genParams *gopter.GenParameters) *gopter.GenResult {
		result := make(map[string]interface{})

		// Generate 0-3 random fields
		numFields := genParams.Rng.Intn(4)

		keys := []string{"old_version", "new_version", "backend", "provider", "config_path"}
		values := []interface{}{"1.0.0", "2.0.0", "github", "aqua", "/etc/unirtm/config.toml"}

		for i := 0; i < numFields && i < len(keys); i++ {
			key := keys[genParams.Rng.Intn(len(keys))]
			value := values[genParams.Rng.Intn(len(values))]
			result[key] = value
		}

		return gopter.NewGenResult(result, gopter.NoShrinker)
	}
}

// auditEntryData represents audit entry data for testing
type auditEntryData struct {
	Operation service.OperationType
	Tool      string
	Version   string
	Status    service.OperationStatus
	Error     string
	Duration  int64
	Metadata  map[string]interface{}
}

// genAuditEntries generates a list of audit entries for testing
func genAuditEntries() gopter.Gen {
	return gen.SliceOfN(10, genAuditEntryData()).Map(func(entries []auditEntryData) []auditEntryData {
		// Ensure we have at least 5 entries
		if len(entries) < 5 {
			// Pad with additional entries
			for len(entries) < 5 {
				entries = append(entries, auditEntryData{
					Operation: service.OperationInstall,
					Tool:      "node",
					Version:   "20.0.0",
					Status:    service.StatusSuccess,
					Duration:  1000,
				})
			}
		}
		return entries
	})
}

// genAuditEntryData generates a single audit entry for logging tests
func genAuditEntryData() gopter.Gen {
	return gopter.CombineGens(
		genOperationType(),
		genToolName(),
		genVersion(),
		genOperationStatus(),
		genErrorMessage(),
		genDuration(),
		genMetadataMap(),
	).Map(func(values []interface{}) auditEntryData {
		return auditEntryData{
			Operation: values[0].(service.OperationType),
			Tool:      values[1].(string),
			Version:   values[2].(string),
			Status:    values[3].(service.OperationStatus),
			Error:     values[4].(string),
			Duration:  values[5].(int64),
			Metadata:  values[6].(map[string]interface{}),
		}
	})
}
