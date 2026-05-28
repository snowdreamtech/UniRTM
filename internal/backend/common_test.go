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
		{"Linux ARM64", "tool-linux-arm64.zip", "tool", -1},   // arch mismatch
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

func TestFindChecksumForAsset(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/checksums.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("deadbeef  app-darwin-amd64\n123456  app-linux-amd64\n"))
	})
	mux.HandleFunc("/checksums_error.txt", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := server.Client()

	assets := []CommonAsset{
		{Name: "app-darwin-amd64", URL: "http://example.com/app"},
		{Name: "checksums.txt", URL: server.URL + "/checksums.txt"},
	}

	target := &CommonAsset{Name: "app-darwin-amd64", URL: "http://example.com/app"}

	// Test success
	checksum, err := FindChecksumForAsset(context.Background(), client, assets, target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if checksum != "deadbeef" {
		t.Errorf("expected deadbeef, got %s", checksum)
	}

	// Test asset not found in checksum file
	target2 := &CommonAsset{Name: "app-windows-amd64", URL: "http://example.com/app"}
	checksum2, err := FindChecksumForAsset(context.Background(), client, assets, target2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if checksum2 != "" {
		t.Errorf("expected empty checksum, got %s", checksum2)
	}

	// Test fetch error
	assetsErr := []CommonAsset{
		{Name: "app-darwin-amd64", URL: "http://example.com/app"},
		{Name: "checksums_error.txt", URL: server.URL + "/checksums_error.txt"},
	}
	_, err = FindChecksumForAsset(context.Background(), client, assetsErr, target)
	if err == nil {
		t.Errorf("expected error for not found checksum file")
	}

	// Test nil target
	checksumNil, err := FindChecksumForAsset(context.Background(), client, assets, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if checksumNil != "" {
		t.Errorf("expected empty checksum, got %s", checksumNil)
	}

	// Test no checksum asset
	assetsNoChecksum := []CommonAsset{
		{Name: "app-darwin-amd64", URL: "http://example.com/app"},
	}
	checksumNoAsset, err := FindChecksumForAsset(context.Background(), client, assetsNoChecksum, target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if checksumNoAsset != "" {
		t.Errorf("expected empty checksum, got %s", checksumNoAsset)
	}
}

func TestFindGPGSignatureForAsset(t *testing.T) {
	assets := []CommonAsset{
		{Name: "app-darwin-amd64", URL: "http://example.com/app"},
		{Name: "app-darwin-amd64.sig", URL: "http://example.com/app.sig"},
	}

	target := &CommonAsset{Name: "app-darwin-amd64", URL: "http://example.com/app"}

	sig := FindGPGSignatureForAsset(assets, target)
	if sig == "" {
		t.Errorf("expected signature url, got empty")
	}

	targetNotFound := &CommonAsset{Name: "app-windows-amd64", URL: "http://example.com/app.exe"}
	sigNotFound := FindGPGSignatureForAsset(assets, targetNotFound)
	if sigNotFound != "" {
		t.Errorf("expected empty signature url, got %s", sigNotFound)
	}
}

func TestCalculateAssetScore_Linux(t *testing.T) {
	platform := Platform{OS: "linux", Arch: "amd64"}
	score := CalculateAssetScore("app-linux-amd64.tar.gz", platform, "app")
	if score <= 0 {
		t.Errorf("expected positive score, got %d", score)
	}
}
func TestCalculateAssetScore_Windows(t *testing.T) {
	platform := Platform{OS: "windows", Arch: "amd64"}
	score := CalculateAssetScore("app-windows-amd64.zip", platform, "app")
	if score <= 0 {
		t.Errorf("expected positive score, got %d", score)
	}

	score = CalculateAssetScore("app-windows-x86_64.exe", platform, "app")
	if score <= 0 {
		t.Errorf("expected positive score, got %d", score)
	}
}

func TestCalculateAssetScore_Darwin(t *testing.T) {
	platform := Platform{OS: "darwin", Arch: "amd64"}
	// darwin with osx
	score := CalculateAssetScore("app-osx-amd64.tar.gz", platform, "app")
	if score <= 0 {
		t.Errorf("expected positive score for osx, got %d", score)
	}
	// darwin with apple
	score = CalculateAssetScore("app-apple-amd64.zip", platform, "app")
	if score <= 0 {
		t.Errorf("expected positive score for apple, got %d", score)
	}
	// darwin with universal (amd64)
	score = CalculateAssetScore("app-darwin-universal.tar.gz", platform, "app")
	if score <= 0 {
		t.Errorf("expected positive score for universal on amd64, got %d", score)
	}
}

func TestCalculateAssetScore_DarwinArm64(t *testing.T) {
	platform := Platform{OS: "darwin", Arch: "arm64"}
	// arm64 with aarch64
	score := CalculateAssetScore("app-darwin-aarch64.tar.gz", platform, "app")
	if score <= 0 {
		t.Errorf("expected positive score for aarch64, got %d", score)
	}
	// arm64 with armv8
	score = CalculateAssetScore("app-darwin-armv8.zip", platform, "app")
	if score <= 0 {
		t.Errorf("expected positive score for armv8, got %d", score)
	}
	// darwin with universal (arm64)
	score = CalculateAssetScore("app-darwin-universal.tar.gz", platform, "app")
	if score <= 0 {
		t.Errorf("expected positive score for universal on arm64, got %d", score)
	}
}

func TestCalculateAssetScore_Linux386(t *testing.T) {
	platform := Platform{OS: "linux", Arch: "386"}
	score := CalculateAssetScore("app-linux-i386.tar.gz", platform, "app")
	if score <= 0 {
		t.Errorf("expected positive score for i386, got %d", score)
	}
	// 32bit
	score = CalculateAssetScore("app-linux-32bit.tar.gz", platform, "app")
	if score <= 0 {
		t.Errorf("expected positive score for 32bit, got %d", score)
	}
}

func TestCalculateAssetScore_ToolNameWithNegativeKeyword(t *testing.T) {
	// Tool name contains negative keyword - should not be excluded
	platform := Platform{OS: "linux", Arch: "amd64"}
	// "addlicense" contains "license"
	score := CalculateAssetScore("addlicense-linux-amd64", platform, "google/addlicense")
	if score <= 0 {
		t.Errorf("expected positive score when negative keyword in tool name, got %d", score)
	}
}

func TestCalculateAssetScore_RawBinary(t *testing.T) {
	// Raw binary (no extension)
	platform := Platform{OS: "linux", Arch: "amd64"}
	score := CalculateAssetScore("app-linux-amd64", platform, "app")
	if score <= 0 {
		t.Errorf("expected positive score for raw binary, got %d", score)
	}
}

func TestCalculateAssetScore_MuslPenalty(t *testing.T) {
	// musl builds get a penalty
	platform := Platform{OS: "linux", Arch: "amd64"}
	scoreMusl := CalculateAssetScore("app-linux-amd64-musl.tar.gz", platform, "app")
	scoreNormal := CalculateAssetScore("app-linux-amd64.tar.gz", platform, "app")
	if scoreMusl >= scoreNormal {
		t.Errorf("expected musl score (%d) < normal score (%d)", scoreMusl, scoreNormal)
	}
}

func TestCalculateAssetScore_TarXz(t *testing.T) {
	platform := Platform{OS: "linux", Arch: "amd64"}
	score := CalculateAssetScore("app-linux-amd64.tar.xz", platform, "app")
	if score <= 0 {
		t.Errorf("expected positive score for tar.xz, got %d", score)
	}
	// .txz
	score = CalculateAssetScore("app-linux-amd64.txz", platform, "app")
	if score <= 0 {
		t.Errorf("expected positive score for txz, got %d", score)
	}
}

func TestCalculateAssetScore_Win(t *testing.T) {
	// "win" substring
	platform := Platform{OS: "windows", Arch: "amd64"}
	score := CalculateAssetScore("app-win-amd64.zip", platform, "app")
	if score <= 0 {
		t.Errorf("expected positive score for win substring, got %d", score)
	}
}

func TestCalculateAssetScore_WindowsExe(t *testing.T) {
	platform := Platform{OS: "windows", Arch: "amd64"}
	// .exe with x86_64 arch
	score := CalculateAssetScore("app-x86_64.exe", platform, "app")
	if score <= 0 {
		t.Errorf("expected positive score for .exe with x86_64, got %d", score)
	}
	// amd64 in name
	score = CalculateAssetScore("app-windows-amd64.exe", platform, "app")
	if score <= 0 {
		t.Errorf("expected positive score for windows amd64 .exe, got %d", score)
	}
}

func TestCalculateAssetScore_Linux64bit(t *testing.T) {
	platform := Platform{OS: "linux", Arch: "amd64"}
	score := CalculateAssetScore("app-linux-64bit.zip", platform, "app")
	if score <= 0 {
		t.Errorf("expected positive score for 64bit, got %d", score)
	}
	// x64
	score = CalculateAssetScore("app-linux-x64.zip", platform, "app")
	if score <= 0 {
		t.Errorf("expected positive score for x64, got %d", score)
	}
}

func TestFetchAndParseChecksumFile_NotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	checksums, err := FetchAndParseChecksumFile(context.Background(), ts.Client(), ts.URL+"/not-found")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if checksums != nil {
		t.Errorf("expected nil for 404, got %v", checksums)
	}
}

func TestFetchAndParseChecksumFile_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	_, err := FetchAndParseChecksumFile(context.Background(), ts.Client(), ts.URL+"/error")
	if err == nil {
		t.Error("expected error for 500 status")
	}
}

func TestFindGPGSignatureForAsset_WithAsc(t *testing.T) {
	assets := []CommonAsset{
		{Name: "app-darwin-amd64", URL: "http://example.com/app"},
		{Name: "app-darwin-amd64.asc", URL: "http://example.com/app.asc"},
	}

	target := &CommonAsset{Name: "app-darwin-amd64", URL: "http://example.com/app"}
	sig := FindGPGSignatureForAsset(assets, target)
	if sig != "http://example.com/app.asc" {
		t.Errorf("expected asc signature url, got %s", sig)
	}
}

func TestFindGPGSignatureForAsset_NilTarget(t *testing.T) {
	assets := []CommonAsset{
		{Name: "app-darwin-amd64", URL: "http://example.com/app"},
	}

	sig := FindGPGSignatureForAsset(assets, nil)
	if sig != "" {
		t.Errorf("expected empty for nil target, got %s", sig)
	}
}

func TestProbeURL_InvalidURL(t *testing.T) {
	client := &http.Client{}
	result := ProbeURL(context.Background(), client, "not-a-url-%%%")
	if result {
		t.Error("expected false for invalid URL")
	}
}
