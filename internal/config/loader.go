// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
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
	data, err := os.ReadFile(globalPath)
	if err != nil {
		return &Config{}, err
	}
	globalCfg := &Config{}
	if err := toml.Unmarshal(data, globalCfg); err != nil {
		return nil, err
	}
	globalCfg.PostLoad()
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
	if data, err := os.ReadFile(globalPath); err == nil {
		globalCfg := &Config{}
		if err := toml.Unmarshal(data, globalCfg); err == nil {
			globalCfg.PostLoad()
			mergedCfg.Merge(globalCfg)
		}
	}

	return mergedCfg, nil
}

// GetGlobalConfigPath returns the path to the global unirtm.toml configuration file.
// (Copied from env package to avoid circular dependency if needed, or just use full path)
func GetGlobalConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".config", "unirtm", "unirtm.toml")
}

// LoadFromDir loads the UniRTM project configuration from the specified directory.
func LoadFromDir(dir string) (*Config, error) {
	configFiles := []string{".unirtm.toml", "unirtm.toml", ".mise.toml", "mise.toml"}
	var data []byte
	var err error
	var foundFile string

	for _, fileName := range configFiles {
		p := filepath.Join(dir, fileName)
		data, err = os.ReadFile(p)
		if err == nil {
			foundFile = p
			break
		}
	}

	if foundFile == "" {
		return &Config{}, fmt.Errorf("no config file found in %s", dir)
	}

	cfg := &Config{}
	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config from %s: %w", foundFile, err)
	}

	cfg.PostLoad()

	return cfg, nil
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
					if _, err := os.Stat(absPath); err == nil {
						// Determine bin directory
						binDir := "bin"
						if runtime.GOOS == "windows" {
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
