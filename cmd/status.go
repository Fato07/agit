package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/fathindos/agit/internal/conflicts"
	"github.com/fathindos/agit/internal/registry"
)

var statusCmd = &cobra.Command{
	Use:   "status [repo]",
	Short: "Show active worktrees, agents, and conflicts",
	Long:  `Displays the current state of all registered repos or a specific repo.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := registry.Open()
		if err != nil {
			return fmt.Errorf("could not open registry: %w", err)
		}
		defer db.Close()

		bold := color.New(color.Bold).SprintFunc()
		gray := color.New(color.FgHiBlack).SprintFunc()
		yellow := color.New(color.FgYellow).SprintFunc()

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
			fmt.Println("No repositories registered. Add one with: agit add <path>")
			return nil
		}

		for _, repo := range repos {
			fmt.Printf("%s %s (%s)\n\n", bold("REPO:"), bold(repo.Name), repo.DefaultBranch)

			// Active worktrees
			activeStatus := "active"
			worktrees, err := db.ListWorktrees(repo.ID, &activeStatus)
			if err != nil {
				return err
			}

			if len(worktrees) > 0 {
				fmt.Println("ACTIVE WORKTREES:")
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
						gray(wt.ID[:12]),
						gray(wt.Branch),
						gray(agentStr),
						gray(taskStr),
					)
				}
				fmt.Println()
			} else {
				fmt.Println("  No active worktrees")
			}

			// Refresh file touches for live conflict scanning
			conflicts.ScanAndUpdate(db, repo)

			// Conflicts
			conflictList, err := db.FindConflicts(repo.ID)
			if err == nil && len(conflictList) > 0 {
				fmt.Println("CONFLICTS:")
				for _, c := range conflictList {
					fmt.Printf("  %s %s modified in %d worktrees\n",
						yellow("WARNING:"),
						c.FilePath,
						len(c.Worktrees),
					)
				}
				fmt.Println()
			}

			// Tasks
			tasks, err := db.ListTasks(repo.ID, nil)
			if err == nil && len(tasks) > 0 {
				fmt.Println("TASKS:")
				for _, t := range tasks {
					agentStr := "(unclaimed)"
					if t.AssignedAgentID != nil {
						agent, err := db.GetAgent(*t.AssignedAgentID)
						if err == nil {
							agentStr = agent.Name
						}
					}
					fmt.Printf("  [%s] %s %s\n",
						gray(t.Status),
						t.Description,
						gray(agentStr),
					)
				}
				fmt.Println()
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
