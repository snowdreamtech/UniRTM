package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestUIWorkspaceRoots(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "unirtm-workspace-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	origWd, _ := os.Getwd()
	_ = os.Chdir(tempDir)
	defer os.Chdir(origWd)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/workspace/roots", handleGetWorkspaceRoots)
	mux.HandleFunc("/api/workspace/roots/add", handleAddWorkspaceRoot)
	mux.HandleFunc("/api/workspace/roots/remove", handleRemoveWorkspaceRoot)

	// 1. Initial GET
	req := httptest.NewRequest("GET", "/api/workspace/roots", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", rr.Code)
	}

	var resp struct {
		Success bool                `json:"success"`
		Roots   []WorkspaceRootItem `json:"roots"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Roots) != 1 {
		t.Fatalf("Expected 1 default root, got %d", len(resp.Roots))
	}

	// 2. Add second root
	extraDir := filepath.Join(tempDir, "pkg-extra")
	_ = os.MkdirAll(extraDir, 0755)

	addBody, _ := json.Marshal(map[string]string{
		"path": "pkg-extra",
		"name": "pkg-extra",
	})
	addReq := httptest.NewRequest("POST", "/api/workspace/roots/add", bytes.NewReader(addBody))
	addRr := httptest.NewRecorder()
	mux.ServeHTTP(addRr, addReq)

	if addRr.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d: %s", addRr.Code, addRr.Body.String())
	}

	var addResp struct {
		Success bool                `json:"success"`
		Roots   []WorkspaceRootItem `json:"roots"`
	}
	_ = json.NewDecoder(addRr.Body).Decode(&addResp)
	if len(addResp.Roots) != 2 {
		t.Fatalf("Expected 2 roots after add, got %d", len(addResp.Roots))
	}

	// 3. Remove second root
	removeBody, _ := json.Marshal(map[string]string{
		"path": "pkg-extra",
	})
	removeReq := httptest.NewRequest("POST", "/api/workspace/roots/remove", bytes.NewReader(removeBody))
	removeRr := httptest.NewRecorder()
	mux.ServeHTTP(removeRr, removeReq)

	if removeRr.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d: %s", removeRr.Code, removeRr.Body.String())
	}

	var removeResp struct {
		Success bool                `json:"success"`
		Roots   []WorkspaceRootItem `json:"roots"`
	}
	_ = json.NewDecoder(removeRr.Body).Decode(&removeResp)
	if len(removeResp.Roots) != 1 {
		t.Fatalf("Expected 1 root after remove, got %d", len(removeResp.Roots))
	}

	// 4. Try removing the only remaining root (should fail)
	failBody, _ := json.Marshal(map[string]string{
		"path": "",
	})
	failReq := httptest.NewRequest("POST", "/api/workspace/roots/remove", bytes.NewReader(failBody))
	failRr := httptest.NewRecorder()
	mux.ServeHTTP(failRr, failReq)

	if failRr.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 Bad Request when removing only root, got %d", failRr.Code)
	}
}

func TestAutoDiscoverSubRepositories(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "unirtm-multi-repo-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// 1. Initially no sub-repos
	roots := loadWorkspaceRoots(tempDir)
	if len(roots) != 1 || roots[0].Path != "" {
		t.Fatalf("Expected 1 default root, got %+v", roots)
	}

	// 2. Create sub-projects:
	// - repo-a with .git
	// - repo-b with .unirtm.toml
	// - normal-dir (not a project)
	repoA := filepath.Join(tempDir, "repo-a")
	_ = os.MkdirAll(filepath.Join(repoA, ".git"), 0755)

	repoB := filepath.Join(tempDir, "repo-b")
	_ = os.MkdirAll(repoB, 0755)
	_ = os.WriteFile(filepath.Join(repoB, ".unirtm.toml"), []byte(""), 0644)

	normalDir := filepath.Join(tempDir, "normal-dir")
	_ = os.MkdirAll(normalDir, 0755)

	// Sub-repositories should be auto-discovered
	roots = loadWorkspaceRoots(tempDir)
	if len(roots) != 2 {
		t.Fatalf("Expected 2 auto-discovered roots, got %d: %+v", len(roots), roots)
	}
	if roots[0].Path != "repo-a" || roots[1].Path != "repo-b" {
		t.Fatalf("Expected repo-a and repo-b, got %+v", roots)
	}

	// 3. If parent folder itself has .git, it should be included as primary root
	_ = os.MkdirAll(filepath.Join(tempDir, ".git"), 0755)
	roots = loadWorkspaceRoots(tempDir)
	if len(roots) != 3 {
		t.Fatalf("Expected 3 roots when root has .git, got %d: %+v", len(roots), roots)
	}
	if roots[0].ID != "root-primary" || roots[0].Path != "" {
		t.Fatalf("Expected primary root first, got %+v", roots[0])
	}
}
