package cmd

import (
	"os/exec"
	"strings"
	"testing"
)

func TestConflictsNoConflicts(t *testing.T) {
	repoPath := setupTestGitRepo(t)
	env := newTestEnv(t)
	env.init()

	_, err := env.run("add", repoPath)
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}

	stdout, err := env.run("conflicts", "test-repo")
	if err != nil {
		t.Fatalf("conflicts failed: %v", err)
	}
	// With <2 worktrees, should mention no conflicts possible or similar
	if !strings.Contains(stdout, "worktree") && !strings.Contains(stdout, "conflict") {
		t.Errorf("expected conflict-related info in output, got: %s", stdout)
	}
}

func TestConflictsJSON(t *testing.T) {
	repoPath := setupTestGitRepo(t)
	env := newTestEnv(t)
	env.init()

	_, err := env.run("add", repoPath)
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}

	stdout, err := env.runJSON("conflicts", "test-repo")
	if err != nil {
		t.Fatalf("conflicts --output json failed: %v", err)
	}
	if !strings.Contains(stdout, "conflicts") {
		t.Errorf("expected JSON with conflicts field, got: %s", stdout)
	}
}

func TestConflictsRepoNotFound(t *testing.T) {
	_, err := executeCommandWithInit(t, "conflicts", "nonexistent-repo")
	if err == nil {
		t.Error("expected error for nonexistent repo")
	}
}

func TestConflictsWithWorktrees(t *testing.T) {
	repoPath := setupTestGitRepo(t)
	env := newTestEnv(t)
	env.init()

	_, err := env.run("add", repoPath)
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}

	// Create two worktrees
	_, err = env.run("spawn", "test-repo", "--task", "first task", "--agent", "a1")
	if err != nil {
		t.Fatalf("spawn 1 failed: %v", err)
	}
	_, err = env.run("spawn", "test-repo", "--task", "second task", "--agent", "a2")
	if err != nil {
		t.Fatalf("spawn 2 failed: %v", err)
	}

	// Now check conflicts - even without overlapping files, it should run the check
	stdout, err := env.run("conflicts", "test-repo")
	if err != nil {
		t.Fatalf("conflicts failed: %v", err)
	}
	// With 2 worktrees, the conflict check should run (may show "No conflicts detected")
	if !strings.Contains(stdout, "conflict") && !strings.Contains(stdout, "Conflict") && !strings.Contains(stdout, "No") {
		t.Errorf("expected conflict check output, got: %s", stdout)
	}
}

func TestConflictsWithWorktreesJSON(t *testing.T) {
	repoPath := setupTestGitRepo(t)
	env := newTestEnv(t)
	env.init()

	_, err := env.run("add", repoPath)
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}

	_, err = env.run("spawn", "test-repo", "--task", "json task 1", "--agent", "j1")
	if err != nil {
		t.Fatalf("spawn 1 failed: %v", err)
	}
	_, err = env.run("spawn", "test-repo", "--task", "json task 2", "--agent", "j2")
	if err != nil {
		t.Fatalf("spawn 2 failed: %v", err)
	}

	stdout, err := env.runJSON("conflicts", "test-repo")
	if err != nil {
		t.Fatalf("conflicts --output json failed: %v", err)
	}
	if !strings.Contains(stdout, "conflicts") {
		t.Errorf("expected JSON with conflicts field, got: %s", stdout)
	}
}

func TestConflictsWithSuggestions(t *testing.T) {
	repoPath := setupTestGitRepo(t)
	env := newTestEnv(t)
	env.init()

	_, err := env.run("add", repoPath)
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}

	// Spawn two worktrees
	stdout1, err := env.runJSON("spawn", "test-repo", "--task", "task1", "--agent", "a1")
	if err != nil {
		t.Fatalf("spawn 1 failed: %v", err)
	}
	_, path1 := extractSpawnJSON(t, stdout1)

	stdout2, err := env.runJSON("spawn", "test-repo", "--task", "task2", "--agent", "a2")
	if err != nil {
		t.Fatalf("spawn 2 failed: %v", err)
	}
	_, path2 := extractSpawnJSON(t, stdout2)

	// Modify the same file in both worktrees and commit
	for _, wtPath := range []string{path1, path2} {
		writeFileInWorktree(t, wtPath, "shared.go", "package main\n// modified in "+wtPath+"\n")
		cmd := exec.Command("git", "-C", wtPath, "add", "shared.go")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git add failed in %s: %v\n%s", wtPath, err, out)
		}
		cmd = exec.Command("git", "-C", wtPath, "commit", "-m", "modify shared.go")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git commit failed in %s: %v\n%s", wtPath, err, out)
		}
	}

	// Run conflicts check — should detect overlapping shared.go and suggest resolution
	stdout, err := env.run("conflicts", "test-repo")
	if err != nil {
		t.Fatalf("conflicts failed: %v", err)
	}
	if !strings.Contains(stdout, "Suggested resolution order") {
		t.Errorf("expected 'Suggested resolution order' in output, got: %s", stdout)
	}

	// Verify JSON output includes suggestions
	stdoutJSON, err := env.runJSON("conflicts", "test-repo")
	if err != nil {
		t.Fatalf("conflicts JSON failed: %v", err)
	}
	if !strings.Contains(stdoutJSON, "suggestions") {
		t.Errorf("expected 'suggestions' in JSON output, got: %s", stdoutJSON)
	}
}
