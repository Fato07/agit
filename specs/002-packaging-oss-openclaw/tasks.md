# Tasks: Packaging, OSS Setup & OpenClaw Integration

**Input**: Design documents from `/specs/002-packaging-oss-openclaw/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: No test tasks — feature is configuration/documentation only; no new Go source code.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Directory structure and prerequisite changes needed by all user stories

- [x] T001 Create `.github/workflows/` directory structure at `.github/workflows/`
- [x] T002 [P] Create `.github/ISSUE_TEMPLATE/` directory structure at `.github/ISSUE_TEMPLATE/`
- [x] T003 [P] Create `integrations/openclaw/` directory structure at `integrations/openclaw/`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: The .gitignore update was already done in spec 001 commit. No additional foundational blockers.

**⚠️ CRITICAL**: .gitignore already updated (committed with spec 001). All user stories can begin after directory setup.

**Checkpoint**: Foundation ready — user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Install agit from a release binary (Priority: P1) 🎯 MVP

**Goal**: Produce distributable binaries for 5 platforms via GoReleaser, with a Homebrew formula and curl install script.

**Independent Test**: Run `goreleaser release --snapshot --clean` and verify 5 platform archives produced. Run `bash -n install.sh` for script syntax validation.

### Implementation for User Story 1

- [x] T004 [US1] Create GoReleaser v2 configuration at `.goreleaser.yml` with builds for linux/darwin/windows on amd64/arm64, CGO_ENABLED=0, ldflags for version injection, tar.gz/zip archives, sha256 checksums, Homebrew formula push to fathindos/homebrew-tap, and deb/rpm nfpms (exact config from `docs/002-packaging-oss-openclaw.md` Section 1.1)
- [x] T005 [P] [US1] Create release workflow at `.github/workflows/release.yml` triggered on `v*` tags, with steps: checkout, setup Go 1.23, `go test ./...`, `goreleaser release --clean` using GITHUB_TOKEN and HOMEBREW_TAP_TOKEN secrets
- [x] T006 [P] [US1] Create install script at `install.sh` with: `set -euo pipefail`, OS/arch detection via `uname`, GitHub API latest release fetch, archive download with curl (wget fallback), sha256 checksum verification via `sha256sum` or `shasum -a 256`, install to `/usr/local/bin` with `~/.local/bin` fallback, colorized output with tty detection, error messages for unsupported OS/arch
- [x] T007 [US1] Update Makefile at `Makefile` to add `release-snapshot` target: `goreleaser release --snapshot --clean` and add `release-snapshot` to `.PHONY`

**Checkpoint**: GoReleaser config + release workflow + install script complete. Binary distribution is functional.

---

## Phase 4: User Story 2 - Contribute to agit as an open source project (Priority: P2)

**Goal**: Provide all OSS scaffolding so contributors find clear guidelines, templates, CI, and licensing.

**Independent Test**: Verify all OSS files exist, CI workflow YAML is valid, linter config is parseable, and README displays badges.

### Implementation for User Story 2

- [x] T008 [P] [US2] Create MIT license file at `LICENSE` with year 2026, copyright holder "fathindos"
- [x] T009 [P] [US2] Create contribution guidelines at `CONTRIBUTING.md` with sections: Welcome (project description, link to architecture spec, link to `good first issue` label), Development Setup (`git clone`, `go mod tidy`, `make build`, `./agit init`, `./agit add`, `./agit repos`), Project Structure (brief description of cmd/, internal/registry/, internal/git/, internal/mcp/), Making Changes (fork, branch, test, lint, conventional commit, PR), Code Style (go fmt, go vet, golangci-lint, 50-line function guideline, exported function comments), Commit Messages (conventional commits with examples), PR Guidelines (single concern, tests, docs, template), Reporting Issues (link templates, include version/OS/Go), Code of Conduct reference
- [x] T010 [P] [US2] Create code of conduct at `CODE_OF_CONDUCT.md` using Contributor Covenant v2.1 full text
- [x] T011 [P] [US2] Create bug report template at `.github/ISSUE_TEMPLATE/bug_report.md` with YAML frontmatter (name: Bug Report, about: Something isn't working, title prefix: "[BUG] ", labels: bug), sections: Describe the bug, To reproduce (steps), Expected behavior, Environment (agit version, OS, Go version, Git version), Additional context
- [x] T012 [P] [US2] Create feature request template at `.github/ISSUE_TEMPLATE/feature_request.md` with YAML frontmatter (name: Feature Request, about: Suggest an enhancement, title prefix: "[FEATURE] ", labels: enhancement), sections: Problem, Proposed solution, Alternatives considered, Additional context
- [x] T013 [P] [US2] Create MCP tool request template at `.github/ISSUE_TEMPLATE/mcp_tool_request.md` with YAML frontmatter (name: MCP Tool Request, about: Request a new MCP tool or resource, title prefix: "[MCP] ", labels: mcp enhancement), sections: Tool or Resource?, Name, Description, Parameters, Use case
- [x] T014 [P] [US2] Create pull request template at `.github/pull_request_template.md` with sections: What does this PR do?, Type of change (checkboxes: Bug fix, New feature, Breaking change, Documentation), How has this been tested?, Checklist (style guidelines, tests added, tests pass, docs updated)
- [x] T015 [P] [US2] Create CI workflow at `.github/workflows/ci.yml` triggered on push to main and all PRs, with jobs: test (matrix Go 1.22 + 1.23 on ubuntu-latest, steps: checkout, setup-go, `go mod tidy`, `go vet ./...`, `go test ./... -race -coverprofile=coverage.out`), lint (golangci-lint via official `golangci/golangci-lint-action`), build (`CGO_ENABLED=0 go build .`)
- [x] T016 [P] [US2] Create golangci-lint configuration at `.golangci.yml` with: `run.timeout: 5m`, linters enabled: errcheck, gosimple, govet, ineffassign, staticcheck, unused, gofmt, misspell, revive; linters-settings: revive rule `exported` with severity warning; issues: exclude errcheck on `_test\.go` files
- [x] T017 [US2] Update README at `README.md` to add badges after the `# agit` heading: CI badge (GitHub Actions ci.yml), Release badge (latest GitHub release), Go Report Card badge, License MIT badge (exact badge markdown from `docs/002-packaging-oss-openclaw.md` Section 2.8)

**Checkpoint**: All OSS files in place. Contributors have clear guidelines, CI validates PRs, linter catches issues.

---

## Phase 5: User Story 3 - Connect agit to an AI agent via MCP (Priority: P2)

**Goal**: Provide MCP integration documentation and an OpenClaw skill file so agent operators can connect agit.

**Independent Test**: Verify integration docs contain valid JSON config snippets for Claude Code, OpenClaw, Cursor, and generic clients. Verify skill file has correct YAML frontmatter and workflow description.

### Implementation for User Story 3

- [x] T018 [P] [US3] Create OpenClaw skill file at `integrations/openclaw/skill.md` with YAML frontmatter (name: agit, description, version: 0.1.0, mcp_servers: [agit]), sections: When to use agit (session start, before changes, before merging, task management), Workflow (register_agent → list_repos → list_tasks → claim_task/spawn_worktree → check_conflicts → complete_task → merge_worktree), Important notes (always use worktrees, check conflicts, complete tasks)
- [x] T019 [P] [US3] Create integration documentation at `docs/integrations.md` with MCP config snippets for: Claude Code (`~/.claude/mcp.json` with agit stdio config), OpenClaw (`openclaw.json` mcpServers section), Cursor (`.cursor/mcp.json`), generic MCP clients (description of stdio default + SSE transport with `--transport sse --port 3847`), verification steps (start session, call `agit_list_repos`, spawn worktree, check conflicts, cleanup), reference to architecture spec Section 5 for full tool/resource list

**Checkpoint**: All integration docs complete. Agent operators can connect agit to any MCP client.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final validation across all user stories

- [x] T020 Verify `go build ./...` still compiles after all changes
- [x] T021 [P] Verify `go vet ./...` still passes after all changes
- [x] T022 [P] Verify `bash -n install.sh` passes syntax validation
- [x] T023 Run quickstart.md validation against implemented files

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Already complete (.gitignore committed with spec 001)
- **User Story 1 (Phase 3)**: Depends on Phase 1 directory setup
- **User Story 2 (Phase 4)**: Depends on Phase 1 directory setup
- **User Story 3 (Phase 5)**: Depends on Phase 1 directory setup
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Phase 1 — No dependencies on other stories
- **User Story 2 (P2)**: Can start after Phase 1 — No dependencies on other stories
- **User Story 3 (P2)**: Can start after Phase 1 — No dependencies on other stories
- All three stories are fully independent and can run in parallel

### Within Each User Story

- T004 (goreleaser) before T007 (Makefile update references goreleaser)
- T005 (release workflow) and T006 (install script) can run in parallel with T004
- All US2 tasks (T008–T016) are fully parallel (different files)
- T017 (README badges) can run in parallel with other US2 tasks
- T018 and T019 are fully parallel (different files)

### Parallel Opportunities

- After Phase 1: US1, US2, US3 can all start simultaneously
- Within US1: T005 and T006 parallel; T004 then T007 sequential
- Within US2: ALL tasks T008–T017 are parallel (10 independent files)
- Within US3: T018 and T019 are parallel

---

## Parallel Example: All User Stories Simultaneously

```bash
# Phase 1: Setup directories (sequential, fast)
Task: "Create .github/workflows/ directory"
Task: "Create .github/ISSUE_TEMPLATE/ directory"
Task: "Create integrations/openclaw/ directory"

# Then launch ALL user stories in parallel:

# US1 batch:
Task: "T004 Create .goreleaser.yml"
Task: "T005 Create .github/workflows/release.yml"
Task: "T006 Create install.sh"

# US2 batch (all 10 tasks parallel):
Task: "T008 Create LICENSE"
Task: "T009 Create CONTRIBUTING.md"
Task: "T010 Create CODE_OF_CONDUCT.md"
Task: "T011-T013 Create issue templates"
Task: "T014 Create PR template"
Task: "T015 Create ci.yml"
Task: "T016 Create .golangci.yml"
Task: "T017 Update README badges"

# US3 batch:
Task: "T018 Create OpenClaw skill"
Task: "T019 Create integration docs"

# Then sequential:
Task: "T007 Update Makefile" (after T004)

# Final: Polish phase
Task: "T020-T023 Verification"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup directories
2. Complete Phase 3: User Story 1 (GoReleaser + release workflow + install script + Makefile)
3. **STOP and VALIDATE**: Run `goreleaser release --snapshot --clean`, `bash -n install.sh`
4. Binary distribution works — MVP complete

### Incremental Delivery

1. Setup directories → Foundation ready
2. Add User Story 1 → Binary distribution works (MVP!)
3. Add User Story 2 → OSS scaffolding complete → Ready for contributors
4. Add User Story 3 → Integration docs complete → Ready for agent operators
5. Polish → All verification checks pass

### Maximum Parallel Strategy

With the task structure above, the entire feature can be completed in 4 sequential steps:
1. Phase 1: Create 3 directories (fast)
2. Phase 3+4+5: All 16 file-creation tasks in parallel (T004–T019, minus T007)
3. T007: Update Makefile (depends on T004)
4. Phase 6: Run 4 verification checks
