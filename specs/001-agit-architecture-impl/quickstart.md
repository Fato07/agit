# Quickstart: agit Architecture Implementation

**Branch**: `001-agit-architecture-impl` | **Date**: 2026-02-20

## Prerequisites

- Go 1.22+
- Git installed
- agit built: `go build -o agit .`

## Build & Verify

```bash
# Build
go build -o agit .

# Verify
./agit version
./agit --help
```

## Development Workflow

### Phase 7: MCP Server (primary new work)

1. Create `internal/mcp/server.go` — server setup and registration
2. Create `internal/mcp/tools.go` — 11 tool handlers
3. Create `internal/mcp/resources.go` — 5 resource handlers
4. Update `cmd/serve.go` — wire MCP server to CLI command

**Test manually**:
```bash
# Initialize and add a repo
./agit init
./agit add /path/to/some/git/repo

# Test stdio transport (pipe JSON-RPC messages)
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' | ./agit serve

# Test SSE transport
./agit serve --transport sse --port 3847 &
curl http://127.0.0.1:3847/sse
```

### Phase 4 Gap: Task --fail and --result flags

```bash
# Create and fail a task
./agit tasks myapp --create "implement feature X"
./agit tasks myapp --claim t-abc123 --agent claude-1
./agit tasks myapp --fail t-abc123 --result "compilation error in main.go"

# Complete with result
./agit tasks myapp --complete t-abc123 --result "merged in PR #42"
```

### Phase 5 Gap: Agent management

```bash
# List agents
./agit agents

# Sweep stale agents
./agit agents --sweep

# Remove an agent (soft release)
./agit agents --remove old-agent
```

## MCP Client Configuration

Add to your AI agent's MCP settings:

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

## Key Files

| File | Purpose |
|------|---------|
| `internal/mcp/server.go` | MCP server creation and registration |
| `internal/mcp/tools.go` | 11 MCP tool handlers |
| `internal/mcp/resources.go` | 5 MCP resource handlers |
| `cmd/serve.go` | CLI wiring for `agit serve` |
| `cmd/tasks.go` | Add `--fail` and `--result` flags |
| `cmd/agents.go` | New `agit agents` subcommand |
| `internal/registry/agents.go` | Add SweepStaleAgents, RemoveAgent |
| `cmd/status.go` | Add live conflict scanning |
