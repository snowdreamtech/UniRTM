// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/snowdreamtech/unirtm/internal/backend"
	"github.com/snowdreamtech/unirtm/internal/pkg/download"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
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
}

// NewInstallationManager creates a new installation manager.
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

	// Get download info
	platform := backend.CurrentPlatform()
	versionInfo, err := b.GetDownloadInfo(ctx, tool, version, platform)
	if err != nil {
		return fmt.Errorf("failed to get download info: %w", err)
	}

	// Download artifact if URL is provided
	var downloadPath string
	var gpgStatus string = "NotRequested"
	if versionInfo.DownloadURL != "" {
		downloadPath = filepath.Join(env.GetDownloadsDir(), fmt.Sprintf("%s-%s", tool, version))
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
		opts = opts.WithVerifyGPG(true, gpgResult)

		if err := downloader.Download(ctx, versionInfo.DownloadURL, downloadPath, opts); err != nil {
			return fmt.Errorf("failed to download: %w", err)
		}
		defer os.Remove(downloadPath)

		// Verify checksum
		if versionInfo.Checksum != "" {
			if err := downloader.VerifyChecksum(ctx, downloadPath, versionInfo.Checksum); err != nil {
				return fmt.Errorf("checksum verification failed: %w", err)
			}
		}
		gpgStatus = gpgResult.Status
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
	auditEntry := &repository.AuditEntry{
		Operation:       "install",
		Tool:            tool,
		Version:         version,
		Status:          "success",
		GpgVerification: gpgStatus,
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

	return nil
}
