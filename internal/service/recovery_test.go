// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/snowdreamtech/unirtm/internal/repository"
)

// MockInstallationRepo for testing CleanupManager
type recoveryMockInstallRepo struct {
	repository.InstallationRepository
	installs []*repository.Installation
}

func (m *recoveryMockInstallRepo) List(ctx context.Context) ([]*repository.Installation, error) {
	return m.installs, nil
}

func (m *recoveryMockInstallRepo) Delete(ctx context.Context, tool, version string) error {
	return nil
}

// MockAuditRepo for testing RecoveryManager
type recoveryMockAuditRepo struct {
	repository.AuditRepository
	entries []*repository.AuditEntry
}

func (m *recoveryMockAuditRepo) Query(ctx context.Context, filter repository.AuditFilter) ([]*repository.AuditEntry, error) {
	var result []*repository.AuditEntry
	for _, e := range m.entries {
		if filter.Status != "" && e.Status != filter.Status {
			continue
		}
		if filter.Tool != "" && e.Tool != filter.Tool {
			continue
		}
		if filter.Operation != "" && e.Operation != filter.Operation {
			continue
		}
		result = append(result, e)
	}
	if filter.Limit > 0 && len(result) > filter.Limit {
		result = result[:filter.Limit]
	}
	return result, nil
}

func (m *recoveryMockAuditRepo) Log(ctx context.Context, entry *repository.AuditEntry) error {
	return nil
}

func TestRecoveryManager_DetectIncomplete(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a partial installation dir
	os.MkdirAll(filepath.Join(tmpDir, "node", "20.0.0"), 0755)

	auditRepo := &recoveryMockAuditRepo{
		entries: []*repository.AuditEntry{
			{
				ID:        1,
				Tool:      "node",
				Version:   "20.0.0",
				Operation: "install",
				Status:    "in_progress",
				Timestamp: time.Now(),
			},
			{
				ID:        2,
				Tool:      "go",
				Version:   "1.21.0",
				Operation: "install",
				Status:    "in_progress",
				Timestamp: time.Now(),
			},
			{
				ID:        3,
				Tool:      "go",
				Version:   "1.21.0",
				Operation: "install",
				Status:    "success",
				Timestamp: time.Now(),
			},
		},
	}

	rm := NewRecoveryManager(nil, auditRepo, tmpDir)

	incomplete, err := rm.DetectIncomplete(context.Background())
	if err != nil {
		t.Fatalf("DetectIncomplete failed: %v", err)
	}

	if len(incomplete) != 1 {
		t.Errorf("expected 1 incomplete operation, got %d", len(incomplete))
	} else {
		if incomplete[0].Tool != "node" {
			t.Errorf("expected incomplete tool 'node', got %q", incomplete[0].Tool)
		}
		if incomplete[0].PartialPath == "" {
			t.Error("expected partial path to be set")
		}
	}
}

func TestRecoveryManager_RecoverAll(t *testing.T) {
	tmpDir := t.TempDir()

	// Create partial install path
	nodePath := filepath.Join(tmpDir, "node", "20.0.0")
	os.MkdirAll(nodePath, 0755)

	auditRepo := &recoveryMockAuditRepo{
		entries: []*repository.AuditEntry{
			{
				ID:        1,
				Tool:      "node",
				Version:   "20.0.0",
				Operation: "install",
				Status:    "in_progress",
				Timestamp: time.Now(),
			},
		},
	}
	installRepo := &recoveryMockInstallRepo{}

	rm := NewRecoveryManager(installRepo, auditRepo, tmpDir)

	report, err := rm.RecoverAll(context.Background())
	if err != nil {
		t.Fatalf("RecoverAll failed: %v", err)
	}

	if len(report.Errors) > 0 {
		t.Errorf("expected 0 errors, got %d", len(report.Errors))
	}
	if len(report.Cleaned) != 1 {
		t.Errorf("expected 1 cleaned path, got %d", len(report.Cleaned))
	}

	if _, err := os.Stat(nodePath); !os.IsNotExist(err) {
		t.Error("expected partial path to be removed")
	}
}

func TestCleanupManager_CleanOrphaned(t *testing.T) {
	tmpDir := t.TempDir()

	// Registered installation
	nodePath := filepath.Join(tmpDir, "node", "20.0.0")
	os.MkdirAll(nodePath, 0755)

	// Orphaned installation
	goPath := filepath.Join(tmpDir, "go", "1.21.0")
	os.MkdirAll(goPath, 0755)

	installRepo := &recoveryMockInstallRepo{
		installs: []*repository.Installation{
			{
				Tool:        "node",
				Version:     "20.0.0",
				InstallPath: nodePath,
			},
		},
	}

	cm := NewCleanupManager(installRepo, tmpDir)

	// Dry run
	removed, err := cm.CleanOrphaned(context.Background(), true)
	if err != nil {
		t.Fatalf("CleanOrphaned (dry-run) failed: %v", err)
	}

	if len(removed) != 1 {
		t.Errorf("expected 1 orphaned path found, got %d", len(removed))
	}

	if _, err := os.Stat(goPath); os.IsNotExist(err) {
		t.Error("expected orphaned path to still exist after dry-run")
	}

	// Real run
	removed, err = cm.CleanOrphaned(context.Background(), false)
	if err != nil {
		t.Fatalf("CleanOrphaned failed: %v", err)
	}

	if len(removed) != 1 {
		t.Errorf("expected 1 orphaned path cleaned, got %d", len(removed))
	}

	if _, err := os.Stat(goPath); !os.IsNotExist(err) {
		t.Error("expected orphaned path to be removed")
	}
}
