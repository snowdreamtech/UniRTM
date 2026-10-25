// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package server

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	"github.com/snowdreamtech/unirtm/internal/config"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
	"github.com/snowdreamtech/unirtm/internal/utils"
)

func getFileSize(path string) string {
	info, err := os.Stat(path)
	if err != nil {
		return "0 B"
	}
	return utils.FormatBytes(info.Size())
}

func getBestEditorWithSource(cfg *config.Config) (string, string) {
	// 1. Env vars
	for _, e := range []string{"UNIRTM_EDITOR", "VISUAL", "EDITOR"} {
		if v := env.Get(e); v != "" {
			return v, "$" + e
		}
	}

	// 2. Config settings
	if cfg != nil && cfg.Settings.Editor != "" {
		return cfg.Settings.Editor, "unirtm settings"
	}

	// 3. System defaults
	var defaults []string
	if runtime.GOOS == "windows" {
		defaults = []string{"notepad", "code", "vim"}
	} else {
		defaults = []string{"vim", "vi", "nano", "code", "emacs"}
	}

	for _, d := range defaults {
		if _, err := exec.LookPath(d); err == nil {
			return d, "system default"
		}
	}

	return "vi", "fallback"
}

func resolveConfigFilePath(global bool) string {
	if global {
		return env.GetGlobalConfigPath()
	}
	candidates := []string{"unirtm.toml", ".unirtm.toml", "unirtm.yaml", ".unirtm.yaml"}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return "unirtm.toml"
}

// APIResponse represents the standard UniRTM API envelope format.
type APIResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// WriteJSON sends a JSON response with status code and payload.
func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// WriteSuccess sends a successful standard envelope response.
func WriteSuccess(w http.ResponseWriter, data any) {
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

// WriteError sends a standard JSON error response with an HTTP status code.
func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, map[string]interface{}{
		"success": false,
		"error":   message,
	})
}
