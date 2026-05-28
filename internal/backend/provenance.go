// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

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
	TrustedMaterials []root.TrustedMaterial
	Identities       []verify.CertificateIdentity
	ExpectedRepo     string
}

// VerifyBundles attempts to verify a list of bundles against the provided identities and trust roots.
// It returns the first successful verification result.
func (v *SigstoreVerifier) VerifyBundles(bundles []json.RawMessage, artifactDigest string) (*ProvenanceResult, error) {
	var errs []string
	for _, rawBundle := range bundles {
		for i, tm := range v.TrustedMaterials {
			for j, identity := range v.Identities {
				result, err := v.verifyBundle(rawBundle, artifactDigest, identity, tm)
				if err == nil {
					return result, nil
				}
				errs = append(errs, fmt.Sprintf("[Root%d/Id%d] %v", i, j, err))
			}
		}
	}

	return nil, fmt.Errorf(
		"provenance: all %d attestation bundle(s) failed verification for %s. Errors: %s",
		len(bundles), v.ExpectedRepo, strings.Join(errs, "; "),
	)
}

func (v *SigstoreVerifier) verifyBundle(rawBundle json.RawMessage, artifactDigest string, identity verify.CertificateIdentity, tm root.TrustedMaterial) (*ProvenanceResult, error) {
	b := &bundle.Bundle{}
	if err := b.UnmarshalJSON(rawBundle); err != nil {
		return nil, fmt.Errorf("parse bundle JSON: %w", err)
	}

	tlogThreshold := 1
	if env.Get("SKIP_REKOR_VERIFY") == "1" {
		tlogThreshold = 0
	} else {
		// Inspect the bundle to see if it actually contains transparency log entries.
		// GitHub Attestations, for example, do not contain tlog entries and only use RFC 3161 timestamps.
		if entries, err := b.TlogEntries(); err != nil || len(entries) == 0 {
			tlogThreshold = 0
		}
	}

	var optionSets [][]verify.VerifierOption
	if tlogThreshold > 0 {
		// Option Set 1: With Observer Timestamps + Transparency Log
		optionSets = append(optionSets, []verify.VerifierOption{
			verify.WithObserverTimestamps(1),
			verify.WithTransparencyLog(tlogThreshold),
		})
		// Option Set 2: With Integrated Timestamps + Transparency Log
		optionSets = append(optionSets, []verify.VerifierOption{
			verify.WithIntegratedTimestamps(1),
			verify.WithTransparencyLog(tlogThreshold),
		})
	} else {
		// No Rekor log: We can only verify using Observer Timestamps (like RFC 3161 timestamps)
		optionSets = append(optionSets, []verify.VerifierOption{
			verify.WithObserverTimestamps(1),
		})
	}

	var result *verify.VerificationResult
	var lastErr error

	for i, opts := range optionSets {
		verifier, err := verify.NewSignedEntityVerifier(tm, opts...)
		if err != nil {
			lastErr = fmt.Errorf("build verifier: %w", err)
			continue
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
			"attempt":       i + 1,
			"tlogThreshold": tlogThreshold,
			"expectedRepo":  v.ExpectedRepo,
		})

		res, err := verifier.Verify(b, policy)
		if err == nil {
			result = res
			break
		}
		logger.Debug("provenance: sigstore-go verification attempt failed", map[string]interface{}{
			"attempt": i + 1,
			"error":   err.Error(),
		})
		lastErr = fmt.Errorf("sigstore verification: %w", err)
	}

	if result == nil {
		return nil, lastErr
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
		} else if result.Statement != nil {
			// Check if the verified in-toto statement contains the expected repository.
			// This is essential for global builders (like dotcom.releases.github.com)
			// where repository identity is encoded in the signed release predicate.
			stmtJSON, err := json.Marshal(result.Statement)
			if err == nil && strings.Contains(string(stmtJSON), v.ExpectedRepo) {
				hasExpectedRepo = true
			}
		}

		if !hasExpectedRepo {
			return nil, fmt.Errorf("spoofed identity: expected repository %s not found in OIDC certificate (SAN=%s SourceRepositoryURI=%s) or signed statement", v.ExpectedRepo, cert.SubjectAlternativeName, cert.Extensions.SourceRepositoryURI)
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

func (f *tufFetcher) DownloadFile(urlPath string, maxLength int64, timeout time.Duration) ([]byte, error) {
	finalURL := urlPath
	if !strings.HasPrefix(urlPath, "http") {
		baseURL := f.repositoryBaseURL
		if baseURL == "" {
			baseURL = "https://tuf-repo-cdn.sigstore.dev"
		}
		finalURL = strings.TrimSuffix(baseURL, "/") + "/" + strings.TrimPrefix(urlPath, "/")
	}

	logger.Debug("provenance: TUF fetching", map[string]interface{}{"url": finalURL, "max_length": maxLength})

	ctx := context.Background()
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, finalURL, nil)
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
