package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/fathindos/agit/internal/ui"
	"github.com/fathindos/agit/internal/ui/interactive"
	"github.com/fathindos/agit/internal/update"
)

var updateCmd = &cobra.Command{
	Use:     "update",
	Aliases: []string{"upgrade"},
	Short:   "Update agit to the latest version",
	Long: `Checks for the latest release on GitHub and updates the agit binary.

For Go-installed binaries, runs "go install" to fetch the latest version.
For standalone binaries, downloads the correct platform archive, verifies
the SHA256 checksum, and replaces the current binary.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if ui.IsJSON() {
			return updateJSON()
		}
		return updateText(cmd)
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

type updateResultJSON struct {
	Status  string `json:"status"`
	From    string `json:"from"`
	To      string `json:"to,omitempty"`
	Message string `json:"message,omitempty"`
}

func updateJSON() error {
	release, err := update.FetchLatestRelease()
	if err != nil {
		return ui.RenderJSON(updateResultJSON{
			Status:  "error",
			From:    Version,
			Message: err.Error(),
		})
	}

	if !update.IsNewer(Version, release.TagName) {
		return ui.RenderJSON(updateResultJSON{
			Status:  "current",
			From:    Version,
			Message: "already up to date",
		})
	}

	if err := update.SelfUpdate(release); err != nil {
		return ui.RenderJSON(updateResultJSON{
			Status:  "error",
			From:    Version,
			To:      release.TagName,
			Message: err.Error(),
		})
	}

	return ui.RenderJSON(updateResultJSON{
		Status: "ok",
		From:   Version,
		To:     release.TagName,
	})
}

func updateText(cmd *cobra.Command) error {
	ui.Info("Checking for updates...")

	var release *update.ReleaseInfo
	err := interactive.WithSpinner("Fetching latest release", func() error {
		var fetchErr error
		release, fetchErr = update.FetchLatestRelease()
		return fetchErr
	})
	if err != nil {
		return fmt.Errorf("could not check for updates: %w", err)
	}

	if !update.IsNewer(Version, release.TagName) {
		ui.Success("agit %s is already the latest version", Version)
		return nil
	}

	ui.Info("New version available: %s (current: %s)", release.TagName, Version)

	err = interactive.WithSpinner("Downloading and installing "+release.TagName, func() error {
		return update.SelfUpdate(release)
	})
	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	ui.Success("Updated agit from %s to %s", Version, release.TagName)

	// Update cache after successful update
	cache := &update.CheckCache{LatestVersion: release.TagName, LatestURL: release.HTMLURL}
	_ = update.SaveCache(cache) // best-effort

	return nil
}
