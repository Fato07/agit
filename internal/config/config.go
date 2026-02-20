package config

import (
	"fmt"
	"os"
	"path/filepath"

	toml "github.com/pelletier/go-toml/v2"
)

// Config represents the agit configuration file (~/.agit/config.toml)
type Config struct {
	Server   ServerConfig   `toml:"server"`
	Defaults DefaultsConfig `toml:"defaults"`
	Agent    AgentConfig    `toml:"agent"`
}

type ServerConfig struct {
	Transport string `toml:"transport"` // "stdio" or "sse"
	Port      int    `toml:"port"`
}

type DefaultsConfig struct {
	BranchPrefix      string `toml:"branch_prefix"`
	WorktreeDir       string `toml:"worktree_dir"`
	CleanupStaleAfter string `toml:"cleanup_stale_after"`
	AutoConflictCheck bool   `toml:"auto_conflict_check"`
}

type AgentConfig struct {
	HeartbeatInterval string `toml:"heartbeat_interval"`
	StaleAfter        string `toml:"stale_after"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Transport: "stdio",
			Port:      3847,
		},
		Defaults: DefaultsConfig{
			BranchPrefix:      "agit/",
			WorktreeDir:       ".worktrees",
			CleanupStaleAfter: "24h",
			AutoConflictCheck: true,
		},
		Agent: AgentConfig{
			HeartbeatInterval: "30s",
			StaleAfter:        "5m",
		},
	}
}

// AgitDir returns the path to the agit config directory (~/.agit/)
func AgitDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}
	return filepath.Join(home, ".agit"), nil
}

// ConfigPath returns the path to the config file
func ConfigPath() (string, error) {
	dir, err := AgitDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.toml"), nil
}

// DBPath returns the path to the SQLite database
func DBPath() (string, error) {
	dir, err := AgitDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "agit.db"), nil
}

// Load reads the config from disk. Returns default config if file doesn't exist.
func Load() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("could not read config: %w", err)
	}

	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("could not parse config: %w", err)
	}

	return cfg, nil
}

// Save writes the config to disk
func Save(cfg *Config) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}

	data, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("could not marshal config: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// EnsureDir creates the agit directory if it doesn't exist
func EnsureDir() error {
	dir, err := AgitDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(dir, 0755)
}
