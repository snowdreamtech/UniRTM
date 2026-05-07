// Package service provides business logic for UniRTM operations.
package service

import (
	"context"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"strings"
	"time"

	"github.com/snowdreamtech/unirtm/internal/pkg/logger"
	"github.com/snowdreamtech/unirtm/internal/repository"
)

// ChecksumAlgorithm represents a checksum algorithm.
type ChecksumAlgorithm string

const (
	AlgorithmSHA256 ChecksumAlgorithm = "sha256"
	AlgorithmSHA512 ChecksumAlgorithm = "sha512"
)

// SecurityVerificationResult represents the result of a security check.
type SecurityVerificationResult struct {
	Tool      string
	Version   string
	FilePath  string
	Algorithm ChecksumAlgorithm
	Expected  string
	Actual    string
	Passed    bool
	Warning   string
}

// SecurityManager handles checksum verification, signature checking, and
// security audit logging for installed tools.
//
// Validates Requirements: 20.3, 20.4, 20.5, 20.6, 20.7
type SecurityManager struct {
	auditRepo repository.AuditRepository
}

// NewSecurityManager creates a new SecurityManager.
func NewSecurityManager(auditRepo repository.AuditRepository) *SecurityManager {
	return &SecurityManager{
		auditRepo: auditRepo,
	}
}

// VerifyChecksum verifies the checksum of a downloaded file.
//
// Validates Requirements: 20.3 (checksum verification), 20.5 (store checksum in database)
func (sm *SecurityManager) VerifyChecksum(ctx context.Context, tool, version, filePath, expected string) (*SecurityVerificationResult, error) {
	if expected == "" {
		// No checksum provided — log a warning
		result := &SecurityVerificationResult{
			Tool:     tool,
			Version:  version,
			FilePath: filePath,
			Warning:  "no checksum provided for verification",
		}

		// Validates Requirement: 20.4 (warn for tools without checksum verification)
		logger.Warn("No checksum provided", map[string]interface{}{
			"tool":    tool,
			"version": version,
		})

		sm.logAudit(ctx, tool, version, "checksum_warning", "warning", result.Warning)
		return result, nil
	}

	// Determine algorithm from prefix or length
	algo, cleanExpected := parseChecksum(expected)

	// Compute actual checksum
	actual, err := computeChecksum(filePath, algo)
	if err != nil {
		sm.logAudit(ctx, tool, version, "checksum_error", "failure", err.Error())
		return nil, fmt.Errorf("compute checksum for %s: %w", filePath, err)
	}

	passed := strings.EqualFold(actual, cleanExpected)

	result := &SecurityVerificationResult{
		Tool:      tool,
		Version:   version,
		FilePath:  filePath,
		Algorithm: algo,
		Expected:  cleanExpected,
		Actual:    actual,
		Passed:    passed,
	}

	if passed {
		logger.Info("Checksum verified", map[string]interface{}{
			"tool":      tool,
			"version":   version,
			"algorithm": algo,
		})
		sm.logAudit(ctx, tool, version, "checksum_verify", "success", "")
	} else {
		errMsg := fmt.Sprintf("checksum mismatch: expected %s, got %s", cleanExpected, actual)
		// Validates Requirement: 20.7 (log security verification failures)
		logger.Error("Checksum mismatch", map[string]interface{}{
			"tool":     tool,
			"version":  version,
			"expected": cleanExpected,
			"actual":   actual,
		})
		sm.logAudit(ctx, tool, version, "checksum_verify", "failure", errMsg)
	}

	return result, nil
}

// VerifyInstallation verifies that an installed tool's checksum matches the stored value.
//
// Validates Requirement: 20.5 (checksum storage and re-verification)
func (sm *SecurityManager) VerifyInstallation(ctx context.Context, installation *repository.Installation, binaryPath string) (*SecurityVerificationResult, error) {
	if installation.Checksum == "" {
		result := &SecurityVerificationResult{
			Tool:     installation.Tool,
			Version:  installation.Version,
			FilePath: binaryPath,
			Warning:  "no stored checksum to verify against",
		}
		logger.Warn("No stored checksum for installation", map[string]interface{}{
			"tool":    installation.Tool,
			"version": installation.Version,
		})
		return result, nil
	}

	return sm.VerifyChecksum(ctx, installation.Tool, installation.Version, binaryPath, installation.Checksum)
}

// logAudit records a security event to the audit log.
//
// Validates Requirement: 20.7 (log security verification failures)
func (sm *SecurityManager) logAudit(ctx context.Context, tool, version, operation, status, errMsg string) {
	if sm.auditRepo == nil {
		return
	}
	_ = sm.auditRepo.Log(ctx, &repository.AuditEntry{
		Timestamp: time.Now(),
		Operation: "security_" + operation,
		Tool:      tool,
		Version:   version,
		Status:    status,
		Error:     errMsg,
	})
}

// parseChecksum detects the algorithm from a checksum string.
// Supports "sha256:<hex>", "sha512:<hex>", and bare hex strings.
func parseChecksum(checksum string) (ChecksumAlgorithm, string) {
	parts := strings.SplitN(checksum, ":", 2)
	if len(parts) == 2 {
		switch strings.ToLower(parts[0]) {
		case "sha256":
			return AlgorithmSHA256, parts[1]
		case "sha512":
			return AlgorithmSHA512, parts[1]
		}
	}
	// Infer from length: SHA-256 = 64 hex chars, SHA-512 = 128 hex chars
	if len(checksum) == 128 {
		return AlgorithmSHA512, checksum
	}
	return AlgorithmSHA256, checksum
}

// computeChecksum computes the checksum of a file.
func computeChecksum(filePath string, algo ChecksumAlgorithm) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	var h hash.Hash
	switch algo {
	case AlgorithmSHA512:
		h = sha512.New()
	default:
		h = sha256.New()
	}

	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("hash file: %w", err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
