# MCP Tool Contracts

**Branch**: `001-agit-architecture-impl` | **Date**: 2026-02-20

All tools use the MCP tool protocol. Parameters are defined via JSON Schema. Responses are JSON text returned via `mcp.NewToolResultText()`.

---

## agit_list_repos

**Description**: List all registered repositories

**Parameters**: None

**Response**:
```json
[
  {
    "name": "myapp",
    "path": "/home/user/myapp",
    "default_branch": "main",
    "remote_url": "git@github.com:user/myapp.git",
    "worktree_count": 2,
    "task_count": 3
  }
]
```

---

## agit_repo_status

**Description**: Get detailed status for a specific repository

**Parameters**:
| Name | Type | Required | Description |
|------|------|----------|-------------|
| repo | string | yes | Repository name |

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
      "agent": "claude-1",
      "task": "fix authentication bug"
    }
  ],
  "tasks": [
    {
      "id": "t-abc12345",
      "description": "fix auth",
      "status": "in_progress",
      "agent": "claude-1"
    }
  ],
  "conflicts": []
}
```

---

## agit_spawn_worktree

**Description**: Create an isolated worktree for an agent

**Parameters**:
| Name | Type | Required | Description |
|------|------|----------|-------------|
| repo | string | yes | Repository name |
| task | string | no | Task description |
| branch | string | no | Custom branch name (auto-generated if omitted) |
| agent | string | no | Agent name to assign |

**Response**:
```json
{
  "worktree_id": "uuid-5678",
  "path": "/home/user/myapp/.worktrees/agit-abc123",
  "branch": "agit/fix-auth-abc123"
}
```

---

## agit_remove_worktree

**Description**: Remove a worktree from disk and registry

**Parameters**:
| Name | Type | Required | Description |
|------|------|----------|-------------|
| repo | string | yes | Repository name |
| worktree_id | string | yes | Worktree ID to remove |

**Response**:
```json
{
  "removed": true,
  "worktree_id": "uuid-5678"
}
```

---

## agit_check_conflicts

**Description**: Scan for file conflicts across active worktrees

**Parameters**:
| Name | Type | Required | Description |
|------|------|----------|-------------|
| repo | string | yes | Repository name |

**Response**:
```json
{
  "conflicts": [
    {
      "file": "src/auth.go",
      "worktrees": ["uuid-1234", "uuid-5678"]
    }
  ],
  "scanned_worktrees": 3
}
```

---

## agit_list_tasks

**Description**: List tasks for a repository

**Parameters**:
| Name | Type | Required | Description |
|------|------|----------|-------------|
| repo | string | yes | Repository name |
| status | string | no | Filter by status (pending/claimed/in_progress/completed/failed) |

**Response**:
```json
[
  {
    "id": "t-abc12345",
    "description": "fix authentication bug",
    "status": "pending",
    "agent": null,
    "created_at": "2026-02-20T12:00:00Z"
  }
]
```

---

## agit_claim_task

**Description**: Atomically claim a pending task for an agent

**Parameters**:
| Name | Type | Required | Description |
|------|------|----------|-------------|
| task_id | string | yes | Task ID to claim |
| agent_id | string | yes | Agent ID claiming the task |

**Response**:
```json
{
  "claimed": true,
  "task_id": "t-abc12345",
  "agent_id": "uuid-agent"
}
```

**Error** (task already claimed):
```json
{
  "error": "task \"t-abc12345\" not found or already claimed"
}
```

---

## agit_complete_task

**Description**: Mark a task as completed with optional result

**Parameters**:
| Name | Type | Required | Description |
|------|------|----------|-------------|
| task_id | string | yes | Task ID to complete |
| result | string | no | Result description |

**Response**:
```json
{
  "completed": true,
  "task_id": "t-abc12345"
}
```

---

## agit_merge_worktree

**Description**: Merge a worktree branch into the default branch, then auto-cleanup

**Parameters**:
| Name | Type | Required | Description |
|------|------|----------|-------------|
| repo | string | yes | Repository name |
| worktree_id | string | yes | Worktree ID to merge |

**Response**:
```json
{
  "merged": true,
  "branch": "agit/fix-auth-abc123",
  "into": "main",
  "worktree_cleaned": true
}
```

**Error** (merge conflict):
```json
{
  "error": "merge would result in conflicts",
  "conflicting_files": ["src/auth.go"]
}
```

---

## agit_register_agent

**Description**: Register a new AI agent

**Parameters**:
| Name | Type | Required | Description |
|------|------|----------|-------------|
| name | string | yes | Agent name |
| type | string | yes | Agent type (e.g., "claude", "custom") |

**Response**:
```json
{
  "agent_id": "uuid-agent-new",
  "name": "claude-1",
  "type": "claude"
}
```

---

## agit_heartbeat

**Description**: Update agent heartbeat timestamp

**Parameters**:
| Name | Type | Required | Description |
|------|------|----------|-------------|
| agent_id | string | yes | Agent ID |

**Response**:
```json
{
  "ok": true,
  "agent_id": "uuid-agent"
}
```
