// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package frontend

import (
	"embed"
	"io/fs"
)

// DistFS embeds the distribution assets produced by Vite.
//
//go:embed dist
var DistFS embed.FS

// GetFS returns an fs.FS rooted at the "dist" directory.
func GetFS() (fs.FS, error) {
	return fs.Sub(DistFS, "dist")
}
