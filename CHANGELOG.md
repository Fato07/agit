# Changelog

All notable changes to agit are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.4.0] - 2026-02-22

### Added
- Atomic task dispatch with priority-based claiming (`agit tasks next`)
- Conflict resolution suggestions with merge-order rationale
- Plugin/hook system for 6 lifecycle events (`worktree.created`, `worktree.removed`, `task.claimed`, `task.completed`, `task.failed`, `conflict.detected`)
- SSE graceful shutdown with SIGTERM/SIGINT handling
- Port-in-use detection with user-friendly error message
- `agit status` command for comprehensive repository overview
- `agit merge` command with automatic worktree cleanup
- `agit config reset` and JSON output format for config
- `agit_next_task` MCP tool (20 total MCP tools)
- Comprehensive CLI integration tests for all commands

### Fixed
- Hooks now wait for completion before process exits

## [0.3.0] - 2026-02-21

### Added
- Auto-update checking (`agit update` / `agit upgrade`)
- Enhanced UI with Unicode symbols
- `agit config` command (`show`, `set`, `path`, `reset`)
- 8 new MCP tools (19 total): `agit_fail_task`, `agit_start_task`, `agit_list_agents`, `agit_list_worktrees`, `agit_get_task`, `agit_add_repo`, `agit_cleanup_worktrees`, `agit_next_task`
- README and test coverage improvements

## [0.2.0] - 2026-02-21

### Added
- CLI UX overhaul with `internal/ui` package
- Shell completions (bash, zsh, fish, powershell)
- JSON output format (`-o json`)
- Interactive TUI mode (`-i`)
- Global flags (`--no-color`, `--quiet`)

## [0.1.1] - 2026-02-20

### Fixed
- Various bug fixes from OpenClaw integration test report

## [0.1.0] - 2026-02-20

### Added
- Initial release
- Repository registry with SQLite backend
- Git worktree management for agent isolation
- Cross-worktree conflict detection
- Task assignment and coordination
- Agent registration and heartbeat tracking
- MCP server with stdio transport (11 tools, 5 resources)

[0.4.0]: https://github.com/Fato07/agit/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/Fato07/agit/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/Fato07/agit/compare/v0.1.1...v0.2.0
[0.1.1]: https://github.com/Fato07/agit/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/Fato07/agit/releases/tag/v0.1.0
