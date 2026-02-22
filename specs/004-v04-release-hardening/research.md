# Research: v0.4.0 Release Hardening

**Date**: 2026-02-22
**Branch**: `004-v04-release-hardening`

## R1: CLI Integration Test Strategy

**Decision**: Use Cobra's `Execute` programmatically with captured stdout/stderr and temp-dir databases. Do not shell out to a compiled binary.

**Rationale**: The existing test infrastructure uses `registry.OpenMemory()` for in-memory SQLite. For CLI tests, we can set up a temp `$HOME` with `t.TempDir()` + `t.Setenv("HOME", ...)`, call `rootCmd.SetArgs(...)` and `rootCmd.Execute()`, and capture output via `rootCmd.SetOut(buf)` / `rootCmd.SetErr(buf)`. This avoids the overhead of compiling a binary per test and keeps tests fast.

**Alternatives considered**:
- `exec.Command("go", "run", ".")` — slow, requires compilation per test
- `TestMain` with binary build — complex setup, fragile CI

## R2: Task Dispatch Atomicity

**Decision**: Add `NextTask(repoID, agentID string) (*Task, error)` method to `registry.DB` using a single SQL statement: `UPDATE tasks SET status='claimed', assigned_agent_id=? WHERE id = (SELECT id FROM tasks WHERE repo_id=? AND status='pending' ORDER BY priority DESC, created_at ASC LIMIT 1) RETURNING *`.

**Rationale**: SQLite's implicit write lock ensures atomicity for single-connection usage. The `UPDATE ... WHERE id = (SELECT ...)` pattern atomically selects and claims in one statement. Using `created_at ASC` as tiebreaker ensures FIFO among equal priority.

**Alternatives considered**:
- Two-step SELECT then UPDATE — race condition between agents
- Application-level mutex — unnecessary given SQLite's locking

**Note**: SQLite with modernc.org driver may not support `RETURNING *`. Fallback: run the UPDATE, then SELECT the claimed task using `assigned_agent_id` + `status='claimed'` + `ORDER BY created_at DESC LIMIT 1`.

## R3: Conflict Resolution Suggestion Algorithm

**Decision**: Suggest merge order based on: (1) fewer conflicting files first (smaller impact), (2) older worktree first (more likely to be closer to completion). Output a numbered list of suggested merge order with rationale.

**Rationale**: Merging the worktree with fewer conflicts first minimizes rebase complexity for the remaining worktrees. Creation time serves as a proxy for "likely closer to done."

**Alternatives considered**:
- Random order — no useful guidance
- Most conflicting files first — increases rebase burden for later worktrees
- Agent priority — no priority concept for agents currently

## R4: Hook System Config Format

**Decision**: Add a `[hooks]` table in `config.toml` with event names as keys and command strings as values. Single command per event for v0.4.0 (simplicity).

```toml
[hooks]
"worktree.created" = "notify-slack --channel dev"
"task.claimed" = "/usr/local/bin/log-task"
"task.completed" = ""
```

**Rationale**: Simple key-value is sufficient for v0.4.0. Array-of-commands-per-event adds complexity with minimal benefit for the initial release. Users can chain commands via shell (`cmd1 && cmd2`).

**Alternatives considered**:
- Array per event (`["cmd1", "cmd2"]`) — deferred to future version
- Separate hooks config file — unnecessary indirection
- Webhook URLs — too opinionated; users can curl from shell commands

## R5: SSE Graceful Shutdown

**Decision**: Use `os/signal` to catch SIGTERM/SIGINT, call `sseServer.Shutdown(ctx)` (if the mcp-go library supports it) or `http.Server.Shutdown()` on the underlying HTTP server. Use a 5-second shutdown timeout.

**Rationale**: The `mcp-go` SSE server wraps an `http.Server`. The standard pattern is signal notification → context with timeout → server shutdown.

**Alternatives considered**:
- Let OS kill the process — no cleanup, potential resource leaks
- Longer timeout — 5s is standard for CLI tools

## R6: cmd/ Test Pattern — Tasks Subcommand Migration

**Decision**: Keep `tasks` as a flag-based command for backward compatibility but add `tasks next` as a subcommand. Cobra supports both: the parent command handles flags, subcommands extend it.

**Rationale**: Existing users and scripts may use `agit tasks myrepo --create "desc"`. Breaking this would violate backward compatibility. Adding `agit tasks next myrepo` follows the `config` subcommand pattern already established.

**Alternatives considered**:
- Full migration to subcommands (`tasks create`, `tasks list`, etc.) — breaking change, deferred
- Separate top-level command `agit next-task` — inconsistent naming
