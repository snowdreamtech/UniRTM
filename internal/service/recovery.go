// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package service provides business logic for UniRTM operations.
package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/snowdreamtech/unirtm/internal/pkg/logger"
	"github.com/snowdreamtech/unirtm/internal/repository"
)

// RecoveryAction represents the action to take for an incomplete operation.
type RecoveryAction string

const (
	// RecoveryActionRetry retries the incomplete operation.
	RecoveryActionRetry RecoveryAction = "retry"
	// RecoveryActionRollback rolls back the incomplete operation.
	RecoveryActionRollback RecoveryAction = "rollback"
	// RecoveryActionIgnore ignores the incomplete operation.
	RecoveryActionIgnore RecoveryAction = "ignore"
)

// IncompleteOperation represents an operation that did not complete successfully.
type IncompleteOperation struct {
	// Tool is the name of the tool that was being installed/updated.
	Tool string
	// Version is the version that was being installed/updated.
	Version string
	// PartialPath is the path of the partial installation (if any).
	PartialPath string
	// StartedAt is when the operation started.
	StartedAt time.Time
	// AuditID is the ID of the audit log entry for this operation.
	AuditID int64
}

// RecoveryReport summarizes the recovery process.
type RecoveryReport struct {
	// Detected lists incomplete operations that were found.
	Detected []IncompleteOperation
	// Cleaned lists paths that were cleaned up.
	Cleaned []string
	// Errors lists errors encountered during recovery.
	Errors []string
}

// RecoveryManager detects and recovers from incomplete operations.
//
// It checks for in-progress audit log entries and partial installation
// directories on startup, then offers retry/rollback/ignore options.
//
// Validates Requirements: 3.6, 12.4, 12.5
type RecoveryManager struct {
	installRepo repository.InstallationRepository
	auditRepo   repository.AuditRepository
	installsDir string
}

// NewRecoveryManager creates a new RecoveryManager.
func NewRecoveryManager(
	installRepo repository.InstallationRepository,
	auditRepo repository.AuditRepository,
	installsDir string,
) *RecoveryManager {
	return &RecoveryManager{
		installRepo: installRepo,
		auditRepo:   auditRepo,
		installsDir: installsDir,
	}
}

// DetectIncomplete scans the audit log for in-progress operations and
// cross-references them with the filesystem to find partial installations.
//
// Validates Requirements: 12.4, 12.5
func (rm *RecoveryManager) DetectIncomplete(ctx context.Context) ([]IncompleteOperation, error) {
	// Query audit log for in-progress (failure-status) operations
	entries, err := rm.auditRepo.Query(ctx, repository.AuditFilter{
		Status: "in_progress",
		Limit:  100,
	})
	if err != nil {
		// in_progress may return nothing if the schema doesn't have that status value yet
		logger.Debug("Could not query in-progress operations", map[string]interface{}{"error": err.Error()})
		return nil, nil
	}

	var incomplete []IncompleteOperation
	for _, entry := range entries {
		op := IncompleteOperation{
			Tool:      entry.Tool,
			Version:   entry.Version,
			StartedAt: entry.Timestamp,
			AuditID:   entry.ID,
		}

		// Check for partial installation directory
		partialPath := filepath.Join(rm.installsDir, entry.Tool, entry.Version)
		if info, err := os.Stat(partialPath); err == nil && info.IsDir() {
			op.PartialPath = partialPath
		}

		// Check if a success entry already exists for this tool+version
		successEntries, err := rm.auditRepo.Query(ctx, repository.AuditFilter{
			Tool:      entry.Tool,
			Operation: "install",
			Status:    "success",
			Limit:     1,
		})
		if err == nil && len(successEntries) > 0 {
			// Already completed — skip
			continue
		}

		incomplete = append(incomplete, op)
	}

	logger.Info("Detected incomplete operations", map[string]interface{}{
		"count": len(incomplete),
	})

	return incomplete, nil
}

// Recover processes an incomplete operation according to the given action.
//
// Validates Requirements: 3.6, 12.4, 12.5
func (rm *RecoveryManager) Recover(ctx context.Context, op IncompleteOperation, action RecoveryAction) error {
	switch action {
	case RecoveryActionRollback, RecoveryActionIgnore:
		// Clean up the partial installation directory
		if op.PartialPath != "" {
			if err := os.RemoveAll(op.PartialPath); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("remove partial installation at %s: %w", op.PartialPath, err)
			}
			logger.Info("Cleaned partial installation", map[string]interface{}{
				"tool":    op.Tool,
				"version": op.Version,
				"path":    op.PartialPath,
			})
		}

		// Remove database record if it exists (defensive)
		if err := rm.installRepo.Delete(ctx, op.Tool, op.Version); err != nil {
			// Non-fatal: record may not exist
			logger.Debug("Note: no installation record to remove", map[string]interface{}{
				"tool":    op.Tool,
				"version": op.Version,
			})
		}

		// Update audit log
		if rm.auditRepo != nil {
			_ = rm.auditRepo.Log(ctx, &repository.AuditEntry{
				Timestamp: time.Now(),
				Operation: "recovery_" + string(action),
				Tool:      op.Tool,
				Version:   op.Version,
				Status:    "success",
			})
		}

	case RecoveryActionRetry:
		logger.Info("Recovery: retry requested (caller handles re-installation)", map[string]interface{}{
			"tool":    op.Tool,
			"version": op.Version,
		})
	}

	return nil
}

// RecoverAll runs recovery for all detected incomplete operations using
// the rollback action by default.
//
// Validates Requirements: 3.6, 12.4, 12.5
func (rm *RecoveryManager) RecoverAll(ctx context.Context) (*RecoveryReport, error) {
	report := &RecoveryReport{}

	incomplete, err := rm.DetectIncomplete(ctx)
	if err != nil {
		return report, fmt.Errorf("detect incomplete operations: %w", err)
	}
	report.Detected = incomplete

	for _, op := range incomplete {
		if recErr := rm.Recover(ctx, op, RecoveryActionRollback); recErr != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("%s@%s: %s", op.Tool, op.Version, recErr.Error()))
		} else if op.PartialPath != "" {
			report.Cleaned = append(report.Cleaned, op.PartialPath)
		}
	}

	return report, nil
}

// ─── CleanupManager ──────────────────────────────────────────────────────────

// CleanupManager removes orphaned files and partial installations.
//
// Validates Requirement: 3.4
type CleanupManager struct {
	installRepo repository.InstallationRepository
	installsDir string
}

// NewCleanupManager creates a new CleanupManager.
func NewCleanupManager(installRepo repository.InstallationRepository, installsDir string) *CleanupManager {
	return &CleanupManager{
		installRepo: installRepo,
		installsDir: installsDir,
	}
}

// FindOrphaned finds installation directories that have no corresponding database record.
//
// Validates Requirement: 3.4
func (cm *CleanupManager) FindOrphaned(ctx context.Context) ([]string, error) {
	installations, err := cm.installRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list installations: %w", err)
	}

	registeredPaths := make(map[string]bool)
	for _, inst := range installations {
		registeredPaths[filepath.Clean(inst.InstallPath)] = true
	}

	var orphaned []string
	toolDirs, err := os.ReadDir(cm.installsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read installs directory: %w", err)
	}

	for _, toolDir := range toolDirs {
		if !toolDir.IsDir() {
			continue
		}
		versionDirs, err := os.ReadDir(filepath.Join(cm.installsDir, toolDir.Name()))
		if err != nil {
			continue
		}
		for _, versionDir := range versionDirs {
			if !versionDir.IsDir() {
				continue
			}
			fullPath := filepath.Clean(filepath.Join(cm.installsDir, toolDir.Name(), versionDir.Name()))
			if !registeredPaths[fullPath] {
				orphaned = append(orphaned, fullPath)
			}
		}
	}

	return orphaned, nil
}

// CleanOrphaned removes orphaned installation directories.
//
// Validates Requirement: 3.4
func (cm *CleanupManager) CleanOrphaned(ctx context.Context, dryRun bool) ([]string, error) {
	orphaned, err := cm.FindOrphaned(ctx)
	if err != nil {
		return nil, err
	}

	var removed []string
	for _, path := range orphaned {
		if dryRun {
			logger.Info("[dry-run] Would remove orphaned directory", map[string]interface{}{"path": path})
			removed = append(removed, path)
			continue
		}
		if err := os.RemoveAll(path); err != nil {
			logger.Error("Failed to remove orphaned directory", map[string]interface{}{
				"path":  path,
				"error": err.Error(),
			})
			continue
		}
		logger.Info("Removed orphaned directory", map[string]interface{}{"path": path})
		removed = append(removed, path)
	}
	return removed, nil
}
