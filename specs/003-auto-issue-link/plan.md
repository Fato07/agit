# Implementation Plan: Automatic Bug Report Issue Link

**Branch**: `003-auto-issue-link` | **Date**: 2026-02-20 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/003-auto-issue-link/spec.md`

## Summary

When agit encounters an unexpected internal error (CLI or MCP), it generates a pre-filled GitHub issue URL containing the error message, version, OS/arch, and triggering command. The URL is appended to the error output so users can report bugs with one click. User input validation errors do not produce issue links. A new `internal/issuelink` package handles URL construction, truncation (2,000-char limit), and opt-out (`AGIT_NO_ISSUE_LINK=1`). A custom `UserError` type in `internal/errors` distinguishes user input errors from internal errors via `errors.As`.

## Technical Context

**Language/Version**: Go 1.23 (go.mod), CI tests against 1.22 + 1.23
**Primary Dependencies**: `github.com/spf13/cobra` (CLI), `github.com/mark3labs/mcp-go` (MCP protocol), stdlib `net/url`, `runtime`, `runtime/debug`
**Storage**: N/A (no persistence; URL generation is pure string computation)
**Testing**: `go test ./... -race` with table-driven tests
**Target Platform**: linux/darwin/windows on amd64/arm64 (CGO_ENABLED=0)
**Project Type**: Single Go binary (existing structure)
**Performance Goals**: URL generation < 1ms (pure string ops, no I/O)
**Constraints**: URL length ≤ 2,000 characters; zero new external dependencies
**Scale/Scope**: Touches ~15 files (1 new package, error type refactoring across cmd/ and internal/mcp/)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Single Static Binary | PASS | No new dependencies; uses stdlib only (`net/url`, `runtime`, `runtime/debug`) |
| II. MCP-First Interface | PASS | Issue links appear in both CLI error output and MCP tool error responses (FR-001, FR-005) |
| III. Worktree Isolation | N/A | Feature does not modify worktree behavior |
| IV. Convention Over Configuration | PASS | Works by default with zero config; opt-out via env var is optional override |
| V. Standard Go Practices | PASS | Uses idiomatic error types (`errors.As`), `go vet`/`golangci-lint` compliant, no build breakage |
| VI. Agent-Agnostic Design | PASS | Issue links are plain URLs; no agent-specific logic |

**Gate result**: PASS — no violations.

## Project Structure

### Documentation (this feature)

```text
specs/003-auto-issue-link/
├── plan.md              # This file
├── research.md          # Phase 0: research findings
├── data-model.md        # Phase 1: error types and entities
├── quickstart.md        # Phase 1: implementation sequence
├── contracts/
│   └── file-manifest.md # Files to create/modify
└── tasks.md             # Phase 2 output (via /speckit.tasks)
```

### Source Code (repository root)

```text
internal/
├── errors/
│   └── errors.go        # NEW: UserError type, IsUserError(), NewUserError()
├── issuelink/
│   ├── issuelink.go     # NEW: Build(), Enabled(), URL construction, truncation
│   └── issuelink_test.go # NEW: table-driven tests for URL generation
├── mcp/
│   ├── tools.go         # MODIFY: wrap internal errors with issue link in error responses
│   └── server.go        # MODIFY: add error wrapping middleware (optional)
└── ...

cmd/
├── root.go              # MODIFY: add panic recovery, issue link in Execute()
├── add.go               # MODIFY: return UserError for input validation
├── spawn.go             # MODIFY: return UserError for input validation
├── merge.go             # MODIFY: return UserError for input validation
├── conflicts.go         # MODIFY: return UserError for input validation
├── repos.go             # MODIFY: return UserError for input validation
├── tasks.go             # MODIFY: return UserError for input validation
├── cleanup.go           # MODIFY: return UserError for input validation
└── serve.go             # MODIFY: return UserError for input validation
```

**Structure Decision**: Follows existing `internal/` package layout. Two new packages (`internal/errors`, `internal/issuelink`) keep the feature self-contained. Modifications to existing `cmd/` and `internal/mcp/` files are surgical — only error return points change.

## Complexity Tracking

> No constitution violations to justify.
