package cmd

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/fathindos/agit/internal/config"
	apperrors "github.com/fathindos/agit/internal/errors"
	"github.com/fathindos/agit/internal/issuelink"
	"github.com/fathindos/agit/internal/ui"
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
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip UI init for serve (MCP uses raw stdout)
		if cmd.Name() == "serve" {
			return nil
		}

		// Load config for UI settings
		cfg, err := config.Load()
		if err != nil {
			// Non-fatal: use defaults if config can't be loaded
			cfg = config.DefaultConfig()
		}

		// Apply --no-color flag or config
		noColor, _ := cmd.Flags().GetBool("no-color")
		switch {
		case noColor:
			color.NoColor = true
		case os.Getenv("NO_COLOR") != "" || os.Getenv("AGIT_NO_COLOR") != "":
			color.NoColor = true
		case cfg.UI.Color == "never":
			color.NoColor = true
		case cfg.UI.Color == "always":
			color.NoColor = false
		default: // "auto" or empty
			if !ui.IsTerminal() {
				color.NoColor = true
			}
		}

		// Re-initialize theme after color state is set
		ui.T = ui.DefaultTheme

		// Apply --quiet flag
		quiet, _ := cmd.Flags().GetBool("quiet")
		ui.Quiet = quiet

		// Apply --output flag (flag overrides config)
		output, _ := cmd.Flags().GetString("output")
		if output == "" {
			output = cfg.UI.OutputFormat
		}
		switch output {
		case "json":
			ui.CurrentFormat = ui.FormatJSON
		default:
			ui.CurrentFormat = ui.FormatText
		}

		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringP("output", "o", "", "Output format: text or json")
	rootCmd.PersistentFlags().Bool("no-color", false, "Disable colored output")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Suppress informational messages")
	rootCmd.PersistentFlags().BoolP("interactive", "i", false, "Enable interactive selection mode")
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
