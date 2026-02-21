package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	apperrors "github.com/fathindos/agit/internal/errors"
	"github.com/fathindos/agit/internal/git"
	"github.com/fathindos/agit/internal/registry"
	"github.com/fathindos/agit/internal/ui"
)

var addCmd = &cobra.Command{
	Use:   "add <path>",
	Short: "Register a Git repository with agit",
	Long:  `Registers a Git repository so agents can discover and work with it.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoPath, err := filepath.Abs(args[0])
		if err != nil {
			return apperrors.NewUserErrorf("could not resolve path: %v", err)
		}

		// Validate it's a git repo
		if !git.IsGitRepo(repoPath) {
			return apperrors.NewUserErrorf("%s is not a Git repository", repoPath)
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

		if ui.IsJSON() {
			return ui.RenderJSON(map[string]string{
				"status":         "ok",
				"message":        "registered",
				"name":           repo.Name,
				"path":           repo.Path,
				"remote_url":     repo.RemoteURL,
				"default_branch": repo.DefaultBranch,
			})
		}

		ui.Success("Registered: %s", ui.T.Bold(repo.Name))
		ui.KeyValue("Path", repo.Path)
		if repo.RemoteURL != "" {
			ui.KeyValue("Remote", repo.RemoteURL)
		}
		ui.KeyValue("Branch", repo.DefaultBranch)

		return nil
	},
}

func init() {
	addCmd.Flags().StringP("name", "n", "", "Alias for the repository (defaults to directory name)")
	rootCmd.AddCommand(addCmd)
}
