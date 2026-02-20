# Tasks: agit Full Architecture Implementation

**Input**: Design documents from `/specs/001-agit-architecture-impl/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: No test tasks included (not requested in feature specification).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Single project**: Go CLI at repository root
- `cmd/` — CLI commands (cobra)
- `internal/` — Business logic packages
- All paths relative to `/Users/fathindosunmu/DEV/MyProjects/agit/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create new package directories and verify build

- [x] T001 Create `internal/mcp/` directory for MCP server package
- [x] T002 Verify `go build ./...` compiles cleanly before any changes

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Add new registry methods that MUST exist before any user story can be implemented

**CRITICAL**: No user story work can begin until this phase is complete

- [x] T003 Add `SweepStaleAgents(staleAfter time.Duration) (int, error)` method to `internal/registry/agents.go` — UPDATE agents SET status='disconnected' WHERE status='active' AND last_seen < now - staleAfter, return count of swept agents
- [x] T004 Add `UnclaimAgentTasks(agentID string) error` method to `internal/registry/agents.go` — UPDATE tasks SET status='pending', assigned_agent_id=NULL WHERE assigned_agent_id=? AND status IN ('claimed','in_progress')
- [x] T005 Add `UnassignAgentWorktrees(agentID string) error` method to `internal/registry/agents.go` — UPDATE worktrees SET agent_id=NULL WHERE agent_id=?
- [x] T006 Add `RemoveAgent(name string) error` method to `internal/registry/agents.go` — transaction: GetAgentByName, UnclaimAgentTasks, UnassignAgentWorktrees, DELETE FROM agents WHERE id=?

**Checkpoint**: Foundation ready — all new registry methods available for CLI and MCP tool handlers

---

## Phase 3: User Story 1 — MCP Server Exposes Tools for AI Agent Orchestration (Priority: P1) MVP

**Goal**: AI agents connect to `agit serve` via MCP and call 11 tools to manage repos, worktrees, tasks, and agents programmatically.

**Independent Test**: Start `agit serve`, pipe a JSON-RPC `initialize` + `tools/call` message for `agit_list_repos` through stdin, verify valid JSON response.

### Implementation for User Story 1

- [x] T007 [US1] Create MCP server factory in `internal/mcp/server.go` — implement `NewServer(db *registry.DB, cfg *config.Config) *server.MCPServer` that creates `server.NewMCPServer("agit", version, server.WithResourceCapabilities(true, true))` and registers all tools
- [x] T008 [US1] Implement `agit_list_repos` tool handler in `internal/mcp/tools.go` — call `db.ListRepos()`, marshal to JSON, return via `mcp.NewToolResultText()`
- [x] T009 [P] [US1] Implement `agit_repo_status` tool handler in `internal/mcp/tools.go` — accept `repo` string param, call `db.GetRepo()`, `db.ListWorktrees()`, `db.ListTasks()`, `db.FindConflicts()`, marshal combined status to JSON
- [x] T010 [P] [US1] Implement `agit_spawn_worktree` tool handler in `internal/mcp/tools.go` — accept `repo` (required), `task`, `branch`, `agent` (optional) params; replicate `cmd/spawn.go` logic: resolve repo, generate branch, create git worktree, record in registry, return JSON with worktree_id, path, branch
- [x] T011 [P] [US1] Implement `agit_remove_worktree` tool handler in `internal/mcp/tools.go` — accept `repo` and `worktree_id` params; call `git.RemoveWorktree()` then `db.DeleteWorktree()`
- [x] T012 [P] [US1] Implement `agit_check_conflicts` tool handler in `internal/mcp/tools.go` — accept `repo` param; call `conflicts.Detect(db, repo)`, marshal result to JSON
- [x] T013 [P] [US1] Implement `agit_list_tasks` tool handler in `internal/mcp/tools.go` — accept `repo` (required) and `status` (optional) params; call `db.ListTasks()`, marshal to JSON
- [x] T014 [P] [US1] Implement `agit_claim_task` tool handler in `internal/mcp/tools.go` — accept `task_id` and `agent_id` required params; call `db.ClaimTask()`, return JSON confirmation or error
- [x] T015 [P] [US1] Implement `agit_complete_task` tool handler in `internal/mcp/tools.go` — accept `task_id` (required) and `result` (optional) params; call `db.CompleteTask()`, return JSON confirmation
- [x] T016 [P] [US1] Implement `agit_merge_worktree` tool handler in `internal/mcp/tools.go` — accept `repo` and `worktree_id` params; call `git.CanMergeCleanly()`, `git.CheckoutBranch()`, `git.MergeBranch()`, then auto-cleanup: `git.RemoveWorktree()`, `db.UpdateWorktreeStatus("completed")`, return JSON
- [x] T017 [P] [US1] Implement `agit_register_agent` tool handler in `internal/mcp/tools.go` — accept `name` and `type` required params; call `db.RegisterAgent()`, return JSON with agent_id
- [x] T018 [P] [US1] Implement `agit_heartbeat` tool handler in `internal/mcp/tools.go` — accept `agent_id` required param; call `db.Heartbeat()`, return JSON confirmation
- [x] T019 [US1] Register all 11 tools in `internal/mcp/server.go` — call `s.AddTool()` for each tool with `mcp.NewTool()` definitions including parameter schemas (WithString, WithDescription, Required)
- [x] T020 [US1] Wire stdio transport in `cmd/serve.go` — replace TODO block: import `internal/mcp`, call `mcpserver.NewServer(db, cfg)`, then `server.ServeStdio(s)` for stdio transport
- [x] T021 [US1] Wire SSE transport in `cmd/serve.go` — when `--transport sse`: create `server.NewSSEServer(s, server.WithBaseURL(...))`, call `.Start(fmt.Sprintf("127.0.0.1:%d", port))`
- [x] T022 [US1] Verify `go build ./...` compiles with MCP server changes

**Checkpoint**: `agit serve` accepts MCP connections and all 11 tools are callable via stdio and SSE

---

## Phase 4: User Story 2 — MCP Resources Provide Read-Only State Discovery (Priority: P1)

**Goal**: AI agents read MCP resources (`agit://repos`, `agit://repos/{name}`, etc.) for state discovery without side effects.

**Independent Test**: Connect to MCP server, read `agit://repos` resource, verify JSON listing.

### Implementation for User Story 2

- [x] T023 [P] [US2] Implement `agit://repos` static resource handler in `internal/mcp/resources.go` — call `db.ListRepos()`, enrich with worktree/task/agent counts, return as `mcp.TextResourceContents` with JSON
- [x] T024 [P] [US2] Implement `agit://repos/{name}` template resource handler in `internal/mcp/resources.go` — extract repo name from URI, call `db.GetRepo()`, `db.ListWorktrees()`, `db.ListTasks()`, return detailed JSON
- [x] T025 [P] [US2] Implement `agit://repos/{name}/conflicts` template resource handler in `internal/mcp/resources.go` — extract repo name, call `conflicts.Detect()`, return JSON conflict report
- [x] T026 [P] [US2] Implement `agit://repos/{name}/tasks` template resource handler in `internal/mcp/resources.go` — extract repo name, call `db.ListTasks()`, enrich with agent names, return JSON
- [x] T027 [P] [US2] Implement `agit://agents` static resource handler in `internal/mcp/resources.go` — call `db.ListAgents()`, enrich with current worktree info, return JSON
- [x] T028 [US2] Register all resources in `internal/mcp/server.go` — call `s.AddResource()` for static resources (`agit://repos`, `agit://agents`) and `s.AddResourceTemplate()` for templates (`agit://repos/{name}`, `agit://repos/{name}/conflicts`, `agit://repos/{name}/tasks`)
- [x] T029 [US2] Verify `go build ./...` compiles with resource handlers

**Checkpoint**: All 5 MCP resources return valid JSON state snapshots

---

## Phase 5: User Story 3 — Task Fail and Result Flags in CLI (Priority: P2)

**Goal**: CLI supports `--fail` and `--result` flags for complete task lifecycle management.

**Independent Test**: Create a task, claim it, run `agit tasks <repo> --fail <id> --result "error"`, verify task status is "failed" with result stored.

### Implementation for User Story 3

- [x] T030 [US3] Add `--fail` string flag to `cmd/tasks.go` — register flag in `init()`, add handler block: if fail != "", call `db.FailTask(fail, resultPtr)`, print confirmation
- [x] T031 [US3] Add `--result` string flag to `cmd/tasks.go` — register flag in `init()`, wire into both `--complete` and `--fail` paths: pass `&result` to `db.CompleteTask()` and `db.FailTask()`
- [x] T032 [US3] Verify `go build ./...` compiles with task flag changes

**Checkpoint**: `agit tasks <repo> --fail <id> --result "msg"` works end-to-end

---

## Phase 6: User Story 4 — Agent Management CLI Subcommand (Priority: P2)

**Goal**: Users manage agent lifecycle via `agit agents` with list, sweep, and remove operations.

**Independent Test**: Register agents via `spawn --agent`, run `agit agents`, verify table output with agent details.

### Implementation for User Story 4

- [x] T033 [US4] Create `cmd/agents.go` — implement `agentsCmd` cobra command with `Use: "agents"`, `Short: "List and manage registered agents"`, default action: list all agents in a table (ID truncated to 12 chars, Name, Type, Status, Last Seen, Current Worktree)
- [x] T034 [US4] Add `--sweep` flag to `cmd/agents.go` — load config, call `db.SweepStaleAgents(cfg.Agent.StaleAfter)`, print count of agents marked disconnected
- [x] T035 [US4] Add `--remove <name>` flag to `cmd/agents.go` — call `db.RemoveAgent(name)`, print confirmation; handle "agent not found" error gracefully
- [x] T036 [US4] Register `agentsCmd` in `cmd/agents.go` `init()` — `rootCmd.AddCommand(agentsCmd)`
- [x] T037 [US4] Verify `go build ./...` compiles with agents command

**Checkpoint**: `agit agents`, `agit agents --sweep`, and `agit agents --remove <name>` all work

---

## Phase 7: User Story 8 — Status Dashboard with Live Conflict Scanning (Priority: P3)

**Goal**: `agit status` runs a fresh conflict scan before displaying results.

**Independent Test**: Modify files in a worktree, run `agit status myapp`, verify fresh conflicts appear.

### Implementation for User Story 8

- [x] T038 [US8] Add live conflict scanning to `cmd/status.go` — import `conflicts` package, before `db.FindConflicts(repo.ID)` call `conflicts.ScanAndUpdate(db, repo)` to refresh file touches for each repo in the loop
- [x] T039 [US8] Verify `go build ./...` compiles with status changes

**Checkpoint**: `agit status` shows fresh conflict data

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Final verification and quality checks across all changes

- [x] T040 Run `go build ./...` — verify zero compilation errors
- [x] T041 Run `go vet ./...` — verify zero issues
- [x] T042 Manual smoke test: `agit init` → `agit add <repo>` → `agit spawn` → `agit tasks` → `agit agents` → `agit status` → `agit conflicts` → verify all commands work
- [x] T043 Manual MCP test: pipe JSON-RPC initialize + `agit_list_repos` tool call through `agit serve` stdin, verify valid response
- [x] T044 Verify `agit serve --transport sse --port 3847` starts and accepts connections on 127.0.0.1

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Setup — BLOCKS all user stories
- **US1 MCP Tools (Phase 3)**: Depends on Foundational
- **US2 MCP Resources (Phase 4)**: Depends on Foundational; can run in parallel with US1
- **US3 Task Flags (Phase 5)**: Depends on Foundational; can run in parallel with US1/US2
- **US4 Agents CLI (Phase 6)**: Depends on Foundational; can run in parallel with US1/US2/US3
- **US8 Status Scan (Phase 7)**: Depends on Foundational; can run in parallel with all above
- **Polish (Phase 8)**: Depends on all user stories being complete

### User Story Dependencies

- **US1 (P1 — MCP Tools)**: After Foundational — no dependency on other stories
- **US2 (P1 — MCP Resources)**: After Foundational — no dependency on other stories (shares `internal/mcp/` package with US1 but different files: `tools.go` vs `resources.go`)
- **US3 (P2 — Task Flags)**: After Foundational — no dependency on other stories
- **US4 (P2 — Agents CLI)**: After Foundational — no dependency on other stories
- **US8 (P3 — Status Scan)**: After Foundational — no dependency on other stories

### Within Each User Story

- Tool/resource handlers before server registration
- Server registration before CLI wiring
- CLI wiring before build verification

### Parallel Opportunities

- T003, T004, T005 can run in parallel (independent registry methods)
- T008 through T018 can run in parallel (independent tool handlers in same file, but [P] only for those with no shared state)
- T023 through T027 can run in parallel (independent resource handlers)
- US1, US2, US3, US4, US8 can all start in parallel after Foundational phase

---

## Parallel Example: User Story 1

```bash
# After T007 (server factory), launch tool handlers in parallel:
Task: "Implement agit_list_repos tool handler in internal/mcp/tools.go"
Task: "Implement agit_repo_status tool handler in internal/mcp/tools.go"
Task: "Implement agit_spawn_worktree tool handler in internal/mcp/tools.go"
Task: "Implement agit_check_conflicts tool handler in internal/mcp/tools.go"
Task: "Implement agit_list_tasks tool handler in internal/mcp/tools.go"
Task: "Implement agit_claim_task tool handler in internal/mcp/tools.go"
Task: "Implement agit_complete_task tool handler in internal/mcp/tools.go"
Task: "Implement agit_merge_worktree tool handler in internal/mcp/tools.go"
Task: "Implement agit_register_agent tool handler in internal/mcp/tools.go"
Task: "Implement agit_heartbeat tool handler in internal/mcp/tools.go"

# Then register all tools and wire CLI:
Task: "Register all 11 tools in internal/mcp/server.go"
Task: "Wire stdio transport in cmd/serve.go"
Task: "Wire SSE transport in cmd/serve.go"
```

## Parallel Example: All User Stories After Foundational

```bash
# After Phase 2 completes, launch all stories in parallel:
Agent 1: US1 (MCP Tools) — internal/mcp/tools.go + cmd/serve.go
Agent 2: US2 (MCP Resources) — internal/mcp/resources.go
Agent 3: US3 (Task Flags) — cmd/tasks.go
Agent 4: US4 (Agents CLI) — cmd/agents.go
Agent 5: US8 (Status Scan) — cmd/status.go
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL — blocks all stories)
3. Complete Phase 3: User Story 1 (MCP Tools)
4. **STOP and VALIDATE**: Pipe JSON-RPC through `agit serve` and verify tool responses
5. This delivers the core MCP server with all 11 tools

### Incremental Delivery

1. Setup + Foundational → Foundation ready
2. Add US1 (MCP Tools) → Test via stdio → MVP!
3. Add US2 (MCP Resources) → Test resource reads → Full MCP surface
4. Add US3 (Task Flags) → Test `--fail`/`--result` → Complete task lifecycle
5. Add US4 (Agents CLI) → Test `agit agents` → Complete agent management
6. Add US8 (Status Scan) → Test live conflicts → Polish complete
7. Each story adds value without breaking previous stories

---

## Notes

- [P] tasks = different files or independent functions, no dependencies
- [Story] label maps task to specific user story for traceability
- All tool handlers in `internal/mcp/tools.go` are independent functions — can be written in parallel
- All resource handlers in `internal/mcp/resources.go` are independent functions — can be written in parallel
- US3, US4, US8 each touch a single different file — fully parallelizable
- Total: 44 tasks across 8 phases
