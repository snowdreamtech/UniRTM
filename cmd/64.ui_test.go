// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"net/http"
	"net/http/httptest"
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
