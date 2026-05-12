// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package native

import (
	"context"
)

// JuliaHandler handles Julia language versions via GitHub releases.
type JuliaHandler struct {
	GithubHandler
}

func (h *JuliaHandler) Name() string {
	return "julia"
}

func (h *JuliaHandler) ResolveVersions(ctx context.Context, baseURL string) ([]VersionInfo, error) {
	h.Owner = "JuliaLang"
	h.Repo = "julia"
	
	return h.GithubHandler.ResolveVersions(ctx, baseURL)
}
