package logger

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/rs/zerolog"
)

func TestLogger_Init(t *testing.T) {
	dir := t.TempDir()
	errLog := filepath.Join(dir, "err.log")
	ginLog := filepath.Join(dir, "gin.log")

	// Test InitLogger
	InitLogger(errLog, ginLog)

	// Test InitUniRTMLogger
	InitUniRTMLogger(filepath.Join(dir, "unirtm.log"), filepath.Join(dir, "error.log"))

	// Test zero logger
	initZeroLogger(os.Stdout)

	// Test Error logger
	initErrorLogger(errLog)
}

func TestLogger_Levels(t *testing.T) {
	// Disable actual exits for Fatal/Panic using a mock approach if needed, or just don't call them.
	// We cover Logf and Log
	Logf(zerolog.InfoLevel, "test %s", "info")
	Log(zerolog.DebugLevel, "test debug")

	ctx := context.Background()
	WithContext(map[string]interface{}{"key": ctx}).Info().Msg("info with context")
	WithFields("key", "val").Debug().Msg("debug with fields")
	WithFields("key").Debug().Msg("debug with odd fields") // test odd fields error
	WithError(errors.New("test err")).Error().Msg("error with err")

	// Methods from Trace to Error
	Trace("trace", map[string]interface{}{"a": 1})
	Trace("trace2")
	Debug("debug", map[string]interface{}{"a": 1})
	Debug("debug2")
	Info("info", map[string]interface{}{"a": 1})
	Warn("warn", map[string]interface{}{"a": 1})
	Warn("warn2")
	Error("error", map[string]interface{}{"a": 1})
	ErrorWithErr(errors.New("err"), "error with err", map[string]interface{}{"a": 1})

	// Test UniRTMErrorHook manually
	hook := UniRTMErrorHook{}
	event := Logger.Info()
	hook.Run(event, zerolog.ErrorLevel, "hook error msg")
	hook.Run(event, zerolog.InfoLevel, "hook info msg")

	// Test ErrorHook manually
	hook2 := ErrorHook{}
	hook2.Run(event, zerolog.ErrorLevel, "hook2 error msg")
}
