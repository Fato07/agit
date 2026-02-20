package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// IsGitRepo checks if the given path is a Git repository
func IsGitRepo(path string) bool {
	gitDir := filepath.Join(path, ".git")
	info, err := os.Stat(gitDir)
	if err != nil {
		return false
	}
	// .git can be a directory (normal repo) or a file (worktree)
	return info.IsDir() || info.Mode().IsRegular()
}

// GetRemoteURL returns the origin remote URL for a repo
func GetRemoteURL(repoPath string) (string, error) {
	out, err := runGit(repoPath, "remote", "get-url", "origin")
	if err != nil {
		return "", fmt.Errorf("could not get remote URL: %w", err)
	}
	return strings.TrimSpace(out), nil
}

// GetDefaultBranch returns the default branch name (main, master, etc.)
func GetDefaultBranch(repoPath string) (string, error) {
	// Try to get the HEAD symbolic ref of origin
	out, err := runGit(repoPath, "symbolic-ref", "refs/remotes/origin/HEAD", "--short")
	if err == nil {
		branch := strings.TrimSpace(out)
		// Remove "origin/" prefix
		parts := strings.SplitN(branch, "/", 2)
		if len(parts) == 2 {
			return parts[1], nil
		}
		return branch, nil
	}

	// Fallback: check if "main" or "master" branch exists
	for _, candidate := range []string{"main", "master"} {
		if err := runGitNoOutput(repoPath, "rev-parse", "--verify", candidate); err == nil {
			return candidate, nil
		}
	}

	// Last resort: whatever HEAD points to
	out, err = runGit(repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "main", nil // absolute fallback
	}
	return strings.TrimSpace(out), nil
}

// HasCommits checks if the repo has at least one commit
func HasCommits(repoPath string) bool {
	return runGitNoOutput(repoPath, "rev-parse", "HEAD") == nil
}

// BranchExists checks if a branch ref can be resolved
func BranchExists(repoPath, branch string) bool {
	return runGitNoOutput(repoPath, "rev-parse", "--verify", branch) == nil
}

// GetCurrentBranch returns the current branch name
func GetCurrentBranch(repoPath string) (string, error) {
	out, err := runGit(repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("could not get current branch: %w", err)
	}
	return strings.TrimSpace(out), nil
}

// runGit executes a git command in the given directory and returns stdout
func runGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), string(exitErr.Stderr))
		}
		return "", err
	}
	return string(out), nil
}

// runGitNoOutput executes a git command and only checks for success
func runGitNoOutput(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	return cmd.Run()
}
