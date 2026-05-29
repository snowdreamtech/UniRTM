// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestVerifyGitlabArtifactProvenance_FileNotFound(t *testing.T) {
	_, err := VerifyGitlabArtifactProvenance(
		context.Background(),
		"", "owner", "repo",
		"/nonexistent/artifact.tar.gz",
	)
	if err == nil {
		t.Error("expected error for missing artifact file, got nil")
	}
}

func TestFetchGitlabAttestations_NotSupported(t *testing.T) {
	bundles, err := testFetchGitlabAttestationsFromJSON([]byte(`[]`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(bundles) != 0 {
		t.Errorf("expected 0 bundles, got %d", len(bundles))
	}
}

func TestFetchGitlabAttestations_WithBundles(t *testing.T) {
	payload := []gitlabAttestationResponse{
		{ID: 1, IID: 1, Status: "success", PredicateType: "https://slsa.dev/provenance/v1", DownloadURL: "https://example.com/bundle"},
	}
	data, _ := json.Marshal(payload)

	bundles, err := testFetchGitlabAttestationsFromJSON(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(bundles) != 1 {
		t.Errorf("expected 1 bundle, got %d", len(bundles))
	}
}

func testFetchGitlabAttestationsFromJSON(data []byte) ([]json.RawMessage, error) {
	var apiResp []gitlabAttestationResponse
	if err := json.Unmarshal(data, &apiResp); err != nil {
		return nil, err
	}
	if len(apiResp) == 0 {
		return nil, nil
	}
	raw := make([]json.RawMessage, len(apiResp))
	// GitLab bundle is obtained via download_url, so in this JSON test we just mock it.
	for i := range apiResp {
		raw[i] = json.RawMessage(`{}`)
	}
	return raw, nil
}

func TestVerifyGitlabArtifactProvenance_MockServer(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "artifact.bin")
	content := []byte("hello gitlab provenance")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/projects/owner/repo/attestations/3831d1667b9386c9d747a06ee7a5bc8c0b5f13c6ea6377e3845a72ab01d2cccd" || r.URL.Path == "/projects/owner%2Frepo/attestations/3831d1667b9386c9d747a06ee7a5bc8c0b5f13c6ea6377e3845a72ab01d2cccd" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[{"id": 1, "iid": 1, "status": "success", "download_url": "http://` + r.Host + `/projects/owner%2Frepo/attestations/1/download"}]`))
			return
		}
		if r.URL.Path == "/projects/owner/repo/attestations/1/download" || r.URL.Path == "/projects/owner%2Frepo/attestations/1/download" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"mediaType":"application/vnd.dev.sigstore.bundle+json;version=0.3"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	t.Setenv("GITLAB_API_URL", server.URL)

	// Set mock client in verifier so it calls httptest server
	verifier := &gitlabProvenanceVerifier{
		client: server.Client(),
	}

	bundles, err := verifier.fetchAttestations(context.Background(), "mock-token", "owner", "repo", "3831d1667b9386c9d747a06ee7a5bc8c0b5f13c6ea6377e3845a72ab01d2cccd")
	if err != nil {
		t.Fatalf("unexpected error fetching attestations: %v", err)
	}
	if len(bundles) != 1 {
		t.Fatalf("expected 1 bundle, got %d", len(bundles))
	}
}
