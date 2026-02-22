package cmd

import (
	"strings"
	"testing"
)

func TestAddSuccess(t *testing.T) {
	repoPath := setupTestGitRepo(t)
	stdout, err := executeCommandWithInit(t, "add", repoPath)
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}
	if !strings.Contains(stdout, "Registered") {
		t.Errorf("expected 'Registered' in output, got: %s", stdout)
	}
}

func TestAddJSON(t *testing.T) {
	repoPath := setupTestGitRepo(t)
	stdout, err := executeCommandJSON(t, "add", repoPath)
	if err != nil {
		t.Fatalf("add --output json failed: %v", err)
	}
	if !strings.Contains(stdout, `"status"`) {
		t.Errorf("expected JSON with status field, got: %s", stdout)
	}
}

func TestAddNotGitRepo(t *testing.T) {
	notRepo := t.TempDir()
	_, err := executeCommandWithInit(t, "add", notRepo)
	if err == nil {
		t.Error("expected error when adding non-git directory")
	}
}

func TestAddMissingPath(t *testing.T) {
	_, err := executeCommandWithInit(t, "add")
	if err == nil {
		t.Error("expected error when no path provided")
	}
}
