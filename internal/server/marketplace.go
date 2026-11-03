// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package server

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/theme"
)

// RegisterMarketplaceRoutes registers theme and extension marketplace routes.
func RegisterMarketplaceRoutes(mux *http.ServeMux) {
	themeMgr := theme.NewManager()

	mux.HandleFunc("/api/themes", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetThemes(w, r, themeMgr)
		case http.MethodDelete:
			handleDeleteTheme(w, r, themeMgr)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "method not allowed"})
		}
	})
	mux.HandleFunc("/api/themes/install", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		handleInstallTheme(w, r, themeMgr)
	})
	mux.HandleFunc("/api/themes/apply", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		handleApplyTheme(w, r)
	})
	mux.HandleFunc("/api/themes/assets/", func(w http.ResponseWriter, r *http.Request) {
		handleThemeAssets(w, r, themeMgr)
	})
	mux.HandleFunc("/api/themes/marketplace/search", func(w http.ResponseWriter, r *http.Request) {
		handleMarketplaceSearch(w, r, themeMgr)
	})
	mux.HandleFunc("/api/themes/marketplace/detail", func(w http.ResponseWriter, r *http.Request) {
		handleMarketplaceDetail(w, r, themeMgr)
	})
	mux.HandleFunc("/api/themes/marketplace/install", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		handleMarketplaceInstall(w, r, themeMgr)
	})
}

// ─── Theme Management Handlers ────────────────────────────────────────────────

func handleGetThemes(w http.ResponseWriter, r *http.Request, themeMgr *theme.Manager) {
	w.Header().Set("Content-Type", "application/json")
	resp, err := themeMgr.ListThemes()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	activeColor := "default"
	activeIcon := "seti"
	activeProduct := "codicons"
	activeAppearance := "system"

	checkCfg := func(c *config.Config) {
		if c == nil {
			return
		}
		if c.Settings.Appearance != "" {
			activeAppearance = c.Settings.Appearance
		}
		if c.Settings.Theme != "" {
			th := c.Settings.Theme
			if th == "light" || th == "dark" || th == "system" {
				activeAppearance = th
				activeColor = "default"
			} else {
				activeColor = th
			}
		}
		if c.Settings.IconTheme != "" {
			activeIcon = c.Settings.IconTheme
		}
		if c.Settings.ProductIconTheme != "" {
			activeProduct = c.Settings.ProductIconTheme
		}
	}

	if globalCfg, err := config.LoadGlobal(); err == nil && globalCfg != nil {
		checkCfg(globalCfg)
	}
	cm := config.NewConfigManager()
	if cfg, _ := cm.LoadHierarchy(r.Context()); cfg != nil {
		checkCfg(cfg)
	}
	resp.Active.Color = activeColor
	resp.Active.Icon = activeIcon
	resp.Active.Product = activeProduct
	resp.Active.Appearance = activeAppearance

	_ = json.NewEncoder(w).Encode(resp)
}

func handleInstallTheme(w http.ResponseWriter, r *http.Request, themeMgr *theme.Manager) {
	w.Header().Set("Content-Type", "application/json")
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "failed to parse multipart form: " + err.Error()})
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "missing 'file' in form data: " + err.Error()})
		return
	}
	defer file.Close()

	items, err := themeMgr.InstallVSIX(file)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "install failed: " + err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"installed": items,
	})
}

func handleDeleteTheme(w http.ResponseWriter, r *http.Request, themeMgr *theme.Manager) {
	w.Header().Set("Content-Type", "application/json")
	themeType := r.URL.Query().Get("type")
	packageID := r.URL.Query().Get("id")

	if themeType == "" || packageID == "" {
		// Fallback to JSON body if provided
		var body struct {
			Type      string `json:"type"`
			PackageID string `json:"package_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
			if themeType == "" {
				themeType = body.Type
			}
			if packageID == "" {
				packageID = body.PackageID
			}
		}
	}

	if themeType == "" || packageID == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "type and id are required"})
		return
	}

	if err := themeMgr.DeleteTheme(theme.ThemeType(themeType), packageID); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

func handleApplyTheme(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var body struct {
		Color      string `json:"color"`
		Icon       string `json:"icon"`
		Product    string `json:"product"`
		Appearance string `json:"appearance"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "invalid json body: " + err.Error()})
		return
	}

	targetFile := config.GetGlobalConfigPath()
	if err := os.MkdirAll(filepath.Dir(targetFile), 0755); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "create config dir: " + err.Error()})
		return
	}

	fileLock, lockErr := config.AcquireFileLock(targetFile, 5*time.Second)
	if lockErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "acquire config file lock: " + lockErr.Error()})
		return
	}
	defer fileLock.Release()

	existingRaw, _ := os.ReadFile(targetFile)
	content := string(existingRaw)
	var err error

	if body.Color != "" {
		content, err = config.UpsertTOMLSetting(content, []string{"settings", "theme"}, body.Color)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "upsert theme: " + err.Error()})
			return
		}
	}

	if body.Icon != "" {
		content, err = config.UpsertTOMLSetting(content, []string{"settings", "icon_theme"}, body.Icon)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "upsert icon_theme: " + err.Error()})
			return
		}
	}

	if body.Product != "" {
		content, err = config.UpsertTOMLSetting(content, []string{"settings", "product_icon_theme"}, body.Product)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "upsert product_icon_theme: " + err.Error()})
			return
		}
	}

	if body.Appearance != "" {
		content, err = config.UpsertTOMLSetting(content, []string{"settings", "appearance"}, body.Appearance)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "upsert appearance: " + err.Error()})
			return
		}
	}

	if err := os.WriteFile(targetFile, []byte(content), 0644); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "write config file: " + err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"color":      body.Color,
		"icon":       body.Icon,
		"product":    body.Product,
		"appearance": body.Appearance,
	})
}

func handleThemeAssets(w http.ResponseWriter, r *http.Request, themeMgr *theme.Manager) {
	rel := strings.TrimPrefix(r.URL.Path, "/api/themes/assets/")
	parts := strings.SplitN(rel, "/", 3)
	if len(parts) < 3 {
		http.NotFound(w, r)
		return
	}
	themeType := parts[0]
	packageID := parts[1]
	subpath := parts[2]

	absPath, err := themeMgr.ResolveThemeFile(theme.ThemeType(themeType), packageID, subpath)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, absPath)
}

func handleMarketplaceSearch(w http.ResponseWriter, r *http.Request, themeMgr *theme.Manager) {
	w.Header().Set("Content-Type", "application/json")
	q := r.URL.Query().Get("q")
	sizeStr := r.URL.Query().Get("size")
	offsetStr := r.URL.Query().Get("offset")

	size := 20
	offset := 0
	if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 {
		size = s
	}
	if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
		offset = o
	}

	installedIDs := make(map[string]bool)
	if themes, err := themeMgr.ListThemes(); err == nil && themes != nil {
		for _, c := range themes.Colors {
			if c.PackageID != "" {
				installedIDs[c.PackageID] = true
			}
		}
		for _, i := range themes.Icons {
			if i.PackageID != "" {
				installedIDs[i.PackageID] = true
			}
		}
		for _, p := range themes.Products {
			if p.PackageID != "" {
				installedIDs[p.PackageID] = true
			}
		}
	}

	result, err := theme.SearchOpenVSXThemes(q, size, offset, installedIDs)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(result)
}

func handleMarketplaceInstall(w http.ResponseWriter, r *http.Request, themeMgr *theme.Manager) {
	w.Header().Set("Content-Type", "application/json")
	var body struct {
		DownloadURL string `json:"download_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.DownloadURL == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "download_url is required"})
		return
	}

	items, err := themeMgr.InstallFromURL(body.DownloadURL)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "online install failed: " + err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"installed": items,
	})
}

func handleMarketplaceDetail(w http.ResponseWriter, r *http.Request, themeMgr *theme.Manager) {
	w.Header().Set("Content-Type", "application/json")
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")
	if namespace == "" || name == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "namespace and name are required"})
		return
	}

	installedIDs := make(map[string]bool)
	if themes, err := themeMgr.ListThemes(); err == nil && themes != nil {
		for _, c := range themes.Colors {
			if c.PackageID != "" {
				installedIDs[c.PackageID] = true
			}
		}
		for _, i := range themes.Icons {
			if i.PackageID != "" {
				installedIDs[i.PackageID] = true
			}
		}
		for _, p := range themes.Products {
			if p.PackageID != "" {
				installedIDs[p.PackageID] = true
			}
		}
	}

	detail, err := theme.GetOpenVSXExtensionDetail(namespace, name, installedIDs)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(detail)
}
