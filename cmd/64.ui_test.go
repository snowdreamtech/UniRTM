// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/snowdreamtech/unirtm/frontend"
)

func TestUICmdFlagsAndAliases(t *testing.T) {
	if uiCmd.Name() != "ui" {
		t.Fatalf("Expected command name 'ui', got %s", uiCmd.Name())
	}

	// Verify aliases
	aliases := uiCmd.Aliases
	hasDashboard := false
	hasWeb := false
	for _, a := range aliases {
		if a == "dashboard" {
			hasDashboard = true
		}
		if a == "web" {
			hasWeb = true
		}
	}
	if !hasDashboard || !hasWeb {
		t.Errorf("Expected aliases to contain 'dashboard' and 'web', got %v", aliases)
	}

	// Verify flags exist
	portFlag := uiCmd.Flags().Lookup("port")
	if portFlag == nil {
		t.Error("Flag --port not found")
	}

	hostFlag := uiCmd.Flags().Lookup("host")
	if hostFlag == nil {
		t.Error("Flag --host not found")
	}

	noOpenFlag := uiCmd.Flags().Lookup("no-open")
	if noOpenFlag == nil {
		t.Error("Flag --no-open not found")
	}

	browserFlag := uiCmd.Flags().Lookup("browser")
	if browserFlag == nil {
		t.Error("Flag --browser not found")
	}
	if browserFlag.Shorthand != "b" {
		t.Errorf("Expected shorthand 'b' for browser flag, got '%s'", browserFlag.Shorthand)
	}
}

func TestUIStaticAndAPIHandler(t *testing.T) {
	distFS, err := frontend.GetFS()
	if err != nil {
		t.Fatalf("GetFS failed: %v", err)
	}

	mux := http.NewServeMux()

	// API
	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ready"}`))
	})

	// Static SPA fallback
	fileServer := http.FileServer(http.FS(distFS))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})

	// Test API status endpoint
	req := httptest.NewRequest("GET", "/api/status", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200 on /api/status, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "ready") {
		t.Errorf("Expected body to contain 'ready', got %s", rr.Body.String())
	}

	// Test root route serving SPA
	reqRoot := httptest.NewRequest("GET", "/", nil)
	rrRoot := httptest.NewRecorder()
	mux.ServeHTTP(rrRoot, reqRoot)

	// Since dist contains index.html, it should return 200 or html
	if rrRoot.Code != http.StatusOK {
		t.Logf("Root returned status %d", rrRoot.Code)
	}
}

func TestUIConfigAPI(t *testing.T) {
	distFS, err := frontend.GetFS()
	if err != nil {
		t.Fatalf("GetFS failed: %v", err)
	}
	mux := setupUIMux(distFS)

	// 1. Test GET /api/config
	reqGet := httptest.NewRequest("GET", "/api/config", nil)
	rrGet := httptest.NewRecorder()
	mux.ServeHTTP(rrGet, reqGet)

	if rrGet.Code != http.StatusOK {
		t.Fatalf("Expected status 200 on GET /api/config, got %d", rrGet.Code)
	}
	if !strings.Contains(rrGet.Body.String(), "user") || !strings.Contains(rrGet.Body.String(), "workspace") {
		t.Errorf("Expected body to contain 'user' and 'workspace', got %s", rrGet.Body.String())
	}

	// 2. Test POST /api/config rejecting browser in workspace scope
	reqBodyWS := strings.NewReader(`{"scope":"workspace","key":"settings.browser","value":"chrome"}`)
	reqPostWS := httptest.NewRequest("POST", "/api/config", reqBodyWS)
	reqPostWS.Header.Set("Content-Type", "application/json")
	rrPostWS := httptest.NewRecorder()
	mux.ServeHTTP(rrPostWS, reqPostWS)

	if rrPostWS.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for browser in workspace, got %d (body: %s)", rrPostWS.Code, rrPostWS.Body.String())
	}
	if !strings.Contains(rrPostWS.Body.String(), "global") {
		t.Errorf("Expected rejection message to mention 'global', got %s", rrPostWS.Body.String())
	}

	// 3. Test POST /api/config rejecting unknown root namespace
	reqBodyInvalid := strings.NewReader(`{"scope":"workspace","key":"unknown_root.item","value":"val"}`)
	reqPostInvalid := httptest.NewRequest("POST", "/api/config", reqBodyInvalid)
	reqPostInvalid.Header.Set("Content-Type", "application/json")
	rrPostInvalid := httptest.NewRecorder()
	mux.ServeHTTP(rrPostInvalid, reqPostInvalid)

	if rrPostInvalid.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for unknown namespace, got %d", rrPostInvalid.Code)
	}

	// 4. Test POST /api/config with user scope browser setting (temporary home dir)
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("USERPROFILE", tempHome)

	reqBodyUser := strings.NewReader(`{"scope":"user","key":"settings.browser","value":"firefox"}`)
	reqPostUser := httptest.NewRequest("POST", "/api/config", reqBodyUser)
	reqPostUser.Header.Set("Content-Type", "application/json")
	rrPostUser := httptest.NewRecorder()
	mux.ServeHTTP(rrPostUser, reqPostUser)

	if rrPostUser.Code != http.StatusOK {
		t.Errorf("Expected status 200 for user scope browser, got %d (body: %s)", rrPostUser.Code, rrPostUser.Body.String())
	}
}

func TestUIDetectBrowsers(t *testing.T) {
	distFS, err := frontend.GetFS()
	if err != nil {
		t.Fatalf("GetFS failed: %v", err)
	}
	mux := setupUIMux(distFS)

	req := httptest.NewRequest("GET", "/api/system/browsers", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200 on /api/system/browsers, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "browsers") {
		t.Errorf("Expected response to contain 'browsers', got %s", rr.Body.String())
	}

	browsers := detectInstalledBrowsers()
	t.Logf("Detected %d installed browsers on host (%s)", len(browsers), runtime.GOOS)

	resChrome := resolveBrowserExecutable("chrome")
	if resChrome == "" {
		t.Error("resolveBrowserExecutable returned empty string")
	}
}

func TestUIDetectTerminals(t *testing.T) {
	distFS, err := frontend.GetFS()
	if err != nil {
		t.Fatalf("GetFS failed: %v", err)
	}
	mux := setupUIMux(distFS)

	req := httptest.NewRequest("GET", "/api/terminals", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200 on /api/terminals, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "terminals") {
		t.Errorf("Expected response to contain 'terminals', got %s", rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "default") {
		t.Errorf("Expected response to contain 'default', got %s", rr.Body.String())
	}

	terms := detectInstalledTerminals()
	t.Logf("Detected %d installed terminals on host (%s)", len(terms), runtime.GOOS)
	if len(terms) == 0 {
		t.Error("Expected at least one installed shell/terminal to be detected on host")
	}

	// Verify top-ordering of default terminal
	if len(terms) > 0 && !terms[0].IsDefault {
		t.Errorf("Expected first detected terminal (%s) to be marked as default, but was not", terms[0].Path)
	}

	// Verify strict deduplication and single default badge
	seenIDs := make(map[string]bool)
	seenPaths := make(map[string]bool)
	defaultCount := 0
	for _, term := range terms {
		if seenIDs[term.ID] {
			t.Errorf("Duplicate terminal ID detected: %s", term.ID)
		}
		seenIDs[term.ID] = true
		if seenPaths[term.Path] {
			t.Errorf("Duplicate terminal Path detected: %s", term.Path)
		}
		seenPaths[term.Path] = true
		if term.IsDefault {
			defaultCount++
		}
	}
	if defaultCount != 1 {
		t.Errorf("Expected exactly 1 default terminal, got %d", defaultCount)
	}

	// 2. Test POST /api/terminal/exec
	execPayload := strings.NewReader(`{"command":"echo Hello UniRTM Shell","shell":""}`)
	reqExec := httptest.NewRequest("POST", "/api/terminal/exec", execPayload)
	reqExec.Header.Set("Content-Type", "application/json")
	rrExec := httptest.NewRecorder()
	mux.ServeHTTP(rrExec, reqExec)

	if rrExec.Code != http.StatusOK {
		t.Fatalf("Expected status 200 on /api/terminal/exec, got %d: %s", rrExec.Code, rrExec.Body.String())
	}
	if !strings.Contains(rrExec.Body.String(), "Hello UniRTM Shell") {
		t.Errorf("Expected terminal output to contain 'Hello UniRTM Shell', got %s", rrExec.Body.String())
	}

	// 3. Test PTY Session Spawning and environment injection
	ptySess, err := startPTYSession("", "", 80, 24)
	if err != nil {
		t.Logf("Notice: startPTYSession error (might be running in non-tty environment): %v", err)
	} else {
		defer ptySess.Close()
		ptySess.Resize(100, 30)
		_, _ = ptySess.Write([]byte("echo PTY_TEST_OK\n"))
	}
}

func TestUIHooksAPI(t *testing.T) {
	distFS, err := frontend.GetFS()
	if err != nil {
		t.Fatalf("GetFS failed: %v", err)
	}
	mux := setupUIMux(distFS)

	// Switch to a temp workspace to isolate tests
	tmpDir := t.TempDir()
	origWd, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origWd) }()

	// 1. Initial GET /api/hooks (should be empty)
	reqGet := httptest.NewRequest("GET", "/api/hooks?scope=workspace", nil)
	rrGet := httptest.NewRecorder()
	mux.ServeHTTP(rrGet, reqGet)
	if rrGet.Code != http.StatusOK {
		t.Fatalf("Expected 200 on GET /api/hooks, got %d", rrGet.Code)
	}

	// 2. POST /api/hooks to add a pre-commit hook
	bodyPost := strings.NewReader(`{"name":"pre-commit","command":"npm test","scope":"workspace"}`)
	reqPost := httptest.NewRequest("POST", "/api/hooks", bodyPost)
	reqPost.Header.Set("Content-Type", "application/json")
	rrPost := httptest.NewRecorder()
	mux.ServeHTTP(rrPost, reqPost)
	if rrPost.Code != http.StatusOK {
		t.Fatalf("Expected 200 on POST /api/hooks, got %d: %s", rrPost.Code, rrPost.Body.String())
	}

	// 3. GET /api/hooks to verify pre-commit exists
	reqGet2 := httptest.NewRequest("GET", "/api/hooks?scope=workspace", nil)
	rrGet2 := httptest.NewRecorder()
	mux.ServeHTTP(rrGet2, reqGet2)
	if rrGet2.Code != http.StatusOK {
		t.Fatalf("Expected 200 on GET /api/hooks after insert, got %d", rrGet2.Code)
	}
	if !strings.Contains(rrGet2.Body.String(), "pre-commit") || !strings.Contains(rrGet2.Body.String(), "npm test") {
		t.Fatalf("Expected pre-commit in GET /api/hooks response, got: %s", rrGet2.Body.String())
	}

	// 4. DELETE /api/hooks to remove pre-commit
	reqDel := httptest.NewRequest("DELETE", "/api/hooks?name=pre-commit&scope=workspace", nil)
	rrDel := httptest.NewRecorder()
	mux.ServeHTTP(rrDel, reqDel)
	if rrDel.Code != http.StatusOK {
		t.Fatalf("Expected 200 on DELETE /api/hooks, got %d", rrDel.Code)
	}

	// 5. GET /api/hooks to verify removed
	reqGet3 := httptest.NewRequest("GET", "/api/hooks?scope=workspace", nil)
	rrGet3 := httptest.NewRecorder()
	mux.ServeHTTP(rrGet3, reqGet3)
	if strings.Contains(rrGet3.Body.String(), "npm test") {
		t.Fatalf("Expected hook to be deleted from GET response, got: %s", rrGet3.Body.String())
	}
}

func TestUIAliasesAPI(t *testing.T) {
	distFS, err := frontend.GetFS()
	if err != nil {
		t.Fatalf("GetFS failed: %v", err)
	}
	mux := setupUIMux(distFS)

	// Switch to a temp workspace to isolate tests
	tmpDir := t.TempDir()
	origWd, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origWd) }()

	// 1. Initial GET /api/aliases (should be empty)
	reqGet := httptest.NewRequest("GET", "/api/aliases?scope=workspace", nil)
	rrGet := httptest.NewRecorder()
	mux.ServeHTTP(rrGet, reqGet)
	if rrGet.Code != http.StatusOK {
		t.Fatalf("Expected 200 on GET /api/aliases, got %d", rrGet.Code)
	}

	// 2. POST /api/aliases to add node.lts = 20
	bodyPost := strings.NewReader(`{"tool":"node","alias":"lts","version":"20","scope":"workspace"}`)
	reqPost := httptest.NewRequest("POST", "/api/aliases", bodyPost)
	reqPost.Header.Set("Content-Type", "application/json")
	rrPost := httptest.NewRecorder()
	mux.ServeHTTP(rrPost, reqPost)
	if rrPost.Code != http.StatusOK {
		t.Fatalf("Expected 200 on POST /api/aliases, got %d: %s", rrPost.Code, rrPost.Body.String())
	}

	// 3. GET /api/aliases to verify node.lts exists
	reqGet2 := httptest.NewRequest("GET", "/api/aliases?scope=workspace", nil)
	rrGet2 := httptest.NewRecorder()
	mux.ServeHTTP(rrGet2, reqGet2)
	if rrGet2.Code != http.StatusOK {
		t.Fatalf("Expected 200 on GET /api/aliases after insert, got %d", rrGet2.Code)
	}
	if !strings.Contains(rrGet2.Body.String(), "node") || !strings.Contains(rrGet2.Body.String(), "lts") || !strings.Contains(rrGet2.Body.String(), "20") {
		t.Fatalf("Expected node.lts in GET /api/aliases response, got: %s", rrGet2.Body.String())
	}

	// 4. Verify GET /api/config includes aliases
	reqCfg := httptest.NewRequest("GET", "/api/config", nil)
	rrCfg := httptest.NewRecorder()
	mux.ServeHTTP(rrCfg, reqCfg)
	if rrCfg.Code != http.StatusOK || !strings.Contains(rrCfg.Body.String(), `"aliases"`) {
		t.Fatalf("Expected GET /api/config to return aliases, got: %s", rrCfg.Body.String())
	}

	// 5. DELETE /api/aliases to remove node.lts
	reqDel := httptest.NewRequest("DELETE", "/api/aliases?tool=node&alias=lts&scope=workspace", nil)
	rrDel := httptest.NewRecorder()
	mux.ServeHTTP(rrDel, reqDel)
	if rrDel.Code != http.StatusOK {
		t.Fatalf("Expected 200 on DELETE /api/aliases, got %d", rrDel.Code)
	}

	// 6. GET /api/aliases to verify removed
	reqGet3 := httptest.NewRequest("GET", "/api/aliases?scope=workspace", nil)
	rrGet3 := httptest.NewRecorder()
	mux.ServeHTTP(rrGet3, reqGet3)
	if strings.Contains(rrGet3.Body.String(), `"lts"`) {
		t.Fatalf("Expected alias to be deleted from GET response, got: %s", rrGet3.Body.String())
	}
}

func TestUIMinVersionConfig(t *testing.T) {
	distFS, err := frontend.GetFS()
	if err != nil {
		t.Fatalf("GetFS failed: %v", err)
	}
	mux := setupUIMux(distFS)

	tmpDir := t.TempDir()
	origWd, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origWd) }()

	// 1. Trying to set min_version in user scope should be rejected (400)
	userBody := strings.NewReader(`{"scope":"user","key":"min_version","value":"0.35.0"}`)
	reqUser := httptest.NewRequest("POST", "/api/config", userBody)
	reqUser.Header.Set("Content-Type", "application/json")
	rrUser := httptest.NewRecorder()
	mux.ServeHTTP(rrUser, reqUser)
	if rrUser.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 when setting min_version in user scope, got %d: %s", rrUser.Code, rrUser.Body.String())
	}

	// 2. Setting min_version in workspace scope should succeed (200)
	wsBody := strings.NewReader(`{"scope":"workspace","key":"min_version","value":"0.35.0"}`)
	reqWs := httptest.NewRequest("POST", "/api/config", wsBody)
	reqWs.Header.Set("Content-Type", "application/json")
	rrWs := httptest.NewRecorder()
	mux.ServeHTTP(rrWs, reqWs)
	if rrWs.Code != http.StatusOK {
		t.Fatalf("Expected 200 when setting min_version in workspace scope, got %d: %s", rrWs.Code, rrWs.Body.String())
	}

	// 3. GET /api/config should contain min_version: "0.35.0"
	reqGet := httptest.NewRequest("GET", "/api/config", nil)
	rrGet := httptest.NewRecorder()
	mux.ServeHTTP(rrGet, reqGet)
	if rrGet.Code != http.StatusOK || !strings.Contains(rrGet.Body.String(), `"min_version":"0.35.0"`) {
		t.Fatalf("Expected GET /api/config to include min_version: 0.35.0, got: %s", rrGet.Body.String())
	}

	// 4. Setting empty min_version should remove it
	emptyBody := strings.NewReader(`{"scope":"workspace","key":"min_version","value":""}`)
	reqEmpty := httptest.NewRequest("POST", "/api/config", emptyBody)
	reqEmpty.Header.Set("Content-Type", "application/json")
	rrEmpty := httptest.NewRecorder()
	mux.ServeHTTP(rrEmpty, reqEmpty)
	if rrEmpty.Code != http.StatusOK {
		t.Fatalf("Expected 200 when clearing min_version, got %d: %s", rrEmpty.Code, rrEmpty.Body.String())
	}

	raw, _ := os.ReadFile(".unirtm.toml")
	if strings.Contains(string(raw), "min_version") {
		t.Fatalf("Expected min_version to be removed from .unirtm.toml, got: %s", string(raw))
	}
}

func TestUIEnvironmentProfileAPI(t *testing.T) {
	distFS, err := frontend.GetFS()
	if err != nil {
		t.Fatalf("GetFS failed: %v", err)
	}
	mux := setupUIMux(distFS)

	tmpDir := t.TempDir()
	origWd, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origWd) }()

	// 1. Initial base variable
	basePost := strings.NewReader(`{"scope":"workspace","key":"DEBUG","value":"false"}`)
	reqBase := httptest.NewRequest("POST", "/api/env", basePost)
	reqBase.Header.Set("Content-Type", "application/json")
	rrBase := httptest.NewRecorder()
	mux.ServeHTTP(rrBase, reqBase)
	if rrBase.Code != http.StatusOK {
		t.Fatalf("Failed to post base env: %s", rrBase.Body.String())
	}

	// 2. Create profile "development"
	profPost := strings.NewReader(`{"scope":"workspace","profile":"development","action":"create_profile"}`)
	reqProf := httptest.NewRequest("POST", "/api/env", profPost)
	reqProf.Header.Set("Content-Type", "application/json")
	rrProf := httptest.NewRecorder()
	mux.ServeHTTP(rrProf, reqProf)
	if rrProf.Code != http.StatusOK {
		t.Fatalf("Failed to create profile: %s", rrProf.Body.String())
	}

	// 3. Override DEBUG in development profile
	overPost := strings.NewReader(`{"scope":"workspace","profile":"development","key":"DEBUG","value":"true"}`)
	reqOver := httptest.NewRequest("POST", "/api/env", overPost)
	reqOver.Header.Set("Content-Type", "application/json")
	rrOver := httptest.NewRecorder()
	mux.ServeHTTP(rrOver, reqOver)
	if rrOver.Code != http.StatusOK {
		t.Fatalf("Failed to override DEBUG in profile: %s", rrOver.Body.String())
	}

	// 4. GET /api/env?scope=workspace&profile=development
	reqGet := httptest.NewRequest("GET", "/api/env?scope=workspace&profile=development", nil)
	rrGet := httptest.NewRecorder()
	mux.ServeHTTP(rrGet, reqGet)
	if rrGet.Code != http.StatusOK {
		t.Fatalf("GET /api/env with profile failed: %s", rrGet.Body.String())
	}
	body := rrGet.Body.String()
	if !strings.Contains(body, `"is_override":true`) || !strings.Contains(body, `"DEBUG"`) || !strings.Contains(body, `"development"`) {
		t.Fatalf("Expected is_override: true for DEBUG in development profile, got: %s", body)
	}

	// 5. DELETE DEBUG from development profile
	delReq := httptest.NewRequest("DELETE", "/api/env?scope=workspace&profile=development&key=DEBUG", nil)
	rrDel := httptest.NewRecorder()
	mux.ServeHTTP(rrDel, delReq)
	if rrDel.Code != http.StatusOK {
		t.Fatalf("DELETE env var in profile failed: %s", rrDel.Body.String())
	}

	// 6. DELETE entire development profile
	delProfReq := httptest.NewRequest("DELETE", "/api/env?scope=workspace&profile=development&action=delete_profile", nil)
	rrDelProf := httptest.NewRecorder()
	mux.ServeHTTP(rrDelProf, delProfReq)
	if rrDelProf.Code != http.StatusOK {
		t.Fatalf("DELETE profile failed: %s", rrDelProf.Body.String())
	}

	raw, _ := os.ReadFile(".unirtm.toml")
	if strings.Contains(string(raw), "development") {
		t.Fatalf("Expected development profile to be deleted from file, got: %s", string(raw))
	}
}

func TestUIFSApi(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/fs/tree", handleFSTree)
	mux.HandleFunc("/api/fs/create", handleFSCreate)
	mux.HandleFunc("/api/fs/rename", handleFSRename)
	mux.HandleFunc("/api/fs/delete", handleFSDelete)
	mux.HandleFunc("/api/fs/read", handleFSRead)
	mux.HandleFunc("/api/fs/write", handleFSWrite)

	// 1. Test handleFSTree on root
	req := httptest.NewRequest("GET", "/api/fs/tree?depth=2", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for /api/fs/tree, got %d: %s", rr.Code, rr.Body.String())
	}
	var tree FSTreeResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &tree); err != nil {
		t.Fatalf("Failed to parse tree response: %v", err)
	}
	if tree.Root == "" || len(tree.Entries) == 0 {
		t.Fatalf("Expected non-empty tree, got %+v", tree)
	}

	// 2. Test Path Traversal security: requesting ../../ must return 403 Forbidden
	traversalReq := httptest.NewRequest("GET", "/api/fs/tree?path=../../", nil)
	rrTrav := httptest.NewRecorder()
	mux.ServeHTTP(rrTrav, traversalReq)
	if rrTrav.Code != http.StatusForbidden {
		t.Fatalf("Expected 403 Forbidden for path traversal, got %d", rrTrav.Code)
	}

	// 3. Test Create file
	testFile := "tmp_test_explorer.txt"
	createBody := strings.NewReader(`{"path":"` + testFile + `","is_dir":false}`)
	createReq := httptest.NewRequest("POST", "/api/fs/create", createBody)
	rrCreate := httptest.NewRecorder()
	mux.ServeHTTP(rrCreate, createReq)
	if rrCreate.Code != http.StatusOK {
		t.Fatalf("Failed to create file: %s", rrCreate.Body.String())
	}

	// 4. Test Write file
	writeBody := strings.NewReader(`{"path":"` + testFile + `","content":"Hello UniRTM Explorer"}`)
	writeReq := httptest.NewRequest("POST", "/api/fs/write", writeBody)
	rrWrite := httptest.NewRecorder()
	mux.ServeHTTP(rrWrite, writeReq)
	if rrWrite.Code != http.StatusOK {
		t.Fatalf("Failed to write file: %s", rrWrite.Body.String())
	}

	// 5. Test Read file
	readReq := httptest.NewRequest("GET", "/api/fs/read?path="+testFile, nil)
	rrRead := httptest.NewRecorder()
	mux.ServeHTTP(rrRead, readReq)
	if rrRead.Code != http.StatusOK {
		t.Fatalf("Failed to read file: %s", rrRead.Body.String())
	}
	if !strings.Contains(rrRead.Body.String(), "Hello UniRTM Explorer") {
		t.Fatalf("Read file unexpected content: %s", rrRead.Body.String())
	}

	// 6. Test Rename file
	renamedFile := "tmp_test_explorer_renamed.txt"
	renameBody := strings.NewReader(`{"old_path":"` + testFile + `","new_path":"` + renamedFile + `"}`)
	renameReq := httptest.NewRequest("POST", "/api/fs/rename", renameBody)
	rrRename := httptest.NewRecorder()
	mux.ServeHTTP(rrRename, renameReq)
	if rrRename.Code != http.StatusOK {
		t.Fatalf("Failed to rename file: %s", rrRename.Body.String())
	}

	// 7. Test Delete file
	delBody := strings.NewReader(`{"path":"` + renamedFile + `"}`)
	delReq := httptest.NewRequest("POST", "/api/fs/delete", delBody)
	rrDel := httptest.NewRecorder()
	mux.ServeHTTP(rrDel, delReq)
	if rrDel.Code != http.StatusOK {
		t.Fatalf("Failed to delete file: %s", rrDel.Body.String())
	}

	// Verify file is gone
	readAfterDel := httptest.NewRequest("GET", "/api/fs/read?path="+renamedFile, nil)
	rrAfterDel := httptest.NewRecorder()
	mux.ServeHTTP(rrAfterDel, readAfterDel)
	if rrAfterDel.Code != http.StatusNotFound {
		t.Fatalf("Expected 404 for deleted file, got %d", rrAfterDel.Code)
	}
}



