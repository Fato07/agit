package cmd

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/fathindos/agit/internal/registry"
)

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
			fmt.Println("No repositories registered. Add one with: agit add <path>")
			return nil
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Name", "Path", "Branch", "Worktrees", "Tasks"})
		table.SetBorder(false)
		table.SetColumnSeparator("")
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetAlignment(tablewriter.ALIGN_LEFT)

		for _, repo := range repos {
			stats, err := db.GetRepoStats(repo.ID)
			if err != nil {
				stats = &registry.RepoStats{}
			}

			wtStr := fmt.Sprintf("%d active", stats.ActiveWorktrees)
			taskStr := fmt.Sprintf("%d pending", stats.PendingTasks)

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
	Use:   "remove <name>",
	Short: "Unregister a repository",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := registry.Open()
		if err != nil {
			return fmt.Errorf("could not open registry: %w", err)
		}
		defer db.Close()

		if err := db.RemoveRepo(args[0]); err != nil {
			return err
		}
		fmt.Printf("Removed: %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(reposCmd)
	rootCmd.AddCommand(removeCmd)
}
