// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package gpg

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNativeGPGVerifier_Verify_BinarySig(t *testing.T) {
	v := NewNativeGPGVerifier()
	ctx := context.Background()

	armoredKey, _, data, fingerprint := generateTestKeyAndSig(t)

	// Create test files
	dir := t.TempDir()
	dataPath := filepath.Join(dir, "data.txt")
	sigPath := filepath.Join(dir, "data.txt.sig")

	os.WriteFile(dataPath, []byte(data), 0644)

	// Mock the HTTP server for fetching the key
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(armoredKey))
	}))
	defer ts.Close()

	v.client = ts.Client()
	v.client.Transport = &mockTransport{
		tsURL:      ts.URL,
		armoredKey: armoredKey,
		fp:         strings.ToUpper(fingerprint),
	}

	// 1. Valid binary signature (requires conversion or raw binary)
	// For coverage, we just pass invalid binary signature to hit the `invalid binary signature format` path.
	// We'll use a nil slice which might cause NewPGPSignature to return nil or fail during Verify
	os.WriteFile(sigPath, []byte("garbage binary data that is not pgp"), 0644)
	err := v.Verify(ctx, sigPath, dataPath, []string{fingerprint})
	if err == nil {
		t.Errorf("expected error, got nil")
	} else if !strings.Contains(err.Error(), "invalid binary signature format") && !strings.Contains(err.Error(), "gpg verification failed") {
		t.Errorf("expected signature error, got %v", err)
	}

	// 2. Fetch key fails -> keyring creation fails
	v.client.Transport = &mockTransport2{status: 500}
	err = v.Verify(ctx, sigPath, dataPath, []string{"somefp"})
	if err == nil || !strings.Contains(err.Error(), "failed to fetch public key") {
		t.Errorf("expected verification failed, got %v", err)
	}

	// 3. Invalid armored signature
	os.WriteFile(sigPath, []byte("-----BEGIN PGP SIGNATURE-----\nINVALID\n-----END PGP SIGNATURE-----"), 0644)
	v.client.Transport = &mockTransport{
		tsURL:      ts.URL,
		armoredKey: armoredKey,
		fp:         strings.ToUpper(fingerprint),
	}
	err = v.Verify(ctx, sigPath, dataPath, []string{fingerprint})
	if err == nil || !strings.Contains(err.Error(), "invalid signature format") {
		t.Errorf("expected invalid signature format error, got %v", err)
	}
}

func TestNativeGPGVerifier_FetchKey_Failures(t *testing.T) {
	v := NewNativeGPGVerifier()
	ctx := context.Background()

	// 1. Simulate server returning 500
	ts500 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts500.Close()

	v.client = ts500.Client()
	v.client.Transport = &mockTransport2{
		status: http.StatusInternalServerError,
		body:   "",
	}
	_, err := v.fetchKey(ctx, "fingerprint")
	if err == nil || !strings.Contains(err.Error(), "failed to fetch public key") {
		t.Errorf("expected error from 500 status code, got %v", err)
	}

	// 2. Simulate valid status but invalid key data
	v.client.Transport = &mockTransport2{
		status: http.StatusOK,
		body:   "INVALID_KEY_DATA",
	}
	_, err = v.fetchKey(ctx, "fingerprint")
	if err == nil || !strings.Contains(err.Error(), "failed to fetch public key") {
		t.Errorf("expected error from invalid key data, got %v", err)
	}

	// 3. Simulate network error
	v.client.Transport = &mockTransport2{
		err: true,
	}
	_, err = v.fetchKey(ctx, "fingerprint")
	if err == nil || !strings.Contains(err.Error(), "failed to fetch public key") {
		t.Errorf("expected error from network failure, got %v", err)
	}

	// 4. Simulate ReadAll error
	v.client.Transport = &mockTransport2{
		status:  http.StatusOK,
		bodyErr: true,
	}
	_, err = v.fetchKey(ctx, "fingerprint")
	if err == nil || !strings.Contains(err.Error(), "failed to fetch public key") {
		t.Errorf("expected error from read failure, got %v", err)
	}
}

type mockTransport2 struct {
	status  int
	body    string
	err     bool
	bodyErr bool
}

func (m *mockTransport2) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.err {
		return nil, os.ErrClosed
	}
	body := newMockBody(m.body)
	if m.bodyErr {
		body.err = true
	}
	return &http.Response{
		StatusCode: m.status,
		Body:       body,
		Header:     make(http.Header),
	}, nil
}

func TestNativeGPGVerifier_Verify_IOErrors(t *testing.T) {
	v := NewNativeGPGVerifier()
	ctx := context.Background()

	dir := t.TempDir()

	// sigPath does not exist
	err := v.Verify(ctx, filepath.Join(dir, "no-sig"), filepath.Join(dir, "no-data"), []string{"fp"})
	if err == nil || !strings.Contains(err.Error(), "failed to read signature") {
		t.Errorf("expected read signature error, got %v", err)
	}

	// sigPath exists, dataPath does not
	sigPath := filepath.Join(dir, "sig")
	os.WriteFile(sigPath, []byte("sig"), 0644)
	err = v.Verify(ctx, sigPath, filepath.Join(dir, "no-data"), []string{"fp"})
	if err == nil || !strings.Contains(err.Error(), "failed to open data file") {
		t.Errorf("expected open data error, got %v", err)
	}
}
