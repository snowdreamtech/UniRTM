// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package native

import (
	"context"
)

// ErlangHandler handles Erlang/OTP versions via RabbitMQ's zero-dependency releases.
type ErlangHandler struct {
	GithubHandler
}

func (h *ErlangHandler) Name() string {
	return "erlang"
}

func (h *ErlangHandler) ResolveVersions(ctx context.Context, baseURL string) ([]VersionInfo, error) {
	h.Owner = "rabbitmq"
	h.Repo = "erlang-relbin"

	return h.GithubHandler.ResolveVersions(ctx, baseURL)
}
