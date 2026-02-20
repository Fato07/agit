package conflicts

import (
	gitops "github.com/fathindos/agit/internal/git"
	"github.com/fathindos/agit/internal/registry"
)

// ScanAndUpdate scans all active worktrees for a repo and updates file_touches
func ScanAndUpdate(db *registry.DB, repo *registry.Repo) error {
	activeStatus := "active"
	worktrees, err := db.ListWorktrees(repo.ID, &activeStatus)
	if err != nil {
		return err
	}

	for _, wt := range worktrees {
		files, err := gitops.ModifiedFilesWithStatus(repo.Path, repo.DefaultBranch, wt.Branch)
		if err != nil {
			continue // skip worktrees we can't diff
		}

		var touches []registry.FileTouch
		for path, changeType := range files {
			touches = append(touches, registry.FileTouch{
				FilePath:   path,
				ChangeType: changeType,
			})
		}

		if err := db.RecordFileTouches(repo.ID, wt.ID, touches); err != nil {
			return err
		}
	}

	return nil
}

// Detect scans and returns conflicts for a repo
func Detect(db *registry.DB, repo *registry.Repo) ([]registry.Conflict, error) {
	if err := ScanAndUpdate(db, repo); err != nil {
		return nil, err
	}
	return db.FindConflicts(repo.ID)
}
