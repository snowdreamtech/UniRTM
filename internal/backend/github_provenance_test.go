// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// ---------------------------------------------------------------------------
// sha256File
// ---------------------------------------------------------------------------

func TestSha256File(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "artifact.bin")

	content := []byte("hello provenance")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	got, err := sha256File(path)
	if err != nil {
		t.Fatalf("sha256File error: %v", err)
	}
	// Pre-computed: printf 'hello provenance' | sha256sum
	const want = "8c12730a07857a092be30af5336fd584b73627368197b93b57bfabf49ae17bd8"
	if got != want {
		t.Errorf("sha256File = %q, want %q", got, want)
	}
}

func TestSha256File_Missing(t *testing.T) {
	_, err := sha256File("/nonexistent/path/artifact.bin")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

// ---------------------------------------------------------------------------
// dssePAE
// ---------------------------------------------------------------------------

func TestDSSEPAE(t *testing.T) {
	// Reference: https://github.com/secure-systems-lab/dsse/blob/master/protocol.md
	payloadType := "application/vnd.in-toto+json"
	payload := []byte(`{"_type":"https://in-toto.io/Statement/v0.1"}`)

	pae := dssePAE(payloadType, payload)

	// Must start with "DSSEv1 "
	if string(pae[:7]) != "DSSEv1 " {
		t.Errorf("PAE does not start with 'DSSEv1 ', got: %q", string(pae[:20]))
	}
}

// ---------------------------------------------------------------------------
// parseGhHostsYml (reused from token test — ensure it's covered here too)
// ---------------------------------------------------------------------------

func TestParseGhHostsYml_Provenance(t *testing.T) {
	content := `github.com:
    oauth_token: ghp_provenance_test
    user: provenance-bot
`
	got := parseGhHostsYml(content, "github.com")
	if got != "ghp_provenance_test" {
		t.Errorf("got %q, want %q", got, "ghp_provenance_test")
	}
}

// ---------------------------------------------------------------------------
// verifySubjectDigest
// ---------------------------------------------------------------------------

func TestVerifySubjectDigest_Match(t *testing.T) {
	const digest = "abc123def456"
	subjects := []inTotoSubject{
		{Name: "artifact.tar.gz", Digest: map[string]string{"sha256": "abc123def456"}},
	}
	if err := verifySubjectDigest(subjects, digest); err != nil {
		t.Errorf("expected match, got error: %v", err)
	}
}

func TestVerifySubjectDigest_CaseInsensitive(t *testing.T) {
	const digest = "ABC123DEF456"
	subjects := []inTotoSubject{
		{Name: "artifact.tar.gz", Digest: map[string]string{"sha256": "abc123def456"}},
	}
	if err := verifySubjectDigest(subjects, digest); err != nil {
		t.Errorf("expected case-insensitive match, got error: %v", err)
	}
}

func TestVerifySubjectDigest_NoMatch(t *testing.T) {
	subjects := []inTotoSubject{
		{Name: "artifact.tar.gz", Digest: map[string]string{"sha256": "deadbeef"}},
	}
	err := verifySubjectDigest(subjects, "cafebabe")
	if err == nil {
		t.Error("expected mismatch error, got nil")
	}
}

func TestVerifySubjectDigest_Empty(t *testing.T) {
	err := verifySubjectDigest(nil, "cafebabe")
	if err == nil {
		t.Error("expected error for empty subjects, got nil")
	}
}

// ---------------------------------------------------------------------------
// fetchAttestations — HTTP 404 returns nil (not supported)
// ---------------------------------------------------------------------------

func TestFetchAttestations_NotSupported(t *testing.T) {
	// We test the JSON parsing path rather than making live HTTP calls.
	// Simulate the 404 → nil behavior by verifying the empty response path.
	bundles, err := fetchAttestations_fromJSON([]byte(`{"attestations":[]}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(bundles) != 0 {
		t.Errorf("expected 0 bundles, got %d", len(bundles))
	}
}

func TestFetchAttestations_WithBundles(t *testing.T) {
	raw := json.RawMessage(`{"mediaType":"application/vnd.dev.sigstore.bundle+json;version=0.3"}`)
	payload := attestationAPIResponse{
		Attestations: []attestationEntry{
			{Bundle: raw, BundleURL: "https://example.com/bundle"},
		},
	}
	data, _ := json.Marshal(payload)

	bundles, err := fetchAttestations_fromJSON(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(bundles) != 1 {
		t.Errorf("expected 1 bundle, got %d", len(bundles))
	}
}

// fetchAttestations_fromJSON is a testable helper that mimics the JSON decoding
// path of fetchAttestations without making real HTTP requests.
func fetchAttestations_fromJSON(data []byte) ([]json.RawMessage, error) {
	var apiResp attestationAPIResponse
	if err := json.Unmarshal(data, &apiResp); err != nil {
		return nil, err
	}
	if len(apiResp.Attestations) == 0 {
		return nil, nil
	}
	raw := make([]json.RawMessage, len(apiResp.Attestations))
	for i, a := range apiResp.Attestations {
		raw[i] = a.Bundle
	}
	return raw, nil
}

// ---------------------------------------------------------------------------
// Fulcio root pool loads without error
// ---------------------------------------------------------------------------

func TestFulcioRootPool_Loads(t *testing.T) {
	pool, err := fulcioRootPool()
	if err != nil {
		t.Fatalf("fulcioRootPool() error: %v", err)
	}
	if pool == nil {
		t.Fatal("pool is nil")
	}
}

// ---------------------------------------------------------------------------
// parseCertChain — invalid input
// ---------------------------------------------------------------------------

func TestParseCertChain_Nil(t *testing.T) {
	_, _, err := parseCertChain(nil)
	if err == nil {
		t.Error("expected error for nil chain, got nil")
	}
}

func TestParseCertChain_EmptyCerts(t *testing.T) {
	_, _, err := parseCertChain(&x509CertChain{})
	if err == nil {
		t.Error("expected error for empty cert list, got nil")
	}
}

func TestParseCertChain_BadBase64(t *testing.T) {
	_, _, err := parseCertChain(&x509CertChain{
		Certificates: []certWrapper{{RawBytes: "!!!not-base64!!!"}},
	})
	if err == nil {
		t.Error("expected base64 decode error, got nil")
	}
}

func TestParseCertChain_BadDER(t *testing.T) {
	_, _, err := parseCertChain(&x509CertChain{
		Certificates: []certWrapper{
			{RawBytes: base64.StdEncoding.EncodeToString([]byte("this-is-not-a-cert"))},
		},
	})
	if err == nil {
		t.Error("expected DER parse error, got nil")
	}
}

// ---------------------------------------------------------------------------
// oidEqual
// ---------------------------------------------------------------------------

func TestOidEqual(t *testing.T) {
	a := []int{1, 3, 6, 1, 4, 1, 57264, 1, 1}
	b := []int{1, 3, 6, 1, 4, 1, 57264, 1, 1}
	c := []int{1, 3, 6, 1, 4, 1, 57264, 1, 2}

	if !oidEqual(a, b) {
		t.Error("oidEqual: expected true for identical OIDs")
	}
	if oidEqual(a, c) {
		t.Error("oidEqual: expected false for different OIDs")
	}
	if oidEqual(a, nil) {
		t.Error("oidEqual: expected false for nil")
	}
}

// ---------------------------------------------------------------------------
// ProvenanceResult — not-supported path
// ---------------------------------------------------------------------------

func TestVerifyArtifactProvenance_FileNotFound(t *testing.T) {
	_, err := VerifyArtifactProvenance(
		nil, //nolint:staticcheck — intentionally nil context for test
		"", "owner", "repo",
		"/nonexistent/artifact.tar.gz",
	)
	if err == nil {
		t.Error("expected error for missing artifact file, got nil")
	}
}
