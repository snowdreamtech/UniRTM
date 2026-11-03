// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGitProfileManager_CRUD(t *testing.T) {
	tmpDir := t.TempDir()
	storeFile := filepath.Join(tmpDir, "git_profiles.json")

	mgr := &GitProfileManager{
		storeFilePath: storeFile,
		profiles:      make(map[string]GitProfile),
	}

	// 1. Save profile
	p1 := GitProfile{
		ID:        "prof-work",
		Name:      "Work Profile",
		UserName:  "Work Engineer",
		UserEmail: "work@company.com",
		GPGSign:   true,
		IsDefault: true,
	}
	if err := mgr.SaveProfile(p1); err != nil {
		t.Fatalf("SaveProfile failed: %v", err)
	}

	profiles, activeID := mgr.GetProfiles()
	if len(profiles) != 1 || activeID != "prof-work" {
		t.Errorf("Expected 1 profile with active prof-work, got %d, active: %s", len(profiles), activeID)
	}

	// 2. Add second profile
	p2 := GitProfile{
		ID:         "prof-oss",
		Name:       "Open Source",
		UserName:   "OSS Contributor",
		UserEmail:  "contributor@oss.org",
		SSHKeyPath: "~/.ssh/id_ed25519_oss",
		GPGSign:    false,
	}
	if err := mgr.SaveProfile(p2); err != nil {
		t.Fatalf("SaveProfile 2 failed: %v", err)
	}

	// 3. Set active
	if err := mgr.SetActiveProfile("prof-oss"); err != nil {
		t.Fatalf("SetActiveProfile failed: %v", err)
	}
	_, activeID = mgr.GetProfiles()
	if activeID != "prof-oss" {
		t.Errorf("Expected activeID prof-oss, got %s", activeID)
	}

	// 4. Persistence verification: reload from file
	mgr2 := &GitProfileManager{
		storeFilePath: storeFile,
		profiles:      make(map[string]GitProfile),
	}
	if err := mgr2.load(); err != nil {
		t.Fatalf("Reload failed: %v", err)
	}
	profiles2, activeID2 := mgr2.GetProfiles()
	if len(profiles2) != 2 || activeID2 != "prof-oss" {
		t.Errorf("Expected 2 profiles on reload, got %d, active %s", len(profiles2), activeID2)
	}

	// 5. Delete profile
	if err := mgr.DeleteProfile("prof-work"); err != nil {
		t.Fatalf("DeleteProfile failed: %v", err)
	}
	profiles3, _ := mgr.GetProfiles()
	if len(profiles3) != 1 {
		t.Errorf("Expected 1 profile after delete, got %d", len(profiles3))
	}
}

func TestGitProfileManager_ApplyProfileToWorkspace(t *testing.T) {
	// Initialize a temporary Git repo
	repoDir := t.TempDir()
	cmdInit := exec.Command("git", "init", repoDir)
	if out, err := cmdInit.CombinedOutput(); err != nil {
		t.Skipf("git init not available: %s (%v)", string(out), err)
	}

	tmpDir := t.TempDir()
	storeFile := filepath.Join(tmpDir, "git_profiles.json")
	mgr := &GitProfileManager{
		storeFilePath: storeFile,
		profiles:      make(map[string]GitProfile),
	}

	p := GitProfile{
		ID:         "prof-test",
		Name:       "Test Isolation Profile",
		UserName:   "Isolated User",
		UserEmail:  "isolated@unirtm.dev",
		SSHKeyPath: "/tmp/custom_key",
		SigningKey: "3AA5C34371567BD2",
		GPGSign:    true,
	}
	if err := mgr.SaveProfile(p); err != nil {
		t.Fatalf("SaveProfile failed: %v", err)
	}

	// Apply to workspace
	if err := mgr.ApplyProfileToWorkspace(repoDir, "prof-test"); err != nil {
		t.Fatalf("ApplyProfileToWorkspace failed: %v", err)
	}

	// Read local git config directly from .git/config file to verify isolation
	gitConfigFile := filepath.Join(repoDir, ".git", "config")
	content, err := os.ReadFile(gitConfigFile)
	if err != nil {
		t.Fatalf("Failed to read local .git/config: %v", err)
	}
	cfgStr := string(content)

	if !strings.Contains(cfgStr, "Isolated User") {
		t.Errorf("Expected local .git/config to contain 'Isolated User', content: %s", cfgStr)
	}
	if !strings.Contains(cfgStr, "isolated@unirtm.dev") {
		t.Errorf("Expected local .git/config to contain 'isolated@unirtm.dev'")
	}
	if !strings.Contains(cfgStr, "ssh -i /tmp/custom_key -o IdentitiesOnly=yes") {
		t.Errorf("Expected local .git/config to contain strict core.sshCommand, got: %s", cfgStr)
	}
	if !strings.Contains(cfgStr, "3AA5C34371567BD2") {
		t.Errorf("Expected local .git/config to contain signingkey")
	}

	// Verify error handling for non-git repository
	notGitDir := t.TempDir()
	if err := mgr.ApplyProfileToWorkspace(notGitDir, "prof-test"); err == nil {
		t.Errorf("Expected error when applying to non-git directory")
	}
}

func TestGitProfileRoutes_HTTP(t *testing.T) {
	mux := http.NewServeMux()
	RegisterGitProfileRoutes(mux)

	// 1. GET /api/git/profiles
	req := httptest.NewRequest(http.MethodGet, "/api/git/profiles", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}

	// 2. POST /api/git/profiles
	newProf := GitProfile{
		ID:        "api-prof",
		Name:      "API Profile",
		UserName:  "API Tester",
		UserEmail: "api@test.com",
	}
	body, _ := json.Marshal(newProf)
	reqPost := httptest.NewRequest(http.MethodPost, "/api/git/profiles", bytes.NewReader(body))
	wPost := httptest.NewRecorder()
	mux.ServeHTTP(wPost, reqPost)
	if wPost.Code != http.StatusOK {
		t.Errorf("Expected 200 OK for POST, got %d", wPost.Code)
	}

	// 3. GET /api/git/system-keys
	reqKeys := httptest.NewRequest(http.MethodGet, "/api/git/system-keys", nil)
	wKeys := httptest.NewRecorder()
	mux.ServeHTTP(wKeys, reqKeys)
	if wKeys.Code != http.StatusOK {
		t.Errorf("Expected 200 OK for system keys, got %d", wKeys.Code)
	}
	var keys SystemKeysSummary
	if err := json.NewDecoder(wKeys.Body).Decode(&keys); err != nil {
		t.Errorf("Failed to decode system keys response: %v", err)
	}

	// 4. DELETE /api/git/profiles?id=api-prof
	reqDel := httptest.NewRequest(http.MethodDelete, "/api/git/profiles?id=api-prof", nil)
	wDel := httptest.NewRecorder()
	mux.ServeHTTP(wDel, reqDel)
	if wDel.Code != http.StatusOK {
		t.Errorf("Expected 200 OK for DELETE, got %d", wDel.Code)
	}
}
