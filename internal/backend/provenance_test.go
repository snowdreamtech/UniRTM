package backend

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/sigstore/sigstore-go/pkg/verify"
)

func TestVerifyBundle_InvalidJSON(t *testing.T) {
	v := &SigstoreVerifier{}
	_, err := v.verifyBundle(json.RawMessage(`invalid`), "abcd", verify.CertificateIdentity{}, nil)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestVerifyBundle_InvalidDigest(t *testing.T) {
	v := &SigstoreVerifier{}
	bundleJSON := `{"mediaType":"application/vnd.dev.sigstore.bundle+json;version=0.3"}`
	_, err := v.verifyBundle(json.RawMessage(bundleJSON), "invalid-hex!", verify.CertificateIdentity{}, nil)
	if err == nil {
		t.Error("expected error for invalid digest hex")
	}
}

func TestVerifyBundle_VerifyFails(t *testing.T) {
	v := &SigstoreVerifier{}
	bundleJSON := `{"mediaType":"application/vnd.dev.sigstore.bundle+json;version=0.3"}`
	_, err := v.verifyBundle(json.RawMessage(bundleJSON), "deadbeef", verify.CertificateIdentity{}, nil)
	if err == nil {
		t.Error("expected error for verify failure")
	} else {
		t.Logf("ReadFile error: %v", err)
	}
}

func TestVerifyBundle(t *testing.T) {
	v := &SigstoreVerifier{}
	_, err := v.verifyBundle([]byte("invalid json"), "digest", verify.CertificateIdentity{}, nil)
	if err == nil {
		t.Errorf("expected error on invalid json, got nil")
	}

	// Invalid bundle but valid JSON
	_, err = v.verifyBundle([]byte(`{"mediaType":"application/vnd.dev.sigstore.bundle+json;version=0.1"}`), "digest", verify.CertificateIdentity{}, nil)
	if err == nil {
		t.Errorf("expected error on missing bundle content, got nil")
	}

	// Read real bundle
	path := "testdata/dummy_bundle.json"
	data, err := os.ReadFile(path)
	if err == nil {
		t.Setenv("SKIP_REKOR_VERIFY", "1")
		_, err = v.verifyBundle(data, "24e4d34078ae81da7c82539616f0ccac3e226cf4f74a38ce6fb3463619e50a55", verify.CertificateIdentity{}, nil)
		t.Logf("verifyBundle err (SKIP=1): %v", err)
		t.Setenv("SKIP_REKOR_VERIFY", "0")
		_, err = v.verifyBundle(data, "24e4d34078ae81da7c82539616f0ccac3e226cf4f74a38ce6fb3463619e50a55", verify.CertificateIdentity{}, nil)
		t.Logf("verifyBundle err (SKIP=0): %v", err)
	} else {
		t.Logf("Failed to read test data: %v", err)
	}
}

func TestDownloadFile(t *testing.T) {
	f := &tufFetcher{
		client: &http.Client{Timeout: 5 * time.Second},
	}
	// Test invalid URL
	_, err := f.DownloadFile("invalid url %%%", 100, 0)
	if err == nil {
		t.Errorf("expected error on invalid url")
	}

	// Test non-existent URL
	_, err = f.DownloadFile("http://localhost:12345/nonexistent", 100, 0)
	if err == nil {
		t.Errorf("expected error on nonexistent url")
	}
}

func TestDownloadFile_RelativePath(t *testing.T) {
	// Test that relative paths get the base URL prepended
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/somefile.json" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"ok":true}`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	f := &tufFetcher{
		client:            &http.Client{Timeout: 5 * time.Second},
		repositoryBaseURL: srv.URL,
	}

	data, err := f.DownloadFile("somefile.json", 1000, 0)
	if err != nil {
		t.Fatalf("expected success for relative path, got: %v", err)
	}
	if string(data) != `{"ok":true}` {
		t.Errorf("unexpected data: %s", string(data))
	}
}

func TestDownloadFile_Non200Status(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	f := &tufFetcher{
		client: &http.Client{Timeout: 5 * time.Second},
	}

	_, err := f.DownloadFile(srv.URL+"/file", 100, 0)
	if err == nil {
		t.Error("expected error for non-200 status")
	}
}

func TestDownloadFile_WithTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	f := &tufFetcher{
		client: &http.Client{Timeout: 5 * time.Second},
	}

	// Pass a non-zero timeout (context with timeout path)
	data, err := f.DownloadFile(srv.URL+"/file", 1000, 2*time.Second)
	if err != nil {
		t.Fatalf("expected success with timeout, got: %v", err)
	}
	if string(data) != `{"ok":true}` {
		t.Errorf("unexpected data: %s", string(data))
	}
}

func TestInitializeTUFRoot_WithCacheDir(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("UNIRTM_TUF_CACHE_DIR", dir)
	// Attempt to initialize - will fail at TUF fetch but should exercise the cache path logic
	_, err := InitializeTUFRoot("test-cache", "http://localhost:12345", nil)
	if err == nil {
		t.Log("InitializeTUFRoot succeeded unexpectedly")
	} else {
		t.Logf("InitializeTUFRoot expected error: %v", err)
	}
}

func TestVerifyBundle_InvalidDigestHex(t *testing.T) {
	// A valid empty bundle structure to bypass JSON parsing
	emptyBundle := `{"mediaType": "application/vnd.dev.sigstore.bundle+json;version=0.3"}`
	v := &SigstoreVerifier{}
	
	// Set SKIP_REKOR_VERIFY to bypass tlog checking which would otherwise error early or use different path
	t.Setenv("UNIRTM_SKIP_REKOR_VERIFY", "1")
	
	dummyIdentity, _ := verify.NewShortCertificateIdentity("https://dummy", "", "", ".*")
	_, err := v.verifyBundle(json.RawMessage(emptyBundle), "invalid-hex-%%%", dummyIdentity, nil)
	if err == nil {
		t.Error("expected error for invalid hex digest")
	}
}
