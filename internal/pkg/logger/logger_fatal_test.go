// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package logger

import (
	"errors"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	Panic("test panic message")
}

func TestPanicWithFields(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	Panic("test panic message", map[string]interface{}{"key": "value"})
}

func TestFatalSubprocess(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		Fatal("test fatal message")
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestFatalSubprocess")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func TestFatalWithFieldsSubprocess(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		Fatal("test fatal message", map[string]interface{}{"key": "value"})
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestFatalWithFieldsSubprocess")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func TestFatalWithErrSubprocess(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		FatalWithErr(errors.New("test error"), "test fatal message")
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestFatalWithErrSubprocess")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func TestFatalWithErrWithFieldsSubprocess(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		FatalWithErr(errors.New("test error"), "test fatal message", map[string]interface{}{"key": "value"})
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestFatalWithErrWithFieldsSubprocess")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}

func TestWithFieldsCoverage(t *testing.T) {
	// Testing the loop branch inside WithFields when len > 0
	l := WithFields(map[string]interface{}{"foo": "bar"})
	assert.NotNil(t, l)
}
