// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type gitStatusCacheEntry struct {
	response GitStatusResponse
	cachedAt time.Time
}

type gitBranchesCacheEntry struct {
	current  string
	branches []GitBranchItem
	cachedAt time.Time
}

var (
	gitStatusCacheMu   sync.RWMutex
	gitStatusCache     = make(map[string]gitStatusCacheEntry)
	gitBranchesCacheMu sync.RWMutex
	gitBranchesCache   = make(map[string]gitBranchesCacheEntry)
)

const (
	gitStatusCacheTTL   = 800 * time.Millisecond
	gitBranchesCacheTTL = 1500 * time.Millisecond
)

// InvalidateGitCache clears any cached Git status across repositories.
func InvalidateGitCache() {
	gitStatusCacheMu.Lock()
	gitStatusCache = make(map[string]gitStatusCacheEntry)
	gitStatusCacheMu.Unlock()

	gitBranchesCacheMu.Lock()
	gitBranchesCache = make(map[string]gitBranchesCacheEntry)
	gitBranchesCacheMu.Unlock()
}

// RegisterGitRoutes registers all Git version control and repository management routes.
func RegisterGitRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/git/repositories", handleGitRepositories)
	mux.HandleFunc("/api/git/status", handleGitStatus)
	mux.HandleFunc("/api/git/stage", handleGitStage)
	mux.HandleFunc("/api/git/unstage", handleGitUnstage)
	mux.HandleFunc("/api/git/discard", handleGitDiscard)
	mux.HandleFunc("/api/git/commit", handleGitCommit)
	mux.HandleFunc("/api/git/diff", handleGitDiff)
	mux.HandleFunc("/api/git/branches", handleGitBranches)
	mux.HandleFunc("/api/git/checkout", handleGitCheckout)
	mux.HandleFunc("/api/git/sync", handleGitSync)
	mux.HandleFunc("/api/git/log", handleGitLog)
	mux.HandleFunc("/api/git/graph", handleGitGraph)
	mux.HandleFunc("/api/git/commit-detail", handleGitCommitDetail)
	mux.HandleFunc("/api/git/commit-diff", handleGitCommitDiff)
	mux.HandleFunc("/api/git/compare", handleGitCompare)
	mux.HandleFunc("/api/git/tag", handleGitTag)
	mux.HandleFunc("/api/git/branch/create", handleGitBranchCreate)
	mux.HandleFunc("/api/git/cherry-pick", handleGitCherryPick)
	mux.HandleFunc("/api/git/revert", handleGitRevert)
	mux.HandleFunc("/api/git/reset", handleGitReset)
	mux.HandleFunc("/api/git/branch/rename", handleGitBranchRename)
	mux.HandleFunc("/api/git/branch/delete", handleGitBranchDelete)
	mux.HandleFunc("/api/git/branch/merge", handleGitBranchMerge)
	mux.HandleFunc("/api/git/branch/rebase", handleGitBranchRebase)
	mux.HandleFunc("/api/git/branch/push", handleGitBranchPush)
	mux.HandleFunc("/api/git/tag/delete", handleGitTagDelete)
	mux.HandleFunc("/api/git/tag/push", handleGitTagPush)
	mux.HandleFunc("/api/git/archive", handleGitArchive)
	mux.HandleFunc("/api/git/stash", handleGitStash)
	mux.HandleFunc("/api/git/stash/apply", handleGitStashApply)
	mux.HandleFunc("/api/git/stash/pop", handleGitStashPop)
	mux.HandleFunc("/api/git/stash/drop", handleGitStashDrop)
	mux.HandleFunc("/api/git/stash/branch", handleGitStashBranch)
	mux.HandleFunc("/api/git/reset-uncommitted", handleGitResetUncommitted)
	mux.HandleFunc("/api/git/clean", handleGitClean)
	mux.HandleFunc("/api/git/pull", handleGitPull)
	mux.HandleFunc("/api/git/fetch", handleGitFetch)
}

type GitChangeItem struct {
	Path   string `json:"path"`
	Status string `json:"status"`
}

type GitStatusResponse struct {
	Success      bool            `json:"success"`
	Branch       string          `json:"branch"`
	Upstream     string          `json:"upstream"`
	Ahead        int             `json:"ahead"`
	Behind       int             `json:"behind"`
	Staged       []GitChangeItem `json:"staged"`
	Unstaged     []GitChangeItem `json:"unstaged"`
	Untracked    []GitChangeItem `json:"untracked"`
	TotalChanges int             `json:"totalChanges"`
}

// GitRepositoryItem represents a git repository in a multi-root/multi-project workspace.
type GitRepositoryItem struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	RootPath     string `json:"rootPath"`
	Branch       string `json:"branch"`
	Upstream     string `json:"upstream"`
	Ahead        int    `json:"ahead"`
	Behind       int    `json:"behind"`
	IsDirty      bool   `json:"isDirty"`
	TotalChanges int    `json:"totalChanges"`
}

func runGitCommandInDir(dir string, args ...string) (string, error) {
	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if askpassPath := GetOrCreateAskPassHelper(); askpassPath != "" {
		cmd.Env = append(os.Environ(),
			"GIT_ASKPASS="+askpassPath,
			"SSH_ASKPASS="+askpassPath,
			"SSH_ASKPASS_REQUIRE=force",
			"UNIRTM_CREDENTIAL_NONCE="+IssueCredentialNonce(),
		)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		errStr := stderr.String()
		if errStr == "" {
			errStr = err.Error()
		}
		return "", errors.New(strings.TrimSpace(errStr))
	}
	return strings.TrimSpace(stdout.String()), nil
}

func resolveGitRepoDir(r *http.Request) string {
	wd, _ := os.Getwd()
	if r == nil {
		return wd
	}
	repo := r.URL.Query().Get("repo")
	if repo == "" {
		repo = r.Header.Get("X-Git-Repo")
	}
	if repo == "" {
		return wd
	}

	// 1. Direct absolute directory
	if filepath.IsAbs(repo) {
		if fi, err := os.Stat(repo); err == nil && fi.IsDir() {
			return repo
		}
	}

	// Find enclosing repo of wd
	baseDir := wd
	curr := wd
	for {
		if _, err := os.Stat(filepath.Join(curr, ".git")); err == nil {
			baseDir = curr
			break
		}
		p := filepath.Dir(curr)
		if p == curr || p == "." || p == "/" {
			break
		}
		curr = p
	}

	// 2. Relative to baseDir
	rel := filepath.Join(baseDir, repo)
	if fi, err := os.Stat(rel); err == nil && fi.IsDir() {
		return rel
	}

	// 3. Match against workspaceRoots
	roots := loadWorkspaceRoots(baseDir)
	for _, root := range roots {
		if root.Name == repo || root.ID == repo || filepath.Base(root.Path) == repo {
			target := root.Path
			if !filepath.IsAbs(target) {
				target = filepath.Join(baseDir, target)
			}
			if fi, err := os.Stat(target); err == nil && fi.IsDir() {
				return target
			}
		}
	}

	// 4. Sibling directory in parent
	parent := filepath.Dir(baseDir)
	sibling := filepath.Join(parent, repo)
	if fi, err := os.Stat(sibling); err == nil && fi.IsDir() {
		return sibling
	}

	return wd
}

var readOnlyGitCmds = map[string]bool{
	"status":    true,
	"log":       true,
	"diff":      true,
	"branch":    true,
	"rev-parse": true,
	"config":    true,
	"remote":    true,
	"show":      true,
	"show-ref":  true,
}

func runGitInReq(r *http.Request, args ...string) (string, error) {
	dir := resolveGitRepoDir(r)
	if len(args) > 0 && !readOnlyGitCmds[args[0]] {
		InvalidateGitCache()
	}
	return runGitCommandInDir(dir, args...)
}

func runGitCommand(args ...string) (string, error) {
	return runGitCommandInDir("", args...)
}

func scanGitRepositories() []GitRepositoryItem {
	wd, _ := os.Getwd()
	visited := make(map[string]bool)
	var candidates []string

	addCandidate := func(dir string) {
		if dir == "" {
			return
		}
		clean := filepath.Clean(dir)
		if visited[clean] {
			return
		}
		visited[clean] = true
		gitDir := filepath.Join(clean, ".git")
		if _, err := os.Stat(gitDir); err == nil {
			candidates = append(candidates, clean)
		}
	}

	// 1. Current working directory and enclosing repository
	curr := wd
	enclosingRepo := ""
	for {
		if _, err := os.Stat(filepath.Join(curr, ".git")); err == nil {
			addCandidate(curr)
			enclosingRepo = curr
			break
		}
		parent := filepath.Dir(curr)
		if parent == curr || parent == "." || parent == "/" {
			break
		}
		curr = parent
	}
	if enclosingRepo == "" {
		addCandidate(wd)
		enclosingRepo = wd
	}

	// 2. Workspace roots
	roots := loadWorkspaceRoots(enclosingRepo)
	for _, root := range roots {
		fullPath := root.Path
		if !filepath.IsAbs(fullPath) {
			fullPath = filepath.Join(enclosingRepo, fullPath)
		}
		addCandidate(fullPath)

		// Check subdirectories of workspace roots
		if entries, err := os.ReadDir(fullPath); err == nil {
			for _, entry := range entries {
				if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
					addCandidate(filepath.Join(fullPath, entry.Name()))
				}
			}
		}
	}

	// 3. Immediate subdirectories of enclosingRepo
	if entries, err := os.ReadDir(enclosingRepo); err == nil {
		for _, entry := range entries {
			if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
				addCandidate(filepath.Join(enclosingRepo, entry.Name()))
			}
		}
	}

	// 4. Sibling directories of enclosingRepo in parent
	parent := filepath.Dir(enclosingRepo)
	if entries, err := os.ReadDir(parent); err == nil {
		for _, entry := range entries {
			if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
				addCandidate(filepath.Join(parent, entry.Name()))
			}
		}
	}

	var results []GitRepositoryItem
	for _, repoDir := range candidates {
		name := filepath.Base(repoDir)
		porcelain, err := runGitCommandInDir(repoDir, "status", "--porcelain=v1", "-b")
		if err != nil {
			continue
		}

		lines := strings.Split(porcelain, "\n")
		branch := "HEAD"
		upstream := ""
		ahead := 0
		behind := 0
		totalChanges := 0

		for _, l := range lines {
			l = strings.TrimRight(l, "\r")
			if strings.TrimSpace(l) == "" {
				continue
			}
			if strings.HasPrefix(l, "## ") {
				header := strings.TrimSpace(strings.TrimPrefix(l, "## "))
				if strings.Contains(header, "...") {
					parts := strings.Split(header, "...")
					branch = parts[0]
					rest := parts[1]
					if idx := strings.Index(rest, " "); idx != -1 {
						upstream = rest[:idx]
						statusPart := rest[idx:]
						if strings.Contains(statusPart, "ahead") {
							var a int
							if n, _ := fmt.Sscanf(statusPart[strings.Index(statusPart, "ahead")+5:], "%d", &a); n > 0 {
								ahead = a
							}
						}
						if strings.Contains(statusPart, "behind") {
							var b int
							if n, _ := fmt.Sscanf(statusPart[strings.Index(statusPart, "behind")+6:], "%d", &b); n > 0 {
								behind = b
							}
						}
					} else {
						upstream = rest
					}
				} else {
					branch = header
				}
			} else {
				totalChanges++
			}
		}

		results = append(results, GitRepositoryItem{
			ID:           name,
			Name:         name,
			RootPath:     repoDir,
			Branch:       branch,
			Upstream:     upstream,
			Ahead:        ahead,
			Behind:       behind,
			IsDirty:      totalChanges > 0,
			TotalChanges: totalChanges,
		})
	}

	// Sort results: place current working directory repo first if present, then alphabetical
	cleanWD := filepath.Clean(wd)
	sort.Slice(results, func(i, j int) bool {
		if filepath.Clean(results[i].RootPath) == cleanWD {
			return true
		}
		if filepath.Clean(results[j].RootPath) == cleanWD {
			return false
		}
		return strings.ToLower(results[i].Name) < strings.ToLower(results[j].Name)
	})

	return results
}

func handleGitRepositories(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	repos := scanGitRepositories()
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"repositories": repos,
	})
}

func handleGitStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	repoDir := resolveGitRepoDir(r)

	// Check short-lived memory cache to prevent process thundering herds
	gitStatusCacheMu.RLock()
	if entry, found := gitStatusCache[repoDir]; found && time.Since(entry.cachedAt) < gitStatusCacheTTL {
		gitStatusCacheMu.RUnlock()
		_ = json.NewEncoder(w).Encode(entry.response)
		return
	}
	gitStatusCacheMu.RUnlock()

	porcelain, err := runGitInReq(r, "status", "--porcelain=v1", "-b")
	if err != nil {
		// Not a git repository or git error
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success":      false,
			"error":        err.Error(),
			"isGitRepo":    false,
			"branch":       "",
			"totalChanges": 0,
		})
		return
	}

	lines := strings.Split(porcelain, "\n")
	branch := "HEAD"
	upstream := ""
	ahead := 0
	behind := 0

	staged := []GitChangeItem{}
	unstaged := []GitChangeItem{}
	untracked := []GitChangeItem{}

	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		if strings.HasPrefix(line, "## ") {
			header := strings.TrimPrefix(line, "## ")
			header = strings.TrimSpace(header)
			if strings.Contains(header, "...") {
				parts := strings.Split(header, "...")
				branch = parts[0]
				rest := parts[1]
				if idx := strings.Index(rest, " "); idx != -1 {
					upstream = rest[:idx]
					statusPart := rest[idx:]
					if strings.Contains(statusPart, "ahead") {
						var a int
						if n, _ := fmt.Sscanf(statusPart[strings.Index(statusPart, "ahead")+5:], "%d", &a); n > 0 {
							ahead = a
						}
					}
					if strings.Contains(statusPart, "behind") {
						var b int
						if n, _ := fmt.Sscanf(statusPart[strings.Index(statusPart, "behind")+6:], "%d", &b); n > 0 {
							behind = b
						}
					}
				} else {
					upstream = rest
				}
			} else {
				branch = header
			}
			continue
		}

		if len(line) < 3 {
			continue
		}

		x := line[0]
		y := line[1]
		fPath := strings.TrimSpace(line[3:])
		fPath = strings.Trim(fPath, "\"")

		if x == '?' && y == '?' {
			untracked = append(untracked, GitChangeItem{Path: fPath, Status: "U"})
		} else {
			if x != ' ' && x != '?' {
				staged = append(staged, GitChangeItem{Path: fPath, Status: string(x)})
			}
			if y != ' ' && y != '?' {
				unstaged = append(unstaged, GitChangeItem{Path: fPath, Status: string(y)})
			}
		}
	}

	resp := GitStatusResponse{
		Success:      true,
		Branch:       branch,
		Upstream:     upstream,
		Ahead:        ahead,
		Behind:       behind,
		Staged:       staged,
		Unstaged:     unstaged,
		Untracked:    untracked,
		TotalChanges: len(staged) + len(unstaged) + len(untracked),
	}

	gitStatusCacheMu.Lock()
	gitStatusCache[repoDir] = gitStatusCacheEntry{
		response: resp,
		cachedAt: time.Now(),
	}
	gitStatusCacheMu.Unlock()

	_ = json.NewEncoder(w).Encode(resp)
}

func handleGitStage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Paths []string `json:"paths"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	args := []string{"add", "--"}
	if len(body.Paths) == 0 || (len(body.Paths) == 1 && (body.Paths[0] == "." || body.Paths[0] == "*")) {
		args = []string{"add", "-A"}
	} else {
		args = append(args, body.Paths...)
	}

	if _, err := runGitInReq(r, args...); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

func handleGitUnstage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Paths []string `json:"paths"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	args := []string{"reset", "HEAD", "--"}
	if len(body.Paths) > 0 && !(len(body.Paths) == 1 && (body.Paths[0] == "." || body.Paths[0] == "*")) {
		args = append(args, body.Paths...)
	}

	if _, err := runGitInReq(r, args...); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

func handleGitDiscard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Paths []string `json:"paths"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	for _, p := range body.Paths {
		if _, err := runGitInReq(r, "checkout", "--", p); err != nil {
			_, _ = runGitInReq(r, "clean", "-fd", "--", p)
		}
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

func handleGitCommit(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Message string `json:"message"`
		Amend   bool   `json:"amend"`
		Sign    *bool  `json:"sign"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.Message) == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "Commit message is required"})
		return
	}

	args := []string{"commit", "-m", strings.TrimSpace(body.Message)}
	if body.Amend {
		args = []string{"commit", "--amend", "-m", strings.TrimSpace(body.Message)}
	}
	if body.Sign != nil {
		if *body.Sign {
			args = append([]string{"-c", "gpg.extraargs=--pinentry-mode=loopback"}, args...)
			args = append(args, "-S")
		} else {
			args = append(args, "--no-gpg-sign")
		}
	}

	out, err := runGitInReq(r, args...)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "output": out})
}

func handleGitDiff(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	filePath := r.URL.Query().Get("path")
	isStaged := r.URL.Query().Get("staged") == "true"

	original := ""
	modified := ""

	if isStaged {
		orig, _ := runGitInReq(r, "show", "HEAD:"+filePath)
		original = orig
		mod, _ := runGitInReq(r, "show", ":"+filePath)
		modified = mod
	} else {
		orig, err := runGitInReq(r, "show", ":"+filePath)
		if err != nil {
			orig, _ = runGitInReq(r, "show", "HEAD:"+filePath)
		}
		original = orig

		repoDir := resolveGitRepoDir(r)
		data, err := os.ReadFile(filepath.Join(repoDir, filePath))
		if err == nil {
			modified = string(data)
		} else {
			wd, safePath, err := getSafeFSPath(filePath)
			if err == nil && wd != "" {
				d, _ := os.ReadFile(safePath)
				modified = string(d)
			}
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"path":     filePath,
		"original": original,
		"modified": modified,
	})
}

type GitBranchItem struct {
	Name      string `json:"name"`
	IsCurrent bool   `json:"isCurrent"`
	IsRemote  bool   `json:"isRemote"`
}

func handleGitBranches(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	repoDir := resolveGitRepoDir(r)
	gitBranchesCacheMu.RLock()
	if entry, found := gitBranchesCache[repoDir]; found && time.Since(entry.cachedAt) < gitBranchesCacheTTL {
		gitBranchesCacheMu.RUnlock()
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"current":  entry.current,
			"branches": entry.branches,
		})
		return
	}
	gitBranchesCacheMu.RUnlock()

	out, err := runGitInReq(r, "branch", "-a", "--format=%(refname:short)|%(HEAD)")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	lines := strings.Split(out, "\n")
	current := ""
	branches := []GitBranchItem{}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		bName := parts[0]
		isHead := len(parts) > 1 && parts[1] == "*"
		if isHead {
			current = bName
		}
		branches = append(branches, GitBranchItem{
			Name:      bName,
			IsCurrent: isHead,
			IsRemote:  strings.HasPrefix(bName, "origin/"),
		})
	}

	gitBranchesCacheMu.Lock()
	gitBranchesCache[repoDir] = gitBranchesCacheEntry{
		current:  current,
		branches: branches,
		cachedAt: time.Now(),
	}
	gitBranchesCacheMu.Unlock()

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"current":  current,
		"branches": branches,
	})
}

func handleGitCheckout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Branch string `json:"branch"`
		Create bool   `json:"create"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.Branch) == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "Branch name is required"})
		return
	}

	args := []string{"checkout", body.Branch}
	if body.Create {
		args = []string{"checkout", "-b", body.Branch}
	}

	if _, err := runGitInReq(r, args...); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "branch": body.Branch})
}

func handleGitSync(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	pullOut, _ := runGitInReq(r, "pull", "--ff-only")
	pushOut, _ := runGitInReq(r, "push")

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"pull":    pullOut,
		"push":    pushOut,
	})
}

type GitGpgInfo struct {
	Status      string `json:"status"` // "verified" | "unverified" | "bad" | "expired" | "revoked" | "none"
	KeyID       string `json:"keyId,omitempty"`
	Signer      string `json:"signer,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"`
}

func parseGpgInfo(rawStatus, keyID, signer, fingerprint string) *GitGpgInfo {
	s := strings.ToUpper(strings.TrimSpace(rawStatus))
	status := "none"
	switch s {
	case "G":
		status = "verified"
	case "U":
		status = "unverified"
	case "B":
		status = "bad"
	case "X", "Y":
		status = "expired"
	case "R":
		status = "revoked"
	default:
		status = "none"
	}
	return &GitGpgInfo{
		Status:      status,
		KeyID:       strings.TrimSpace(keyID),
		Signer:      strings.TrimSpace(signer),
		Fingerprint: strings.TrimSpace(fingerprint),
	}
}

type GitCommitItem struct {
	Hash      string      `json:"hash"`
	ShortHash string      `json:"shortHash"`
	Author    string      `json:"author"`
	Email     string      `json:"email"`
	Timestamp int64       `json:"timestamp"`
	Message   string      `json:"message"`
	Gpg       *GitGpgInfo `json:"gpg,omitempty"`
}

func handleGitLog(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	out, err := runGitInReq(r, "log", "-n", "50", "--pretty=format:%H%x09%h%x09%an%x09%ae%x09%at%x09%s%x09%G?%x09%GK%x09%GS%x09%GF")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	lines := strings.Split(out, "\n")
	commits := []GitCommitItem{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) >= 6 {
			var ts int64
			_, _ = fmt.Sscanf(parts[4], "%d", &ts)
			var gpg *GitGpgInfo
			if len(parts) >= 10 {
				gpg = parseGpgInfo(parts[6], parts[7], parts[8], parts[9])
			} else {
				gpg = parseGpgInfo("", "", "", "")
			}
			commits = append(commits, GitCommitItem{
				Hash:      parts[0],
				ShortHash: parts[1],
				Author:    parts[2],
				Email:     parts[3],
				Timestamp: ts,
				Message:   parts[5],
				Gpg:       gpg,
			})
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"commits": commits,
	})
}

type GitGraphRef struct {
	Name    string `json:"name"`
	Type    string `json:"type"` // "head" | "branch" | "remote" | "tag"
	Current bool   `json:"current"`
}

type GitGraphCommit struct {
	Hash      string        `json:"hash"`
	ShortHash string        `json:"shortHash"`
	Parents   []string      `json:"parents"`
	Author    string        `json:"author"`
	Email     string        `json:"email"`
	Timestamp int64         `json:"timestamp"`
	Message   string        `json:"message"`
	Refs      []GitGraphRef `json:"refs"`
	Gpg       *GitGpgInfo   `json:"gpg,omitempty"`
	IsStash   bool          `json:"isStash,omitempty"`
}

func handleGitGraph(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	rawBranches := r.URL.Query()["branch"]
	var branchList []string
	for _, b := range rawBranches {
		for _, part := range strings.Split(b, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				branchList = append(branchList, part)
			}
		}
	}

	isAll := len(branchList) == 0
	var validBranches []string
	for _, b := range branchList {
		if b == "all" {
			isAll = true
			break
		}
		validBranches = append(validBranches, b)
	}

	filePath := r.URL.Query().Get("path")
	limitStr := r.URL.Query().Get("limit")
	if limitStr == "" {
		limitStr = "200"
	}

	args := []string{"log", "--date-order", "-n", limitStr, "--pretty=format:%H%x09%h%x09%P%x09%an%x09%ae%x09%at%x09%s%x09%D%x09%G?%x09%GK%x09%GS%x09%GF"}
	if !isAll && len(validBranches) > 0 {
		args = append(args, validBranches...)
	} else {
		args = append(args, "--all")
	}
	if filePath != "" {
		args = append(args, "--", filePath)
	}

	out, err := runGitInReq(r, args...)
	if err != nil {
		out = ""
	}

	lines := strings.Split(out, "\n")
	commits := []GitGraphCommit{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) >= 7 {
			var ts int64
			_, _ = fmt.Sscanf(parts[5], "%d", &ts)
			parentsStr := strings.TrimSpace(parts[2])
			parents := []string{}
			if parentsStr != "" {
				parents = strings.Fields(parentsStr)
			}

			refsStr := ""
			if len(parts) >= 8 {
				refsStr = parts[7]
			}
			var refs []GitGraphRef
			if refsStr != "" {
				for _, raw := range strings.Split(refsStr, ",") {
					raw = strings.TrimSpace(raw)
					if raw == "" {
						continue
					}
					if strings.HasPrefix(raw, "HEAD -> ") {
						refs = append(refs, GitGraphRef{Name: strings.TrimPrefix(raw, "HEAD -> "), Type: "branch", Current: true})
					} else if raw == "HEAD" {
						refs = append(refs, GitGraphRef{Name: "HEAD", Type: "head", Current: true})
					} else if strings.HasPrefix(raw, "tag: ") {
						refs = append(refs, GitGraphRef{Name: strings.TrimPrefix(raw, "tag: "), Type: "tag", Current: false})
					} else if strings.HasPrefix(raw, "refs/tags/") {
						refs = append(refs, GitGraphRef{Name: strings.TrimPrefix(raw, "refs/tags/"), Type: "tag", Current: false})
					} else if strings.Contains(raw, "/") {
						refs = append(refs, GitGraphRef{Name: raw, Type: "remote", Current: false})
					} else {
						refs = append(refs, GitGraphRef{Name: raw, Type: "branch", Current: false})
					}
				}
			}

			var gpg *GitGpgInfo
			if len(parts) >= 12 {
				gpg = parseGpgInfo(parts[8], parts[9], parts[10], parts[11])
			} else {
				gpg = parseGpgInfo("", "", "", "")
			}

			commits = append(commits, GitGraphCommit{
				Hash:      parts[0],
				ShortHash: parts[1],
				Parents:   parents,
				Author:    parts[3],
				Email:     parts[4],
				Timestamp: ts,
				Message:   parts[6],
				Refs:      refs,
				Gpg:       gpg,
			})
		}
	}

	// Also parse stashes if showStashes != "false"
	showStashes := r.URL.Query().Get("showStashes")
	if showStashes != "false" {
		stashOut, err := runGitInReq(r, "stash", "list", "--pretty=format:%H%x09%h%x09%P%x09%an%x09%ae%x09%at%x09%s%x09%gd")
		if err == nil && strings.TrimSpace(stashOut) != "" {
			for _, sLine := range strings.Split(stashOut, "\n") {
				sLine = strings.TrimSpace(sLine)
				if sLine == "" {
					continue
				}
				sParts := strings.Split(sLine, "\t")
				if len(sParts) >= 8 {
					var sTs int64
					_, _ = fmt.Sscanf(sParts[5], "%d", &sTs)
					sParents := []string{}
					if sParts[2] != "" {
						sParents = strings.Fields(sParts[2])
					}
					sRefName := sParts[7]
					sCommit := GitGraphCommit{
						Hash:      sParts[0],
						ShortHash: sParts[1],
						Parents:   sParents,
						Author:    sParts[3],
						Email:     sParts[4],
						Timestamp: sTs,
						Message:   sParts[6],
						Refs: []GitGraphRef{
							{
								Name:    sRefName,
								Type:    "stash",
								Current: false,
							},
						},
						Gpg:     &GitGpgInfo{Status: "none"},
						IsStash: true,
					}
					inserted := false
					for idx, c := range commits {
						if sCommit.Timestamp >= c.Timestamp {
							commits = append(commits[:idx], append([]GitGraphCommit{sCommit}, commits[idx:]...)...)
							inserted = true
							break
						}
					}
					if !inserted {
						commits = append(commits, sCommit)
					}
				}
			}
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"commits": commits,
	})
}

type GitCommitFileItem struct {
	Path      string `json:"path"`
	OldPath   string `json:"oldPath,omitempty"`
	Status    string `json:"status"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
}

func handleGitCommitDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	hash := r.URL.Query().Get("hash")
	if hash == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "Hash is required"})
		return
	}

	metaOut, err := runGitInReq(r, "show", "-s", "--format=%H%x09%h%x09%P%x09%an%x09%ae%x09%at%x09%cn%x09%ce%x09%ct%x09%G?%x09%GK%x09%GS%x09%GF%x09%B", hash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	metaParts := strings.Split(metaOut, "\t")
	var fullHash, shortHash, author, email, committer, committerEmail, fullBody string
	var ts, cTs int64
	var parents []string
	var gpg *GitGpgInfo

	if len(metaParts) >= 14 {
		fullHash = metaParts[0]
		shortHash = metaParts[1]
		if pStr := strings.TrimSpace(metaParts[2]); pStr != "" {
			parents = strings.Fields(pStr)
		}
		author = metaParts[3]
		email = metaParts[4]
		_, _ = fmt.Sscanf(metaParts[5], "%d", &ts)
		committer = metaParts[6]
		committerEmail = metaParts[7]
		_, _ = fmt.Sscanf(metaParts[8], "%d", &cTs)
		gpg = parseGpgInfo(metaParts[9], metaParts[10], metaParts[11], metaParts[12])
		fullBody = strings.Join(metaParts[13:], "\t")
	} else if len(metaParts) >= 10 {
		fullHash = metaParts[0]
		shortHash = metaParts[1]
		if pStr := strings.TrimSpace(metaParts[2]); pStr != "" {
			parents = strings.Fields(pStr)
		}
		author = metaParts[3]
		email = metaParts[4]
		_, _ = fmt.Sscanf(metaParts[5], "%d", &ts)
		committer = metaParts[6]
		committerEmail = metaParts[7]
		_, _ = fmt.Sscanf(metaParts[8], "%d", &cTs)
		fullBody = metaParts[9]
		gpg = parseGpgInfo("", "", "", "")
	}

	nameStatusOut, _ := runGitInReq(r, "diff-tree", "--no-commit-id", "--name-status", "-r", "-m", "--first-parent", hash)
	numstatOut, _ := runGitInReq(r, "show", "--numstat", "--pretty=", hash)

	numstatMap := make(map[string][2]int)
	for _, line := range strings.Split(numstatOut, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) >= 3 {
			var a, d int
			_, _ = fmt.Sscanf(parts[0], "%d", &a)
			_, _ = fmt.Sscanf(parts[1], "%d", &d)
			numstatMap[strings.TrimSpace(parts[2])] = [2]int{a, d}
		}
	}

	var files []GitCommitFileItem
	for _, line := range strings.Split(nameStatusOut, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) >= 2 {
			statusCode := parts[0]
			p1 := parts[1]
			p2 := ""
			if len(parts) >= 3 {
				p2 = parts[2]
			}
			finalPath := p1
			oldPath := ""
			if p2 != "" {
				oldPath = p1
				finalPath = p2
			}
			st := "M"
			if len(statusCode) > 0 {
				st = strings.ToUpper(string(statusCode[0]))
			}
			stat := numstatMap[finalPath]
			files = append(files, GitCommitFileItem{
				Path:      finalPath,
				OldPath:   oldPath,
				Status:    st,
				Additions: stat[0],
				Deletions: stat[1],
			})
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"commit": map[string]interface{}{
			"hash":               fullHash,
			"shortHash":          shortHash,
			"parents":            parents,
			"author":             author,
			"email":              email,
			"timestamp":          ts,
			"committer":          committer,
			"committerEmail":     committerEmail,
			"committerTimestamp": cTs,
			"message":            strings.TrimSpace(fullBody),
			"gpg":                gpg,
			"files":              files,
		},
	})
}

func handleGitCommitDiff(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	hash := r.URL.Query().Get("hash")
	filePath := r.URL.Query().Get("path")
	parent := r.URL.Query().Get("parent")

	if hash == "" || filePath == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "hash and path are required"})
		return
	}

	if parent == "" {
		p, err := runGitInReq(r, "rev-parse", hash+"^")
		if err == nil {
			parent = strings.TrimSpace(p)
		}
	}

	original := ""
	if parent != "" {
		orig, err := runGitInReq(r, "show", parent+":"+filePath)
		if err == nil {
			original = orig
		}
	}

	modified := ""
	mod, err := runGitInReq(r, "show", hash+":"+filePath)
	if err == nil {
		modified = mod
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"path":     filePath,
		"hash":     hash,
		"parent":   parent,
		"original": original,
		"modified": modified,
	})
}

type GitCompareCommitInfo struct {
	Hash      string `json:"hash"`
	ShortHash string `json:"shortHash"`
	Author    string `json:"author"`
	Timestamp int64  `json:"timestamp"`
	Message   string `json:"message"`
}

type GitCompareResult struct {
	Base           GitCompareCommitInfo `json:"base"`
	Target         GitCompareCommitInfo `json:"target"`
	Files          []GitCommitFileItem  `json:"files"`
	TotalAdditions int                  `json:"totalAdditions"`
	TotalDeletions int                  `json:"totalDeletions"`
}

func handleGitCompare(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	base := r.URL.Query().Get("base")
	target := r.URL.Query().Get("target")

	if base == "" || target == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "base and target are required"})
		return
	}

	parseMeta := func(h string) GitCompareCommitInfo {
		info := GitCompareCommitInfo{
			Hash:      h,
			ShortHash: h,
		}
		if len(h) >= 8 {
			info.ShortHash = h[:8]
		}
		out, err := runGitInReq(r, "show", "-s", "--format=%H%x09%h%x09%an%x09%at%x09%s", h)
		if err == nil {
			parts := strings.Split(out, "\t")
			if len(parts) >= 5 {
				info.Hash = parts[0]
				info.ShortHash = parts[1]
				info.Author = parts[2]
				_, _ = fmt.Sscanf(parts[3], "%d", &info.Timestamp)
				info.Message = strings.TrimSpace(parts[4])
			}
		}
		return info
	}

	baseMeta := parseMeta(base)
	targetMeta := parseMeta(target)

	nameStatusOut, _ := runGitInReq(r, "diff-tree", "--no-commit-id", "--name-status", "-r", base, target)
	numstatOut, _ := runGitInReq(r, "diff", "--numstat", base, target)

	numstatMap := make(map[string][2]int)
	for _, line := range strings.Split(numstatOut, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) >= 3 {
			var a, d int
			_, _ = fmt.Sscanf(parts[0], "%d", &a)
			_, _ = fmt.Sscanf(parts[1], "%d", &d)
			numstatMap[strings.TrimSpace(parts[2])] = [2]int{a, d}
		}
	}

	var files []GitCommitFileItem
	totalAdd := 0
	totalDel := 0

	for _, line := range strings.Split(nameStatusOut, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		statusCode := "M"
		if len(parts) > 0 && len(parts[0]) > 0 {
			statusCode = strings.ToUpper(string(parts[0][0]))
		}

		var finalPath, oldPath string
		if len(parts) >= 3 {
			oldPath = parts[1]
			finalPath = parts[2]
		} else if len(parts) >= 2 {
			finalPath = parts[1]
		}

		if finalPath != "" {
			var a, d int
			if stats, ok := numstatMap[finalPath]; ok {
				a = stats[0]
				d = stats[1]
			}
			totalAdd += a
			totalDel += d
			files = append(files, GitCommitFileItem{
				Path:      finalPath,
				OldPath:   oldPath,
				Status:    statusCode,
				Additions: a,
				Deletions: d,
			})
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"compare": GitCompareResult{
			Base:           baseMeta,
			Target:         targetMeta,
			Files:          files,
			TotalAdditions: totalAdd,
			TotalDeletions: totalDel,
		},
	})
}

func handleGitTag(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Name    string `json:"name"`
		Hash    string `json:"hash"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" || body.Hash == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "Tag name and commit hash are required"})
		return
	}

	var err error
	if body.Message != "" {
		_, err = runGitInReq(r, "tag", "-a", body.Name, "-m", body.Message, body.Hash)
	} else {
		_, err = runGitInReq(r, "tag", body.Name, body.Hash)
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "name": body.Name})
}

func handleGitBranchCreate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Name     string `json:"name"`
		Hash     string `json:"hash"`
		Checkout bool   `json:"checkout"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" || body.Hash == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "Branch name and commit hash are required"})
		return
	}

	if _, err := runGitInReq(r, "branch", body.Name, body.Hash); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	if body.Checkout {
		if _, err := runGitInReq(r, "checkout", body.Name); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
			return
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "name": body.Name, "checkout": body.Checkout})
}

func handleGitCherryPick(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Hash string `json:"hash"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Hash == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "Commit hash is required"})
		return
	}

	out, err := runGitInReq(r, "cherry-pick", body.Hash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "output": out})
}

func handleGitRevert(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Hash string `json:"hash"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Hash == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "Commit hash is required"})
		return
	}

	out, err := runGitInReq(r, "revert", "--no-edit", body.Hash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "output": out})
}

func handleGitReset(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Hash string `json:"hash"`
		Mode string `json:"mode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Hash == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "Commit hash is required"})
		return
	}

	mode := body.Mode
	if mode != "soft" && mode != "hard" {
		mode = "mixed"
	}

	out, err := runGitInReq(r, "reset", "--"+mode, body.Hash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "mode": mode, "output": out})
}

func handleGitBranchRename(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		OldName string `json:"oldName"`
		NewName string `json:"newName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.OldName == "" || body.NewName == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "Both oldName and newName are required"})
		return
	}
	if _, err := runGitInReq(r, "branch", "-m", body.OldName, body.NewName); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "oldName": body.OldName, "newName": body.NewName})
}

func handleGitBranchDelete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Name  string `json:"name"`
		Force bool   `json:"force"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "Branch name is required"})
		return
	}
	flag := "-d"
	if body.Force {
		flag = "-D"
	}
	if _, err := runGitInReq(r, "branch", flag, body.Name); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "name": body.Name})
}

func handleGitBranchMerge(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Branch string `json:"branch"`
		NoFf   bool   `json:"noFf"`
		Squash bool   `json:"squash"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Branch == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "Branch is required"})
		return
	}
	args := []string{"merge"}
	if body.NoFf {
		args = append(args, "--no-ff")
	}
	if body.Squash {
		args = append(args, "--squash")
	}
	args = append(args, body.Branch)
	out, err := runGitInReq(r, args...)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "branch": body.Branch, "output": out})
}

func handleGitBranchRebase(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Branch string `json:"branch"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Branch == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "Branch is required"})
		return
	}
	out, err := runGitInReq(r, "rebase", body.Branch)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "branch": body.Branch, "output": out})
}

func handleGitBranchPush(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Branch      string `json:"branch"`
		SetUpstream bool   `json:"setUpstream"`
		Force       bool   `json:"force"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Branch == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "Branch is required"})
		return
	}
	args := []string{"push"}
	if body.Force {
		args = append(args, "--force")
	}
	if body.SetUpstream {
		args = append(args, "-u", "origin", body.Branch)
	} else {
		args = append(args, "origin", body.Branch)
	}
	out, err := runGitInReq(r, args...)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "branch": body.Branch, "output": out})
}

func handleGitTagDelete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "Tag name is required"})
		return
	}
	if _, err := runGitInReq(r, "tag", "-d", body.Name); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "name": body.Name})
}

func handleGitTagPush(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Name  string `json:"name"`
		Force bool   `json:"force"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "Tag name is required"})
		return
	}
	args := []string{"push"}
	if body.Force {
		args = append(args, "--force")
	}
	args = append(args, "origin", body.Name)
	out, err := runGitInReq(r, args...)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "name": body.Name, "output": out})
}

func handleGitArchive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	ref := r.URL.Query().Get("ref")
	if ref == "" {
		ref = "HEAD"
	}
	cleanRef := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '_' || r == '-' || r == '/' {
			return r
		}
		return -1
	}, ref)
	filename := strings.ReplaceAll(cleanRef, "/", "_") + ".zip"
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	cmd := exec.Command("git", "archive", "--format=zip", cleanRef)
	cmd.Stdout = w
	_ = cmd.Run()
}

func handleGitStash(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Message          string `json:"message"`
		IncludeUntracked bool   `json:"includeUntracked"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	args := []string{"stash", "push"}
	if body.IncludeUntracked {
		args = append(args, "--include-untracked")
	}
	if strings.TrimSpace(body.Message) != "" {
		args = append(args, "-m", strings.TrimSpace(body.Message))
	}

	out, err := runGitInReq(r, args...)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "output": out})
}

func handleGitResetUncommitted(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	out, err := runGitInReq(r, "reset", "--hard", "HEAD")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "output": out})
}

func handleGitClean(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	out, err := runGitInReq(r, "clean", "-fd")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "output": out})
}

func handleGitPull(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Rebase bool `json:"rebase"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	args := []string{"pull"}
	if body.Rebase {
		args = append(args, "--rebase")
	}

	out, err := runGitInReq(r, args...)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "output": out})
}

func handleGitFetch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	out, err := runGitInReq(r, "fetch", "--all", "--prune")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "output": out})
}

func handleGitStashApply(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Stash string `json:"stash"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	args := []string{"stash", "apply"}
	if strings.TrimSpace(body.Stash) != "" {
		args = append(args, strings.TrimSpace(body.Stash))
	}
	out, err := runGitInReq(r, args...)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "output": out})
}

func handleGitStashPop(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Stash string `json:"stash"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	args := []string{"stash", "pop"}
	if strings.TrimSpace(body.Stash) != "" {
		args = append(args, strings.TrimSpace(body.Stash))
	}
	out, err := runGitInReq(r, args...)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "output": out})
}

func handleGitStashDrop(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Stash string `json:"stash"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	args := []string{"stash", "drop"}
	if strings.TrimSpace(body.Stash) != "" {
		args = append(args, strings.TrimSpace(body.Stash))
	}
	out, err := runGitInReq(r, args...)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "output": out})
}

func handleGitStashBranch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Branch string `json:"branch"`
		Stash  string `json:"stash"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if strings.TrimSpace(body.Branch) == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "branch is required"})
		return
	}
	stashRef := "stash@{0}"
	if strings.TrimSpace(body.Stash) != "" {
		stashRef = strings.TrimSpace(body.Stash)
	}
	out, err := runGitInReq(r, "stash", "branch", strings.TrimSpace(body.Branch), stashRef)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "output": out})
}
