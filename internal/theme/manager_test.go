// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package theme

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestThemeManager_Lifecycle(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "unirtm-theme-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mgr := &Manager{
		baseDir:   tempDir,
		iconsDir:  filepath.Join(tempDir, "icons"),
		colorsDir: filepath.Join(tempDir, "colors"),
	}

	if err := mgr.EnsureDirectories(); err != nil {
		t.Fatalf("EnsureDirectories failed: %v", err)
	}

	// 1. Initial list should contain defaults
	initial, err := mgr.ListThemes()
	if err != nil {
		t.Fatalf("ListThemes failed: %v", err)
	}
	if len(initial.Colors) < 2 {
		t.Errorf("expected at least 2 default color themes, got %d", len(initial.Colors))
	}
	if len(initial.Icons) < 1 {
		t.Errorf("expected at least 1 default icon theme, got %d", len(initial.Icons))
	}

	// 2. Build mock .vsix with color theme & icon theme
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	pkgJSON := `{
		"name": "test-theme-pack",
		"displayName": "Test Theme Pack",
		"version": "1.0.0",
		"description": "Mock test theme pack",
		"contributes": {
			"themes": [
				{
					"label": "Test Dark Color",
					"uiTheme": "vs-dark",
					"path": "./themes/test-dark.json"
				}
			],
			"iconThemes": [
				{
					"id": "test-icons",
					"label": "Test Icon Theme",
					"path": "./icons/test-icons.json"
				}
			]
		}
	}`

	w, _ := zw.Create("extension/package.json")
	w.Write([]byte(pkgJSON))

	wTheme, _ := zw.Create("extension/themes/test-dark.json")
	wTheme.Write([]byte(`{"colors": {"activityBar.background": "#123456"}}`))

	wIcons, _ := zw.Create("extension/icons/test-icons.json")
	wIcons.Write([]byte(`{"iconDefinitions": {"_file": {"fontCharacter": "\\E023"}}}`))

	zw.Close()

	// 3. Install VSIX
	installed, err := mgr.InstallVSIX(&buf)
	if err != nil {
		t.Fatalf("InstallVSIX failed: %v", err)
	}
	if len(installed) != 2 {
		t.Fatalf("expected 2 installed items, got %d", len(installed))
	}

	// 4. Verify listed themes
	afterInstall, err := mgr.ListThemes()
	if err != nil {
		t.Fatalf("ListThemes after install failed: %v", err)
	}

	foundColor := false
	for _, c := range afterInstall.Colors {
		if c.ID == "Test Dark Color" && c.PackageID == "test-theme-pack" {
			foundColor = true
			break
		}
	}
	if !foundColor {
		t.Errorf("installed color theme not found in ListThemes")
	}

	foundIcon := false
	for _, ic := range afterInstall.Icons {
		if ic.ID == "test-icons" && ic.PackageID == "test-theme-pack" {
			foundIcon = true
			break
		}
	}
	if !foundIcon {
		t.Errorf("installed icon theme not found in ListThemes")
	}

	// 5. Test resolve file
	resolved, err := mgr.ResolveThemeFile(ThemeTypeColor, "test-theme-pack", "themes/test-dark.json")
	if err != nil {
		t.Errorf("ResolveThemeFile failed: %v", err)
	}
	if _, err := os.Stat(resolved); err != nil {
		t.Errorf("resolved file does not exist: %v", err)
	}

	// 6. Test delete
	if err := mgr.DeleteTheme(ThemeTypeColor, "test-theme-pack"); err != nil {
		t.Errorf("DeleteTheme failed: %v", err)
	}
	if err := mgr.DeleteTheme(ThemeTypeIcon, "test-theme-pack"); err != nil {
		t.Errorf("DeleteTheme icon failed: %v", err)
	}

	afterDelete, _ := mgr.ListThemes()
	for _, c := range afterDelete.Colors {
		if c.PackageID == "test-theme-pack" {
			t.Errorf("color theme still exists after delete")
		}
	}
}

func TestThemeManager_SecurityZipSlip(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "unirtm-theme-slip-*")
	defer os.RemoveAll(tempDir)

	mgr := &Manager{
		baseDir:   tempDir,
		iconsDir:  filepath.Join(tempDir, "icons"),
		colorsDir: filepath.Join(tempDir, "colors"),
	}
	mgr.EnsureDirectories()

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	w, _ := zw.Create("extension/package.json")
	w.Write([]byte(`{"name": "evil", "contributes": {"themes": [{"label": "evil", "path": "evil.json"}]}}`))

	wEvil, _ := zw.Create("extension/../../../../etc/passwd")
	wEvil.Write([]byte("malicious"))

	zw.Close()

	_, err := mgr.InstallVSIX(&buf)
	if err == nil {
		t.Errorf("expected Zip Slip detection error, got nil")
	}
}
