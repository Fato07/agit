package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	toml "github.com/pelletier/go-toml/v2"
)

// Config represents the agit configuration file (~/.agit/config.toml)
type Config struct {
	Server   ServerConfig   `toml:"server"`
	Defaults DefaultsConfig `toml:"defaults"`
	Agent    AgentConfig    `toml:"agent"`
	UI       UIConfig       `toml:"ui"`
	Updates  UpdatesConfig  `toml:"updates"`
}

// UpdatesConfig controls automatic update checking.
type UpdatesConfig struct {
	Enabled       bool   `toml:"enabled"`        // default: true
	CheckInterval string `toml:"check_interval"` // default: "24h", "0" to disable
}

// UIConfig controls CLI display behavior.
type UIConfig struct {
	Color        string `toml:"color"`         // "auto", "always", "never"
	OutputFormat string `toml:"output_format"` // "text", "json"
	Compact      bool   `toml:"compact"`
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
		Updates: UpdatesConfig{
			Enabled:       true,
			CheckInterval: "24h",
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

// Validate checks that all configuration values are within valid ranges.
func (c *Config) Validate() error {
	// Server
	switch c.Server.Transport {
	case "stdio", "sse":
	default:
		return fmt.Errorf("invalid server.transport %q: must be stdio or sse", c.Server.Transport)
	}
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server.port %d: must be 1-65535", c.Server.Port)
	}

	// Durations
	for _, pair := range []struct {
		key, val string
	}{
		{"defaults.cleanup_stale_after", c.Defaults.CleanupStaleAfter},
		{"agent.heartbeat_interval", c.Agent.HeartbeatInterval},
		{"agent.stale_after", c.Agent.StaleAfter},
		{"updates.check_interval", c.Updates.CheckInterval},
	} {
		if _, err := time.ParseDuration(pair.val); err != nil {
			return fmt.Errorf("invalid %s %q: %w", pair.key, pair.val, err)
		}
	}

	// UI
	switch c.UI.Color {
	case "", "auto", "always", "never":
	default:
		return fmt.Errorf("invalid ui.color %q: must be auto, always, or never", c.UI.Color)
	}
	switch c.UI.OutputFormat {
	case "", "text", "json":
	default:
		return fmt.Errorf("invalid ui.output_format %q: must be text or json", c.UI.OutputFormat)
	}

	return nil
}

// AllKeys returns the list of all valid dot-notation configuration keys.
func AllKeys() []string {
	return []string{
		"server.transport",
		"server.port",
		"defaults.branch_prefix",
		"defaults.worktree_dir",
		"defaults.cleanup_stale_after",
		"defaults.auto_conflict_check",
		"agent.heartbeat_interval",
		"agent.stale_after",
		"ui.color",
		"ui.output_format",
		"ui.compact",
		"updates.enabled",
		"updates.check_interval",
	}
}

// SetByDotKey sets a configuration value by dot-notation key.
func (c *Config) SetByDotKey(key, value string) error {
	switch key {
	case "server.transport":
		c.Server.Transport = value
	case "server.port":
		v, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid value for server.port: %w", err)
		}
		c.Server.Port = v
	case "defaults.branch_prefix":
		c.Defaults.BranchPrefix = value
	case "defaults.worktree_dir":
		c.Defaults.WorktreeDir = value
	case "defaults.cleanup_stale_after":
		c.Defaults.CleanupStaleAfter = value
	case "defaults.auto_conflict_check":
		v, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid value for defaults.auto_conflict_check: %w", err)
		}
		c.Defaults.AutoConflictCheck = v
	case "agent.heartbeat_interval":
		c.Agent.HeartbeatInterval = value
	case "agent.stale_after":
		c.Agent.StaleAfter = value
	case "ui.color":
		c.UI.Color = value
	case "ui.output_format":
		c.UI.OutputFormat = value
	case "ui.compact":
		v, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid value for ui.compact: %w", err)
		}
		c.UI.Compact = v
	case "updates.enabled":
		v, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid value for updates.enabled: %w", err)
		}
		c.Updates.Enabled = v
	case "updates.check_interval":
		c.Updates.CheckInterval = value
	default:
		return fmt.Errorf("unknown config key %q", key)
	}
	return nil
}

// GetByDotKey returns the current value for a dot-notation key as a string.
func (c *Config) GetByDotKey(key string) (string, error) {
	switch key {
	case "server.transport":
		return c.Server.Transport, nil
	case "server.port":
		return strconv.Itoa(c.Server.Port), nil
	case "defaults.branch_prefix":
		return c.Defaults.BranchPrefix, nil
	case "defaults.worktree_dir":
		return c.Defaults.WorktreeDir, nil
	case "defaults.cleanup_stale_after":
		return c.Defaults.CleanupStaleAfter, nil
	case "defaults.auto_conflict_check":
		return strconv.FormatBool(c.Defaults.AutoConflictCheck), nil
	case "agent.heartbeat_interval":
		return c.Agent.HeartbeatInterval, nil
	case "agent.stale_after":
		return c.Agent.StaleAfter, nil
	case "ui.color":
		return c.UI.Color, nil
	case "ui.output_format":
		return c.UI.OutputFormat, nil
	case "ui.compact":
		return strconv.FormatBool(c.UI.Compact), nil
	case "updates.enabled":
		return strconv.FormatBool(c.Updates.Enabled), nil
	case "updates.check_interval":
		return c.Updates.CheckInterval, nil
	default:
		return "", fmt.Errorf("unknown config key %q", key)
	}
}

// EnsureDir creates the agit directory if it doesn't exist
func EnsureDir() error {
	dir, err := AgitDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(dir, 0755)
}
