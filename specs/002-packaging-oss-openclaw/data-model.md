# Data Model: Packaging, OSS Setup & OpenClaw Integration

**Feature**: `002-packaging-oss-openclaw`
**Date**: 2026-02-20

## Overview

This feature introduces no new data entities or database changes. All artifacts are static configuration files, documentation, and CI/CD workflow definitions.

## Entities

### Release Archive (produced by GoReleaser)

- **Name pattern**: `agit_{version}_{os}_{arch}.{ext}`
- **Platforms**: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64
- **Formats**: tar.gz (Linux/macOS), zip (Windows)
- **Contents**: Single `agit` binary (or `agit.exe` for Windows)
- **Checksum**: sha256 in `checksums.txt`

### Homebrew Formula (auto-generated)

- **Repository**: `fathindos/homebrew-tap`
- **Path**: `Formula/agit.rb`
- **Lifecycle**: Auto-updated by GoReleaser on each tagged release

### Version String (injected at build time)

- **Location**: `cmd.Version` variable
- **Source**: Git tag via GoReleaser `{{.Version}}`
- **Format**: Semantic version without `v` prefix (e.g., `0.1.0`)

## State Transitions

None. This feature has no stateful components.

## Relationships to Existing Data

- The `cmd.Version` variable already exists in `cmd/root.go` and is set via ldflags in the Makefile
- GoReleaser uses the same ldflags pattern, ensuring consistency between `make build` and release builds
