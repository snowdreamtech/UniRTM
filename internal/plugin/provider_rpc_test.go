package plugin

import (
	"context"
	"net"
	"net/rpc"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockProvider struct {
	mock.Mock
}

func (m *mockProvider) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockProvider) Install(ctx context.Context, tool string, installPath string, artifactPath string, version string) error {
	args := m.Called(ctx, tool, installPath, artifactPath, version)
	return args.Error(0)
}

func (m *mockProvider) PostInstall(ctx context.Context, tool string, installPath string, version string) error {
	args := m.Called(ctx, tool, installPath, version)
	return args.Error(0)
}

func (m *mockProvider) GenerateShims(tool string, installPath string, version string) (map[string]string, error) {
	args := m.Called(tool, installPath, version)
	if args.Get(0) != nil {
		return args.Get(0).(map[string]string), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockProvider) DetectVersion(ctx context.Context, tool string, installPath string) (string, error) {
	args := m.Called(ctx, tool, installPath)
	return args.String(0), args.Error(1)
}

func (m *mockProvider) ListExecutables(tool string, installPath string, version string) ([]string, error) {
	args := m.Called(tool, installPath, version)
	if args.Get(0) != nil {
		return args.Get(0).([]string), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockProvider) GetBinPaths(tool string, installPath string, version string) ([]string, error) {
	args := m.Called(tool, installPath, version)
	if args.Get(0) != nil {
		return args.Get(0).([]string), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockProvider) GetEnvVars(tool string, installPath string, version string) (map[string]string, error) {
	args := m.Called(tool, installPath, version)
	if args.Get(0) != nil {
		return args.Get(0).(map[string]string), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockProvider) Uninstall(ctx context.Context, tool string, installPath string, version string) error {
	args := m.Called(ctx, tool, installPath, version)
	return args.Error(0)
}

// Add dummy interface check
var _ provider.Provider = (*mockProvider)(nil)

func setupRPCClientServer(t *testing.T, impl provider.Provider) *ProviderRPCClient {
	server := rpc.NewServer()
	err := server.RegisterName("Plugin", &ProviderRPCServer{Impl: impl})
	assert.NoError(t, err)

	clientConn, serverConn := net.Pipe()
	go server.ServeConn(serverConn)

	client := rpc.NewClient(clientConn)
	t.Cleanup(func() {
		client.Close()
		clientConn.Close()
		serverConn.Close()
	})

	return &ProviderRPCClient{client: client}
}

func TestProviderRPC_Name(t *testing.T) {
	mockP := new(mockProvider)
	mockP.On("Name").Return("mock-provider")
	client := setupRPCClientServer(t, mockP)

	name := client.Name()
	assert.Equal(t, "mock-provider", name)
	mockP.AssertExpectations(t)
}

func TestProviderRPC_Install(t *testing.T) {
	mockP := new(mockProvider)
	mockP.On("Install", mock.Anything, "go", "/path", "/artifact", "1.20").Return(nil)
	client := setupRPCClientServer(t, mockP)

	err := client.Install(context.Background(), "go", "/path", "/artifact", "1.20")
	assert.NoError(t, err)
	mockP.AssertExpectations(t)
}

func TestProviderRPC_GetBinPaths(t *testing.T) {
	mockP := new(mockProvider)
	mockP.On("GetBinPaths", "go", "/path", "1.20").Return([]string{"/path/bin"}, nil)
	client := setupRPCClientServer(t, mockP)

	paths, err := client.GetBinPaths("go", "/path", "1.20")
	assert.NoError(t, err)
	assert.Equal(t, []string{"/path/bin"}, paths)
	mockP.AssertExpectations(t)
}
