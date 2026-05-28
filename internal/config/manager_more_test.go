// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestManager_tryLoad(t *testing.T) {
	m := NewConfigManager()

	// Create a dummy config file
	f, _ := os.CreateTemp("", "unirtm-*.toml")
	f.WriteString(`[tools]
t1 = "1.0"
`)
	f.Close()
	defer os.Remove(f.Name())

	cm := m.(*defaultConfigManager)
	c, err := cm.tryLoad(context.Background(), f.Name(), false, nil)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if c.ToolsRaw["t1"] != "1.0" {
		t.Errorf("expected t1 version 1.0")
	}

	// Unreadable
	os.Chmod(f.Name(), 0000)
	_, err = cm.tryLoad(context.Background(), f.Name(), false, nil)
	if err == nil {
		t.Errorf("expected error on unreadable")
	}
	os.Chmod(f.Name(), 0644)

	// Missing
	cMissing, err := cm.tryLoad(context.Background(), f.Name()+".missing", false, nil)
	if err != nil || cMissing != nil {
		t.Errorf("expected nil error and nil config on missing")
	}
}

func TestManager_renderTemplate(t *testing.T) {
	s := renderTemplate("hello {{ env.USER }}", nil)
	if s == "" {
		t.Errorf("expected non-empty string")
	}

	s = renderTemplate("hello {{ env.USER ", nil)
	if s == "" {
		// Just checking it doesn't panic
	}
}

func TestManager_MergeConfig(t *testing.T) {
	m := NewConfigManager()
	c1 := &Config{Env: map[string]interface{}{"A": "1"}}
	c2 := &Config{Env: map[string]interface{}{"B": "2"}}
	cm := m.(*defaultConfigManager)
	c3, _ := cm.Merge(c1, c2)
	if c3.Env["B"] != "2" {
		t.Errorf("Merge failed")
	}
}

type mockTrustManagerMore struct {
	status TrustStatus
}

func (m *mockTrustManagerMore) Trust(path string) error             { return nil }
func (m *mockTrustManagerMore) Untrust(path string) error           { return nil }
func (m *mockTrustManagerMore) TrustStatus(path string) TrustStatus { return m.status }
func (m *mockTrustManagerMore) List() (map[string]string, error)    { return nil, nil }

func TestManager_tryLoadTrust(t *testing.T) {
	tmpDir := t.TempDir()
	p := filepath.Join(tmpDir, "unirtm.toml")
	os.WriteFile(p, []byte(`[tools]
t1="1.0"
[env]
a="b"
[tasks.test]
run = "echo"
`), 0644)

	m := NewConfigManager().(*defaultConfigManager)
	m.trustManager = &mockTrustManagerMore{status: TrustStatusUntrusted}

	ctx := context.Background()
	initialSettings := &Settings{}

	// Untrusted
	cfg, err := m.tryLoad(ctx, p, true, initialSettings)
	if err != nil {
		t.Errorf("expected no error")
	}
	if cfg != nil {
		t.Errorf("expected nil cfg for untrusted")
	}

	// Modified
	m.trustManager = &mockTrustManagerMore{status: TrustStatusModified}
	cfg, err = m.tryLoad(ctx, p, true, initialSettings)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg == nil {
		t.Fatalf("expected cfg")
	}
	if len(cfg.Env) != 0 || len(cfg.Tasks) != 0 {
		t.Errorf("expected stripped env and tasks")
	}

	// Trusted
	m.trustManager = &mockTrustManagerMore{status: TrustStatusTrusted}
	cfg, err = m.tryLoad(ctx, p, true, initialSettings)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg == nil {
		t.Fatalf("expected cfg")
	}
	if len(cfg.Env) == 0 || len(cfg.Tasks) == 0 {
		t.Errorf("expected env and tasks present")
	}
}
