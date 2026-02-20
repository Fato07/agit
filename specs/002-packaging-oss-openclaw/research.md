# Research: Packaging, OSS Setup & OpenClaw Integration

**Feature**: `002-packaging-oss-openclaw`
**Date**: 2026-02-20

## R1: GoReleaser v2 Configuration

**Decision**: Use GoReleaser v2 with `version: 2` header, CGO_ENABLED=0, and ldflags for version injection.

**Rationale**:
- GoReleaser v2 is the current stable version with the `version: 2` YAML header
- CGO_ENABLED=0 is mandatory because agit uses `modernc.org/sqlite` (pure Go) and must cross-compile
- The `-s -w` ldflags strip debug info, reducing binary size by ~30%
- Version injection via `-X github.com/fathindos/agit/cmd.Version={{.Version}}` follows Go convention

**Alternatives considered**:
- **goreleaser v1**: Deprecated, v2 is current
- **Manual release scripts**: More maintenance burden, less ecosystem support
- **ko (container-focused)**: Not appropriate for CLI distribution

**Key details**:
- Archives: tar.gz for Linux/macOS, zip for Windows (standard convention)
- Checksums: sha256 in `checksums.txt` (used by install script for verification)
- Homebrew: Auto-push formula to `fathindos/homebrew-tap` via `brews` section
- nfpms: deb + rpm for Linux package managers
- The `before.hooks` section runs `go mod tidy` to ensure clean module state

## R2: CI Pipeline Design

**Decision**: Two-job CI with Go 1.22/1.23 matrix on ubuntu-latest, plus golangci-lint via official action.

**Rationale**:
- Go 1.22 is the minimum supported version (spec requirement); 1.23 is the version in go.mod
- The official `golangci/golangci-lint-action` handles caching and version pinning
- Separate lint and test jobs allow parallel execution and clear failure attribution
- `CGO_ENABLED=0 go build .` as a build verification step catches compilation issues

**Alternatives considered**:
- **Single Go version**: Would miss compatibility issues between Go releases
- **Three versions (1.21+)**: Unnecessary — go.mod specifies 1.23, 1.22 is reasonable minimum
- **Self-hosted runners**: Overkill for a small project

## R3: golangci-lint Configuration

**Decision**: Enable errcheck, gosimple, govet, ineffassign, staticcheck, unused, gofmt, misspell, and revive.

**Rationale**:
- This set covers the most impactful Go linters without being overly strict
- `errcheck`: Catches unchecked errors (critical for reliability)
- `gosimple` + `staticcheck`: Standard static analysis
- `govet`: Catches common Go mistakes
- `gofmt`: Enforces consistent formatting
- `misspell`: Catches typos in strings and comments
- `revive`: Modern replacement for golint, configurable
- Test files excluded from `errcheck` (common pattern for test helpers)

**Alternatives considered**:
- **Default linter set only**: Misses valuable checks like misspell and revive
- **Exhaustive linter set (50+ linters)**: Too noisy, slows CI, causes contributor friction
- **Custom revive rules**: Only `exported` rule enabled to avoid excessive strictness

## R4: Install Script Patterns

**Decision**: Shell script with OS/arch detection, GitHub API for latest release, checksum verification, fallback install paths.

**Rationale**:
- Follows patterns established by goreleaser's own install script and similar tools (buf, golangci-lint)
- `set -euo pipefail` for safety (fail on any error, undefined vars, pipe failures)
- OS detection via `uname -s` / `uname -m` with mapping to GoReleaser archive names
- GitHub API (`/repos/{owner}/{repo}/releases/latest`) for release discovery
- curl preferred, wget as fallback (covers most Linux/macOS environments)
- Checksum verification via `sha256sum` or `shasum -a 256` (macOS compatibility)
- Install to `/usr/local/bin` with fallback to `~/.local/bin` if no write permission

**Alternatives considered**:
- **Go install only**: Requires Go toolchain, limits audience
- **Docker distribution**: Adds complexity for a CLI tool
- **Snap/Flatpak**: Limited audience, complex packaging

## R5: Open Source File Standards

**Decision**: MIT license, Contributor Covenant v2.1, conventional commits, GitHub issue/PR templates.

**Rationale**:
- MIT is the most permissive standard license, appropriate for developer tools
- Contributor Covenant v2.1 is the most widely adopted code of conduct
- Conventional commits (feat/fix/docs/test/ci/refactor/chore) enable automated changelogs
- Three issue templates (bug, feature, MCP tool) cover the project's primary contribution types
- PR template with checklist ensures consistent quality across contributions

**Alternatives considered**:
- **Apache 2.0**: More complex, patent clause unnecessary for this project
- **No code of conduct**: Poor practice for OSS, signals unwelcoming community
- **Custom commit format**: Conventional commits have the broadest tooling support

## R6: MCP Integration Documentation

**Decision**: Provide JSON config snippets for Claude Code, OpenClaw, Cursor, and generic MCP clients. Include both stdio and SSE transport examples.

**Rationale**:
- Copy-paste JSON snippets minimize friction for users
- stdio is the default transport (simplest, no network setup)
- SSE is needed for remote deployments (agit on a different machine than the agent)
- OpenClaw skill file provides agent-friendly prompting for the recommended workflow

**Alternatives considered**:
- **Only stdio**: Would limit remote deployment scenarios
- **Only documentation, no skill file**: Agents would lack guidance on when/how to use agit tools
- **Separate npm/pip package for integration**: Unnecessary complexity for JSON config
