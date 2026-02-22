package cmd

import (
	"testing"
)

func TestUpdateHandlesNoNetwork(t *testing.T) {
	// The update command tries to fetch from GitHub.
	// In test environment this may fail with a network error or succeed.
	// We verify it doesn't panic and returns gracefully.
	_, err := executeCommandWithInit(t, "update")
	// Error is acceptable (no network), but shouldn't panic
	_ = err
}

func TestUpdateJSON(t *testing.T) {
	stdout, err := executeCommandJSON(t, "update")
	// In JSON mode, update should return structured output even on error
	_ = err
	if len(stdout) > 0 && stdout[0] != '{' {
		t.Errorf("expected JSON output, got: %s", stdout)
	}
}
