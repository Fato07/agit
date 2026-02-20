# Quickstart: Packaging, OSS Setup & OpenClaw Integration

**Feature**: `002-packaging-oss-openclaw`
**Date**: 2026-02-20

## Prerequisites

- Go 1.22+ installed
- Git repository with spec 001 committed
- Branch: `002-packaging-oss-openclaw`

## Implementation Order

### Phase A: Packaging (do first)

1. Create `.goreleaser.yml` at project root (exact config from spec)
2. Create `.github/workflows/release.yml`
3. Create `install.sh` at project root
4. Update `Makefile` — add `release-snapshot` target

**Verify**: `goreleaser release --snapshot --clean` (requires goreleaser installed)

### Phase B: Open Source Scaffolding

1. Create `LICENSE` (MIT, 2026, fathindos)
2. Create `CONTRIBUTING.md`
3. Create `CODE_OF_CONDUCT.md` (Contributor Covenant v2.1)
4. Create `.github/ISSUE_TEMPLATE/bug_report.md`
5. Create `.github/ISSUE_TEMPLATE/feature_request.md`
6. Create `.github/ISSUE_TEMPLATE/mcp_tool_request.md`
7. Create `.github/pull_request_template.md`
8. Create `.github/workflows/ci.yml`
9. Create `.golangci.yml`
10. Update `README.md` — add badges at top

**Verify**: `go build ./...`, `go vet ./...`, `bash -n install.sh`

### Phase C: OpenClaw Integration

1. Create `integrations/openclaw/skill.md`
2. Create `docs/integrations.md`

**Verify**: Review JSON snippets are valid, skill file follows MCP skill format

## Key Decisions

- **GoReleaser v2** syntax with `version: 2` header
- **CGO_ENABLED=0** for all builds (pure Go SQLite)
- **Go 1.22 + 1.23** CI matrix
- **golangci-lint** with 9 linters: errcheck, gosimple, govet, ineffassign, staticcheck, unused, gofmt, misspell, revive
- **Install script** uses curl (wget fallback), sha256 checksum verification, `/usr/local/bin` → `~/.local/bin` fallback

## Post-Implementation

1. `go build ./...` — still compiles
2. `go vet ./...` — still clean
3. `bash -n install.sh` — script syntax valid
4. Optionally: `goreleaser release --snapshot --clean` (needs goreleaser binary)
5. Commit all changes on `002-packaging-oss-openclaw` branch
