package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	apperrors "github.com/fathindos/agit/internal/errors"
	"github.com/fathindos/agit/internal/registry"
)

var tasksCmd = &cobra.Command{
	Use:   "tasks <repo>",
	Short: "Manage tasks for a repository",
	Long:  `Create, list, claim, and complete tasks for a repository.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoName := args[0]
		create, _ := cmd.Flags().GetString("create")
		claim, _ := cmd.Flags().GetString("claim")
		complete, _ := cmd.Flags().GetString("complete")
		fail, _ := cmd.Flags().GetString("fail")
		result, _ := cmd.Flags().GetString("result")
		agent, _ := cmd.Flags().GetString("agent")

		db, err := registry.Open()
		if err != nil {
			return fmt.Errorf("could not open registry: %w", err)
		}
		defer db.Close()

		repo, err := db.GetRepo(repoName)
		if err != nil {
			return err
		}

		green := color.New(color.FgGreen).SprintFunc()

		var resultPtr *string
		if result != "" {
			resultPtr = &result
		}

		// Create task
		if create != "" {
			task, err := db.CreateTask(repo.ID, create)
			if err != nil {
				return err
			}
			fmt.Printf("%s Created task: %s - %s\n", green("✓"), task.ID, task.Description)
			return nil
		}

		// Claim task
		if claim != "" {
			if agent == "" {
				return apperrors.NewUserError("--agent is required when claiming a task")
			}
			agentObj, err := db.GetAgentByName(agent)
			if err != nil {
				return err
			}
			if agentObj == nil {
				agentObj, err = db.RegisterAgent(agent, "custom")
				if err != nil {
					return err
				}
			}
			if err := db.ClaimTask(claim, agentObj.ID); err != nil {
				return err
			}
			fmt.Printf("%s Task %s claimed by %s\n", green("✓"), claim, agent)
			return nil
		}

		// Complete task
		if complete != "" {
			if err := db.CompleteTask(complete, resultPtr); err != nil {
				return err
			}
			fmt.Printf("%s Task %s completed\n", green("✓"), complete)
			return nil
		}

		// Fail task
		if fail != "" {
			if err := db.FailTask(fail, resultPtr); err != nil {
				return err
			}
			fmt.Printf("%s Task %s marked as failed\n", green("✓"), fail)
			return nil
		}

		// List tasks
		tasks, err := db.ListTasks(repo.ID, nil)
		if err != nil {
			return err
		}

		if len(tasks) == 0 {
			fmt.Printf("No tasks for %s. Create one with: agit tasks %s --create \"description\"\n", repoName, repoName)
			return nil
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "Status", "Agent", "Description"})
		table.SetBorder(false)
		table.SetColumnSeparator("")
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetAlignment(tablewriter.ALIGN_LEFT)

		for _, t := range tasks {
			agentStr := "-"
			if t.AssignedAgentID != nil {
				a, err := db.GetAgent(*t.AssignedAgentID)
				if err == nil {
					agentStr = a.Name
				}
			}
			table.Append([]string{t.ID, t.Status, agentStr, t.Description})
		}

		table.Render()
		return nil
	},
}

func init() {
	tasksCmd.Flags().String("create", "", "Create a new task with this description")
	tasksCmd.Flags().String("claim", "", "Claim a task by ID")
	tasksCmd.Flags().String("complete", "", "Complete a task by ID")
	tasksCmd.Flags().String("fail", "", "Fail a task by ID")
	tasksCmd.Flags().String("result", "", "Result message (used with --complete or --fail)")
	tasksCmd.Flags().StringP("agent", "a", "", "Agent name (required for --claim)")
	rootCmd.AddCommand(tasksCmd)
}
