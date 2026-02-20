# File Manifest: Packaging, OSS Setup & OpenClaw Integration

**Feature**: `002-packaging-oss-openclaw`
**Date**: 2026-02-20

## Overview

This feature has no API contracts (no new endpoints, tools, or resources). The "contracts" are the file artifacts that must be created and their required content structure.

## New Files (15)

| File | Type | Content Source |
|------|------|---------------|
| `.goreleaser.yml` | Config | GoReleaser v2 build/release config |
| `.golangci.yml` | Config | Linter configuration |
| `.github/workflows/ci.yml` | CI/CD | Test/lint/build pipeline |
| `.github/workflows/release.yml` | CI/CD | Tag-triggered release |
| `.github/ISSUE_TEMPLATE/bug_report.md` | Template | Bug report with env info |
| `.github/ISSUE_TEMPLATE/feature_request.md` | Template | Feature proposal |
| `.github/ISSUE_TEMPLATE/mcp_tool_request.md` | Template | MCP tool/resource request |
| `.github/pull_request_template.md` | Template | PR checklist |
| `install.sh` | Script | OS-detect + download + verify + install |
| `LICENSE` | Legal | MIT, 2026, fathindos |
| `CONTRIBUTING.md` | Docs | Dev setup, code style, PR process |
| `CODE_OF_CONDUCT.md` | Docs | Contributor Covenant v2.1 |
| `docs/integrations.md` | Docs | MCP config for 4 clients |
| `integrations/openclaw/skill.md` | Skill | Agent workflow guidance |

## Modified Files (3)

| File | Change |
|------|--------|
| `.gitignore` | Add `.worktrees/`, `dist/`; remove `docs/*`, `specs/*` |
| `Makefile` | Add `release-snapshot` target |
| `README.md` | Add CI/release/Go Report/license badges |

## Validation Contracts

Each file must satisfy these validation checks:

| Check | Command | Expected |
|-------|---------|----------|
| Go builds | `go build ./...` | Exit 0 |
| Go vet | `go vet ./...` | Exit 0 |
| Install script syntax | `bash -n install.sh` | Exit 0 |
| GoReleaser snapshot | `goreleaser release --snapshot --clean` | Produces 5 platform binaries |
| golangci-lint parseable | `golangci-lint run --help` with `.golangci.yml` present | Config recognized |
