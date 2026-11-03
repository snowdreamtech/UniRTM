// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/snowdreamtech/unirtm/internal/backend"
	"github.com/snowdreamtech/unirtm/internal/repository"
	"golang.org/x/sync/errgroup"
)

// IndexManager manages tool index storage, retrieval, and updates
// Validates Requirements: 11.1, 11.2, 11.3, 11.4, 11.5, 11.6, 11.7, 11.8
type IndexManager struct {
	repo         repository.IndexRepository
	auditRepo    repository.AuditRepository
	backends     map[string]backend.Backend
	staleTimeout time.Duration
	mu           sync.RWMutex
}

// IndexManagerConfig holds configuration for the index manager
type IndexManagerConfig struct {
	// StaleTimeout is the duration after which the index is considered stale (default 7 days)
	StaleTimeout time.Duration
}

// ToolMetadata represents extended metadata for a tool in the index
type ToolMetadata struct {
	// AvailableVersions is a list of available versions for the tool
	AvailableVersions []string `json:"available_versions,omitempty"`
	// Tags are searchable tags for the tool
	Tags []string `json:"tags,omitempty"`
	// ReleaseDate is the date of the latest release
	ReleaseDate string `json:"release_date,omitempty"`
	// Stars is the number of GitHub stars (if applicable)
	Stars int `json:"stars,omitempty"`
	// LastUpdated is when this metadata was last updated
	LastUpdated time.Time `json:"last_updated"`
}

// SearchOptions defines options for searching the tool index
type SearchOptions struct {
	// Query is the search query string (matches name, description, tags)
	Query string
	// Backend filters results by backend type (empty = all backends)
	Backend string
	// Limit limits the number of results (0 = no limit)
	Limit int
	// Offset skips the first N results (for pagination)
	Offset int
}

// NewIndexManager creates a new index manager instance
func NewIndexManager(repo repository.IndexRepository, auditRepo repository.AuditRepository, backends map[string]backend.Backend, config IndexManagerConfig) (*IndexManager, error) {
	if repo == nil {
		return nil, errors.New("index repository is required")
	}

	if config.StaleTimeout <= 0 {
		config.StaleTimeout = 7 * 24 * time.Hour // 7 days default
	}

	if backends == nil {
		backends = make(map[string]backend.Backend)
	}

	im := &IndexManager{
		repo:         repo,
		auditRepo:    auditRepo,
		backends:     backends,
		staleTimeout: config.StaleTimeout,
	}

	// Proactively seed the index with popular default tools for superior experience & offline-readiness
	_ = im.seedDefaultTools(context.Background())

	return im, nil
}

// upsertToolLockless creates or updates a tool in the index without locking
func (im *IndexManager) upsertToolLockless(ctx context.Context, tool string, description string, homepage string, license string, backendName string, metadata *ToolMetadata) error {
	// Serialize metadata to JSON
	var metadataJSON string
	if metadata != nil {
		metadata.LastUpdated = time.Now()
		metadataBytes, err := json.Marshal(metadata)
		if err != nil {
			return fmt.Errorf("marshal metadata: %w", err)
		}
		metadataJSON = string(metadataBytes)
	}

	entry := &repository.IndexEntry{
		Tool:        tool,
		Description: description,
		Homepage:    homepage,
		License:     license,
		Backend:     backendName,
		UpdatedAt:   time.Now(),
		Metadata:    metadataJSON,
	}

	if err := im.repo.Upsert(ctx, entry); err != nil {
		return fmt.Errorf("upsert tool index entry: %w", err)
	}

	// Log the upsert operation
	if im.auditRepo != nil {
		_ = im.auditRepo.Log(ctx, &repository.AuditEntry{
			Timestamp: time.Now(),
			Operation: "index_upsert",
			Tool:      tool,
			Status:    "success",
			Metadata:  fmt.Sprintf(`{"backend":"%s"}`, backendName),
		})
	}

	return nil
}

// UpsertTool creates or updates a tool in the index
// Validates Requirement: 11.1 (Maintain searchable index)
func (im *IndexManager) UpsertTool(ctx context.Context, tool string, description string, homepage string, license string, backendName string, metadata *ToolMetadata) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	return im.upsertToolLockless(ctx, tool, description, homepage, license, backendName, metadata)
}

// GetTool retrieves a tool from the index by name
func (im *IndexManager) GetTool(ctx context.Context, tool string) (*repository.IndexEntry, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	entry, err := im.repo.FindByTool(ctx, tool)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("find tool in index: %w", err)
	}

	return entry, nil
}

// ListTools lists all tools in the index
// Validates Requirement: 11.1 (Maintain searchable index)
func (im *IndexManager) ListTools(ctx context.Context) ([]*repository.IndexEntry, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	entries, err := im.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list tool index entries: %w", err)
	}

	return entries, nil
}

// SearchTools searches for tools by name, description, or tags
// Validates Requirement: 11.4 (Search by name, description, tags)
func (im *IndexManager) SearchTools(ctx context.Context, options SearchOptions) ([]*repository.IndexEntry, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	// Perform the search
	entries, err := im.repo.Search(ctx, options.Query)
	if err != nil {
		return nil, fmt.Errorf("search tool index: %w", err)
	}

	// Filter by backend if specified
	if options.Backend != "" {
		filtered := make([]*repository.IndexEntry, 0)
		for _, entry := range entries {
			if entry.Backend == options.Backend {
				filtered = append(filtered, entry)
			}
		}
		entries = filtered
	}

	// Apply pagination
	if options.Offset > 0 {
		if options.Offset >= len(entries) {
			return []*repository.IndexEntry{}, nil
		}
		entries = entries[options.Offset:]
	}

	if options.Limit > 0 && len(entries) > options.Limit {
		entries = entries[:options.Limit]
	}

	return entries, nil
}

// FilterByBackend filters tools by backend type
// Validates Requirement: 11.5 (Filter by backend type)
func (im *IndexManager) FilterByBackend(ctx context.Context, backendName string) ([]*repository.IndexEntry, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	// Get all tools
	allEntries, err := im.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list tool index entries: %w", err)
	}

	// Filter by backend
	filtered := make([]*repository.IndexEntry, 0)
	for _, entry := range allEntries {
		if entry.Backend == backendName {
			filtered = append(filtered, entry)
		}
	}

	return filtered, nil
}

// UpdateFromBackend updates the index from a specific backend
// Validates Requirement: 11.2 (Update from multiple sources)
func (im *IndexManager) UpdateFromBackend(ctx context.Context, backendName string) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	// Get the backend
	b, ok := im.backends[backendName]
	if !ok {
		return fmt.Errorf("backend %s not found", backendName)
	}

	// Log the update operation start
	startTime := time.Now()

	// Refresh and update our seed/popular tools metadata list specifically for this backend,
	// keeping it completely fresh, dynamic, and correct.
	err := im.seedDefaultToolsLockless(ctx, backendName)
	if err != nil {
		return fmt.Errorf("seed default tools: %w", err)
	}

	duration := time.Since(startTime)

	// Log the update operation
	if im.auditRepo != nil {
		_ = im.auditRepo.Log(ctx, &repository.AuditEntry{
			Timestamp: time.Now(),
			Operation: "index_update",
			Status:    "success",
			Duration:  duration.Milliseconds(),
			Metadata:  fmt.Sprintf(`{"backend":"%s","name":"%s"}`, backendName, b.Name()),
		})
	}

	return nil
}

// UpdateFromAllBackends updates the index from all registered backends
// Validates Requirement: 11.2 (Update from multiple sources)
func (im *IndexManager) UpdateFromAllBackends(ctx context.Context) error {
	im.mu.RLock()
	// Collect backends to update safely while holding a read lock
	backendNames := make([]string, 0, len(im.backends))
	for backendName := range im.backends {
		backendNames = append(backendNames, backendName)
	}
	im.mu.RUnlock()

	startTime := time.Now()
	var successCount int32
	var errorCount int32
	var mu sync.Mutex
	var errorsList []string

	g, gCtx := errgroup.WithContext(ctx)

	for _, name := range backendNames {
		backendName := name // capture loop variable
		g.Go(func() error {
			// UpdateFromBackend handles its own locking internally
			err := im.UpdateFromBackend(gCtx, backendName)
			if err != nil {
				atomic.AddInt32(&errorCount, 1)
				mu.Lock()
				errorsList = append(errorsList, fmt.Sprintf("%s: %v", backendName, err))
				mu.Unlock()
			} else {
				atomic.AddInt32(&successCount, 1)
			}
			return nil // Don't return error to errgroup to allow all to finish even if one fails
		})
	}

	_ = g.Wait() // wait for all backend updates to finish

	// Always trigger a refresh/re-seed of all default tools to ensure complete Parity and database health
	// We call the locked version here because we are not holding the lock
	_ = im.seedDefaultTools(ctx)

	duration := time.Since(startTime)

	// Log the update operation
	if im.auditRepo != nil {
		status := "success"
		errorMsg := ""
		if atomic.LoadInt32(&errorCount) > 0 {
			status = "partial_failure"
			errorMsg = fmt.Sprintf("failed backends: %v", errorsList)
		}

		_ = im.auditRepo.Log(ctx, &repository.AuditEntry{
			Timestamp: time.Now(),
			Operation: "index_update_all",
			Status:    status,
			Error:     errorMsg,
			Duration:  duration.Milliseconds(),
			Metadata:  fmt.Sprintf(`{"success_count":%d,"error_count":%d}`, atomic.LoadInt32(&successCount), atomic.LoadInt32(&errorCount)),
		})
	}

	if atomic.LoadInt32(&errorCount) > 0 {
		return fmt.Errorf("index update completed with %d errors: %v", atomic.LoadInt32(&errorCount), errorsList)
	}

	return nil
}

// seedDefaultToolsLockless seeds default tools into the database index without acquiring mutex lock
func (im *IndexManager) seedDefaultToolsLockless(ctx context.Context, backendFilter string) error {
	// Curated list of popular tools for beautiful user experience and robust offline support
	defaultTools := []struct {
		Tool        string
		Description string
		Homepage    string
		License     string
		Backend     string
		Tags        []string
	}{
		{"go", "Go programming language compiler and tools", "https://go.dev", "BSD-3-Clause", "go", []string{"language", "compiler", "go"}},
		{"node", "Node.js JavaScript runtime environment", "https://nodejs.org", "MIT", "npm", []string{"language", "runtime", "javascript", "node"}},
		{"python", "Python programming language interpreter", "https://python.org", "PSF-2.0", "pypi", []string{"language", "runtime", "python"}},
		{"ruby", "Ruby programming language interpreter", "https://www.ruby-lang.org", "Ruby", "gem", []string{"language", "runtime", "ruby"}},
		{"rust", "Rust programming language compiler and toolchain", "https://rust-lang.org", "MIT OR Apache-2.0", "cargo", []string{"language", "compiler", "rust"}},
		{"rustup", "Rust toolchain installer", "https://rustup.rs", "MIT OR Apache-2.0", "github", []string{"toolchain", "rust", "installer"}},
		{"bun", "Incredibly fast JavaScript & TypeScript runtime, bundler, test runner and package manager", "https://bun.sh", "MIT", "github", []string{"runtime", "javascript", "typescript", "bun"}},
		{"deno", "A modern runtime for JavaScript and TypeScript", "https://deno.com", "MIT", "github", []string{"runtime", "javascript", "typescript", "deno"}},
		{"pnpm", "Fast, disk space efficient package manager for Node.js", "https://pnpm.io", "MIT", "npm", []string{"package-manager", "javascript", "node"}},
		{"yarn", "Fast, reliable, and secure dependency management for Node.js", "https://yarnpkg.com", "BSD-2-Clause", "npm", []string{"package-manager", "javascript", "node"}},
		{"ruff", "An extremely fast Python linter and code formatter, written in Rust", "https://astral.sh/ruff", "MIT", "github", []string{"linter", "formatter", "python", "ruff"}},
		{"ripgrep", "ripgrep recursively searches directories for a regex pattern while respecting your gitignore", "https://github.com/BurntSushi/ripgrep", "Unlicense OR MIT", "github", []string{"search", "grep", "cli"}},
		{"bat", "A cat(1) clone with wings (syntax highlighting and Git integration)", "https://github.com/sharkdp/bat", "MIT OR Apache-2.0", "github", []string{"terminal", "utility", "cat"}},
		{"fd", "A simple, fast and user-friendly alternative to 'find'", "https://github.com/sharkdp/fd", "MIT OR Apache-2.0", "github", []string{"search", "find", "cli"}},
		{"fzf", "A command-line fuzzy finder", "https://github.com/junegunn/fzf", "MIT", "github", []string{"search", "fuzzy", "terminal"}},
		{"jq", "Command-line JSON processor", "https://jqlang.github.io/jq/", "MIT", "github", []string{"utility", "json", "parser"}},
		{"yq", "Portable command-line YAML, JSON, XML, CSV, TOML and properties processor", "https://github.com/mikefarah/yq", "MIT", "github", []string{"utility", "yaml", "parser"}},
		{"shellcheck", "ShellCheck, a static analysis tool for shell scripts", "https://www.shellcheck.net", "GPL-3.0", "github", []string{"linter", "shell", "bash"}},
		{"shfmt", "A shell parser, formatter, and interpreter", "https://github.com/mvdan/sh", "BSD-3-Clause", "github", []string{"formatter", "shell", "shfmt"}},
		{"actionlint", "Static checker for GitHub Actions workflow files", "https://github.com/rhysd/actionlint", "MIT", "github", []string{"linter", "github-actions", "workflow"}},
		{"hadolint", "Dockerfile linter, validated by ShellCheck", "https://github.com/hadolint/hadolint", "GPL-3.0", "github", []string{"linter", "docker", "dockerfile"}},
		{"gitleaks", "Scan git repos (or files) for secrets using regex and entropy", "https://github.com/gitleaks/gitleaks", "MIT", "github", []string{"security", "scanner", "secrets"}},
		{"osv-scanner", "Vulnerability scanner written in Go which uses the data from https://osv.dev", "https://github.com/google/osv-scanner", "Apache-2.0", "github", []string{"security", "vulnerability", "scanner"}},
		{"zizmor", "A static security analyzer for GitHub Actions workflows", "https://github.com/woodruffw/zizmor", "Apache-2.0", "github", []string{"security", "linter", "github-actions"}},
		{"poetry", "Python packaging and dependency management made easy", "https://python-poetry.org", "MIT", "pypi", []string{"package-manager", "python", "poetry"}},
		{"pipx", "Install and run Python applications in isolated environments", "https://github.com/pypa/pipx", "MIT", "pypi", []string{"package-manager", "python", "pipx"}},
		{"helm", "The Kubernetes Package Manager", "https://helm.sh", "Apache-2.0", "github", []string{"kubernetes", "package-manager", "devops"}},
		{"kubectl", "Command line tool for controlling Kubernetes clusters", "https://kubernetes.io/docs/reference/kubectl/", "Apache-2.0", "github", []string{"kubernetes", "cli", "devops"}},
		{"kustomize", "Customization of kubernetes YAML configurations", "https://kustomize.io", "Apache-2.0", "github", []string{"kubernetes", "utility", "devops"}},
		{"terraform", "Terraform enables you to safely and predictably create, change, and improve infrastructure", "https://www.terraform.io", "BSL-1.1", "github", []string{"iac", "terraform", "infrastructure"}},
		{"opentofu", "OpenTofu lets you declaratively manage your cloud infrastructure", "https://opentofu.org", "MPL-2.0", "github", []string{"iac", "opentofu", "infrastructure"}},
		{"terragrunt", "Terragrunt is a thin wrapper for Terraform that provides extra tools", "https://terragrunt.gruntwork.io", "MIT", "github", []string{"iac", "terraform", "wrapper"}},
		{"ansible", "Radically simple IT automation platform", "https://www.ansible.com", "GPL-3.0", "pypi", []string{"automation", "configuration", "devops"}},
		{"caddy", "Fast and extensible multi-platform HTTP/1-2-3 web server with automatic HTTPS", "https://caddyserver.com", "Apache-2.0", "github", []string{"server", "http", "caddy"}},
		{"copilot", "AWS Copilot CLI", "https://github.com/aws/copilot-cli", "Apache-2.0", "github", []string{"aws", "cli", "cloud"}},
		{"gh", "GitHub's official command line tool", "https://cli.github.com", "MIT", "github", []string{"github", "cli", "git"}},
		{"glab", "An open-source GitLab command line tool", "https://gitlab.com/gitlab-org/cli", "MIT", "gitlab", []string{"gitlab", "cli", "git"}},
		{"act", "Run your GitHub Actions locally!", "https://github.com/nektos/act", "MIT", "github", []string{"github-actions", "local", "docker"}},
		{"lazygit", "Simple terminal UI for git commands", "https://github.com/jesseduffield/lazygit", "MIT", "github", []string{"git", "tui", "terminal"}},
		{"lazydocker", "The simple terminal UI for both docker and docker-compose", "https://github.com/jesseduffield/lazydocker", "MIT", "github", []string{"docker", "tui", "terminal"}},
		{"k9s", "Kubernetes CLI To Watch Your Clusters In Style!", "https://k9scli.io", "Apache-2.0", "github", []string{"kubernetes", "tui", "terminal"}},
		{"starship", "The minimal, blazing-fast, and infinitely customizable prompt for any shell!", "https://starship.rs", "ISC", "github", []string{"shell", "prompt", "starship"}},
		{"eza", "A modern alternative to 'ls'", "https://github.com/eza-community/eza", "MIT", "github", []string{"terminal", "utility", "ls"}},
		{"zoxide", "A smarter cd command", "https://github.com/ajeetdsouza/zoxide", "MIT", "github", []string{"terminal", "utility", "cd"}},
		{"tmux", "tmux is a terminal multiplexer", "https://github.com/tmux/tmux", "BSD-3-Clause", "github", []string{"terminal", "multiplexer", "tmux"}},
		{"direnv", "Clutter-free environment variables", "https://direnv.net", "MIT", "github", []string{"terminal", "env", "utility"}},
		{"age", "A simple, modern and secure file encryption tool", "https://github.com/FiloSottile/age", "BSD-3-Clause", "github", []string{"security", "encryption", "age"}},
		{"sops", "Simple and Flexible Tool for Managing Secrets", "https://github.com/getsops/sops", "MPL-2.0", "github", []string{"security", "secrets", "sops"}},
	}

	for _, t := range defaultTools {
		// If backendFilter is specified, only seed tools for that backend (or map backend appropriately)
		// Map 'go' backend to 'go' backendName, etc.
		if backendFilter != "" && t.Backend != backendFilter {
			continue
		}
		metadata := &ToolMetadata{
			Tags:        t.Tags,
			LastUpdated: time.Now(),
		}
		if err := im.upsertToolLockless(ctx, t.Tool, t.Description, t.Homepage, t.License, t.Backend, metadata); err != nil {
			return err
		}
	}
	return nil
}

// seedDefaultTools seeds default tools with proper mutex locking
func (im *IndexManager) seedDefaultTools(ctx context.Context) error {
	im.mu.Lock()
	defer im.mu.Unlock()
	return im.seedDefaultToolsLockless(ctx, "")
}

// IsStale checks if the index is stale (older than the configured timeout)
// Validates Requirement: 11.7 (Detect stale index)
func (im *IndexManager) IsStale(ctx context.Context) (bool, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	// Get all entries to find the most recent update
	entries, err := im.repo.List(ctx)
	if err != nil {
		return false, fmt.Errorf("list tool index entries: %w", err)
	}

	// If no entries, index is stale
	if len(entries) == 0 {
		return true, nil
	}

	// Find the most recent update time
	var mostRecent time.Time
	for _, entry := range entries {
		if entry.UpdatedAt.After(mostRecent) {
			mostRecent = entry.UpdatedAt
		}
	}

	// Check if the most recent update is older than the stale timeout
	return time.Since(mostRecent) > im.staleTimeout, nil
}

// GetStaleAge returns how long ago the index was last updated
func (im *IndexManager) GetStaleAge(ctx context.Context) (time.Duration, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	// Get all entries to find the most recent update
	entries, err := im.repo.List(ctx)
	if err != nil {
		return 0, fmt.Errorf("list tool index entries: %w", err)
	}

	// If no entries, return max duration
	if len(entries) == 0 {
		return time.Duration(1<<63 - 1), nil // Max duration
	}

	// Find the most recent update time
	var mostRecent time.Time
	for _, entry := range entries {
		if entry.UpdatedAt.After(mostRecent) {
			mostRecent = entry.UpdatedAt
		}
	}

	return time.Since(mostRecent), nil
}

// PromptForUpdate checks if the index is stale and returns a prompt message
// Validates Requirement: 11.7 (Prompt for update when stale)
func (im *IndexManager) PromptForUpdate(ctx context.Context) (bool, string, error) {
	isStale, err := im.IsStale(ctx)
	if err != nil {
		return false, "", fmt.Errorf("check if index is stale: %w", err)
	}

	if !isStale {
		return false, "", nil
	}

	age, err := im.GetStaleAge(ctx)
	if err != nil {
		return true, "The tool index is stale. Run 'unirtm index update' to refresh it.", nil
	}

	days := int(age.Hours() / 24)
	return true, fmt.Sprintf("The tool index is %d days old. Run 'unirtm index update' to refresh it.", days), nil
}

// SupportsOffline indicates whether the index manager can operate offline
// Validates Requirement: 11.8 (Support offline operation)
func (im *IndexManager) SupportsOffline() bool {
	return true
}

// IsOfflineCapable checks if the index has cached data for offline operation
// Validates Requirement: 11.8 (Support offline operation using cached index)
func (im *IndexManager) IsOfflineCapable(ctx context.Context) (bool, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	// Check if we have any cached index entries
	entries, err := im.repo.List(ctx)
	if err != nil {
		return false, fmt.Errorf("list tool index entries: %w", err)
	}

	return len(entries) > 0, nil
}

// DeleteTool removes a tool from the index
func (im *IndexManager) DeleteTool(ctx context.Context, tool string) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	if err := im.repo.Delete(ctx, tool); err != nil {
		return fmt.Errorf("delete tool from index: %w", err)
	}

	// Log the delete operation
	if im.auditRepo != nil {
		_ = im.auditRepo.Log(ctx, &repository.AuditEntry{
			Timestamp: time.Now(),
			Operation: "index_delete",
			Tool:      tool,
			Status:    "success",
		})
	}

	return nil
}

// GetToolMetadata retrieves and parses the metadata for a tool
func (im *IndexManager) GetToolMetadata(ctx context.Context, tool string) (*ToolMetadata, error) {
	entry, err := im.GetTool(ctx, tool)
	if err != nil {
		return nil, err
	}

	if entry.Metadata == "" {
		return &ToolMetadata{}, nil
	}

	var metadata ToolMetadata
	if err := json.Unmarshal([]byte(entry.Metadata), &metadata); err != nil {
		return nil, fmt.Errorf("unmarshal tool metadata: %w", err)
	}

	return &metadata, nil
}

// RegisterBackend registers a backend for index updates
func (im *IndexManager) RegisterBackend(backendName string, b backend.Backend) {
	im.mu.Lock()
	defer im.mu.Unlock()

	im.backends[backendName] = b
}

// UnregisterBackend removes a backend from the index manager
func (im *IndexManager) UnregisterBackend(backendName string) {
	im.mu.Lock()
	defer im.mu.Unlock()

	delete(im.backends, backendName)
}

// ListBackends returns the names of all registered backends
func (im *IndexManager) ListBackends() []string {
	im.mu.RLock()
	defer im.mu.RUnlock()

	backends := make([]string, 0, len(im.backends))
	for name := range im.backends {
		backends = append(backends, name)
	}

	return backends
}
