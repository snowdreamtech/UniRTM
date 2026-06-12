// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"testing"
)

func TestResolveLuaDownloadURL(t *testing.T) {
	tests := []struct {
		name    string
		version string
		goos    string
		goarch  string
		want    string
		wantErr bool
	}{
		{
			name:    "Windows x64",
			version: "5.4.2",
			goos:    "windows",
			goarch:  "amd64",
			want:    "https://sourceforge.net/projects/luabinaries/files/5.4.2/Tools%20Executables/lua-5.4.2_Win64_bin.zip/download",
			wantErr: false,
		},
		{
			name:    "Linux x64",
			version: "5.4.2",
			goos:    "linux",
			goarch:  "amd64",
			want:    "https://sourceforge.net/projects/luabinaries/files/5.4.2/Tools%20Executables/lua-5.4.2_Linux54_64_bin.tar.gz/download",
			wantErr: false,
		},
		{
			name:    "MacOS x64",
			version: "5.4.2",
			goos:    "darwin",
			goarch:  "amd64",
			want:    "https://sourceforge.net/projects/luabinaries/files/5.4.2/Tools%20Executables/lua-5.4.2_MacOS1015_64_bin.tar.gz/download",
			wantErr: false,
		},
		{
			name:    "MacOS arm64 (fallback to x64)",
			version: "5.4.2",
			goos:    "darwin",
			goarch:  "arm64",
			want:    "https://sourceforge.net/projects/luabinaries/files/5.4.2/Tools%20Executables/lua-5.4.2_MacOS1015_64_bin.tar.gz/download",
			wantErr: false,
		},
		{
			name:    "Unsupported OS",
			version: "5.4.2",
			goos:    "freebsd",
			goarch:  "amd64",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveLuaDownloadURL(tt.version, tt.goos, tt.goarch)
			if (err != nil) != tt.wantErr {
				t.Errorf("resolveLuaDownloadURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("resolveLuaDownloadURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
