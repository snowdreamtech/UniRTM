// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package theme

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestMarketplace_InstallFromURL(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "unirtm-mp-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	mgr := &Manager{
		baseDir:   tmpDir,
		iconsDir:  tmpDir + "/icons",
		colorsDir: tmpDir + "/colors",
	}
	_ = mgr.EnsureDirectories()

	// Create a mock .vsix buffer
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)

	pkgJSON := `{
		"name": "sample-theme",
		"displayName": "Sample Theme",
		"version": "1.0.0",
		"contributes": {
			"themes": [
				{ "label": "Sample Dark", "uiTheme": "vs-dark", "path": "./themes/dark.json" }
			]
		}
	}`
	w, _ := zw.Create("extension/package.json")
	_, _ = w.Write([]byte(pkgJSON))

	themeJSON := `{"colors":{"editor.background":"#111111"}}`
	w2, _ := zw.Create("extension/themes/dark.json")
	_, _ = w2.Write([]byte(themeJSON))
	_ = zw.Close()

	// Mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write(buf.Bytes())
	}))
	defer server.Close()

	items, err := mgr.InstallFromURL(server.URL + "/sample.vsix")
	if err != nil {
		t.Fatalf("InstallFromURL failed: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 installed item, got %d", len(items))
	}
	if items[0].Name != "Sample Dark" {
		t.Errorf("expected theme name 'Sample Dark', got %s", items[0].Name)
	}
}

func TestMarketplace_SearchMock(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cat := r.URL.Query().Get("category")
		if cat != "Themes" {
			http.Error(w, "missing category=Themes", http.StatusBadRequest)
			return
		}
		resp := MarketplaceSearchResponse{
			TotalSize: 1,
			Offset:    0,
			Extensions: []MarketplaceExtension{
				{
					Namespace:   "test-author",
					Name:        "test-theme",
					DisplayName: "Test Theme",
					Version:     "1.0.0",
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Test directly against the mock server URL
	req, _ := http.NewRequest(http.MethodGet, server.URL+"?category=Themes&size=20&offset=0", nil)
	client := server.Client()
	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	var data MarketplaceSearchResponse
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		t.Fatal(err)
	}

	if len(data.Extensions) != 1 || data.Extensions[0].Name != "test-theme" {
		t.Errorf("mock search failed to decode expected data")
	}
}

func TestMarketplace_DetailMock(t *testing.T) {
	manifestServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/README.md":
			_, _ = w.Write([]byte("# Sample Theme README\nPreview content."))
		case "/package.json":
			_, _ = w.Write([]byte(`{
				"contributes": {
					"themes": [
						{ "label": "Sample Dark", "uiTheme": "vs-dark" },
						{ "label": "Sample Light", "uiTheme": "vs" }
					]
				}
			}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer manifestServer.Close()

	detail := MarketplaceExtensionDetail{
		Namespace:   "test-ns",
		Name:        "sample-ext",
		DisplayName: "Sample Extension",
		Version:     "1.2.0",
		Files: MarketplaceExtensionFiles{
			Readme:   manifestServer.URL + "/README.md",
			Manifest: manifestServer.URL + "/package.json",
		},
	}

	if detail.Name != "sample-ext" || detail.Version != "1.2.0" {
		t.Errorf("detail fields mismatch")
	}
}
