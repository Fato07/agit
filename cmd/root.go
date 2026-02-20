package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var Version = "0.1.0"

var rootCmd = &cobra.Command{
	Use:   "agit",
	Short: "Infrastructure-aware Git orchestration for AI agents",
	Long: `agit provides AI agents with persistent, queryable awareness of their
Git infrastructure. It manages a local registry of repositories, orchestrates
Git worktrees for agent isolation, detects cross-worktree conflicts, and
coordinates task assignment across multiple agents.

Get started:
  agit init          Initialize agit (~/.agit/)
  agit add <path>    Register a Git repository
  agit repos         List registered repositories
  agit spawn <repo>  Create an isolated worktree for an agent
  agit status        Show active worktrees and conflicts
  agit serve         Start the MCP server for agent integration`,
	Version: Version,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
