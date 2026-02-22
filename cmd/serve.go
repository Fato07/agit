package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"

	"github.com/fathindos/agit/internal/config"
	apperrors "github.com/fathindos/agit/internal/errors"
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

		if cmd.Flags().Changed("port") && !cmd.Flags().Changed("transport") {
			transport = "sse"
		}
		if cmd.Flags().Changed("port") && cmd.Flags().Changed("transport") && transport == "stdio" {
			fmt.Fprintln(os.Stderr, "Warning: --port is only used with --transport=sse, ignoring")
		}

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

			// Handle graceful shutdown on SIGTERM/SIGINT
			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

			errCh := make(chan error, 1)
			go func() {
				errCh <- sseServer.Start(addr)
			}()

			select {
			case err := <-errCh:
				if err != nil {
					if strings.Contains(err.Error(), "address already in use") {
						return apperrors.NewUserErrorf("port %d is already in use — try a different port with --port", port)
					}
					return fmt.Errorf("SSE server error: %w", err)
				}
			case sig := <-sigCh:
				log.Printf("received %s, shutting down gracefully...", sig)
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := sseServer.Shutdown(ctx); err != nil {
					log.Printf("shutdown error: %v", err)
				}
				log.Println("server stopped")
			}
		default:
			return apperrors.NewUserErrorf("unknown transport %q (use stdio or sse)", transport)
		}

		return nil
	},
}

func init() {
	serveCmd.Flags().String("transport", "stdio", "Transport: stdio or sse")
	serveCmd.Flags().Int("port", 3847, "Port for SSE transport")
	rootCmd.AddCommand(serveCmd)
}
