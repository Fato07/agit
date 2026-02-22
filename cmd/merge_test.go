package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMergeMissingWorktreeID(t *testing.T) {
	_, err := executeCommandWithInit(t, "merge")
	if err == nil {
		t.Error("expected error when no worktree-id specified")
	}
}

func TestMergeNonexistentWorktree(t *testing.T) {
	_, err := executeCommandWithInit(t, "merge", "nonexistent-id")
	if err == nil {
		t.Error("expected error for nonexistent worktree ID")
	}
}

func TestMergeSuccess(t *testing.T) {
	repoPath := setupTestGitRepo(t)
	env := newTestEnv(t)
	env.init()

	_, err := env.run("add", repoPath)
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}

	// Use JSON output to reliably extract worktree ID and path
	stdout, err := env.runJSON("spawn", "test-repo", "--task", "merge test", "--agent", "m-agent")
	if err != nil {
		t.Fatalf("spawn failed: %v", err)
	}

	wtID, wtPath := extractSpawnJSON(t, stdout)

	// Create a file and commit in the worktree
	testFile := filepath.Join(wtPath, "test.txt")
	if err := os.WriteFile(testFile, []byte("merge test\n"), 0o644); err != nil {
		t.Fatalf("could not write test file: %v", err)
	}
	runGit(t, wtPath, "add", "test.txt")
	runGit(t, wtPath, "commit", "-m", "Add test file")

	// Merge
	stdout, err = env.run("merge", wtID, "--skip-conflict-check")
	if err != nil {
		t.Fatalf("merge failed: %v\nOutput: %s", err, stdout)
	}
	if !strings.Contains(stdout, "Merged") {
		t.Errorf("expected 'Merged' in output, got: %s", stdout)
	}
}

func TestMergeWithCleanup(t *testing.T) {
	repoPath := setupTestGitRepo(t)
	env := newTestEnv(t)
	env.init()

	_, err := env.run("add", repoPath)
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}

	stdout, err := env.runJSON("spawn", "test-repo", "--task", "cleanup merge", "--agent", "c-agent")
	if err != nil {
		t.Fatalf("spawn failed: %v", err)
	}

	wtID, wtPath := extractSpawnJSON(t, stdout)

	testFile := filepath.Join(wtPath, "cleanup.txt")
	if err := os.WriteFile(testFile, []byte("cleanup merge\n"), 0o644); err != nil {
		t.Fatalf("could not write test file: %v", err)
	}
	runGit(t, wtPath, "add", "cleanup.txt")
	runGit(t, wtPath, "commit", "-m", "Add cleanup file")

	stdout, err = env.run("merge", wtID, "--skip-conflict-check", "--cleanup")
	if err != nil {
		t.Fatalf("merge --cleanup failed: %v\nOutput: %s", err, stdout)
	}
	if !strings.Contains(stdout, "Merged") {
		t.Errorf("expected 'Merged' in output, got: %s", stdout)
	}
}

func TestMergeJSON(t *testing.T) {
	repoPath := setupTestGitRepo(t)
	env := newTestEnv(t)
	env.init()

	_, err := env.run("add", repoPath)
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}

	stdout, err := env.runJSON("spawn", "test-repo", "--task", "json merge", "--agent", "jm-agent")
	if err != nil {
		t.Fatalf("spawn failed: %v", err)
	}

	wtID, wtPath := extractSpawnJSON(t, stdout)

	testFile := filepath.Join(wtPath, "json.txt")
	if err := os.WriteFile(testFile, []byte("json merge\n"), 0o644); err != nil {
		t.Fatalf("could not write test file: %v", err)
	}
	runGit(t, wtPath, "add", "json.txt")
	runGit(t, wtPath, "commit", "-m", "Add json file")

	stdout, err = env.runJSON("merge", wtID, "--skip-conflict-check")
	if err != nil {
		t.Fatalf("merge --output json failed: %v\nOutput: %s", err, stdout)
	}
	if !strings.Contains(stdout, `"merged"`) {
		t.Errorf("expected 'merged' in JSON output, got: %s", stdout)
	}
}

// extractSpawnJSON parses worktree ID and path from JSON spawn output.
func extractSpawnJSON(t *testing.T, output string) (id, path string) {
	t.Helper()
	var result map[string]string
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &result); err != nil {
		t.Fatalf("could not parse spawn JSON: %v\nOutput: %s", err, output)
	}
	id = result["worktree"]
	path = result["path"]
	if id == "" || path == "" {
		t.Fatalf("missing worktree or path in spawn JSON: %v", result)
	}
	return id, path
}
