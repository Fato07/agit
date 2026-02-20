# Research: Automatic Bug Report Issue Link

**Feature Branch**: `003-auto-issue-link`
**Date**: 2026-02-20

## R1: Error Classification Pattern

**Decision**: Custom `UserError` type with `errors.As` traversal.

**Rationale**: Go's `errors.As` checks the entire error chain, so a `UserError` wrapped with `fmt.Errorf("context: %w", userErr)` is still detected. This is the standard Go pattern (used by `os.IsNotExist`, `net.Error`). The alternative — sentinel errors via `errors.Is` — works but cannot carry structured metadata.

**Alternatives considered**:
- Sentinel error (`var ErrUserInput = errors.New(...)`) — simpler but less extensible
- String prefix matching — fragile, not idiomatic
- Error code integers — over-engineering for this use case

**Implementation**: New `internal/errors/errors.go` with `UserError` struct, `IsUserError()` helper, and `NewUserError()`/`NewUserErrorf()` constructors.

## R2: Panic Recovery in Cobra

**Decision**: Wrap `Execute()` with a top-level `defer/recover` in `cmd/root.go`.

**Rationale**: All subcommand panics propagate to `rootCmd.Execute()`. A single recovery point in `Execute()` catches everything without wrapping individual `RunE` functions. Uses `runtime/debug.Stack()` for stack trace capture.

**Alternatives considered**:
- Per-command `withPanicRecovery()` wrapper — more granular but duplicative
- `PersistentPreRunE` — runs before subcommand, does not wrap execution

## R3: GitHub Issue URL Format

**Decision**: Use `https://github.com/{owner}/{repo}/issues/new` with query params `title`, `body`, `labels`, `template`.

**Rationale**: GitHub's new-issue page accepts URL query parameters for pre-filling. The `template` param selects the issue template by filename. The `body` param overrides the template body with diagnostic content.

**Supported parameters**: `title`, `body`, `labels` (comma-separated or `labels[]=`), `template` (filename relative to `.github/ISSUE_TEMPLATE/`), `assignee`, `milestone`, `projects`.

## R4: URL Encoding and Truncation

**Decision**: Use `net/url.Values` for encoding; truncate body to keep total URL ≤ 2,000 characters.

**Rationale**: `url.Values.Encode()` handles all special characters (spaces, ampersands, newlines, non-ASCII) per RFC 3986. The 2,000-character limit is a safe cross-browser threshold. GitHub's server handles ~8,000 but browser address bars may not.

**Truncation strategy**: Build URL with all params. If over limit, shorten the `body` field (largest param) and append `[truncated]` note.

## R5: Runtime Environment Detection

**Decision**: Use `runtime.GOOS` and `runtime.GOARCH` for OS/arch; use `cmd.Version` (ldflags-injected) for version.

**Rationale**: `runtime.GOOS`/`GOARCH` always reflect the actual execution environment. Version is already injected via Makefile ldflags (`-X github.com/fathindos/agit/cmd.Version`).

## R6: Module Path / Repository URL

**Decision**: Use a constant `RepoURL` derived from the module path in `go.mod`.

**Rationale**: The module path `github.com/fathindos/agit` directly maps to `https://github.com/fathindos/agit`. A constant is simpler and more reliable than `debug.ReadBuildInfo()`, which returns `"command-line-arguments"` for local `go build` invocations. If the repo is ever renamed/forked, the constant is a single-line change.

**Alternative considered**: `debug.ReadBuildInfo().Main.Path` — works for `go install` but not for local builds.

## R7: Existing Error Handling Integration Points

**Decision**: Modify `cmd/root.go:Execute()` as the single CLI integration point; wrap MCP tool errors in `internal/mcp/tools.go`.

**Rationale**: All CLI errors converge in `Execute()` (root.go:30-36). Currently: `fmt.Fprintln(os.Stderr, err); os.Exit(1)`. This is the natural injection point. For MCP, errors return as `(*mcp.CallToolResult, error)` from tool handlers — the issue link should be appended to the error message string.

**Current error pattern**: All subcommands use `RunE` returning `error`. Errors wrap with `fmt.Errorf("context: %w", err)` using `%w` for chain preservation. No custom error types exist yet.

**Files requiring UserError migration** (input validation returns):
- `cmd/add.go`: "not a Git repository", path resolution failures
- `cmd/spawn.go`: missing repo argument, invalid branch names
- `cmd/merge.go`: missing worktree argument, conflict pre-check failures
- `cmd/conflicts.go`: missing repo argument
- `cmd/repos.go`: missing repo name for remove
- `cmd/tasks.go`: missing required arguments for task operations
- `cmd/cleanup.go`: missing/invalid arguments
- `cmd/serve.go`: invalid transport argument
- `internal/mcp/tools.go`: "parameter is required" errors
