# Implementation Plan: agit Full Architecture Implementation

**Branch**: `001-agit-architecture-impl` | **Date**: 2026-02-20 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/001-agit-architecture-impl/spec.md`

## Summary

Implement Phase 7 (MCP Server) — the only entirely unimplemented phase — plus small gap fixes for Phases 4, 5, and 6. The MCP server exposes 11 tools and 5 resources via `github.com/mark3labs/mcp-go`, enabling AI agents to programmatically discover repos, spawn worktrees, manage tasks, and coordinate work over stdio or SSE transport. Gap fixes add `--fail`/`--result` task flags, `agit agents` CLI subcommand, and live conflict scanning in status.

## Technical Context

**Language/Version**: Go 1.22
**Primary Dependencies**: `github.com/mark3labs/mcp-go` v0.20.1 (MCP protocol), `github.com/spf13/cobra` v1.8.1 (CLI), `modernc.org/sqlite` v1.34.4 (database)
**Storage**: SQLite at `~/.agit/agit.db`, WAL mode, foreign keys enabled
**Testing**: `go test ./...`, manual MCP client testing via JSON-RPC over stdio
**Target Platform**: macOS/Linux CLI tool
**Project Type**: Single Go module CLI application
**Performance Goals**: MCP tool responses < 200ms for local operations
**Constraints**: Single-machine tool, SSE binds to localhost only, no authentication
**Scale/Scope**: Single developer machine, typically 1-10 repos, 1-20 worktrees

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Constitution file contains template placeholders only (not yet filled). No gates to enforce. **PASS** — no violations.

## Project Structure

### Documentation (this feature)

```text
specs/001-agit-architecture-impl/
├── plan.md              # This file
├── research.md          # Phase 0 output — research decisions
├── data-model.md        # Phase 1 output — entity model
├── quickstart.md        # Phase 1 output — developer quickstart
├── contracts/
│   ├── mcp-tools.md     # Phase 1 output — 11 tool contracts
│   └── mcp-resources.md # Phase 1 output — 5 resource contracts
├── checklists/
│   └── requirements.md  # Quality checklist (39/39 pass)
└── tasks.md             # Phase 2 output (created by /speckit.tasks)
```

### Source Code (repository root)

```text
cmd/
├── add.go           # Existing — agit add
├── agents.go        # NEW — agit agents subcommand (Phase 5 gap)
├── cleanup.go       # Existing — agit cleanup
├── conflicts.go     # Existing — agit conflicts
├── init.go          # Existing — agit init
├── merge.go         # Existing — agit merge
├── repos.go         # Existing — agit repos
├── root.go          # Existing — root command
├── serve.go         # MODIFY — wire MCP server (Phase 7)
├── spawn.go         # Existing — agit spawn
├── status.go        # MODIFY — add live conflict scan (Phase 6 gap)
└── tasks.go         # MODIFY — add --fail, --result flags (Phase 4 gap)

internal/
├── config/
│   └── config.go    # Existing — no changes
├── conflicts/
│   ├── detector.go  # Existing — no changes
│   └── report.go    # Existing — no changes
├── git/
│   ├── diff.go      # Existing — no changes
│   ├── merge.go     # Existing — no changes
│   ├── repo.go      # Existing — no changes
│   └── worktree.go  # Existing — no changes
├── mcp/             # NEW — MCP server package (Phase 7)
│   ├── server.go    # Server creation, tool/resource registration
│   ├── tools.go     # 11 tool handler implementations
│   └── resources.go # 5 resource handler implementations
└── registry/
    ├── agents.go    # MODIFY — add SweepStaleAgents, RemoveAgent (Phase 5 gap)
    ├── db.go        # Existing — no changes
    ├── repos.go     # Existing — no changes
    ├── tasks.go     # Existing — no changes
    ├── touches.go   # Existing — no changes
    └── worktrees.go # Existing — no changes
```

**Structure Decision**: Existing single-module Go CLI structure is maintained. New code goes into `internal/mcp/` (3 files) and `cmd/agents.go` (1 file). Three existing files are modified with small additions.

## Implementation Phases

### Phase A: Registry Gap Methods (Foundation)

Add new methods to `internal/registry/agents.go`:

1. `SweepStaleAgents(staleAfter time.Duration) (int, error)` — UPDATE agents SET status='disconnected' WHERE status='active' AND last_seen < now - staleAfter
2. `UnclaimAgentTasks(agentID string) error` — UPDATE tasks SET status='pending', assigned_agent_id=NULL WHERE assigned_agent_id=? AND status IN ('claimed','in_progress')
3. `UnassignAgentWorktrees(agentID string) error` — UPDATE worktrees SET agent_id=NULL WHERE agent_id=?
4. `RemoveAgent(name string) error` — transaction: get agent by name, unclaim tasks, unassign worktrees, DELETE agent

### Phase B: CLI Gap Fixes

1. **cmd/tasks.go** — Add `--fail` and `--result` flags:
   - Add `fail` string flag (task ID to fail)
   - Add `result` string flag (result text)
   - Wire `--fail` to `db.FailTask(fail, &result)`
   - Wire `--result` into existing `--complete` path: `db.CompleteTask(complete, &result)`

2. **cmd/agents.go** — New file:
   - `agit agents` — list all agents in a table (ID truncated, Name, Type, Status, Last Seen, Current Worktree)
   - `--sweep` flag — call `db.SweepStaleAgents(cfg.Agent.StaleAfter)`
   - `--remove <name>` flag — call `db.RemoveAgent(name)`

3. **cmd/status.go** — Add live conflict scanning:
   - Import `conflicts` package
   - Before `db.FindConflicts()`, call `conflicts.ScanAndUpdate(db, repo)` to refresh file touches

### Phase C: MCP Server Implementation

1. **internal/mcp/server.go**:
   - `NewServer(db *registry.DB, cfg *config.Config) *server.MCPServer`
   - Create `server.NewMCPServer("agit", version, server.WithResourceCapabilities(true, true))`
   - Register all 11 tools via `s.AddTool()`
   - Register static resources via `s.AddResource()`
   - Register template resources via `s.AddResourceTemplate()`

2. **internal/mcp/tools.go** — 11 tool handlers:
   - Each handler opens registry, performs operation, returns JSON via `mcp.NewToolResultText()`
   - Handlers wrap existing registry/git methods — no new business logic
   - Error handling: return `nil, fmt.Errorf(...)` for operational errors
   - `agit_merge_worktree`: merge → remove worktree from disk → mark completed

3. **internal/mcp/resources.go** — 5 resource handlers:
   - Each handler queries registry and returns JSON via `mcp.TextResourceContents`
   - `agit://repos` and `agit://agents` — static resources
   - `agit://repos/{name}`, `agit://repos/{name}/conflicts`, `agit://repos/{name}/tasks` — template resources

4. **cmd/serve.go** — Wire MCP server:
   - Replace TODO with actual MCP server creation
   - stdio: `server.ServeStdio(mcpServer)`
   - sse: `server.NewSSEServer(mcpServer).Start(fmt.Sprintf("127.0.0.1:%d", port))`

### Phase D: Integration & Verification

1. `go build ./...` — verify compilation
2. `go vet ./...` — verify no issues
3. Manual testing: pipe JSON-RPC initialize + tool calls through stdio
4. Verify all existing CLI commands still work (regression)

## Complexity Tracking

No constitution violations to justify. The implementation adds 4 new files and modifies 4 existing files — well within a single-feature scope.

## Dependencies Between Phases

```
Phase A (registry methods) → Phase B (CLI uses new methods)
Phase A (registry methods) → Phase C (MCP tools use registry)
Phase C (MCP server)       → Phase D (integration testing)
Phase B (CLI gaps)         → Phase D (regression testing)
```

Phase A must be completed first. Phases B and C can proceed in parallel after A.
