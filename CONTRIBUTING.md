# Contributing to agit

Thank you for your interest in contributing to agit! This guide will help you get started.

agit is an infrastructure-aware Git orchestration tool for AI agents. We welcome contributions of all kinds: bug fixes, new features, documentation improvements, and MCP tool/resource additions.

- [Architecture specification](docs/agit-architecture-spec.pdf)
- [Open issues labeled `good first issue`](https://github.com/fathindos/agit/labels/good%20first%20issue)

## Development Setup

```bash
git clone https://github.com/fathindos/agit.git
cd agit
go mod tidy
make build
./agit init
./agit add /path/to/any/git/repo
./agit repos
```

Requires Go 1.22 or later.

## Project Structure

| Directory | Purpose |
|-----------|---------|
| `cmd/` | CLI command definitions (Cobra) |
| `internal/registry/` | SQLite-backed registry for repos, worktrees, tasks, agents |
| `internal/git/` | Git operations (worktrees, merging, diffs) |
| `internal/mcp/` | MCP server (tools, resources, server factory) |
| `internal/conflicts/` | Cross-worktree conflict detection |
| `internal/config/` | Configuration management |
| `docs/` | Architecture spec and integration guides |

See the [architecture specification](docs/agit-architecture-spec.pdf) for detailed design.

## Making Changes

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Make your changes
4. Run tests: `make test`
5. Run linter: `make lint` (requires [golangci-lint](https://golangci-lint.run/))
6. Commit with a descriptive message following conventional commits
7. Push and open a PR

## Code Style

- Follow standard Go conventions (`go fmt`, `go vet`)
- Use `golangci-lint` with the project's `.golangci.yml` config
- Keep functions focused and under 50 lines where possible
- Add comments for exported functions
- Write tests for new functionality

## Commit Messages

Use [conventional commits](https://www.conventionalcommits.org/):

- `feat: add SSE transport support`
- `fix: handle empty repos in spawn`
- `docs: update MCP tool descriptions`
- `test: add conflict detection tests`
- `ci: update Go version matrix`
- `refactor: simplify worktree cleanup logic`
- `chore: update dependencies`

Keep the first line under 72 characters. Reference issues when applicable: `feat: add task priority (#42)`

## Pull Request Guidelines

- PRs should address a single concern
- Include tests for new features
- Update documentation if behavior changes
- Fill out the PR template completely

## Reporting Issues

Use the [issue templates](https://github.com/fathindos/agit/issues/new/choose) to report bugs, request features, or propose new MCP tools. Please include:

- agit version (`agit --version`)
- Operating system
- Go version (if building from source)
- Steps to reproduce

## Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.
