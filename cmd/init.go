package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/fathindos/agit/internal/config"
	"github.com/fathindos/agit/internal/registry"
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

		green := color.New(color.FgGreen).SprintFunc()
		gray := color.New(color.FgHiBlack).SprintFunc()

		fmt.Printf("%s agit initialized\n\n", green("✓"))
		fmt.Printf("  Config:   %s\n", gray(configPath))
		fmt.Printf("  Database: %s\n", gray(dbPath))
		fmt.Printf("  Dir:      %s\n\n", gray(dir))
		fmt.Printf("Get started:\n")
		fmt.Printf("  agit add <path>    Register a Git repository\n")
		fmt.Printf("  agit repos         List registered repositories\n")
		fmt.Printf("  agit serve         Start MCP server for agents\n")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
