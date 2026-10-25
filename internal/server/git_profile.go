// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package server

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/snowdreamtech/unirtm/internal/pkg/env"
)

// GitProfile represents an isolated Git identity configuration.
type GitProfile struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	UserName   string `json:"userName"`
	UserEmail  string `json:"userEmail"`
	SSHKeyPath string `json:"sshKeyPath,omitempty"` // Local path to SSH key, applied via core.sshCommand
	SigningKey string `json:"signingKey,omitempty"` // GPG Key ID / Fingerprint
	GPGSign    bool   `json:"gpgSign"`              // Whether to sign commits by default
	IsDefault  bool   `json:"isDefault"`
	UpdatedAt  int64  `json:"updatedAt"`
}

// GitProfileStore wraps the in-memory and persistent storage of Git profiles.
type GitProfileStore struct {
	ActiveProfileID string       `json:"activeProfileId"`
	Profiles        []GitProfile `json:"profiles"`
}

// SystemSSHKey represents an enumerated public key on the local system.
type SystemSSHKey struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Algorithm string `json:"algorithm"`
	Comment   string `json:"comment"`
}

// SystemGPGKey represents an enumerated GPG public key.
type SystemGPGKey struct {
	KeyID   string `json:"keyId"`
	UserID  string `json:"userId"`
	Algo    string `json:"algo,omitempty"`
	Created string `json:"created,omitempty"`
}

// SystemKeysSummary bundles detected SSH and GPG keys.
type SystemKeysSummary struct {
	SSHKeys []SystemSSHKey `json:"sshKeys"`
	GPGKeys []SystemGPGKey `json:"gpgKeys"`
}

// GitProfileManager coordinates identity profiles and local repository configuration.
type GitProfileManager struct {
	mu              sync.RWMutex
	storeFilePath   string
	profiles        map[string]GitProfile
	activeProfileID string
}

func newGitProfileManager() *GitProfileManager {
	filePath := filepath.Join(env.GetConfigDir(), "git_profiles.json")
	m := &GitProfileManager{
		storeFilePath: filePath,
		profiles:      make(map[string]GitProfile),
	}
	_ = m.load()
	return m
}

var globalGitProfileManager = newGitProfileManager()

func (m *GitProfileManager) load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.storeFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Populate default profile if none exists
			defaultProfile := GitProfile{
				ID:        "default",
				Name:      "Default Profile",
				UserName:  "",
				UserEmail: "",
				GPGSign:   false,
				IsDefault: true,
				UpdatedAt: time.Now().UnixMilli(),
			}
			m.profiles["default"] = defaultProfile
			m.activeProfileID = "default"
		}
		return nil
	}

	var store GitProfileStore
	if err := json.Unmarshal(data, &store); err != nil {
		return err
	}

	m.profiles = make(map[string]GitProfile)
	for _, p := range store.Profiles {
		m.profiles[p.ID] = p
	}
	m.activeProfileID = store.ActiveProfileID
	if m.activeProfileID == "" && len(store.Profiles) > 0 {
		m.activeProfileID = store.Profiles[0].ID
	}
	return nil
}

func (m *GitProfileManager) saveLocked() error {
	var store GitProfileStore
	store.ActiveProfileID = m.activeProfileID
	for _, p := range m.profiles {
		store.Profiles = append(store.Profiles, p)
	}

	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(m.storeFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(m.storeFilePath, data, 0600)
}

// GetProfiles returns a snapshot of all stored profiles and active ID.
func (m *GitProfileManager) GetProfiles() ([]GitProfile, string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := make([]GitProfile, 0, len(m.profiles))
	for _, p := range m.profiles {
		list = append(list, p)
	}
	return list, m.activeProfileID
}

// SaveProfile saves or updates a profile.
func (m *GitProfileManager) SaveProfile(profile GitProfile) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if profile.ID == "" {
		profile.ID = fmt.Sprintf("profile-%d", time.Now().UnixNano())
	}
	profile.UpdatedAt = time.Now().UnixMilli()
	if profile.IsDefault {
		for id, p := range m.profiles {
			if id != profile.ID && p.IsDefault {
				p.IsDefault = false
				m.profiles[id] = p
			}
		}
	}

	m.profiles[profile.ID] = profile
	if m.activeProfileID == "" || profile.IsDefault {
		m.activeProfileID = profile.ID
	}

	return m.saveLocked()
}

// DeleteProfile removes a profile.
func (m *GitProfileManager) DeleteProfile(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.profiles, id)
	if m.activeProfileID == id {
		m.activeProfileID = ""
		for pid := range m.profiles {
			m.activeProfileID = pid
			break
		}
	}
	return m.saveLocked()
}

// SetActiveProfile switches active profile.
func (m *GitProfileManager) SetActiveProfile(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.profiles[id]; !exists {
		return fmt.Errorf("profile not found: %s", id)
	}
	m.activeProfileID = id
	return m.saveLocked()
}

// ApplyProfileToWorkspace strictly modifies only LOCAL .git/config, preventing any global pollution.
func (m *GitProfileManager) ApplyProfileToWorkspace(workspaceDir string, profileID string) error {
	m.mu.RLock()
	profile, exists := m.profiles[profileID]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("profile with ID %q not found", profileID)
	}

	if workspaceDir == "" {
		workspaceDir = "."
	}

	// Verify local git repo exists
	gitDir := filepath.Join(workspaceDir, ".git")
	if fi, err := os.Stat(gitDir); err != nil || (!fi.IsDir() && !strings.HasSuffix(fi.Name(), ".git")) {
		return fmt.Errorf("not a git repository: %s", workspaceDir)
	}

	// 1. user.name
	if profile.UserName != "" {
		if out, err := exec.Command("git", "-C", workspaceDir, "config", "--local", "user.name", profile.UserName).CombinedOutput(); err != nil {
			return fmt.Errorf("failed to set local user.name: %s (%w)", strings.TrimSpace(string(out)), err)
		}
	}

	// 2. user.email
	if profile.UserEmail != "" {
		if out, err := exec.Command("git", "-C", workspaceDir, "config", "--local", "user.email", profile.UserEmail).CombinedOutput(); err != nil {
			return fmt.Errorf("failed to set local user.email: %s (%w)", strings.TrimSpace(string(out)), err)
		}
	}

	// 3. GPG signing key and commit.gpgsign
	if profile.SigningKey != "" {
		_ = exec.Command("git", "-C", workspaceDir, "config", "--local", "user.signingkey", profile.SigningKey).Run()
		signVal := "false"
		if profile.GPGSign {
			signVal = "true"
		}
		_ = exec.Command("git", "-C", workspaceDir, "config", "--local", "commit.gpgsign", signVal).Run()
	} else {
		_ = exec.Command("git", "-C", workspaceDir, "config", "--local", "--unset", "user.signingkey").Run()
		_ = exec.Command("git", "-C", workspaceDir, "config", "--local", "commit.gpgsign", "false").Run()
	}

	// 4. SSH key isolation via core.sshCommand (Zero global pollution, strict identities only)
	if profile.SSHKeyPath != "" {
		// Expand ~ if present
		expandedPath := profile.SSHKeyPath
		if strings.HasPrefix(expandedPath, "~/") {
			if home, err := os.UserHomeDir(); err == nil {
				expandedPath = filepath.Join(home, expandedPath[2:])
			}
		}
		sshCmd := fmt.Sprintf("ssh -i %s -o IdentitiesOnly=yes", expandedPath)
		if out, err := exec.Command("git", "-C", workspaceDir, "config", "--local", "core.sshCommand", sshCmd).CombinedOutput(); err != nil {
			return fmt.Errorf("failed to set local core.sshCommand: %s (%w)", strings.TrimSpace(string(out)), err)
		}
	} else {
		// Unset local sshCommand to fallback to default
		_ = exec.Command("git", "-C", workspaceDir, "config", "--local", "--unset", "core.sshCommand").Run()
	}

	_ = m.SetActiveProfile(profileID)
	return nil
}

// DiscoverSystemKeys scans only PUBLIC keys and GPG public metadata, strictly adhering to Zero Private Key Touch.
func DiscoverSystemKeys() (*SystemKeysSummary, error) {
	summary := &SystemKeysSummary{
		SSHKeys: make([]SystemSSHKey, 0),
		GPGKeys: make([]SystemGPGKey, 0),
	}

	// 1. Scan SSH Public keys only (~/.ssh/*.pub)
	if home, err := os.UserHomeDir(); err == nil {
		sshDir := filepath.Join(home, ".ssh")
		entries, _ := os.ReadDir(sshDir)
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".pub") {
				pubPath := filepath.Join(sshDir, e.Name())
				content, err := os.ReadFile(pubPath)
				if err == nil {
					line := strings.TrimSpace(string(content))
					parts := strings.Fields(line)
					algo := ""
					comment := ""
					if len(parts) >= 1 {
						algo = parts[0]
					}
					if len(parts) >= 3 {
						comment = strings.Join(parts[2:], " ")
					}
					// Corresponding private key path (without .pub)
					keyPath := strings.TrimSuffix(pubPath, ".pub")
					summary.SSHKeys = append(summary.SSHKeys, SystemSSHKey{
						Name:      e.Name(),
						Path:      keyPath,
						Algorithm: algo,
						Comment:   comment,
					})
				}
			}
		}
	}

	// 2. Discover GPG public keys via gpg command (zero touching of secret key contents)
	if gpgPath, err := exec.LookPath("gpg"); err == nil {
		cmd := exec.Command(gpgPath, "--list-public-keys", "--with-colons")
		out, err := cmd.Output()
		if err == nil {
			scanner := bufio.NewScanner(bytes.NewReader(out))
			var currentKeyID string
			for scanner.Scan() {
				line := scanner.Text()
				fields := strings.Split(line, ":")
				if len(fields) < 5 {
					continue
				}
				recordType := fields[0]
				if recordType == "pub" || recordType == "sec" {
					currentKeyID = fields[4]
				} else if recordType == "fpr" && len(fields) >= 10 && fields[9] != "" {
					currentKeyID = fields[9]
				} else if recordType == "uid" && currentKeyID != "" {
					uid := fields[9]
					summary.GPGKeys = append(summary.GPGKeys, SystemGPGKey{
						KeyID:  currentKeyID,
						UserID: uid,
					})
				}
			}
		}
	}

	return summary, nil
}

// RegisterGitProfileRoutes mounts the REST API endpoints.
func RegisterGitProfileRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/git/profiles", handleGitProfiles)
	mux.HandleFunc("/api/git/profile/apply", handleGitProfileApply)
	mux.HandleFunc("/api/git/system-keys", handleGitSystemKeys)
}

func handleGitProfiles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		profiles, activeID := globalGitProfileManager.GetProfiles()
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"profiles":        profiles,
			"activeProfileId": activeID,
		})
	case http.MethodPost:
		var profile GitProfile
		if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
			http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
			return
		}
		if profile.Name == "" {
			http.Error(w, `{"error":"profile name cannot be empty"}`, http.StatusBadRequest)
			return
		}
		if err := globalGitProfileManager.SaveProfile(profile); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"profile": profile,
		})
	case http.MethodDelete:
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, `{"error":"missing profile id parameter"}`, http.StatusBadRequest)
			return
		}
		if err := globalGitProfileManager.DeleteProfile(id); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
		})
	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func handleGitProfileApply(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		WorkspaceDir string `json:"workspaceDir"`
		ProfileID    string `json:"profileId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}

	if req.ProfileID == "" {
		http.Error(w, `{"error":"profileId is required"}`, http.StatusBadRequest)
		return
	}

	if err := globalGitProfileManager.ApplyProfileToWorkspace(req.WorkspaceDir, req.ProfileID); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusBadRequest)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"profileId": req.ProfileID,
	})
}

func handleGitSystemKeys(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	summary, err := DiscoverSystemKeys()
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(summary)
}
