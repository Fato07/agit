package git

import (
	"fmt"
	"strings"
)

// MergeBranch merges a branch into the current branch
func MergeBranch(repoPath, branchName string) error {
	_, err := runGit(repoPath, "merge", branchName, "--no-ff", "-m",
		fmt.Sprintf("Merge %s via agit", branchName))
	if err != nil {
		return fmt.Errorf("merge failed: %w", err)
	}
	return nil
}

// CheckoutBranch switches to a branch
func CheckoutBranch(repoPath, branchName string) error {
	_, err := runGit(repoPath, "checkout", branchName)
	if err != nil {
		return fmt.Errorf("could not checkout branch: %w", err)
	}
	return nil
}

// HasUnmergedChanges checks if a branch has changes not in the base branch
func HasUnmergedChanges(repoPath, baseBranch, branch string) (bool, error) {
	out, err := runGit(repoPath, "log", "--oneline", baseBranch+".."+branch)
	if err != nil {
		return false, fmt.Errorf("could not check unmerged changes: %w", err)
	}
	return strings.TrimSpace(out) != "", nil
}

// CanMergeCleanly does a dry-run merge to check for conflicts
func CanMergeCleanly(repoPath, branchName string) (bool, error) {
	// Try a merge with --no-commit --no-ff
	err := runGitNoOutput(repoPath, "merge", "--no-commit", "--no-ff", branchName)
	// Abort the merge attempt regardless of outcome
	runGitNoOutput(repoPath, "merge", "--abort")

	if err != nil {
		return false, nil // merge would conflict
	}
	return true, nil
}
