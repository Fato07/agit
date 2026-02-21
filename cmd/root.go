package cmd

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/fathindos/agit/internal/config"
	apperrors "github.com/fathindos/agit/internal/errors"
	"github.com/fathindos/agit/internal/issuelink"
	"github.com/fathindos/agit/internal/ui"
	"github.com/fathindos/agit/internal/update"
)

var Version = "0.2.0"

// updateNotice carries a message from the background update checker.
var updateNotice chan string

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

		// Launch background update check (skip for update, serve, completion commands)
		name := cmd.Name()
		if !quiet && !ui.IsJSON() && name != "update" && name != "upgrade" && !strings.HasPrefix(name, "__") {
			startBackgroundUpdateCheck(cfg)
		}

		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, _ []string) error {
		if updateNotice == nil {
			return nil
		}
		// Non-blocking read: if the check hasn't finished, skip silently
		select {
		case msg := <-updateNotice:
			if msg != "" {
				fmt.Fprintln(os.Stderr, msg)
			}
		default:
		}
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringP("output", "o", "", "Output format: text or json")
	rootCmd.PersistentFlags().Bool("no-color", false, "Disable colored output")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Suppress informational messages")
	rootCmd.PersistentFlags().BoolP("interactive", "i", false, "Enable interactive selection mode")

	// Branded version output
	rootCmd.SetVersionTemplate(ui.T.Brand(ui.Sym.Zap+" agit v"+Version) + "\n")
}

// startBackgroundUpdateCheck launches a goroutine that checks for new versions.
func startBackgroundUpdateCheck(cfg *config.Config) {
	if !cfg.Updates.Enabled {
		return
	}
	interval, err := time.ParseDuration(cfg.Updates.CheckInterval)
	if err != nil || interval <= 0 {
		return
	}

	updateNotice = make(chan string, 1)

	go func() {
		cache, err := update.LoadCache()
		if err != nil {
			return
		}

		if !cache.ShouldCheck(interval) {
			// Use cached result
			if cache.LatestVersion != "" && update.IsNewer(Version, cache.LatestVersion) {
				updateNotice <- fmt.Sprintf(
					"\nA new version of agit is available: %s (current: v%s). Run \"agit update\" to upgrade.",
					cache.LatestVersion, Version,
				)
			}
			return
		}

		release, err := update.FetchLatestRelease()
		if err != nil {
			// Network failure — don't cache the error, just skip silently
			return
		}

		// Update cache
		cache.LastCheck = time.Now()
		cache.LatestVersion = release.TagName
		cache.LatestURL = release.HTMLURL
		_ = update.SaveCache(cache) // best-effort

		if update.IsNewer(Version, release.TagName) {
			updateNotice <- fmt.Sprintf(
				"\nA new version of agit is available: %s (current: v%s). Run \"agit update\" to upgrade.",
				release.TagName, Version,
			)
		}
	}()
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
