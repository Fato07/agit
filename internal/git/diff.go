package git

import (
	"fmt"
	"strings"
)

// ModifiedFiles returns the list of files modified in a branch compared to a base branch
func ModifiedFiles(repoPath, baseBranch, branch string) ([]string, error) {
	out, err := runGit(repoPath, "diff", "--name-only", baseBranch+"..."+branch)
	if err != nil {
		return nil, fmt.Errorf("could not get modified files: %w", err)
	}

	var files []string
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}

// ModifiedFilesWithStatus returns files with their change type (A/M/D/R)
func ModifiedFilesWithStatus(repoPath, baseBranch, branch string) (map[string]string, error) {
	out, err := runGit(repoPath, "diff", "--name-status", baseBranch+"..."+branch)
	if err != nil {
		return nil, fmt.Errorf("could not get modified files: %w", err)
	}

	files := make(map[string]string)
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			continue
		}
		status := parts[0]
		path := parts[1]

		switch {
		case strings.HasPrefix(status, "A"):
			files[path] = "added"
		case strings.HasPrefix(status, "D"):
			files[path] = "deleted"
		case strings.HasPrefix(status, "R"):
			files[path] = "renamed"
		default:
			files[path] = "modified"
		}
	}
	return files, nil
}
