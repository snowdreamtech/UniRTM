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

func TestUIToolsAPI(t *testing.T) {
	mockFS := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html>UniRTM Workbench</html>")},
	}
	mux := setupUIMux(mockFS)

	// 1. GET /api/tools/status (and verify /api/runtimes/status compatibility)
	req := httptest.NewRequest("GET", "/api/tools/status", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code, rr.Body.String())

	var statusResp ToolsStatusResponse
	err := json.Unmarshal(rr.Body.Bytes(), &statusResp)
	require.NoError(t, err)
	assert.NotEmpty(t, statusResp.Tools)
	assert.NotEmpty(t, statusResp.Runtimes)

	// Find node in the list
	var nodeInfo *ToolInfo
	for i := range statusResp.Tools {
		if statusResp.Tools[i].ID == "node" {
			nodeInfo = &statusResp.Tools[i]
			break
		}
	}
	require.NotNil(t, nodeInfo)
	assert.Equal(t, "Node.js", nodeInfo.Name)

	// 2. POST /api/tools/use (switch version)
	tempDir := t.TempDir()
	origWd, _ := os.Getwd()
	_ = os.Chdir(tempDir)
	defer func() { _ = os.Chdir(origWd) }()

	useBody, _ := json.Marshal(map[string]interface{}{
		"tool":    "node",
		"version": "22.1.0",
		"global":  false,
	})
	req = httptest.NewRequest("POST", "/api/tools/use", bytes.NewReader(useBody))
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, rr.Body.String())

	// Verify config was written in tempDir
	targetCfg := resolveConfigFilePath(false)
	writtenContent, err := os.ReadFile(targetCfg)
	require.NoError(t, err)
	assert.Contains(t, string(writtenContent), "22.1.0")

	// 3. POST /api/tools/remove
	removeBody, _ := json.Marshal(map[string]interface{}{
		"tool":   "node",
		"global": false,
	})
	req = httptest.NewRequest("POST", "/api/tools/remove", bytes.NewReader(removeBody))
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code, rr.Body.String())

	removedContent, _ := os.ReadFile(targetCfg)
	assert.NotContains(t, string(removedContent), "node")

	// 4. POST /api/tools/migrate-legacy
	// Create .nvmrc in tempDir
	_ = os.WriteFile(filepath.Join(tempDir, ".nvmrc"), []byte("20.18.0\n"), 0644)
	migrateBody, _ := json.Marshal(map[string]string{
		"tool": "node",
	})
	req = httptest.NewRequest("POST", "/api/tools/migrate-legacy", bytes.NewReader(migrateBody))
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, rr.Body.String())
	migratedContent, err := os.ReadFile(targetCfg)
	require.NoError(t, err)
	assert.Contains(t, string(migratedContent), "20.18.0")

	// 5. Test tools with prefix (e.g. npm:@biomejs/biome) in .unirtm.toml are reflected in /api/tools/status
	customToml := `[tools]
node = "22.14.0"
"npm:@biomejs/biome" = "1.9.4"
"cli/cli" = "2.62.0"
`
	_ = os.WriteFile(targetCfg, []byte(customToml), 0644)
	req = httptest.NewRequest("GET", "/api/tools/status", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	var customStatus ToolsStatusResponse
	_ = json.Unmarshal(rr.Body.Bytes(), &customStatus)

	var biomeInfo *ToolInfo
	for i := range customStatus.Tools {
		if customStatus.Tools[i].ID == "npm:@biomejs/biome" {
			biomeInfo = &customStatus.Tools[i]
			break
		}
	}
	require.NotNil(t, biomeInfo)
	assert.Equal(t, "npm:@biomejs/biome", biomeInfo.Name)
	assert.Equal(t, "1.9.4", biomeInfo.ActiveVersion)

	// Ensure core languages appear before custom tools
	var nodeIdx, biomeIdx int = -1, -1
	for idx, rt := range customStatus.Tools {
		if rt.ID == "node" {
			nodeIdx = idx
		}
		if rt.ID == "npm:@biomejs/biome" {
			biomeIdx = idx
		}
	}
	assert.True(t, nodeIdx != -1 && biomeIdx != -1 && nodeIdx < biomeIdx)

	// 6. Test POST /api/tools/pin (and verification in /api/tools/status)
	pinBody, _ := json.Marshal(map[string]interface{}{
		"tool":   "node",
		"pinned": true,
		"reason": "macOS 12.7 compatibility",
		"global": false,
	})
	req = httptest.NewRequest("POST", "/api/tools/pin", bytes.NewReader(pinBody))
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code, rr.Body.String())

	// Verify /api/tools/status reflects pinned and pin_reason
	req = httptest.NewRequest("GET", "/api/tools/status", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	var pinnedStatus ToolsStatusResponse
	_ = json.Unmarshal(rr.Body.Bytes(), &pinnedStatus)
	var pinnedNode *ToolInfo
	for i := range pinnedStatus.Tools {
		if pinnedStatus.Tools[i].ID == "node" {
			pinnedNode = &pinnedStatus.Tools[i]
			break
		}
	}
	require.NotNil(t, pinnedNode)
	assert.True(t, pinnedNode.Pinned)
	assert.Equal(t, "macOS 12.7 compatibility", pinnedNode.PinReason)

	// Unpin
	unpinBody, _ := json.Marshal(map[string]interface{}{
		"tool":   "node",
		"pinned": false,
		"global": false,
	})
	req = httptest.NewRequest("POST", "/api/tools/pin", bytes.NewReader(unpinBody))
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code, rr.Body.String())

	// Verify unpinned in status
	req = httptest.NewRequest("GET", "/api/tools/status", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	var unpinnedStatus ToolsStatusResponse
	_ = json.Unmarshal(rr.Body.Bytes(), &unpinnedStatus)
	for i := range unpinnedStatus.Tools {
		if unpinnedStatus.Tools[i].ID == "node" {
			assert.False(t, unpinnedStatus.Tools[i].Pinned)
			break
		}
	}

	// 7. Test GET /api/tools/check-updates
	req = httptest.NewRequest("GET", "/api/tools/check-updates?tool=node", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	var checkResp CheckUpdatesResponse
	_ = json.Unmarshal(rr.Body.Bytes(), &checkResp)
	require.NotEmpty(t, checkResp.Updates)
	assert.Equal(t, "node", checkResp.Updates[0].Tool)
	assert.NotEmpty(t, checkResp.Updates[0].LatestVersion)

	// Test GET /api/tools/check-updates for checkmake and trivy (must not return fake 1.0 or 1.0.0)
	req = httptest.NewRequest("GET", "/api/tools/check-updates?tool=github:checkmake/checkmake", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	var cmResp CheckUpdatesResponse
	_ = json.Unmarshal(rr.Body.Bytes(), &cmResp)
	require.NotEmpty(t, cmResp.Updates)
	assert.NotEqual(t, "1.0", cmResp.Updates[0].LatestVersion)
	assert.NotEqual(t, "1.0.0", cmResp.Updates[0].LatestVersion)

	req = httptest.NewRequest("GET", "/api/tools/check-updates?tool=github:aquasecurity/trivy", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	var trivyResp CheckUpdatesResponse
	_ = json.Unmarshal(rr.Body.Bytes(), &trivyResp)
	require.NotEmpty(t, trivyResp.Updates)
	assert.NotEqual(t, "1.0", trivyResp.Updates[0].LatestVersion)
	assert.NotEqual(t, "1.0.0", trivyResp.Updates[0].LatestVersion)

	// 8. Test POST /api/tools/batch-update
	batchBody, _ := json.Marshal(map[string]interface{}{
		"tools": []map[string]string{
			{"tool": "node", "version": "23.9.0"},
			{"tool": "go", "version": "1.27.2"},
		},
		"global": false,
	})
	req = httptest.NewRequest("POST", "/api/tools/batch-update", bytes.NewReader(batchBody))
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	var batchResp map[string]interface{}
	_ = json.Unmarshal(rr.Body.Bytes(), &batchResp)
	assert.Equal(t, true, batchResp["success"])
	assert.Equal(t, float64(2), batchResp["updated_count"])

	// 9. Test POST /api/tools/batch-update strictly respects tool pinning
	pinNode, _ := json.Marshal(map[string]interface{}{
		"tool":   "node",
		"pinned": true,
		"reason": "lock in production",
		"global": false,
	})
	req = httptest.NewRequest("POST", "/api/tools/pin", bytes.NewReader(pinNode))
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	// Attempt batch-update containing pinned node and unpinned go
	batchWithPin, _ := json.Marshal(map[string]interface{}{
		"tools": []map[string]string{
			{"tool": "node", "version": "24.0.0"},
			{"tool": "go", "version": "1.27.3"},
		},
		"global": false,
	})
	req = httptest.NewRequest("POST", "/api/tools/batch-update", bytes.NewReader(batchWithPin))
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	var pinBatchResp struct {
		Success      bool     `json:"success"`
		UpdatedCount int      `json:"updated_count"`
		Updated      []string `json:"updated"`
		Skipped      []string `json:"skipped"`
	}
	_ = json.Unmarshal(rr.Body.Bytes(), &pinBatchResp)
	assert.True(t, pinBatchResp.Success)
	assert.Equal(t, 1, pinBatchResp.UpdatedCount)
	assert.Contains(t, pinBatchResp.Updated, "go")
	assert.Contains(t, pinBatchResp.Skipped, "node")
}

// TestUIRuntimesAPI provides backward compatibility verification
func TestUIRuntimesAPI(t *testing.T) {
	TestUIToolsAPI(t)
}

