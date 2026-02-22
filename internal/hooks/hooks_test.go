package hooks

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fathindos/agit/internal/config"
)

func TestFireWritesFile(t *testing.T) {
	dir := t.TempDir()
	outFile := filepath.Join(dir, "hook_output.txt")

	cfg := config.DefaultConfig()
	cfg.Hooks = map[string]string{
		"worktree.created": "echo $AGIT_EVENT > " + outFile,
	}
	cfg.HookTimeout = "5s"

	r := NewRunner(cfg)
	r.Fire("worktree.created", map[string]string{
		"AGIT_REPO": "test-repo",
	})

	// Wait for async hook to complete
	deadline := time.After(3 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatal("hook did not write file within 3s")
		default:
			if _, err := os.Stat(outFile); err == nil {
				data, _ := os.ReadFile(outFile)
				if len(data) == 0 {
					time.Sleep(50 * time.Millisecond)
					continue
				}
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

func TestFireTimeout(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Hooks = map[string]string{
		"slow.event": "sleep 10",
	}
	cfg.HookTimeout = "100ms"

	r := NewRunner(cfg)
	start := time.Now()
	r.Fire("slow.event", nil)

	// Fire returns immediately (async)
	if time.Since(start) > 50*time.Millisecond {
		t.Error("Fire should return immediately, took too long")
	}

	// Wait for the goroutine to finish (timeout should kill it)
	time.Sleep(300 * time.Millisecond)
}

func TestFireNoHookConfigured(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Hooks = map[string]string{
		"other.event": "echo hello",
	}

	r := NewRunner(cfg)
	// Fire for an event with no hook — should be a no-op
	r.Fire("nonexistent.event", nil)
}

func TestFireNilRunner(t *testing.T) {
	// Fire on nil runner should not panic
	var r *Runner
	r.Fire("any.event", nil)
}

func TestNewRunnerNoHooks(t *testing.T) {
	cfg := config.DefaultConfig()
	r := NewRunner(cfg)
	if r != nil {
		t.Error("expected nil runner when no hooks configured")
	}
}

func TestFireReturnsQuickly(t *testing.T) {
	dir := t.TempDir()
	outFile := filepath.Join(dir, "async_test.txt")

	cfg := config.DefaultConfig()
	cfg.Hooks = map[string]string{
		"test.event": "sleep 0.5 && echo done > " + outFile,
	}

	r := NewRunner(cfg)
	start := time.Now()
	r.Fire("test.event", map[string]string{"AGIT_REPO": "test"})
	elapsed := time.Since(start)

	if elapsed > 50*time.Millisecond {
		t.Errorf("Fire should return in <50ms (async), took %v", elapsed)
	}
}
