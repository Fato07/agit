package cmd

import (
	"fmt"

	toml "github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"

	"github.com/fathindos/agit/internal/config"
	apperrors "github.com/fathindos/agit/internal/errors"
	"github.com/fathindos/agit/internal/ui"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage agit configuration",
	Long:  `View and modify agit configuration stored in ~/.agit/config.toml.`,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("could not load config: %w", err)
		}

		if ui.IsJSON() {
			return ui.RenderJSON(cfg)
		}

		data, err := toml.Marshal(cfg)
		if err != nil {
			return fmt.Errorf("could not format config: %w", err)
		}
		fmt.Print(string(data))
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return config.AllKeys(), cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		key, value := args[0], args[1]

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("could not load config: %w", err)
		}

		if err := cfg.SetByDotKey(key, value); err != nil {
			return apperrors.NewUserErrorf("%v", err)
		}

		if err := cfg.Validate(); err != nil {
			return apperrors.NewUserErrorf("invalid value: %v", err)
		}

		if err := config.EnsureDir(); err != nil {
			return fmt.Errorf("could not create config directory: %w", err)
		}

		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("could not save config: %w", err)
		}

		if ui.IsJSON() {
			return ui.RenderJSON(map[string]string{
				"status": "ok",
				"key":    key,
				"value":  value,
			})
		}

		ui.Success("Set %s = %s", key, value)
		return nil
	},
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Print the configuration file path",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := config.ConfigPath()
		if err != nil {
			return err
		}

		if ui.IsJSON() {
			return ui.RenderJSON(map[string]string{"path": path})
		}

		fmt.Println(path)
		return nil
	},
}

var configResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset configuration to defaults",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.EnsureDir(); err != nil {
			return fmt.Errorf("could not create config directory: %w", err)
		}

		cfg := config.DefaultConfig()
		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("could not save config: %w", err)
		}

		if ui.IsJSON() {
			return ui.RenderJSON(map[string]string{"status": "ok", "message": "reset to defaults"})
		}

		ui.Success("Configuration reset to defaults")
		return nil
	},
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configPathCmd)
	configCmd.AddCommand(configResetCmd)
	rootCmd.AddCommand(configCmd)
}
