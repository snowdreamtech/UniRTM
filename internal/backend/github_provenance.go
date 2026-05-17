// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package backend provides the GitHub Provenance (SLSA attestation) verifier.
package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang/snappy"
	"github.com/sigstore/sigstore-go/pkg/root"
	"github.com/sigstore/sigstore-go/pkg/verify"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
	"github.com/snowdreamtech/unirtm/internal/pkg/logger"
)

// VerifyArtifactProvenance checks the GitHub attestation for the artifact at
// artifactPath against the repository owner/repo.
func VerifyArtifactProvenance(
	ctx context.Context,
	token, owner, repo, artifactPath string,
) (*ProvenanceResult, error) {
	if trans, ok := http.DefaultTransport.(*http.Transport); ok {
		if env.Get("HTTP2") == "0" {
			pkgHttp.DisableHTTP2(trans)
			logger.Debug("provenance: globally disabled HTTP/2 for verification (manual via env)")
		}
	}

	verifier := &provenanceVerifier{
		client: pkgHttp.NewClientWithTimeout(30 * time.Second),
	}

	result, err := verifier.verify(ctx, token, owner, repo, artifactPath)
	if err != nil && strings.Contains(err.Error(), "malformed HTTP response") {
		if trans, ok := http.DefaultTransport.(*http.Transport); ok {
			logger.Warn("provenance: detected malformed HTTP response, smartly downgrading to HTTP/1.1 and retrying...")
			pkgHttp.DisableHTTP2(trans)
			verifier = &provenanceVerifier{
				client: pkgHttp.NewClientWithTimeout(30 * time.Second),
			}
			return verifier.verify(ctx, token, owner, repo, artifactPath)
		}
	}
	return result, err
}

type provenanceVerifier struct {
	client *http.Client
}

func (v *provenanceVerifier) verify(
	ctx context.Context,
	token, owner, repo, artifactPath string,
) (*ProvenanceResult, error) {
	digest, err := sha256File(artifactPath)
	if err != nil {
		return nil, fmt.Errorf("provenance: compute digest: %w", err)
	}

	logger.Debug("provenance: fetching attestations from GitHub", map[string]interface{}{"owner": owner, "repo": repo, "digest": digest})
	bundles, err := v.fetchAttestations(ctx, token, owner, repo, digest)
	if err != nil {
		return nil, err
	}
	logger.Debug("provenance: found attestation bundles", map[string]interface{}{"count": len(bundles)})

	if len(bundles) == 0 {
		return &ProvenanceResult{Supported: false}, nil
	}

	trustedMaterial, err := sigstoreTrustedRoot()
	if err != nil {
		return nil, fmt.Errorf("provenance: load TUF trusted root: %w", err)
	}

	expectedRepo := owner + "/" + repo
	repoRegex := "^https://github\\.com/" + owner + "/" + repo + "/"
	
	mainIdentity, err := verify.NewShortCertificateIdentity(
		"https://token.actions.githubusercontent.com",
		"",
		"",
		repoRegex,
	)
	if err != nil {
		return nil, fmt.Errorf("build main identity: %w", err)
	}

	globalIdentity, err := verify.NewShortCertificateIdentity(
		"",
		".*",
		"",
		"^https://dotcom\\.releases\\.github\\.com$",
	)
	if err != nil {
		return nil, fmt.Errorf("build global identity: %w", err)
	}

	sigstoreVerifier := &SigstoreVerifier{
		TrustedMaterial: trustedMaterial,
		Identities:      []verify.CertificateIdentity{mainIdentity, globalIdentity},
		ExpectedRepo:    expectedRepo,
	}

	return sigstoreVerifier.VerifyBundles(bundles, digest)
}

// -----------------------------------------------------------------------------
// TUF-backed Sigstore trusted root (singleton with 24 h refresh)
// -----------------------------------------------------------------------------

var (
	liveTrustedRootOnce sync.Once
	liveTrustedRoot     *root.LiveTrustedRoot
	liveTrustedRootErr  error
)

func sigstoreTrustedRoot() (*root.LiveTrustedRoot, error) {
	liveTrustedRootOnce.Do(func() {
		logger.Debug("provenance: initializing GitHub TUF trusted root")
		liveTrustedRoot, liveTrustedRootErr = InitializeTUFRoot(
			"github-tuf",
			"https://tuf-repo.github.com",
			githubTufRoot,
		)
		if liveTrustedRootErr != nil {
			logger.Error("provenance: failed to initialize GitHub TUF root", map[string]interface{}{"error": liveTrustedRootErr.Error()})
		} else {
			logger.Debug("provenance: GitHub TUF root initialized successfully")
		}
	})
	return liveTrustedRoot, liveTrustedRootErr
}

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

func (v *provenanceVerifier) fetchAttestations(
	ctx context.Context,
	token, owner, repo, digest string,
) ([]json.RawMessage, error) {
	apiBase := env.Get("GITHUB_API_BASEURL")
	if apiBase == "" {
		apiBase = "https://api.github.com"
	}
	apiBase = strings.TrimSuffix(apiBase, "/")

	url := fmt.Sprintf(
		"%s/repos/%s/%s/attestations/sha256:%s",
		apiBase, owner, repo, digest,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("provenance: create request: %w", err)
	}
	req.Header.Set("User-Agent", "unirtm/"+env.GitTag)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := v.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("provenance: fetch attestations: %w", err)
	}
	defer resp.Body.Close()

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

		if a.BundleURL != "" {
			fmt.Printf("ℹ provenance: fetching external bundle %d/%d from URL...\n", i+1, len(apiResp.Attestations))
			bundleData, err := v.fetchExternalBundle(ctx, a.BundleURL)
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

func (v *provenanceVerifier) fetchExternalBundle(ctx context.Context, urlStr string) (json.RawMessage, error) {
	finalURL := urlStr
	githubProxy := env.Get("GITHUB_PROXY")
	if githubProxy != "" && env.Get("ENABLE_GITHUB_PROXY") == "1" {
		if !strings.HasPrefix(urlStr, githubProxy) {
			finalURL = githubProxy + urlStr
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, finalURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "unirtm/"+env.GitTag)

	resp, err := v.client.Do(req)
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

	if strings.Contains(urlStr, ".sn") || strings.HasSuffix(urlStr, ".sn") {
		decoded, err := snappy.Decode(nil, body)
		if err != nil {
			return nil, fmt.Errorf("provenance: decompress snappy bundle: %w", err)
		}
		return decoded, nil
	}

	return body, nil
}
