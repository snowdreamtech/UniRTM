// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sigstore/sigstore-go/pkg/bundle"
	"github.com/sigstore/sigstore-go/pkg/root"
	"github.com/sigstore/sigstore-go/pkg/tuf"
	"github.com/sigstore/sigstore-go/pkg/verify"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
	"github.com/snowdreamtech/unirtm/internal/pkg/logger"
	"github.com/theupdateframework/go-tuf/v2/metadata"
	"github.com/theupdateframework/go-tuf/v2/metadata/fetcher"
)

// ProvenanceResult represents the metadata extracted from a verified provenance bundle.
type ProvenanceResult struct {
	// Supported is false when the project publishes no attestations.
	Supported bool
	// Verified is true when the attestation bundle passed all checks.
	Verified bool
	// Repository is the source repository recorded in the Fulcio cert SAN.
	Repository string
	// WorkflowRef is the triggering workflow path recorded in the cert.
	WorkflowRef string
	// PredicateType is the in-toto predicate URI in the signed statement.
	PredicateType string
	// BuilderID is the SLSA builder identifier from the certificate extension.
	BuilderID string
}

// SigstoreVerifier acts as the common orchestration engine for verifying 
// Sigstore bundles across different platforms (GitHub, GitLab, etc.)
type SigstoreVerifier struct {
	TrustedMaterial root.TrustedMaterial
	Identities      []verify.CertificateIdentity
	ExpectedRepo    string
}

// VerifyBundles attempts to verify a list of bundles against the provided identities.
// It returns the first successful verification result.
func (v *SigstoreVerifier) VerifyBundles(bundles []json.RawMessage, artifactDigest string) (*ProvenanceResult, error) {
	var lastErr error
	for _, rawBundle := range bundles {
		for _, identity := range v.Identities {
			result, err := v.verifyBundle(rawBundle, artifactDigest, identity)
			if err == nil {
				return result, nil
			}
			lastErr = err
		}
	}

	return nil, fmt.Errorf(
		"provenance: all %d attestation bundle(s) failed verification for %s. Last error: %v",
		len(bundles), v.ExpectedRepo, lastErr,
	)
}

func (v *SigstoreVerifier) verifyBundle(rawBundle json.RawMessage, artifactDigest string, identity verify.CertificateIdentity) (*ProvenanceResult, error) {
	b := &bundle.Bundle{}
	if err := b.UnmarshalJSON(rawBundle); err != nil {
		return nil, fmt.Errorf("parse bundle JSON: %w", err)
	}

	tlogThreshold := 1
	if env.Get("SKIP_REKOR_VERIFY") == "1" {
		tlogThreshold = 0
	}

	verifierOpts := []verify.VerifierOption{
		verify.WithObserverTimestamps(1),
	}
	if tlogThreshold > 0 {
		verifierOpts = append(verifierOpts, verify.WithTransparencyLog(tlogThreshold))
	}

	verifier, err := verify.NewSignedEntityVerifier(v.TrustedMaterial, verifierOpts...)
	if err != nil {
		return nil, fmt.Errorf("build verifier: %w", err)
	}

	digestBytes, err := hex.DecodeString(artifactDigest)
	if err != nil {
		return nil, fmt.Errorf("decode artifact digest: %w", err)
	}

	policy := verify.NewPolicy(
		verify.WithArtifactDigest("sha256", digestBytes),
		verify.WithCertificateIdentity(identity),
	)

	logger.Debug("provenance: performing sigstore-go verification", map[string]interface{}{
		"tlogThreshold": tlogThreshold,
		"expectedRepo":  v.ExpectedRepo,
	})
	
	result, err := verifier.Verify(b, policy)
	if err != nil {
		logger.Debug("provenance: sigstore-go verification failed", map[string]interface{}{"error": err.Error()})
		return nil, fmt.Errorf("sigstore verification: %w", err)
	}

	provResult := &ProvenanceResult{
		Supported:  true,
		Verified:   true,
		Repository: v.ExpectedRepo,
	}

	if result.Signature != nil && result.Signature.Certificate != nil {
		cert := result.Signature.Certificate
		provResult.WorkflowRef = cert.SubjectAlternativeName
		provResult.BuilderID = cert.Extensions.BuildSignerURI

		// Post-verification spoofing check: Ensure expected repository is in OIDC extensions
		// Specifically critical for global identities (e.g. dotcom.releases.github.com)
		hasExpectedRepo := false
		if strings.Contains(cert.SubjectAlternativeName, v.ExpectedRepo) {
			hasExpectedRepo = true
		} else if strings.Contains(cert.Extensions.SourceRepositoryURI, v.ExpectedRepo) {
			hasExpectedRepo = true
		} else if strings.Contains(cert.Extensions.BuildSignerURI, v.ExpectedRepo) {
			hasExpectedRepo = true
		}
		
		if !hasExpectedRepo {
			return nil, fmt.Errorf("spoofed identity: expected repository %s not found in OIDC certificate", v.ExpectedRepo)
		}
	}

	if result.Statement != nil {
		provResult.PredicateType = result.Statement.GetPredicateType()
	}

	logger.Info("✓ provenance: verified signature and identity")
	return provResult, nil
}

// -----------------------------------------------------------------------------
// Common Utilities & TUF Root
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

// tufFetcher implements fetcher.Fetcher using our standard HTTP client.
type tufFetcher struct {
	client            *http.Client
	repositoryBaseURL string
}

var _ fetcher.Fetcher = (*tufFetcher)(nil)

func (f *tufFetcher) DownloadFile(urlPath string, maxLength int64, _ time.Duration) ([]byte, error) {
	finalURL := urlPath
	if !strings.HasPrefix(urlPath, "http") {
		baseURL := f.repositoryBaseURL
		if baseURL == "" {
			baseURL = "https://tuf-repo-cdn.sigstore.dev"
		}
		finalURL = strings.TrimSuffix(baseURL, "/") + "/" + strings.TrimPrefix(urlPath, "/")
	}

	logger.Debug("provenance: TUF fetching", map[string]interface{}{"url": finalURL, "max_length": maxLength})

	req, err := http.NewRequest(http.MethodGet, finalURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "unirtm/"+env.GitTag)

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Debug("provenance: TUF fetch returned non-200", map[string]interface{}{
			"url":    finalURL,
			"status": resp.StatusCode,
		})
		return nil, &metadata.ErrDownloadHTTP{StatusCode: resp.StatusCode, URL: finalURL}
	}

	return io.ReadAll(io.LimitReader(resp.Body, maxLength))
}

// InitializeTUFRoot sets up a Sigstore TUF root with the specified parameters.
func InitializeTUFRoot(cacheDirName, repoBaseURL string, customRoot []byte) (*root.LiveTrustedRoot, error) {
	opts := tuf.DefaultOptions()
	
	if repoBaseURL != "" {
		opts.RepositoryBaseURL = repoBaseURL
	}
	if customRoot != nil {
		opts.Root = customRoot
	}

	opts.Fetcher = &tufFetcher{
		client:            pkgHttp.NewClientWithTimeout(60 * time.Second),
		repositoryBaseURL: opts.RepositoryBaseURL,
	}

	cachePath := opts.CachePath
	if cacheDir := env.Get("TUF_CACHE_DIR"); cacheDir != "" {
		cachePath = filepath.Join(cacheDir, cacheDirName)
	} else {
		cachePath = filepath.Join(filepath.Dir(cachePath), cacheDirName)
	}
	opts.CachePath = cachePath

	logger.Debug("provenance: TUF cache path configured", map[string]interface{}{"path": cachePath})
	return root.NewLiveTrustedRoot(opts)
}
