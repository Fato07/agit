package cmd

import (
	"strings"
	"testing"
)

func TestTasksCreateAndList(t *testing.T) {
	repoPath := setupTestGitRepo(t)
	env := newTestEnv(t)
	env.init()

	_, err := env.run("add", repoPath)
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}

	// Create a task
	stdout, err := env.run("tasks", "test-repo", "--create", "implement feature X")
	if err != nil {
		t.Fatalf("tasks --create failed: %v", err)
	}
	if !strings.Contains(stdout, "Created task") {
		t.Errorf("expected 'Created task' in output, got: %s", stdout)
	}

	// List tasks
	stdout, err = env.run("tasks", "test-repo")
	if err != nil {
		t.Fatalf("tasks list failed: %v", err)
	}
	if !strings.Contains(stdout, "implement feature X") {
		t.Errorf("expected task description in list, got: %s", stdout)
	}
}

func TestTasksClaimAndComplete(t *testing.T) {
	repoPath := setupTestGitRepo(t)
	env := newTestEnv(t)
	env.init()

	_, err := env.run("add", repoPath)
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}

	// Create task
	stdout, err := env.run("tasks", "test-repo", "--create", "test task")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	taskID := extractTaskID(t, stdout)

	// Claim it
	stdout, err = env.run("tasks", "test-repo", "--claim", taskID, "--agent", "claude-1")
	if err != nil {
		t.Fatalf("claim failed: %v", err)
	}
	if !strings.Contains(stdout, "claimed") {
		t.Errorf("expected 'claimed' in output, got: %s", stdout)
	}

	// Complete it
	stdout, err = env.run("tasks", "test-repo", "--complete", taskID, "--result", "done")
	if err != nil {
		t.Fatalf("complete failed: %v", err)
	}
	if !strings.Contains(stdout, "completed") {
		t.Errorf("expected 'completed' in output, got: %s", stdout)
	}
}

func TestTasksFail(t *testing.T) {
	repoPath := setupTestGitRepo(t)
	env := newTestEnv(t)
	env.init()

	_, err := env.run("add", repoPath)
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}

	// Create and claim
	stdout, err := env.run("tasks", "test-repo", "--create", "failing task")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	taskID := extractTaskID(t, stdout)

	_, err = env.run("tasks", "test-repo", "--claim", taskID, "--agent", "agent-1")
	if err != nil {
		t.Fatalf("claim failed: %v", err)
	}

	// Fail it
	stdout, err = env.run("tasks", "test-repo", "--fail", taskID, "--result", "error occurred")
	if err != nil {
		t.Fatalf("fail failed: %v", err)
	}
	if !strings.Contains(stdout, "failed") {
		t.Errorf("expected 'failed' in output, got: %s", stdout)
	}
}

func TestTasksJSON(t *testing.T) {
	repoPath := setupTestGitRepo(t)
	env := newTestEnv(t)
	env.init()

	_, err := env.run("add", repoPath)
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}

	// Create task then list as JSON
	_, err = env.run("tasks", "test-repo", "--create", "json task")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	stdout, err := env.runJSON("tasks", "test-repo")
	if err != nil {
		t.Fatalf("tasks --output json failed: %v", err)
	}
	if !strings.Contains(stdout, `"description"`) {
		t.Errorf("expected JSON with description field, got: %s", stdout)
	}
}

func TestTasksMissingRepo(t *testing.T) {
	_, err := executeCommandWithInit(t, "tasks")
	if err == nil {
		t.Error("expected error when no repo specified")
	}
}

func TestTasksCreateJSON(t *testing.T) {
	repoPath := setupTestGitRepo(t)
	env := newTestEnv(t)
	env.init()

	_, err := env.run("add", repoPath)
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}

	stdout, err := env.runJSON("tasks", "test-repo", "--create", "json task creation")
	if err != nil {
		t.Fatalf("tasks --create --output json failed: %v", err)
	}
	if !strings.Contains(stdout, `"status"`) {
		t.Errorf("expected JSON with status field, got: %s", stdout)
	}
}

func TestTasksClaimJSON(t *testing.T) {
	repoPath := setupTestGitRepo(t)
	env := newTestEnv(t)
	env.init()

	_, err := env.run("add", repoPath)
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}

	stdout, err := env.run("tasks", "test-repo", "--create", "claim json task")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	taskID := extractTaskID(t, stdout)

	stdout, err = env.runJSON("tasks", "test-repo", "--claim", taskID, "--agent", "j-agent")
	if err != nil {
		t.Fatalf("claim --output json failed: %v", err)
	}
	if !strings.Contains(stdout, `"status"`) {
		t.Errorf("expected JSON with status field, got: %s", stdout)
	}
}

func TestTasksCompleteJSON(t *testing.T) {
	repoPath := setupTestGitRepo(t)
	env := newTestEnv(t)
	env.init()

	_, err := env.run("add", repoPath)
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}

	stdout, err := env.run("tasks", "test-repo", "--create", "complete json task")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	taskID := extractTaskID(t, stdout)

	_, err = env.run("tasks", "test-repo", "--claim", taskID, "--agent", "cj-agent")
	if err != nil {
		t.Fatalf("claim failed: %v", err)
	}

	stdout, err = env.runJSON("tasks", "test-repo", "--complete", taskID, "--result", "done")
	if err != nil {
		t.Fatalf("complete --output json failed: %v", err)
	}
	if !strings.Contains(stdout, `"status"`) {
		t.Errorf("expected JSON with status field, got: %s", stdout)
	}
}

func TestTasksNextPriority(t *testing.T) {
	repoPath := setupTestGitRepo(t)
	env := newTestEnv(t)
	env.init()

	_, err := env.run("add", repoPath)
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}

	// Create tasks with different priorities
	_, err = env.run("tasks", "test-repo", "--create", "low prio task", "--priority", "1")
	if err != nil {
		t.Fatalf("create low failed: %v", err)
	}
	_, err = env.run("tasks", "test-repo", "--create", "high prio task", "--priority", "10")
	if err != nil {
		t.Fatalf("create high failed: %v", err)
	}
	_, err = env.run("tasks", "test-repo", "--create", "medium prio task", "--priority", "5")
	if err != nil {
		t.Fatalf("create medium failed: %v", err)
	}

	// Next should return the highest priority task
	stdout, err := env.run("tasks", "next", "test-repo", "--agent", "next-agent")
	if err != nil {
		t.Fatalf("tasks next failed: %v", err)
	}
	if !strings.Contains(stdout, "high prio task") {
		t.Errorf("expected highest priority task, got: %s", stdout)
	}
}

func TestTasksNextNoPending(t *testing.T) {
	repoPath := setupTestGitRepo(t)
	env := newTestEnv(t)
	env.init()

	_, err := env.run("add", repoPath)
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}

	stdout, err := env.run("tasks", "next", "test-repo", "--agent", "empty-agent")
	if err != nil {
		t.Fatalf("tasks next failed: %v", err)
	}
	if !strings.Contains(stdout, "No pending tasks") {
		t.Errorf("expected 'No pending tasks', got: %s", stdout)
	}
}

func TestTasksNextJSON(t *testing.T) {
	repoPath := setupTestGitRepo(t)
	env := newTestEnv(t)
	env.init()

	_, err := env.run("add", repoPath)
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}

	_, err = env.run("tasks", "test-repo", "--create", "json next task", "--priority", "5")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	stdout, err := env.runJSON("tasks", "next", "test-repo", "--agent", "json-next-agent")
	if err != nil {
		t.Fatalf("tasks next --output json failed: %v", err)
	}
	if !strings.Contains(stdout, `"claimed"`) {
		t.Errorf("expected 'claimed' in JSON output, got: %s", stdout)
	}
}

func TestTasksNextMissingAgent(t *testing.T) {
	_, err := executeCommandWithInit(t, "tasks", "next", "test-repo")
	if err == nil {
		t.Error("expected error when --agent missing")
	}
}

func TestTasksFailJSON(t *testing.T) {
	repoPath := setupTestGitRepo(t)
	env := newTestEnv(t)
	env.init()

	_, err := env.run("add", repoPath)
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}

	stdout, err := env.run("tasks", "test-repo", "--create", "fail json task")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	taskID := extractTaskID(t, stdout)

	_, err = env.run("tasks", "test-repo", "--claim", taskID, "--agent", "fj-agent")
	if err != nil {
		t.Fatalf("claim failed: %v", err)
	}

	stdout, err = env.runJSON("tasks", "test-repo", "--fail", taskID, "--result", "error")
	if err != nil {
		t.Fatalf("fail --output json failed: %v", err)
	}
	if !strings.Contains(stdout, `"status"`) {
		t.Errorf("expected JSON with status field, got: %s", stdout)
	}
}

// extractTaskID parses the task ID from "Created task: t-xxxxxxxx - description" output.
func extractTaskID(t *testing.T, output string) string {
	t.Helper()
	idx := strings.Index(output, "t-")
	if idx == -1 {
		t.Fatalf("could not find task ID in output: %s", output)
	}
	// Task IDs are "t-" followed by 8 hex chars
	end := idx + 10
	if end > len(output) {
		end = len(output)
	}
	id := output[idx:end]
	// Trim trailing non-hex characters
	for len(id) > 2 {
		c := id[len(id)-1]
		if (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || c == '-' {
			break
		}
		id = id[:len(id)-1]
	}
	return id
}
