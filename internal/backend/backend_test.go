// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"errors"
	"strings"
	"testing"
)

func TestPlatform_String(t *testing.T) {
	p := Platform{OS: "linux", Arch: "amd64"}
	if p.String() != "linux-amd64" {
		t.Errorf("expected linux-amd64, got %s", p.String())
	}
}

func TestCurrentPlatform(t *testing.T) {
	p := CurrentPlatform()
	if p.OS == "" || p.Arch == "" {
		t.Error("CurrentPlatform should return non-empty OS and Arch")
	}
}

func TestBackendError(t *testing.T) {
	err1 := NewBackendError("test", "mytool", "something failed", nil)
	if !strings.Contains(err1.Error(), "test backend error for mytool: something failed") {
		t.Errorf("unexpected error string: %s", err1.Error())
	}
	if err1.Unwrap() != nil {
		t.Error("expected nil cause")
	}

	cause := errors.New("underlying")
	err2 := NewBackendError("test", "mytool", "something failed", cause)
	if !strings.Contains(err2.Error(), "test backend error for mytool: something failed: underlying") {
		t.Errorf("unexpected error string: %s", err2.Error())
	}
	if err2.Unwrap() != cause {
		t.Error("expected to unwrap to underlying cause")
	}
}
