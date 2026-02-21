package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/fathindos/agit/internal/config"
	apperrors "github.com/fathindos/agit/internal/errors"
	gitops "github.com/fathindos/agit/internal/git"
	"github.com/fathindos/agit/internal/registry"
	"github.com/fathindos/agit/internal/ui"
	"github.com/fathindos/agit/internal/ui/interactive"
)

var spawnCmd = &cobra.Command{
	Use:   "spawn [repo]",
	Short: "Create an isolated worktree for an agent",
	Long: `Creates a new Git worktree branched from the repo's default branch.
The worktree provides an isolated workspace where an agent can make changes
without affecting other agents or the main branch.

With -i (interactive), presents a selector if no repo is specified.`,
	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: completeRepoNames,
	RunE: func(cmd *cobra.Command, args []string) error {
		isInteractive, _ := cmd.Flags().GetBool("interactive")
		task, _ := cmd.Flags().GetString("task")
		branch, _ := cmd.Flags().GetString("branch")
		agentName, _ := cmd.Flags().GetString("agent")

		// Open registry
		db, err := registry.Open()
		if err != nil {
			return fmt.Errorf("could not open registry: %w", err)
		}
		defer db.Close()

		var repoName string
		if len(args) > 0 {
			repoName = args[0]
		} else if isInteractive {
			repos, err := db.ListRepos()
			if err != nil {
				return err
			}
			if len(repos) == 0 {
				fmt.Println("No repositories registered. Add one with: agit add <path>")
				return nil
			}
			var items []interactive.Item
			for _, r := range repos {
				items = append(items, interactive.Item{
					ID:    r.Name,
					Label: r.Name,
					Desc:  r.Path,
				})
			}
			selected, err := interactive.Select("Select a repository:", items)
			if err != nil {
				return err
			}
			repoName = selected.ID
		} else {
			return apperrors.NewUserError("requires a repo argument (or use -i for interactive mode)")
		}

		// Load config
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("could not load config: %w", err)
		}

		// Get repo
		repo, err := db.GetRepo(repoName)
		if err != nil {
			return err
		}

		// Generate worktree ID and branch name
		shortID := uuid.New().String()[:8]
		if branch == "" {
			if task != "" {
				// Slugify the task description
				slug := strings.ToLower(task)
				slug = strings.ReplaceAll(slug, " ", "-")
				cleaned := ""
				for _, c := range slug {
					if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' {
						cleaned += string(c)
					}
				}
				if len(cleaned) > 40 {
					cleaned = cleaned[:40]
				}
				branch = cfg.Defaults.BranchPrefix + cleaned + "-" + shortID
			} else {
				branch = cfg.Defaults.BranchPrefix + shortID
			}
		}

		// Worktree path
		worktreePath := filepath.Join(repo.Path, cfg.Defaults.WorktreeDir, "agit-"+shortID)

		// Create the git worktree
		if err := gitops.CreateWorktree(repo.Path, worktreePath, branch, repo.DefaultBranch); err != nil {
			return fmt.Errorf("could not create worktree: %w", err)
		}

		// Resolve agent
		var agentID *string
		if agentName != "" {
			agent, err := db.GetAgentByName(agentName)
			if err != nil {
				return err
			}
			if agent == nil {
				agent, err = db.RegisterAgent(agentName, "custom")
				if err != nil {
					return err
				}
			}
			agentID = &agent.ID
		}

		var taskDesc *string
		if task != "" {
			taskDesc = &task
		}

		// Record in registry
		wt, err := db.CreateWorktree(repo.ID, worktreePath, branch, agentID, taskDesc)
		if err != nil {
			gitops.RemoveWorktree(repo.Path, worktreePath)
			return fmt.Errorf("could not record worktree: %w", err)
		}

		if agentID != nil {
			db.UpdateAgentWorktree(*agentID, &wt.ID)
		}

		if ui.IsJSON() {
			result := map[string]string{
				"status":   "ok",
				"message":  "created",
				"path":     worktreePath,
				"branch":   branch,
				"worktree": wt.ID,
			}
			if agentName != "" {
				result["agent"] = agentName
			}
			if task != "" {
				result["task"] = task
			}
			return ui.RenderJSON(result)
		}

		ui.Success("Created worktree: %s", ui.T.Muted(worktreePath))
		ui.KeyValue("Branch", branch)
		if agentName != "" {
			ui.KeyValue("Agent", agentName)
		}
		if task != "" {
			ui.KeyValue("Task", task)
		}
		fmt.Printf("\nAgent can work in: %s\n", ui.T.Bold(worktreePath))

		return nil
	},
}

func init() {
	spawnCmd.Flags().StringP("task", "t", "", "Description of what the agent will do")
	spawnCmd.Flags().StringP("branch", "b", "", "Custom branch name (auto-generated if omitted)")
	spawnCmd.Flags().StringP("agent", "a", "", "Agent name to assign this worktree to")
	_ = spawnCmd.RegisterFlagCompletionFunc("agent", completeAgentNames)
	rootCmd.AddCommand(spawnCmd)
}
