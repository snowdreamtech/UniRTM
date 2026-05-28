// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/snowdreamtech/unirtm/internal/repository"
)

func TestInstallationRepository_Upsert(t *testing.T) {
	db, closeFunc := setupTestDB(t)
	defer closeFunc()

	repo, err := NewInstallationRepository(db.Conn())
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}

	ctx := context.Background()

	inst := &repository.Installation{
		Tool:        "tool-1",
		Version:     "1.0.0",
		Backend:     "github",
		Provider:    "test",
		InstallPath: "/path",
		Checksum:    "abc",
		InstalledAt: time.Now(),
		Metadata:    "meta",
	}

	// insert
	err = repo.Upsert(ctx, inst)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// update
	inst.Metadata = "inactive"
	err = repo.Upsert(ctx, inst)
	if err != nil {
		t.Errorf("expected no error on update, got %v", err)
	}
}

func TestInstallationRepository_ListByTool(t *testing.T) {
	db, closeFunc := setupTestDB(t)
	defer closeFunc()

	repo, err := NewInstallationRepository(db.Conn())
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}

	ctx := context.Background()

	inst := &repository.Installation{
		Tool:        "tool-xyz",
		Version:     "1.0.0",
		Backend:     "github",
		Provider:    "test",
		InstallPath: "/path1",
		Checksum:    "def",
		InstalledAt: time.Now(),
		Metadata:    "meta",
	}

	repo.Create(ctx, inst)

	res, err := repo.ListByTool(ctx, "tool-xyz")
	if err != nil {
		t.Errorf("expected no error listing by tool, got %v", err)
	}
	if len(res) != 1 {
		t.Errorf("expected 1 result, got %d", len(res))
	}
}
