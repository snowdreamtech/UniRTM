// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

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

	cfg, err := guard.LoadGuardConfig(pwd)
	if err != nil {
		cfg = guard.GuardConfig{Enabled: true}
	}
	cfg.Mode = mode

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
		"rule_reports":   res.RuleReports,
		"logs":           res.Logs,
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

	cfg, err := guard.LoadGuardConfig(pwd)
	if err != nil {
		cfg = guard.GuardConfig{Enabled: true}
	}
	cfg.Mode = req.Mode

	pipeline := guard.NewPipeline(pwd, cfg)
	res, err := pipeline.Run(r.Context(), files, guard.ExecuteOptions{
		Mode: req.Mode,
		Fix:  req.Fix,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"summary":      res.Summary,
		"issues":       res.Issues,
		"rule_reports": res.RuleReports,
		"logs":         res.Logs,
	})
}

func handleGuardRules(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "method not allowed"})
		return
	}

	pwd, err := os.Getwd()
	if err != nil {
		pwd = "."
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

	type ConfigFileMeta struct {
		Name        string `json:"name"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Category    string `json:"category"`
		Phase       string `json:"phase"`
		Exists      bool   `json:"exists"`
		Path        string `json:"path"`
	}

	standardConfigs := []struct {
		Name     string
		Title    string
		Desc     string
		Category string
		Phase    string
	}{
		{".unirtm.toml", "UniRTM 守卫总配置", "定义质量守卫运行模式、排除路径与工具槽位绑定", "core", "all"},
		{".golangci.yml", "Go 静态检查门禁", "Go 语言代码质量规则集与静态分析配置", "lint", "Phase 3"},
		{".prettierrc", "Web 统一格式化", "JavaScript/TypeScript/Vue/JSON 视觉排版规范", "format", "Phase 2"},
		{".gitleaks.toml", "密钥泄露防护规则", "密码、私钥、API Key 与 Token 泄露拦截策略", "secret", "Phase 3"},
		{".editorconfig", "跨编辑器基础规范", "统一文件编码、换行符及缩进风格", "hygiene", "Phase 1"},
		{"eslint.config.mjs", "Web 前端代码质量", "TypeScript 与 Vue 语法规约与静态拦截", "lint", "Phase 3"},
		{".shellcheckrc", "Shell 脚本检查规范", "POSIX / Bash 陷阱与质量规范配置", "lint", "Phase 3"},
		{"ruff.toml", "Python 快速检查排版", "Python 代码语法树排版与规范拦截配置", "lint", "Phase 2/3"},
	}

	var configMetas []ConfigFileMeta
	for _, sc := range standardConfigs {
		fullPath := filepath.Join(pwd, sc.Name)
		fi, err := os.Stat(fullPath)
		exists := err == nil && !fi.IsDir()
		configMetas = append(configMetas, ConfigFileMeta{
			Name:        sc.Name,
			Title:       sc.Title,
			Description: sc.Desc,
			Category:    sc.Category,
			Phase:       sc.Phase,
			Exists:      exists,
			Path:        sc.Name,
		})
	}

	cfg, _ := guard.LoadGuardConfig(pwd)
	resolvedSlots := guard.GetResolvedToolSlots(cfg)

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"builtin_rules": meta,
		"slots":         resolvedSlots,
		"config_files":  configMetas,
	})
}

func handleGuardSlots(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	pwd, err := os.Getwd()
	if err != nil {
		pwd = "."
	}

	switch r.Method {
	case http.MethodGet:
		cfg, _ := guard.LoadGuardConfig(pwd)
		slots := guard.GetResolvedToolSlots(cfg)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"slots":   slots,
		})

	case http.MethodPost:
		var slot guard.SlotConfig
		if err := json.NewDecoder(r.Body).Decode(&slot); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "invalid request payload: " + err.Error()})
			return
		}
		if slot.ID == "" {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "slot ID is required"})
			return
		}

		if err := guard.SaveSlotConfig(pwd, slot); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "failed to save slot: " + err.Error()})
			return
		}

		cfg, _ := guard.LoadGuardConfig(pwd)
		slots := guard.GetResolvedToolSlots(cfg)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"slots":   slots,
		})

	case http.MethodDelete:
		slotID := r.URL.Query().Get("id")
		if slotID == "" {
			var body struct {
				ID string `json:"id"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
				slotID = body.ID
			}
		}
		if slotID == "" {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "slot ID is required"})
			return
		}

		if err := guard.DeleteSlotConfig(pwd, slotID); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "failed to delete slot: " + err.Error()})
			return
		}

		cfg, _ := guard.LoadGuardConfig(pwd)
		slots := guard.GetResolvedToolSlots(cfg)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"slots":   slots,
		})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "method not allowed"})
	}
}

func handleGuardToml(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	pwd, err := os.Getwd()
	if err != nil {
		pwd = "."
	}

	switch r.Method {
	case http.MethodGet:
		content, err := guard.GetGuardToml(pwd)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"content": content,
		})

	case http.MethodPost:
		var req struct {
			Content string `json:"content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "invalid payload: " + err.Error()})
			return
		}

		if err := guard.SaveGuardToml(pwd, req.Content); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "failed to save TOML: " + err.Error()})
			return
		}

		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
		})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "method not allowed"})
	}
}
