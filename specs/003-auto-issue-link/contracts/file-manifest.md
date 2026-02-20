# File Manifest: Automatic Bug Report Issue Link

**Feature Branch**: `003-auto-issue-link`
**Date**: 2026-02-20

## New Files

| File | Purpose |
|------|---------|
| `internal/errors/errors.go` | `UserError` type, `IsUserError()`, `NewUserError()`, `NewUserErrorf()` |
| `internal/errors/errors_test.go` | Tests for UserError detection through wrap chains |
| `internal/issuelink/issuelink.go` | `Build()`, `Enabled()`, URL construction, truncation logic |
| `internal/issuelink/issuelink_test.go` | Table-driven tests: URL format, encoding, truncation, opt-out |

## Modified Files

| File | Change |
|------|--------|
| `cmd/root.go` | Add panic recovery defer; add issue link to error output; set `SilenceErrors = true` |
| `cmd/add.go` | Return `UserError` for input validation (not a git repo, path resolution) |
| `cmd/spawn.go` | Return `UserError` for missing repo, invalid branch |
| `cmd/merge.go` | Return `UserError` for missing worktree, conflict pre-check |
| `cmd/conflicts.go` | Return `UserError` for missing repo argument |
| `cmd/repos.go` | Return `UserError` for missing repo name on remove |
| `cmd/tasks.go` | Return `UserError` for missing required arguments |
| `cmd/cleanup.go` | Return `UserError` for missing/invalid arguments |
| `cmd/serve.go` | Return `UserError` for invalid transport argument |
| `internal/mcp/tools.go` | Return `UserError` for "parameter is required" errors; append issue link to internal error messages |

## Unchanged Files

All files under `internal/git/`, `internal/registry/`, `internal/config/`, and `internal/conflicts/` are unchanged. These packages return internal errors that will naturally get issue links through the centralized error handler.
