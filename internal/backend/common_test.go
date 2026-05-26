// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCalculateAssetScore(t *testing.T) {
	p := Platform{OS: "linux", Arch: "amd64"}

	tests := []struct {
		name      string
		assetName string
		tool      string
		minScore  int
	}{
		{"Linux AMD64 zip", "tool-linux-amd64.zip", "tool", 240},
		{"Linux x86_64 tar.gz", "tool-linux-x86_64.tar.gz", "tool", 250},
		{"Linux ARM64", "tool-linux-arm64.zip", "tool", -1}, // arch mismatch
		{"Windows zip", "tool-windows-amd64.zip", "tool", -1}, // os mismatch
		{"Hard exclude", "tool-linux-amd64.zip.sha256", "tool", -1},
		{"Source exclude", "tool-source.tar.gz", "tool", -1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			score := CalculateAssetScore(tc.assetName, p, tc.tool)
			if tc.minScore == -1 {
				if score != -1 {
					t.Errorf("expected score -1 for %s, got %d", tc.assetName, score)
				}
			} else {
				if score < tc.minScore {
					t.Errorf("expected score >= %d for %s, got %d", tc.minScore, tc.assetName, score)
				}
			}
		})
	}
}

func TestFindBestAsset(t *testing.T) {
	assets := []CommonAsset{
		{Name: "tool-windows-amd64.zip", URL: "url1"},
		{Name: "tool-linux-arm64.tar.gz", URL: "url2"},
		{Name: "tool-linux-amd64.tar.gz", URL: "url3"},
	}

	p := Platform{OS: "linux", Arch: "amd64"}
	best, score := FindBestAsset(assets, p, "tool")
	if best == nil || score <= 0 {
		t.Fatalf("failed to find best asset")
	}
	if best.URL != "url3" {
		t.Errorf("expected url3, got %s", best.URL)
	}
}

func TestProbeURL(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/good" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	client := ts.Client()
	if !ProbeURL(context.Background(), client, ts.URL+"/good") {
		t.Error("expected /good to be accessible")
	}
	if ProbeURL(context.Background(), client, ts.URL+"/bad") {
		t.Error("expected /bad to be inaccessible")
	}
}

func TestFetchAndParseChecksumFile(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/checksums.txt" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("abcd  file1.txt\nefgh *file2.txt\n"))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	checksums, err := FetchAndParseChecksumFile(context.Background(), ts.Client(), ts.URL+"/checksums.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if checksums["file1.txt"] != "abcd" {
		t.Errorf("expected abcd for file1.txt, got %s", checksums["file1.txt"])
	}
	if checksums["file2.txt"] != "efgh" {
		t.Errorf("expected efgh for file2.txt, got %s", checksums["file2.txt"])
	}
}
