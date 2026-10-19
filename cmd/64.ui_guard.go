// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/snowdreamtech/unirtm/internal/guard"
)

func handleGuardStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "method not allowed"})
		return
	}

	pwd, err := os.Getwd()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	mode := r.URL.Query().Get("mode")
	if mode == "" {
		mode = "staged"
	}

	walkOpts := guard.WalkOptions{
		WorkspaceDir: pwd,
		Mode:         mode,
	}

	files, err := guard.CollectFiles(r.Context(), walkOpts)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	cfg := guard.GuardConfig{
		Enabled: true,
		Mode:    mode,
	}

	pipeline := guard.NewPipeline(pwd, cfg)
	res, err := pipeline.Run(r.Context(), files, guard.ExecuteOptions{
		Fix: false,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	fixableCount := 0
	for _, iss := range res.Issues {
		if iss.Fixable {
			fixableCount++
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"mode":           mode,
		"scanned":        res.Summary.TotalFilesScanned,
		"clean":          !res.HasErrors && len(res.Issues) == 0,
		"errors_count":   res.Summary.ErrorsCount,
		"warnings_count": res.Summary.WarningsCount,
		"fixable_count":  fixableCount,
		"issues":         res.Issues,
	})
}

func handleGuardRun(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "method not allowed"})
		return
	}

	var req struct {
		Mode  string   `json:"mode"`
		Fix   bool     `json:"fix"`
		Files []string `json:"files"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Mode = "staged"
		req.Fix = true
	}

	pwd, err := os.Getwd()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	if req.Mode == "" {
		req.Mode = "staged"
	}

	var files []string
	if len(req.Files) > 0 {
		files = req.Files
	} else {
		walkOpts := guard.WalkOptions{
			WorkspaceDir: pwd,
			Mode:         req.Mode,
		}
		files, err = guard.CollectFiles(r.Context(), walkOpts)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
			return
		}
	}

	cfg := guard.GuardConfig{
		Enabled: true,
		Mode:    req.Mode,
	}

	pipeline := guard.NewPipeline(pwd, cfg)
	res, err := pipeline.Run(r.Context(), files, guard.ExecuteOptions{
		Fix: req.Fix,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"summary": res.Summary,
		"issues":  res.Issues,
	})
}

func handleGuardRules(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "method not allowed"})
		return
	}

	rules := guard.GetAllBuiltinRules()
	type RuleMeta struct {
		ID         string   `json:"id"`
		Category   string   `json:"category"`
		Phase      string   `json:"phase"`
		Extensions []string `json:"extensions"`
	}

	var meta []RuleMeta
	for _, rule := range rules {
		meta = append(meta, RuleMeta{
			ID:         rule.ID(),
			Category:   string(rule.Category()),
			Phase:      string(rule.Phase()),
			Extensions: rule.DefaultExtensions(),
		})
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"builtin_rules": meta,
		"slots":         guard.DefaultToolSlots,
	})
}
