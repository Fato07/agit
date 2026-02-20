package registry

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Task represents a work item for agents
type Task struct {
	ID              string
	RepoID          string
	Description     string
	Priority        int
	Status          string
	AssignedAgentID *string
	WorktreeID      *string
	CreatedAt       time.Time
	CompletedAt     *time.Time
	Result          *string
}

// CreateTask creates a new task
func (db *DB) CreateTask(repoID, description string, priority int) (*Task, error) {
	id := "t-" + uuid.New().String()[:8]
	now := time.Now()

	_, err := db.conn.Exec(
		`INSERT INTO tasks (id, repo_id, description, priority, status, created_at) VALUES (?, ?, ?, ?, 'pending', ?)`,
		id, repoID, description, priority, now,
	)
	if err != nil {
		return nil, fmt.Errorf("could not create task: %w", err)
	}

	return &Task{
		ID:          id,
		RepoID:      repoID,
		Description: description,
		Priority:    priority,
		Status:      "pending",
		CreatedAt:   now,
	}, nil
}

// ClaimTask assigns a task to an agent
func (db *DB) ClaimTask(taskID, agentID string) error {
	result, err := db.conn.Exec(
		`UPDATE tasks SET status = 'claimed', assigned_agent_id = ? WHERE id = ? AND status = 'pending'`,
		agentID, taskID,
	)
	if err != nil {
		return fmt.Errorf("could not claim task: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("task %q not found or already claimed", taskID)
	}
	return nil
}

// StartTask marks a task as in_progress and associates a worktree
func (db *DB) StartTask(taskID, worktreeID string) error {
	_, err := db.conn.Exec(
		`UPDATE tasks SET status = 'in_progress', worktree_id = ? WHERE id = ?`,
		worktreeID, taskID,
	)
	return err
}

// CompleteTask marks a task as completed
func (db *DB) CompleteTask(taskID string, result *string) error {
	now := time.Now()
	res, err := db.conn.Exec(
		`UPDATE tasks SET status = 'completed', completed_at = ?, result = ? WHERE id = ?`,
		now, result, taskID,
	)
	if err != nil {
		return fmt.Errorf("could not complete task: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("task %q not found", taskID)
	}
	return nil
}

// FailTask marks a task as failed
func (db *DB) FailTask(taskID string, result *string) error {
	now := time.Now()
	_, err := db.conn.Exec(
		`UPDATE tasks SET status = 'failed', completed_at = ?, result = ? WHERE id = ?`,
		now, result, taskID,
	)
	return err
}

// GetTask retrieves a task by ID
func (db *DB) GetTask(id string) (*Task, error) {
	t := &Task{}
	err := db.conn.QueryRow(
		`SELECT id, repo_id, description, priority, status, assigned_agent_id, worktree_id, created_at, completed_at, result
		 FROM tasks WHERE id = ?`, id,
	).Scan(&t.ID, &t.RepoID, &t.Description, &t.Priority, &t.Status, &t.AssignedAgentID,
		&t.WorktreeID, &t.CreatedAt, &t.CompletedAt, &t.Result)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("task %q not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("could not get task: %w", err)
	}
	return t, nil
}

// ListTasks returns tasks for a repo, optionally filtered by status
func (db *DB) ListTasks(repoID string, status *string) ([]*Task, error) {
	var rows *sql.Rows
	var err error

	if status != nil {
		rows, err = db.conn.Query(
			`SELECT id, repo_id, description, priority, status, assigned_agent_id, worktree_id, created_at, completed_at, result
			 FROM tasks WHERE repo_id = ? AND status = ? ORDER BY priority DESC, created_at DESC`,
			repoID, *status,
		)
	} else {
		rows, err = db.conn.Query(
			`SELECT id, repo_id, description, priority, status, assigned_agent_id, worktree_id, created_at, completed_at, result
			 FROM tasks WHERE repo_id = ? ORDER BY priority DESC, created_at DESC`,
			repoID,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("could not list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		t := &Task{}
		if err := rows.Scan(&t.ID, &t.RepoID, &t.Description, &t.Priority, &t.Status, &t.AssignedAgentID,
			&t.WorktreeID, &t.CreatedAt, &t.CompletedAt, &t.Result); err != nil {
			return nil, fmt.Errorf("could not scan task: %w", err)
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}
