// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package download

import (
	"context"
	"testing"
)

// mockDownloader implements Downloader for testing.
type mockDownloader struct{}

func (m *mockDownloader) Download(ctx context.Context, url, destination string, options DownloadOptions) error {
	return nil
}

func (m *mockDownloader) VerifyChecksum(ctx context.Context, file string, expectedChecksum string) error {
	return nil
}

func TestManager(t *testing.T) {
	m := NewManager()

	// Initial state
	if !m.HasScheme("http") {
		t.Error("expected http scheme to be registered")
	}
	if !m.HasScheme("https") {
		t.Error("expected https scheme to be registered")
	}

	// List schemes
	schemes := m.ListSchemes()
	if len(schemes) != 2 {
		t.Errorf("expected 2 schemes, got %d", len(schemes))
	}

	// Get existing
	d, err := m.Get("http")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if d == nil {
		t.Error("expected downloader to be non-nil")
	}

	// Get non-existing
	_, err = m.Get("ftp")
	if err == nil {
		t.Error("expected error for non-existing scheme")
	}

	// Register custom
	m.Register("ftp", &mockDownloader{})
	if !m.HasScheme("ftp") {
		t.Error("expected ftp scheme to be registered")
	}

	// Unregister custom
	m.Unregister("ftp")
	if m.HasScheme("ftp") {
		t.Error("expected ftp scheme to be unregistered")
	}
}

func TestDefaultManager(t *testing.T) {
	// Test global functions
	Register("test", &mockDownloader{})
	if !HasScheme("test") {
		t.Error("expected test scheme to be registered in default manager")
	}

	d, err := Get("test")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if d == nil {
		t.Error("expected downloader to be non-nil")
	}

	schemes := ListSchemes()
	found := false
	for _, s := range schemes {
		if s == "test" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected test scheme in ListSchemes")
	}

	Unregister("test")
	if HasScheme("test") {
		t.Error("expected test scheme to be unregistered in default manager")
	}
}
