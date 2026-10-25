// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSessionKeyringCache(t *testing.T) {
	keyring := &sessionKeyringCache{
		items: make(map[string]string),
	}

	// 1. Initial state
	_, ok := keyring.Get("id_rsa")
	assert.False(t, ok)

	// 2. Set and Get
	keyring.Set("id_rsa", "mypass123")
	val, ok := keyring.Get("id_rsa")
	assert.True(t, ok)
	assert.Equal(t, "mypass123", val)

	// 3. Clear
	keyring.Clear()
	_, ok = keyring.Get("id_rsa")
	assert.False(t, ok)
}

func TestCredentialCoordinator_Flow(t *testing.T) {
	coord := &credentialCoordinator{
		activeChan: make(map[string]chan GitCredentialResponse),
	}
	globalSessionKeyring.Clear()

	// 1. Success Flow with Response
	done := make(chan struct{})
	go func() {
		// Wait slightly until pendingReq is registered
		for i := 0; i < 20; i++ {
			req := coord.GetPendingRequest()
			if req != nil && req.Type == "ssh" {
				ok := coord.RespondCredential(GitCredentialResponse{
					ID:              req.ID,
					Passphrase:      "secret_ssh_pass",
					RememberSession: true,
					Canceled:        false,
				})
				assert.True(t, ok)
				close(done)
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	pass, err := coord.RequestGitCredential("ssh", "Enter passphrase for id_ed25519:", "/home/user/.ssh/id_ed25519", 2*time.Second)
	assert.NoError(t, err)
	assert.Equal(t, "secret_ssh_pass", pass)
	<-done

	// 2. Cache Hit Flow: Next request for same key should hit cache immediately
	passCached, errCached := coord.RequestGitCredential("ssh", "Enter passphrase for id_ed25519:", "/home/user/.ssh/id_ed25519", 100*time.Millisecond)
	assert.NoError(t, errCached)
	assert.Equal(t, "secret_ssh_pass", passCached)

	// 3. User Cancellation Flow
	globalSessionKeyring.Clear()
	go func() {
		for i := 0; i < 20; i++ {
			req := coord.GetPendingRequest()
			if req != nil {
				coord.RespondCredential(GitCredentialResponse{
					ID:       req.ID,
					Canceled: true,
				})
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	_, cancelErr := coord.RequestGitCredential("gpg", "Enter GPG Passphrase:", "0x9876", 2*time.Second)
	assert.Error(t, cancelErr)
	assert.Contains(t, cancelErr.Error(), "canceled")

	// 4. Timeout Flow
	_, timeoutErr := coord.RequestGitCredential("git-auth", "Enter password for git:", "https://github.com", 50*time.Millisecond)
	assert.Error(t, timeoutErr)
	assert.Contains(t, timeoutErr.Error(), "timed out")
}

func TestCredentialHTTPHandlers(t *testing.T) {
	globalSessionKeyring.Clear()

	// 1. GET /api/git/credential-pending when empty
	req := httptest.NewRequest(http.MethodGet, "/api/git/credential-pending", nil)
	rec := httptest.NewRecorder()
	HandleGitCredentialPending(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var pendingResp map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &pendingResp)
	assert.NoError(t, err)
	assert.True(t, pendingResp["success"].(bool))
	assert.Nil(t, pendingResp["request"])

	// 2. POST /api/git/credential-clear-cache
	clearReq := httptest.NewRequest(http.MethodPost, "/api/git/credential-clear-cache", nil)
	clearRec := httptest.NewRecorder()
	HandleGitCredentialClearCache(clearRec, clearReq)
	assert.Equal(t, http.StatusOK, clearRec.Code)

	// 3. POST /api/git/credential-respond with non-existent ID
	bodyBytes, _ := json.Marshal(GitCredentialResponse{
		ID:         "invalid-id",
		Passphrase: "test",
	})
	respReq := httptest.NewRequest(http.MethodPost, "/api/git/credential-respond", bytes.NewReader(bodyBytes))
	respRec := httptest.NewRecorder()
	HandleGitCredentialRespond(respRec, respReq)
	assert.Equal(t, http.StatusOK, respRec.Code)
	var respData map[string]interface{}
	_ = json.Unmarshal(respRec.Body.Bytes(), &respData)
	assert.False(t, respData["success"].(bool))

	// 4. Test GetOrCreateAskPassHelper
	helperPath := GetOrCreateAskPassHelper()
	assert.NotEmpty(t, helperPath)

	// 5. POST /api/git/credential-internal-ask with cache hit
	globalSessionKeyring.Set("/home/user/.ssh/id_rsa", "auto_cached_secret")
	internalReq := httptest.NewRequest(
		http.MethodPost,
		"/api/git/credential-internal-ask",
		bytes.NewBufferString("Enter passphrase for key '/home/user/.ssh/id_rsa':"),
	)
	internalRec := httptest.NewRecorder()
	HandleGitCredentialInternalAsk(internalRec, internalReq)
	assert.Equal(t, http.StatusOK, internalRec.Code)
	assert.Equal(t, "auto_cached_secret", internalRec.Body.String())
}

func TestCredentialLongPolling(t *testing.T) {
	globalSessionKeyring.Clear()

	// 1. Long-polling timeout test (wait=40ms)
	start := time.Now()
	req := httptest.NewRequest(http.MethodGet, "/api/git/credential-pending?wait=40ms", nil)
	rec := httptest.NewRecorder()
	HandleGitCredentialPending(rec, req)
	elapsed := time.Since(start)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, elapsed >= 35*time.Millisecond)

	var timeoutResp map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &timeoutResp)
	assert.NoError(t, err)
	assert.True(t, timeoutResp["success"].(bool))
	assert.Nil(t, timeoutResp["request"])

	// 2. Long-polling instant wake test when request arrives
	woken := make(chan struct{})
	go func() {
		reqWake := httptest.NewRequest(http.MethodGet, "/api/git/credential-pending?wait=2s", nil)
		recWake := httptest.NewRecorder()
		HandleGitCredentialPending(recWake, reqWake)

		assert.Equal(t, http.StatusOK, recWake.Code)
		var wakeResp map[string]interface{}
		_ = json.Unmarshal(recWake.Body.Bytes(), &wakeResp)
		assert.True(t, wakeResp["success"].(bool))
		reqMap := wakeResp["request"].(map[string]interface{})
		assert.Equal(t, "instant-test", reqMap["id"])
		close(woken)
	}()

	time.Sleep(30 * time.Millisecond)
	globalCredentialCoordinator.mu.Lock()
	globalCredentialCoordinator.pendingReq = &GitCredentialRequest{
		ID:        "instant-test",
		Type:      "ssh",
		Prompt:    "Enter passphrase:",
		Timestamp: time.Now().UnixMilli(),
	}
	if globalCredentialCoordinator.notifyCh != nil {
		close(globalCredentialCoordinator.notifyCh)
		globalCredentialCoordinator.notifyCh = make(chan struct{})
	}
	globalCredentialCoordinator.mu.Unlock()

	select {
	case <-woken:
		// Succeeded immediately on event
	case <-time.After(1 * time.Second):
		t.Fatal("Long polling did not wake up promptly upon request arrival")
	}

	// Clean up pendingReq
	globalCredentialCoordinator.mu.Lock()
	globalCredentialCoordinator.pendingReq = nil
	globalCredentialCoordinator.mu.Unlock()
}

func TestCredentialStream(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req := httptest.NewRequest(http.MethodGet, "/api/git/credential-stream", nil).WithContext(ctx)
	rec := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		HandleGitCredentialStream(rec, req)
		close(done)
	}()

	// Wait for handler to subscribe
	time.Sleep(30 * time.Millisecond)

	// Simulate credential prompt arrival
	go func() {
		_, _ = globalCredentialCoordinator.RequestGitCredential("gpg", "Enter GPG Passphrase:", "0x5678", 500*time.Millisecond)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel() // cancel stream connection
	<-done

	body := rec.Body.String()
	assert.Contains(t, body, "event: credential")
	assert.Contains(t, body, "0x5678")
}

func TestCredentialNonceManager(t *testing.T) {
	nonce := IssueCredentialNonce()
	assert.NotEmpty(t, nonce)

	// First verify and consume should succeed
	ok := VerifyAndConsumeNonce(nonce)
	assert.True(t, ok)

	// Second verify should fail (Single-use burned)
	okAgain := VerifyAndConsumeNonce(nonce)
	assert.False(t, okAgain)

	// Empty nonce should fail
	assert.False(t, VerifyAndConsumeNonce(""))
}

func TestCredentialAuditTrail(t *testing.T) {
	globalAuditTrail.Clear()

	rec := CredentialAuditRecord{
		ID:        "audit-test-1",
		Timestamp: time.Now().UnixMilli(),
		Action:    "ssh_push",
		Type:      "ssh",
		TargetKey: "/home/user/.ssh/id_rsa",
		Result:    "success",
		ClientIP:  "127.0.0.1",
	}
	globalAuditTrail.Record(rec)

	records := globalAuditTrail.GetAll()
	assert.Len(t, records, 1)
	assert.Equal(t, "audit-test-1", records[0].ID)
	assert.Equal(t, "success", records[0].Result)

	// HTTP endpoint test
	req := httptest.NewRequest(http.MethodGet, "/api/git/credential-audit", nil)
	w := httptest.NewRecorder()
	HandleGitCredentialAudit(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "audit-test-1")
}

func TestNativeAgentStatusEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/git/agent-status", nil)
	w := httptest.NewRecorder()
	HandleGitAgentStatus(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var status NativeAgentStatus
	err := json.Unmarshal(w.Body.Bytes(), &status)
	assert.NoError(t, err)
}

func TestCredentialAskPass_NonceSecurity(t *testing.T) {
	// Set strict nonce mode
	t.Setenv("UNIRTM_CREDENTIAL_NONCE_STRICT", "true")

	// 1. Forbidden without valid nonce
	reqBad := httptest.NewRequest(http.MethodPost, "/api/git/credential-internal-ask", bytes.NewReader([]byte("Enter passphrase:")))
	reqBad.Header.Set("X-Unirtm-Nonce", "malicious_fake_nonce")
	wBad := httptest.NewRecorder()
	HandleGitCredentialInternalAsk(wBad, reqBad)
	assert.Equal(t, http.StatusForbidden, wBad.Code)

	// Verify that the security violation was logged to AuditTrail
	records := globalAuditTrail.GetAll()
	assert.True(t, len(records) > 0)
	lastRecord := records[len(records)-1]
	assert.Equal(t, "denied_invalid_nonce", lastRecord.Result)
}



