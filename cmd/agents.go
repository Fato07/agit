package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/fathindos/agit/internal/config"
	"github.com/fathindos/agit/internal/registry"
)

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

		green := color.New(color.FgGreen).SprintFunc()

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
			fmt.Printf("%s Swept %d stale agent(s)\n", green("✓"), count)
			return nil
		}

		// Remove agent
		if remove != "" {
			if err := db.RemoveAgent(remove); err != nil {
				return err
			}
			fmt.Printf("%s Removed agent %q\n", green("✓"), remove)
			return nil
		}

		// List agents
		agents, err := db.ListAgents()
		if err != nil {
			return err
		}

		if len(agents) == 0 {
			fmt.Println("No agents registered. Agents are created when spawning worktrees with --agent.")
			return nil
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "Name", "Type", "Status", "Last Seen", "Worktree"})
		table.SetBorder(false)
		table.SetColumnSeparator("")
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetAlignment(tablewriter.ALIGN_LEFT)

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
				a.Status,
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
