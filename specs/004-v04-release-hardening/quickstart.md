# Quickstart: v0.4.0 Release Hardening

**Branch**: `004-v04-release-hardening` | **Date**: 2026-02-22

## Prerequisites

- Go 1.22+ installed
- agit repository cloned and on `004-v04-release-hardening` branch
- `golangci-lint` installed

## Build & Test

```bash
# Build
go build ./...

# Run all tests with race detection
go test ./... -race

# Lint
golangci-lint run
```

## Feature Verification Scenarios

### US1: CLI Integration Tests

```bash
# Run CLI-level tests specifically
go test ./cmd/... -race -v

# Verify coverage target (≥60%)
go test ./cmd/... -coverprofile=coverage.out
go tool cover -func=coverage.out | grep total
```

### US2: Version Bump

```bash
# After version bump, verify
go run . --version
# Expected: agit version 0.4.0
```

### US3: Task Dispatch

```bash
# Create tasks with different priorities
agit tasks myrepo --create "low priority task" --priority 1
agit tasks myrepo --create "high priority task" --priority 10
agit tasks myrepo --create "medium priority task" --priority 5

# Dispatch next task (should return highest priority)
agit tasks next myrepo --agent agent-1
# Expected: claims "high priority task"

# JSON output
agit tasks next myrepo --agent agent-2 --output json
# Expected: JSON with "medium priority task"
```

### US4: Conflict Resolution Suggestions

```bash
# With two worktrees modifying same files
agit conflicts myrepo
# Expected: conflict list with "Suggested resolution order:" section

agit conflicts myrepo --output json
# Expected: JSON with "suggestions" array
```

### US5: Hook System

```bash
# Configure a hook
agit config set hooks.worktree.created "echo 'worktree created' >> /tmp/agit-hooks.log"

# Trigger the hook
agit spawn myrepo --task "test hooks" --agent test-agent

# Verify hook fired
cat /tmp/agit-hooks.log
# Expected: "worktree created"
```

### US6: SSE Graceful Shutdown

```bash
# Start SSE server
agit serve --transport=sse --port=3847 &
SSE_PID=$!

# Send SIGTERM
kill $SSE_PID

# Server should shut down cleanly within 5 seconds
# No "panic" or "leaked goroutine" messages
```

## MCP Integration Test

```bash
# Start MCP server and verify new tool is available
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | agit serve
# Expected: response includes "agit_next_task" tool
```
