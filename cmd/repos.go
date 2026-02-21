package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/fathindos/agit/internal/registry"
	"github.com/fathindos/agit/internal/ui"
)

type repoJSON struct {
	Name            string `json:"name"`
	Path            string `json:"path"`
	DefaultBranch   string `json:"default_branch"`
	ActiveWorktrees int    `json:"active_worktrees"`
	PendingTasks    int    `json:"pending_tasks"`
}

var reposCmd = &cobra.Command{
	Use:   "repos",
	Short: "List registered repositories",
	Long:  `Shows all repositories registered with agit and their current state.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := registry.Open()
		if err != nil {
			return fmt.Errorf("could not open registry: %w", err)
		}
		defer db.Close()

		repos, err := db.ListRepos()
		if err != nil {
			return fmt.Errorf("could not list repos: %w", err)
		}

		if len(repos) == 0 {
			if ui.IsJSON() {
				return ui.RenderJSON([]interface{}{})
			}
			fmt.Println("No repositories registered. Add one with: agit add <path>")
			return nil
		}

		if ui.IsJSON() {
			var items []repoJSON
			for _, repo := range repos {
				stats, err := db.GetRepoStats(repo.ID)
				if err != nil {
					stats = &registry.RepoStats{}
				}
				items = append(items, repoJSON{
					Name:            repo.Name,
					Path:            repo.Path,
					DefaultBranch:   repo.DefaultBranch,
					ActiveWorktrees: stats.ActiveWorktrees,
					PendingTasks:    stats.PendingTasks,
				})
			}
			return ui.RenderJSON(items)
		}

		table := ui.NewTable("Name", "Path", "Branch", "Worktrees", "Tasks")

		for _, repo := range repos {
			stats, err := db.GetRepoStats(repo.ID)
			if err != nil {
				stats = &registry.RepoStats{}
			}

			wtStr := fmt.Sprintf("%d active", stats.ActiveWorktrees)
			taskStr := fmt.Sprintf("%d pending", stats.PendingTasks)

			if stats.ActiveWorktrees > 0 {
				wtStr = ui.T.Success(wtStr)
			}
			if stats.PendingTasks > 0 {
				taskStr = ui.T.Info(taskStr)
			}

			table.Append([]string{
				repo.Name,
				repo.Path,
				repo.DefaultBranch,
				wtStr,
				taskStr,
			})
		}

		table.Render()
		return nil
	},
}

var removeCmd = &cobra.Command{
	Use:               "remove <name>",
	Short:             "Unregister a repository",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeRepoNames,
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := registry.Open()
		if err != nil {
			return fmt.Errorf("could not open registry: %w", err)
		}
		defer db.Close()

		if err := db.RemoveRepo(args[0]); err != nil {
			return err
		}

		if ui.IsJSON() {
			return ui.RenderJSON(map[string]string{"status": "ok", "message": "removed", "name": args[0]})
		}

		ui.Success("Removed: %s", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(reposCmd)
	rootCmd.AddCommand(removeCmd)
}
