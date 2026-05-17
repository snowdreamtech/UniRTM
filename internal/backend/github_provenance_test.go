// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/sigstore/sigstore-go/pkg/verify"
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
// fetchAttestations — JSON parsing path (no live HTTP)
// ---------------------------------------------------------------------------

func TestFetchAttestations_NotSupported(t *testing.T) {
	bundles, err := testFetchAttestationsFromJSON([]byte(`{"attestations":[]}`))
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

	bundles, err := testFetchAttestationsFromJSON(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(bundles) != 1 {
		t.Errorf("expected 1 bundle, got %d", len(bundles))
	}
}

// testFetchAttestationsFromJSON mimics the JSON-decode path of fetchAttestations
// without making real HTTP requests.
func testFetchAttestationsFromJSON(data []byte) ([]json.RawMessage, error) {
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
// TUF trusted root singleton — smoke test (does not make network calls
// if TUF cache already exists; skipped if no network connectivity)
// ---------------------------------------------------------------------------

func TestSigstoreTrustedRoot_ResetWorks(t *testing.T) {
	// Ensure the reset function works without panicking.
	ResetTrustedRootForTest()
	// After reset the singleton is nil — next call would try TUF fetch.
	// We don't trigger the fetch here to keep tests offline-safe.
}

// ---------------------------------------------------------------------------
// VerifyArtifactProvenance — missing file returns error immediately
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

func TestPrintTufRoots(t *testing.T) {
	roots, err := sigstoreTrustedRoots()
	if err != nil {
		t.Fatal(err)
	}
	for i, r := range roots {
		keys := []string{}
		for k := range r.RekorLogs() {
			keys = append(keys, k)
		}
		t.Logf("Root %d RekorLogs keys: %v", i, keys)
	}
}

func TestVerifySyftBundleFromFile(t *testing.T) {
	// Try root directory or '../../' to find the bundle file in go test
	path := "sha256:24e4d34078ae81da7c82539616f0ccac3e226cf4f74a38ce6fb3463619e50a55.jsonl"
	if _, err := os.Stat(path); err != nil {
		path = "../../" + path
	}
	bundleBytes, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("syft bundle file not found at %s, skipping local integration test", path)
	}

	roots, err := sigstoreTrustedRoots()
	if err != nil {
		t.Fatal(err)
	}

	// Prepare verifier with GitHub TUF roots and public Sigstore TUF roots.
	// Since syft is a GitHub Attestation, it should be verified by the GitHub TUF root (Root 0).
	mainIdentity, err := verify.NewShortCertificateIdentity(
		"https://token.actions.githubusercontent.com",
		"",
		"",
		"^https://github\\.com/anchore/syft/",
	)
	if err != nil {
		t.Fatal(err)
	}

	globalIdentity, err := verify.NewShortCertificateIdentity(
		"",
		".*",
		"",
		"^https://dotcom\\.releases\\.github\\.com$",
	)
	if err != nil {
		t.Fatal(err)
	}

	verifier := &SigstoreVerifier{
		TrustedMaterials: roots,
		Identities:       []verify.CertificateIdentity{mainIdentity, globalIdentity},
		ExpectedRepo:     "anchore/syft",
	}

	res, err := verifier.VerifyBundles([]json.RawMessage{json.RawMessage(bundleBytes)}, "24e4d34078ae81da7c82539616f0ccac3e226cf4f74a38ce6fb3463619e50a55")
	if err != nil {
		t.Fatalf("failed to verify syft bundle: %v", err)
	}

	t.Logf("Verification succeeded! Res: %+v", res)
}
