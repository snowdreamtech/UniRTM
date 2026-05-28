// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package backend

import (
	"context"
	"testing"
)

type dummyBackend struct {
	name string
}

func (d *dummyBackend) Name() string           { return d.name }
func (d *dummyBackend) Dependencies() []string { return nil }
func (d *dummyBackend) ListVersions(ctx context.Context, tool string, platform Platform) ([]VersionInfo, error) {
	return nil, nil
}
func (d *dummyBackend) ResolveVersion(ctx context.Context, tool string, versionRequest string, platform Platform) (*VersionInfo, error) {
	return nil, nil
}
func (d *dummyBackend) GetDownloadInfo(ctx context.Context, tool string, version string, platform Platform) (*VersionInfo, error) {
	return nil, nil
}
func (d *dummyBackend) SupportsChecksum() bool  { return false }
func (d *dummyBackend) SupportsGPG() bool       { return false }
func (d *dummyBackend) AttestationType() string { return "" }
func (d *dummyBackend) IsRecommended() bool     { return false }
func (d *dummyBackend) IsScriptless() bool      { return true }
func (d *dummyBackend) GetReach() string        { return "Small" }
func (d *dummyBackend) IsStable() bool          { return false }
func (d *dummyBackend) SupportsOffline() bool   { return false }

func TestRegistry_RegisterAndGet(t *testing.T) {
	r := NewRegistry()
	b := &dummyBackend{name: "dummy"}

	r.Register(b)

	retrieved, err := r.Get("dummy")
	if err != nil {
		t.Fatalf("expected to get dummy backend, got error: %v", err)
	}
	if retrieved.Name() != "dummy" {
		t.Errorf("expected dummy, got %s", retrieved.Name())
	}

	_, err = r.Get("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent backend")
	}
}

func TestRegistry_HasAndUnregister(t *testing.T) {
	r := NewRegistry()
	b := &dummyBackend{name: "dummy"}
	r.Register(b)

	if !r.Has("dummy") {
		t.Error("expected registry to have dummy backend")
	}

	r.Unregister("dummy")

	if r.Has("dummy") {
		t.Error("expected registry to not have dummy backend after unregister")
	}
}

func TestRegistry_ListAndBackends(t *testing.T) {
	r := NewRegistry()
	// Clear the default backends for clean test
	r.backends = make(map[string]Backend)

	r.Register(&dummyBackend{name: "dummy1"})
	r.Register(&dummyBackend{name: "dummy2"})

	list := r.List()
	if len(list) != 2 {
		t.Errorf("expected 2 backends, got %d", len(list))
	}

	backends := r.Backends()
	if len(backends) != 2 {
		t.Errorf("expected 2 backends map, got %d", len(backends))
	}
	if _, ok := backends["dummy1"]; !ok {
		t.Error("expected dummy1 in backends map")
	}
}

func TestDefaultRegistry(t *testing.T) {
	b := &dummyBackend{name: "default_dummy"}

	Register(b)

	if !Has("default_dummy") {
		t.Error("expected default registry to have default_dummy")
	}

	retrieved, err := Get("default_dummy")
	if err != nil {
		t.Fatalf("expected to get default_dummy, got error: %v", err)
	}
	if retrieved.Name() != "default_dummy" {
		t.Errorf("expected default_dummy, got %s", retrieved.Name())
	}

	Unregister("default_dummy")

	if Has("default_dummy") {
		t.Error("expected default registry to not have default_dummy after unregister")
	}
}
