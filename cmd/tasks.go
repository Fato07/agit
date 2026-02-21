package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	apperrors "github.com/fathindos/agit/internal/errors"
	"github.com/fathindos/agit/internal/registry"
	"github.com/fathindos/agit/internal/ui"
	"github.com/fathindos/agit/internal/ui/interactive"
)

type taskJSON struct {
	ID          string `json:"id"`
	Priority    string `json:"priority"`
	Status      string `json:"status"`
	Agent       string `json:"agent"`
	Description string `json:"description"`
}

var tasksCmd = &cobra.Command{
	Use:               "tasks <repo>",
	Short:             "Manage tasks for a repository",
	Long:              `Create, list, claim, and complete tasks for a repository.`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeRepoNames,
	RunE: func(cmd *cobra.Command, args []string) error {
		repoName := args[0]
		create, _ := cmd.Flags().GetString("create")
		claim, _ := cmd.Flags().GetString("claim")
		complete, _ := cmd.Flags().GetString("complete")
		fail, _ := cmd.Flags().GetString("fail")
		result, _ := cmd.Flags().GetString("result")
		agent, _ := cmd.Flags().GetString("agent")
		priority, _ := cmd.Flags().GetInt("priority")
		isInteractive, _ := cmd.Flags().GetBool("interactive")

		db, err := registry.Open()
		if err != nil {
			return fmt.Errorf("could not open registry: %w", err)
		}
		defer db.Close()

		repo, err := db.GetRepo(repoName)
		if err != nil {
			return err
		}

		var resultPtr *string
		if result != "" {
			resultPtr = &result
		}

		// Create task
		if create != "" {
			task, err := db.CreateTask(repo.ID, create, priority)
			if err != nil {
				return err
			}
			if ui.IsJSON() {
				return ui.RenderJSON(map[string]string{"status": "ok", "message": "created", "id": task.ID, "description": task.Description})
			}
			ui.Success("Created task: %s - %s", task.ID, task.Description)
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
			if ui.IsJSON() {
				return ui.RenderJSON(map[string]string{"status": "ok", "message": "claimed", "task": claim, "agent": agent})
			}
			ui.Success("Task %s claimed by %s", claim, agent)
			return nil
		}

		// Complete task
		if complete != "" {
			if err := db.CompleteTask(complete, resultPtr); err != nil {
				return err
			}
			if ui.IsJSON() {
				return ui.RenderJSON(map[string]string{"status": "ok", "message": "completed", "task": complete})
			}
			ui.Success("Task %s completed", complete)
			return nil
		}

		// Fail task
		if fail != "" {
			if err := db.FailTask(fail, resultPtr); err != nil {
				return err
			}
			if ui.IsJSON() {
				return ui.RenderJSON(map[string]string{"status": "ok", "message": "failed", "task": fail})
			}
			ui.Success("Task %s marked as failed", fail)
			return nil
		}

		// Interactive mode: select a task and choose an action
		if isInteractive {
			return tasksInteractive(db, repo)
		}

		// List tasks
		tasks, err := db.ListTasks(repo.ID, nil)
		if err != nil {
			return err
		}

		if len(tasks) == 0 {
			if ui.IsJSON() {
				return ui.RenderJSON([]interface{}{})
			}
			fmt.Printf("No tasks for %s. Create one with: agit tasks %s --create \"description\"\n", repoName, repoName)
			return nil
		}

		priorityLabel := func(p int) string {
			switch p {
			case 2:
				return "critical"
			case 1:
				return "high"
			default:
				return "normal"
			}
		}

		if ui.IsJSON() {
			var items []taskJSON
			for _, t := range tasks {
				agentStr := "-"
				if t.AssignedAgentID != nil {
					a, err := db.GetAgent(*t.AssignedAgentID)
					if err == nil {
						agentStr = a.Name
					}
				}
				items = append(items, taskJSON{
					ID:          t.ID,
					Priority:    priorityLabel(t.Priority),
					Status:      t.Status,
					Agent:       agentStr,
					Description: t.Description,
				})
			}
			return ui.RenderJSON(items)
		}

		table := ui.NewTable("ID", "Priority", "Status", "Agent", "Description")

		for _, t := range tasks {
			agentStr := "-"
			if t.AssignedAgentID != nil {
				a, err := db.GetAgent(*t.AssignedAgentID)
				if err == nil {
					agentStr = a.Name
				}
			}
			pLabel := priorityLabel(t.Priority)
			table.Append([]string{
				t.ID,
				ui.PriorityColor(pLabel),
				ui.StatusColor(t.Status),
				agentStr,
				t.Description,
			})
		}

		table.Render()
		return nil
	},
}

func tasksInteractive(db *registry.DB, repo *registry.Repo) error {
	tasks, err := db.ListTasks(repo.ID, nil)
	if err != nil {
		return err
	}
	if len(tasks) == 0 {
		fmt.Printf("No tasks for %s.\n", repo.Name)
		return nil
	}

	var items []interactive.Item
	for _, t := range tasks {
		items = append(items, interactive.Item{
			ID:    t.ID,
			Label: fmt.Sprintf("[%s] %s", t.Status, t.Description),
			Desc:  t.ID,
		})
	}

	selected, err := interactive.Select("Select a task:", items)
	if err != nil {
		return err
	}

	// Choose action
	actions := []interactive.Item{
		{ID: "complete", Label: "Complete", Desc: "Mark task as completed"},
		{ID: "fail", Label: "Fail", Desc: "Mark task as failed"},
	}

	action, err := interactive.Select("Choose action for "+selected.ID+":", actions)
	if err != nil {
		return err
	}

	switch action.ID {
	case "complete":
		if err := db.CompleteTask(selected.ID, nil); err != nil {
			return err
		}
		ui.Success("Task %s completed", selected.ID)
	case "fail":
		if err := db.FailTask(selected.ID, nil); err != nil {
			return err
		}
		ui.Success("Task %s marked as failed", selected.ID)
	}

	return nil
}

func init() {
	tasksCmd.Flags().String("create", "", "Create a new task with this description")
	tasksCmd.Flags().Int("priority", 0, "Task priority (0=normal, 1=high, 2=critical)")
	tasksCmd.Flags().String("claim", "", "Claim a task by ID")
	tasksCmd.Flags().String("complete", "", "Complete a task by ID")
	tasksCmd.Flags().String("fail", "", "Fail a task by ID")
	tasksCmd.Flags().String("result", "", "Result message (used with --complete or --fail)")
	tasksCmd.Flags().StringP("agent", "a", "", "Agent name (required for --claim)")
	rootCmd.AddCommand(tasksCmd)
}
