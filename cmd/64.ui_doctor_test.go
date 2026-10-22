// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleDoctorDiagnostic_Success(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/doctor", nil)
	rr := httptest.NewRecorder()

	handleDoctorDiagnostic(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Expected HTTP 200 OK")
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var report DoctorDiagnosticReport
	err := json.Unmarshal(rr.Body.Bytes(), &report)
	assert.NoError(t, err, "Response body should be valid JSON DoctorDiagnosticReport")

	// Validate top-level fields
	assert.Greater(t, report.Timestamp, int64(0), "Timestamp should be non-zero")
	assert.GreaterOrEqual(t, report.Score, 0, "Score must be >= 0")
	assert.LessOrEqual(t, report.Score, 100, "Score must be <= 100")
	assert.Contains(t, []string{"healthy", "warning", "critical"}, report.Status, "Status must be valid")

	// Validate Core status
	assert.NotEmpty(t, report.Core.TrustStatus, "Core TrustStatus should be set")
	assert.NotEmpty(t, report.Core.ShimsDir, "Core ShimsDir should not be empty")

	// Validate Build info
	assert.NotEmpty(t, report.Build.Target, "Target architecture should be present")
	assert.NotEmpty(t, report.Build.GoVersion, "GoVersion should be present")
	assert.NotEmpty(t, report.Build.Features, "Build features list should not be empty")

	// Validate Context info
	assert.NotEmpty(t, report.Context.WorkDir, "WorkDir should not be empty")
	assert.NotEmpty(t, report.Context.ActiveEnv, "ActiveEnv should not be empty")

	// Validate Directories
	assert.GreaterOrEqual(t, len(report.Directories), 4, "Should report at least 4 key directories (cache, config, data, shims)")
	dirTypes := make(map[string]bool)
	for _, d := range report.Directories {
		dirTypes[d.Type] = true
		assert.NotEmpty(t, d.Path, "Directory path should not be empty")
	}
	assert.True(t, dirTypes["cache"], "cache directory should be included")
	assert.True(t, dirTypes["data"], "data directory should be included")
	assert.True(t, dirTypes["shims"], "shims directory should be included")

	// Validate Health checks
	assert.NotEmpty(t, report.Health.DatabasePath, "DatabasePath should not be empty")
}

func TestHandleDoctorDiagnostic_PostMethodAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/doctor", nil)
	rr := httptest.NewRecorder()

	handleDoctorDiagnostic(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Expected HTTP 200 OK for POST triggering re-diagnosis")
}

func TestHandleDoctorDiagnostic_MethodNotAllowed(t *testing.T) {
	disallowedMethods := []string{http.MethodDelete, http.MethodPut, http.MethodPatch}
	for _, method := range disallowedMethods {
		req := httptest.NewRequest(method, "/api/doctor", nil)
		rr := httptest.NewRecorder()

		handleDoctorDiagnostic(rr, req)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code, "Expected HTTP 405 Method Not Allowed for %s", method)
	}
}
