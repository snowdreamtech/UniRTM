// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package guard_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/guard"
)

func TestBuiltinRules_WhitespaceAndFix(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "dirty.txt")
	// Trailing spaces on line 1 and line 2
	err := os.WriteFile(filePath, []byte("hello   \r\nworld \n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	rule := guard.NewTrailingWhitespaceRule()
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

	rule := guard.NewEndOfFileRule()
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

	rule := guard.NewUtf8BOMRule()
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

	rule := guard.NewMergeConflictRule()
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

	rule := guard.NewPrivateKeyRule()
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

	rule := guard.NewEnvSafetyRule()
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

	rule := guard.NewLargeFilesRule()
	ctx := context.Background()

	// 1. Default limit is now 10MB -> 1.5MB should PASS
	issuesDefault, err := rule.Execute(ctx, []string{"big.bin"}, guard.ExecuteOptions{
		WorkspaceDir: tmpDir,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(issuesDefault) != 0 {
		t.Fatalf("Expected 1.5MB file to pass default 10MB limit, got: %v", issuesDefault)
	}

	// 2. Custom limit (1MB) -> 1.5MB should trigger error
	issuesCustom, err := rule.Execute(ctx, []string{"big.bin"}, guard.ExecuteOptions{
		WorkspaceDir: tmpDir,
		MaxFileSize:  1024 * 1024, // 1MB limit
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(issuesCustom) != 1 {
		t.Fatalf("Expected large file error for custom limit, got: %v", issuesCustom)
	}

	// 3. Huge file (11MB) with default limit -> should trigger error
	hugePath := filepath.Join(tmpDir, "huge.bin")
	_ = os.WriteFile(hugePath, make([]byte, 11*1024*1024), 0644)
	issuesHuge, err := rule.Execute(ctx, []string{"huge.bin"}, guard.ExecuteOptions{
		WorkspaceDir: tmpDir,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(issuesHuge) != 1 {
		t.Fatalf("Expected 11MB file to exceed default 10MB limit, got: %v", issuesHuge)
	}
}


func TestBuiltinRules_CaseConflict(t *testing.T) {
	rule := guard.NewCaseConflictRule()
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
	ruleCoex := guard.NewLockfileCoexistenceRule()
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

	ruleConflict := guard.NewLockfileConflictsRule()
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

func TestBuiltinRules_ForbiddenFiles(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, ".DS_Store"), []byte("garbage"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "app.bak"), []byte("backup"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "normal.go"), []byte("package main"), 0644)

	rule := guard.NewForbiddenFilesRule()
	ctx := context.Background()

	issues, err := rule.Execute(ctx, []string{".DS_Store", "app.bak", "normal.go"}, guard.ExecuteOptions{
		WorkspaceDir: tmpDir,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 2 {
		t.Fatalf("Expected 2 forbidden file issues, got: %v", issues)
	}
}

func TestBuiltinRules_Symlinks(t *testing.T) {
	tmpDir := t.TempDir()
	targetFile := filepath.Join(tmpDir, "target.txt")
	_ = os.WriteFile(targetFile, []byte("content"), 0644)

	// Valid relative symlink
	validLink := filepath.Join(tmpDir, "valid_link")
	_ = os.Symlink("target.txt", validLink)

	// Absolute symlink
	absLink := filepath.Join(tmpDir, "abs_link")
	_ = os.Symlink(targetFile, absLink)

	// Broken symlink
	brokenLink := filepath.Join(tmpDir, "broken_link")
	_ = os.Symlink("non_existent.txt", brokenLink)

	rule := guard.NewSymlinksRule()
	ctx := context.Background()

	issues, err := rule.Execute(ctx, []string{"valid_link", "abs_link", "broken_link"}, guard.ExecuteOptions{
		WorkspaceDir: tmpDir,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 2 {
		t.Fatalf("Expected 2 symlink issues (1 abs, 1 broken), got: %v", issues)
	}
}

func TestBuiltinRules_ShebangAndFix(t *testing.T) {
	tmpDir := t.TempDir()

	// 1. Script with shebang but lacking +x
	scriptPath := filepath.Join(tmpDir, "build.sh")
	_ = os.WriteFile(scriptPath, []byte("#!/bin/bash\necho ok\n"), 0644)

	ruleExec := guard.NewShebangScriptsAreExecutableRule()
	ctx := context.Background()

	// Check without fix
	issues, err := ruleExec.Execute(ctx, []string{"build.sh"}, guard.ExecuteOptions{
		WorkspaceDir: tmpDir,
		Fix:          false,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 shebang permission issue, got: %v", issues)
	}

	// Check with fix
	issuesFixed, err := ruleExec.Execute(ctx, []string{"build.sh"}, guard.ExecuteOptions{
		WorkspaceDir: tmpDir,
		Fix:          true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(issuesFixed) != 1 || !issuesFixed[0].Fixed {
		t.Fatalf("Expected issue to be fixed, got: %v", issuesFixed)
	}

	// Verify chmod +x on disk
	fi, _ := os.Stat(scriptPath)
	if fi.Mode()&0111 == 0 {
		t.Errorf("Expected script to have executable bit set after fix")
	}

	// 2. Executable lacking shebang
	noShebangPath := filepath.Join(tmpDir, "run_no_shebang.sh")
	_ = os.WriteFile(noShebangPath, []byte("echo 123\n"), 0755)

	ruleLack := guard.NewExecutablesHaveShebangsRule()
	issuesLack, err := ruleLack.Execute(ctx, []string{"run_no_shebang.sh"}, guard.ExecuteOptions{
		WorkspaceDir: tmpDir,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(issuesLack) != 1 {
		t.Fatalf("Expected 1 lack-shebang issue, got: %v", issuesLack)
	}
}

func TestBuiltinRules_SvgSafety(t *testing.T) {
	tmpDir := t.TempDir()

	safeSvg := filepath.Join(tmpDir, "safe.svg")
	_ = os.WriteFile(safeSvg, []byte(`<svg viewBox="0 0 100 100"><circle cx="50" cy="50" r="40"/></svg>`), 0644)

	unsafeSvg := filepath.Join(tmpDir, "unsafe.svg")
	_ = os.WriteFile(unsafeSvg, []byte(`<svg><script>alert('xss')</script><circle/></svg>`), 0644)

	unsafeEventSvg := filepath.Join(tmpDir, "onload.svg")
	_ = os.WriteFile(unsafeEventSvg, []byte(`<svg onload="alert(1)"><circle/></svg>`), 0644)

	rule := guard.NewSvgSafetyRule()
	ctx := context.Background()

	issues, err := rule.Execute(ctx, []string{"safe.svg", "unsafe.svg", "onload.svg"}, guard.ExecuteOptions{
		WorkspaceDir: tmpDir,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 2 {
		t.Fatalf("Expected 2 unsafe SVG issues, got: %v", issues)
	}
}

func TestBuiltinRegistry(t *testing.T) {
	rules := guard.GetAllBuiltinRules()
	if len(rules) < 14 {
		t.Errorf("Expected at least 14 builtin rules, got %d", len(rules))
	}
}

