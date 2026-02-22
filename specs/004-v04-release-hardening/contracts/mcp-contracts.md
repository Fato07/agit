# MCP Tool Contracts: v0.4.0 Release Hardening

## New MCP Tool

### `agit_next_task`

**Purpose**: Atomically dispatch the highest-priority pending task to an agent.

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "repo": {
      "type": "string",
      "description": "Repository name or ID"
    },
    "agent_id": {
      "type": "string",
      "description": "ID of the agent requesting work"
    }
  },
  "required": ["repo", "agent_id"]
}
```

**Success Response**:
```json
{
  "content": [
    {
      "type": "text",
      "text": "{\"id\":\"uuid\",\"description\":\"task desc\",\"priority\":10,\"status\":\"claimed\",\"assigned_agent_id\":\"agent-1\"}"
    }
  ]
}
```

**No Tasks Response**:
```json
{
  "content": [
    {
      "type": "text",
      "text": "{\"message\":\"no pending tasks for repository\"}"
    }
  ]
}
```

**Error Response**:
```json
{
  "isError": true,
  "content": [
    {
      "type": "text",
      "text": "repository not found: myrepo"
    }
  ]
}
```

---

## Enhanced MCP Tool

### `agit_check_conflicts` (enhanced response)

**Existing behavior**: Returns conflict list as JSON text.

**New behavior**: Response JSON includes `suggestions` array alongside `conflicts`.

**Enhanced response text field**:
```json
{
  "conflicts": [
    {
      "file": "src/main.go",
      "worktrees": ["wt-abc", "wt-def"],
      "agents": ["claude-1", "cursor-1"],
      "tasks": ["refactor auth", "add tests"]
    }
  ],
  "suggestions": [
    {
      "order": 1,
      "worktree_id": "wt-def",
      "conflicting_files": 1,
      "rationale": "Fewest conflicting files, merge first to minimize rebase complexity"
    }
  ]
}
```

---

## SSE Transport Contract

### Graceful Shutdown

**Signal**: SIGTERM or SIGINT

**Behavior**:
1. Stop accepting new SSE connections
2. Allow in-flight requests to complete (up to 5-second timeout)
3. Close all active SSE connections
4. Exit with code 0

**Health Check** (future consideration, not in v0.4.0 scope):
- Not implemented in v0.4.0
- SSE server responds to standard MCP protocol messages
