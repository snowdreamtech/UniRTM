package hook

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestInstallBridgeScript(t *testing.T) {
	tmpDir := t.TempDir()
	
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Skipf("git init failed, skipping test: %v", err)
	}

	err := InstallBridgeScript(context.Background(), tmpDir, "pre-commit")
	if err != nil {
		t.Errorf("InstallBridgeScript failed: %v", err)
	}

	hookPath := filepath.Join(tmpDir, ".git", "hooks", "pre-commit")
	if _, err := os.Stat(hookPath); os.IsNotExist(err) {
		t.Errorf("expected bridge script to be created at %s", hookPath)
	}
}
