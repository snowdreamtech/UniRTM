// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"fmt"
	"os"
	"strings"
	"github.com/flosch/pongo2/v6"
	"github.com/joho/godotenv"
	"github.com/pelletier/go-toml/v2"
	"path/filepath"
	"runtime"
)

// Load loads the UniRTM project configuration from .unirtm.toml or unirtm.toml.
func Load() (*Config, error) {
	data, err := os.ReadFile(".unirtm.toml")
	if err != nil {
		data, err = os.ReadFile("unirtm.toml")
		if err != nil {
			return &Config{}, nil
		}
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// ResolveEnvironment resolves environment variables defined in the configuration.
// It returns a map of rendered environment variables and a list of scripts to source.
// It supports pongo2 (Jinja2-like) templates.
func (c *Config) ResolveEnvironment() (map[string]string, []string) {
	resolved := make(map[string]string)
	var sources []string
	if c.Env == nil {
		return resolved, sources
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
					
					currentPath := resolved["PATH"]
					if currentPath == "" {
						currentPath = os.Getenv("PATH")
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
							currentPath = os.Getenv("PATH")
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
		switch val := v.(type) {
		case string:
			valStr = val
		case int, int64, float64, bool:
			valStr = fmt.Sprintf("%v", val)
		default:
			continue
		}

		rendered := valStr
		if strings.Contains(valStr, "{%") || strings.Contains(valStr, "{{") {
			tpl, err := pongo2.FromString(valStr)
			if err == nil {
				if out, err := tpl.Execute(ctx); err == nil {
					rendered = out
				}
			}
		}

		resolved[k] = rendered
		// Update context for subsequent variables in the same block
		ctx["env"].(map[string]string)[k] = rendered
	}

	return resolved, sources
}

// ApplyEnvironment sets the environment variables defined in the configuration
// to the current process environment.
func (c *Config) ApplyEnvironment() {
	resolved, _ := c.ResolveEnvironment()
	for k, v := range resolved {
		os.Setenv(k, v)
	}
}
