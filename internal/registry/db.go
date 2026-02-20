package registry

import (
	"database/sql"
	"fmt"
	"strings"

	_ "modernc.org/sqlite"

	"github.com/fathindos/agit/internal/config"
)

// DB wraps the SQLite database connection
type DB struct {
	conn *sql.DB
}

// Open opens (or creates) the agit database
func Open() (*DB, error) {
	path, err := config.DBPath()
	if err != nil {
		return nil, err
	}

	conn, err := sql.Open("sqlite", path+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("could not open database: %w", err)
	}

	// Enable WAL mode and foreign keys
	if _, err := conn.Exec("PRAGMA journal_mode=WAL"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("could not set WAL mode: %w", err)
	}
	if _, err := conn.Exec("PRAGMA foreign_keys=ON"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("could not enable foreign keys: %w", err)
	}

	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("could not run migrations: %w", err)
	}

	return db, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// Conn returns the underlying sql.DB for direct queries
func (db *DB) Conn() *sql.DB {
	return db.conn
}

// migrate creates or updates the database schema
func (db *DB) migrate() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS repos (
			id TEXT PRIMARY KEY,
			name TEXT UNIQUE NOT NULL,
			path TEXT NOT NULL,
			remote_url TEXT NOT NULL DEFAULT '',
			default_branch TEXT NOT NULL DEFAULT 'main',
			added_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			last_synced TIMESTAMP,
			metadata JSON DEFAULT '{}'
		)`,

		`CREATE TABLE IF NOT EXISTS worktrees (
			id TEXT PRIMARY KEY,
			repo_id TEXT NOT NULL REFERENCES repos(id) ON DELETE CASCADE,
			path TEXT NOT NULL,
			branch TEXT NOT NULL,
			agent_id TEXT,
			task_description TEXT,
			status TEXT NOT NULL DEFAULT 'active' CHECK(status IN ('active', 'completed', 'stale', 'conflict')),
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS agents (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			type TEXT NOT NULL DEFAULT 'custom',
			status TEXT NOT NULL DEFAULT 'active' CHECK(status IN ('active', 'idle', 'disconnected')),
			current_worktree_id TEXT REFERENCES worktrees(id) ON DELETE SET NULL,
			last_seen TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS tasks (
			id TEXT PRIMARY KEY,
			repo_id TEXT NOT NULL REFERENCES repos(id) ON DELETE CASCADE,
			description TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'claimed', 'in_progress', 'completed', 'failed')),
			assigned_agent_id TEXT REFERENCES agents(id) ON DELETE SET NULL,
			worktree_id TEXT REFERENCES worktrees(id) ON DELETE SET NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			completed_at TIMESTAMP,
			result TEXT
		)`,

		`CREATE TABLE IF NOT EXISTS file_touches (
			repo_id TEXT NOT NULL REFERENCES repos(id) ON DELETE CASCADE,
			worktree_id TEXT NOT NULL REFERENCES worktrees(id) ON DELETE CASCADE,
			file_path TEXT NOT NULL,
			change_type TEXT NOT NULL DEFAULT 'modified' CHECK(change_type IN ('added', 'modified', 'deleted', 'renamed')),
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (repo_id, worktree_id, file_path)
		)`,

		// Indexes for common queries
		`CREATE INDEX IF NOT EXISTS idx_worktrees_repo_id ON worktrees(repo_id)`,
		`CREATE INDEX IF NOT EXISTS idx_worktrees_status ON worktrees(status)`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_repo_id ON tasks(repo_id)`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status)`,
		`CREATE INDEX IF NOT EXISTS idx_file_touches_repo_worktree ON file_touches(repo_id, worktree_id)`,
	}

	for _, m := range migrations {
		if _, err := db.conn.Exec(m); err != nil {
			return fmt.Errorf("migration failed: %s: %w", m[:60], err)
		}
	}

	// Additive schema migrations (ignore "duplicate column" errors)
	alterMigrations := []string{
		`ALTER TABLE tasks ADD COLUMN priority INTEGER NOT NULL DEFAULT 0`,
	}
	for _, m := range alterMigrations {
		if _, err := db.conn.Exec(m); err != nil {
			if !strings.Contains(err.Error(), "duplicate column") {
				return fmt.Errorf("migration failed: %w", err)
			}
		}
	}

	return nil
}
