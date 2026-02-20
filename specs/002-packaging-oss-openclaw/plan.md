# Implementation Plan: Packaging, OSS Setup & OpenClaw Integration

**Branch**: `002-packaging-oss-openclaw` | **Date**: 2026-02-20 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/002-packaging-oss-openclaw/spec.md`

## Summary

Prepare agit for public release by adding GoReleaser packaging (5 OS/arch targets), open source scaffolding (LICENSE, CONTRIBUTING, CI, templates), and MCP integration documentation for Claude Code, OpenClaw, Cursor, and generic MCP clients. All changes are new files or minor edits to existing files — no Go source code changes required.

## Technical Context

**Language/Version**: Go 1.23 (go.mod), CI tests against 1.22 + 1.23
**Primary Dependencies**: GoReleaser v2 (build/release), golangci-lint (linting), GitHub Actions (CI/CD)
**Storage**: N/A (no data model changes)
**Testing**: `go test ./...`, `go vet ./...`, `golangci-lint run`, `bash -n install.sh`
**Target Platform**: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64
**Project Type**: Single Go CLI binary
**Performance Goals**: N/A (packaging concern, not runtime)
**Constraints**: CGO_ENABLED=0 mandatory (pure Go SQLite via modernc.org/sqlite)
**Scale/Scope**: 15 new files, 3 modified files. No Go source changes.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Constitution v1.0.0 (6 principles):
- **I. Single Static Binary**: Pass — GoReleaser uses CGO_ENABLED=0; build verification in polish phase
- **II. MCP-First Interface**: N/A — no new CLI/MCP features in this spec
- **III. Worktree Isolation**: N/A — no worktree changes
- **IV. Convention Over Configuration**: N/A — no configuration changes
- **V. Standard Go Practices**: Pass — golangci-lint config added; CI enforces go vet + tests
- **VI. Agent-Agnostic Design**: Pass — integration docs cover 4 MCP clients; OpenClaw skill is additive

No violations. Gate passed.

## Project Structure

### Documentation (this feature)

```text
specs/002-packaging-oss-openclaw/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output (minimal — no data changes)
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output (file manifest)
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
# New files created by this feature
.goreleaser.yml                         # GoReleaser v2 config
.golangci.yml                           # Linter config
.github/
├── workflows/
│   ├── ci.yml                          # CI pipeline (test/lint/build)
│   └── release.yml                     # Release workflow (GoReleaser)
├── ISSUE_TEMPLATE/
│   ├── bug_report.md                   # Bug report template
│   ├── feature_request.md              # Feature request template
│   └── mcp_tool_request.md             # MCP tool/resource request
└── pull_request_template.md            # PR checklist
install.sh                              # Curl-based installer
LICENSE                                 # MIT 2026
CONTRIBUTING.md                         # Contribution guidelines
CODE_OF_CONDUCT.md                      # Contributor Covenant v2.1
docs/integrations.md                    # MCP client setup guides
integrations/openclaw/skill.md          # OpenClaw agent skill

# Modified files
.gitignore                              # Add .worktrees/, dist/
Makefile                                # Add release-snapshot target
README.md                               # Add badges
```

**Structure Decision**: All new files are project-root configuration, GitHub templates, or documentation. No changes to the existing `cmd/`, `internal/`, or Go source structure.

## Complexity Tracking

No constitution violations to justify. This feature adds only configuration and documentation files.
