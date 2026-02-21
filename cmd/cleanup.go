package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	gitops "github.com/fathindos/agit/internal/git"
	"github.com/fathindos/agit/internal/registry"
	"github.com/fathindos/agit/internal/ui"
	"github.com/fathindos/agit/internal/ui/interactive"
)

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Remove completed or stale worktrees",
	Long: `Cleans up worktrees that are completed, stale, or no longer needed.

With -i (interactive), presents a multi-select list of eligible worktrees
and asks for confirmation before removal.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")
		staleOnly, _ := cmd.Flags().GetBool("stale")
		isInteractive, _ := cmd.Flags().GetBool("interactive")

		db, err := registry.Open()
		if err != nil {
			return fmt.Errorf("could not open registry: %w", err)
		}
		defer db.Close()

		repos, err := db.ListRepos()
		if err != nil {
			return err
		}

		// Prune orphaned worktrees first
		for _, repo := range repos {
			db.PruneOrphanedWorktrees(repo.ID)
		}

		if isInteractive {
			return cleanupInteractive(db, repos)
		}

		type removedEntry struct {
			ID     string `json:"id"`
			Repo   string `json:"repo"`
			Status string `json:"status"`
		}
		var removedItems []removedEntry

		removed := 0
		for _, repo := range repos {
			worktrees, err := db.ListWorktrees(repo.ID, nil)
			if err != nil {
				continue
			}

			for _, wt := range worktrees {
				shouldRemove := false
				if all {
					shouldRemove = wt.Status == "completed" || wt.Status == "stale"
				} else if staleOnly {
					shouldRemove = wt.Status == "stale"
				} else {
					shouldRemove = wt.Status == "completed" || wt.Status == "stale"
				}

				if !shouldRemove {
					continue
				}

				if err := gitops.RemoveWorktree(repo.Path, wt.Path); err != nil {
					fmt.Fprintf(os.Stderr, "  Warning: could not remove worktree %s: %v\n", wt.ID[:12], err)
				}
				gitops.DeleteBranch(repo.Path, wt.Branch)
				db.DeleteWorktree(wt.ID)

				removedItems = append(removedItems, removedEntry{
					ID:     wt.ID[:12],
					Repo:   repo.Name,
					Status: wt.Status,
				})

				if !ui.IsJSON() {
					fmt.Printf("  Removed: %s (%s) - %s\n", ui.T.Muted(wt.ID[:12]), repo.Name, ui.StatusColor(wt.Status))
				}
				removed++
			}
		}

		if ui.IsJSON() {
			return ui.RenderJSON(map[string]interface{}{
				"status":  "ok",
				"removed": removedItems,
				"count":   removed,
			})
		}

		if removed == 0 {
			fmt.Println("Nothing to clean up.")
		} else {
			ui.Blank()
			ui.Success("Cleaned up %d worktree(s).", removed)
		}

		return nil
	},
}

func cleanupInteractive(db *registry.DB, repos []*registry.Repo) error {
	// Collect all eligible worktrees
	type candidate struct {
		wt   *registry.Worktree
		repo *registry.Repo
	}
	var candidates []candidate

	for _, repo := range repos {
		worktrees, err := db.ListWorktrees(repo.ID, nil)
		if err != nil {
			continue
		}
		for _, wt := range worktrees {
			if wt.Status == "completed" || wt.Status == "stale" {
				candidates = append(candidates, candidate{wt: wt, repo: repo})
			}
		}
	}

	if len(candidates) == 0 {
		fmt.Println("Nothing to clean up.")
		return nil
	}

	var items []interactive.Item
	for _, c := range candidates {
		items = append(items, interactive.Item{
			ID:    c.wt.ID,
			Label: fmt.Sprintf("%s (%s)", c.wt.ID[:12], c.repo.Name),
			Desc:  fmt.Sprintf("[%s] %s", c.wt.Status, c.wt.Branch),
		})
	}

	selected, err := interactive.MultiSelect("Select worktrees to clean up:", items)
	if err != nil {
		return err
	}

	if len(selected) == 0 {
		fmt.Println("No worktrees selected.")
		return nil
	}

	confirmed, err := interactive.Confirm(
		fmt.Sprintf("Remove %d worktree(s)?", len(selected)),
		false,
	)
	if err != nil {
		return err
	}
	if !confirmed {
		fmt.Println("Cancelled.")
		return nil
	}

	selectedIDs := make(map[string]bool)
	for _, s := range selected {
		selectedIDs[s.ID] = true
	}

	removed := 0
	for _, c := range candidates {
		if !selectedIDs[c.wt.ID] {
			continue
		}
		if err := gitops.RemoveWorktree(c.repo.Path, c.wt.Path); err != nil {
			fmt.Fprintf(os.Stderr, "  Warning: could not remove worktree %s: %v\n", c.wt.ID[:12], err)
		}
		gitops.DeleteBranch(c.repo.Path, c.wt.Branch)
		db.DeleteWorktree(c.wt.ID)
		fmt.Printf("  Removed: %s (%s) - %s\n", ui.T.Muted(c.wt.ID[:12]), c.repo.Name, ui.StatusColor(c.wt.Status))
		removed++
	}

	ui.Blank()
	ui.Success("Cleaned up %d worktree(s).", removed)
	return nil
}

func init() {
	cleanupCmd.Flags().Bool("all", false, "Remove all completed and stale worktrees")
	cleanupCmd.Flags().Bool("stale", false, "Remove only stale worktrees")
	rootCmd.AddCommand(cleanupCmd)
}
