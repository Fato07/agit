package cmd

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/fathindos/agit/internal/ui"
)

// testEnv holds a reusable test environment for multi-command sequences.
type testEnv struct {
	t    *testing.T
	home string
}

// newTestEnv creates an isolated test environment with a temp HOME.
func newTestEnv(t *testing.T) *testEnv {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	return &testEnv{t: t, home: home}
}

// captureStdout redirects os.Stdout to a pipe and returns captured output after fn completes.
func captureStdout(fn func()) string {
	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	var buf bytes.Buffer
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		io.Copy(&buf, r)
	}()

	fn()

	w.Close()
	wg.Wait()
	os.Stdout = origStdout
	return buf.String()
}

// resetAllFlags resets all flags on a command and its subcommands to their defaults.
func resetAllFlags(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		f.Value.Set(f.DefValue)
		f.Changed = false
	})
	for _, sub := range cmd.Commands() {
		resetAllFlags(sub)
	}
}

// resetUIState resets all global UI state to defaults for test isolation.
func resetUIState() {
	color.NoColor = true
	ui.T = ui.DefaultTheme
	ui.Quiet = false
	ui.CurrentFormat = ui.FormatText
}

// run executes a command in this test environment and captures all stdout output.
func (e *testEnv) run(args ...string) (stdout string, err error) {
	e.t.Helper()

	e.t.Setenv("HOME", e.home)
	resetUIState()

	// Reset all flags to defaults to avoid leaking between test calls.
	// Cobra retains flag values across Execute() calls.
	resetAllFlags(rootCmd)

	// Suppress stderr noise
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs(args)

	stdout = captureStdout(func() {
		err = rootCmd.Execute()
	})

	// Reset UI state after execution to prevent leaking format changes
	resetUIState()

	return stdout, err
}

// runJSON executes a command with --output json.
func (e *testEnv) runJSON(args ...string) (stdout string, err error) {
	e.t.Helper()
	fullArgs := append([]string{"--output", "json"}, args...)
	return e.run(fullArgs...)
}

// init runs agit init (suppresses output).
func (e *testEnv) init() {
	e.t.Helper()
	_, err := e.run("init")
	if err != nil {
		e.t.Fatalf("agit init failed: %v", err)
	}
}

// executeCommand runs a command in a fresh test environment (one-shot).
func executeCommand(t *testing.T, args ...string) (stdout string, err error) {
	t.Helper()
	env := newTestEnv(t)
	return env.run(args...)
}

// executeCommandWithInit runs a command after initializing agit (one-shot).
func executeCommandWithInit(t *testing.T, args ...string) (stdout string, err error) {
	t.Helper()
	env := newTestEnv(t)
	env.init()
	return env.run(args...)
}

// executeCommandJSON runs a command with --output json after init (one-shot).
func executeCommandJSON(t *testing.T, args ...string) (stdout string, err error) {
	t.Helper()
	env := newTestEnv(t)
	env.init()
	return env.runJSON(args...)
}

// setupTestGitRepo creates a minimal git repo in a temp dir and returns its path.
// Hooks are disabled to prevent interference from system-level git hooks.
func setupTestGitRepo(t *testing.T) string {
	t.Helper()

	dir := filepath.Join(t.TempDir(), "test-repo")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("could not create test repo dir: %v", err)
	}

	runGit(t, dir, "init", "--initial-branch=main")
	runGit(t, dir, "config", "user.email", "test@agit.dev")
	runGit(t, dir, "config", "user.name", "agit-test")
	// Disable hooks to avoid interference from system-level hooks (e.g., kb-daemon)
	runGit(t, dir, "config", "core.hooksPath", "/dev/null")

	readmePath := filepath.Join(dir, "README.md")
	if err := os.WriteFile(readmePath, []byte("# Test Repo\n"), 0o644); err != nil {
		t.Fatalf("could not create README: %v", err)
	}
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "Initial commit")

	return dir
}

// writeFileInWorktree creates a file in a worktree directory.
func writeFileInWorktree(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("could not write file %s: %v", name, err)
	}
}

// runGit runs a git command in the given directory.
func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, out)
	}
}
