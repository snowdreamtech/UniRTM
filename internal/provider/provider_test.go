// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"errors"
	"strings"
	"testing"
)

func TestProviderError(t *testing.T) {
	err1 := NewProviderError("test_prov", "test_tool", "1.0", "something failed", nil)
	expected1 := "test_prov provider error for test_tool 1.0: something failed"
	if err1.Error() != expected1 {
		t.Errorf("expected '%s', got '%s'", expected1, err1.Error())
	}
	if err1.Unwrap() != nil {
		t.Error("expected nil cause")
	}

	cause := errors.New("underlying cause")
	err2 := NewProviderError("test_prov", "test_tool", "1.0", "something failed", cause)
	expected2 := "test_prov provider error for test_tool 1.0: something failed: underlying cause"
	if err2.Error() != expected2 {
		t.Errorf("expected '%s', got '%s'", expected2, err2.Error())
	}
	if err2.Unwrap() != cause {
		t.Error("expected cause to match")
	}
}

func TestDomainFromURL(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"https://example.com/path", "example.com"},
		{"http://test.org:8080/foo", "test.org"},
		{"example.net/bar", "example.net"},
		{"", ""},
	}

	for _, tc := range tests {
		t.Run(tc.url, func(t *testing.T) {
			result := DomainFromURL(tc.url)
			if result != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, result)
			}
		})
	}
}

func TestGetNoProxyEnv(t *testing.T) {
	// Temporarily override GlobalNoProxy to avoid side effects
	originalGlobalNoProxy := GlobalNoProxy
	defer func() { GlobalNoProxy = originalGlobalNoProxy }()
	GlobalNoProxy = []string{"global.example.com"}

	env := GetNoProxyEnv("extra.example.com")

	found := false
	for _, e := range env {
		if strings.HasPrefix(strings.ToUpper(e), "NO_PROXY=") {
			found = true
			if !strings.Contains(e, "global.example.com") {
				t.Errorf("expected NO_PROXY to contain global.example.com, got %s", e)
			}
			if !strings.Contains(e, "extra.example.com") {
				t.Errorf("expected NO_PROXY to contain extra.example.com, got %s", e)
			}
			break
		}
	}

	if !found {
		t.Error("expected to find NO_PROXY in environment")
	}
}
