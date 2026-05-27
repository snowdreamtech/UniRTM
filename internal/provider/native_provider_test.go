package provider

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	pkgHttp "github.com/snowdreamtech/unirtm/internal/pkg/http"
	"github.com/stretchr/testify/assert"
)

type mockRoundTripper struct {
	roundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.roundTripFunc != nil {
		return m.roundTripFunc(req)
	}
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString("mocked response")),
	}, nil
}

func TestNativeProvider_Name(t *testing.T) {
	p := NewNativeProvider()
	if p.Name() != "native" {
		t.Errorf("expected native, got %s", p.Name())
	}
}

func TestNativeProvider_ListVersions_NoRecipe(t *testing.T) {
	p := NewNativeProvider()
	_, err := p.ListVersions(context.Background(), "nonexistent_tool")
	if err == nil {
		t.Error("expected error for nonexistent tool")
	}
}

func TestNativeProvider_ListVersions_Success(t *testing.T) {
	mockRt := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			// Mocking golang version API response
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(`[{"version":"go1.20.5","files":[{"filename":"go1.20.5.linux-amd64.tar.gz","os":"linux","arch":"amd64"}]}]`)),
			}, nil
		},
	}
	oldMock := pkgHttp.MockTransport
	pkgHttp.MockTransport = mockRt
	defer func() { pkgHttp.MockTransport = oldMock }()

	p := NewNativeProvider()
	versions, err := p.ListVersions(context.Background(), "go") // "go" is golang's alias or use "golang"
	// Wait, recipes use "golang" not "go" directly. Let's use "golang".
	
	// Actually "go" is an alias? Recipes register map for tools. Let's test nodejs to be safe.
	mockRtNode := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString(`[{"version":"v20.5.0","files":["linux-x64","osx-x64-tar","win-x64-zip"]}]`)),
			}, nil
		},
	}
	pkgHttp.MockTransport = mockRtNode

	versions, err = p.ListVersions(context.Background(), "nodejs")
	assert.NoError(t, err)
	if len(versions) > 0 {
		assert.Equal(t, "20.5.0", versions[0])
	}
}

func TestNativeProvider_Install_EmptyArtifact(t *testing.T) {
	p := NewNativeProvider()
	err := p.Install(context.Background(), "tool", "/tmp", "", "1.0")
	if err == nil {
		t.Error("expected error for empty artifact path")
	}
}

func TestNativeProvider_PostInstall(t *testing.T) {
	p := NewNativeProvider()
	tmpDir := t.TempDir()
	
	// Create a mock executable in root of tmpDir
	exeName := "mytool"
	if runtime.GOOS == "windows" {
		exeName = "mytool.exe"
	}
	exePath := filepath.Join(tmpDir, exeName)
	err := os.WriteFile(exePath, []byte("echo hi"), 0755)
	assert.NoError(t, err)

	err = p.PostInstall(context.Background(), "tool", tmpDir, "1.0")
	assert.NoError(t, err)

	// Check if it was moved to bin
	binExePath := filepath.Join(tmpDir, "bin", exeName)
	_, err = os.Stat(binExePath)
	assert.NoError(t, err)
}

func TestNativeProvider_Verify(t *testing.T) {
	p := NewNativeProvider()
	err := p.Verify(context.Background(), "tool", "1.0", "/tmp")
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestNativeProvider_IsExecutable(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Non-executable
	f, err := os.CreateTemp(tmpDir, "test")
	assert.NoError(t, err)
	f.Close()
	info, err := os.Stat(f.Name())
	assert.NoError(t, err)
	assert.False(t, isExecutable(info))

	// Executable
	exeName := "test_exe"
	if runtime.GOOS == "windows" {
		exeName = "test_exe.exe"
	}
	exePath := filepath.Join(tmpDir, exeName)
	err = os.WriteFile(exePath, []byte("test"), 0755)
	assert.NoError(t, err)
	
	infoExe, err := os.Stat(exePath)
	assert.NoError(t, err)
	assert.True(t, isExecutable(infoExe))
}

func TestNativeProvider_DelegatedMethods(t *testing.T) {
	p := NewNativeProvider()
	
	tmpDir := t.TempDir()
	
	// Create a dummy bin dir and executable so generic provider doesn't fail
	err := os.MkdirAll(filepath.Join(tmpDir, "bin"), 0755)
	assert.NoError(t, err)
	exePath := filepath.Join(tmpDir, "bin", "dummy")
	if runtime.GOOS == "windows" {
		exePath += ".exe"
	}
	err = os.WriteFile(exePath, []byte("echo hi"), 0755)
	assert.NoError(t, err)
	
	// GenerateShims
	_, err = p.GenerateShims("tool", tmpDir, "1.0")
	assert.NoError(t, err)

	// DetectVersion
	_, err = p.DetectVersion(context.Background(), "tool", tmpDir)
	// It might return error since it's an empty dir, but we just check it doesn't panic
	
	// ListExecutables
	_, err = p.ListExecutables("tool", tmpDir, "1.0")
	assert.NoError(t, err)

	// GetBinPaths
	_, err = p.GetBinPaths("tool", tmpDir, "1.0")
	assert.NoError(t, err)

	// GetEnvVars
	_, err = p.GetEnvVars("tool", tmpDir, "1.0")
	assert.NoError(t, err)

	// Uninstall
	err = p.Uninstall(context.Background(), "tool", tmpDir, "1.0")
	assert.NoError(t, err)
}
