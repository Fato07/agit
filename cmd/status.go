package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/fathindos/agit/internal/conflicts"
	"github.com/fathindos/agit/internal/registry"
	"github.com/fathindos/agit/internal/ui"
)

type statusJSON struct {
	Repos []statusRepoJSON `json:"repos"`
}

type statusRepoJSON struct {
	Name          string               `json:"name"`
	DefaultBranch string               `json:"default_branch"`
	Worktrees     []statusWorktreeJSON `json:"worktrees"`
	Conflicts     []statusConflictJSON `json:"conflicts,omitempty"`
	Tasks         []statusTaskJSON     `json:"tasks,omitempty"`
}

type statusWorktreeJSON struct {
	ID     string `json:"id"`
	Branch string `json:"branch"`
	Agent  string `json:"agent"`
	Task   string `json:"task,omitempty"`
}

type statusConflictJSON struct {
	File      string `json:"file"`
	Worktrees int    `json:"worktrees"`
}

type statusTaskJSON struct {
	Status      string `json:"status"`
	Description string `json:"description"`
	Agent       string `json:"agent"`
}

var statusCmd = &cobra.Command{
	Use:               "status [repo]",
	Short:             "Show active worktrees, agents, and conflicts",
	Long:              `Displays the current state of all registered repos or a specific repo.`,
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

		if len(repos) == 0 {
			if ui.IsJSON() {
				return ui.RenderJSON(statusJSON{Repos: []statusRepoJSON{}})
			}
			fmt.Println("No repositories registered. Add one with: agit add <path>")
			return nil
		}

		var jsonResult statusJSON

		for _, repo := range repos {
			var repoData statusRepoJSON
			if ui.IsJSON() {
				repoData = statusRepoJSON{
					Name:          repo.Name,
					DefaultBranch: repo.DefaultBranch,
				}
			}

			if !ui.IsJSON() {
				fmt.Printf("%s %s (%s)\n", ui.T.Bold("REPO:"), ui.T.Bold(repo.Name), repo.DefaultBranch)
			}

			// Prune orphaned worktrees before display
			pruned, _ := db.PruneOrphanedWorktrees(repo.ID)
			if pruned > 0 && !ui.IsJSON() {
				ui.Blank()
				ui.Warning("%d worktree(s) marked stale (directory removed)", pruned)
			}

			// Active worktrees
			activeStatus := "active"
			worktrees, err := db.ListWorktrees(repo.ID, &activeStatus)
			if err != nil {
				return err
			}

			if ui.IsJSON() {
				for _, wt := range worktrees {
					wtJSON := statusWorktreeJSON{
						ID:     wt.ID[:12],
						Branch: wt.Branch,
						Agent:  "-",
					}
					if wt.AgentID != nil {
						agent, err := db.GetAgent(*wt.AgentID)
						if err == nil {
							wtJSON.Agent = agent.Name
						}
					}
					if wt.TaskDescription != nil {
						wtJSON.Task = *wt.TaskDescription
					}
					repoData.Worktrees = append(repoData.Worktrees, wtJSON)
				}
			} else if len(worktrees) > 0 {
				ui.Section("Active Worktrees")
				for _, wt := range worktrees {
					agentStr := "-"
					if wt.AgentID != nil {
						agent, err := db.GetAgent(*wt.AgentID)
						if err == nil {
							agentStr = agent.Name
						}
					}
					taskStr := ""
					if wt.TaskDescription != nil {
						taskStr = *wt.TaskDescription
					}
					fmt.Printf("  %s  branch:%s  agent:%s  task:%s\n",
						ui.T.Muted(wt.ID[:12]),
						ui.T.Muted(wt.Branch),
						ui.T.Muted(agentStr),
						ui.T.Muted(taskStr),
					)
				}
				ui.Blank()
			} else if !ui.IsJSON() {
				fmt.Println("  No active worktrees")
			}

			// Refresh file touches for live conflict scanning
			conflicts.ScanAndUpdate(db, repo)

			// Conflicts
			conflictList, err := db.FindConflicts(repo.ID)
			if err == nil && len(conflictList) > 0 {
				if ui.IsJSON() {
					for _, c := range conflictList {
						repoData.Conflicts = append(repoData.Conflicts, statusConflictJSON{
							File:      c.FilePath,
							Worktrees: len(c.Worktrees),
						})
					}
				} else {
					ui.Section("Conflicts")
					for _, c := range conflictList {
						ui.Warning("%s modified in %d worktrees", c.FilePath, len(c.Worktrees))
					}
					ui.Blank()
				}
			}

			// Tasks
			tasks, err := db.ListTasks(repo.ID, nil)
			if err == nil && len(tasks) > 0 {
				if ui.IsJSON() {
					for _, t := range tasks {
						agentStr := ""
						if t.AssignedAgentID != nil {
							agent, err := db.GetAgent(*t.AssignedAgentID)
							if err == nil {
								agentStr = agent.Name
							}
						}
						repoData.Tasks = append(repoData.Tasks, statusTaskJSON{
							Status:      t.Status,
							Description: t.Description,
							Agent:       agentStr,
						})
					}
				} else {
					ui.Section("Tasks")
					for _, t := range tasks {
						agentStr := "(unclaimed)"
						if t.AssignedAgentID != nil {
							agent, err := db.GetAgent(*t.AssignedAgentID)
							if err == nil {
								agentStr = agent.Name
							}
						}
						fmt.Printf("  [%s] %s %s\n",
							ui.StatusColor(t.Status),
							t.Description,
							ui.T.Muted(agentStr),
						)
					}
					ui.Blank()
				}
			}

			if ui.IsJSON() {
				jsonResult.Repos = append(jsonResult.Repos, repoData)
			}
		}

		if ui.IsJSON() {
			return ui.RenderJSON(jsonResult)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
