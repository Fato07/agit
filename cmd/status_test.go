package cmd

import (
	"strings"
	"testing"
)

func TestStatusEmpty(t *testing.T) {
	stdout, err := executeCommandWithInit(t, "status")
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if !strings.Contains(stdout, "No repositories") {
		t.Errorf("expected 'No repositories' message, got: %s", stdout)
	}
}

func TestStatusWithRepo(t *testing.T) {
	repoPath := setupTestGitRepo(t)
	env := newTestEnv(t)
	env.init()

	_, err := env.run("add", repoPath)
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}

	stdout, err := env.run("status")
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if !strings.Contains(stdout, "test-repo") {
		t.Errorf("expected repo name in status output, got: %s", stdout)
	}
}

func TestStatusJSON(t *testing.T) {
	stdout, err := executeCommandJSON(t, "status")
	if err != nil {
		t.Fatalf("status --output json failed: %v", err)
	}
	if !strings.Contains(stdout, `"repos"`) {
		t.Errorf("expected JSON with repos field, got: %s", stdout)
	}
}

func TestStatusPerRepo(t *testing.T) {
	repoPath := setupTestGitRepo(t)
	env := newTestEnv(t)
	env.init()

	_, err := env.run("add", repoPath)
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}

	stdout, err := env.run("status", "test-repo")
	if err != nil {
		t.Fatalf("status test-repo failed: %v", err)
	}
	if !strings.Contains(stdout, "test-repo") {
		t.Errorf("expected repo name in output, got: %s", stdout)
	}
}

func TestStatusWithWorktree(t *testing.T) {
	repoPath := setupTestGitRepo(t)
	env := newTestEnv(t)
	env.init()

	_, err := env.run("add", repoPath)
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}

	_, err = env.run("spawn", "test-repo", "--task", "status test", "--agent", "s-agent")
	if err != nil {
		t.Fatalf("spawn failed: %v", err)
	}

	stdout, err := env.run("status")
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if !strings.Contains(stdout, "test-repo") {
		t.Errorf("expected repo in status output, got: %s", stdout)
	}
}

func TestStatusWithWorktreeJSON(t *testing.T) {
	repoPath := setupTestGitRepo(t)
	env := newTestEnv(t)
	env.init()

	_, err := env.run("add", repoPath)
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}

	_, err = env.run("spawn", "test-repo", "--task", "json status", "--agent", "js-agent")
	if err != nil {
		t.Fatalf("spawn failed: %v", err)
	}

	stdout, err := env.runJSON("status")
	if err != nil {
		t.Fatalf("status --output json failed: %v", err)
	}
	if !strings.Contains(stdout, `"repos"`) {
		t.Errorf("expected repos field in JSON, got: %s", stdout)
	}
}

func TestStatusPerRepoJSON(t *testing.T) {
	repoPath := setupTestGitRepo(t)
	env := newTestEnv(t)
	env.init()

	_, err := env.run("add", repoPath)
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}

	stdout, err := env.runJSON("status", "test-repo")
	if err != nil {
		t.Fatalf("status test-repo --output json failed: %v", err)
	}
	if !strings.Contains(stdout, `"name"`) {
		t.Errorf("expected name field in JSON, got: %s", stdout)
	}
}
