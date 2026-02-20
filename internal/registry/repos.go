package registry

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Repo represents a registered Git repository
type Repo struct {
	ID            string
	Name          string
	Path          string
	RemoteURL     string
	DefaultBranch string
	AddedAt       time.Time
	LastSynced    *time.Time
	Metadata      string
}

// AddRepo registers a new repository
func (db *DB) AddRepo(name, path, remoteURL, defaultBranch string) (*Repo, error) {
	id := uuid.New().String()
	now := time.Now()

	_, err := db.conn.Exec(
		`INSERT INTO repos (id, name, path, remote_url, default_branch, added_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		id, name, path, remoteURL, defaultBranch, now,
	)
	if err != nil {
		return nil, fmt.Errorf("could not add repo: %w", err)
	}

	return &Repo{
		ID:            id,
		Name:          name,
		Path:          path,
		RemoteURL:     remoteURL,
		DefaultBranch: defaultBranch,
		AddedAt:       now,
	}, nil
}

// GetRepo retrieves a repo by name
func (db *DB) GetRepo(name string) (*Repo, error) {
	repo := &Repo{}
	err := db.conn.QueryRow(
		`SELECT id, name, path, remote_url, default_branch, added_at, last_synced, metadata
		 FROM repos WHERE name = ?`, name,
	).Scan(&repo.ID, &repo.Name, &repo.Path, &repo.RemoteURL,
		&repo.DefaultBranch, &repo.AddedAt, &repo.LastSynced, &repo.Metadata)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("repo %q not found", name)
	}
	if err != nil {
		return nil, fmt.Errorf("could not get repo: %w", err)
	}
	return repo, nil
}

// GetRepoByID retrieves a repo by ID
func (db *DB) GetRepoByID(id string) (*Repo, error) {
	repo := &Repo{}
	err := db.conn.QueryRow(
		`SELECT id, name, path, remote_url, default_branch, added_at, last_synced, metadata
		 FROM repos WHERE id = ?`, id,
	).Scan(&repo.ID, &repo.Name, &repo.Path, &repo.RemoteURL,
		&repo.DefaultBranch, &repo.AddedAt, &repo.LastSynced, &repo.Metadata)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("repo with id %q not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("could not get repo: %w", err)
	}
	return repo, nil
}

// ListRepos returns all registered repositories
func (db *DB) ListRepos() ([]*Repo, error) {
	rows, err := db.conn.Query(
		`SELECT id, name, path, remote_url, default_branch, added_at, last_synced, metadata
		 FROM repos ORDER BY name`,
	)
	if err != nil {
		return nil, fmt.Errorf("could not list repos: %w", err)
	}
	defer rows.Close()

	var repos []*Repo
	for rows.Next() {
		repo := &Repo{}
		if err := rows.Scan(&repo.ID, &repo.Name, &repo.Path, &repo.RemoteURL,
			&repo.DefaultBranch, &repo.AddedAt, &repo.LastSynced, &repo.Metadata); err != nil {
			return nil, fmt.Errorf("could not scan repo: %w", err)
		}
		repos = append(repos, repo)
	}
	return repos, nil
}

// RemoveRepo deletes a repo by name (cascades to worktrees, tasks, file_touches)
func (db *DB) RemoveRepo(name string) error {
	result, err := db.conn.Exec(`DELETE FROM repos WHERE name = ?`, name)
	if err != nil {
		return fmt.Errorf("could not remove repo: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("repo %q not found", name)
	}
	return nil
}

// RepoStats holds summary stats for a repo
type RepoStats struct {
	ActiveWorktrees int
	PendingTasks    int
}

// GetRepoStats returns worktree and task counts for a repo
func (db *DB) GetRepoStats(repoID string) (*RepoStats, error) {
	stats := &RepoStats{}

	err := db.conn.QueryRow(
		`SELECT COUNT(*) FROM worktrees WHERE repo_id = ? AND status = 'active'`, repoID,
	).Scan(&stats.ActiveWorktrees)
	if err != nil {
		return nil, err
	}

	err = db.conn.QueryRow(
		`SELECT COUNT(*) FROM tasks WHERE repo_id = ? AND status = 'pending'`, repoID,
	).Scan(&stats.PendingTasks)
	if err != nil {
		return nil, err
	}

	return stats, nil
}
