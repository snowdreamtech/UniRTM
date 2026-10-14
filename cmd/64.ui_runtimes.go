// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/snowdreamtech/unirtm/internal/backend"
	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/database"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/repository/sqlite"
)

type RuntimeInfo struct {
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

type RuntimesStatusResponse struct {
	Runtimes         []RuntimeInfo `json:"runtimes"`
	TotalInstalled   int           `json:"total_installed"`
	DetectedLegacies []string      `json:"detected_legacies"`
}

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

func handleGetRuntimesStatus(w http.ResponseWriter, r *http.Request) {
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

	// 3. Assemble runtime info
	runtimeList := make([]RuntimeInfo, 0, len(coreRuntimes))
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

		runtimeList = append(runtimeList, RuntimeInfo{
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

	_ = json.NewEncoder(w).Encode(RuntimesStatusResponse{
		Runtimes:         runtimeList,
		TotalInstalled:   totalInstalled,
		DetectedLegacies: detectedLegacies,
	})
}

func handleGetRuntimesRemote(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tool := strings.TrimSpace(r.URL.Query().Get("tool"))
	if tool == "" {
		tool = "node"
	}

	nb := backend.NewNativeBackend()
	versions, err := nb.ListVersions(r.Context(), tool, backend.CurrentPlatform())
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	type remoteVerDTO struct {
		Version string `json:"version"`
	}

	dtos := make([]remoteVerDTO, 0, len(versions))
	for _, v := range versions {
		dtos = append(dtos, remoteVerDTO{Version: v.Version})
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"tool":     tool,
		"versions": dtos,
	})
}

func handleUseRuntime(w http.ResponseWriter, r *http.Request) {
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

func handleMigrateLegacyRuntime(w http.ResponseWriter, r *http.Request) {
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

	if err := os.WriteFile(targetFile, []byte(content), 0644); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	_, _ = config.FormatFile(targetFile, false)

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success":        true,
		"migrated_count": migratedCount,
		"file":           targetFile,
	})
}
