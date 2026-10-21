// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package guard

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadGuardConfig_OverridesAndNewSlots(t *testing.T) {
	tmpDir := t.TempDir()
	tomlContent := `
[guard]
enabled = true
mode = "staged"

[guard.slots."web.formatter"]
tool = "biome"
command = "biome format {{files}}"
fix_command = "biome format --write {{files}}"
extensions = [".ts", ".tsx", ".js"]

[guard.slots."sql.linter"]
tool = "sqlfluff"
command = "sqlfluff lint {{files}}"
extensions = [".sql"]
category = "lint"
phase = "linter"

[guard.slots."go.imports"]
enabled = false
`
	configPath := filepath.Join(tmpDir, ".unirtm.toml")
	if err := os.WriteFile(configPath, []byte(tomlContent), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadGuardConfig(tmpDir)
	if err != nil {
		t.Fatalf("LoadGuardConfig failed: %v", err)
	}

	if !cfg.Enabled {
		t.Fatalf("expected guard enabled to be true")
	}
	if cfg.Mode != "staged" {
		t.Fatalf("expected mode staged, got %s", cfg.Mode)
	}

	pipeline := NewPipeline(tmpDir, cfg)

	// Verify web.formatter was overridden with biome
	var foundWeb, foundSql, foundGoImports bool
	for _, r := range pipeline.Rules {
		sr, ok := r.(*SlotRule)
		if !ok {
			continue
		}
		if sr.ID() == "slot:web.formatter" {
			foundWeb = true
			if sr.slot.DefaultToolName != "biome" {
				t.Errorf("expected tool biome, got %s", sr.slot.DefaultToolName)
			}
			if sr.slot.DefaultCommand != "biome format {{files}}" {
				t.Errorf("expected command biome format, got %s", sr.slot.DefaultCommand)
			}
		}
		if sr.ID() == "slot:sql.linter" {
			foundSql = true
			if sr.slot.DefaultToolName != "sqlfluff" {
				t.Errorf("expected tool sqlfluff, got %s", sr.slot.DefaultToolName)
			}
			if len(sr.slot.DefaultExtensions) != 1 || sr.slot.DefaultExtensions[0] != ".sql" {
				t.Errorf("expected extension .sql, got %v", sr.slot.DefaultExtensions)
			}
		}
		if sr.ID() == "slot:go.imports" {
			foundGoImports = true
		}
	}

	if !foundWeb {
		t.Errorf("expected slot:web.formatter to be in pipeline rules")
	}
	if !foundSql {
		t.Errorf("expected custom slot:sql.linter to be in pipeline rules")
	}
	if foundGoImports {
		t.Errorf("expected disabled slot:go.imports to NOT be in pipeline rules")
	}
}
