// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"os"
	"strings"
	"github.com/pelletier/go-toml/v2"
)

// LoadProjectConfig loads the UniRTM project configuration from .unirtm.toml or unirtm.toml.
func LoadProjectConfig() (*Config, error) {
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

// ApplyEnvironment sets the environment variables defined in the configuration
// to the current process environment.
func (c *Config) ApplyEnvironment() {
	if c.Env == nil {
		return
	}

	for k, v := range c.Env {
		// Basic template rendering for Jinja-like templates commonly found in mise configs.
		// Example: {% if env.CI is defined %}0{% else %}1{% endif %}
		rendered := v
		if strings.Contains(v, "{%") && strings.Contains(v, "%}") {
			isCI := os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != ""
			
			// Handle {% if env.CI is defined %}...{% else %}...{% endif %}
			if strings.Contains(v, "if env.CI is defined") {
				parts := strings.Split(v, "{% else %}")
				if len(parts) == 2 {
					if isCI {
						// Extract content between if and else
						start := strings.Index(parts[0], "%}") + 2
						rendered = strings.TrimSpace(parts[0][start:])
					} else {
						// Extract content between else and endif
						end := strings.Index(parts[1], "{% endif %}")
						if end != -1 {
							rendered = strings.TrimSpace(parts[1][:end])
						} else {
							// fallback
							start := strings.Index(parts[1], "%}") + 2
							rendered = strings.TrimSpace(parts[1][start:])
						}
					}
				}
			}
		}

		// Only set if not already set, or if we want to override.
		// For now, let's override to ensure configuration takes precedence.
		os.Setenv(k, rendered)
	}
}
