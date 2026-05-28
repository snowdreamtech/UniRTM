// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestMigrationManager_MigrateFile_ToolVersions(t *testing.T) {
	tmpDir := t.TempDir()
	sourcePath := filepath.Join(tmpDir, ".tool-versions")

	content := []byte("nodejs 20.0.0\ngo 1.21.0\n# comment\n")
	if err := os.WriteFile(sourcePath, content, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	mm := NewMigrationManager()
	outputPath := filepath.Join(tmpDir, ".unirtm.toml")

	report, err := mm.MigrateFile(context.Background(), sourcePath, outputPath, false)
	if err != nil {
		t.Fatalf("MigrateFile failed: %v", err)
	}

	if report.Source != string(MigrationSourceToolVersions) {
		t.Errorf("expected source .tool-versions, got %q", report.Source)
	}

	if len(report.Tools) != 2 {
		t.Errorf("expected 2 tools, got %d", len(report.Tools))
	}

	// Verify file was written
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("expected output file to be written")
	}
}

func TestMigrationManager_MigrateFile_MiseToml(t *testing.T) {
	tmpDir := t.TempDir()
	sourcePath := filepath.Join(tmpDir, ".mise.toml")

	content := []byte(`
[tools]
node = "20.0.0"
go = { version = "1.21.0", backend = "go", provider = "go" }

[env]
MISE_ENV = "test"

[tasks.build]
run = "go build"

[settings]
cache_ttl = 3600
`)
	if err := os.WriteFile(sourcePath, content, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	mm := NewMigrationManager()
	outputPath := filepath.Join(tmpDir, ".unirtm.toml")

	report, err := mm.MigrateFile(context.Background(), sourcePath, outputPath, false)
	if err != nil {
		t.Fatalf("MigrateFile failed: %v", err)
	}

	if report.Source != string(MigrationSourceMiseToml) {
		t.Errorf("expected source mise.toml, got %q", report.Source)
	}

	if len(report.Tools) != 2 {
		t.Errorf("expected 2 tools, got %d", len(report.Tools))
	}

	reportStr := mm.FormatReport(report)
	if len(reportStr) == 0 {
		t.Error("expected non-empty report string")
	}
}

func TestMigrationManager_MigrateDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	sourcePath := filepath.Join(tmpDir, ".mise.toml")

	content := []byte(`
[tools]
node = "20.0.0"
`)
	if err := os.WriteFile(sourcePath, content, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	mm := NewMigrationManager()

	reports, err := mm.MigrateDirectory(context.Background(), tmpDir, true) // dry run
	if err != nil {
		t.Fatalf("MigrateDirectory failed: %v", err)
	}

	if len(reports) != 1 {
		t.Errorf("expected 1 report, got %d", len(reports))
	}

	if !reports[0].DryRun {
		t.Error("expected dry-run to be true")
	}
}
