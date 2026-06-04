// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package gpg

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
)

// generateTestKeyAndSig creates a test key, signs data, and returns key, sig, data, and fingerprint
func generateTestKeyAndSig(t *testing.T) (armoredKey, armoredSig, data string, fingerprint string) {
	// 1. Generate a key
	key, err := crypto.GenerateKey("test", "test@example.com", "rsa", 2048)
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	armoredKey, err = key.Armor()
	if err != nil {
		t.Fatalf("Failed to armor key: %v", err)
	}
	fingerprint = key.GetFingerprint()

	// 2. Sign some data
	data = "test data for signature verification"
	msg := crypto.NewPlainMessageFromString(data)

	keyRing, err := crypto.NewKeyRing(key)
	if err != nil {
		t.Fatalf("Failed to create keyring: %v", err)
	}

	sig, err := keyRing.SignDetached(msg)
	if err != nil {
		t.Fatalf("Failed to sign data: %v", err)
	}

	armoredSig, err = sig.GetArmored()
	if err != nil {
		t.Fatalf("Failed to armor signature: %v", err)
	}

	return armoredKey, armoredSig, data, fingerprint
}

func TestNewVerifier(t *testing.T) {
	v := NewVerifier()
	if _, ok := v.(*NativeGPGVerifier); !ok {
		t.Errorf("NewVerifier should return a NativeGPGVerifier by default")
	}
}

func TestNativeGPGVerifier_IsAvailable(t *testing.T) {
	v := NewNativeGPGVerifier()
	if !v.IsAvailable(context.Background()) {
		t.Errorf("NativeGPGVerifier should always be available")
	}
}

func TestNativeGPGVerifier_Verify(t *testing.T) {
	v := NewNativeGPGVerifier()
	ctx := context.Background()

	armoredKey, armoredSig, data, fingerprint := generateTestKeyAndSig(t)

	// Create test files
	dir := t.TempDir()
	dataPath := filepath.Join(dir, "data.txt")
	sigPath := filepath.Join(dir, "data.txt.sig")

	os.WriteFile(dataPath, []byte(data), 0644)
	os.WriteFile(sigPath, []byte(armoredSig), 0644)

	// Mock the HTTP server for fetching the key
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, strings.ToUpper(fingerprint)) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(armoredKey))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	// Inject the mock transport into the verifier's client
	v.client = ts.Client()
	// Override the keyservers to use our mock server for testing
	// We need to temporarily modify fetchKey logic by overriding the URLs in the test
	// But since keyservers is hardcoded, we will just use a custom RoundTripper

	v.client.Transport = &mockTransport{
		tsURL:      ts.URL,
		armoredKey: armoredKey,
		fp:         strings.ToUpper(fingerprint),
	}

	// 1. Valid verification
	err := v.Verify(ctx, sigPath, dataPath, []string{fingerprint})
	if err != nil {
		t.Errorf("expected verification to succeed, got: %v", err)
	}

	// 2. Missing sig file
	err = v.Verify(ctx, "non_existent.sig", dataPath, []string{fingerprint})
	if err == nil {
		t.Errorf("expected error for missing sig file")
	}

	// 3. Missing data file
	err = v.Verify(ctx, sigPath, "non_existent.txt", []string{fingerprint})
	if err == nil {
		t.Errorf("expected error for missing data file")
	}

	// 4. Invalid fingerprint (will fail to fetch key)
	err = v.Verify(ctx, sigPath, dataPath, []string{"invalid_fp"})
	if err == nil {
		t.Errorf("expected error for invalid fingerprint")
	}

	// 5. Corrupted signature
	badSigPath := filepath.Join(dir, "bad.sig")
	os.WriteFile(badSigPath, []byte("NOT A VALID SIG"), 0644)
	err = v.Verify(ctx, badSigPath, dataPath, []string{fingerprint})
	if err == nil {
		t.Errorf("expected error for corrupted signature")
	}
}

func TestNativeGPGVerifier_ImportKey(t *testing.T) {
	v := NewNativeGPGVerifier()
	ctx := context.Background()

	armoredKey, _, _, fingerprint := generateTestKeyAndSig(t)

	v.client.Transport = &mockTransport{
		armoredKey: armoredKey,
		fp:         strings.ToUpper(fingerprint),
	}

	err := v.ImportKey(ctx, fingerprint)
	if err != nil {
		t.Errorf("expected ImportKey to succeed, got: %v", err)
	}

	err = v.ImportKey(ctx, "invalid")
	if err == nil {
		t.Errorf("expected ImportKey to fail with invalid fingerprint")
	}
}

// mockTransport intercepts requests and returns our mocked key
type mockTransport struct {
	tsURL      string
	armoredKey string
	fp         string
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if strings.Contains(req.URL.String(), m.fp) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       newMockBody(m.armoredKey),
			Header:     make(http.Header),
		}, nil
	}
	return &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       newMockBody(""),
		Header:     make(http.Header),
	}, nil
}

type mockBody struct {
	data *strings.Reader
	err  bool
}

func newMockBody(s string) *mockBody {
	return &mockBody{data: strings.NewReader(s)}
}
func (m *mockBody) Read(p []byte) (n int, err error) {
	if m.err {
		return 0, os.ErrClosed
	}
	return m.data.Read(p)
}
func (m *mockBody) Close() error { return nil }

// --- SystemGPGVerifier Tests ---

func TestSystemGPGVerifier(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping SystemGPGVerifier test on Windows due to mock gpg.bat flakiness")
	}
	v := NewSystemGPGVerifier()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 1. Test when gpg is NOT in PATH
	// We do this by temporarily setting a broken PATH
	oldPath := os.Getenv("PATH")
	t.Setenv("PATH", "/invalid/path/for/test")
	defer os.Setenv("PATH", oldPath)

	if v.IsAvailable(ctx) {
		t.Errorf("IsAvailable should return false when gpg is missing")
	}

	err := v.Verify(ctx, "sig", "data", nil)
	if err == nil || !strings.Contains(err.Error(), "gpg command not found") {
		t.Errorf("expected 'gpg command not found', got %v", err)
	}
	os.Setenv("PATH", oldPath) // restore early

	// 2. Create a mock gpg executable
	mockGpgDir := t.TempDir()

	// Create the mock script
	var mockScript, ext string
	if runtime.GOOS == "windows" {
		ext = ".bat"
		mockScript = `@echo off
echo %* | findstr /C:"bad_key.sig" >nul
if %errorlevel% equ 0 (
    echo NO_PUBKEY 123456
    exit /b 2
)
echo %* | findstr /C:"bad_sig.sig" >nul
if %errorlevel% equ 0 (
    echo BADSIG
    exit /b 1
)
echo %* | findstr /C:"good_sig_no_match.sig" >nul
if %errorlevel% equ 0 (
    echo [GNUPG:] VALIDSIG ABCDEF
    echo GOODSIG
    exit /b 0
)
echo %* | findstr /C:"good_sig_match.sig" >nul
if %errorlevel% equ 0 (
    echo [GNUPG:] VALIDSIG MYFINGERPRINT
    echo GOODSIG
    exit /b 0
)
echo %* | findstr /C:"no_goodsig.sig" >nul
if %errorlevel% equ 0 (
    echo Something else
    exit /b 0
)
echo %* | findstr /C:"fail_key" >nul
if %errorlevel% equ 0 (
    exit /b 1
)
exit /b 0
`
	} else {
		mockScript = `#!/bin/sh
case "$*" in
    *bad_key.sig*)
        echo "NO_PUBKEY 123456"
        exit 2
        ;;
    *bad_sig.sig*)
        echo "BADSIG"
        exit 1
        ;;
    *good_sig_no_match.sig*)
        echo "[GNUPG:] VALIDSIG ABCDEF"
        echo "GOODSIG"
        exit 0
        ;;
    *good_sig_match.sig*)
        echo "[GNUPG:] VALIDSIG MYFINGERPRINT"
        echo "GOODSIG"
        exit 0
        ;;
    *no_goodsig.sig*)
        echo "Something else"
        exit 0
        ;;
    *fail_key*)
        exit 1
        ;;
esac
exit 0
`
	}
	mockGpgPath := filepath.Join(mockGpgDir, "gpg"+ext)
	os.WriteFile(mockGpgPath, []byte(mockScript), 0755)

	// Prepend mock to PATH
	t.Setenv("PATH", mockGpgDir+string(os.PathListSeparator)+oldPath)
	defer os.Setenv("PATH", oldPath)

	// 3. Test Verify NO_PUBKEY
	err = v.Verify(ctx, "bad_key.sig", "data.txt", nil)
	if err == nil || !strings.Contains(err.Error(), "missing public key") {
		t.Errorf("expected missing public key error, got %v", err)
	}

	// 4. Test Verify BADSIG
	err = v.Verify(ctx, "bad_sig.sig", "data.txt", nil)
	if err == nil || !strings.Contains(err.Error(), "verification failed") {
		t.Errorf("expected verification failed error, got %v", err)
	}

	// 5. Test Verify GOODSIG but unmatched fingerprint
	err = v.Verify(ctx, "good_sig_no_match.sig", "data.txt", []string{"MYFINGERPRINT"})
	if err == nil || !strings.Contains(err.Error(), "security violation") {
		t.Errorf("expected security violation error, got %v", err)
	}

	// 6. Test Verify GOODSIG and matched fingerprint
	err = v.Verify(ctx, "good_sig_match.sig", "data.txt", []string{"MYFINGERPRINT"})
	if err != nil {
		t.Errorf("expected success, got %v", err)
	}

	// 7. Test Verify with output missing GOODSIG
	err = v.Verify(ctx, "no_goodsig.sig", "data.txt", nil)
	if err == nil || !strings.Contains(err.Error(), "no valid signature found") {
		t.Errorf("expected no valid signature found error, got %v", err)
	}

	// 8. Test ImportKey Success
	err = v.ImportKey(ctx, "MYFINGERPRINT")
	if err != nil {
		t.Errorf("expected ImportKey success, got %v", err)
	}

	// 9. Test ImportKey Fail
	err = v.ImportKey(ctx, "fail_key")
	if err == nil {
		t.Errorf("expected ImportKey to fail")
	}
}

func TestNativeGPGVerifier_Verify_Errors(t *testing.T) {
	v := NewNativeGPGVerifier()
	ctx := context.Background()

	// 1. Missing signature file
	err := v.Verify(ctx, "nonexistent.sig", "data.txt", []string{"FP"})
	if err == nil || !strings.Contains(err.Error(), "failed to read signature file") {
		t.Errorf("expected failed to read signature file error, got %v", err)
	}

	// 2. Missing data file
	dir := t.TempDir()
	sigPath := filepath.Join(dir, "sig.sig")
	os.WriteFile(sigPath, []byte("some sig"), 0644)
	err = v.Verify(ctx, sigPath, "nonexistent.txt", []string{"FP"})
	if err == nil || !strings.Contains(err.Error(), "failed to open data file") {
		t.Errorf("expected failed to open data file error, got %v", err)
	}

	// 3. No fingerprints
	dataPath := filepath.Join(dir, "data.txt")
	os.WriteFile(dataPath, []byte("data"), 0644)
	err = v.Verify(ctx, sigPath, dataPath, nil)
	if err == nil || !strings.Contains(err.Error(), "no trusted keys matched") {
		t.Errorf("expected no trusted keys matched error, got %v", err)
	}

	// 4. Invalid signature format
	// Setup a mock transport that returns a valid key, but signature is bad
	armoredKey, _, _, fingerprint := generateTestKeyAndSig(t)
	v.client.Transport = &mockTransport{
		armoredKey: armoredKey,
		fp:         strings.ToUpper(fingerprint),
	}
	err = v.Verify(ctx, sigPath, dataPath, []string{fingerprint})
	if err == nil || (!strings.Contains(err.Error(), "invalid signature format") && !strings.Contains(err.Error(), "gpg verification failed")) {
		t.Errorf("expected signature error, got %v", err)
	}
}
