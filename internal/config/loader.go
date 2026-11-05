// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/flosch/pongo2/v6"
	"github.com/joho/godotenv"
	"github.com/pelletier/go-toml/v2"
	"github.com/snowdreamtech/unirtm/internal/pkg/env"
)

// Load loads the UniRTM project configuration from the current directory.
func Load() (*Config, error) {
	return LoadFromDir(".")
}

// LoadFull loads the UniRTM configuration by merging current directory configs,
// parent directory configs, and the global configuration (hierarchy).
// This provides full alignment with mise's hierarchical config loading.
func LoadFull() (*Config, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return Load() // Fallback to Load() if we can't get current directory
	}
	return LoadHierarchy(pwd)
}

// LoadGlobal loads only the global UniRTM configuration.
func LoadGlobal() (*Config, error) {
	globalPath := GetGlobalConfigPath()
	data, err := OsReadFile(globalPath)
	if err != nil {
		globalCfg := &Config{Plugins: make(PluginMap)}
		globalSidecars, _ := DiscoverGlobalSidecars()
		if len(globalSidecars) > 0 {
			MergeSidecarConfigs(globalCfg, globalSidecars)
			return globalCfg, nil
		}
		return globalCfg, err
	}
	globalCfg := &Config{}
	if err := toml.Unmarshal(data, globalCfg); err != nil {
		return nil, err
	}
	globalCfg.PostLoad()
	globalSidecars, _ := DiscoverGlobalSidecars()
	if len(globalSidecars) > 0 {
		MergeSidecarConfigs(globalCfg, globalSidecars)
	}
	return globalCfg, nil
}

// LoadHierarchy loads the UniRTM configuration by merging configs starting from the given directory,
// parent directories up to root, and the global configuration.
func LoadHierarchy(startDir string) (*Config, error) {
	mergedCfg := &Config{}

	// 1. Walk up from startDir to root to find project configs
	curr := startDir
	for {
		cfg, err := LoadFromDir(curr)
		if err == nil {
			mergedCfg.Merge(cfg)
		}

		// Stop at root
		parent := filepath.Dir(curr)
		if parent == curr {
			break
		}
		curr = parent
	}

	// 2. Load global config
	globalPath := GetGlobalConfigPath()
	if data, err := OsReadFile(globalPath); err == nil {
		globalCfg := &Config{}
		if err := toml.Unmarshal(data, globalCfg); err == nil {
			globalCfg.PostLoad()
			mergedCfg.Merge(globalCfg)
		}
	}

	return mergedCfg, nil
}

// GetGlobalConfigPath returns the path to the global unirtm.toml configuration file.
// It delegates to env.GetGlobalConfigPath() which respects UNIRTM_CONFIG_DIR override.
func GetGlobalConfigPath() string {
	return env.GetGlobalConfigPath()
}

// LoadFromDir loads the UniRTM project configuration from the specified directory.
func LoadFromDir(dir string) (*Config, error) {
	configFiles := []string{".unirtm.toml", "unirtm.toml", ".mise.toml", "mise.toml"}
	var data []byte
	var err error
	var foundFile string

	for _, fileName := range configFiles {
		p := filepath.Join(dir, fileName)
		data, err = OsReadFile(p)
		if err == nil {
			foundFile = p
			break
		}
	}

	if foundFile == "" {
		// If no primary config file exists, inspect directory for legacy version files
		// (.nvmrc, .node-version, .go-version, .python-version, .tool-versions)
		legacyTools, legacySources := ReadLegacyVersionFiles(dir)
		if len(legacyTools) > 0 {
			cfg := &Config{
				Tools:         legacyTools,
				LegacySources: legacySources,
			}
			for _, src := range legacySources {
				cfg.SourceFile = filepath.Join(dir, src)
				break
			}
			cfg.PostLoad()

			projSidecars, _ := DiscoverProjectSidecars(dir)
			if len(projSidecars) > 0 {
				MergeSidecarConfigs(cfg, projSidecars)
			}
			return cfg, nil
		}

		// Check for project sidecars
		projSidecars, _ := DiscoverProjectSidecars(dir)
		if len(projSidecars) > 0 {
			sidecarCfg := &Config{Plugins: make(PluginMap)}
			MergeSidecarConfigs(sidecarCfg, projSidecars)
			return sidecarCfg, nil
		}
		return &Config{}, fmt.Errorf("no config file found in %s", dir)
	}

	cfg := &Config{}
	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config from %s: %w", foundFile, err)
	}

	cfg.SourceFile = foundFile
	cfg.PostLoad()

	// Fill any tools not explicitly declared in primary config from legacy files
	legacyTools, legacySources := ReadLegacyVersionFiles(dir)
	if len(legacyTools) > 0 {
		if cfg.Tools == nil {
			cfg.Tools = make(ToolMap)
		}
		if cfg.LegacySources == nil {
			cfg.LegacySources = make(map[string]string)
		}
		for t, tCfg := range legacyTools {
			if _, exists := cfg.Tools[t]; !exists {
				cfg.Tools[t] = tCfg
				cfg.LegacySources[t] = legacySources[t]
			}
		}
	}

	// Discover and merge sidecars (.unirtm/plugins/*.toml)
	projSidecars, _ := DiscoverProjectSidecars(dir)
	if len(projSidecars) > 0 {
		MergeSidecarConfigs(cfg, projSidecars)
	}

	// Apply environment overlay (.unirtm.<env>.toml) if active environment is specified
	activeEnv := GetActiveEnvironment()
	if activeEnv != "" && activeEnv != "local" {
		if envOverlay := LoadEnvironmentOverlay(dir, activeEnv); envOverlay != nil {
			cfg.MergeOverlay(envOverlay)
		}
	}

	// Apply .unirtm.local.toml private override if present
	if localOverlay := LoadLocalOverlay(dir); localOverlay != nil {
		cfg.MergeOverlay(localOverlay)
	}

	return cfg, nil
}

// GetActiveEnvironment returns the active environment name from env.EnvName, UNIRTM_ENV,
// or defaults to "ci" when running in automated CI environments (CI=true / GITHUB_ACTIONS=true).
func GetActiveEnvironment() string {
	if env.EnvName != "" {
		return env.EnvName
	}
	if v := os.Getenv("UNIRTM_ENV"); v != "" {
		return v
	}
	if os.Getenv("CI") == "true" || os.Getenv("GITHUB_ACTIONS") == "true" {
		return "ci"
	}
	return ""
}

// LoadEnvironmentOverlay loads and returns the overlay config for activeEnv from dir if present.
func LoadEnvironmentOverlay(dir, activeEnv string) *Config {
	if activeEnv == "" {
		return nil
	}
	envFiles := []string{
		fmt.Sprintf(".unirtm.%s.toml", activeEnv),
		fmt.Sprintf("unirtm.%s.toml", activeEnv),
		fmt.Sprintf(".mise.%s.toml", activeEnv),
		fmt.Sprintf("mise.%s.toml", activeEnv),
	}
	for _, fileName := range envFiles {
		p := filepath.Join(dir, fileName)
		data, err := OsReadFile(p)
		if err == nil {
			overlay := &Config{}
			if err := toml.Unmarshal(data, overlay); err == nil {
				overlay.SourceFile = p
				overlay.PostLoadOverlay()
				return overlay
			}
		}
	}
	return nil
}

// LoadLocalOverlay loads and returns the local override config (.unirtm.local.toml) from dir if present.
func LoadLocalOverlay(dir string) *Config {
	localFiles := []string{
		".unirtm.local.toml",
		"unirtm.local.toml",
		".mise.local.toml",
		"mise.local.toml",
	}
	for _, fileName := range localFiles {
		p := filepath.Join(dir, fileName)
		data, err := OsReadFile(p)
		if err == nil {
			overlay := &Config{}
			if err := toml.Unmarshal(data, overlay); err == nil {
				overlay.SourceFile = p
				overlay.PostLoadOverlay()
				return overlay
			}
		}
	}
	return nil
}

// DiscoverAllProjectTools scans dir for base config and all environment overlay configs (.unirtm.*.toml, unirtm.*.toml),
// returning a merged ToolMap of all tools declared across all environment files.
// This allows unirtm lock to lock the complete dependency graph for all environments.
func DiscoverAllProjectTools(dir string) (ToolMap, error) {
	allTools := make(ToolMap)

	// 1. Base tools from dir
	baseCfg, err := LoadFromDir(dir)
	if err == nil && baseCfg != nil {
		for k, v := range baseCfg.Tools {
			allTools[k] = v
		}
	}

	// 2. Discover all .unirtm.*.toml and unirtm.*.toml in dir
	entries, readErr := os.ReadDir(dir)
	if readErr == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			if strings.HasSuffix(name, ".toml") &&
				((strings.HasPrefix(name, ".unirtm.") && name != ".unirtm.toml" && name != ".unirtm.local.toml") ||
					(strings.HasPrefix(name, "unirtm.") && name != "unirtm.toml" && name != "unirtm.local.toml") ||
					(strings.HasPrefix(name, ".mise.") && name != ".mise.toml" && name != ".mise.local.toml") ||
					(strings.HasPrefix(name, "mise.") && name != "mise.toml" && name != "mise.local.toml")) {
				p := filepath.Join(dir, name)
				if data, err := OsReadFile(p); err == nil {
					overlay := &Config{}
					if err := toml.Unmarshal(data, overlay); err == nil {
						overlay.PostLoadOverlay()
						for k, v := range overlay.Tools {
							allTools[k] = v
						}
					}
				}
			}
		}
	}

	return allTools, nil
}

// DiscoverEnvironmentProfiles returns all environment profile names found in dir.
func DiscoverEnvironmentProfiles(dir string) []string {
	var profiles []string
	seen := make(map[string]bool)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return profiles
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".toml") {
			var envName string
			if strings.HasPrefix(name, ".unirtm.") && name != ".unirtm.toml" && name != ".unirtm.local.toml" {
				envName = strings.TrimSuffix(strings.TrimPrefix(name, ".unirtm."), ".toml")
			} else if strings.HasPrefix(name, "unirtm.") && name != "unirtm.toml" && name != "unirtm.local.toml" {
				envName = strings.TrimSuffix(strings.TrimPrefix(name, "unirtm."), ".toml")
			}
			if envName != "" && !seen[envName] {
				seen[envName] = true
				profiles = append(profiles, envName)
			}
		}
	}
	sort.Strings(profiles)
	return profiles
}

// ResolveEnvironment resolves environment variables defined in the configuration.
// It returns a map of rendered environment variables, a list of scripts to source,
// a list of redacted keys, and an error if any required variables are missing.
// It supports pongo2 (Jinja2-like) templates.
func (c *Config) ResolveEnvironment() (map[string]string, []string, []string, error) {
	resolved := make(map[string]string)
	var sources []string
	var redacted []string
	var errs []string

	if c.Env == nil {
		return resolved, sources, redacted, nil
	}

	// Prepare template context
	ctx := pongo2.Context{
		"env": make(map[string]string),
	}
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if len(pair) == 2 {
			ctx["env"].(map[string]string)[pair[0]] = pair[1]
		}
	}

	for k, v := range c.Env {
		// Handle special directives starting with underscore
		if strings.HasPrefix(k, "_.") {
			switch k {
			case "_.file":
				if path, ok := v.(string); ok {
					if envs, err := godotenv.Read(path); err == nil {
						for ek, ev := range envs {
							resolved[ek] = ev
							ctx["env"].(map[string]string)[ek] = ev
						}
					}
				}
			case "_.path":
				var paths []string
				switch val := v.(type) {
				case string:
					paths = []string{val}
				case []interface{}:
					for _, item := range val {
						if s, ok := item.(string); ok {
							paths = append(paths, s)
						}
					}
				}
				for _, p := range paths {
					// Render template for path
					rendered := p
					if strings.Contains(p, "{%") || strings.Contains(p, "{{") {
						if tpl, err := pongo2.FromString(p); err == nil {
							if out, err := tpl.Execute(ctx); err == nil {
								rendered = out
							}
						}
					}

					// Apply shell-style expansion
					rendered = os.Expand(rendered, func(name string) string {
						if v, ok := resolved[name]; ok {
							return v
						}
						return env.Get(name)
					})

					currentPath := resolved["PATH"]
					if currentPath == "" {
						currentPath = env.Get("PATH")
					}

					if currentPath == "" {
						resolved["PATH"] = rendered
					} else {
						resolved["PATH"] = rendered + string(os.PathListSeparator) + currentPath
					}
					// Update context
					ctx["env"].(map[string]string)["PATH"] = resolved["PATH"]
				}
			case "_.source":
				if path, ok := v.(string); ok {
					// Render template for source path
					rendered := path
					if strings.Contains(path, "{%") || strings.Contains(path, "{{") {
						if tpl, err := pongo2.FromString(path); err == nil {
							if out, err := tpl.Execute(ctx); err == nil {
								rendered = out
							}
						}
					}

					// Apply shell-style expansion
					rendered = os.Expand(rendered, func(name string) string {
						if v, ok := resolved[name]; ok {
							return v
						}
						return env.Get(name)
					})
					sources = append(sources, rendered)
				}
			case "_.python_venv":
				if venvPath, ok := v.(string); ok {
					// Render template for venv path
					rendered := venvPath
					if strings.Contains(venvPath, "{%") || strings.Contains(venvPath, "{{") {
						if tpl, err := pongo2.FromString(venvPath); err == nil {
							if out, err := tpl.Execute(ctx); err == nil {
								rendered = out
							}
						}
					}

					// Apply shell-style expansion
					rendered = os.Expand(rendered, func(name string) string {
						if v, ok := resolved[name]; ok {
							return v
						}
						return env.Get(name)
					})

					// Normalize path
					absPath := rendered
					if !filepath.IsAbs(rendered) {
						// Assume relative to current directory for now
						// In a real implementation, it might be relative to config_root
						if cwd, err := os.Getwd(); err == nil {
							absPath = filepath.Join(cwd, rendered)
						}
					}

					// Check if venv exists and activate it
					if _, err := OsStat(absPath); err == nil {
						// Determine bin directory
						binDir := "bin"
						if env.RuntimeGOOS == "windows" {
							binDir = "Scripts"
						}
						venvBin := filepath.Join(absPath, binDir)

						// Update PATH
						currentPath := resolved["PATH"]
						if currentPath == "" {
							currentPath = env.Get("PATH")
						}
						if currentPath == "" {
							resolved["PATH"] = venvBin
						} else {
							resolved["PATH"] = venvBin + string(os.PathListSeparator) + currentPath
						}
						ctx["env"].(map[string]string)["PATH"] = resolved["PATH"]

						// Set VIRTUAL_ENV
						resolved["VIRTUAL_ENV"] = absPath
						ctx["env"].(map[string]string)["VIRTUAL_ENV"] = absPath
					}
				}
			}
			continue
		}

		// Regular environment variables
		var valStr string
		isRequired := false
		isRedact := false
		isRm := false
		helpText := ""

		switch val := v.(type) {
		case string:
			valStr = val
		case int, int64, float64, bool:
			valStr = fmt.Sprintf("%v", val)
		case map[string]interface{}:
			if rm, ok := val["rm"].(bool); ok && rm {
				isRm = true
			}
			if v, ok := val["value"]; ok {
				valStr = fmt.Sprintf("%v", v)
			}
			if req, ok := val["required"].(bool); ok && req {
				isRequired = true
			}
			if red, ok := val["redact"].(bool); ok && red {
				isRedact = true
			}
			if h, ok := val["help"].(string); ok {
				helpText = h
			}
		default:
			continue
		}

		if isRm {
			delete(resolved, k)
			delete(ctx["env"].(map[string]string), k)
			continue
		}

		rendered := valStr
		if strings.Contains(valStr, "{%") || strings.Contains(valStr, "{{") {
			// Pre-process to support Jinja2/mise 'is defined' syntax
			reNotDefined := regexp.MustCompile(`(\S+)\s+is\s+not\s+defined`)
			processedTpl := reNotDefined.ReplaceAllString(valStr, "not $1")

			reDefined := regexp.MustCompile(`(\S+)\s+is\s+defined`)
			processedTpl = reDefined.ReplaceAllString(processedTpl, "$1")

			tpl, err := pongo2.FromString(processedTpl)
			if err == nil {
				if out, err := tpl.Execute(ctx); err == nil {
					rendered = out
				}
			}
		}

		// Apply shell-style expansion ($VAR or ${VAR})
		rendered = os.Expand(rendered, func(name string) string {
			if v, ok := resolved[name]; ok {
				return v
			}
			return env.Get(name)
		})

		// Handle required check
		if isRequired && rendered == "" {
			msg := fmt.Sprintf("Environment variable %q is required but not set.", k)
			if helpText != "" {
				msg += " Help: " + helpText
			}
			errs = append(errs, msg)
		}

		// Handle redact
		if isRedact {
			redacted = append(redacted, k)
		}

		resolved[k] = rendered
		// Update context for subsequent variables in the same block
		ctx["env"].(map[string]string)[k] = rendered
	}

	var finalErr error
	if len(errs) > 0 {
		finalErr = fmt.Errorf("environment resolution errors:\n  - %s", strings.Join(errs, "\n  - "))
	}

	return resolved, sources, redacted, finalErr
}

// ApplyEnvironment sets the environment variables defined in the configuration
// to the current process environment.
func (c *Config) ApplyEnvironment() {
	resolved, _, _, err := c.ResolveEnvironment()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	}
	for k, v := range resolved {
		os.Setenv(k, v)
	}
}
