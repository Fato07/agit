package cmd

import (
	"strings"
	"testing"
)

func TestAgentsEmpty(t *testing.T) {
	stdout, err := executeCommandWithInit(t, "agents")
	if err != nil {
		t.Fatalf("agents failed: %v", err)
	}
	if !strings.Contains(stdout, "No agents") {
		t.Errorf("expected 'No agents' message, got: %s", stdout)
	}
}

func TestAgentsSweep(t *testing.T) {
	stdout, err := executeCommandWithInit(t, "agents", "--sweep")
	if err != nil {
		t.Fatalf("agents --sweep failed: %v", err)
	}
	if !strings.Contains(stdout, "Swept") {
		t.Errorf("expected 'Swept' in output, got: %s", stdout)
	}
}

func TestAgentsJSON(t *testing.T) {
	stdout, err := executeCommandJSON(t, "agents")
	if err != nil {
		t.Fatalf("agents --output json failed: %v", err)
	}
	if !strings.Contains(stdout, "[") {
		t.Errorf("expected JSON array, got: %s", stdout)
	}
}

func TestAgentsSweepJSON(t *testing.T) {
	stdout, err := executeCommandJSON(t, "agents", "--sweep")
	if err != nil {
		t.Fatalf("agents --sweep --output json failed: %v", err)
	}
	if !strings.Contains(stdout, `"status"`) {
		t.Errorf("expected JSON with status field, got: %s", stdout)
	}
}

func TestAgentsListWithAgent(t *testing.T) {
	repoPath := setupTestGitRepo(t)
	env := newTestEnv(t)
	env.init()

	_, err := env.run("add", repoPath)
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}

	// Spawn registers an agent
	_, err = env.run("spawn", "test-repo", "--task", "agent test", "--agent", "my-agent")
	if err != nil {
		t.Fatalf("spawn failed: %v", err)
	}

	// List agents - should show the registered agent
	stdout, err := env.run("agents")
	if err != nil {
		t.Fatalf("agents failed: %v", err)
	}
	if !strings.Contains(stdout, "my-agent") {
		t.Errorf("expected 'my-agent' in output, got: %s", stdout)
	}
}

func TestAgentsListWithAgentJSON(t *testing.T) {
	repoPath := setupTestGitRepo(t)
	env := newTestEnv(t)
	env.init()

	_, err := env.run("add", repoPath)
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}

	_, err = env.run("spawn", "test-repo", "--task", "json agent test", "--agent", "json-agent")
	if err != nil {
		t.Fatalf("spawn failed: %v", err)
	}

	stdout, err := env.runJSON("agents")
	if err != nil {
		t.Fatalf("agents --output json failed: %v", err)
	}
	if !strings.Contains(stdout, `"json-agent"`) {
		t.Errorf("expected agent name in JSON output, got: %s", stdout)
	}
}

func TestAgentsRemove(t *testing.T) {
	repoPath := setupTestGitRepo(t)
	env := newTestEnv(t)
	env.init()

	_, err := env.run("add", repoPath)
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}

	_, err = env.run("spawn", "test-repo", "--task", "remove test", "--agent", "remove-me")
	if err != nil {
		t.Fatalf("spawn failed: %v", err)
	}

	stdout, err := env.run("agents", "--remove", "remove-me")
	if err != nil {
		t.Fatalf("agents --remove failed: %v", err)
	}
	if !strings.Contains(stdout, "Removed") {
		t.Errorf("expected 'Removed' in output, got: %s", stdout)
	}
}

func TestAgentsRemoveJSON(t *testing.T) {
	repoPath := setupTestGitRepo(t)
	env := newTestEnv(t)
	env.init()

	_, err := env.run("add", repoPath)
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}

	_, err = env.run("spawn", "test-repo", "--task", "remove json test", "--agent", "rm-json")
	if err != nil {
		t.Fatalf("spawn failed: %v", err)
	}

	stdout, err := env.runJSON("agents", "--remove", "rm-json")
	if err != nil {
		t.Fatalf("agents --remove --output json failed: %v", err)
	}
	if !strings.Contains(stdout, `"removed"`) {
		t.Errorf("expected 'removed' in JSON output, got: %s", stdout)
	}
}
