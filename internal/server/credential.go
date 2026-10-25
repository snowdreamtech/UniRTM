// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// GitCredentialRequest represents an incoming passphrase/credential prompt.
type GitCredentialRequest struct {
	ID        string `json:"id"`
	Type      string `json:"type"` // "ssh" | "gpg" | "git-auth"
	Prompt    string `json:"prompt"`
	TargetKey string `json:"targetKey,omitempty"`
	Timestamp int64  `json:"timestamp"`
}

// GitCredentialResponse represents user response from the Web UI.
type GitCredentialResponse struct {
	ID              string `json:"id"`
	Passphrase      string `json:"passphrase"`
	RememberSession bool   `json:"rememberSession"`
	Canceled        bool   `json:"canceled"`
}

// sessionKeyringCache safely stores in-memory passphrases during the active session.
type sessionKeyringCache struct {
	mu    sync.RWMutex
	items map[string]string // key identifier -> passphrase
}

var globalSessionKeyring = &sessionKeyringCache{
	items: make(map[string]string),
}

func (s *sessionKeyringCache) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.items[key]
	return val, ok
}

func (s *sessionKeyringCache) Set(key, val string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[key] = val
}

func (s *sessionKeyringCache) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = make(map[string]string)
}

// credentialCoordinator manages blocking waiting channels between CLI/subprocesses and Web UI.
type credentialCoordinator struct {
	mu          sync.Mutex
	activeChan  map[string]chan GitCredentialResponse
	pendingReq  *GitCredentialRequest
	notifyCh    chan struct{}
	subscribers map[chan *GitCredentialRequest]struct{}
}

func newCredentialCoordinator() *credentialCoordinator {
	return &credentialCoordinator{
		activeChan:  make(map[string]chan GitCredentialResponse),
		notifyCh:    make(chan struct{}),
		subscribers: make(map[chan *GitCredentialRequest]struct{}),
	}
}

var globalCredentialCoordinator = newCredentialCoordinator()

func generateCredentialID() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return fmt.Sprintf("cred-%d-%s", time.Now().UnixNano()%100000000, hex.EncodeToString(b))
}

// RequestGitCredential blocks until user enters password in Web UI or timeout occurs.
func (c *credentialCoordinator) RequestGitCredential(reqType, prompt, targetKey string, timeout time.Duration) (string, error) {
	cacheKey := targetKey
	if cacheKey == "" {
		cacheKey = prompt
	}

	// 1. Check in-memory session keyring cache first
	if cachedPass, ok := globalSessionKeyring.Get(cacheKey); ok {
		return cachedPass, nil
	}

	id := generateCredentialID()
	respCh := make(chan GitCredentialResponse, 1)

	c.mu.Lock()
	c.activeChan[id] = respCh
	c.pendingReq = &GitCredentialRequest{
		ID:        id,
		Type:      reqType,
		Prompt:    prompt,
		TargetKey: targetKey,
		Timestamp: time.Now().UnixMilli(),
	}
	if c.notifyCh != nil {
		close(c.notifyCh)
		c.notifyCh = make(chan struct{})
	}
	for sub := range c.subscribers {
		select {
		case sub <- c.pendingReq:
		default:
		}
	}
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		delete(c.activeChan, id)
		if c.pendingReq != nil && c.pendingReq.ID == id {
			c.pendingReq = nil
		}
		c.mu.Unlock()
	}()

	select {
	case resp := <-respCh:
		if resp.Canceled {
			return "", errors.New("user canceled credential prompt")
		}
		if resp.RememberSession {
			globalSessionKeyring.Set(cacheKey, resp.Passphrase)
		}
		return resp.Passphrase, nil
	case <-time.After(timeout):
		return "", errors.New("credential prompt timed out waiting for user input")
	}
}

// RespondCredential dispatches user password response to the waiting process channel.
func (c *credentialCoordinator) RespondCredential(resp GitCredentialResponse) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	ch, exists := c.activeChan[resp.ID]
	if !exists {
		return false
	}
	ch <- resp
	if c.pendingReq != nil && c.pendingReq.ID == resp.ID {
		c.pendingReq = nil
	}
	return true
}

func (c *credentialCoordinator) GetPendingRequest() *GitCredentialRequest {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.pendingReq
}

// WaitForPendingRequest blocks until a pending request is available,
// the timeout expires, or the client context is cancelled.
func (c *credentialCoordinator) WaitForPendingRequest(ctx context.Context, timeout time.Duration) *GitCredentialRequest {
	c.mu.Lock()
	if c.pendingReq != nil {
		req := c.pendingReq
		c.mu.Unlock()
		return req
	}
	if c.notifyCh == nil {
		c.notifyCh = make(chan struct{})
	}
	notify := c.notifyCh
	c.mu.Unlock()

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-notify:
		c.mu.Lock()
		req := c.pendingReq
		c.mu.Unlock()
		return req
	case <-timer.C:
		c.mu.Lock()
		req := c.pendingReq
		c.mu.Unlock()
		return req
	case <-ctx.Done():
		return nil
	}
}

// Subscribe registers a live event listener for incoming Git credential requests.
func (c *credentialCoordinator) Subscribe() (chan *GitCredentialRequest, func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.subscribers == nil {
		c.subscribers = make(map[chan *GitCredentialRequest]struct{})
	}
	ch := make(chan *GitCredentialRequest, 10)
	c.subscribers[ch] = struct{}{}
	unsubscribe := func() {
		c.mu.Lock()
		delete(c.subscribers, ch)
		c.mu.Unlock()
	}
	return ch, unsubscribe
}

// HTTP API: GET /api/git/credential-stream (Server-Sent Events)
func HandleGitCredentialStream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	subCh, unsubscribe := globalCredentialCoordinator.Subscribe()
	defer unsubscribe()

	// If there is already a pending credential request, push it immediately upon connect
	if initialReq := globalCredentialCoordinator.GetPendingRequest(); initialReq != nil {
		data, _ := json.Marshal(initialReq)
		fmt.Fprintf(w, "event: credential\ndata: %s\n\n", data)
		flusher.Flush()
	}

	keepAliveTicker := time.NewTicker(15 * time.Second)
	defer keepAliveTicker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case req, open := <-subCh:
			if !open {
				return
			}
			if req != nil {
				data, _ := json.Marshal(req)
				fmt.Fprintf(w, "event: credential\ndata: %s\n\n", data)
				flusher.Flush()
			}
		case <-keepAliveTicker.C:
			fmt.Fprintf(w, ": keep-alive\n\n")
			flusher.Flush()
		}
	}
}

// HTTP API: GET /api/git/credential-pending?wait=25s
func HandleGitCredentialPending(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req *GitCredentialRequest
	waitParam := r.URL.Query().Get("wait")
	if waitParam != "" {
		waitDuration, err := time.ParseDuration(waitParam)
		if err != nil || waitDuration <= 0 {
			waitDuration = 25 * time.Second
		}
		if waitDuration > 60*time.Second {
			waitDuration = 60 * time.Second
		}
		req = globalCredentialCoordinator.WaitForPendingRequest(r.Context(), waitDuration)
	} else {
		req = globalCredentialCoordinator.GetPendingRequest()
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"request": req,
	})
}

// HTTP API: POST /api/git/credential-respond
func HandleGitCredentialRespond(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var resp GitCredentialResponse
	if err := json.NewDecoder(r.Body).Decode(&resp); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "invalid JSON payload",
		})
		return
	}

	ok := globalCredentialCoordinator.RespondCredential(resp)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": ok,
	})
}

// HTTP API: POST /api/git/credential-clear-cache
func HandleGitCredentialClearCache(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	globalSessionKeyring.Clear()
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

// NonceManager handles single-use cryptographic tokens for AskPass IPC security.
type NonceManager struct {
	mu          sync.Mutex
	validNonces map[string]time.Time
}

var globalNonceManager = &NonceManager{
	validNonces: make(map[string]time.Time),
}

// IssueCredentialNonce creates a cryptographically random single-use token valid for 2 minutes.
func IssueCredentialNonce() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	nonce := hex.EncodeToString(b)

	globalNonceManager.mu.Lock()
	defer globalNonceManager.mu.Unlock()
	globalNonceManager.validNonces[nonce] = time.Now().Add(2 * time.Minute)
	return nonce
}

// VerifyAndConsumeNonce checks if the nonce is valid and burns it immediately (Single-use).
func VerifyAndConsumeNonce(nonce string) bool {
	if nonce == "" {
		return false
	}
	globalNonceManager.mu.Lock()
	defer globalNonceManager.mu.Unlock()

	exp, ok := globalNonceManager.validNonces[nonce]
	if !ok {
		return false
	}
	delete(globalNonceManager.validNonces, nonce)
	return time.Now().Before(exp)
}

// CredentialAuditRecord stores security transparency logs for credential accesses.
type CredentialAuditRecord struct {
	ID        string `json:"id"`
	Timestamp int64  `json:"timestamp"`
	Action    string `json:"action"`
	Type      string `json:"type"`
	TargetKey string `json:"targetKey"`
	Result    string `json:"result"` // "success" | "cached" | "denied_invalid_nonce" | "timeout" | "canceled"
	ClientIP  string `json:"clientIp"`
}

type credentialAuditTrail struct {
	mu      sync.RWMutex
	records []CredentialAuditRecord
}

var globalAuditTrail = &credentialAuditTrail{
	records: make([]CredentialAuditRecord, 0, 100),
}

func (a *credentialAuditTrail) Record(rec CredentialAuditRecord) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if len(a.records) >= 100 {
		a.records = a.records[1:]
	}
	a.records = append(a.records, rec)
}

func (a *credentialAuditTrail) GetAll() []CredentialAuditRecord {
	a.mu.RLock()
	defer a.mu.RUnlock()
	res := make([]CredentialAuditRecord, len(a.records))
	copy(res, a.records)
	return res
}

func (a *credentialAuditTrail) Clear() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.records = make([]CredentialAuditRecord, 0, 100)
}

// NativeAgentStatus provides diagnostics on host ssh-agent and gpg-agent.
type NativeAgentStatus struct {
	SSHAgentActive bool   `json:"sshAgentActive"`
	SSHAgentSock   string `json:"sshAgentSock,omitempty"`
	SSHAgentKeys   int    `json:"sshAgentKeys"`
	GPGAgentActive bool   `json:"gpgAgentActive"`
}

// DetectNativeAgents probes the operating system for active key agents.
func DetectNativeAgents() NativeAgentStatus {
	status := NativeAgentStatus{}

	// 1. SSH Agent probe
	sock := os.Getenv("SSH_AUTH_SOCK")
	if sock != "" {
		status.SSHAgentActive = true
		// Mask path for privacy
		status.SSHAgentSock = filepath.Base(sock)
		// Check loaded keys via ssh-add -l
		cmd := exec.Command("ssh-add", "-l")
		out, err := cmd.CombinedOutput()
		if err == nil {
			lines := strings.Split(strings.TrimSpace(string(out)), "\n")
			status.SSHAgentKeys = len(lines)
		}
	}

	// 2. GPG Agent probe
	gpgCmd := exec.Command("gpg-connect-agent", "/bye")
	if err := gpgCmd.Run(); err == nil {
		status.GPGAgentActive = true
	}

	return status
}

var (
	askpassHelperPath string
	askpassHelperOnce sync.Once
)

// ActivePort stores the bound listening port of the Workbench server.
var ActivePort int = 5173

// GetOrCreateAskPassHelper creates a standalone lightweight POSIX shell script
// that relays Git / SSH passphrase prompts to UniRTM's broker HTTP endpoint with Nonce header.
func GetOrCreateAskPassHelper() string {
	askpassHelperOnce.Do(func() {
		tmpDir := os.TempDir()
		scriptFile := filepath.Join(tmpDir, fmt.Sprintf("unirtm-askpass-%d.sh", os.Getpid()))

		targetPort := ActivePort
		if targetPort <= 0 {
			targetPort = 5173
		}

		content := fmt.Sprintf(`#!/bin/sh
PROMPT="$*"
NONCE="${UNIRTM_CREDENTIAL_NONCE:-}"
if command -v curl >/dev/null 2>&1; then
    curl -s -X POST "http://127.0.0.1:%d/api/git/credential-internal-ask" \
         -H "Content-Type: text/plain" \
         -H "X-Unirtm-Nonce: $NONCE" \
         --data-binary "$PROMPT"
elif command -v wget >/dev/null 2>&1; then
    wget -qO- --header="X-Unirtm-Nonce: $NONCE" --post-data="$PROMPT" "http://127.0.0.1:%d/api/git/credential-internal-ask"
fi
`, targetPort, targetPort)

		if err := os.WriteFile(scriptFile, []byte(content), 0700); err == nil {
			askpassHelperPath = scriptFile
		}
	})
	return askpassHelperPath
}

// HTTP API: POST /api/git/credential-internal-ask
// Invoked by the askpass helper script when Git or SSH prompts for password.
func HandleGitCredentialInternalAsk(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	nonce := r.Header.Get("X-Unirtm-Nonce")
	clientIP := r.RemoteAddr

	// Verify Nonce if present or if strictly enforced
	strictNonce := os.Getenv("UNIRTM_CREDENTIAL_NONCE_STRICT") != "false"
	if strictNonce && nonce != "" {
		if !VerifyAndConsumeNonce(nonce) {
			globalAuditTrail.Record(CredentialAuditRecord{
				ID:        fmt.Sprintf("audit-%d", time.Now().UnixNano()),
				Timestamp: time.Now().UnixMilli(),
				Action:    "askpass_prompt",
				Type:      "security_violation",
				TargetKey: "N/A",
				Result:    "denied_invalid_nonce",
				ClientIP:  clientIP,
			})
			http.Error(w, "Forbidden: invalid or expired credential nonce", http.StatusForbidden)
			return
		}
	}

	promptBytes, _ := io.ReadAll(r.Body)
	prompt := strings.TrimSpace(string(promptBytes))

	// Detect whether prompt is for SSH, GPG, or HTTPS auth
	reqType := "ssh"
	targetKey := ""

	lowerPrompt := strings.ToLower(prompt)
	if strings.Contains(lowerPrompt, "gpg") || strings.Contains(lowerPrompt, "pinentry") {
		reqType = "gpg"
	} else if strings.Contains(lowerPrompt, "username") || strings.Contains(lowerPrompt, "password for 'https") {
		reqType = "git-auth"
	}

	// Extract key path from prompt e.g. "Enter passphrase for key '/root/.ssh/id_rsa':"
	if idx := strings.Index(prompt, "'"); idx != -1 {
		rest := prompt[idx+1:]
		if endIdx := strings.Index(rest, "'"); endIdx != -1 {
			targetKey = rest[:endIdx]
		}
	}

	pass, err := globalCredentialCoordinator.RequestGitCredential(reqType, prompt, targetKey, 60*time.Second)
	if err != nil {
		globalAuditTrail.Record(CredentialAuditRecord{
			ID:        fmt.Sprintf("audit-%d", time.Now().UnixNano()),
			Timestamp: time.Now().UnixMilli(),
			Action:    "askpass_prompt",
			Type:      reqType,
			TargetKey: targetKey,
			Result:    "timeout_or_canceled",
			ClientIP:  clientIP,
		})
		w.WriteHeader(http.StatusRequestTimeout)
		return
	}

	globalAuditTrail.Record(CredentialAuditRecord{
		ID:        fmt.Sprintf("audit-%d", time.Now().UnixNano()),
		Timestamp: time.Now().UnixMilli(),
		Action:    "askpass_prompt",
		Type:      reqType,
		TargetKey: targetKey,
		Result:    "success",
		ClientIP:  clientIP,
	})

	w.Header().Set("Content-Type", "text/plain")
	_, _ = w.Write([]byte(pass))
}

// HTTP API: GET /api/git/agent-status
func HandleGitAgentStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	status := DetectNativeAgents()
	_ = json.NewEncoder(w).Encode(status)
}

// HTTP API: GET /api/git/credential-audit
func HandleGitCredentialAudit(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	records := globalAuditTrail.GetAll()
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"records": records,
	})
}

// RegisterCredentialRoutes registers Git credential broker endpoints.
func RegisterCredentialRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/credentials/stream", HandleGitCredentialStream)
	mux.HandleFunc("/api/credentials/pending", HandleGitCredentialPending)
	mux.HandleFunc("/api/credentials/respond", HandleGitCredentialRespond)
	mux.HandleFunc("/api/credentials/clear", HandleGitCredentialClearCache)
	mux.HandleFunc("/api/credentials/ask", HandleGitCredentialInternalAsk)
	mux.HandleFunc("/api/git/credential-stream", HandleGitCredentialStream)
	mux.HandleFunc("/api/git/credential-pending", HandleGitCredentialPending)
	mux.HandleFunc("/api/git/credential-respond", HandleGitCredentialRespond)
	mux.HandleFunc("/api/git/credential-clear-cache", HandleGitCredentialClearCache)
	mux.HandleFunc("/api/git/credential-internal-ask", HandleGitCredentialInternalAsk)
	mux.HandleFunc("/api/git/agent-status", HandleGitAgentStatus)
	mux.HandleFunc("/api/git/credential-audit", HandleGitCredentialAudit)
}
