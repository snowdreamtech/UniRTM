package native

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAllNativeProviders_Name(t *testing.T) {
	tests := []struct {
		handler interface{ Name() string }
		want    string
	}{
		{&GolangHandler{}, "golang"},
		{&JavaHandler{}, "java"},
		{&PythonHandler{}, "python_standalone"},
		{&RustHandler{}, "rust"},
		{&NodeJSHandler{}, "nodejs"},
		{&CMakeHandler{}, "cmake"},
		{&ElixirHandler{}, "elixir"},
		{&ErlangHandler{}, "erlang"},
		{&FlutterHandler{}, "flutter"},
		{&HelmHandler{}, "helm"},
		{&JuliaHandler{}, "julia"},
		{&KubectlHandler{}, "kubectl"},
		{&NinjaHandler{}, "ninja"},
		{&RubyHandler{}, "ruby"},
		{&ZigHandler{}, "zig"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.handler.Name())
		})
	}
}
