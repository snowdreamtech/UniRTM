// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package provider

import (
	"testing"
)

func TestRegistry_RegisterAndGet(t *testing.T) {
	r := NewRegistry()
	
	// Create a dummy provider
	p := NewGenericProvider()
	
	r.Register("dummytool", p)
	
	retrieved, err := r.GetExact("dummytool")
	if err != nil {
		t.Fatalf("expected to get dummytool, got error: %v", err)
	}
	if retrieved.Name() != p.Name() {
		t.Errorf("expected %s, got %s", p.Name(), retrieved.Name())
	}
	
	_, err = r.GetExact("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent tool")
	}
}

func TestRegistry_GetWithFallback(t *testing.T) {
	r := NewRegistry()
	p := r.Get("nonexistent")
	if p == nil {
		t.Fatal("expected generic provider fallback, got nil")
	}
	if p.Name() != "generic" {
		t.Errorf("expected generic provider, got %s", p.Name())
	}
}

func TestRegistry_GetWithBackend(t *testing.T) {
	r := NewRegistry()
	
	// 'node' is registered to NodeProvider
	p1 := r.GetWithBackend("mytool", "node")
	if p1.Name() != "node" {
		t.Errorf("expected node provider, got %s", p1.Name())
	}
	
	// 'npm:prettier' should resolve to npm
	p2 := r.Get("npm:prettier")
	if p2.Name() != "npm" {
		t.Errorf("expected npm provider, got %s", p2.Name())
	}
}

func TestRegistry_HasAndUnregister(t *testing.T) {
	r := NewRegistry()
	p := NewGenericProvider()
	r.Register("dummytool", p)
	
	if !r.Has("dummytool") {
		t.Error("expected registry to have dummytool")
	}
	
	r.Unregister("dummytool")
	
	if r.Has("dummytool") {
		t.Error("expected registry to not have dummytool after unregister")
	}
}

func TestDefaultRegistry(t *testing.T) {
	p := NewGenericProvider()
	
	Register("dummytool2", p)
	
	if !Has("dummytool2") {
		t.Error("expected default registry to have dummytool2")
	}
	
	retrieved := Get("dummytool2")
	if retrieved.Name() != "generic" { // 'generic' is the name of GenericProvider
		t.Errorf("expected generic, got %s", retrieved.Name())
	}
	
	Unregister("dummytool2")
	
	if Has("dummytool2") {
		t.Error("expected default registry to not have dummytool2 after unregister")
	}
}
