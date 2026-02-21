package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/fathindos/agit/internal/config"
	"github.com/fathindos/agit/internal/registry"
	"github.com/fathindos/agit/internal/ui"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize agit",
	Long:  `Creates the ~/.agit/ directory, default config, and SQLite database.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create ~/.agit/ directory
		if err := config.EnsureDir(); err != nil {
			return fmt.Errorf("could not create agit directory: %w", err)
		}

		dir, _ := config.AgitDir()
		configPath, _ := config.ConfigPath()
		dbPath, _ := config.DBPath()

		// Write default config
		cfg := config.DefaultConfig()
		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("could not write config: %w", err)
		}

		// Initialize database
		db, err := registry.Open()
		if err != nil {
			return fmt.Errorf("could not initialize database: %w", err)
		}
		db.Close()

		if ui.IsJSON() {
			return ui.RenderJSON(map[string]string{
				"status":   "ok",
				"message":  "agit initialized",
				"config":   configPath,
				"database": dbPath,
				"dir":      dir,
			})
		}

		ui.Success("agit initialized")
		ui.Blank()
		ui.KeyValue("Config", configPath)
		ui.KeyValue("Database", dbPath)
		ui.KeyValue("Dir", dir)
		ui.Blank()
		fmt.Println("Get started:")
		fmt.Println("  agit add <path>    Register a Git repository")
		fmt.Println("  agit repos         List registered repositories")
		fmt.Println("  agit serve         Start MCP server for agents")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
