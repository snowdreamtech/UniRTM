// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package backend provides the GitLab Provenance (SLSA attestation) verifier.
package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sigstore/sigstore-go/pkg/verify"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
	"github.com/snowdreamtech/unirtm/internal/pkg/logger"
)

// VerifyGitlabArtifactProvenance checks the GitLab attestation for the artifact at
// artifactPath against the repository owner/repo.
func VerifyGitlabArtifactProvenance(
	ctx context.Context,
	token, owner, repo, artifactPath string,
) (*ProvenanceResult, error) {

	verifier := &gitlabProvenanceVerifier{
		client: pkgHttp.NewClientWithTimeout(30 * time.Second),
	}

	result, err := verifier.verify(ctx, token, owner, repo, artifactPath)
	if err != nil && strings.Contains(err.Error(), "malformed HTTP response") {
		logger.Warn("provenance: detected malformed HTTP response, smartly downgrading to HTTP/1.1 and retrying...")
		verifier = &gitlabProvenanceVerifier{
			client: pkgHttp.NewClientWithTimeout(30 * time.Second),
		}
		if trans, ok := verifier.client.Transport.(*http.Transport); ok {
			pkgHttp.DisableHTTP2(trans)
		}
		return verifier.verify(ctx, token, owner, repo, artifactPath)
	}
	return result, err
}

type gitlabProvenanceVerifier struct {
	client *http.Client
}

type gitlabAttestationResponse struct {
	ID            int    `json:"id"`
	IID           int    `json:"iid"`
	Status        string `json:"status"`
	PredicateType string `json:"predicate_type"`
	DownloadURL   string `json:"download_url"`
}

func (v *gitlabProvenanceVerifier) verify(
	ctx context.Context,
	token, owner, repo, artifactPath string,
) (*ProvenanceResult, error) {
	digest, err := sha256File(artifactPath)
	if err != nil {
		return nil, fmt.Errorf("provenance: compute digest: %w", err)
	}

	logger.Debug("provenance: fetching attestations from GitLab", map[string]interface{}{"owner": owner, "repo": repo, "digest": digest})
	bundles, err := v.fetchAttestations(ctx, token, owner, repo, digest)
	if err != nil {
		return nil, err
	}
	logger.Debug("provenance: found GitLab attestation bundles", map[string]interface{}{"count": len(bundles)})

	if len(bundles) == 0 {
		return &ProvenanceResult{Supported: false}, nil
	}

	trustedMaterials, err := sigstoreTrustedRoots()
	if err != nil {
		return nil, fmt.Errorf("provenance: load TUF trusted roots: %w", err)
	}

	expectedRepo := owner + "/" + repo

	// GitLab Fulcio issuer is typically the GitLab domain itself.
	issuer := env.Get("GITLAB_API_URL")
	if issuer == "" {
		issuer = "https://gitlab.com"
	} else {
		// Clean up issuer base URL from API URL
		if idx := strings.Index(issuer, "/api/v4"); idx != -1 {
			issuer = issuer[:idx]
		}
	}

	// For GitLab standard OIDC:
	// OIDC issuer is https://gitlab.com
	// SAN is typically: https://gitlab.com/owner/repo//.gitlab-ci.yml@refs/heads/main
	repoRegex := "^https://" + strings.ReplaceAll(strings.TrimPrefix(issuer, "https://"), ".", "\\.") + "/" + owner + "/" + repo + "/"

	mainIdentity, err := verify.NewShortCertificateIdentity(
		issuer,
		"",
		"",
		repoRegex,
	)
	if err != nil {
		return nil, fmt.Errorf("build main identity: %w", err)
	}

	sigstoreVerifier := &SigstoreVerifier{
		TrustedMaterials: trustedMaterials,
		Identities:       []verify.CertificateIdentity{mainIdentity},
		ExpectedRepo:     expectedRepo,
	}

	return sigstoreVerifier.VerifyBundles(bundles, digest)
}

func (v *gitlabProvenanceVerifier) fetchAttestations(
	ctx context.Context,
	token, owner, repo, digest string,
) ([]json.RawMessage, error) {
	baseURL := env.Get("GITLAB_API_URL")
	if baseURL == "" {
		baseURL = "https://gitlab.com/api/v4"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	encodedRepo := url.PathEscape(owner + "/" + repo)
	urlStr := fmt.Sprintf(
		"%s/projects/%s/attestations/%s",
		baseURL, encodedRepo, digest,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("provenance: create request: %w", err)
	}
	req.Header.Set("User-Agent", "unirtm/"+env.GitTag)
	req.Header.Set("Accept", "application/json")
	if token != "" {
		req.Header.Set("PRIVATE-TOKEN", token)
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
		return nil, fmt.Errorf("provenance: GitLab attestations API returned %d: %s", resp.StatusCode, body)
	}

	var apiResp []gitlabAttestationResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("provenance: decode API response: %w", err)
	}

	fmt.Printf("ℹ provenance: found %d GitLab attestation(s)\n", len(apiResp))

	raw := make([]json.RawMessage, 0, len(apiResp))
	for i, a := range apiResp {
		if a.DownloadURL != "" {
			fmt.Printf("ℹ provenance: downloading GitLab bundle %d/%d from URL...\n", i+1, len(apiResp))
			bundleData, err := v.downloadBundle(ctx, token, a.DownloadURL)
			if err != nil {
				logger.Warn("provenance: failed to download GitLab bundle", map[string]interface{}{
					"url":   a.DownloadURL,
					"error": err.Error(),
				})
				continue
			}
			raw = append(raw, bundleData)
		}
	}
	return raw, nil
}

func (v *gitlabProvenanceVerifier) downloadBundle(ctx context.Context, token, downloadURL string) (json.RawMessage, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "unirtm/"+env.GitTag)
	req.Header.Set("Accept", "application/json")
	if token != "" {
		req.Header.Set("PRIVATE-TOKEN", token)
	}

	resp, err := v.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download bundle returned status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
