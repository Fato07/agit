# Implementation Plan: v0.4.0 Release Hardening

**Branch**: `004-v04-release-hardening` | **Date**: 2026-02-22 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/004-v04-release-hardening/spec.md`

## Summary

Harden agit for v0.4.0 release across six workstreams: CLI integration tests for the `cmd/` package (P1), version bump and tagged release (P1), atomic task dispatch with priority-based queue (P2), conflict resolution suggestions (P2), event-driven hook/plugin system (P3), and SSE transport graceful shutdown (P3). Research decisions favor Cobra programmatic execution for tests, single-SQL atomic dispatch, fewer-conflicts-first merge ordering, TOML-based hook config, and signal-based SSE shutdown.

## Technical Context

**Language/Version**: Go 1.23 (go.mod), CI tests against 1.22 + 1.23
**Primary Dependencies**: `github.com/spf13/cobra` v1.8.1 (CLI), `github.com/mark3labs/mcp-go` v0.20.1 (MCP protocol), `modernc.org/sqlite` v1.34.4 (database), `github.com/pelletier/go-toml/v2` (config)
**Storage**: SQLite via `modernc.org/sqlite` (pure Go, CGO_ENABLED=0)
**Testing**: `go test ./... -race`, `golangci-lint run`
**Target Platform**: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64
**Project Type**: Single Go CLI binary
**Performance Goals**: Task dispatch <100ms for 1,000 tasks; hook execution adds <50ms latency (async)
**Constraints**: CGO_ENABLED=0; single static binary; no external runtime dependencies
**Scale/Scope**: CLI tool for individual developers and small teams; up to 1,000 tasks per repository

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Single Static Binary | PASS | No new CGO dependencies. Hook execution uses `os/exec` (stdlib). SSE uses existing `net/http`. |
| II. MCP-First Interface | PASS | `agit_next_task` MCP tool added alongside `tasks next` CLI subcommand. Conflict suggestions exposed via both CLI and MCP. |
| III. Worktree Isolation | PASS | No changes to worktree isolation model. Hooks receive worktree context but don't modify isolation. |
| IV. Convention Over Configuration | PASS | Hooks are optional config. Task dispatch works without config. All features have sensible defaults. |
| V. Standard Go Practices | PASS | CLI integration tests follow Go testing conventions. All code passes `go vet`, `golangci-lint`, `go build ./...`. |
| VI. Agent-Agnostic Design | PASS | Task dispatch accepts any agent ID. Hooks are generic shell commands, not agent-specific. |

**Post-Phase 1 Re-check**: All principles still satisfied. Hook system adds optional config but doesn't violate Convention Over Configuration (empty `[hooks]` section = no-op). SSE shutdown is standard Go `http.Server.Shutdown()` pattern.

## Project Structure

### Documentation (this feature)

```text
specs/004-v04-release-hardening/
├── plan.md              # This file
├── research.md          # Phase 0 output (R1-R6 decisions)
├── data-model.md        # Phase 1 output (entity descriptions)
├── quickstart.md        # Phase 1 output (verification scenarios)
├── contracts/           # Phase 1 output
│   ├── cli-contracts.md # CLI command contracts
│   └── mcp-contracts.md # MCP tool contracts
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
cmd/                          # Cobra CLI commands
├── root.go                   # Root command, global flags, update check
├── tasks.go                  # Tasks command (modify: add "next" subcommand)
├── conflicts.go              # Conflicts command (modify: add resolution suggestions)
├── serve.go                  # Serve command (modify: add graceful shutdown)
├── config.go                 # Config command (existing)
└── *_test.go                 # NEW: CLI integration tests for all commands

internal/
├── config/
│   └── config.go             # Modify: add hooks config struct
├── registry/
│   ├── db.go                 # Modify: add NextTask method
│   └── tasks.go              # Existing task CRUD
├── conflicts/
│   ├── detector.go           # Existing conflict detection
│   └── resolver.go           # NEW: resolution suggestion algorithm
├── hooks/
│   └── hooks.go              # NEW: hook execution engine
├── mcp/
│   ├── server.go             # Modify: register agit_next_task tool
│   └── tools.go              # Modify: add handleNextTask handler
└── git/
    └── diff.go               # Existing (no changes)
```

**Structure Decision**: Extends the existing single-project Go structure. Two new packages (`internal/hooks`, `internal/conflicts/resolver.go`) and CLI integration tests added to `cmd/`. No structural reorganization needed.

## Complexity Tracking

No constitution violations. All features fit within the existing architecture.
