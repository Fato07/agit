# Research: agit Full Architecture Implementation

**Branch**: `001-agit-architecture-impl` | **Date**: 2026-02-20

## R1: mcp-go Library API Patterns

**Decision**: Use `github.com/mark3labs/mcp-go` v0.20.1 (already in go.mod) for MCP server implementation.

**Rationale**: It is the standard Go MCP SDK, already declared as a dependency. The API is straightforward:
- `server.NewMCPServer(name, version, opts...)` creates the server
- `s.AddTool(tool, handler)` registers tools with typed handlers
- `s.AddResource(resource, handler)` registers static resources
- `s.AddResourceTemplate(template, handler)` registers dynamic URI-templated resources
- `server.ServeStdio(s)` starts stdio transport with signal handling
- `server.NewSSEServer(s, opts...).Start(addr)` starts SSE transport

**Key Types**:
- Tool handler: `func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)`
- Resource handler: `func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error)`
- Tool definition: `mcp.NewTool(name, mcp.WithDescription(...), mcp.WithString(..., mcp.Required()), ...)`
- Resource definition: `mcp.NewResource(uri, name, opts...)` or `mcp.NewResourceTemplate(uriTemplate, name, opts...)`
- Results: `mcp.NewToolResultText(text)` for string results; for JSON, marshal and use text result

**Alternatives considered**: Building raw JSON-RPC over stdio — rejected because mcp-go handles protocol details, session management, and signal handling.

## R2: MCP Server Architecture Pattern

**Decision**: Create `internal/mcp/` package with server setup and tool/resource handlers. Wire into `cmd/serve.go`.

**Rationale**: Follows existing project convention where `internal/` contains business logic and `cmd/` contains CLI wiring. Separating MCP handlers into their own package keeps them testable and decoupled from cobra.

**File structure**:
- `internal/mcp/server.go` — Server creation, tool/resource registration
- `internal/mcp/tools.go` — All 11 tool handlers
- `internal/mcp/resources.go` — All 5 resource handlers

**Alternatives considered**: Single file — rejected because 11 tools + 5 resources would create a 500+ line file; embedding handlers in cmd/serve.go — rejected because it mixes CLI wiring with business logic.

## R3: SSE Localhost Binding

**Decision**: SSE transport binds to `127.0.0.1:<port>` only. No auth needed.

**Rationale**: agit is a single-machine developer tool. The SSE transport is for local debugging and agent integration testing. Network exposure is unnecessary and would require auth complexity.

**Implementation**: Pass `fmt.Sprintf("127.0.0.1:%d", port)` to `sseServer.Start()`.

**Alternatives considered**: Binding to 0.0.0.0 with token auth — rejected per clarification session (over-engineering for the use case).

## R4: Agent Removal Cascading

**Decision**: Soft release on agent removal — unclaim tasks (revert to pending), unassign worktrees (clear agent_id), delete agent record.

**Rationale**: Preserves work already done in worktrees. Tasks become available for other agents. No data is lost.

**Implementation**: Add `RemoveAgent(name)` method to registry that performs the soft release in a transaction, then deletes the agent.

## R5: Merge Worktree Auto-Cleanup

**Decision**: After successful merge, remove worktree from disk and mark "completed" in registry.

**Rationale**: Merge is the natural end of a worktree lifecycle. Leaving orphaned worktrees creates clutter.

**Implementation**: `agit_merge_worktree` handler calls git merge, then git worktree remove, then registry update — in sequence with rollback on failure.

## R6: Agent Sweep Implementation

**Decision**: Use `config.Agent.StaleAfter` duration (default 5m) to determine stale agents.

**Rationale**: Already defined in config.go. The sweep compares `last_seen + stale_after` against current time.

**Implementation**: Add `SweepStaleAgents(staleAfter time.Duration)` method to registry that updates status to "disconnected" for agents where `last_seen < now - staleAfter`.

## R7: Status Live Conflict Scanning

**Decision**: Run `conflicts.ScanAndUpdate()` before displaying conflict data in `agit status`.

**Rationale**: Currently status shows stale touch data. A fresh scan ensures accuracy without requiring a separate `agit conflicts` invocation.

**Implementation**: Import `conflicts` package in `cmd/status.go`, call `ScanAndUpdate(db, repo)` before `FindConflicts()`.
