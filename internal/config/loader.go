// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"os"
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
		// Only set if not already set, or if we want to override.
		// For now, let's override to ensure configuration takes precedence.
		os.Setenv(k, v)
	}
}
