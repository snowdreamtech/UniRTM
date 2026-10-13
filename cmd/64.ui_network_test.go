// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func TestUINetworkAPI(t *testing.T) {
	mockFS := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html>UniRTM</html>")},
	}
	mux := setupUIMux(mockFS)

	// 1. GET /api/network/status
	req := httptest.NewRequest("GET", "/api/network/status", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 from /api/network/status, got %d: %s", rr.Code, rr.Body.String())
	}

	var statusResp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &statusResp); err != nil {
		t.Fatalf("failed to decode status response: %v", err)
	}
	if statusResp["status"] == nil {
		t.Fatalf("expected 'status' key in response")
	}

	// 2. POST /api/network/profile/use
	useBody, _ := json.Marshal(map[string]string{
		"profile_id": "official",
	})
	req = httptest.NewRequest("POST", "/api/network/profile/use", bytes.NewReader(useBody))
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 from /api/network/profile/use, got %d: %s", rr.Code, rr.Body.String())
	}

	// 3. POST /api/network/registry/override
	overrideBody, _ := json.Marshal(map[string]string{
		"ecosystem": "npm",
		"url":       "https://test.npm.internal",
	})
	req = httptest.NewRequest("POST", "/api/network/registry/override", bytes.NewReader(overrideBody))
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 from /api/network/registry/override, got %d: %s", rr.Code, rr.Body.String())
	}

	// 4. POST /api/network/proxy
	proxyBody, _ := json.Marshal(map[string]interface{}{
		"enabled":  false,
		"http":     "",
		"no_proxy": "localhost",
	})
	req = httptest.NewRequest("POST", "/api/network/proxy", bytes.NewReader(proxyBody))
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 from /api/network/proxy, got %d: %s", rr.Code, rr.Body.String())
	}
}
