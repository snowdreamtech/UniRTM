// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package native

import (
	"context"
)

// HelmHandler handles Helm tool versions via GitHub releases.
type HelmHandler struct {
	GithubHandler
}

func (h *HelmHandler) Name() string {
	return "helm"
}

func (h *HelmHandler) ResolveVersions(ctx context.Context, baseURL string) ([]VersionInfo, error) {
	h.Owner = "helm"
	h.Repo = "helm"
	
	return h.GithubHandler.ResolveVersions(ctx, baseURL)
}
