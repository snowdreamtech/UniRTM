package task

import (
	"context"
	"testing"

	"github.com/snowdreamtech/unirtm/internal/config"
)

func TestNativeRunner_runTaskWithGraph(t *testing.T) {
	r := &NativeRunner{
		tasks: map[string]config.Task{
			"t1":      {Depends: []string{"t2"}},
			"t2":      {Depends: []string{"t3"}},
			"t3":      {Run: "echo t3"},
			"cycle1":  {Depends: []string{"cycle2"}},
			"cycle2":  {Depends: []string{"cycle1"}},
			"bad_dep": {Depends: []string{"nonexistent"}},
		},
	}

	ctx := context.Background()

	// success
	err := r.Run(ctx, ".", "t1", nil, nil)
	if err != nil {
		t.Errorf("expected no error running t1")
	}

	// cycle
	err = r.Run(ctx, ".", "cycle1", nil, nil)
	if err == nil {
		t.Errorf("expected error for circular dependency")
	}

	// task not found
	err = r.Run(ctx, ".", "nonexistent", nil, nil)
	if err == nil {
		t.Errorf("expected error for nonexistent task")
	}

	// dependency not found
	err = r.Run(ctx, ".", "bad_dep", nil, nil)
	if err == nil {
		t.Errorf("expected error for missing dependency")
	}
}
