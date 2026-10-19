package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestMultiRepositoryAPI(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/git/repositories", handleGitRepositories)
	mux.HandleFunc("/api/git/status", handleGitStatus)

	// 1. GET /api/git/repositories
	req := httptest.NewRequest(http.MethodGet, "/api/git/repositories", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", rr.Code)
	}

	var resp struct {
		Success      bool                `json:"success"`
		Repositories []GitRepositoryItem `json:"repositories"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}

	if !resp.Success {
		t.Fatalf("Expected success to be true")
	}

	if len(resp.Repositories) == 0 {
		t.Fatalf("Expected at least 1 repository discovered")
	}

	firstRepo := resp.Repositories[0]
	if firstRepo.Name == "" || firstRepo.RootPath == "" {
		t.Fatalf("Repository item missing name or rootPath: %+v", firstRepo)
	}

	// 2. GET /api/git/status with ?repo=<firstRepo.Name>
	statusReq := httptest.NewRequest(http.MethodGet, "/api/git/status?repo="+firstRepo.Name, nil)
	statusRr := httptest.NewRecorder()
	mux.ServeHTTP(statusRr, statusReq)

	if statusRr.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for repo status, got %d", statusRr.Code)
	}

	var statusResp struct {
		Success bool   `json:"success"`
		Branch  string `json:"branch"`
	}
	if err := json.NewDecoder(statusRr.Body).Decode(&statusResp); err != nil {
		t.Fatal(err)
	}
	if !statusResp.Success {
		t.Fatalf("Expected status success to be true")
	}

	// 3. Test resolveGitRepoDir resolution logic
	dummyReq := httptest.NewRequest(http.MethodGet, "/api/git/status?repo="+firstRepo.RootPath, nil)
	resolvedDir := resolveGitRepoDir(dummyReq)
	cleanExpected := filepath.Clean(firstRepo.RootPath)
	cleanResolved := filepath.Clean(resolvedDir)
	if cleanExpected != cleanResolved {
		t.Fatalf("resolveGitRepoDir failed: expected %s, got %s", cleanExpected, cleanResolved)
	}
}
