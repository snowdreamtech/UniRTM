// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/snowdreamtech/unirtm/internal/backend"
	"github.com/snowdreamtech/unirtm/internal/pkg/download"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/pkg/logger"
	"github.com/snowdreamtech/unirtm/internal/provider"
	"github.com/snowdreamtech/unirtm/internal/repository"
	"github.com/snowdreamtech/unirtm/internal/transaction"
)

// InstallationManager manages tool installation workflow.
type InstallationManager struct {
	backendRegistry  *backend.Registry
	providerRegistry *provider.Registry
	downloadManager  *download.Manager
	installRepo      repository.InstallationRepository
	txManager        transaction.TransactionManager
	lockService      *LockService // optional; nil = lockfile disabled
}

// NewInstallationManager creates a new installation manager without lockfile support.
func NewInstallationManager(
	backendRegistry *backend.Registry,
	providerRegistry *provider.Registry,
	downloadManager *download.Manager,
	installRepo repository.InstallationRepository,
	txManager transaction.TransactionManager,
) *InstallationManager {
	return &InstallationManager{
		backendRegistry:  backendRegistry,
		providerRegistry: providerRegistry,
		downloadManager:  downloadManager,
		installRepo:      installRepo,
		txManager:        txManager,
	}
}

// NewInstallationManagerWithLock creates an InstallationManager that reads and
// writes unirtm.lock for reproducible, API-call-free installations.
func NewInstallationManagerWithLock(
	backendRegistry *backend.Registry,
	providerRegistry *provider.Registry,
	downloadManager *download.Manager,
	installRepo repository.InstallationRepository,
	txManager transaction.TransactionManager,
	lockService *LockService,
) *InstallationManager {
	im := NewInstallationManager(backendRegistry, providerRegistry, downloadManager, installRepo, txManager)
	im.lockService = lockService
	return im
}

// Install performs the complete installation workflow for a tool.
// Workflow: check → download → verify → extract → activate → record
func (im *InstallationManager) Install(ctx context.Context, tool, version, backendName string) error {
	// Start transaction
	tx, err := im.txManager.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Check if already installed
	existing, err := tx.InstallationRepo().FindByToolAndVersion(ctx, tool, version)
	if err == nil && existing != nil {
		// Verify if the installation directory actually exists on disk
		if _, statErr := os.Stat(existing.InstallPath); os.IsNotExist(statErr) {
			// Path doesn't exist, this is a stale database record. Clean it up.
			if delErr := tx.InstallationRepo().Delete(ctx, tool, version); delErr != nil {
				return fmt.Errorf("failed to clean up stale installation record: %w", delErr)
			}
		} else {
			return fmt.Errorf("tool %s version %s already installed", tool, version)
		}
	}

	// Get backend
	b, err := im.backendRegistry.Get(backendName)
	if err != nil {
		return fmt.Errorf("backend not found: %w", err)
	}

	// Get download info — check lockfile first to avoid remote API calls.
	platform := backend.CurrentPlatform()
	var versionInfo *backend.VersionInfo

	if im.lockService != nil {
		// Enforce strict mode before any API call.
		if err := im.lockService.CheckStrict(tool, version, platform); err != nil {
			return err
		}
		// Try to resolve directly from the lockfile.
		if info, ok := im.lockService.Resolve(tool, version, platform); ok {
			logger.Debug("lockfile hit: using cached URL", map[string]interface{}{
				"tool":     tool,
				"version":  version,
				"platform": platform.String(),
			})
			versionInfo = info
		}
	}

	if versionInfo == nil {
		// Lockfile miss — fall back to the remote backend.
		fmt.Printf("ℹ resolving download info for %s@%s...\n", tool, version)
		info, err := b.ResolveVersion(ctx, tool, version, platform)
		if err != nil {
			return fmt.Errorf("failed to resolve version: %w", err)
		}
		version = info.Version // Update to the concrete resolved version
		fmt.Printf("✓ resolved %s to version %s\n", tool, version)
		versionInfo = info
	}

	// Download artifact if URL is provided
	var downloadPath string
	var gpgStatus string = "NotRequested"
	if versionInfo.DownloadURL != "" {
		downloadPath = filepath.Join(env.GetDownloadsDir(), fmt.Sprintf("%s-%s", tool, version))
		fmt.Printf("ℹ downloading %s@%s to %s...\n", tool, version, downloadPath)
		if err := os.MkdirAll(filepath.Dir(downloadPath), 0755); err != nil {
			return fmt.Errorf("failed to create downloads directory: %w", err)
		}
		downloader, err := im.downloadManager.Get("https")
		if err != nil {
			return fmt.Errorf("failed to get downloader: %w", err)
		}

		opts := download.DefaultDownloadOptions()
		if versionInfo.Checksum != "" {
			opts = opts.WithChecksum(versionInfo.Checksum)
		}
		gpgResult := &download.GPGResult{}
		// Only attempt GPG verification if the keyring exists
		keyringPath := filepath.Join(env.GetDataDir(), "keyring.gpg")
		verifyGPG := false
		if _, err := os.Stat(keyringPath); err == nil {
			verifyGPG = true
		}
		opts = opts.WithVerifyGPG(verifyGPG, gpgResult)

		fmt.Printf("ℹ downloading %s@%s...\n", tool, version)
		if err := downloader.Download(ctx, versionInfo.DownloadURL, downloadPath, opts); err != nil {
			return fmt.Errorf("failed to download: %w", err)
		}
		fmt.Printf("✓ downloaded to %s\n", downloadPath)
		defer func() {
			os.Remove(downloadPath)
			// Clean up empty parent directories up to the downloads root
			im.removeEmptyDirs(downloadPath, env.GetDownloadsDir())
		}()

		// Verify checksum
		if versionInfo.Checksum != "" {
			if err := downloader.VerifyChecksum(ctx, downloadPath, versionInfo.Checksum); err != nil {
				return fmt.Errorf("checksum verification failed: %w", err)
			}
		}
		gpgStatus = gpgResult.Status

		// Verify GitHub provenance (SLSA attestation) if the backend is github or ubi.
		// If the project does not publish attestations, this is a no-op.
		// If attestations exist, verification MUST pass to prevent supply chain attacks.
		if backendName == "github" || backendName == "ubi" {
			// Extract owner/repo from tool string (expected format: "owner/repo").
			provenanceStatus, provenanceErr := tryVerifyProvenance(ctx, tool, downloadPath)
			if provenanceErr != nil {
				return fmt.Errorf("github provenance verification failed: %w", provenanceErr)
			}
			// Store result in versionInfo metadata for audit logging below.
			if versionInfo.Metadata == nil {
				versionInfo.Metadata = make(map[string]string)
			}
			versionInfo.Metadata["provenance_status"] = provenanceStatus
		}

		// Write resolved info back into the lockfile so future installs can
		// skip the remote API call (lock hit) and use the cached URL directly.
		if im.lockService != nil {
			if lockErr := im.lockService.RecordInstall(tool, backendName, versionInfo); lockErr != nil {
				// Non-fatal: log but don't abort the install.
				logger.Warn("lockfile: failed to record install", map[string]interface{}{
					"tool":  tool,
					"error": lockErr.Error(),
				})
			}
		}
	}

	// Install using provider
	installPath := filepath.Join(env.GetInstallsDir(), tool, version)
	p := im.providerRegistry.GetWithBackend(tool, backendName)

	if err := p.Install(ctx, installPath, downloadPath, version); err != nil {
		os.RemoveAll(installPath)
		return fmt.Errorf("installation failed: %w", err)
	}

	if err := p.PostInstall(ctx, installPath, version); err != nil {
		os.RemoveAll(installPath)
		return fmt.Errorf("post-install failed: %w", err)
	}

	// Record installation
	installation := &repository.Installation{
		Tool:        tool,
		Version:     version,
		Backend:     backendName,
		InstallPath: installPath,
		Checksum:    versionInfo.Checksum,
	}

	if err := tx.InstallationRepo().Create(ctx, installation); err != nil {
		os.RemoveAll(installPath)
		return fmt.Errorf("failed to record installation: %w", err)
	}

	// Record audit entry
	auditMeta := map[string]string{"gpg": gpgStatus}
	if versionInfo.Metadata != nil {
		if ps, ok := versionInfo.Metadata["provenance_status"]; ok && ps != "" {
			auditMeta["provenance"] = ps
		}
	}
	auditMetaJSON, _ := json.Marshal(auditMeta)
	auditEntry := &repository.AuditEntry{
		Operation:       "install",
		Tool:            tool,
		Version:         version,
		Status:          "success",
		GpgVerification: gpgStatus,
		Metadata:        string(auditMetaJSON),
	}
	if err := tx.AuditRepo().Log(ctx, auditEntry); err != nil {
		// Log but don't fail
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		os.RemoveAll(installPath)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Uninstall removes a tool installation.
func (im *InstallationManager) Uninstall(ctx context.Context, tool, version string) error {
	installation, err := im.installRepo.FindByToolAndVersion(ctx, tool, version)
	if err != nil {
		return fmt.Errorf("installation not found: %w", err)
	}

	// Run provider uninstall
	p := im.providerRegistry.GetWithBackend(tool, installation.Backend)
	if err := p.Uninstall(ctx, installation.InstallPath, version); err != nil {
		return fmt.Errorf("provider uninstall failed: %w", err)
	}

	// Remove installation directory
	if err := os.RemoveAll(installation.InstallPath); err != nil {
		return fmt.Errorf("failed to remove installation directory: %w", err)
	}

	// Remove from database
	if err := im.installRepo.Delete(ctx, tool, version); err != nil {
		return fmt.Errorf("failed to remove installation record: %w", err)
	}

	// Remove from lockfile
	if im.lockService != nil {
		_ = im.lockService.RemoveTool(tool)
	}

	return nil
}

// tryVerifyProvenance runs GitHub provenance (SLSA attestation) verification
// for tools whose tool string is in "owner/repo" format.
//
// Returns a short status string for audit logging:
//   - "not_applicable" — tool format is not owner/repo (e.g. language runtimes)
//   - "not_supported"  — project published no attestations (safely skipped)
//   - "verified"       — attestation present and verified successfully
//
// An error is returned only when attestations exist but verification fails,
// which is treated as a hard security failure.
func tryVerifyProvenance(ctx context.Context, tool, artifactPath string) (string, error) {
	// Provenance only applies to GitHub-hosted tools in "owner/repo" format.
	parts := strings.SplitN(tool, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "not_applicable", nil
	}
	owner, repo := parts[0], parts[1]

	// Resolve a token (reuses the 6-tier resolver from the GitHub backend).
	token := backend.ResolveGitHubTokenPublic("github.com")

	result, err := backend.VerifyArtifactProvenance(ctx, token, owner, repo, artifactPath)
	if err != nil {
		return "failed", err
	}

	if !result.Supported {
		logger.Debug("GitHub provenance not supported, skipping", map[string]interface{}{"tool": tool})
		return "not_supported", nil
	}

	logger.Debug("GitHub provenance verified", map[string]interface{}{
		"tool":          tool,
		"repository":    result.Repository,
		"workflowRef":   result.WorkflowRef,
		"predicateType": result.PredicateType,
	})
	return "verified", nil
}

// removeEmptyDirs recursively removes empty parent directories of path up to root.
func (im *InstallationManager) removeEmptyDirs(path string, root string) {
	dir := filepath.Dir(path)
	for {
		if dir == root || dir == "." || dir == filepath.Dir(root) {
			break
		}
		// Try to remove the directory. os.Remove only removes if empty.
		if err := os.Remove(dir); err != nil {
			break
		}
		dir = filepath.Dir(dir)
	}
}
