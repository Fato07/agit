# Data Model: agit Full Architecture Implementation

**Branch**: `001-agit-architecture-impl` | **Date**: 2026-02-20

## Existing Entities (Phase 1 — no changes)

### Repo
| Field | Type | Constraints |
|-------|------|-------------|
| id | string (UUID) | Primary key |
| name | string | Unique, derived from directory name |
| path | string | Absolute filesystem path |
| remote_url | string | Nullable, auto-detected from git |
| default_branch | string | Auto-detected (e.g., "main") |
| added_at | timestamp | Set on creation |
| last_synced | timestamp | Nullable |
| metadata | JSON | Nullable, extensible key-value |

### Worktree
| Field | Type | Constraints |
|-------|------|-------------|
| id | string (UUID) | Primary key |
| repo_id | string | FK → repos.id |
| path | string | Absolute filesystem path |
| branch | string | Git branch name |
| agent_id | string | Nullable, FK → agents.id |
| task_description | string | Nullable |
| status | enum | active / completed / stale / conflict |
| created_at | timestamp | Set on creation |
| updated_at | timestamp | Updated on status change |

### Agent
| Field | Type | Constraints |
|-------|------|-------------|
| id | string (UUID) | Primary key |
| name | string | Unique identifier |
| type | string | Agent type (e.g., "custom", "claude") |
| status | enum | active / idle / disconnected |
| current_worktree_id | string | Nullable, FK → worktrees.id |
| last_seen | timestamp | Updated on heartbeat |

### Task
| Field | Type | Constraints |
|-------|------|-------------|
| id | string | Primary key (t-XXXXXXXX format) |
| repo_id | string | FK → repos.id |
| description | string | Human-readable task description |
| status | enum | pending / claimed / in_progress / completed / failed |
| assigned_agent_id | string | Nullable, FK → agents.id |
| worktree_id | string | Nullable, FK → worktrees.id |
| created_at | timestamp | Set on creation |
| completed_at | timestamp | Nullable, set on complete/fail |
| result | string | Nullable, outcome description |

### FileTouch
| Field | Type | Constraints |
|-------|------|-------------|
| repo_id | string | PK part, FK → repos.id |
| worktree_id | string | PK part, FK → worktrees.id |
| file_path | string | PK part, relative to repo root |
| change_type | enum | added / modified / deleted / renamed |
| updated_at | timestamp | Last scan time |

## State Transitions

### Task Lifecycle
```
pending → claimed (via ClaimTask: assigns agent)
claimed → in_progress (via StartTask: assigns worktree)
in_progress → completed (via CompleteTask: sets result, completed_at)
in_progress → failed (via FailTask: sets result, completed_at)
claimed → failed (via FailTask: agent fails before starting)
```

On agent removal: claimed/in_progress → pending (soft release)

### Agent Lifecycle
```
(new) → active (via RegisterAgent)
active → active (via Heartbeat: updates last_seen)
active → disconnected (via SweepStaleAgents: last_seen exceeded stale_after)
disconnected → active (via Heartbeat: agent reconnects)
any → (deleted) (via RemoveAgent: soft release then delete)
```

### Worktree Lifecycle
```
(new) → active (via CreateWorktree)
active → completed (via merge or manual cleanup)
active → stale (via cleanup: orphaned worktrees)
active → conflict (via conflict detection flagging)
completed → (deleted from registry and disk via cleanup)
```

## New Registry Methods (to implement)

| Method | Purpose | Phase |
|--------|---------|-------|
| `SweepStaleAgents(staleAfter time.Duration) (int, error)` | Mark stale agents as disconnected | P5 gap |
| `RemoveAgent(name string) error` | Soft release + delete agent | P5 gap |
| `UnclaimAgentTasks(agentID string) error` | Revert agent's tasks to pending | P5 gap |
| `UnassignAgentWorktrees(agentID string) error` | Clear agent_id on worktrees | P5 gap |
