// Package service provides business logic for UniRTM operations.
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/snowdreamtech/unirtm/internal/backend"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/pkg/download"
	"github.com/snowdreamtech/unirtm/internal/provider"
	"github.com/snowdreamtech/unirtm/internal/repository"
	"github.com/snowdreamtech/unirtm/internal/transaction"
)

// UpdateInfo represents information about an available update for a tool.
type UpdateInfo struct {
	Tool           string // Tool name
	CurrentVersion string // Currently installed version
	LatestVersion  string // Latest available version
	Backend        string // Backend used for the tool
	UpdateRequired bool   // Whether an update is available
}

// UpdatePreview represents a preview of what will be updated.
type UpdatePreview struct {
	Updates       []UpdateInfo  // List of tools that will be updated
	TotalUpdates  int           // Total number of updates
	EstimatedTime time.Duration // Estimated time for all updates
}

// UpdateResult represents the result of an update operation.
type UpdateResult struct {
	Tool           string        // Tool name
	OldVersion     string        // Previous version
	NewVersion     string        // New version after update
	Success        bool          // Whether the update succeeded
	Error          string        // Error message if update failed
	Duration       time.Duration // Time taken for the update
	RolledBack     bool          // Whether a rollback was performed
	RollbackReason string        // Reason for rollback if applicable
}

// UpdateManager manages tool updates with version checking, preview, and rollback support.
//
// Validates Requirements: 25.1, 25.2, 25.3, 25.4, 25.5, 25.6, 25.7
type UpdateManager struct {
	backendRegistry  *backend.Registry
	providerRegistry *provider.Registry
	downloadManager  *download.Manager
	installRepo      repository.InstallationRepository
	auditRepo        repository.AuditRepository
	txManager        transaction.TransactionManager
	configManager    *config.Config
}

// NewUpdateManager creates a new update manager.
func NewUpdateManager(
	backendRegistry *backend.Registry,
	providerRegistry *provider.Registry,
	downloadManager *download.Manager,
	installRepo repository.InstallationRepository,
	auditRepo repository.AuditRepository,
	txManager transaction.TransactionManager,
	configManager *config.Config,
) *UpdateManager {
	return &UpdateManager{
		backendRegistry:  backendRegistry,
		providerRegistry: providerRegistry,
		downloadManager:  downloadManager,
		installRepo:      installRepo,
		auditRepo:        auditRepo,
		txManager:        txManager,
		configManager:    configManager,
	}
}

// CheckForUpdates checks for newer versions of installed tools.
//
// Validates Requirement 25.1: Check for newer versions of installed tools
func (um *UpdateManager) CheckForUpdates(ctx context.Context) ([]UpdateInfo, error) {
	// Get all installed tools
	installations, err := um.installRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list installations: %w", err)
	}

	var updates []UpdateInfo
	platform := backend.CurrentPlatform()

	for _, installation := range installations {
		// Get backend for this tool
		b, err := um.backendRegistry.Get(installation.Backend)
		if err != nil {
			// Skip tools with unavailable backends
			continue
		}

		// Resolve latest version
		latestInfo, err := b.ResolveVersion(ctx, installation.Tool, "latest", platform)
		if err != nil {
			// Skip tools where we can't determine latest version
			continue
		}

		// Check if update is available
		updateRequired := latestInfo.Version != installation.Version

		// Check version constraints from config if available
		if um.configManager != nil {
			if toolConfig, exists := um.configManager.Tools[installation.Tool]; exists {
				// If a specific version is pinned in config, respect it
				if toolConfig.Version != "" && toolConfig.Version != "latest" {
					// Try to resolve the configured version
					configInfo, err := b.ResolveVersion(ctx, installation.Tool, toolConfig.Version, platform)
					if err == nil {
						// Use the configured version as the target
						latestInfo = configInfo
						updateRequired = configInfo.Version != installation.Version
					}
				}
			}
		}

		updates = append(updates, UpdateInfo{
			Tool:           installation.Tool,
			CurrentVersion: installation.Version,
			LatestVersion:  latestInfo.Version,
			Backend:        installation.Backend,
			UpdateRequired: updateRequired,
		})
	}

	return updates, nil
}

// UpdateTool updates a specific tool to a specific version.
//
// Validates Requirement 25.2: Support updating a specific tool to a specific version
// Validates Requirement 25.7: Rollback to the previous version when an update fails
func (um *UpdateManager) UpdateTool(ctx context.Context, tool, targetVersion string) (*UpdateResult, error) {
	startTime := time.Now()

	// Get current installation
	installation, err := um.installRepo.FindByToolAndVersion(ctx, tool, "")
	if err != nil {
		return nil, fmt.Errorf("tool %s not installed: %w", tool, err)
	}

	oldVersion := installation.Version
	oldInstallPath := installation.InstallPath

	// Check if already at target version
	if oldVersion == targetVersion {
		return &UpdateResult{
			Tool:       tool,
			OldVersion: oldVersion,
			NewVersion: targetVersion,
			Success:    true,
			Duration:   time.Since(startTime),
		}, nil
	}

	// Start transaction for atomic update
	tx, err := um.txManager.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Audit log: start update
	auditEntry := &repository.AuditEntry{
		Timestamp: time.Now(),
		Operation: "update",
		Tool:      tool,
		Version:   targetVersion,
		Status:    "in_progress",
		Metadata:  fmt.Sprintf(`{"old_version":"%s","new_version":"%s"}`, oldVersion, targetVersion),
	}
	if err := tx.AuditRepo().Log(ctx, auditEntry); err != nil {
		return nil, fmt.Errorf("failed to log audit entry: %w", err)
	}

	// Get backend
	b, err := um.backendRegistry.Get(installation.Backend)
	if err != nil {
		return um.createFailureResult(tool, oldVersion, targetVersion, startTime, fmt.Errorf("backend not found: %w", err), false, "")
	}

	// Get download info for target version
	platform := backend.CurrentPlatform()
	versionInfo, err := b.GetDownloadInfo(ctx, tool, targetVersion, platform)
	if err != nil {
		return um.createFailureResult(tool, oldVersion, targetVersion, startTime, fmt.Errorf("failed to get download info: %w", err), false, "")
	}

	// Download new version
	downloadPath := filepath.Join(os.TempDir(), fmt.Sprintf("%s-%s", tool, targetVersion))
	downloader, err := um.downloadManager.Get("https")
	if err != nil {
		return um.createFailureResult(tool, oldVersion, targetVersion, startTime, fmt.Errorf("failed to get downloader: %w", err), false, "")
	}

	opts := download.DefaultDownloadOptions()
	if versionInfo.Checksum != "" {
		opts = opts.WithChecksum(versionInfo.Checksum)
	}

	if err := downloader.Download(ctx, versionInfo.DownloadURL, downloadPath, opts); err != nil {
		return um.createFailureResult(tool, oldVersion, targetVersion, startTime, fmt.Errorf("failed to download: %w", err), false, "")
	}
	defer os.Remove(downloadPath)

	// Verify checksum
	if versionInfo.Checksum != "" {
		if err := downloader.VerifyChecksum(ctx, downloadPath, versionInfo.Checksum); err != nil {
			return um.createFailureResult(tool, oldVersion, targetVersion, startTime, fmt.Errorf("checksum verification failed: %w", err), false, "")
		}
	}

	// Install new version
	newInstallPath := filepath.Join(filepath.Dir(oldInstallPath), targetVersion)
	p := um.providerRegistry.Get(tool)

	if err := p.Install(ctx, newInstallPath, downloadPath, targetVersion); err != nil {
		os.RemoveAll(newInstallPath)
		return um.createFailureResult(tool, oldVersion, targetVersion, startTime, fmt.Errorf("installation failed: %w", err), false, "")
	}

	if err := p.PostInstall(ctx, newInstallPath, targetVersion); err != nil {
		os.RemoveAll(newInstallPath)
		return um.createFailureResult(tool, oldVersion, targetVersion, startTime, fmt.Errorf("post-install failed: %w", err), false, "")
	}

	// Update database record
	newInstallation := &repository.Installation{
		Tool:        tool,
		Version:     targetVersion,
		Backend:     installation.Backend,
		Provider:    installation.Provider,
		InstallPath: newInstallPath,
		Checksum:    versionInfo.Checksum,
		InstalledAt: time.Now(),
		Metadata:    installation.Metadata,
	}

	// Delete old installation record
	if err := tx.InstallationRepo().Delete(ctx, tool, oldVersion); err != nil {
		os.RemoveAll(newInstallPath)
		return um.createFailureResult(tool, oldVersion, targetVersion, startTime, fmt.Errorf("failed to delete old installation record: %w", err), false, "")
	}

	// Create new installation record
	if err := tx.InstallationRepo().Create(ctx, newInstallation); err != nil {
		os.RemoveAll(newInstallPath)
		return um.createFailureResult(tool, oldVersion, targetVersion, startTime, fmt.Errorf("failed to create installation record: %w", err), true, "database update failed")
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		os.RemoveAll(newInstallPath)
		return um.createFailureResult(tool, oldVersion, targetVersion, startTime, fmt.Errorf("failed to commit transaction: %w", err), true, "transaction commit failed")
	}

	// Remove old installation directory (after successful commit)
	if err := os.RemoveAll(oldInstallPath); err != nil {
		// Log but don't fail - the update was successful
		// The old version directory can be cleaned up manually
	}

	// Update audit log: success
	auditEntry.Status = "success"
	auditEntry.Duration = time.Since(startTime).Milliseconds()
	if err := um.auditRepo.Log(ctx, auditEntry); err != nil {
		// Log but don't fail - the update was successful
	}

	return &UpdateResult{
		Tool:       tool,
		OldVersion: oldVersion,
		NewVersion: targetVersion,
		Success:    true,
		Duration:   time.Since(startTime),
	}, nil
}

// UpdateAll updates all tools to their latest versions.
//
// Validates Requirement 25.3: Support updating all tools to their latest versions
// Validates Requirement 25.4: Respect version constraints in configuration files
func (um *UpdateManager) UpdateAll(ctx context.Context) ([]UpdateResult, error) {
	// Check for updates
	updates, err := um.CheckForUpdates(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check for updates: %w", err)
	}

	var results []UpdateResult

	// Update each tool that has an available update
	for _, update := range updates {
		if !update.UpdateRequired {
			continue
		}

		result, err := um.UpdateTool(ctx, update.Tool, update.LatestVersion)
		if err != nil {
			// Continue with other updates even if one fails
			results = append(results, UpdateResult{
				Tool:       update.Tool,
				OldVersion: update.CurrentVersion,
				NewVersion: update.LatestVersion,
				Success:    false,
				Error:      err.Error(),
			})
			continue
		}

		results = append(results, *result)
	}

	return results, nil
}

// PreviewUpdates shows a preview of what will be updated before applying.
//
// Validates Requirement 25.5: Show a preview of what will be updated before applying
func (um *UpdateManager) PreviewUpdates(ctx context.Context) (*UpdatePreview, error) {
	updates, err := um.CheckForUpdates(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check for updates: %w", err)
	}

	// Filter to only updates that are required
	var requiredUpdates []UpdateInfo
	for _, update := range updates {
		if update.UpdateRequired {
			requiredUpdates = append(requiredUpdates, update)
		}
	}

	// Estimate time (rough estimate: 30 seconds per tool)
	estimatedTime := time.Duration(len(requiredUpdates)) * 30 * time.Second

	return &UpdatePreview{
		Updates:       requiredUpdates,
		TotalUpdates:  len(requiredUpdates),
		EstimatedTime: estimatedTime,
	}, nil
}

// EnableAutomaticUpdates enables automatic updates with the specified schedule.
//
// Validates Requirement 25.6: Support automatic updates (opt-in, configurable schedule)
//
// Note: This method stores the automatic update configuration. The actual
// scheduling and execution would be handled by a separate scheduler service
// or cron job that calls UpdateAll periodically.
func (um *UpdateManager) EnableAutomaticUpdates(ctx context.Context, schedule string) error {
	// Store automatic update configuration in metadata
	metadata := map[string]interface{}{
		"automatic_updates_enabled": true,
		"schedule":                  schedule,
		"enabled_at":                time.Now().Format(time.RFC3339),
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Store in audit log for tracking
	auditEntry := &repository.AuditEntry{
		Timestamp: time.Now(),
		Operation: "enable_automatic_updates",
		Status:    "success",
		Metadata:  string(metadataJSON),
	}

	if err := um.auditRepo.Log(ctx, auditEntry); err != nil {
		return fmt.Errorf("failed to log automatic updates configuration: %w", err)
	}

	return nil
}

// DisableAutomaticUpdates disables automatic updates.
func (um *UpdateManager) DisableAutomaticUpdates(ctx context.Context) error {
	// Store automatic update configuration in metadata
	metadata := map[string]interface{}{
		"automatic_updates_enabled": false,
		"disabled_at":               time.Now().Format(time.RFC3339),
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Store in audit log for tracking
	auditEntry := &repository.AuditEntry{
		Timestamp: time.Now(),
		Operation: "disable_automatic_updates",
		Status:    "success",
		Metadata:  string(metadataJSON),
	}

	if err := um.auditRepo.Log(ctx, auditEntry); err != nil {
		return fmt.Errorf("failed to log automatic updates configuration: %w", err)
	}

	return nil
}

// createFailureResult creates an UpdateResult for a failed update.
func (um *UpdateManager) createFailureResult(
	tool, oldVersion, newVersion string,
	startTime time.Time,
	err error,
	rolledBack bool,
	rollbackReason string,
) (*UpdateResult, error) {
	return &UpdateResult{
		Tool:           tool,
		OldVersion:     oldVersion,
		NewVersion:     newVersion,
		Success:        false,
		Error:          err.Error(),
		Duration:       time.Since(startTime),
		RolledBack:     rolledBack,
		RollbackReason: rollbackReason,
	}, err
}
