package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/fathindos/agit/internal/git"
	"github.com/fathindos/agit/internal/registry"
)

var addCmd = &cobra.Command{
	Use:   "add <path>",
	Short: "Register a Git repository with agit",
	Long:  `Registers a Git repository so agents can discover and work with it.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoPath, err := filepath.Abs(args[0])
		if err != nil {
			return fmt.Errorf("could not resolve path: %w", err)
		}

		// Validate it's a git repo
		if !git.IsGitRepo(repoPath) {
			return fmt.Errorf("%s is not a Git repository", repoPath)
		}

		// Get name (flag or directory name)
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			name = filepath.Base(repoPath)
		}

		// Auto-detect remote and default branch
		remoteURL, err := git.GetRemoteURL(repoPath)
		if err != nil {
			remoteURL = "" // local-only repo, no remote
		}

		defaultBranch, err := git.GetDefaultBranch(repoPath)
		if err != nil {
			defaultBranch = "main"
		}

		// Open registry and add
		db, err := registry.Open()
		if err != nil {
			return fmt.Errorf("could not open registry: %w", err)
		}
		defer db.Close()

		repo, err := db.AddRepo(name, repoPath, remoteURL, defaultBranch)
		if err != nil {
			return fmt.Errorf("could not register repo: %w", err)
		}

		green := color.New(color.FgGreen).SprintFunc()
		gray := color.New(color.FgHiBlack).SprintFunc()

		fmt.Printf("%s Registered: %s\n", green("✓"), color.New(color.Bold).Sprint(repo.Name))
		fmt.Printf("  Path:   %s\n", gray(repo.Path))
		if repo.RemoteURL != "" {
			fmt.Printf("  Remote: %s\n", gray(repo.RemoteURL))
		}
		fmt.Printf("  Branch: %s\n", gray(repo.DefaultBranch))

		return nil
	},
}

func init() {
	addCmd.Flags().StringP("name", "n", "", "Alias for the repository (defaults to directory name)")
	rootCmd.AddCommand(addCmd)
}
