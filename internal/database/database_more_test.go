package database

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestDatabase_EdgeCases(t *testing.T) {
	tempDir := t.TempDir()
	nestedPath := filepath.Join(tempDir, "nested", "test.db")
	ctx := context.Background()

	// 1. Successful initialization
	config := Config{
		Path:    nestedPath,
		WALMode: true,
	}
	db, err := Open(ctx, config)
	if err != nil {
		t.Fatalf("Failed to open nested db: %v", err)
	}

	if p := db.Path(); p != nestedPath {
		t.Errorf("Path() returned %s, expected %s", p, nestedPath)
	}

	if conn := db.Conn(); conn == nil {
		t.Error("Conn() returned nil")
	}

	if err := db.Ping(ctx); err != nil {
		t.Errorf("Ping failed: %v", err)
	}

	// 2. BeginTx
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Errorf("BeginTx failed: %v", err)
	} else {
		tx.Rollback()
	}

	// 3. GetSchemaVersion
	version, err := db.GetSchemaVersion(ctx)
	if err != nil {
		t.Errorf("GetSchemaVersion failed: %v", err)
	}
	if version < 0 {
		t.Errorf("Invalid version %d", version)
	}

	db.Close()

	// Double close should be fine or return an error gracefully
	_ = db.Close()

	// 4. Invalid Path for MkdirAll to fail
	// A file as a parent directory
	invalidPathDir := filepath.Join(tempDir, "file_as_dir")
	os.WriteFile(invalidPathDir, []byte("content"), 0644)
	invalidConfig := Config{Path: filepath.Join(invalidPathDir, "test.db")}

	_, err = Open(ctx, invalidConfig)
	if err == nil {
		t.Error("Expected error opening db where parent is a file")
	}
}

func TestMigrationManager_Rollback(t *testing.T) {
	tempDir := t.TempDir()
	ctx := context.Background()
	config := Config{Path: filepath.Join(tempDir, "test.db")}
	db, err := Open(ctx, config)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	m := NewMigrationManager(db.Conn())

	// Open runs ApplyMigrations, so we have some migrations.
	// Let's get current version
	v, err := m.GetCurrentVersion(ctx)
	if err != nil {
		t.Fatalf("GetCurrentVersion: %v", err)
	}

	if v == 0 {
		// Insert a fake migration record if none
		_, err := db.Conn().ExecContext(ctx, "CREATE TABLE IF NOT EXISTS schema_migrations (version INTEGER PRIMARY KEY, description TEXT)")
		if err == nil {
			db.Conn().ExecContext(ctx, "INSERT INTO schema_migrations (version, description) VALUES (9999, 'Test')")
		}
	}

	err = m.Rollback(ctx)
	if err != nil {
		t.Logf("Rollback error (expected if down sql is missing or migration not found): %v", err)
	}

	// Create a table schema_migrations manually if GetCurrentVersion fails
	db2, _ := Open(ctx, Config{Path: filepath.Join(tempDir, "test2.db")})
	defer db2.Close()
	// test close connection cases
	db2.Close()
	m2 := NewMigrationManager(db2.Conn())
	_ = m2.ApplyMigrations(ctx)
	_, _ = m2.GetCurrentVersion(ctx)
	_ = m2.Rollback(ctx)
}
