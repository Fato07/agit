package registry

import (
	"fmt"
	"time"
)

// FileTouch represents a file modification in a worktree
type FileTouch struct {
	RepoID     string
	WorktreeID string
	FilePath   string
	ChangeType string
	UpdatedAt  time.Time
}

// RecordFileTouches replaces all file touch records for a worktree
func (db *DB) RecordFileTouches(repoID, worktreeID string, touches []FileTouch) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return fmt.Errorf("could not begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Clear existing touches for this worktree
	if _, err := tx.Exec(
		`DELETE FROM file_touches WHERE repo_id = ? AND worktree_id = ?`,
		repoID, worktreeID,
	); err != nil {
		return fmt.Errorf("could not clear file touches: %w", err)
	}

	// Insert new touches
	now := time.Now()
	stmt, err := tx.Prepare(
		`INSERT INTO file_touches (repo_id, worktree_id, file_path, change_type, updated_at)
		 VALUES (?, ?, ?, ?, ?)`,
	)
	if err != nil {
		return fmt.Errorf("could not prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, t := range touches {
		changeType := t.ChangeType
		if changeType == "" {
			changeType = "modified"
		}
		if _, err := stmt.Exec(repoID, worktreeID, t.FilePath, changeType, now); err != nil {
			return fmt.Errorf("could not insert file touch: %w", err)
		}
	}

	return tx.Commit()
}

// Conflict represents overlapping file modifications across worktrees
type Conflict struct {
	FilePath   string
	Worktrees  []string // worktree IDs
	AgentIDs   []string // corresponding agent IDs (may be empty)
	TaskDescs  []string // corresponding task descriptions (may be empty)
}

// FindConflicts detects files modified in multiple active worktrees for a repo
func (db *DB) FindConflicts(repoID string) ([]Conflict, error) {
	rows, err := db.conn.Query(
		`SELECT ft.file_path, ft.worktree_id, w.agent_id, w.task_description
		 FROM file_touches ft
		 JOIN worktrees w ON ft.worktree_id = w.id
		 WHERE ft.repo_id = ? AND w.status = 'active'
		 ORDER BY ft.file_path, ft.worktree_id`,
		repoID,
	)
	if err != nil {
		return nil, fmt.Errorf("could not query file touches: %w", err)
	}
	defer rows.Close()

	// Build map of file -> worktree details
	type wtDetail struct {
		worktreeID string
		agentID    string
		taskDesc   string
	}
	fileMap := make(map[string][]wtDetail)

	for rows.Next() {
		var filePath, worktreeID string
		var agentID, taskDesc *string
		if err := rows.Scan(&filePath, &worktreeID, &agentID, &taskDesc); err != nil {
			return nil, fmt.Errorf("could not scan row: %w", err)
		}
		aid := ""
		if agentID != nil {
			aid = *agentID
		}
		td := ""
		if taskDesc != nil {
			td = *taskDesc
		}
		fileMap[filePath] = append(fileMap[filePath], wtDetail{worktreeID, aid, td})
	}

	// Filter to files with 2+ worktrees
	var conflicts []Conflict
	for filePath, details := range fileMap {
		if len(details) < 2 {
			continue
		}
		c := Conflict{FilePath: filePath}
		for _, d := range details {
			c.Worktrees = append(c.Worktrees, d.worktreeID)
			c.AgentIDs = append(c.AgentIDs, d.agentID)
			c.TaskDescs = append(c.TaskDescs, d.taskDesc)
		}
		conflicts = append(conflicts, c)
	}

	return conflicts, nil
}
