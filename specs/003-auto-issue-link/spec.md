# Feature Specification: Automatic Bug Report Issue Link

**Feature Branch**: `003-auto-issue-link`
**Created**: 2026-02-20
**Status**: Draft
**Input**: User description: "When a user encounters a bug while using agit, the tool should automatically generate a pre-filled GitHub issue link. The user clicks the link to create the issue, making it easy to report bugs and improve the project."

## Clarifications

### Session 2026-02-20

- Q: Should the issue link appear on all errors or only unexpected/internal errors? → A: Only unexpected/internal errors. User input validation errors (wrong path, missing arg) are not bugs and should not show an issue link.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - CLI Error Produces Issue Link (Priority: P1)

A user runs an agit command that fails with an unexpected internal error (not a user input validation error). Along with the error message, agit displays a clickable URL that opens a new GitHub issue pre-filled with relevant context (error message, agit version, OS, and the command that was run). The user clicks the link, reviews the pre-filled content, adds any extra details, and submits the issue. User input errors (e.g., invalid path, missing required argument) display only the error message without an issue link.

**Why this priority**: This is the core value proposition. Without this, the feature does not exist. Every other story builds on the ability to generate a useful issue link from an error.

**Independent Test**: Can be fully tested by triggering an unexpected internal error (e.g., database corruption, git operation failure) and verifying the output contains a valid, clickable GitHub issue URL with pre-filled fields. Additionally, verify that a user input error (e.g., missing argument) does NOT produce an issue link.

**Acceptance Scenarios**:

1. **Given** a user runs an agit CLI command, **When** the command fails with an unexpected internal error, **Then** the error message is displayed followed by a pre-filled GitHub issue URL on a separate line.
2. **Given** an internal error occurs, **When** the issue link is generated, **Then** the URL opens a new issue form pre-filled with the error message, agit version, operating system, and the command that was run.
3. **Given** an internal error occurs, **When** the user copies or clicks the URL, **Then** the GitHub new-issue page loads with the "bug_report" issue template selected and fields populated.
4. **Given** a user runs an agit CLI command with invalid input (e.g., wrong path, missing argument), **When** the command fails with a validation error, **Then** only the error message is displayed without an issue link.

---

### User Story 2 - MCP Server Error Produces Issue Link (Priority: P2)

An AI agent calls an agit MCP tool that returns an error. The error response includes the same pre-filled GitHub issue URL so that the agent (or the human supervising it) can easily report the problem.

**Why this priority**: MCP is the primary interface for AI agents. Surfacing issue links in MCP error responses ensures the reporting mechanism works across both human and agent interfaces.

**Independent Test**: Can be tested by calling an MCP tool with invalid parameters and verifying the error response includes a valid GitHub issue URL.

**Acceptance Scenarios**:

1. **Given** an AI agent calls an agit MCP tool, **When** the tool returns an error, **Then** the error response text includes a pre-filled GitHub issue URL.
2. **Given** an MCP error with an issue link, **When** the link is opened in a browser, **Then** the new-issue form is pre-filled with MCP tool name, error details, and environment info.

---

### User Story 3 - Opt-Out of Issue Links (Priority: P3)

A user or automated pipeline finds the issue links noisy and wants to suppress them. They can disable the feature via an environment variable so that errors display without the appended URL.

**Why this priority**: Power users and CI/CD pipelines need the ability to suppress extra output. This respects user control without complicating the default experience.

**Independent Test**: Can be tested by setting the environment variable, triggering an error, and verifying no issue URL appears in the output.

**Acceptance Scenarios**:

1. **Given** a user has set the opt-out environment variable, **When** a CLI command fails, **Then** only the error message is displayed without an issue link.
2. **Given** the opt-out variable is not set (default), **When** a CLI command fails, **Then** the issue link is displayed as normal.

---

### Edge Cases

- What happens when the error message contains special characters (quotes, ampersands, newlines) that need URL-encoding?
- What happens when the generated URL exceeds browser URL length limits (~2,000 characters)? The system should truncate the pre-filled body gracefully.
- What happens when the user is offline? The link is still displayed (it's a URL, not a network call), and will work when the user is back online.
- What happens when the GitHub repository URL is not configured or changes? The system should use the module path from go.mod as the default.
- How does the system handle panics vs. returned errors? Panics should also produce an issue link via a recovery handler.
- What happens when an error is ambiguous (could be user input or internal)? The system should err on the side of showing the issue link -- it is better to occasionally show a link on a user error than to miss a real bug.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST append a pre-filled GitHub issue URL to CLI error messages caused by unexpected internal errors (e.g., database failures, git operation errors, panics). User input validation errors (e.g., invalid path, missing argument) MUST NOT include an issue link.
- **FR-011**: System MUST distinguish between user input validation errors and unexpected internal errors to determine whether an issue link is appropriate.
- **FR-002**: System MUST include the following context in the pre-filled issue: error message, agit version, operating system, architecture, and the command that triggered the error.
- **FR-003**: System MUST URL-encode all dynamic content in the generated link to handle special characters safely.
- **FR-004**: System MUST truncate the pre-filled issue body if the total URL length would exceed 2,000 characters.
- **FR-005**: System MUST include the issue link in MCP tool error responses so AI agents and their operators can report bugs.
- **FR-006**: System MUST allow users to disable issue link generation via an environment variable (`AGIT_NO_ISSUE_LINK=1`).
- **FR-007**: System MUST select the "bug_report" issue template when generating the GitHub URL.
- **FR-008**: System MUST recover from panics and display an issue link with panic details before exiting.
- **FR-009**: System MUST derive the GitHub repository URL from the project's module path rather than hardcoding it.
- **FR-010**: System MUST format the issue link on its own line, visually separated from the error message, so it is easy to identify and click.

### Key Entities

- **Error Context**: The collection of diagnostic information gathered at the point of failure -- error message, command, version, OS, architecture.
- **Issue URL**: A generated GitHub "new issue" URL containing query parameters for title, body, labels, and template selection.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of unexpected internal CLI errors include a valid, clickable GitHub issue URL (when opt-out is not set). 0% of user input validation errors include an issue link.
- **SC-002**: 100% of unexpected internal MCP tool errors include a valid GitHub issue URL in the error text. User input validation errors in MCP responses do not include an issue link.
- **SC-003**: Pre-filled issue content is accurate -- the error message, version, OS, and triggering command match the actual error context.
- **SC-004**: URLs with special characters in error messages are properly encoded and open correctly in all major browsers.
- **SC-005**: No issue URL appears in output when the opt-out environment variable is set.
- **SC-006**: Generated URLs do not exceed 2,000 characters; longer content is truncated with a note indicating truncation.

## Assumptions

- The GitHub repository is public, so anyone with the link can open a new issue.
- The `bug_report.md` issue template (created in spec 002) is present in `.github/ISSUE_TEMPLATE/`.
- The agit version is embedded at build time (already handled by the Makefile's `-ldflags` pattern).
- Standard GitHub "new issue" URL query parameters (`title`, `body`, `labels`, `template`) are used -- no GitHub API calls are needed.
- URL length limit of 2,000 characters is a safe cross-browser threshold.

## Dependencies

- Spec 002 (packaging & OSS scaffolding) -- provides the `bug_report.md` issue template that this feature references.
