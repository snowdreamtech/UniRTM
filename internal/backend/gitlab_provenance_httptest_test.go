// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
)

func TestGitlabFetchAttestations_LiveMock(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/projects/owner%2Frepo/attestations/abcd", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"download_url":"http://` + r.Host + `/bundle.json"}]`))
	})
	mux.HandleFunc("/bundle.json", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"fake":"gitlab_bundle"}`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	t.Setenv("UNIRTM_GITLAB_API_URL", server.URL)

	verifier := &gitlabProvenanceVerifier{
		client: pkgHttp.NewClientWithTimeout(30 * time.Second),
	}

	bundles, err := verifier.fetchAttestations(context.Background(), "token", "owner", "repo", "abcd")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(bundles) != 1 {
		t.Fatalf("expected 1 bundle, got %d", len(bundles))
	}
	if string(bundles[0]) != `{"fake":"gitlab_bundle"}` {
		t.Fatalf("expected gitlab_bundle, got %s", string(bundles[0]))
	}
}

func TestGitlabFetchAttestations_NotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/projects/owner%2Frepo/attestations/abcd", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	t.Setenv("UNIRTM_GITLAB_API_URL", server.URL)

	verifier := &gitlabProvenanceVerifier{
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

func TestVerifyGitlabArtifactProvenance_MalformedResponse(t *testing.T) {
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
	t.Setenv("UNIRTM_GITLAB_API_URL", server.URL)

	tmpFile, _ := os.CreateTemp("", "artifact")
	defer os.Remove(tmpFile.Name())
	tmpFile.Write([]byte("data"))
	tmpFile.Close()

	_, _ = VerifyGitlabArtifactProvenance(context.Background(), "", "owner", "repo", tmpFile.Name())
}

func TestVerifyGitlabArtifactProvenance_InvalidBundle(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"download_url":"http://` + r.Host + `/bundle.json"}]`))
	})
	mux.HandleFunc("/bundle.json", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"fake":"gitlab_bundle"}`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()
	t.Setenv("UNIRTM_GITLAB_API_URL", server.URL)

	tmpFile, _ := os.CreateTemp("", "artifact")
	defer os.Remove(tmpFile.Name())
	tmpFile.Write([]byte("data"))
	tmpFile.Close()

	_, _ = VerifyGitlabArtifactProvenance(context.Background(), "", "owner", "repo", tmpFile.Name())
}

func TestVerifyGitlabArtifactProvenance_HTTP2Disabled(t *testing.T) {
	t.Setenv("UNIRTM_HTTP2", "0")
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	server := httptest.NewServer(mux)
	defer server.Close()
	t.Setenv("UNIRTM_GITLAB_API_URL", server.URL)

	tmpFile, _ := os.CreateTemp("", "artifact")
	defer os.Remove(tmpFile.Name())
	tmpFile.Write([]byte("data"))
	tmpFile.Close()

	_, _ = VerifyGitlabArtifactProvenance(context.Background(), "", "owner", "repo", tmpFile.Name())
}

func TestGitlabFetchAttestations_NonOKStatus(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/projects/owner%2Frepo/attestations/abcd", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"message":"forbidden"}`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	t.Setenv("UNIRTM_GITLAB_API_URL", server.URL)

	verifier := &gitlabProvenanceVerifier{
		client: pkgHttp.NewClientWithTimeout(30 * time.Second),
	}

	_, err := verifier.fetchAttestations(context.Background(), "token", "owner", "repo", "abcd")
	if err == nil {
		t.Fatal("expected error for non-OK status")
	}
}

func TestGitlabFetchAttestations_InvalidJSON(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/projects/owner%2Frepo/attestations/abcd", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`invalid json{}`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	t.Setenv("UNIRTM_GITLAB_API_URL", server.URL)

	verifier := &gitlabProvenanceVerifier{
		client: pkgHttp.NewClientWithTimeout(30 * time.Second),
	}

	_, err := verifier.fetchAttestations(context.Background(), "token", "owner", "repo", "abcd")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestGitlabDownloadBundle_WithToken(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/bundle.json", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("PRIVATE-TOKEN") != "my-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"mediaType":"test"}`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	verifier := &gitlabProvenanceVerifier{
		client: pkgHttp.NewClientWithTimeout(30 * time.Second),
	}

	data, err := verifier.downloadBundle(context.Background(), "my-token", server.URL+"/bundle.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != `{"mediaType":"test"}` {
		t.Errorf("unexpected data: %s", string(data))
	}
}

func TestGitlabDownloadBundle_NonOKStatus(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/bundle.json", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	verifier := &gitlabProvenanceVerifier{
		client: pkgHttp.NewClientWithTimeout(30 * time.Second),
	}

	_, err := verifier.downloadBundle(context.Background(), "", server.URL+"/bundle.json")
	if err == nil {
		t.Fatal("expected error for non-OK status in downloadBundle")
	}
}

func TestGitlabFetchAttestations_BundleDownloadFails(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/projects/owner%2Frepo/attestations/abcd", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Return an attestation with a download_url that will fail
		w.Write([]byte(`[{"download_url":"http://localhost:1/nonexistent"}]`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	t.Setenv("UNIRTM_GITLAB_API_URL", server.URL)

	verifier := &gitlabProvenanceVerifier{
		client: pkgHttp.NewClientWithTimeout(5 * time.Second),
	}

	// Should return 0 bundles (download fails, logged as warning, continued)
	bundles, err := verifier.fetchAttestations(context.Background(), "token", "owner", "repo", "abcd")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(bundles) != 0 {
		t.Errorf("expected 0 bundles when download fails, got %d", len(bundles))
	}
}

func TestGitlabFetchAttestations_WithCustomIssuer(t *testing.T) {
	mux := http.NewServeMux()
	// When GITLAB_API_URL is server.URL+"/api/v4", the request goes to /api/v4/projects/...
	mux.HandleFunc("/api/v4/projects/owner%2Frepo/attestations/abcd", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"download_url":"http://` + r.Host + `/bundle.json"}]`))
	})
	mux.HandleFunc("/bundle.json", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"fake":"bundle"}`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	// Test the issuer logic: with /api/v4 suffix
	t.Setenv("UNIRTM_GITLAB_API_URL", server.URL+"/api/v4")

	verifier := &gitlabProvenanceVerifier{
		client: pkgHttp.NewClientWithTimeout(30 * time.Second),
	}

	bundles, err := verifier.fetchAttestations(context.Background(), "", "owner", "repo", "abcd")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(bundles) != 1 {
		t.Errorf("expected 1 bundle, got %d", len(bundles))
	}
}
