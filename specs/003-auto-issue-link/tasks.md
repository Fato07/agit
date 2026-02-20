# Tasks: Automatic Bug Report Issue Link

**Input**: Design documents from `/specs/003-auto-issue-link/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/file-manifest.md, quickstart.md

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create the two new packages that all user stories depend on

- [x] T001 [P] Create `UserError` type with `IsUserError()` and constructors in `internal/errors/errors.go`
- [x] T002 [P] Create issue link builder with `Build()`, `Enabled()`, URL construction, and truncation in `internal/issuelink/issuelink.go`

**Checkpoint**: Both new packages compile (`go build ./...`). No existing code depends on them yet.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Wire the issue link into the central CLI error handler. MUST complete before user story migration work.

**CRITICAL**: No user story work can begin until this phase is complete.

- [x] T003 Modify `Execute()` in `cmd/root.go` to use `issuelink.Build()` for internal errors and suppress issue links for `UserError` via `IsUserError()` check. Set `rootCmd.SilenceErrors = true`.
- [x] T004 Add panic recovery with `defer/recover` in `Execute()` in `cmd/root.go` that captures stack trace via `runtime/debug.Stack()` and displays issue link before exiting with code 2.

**Checkpoint**: Foundation ready. `go build ./...` passes. Internal errors (e.g., forced DB error) show issue link. Panics show crash message with issue link. User errors still show the old way (no UserError migration yet).

---

## Phase 3: User Story 1 - CLI Error Produces Issue Link (Priority: P1) MVP

**Goal**: Every unexpected internal CLI error displays a pre-filled GitHub issue URL. User input validation errors display only the error message.

**Independent Test**: Trigger an internal error (e.g., corrupt DB path) and verify issue link appears. Trigger a user input error (e.g., `agit add /nonexistent`) and verify NO issue link appears.

### Implementation for User Story 1

- [x] T005 [P] [US1] Migrate user input validation errors to `UserError` in `cmd/add.go`
- [x] T006 [P] [US1] Migrate user input validation errors to `UserError` in `cmd/spawn.go` (no user validation errors found — cobra.ExactArgs handles arg validation)
- [x] T007 [P] [US1] Migrate user input validation errors to `UserError` in `cmd/merge.go`
- [x] T008 [P] [US1] Migrate user input validation errors to `UserError` in `cmd/conflicts.go` (no user validation errors found — all errors are internal)
- [x] T009 [P] [US1] Migrate user input validation errors to `UserError` in `cmd/repos.go` (no user validation errors found — cobra.ExactArgs handles arg validation)
- [x] T010 [P] [US1] Migrate user input validation errors to `UserError` in `cmd/tasks.go`
- [x] T011 [P] [US1] Migrate user input validation errors to `UserError` in `cmd/cleanup.go` (no user validation errors found — all errors are internal)
- [x] T012 [P] [US1] Migrate user input validation errors to `UserError` in `cmd/serve.go`
- [x] T013 [US1] Verify `go build ./...` and `go vet ./...` pass after all cmd/ migrations

**Checkpoint**: User Story 1 fully functional. Internal CLI errors show issue link. User input errors do not. All cmd/ files use `UserError` for validation.

---

## Phase 4: User Story 2 - MCP Server Error Produces Issue Link (Priority: P2)

**Goal**: MCP tool error responses include a pre-filled GitHub issue URL for internal errors. Validation errors (e.g., "parameter is required") do not.

**Independent Test**: Call an MCP tool with invalid parameters — verify no issue link. Force an internal MCP error (e.g., DB failure during tool call) — verify issue link appears in error text.

### Implementation for User Story 2

- [x] T014 [US2] Migrate "parameter is required" validation errors to `UserError` in `internal/mcp/tools.go`
- [x] T015 [US2] Wrap internal MCP tool errors with issue link URL appended to error message in `internal/mcp/tools.go`
- [x] T016 [US2] Verify `go build ./...` and `go vet ./...` pass after MCP changes

**Checkpoint**: User Story 2 fully functional. MCP internal errors include issue link. MCP validation errors do not.

---

## Phase 5: User Story 3 - Opt-Out of Issue Links (Priority: P3)

**Goal**: Users can suppress issue links via `AGIT_NO_ISSUE_LINK=1` environment variable.

**Independent Test**: Set `AGIT_NO_ISSUE_LINK=1`, trigger an internal error — verify no issue link in output. Unset the variable, trigger the same error — verify issue link appears.

### Implementation for User Story 3

- [x] T017 [US3] Verify opt-out is respected in CLI error path in `cmd/root.go` (both normal errors and panics check `issuelink.Enabled()`)
- [x] T018 [US3] Verify opt-out is respected in MCP error wrapping in `internal/mcp/tools.go` (skip issue link when `issuelink.Enabled()` returns false)

**Checkpoint**: Opt-out works across both CLI and MCP interfaces.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Tests, build verification, and final validation

- [x] T019 [P] Create table-driven tests for `UserError` and `IsUserError()` in `internal/errors/errors_test.go`
- [x] T020 [P] Create table-driven tests for `Build()`, URL encoding, truncation, and opt-out in `internal/issuelink/issuelink_test.go`
- [x] T021 Run full test suite: `go test ./... -race`
- [x] T022 Run linter: `golangci-lint run`
- [x] T023 Run quickstart.md verification checklist (manual smoke tests)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately. T001 and T002 are parallel.
- **Foundational (Phase 2)**: Depends on Phase 1. T003 depends on T001 + T002. T004 depends on T003.
- **User Story 1 (Phase 3)**: Depends on Phase 2. All T005–T012 are parallel (different files). T013 depends on T005–T012.
- **User Story 2 (Phase 4)**: Depends on Phase 2 only. Can run in parallel with US1 (different files).
- **User Story 3 (Phase 5)**: Depends on Phase 3 + Phase 4 (verifies opt-out across both paths).
- **Polish (Phase 6)**: T019–T020 can start after Phase 1. T021–T023 depend on all phases complete.

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) — no other story dependencies
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) — independent of US1 (different files)
- **User Story 3 (P3)**: Depends on US1 + US2 being implemented (verifies opt-out in both code paths)

### Parallel Opportunities

- T001 + T002 (Phase 1): different packages, no dependencies
- T005 through T012 (Phase 3): all modify different cmd/ files
- T019 + T020 (Phase 6): different test files
- US1 (Phase 3) + US2 (Phase 4): can run in parallel after Phase 2

---

## Parallel Example: User Story 1

```bash
# Launch all cmd/ migrations together (all [P] tasks):
Task: "Migrate user input validation errors to UserError in cmd/add.go"
Task: "Migrate user input validation errors to UserError in cmd/spawn.go"
Task: "Migrate user input validation errors to UserError in cmd/merge.go"
Task: "Migrate user input validation errors to UserError in cmd/conflicts.go"
Task: "Migrate user input validation errors to UserError in cmd/repos.go"
Task: "Migrate user input validation errors to UserError in cmd/tasks.go"
Task: "Migrate user input validation errors to UserError in cmd/cleanup.go"
Task: "Migrate user input validation errors to UserError in cmd/serve.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001–T002)
2. Complete Phase 2: Foundational (T003–T004)
3. Complete Phase 3: User Story 1 (T005–T013)
4. **STOP and VALIDATE**: Internal errors show issue link, user errors do not
5. Deploy/demo if ready

### Incremental Delivery

1. Setup + Foundational → Foundation ready
2. Add User Story 1 → CLI errors classified and issue-linked → MVP!
3. Add User Story 2 → MCP errors classified and issue-linked
4. Add User Story 3 → Opt-out verified across both interfaces
5. Polish → Tests, lint, smoke tests

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- All 8 cmd/ file migrations in Phase 3 are independent and parallelizable
- Tests (Phase 6) are included because the spec references testable acceptance criteria
- Commit after each phase for clean bisectability
