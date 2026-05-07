package logger

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitUniRTMLogger(t *testing.T) {
	// Create temporary directory for test logs
	tmpDir := t.TempDir()
	operationLogPath := filepath.Join(tmpDir, "unirtm.log")
	errorLogPath := filepath.Join(tmpDir, "error.log")

	// Initialize logger
	opWriter, errWriter := InitUniRTMLogger(operationLogPath, errorLogPath)

	// Verify writers are not nil
	assert.NotNil(t, opWriter)
	assert.NotNil(t, errWriter)

	// Write a log to trigger file creation
	Logger.Info().Msg("test log")

	// Verify log files are created after writing
	_, err := os.Stat(operationLogPath)
	assert.NoError(t, err, "operation log file should be created after writing")

	// Write an error to trigger error log creation
	Logger.Error().Msg("test error")

	_, err = os.Stat(errorLogPath)
	assert.NoError(t, err, "error log file should be created after writing")
}

func TestInitUniRTMLoggerWithDefaults(t *testing.T) {
	// Save current directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	// Create temporary directory and change to it
	tmpDir := t.TempDir()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(originalDir)
	}()

	// Initialize logger with empty paths (should use defaults)
	opWriter, errWriter := InitUniRTMLogger("", "")

	// Verify writers are not nil
	assert.NotNil(t, opWriter)
	assert.NotNil(t, errWriter)

	// Write logs to trigger file creation
	Logger.Info().Msg("test log")
	Logger.Error().Msg("test error")

	// Verify default log files are created after writing
	_, err = os.Stat("unirtm.log")
	assert.NoError(t, err, "default operation log file should be created after writing")

	_, err = os.Stat("error.log")
	assert.NoError(t, err, "default error log file should be created after writing")
}

func TestWithContext(t *testing.T) {
	// Create logger with context
	fields := map[string]interface{}{
		"user_id":   123,
		"operation": "install",
		"tool":      "node",
		"version":   "20.0.0",
	}

	logger := WithContext(fields)
	assert.NotNil(t, logger)
}

func TestWithError(t *testing.T) {
	// Create logger with error
	testErr := assert.AnError
	logger := WithError(testErr)
	assert.NotNil(t, logger)
}

func TestWithFields(t *testing.T) {
	// Test with valid key-value pairs
	logger := WithFields(
		"user_id", 123,
		"operation", "install",
		"tool", "node",
	)
	assert.NotNil(t, logger)

	// Test with odd number of arguments (should log error and return Logger)
	logger = WithFields("key1", "value1", "key2")
	assert.NotNil(t, logger)
}

func TestLogLevels(t *testing.T) {
	// Create temporary directory for test logs
	tmpDir := t.TempDir()
	operationLogPath := filepath.Join(tmpDir, "unirtm.log")
	errorLogPath := filepath.Join(tmpDir, "error.log")

	// Initialize logger
	InitUniRTMLogger(operationLogPath, errorLogPath)

	// Set log level to trace to capture all logs
	zerolog.SetGlobalLevel(zerolog.TraceLevel)

	// Test all log levels
	Trace("trace message")
	Debug("debug message")
	Info("info message")
	Warn("warn message")
	Error("error message")

	// Test with context fields
	fields := map[string]interface{}{
		"test": "value",
	}
	Trace("trace with context", fields)
	Debug("debug with context", fields)
	Info("info with context", fields)
	Warn("warn with context", fields)
	Error("error with context", fields)

	// Verify log files exist and have content
	opInfo, err := os.Stat(operationLogPath)
	assert.NoError(t, err)
	assert.Greater(t, opInfo.Size(), int64(0), "operation log should have content")

	errInfo, err := os.Stat(errorLogPath)
	assert.NoError(t, err)
	assert.Greater(t, errInfo.Size(), int64(0), "error log should have content")
}

func TestErrorWithErr(t *testing.T) {
	// Create temporary directory for test logs
	tmpDir := t.TempDir()
	operationLogPath := filepath.Join(tmpDir, "unirtm.log")
	errorLogPath := filepath.Join(tmpDir, "error.log")

	// Initialize logger
	InitUniRTMLogger(operationLogPath, errorLogPath)

	// Test error logging with error object
	testErr := assert.AnError
	ErrorWithErr(testErr, "operation failed")

	// Test with context fields
	fields := map[string]interface{}{
		"operation": "install",
		"tool":      "node",
	}
	ErrorWithErr(testErr, "installation failed", fields)

	// Verify error log has content
	errInfo, err := os.Stat(errorLogPath)
	assert.NoError(t, err)
	assert.Greater(t, errInfo.Size(), int64(0), "error log should have content")
}

func TestUniRTMErrorHook(t *testing.T) {
	// Create temporary directory for test logs
	tmpDir := t.TempDir()
	operationLogPath := filepath.Join(tmpDir, "unirtm.log")
	errorLogPath := filepath.Join(tmpDir, "error.log")

	// Initialize logger
	InitUniRTMLogger(operationLogPath, errorLogPath)

	// Log an error to trigger the hook
	Logger.Error().Msg("test error with stack trace")

	// Verify error log has content (stack trace should be captured)
	errInfo, err := os.Stat(errorLogPath)
	assert.NoError(t, err)
	assert.Greater(t, errInfo.Size(), int64(0), "error log should have stack trace")
}

func TestColorize(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		color    int
		disabled bool
		wantLen  int // expected length > input length when color codes are added
	}{
		{
			name:     "with color",
			input:    "test",
			color:    colorRed,
			disabled: false,
			wantLen:  13, // "test" + ANSI codes (\x1b[31m + \x1b[0m = 9 chars)
		},
		{
			name:     "disabled color",
			input:    "test",
			color:    colorRed,
			disabled: true,
			wantLen:  4, // just "test"
		},
		{
			name:     "zero color code",
			input:    "test",
			color:    0,
			disabled: false,
			wantLen:  4, // just "test"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := colorize(tt.input, tt.color, tt.disabled)
			assert.Equal(t, tt.wantLen, len(result))
		})
	}
}

func TestShould(t *testing.T) {
	// Save original level
	originalLevel := zerolog.GlobalLevel()
	defer zerolog.SetGlobalLevel(originalLevel)

	tests := []struct {
		name        string
		globalLevel zerolog.Level
		testLevel   zerolog.Level
		want        bool
	}{
		{
			name:        "disabled global level",
			globalLevel: zerolog.Disabled,
			testLevel:   zerolog.ErrorLevel,
			want:        false,
		},
		{
			name:        "level below global",
			globalLevel: zerolog.ErrorLevel,
			testLevel:   zerolog.DebugLevel,
			want:        false,
		},
		{
			name:        "level at global",
			globalLevel: zerolog.InfoLevel,
			testLevel:   zerolog.InfoLevel,
			want:        true,
		},
		{
			name:        "level above global",
			globalLevel: zerolog.InfoLevel,
			testLevel:   zerolog.ErrorLevel,
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zerolog.SetGlobalLevel(tt.globalLevel)
			got := should(tt.testLevel)
			assert.Equal(t, tt.want, got)
		})
	}
}
