// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package backend provides the GitHub Provenance (SLSA attestation) verifier.
//
// Verification is backed by sigstore-go v1.1.4 with TUF-managed trust roots,
// which means:
//   - Fulcio CA certificates are fetched from the Sigstore TUF repository
//     (https://tuf-repo-cdn.sigstore.dev) and cached locally
//   - Revoked or expired CA certificates are automatically rejected after
//     the next TUF refresh (default: every 24 hours)
//   - The local TUF cache lives in $UNIRTM_TUF_CACHE_DIR or the OS temp dir
package backend

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/golang/snappy"
	"github.com/sigstore/sigstore-go/pkg/bundle"
	"github.com/sigstore/sigstore-go/pkg/root"
	"github.com/sigstore/sigstore-go/pkg/tuf"
	"github.com/sigstore/sigstore-go/pkg/verify"
	"github.com/snowdreamtech/unirtm/internal/pkg/logger"
)

// -----------------------------------------------------------------------------
// Public API
// -----------------------------------------------------------------------------

// ProvenanceResult is returned by VerifyArtifactProvenance.
type ProvenanceResult struct {
	// Supported is false when the project publishes no GitHub attestations.
	// Callers MUST NOT treat this as an error — simply skip further checks.
	Supported bool

	// Verified is true when the attestation bundle passed all checks.
	Verified bool

	// Repository is the source repository recorded in the Fulcio cert SAN,
	// e.g. "octocat/hello-world".
	Repository string

	// WorkflowRef is the triggering workflow path recorded in the cert, e.g.
	// "octocat/hello-world/.github/workflows/release.yml@refs/heads/main".
	WorkflowRef string

	// PredicateType is the in-toto predicate URI in the signed statement, e.g.
	// "https://slsa.dev/provenance/v1".
	PredicateType string

	// BuilderID is the SLSA builder identifier from the certificate extension.
	BuilderID string
}

// VerifyArtifactProvenance checks the GitHub attestation for the artifact at
// artifactPath against the repository owner/repo.
//
//   - result.Supported == false, err == nil  → project has no attestations; skip.
//   - result.Supported == true, err != nil   → attestations exist but failed; hard error.
//   - result.Supported == true, err == nil   → fully verified; proceed safely.
func VerifyArtifactProvenance(
	ctx context.Context,
	token, owner, repo, artifactPath string,
) (*ProvenanceResult, error) {
	// 1. Compute SHA-256 of the artifact on disk.
	digest, err := sha256File(artifactPath)
	if err != nil {
		return nil, fmt.Errorf("provenance: compute digest: %w", err)
	}

	// 2. Fetch attestation bundles from the GitHub API.
	logger.Debug("provenance: fetching attestations from GitHub", map[string]interface{}{"owner": owner, "repo": repo, "digest": digest})
	bundles, err := fetchAttestations(ctx, token, owner, repo, digest)
	if err != nil {
		return nil, err
	}
	logger.Debug("provenance: found attestation bundles", map[string]interface{}{"count": len(bundles)})

	// 3. No attestations → project does not publish provenance → graceful skip.
	if len(bundles) == 0 {
		return &ProvenanceResult{Supported: false}, nil
	}

	// 4. Obtain the TUF-backed Sigstore trusted root (cached, refreshed every 24h).
	trustedMaterial, err := sigstoreTrustedRoot()
	if err != nil {
		return nil, fmt.Errorf("provenance: load TUF trusted root: %w", err)
	}

	// 5. Verify all returned bundles; at least one must pass.
	expectedRepo := owner + "/" + repo
	var lastErr error
	for _, rawBundle := range bundles {
		result, err := verifyBundleWithSigstore(rawBundle, digest, expectedRepo, trustedMaterial)
		if err == nil {
			return result, nil
		}
		lastErr = err
	}

	// All bundles failed verification.
	return nil, fmt.Errorf(
		"provenance: all %d attestation bundle(s) failed verification for %s/%s. Last error: %v",
		len(bundles), owner, repo, lastErr,
	)
}

// -----------------------------------------------------------------------------
// TUF-backed Sigstore trusted root (singleton with 24 h refresh)
// -----------------------------------------------------------------------------

var (
	liveTrustedRootOnce sync.Once
	liveTrustedRoot     *root.LiveTrustedRoot
	liveTrustedRootErr  error
)

// sigstoreTrustedRoot returns the Sigstore public good TUF trusted root.
// The root is initialized once and refreshed automatically every 24 hours
// by the LiveTrustedRoot mechanism inside sigstore-go.
func sigstoreTrustedRoot() (*root.LiveTrustedRoot, error) {
	liveTrustedRootOnce.Do(func() {
		fmt.Println("ℹ provenance: initializing Sigstore TUF trusted root (this may take 30-60s on first run)")
		logger.Debug("provenance: initializing Sigstore TUF trusted root (this may take 30-60s on first run)")
		opts := tuf.DefaultOptions()

		// Allow users to override the TUF cache directory.
		if cacheDir := os.Getenv("UNIRTM_TUF_CACHE_DIR"); cacheDir != "" {
			opts.CachePath = filepath.Join(cacheDir, "sigstore-tuf")
		}

		// TUF initialization happens here (implicitly uses background context inside sigstore-go)
		liveTrustedRoot, liveTrustedRootErr = root.NewLiveTrustedRoot(opts)
		if liveTrustedRootErr != nil {
			logger.Error("provenance: failed to initialize TUF root", map[string]interface{}{"error": liveTrustedRootErr.Error()})
		} else {
			logger.Debug("provenance: Sigstore TUF trusted root initialized")
		}
	})
	return liveTrustedRoot, liveTrustedRootErr
}

// ResetTrustedRootForTest resets the singleton — for testing only.
func ResetTrustedRootForTest() {
	liveTrustedRootOnce = sync.Once{}
	liveTrustedRoot = nil
	liveTrustedRootErr = nil
}

// -----------------------------------------------------------------------------
// GitHub Attestations REST API
// -----------------------------------------------------------------------------

type attestationAPIResponse struct {
	Attestations []attestationEntry `json:"attestations"`
}

type attestationEntry struct {
	Bundle    json.RawMessage `json:"bundle"`
	BundleURL string          `json:"bundle_url"`
}

// fetchAttestations queries GET /repos/{owner}/{repo}/attestations/sha256:{digest}.
// Returns nil slice (not an error) when the project publishes no attestations (HTTP 404).
func fetchAttestations(
	ctx context.Context,
	token, owner, repo, digest string,
) ([]json.RawMessage, error) {
	url := fmt.Sprintf(
		"https://api.github.com/repos/%s/%s/attestations/sha256:%s",
		owner, repo, digest,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("provenance: create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("provenance: fetch attestations: %w", err)
	}
	defer resp.Body.Close()

	// 404 → project does not publish attestations → gracefully skip.
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("provenance: attestations API returned %d: %s", resp.StatusCode, body)
	}

	var apiResp attestationAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("provenance: decode API response: %w", err)
	}
	fmt.Printf("ℹ provenance: found %d attestation(s)\n", len(apiResp.Attestations))
	raw := make([]json.RawMessage, 0, len(apiResp.Attestations))
	for i, a := range apiResp.Attestations {
		if len(a.Bundle) > 0 && string(a.Bundle) != "null" {
			raw = append(raw, a.Bundle)
			continue
		}

		// Fallback to bundle_url if inline bundle is missing/null
		if a.BundleURL != "" {
			fmt.Printf("ℹ provenance: fetching external bundle %d/%d from URL...\n", i+1, len(apiResp.Attestations))
			bundleData, err := fetchExternalBundle(ctx, a.BundleURL)
			if err != nil {
				logger.Warn("provenance: failed to fetch external bundle", map[string]interface{}{
					"url":   a.BundleURL,
					"error": err.Error(),
				})
				continue
			}
			raw = append(raw, bundleData)
		}
	}
	return raw, nil
}

// fetchExternalBundle downloads a Sigstore bundle from an external URL.
func fetchExternalBundle(ctx context.Context, url string) (json.RawMessage, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("external bundle API returned %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// GitHub often returns Snappy-compressed bundles (with .sn suffix).
	// These are usually in Snappy Raw format (not framed).
	if strings.Contains(url, ".sn") || strings.HasSuffix(url, ".sn") {
		decoded, err := snappy.Decode(nil, body)
		if err != nil {
			// If raw decode fails, try framed if it might be framed (though usually it's raw)
			// But for now, if raw fails, we return error as it's the expected format.
			return nil, fmt.Errorf("provenance: decompress snappy bundle: %w", err)
		}
		return decoded, nil
	}

	return body, nil
}

// -----------------------------------------------------------------------------
// Bundle verification via sigstore-go
// -----------------------------------------------------------------------------

// verifyBundleWithSigstore verifies a single Sigstore bundle using the official
// sigstore-go library backed by TUF-managed trust material.
//
// Verification chain:
//   1. Parse the bundle JSON into a sigstore-go Bundle.
//   2. Build the verifier with transparency log + observer timestamp enforcement.
//   3. Build an identity policy: issuer must be GitHub Actions OIDC; SAN must
//      contain the expected repository path.
//   4. Build an artifact digest policy enforcing SHA-256 match.
//   5. Call verifier.Verify() — sigstore-go handles cert chain, Rekor,
//      DSSE signature, and subject digest internally.
//   6. Extract the workflow and predicate info from the verification result.
func verifyBundleWithSigstore(
	rawBundle json.RawMessage,
	artifactDigest, expectedRepo string,
	trustedMaterial root.TrustedMaterial,
) (*ProvenanceResult, error) {
	// Step 1: Parse bundle
	b := &bundle.Bundle{}
	if err := b.UnmarshalJSON(rawBundle); err != nil {
		return nil, fmt.Errorf("parse bundle JSON: %w", err)
	}

	// Step 2: Build verifier
	// - WithTransparencyLog(1): require at least 1 Rekor tlog entry
	// - WithObserverTimestamps(1): require at least 1 observer timestamp
	//   (either tlog integrated time or RFC 3161 TSA timestamp)
	verifier, err := verify.NewSignedEntityVerifier(
		trustedMaterial,
		verify.WithTransparencyLog(1),
		verify.WithObserverTimestamps(1),
	)
	if err != nil {
		return nil, fmt.Errorf("build verifier: %w", err)
	}

	// Step 3: Certificate identity policy
	// GitHub Actions OIDC issuer + SAN regex matching the repository.
	//
	// SAN format for GitHub Actions:
	//   https://github.com/{owner}/{repo}/.github/workflows/{workflow}@refs/...
	sanRegex := fmt.Sprintf("^https://github\\.com/%s/", regexp_escape(expectedRepo))
	identity, err := verify.NewShortCertificateIdentity(
		"https://token.actions.githubusercontent.com", // GitHub Actions OIDC issuer
		"",        // issuerRegex: empty (exact match above)
		"",        // sanValue:    empty (use regex below)
		sanRegex,  // sanRegex
	)
	if err != nil {
		return nil, fmt.Errorf("build certificate identity: %w", err)
	}

	// Step 4: Artifact digest policy
	digestBytes, err := hex.DecodeString(artifactDigest)
	if err != nil {
		return nil, fmt.Errorf("decode artifact digest: %w", err)
	}

	policy := verify.NewPolicy(
		verify.WithArtifactDigest("sha256", digestBytes),
		verify.WithCertificateIdentity(identity),
	)

	// Step 5: Verify
	result, err := verifier.Verify(b, policy)
	if err != nil {
		return nil, fmt.Errorf("sigstore verification: %w", err)
	}

	// Step 6: Extract metadata from the verification result
	provResult := &ProvenanceResult{
		Supported:  true,
		Verified:   true,
		Repository: expectedRepo,
	}

	if result.Signature != nil && result.Signature.Certificate != nil {
		// Certificate is a *certificate.Summary
		provResult.WorkflowRef = result.Signature.Certificate.SubjectAlternativeName
		provResult.BuilderID = result.Signature.Certificate.Extensions.BuildSignerURI
	}

	if result.Statement != nil {
		provResult.PredicateType = result.Statement.GetPredicateType()
	}

	return provResult, nil
}

// regexp_escape escapes special regex characters in a plain string.
func regexp_escape(s string) string {
	replacer := strings.NewReplacer(
		`.`, `\.`,
		`-`, `\-`,
		`_`, `\_`,
		`/`, `\/`,
	)
	return replacer.Replace(s)
}

// -----------------------------------------------------------------------------
// Utilities
// -----------------------------------------------------------------------------

// sha256File computes the hex-encoded SHA-256 digest of the file at path.
func sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
