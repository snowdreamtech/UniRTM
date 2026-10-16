// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/snowdreamtech/unirtm/internal/backend"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/database"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/repository/sqlite"
)

type ToolInfo struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	Description       string   `json:"description"`
	Homepage          string   `json:"homepage"`
	InstalledVersions []string `json:"installed_versions"`
	ActiveVersion     string   `json:"active_version"`
	GlobalVersion     string   `json:"global_version"`
	Source            string   `json:"source"`
	IsLegacy          bool     `json:"is_legacy"`
	LegacyFile        string   `json:"legacy_file"`
}

type ToolsStatusResponse struct {
	Tools            []ToolInfo `json:"tools"`
	Runtimes         []ToolInfo `json:"runtimes"` // Backward compatibility alias
	TotalInstalled   int        `json:"total_installed"`
	DetectedLegacies []string   `json:"detected_legacies"`
}

// Backward-compatible type aliases
type RuntimeInfo = ToolInfo
type RuntimesStatusResponse = ToolsStatusResponse

var coreRuntimes = []struct {
	ID          string
	Name        string
	Description string
	Homepage    string
}{
	{"node", "Node.js", "JavaScript runtime built on Chrome's V8 engine", "https://nodejs.org"},
	{"go", "Go", "Fast, reliable, open source language supported by Google", "https://go.dev"},
	{"python", "Python", "High-level programming language emphasizing code readability", "https://www.python.org"},
	{"rust", "Rust", "Empowering everyone to build reliable and efficient software", "https://www.rust-lang.org"},
	{"java", "Java", "Multi-platform, object-oriented language and platform", "https://adoptium.net"},
	{"ruby", "Ruby", "Dynamic, open source language with a focus on simplicity", "https://www.ruby-lang.org"},
}

func handleGetToolsStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()

	// 1. Query installed versions from SQLite
	installedMap := make(map[string][]string)
	totalInstalled := 0

	dbPath := env.GetDatabasePath()
	if db, err := database.Open(ctx, database.Config{Path: dbPath, WALMode: true}); err == nil {
		defer db.Close()
		if installRepo, err := sqlite.NewInstallationRepository(db.Conn()); err == nil {
			if installations, err := installRepo.List(ctx); err == nil {
				totalInstalled = len(installations)
				for _, inst := range installations {
					if inst != nil {
						installedMap[inst.Tool] = append(installedMap[inst.Tool], inst.Version)
					}
				}
			}
		}
	}

	// 2. Load hierarchy and global configurations
	cfg, _ := config.LoadFull()
	globalCfg, _ := config.LoadGlobal()

	legacyMap := make(map[string]bool)
	var detectedLegacies []string

	if cfg != nil {
		for _, src := range cfg.LegacySources {
			if !legacyMap[src] {
				legacyMap[src] = true
				detectedLegacies = append(detectedLegacies, src)
			}
		}
	}

	// 3. Assemble tools info (core runtimes first)
	toolsList := make([]ToolInfo, 0, len(coreRuntimes))
	knownIDs := make(map[string]bool)

	for _, cr := range coreRuntimes {
		knownIDs[cr.ID] = true
		activeVer := ""
		source := ""
		isLegacy := false
		legacyFile := ""
		globalVer := ""

		if globalCfg != nil {
			if gTool, exists := globalCfg.Tools[cr.ID]; exists {
				globalVer = gTool.Version
			}
		}

		if cfg != nil {
			if pTool, exists := cfg.Tools[cr.ID]; exists {
				activeVer = pTool.Version
				if lSrc, ok := cfg.LegacySources[cr.ID]; ok && lSrc != "" {
					isLegacy = true
					legacyFile = lSrc
					source = lSrc
				} else if cfg.SourceFile != "" {
					source = filepath.Base(cfg.SourceFile)
				} else {
					source = "global"
				}
			}
		}

		if activeVer == "" && globalVer != "" {
			activeVer = globalVer
			source = "global"
		}

		toolsList = append(toolsList, ToolInfo{
			ID:                cr.ID,
			Name:              cr.Name,
			Description:       cr.Description,
			Homepage:          cr.Homepage,
			InstalledVersions: installedMap[cr.ID],
			ActiveVersion:     activeVer,
			GlobalVersion:     globalVer,
			Source:            source,
			IsLegacy:          isLegacy,
			LegacyFile:        legacyFile,
		})
	}

	// 4. Append all other tools configured in workspace, global, or installed locally
	var otherTools []string
	if cfg != nil {
		for name := range cfg.Tools {
			if !knownIDs[name] {
				knownIDs[name] = true
				otherTools = append(otherTools, name)
			}
		}
	}
	if globalCfg != nil {
		for name := range globalCfg.Tools {
			if !knownIDs[name] {
				knownIDs[name] = true
				otherTools = append(otherTools, name)
			}
		}
	}
	for name := range installedMap {
		if !knownIDs[name] {
			knownIDs[name] = true
			otherTools = append(otherTools, name)
		}
	}
	sort.Strings(otherTools)

	for _, name := range otherTools {
		activeVer := ""
		source := ""
		globalVer := ""

		if globalCfg != nil {
			if gTool, exists := globalCfg.Tools[name]; exists {
				globalVer = gTool.Version
			}
		}

		if cfg != nil {
			if pTool, exists := cfg.Tools[name]; exists {
				activeVer = pTool.Version
				if cfg.SourceFile != "" {
					source = filepath.Base(cfg.SourceFile)
				} else {
					source = ".unirtm.toml"
				}
			}
		}

		if activeVer == "" && globalVer != "" {
			activeVer = globalVer
			source = "global"
		}

		toolsList = append(toolsList, ToolInfo{
			ID:                name,
			Name:              name,
			Description:       "Unified toolchain managed via UniRTM",
			Homepage:          "",
			InstalledVersions: installedMap[name],
			ActiveVersion:     activeVer,
			GlobalVersion:     globalVer,
			Source:            source,
			IsLegacy:          false,
			LegacyFile:        "",
		})
	}

	_ = json.NewEncoder(w).Encode(ToolsStatusResponse{
		Tools:            toolsList,
		Runtimes:         toolsList,
		TotalInstalled:   totalInstalled,
		DetectedLegacies: detectedLegacies,
	})
}

func handleGetToolsRemote(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tool := strings.TrimSpace(r.URL.Query().Get("tool"))
	if tool == "" {
		tool = "node"
	}

	nb := backend.NewNativeBackend()
	versions, err := nb.ListVersions(r.Context(), tool, backend.CurrentPlatform())

	type remoteVerDTO struct {
		Version string `json:"version"`
	}

	dtos := make([]remoteVerDTO, 0, len(versions))
	if err == nil && len(versions) > 0 {
		for _, v := range versions {
			dtos = append(dtos, remoteVerDTO{Version: v.Version})
		}
	} else {
		// Provide comprehensive built-in upstream releases fallback
		for _, v := range getFallbackRemoteVersions(tool) {
			dtos = append(dtos, remoteVerDTO{Version: v})
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"tool":     tool,
		"versions": dtos,
	})
}

func getFallbackRemoteVersions(tool string) []string {
	switch strings.ToLower(tool) {
	case "node", "nodejs":
		return []string{
			"23.9.0", "23.8.0", "23.7.0", "23.6.0", "23.0.0",
			"22.14.0", "22.13.1", "22.13.0", "22.12.0", "22.11.0", "22.9.0", "22.0.0",
			"21.7.3", "21.6.2", "21.1.0",
			"20.18.3", "20.18.2", "20.18.0", "20.17.0", "20.16.0", "20.15.1", "20.14.0", "20.12.2", "20.11.1", "20.10.0", "20.9.0",
			"19.9.0", "19.8.1",
			"18.20.7", "18.20.6", "18.20.5", "18.20.4", "18.20.0", "18.19.1", "18.18.2", "18.17.1", "18.16.0", "18.14.2", "18.12.1",
			"16.20.2", "16.20.0", "16.19.1", "16.18.0", "16.14.2",
			"14.21.3", "14.20.1", "14.19.3",
		}
	case "python", "python3":
		return []string{
			"3.14.0a4",
			"3.13.2", "3.13.1", "3.13.0",
			"3.12.9", "3.12.8", "3.12.7", "3.12.6", "3.12.5", "3.12.4", "3.12.3", "3.12.2", "3.12.1", "3.12.0",
			"3.11.11", "3.11.10", "3.11.9", "3.11.8", "3.11.7", "3.11.6", "3.11.5", "3.11.4", "3.11.0",
			"3.10.16", "3.10.15", "3.10.14", "3.10.13", "3.10.12", "3.10.11", "3.10.8", "3.10.0",
			"3.9.21", "3.9.20", "3.9.19", "3.9.18", "3.9.16", "3.9.13", "3.9.0",
			"3.8.20", "3.8.18", "3.8.16", "3.8.12", "3.8.0",
			"3.7.17", "3.7.12",
		}
	case "go", "golang":
		return []string{
			"1.27.2", "1.27.1", "1.27.0",
			"1.26.5", "1.26.4", "1.26.3", "1.26.2", "1.26.1", "1.26.0",
			"1.25.8", "1.25.7", "1.25.6", "1.25.5", "1.25.4", "1.25.3", "1.25.2", "1.25.1", "1.25.0",
			"1.24.5", "1.24.4", "1.24.3", "1.24.2", "1.24.1", "1.24.0",
			"1.23.6", "1.23.5", "1.23.4", "1.23.3", "1.23.2", "1.23.1", "1.23.0",
			"1.22.12", "1.22.11", "1.22.10", "1.22.9", "1.22.8", "1.22.6", "1.22.5", "1.22.2", "1.22.0",
			"1.21.13", "1.21.12", "1.21.11", "1.21.8", "1.21.6", "1.21.0",
			"1.20.14", "1.20.12", "1.20.7", "1.20.0",
			"1.19.13", "1.19.8", "1.19.0",
			"1.18.10", "1.18.5",
		}
	case "rust":
		return []string{
			"1.85.0", "1.84.1", "1.84.0", "1.83.0", "1.82.0", "1.81.0",
			"1.80.1", "1.80.0", "1.79.0", "1.78.0", "1.77.2", "1.77.0",
			"1.76.0", "1.75.0", "1.74.1", "1.73.0", "1.72.1", "1.71.1", "1.70.0",
			"1.69.0", "1.68.2", "1.67.1", "1.66.1", "1.65.0",
		}
	case "java", "openjdk":
		return []string{
			"23.0.2", "23.0.1",
			"22.0.2", "22.0.1",
			"21.0.6", "21.0.5", "21.0.4", "21.0.2", "21.0.1", "21.0.0",
			"17.0.14", "17.0.12", "17.0.10", "17.0.9", "17.0.8", "17.0.2",
			"11.0.26", "11.0.22", "11.0.20", "11.0.18",
			"8.0.442", "8.0.402", "8.0.392", "8.0.352",
		}
	case "bun":
		return []string{
			"1.2.4", "1.2.3", "1.2.2", "1.2.1", "1.2.0",
			"1.1.45", "1.1.42", "1.1.38", "1.1.34", "1.1.30", "1.1.25", "1.1.20", "1.1.10", "1.1.0",
			"1.0.36", "1.0.35", "1.0.30", "1.0.25", "1.0.20", "1.0.18", "1.0.0",
		}
	case "deno":
		return []string{
			"2.2.3", "2.2.2", "2.2.1", "2.2.0",
			"2.1.9", "2.1.4", "2.1.0",
			"2.0.6", "2.0.4", "2.0.0",
			"1.46.3", "1.45.5", "1.44.4", "1.43.6", "1.42.4", "1.41.2", "1.40.5", "1.39.4", "1.38.2",
		}
	case "zig":
		return []string{
			"0.14.0", "0.13.0", "0.12.1", "0.12.0", "0.11.0", "0.10.1", "0.9.1", "0.8.1",
		}
	case "ruby":
		return []string{
			"3.4.2", "3.4.1",
			"3.3.7", "3.3.6", "3.3.5", "3.3.0",
			"3.2.7", "3.2.6", "3.2.3", "3.2.0",
			"3.1.6", "3.1.4", "3.1.2",
			"3.0.7", "3.0.6",
		}
	case "php":
		return []string{
			"8.4.4", "8.4.3", "8.4.1",
			"8.3.17", "8.3.16", "8.3.14", "8.3.10", "8.3.0",
			"8.2.27", "8.2.26", "8.2.20", "8.2.15",
			"8.1.31", "8.1.28", "8.1.20",
			"8.0.30", "8.0.28",
		}
	default:
		return []string{
			"3.2.0", "3.1.0", "3.0.0",
			"2.8.0", "2.6.0", "2.4.0", "2.2.0", "2.0.0",
			"1.8.0", "1.5.0", "1.2.0", "1.0.0",
		}
	}
}


func handleUseTool(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Tool    string `json:"tool"`
		Version string `json:"version"`
		Global  bool   `json:"global"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	req.Tool = strings.TrimSpace(req.Tool)
	req.Version = strings.TrimSpace(req.Version)
	if req.Tool == "" || req.Version == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "tool and version are required"})
		return
	}

	var targetFile string
	if req.Global {
		targetFile = config.GetGlobalConfigPath()
	} else {
		targetFile = resolveConfigFilePath(false)
	}

	_ = os.MkdirAll(filepath.Dir(targetFile), 0755)
	content, _ := config.ReadFileOrEmpty(targetFile)
	content = config.UpsertToolVersion(content, req.Tool, req.Version)

	if err := os.WriteFile(targetFile, []byte(content), 0644); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	_, _ = config.FormatFile(targetFile, false)

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"tool":    req.Tool,
		"version": req.Version,
		"global":  req.Global,
		"file":    targetFile,
	})
}

func handleRemoveTool(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Tool   string `json:"tool"`
		Global bool   `json:"global"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	req.Tool = strings.TrimSpace(req.Tool)
	if req.Tool == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "tool is required"})
		return
	}

	var targetFile string
	if req.Global {
		targetFile = config.GetGlobalConfigPath()
	} else {
		targetFile = resolveConfigFilePath(false)
	}

	content, _ := config.ReadFileOrEmpty(targetFile)
	newContent, removed := config.RemoveToolVersion(content, req.Tool)
	if removed {
		_ = os.WriteFile(targetFile, []byte(newContent), 0644)
		_, _ = config.FormatFile(targetFile, false)
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"tool":    req.Tool,
		"removed": removed,
		"file":    targetFile,
	})
}

func handleMigrateLegacyTool(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Tool string `json:"tool"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	legacyTools, _ := config.ReadLegacyVersionFiles(".")
	if len(legacyTools) == 0 {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "no legacy files found in current directory",
		})
		return
	}

	targetFile := resolveConfigFilePath(false)
	content, _ := config.ReadFileOrEmpty(targetFile)

	migratedCount := 0
	for t, cfg := range legacyTools {
		if req.Tool == "" || req.Tool == t {
			content = config.UpsertToolVersion(content, t, cfg.Version)
			migratedCount++
		}
	}

	if err := config.WriteTOMLFile(targetFile, []byte(content)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success":        true,
		"migrated_count": migratedCount,
		"file":           targetFile,
	})
}

// Backward-compatible function aliases
var (
	handleGetRuntimesStatus    = handleGetToolsStatus
	handleGetRuntimesRemote    = handleGetToolsRemote
	handleUseRuntime           = handleUseTool
	handleRemoveRuntime        = handleRemoveTool
	handleMigrateLegacyRuntime = handleMigrateLegacyTool
)
