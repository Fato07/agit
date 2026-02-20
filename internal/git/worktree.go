package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CreateWorktree creates a new git worktree at the specified path on a new branch
func CreateWorktree(repoPath, worktreePath, branchName, baseBranch string) error {
	// Ensure the parent directory exists
	parentDir := filepath.Dir(worktreePath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("could not create worktree parent directory: %w", err)
	}

	// Resolve the base ref. If baseBranch doesn't exist, ensure there is
	// at least one commit (create an empty one if needed) and use HEAD.
	base := baseBranch
	if !BranchExists(repoPath, base) {
		// Check if repo has any commits at all
		if !HasCommits(repoPath) {
			if _, err := runGit(repoPath, "commit", "--allow-empty", "-m", "initial commit (auto-created by agit)"); err != nil {
				return fmt.Errorf("could not create initial commit for worktree: %w", err)
			}
		}
		base = "HEAD"
	}

	// Create the worktree with a new branch from the resolved base
	_, err := runGit(repoPath, "worktree", "add", "-b", branchName, worktreePath, base)
	if err != nil {
		return fmt.Errorf("could not create worktree: %w", err)
	}

	return nil
}

// RemoveWorktree removes a git worktree
func RemoveWorktree(repoPath, worktreePath string) error {
	_, err := runGit(repoPath, "worktree", "remove", worktreePath, "--force")
	if err != nil {
		// If git worktree remove fails, try manual cleanup
		os.RemoveAll(worktreePath)
		// Prune stale worktree refs
		runGit(repoPath, "worktree", "prune")
	}
	return nil
}

// ListWorktrees returns the list of git worktrees for a repo
func ListWorktrees(repoPath string) ([]string, error) {
	out, err := runGit(repoPath, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, fmt.Errorf("could not list worktrees: %w", err)
	}

	var paths []string
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "worktree ") {
			paths = append(paths, strings.TrimPrefix(line, "worktree "))
		}
	}
	return paths, nil
}

// DeleteBranch deletes a local branch
func DeleteBranch(repoPath, branchName string) error {
	_, err := runGit(repoPath, "branch", "-D", branchName)
	return err
}
