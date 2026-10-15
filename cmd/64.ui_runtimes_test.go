// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUIRuntimesAPI(t *testing.T) {
	mockFS := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html>UniRTM Workbench</html>")},
	}
	mux := setupUIMux(mockFS)

	// 1. GET /api/runtimes/status
	req := httptest.NewRequest("GET", "/api/runtimes/status", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code, rr.Body.String())

	var statusResp RuntimesStatusResponse
	err := json.Unmarshal(rr.Body.Bytes(), &statusResp)
	require.NoError(t, err)
	assert.NotEmpty(t, statusResp.Runtimes)

	// Find node in the list
	var nodeInfo *RuntimeInfo
	for i := range statusResp.Runtimes {
		if statusResp.Runtimes[i].ID == "node" {
			nodeInfo = &statusResp.Runtimes[i]
			break
		}
	}
	require.NotNil(t, nodeInfo)
	assert.Equal(t, "Node.js", nodeInfo.Name)

	// 2. POST /api/runtimes/use (switch version)
	tempDir := t.TempDir()
	origWd, _ := os.Getwd()
	_ = os.Chdir(tempDir)
	defer func() { _ = os.Chdir(origWd) }()

	useBody, _ := json.Marshal(map[string]interface{}{
		"tool":    "node",
		"version": "22.1.0",
		"global":  false,
	})
	req = httptest.NewRequest("POST", "/api/runtimes/use", bytes.NewReader(useBody))
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, rr.Body.String())

	// Verify config was written in tempDir
	targetCfg := resolveConfigFilePath(false)
	writtenContent, err := os.ReadFile(targetCfg)
	require.NoError(t, err)
	assert.Contains(t, string(writtenContent), "22.1.0")

	// 3. POST /api/runtimes/migrate-legacy
	// Create .nvmrc in tempDir
	_ = os.WriteFile(filepath.Join(tempDir, ".nvmrc"), []byte("20.18.0\n"), 0644)
	migrateBody, _ := json.Marshal(map[string]string{
		"tool": "node",
	})
	req = httptest.NewRequest("POST", "/api/runtimes/migrate-legacy", bytes.NewReader(migrateBody))
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, rr.Body.String())
	migratedContent, err := os.ReadFile(targetCfg)
	require.NoError(t, err)
	assert.Contains(t, string(migratedContent), "20.18.0")

	// 4. Test tools with prefix (e.g. npm:@biomejs/biome) in .unirtm.toml are reflected in /api/runtimes/status
	customToml := `[tools]
node = "22.14.0"
"npm:@biomejs/biome" = "1.9.4"
"cli/cli" = "2.62.0"
`
	_ = os.WriteFile(filepath.Join(tempDir, ".unirtm.toml"), []byte(customToml), 0644)
	req = httptest.NewRequest("GET", "/api/runtimes/status", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	var customStatus RuntimesStatusResponse
	_ = json.Unmarshal(rr.Body.Bytes(), &customStatus)

	var biomeInfo *RuntimeInfo
	for i := range customStatus.Runtimes {
		if customStatus.Runtimes[i].ID == "npm:@biomejs/biome" {
			biomeInfo = &customStatus.Runtimes[i]
			break
		}
	}
	require.NotNil(t, biomeInfo)
	assert.Equal(t, "npm:@biomejs/biome", biomeInfo.Name)
	assert.Equal(t, "1.9.4", biomeInfo.ActiveVersion)

	// Ensure core languages appear before custom tools
	var nodeIdx, biomeIdx int = -1, -1
	for idx, rt := range customStatus.Runtimes {
		if rt.ID == "node" {
			nodeIdx = idx
		}
		if rt.ID == "npm:@biomejs/biome" {
			biomeIdx = idx
		}
	}
	assert.True(t, nodeIdx != -1 && biomeIdx != -1 && nodeIdx < biomeIdx)
}
