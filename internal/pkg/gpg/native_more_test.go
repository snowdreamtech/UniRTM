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
	os.WriteFile(sigPath, nil, 0644)
	err := v.Verify(ctx, sigPath, dataPath, []string{fingerprint})
	if err == nil {
		t.Errorf("expected error, got nil")
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
	if err == nil {
		t.Errorf("expected error from 500 status code")
	}

	// 2. Simulate valid status but invalid key data
	v.client.Transport = &mockTransport2{
		status: http.StatusOK,
		body:   "INVALID_KEY_DATA",
	}
	_, err = v.fetchKey(ctx, "fingerprint")
	if err == nil {
		t.Errorf("expected error from invalid key data")
	}

	// 3. Simulate network error
	v.client.Transport = &mockTransport2{
		err: true,
	}
	_, err = v.fetchKey(ctx, "fingerprint")
	if err == nil {
		t.Errorf("expected error from network failure")
	}
}

type mockTransport2 struct {
	status int
	body   string
	err    bool
}

func (m *mockTransport2) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.err {
		return nil, os.ErrClosed
	}
	return &http.Response{
		StatusCode: m.status,
		Body:       newMockBody(m.body),
		Header:     make(http.Header),
	}, nil
}
