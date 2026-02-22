# Feature Specification: v0.4.0 Release Hardening

**Feature Branch**: `004-v04-release-hardening`
**Created**: 2026-02-21
**Status**: Draft
**Input**: Version bump to 0.4.0, integration/E2E tests, agents command improvements, MCP SSE transport hardening, conflict resolution workflows, task prioritization/scheduling, and plugin/hook system.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - CLI Integration Tests Catch Regressions (Priority: P1)

A developer contributing to agit runs `go test ./...` and gets confidence that all CLI commands work end-to-end — not just internal packages. Currently the `cmd/` package has zero tests, so a broken flag or wiring error goes undetected until manual testing.

**Why this priority**: Without CLI-level tests, every release risks shipping broken commands. This is the foundation that enables safe iteration on all other stories.

**Independent Test**: Can be fully tested by running `go test ./cmd/... -race` and verifying all commands exercise their core paths against an in-memory or temp-dir database.

**Acceptance Scenarios**:

1. **Given** a fresh test environment, **When** each CLI command is invoked programmatically with valid arguments, **Then** it exits with code 0 and produces expected output.
2. **Given** a command is invoked with invalid arguments, **When** it runs, **Then** it exits with a non-zero code and prints a user-friendly error.
3. **Given** the `--output json` flag is passed to any command, **When** the command succeeds, **Then** the output is valid JSON.
4. **Given** any new feature is added in the future, **When** CI runs, **Then** the existing CLI tests still pass (regression protection).

---

### User Story 2 - Version Bump and Release (Priority: P1)

A project maintainer bumps the version to 0.4.0 and creates a tagged release so users running `agit update` receive all recent improvements (config command, 19 MCP tools, test coverage).

**Why this priority**: Shipping the work already merged to main delivers immediate value. Tied P1 with tests because both are release prerequisites.

**Independent Test**: Can be verified by running `agit --version` and confirming it prints `0.4.0`, and by checking that the GitHub release page shows the new tag.

**Acceptance Scenarios**:

1. **Given** all CI checks pass on main, **When** a v0.4.0 tag is pushed, **Then** GoReleaser builds binaries for all platforms and publishes a GitHub release.
2. **Given** a user has agit v0.3.0 installed, **When** they run `agit update`, **Then** they are upgraded to v0.4.0.

---

### User Story 3 - Task Queue-Based Dispatch (Priority: P2)

An AI agent orchestrator wants to create a batch of tasks with different priorities and have agents automatically pick up the highest-priority pending task. Currently priority is stored but there is no "next task" dispatch — agents must list all tasks and manually choose.

**Why this priority**: Task coordination is a core value proposition. Priority-based dispatch makes multi-agent workflows practical without human coordination.

**Independent Test**: Can be tested by creating tasks with different priorities, calling a dispatch endpoint, and verifying the highest-priority pending task is returned and atomically claimed.

**Acceptance Scenarios**:

1. **Given** multiple pending tasks with different priorities exist, **When** an agent requests the next task, **Then** the system returns the highest-priority pending task and atomically claims it.
2. **Given** no pending tasks exist for a repository, **When** an agent requests the next task, **Then** the system returns an empty result (not an error).
3. **Given** two agents request the next task simultaneously, **When** both requests arrive, **Then** each agent receives a different task (no double-claiming).

---

### User Story 4 - Conflict Resolution Guidance (Priority: P2)

A developer or agent sees that two worktrees have conflicting file changes. They want agit to show which specific files conflict, which worktrees/agents are involved, and suggest a resolution path (e.g., "merge worktree A first, then rebase worktree B").

**Why this priority**: Conflict detection already exists; resolution guidance is the natural next step that turns detection into actionable workflow.

**Independent Test**: Can be tested by creating two worktrees with overlapping file touches, running the conflicts command, and verifying the output includes resolution suggestions.

**Acceptance Scenarios**:

1. **Given** two active worktrees have modified the same file, **When** the user runs `agit conflicts <repo>`, **Then** the output shows the conflicting files, involved worktrees, agents, and a suggested resolution order.
2. **Given** conflicts exist, **When** the output is rendered in JSON mode, **Then** the JSON includes a `suggestions` array with structured resolution steps.
3. **Given** no conflicts exist, **When** the user runs `agit conflicts <repo>`, **Then** a clean message is displayed with no suggestions.

---

### User Story 5 - Plugin / Hook System for Event-Driven Automation (Priority: P3)

A team wants to trigger custom actions when certain agit events occur — for example, send a Slack notification when a task is claimed, run linting when a worktree is created, or update a dashboard when conflicts are detected. They configure hooks in the agit config file.

**Why this priority**: Hooks enable extensibility without modifying agit core. Lower priority because the core workflow works without them, but they unlock significant automation potential.

**Independent Test**: Can be tested by configuring a hook that writes to a temp file, triggering the event, and verifying the file was created with expected content.

**Acceptance Scenarios**:

1. **Given** a hook is configured for the `worktree.created` event, **When** a worktree is spawned, **Then** the hook command executes with event details as environment variables.
2. **Given** a hook is configured for `task.claimed`, **When** a task is claimed by an agent, **Then** the hook receives task ID, agent ID, and description.
3. **Given** a hook command fails (non-zero exit), **When** the triggering operation runs, **Then** the operation still succeeds and the hook failure is logged as a warning.
4. **Given** no hooks are configured, **When** events fire, **Then** there is no performance impact or error output.

---

### User Story 6 - MCP SSE Transport Hardening (Priority: P3)

An operator runs `agit serve --transport=sse` to expose the MCP server over HTTP for remote agents. They expect graceful shutdown, connection health monitoring, and proper error responses.

**Why this priority**: SSE transport works at a basic level already. Hardening is important for production use but not blocking for local development workflows.

**Independent Test**: Can be tested by starting the SSE server, making HTTP requests, verifying responses, and testing graceful shutdown behavior.

**Acceptance Scenarios**:

1. **Given** the SSE server is running, **When** a client connects and calls a tool, **Then** the server returns a valid MCP response.
2. **Given** the SSE server is running, **When** a SIGTERM is received, **Then** the server shuts down gracefully within a reasonable window.
3. **Given** the SSE server is running, **When** an invalid request arrives, **Then** the server returns an appropriate error without crashing.

---

### Edge Cases

- What happens when a hook command hangs or takes too long? (Assumed: hooks have a global configurable timeout, default 30 seconds. Per-event timeout configuration deferred to a future version.)
- What happens when priority dispatch is called concurrently by 10+ agents? (Must be atomic — SQLite's built-in locking handles this)
- What happens when conflict resolution suggests merging a worktree that has since been deleted? (Suggestions are advisory; stale worktrees are filtered out)
- What happens when the SSE server port is already in use? (Clear error message on startup)
- What happens when a CLI test's temp database is corrupted? (Each test uses a fresh in-memory DB; no shared state)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide CLI-level integration tests that exercise all commands (`init`, `add`, `repos`, `spawn`, `status`, `conflicts`, `tasks`, `merge`, `cleanup`, `serve`, `update`, `config`, `agents`) with both success and error paths.
- **FR-002**: System MUST bump the version to 0.4.0 and produce a tagged release with binaries for all supported platforms.
- **FR-003**: System MUST provide a "next task" dispatch mechanism that atomically returns and claims the highest-priority pending task for a given repository.
- **FR-004**: The dispatch mechanism MUST be exposed as both a CLI subcommand (`agit tasks next <repo>`) and an MCP tool (`agit_next_task`).
- **FR-005**: System MUST enrich conflict output with resolution suggestions, including a recommended merge order based on worktree creation time and file overlap count.
- **FR-006**: Resolution suggestions MUST be available in both text and JSON output formats.
- **FR-007**: System MUST support event hooks configured in `~/.agit/config.toml` under a `[hooks]` section.
- **FR-008**: Supported hook events MUST include: `worktree.created`, `worktree.removed`, `task.claimed`, `task.completed`, `task.failed`, `conflict.detected`.
- **FR-009**: Hook commands MUST receive event context as environment variables including `AGIT_EVENT`, `AGIT_REPO`, `AGIT_REPO_ID`, `AGIT_TASK_ID`, `AGIT_AGENT_ID`, `AGIT_WORKTREE_ID`, and `AGIT_WORKTREE_PATH` (set when applicable to the event).
- **FR-010**: Hook execution MUST be non-blocking — a failing or slow hook MUST NOT prevent the triggering operation from completing.
- **FR-011**: SSE transport MUST handle graceful shutdown on SIGTERM/SIGINT.
- **FR-012**: All new features MUST pass `go build ./...`, `go test ./... -race`, and `golangci-lint run`.

### Key Entities

- **Hook**: An event name mapped to a shell command, configured in TOML. Attributes: event name, command string, timeout duration.
- **Task Dispatch**: A mechanism that atomically selects and claims the highest-priority pending task. Implemented as a single SQL query with `UPDATE ... WHERE ... ORDER BY priority DESC LIMIT 1`.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: All CLI commands have at least one integration test covering the happy path; `cmd/` package test coverage reaches at least 60%.
- **SC-002**: `go test ./... -race` passes with zero failures across all packages.
- **SC-003**: Task dispatch returns the correct highest-priority task in under 100ms for repositories with up to 1,000 tasks.
- **SC-004**: Conflict resolution suggestions are present in output whenever 2+ worktrees touch the same file.
- **SC-005**: Hook execution adds no more than 50ms latency to the triggering operation (hooks run asynchronously).
- **SC-006**: SSE server starts, handles at least one tool call, and shuts down cleanly without leaked resources.
- **SC-007**: v0.4.0 release is published with binaries for linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, and windows/amd64.

## Assumptions

- Hook commands are standard shell commands executed via `sh -c` on Unix and `cmd /c` on Windows.
- Hook timeout defaults to 30 seconds (global setting). Per-event timeout configuration is deferred to a future version.
- Task dispatch uses SQLite's row-level locking (via `UPDATE ... WHERE`) for atomicity — no external lock manager needed.
- Conflict resolution suggestions are advisory text, not automated actions. Automated merge/rebase is out of scope.
- SSE hardening focuses on graceful shutdown and error handling, not TLS or authentication (those are assumed to be handled by a reverse proxy in production).
- Integration tests use `exec.Command` to invoke the compiled `agit` binary or call Cobra's `Execute` programmatically with redirected I/O.
