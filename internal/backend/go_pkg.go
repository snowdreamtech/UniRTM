// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

// GoPkgBackend implements the backend for go packages.
// It embeds GoBackend to reuse its version resolution logic (proxy.golang.org).
type GoPkgBackend struct {
	*GoBackend
}

func NewGoPkgBackend() *GoPkgBackend {
	return &GoPkgBackend{
		GoBackend: NewGoBackend(),
	}
}

func (b *GoPkgBackend) Name() string {
	return "go-pkg"
}
