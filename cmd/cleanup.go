package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	gitops "github.com/fathindos/agit/internal/git"
	"github.com/fathindos/agit/internal/registry"
)

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Remove completed or stale worktrees",
	Long:  `Cleans up worktrees that are completed, stale, or no longer needed.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")
		staleOnly, _ := cmd.Flags().GetBool("stale")

		db, err := registry.Open()
		if err != nil {
			return fmt.Errorf("could not open registry: %w", err)
		}
		defer db.Close()

		green := color.New(color.FgGreen).SprintFunc()
		gray := color.New(color.FgHiBlack).SprintFunc()

		repos, err := db.ListRepos()
		if err != nil {
			return err
		}

		removed := 0
		for _, repo := range repos {
			// Prune orphaned worktrees so they become available for cleanup
			db.PruneOrphanedWorktrees(repo.ID)

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

				// Remove git worktree
				if err := gitops.RemoveWorktree(repo.Path, wt.Path); err != nil {
					fmt.Fprintf(os.Stderr, "  Warning: could not remove worktree %s: %v\n", wt.ID[:12], err)
				}
				// Delete branch
				gitops.DeleteBranch(repo.Path, wt.Branch)
				// Remove from registry
				db.DeleteWorktree(wt.ID)

				fmt.Printf("  Removed: %s (%s) - %s\n", gray(wt.ID[:12]), repo.Name, wt.Status)
				removed++
			}
		}

		if removed == 0 {
			fmt.Println("Nothing to clean up.")
		} else {
			fmt.Printf("\n%s Cleaned up %d worktree(s).\n", green("✓"), removed)
		}

		return nil
	},
}

func init() {
	cleanupCmd.Flags().Bool("all", false, "Remove all completed and stale worktrees")
	cleanupCmd.Flags().Bool("stale", false, "Remove only stale worktrees")
	rootCmd.AddCommand(cleanupCmd)
}
