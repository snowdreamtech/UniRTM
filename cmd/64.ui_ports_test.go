// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
