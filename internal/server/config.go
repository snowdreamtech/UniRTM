// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
)

func handleGetConfig(w http.ResponseWriter, r *http.Request) {
	userPath := config.GetGlobalConfigPath()
	userExists := false
	var userRaw string
	var userCfg *config.Config
	if data, err := os.ReadFile(userPath); err == nil {
		userExists = true
		userRaw = string(data)
	}
	userCfg, _ = config.LoadGlobal()
	if userCfg == nil {
		userCfg = &config.Config{}
	}

	wsPath := ".unirtm.toml"
	wsExists := false
	var wsRaw string
	if data, err := os.ReadFile(wsPath); err == nil {
		wsExists = true
		wsRaw = string(data)
	}
	cm := config.NewConfigManager()
	wsCfg, _ := cm.LoadHierarchy(r.Context())
	if wsCfg == nil {
		wsCfg = &config.Config{}
	}
	if wsCfg.MinVersion == "" && wsRaw != "" {
		var doc struct {
			MinVersion string `toml:"min_version"`
		}
		if err := toml.Unmarshal([]byte(wsRaw), &doc); err == nil && doc.MinVersion != "" {
			wsCfg.MinVersion = doc.MinVersion
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"user": map[string]interface{}{
			"path":        userPath,
			"exists":      userExists,
			"raw":         userRaw,
			"min_version": userCfg.MinVersion,
			"settings":    userCfg.Settings,
			"plugins":     userCfg.Plugins,
			"env":         userCfg.Env,
			"hooks":       userCfg.Hooks,
			"aliases":     userCfg.Aliases,
		},
		"workspace": map[string]interface{}{
			"path":        wsPath,
			"exists":      wsExists,
			"raw":         wsRaw,
			"min_version": wsCfg.MinVersion,
			"settings":    wsCfg.Settings,
			"plugins":     wsCfg.Plugins,
			"env":         wsCfg.Env,
			"tools":       wsCfg.Tools,
			"tasks":       wsCfg.Tasks,
			"hooks":       wsCfg.Hooks,
			"aliases":     wsCfg.Aliases,
		},
		"detected_browsers": detectInstalledBrowsers(),
	})
}

func handleGetConfigRaw(w http.ResponseWriter, r *http.Request) {
	scope := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("scope")))
	var targetPath string
	if scope == "user" {
		targetPath = config.GetGlobalConfigPath()
	} else {
		targetPath = ".unirtm.toml"
	}
	exists := false
	var content string
	if data, err := os.ReadFile(targetPath); err == nil {
		exists = true
		content = string(data)
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"scope":   scope,
		"path":    targetPath,
		"exists":  exists,
		"content": content,
	})
}

func handlePostConfigRaw(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Scope   string `json:"scope"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "invalid json body: " + err.Error()})
		return
	}
	scope := strings.ToLower(strings.TrimSpace(body.Scope))
	if scope != "user" && scope != "workspace" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "scope must be 'user' or 'workspace'"})
		return
	}

	// Validate TOML syntax before saving
	var dummy map[string]interface{}
	if err := toml.Unmarshal([]byte(body.Content), &dummy); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "invalid TOML syntax: " + err.Error()})
		return
	}

	var targetPath string
	if scope == "user" {
		targetPath = config.GetGlobalConfigPath()
	} else {
		targetPath = ".unirtm.toml"
	}

	parent := filepath.Dir(targetPath)
	if err := os.MkdirAll(parent, 0755); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "failed to create directory: " + err.Error()})
		return
	}

	if err := os.WriteFile(targetPath, []byte(body.Content), 0644); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "failed to write file: " + err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"scope":   scope,
		"path":    targetPath,
		"size":    len(body.Content),
	})
}

func getKeybindingsPath(scope string) string {
	if strings.ToLower(strings.TrimSpace(scope)) == "workspace" {
		return filepath.Join(".unirtm", "keybindings.json")
	}
	return filepath.Join(env.GetConfigDir(), "keybindings.json")
}

func handleGetKeybindings(w http.ResponseWriter, r *http.Request) {
	scope := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("scope")))
	if scope == "" {
		scope = "workspace"
	}
	targetPath := getKeybindingsPath(scope)
	exists := false
	raw := "[]"
	var list []map[string]interface{}
	if data, err := os.ReadFile(targetPath); err == nil {
		exists = true
		raw = string(data)
		_ = json.Unmarshal(data, &list)
	}
	if list == nil {
		list = []map[string]interface{}{}
	}
	userPath := getKeybindingsPath("user")
	workspacePath := getKeybindingsPath("workspace")

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"scope":         scope,
		"path":          targetPath,
		"exists":        exists,
		"raw":           raw,
		"keybindings":   list,
		"userPath":      userPath,
		"workspacePath": workspacePath,
	})
}

func handlePostKeybindings(w http.ResponseWriter, r *http.Request) {
	scope := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("scope")))
	if scope == "" {
		scope = "workspace"
	}
	var body struct {
		Raw         string                   `json:"raw"`
		Keybindings []map[string]interface{} `json:"keybindings"`
		Scope       string                   `json:"scope"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "invalid json: " + err.Error()})
		return
	}
	if body.Scope != "" {
		scope = strings.ToLower(strings.TrimSpace(body.Scope))
	}

	targetPath := getKeybindingsPath(scope)
	parent := filepath.Dir(targetPath)
	if err := os.MkdirAll(parent, 0755); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "failed to create config dir: " + err.Error()})
		return
	}

	var contentBytes []byte
	if strings.TrimSpace(body.Raw) != "" {
		var dummy interface{}
		if err := json.Unmarshal([]byte(body.Raw), &dummy); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "raw content is not valid JSON: " + err.Error()})
			return
		}
		contentBytes = []byte(body.Raw)
	} else {
		if body.Keybindings == nil {
			body.Keybindings = []map[string]interface{}{}
		}
		prettyBytes, err := json.MarshalIndent(body.Keybindings, "", "  ")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "failed to encode keybindings: " + err.Error()})
			return
		}
		contentBytes = append(prettyBytes, '\n')
	}

	if err := os.WriteFile(targetPath, contentBytes, 0644); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "failed to write keybindings.json: " + err.Error()})
		return
	}

	var parsedList []map[string]interface{}
	_ = json.Unmarshal(contentBytes, &parsedList)
	if parsedList == nil {
		parsedList = []map[string]interface{}{}
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"scope":       scope,
		"path":        targetPath,
		"raw":         string(contentBytes),
		"keybindings": parsedList,
	})
}

// ─── Version Aliases Management Handlers ─────────────────────────────────────

func handleGetAliases(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	scope := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("scope")))
	if scope == "" {
		scope = "workspace"
	}

	targetFile := ".unirtm.toml"
	if scope == "user" {
		targetFile = config.GetGlobalConfigPath()
	}

	content, _ := config.ReadFileOrEmpty(targetFile)
	aliases := config.GetAliasesFromContent(content)
	if aliases == nil {
		aliases = make(map[string]map[string]string)
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"scope":   scope,
		"aliases": aliases,
	})
}

func handleUpsertAlias(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		Tool    string `json:"tool"`
		Alias   string `json:"alias"`
		Version string `json:"version"`
		Scope   string `json:"scope"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "invalid request body: " + err.Error()})
		return
	}

	req.Tool = strings.TrimSpace(req.Tool)
	req.Alias = strings.TrimSpace(req.Alias)
	req.Version = strings.TrimSpace(req.Version)
	if req.Scope == "" {
		req.Scope = "workspace"
	}
	req.Scope = strings.ToLower(strings.TrimSpace(req.Scope))

	if req.Tool == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "tool name cannot be empty"})
		return
	}
	if req.Alias == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "alias name cannot be empty"})
		return
	}
	if req.Version == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "version cannot be empty"})
		return
	}

	targetFile := ".unirtm.toml"
	if req.Scope == "user" {
		targetFile = config.GetGlobalConfigPath()
	}

	content, _ := config.ReadFileOrEmpty(targetFile)
	updated := config.UpsertAlias(content, req.Tool, req.Alias, req.Version)
	if err := config.WriteTOMLFile(targetFile, []byte(updated)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "failed to save config: " + err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"tool":    req.Tool,
		"alias":   req.Alias,
		"version": req.Version,
		"scope":   req.Scope,
	})
}

func handleDeleteAlias(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tool := strings.TrimSpace(r.URL.Query().Get("tool"))
	alias := strings.TrimSpace(r.URL.Query().Get("alias"))
	scope := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("scope")))

	if tool == "" || alias == "" {
		var body struct {
			Tool  string `json:"tool"`
			Alias string `json:"alias"`
			Scope string `json:"scope"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
			if tool == "" {
				tool = strings.TrimSpace(body.Tool)
			}
			if alias == "" {
				alias = strings.TrimSpace(body.Alias)
			}
			if scope == "" {
				scope = strings.ToLower(strings.TrimSpace(body.Scope))
			}
		}
	}

	if scope == "" {
		scope = "workspace"
	}

	if tool == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "tool name is required"})
		return
	}
	if alias == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "alias name is required"})
		return
	}

	targetFile := ".unirtm.toml"
	if scope == "user" {
		targetFile = config.GetGlobalConfigPath()
	}

	content, err := config.ReadFileOrEmpty(targetFile)
	if err != nil || content == "" {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
		return
	}

	updated, removed := config.RemoveAlias(content, tool, alias)
	if removed {
		if err := config.WriteTOMLFile(targetFile, []byte(updated)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "failed to save config: " + err.Error()})
			return
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"tool":    tool,
		"alias":   alias,
		"scope":   scope,
	})
}

func handlePostConfig(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Scope string      `json:"scope"`
		Key   string      `json:"key"`
		Value interface{} `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "invalid json body: " + err.Error()})
		return
	}

	scope := strings.ToLower(strings.TrimSpace(body.Scope))
	key := strings.TrimSpace(body.Key)
	if scope != "user" && scope != "workspace" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "scope must be 'user' or 'workspace'"})
		return
	}

	if key == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "key cannot be empty"})
		return
	}

	// Machine-specific settings check: browser, editor and terminal are global-only
	cleanKey := strings.TrimPrefix(key, "settings.")
	if scope == "workspace" && (cleanKey == "browser" || cleanKey == "editor" || cleanKey == "terminal") {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": fmt.Sprintf("%s configuration is only permitted in global config to ensure project portability across team machines.", cleanKey),
			"hint":  fmt.Sprintf("Please configure settings.%s in User (global) scope instead.", cleanKey),
		})
		return
	}

	// Workspace-specific settings check: min_version is workspace-only
	if scope == "user" && key == "min_version" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "min_version configuration is only permitted in workspace config to ensure project compatibility.",
			"hint":  "Please configure min_version in Workspace scope instead.",
		})
		return
	}

	// Validate namespace adhering to reserved core tables and plugin namespaces
	if err := config.ValidateKeyNamespace(key); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	valStr := fmt.Sprintf("%v", body.Value)

	// Validate plugin schema if applicable
	segments := config.SplitConfigKey(key)
	if len(segments) >= 4 && segments[0] == "plugins" && segments[2] == "config" {
		pluginID := segments[1]
		property := segments[3]
		cwdDir, _ := os.Getwd()
		manifest, _ := config.FindPluginManifest(pluginID, cwdDir)
		if manifest != nil && manifest.Contributes.Configuration != nil {
			if schema, exists := manifest.Contributes.Configuration[property]; exists {
				if err := config.ValidateConfigValue(property, valStr, schema); err != nil {
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(map[string]interface{}{
						"error": fmt.Sprintf("plugin schema validation failed: %v", err),
					})
					return
				}
			}
		}
	}

	var targetFile string
	if scope == "user" {
		targetFile = config.GetGlobalConfigPath()
		if err := os.MkdirAll(filepath.Dir(targetFile), 0755); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "create global config dir: " + err.Error()})
			return
		}
	} else {
		targetFile = ".unirtm.toml"
	}

	// Normalize settings shorthand (e.g. "browser" -> "settings.browser")
	storeKey := key
	if !strings.HasPrefix(storeKey, "settings.") && (storeKey == "browser" || storeKey == "cache_ttl" || storeKey == "theme" || storeKey == "icon_theme" || storeKey == "language") {
		storeKey = "settings." + storeKey
	}

	fileLock, lockErr := config.AcquireFileLock(targetFile, 5*time.Second)
	if lockErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "acquire config file lock: " + lockErr.Error()})
		return
	}
	defer fileLock.Release()

	existingRaw, _ := os.ReadFile(targetFile)
	var updatedTOML string
	if key == "min_version" && strings.TrimSpace(valStr) == "" {
		var removed bool
		updatedTOML, removed = config.RemoveRootField(string(existingRaw), "min_version")
		if !removed {
			updatedTOML = string(existingRaw)
		}
	} else {
		var err error
		updatedTOML, err = config.UpsertTOMLSetting(string(existingRaw), config.SplitConfigKey(storeKey), valStr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "update toml setting: " + err.Error()})
			return
		}
	}

	if err := os.WriteFile(targetFile, []byte(updatedTOML), 0644); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "write config file: " + err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"scope":   scope,
		"key":     storeKey,
		"value":   valStr,
		"file":    targetFile,
	})
}
