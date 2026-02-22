# Tasks: v0.4.0 Release Hardening

**Input**: Design documents from `/specs/004-v04-release-hardening/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Included — FR-001 requires CLI integration tests, FR-012 requires all new features pass tests.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Test infrastructure and helper utilities needed across multiple stories

- [X] T001 Create CLI test helper with `executeCommand(args ...string) (stdout, stderr string, err error)` that sets up temp HOME, initializes agit, captures output via `rootCmd.SetOut/SetErr`, and calls `rootCmd.Execute()` in `cmd/test_helpers_test.go`
- [X] T002 [P] Verify existing `registry.OpenMemory()` works for CLI test isolation by writing a smoke test in `cmd/cmd_smoke_test.go` that runs `agit --version`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before user story features can be implemented

**No foundational tasks required** — the existing codebase already has the database schema, config system, MCP server, and CLI framework. T001/T002 provide the test infrastructure needed.

**Checkpoint**: Setup ready — user story implementation can begin

---

## Phase 3: User Story 1 — CLI Integration Tests Catch Regressions (Priority: P1) MVP

**Goal**: Every CLI command has at least one integration test covering the happy path; `cmd/` package test coverage reaches at least 60%.

**Independent Test**: Run `go test ./cmd/... -race -v` and verify all tests pass. Check coverage with `go test ./cmd/... -coverprofile=coverage.out && go tool cover -func=coverage.out | grep total`.

### Implementation for User Story 1

- [X] T003 [P] [US1] Write integration test for `agit init` command (success + already-initialized error) in `cmd/init_test.go`
- [X] T004 [P] [US1] Write integration test for `agit add` command (success with temp git repo + missing path error) in `cmd/add_test.go`
- [X] T005 [P] [US1] Write integration test for `agit repos` command (list empty + list after add + JSON output) in `cmd/repos_test.go`
- [X] T006 [P] [US1] Write integration test for `agit spawn` command (success with temp git repo + missing repo error) in `cmd/spawn_test.go`
- [X] T007 [P] [US1] Write integration test for `agit status` command (global status + per-repo status + JSON output) in `cmd/status_test.go`
- [X] T008 [P] [US1] Write integration test for `agit conflicts` command (no conflicts + JSON output) in `cmd/conflicts_test.go`
- [X] T009 [P] [US1] Write integration test for `agit tasks` command (create + list + claim + complete + fail + JSON output) in `cmd/tasks_test.go`
- [X] T010 [P] [US1] Write integration test for `agit agents` command (list empty + sweep + JSON output) in `cmd/agents_test.go`
- [X] T011 [P] [US1] Write integration test for `agit config` subcommands (show + set + path + reset + JSON output) in `cmd/config_test.go`
- [X] T012 [P] [US1] Write integration test for `agit merge` command (missing worktree-id error) in `cmd/merge_test.go`
- [X] T013 [P] [US1] Write integration test for `agit cleanup` command (nothing to clean + JSON output) in `cmd/cleanup_test.go`
- [X] T014 [P] [US1] Write integration test for `agit serve` command (verify it starts without error using context cancellation) in `cmd/serve_test.go`
- [X] T015 [P] [US1] Write integration test for `agit update` command (verify version check output + error handling for no network) in `cmd/update_test.go`
- [X] T016 [US1] Verify `cmd/` test coverage meets 60% target by running `go test ./cmd/... -coverprofile=coverage.out` and adjusting tests if needed

**Checkpoint**: All CLI commands have integration tests. `go test ./cmd/... -race` passes.

---

## Phase 4: User Story 2 — Version Bump and Release (Priority: P1)

**Goal**: Version bumped to 0.4.0, tagged release with binaries for all platforms.

**Independent Test**: Run `go run . --version` and verify output contains `0.4.0`.

### Implementation for User Story 2

- [X] T017 [US2] Update version constant from `0.3.0` to `0.4.0` in `cmd/root.go`
- [X] T018 [US2] Verify `go build ./...` succeeds and `./agit --version` prints `0.4.0`
- [X] T019 [US2] Verify GoReleaser config (`.goreleaser.yaml`) builds for all target platforms (linux/darwin amd64+arm64, windows/amd64)
- [X] T020 [US2] Update version test assertion in `cmd/cmd_smoke_test.go` (from T002) to expect `0.4.0`

**Checkpoint**: Version is 0.4.0. Build succeeds on all platforms.

---

## Phase 5: User Story 3 — Task Queue-Based Dispatch (Priority: P2)

**Goal**: Atomic `NextTask` dispatch returns and claims the highest-priority pending task. Exposed as CLI subcommand and MCP tool.

**Independent Test**: Create tasks with different priorities, call `agit tasks next`, verify highest-priority task is returned and atomically claimed.

### Implementation for User Story 3

- [X] T021 [US3] Add `NextTask(repoID, agentID string) (*Task, error)` method to `internal/registry/tasks.go` using atomic `UPDATE...WHERE id=(SELECT id FROM tasks WHERE repo_id=? AND status='pending' ORDER BY priority DESC, created_at ASC LIMIT 1)` then `SELECT` fallback
- [X] T022 [US3] Write unit tests for `NextTask` in `internal/registry/registry_test.go`: highest-priority returned, FIFO tiebreak on equal priority, no pending tasks returns nil, concurrent calls claim different tasks
- [X] T023 [US3] Add `tasksNextCmd` Cobra subcommand to `cmd/tasks.go` with `--agent` required flag, text + JSON output per CLI contract
- [X] T024 [US3] Add `handleNextTask` handler in `internal/mcp/tools.go` following existing pattern (param extraction, `db.NextTask()`, `jsonResult()`)
- [X] T025 [US3] Register `agit_next_task` tool in `internal/mcp/server.go` with `repo` and `agent_id` required params per MCP contract
- [X] T026 [US3] Write integration test for `agit tasks next` in `cmd/tasks_test.go` (append to existing): create 3 tasks with priorities 1/10/5, call next, verify priority-10 task claimed
- [X] T027 [P] [US3] Write MCP handler test for `handleNextTask` in `internal/mcp/tools_test.go`: success, no tasks, missing params
- [X] T028 [US3] Write benchmark test `BenchmarkNextTask` in `internal/registry/registry_test.go`: insert 1,000 tasks with varying priorities, verify `NextTask` completes in <100ms (SC-003)

**Checkpoint**: `agit tasks next myrepo --agent a1` returns highest-priority task. MCP tool `agit_next_task` works.

---

## Phase 6: User Story 4 — Conflict Resolution Guidance (Priority: P2)

**Goal**: `agit conflicts` output includes resolution suggestions with recommended merge order.

**Independent Test**: Create two worktrees with overlapping file touches, run `agit conflicts`, verify suggestions appear.

### Implementation for User Story 4

- [X] T029 [US4] Create `internal/conflicts/resolver.go` with `SuggestResolutionOrder(conflicts []registry.Conflict, worktrees []registry.Worktree) []Suggestion` function that sorts by (1) fewer conflicting files, (2) older creation time
- [X] T030 [US4] Define `Suggestion` struct in `internal/conflicts/resolver.go` with fields: `Order int`, `WorktreeID string`, `ConflictingFiles int`, `CreatedAt time.Time`, `Rationale string`
- [X] T031 [US4] Write unit tests for `SuggestResolutionOrder` in `internal/conflicts/resolver_test.go`: 2 worktrees, 3 worktrees, single worktree (no suggestions), equal conflicts sorted by creation time
- [X] T032 [US4] Modify `cmd/conflicts.go` to call `resolver.SuggestResolutionOrder()` after listing conflicts and print "Suggested resolution order:" section in text output
- [X] T033 [US4] Modify `cmd/conflicts.go` JSON output to include `suggestions` array per CLI contract
- [X] T034 [US4] Modify `handleCheckConflicts` in `internal/mcp/tools.go` to include `suggestions` in JSON response per MCP contract
- [X] T035 [US4] Write integration test for `agit conflicts` with suggestions in `cmd/conflicts_test.go` (append): set up overlapping file touches, verify suggestion output contains "Suggested resolution order"

**Checkpoint**: `agit conflicts myrepo` shows suggestions. JSON output includes `suggestions` array.

---

## Phase 7: User Story 5 — Plugin/Hook System (Priority: P3)

**Goal**: Event hooks configured in config.toml fire asynchronously on agit events with context via env vars.

**Independent Test**: Configure a hook that writes to a temp file, trigger the event, verify the file exists.

### Implementation for User Story 5

- [X] T036 [US5] Add `Hooks map[string]string` field and `HookTimeout string` field to config struct in `internal/config/config.go`, with default timeout `"30s"` (global timeout, per-event deferred to future version)
- [X] T037 [US5] Update `SetByDotKey`/`GetByDotKey`/`AllKeys` in `internal/config/config.go` to handle `hooks.*` keys (e.g., `hooks.worktree.created`)
- [X] T038 [US5] Create `internal/hooks/hooks.go` with `Runner` struct holding config, and `Fire(event string, env map[string]string)` method that spawns goroutine, runs `sh -c <command>` with timeout context and env vars (`AGIT_EVENT`, `AGIT_REPO`, `AGIT_TASK_ID`, `AGIT_AGENT_ID`, `AGIT_WORKTREE_ID`)
- [X] T039 [US5] Write unit tests for hook `Runner` in `internal/hooks/hooks_test.go`: hook fires and writes file, hook timeout kills process, no hook configured is no-op, failing hook logs warning but returns no error, and verify `Fire()` returns in <50ms (SC-005 — hooks are async so caller must not block)
- [X] T040 [US5] Integrate hook firing into `cmd/spawn.go` for `worktree.created` event after successful worktree creation
- [X] T041 [US5] Integrate hook firing into `cmd/tasks.go` for `task.claimed`, `task.completed`, `task.failed` events after respective operations
- [X] T042 [US5] Integrate hook firing into `cmd/cleanup.go` for `worktree.removed` event after worktree removal
- [X] T043 [US5] Integrate hook firing into `cmd/conflicts.go` for `conflict.detected` event when conflicts are found
- [X] T044 [US5] Write integration test for hooks in `cmd/hooks_test.go`: configure hook via config, run spawn command, verify hook output file exists

**Checkpoint**: Hooks fire on events. Failing hooks don't block operations. No hooks = no impact.

---

## Phase 8: User Story 6 — MCP SSE Transport Hardening (Priority: P3)

**Goal**: SSE server shuts down gracefully on SIGTERM/SIGINT within 5 seconds.

**Independent Test**: Start SSE server, send SIGTERM, verify clean shutdown without panics or leaked goroutines.

### Implementation for User Story 6

- [X] T045 [US6] Modify `cmd/serve.go` to set up `os/signal` notification for `SIGTERM` and `SIGINT`, create shutdown context with 5-second timeout, call `httpServer.Shutdown(ctx)` on signal receipt
- [X] T046 [US6] Add port-in-use error handling to `cmd/serve.go` SSE mode: detect `bind: address already in use` error and print user-friendly message
- [X] T047 [US6] Write test for graceful shutdown in `cmd/serve_test.go` (append): start SSE server in goroutine, send MCP tool call via HTTP, then send signal via `syscall.Kill`, verify tool call response received and clean exit within 5 seconds (SC-006)
- [X] T048 [P] [US6] Write test for port-in-use error in `cmd/serve_test.go` (append): bind a port, attempt SSE start on same port, verify error message

**Checkpoint**: SSE server handles SIGTERM/SIGINT gracefully. Port conflict shows clear error.

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: Final validation and cleanup

- [X] T049 Run `go build ./...` and verify zero errors
- [X] T050 Run `go test ./... -race` and verify zero failures across all packages
- [X] T051 Run `golangci-lint run` and fix any new lint warnings
- [X] T052 Verify `cmd/` package test coverage reaches 60% with `go test ./cmd/... -coverprofile=coverage.out && go tool cover -func=coverage.out`
- [X] T053 Update README.md to document `agit tasks next` subcommand and hook configuration in Commands table and Configuration section
- [X] T054 Run quickstart.md verification scenarios manually for US3 (task dispatch) and US4 (conflict suggestions)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: No separate phase needed — existing infrastructure sufficient
- **US1 (Phase 3)**: Depends on T001/T002 (test helpers). Independent of all other stories.
- **US2 (Phase 4)**: Independent. Can run in parallel with US1.
- **US3 (Phase 5)**: Independent. Adds to `registry/tasks.go`, `cmd/tasks.go`, `mcp/tools.go`.
- **US4 (Phase 6)**: Independent. Adds `conflicts/resolver.go`, modifies `cmd/conflicts.go`.
- **US5 (Phase 7)**: Independent new package. Integrates with multiple cmd files (spawn, tasks, cleanup, conflicts).
- **US6 (Phase 8)**: Independent. Modifies `cmd/serve.go` only.
- **Polish (Phase 9)**: Depends on all desired user stories being complete.

### User Story Dependencies

- **US1 (P1)**: Can start after Setup — no dependencies on other stories
- **US2 (P1)**: Can start immediately — no dependencies on other stories
- **US3 (P2)**: Can start after Setup — no dependencies on other stories
- **US4 (P2)**: Can start after Setup — no dependencies on other stories
- **US5 (P3)**: Can start after Setup — integrates with cmd files but creates new package first
- **US6 (P3)**: Can start after Setup — no dependencies on other stories

### Within Each User Story

- Core logic before CLI/MCP integration
- Unit tests alongside or after implementation
- Integration tests after CLI commands are wired

### Parallel Opportunities

- T003-T015 (US1 tests): All [P] — can run in parallel (different test files)
- T021-T025 vs T029-T034: US3 and US4 can be developed in parallel (different packages)
- T036-T038 vs T045-T046: US5 and US6 can be developed in parallel (different packages)
- T027 and T031: Test files are independent and parallelizable

---

## Parallel Example: User Story 1

```bash
# All CLI integration test files can be written in parallel:
Task: T003 "Write init test in cmd/init_test.go"
Task: T004 "Write add test in cmd/add_test.go"
Task: T005 "Write repos test in cmd/repos_test.go"
Task: T006 "Write spawn test in cmd/spawn_test.go"
Task: T007 "Write status test in cmd/status_test.go"
Task: T008 "Write conflicts test in cmd/conflicts_test.go"
Task: T009 "Write tasks test in cmd/tasks_test.go"
Task: T010 "Write agents test in cmd/agents_test.go"
Task: T011 "Write config test in cmd/config_test.go"
Task: T012 "Write merge test in cmd/merge_test.go"
Task: T013 "Write cleanup test in cmd/cleanup_test.go"
Task: T014 "Write serve test in cmd/serve_test.go"
Task: T015 "Write update test in cmd/update_test.go"
```

## Parallel Example: User Stories 3 & 4

```bash
# US3 and US4 can be developed in parallel (different packages):
# Agent A: US3 (registry/tasks.go + cmd/tasks.go + mcp/tools.go)
# Agent B: US4 (conflicts/resolver.go + cmd/conflicts.go)
```

---

## Implementation Strategy

### MVP First (User Stories 1 + 2 Only)

1. Complete Phase 1: Setup (T001-T002)
2. Complete Phase 3: US1 — CLI Integration Tests (T003-T016)
3. Complete Phase 4: US2 — Version Bump (T017-T020)
4. **STOP and VALIDATE**: Run `go test ./cmd/... -race` and `go build ./...`
5. Tag v0.4.0 and release

### Incremental Delivery

1. Setup → CLI Tests (US1) → Version Bump (US2) → **v0.4.0 MVP release**
2. Add Task Dispatch (US3) → Test independently → patch release
3. Add Conflict Resolution (US4) → Test independently → patch release
4. Add Hooks (US5) + SSE Hardening (US6) → Test → patch release

### Parallel Team Strategy

With multiple developers:

1. Complete Setup together (T001-T002)
2. Once setup is done:
   - Developer A: US1 (CLI tests) + US2 (version bump)
   - Developer B: US3 (task dispatch) + US4 (conflict resolution)
   - Developer C: US5 (hooks) + US6 (SSE hardening)
3. Stories complete and integrate independently

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story is independently completable and testable
- US1 is the MVP — CLI tests enable safe iteration on everything else
- US2 (version bump) should be the last thing merged before tagging
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
