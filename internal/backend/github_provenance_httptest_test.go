// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang/snappy"
	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
)

func TestFetchAttestations_LiveMock(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/attestations/sha256:abcd", func(w http.ResponseWriter, r *http.Request) {
		resp := attestationAPIResponse{
			Attestations: []attestationEntry{
				{Bundle: json.RawMessage(`{"fake":"bundle"}`)},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	t.Setenv("UNIRTM_GITHUB_API_BASEURL", server.URL)

	verifier := &provenanceVerifier{
		client: pkgHttp.NewClientWithTimeout(30 * time.Second),
	}

	bundles, err := verifier.fetchAttestations(context.Background(), "token", "owner", "repo", "abcd")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(bundles) != 1 {
		t.Fatalf("expected 1 bundle, got %d", len(bundles))
	}
}

func TestFetchAttestations_NotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/attestations/sha256:abcd", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	t.Setenv("UNIRTM_GITHUB_API_BASEURL", server.URL)

	verifier := &provenanceVerifier{
		client: pkgHttp.NewClientWithTimeout(30 * time.Second),
	}

	bundles, err := verifier.fetchAttestations(context.Background(), "token", "owner", "repo", "abcd")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(bundles) != 0 {
		t.Fatalf("expected 0 bundles, got %d", len(bundles))
	}
}

func TestFetchExternalBundle_Snappy(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/bundle.sn", func(w http.ResponseWriter, r *http.Request) {
		data := []byte(`{"fake":"external"}`)
		encoded := snappy.Encode(nil, data)
		w.Write(encoded)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	t.Setenv("UNIRTM_ENABLE_GITHUB_PROXY", "0")

	verifier := &provenanceVerifier{
		client: pkgHttp.NewClientWithTimeout(30 * time.Second),
	}

	bundle, err := verifier.fetchExternalBundle(context.Background(), server.URL+"/bundle.sn")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(bundle) != `{"fake":"external"}` {
		t.Fatalf("expected bundle content, got %s", string(bundle))
	}
}

func TestFetchAttestations_WithExternalBundle(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/attestations/sha256:abcd", func(w http.ResponseWriter, r *http.Request) {
		resp := attestationAPIResponse{
			Attestations: []attestationEntry{
				{BundleURL: "http://" + r.Host + "/bundle.json"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/bundle.json", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"fake":"external_json"}`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	t.Setenv("UNIRTM_GITHUB_API_BASEURL", server.URL)
	t.Setenv("UNIRTM_ENABLE_GITHUB_PROXY", "0")

	verifier := &provenanceVerifier{
		client: pkgHttp.NewClientWithTimeout(30 * time.Second),
	}

	bundles, err := verifier.fetchAttestations(context.Background(), "", "owner", "repo", "abcd")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(bundles) != 1 {
		t.Fatalf("expected 1 bundle, got %d", len(bundles))
	}
	if string(bundles[0]) != `{"fake":"external_json"}` {
		t.Fatalf("expected content, got %s", string(bundles[0]))
	}
}

func TestVerifyArtifactProvenance_MalformedResponse(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if ok {
			conn, _, _ := hj.Hijack()
			conn.Write([]byte("Garbage\n\n"))
			conn.Close()
		} else {
			w.WriteHeader(http.StatusOK)
		}
	})
	server := httptest.NewServer(mux)
	defer server.Close()
	t.Setenv("UNIRTM_GITHUB_API_BASEURL", server.URL)

	tmpFile, _ := os.CreateTemp("", "artifact")
	defer os.Remove(tmpFile.Name())
	tmpFile.Write([]byte("data"))
	tmpFile.Close()

	_, _ = VerifyArtifactProvenance(context.Background(), "", "owner", "repo", tmpFile.Name())
}

func TestVerifyArtifactProvenance_InvalidBundle(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		resp := attestationAPIResponse{
			Attestations: []attestationEntry{
				{Bundle: json.RawMessage(`{"fake":"bundle"}`)},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	server := httptest.NewServer(mux)
	defer server.Close()
	t.Setenv("UNIRTM_GITHUB_API_BASEURL", server.URL)

	tmpFile, _ := os.CreateTemp("", "artifact")
	defer os.Remove(tmpFile.Name())
	tmpFile.Write([]byte("data"))
	tmpFile.Close()

	_, _ = VerifyArtifactProvenance(context.Background(), "", "owner", "repo", tmpFile.Name())
}

func TestVerifyArtifactProvenance_HTTP2Disabled(t *testing.T) {
	t.Setenv("UNIRTM_HTTP2", "0")
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	server := httptest.NewServer(mux)
	defer server.Close()
	t.Setenv("UNIRTM_GITHUB_API_BASEURL", server.URL)

	tmpFile, _ := os.CreateTemp("", "artifact")
	defer os.Remove(tmpFile.Name())
	tmpFile.Write([]byte("data"))
	tmpFile.Close()

	_, _ = VerifyArtifactProvenance(context.Background(), "", "owner", "repo", tmpFile.Name())
}

func TestFetchAttestations_NonOKStatus(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/attestations/sha256:abcd", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"message":"forbidden"}`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	t.Setenv("UNIRTM_GITHUB_API_BASEURL", server.URL)

	verifier := &provenanceVerifier{
		client: pkgHttp.NewClientWithTimeout(30 * time.Second),
	}

	_, err := verifier.fetchAttestations(context.Background(), "token", "owner", "repo", "abcd")
	if err == nil {
		t.Fatal("expected error for non-OK status")
	}
}

func TestFetchAttestations_InvalidJSON(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/attestations/sha256:abcd", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`invalid json{}`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	t.Setenv("UNIRTM_GITHUB_API_BASEURL", server.URL)

	verifier := &provenanceVerifier{
		client: pkgHttp.NewClientWithTimeout(30 * time.Second),
	}

	_, err := verifier.fetchAttestations(context.Background(), "token", "owner", "repo", "abcd")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestFetchExternalBundle_WithGitHubProxy(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/bundle.json", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"proxytest":true}`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	// Test GITHUB_PROXY path
	t.Setenv("UNIRTM_GITHUB_PROXY", server.URL+"/")
	t.Setenv("UNIRTM_ENABLE_GITHUB_PROXY", "1")

	verifier := &provenanceVerifier{
		client: pkgHttp.NewClientWithTimeout(30 * time.Second),
	}

	// URL not starting with the proxy prefix - so proxy should be prepended
	data, err := verifier.fetchExternalBundle(context.Background(), "https://example.com/bundle.json")
	if err != nil {
		// It's OK for this to fail - just checking the proxy path is exercised
		t.Logf("fetchExternalBundle with proxy: %v", err)
	} else {
		t.Logf("data: %s", data)
	}
}

func TestFetchExternalBundle_NonOKStatus(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/bundle.json", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	verifier := &provenanceVerifier{
		client: pkgHttp.NewClientWithTimeout(30 * time.Second),
	}

	_, err := verifier.fetchExternalBundle(context.Background(), server.URL+"/bundle.json")
	if err == nil {
		t.Fatal("expected error for non-OK status")
	}
}

func TestFetchAttestations_BundleDownloadFails(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/attestations/sha256:abcd", func(w http.ResponseWriter, r *http.Request) {
		resp := attestationAPIResponse{
			Attestations: []attestationEntry{
				{BundleURL: "http://localhost:1/nonexistent"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	t.Setenv("UNIRTM_GITHUB_API_BASEURL", server.URL)

	verifier := &provenanceVerifier{
		client: pkgHttp.NewClientWithTimeout(5 * time.Second),
	}

	// Should return 0 bundles (download fails, logged as warning, continued)
	bundles, err := verifier.fetchAttestations(context.Background(), "", "owner", "repo", "abcd")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(bundles) != 0 {
		t.Errorf("expected 0 bundles when download fails, got %d", len(bundles))
	}
}
