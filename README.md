# agit

[![CI](https://github.com/Fato07/agit/actions/workflows/ci.yml/badge.svg)](https://github.com/Fato07/agit/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/Fato07/agit)](https://github.com/Fato07/agit/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/fathindos/agit)](https://goreportcard.com/report/github.com/fathindos/agit)
[![Go Version](https://img.shields.io/github/go-mod/go-version/Fato07/agit)](https://github.com/Fato07/agit/blob/main/go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Infrastructure-aware Git orchestration for AI agents.

**agit** gives AI coding agents persistent, queryable awareness of their Git infrastructure. It manages a local registry of repositories, orchestrates Git worktrees for agent isolation, detects cross-worktree conflicts, and coordinates task assignment across multiple agents.

## The Problem

AI coding agents (Claude Code, Cursor, Codex, OpenClaw) operate without persistent memory of their infrastructure. Each session starts blind — no knowledge of available repos, active work, or what other agents are doing. When multiple agents work on the same repo, conflicts are discovered only at merge time.

## The Solution

agit provides:

- **Persistent registry** — agents always know what repos are available via MCP
- **Worktree isolation** — each agent gets its own workspace, no stepping on toes
- **Conflict detection** — know about overlapping changes before merge time
- **Task coordination** — agents claim work, preventing duplication

## Quick Start

```bash
# Install
go install github.com/fathindos/agit@latest

# Initialize
agit init

# Register your repos
agit add ~/projects/my-app --name my-app
agit add ~/projects/api --name api

# Spawn isolated worktrees for agents
agit spawn my-app --task "refactor auth" --agent claude-1
agit spawn my-app --task "add tests" --agent cursor-1

# Check for conflicts
agit conflicts my-app

# See everything at a glance
agit status

# Merge completed work back
agit merge <worktree-id> --cleanup
```

## MCP Integration

Add agit to your agent's MCP config for automatic infrastructure awareness:

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

Now any MCP-compatible agent can call `agit_list_repos()` on startup and immediately know what's available.

## Commands

| Command | Description |
|---------|-------------|
| `agit init` | Initialize agit (~/.agit/) |
| `agit add <path>` | Register a Git repository |
| `agit repos` | List registered repositories |
| `agit spawn <repo>` | Create isolated worktree for an agent |
| `agit status [repo]` | Show worktrees, agents, conflicts |
| `agit conflicts [repo]` | Check for overlapping file changes |
| `agit tasks <repo>` | Manage tasks (create/claim/complete) |
| `agit merge <id>` | Merge worktree back to base branch |
| `agit cleanup` | Remove completed/stale worktrees |
| `agit serve` | Start MCP server |

## Architecture

See [docs/integrations.md](docs/integrations.md) for MCP integration details.

## License

MIT
