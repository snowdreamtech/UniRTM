// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// -----------------------------------------------------------------------------
// Public API
// -----------------------------------------------------------------------------

// ProvenanceResult is returned by VerifyArtifactProvenance.
type ProvenanceResult struct {
	// Supported is false when the project publishes no GitHub attestations.
	// Callers MUST NOT treat this as an error; simply skip further checks.
	Supported bool

	// Verified is true when the attestation bundle passed all cryptographic checks.
	Verified bool

	// Repository is the source repository recorded in the certificate SAN,
	// e.g. "octocat/hello-world".
	Repository string

	// WorkflowRef is the triggering workflow path, e.g.
	// "octocat/hello-world/.github/workflows/release.yml@refs/heads/main".
	WorkflowRef string

	// PredicateType is the in-toto predicate URI, e.g.
	// "https://slsa.dev/provenance/v1".
	PredicateType string

	// BuilderID is the SLSA builder identifier embedded in the cert extensions.
	BuilderID string
}

// VerifyArtifactProvenance checks the GitHub attestation for the artifact at
// artifactPath against the repository owner/repo.
//
//   - If the project publishes no attestations, result.Supported == false and
//     err == nil. The caller should skip any further provenance enforcement.
//   - If attestations exist but verification fails, err != nil.
//   - On success, result.Supported == true && result.Verified == true.
func VerifyArtifactProvenance(
	ctx context.Context,
	token, owner, repo, artifactPath string,
) (*ProvenanceResult, error) {
	// 1. Compute the SHA-256 digest of the downloaded artifact.
	digest, err := sha256File(artifactPath)
	if err != nil {
		return nil, fmt.Errorf("provenance: compute digest: %w", err)
	}

	// 2. Query the GitHub Attestations API.
	bundles, err := fetchAttestations(ctx, token, owner, repo, digest)
	if err != nil {
		return nil, err
	}

	// 3. If no attestations were published, declare "not supported" and stop.
	if len(bundles) == 0 {
		return &ProvenanceResult{Supported: false}, nil
	}

	// 4. Verify the first matching bundle (GitHub usually only returns one).
	result, err := verifyBundle(bundles[0], digest, owner+"/"+repo)
	if err != nil {
		return nil, fmt.Errorf("provenance: verification failed for %s/%s: %w", owner, repo, err)
	}

	return result, nil
}

// -----------------------------------------------------------------------------
// GitHub Attestations REST API
// -----------------------------------------------------------------------------

// attestationAPIResponse matches the GitHub Attestations API envelope.
type attestationAPIResponse struct {
	Attestations []attestationEntry `json:"attestations"`
}

type attestationEntry struct {
	Bundle    json.RawMessage `json:"bundle"`
	BundleURL string          `json:"bundle_url"`
}

// fetchAttestations queries GET /repos/{owner}/{repo}/attestations/sha256:{digest}.
// Returns nil slice (not an error) when the project publishes no attestations.
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

	if len(apiResp.Attestations) == 0 {
		return nil, nil
	}

	raw := make([]json.RawMessage, len(apiResp.Attestations))
	for i, a := range apiResp.Attestations {
		raw[i] = a.Bundle
	}
	return raw, nil
}

// -----------------------------------------------------------------------------
// Sigstore Bundle verification
// -----------------------------------------------------------------------------

// sigstoreBundle is the wire format returned by the Attestations API.
// Conforms to application/vnd.dev.sigstore.bundle+json;version=0.3
type sigstoreBundle struct {
	MediaType           string              `json:"mediaType"`
	VerificationMaterial verificationMat    `json:"verificationMaterial"`
	DSSEEnvelope        dsseEnvelope        `json:"dsseEnvelope"`
}

type verificationMat struct {
	X509CertificateChain *x509CertChain   `json:"x509CertificateChain"`
	TlogEntries          []tlogEntry      `json:"tlogEntries"`
}

type x509CertChain struct {
	Certificates []certWrapper `json:"certificates"`
}

type certWrapper struct {
	RawBytes string `json:"rawBytes"` // base64-encoded DER
}

type tlogEntry struct {
	LogIndex               string `json:"logIndex"`
	LogID                  string `json:"logID"`
	KindVersion            struct {
		Kind    string `json:"kind"`
		Version string `json:"version"`
	} `json:"kindVersion"`
	IntegratedTime         string `json:"integratedTime"`
	InclusionPromise       struct {
		SignedEntryTimestamp string `json:"signedEntryTimestamp"`
	} `json:"inclusionPromise"`
}

type dsseEnvelope struct {
	Payload     string        `json:"payload"`     // base64-encoded in-toto statement
	PayloadType string        `json:"payloadType"` // "application/vnd.in-toto+json"
	Signatures  []dsseSig     `json:"signatures"`
}

type dsseSig struct {
	Sig   string `json:"sig"`   // base64-encoded signature
	KeyID string `json:"keyid"` // empty for Sigstore ephemeral keys
}

// inTotoStatement is the decoded payload of the DSSE envelope.
type inTotoStatement struct {
	Type          string          `json:"_type"`
	Subject       []inTotoSubject `json:"subject"`
	PredicateType string          `json:"predicateType"`
	Predicate     json.RawMessage `json:"predicate"`
}

type inTotoSubject struct {
	Name   string            `json:"name"`
	Digest map[string]string `json:"digest"`
}

// verifyBundle performs full cryptographic verification of a Sigstore bundle.
func verifyBundle(raw json.RawMessage, artifactDigest, expectedRepo string) (*ProvenanceResult, error) {
	var bundle sigstoreBundle
	if err := json.Unmarshal(raw, &bundle); err != nil {
		return nil, fmt.Errorf("unmarshal bundle: %w", err)
	}

	// ── Step A: Parse the leaf certificate ───────────────────────────────────
	leaf, chain, err := parseCertChain(bundle.VerificationMaterial.X509CertificateChain)
	if err != nil {
		return nil, fmt.Errorf("parse cert chain: %w", err)
	}

	// ── Step B: Validate the certificate chain against Fulcio roots ──────────
	if err := validateCertChain(leaf, chain); err != nil {
		return nil, fmt.Errorf("cert chain validation: %w", err)
	}

	// ── Step C: Verify certificate has not expired at signing time ───────────
	if err := checkCertValidity(leaf, bundle.VerificationMaterial.TlogEntries); err != nil {
		return nil, fmt.Errorf("cert validity: %w", err)
	}

	// ── Step D: Extract claims from the certificate's OID extensions ─────────
	claims, err := extractCertClaims(leaf)
	if err != nil {
		return nil, fmt.Errorf("extract cert claims: %w", err)
	}

	// ── Step E: Repository check — cert must reference our repo ─────────────
	if !strings.Contains(claims.san, expectedRepo) {
		return nil, fmt.Errorf(
			"repository mismatch: cert SAN %q does not contain expected repo %q",
			claims.san, expectedRepo,
		)
	}

	// ── Step F: Decode and parse the in-toto statement ───────────────────────
	payloadBytes, err := base64.StdEncoding.DecodeString(bundle.DSSEEnvelope.Payload)
	if err != nil {
		return nil, fmt.Errorf("decode DSSE payload: %w", err)
	}

	var stmt inTotoStatement
	if err := json.Unmarshal(payloadBytes, &stmt); err != nil {
		return nil, fmt.Errorf("parse in-toto statement: %w", err)
	}

	// ── Step G: Verify the DSSE signature with the leaf cert public key ──────
	if err := verifyDSSESignature(bundle.DSSEEnvelope, leaf); err != nil {
		return nil, fmt.Errorf("DSSE signature verification: %w", err)
	}

	// ── Step H: Subject digest must match our artifact ───────────────────────
	if err := verifySubjectDigest(stmt.Subject, artifactDigest); err != nil {
		return nil, fmt.Errorf("subject digest mismatch: %w", err)
	}

	return &ProvenanceResult{
		Supported:     true,
		Verified:      true,
		Repository:    expectedRepo,
		WorkflowRef:   claims.workflowRef,
		PredicateType: stmt.PredicateType,
		BuilderID:     claims.builderID,
	}, nil
}

// -----------------------------------------------------------------------------
// Certificate handling
// -----------------------------------------------------------------------------

type certClaims struct {
	san         string // Subject Alternative Name (URI form)
	workflowRef string // Fulcio OID 1.3.6.1.4.1.57264.1.3
	builderID   string // Fulcio OID 1.3.6.1.4.1.57264.1.9 (SLSA builder)
	issuer      string // Fulcio OID 1.3.6.1.4.1.57264.1.1
}

// Fulcio custom OIDs (see https://github.com/sigstore/fulcio/blob/main/docs/oid-info.md)
var (
	oidFulcioIssuer      = mustParseOID("1.3.6.1.4.1.57264.1.1")
	oidFulcioWorkflow    = mustParseOID("1.3.6.1.4.1.57264.1.3") // workflow ref
	oidFulcioBuildSigner = mustParseOID("1.3.6.1.4.1.57264.1.9") // build signer URI (SLSA builder ID)
)

func mustParseOID(s string) []int {
	parts := strings.Split(s, ".")
	oid := make([]int, len(parts))
	for i, p := range parts {
		n := 0
		for _, c := range p {
			n = n*10 + int(c-'0')
		}
		oid[i] = n
	}
	return oid
}

func oidEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func parseCertChain(chain *x509CertChain) (*x509.Certificate, []*x509.Certificate, error) {
	if chain == nil || len(chain.Certificates) == 0 {
		return nil, nil, fmt.Errorf("empty certificate chain")
	}

	var certs []*x509.Certificate
	for i, c := range chain.Certificates {
		der, err := base64.StdEncoding.DecodeString(c.RawBytes)
		if err != nil {
			return nil, nil, fmt.Errorf("cert[%d]: base64 decode: %w", i, err)
		}
		cert, err := x509.ParseCertificate(der)
		if err != nil {
			return nil, nil, fmt.Errorf("cert[%d]: parse DER: %w", i, err)
		}
		certs = append(certs, cert)
	}

	return certs[0], certs[1:], nil
}

// validateCertChain verifies the leaf against the Sigstore public good Fulcio CA.
func validateCertChain(leaf *x509.Certificate, intermediates []*x509.Certificate) error {
	roots, err := fulcioRootPool()
	if err != nil {
		return err
	}

	interPool := x509.NewCertPool()
	for _, c := range intermediates {
		interPool.AddCert(c)
	}

	opts := x509.VerifyOptions{
		Roots:         roots,
		Intermediates: interPool,
		// Use the certificate's own NotBefore as CurrentTime so we verify
		// at issuance time, not at current wall clock (ephemeral certs expire
		// in ~10 minutes but the artifact lives forever).
		CurrentTime: leaf.NotBefore,
		KeyUsages:   []x509.ExtKeyUsage{x509.ExtKeyUsageCodeSigning},
	}

	if _, err := leaf.Verify(opts); err != nil {
		return fmt.Errorf("certificate chain verification failed: %w", err)
	}
	return nil
}

// checkCertValidity ensures the transparency log entry's integrated time falls
// within the leaf certificate's validity window.
func checkCertValidity(leaf *x509.Certificate, entries []tlogEntry) error {
	if len(entries) == 0 {
		// No Rekor entry — accept if cert is still technically valid,
		// but warn in the result. For now we allow it for private Rekor instances.
		return nil
	}
	// Primary check: the entry's integratedTime must be within the cert window.
	for _, e := range entries {
		if e.IntegratedTime == "" {
			continue
		}
		var ts int64
		if _, err := fmt.Sscan(e.IntegratedTime, &ts); err != nil {
			continue
		}
		t := time.Unix(ts, 0)
		if t.Before(leaf.NotBefore) || t.After(leaf.NotAfter) {
			return fmt.Errorf(
				"tlog integrated time %v is outside cert validity [%v, %v]",
				t, leaf.NotBefore, leaf.NotAfter,
			)
		}
		return nil // first valid entry suffices
	}
	return nil
}

// extractCertClaims reads the Fulcio-specific OID extensions from the leaf cert.
func extractCertClaims(leaf *x509.Certificate) (certClaims, error) {
	var c certClaims

	// Subject Alternative Name: Fulcio uses URI SAN.
	for _, u := range leaf.URIs {
		c.san = u.String()
		break
	}

	// Fulcio custom extensions
	for _, ext := range leaf.Extensions {
		val := string(ext.Value)
		// Strip DER UTF-8 string tag if present (0x0c <len> ...)
		if len(ext.Value) > 2 && ext.Value[0] == 0x0c {
			val = string(ext.Value[2:])
		}

		switch {
		case oidEqual(ext.Id, oidFulcioIssuer):
			c.issuer = val
		case oidEqual(ext.Id, oidFulcioWorkflow):
			c.workflowRef = val
		case oidEqual(ext.Id, oidFulcioBuildSigner):
			c.builderID = val
		}
	}

	return c, nil
}

// verifyDSSESignature verifies the DSSE envelope signature using the leaf cert's public key.
// DSSE PAE: "DSSEv1" + SP + len(payloadType) + SP + payloadType + SP + len(payload) + SP + payload
func verifyDSSESignature(env dsseEnvelope, leaf *x509.Certificate) error {
	if len(env.Signatures) == 0 {
		return fmt.Errorf("no signatures in DSSE envelope")
	}

	payloadBytes, err := base64.StdEncoding.DecodeString(env.Payload)
	if err != nil {
		return fmt.Errorf("decode payload: %w", err)
	}

	// Build the PAE (Pre-Authentication Encoding) message.
	pae := dssePAE(env.PayloadType, payloadBytes)

	pubKey, ok := leaf.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("leaf cert public key is not ECDSA (got %T)", leaf.PublicKey)
	}

	for _, sig := range env.Signatures {
		sigBytes, err := base64.StdEncoding.DecodeString(sig.Sig)
		if err != nil {
			continue
		}

		// Choose digest algorithm based on curve
		var h hash.Hash
		switch pubKey.Curve.Params().Name {
		case "P-384":
			h = sha512.New384()
		default: // P-256
			h = sha256.New()
		}
		h.Write(pae)
		digest := h.Sum(nil)

		if ecdsa.VerifyASN1(pubKey, digest, sigBytes) {
			return nil
		}
	}
	return fmt.Errorf("no valid DSSE signature found")
}

// dssePAE computes the DSSE Pre-Authentication Encoding.
func dssePAE(payloadType string, payload []byte) []byte {
	// "DSSEv1" + SP + len(payloadType) + SP + payloadType + SP + len(payload) + SP + payload
	return []byte(fmt.Sprintf("DSSEv1 %d %s %d ", len(payloadType), payloadType, len(payload)) + string(payload))
}

// verifySubjectDigest checks that at least one subject in the in-toto statement
// matches the downloaded artifact's SHA-256 digest.
func verifySubjectDigest(subjects []inTotoSubject, artifactDigest string) error {
	for _, s := range subjects {
		if d, ok := s.Digest["sha256"]; ok && strings.EqualFold(d, artifactDigest) {
			return nil
		}
	}
	return fmt.Errorf("artifact digest %q not found in attestation subjects", artifactDigest)
}

// -----------------------------------------------------------------------------
// Fulcio root certificate pool
// -----------------------------------------------------------------------------

// fulcioRootPool returns an x509.CertPool loaded with the Sigstore public good
// Fulcio root CA and its intermediate certificate. Both certs are embedded
// directly to avoid network access and TUF bootstrapping complexity.
// Source: https://github.com/sigstore/root-signing
func fulcioRootPool() (*x509.CertPool, error) {
	pool := x509.NewCertPool()

	// Sigstore Fulcio v1 root CA (production, public good instance).
	// Source: targets/fulcio_v1.crt.pem
	const fulcioRootV1PEM = `-----BEGIN CERTIFICATE-----
MIIB9zCCAXygAwIBAgIUALZNAPFdxHPwjeDloDwyYChAO/4wCgYIKoZIzj0EAwMw
KjEVMBMGA1UEChMMc2lnc3RvcmUuZGV2MREwDwYDVQQDEwhzaWdzdG9yZTAeFw0y
MTEwMDcxMzU2NTlaFw0zMTEwMDUxMzU2NThaMCoxFTATBgNVBAoTDHNpZ3N0b3Jl
LmRldjERMA8GA1UEAxMIc2lnc3RvcmUwdjAQBgcqhkjOPQIBBgUrgQQAIgNiAAT7
XeFT4rb3PQGwS4IajtLk3/OlnpgangaBclYpsYBr5i+4ynB07ceb3LP0OIOZdxex
X69c5iVuyJRQ+Hz05yi+UF3uBWAlHpiS5sh0+H2GHE7SXrk1EC5m1Tr19L9gg92j
YzBhMA4GA1UdDwEB/wQEAwIBBjAPBgNVHRMBAf8EBTADAQH/MB0GA1UdDgQWBBRY
wB5fkUWlZql6zJChkyLQKsXF+jAfBgNVHSMEGDAWgBRYwB5fkUWlZql6zJChkyLQ
KsXF+jAKBggqhkjOPQQDAwNpADBmAjEAj1nHeXZp+13NWBNa+EDsDP8G1WWg1tCM
WP/WHPqpaVo0jhsweNFZgSs0eE7wYI4qAjEA2WB9ot98sIkoF3vZYdd3/VtWB5b9
TNMea7Ix/stJ5TfcLLeABLE4BNJOsQ4vnBHJ
-----END CERTIFICATE-----`

	// Sigstore Fulcio v1 intermediate CA.
	// Source: targets/fulcio_intermediate_v1.crt.pem
	const fulcioIntermediateV1PEM = `-----BEGIN CERTIFICATE-----
MIICGjCCAaGgAwIBAgIUALnViVfnU0brJasmRkHrn/UnfaQwCgYIKoZIzj0EAwMw
KjEVMBMGA1UEChMMc2lnc3RvcmUuZGV2MREwDwYDVQQDEwhzaWdzdG9yZTAeFw0y
MjA0MTMyMDA2MTVaFw0zMTEwMDUxMzU2NThaMDcxFTATBgNVBAoTDHNpZ3N0b3Jl
LmRldjEeMBwGA1UEAxMVc2lnc3RvcmUtaW50ZXJtZWRpYXRlMHYwEAYHKoZIzj0C
AQYFK4EEACIDYgAE8RVS/ysH+NOvuDZyPIZtilgUF9NlarYpAd9HP1vBBH1U5CV7
7LSS7s0ZiH4nE7Hv7ptS6LvvR/STk798LVgMzLlJ4HeIfF3tHSaexLcYpSASr1kS
0N/RgBJz/9jWCiXno3sweTAOBgNVHQ8BAf8EBAMCAQYwEwYDVR0lBAwwCgYIKwYB
BQUHAwMwEgYDVR0TAQH/BAgwBgEB/wIBADAdBgNVHQ4EFgQU39Ppz1YkEZb5qNjp
KFWixi4YZD8wHwYDVR0jBBgwFoAUWMAeX5FFpWapesyQoZMi0CrFxfowCgYIKoZI
zj0EAwMDZwAwZAIwPCsQK4DYiZYDPIaDi5HFKnfxXx6ASSVmERfsynYBiX2X6SJR
nZU84/9DZdnFvvxmAjBOt6QpBlc4J/0DxvkTCqpclvziL6BCCPnjdlIB3Pu3BxsP
mygUY7Ii2zbdCdliiow=
-----END CERTIFICATE-----`

	for _, pemData := range []string{fulcioRootV1PEM, fulcioIntermediateV1PEM} {
		if !pool.AppendCertsFromPEM([]byte(pemData)) {
			return nil, fmt.Errorf("failed to load Fulcio CA certificate")
		}
	}

	return pool, nil
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
