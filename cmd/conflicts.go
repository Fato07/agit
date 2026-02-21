package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	gitops "github.com/fathindos/agit/internal/git"
	"github.com/fathindos/agit/internal/registry"
	"github.com/fathindos/agit/internal/ui"
)

type conflictJSON struct {
	Repo      string           `json:"repo"`
	File      string           `json:"file"`
	Worktrees []conflictWtJSON `json:"worktrees"`
}

type conflictWtJSON struct {
	ID    string `json:"id"`
	Agent string `json:"agent,omitempty"`
	Task  string `json:"task,omitempty"`
}

var conflictsCmd = &cobra.Command{
	Use:   "conflicts [repo]",
	Short: "Check for overlapping file changes across worktrees",
	Long: `Scans all active worktrees and detects files that have been modified
in more than one worktree, indicating potential merge conflicts.`,
	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: completeRepoNames,
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := registry.Open()
		if err != nil {
			return fmt.Errorf("could not open registry: %w", err)
		}
		defer db.Close()

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

		var allConflicts []conflictJSON
		totalConflicts := 0

		for _, repo := range repos {
			activeStatus := "active"
			worktrees, err := db.ListWorktrees(repo.ID, &activeStatus)
			if err != nil {
				return err
			}

			if len(worktrees) < 2 {
				if !ui.IsJSON() {
					ui.Info("%s: < 2 active worktrees, no conflicts possible", repo.Name)
				}
				continue
			}

			if !ui.IsJSON() {
				ui.Info("Scanning %d active worktrees in %s...", len(worktrees), repo.Name)
				ui.Blank()
			}

			// Update file touches for each worktree
			for _, wt := range worktrees {
				files, err := gitops.ModifiedFilesWithStatus(repo.Path, repo.DefaultBranch, wt.Branch)
				if err != nil {
					if !ui.IsJSON() {
						ui.Warning("could not get diff for %s: %v", wt.ID[:8], err)
					}
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
				if !ui.IsJSON() {
					ui.Success("No conflicts detected in %s", repo.Name)
					ui.Blank()
				}
				continue
			}

			for _, c := range conflicts {
				if ui.IsJSON() {
					cj := conflictJSON{Repo: repo.Name, File: c.FilePath}
					for i, wtID := range c.Worktrees {
						wj := conflictWtJSON{ID: wtID[:12]}
						if i < len(c.AgentIDs) && c.AgentIDs[i] != "" {
							agent, err := db.GetAgent(c.AgentIDs[i])
							if err == nil {
								wj.Agent = agent.Name
							}
						}
						if i < len(c.TaskDescs) && c.TaskDescs[i] != "" {
							wj.Task = c.TaskDescs[i]
						}
						cj.Worktrees = append(cj.Worktrees, wj)
					}
					allConflicts = append(allConflicts, cj)
				} else {
					fmt.Printf("%s %s\n", ui.T.Warning("CONFLICT:"), c.FilePath)
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
						desc := ui.T.Muted(wtID[:12])
						if agentStr != "" {
							desc = fmt.Sprintf("%s (%s: %s)", ui.T.Muted(wtID[:12]), agentStr, taskStr)
						}
						fmt.Printf("  Modified in: %s\n", desc)
					}
					ui.Blank()
				}
			}

			totalConflicts += len(conflicts)
		}

		if ui.IsJSON() {
			return ui.RenderJSON(allConflicts)
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
