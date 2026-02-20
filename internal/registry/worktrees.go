package registry

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
)

// Worktree represents an agent worktree
type Worktree struct {
	ID              string
	RepoID          string
	Path            string
	Branch          string
	AgentID         *string
	TaskDescription *string
	Status          string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// CreateWorktree records a new worktree in the registry
func (db *DB) CreateWorktree(repoID, path, branch string, agentID, taskDesc *string) (*Worktree, error) {
	id := uuid.New().String()
	now := time.Now()

	_, err := db.conn.Exec(
		`INSERT INTO worktrees (id, repo_id, path, branch, agent_id, task_description, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, 'active', ?, ?)`,
		id, repoID, path, branch, agentID, taskDesc, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("could not create worktree record: %w", err)
	}

	return &Worktree{
		ID:              id,
		RepoID:          repoID,
		Path:            path,
		Branch:          branch,
		AgentID:         agentID,
		TaskDescription: taskDesc,
		Status:          "active",
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

// GetWorktree retrieves a worktree by ID
func (db *DB) GetWorktree(id string) (*Worktree, error) {
	wt := &Worktree{}
	err := db.conn.QueryRow(
		`SELECT id, repo_id, path, branch, agent_id, task_description, status, created_at, updated_at
		 FROM worktrees WHERE id = ?`, id,
	).Scan(&wt.ID, &wt.RepoID, &wt.Path, &wt.Branch, &wt.AgentID,
		&wt.TaskDescription, &wt.Status, &wt.CreatedAt, &wt.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("worktree %q not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("could not get worktree: %w", err)
	}
	return wt, nil
}

// ListWorktrees returns worktrees for a repo, optionally filtered by status
func (db *DB) ListWorktrees(repoID string, status *string) ([]*Worktree, error) {
	var rows *sql.Rows
	var err error

	if status != nil {
		rows, err = db.conn.Query(
			`SELECT id, repo_id, path, branch, agent_id, task_description, status, created_at, updated_at
			 FROM worktrees WHERE repo_id = ? AND status = ? ORDER BY created_at DESC`,
			repoID, *status,
		)
	} else {
		rows, err = db.conn.Query(
			`SELECT id, repo_id, path, branch, agent_id, task_description, status, created_at, updated_at
			 FROM worktrees WHERE repo_id = ? ORDER BY created_at DESC`,
			repoID,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("could not list worktrees: %w", err)
	}
	defer rows.Close()

	var worktrees []*Worktree
	for rows.Next() {
		wt := &Worktree{}
		if err := rows.Scan(&wt.ID, &wt.RepoID, &wt.Path, &wt.Branch, &wt.AgentID,
			&wt.TaskDescription, &wt.Status, &wt.CreatedAt, &wt.UpdatedAt); err != nil {
			return nil, fmt.Errorf("could not scan worktree: %w", err)
		}
		worktrees = append(worktrees, wt)
	}
	return worktrees, nil
}

// UpdateWorktreeStatus updates the status of a worktree
func (db *DB) UpdateWorktreeStatus(id, status string) error {
	result, err := db.conn.Exec(
		`UPDATE worktrees SET status = ?, updated_at = ? WHERE id = ?`,
		status, time.Now(), id,
	)
	if err != nil {
		return fmt.Errorf("could not update worktree status: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("worktree %q not found", id)
	}
	return nil
}

// DeleteWorktree removes a worktree record
func (db *DB) DeleteWorktree(id string) error {
	_, err := db.conn.Exec(`DELETE FROM worktrees WHERE id = ?`, id)
	return err
}

// PruneOrphanedWorktrees marks active worktrees as stale if their directory no longer exists
func (db *DB) PruneOrphanedWorktrees(repoID string) (int, error) {
	activeStatus := "active"
	worktrees, err := db.ListWorktrees(repoID, &activeStatus)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, wt := range worktrees {
		if _, err := os.Stat(wt.Path); os.IsNotExist(err) {
			if err := db.UpdateWorktreeStatus(wt.ID, "stale"); err != nil {
				return count, fmt.Errorf("could not mark worktree %s as stale: %w", wt.ID, err)
			}
			count++
		}
	}
	return count, nil
}

// FindWorktreeByPrefix finds a worktree by ID prefix within a repo
func (db *DB) FindWorktreeByPrefix(repoID, prefix string) (*Worktree, error) {
	if len(prefix) < 4 {
		return nil, fmt.Errorf("worktree ID prefix must be at least 4 characters")
	}

	rows, err := db.conn.Query(
		`SELECT id, repo_id, path, branch, agent_id, task_description, status, created_at, updated_at
		 FROM worktrees WHERE repo_id = ? AND id LIKE ?`,
		repoID, prefix+"%",
	)
	if err != nil {
		return nil, fmt.Errorf("could not search worktrees: %w", err)
	}
	defer rows.Close()

	var matches []*Worktree
	for rows.Next() {
		wt := &Worktree{}
		if err := rows.Scan(&wt.ID, &wt.RepoID, &wt.Path, &wt.Branch, &wt.AgentID,
			&wt.TaskDescription, &wt.Status, &wt.CreatedAt, &wt.UpdatedAt); err != nil {
			return nil, fmt.Errorf("could not scan worktree: %w", err)
		}
		matches = append(matches, wt)
	}

	switch len(matches) {
	case 0:
		return nil, fmt.Errorf("no worktree matching prefix %q", prefix)
	case 1:
		return matches[0], nil
	default:
		return nil, fmt.Errorf("ambiguous prefix %q: matches %d worktrees", prefix, len(matches))
	}
}

// ResolveWorktree finds a worktree by exact ID or prefix within a repo
func (db *DB) ResolveWorktree(repoID, idOrPrefix string) (*Worktree, error) {
	wt, err := db.GetWorktree(idOrPrefix)
	if err == nil {
		if wt.RepoID != repoID {
			return nil, fmt.Errorf("worktree %q does not belong to this repository", idOrPrefix)
		}
		return wt, nil
	}
	return db.FindWorktreeByPrefix(repoID, idOrPrefix)
}

// ListAllActiveWorktrees returns all active worktrees across all repos
func (db *DB) ListAllActiveWorktrees() ([]*Worktree, error) {
	rows, err := db.conn.Query(
		`SELECT id, repo_id, path, branch, agent_id, task_description, status, created_at, updated_at
		 FROM worktrees WHERE status = 'active' ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("could not list active worktrees: %w", err)
	}
	defer rows.Close()

	var worktrees []*Worktree
	for rows.Next() {
		wt := &Worktree{}
		if err := rows.Scan(&wt.ID, &wt.RepoID, &wt.Path, &wt.Branch, &wt.AgentID,
			&wt.TaskDescription, &wt.Status, &wt.CreatedAt, &wt.UpdatedAt); err != nil {
			return nil, fmt.Errorf("could not scan worktree: %w", err)
		}
		worktrees = append(worktrees, wt)
	}
	return worktrees, nil
}
