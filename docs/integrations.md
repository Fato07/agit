# agit MCP Integration Guide

agit exposes its functionality via the [Model Context Protocol (MCP)](https://modelcontextprotocol.io/). Any MCP-compatible AI agent can use agit to discover repositories, manage worktrees, coordinate tasks, and detect conflicts.

## Quick Setup

agit uses **stdio transport** by default (simplest, recommended for local use). Add the following to your agent's MCP configuration:

### Claude Code

```json
// ~/.claude/mcp.json
{
  "mcpServers": {
    "agit": {
      "command": "agit",
      "args": ["serve"]
    }
  }
}
```

### OpenClaw

```json
// openclaw.json or Gateway settings
{
  "mcpServers": {
    "agit": {
      "command": "agit",
      "args": ["serve"],
      "env": {}
    }
  }
}
```

### Cursor

```json
// .cursor/mcp.json
{
  "mcpServers": {
    "agit": {
      "command": "agit",
      "args": ["serve"]
    }
  }
}
```

### Any MCP-Compatible Agent

agit implements the standard MCP protocol. Point your agent's MCP config to `agit serve`:

```json
{
  "mcpServers": {
    "agit": {
      "command": "agit",
      "args": ["serve"]
    }
  }
}
```

## SSE Transport (Remote Deployments)

For scenarios where agit runs on a different machine than the agent, use SSE transport:

```json
{
  "mcpServers": {
    "agit": {
      "command": "agit",
      "args": ["serve", "--transport", "sse", "--port", "3847"],
      "env": {}
    }
  }
}
```

## Available Tools and Resources

agit exposes **20 MCP tools** and **5 MCP resources**.

**Tools** (actions):
- `agit_list_repos` - List registered repositories
- `agit_add_repo` - Register a Git repository via MCP
- `agit_repo_status` - Get detailed status for a specific repository
- `agit_spawn_worktree` - Create an isolated worktree for an agent
- `agit_remove_worktree` - Remove a worktree from disk and registry
- `agit_list_worktrees` - List worktrees for a repository
- `agit_merge_worktree` - Merge a worktree branch into the default branch
- `agit_check_conflicts` - Detect overlapping changes across active worktrees (includes resolution suggestions)
- `agit_list_tasks` - List tasks for a repository
- `agit_create_task` - Create a new task for a repository
- `agit_get_task` - Get detailed information about a specific task
- `agit_claim_task` - Atomically claim a pending task for an agent
- `agit_next_task` - Atomically claim the highest-priority pending task
- `agit_start_task` - Mark a claimed task as in-progress with a worktree
- `agit_complete_task` - Mark a task as completed with optional result
- `agit_fail_task` - Mark a task as failed with optional reason
- `agit_register_agent` - Register a new AI agent
- `agit_list_agents` - List all registered AI agents
- `agit_heartbeat` - Update agent heartbeat timestamp
- `agit_cleanup_worktrees` - Prune orphaned worktrees

**Resources** (read-only state):
- `agit://repos` - All registered repositories
- `agit://repos/{name}` - Single repository details
- `agit://repos/{name}/conflicts` - Conflict report
- `agit://repos/{name}/tasks` - Tasks for a repo
- `agit://agents` - All registered agents

## Verification

After configuring agit, verify the integration:

1. Start a new agent session
2. The agent should be able to call `agit_list_repos`
3. If repos were previously registered, they should appear in the response
4. Try `agit_spawn_worktree` to create an isolated workspace
5. Verify the agent can work in the spawned worktree path
6. Call `agit_check_conflicts` — should return empty with only one worktree
7. Call `agit_merge_worktree` to clean up

## Prerequisites

agit must be initialized before MCP use:

```bash
agit init
agit add /path/to/your/repo --name my-repo
```
