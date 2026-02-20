# Quickstart: Automatic Bug Report Issue Link

**Feature Branch**: `003-auto-issue-link`
**Date**: 2026-02-20

## Implementation Sequence

### Phase 1: Core Infrastructure (User Story 1 foundation)

Build the two new packages that everything else depends on.

1. **Create `internal/errors/errors.go`** — Define `UserError` type with `Error()` method, `IsUserError()` checker using `errors.As`, and constructor functions `NewUserError()` / `NewUserErrorf()`.

2. **Create `internal/errors/errors_test.go`** — Test that `IsUserError` detects `UserError` through `fmt.Errorf("%w")` wrap chains, returns false for regular errors, and handles nil.

3. **Create `internal/issuelink/issuelink.go`** — Implement `Build(ctx Context) string` for URL generation, `Enabled() bool` for env var check, URL truncation to 2,000 chars, and `RepoURL` constant.

4. **Create `internal/issuelink/issuelink_test.go`** — Table-driven tests: valid URL output, special character encoding, truncation at limit, opt-out via env var, panic context formatting.

### Phase 2: CLI Integration (User Story 1)

Wire the issue link into the CLI error path.

5. **Modify `cmd/root.go`** — Set `rootCmd.SilenceErrors = true`. In `Execute()`, after `rootCmd.Execute()` returns an error, check `!errors.IsUserError(err)` and if true, call `issuelink.Build()` and print the URL on a separate line. Add `defer/recover` for panic handling with issue link.

6. **Migrate user input errors in `cmd/` files** — In each command file (`add.go`, `spawn.go`, `merge.go`, `conflicts.go`, `repos.go`, `tasks.go`, `cleanup.go`, `serve.go`), replace `fmt.Errorf(...)` with `errors.NewUserError(...)` or `errors.NewUserErrorf(...)` for input validation returns. Internal errors (database, git ops) remain as `fmt.Errorf`.

### Phase 3: MCP Integration (User Story 2)

7. **Modify `internal/mcp/tools.go`** — For "parameter is required" validation errors, return `UserError`. For internal errors, append the issue link URL to the error message string before returning.

### Phase 4: Opt-Out and Polish (User Story 3)

8. **Verify opt-out** — `Enabled()` already checks `AGIT_NO_ISSUE_LINK`. Verify that `Execute()` and MCP error handler both respect it.

9. **Verify build and lint** — Run `go build ./...`, `go vet ./...`, `go test ./... -race`. Ensure no regressions.

## Build Sequence

```bash
# After each phase, verify:
go build ./...
go vet ./...
go test ./... -race
```

## Verification Checklist

- [ ] `agit add /nonexistent` shows error WITHOUT issue link (user input error)
- [ ] Forcing a database error shows error WITH issue link
- [ ] Issue link URL is valid and opens GitHub with pre-filled fields
- [ ] `AGIT_NO_ISSUE_LINK=1 agit ...` suppresses issue link on internal errors
- [ ] URL with special characters in error message encodes correctly
- [ ] Long error messages produce truncated URLs ≤ 2,000 chars
- [ ] Panic produces crash message with issue link and stack trace
