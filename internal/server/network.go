// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package server

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/network"
)

func resolveNetworkPathFromRequest(r *http.Request, defaultScope string) string {
	scope := r.URL.Query().Get("scope")
	if scope == "" {
		scope = defaultScope
	}
	if scope == "user" || scope == "global" {
		return config.GetGlobalConfigPath()
	}
	if _, err := os.Stat(".unirtm.toml"); err == nil {
		return ".unirtm.toml"
	}
	return config.GetGlobalConfigPath()
}

func handleGetNetworkStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	targetPath := resolveNetworkPathFromRequest(r, "workspace")
	mgr := network.NewManager(nil)
	_ = mgr.LoadFromPath(targetPath)

	status := mgr.GetStatus()
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":      status,
		"recommended": network.RecommendedMirrors(),
		"path":        targetPath,
	})
}

func handleUseNetworkProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		ProfileID string `json:"profile_id"`
		Scope     string `json:"scope"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	targetPath := resolveNetworkPathFromRequest(r, req.Scope)
	mgr := network.NewManager(nil)
	_ = mgr.LoadFromPath(targetPath)

	if err := mgr.UseProfile(req.ProfileID); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	if err := mgr.SaveToPath(targetPath); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "status": mgr.GetStatus()})
}

func handleSaveNetworkProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		ProfileID   string `json:"profile_id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Scope       string `json:"scope"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	targetPath := resolveNetworkPathFromRequest(r, req.Scope)
	mgr := network.NewManager(nil)
	_ = mgr.LoadFromPath(targetPath)

	if err := mgr.SaveAsProfile(req.ProfileID, req.Name, req.Description); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	if err := mgr.SaveToPath(targetPath); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "status": mgr.GetStatus()})
}

func handleResetNetworkProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		ProfileID string `json:"profile_id"`
		Scope     string `json:"scope"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	targetPath := resolveNetworkPathFromRequest(r, req.Scope)
	mgr := network.NewManager(nil)
	_ = mgr.LoadFromPath(targetPath)

	if err := mgr.ResetProfile(req.ProfileID); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	if err := mgr.SaveToPath(targetPath); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "status": mgr.GetStatus()})
}

func handleDeleteNetworkProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		ProfileID string `json:"profile_id"`
		Scope     string `json:"scope"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	targetPath := resolveNetworkPathFromRequest(r, req.Scope)
	mgr := network.NewManager(nil)
	_ = mgr.LoadFromPath(targetPath)

	if err := mgr.DeleteProfile(req.ProfileID); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	if err := mgr.SaveToPath(targetPath); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "status": mgr.GetStatus()})
}

func handleOverrideNetworkRegistry(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		Ecosystem string `json:"ecosystem"`
		URL       string `json:"url"`
		Scope     string `json:"scope"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	targetPath := resolveNetworkPathFromRequest(r, req.Scope)
	mgr := network.NewManager(nil)
	_ = mgr.LoadFromPath(targetPath)

	if err := mgr.SetOverride(req.Ecosystem, req.URL); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	if err := mgr.SaveToPath(targetPath); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "status": mgr.GetStatus()})
}

func handleResetNetworkRegistry(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		Ecosystem string `json:"ecosystem"`
		Scope     string `json:"scope"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	targetPath := resolveNetworkPathFromRequest(r, req.Scope)
	mgr := network.NewManager(nil)
	_ = mgr.LoadFromPath(targetPath)

	if err := mgr.ResetOverride(req.Ecosystem); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	if err := mgr.SaveToPath(targetPath); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "status": mgr.GetStatus()})
}

func handleAddNetworkRegistry(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		Ecosystem string `json:"ecosystem"`
		URL       string `json:"url"`
		Scope     string `json:"scope"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	targetPath := resolveNetworkPathFromRequest(r, req.Scope)
	mgr := network.NewManager(nil)
	_ = mgr.LoadFromPath(targetPath)

	if err := mgr.AddEcosystem(req.Ecosystem, req.URL); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	if err := mgr.SaveToPath(targetPath); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "status": mgr.GetStatus()})
}

func handleDeleteNetworkRegistry(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		Ecosystem string `json:"ecosystem"`
		Scope     string `json:"scope"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	targetPath := resolveNetworkPathFromRequest(r, req.Scope)
	mgr := network.NewManager(nil)
	_ = mgr.LoadFromPath(targetPath)

	if err := mgr.DeleteEcosystem(req.Ecosystem); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	if err := mgr.SaveToPath(targetPath); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "status": mgr.GetStatus()})
}

func handleResetAllNetworkRegistries(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		Scope string `json:"scope"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	targetPath := resolveNetworkPathFromRequest(r, req.Scope)
	mgr := network.NewManager(nil)
	_ = mgr.LoadFromPath(targetPath)

	if err := mgr.ResetDefaultEcosystems(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	if err := mgr.SaveToPath(targetPath); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "status": mgr.GetStatus()})
}

func handleSetNetworkProxy(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		Enabled bool   `json:"enabled"`
		HTTP    string `json:"http"`
		HTTPS   string `json:"https"`
		SOCKS5  string `json:"socks5"`
		NoProxy string `json:"no_proxy"`
		Scope   string `json:"scope"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	targetPath := resolveNetworkPathFromRequest(r, req.Scope)
	mgr := network.NewManager(nil)
	_ = mgr.LoadFromPath(targetPath)

	proxyCfg := network.ProxyConfig{
		Enabled: req.Enabled,
		HTTP:    req.HTTP,
		HTTPS:   req.HTTPS,
		SOCKS5:  req.SOCKS5,
		NoProxy: req.NoProxy,
	}

	if err := mgr.SetProxy(proxyCfg); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	if err := mgr.SaveToPath(targetPath); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "status": mgr.GetStatus()})
}

func handleDetectNetworkProxy(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	detected := network.DetectSystemProxy()
	_ = json.NewEncoder(w).Encode(detected)
}

func handleTestNetworkProxy(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		HTTP   string `json:"http"`
		HTTPS  string `json:"https"`
		SOCKS5 string `json:"socks5"`
	}
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	res := network.TestProxyConnectivity(network.ProxyConfig{
		Enabled: true,
		HTTP:    req.HTTP,
		HTTPS:   req.HTTPS,
		SOCKS5:  req.SOCKS5,
	}, 4*time.Second)

	_ = json.NewEncoder(w).Encode(res)
}

func handlePingNetwork(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var targets []network.ProbeTarget

	if r.Method == http.MethodPost && r.Body != nil {
		var req struct {
			Targets []struct {
				Ecosystem string `json:"ecosystem"`
				Name      string `json:"name"`
				URL       string `json:"url"`
			} `json:"targets"`
			Ecosystem string `json:"ecosystem"`
			URL       string `json:"url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			if len(req.Targets) > 0 {
				for _, t := range req.Targets {
					targets = append(targets, network.ProbeTarget{
						Ecosystem: t.Ecosystem,
						Name:      t.Name,
						Endpoint:  t.URL,
					})
				}
				targets = append(targets, network.ProbeTarget{
					Ecosystem: req.Ecosystem,
					Name:      req.Ecosystem,
					Endpoint:  req.URL,
				})
			}
		}
	}

	if len(targets) == 0 {
		targets = network.DefaultProbeTargets()
	}

	results := network.ProbeAll(targets, 3000*time.Millisecond)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"results": results,
	})
}
