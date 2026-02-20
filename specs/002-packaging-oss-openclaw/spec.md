# Feature Specification: Packaging, OSS Setup & OpenClaw Integration

**Feature Branch**: `002-packaging-oss-openclaw`
**Created**: 2026-02-20
**Status**: Draft
**Input**: User description: "Packaging, Open Source Setup, and OpenClaw Integration"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Install agit from a release binary (Priority: P1)

A developer wants to install agit on their machine without building from source. They run a single shell command or use Homebrew, and agit is immediately available in their PATH.

**Why this priority**: Without distributable binaries, agit cannot reach users beyond Go developers. This is the critical path to adoption.

**Independent Test**: Can be fully tested by running `goreleaser release --snapshot --clean` locally and verifying binaries for all 5 target platforms are produced, then testing the install script on a clean environment.

**Acceptance Scenarios**:

1. **Given** a tagged release exists on GitHub, **When** a macOS user runs `brew tap fathindos/tap && brew install agit`, **Then** `agit --version` outputs the correct version.
2. **Given** a tagged release exists on GitHub, **When** a Linux user runs the curl install script, **Then** the script downloads the correct binary for their OS/arch, verifies the checksum, and places it in a writable bin directory.
3. **Given** the project source code, **When** GoReleaser runs with `--snapshot`, **Then** archives are produced for linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, and windows/amd64.
4. **Given** a Windows user downloads the release zip, **When** they extract it, **Then** `agit.exe` is present and runs correctly.

---

### User Story 2 - Contribute to agit as an open source project (Priority: P2)

A developer discovers agit on GitHub and wants to contribute. They find clear contribution guidelines, issue templates, a code of conduct, and CI that validates their PR automatically.

**Why this priority**: Open source scaffolding must exist before publicizing the project. Without it, early contributors have a poor experience and may not return.

**Independent Test**: Can be tested by verifying all OSS files exist (LICENSE, CONTRIBUTING.md, CODE_OF_CONDUCT.md, issue templates, PR template), CI workflow runs on PR, and linter config catches common issues.

**Acceptance Scenarios**:

1. **Given** a new contributor visits the GitHub repo, **When** they click "New Issue", **Then** they see templates for bug reports, feature requests, and MCP tool requests.
2. **Given** a contributor opens a PR, **When** CI runs, **Then** it executes tests on Go 1.22 and 1.23, runs the linter, and verifies the build compiles with CGO disabled.
3. **Given** a contributor reads CONTRIBUTING.md, **When** they follow the dev setup instructions, **Then** they can build and test the project locally.
4. **Given** the README, **When** a visitor views it, **Then** they see CI status, latest release version, Go Report Card, and license badges.

---

### User Story 3 - Connect agit to an AI agent via MCP (Priority: P2)

An AI agent operator (using Claude Code, OpenClaw, Cursor, or another MCP client) wants to add agit to their agent's tool set. They find integration documentation with copy-paste config snippets.

**Why this priority**: MCP integration is agit's core value proposition. Clear integration docs are essential for adoption alongside the binary distribution.

**Independent Test**: Can be tested by verifying the integration docs contain valid JSON config snippets for each supported client, and the OpenClaw skill file follows the expected format.

**Acceptance Scenarios**:

1. **Given** a Claude Code user, **When** they add the agit MCP config snippet to their `mcp.json`, **Then** agit tools become available in their Claude Code session.
2. **Given** an OpenClaw operator, **When** they add the agit config to their gateway settings, **Then** agents can call `agit_list_repos` immediately.
3. **Given** the integration docs, **When** a user follows the verification steps, **Then** they can confirm agit is working by listing repos through the MCP interface.
4. **Given** the OpenClaw skill file, **When** an agent loads it, **Then** the agent understands the recommended workflow for using agit tools.

---

### Edge Cases

- What happens when the install script runs on an unsupported OS or architecture?
- What happens when the install script cannot write to `/usr/local/bin` and `~/.local/bin` doesn't exist?
- What happens when the GitHub API is unreachable during install script execution?
- What happens when a contributor's PR fails CI due to lint issues introduced by the new golangci-lint config?
- What happens when GoReleaser runs but the Homebrew tap token is not configured?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Project MUST produce static binaries for linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, and windows/amd64 via GoReleaser
- **FR-002**: Project MUST generate sha256 checksums for all release archives
- **FR-003**: Project MUST produce tar.gz archives for Linux/macOS and zip archives for Windows
- **FR-004**: Project MUST auto-publish a Homebrew formula to a tap repository on tagged releases
- **FR-005**: Project MUST produce deb and rpm packages for Linux distribution
- **FR-006**: An install script MUST detect OS and architecture, download the correct binary, verify its checksum, and install it to a writable bin directory
- **FR-007**: The install script MUST work with curl or wget as the HTTP client
- **FR-008**: The install script MUST fall back to `~/.local/bin` if `/usr/local/bin` requires elevated permissions
- **FR-009**: A CI pipeline MUST run tests, linting, and build verification on every push to main and every pull request
- **FR-010**: CI MUST test against Go 1.22 and Go 1.23
- **FR-011**: A release workflow MUST trigger on version tags (`v*`) and run GoReleaser
- **FR-012**: Project MUST include an MIT license file
- **FR-013**: Project MUST include contribution guidelines with dev setup, code style, and PR process
- **FR-014**: Project MUST include a Contributor Covenant v2.1 code of conduct
- **FR-015**: Project MUST include GitHub issue templates for bug reports, feature requests, and MCP tool requests
- **FR-016**: Project MUST include a pull request template with type classification and checklist
- **FR-017**: Project MUST include a golangci-lint configuration with errcheck, gosimple, govet, ineffassign, staticcheck, unused, gofmt, misspell, and revive linters
- **FR-018**: README MUST display CI, release, Go Report Card, and license badges
- **FR-019**: Integration documentation MUST provide MCP config snippets for Claude Code, OpenClaw, Cursor, and generic MCP clients
- **FR-020**: Integration documentation MUST cover both stdio and SSE transport modes
- **FR-021**: An OpenClaw skill file MUST describe when and how agents should use agit tools
- **FR-022**: The Makefile MUST include a `release-snapshot` target for local GoReleaser testing

### Key Entities

- **Release Archive**: A platform-specific compressed package containing the agit binary, produced by GoReleaser for each OS/arch combination
- **Homebrew Formula**: An auto-generated package definition pushed to the tap repository, enabling `brew install agit`
- **CI Pipeline**: Automated quality gates that validate code on every change (tests, linting, build verification)
- **MCP Config Snippet**: A JSON configuration block that users paste into their agent's settings to connect agit as an MCP server
- **OpenClaw Skill**: A markdown-based definition that teaches agents the recommended workflow for using agit tools

## Assumptions

- The Homebrew tap repository (`fathindos/homebrew-tap`) will be created manually on GitHub before the first release
- The `HOMEBREW_TAP_TOKEN` secret will be configured manually in the GitHub repository settings
- CGO-free compilation has already been verified (modernc.org/sqlite is pure Go) — confirmed in spec 001
- The project compiles and passes `go vet` on the current codebase (verified in spec 001)
- GoReleaser v2 syntax is used (with `version: 2` header)
- golangci-lint is available in CI via the official GitHub Action

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: GoReleaser snapshot produces binaries for all 5 target platforms without errors
- **SC-002**: The install script passes syntax validation (`bash -n install.sh`) and handles both curl and wget code paths
- **SC-003**: All required OSS files exist and follow their respective format standards (15 new files, 3 modified files)
- **SC-004**: CI workflow configuration is syntactically valid and covers the Go 1.22 + 1.23 test matrix
- **SC-005**: The project builds and passes `go vet` after all changes are applied
- **SC-006**: Integration docs contain working JSON config snippets for at least 3 MCP clients
- **SC-007**: The golangci-lint configuration is parseable and the enabled linters match the specification
