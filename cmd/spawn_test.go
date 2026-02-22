package cmd

import (
	"strings"
	"testing"
)

func TestSpawnSuccess(t *testing.T) {
	repoPath := setupTestGitRepo(t)
	env := newTestEnv(t)
	env.init()

	_, err := env.run("add", repoPath)
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}

	stdout, err := env.run("spawn", "test-repo", "--task", "test task", "--agent", "test-agent")
	if err != nil {
		t.Fatalf("spawn failed: %v", err)
	}
	if !strings.Contains(stdout, "Created worktree") {
		t.Errorf("expected 'Created worktree' in output, got: %s", stdout)
	}
}

func TestSpawnJSON(t *testing.T) {
	repoPath := setupTestGitRepo(t)
	env := newTestEnv(t)
	env.init()

	_, err := env.run("add", repoPath)
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}

	stdout, err := env.runJSON("spawn", "test-repo", "--task", "test task", "--agent", "test-agent")
	if err != nil {
		t.Fatalf("spawn --output json failed: %v", err)
	}
	if !strings.Contains(stdout, `"status"`) {
		t.Errorf("expected JSON with status field, got: %s", stdout)
	}
}

func TestSpawnMissingRepo(t *testing.T) {
	_, err := executeCommandWithInit(t, "spawn")
	if err == nil {
		t.Error("expected error when no repo specified")
	}
}

func TestSpawnRepoNotFound(t *testing.T) {
	_, err := executeCommandWithInit(t, "spawn", "nonexistent-repo")
	if err == nil {
		t.Error("expected error for nonexistent repo")
	}
}
