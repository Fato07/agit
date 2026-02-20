---
name: agit
description: Git infrastructure awareness and multi-agent coordination
version: 0.1.0
mcp_servers:
  - agit
---

# agit - Git Infrastructure for AI Agents

You have access to agit, which provides persistent awareness of Git repositories
and coordinates work across multiple agents.

## When to use agit

- **Starting a new session**: Always call `agit_list_repos` first to see what
  repositories are available. Never assume you know what repos exist.
- **Before making changes**: Call `agit_spawn_worktree` to get an isolated
  workspace. Never work directly on the main branch.
- **Before merging**: Call `agit_check_conflicts` to see if your changes
  overlap with other agents' work.
- **When given a task**: Call `agit_list_tasks` to see if there are existing
  tasks, and `agit_claim_task` before starting work.

## Workflow

1. `agit_register_agent` - Announce yourself
2. `agit_list_repos` - Discover available repos
3. `agit_list_tasks` - Check for existing work
4. `agit_claim_task` or `agit_spawn_worktree` - Start working
5. `agit_check_conflicts` - Check periodically while working
6. `agit_complete_task` - Mark work as done
7. `agit_merge_worktree` - Merge when ready

## Important

- Always use worktrees. Never commit directly to the default branch.
- Check for conflicts before merging.
- Complete tasks when done so other agents know the work is finished.
