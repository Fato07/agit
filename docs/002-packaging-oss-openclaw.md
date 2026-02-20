# Spec 002: Packaging, Open Source, and OpenClaw Integration

## Overview

This spec covers three workstreams that prepare agit for public release:

1. **Packaging & Distribution** — GoReleaser, Homebrew tap, install script, GitHub Releases
2. **Open Source Setup** — CONTRIBUTING.md, CI pipeline, issue/PR templates, code of conduct, license
3. **OpenClaw Integration** — MCP config, OpenClaw skill, integration guide

All three should be implemented in order. The packaging must work before we publicize; the OSS scaffolding must exist before contributors arrive; the OpenClaw integration is the first "proof of value" for the agent ecosystem.

---

## Part 1: Packaging & Distribution

### 1.1 GoReleaser Configuration

Create `.goreleaser.yml` at the project root.

**Build targets:**
- linux/amd64, linux/arm64
- darwin/amd64, darwin/arm64 (macOS Intel + Apple Silicon)
- windows/amd64

**Configuration:**

```yaml
version: 2

project_name: agit

before:
  hooks:
    - go mod tidy

builds:
  - id: agit
    main: .
    binary: agit
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w
      - -X github.com/fathindos/agit/cmd.Version={{.Version}}

archives:
  - id: default
    format: tar.gz
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
    format_overrides:
      - goos: windows
        format: zip

checksum:
  name_template: "checksums.txt"
  algorithm: sha256

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^test:"
      - "^ci:"
      - "^chore:"

brews:
  - repository:
      owner: fathindos
      name: homebrew-tap
      token: "{{ .Env.HOMEBREW_TAP_TOKEN }}"
    commit_author:
      name: agit-bot
      email: bot@agit.dev
    directory: Formula
    homepage: "https://github.com/fathindos/agit"
    description: "Infrastructure-aware Git orchestration for AI agents"
    license: "MIT"
    install: |
      bin.install "agit"
    test: |
      system "#{bin}/agit", "--version"

release:
  github:
    owner: fathindos
    name: agit
  draft: false
  prerelease: auto
  name_template: "v{{.Version}}"

nfpms:
  - id: packages
    package_name: agit
    vendor: fathindos
    homepage: https://github.com/fathindos/agit
    description: "Infrastructure-aware Git orchestration for AI agents"
    license: MIT
    formats:
      - deb
      - rpm
    bindir: /usr/local/bin
```

### 1.2 Homebrew Tap Repository

Create a new GitHub repository: `fathindos/homebrew-tap`

This is an empty repo. GoReleaser will automatically push the formula to it on release. Users install via:

```bash
brew tap fathindos/tap
brew install agit
```

### 1.3 Install Script

Create `install.sh` at the project root. This is for users who don't use Homebrew or Go.

**Behavior:**
1. Detect OS (linux/darwin) and architecture (amd64/arm64)
2. Fetch the latest release from GitHub API
3. Download the correct archive
4. Verify checksum
5. Extract binary to `/usr/local/bin/agit` (or `~/.local/bin/agit` if no sudo)
6. Print success message with next steps

**Script requirements:**
- Must work with only `curl` and `tar` (no dependencies)
- Support both `curl` and `wget` as fallback
- Colorized output (but degrade gracefully if no tty)
- `set -euo pipefail` for safety
- Usage: `curl -sSfL https://raw.githubusercontent.com/fathindos/agit/main/install.sh | sh`

### 1.4 GitHub Actions Release Workflow

Create `.github/workflows/release.yml`:

**Trigger:** Push of a tag matching `v*` (e.g., `v0.1.0`)

**Steps:**
1. Checkout code
2. Set up Go 1.22+
3. Run tests (`go test ./...`)
4. Run GoReleaser (`goreleaser release --clean`)

**Secrets needed:**
- `GITHUB_TOKEN` (automatic)
- `HOMEBREW_TAP_TOKEN` (personal access token with repo scope for the homebrew-tap repo)

### 1.5 CGO-Free SQLite

**IMPORTANT:** The current code uses `modernc.org/sqlite` which is pure Go (no CGO). This is critical for cross-compilation. Verify that `CGO_ENABLED=0` builds work. If there are any CGO dependencies, replace them with pure Go alternatives.

**Test command:**
```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o agit-linux .
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o agit-darwin .
```

Both must succeed.

---

## Part 2: Open Source Setup

### 2.1 LICENSE

MIT license. File already exists. Verify it has the correct year and author.

### 2.2 CONTRIBUTING.md

Create `CONTRIBUTING.md` at the project root with these sections:

**Welcome**
- Brief description of the project and what contributions are welcome
- Link to the architecture spec (`docs/architecture-spec.docx`)
- Link to open issues labeled `good first issue`

**Development Setup**
```
git clone https://github.com/fathindos/agit.git
cd agit
go mod tidy
make build
./agit init
./agit add /path/to/any/git/repo
./agit repos
```

**Project Structure**
- Brief description of each directory (cmd/, internal/registry/, internal/git/, internal/mcp/, etc.)
- Point to the architecture spec for detailed design

**Making Changes**
1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Make your changes
4. Run tests: `make test`
5. Run linter: `make lint` (requires golangci-lint)
6. Commit with a descriptive message following conventional commits:
   - `feat: add SSE transport support`
   - `fix: handle empty repos in spawn`
   - `docs: update MCP tool descriptions`
7. Push and open a PR

**Code Style**
- Follow standard Go conventions (`go fmt`, `go vet`)
- Use `golangci-lint` with the project's config
- Keep functions focused and under 50 lines where possible
- Add comments for exported functions
- Write tests for new functionality

**Commit Messages**
- Use conventional commits: `feat:`, `fix:`, `docs:`, `test:`, `ci:`, `refactor:`, `chore:`
- Keep the first line under 72 characters
- Reference issues: `feat: add task priority (#42)`

**Pull Request Guidelines**
- PRs should address a single concern
- Include tests for new features
- Update documentation if behavior changes
- Fill out the PR template

**Reporting Issues**
- Link to the issue templates
- Ask reporters to include: agit version, OS, Go version, steps to reproduce

**Code of Conduct**
- Reference CODE_OF_CONDUCT.md

### 2.3 CODE_OF_CONDUCT.md

Use the Contributor Covenant v2.1 (standard for open source projects). Download from https://www.contributor-covenant.org/version/2/1/code_of_conduct/

### 2.4 Issue Templates

Create `.github/ISSUE_TEMPLATE/` with three templates:

**bug_report.md:**
```yaml
---
name: Bug Report
about: Something isn't working
title: "[BUG] "
labels: bug
assignees: ''
---

**Describe the bug**
A clear description of what happened.

**To reproduce**
Steps to reproduce:
1. Run `agit ...`
2. ...

**Expected behavior**
What should have happened.

**Environment**
- agit version: (run `agit --version`)
- OS:
- Go version: (if building from source)
- Git version:

**Additional context**
Logs, screenshots, etc.
```

**feature_request.md:**
```yaml
---
name: Feature Request
about: Suggest an enhancement
title: "[FEATURE] "
labels: enhancement
assignees: ''
---

**Problem**
What problem does this solve?

**Proposed solution**
How should it work?

**Alternatives considered**
Other approaches you've thought about.

**Additional context**
Mockups, examples, related issues.
```

**mcp_tool_request.md:**
```yaml
---
name: MCP Tool Request
about: Request a new MCP tool or resource
title: "[MCP] "
labels: mcp, enhancement
assignees: ''
---

**Tool or Resource?**
Tool / Resource

**Name**
e.g., `agit_create_pr`

**Description**
What should it do?

**Parameters**
List the input parameters.

**Use case**
How would an agent use this?
```

### 2.5 Pull Request Template

Create `.github/pull_request_template.md`:

```markdown
## What does this PR do?

<!-- Brief description -->

## Type of change

- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation

## How has this been tested?

<!-- Describe tests run -->

## Checklist

- [ ] My code follows the project's style guidelines
- [ ] I have added tests for my changes
- [ ] All new and existing tests pass (`make test`)
- [ ] I have updated documentation if needed
```

### 2.6 CI Pipeline

Create `.github/workflows/ci.yml`:

**Trigger:** Push to main, all PRs

**Jobs:**

**test:**
- Matrix: Go 1.22, 1.23 on ubuntu-latest
- Steps: checkout, setup Go, `go mod tidy`, `go vet ./...`, `go test ./... -race -coverprofile=coverage.out`
- Upload coverage to Codecov (optional but nice)

**lint:**
- Run golangci-lint via the official GitHub Action
- Use `.golangci.yml` config

**build:**
- Run `CGO_ENABLED=0 go build .` to verify it compiles
- Run on linux, darwin, windows matrix

### 2.7 golangci-lint Configuration

Create `.golangci.yml`:

```yaml
run:
  timeout: 5m

linters:
  enable:
    - errcheck
    - gosimple
    - govet
    - ineffassign
    - staticcheck
    - unused
    - gofmt
    - misspell
    - revive

linters-settings:
  revive:
    rules:
      - name: exported
        severity: warning

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - errcheck
```

### 2.8 README Badges

Add to the top of README.md:

```markdown
[![CI](https://github.com/fathindos/agit/actions/workflows/ci.yml/badge.svg)](https://github.com/fathindos/agit/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/fathindos/agit)](https://github.com/fathindos/agit/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/fathindos/agit)](https://goreportcard.com/report/github.com/fathindos/agit)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
```

---

## Part 3: OpenClaw Integration

### 3.1 OpenClaw MCP Configuration

Users add agit to their OpenClaw config (`openclaw.json` or via the Gateway settings):

```json
{
  "mcpServers": {
    "agit": {
      "command": "agit",
      "args": ["serve"],
      "env": {}
    }
  }
}
```

For SSE transport (e.g., if OpenClaw and agit run on different machines):

```json
{
  "mcpServers": {
    "agit": {
      "command": "agit",
      "args": ["serve", "--transport", "sse", "--port", "3847"],
      "env": {}
    }
  }
}
```

**Verify:** After adding the config, restart OpenClaw Gateway. The agent should be able to call `agit_list_repos` immediately.

### 3.2 OpenClaw Skill (Optional Enhancement)

Create an OpenClaw skill that wraps agit MCP with agent-friendly prompting. This is NOT required (the MCP server works directly), but it improves the agent experience by providing context about when to use agit tools.

**Skill location:** Publish to ClawHub or include in the agit repo at `integrations/openclaw/`

**Skill file:** `integrations/openclaw/skill.md`

Contents of the skill:

```markdown
---
name: agit
description: Git infrastructure awareness and multi-agent coordination
version: 0.1.0
mcp_servers:
  - agit
---

# agit - Git Infrastructure for AI Agents

You have access to agit, which provides persistent awareness of Git repositories
and coordinates work across multiple agents.

## When to use agit

- **Starting a new session**: Always call `agit_list_repos` first to see what
  repositories are available. Never assume you know what repos exist.
- **Before making changes**: Call `agit_spawn_worktree` to get an isolated
  workspace. Never work directly on the main branch.
- **Before merging**: Call `agit_check_conflicts` to see if your changes
  overlap with other agents' work.
- **When given a task**: Call `agit_list_tasks` to see if there are existing
  tasks, and `agit_claim_task` before starting work.

## Workflow

1. `agit_register_agent` - Announce yourself
2. `agit_list_repos` - Discover available repos
3. `agit_list_tasks` - Check for existing work
4. `agit_claim_task` or `agit_spawn_worktree` - Start working
5. `agit_check_conflicts` - Check periodically while working
6. `agit_complete_task` - Mark work as done
7. `agit_merge_worktree` - Merge when ready

## Important

- Always use worktrees. Never commit directly to the default branch.
- Check for conflicts before merging.
- Complete tasks when done so other agents know the work is finished.
```

### 3.3 Integration Documentation

Create `docs/integrations.md` with setup instructions for:

**Claude Code:**
```json
// ~/.claude/mcp.json
{
  "mcpServers": {
    "agit": {
      "command": "agit",
      "args": ["serve"]
    }
  }
}
```

**OpenClaw:**
```json
// openclaw.json mcpServers section
{
  "agit": {
    "command": "agit",
    "args": ["serve"]
  }
}
```

**Cursor:**
```json
// .cursor/mcp.json
{
  "mcpServers": {
    "agit": {
      "command": "agit",
      "args": ["serve"]
    }
  }
}
```

**Any MCP-compatible agent:**
- agit uses the standard MCP protocol (stdio by default, SSE optional)
- Just point your agent's MCP config to `agit serve`
- The server exposes 11 tools and 6 resources (see architecture spec Section 5)

### 3.4 Verification Test

After setting up the integration, verify with this test:

1. Start a new agent session
2. The agent should be able to call `agit_list_repos`
3. If repos were previously registered, they should appear
4. Try `agit_spawn_worktree` from the agent
5. Verify the agent can work in the spawned worktree path
6. Call `agit_check_conflicts` — should return empty with 1 worktree
7. Call `agit_complete_task` or `agit_merge_worktree` to clean up

---

## Implementation Order

### Phase A: Packaging (do first)
1. Verify `CGO_ENABLED=0` builds work
2. Create `.goreleaser.yml`
3. Create `.github/workflows/release.yml`
4. Create `install.sh`
5. Create `fathindos/homebrew-tap` repo on GitHub
6. Test: `goreleaser release --snapshot --clean` locally
7. Tag `v0.1.0` and push to trigger first release

### Phase B: Open Source Scaffolding
1. Create `LICENSE` (MIT, verify year/author)
2. Create `CONTRIBUTING.md`
3. Create `CODE_OF_CONDUCT.md`
4. Create `.github/ISSUE_TEMPLATE/` (3 templates)
5. Create `.github/pull_request_template.md`
6. Create `.github/workflows/ci.yml`
7. Create `.golangci.yml`
8. Add badges to `README.md`
9. Label initial issues as `good first issue`

### Phase C: OpenClaw Integration
1. Create `integrations/openclaw/skill.md`
2. Create `docs/integrations.md`
3. Test with local OpenClaw instance
4. Submit skill to ClawHub (optional)

---

## Files Created by This Spec

```
.goreleaser.yml
.golangci.yml
.github/
  workflows/
    ci.yml
    release.yml
  ISSUE_TEMPLATE/
    bug_report.md
    feature_request.md
    mcp_tool_request.md
  pull_request_template.md
install.sh
CONTRIBUTING.md
CODE_OF_CONDUCT.md
LICENSE (verify/update)
README.md (update with badges)
docs/integrations.md
integrations/openclaw/skill.md
```

Total: 14 new files, 1 updated file (README.md)
