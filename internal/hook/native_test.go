package hook

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestNativeRunner(t *testing.T) {
	tmpDir := t.TempDir()
	tomlPath := filepath.Join(tmpDir, ".unirtm.toml")
	
	configContent := `
[hooks]
pre-commit = "echo 'success'"
`
	err := os.WriteFile(tomlPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	n := NativeRunner{}
	if !n.Detect(tmpDir) {
		t.Errorf("Detect should return true for config with hooks")
	}

	// Change dir because config.LoadHierarchy loads from current directory by default in Run()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	err = n.Run(context.Background(), "pre-commit", []string{"safe arg"})
	if err != nil {
		t.Errorf("Run failed: %v", err)
	}

	err = n.Run(context.Background(), "commit-msg", []string{})
	if err != nil {
		t.Errorf("Run should silently succeed for undefined hook, got: %v", err)
	}
}
