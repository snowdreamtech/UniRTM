// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func TestVerifyArtifactProvenance_MockServer(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "artifact.bin")
	if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	client := &http.Client{
		Transport: &mockCargoTransport{
			roundTripFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(`{"attestations": []}`)),
				}, nil
			},
		},
	}

	verifier := &provenanceVerifier{client: client}
	ctx := context.Background()

	res, err := verifier.verify(ctx, "", "owner", "repo", path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Supported {
		t.Errorf("expected Supported to be false")
	}
}
