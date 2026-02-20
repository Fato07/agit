# agit Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-02-20

## Active Technologies
- Go 1.23 (go.mod), CI tests against 1.22 + 1.23 + GoReleaser v2 (build/release), golangci-lint (linting), GitHub Actions (CI/CD) (002-packaging-oss-openclaw)
- N/A (no data model changes) (002-packaging-oss-openclaw)
- Go 1.23 (go.mod), CI tests against 1.22 + 1.23 + `github.com/spf13/cobra` (CLI), `github.com/mark3labs/mcp-go` (MCP protocol), stdlib `net/url`, `runtime`, `runtime/debug` (003-auto-issue-link)
- N/A (no persistence; URL generation is pure string computation) (003-auto-issue-link)

- Go 1.22 + `github.com/mark3labs/mcp-go` v0.20.1 (MCP protocol), `github.com/spf13/cobra` v1.8.1 (CLI), `modernc.org/sqlite` v1.34.4 (database) (001-agit-architecture-impl)

## Project Structure

```text
src/
tests/
```

## Commands

# Add commands for Go 1.22

## Code Style

Go 1.22: Follow standard conventions

## Recent Changes
- 003-auto-issue-link: Added Go 1.23 (go.mod), CI tests against 1.22 + 1.23 + `github.com/spf13/cobra` (CLI), `github.com/mark3labs/mcp-go` (MCP protocol), stdlib `net/url`, `runtime`, `runtime/debug`
- 002-packaging-oss-openclaw: Added Go 1.23 (go.mod), CI tests against 1.22 + 1.23 + GoReleaser v2 (build/release), golangci-lint (linting), GitHub Actions (CI/CD)

- 001-agit-architecture-impl: Added Go 1.22 + `github.com/mark3labs/mcp-go` v0.20.1 (MCP protocol), `github.com/spf13/cobra` v1.8.1 (CLI), `modernc.org/sqlite` v1.34.4 (database)

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
