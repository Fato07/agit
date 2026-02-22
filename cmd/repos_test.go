package cmd

import (
	"strings"
	"testing"
)

func TestReposEmpty(t *testing.T) {
	stdout, err := executeCommandWithInit(t, "repos")
	if err != nil {
		t.Fatalf("repos failed: %v", err)
	}
	if !strings.Contains(stdout, "No repositories") {
		t.Errorf("expected 'No repositories' message, got: %s", stdout)
	}
}

func TestReposAfterAdd(t *testing.T) {
	repoPath := setupTestGitRepo(t)
	env := newTestEnv(t)
	env.init()

	_, err := env.run("add", repoPath)
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}

	stdout, err := env.run("repos")
	if err != nil {
		t.Fatalf("repos failed: %v", err)
	}
	if !strings.Contains(stdout, "test-repo") {
		t.Errorf("expected repo name in output, got: %s", stdout)
	}
}

func TestReposJSON(t *testing.T) {
	stdout, err := executeCommandJSON(t, "repos")
	if err != nil {
		t.Fatalf("repos --output json failed: %v", err)
	}
	if !strings.Contains(stdout, "[") {
		t.Errorf("expected JSON array, got: %s", stdout)
	}
}

func TestReposAfterAddJSON(t *testing.T) {
	repoPath := setupTestGitRepo(t)
	env := newTestEnv(t)
	env.init()

	_, err := env.run("add", repoPath)
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}

	stdout, err := env.runJSON("repos")
	if err != nil {
		t.Fatalf("repos --output json failed: %v", err)
	}
	if !strings.Contains(stdout, `"test-repo"`) {
		t.Errorf("expected repo name in JSON output, got: %s", stdout)
	}
}
