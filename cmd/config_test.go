package cmd

import (
	"strings"
	"testing"
)

func TestConfigShow(t *testing.T) {
	stdout, err := executeCommandWithInit(t, "config", "show")
	if err != nil {
		t.Fatalf("config show failed: %v", err)
	}
	if !strings.Contains(stdout, "transport") {
		t.Errorf("expected config content in output, got: %s", stdout)
	}
}

func TestConfigShowJSON(t *testing.T) {
	stdout, err := executeCommandJSON(t, "config", "show")
	if err != nil {
		t.Fatalf("config show --output json failed: %v", err)
	}
	if !strings.Contains(stdout, `"Transport"`) {
		t.Errorf("expected JSON config with Transport field, got: %s", stdout)
	}
}

func TestConfigSet(t *testing.T) {
	env := newTestEnv(t)
	env.init()

	stdout, err := env.run("config", "set", "ui.color", "never")
	if err != nil {
		t.Fatalf("config set failed: %v", err)
	}
	if !strings.Contains(stdout, "Set") {
		t.Errorf("expected 'Set' confirmation, got: %s", stdout)
	}

	// Verify it was set
	stdout, err = env.run("config", "show")
	if err != nil {
		t.Fatalf("config show failed: %v", err)
	}
	if !strings.Contains(stdout, "never") {
		t.Errorf("expected 'never' in config output, got: %s", stdout)
	}
}

func TestConfigSetInvalid(t *testing.T) {
	_, err := executeCommandWithInit(t, "config", "set", "server.transport", "invalid")
	if err == nil {
		t.Error("expected error for invalid transport value")
	}
}

func TestConfigPath(t *testing.T) {
	stdout, err := executeCommandWithInit(t, "config", "path")
	if err != nil {
		t.Fatalf("config path failed: %v", err)
	}
	if !strings.Contains(stdout, "config.toml") {
		t.Errorf("expected config path in output, got: %s", stdout)
	}
}

func TestConfigReset(t *testing.T) {
	env := newTestEnv(t)
	env.init()

	_, err := env.run("config", "set", "ui.color", "never")
	if err != nil {
		t.Fatalf("config set failed: %v", err)
	}

	stdout, err := env.run("config", "reset")
	if err != nil {
		t.Fatalf("config reset failed: %v", err)
	}
	if !strings.Contains(stdout, "reset") {
		t.Errorf("expected 'reset' in output, got: %s", stdout)
	}
}

func TestConfigSetJSON(t *testing.T) {
	env := newTestEnv(t)
	env.init()

	stdout, err := env.runJSON("config", "set", "ui.color", "always")
	if err != nil {
		t.Fatalf("config set --output json failed: %v", err)
	}
	if !strings.Contains(stdout, `"status"`) {
		t.Errorf("expected JSON with status field, got: %s", stdout)
	}
}
