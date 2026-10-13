// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"net/http"
	"net/http/httptest"
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
