// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"testing"
)

func TestResolveLuaBinariesURL(t *testing.T) {
	tests := []struct {
		name    string
		version string
		goarch  string
		want    string
		wantErr bool
	}{
		{
			name:    "Windows x64",
			version: "5.4.2",
			goarch:  "amd64",
			want:    "https://sourceforge.net/projects/luabinaries/files/5.4.2/Tools%20Executables/lua-5.4.2_Win64_bin.zip/download",
			wantErr: false,
		},
		{
			name:    "Windows x86",
			version: "5.4.2",
			goarch:  "386",
			want:    "https://sourceforge.net/projects/luabinaries/files/5.4.2/Tools%20Executables/lua-5.4.2_Win32_bin.zip/download",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveLuaBinariesURL(tt.version, tt.goarch)
			if (err != nil) != tt.wantErr {
				t.Errorf("resolveLuaBinariesURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("resolveLuaBinariesURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
