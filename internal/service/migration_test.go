// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package service

import (
	"context"
	"os"
	"path/filepath"
	"strings"
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

func TestMigrationManager_ParseMiseToml_Branches(t *testing.T) {
	tmpDir := t.TempDir()
	sourcePath := filepath.Join(tmpDir, ".mise.toml")

	content := []byte(`
[tools]
node = "20.0.0"
complex = { noversion = "test" }
badtool = 123

[env]
MISE_ENV = "test"
BAD_ENV = 123

[tasks.build]
run = ["go", "build"]
description = "build it"
depends = ["lint", 123]
env = { FOO = "bar" }

[settings]
cache_ttl = "2h"
locked = true
`)
	os.WriteFile(sourcePath, content, 0644)

	mm := NewMigrationManager()
	report, err := mm.MigrateFile(context.Background(), sourcePath, filepath.Join(tmpDir, "out.toml"), false)
	if err != nil {
		t.Fatalf("MigrateFile failed: %v", err)
	}

	foundComplex := false
	foundBadTool := false
	foundBadEnv := false
	for _, u := range report.UnsupportedFields {
		if u == `tool "complex" has complex configuration that may be lost` {
			foundComplex = true
		}
		if u == `tool "badtool" has unsupported value type int64` {
			foundBadTool = true
		}
		if u == `env var "BAD_ENV" has non-string value and was skipped` {
			foundBadEnv = true
		}
	}
	if !foundComplex || !foundBadTool || !foundBadEnv {
		t.Errorf("missing unsupported fields warnings: %v", report.UnsupportedFields)
	}
}

func TestMigrationManager_ParseMiseToml_Errors(t *testing.T) {
	mm := NewMigrationManager()

	// Read error
	_, err := mm.MigrateFile(context.Background(), "/nonexistent/file.toml", "", false)
	if err == nil {
		t.Error("expected error for nonexistent file")
	}

	// Unmarshal error
	tmpDir := t.TempDir()
	sourcePath := filepath.Join(tmpDir, ".mise.toml")
	os.WriteFile(sourcePath, []byte("invalid \n toml"), 0644)
	_, err = mm.MigrateFile(context.Background(), sourcePath, "", false)
	if err == nil {
		t.Error("expected error for invalid toml")
	}

	// Empty file
	os.WriteFile(sourcePath, []byte(""), 0644)
	report, err := mm.MigrateFile(context.Background(), sourcePath, "", false)
	if err != nil {
		t.Errorf("expected no error for empty file, got: %v", err)
	}
	if len(report.Errors) == 0 || report.Errors[0] != "no tools, env vars, or tasks found in source file" {
		t.Error("expected error message in report for empty file")
	}
}

func TestMigrationManager_FormatReport_Errors(t *testing.T) {
	mm := NewMigrationManager()
	report := &MigrationReport{
		Source:            "src",
		OutputFile:        "out",
		DryRun:            true,
		Tools:             []MigrationTool{{Name: "a", Version: "1"}},
		UnsupportedFields: []string{"warn1"},
		Errors:            []string{"err1"},
	}
	str := mm.FormatReport(report)
	if !strings.Contains(str, "warn1") {
		t.Error("expected report to contain warnings")
	}
	if !strings.Contains(str, "err1") {
		t.Error("expected report to contain errors")
	}
	if !strings.Contains(str, "dry-run") {
		t.Error("expected report to contain dry-run mode")
	}
}
