// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package service

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/snowdreamtech/unirtm/internal/backend"
	"github.com/snowdreamtech/unirtm/internal/pkg/download"
	"github.com/snowdreamtech/unirtm/internal/provider"
	"github.com/snowdreamtech/unirtm/internal/repository"
)

func TestUpdateManager_UpdateAll(t *testing.T) {
	tempDataDir := t.TempDir()
	os.Setenv("UNIRTM_DATA_DIR", tempDataDir)
	defer os.Unsetenv("UNIRTM_DATA_DIR")

	tests := []struct {
		name          string
		installations []*repository.Installation
		backends      map[string]backend.Backend
		downloader    download.Downloader
		wantSuccess   int
		wantFail      int
		wantErr       bool
	}{
		{
			name: "all updates succeed",
			installations: []*repository.Installation{
				{
					Tool:        "node",
					Version:     "18.0.0",
					Backend:     "github",
					InstallPath: "/opt/node",
				},
				{
					Tool:        "go",
					Version:     "1.20.0",
					Backend:     "github",
					InstallPath: "/opt/go",
				},
			},
			backends: map[string]backend.Backend{
				"github": &mockUpdateBackend{
					name: "github",
					versions: map[string]*backend.VersionInfo{
						"20.0.0": {
							Version:     "20.0.0",
							DownloadURL: "https://example.com/20",
							Checksum:    "abc",
						},
					},
				},
			},
			downloader:  &mockDownloader{},
			wantSuccess: 2,
			wantFail:    0,
			wantErr:     false,
		},
		{
			name: "one update fails",
			installations: []*repository.Installation{
				{
					Tool:        "node",
					Version:     "18.0.0",
					Backend:     "github",
					InstallPath: "/opt/node",
				},
				{
					Tool:        "python",
					Version:     "3.9.0",
					Backend:     "github",
					InstallPath: "/opt/python",
				},
			},
			backends: map[string]backend.Backend{
				"github": &mockUpdateBackend{
					name: "github",
					versions: map[string]*backend.VersionInfo{
						"20.0.0": {
							Version:     "20.0.0",
							DownloadURL: "https://example.com/20",
							Checksum:    "abc",
						},
					},
				},
			},
			downloader:  &mockDownloader{downloadErr: errors.New("network error")}, // download fails for all, so both fail UpdateTool
			wantSuccess: 0,
			wantFail:    2,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			installRepo := &mockInstallationRepo{
				installations: make(map[string]*repository.Installation),
			}
			for _, inst := range tt.installations {
				key := inst.Tool + "-" + inst.Version
				installRepo.installations[key] = inst
			}

			backendRegistry := backend.NewRegistry()
			for _, b := range tt.backends {
				backendRegistry.Register(b)
			}

			providerRegistry := provider.NewRegistry()
			providerRegistry.Register("node", &mockProvider{name: "node"})
			providerRegistry.Register("go", &mockProvider{name: "go"})
			providerRegistry.Register("python", &mockProvider{name: "python"})

			downloadManager := download.NewManager()
			downloadManager.Register("https", tt.downloader)

			auditRepo := &mockAuditRepo{}

			txManager := &mockTransactionManager{
				tx: &mockTransaction{
					installationRepo: installRepo,
					auditRepo:        auditRepo,
				},
			}

			um := NewUpdateManager(
				backendRegistry,
				providerRegistry,
				downloadManager,
				installRepo,
				auditRepo,
				txManager,
				nil,
			)

			results, err := um.UpdateAll(context.Background())

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			var successes, failures int
			for _, res := range results {
				if res.Success {
					successes++
				} else {
					failures++
					t.Logf("Failed update for %s: %s", res.Tool, res.Error)
				}
			}

			assert.Equal(t, tt.wantSuccess, successes)
			assert.Equal(t, tt.wantFail, failures)
		})
	}
}
