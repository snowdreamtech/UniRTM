// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
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
	mu         sync.Mutex
	activeChan map[string]chan GitCredentialResponse
	pendingReq *GitCredentialRequest
}

var globalCredentialCoordinator = &credentialCoordinator{
	activeChan: make(map[string]chan GitCredentialResponse),
}

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

// HTTP API: GET /api/git/credential-pending
func handleGitCredentialPending(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	req := globalCredentialCoordinator.GetPendingRequest()
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"request": req,
	})
}

// HTTP API: POST /api/git/credential-respond
func handleGitCredentialRespond(w http.ResponseWriter, r *http.Request) {
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
func handleGitCredentialClearCache(w http.ResponseWriter, r *http.Request) {
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

var (
	askpassHelperPath string
	askpassHelperOnce sync.Once
)

// getOrCreateAskPassHelper creates a standalone lightweight POSIX shell script
// that relays Git / SSH passphrase prompts to UniRTM's broker HTTP endpoint.
func getOrCreateAskPassHelper() string {
	askpassHelperOnce.Do(func() {
		tmpDir := os.TempDir()
		scriptFile := filepath.Join(tmpDir, fmt.Sprintf("unirtm-askpass-%d.sh", os.Getpid()))

		targetPort := uiPort
		if targetPort <= 0 {
			targetPort = 5173
		}

		content := fmt.Sprintf(`#!/bin/sh
PROMPT="$*"
if command -v curl >/dev/null 2>&1; then
    curl -s -X POST "http://127.0.0.1:%d/api/git/credential-internal-ask" -H "Content-Type: text/plain" --data-binary "$PROMPT"
elif command -v wget >/dev/null 2>&1; then
    wget -qO- --post-data="$PROMPT" "http://127.0.0.1:%d/api/git/credential-internal-ask"
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
func handleGitCredentialInternalAsk(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
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
		w.WriteHeader(http.StatusRequestTimeout)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	_, _ = w.Write([]byte(pass))
}

