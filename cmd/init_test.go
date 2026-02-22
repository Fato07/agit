package cmd

import (
	"strings"
	"testing"
)

func TestInitSuccess(t *testing.T) {
	stdout, err := executeCommand(t, "init")
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if !strings.Contains(stdout, "initialized") {
		t.Errorf("expected 'initialized' in output, got: %s", stdout)
	}
}

func TestInitJSON(t *testing.T) {
	stdout, err := executeCommand(t, "--output", "json", "init")
	if err != nil {
		t.Fatalf("init --output json failed: %v", err)
	}
	if !strings.Contains(stdout, `"status"`) {
		t.Errorf("expected JSON with status field, got: %s", stdout)
	}
}

func TestInitIdempotent(t *testing.T) {
	env := newTestEnv(t)
	_, err := env.run("init")
	if err != nil {
		t.Fatalf("first init failed: %v", err)
	}
	_, err = env.run("init")
	if err != nil {
		t.Errorf("second init should be idempotent, got error: %v", err)
	}
}
