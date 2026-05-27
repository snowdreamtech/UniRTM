package sqlite

import (
	"context"
	"database/sql"
	"sync"
	"testing"
	"time"

	"github.com/snowdreamtech/unirtm/internal/database"
	"github.com/snowdreamtech/unirtm/internal/repository"
)

func TestInstallationRepository_UpsertAndListByTool(t *testing.T) {
	dbPath := ":memory:"
	db, err := database.Open(context.Background(), database.Config{Path: dbPath})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	repo, err := NewInstallationRepository(db.Conn())
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()

	inst := &repository.Installation{
		Tool:        "node",
		Version:     "18.0.0",
		InstallPath: "/opt/node",
		InstalledAt: time.Now(),
	}

	err = repo.Upsert(ctx, inst)
	if err != nil {
		t.Fatalf("failed to upsert: %v", err)
	}

	// Upsert again to test update
	inst.InstallPath = "/opt/node2"
	err = repo.Upsert(ctx, inst)
	if err != nil {
		t.Fatalf("failed to upsert update: %v", err)
	}

	list, err := repo.ListByTool(ctx, "node")
	if err != nil {
		t.Fatalf("failed to ListByTool: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 item, got %d", len(list))
	}
	if list[0].InstallPath != "/opt/node2" {
		t.Errorf("expected updated path, got %s", list[0].InstallPath)
	}
}

func TestAuditRepository_BuildQueryWithFilters(t *testing.T) {
	dbPath := ":memory:"
	db, err := database.Open(context.Background(), database.Config{Path: dbPath})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	repo, err := NewAuditRepository(db.Conn())
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()

	// Add an entry
	err = repo.Log(ctx, &repository.AuditEntry{
		Operation: "install",
		Tool:      "node",
		Version:   "18.0.0",
		Status:    "success",
		Timestamp: time.Now(),
	})
	if err != nil {
		t.Fatalf("failed to log: %v", err)
	}

	start := time.Now().Add(-24 * time.Hour)
	end := time.Now().Add(24 * time.Hour)
	filter := repository.AuditFilter{
		Operation: "install",
		Tool:      "node",
		Status:    "success",
		StartTime: &start,
		EndTime:   &end,
		Limit:     10,
		Offset:    0,
	}

	logs, err := repo.Query(ctx, filter)
	if err != nil {
		t.Fatalf("failed to query with filters: %v", err)
	}
	if len(logs) != 1 {
		t.Errorf("expected 1 log, got %d", len(logs))
	}
}

func TestDatabase_ConcurrentOperations(t *testing.T) {
	dbPath := ":memory:"
	db, err := database.Open(context.Background(), database.Config{Path: dbPath})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	repo, err := NewCacheRepository(db.Conn())
	if err != nil {
		t.Fatalf("failed to create cache repo: %v", err)
	}

	var wg sync.WaitGroup
	ctx := context.Background()
	concurrency := 100 // High concurrency to trigger DB races if any

	// Test concurrent inserts
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			err := repo.Set(ctx, "key", []byte("val"), 1*time.Hour)
			if err != nil {
				t.Errorf("concurrent insert failed: %v", err)
			}
		}(i)
	}
	wg.Wait()

	_, err = repo.Get(ctx, "key")
	if err != nil && err != sql.ErrNoRows {
		t.Fatalf("failed to get: %v", err)
	}

	// Concurrent deletes
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_ = repo.Delete(ctx, "key")
		}(i)
	}
	wg.Wait()
}
