// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"
)

// S3Backend implements the Backend interface for Amazon S3 buckets.
type S3Backend struct {
	client *http.Client
}

// NewS3Backend creates a new S3 backend.
func NewS3Backend() *S3Backend {
	return &S3Backend{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (b *S3Backend) Name() string {
	return "s3"
}

func (b *S3Backend) ListVersions(ctx context.Context, tool string, platform Platform) ([]VersionInfo, error) {
	// S3 bucket listing requires parsing XML from ?list-type=2 API, which can be complex
	// For simplicity, we assume explicit version requests or a predefined versions.txt
	return nil, NewBackendError(b.Name(), tool, "listing versions from s3 bucket root is not supported dynamically", nil)
}

func (b *S3Backend) ResolveVersion(ctx context.Context, tool string, versionRequest string, platform Platform) (*VersionInfo, error) {
	if versionRequest == "latest" {
		return &VersionInfo{
			Version:  "latest",
			Platform: platform,
		}, nil
	}

	return &VersionInfo{
		Version:  versionRequest,
		Platform: platform,
	}, nil
}

func (b *S3Backend) GetDownloadInfo(ctx context.Context, tool string, version string, platform Platform) (*VersionInfo, error) {
	bucket := os.Getenv("UNIRTM_S3_BUCKET")
	if bucket == "" {
		bucket = "unirtm-binaries" // Fallback or could fail here
	}
	region := os.Getenv("UNIRTM_S3_REGION")
	if region == "" {
		region = "us-east-1"
	}

	// Example construction: https://bucket.s3.region.amazonaws.com/tool/version/os-arch/tool.tar.gz
	// In reality, this requires a standard layout in the S3 bucket.
	ext := "tar.gz"
	if platform.OS == "windows" {
		ext = "zip"
	}

	url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s/%s/%s-%s/%s.%s",
		bucket, region, tool, version, platform.OS, platform.Arch, tool, ext)

	return &VersionInfo{
		Version:     version,
		DownloadURL: url,
		Platform:    platform,
	}, nil
}

func (b *S3Backend) SupportsChecksum() bool {
	return true
}

func (b *S3Backend) SupportsGPG() bool {
	return false
}

func (b *S3Backend) AttestationType() string {
	return ""
}
