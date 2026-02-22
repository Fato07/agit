package conflicts

import (
	"fmt"
	"sort"
	"time"

	"github.com/fathindos/agit/internal/registry"
)

// Suggestion represents a recommended merge order entry for conflict resolution.
type Suggestion struct {
	Order            int       `json:"order"`
	WorktreeID       string    `json:"worktree_id"`
	ConflictingFiles int       `json:"conflicting_files"`
	CreatedAt        time.Time `json:"created_at"`
	Rationale        string    `json:"rationale"`
}

// SuggestResolutionOrder analyzes conflicts and worktrees to produce a
// recommended merge order. Worktrees with fewer conflicting files should
// be merged first (less risk). Ties are broken by creation time (older first).
func SuggestResolutionOrder(conflicts []registry.Conflict, worktrees []*registry.Worktree) []Suggestion {
	if len(conflicts) == 0 || len(worktrees) < 2 {
		return nil
	}

	// Build a set of worktree IDs involved in conflicts
	conflictCount := make(map[string]int)
	for _, c := range conflicts {
		for _, wtID := range c.Worktrees {
			conflictCount[wtID]++
		}
	}

	if len(conflictCount) < 2 {
		return nil
	}

	// Build worktree lookup
	wtMap := make(map[string]*registry.Worktree)
	for _, wt := range worktrees {
		wtMap[wt.ID] = wt
	}

	// Collect entries for conflicting worktrees
	type entry struct {
		wtID      string
		conflicts int
		createdAt time.Time
	}

	var entries []entry
	for wtID, count := range conflictCount {
		wt, ok := wtMap[wtID]
		if !ok {
			continue
		}
		entries = append(entries, entry{
			wtID:      wtID,
			conflicts: count,
			createdAt: wt.CreatedAt,
		})
	}

	// Sort: fewer conflicts first, then older creation time first
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].conflicts != entries[j].conflicts {
			return entries[i].conflicts < entries[j].conflicts
		}
		return entries[i].createdAt.Before(entries[j].createdAt)
	})

	suggestions := make([]Suggestion, len(entries))
	for i, e := range entries {
		rationale := fmt.Sprintf("Merge first: %d conflicting file(s)", e.conflicts)
		if i == 0 {
			rationale = fmt.Sprintf("Fewest conflicts (%d file(s)) — merge first to minimize risk", e.conflicts)
		}
		suggestions[i] = Suggestion{
			Order:            i + 1,
			WorktreeID:       e.wtID,
			ConflictingFiles: e.conflicts,
			CreatedAt:        e.createdAt,
			Rationale:        rationale,
		}
	}

	return suggestions
}
