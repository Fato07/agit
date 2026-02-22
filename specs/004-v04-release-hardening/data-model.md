# Data Model: v0.4.0 Release Hardening

**Branch**: `004-v04-release-hardening` | **Date**: 2026-02-22

## Existing Entities (no schema changes)

The v0.4.0 release hardening does not introduce new database tables or modify the existing SQLite schema. All six user stories operate on the existing data model:

### Task (existing table: `tasks`)

| Field | Type | Description |
|-------|------|-------------|
| id | TEXT PK | UUID |
| repo_id | TEXT FK | Repository reference |
| description | TEXT | Task description |
| status | TEXT | pending / claimed / in_progress / completed / failed |
| priority | INTEGER | Higher = more important (default 0) |
| assigned_agent_id | TEXT | Agent that claimed the task |
| worktree_id | TEXT | Associated worktree (set on start) |
| result | TEXT | Completion result or failure reason |
| created_at | DATETIME | Creation timestamp |
| updated_at | DATETIME | Last modification timestamp |

**New behavior (US3)**: The `NextTask` dispatch method uses `priority DESC, created_at ASC` ordering to atomically select and claim the highest-priority pending task.

### Conflict (computed, not persisted)

| Field | Type | Description |
|-------|------|-------------|
| file_path | TEXT | Path of the conflicting file |
| worktree_ids | []TEXT | Worktrees that modified this file |
| agents | []TEXT | Agents associated with those worktrees |
| tasks | []TEXT | Tasks associated with those worktrees |

**New behavior (US4)**: Resolution suggestions are computed at query time based on worktree creation time and file overlap count. Not persisted.

## New Configuration Entity

### Hook (config-based, not DB)

Hooks are stored in `~/.agit/config.toml` under the `[hooks]` section, not in the database.

| Field | Type | Description |
|-------|------|-------------|
| event | string (TOML key) | Event name (e.g., `worktree.created`) |
| command | string (TOML value) | Shell command to execute |

**Supported events**: `worktree.created`, `worktree.removed`, `task.claimed`, `task.completed`, `task.failed`, `conflict.detected`

**Runtime context**: Hook commands receive event details via environment variables (`AGIT_EVENT`, `AGIT_REPO`, `AGIT_TASK_ID`, `AGIT_AGENT_ID`, `AGIT_WORKTREE_ID`).

## State Transitions

### Task Dispatch (US3 — new transition path)

```
pending ──[NextTask]──> claimed ──[StartTask]──> in_progress ──> completed/failed
```

The `NextTask` transition is atomic: a single SQL statement selects the highest-priority pending task and sets `status='claimed'` + `assigned_agent_id` in one operation.

### Hook Lifecycle (US5)

```
event fired ──> hook lookup ──> [no hook] ──> no-op
                              ──> [hook found] ──> spawn goroutine
                                                    ──> exec with timeout (30s default)
                                                    ──> log result (success/failure)
```

Hook execution is fire-and-forget. Failures are logged as warnings but never block the triggering operation.
