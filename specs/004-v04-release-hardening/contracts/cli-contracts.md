# CLI Contracts: v0.4.0 Release Hardening

## New CLI Commands

### `agit tasks next <repo>`

**Purpose**: Atomically claim the highest-priority pending task for a repository.

**Arguments**:
| Arg | Required | Description |
|-----|----------|-------------|
| `repo` | Yes | Repository name or ID |

**Flags**:
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--agent` | string | required | Agent ID claiming the task |
| `--output` | string | `text` | Output format: `text` or `json` |

**Success Response (text)**:
```
Claimed task <task-id>: <description> (priority: <n>)
```

**Success Response (JSON)**:
```json
{
  "id": "uuid",
  "description": "task description",
  "priority": 10,
  "status": "claimed",
  "assigned_agent_id": "agent-1",
  "repo_id": "repo-uuid"
}
```

**No Tasks Response (text)**:
```
No pending tasks for repository <repo>
```

**No Tasks Response (JSON)**:
```json
{
  "message": "no pending tasks"
}
```

**Error Cases**:
- Repository not found → exit code 1, user-friendly error
- Missing `--agent` flag → exit code 1, usage help

---

## Modified CLI Output

### `agit conflicts <repo>` (enhanced)

**Current behavior**: Lists conflicting files and involved worktrees.

**New behavior**: Appends resolution suggestions after the conflict list.

**Enhanced text output**:
```
Conflicts in <repo>:

  src/main.go
    ├─ worktree-abc (agent: claude-1, task: refactor auth)
    └─ worktree-def (agent: cursor-1, task: add tests)

  internal/config.go
    ├─ worktree-abc (agent: claude-1, task: refactor auth)
    └─ worktree-ghi (agent: claude-2, task: update config)

Suggested resolution order:
  1. Merge worktree-def first (1 conflicting file, created 2h ago)
  2. Merge worktree-ghi next (1 conflicting file, created 1h ago)
  3. Merge worktree-abc last (2 conflicting files, created 3h ago)

Rationale: Merging worktrees with fewer conflicts first minimizes rebase
complexity for remaining worktrees.
```

**Enhanced JSON output**:
```json
{
  "conflicts": [...],
  "suggestions": [
    {
      "order": 1,
      "worktree_id": "worktree-def",
      "conflicting_files": 1,
      "created_at": "2026-02-22T10:00:00Z",
      "rationale": "Fewest conflicting files"
    }
  ]
}
```

---

## Hook Configuration Contract

### Config format (`~/.agit/config.toml`)

```toml
[hooks]
"worktree.created" = "notify-slack --channel dev"
"worktree.removed" = ""
"task.claimed" = "/usr/local/bin/log-task"
"task.completed" = ""
"task.failed" = ""
"conflict.detected" = ""

[hooks.timeout]
default = "30s"
```

### Hook Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `AGIT_EVENT` | Event name | `worktree.created` |
| `AGIT_REPO` | Repository name | `my-app` |
| `AGIT_REPO_ID` | Repository UUID | `uuid` |
| `AGIT_TASK_ID` | Task UUID (if applicable) | `uuid` |
| `AGIT_AGENT_ID` | Agent ID (if applicable) | `claude-1` |
| `AGIT_WORKTREE_ID` | Worktree UUID (if applicable) | `uuid` |
| `AGIT_WORKTREE_PATH` | Worktree path (if applicable) | `/path/to/worktree` |
