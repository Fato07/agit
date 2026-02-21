package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Server.Transport != "stdio" {
		t.Errorf("expected transport stdio, got %s", cfg.Server.Transport)
	}
	if cfg.Server.Port != 3847 {
		t.Errorf("expected port 3847, got %d", cfg.Server.Port)
	}
	if cfg.Defaults.BranchPrefix != "agit/" {
		t.Errorf("expected branch prefix agit/, got %s", cfg.Defaults.BranchPrefix)
	}
	if cfg.Defaults.WorktreeDir != ".worktrees" {
		t.Errorf("expected worktree dir .worktrees, got %s", cfg.Defaults.WorktreeDir)
	}
	if cfg.Defaults.CleanupStaleAfter != "24h" {
		t.Errorf("expected cleanup_stale_after 24h, got %s", cfg.Defaults.CleanupStaleAfter)
	}
	if !cfg.Defaults.AutoConflictCheck {
		t.Error("expected auto_conflict_check true")
	}
	if cfg.Agent.HeartbeatInterval != "30s" {
		t.Errorf("expected heartbeat_interval 30s, got %s", cfg.Agent.HeartbeatInterval)
	}
	if cfg.Agent.StaleAfter != "5m" {
		t.Errorf("expected stale_after 5m, got %s", cfg.Agent.StaleAfter)
	}
	if !cfg.Updates.Enabled {
		t.Error("expected updates.enabled true")
	}
	if cfg.Updates.CheckInterval != "24h" {
		t.Errorf("expected check_interval 24h, got %s", cfg.Updates.CheckInterval)
	}
}

func TestLoadMissingFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should return defaults
	if cfg.Server.Transport != "stdio" {
		t.Errorf("expected default transport, got %s", cfg.Server.Transport)
	}
}

func TestLoadExistingFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := filepath.Join(home, ".agit")
	os.MkdirAll(dir, 0755)

	content := []byte("[server]\ntransport = \"sse\"\nport = 9999\n")
	os.WriteFile(filepath.Join(dir, "config.toml"), content, 0644)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server.Transport != "sse" {
		t.Errorf("expected sse, got %s", cfg.Server.Transport)
	}
	if cfg.Server.Port != 9999 {
		t.Errorf("expected port 9999, got %d", cfg.Server.Port)
	}
	// Unspecified values should be defaults
	if cfg.Defaults.BranchPrefix != "agit/" {
		t.Errorf("expected default branch prefix, got %s", cfg.Defaults.BranchPrefix)
	}
}

func TestLoadMalformed(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := filepath.Join(home, ".agit")
	os.MkdirAll(dir, 0755)

	content := []byte("this is not valid TOML {{{}}")
	os.WriteFile(filepath.Join(dir, "config.toml"), content, 0644)

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for malformed TOML")
	}
}

func TestSaveAndLoad(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if err := EnsureDir(); err != nil {
		t.Fatalf("could not create dir: %v", err)
	}

	cfg := DefaultConfig()
	cfg.Server.Transport = "sse"
	cfg.Server.Port = 4000
	cfg.UI.Color = "never"

	if err := Save(cfg); err != nil {
		t.Fatalf("could not save: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("could not load: %v", err)
	}

	if loaded.Server.Transport != "sse" {
		t.Errorf("expected sse, got %s", loaded.Server.Transport)
	}
	if loaded.Server.Port != 4000 {
		t.Errorf("expected 4000, got %d", loaded.Server.Port)
	}
	if loaded.UI.Color != "never" {
		t.Errorf("expected never, got %s", loaded.UI.Color)
	}
}

func TestConfigPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	path, err := ConfigPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := filepath.Join(home, ".agit", "config.toml")
	if path != expected {
		t.Errorf("expected %s, got %s", expected, path)
	}
}

func TestDBPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	path, err := DBPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := filepath.Join(home, ".agit", "agit.db")
	if path != expected {
		t.Errorf("expected %s, got %s", expected, path)
	}
}

func TestEnsureDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if err := EnsureDir(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dir := filepath.Join(home, ".agit")
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected directory")
	}
}

func TestValidate(t *testing.T) {
	cfg := DefaultConfig()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("default config should be valid: %v", err)
	}

	// Invalid transport
	bad := DefaultConfig()
	bad.Server.Transport = "http"
	if err := bad.Validate(); err == nil {
		t.Error("expected error for invalid transport")
	}

	// Invalid port
	bad = DefaultConfig()
	bad.Server.Port = 0
	if err := bad.Validate(); err == nil {
		t.Error("expected error for invalid port")
	}

	// Invalid duration
	bad = DefaultConfig()
	bad.Agent.HeartbeatInterval = "notaduration"
	if err := bad.Validate(); err == nil {
		t.Error("expected error for invalid duration")
	}

	// Invalid color
	bad = DefaultConfig()
	bad.UI.Color = "rainbow"
	if err := bad.Validate(); err == nil {
		t.Error("expected error for invalid color")
	}

	// Invalid output format
	bad = DefaultConfig()
	bad.UI.OutputFormat = "xml"
	if err := bad.Validate(); err == nil {
		t.Error("expected error for invalid output format")
	}
}

func TestSetByDotKey(t *testing.T) {
	cfg := DefaultConfig()

	tests := []struct {
		key, value string
		check      func() bool
	}{
		{"server.transport", "sse", func() bool { return cfg.Server.Transport == "sse" }},
		{"server.port", "8080", func() bool { return cfg.Server.Port == 8080 }},
		{"defaults.branch_prefix", "feat/", func() bool { return cfg.Defaults.BranchPrefix == "feat/" }},
		{"defaults.worktree_dir", ".wt", func() bool { return cfg.Defaults.WorktreeDir == ".wt" }},
		{"defaults.cleanup_stale_after", "48h", func() bool { return cfg.Defaults.CleanupStaleAfter == "48h" }},
		{"defaults.auto_conflict_check", "false", func() bool { return !cfg.Defaults.AutoConflictCheck }},
		{"agent.heartbeat_interval", "1m", func() bool { return cfg.Agent.HeartbeatInterval == "1m" }},
		{"agent.stale_after", "10m", func() bool { return cfg.Agent.StaleAfter == "10m" }},
		{"ui.color", "never", func() bool { return cfg.UI.Color == "never" }},
		{"ui.output_format", "json", func() bool { return cfg.UI.OutputFormat == "json" }},
		{"ui.compact", "true", func() bool { return cfg.UI.Compact }},
		{"updates.enabled", "false", func() bool { return !cfg.Updates.Enabled }},
		{"updates.check_interval", "12h", func() bool { return cfg.Updates.CheckInterval == "12h" }},
	}

	for _, tt := range tests {
		if err := cfg.SetByDotKey(tt.key, tt.value); err != nil {
			t.Errorf("SetByDotKey(%s, %s) error: %v", tt.key, tt.value, err)
			continue
		}
		if !tt.check() {
			t.Errorf("SetByDotKey(%s, %s) did not apply", tt.key, tt.value)
		}
	}

	// Unknown key
	if err := cfg.SetByDotKey("unknown.key", "val"); err == nil {
		t.Error("expected error for unknown key")
	}
}

func TestGetByDotKey(t *testing.T) {
	cfg := DefaultConfig()

	for _, key := range AllKeys() {
		val, err := cfg.GetByDotKey(key)
		if err != nil {
			t.Errorf("GetByDotKey(%s) error: %v", key, err)
		}
		if val == "" && key != "ui.color" && key != "ui.output_format" {
			t.Errorf("GetByDotKey(%s) returned empty string", key)
		}
	}

	// Unknown key
	if _, err := cfg.GetByDotKey("unknown.key"); err == nil {
		t.Error("expected error for unknown key")
	}
}
