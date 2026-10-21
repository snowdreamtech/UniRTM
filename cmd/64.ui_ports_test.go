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
)

func TestPortsAPI(t *testing.T) {
	// 1. Test GET /api/ports/list
	reqList := httptest.NewRequest("GET", "/api/ports/list", nil)
	rrList := httptest.NewRecorder()
	handlePortsList(rrList, reqList)

	if rrList.Code != http.StatusOK {
		t.Fatalf("Expected 200 for /api/ports/list, got %d: %s", rrList.Code, rrList.Body.String())
	}

	var listResp struct {
		Success   bool                  `json:"success"`
		Listening []SystemListeningPort `json:"listening"`
	}
	if err := json.Unmarshal(rrList.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("Failed to parse /api/ports/list JSON: %v", err)
	}
	if !listResp.Success {
		t.Fatalf("Expected success: true, got %v", listResp.Success)
	}

	// 2. Test POST /api/ports/forward
	payload := PortForwardRequest{
		Port:       3000,
		LocalPort:  3000,
		Label:      "Web App",
		Protocol:   "http",
		Visibility: "private",
	}
	bodyBytes, _ := json.Marshal(payload)
	reqFwd := httptest.NewRequest("POST", "/api/ports/forward", bytes.NewReader(bodyBytes))
	rrFwd := httptest.NewRecorder()
	handlePortForward(rrFwd, reqFwd)

	if rrFwd.Code != http.StatusOK {
		t.Fatalf("Expected 200 for /api/ports/forward, got %d: %s", rrFwd.Code, rrFwd.Body.String())
	}

	// 3. Test POST /api/ports/update
	updatePayload := PortForwardRequest{
		Port:       3000,
		Label:      "Next.js App",
		Visibility: "public",
	}
	updateBytes, _ := json.Marshal(updatePayload)
	reqUp := httptest.NewRequest("POST", "/api/ports/update", bytes.NewReader(updateBytes))
	rrUp := httptest.NewRecorder()
	handlePortUpdate(rrUp, reqUp)

	if rrUp.Code != http.StatusOK {
		t.Fatalf("Expected 200 for /api/ports/update, got %d: %s", rrUp.Code, rrUp.Body.String())
	}

	// 4. Test POST /api/ports/stop
	stopPayload := PortForwardRequest{Port: 3000}
	stopBytes, _ := json.Marshal(stopPayload)
	reqStop := httptest.NewRequest("POST", "/api/ports/stop", bytes.NewReader(stopBytes))
	rrStop := httptest.NewRecorder()
	handlePortStop(rrStop, reqStop)

	if rrStop.Code != http.StatusOK {
		t.Fatalf("Expected 200 for /api/ports/stop, got %d: %s", rrStop.Code, rrStop.Body.String())
	}

	// 5. Test POST /api/ports/unforward
	unfwdPayload := PortForwardRequest{Port: 3000}
	unfwdBytes, _ := json.Marshal(unfwdPayload)
	reqUnfwd := httptest.NewRequest("POST", "/api/ports/unforward", bytes.NewReader(unfwdBytes))
	rrUnfwd := httptest.NewRecorder()
	handlePortUnforward(rrUnfwd, reqUnfwd)

	if rrUnfwd.Code != http.StatusOK {
		t.Fatalf("Expected 200 for /api/ports/unforward, got %d: %s", rrUnfwd.Code, rrUnfwd.Body.String())
	}
}

func TestPortExtractionHelpers(t *testing.T) {
	if p := extractPortFromAddr("127.0.0.1:5173"); p != 5173 {
		t.Errorf("Expected 5173, got %d", p)
	}
	if p := extractPortFromAddr("[::1]:8080"); p != 8080 {
		t.Errorf("Expected 8080, got %d", p)
	}
	if p := extractPortFromAddr("*:3306"); p != 3306 {
		t.Errorf("Expected 3306, got %d", p)
	}

	userStr := `users:(("node",pid=727442,fd=32))`
	if pid := extractPidFromUsers(userStr); pid != 727442 {
		t.Errorf("Expected pid 727442, got %d", pid)
	}
	if proc := extractProcessNameFromUsers(userStr); proc != "node" {
		t.Errorf("Expected process node, got %s", proc)
	}
}

func TestPortProbeHTTPParsing(t *testing.T) {
	// 1. Test FastAPI / Uvicorn header parsing
	respUvicorn := "HTTP/1.1 200 OK\r\nServer: uvicorn\r\nContent-Type: application/json\r\n\r\n{}"
	var infoUvicorn PortProbeInfo
	parseHTTPProbeResponse(respUvicorn, &infoUvicorn)
	if infoUvicorn.HTTPStatus != 200 {
		t.Errorf("Expected status 200, got %d", infoUvicorn.HTTPStatus)
	}
	if infoUvicorn.ServerHeader != "uvicorn" {
		t.Errorf("Expected Server uvicorn, got %s", infoUvicorn.ServerHeader)
	}
	if infoUvicorn.DetectedService != "FastAPI / Starlette (Uvicorn)" {
		t.Errorf("Expected FastAPI / Starlette (Uvicorn), got %s", infoUvicorn.DetectedService)
	}

	// 2. Test Express header parsing
	respExpress := "HTTP/1.1 404 Not Found\r\nX-Powered-By: Express\r\n\r\n"
	var infoExpress PortProbeInfo
	parseHTTPProbeResponse(respExpress, &infoExpress)
	if infoExpress.HTTPStatus != 404 {
		t.Errorf("Expected status 404, got %d", infoExpress.HTTPStatus)
	}
	if infoExpress.DetectedService != "Express (Node.js)" {
		t.Errorf("Expected Express (Node.js), got %s", infoExpress.DetectedService)
	}

	// 3. Test non-HTTP protocol detection
	var infoRedis PortProbeInfo
	detectNonHTTPProtocols("5.7.34-log\x00\x00\x00caching_sha2_password", &infoRedis)
	if infoRedis.DetectedService != "MySQL / MariaDB" {
		t.Errorf("Expected MySQL / MariaDB, got %s", infoRedis.DetectedService)
	}
}

func TestPortProbeAPI(t *testing.T) {
	// Test POST /api/ports/probe
	probePayload := map[string]interface{}{"port": 9999}
	bodyBytes, _ := json.Marshal(probePayload)
	reqProbe := httptest.NewRequest("POST", "/api/ports/probe", bytes.NewReader(bodyBytes))
	rrProbe := httptest.NewRecorder()
	handlePortProbe(rrProbe, reqProbe)

	if rrProbe.Code != http.StatusOK {
		t.Fatalf("Expected 200 for /api/ports/probe, got %d: %s", rrProbe.Code, rrProbe.Body.String())
	}

	var resp struct {
		Success bool          `json:"success"`
		Probe   PortProbeInfo `json:"probe"`
	}
	if err := json.Unmarshal(rrProbe.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse probe JSON: %v", err)
	}
	if !resp.Success || resp.Probe.Port != 9999 {
		t.Fatalf("Unexpected probe response: %+v", resp)
	}

	// Test GET /api/ports/list?probe=true
	reqListProbe := httptest.NewRequest("GET", "/api/ports/list?probe=true", nil)
	rrListProbe := httptest.NewRecorder()
	handlePortsList(rrListProbe, reqListProbe)

	if rrListProbe.Code != http.StatusOK {
		t.Fatalf("Expected 200 for /api/ports/list?probe=true, got %d", rrListProbe.Code)
	}
}

func TestWorkspaceServiceHints(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Write mock package.json
	pkgContent := `{
		"name": "my-mock-app",
		"scripts": {
			"dev": "vite --port 5173",
			"start": "next start -p 3000"
		},
		"dependencies": {
			"vite": "^5.0.0"
		}
	}`
	_ = os.WriteFile(filepath.Join(tempDir, "package.json"), []byte(pkgContent), 0644)

	// 2. Write mock docker-compose.yml
	composeContent := `version: '3.8'
services:
  db:
    image: postgres:15
    ports:
      - "5432:5432"
  redis:
    image: redis:7
    ports:
      - "6379:6379"
`
	_ = os.WriteFile(filepath.Join(tempDir, "docker-compose.yml"), []byte(composeContent), 0644)

	// 3. Write mock .env
	envContent := `PORT=8000
DATABASE_URL=postgres://localhost:5432/db
`
	_ = os.WriteFile(filepath.Join(tempDir, ".env"), []byte(envContent), 0644)

	hints := scanWorkspaceServiceHints(tempDir)
	if len(hints) < 4 {
		t.Fatalf("Expected at least 4 hints, got %d: %+v", len(hints), hints)
	}

	hintMap := make(map[int]WorkspaceServiceHint)
	for _, h := range hints {
		hintMap[h.Port] = h
	}

	if h, ok := hintMap[5173]; !ok || h.Framework != "Vite" {
		t.Errorf("Expected port 5173 with Vite framework, got %+v", h)
	}
	if h, ok := hintMap[5432]; !ok || h.Framework != "PostgreSQL" {
		t.Errorf("Expected port 5432 with PostgreSQL framework, got %+v", h)
	}
	if h, ok := hintMap[8000]; !ok || h.ServiceType != "env-var" {
		t.Errorf("Expected port 8000 from env-var, got %+v", h)
	}

	// Test GET /api/ports/workspace-hints endpoint
	req := httptest.NewRequest("GET", "/api/ports/workspace-hints", nil)
	rr := httptest.NewRecorder()
	handlePortsWorkspaceHints(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected 200 for /api/ports/workspace-hints, got %d", rr.Code)
	}
}
