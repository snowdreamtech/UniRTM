// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package service

import (
	"context"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/repository"
)

func TestSecurityManager_VerifyChecksum_Success(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.bin")
	content := []byte("hello world")
	os.WriteFile(filePath, content, 0644)

	// Compute expected SHA256
	h := sha256.New()
	h.Write(content)
	expectedHash := hex.EncodeToString(h.Sum(nil))

	auditRepo := &recoveryMockAuditRepo{}
	sm := NewSecurityManager(auditRepo)

	res, err := sm.VerifyChecksum(context.Background(), "test-tool", "1.0.0", filePath, expectedHash)
	if err != nil {
		t.Fatalf("VerifyChecksum failed: %v", err)
	}

	if !res.Passed {
		t.Error("expected checksum to pass")
	}
	if res.Actual != expectedHash {
		t.Errorf("expected %q, got %q", expectedHash, res.Actual)
	}
}

func TestSecurityManager_VerifyChecksum_Failure(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.bin")
	os.WriteFile(filePath, []byte("hello world"), 0644)

	sm := NewSecurityManager(nil)

	res, err := sm.VerifyChecksum(context.Background(), "test-tool", "1.0.0", filePath, "invalid-hash-that-is-long-enough-to-not-crash-i-guess-wait-it-doesnt-matter")
	if err != nil {
		t.Fatalf("VerifyChecksum failed: %v", err)
	}

	if res.Passed {
		t.Error("expected checksum to fail")
	}
}

func TestSecurityManager_VerifyChecksum_NoChecksum(t *testing.T) {
	sm := NewSecurityManager(nil)
	res, err := sm.VerifyChecksum(context.Background(), "test-tool", "1.0.0", "some-path", "")
	if err != nil {
		t.Fatalf("VerifyChecksum failed: %v", err)
	}
	if res.Warning == "" {
		t.Error("expected a warning when no checksum provided")
	}
}

func TestSecurityManager_VerifyChecksum_SHA512(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.bin")
	content := []byte("hello world")
	os.WriteFile(filePath, content, 0644)

	h := sha512.New()
	h.Write(content)
	expectedHash := hex.EncodeToString(h.Sum(nil))

	sm := NewSecurityManager(nil)

	// Test with prefix
	res, err := sm.VerifyChecksum(context.Background(), "test-tool", "1.0.0", filePath, "sha512:"+expectedHash)
	if err != nil {
		t.Fatalf("VerifyChecksum failed: %v", err)
	}
	if !res.Passed {
		t.Error("expected SHA512 checksum to pass with prefix")
	}

	// Test without prefix (length inference)
	res, err = sm.VerifyChecksum(context.Background(), "test-tool", "1.0.0", filePath, expectedHash)
	if err != nil {
		t.Fatalf("VerifyChecksum failed: %v", err)
	}
	if !res.Passed {
		t.Error("expected SHA512 checksum to pass with inferred length")
	}
}

func TestSecurityManager_VerifyInstallation(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.bin")
	content := []byte("hello world")
	os.WriteFile(filePath, content, 0644)

	h := sha256.New()
	h.Write(content)
	expectedHash := hex.EncodeToString(h.Sum(nil))

	sm := NewSecurityManager(nil)

	inst := &repository.Installation{
		Tool:     "test-tool",
		Version:  "1.0.0",
		Checksum: expectedHash,
	}

	res, err := sm.VerifyInstallation(context.Background(), inst, filePath)
	if err != nil {
		t.Fatalf("VerifyInstallation failed: %v", err)
	}
	if !res.Passed {
		t.Error("expected installation verification to pass")
	}

	// Test without checksum
	inst.Checksum = ""
	res, err = sm.VerifyInstallation(context.Background(), inst, filePath)
	if err != nil {
		t.Fatalf("VerifyInstallation failed: %v", err)
	}
	if res.Warning == "" {
		t.Error("expected warning for empty checksum")
	}
}

func TestSecurityManager_FileError(t *testing.T) {
	sm := NewSecurityManager(nil)
	_, err := sm.VerifyChecksum(context.Background(), "test-tool", "1.0.0", "/non/existent/file", "somehash")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}
