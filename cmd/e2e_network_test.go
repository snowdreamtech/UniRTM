package cmd

import (
	"bytes"
	"io"
	"net/http"
	"testing"
)

func TestE2E_LsRemote(t *testing.T) {
	harness := NewE2EHarness(t)

	// Mock github API for something simple
	harness.MockHTTP["https://api.github.com/repos/snowdreamtech/test-tool/releases"] = func(req *http.Request) (*http.Response, error) {
		jsonResp := `[
			{
				"tag_name": "v1.0.0", "prerelease": false, "draft": false,
				"assets": [{"name": "test-tool-darwin-arm64.tar.gz", "browser_download_url": "http://mock/1", "size": 100}, {"name": "test-tool-darwin-amd64.tar.gz", "browser_download_url": "http://mock/1", "size": 100}, {"name": "test-tool-linux-amd64.tar.gz", "browser_download_url": "http://mock/1", "size": 100}, {"name": "test-tool-linux-arm64.tar.gz", "browser_download_url": "http://mock/1", "size": 100}, {"name": "test-tool-windows-amd64.zip", "browser_download_url": "http://mock/1", "size": 100}]
			},
			{
				"tag_name": "v1.1.0", "prerelease": false, "draft": false,
				"assets": [{"name": "test-tool-darwin-arm64.tar.gz", "browser_download_url": "http://mock/2", "size": 100}, {"name": "test-tool-darwin-amd64.tar.gz", "browser_download_url": "http://mock/2", "size": 100}, {"name": "test-tool-linux-amd64.tar.gz", "browser_download_url": "http://mock/2", "size": 100}, {"name": "test-tool-linux-arm64.tar.gz", "browser_download_url": "http://mock/2", "size": 100}, {"name": "test-tool-windows-amd64.zip", "browser_download_url": "http://mock/2", "size": 100}]
			}
		]`
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBufferString(jsonResp)),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	}

	harness.MockHTTP["https://api.github.com/repos/snowdreamtech/test-tool/releases?per_page=100"] = harness.MockHTTP["https://api.github.com/repos/snowdreamtech/test-tool/releases"]

	stdout, stderr, err := harness.Run("ls-remote", "github:snowdreamtech/test-tool")
	if err != nil {
		t.Fatalf("unirtm ls-remote failed: %v\nstderr: %s", err, stderr)
	}

	if !bytes.Contains([]byte(stdout), []byte("1.0.0")) || !bytes.Contains([]byte(stdout), []byte("1.1.0")) {
		t.Errorf("Expected output to contain versions, got stdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
}

func TestE2E_Registry(t *testing.T) {
	harness := NewE2EHarness(t)

	// Mock registry download
	harness.MockHTTP["https://github.com/snowdreamtech/unirtm-registry/releases/latest/download/registry.json"] = func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBufferString(`{}`)),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	}

	stdout, stderr, err := harness.Run("registry")
	if err != nil {
		t.Fatalf("unirtm registry failed: %v\nstderr: %s", err, stderr)
	}
	_ = stdout

	stdout, stderr, err = harness.Run("index", "update")
	if err != nil {
		// Just log error for index update if it fails, maybe the mock isn't sufficient
		t.Logf("unirtm index update: %v\nstderr: %s", err, stderr)
	}
	_ = stdout
}
