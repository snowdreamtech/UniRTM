// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUIGuard_Endpoints(t *testing.T) {
	// 1. Test GET /api/guard/rules
	reqRules := httptest.NewRequest("GET", "/api/guard/rules", nil)
	rrRules := httptest.NewRecorder()
	handleGuardRules(rrRules, reqRules)

	if rrRules.Code != http.StatusOK {
		t.Fatalf("Expected 200 on /api/guard/rules, got %d", rrRules.Code)
	}

	var rulesResp struct {
		BuiltinRules []struct {
			ID string `json:"id"`
		} `json:"builtin_rules"`
	}
	if err := json.Unmarshal(rrRules.Body.Bytes(), &rulesResp); err != nil {
		t.Fatalf("Failed to decode rules response: %v", err)
	}
	if len(rulesResp.BuiltinRules) < 8 {
		t.Errorf("Expected at least 8 builtin rules, got %d", len(rulesResp.BuiltinRules))
	}

	// 2. Test GET /api/guard/status?mode=changed
	reqStatus := httptest.NewRequest("GET", "/api/guard/status?mode=changed", nil)
	rrStatus := httptest.NewRecorder()
	handleGuardStatus(rrStatus, reqStatus)

	if rrStatus.Code != http.StatusOK {
		t.Fatalf("Expected 200 on /api/guard/status, got %d", rrStatus.Code)
	}

	// 3. Test POST /api/guard/run with empty files
	bodyRun := strings.NewReader(`{"mode":"staged","fix":false,"files":[]}`)
	reqRun := httptest.NewRequest("POST", "/api/guard/run", bodyRun)
	rrRun := httptest.NewRecorder()
	handleGuardRun(rrRun, reqRun)

	if rrRun.Code != http.StatusOK {
		t.Fatalf("Expected 200 on /api/guard/run, got %d", rrRun.Code)
	}
}
