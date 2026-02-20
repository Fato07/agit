# Feature Specification: agit Full Architecture Implementation

**Feature Branch**: `001-agit-architecture-impl`
**Created**: 2026-02-20
**Status**: Draft
**Input**: Full agit architecture specification and implementation covering all 7 phases

## User Scenarios & Testing *(mandatory)*

### User Story 1 - MCP Server Exposes Tools for AI Agent Orchestration (Priority: P1)

An AI agent (e.g., Claude Code) connects to `agit serve` via MCP and uses tools to discover repos, spawn worktrees, manage tasks, and coordinate work — all without CLI invocation.

**Why this priority**: The MCP server (Phase 7) is the only entirely unimplemented phase. It is the primary integration surface that allows AI agents to programmatically interact with agit. Without it, agents must shell out to CLI commands.

**Independent Test**: Start `agit serve`, connect an MCP client, call `agit_list_repos`, and verify a JSON response listing registered repositories.

**Acceptance Scenarios**:

1. **Given** agit is initialized with at least one registered repo, **When** an MCP client calls the `agit_list_repos` tool, **Then** the server returns a JSON array of repo objects with name, path, default_branch, and remote_url.
2. **Given** a registered repo "myapp", **When** an MCP client calls `agit_spawn_worktree` with `{"repo": "myapp", "task": "fix auth"}`, **Then** a new worktree is created on disk, recorded in the registry, and the response includes the worktree path and branch name.
3. **Given** two active worktrees modifying overlapping files, **When** an MCP client calls `agit_check_conflicts` with the repo name, **Then** the response lists each conflicting file path and the worktree IDs that touch it.
4. **Given** a pending task exists, **When** an MCP client calls `agit_claim_task` with the task ID and agent ID, **Then** the task status changes to "claimed" and the assigned agent is recorded.
5. **Given** a claimed task, **When** `agit_complete_task` is called with a result string, **Then** the task status changes to "completed" and the result is persisted.
6. **Given** an agent is registered, **When** `agit_heartbeat` is called with the agent ID, **Then** the agent's `last_seen` timestamp is updated and status remains "active".
7. **Given** the server is started with `--transport sse --port 3847`, **When** an HTTP client connects to the SSE endpoint, **Then** the server accepts the connection and processes MCP messages over SSE.

---

### User Story 2 - MCP Resources Provide Read-Only State Discovery (Priority: P1)

An AI agent reads MCP resources (e.g., `agit://repos`, `agit://repos/{name}`) to understand the current state of the orchestration environment without modifying anything.

**Why this priority**: Resources complement tools by providing a read-only discovery mechanism. Together with tools, they complete the MCP server surface.

**Independent Test**: Connect to the MCP server, read `agit://repos`, and verify a JSON listing of all registered repos.

**Acceptance Scenarios**:

1. **Given** three repos are registered, **When** an MCP client reads `agit://repos`, **Then** it receives a JSON array with all three repos and their metadata.
2. **Given** repo "myapp" has two active worktrees and one pending task, **When** an MCP client reads `agit://repos/myapp`, **Then** it receives the repo details including worktree and task counts.
3. **Given** two agents are registered (one active, one stale), **When** an MCP client reads `agit://agents`, **Then** it receives both agents with their status, last_seen, and current worktree info.
4. **Given** repo "myapp" has file conflicts between two worktrees, **When** an MCP client reads `agit://repos/myapp/conflicts`, **Then** it receives the conflict report with file paths and involved worktrees.
5. **Given** repo "myapp" has tasks in various states, **When** an MCP client reads `agit://repos/myapp/tasks`, **Then** it receives all tasks with their status, assigned agent, and timestamps.

---

### User Story 3 - Task Fail and Result Flags in CLI (Priority: P2)

A user or script can mark a task as failed via the CLI with `--fail` and attach result metadata with `--result`, completing the task lifecycle from the command line.

**Why this priority**: The registry already supports `FailTask` and result storage, but the CLI `tasks` command lacks `--fail` and `--result` flags. This is a small gap that completes Phase 4.

**Independent Test**: Create a task, claim it, then run `agit tasks <repo> --fail <id> --result "compilation error"` and verify the task is marked failed with the result stored.

**Acceptance Scenarios**:

1. **Given** a task in "claimed" or "in_progress" status, **When** `agit tasks <repo> --fail <id>` is executed, **Then** the task status changes to "failed" with `completed_at` set.
2. **Given** a task being completed, **When** `agit tasks <repo> --complete <id> --result "merged PR #42"` is executed, **Then** the task status is "completed" and the result string is persisted.
3. **Given** a task being failed, **When** `agit tasks <repo> --fail <id> --result "tests broke"` is executed, **Then** the task status is "failed" and the result string is persisted.

---

### User Story 4 - Agent Management CLI Subcommand (Priority: P2)

A user can list, inspect, and clean up agents via `agit agents` — seeing which agents are active, which are stale, and removing dead registrations.

**Why this priority**: Agent data exists in the registry and is used by `spawn` and `tasks`, but there is no dedicated CLI command for agent lifecycle management. This completes Phase 5.

**Independent Test**: Register two agents (one via `spawn --agent`), then run `agit agents` and verify both appear with their status and last_seen time.

**Acceptance Scenarios**:

1. **Given** three agents are registered, **When** `agit agents` is executed, **Then** a table is displayed with columns: ID (truncated), Name, Type, Status, Last Seen, Current Worktree.
2. **Given** an agent has not sent a heartbeat for longer than `stale_after` (default 5m), **When** `agit agents --sweep` is executed, **Then** that agent's status is updated to "disconnected".
3. **Given** agent "builder-1" is registered, **When** `agit agents --remove builder-1` is executed, **Then** the agent record is deleted from the registry.

---

### User Story 5 - Core Registry Initialization and Repo Management (Priority: P3)

A user initializes agit, registers Git repositories, and manages the repo list — the foundational data layer that all other phases depend on.

**Why this priority**: Already fully implemented in Phase 1. Included for completeness and regression coverage.

**Independent Test**: Run `agit init`, then `agit add /path/to/repo`, then `agit repos` and verify the repo appears.

**Acceptance Scenarios**:

1. **Given** agit is not initialized, **When** `agit init` is run, **Then** `~/.agit/` is created with `config.toml` and `agit.db` (SQLite with WAL mode and foreign keys enabled).
2. **Given** a valid Git repository at `/tmp/myrepo`, **When** `agit add /tmp/myrepo` is run, **Then** the repo is recorded in the `repos` table with auto-detected name, remote_url, and default_branch.
3. **Given** two repos are registered, **When** `agit repos` is run, **Then** a table lists both with name, path, branch, and worktree counts.
4. **Given** repo "myapp" is registered, **When** `agit repos --remove myapp` is run, **Then** the repo and its associated worktrees, tasks, and touches are deleted (cascading).

---

### User Story 6 - Worktree Spawn and Lifecycle Management (Priority: P3)

A user spawns isolated worktrees for agents, lists active worktrees, and cleans up completed ones — the core isolation mechanism.

**Why this priority**: Already fully implemented in Phase 2. Included for completeness.

**Independent Test**: Run `agit spawn myapp --task "fix bug" --agent "claude"` and verify the worktree directory exists and is registered.

**Acceptance Scenarios**:

1. **Given** repo "myapp" is registered, **When** `agit spawn myapp --task "fix bug" --branch agit/fix-bug-abc123` is run, **Then** a Git worktree is created at `<repo>/.worktrees/agit-abc123` on a new branch.
2. **Given** an active worktree exists, **When** `agit cleanup myapp` is run, **Then** completed/stale worktrees are removed from disk and registry.
3. **Given** two agents have active worktrees, **When** `agit status myapp` is run, **Then** all active worktrees are displayed with their branch, agent, and task info.

---

### User Story 7 - Cross-Worktree Conflict Detection (Priority: P3)

The system detects when multiple worktrees modify the same files and warns users before merge conflicts occur.

**Why this priority**: Already fully implemented in Phase 3. Included for completeness.

**Independent Test**: Create two worktrees that both modify `src/main.go`, run `agit conflicts myapp`, and verify the overlap is reported.

**Acceptance Scenarios**:

1. **Given** worktree A modifies `src/auth.go` and worktree B modifies `src/auth.go`, **When** `agit conflicts myapp` is run, **Then** the report shows `src/auth.go` is modified in both worktrees.
2. **Given** worktree A modifies `README.md` and worktree B modifies `src/main.go` (no overlap), **When** `agit conflicts myapp` is run, **Then** no conflicts are reported.
3. **Given** a worktree's status is "completed", **When** conflict detection runs, **Then** the completed worktree's file touches are excluded from overlap analysis.

---

### User Story 8 - Status Dashboard with Live Conflict Scanning (Priority: P3)

The status command provides a unified dashboard showing repos, worktrees, agents, tasks, and conflicts — optionally running a live conflict scan before display.

**Why this priority**: Status display is implemented. The gap is triggering a fresh conflict scan before displaying results (currently shows stale touch data).

**Independent Test**: Modify files in a worktree, run `agit status myapp`, and verify freshly-detected conflicts appear in the output.

**Acceptance Scenarios**:

1. **Given** repo "myapp" has active worktrees with recent changes, **When** `agit status myapp` is run, **Then** the conflict section reflects the current file state (not stale data from the last explicit scan).
2. **Given** a repo with many worktrees, **When** `agit status` (no args) is run, **Then** all registered repos are displayed with their worktrees, tasks, and conflict summaries.

---

### Edge Cases

- What happens when the MCP server receives a tool call for a non-existent repo? The server returns a structured error with a clear message (not a crash or stack trace).
- What happens when two MCP clients call `agit_claim_task` for the same task simultaneously? SQLite's atomic UPDATE with WHERE clause ensures only one succeeds; the other receives a "task already claimed" error.
- What happens when `agit serve --transport sse` is started but the port is already in use? The server returns a clear error message indicating the port conflict.
- How does the MCP server handle a malformed tool input (missing required parameters)? The server validates inputs against the MCP tool schema and returns an error with the expected parameter format.
- What happens when `agit_spawn_worktree` is called but the disk is full or the path is invalid? The tool returns an error and no partial state is left in the registry (transactional cleanup).
- What happens when `agit agents --sweep` is run and a stale agent has an active worktree? The agent status is updated to "disconnected" but the worktree remains active — it is not automatically deleted.

## Clarifications

### Session 2026-02-20

- Q: Should the MCP server require authentication for SSE connections? → A: No auth; SSE binds to localhost only (127.0.0.1). agit is a single-machine developer tool; the threat model does not require network auth.
- Q: When `agit agents --remove` deletes an agent, what happens to their tasks and worktrees? → A: Soft release — unclaim tasks (revert to "pending"), unassign worktrees (clear agent_id), keep both intact.
- Q: Should `agit_merge_worktree` auto-cleanup after successful merge? → A: Yes — remove worktree from disk and mark "completed" in registry after successful merge.

## Requirements *(mandatory)*

### Functional Requirements

**Phase 7 - MCP Server (New)**

- **FR-001**: System MUST expose an MCP server via `agit serve` supporting `stdio` transport by default.
- **FR-002**: System MUST support SSE transport via `agit serve --transport sse --port <N>`, binding to localhost (127.0.0.1) only. No authentication is required.
- **FR-003**: Server MUST expose the following 11 MCP tools: `agit_list_repos`, `agit_repo_status`, `agit_spawn_worktree`, `agit_remove_worktree`, `agit_check_conflicts`, `agit_list_tasks`, `agit_claim_task`, `agit_complete_task`, `agit_merge_worktree`, `agit_register_agent`, `agit_heartbeat`.
- **FR-004**: Server MUST expose MCP resources: `agit://repos`, `agit://repos/{name}`, `agit://repos/{name}/conflicts`, `agit://repos/{name}/tasks`, `agit://agents`.
- **FR-005**: Each MCP tool MUST validate its input parameters against a defined schema and return structured errors for invalid inputs.
- **FR-006**: Tool `agit_spawn_worktree` MUST accept parameters: `repo` (required), `task` (optional), `branch` (optional), `agent` (optional) — matching the CLI `spawn` command semantics.
- **FR-007**: Tool `agit_claim_task` MUST accept `task_id` and `agent_id` and perform an atomic claim (only succeeds if task is in "pending" status).
- **FR-008**: Tool `agit_complete_task` MUST accept `task_id` and optional `result` string.
- **FR-009**: Tool `agit_merge_worktree` MUST perform a no-fast-forward merge of the worktree branch into the repo's default branch, with a pre-merge conflict check. After a successful merge, the worktree MUST be removed from disk and marked "completed" in the registry.
- **FR-010**: Tool `agit_register_agent` MUST accept `name` and `type` and return the agent ID.
- **FR-011**: Tool `agit_heartbeat` MUST accept `agent_id` and update the agent's `last_seen` timestamp.
- **FR-012**: All MCP tool responses MUST be JSON-serializable objects (not raw strings).
- **FR-013**: Resource `agit://repos/{name}` MUST include worktree count, task count, and active agent count in its response.

**Phase 4 - Task Coordination (Gap Fix)**

- **FR-014**: CLI `tasks` command MUST support `--fail <id>` flag to mark a task as failed.
- **FR-015**: CLI `tasks` command MUST support `--result <text>` flag to attach a result string when completing or failing a task.

**Phase 5 - Agent Registration (Gap Fix)**

- **FR-016**: System MUST provide an `agit agents` subcommand that lists all registered agents in a table.
- **FR-017**: `agit agents --sweep` MUST mark agents as "disconnected" if their `last_seen` exceeds the configured `stale_after` duration.
- **FR-018**: `agit agents --remove <name>` MUST delete an agent record from the registry. Before deletion, any claimed/in-progress tasks assigned to the agent MUST be reverted to "pending" status, and any worktrees assigned to the agent MUST have their agent_id cleared (worktrees and tasks are preserved, not deleted).

**Phase 1 - Core Registry (Existing)**

- **FR-019**: System MUST store data in a SQLite database at `~/.agit/agit.db` with WAL mode and foreign keys enabled.
- **FR-020**: Schema MUST include 5 tables: `repos`, `worktrees`, `agents`, `tasks`, `file_touches`.
- **FR-021**: `agit init` MUST create the `~/.agit/` directory, `config.toml`, and the SQLite database with schema.
- **FR-022**: `agit add <path>` MUST auto-detect repo name, remote URL, and default branch from the Git repository.
- **FR-023**: `agit repos` MUST list all registered repositories with worktree and task counts.

**Phase 2 - Worktree Management (Existing)**

- **FR-024**: `agit spawn <repo>` MUST create a Git worktree via `git worktree add -b` and record it in the registry.
- **FR-025**: Worktree branch names MUST follow the pattern `<prefix><slug>-<shortid>` where prefix defaults to `agit/`.
- **FR-026**: `agit cleanup` MUST remove completed or stale worktrees from both disk and registry.

**Phase 3 - Conflict Detection (Existing)**

- **FR-027**: `agit conflicts <repo>` MUST scan all active worktrees via `git diff --name-status` and detect file overlaps.
- **FR-028**: Conflict reports MUST show each conflicting file and the list of worktrees modifying it.

**Phase 6 - CLI Polish (Existing + Gap)**

- **FR-029**: `agit status` MUST display worktrees, tasks, agents, and conflicts for all repos (or a specific repo if named).
- **FR-030**: `agit merge <repo> <worktree-id>` MUST perform a no-fast-forward merge with pre-merge conflict check.

### Key Entities

- **Repo**: A registered Git repository — has a name, local path, remote URL, default branch, and associated worktrees/tasks.
- **Worktree**: An isolated Git worktree created for an agent — has a path, branch name, status (active/completed/stale/conflict), optional agent assignment and task description.
- **Agent**: A registered AI agent — has a name, type, status (active/idle/disconnected), heartbeat timestamp, and optional current worktree.
- **Task**: A work item assigned to an agent — has a description, lifecycle status (pending/claimed/in_progress/completed/failed), optional result, and timestamps.
- **FileTouch**: A record of a file modification in a worktree — used for cross-worktree conflict detection. Composite key of (repo_id, worktree_id, file_path).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: An MCP client can connect to `agit serve` (stdio) and successfully call all 11 tools, receiving valid JSON responses.
- **SC-002**: An MCP client can read all 5 resource URIs and receive valid JSON state snapshots.
- **SC-003**: `agit serve --transport sse --port 3847` accepts HTTP connections and processes MCP messages over SSE.
- **SC-004**: `agit tasks <repo> --fail <id> --result "error msg"` transitions the task to "failed" status with the result stored.
- **SC-005**: `agit agents` lists all agents with their status, and `--sweep` marks stale agents as "disconnected".
- **SC-006**: All existing CLI commands (`init`, `add`, `repos`, `spawn`, `cleanup`, `conflicts`, `status`, `merge`, `tasks`) continue to function correctly after MCP server integration (no regressions).
- **SC-007**: The MCP server handles concurrent tool calls without data corruption (SQLite WAL mode ensures this).
- **SC-008**: All MCP tool errors return structured JSON error responses (not panics or unformatted strings).
- **SC-009**: `go build ./...` succeeds with zero compilation errors after all changes.
- **SC-010**: `go vet ./...` reports zero issues.
