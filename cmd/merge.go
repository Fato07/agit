package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	gitops "github.com/fathindos/agit/internal/git"
	"github.com/fathindos/agit/internal/registry"
)

var mergeCmd = &cobra.Command{
	Use:   "merge <worktree-id>",
	Short: "Merge a worktree branch back into the base branch",
	Long: `Merges the worktree's branch into the repository's default branch.
Runs a conflict check first unless --skip-conflict-check is set.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		worktreeID := args[0]
		skipCheck, _ := cmd.Flags().GetBool("skip-conflict-check")
		cleanup, _ := cmd.Flags().GetBool("cleanup")

		db, err := registry.Open()
		if err != nil {
			return fmt.Errorf("could not open registry: %w", err)
		}
		defer db.Close()

		green := color.New(color.FgGreen).SprintFunc()
		yellow := color.New(color.FgYellow).SprintFunc()

		// Find worktree (support partial ID match)
		wt, err := db.GetWorktree(worktreeID)
		if err != nil {
			return err
		}

		repo, err := db.GetRepoByID(wt.RepoID)
		if err != nil {
			return err
		}

		// Pre-merge conflict check
		if !skipCheck {
			canMerge, err := gitops.CanMergeCleanly(repo.Path, wt.Branch)
			if err != nil {
				fmt.Printf("%s Could not check merge compatibility: %v\n", yellow("WARNING:"), err)
			} else if !canMerge {
				return fmt.Errorf("merge would produce conflicts. Use --skip-conflict-check to force, or resolve manually")
			}
		}

		// Checkout default branch and merge
		if err := gitops.CheckoutBranch(repo.Path, repo.DefaultBranch); err != nil {
			return fmt.Errorf("could not checkout %s: %w", repo.DefaultBranch, err)
		}

		if err := gitops.MergeBranch(repo.Path, wt.Branch); err != nil {
			return err
		}

		fmt.Printf("%s Merged %s into %s\n", green("✓"), wt.Branch, repo.DefaultBranch)

		// Update worktree status
		db.UpdateWorktreeStatus(wt.ID, "completed")

		// Cleanup if requested
		if cleanup {
			gitops.RemoveWorktree(repo.Path, wt.Path)
			gitops.DeleteBranch(repo.Path, wt.Branch)
			db.DeleteWorktree(wt.ID)
			fmt.Printf("%s Cleaned up worktree and branch\n", green("✓"))
		}

		return nil
	},
}

func init() {
	mergeCmd.Flags().Bool("skip-conflict-check", false, "Skip pre-merge conflict check")
	mergeCmd.Flags().Bool("cleanup", false, "Remove worktree and branch after merge")
	rootCmd.AddCommand(mergeCmd)
}
