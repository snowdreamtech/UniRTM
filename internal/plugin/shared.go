// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package plugin

import (
	"github.com/hashicorp/go-plugin"
)

// HandshakeConfig is a common handshake that is shared by plugin and host.
var HandshakeConfig = plugin.HandshakeConfig{
	// This isn't required when using VersionedPlugins
	ProtocolVersion:  1,
	MagicCookieKey:   "UNIRTM_PLUGIN",
	MagicCookieValue: "hello",
}

// Map is the map of plugins we can dispense.
var PluginMap = map[string]plugin.Plugin{
	"backend":  &BackendPlugin{},
	"provider": &ProviderPlugin{},
}
