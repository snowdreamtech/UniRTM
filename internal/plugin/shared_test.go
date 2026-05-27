package plugin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandshakeConfig(t *testing.T) {
	assert.Equal(t, uint(1), HandshakeConfig.ProtocolVersion)
	assert.Equal(t, "UNIRTM_PLUGIN", HandshakeConfig.MagicCookieKey)
	assert.Equal(t, "hello", HandshakeConfig.MagicCookieValue)
}

func TestPluginMap(t *testing.T) {
	assert.Contains(t, PluginMap, "backend")
	assert.Contains(t, PluginMap, "provider")

	assert.IsType(t, &BackendPlugin{}, PluginMap["backend"])
	assert.IsType(t, &ProviderPlugin{}, PluginMap["provider"])
}
