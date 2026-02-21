package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/fathindos/agit/internal/config"
	"github.com/fathindos/agit/internal/registry"
	"github.com/fathindos/agit/internal/ui"
)

type agentJSON struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Status   string `json:"status"`
	LastSeen string `json:"last_seen"`
	Worktree string `json:"worktree"`
}

var agentsCmd = &cobra.Command{
	Use:   "agents",
	Short: "List and manage registered agents",
	Long:  `List all registered agents, sweep stale agents, or remove an agent by name.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		sweep, _ := cmd.Flags().GetBool("sweep")
		remove, _ := cmd.Flags().GetString("remove")

		db, err := registry.Open()
		if err != nil {
			return fmt.Errorf("could not open registry: %w", err)
		}
		defer db.Close()

		// Sweep stale agents
		if sweep {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("could not load config: %w", err)
			}
			staleAfter, err := time.ParseDuration(cfg.Agent.StaleAfter)
			if err != nil {
				return fmt.Errorf("invalid stale_after duration %q: %w", cfg.Agent.StaleAfter, err)
			}
			count, err := db.SweepStaleAgents(staleAfter)
			if err != nil {
				return err
			}
			if ui.IsJSON() {
				return ui.RenderJSON(map[string]interface{}{"status": "ok", "message": "swept", "count": count})
			}
			ui.Success("Swept %d stale agent(s)", count)
			return nil
		}

		// Remove agent
		if remove != "" {
			if err := db.RemoveAgent(remove); err != nil {
				return err
			}
			if ui.IsJSON() {
				return ui.RenderJSON(map[string]string{"status": "ok", "message": "removed", "name": remove})
			}
			ui.Success("Removed agent %q", remove)
			return nil
		}

		// List agents
		agents, err := db.ListAgents()
		if err != nil {
			return err
		}

		if len(agents) == 0 {
			if ui.IsJSON() {
				return ui.RenderJSON([]interface{}{})
			}
			fmt.Println("No agents registered. Agents are created when spawning worktrees with --agent.")
			return nil
		}

		if ui.IsJSON() {
			var items []agentJSON
			for _, a := range agents {
				wtStr := "-"
				if a.CurrentWorktreeID != nil {
					wt, err := db.GetWorktree(*a.CurrentWorktreeID)
					if err == nil {
						wtStr = wt.Branch
					}
				}
				idShort := a.ID
				if len(idShort) > 12 {
					idShort = idShort[:12]
				}
				items = append(items, agentJSON{
					ID:       idShort,
					Name:     a.Name,
					Type:     a.Type,
					Status:   a.Status,
					LastSeen: a.LastSeen.Format("2006-01-02 15:04"),
					Worktree: wtStr,
				})
			}
			return ui.RenderJSON(items)
		}

		table := ui.NewTable("ID", "Name", "Type", "Status", "Last Seen", "Worktree")

		for _, a := range agents {
			wtStr := "-"
			if a.CurrentWorktreeID != nil {
				wt, err := db.GetWorktree(*a.CurrentWorktreeID)
				if err == nil {
					wtStr = wt.Branch
				}
			}
			idShort := a.ID
			if len(idShort) > 12 {
				idShort = idShort[:12]
			}
			table.Append([]string{
				idShort,
				a.Name,
				a.Type,
				ui.StatusColor(a.Status),
				a.LastSeen.Format("2006-01-02 15:04"),
				wtStr,
			})
		}

		table.Render()
		return nil
	},
}

func init() {
	agentsCmd.Flags().Bool("sweep", false, "Mark stale agents as disconnected")
	agentsCmd.Flags().String("remove", "", "Remove an agent by name")
	rootCmd.AddCommand(agentsCmd)
}
