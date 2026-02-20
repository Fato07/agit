# MCP Resource Contracts

**Branch**: `001-agit-architecture-impl` | **Date**: 2026-02-20

Resources are read-only MCP endpoints that provide state snapshots. All responses are JSON text via `mcp.TextResourceContents`.

---

## agit://repos

**Type**: Static resource
**Description**: List all registered repositories with summary stats

**Response**:
```json
[
  {
    "name": "myapp",
    "path": "/home/user/myapp",
    "default_branch": "main",
    "remote_url": "git@github.com:user/myapp.git",
    "active_worktrees": 2,
    "pending_tasks": 1,
    "active_agents": 1
  }
]
```

---

## agit://repos/{name}

**Type**: Resource template (dynamic URI)
**Description**: Detailed status for a single repository

**URI Parameters**:
| Name | Description |
|------|-------------|
| name | Repository name |

**Response**:
```json
{
  "name": "myapp",
  "path": "/home/user/myapp",
  "default_branch": "main",
  "remote_url": "git@github.com:user/myapp.git",
  "worktrees": [
    {
      "id": "uuid-1234",
      "branch": "agit/fix-auth-abc123",
      "status": "active",
      "agent_name": "claude-1",
      "task": "fix authentication bug",
      "created_at": "2026-02-20T12:00:00Z"
    }
  ],
  "task_summary": {
    "pending": 1,
    "claimed": 0,
    "in_progress": 1,
    "completed": 3,
    "failed": 0
  },
  "active_agents": 1
}
```

---

## agit://repos/{name}/conflicts

**Type**: Resource template (dynamic URI)
**Description**: Current file conflicts for a repository

**URI Parameters**:
| Name | Description |
|------|-------------|
| name | Repository name |

**Response**:
```json
{
  "repo": "myapp",
  "conflicts": [
    {
      "file": "src/auth.go",
      "change_type": "modified",
      "worktrees": [
        {"id": "uuid-1234", "branch": "agit/fix-auth-abc123"},
        {"id": "uuid-5678", "branch": "agit/add-oauth-def456"}
      ]
    }
  ],
  "total_conflicts": 1,
  "scanned_at": "2026-02-20T12:30:00Z"
}
```

---

## agit://repos/{name}/tasks

**Type**: Resource template (dynamic URI)
**Description**: All tasks for a repository

**URI Parameters**:
| Name | Description |
|------|-------------|
| name | Repository name |

**Response**:
```json
{
  "repo": "myapp",
  "tasks": [
    {
      "id": "t-abc12345",
      "description": "fix authentication bug",
      "status": "in_progress",
      "assigned_agent": "claude-1",
      "worktree_branch": "agit/fix-auth-abc123",
      "created_at": "2026-02-20T12:00:00Z",
      "completed_at": null,
      "result": null
    }
  ]
}
```

---

## agit://agents

**Type**: Static resource
**Description**: All registered agents with their status

**Response**:
```json
[
  {
    "id": "uuid-agent-1",
    "name": "claude-1",
    "type": "claude",
    "status": "active",
    "last_seen": "2026-02-20T12:29:00Z",
    "current_worktree": {
      "id": "uuid-1234",
      "branch": "agit/fix-auth-abc123",
      "repo": "myapp"
    }
  },
  {
    "id": "uuid-agent-2",
    "name": "builder-old",
    "type": "custom",
    "status": "disconnected",
    "last_seen": "2026-02-20T11:00:00Z",
    "current_worktree": null
  }
]
```
