// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package builtin_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/guard"
	"github.com/snowdreamtech/unirtm/internal/guard/builtin"
)

func TestBuiltinRules_WhitespaceAndFix(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "dirty.txt")
	// Trailing spaces on line 1 and line 2
	err := os.WriteFile(filePath, []byte("hello   \r\nworld \n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	rule := builtin.NewTrailingWhitespaceRule()
	ctx := context.Background()

	// 1. Check without fix
	issues, err := rule.Execute(ctx, []string{"dirty.txt"}, guard.ExecuteOptions{
		WorkspaceDir: tmpDir,
		Fix:          false,
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].Fixed {
		t.Errorf("Should not be marked fixed when Fix=false")
	}

	// 2. Execute with fix
	issuesFixed, err := rule.Execute(ctx, []string{"dirty.txt"}, guard.ExecuteOptions{
		WorkspaceDir: tmpDir,
		Fix:          true,
	})
	if err != nil {
		t.Fatalf("Execute with fix failed: %v", err)
	}
	if len(issuesFixed) != 1 || !issuesFixed[0].Fixed {
		t.Fatalf("Expected 1 fixed issue, got: %v", issuesFixed)
	}

	// Verify content on disk
	fixedBytes, _ := os.ReadFile(filePath)
	if string(fixedBytes) != "hello\r\nworld\n" {
		t.Errorf("Unexpected fixed content: %q", string(fixedBytes))
	}
}

func TestBuiltinRules_EndOfFileAndFix(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "noeol.txt")
	// Missing trailing newline
	_ = os.WriteFile(filePath, []byte("line without newline"), 0644)

	rule := builtin.NewEndOfFileRule()
	ctx := context.Background()

	// Check with fix
	issues, err := rule.Execute(ctx, []string{"noeol.txt"}, guard.ExecuteOptions{
		WorkspaceDir: tmpDir,
		Fix:          true,
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if len(issues) != 1 || !issues[0].Fixed {
		t.Fatalf("Expected issue to be fixed, got: %v", issues)
	}

	fixedBytes, _ := os.ReadFile(filePath)
	if string(fixedBytes) != "line without newline\n" {
		t.Errorf("Expected single newline, got: %q", string(fixedBytes))
	}
}

func TestBuiltinRules_Utf8BOM(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "bom.txt")
	// Write UTF-8 BOM + "hello\n"
	_ = os.WriteFile(filePath, append([]byte{0xEF, 0xBB, 0xBF}, []byte("hello\n")...), 0644)

	rule := builtin.NewUtf8BOMRule()
	ctx := context.Background()

	issues, err := rule.Execute(ctx, []string{"bom.txt"}, guard.ExecuteOptions{
		WorkspaceDir: tmpDir,
		Fix:          true,
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if len(issues) != 1 || !issues[0].Fixed {
		t.Fatalf("Expected 1 fixed BOM issue, got: %v", issues)
	}

	fixedBytes, _ := os.ReadFile(filePath)
	if string(fixedBytes) != "hello\n" {
		t.Errorf("Expected BOM to be removed, got: %v", fixedBytes)
	}
}

func TestBuiltinRules_MergeConflict(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "conflict.go")
	conflictContent := `package main

<<<<<<< HEAD
func foo() {}
=======
func bar() {}
>>>>>>> branch
`
	_ = os.WriteFile(filePath, []byte(conflictContent), 0644)

	rule := builtin.NewMergeConflictRule()
	ctx := context.Background()

	issues, err := rule.Execute(ctx, []string{"conflict.go"}, guard.ExecuteOptions{
		WorkspaceDir: tmpDir,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 conflict issue, got: %v", issues)
	}
	if issues[0].Severity != guard.SeverityError {
		t.Errorf("Expected SeverityError for merge conflict")
	}
}

func TestBuiltinRules_PrivateKey(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "secret.key")
	_ = os.WriteFile(filePath, []byte("-----BEGIN RSA PRIVATE KEY-----\nMIIE...\n-----END RSA PRIVATE KEY-----\n"), 0600)

	rule := builtin.NewPrivateKeyRule()
	ctx := context.Background()

	issues, err := rule.Execute(ctx, []string{"secret.key"}, guard.ExecuteOptions{
		WorkspaceDir: tmpDir,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 1 || issues[0].Category != guard.CategorySecret {
		t.Fatalf("Expected 1 secret issue, got: %v", issues)
	}
}

func TestBuiltinRules_EnvSafety(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")
	_ = os.WriteFile(envPath, []byte("DATABASE_URL=postgres://user:super_secret_password_123@db:5432/prod\n"), 0644)

	rule := builtin.NewEnvSafetyRule()
	ctx := context.Background()

	issues, err := rule.Execute(ctx, []string{".env"}, guard.ExecuteOptions{
		WorkspaceDir: tmpDir,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 1 || issues[0].Severity != guard.SeverityError {
		t.Fatalf("Expected 1 env safety error, got: %v", issues)
	}
}

func TestBuiltinRules_LargeFiles(t *testing.T) {
	tmpDir := t.TempDir()
	bigPath := filepath.Join(tmpDir, "big.bin")
	// Write 1.5MB file
	_ = os.WriteFile(bigPath, make([]byte, 1500*1024), 0644)

	rule := builtin.NewLargeFilesRule()
	ctx := context.Background()

	issues, err := rule.Execute(ctx, []string{"big.bin"}, guard.ExecuteOptions{
		WorkspaceDir: tmpDir,
		MaxFileSize:  1024 * 1024, // 1MB limit
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected large file error, got: %v", issues)
	}
}

func TestBuiltinRules_CaseConflict(t *testing.T) {
	rule := builtin.NewCaseConflictRule()
	ctx := context.Background()

	issues, err := rule.Execute(ctx, []string{"pkg/User.go", "pkg/user.go"}, guard.ExecuteOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 case conflict issue, got: %v", issues)
	}
	if !strings.Contains(issues[0].Message, "conflict") {
		t.Errorf("Unexpected message: %s", issues[0].Message)
	}
}

func TestBuiltinRules_Lockfiles(t *testing.T) {
	ruleCoex := builtin.NewLockfileCoexistenceRule()
	ctx := context.Background()

	// 1. Coexistence of pnpm-lock.yaml and package-lock.json in same dir
	issues, err := ruleCoex.Execute(ctx, []string{"pnpm-lock.yaml", "package-lock.json"}, guard.ExecuteOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 coexistence issue, got: %v", issues)
	}

	// 2. Lockfile conflict markers
	tmpDir := t.TempDir()
	lockPath := filepath.Join(tmpDir, "Cargo.lock")
	_ = os.WriteFile(lockPath, []byte("version = 3\n<<<<<<< HEAD\nfoo = 1\n=======\nfoo = 2\n>>>>>>> main\n"), 0644)

	ruleConflict := builtin.NewLockfileConflictsRule()
	issuesConflict, err := ruleConflict.Execute(ctx, []string{"Cargo.lock"}, guard.ExecuteOptions{
		WorkspaceDir: tmpDir,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(issuesConflict) != 1 {
		t.Fatalf("Expected 1 lockfile conflict issue, got: %v", issuesConflict)
	}
}

func TestBuiltinRegistry(t *testing.T) {
	rules := builtin.GetAllBuiltinRules()
	if len(rules) < 9 {
		t.Errorf("Expected at least 9 builtin rules, got %d", len(rules))
	}
}
