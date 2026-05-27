package service

import (
	"context"
	"testing"
	"os"

	"github.com/stretchr/testify/require"
)

func Test_tryVerifyProvenance(t *testing.T) {
	// Set UNIRTM_VERIFY_PROVENANCE to false
	os.Setenv("UNIRTM_VERIFY_PROVENANCE", "false")
	status, err := tryVerifyProvenance(context.Background(), "github", "node", "/tmp/some-nonexistent-path")
	require.NoError(t, err)
	require.Equal(t, "skipped", status)
	os.Unsetenv("UNIRTM_VERIFY_PROVENANCE")

	// Set MISE_VERIFY_PROVENANCE to false
	os.Setenv("MISE_VERIFY_PROVENANCE", "false")
	status, err = tryVerifyProvenance(context.Background(), "github", "node", "/tmp/some-nonexistent-path")
	require.NoError(t, err)
	require.Equal(t, "skipped", status)
	os.Unsetenv("MISE_VERIFY_PROVENANCE")

	// Default, no provenance info available
	status, err = tryVerifyProvenance(context.Background(), "unknown", "node", "/tmp/some-nonexistent-path")
	require.NoError(t, err)
	require.Equal(t, "not_applicable", status)


	// Test github backend with non-existent file to trigger error
	status, err = tryVerifyProvenance(context.Background(), "github", "foo/bar", "/tmp/non-existent-artifact.tar.gz")
	require.Error(t, err)
	require.Equal(t, "failed", status)

	// Test gitlab backend with non-existent file to trigger error
	status, err = tryVerifyProvenance(context.Background(), "gitlab", "foo/bar", "/tmp/non-existent-artifact.tar.gz")
	require.Error(t, err)
	require.Equal(t, "failed", status)
}
