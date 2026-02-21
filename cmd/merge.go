package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	apperrors "github.com/fathindos/agit/internal/errors"
	gitops "github.com/fathindos/agit/internal/git"
	"github.com/fathindos/agit/internal/registry"
	"github.com/fathindos/agit/internal/ui"
	"github.com/fathindos/agit/internal/ui/interactive"
)

var mergeCmd = &cobra.Command{
	Use:   "merge [worktree-id]",
	Short: "Merge a worktree branch back into the base branch",
	Long: `Merges the worktree's branch into the repository's default branch.
Runs a conflict check first unless --skip-conflict-check is set.

With -i (interactive), presents a selector if no worktree ID is specified.`,
	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: completeWorktreeIDs,
	RunE: func(cmd *cobra.Command, args []string) error {
		isInteractive, _ := cmd.Flags().GetBool("interactive")
		skipCheck, _ := cmd.Flags().GetBool("skip-conflict-check")
		cleanup, _ := cmd.Flags().GetBool("cleanup")

		db, err := registry.Open()
		if err != nil {
			return fmt.Errorf("could not open registry: %w", err)
		}
		defer db.Close()

		var wt *registry.Worktree

		if len(args) > 0 {
			worktreeID := args[0]
			wt, err = db.GetWorktree(worktreeID)
			if err != nil {
				repos, repoErr := db.ListRepos()
				if repoErr != nil {
					return err
				}
				for _, r := range repos {
					if found, findErr := db.FindWorktreeByPrefix(r.ID, worktreeID); findErr == nil {
						wt = found
						break
					}
				}
				if wt == nil {
					return err
				}
			}
		} else if isInteractive {
			worktrees, err := db.ListAllActiveWorktrees()
			if err != nil {
				return err
			}
			if len(worktrees) == 0 {
				fmt.Println("No active worktrees to merge.")
				return nil
			}
			var items []interactive.Item
			for _, w := range worktrees {
				desc := w.Branch
				if w.TaskDescription != nil {
					desc += " - " + *w.TaskDescription
				}
				items = append(items, interactive.Item{
					ID:    w.ID,
					Label: w.ID[:12],
					Desc:  desc,
				})
			}
			selected, err := interactive.Select("Select a worktree to merge:", items)
			if err != nil {
				return err
			}
			wt, err = db.GetWorktree(selected.ID)
			if err != nil {
				return err
			}
		} else {
			return apperrors.NewUserError("requires a worktree-id argument (or use -i for interactive mode)")
		}

		repo, err := db.GetRepoByID(wt.RepoID)
		if err != nil {
			return err
		}

		// Pre-merge conflict check
		if !skipCheck {
			canMerge, err := gitops.CanMergeCleanly(repo.Path, wt.Branch)
			if err != nil {
				ui.Warning("Could not check merge compatibility: %v", err)
			} else if !canMerge {
				return apperrors.NewUserError("merge would produce conflicts. Use --skip-conflict-check to force, or resolve manually")
			}
		}

		// Checkout default branch and merge
		if err := gitops.CheckoutBranch(repo.Path, repo.DefaultBranch); err != nil {
			return fmt.Errorf("could not checkout %s: %w", repo.DefaultBranch, err)
		}

		if err := gitops.MergeBranch(repo.Path, wt.Branch); err != nil {
			return err
		}

		db.UpdateWorktreeStatus(wt.ID, "completed")

		if ui.IsJSON() {
			result := map[string]string{
				"status":  "ok",
				"message": "merged",
				"branch":  wt.Branch,
				"into":    repo.DefaultBranch,
			}
			if cleanup {
				gitops.RemoveWorktree(repo.Path, wt.Path)
				gitops.DeleteBranch(repo.Path, wt.Branch)
				db.DeleteWorktree(wt.ID)
				result["cleanup"] = "done"
			}
			return ui.RenderJSON(result)
		}

		ui.Success("Merged %s into %s", wt.Branch, repo.DefaultBranch)

		if cleanup {
			gitops.RemoveWorktree(repo.Path, wt.Path)
			gitops.DeleteBranch(repo.Path, wt.Branch)
			db.DeleteWorktree(wt.ID)
			ui.Success("Cleaned up worktree and branch")
		}

		return nil
	},
}

func init() {
	mergeCmd.Flags().Bool("skip-conflict-check", false, "Skip pre-merge conflict check")
	mergeCmd.Flags().Bool("cleanup", false, "Remove worktree and branch after merge")
	rootCmd.AddCommand(mergeCmd)
}
