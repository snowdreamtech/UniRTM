// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package guard

import (
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// LoadGuardConfig discovers and unmarshals the [guard] section from workspace TOML configurations
func LoadGuardConfig(workspaceDir string) (GuardConfig, error) {
	defaultCfg := GuardConfig{
		Enabled: true,
		Mode:    "changed",
	}

	if workspaceDir == "" {
		return defaultCfg, nil
	}

	candidates := []string{
		filepath.Join(workspaceDir, ".unirtm.toml"),
		filepath.Join(workspaceDir, "unirtm.toml"),
		filepath.Join(workspaceDir, ".unirtm.local.toml"),
	}

	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var raw struct {
			Guard GuardConfig `toml:"guard"`
		}

		if err := toml.Unmarshal(data, &raw); err != nil {
			return defaultCfg, err
		}

		// Ensure defaults if omitted in config
		cfg := raw.Guard
		if cfg.Mode == "" {
			cfg.Mode = "changed"
		}
		return cfg, nil
	}

	return defaultCfg, nil
}
