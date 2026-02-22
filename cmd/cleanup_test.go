package cmd

import (
	"strings"
	"testing"
)

func TestCleanupNothingToClean(t *testing.T) {
	stdout, err := executeCommandWithInit(t, "cleanup")
	if err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}
	if !strings.Contains(stdout, "Nothing to clean") {
		t.Errorf("expected 'Nothing to clean' message, got: %s", stdout)
	}
}

func TestCleanupJSON(t *testing.T) {
	stdout, err := executeCommandJSON(t, "cleanup")
	if err != nil {
		t.Fatalf("cleanup --output json failed: %v", err)
	}
	if !strings.Contains(stdout, `"count"`) {
		t.Errorf("expected JSON with count field, got: %s", stdout)
	}
}

func TestCleanupAfterMerge(t *testing.T) {
	repoPath := setupTestGitRepo(t)
	env := newTestEnv(t)
	env.init()

	_, err := env.run("add", repoPath)
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}

	// Spawn and merge to create a "completed" worktree
	spawnOut, err := env.runJSON("spawn", "test-repo", "--task", "cleanup test", "--agent", "cu-agent")
	if err != nil {
		t.Fatalf("spawn failed: %v", err)
	}

	wtID, wtPath := extractSpawnJSON(t, spawnOut)

	writeFileInWorktree(t, wtPath, "cu.txt", "cleanup test\n")
	runGit(t, wtPath, "add", "cu.txt")
	runGit(t, wtPath, "commit", "-m", "Add cu file")

	_, err = env.run("merge", wtID, "--skip-conflict-check")
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}

	// Now cleanup should find and remove the completed worktree
	stdout, err := env.run("cleanup")
	if err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}
	if !strings.Contains(stdout, "Cleaned up") {
		t.Errorf("expected 'Cleaned up' in output, got: %s", stdout)
	}
}

func TestCleanupAllFlag(t *testing.T) {
	repoPath := setupTestGitRepo(t)
	env := newTestEnv(t)
	env.init()

	_, err := env.run("add", repoPath)
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}

	spawnOut, err := env.runJSON("spawn", "test-repo", "--task", "all cleanup", "--agent", "all-agent")
	if err != nil {
		t.Fatalf("spawn failed: %v", err)
	}

	wtID, wtPath := extractSpawnJSON(t, spawnOut)

	writeFileInWorktree(t, wtPath, "all.txt", "all cleanup\n")
	runGit(t, wtPath, "add", "all.txt")
	runGit(t, wtPath, "commit", "-m", "Add all file")

	_, err = env.run("merge", wtID, "--skip-conflict-check")
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}

	stdout, err := env.run("cleanup", "--all")
	if err != nil {
		t.Fatalf("cleanup --all failed: %v", err)
	}
	if !strings.Contains(stdout, "Cleaned up") && !strings.Contains(stdout, "Nothing") {
		t.Errorf("expected cleanup output, got: %s", stdout)
	}
}
