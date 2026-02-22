package cmd

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestHookFiresOnSpawn(t *testing.T) {
	repoPath := setupTestGitRepo(t)
	env := newTestEnv(t)
	env.init()

	_, err := env.run("add", repoPath)
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}

	// Configure a hook that writes to a file on worktree.created
	hookOutput := filepath.Join(env.home, "hook_output.txt")
	_, err = env.run("config", "set", "hooks.worktree.created", "echo $AGIT_EVENT > "+hookOutput)
	if err != nil {
		t.Fatalf("config set hook failed: %v", err)
	}

	// Spawn a worktree
	_, err = env.run("spawn", "test-repo", "--task", "hook test", "--agent", "hook-agent")
	if err != nil {
		t.Fatalf("spawn failed: %v", err)
	}

	// Wait for the async hook to execute
	deadline := time.After(3 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatal("hook did not write file within 3s")
		default:
			if data, err := os.ReadFile(hookOutput); err == nil && len(data) > 0 {
				got := string(data)
				if got != "worktree.created\n" {
					t.Errorf("expected 'worktree.created\\n', got %q", got)
				}
				return
			}
			time.Sleep(50 * time.Millisecond)
		}
	}
}
