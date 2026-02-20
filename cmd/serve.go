package cmd

import (
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"

	"github.com/fathindos/agit/internal/config"
	mcpserver "github.com/fathindos/agit/internal/mcp"
	"github.com/fathindos/agit/internal/registry"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the MCP server for agent integration",
	Long: `Starts the agit MCP server so AI agents can discover repositories,
spawn worktrees, check conflicts, and coordinate tasks.

Configure in your agent's MCP settings:
  {
    "mcpServers": {
      "agit": {
        "command": "agit",
        "args": ["serve"]
      }
    }
  }`,
	RunE: func(cmd *cobra.Command, args []string) error {
		transport, _ := cmd.Flags().GetString("transport")
		port, _ := cmd.Flags().GetInt("port")

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("could not load config: %w", err)
		}

		db, err := registry.Open()
		if err != nil {
			return fmt.Errorf("could not open registry: %w", err)
		}
		defer db.Close()

		s := mcpserver.NewServer(db, cfg)

		switch transport {
		case "stdio":
			if err := server.ServeStdio(s); err != nil {
				return fmt.Errorf("stdio server error: %w", err)
			}
		case "sse":
			addr := fmt.Sprintf("127.0.0.1:%d", port)
			sseServer := server.NewSSEServer(s)
			log.Printf("agit MCP server listening on %s (SSE)\n", addr)
			if err := sseServer.Start(addr); err != nil {
				return fmt.Errorf("SSE server error: %w", err)
			}
		default:
			return fmt.Errorf("unknown transport %q (use stdio or sse)", transport)
		}

		return nil
	},
}

func init() {
	serveCmd.Flags().String("transport", "stdio", "Transport: stdio or sse")
	serveCmd.Flags().Int("port", 3847, "Port for SSE transport")
	rootCmd.AddCommand(serveCmd)
}
