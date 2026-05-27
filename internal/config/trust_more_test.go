package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTrustManager(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	tm := NewTrustManager()
	
	// Create a dummy config file to trust
	testConfig := filepath.Join(tmpDir, "test.toml")
	os.WriteFile(testConfig, []byte("test=1"), 0644)
	
	status := tm.TrustStatus(testConfig)
	if status != TrustStatusUntrusted {
		t.Errorf("expected untrusted initially")
	}
	
	err := tm.Trust(testConfig)
	if err != nil {
		t.Errorf("expected no error trusting")
	}
	
	status = tm.TrustStatus(testConfig)
	if status != TrustStatusTrusted {
		t.Errorf("expected trusted")
	}
	
	// Modify
	os.WriteFile(testConfig, []byte("test=2"), 0644)
	status = tm.TrustStatus(testConfig)
	if status != TrustStatusModified {
		t.Errorf("expected modified")
	}
	
	// Untrust
	err = tm.Untrust(testConfig)
	if err != nil {
		t.Errorf("expected no error untrusting")
	}
	status = tm.TrustStatus(testConfig)
	if status != TrustStatusUntrusted {
		t.Errorf("expected untrusted after untrust")
	}
	
	// List
	tm.Trust(testConfig)
	list, err := tm.List()
	if err != nil {
		t.Errorf("expected no error listing")
	}
	if len(list) != 1 {
		t.Errorf("expected 1 item in list")
	}
}
