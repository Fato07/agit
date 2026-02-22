package cmd

import (
	"strings"
	"testing"
)

func TestVersionOutput(t *testing.T) {
	stdout, err := executeCommand(t, "--version")
	if err != nil {
		t.Fatalf("--version failed: %v", err)
	}
	if !strings.Contains(stdout, Version) {
		t.Errorf("expected version %q in output, got: %s", Version, stdout)
	}
}

func TestHelpOutput(t *testing.T) {
	stdout, err := executeCommand(t, "--help")
	if err != nil {
		t.Fatalf("--help failed: %v", err)
	}
	if !strings.Contains(stdout, "agit") {
		t.Errorf("expected 'agit' in help output, got: %s", stdout)
	}
}
