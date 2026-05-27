package plugin

import (
	"context"
	"net"
	"net/rpc"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/backend"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockBackend struct {
	mock.Mock
}

func (m *mockBackend) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockBackend) ListVersions(ctx context.Context, tool string, platform backend.Platform) ([]backend.VersionInfo, error) {
	args := m.Called(ctx, tool, platform)
	if args.Get(0) != nil {
		return args.Get(0).([]backend.VersionInfo), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockBackend) ResolveVersion(ctx context.Context, tool string, versionRequest string, platform backend.Platform) (*backend.VersionInfo, error) {
	args := m.Called(ctx, tool, versionRequest, platform)
	if args.Get(0) != nil {
		return args.Get(0).(*backend.VersionInfo), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockBackend) GetDownloadInfo(ctx context.Context, tool string, version string, platform backend.Platform) (*backend.VersionInfo, error) {
	args := m.Called(ctx, tool, version, platform)
	if args.Get(0) != nil {
		return args.Get(0).(*backend.VersionInfo), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockBackend) SupportsChecksum() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *mockBackend) SupportsGPG() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *mockBackend) AttestationType() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockBackend) IsRecommended() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *mockBackend) IsScriptless() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *mockBackend) GetReach() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockBackend) IsStable() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *mockBackend) SupportsOffline() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *mockBackend) Dependencies() []string {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Get(0).([]string)
	}
	return nil
}

// Add dummy interface check
var _ backend.Backend = (*mockBackend)(nil)

func setupBackendRPCClientServer(t *testing.T, impl backend.Backend) *BackendRPCClient {
	server := rpc.NewServer()
	err := server.RegisterName("Plugin", &BackendRPCServer{Impl: impl})
	assert.NoError(t, err)

	clientConn, serverConn := net.Pipe()
	go server.ServeConn(serverConn)

	client := rpc.NewClient(clientConn)
	t.Cleanup(func() {
		client.Close()
		clientConn.Close()
		serverConn.Close()
	})

	return &BackendRPCClient{client: client}
}

func TestBackendRPC_Name(t *testing.T) {
	mockB := new(mockBackend)
	mockB.On("Name").Return("mock-backend")
	client := setupBackendRPCClientServer(t, mockB)

	name := client.Name()
	assert.Equal(t, "mock-backend", name)
	mockB.AssertExpectations(t)
}

func TestBackendRPC_ResolveVersion(t *testing.T) {
	mockB := new(mockBackend)
	expectedInfo := &backend.VersionInfo{
		Version:     "1.20",
		DownloadURL: "https://example.com/go1.20",
	}
	mockB.On("ResolveVersion", mock.Anything, "go", "latest", backend.Platform{OS: "linux", Arch: "amd64"}).
		Return(expectedInfo, nil)
	
	client := setupBackendRPCClientServer(t, mockB)

	info, err := client.ResolveVersion(context.Background(), "go", "latest", backend.Platform{OS: "linux", Arch: "amd64"})
	assert.NoError(t, err)
	assert.Equal(t, expectedInfo, info)
	mockB.AssertExpectations(t)
}

func TestBackendRPC_SupportsGPG(t *testing.T) {
	mockB := new(mockBackend)
	mockB.On("SupportsGPG").Return(true)
	
	client := setupBackendRPCClientServer(t, mockB)

	assert.True(t, client.SupportsGPG())
	mockB.AssertExpectations(t)
}

func TestBackendRPC_ListVersions(t *testing.T) {
	mockB := new(mockBackend)
	expected := []backend.VersionInfo{{Version: "1.0"}}
	mockB.On("ListVersions", mock.Anything, "go", backend.Platform{OS: "linux", Arch: "amd64"}).
		Return(expected, nil)
	client := setupBackendRPCClientServer(t, mockB)
	res, err := client.ListVersions(context.Background(), "go", backend.Platform{OS: "linux", Arch: "amd64"})
	assert.NoError(t, err)
	assert.Equal(t, expected, res)
	mockB.AssertExpectations(t)
}

func TestBackendRPC_GetDownloadInfo(t *testing.T) {
	mockB := new(mockBackend)
	expected := &backend.VersionInfo{Version: "1.0"}
	mockB.On("GetDownloadInfo", mock.Anything, "go", "1.0", backend.Platform{OS: "linux", Arch: "amd64"}).
		Return(expected, nil)
	client := setupBackendRPCClientServer(t, mockB)
	res, err := client.GetDownloadInfo(context.Background(), "go", "1.0", backend.Platform{OS: "linux", Arch: "amd64"})
	assert.NoError(t, err)
	assert.Equal(t, expected, res)
	mockB.AssertExpectations(t)
}

func TestBackendRPC_SupportsChecksum(t *testing.T) {
	mockB := new(mockBackend)
	mockB.On("SupportsChecksum").Return(true)
	client := setupBackendRPCClientServer(t, mockB)
	assert.True(t, client.SupportsChecksum())
	mockB.AssertExpectations(t)
}

func TestBackendRPC_AttestationType(t *testing.T) {
	mockB := new(mockBackend)
	mockB.On("AttestationType").Return("pgp")
	client := setupBackendRPCClientServer(t, mockB)
	assert.Equal(t, "pgp", client.AttestationType())
	mockB.AssertExpectations(t)
}

func TestBackendRPC_IsRecommended(t *testing.T) {
	mockB := new(mockBackend)
	mockB.On("IsRecommended").Return(true)
	client := setupBackendRPCClientServer(t, mockB)
	assert.True(t, client.IsRecommended())
	mockB.AssertExpectations(t)
}

func TestBackendRPC_IsScriptless(t *testing.T) {
	mockB := new(mockBackend)
	mockB.On("IsScriptless").Return(true)
	client := setupBackendRPCClientServer(t, mockB)
	assert.True(t, client.IsScriptless())
	mockB.AssertExpectations(t)
}

func TestBackendRPC_GetReach(t *testing.T) {
	mockB := new(mockBackend)
	mockB.On("GetReach").Return("global")
	client := setupBackendRPCClientServer(t, mockB)
	assert.Equal(t, "global", client.GetReach())
	mockB.AssertExpectations(t)
}

func TestBackendRPC_IsStable(t *testing.T) {
	mockB := new(mockBackend)
	mockB.On("IsStable").Return(true)
	client := setupBackendRPCClientServer(t, mockB)
	assert.True(t, client.IsStable())
	mockB.AssertExpectations(t)
}

func TestBackendRPC_SupportsOffline(t *testing.T) {
	mockB := new(mockBackend)
	mockB.On("SupportsOffline").Return(true)
	client := setupBackendRPCClientServer(t, mockB)
	assert.True(t, client.SupportsOffline())
	mockB.AssertExpectations(t)
}

func TestBackendRPC_Dependencies(t *testing.T) {
	mockB := new(mockBackend)
	expected := []string{"dep1"}
	mockB.On("Dependencies").Return(expected)
	client := setupBackendRPCClientServer(t, mockB)
	assert.Equal(t, expected, client.Dependencies())
	mockB.AssertExpectations(t)
}
