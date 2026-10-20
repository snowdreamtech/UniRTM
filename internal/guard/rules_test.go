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

func TestBuiltinRules_Json(t *testing.T) {
	tmpDir := t.TempDir()

	validJson := filepath.Join(tmpDir, "valid.json")
	_ = os.WriteFile(validJson, []byte(`{"name": "unirtm", "version": 1}`), 0644)

	invalidJson := filepath.Join(tmpDir, "invalid.json")
	_ = os.WriteFile(invalidJson, []byte(`{"name": "unirtm", "unclosed": }`), 0644)

	jsonc := filepath.Join(tmpDir, "tsconfig.json")
	_ = os.WriteFile(jsonc, []byte(`{
		// Single line comment
		/* Multi-line
		   comment */
		"compilerOptions": {
			"target": "es2022"
		}
	}`), 0644)

	rule := guard.NewJsonRule()
	ctx := context.Background()

	issues, err := rule.Execute(ctx, []string{"valid.json", "invalid.json", "tsconfig.json"}, guard.ExecuteOptions{
		WorkspaceDir: tmpDir,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 invalid JSON issue, got %d: %v", len(issues), issues)
	}
	if issues[0].FilePath != "invalid.json" {
		t.Errorf("Expected invalid.json to fail, got: %s", issues[0].FilePath)
	}
}

func TestBuiltinRules_Toml(t *testing.T) {
	tmpDir := t.TempDir()

	validToml := filepath.Join(tmpDir, "valid.toml")
	_ = os.WriteFile(validToml, []byte("[project]\nname = \"unirtm\"\nversion = \"0.35.0\"\n"), 0644)

	invalidToml := filepath.Join(tmpDir, "invalid.toml")
	_ = os.WriteFile(invalidToml, []byte("[project\nname = \"unirtm\"\n"), 0644)

	rule := guard.NewTomlRule()
	ctx := context.Background()

	issues, err := rule.Execute(ctx, []string{"valid.toml", "invalid.toml"}, guard.ExecuteOptions{
		WorkspaceDir: tmpDir,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 invalid TOML issue, got %d: %v", len(issues), issues)
	}
	if issues[0].FilePath != "invalid.toml" {
		t.Errorf("Expected invalid.toml to fail, got: %s", issues[0].FilePath)
	}
}

func TestBuiltinRules_MixedLineEndingsAndFix(t *testing.T) {
	tmpDir := t.TempDir()

	mixedPath := filepath.Join(tmpDir, "mixed.txt")
	// Line 1 ends with CRLF, Line 2 ends with LF
	_ = os.WriteFile(mixedPath, []byte("line1\r\nline2\nline3\r\n"), 0644)

	rule := guard.NewMixedLineEndingsRule()
	ctx := context.Background()

	// 1. Without fix
	issues, err := rule.Execute(ctx, []string{"mixed.txt"}, guard.ExecuteOptions{
		WorkspaceDir: tmpDir,
		Fix:          false,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 mixed line ending issue, got %d: %v", len(issues), issues)
	}
	if issues[0].Fixed {
		t.Errorf("Should not be marked fixed when Fix=false")
	}

	// 2. With fix
	issuesFixed, err := rule.Execute(ctx, []string{"mixed.txt"}, guard.ExecuteOptions{
		WorkspaceDir: tmpDir,
		Fix:          true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(issuesFixed) != 1 || !issuesFixed[0].Fixed {
		t.Fatalf("Expected issue to be fixed, got: %v", issuesFixed)
	}

	// Verify file content on disk is now strictly LF
	fixedContent, _ := os.ReadFile(mixedPath)
	if string(fixedContent) != "line1\nline2\nline3\n" {
		t.Errorf("Unexpected content after fix: %q", string(fixedContent))
	}
}

func TestBuiltinRules_Xml(t *testing.T) {
	tmpDir := t.TempDir()

	validXml := filepath.Join(tmpDir, "valid.xml")
	_ = os.WriteFile(validXml, []byte(`<?xml version="1.0" encoding="UTF-8"?><root><item id="1">Text</item></root>`), 0644)

	invalidXml := filepath.Join(tmpDir, "invalid.xml")
	_ = os.WriteFile(invalidXml, []byte(`<?xml version="1.0"?><root><unclosed>text</root>`), 0644)

	emptyXml := filepath.Join(tmpDir, "empty.xml")
	_ = os.WriteFile(emptyXml, []byte("   \n"), 0644)

	rule := guard.NewXmlRule()
	ctx := context.Background()

	issues, err := rule.Execute(ctx, []string{"valid.xml", "invalid.xml", "empty.xml"}, guard.ExecuteOptions{
		WorkspaceDir: tmpDir,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 2 {
		t.Fatalf("Expected 2 invalid XML issues, got %d: %v", len(issues), issues)
	}
}

func TestBuiltinRules_CloudSecrets(t *testing.T) {
	tmpDir := t.TempDir()

	dirtyFile := filepath.Join(tmpDir, "config.py")
	_ = os.WriteFile(dirtyFile, []byte("aws_key = 'AKIA1234567890ABCDEF'\ngithub_token = 'ghp_1234567890abcdefghijklmnopqrstuvwxyz'\n"), 0644)

	cleanFile := filepath.Join(tmpDir, "clean.go")
	_ = os.WriteFile(cleanFile, []byte("package main\nconst token = \"normal_string\"\n"), 0644)

	rule := guard.NewCloudSecretsRule()
	ctx := context.Background()

	issues, err := rule.Execute(ctx, []string{"config.py", "clean.go"}, guard.ExecuteOptions{
		WorkspaceDir: tmpDir,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 2 {
		t.Fatalf("Expected 2 cloud secret issues in config.py, got %d: %v", len(issues), issues)
	}
}

func TestBuiltinRules_CommitMsg(t *testing.T) {
	// 1. Direct string validation
	ok, _ := guard.ValidateCommitMessage("feat(guard): add commit-msg validation")
	if !ok {
		t.Errorf("Expected valid conventional commit to pass")
	}

	ok, _ = guard.ValidateCommitMessage("Merge branch 'main' into dev")
	if !ok {
		t.Errorf("Expected git merge commit to pass")
	}

	okBadType, _ := guard.ValidateCommitMessage("unknown(guard): something")
	if okBadType {
		t.Errorf("Expected invalid type to fail")
	}

	okTrailingPeriod, _ := guard.ValidateCommitMessage("fix: fix bug.")
	if okTrailingPeriod {
		t.Errorf("Expected subject with trailing period to fail")
	}

	// 2. Rule execution against COMMIT_EDITMSG file
	tmpDir := t.TempDir()
	validMsgFile := filepath.Join(tmpDir, "COMMIT_EDITMSG")
	_ = os.WriteFile(validMsgFile, []byte("docs: update README\n\nDetailed explanation here.\n"), 0644)

	rule := guard.NewCommitMsgRule()
	ctx := context.Background()

	issues, err := rule.Execute(ctx, []string{"COMMIT_EDITMSG"}, guard.ExecuteOptions{
		WorkspaceDir: tmpDir,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 0 {
		t.Fatalf("Expected 0 issues for valid commit message, got: %v", issues)
	}

	// Corrupt commit message
	_ = os.WriteFile(validMsgFile, []byte("invalid commit message without type\n"), 0644)
	issuesBad, err := rule.Execute(ctx, []string{"COMMIT_EDITMSG"}, guard.ExecuteOptions{
		WorkspaceDir: tmpDir,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(issuesBad) != 1 {
		t.Fatalf("Expected 1 issue for invalid commit message, got: %v", issuesBad)
	}
}

func TestBuiltinRules_Yaml(t *testing.T) {
	tmpDir := t.TempDir()

	validYaml := filepath.Join(tmpDir, "config.yaml")
	_ = os.WriteFile(validYaml, []byte("version: '3.8'\nservices:\n  app:\n    image: nginx:latest\n---\nkind: ConfigMap\n"), 0644)

	invalidYaml := filepath.Join(tmpDir, "bad.yaml")
	_ = os.WriteFile(invalidYaml, []byte("key: [unclosed bracket\n"), 0644)

	emptyYaml := filepath.Join(tmpDir, "empty.yml")
	_ = os.WriteFile(emptyYaml, []byte("   \n"), 0644)

	rule := guard.NewYamlRule()
	ctx := context.Background()

	issues, err := rule.Execute(ctx, []string{"config.yaml", "bad.yaml", "empty.yml"}, guard.ExecuteOptions{
		WorkspaceDir: tmpDir,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 invalid YAML issue, got %d: %v", len(issues), issues)
	}
	if issues[0].FilePath != "bad.yaml" {
		t.Errorf("Expected bad.yaml to fail, got: %s", issues[0].FilePath)
	}
}

func TestBuiltinRules_DebugStatements(t *testing.T) {
	tmpDir := t.TempDir()

	goodJs := filepath.Join(tmpDir, "index.js")
	_ = os.WriteFile(goodJs, []byte("console.log('hello');\n// debugger;\n"), 0644)

	badJs := filepath.Join(tmpDir, "app.ts")
	_ = os.WriteFile(badJs, []byte("function test() {\n  debugger;\n}\n"), 0644)

	badPy := filepath.Join(tmpDir, "main.py")
	_ = os.WriteFile(badPy, []byte("import pdb; pdb.set_trace()\n"), 0644)

	testJs := filepath.Join(tmpDir, "app.test.ts")
	_ = os.WriteFile(testJs, []byte("debugger;\n"), 0644)

	rule := guard.NewDebugStatementsRule()
	ctx := context.Background()

	issues, err := rule.Execute(ctx, []string{"index.js", "app.ts", "main.py", "app.test.ts"}, guard.ExecuteOptions{
		WorkspaceDir: tmpDir,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 2 {
		t.Fatalf("Expected 2 debug statement issues, got %d: %v", len(issues), issues)
	}
}

func TestBuiltinRules_BinaryArtifacts(t *testing.T) {
	tmpDir := t.TempDir()

	goodSource := filepath.Join(tmpDir, "main.go")
	_ = os.WriteFile(goodSource, []byte("package main\n"), 0644)

	badExe := filepath.Join(tmpDir, "app.exe")
	_ = os.WriteFile(badExe, []byte("MZ\x90\x00"), 0644)

	badZip := filepath.Join(tmpDir, "archive.tar.gz")
	_ = os.WriteFile(badZip, []byte("fake tar gz content"), 0644)

	rule := guard.NewBinaryArtifactsRule()
	ctx := context.Background()

	issues, err := rule.Execute(ctx, []string{"main.go", "app.exe", "archive.tar.gz"}, guard.ExecuteOptions{
		WorkspaceDir: tmpDir,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 2 {
		t.Fatalf("Expected 2 binary artifact issues, got %d: %v", len(issues), issues)
	}
}

func TestBuiltinRules_EmptyCodeFiles(t *testing.T) {
	tmpDir := t.TempDir()

	goodSource := filepath.Join(tmpDir, "main.go")
	_ = os.WriteFile(goodSource, []byte("package main\n"), 0644)

	whitelistedInit := filepath.Join(tmpDir, "__init__.py")
	_ = os.WriteFile(whitelistedInit, []byte(""), 0644)

	whitelistedKeep := filepath.Join(tmpDir, ".gitkeep")
	_ = os.WriteFile(whitelistedKeep, []byte(""), 0644)

	badEmptyGo := filepath.Join(tmpDir, "empty.go")
	_ = os.WriteFile(badEmptyGo, []byte(""), 0644)

	badEmptyTs := filepath.Join(tmpDir, "empty.ts")
	_ = os.WriteFile(badEmptyTs, []byte(""), 0644)

	rule := guard.NewEmptyCodeFilesRule()
	ctx := context.Background()

	issues, err := rule.Execute(ctx, []string{"main.go", "__init__.py", ".gitkeep", "empty.go", "empty.ts"}, guard.ExecuteOptions{
		WorkspaceDir: tmpDir,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 2 {
		t.Fatalf("Expected 2 empty file issues, got %d: %v", len(issues), issues)
	}
}

func TestBuiltinRegistry(t *testing.T) {
	rules := guard.GetAllBuiltinRules()
	if len(rules) < 24 {
		t.Errorf("Expected at least 24 builtin rules, got %d", len(rules))
	}
}







