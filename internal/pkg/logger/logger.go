// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

// ANSI color codes
const (
	colorBlack = iota + 30
	colorRed
	colorGreen
	colorYellow
	colorBlue
	colorMagenta
	colorCyan
	colorWhite

	colorBold     = 1
	colorDarkGray = 90

	unknownLevel = "???"
)

var (
	// Logger is the global logger.
	Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()

	// NoColor disables the colorized output.
	NoColor bool

	// operationLogWriter is the writer for operation logs (unirtm.log)
	operationLogWriter io.Writer

	// errorLogWriter is the writer for error logs (error.log)
	errorLogWriter io.Writer
)

// InitLogger initializes the logger with the specified error and gin log paths.
// It returns the error and gin writers.
//
// Parameters:
//   - errorLogPath: The path to the error log file.
//   - ginLogPath: The path to the gin log file.
//
// Returns:
//   - errorWriter: The error writer.
//   - ginWriter: The gin writer.
func InitLogger(errorLogPath, ginLogPath string) (errorWriter io.Writer, ginWriter io.Writer) {

	// Set error.log
	errorWriter = initErrorLogger(errorLogPath)

	// Set gin.log
	ginWriter = initGinLogger(ginLogPath)

	// Set zerologger as the default logger
	initZeroLogger(ginWriter)

	return errorWriter, ginWriter
}

// InitUniRTMLogger initializes the logger for UniRTM with rotating file writers
// for both operation logs (unirtm.log) and error logs (error.log).
//
// This function configures:
//   - Rotating file writer for unirtm.log (operations, max 500MB, 10 backups, 30 days retention)
//   - Rotating file writer for error.log (errors only, max 500MB, 10 backups, 30 days retention)
//   - Console output with color-coded log levels
//   - Structured logging with context fields
//   - Stack trace capture for errors
//
// Parameters:
//   - operationLogPath: The path to the operation log file (default: "unirtm.log")
//   - errorLogPath: The path to the error log file (default: "error.log")
//
// Returns:
//   - operationWriter: The operation log writer
//   - errorWriter: The error log writer
func InitUniRTMLogger(operationLogPath, errorLogPath string) (operationWriter io.Writer, errorWriter io.Writer) {
	// Set default paths if not provided
	if operationLogPath == "" {
		operationLogPath = "unirtm.log"
	}
	if errorLogPath == "" {
		errorLogPath = "error.log"
	}

	// Initialize operation log with rotating file writer
	operationLog := &lumberjack.Logger{
		Filename:   operationLogPath,
		MaxSize:    500, // megabytes
		MaxBackups: 10,
		MaxAge:     30,   // days
		Compress:   true, // compress rotated files
	}

	// Initialize error log with rotating file writer
	errorLog := &lumberjack.Logger{
		Filename:   errorLogPath,
		MaxSize:    500, // megabytes
		MaxBackups: 10,
		MaxAge:     30,   // days
		Compress:   true, // compress rotated files
	}

	// Create multi-writers for operation logs (file + console)
	operationWriter = io.MultiWriter(operationLog, os.Stdout)
	operationLogWriter = operationWriter

	// Create multi-writers for error logs (file + stderr)
	errorWriter = io.MultiWriter(errorLog, os.Stderr)
	errorLogWriter = errorWriter

	// Configure zerolog with console writer for colored output
	consoleWriter := zerolog.ConsoleWriter{
		Out:        operationWriter,
		NoColor:    NoColor,
		TimeFormat: time.RFC3339,
	}

	// Create the main logger with console writer and error hook
	Logger = zerolog.New(consoleWriter).
		With().
		Timestamp().
		Caller(). // Add caller information for better debugging
		Logger().
		Hook(UniRTMErrorHook{})

	// Set as default logger
	zlog.Logger = Logger

	// Configure standard log to use zerolog
	log.SetOutput(Logger)
	log.SetFlags(0) // zerolog handles timestamps and formatting

	return operationWriter, errorWriter
}

// initErrorLogger initializes the error logger with the specified error log path.
// It returns the error writer.
//
// Parameters:
//   - filepath: The path to the error log file.
//
// Returns:
//   - errorWriter: The error writer.
func initErrorLogger(filepath string) io.Writer {
	if filepath == "" {
		filepath = "error.log"
	}

	// Set error.log
	errorlog := &lumberjack.Logger{
		Filename:   filepath,
		MaxSize:    500, // megabytes
		MaxBackups: 10,
		MaxAge:     30,   //days
		Compress:   true, // disabled by default
	}

	errorWriter := io.MultiWriter(errorlog, os.Stderr)

	// gin.DefaultErrorWriter = io.MultiWriter(errorlog)
	// Use the following code if you need to write the logs to file and console at the same time.
	gin.DefaultErrorWriter = errorWriter

	return errorWriter
}

// InitGinLogger initializes the gin logger with the specified gin log path.
// It returns the gin writer.
//
// Parameters:
//   - filepath: The path to the gin log file.
//
// Returns:
//   - ginWriter: The gin writer.
func initGinLogger(filepath string) io.Writer {
	if filepath == "" {
		filepath = "gin.log"
	}

	// Set gin.log
	ginlog := &lumberjack.Logger{
		Filename:   filepath,
		MaxSize:    500, // megabytes
		MaxBackups: 10,
		MaxAge:     30,   //days
		Compress:   true, // disabled by default
	}

	ginWriter := io.MultiWriter(ginlog, os.Stdout)

	// gin.DefaultWriter = io.MultiWriter(accesslog)
	// Use the following code if you need to write the logs to file and console at the same time.
	gin.DefaultWriter = ginWriter

	return ginWriter
}

// ErrorHook is a hook for logging errors.
type ErrorHook struct{}

// Run executes the ErrorHook, logging the event at the specified level with the given message.
// It writes the log output to the default error writer for the gin framework.
//
// Parameters:
//   - e: The zerolog event to be logged.
//   - level: The logging level of the event.
//   - msg: The message to be logged.
func (h ErrorHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	// Write the log output to the default error writer for the gin framework.
	if zerolog.GlobalLevel() != zerolog.Disabled && level >= zerolog.ErrorLevel && level <= zerolog.PanicLevel {
		t := colorize(time.Now().Format(time.RFC3339), colorDarkGray, NoColor)
		l := colorize(zerolog.FormattedLevels[level], zerolog.LevelColors[level], NoColor)
		s := colorize(fmt.Sprintf("[GIN-debug] [ERROR] %s\n", msg), colorBold, NoColor)

		fmt.Fprintf(gin.DefaultErrorWriter, "%s %s %s\n", t, l, s)
	}
}

// UniRTMErrorHook is a hook for logging errors in UniRTM.
// It captures stack traces for errors and writes them to the error log file.
type UniRTMErrorHook struct{}

// Run executes the UniRTMErrorHook, capturing stack traces for error-level logs
// and writing them to the error log file.
//
// Parameters:
//   - e: The zerolog event to be logged.
//   - level: The logging level of the event.
//   - msg: The message to be logged.
func (h UniRTMErrorHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	// Only process error-level and above logs
	if level < zerolog.ErrorLevel {
		return
	}

	// Capture stack trace for errors
	stackTrace := string(debug.Stack())

	// Add stack trace to the event
	e.Str("stack_trace", stackTrace)

	// Write error logs to the error log file if configured
	if errorLogWriter != nil && level >= zerolog.ErrorLevel && level <= zerolog.PanicLevel {
		// Create a separate logger for error file output
		errorLogger := zerolog.New(errorLogWriter).
			With().
			Timestamp().
			Caller().
			Logger()

		// Log to error file with stack trace
		errorLogger.WithLevel(level).
			Str("stack_trace", stackTrace).
			Msg(msg)
	}
}

// initZeroLogger initializes the zerolog logger with the specified output writer.
// It sets the time field format to UNIX timestamp and the logger as the default logger.
//
// Parameters:
//   - out: The output writer for the logger.
func initZeroLogger(out io.Writer) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix // UNIX timestamp
	zerologger := zerolog.New(zerolog.ConsoleWriter{Out: out, NoColor: NoColor, TimeFormat: time.RFC3339}).With().Timestamp().Logger().Hook(ErrorHook{})

	// Set zerologger as the default logger
	log.SetOutput(zerologger)
	log.SetPrefix("[GIN-debug] ")
	log.SetFlags(log.LUTC)

	zlog.Logger = zerologger
	Logger = zerologger
}

// colorize returns the string s wrapped in ANSI code c, unless disabled is true or c is 0.
func colorize(s interface{}, c int, disabled bool) string {
	e := os.Getenv("NO_COLOR")
	if e != "" || c == 0 {
		disabled = true
	}

	if disabled {
		return fmt.Sprintf("%s", s)
	}
	return fmt.Sprintf("\x1b[%dm%v\x1b[0m", c, s)
}

// WithContext returns a logger with the specified context fields.
// This enables structured logging with key-value pairs.
//
// Example:
//
//	logger.WithContext(map[string]interface{}{
//	    "user_id": 123,
//	    "operation": "install",
//	    "tool": "node",
//	    "version": "20.0.0",
//	}).Info("Installing tool")
func WithContext(fields map[string]interface{}) *zerolog.Logger {
	ctx := Logger.With()
	for key, value := range fields {
		ctx = ctx.Interface(key, value)
	}
	logger := ctx.Logger()
	return &logger
}

// WithError returns a logger with the error field set.
// This is a convenience function for logging errors with context.
//
// Example:
//
//	logger.WithError(err).Error("Failed to install tool")
func WithError(err error) *zerolog.Logger {
	logger := Logger.With().Err(err).Logger()
	return &logger
}

// WithFields returns a logger with multiple fields set.
// This is a convenience function for structured logging.
//
// Example:
//
//	logger.WithFields(
//	    "user_id", 123,
//	    "operation", "install",
//	).Info("Starting operation")
func WithFields(keysAndValues ...interface{}) *zerolog.Logger {
	if len(keysAndValues)%2 != 0 {
		Logger.Error().Msg("WithFields called with odd number of arguments")
		return &Logger
	}

	ctx := Logger.With()
	for i := 0; i < len(keysAndValues); i += 2 {
		key, ok := keysAndValues[i].(string)
		if !ok {
			Logger.Error().Msgf("WithFields key at index %d is not a string", i)
			continue
		}
		ctx = ctx.Interface(key, keysAndValues[i+1])
	}
	logger := ctx.Logger()
	return &logger
}

// Logf logs a formatted message at the specified log level.
// The log message includes a timestamp, the log level, and the formatted message.
//
// Parameters:
//   - level: The log level (e.g., "INF", "ERR").
//
// TraceLevel: "TRC",
// DebugLevel: "DBG",
// InfoLevel:  "INF",
// WarnLevel:  "WRN",
// ErrorLevel: "ERR",
// FatalLevel: "FTL",
// PanicLevel: "PNC",
//
//   - format: The format string for the log message.
//   - v: The values to be formatted according to the format string.
//
// Deprecated: Use Zerolog instead.
func Logf(level zerolog.Level, format string, v ...interface{}) {
	if should(level) {
		t := colorize(time.Now().Format(time.RFC3339), colorDarkGray, NoColor)
		l := colorize(zerolog.FormattedLevels[level], zerolog.LevelColors[level], NoColor)
		s := colorize(fmt.Sprintf(format, v...), colorBold, NoColor)

		fmt.Fprintf(gin.DefaultWriter, "%s %s %s\n", t, l, s)
	}
}

// Log logs a message at the specified log level.
// The log message is formatted with the current time, log level, and the provided arguments.
// The log output is written to the default writer of the gin framework.
//
// Parameters:
//   - level: The log level as a string (e.g., "INF", "ERR").
//
// TraceLevel: "TRC",
// DebugLevel: "DBG",
// InfoLevel:  "INF",
// WarnLevel:  "WRN",
// ErrorLevel: "ERR",
// FatalLevel: "FTL",
// PanicLevel: "PNC",
//
//   - v: Variadic arguments to be included in the log message.
//
// Deprecated: Use Zerolog instead.
func Log(level zerolog.Level, v ...interface{}) {
	if should(level) {
		t := colorize(time.Now().Format(time.RFC3339), colorDarkGray, NoColor)
		l := colorize(zerolog.FormattedLevels[level], zerolog.LevelColors[level], NoColor)
		s := colorize(fmt.Sprint(v...), colorBold, NoColor)

		fmt.Fprintf(gin.DefaultWriter, "%s %s %s\n", t, l, s)
	}
}

// should returns true if the log should be logged.
func should(level zerolog.Level) bool {
	if zerolog.GlobalLevel() == zerolog.Disabled {
		return false
	}

	if level < zerolog.GlobalLevel() {
		return false
	}

	return true
}

// Trace logs a message at trace level with optional context fields.
func Trace(msg string, fields ...map[string]interface{}) {
	if len(fields) > 0 {
		logger := WithContext(fields[0])
		logger.Trace().Msg(msg)
	} else {
		Logger.Trace().Msg(msg)
	}
}

// Debug logs a message at debug level with optional context fields.
func Debug(msg string, fields ...map[string]interface{}) {
	if len(fields) > 0 {
		logger := WithContext(fields[0])
		logger.Debug().Msg(msg)
	} else {
		Logger.Debug().Msg(msg)
	}
}

// Info logs a message at info level with optional context fields.
func Info(msg string, fields ...map[string]interface{}) {
	if len(fields) > 0 {
		logger := WithContext(fields[0])
		logger.Info().Msg(msg)
	} else {
		Logger.Info().Msg(msg)
	}
}

// Warn logs a message at warn level with optional context fields.
func Warn(msg string, fields ...map[string]interface{}) {
	if len(fields) > 0 {
		logger := WithContext(fields[0])
		logger.Warn().Msg(msg)
	} else {
		Logger.Warn().Msg(msg)
	}
}

// Error logs a message at error level with optional context fields.
// Stack traces are automatically captured for error-level logs.
func Error(msg string, fields ...map[string]interface{}) {
	if len(fields) > 0 {
		logger := WithContext(fields[0])
		logger.Error().Msg(msg)
	} else {
		Logger.Error().Msg(msg)
	}
}

// Fatal logs a message at fatal level with optional context fields and exits.
// Stack traces are automatically captured for fatal-level logs.
func Fatal(msg string, fields ...map[string]interface{}) {
	if len(fields) > 0 {
		logger := WithContext(fields[0])
		logger.Fatal().Msg(msg)
	} else {
		Logger.Fatal().Msg(msg)
	}
}

// Panic logs a message at panic level with optional context fields and panics.
// Stack traces are automatically captured for panic-level logs.
func Panic(msg string, fields ...map[string]interface{}) {
	if len(fields) > 0 {
		logger := WithContext(fields[0])
		logger.Panic().Msg(msg)
	} else {
		Logger.Panic().Msg(msg)
	}
}

// ErrorWithErr logs an error with its message and optional context fields.
// This is a convenience function for logging errors with automatic error field extraction.
func ErrorWithErr(err error, msg string, fields ...map[string]interface{}) {
	if len(fields) > 0 {
		ctx := Logger.With().Err(err)
		for key, value := range fields[0] {
			ctx = ctx.Interface(key, value)
		}
		logger := ctx.Logger()
		logger.Error().Msg(msg)
	} else {
		Logger.Error().Err(err).Msg(msg)
	}
}

// FatalWithErr logs an error with its message and optional context fields, then exits.
func FatalWithErr(err error, msg string, fields ...map[string]interface{}) {
	if len(fields) > 0 {
		ctx := Logger.With().Err(err)
		for key, value := range fields[0] {
			ctx = ctx.Interface(key, value)
		}
		logger := ctx.Logger()
		logger.Fatal().Msg(msg)
	} else {
		Logger.Fatal().Err(err).Msg(msg)
	}
}
