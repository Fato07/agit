package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/fathindos/agit/internal/config"
	gitops "github.com/fathindos/agit/internal/git"
	"github.com/fathindos/agit/internal/registry"
)

var spawnCmd = &cobra.Command{
	Use:   "spawn <repo>",
	Short: "Create an isolated worktree for an agent",
	Long: `Creates a new Git worktree branched from the repo's default branch.
The worktree provides an isolated workspace where an agent can make changes
without affecting other agents or the main branch.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoName := args[0]
		task, _ := cmd.Flags().GetString("task")
		branch, _ := cmd.Flags().GetString("branch")
		agentName, _ := cmd.Flags().GetString("agent")

		// Load config
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("could not load config: %w", err)
		}

		// Open registry
		db, err := registry.Open()
		if err != nil {
			return fmt.Errorf("could not open registry: %w", err)
		}
		defer db.Close()

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
				// Keep only alphanumeric and hyphens
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
				// Auto-register agent
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
			// Clean up the git worktree if registry insert fails
			gitops.RemoveWorktree(repo.Path, worktreePath)
			return fmt.Errorf("could not record worktree: %w", err)
		}

		// Update agent's current worktree
		if agentID != nil {
			db.UpdateAgentWorktree(*agentID, &wt.ID)
		}

		green := color.New(color.FgGreen).SprintFunc()
		gray := color.New(color.FgHiBlack).SprintFunc()

		fmt.Printf("%s Created worktree: %s\n", green("✓"), gray(worktreePath))
		fmt.Printf("  Branch: %s\n", gray(branch))
		if agentName != "" {
			fmt.Printf("  Agent:  %s\n", gray(agentName))
		}
		if task != "" {
			fmt.Printf("  Task:   %s\n", gray(task))
		}
		fmt.Printf("\nAgent can work in: %s\n", color.New(color.Bold).Sprint(worktreePath))

		return nil
	},
}

func init() {
	spawnCmd.Flags().StringP("task", "t", "", "Description of what the agent will do")
	spawnCmd.Flags().StringP("branch", "b", "", "Custom branch name (auto-generated if omitted)")
	spawnCmd.Flags().StringP("agent", "a", "", "Agent name to assign this worktree to")
	rootCmd.AddCommand(spawnCmd)
}
