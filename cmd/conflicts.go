package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	gitops "github.com/fathindos/agit/internal/git"
	"github.com/fathindos/agit/internal/registry"
)

var conflictsCmd = &cobra.Command{
	Use:   "conflicts [repo]",
	Short: "Check for overlapping file changes across worktrees",
	Long: `Scans all active worktrees and detects files that have been modified
in more than one worktree, indicating potential merge conflicts.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := registry.Open()
		if err != nil {
			return fmt.Errorf("could not open registry: %w", err)
		}
		defer db.Close()

		yellow := color.New(color.FgYellow).SprintFunc()
		green := color.New(color.FgGreen).SprintFunc()
		gray := color.New(color.FgHiBlack).SprintFunc()

		var repos []*registry.Repo
		if len(args) > 0 {
			repo, err := db.GetRepo(args[0])
			if err != nil {
				return err
			}
			repos = []*registry.Repo{repo}
		} else {
			repos, err = db.ListRepos()
			if err != nil {
				return err
			}
		}

		totalConflicts := 0

		for _, repo := range repos {
			activeStatus := "active"
			worktrees, err := db.ListWorktrees(repo.ID, &activeStatus)
			if err != nil {
				return err
			}

			if len(worktrees) < 2 {
				fmt.Printf("%s: %s\n", repo.Name, gray("< 2 active worktrees, no conflicts possible"))
				continue
			}

			fmt.Printf("Scanning %d active worktrees in %s...\n\n", len(worktrees), repo.Name)

			// Update file touches for each worktree
			for _, wt := range worktrees {
				files, err := gitops.ModifiedFilesWithStatus(repo.Path, repo.DefaultBranch, wt.Branch)
				if err != nil {
					fmt.Printf("  Warning: could not get diff for %s: %v\n", wt.ID[:8], err)
					continue
				}

				var touches []registry.FileTouch
				for path, changeType := range files {
					touches = append(touches, registry.FileTouch{
						FilePath:   path,
						ChangeType: changeType,
					})
				}
				db.RecordFileTouches(repo.ID, wt.ID, touches)
			}

			// Find conflicts
			conflicts, err := db.FindConflicts(repo.ID)
			if err != nil {
				return fmt.Errorf("could not check conflicts: %w", err)
			}

			if len(conflicts) == 0 {
				fmt.Printf("%s No conflicts detected in %s\n\n", green("✓"), repo.Name)
				continue
			}

			for _, c := range conflicts {
				fmt.Printf("%s %s\n", yellow("CONFLICT:"), c.FilePath)
				for i, wtID := range c.Worktrees {
					agentStr := ""
					if i < len(c.AgentIDs) && c.AgentIDs[i] != "" {
						agent, err := db.GetAgent(c.AgentIDs[i])
						if err == nil {
							agentStr = agent.Name
						}
					}
					taskStr := ""
					if i < len(c.TaskDescs) && c.TaskDescs[i] != "" {
						taskStr = c.TaskDescs[i]
					}
					desc := gray(wtID[:12])
					if agentStr != "" {
						desc = fmt.Sprintf("%s (%s: %s)", gray(wtID[:12]), agentStr, taskStr)
					}
					fmt.Printf("  Modified in: %s\n", desc)
				}
				fmt.Println()
			}

			totalConflicts += len(conflicts)
		}

		if totalConflicts > 0 {
			fmt.Printf("%d conflict(s) detected.\n", totalConflicts)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(conflictsCmd)
}
