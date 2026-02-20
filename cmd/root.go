package cmd

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/spf13/cobra"

	apperrors "github.com/fathindos/agit/internal/errors"
	"github.com/fathindos/agit/internal/issuelink"
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
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			fmt.Fprintf(os.Stderr, "agit crashed unexpectedly: %v\n\nStack trace:\n%s\n", r, stack)
			if issuelink.Enabled() {
				panicErr := fmt.Errorf("panic: %v", r)
				link := issuelink.Build(issuelink.Context{
					Err:     panicErr,
					Command: os.Args,
					Version: Version,
				})
				fmt.Fprintf(os.Stderr, "\nTo report this bug, open:\n  %s\n", link)
			}
			os.Exit(2)
		}
	}()

	issuelink.AppVersion = Version
	rootCmd.SilenceErrors = true

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		if issuelink.Enabled() && !apperrors.IsUserError(err) {
			link := issuelink.Build(issuelink.Context{
				Err:     err,
				Command: os.Args,
				Version: Version,
			})
			fmt.Fprintf(os.Stderr, "\nTo report this bug, open:\n  %s\n", link)
		}
		os.Exit(1)
	}
}
