# Data Model: Automatic Bug Report Issue Link

**Feature Branch**: `003-auto-issue-link`
**Date**: 2026-02-20

## Entities

### UserError (new type)

Marks an error as caused by invalid user input. These errors do NOT trigger issue link generation.

| Field | Type | Description |
|-------|------|-------------|
| msg | string | Human-readable error message |

**Behavior**:
- Implements `error` interface via `Error() string`
- Detectable via `errors.As` through any wrap chain
- Created via `NewUserError(msg)` or `NewUserErrorf(format, args...)`
- Checked via `IsUserError(err) bool`

### ErrorContext (value object)

Diagnostic information gathered at the point of failure, used to build the issue URL.

| Field | Type | Description |
|-------|------|-------------|
| Err | error | The original error |
| Command | []string | The full command line (`os.Args`) |
| Version | string | agit version (ldflags-injected) |
| OS | string | Runtime operating system (`runtime.GOOS`) |
| Arch | string | Runtime architecture (`runtime.GOARCH`) |

**Derived at runtime** — OS and Arch are populated from `runtime` package, not stored.

### IssueURL (computed value)

A GitHub "new issue" URL with pre-filled query parameters. Not persisted — generated on-the-fly from ErrorContext.

| Parameter | Source | Example |
|-----------|--------|---------|
| title | First line of error message | `Bug: could not open database` |
| body | Formatted ErrorContext | Error + version + OS + command |
| labels | Constant | `bug` |
| template | Constant | `bug_report.md` |

**Constraints**:
- Total URL length ≤ 2,000 characters
- Body truncated with `[truncated]` note if over limit
- All values URL-encoded via `net/url.Values`

## Relationships

```
UserError ──(is-a)──> error interface
ErrorContext ──(contains)──> error
ErrorContext ──(produces)──> IssueURL (via Build function)
```

## State Transitions

No state machines. Error classification is a one-time check at the error handling boundary:

```
error occurs → IsUserError? → yes → display error only
                             → no  → display error + IssueURL
```

## No Persistence

This feature has no database, file, or network storage requirements. All entities are ephemeral values computed at error time.
