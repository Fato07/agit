package conflicts

import (
	"testing"
	"time"

	"github.com/fathindos/agit/internal/registry"
)

func TestSuggestResolutionOrderTwoWorktrees(t *testing.T) {
	now := time.Now()
	conflicts := []registry.Conflict{
		{FilePath: "a.go", Worktrees: []string{"wt-1", "wt-2"}},
		{FilePath: "b.go", Worktrees: []string{"wt-1", "wt-2"}},
		{FilePath: "c.go", Worktrees: []string{"wt-2"}}, // only wt-2 listed once (won't count as conflict alone, but comes from FindConflicts)
	}
	// wt-1 appears in 2 conflicts, wt-2 appears in 3 — wt-1 should be merged first
	worktrees := []*registry.Worktree{
		{ID: "wt-1", CreatedAt: now.Add(-time.Hour)},
		{ID: "wt-2", CreatedAt: now},
	}

	suggestions := SuggestResolutionOrder(conflicts, worktrees)
	if len(suggestions) != 2 {
		t.Fatalf("expected 2 suggestions, got %d", len(suggestions))
	}
	if suggestions[0].WorktreeID != "wt-1" {
		t.Errorf("expected wt-1 first (fewer conflicts), got %s", suggestions[0].WorktreeID)
	}
	if suggestions[0].Order != 1 {
		t.Errorf("expected order 1, got %d", suggestions[0].Order)
	}
	if suggestions[1].WorktreeID != "wt-2" {
		t.Errorf("expected wt-2 second, got %s", suggestions[1].WorktreeID)
	}
}

func TestSuggestResolutionOrderThreeWorktrees(t *testing.T) {
	now := time.Now()
	conflicts := []registry.Conflict{
		{FilePath: "a.go", Worktrees: []string{"wt-1", "wt-2", "wt-3"}},
		{FilePath: "b.go", Worktrees: []string{"wt-2", "wt-3"}},
	}
	// wt-1: 1 conflict, wt-2: 2 conflicts, wt-3: 2 conflicts
	worktrees := []*registry.Worktree{
		{ID: "wt-1", CreatedAt: now},
		{ID: "wt-2", CreatedAt: now.Add(-2 * time.Hour)},
		{ID: "wt-3", CreatedAt: now.Add(-time.Hour)},
	}

	suggestions := SuggestResolutionOrder(conflicts, worktrees)
	if len(suggestions) != 3 {
		t.Fatalf("expected 3 suggestions, got %d", len(suggestions))
	}
	// wt-1 first (1 conflict), then wt-2 (2 conflicts, older), then wt-3 (2 conflicts, newer)
	if suggestions[0].WorktreeID != "wt-1" {
		t.Errorf("expected wt-1 first, got %s", suggestions[0].WorktreeID)
	}
	if suggestions[1].WorktreeID != "wt-2" {
		t.Errorf("expected wt-2 second (older), got %s", suggestions[1].WorktreeID)
	}
	if suggestions[2].WorktreeID != "wt-3" {
		t.Errorf("expected wt-3 third, got %s", suggestions[2].WorktreeID)
	}
}

func TestSuggestResolutionOrderSingleWorktree(t *testing.T) {
	suggestions := SuggestResolutionOrder(
		[]registry.Conflict{{FilePath: "a.go", Worktrees: []string{"wt-1"}}},
		[]*registry.Worktree{{ID: "wt-1", CreatedAt: time.Now()}},
	)
	if suggestions != nil {
		t.Errorf("expected nil suggestions for single worktree, got %d", len(suggestions))
	}
}

func TestSuggestResolutionOrderNoConflicts(t *testing.T) {
	suggestions := SuggestResolutionOrder(
		nil,
		[]*registry.Worktree{{ID: "wt-1"}, {ID: "wt-2"}},
	)
	if suggestions != nil {
		t.Errorf("expected nil suggestions for no conflicts, got %d", len(suggestions))
	}
}

func TestSuggestResolutionOrderEqualConflictsSortByTime(t *testing.T) {
	now := time.Now()
	conflicts := []registry.Conflict{
		{FilePath: "a.go", Worktrees: []string{"wt-1", "wt-2"}},
	}
	// Both have 1 conflict; wt-2 is older so should be first
	worktrees := []*registry.Worktree{
		{ID: "wt-1", CreatedAt: now},
		{ID: "wt-2", CreatedAt: now.Add(-time.Hour)},
	}

	suggestions := SuggestResolutionOrder(conflicts, worktrees)
	if len(suggestions) != 2 {
		t.Fatalf("expected 2 suggestions, got %d", len(suggestions))
	}
	if suggestions[0].WorktreeID != "wt-2" {
		t.Errorf("expected wt-2 first (older), got %s", suggestions[0].WorktreeID)
	}
}
