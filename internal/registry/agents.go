package registry

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Agent represents a registered AI agent
type Agent struct {
	ID                string
	Name              string
	Type              string
	Status            string
	CurrentWorktreeID *string
	LastSeen          time.Time
}

// RegisterAgent creates a new agent record
func (db *DB) RegisterAgent(name, agentType string) (*Agent, error) {
	id := uuid.New().String()
	now := time.Now()

	_, err := db.conn.Exec(
		`INSERT INTO agents (id, name, type, status, last_seen) VALUES (?, ?, ?, 'active', ?)`,
		id, name, agentType, now,
	)
	if err != nil {
		return nil, fmt.Errorf("could not register agent: %w", err)
	}

	return &Agent{
		ID:       id,
		Name:     name,
		Type:     agentType,
		Status:   "active",
		LastSeen: now,
	}, nil
}

// GetAgent retrieves an agent by ID
func (db *DB) GetAgent(id string) (*Agent, error) {
	agent := &Agent{}
	err := db.conn.QueryRow(
		`SELECT id, name, type, status, current_worktree_id, last_seen FROM agents WHERE id = ?`, id,
	).Scan(&agent.ID, &agent.Name, &agent.Type, &agent.Status, &agent.CurrentWorktreeID, &agent.LastSeen)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("agent %q not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("could not get agent: %w", err)
	}
	return agent, nil
}

// GetAgentByName retrieves an agent by name
func (db *DB) GetAgentByName(name string) (*Agent, error) {
	agent := &Agent{}
	err := db.conn.QueryRow(
		`SELECT id, name, type, status, current_worktree_id, last_seen FROM agents WHERE name = ?`, name,
	).Scan(&agent.ID, &agent.Name, &agent.Type, &agent.Status, &agent.CurrentWorktreeID, &agent.LastSeen)

	if err == sql.ErrNoRows {
		return nil, nil // not found is not an error for lookup
	}
	if err != nil {
		return nil, fmt.Errorf("could not get agent: %w", err)
	}
	return agent, nil
}

// ListAgents returns all agents
func (db *DB) ListAgents() ([]*Agent, error) {
	rows, err := db.conn.Query(
		`SELECT id, name, type, status, current_worktree_id, last_seen FROM agents ORDER BY name`,
	)
	if err != nil {
		return nil, fmt.Errorf("could not list agents: %w", err)
	}
	defer rows.Close()

	var agents []*Agent
	for rows.Next() {
		a := &Agent{}
		if err := rows.Scan(&a.ID, &a.Name, &a.Type, &a.Status, &a.CurrentWorktreeID, &a.LastSeen); err != nil {
			return nil, fmt.Errorf("could not scan agent: %w", err)
		}
		agents = append(agents, a)
	}
	return agents, nil
}

// Heartbeat updates an agent's last_seen timestamp
func (db *DB) Heartbeat(agentID string) error {
	result, err := db.conn.Exec(
		`UPDATE agents SET last_seen = ?, status = 'active' WHERE id = ?`,
		time.Now(), agentID,
	)
	if err != nil {
		return fmt.Errorf("could not update heartbeat: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("agent %q not found", agentID)
	}
	return nil
}

// UpdateAgentWorktree sets the current worktree for an agent
func (db *DB) UpdateAgentWorktree(agentID string, worktreeID *string) error {
	_, err := db.conn.Exec(
		`UPDATE agents SET current_worktree_id = ?, last_seen = ? WHERE id = ?`,
		worktreeID, time.Now(), agentID,
	)
	return err
}

// SweepStaleAgents marks agents as disconnected if their last_seen exceeds staleAfter
func (db *DB) SweepStaleAgents(staleAfter time.Duration) (int, error) {
	cutoff := time.Now().Add(-staleAfter)
	result, err := db.conn.Exec(
		`UPDATE agents SET status = 'disconnected' WHERE status = 'active' AND last_seen < ?`,
		cutoff,
	)
	if err != nil {
		return 0, fmt.Errorf("could not sweep stale agents: %w", err)
	}
	rows, _ := result.RowsAffected()
	return int(rows), nil
}

// UnclaimAgentTasks reverts an agent's claimed/in_progress tasks to pending
func (db *DB) UnclaimAgentTasks(agentID string) error {
	_, err := db.conn.Exec(
		`UPDATE tasks SET status = 'pending', assigned_agent_id = NULL
		 WHERE assigned_agent_id = ? AND status IN ('claimed', 'in_progress')`,
		agentID,
	)
	if err != nil {
		return fmt.Errorf("could not unclaim agent tasks: %w", err)
	}
	return nil
}

// UnassignAgentWorktrees clears agent_id on worktrees assigned to this agent
func (db *DB) UnassignAgentWorktrees(agentID string) error {
	_, err := db.conn.Exec(
		`UPDATE worktrees SET agent_id = NULL WHERE agent_id = ?`,
		agentID,
	)
	if err != nil {
		return fmt.Errorf("could not unassign agent worktrees: %w", err)
	}
	return nil
}

// RemoveAgent performs a soft release (unclaim tasks, unassign worktrees) then deletes the agent
func (db *DB) RemoveAgent(name string) error {
	agent, err := db.GetAgentByName(name)
	if err != nil {
		return err
	}
	if agent == nil {
		return fmt.Errorf("agent %q not found", name)
	}

	tx, err := db.conn.Begin()
	if err != nil {
		return fmt.Errorf("could not begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Unclaim tasks
	if _, err := tx.Exec(
		`UPDATE tasks SET status = 'pending', assigned_agent_id = NULL
		 WHERE assigned_agent_id = ? AND status IN ('claimed', 'in_progress')`,
		agent.ID,
	); err != nil {
		return fmt.Errorf("could not unclaim tasks: %w", err)
	}

	// Unassign worktrees
	if _, err := tx.Exec(
		`UPDATE worktrees SET agent_id = NULL WHERE agent_id = ?`,
		agent.ID,
	); err != nil {
		return fmt.Errorf("could not unassign worktrees: %w", err)
	}

	// Delete agent
	if _, err := tx.Exec(`DELETE FROM agents WHERE id = ?`, agent.ID); err != nil {
		return fmt.Errorf("could not delete agent: %w", err)
	}

	return tx.Commit()
}
