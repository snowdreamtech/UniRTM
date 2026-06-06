// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package hook

import (
	"context"
	"sort"
	"testing"
)

type mockRunner struct {
	name   string
	detect bool
	runErr error
	called bool
}

func (m *mockRunner) Name() string                                  { return m.name }
func (m *mockRunner) Detect(dir string) bool                        { return m.detect }
func (m *mockRunner) Install(ctx context.Context, dir string) error { return nil }
func (m *mockRunner) Run(ctx context.Context, hookName string, args []string) error {
	m.called = true
	return m.runErr
}

func TestRouterPriority(t *testing.T) {
	// Temporarily override runners
	originalRunners := runners
	defer func() { runners = originalRunners }()

	// Reset runners
	runners = nil

	mPrecommit := &mockRunner{name: "pre-commit", detect: true}
	mLefthook := &mockRunner{name: "lefthook", detect: true}
	mHusky := &mockRunner{name: "husky", detect: true}
	mNative := &mockRunner{name: "unirtm", detect: true}

	// Register in random order
	RegisterRunner(mNative)
	RegisterRunner(mLefthook)
	RegisterRunner(mPrecommit)
	RegisterRunner(mHusky)

	// Sort manually as Run does
	sort.SliceStable(runners, func(i, j int) bool {
		return getPriority(runners[i].Name()) > getPriority(runners[j].Name())
	})

	if runners[0].Name() != "lefthook" {
		t.Errorf("expected highest priority to be lefthook, got %s", runners[0].Name())
	}
	if runners[1].Name() != "husky" {
		t.Errorf("expected second priority to be husky, got %s", runners[1].Name())
	}
	if runners[2].Name() != "pre-commit" {
		t.Errorf("expected third priority to be pre-commit, got %s", runners[2].Name())
	}
	if runners[3].Name() != "unirtm" {
		t.Errorf("expected lowest priority to be unirtm, got %s", runners[3].Name())
	}
}

func TestRouterRunFallback(t *testing.T) {
	originalRunners := runners
	defer func() { runners = originalRunners }()
	runners = nil

	// Register runners that don't detect anything
	mPrecommit := &mockRunner{name: "pre-commit", detect: false}
	RegisterRunner(mPrecommit)

	err := Run(context.Background(), ".", "pre-commit", nil)
	if err != nil {
		t.Errorf("expected no error when no runners detect, got %v", err)
	}
	if mPrecommit.called {
		t.Errorf("expected runner not to be called")
	}
}
